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
	return mercuryConjunction(jde, 0)
}

func NextMercuryConjunction(jde float64) float64 {
	return mercuryConjunction(jde, 1)
}

func NextMercuryInferiorConjunction(jde float64) float64 {
	date := NextMercuryConjunction(jde)
	if EarthMercuryAway(date) > EarthAway(date) {
		return NextMercuryConjunction(date + 2)
	}
	return date
}

func NextMercurySuperiorConjunction(jde float64) float64 {
	date := NextMercuryConjunction(jde)
	if EarthMercuryAway(date) < EarthAway(date) {
		return NextMercuryConjunction(date + 2)
	}
	return date
}

func LastMercuryInferiorConjunction(jde float64) float64 {
	date := LastMercuryConjunction(jde)
	if EarthMercuryAway(date) > EarthAway(date) {
		return LastMercuryConjunction(date - 2)
	}
	return date
}

func LastMercurySuperiorConjunction(jde float64) float64 {
	date := LastMercuryConjunction(jde)
	if EarthMercuryAway(date) < EarthAway(date) {
		return LastMercuryConjunction(date - 2)
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
	raRate := func(jde float64, delta float64) float64 {
		sub := MercuryApparentRa(jde+delta) - MercuryApparentRa(jde-delta)
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub / (2 * delta)
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
		currentRate := raRate(jde, 1.0/86400.0)
		if math.Abs(currentRate) > 0.55 {
			jde += 2
			continue
		}
		break
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		rateValue := raRate(prevJD, 2.0/86400.0)
		rateSlope := (raRate(prevJD+15.0/86400.0, 2.0/86400.0) - raRate(prevJD-15.0/86400.0, 2.0/86400.0)) / (30.0 / 86400.0)
		estimateJD = prevJD - rateValue/rateSlope
		if math.Abs(estimateJD-prevJD) <= 30.0/86400.0 {
			break
		}
	}
	bestJD := eventZeroRefine(estimateJD, 15.0/86400.0, 0.5/86400.0, func(jd float64) float64 {
		return raRate(jd, 0.5/86400.0)
	})
	//fmt.Println((bestJD - lastConjunction) / (nextConjunction - lastConjunction))
	return TD2UT(bestJD, false)
}

func NextMercuryRetrograde(jde float64) float64 {
	date := mercuryRetrograde(jde)
	if date < jde {
		nextConjunction := mercuryConjunctionLegacy(jde, 1)
		return mercuryRetrograde(nextConjunction + 2)
	}
	return date
}

func LastMercuryRetrograde(jde float64) float64 {
	lastConjunction := mercuryConjunctionLegacy(jde, 0)
	date := mercuryRetrograde(lastConjunction + 2)
	if date > jde {
		previousConjunction := mercuryConjunctionLegacy(lastConjunction-2, 0)
		return mercuryRetrograde(previousConjunction + 2)
	}
	return date
}

func NextMercuryProgradeToRetrograde(jde float64) float64 {
	date := NextMercuryRetrograde(jde)
	sub := Limit360(MercuryApparentRa(date) - SunApparentRa(date))
	if sub > 180 {
		return NextMercuryRetrograde(date + MERCURY_S_PERIOD/2)
	}
	return date
}

func NextMercuryRetrogradeToPrograde(jde float64) float64 {
	date := NextMercuryRetrograde(jde)
	sub := Limit360(MercuryApparentRa(date) - SunApparentRa(date))
	if sub < 180 {
		return NextMercuryRetrograde(date + 12)
	}
	return date
}

func LastMercuryProgradeToRetrograde(jde float64) float64 {
	date := LastMercuryRetrograde(jde)
	sub := Limit360(MercuryApparentRa(date) - SunApparentRa(date))
	if sub > 180 {
		return LastMercuryRetrograde(date - 12)
	}
	return date
}

func LastMercuryRetrogradeToPrograde(jde float64) float64 {
	date := LastMercuryRetrograde(jde)
	sub := Limit360(MercuryApparentRa(date) - SunApparentRa(date))
	if sub < 180 {
		return LastMercuryRetrograde(date - MERCURY_S_PERIOD/2)
	}
	return date
}

func MercurySunElongation(jde float64) float64 {
	lo1, bo1 := MercuryApparentLoBo(jde)
	lo2 := SunApparentLo(jde)
	bo2 := HSunTrueBo(jde)
	return StarAngularSeparation(lo1, bo1, lo2, bo2)
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
	lastConjunction := mercuryConjunctionLegacy(jde, 0)
	nextConjunction := mercuryConjunctionLegacy(jde, 1)
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
	date := mercuryGreatestElongation(jde)
	if date < jde {
		nextConjunction := mercuryConjunctionLegacy(jde, 1)
		return mercuryGreatestElongation(nextConjunction + 2)
	}
	return date
}

func LastMercuryGreatestElongation(jde float64) float64 {
	lastConjunction := mercuryConjunctionLegacy(jde, 0)
	date := mercuryGreatestElongation(lastConjunction + 2)
	if date > jde {
		previousConjunction := mercuryConjunctionLegacy(lastConjunction-2, 0)
		return mercuryGreatestElongation(previousConjunction + 2)
	}
	return date
}

func NextMercuryGreatestElongationEast(jde float64) float64 {
	date := NextMercuryGreatestElongation(jde)
	sub := Limit360(MercuryApparentRa(date) - SunApparentRa(date))
	if sub > 180 {
		return NextMercuryGreatestElongation(date + 1)
	}
	return date
}

func NextMercuryGreatestElongationWest(jde float64) float64 {
	date := NextMercuryGreatestElongation(jde)
	sub := Limit360(MercuryApparentRa(date) - SunApparentRa(date))
	if sub < 180 {
		return NextMercuryGreatestElongation(date + 1)
	}
	return date
}

func LastMercuryGreatestElongationEast(jde float64) float64 {
	date := LastMercuryGreatestElongation(jde)
	sub := Limit360(MercuryApparentRa(date) - SunApparentRa(date))
	if sub > 180 {
		return LastMercuryGreatestElongation(date - 1)
	}
	return date
}

func LastMercuryGreatestElongationWest(jde float64) float64 {
	date := LastMercuryGreatestElongation(jde)
	sub := Limit360(MercuryApparentRa(date) - SunApparentRa(date))
	if sub < 180 {
		return LastMercuryGreatestElongation(date - 1)
	}
	return date
}
