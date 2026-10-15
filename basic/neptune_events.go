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

func neptuneSunLongitudeDelta(jde, degree float64, filter bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := Limit360(Limit360(NeptuneApparentLoWithInstance(jde, inst)-HSunApparentLoWithInstance(jde, inst)) - degree)
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

func neptuneSunLongitudeDeltaN(jde, degree float64, filter bool, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := Limit360(Limit360(NeptuneApparentLoNWithInstance(jde, n, inst)-HSunApparentLoNWithInstance(jde, n, inst)) - degree)
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

func neptuneRADerivative(jde, delta float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := NeptuneApparentRaWithInstance(jde+delta, inst) - NeptuneApparentRaWithInstance(jde-delta, inst)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func neptuneRADerivativeN(jde, delta float64, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := NeptuneApparentRaNWithInstance(jde+delta, n, inst) - NeptuneApparentRaNWithInstance(jde-delta, n, inst)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func neptuneConjunctionFull(jde, degree float64, next uint8, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	//0=last 1=next
	if !isFiniteFloat(jde) || !isFiniteFloat(degree) {
		return math.NaN()
	}
	daysPerDegree := NEPTUNE_S_PERIOD / 360
	currentDelta := neptuneSunLongitudeDelta(jde, degree, false, inst)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJDE := jde
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJDE := estimateJDE
		longitudeDelta := neptuneSunLongitudeDelta(prevJDE, degree, true, inst)
		longitudeSlope := (neptuneSunLongitudeDelta(prevJDE+0.000005, degree, true, inst) - neptuneSunLongitudeDelta(prevJDE-0.000005, degree, true, inst)) / 0.00001
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

func neptuneConjunction(jde, degree float64, next uint8, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	//0=last 1=next
	if !isFiniteFloat(jde) || !isFiniteFloat(degree) {
		return math.NaN()
	}
	daysPerDegree := NEPTUNE_S_PERIOD / 360
	currentDelta := neptuneSunLongitudeDelta(jde, degree, false, inst)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJDE := jde
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJDE := estimateJDE
		longitudeDelta := neptuneSunLongitudeDeltaN(prevJDE, degree, true, neptuneEventSearchN, inst)
		longitudeSlope := (neptuneSunLongitudeDeltaN(prevJDE+0.000005, degree, true, neptuneEventSearchN, inst) - neptuneSunLongitudeDeltaN(prevJDE-0.000005, degree, true, neptuneEventSearchN, inst)) / 0.00001
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
		longitudeDelta := neptuneSunLongitudeDelta(prevJDE, degree, true, inst)
		longitudeSlope := (neptuneSunLongitudeDelta(prevJDE+0.000005, degree, true, inst) - neptuneSunLongitudeDelta(prevJDE-0.000005, degree, true, inst)) / 0.00001
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

// LastNeptuneConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastNeptuneConjunctionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 0, func(jde, degree float64, next uint8) float64 {
		return neptuneConjunction(jde, degree, next, inst)
	})
}

// NextNeptuneConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextNeptuneConjunctionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 0, func(jde, degree float64, next uint8) float64 {
		return neptuneConjunction(jde, degree, next, inst)
	})
}

// LastNeptuneOppositionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastNeptuneOppositionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 180, func(jde, degree float64, next uint8) float64 {
		return neptuneConjunction(jde, degree, next, inst)
	})
}

// NextNeptuneOppositionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextNeptuneOppositionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 180, func(jde, degree float64, next uint8) float64 {
		return neptuneConjunction(jde, degree, next, inst)
	})
}

// NextNeptuneEasternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextNeptuneEasternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 90, func(jde, degree float64, next uint8) float64 {
		return neptuneConjunction(jde, degree, next, inst)
	})
}

// LastNeptuneEasternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastNeptuneEasternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 90, func(jde, degree float64, next uint8) float64 {
		return neptuneConjunction(jde, degree, next, inst)
	})
}

// NextNeptuneWesternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextNeptuneWesternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 270, func(jde, degree float64, next uint8) float64 {
		return neptuneConjunction(jde, degree, next, inst)
	})
}

// LastNeptuneWesternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastNeptuneWesternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 270, func(jde, degree float64, next uint8) float64 {
		return neptuneConjunction(jde, degree, next, inst)
	})
}

func neptuneRetrogradeAroundOpposition(oppositionJD float64, searchBeforeOpposition bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	if !isFiniteFloat(oppositionJD) {
		return math.NaN()
	}
	oppositionTT := UTC2TT(oppositionJD)
	startTT := oppositionTT
	endTT := oppositionTT
	if searchBeforeOpposition {
		easternQuadratureUT := neptuneConjunction(oppositionTT, 90, 0, inst)
		startTT = UTC2TT(easternQuadratureUT)
	} else {
		westernQuadratureUT := neptuneConjunction(oppositionTT, 270, 1, inst)
		endTT = UTC2TT(westernQuadratureUT)
	}
	bestJDE := zeroEventInWindow(startTT, endTT, 2.0, 2.0, 30.0/86400.0, func(jd float64) float64 {
		return neptuneRADerivativeN(jd, stationDerivativeStepDay, neptuneEventSearchN, inst)
	}, func(jd float64) float64 {
		return neptuneRADerivative(jd, stationDerivativeStepDay, inst)
	})
	return TT2UTC(bestJDE)
}

// NextNeptuneRetrogradeToProgradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextNeptuneRetrogradeToProgradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	lastOppositionJD := neptuneConjunctionFull(jde, 180, 0, inst)
	date := neptuneRetrogradeAroundOpposition(lastOppositionJD, false, inst)
	if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
		return date
	}
	nextOppositionJD := neptuneConjunctionFull(jde, 180, 1, inst)
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	date = neptuneRetrogradeAroundOpposition(nextOppositionJD, false, inst)
	if !isFiniteFloat(date) || !stationUTQueryAfterOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

// LastNeptuneRetrogradeToProgradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastNeptuneRetrogradeToProgradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	lastOppositionJD := neptuneConjunctionFull(jde, 180, 0, inst)
	date := neptuneRetrogradeAroundOpposition(lastOppositionJD, false, inst)
	if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(lastOppositionJD) {
		return math.NaN()
	}
	previousOppositionJD := neptuneConjunctionFull(eventUTLastQueryTT(lastOppositionJD), 180, 0, inst)
	if !isFiniteFloat(previousOppositionJD) {
		return math.NaN()
	}
	date = neptuneRetrogradeAroundOpposition(previousOppositionJD, false, inst)
	if !isFiniteFloat(date) || !stationUTQueryBeforeOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

// NextNeptuneProgradeToRetrogradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextNeptuneProgradeToRetrogradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	nextOppositionJD := neptuneConjunctionFull(jde, 180, 1, inst)
	date := neptuneRetrogradeAroundOpposition(nextOppositionJD, true, inst)
	if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	followingOppositionJD := neptuneConjunctionFull(eventUTNextQueryTT(nextOppositionJD), 180, 1, inst)
	if !isFiniteFloat(followingOppositionJD) {
		return math.NaN()
	}
	date = neptuneRetrogradeAroundOpposition(followingOppositionJD, true, inst)
	if !isFiniteFloat(date) || !stationUTQueryAfterOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

// LastNeptuneProgradeToRetrogradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastNeptuneProgradeToRetrogradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	nextOppositionJD := neptuneConjunctionFull(jde, 180, 1, inst)
	date := neptuneRetrogradeAroundOpposition(nextOppositionJD, true, inst)
	if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	lastOppositionJD := neptuneConjunctionFull(jde, 180, 0, inst)
	if !isFiniteFloat(lastOppositionJD) {
		return math.NaN()
	}
	date = neptuneRetrogradeAroundOpposition(lastOppositionJD, true, inst)
	if !isFiniteFloat(date) || !stationUTQueryBeforeOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}
