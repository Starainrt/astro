package basic

import (
	"math"

	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func moonCalcNewNWithInstance(coordIndex int, jd float64, n int, inst *Instance) float64 {
	if inst == nil {
		return planet.WhereMoonN(coordIndex, jd, n)
	}
	if inst.Model() != ephem.Custom {
		return planet.WhereMoonModelN(inst.Model(), coordIndex, jd, n)
	}
	if lo, bo, distAU, ok := instanceStateEcliptic(ephem.Moon, jd, inst); ok {
		switch coordIndex {
		case 0:
			return lo * math.Pi / 180
		case 1:
			return bo * math.Pi / 180
		case 2:
			return distAU * astronomicalUnitKM
		}
	}
	return math.NaN()
}

// MoonCalcNewWithInstance 指定星历实例的月球几何量 / lunar geometry with an explicit ephemeris instance.
func MoonCalcNewWithInstance(coordIndex int, jd float64, inst *Instance) float64 {
	return moonCalcNewNWithInstance(coordIndex, jd, -1, inst)
}

// MoonCalcNewNWithInstance 指定星历实例的月球几何量（n 变体） / n-parameterised lunar geometry with an explicit ephemeris instance.
func MoonCalcNewNWithInstance(coordIndex int, jd float64, n int, inst *Instance) float64 {
	return moonCalcNewNWithInstance(coordIndex, jd, n, inst)
}

// HMoonTrueLoWithInstance 指定星历实例的月球真黄经 / true lunar longitude with an explicit ephemeris instance.
func HMoonTrueLoWithInstance(jde float64, inst *Instance) float64 {
	return HMoonTrueLoNWithInstance(jde, -1, inst)
}

// HMoonTrueLoNWithInstance 指定星历实例的月球真黄经（n 变体） / n-parameterised true lunar longitude with an explicit ephemeris instance.
func HMoonTrueLoNWithInstance(jde float64, n int, inst *Instance) float64 {
	return Limit360(moonCalcNewNWithInstance(0, jde, n, inst) * 180 / math.Pi)
}

// HMoonTrueBoWithInstance 指定星历实例的月球真黄纬 / true lunar latitude with an explicit ephemeris instance.
func HMoonTrueBoWithInstance(jde float64, inst *Instance) float64 {
	return HMoonTrueBoNWithInstance(jde, -1, inst)
}

// HMoonTrueBoNWithInstance 指定星历实例的月球真黄纬（n 变体） / n-parameterised true lunar latitude with an explicit ephemeris instance.
func HMoonTrueBoNWithInstance(jde float64, n int, inst *Instance) float64 {
	return moonCalcNewNWithInstance(1, jde, n, inst) * 180 / math.Pi
}

// HMoonAwayWithInstance 指定星历实例的月地距离 / lunar distance with an explicit ephemeris instance.
func HMoonAwayWithInstance(jde float64, inst *Instance) float64 {
	return HMoonAwayNWithInstance(jde, -1, inst)
}

// HMoonAwayNWithInstance 指定星历实例的月地距离（n 变体） / n-parameterised lunar distance with an explicit ephemeris instance.
func HMoonAwayNWithInstance(jde float64, n int, inst *Instance) float64 {
	return moonCalcNewNWithInstance(2, jde, n, inst)
}

// HMoonApparentLoWithInstance 指定星历实例的月球视黄经 / apparent lunar longitude with an explicit ephemeris instance.
func HMoonApparentLoWithInstance(jde float64, inst *Instance) float64 {
	return HMoonApparentLoNWithInstance(jde, -1, inst)
}

// HMoonApparentLoNWithInstance 指定星历实例的月球视黄经（n 变体） / n-parameterised apparent lunar longitude with an explicit ephemeris instance.
func HMoonApparentLoNWithInstance(jde float64, n int, inst *Instance) float64 {
	if lo, _, _, _, ok, err := vectorApparentWithInstance(ephem.Moon, jde, n, inst); err != nil {
		return math.NaN()
	} else if ok {
		return lo
	}
	return HMoonTrueLoNWithInstance(jde, n, inst) + Nutation2000Bi(jde)
}

// HMoonGeocentricApparentRaDecWithInstance 指定星历实例的月球地心视赤经纬 / apparent geocentric lunar right ascension and declination with an explicit ephemeris instance.
func HMoonGeocentricApparentRaDecWithInstance(jde float64, inst *Instance) (float64, float64) {
	return HMoonGeocentricApparentRaDecNWithInstance(jde, -1, inst)
}

// HMoonGeocentricApparentRaDecNWithInstance 指定星历实例的月球地心视赤经纬（n 变体） / n-parameterised apparent geocentric lunar right ascension and declination with an explicit ephemeris instance.
func HMoonGeocentricApparentRaDecNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	return moonApparentRaDecN(jde, n, inst)
}

// HMoonTrueRaDecWithInstance 指定星历实例的月球真赤经纬 / true lunar right ascension and declination with an explicit ephemeris instance.
func HMoonTrueRaDecWithInstance(jde float64, inst *Instance) (float64, float64) {
	return moonApparentRaDecN(jde, -1, inst)
}

// HMoonApparentRaDecWithInstance 指定星历实例的月球站心视赤经纬 / topocentric apparent lunar right ascension and declination with an explicit ephemeris instance.
func HMoonApparentRaDecWithInstance(jd, lon, lat, tz float64, inst *Instance) (float64, float64) {
	jde := UTC2TT(jd) - tz/24
	ra, dec := HMoonTrueRaDecWithInstance(jde, inst)
	away := HMoonAwayWithInstance(jde, inst) / astronomicalUnitKM
	return TopocentricRaDec(ra, dec, lat, lon, jd-tz/24, away, 0)
}
