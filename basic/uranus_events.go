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
	if !isFiniteFloat(jde) || !isFiniteFloat(degree) {
		return math.NaN()
	}
	daysPerDegree := URANUS_S_PERIOD / 360
	currentDelta := uranusSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJDE := jde
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJDE := estimateJDE
		longitudeDelta := uranusSunLongitudeDelta(prevJDE, degree, true)
		longitudeSlope := (uranusSunLongitudeDelta(prevJDE+0.000005, degree, true) - uranusSunLongitudeDelta(prevJDE-0.000005, degree, true)) / 0.00001
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

func uranusConjunction(jde, degree float64, next uint8) float64 {
	//0=last 1=next
	if !isFiniteFloat(jde) || !isFiniteFloat(degree) {
		return math.NaN()
	}
	daysPerDegree := URANUS_S_PERIOD / 360
	currentDelta := uranusSunLongitudeDelta(jde, degree, false)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJDE := jde
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJDE := estimateJDE
		longitudeDelta := uranusSunLongitudeDeltaN(prevJDE, degree, true, uranusEventSearchN)
		longitudeSlope := (uranusSunLongitudeDeltaN(prevJDE+0.000005, degree, true, uranusEventSearchN) - uranusSunLongitudeDeltaN(prevJDE-0.000005, degree, true, uranusEventSearchN)) / 0.00001
		nextJD := prevJDE - longitudeDelta/longitudeSlope
		estimateJDE = nextJD
		if math.Abs(nextJD-prevJDE) <= uranusPhaseCoarseTolerance {
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
		longitudeDelta := uranusSunLongitudeDelta(prevJDE, degree, true)
		longitudeSlope := (uranusSunLongitudeDelta(prevJDE+0.000005, degree, true) - uranusSunLongitudeDelta(prevJDE-0.000005, degree, true)) / 0.00001
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
	if !isFiniteFloat(oppositionJD) {
		return math.NaN()
	}
	oppositionTT := UTC2TT(oppositionJD)
	startTT := oppositionTT
	endTT := oppositionTT
	if searchBeforeOpposition {
		easternQuadratureUT := uranusConjunction(oppositionTT, 90, 0)
		startTT = UTC2TT(easternQuadratureUT)
	} else {
		westernQuadratureUT := uranusConjunction(oppositionTT, 270, 1)
		endTT = UTC2TT(westernQuadratureUT)
	}
	bestJDE := zeroEventInWindow(startTT, endTT, 2.0, 2.0, 30.0/86400.0, func(jd float64) float64 {
		return uranusRADerivativeN(jd, stationDerivativeStepDay, uranusEventSearchN)
	}, func(jd float64) float64 {
		return uranusRADerivative(jd, stationDerivativeStepDay)
	})
	return TT2UTC(bestJDE)
}

func NextUranusRetrogradeToPrograde(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	lastOppositionJD := uranusConjunctionFull(jde, 180, 0)
	date := uranusRetrogradeAroundOpposition(lastOppositionJD, false)
	if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
		return date
	}
	nextOppositionJD := uranusConjunctionFull(jde, 180, 1)
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	date = uranusRetrogradeAroundOpposition(nextOppositionJD, false)
	if !isFiniteFloat(date) || !stationUTQueryAfterOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

func LastUranusRetrogradeToPrograde(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	lastOppositionJD := uranusConjunctionFull(jde, 180, 0)
	date := uranusRetrogradeAroundOpposition(lastOppositionJD, false)
	if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(lastOppositionJD) {
		return math.NaN()
	}
	previousOppositionJD := uranusConjunctionFull(eventUTLastQueryTT(lastOppositionJD), 180, 0)
	if !isFiniteFloat(previousOppositionJD) {
		return math.NaN()
	}
	date = uranusRetrogradeAroundOpposition(previousOppositionJD, false)
	if !isFiniteFloat(date) || !stationUTQueryBeforeOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

func NextUranusProgradeToRetrograde(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	nextOppositionJD := uranusConjunctionFull(jde, 180, 1)
	date := uranusRetrogradeAroundOpposition(nextOppositionJD, true)
	if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	followingOppositionJD := uranusConjunctionFull(eventUTNextQueryTT(nextOppositionJD), 180, 1)
	if !isFiniteFloat(followingOppositionJD) {
		return math.NaN()
	}
	date = uranusRetrogradeAroundOpposition(followingOppositionJD, true)
	if !isFiniteFloat(date) || !stationUTQueryAfterOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

func LastUranusProgradeToRetrograde(jde float64) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	nextOppositionJD := uranusConjunctionFull(jde, 180, 1)
	date := uranusRetrogradeAroundOpposition(nextOppositionJD, true)
	if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	lastOppositionJD := uranusConjunctionFull(jde, 180, 0)
	if !isFiniteFloat(lastOppositionJD) {
		return math.NaN()
	}
	date = uranusRetrogradeAroundOpposition(lastOppositionJD, true)
	if !isFiniteFloat(date) || !stationUTQueryBeforeOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}
