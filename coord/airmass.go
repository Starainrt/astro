package coord

import "github.com/starainrt/astro/formula"

// AirmassPlaneParallelFromTrueAltitude 平行平板大气质量 / plane-parallel airmass from true altitude.
//
// 输入为真高度角，单位度。适合中高空几何近似，接近地平线时会发散。
func AirmassPlaneParallelFromTrueAltitude(trueAltitude float64) float64 {
	return formula.AirmassPlaneParallel(trueAltitude)
}

// AirmassKastenYoungFromApparentAltitude Kasten-Young 大气质量 / Kasten-Young airmass from apparent altitude.
//
// 输入为视高度角，单位度。
func AirmassKastenYoungFromApparentAltitude(apparentAltitude float64) float64 {
	return formula.AirmassKastenYoung(apparentAltitude)
}

// AirmassPickeringFromApparentAltitude Pickering 大气质量 / Pickering airmass from apparent altitude.
//
// 输入为视高度角，单位度。
func AirmassPickeringFromApparentAltitude(apparentAltitude float64) float64 {
	return formula.AirmassPickering(apparentAltitude)
}

// AirmassKastenYoungFromTrueAltitude Kasten-Young 大气质量 / Kasten-Young airmass from true altitude.
//
// 先用 pressureHPa / temperatureC 估算大气折射，将真高度角换算为视高度角，再代入经验公式。
func AirmassKastenYoungFromTrueAltitude(trueAltitude, pressureHPa, temperatureC float64) float64 {
	return formula.AirmassKastenYoung(ApparentAltitude(trueAltitude, pressureHPa, temperatureC))
}

// AirmassPickeringFromTrueAltitude Pickering 大气质量 / Pickering airmass from true altitude.
//
// 先用 pressureHPa / temperatureC 估算大气折射，将真高度角换算为视高度角，再代入经验公式。
func AirmassPickeringFromTrueAltitude(trueAltitude, pressureHPa, temperatureC float64) float64 {
	return formula.AirmassPickering(ApparentAltitude(trueAltitude, pressureHPa, temperatureC))
}
