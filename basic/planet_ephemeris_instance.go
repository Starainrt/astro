package basic

import (
	"math"

	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func planetBodyForInstance(index int) (ephem.Body, bool) {
	if index == -1 {
		return ephem.Earth, true
	}
	if index < 1 || index > 7 {
		return 0, false
	}
	return planetBody[index], true
}

func planetHeliocentricCoordinateWithInstance(index, coordinate int, jde float64, inst *Instance) (float64, bool) {
	if coordinate < 0 || coordinate > 2 {
		return math.NaN(), false
	}
	if inst == nil {
		return planet.WherePlanet(index, coordinate, jde), true
	}
	body, ok := planetBodyForInstance(index)
	if !ok {
		return math.NaN(), false
	}
	if lo, bo, distance, ok := instanceStateEclipticRelative(body, ephem.Sun, jde, jde, inst); ok {
		switch coordinate {
		case 0:
			return lo, true
		case 1:
			return bo, true
		default:
			return distance, true
		}
	}
	model := inst.Model()
	if model == ephem.Legacy || model == ephem.Semianalytic {
		return planet.WherePlanetModelN(model, index, coordinate, jde, -1), true
	}
	return math.NaN(), false
}

func planetGeocentricXYZWithInstance(index int, jde float64, inst *Instance) (x, y, z float64, ok bool) {
	if inst != nil {
		body, bodyOK := planetBodyForInstance(index)
		if bodyOK {
			if lo, bo, distance, stateOK := instanceStateEclipticRelative(body, ephem.Earth, jde, jde, inst); stateOK {
				x, y, z := sphericalToRectangular(lo, bo, distance)
				return x, y, z, true
			}
		}
	}
	pl, plOK := planetHeliocentricCoordinateWithInstance(index, 0, jde, inst)
	pb, pbOK := planetHeliocentricCoordinateWithInstance(index, 1, jde, inst)
	pr, prOK := planetHeliocentricCoordinateWithInstance(index, 2, jde, inst)
	el, elOK := planetHeliocentricCoordinateWithInstance(-1, 0, jde, inst)
	eb, ebOK := planetHeliocentricCoordinateWithInstance(-1, 1, jde, inst)
	er, erOK := planetHeliocentricCoordinateWithInstance(-1, 2, jde, inst)
	if !plOK || !pbOK || !prOK || !elOK || !ebOK || !erOK {
		return 0, 0, 0, false
	}
	px, py, pz := sphericalToRectangular(pl, pb, pr)
	ex, ey, ez := sphericalToRectangular(el, eb, er)
	return px - ex, py - ey, pz - ez, true
}

func planetApparentPositionModel(index int, jde float64, n int, model ephem.Model) (planetGeocentricPosition, float64, float64, bool) {
	ex, ey, ez := planetHeliocentricXYZModel(-1, jde, n, model)
	px, py, pz := planetHeliocentricXYZModel(index, jde, n, model)
	geoNow := geocentricPositionFromRectangular(px-ex, py-ey, pz-ez)
	distance := math.Sqrt(geoNow.x*geoNow.x + geoNow.y*geoNow.y + geoNow.z*geoNow.z)
	if !isFiniteFloat(distance) || distance <= 0 {
		return planetGeocentricPosition{}, 0, 0, false
	}
	tau := 0.0057755183 * distance
	px, py, pz = planetHeliocentricXYZModel(index, jde-tau, n, model)
	geo := geocentricPositionFromRectangular(px-ex, py-ey, pz-ez)
	baseLo, baseBo := geo.lo, geo.bo
	geo.lo = Limit360(baseLo + GXCLo(baseLo, baseBo, jde)/3600.0 + Nutation2000Bi(jde))
	geo.bo = baseBo + GXCBo(baseLo, baseBo, jde)/3600.0
	return geo, tau, distance, true
}

func planetHeliocentricXYZModel(index int, jde float64, n int, model ephem.Model) (x, y, z float64) {
	lo := planet.WherePlanetModelN(model, index, 0, jde, n)
	bo := planet.WherePlanetModelN(model, index, 1, jde, n)
	distance := planet.WherePlanetModelN(model, index, 2, jde, n)
	return sphericalToRectangular(lo, bo, distance)
}

func planetApparentPositionWithInstance(index int, jde float64, n int, inst *Instance) (planetGeocentricPosition, float64, float64, bool) {
	if inst == nil {
		geo, tau, distance := planetApparentGeocentricPositionAndDistanceN(index, jde, n)
		return geo, tau, distance, true
	}
	if lo, bo, tau, distance, ok, err := vectorApparentWithInstance(planetBody[index], jde, n, inst); err != nil {
		return planetGeocentricPosition{}, 0, 0, false
	} else if ok {
		x, y, z := sphericalToRectangular(lo, bo, distance)
		return planetGeocentricPosition{x: x, y: y, z: z, lo: lo, bo: bo}, tau, distance, true
	}
	if model := inst.Model(); model == ephem.Legacy || model == ephem.Semianalytic {
		return planetApparentPositionModel(index, jde, n, model)
	}
	body := planetBody[index]
	_, _, currentDistance, ok := instanceStateEclipticRelative(body, ephem.Earth, jde, jde, inst)
	if !ok || currentDistance <= 0 {
		return planetGeocentricPosition{}, 0, 0, false
	}
	tau := 0.0057755183 * currentDistance
	lo, bo, _, ok := instanceStateEclipticRelative(body, ephem.Earth, jde-tau, jde, inst)
	if !ok {
		return planetGeocentricPosition{}, 0, 0, false
	}
	x, y, z := sphericalToRectangular(lo, bo, currentDistance)
	return planetGeocentricPosition{x: x, y: y, z: z, lo: lo, bo: bo}, tau, currentDistance, true
}

func planetEarthAwayWithInstance(index int, jde float64, inst *Instance) float64 {
	return planetEarthAwayNWithInstance(index, jde, -1, inst)
}

func planetEarthAwayNWithInstance(index int, jde float64, n int, inst *Instance) float64 {
	if inst == nil {
		return planetEarthAwayExplicitN(index, jde, n)
	}
	if _, _, distance, ok := instanceStateEclipticRelative(planetBody[index], ephem.Earth, jde, jde, inst); ok {
		return distance
	}
	// 与 planetEarthAwayExplicitN 对称：新模型全精度时取视位置管线的距离。
	if index >= 1 && index <= 7 {
		if _, _, _, distance, ok, err := vectorApparentWithInstance(planetBody[index], jde, n, inst); err == nil && ok {
			return distance
		}
	}
	model := inst.Model()
	if model != ephem.Legacy && model != ephem.Semianalytic {
		return math.NaN()
	}
	x, y, z := planetHeliocentricXYZModel(index, jde, n, model)
	ex, ey, ez := planetHeliocentricXYZModel(-1, jde, n, model)
	return math.Sqrt((x-ex)*(x-ex) + (y-ey)*(y-ey) + (z-ez)*(z-ez))
}

func planetLWithInstance(index int, jde float64, inst *Instance) float64 {
	value, _ := planetHeliocentricCoordinateWithInstance(index, 0, jde, inst)
	return value
}

func planetBWithInstance(index int, jde float64, inst *Instance) float64 {
	value, _ := planetHeliocentricCoordinateWithInstance(index, 1, jde, inst)
	return value
}

func planetRWithInstance(index int, jde float64, inst *Instance) float64 {
	value, _ := planetHeliocentricCoordinateWithInstance(index, 2, jde, inst)
	return value
}

func planetApparentLoBoWithInstance(index int, jde float64, inst *Instance) (float64, float64) {
	return planetApparentLoBoNWithInstance(index, jde, -1, inst)
}

func planetApparentLoBoNWithInstance(index int, jde float64, n int, inst *Instance) (float64, float64) {
	geo, _, _, ok := planetApparentPositionWithInstance(index, jde, n, inst)
	if !ok {
		return math.NaN(), math.NaN()
	}
	return geo.lo, geo.bo
}

func planetApparentRaDecWithInstance(index int, jde float64, inst *Instance) (float64, float64) {
	lo, bo := planetApparentLoBoWithInstance(index, jde, inst)
	if math.IsNaN(lo) || math.IsNaN(bo) {
		return math.NaN(), math.NaN()
	}
	return planetApparentRaDecFromLoBo(jde, lo, bo)
}

func planetMagnitudeWithInstance(index int, jde float64, inst *Instance, base, phaseLinear, phaseQuadratic, phaseCubic float64) float64 {
	sunDistance := planetRWithInstance(index, jde, inst)
	earthDistance := planetEarthAwayWithInstance(index, jde, inst)
	earthSunDistance := planetRWithInstance(-1, jde, inst)
	phase := (sunDistance*sunDistance + earthDistance*earthDistance - earthSunDistance*earthSunDistance) / (2 * sunDistance * earthDistance)
	phase = ArcCos(phase)
	mag := base + 5*math.Log10(sunDistance*earthDistance) + phaseLinear*phase + phaseQuadratic*phase*phase + phaseCubic*phase*phase*phase
	return FloatRound(mag, 2)
}

// MercuryLWithInstance 水星日心黄经 / heliocentric longitude of Mercury with an explicit ephemeris instance.
func MercuryLWithInstance(jde float64, inst *Instance) float64 {
	return planetLWithInstance(1, jde, inst)
}

// MercuryBWithInstance 水星日心黄纬 / heliocentric latitude of Mercury with an explicit ephemeris instance.
func MercuryBWithInstance(jde float64, inst *Instance) float64 {
	return planetBWithInstance(1, jde, inst)
}

// MercuryRWithInstance 水星日心距离 / heliocentric distance of Mercury with an explicit ephemeris instance.
func MercuryRWithInstance(jde float64, inst *Instance) float64 {
	return planetRWithInstance(1, jde, inst)
}

// AMercuryXYZWithInstance 水星地心直角坐标 / geocentric rectangular coordinates of Mercury with an explicit ephemeris instance.
func AMercuryXYZWithInstance(jde float64, inst *Instance) (float64, float64, float64) {
	x, y, z, _ := planetGeocentricXYZWithInstance(1, jde, inst)
	return x, y, z
}

// AMercuryXWithInstance 水星地心 X 坐标 / geocentric X coordinate of Mercury with an explicit ephemeris instance.
func AMercuryXWithInstance(jde float64, inst *Instance) float64 {
	x, _, _ := AMercuryXYZWithInstance(jde, inst)
	return x
}

// AMercuryYWithInstance 水星地心 Y 坐标 / geocentric Y coordinate of Mercury with an explicit ephemeris instance.
func AMercuryYWithInstance(jde float64, inst *Instance) float64 {
	_, y, _ := AMercuryXYZWithInstance(jde, inst)
	return y
}

// AMercuryZWithInstance 水星地心 Z 坐标 / geocentric Z coordinate of Mercury with an explicit ephemeris instance.
func AMercuryZWithInstance(jde float64, inst *Instance) float64 {
	_, _, z := AMercuryXYZWithInstance(jde, inst)
	return z
}

// MercuryApparentLoBoWithInstance 水星视黄经纬 / apparent ecliptic longitude and latitude of Mercury with an explicit ephemeris instance.
func MercuryApparentLoBoWithInstance(jde float64, inst *Instance) (float64, float64) {
	return planetApparentLoBoWithInstance(1, jde, inst)
}

// MercuryApparentLoWithInstance 水星视黄经 / apparent ecliptic longitude of Mercury with an explicit ephemeris instance.
func MercuryApparentLoWithInstance(jde float64, inst *Instance) float64 {
	lo, _ := MercuryApparentLoBoWithInstance(jde, inst)
	return lo
}

// MercuryApparentBoWithInstance 水星视黄纬 / apparent ecliptic latitude of Mercury with an explicit ephemeris instance.
func MercuryApparentBoWithInstance(jde float64, inst *Instance) float64 {
	_, bo := MercuryApparentLoBoWithInstance(jde, inst)
	return bo
}

// MercuryApparentRaDecWithInstance 水星视赤经纬 / apparent right ascension and declination of Mercury with an explicit ephemeris instance.
func MercuryApparentRaDecWithInstance(jde float64, inst *Instance) (float64, float64) {
	return planetApparentRaDecWithInstance(1, jde, inst)
}

// MercuryApparentRaWithInstance 水星视赤经 / apparent right ascension of Mercury with an explicit ephemeris instance.
func MercuryApparentRaWithInstance(jde float64, inst *Instance) float64 {
	ra, _ := MercuryApparentRaDecWithInstance(jde, inst)
	return ra
}

// MercuryApparentDecWithInstance 水星视赤纬 / apparent declination of Mercury with an explicit ephemeris instance.
func MercuryApparentDecWithInstance(jde float64, inst *Instance) float64 {
	_, dec := MercuryApparentRaDecWithInstance(jde, inst)
	return dec
}

// MercuryApparentLoBoNWithInstance 水星视黄经纬（n 变体） / n-parameterised apparent ecliptic longitude and latitude of Mercury with an explicit ephemeris instance.
func MercuryApparentLoBoNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	return planetApparentLoBoNWithInstance(1, jde, n, inst)
}

// MercuryApparentLoNWithInstance 水星视黄经（n 变体） / n-parameterised apparent ecliptic longitude of Mercury with an explicit ephemeris instance.
func MercuryApparentLoNWithInstance(jde float64, n int, inst *Instance) float64 {
	lo, _ := MercuryApparentLoBoNWithInstance(jde, n, inst)
	return lo
}

// MercuryApparentBoNWithInstance 水星视黄纬（n 变体） / n-parameterised apparent ecliptic latitude of Mercury with an explicit ephemeris instance.
func MercuryApparentBoNWithInstance(jde float64, n int, inst *Instance) float64 {
	_, bo := MercuryApparentLoBoNWithInstance(jde, n, inst)
	return bo
}

// MercuryApparentRaDecNWithInstance 水星视赤经纬（n 变体） / n-parameterised apparent right ascension and declination of Mercury with an explicit ephemeris instance.
func MercuryApparentRaDecNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	lo, bo := MercuryApparentLoBoNWithInstance(jde, n, inst)
	return planetApparentRaDecFromLoBo(jde, lo, bo)
}

// MercuryApparentRaNWithInstance 水星视赤经（n 变体） / n-parameterised apparent right ascension of Mercury with an explicit ephemeris instance.
func MercuryApparentRaNWithInstance(jde float64, n int, inst *Instance) float64 {
	ra, _ := MercuryApparentRaDecNWithInstance(jde, n, inst)
	return ra
}

// MercuryApparentDecNWithInstance 水星视赤纬（n 变体） / n-parameterised apparent declination of Mercury with an explicit ephemeris instance.
func MercuryApparentDecNWithInstance(jde float64, n int, inst *Instance) float64 {
	_, dec := MercuryApparentRaDecNWithInstance(jde, n, inst)
	return dec
}

// EarthMercuryAwayWithInstance 地水距离 / Earth-Mercury distance with an explicit ephemeris instance.
func EarthMercuryAwayWithInstance(jde float64, inst *Instance) float64 {
	return planetEarthAwayWithInstance(1, jde, inst)
}

// EarthMercuryAwayNWithInstance 水星地心距离（n 变体） / n-parameterised Earth-Mercury distance with an explicit ephemeris instance.
func EarthMercuryAwayNWithInstance(jde float64, n int, inst *Instance) float64 {
	return planetEarthAwayNWithInstance(1, jde, n, inst)
}

// VenusLWithInstance 金星日心黄经 / heliocentric longitude of Venus with an explicit ephemeris instance.
func VenusLWithInstance(jde float64, inst *Instance) float64 {
	return planetLWithInstance(2, jde, inst)
}

// VenusBWithInstance 金星日心黄纬 / heliocentric latitude of Venus with an explicit ephemeris instance.
func VenusBWithInstance(jde float64, inst *Instance) float64 {
	return planetBWithInstance(2, jde, inst)
}

// VenusRWithInstance 金星日心距离 / heliocentric distance of Venus with an explicit ephemeris instance.
func VenusRWithInstance(jde float64, inst *Instance) float64 {
	return planetRWithInstance(2, jde, inst)
}

// AVenusXYZWithInstance 金星地心直角坐标 / geocentric rectangular coordinates of Venus with an explicit ephemeris instance.
func AVenusXYZWithInstance(jde float64, inst *Instance) (float64, float64, float64) {
	x, y, z, _ := planetGeocentricXYZWithInstance(2, jde, inst)
	return x, y, z
}

// AVenusXWithInstance 金星地心 X 坐标 / geocentric X coordinate of Venus with an explicit ephemeris instance.
func AVenusXWithInstance(jde float64, inst *Instance) float64 {
	x, _, _ := AVenusXYZWithInstance(jde, inst)
	return x
}

// AVenusYWithInstance 金星地心 Y 坐标 / geocentric Y coordinate of Venus with an explicit ephemeris instance.
func AVenusYWithInstance(jde float64, inst *Instance) float64 {
	_, y, _ := AVenusXYZWithInstance(jde, inst)
	return y
}

// AVenusZWithInstance 金星地心 Z 坐标 / geocentric Z coordinate of Venus with an explicit ephemeris instance.
func AVenusZWithInstance(jde float64, inst *Instance) float64 {
	_, _, z := AVenusXYZWithInstance(jde, inst)
	return z
}

// VenusApparentLoBoWithInstance 金星视黄经纬 / apparent ecliptic longitude and latitude of Venus with an explicit ephemeris instance.
func VenusApparentLoBoWithInstance(jde float64, inst *Instance) (float64, float64) {
	return planetApparentLoBoWithInstance(2, jde, inst)
}

// VenusApparentLoWithInstance 金星视黄经 / apparent ecliptic longitude of Venus with an explicit ephemeris instance.
func VenusApparentLoWithInstance(jde float64, inst *Instance) float64 {
	lo, _ := VenusApparentLoBoWithInstance(jde, inst)
	return lo
}

// VenusApparentBoWithInstance 金星视黄纬 / apparent ecliptic latitude of Venus with an explicit ephemeris instance.
func VenusApparentBoWithInstance(jde float64, inst *Instance) float64 {
	_, bo := VenusApparentLoBoWithInstance(jde, inst)
	return bo
}

// VenusApparentRaDecWithInstance 金星视赤经纬 / apparent right ascension and declination of Venus with an explicit ephemeris instance.
func VenusApparentRaDecWithInstance(jde float64, inst *Instance) (float64, float64) {
	return planetApparentRaDecWithInstance(2, jde, inst)
}

// VenusApparentRaWithInstance 金星视赤经 / apparent right ascension of Venus with an explicit ephemeris instance.
func VenusApparentRaWithInstance(jde float64, inst *Instance) float64 {
	ra, _ := VenusApparentRaDecWithInstance(jde, inst)
	return ra
}

// VenusApparentDecWithInstance 金星视赤纬 / apparent declination of Venus with an explicit ephemeris instance.
func VenusApparentDecWithInstance(jde float64, inst *Instance) float64 {
	_, dec := VenusApparentRaDecWithInstance(jde, inst)
	return dec
}

// VenusApparentLoBoNWithInstance 金星视黄经纬（n 变体） / n-parameterised apparent ecliptic longitude and latitude of Venus with an explicit ephemeris instance.
func VenusApparentLoBoNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	return planetApparentLoBoNWithInstance(2, jde, n, inst)
}

// VenusApparentLoNWithInstance 金星视黄经（n 变体） / n-parameterised apparent ecliptic longitude of Venus with an explicit ephemeris instance.
func VenusApparentLoNWithInstance(jde float64, n int, inst *Instance) float64 {
	lo, _ := VenusApparentLoBoNWithInstance(jde, n, inst)
	return lo
}

// VenusApparentBoNWithInstance 金星视黄纬（n 变体） / n-parameterised apparent ecliptic latitude of Venus with an explicit ephemeris instance.
func VenusApparentBoNWithInstance(jde float64, n int, inst *Instance) float64 {
	_, bo := VenusApparentLoBoNWithInstance(jde, n, inst)
	return bo
}

// VenusApparentRaDecNWithInstance 金星视赤经纬（n 变体） / n-parameterised apparent right ascension and declination of Venus with an explicit ephemeris instance.
func VenusApparentRaDecNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	lo, bo := VenusApparentLoBoNWithInstance(jde, n, inst)
	return planetApparentRaDecFromLoBo(jde, lo, bo)
}

// VenusApparentRaNWithInstance 金星视赤经（n 变体） / n-parameterised apparent right ascension of Venus with an explicit ephemeris instance.
func VenusApparentRaNWithInstance(jde float64, n int, inst *Instance) float64 {
	ra, _ := VenusApparentRaDecNWithInstance(jde, n, inst)
	return ra
}

// VenusApparentDecNWithInstance 金星视赤纬（n 变体） / n-parameterised apparent declination of Venus with an explicit ephemeris instance.
func VenusApparentDecNWithInstance(jde float64, n int, inst *Instance) float64 {
	_, dec := VenusApparentRaDecNWithInstance(jde, n, inst)
	return dec
}

// EarthVenusAwayWithInstance 地金距离 / Earth-Venus distance with an explicit ephemeris instance.
func EarthVenusAwayWithInstance(jde float64, inst *Instance) float64 {
	return planetEarthAwayWithInstance(2, jde, inst)
}

// EarthVenusAwayNWithInstance 金星地心距离（n 变体） / n-parameterised Earth-Venus distance with an explicit ephemeris instance.
func EarthVenusAwayNWithInstance(jde float64, n int, inst *Instance) float64 {
	return planetEarthAwayNWithInstance(2, jde, n, inst)
}

// MarsLWithInstance 火星日心黄经 / heliocentric longitude of Mars with an explicit ephemeris instance.
func MarsLWithInstance(jde float64, inst *Instance) float64 {
	return planetLWithInstance(3, jde, inst)
}

// MarsBWithInstance 火星日心黄纬 / heliocentric latitude of Mars with an explicit ephemeris instance.
func MarsBWithInstance(jde float64, inst *Instance) float64 {
	return planetBWithInstance(3, jde, inst)
}

// MarsRWithInstance 火星日心距离 / heliocentric distance of Mars with an explicit ephemeris instance.
func MarsRWithInstance(jde float64, inst *Instance) float64 {
	return planetRWithInstance(3, jde, inst)
}

// AMarsXYZWithInstance 火星地心直角坐标 / geocentric rectangular coordinates of Mars with an explicit ephemeris instance.
func AMarsXYZWithInstance(jde float64, inst *Instance) (float64, float64, float64) {
	x, y, z, _ := planetGeocentricXYZWithInstance(3, jde, inst)
	return x, y, z
}

// AMarsXWithInstance 火星地心 X 坐标 / geocentric X coordinate of Mars with an explicit ephemeris instance.
func AMarsXWithInstance(jde float64, inst *Instance) float64 {
	x, _, _ := AMarsXYZWithInstance(jde, inst)
	return x
}

// AMarsYWithInstance 火星地心 Y 坐标 / geocentric Y coordinate of Mars with an explicit ephemeris instance.
func AMarsYWithInstance(jde float64, inst *Instance) float64 {
	_, y, _ := AMarsXYZWithInstance(jde, inst)
	return y
}

// AMarsZWithInstance 火星地心 Z 坐标 / geocentric Z coordinate of Mars with an explicit ephemeris instance.
func AMarsZWithInstance(jde float64, inst *Instance) float64 {
	_, _, z := AMarsXYZWithInstance(jde, inst)
	return z
}

// MarsApparentLoBoWithInstance 火星视黄经纬 / apparent ecliptic longitude and latitude of Mars with an explicit ephemeris instance.
func MarsApparentLoBoWithInstance(jde float64, inst *Instance) (float64, float64) {
	return planetApparentLoBoWithInstance(3, jde, inst)
}

// MarsApparentLoWithInstance 火星视黄经 / apparent ecliptic longitude of Mars with an explicit ephemeris instance.
func MarsApparentLoWithInstance(jde float64, inst *Instance) float64 {
	lo, _ := MarsApparentLoBoWithInstance(jde, inst)
	return lo
}

// MarsApparentBoWithInstance 火星视黄纬 / apparent ecliptic latitude of Mars with an explicit ephemeris instance.
func MarsApparentBoWithInstance(jde float64, inst *Instance) float64 {
	_, bo := MarsApparentLoBoWithInstance(jde, inst)
	return bo
}

// MarsApparentRaDecWithInstance 火星视赤经纬 / apparent right ascension and declination of Mars with an explicit ephemeris instance.
func MarsApparentRaDecWithInstance(jde float64, inst *Instance) (float64, float64) {
	return planetApparentRaDecWithInstance(3, jde, inst)
}

// MarsApparentRaWithInstance 火星视赤经 / apparent right ascension of Mars with an explicit ephemeris instance.
func MarsApparentRaWithInstance(jde float64, inst *Instance) float64 {
	ra, _ := MarsApparentRaDecWithInstance(jde, inst)
	return ra
}

// MarsApparentDecWithInstance 火星视赤纬 / apparent declination of Mars with an explicit ephemeris instance.
func MarsApparentDecWithInstance(jde float64, inst *Instance) float64 {
	_, dec := MarsApparentRaDecWithInstance(jde, inst)
	return dec
}

// MarsApparentLoBoNWithInstance 火星视黄经纬（n 变体） / n-parameterised apparent ecliptic longitude and latitude of Mars with an explicit ephemeris instance.
func MarsApparentLoBoNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	return planetApparentLoBoNWithInstance(3, jde, n, inst)
}

// MarsApparentLoNWithInstance 火星视黄经（n 变体） / n-parameterised apparent ecliptic longitude of Mars with an explicit ephemeris instance.
func MarsApparentLoNWithInstance(jde float64, n int, inst *Instance) float64 {
	lo, _ := MarsApparentLoBoNWithInstance(jde, n, inst)
	return lo
}

// MarsApparentBoNWithInstance 火星视黄纬（n 变体） / n-parameterised apparent ecliptic latitude of Mars with an explicit ephemeris instance.
func MarsApparentBoNWithInstance(jde float64, n int, inst *Instance) float64 {
	_, bo := MarsApparentLoBoNWithInstance(jde, n, inst)
	return bo
}

// MarsApparentRaDecNWithInstance 火星视赤经纬（n 变体） / n-parameterised apparent right ascension and declination of Mars with an explicit ephemeris instance.
func MarsApparentRaDecNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	lo, bo := MarsApparentLoBoNWithInstance(jde, n, inst)
	return planetApparentRaDecFromLoBo(jde, lo, bo)
}

// MarsApparentRaNWithInstance 火星视赤经（n 变体） / n-parameterised apparent right ascension of Mars with an explicit ephemeris instance.
func MarsApparentRaNWithInstance(jde float64, n int, inst *Instance) float64 {
	ra, _ := MarsApparentRaDecNWithInstance(jde, n, inst)
	return ra
}

// MarsApparentDecNWithInstance 火星视赤纬（n 变体） / n-parameterised apparent declination of Mars with an explicit ephemeris instance.
func MarsApparentDecNWithInstance(jde float64, n int, inst *Instance) float64 {
	_, dec := MarsApparentRaDecNWithInstance(jde, n, inst)
	return dec
}

// EarthMarsAwayWithInstance 地火距离 / Earth-Mars distance with an explicit ephemeris instance.
func EarthMarsAwayWithInstance(jde float64, inst *Instance) float64 {
	return planetEarthAwayWithInstance(3, jde, inst)
}

// EarthMarsAwayNWithInstance 火星地心距离（n 变体） / n-parameterised Earth-Mars distance with an explicit ephemeris instance.
func EarthMarsAwayNWithInstance(jde float64, n int, inst *Instance) float64 {
	return planetEarthAwayNWithInstance(3, jde, n, inst)
}

// JupiterLWithInstance 木星日心黄经 / heliocentric longitude of Jupiter with an explicit ephemeris instance.
func JupiterLWithInstance(jde float64, inst *Instance) float64 {
	return planetLWithInstance(4, jde, inst)
}

// JupiterBWithInstance 木星日心黄纬 / heliocentric latitude of Jupiter with an explicit ephemeris instance.
func JupiterBWithInstance(jde float64, inst *Instance) float64 {
	return planetBWithInstance(4, jde, inst)
}

// JupiterRWithInstance 木星日心距离 / heliocentric distance of Jupiter with an explicit ephemeris instance.
func JupiterRWithInstance(jde float64, inst *Instance) float64 {
	return planetRWithInstance(4, jde, inst)
}

// AJupiterXYZWithInstance 木星地心直角坐标 / geocentric rectangular coordinates of Jupiter with an explicit ephemeris instance.
func AJupiterXYZWithInstance(jde float64, inst *Instance) (float64, float64, float64) {
	x, y, z, _ := planetGeocentricXYZWithInstance(4, jde, inst)
	return x, y, z
}

// AJupiterXWithInstance 木星地心 X 坐标 / geocentric X coordinate of Jupiter with an explicit ephemeris instance.
func AJupiterXWithInstance(jde float64, inst *Instance) float64 {
	x, _, _ := AJupiterXYZWithInstance(jde, inst)
	return x
}

// AJupiterYWithInstance 木星地心 Y 坐标 / geocentric Y coordinate of Jupiter with an explicit ephemeris instance.
func AJupiterYWithInstance(jde float64, inst *Instance) float64 {
	_, y, _ := AJupiterXYZWithInstance(jde, inst)
	return y
}

// AJupiterZWithInstance 木星地心 Z 坐标 / geocentric Z coordinate of Jupiter with an explicit ephemeris instance.
func AJupiterZWithInstance(jde float64, inst *Instance) float64 {
	_, _, z := AJupiterXYZWithInstance(jde, inst)
	return z
}

// JupiterApparentLoBoWithInstance 木星视黄经纬 / apparent ecliptic longitude and latitude of Jupiter with an explicit ephemeris instance.
func JupiterApparentLoBoWithInstance(jde float64, inst *Instance) (float64, float64) {
	return planetApparentLoBoWithInstance(4, jde, inst)
}

// JupiterApparentLoWithInstance 木星视黄经 / apparent ecliptic longitude of Jupiter with an explicit ephemeris instance.
func JupiterApparentLoWithInstance(jde float64, inst *Instance) float64 {
	lo, _ := JupiterApparentLoBoWithInstance(jde, inst)
	return lo
}

// JupiterApparentBoWithInstance 木星视黄纬 / apparent ecliptic latitude of Jupiter with an explicit ephemeris instance.
func JupiterApparentBoWithInstance(jde float64, inst *Instance) float64 {
	_, bo := JupiterApparentLoBoWithInstance(jde, inst)
	return bo
}

// JupiterApparentRaDecWithInstance 木星视赤经纬 / apparent right ascension and declination of Jupiter with an explicit ephemeris instance.
func JupiterApparentRaDecWithInstance(jde float64, inst *Instance) (float64, float64) {
	return planetApparentRaDecWithInstance(4, jde, inst)
}

// JupiterApparentRaWithInstance 木星视赤经 / apparent right ascension of Jupiter with an explicit ephemeris instance.
func JupiterApparentRaWithInstance(jde float64, inst *Instance) float64 {
	ra, _ := JupiterApparentRaDecWithInstance(jde, inst)
	return ra
}

// JupiterApparentDecWithInstance 木星视赤纬 / apparent declination of Jupiter with an explicit ephemeris instance.
func JupiterApparentDecWithInstance(jde float64, inst *Instance) float64 {
	_, dec := JupiterApparentRaDecWithInstance(jde, inst)
	return dec
}

// JupiterApparentLoBoNWithInstance 木星视黄经纬（n 变体） / n-parameterised apparent ecliptic longitude and latitude of Jupiter with an explicit ephemeris instance.
func JupiterApparentLoBoNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	return planetApparentLoBoNWithInstance(4, jde, n, inst)
}

// JupiterApparentLoNWithInstance 木星视黄经（n 变体） / n-parameterised apparent ecliptic longitude of Jupiter with an explicit ephemeris instance.
func JupiterApparentLoNWithInstance(jde float64, n int, inst *Instance) float64 {
	lo, _ := JupiterApparentLoBoNWithInstance(jde, n, inst)
	return lo
}

// JupiterApparentBoNWithInstance 木星视黄纬（n 变体） / n-parameterised apparent ecliptic latitude of Jupiter with an explicit ephemeris instance.
func JupiterApparentBoNWithInstance(jde float64, n int, inst *Instance) float64 {
	_, bo := JupiterApparentLoBoNWithInstance(jde, n, inst)
	return bo
}

// JupiterApparentRaDecNWithInstance 木星视赤经纬（n 变体） / n-parameterised apparent right ascension and declination of Jupiter with an explicit ephemeris instance.
func JupiterApparentRaDecNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	lo, bo := JupiterApparentLoBoNWithInstance(jde, n, inst)
	return planetApparentRaDecFromLoBo(jde, lo, bo)
}

// JupiterApparentRaNWithInstance 木星视赤经（n 变体） / n-parameterised apparent right ascension of Jupiter with an explicit ephemeris instance.
func JupiterApparentRaNWithInstance(jde float64, n int, inst *Instance) float64 {
	ra, _ := JupiterApparentRaDecNWithInstance(jde, n, inst)
	return ra
}

// JupiterApparentDecNWithInstance 木星视赤纬（n 变体） / n-parameterised apparent declination of Jupiter with an explicit ephemeris instance.
func JupiterApparentDecNWithInstance(jde float64, n int, inst *Instance) float64 {
	_, dec := JupiterApparentRaDecNWithInstance(jde, n, inst)
	return dec
}

// EarthJupiterAwayWithInstance 地木距离 / Earth-Jupiter distance with an explicit ephemeris instance.
func EarthJupiterAwayWithInstance(jde float64, inst *Instance) float64 {
	return planetEarthAwayWithInstance(4, jde, inst)
}

// EarthJupiterAwayNWithInstance 木星地心距离（n 变体） / n-parameterised Earth-Jupiter distance with an explicit ephemeris instance.
func EarthJupiterAwayNWithInstance(jde float64, n int, inst *Instance) float64 {
	return planetEarthAwayNWithInstance(4, jde, n, inst)
}

// SaturnLWithInstance 土星日心黄经 / heliocentric longitude of Saturn with an explicit ephemeris instance.
func SaturnLWithInstance(jde float64, inst *Instance) float64 {
	return planetLWithInstance(5, jde, inst)
}

// SaturnBWithInstance 土星日心黄纬 / heliocentric latitude of Saturn with an explicit ephemeris instance.
func SaturnBWithInstance(jde float64, inst *Instance) float64 {
	return planetBWithInstance(5, jde, inst)
}

// SaturnRWithInstance 土星日心距离 / heliocentric distance of Saturn with an explicit ephemeris instance.
func SaturnRWithInstance(jde float64, inst *Instance) float64 {
	return planetRWithInstance(5, jde, inst)
}

// ASaturnXYZWithInstance 土星地心直角坐标 / geocentric rectangular coordinates of Saturn with an explicit ephemeris instance.
func ASaturnXYZWithInstance(jde float64, inst *Instance) (float64, float64, float64) {
	x, y, z, _ := planetGeocentricXYZWithInstance(5, jde, inst)
	return x, y, z
}

// ASaturnXWithInstance 土星地心 X 坐标 / geocentric X coordinate of Saturn with an explicit ephemeris instance.
func ASaturnXWithInstance(jde float64, inst *Instance) float64 {
	x, _, _ := ASaturnXYZWithInstance(jde, inst)
	return x
}

// ASaturnYWithInstance 土星地心 Y 坐标 / geocentric Y coordinate of Saturn with an explicit ephemeris instance.
func ASaturnYWithInstance(jde float64, inst *Instance) float64 {
	_, y, _ := ASaturnXYZWithInstance(jde, inst)
	return y
}

// ASaturnZWithInstance 土星地心 Z 坐标 / geocentric Z coordinate of Saturn with an explicit ephemeris instance.
func ASaturnZWithInstance(jde float64, inst *Instance) float64 {
	_, _, z := ASaturnXYZWithInstance(jde, inst)
	return z
}

// SaturnApparentLoBoWithInstance 土星视黄经纬 / apparent ecliptic longitude and latitude of Saturn with an explicit ephemeris instance.
func SaturnApparentLoBoWithInstance(jde float64, inst *Instance) (float64, float64) {
	return planetApparentLoBoWithInstance(5, jde, inst)
}

// SaturnApparentLoWithInstance 土星视黄经 / apparent ecliptic longitude of Saturn with an explicit ephemeris instance.
func SaturnApparentLoWithInstance(jde float64, inst *Instance) float64 {
	lo, _ := SaturnApparentLoBoWithInstance(jde, inst)
	return lo
}

// SaturnApparentBoWithInstance 土星视黄纬 / apparent ecliptic latitude of Saturn with an explicit ephemeris instance.
func SaturnApparentBoWithInstance(jde float64, inst *Instance) float64 {
	_, bo := SaturnApparentLoBoWithInstance(jde, inst)
	return bo
}

// SaturnApparentRaDecWithInstance 土星视赤经纬 / apparent right ascension and declination of Saturn with an explicit ephemeris instance.
func SaturnApparentRaDecWithInstance(jde float64, inst *Instance) (float64, float64) {
	return planetApparentRaDecWithInstance(5, jde, inst)
}

// SaturnApparentRaWithInstance 土星视赤经 / apparent right ascension of Saturn with an explicit ephemeris instance.
func SaturnApparentRaWithInstance(jde float64, inst *Instance) float64 {
	ra, _ := SaturnApparentRaDecWithInstance(jde, inst)
	return ra
}

// SaturnApparentDecWithInstance 土星视赤纬 / apparent declination of Saturn with an explicit ephemeris instance.
func SaturnApparentDecWithInstance(jde float64, inst *Instance) float64 {
	_, dec := SaturnApparentRaDecWithInstance(jde, inst)
	return dec
}

// SaturnApparentLoBoNWithInstance 土星视黄经纬（n 变体） / n-parameterised apparent ecliptic longitude and latitude of Saturn with an explicit ephemeris instance.
func SaturnApparentLoBoNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	return planetApparentLoBoNWithInstance(5, jde, n, inst)
}

// SaturnApparentLoNWithInstance 土星视黄经（n 变体） / n-parameterised apparent ecliptic longitude of Saturn with an explicit ephemeris instance.
func SaturnApparentLoNWithInstance(jde float64, n int, inst *Instance) float64 {
	lo, _ := SaturnApparentLoBoNWithInstance(jde, n, inst)
	return lo
}

// SaturnApparentBoNWithInstance 土星视黄纬（n 变体） / n-parameterised apparent ecliptic latitude of Saturn with an explicit ephemeris instance.
func SaturnApparentBoNWithInstance(jde float64, n int, inst *Instance) float64 {
	_, bo := SaturnApparentLoBoNWithInstance(jde, n, inst)
	return bo
}

// SaturnApparentRaDecNWithInstance 土星视赤经纬（n 变体） / n-parameterised apparent right ascension and declination of Saturn with an explicit ephemeris instance.
func SaturnApparentRaDecNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	lo, bo := SaturnApparentLoBoNWithInstance(jde, n, inst)
	return planetApparentRaDecFromLoBo(jde, lo, bo)
}

// SaturnApparentRaNWithInstance 土星视赤经（n 变体） / n-parameterised apparent right ascension of Saturn with an explicit ephemeris instance.
func SaturnApparentRaNWithInstance(jde float64, n int, inst *Instance) float64 {
	ra, _ := SaturnApparentRaDecNWithInstance(jde, n, inst)
	return ra
}

// SaturnApparentDecNWithInstance 土星视赤纬（n 变体） / n-parameterised apparent declination of Saturn with an explicit ephemeris instance.
func SaturnApparentDecNWithInstance(jde float64, n int, inst *Instance) float64 {
	_, dec := SaturnApparentRaDecNWithInstance(jde, n, inst)
	return dec
}

// EarthSaturnAwayWithInstance 地土距离 / Earth-Saturn distance with an explicit ephemeris instance.
func EarthSaturnAwayWithInstance(jde float64, inst *Instance) float64 {
	return planetEarthAwayWithInstance(5, jde, inst)
}

// EarthSaturnAwayNWithInstance 土星地心距离（n 变体） / n-parameterised Earth-Saturn distance with an explicit ephemeris instance.
func EarthSaturnAwayNWithInstance(jde float64, n int, inst *Instance) float64 {
	return planetEarthAwayNWithInstance(5, jde, n, inst)
}

// UranusLWithInstance 天王星日心黄经 / heliocentric longitude of Uranus with an explicit ephemeris instance.
func UranusLWithInstance(jde float64, inst *Instance) float64 {
	return planetLWithInstance(6, jde, inst)
}

// UranusBWithInstance 天王星日心黄纬 / heliocentric latitude of Uranus with an explicit ephemeris instance.
func UranusBWithInstance(jde float64, inst *Instance) float64 {
	return planetBWithInstance(6, jde, inst)
}

// UranusRWithInstance 天王星日心距离 / heliocentric distance of Uranus with an explicit ephemeris instance.
func UranusRWithInstance(jde float64, inst *Instance) float64 {
	return planetRWithInstance(6, jde, inst)
}

// AUranusXYZWithInstance 天王星地心直角坐标 / geocentric rectangular coordinates of Uranus with an explicit ephemeris instance.
func AUranusXYZWithInstance(jde float64, inst *Instance) (float64, float64, float64) {
	x, y, z, _ := planetGeocentricXYZWithInstance(6, jde, inst)
	return x, y, z
}

// AUranusXWithInstance 天王星地心 X 坐标 / geocentric X coordinate of Uranus with an explicit ephemeris instance.
func AUranusXWithInstance(jde float64, inst *Instance) float64 {
	x, _, _ := AUranusXYZWithInstance(jde, inst)
	return x
}

// AUranusYWithInstance 天王星地心 Y 坐标 / geocentric Y coordinate of Uranus with an explicit ephemeris instance.
func AUranusYWithInstance(jde float64, inst *Instance) float64 {
	_, y, _ := AUranusXYZWithInstance(jde, inst)
	return y
}

// AUranusZWithInstance 天王星地心 Z 坐标 / geocentric Z coordinate of Uranus with an explicit ephemeris instance.
func AUranusZWithInstance(jde float64, inst *Instance) float64 {
	_, _, z := AUranusXYZWithInstance(jde, inst)
	return z
}

// UranusApparentLoBoWithInstance 天王星视黄经纬 / apparent ecliptic longitude and latitude of Uranus with an explicit ephemeris instance.
func UranusApparentLoBoWithInstance(jde float64, inst *Instance) (float64, float64) {
	return planetApparentLoBoWithInstance(6, jde, inst)
}

// UranusApparentLoWithInstance 天王星视黄经 / apparent ecliptic longitude of Uranus with an explicit ephemeris instance.
func UranusApparentLoWithInstance(jde float64, inst *Instance) float64 {
	lo, _ := UranusApparentLoBoWithInstance(jde, inst)
	return lo
}

// UranusApparentBoWithInstance 天王星视黄纬 / apparent ecliptic latitude of Uranus with an explicit ephemeris instance.
func UranusApparentBoWithInstance(jde float64, inst *Instance) float64 {
	_, bo := UranusApparentLoBoWithInstance(jde, inst)
	return bo
}

// UranusApparentRaDecWithInstance 天王星视赤经纬 / apparent right ascension and declination of Uranus with an explicit ephemeris instance.
func UranusApparentRaDecWithInstance(jde float64, inst *Instance) (float64, float64) {
	return planetApparentRaDecWithInstance(6, jde, inst)
}

// UranusApparentRaWithInstance 天王星视赤经 / apparent right ascension of Uranus with an explicit ephemeris instance.
func UranusApparentRaWithInstance(jde float64, inst *Instance) float64 {
	ra, _ := UranusApparentRaDecWithInstance(jde, inst)
	return ra
}

// UranusApparentDecWithInstance 天王星视赤纬 / apparent declination of Uranus with an explicit ephemeris instance.
func UranusApparentDecWithInstance(jde float64, inst *Instance) float64 {
	_, dec := UranusApparentRaDecWithInstance(jde, inst)
	return dec
}

// UranusApparentLoBoNWithInstance 天王星视黄经纬（n 变体） / n-parameterised apparent ecliptic longitude and latitude of Uranus with an explicit ephemeris instance.
func UranusApparentLoBoNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	return planetApparentLoBoNWithInstance(6, jde, n, inst)
}

// UranusApparentLoNWithInstance 天王星视黄经（n 变体） / n-parameterised apparent ecliptic longitude of Uranus with an explicit ephemeris instance.
func UranusApparentLoNWithInstance(jde float64, n int, inst *Instance) float64 {
	lo, _ := UranusApparentLoBoNWithInstance(jde, n, inst)
	return lo
}

// UranusApparentBoNWithInstance 天王星视黄纬（n 变体） / n-parameterised apparent ecliptic latitude of Uranus with an explicit ephemeris instance.
func UranusApparentBoNWithInstance(jde float64, n int, inst *Instance) float64 {
	_, bo := UranusApparentLoBoNWithInstance(jde, n, inst)
	return bo
}

// UranusApparentRaDecNWithInstance 天王星视赤经纬（n 变体） / n-parameterised apparent right ascension and declination of Uranus with an explicit ephemeris instance.
func UranusApparentRaDecNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	lo, bo := UranusApparentLoBoNWithInstance(jde, n, inst)
	return planetApparentRaDecFromLoBo(jde, lo, bo)
}

// UranusApparentRaNWithInstance 天王星视赤经（n 变体） / n-parameterised apparent right ascension of Uranus with an explicit ephemeris instance.
func UranusApparentRaNWithInstance(jde float64, n int, inst *Instance) float64 {
	ra, _ := UranusApparentRaDecNWithInstance(jde, n, inst)
	return ra
}

// UranusApparentDecNWithInstance 天王星视赤纬（n 变体） / n-parameterised apparent declination of Uranus with an explicit ephemeris instance.
func UranusApparentDecNWithInstance(jde float64, n int, inst *Instance) float64 {
	_, dec := UranusApparentRaDecNWithInstance(jde, n, inst)
	return dec
}

// EarthUranusAwayWithInstance 地天王距离 / Earth-Uranus distance with an explicit ephemeris instance.
func EarthUranusAwayWithInstance(jde float64, inst *Instance) float64 {
	return planetEarthAwayWithInstance(6, jde, inst)
}

// EarthUranusAwayNWithInstance 天王星地心距离（n 变体） / n-parameterised Earth-Uranus distance with an explicit ephemeris instance.
func EarthUranusAwayNWithInstance(jde float64, n int, inst *Instance) float64 {
	return planetEarthAwayNWithInstance(6, jde, n, inst)
}

// NeptuneLWithInstance 海王星日心黄经 / heliocentric longitude of Neptune with an explicit ephemeris instance.
func NeptuneLWithInstance(jde float64, inst *Instance) float64 {
	return planetLWithInstance(7, jde, inst)
}

// NeptuneBWithInstance 海王星日心黄纬 / heliocentric latitude of Neptune with an explicit ephemeris instance.
func NeptuneBWithInstance(jde float64, inst *Instance) float64 {
	return planetBWithInstance(7, jde, inst)
}

// NeptuneRWithInstance 海王星日心距离 / heliocentric distance of Neptune with an explicit ephemeris instance.
func NeptuneRWithInstance(jde float64, inst *Instance) float64 {
	return planetRWithInstance(7, jde, inst)
}

// ANeptuneXYZWithInstance 海王星地心直角坐标 / geocentric rectangular coordinates of Neptune with an explicit ephemeris instance.
func ANeptuneXYZWithInstance(jde float64, inst *Instance) (float64, float64, float64) {
	x, y, z, _ := planetGeocentricXYZWithInstance(7, jde, inst)
	return x, y, z
}

// ANeptuneXWithInstance 海王星地心 X 坐标 / geocentric X coordinate of Neptune with an explicit ephemeris instance.
func ANeptuneXWithInstance(jde float64, inst *Instance) float64 {
	x, _, _ := ANeptuneXYZWithInstance(jde, inst)
	return x
}

// ANeptuneYWithInstance 海王星地心 Y 坐标 / geocentric Y coordinate of Neptune with an explicit ephemeris instance.
func ANeptuneYWithInstance(jde float64, inst *Instance) float64 {
	_, y, _ := ANeptuneXYZWithInstance(jde, inst)
	return y
}

// ANeptuneZWithInstance 海王星地心 Z 坐标 / geocentric Z coordinate of Neptune with an explicit ephemeris instance.
func ANeptuneZWithInstance(jde float64, inst *Instance) float64 {
	_, _, z := ANeptuneXYZWithInstance(jde, inst)
	return z
}

// NeptuneApparentLoBoWithInstance 海王星视黄经纬 / apparent ecliptic longitude and latitude of Neptune with an explicit ephemeris instance.
func NeptuneApparentLoBoWithInstance(jde float64, inst *Instance) (float64, float64) {
	return planetApparentLoBoWithInstance(7, jde, inst)
}

// NeptuneApparentLoWithInstance 海王星视黄经 / apparent ecliptic longitude of Neptune with an explicit ephemeris instance.
func NeptuneApparentLoWithInstance(jde float64, inst *Instance) float64 {
	lo, _ := NeptuneApparentLoBoWithInstance(jde, inst)
	return lo
}

// NeptuneApparentBoWithInstance 海王星视黄纬 / apparent ecliptic latitude of Neptune with an explicit ephemeris instance.
func NeptuneApparentBoWithInstance(jde float64, inst *Instance) float64 {
	_, bo := NeptuneApparentLoBoWithInstance(jde, inst)
	return bo
}

// NeptuneApparentRaDecWithInstance 海王星视赤经纬 / apparent right ascension and declination of Neptune with an explicit ephemeris instance.
func NeptuneApparentRaDecWithInstance(jde float64, inst *Instance) (float64, float64) {
	return planetApparentRaDecWithInstance(7, jde, inst)
}

// NeptuneApparentRaWithInstance 海王星视赤经 / apparent right ascension of Neptune with an explicit ephemeris instance.
func NeptuneApparentRaWithInstance(jde float64, inst *Instance) float64 {
	ra, _ := NeptuneApparentRaDecWithInstance(jde, inst)
	return ra
}

// NeptuneApparentDecWithInstance 海王星视赤纬 / apparent declination of Neptune with an explicit ephemeris instance.
func NeptuneApparentDecWithInstance(jde float64, inst *Instance) float64 {
	_, dec := NeptuneApparentRaDecWithInstance(jde, inst)
	return dec
}

// NeptuneApparentLoBoNWithInstance 海王星视黄经纬（n 变体） / n-parameterised apparent ecliptic longitude and latitude of Neptune with an explicit ephemeris instance.
func NeptuneApparentLoBoNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	return planetApparentLoBoNWithInstance(7, jde, n, inst)
}

// NeptuneApparentLoNWithInstance 海王星视黄经（n 变体） / n-parameterised apparent ecliptic longitude of Neptune with an explicit ephemeris instance.
func NeptuneApparentLoNWithInstance(jde float64, n int, inst *Instance) float64 {
	lo, _ := NeptuneApparentLoBoNWithInstance(jde, n, inst)
	return lo
}

// NeptuneApparentBoNWithInstance 海王星视黄纬（n 变体） / n-parameterised apparent ecliptic latitude of Neptune with an explicit ephemeris instance.
func NeptuneApparentBoNWithInstance(jde float64, n int, inst *Instance) float64 {
	_, bo := NeptuneApparentLoBoNWithInstance(jde, n, inst)
	return bo
}

// NeptuneApparentRaDecNWithInstance 海王星视赤经纬（n 变体） / n-parameterised apparent right ascension and declination of Neptune with an explicit ephemeris instance.
func NeptuneApparentRaDecNWithInstance(jde float64, n int, inst *Instance) (float64, float64) {
	lo, bo := NeptuneApparentLoBoNWithInstance(jde, n, inst)
	return planetApparentRaDecFromLoBo(jde, lo, bo)
}

// NeptuneApparentRaNWithInstance 海王星视赤经（n 变体） / n-parameterised apparent right ascension of Neptune with an explicit ephemeris instance.
func NeptuneApparentRaNWithInstance(jde float64, n int, inst *Instance) float64 {
	ra, _ := NeptuneApparentRaDecNWithInstance(jde, n, inst)
	return ra
}

// NeptuneApparentDecNWithInstance 海王星视赤纬（n 变体） / n-parameterised apparent declination of Neptune with an explicit ephemeris instance.
func NeptuneApparentDecNWithInstance(jde float64, n int, inst *Instance) float64 {
	_, dec := NeptuneApparentRaDecNWithInstance(jde, n, inst)
	return dec
}

// EarthNeptuneAwayWithInstance 地海距离 / Earth-Neptune distance with an explicit ephemeris instance.
func EarthNeptuneAwayWithInstance(jde float64, inst *Instance) float64 {
	return planetEarthAwayWithInstance(7, jde, inst)
}

// EarthNeptuneAwayNWithInstance 海王星地心距离（n 变体） / n-parameterised Earth-Neptune distance with an explicit ephemeris instance.
func EarthNeptuneAwayNWithInstance(jde float64, n int, inst *Instance) float64 {
	return planetEarthAwayNWithInstance(7, jde, n, inst)
}

// MercuryMagWithInstance 水星视星等 / apparent magnitude of Mercury with an explicit ephemeris instance.
func MercuryMagWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return MercuryMag(jde)
	}
	return planetMagnitudeWithInstance(1, jde, inst, -0.42, 0.0380, -0.000273, 0.000002)
}

// VenusMagWithInstance 金星视星等 / apparent magnitude of Venus with an explicit ephemeris instance.
func VenusMagWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return VenusMag(jde)
	}
	return planetMagnitudeWithInstance(2, jde, inst, -4.40, 0.0009, 0.000239, -0.00000065)
}

// MarsMagWithInstance 火星视星等 / apparent magnitude of Mars with an explicit ephemeris instance.
func MarsMagWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return MarsMag(jde)
	}
	return planetMagnitudeWithInstance(3, jde, inst, -1.52, 0.016, 0, 0)
}

// JupiterMagWithInstance 木星视星等 / apparent magnitude of Jupiter with an explicit ephemeris instance.
func JupiterMagWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return JupiterMag(jde)
	}
	return planetMagnitudeWithInstance(4, jde, inst, -9.40, 0.0005, 0, 0)
}

// UranusMagWithInstance 天王星视星等 / apparent magnitude of Uranus with an explicit ephemeris instance.
func UranusMagWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return UranusMag(jde)
	}
	return planetMagnitudeWithInstance(6, jde, inst, -7.19, 0.016, 0, 0)
}

// NeptuneMagWithInstance 海王星视星等 / apparent magnitude of Neptune with an explicit ephemeris instance.
func NeptuneMagWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return NeptuneMag(jde)
	}
	sunDistance := planetRWithInstance(7, jde, inst)
	earthDistance := planetEarthAwayWithInstance(7, jde, inst)
	return FloatRound(-6.87+5*math.Log10(sunDistance*earthDistance), 2)
}
