package basic

import "math"

const (
	refractionStandardPressureHPa   = 1010.0
	refractionStandardTemperatureK  = 283.0
	refractionAbsoluteZeroC         = -273.15
	refractionLowerLimitAltitudeDeg = -5.0
	refractionUpperLimitAltitudeDeg = 90.0
)

// RefractionFromApparentAltitude 大气折射修正量，单位度；输入为视高度角。
// 返回值应从真高度角加上后得到视高度角。
func RefractionFromApparentAltitude(apparentAltitude, pressureHPa, temperatureC float64) float64 {
	if !validRefractionInputs(apparentAltitude, pressureHPa, temperatureC) {
		return math.NaN()
	}
	if apparentAltitude < refractionLowerLimitAltitudeDeg || apparentAltitude > refractionUpperLimitAltitudeDeg {
		return 0
	}
	angle := (apparentAltitude + 10.3/(apparentAltitude+5.11)) * math.Pi / 180
	return refractionScale(pressureHPa, temperatureC) * (1.02 / math.Tan(angle)) / 60
}

// TrueAltitude 真高度角，单位度；输入为视高度角。
func TrueAltitude(apparentAltitude, pressureHPa, temperatureC float64) float64 {
	refraction := RefractionFromApparentAltitude(apparentAltitude, pressureHPa, temperatureC)
	if math.IsNaN(refraction) {
		return math.NaN()
	}
	return apparentAltitude - refraction
}

// ApparentAltitude 视高度角，单位度；输入为真高度角。
func ApparentAltitude(trueAltitude, pressureHPa, temperatureC float64) float64 {
	if !validRefractionInputs(trueAltitude, pressureHPa, temperatureC) {
		return math.NaN()
	}
	if trueAltitude < refractionLowerLimitAltitudeDeg || trueAltitude > refractionUpperLimitAltitudeDeg {
		return trueAltitude
	}

	estimate := trueAltitude + RefractionFromApparentAltitude(trueAltitude, pressureHPa, temperatureC)
	for i := 0; i < 8; i++ {
		refraction := RefractionFromApparentAltitude(estimate, pressureHPa, temperatureC)
		if math.IsNaN(refraction) {
			return math.NaN()
		}
		value := estimate - refraction - trueAltitude
		if math.Abs(value) < 1e-12 {
			break
		}

		const delta = 1e-6
		refractionPlus := RefractionFromApparentAltitude(estimate+delta, pressureHPa, temperatureC)
		refractionMinus := RefractionFromApparentAltitude(estimate-delta, pressureHPa, temperatureC)
		if math.IsNaN(refractionPlus) || math.IsNaN(refractionMinus) {
			return math.NaN()
		}

		derivative := 1 - (refractionPlus-refractionMinus)/(2*delta)
		if derivative == 0 {
			break
		}
		estimate -= value / derivative
	}
	return estimate
}

// RefractionFromTrueAltitude 大气折射修正量，单位度；输入为真高度角。
// 返回值应从真高度角加上后得到视高度角。
func RefractionFromTrueAltitude(trueAltitude, pressureHPa, temperatureC float64) float64 {
	apparentAltitude := ApparentAltitude(trueAltitude, pressureHPa, temperatureC)
	if math.IsNaN(apparentAltitude) {
		return math.NaN()
	}
	return apparentAltitude - trueAltitude
}

func validRefractionInputs(altitude, pressureHPa, temperatureC float64) bool {
	return !(math.IsNaN(altitude) || math.IsInf(altitude, 0) ||
		math.IsNaN(pressureHPa) || math.IsInf(pressureHPa, 0) || pressureHPa <= 0 ||
		math.IsNaN(temperatureC) || math.IsInf(temperatureC, 0) || temperatureC <= refractionAbsoluteZeroC)
}

func refractionScale(pressureHPa, temperatureC float64) float64 {
	return pressureHPa / refractionStandardPressureHPa * refractionStandardTemperatureK / (273 + temperatureC)
}
