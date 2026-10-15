// Package refraction Saemundsson 大气折射及其数值逆解，供 basic 与 formula 共用。
package refraction

import "math"

// StandardPressureHPa 与 StandardTemperatureC 是折射公式的标准气压与气温 / standard pressure and temperature of the refraction formula.
const (
	StandardPressureHPa  = 1010.0
	StandardTemperatureC = 10.0
)

const (
	refractionStandardTemperatureK  = 283.15
	refractionAbsoluteZeroC         = -273.15
	refractionLowerLimitAltitudeDeg = -5.0
	refractionUpperLimitAltitudeDeg = 90.0
)

// RefractionFromApparentAltitude 由视高度角求折射量（度，视高度减去它得真高度），无逆解时为 NaN / refraction from apparent altitude.
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

// TrueAltitude 视高度角转真高度角（度），无逆解时为 NaN / true altitude from apparent altitude.
func TrueAltitude(apparentAltitude, pressureHPa, temperatureC float64) float64 {
	if !validRefractionInputs(apparentAltitude, pressureHPa, temperatureC) {
		return math.NaN()
	}
	if apparentAltitude <= refractionLowerLimitAltitudeDeg || apparentAltitude >= refractionUpperLimitAltitudeDeg {
		return apparentAltitude
	}
	return trueAltitudeFromApparent(apparentAltitude, pressureHPa, temperatureC)
}

// ApparentAltitude 真高度角转视高度角（度） / apparent altitude from true altitude.
func ApparentAltitude(trueAltitude, pressureHPa, temperatureC float64) float64 {
	if !validRefractionInputs(trueAltitude, pressureHPa, temperatureC) {
		return math.NaN()
	}
	return trueAltitude + refractionFromTrueAltitude(trueAltitude, pressureHPa, temperatureC)
}

// RefractionFromTrueAltitude 由真高度角求折射量（度，加到真高度上得视高度） / refraction from true altitude.
func RefractionFromTrueAltitude(trueAltitude, pressureHPa, temperatureC float64) float64 {
	if !validRefractionInputs(trueAltitude, pressureHPa, temperatureC) {
		return math.NaN()
	}
	return refractionFromTrueAltitude(trueAltitude, pressureHPa, temperatureC)
}

// Saemundsson 公式以真高度角为输入；逆函数对同一模型数值求解，保证真/视高度互逆。
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
	// 近 90° 时折射略为负，真高度高于视高度，根区间上界须放宽到 90°。
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
	scale := refractionScale(pressureHPa, temperatureC) * 1.02 / 60
	for i := 0; i < 12; i++ {
		if estimate < lower || estimate > upper || !finiteRefractionValue(estimate) {
			return 0, false
		}
		value := refractionInverseResidual(estimate, apparentAltitude, pressureHPa, temperatureC)
		if math.Abs(value) < 1e-12 {
			return estimate, true
		}
		shifted := estimate + 5.11
		sinAngle := math.Sin((estimate + 10.3/shifted) * math.Pi / 180)
		// d/dh cot(A) = −(dA/dh)/sin²A，A = h + 10.3/(h+5.11)（度）。
		derivative := 1 - scale*(1-10.3/(shifted*shifted))*(math.Pi/180)/(sinAngle*sinAngle)
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
	return pressureHPa / StandardPressureHPa * refractionStandardTemperatureK / (temperatureC - refractionAbsoluteZeroC)
}
