package formula

import "math"

// AirmassPlaneParallel 平行平板大气模型 / plane-parallel airmass.
//
// altitude 为真高度角，单位度。该模型等价于 sec(z)，其中 z 为天顶距。
// 中高空可作几何近似；接近地平线时会发散，不宜用于低空精细估算。
// altitude is true altitude in degrees. The model is sec(z) with zenith
// distance z = 90° - altitude. It is a geometric approximation that diverges
// near the horizon.
func AirmassPlaneParallel(altitude float64) float64 {
	if !isFinitePhotometry(altitude) || altitude < 0 || altitude > 90 {
		return math.NaN()
	}
	if altitude == 0 {
		return math.Inf(1)
	}
	return 1 / math.Sin(altitude*math.Pi/180)
}

// AirmassPlaneParallelByZenithDistance 按天顶距的平行平板大气质量 / plane-parallel airmass by zenith distance.
func AirmassPlaneParallelByZenithDistance(zenithDistance float64) float64 {
	if !isFinitePhotometry(zenithDistance) || zenithDistance < 0 || zenithDistance > 90 {
		return math.NaN()
	}
	if zenithDistance == 90 {
		return math.Inf(1)
	}
	return 1 / math.Cos(zenithDistance*math.Pi/180)
}

// AirmassKastenYoung Kasten-Young 1989 大气质量模型 / Kasten-Young 1989 airmass.
//
// apparentAltitude 为视高度角，单位度。该经验公式在低空通常比 sec(z) 更稳健。
// apparentAltitude is apparent altitude in degrees. This empirical model is
// generally more robust than sec(z) at low altitude.
func AirmassKastenYoung(apparentAltitude float64) float64 {
	if !isFinitePhotometry(apparentAltitude) || apparentAltitude < 0 || apparentAltitude > 90 {
		return math.NaN()
	}
	return 1 / (math.Sin(apparentAltitude*math.Pi/180) + 0.50572*math.Pow(apparentAltitude+6.07995, -1.6364))
}

// AirmassPickering Pickering 2002 大气质量模型 / Pickering 2002 airmass.
//
// apparentAltitude 为视高度角，单位度。该经验公式专门面向低空观测修正。
// apparentAltitude is apparent altitude in degrees. This empirical model is
// intended for low-altitude observational use.
func AirmassPickering(apparentAltitude float64) float64 {
	if !isFinitePhotometry(apparentAltitude) || apparentAltitude < 0 || apparentAltitude > 90 {
		return math.NaN()
	}
	correctedAltitude := apparentAltitude + 244/(165+47*math.Pow(apparentAltitude, 1.1))
	return 1 / math.Sin(correctedAltitude*math.Pi/180)
}

func isFinitePhotometry(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
