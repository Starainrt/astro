package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

// Pos

const (
	NEPTUNE_S_PERIOD            = 1 / ((1 / 365.256363004) - (1 / 60190.03))
	neptuneEventSearchN         = 16
	neptunePhaseCoarseTolerance = 30.0 / 86400.0
)

func neptuneSunLongitudeDelta(jde, degree float64, filter bool) float64 {
	sub := Limit360(Limit360(NeptuneApparentLo(jde)-HSunApparentLo(jde)) - degree)
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

func neptuneSunLongitudeDeltaN(jde, degree float64, filter bool, n int) float64 {
	sub := Limit360(Limit360(NeptuneApparentLoN(jde, n)-HSunApparentLoN(jde, n)) - degree)
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

func neptuneRADerivative(jde, delta float64) float64 {
	sub := NeptuneApparentRa(jde+delta) - NeptuneApparentRa(jde-delta)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func neptuneRADerivativeN(jde, delta float64, n int) float64 {
	sub := NeptuneApparentRaN(jde+delta, n) - NeptuneApparentRaN(jde-delta, n)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func neptuneConjunctionFull(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	daysPerDegree := NEPTUNE_S_PERIOD / 360
	currentDelta := neptuneSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		longitudeDelta := neptuneSunLongitudeDelta(prevJD, degree, true)
		longitudeSlope := (neptuneSunLongitudeDelta(prevJD+0.000005, degree, true) - neptuneSunLongitudeDelta(prevJD-0.000005, degree, true)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return TD2UT(estimateJD, false)
}

func neptuneConjunction(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	daysPerDegree := NEPTUNE_S_PERIOD / 360
	currentDelta := neptuneSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		longitudeDelta := neptuneSunLongitudeDeltaN(prevJD, degree, true, neptuneEventSearchN)
		longitudeSlope := (neptuneSunLongitudeDeltaN(prevJD+0.000005, degree, true, neptuneEventSearchN) - neptuneSunLongitudeDeltaN(prevJD-0.000005, degree, true, neptuneEventSearchN)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= neptunePhaseCoarseTolerance {
			break
		}
	}
	for {
		prevJD := estimateJD
		longitudeDelta := neptuneSunLongitudeDelta(prevJD, degree, true)
		longitudeSlope := (neptuneSunLongitudeDelta(prevJD+0.000005, degree, true) - neptuneSunLongitudeDelta(prevJD-0.000005, degree, true)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return TD2UT(estimateJD, false)
}

func LastNeptuneConjunction(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 0, neptuneConjunction)
}

func NextNeptuneConjunction(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 0, neptuneConjunction)
}

func LastNeptuneOpposition(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 180, neptuneConjunction)
}

func NextNeptuneOpposition(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 180, neptuneConjunction)
}

func NextNeptuneEasternQuadrature(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 90, neptuneConjunction)
}

func LastNeptuneEasternQuadrature(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 90, neptuneConjunction)
}

func NextNeptuneWesternQuadrature(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 270, neptuneConjunction)
}

func LastNeptuneWesternQuadrature(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 270, neptuneConjunction)
}

func neptuneRetrogradeAroundOpposition(oppositionJD float64, searchBeforeOpposition bool) float64 {
	oppositionTT := TD2UT(oppositionJD, true)
	startTT := oppositionTT
	endTT := oppositionTT
	if searchBeforeOpposition {
		easternQuadratureUT := neptuneConjunction(oppositionTT, 90, 0)
		startTT = TD2UT(easternQuadratureUT, true)
	} else {
		westernQuadratureUT := neptuneConjunction(oppositionTT, 270, 1)
		endTT = TD2UT(westernQuadratureUT, true)
	}
	bestJD := zeroEventInWindow(startTT, endTT, 2.0, 2.0, 30.0/86400.0, func(jd float64) float64 {
		return neptuneRADerivativeN(jd, 1.0/86400.0, neptuneEventSearchN)
	}, func(jd float64) float64 {
		return neptuneRADerivative(jd, 0.5/86400.0)
	})
	return TD2UT(bestJD, false)
}

func NextNeptuneRetrogradeToPrograde(jde float64) float64 {
	lastOppositionJD := neptuneConjunctionFull(jde, 180, 0)
	date := neptuneRetrogradeAroundOpposition(lastOppositionJD, false)
	if sameEventUTQueryTT(date, jde) || eventUTQueryAfterOrEqual(date, jde) {
		return date
	}
	nextOppositionJD := neptuneConjunctionFull(jde, 180, 1)
	return neptuneRetrogradeAroundOpposition(nextOppositionJD, false)
}

func LastNeptuneRetrogradeToPrograde(jde float64) float64 {
	lastOppositionJD := neptuneConjunctionFull(jde, 180, 0)
	date := neptuneRetrogradeAroundOpposition(lastOppositionJD, false)
	if sameEventUTQueryTT(date, jde) || eventUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	previousOppositionJD := neptuneConjunctionFull(eventUTLastQueryTT(lastOppositionJD), 180, 0)
	return neptuneRetrogradeAroundOpposition(previousOppositionJD, false)
}

func NextNeptuneProgradeToRetrograde(jde float64) float64 {
	nextOppositionJD := neptuneConjunctionFull(jde, 180, 1)
	date := neptuneRetrogradeAroundOpposition(nextOppositionJD, true)
	if sameEventUTQueryTT(date, jde) || eventUTQueryAfterOrEqual(date, jde) {
		return date
	}
	followingOppositionJD := neptuneConjunctionFull(eventUTNextQueryTT(nextOppositionJD), 180, 1)
	return neptuneRetrogradeAroundOpposition(followingOppositionJD, true)
}

func LastNeptuneProgradeToRetrograde(jde float64) float64 {
	nextOppositionJD := neptuneConjunctionFull(jde, 180, 1)
	date := neptuneRetrogradeAroundOpposition(nextOppositionJD, true)
	if sameEventUTQueryTT(date, jde) || eventUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	lastOppositionJD := neptuneConjunctionFull(jde, 180, 0)
	return neptuneRetrogradeAroundOpposition(lastOppositionJD, true)
}
