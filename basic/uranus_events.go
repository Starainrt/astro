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

func uranusSunLongitudeDelta(jde, degree float64, filter bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := Limit360(Limit360(UranusApparentLoWithInstance(jde, inst)-HSunApparentLoWithInstance(jde, inst)) - degree)
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

func uranusSunLongitudeDeltaN(jde, degree float64, filter bool, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := Limit360(Limit360(UranusApparentLoNWithInstance(jde, n, inst)-HSunApparentLoNWithInstance(jde, n, inst)) - degree)
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

func uranusRADerivative(jde, delta float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := UranusApparentRaWithInstance(jde+delta, inst) - UranusApparentRaWithInstance(jde-delta, inst)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func uranusRADerivativeN(jde, delta float64, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := UranusApparentRaNWithInstance(jde+delta, n, inst) - UranusApparentRaNWithInstance(jde-delta, n, inst)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func uranusConjunctionFull(jde, degree float64, next uint8, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	//0=last 1=next
	if !isFiniteFloat(jde) || !isFiniteFloat(degree) {
		return math.NaN()
	}
	daysPerDegree := URANUS_S_PERIOD / 360
	currentDelta := uranusSunLongitudeDelta(jde, degree, false, inst)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJDE := jde
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJDE := estimateJDE
		longitudeDelta := uranusSunLongitudeDelta(prevJDE, degree, true, inst)
		longitudeSlope := (uranusSunLongitudeDelta(prevJDE+0.000005, degree, true, inst) - uranusSunLongitudeDelta(prevJDE-0.000005, degree, true, inst)) / 0.00001
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

func uranusConjunction(jde, degree float64, next uint8, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	//0=last 1=next
	if !isFiniteFloat(jde) || !isFiniteFloat(degree) {
		return math.NaN()
	}
	daysPerDegree := URANUS_S_PERIOD / 360
	currentDelta := uranusSunLongitudeDelta(jde, degree, false, inst)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJDE := jde
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJDE := estimateJDE
		longitudeDelta := uranusSunLongitudeDeltaN(prevJDE, degree, true, uranusEventSearchN, inst)
		longitudeSlope := (uranusSunLongitudeDeltaN(prevJDE+0.000005, degree, true, uranusEventSearchN, inst) - uranusSunLongitudeDeltaN(prevJDE-0.000005, degree, true, uranusEventSearchN, inst)) / 0.00001
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
		longitudeDelta := uranusSunLongitudeDelta(prevJDE, degree, true, inst)
		longitudeSlope := (uranusSunLongitudeDelta(prevJDE+0.000005, degree, true, inst) - uranusSunLongitudeDelta(prevJDE-0.000005, degree, true, inst)) / 0.00001
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

// LastUranusConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastUranusConjunctionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 0, func(jde, degree float64, next uint8) float64 {
		return uranusConjunction(jde, degree, next, inst)
	})
}

// NextUranusConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextUranusConjunctionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 0, func(jde, degree float64, next uint8) float64 {
		return uranusConjunction(jde, degree, next, inst)
	})
}

// LastUranusOppositionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastUranusOppositionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 180, func(jde, degree float64, next uint8) float64 {
		return uranusConjunction(jde, degree, next, inst)
	})
}

// NextUranusOppositionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextUranusOppositionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 180, func(jde, degree float64, next uint8) float64 {
		return uranusConjunction(jde, degree, next, inst)
	})
}

// NextUranusEasternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextUranusEasternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 90, func(jde, degree float64, next uint8) float64 {
		return uranusConjunction(jde, degree, next, inst)
	})
}

// LastUranusEasternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastUranusEasternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 90, func(jde, degree float64, next uint8) float64 {
		return uranusConjunction(jde, degree, next, inst)
	})
}

// NextUranusWesternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextUranusWesternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 270, func(jde, degree float64, next uint8) float64 {
		return uranusConjunction(jde, degree, next, inst)
	})
}

// LastUranusWesternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastUranusWesternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 270, func(jde, degree float64, next uint8) float64 {
		return uranusConjunction(jde, degree, next, inst)
	})
}

func uranusRetrogradeAroundOpposition(oppositionJD float64, searchBeforeOpposition bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	if !isFiniteFloat(oppositionJD) {
		return math.NaN()
	}
	oppositionTT := UTC2TT(oppositionJD)
	startTT := oppositionTT
	endTT := oppositionTT
	if searchBeforeOpposition {
		easternQuadratureUT := uranusConjunction(oppositionTT, 90, 0, inst)
		startTT = UTC2TT(easternQuadratureUT)
	} else {
		westernQuadratureUT := uranusConjunction(oppositionTT, 270, 1, inst)
		endTT = UTC2TT(westernQuadratureUT)
	}
	bestJDE := zeroEventInWindow(startTT, endTT, 2.0, 2.0, 30.0/86400.0, func(jd float64) float64 {
		return uranusRADerivativeN(jd, stationDerivativeStepDay, uranusEventSearchN, inst)
	}, func(jd float64) float64 {
		return uranusRADerivative(jd, stationDerivativeStepDay, inst)
	})
	return TT2UTC(bestJDE)
}

// NextUranusRetrogradeToProgradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextUranusRetrogradeToProgradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	lastOppositionJD := uranusConjunctionFull(jde, 180, 0, inst)
	date := uranusRetrogradeAroundOpposition(lastOppositionJD, false, inst)
	if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
		return date
	}
	nextOppositionJD := uranusConjunctionFull(jde, 180, 1, inst)
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	date = uranusRetrogradeAroundOpposition(nextOppositionJD, false, inst)
	if !isFiniteFloat(date) || !stationUTQueryAfterOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

// LastUranusRetrogradeToProgradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastUranusRetrogradeToProgradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	lastOppositionJD := uranusConjunctionFull(jde, 180, 0, inst)
	date := uranusRetrogradeAroundOpposition(lastOppositionJD, false, inst)
	if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(lastOppositionJD) {
		return math.NaN()
	}
	previousOppositionJD := uranusConjunctionFull(eventUTLastQueryTT(lastOppositionJD), 180, 0, inst)
	if !isFiniteFloat(previousOppositionJD) {
		return math.NaN()
	}
	date = uranusRetrogradeAroundOpposition(previousOppositionJD, false, inst)
	if !isFiniteFloat(date) || !stationUTQueryBeforeOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

// NextUranusProgradeToRetrogradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextUranusProgradeToRetrogradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	nextOppositionJD := uranusConjunctionFull(jde, 180, 1, inst)
	date := uranusRetrogradeAroundOpposition(nextOppositionJD, true, inst)
	if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	followingOppositionJD := uranusConjunctionFull(eventUTNextQueryTT(nextOppositionJD), 180, 1, inst)
	if !isFiniteFloat(followingOppositionJD) {
		return math.NaN()
	}
	date = uranusRetrogradeAroundOpposition(followingOppositionJD, true, inst)
	if !isFiniteFloat(date) || !stationUTQueryAfterOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

// LastUranusProgradeToRetrogradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastUranusProgradeToRetrogradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	nextOppositionJD := uranusConjunctionFull(jde, 180, 1, inst)
	date := uranusRetrogradeAroundOpposition(nextOppositionJD, true, inst)
	if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	lastOppositionJD := uranusConjunctionFull(jde, 180, 0, inst)
	if !isFiniteFloat(lastOppositionJD) {
		return math.NaN()
	}
	date = uranusRetrogradeAroundOpposition(lastOppositionJD, true, inst)
	if !isFiniteFloat(date) || !stationUTQueryBeforeOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}
