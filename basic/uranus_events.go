package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

// Pos

const (
	URANUS_S_PERIOD            = 1 / ((1 / 365.256363004) - (1 / 30799.095))
	uranusEventSearchN         = 16
	uranusPhaseCoarseTolerance = 30.0 / 86400.0
)

func uranusSunLongitudeDelta(jde, degree float64, filter bool) float64 {
	sub := Limit360(Limit360(UranusApparentLo(jde)-HSunApparentLo(jde)) - degree)
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

func uranusSunLongitudeDeltaN(jde, degree float64, filter bool, n int) float64 {
	sub := Limit360(Limit360(UranusApparentLoN(jde, n)-HSunApparentLoN(jde, n)) - degree)
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

func uranusRADerivative(jde, delta float64) float64 {
	sub := UranusApparentRa(jde+delta) - UranusApparentRa(jde-delta)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func uranusRADerivativeN(jde, delta float64, n int) float64 {
	sub := UranusApparentRaN(jde+delta, n) - UranusApparentRaN(jde-delta, n)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func uranusConjunctionFull(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	daysPerDegree := URANUS_S_PERIOD / 360
	currentDelta := uranusSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		longitudeDelta := uranusSunLongitudeDelta(prevJD, degree, true)
		longitudeSlope := (uranusSunLongitudeDelta(prevJD+0.000005, degree, true) - uranusSunLongitudeDelta(prevJD-0.000005, degree, true)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return TD2UT(estimateJD, false)
}

func uranusConjunction(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	daysPerDegree := URANUS_S_PERIOD / 360
	currentDelta := uranusSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJD := jde
	for {
		prevJD := estimateJD
		longitudeDelta := uranusSunLongitudeDeltaN(prevJD, degree, true, uranusEventSearchN)
		longitudeSlope := (uranusSunLongitudeDeltaN(prevJD+0.000005, degree, true, uranusEventSearchN) - uranusSunLongitudeDeltaN(prevJD-0.000005, degree, true, uranusEventSearchN)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= uranusPhaseCoarseTolerance {
			break
		}
	}
	for {
		prevJD := estimateJD
		longitudeDelta := uranusSunLongitudeDelta(prevJD, degree, true)
		longitudeSlope := (uranusSunLongitudeDelta(prevJD+0.000005, degree, true) - uranusSunLongitudeDelta(prevJD-0.000005, degree, true)) / 0.00001
		estimateJD = prevJD - longitudeDelta/longitudeSlope
		if math.Abs(estimateJD-prevJD) <= 0.00001 {
			break
		}
	}
	return TD2UT(estimateJD, false)
}

func LastUranusConjunction(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 0, uranusConjunction)
}

func NextUranusConjunction(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 0, uranusConjunction)
}

func LastUranusOpposition(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 180, uranusConjunction)
}

func NextUranusOpposition(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 180, uranusConjunction)
}

func NextUranusEasternQuadrature(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 90, uranusConjunction)
}

func LastUranusEasternQuadrature(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 90, uranusConjunction)
}

func NextUranusWesternQuadrature(jde float64) float64 {
	return inclusiveNextPhaseEvent(jde, 270, uranusConjunction)
}

func LastUranusWesternQuadrature(jde float64) float64 {
	return inclusiveLastPhaseEvent(jde, 270, uranusConjunction)
}

func uranusRetrogradeAroundOpposition(oppositionJD float64, searchBeforeOpposition bool) float64 {
	oppositionTT := TD2UT(oppositionJD, true)
	startTT := oppositionTT
	endTT := oppositionTT
	if searchBeforeOpposition {
		easternQuadratureUT := uranusConjunction(oppositionTT, 90, 0)
		startTT = TD2UT(easternQuadratureUT, true)
	} else {
		westernQuadratureUT := uranusConjunction(oppositionTT, 270, 1)
		endTT = TD2UT(westernQuadratureUT, true)
	}
	bestJD := zeroEventInWindow(startTT, endTT, 2.0, 2.0, 30.0/86400.0, func(jd float64) float64 {
		return uranusRADerivativeN(jd, 1.0/86400.0, uranusEventSearchN)
	}, func(jd float64) float64 {
		return uranusRADerivative(jd, 0.5/86400.0)
	})
	return TD2UT(bestJD, false)
}

func NextUranusRetrogradeToPrograde(jde float64) float64 {
	lastOppositionJD := uranusConjunctionFull(jde, 180, 0)
	date := uranusRetrogradeAroundOpposition(lastOppositionJD, false)
	if sameEventUTQueryTT(date, jde) || eventUTQueryAfterOrEqual(date, jde) {
		return date
	}
	nextOppositionJD := uranusConjunctionFull(jde, 180, 1)
	return uranusRetrogradeAroundOpposition(nextOppositionJD, false)
}

func LastUranusRetrogradeToPrograde(jde float64) float64 {
	lastOppositionJD := uranusConjunctionFull(jde, 180, 0)
	date := uranusRetrogradeAroundOpposition(lastOppositionJD, false)
	if sameEventUTQueryTT(date, jde) || eventUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	previousOppositionJD := uranusConjunctionFull(eventUTLastQueryTT(lastOppositionJD), 180, 0)
	return uranusRetrogradeAroundOpposition(previousOppositionJD, false)
}

func NextUranusProgradeToRetrograde(jde float64) float64 {
	nextOppositionJD := uranusConjunctionFull(jde, 180, 1)
	date := uranusRetrogradeAroundOpposition(nextOppositionJD, true)
	if sameEventUTQueryTT(date, jde) || eventUTQueryAfterOrEqual(date, jde) {
		return date
	}
	followingOppositionJD := uranusConjunctionFull(eventUTNextQueryTT(nextOppositionJD), 180, 1)
	return uranusRetrogradeAroundOpposition(followingOppositionJD, true)
}

func LastUranusProgradeToRetrograde(jde float64) float64 {
	nextOppositionJD := uranusConjunctionFull(jde, 180, 1)
	date := uranusRetrogradeAroundOpposition(nextOppositionJD, true)
	if sameEventUTQueryTT(date, jde) || eventUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	lastOppositionJD := uranusConjunctionFull(jde, 180, 0)
	return uranusRetrogradeAroundOpposition(lastOppositionJD, true)
}
