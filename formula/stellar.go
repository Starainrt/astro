package formula

import "math"

const (
	solarLuminosityW    = 3.828e26
	solarRadiusM        = 6.957e8
	solarEffectiveTempK = 5772.0
)

// LuminosityFromRadiusTemperature 由半径和温度求光度 / luminosity from radius and temperature.
//
//	radiusM: 恒星半径，单位米
//	temperatureK: 恒星有效温度，单位开尔文
//
// 返回：
//
//	总光度，单位瓦特
func LuminosityFromRadiusTemperature(radiusM, temperatureK float64) float64 {
	if radiusM <= 0 || temperatureK <= 0 ||
		math.IsNaN(radiusM) || math.IsInf(radiusM, 0) ||
		math.IsNaN(temperatureK) || math.IsInf(temperatureK, 0) {
		return math.NaN()
	}
	return 4 * math.Pi * radiusM * radiusM * StefanBoltzmannFlux(temperatureK)
}

// RadiusFromLuminosityTemperature 由光度和温度求半径 / radius from luminosity and temperature.
//
//	luminosityW: 恒星总光度，单位瓦特
//	temperatureK: 恒星有效温度，单位开尔文
//
// 返回：
//
//	恒星半径，单位米
func RadiusFromLuminosityTemperature(luminosityW, temperatureK float64) float64 {
	if luminosityW <= 0 || temperatureK <= 0 ||
		math.IsNaN(luminosityW) || math.IsInf(luminosityW, 0) ||
		math.IsNaN(temperatureK) || math.IsInf(temperatureK, 0) {
		return math.NaN()
	}
	denominator := 4 * math.Pi * StefanBoltzmannFlux(temperatureK)
	if denominator == 0 {
		return math.NaN()
	}
	return math.Sqrt(luminosityW / denominator)
}

// EffectiveTemperatureFromLuminosityRadius 由光度和半径求温度 / effective temperature from luminosity and radius.
//
//	luminosityW: 恒星总光度，单位瓦特
//	radiusM: 恒星半径，单位米
//
// 返回：
//
//	恒星有效温度，单位开尔文
func EffectiveTemperatureFromLuminosityRadius(luminosityW, radiusM float64) float64 {
	if luminosityW <= 0 || radiusM <= 0 ||
		math.IsNaN(luminosityW) || math.IsInf(luminosityW, 0) ||
		math.IsNaN(radiusM) || math.IsInf(radiusM, 0) {
		return math.NaN()
	}
	denominator := 4 * math.Pi * radiusM * radiusM * stefanBoltzmannConstant
	if denominator == 0 {
		return math.NaN()
	}
	return math.Pow(luminosityW/denominator, 0.25)
}

// LuminositySolarFromRadiusTemperature 由太阳半径单位和温度求光度 / luminosity in solar units from radius and temperature.
//
//	radiusSolar: 恒星半径，单位为太阳半径
//	temperatureK: 恒星有效温度，单位开尔文
//
// 返回：
//
//	总光度，单位为太阳光度 L☉
func LuminositySolarFromRadiusTemperature(radiusSolar, temperatureK float64) float64 {
	if radiusSolar <= 0 || temperatureK <= 0 ||
		math.IsNaN(radiusSolar) || math.IsInf(radiusSolar, 0) ||
		math.IsNaN(temperatureK) || math.IsInf(temperatureK, 0) {
		return math.NaN()
	}
	return LuminosityFromRadiusTemperature(radiusSolar*solarRadiusM, temperatureK) / solarLuminosityW
}

// RadiusSolarFromLuminosityTemperature 由太阳光度单位和温度求半径 / radius in solar units from luminosity and temperature.
//
//	luminositySolar: 恒星总光度，单位为太阳光度 L☉
//	temperatureK: 恒星有效温度，单位开尔文
//
// 返回：
//
//	恒星半径，单位为太阳半径 R☉
func RadiusSolarFromLuminosityTemperature(luminositySolar, temperatureK float64) float64 {
	if luminositySolar <= 0 || temperatureK <= 0 ||
		math.IsNaN(luminositySolar) || math.IsInf(luminositySolar, 0) ||
		math.IsNaN(temperatureK) || math.IsInf(temperatureK, 0) {
		return math.NaN()
	}
	return RadiusFromLuminosityTemperature(luminositySolar*solarLuminosityW, temperatureK) / solarRadiusM
}

// EffectiveTemperatureFromLuminositySolarRadius 由太阳光度和半径单位求温度 / effective temperature from solar luminosity and radius.
//
//	luminositySolar: 恒星总光度，单位为太阳光度 L☉
//	radiusSolar: 恒星半径，单位为太阳半径 R☉
//
// 返回：
//
//	恒星有效温度，单位开尔文
//
// 例：
//
//	// 半径 2.5 R☉、光度 20 L☉ 的主序星
//	t := formula.EffectiveTemperatureFromLuminositySolarRadius(20, 2.5)
func EffectiveTemperatureFromLuminositySolarRadius(luminositySolar, radiusSolar float64) float64 {
	if luminositySolar <= 0 || radiusSolar <= 0 ||
		math.IsNaN(luminositySolar) || math.IsInf(luminositySolar, 0) ||
		math.IsNaN(radiusSolar) || math.IsInf(radiusSolar, 0) {
		return math.NaN()
	}
	return EffectiveTemperatureFromLuminosityRadius(luminositySolar*solarLuminosityW, radiusSolar*solarRadiusM)
}

// SolarEffectiveTemperature 太阳有效温度常数 / solar effective temperature constant.
//
// 返回：
//
//	太阳有效温度，单位开尔文
func SolarEffectiveTemperature() float64 {
	return solarEffectiveTempK
}
