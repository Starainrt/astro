package basic

import "math"

const (
	refractionStandardPressureHPa   = 1010.0
	refractionStandardTemperatureC  = 10.0
	refractionStandardTemperatureK  = 283.15
	refractionAbsoluteZeroC         = -273.15
	refractionLowerLimitAltitudeDeg = -5.0
	refractionUpperLimitAltitudeDeg = 90.0
)

// RefractionFromApparentAltitude 大气折射修正量，单位度；输入为视高度角。
// 返回值应从视高度角减去后得到真高度角。
// 若模型在支持的真高度范围内没有逆解，则返回 NaN。
func RefractionFromApparentAltitude(apparentAltitude, pressureHPa, temperatureC float64) float64 {
	if !validRefractionInputs(apparentAltitude, pressureHPa, temperatureC) {
		return math.NaN()
	}
	if apparentAltitude <= refractionLowerLimitAltitudeDeg || apparentAltitude >= refractionUpperLimitAltitudeDeg {
		return 0
	}
	trueAltitude := trueAltitudeFromApparent(apparentAltitude, pressureHPa, temperatureC)
	return apparentAltitude - trueAltitude
}

// TrueAltitude 真高度角，单位度；输入为视高度角。若折射模型在支持的
// 真高度范围内没有逆解，则返回 NaN。
func TrueAltitude(apparentAltitude, pressureHPa, temperatureC float64) float64 {
	if !validRefractionInputs(apparentAltitude, pressureHPa, temperatureC) {
		return math.NaN()
	}
	if apparentAltitude <= refractionLowerLimitAltitudeDeg || apparentAltitude >= refractionUpperLimitAltitudeDeg {
		return apparentAltitude
	}
	return trueAltitudeFromApparent(apparentAltitude, pressureHPa, temperatureC)
}

// ApparentAltitude 视高度角，单位度；输入为真高度角。
func ApparentAltitude(trueAltitude, pressureHPa, temperatureC float64) float64 {
	if !validRefractionInputs(trueAltitude, pressureHPa, temperatureC) {
		return math.NaN()
	}
	return trueAltitude + refractionFromTrueAltitude(trueAltitude, pressureHPa, temperatureC)
}

// RefractionFromTrueAltitude 大气折射修正量，单位度；输入为真高度角。
// 返回值应从真高度角加上后得到视高度角。
func RefractionFromTrueAltitude(trueAltitude, pressureHPa, temperatureC float64) float64 {
	if !validRefractionInputs(trueAltitude, pressureHPa, temperatureC) {
		return math.NaN()
	}
	return refractionFromTrueAltitude(trueAltitude, pressureHPa, temperatureC)
}

// Saemundsson 公式以真高度角为输入；逆 API 对同一模型做数值求解，保持公开真/视高度语义一致。
// Saemundsson's formula takes true altitude. The inverse APIs solve the same model numerically so the public true/apparent semantics remain consistent.
func refractionFromTrueAltitude(trueAltitude, pressureHPa, temperatureC float64) float64 {
	if trueAltitude <= refractionLowerLimitAltitudeDeg || trueAltitude >= refractionUpperLimitAltitudeDeg {
		return 0
	}
	angle := (trueAltitude + 10.3/(trueAltitude+5.11)) * math.Pi / 180
	return refractionScale(pressureHPa, temperatureC) * (1.02 / math.Tan(angle)) / 60
}

func trueAltitudeFromApparent(apparentAltitude, pressureHPa, temperatureC float64) float64 {
	lower := math.Nextafter(refractionLowerLimitAltitudeDeg, math.Inf(1))
	upper := apparentAltitude
	// Saemundsson 近似在 90 度以下极窄范围会略为负值；此时真高度角高于视高度角。若用视高度角作为根区间上界会漏掉有效解，因此保留模型明确的 90 度边界，允许逆解越过输入的视高度角。
	// Saemundsson's approximation becomes slightly negative just below 90 degrees. In that narrow range the true altitude is above the apparent altitude, so using the apparent altitude as the root bracket would omit the valid solution. Keep the model's explicit 90-degree boundary while allowing the inverse to cross above the apparent input.
	if refractionFromTrueAltitude(apparentAltitude, pressureHPa, temperatureC) < 0 {
		upper = math.Nextafter(refractionUpperLimitAltitudeDeg, math.Inf(-1))
	}
	if estimate, ok := trueAltitudeFromApparentNewton(apparentAltitude, pressureHPa, temperatureC, lower, upper); ok {
		return estimate
	}
	return trueAltitudeFromApparentBisection(apparentAltitude, pressureHPa, temperatureC, lower, upper)
}

func trueAltitudeFromApparentNewton(apparentAltitude, pressureHPa, temperatureC, lower, upper float64) (float64, bool) {
	estimate := apparentAltitude - refractionFromTrueAltitude(apparentAltitude, pressureHPa, temperatureC)
	const delta = 1e-6
	for i := 0; i < 12; i++ {
		if estimate < lower || estimate > upper || !finiteRefractionValue(estimate) {
			return 0, false
		}
		value := refractionInverseResidual(estimate, apparentAltitude, pressureHPa, temperatureC)
		if math.Abs(value) < 1e-12 {
			return estimate, true
		}
		refractionPlus := refractionFromTrueAltitude(estimate+delta, pressureHPa, temperatureC)
		refractionMinus := refractionFromTrueAltitude(estimate-delta, pressureHPa, temperatureC)
		derivative := 1 + (refractionPlus-refractionMinus)/(2*delta)
		if math.Abs(derivative) < 1e-12 || !finiteRefractionValue(derivative) {
			return 0, false
		}
		next := estimate - value/derivative
		if next < lower || next > upper || !finiteRefractionValue(next) {
			return 0, false
		}
		estimate = next
	}
	return 0, false
}

func trueAltitudeFromApparentBisection(apparentAltitude, pressureHPa, temperatureC, lower, upper float64) float64 {
	lowerValue := refractionInverseResidual(lower, apparentAltitude, pressureHPa, temperatureC)
	upperValue := refractionInverseResidual(upper, apparentAltitude, pressureHPa, temperatureC)
	if !finiteRefractionValue(lowerValue) || !finiteRefractionValue(upperValue) || lowerValue > 0 || upperValue < 0 {
		return math.NaN()
	}
	if math.Abs(lowerValue) < 1e-12 {
		return lower
	}
	if math.Abs(upperValue) < 1e-12 {
		return upper
	}

	best, bestResidual := lower, math.Abs(lowerValue)
	if math.Abs(upperValue) < bestResidual {
		best, bestResidual = upper, math.Abs(upperValue)
	}
	for i := 0; i < 96; i++ {
		midpoint := lower + (upper-lower)/2
		if midpoint == lower || midpoint == upper {
			break
		}
		value := refractionInverseResidual(midpoint, apparentAltitude, pressureHPa, temperatureC)
		if !finiteRefractionValue(value) {
			return math.NaN()
		}
		if residual := math.Abs(value); residual < bestResidual {
			best, bestResidual = midpoint, residual
		}
		if math.Abs(value) < 1e-12 {
			return midpoint
		}
		if value > 0 {
			upper = midpoint
		} else {
			lower = midpoint
		}
	}
	if bestResidual <= 1e-9 {
		return best
	}
	return math.NaN()
}

func refractionInverseResidual(trueAltitude, apparentAltitude, pressureHPa, temperatureC float64) float64 {
	return trueAltitude + refractionFromTrueAltitude(trueAltitude, pressureHPa, temperatureC) - apparentAltitude
}

func finiteRefractionValue(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func validRefractionInputs(altitude, pressureHPa, temperatureC float64) bool {
	return !(math.IsNaN(altitude) || math.IsInf(altitude, 0) ||
		math.IsNaN(pressureHPa) || math.IsInf(pressureHPa, 0) || pressureHPa <= 0 ||
		math.IsNaN(temperatureC) || math.IsInf(temperatureC, 0) || temperatureC <= refractionAbsoluteZeroC)
}

func refractionScale(pressureHPa, temperatureC float64) float64 {
	return pressureHPa / refractionStandardPressureHPa * refractionStandardTemperatureK / (temperatureC - refractionAbsoluteZeroC)
}
