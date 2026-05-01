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

// LocalLunarEclipseInfo 站点月食信息, local lunar eclipse information.
//
// 所有时刻字段都保持用户输入的时区。
// 不存在的阶段使用零值 time.Time。
type LocalLunarEclipseInfo struct {
	// Type 月食类型, eclipse type.
	Type LunarEclipseType
	// HasSaros 存在沙罗序列信息, has Saros series metadata.
	HasSaros bool
	// Saros 是沙罗序列信息，包括系列号、系列内序号和总成员数。
	// Saros is Saros series metadata with the series number, member index, and total member count.
	Saros SarosInfo

	// Longitude 观测点经度，东正西负, observer longitude, east positive.
	Longitude float64
	// Latitude 观测点纬度，北正南负, observer latitude, north positive.
	Latitude float64
	// Height 观测点海拔高度，单位米, observer height in meters.
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
	switch {
	case hasLast && !hasNext:
		return last
	case !hasLast && hasNext:
		return next
	case !hasLast && !hasNext:
		return LocalLunarEclipseInfo{}
	}

	lastDistance := math.Abs(date.Sub(last.Maximum).Seconds())
	nextDistance := math.Abs(next.Maximum.Sub(date).Seconds())
	if lastDistance <= nextDistance {
		return last
	}
	return next
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
		candidateTT = basic.CalcMoonSHByJDE(candidateTT+float64(direction)*lunarEclipseSynodicMonthDays, 1)
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

	return LocalLunarEclipseInfo{
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

func localLunarEclipseVisibleDuring(info LocalLunarEclipseInfo, start, end time.Time) bool {
	if localLunarEclipseAltitudeVisible(start, info) || localLunarEclipseAltitudeVisible(end, info) {
		return true
	}

	for dayStart, _, _ := lunarEclipseLocalDayBounds(start); !dayStart.After(end); dayStart = nextLunarEclipseLocalDayStart(dayStart) {
		_, culminationSeed, _ := lunarEclipseLocalDayBounds(dayStart)
		culmination := lunarCulminationTime(culminationSeed, info.Longitude, info.Latitude)
		if culmination.Before(start) || culmination.After(end) {
			continue
		}
		if localLunarEclipseAltitudeVisible(culmination, info) {
			return true
		}
	}

	return false
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
