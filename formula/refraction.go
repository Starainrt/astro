package formula

import "github.com/starainrt/astro/internal/refraction"

// RefractionStandardPressureHPa 与 RefractionStandardTemperatureC 是折射公式的标准气压与气温 / standard pressure and temperature of the refraction formula.
const (
	RefractionStandardPressureHPa  = refraction.StandardPressureHPa
	RefractionStandardTemperatureC = refraction.StandardTemperatureC
)

// RefractionFromApparentAltitude 由视高度角求折射量（度，视高度减去它得真高度），无逆解时为 NaN / refraction from apparent altitude.
func RefractionFromApparentAltitude(apparentAltitude, pressureHPa, temperatureC float64) float64 {
	return refraction.RefractionFromApparentAltitude(apparentAltitude, pressureHPa, temperatureC)
}

// TrueAltitude 视高度角转真高度角（度），无逆解时为 NaN / true altitude from apparent altitude.
func TrueAltitude(apparentAltitude, pressureHPa, temperatureC float64) float64 {
	return refraction.TrueAltitude(apparentAltitude, pressureHPa, temperatureC)
}

// ApparentAltitude 真高度角转视高度角（度） / apparent altitude from true altitude.
func ApparentAltitude(trueAltitude, pressureHPa, temperatureC float64) float64 {
	return refraction.ApparentAltitude(trueAltitude, pressureHPa, temperatureC)
}

// RefractionFromTrueAltitude 由真高度角求折射量（度，加到真高度上得视高度） / refraction from true altitude.
func RefractionFromTrueAltitude(trueAltitude, pressureHPa, temperatureC float64) float64 {
	return refraction.RefractionFromTrueAltitude(trueAltitude, pressureHPa, temperatureC)
}
