package basic

import (
	"github.com/starainrt/astro/internal/ephem"
	. "github.com/starainrt/astro/tools"
)

// sunLowPrecisionAlias 报告低精度太阳别名是否沿用与模型无关的解析公式，只有自定义引擎改走实例。
func sunLowPrecisionAlias(inst *Instance) bool {
	return inst == nil || inst.Model() != ephem.Custom
}

// SunTrueLoWithInstance 太阳真黄经 / true solar longitude with an explicit ephemeris instance.
func SunTrueLoWithInstance(jde float64, inst *Instance) float64 {
	if sunLowPrecisionAlias(inst) {
		return SunTrueLo(jde)
	}
	return HSunTrueLoWithInstance(jde, inst)
}

// SunTrueLoNWithInstance 太阳真黄经（n 变体） / n-parameterised true solar longitude with an explicit ephemeris instance.
func SunTrueLoNWithInstance(jde float64, n int, inst *Instance) float64 {
	return HSunTrueLoNWithInstance(jde, n, inst)
}

// SunTrueBoWithInstance 太阳真黄纬 / true solar latitude with an explicit ephemeris instance.
func SunTrueBoWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return HSunTrueBo(jde)
	}
	return HSunTrueBoWithInstance(jde, inst)
}

// SunTrueBoNWithInstance 太阳真黄纬（n 变体） / n-parameterised true solar latitude with an explicit ephemeris instance.
func SunTrueBoNWithInstance(jde float64, n int, inst *Instance) float64 {
	return HSunTrueBoNWithInstance(jde, n, inst)
}

// SunApparentLoWithInstance 太阳视黄经 / apparent solar longitude with an explicit ephemeris instance.
func SunApparentLoWithInstance(jde float64, inst *Instance) float64 {
	if sunLowPrecisionAlias(inst) {
		return SunApparentLo(jde)
	}
	return HSunApparentLoWithInstance(jde, inst)
}

// SunApparentLoNWithInstance 太阳视黄经（n 变体） / n-parameterised apparent solar longitude with an explicit ephemeris instance.
func SunApparentLoNWithInstance(jde float64, n int, inst *Instance) float64 {
	return HSunApparentLoNWithInstance(jde, n, inst)
}

// SunApparentRaWithInstance 太阳视赤经 / apparent solar right ascension with an explicit ephemeris instance.
func SunApparentRaWithInstance(jde float64, inst *Instance) float64 {
	if sunLowPrecisionAlias(inst) {
		return SunApparentRa(jde)
	}
	return HSunApparentRaWithInstance(jde, inst)
}

// SunApparentRaNWithInstance 太阳视赤经（n 变体） / n-parameterised apparent solar right ascension with an explicit ephemeris instance.
func SunApparentRaNWithInstance(jde float64, n int, inst *Instance) float64 {
	return HSunApparentRaNWithInstance(jde, n, inst)
}

// SunApparentDecWithInstance 太阳视赤纬 / apparent solar declination with an explicit ephemeris instance.
func SunApparentDecWithInstance(jde float64, inst *Instance) float64 {
	if sunLowPrecisionAlias(inst) {
		return SunApparentDec(jde)
	}
	return HSunApparentDecWithInstance(jde, inst)
}

// SunApparentDecNWithInstance 太阳视赤纬（n 变体） / n-parameterised apparent solar declination with an explicit ephemeris instance.
func SunApparentDecNWithInstance(jde float64, n int, inst *Instance) float64 {
	return HSunApparentDecNWithInstance(jde, n, inst)
}

// SunApparentRaDecWithInstance 太阳视赤经纬 / apparent solar right ascension and declination with an explicit ephemeris instance.
func SunApparentRaDecWithInstance(jde float64, inst *Instance) (float64, float64) {
	if sunLowPrecisionAlias(inst) {
		return SunApparentRaDec(jde)
	}
	return HSunApparentRaDecWithInstance(jde, inst)
}

// SunApparentRaDecNWithInstance 太阳视赤经纬（n 变体） / n-parameterised apparent solar right ascension and declination with an explicit ephemeris instance.
func SunApparentRaDecNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	return HSunApparentRaDecNWithInstance(jde, n, inst)
}

// SunTrueRaWithInstance 太阳真赤经 / true solar right ascension with an explicit ephemeris instance.
func SunTrueRaWithInstance(jde float64, inst *Instance) float64 {
	if sunLowPrecisionAlias(inst) {
		return SunTrueRa(jde)
	}
	return HSunTrueRaWithInstance(jde, inst)
}

// SunTrueRaNWithInstance 太阳真赤经（n 变体） / n-parameterised true solar right ascension with an explicit ephemeris instance.
func SunTrueRaNWithInstance(jde float64, n int, inst *Instance) float64 {
	return HSunTrueRaNWithInstance(jde, n, inst)
}

// SunTrueDecWithInstance 太阳真赤纬 / true solar declination with an explicit ephemeris instance.
func SunTrueDecWithInstance(jde float64, inst *Instance) float64 {
	if sunLowPrecisionAlias(inst) {
		return SunTrueDec(jde)
	}
	return HSunTrueDecWithInstance(jde, inst)
}

// SunTrueDecNWithInstance 太阳真赤纬（n 变体） / n-parameterised true solar declination with an explicit ephemeris instance.
func SunTrueDecNWithInstance(jde float64, n int, inst *Instance) float64 {
	return HSunTrueDecNWithInstance(jde, n, inst)
}

// SunTrueRaDecWithInstance 太阳真赤经纬 / true solar right ascension and declination with an explicit ephemeris instance.
func SunTrueRaDecWithInstance(jde float64, inst *Instance) (float64, float64) {
	if sunLowPrecisionAlias(inst) {
		return SunTrueRa(jde), SunTrueDec(jde)
	}
	return HSunTrueRaWithInstance(jde, inst), HSunTrueDecWithInstance(jde, inst)
}

// SunTrueRaDecNWithInstance 太阳真赤经纬（n 变体） / n-parameterised true solar right ascension and declination with an explicit ephemeris instance.
func SunTrueRaDecNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	return HSunTrueRaNWithInstance(jde, n, inst), HSunTrueDecNWithInstance(jde, n, inst)
}

// EarthAwayWithInstance 日地距离 / Earth-Sun distance with an explicit ephemeris instance.
func EarthAwayWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return EarthAway(jde)
	}
	return EarthAwayNWithInstance(jde, -1, inst)
}

// MoonGeocentricApparentRaWithInstance 月球地心视赤经 / apparent geocentric lunar right ascension with an explicit ephemeris instance.
func MoonGeocentricApparentRaWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return HMoonGeocentricApparentRa(jde)
	}
	return MoonGeocentricApparentRaNWithInstance(jde, -1, inst)
}

// MoonGeocentricApparentRaNWithInstance 月球地心视赤经（n 变体） / n-parameterised apparent geocentric lunar right ascension with an explicit ephemeris instance.
func MoonGeocentricApparentRaNWithInstance(jde float64, n int, inst *Instance) float64 {
	return moonApparentRaN(jde, n, inst)
}

// MoonGeocentricApparentDecWithInstance 月球地心视赤纬 / apparent geocentric lunar declination with an explicit ephemeris instance.
func MoonGeocentricApparentDecWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return HMoonGeocentricApparentDec(jde)
	}
	return MoonGeocentricApparentDecNWithInstance(jde, -1, inst)
}

// MoonGeocentricApparentDecNWithInstance 月球地心视赤纬（n 变体） / n-parameterised apparent geocentric lunar declination with an explicit ephemeris instance.
func MoonGeocentricApparentDecNWithInstance(jde float64, n int, inst *Instance) float64 {
	bo := HMoonTrueBoNWithInstance(jde, n, inst)
	lo := HMoonApparentLoNWithInstance(jde, n, inst)
	return ArcSin(Sin(bo)*Cos(TrueObliquity(jde)) + Cos(bo)*Sin(TrueObliquity(jde))*Sin(lo))
}

// MoonGeocentricApparentRaDecWithInstance 月球地心视赤经纬 / apparent geocentric lunar right ascension and declination with an explicit ephemeris instance.
func MoonGeocentricApparentRaDecWithInstance(jde float64, inst *Instance) (float64, float64) {
	if inst == nil {
		return HMoonGeocentricApparentRaDec(jde)
	}
	return MoonGeocentricApparentRaNWithInstance(jde, -1, inst), MoonGeocentricApparentDecNWithInstance(jde, -1, inst)
}

// MoonGeocentricApparentRaDecNWithInstance 月球地心视赤经纬（n 变体） / n-parameterised apparent geocentric lunar right ascension and declination with an explicit ephemeris instance.
func MoonGeocentricApparentRaDecNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	return MoonGeocentricApparentRaNWithInstance(jde, n, inst), MoonGeocentricApparentDecNWithInstance(jde, n, inst)
}

// MoonGeocentricTrueRaWithInstance 月球地心真赤经 / true geocentric lunar right ascension with an explicit ephemeris instance.
func MoonGeocentricTrueRaWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return HMoonGeocentricTrueRa(jde)
	}
	return MoonGeocentricTrueRaNWithInstance(jde, -1, inst)
}

// MoonGeocentricTrueRaNWithInstance 月球地心真赤经（n 变体） / n-parameterised true geocentric lunar right ascension with an explicit ephemeris instance.
func MoonGeocentricTrueRaNWithInstance(jde float64, n int, inst *Instance) float64 {
	return LoToRa(jde, HMoonTrueLoNWithInstance(jde, n, inst), HMoonTrueBoNWithInstance(jde, n, inst))
}

// MoonGeocentricTrueDecWithInstance 月球地心真赤纬 / true geocentric lunar declination with an explicit ephemeris instance.
func MoonGeocentricTrueDecWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return HMoonGeocentricTrueDec(jde)
	}
	return MoonGeocentricTrueDecNWithInstance(jde, -1, inst)
}

// MoonGeocentricTrueDecNWithInstance 月球地心真赤纬（n 变体） / n-parameterised true geocentric lunar declination with an explicit ephemeris instance.
func MoonGeocentricTrueDecNWithInstance(jde float64, n int, inst *Instance) float64 {
	lo := HMoonTrueLoNWithInstance(jde, n, inst)
	bo := HMoonTrueBoNWithInstance(jde, n, inst)
	return ArcSin(Sin(bo)*Cos(TrueObliquity(jde)) + Cos(bo)*Sin(TrueObliquity(jde))*Sin(lo))
}

// MoonGeocentricTrueRaDecWithInstance 月球地心真赤经纬 / true geocentric lunar right ascension and declination with an explicit ephemeris instance.
func MoonGeocentricTrueRaDecWithInstance(jde float64, inst *Instance) (float64, float64) {
	if inst == nil {
		return HMoonGeocentricTrueRaDec(jde)
	}
	return MoonGeocentricTrueRaWithInstance(jde, inst), MoonGeocentricTrueDecWithInstance(jde, inst)
}

// MoonGeocentricTrueRaDecNWithInstance 月球地心真赤经纬（n 变体） / n-parameterised true geocentric lunar right ascension and declination with an explicit ephemeris instance.
func MoonGeocentricTrueRaDecNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	return MoonGeocentricTrueRaNWithInstance(jde, n, inst), MoonGeocentricTrueDecNWithInstance(jde, n, inst)
}

// MoonTrueRaWithInstance 月球真赤经 / lunar true right ascension with an explicit ephemeris instance.
func MoonTrueRaWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return HMoonTrueRa(jde)
	}
	return moonApparentRaN(jde, -1, inst)
}

// MoonTrueRaNWithInstance 月球真赤经（n 变体） / n-parameterised lunar true right ascension with an explicit ephemeris instance.
func MoonTrueRaNWithInstance(jde float64, n int, inst *Instance) float64 {
	return moonApparentRaN(jde, n, inst)
}

// MoonTrueDecWithInstance 月球真赤纬 / lunar true declination with an explicit ephemeris instance.
func MoonTrueDecWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return HMoonTrueDec(jde)
	}
	return MoonTrueDecNWithInstance(jde, -1, inst)
}

// MoonTrueDecNWithInstance 月球真赤纬（n 变体） / n-parameterised lunar true declination with an explicit ephemeris instance.
func MoonTrueDecNWithInstance(jde float64, n int, inst *Instance) float64 {
	lo := HMoonApparentLoNWithInstance(jde, n, inst)
	bo := HMoonTrueBoNWithInstance(jde, n, inst)
	return ArcSin(Sin(bo)*Cos(TrueObliquity(jde)) + Cos(bo)*Sin(TrueObliquity(jde))*Sin(lo))
}

// MoonTrueRaDecWithInstance 月球真赤经纬 / lunar true right ascension and declination with an explicit ephemeris instance.
func MoonTrueRaDecWithInstance(jde float64, inst *Instance) (float64, float64) {
	if inst == nil {
		return HMoonTrueRaDec(jde)
	}
	return MoonTrueRaWithInstance(jde, inst), MoonTrueDecWithInstance(jde, inst)
}

// MoonTrueRaDecNWithInstance 月球真赤经纬（n 变体） / n-parameterised lunar true right ascension and declination with an explicit ephemeris instance.
func MoonTrueRaDecNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	return MoonTrueRaNWithInstance(jde, n, inst), MoonTrueDecNWithInstance(jde, n, inst)
}

// MoonApparentRaWithInstance 月球站心视赤经 / topocentric apparent lunar right ascension with an explicit ephemeris instance.
func MoonApparentRaWithInstance(jd, lon, lat, tz float64, inst *Instance) float64 {
	if inst == nil {
		return HMoonApparentRa(jd, lon, lat, tz)
	}
	ra, _ := HMoonApparentRaDecWithInstance(jd, lon, lat, tz, inst)
	return ra
}

// MoonApparentDecWithInstance 月球站心视赤纬 / topocentric apparent lunar declination with an explicit ephemeris instance.
func MoonApparentDecWithInstance(jd, lon, lat, tz float64, inst *Instance) float64 {
	if inst == nil {
		return HMoonApparentDec(jd, lon, lat, tz)
	}
	_, dec := HMoonApparentRaDecWithInstance(jd, lon, lat, tz, inst)
	return dec
}
