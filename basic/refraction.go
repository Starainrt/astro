package basic

import "github.com/starainrt/astro/internal/refraction"

const (
	refractionStandardPressureHPa  = refraction.StandardPressureHPa
	refractionStandardTemperatureC = refraction.StandardTemperatureC
)

// RefractionFromApparentAltitude 大气折射修正量，单位度；输入为视高度角。
// 返回值应从视高度角减去后得到真高度角。
// 若模型在支持的真高度范围内没有逆解，则返回 NaN。
func RefractionFromApparentAltitude(apparentAltitude, pressureHPa, temperatureC float64) float64 {
	return refraction.RefractionFromApparentAltitude(apparentAltitude, pressureHPa, temperatureC)
}

// TrueAltitude 真高度角，单位度；输入为视高度角。若折射模型在支持的
// 真高度范围内没有逆解，则返回 NaN。
func TrueAltitude(apparentAltitude, pressureHPa, temperatureC float64) float64 {
	return refraction.TrueAltitude(apparentAltitude, pressureHPa, temperatureC)
}

// ApparentAltitude 视高度角，单位度；输入为真高度角。
func ApparentAltitude(trueAltitude, pressureHPa, temperatureC float64) float64 {
	return refraction.ApparentAltitude(trueAltitude, pressureHPa, temperatureC)
}

// RefractionFromTrueAltitude 大气折射修正量，单位度；输入为真高度角。
// 返回值应从真高度角加上后得到视高度角。
func RefractionFromTrueAltitude(trueAltitude, pressureHPa, temperatureC float64) float64 {
	return refraction.RefractionFromTrueAltitude(trueAltitude, pressureHPa, temperatureC)
}
