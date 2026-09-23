package calendar

import (
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

// L2 契约：儒略历独有闰日（1582 年前百年非 400 闰年的 2 月 29 日）的 Solar 取后继日作为规范日，
// JD 取精确儒略日，JulianOnly 为真，农历身份字段描述这一天本身。
var julianOnlyYears = []int{
	100, 200, 300, 500, 600, 700, 900, 1000, 1100, 1300, 1400, 1500,
	-700, -600, -500, -300, -200, -100,
}

// lunarIdentity 是农历记录的身份字段，用于比较正向与反向是否描述同一天。
type lunarIdentity struct {
	year, month, day int
	leap             bool
	desc             string
	ganzhiDay        string
	ganzhiMonth      string
	calendars        int
}

func identityOf(t Time) lunarIdentity {
	l := t.Lunar()
	return lunarIdentity{
		year:        l.LunarYear(),
		month:       l.LunarMonth(),
		day:         l.LunarDay(),
		leap:        l.IsLeap(),
		desc:        l.MonthDay(),
		ganzhiDay:   l.GanZhiDay(),
		ganzhiMonth: l.GanZhiMonth(),
		calendars:   len(t.LunarInfo()),
	}
}

func TestJulianOnlyLeapDayForwardReverseAgree(t *testing.T) {
	for _, year := range julianOnlyYears {
		phantom := basic.JDCalc(year, 2, 29)
		if phantom != phantom { // NaN 保护：这些年份必须被历法层承认
			t.Fatalf("year %d: basic.JDCalc(%d,2,29) is NaN", year, year)
		}
		canonicalDate := basic.JD2DateByZone(phantom, time.FixedZone("CST", 8*3600), true)
		fwd, err := SolarToLunarByYMD(year, 2, 29)
		if err != nil {
			t.Fatalf("year %d: SolarToLunarByYMD(%d,2,29) error: %v", year, year, err)
		}
		if !fwd.JulianOnly() || fwd.JD() != phantom {
			t.Errorf("year %d: forward JulianOnly=%v JD=%.1f, want true %.1f",
				year, fwd.JulianOnly(), fwd.JD(), phantom)
		}
		if got := fwd.Solar().Format("2006-01-02"); got != canonicalDate.Format("2006-01-02") {
			t.Errorf("year %d: forward Solar=%s, want canonical day %s", year, got, canonicalDate.Format("2006-01-02"))
		}
		want := identityOf(fwd)

		rev, err := LunarToSolarByYMD(want.year, want.month, want.day, want.leap)
		if err != nil {
			t.Fatalf("year %d: LunarToSolarByYMD(%d,%d,%d,%v) error: %v",
				year, want.year, want.month, want.day, want.leap, err)
		}
		if !rev.JulianOnly() || rev.JD() != phantom {
			t.Errorf("year %d: reverse JulianOnly=%v JD=%.1f, want true %.1f",
				year, rev.JulianOnly(), rev.JD(), phantom)
		}
		if got := rev.Solar().Format("2006-01-02"); got != canonicalDate.Format("2006-01-02") {
			t.Errorf("year %d: reverse Solar=%s, want canonical day %s", year, got, canonicalDate.Format("2006-01-02"))
		}
		if got := identityOf(rev); got != want {
			t.Errorf("year %d: reverse identity %+v, want %+v (requested lunar %d/%d/%d leap=%v)",
				year, got, want, want.year, want.month, want.day, want.leap)
		}
		info := rev.LunarInfo()[0]
		if !info.JulianOnly || info.JD != phantom {
			t.Errorf("year %d: LunarInfo JulianOnly=%v JD=%.1f, want true %.1f", year, info.JulianOnly, info.JD, phantom)
		}
		// 日干支必须取自闰日自己的儒略日，而不是规范日 Solar()。
		// The sexagenary day name must come from the leap day's own Julian day, not from the canonical day.
		if got, canonicalGanzhi := rev.Lunar().GanZhiDay(), GanZhiOfDay(rev.Solar()); got != ganZhiOfJD(phantom) {
			t.Errorf("year %d: GanZhiDay=%s, want %s (canonical day %s gives %s)",
				year, got, ganZhiOfJD(phantom), rev.Solar().Format("2006-01-02"), canonicalGanzhi)
		}
	}
}

// 儒略历闰日两侧仍是普通日子：往返正常且农历日不与闰日重复。
func TestJulianOnlyLeapDayNeighboursStayOrdinary(t *testing.T) {
	for _, year := range julianOnlyYears {
		phantom := basic.JDCalc(year, 2, 29)
		fwd, err := SolarToLunarByYMD(year, 2, 29)
		if err != nil {
			t.Fatalf("year %d: SolarToLunarByYMD(%d,2,29) error: %v", year, year, err)
		}
		leapLunar := identityOf(fwd)
		for _, c := range []struct {
			month, day int
			wantJD     float64
		}{
			{2, 28, phantom - 1},
			{3, 1, phantom + 1},
		} {
			near, err := SolarToLunarByYMD(year, c.month, c.day)
			if err != nil {
				t.Fatalf("year %d: SolarToLunarByYMD(%d,%d,%d) error: %v", year, year, c.month, c.day, err)
			}
			nearID := identityOf(near)
			if nearID == leapLunar {
				t.Errorf("year %d: %d-%02d shares the leap day's lunar date %d/%d/%d leap=%v",
					year, year, c.day, nearID.year, nearID.month, nearID.day, nearID.leap)
			}
			rev, err := LunarToSolarByYMD(nearID.year, nearID.month, nearID.day, nearID.leap)
			if err != nil {
				t.Fatalf("year %d: reverse of %d-%02d lunar %d/%d/%d error: %v",
					year, year, c.day, nearID.year, nearID.month, nearID.day, err)
			}
			if rev.JulianOnly() {
				t.Errorf("year %d: %d-%02d flagged JulianOnly", year, year, c.day)
			}
			if rev.JD() != c.wantJD {
				t.Errorf("year %d: %d-%02d JD=%.1f, want %.1f", year, year, c.day, rev.JD(), c.wantJD)
			}
			if got := identityOf(rev); got != nearID {
				t.Errorf("year %d: %d-%02d reverse identity %+v, want %+v", year, year, c.day, got, nearID)
			}
		}
	}
}

// julianOnlyCivilDay 的分类必须只覆盖"仅儒略历存在"的闰日。
func TestJulianOnlyCivilDayClassification(t *testing.T) {
	for _, year := range append(append([]int{}, julianOnlyYears...), 400, 800, 1200, 1600, 2000) {
		want := false
		for _, v := range julianOnlyYears {
			if v == year {
				want = true
				break
			}
		}
		if got := julianOnlyCivilDay(year, 2, 29); got != want {
			t.Errorf("julianOnlyCivilDay(%d,2,29)=%v, want %v", year, got, want)
		}
	}
	// 不存在的日期与普通闰日都不属于这一类。
	for _, c := range []struct{ y, m, d int }{
		{1900, 2, 29}, {1582, 2, 29}, {2024, 2, 29}, {2023, 2, 29}, {100, 2, 28}, {100, 3, 1},
	} {
		if julianOnlyCivilDay(c.y, c.m, c.d) {
			t.Errorf("julianOnlyCivilDay(%d,%d,%d)=true, want false", c.y, c.m, c.d)
		}
	}
}

// 普通日期不受 L2 影响：JulianOnly 为假、JD 与 Solar 一致、身份与正向一致。
func TestOrdinaryDatesCarryNoJulianOnlyFlag(t *testing.T) {
	cases := []struct{ y, m, d int }{
		{2025, 1, 1}, {2025, 6, 15}, {1912, 2, 12}, {1900, 5, 5}, {1582, 10, 15},
		{1000, 7, 7}, {700, 3, 2}, {500, 3, 1}, {1, 1, 1}, {0, 6, 1}, {-100, 3, 2}, {-500, 2, 28},
	}
	for _, c := range cases {
		t.Run(time.Date(c.y, time.Month(c.m), c.d, 0, 0, 0, 0, time.UTC).Format("2006-01-02"), func(t *testing.T) {
			fwd, err := SolarToLunarByYMD(c.y, c.m, c.d)
			if err != nil {
				t.Fatalf("SolarToLunarByYMD(%d,%d,%d) error: %v", c.y, c.m, c.d, err)
			}
			id := identityOf(fwd)
			rev, err := LunarToSolarByYMD(id.year, id.month, id.day, id.leap)
			if err != nil {
				t.Fatalf("reverse of %d/%d/%d leap=%v error: %v", id.year, id.month, id.day, id.leap, err)
			}
			if rev.JulianOnly() {
				t.Errorf("%d-%02d-%02d flagged JulianOnly", c.y, c.m, c.d)
			}
			if want := Date2JD(rev.Solar()); rev.JD() != want {
				t.Errorf("%d-%02d-%02d JD=%.1f, want %.1f (Solar 推出值)", c.y, c.m, c.d, rev.JD(), want)
			}
			if got := identityOf(rev); got != id {
				t.Errorf("%d-%02d-%02d reverse identity %+v, want %+v", c.y, c.m, c.d, got, id)
			}
			info := rev.LunarInfo()
			for _, v := range info {
				if v.JulianOnly {
					t.Errorf("%d-%02d-%02d LunarInfo flagged JulianOnly", c.y, c.m, c.d)
				}
				if v.JD != rev.JD() {
					t.Errorf("%d-%02d-%02d LunarInfo JD=%.1f, want %.1f", c.y, c.m, c.d, v.JD, rev.JD())
				}
			}
		})
	}
}

// 字符串入口 LunarToSolar 曾把儒略历独有闰日整体偏一天（先解析成 time.Time 再回代）；
// 现在按"结果紧邻闰日且闰日记录正是描述所指"复核。用例含武则天年号写法。
func TestLunarToSolarTextHandlesJulianOnlyLeapDays(t *testing.T) {
	cases := []struct {
		desc  string
		year  int
		month int
		day   int
		leap  bool
	}{
		{"100年二月初二", 100, 2, 2, false},
		{"200年正月廿七", 200, 1, 27, false},
		{"300年正月廿三", 300, 1, 23, false},
		{"500年正月十五", 500, 1, 15, false},
		{"600年闰正月初十", 600, 1, 10, true},
		{"700年四月初五", 700, 4, 5, false},
		{"900年正月廿六", 900, 1, 26, false},
		{"1000年正月廿二", 1000, 1, 22, false},
		{"1100年正月十八", 1100, 1, 18, false},
		{"1300年二月初九", 1300, 2, 9, false},
		{"1400年二月初五", 1400, 2, 5, false},
		{"1500年二月初一", 1500, 2, 1, false},
		{"前701年二月初八", -700, 2, 8, false},
		{"前601年三月初三", -600, 3, 3, false},
		{"前501年三月廿八", -500, 3, 28, false},
		{"前301年三月二十", -300, 3, 20, false},
		{"前201年正月十五", -200, 1, 15, false},
		{"前101年正月十一", -100, 1, 11, false},
		{"圣历三年二月初五", 700, 4, 5, false},
		{"久视元年二月初五", 700, 4, 5, false},
		{"永元十二年二月初二", 100, 2, 2, false},
	}
	for _, c := range cases {
		phantom := basic.JDCalc(c.year, 2, 29)
		canonicalDate := basic.JD2DateByZone(phantom, time.FixedZone("CST", 8*3600), true)
		results, err := LunarToSolar(c.desc)
		if err != nil {
			t.Fatalf("%s: LunarToSolar error: %v", c.desc, err)
		}
		if len(results) != 1 {
			t.Errorf("%s: got %d results, want 1", c.desc, len(results))
			continue
		}
		got := results[0]
		if !got.JulianOnly() || got.JD() != phantom {
			t.Errorf("%s: JulianOnly=%v JD=%.1f, want true %.1f", c.desc, got.JulianOnly(), got.JD(), phantom)
		}
		if solar := got.Solar().Format("2006-01-02"); solar != canonicalDate.Format("2006-01-02") {
			t.Errorf("%s: Solar=%s, want canonical day %s", c.desc, solar, canonicalDate.Format("2006-01-02"))
		}
		lunar := got.Lunar()
		if lunar.LunarYear() != c.year || lunar.LunarMonth() != c.month || lunar.LunarDay() != c.day || lunar.IsLeap() != c.leap {
			t.Errorf("%s: lunar %d/%d/%d leap=%v, want %d/%d/%d leap=%v", c.desc,
				lunar.LunarYear(), lunar.LunarMonth(), lunar.LunarDay(), lunar.IsLeap(), c.year, c.month, c.day, c.leap)
		}
	}
}

// 闰日前后的描述必须仍是普通日期。
func TestLunarToSolarTextKeepsLeapDayNeighboursOrdinary(t *testing.T) {
	for _, year := range julianOnlyYears {
		phantom := basic.JDCalc(year, 2, 29)
		for _, offset := range []int{-1, 1} {
			// 取闰日前后各一天，用它们自己的农历日构造描述。
			neighbour := basic.JD2DateByZone(phantom+float64(offset), time.FixedZone("CST", 8*3600), true)
			fwd, err := SolarToLunarByYMD(neighbour.Year(), int(neighbour.Month()), neighbour.Day())
			if err != nil {
				t.Fatalf("year %d: SolarToLunarByYMD(%s) error: %v", year, neighbour.Format("2006-01-02"), err)
			}
			id := identityOf(fwd)
			prefix := ""
			if id.year < 0 {
				prefix = "前" + number2Chinese(1-id.year, true)
			} else {
				prefix = number2Chinese(id.year, true)
			}
			desc := prefix + "年" + formatLunarDateString(id.month, id.day, id.leap, 0)
			results, err := LunarToSolar(desc)
			if err != nil {
				t.Fatalf("%s: LunarToSolar error: %v", desc, err)
			}
			if len(results) != 1 {
				t.Errorf("%s: got %d results, want 1", desc, len(results))
				continue
			}
			got := results[0]
			if got.JulianOnly() {
				t.Errorf("%s: unexpectedly flagged JulianOnly", desc)
			}
			if want := phantom + float64(offset); got.JD() != want {
				t.Errorf("%s: JD=%.1f, want %.1f", desc, got.JD(), want)
			}
			if got := identityOf(results[0]); got != id {
				t.Errorf("%s: identity %+v, want %+v", desc, got, id)
			}
		}
	}
}
