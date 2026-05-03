package eclipse

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

const (
	localSolarEclipseSynodicMonthDays = 29.530588853
	localSolarEclipseSearchLimit      = 6000
	localSolarEclipseSearchEpsilonDay = 1e-8
	localSolarEclipseLatitudeLimitDeg = 2.0
)

type localSolarEclipseCalculator struct {
	global func(float64) basic.SolarEclipseResult
	local  func(float64, float64, float64, float64) basic.LocalSolarEclipseResult
}

type localSolarEclipseQueryMode int

const (
	localSolarEclipseQueryVisible localSolarEclipseQueryMode = iota
	localSolarEclipseQueryGeometric
)

// LocalSolarEclipseContactPoint 表示站心日食在日面上的接触点方位。
// LocalSolarEclipseContactPoint describes a local solar eclipse contact point on the Sun limb.
type LocalSolarEclipseContactPoint struct {
	// Label 是接触标签，如 C1/C2/C3/C4。
	// Label is the contact label, such as C1/C2/C3/C4.
	Label string
	// Time 是该接触时刻，保持用户输入时区。
	// Time is the contact time, preserving the input timezone.
	Time time.Time
	// ContactPositionAngle 是日面接触点位置角，从天球北点起向东量，单位度。
	// ContactPositionAngle is the Sun-limb contact position angle from celestial north toward east, in degrees.
	ContactPositionAngle float64
	// ContactClockwiseAngle 是图面上从北点顺时针量到接触点的角度，单位度。
	// ContactClockwiseAngle is the chart clockwise angle from north to the contact point, in degrees.
	ContactClockwiseAngle float64
	// MoonCenterPositionAngle 是月心相对日心的位置角，从北点起向东量，单位度。
	// MoonCenterPositionAngle is the Moon-center position angle from the Sun center, in degrees.
	MoonCenterPositionAngle float64
}

// LocalSolarEclipseInfo 站心日食信息, local solar eclipse information.
//
// 所有时刻字段都保持用户输入的时区。
// 不存在的阶段使用零值 time.Time。
type LocalSolarEclipseInfo struct {
	// Model 日食月亮半径模型, eclipse lunar radius model.
	Model SolarEclipseRadiusModel
	// Type 站心食型, local eclipse type.
	Type SolarEclipseType
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

	// GreatestEclipse 食甚时刻, greatest eclipse.
	GreatestEclipse time.Time
	// PartialStart 偏食始 / 初亏, partial eclipse begins.
	PartialStart time.Time
	// PartialEnd 偏食终 / 复圆, partial eclipse ends.
	PartialEnd time.Time
	// CentralStart 中心食始；对全食为食既，对环食为环食始, central eclipse begins.
	CentralStart time.Time
	// CentralEnd 中心食终；对全食为生光，对环食为环食终, central eclipse ends.
	CentralEnd time.Time

	// Magnitude 站心食分, local eclipse magnitude.
	Magnitude float64
	// Obscuration 食甚时太阳视圆面遮蔽率, obscuration at greatest eclipse.
	Obscuration float64
	// Separation 食甚时日月中心角距，单位度, center separation at greatest eclipse in degrees.
	Separation float64
	// SunAltitude 食甚时太阳高度角，单位度, Sun altitude at greatest eclipse in degrees.
	SunAltitude float64
	// SunAzimuth 食甚时太阳方位角，单位度, Sun azimuth at greatest eclipse in degrees.
	SunAzimuth float64

	// VisibleAtGreatest 食甚时太阳中心在地平线上方, Sun center above horizon at greatest eclipse.
	VisibleAtGreatest bool

	// ContactPoints 是各接触时刻在日面上的接触点方位。
	// ContactPoints are Sun-limb contact position angles at eclipse contacts.
	ContactPoints []LocalSolarEclipseContactPoint

	// HasPartial 存在偏食阶段, has partial phase.
	HasPartial bool
	// HasCentral 存在中心食阶段, has central phase.
	HasCentral bool
	// HasAnnular 存在环食阶段, has annular phase.
	HasAnnular bool
	// HasTotal 存在全食阶段, has total phase.
	HasTotal bool
}

// LocalSolarEclipseOnDate 当地站心日食查询 / local topocentric solar eclipse query.
// Determine whether a visible local solar eclipse overlaps the local date, using NASA bulletin Split-K by default.
func LocalSolarEclipseOnDate(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return LocalSolarEclipseOnDateNASABulletinSplitK(date, lon, lat, height)
}

// LocalSolarEclipseOnDateNASABulletinSplitK 当地站心日食查询（NASA bulletin Split-K） / local topocentric solar eclipse query with NASA bulletin Split-K.
// Determine whether a visible local solar eclipse overlaps the local date with the NASA bulletin Split-K model.
func LocalSolarEclipseOnDateNASABulletinSplitK(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return localSolarEclipseOnDate(date, lon, lat, height, localSolarEclipseNASABulletinSplitK, localSolarEclipseQueryVisible)
}

// LocalSolarEclipseOnDateIAUSingleK 当地站心日食查询（IAU Single-K） / local topocentric solar eclipse query with IAU Single-K.
// Determine whether a visible local solar eclipse overlaps the local date with the IAU Single-K model.
func LocalSolarEclipseOnDateIAUSingleK(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return localSolarEclipseOnDate(date, lon, lat, height, localSolarEclipseIAUSingleK, localSolarEclipseQueryVisible)
}

// GeometricLocalSolarEclipseOnDate 当地站心几何日食查询 / local geometric solar eclipse query.
// Determine whether a geometric local solar eclipse overlaps the local date, using NASA bulletin Split-K by default.
func GeometricLocalSolarEclipseOnDate(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return GeometricLocalSolarEclipseOnDateNASABulletinSplitK(date, lon, lat, height)
}

// GeometricLocalSolarEclipseOnDateNASABulletinSplitK 当地站心几何日食查询（NASA bulletin Split-K） / local geometric solar eclipse query with NASA bulletin Split-K.
// Determine whether a geometric local solar eclipse overlaps the local date with the NASA bulletin Split-K model.
func GeometricLocalSolarEclipseOnDateNASABulletinSplitK(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return localSolarEclipseOnDate(date, lon, lat, height, localSolarEclipseNASABulletinSplitK, localSolarEclipseQueryGeometric)
}

// GeometricLocalSolarEclipseOnDateIAUSingleK 当地站心几何日食查询（IAU Single-K） / local geometric solar eclipse query with IAU Single-K.
// Determine whether a geometric local solar eclipse overlaps the local date with the IAU Single-K model.
func GeometricLocalSolarEclipseOnDateIAUSingleK(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return localSolarEclipseOnDate(date, lon, lat, height, localSolarEclipseIAUSingleK, localSolarEclipseQueryGeometric)
}

func localSolarEclipseOnDate(
	date time.Time,
	lon, lat, height float64,
	calculator localSolarEclipseCalculator,
	mode localSolarEclipseQueryMode,
) (LocalSolarEclipseInfo, bool) {
	location := date.Location()
	dayStart, dayMid, dayEnd := solarEclipseLocalDayBounds(date)

	candidateTT := basic.CalcMoonSHByJDE(solarEclipseTimeToTTJDE(dayMid), 0)
	if !isPotentialLocalSolarEclipse(candidateTT) {
		return LocalSolarEclipseInfo{}, false
	}

	globalResult := calculator.global(candidateTT)
	if globalResult.Type == basic.SolarEclipseNone {
		return LocalSolarEclipseInfo{}, false
	}

	result := calculator.local(globalResult.GreatestEclipse, lon, lat, height)
	if result.Type == basic.SolarEclipseNone {
		return LocalSolarEclipseInfo{}, false
	}

	info := localSolarEclipseInfoFromBasic(result, lon, lat, height, location)
	if !localSolarEclipseOverlapsDate(info, dayStart, dayEnd) {
		return LocalSolarEclipseInfo{}, false
	}
	if mode == localSolarEclipseQueryVisible && !localSolarEclipseVisibleOnDate(info, dayStart, dayEnd) {
		return LocalSolarEclipseInfo{}, false
	}
	return info, true
}

// LastLocalSolarEclipse 上次站心日食 / previous local solar eclipse.
// Previous visible local solar eclipse, using NASA bulletin Split-K by default.
func LastLocalSolarEclipse(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return LastLocalSolarEclipseNASABulletinSplitK(date, lon, lat, height)
}

// LastLocalSolarEclipseNASABulletinSplitK 上次站心日食（NASA bulletin Split-K） / previous local solar eclipse with NASA bulletin Split-K.
// Previous visible local solar eclipse with the NASA bulletin Split-K model.
func LastLocalSolarEclipseNASABulletinSplitK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	info, _ := searchLocalSolarEclipse(date, lon, lat, height, -1, true, localSolarEclipseNASABulletinSplitK, localSolarEclipseQueryVisible)
	return info
}

// LastLocalSolarEclipseIAUSingleK 上次站心日食（IAU Single-K） / previous local solar eclipse with IAU Single-K.
// Previous visible local solar eclipse with the IAU Single-K model.
func LastLocalSolarEclipseIAUSingleK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	info, _ := searchLocalSolarEclipse(date, lon, lat, height, -1, true, localSolarEclipseIAUSingleK, localSolarEclipseQueryVisible)
	return info
}

// LastGeometricLocalSolarEclipse 上次站心几何日食 / previous geometric local solar eclipse.
// Previous geometric local solar eclipse, using NASA bulletin Split-K by default.
func LastGeometricLocalSolarEclipse(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return LastGeometricLocalSolarEclipseNASABulletinSplitK(date, lon, lat, height)
}

// LastGeometricLocalSolarEclipseNASABulletinSplitK 上次站心几何日食（NASA bulletin Split-K） / previous geometric local solar eclipse with NASA bulletin Split-K.
// Previous geometric local solar eclipse with the NASA bulletin Split-K model.
func LastGeometricLocalSolarEclipseNASABulletinSplitK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	info, _ := searchLocalSolarEclipse(date, lon, lat, height, -1, true, localSolarEclipseNASABulletinSplitK, localSolarEclipseQueryGeometric)
	return info
}

// LastGeometricLocalSolarEclipseIAUSingleK 上次站心几何日食（IAU Single-K） / previous geometric local solar eclipse with IAU Single-K.
// Previous geometric local solar eclipse with the IAU Single-K model.
func LastGeometricLocalSolarEclipseIAUSingleK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	info, _ := searchLocalSolarEclipse(date, lon, lat, height, -1, true, localSolarEclipseIAUSingleK, localSolarEclipseQueryGeometric)
	return info
}

// LastLocalTotalSolarEclipse 上次站心日全食 / previous local total solar eclipse.
// Previous visible local total solar eclipse, using NASA bulletin Split-K by default.
func LastLocalTotalSolarEclipse(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return searchLocalTotalSolarEclipse(date, lon, lat, height, -1, true, localSolarEclipseNASABulletinSplitK, localSolarEclipseQueryVisible)
}

// LastLocalAnnularSolarEclipse 上次站心日环食 / previous local annular solar eclipse.
// Previous visible local annular solar eclipse, using NASA bulletin Split-K by default.
func LastLocalAnnularSolarEclipse(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return searchLocalAnnularSolarEclipse(date, lon, lat, height, -1, true, localSolarEclipseNASABulletinSplitK, localSolarEclipseQueryVisible)
}

// NextLocalSolarEclipse 下次站心日食 / next local solar eclipse.
// Next visible local solar eclipse, using NASA bulletin Split-K by default.
func NextLocalSolarEclipse(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return NextLocalSolarEclipseNASABulletinSplitK(date, lon, lat, height)
}

// NextLocalSolarEclipseNASABulletinSplitK 下次站心日食（NASA bulletin Split-K） / next local solar eclipse with NASA bulletin Split-K.
// Next visible local solar eclipse with the NASA bulletin Split-K model.
func NextLocalSolarEclipseNASABulletinSplitK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	info, _ := searchLocalSolarEclipse(date, lon, lat, height, 1, false, localSolarEclipseNASABulletinSplitK, localSolarEclipseQueryVisible)
	return info
}

// NextLocalSolarEclipseIAUSingleK 下次站心日食（IAU Single-K） / next local solar eclipse with IAU Single-K.
// Next visible local solar eclipse with the IAU Single-K model.
func NextLocalSolarEclipseIAUSingleK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	info, _ := searchLocalSolarEclipse(date, lon, lat, height, 1, false, localSolarEclipseIAUSingleK, localSolarEclipseQueryVisible)
	return info
}

// NextGeometricLocalSolarEclipse 下次站心几何日食 / next geometric local solar eclipse.
// Next geometric local solar eclipse, using NASA bulletin Split-K by default.
func NextGeometricLocalSolarEclipse(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return NextGeometricLocalSolarEclipseNASABulletinSplitK(date, lon, lat, height)
}

// NextGeometricLocalSolarEclipseNASABulletinSplitK 下次站心几何日食（NASA bulletin Split-K） / next geometric local solar eclipse with NASA bulletin Split-K.
// Next geometric local solar eclipse with the NASA bulletin Split-K model.
func NextGeometricLocalSolarEclipseNASABulletinSplitK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	info, _ := searchLocalSolarEclipse(date, lon, lat, height, 1, false, localSolarEclipseNASABulletinSplitK, localSolarEclipseQueryGeometric)
	return info
}

// NextGeometricLocalSolarEclipseIAUSingleK 下次站心几何日食（IAU Single-K） / next geometric local solar eclipse with IAU Single-K.
// Next geometric local solar eclipse with the IAU Single-K model.
func NextGeometricLocalSolarEclipseIAUSingleK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	info, _ := searchLocalSolarEclipse(date, lon, lat, height, 1, false, localSolarEclipseIAUSingleK, localSolarEclipseQueryGeometric)
	return info
}

// NextLocalTotalSolarEclipse 下次站心日全食 / next local total solar eclipse.
// Next visible local total solar eclipse, using NASA bulletin Split-K by default.
func NextLocalTotalSolarEclipse(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return searchLocalTotalSolarEclipse(date, lon, lat, height, 1, false, localSolarEclipseNASABulletinSplitK, localSolarEclipseQueryVisible)
}

// NextLocalAnnularSolarEclipse 下次站心日环食 / next local annular solar eclipse.
// Next visible local annular solar eclipse, using NASA bulletin Split-K by default.
func NextLocalAnnularSolarEclipse(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	return searchLocalAnnularSolarEclipse(date, lon, lat, height, 1, false, localSolarEclipseNASABulletinSplitK, localSolarEclipseQueryVisible)
}

// ClosestLocalSolarEclipse 最近一次站心日食 / closest local solar eclipse.
// Closest visible local solar eclipse, using NASA bulletin Split-K by default.
func ClosestLocalSolarEclipse(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return ClosestLocalSolarEclipseNASABulletinSplitK(date, lon, lat, height)
}

// ClosestLocalSolarEclipseNASABulletinSplitK 最近一次站心日食（NASA bulletin Split-K） / closest local solar eclipse with NASA bulletin Split-K.
// Closest visible local solar eclipse with the NASA bulletin Split-K model.
func ClosestLocalSolarEclipseNASABulletinSplitK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	last, hasLast := searchLocalSolarEclipse(date, lon, lat, height, -1, true, localSolarEclipseNASABulletinSplitK, localSolarEclipseQueryVisible)
	next, hasNext := searchLocalSolarEclipse(date, lon, lat, height, 1, false, localSolarEclipseNASABulletinSplitK, localSolarEclipseQueryVisible)
	return closestLocalSolarEclipse(date, last, hasLast, next, hasNext)
}

// ClosestLocalSolarEclipseIAUSingleK 最近一次站心日食（IAU Single-K） / closest local solar eclipse with IAU Single-K.
// Closest visible local solar eclipse with the IAU Single-K model.
func ClosestLocalSolarEclipseIAUSingleK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	last, hasLast := searchLocalSolarEclipse(date, lon, lat, height, -1, true, localSolarEclipseIAUSingleK, localSolarEclipseQueryVisible)
	next, hasNext := searchLocalSolarEclipse(date, lon, lat, height, 1, false, localSolarEclipseIAUSingleK, localSolarEclipseQueryVisible)
	return closestLocalSolarEclipse(date, last, hasLast, next, hasNext)
}

// ClosestGeometricLocalSolarEclipse 最近一次站心几何日食 / closest geometric local solar eclipse.
// Closest geometric local solar eclipse, using NASA bulletin Split-K by default.
func ClosestGeometricLocalSolarEclipse(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	return ClosestGeometricLocalSolarEclipseNASABulletinSplitK(date, lon, lat, height)
}

// ClosestGeometricLocalSolarEclipseNASABulletinSplitK 最近一次站心几何日食（NASA bulletin Split-K） / closest geometric local solar eclipse with NASA bulletin Split-K.
// Closest geometric local solar eclipse with the NASA bulletin Split-K model.
func ClosestGeometricLocalSolarEclipseNASABulletinSplitK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	last, hasLast := searchLocalSolarEclipse(date, lon, lat, height, -1, true, localSolarEclipseNASABulletinSplitK, localSolarEclipseQueryGeometric)
	next, hasNext := searchLocalSolarEclipse(date, lon, lat, height, 1, false, localSolarEclipseNASABulletinSplitK, localSolarEclipseQueryGeometric)
	return closestLocalSolarEclipse(date, last, hasLast, next, hasNext)
}

// ClosestGeometricLocalSolarEclipseIAUSingleK 最近一次站心几何日食（IAU Single-K） / closest geometric local solar eclipse with IAU Single-K.
// Closest geometric local solar eclipse with the IAU Single-K model.
func ClosestGeometricLocalSolarEclipseIAUSingleK(date time.Time, lon, lat, height float64) LocalSolarEclipseInfo {
	last, hasLast := searchLocalSolarEclipse(date, lon, lat, height, -1, true, localSolarEclipseIAUSingleK, localSolarEclipseQueryGeometric)
	next, hasNext := searchLocalSolarEclipse(date, lon, lat, height, 1, false, localSolarEclipseIAUSingleK, localSolarEclipseQueryGeometric)
	return closestLocalSolarEclipse(date, last, hasLast, next, hasNext)
}

// ClosestLocalTotalSolarEclipse 最近一次站心日全食 / closest local total solar eclipse.
// Closest visible local total solar eclipse, using NASA bulletin Split-K by default.
func ClosestLocalTotalSolarEclipse(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	last, hasLast := searchLocalTotalSolarEclipse(date, lon, lat, height, -1, true, localSolarEclipseNASABulletinSplitK, localSolarEclipseQueryVisible)
	next, hasNext := searchLocalTotalSolarEclipse(date, lon, lat, height, 1, false, localSolarEclipseNASABulletinSplitK, localSolarEclipseQueryVisible)
	return closestLocalSolarEclipseResult(date, last, hasLast, next, hasNext)
}

// ClosestLocalAnnularSolarEclipse 最近一次站心日环食 / closest local annular solar eclipse.
// Closest visible local annular solar eclipse, using NASA bulletin Split-K by default.
func ClosestLocalAnnularSolarEclipse(date time.Time, lon, lat, height float64) (LocalSolarEclipseInfo, bool) {
	last, hasLast := searchLocalAnnularSolarEclipse(date, lon, lat, height, -1, true, localSolarEclipseNASABulletinSplitK, localSolarEclipseQueryVisible)
	next, hasNext := searchLocalAnnularSolarEclipse(date, lon, lat, height, 1, false, localSolarEclipseNASABulletinSplitK, localSolarEclipseQueryVisible)
	return closestLocalSolarEclipseResult(date, last, hasLast, next, hasNext)
}

func closestLocalSolarEclipse(
	date time.Time,
	last LocalSolarEclipseInfo,
	hasLast bool,
	next LocalSolarEclipseInfo,
	hasNext bool,
) LocalSolarEclipseInfo {
	info, _ := closestLocalSolarEclipseResult(date, last, hasLast, next, hasNext)
	return info
}

func closestLocalSolarEclipseResult(
	date time.Time,
	last LocalSolarEclipseInfo,
	hasLast bool,
	next LocalSolarEclipseInfo,
	hasNext bool,
) (LocalSolarEclipseInfo, bool) {
	switch {
	case hasLast && !hasNext:
		return last, true
	case !hasLast && hasNext:
		return next, true
	case !hasLast && !hasNext:
		return LocalSolarEclipseInfo{}, false
	}

	lastDistance := math.Abs(date.Sub(last.GreatestEclipse).Seconds())
	nextDistance := math.Abs(next.GreatestEclipse.Sub(date).Seconds())
	if lastDistance <= nextDistance {
		return last, true
	}
	return next, true
}

func searchLocalSolarEclipse(
	date time.Time,
	lon, lat, height float64,
	direction int,
	includeCurrent bool,
	calculator localSolarEclipseCalculator,
	mode localSolarEclipseQueryMode,
) (LocalSolarEclipseInfo, bool) {
	targetTT := solarEclipseTimeToTTJDE(date)
	candidateTT := basic.CalcMoonSHByJDE(targetTT, 0)

	for i := 0; i < localSolarEclipseSearchLimit; i++ {
		if isPotentialLocalSolarEclipse(candidateTT) {
			globalResult := calculator.global(candidateTT)
			if globalResult.Type != basic.SolarEclipseNone {
				result := calculator.local(globalResult.GreatestEclipse, lon, lat, height)
				if result.Type != basic.SolarEclipseNone {
					info := localSolarEclipseInfoFromBasic(result, lon, lat, height, date.Location())
					if (mode != localSolarEclipseQueryVisible || localSolarEclipseVisible(info)) &&
						localSolarEclipseMatchesDirection(result.GreatestEclipse, targetTT, direction, includeCurrent) {
						return info, true
					}
				}
			}
		}
		candidateTT = nextEclipseSearchCandidateTT(candidateTT, 0, direction, localSolarEclipseSynodicMonthDays)
	}

	return LocalSolarEclipseInfo{}, false
}

func searchLocalTotalSolarEclipse(
	date time.Time,
	lon, lat, height float64,
	direction int,
	includeCurrent bool,
	calculator localSolarEclipseCalculator,
	mode localSolarEclipseQueryMode,
) (LocalSolarEclipseInfo, bool) {
	targetTT := solarEclipseTimeToTTJDE(date)
	candidateTT := basic.CalcMoonSHByJDE(targetTT, 0)

	for i := 0; i < localSolarEclipseSearchLimit; i++ {
		if isPotentialLocalSolarEclipse(candidateTT) {
			globalResult := calculator.global(candidateTT)
			if globalResult.HasTotal || globalResult.HasHybrid {
				result := calculator.local(globalResult.GreatestEclipse, lon, lat, height)
				if result.HasTotal {
					info := localSolarEclipseInfoFromBasic(result, lon, lat, height, date.Location())
					if (mode != localSolarEclipseQueryVisible || localCentralSolarEclipseVisible(info)) &&
						localSolarEclipseMatchesDirection(result.GreatestEclipse, targetTT, direction, includeCurrent) {
						return info, true
					}
				}
			}
		}
		candidateTT = nextEclipseSearchCandidateTT(candidateTT, 0, direction, localSolarEclipseSynodicMonthDays)
	}

	return LocalSolarEclipseInfo{}, false
}

func searchLocalAnnularSolarEclipse(
	date time.Time,
	lon, lat, height float64,
	direction int,
	includeCurrent bool,
	calculator localSolarEclipseCalculator,
	mode localSolarEclipseQueryMode,
) (LocalSolarEclipseInfo, bool) {
	targetTT := solarEclipseTimeToTTJDE(date)
	candidateTT := basic.CalcMoonSHByJDE(targetTT, 0)

	for i := 0; i < localSolarEclipseSearchLimit; i++ {
		if isPotentialLocalSolarEclipse(candidateTT) {
			globalResult := calculator.global(candidateTT)
			if globalResult.HasAnnular || globalResult.HasHybrid {
				result := calculator.local(globalResult.GreatestEclipse, lon, lat, height)
				if result.HasAnnular && !result.HasTotal {
					info := localSolarEclipseInfoFromBasic(result, lon, lat, height, date.Location())
					if (mode != localSolarEclipseQueryVisible || localCentralSolarEclipseVisible(info)) &&
						localSolarEclipseMatchesDirection(result.GreatestEclipse, targetTT, direction, includeCurrent) {
						return info, true
					}
				}
			}
		}
		candidateTT = nextEclipseSearchCandidateTT(candidateTT, 0, direction, localSolarEclipseSynodicMonthDays)
	}

	return LocalSolarEclipseInfo{}, false
}

func isPotentialLocalSolarEclipse(newMoonTT float64) bool {
	return math.Abs(basic.HMoonTrueBo(newMoonTT)) <= localSolarEclipseLatitudeLimitDeg
}

func localSolarEclipseMatchesDirection(greatestTT, targetTT float64, direction int, includeCurrent bool) bool {
	delta := greatestTT - targetTT
	if math.Abs(delta) <= localSolarEclipseSearchEpsilonDay {
		return direction < 0 && includeCurrent
	}
	if direction > 0 {
		return delta > 0
	}
	return delta < 0
}

func localSolarEclipseInfoFromBasic(
	result basic.LocalSolarEclipseResult,
	lon, lat, height float64,
	location *time.Location,
) LocalSolarEclipseInfo {
	info := localSolarEclipseInfoFieldsFromBasic(result, lon, lat, height, location)
	info.ContactPoints = localSolarEclipseContactPointsFromBasic(result, lon, lat, height, location)
	return info
}

func localSolarEclipseInfoFromDiagram(
	diagram basic.LocalSolarEclipseDiagramResult,
	lon, lat, height float64,
	location *time.Location,
) LocalSolarEclipseInfo {
	info := localSolarEclipseInfoFieldsFromBasic(diagram.Eclipse, lon, lat, height, location)
	info.ContactPoints = localSolarEclipseContactPointsFromFrames(diagram.Frames, location)
	return info
}

func localSolarEclipseInfoFieldsFromBasic(
	result basic.LocalSolarEclipseResult,
	lon, lat, height float64,
	location *time.Location,
) LocalSolarEclipseInfo {
	visibleThreshold := localSolarEclipseVisibilityThreshold(height, lat)
	saros, hasSaros := solarSarosInfo(result.GreatestEclipse)
	return LocalSolarEclipseInfo{
		Model:             mapBasicSolarEclipseModel(result.Model),
		Type:              mapBasicSolarEclipseType(result.Type),
		HasSaros:          hasSaros,
		Saros:             saros,
		Longitude:         lon,
		Latitude:          lat,
		Height:            height,
		GreatestEclipse:   solarEclipseTTJDEToTime(result.GreatestEclipse, location),
		PartialStart:      solarEclipseTTJDEToTime(result.PartialStart, location),
		PartialEnd:        solarEclipseTTJDEToTime(result.PartialEnd, location),
		CentralStart:      solarEclipseTTJDEToTime(result.CentralStart, location),
		CentralEnd:        solarEclipseTTJDEToTime(result.CentralEnd, location),
		Magnitude:         result.Magnitude,
		Obscuration:       result.Obscuration,
		Separation:        result.Separation,
		SunAltitude:       result.SunAltitude,
		SunAzimuth:        result.SunAzimuth,
		VisibleAtGreatest: result.SunAltitude > visibleThreshold,
		HasPartial:        result.HasPartial,
		HasCentral:        result.HasCentral,
		HasAnnular:        result.HasAnnular,
		HasTotal:          result.HasTotal,
	}
}

func localSolarEclipseContactPointsFromBasic(
	result basic.LocalSolarEclipseResult,
	lon, lat, height float64,
	location *time.Location,
) []LocalSolarEclipseContactPoint {
	if !result.HasPartial {
		return nil
	}
	options := basic.LocalSolarEclipseDiagramOptions{StepDays: 1}
	var diagram basic.LocalSolarEclipseDiagramResult
	if result.Model == basic.SolarEclipseModelIAUSingleK {
		diagram = basic.LocalSolarEclipseDiagramIAUSingleK(result.GreatestEclipse, lon, lat, height, options)
	} else {
		diagram = basic.LocalSolarEclipseDiagramNASABulletinSplitK(result.GreatestEclipse, lon, lat, height, options)
	}
	return localSolarEclipseContactPointsFromFrames(diagram.Frames, location)
}

func localSolarEclipseContactPointsFromFrames(
	frames []basic.LocalSolarEclipseDiagramFrame,
	location *time.Location,
) []LocalSolarEclipseContactPoint {
	contacts := make([]LocalSolarEclipseContactPoint, 0, 4)
	for _, frame := range frames {
		for _, label := range localSolarEclipseFrameLabels(frame) {
			switch label {
			case "C1", "C2", "C3", "C4":
				contactPA := frame.PositionAngle
				if (label == "C2" || label == "C3") && frame.MoonRadius >= frame.SunRadius {
					contactPA = normalizeSolarEclipseDegree360(contactPA + 180)
				}
				contacts = append(contacts, LocalSolarEclipseContactPoint{
					Label:                   label,
					Time:                    solarEclipseTTJDEToTime(frame.JDE, location),
					ContactPositionAngle:    contactPA,
					ContactClockwiseAngle:   normalizeSolarEclipseDegree360(360 - contactPA),
					MoonCenterPositionAngle: frame.PositionAngle,
				})
			}
		}
	}
	return contacts
}

func localSolarEclipseFrameLabels(frame basic.LocalSolarEclipseDiagramFrame) []string {
	if len(frame.Labels) > 0 {
		return frame.Labels
	}
	if frame.Label == "" {
		return nil
	}
	return []string{frame.Label}
}

func localSolarEclipseOverlapsDate(info LocalSolarEclipseInfo, dayStart, dayEnd time.Time) bool {
	eventStart, eventEnd, ok := localSolarEclipseRange(info)
	if !ok {
		return false
	}
	return !eventEnd.Before(dayStart) && eventStart.Before(dayEnd)
}

func localSolarEclipseRange(info LocalSolarEclipseInfo) (time.Time, time.Time, bool) {
	if !info.HasPartial {
		return time.Time{}, time.Time{}, false
	}
	return info.PartialStart, info.PartialEnd, true
}

func localSolarEclipseVisible(info LocalSolarEclipseInfo) bool {
	eventStart, eventEnd, ok := localSolarEclipseRange(info)
	if !ok {
		return false
	}
	return localSolarEclipseVisibleDuring(info, eventStart, eventEnd)
}

func localCentralSolarEclipseVisible(info LocalSolarEclipseInfo) bool {
	if !info.HasCentral || info.CentralStart.IsZero() || info.CentralEnd.IsZero() {
		return false
	}
	return localSolarEclipseVisibleDuring(info, info.CentralStart, info.CentralEnd)
}

func localSolarEclipseVisibleOnDate(info LocalSolarEclipseInfo, dayStart, dayEnd time.Time) bool {
	eventStart, eventEnd, ok := localSolarEclipseRange(info)
	if !ok {
		return false
	}
	segmentStart := maxLocalSolarTime(eventStart, dayStart)
	segmentEnd := minLocalSolarTime(eventEnd, dayEnd)
	if !segmentStart.Before(segmentEnd) {
		return false
	}
	return localSolarEclipseVisibleDuring(info, segmentStart, segmentEnd)
}

func localSolarEclipseVisibleDuring(info LocalSolarEclipseInfo, start, end time.Time) bool {
	if !start.Before(end) && !start.Equal(end) {
		return false
	}

	if localSolarEclipseAltitudeVisible(start, info) || localSolarEclipseAltitudeVisible(end, info) {
		return true
	}

	for dayStart, _, _ := solarEclipseLocalDayBounds(start); !dayStart.After(end); dayStart = nextSolarEclipseLocalDayStart(dayStart) {
		_, culminationSeed, _ := solarEclipseLocalDayBounds(dayStart)
		culmination := solarCulminationTime(culminationSeed, info.Longitude)
		if culmination.Before(start) || culmination.After(end) {
			continue
		}
		if localSolarEclipseAltitudeVisible(culmination, info) {
			return true
		}
	}

	return false
}

func localSolarEclipseAltitudeVisible(date time.Time, info LocalSolarEclipseInfo) bool {
	return solarAltitude(date, info.Longitude, info.Latitude) > localSolarEclipseVisibilityThreshold(info.Height, info.Latitude)
}

func localSolarEclipseVisibilityThreshold(height, latitude float64) float64 {
	if height <= 0 {
		return 0
	}
	return -basic.HeightDegreeByLat(height, latitude)
}

func normalizeSolarEclipseDegree360(angle float64) float64 {
	angle = math.Mod(angle, 360)
	if angle < 0 {
		angle += 360
	}
	return angle
}

func maxLocalSolarTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minLocalSolarTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

var (
	localSolarEclipseNASABulletinSplitK = localSolarEclipseCalculator{
		global: basic.SolarEclipseNASABulletinSplitK,
		local:  basic.LocalSolarEclipseNASABulletinSplitK,
	}
	localSolarEclipseIAUSingleK = localSolarEclipseCalculator{
		global: basic.SolarEclipseIAUSingleK,
		local:  basic.LocalSolarEclipseIAUSingleK,
	}
)
