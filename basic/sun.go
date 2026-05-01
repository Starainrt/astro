package basic

import (
	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

// SunLo 太阳几何黄经
func SunLo(jd float64) float64 {
	return planet.SunLo(jd)
}

func SunM(jd float64) float64 {
	return planet.SunM(jd)
}

/*
@name 地球偏心率
*/
func Earthe(jd float64) float64 { //'地球偏心率
	return planet.Earthe(jd)
}

func EarthPI(jd float64) float64 { //近日點經度
	return planet.EarthPI(jd)
}

func SunMidFun(jd float64) float64 { //'太阳中间方程
	return planet.SunMidFun(jd)
}

func SunTrueLo(jd float64) float64 { // '太阳真黄经
	return planet.SunTrueLo(jd)
}

func SunApparentLo(jd float64) float64 { //'太阳视黄经
	return planet.SunApparentLo(jd)
}

func SunApparentRa(jd float64) float64 { // '太阳视赤经
	return LoToRa(jd, SunApparentLo(jd), 0)
}

func SunApparentRaDec(jd float64) (float64, float64) {
	return LoBoToRaDec(jd, SunApparentLo(jd), 0)
}

func SunTrueRa(jd float64) float64 { //'太阳真赤经
	eps := TrueObliquity(jd)
	sunTrueRa := ArcTan(Cos(eps) * Sin(SunTrueLo(jd)) / Cos(SunTrueLo(jd)))
	//Select Case SunTrueLo(JD)
	sunTrueLo := SunTrueLo(jd)
	if sunTrueLo >= 90 && sunTrueLo < 180 {
		sunTrueRa = 180 + sunTrueRa
	} else if sunTrueLo >= 180 && sunTrueLo < 270 {
		sunTrueRa = 180 + sunTrueRa
	} else if sunTrueLo >= 270 && sunTrueLo <= 360 {
		sunTrueRa = 360 + sunTrueRa
	}
	return sunTrueRa
}

func SunApparentDec(jd float64) float64 { // '太阳视赤纬
	t := (jd - 2451545) / 36525
	eps := TrueObliquity(jd) + 0.00256*Cos(125.04-1934.136*t)
	sunApparentDec := ArcSin(Sin(eps) * Sin(SunApparentLo(jd)))
	return sunApparentDec
}

func SunTrueDec(jd float64) float64 { // '太阳真赤纬
	eps := TrueObliquity(jd)
	sunTrueDec := ArcSin(Sin(eps) * Sin(SunTrueLo(jd)))
	return sunTrueDec
}

func SunTime(jd float64) float64 { //均时差
	tm := (SunLo(jd) - 0.0057183 - (HSunApparentRa(jd)) + (Nutation2000Bi(jd))*Cos(TrueObliquity(jd))) / 15
	if tm > 23 {
		tm = -24 + tm
	}
	return tm
}

func SunTimeN(jd float64, n int) float64 { //均时差
	tm := (SunLo(jd) - 0.0057183 - (HSunApparentRaN(jd, n)) + (Nutation2000Bi(jd))*Cos(TrueObliquity(jd))) / 15
	if tm > 23 {
		tm = -24 + tm
	}
	return tm
}

func SunSC(lo, jd float64) float64 { //黄道上的岁差，仅黄纬=0时
	t := (jd - 2451545) / 36525
	//n := 47.0029/3600*t - 0.03302/3600*t*t + 0.000060/3600*t*t*t
	//m := 174.876384/3600 - 869.8089/3600*t + 0.03536/3600*t*t
	pk := 5029.0966/3600.00*t + 1.11113/3600.00*t*t - 0.000006/3600.00*t*t*t
	return lo + pk
}

// 高精度，使用VSOP87
func HSunTrueLo(jd float64) float64 {
	return HSunTrueLoN(jd, -1)
}

// HSunTrueLoN 高精度太阳真黄经，n<0 时取全量 VSOP 项。
func HSunTrueLoN(jd float64, n int) float64 {
	return planet.WherePlanetN(0, 0, jd, n)
}

func HSunTrueBo(jd float64) float64 {
	return HSunTrueBoN(jd, -1)
}

// HSunTrueBoN 高精度太阳真黄纬，n<0 时取全量 VSOP 项。
func HSunTrueBoN(jd float64, n int) float64 {
	return planet.WherePlanetN(0, 1, jd, n)
}

func HSunApparentLo(jd float64) float64 {
	return HSunApparentLoN(jd, -1)
}

func HSunApparentLoN(jd float64, n int) float64 {
	lo := HSunTrueLoN(jd, n)
	lo = lo + Nutation2000Bi(jd) + SunLoGXCN(jd, n)
	return lo
}

func SunLoGXC(jd float64) float64 {
	return SunLoGXCN(jd, -1)
}

func SunLoGXCN(jd float64, n int) float64 {
	radius := EarthAwayN(jd, n)
	return -20.49552 / radius / 3600
}

func EarthAway(jd float64) float64 {
	return EarthAwayN(jd, -1)
}

func EarthAwayN(jd float64, n int) float64 {
	return planet.WherePlanetN(0, 2, jd, n)
}

func HSunApparentRaDec(jd float64) (float64, float64) {
	return HSunApparentRaDecN(jd, -1)
}

func HSunApparentRaDecN(jd float64, n int) (float64, float64) {
	return LoBoToRaDec(jd, HSunApparentLoN(jd, n), HSunTrueBoN(jd, n))
}

func HSunApparentRa(jd float64) float64 { // '太阳视赤经
	return HSunApparentRaN(jd, -1)
}

func HSunApparentRaN(jd float64, n int) float64 { // '太阳视赤经
	return LoToRa(jd, HSunApparentLoN(jd, n), HSunTrueBoN(jd, n))
}

func HSunTrueRa(jd float64) float64 {
	return HSunTrueRaN(jd, -1)
}

func HSunTrueRaN(jd float64, n int) float64 {
	sunTrueLo := HSunTrueLoN(jd, n)
	eps := TrueObliquity(jd)

	numerator := Cos(eps) * Sin(sunTrueLo)
	denominator := Cos(sunTrueLo)

	return ArcTan2(numerator, denominator)
}

func HSunApparentDec(jd float64) float64 { // '太阳视赤纬
	return HSunApparentDecN(jd, -1)
}

func HSunApparentDecN(jd float64, n int) float64 { // '太阳视赤纬
	return ArcSin(Sin(EclipticObliquity(jd, true)) * Sin(HSunApparentLoN(jd, n)))
}

func HSunTrueDec(jd float64) float64 { // '太阳真赤纬
	return HSunTrueDecN(jd, -1)
}

func HSunTrueDecN(jd float64, n int) float64 { // '太阳真赤纬
	return ArcSin(Sin(EclipticObliquity(jd, false)) * Sin(HSunTrueLoN(jd, n)))
}

func Distance(jd float64) float64 { //ri di ju li
	return planet.Distance(jd)
}
