package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

// Pos

const (
	SATURN_S_PERIOD            = 1 / ((1 / 365.256363004) - (1 / 10759.0))
	saturnEventSearchN         = 16
	saturnPhaseCoarseTolerance = 30.0 / 86400.0
)

func saturnSunLongitudeDelta(jde, degree float64, filter bool) float64 {
	sub := Limit360(Limit360(SaturnApparentLo(jde)-HSunApparentLo(jde)) - degree)
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

func saturnSunLongitudeDeltaN(jde, degree float64, filter bool, n int) float64 {
	sub := Limit360(Limit360(SaturnApparentLoN(jde, n)-HSunApparentLoN(jde, n)) - degree)
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

func saturnRADerivative(jde, delta float64) float64 {
	sub := SaturnApparentRa(jde+delta) - SaturnApparentRa(jde-delta)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func saturnRADerivativeN(jde, delta float64, n int) float64 {
	sub := SaturnApparentRaN(jde+delta, n) - SaturnApparentRaN(jde-delta, n)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func saturnConjunctionFull(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	daysPerDegree := SATURN_S_PERIOD / 360
	currentDelta := saturnSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		longitudeDelta := saturnSunLongitudeDelta(prevJD, degree, true)
		longitudeSlope := (saturnSunLongitudeDelta(prevJD+0.000005, degree, true) - saturnSunLongitudeDelta(prevJD-0.000005, degree, true)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return TD2UT(estimateJD, false)
}

func saturnConjunction(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	daysPerDegree := SATURN_S_PERIOD / 360
	currentDelta := saturnSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		longitudeDelta := saturnSunLongitudeDeltaN(prevJD, degree, true, saturnEventSearchN)
		longitudeSlope := (saturnSunLongitudeDeltaN(prevJD+0.000005, degree, true, saturnEventSearchN) - saturnSunLongitudeDeltaN(prevJD-0.000005, degree, true, saturnEventSearchN)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= saturnPhaseCoarseTolerance {
			break
		}
	}
	for {
		prevJD := estimateJD
		longitudeDelta := saturnSunLongitudeDelta(prevJD, degree, true)
		longitudeSlope := (saturnSunLongitudeDelta(prevJD+0.000005, degree, true) - saturnSunLongitudeDelta(prevJD-0.000005, degree, true)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return TD2UT(estimateJD, false)
}

func LastSaturnConjunction(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 0, saturnConjunction)
}

func NextSaturnConjunction(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 0, saturnConjunction)
}

func LastSaturnOpposition(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 180, saturnConjunction)
}

func NextSaturnOpposition(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 180, saturnConjunction)
}

func NextSaturnEasternQuadrature(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 90, saturnConjunction)
}

func LastSaturnEasternQuadrature(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 90, saturnConjunction)
}

func NextSaturnWesternQuadrature(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 270, saturnConjunction)
}

func LastSaturnWesternQuadrature(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 270, saturnConjunction)
}

func saturnRetrogradeAroundOpposition(oppositionJD float64, searchBeforeOpposition bool) float64 {
	oppositionTT := TD2UT(oppositionJD, true)
	startTT := oppositionTT
	endTT := oppositionTT
	if searchBeforeOpposition {
		easternQuadratureUT := saturnConjunction(oppositionTT, 90, 0)
		startTT = TD2UT(easternQuadratureUT, true)
	} else {
		westernQuadratureUT := saturnConjunction(oppositionTT, 270, 1)
		endTT = TD2UT(westernQuadratureUT, true)
	}
	bestJD := zeroEventInWindow(startTT, endTT, 2.0, 2.0, 30.0/86400.0, func(jd float64) float64 {
		return saturnRADerivativeN(jd, 1.0/86400.0, saturnEventSearchN)
	}, func(jd float64) float64 {
		return saturnRADerivative(jd, 0.5/86400.0)
	})
	return TD2UT(bestJD, false)
}

func NextSaturnRetrogradeToPrograde(jde float64) float64 {
	lastOppositionJD := saturnConjunctionFull(jde, 180, 0)
	date := saturnRetrogradeAroundOpposition(lastOppositionJD, false)
	if sameEventUTQueryTT(date, jde) || eventUTQueryAfterOrEqual(date, jde) {
		return date
	}
	nextOppositionJD := saturnConjunctionFull(jde, 180, 1)
	return saturnRetrogradeAroundOpposition(nextOppositionJD, false)
}

func LastSaturnRetrogradeToPrograde(jde float64) float64 {
	lastOppositionJD := saturnConjunctionFull(jde, 180, 0)
	date := saturnRetrogradeAroundOpposition(lastOppositionJD, false)
	if sameEventUTQueryTT(date, jde) || eventUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	previousOppositionJD := saturnConjunctionFull(eventUTLastQueryTT(lastOppositionJD), 180, 0)
	return saturnRetrogradeAroundOpposition(previousOppositionJD, false)
}

func NextSaturnProgradeToRetrograde(jde float64) float64 {
	nextOppositionJD := saturnConjunctionFull(jde, 180, 1)
	date := saturnRetrogradeAroundOpposition(nextOppositionJD, true)
	if sameEventUTQueryTT(date, jde) || eventUTQueryAfterOrEqual(date, jde) {
		return date
	}
	followingOppositionJD := saturnConjunctionFull(eventUTNextQueryTT(nextOppositionJD), 180, 1)
	return saturnRetrogradeAroundOpposition(followingOppositionJD, true)
}

func LastSaturnProgradeToRetrograde(jde float64) float64 {
	nextOppositionJD := saturnConjunctionFull(jde, 180, 1)
	date := saturnRetrogradeAroundOpposition(nextOppositionJD, true)
	if sameEventUTQueryTT(date, jde) || eventUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	lastOppositionJD := saturnConjunctionFull(jde, 180, 0)
	return saturnRetrogradeAroundOpposition(lastOppositionJD, true)
}
