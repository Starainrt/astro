package eclipse

import (
	"math"
	"sync"

	"github.com/starainrt/astro/basic"
)

const (
	sarosCycleLunations = 223
	sarosCycleDays      = float64(sarosCycleLunations) * solarEclipseSynodicMonthDays
	// The generated tables show at most 90 members in one series. A window of
	// 90 returns therefore covers a complete series without evaluating remote
	// samples whose ancient ephemeris drift can reject an otherwise valid event.
	sarosExtrapolationWindow = 90

	sarosMagicYearOffset    = 3000
	sarosMagicCountMask     = 0x7f
	sarosMagicDayMask       = 0x1f
	sarosMagicMonthMask     = 0x0f
	sarosMagicYearMask      = 0x1fff
	sarosMagicCountShift    = 0
	sarosMagicDayShift      = 7
	sarosMagicMonthShift    = 12
	sarosMagicYearShift     = 16
	sarosMagicMatchLimitDay = 12.0
	sarosMagicTieEpsilonDay = 1e-9
)

// SarosInfo 沙罗序列信息, Saros series metadata.
type SarosInfo struct {
	// Series 是权威目录系列编号；Verified=false 时为根据周期外推得到的暂定编号。
	// Series is the authoritative catalog series number, or a provisional number when Verified is false.
	Series int
	// Member 是本次食在该系列中的序号，从 1 开始计数。
	// Member is the 1-based index of this eclipse within the series.
	Member int
	// Count 是该系列的总成员数。
	// Count is the total number of eclipses in the series.
	Count int
	// Verified 表示系列头和成员范围已与内置权威目录锚点核验。
	// Verified reports whether the series head and member range match an authoritative catalog anchor.
	Verified bool
}

// sarosAnchorRange is the authoritative member window of one series as
// published in the anchor catalog. Extended (provisional) tables may enumerate
// one or two extra members at the head and tail of a series, which collides
// with the anchored numbering: the same member index then appears twice and the
// series count changes mid-series. Reconciling against the anchor keeps the
// member numbering and the total count consistent for every event.
type sarosAnchorRange struct {
	headTT float64
	count  int
	offset int
}

var (
	sarosAnchorRangesOnce sync.Once
	sarosAnchorRanges     [2]map[int]sarosAnchorRange
)

func sarosAnchorRangeFor(phase, series int) (sarosAnchorRange, bool) {
	sarosAnchorRangesOnce.Do(buildSarosAnchorRanges)
	if phase < 0 || phase >= len(sarosAnchorRanges) {
		return sarosAnchorRange{}, false
	}
	value, ok := sarosAnchorRanges[phase][series]
	return value, ok
}

func buildSarosAnchorRanges() {
	anchors := solarSarosAnchors[:]
	base := 0
	overrides := solarSarosHeadOverrides[:]
	for phase := 0; phase < 2; phase++ {
		if phase == 1 {
			anchors = lunarSarosAnchors[:]
			base = 1
			overrides = lunarSarosHeadOverrides[:]
		}
		ranges := make(map[int]sarosAnchorRange, len(anchors)+len(overrides))
		for index, magic := range anchors {
			anchor := decodeSarosMagic(magic, base+index)
			headTT := basic.JDECalc(int(anchor.Year), int(anchor.Month), float64(anchor.Day))
			if math.IsNaN(headTT) {
				continue
			}
			ranges[int(anchor.Series)] = sarosAnchorRange{headTT: headTT, count: int(anchor.Count)}
		}
		for _, override := range overrides {
			headTT := basic.JDECalc(int(override.HeadYear), int(override.HeadMonth), float64(override.HeadDay))
			if math.IsNaN(headTT) {
				continue
			}
			ranges[int(override.Series)] = sarosAnchorRange{
				headTT: headTT,
				count:  int(override.Count),
				offset: int(override.MemberOffset),
			}
		}
		sarosAnchorRanges[phase] = ranges
	}
}

// reconcileSarosSpan anchors a provisional member to the authoritative series
// window. Members outside that window are not part of the catalog series, so
// they report no Saros metadata instead of a conflicting provisional number.
// The member index uses the same relation as the anchor matcher, which keeps a
// provisional event and an anchored event of one series in step.
func reconcileSarosSpan(info SarosInfo, ttJDE float64, phase int) (SarosInfo, bool) {
	anchor, ok := sarosAnchorRangeFor(phase, info.Series)
	if !ok {
		return info, true
	}
	member := int(math.Round((ttJDE-anchor.headTT)/sarosCycleDays)) + 1 + anchor.offset
	if member < 1 || member > anchor.count {
		return SarosInfo{}, false
	}
	info.Member = member
	info.Count = anchor.count
	return info, true
}

type sarosMagic uint32

type sarosAnchor struct {
	Series int16
	Count  uint8
	Year   int16
	Month  uint8
	Day    uint8
}

type sarosHeadOverride struct {
	Series       int16
	Count        uint8
	HeadYear     int16
	HeadMonth    uint8
	HeadDay      uint8
	MemberOffset int8
}

var solarSarosHeadOverrides = [...]sarosHeadOverride{
	{Series: 22, Count: 71, HeadYear: -2192, HeadMonth: 5, HeadDay: 17, MemberOffset: -1},
}

var lunarSarosHeadOverrides = [...]sarosHeadOverride{
	// 4 号锚表头提前了一个分支：目录成员 1 落在本模型认不出的浅半影回次上，
	// 头按目录记作第 3 个成员，尾部最后一场才与目录的 Count 对齐。
	{Series: 4, Count: 78, HeadYear: -2483, HeadMonth: 1, HeadDay: 12, MemberOffset: 2},
	{Series: 8, Count: 86, HeadYear: -2494, HeadMonth: 8, HeadDay: 7, MemberOffset: 0},
	{Series: 61, Count: 78, HeadYear: -762, HeadMonth: 12, HeadDay: 24, MemberOffset: 1},
}

func solarSarosInfo(ttJDE float64) (SarosInfo, bool) {
	if info, ok := matchSarosMagic(solarSarosAnchors[:], 0, solarSarosHeadOverrides[:], ttJDE); ok {
		return info, true
	}
	return reconciledSarosInfo(extendedSarosInfo, ttJDE, 0)
}

func lunarSarosInfo(ttJDE float64) (SarosInfo, bool) {
	if info, ok := matchSarosMagic(lunarSarosAnchors[:], 1, lunarSarosHeadOverrides[:], ttJDE); ok {
		return info, true
	}
	return reconciledSarosInfo(extendedSarosInfo, ttJDE, 1)
}

// reconciledSarosInfo applies the anchor window to a provisional lookup so the
// public metadata never mixes two numbering baselines for one series.
func reconciledSarosInfo(
	provisional func(float64, int) (SarosInfo, bool),
	ttJDE float64,
	phase int,
) (SarosInfo, bool) {
	info, ok := provisional(ttJDE, phase)
	if !ok {
		return SarosInfo{}, false
	}
	return reconcileSarosSpan(info, ttJDE, phase)
}

func matchSarosMagic(anchors []sarosMagic, seriesBase int, overrides []sarosHeadOverride, ttJDE float64) (SarosInfo, bool) {
	if info, ok := matchSarosMagicOverrides(overrides, ttJDE); ok {
		return info, true
	}
	// 带 override 的序列只认 override 头，与 buildSarosAnchorRanges 的覆盖语义一致。
	bestDistance := math.Inf(1)
	best := SarosInfo{}
	for index, magic := range anchors {
		anchor := decodeSarosMagic(magic, seriesBase+index)
		if sarosSeriesOverridden(overrides, int(anchor.Series)) {
			continue
		}
		info, distance, ok := matchSarosMagicCandidate(ttJDE, anchor, 0)
		if !ok {
			continue
		}
		if betterSarosMagicMatch(info, distance, best, bestDistance) {
			bestDistance = distance
			best = info
		}
	}
	if bestDistance <= sarosMagicMatchLimitDay {
		return best, true
	}
	return SarosInfo{}, false
}

// sarosSeriesOverridden 报告该序列是否带有 override 头。
func sarosSeriesOverridden(overrides []sarosHeadOverride, series int) bool {
	for _, override := range overrides {
		if int(override.Series) == series {
			return true
		}
	}
	return false
}

func matchSarosMagicOverrides(overrides []sarosHeadOverride, ttJDE float64) (SarosInfo, bool) {
	bestDistance := math.Inf(1)
	best := SarosInfo{}
	for _, override := range overrides {
		anchor := sarosAnchor{
			Series: override.Series,
			Count:  override.Count,
			Year:   override.HeadYear,
			Month:  override.HeadMonth,
			Day:    override.HeadDay,
		}
		info, distance, ok := matchSarosMagicCandidate(ttJDE, anchor, int(override.MemberOffset))
		if !ok {
			continue
		}
		if betterSarosMagicMatch(info, distance, best, bestDistance) {
			bestDistance = distance
			best = info
		}
	}
	if bestDistance <= sarosMagicMatchLimitDay {
		return best, true
	}
	return SarosInfo{}, false
}

func matchSarosMagicCandidate(ttJDE float64, anchor sarosAnchor, memberOffset int) (SarosInfo, float64, bool) {
	headTT := basic.JDECalc(int(anchor.Year), int(anchor.Month), float64(anchor.Day))
	if math.IsNaN(headTT) {
		return SarosInfo{}, 0, false
	}
	member := int(math.Round((ttJDE-headTT)/sarosCycleDays)) + 1 + memberOffset
	if member < 1 || member > int(anchor.Count) {
		return SarosInfo{}, 0, false
	}
	expectedTT := headTT + float64(member-1-memberOffset)*sarosCycleDays
	return SarosInfo{
		Series:   int(anchor.Series),
		Member:   member,
		Count:    int(anchor.Count),
		Verified: true,
	}, math.Abs(ttJDE - expectedTT), true
}

func betterSarosMagicMatch(info SarosInfo, distance float64, best SarosInfo, bestDistance float64) bool {
	if distance < bestDistance-sarosMagicTieEpsilonDay {
		return true
	}
	if math.Abs(distance-bestDistance) > sarosMagicTieEpsilonDay {
		return false
	}
	if info.Series != best.Series {
		return info.Series < best.Series
	}
	return info.Member < best.Member
}

func decodeSarosMagic(magic sarosMagic, series int) sarosAnchor {
	value := uint32(magic)
	return sarosAnchor{
		Series: int16(series),
		Count:  uint8((value >> sarosMagicCountShift) & sarosMagicCountMask),
		Year:   int16(int((value>>sarosMagicYearShift)&sarosMagicYearMask) - sarosMagicYearOffset),
		Month:  uint8((value >> sarosMagicMonthShift) & sarosMagicMonthMask),
		Day:    uint8((value >> sarosMagicDayShift) & sarosMagicDayMask),
	}
}
