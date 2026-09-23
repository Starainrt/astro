package eclipse

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

const localLunarEclipseSearchLimit = 6000

type localLunarEclipseQueryMode int

const (
	localLunarEclipseQueryVisible localLunarEclipseQueryMode = iota
	localLunarEclipseQueryGeometric
)

// LocalLunarEclipseVisibility 站点在本次月食中的可见性类别 / how much of the eclipse stays above the local horizon.
type LocalLunarEclipseVisibility string

const (
	// LocalLunarEclipseInvisible 食内月亮始终在地平线下 / the Moon never clears the local horizon.
	LocalLunarEclipseInvisible LocalLunarEclipseVisibility = "invisible"
	// LocalLunarEclipseMoonrise 带食月出：食始时月亮尚未升起 / moonrise falls inside the eclipse.
	LocalLunarEclipseMoonrise LocalLunarEclipseVisibility = "moonrise"
	// LocalLunarEclipseMoonset 带食月落：食终时月亮已经落下 / moonset falls inside the eclipse.
	LocalLunarEclipseMoonset LocalLunarEclipseVisibility = "moonset"
	// LocalLunarEclipseRiseAndSet 食内先月出后月落：半影首尾都在地平线下 / both a moonrise and a moonset fall inside the eclipse.
	LocalLunarEclipseRiseAndSet LocalLunarEclipseVisibility = "rise-and-set"
	// LocalLunarEclipseInterrupted 半影首尾可见，中途却落到地平线下 / above the horizon at both penumbral contacts but not in between.
	LocalLunarEclipseInterrupted LocalLunarEclipseVisibility = "interrupted"
	// LocalLunarEclipseFull 半影全程可见 / the Moon stays above the horizon for the whole eclipse.
	LocalLunarEclipseFull LocalLunarEclipseVisibility = "full"
	// LocalLunarEclipsePenumbraMoonrise 仅见半影的月出：月出落在食内，本影阶段整段在地平线下 / moonrise inside the eclipse with the whole umbral phase below the horizon.
	LocalLunarEclipsePenumbraMoonrise LocalLunarEclipseVisibility = "penumbra-moonrise"
	// LocalLunarEclipsePenumbraMoonset 仅见半影的月落：月落落在食内，本影阶段整段在地平线下 / moonset inside the eclipse with the whole umbral phase below the horizon.
	LocalLunarEclipsePenumbraMoonset LocalLunarEclipseVisibility = "penumbra-moonset"
)

// LocalLunarEclipseInfo 站点月食信息, local lunar eclipse information.
//
// 所有时刻字段都保持用户输入的时区。
// 不存在的阶段使用零值 time.Time。
type LocalLunarEclipseInfo struct {
	// Type 月食类型, eclipse type.
	Type LunarEclipseType
	// HasSaros 存在沙罗序列信息（可能是锚点外推结果）, has Saros series metadata (possibly extrapolated).
	HasSaros bool
	// Saros 是沙罗序列信息，包括系列号、系列内序号和总成员数。
	// Saros is Saros series metadata with the series number, member index, and total member count.
	Saros SarosInfo

	// Longitude 观测点经度，东正西负, observer longitude, east positive.
	Longitude float64
	// Latitude 观测点纬度，北正南负, observer latitude, north positive.
	Latitude float64
	// Height 观测点椭球高（大地高），单位米；只有正高 H 时须先加大地水准面差距 N, observer ellipsoidal height in meters.
	Height float64

	// PenumbralMagnitude 半影食分, penumbral magnitude.
	PenumbralMagnitude float64
	// UmbralMagnitude 本影食分；纯半影月食时可为负值, umbral magnitude; can be negative for purely penumbral eclipses.
	UmbralMagnitude float64

	// PenumbralStart 半影始, penumbral eclipse begins.
	PenumbralStart time.Time
	// PartialStart 初亏, partial eclipse begins.
	PartialStart time.Time
	// TotalStart 食既, total eclipse begins.
	TotalStart time.Time
	// Maximum 食甚, greatest eclipse.
	Maximum time.Time
	// TotalEnd 生光, total eclipse ends.
	TotalEnd time.Time
	// PartialEnd 复圆, partial eclipse ends.
	PartialEnd time.Time
	// PenumbralEnd 半影终, penumbral eclipse ends.
	PenumbralEnd time.Time

	// MoonAltitude 食甚时月亮高度角，单位度, Moon altitude at maximum in degrees.
	MoonAltitude float64
	// MoonAzimuth 食甚时月亮方位角，单位度, Moon azimuth at maximum in degrees.
	MoonAzimuth float64
	// VisibleAtMaximum 食甚时月亮中心在本地几何地平线上方, Moon center above the local geometric horizon at maximum.
	VisibleAtMaximum bool
	// Visibility 本次月食在站点的可见性类别；月出/月落落在食内且本影阶段整段在地平线下时取
	// penumbra-moonrise / penumbra-moonset（只看到半影）/ how much of the eclipse is above the local
	// horizon; of the umbral phase stays below it, moonrise and moonset are reported as
	// penumbra-moonrise and penumbra-moonset.
	Visibility LocalLunarEclipseVisibility

	// HasPenumbral 有半影阶段, has penumbral phase.
	HasPenumbral bool
	// HasPartial 有偏食阶段, has partial phase.
	HasPartial bool
	// HasTotal 有全食阶段, has total phase.
	HasTotal bool
}

// LocalLunarEclipseOnDate 当地可见月食查询 / local visible lunar eclipse query.
// Determine whether a visible local lunar eclipse occurs on the local date, using Danjon by default.
func LocalLunarEclipseOnDate(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool) {
	return LocalLunarEclipseOnDateDanjon(date, lon, lat, height)
}

// LocalLunarEclipseOnDateDanjon 当地可见月食查询（Danjon） / local visible lunar eclipse query with Danjon model.
// Determine whether a visible local lunar eclipse occurs on the local date with the Danjon model.
func LocalLunarEclipseOnDateDanjon(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool) {
	return localLunarEclipseOnDate(date, lon, lat, height, basic.LunarEclipseDanjon, localLunarEclipseQueryVisible)
}

// LocalLunarEclipseOnDateChauvenet 当地可见月食查询（Chauvenet） / local visible lunar eclipse query with Chauvenet model.
// Determine whether a visible local lunar eclipse occurs on the local date with the Chauvenet model.
func LocalLunarEclipseOnDateChauvenet(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool) {
	return localLunarEclipseOnDate(date, lon, lat, height, basic.LunarEclipseChauvenet, localLunarEclipseQueryVisible)
}

// GeometricLocalLunarEclipseOnDate 当地几何月食查询 / local geometric lunar eclipse query.
// Determine whether a geometric local lunar eclipse occurs on the local date, using Danjon by default.
func GeometricLocalLunarEclipseOnDate(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool) {
	return GeometricLocalLunarEclipseOnDateDanjon(date, lon, lat, height)
}

// GeometricLocalLunarEclipseOnDateDanjon 当地几何月食查询（Danjon） / local geometric lunar eclipse query with Danjon model.
// Determine whether a geometric local lunar eclipse occurs on the local date with the Danjon model.
func GeometricLocalLunarEclipseOnDateDanjon(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool) {
	return localLunarEclipseOnDate(date, lon, lat, height, basic.LunarEclipseDanjon, localLunarEclipseQueryGeometric)
}

// GeometricLocalLunarEclipseOnDateChauvenet 当地几何月食查询（Chauvenet） / local geometric lunar eclipse query with Chauvenet model.
// Determine whether a geometric local lunar eclipse occurs on the local date with the Chauvenet model.
func GeometricLocalLunarEclipseOnDateChauvenet(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool) {
	return localLunarEclipseOnDate(date, lon, lat, height, basic.LunarEclipseChauvenet, localLunarEclipseQueryGeometric)
}

func localLunarEclipseOnDate(
	date time.Time,
	lon, lat, height float64,
	calculator lunarEclipseCalculator,
	mode localLunarEclipseQueryMode,
) (LocalLunarEclipseInfo, bool) {
	location := date.Location()
	dayStart, dayMid, dayEnd := lunarEclipseLocalDayBounds(date)

	phaseDiff := moonSunLoDiff(dayMid)
	if phaseDiff < lunarEclipseDayPhaseMin || phaseDiff > lunarEclipseDayPhaseMax {
		return LocalLunarEclipseInfo{}, false
	}

	candidateTT := basic.CalcMoonSHByJDE(timeToTTJDE(dayMid), 1)
	if !isPotentialLunarEclipse(candidateTT) {
		return LocalLunarEclipseInfo{}, false
	}

	result := calculator(candidateTT)
	if result.Type == basic.LunarEclipseNone {
		return LocalLunarEclipseInfo{}, false
	}

	info := localLunarEclipseInfoFromBasic(result, lon, lat, height, location)
	if !localLunarEclipseOverlapsDate(info, dayStart, dayEnd) {
		return LocalLunarEclipseInfo{}, false
	}
	if mode == localLunarEclipseQueryVisible && !localLunarEclipseVisibleOnDate(info, dayStart, dayEnd) {
		return LocalLunarEclipseInfo{}, false
	}
	return info, true
}

// LastLocalLunarEclipse 上次可见月食 / previous visible local lunar eclipse.
// Previous visible local lunar eclipse, using Danjon by default.
func LastLocalLunarEclipse(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return LastLocalLunarEclipseDanjon(date, lon, lat, height)
}

// LastLocalTotalLunarEclipse 上次可见月全食 / previous visible local total lunar eclipse.
// Previous visible local total lunar eclipse, using Danjon by default.
func LastLocalTotalLunarEclipse(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool) {
	return searchLocalTotalLunarEclipse(date, lon, lat, height, -1, true, basic.LunarEclipseDanjon, localLunarEclipseQueryVisible)
}

// LastLocalLunarEclipseDanjon 上次可见月食（Danjon） / previous visible local lunar eclipse with Danjon model.
// Previous visible local lunar eclipse with the Danjon model.
func LastLocalLunarEclipseDanjon(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	info, _ := searchLocalLunarEclipse(date, lon, lat, height, -1, true, basic.LunarEclipseDanjon, localLunarEclipseQueryVisible)
	return info
}

// LastLocalLunarEclipseChauvenet 上次可见月食（Chauvenet） / previous visible local lunar eclipse with Chauvenet model.
// Previous visible local lunar eclipse with the Chauvenet model.
func LastLocalLunarEclipseChauvenet(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	info, _ := searchLocalLunarEclipse(date, lon, lat, height, -1, true, basic.LunarEclipseChauvenet, localLunarEclipseQueryVisible)
	return info
}

// LastGeometricLocalLunarEclipse 上次几何月食 / previous geometric local lunar eclipse.
// Previous geometric local lunar eclipse, using Danjon by default.
func LastGeometricLocalLunarEclipse(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return LastGeometricLocalLunarEclipseDanjon(date, lon, lat, height)
}

// LastGeometricLocalLunarEclipseDanjon 上次几何月食（Danjon） / previous geometric local lunar eclipse with Danjon model.
// Previous geometric local lunar eclipse with the Danjon model.
func LastGeometricLocalLunarEclipseDanjon(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	info, _ := searchLocalLunarEclipse(date, lon, lat, height, -1, true, basic.LunarEclipseDanjon, localLunarEclipseQueryGeometric)
	return info
}

// LastGeometricLocalLunarEclipseChauvenet 上次几何月食（Chauvenet） / previous geometric local lunar eclipse with Chauvenet model.
// Previous geometric local lunar eclipse with the Chauvenet model.
func LastGeometricLocalLunarEclipseChauvenet(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	info, _ := searchLocalLunarEclipse(date, lon, lat, height, -1, true, basic.LunarEclipseChauvenet, localLunarEclipseQueryGeometric)
	return info
}

// NextLocalLunarEclipse 下次可见月食 / next visible local lunar eclipse.
// Next visible local lunar eclipse, using Danjon by default.
func NextLocalLunarEclipse(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return NextLocalLunarEclipseDanjon(date, lon, lat, height)
}

// NextLocalTotalLunarEclipse 下次可见月全食 / next visible local total lunar eclipse.
// Next visible local total lunar eclipse, using Danjon by default.
func NextLocalTotalLunarEclipse(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool) {
	return searchLocalTotalLunarEclipse(date, lon, lat, height, 1, false, basic.LunarEclipseDanjon, localLunarEclipseQueryVisible)
}

// NextLocalLunarEclipseDanjon 下次可见月食（Danjon） / next visible local lunar eclipse with Danjon model.
// Next visible local lunar eclipse with the Danjon model.
func NextLocalLunarEclipseDanjon(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	info, _ := searchLocalLunarEclipse(date, lon, lat, height, 1, false, basic.LunarEclipseDanjon, localLunarEclipseQueryVisible)
	return info
}

// NextLocalLunarEclipseChauvenet 下次可见月食（Chauvenet） / next visible local lunar eclipse with Chauvenet model.
// Next visible local lunar eclipse with the Chauvenet model.
func NextLocalLunarEclipseChauvenet(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	info, _ := searchLocalLunarEclipse(date, lon, lat, height, 1, false, basic.LunarEclipseChauvenet, localLunarEclipseQueryVisible)
	return info
}

// NextGeometricLocalLunarEclipse 下次几何月食 / next geometric local lunar eclipse.
// Next geometric local lunar eclipse, using Danjon by default.
func NextGeometricLocalLunarEclipse(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return NextGeometricLocalLunarEclipseDanjon(date, lon, lat, height)
}

// NextGeometricLocalLunarEclipseDanjon 下次几何月食（Danjon） / next geometric local lunar eclipse with Danjon model.
// Next geometric local lunar eclipse with the Danjon model.
func NextGeometricLocalLunarEclipseDanjon(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	info, _ := searchLocalLunarEclipse(date, lon, lat, height, 1, false, basic.LunarEclipseDanjon, localLunarEclipseQueryGeometric)
	return info
}

// NextGeometricLocalLunarEclipseChauvenet 下次几何月食（Chauvenet） / next geometric local lunar eclipse with Chauvenet model.
// Next geometric local lunar eclipse with the Chauvenet model.
func NextGeometricLocalLunarEclipseChauvenet(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	info, _ := searchLocalLunarEclipse(date, lon, lat, height, 1, false, basic.LunarEclipseChauvenet, localLunarEclipseQueryGeometric)
	return info
}

// ClosestLocalLunarEclipse 最近一次可见月食 / closest visible local lunar eclipse.
// Closest visible local lunar eclipse, using Danjon by default.
func ClosestLocalLunarEclipse(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return ClosestLocalLunarEclipseDanjon(date, lon, lat, height)
}

// ClosestLocalTotalLunarEclipse 最近一次可见月全食 / closest visible local total lunar eclipse.
// Closest visible local total lunar eclipse, using Danjon by default.
func ClosestLocalTotalLunarEclipse(date time.Time, lon, lat, height float64) (LocalLunarEclipseInfo, bool) {
	last, hasLast := searchLocalTotalLunarEclipse(date, lon, lat, height, -1, true, basic.LunarEclipseDanjon, localLunarEclipseQueryVisible)
	next, hasNext := searchLocalTotalLunarEclipse(date, lon, lat, height, 1, false, basic.LunarEclipseDanjon, localLunarEclipseQueryVisible)
	return closestLocalLunarEclipseResult(date, last, hasLast, next, hasNext)
}

// ClosestLocalLunarEclipseDanjon 最近一次可见月食（Danjon） / closest visible local lunar eclipse with Danjon model.
// Closest visible local lunar eclipse with the Danjon model.
func ClosestLocalLunarEclipseDanjon(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	last, hasLast := searchLocalLunarEclipse(date, lon, lat, height, -1, true, basic.LunarEclipseDanjon, localLunarEclipseQueryVisible)
	next, hasNext := searchLocalLunarEclipse(date, lon, lat, height, 1, false, basic.LunarEclipseDanjon, localLunarEclipseQueryVisible)
	return closestLocalLunarEclipse(date, last, hasLast, next, hasNext)
}

// ClosestLocalLunarEclipseChauvenet 最近一次可见月食（Chauvenet） / closest visible local lunar eclipse with Chauvenet model.
// Closest visible local lunar eclipse with the Chauvenet model.
func ClosestLocalLunarEclipseChauvenet(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	last, hasLast := searchLocalLunarEclipse(date, lon, lat, height, -1, true, basic.LunarEclipseChauvenet, localLunarEclipseQueryVisible)
	next, hasNext := searchLocalLunarEclipse(date, lon, lat, height, 1, false, basic.LunarEclipseChauvenet, localLunarEclipseQueryVisible)
	return closestLocalLunarEclipse(date, last, hasLast, next, hasNext)
}

// ClosestGeometricLocalLunarEclipse 最近一次几何月食 / closest geometric local lunar eclipse.
// Closest geometric local lunar eclipse, using Danjon by default.
func ClosestGeometricLocalLunarEclipse(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	return ClosestGeometricLocalLunarEclipseDanjon(date, lon, lat, height)
}

// ClosestGeometricLocalLunarEclipseDanjon 最近一次几何月食（Danjon） / closest geometric local lunar eclipse with Danjon model.
// Closest geometric local lunar eclipse with the Danjon model.
func ClosestGeometricLocalLunarEclipseDanjon(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	last, hasLast := searchLocalLunarEclipse(date, lon, lat, height, -1, true, basic.LunarEclipseDanjon, localLunarEclipseQueryGeometric)
	next, hasNext := searchLocalLunarEclipse(date, lon, lat, height, 1, false, basic.LunarEclipseDanjon, localLunarEclipseQueryGeometric)
	return closestLocalLunarEclipse(date, last, hasLast, next, hasNext)
}

// ClosestGeometricLocalLunarEclipseChauvenet 最近一次几何月食（Chauvenet） / closest geometric local lunar eclipse with Chauvenet model.
// Closest geometric local lunar eclipse with the Chauvenet model.
func ClosestGeometricLocalLunarEclipseChauvenet(date time.Time, lon, lat, height float64) LocalLunarEclipseInfo {
	last, hasLast := searchLocalLunarEclipse(date, lon, lat, height, -1, true, basic.LunarEclipseChauvenet, localLunarEclipseQueryGeometric)
	next, hasNext := searchLocalLunarEclipse(date, lon, lat, height, 1, false, basic.LunarEclipseChauvenet, localLunarEclipseQueryGeometric)
	return closestLocalLunarEclipse(date, last, hasLast, next, hasNext)
}

func closestLocalLunarEclipse(
	date time.Time,
	last LocalLunarEclipseInfo,
	hasLast bool,
	next LocalLunarEclipseInfo,
	hasNext bool,
) LocalLunarEclipseInfo {
	info, _ := closestLocalLunarEclipseResult(date, last, hasLast, next, hasNext)
	return info
}

func closestLocalLunarEclipseResult(
	date time.Time,
	last LocalLunarEclipseInfo,
	hasLast bool,
	next LocalLunarEclipseInfo,
	hasNext bool,
) (LocalLunarEclipseInfo, bool) {
	switch {
	case hasLast && !hasNext:
		return last, true
	case !hasLast && hasNext:
		return next, true
	case !hasLast && !hasNext:
		return LocalLunarEclipseInfo{}, false
	}

	lastDistance := math.Abs(date.Sub(last.Maximum).Seconds())
	nextDistance := math.Abs(next.Maximum.Sub(date).Seconds())
	if lastDistance <= nextDistance {
		return last, true
	}
	return next, true
}

func searchLocalLunarEclipse(
	date time.Time,
	lon, lat, height float64,
	direction int,
	includeCurrent bool,
	calculator lunarEclipseCalculator,
	mode localLunarEclipseQueryMode,
) (LocalLunarEclipseInfo, bool) {
	targetTT := timeToTTJDE(date)
	candidateTT := basic.CalcMoonSHByJDE(targetTT, 1)

	for i := 0; i < localLunarEclipseSearchLimit; i++ {
		if isPotentialLunarEclipse(candidateTT) {
			result := calculator(candidateTT)
			if result.Type != basic.LunarEclipseNone {
				info := localLunarEclipseInfoFromBasic(result, lon, lat, height, date.Location())
				if (mode != localLunarEclipseQueryVisible || localLunarEclipseVisible(info)) &&
					lunarEclipseMatchesDirection(result.Maximum, targetTT, direction, includeCurrent) {
					return info, true
				}
			}
		}
		candidateTT = nextEclipseSearchCandidateTT(candidateTT, 1, direction, lunarEclipseSynodicMonthDays)
	}

	return LocalLunarEclipseInfo{}, false
}

func searchLocalTotalLunarEclipse(
	date time.Time,
	lon, lat, height float64,
	direction int,
	includeCurrent bool,
	calculator lunarEclipseCalculator,
	mode localLunarEclipseQueryMode,
) (LocalLunarEclipseInfo, bool) {
	targetTT := timeToTTJDE(date)
	candidateTT := basic.CalcMoonSHByJDE(targetTT, 1)

	for i := 0; i < localLunarEclipseSearchLimit; i++ {
		if isPotentialLunarEclipse(candidateTT) {
			result := calculator(candidateTT)
			if result.HasTotal {
				info := localLunarEclipseInfoFromBasic(result, lon, lat, height, date.Location())
				if (mode != localLunarEclipseQueryVisible || localTotalLunarEclipseVisible(info)) &&
					lunarEclipseMatchesDirection(result.Maximum, targetTT, direction, includeCurrent) {
					return info, true
				}
			}
		}
		candidateTT = nextEclipseSearchCandidateTT(candidateTT, 1, direction, lunarEclipseSynodicMonthDays)
	}

	return LocalLunarEclipseInfo{}, false
}

func localLunarEclipseInfoFromBasic(
	result basic.LunarEclipseResult,
	lon, lat, height float64,
	location *time.Location,
) LocalLunarEclipseInfo {
	maximum := ttJDEToTime(result.Maximum, location)
	visibleThreshold := localLunarEclipseVisibilityThreshold(height, lat)
	moonAltitude := lunarAltitude(maximum, lon, lat)
	saros, hasSaros := lunarSarosInfo(result.Maximum)

	info := LocalLunarEclipseInfo{
		HasSaros:           hasSaros,
		Saros:              saros,
		Type:               mapBasicLunarEclipseType(result.Type),
		Longitude:          lon,
		Latitude:           lat,
		Height:             height,
		PenumbralMagnitude: result.PenumbralMagnitude,
		UmbralMagnitude:    result.Magnitude,
		PenumbralStart:     ttJDEToTime(result.PenumbralStart, location),
		PartialStart:       ttJDEToTime(result.PartialStart, location),
		TotalStart:         ttJDEToTime(result.TotalStart, location),
		Maximum:            maximum,
		TotalEnd:           ttJDEToTime(result.TotalEnd, location),
		PartialEnd:         ttJDEToTime(result.PartialEnd, location),
		PenumbralEnd:       ttJDEToTime(result.PenumbralEnd, location),
		MoonAltitude:       moonAltitude,
		MoonAzimuth:        lunarAzimuth(maximum, lon, lat),
		VisibleAtMaximum:   moonAltitude > visibleThreshold,
		HasPenumbral:       result.HasPenumbral,
		HasPartial:         result.HasPartial,
		HasTotal:           result.HasTotal,
	}
	info.Visibility = localLunarEclipseVisibility(info)
	return info
}

func localLunarEclipseOverlapsDate(info LocalLunarEclipseInfo, dayStart, dayEnd time.Time) bool {
	eventStart, eventEnd, ok := localLunarEclipseRange(info)
	if !ok {
		return false
	}
	return !eventEnd.Before(dayStart) && eventStart.Before(dayEnd)
}

func localLunarEclipseRange(info LocalLunarEclipseInfo) (time.Time, time.Time, bool) {
	if !info.HasPenumbral {
		return time.Time{}, time.Time{}, false
	}
	return info.PenumbralStart, info.PenumbralEnd, true
}

func localLunarEclipseVisible(info LocalLunarEclipseInfo) bool {
	eventStart, eventEnd, ok := localLunarEclipseRange(info)
	if !ok {
		return false
	}
	return localLunarEclipseVisibleDuring(info, eventStart, eventEnd)
}

func localTotalLunarEclipseVisible(info LocalLunarEclipseInfo) bool {
	if !info.HasTotal || info.TotalStart.IsZero() || info.TotalEnd.IsZero() {
		return false
	}
	return localLunarEclipseVisibleDuring(info, info.TotalStart, info.TotalEnd)
}

func localLunarEclipseVisibleOnDate(info LocalLunarEclipseInfo, dayStart, dayEnd time.Time) bool {
	eventStart, eventEnd, ok := localLunarEclipseRange(info)
	if !ok {
		return false
	}
	segmentStart := maxLocalLunarTime(eventStart, dayStart)
	segmentEnd := minLocalLunarTime(eventEnd, dayEnd)
	if !segmentStart.Before(segmentEnd) {
		return false
	}
	return localLunarEclipseVisibleDuring(info, segmentStart, segmentEnd)
}

// localLunarEclipseVisibility 按站点自己的中天对食内高度分类。
func localLunarEclipseVisibility(info LocalLunarEclipseInfo) LocalLunarEclipseVisibility {
	start, end, ok := localLunarEclipseRange(info)
	if !ok {
		return LocalLunarEclipseInvisible
	}
	startVisible := localLunarEclipseAltitudeVisible(start, info)
	endVisible := localLunarEclipseAltitudeVisible(end, info)
	umbralVisible := localLunarEclipseUmbralVisible(info)
	switch {
	case startVisible && endVisible:
		if localLunarEclipseExtremumAltitude(info, start, end, false) > localLunarEclipseVisibilityThreshold(info.Height, info.Latitude) {
			return LocalLunarEclipseFull
		}
		return LocalLunarEclipseInterrupted
	case endVisible:
		if !umbralVisible {
			return LocalLunarEclipsePenumbraMoonrise
		}
		return LocalLunarEclipseMoonrise
	case startVisible:
		if !umbralVisible {
			return LocalLunarEclipsePenumbraMoonset
		}
		return LocalLunarEclipseMoonset
	case localLunarEclipseVisibleDuring(info, start, end):
		return LocalLunarEclipseRiseAndSet
	default:
		return LocalLunarEclipseInvisible
	}
}

// localLunarEclipseUmbralVisible 本影阶段是否有任何时刻月亮在地平线上；纯半影月食没有本影可看，
// 恒为 true，不把这类站点降级成"仅见半影"。极区掠射时 U1、食甚、U4 三点都可能在地平下而中途
// 短暂露出，所以这里取整个本影区间的高度极值，不能只比三个采样点。
func localLunarEclipseUmbralVisible(info LocalLunarEclipseInfo) bool {
	if !info.HasPartial {
		return true
	}
	start, end := info.PartialStart, info.PartialEnd
	if start.IsZero() || end.IsZero() || !start.Before(end) {
		return false
	}
	if localLunarEclipseAltitudeVisible(start, info) || localLunarEclipseAltitudeVisible(end, info) {
		return true
	}
	return localLunarEclipseExtremumAltitude(info, start, end, true) >
		localLunarEclipseVisibilityThreshold(info.Height, info.Latitude)
}

const lunarEclipseVisibilityScanSamples = 48

// localLunarEclipseExtremumAltitude 先粗扫定位、再在相邻两格内三分细化地取有限窗口内的月心高度极值。
func localLunarEclipseExtremumAltitude(info LocalLunarEclipseInfo, start, end time.Time, maximum bool) float64 {
	if !start.Before(end) {
		return lunarAltitude(start, info.Longitude, info.Latitude)
	}
	step := end.Sub(start) / lunarEclipseVisibilityScanSamples
	best := lunarAltitude(start, info.Longitude, info.Latitude)
	bestIndex := 0
	for index := 1; index <= lunarEclipseVisibilityScanSamples; index++ {
		value := lunarAltitude(start.Add(step*time.Duration(index)), info.Longitude, info.Latitude)
		if (value > best) == maximum {
			best, bestIndex = value, index
		}
	}
	lowIndex, highIndex := bestIndex-1, bestIndex+1
	if lowIndex < 0 {
		lowIndex = 0
	}
	if highIndex > lunarEclipseVisibilityScanSamples {
		highIndex = lunarEclipseVisibilityScanSamples
	}
	low := start.Add(step * time.Duration(lowIndex))
	high := start.Add(step * time.Duration(highIndex))
	for iteration := 0; iteration < 30; iteration++ {
		third := high.Sub(low) / 3
		first, second := low.Add(third), high.Add(-third)
		if (lunarAltitude(first, info.Longitude, info.Latitude) < lunarAltitude(second, info.Longitude, info.Latitude)) == maximum {
			low = first
		} else {
			high = second
		}
	}
	refined := lunarAltitude(low.Add(high.Sub(low)/2), info.Longitude, info.Latitude)
	if maximum {
		return math.Max(best, refined)
	}
	return math.Min(best, refined)
}

func localLunarEclipseVisibleDuring(info LocalLunarEclipseInfo, start, end time.Time) bool {
	if localLunarEclipseAltitudeVisible(start, info) || localLunarEclipseAltitudeVisible(end, info) {
		return true
	}
	return localLunarEclipseExtremumAltitude(info, start, end, true) >
		localLunarEclipseVisibilityThreshold(info.Height, info.Latitude)
}

func localLunarEclipseAltitudeVisible(date time.Time, info LocalLunarEclipseInfo) bool {
	return lunarAltitude(date, info.Longitude, info.Latitude) > localLunarEclipseVisibilityThreshold(info.Height, info.Latitude)
}

func localLunarEclipseVisibilityThreshold(height, latitude float64) float64 {
	if height <= 0 {
		return 0
	}
	return -basic.HeightDegreeByLat(height, latitude)
}

func maxLocalLunarTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minLocalLunarTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
