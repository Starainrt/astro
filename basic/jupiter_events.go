package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

// Pos

const (
	JUPITER_S_PERIOD            = 1 / ((1 / 365.256363004) - (1 / 4332.59))
	jupiterEventSearchN         = 16
	jupiterPhaseCoarseTolerance = 30.0 / 86400.0
)

func jupiterSunLongitudeDelta(jde, degree float64, filter bool) float64 {
	sub := Limit360(Limit360(JupiterApparentLo(jde)-HSunApparentLo(jde)) - degree)
	if filter {
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
	}
	return sub
}

func jupiterSunLongitudeDeltaN(jde, degree float64, filter bool, n int) float64 {
	sub := Limit360(Limit360(JupiterApparentLoN(jde, n)-HSunApparentLoN(jde, n)) - degree)
	if filter {
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
	}
	return sub
}

func jupiterConjunctionFull(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	daysPerDegree := JUPITER_S_PERIOD / 360
	currentDelta := jupiterSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		longitudeDelta := jupiterSunLongitudeDelta(prevJD, degree, true)
		longitudeSlope := (jupiterSunLongitudeDelta(prevJD+0.000005, degree, true) - jupiterSunLongitudeDelta(prevJD-0.000005, degree, true)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return TD2UT(estimateJD, false)
}

func jupiterConjunction(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	daysPerDegree := JUPITER_S_PERIOD / 360
	currentDelta := jupiterSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		longitudeDelta := jupiterSunLongitudeDeltaN(prevJD, degree, true, jupiterEventSearchN)
		longitudeSlope := (jupiterSunLongitudeDeltaN(prevJD+0.000005, degree, true, jupiterEventSearchN) - jupiterSunLongitudeDeltaN(prevJD-0.000005, degree, true, jupiterEventSearchN)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= jupiterPhaseCoarseTolerance {
			break
		}
	}
	for {
		prevJD := estimateJD
		longitudeDelta := jupiterSunLongitudeDelta(prevJD, degree, true)
		longitudeSlope := (jupiterSunLongitudeDelta(prevJD+0.000005, degree, true) - jupiterSunLongitudeDelta(prevJD-0.000005, degree, true)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return TD2UT(estimateJD, false)
}

func LastJupiterConjunction(jde float64) float64 {
	return jupiterConjunction(jde, 0, 0)
}

func NextJupiterConjunction(jde float64) float64 {
	return jupiterConjunction(jde, 0, 1)
}

func LastJupiterOpposition(jde float64) float64 {
	return jupiterConjunction(jde, 180, 0)
}

func NextJupiterOpposition(jde float64) float64 {
	return jupiterConjunction(jde, 180, 1)
}

func NextJupiterEasternQuadrature(jde float64) float64 {
	return jupiterConjunction(jde, 90, 1)
}

func LastJupiterEasternQuadrature(jde float64) float64 {
	return jupiterConjunction(jde, 90, 0)
}

func NextJupiterWesternQuadrature(jde float64) float64 {
	return jupiterConjunction(jde, 270, 1)
}

func LastJupiterWesternQuadrature(jde float64) float64 {
	return jupiterConjunction(jde, 270, 0)
}

func jupiterRetrograde(jde float64, searchBeforeOpposition bool) float64 {
	//0=last 1=next
	raRate := func(jde float64, delta float64) float64 {
		sub := JupiterApparentRa(jde+delta) - JupiterApparentRa(jde-delta)
		if sub > 180 {
			sub -= 360
		}
		if sub < -180 {
			sub += 360
		}
		return sub / (2 * delta)
	}
	jde = jupiterConjunctionFull(jde, 180, 1)
	if searchBeforeOpposition {
		jde -= 60
	} else {
		jde += 60
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
	return TD2UT(bestJD, false)
}

func NextJupiterRetrogradeToPrograde(jde float64) float64 {
	date := jupiterRetrograde(jde, false)
	if date < jde {
		oppositionJD := jupiterConjunctionFull(jde, 180, 1)
		return jupiterRetrograde(oppositionJD+10, false)
	}
	return date
}

func LastJupiterRetrogradeToPrograde(jde float64) float64 {
	jde = jupiterConjunctionFull(jde, 180, 0) - 10
	date := jupiterRetrograde(jde, false)
	if date > jde {
		oppositionJD := jupiterConjunctionFull(jde, 180, 0)
		return jupiterRetrograde(oppositionJD-10, false)
	}
	return date
}

func NextJupiterProgradeToRetrograde(jde float64) float64 {
	date := jupiterRetrograde(jde, true)
	if date < jde {
		oppositionJD := jupiterConjunctionFull(jde, 180, 1)
		return jupiterRetrograde(oppositionJD+10, true)
	}
	return date
}

func LastJupiterProgradeToRetrograde(jde float64) float64 {
	jde = jupiterConjunctionFull(jde, 180, 0) - 10
	date := jupiterRetrograde(jde, true)
	if date > jde {
		oppositionJD := jupiterConjunctionFull(jde, 180, 0)
		return jupiterRetrograde(oppositionJD-10, true)
	}
	return date
}
