package basic

import "math"

// LunarEclipseType 表示月食类型。
type LunarEclipseType string

const (
	// LunarEclipseNone 表示该次望月没有发生月食。
	LunarEclipseNone LunarEclipseType = "none"
	// LunarEclipsePenumbral 表示半影月食。
	LunarEclipsePenumbral LunarEclipseType = "penumbral"
	// LunarEclipsePartial 表示月偏食。
	LunarEclipsePartial LunarEclipseType = "partial"
	// LunarEclipseTotal 表示月全食。
	LunarEclipseTotal LunarEclipseType = "total"
)

// LunarEclipseResult 表示一次望月附近的月食几何结果。
//
// 所有时刻字段都使用力学时儒略日（JDE, TT）。
// 输入 seedJDE 只需要落在目标望月附近，允许相差数天。
type LunarEclipseResult struct {
	Type LunarEclipseType

	// Maximum 是食甚时刻；即使最终没有月食，也会返回该次望月附近
	// “月面中心最接近地影中心”的几何极值时刻。
	Maximum float64

	// Magnitude 是本影食分。纯半影月食时可为负值；无月食时为 0。
	Magnitude float64
	// PenumbralMagnitude 是半影食分。无半影接触时为 0。
	PenumbralMagnitude float64

	// MinimumDistance 是食甚时月心到地影中心的最小角距离，单位为弧度。
	MinimumDistance float64

	// Contact times:
	// PenumbralStart / PenumbralEnd: 半影食始 / 半影食终
	// PartialStart / PartialEnd: 初亏 / 复圆
	// TotalStart / TotalEnd: 食既 / 生光
	PenumbralStart float64
	PenumbralEnd   float64
	PartialStart   float64
	PartialEnd     float64
	TotalStart     float64
	TotalEnd       float64

	HasPenumbral bool
	HasPartial   bool
	HasTotal     bool
}

type lunarShadowState struct {
	jde               float64
	x                 float64
	y                 float64
	moonRadiusRad     float64
	umbraRadiusRad    float64
	penumbraRadiusRad float64
}

type lunarEclipseShadowModel int

const (
	lunarEclipseShadowDanjon lunarEclipseShadowModel = iota
	lunarEclipseShadowChauvenet
)

const (
	lunarEarthEquatorialRadiusKM = 6378.1366
	lunarAstronomicalUnitKM      = 1.49597870691e8

	// 沿用月食常量：
	// - 0.2725076 用于月亮视半径和半影几何
	// - 959.63 / 8.794 分别为太阳视半径与太阳视差的常用角秒常量
	lunarMoonRadiusRatio      = 0.2725076
	lunarMoonRadiusScale      = lunarMoonRadiusRatio * lunarEarthEquatorialRadiusKM * 1.0000036
	lunarSolarRadiusArcsec    = 959.63
	lunarSolarParallaxArcsec  = 8.794
	lunarLongitudeAberration  = -3.4e-6
	lunarFiniteDifferenceStep = 60.0 / 86400.0

	// Chauvenet 体系：
	// - 地球有效半径取 0.99834 * 赤道半径
	// - 再统一乘 51/50 的大气放大因子
	lunarChauvenetEarthScale = 0.99834
	lunarChauvenetShadowGain = 51.0 / 50.0

	// Danjon 体系：
	// - 影半径只对月球水平视差项乘 1.01
	// - 太阳视半径与太阳视差项不再统一乘 1.02
	lunarDanjonParallaxScale = 1.01
)

var lunarArcsecPerRadian = 180.0 * 3600.0 / math.Pi

// LunarEclipse 计算给定近望时刻附近的一次月食，默认使用 Danjon 影半径模型。
//
// seedJDE 为力学时儒略日（TT），只需落在目标望月附近，允许相差数天。
// 返回值中的所有接触时刻也都是力学时儒略日。
func LunarEclipse(seedJDE float64) LunarEclipseResult {
	return LunarEclipseDanjon(seedJDE)
}

// LunarEclipseDanjon 计算给定近望时刻附近的一次月食，使用 Danjon 影半径模型。
func LunarEclipseDanjon(seedJDE float64) LunarEclipseResult {
	return lunarEclipse(seedJDE, lunarEclipseShadowDanjon)
}

// LunarEclipseChauvenet 计算给定近望时刻附近的一次月食，使用 Chauvenet 影半径模型。
func LunarEclipseChauvenet(seedJDE float64) LunarEclipseResult {
	return lunarEclipse(seedJDE, lunarEclipseShadowChauvenet)
}

func lunarEclipse(seedJDE float64, shadowModel lunarEclipseShadowModel) LunarEclipseResult {
	fullMoonJDE := CalcMoonSHByJDE(seedJDE, 1)
	maximumJDE, state, dxdt, dydt, minimumDistance := refineLunarEclipseMaximum(fullMoonJDE, shadowModel)

	result := LunarEclipseResult{
		Type:               LunarEclipseNone,
		Maximum:            maximumJDE,
		MinimumDistance:    minimumDistance,
		PenumbralMagnitude: (state.moonRadiusRad + state.penumbraRadiusRad - minimumDistance) / (2 * state.moonRadiusRad),
	}
	rawUmbralMagnitude := (state.moonRadiusRad + state.umbraRadiusRad - minimumDistance) / (2 * state.moonRadiusRad)

	if result.PenumbralMagnitude < 0 {
		result.PenumbralMagnitude = 0
	}

	if minimumDistance <= state.moonRadiusRad+state.penumbraRadiusRad {
		result.Type = LunarEclipsePenumbral
		result.HasPenumbral = true
		result.Magnitude = rawUmbralMagnitude
		result.PenumbralStart = refineLunarEclipseContact(
			state, dxdt, dydt, state.moonRadiusRad+state.penumbraRadiusRad, false, shadowModel,
		)
		result.PenumbralEnd = refineLunarEclipseContact(
			state, dxdt, dydt, state.moonRadiusRad+state.penumbraRadiusRad, true, shadowModel,
		)
	}

	if minimumDistance <= state.moonRadiusRad+state.umbraRadiusRad {
		result.Type = LunarEclipsePartial
		result.HasPartial = true
		result.Magnitude = rawUmbralMagnitude
		result.PartialStart = refineLunarEclipseContact(
			state, dxdt, dydt, state.moonRadiusRad+state.umbraRadiusRad, false, shadowModel,
		)
		result.PartialEnd = refineLunarEclipseContact(
			state, dxdt, dydt, state.moonRadiusRad+state.umbraRadiusRad, true, shadowModel,
		)
	}

	if minimumDistance <= state.umbraRadiusRad-state.moonRadiusRad {
		result.Type = LunarEclipseTotal
		result.HasTotal = true
		result.TotalStart = refineLunarEclipseContact(
			state, dxdt, dydt, state.umbraRadiusRad-state.moonRadiusRad, false, shadowModel,
		)
		result.TotalEnd = refineLunarEclipseContact(
			state, dxdt, dydt, state.umbraRadiusRad-state.moonRadiusRad, true, shadowModel,
		)
	}

	return result
}

// refineLunarEclipseMaximum 从近望初值出发，用有限差分速度做两轮几何极值修正。
//
// 这里直接在月心相对地影中心的二维平面上求 |r| 的极小值，
func refineLunarEclipseMaximum(
	seedJDE float64,
	shadowModel lunarEclipseShadowModel,
) (float64, lunarShadowState, float64, float64, float64) {
	currentJDE := seedJDE

	for i := 0; i < 2; i++ {
		state := computeLunarShadowState(currentJDE, shadowModel)
		nextState := computeLunarShadowState(currentJDE+lunarFiniteDifferenceStep, shadowModel)
		dxdt := (nextState.x - state.x) / lunarFiniteDifferenceStep
		dydt := (nextState.y - state.y) / lunarFiniteDifferenceStep

		denominator := dxdt*dxdt + dydt*dydt
		if denominator == 0 {
			finalState := computeLunarShadowState(currentJDE, shadowModel)
			return currentJDE, finalState, 0, 0, math.Hypot(finalState.x, finalState.y)
		}

		correction := -(state.x*dxdt + state.y*dydt) / denominator
		currentJDE += correction
	}

	linearState := computeLunarShadowState(currentJDE, shadowModel)
	nextLinearState := computeLunarShadowState(currentJDE+lunarFiniteDifferenceStep, shadowModel)
	dxdt := (nextLinearState.x - linearState.x) / lunarFiniteDifferenceStep
	dydt := (nextLinearState.y - linearState.y) / lunarFiniteDifferenceStep

	denominator := dxdt*dxdt + dydt*dydt
	if denominator == 0 {
		return currentJDE, linearState, dxdt, dydt, math.Hypot(linearState.x, linearState.y)
	}

	correction := -(linearState.x*dxdt + linearState.y*dydt) / denominator
	maximumJDE := currentJDE + correction

	finalState := computeLunarShadowState(maximumJDE, shadowModel)
	nextState := computeLunarShadowState(maximumJDE+lunarFiniteDifferenceStep, shadowModel)
	dxdt = (nextState.x - finalState.x) / lunarFiniteDifferenceStep
	dydt = (nextState.y - finalState.y) / lunarFiniteDifferenceStep

	return maximumJDE, finalState, dxdt, dydt, math.Hypot(finalState.x, finalState.y)
}

// refineLunarEclipseContact 先用固定速度近似求一次接触时刻，
// 再在接触点重算半径并修正一次。
func refineLunarEclipseContact(
	maximumState lunarShadowState,
	dxdt, dydt, boundaryRadius float64,
	afterMaximum bool,
	shadowModel lunarEclipseShadowModel,
) float64 {
	firstGuess, ok := solveLineCircleContact(maximumState, dxdt, dydt, boundaryRadius, afterMaximum)
	if !ok {
		return 0
	}

	contactState := computeLunarShadowState(firstGuess, shadowModel)
	refinedRadius := boundaryRadius

	switch {
	case math.Abs(boundaryRadius-(maximumState.moonRadiusRad+maximumState.umbraRadiusRad)) < 1e-18:
		refinedRadius = contactState.moonRadiusRad + contactState.umbraRadiusRad
	case math.Abs(boundaryRadius-(maximumState.moonRadiusRad+maximumState.penumbraRadiusRad)) < 1e-18:
		refinedRadius = contactState.moonRadiusRad + contactState.penumbraRadiusRad
	case math.Abs(boundaryRadius-(maximumState.umbraRadiusRad-maximumState.moonRadiusRad)) < 1e-18:
		refinedRadius = contactState.umbraRadiusRad - contactState.moonRadiusRad
	}

	refinedGuess, ok := solveLineCircleContact(contactState, dxdt, dydt, refinedRadius, afterMaximum)
	if !ok {
		return firstGuess
	}
	return refinedGuess
}

// solveLineCircleContact 求月心轨迹与某个影界圆的交点时刻。
func solveLineCircleContact(
	state lunarShadowState,
	dxdt, dydt, radius float64,
	afterMaximum bool,
) (float64, bool) {
	a := dxdt*dxdt + dydt*dydt
	if a == 0 {
		return 0, false
	}

	b := 2 * (state.x*dxdt + state.y*dydt)
	c := state.x*state.x + state.y*state.y - radius*radius
	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return 0, false
	}

	root := math.Sqrt(discriminant)
	delta := (-b - root) / (2 * a)
	if afterMaximum {
		delta = (-b + root) / (2 * a)
	}

	return state.jde + delta, true
}

// computeLunarShadowState 计算某一力学时刻下，月心相对地影中心的二维几何状态。
//
// 所有内部角量统一使用弧度。影半径模型允许在 Danjon 与 Chauvenet 之间切换，
// 其余月心轨迹与几何求交框架保持一致。
func computeLunarShadowState(jde float64, shadowModel lunarEclipseShadowModel) lunarShadowState {
	julianCentury := (jde - 2451545.0) / 36525.0

	sunLongitude := HSunTrueLo(jde)*rad + sunLongitudeAberrationRad(julianCentury)
	sunLatitude := HSunTrueBo(jde) * rad
	moonLongitude := HMoonTrueLo(jde)*rad + lunarLongitudeAberration
	moonLatitude := HMoonTrueBo(jde)*rad + moonLatitudeAberrationRad(julianCentury)

	moonDistanceKM := HMoonAway(jde)
	sunDistanceAU := EarthAway(jde)

	moonRadiusArcsec := lunarMoonRadiusScale * lunarArcsecPerRadian / moonDistanceKM
	earthParallaxArcsec := lunarEarthEquatorialRadiusKM / moonDistanceKM * lunarArcsecPerRadian
	solarRadiusArcsec := lunarSolarRadiusArcsec / sunDistanceAU
	solarParallaxArcsec := lunarSolarParallaxArcsec / sunDistanceAU
	umbraRadiusArcsec, penumbraRadiusArcsec := lunarEclipseShadowRadiiArcsec(
		earthParallaxArcsec,
		solarRadiusArcsec,
		solarParallaxArcsec,
		shadowModel,
	)

	return lunarShadowState{
		jde:               jde,
		x:                 normalizeRadians(moonLongitude+math.Pi-sunLongitude) * math.Cos((moonLatitude-sunLatitude)/2),
		y:                 moonLatitude + sunLatitude,
		moonRadiusRad:     moonRadiusArcsec / lunarArcsecPerRadian,
		umbraRadiusRad:    umbraRadiusArcsec / lunarArcsecPerRadian,
		penumbraRadiusRad: penumbraRadiusArcsec / lunarArcsecPerRadian,
	}
}

func lunarEclipseShadowRadiiArcsec(
	earthParallaxArcsec, solarRadiusArcsec, solarParallaxArcsec float64,
	shadowModel lunarEclipseShadowModel,
) (float64, float64) {
	switch shadowModel {
	case lunarEclipseShadowDanjon:
		earthTerm := lunarDanjonParallaxScale * earthParallaxArcsec
		return earthTerm - solarRadiusArcsec + solarParallaxArcsec,
			earthTerm + solarRadiusArcsec + solarParallaxArcsec
	default:
		earthTerm := lunarChauvenetEarthScale * earthParallaxArcsec
		return (earthTerm - solarRadiusArcsec + solarParallaxArcsec) * lunarChauvenetShadowGain,
			(earthTerm + solarRadiusArcsec + solarParallaxArcsec) * lunarChauvenetShadowGain
	}
}

func normalizeRadians(angle float64) float64 {
	angle = math.Mod(angle, 2*math.Pi)
	if angle > math.Pi {
		angle -= 2 * math.Pi
	}
	if angle <= -math.Pi {
		angle += 2 * math.Pi
	}
	return angle
}

func sunLongitudeAberrationRad(julianCentury float64) float64 {
	meanAnomaly := -0.043126 + 628.301955*julianCentury - 0.000002732*julianCentury*julianCentury
	eccentricity := 0.016708634 - 0.000042037*julianCentury - 0.0000001267*julianCentury*julianCentury
	return -20.49552 * (1 + eccentricity*math.Cos(meanAnomaly)) / lunarArcsecPerRadian
}

func moonLatitudeAberrationRad(julianCentury float64) float64 {
	argument := 0.057 + 8433.4662*julianCentury + 0.000064*julianCentury*julianCentury
	return 0.063 * math.Sin(argument) / lunarArcsecPerRadian
}
