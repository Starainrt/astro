package basic

import (
	"math"

	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

// GetMoonCir 返回旧月球级数系数的独立副本 / an independent copy of the legacy lunar-series coefficients.
func GetMoonCir() [][][]float64 { return planet.GetMoonCir() }

func MoonCalcNew(coordIndex int, jd float64) float64 {
	return MoonCalcNewN(coordIndex, jd, -1)
}

func MoonCalcNewN(coordIndex int, jd float64, n int) float64 {
	return planet.WhereMoonN(coordIndex, jd, n)
}

func moonCalcNewN(coordIndex int, jd float64, n int) float64 {
	return planet.WhereMoonN(coordIndex, jd, n)
}

func HMoonTrueLo(jde float64) float64 { //计算月亮
	return HMoonTrueLoN(jde, -1)
}

func HMoonTrueLoN(jd float64, n int) float64 { //计算月亮
	v := moonCalcNewN(0, jd, n) * 180 / math.Pi
	return Limit360(v)
}

func HMoonTrueBo(jde float64) float64 {
	return HMoonTrueBoN(jde, -1)
}

func HMoonTrueBoN(jd float64, n int) float64 {
	v := MoonCalcNewN(1, jd, n) * 180 / math.Pi
	return v
}

func HMoonAway(jde float64) float64 { //'月地距离
	return HMoonAwayN(jde, -1)
}

func HMoonAwayN(jd float64, n int) float64 { //'月地距离
	return MoonCalcNewN(2, jd, n)
}

/*
 * @name 月球视黄经
 */
func HMoonApparentLo(jde float64) float64 {
	return HMoonApparentLoN(jde, -1)
}

func HMoonApparentLoN(jde float64, n int) float64 {
	if lo, _, _, _, ok := vectorApparent(ephem.Moon, jde, n); ok {
		return lo
	}
	return HMoonTrueLoN(jde, n) + Nutation2000Bi(jde)
}

// HMoonGeocentricApparentRaWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func HMoonGeocentricApparentRaWithInstance(jde float64, inst *Instance) float64 {
	return HMoonGeocentricApparentRaNWithInstance(jde, -1, inst)
}

// HMoonGeocentricApparentRaNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func HMoonGeocentricApparentRaNWithInstance(jde float64, n int, inst *Instance) float64 {
	return moonApparentRaN(jde, n, inst)
}

// HMoonGeocentricApparentDecWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func HMoonGeocentricApparentDecWithInstance(jde float64, inst *Instance) float64 {
	return HMoonGeocentricApparentDecNWithInstance(jde, -1, inst)
}

// HMoonGeocentricApparentDecNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func HMoonGeocentricApparentDecNWithInstance(jde float64, n int, inst *Instance) float64 {
	_, dec := moonApparentRaDecN(jde, n, inst)
	return dec
}

// HMoonGeocentricApparentRaDec 月亮地心视赤经、视赤纬 / apparent geocentric right ascension and declination of the Moon.
func HMoonGeocentricApparentRaDec(jde float64) (float64, float64) {
	return HMoonGeocentricApparentRaDecN(jde, -1)
}

// HMoonGeocentricApparentRaDecN 月亮地心视赤经、视赤纬（n 变体） / n-parameterised apparent geocentric right ascension and declination of the Moon.
func HMoonGeocentricApparentRaDecN(jde float64, n int) (float64, float64) {
	return moonApparentRaDecN(jde, n, nil)
}

// HMoonGeocentricTrueRaWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func HMoonGeocentricTrueRaWithInstance(jde float64, inst *Instance) float64 {
	return HMoonGeocentricTrueRaNWithInstance(jde, -1, inst)
}

// HMoonGeocentricTrueRaNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func HMoonGeocentricTrueRaNWithInstance(jde float64, n int, inst *Instance) float64 {
	return LoToRa(jde, HMoonTrueLoNWithInstance(jde, n, inst), HMoonTrueBoNWithInstance(jde, n, inst))
}

// HMoonGeocentricTrueDecWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func HMoonGeocentricTrueDecWithInstance(jde float64, inst *Instance) float64 {
	return HMoonGeocentricTrueDecNWithInstance(jde, -1, inst)
}

// HMoonGeocentricTrueDecNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func HMoonGeocentricTrueDecNWithInstance(jde float64, n int, inst *Instance) float64 {
	return ArcSin(Sin(HMoonTrueBoNWithInstance(jde, n, inst))*Cos(TrueObliquity(jde)) +
		Cos(HMoonTrueBoNWithInstance(jde, n, inst))*Sin(TrueObliquity(jde))*Sin(HMoonTrueLoNWithInstance(jde, n, inst)))
}

// HMoonGeocentricTrueRaDecWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func HMoonGeocentricTrueRaDecWithInstance(jde float64, inst *Instance) (float64, float64) {
	return HMoonGeocentricTrueRaDecNWithInstance(jde, -1, inst)
}

// HMoonGeocentricTrueRaDecNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func HMoonGeocentricTrueRaDecNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	return LoBoToRaDec(jde, HMoonTrueLoNWithInstance(jde, n, inst), HMoonTrueBoNWithInstance(jde, n, inst))
}

func HMoonTrueRaDec(jde float64) (float64, float64) {
	return HMoonTrueRaDecN(jde, -1)
}

// HMoonTrueRaDecNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func HMoonTrueRaDecNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	return moonApparentRaDecN(jde, n, inst)
}

/*
 * 月球真赤纬
 */
func HMoonTrueDec(jde float64) float64 {
	return HMoonTrueDecN(jde, -1)
}

// HMoonTrueDecNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func HMoonTrueDecNWithInstance(jde float64, n int, inst *Instance) float64 {
	_, dec := moonApparentRaDecN(jde, n, inst)
	return dec
}

/*
 * 月球真赤经
 */
func HMoonTrueRa(jde float64) float64 {
	return HMoonTrueRaN(jde, -1)
}

// HMoonTrueRaNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func HMoonTrueRaNWithInstance(jde float64, n int, inst *Instance) float64 {
	return moonApparentRaN(jde, n, inst)
}

/*
 * HMoonApparentRaDec 本地民用时下的月球视赤道坐标；jd 是本地民用时（内部按 jd-tz/24 换算 UT），不是世界时
 */
func HMoonApparentRaDec(jd, lon, lat, tz float64) (float64, float64) {
	return HMoonApparentRaDecN(jd, lon, lat, tz, -1)
}

func HMoonApparentRaDecN(jd, lon, lat, tz float64, n int) (float64, float64) {
	jde := UTC2TT(jd)
	ra := HMoonTrueRaN(jde-tz/24, n)
	dec := HMoonTrueDecN(jde-tz/24, n)
	away := HMoonAwayN(jde-tz/24, n) / 149597870.7
	topoRA, topoDec := TopocentricRaDec(ra, dec, lat, lon, jd-tz/24, away, 0)
	return topoRA, topoDec
}

// HMoonApparentRaWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func HMoonApparentRaWithInstance(jd, lon, lat, tz float64, inst *Instance) float64 {
	return HMoonApparentRaNWithInstance(jd, lon, lat, tz, -1, inst)
}

// HMoonApparentRaNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func HMoonApparentRaNWithInstance(jd, lon, lat, tz float64, n int, inst *Instance) float64 {
	jde := UTC2TT(jd) - tz/24
	ra := HMoonTrueRaNWithInstance(jde, n, inst)
	dec := HMoonTrueDecNWithInstance(jde, n, inst)
	away := HMoonAwayNWithInstance(jde, n, inst) / 149597870.7
	topoRA := TopocentricRa(ra, dec, lat, lon, jd-tz/24, away, 0)
	return topoRA
}

// HMoonApparentDecWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func HMoonApparentDecWithInstance(jd, lon, lat, tz float64, inst *Instance) float64 {
	return HMoonApparentDecNWithInstance(jd, lon, lat, tz, -1, inst)
}

// HMoonApparentDecNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func HMoonApparentDecNWithInstance(jd, lon, lat, tz float64, n int, inst *Instance) float64 {
	jde := UTC2TT(jd) - tz/24
	ra := HMoonTrueRaNWithInstance(jde, n, inst)
	dec := HMoonTrueDecNWithInstance(jde, n, inst)
	away := HMoonAwayNWithInstance(jde, n, inst) / 149597870.7
	topoDec := TopocentricDec(ra, dec, lat, lon, jd-tz/24, away, 0)
	return topoDec
}
