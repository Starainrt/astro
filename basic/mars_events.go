package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

// Pos

const (
	MARS_S_PERIOD            = 1 / ((1 / 365.256363004) - (1 / 686.98))
	marsEventSearchN         = 16
	marsPhaseCoarseTolerance = 30.0 / 86400.0
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
	daysPerDegree := MARS_S_PERIOD / 360
	currentDelta := marsSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		longitudeDelta := marsSunLongitudeDelta(prevJD, degree, true)
		longitudeSlope := (marsSunLongitudeDelta(prevJD+0.000005, degree, true) - marsSunLongitudeDelta(prevJD-0.000005, degree, true)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return TD2UT(estimateJD, false)
}

func marsConjunction(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	daysPerDegree := MARS_S_PERIOD / 360
	currentDelta := marsSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		longitudeDelta := marsSunLongitudeDeltaN(prevJD, degree, true, marsEventSearchN)
		longitudeSlope := (marsSunLongitudeDeltaN(prevJD+0.000005, degree, true, marsEventSearchN) - marsSunLongitudeDeltaN(prevJD-0.000005, degree, true, marsEventSearchN)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= marsPhaseCoarseTolerance {
			break
		}
	}
	for {
		prevJD := estimateJD
		longitudeDelta := marsSunLongitudeDelta(prevJD, degree, true)
		longitudeSlope := (marsSunLongitudeDelta(prevJD+0.000005, degree, true) - marsSunLongitudeDelta(prevJD-0.000005, degree, true)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
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
	jde := oppositionJD
	if searchBeforeOpposition {
		jde -= 60
	} else {
		jde += 60
	}
	for {
		currentRate := marsRADerivative(jde, 1.0/86400.0)
		if math.Abs(currentRate) > 0.55 {
			jde += 2
			continue
		}
		break
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		rateValue := marsRADerivative(prevJD, 2.0/86400.0)
		rateSlope := (marsRADerivative(prevJD+15.0/86400.0, 2.0/86400.0) - marsRADerivative(prevJD-15.0/86400.0, 2.0/86400.0)) / (30.0 / 86400.0)
		estimateJD = prevJD - rateValue/rateSlope
		if math.Abs(estimateJD-prevJD) <= 30.0/86400.0 {
			break
		}
	}
	bestJD := eventZeroRefine(estimateJD, 15.0/86400.0, 0.5/86400.0, func(jd float64) float64 {
		return marsRADerivative(jd, 0.5/86400.0)
	})
	return TD2UT(bestJD, false)
}

func marsOppositionFromBefore(oppositionJD float64) float64 {
	return marsConjunctionFull(eventUTLastQueryTT(oppositionJD), 180, 1)
}

func marsOppositionFromAfter(oppositionJD float64) float64 {
	return marsConjunctionFull(eventUTNextQueryTT(oppositionJD), 180, 0)
}

func NextMarsRetrogradeToPrograde(jde float64) float64 {
	lastOppositionJD := marsConjunctionFull(jde, 180, 0)
	date := marsRetrogradeAroundOpposition(lastOppositionJD, false)
	if sameEventUTQueryTT(date, jde) {
		sameOppositionJD := marsOppositionFromBefore(lastOppositionJD)
		return closestEventUTToQueryTT(jde, date, marsRetrogradeAroundOpposition(sameOppositionJD, false))
	}
	if !eventUTQueryAfterOrEqual(date, jde) {
		nextOppositionJD := marsConjunctionFull(jde, 180, 1)
		return marsRetrogradeAroundOpposition(nextOppositionJD, false)
	}
	return date
}

func LastMarsRetrogradeToPrograde(jde float64) float64 {
	lastOppositionJD := marsConjunctionFull(jde, 180, 0)
	date := marsRetrogradeAroundOpposition(lastOppositionJD, false)
	if sameEventUTQueryTT(date, jde) {
		sameOppositionJD := marsOppositionFromBefore(lastOppositionJD)
		return closestEventUTToQueryTT(jde, date, marsRetrogradeAroundOpposition(sameOppositionJD, false))
	}
	if !eventUTQueryBeforeOrEqual(date, jde) {
		previousOppositionJD := marsConjunctionFull(eventUTLastQueryTT(lastOppositionJD), 180, 0)
		return marsRetrogradeAroundOpposition(previousOppositionJD, false)
	}
	return date
}

func NextMarsProgradeToRetrograde(jde float64) float64 {
	nextOppositionJD := marsConjunctionFull(jde, 180, 1)
	date := marsRetrogradeAroundOpposition(nextOppositionJD, true)
	if sameEventUTQueryTT(date, jde) {
		sameOppositionJD := marsOppositionFromAfter(nextOppositionJD)
		return closestEventUTToQueryTT(jde, date, marsRetrogradeAroundOpposition(sameOppositionJD, true))
	}
	if !eventUTQueryAfterOrEqual(date, jde) {
		followingOppositionJD := marsConjunctionFull(eventUTNextQueryTT(nextOppositionJD), 180, 1)
		return marsRetrogradeAroundOpposition(followingOppositionJD, true)
	}
	return date
}

func LastMarsProgradeToRetrograde(jde float64) float64 {
	nextOppositionJD := marsConjunctionFull(jde, 180, 1)
	date := marsRetrogradeAroundOpposition(nextOppositionJD, true)
	if sameEventUTQueryTT(date, jde) {
		sameOppositionJD := marsOppositionFromAfter(nextOppositionJD)
		return closestEventUTToQueryTT(jde, date, marsRetrogradeAroundOpposition(sameOppositionJD, true))
	}
	if !eventUTQueryBeforeOrEqual(date, jde) {
		lastOppositionJD := marsConjunctionFull(jde, 180, 0)
		return marsRetrogradeAroundOpposition(lastOppositionJD, true)
	}
	return date
}
