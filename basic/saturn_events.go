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
	if !isFiniteFloat(jde) || !isFiniteFloat(degree) {
		return math.NaN()
	}
	daysPerDegree := SATURN_S_PERIOD / 360
	currentDelta := saturnSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJD := jde
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJD := estimateJD
		longitudeDelta := saturnSunLongitudeDelta(prevJD, degree, true)
		longitudeSlope := (saturnSunLongitudeDelta(prevJD+0.000005, degree, true) - saturnSunLongitudeDelta(prevJD-0.000005, degree, true)) / 0.00001
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

func saturnConjunction(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	if !isFiniteFloat(jde) || !isFiniteFloat(degree) {
		return math.NaN()
	}
	daysPerDegree := SATURN_S_PERIOD / 360
	currentDelta := saturnSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJD := jde
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJD := estimateJD
		longitudeDelta := saturnSunLongitudeDeltaN(prevJD, degree, true, saturnEventSearchN)
		longitudeSlope := (saturnSunLongitudeDeltaN(prevJD+0.000005, degree, true, saturnEventSearchN) - saturnSunLongitudeDeltaN(prevJD-0.000005, degree, true, saturnEventSearchN)) / 0.00001
		nextJD := prevJD - longitudeDelta/longitudeSlope
		estimateJD = nextJD
		if math.Abs(nextJD-prevJD) <= saturnPhaseCoarseTolerance {
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
		longitudeDelta := saturnSunLongitudeDelta(prevJD, degree, true)
		longitudeSlope := (saturnSunLongitudeDelta(prevJD+0.000005, degree, true) - saturnSunLongitudeDelta(prevJD-0.000005, degree, true)) / 0.00001
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
	if !isFiniteFloat(oppositionJD) {
		return math.NaN()
	}
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
		return saturnRADerivativeN(jd, stationDerivativeStepDay, saturnEventSearchN)
	}, func(jd float64) float64 {
		return saturnRADerivative(jd, stationDerivativeStepDay)
	})
	return TD2UT(bestJD, false)
}

func NextSaturnRetrogradeToPrograde(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	lastOppositionJD := saturnConjunctionFull(jde, 180, 0)
	date := saturnRetrogradeAroundOpposition(lastOppositionJD, false)
	if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
		return date
	}
	nextOppositionJD := saturnConjunctionFull(jde, 180, 1)
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	date = saturnRetrogradeAroundOpposition(nextOppositionJD, false)
	if !isFiniteFloat(date) || !stationUTQueryAfterOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

func LastSaturnRetrogradeToPrograde(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	lastOppositionJD := saturnConjunctionFull(jde, 180, 0)
	date := saturnRetrogradeAroundOpposition(lastOppositionJD, false)
	if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(lastOppositionJD) {
		return math.NaN()
	}
	previousOppositionJD := saturnConjunctionFull(eventUTLastQueryTT(lastOppositionJD), 180, 0)
	if !isFiniteFloat(previousOppositionJD) {
		return math.NaN()
	}
	date = saturnRetrogradeAroundOpposition(previousOppositionJD, false)
	if !isFiniteFloat(date) || !stationUTQueryBeforeOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

func NextSaturnProgradeToRetrograde(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	nextOppositionJD := saturnConjunctionFull(jde, 180, 1)
	date := saturnRetrogradeAroundOpposition(nextOppositionJD, true)
	if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	followingOppositionJD := saturnConjunctionFull(eventUTNextQueryTT(nextOppositionJD), 180, 1)
	if !isFiniteFloat(followingOppositionJD) {
		return math.NaN()
	}
	date = saturnRetrogradeAroundOpposition(followingOppositionJD, true)
	if !isFiniteFloat(date) || !stationUTQueryAfterOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

func LastSaturnProgradeToRetrograde(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	nextOppositionJD := saturnConjunctionFull(jde, 180, 1)
	date := saturnRetrogradeAroundOpposition(nextOppositionJD, true)
	if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	lastOppositionJD := saturnConjunctionFull(jde, 180, 0)
	if !isFiniteFloat(lastOppositionJD) {
		return math.NaN()
	}
	date = saturnRetrogradeAroundOpposition(lastOppositionJD, true)
	if !isFiniteFloat(date) || !stationUTQueryBeforeOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}
