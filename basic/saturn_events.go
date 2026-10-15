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

func saturnSunLongitudeDelta(jde, degree float64, filter bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := Limit360(Limit360(SaturnApparentLoWithInstance(jde, inst)-HSunApparentLoWithInstance(jde, inst)) - degree)
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

func saturnSunLongitudeDeltaN(jde, degree float64, filter bool, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := Limit360(Limit360(SaturnApparentLoNWithInstance(jde, n, inst)-HSunApparentLoNWithInstance(jde, n, inst)) - degree)
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

func saturnRADerivative(jde, delta float64, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := SaturnApparentRaWithInstance(jde+delta, inst) - SaturnApparentRaWithInstance(jde-delta, inst)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func saturnRADerivativeN(jde, delta float64, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sub := SaturnApparentRaNWithInstance(jde+delta, n, inst) - SaturnApparentRaNWithInstance(jde-delta, n, inst)
	if sub > 180 {
		sub -= 360
	}
	if sub < -180 {
		sub += 360
	}
	return sub / (2 * delta)
}

func saturnConjunctionFull(jde, degree float64, next uint8, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	//0=last 1=next
	if !isFiniteFloat(jde) || !isFiniteFloat(degree) {
		return math.NaN()
	}
	daysPerDegree := SATURN_S_PERIOD / 360
	currentDelta := saturnSunLongitudeDelta(jde, degree, false, inst)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJDE := jde
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJDE := estimateJDE
		longitudeDelta := saturnSunLongitudeDelta(prevJDE, degree, true, inst)
		longitudeSlope := (saturnSunLongitudeDelta(prevJDE+0.000005, degree, true, inst) - saturnSunLongitudeDelta(prevJDE-0.000005, degree, true, inst)) / 0.00001
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

func saturnConjunction(jde, degree float64, next uint8, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	//0=last 1=next
	if !isFiniteFloat(jde) || !isFiniteFloat(degree) {
		return math.NaN()
	}
	daysPerDegree := SATURN_S_PERIOD / 360
	currentDelta := saturnSunLongitudeDelta(jde, degree, false, inst)
	if next == 0 {
		jde -= (360 - currentDelta) * daysPerDegree
	} else {
		jde += daysPerDegree * currentDelta
	}
	estimateJDE := jde
	converged := false
	for i := 0; i < eventNewtonMaxIterations; i++ {
		prevJDE := estimateJDE
		longitudeDelta := saturnSunLongitudeDeltaN(prevJDE, degree, true, saturnEventSearchN, inst)
		longitudeSlope := (saturnSunLongitudeDeltaN(prevJDE+0.000005, degree, true, saturnEventSearchN, inst) - saturnSunLongitudeDeltaN(prevJDE-0.000005, degree, true, saturnEventSearchN, inst)) / 0.00001
		nextJD := prevJDE - longitudeDelta/longitudeSlope
		estimateJDE = nextJD
		if math.Abs(nextJD-prevJDE) <= saturnPhaseCoarseTolerance {
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
		longitudeDelta := saturnSunLongitudeDelta(prevJDE, degree, true, inst)
		longitudeSlope := (saturnSunLongitudeDelta(prevJDE+0.000005, degree, true, inst) - saturnSunLongitudeDelta(prevJDE-0.000005, degree, true, inst)) / 0.00001
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

// LastSaturnConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastSaturnConjunctionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 0, func(jde, degree float64, next uint8) float64 {
		return saturnConjunction(jde, degree, next, inst)
	})
}

// NextSaturnConjunctionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextSaturnConjunctionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 0, func(jde, degree float64, next uint8) float64 {
		return saturnConjunction(jde, degree, next, inst)
	})
}

// LastSaturnOppositionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastSaturnOppositionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 180, func(jde, degree float64, next uint8) float64 {
		return saturnConjunction(jde, degree, next, inst)
	})
}

// NextSaturnOppositionWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextSaturnOppositionWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 180, func(jde, degree float64, next uint8) float64 {
		return saturnConjunction(jde, degree, next, inst)
	})
}

// NextSaturnEasternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextSaturnEasternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 90, func(jde, degree float64, next uint8) float64 {
		return saturnConjunction(jde, degree, next, inst)
	})
}

// LastSaturnEasternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastSaturnEasternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 90, func(jde, degree float64, next uint8) float64 {
		return saturnConjunction(jde, degree, next, inst)
	})
}

// NextSaturnWesternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextSaturnWesternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveNextPhaseEvent(jde, 270, func(jde, degree float64, next uint8) float64 {
		return saturnConjunction(jde, degree, next, inst)
	})
}

// LastSaturnWesternQuadratureWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastSaturnWesternQuadratureWithInstance(jde float64, inst *Instance) float64 {
	return inclusiveLastPhaseEvent(jde, 270, func(jde, degree float64, next uint8) float64 {
		return saturnConjunction(jde, degree, next, inst)
	})
}

func saturnRetrogradeAroundOpposition(oppositionJD float64, searchBeforeOpposition bool, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	if !isFiniteFloat(oppositionJD) {
		return math.NaN()
	}
	oppositionTT := UTC2TT(oppositionJD)
	startTT := oppositionTT
	endTT := oppositionTT
	if searchBeforeOpposition {
		easternQuadratureUT := saturnConjunction(oppositionTT, 90, 0, inst)
		startTT = UTC2TT(easternQuadratureUT)
	} else {
		westernQuadratureUT := saturnConjunction(oppositionTT, 270, 1, inst)
		endTT = UTC2TT(westernQuadratureUT)
	}
	bestJDE := zeroEventInWindow(startTT, endTT, 2.0, 2.0, 30.0/86400.0, func(jd float64) float64 {
		return saturnRADerivativeN(jd, stationDerivativeStepDay, saturnEventSearchN, inst)
	}, func(jd float64) float64 {
		return saturnRADerivative(jd, stationDerivativeStepDay, inst)
	})
	return TT2UTC(bestJDE)
}

// NextSaturnRetrogradeToProgradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextSaturnRetrogradeToProgradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	lastOppositionJD := saturnConjunctionFull(jde, 180, 0, inst)
	date := saturnRetrogradeAroundOpposition(lastOppositionJD, false, inst)
	if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
		return date
	}
	nextOppositionJD := saturnConjunctionFull(jde, 180, 1, inst)
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	date = saturnRetrogradeAroundOpposition(nextOppositionJD, false, inst)
	if !isFiniteFloat(date) || !stationUTQueryAfterOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

// LastSaturnRetrogradeToProgradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastSaturnRetrogradeToProgradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	lastOppositionJD := saturnConjunctionFull(jde, 180, 0, inst)
	date := saturnRetrogradeAroundOpposition(lastOppositionJD, false, inst)
	if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(lastOppositionJD) {
		return math.NaN()
	}
	previousOppositionJD := saturnConjunctionFull(eventUTLastQueryTT(lastOppositionJD), 180, 0, inst)
	if !isFiniteFloat(previousOppositionJD) {
		return math.NaN()
	}
	date = saturnRetrogradeAroundOpposition(previousOppositionJD, false, inst)
	if !isFiniteFloat(date) || !stationUTQueryBeforeOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

// NextSaturnProgradeToRetrogradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextSaturnProgradeToRetrogradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	nextOppositionJD := saturnConjunctionFull(jde, 180, 1, inst)
	date := saturnRetrogradeAroundOpposition(nextOppositionJD, true, inst)
	if isFiniteFloat(date) && stationUTQueryAfterOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	followingOppositionJD := saturnConjunctionFull(eventUTNextQueryTT(nextOppositionJD), 180, 1, inst)
	if !isFiniteFloat(followingOppositionJD) {
		return math.NaN()
	}
	date = saturnRetrogradeAroundOpposition(followingOppositionJD, true, inst)
	if !isFiniteFloat(date) || !stationUTQueryAfterOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}

// LastSaturnProgradeToRetrogradeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastSaturnProgradeToRetrogradeWithInstance(jde float64, inst *Instance) float64 {
	if !isFiniteFloat(jde) {
		return math.NaN()
	}
	nextOppositionJD := saturnConjunctionFull(jde, 180, 1, inst)
	date := saturnRetrogradeAroundOpposition(nextOppositionJD, true, inst)
	if isFiniteFloat(date) && stationUTQueryBeforeOrEqual(date, jde) {
		return date
	}
	if !isFiniteFloat(nextOppositionJD) {
		return math.NaN()
	}
	lastOppositionJD := saturnConjunctionFull(jde, 180, 0, inst)
	if !isFiniteFloat(lastOppositionJD) {
		return math.NaN()
	}
	date = saturnRetrogradeAroundOpposition(lastOppositionJD, true, inst)
	if !isFiniteFloat(date) || !stationUTQueryBeforeOrEqual(date, jde) {
		return math.NaN()
	}
	return date
}
