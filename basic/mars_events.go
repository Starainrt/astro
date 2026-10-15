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

func marsSunLongitudeDelta(jde, degree float64, filter bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := Limit360(Limit360(MarsApparentLoWithInstance(jde, inst)-HSunApparentLoWithInstance(jde, inst)) - degree)
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

func marsSunLongitudeDeltaN(jde, degree float64, filter bool, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := Limit360(Limit360(MarsApparentLoNWithInstance(jde, n, inst)-HSunApparentLoNWithInstance(jde, n, inst)) - degree)
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

func marsRADerivative(jde, val float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := MarsApparentRaWithInstance(jde+val, inst) - MarsApparentRaWithInstance(jde-val, inst)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * val)
}

func marsRADerivativeN(jde, val float64, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := MarsApparentRaNWithInstance(jde+val, n, inst) - MarsApparentRaNWithInstance(jde-val, n, inst)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * val)
}

func marsConjunctionFull(jde, degree float64, next uint8, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	//0=last 1=next
	if !isFiniteFloat(jde) || !isFiniteFloat(degree) {
		return math.NaN()
	}
	daysPerDegree := MARS_S_PERIOD / 360
	currentDelta := marsSunLongitudeDelta(jde, degree, false, inst)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJDE := jde
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJDE := estimateJDE
		longitudeDelta := marsSunLongitudeDelta(prevJDE, degree, true, inst)
		longitudeSlope := (marsSunLongitudeDelta(prevJDE+0.000005, degree, true, inst) - marsSunLongitudeDelta(prevJDE-0.000005, degree, true, inst)) / 0.00001
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

func marsConjunction(jde, degree float64, next uint8, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	//0=last 1=next
	if !isFiniteFloat(jde) || !isFiniteFloat(degree) {
		return math.NaN()
	}
	daysPerDegree := MARS_S_PERIOD / 360
	currentDelta := marsSunLongitudeDelta(jde, degree, false, inst)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJDE := jde
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJDE := estimateJDE
		longitudeDelta := marsSunLongitudeDeltaN(prevJDE, degree, true, marsEventSearchN, inst)
		longitudeSlope := (marsSunLongitudeDeltaN(prevJDE+0.000005, degree, true, marsEventSearchN, inst) - marsSunLongitudeDeltaN(prevJDE-0.000005, degree, true, marsEventSearchN, inst)) / 0.00001
		nextJD := prevJDE - longitudeDelta/longitudeSlope
		estimateJDE = nextJD
		if math.Abs(nextJD-prevJDE) <= marsPhaseCoarseTolerance {
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
		longitudeDelta := marsSunLongitudeDelta(prevJDE, degree, true, inst)
		longitudeSlope := (marsSunLongitudeDelta(prevJDE+0.000005, degree, true, inst) - marsSunLongitudeDelta(prevJDE-0.000005, degree, true, inst)) / 0.00001
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

// LastMarsConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMarsConjunctionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 0, func(jde, degree float64, next uint8) float64 {
		return marsConjunction(jde, degree, next, inst)
	})
}

// NextMarsConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMarsConjunctionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 0, func(jde, degree float64, next uint8) float64 {
		return marsConjunction(jde, degree, next, inst)
	})
}

// LastMarsOppositionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMarsOppositionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 180, func(jde, degree float64, next uint8) float64 {
		return marsConjunction(jde, degree, next, inst)
	})
}

// NextMarsOppositionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMarsOppositionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 180, func(jde, degree float64, next uint8) float64 {
		return marsConjunction(jde, degree, next, inst)
	})
}

// NextMarsEasternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMarsEasternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 90, func(jde, degree float64, next uint8) float64 {
		return marsConjunction(jde, degree, next, inst)
	})
}

// LastMarsEasternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMarsEasternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 90, func(jde, degree float64, next uint8) float64 {
		return marsConjunction(jde, degree, next, inst)
	})
}

// NextMarsWesternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMarsWesternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 270, func(jde, degree float64, next uint8) float64 {
		return marsConjunction(jde, degree, next, inst)
	})
}

// LastMarsWesternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMarsWesternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 270, func(jde, degree float64, next uint8) float64 {
		return marsConjunction(jde, degree, next, inst)
	})
}

func marsRetrogradeAroundOpposition(oppositionJD float64, searchBeforeOpposition bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	if !isFiniteFloat(oppositionJD) {
		return math.NaN()
	}
	oppositionTT := UTC2TT(oppositionJD)
	startTT := oppositionTT
	endTT := oppositionTT
	if searchBeforeOpposition {
		easternQuadratureUT := marsConjunction(oppositionTT, 90, 0, inst)
		startTT = UTC2TT(easternQuadratureUT)
	} else {
		westernQuadratureUT := marsConjunction(oppositionTT, 270, 1, inst)
		endTT = UTC2TT(westernQuadratureUT)
	}
	bestJDE := zeroEventInWindow(startTT, endTT, marsStationCoarseStepDay, marsStationHalfWindowDay, 30.0/86400.0, func(jd float64) float64 {
		return marsRADerivativeN(jd, marsStationDerivativeStepDay, marsEventSearchN, inst)
	}, func(jd float64) float64 {
		return marsRADerivative(jd, marsStationDerivativeStepDay, inst)
	})
	return TT2UTC(bestJDE)
}

// NextMarsRetrogradeToProgradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMarsRetrogradeToProgradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	lastOppositionJD := marsConjunctionFull(jde, 180, 0, inst)
	date := marsRetrogradeAroundOpposition(lastOppositionJD, false, inst)
	if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
		return date
	}
	nextOppositionJD := marsConjunctionFull(jde, 180, 1, inst)
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	date = marsRetrogradeAroundOpposition(nextOppositionJD, false, inst)
	if !isFiniteFloat(date) || !stationUTQueryAfterOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

// LastMarsRetrogradeToProgradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMarsRetrogradeToProgradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	lastOppositionJD := marsConjunctionFull(jde, 180, 0, inst)
	date := marsRetrogradeAroundOpposition(lastOppositionJD, false, inst)
	if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(lastOppositionJD) {
		return math.NaN()
	}
	previousOppositionJD := marsConjunctionFull(eventUTLastQueryTT(lastOppositionJD), 180, 0, inst)
	if !isFiniteFloat(previousOppositionJD) {
		return math.NaN()
	}
	date = marsRetrogradeAroundOpposition(previousOppositionJD, false, inst)
	if !isFiniteFloat(date) || !stationUTQueryBeforeOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

// NextMarsProgradeToRetrogradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextMarsProgradeToRetrogradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	nextOppositionJD := marsConjunctionFull(jde, 180, 1, inst)
	date := marsRetrogradeAroundOpposition(nextOppositionJD, true, inst)
	if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	followingOppositionJD := marsConjunctionFull(eventUTNextQueryTT(nextOppositionJD), 180, 1, inst)
	if !isFiniteFloat(followingOppositionJD) {
		return math.NaN()
	}
	date = marsRetrogradeAroundOpposition(followingOppositionJD, true, inst)
	if !isFiniteFloat(date) || !stationUTQueryAfterOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

// LastMarsProgradeToRetrogradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastMarsProgradeToRetrogradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	nextOppositionJD := marsConjunctionFull(jde, 180, 1, inst)
	date := marsRetrogradeAroundOpposition(nextOppositionJD, true, inst)
	if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	lastOppositionJD := marsConjunctionFull(jde, 180, 0, inst)
	if !isFiniteFloat(lastOppositionJD) {
		return math.NaN()
	}
	date = marsRetrogradeAroundOpposition(lastOppositionJD, true, inst)
	if !isFiniteFloat(date) || !stationUTQueryBeforeOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}
