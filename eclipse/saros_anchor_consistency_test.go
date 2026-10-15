package eclipse

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

// 本文件独立解析锚表与扩展表：成员号、唯一性、连续性与成员总数都由测试自己推导，
const (
	sarosTestLunationEpoch        = 2451550.09765
	sarosTestSynodicMonthDays     = 29.530588853
	sarosTestAnchorYearOffset     = 3000
	sarosTestAnchorCountMask      = 0x7f
	sarosTestAnchorDayMask        = 0x1f
	sarosTestAnchorMonthMask      = 0x0f
	sarosTestAnchorYearMask       = 0x1fff
	sarosTestAnchorDayShift       = 7
	sarosTestAnchorMonthShift     = 12
	sarosTestAnchorYearShift      = 16
	sarosTestAnchorMemberInterval = 223
)

type sarosTestWindow struct {
	headTT float64
	count  int
	offset int
}

func sarosTestWindows(t *testing.T, phase int) map[int]sarosTestWindow {
	t.Helper()
	anchors := solarSarosAnchors[:]
	overrides := solarSarosHeadOverrides[:]
	base := 0
	if phase == 1 {
		anchors = lunarSarosAnchors[:]
		overrides = lunarSarosHeadOverrides[:]
		base = 1
	}
	windows := make(map[int]sarosTestWindow, len(anchors)+len(overrides))
	for index, magic := range anchors {
		value := uint32(magic)
		headTT := basic.JDCalc(
			int((value>>sarosTestAnchorYearShift)&sarosTestAnchorYearMask)-sarosTestAnchorYearOffset,
			int((value>>sarosTestAnchorMonthShift)&sarosTestAnchorMonthMask),
			float64((value>>sarosTestAnchorDayShift)&sarosTestAnchorDayMask),
		)
		if math.IsNaN(headTT) {
			t.Fatalf("invalid anchor head for series %d", base+index)
		}
		windows[base+index] = sarosTestWindow{
			headTT: headTT,
			count:  int(value & sarosTestAnchorCountMask),
		}
	}
	for _, override := range overrides {
		headTT := basic.JDCalc(int(override.HeadYear), int(override.HeadMonth), float64(override.HeadDay))
		if math.IsNaN(headTT) {
			t.Fatalf("invalid override head for series %d", override.Series)
		}
		windows[int(override.Series)] = sarosTestWindow{
			headTT: headTT,
			count:  int(override.Count),
			offset: int(override.MemberOffset),
		}
	}
	return windows
}

func sarosTestReturnTT(k, phase int) float64 {
	return sarosTestLunationEpoch + (float64(k)+float64(phase)/2)*sarosTestSynodicMonthDays
}

func sarosTestMember(window sarosTestWindow, ttJDE float64) int {
	return int(math.Round((ttJDE-window.headTT)/(float64(sarosTestAnchorMemberInterval)*sarosTestSynodicMonthDays))) +
		1 + window.offset
}

func sarosTestSpans(phase int) []sarosSpan {
	if phase == 1 {
		return sarosExtensionSpans(1)
	}
	return sarosExtensionSpans(0)
}

func sarosTestInfo(ttJDE float64, phase int) (SarosInfo, bool) {
	if phase == 0 {
		return solarSarosInfo(ttJDE)
	}
	return lunarSarosInfo(ttJDE)
}

// 扩展表只对窗口内的时刻生效；窗口外的序列行由外推路径接管。
func sarosTestInsideTable(ttJDE float64) bool {
	return ttJDE >= sarosExtendedStartTT && ttJDE < sarosExtendedEndTT
}

// 锚表窗口、扩展表成员号与公开查询必须一致：成员号唯一、连续，总数等于锚表的 Count，
// 序列号与扩展表相同；目录窗口之外的成员只报"没有沙罗信息"。
func TestSarosAnchorWindowsMatchExtensionSpans(t *testing.T) {
	for _, phase := range []int{0, 1} {
		name := "solar"
		if phase == 1 {
			name = "lunar"
		}
		t.Run(name, func(t *testing.T) {
			windows := sarosTestWindows(t, phase)
			reported := make(map[int]map[int]int)
			for _, span := range sarosTestSpans(phase) {
				if span.First > span.Last || (span.Last-span.First)%sarosTestAnchorMemberInterval != 0 {
					t.Fatalf("malformed span %+v", span)
				}
				window, anchored := windows[span.Series]
				for index, k := 0, span.First; k <= span.Last; k, index = k+sarosTestAnchorMemberInterval, index+1 {
					ttJDE := sarosTestReturnTT(k, phase)
					if !sarosTestInsideTable(ttJDE) {
						continue
					}
					spanMember := span.Member + index
					want := SarosInfo{Series: span.Series, Member: spanMember, Count: span.Count}
					if anchored {
						want.Member = sarosTestMember(window, ttJDE)
						want.Count = window.count
						if want.Member < 1 || want.Member > window.count {
							want = SarosInfo{}
						}
					}
					// 扩展表派发本身就要采用锚表的成员总数；协调会再覆盖一次，所以单独钉住输入侧。
					provisional, provisionalOK := extendedSarosInfo(ttJDE, phase)
					if !provisionalOK || provisional.Series != span.Series || provisional.Member != spanMember {
						t.Fatalf("series %d k=%d extension dispatch=%+v/%v, want member %d",
							span.Series, k, provisional, provisionalOK, spanMember)
					}
					wantProvisionalCount := span.Count
					if anchored {
						wantProvisionalCount = window.count
					}
					if provisional.Count != wantProvisionalCount {
						t.Fatalf("series %d k=%d extension count=%d, want %d", span.Series, k, provisional.Count, wantProvisionalCount)
					}
					info, ok := reconciledSarosInfo(extendedSarosInfo, ttJDE, phase)
					if ok != (want != SarosInfo{}) || info.Series != want.Series ||
						info.Member != want.Member || info.Count != want.Count {
						t.Fatalf("series %d k=%d dispatch=%+v/%v, independently derived %+v", span.Series, k, info, ok, want)
					}
					if !ok {
						continue
					}
					// 唯一性与连续性按测试自己推导的成员号记账，覆盖窗口内每个回次。
					if reported[span.Series] == nil {
						reported[span.Series] = make(map[int]int)
					}
					if previous, seen := reported[span.Series][want.Member]; seen {
						t.Fatalf("series %d member %d reported for both k=%d and k=%d", span.Series, want.Member, previous, k)
					}
					reported[span.Series][want.Member] = k
					// 序列头、尾与中间各抽一点走公开入口，锚表匹配器与扩展表必须给出同一成员。
					if index != 0 && index != 1 && k != span.Last && k != span.Last-sarosTestAnchorMemberInterval {
						continue
					}
					public, publicOK := sarosTestInfo(ttJDE, phase)
					wantPublic := SarosInfo{Series: want.Series, Member: want.Member, Count: want.Count, Verified: anchored}
					if !publicOK || public != wantPublic {
						t.Fatalf("series %d k=%d public=%+v/%v, want %+v", span.Series, k, public, publicOK, wantPublic)
					}
				}
			}
			for series, members := range reported {
				minimum, maximum := 1<<30, -1<<30
				for member := range members {
					if member < minimum {
						minimum = member
					}
					if member > maximum {
						maximum = member
					}
				}
				if maximum-minimum+1 != len(members) {
					t.Fatalf("series %d members are not consecutive: %d..%d with %d entries",
						series, minimum, maximum, len(members))
				}
				window, anchored := windows[series]
				if !anchored {
					continue
				}
				if minimum < 1 || maximum > window.count {
					t.Fatalf("series %d covers member %d..%d outside the anchor window 1..%d",
						series, minimum, maximum, window.count)
				}
				if minimum == 1 && maximum == window.count {
					continue
				}
				// 4 号头部连续三个回次缺测，最后一个正是目录成员 1；其余序列必须铺满窗口。
				// Lunar series 4 catalog member 1 sits on a return the library's lunar
				// detection cannot resolve; every other anchored series must be complete.
				if phase == 1 && series == 4 && minimum == 2 && maximum == window.count {
					continue
				}
				t.Fatalf("series %d covers member %d..%d of the anchor window 1..%d",
					series, minimum, maximum, window.count)
			}
		})
	}
}

// override 声明的成员区间必须真的可达：首成员与末成员都要按格点报出，且标注为已核验。
func TestSarosOverrideWindowsCoverDeclaredMemberRange(t *testing.T) {
	for _, phase := range []int{0, 1} {
		overrides := solarSarosHeadOverrides[:]
		if phase == 1 {
			overrides = lunarSarosHeadOverrides[:]
		}
		cycleDays := float64(sarosTestAnchorMemberInterval) * sarosTestSynodicMonthDays
		for _, override := range overrides {
			headTT := basic.JDCalc(int(override.HeadYear), int(override.HeadMonth), float64(override.HeadDay))
			for _, member := range []int{1, int(override.Count)} {
				ttJDE := headTT + float64(member-1-int(override.MemberOffset))*cycleDays
				info, ok := sarosTestInfo(ttJDE, phase)
				want := SarosInfo{
					Series:   int(override.Series),
					Member:   member,
					Count:    int(override.Count),
					Verified: true,
				}
				if !ok || info != want {
					t.Fatalf("phase=%d series %d member %d = %+v/%v, want %+v",
						phase, override.Series, member, info, ok, want)
				}
			}
		}
	}
}

// 锚表 Count 的裁剪只作用于协调输入：扩展表派发结果与"原始跨度 + 协调"必须逐字段相同。
func TestSarosAnchorCountClipKeepsReconciledResult(t *testing.T) {
	for _, phase := range []int{0, 1} {
		spans := sarosTestSpans(phase)
		for _, span := range spans {
			for index, k := 0, span.First; k <= span.Last; k, index = k+sarosTestAnchorMemberInterval, index+1 {
				ttJDE := sarosTestReturnTT(k, phase)
				if !sarosTestInsideTable(ttJDE) {
					continue
				}
				if index != 0 && k != span.Last {
					continue
				}
				raw, rawOK := matchSarosSpans(spans, k)
				want, wantOK := reconcileSarosSpan(raw, ttJDE, phase)
				if !rawOK {
					wantOK = false
				}
				got, ok := reconciledSarosInfo(extendedSarosInfo, ttJDE, phase)
				if ok != wantOK || got != want {
					t.Fatalf("phase=%d k=%d dispatch=%+v/%v, raw reconcile=%+v/%v", phase, k, got, ok, want, wantOK)
				}
			}
		}
	}
}

// Anchored series keep the catalog numbering, so a provisional span may no
// longer present a second "member 1" for the same series.
func TestLunarSaros_ProvisionalSpanDoesNotDuplicateAnchoredMember(t *testing.T) {
	head, ok := eclipseOnDateForTest(-1926, 2, 2)
	if !ok {
		t.Skip("anchored series head is not an eclipse in this build")
	}
	if head.Saros.Series != 22 || head.Saros.Member != 1 {
		t.Fatalf("series 22 head = %d/%d, want 22/1", head.Saros.Series, head.Saros.Member)
	}
	earlier, ok := eclipseOnDateForTest(-1944, 1, 22)
	if !ok {
		t.Fatal("expected a lunar eclipse one Saros before the anchored head")
	}
	if earlier.HasSaros && earlier.Saros.Series == head.Saros.Series && earlier.Saros.Member == head.Saros.Member {
		t.Fatalf("provisional span repeats series %d member %d", earlier.Saros.Series, earlier.Saros.Member)
	}
}

// Modern catalog values must keep matching NASA's published member numbers.
func TestLunarSaros_ModernMembersUnchanged(t *testing.T) {
	cases := []struct {
		year, month, day int
		series, member   int
	}{
		{2000, 7, 16, 129, 37},
		{2018, 7, 27, 129, 38},
		{2022, 11, 8, 136, 20},
		{2025, 9, 7, 128, 41},
	}
	for _, testCase := range cases {
		info, ok := eclipseOnDateForTest(testCase.year, testCase.month, testCase.day)
		if !ok {
			t.Fatalf("missing lunar eclipse on %04d-%02d-%02d", testCase.year, testCase.month, testCase.day)
		}
		if info.Saros.Series != testCase.series || info.Saros.Member != testCase.member {
			t.Fatalf("%04d-%02d-%02d saros = %d/%d, want %d/%d",
				testCase.year, testCase.month, testCase.day,
				info.Saros.Series, info.Saros.Member, testCase.series, testCase.member)
		}
	}
}

func eclipseOnDateForTest(year, month, day int) (LunarEclipseInfo, bool) {
	return LunarEclipseOnDate(time.Date(year, time.Month(month), day, 12, 0, 0, 0, time.UTC))
}
