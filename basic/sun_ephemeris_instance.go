package basic

import (
	"math"

	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

// HSunTrueLoWithInstance 指定星历实例的太阳真黄经 / true solar longitude with an explicit ephemeris instance.
func HSunTrueLoWithInstance(jde float64, inst *Instance) float64 {
	return HSunTrueLoNWithInstance(jde, -1, inst)
}

// HSunTrueLoNWithInstance 指定星历实例的太阳真黄经（n 变体） / n-parameterised true solar longitude with an explicit ephemeris instance.
func HSunTrueLoNWithInstance(jde float64, n int, inst *Instance) float64 {
	if inst == nil {
		return HSunTrueLoN(jde, n)
	}
	if inst.Model() != ephem.Custom {
		return planet.WherePlanetModelN(inst.Model(), 0, 0, jde, n)
	}
	if lo, _, _, ok := instanceStateEcliptic(ephem.Sun, jde, inst); ok {
		return lo
	}
	return math.NaN()
}

// HSunTrueBoWithInstance 指定星历实例的太阳真黄纬 / true solar latitude with an explicit ephemeris instance.
func HSunTrueBoWithInstance(jde float64, inst *Instance) float64 {
	return HSunTrueBoNWithInstance(jde, -1, inst)
}

// HSunTrueBoNWithInstance 指定星历实例的太阳真黄纬（n 变体） / n-parameterised true solar latitude with an explicit ephemeris instance.
func HSunTrueBoNWithInstance(jde float64, n int, inst *Instance) float64 {
	if inst == nil {
		return HSunTrueBoN(jde, n)
	}
	if inst.Model() != ephem.Custom {
		return planet.WherePlanetModelN(inst.Model(), 0, 1, jde, n)
	}
	if _, bo, _, ok := instanceStateEcliptic(ephem.Sun, jde, inst); ok {
		return bo
	}
	return math.NaN()
}

// EarthAwayNWithInstance 指定星历实例的日地距离 / Sun-Earth distance with an explicit ephemeris instance.
func EarthAwayNWithInstance(jde float64, n int, inst *Instance) float64 {
	if inst == nil {
		return EarthAwayN(jde, n)
	}
	if inst.Model() != ephem.Custom {
		return planet.WherePlanetModelN(inst.Model(), 0, 2, jde, n)
	}
	if _, _, distanceAU, ok := instanceStateEcliptic(ephem.Sun, jde, inst); ok {
		return distanceAU
	}
	return math.NaN()
}

// HSunApparentLoWithInstance 指定星历实例的太阳视黄经 / apparent solar longitude with an explicit ephemeris instance.
func HSunApparentLoWithInstance(jde float64, inst *Instance) float64 {
	return HSunApparentLoNWithInstance(jde, -1, inst)
}

// HSunApparentLoNWithInstance 指定星历实例的太阳视黄经（n 变体） / n-parameterised apparent solar longitude with an explicit ephemeris instance.
func HSunApparentLoNWithInstance(jde float64, n int, inst *Instance) float64 {
	if lo, _, _, _, ok, err := vectorApparentWithInstance(ephem.Sun, jde, n, inst); err != nil {
		return math.NaN()
	} else if ok {
		return lo
	}
	lo := HSunTrueLoNWithInstance(jde, n, inst)
	return lo + Nutation2000Bi(jde) - 20.49552/EarthAwayNWithInstance(jde, n, inst)/3600
}

// HSunApparentRaDecWithInstance 指定星历实例的太阳视赤经纬 / apparent solar right ascension and declination with an explicit ephemeris instance.
func HSunApparentRaDecWithInstance(jde float64, inst *Instance) (float64, float64) {
	return HSunApparentRaDecNWithInstance(jde, -1, inst)
}

// HSunApparentRaDecNWithInstance 指定星历实例的太阳视赤经纬（n 变体） / n-parameterised apparent solar right ascension and declination with an explicit ephemeris instance.
func HSunApparentRaDecNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	// 内置模型与包级 HSunApparentRaDecN 同口径。
	if inst == nil {
		return HSunApparentRaDecN(jde, n)
	}
	if inst.Model() != ephem.Custom {
		ra, dec, _ := hSunApparentRaDecDistanceN(jde, n, inst)
		return ra, dec
	}
	if lo, bo, _, _, ok, err := vectorApparentWithInstance(ephem.Sun, jde, n, inst); err != nil {
		return math.NaN(), math.NaN()
	} else if ok {
		return LoBoToRaDec(jde, lo, bo)
	}
	return sunApparentRaDecN(jde, n, inst)
}

// HSunApparentRaWithInstance 指定星历实例的太阳视赤经 / apparent solar right ascension with an explicit ephemeris instance.
func HSunApparentRaWithInstance(jde float64, inst *Instance) float64 {
	return HSunApparentRaNWithInstance(jde, -1, inst)
}

// HSunApparentRaNWithInstance 指定星历实例的太阳视赤经（n 变体） / n-parameterised apparent solar right ascension with an explicit ephemeris instance.
func HSunApparentRaNWithInstance(jde float64, n int, inst *Instance) float64 {
	return sunApparentRaN(jde, n, inst)
}

// HSunApparentDecWithInstance 指定星历实例的太阳视赤纬 / apparent solar declination with an explicit ephemeris instance.
func HSunApparentDecWithInstance(jde float64, inst *Instance) float64 {
	return HSunApparentDecNWithInstance(jde, -1, inst)
}

// HSunApparentDecNWithInstance 指定星历实例的太阳视赤纬（n 变体） / n-parameterised apparent solar declination with an explicit ephemeris instance.
func HSunApparentDecNWithInstance(jde float64, n int, inst *Instance) float64 {
	_, dec := sunApparentRaDecN(jde, n, inst)
	return dec
}

// HSunTrueRaWithInstance 指定星历实例的太阳真赤经 / true solar right ascension with an explicit ephemeris instance.
func HSunTrueRaWithInstance(jde float64, inst *Instance) float64 {
	return HSunTrueRaNWithInstance(jde, -1, inst)
}

// HSunTrueRaNWithInstance 指定星历实例的太阳真赤经（n 变体） / n-parameterised true solar right ascension with an explicit ephemeris instance.
func HSunTrueRaNWithInstance(jde float64, n int, inst *Instance) float64 {
	lo := HSunTrueLoNWithInstance(jde, n, inst)
	eps := TrueObliquity(jde)
	return ArcTan2(Cos(eps)*Sin(lo), Cos(lo))
}

// HSunTrueDecWithInstance 指定星历实例的太阳真赤纬 / true solar declination with an explicit ephemeris instance.
func HSunTrueDecWithInstance(jde float64, inst *Instance) float64 {
	return HSunTrueDecNWithInstance(jde, -1, inst)
}

// HSunTrueDecNWithInstance 指定星历实例的太阳真赤纬（n 变体） / n-parameterised true solar declination with an explicit ephemeris instance.
func HSunTrueDecNWithInstance(jde float64, n int, inst *Instance) float64 {
	return ArcSin(Sin(EclipticObliquity(jde, false)) * Sin(HSunTrueLoNWithInstance(jde, n, inst)))
}
