package eclipse

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

const (
	solarEclipseSynodicMonthDays = 29.530588853
	solarEclipseSearchLimit      = 36
	solarEclipseSearchEpsilonDay = 1e-8
	solarEclipseLatitudeLimitDeg = 2.0
)

type solarEclipseCalculator func(float64) basic.SolarEclipseResult

// SolarEclipseRadiusModel 日食月亮半径模型, lunar radius model for solar eclipses.
type SolarEclipseRadiusModel string

const (
	// SolarEclipseModelIAUSingleK IAU 单一 k 值模型, IAU single-k model.
	SolarEclipseModelIAUSingleK SolarEclipseRadiusModel = "iau_single_k"
	// SolarEclipseModelNASABulletinSplitK NASA bulletin 分裂 k 值模型, NASA bulletin split-k model.
	SolarEclipseModelNASABulletinSplitK SolarEclipseRadiusModel = "nasa_bulletin_split_k"
)

// SolarEclipseType 全局日食食型, global solar eclipse type.
type SolarEclipseType string

const (
	// SolarEclipseNone 无日食, no solar eclipse.
	SolarEclipseNone SolarEclipseType = "none"
	// SolarEclipsePartial 日偏食, partial solar eclipse.
	SolarEclipsePartial SolarEclipseType = "partial"
	// SolarEclipseAnnular 日环食, annular solar eclipse.
	SolarEclipseAnnular SolarEclipseType = "annular"
	// SolarEclipseTotal 日全食, total solar eclipse.
	SolarEclipseTotal SolarEclipseType = "total"
	// SolarEclipseHybrid 全环食, hybrid solar eclipse.
	SolarEclipseHybrid SolarEclipseType = "hybrid"
)

// SolarEclipseCentrality 中心线进入地球的方式, global eclipse centrality.
type SolarEclipseCentrality string

const (
	// SolarEclipseNonCentral 非中心食, non-central eclipse.
	SolarEclipseNonCentral SolarEclipseCentrality = "non_central"
	// SolarEclipseCentralOneLimit 单界中心食, central eclipse with one limit.
	SolarEclipseCentralOneLimit SolarEclipseCentrality = "central_one_limit"
	// SolarEclipseCentralTwoLimits 双界中心食, central eclipse with two limits.
	SolarEclipseCentralTwoLimits SolarEclipseCentrality = "central_two_limits"
)

// SolarEclipseInfo 全局日食信息, global solar eclipse information.
//
// 所有时刻字段都保持用户输入的时区。
// 不存在的阶段使用零值 time.Time。
type SolarEclipseInfo struct {
	// Model 日食月亮半径模型, eclipse lunar radius model.
	Model SolarEclipseRadiusModel
	// Type 全局食型, global eclipse type.
	Type SolarEclipseType
	// Centrality 中心性, eclipse centrality.
	Centrality SolarEclipseCentrality
	// HasSaros 存在沙罗序列信息, has Saros series metadata.
	HasSaros bool
	// Saros 是沙罗序列信息，包括系列号、系列内序号和总成员数。
	// Saros is Saros series metadata with the series number, member index, and total member count.
	Saros SarosInfo

	// GreatestEclipse 食甚时刻, greatest eclipse.
	GreatestEclipse time.Time
	// PartialBeginOnEarth 地球范围偏食始, partial eclipse begins on Earth.
	PartialBeginOnEarth time.Time
	// PartialEndOnEarth 地球范围偏食终, partial eclipse ends on Earth.
	PartialEndOnEarth time.Time
	// CentralBeginOnEarth 地球范围中心食始, central eclipse begins on Earth.
	CentralBeginOnEarth time.Time
	// CentralEndOnEarth 地球范围中心食终, central eclipse ends on Earth.
	CentralEndOnEarth time.Time

	// Magnitude 全局食分, global eclipse magnitude.
	Magnitude float64
	// Gamma 食甚时影轴到地心的距离, gamma at greatest eclipse.
	Gamma float64
	// PathWidthKM 食甚处中心食带宽度, central path width at greatest eclipse.
	PathWidthKM float64

	// GreatestLongitude 食甚点经度，东正西负, longitude of greatest eclipse, east positive.
	GreatestLongitude float64
	// GreatestLatitude 食甚点纬度，北正南负, latitude of greatest eclipse, north positive.
	GreatestLatitude float64

	// HasPartial 存在偏食阶段, has partial phase.
	HasPartial bool
	// HasCentral 存在中心食阶段, has central phase.
	HasCentral bool
	// HasAnnular 存在环食阶段, has annular phase.
	HasAnnular bool
	// HasTotal 存在全食阶段, has total phase.
	HasTotal bool
	// HasHybrid 为混合食, is hybrid eclipse.
	HasHybrid bool
}

// SolarEclipseOnDate 当地自然日全局日食查询 / local-date global solar eclipse query.
// Determine whether a global solar eclipse overlaps the local date, using NASA bulletin Split-K by default.
func SolarEclipseOnDate(date time.Time) (SolarEclipseInfo, bool) {
	return SolarEclipseOnDateNASABulletinSplitK(date)
}

// SolarEclipseOnDateNASABulletinSplitK 当地自然日全局日食查询（NASA bulletin Split-K） / local-date global solar eclipse query with NASA bulletin Split-K.
// Determine whether a global solar eclipse overlaps the local date with the NASA bulletin Split-K model.
func SolarEclipseOnDateNASABulletinSplitK(date time.Time) (SolarEclipseInfo, bool) {
	return solarEclipseOnDate(date, basic.SolarEclipseNASABulletinSplitK)
}

// SolarEclipseOnDateIAUSingleK 当地自然日全局日食查询（IAU Single-K） / local-date global solar eclipse query with IAU Single-K.
// Determine whether a global solar eclipse overlaps the local date with the IAU Single-K model.
func SolarEclipseOnDateIAUSingleK(date time.Time) (SolarEclipseInfo, bool) {
	return solarEclipseOnDate(date, basic.SolarEclipseIAUSingleK)
}

func solarEclipseOnDate(date time.Time, calculator solarEclipseCalculator) (SolarEclipseInfo, bool) {
	location := date.Location()
	dayStart, dayMid, dayEnd := solarEclipseLocalDayBounds(date)

	candidateTT := basic.CalcMoonSHByJDE(solarEclipseTimeToTTJDE(dayMid), 0)
	result := calculator(candidateTT)
	if result.Type == basic.SolarEclipseNone {
		return SolarEclipseInfo{}, false
	}

	info := solarEclipseInfoFromBasic(result, location)
	if !solarEclipseOverlapsDate(info, dayStart, dayEnd) {
		return SolarEclipseInfo{}, false
	}
	return info, true
}

// LastSolarEclipse 上次日食 / previous solar eclipse.
// Previous solar eclipse, using NASA bulletin Split-K by default.
func LastSolarEclipse(date time.Time) SolarEclipseInfo {
	return LastSolarEclipseNASABulletinSplitK(date)
}

// LastSolarEclipseNASABulletinSplitK 上次日食（NASA bulletin Split-K） / previous solar eclipse with NASA bulletin Split-K.
// Previous solar eclipse with the NASA bulletin Split-K model.
func LastSolarEclipseNASABulletinSplitK(date time.Time) SolarEclipseInfo {
	info, _ := searchSolarEclipse(date, -1, true, basic.SolarEclipseNASABulletinSplitK)
	return info
}

// LastSolarEclipseIAUSingleK 上次日食（IAU Single-K） / previous solar eclipse with IAU Single-K.
// Previous solar eclipse with the IAU Single-K model.
func LastSolarEclipseIAUSingleK(date time.Time) SolarEclipseInfo {
	info, _ := searchSolarEclipse(date, -1, true, basic.SolarEclipseIAUSingleK)
	return info
}

// NextSolarEclipse 下次日食 / next solar eclipse.
// Next solar eclipse, using NASA bulletin Split-K by default.
func NextSolarEclipse(date time.Time) SolarEclipseInfo {
	return NextSolarEclipseNASABulletinSplitK(date)
}

// NextSolarEclipseNASABulletinSplitK 下次日食（NASA bulletin Split-K） / next solar eclipse with NASA bulletin Split-K.
// Next solar eclipse with the NASA bulletin Split-K model.
func NextSolarEclipseNASABulletinSplitK(date time.Time) SolarEclipseInfo {
	info, _ := searchSolarEclipse(date, 1, false, basic.SolarEclipseNASABulletinSplitK)
	return info
}

// NextSolarEclipseIAUSingleK 下次日食（IAU Single-K） / next solar eclipse with IAU Single-K.
// Next solar eclipse with the IAU Single-K model.
func NextSolarEclipseIAUSingleK(date time.Time) SolarEclipseInfo {
	info, _ := searchSolarEclipse(date, 1, false, basic.SolarEclipseIAUSingleK)
	return info
}

// ClosestSolarEclipse 最近一次日食 / closest solar eclipse.
// Closest solar eclipse, using NASA bulletin Split-K by default.
func ClosestSolarEclipse(date time.Time) SolarEclipseInfo {
	return ClosestSolarEclipseNASABulletinSplitK(date)
}

// ClosestSolarEclipseNASABulletinSplitK 最近一次日食（NASA bulletin Split-K） / closest solar eclipse with NASA bulletin Split-K.
// Closest solar eclipse with the NASA bulletin Split-K model.
func ClosestSolarEclipseNASABulletinSplitK(date time.Time) SolarEclipseInfo {
	last, hasLast := searchSolarEclipse(date, -1, true, basic.SolarEclipseNASABulletinSplitK)
	next, hasNext := searchSolarEclipse(date, 1, false, basic.SolarEclipseNASABulletinSplitK)
	return closestSolarEclipse(date, last, hasLast, next, hasNext)
}

// ClosestSolarEclipseIAUSingleK 最近一次日食（IAU Single-K） / closest solar eclipse with IAU Single-K.
// Closest solar eclipse with the IAU Single-K model.
func ClosestSolarEclipseIAUSingleK(date time.Time) SolarEclipseInfo {
	last, hasLast := searchSolarEclipse(date, -1, true, basic.SolarEclipseIAUSingleK)
	next, hasNext := searchSolarEclipse(date, 1, false, basic.SolarEclipseIAUSingleK)
	return closestSolarEclipse(date, last, hasLast, next, hasNext)
}

func closestSolarEclipse(
	date time.Time,
	last SolarEclipseInfo,
	hasLast bool,
	next SolarEclipseInfo,
	hasNext bool,
) SolarEclipseInfo {
	switch {
	case hasLast && !hasNext:
		return last
	case !hasLast && hasNext:
		return next
	case !hasLast && !hasNext:
		return SolarEclipseInfo{}
	}

	lastDistance := math.Abs(date.Sub(last.GreatestEclipse).Seconds())
	nextDistance := math.Abs(next.GreatestEclipse.Sub(date).Seconds())
	if lastDistance <= nextDistance {
		return last
	}
	return next
}

func searchSolarEclipse(
	date time.Time,
	direction int,
	includeCurrent bool,
	calculator solarEclipseCalculator,
) (SolarEclipseInfo, bool) {
	targetTT := solarEclipseTimeToTTJDE(date)
	candidateTT := basic.CalcMoonSHByJDE(targetTT, 0)

	for i := 0; i < solarEclipseSearchLimit; i++ {
		if isPotentialSolarEclipse(candidateTT) {
			result := calculator(candidateTT)
			if result.Type != basic.SolarEclipseNone && solarEclipseMatchesDirection(result.GreatestEclipse, targetTT, direction, includeCurrent) {
				return solarEclipseInfoFromBasic(result, date.Location()), true
			}
		}
		candidateTT = nextEclipseSearchCandidateTT(candidateTT, 0, direction, solarEclipseSynodicMonthDays)
	}

	return SolarEclipseInfo{}, false
}

func isPotentialSolarEclipse(newMoonTT float64) bool {
	return math.Abs(basic.HMoonTrueBo(newMoonTT)) <= solarEclipseLatitudeLimitDeg
}

func solarEclipseMatchesDirection(greatestTT, targetTT float64, direction int, includeCurrent bool) bool {
	delta := greatestTT - targetTT
	if math.Abs(delta) <= solarEclipseSearchEpsilonDay {
		return direction < 0 && includeCurrent
	}
	if direction > 0 {
		return delta > 0
	}
	return delta < 0
}

func solarEclipseLocalDayBounds(date time.Time) (time.Time, time.Time, time.Time) {
	location := date.Location()
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, location)
	dayMid := time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, location)
	dayEnd := time.Date(date.Year(), date.Month(), date.Day()+1, 0, 0, 0, 0, location)
	return dayStart, dayMid, dayEnd
}

func nextSolarEclipseLocalDayStart(dayStart time.Time) time.Time {
	location := dayStart.Location()
	return time.Date(dayStart.Year(), dayStart.Month(), dayStart.Day()+1, 0, 0, 0, 0, location)
}

func solarEclipseInfoFromBasic(result basic.SolarEclipseResult, location *time.Location) SolarEclipseInfo {
	saros, hasSaros := solarSarosInfo(result.GreatestEclipse)
	return SolarEclipseInfo{
		Model:               mapBasicSolarEclipseModel(result.Model),
		Type:                mapBasicSolarEclipseType(result.Type),
		Centrality:          mapBasicSolarEclipseCentrality(result.Centrality),
		HasSaros:            hasSaros,
		Saros:               saros,
		GreatestEclipse:     solarEclipseTTJDEToTime(result.GreatestEclipse, location),
		PartialBeginOnEarth: solarEclipseTTJDEToTime(result.PartialBeginOnEarth, location),
		PartialEndOnEarth:   solarEclipseTTJDEToTime(result.PartialEndOnEarth, location),
		CentralBeginOnEarth: solarEclipseTTJDEToTime(result.CentralBeginOnEarth, location),
		CentralEndOnEarth:   solarEclipseTTJDEToTime(result.CentralEndOnEarth, location),
		Magnitude:           result.Magnitude,
		Gamma:               result.Gamma,
		PathWidthKM:         result.PathWidthKM,
		GreatestLongitude:   result.GreatestLongitude,
		GreatestLatitude:    result.GreatestLatitude,
		HasPartial:          result.HasPartial,
		HasCentral:          result.HasCentral,
		HasAnnular:          result.HasAnnular,
		HasTotal:            result.HasTotal,
		HasHybrid:           result.HasHybrid,
	}
}

func mapBasicSolarEclipseModel(model basic.SolarEclipseRadiusModel) SolarEclipseRadiusModel {
	switch model {
	case basic.SolarEclipseModelIAUSingleK:
		return SolarEclipseModelIAUSingleK
	default:
		return SolarEclipseModelNASABulletinSplitK
	}
}

func mapBasicSolarEclipseType(eclipseType basic.SolarEclipseType) SolarEclipseType {
	switch eclipseType {
	case basic.SolarEclipsePartial:
		return SolarEclipsePartial
	case basic.SolarEclipseAnnular:
		return SolarEclipseAnnular
	case basic.SolarEclipseTotal:
		return SolarEclipseTotal
	case basic.SolarEclipseHybrid:
		return SolarEclipseHybrid
	default:
		return SolarEclipseNone
	}
}

func mapBasicSolarEclipseCentrality(centrality basic.SolarEclipseCentrality) SolarEclipseCentrality {
	switch centrality {
	case basic.SolarEclipseCentralOneLimit:
		return SolarEclipseCentralOneLimit
	case basic.SolarEclipseCentralTwoLimits:
		return SolarEclipseCentralTwoLimits
	default:
		return SolarEclipseNonCentral
	}
}

func solarEclipseOverlapsDate(info SolarEclipseInfo, dayStart, dayEnd time.Time) bool {
	eventStart, eventEnd, ok := solarEclipseRange(info)
	if !ok {
		return false
	}
	return !eventEnd.Before(dayStart) && eventStart.Before(dayEnd)
}

func solarEclipseRange(info SolarEclipseInfo) (time.Time, time.Time, bool) {
	if !info.HasPartial {
		return time.Time{}, time.Time{}, false
	}
	return info.PartialBeginOnEarth, info.PartialEndOnEarth, true
}

func solarEclipseTTJDEToTime(ttJDE float64, location *time.Location) time.Time {
	if ttJDE == 0 {
		return time.Time{}
	}
	utcJDE := basic.TD2UT(ttJDE, false)
	return basic.JDE2DateByZone(utcJDE, location, false)
}

func solarEclipseTimeToTTJDE(date time.Time) float64 {
	utcJDE := basic.Date2JDE(date.UTC())
	return basic.TD2UT(utcJDE, true)
}
