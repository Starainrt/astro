package formula

import "math"

// ExtinctionComponents V 波段天顶消光系数分量，单位星等/大气质量 / V-band zenith extinction components in magnitudes per airmass.
type ExtinctionComponents struct {
	Rayleigh float64 // 分子散射 / molecular scattering
	Aerosol  float64 // 气溶胶 / aerosol
	Ozone    float64 // 臭氧 / ozone
	Water    float64 // 水汽 / water vapour
	Total    float64 // 四项之和 / sum of the four terms
}

// ExtinctionSchaefer 按 Schaefer 经验模型由气象条件合成 V 波段天顶消光系数 / V-band zenith extinction from weather by Schaefer's empirical model.
//
// heightM 为测站海拔（米），temperatureC 为气温（°C），relativeHumidity 为相对湿度（百分比，
// 须在 (0, 100) 内），latitude 为测站纬度（度），sunRA 为太阳赤经（度，用于季节项）。
// Total 可直接作为 basic.HeliacalOptions.ExtinctionCoefficient；输入越界时各分量为 NaN。
// heightM is site elevation in metres, temperatureC is air temperature,
// relativeHumidity is a percentage in (0, 100), latitude is the site latitude,
// and sunRA is the Sun's right ascension in degrees for the seasonal terms.
// Total can be used as basic.HeliacalOptions.ExtinctionCoefficient. Out-of-range
// inputs yield NaN components.
func ExtinctionSchaefer(heightM, temperatureC, relativeHumidity, latitude, sunRA float64) ExtinctionComponents {
	if !isFinitePhotometry(heightM) || !isFinitePhotometry(temperatureC) || !isFinitePhotometry(relativeHumidity) ||
		!isFinitePhotometry(latitude) || !isFinitePhotometry(sunRA) ||
		relativeHumidity <= 0 || relativeHumidity >= 100 || latitude < -90 || latitude > 90 {
		n := math.NaN()
		return ExtinctionComponents{n, n, n, n, n}
	}
	lat := latitude * math.Pi / 180
	ra := sunRA * math.Pi / 180
	hemisphere := 1.0
	if latitude < 0 {
		hemisphere = -1
	}
	// 标高：分子与水汽 8200 m，气溶胶 1500 m；臭氧层在高空，不随测站海拔变化。
	components := ExtinctionComponents{
		Rayleigh: 0.1066 * math.Exp(-heightM/8200),
		Aerosol: 0.1 * math.Exp(-heightM/1500) * math.Pow(1-0.32/math.Log(relativeHumidity/100), 1.33) *
			(1 + 0.33*hemisphere*math.Sin(ra)),
		Ozone: 0.031 * (3 + 0.4*(lat*math.Cos(ra)-math.Cos(3*lat))) / 3,
		Water: 0.031 * 0.94 * (relativeHumidity / 100) * math.Exp(temperatureC/15) * math.Exp(-heightM/8200),
	}
	components.Total = components.Rayleigh + components.Aerosol + components.Ozone + components.Water
	return components
}
