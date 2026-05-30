package coord

import "github.com/starainrt/astro/formula"

// AirmassPlaneParallelFromTrueAltitude 平行平板大气质量 / plane-parallel airmass from true altitude.
//
// 输入为真高度角，单位度。适合中高空几何近似，接近地平线时会发散。
// Input is true altitude in degrees. This geometric approximation is suitable at moderate and high altitudes but diverges near the horizon.
func AirmassPlaneParallelFromTrueAltitude(trueAltitude float64) float64 {
	return formula.AirmassPlaneParallel(trueAltitude)
}

// AirmassKastenYoungFromApparentAltitude Kasten-Young 大气质量 / Kasten-Young airmass from apparent altitude.
//
// 输入为视高度角，单位度。
// Input is apparent altitude in degrees.
func AirmassKastenYoungFromApparentAltitude(apparentAltitude float64) float64 {
	return formula.AirmassKastenYoung(apparentAltitude)
}

// AirmassPickeringFromApparentAltitude Pickering 大气质量 / Pickering airmass from apparent altitude.
//
// 输入为视高度角，单位度。
// Input is apparent altitude in degrees.
func AirmassPickeringFromApparentAltitude(apparentAltitude float64) float64 {
	return formula.AirmassPickering(apparentAltitude)
}

// AirmassKastenYoungFromTrueAltitude Kasten-Young 大气质量 / Kasten-Young airmass from true altitude.
//
// 先用 pressureHPa / temperatureC 估算大气折射，将真高度角换算为视高度角，再代入经验公式。
// First estimates atmospheric refraction from pressureHPa and temperatureC, converts true altitude to apparent altitude, and then applies the empirical formula.
func AirmassKastenYoungFromTrueAltitude(trueAltitude, pressureHPa, temperatureC float64) float64 {
	return formula.AirmassKastenYoung(ApparentAltitude(trueAltitude, pressureHPa, temperatureC))
}

// AirmassPickeringFromTrueAltitude Pickering 大气质量 / Pickering airmass from true altitude.
//
// 先用 pressureHPa / temperatureC 估算大气折射，将真高度角换算为视高度角，再代入经验公式。
// First estimates atmospheric refraction from pressureHPa and temperatureC, converts true altitude to apparent altitude, and then applies the empirical formula.
func AirmassPickeringFromTrueAltitude(trueAltitude, pressureHPa, temperatureC float64) float64 {
	return formula.AirmassPickering(ApparentAltitude(trueAltitude, pressureHPa, temperatureC))
}
