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

func (state localSolarEclipseState) movingDiskContactState() movingDiskContactState {
	return movingDiskContactState{
		separation:           state.separationRad,
		occultingOuterRadius: state.moonOuterRadiusRad,
		occultingInnerRadius: state.moonInnerRadiusRad,
		targetRadius:         state.sunRadiusRad,
		valid: movingDiskContactStateValid(
			state.separationRad,
			state.moonOuterRadiusRad,
			state.moonInnerRadiusRad,
			state.sunRadiusRad,
		),
	}
}

type localSolarEclipseStateContext struct {
	sunXYZ  [3]float64
	moonXYZ [3]float64
	gst     float64
	params  solarEclipseModelParameters
	// generation 记录写入缓存时的 ΔT 世代（gst 依赖 ΔT，覆盖 ΔT 后条目必须失效）。
	// generation is the ΔT generation the cached context was built under; gst depends on
	// ΔT, so an override must invalidate the entry.
	generation uint64
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
	contactEvaluator := newLocalSolarEclipseContactEvaluator(lonRad, latRad, heightKM, params)
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

	contactState := state.movingDiskContactState()
	contactEvaluator.prime(greatestEclipseJDE, contactState, contactState.valid)
	partialGap := contactState.externalContactGap()
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

	if partialStart, ok := localSolarEclipseContactWithEvaluator(greatestEclipseJDE, contactEvaluator, false, true); ok {
		result.PartialStart = partialStart
	}
	if partialEnd, ok := localSolarEclipseContactWithEvaluator(greatestEclipseJDE, contactEvaluator, false, false); ok {
		result.PartialEnd = partialEnd
	}

	if contactState.internalContactGap() > 0 {
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

	if centralStart, ok := localSolarEclipseContactWithEvaluator(greatestEclipseJDE, contactEvaluator, true, true); ok {
		result.CentralStart = centralStart
	}
	if centralEnd, ok := localSolarEclipseContactWithEvaluator(greatestEclipseJDE, contactEvaluator, true, false); ok {
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
	return localSolarEclipseGreatestWith(newMoonJDE, func(jd float64) localSolarEclipseState {
		return localSolarEclipseStateAt(jd, lonRad, latRad, heightKM, params)
	})
}

func localSolarEclipseGreatestWith(
	newMoonJDE float64,
	stateAt func(float64) localSolarEclipseState,
) float64 {
	left := newMoonJDE - localSolarEclipseGreatestWindowDays
	right := newMoonJDE + localSolarEclipseGreatestWindowDays
	return solarEclipseMovingDiskEngine().greatest(
		newMoonJDE, left, right,
		func(jd float64) (float64, bool) {
			return stateAt(jd).separationSquared, true
		},
		80,
	)
}

// centralPhaseDurationDaysAt 用事件局部插值星历求解某点的中心相时长（日）：先求局部食甚，
// 再解本影/反本影的内切接触。没有中心相或接触退化时返回 0。
func (solver solarEclipseSolver) centralPhaseDurationDaysAt(jd, lonDeg, latDeg float64) float64 {
	solver = solver.withLocalEphemeris()
	lonRad, latRad := lonDeg*rad, latDeg*rad
	stateAt := func(jd float64) localSolarEclipseState {
		return solver.localStateContextCandidateAt(jd).stateAt(lonRad, latRad, 0)
	}
	greatestJDE := localSolarEclipseGreatestWith(solver.newMoonJDE, stateAt)
	contactState := stateAt(greatestJDE).movingDiskContactState()
	if !contactState.valid || contactState.internalContactGap() > 0 {
		return 0
	}
	evaluator := newMovingDiskContactEvaluator(func(jd float64) (movingDiskContactState, bool) {
		contact := stateAt(jd).movingDiskContactState()
		return contact, contact.valid
	})
	evaluator.prime(greatestJDE, contactState, true)
	start, ok := localSolarEclipseContactWithEvaluator(greatestJDE, evaluator, true, true)
	if !ok {
		return 0
	}
	end, ok := localSolarEclipseContactWithEvaluator(greatestJDE, evaluator, true, false)
	if !ok || end <= start {
		return 0
	}
	return end - start
}

func newLocalSolarEclipseContactEvaluator(
	lonRad, latRad, heightKM float64,
	params solarEclipseModelParameters,
) *movingDiskContactEvaluator {
	return newMovingDiskContactEvaluator(func(jdTT float64) (movingDiskContactState, bool) {
		state := localSolarEclipseStateAt(jdTT, lonRad, latRad, heightKM, params)
		contactState := state.movingDiskContactState()
		return contactState, contactState.valid
	})
}

func localSolarEclipseContactWithEvaluator(
	greatestEclipseJDE float64,
	evaluator *movingDiskContactEvaluator,
	central bool,
	beforeGreatest bool,
) (float64, bool) {
	centerGap, centerOK := evaluator.gap(greatestEclipseJDE, central)
	if !centerOK {
		return 0, false
	}
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
	return solarEclipseMovingDiskEngine().contactRoot(
		greatestEclipseJDE,
		direction,
		localSolarEclipseContactStepDays,
		localSolarEclipseContactStepDays*localSolarEclipseContactSearchSteps,
		localSolarEclipseContactTolerance,
		func(jd float64) (float64, bool) {
			return evaluator.gap(jd, central)
		},
		80,
	)
}

func localSolarEclipseStateAt(
	jdTT, lonRad, latRad, heightKM float64,
	params solarEclipseModelParameters,
) localSolarEclipseState {
	return newLocalSolarEclipseStateContext(jdTT, params).stateAt(lonRad, latRad, heightKM)
}

func newLocalSolarEclipseStateContext(
	jdTT float64,
	params solarEclipseModelParameters,
) localSolarEclipseStateContext {
	return newLocalSolarEclipseStateContextWithOverride(jdTT, math.NaN(), params)
}

// newLocalSolarEclipseStateContextWithOverride 用显式 ΔT 构造站心状态上下文：TT 时刻不变，
// ΔT 只决定恒星时相位；deltaTSeconds 为 NaN 时走进程级模型。
func newLocalSolarEclipseStateContextWithOverride(
	jdTT, deltaTSeconds float64,
	params solarEclipseModelParameters,
) localSolarEclipseStateContext {
	sunEquatorial, moonEquatorial := solarEclipseSunMoonEquatorial(jdTT)
	utJDE := TD2UT(jdTT, false)
	if !math.IsNaN(deltaTSeconds) {
		utJDE = jdTT - deltaTSeconds/86400
	}
	return localSolarEclipseStateContext{
		sunXYZ:  solarEclipseLLRToXYZ(sunEquatorial[0], sunEquatorial[1], sunEquatorial[2]),
		moonXYZ: solarEclipseLLRToXYZ(moonEquatorial[0], moonEquatorial[1], moonEquatorial[2]),
		gst:     ApparentSiderealTime(utJDE) * 15 * rad,
		params:  params,
	}
}

func (context localSolarEclipseStateContext) stateAt(lonRad, latRad, heightKM float64) localSolarEclipseState {
	observerXYZ := localSolarEclipseObserverXYZ(context.gst, lonRad, latRad, heightKM)

	sunTopocentric := solarEclipseXYZToLLR(
		context.sunXYZ[0]-observerXYZ[0],
		context.sunXYZ[1]-observerXYZ[1],
		context.sunXYZ[2]-observerXYZ[2],
	)
	moonTopocentric := solarEclipseXYZToLLR(
		context.moonXYZ[0]-observerXYZ[0],
		context.moonXYZ[1]-observerXYZ[1],
		context.moonXYZ[2]-observerXYZ[2],
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
		solarEclipseEarthEquatorialRadiusKM * context.params.umbralK * localSolarMoonRadiusScale / moonTopocentric[2],
	))

	sunHorizontal := solarEclipseEquatorialToHorizontal(
		sunTopocentric[0],
		sunTopocentric[1],
		sunTopocentric[2],
		lonRad,
		latRad,
		context.gst,
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
