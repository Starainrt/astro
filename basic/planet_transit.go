package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

const (
	planetTransitMeanSolarMotionDegPerDay = 360.0 / 365.2422
	planetTransitTropicalYearDays         = 365.2422
	planetTransitSeasonProbeStepDays      = 20.0
	planetTransitSearchLimit              = 2400
	planetTransitSearchEpsilonDays        = 1.0 / 86400.0
	planetTransitGreatestWindowDays       = 1.2
	planetTransitGreatestToleranceDays    = 0.25 / 86400.0
	planetTransitContactStepDays          = 0.02
	planetTransitContactSpanDays          = 1.0
	planetTransitContactToleranceDays     = 0.25 / 86400.0
	planetTransitCoarseN                  = 16
)

// PlanetTransitResult 表示一次地心行星凌日结果。
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
}

type planetTransitConfig struct {
	planetIndex int

	synodicPeriodDays float64
	anchorInferiorTT  float64

	seasonWindowDays   float64
	latitudePrefilter  float64
	conjunctionStepDay float64

	apparentLoN    func(float64, int) float64
	apparentBoN    func(float64, int) float64
	apparentRaDecN func(float64, int) (float64, float64)
	semidiameterN  func(float64, int) float64
	earthDistanceN func(float64, int) float64
	nodeN          func(float64, int) float64
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

func mercuryTransitConfig() planetTransitConfig {
	return planetTransitConfig{
		planetIndex:        1,
		synodicPeriodDays:  MERCURY_S_PERIOD,
		anchorInferiorTT:   TD2UT(JDECalc(2019, 11, 11+(15+21.0/60+40.0/3600)/24), true),
		seasonWindowDays:   12,
		latitudePrefilter:  1.0,
		conjunctionStepDay: 0.00001,
		apparentLoN:        MercuryApparentLoN,
		apparentBoN:        MercuryApparentBoN,
		apparentRaDecN:     MercuryApparentRaDecN,
		semidiameterN:      MercurySemidiameterN,
		earthDistanceN:     EarthMercuryAwayN,
		nodeN:              MercuryAscendingNodeN,
	}
}

func venusTransitConfig() planetTransitConfig {
	return planetTransitConfig{
		planetIndex:        2,
		synodicPeriodDays:  VENUS_S_PERIOD,
		anchorInferiorTT:   TD2UT(JDECalc(2012, 6, 6+(1+29.0/60)/24), true),
		seasonWindowDays:   8,
		latitudePrefilter:  0.8,
		conjunctionStepDay: 0.00001,
		apparentLoN:        VenusApparentLoN,
		apparentBoN:        VenusApparentBoN,
		apparentRaDecN:     VenusApparentRaDecN,
		semidiameterN:      VenusSemidiameterN,
		earthDistanceN:     EarthVenusAwayN,
		nodeN:              VenusAscendingNodeN,
	}
}

// NextMercuryTransit 返回给定时刻之后的下一次地心水星凌日。
func NextMercuryTransit(jde float64) PlanetTransitResult {
	result, _ := searchPlanetTransit(jde, mercuryTransitConfig(), 1, false)
	return result
}

// LastMercuryTransit 返回给定时刻之前的上一次地心水星凌日。
func LastMercuryTransit(jde float64) PlanetTransitResult {
	result, _ := searchPlanetTransit(jde, mercuryTransitConfig(), -1, true)
	return result
}

// ClosestMercuryTransit 返回距给定时刻最近的一次地心水星凌日。
func ClosestMercuryTransit(jde float64) PlanetTransitResult {
	return closestPlanetTransit(jde, mercuryTransitConfig())
}

// NextVenusTransit 返回给定时刻之后的下一次地心金星凌日。
func NextVenusTransit(jde float64) PlanetTransitResult {
	result, _ := searchPlanetTransit(jde, venusTransitConfig(), 1, false)
	return result
}

// LastVenusTransit 返回给定时刻之前的上一次地心金星凌日。
func LastVenusTransit(jde float64) PlanetTransitResult {
	result, _ := searchPlanetTransit(jde, venusTransitConfig(), -1, true)
	return result
}

// ClosestVenusTransit 返回距给定时刻最近的一次地心金星凌日。
func ClosestVenusTransit(jde float64) PlanetTransitResult {
	return closestPlanetTransit(jde, venusTransitConfig())
}

func closestPlanetTransit(jde float64, cfg planetTransitConfig) PlanetTransitResult {
	last, hasLast := searchPlanetTransit(jde, cfg, -1, true)
	next, hasNext := searchPlanetTransit(jde, cfg, 1, false)
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

func searchPlanetTransit(jde float64, cfg planetTransitConfig, direction int, includeCurrent bool) (PlanetTransitResult, bool) {
	if !isFiniteFloat(jde) || direction == 0 {
		return PlanetTransitResult{}, false
	}

	targetTT := TD2UT(jde, true)
	probeTT := targetTT
	for i := 0; i < planetTransitSearchLimit; i++ {
		seasonTT, ok := nextPlanetTransitSeasonTT(probeTT, cfg, direction)
		if !ok {
			return PlanetTransitResult{}, false
		}

		seedTT := nearestPlanetTransitInferiorSeedTT(seasonTT, cfg)
		if math.Abs(seedTT-seasonTT) <= cfg.seasonWindowDays {
			conjunctionTT := refinePlanetTransitInferiorConjunctionTT(seedTT, cfg)
			if math.Abs(conjunctionTT-seasonTT) <= cfg.seasonWindowDays+1 && isPotentialPlanetTransit(conjunctionTT, cfg) {
				resultTT, ok := planetTransitAtInferiorConjunctionTT(conjunctionTT, cfg)
				if ok && planetTransitMatchesDirection(resultTT.Greatest, targetTT, direction, includeCurrent) {
					return planetTransitResultTTToUT(resultTT), true
				}
			}
		}

		probeTT = seasonTT + float64(direction)*planetTransitSeasonProbeStepDays
	}

	return PlanetTransitResult{}, false
}

func nextPlanetTransitSeasonTT(jdTT float64, cfg planetTransitConfig, direction int) (float64, bool) {
	best := math.NaN()
	for nodeOffset := 0; nodeOffset <= 1; nodeOffset++ {
		candidate := estimatePlanetTransitSeasonTT(jdTT, cfg, nodeOffset, direction)
		candidate = refinePlanetTransitSeasonTT(candidate, cfg, nodeOffset)
		for !planetTransitMatchesDirection(candidate, jdTT, direction, false) {
			candidate += float64(direction) * planetTransitTropicalYearDays
			candidate = refinePlanetTransitSeasonTT(candidate, cfg, nodeOffset)
		}
		if !isFiniteFloat(best) || math.Abs(candidate-jdTT) < math.Abs(best-jdTT) {
			best = candidate
		}
	}
	if !isFiniteFloat(best) {
		return 0, false
	}
	return best, true
}

func estimatePlanetTransitSeasonTT(jdTT float64, cfg planetTransitConfig, nodeOffset int, direction int) float64 {
	sunLongitude := HSunApparentLoN(jdTT, planetTransitCoarseN)
	nodeLongitude := planetTransitNodeLongitude(jdTT, cfg, nodeOffset, planetTransitCoarseN)
	if direction > 0 {
		delta := Limit360(nodeLongitude - sunLongitude)
		if delta <= planetTransitSearchEpsilonDays {
			delta += 360
		}
		return jdTT + delta/planetTransitMeanSolarMotionDegPerDay
	}
	delta := Limit360(sunLongitude - nodeLongitude)
	if delta <= planetTransitSearchEpsilonDays {
		delta += 360
	}
	return jdTT - delta/planetTransitMeanSolarMotionDegPerDay
}

func refinePlanetTransitSeasonTT(seedTT float64, cfg planetTransitConfig, nodeOffset int) float64 {
	current := seedTT
	for i := 0; i < 8; i++ {
		prev := current
		value := planetTransitSunNodeLongitudeDelta(prev, cfg, nodeOffset)
		slope := (planetTransitSunNodeLongitudeDelta(prev+0.5, cfg, nodeOffset) -
			planetTransitSunNodeLongitudeDelta(prev-0.5, cfg, nodeOffset)) / 1.0
		if slope == 0 || !isFiniteFloat(slope) {
			break
		}
		current = prev - value/slope
		if math.Abs(current-prev) <= 0.00001 {
			break
		}
	}
	return current
}

func planetTransitSunNodeLongitudeDelta(jdTT float64, cfg planetTransitConfig, nodeOffset int) float64 {
	return planetTransitAngleDelta(HSunApparentLoN(jdTT, planetTransitCoarseN) -
		planetTransitNodeLongitude(jdTT, cfg, nodeOffset, planetTransitCoarseN))
}

func planetTransitNodeLongitude(jdTT float64, cfg planetTransitConfig, nodeOffset int, n int) float64 {
	return Limit360(cfg.nodeN(jdTT, n) + float64(nodeOffset)*180)
}

func nearestPlanetTransitInferiorSeedTT(seasonTT float64, cfg planetTransitConfig) float64 {
	k := math.Round((seasonTT - cfg.anchorInferiorTT) / cfg.synodicPeriodDays)
	return cfg.anchorInferiorTT + k*cfg.synodicPeriodDays
}

func refinePlanetTransitInferiorConjunctionTT(seedTT float64, cfg planetTransitConfig) float64 {
	current := seedTT
	for i := 0; i < 4; i++ {
		prev := current
		value := planetTransitLongitudeDeltaN(prev, cfg, planetTransitCoarseN)
		slope := (planetTransitLongitudeDeltaN(prev+cfg.conjunctionStepDay, cfg, planetTransitCoarseN) -
			planetTransitLongitudeDeltaN(prev-cfg.conjunctionStepDay, cfg, planetTransitCoarseN)) / (2 * cfg.conjunctionStepDay)
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
		value := planetTransitLongitudeDeltaN(prev, cfg, -1)
		slope := (planetTransitLongitudeDeltaN(prev+cfg.conjunctionStepDay, cfg, -1) -
			planetTransitLongitudeDeltaN(prev-cfg.conjunctionStepDay, cfg, -1)) / (2 * cfg.conjunctionStepDay)
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

func planetTransitLongitudeDeltaN(jdTT float64, cfg planetTransitConfig, n int) float64 {
	return planetTransitAngleDelta(cfg.apparentLoN(jdTT, n) - HSunApparentLoN(jdTT, n))
}

func isPotentialPlanetTransit(conjunctionTT float64, cfg planetTransitConfig) bool {
	if cfg.earthDistanceN(conjunctionTT, planetTransitCoarseN) > EarthAwayN(conjunctionTT, planetTransitCoarseN) {
		return false
	}
	return math.Abs(cfg.apparentBoN(conjunctionTT, planetTransitCoarseN)) <= cfg.latitudePrefilter
}

func planetTransitAtInferiorConjunctionTT(conjunctionTT float64, cfg planetTransitConfig) (PlanetTransitResult, bool) {
	greatestTT := greatestPlanetTransitTT(conjunctionTT, cfg)
	greatestState := planetTransitStateAt(greatestTT, cfg, -1)
	if !isFiniteFloat(greatestState.externalContactMetric) || greatestState.externalContactMetric > 0 {
		return PlanetTransitResult{}, false
	}

	result := PlanetTransitResult{
		Valid:                    true,
		PlanetIndex:              cfg.planetIndex,
		Greatest:                 greatestTT,
		MinimumSeparationArcsec:  greatestState.separationArcsec,
		SunSemidiameterArcsec:    greatestState.sunSemidiameter,
		PlanetSemidiameterArcsec: greatestState.planetSemidiameter,
		HasExternal:              true,
		HasInternal:              greatestState.internalContactMetric <= 0,
	}

	externalIngress, ok := refinePlanetTransitContactTT(greatestTT, cfg, -1, false)
	if !ok {
		return PlanetTransitResult{}, false
	}
	externalEgress, ok := refinePlanetTransitContactTT(greatestTT, cfg, 1, false)
	if !ok || externalEgress <= externalIngress {
		return PlanetTransitResult{}, false
	}
	result.ExternalIngress = externalIngress
	result.ExternalEgress = externalEgress

	if result.HasInternal {
		internalIngress, ok := refinePlanetTransitContactTT(greatestTT, cfg, -1, true)
		if ok {
			result.InternalIngress = internalIngress
		}
		internalEgress, ok := refinePlanetTransitContactTT(greatestTT, cfg, 1, true)
		if ok && internalEgress > internalIngress {
			result.InternalEgress = internalEgress
		}
		result.HasInternal = result.InternalIngress != 0 && result.InternalEgress != 0
	}

	return result, true
}

func greatestPlanetTransitTT(seedTT float64, cfg planetTransitConfig) float64 {
	left := seedTT - planetTransitGreatestWindowDays
	right := seedTT + planetTransitGreatestWindowDays
	goldenRatio := (math.Sqrt(5) - 1) / 2

	x1 := right - goldenRatio*(right-left)
	x2 := left + goldenRatio*(right-left)
	f1 := planetTransitStateAt(x1, cfg, planetTransitCoarseN).separationSquared
	f2 := planetTransitStateAt(x2, cfg, planetTransitCoarseN).separationSquared

	for i := 0; i < 80 && right-left > planetTransitGreatestToleranceDays; i++ {
		if f1 <= f2 {
			right = x2
			x2 = x1
			f2 = f1
			x1 = right - goldenRatio*(right-left)
			f1 = planetTransitStateAt(x1, cfg, planetTransitCoarseN).separationSquared
			continue
		}
		left = x1
		x1 = x2
		f1 = f2
		x2 = left + goldenRatio*(right-left)
		f2 = planetTransitStateAt(x2, cfg, planetTransitCoarseN).separationSquared
	}

	center := (left + right) / 2
	left = center - 2.0/24.0
	right = center + 2.0/24.0
	x1 = right - goldenRatio*(right-left)
	x2 = left + goldenRatio*(right-left)
	f1 = planetTransitStateAt(x1, cfg, -1).separationSquared
	f2 = planetTransitStateAt(x2, cfg, -1).separationSquared
	for i := 0; i < 80 && right-left > planetTransitGreatestToleranceDays; i++ {
		if f1 <= f2 {
			right = x2
			x2 = x1
			f2 = f1
			x1 = right - goldenRatio*(right-left)
			f1 = planetTransitStateAt(x1, cfg, -1).separationSquared
			continue
		}
		left = x1
		x1 = x2
		f1 = f2
		x2 = left + goldenRatio*(right-left)
		f2 = planetTransitStateAt(x2, cfg, -1).separationSquared
	}
	return (left + right) / 2
}

func planetTransitStateAt(jdTT float64, cfg planetTransitConfig, n int) planetTransitState {
	planetRA, planetDec := cfg.apparentRaDecN(jdTT, n)
	sunRA, sunDec := HSunApparentRaDecN(jdTT, n)
	separationArcsec := StarAngularSeparation(planetRA, planetDec, sunRA, sunDec) * 3600
	sunSemidiameter := SunSemidiameterN(jdTT, n)
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

func refinePlanetTransitContactTT(greatestTT float64, cfg planetTransitConfig, direction int, internal bool) (float64, bool) {
	if direction != -1 && direction != 1 {
		return 0, false
	}
	metric := func(jdTT float64) float64 {
		state := planetTransitStateAt(jdTT, cfg, -1)
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
	result.Greatest = TD2UT(result.Greatest, false)
	result.ExternalIngress = TD2UT(result.ExternalIngress, false)
	result.ExternalEgress = TD2UT(result.ExternalEgress, false)
	if result.InternalIngress != 0 {
		result.InternalIngress = TD2UT(result.InternalIngress, false)
	}
	if result.InternalEgress != 0 {
		result.InternalEgress = TD2UT(result.InternalEgress, false)
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

func isFiniteFloat(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
