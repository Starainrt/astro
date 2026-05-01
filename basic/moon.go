package basic

import (
	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func MoonLo(jd float64) float64 { //'月球平黄经
	return planet.MoonLo(jd)
}

func SunMoonAngle(jd float64) float64 { // '月日距角
	return planet.SunMoonAngle(jd)
}

func MoonM(jd float64) float64 { // '月平近点角
	return planet.MoonM(jd)
}

func MoonLonX(jd float64) float64 { // As Double '月球经度参数(到升交点的平角距离)
	return planet.MoonLonX(jd)
}

func MoonI(jd float64) float64 {
	return planet.MoonI(jd)
}

func MoonR(jd float64) float64 {
	return planet.MoonR(jd)
}

func MoonB(jd float64) float64 {
	return planet.MoonB(jd)
}

func MoonTrueLo(jd float64) float64 {
	return planet.MoonTrueLo(jd)
}
func MoonTrueBo(jd float64) float64 {
	return planet.MoonTrueBo(jd)
}
func MoonAway(jd float64) float64 { //'月地距离
	return planet.MoonAway(jd)
}

/*
 * @name 月球视黄经
 */
func MoonApparentLo(jd float64) float64 {
	return MoonTrueLo(jd) + Nutation2000Bi(jd)
}

/*
 * 月球真赤纬
 */
func MoonTrueDec(jd float64) float64 {
	moonLo := MoonApparentLo(jd)
	moonBo := MoonTrueBo(jd)
	tmp := Sin(moonBo)*Cos(TrueObliquity(jd)) + Cos(moonBo)*Sin(TrueObliquity(jd))*Sin(moonLo)
	res := ArcSin(tmp)
	return res
}

/*
 * 月球真赤经
 */
func MoonTrueRa(jd float64) float64 {
	return LoToRa(jd, MoonApparentLo(jd), MoonTrueBo(jd))
}

func MoonTrueRaDec(jd float64) (float64, float64) {
	return LoBoToRaDec(jd, MoonApparentLo(jd), MoonTrueBo(jd))
}

/*
*

	*
	传入世界时
*/
func MoonApparentRa(jd, lon, lat float64, tz int) float64 {
	jde := TD2UT(jd, true)
	utcJD := jde - float64(tz)/24.000
	ra := MoonTrueRa(utcJD)
	dec := MoonTrueDec(utcJD)
	away := MoonAway(utcJD) / 149597870.7
	topoRA := TopocentricRa(ra, dec, lat, lon, jd-float64(tz)/24.000, away, 0)
	return topoRA
}

func MoonApparentDec(jd, lon, lat, tz float64) float64 {
	jde := TD2UT(jd, true)
	utcJD := jde - tz/24
	ra := MoonTrueRa(utcJD)
	dec := MoonTrueDec(utcJD)
	away := MoonAway(utcJD) / 149597870.7
	topoDec := TopocentricDec(ra, dec, lat, lon, jd-tz/24, away, 0)
	return topoDec
}
