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

func jupiterSunLongitudeDelta(jde, degree float64, filter bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := Limit360(Limit360(JupiterApparentLoWithInstance(jde, inst)-HSunApparentLoWithInstance(jde, inst)) - degree)
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

func jupiterSunLongitudeDeltaN(jde, degree float64, filter bool, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := Limit360(Limit360(JupiterApparentLoNWithInstance(jde, n, inst)-HSunApparentLoNWithInstance(jde, n, inst)) - degree)
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

func jupiterRADerivative(jde, delta float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := JupiterApparentRaWithInstance(jde+delta, inst) - JupiterApparentRaWithInstance(jde-delta, inst)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func jupiterRADerivativeN(jde, delta float64, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := JupiterApparentRaNWithInstance(jde+delta, n, inst) - JupiterApparentRaNWithInstance(jde-delta, n, inst)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func jupiterConjunctionFull(jde, degree float64, next uint8, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	//0=last 1=next
	if !isFiniteFloat(jde) || !isFiniteFloat(degree) {
		return math.NaN()
	}
	daysPerDegree := JUPITER_S_PERIOD / 360
	currentDelta := jupiterSunLongitudeDelta(jde, degree, false, inst)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJDE := jde
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJDE := estimateJDE
		longitudeDelta := jupiterSunLongitudeDelta(prevJDE, degree, true, inst)
		longitudeSlope := (jupiterSunLongitudeDelta(prevJDE+0.000005, degree, true, inst) - jupiterSunLongitudeDelta(prevJDE-0.000005, degree, true, inst)) / 0.00001
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

func jupiterConjunction(jde, degree float64, next uint8, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	//0=last 1=next
	if !isFiniteFloat(jde) || !isFiniteFloat(degree) {
		return math.NaN()
	}
	daysPerDegree := JUPITER_S_PERIOD / 360
	currentDelta := jupiterSunLongitudeDelta(jde, degree, false, inst)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJDE := jde
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJDE := estimateJDE
		longitudeDelta := jupiterSunLongitudeDeltaN(prevJDE, degree, true, jupiterEventSearchN, inst)
		longitudeSlope := (jupiterSunLongitudeDeltaN(prevJDE+0.000005, degree, true, jupiterEventSearchN, inst) - jupiterSunLongitudeDeltaN(prevJDE-0.000005, degree, true, jupiterEventSearchN, inst)) / 0.00001
		nextJD := prevJDE - longitudeDelta/longitudeSlope
		estimateJDE = nextJD
		if math.Abs(nextJD-prevJDE) <= jupiterPhaseCoarseTolerance {
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
		longitudeDelta := jupiterSunLongitudeDelta(prevJDE, degree, true, inst)
		longitudeSlope := (jupiterSunLongitudeDelta(prevJDE+0.000005, degree, true, inst) - jupiterSunLongitudeDelta(prevJDE-0.000005, degree, true, inst)) / 0.00001
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

// LastJupiterConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastJupiterConjunctionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 0, func(jde, degree float64, next uint8) float64 {
		return jupiterConjunction(jde, degree, next, inst)
	})
}

// NextJupiterConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextJupiterConjunctionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 0, func(jde, degree float64, next uint8) float64 {
		return jupiterConjunction(jde, degree, next, inst)
	})
}

// LastJupiterOppositionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastJupiterOppositionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 180, func(jde, degree float64, next uint8) float64 {
		return jupiterConjunction(jde, degree, next, inst)
	})
}

// NextJupiterOppositionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextJupiterOppositionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 180, func(jde, degree float64, next uint8) float64 {
		return jupiterConjunction(jde, degree, next, inst)
	})
}

// NextJupiterEasternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextJupiterEasternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 90, func(jde, degree float64, next uint8) float64 {
		return jupiterConjunction(jde, degree, next, inst)
	})
}

// LastJupiterEasternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastJupiterEasternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 90, func(jde, degree float64, next uint8) float64 {
		return jupiterConjunction(jde, degree, next, inst)
	})
}

// NextJupiterWesternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextJupiterWesternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 270, func(jde, degree float64, next uint8) float64 {
		return jupiterConjunction(jde, degree, next, inst)
	})
}

// LastJupiterWesternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastJupiterWesternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 270, func(jde, degree float64, next uint8) float64 {
		return jupiterConjunction(jde, degree, next, inst)
	})
}

func jupiterRetrogradeAroundOpposition(oppositionJD float64, searchBeforeOpposition bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	if !isFiniteFloat(oppositionJD) {
		return math.NaN()
	}
	oppositionTT := UTC2TT(oppositionJD)
	startTT := oppositionTT
	endTT := oppositionTT
	if searchBeforeOpposition {
		easternQuadratureUT := jupiterConjunction(oppositionTT, 90, 0, inst)
		startTT = UTC2TT(easternQuadratureUT)
	} else {
		westernQuadratureUT := jupiterConjunction(oppositionTT, 270, 1, inst)
		endTT = UTC2TT(westernQuadratureUT)
	}
	bestJDE := zeroEventInWindow(startTT, endTT, 2.0, 2.0, 30.0/86400.0, func(jd float64) float64 {
		return jupiterRADerivativeN(jd, stationDerivativeStepDay, jupiterEventSearchN, inst)
	}, func(jd float64) float64 {
		return jupiterRADerivative(jd, stationDerivativeStepDay, inst)
	})
	return TT2UTC(bestJDE)
}

// NextJupiterRetrogradeToProgradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextJupiterRetrogradeToProgradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	lastOppositionJD := jupiterConjunctionFull(jde, 180, 0, inst)
	date := jupiterRetrogradeAroundOpposition(lastOppositionJD, false, inst)
	if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
		return date
	}
	nextOppositionJD := jupiterConjunctionFull(jde, 180, 1, inst)
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	date = jupiterRetrogradeAroundOpposition(nextOppositionJD, false, inst)
	if !isFiniteFloat(date) || !stationUTQueryAfterOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

// LastJupiterRetrogradeToProgradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastJupiterRetrogradeToProgradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	lastOppositionJD := jupiterConjunctionFull(jde, 180, 0, inst)
	date := jupiterRetrogradeAroundOpposition(lastOppositionJD, false, inst)
	if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(lastOppositionJD) {
		return math.NaN()
	}
	previousOppositionJD := jupiterConjunctionFull(eventUTLastQueryTT(lastOppositionJD), 180, 0, inst)
	if !isFiniteFloat(previousOppositionJD) {
		return math.NaN()
	}
	date = jupiterRetrogradeAroundOpposition(previousOppositionJD, false, inst)
	if !isFiniteFloat(date) || !stationUTQueryBeforeOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

// NextJupiterProgradeToRetrogradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextJupiterProgradeToRetrogradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	nextOppositionJD := jupiterConjunctionFull(jde, 180, 1, inst)
	date := jupiterRetrogradeAroundOpposition(nextOppositionJD, true, inst)
	if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	followingOppositionJD := jupiterConjunctionFull(eventUTNextQueryTT(nextOppositionJD), 180, 1, inst)
	if !isFiniteFloat(followingOppositionJD) {
		return math.NaN()
	}
	date = jupiterRetrogradeAroundOpposition(followingOppositionJD, true, inst)
	if !isFiniteFloat(date) || !stationUTQueryAfterOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

// LastJupiterProgradeToRetrogradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastJupiterProgradeToRetrogradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	nextOppositionJD := jupiterConjunctionFull(jde, 180, 1, inst)
	date := jupiterRetrogradeAroundOpposition(nextOppositionJD, true, inst)
	if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	lastOppositionJD := jupiterConjunctionFull(jde, 180, 0, inst)
	if !isFiniteFloat(lastOppositionJD) {
		return math.NaN()
	}
	date = jupiterRetrogradeAroundOpposition(lastOppositionJD, true, inst)
	if !isFiniteFloat(date) || !stationUTQueryBeforeOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}
