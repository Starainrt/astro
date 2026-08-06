package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

// Pos

const (
	MARS_S_PERIOD                = 1 / ((1 / 365.256363004) - (1 / 686.98))
	marsEventSearchN             = 16
	marsPhaseCoarseTolerance     = 30.0 / 86400.0
	marsStationDerivativeStepDay = 0.01
	marsStationCoarseStepDay     = 6.0
	marsStationHalfWindowDay     = 6.0
)

func marsSunLongitudeDelta(jde, degree float64, filter bool) float64 {
	sub := Limit360(Limit360(MarsApparentLo(jde)-HSunApparentLo(jde)) - degree)
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

func marsSunLongitudeDeltaN(jde, degree float64, filter bool, n int) float64 {
	sub := Limit360(Limit360(MarsApparentLoN(jde, n)-HSunApparentLoN(jde, n)) - degree)
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

func marsRADerivative(jde, val float64) float64 {
	sub := MarsApparentRa(jde+val) - MarsApparentRa(jde-val)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * val)
}

func marsRADerivativeN(jde, val float64, n int) float64 {
	sub := MarsApparentRaN(jde+val, n) - MarsApparentRaN(jde-val, n)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * val)
}

func marsConjunctionFull(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	if !isFiniteFloat(jde) || !isFiniteFloat(degree) {
		return math.NaN()
	}
	daysPerDegree := MARS_S_PERIOD / 360
	currentDelta := marsSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJD := jde
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJD := estimateJD
		longitudeDelta := marsSunLongitudeDelta(prevJD, degree, true)
		longitudeSlope := (marsSunLongitudeDelta(prevJD+0.000005, degree, true) - marsSunLongitudeDelta(prevJD-0.000005, degree, true)) / 0.00001
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
	return TD2UT(estimateJD, false)
}

func marsConjunction(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	if !isFiniteFloat(jde) || !isFiniteFloat(degree) {
		return math.NaN()
	}
	daysPerDegree := MARS_S_PERIOD / 360
	currentDelta := marsSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJD := jde
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJD := estimateJD
		longitudeDelta := marsSunLongitudeDeltaN(prevJD, degree, true, marsEventSearchN)
		longitudeSlope := (marsSunLongitudeDeltaN(prevJD+0.000005, degree, true, marsEventSearchN) - marsSunLongitudeDeltaN(prevJD-0.000005, degree, true, marsEventSearchN)) / 0.00001
		nextJD := prevJD - longitudeDelta/longitudeSlope
		estimateJD = nextJD
		if math.Abs(nextJD-prevJD) <= marsPhaseCoarseTolerance {
			converged = true
			break
		}
	}
	if !converged {
		return math.NaN()
	}
	converged = false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJD := estimateJD
		longitudeDelta := marsSunLongitudeDelta(prevJD, degree, true)
		longitudeSlope := (marsSunLongitudeDelta(prevJD+0.000005, degree, true) - marsSunLongitudeDelta(prevJD-0.000005, degree, true)) / 0.00001
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
	return TD2UT(estimateJD, false)
}

func LastMarsConjunction(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 0, marsConjunction)
}

func NextMarsConjunction(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 0, marsConjunction)
}

func LastMarsOpposition(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 180, marsConjunction)
}

func NextMarsOpposition(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 180, marsConjunction)
}

func NextMarsEasternQuadrature(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 90, marsConjunction)
}

func LastMarsEasternQuadrature(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 90, marsConjunction)
}

func NextMarsWesternQuadrature(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 270, marsConjunction)
}

func LastMarsWesternQuadrature(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 270, marsConjunction)
}

func marsRetrogradeAroundOpposition(oppositionJD float64, searchBeforeOpposition bool) float64 {
	oppositionTT := TD2UT(oppositionJD, true)
	startTT := oppositionTT
	endTT := oppositionTT
	if searchBeforeOpposition {
		easternQuadratureUT := marsConjunction(oppositionTT, 90, 0)
		startTT = TD2UT(easternQuadratureUT, true)
	} else {
		westernQuadratureUT := marsConjunction(oppositionTT, 270, 1)
		endTT = TD2UT(westernQuadratureUT, true)
	}
	bestJD := zeroEventInWindow(startTT, endTT, marsStationCoarseStepDay, marsStationHalfWindowDay, 30.0/86400.0, func(jd float64) float64 {
		return marsRADerivativeN(jd, marsStationDerivativeStepDay, marsEventSearchN)
	}, func(jd float64) float64 {
		return marsRADerivative(jd, marsStationDerivativeStepDay)
	})
	return TD2UT(bestJD, false)
}

func NextMarsRetrogradeToPrograde(jde float64) float64 {
	lastOppositionJD := marsConjunctionFull(jde, 180, 0)
	date := marsRetrogradeAroundOpposition(lastOppositionJD, false)
	if sameEventUTQueryTT(date, jde) || eventUTQueryAfterOrEqual(date, jde) {
		return date
	}
	nextOppositionJD := marsConjunctionFull(jde, 180, 1)
	return marsRetrogradeAroundOpposition(nextOppositionJD, false)
}

func LastMarsRetrogradeToPrograde(jde float64) float64 {
	lastOppositionJD := marsConjunctionFull(jde, 180, 0)
	date := marsRetrogradeAroundOpposition(lastOppositionJD, false)
	if sameEventUTQueryTT(date, jde) || eventUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	previousOppositionJD := marsConjunctionFull(eventUTLastQueryTT(lastOppositionJD), 180, 0)
	return marsRetrogradeAroundOpposition(previousOppositionJD, false)
}

func NextMarsProgradeToRetrograde(jde float64) float64 {
	nextOppositionJD := marsConjunctionFull(jde, 180, 1)
	date := marsRetrogradeAroundOpposition(nextOppositionJD, true)
	if sameEventUTQueryTT(date, jde) || eventUTQueryAfterOrEqual(date, jde) {
		return date
	}
	followingOppositionJD := marsConjunctionFull(eventUTNextQueryTT(nextOppositionJD), 180, 1)
	return marsRetrogradeAroundOpposition(followingOppositionJD, true)
}

func LastMarsProgradeToRetrograde(jde float64) float64 {
	nextOppositionJD := marsConjunctionFull(jde, 180, 1)
	date := marsRetrogradeAroundOpposition(nextOppositionJD, true)
	if sameEventUTQueryTT(date, jde) || eventUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	lastOppositionJD := marsConjunctionFull(jde, 180, 0)
	return marsRetrogradeAroundOpposition(lastOppositionJD, true)
}
