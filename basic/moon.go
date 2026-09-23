package basic

import (
	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func MoonLo(jde float64) float64 { //'月球平黄经
	return planet.MoonLo(jde)
}

func SunMoonAngle(jde float64) float64 { // '月日距角
	return planet.SunMoonAngle(jde)
}

func MoonM(jde float64) float64 { // '月平近点角
	return planet.MoonM(jde)
}

func MoonLonX(jde float64) float64 { // As Double '月球经度参数(到升交点的平角距离)
	return planet.MoonLonX(jde)
}

func MoonI(jde float64) float64 {
	return planet.MoonI(jde)
}

func MoonR(jde float64) float64 {
	return planet.MoonR(jde)
}

func MoonB(jde float64) float64 {
	return planet.MoonB(jde)
}

func MoonTrueLo(jde float64) float64 {
	return planet.MoonTrueLo(jde)
}
func MoonTrueBo(jde float64) float64 {
	return planet.MoonTrueBo(jde)
}
func MoonAway(jde float64) float64 { //'月地距离
	return planet.MoonAway(jde)
}

/*
 * @name 月球视黄经
 */
func MoonApparentLo(jde float64) float64 {
	return MoonTrueLo(jde) + Nutation2000Bi(jde)
}

/*
 * 月球真赤纬
 */
func MoonTrueDec(jde float64) float64 {
	moonLo := MoonApparentLo(jde)
	moonBo := MoonTrueBo(jde)
	tmp := Sin(moonBo)*Cos(TrueObliquity(jde)) + Cos(moonBo)*Sin(TrueObliquity(jde))*Sin(moonLo)
	res := ArcSin(tmp)
	return res
}

/*
 * 月球真赤经
 */
func MoonTrueRa(jde float64) float64 {
	return LoToRa(jde, MoonApparentLo(jde), MoonTrueBo(jde))
}

func MoonTrueRaDec(jde float64) (float64, float64) {
	return LoBoToRaDec(jde, MoonApparentLo(jde), MoonTrueBo(jde))
}

// MoonApparentRa 站心视赤经；jd 为当地时儒略日，tz 为时区小时数 / topocentric apparent right ascension; jd is local civil time and tz is the zone offset in hours.
func MoonApparentRa(jd, lon, lat float64, tz int) float64 {
	// 本地时刻加 ΔT 再减时区偏移即该地时刻的 TT（加法可交换），故 jde 就是 TT。
	jde := UTC2TT(jd) - float64(tz)/24.000
	ra := MoonTrueRa(jde)
	dec := MoonTrueDec(jde)
	away := MoonAway(jde) / 149597870.7
	topoRA := TopocentricRa(ra, dec, lat, lon, jd-float64(tz)/24.000, away, 0)
	return topoRA
}

func MoonApparentDec(jd, lon, lat, tz float64) float64 {
	// 同上：jde 是当地时刻的 TT。
	jde := UTC2TT(jd) - tz/24
	ra := MoonTrueRa(jde)
	dec := MoonTrueDec(jde)
	away := MoonAway(jde) / 149597870.7
	topoDec := TopocentricDec(ra, dec, lat, lon, jd-tz/24, away, 0)
	return topoDec
}
