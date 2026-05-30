// Package formula 提供与具体时刻、星历表无关的研究型天文公式。
package formula

import "math"

const (
	planckConstant           = 6.62607015e-34
	speedOfLight             = 299792458.0
	boltzmannConstant        = 1.380649e-23
	stefanBoltzmannConstant  = 5.670374419e-8
	wienDisplacementConstant = 2.897771955e-3
)

// WienPeakWavelength 维恩峰值波长 / Wien peak wavelength.
//
//	temperatureK: 黑体温度，单位开尔文
//
// 返回：
//
//	峰值波长，单位米
//
// Returns the wavelength of maximum emission in meters for a blackbody at the supplied temperature.
func WienPeakWavelength(temperatureK float64) float64 {
	if temperatureK <= 0 || math.IsNaN(temperatureK) || math.IsInf(temperatureK, 0) {
		return math.NaN()
	}
	return wienDisplacementConstant / temperatureK
}

// StefanBoltzmannFlux 斯特藩-玻尔兹曼通量 / Stefan-Boltzmann flux.
//
//	temperatureK: 黑体温度，单位开尔文
//
// 返回：
//
//	单位面积总出射度，单位 W/m^2
//
// Returns the total radiant exitance in W/m^2 for a blackbody at the supplied temperature.
func StefanBoltzmannFlux(temperatureK float64) float64 {
	if temperatureK < 0 || math.IsNaN(temperatureK) || math.IsInf(temperatureK, 0) {
		return math.NaN()
	}
	return stefanBoltzmannConstant * math.Pow(temperatureK, 4)
}

// PlanckRadianceByWavelength 按波长的普朗克谱辐亮度 / Planck spectral radiance by wavelength.
//
//	wavelengthM: 波长，单位米
//	temperatureK: 黑体温度，单位开尔文
//
// 返回：
//
//	谱辐亮度，单位 W·sr^-1·m^-3
//
// Returns spectral radiance in W·sr^-1·m^-3 at the supplied wavelength and temperature.
//
// 例：
//
//	b := formula.PlanckRadianceByWavelength(500e-9, 5772)
func PlanckRadianceByWavelength(wavelengthM, temperatureK float64) float64 {
	if wavelengthM <= 0 || temperatureK <= 0 ||
		math.IsNaN(wavelengthM) || math.IsInf(wavelengthM, 0) ||
		math.IsNaN(temperatureK) || math.IsInf(temperatureK, 0) {
		return math.NaN()
	}

	exponent := planckConstant * speedOfLight / (wavelengthM * boltzmannConstant * temperatureK)
	denominator := math.Expm1(exponent)
	if denominator == 0 {
		return math.Inf(1)
	}
	return 2 * planckConstant * speedOfLight * speedOfLight / math.Pow(wavelengthM, 5) / denominator
}
