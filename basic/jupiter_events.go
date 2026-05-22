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

func jupiterRADerivative(jde, delta float64) float64 {
	sub := JupiterApparentRa(jde+delta) - JupiterApparentRa(jde-delta)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func jupiterRADerivativeN(jde, delta float64, n int) float64 {
	sub := JupiterApparentRaN(jde+delta, n) - JupiterApparentRaN(jde-delta, n)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
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
	return inclusiveLastPhaseEvent(jde, 0, jupiterConjunction)
}

func NextJupiterConjunction(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 0, jupiterConjunction)
}

func LastJupiterOpposition(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 180, jupiterConjunction)
}

func NextJupiterOpposition(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 180, jupiterConjunction)
}

func NextJupiterEasternQuadrature(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 90, jupiterConjunction)
}

func LastJupiterEasternQuadrature(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 90, jupiterConjunction)
}

func NextJupiterWesternQuadrature(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 270, jupiterConjunction)
}

func LastJupiterWesternQuadrature(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 270, jupiterConjunction)
}

func jupiterRetrogradeAroundOpposition(oppositionJD float64, searchBeforeOpposition bool) float64 {
	oppositionTT := TD2UT(oppositionJD, true)
	startTT := oppositionTT
	endTT := oppositionTT
	if searchBeforeOpposition {
		easternQuadratureUT := jupiterConjunction(oppositionTT, 90, 0)
		startTT = TD2UT(easternQuadratureUT, true)
	} else {
		westernQuadratureUT := jupiterConjunction(oppositionTT, 270, 1)
		endTT = TD2UT(westernQuadratureUT, true)
	}
	bestJD := zeroEventInWindow(startTT, endTT, 2.0, 2.0, 30.0/86400.0, func(jd float64) float64 {
		return jupiterRADerivativeN(jd, 1.0/86400.0, jupiterEventSearchN)
	}, func(jd float64) float64 {
		return jupiterRADerivative(jd, 0.5/86400.0)
	})
	return TD2UT(bestJD, false)
}

func NextJupiterRetrogradeToPrograde(jde float64) float64 {
	lastOppositionJD := jupiterConjunctionFull(jde, 180, 0)
	date := jupiterRetrogradeAroundOpposition(lastOppositionJD, false)
	if sameEventUTQueryTT(date, jde) || eventUTQueryAfterOrEqual(date, jde) {
		return date
	}
	nextOppositionJD := jupiterConjunctionFull(jde, 180, 1)
	return jupiterRetrogradeAroundOpposition(nextOppositionJD, false)
}

func LastJupiterRetrogradeToPrograde(jde float64) float64 {
	lastOppositionJD := jupiterConjunctionFull(jde, 180, 0)
	date := jupiterRetrogradeAroundOpposition(lastOppositionJD, false)
	if sameEventUTQueryTT(date, jde) || eventUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	previousOppositionJD := jupiterConjunctionFull(eventUTLastQueryTT(lastOppositionJD), 180, 0)
	return jupiterRetrogradeAroundOpposition(previousOppositionJD, false)
}

func NextJupiterProgradeToRetrograde(jde float64) float64 {
	nextOppositionJD := jupiterConjunctionFull(jde, 180, 1)
	date := jupiterRetrogradeAroundOpposition(nextOppositionJD, true)
	if sameEventUTQueryTT(date, jde) || eventUTQueryAfterOrEqual(date, jde) {
		return date
	}
	followingOppositionJD := jupiterConjunctionFull(eventUTNextQueryTT(nextOppositionJD), 180, 1)
	return jupiterRetrogradeAroundOpposition(followingOppositionJD, true)
}

func LastJupiterProgradeToRetrograde(jde float64) float64 {
	nextOppositionJD := jupiterConjunctionFull(jde, 180, 1)
	date := jupiterRetrogradeAroundOpposition(nextOppositionJD, true)
	if sameEventUTQueryTT(date, jde) || eventUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	lastOppositionJD := jupiterConjunctionFull(jde, 180, 0)
	return jupiterRetrogradeAroundOpposition(lastOppositionJD, true)
}
