package planet

import . "github.com/starainrt/astro/tools"

// SunLo 太阳几何黄经
func SunLo(jd float64) float64 {
	T := (jd - 2451545) / 365250
	SunLo := 280.4664567 + 360007.6982779*T + 0.03032028*T*T + T*T*T/49931 - T*T*T*T/15299 - T*T*T*T*T/1988000
	return Limit360(SunLo)
}

func SunM(JD float64) float64 {
	T := (JD - 2451545) / 36525
	sunM := 357.5291092 + 35999.0502909*T - 0.0001559*T*T - 0.00000048*T*T*T
	return Limit360(sunM)
}

// Earthe 地球偏心率
func Earthe(JD float64) float64 {
	T := (JD - 2451545) / 36525
	Earthe := 0.016708617 - 0.000042037*T - 0.0000001236*T*T
	return Earthe
}

func EarthPI(JD float64) float64 {
	T := (JD - 2451545) / 36525
	return 102.93735 + 1.71953*T + 000046*T*T
}

func SunMidFun(JD float64) float64 {
	T := (JD - 2451545) / 36525
	M := SunM(JD)
	SunMidFun := (1.9146-0.004817*T-0.000014*T*T)*Sin(M) + (0.019993-0.000101*T)*Sin(2*M) + 0.00029*Sin(3*M)
	return SunMidFun
}

func SunTrueLo(JD float64) float64 {
	SunTrueLo := SunLo(JD) + SunMidFun(JD)
	return SunTrueLo
}

func SunApparentLo(JD float64) float64 {
	T := (JD - 2451545) / 36525
	SunApparentLo := SunTrueLo(JD) - 0.00569 - 0.00478*Sin(125.04-1934.136*T)
	return SunApparentLo
}

func Distance(jd float64) float64 {
	f := SunMidFun(jd)
	m := SunM(jd)
	e := Earthe(jd)
	return 1.000001018 * (1 - e*e) / (1 + e*Cos(f+m))
}
