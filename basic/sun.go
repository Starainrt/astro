package basic

import (
	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

// SunLo 太阳几何黄经
func SunLo(jde float64) float64 {
	return planet.SunLo(jde)
}

func SunM(jde float64) float64 {
	return planet.SunM(jde)
}

/*
@name 地球偏心率
*/
func Earthe(jde float64) float64 { //'地球偏心率
	return planet.Earthe(jde)
}

func EarthPI(jde float64) float64 { //近日點經度
	return planet.EarthPI(jde)
}

func SunMidFun(jde float64) float64 { //'太阳中间方程
	return planet.SunMidFun(jde)
}

func SunTrueLo(jde float64) float64 { // '太阳真黄经
	return planet.SunTrueLo(jde)
}

func SunApparentLo(jde float64) float64 { //'太阳视黄经
	return planet.SunApparentLo(jde)
}

func SunApparentRa(jde float64) float64 { // '太阳视赤经
	return LoToRa(jde, SunApparentLo(jde), 0)
}

func SunApparentRaDec(jde float64) (float64, float64) {
	return LoBoToRaDec(jde, SunApparentLo(jde), 0)
}

func SunTrueRa(jde float64) float64 { //'太阳真赤经
	eps := TrueObliquity(jde)
	sunTrueRa := ArcTan(Cos(eps) * Sin(SunTrueLo(jde)) / Cos(SunTrueLo(jde)))
	//Select Case SunTrueLo(JD)
	sunTrueLo := SunTrueLo(jde)
	if sunTrueLo >= 90 && sunTrueLo < 180 {
		sunTrueRa = 180 + sunTrueRa
	} else if sunTrueLo >= 180 && sunTrueLo < 270 {
		sunTrueRa = 180 + sunTrueRa
	} else if sunTrueLo >= 270 && sunTrueLo <= 360 {
		sunTrueRa = 360 + sunTrueRa
	}
	return sunTrueRa
}

func SunApparentDec(jde float64) float64 { // '太阳视赤纬
	// TrueObliquity 已是“平交角 + 交角章动”，不能再加 0.00256*cos(Ω)：
	// 那一项就是交角章动的近似值（见《天文算法》24 章译者注），加了会把章动计两次。
	eps := TrueObliquity(jde)
	sunApparentDec := ArcSin(Sin(eps) * Sin(SunApparentLo(jde)))
	return sunApparentDec
}

func SunTrueDec(jde float64) float64 { // '太阳真赤纬
	eps := TrueObliquity(jde)
	sunTrueDec := ArcSin(Sin(eps) * Sin(SunTrueLo(jde)))
	return sunTrueDec
}

// SunTime 均时差，单位小时 / equation of time in hours.
//
// jde 按力学时取用；CulminationTime 传的是本地民用日锚点，历元差最大半天，
// 对均时差的影响在 10 s 量级，是该近似入口的既有取舍。
func SunTime(jde float64) float64 { //均时差
	tm := (SunLo(jde) - 0.0057183 - (HSunApparentRa(jde)) + (Nutation2000Bi(jde))*Cos(TrueObliquity(jde))) / 15
	if tm > 23 {
		tm = -24 + tm
	}
	return tm
}

func SunTimeN(jde float64, n int) float64 { //均时差
	tm := (SunLo(jde) - 0.0057183 - (HSunApparentRaN(jde, n)) + (Nutation2000Bi(jde))*Cos(TrueObliquity(jde))) / 15
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
func HSunTrueLo(jde float64) float64 {
	return HSunTrueLoN(jde, -1)
}

// HSunTrueLoN 高精度太阳真黄经，n<0 时取全量 VSOP 项。
func HSunTrueLoN(jde float64, n int) float64 {
	return planet.WherePlanetN(0, 0, jde, n)
}

func HSunTrueBo(jde float64) float64 {
	return HSunTrueBoN(jde, -1)
}

// HSunTrueBoN 高精度太阳真黄纬，n<0 时取全量 VSOP 项。
func HSunTrueBoN(jde float64, n int) float64 {
	return planet.WherePlanetN(0, 1, jde, n)
}

func HSunApparentLo(jde float64) float64 {
	return HSunApparentLoN(jde, -1)
}

func HSunApparentLoN(jde float64, n int) float64 {
	if lo, _, _, _, ok := vectorApparent(ephem.Sun, jde, n); ok {
		return lo
	}
	lo := HSunTrueLoN(jde, n)
	lo = lo + Nutation2000Bi(jde) + SunLoGXCN(jde, n)
	return lo
}

func SunLoGXC(jde float64) float64 {
	return SunLoGXCN(jde, -1)
}

func SunLoGXCN(jde float64, n int) float64 {
	radius := EarthAwayN(jde, n)
	return -20.49552 / radius / 3600
}

func EarthAway(jde float64) float64 {
	return EarthAwayN(jde, -1)
}

func EarthAwayN(jde float64, n int) float64 {
	return planet.WherePlanetN(0, 2, jde, n)
}

func HSunApparentRaDec(jde float64) (float64, float64) {
	return HSunApparentRaDecN(jde, -1)
}

func HSunApparentRaDecN(jde float64, n int) (float64, float64) {
	ra, dec, _ := hSunApparentRaDecDistanceN(jde, n)
	return ra, dec
}

func hSunApparentRaDecDistanceN(jde float64, n int, instances ...*Instance) (ra, dec, distanceAU float64) {
	inst := instanceValue(instances)

	ra, dec = sunApparentRaDecN(jde, n, inst)
	return ra, dec, EarthAwayNWithInstance(jde, n, inst)
}

func HSunApparentRa(jde float64) float64 { // '太阳视赤经
	return HSunApparentRaN(jde, -1)
}

func HSunApparentRaN(jde float64, n int) float64 { // '太阳视赤经
	return sunApparentRaN(jde, n, nil)
}

func HSunTrueRa(jde float64) float64 {
	return HSunTrueRaN(jde, -1)
}

func HSunTrueRaN(jde float64, n int) float64 {
	sunTrueLo := HSunTrueLoN(jde, n)
	eps := TrueObliquity(jde)

	numerator := Cos(eps) * Sin(sunTrueLo)
	denominator := Cos(sunTrueLo)

	return ArcTan2(numerator, denominator)
}

func HSunApparentDec(jde float64) float64 { // '太阳视赤纬
	return HSunApparentDecN(jde, -1)
}

func HSunApparentDecN(jde float64, n int) float64 { // '太阳视赤纬
	_, dec := sunApparentRaDecN(jde, n, nil)
	return dec
}

func HSunTrueDec(jde float64) float64 { // '太阳真赤纬
	return HSunTrueDecN(jde, -1)
}

func HSunTrueDecN(jde float64, n int) float64 { // '太阳真赤纬
	return ArcSin(Sin(EclipticObliquity(jde, false)) * Sin(HSunTrueLoN(jde, n)))
}

func Distance(jde float64) float64 { //ri di ju li
	return planet.Distance(jde)
}
