package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

const (
	MERCURY_S_PERIOD                    = 1 / ((1 / 87.9691) - (1 / 365.256363004))
	mercuryConjunctionDerivativeStepDay = 2e-5 * 36525.0
	mercuryLightTimeDaysPerAU           = 0.0057755183
	mercuryEventSearchN                 = 16
	mercuryStationWindowDays            = 30.0
	mercuryStationDerivativeStepDay     = 0.01
	mercuryStationCoarseStepDay         = 2.0
	mercuryStationHalfWindowDay         = 2.0
	mercuryStationMotionTolerance       = 1e-3
)

type mercuryConjunctionLBR struct {
	lo float64
	bo float64
	r  float64
}

type mercuryConjunctionGeo struct {
	lo   float64
	bo   float64
	dist float64
}

type mercuryConjunctionResult struct {
	diff         float64
	sunLightDays float64
	geoLightDays float64
}

func mercuryHelioN(planetIndex int, jd float64, n int) mercuryConjunctionLBR {
	return mercuryConjunctionLBR{
		lo: planet.WherePlanetN(planetIndex, 0, jd, n),
		bo: planet.WherePlanetN(planetIndex, 1, jd, n),
		r:  planet.WherePlanetN(planetIndex, 2, jd, n),
	}
}

func mercuryGeocentric(planetPos, earthPos mercuryConjunctionLBR) mercuryConjunctionGeo {
	x := planetPos.r*Cos(planetPos.bo)*Cos(planetPos.lo) - earthPos.r*Cos(earthPos.bo)*Cos(earthPos.lo)
	y := planetPos.r*Cos(planetPos.bo)*Sin(planetPos.lo) - earthPos.r*Cos(earthPos.bo)*Sin(earthPos.lo)
	z := planetPos.r*Sin(planetPos.bo) - earthPos.r*Sin(earthPos.bo)
	dist := math.Sqrt(x*x + y*y + z*z)
	return mercuryConjunctionGeo{
		lo:   Limit360(math.Atan2(y, x) * 180 / math.Pi),
		bo:   math.Atan2(z, math.Sqrt(x*x+y*y)) * 180 / math.Pi,
		dist: dist,
	}
}

func mercuryConjunctionAngleDelta(diff float64) float64 {
	diff = Limit360(diff)
	if diff > 180 {
		diff -= 360
	}
	if diff < -180 {
		diff += 360
	}
	return diff
}

func mercuryConjunctionHeliocentricDelta(jd, targetDeg float64, n int) float64 {
	planetLo := planet.WherePlanetN(1, 0, jd, n)
	earthLo := planet.WherePlanetN(-1, 0, jd, n)
	return mercuryConjunctionAngleDelta(planetLo - earthLo - targetDeg)
}

func mercuryConjunctionDifference(jd float64, n int, targetDeg, sunLightDays, geoLightDays float64) mercuryConjunctionResult {
	earthForSun := mercuryHelioN(-1, jd-sunLightDays, n)
	sunLo := Limit360(earthForSun.lo + 180)
	earth := mercuryHelioN(-1, jd-geoLightDays, n)
	planetPos := mercuryHelioN(1, jd-geoLightDays, n)
	geo := mercuryGeocentric(planetPos, earth)
	return mercuryConjunctionResult{
		diff:         mercuryConjunctionAngleDelta(geo.lo - sunLo - targetDeg),
		sunLightDays: earthForSun.r * mercuryLightTimeDaysPerAU,
		geoLightDays: geo.dist * mercuryLightTimeDaysPerAU,
	}
}

func mercuryConjunctionExactDelta(jd float64) float64 {
	return mercuryConjunctionAngleDelta(MercuryApparentLo(jd) - HSunApparentLo(jd))
}

func mercuryConjunctionApproxTT(seed float64, inferior bool) float64 {
	heliocentricTarget := 180.0
	if inferior {
		heliocentricTarget = 0
	}
	jd := seed
	for i := 0; i < 6; i++ {
		jd -= mercuryConjunctionHeliocentricDelta(jd, heliocentricTarget, 8) / (360.0 / MERCURY_S_PERIOD)
	}

	startSample := mercuryConjunctionDifference(jd, 8, 0, 0, 0)
	nextSample := mercuryConjunctionDifference(jd+mercuryConjunctionDerivativeStepDay, 8, 0, 0, 0)
	diffSlope := mercuryConjunctionAngleDelta(nextSample.diff-startSample.diff) / mercuryConjunctionDerivativeStepDay

	refined := mercuryConjunctionDifference(jd, 40, 0, startSample.sunLightDays, startSample.geoLightDays)
	jd -= refined.diff / diffSlope
	final := mercuryConjunctionDifference(jd, -1, 0, refined.sunLightDays, refined.geoLightDays)
	jd -= final.diff / diffSlope
	return jd
}

func mercuryConjunctionExactTT(seed float64, inferior bool) float64 {
	estimateJD := mercuryConjunctionApproxTT(seed, inferior)
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJD := estimateJD
		longitudeDelta := mercuryConjunctionExactDelta(prevJD)
		longitudeSlope := (mercuryConjunctionExactDelta(prevJD+0.000005) - mercuryConjunctionExactDelta(prevJD-0.000005)) / 0.00001
		nextJD := prevJD - longitudeDelta/longitudeSlope
		estimateJD = nextJD
		if math.Abs(nextJD-prevJD) <= 0.00001 {
			converged = true
			break
		}
	}
	if !converged {
		return math.NaN()
	}
	return estimateJD
}

func mercuryConjunction(jde float64, next uint8) float64 {
	//0=last 1=next
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	if math.Abs(mercuryConjunctionExactDelta(jde)) <= 30.0/86400.0 {
		best := math.NaN()
		consider := func(inferior bool) {
			eventUT := TD2UT(mercuryConjunctionExactTT(jde, inferior), false)
			if next == 0 && !eventUTQueryBeforeOrEqual(eventUT, jde) {
				return
			}
			if next == 1 && !eventUTQueryAfterOrEqual(eventUT, jde) {
				return
			}
			if math.IsNaN(best) || math.Abs(eventUTQueryTTDelta(eventUT, jde)) < math.Abs(eventUTQueryTTDelta(best, jde)) {
				best = eventUT
			}
		}
		consider(true)
		consider(false)
		if !math.IsNaN(best) {
			return best
		}
	}
	currentDelta := mercuryConjunctionExactDelta(jde)
	// pos 大于0:远离太阳 小于0:靠近太阳
	distanceTrend := math.Abs(mercuryConjunctionExactDelta(jde+1/86400.0)) - math.Abs(currentDelta)
	if distanceTrend >= 0 && next == 1 && currentDelta > 0 {
		jde += MERCURY_S_PERIOD/8.0 + 2
	}
	if distanceTrend >= 0 && next == 1 && currentDelta < 0 {
		jde += MERCURY_S_PERIOD/6.0 + 2
	}
	if distanceTrend <= 0 && next == 0 && currentDelta < 0 {
		jde -= MERCURY_S_PERIOD/8.0 + 2
	}
	if distanceTrend <= 0 && next == 0 && currentDelta > 0 {
		jde -= MERCURY_S_PERIOD/6.0 + 2
	}
	found := false
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		currentDelta := mercuryConjunctionExactDelta(jde)
		nextDelta := mercuryConjunctionExactDelta(jde + 1/86400.0)
		if !isFiniteFloat(currentDelta) || !isFiniteFloat(nextDelta) {
			return math.NaN()
		}
		distanceTrend := math.Abs(nextDelta) - math.Abs(currentDelta)
		if math.Abs(currentDelta) > 12 || (distanceTrend > 0 && next == 1) || (distanceTrend < 0 && next == 0) {
			if next == 1 {
				jde += 2
			} else {
				jde -= 2
			}
			continue
		}
		found = true
		break
	}
	if !found {
		return math.NaN()
	}

	inferior := mercuryConjunctionExactTT(jde, true)
	superior := mercuryConjunctionExactTT(jde, false)
	if !isFiniteFloat(inferior) || !isFiniteFloat(superior) {
		return math.NaN()
	}
	best := inferior
	if math.Abs(superior-jde) < math.Abs(inferior-jde) {
		best = superior
	}
	return TD2UT(best, false)
}

func LastMercuryConjunction(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastMercuryConjunctionStrict, NextMercuryConjunctionStrict)
}

func NextMercuryConjunction(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastMercuryConjunctionStrict, NextMercuryConjunctionStrict)
}

func LastMercuryConjunctionStrict(jde float64) float64 {
	return mercuryConjunction(jde, 0)
}

func NextMercuryConjunctionStrict(jde float64) float64 {
	return mercuryConjunction(jde, 1)
}

func NextMercuryInferiorConjunction(jde float64) float64 {
	date := NextMercuryConjunctionStrict(jde)
	if EarthMercuryAway(date) > EarthAway(date) {
		return NextMercuryConjunctionStrict(date + 2)
	}
	return date
}

func NextMercurySuperiorConjunction(jde float64) float64 {
	date := NextMercuryConjunctionStrict(jde)
	if EarthMercuryAway(date) < EarthAway(date) {
		return NextMercuryConjunctionStrict(date + 2)
	}
	return date
}

func LastMercuryInferiorConjunction(jde float64) float64 {
	date := LastMercuryConjunctionStrict(jde)
	if EarthMercuryAway(date) > EarthAway(date) {
		return LastMercuryConjunctionStrict(date - 2)
	}
	return date
}

func LastMercurySuperiorConjunction(jde float64) float64 {
	date := LastMercuryConjunctionStrict(jde)
	if EarthMercuryAway(date) < EarthAway(date) {
		return LastMercuryConjunctionStrict(date - 2)
	}
	return date
}

func mercuryRADerivative(jde, delta float64) float64 {
	sub := MercuryApparentRa(jde+delta) - MercuryApparentRa(jde-delta)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func mercuryRADerivativeN(jde, delta float64, n int) float64 {
	sub := MercuryApparentRaN(jde+delta, n) - MercuryApparentRaN(jde-delta, n)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func mercuryStationInWindow(startTT, endTT float64) float64 {
	bestJD := zeroEventInWindow(startTT, endTT, mercuryStationCoarseStepDay, mercuryStationHalfWindowDay, 30.0/86400.0, func(jd float64) float64 {
		return mercuryRADerivativeN(jd, mercuryStationDerivativeStepDay, mercuryEventSearchN)
	}, func(jd float64) float64 {
		return mercuryRADerivative(jd, mercuryStationDerivativeStepDay)
	})
	return TD2UT(bestJD, false)
}

func mercuryStationBetween(startTT, endTT float64) bool {
	if endTT < startTT {
		startTT, endTT = endTT, startTT
	}
	if endTT-startTT <= 0 {
		return false
	}
	if endTT-startTT > mercuryStationWindowDays {
		return true
	}
	// 截断扫描足以判断单候选快速路径是否安全 / A truncated scan is enough to decide whether the one-candidate fast path is safe.
	left := startTT
	leftValue := mercuryRADerivativeN(left, mercuryStationDerivativeStepDay, mercuryEventSearchN)
	for left < endTT {
		right := left + mercuryStationCoarseStepDay
		if right > endTT {
			right = endTT
		}
		rightValue := mercuryRADerivativeN(right, mercuryStationDerivativeStepDay, mercuryEventSearchN)
		if leftValue == 0 || leftValue*rightValue < 0 || rightValue == 0 {
			return true
		}
		left = right
		leftValue = rightValue
	}
	return false
}

func mercuryProgradeToRetrogradeAroundInferior(inferiorUT float64) float64 {
	inferiorTT := TD2UT(inferiorUT, true)
	return mercuryStationInWindow(inferiorTT-mercuryStationWindowDays, inferiorTT)
}

func mercuryRetrogradeToProgradeAroundInferior(inferiorUT float64) float64 {
	inferiorTT := TD2UT(inferiorUT, true)
	return mercuryStationInWindow(inferiorTT, inferiorTT+mercuryStationWindowDays)
}

func NextMercuryProgradeToRetrograde(jde float64) float64 {
	inferior := NextMercuryInferiorConjunction(jde)
	date := mercuryProgradeToRetrogradeAroundInferior(inferior)
	if eventUTQueryAfterOrEqual(date, jde) {
		return date
	}
	followingInferior := NextMercuryInferiorConjunction(eventUTNextQueryTT(inferior))
	return mercuryProgradeToRetrogradeAroundInferior(followingInferior)
}

func NextMercuryRetrogradeToPrograde(jde float64) float64 {
	inferior := LastMercuryInferiorConjunction(jde)
	date := mercuryRetrogradeToProgradeAroundInferior(inferior)
	if eventUTQueryAfterOrEqual(date, jde) {
		return date
	}
	nextInferior := NextMercuryInferiorConjunction(eventUTNextQueryTT(inferior))
	return mercuryRetrogradeToProgradeAroundInferior(nextInferior)
}

func LastMercuryProgradeToRetrograde(jde float64) float64 {
	inferior := NextMercuryInferiorConjunction(jde)
	date := mercuryProgradeToRetrogradeAroundInferior(inferior)
	if eventUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	previousInferior := LastMercuryInferiorConjunction(eventUTLastQueryTT(inferior))
	return mercuryProgradeToRetrogradeAroundInferior(previousInferior)
}

func LastMercuryRetrogradeToPrograde(jde float64) float64 {
	inferior := LastMercuryInferiorConjunction(jde)
	date := mercuryRetrogradeToProgradeAroundInferior(inferior)
	if eventUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	previousInferior := LastMercuryInferiorConjunction(eventUTLastQueryTT(inferior))
	return mercuryRetrogradeToProgradeAroundInferior(previousInferior)
}

func nextMercuryRetrogradeFromTyped(jde float64) float64 {
	p2r := NextMercuryProgradeToRetrograde(jde)
	r2p := NextMercuryRetrogradeToPrograde(jde)
	if p2r < r2p {
		return p2r
	}
	return r2p
}

func NextMercuryRetrograde(jde float64) float64 {
	motion := mercuryRADerivative(jde, mercuryStationDerivativeStepDay)
	if motion > mercuryStationMotionTolerance {
		p2r := NextMercuryProgradeToRetrograde(jde)
		if !mercuryStationBetween(jde, TD2UT(p2r, true)) {
			return p2r
		}
		r2p := NextMercuryRetrogradeToPrograde(jde)
		if p2r < r2p {
			return p2r
		}
		return r2p
	}
	if motion < -mercuryStationMotionTolerance {
		r2p := NextMercuryRetrogradeToPrograde(jde)
		if !mercuryStationBetween(jde, TD2UT(r2p, true)) {
			return r2p
		}
		p2r := NextMercuryProgradeToRetrograde(jde)
		if p2r < r2p {
			return p2r
		}
		return r2p
	}
	return nextMercuryRetrogradeFromTyped(jde)
}

func lastMercuryRetrogradeFromTyped(jde float64) float64 {
	p2r := LastMercuryProgradeToRetrograde(jde)
	r2p := LastMercuryRetrogradeToPrograde(jde)
	if p2r > r2p {
		return p2r
	}
	return r2p
}

func LastMercuryRetrograde(jde float64) float64 {
	motion := mercuryRADerivative(jde, mercuryStationDerivativeStepDay)
	if motion > mercuryStationMotionTolerance {
		r2p := LastMercuryRetrogradeToPrograde(jde)
		if !mercuryStationBetween(TD2UT(r2p, true), jde) {
			return r2p
		}
		p2r := LastMercuryProgradeToRetrograde(jde)
		if p2r > r2p {
			return p2r
		}
		return r2p
	}
	if motion < -mercuryStationMotionTolerance {
		p2r := LastMercuryProgradeToRetrograde(jde)
		if !mercuryStationBetween(TD2UT(p2r, true), jde) {
			return p2r
		}
		r2p := LastMercuryRetrogradeToPrograde(jde)
		if p2r > r2p {
			return p2r
		}
		return r2p
	}
	return lastMercuryRetrogradeFromTyped(jde)
}

func LastMercuryRetrogradeStrict(jde float64) float64 {
	return LastMercuryRetrograde(jde)
}

func NextMercuryRetrogradeStrict(jde float64) float64 {
	return NextMercuryRetrograde(jde)
}

func MercurySunElongation(jde float64) float64 {
	lo1, bo1 := MercuryApparentLoBo(jde)
	lo2 := HSunApparentLo(jde)
	bo2 := HSunTrueBo(jde)
	return StarAngularSeparation(lo1, bo1, lo2, bo2)
}

func mercurySunElongationN(jde float64, n int) float64 {
	lo1, bo1 := MercuryApparentLoBoN(jde, n)
	lo2 := HSunApparentLoN(jde, n)
	bo2 := HSunTrueBoN(jde, n)
	return StarAngularSeparation(lo1, bo1, lo2, bo2)
}

func mercuryTrueElongationN(jde float64, n int) float64 {
	earth := mercuryHelioN(-1, jde, n)
	planetPos := mercuryHelioN(1, jde, n)
	geo := mercuryGeocentric(planetPos, earth)
	return StarAngularSeparation(geo.lo, geo.bo, HSunTrueLoN(jde, n), HSunTrueBoN(jde, n))
}

func mercuryGreatestElongationInWindow(start, end float64) float64 {
	best := maximizeInWindow(start, end, 2.0, func(jd float64) float64 {
		return mercuryTrueElongationN(jd, mercuryEventSearchN)
	}, func(jd float64) float64 {
		return mercuryTrueElongationN(jd, -1)
	})
	return TD2UT(best, false)
}

func mercuryEastElongationWindowEndingAt(inferior float64) (float64, float64) {
	lastSuperior := LastMercurySuperiorConjunction(eventUTLastQueryTT(inferior))
	return lastSuperior + innerEventWindowPadding, inferior - innerEventWindowPadding
}

func mercuryWestElongationWindowEndingAt(superior float64) (float64, float64) {
	lastInferior := LastMercuryInferiorConjunction(eventUTLastQueryTT(superior))
	return lastInferior + innerEventWindowPadding, superior - innerEventWindowPadding
}

func mercuryEastElongationWindowContaining(jde float64) (float64, float64) {
	nextInferior := NextMercuryInferiorConjunction(jde)
	start, end := mercuryEastElongationWindowEndingAt(nextInferior)
	if eventUTQueryBeforeOrEqual(start, jde) {
		return start, end
	}
	currentInferior := LastMercuryInferiorConjunction(jde)
	return mercuryEastElongationWindowEndingAt(currentInferior)
}

func mercuryWestElongationWindowContaining(jde float64) (float64, float64) {
	nextSuperior := NextMercurySuperiorConjunction(jde)
	start, end := mercuryWestElongationWindowEndingAt(nextSuperior)
	if eventUTQueryBeforeOrEqual(start, jde) {
		return start, end
	}
	currentSuperior := LastMercurySuperiorConjunction(jde)
	return mercuryWestElongationWindowEndingAt(currentSuperior)
}

func nextMercuryGreatestElongationTyped(jde float64, east bool) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	if east {
		start, windowEnd := mercuryEastElongationWindowContaining(jde)
		for i := 0; i < eventDirectionalSearchIterations; i++ {
			date := mercuryGreatestElongationInWindow(start, windowEnd)
			if !isFiniteFloat(date) {
				return math.NaN()
			}
			if eventUTQueryAfterOrEqual(date, jde) {
				return date
			}
			nextInferior := NextMercuryInferiorConjunction(eventUTNextQueryTT(windowEnd))
			start, windowEnd = mercuryEastElongationWindowEndingAt(nextInferior)
		}
		return math.NaN()
	}
	start, windowEnd := mercuryWestElongationWindowContaining(jde)
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		date := mercuryGreatestElongationInWindow(start, windowEnd)
		if !isFiniteFloat(date) {
			return math.NaN()
		}
		if eventUTQueryAfterOrEqual(date, jde) {
			return date
		}
		nextSuperior := NextMercurySuperiorConjunction(eventUTNextQueryTT(windowEnd))
		start, windowEnd = mercuryWestElongationWindowEndingAt(nextSuperior)
	}
	return math.NaN()
}

func lastMercuryGreatestElongationTyped(jde float64, east bool) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	if east {
		start, windowEnd := mercuryEastElongationWindowContaining(jde)
		for i := 0; i < eventDirectionalSearchIterations; i++ {
			date := mercuryGreatestElongationInWindow(start, windowEnd)
			if !isFiniteFloat(date) {
				return math.NaN()
			}
			if eventUTQueryBeforeOrEqual(date, jde) {
				return date
			}
			prevInferior := LastMercuryInferiorConjunction(eventUTLastQueryTT(start))
			start, windowEnd = mercuryEastElongationWindowEndingAt(prevInferior)
		}
		return math.NaN()
	}
	start, windowEnd := mercuryWestElongationWindowContaining(jde)
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		date := mercuryGreatestElongationInWindow(start, windowEnd)
		if !isFiniteFloat(date) {
			return math.NaN()
		}
		if eventUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		prevSuperior := LastMercurySuperiorConjunction(eventUTLastQueryTT(start))
		start, windowEnd = mercuryWestElongationWindowEndingAt(prevSuperior)
	}
	return math.NaN()
}

func mercuryGreatestElongation(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	solarRADelta := func(jde float64) float64 {
		sub := Limit360(MercuryApparentRa(jde) - SunApparentRa(jde))
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub
	}
	elongationRate := func(jde float64, delta float64) float64 {
		sub := MercurySunElongation(jde+delta) - MercurySunElongation(jde-delta)
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub / (2 * delta)
	}
	lastConjunction := LastMercuryConjunctionStrict(jde)
	nextConjunction := NextMercuryConjunctionStrict(jde)
	currentRADelta := solarRADelta(jde)
	if currentRADelta > 0 {
		jde = lastConjunction + ((nextConjunction - lastConjunction) / 5.0 * 2.0)
	} else {
		jde = lastConjunction + ((nextConjunction - lastConjunction) / 6.0)
	}
	found := false
	for i := 0; i < eventDirectionalSearchIterations; i++ {
		currentRate := elongationRate(jde, 1.0/86400.0)
		if !isFiniteFloat(currentRate) {
			return math.NaN()
		}
		if math.Abs(currentRate) > 0.4 {
			jde += 2
			continue
		}
		found = true
		break
	}
	if !found {
		return math.NaN()
	}
	estimateJD := jde
	var ok bool
	estimateJD, ok = eventNewtonRefine(estimateJD, 30.0/86400.0, func(prevJD float64) float64 {
		rateValue := elongationRate(prevJD, 2.0/86400.0)
		rateSlope := (elongationRate(prevJD+15.0/86400.0, 2.0/86400.0) - elongationRate(prevJD-15.0/86400.0, 2.0/86400.0)) / (30.0 / 86400.0)
		return rateValue / rateSlope
	})
	if !ok {
		return math.NaN()
	}
	bestJD := eventZeroRefine(estimateJD, 15.0/86400.0, 0.5/86400.0, func(jd float64) float64 {
		return elongationRate(jd, 0.5/86400.0)
	})
	//fmt.Println((bestJD - lastConjunction) / (nextConjunction - lastConjunction))
	return TD2UT(bestJD, false)
}

func NextMercuryGreatestElongation(jde float64) float64 {
	east := NextMercuryGreatestElongationEast(jde)
	west := NextMercuryGreatestElongationWest(jde)
	if sameEventJD(east, west) {
		return east
	}
	if east < west {
		return east
	}
	return west
}

func LastMercuryGreatestElongation(jde float64) float64 {
	east := LastMercuryGreatestElongationEast(jde)
	west := LastMercuryGreatestElongationWest(jde)
	if sameEventJD(east, west) {
		return east
	}
	if east > west {
		return east
	}
	return west
}

func LastMercuryInferiorConjunctionInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastMercuryInferiorConjunction, NextMercuryInferiorConjunction)
}

func NextMercuryInferiorConjunctionInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastMercuryInferiorConjunction, NextMercuryInferiorConjunction)
}

func LastMercurySuperiorConjunctionInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastMercurySuperiorConjunction, NextMercurySuperiorConjunction)
}

func NextMercurySuperiorConjunctionInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastMercurySuperiorConjunction, NextMercurySuperiorConjunction)
}

func LastMercuryRetrogradeInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastMercuryRetrograde, NextMercuryRetrograde)
}

func NextMercuryRetrogradeInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastMercuryRetrograde, NextMercuryRetrograde)
}

func LastMercuryProgradeToRetrogradeInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastMercuryProgradeToRetrograde, NextMercuryProgradeToRetrograde)
}

func NextMercuryProgradeToRetrogradeInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastMercuryProgradeToRetrograde, NextMercuryProgradeToRetrograde)
}

func LastMercuryRetrogradeToProgradeInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastMercuryRetrogradeToPrograde, NextMercuryRetrogradeToPrograde)
}

func NextMercuryRetrogradeToProgradeInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastMercuryRetrogradeToPrograde, NextMercuryRetrogradeToPrograde)
}

func LastMercuryGreatestElongationInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastMercuryGreatestElongation, NextMercuryGreatestElongation)
}

func NextMercuryGreatestElongationInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastMercuryGreatestElongation, NextMercuryGreatestElongation)
}

func LastMercuryGreatestElongationEastInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastMercuryGreatestElongationEast, NextMercuryGreatestElongationEast)
}

func NextMercuryGreatestElongationEastInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastMercuryGreatestElongationEast, NextMercuryGreatestElongationEast)
}

func LastMercuryGreatestElongationWestInclusive(jde float64) float64 {
	return inclusiveLastSimpleEvent(jde, LastMercuryGreatestElongationWest, NextMercuryGreatestElongationWest)
}

func NextMercuryGreatestElongationWestInclusive(jde float64) float64 {
	return inclusiveNextSimpleEvent(jde, LastMercuryGreatestElongationWest, NextMercuryGreatestElongationWest)
}

func NextMercuryGreatestElongationEast(jde float64) float64 {
	return nextMercuryGreatestElongationTyped(jde, true)
}

func NextMercuryGreatestElongationWest(jde float64) float64 {
	return nextMercuryGreatestElongationTyped(jde, false)
}

func LastMercuryGreatestElongationEast(jde float64) float64 {
	return lastMercuryGreatestElongationTyped(jde, true)
}

func LastMercuryGreatestElongationWest(jde float64) float64 {
	return lastMercuryGreatestElongationTyped(jde, false)
}
