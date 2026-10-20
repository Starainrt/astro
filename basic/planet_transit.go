package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

const (
	planetTransitSearchLimit           = 2400
	planetTransitSearchEpsilonDays     = 1.0 / 86400.0
	planetTransitGreatestWindowDays    = 1.2
	planetTransitGreatestToleranceDays = 0.25 / 86400.0
	planetTransitContactStepDays       = 0.02
	planetTransitContactSpanDays       = 1.0
	planetTransitContactToleranceDays  = 0.25 / 86400.0
	planetTransitCoarseN               = 16
)

// PlanetTransitResult 全球行星凌日与地心接触 / global planetary transit and geocentric contacts.
//
// Valid 为 false 时表示没有找到有效凌日。所有时刻字段均为 UT 儒略日。
// MinimumSeparationArcsec、SunSemidiameterArcsec、PlanetSemidiameterArcsec 的单位均为角秒。
type PlanetTransitResult struct {
	Valid bool

	// PlanetIndex 为行星序号，1 表示水星，2 表示金星。
	PlanetIndex int

	// ExternalIngress / ExternalEgress 为一触 / 四触。
	ExternalIngress float64
	ExternalEgress  float64
	// InternalIngress / InternalEgress 为二触 / 三触。掠凌没有内切接触时为 0。
	InternalIngress float64
	InternalEgress  float64
	// Greatest 为凌甚，即行星中心最接近太阳中心的时刻。
	Greatest float64

	MinimumSeparationArcsec  float64
	SunSemidiameterArcsec    float64
	PlanetSemidiameterArcsec float64

	HasExternal bool
	HasInternal bool

	// WindowStart、WindowEnd 是站心求解窗口，不是接触时刻。
	WindowStart float64
	WindowEnd   float64
	// Gamma 是凌甚时影轴到地心的有符号距离，以地球赤道半径为单位，北侧为正。
	Gamma float64
	// PenumbralRadius 是地心基本面上的半影半径，以地球赤道半径为单位。
	PenumbralRadius float64
	Series          PlanetTransitSeries
}

type planetTransitConfig struct {
	planetIndex int

	synodicPeriodDays float64
	anchorInferiorTT  float64

	latitudePrefilter  float64
	conjunctionStepDay float64

	apparentLoN    func(float64, int) float64
	apparentBoN    func(float64, int) float64
	apparentRaDecN func(float64, int) (float64, float64)
	semidiameterN  func(float64, int) float64
	earthDistanceN func(float64, int) float64
}

type planetTransitState struct {
	jdTT                  float64
	separationArcsec      float64
	separationSquared     float64
	sunSemidiameter       float64
	planetSemidiameter    float64
	externalContactMetric float64
	internalContactMetric float64
}

func mercuryTransitConfig(instances ...*Instance) planetTransitConfig {
	inst := instanceValue(instances)

	return planetTransitConfig{
		planetIndex:        1,
		synodicPeriodDays:  MERCURY_S_PERIOD,
		anchorInferiorTT:   UTC2TT(JDCalc(2019, 11, 11+(15+21.0/60+40.0/3600)/24)),
		latitudePrefilter:  1.0,
		conjunctionStepDay: 0.00001,
		apparentLoN:        func(jd float64, n int) float64 { return MercuryApparentLoNWithInstance(jd, n, inst) },
		apparentBoN:        func(jd float64, n int) float64 { return MercuryApparentBoNWithInstance(jd, n, inst) },
		apparentRaDecN:     func(jde float64, n int) (float64, float64) { return MercuryApparentRaDecNWithInstance(jde, n, inst) },
		semidiameterN:      func(jde float64, n int) float64 { return MercurySemidiameterNWithInstance(jde, n, inst) },
		earthDistanceN:     func(jd float64, n int) float64 { return EarthMercuryAwayNWithInstance(jd, n, inst) },
	}
}

func venusTransitConfig(instances ...*Instance) planetTransitConfig {
	inst := instanceValue(instances)

	return planetTransitConfig{
		planetIndex:        2,
		synodicPeriodDays:  VENUS_S_PERIOD,
		anchorInferiorTT:   UTC2TT(JDCalc(2012, 6, 6+(1+29.0/60)/24)),
		latitudePrefilter:  0.8,
		conjunctionStepDay: 0.00001,
		apparentLoN:        func(jd float64, n int) float64 { return VenusApparentLoNWithInstance(jd, n, inst) },
		apparentBoN:        func(jd float64, n int) float64 { return VenusApparentBoNWithInstance(jd, n, inst) },
		apparentRaDecN:     func(jd float64, n int) (float64, float64) { return VenusApparentRaDecNWithInstance(jd, n, inst) },
		semidiameterN:      func(jde float64, n int) float64 { return VenusSemidiameterNWithInstance(jde, n, inst) },
		earthDistanceN:     func(jd float64, n int) float64 { return EarthVenusAwayNWithInstance(jd, n, inst) },
	}
}

// NextMercuryTransitWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMercuryTransitWithInstance(jde float64, inst *Instance) PlanetTransitResult {
	result, _ := searchPlanetTransit(jde, mercuryTransitConfig(inst), 1, false, inst)
	return result
}

// LastMercuryTransitWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMercuryTransitWithInstance(jde float64, inst *Instance) PlanetTransitResult {
	result, _ := searchPlanetTransit(jde, mercuryTransitConfig(inst), -1, true, inst)
	return result
}

// ClosestMercuryTransitWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func ClosestMercuryTransitWithInstance(jde float64, inst *Instance) PlanetTransitResult {
	return closestPlanetTransit(jde, mercuryTransitConfig(inst), inst)
}

// NextVenusTransitWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextVenusTransitWithInstance(jde float64, inst *Instance) PlanetTransitResult {
	result, _ := searchPlanetTransit(jde, venusTransitConfig(inst), 1, false, inst)
	return result
}

// LastVenusTransitWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastVenusTransitWithInstance(jde float64, inst *Instance) PlanetTransitResult {
	result, _ := searchPlanetTransit(jde, venusTransitConfig(inst), -1, true, inst)
	return result
}

// ClosestVenusTransitWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func ClosestVenusTransitWithInstance(jde float64, inst *Instance) PlanetTransitResult {
	return closestPlanetTransit(jde, venusTransitConfig(inst), inst)
}

func closestPlanetTransit(jde float64, cfg planetTransitConfig, instances ...*Instance) PlanetTransitResult {
	inst := instanceValue(instances)

	last, hasLast := searchPlanetTransit(jde, cfg, -1, true, inst)
	next, hasNext := searchPlanetTransit(jde, cfg, 1, false, inst)
	switch {
	case hasLast && !hasNext:
		return last
	case !hasLast && hasNext:
		return next
	case !hasLast && !hasNext:
		return PlanetTransitResult{}
	}
	if math.Abs(last.Greatest-jde) <= math.Abs(next.Greatest-jde) {
		return last
	}
	return next
}

func searchPlanetTransit(jd float64, cfg planetTransitConfig, direction int, includeCurrent bool, instances ...*Instance) (PlanetTransitResult, bool) {
	inst := instanceValue(instances)

	if !isFiniteFloat(jd) || direction == 0 {
		return PlanetTransitResult{}, false
	}

	targetTT := UTC2TT(jd)
	if !isFiniteFloat(targetTT) {
		return PlanetTransitResult{}, false
	}
	cycle := math.Floor((targetTT-cfg.anchorInferiorTT)/cfg.synodicPeriodDays) - 1
	if direction < 0 {
		cycle += 3
	}
	for i := 0; i < planetTransitSearchLimit; i++ {
		seedTT := cfg.anchorInferiorTT + cycle*cfg.synodicPeriodDays
		if planetTransitPredictedSeason(seedTT, cfg, inst) {
			// 单个会合周算不出粗时刻（例如候选落在有限覆盖星历之外）时跳过该周，后面的周期仍可能落在覆盖内。
			coarseTT := planetTransitCoarseConjunction(seedTT, cfg, inst)
			if !isFiniteFloat(coarseTT) {
				cycle += float64(direction)
				continue
			}
			if isPotentialPlanetTransit(coarseTT, cfg, inst) {
				conjunctionTT := refinePlanetTransitInferiorConjunctionTT(coarseTT, cfg, inst)
				resultTT, ok := planetTransitAtInferiorConjunctionTT(conjunctionTT, cfg, inst)
				if ok && planetTransitMatchesDirection(resultTT.Greatest, targetTT, direction, includeCurrent) {
					return planetTransitResultTTToUT(resultTT), true
				}
			}
		}

		cycle += float64(direction)
	}

	return PlanetTransitResult{}, false
}

func refinePlanetTransitInferiorConjunctionTT(seedTT float64, cfg planetTransitConfig, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	current := seedTT
	for i := 0; i < 4; i++ {
		prev := current
		value := planetTransitLongitudeDeltaN(prev, cfg, planetTransitCoarseN, inst)
		slope := (planetTransitLongitudeDeltaN(prev+cfg.conjunctionStepDay, cfg, planetTransitCoarseN, inst) -
			planetTransitLongitudeDeltaN(prev-cfg.conjunctionStepDay, cfg, planetTransitCoarseN, inst)) / (2 * cfg.conjunctionStepDay)
		if slope == 0 || !isFiniteFloat(slope) {
			break
		}
		current = prev - value/slope
		if math.Abs(current-prev) <= 30.0/86400.0 {
			break
		}
	}
	for i := 0; i < 8; i++ {
		prev := current
		value := planetTransitLongitudeDeltaN(prev, cfg, -1, inst)
		slope := (planetTransitLongitudeDeltaN(prev+cfg.conjunctionStepDay, cfg, -1, inst) -
			planetTransitLongitudeDeltaN(prev-cfg.conjunctionStepDay, cfg, -1, inst)) / (2 * cfg.conjunctionStepDay)
		if slope == 0 || !isFiniteFloat(slope) {
			break
		}
		current = prev - value/slope
		if math.Abs(current-prev) <= cfg.conjunctionStepDay {
			break
		}
	}
	return current
}

func planetTransitLongitudeDeltaN(jdTT float64, cfg planetTransitConfig, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	return planetTransitAngleDelta(cfg.apparentLoN(jdTT, n) - HSunApparentLoNWithInstance(jdTT, n, inst))
}

func isPotentialPlanetTransit(conjunctionTT float64, cfg planetTransitConfig, instances ...*Instance) bool {
	inst := instanceValue(instances)

	if cfg.earthDistanceN(conjunctionTT, planetTransitCoarseN) > EarthAwayNWithInstance(conjunctionTT, planetTransitCoarseN, inst) {
		return false
	}
	return math.Abs(cfg.apparentBoN(conjunctionTT, planetTransitCoarseN)) <= cfg.latitudePrefilter
}

func planetTransitAtInferiorConjunctionTT(conjunctionTT float64, cfg planetTransitConfig, instances ...*Instance) (PlanetTransitResult, bool) {
	inst := instanceValue(instances)

	greatestTT := greatestPlanetTransitTT(conjunctionTT, cfg, inst)
	greatestState := planetTransitStateAt(greatestTT, cfg, -1, inst)
	if !isFiniteFloat(greatestState.externalContactMetric) {
		return PlanetTransitResult{}, false
	}
	gamma, penumbra, surfaceSeed, possible := planetTransitShadowGeometry(greatestTT, cfg, inst)
	if !possible {
		return PlanetTransitResult{}, false
	}

	result := PlanetTransitResult{
		Valid:                    true,
		PlanetIndex:              cfg.planetIndex,
		Greatest:                 greatestTT,
		MinimumSeparationArcsec:  greatestState.separationArcsec,
		SunSemidiameterArcsec:    greatestState.sunSemidiameter,
		PlanetSemidiameterArcsec: greatestState.planetSemidiameter,
		HasExternal:              greatestState.externalContactMetric <= 0,
		HasInternal:              greatestState.internalContactMetric <= 0,
		WindowStart:              greatestTT - 0.25,
		WindowEnd:                greatestTT + 0.25,
		Gamma:                    gamma,
		PenumbralRadius:          penumbra,
		Series:                   planetTransitSeriesAt(greatestTT, cfg),
	}
	if !result.HasExternal {
		if !planetTransitSurfaceOverlap(planetTransitResultTTToUT(result), surfaceSeed, inst) {
			return PlanetTransitResult{}, false
		}
		return result, true
	}

	externalIngress, ok := refinePlanetTransitContactTT(greatestTT, cfg, -1, false, inst)
	if !ok {
		return PlanetTransitResult{}, false
	}
	externalEgress, ok := refinePlanetTransitContactTT(greatestTT, cfg, 1, false, inst)
	if !ok || externalEgress <= externalIngress {
		return PlanetTransitResult{}, false
	}
	result.ExternalIngress = externalIngress
	result.ExternalEgress = externalEgress
	result.WindowStart = externalIngress - planetTransitSnapshotMarginDays
	result.WindowEnd = externalEgress + planetTransitSnapshotMarginDays

	if result.HasInternal {
		internalIngress, ok := refinePlanetTransitContactTT(greatestTT, cfg, -1, true, inst)
		if ok {
			result.InternalIngress = internalIngress
		}
		internalEgress, ok := refinePlanetTransitContactTT(greatestTT, cfg, 1, true, inst)
		if ok && internalEgress > internalIngress {
			result.InternalEgress = internalEgress
		}
		result.HasInternal = result.InternalIngress != 0 && result.InternalEgress != 0
	}

	return result, true
}

func greatestPlanetTransitTT(seedTT float64, cfg planetTransitConfig, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	left := seedTT - planetTransitGreatestWindowDays
	right := seedTT + planetTransitGreatestWindowDays
	goldenRatio := (math.Sqrt(5) - 1) / 2

	x1 := right - goldenRatio*(right-left)
	x2 := left + goldenRatio*(right-left)
	f1 := planetTransitStateAt(x1, cfg, planetTransitCoarseN, inst).separationSquared
	f2 := planetTransitStateAt(x2, cfg, planetTransitCoarseN, inst).separationSquared

	for i := 0; i < 80 && right-left > planetTransitGreatestToleranceDays; i++ {
		if f1 <= f2 {
			right = x2
			x2 = x1
			f2 = f1
			x1 = right - goldenRatio*(right-left)
			f1 = planetTransitStateAt(x1, cfg, planetTransitCoarseN, inst).separationSquared
			continue
		}
		left = x1
		x1 = x2
		f1 = f2
		x2 = left + goldenRatio*(right-left)
		f2 = planetTransitStateAt(x2, cfg, planetTransitCoarseN, inst).separationSquared
	}

	center := (left + right) / 2
	left = center - 2.0/24.0
	right = center + 2.0/24.0
	x1 = right - goldenRatio*(right-left)
	x2 = left + goldenRatio*(right-left)
	f1 = planetTransitStateAt(x1, cfg, -1, inst).separationSquared
	f2 = planetTransitStateAt(x2, cfg, -1, inst).separationSquared
	for i := 0; i < 80 && right-left > planetTransitGreatestToleranceDays; i++ {
		if f1 <= f2 {
			right = x2
			x2 = x1
			f2 = f1
			x1 = right - goldenRatio*(right-left)
			f1 = planetTransitStateAt(x1, cfg, -1, inst).separationSquared
			continue
		}
		left = x1
		x1 = x2
		f1 = f2
		x2 = left + goldenRatio*(right-left)
		f2 = planetTransitStateAt(x2, cfg, -1, inst).separationSquared
	}
	return (left + right) / 2
}

func planetTransitStateAt(jdTT float64, cfg planetTransitConfig, n int, instances ...*Instance) planetTransitState {
	inst := instanceValue(instances)

	planetRA, planetDec := cfg.apparentRaDecN(jdTT, n)
	sunRA, sunDec := HSunApparentRaDecNWithInstance(jdTT, n, inst)
	separationArcsec := StarAngularSeparation(planetRA, planetDec, sunRA, sunDec) * 3600
	sunSemidiameter := SunSemidiameterNWithInstance(jdTT, n, inst)
	planetSemidiameter := cfg.semidiameterN(jdTT, n)
	return planetTransitState{
		jdTT:                  jdTT,
		separationArcsec:      separationArcsec,
		separationSquared:     separationArcsec * separationArcsec,
		sunSemidiameter:       sunSemidiameter,
		planetSemidiameter:    planetSemidiameter,
		externalContactMetric: separationArcsec - (sunSemidiameter + planetSemidiameter),
		internalContactMetric: separationArcsec - (sunSemidiameter - planetSemidiameter),
	}
}

func refinePlanetTransitContactTT(greatestTT float64, cfg planetTransitConfig, direction int, internal bool, instances ...*Instance) (float64, bool) {
	inst := instanceValue(instances)

	if direction != -1 && direction != 1 {
		return 0, false
	}
	metric := func(jdTT float64) float64 {
		state := planetTransitStateAt(jdTT, cfg, -1, inst)
		if internal {
			return state.internalContactMetric
		}
		return state.externalContactMetric
	}

	nearJD := greatestTT
	nearValue := metric(nearJD)
	if !isFiniteFloat(nearValue) || nearValue > 0 {
		return 0, false
	}
	maxSteps := int(math.Ceil(planetTransitContactSpanDays / planetTransitContactStepDays))
	for i := 1; i <= maxSteps; i++ {
		farJD := greatestTT + float64(direction)*planetTransitContactStepDays*float64(i)
		farValue := metric(farJD)
		if !isFiniteFloat(farValue) {
			continue
		}
		if farValue >= 0 {
			return bisectPlanetTransitContactTT(nearJD, nearValue, farJD, farValue, metric)
		}
		nearJD = farJD
		nearValue = farValue
	}
	return 0, false
}

func bisectPlanetTransitContactTT(leftJD, leftValue, rightJD, rightValue float64, metric func(float64) float64) (float64, bool) {
	if leftJD > rightJD {
		leftJD, rightJD = rightJD, leftJD
		leftValue, rightValue = rightValue, leftValue
	}
	if leftValue == 0 {
		return leftJD, true
	}
	if rightValue == 0 {
		return rightJD, true
	}
	if leftValue*rightValue > 0 {
		return 0, false
	}

	for i := 0; i < 80 && rightJD-leftJD > planetTransitContactToleranceDays; i++ {
		midJD := (leftJD + rightJD) / 2
		midValue := metric(midJD)
		if !isFiniteFloat(midValue) {
			return 0, false
		}
		if midValue == 0 {
			return midJD, true
		}
		if leftValue*midValue <= 0 {
			rightJD = midJD
			rightValue = midValue
			continue
		}
		leftJD = midJD
		leftValue = midValue
	}
	return (leftJD + rightJD) / 2, true
}

func planetTransitResultTTToUT(result PlanetTransitResult) PlanetTransitResult {
	result.Greatest = TT2UTC(result.Greatest)
	if result.ExternalIngress != 0 {
		result.ExternalIngress = TT2UTC(result.ExternalIngress)
	}
	if result.ExternalEgress != 0 {
		result.ExternalEgress = TT2UTC(result.ExternalEgress)
	}
	if result.WindowStart != 0 {
		result.WindowStart = TT2UTC(result.WindowStart)
	}
	if result.WindowEnd != 0 {
		result.WindowEnd = TT2UTC(result.WindowEnd)
	}
	if result.InternalIngress != 0 {
		result.InternalIngress = TT2UTC(result.InternalIngress)
	}
	if result.InternalEgress != 0 {
		result.InternalEgress = TT2UTC(result.InternalEgress)
	}
	return result
}

func planetTransitMatchesDirection(eventJDE, targetJDE float64, direction int, includeCurrent bool) bool {
	delta := eventJDE - targetJDE
	if math.Abs(delta) <= planetTransitSearchEpsilonDays {
		return includeCurrent
	}
	if direction > 0 {
		return delta > 0
	}
	return delta < 0
}

func planetTransitAngleDelta(diff float64) float64 {
	diff = Limit360(diff)
	if diff > 180 {
		diff -= 360
	}
	if diff < -180 {
		diff += 360
	}
	return diff
}
