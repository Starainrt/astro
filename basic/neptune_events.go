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
	if !isFiniteFloat(jde) || !isFiniteFloat(degree) {
		return math.NaN()
	}
	daysPerDegree := NEPTUNE_S_PERIOD / 360
	currentDelta := neptuneSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJDE := jde
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJDE := estimateJDE
		longitudeDelta := neptuneSunLongitudeDelta(prevJDE, degree, true)
		longitudeSlope := (neptuneSunLongitudeDelta(prevJDE+0.000005, degree, true) - neptuneSunLongitudeDelta(prevJDE-0.000005, degree, true)) / 0.00001
		nextJD := prevJDE - longitudeDelta/longitudeSlope
		estimateJDE = nextJD
		if math.Abs(nextJD-prevJDE) <= 0.00001 {
			converged = true
			break
		}
	}
	if !converged {
		return math.NaN()
	}
	return TT2UTC(estimateJDE)
}

func neptuneConjunction(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	if !isFiniteFloat(jde) || !isFiniteFloat(degree) {
		return math.NaN()
	}
	daysPerDegree := NEPTUNE_S_PERIOD / 360
	currentDelta := neptuneSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJDE := jde
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJDE := estimateJDE
		longitudeDelta := neptuneSunLongitudeDeltaN(prevJDE, degree, true, neptuneEventSearchN)
		longitudeSlope := (neptuneSunLongitudeDeltaN(prevJDE+0.000005, degree, true, neptuneEventSearchN) - neptuneSunLongitudeDeltaN(prevJDE-0.000005, degree, true, neptuneEventSearchN)) / 0.00001
		nextJD := prevJDE - longitudeDelta/longitudeSlope
		estimateJDE = nextJD
		if math.Abs(nextJD-prevJDE) <= neptunePhaseCoarseTolerance {
			converged = true
			break
		}
	}
	if !converged {
		return math.NaN()
	}
	converged = false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJDE := estimateJDE
		longitudeDelta := neptuneSunLongitudeDelta(prevJDE, degree, true)
		longitudeSlope := (neptuneSunLongitudeDelta(prevJDE+0.000005, degree, true) - neptuneSunLongitudeDelta(prevJDE-0.000005, degree, true)) / 0.00001
		nextJD := prevJDE - longitudeDelta/longitudeSlope
		estimateJDE = nextJD
		if math.Abs(nextJD-prevJDE) <= 0.00001 {
			converged = true
			break
		}
	}
	if !converged {
		return math.NaN()
	}
	return TT2UTC(estimateJDE)
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
	if !isFiniteFloat(oppositionJD) {
		return math.NaN()
	}
	oppositionTT := UTC2TT(oppositionJD)
	startTT := oppositionTT
	endTT := oppositionTT
	if searchBeforeOpposition {
		easternQuadratureUT := neptuneConjunction(oppositionTT, 90, 0)
		startTT = UTC2TT(easternQuadratureUT)
	} else {
		westernQuadratureUT := neptuneConjunction(oppositionTT, 270, 1)
		endTT = UTC2TT(westernQuadratureUT)
	}
	bestJDE := zeroEventInWindow(startTT, endTT, 2.0, 2.0, 30.0/86400.0, func(jd float64) float64 {
		return neptuneRADerivativeN(jd, stationDerivativeStepDay, neptuneEventSearchN)
	}, func(jd float64) float64 {
		return neptuneRADerivative(jd, stationDerivativeStepDay)
	})
	return TT2UTC(bestJDE)
}

func NextNeptuneRetrogradeToPrograde(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	lastOppositionJD := neptuneConjunctionFull(jde, 180, 0)
	date := neptuneRetrogradeAroundOpposition(lastOppositionJD, false)
	if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
		return date
	}
	nextOppositionJD := neptuneConjunctionFull(jde, 180, 1)
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	date = neptuneRetrogradeAroundOpposition(nextOppositionJD, false)
	if !isFiniteFloat(date) || !stationUTQueryAfterOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

func LastNeptuneRetrogradeToPrograde(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	lastOppositionJD := neptuneConjunctionFull(jde, 180, 0)
	date := neptuneRetrogradeAroundOpposition(lastOppositionJD, false)
	if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(lastOppositionJD) {
		return math.NaN()
	}
	previousOppositionJD := neptuneConjunctionFull(eventUTLastQueryTT(lastOppositionJD), 180, 0)
	if !isFiniteFloat(previousOppositionJD) {
		return math.NaN()
	}
	date = neptuneRetrogradeAroundOpposition(previousOppositionJD, false)
	if !isFiniteFloat(date) || !stationUTQueryBeforeOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

func NextNeptuneProgradeToRetrograde(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	nextOppositionJD := neptuneConjunctionFull(jde, 180, 1)
	date := neptuneRetrogradeAroundOpposition(nextOppositionJD, true)
	if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	followingOppositionJD := neptuneConjunctionFull(eventUTNextQueryTT(nextOppositionJD), 180, 1)
	if !isFiniteFloat(followingOppositionJD) {
		return math.NaN()
	}
	date = neptuneRetrogradeAroundOpposition(followingOppositionJD, true)
	if !isFiniteFloat(date) || !stationUTQueryAfterOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

func LastNeptuneProgradeToRetrograde(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	nextOppositionJD := neptuneConjunctionFull(jde, 180, 1)
	date := neptuneRetrogradeAroundOpposition(nextOppositionJD, true)
	if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	lastOppositionJD := neptuneConjunctionFull(jde, 180, 0)
	if !isFiniteFloat(lastOppositionJD) {
		return math.NaN()
	}
	date = neptuneRetrogradeAroundOpposition(lastOppositionJD, true)
	if !isFiniteFloat(date) || !stationUTQueryBeforeOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}
