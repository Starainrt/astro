package basic

import "math"

// solarEclipseLimbMaxShiftDays 是全球偏食始末月缘修正的上限（2 分钟），超出视为不可信并保留圆月面结果。
const solarEclipseLimbMaxShiftDays = 120.0 / 86400.0

// solarEclipseLimbDerivativeStepDays 是求外切间隙时间导数的差分半步长（10 秒）。
const solarEclipseLimbDerivativeStepDays = 10.0 / 86400.0

// SolarEclipseLimbContact 月缘修正后的全球偏食始末接触点，与修正后的时刻同源 / limb-corrected global contact location.
type SolarEclipseLimbContact struct {
	// Defined 为真时经纬度有效；月缘未生效时保持 false 并且经纬度为零值 / the location is valid.
	Defined bool
	// Longitude 经度，东正西负，单位度 / longitude in degrees, east positive.
	Longitude float64
	// Latitude 纬度，北正南负，单位度 / latitude in degrees, north positive.
	Latitude float64
}

// withLunarLimb 在圆月面全球结果上按月缘修正偏食始末（P1/P4）与食甚点中心食时长；中心线始末、食甚、食分与食型不变，
// 本函数不改带宽，月缘带宽增量由路径层在成对横截面上叠加。
func (solver solarEclipseSolver) withLunarLimb(result SolarEclipseResult, limb LunarLimb) SolarEclipseResult {
	if limb == nil || result.Type == SolarEclipseNone {
		return result
	}
	limbSolver := solver.withLimbParams(limb)

	original := result
	if result.HasPartial {
		result.PartialBeginOnEarth, result.PartialBeginContact = solver.limbPartialContactPoint(limbSolver, result.PartialBeginOnEarth)
		result.PartialEndOnEarth, result.PartialEndContact = solver.limbPartialContactPoint(limbSolver, result.PartialEndOnEarth)
	}
	if result.HasCentral && result.GreatestEclipse > 0 {
		// 剖面在任一求解状态上无值时，半径已退回圆月面，保留圆月面时长而不是采纳离散差。
		if duration, limbOK := limbSolver.centralPhaseDurationDaysWithLimbAt(
			result.GreatestEclipse, result.GreatestLongitude, result.GreatestLatitude,
		); limbOK && duration > 0 {
			result.CentralDurationDays = duration
		}
	}
	result.LunarLimbApplied = solarEclipseLimbChanged(result, original)
	// 时刻未真正改变时位置也不该被声明，避免出图标记用上噪声级的平移。
	if !result.LunarLimbApplied {
		result.PartialBeginContact, result.PartialEndContact = SolarEclipseLimbContact{}, SolarEclipseLimbContact{}
	}
	return result
}

// limbPartialContactPoint 在修正后的时刻重解半影与地球的外切点，使接触点位置与新时刻同源。
func (solver solarEclipseSolver) limbPartialContactPoint(
	limbSolver solarEclipseSolver,
	contactJDE float64,
) (float64, SolarEclipseLimbContact) {
	jde, ok := solver.limbPartialContact(limbSolver, contactJDE)
	if !ok {
		return jde, SolarEclipseLimbContact{}
	}
	point, pointOK := solver.shadowContactPointAt(jde, solarEclipsePenumbralShadow, false)
	if !pointOK {
		return jde, SolarEclipseLimbContact{}
	}
	return jde, SolarEclipseLimbContact{Defined: true, Longitude: point.Longitude, Latitude: point.Latitude}
}

func solarEclipseLimbChanged(result, original SolarEclipseResult) bool {
	for _, pair := range [][2]float64{
		{result.PartialBeginOnEarth, original.PartialBeginOnEarth},
		{result.PartialEndOnEarth, original.PartialEndOnEarth},
		{result.CentralDurationDays, original.CentralDurationDays},
	} {
		if limbTimeChanged(pair[0], pair[1]) {
			return true
		}
	}
	return false
}

// solarEclipseLimbSiteSpanRad 与 solarEclipseLimbSiteStepRad 是切点两侧沿地球边缘扫描的半宽与步长：
// 月缘起伏约 ±5 km，而偏离切点 δ 时影锥外距约 R⊕·δ²/2，δ 超过约 3° 后不可能更早接触。
const (
	solarEclipseLimbSiteSpanRad = 3.0 * math.Pi / 180
	solarEclipseLimbSiteStepRad = 0.1 * math.Pi / 180
)

// limbPartialContact 修正一次全球偏食始末，P1 取最早、P4 取最晚；第二个返回值报告是否取到月缘解。
func (solver solarEclipseSolver) limbPartialContact(limbSolver solarEclipseSolver, contactJDE float64) (float64, bool) {
	if !finite(contactJDE) || contactJDE <= 0 {
		return contactJDE, false
	}
	moon := solver.besselMoonAt(contactJDE)
	norm := math.Hypot(moon[0], moon[1])
	if !finite(norm) || norm == 0 {
		return contactJDE, false
	}
	axis := solver.besselAxisAt(contactJDE)
	center := math.Atan2(moon[1], moon[0])
	siteAt := func(angle float64) (float64, float64, bool) {
		x, y := math.Cos(angle), math.Sin(angle)
		// 与 quickContactAt 同口径的椭球有效半径，再稍向内收保证视线与椭球相交。
		effective := 1 - (1/solarEclipseEarthPolarRatioSquared-1)*y*y/2
		lonDeg, latDeg, ok := solarEclipseBesselXYToGeodetic(x*effective*0.999, y*effective*0.999, axis, true)
		return lonDeg * rad, latDeg * rad, ok
	}
	gapAt := func(s solarEclipseSolver, lonRad, latRad, jde float64) float64 {
		return s.localStateContextAt(jde).stateAt(lonRad, latRad, 0).movingDiskContactState().externalContactGap()
	}
	// 剖面在该点无值时半径退回圆月面，这样的点不能参与修正，否则圆月面离散误差会冒充月缘修正。
	limbGapAt := func(lonRad, latRad, jde float64) (float64, bool) {
		state := limbSolver.localStateContextAt(jde).stateAt(lonRad, latRad, 0)
		return state.movingDiskContactState().externalContactGap(), state.outerLimbApplied
	}
	lon0, lat0, ok := siteAt(center)
	if !ok {
		return contactJDE, false
	}
	reference := gapAt(solver, lon0, lat0, contactJDE)
	after, okAfter := limbGapAt(lon0, lat0, contactJDE+solarEclipseLimbDerivativeStepDays)
	before, okBefore := limbGapAt(lon0, lat0, contactJDE-solarEclipseLimbDerivativeStepDays)
	derivative := (after - before) / (2 * solarEclipseLimbDerivativeStepDays)
	if !okAfter || !okBefore || !finite(reference) || !finite(derivative) || derivative == 0 {
		return contactJDE, false
	}
	// derivative < 0 表示间隙随时间缩小（P1），此时取最早接触；否则取最晚（P4）。
	earliest := derivative < 0
	best, found := 0.0, false
	steps := int(math.Round(solarEclipseLimbSiteSpanRad / solarEclipseLimbSiteStepRad))
	for i := -steps; i <= steps; i++ {
		lonRad, latRad, ok := siteAt(center + float64(i)*solarEclipseLimbSiteStepRad)
		if !ok {
			continue
		}
		gap, limbOK := limbGapAt(lonRad, latRad, contactJDE)
		if !limbOK || !finite(gap) {
			continue
		}
		shift := -(gap - reference) / derivative
		if !found || (earliest && shift < best) || (!earliest && shift > best) {
			best, found = shift, true
		}
	}
	if !found || !finite(best) || math.Abs(best) > solarEclipseLimbMaxShiftDays {
		return contactJDE, false
	}
	return contactJDE + best, true
}

// withLimbParams 返回带月缘、独立状态缓存的求解器副本；原求解器的圆月面缓存不受影响。
func (solver solarEclipseSolver) withLimbParams(limb LunarLimb) solarEclipseSolver {
	limbSolver := solver
	limbSolver.params.lunarLimb = limb
	limbSolver.localStateContextCache = make(map[uint64]localSolarEclipseStateContext)
	limbSolver.limbAxes = nil
	return limbSolver
}
