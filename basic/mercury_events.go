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
	for {
		prevJD := estimateJD
		longitudeDelta := mercuryConjunctionExactDelta(prevJD)
		longitudeSlope := (mercuryConjunctionExactDelta(prevJD+0.000005) - mercuryConjunctionExactDelta(prevJD-0.000005)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return estimateJD
}

func mercuryConjunctionLegacy(jde float64, next uint8) float64 {
	//0=last 1=next
	longitudeDeltaAt := func(jde float64) float64 {
		return mercuryConjunctionExactDelta(jde)
	}
	currentDelta := longitudeDeltaAt(jde)
	distanceTrend := math.Abs(longitudeDeltaAt(jde+1/86400.0)) - math.Abs(currentDelta)
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
	for {
		currentDelta := longitudeDeltaAt(jde)
		distanceTrend := math.Abs(longitudeDeltaAt(jde+1/86400.0)) - math.Abs(currentDelta)
		if math.Abs(currentDelta) > 12 || (distanceTrend > 0 && next == 1) || (distanceTrend < 0 && next == 0) {
			if next == 1 {
				jde += 2
			} else {
				jde -= 2
			}
			continue
		}
		break
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		longitudeDelta := longitudeDeltaAt(prevJD)
		longitudeSlope := (longitudeDeltaAt(prevJD+0.000005) - longitudeDeltaAt(prevJD-0.000005)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return TD2UT(estimateJD, false)
}

func mercuryConjunction(jde float64, next uint8) float64 {
	//0=last 1=next
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
	for {
		currentDelta := mercuryConjunctionExactDelta(jde)
		distanceTrend := math.Abs(mercuryConjunctionExactDelta(jde+1/86400.0)) - math.Abs(currentDelta)
		if math.Abs(currentDelta) > 12 || (distanceTrend > 0 && next == 1) || (distanceTrend < 0 && next == 0) {
			if next == 1 {
				jde += 2
			} else {
				jde -= 2
			}
			continue
		}
		break
	}

	inferior := mercuryConjunctionExactTT(jde, true)
	superior := mercuryConjunctionExactTT(jde, false)
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

func mercuryRetrograde(jde float64) float64 {
	//0=last 1=next
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
	lastConjunction := mercuryConjunctionLegacy(jde, 0)
	nextConjunction := mercuryConjunctionLegacy(jde, 1)
	currentRADelta := solarRADelta(jde)
	if currentRADelta > 0 {
		jde = lastConjunction + ((nextConjunction - lastConjunction) / 5.0 * 3.5)
	} else {
		jde = lastConjunction + ((nextConjunction - lastConjunction) / 5.5)
	}
	for {
		currentRate := mercuryRADerivative(jde, 1.0/86400.0)
		if math.Abs(currentRate) > 0.55 {
			jde += 2
			continue
		}
		break
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		rateValue := mercuryRADerivative(prevJD, 2.0/86400.0)
		rateSlope := (mercuryRADerivative(prevJD+15.0/86400.0, 2.0/86400.0) - mercuryRADerivative(prevJD-15.0/86400.0, 2.0/86400.0)) / (30.0 / 86400.0)
		estimateJD = prevJD - rateValue/rateSlope
		if math.Abs(estimateJD-prevJD) <= 30.0/86400.0 {
			break
		}
	}
	bestJD := eventZeroRefine(estimateJD, 15.0/86400.0, 0.5/86400.0, func(jd float64) float64 {
		return mercuryRADerivative(jd, 0.5/86400.0)
	})
	//fmt.Println((bestJD - lastConjunction) / (nextConjunction - lastConjunction))
	return TD2UT(bestJD, false)
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

func mercuryStationIsProgradeToRetrograde(eventUT float64) bool {
	for _, offset := range []float64{0.25, 0.5, 1.0} {
		before := mercuryRADerivative(eventUT-offset, 0.5/86400.0)
		after := mercuryRADerivative(eventUT+offset, 0.5/86400.0)
		if before > 0 && after < 0 {
			return true
		}
		if before < 0 && after > 0 {
			return false
		}
	}
	before := mercuryRADerivative(eventUT-0.25, 0.5/86400.0)
	after := mercuryRADerivative(eventUT+0.25, 0.5/86400.0)
	return before > after
}

func nextMercuryTypedStation(jde float64, progradeToRetrograde bool) float64 {
	date := NextMercuryRetrogradeStrict(jde)
	for mercuryStationIsProgradeToRetrograde(date) != progradeToRetrograde {
		date = NextMercuryRetrogradeStrict(eventUTNextQueryTT(date))
	}
	return date
}

func lastMercuryTypedStation(jde float64, progradeToRetrograde bool) float64 {
	date := LastMercuryRetrogradeStrict(jde)
	for mercuryStationIsProgradeToRetrograde(date) != progradeToRetrograde {
		date = LastMercuryRetrogradeStrict(eventUTLastQueryTT(date))
	}
	return date
}

func NextMercuryRetrograde(jde float64) float64 {
	date := mercuryRetrograde(jde)
	if !eventUTQueryAfterOrEqual(date, jde) {
		nextConjunction := NextMercuryConjunctionStrict(jde)
		return mercuryRetrograde(nextConjunction + 2)
	}
	return date
}

func LastMercuryRetrograde(jde float64) float64 {
	lastConjunction := LastMercuryConjunctionStrict(jde)
	date := mercuryRetrograde(lastConjunction + 2)
	if !eventUTQueryBeforeOrEqual(date, jde) {
		previousConjunction := LastMercuryConjunctionStrict(eventUTLastQueryTT(lastConjunction))
		return mercuryRetrograde(previousConjunction + 2)
	}
	return date
}

func LastMercuryRetrogradeStrict(jde float64) float64 {
	return LastMercuryRetrograde(jde)
}

func NextMercuryRetrogradeStrict(jde float64) float64 {
	return NextMercuryRetrograde(jde)
}

func NextMercuryProgradeToRetrograde(jde float64) float64 {
	return nextMercuryTypedStation(jde, true)
}

func NextMercuryRetrogradeToPrograde(jde float64) float64 {
	return nextMercuryTypedStation(jde, false)
}

func LastMercuryProgradeToRetrograde(jde float64) float64 {
	return lastMercuryTypedStation(jde, true)
}

func LastMercuryRetrogradeToPrograde(jde float64) float64 {
	return lastMercuryTypedStation(jde, false)
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
	return lastSuperior + innerEventEpsilon, inferior - innerEventEpsilon
}

func mercuryWestElongationWindowEndingAt(superior float64) (float64, float64) {
	lastInferior := LastMercuryInferiorConjunction(eventUTLastQueryTT(superior))
	return lastInferior + innerEventEpsilon, superior - innerEventEpsilon
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
	if east {
		start, windowEnd := mercuryEastElongationWindowContaining(jde)
		for {
			date := mercuryGreatestElongationInWindow(start, windowEnd)
			if eventUTQueryAfterOrEqual(date, jde) {
				return date
			}
			nextInferior := NextMercuryInferiorConjunction(eventUTNextQueryTT(windowEnd))
			start, windowEnd = mercuryEastElongationWindowEndingAt(nextInferior)
		}
	}
	start, windowEnd := mercuryWestElongationWindowContaining(jde)
	for {
		date := mercuryGreatestElongationInWindow(start, windowEnd)
		if eventUTQueryAfterOrEqual(date, jde) {
			return date
		}
		nextSuperior := NextMercurySuperiorConjunction(eventUTNextQueryTT(windowEnd))
		start, windowEnd = mercuryWestElongationWindowEndingAt(nextSuperior)
	}
}

func lastMercuryGreatestElongationTyped(jde float64, east bool) float64 {
	if east {
		start, windowEnd := mercuryEastElongationWindowContaining(jde)
		for {
			date := mercuryGreatestElongationInWindow(start, windowEnd)
			if eventUTQueryBeforeOrEqual(date, jde) {
				return date
			}
			prevInferior := LastMercuryInferiorConjunction(eventUTLastQueryTT(start))
			start, windowEnd = mercuryEastElongationWindowEndingAt(prevInferior)
		}
	}
	start, windowEnd := mercuryWestElongationWindowContaining(jde)
	for {
		date := mercuryGreatestElongationInWindow(start, windowEnd)
		if eventUTQueryBeforeOrEqual(date, jde) {
			return date
		}
		prevSuperior := LastMercurySuperiorConjunction(eventUTLastQueryTT(start))
		start, windowEnd = mercuryWestElongationWindowEndingAt(prevSuperior)
	}
}

func mercuryGreatestElongation(jde float64) float64 {
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
	for {
		currentRate := elongationRate(jde, 1.0/86400.0)
		if math.Abs(currentRate) > 0.4 {
			jde += 2
			continue
		}
		break
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		rateValue := elongationRate(prevJD, 2.0/86400.0)
		rateSlope := (elongationRate(prevJD+15.0/86400.0, 2.0/86400.0) - elongationRate(prevJD-15.0/86400.0, 2.0/86400.0)) / (30.0 / 86400.0)
		estimateJD = prevJD - rateValue/rateSlope
		if math.Abs(estimateJD-prevJD) <= 30.0/86400.0 {
			break
		}
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
