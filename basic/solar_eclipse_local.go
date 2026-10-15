package basic

import (
	"math"

	"github.com/starainrt/astro/internal/geodata"
)

// LocalSolarEclipseResult 表示某个站点的一次站心日食几何结果。
//
// 所有时刻字段都使用力学时儒略日（JDE, TT）。
// 输入 seedJDE 只需要落在目标朔月附近，允许相差数天。
type LocalSolarEclipseResult struct {
	// 下列字段是决定上述数值的口径，随结果一起保留。
	// The fields below are the conventions that fix the numbers above.
	Model          SolarEclipseRadiusModel
	SunRadiusModel SolarEclipseSunRadiusModel
	// LunarLimbApplied 表示接触时刻已按月缘剖面修正 / reports whether lunar-limb refinement changed the contacts.
	LunarLimbApplied bool
	Type             SolarEclipseType

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
	// innerTotal 为真表示内切按全食求、食型为全食；有月缘时取全食、环食两支中内切间隙较小者，不能由被钳过的等效内半径反推。
	innerTotal     bool
	sunAltitudeRad float64
	sunAzimuthRad  float64
	// limbApplied 表示内外半径都取自月缘剖面；outerLimbApplied 只看外切半径，偏食界线与全球始末只需要它。
	limbApplied      bool
	outerLimbApplied bool
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
	// 仅设置月缘时非 nil；用指针让圆月面路径的缓存条目保持免分配。
	limbFrame *lunarLimbFrame
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

// LocalSolarEclipse 计算给定近朔时刻附近的一次站心日食，默认使用 NASA bulletin Split-K 模型与标准太阳半径 / one local solar eclipse near a new moon.
// seedJDE 为力学时儒略日（TT），落在目标朔月附近即可；height 按椭球高（大地高）解读，只有正高 H 时须先加 N。
func LocalSolarEclipse(seedJDE, lon, lat, height float64) LocalSolarEclipseResult {
	return LocalSolarEclipseNASABulletinSplitK(seedJDE, lon, lat, height)
}

// LocalSolarEclipseWithOptionsWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LocalSolarEclipseWithOptionsWithInstance(seedJDE, lon, lat, height float64, options SolarEclipseOptions, inst *Instance) LocalSolarEclipseResult {
	return localSolarEclipse(seedJDE, lon, lat, height, options, inst)
}

// LocalSolarEclipseIAUSingleK 计算给定近朔时刻附近的一次站心日食，使用 IAU Single-K 模型。
func LocalSolarEclipseIAUSingleK(seedJDE, lon, lat, height float64) LocalSolarEclipseResult {
	return localSolarEclipse(seedJDE, lon, lat, height, SolarEclipseOptions{RadiusModel: SolarEclipseModelIAUSingleK})
}

// LocalSolarEclipseNASABulletinSplitK 计算给定近朔时刻附近的一次站心日食，使用 NASA bulletin Split-K 模型。
func LocalSolarEclipseNASABulletinSplitK(seedJDE, lon, lat, height float64) LocalSolarEclipseResult {
	return localSolarEclipse(seedJDE, lon, lat, height, SolarEclipseOptions{RadiusModel: SolarEclipseModelNASABulletinSplitK})
}

func localSolarEclipse(seedJDE, lonDeg, latDeg, heightMeters float64, options SolarEclipseOptions, instances ...*Instance) LocalSolarEclipseResult {
	if options.LunarLimb != nil {
		circularOptions := options
		circularOptions.LunarLimb = nil
		circular := localSolarEclipseCore(seedJDE, lonDeg, latDeg, heightMeters, circularOptions, instances...)
		result := localSolarEclipseCore(seedJDE, lonDeg, latDeg, heightMeters, options, instances...)
		result.LunarLimbApplied = localSolarEclipseLimbChanged(result, circular)
		return result
	}
	return localSolarEclipseCore(seedJDE, lonDeg, latDeg, heightMeters, options, instances...)
}

// localSolarEclipseLimbChanged 报告月缘是否改变了站心结果中任何对外可见的数值：
// 四个接触时刻、食甚时的食分与遮蔽率，以及站心食型。食甚时刻本身按圆月面最近距离求，不参与比较。
func localSolarEclipseLimbChanged(result, circular LocalSolarEclipseResult) bool {
	if result.Type == SolarEclipseNone || circular.Type == SolarEclipseNone {
		return result.Type != circular.Type
	}
	if result.Type != circular.Type {
		return true
	}
	for _, pair := range [][2]float64{
		{result.PartialStart, circular.PartialStart},
		{result.PartialEnd, circular.PartialEnd},
		{result.CentralStart, circular.CentralStart},
		{result.CentralEnd, circular.CentralEnd},
	} {
		if limbTimeChanged(pair[0], pair[1]) {
			return true
		}
	}
	for _, pair := range [][2]float64{
		{result.Magnitude, circular.Magnitude},
		{result.Obscuration, circular.Obscuration},
	} {
		if limbMagnitudeChanged(pair[0], pair[1]) {
			return true
		}
	}
	return false
}

func localSolarEclipseCore(seedJDE, lonDeg, latDeg, heightMeters float64, options SolarEclipseOptions, instances ...*Instance) LocalSolarEclipseResult {
	inst := instanceValue(instances)

	newMoonJDE := solarEclipseAnchorNewMoon(seedJDE, inst)
	lonRad := lonDeg * rad
	latRad := latDeg * rad
	heightKM := heightMeters / 1000.0
	options.RadiusModel = normalizeSolarEclipseRadiusModel(options.RadiusModel)
	options.SunRadiusModel = normalizeSolarEclipseSunRadiusModel(options.SunRadiusModel)
	params := solarEclipseModelParams(options.RadiusModel, options.SunRadiusModel)
	params.lunarLimb = options.LunarLimb

	greatestEclipseJDE := localSolarEclipseGreatest(newMoonJDE, lonRad, latRad, heightKM, params, inst)
	state := localSolarEclipseStateAt(greatestEclipseJDE, lonRad, latRad, heightKM, params, inst)
	contactEvaluator := newLocalSolarEclipseContactEvaluator(lonRad, latRad, heightKM, params, inst)
	visibleThresholdRad := 0.0
	if heightMeters > 0 {
		visibleThresholdRad = -HeightDegreeByLat(heightMeters, latDeg) * rad
	}

	result := LocalSolarEclipseResult{
		Model:             options.RadiusModel,
		SunRadiusModel:    options.SunRadiusModel,
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
	if state.innerTotal {
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

// solarEclipsePenumbralRadiusNumerator 是半影外半径公式的分子（地球赤道半径 × k1 × 月亮半径尺度）。
// Split-K 的 k1 是编译期常量，这里保持常量折叠后的末位，使该模型的既有输出逐字节不变；IAU Single-K 只能走运行时值。
func solarEclipsePenumbralRadiusNumerator(params solarEclipseModelParameters) float64 {
	if params.penumbralK == solarEclipsePenumbralK {
		return geodata.EarthEquatorialRadiusKM * solarEclipsePenumbralK * localSolarMoonRadiusScale
	}
	return geodata.EarthEquatorialRadiusKM * params.penumbralK * localSolarMoonRadiusScale
}

func solarEclipseModelParams(model SolarEclipseRadiusModel, sunRadiusModel SolarEclipseSunRadiusModel) solarEclipseModelParameters {
	params := solarEclipseModelParameters{
		penumbralK:     solarEclipsePenumbralK,
		umbralK:        solarEclipseUmbralK,
		sunRadiusRatio: solarEclipseSunRadiusRatio(sunRadiusModel),
	}
	if model == SolarEclipseModelIAUSingleK {
		params.penumbralK = solarEclipseIAUSingleRadiusK
		params.umbralK = solarEclipseIAUSingleRadiusK
	}
	return params
}

func localSolarEclipseGreatest(
	newMoonJDE, lonRad, latRad, heightKM float64,
	params solarEclipseModelParameters, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	// 食甚只看日月中心角距，与月缘无关，免去每步的全方向月缘扫描。
	params.lunarLimb = nil
	return localSolarEclipseGreatestWith(newMoonJDE, func(jd float64) localSolarEclipseState {
		return localSolarEclipseStateAt(jd, lonRad, latRad, heightKM, params, inst)
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
func (solver solarEclipseSolver) centralPhaseDurationDaysAt(jde, lonDeg, latDeg float64) float64 {
	duration, _ := solver.centralPhaseDurationDaysWithLimbAt(jde, lonDeg, latDeg)
	return duration
}

// centralPhaseDurationDaysWithLimbAt 同 centralPhaseDurationDaysAt，并报告求解用到的每个状态是否都取到了月缘半径。
func (solver solarEclipseSolver) centralPhaseDurationDaysWithLimbAt(jde, lonDeg, latDeg float64) (float64, bool) {
	solver = solver.withLocalEphemeris()
	lonRad, latRad := lonDeg*rad, latDeg*rad
	limbOK := solver.params.lunarLimb != nil
	stateAt := func(jde float64) localSolarEclipseState {
		state := solver.localStateContextCandidateAt(jde).stateAt(lonRad, latRad, 0)
		if !state.limbApplied {
			limbOK = false
		}
		return state
	}
	greatestJDE := localSolarEclipseGreatestWith(solver.newMoonJDE, func(jde float64) localSolarEclipseState {
		context := solver.localStateContextCandidateAt(jde)
		context.limbFrame = nil
		return context.stateAt(lonRad, latRad, 0)
	})
	contactState := stateAt(greatestJDE).movingDiskContactState()
	if !contactState.valid || contactState.internalContactGap() > 0 {
		return 0, false
	}
	evaluator := newMovingDiskContactEvaluator(func(jde float64) (movingDiskContactState, bool) {
		contact := stateAt(jde).movingDiskContactState()
		return contact, contact.valid
	})
	evaluator.prime(greatestJDE, contactState, true)
	start, ok := localSolarEclipseContactWithEvaluator(greatestJDE, evaluator, true, true)
	if !ok {
		return 0, false
	}
	end, ok := localSolarEclipseContactWithEvaluator(greatestJDE, evaluator, true, false)
	if !ok || end <= start {
		return 0, false
	}
	return end - start, limbOK
}

func newLocalSolarEclipseContactEvaluator(
	lonRad, latRad, heightKM float64,
	params solarEclipseModelParameters, instances ...*Instance) *movingDiskContactEvaluator {
	inst := instanceValue(instances)

	return newMovingDiskContactEvaluator(func(jdTT float64) (movingDiskContactState, bool) {
		state := localSolarEclipseStateAt(jdTT, lonRad, latRad, heightKM, params, inst)
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
	params solarEclipseModelParameters, instances ...*Instance) localSolarEclipseState {
	inst := instanceValue(instances)

	return newLocalSolarEclipseStateContext(jdTT, params, inst).stateAt(lonRad, latRad, heightKM)
}

func newLocalSolarEclipseStateContext(
	jdTT float64,
	params solarEclipseModelParameters, instances ...*Instance) localSolarEclipseStateContext {
	inst := instanceValue(instances)

	return newLocalSolarEclipseStateContextWithOverride(jdTT, math.NaN(), params, inst)
}

// newLocalSolarEclipseStateContextWithOverride 用显式 ΔT 构造站心状态上下文：TT 时刻不变，
// ΔT 只决定恒星时相位；deltaTSeconds 为 NaN 时走进程级模型。
func newLocalSolarEclipseStateContextWithOverride(
	jdTT, deltaTSeconds float64,
	params solarEclipseModelParameters, instances ...*Instance) localSolarEclipseStateContext {
	inst := instanceValue(instances)

	sunEquatorial, moonEquatorial := solarEclipseSunMoonEquatorial(jdTT, inst)
	ut1JDE := TT2UT1(jdTT)
	if !math.IsNaN(deltaTSeconds) {
		ut1JDE = jdTT - deltaTSeconds/86400
	}
	context := localSolarEclipseStateContext{
		sunXYZ:  solarEclipseLLRToXYZ(sunEquatorial[0], sunEquatorial[1], sunEquatorial[2]),
		moonXYZ: solarEclipseLLRToXYZ(moonEquatorial[0], moonEquatorial[1], moonEquatorial[2]),
		gst:     ApparentSiderealTime(ut1JDE) * 15 * rad,
		params:  params,
	}
	if params.lunarLimb != nil {
		if frame := newLunarLimbFrame(params.lunarLimb, jdTT, context.moonXYZ, inst); frame.valid {
			context.limbFrame = &frame
		}
	}
	return context
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
		geodata.EarthEquatorialRadiusKM * context.params.sunRadiusRatio / sunTopocentric[2],
	))
	moonOuterRadiusRad := math.Asin(localSolarEclipseClampUnit(
		solarEclipsePenumbralRadiusNumerator(context.params) / moonTopocentric[2],
	))
	moonInnerRadiusRad := math.Asin(localSolarEclipseClampUnit(
		geodata.EarthEquatorialRadiusKM * context.params.umbralK * localSolarMoonRadiusScale / moonTopocentric[2],
	))
	innerTotal := moonInnerRadiusRad >= sunRadiusRad
	limbApplied, outerLimbApplied := false, false
	if context.limbFrame != nil {
		outer, inner, total, outerOK, innerOK := context.limbRadii(
			observerXYZ, math.Acos(dot), sunRadiusRad, moonTopocentric[2], moonInnerRadiusRad,
		)
		// 剖面在某侧无值时该侧保持圆月面。
		if outerOK {
			moonOuterRadiusRad, outerLimbApplied = outer, true
		}
		if innerOK {
			moonInnerRadiusRad, innerTotal = inner, total
		}
		limbApplied = outerOK && innerOK
	}

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
		innerTotal:         innerTotal,
		limbApplied:        limbApplied,
		outerLimbApplied:   outerLimbApplied,
	}
}

// limbRadii 用月缘剖面求外、内接触的等效月半径（弧度）：月缘在全部位置角上的最不利余量折算成圆月面半径，
// 使圆月面接触间隙恰为月缘余量的相反数。圆月面内半径离日面半径不到一个起伏幅度时全食、环食两支都算，
// 取内切间隙较小的一支并据此报告 total；内外两侧分别报告是否取到剖面值。
func (context localSolarEclipseStateContext) limbRadii(
	observerXYZ [3]float64,
	separationRad, sunRadiusRad, moonDistanceKM, circularInnerRad float64,
) (outer, inner float64, total, outerOK, innerOK bool) {
	moonTopocentric := [3]float64{
		context.moonXYZ[0] - observerXYZ[0],
		context.moonXYZ[1] - observerXYZ[1],
		context.moonXYZ[2] - observerXYZ[2],
	}
	sunTopocentric := [3]float64{
		context.sunXYZ[0] - observerXYZ[0],
		context.sunXYZ[1] - observerXYZ[1],
		context.sunXYZ[2] - observerXYZ[2],
	}
	total = circularInnerRad >= sunRadiusRad
	view, ok := context.limbFrame.viewToward(moonTopocentric, sunTopocentric)
	if !ok {
		return 0, 0, total, false, false
	}
	contact := lunarLimbContact{
		frame:        context.limbFrame,
		view:         view,
		distanceKM:   moonDistanceKM,
		separation:   separationRad,
		targetRadius: sunRadiusRad,
	}
	outer, outerOK = contact.outerRadius()
	inner, innerOK = contact.innerRadius(total)
	low, high, ranged := contact.bounds()
	// 混合食附近月缘可让食型翻转：剖面不报告范围时无从判断，同样两支都算。
	if innerOK && (!ranged || math.Abs(circularInnerRad-sunRadiusRad) < high-low) {
		if other, otherOK := contact.innerRadius(!total); otherOK &&
			separationRad-math.Abs(other-sunRadiusRad) < separationRad-math.Abs(inner-sunRadiusRad) {
			inner, total = other, !total
		}
	}
	return outer, inner, total, outerOK, innerOK
}

func localSolarEclipseObserverXYZ(gst, lonRad, latRad, heightKM float64) [3]float64 {
	equatorialRadius := geodata.EarthEquatorialRadiusKM
	polarRadius := geodata.EarthEquatorialRadiusKM * solarEclipseEarthPolarRatio

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
