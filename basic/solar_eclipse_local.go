package basic

import "math"

// LocalSolarEclipseResult 表示某个站点的一次站心日食几何结果。
//
// 所有时刻字段都使用力学时儒略日（JDE, TT）。
// 输入 seedJDE 只需要落在目标朔月附近，允许相差数天。
type LocalSolarEclipseResult struct {
	Model SolarEclipseRadiusModel
	Type  SolarEclipseType

	// GreatestEclipse 是站心盘面中心角距最小的时刻。
	GreatestEclipse float64

	// PartialStart / PartialEnd 是站心偏食始 / 偏食终。
	PartialStart float64
	PartialEnd   float64
	// CentralStart / CentralEnd 是站心中心食始 / 中心食终。
	//
	// 对全食对应食既 / 生光；
	// 对环食对应环食始 / 环食终。
	CentralStart float64
	CentralEnd   float64

	// Magnitude 是站心食分。
	Magnitude float64
	// Obscuration 是食甚时太阳视圆面被月面遮蔽的面积比例，范围 [0, 1]。
	Obscuration float64
	// Separation 是食甚时日月中心的站心角距，单位为度。
	Separation float64

	// SunAltitude / SunAzimuth 是食甚时太阳的站心高度角 / 方位角，单位为度。
	SunAltitude float64
	SunAzimuth  float64

	// VisibleAtGreatest 表示食甚时太阳中心在地平线上方。
	VisibleAtGreatest bool

	HasPartial bool
	HasCentral bool
	HasAnnular bool
	HasTotal   bool
}

type localSolarEclipseState struct {
	separationRad      float64
	separationSquared  float64
	sunRadiusRad       float64
	moonOuterRadiusRad float64
	moonInnerRadiusRad float64
	sunAltitudeRad     float64
	sunAzimuthRad      float64
}

const (
	localSolarEclipseGreatestWindowDays = 0.5
	localSolarEclipseGreatestTolerance  = 1e-8
	localSolarEclipseContactStepDays    = 10.0 / 1440.0
	localSolarEclipseContactSearchSteps = 72
	localSolarEclipseContactTolerance   = 1e-8
	localSolarMoonRadiusScale           = 1.0000036
)

// LocalSolarEclipse 计算给定近朔时刻附近的一次站心日食，默认使用 NASA bulletin Split-K 模型。
//
// seedJDE 为力学时儒略日（TT），只需落在目标朔月附近，允许相差数天。
// lon 为经度，东正西负；lat 为纬度，北正南负；height 为海拔高度，单位米。
func LocalSolarEclipse(seedJDE, lon, lat, height float64) LocalSolarEclipseResult {
	return LocalSolarEclipseNASABulletinSplitK(seedJDE, lon, lat, height)
}

// LocalSolarEclipseIAUSingleK 计算给定近朔时刻附近的一次站心日食，使用 IAU Single-K 模型。
func LocalSolarEclipseIAUSingleK(seedJDE, lon, lat, height float64) LocalSolarEclipseResult {
	return localSolarEclipse(seedJDE, lon, lat, height, SolarEclipseModelIAUSingleK)
}

// LocalSolarEclipseNASABulletinSplitK 计算给定近朔时刻附近的一次站心日食，使用 NASA bulletin Split-K 模型。
func LocalSolarEclipseNASABulletinSplitK(seedJDE, lon, lat, height float64) LocalSolarEclipseResult {
	return localSolarEclipse(seedJDE, lon, lat, height, SolarEclipseModelNASABulletinSplitK)
}

func localSolarEclipse(seedJDE, lonDeg, latDeg, heightMeters float64, model SolarEclipseRadiusModel) LocalSolarEclipseResult {
	newMoonJDE := CalcMoonSHByJDE(seedJDE, 0)
	lonRad := lonDeg * rad
	latRad := latDeg * rad
	heightKM := heightMeters / 1000.0
	params := solarEclipseModelParams(model)

	greatestEclipseJDE := localSolarEclipseGreatest(newMoonJDE, lonRad, latRad, heightKM, params)
	state := localSolarEclipseStateAt(greatestEclipseJDE, lonRad, latRad, heightKM, params)
	visibleThresholdRad := 0.0
	if heightMeters > 0 {
		visibleThresholdRad = -HeightDegreeByLat(heightMeters, latDeg) * rad
	}

	result := LocalSolarEclipseResult{
		Model:             model,
		Type:              SolarEclipseNone,
		GreatestEclipse:   greatestEclipseJDE,
		Separation:        state.separationRad / rad,
		SunAltitude:       state.sunAltitudeRad / rad,
		SunAzimuth:        state.sunAzimuthRad / rad,
		VisibleAtGreatest: state.sunAltitudeRad > visibleThresholdRad,
	}

	partialBoundary := state.sunRadiusRad + state.moonOuterRadiusRad
	partialGap := state.separationRad - partialBoundary
	if partialGap > 0 {
		return result
	}

	result.Type = SolarEclipsePartial
	result.HasPartial = true
	result.Magnitude = (state.moonOuterRadiusRad + state.sunRadiusRad - state.separationRad) / (2 * state.sunRadiusRad)
	if result.Magnitude < 0 {
		result.Magnitude = 0
	}
	result.Obscuration = localSolarEclipseObscuration(
		state.sunRadiusRad,
		state.moonOuterRadiusRad,
		state.separationRad,
	)

	if partialStart, ok := localSolarEclipseContact(greatestEclipseJDE, lonRad, latRad, heightKM, params, false, true); ok {
		result.PartialStart = partialStart
	}
	if partialEnd, ok := localSolarEclipseContact(greatestEclipseJDE, lonRad, latRad, heightKM, params, false, false); ok {
		result.PartialEnd = partialEnd
	}

	centralBoundary := math.Abs(state.sunRadiusRad - state.moonInnerRadiusRad)
	if state.separationRad > centralBoundary {
		return result
	}

	result.HasCentral = true
	if state.moonInnerRadiusRad >= state.sunRadiusRad {
		result.Type = SolarEclipseTotal
		result.HasTotal = true
	} else {
		result.Type = SolarEclipseAnnular
		result.HasAnnular = true
	}
	result.Magnitude = state.moonInnerRadiusRad / state.sunRadiusRad

	if centralStart, ok := localSolarEclipseContact(greatestEclipseJDE, lonRad, latRad, heightKM, params, true, true); ok {
		result.CentralStart = centralStart
	}
	if centralEnd, ok := localSolarEclipseContact(greatestEclipseJDE, lonRad, latRad, heightKM, params, true, false); ok {
		result.CentralEnd = centralEnd
	}

	return result
}

func solarEclipseModelParams(model SolarEclipseRadiusModel) solarEclipseModelParameters {
	params := solarEclipseModelParameters{
		penumbralK: solarEclipsePenumbralK,
		umbralK:    solarEclipsePenumbralK,
	}
	if model == SolarEclipseModelNASABulletinSplitK {
		params.umbralK = solarEclipseUmbralK
	}
	return params
}

func localSolarEclipseGreatest(
	newMoonJDE, lonRad, latRad, heightKM float64,
	params solarEclipseModelParameters,
) float64 {
	left := newMoonJDE - localSolarEclipseGreatestWindowDays
	right := newMoonJDE + localSolarEclipseGreatestWindowDays
	goldenRatio := (math.Sqrt(5) - 1) / 2

	x1 := right - goldenRatio*(right-left)
	x2 := left + goldenRatio*(right-left)
	f1 := localSolarEclipseStateAt(x1, lonRad, latRad, heightKM, params).separationSquared
	f2 := localSolarEclipseStateAt(x2, lonRad, latRad, heightKM, params).separationSquared

	for i := 0; i < 80 && right-left > localSolarEclipseGreatestTolerance; i++ {
		if f1 <= f2 {
			right = x2
			x2 = x1
			f2 = f1
			x1 = right - goldenRatio*(right-left)
			f1 = localSolarEclipseStateAt(x1, lonRad, latRad, heightKM, params).separationSquared
			continue
		}

		left = x1
		x1 = x2
		f1 = f2
		x2 = left + goldenRatio*(right-left)
		f2 = localSolarEclipseStateAt(x2, lonRad, latRad, heightKM, params).separationSquared
	}

	return (left + right) / 2
}

func localSolarEclipseContact(
	greatestEclipseJDE, lonRad, latRad, heightKM float64,
	params solarEclipseModelParameters,
	central bool,
	beforeGreatest bool,
) (float64, bool) {
	centerGap := localSolarEclipseGap(greatestEclipseJDE, lonRad, latRad, heightKM, params, central)
	if centerGap > 0 {
		return 0, false
	}
	if math.Abs(centerGap) <= 1e-14 {
		return greatestEclipseJDE, true
	}

	direction := 1.0
	if beforeGreatest {
		direction = -1.0
	}

	previousJDE := greatestEclipseJDE
	for i := 1; i <= localSolarEclipseContactSearchSteps; i++ {
		currentJDE := greatestEclipseJDE + direction*localSolarEclipseContactStepDays*float64(i)
		currentGap := localSolarEclipseGap(currentJDE, lonRad, latRad, heightKM, params, central)
		if currentGap >= 0 {
			left := previousJDE
			right := currentJDE
			if beforeGreatest {
				left = currentJDE
				right = previousJDE
			}
			return localSolarEclipseContactBisection(left, right, lonRad, latRad, heightKM, params, central)
		}
		previousJDE = currentJDE
	}

	return 0, false
}

func localSolarEclipseContactBisection(
	leftJDE, rightJDE, lonRad, latRad, heightKM float64,
	params solarEclipseModelParameters,
	central bool,
) (float64, bool) {
	leftGap := localSolarEclipseGap(leftJDE, lonRad, latRad, heightKM, params, central)
	rightGap := localSolarEclipseGap(rightJDE, lonRad, latRad, heightKM, params, central)
	if leftGap == 0 {
		return leftJDE, true
	}
	if rightGap == 0 {
		return rightJDE, true
	}
	if leftGap*rightGap > 0 {
		return 0, false
	}

	for i := 0; i < 80 && rightJDE-leftJDE > localSolarEclipseContactTolerance; i++ {
		midJDE := (leftJDE + rightJDE) / 2
		midGap := localSolarEclipseGap(midJDE, lonRad, latRad, heightKM, params, central)
		if leftGap*midGap > 0 {
			leftJDE = midJDE
			leftGap = midGap
			continue
		}
		rightJDE = midJDE
		rightGap = midGap
	}

	return (leftJDE + rightJDE) / 2, true
}

func localSolarEclipseGap(
	jdTT, lonRad, latRad, heightKM float64,
	params solarEclipseModelParameters,
	central bool,
) float64 {
	state := localSolarEclipseStateAt(jdTT, lonRad, latRad, heightKM, params)
	boundary := state.sunRadiusRad + state.moonOuterRadiusRad
	if central {
		boundary = math.Abs(state.sunRadiusRad - state.moonInnerRadiusRad)
	}
	return state.separationRad - boundary
}

func localSolarEclipseStateAt(
	jdTT, lonRad, latRad, heightKM float64,
	params solarEclipseModelParameters,
) localSolarEclipseState {
	sunEquatorial, moonEquatorial := solarEclipseSunMoonEquatorial(jdTT)
	sunXYZ := solarEclipseLLRToXYZ(sunEquatorial[0], sunEquatorial[1], sunEquatorial[2])
	moonXYZ := solarEclipseLLRToXYZ(moonEquatorial[0], moonEquatorial[1], moonEquatorial[2])

	utJDE := TD2UT(jdTT, false)
	gst := ApparentSiderealTime(utJDE) * 15 * rad
	observerXYZ := localSolarEclipseObserverXYZ(gst, lonRad, latRad, heightKM)

	sunTopocentric := solarEclipseXYZToLLR(
		sunXYZ[0]-observerXYZ[0],
		sunXYZ[1]-observerXYZ[1],
		sunXYZ[2]-observerXYZ[2],
	)
	moonTopocentric := solarEclipseXYZToLLR(
		moonXYZ[0]-observerXYZ[0],
		moonXYZ[1]-observerXYZ[1],
		moonXYZ[2]-observerXYZ[2],
	)

	sunUnit := solarEclipseLLRToXYZ(sunTopocentric[0], sunTopocentric[1], 1)
	moonUnit := solarEclipseLLRToXYZ(moonTopocentric[0], moonTopocentric[1], 1)
	dot := sunUnit[0]*moonUnit[0] + sunUnit[1]*moonUnit[1] + sunUnit[2]*moonUnit[2]
	if dot > 1 {
		dot = 1
	}
	if dot < -1 {
		dot = -1
	}

	sunRadiusRad := math.Asin(localSolarEclipseClampUnit(
		solarEclipseEarthEquatorialRadiusKM * solarEclipseSolarRadiusRatio / sunTopocentric[2],
	))
	moonOuterRadiusRad := math.Asin(localSolarEclipseClampUnit(
		solarEclipseEarthEquatorialRadiusKM * solarEclipsePenumbralK * localSolarMoonRadiusScale / moonTopocentric[2],
	))
	moonInnerRadiusRad := math.Asin(localSolarEclipseClampUnit(
		solarEclipseEarthEquatorialRadiusKM * params.umbralK * localSolarMoonRadiusScale / moonTopocentric[2],
	))

	sunHorizontal := solarEclipseEquatorialToHorizontal(
		sunTopocentric[0],
		sunTopocentric[1],
		sunTopocentric[2],
		lonRad,
		latRad,
		gst,
	)

	return localSolarEclipseState{
		separationRad:      math.Acos(dot),
		separationSquared:  2 - 2*dot,
		sunRadiusRad:       sunRadiusRad,
		moonOuterRadiusRad: moonOuterRadiusRad,
		moonInnerRadiusRad: moonInnerRadiusRad,
		sunAltitudeRad:     sunHorizontal[1],
		sunAzimuthRad:      solarEclipseNormalizeRadians(sunHorizontal[0] + math.Pi),
	}
}

func localSolarEclipseObserverXYZ(gst, lonRad, latRad, heightKM float64) [3]float64 {
	equatorialRadius := solarEclipseEarthEquatorialRadiusKM
	polarRadius := solarEclipseEarthEquatorialRadiusKM * solarEclipseEarthPolarRatio

	u := math.Atan((polarRadius / equatorialRadius) * math.Tan(latRad))
	radiusXY := equatorialRadius*math.Cos(u) + heightKM*math.Cos(latRad)
	radiusZ := polarRadius*math.Sin(u) + heightKM*math.Sin(latRad)
	angle := gst + lonRad

	return [3]float64{
		radiusXY * math.Cos(angle),
		radiusXY * math.Sin(angle),
		radiusZ,
	}
}

func localSolarEclipseObscuration(sunRadius, moonRadius, separation float64) float64 {
	if separation >= sunRadius+moonRadius {
		return 0
	}

	sunArea := math.Pi * sunRadius * sunRadius
	if separation <= math.Abs(sunRadius-moonRadius) {
		if moonRadius >= sunRadius {
			return 1
		}
		return moonRadius * moonRadius / (sunRadius * sunRadius)
	}

	partSun := localSolarEclipseClampUnit((separation*separation + sunRadius*sunRadius - moonRadius*moonRadius) / (2 * separation * sunRadius))
	partMoon := localSolarEclipseClampUnit((separation*separation + moonRadius*moonRadius - sunRadius*sunRadius) / (2 * separation * moonRadius))
	term := (-separation + sunRadius + moonRadius) *
		(separation + sunRadius - moonRadius) *
		(separation - sunRadius + moonRadius) *
		(separation + sunRadius + moonRadius)
	if term < 0 {
		term = 0
	}

	overlapArea := sunRadius*sunRadius*math.Acos(partSun) +
		moonRadius*moonRadius*math.Acos(partMoon) -
		0.5*math.Sqrt(term)

	return overlapArea / sunArea
}

func localSolarEclipseClampUnit(value float64) float64 {
	if value > 1 {
		return 1
	}
	if value < -1 {
		return -1
	}
	return value
}
