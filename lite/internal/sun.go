package internal

import . "github.com/starainrt/astro/tools"

func SunLo(jde float64) float64 {
	t := (jde - 2451545.0) / 365250.0
	return Limit360(280.4664567 + 360007.6982779*t + 0.03032028*t*t + t*t*t/49931.0 - t*t*t*t/15299.0 - t*t*t*t*t/1988000.0)
}

func SunMeanAnomaly(jde float64) float64 {
	t := (jde - 2451545.0) / 36525.0
	return Limit360(357.5291092 + 35999.0502909*t - 0.0001559*t*t - 0.00000048*t*t*t)
}

func EarthEccentricity(jde float64) float64 {
	t := (jde - 2451545.0) / 36525.0
	return 0.016708617 - 0.000042037*t - 0.0000001236*t*t
}

func SunCenter(jde float64) float64 {
	t := (jde - 2451545.0) / 36525.0
	m := SunMeanAnomaly(jde)
	return (1.9146-0.004817*t-0.000014*t*t)*Sin(m) + (0.019993-0.000101*t)*Sin(2*m) + 0.00029*Sin(3*m)
}

// 本文件以 jde（力学时儒略日）键控；lite 包装层的实参是 UTC 民用 JD，69 s 的历元差远低于轻量公式精度。
func SunTrueLo(jde float64) float64 {
	return Limit360(SunLo(jde) + SunCenter(jde))
}

func SunApparentLo(jde float64) float64 {
	t := (jde - 2451545.0) / 36525.0
	return Limit360(SunTrueLo(jde) - 0.00569 - 0.00478*Sin(125.04-1934.136*t))
}

func SunDistanceAU(jde float64) float64 {
	c := SunCenter(jde)
	m := SunMeanAnomaly(jde)
	e := EarthEccentricity(jde)
	return 1.000001018 * (1 - e*e) / (1 + e*Cos(m+c))
}

func SunTrueRaDec(jde float64) (float64, float64) {
	return EclipticToEquatorial(jde, SunTrueLo(jde), 0)
}

func SunApparentRaDec(jde float64) (float64, float64) {
	return EclipticToEquatorial(jde, SunApparentLo(jde), 0)
}
