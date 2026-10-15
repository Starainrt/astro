package internal

import "math"

const (
	refractionStandardPressureHPa  = 1010.0
	refractionStandardTemperatureK = 283.15
	refractionAbsoluteZeroC        = -273.15
	refractionLowerLimitAltitude   = -5.0
	refractionUpperLimitAltitude   = 90.0
)

// RefractionFromTrueAltitude 真高度到视高度的折射修正 / true-to-apparent altitude refraction.
func RefractionFromTrueAltitude(trueAltitude, pressureHPa, temperatureC float64) float64 {
	if !validRefractionInputs(trueAltitude, pressureHPa, temperatureC) || trueAltitude <= refractionLowerLimitAltitude || trueAltitude >= refractionUpperLimitAltitude {
		return 0
	}
	angle := (trueAltitude + 10.3/(trueAltitude+5.11)) * math.Pi / 180
	scale := pressureHPa / refractionStandardPressureHPa * refractionStandardTemperatureK / (temperatureC - refractionAbsoluteZeroC)
	return scale * (1.02 / math.Tan(angle)) / 60
}

func validRefractionInputs(altitude, pressureHPa, temperatureC float64) bool {
	return !(math.IsNaN(altitude) || math.IsInf(altitude, 0) || math.IsNaN(pressureHPa) || math.IsInf(pressureHPa, 0) || pressureHPa <= 0 || math.IsNaN(temperatureC) || math.IsInf(temperatureC, 0) || temperatureC <= refractionAbsoluteZeroC)
}
