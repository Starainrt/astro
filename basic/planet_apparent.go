package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

type planetGeocentricPosition struct {
	x  float64
	y  float64
	z  float64
	lo float64
	bo float64
}

func planetHeliocentricXYZN(planetIndex int, jde float64, n int, instances ...*Instance) (float64, float64, float64) {
	inst := instanceValue(instances)

	l := PlanetCoordinateNWithInstance(planetIndex, 0, jde, n, inst)
	b := PlanetCoordinateNWithInstance(planetIndex, 1, jde, n, inst)
	r := PlanetCoordinateNWithInstance(planetIndex, 2, jde, n, inst)
	return sphericalToRectangular(l, b, r)
}

func earthHeliocentricXYZN(jde float64, n int, instances ...*Instance) (float64, float64, float64) {
	inst := instanceValue(instances)

	l := PlanetCoordinateNWithInstance(-1, 0, jde, n, inst)
	b := PlanetCoordinateNWithInstance(-1, 1, jde, n, inst)
	r := PlanetCoordinateNWithInstance(-1, 2, jde, n, inst)
	return sphericalToRectangular(l, b, r)
}

func sphericalToRectangular(lo, bo, radius float64) (float64, float64, float64) {
	cosBo := math.Cos(bo * math.Pi / 180)
	return radius * cosBo * math.Cos(lo*math.Pi/180),
		radius * cosBo * math.Sin(lo*math.Pi/180),
		radius * math.Sin(bo*math.Pi/180)
}

func geocentricPositionFromRectangular(x, y, z float64) planetGeocentricPosition {
	lo := math.Atan2(y, x) * 180 / math.Pi
	bo := math.Atan2(z, math.Sqrt(x*x+y*y)) * 180 / math.Pi
	return planetGeocentricPosition{
		x:  x,
		y:  y,
		z:  z,
		lo: Limit360(lo),
		bo: bo,
	}
}

func planetGeocentricPositionN(planetIndex int, planetJD, earthJDE float64, n int) planetGeocentricPosition {
	px, py, pz := planetHeliocentricXYZN(planetIndex, planetJD, n)
	ex, ey, ez := earthHeliocentricXYZN(earthJDE, n)
	return geocentricPositionFromRectangular(px-ex, py-ey, pz-ez)
}

func planetGeocentricPositionWithEarthN(planetIndex int, planetJD float64, ex, ey, ez float64, n int, instances ...*Instance) planetGeocentricPosition {
	inst := instanceValue(instances)

	px, py, pz := planetHeliocentricXYZN(planetIndex, planetJD, n, inst)
	return geocentricPositionFromRectangular(px-ex, py-ey, pz-ez)
}

func planetApparentGeocentricPositionN(planetIndex int, jd float64, n int) (planetGeocentricPosition, float64) {
	geo, tau, _ := planetApparentGeocentricPositionAndDistanceN(planetIndex, jd, n)
	return geo, tau
}

func planetApparentGeocentricPositionAndDistanceN(
	planetIndex int,
	jde float64,
	n int, instances ...*Instance) (planetGeocentricPosition, float64, float64) {
	inst := instanceValue(instances)

	if planetIndex >= 1 && planetIndex <= 7 {
		if lo, bo, tau, dist, ok := boundVectorApparent(planetBody[planetIndex], jde, n, inst); ok {
			x, y, z := sphericalToRectangular(lo, bo, dist)
			return planetGeocentricPosition{x: x, y: y, z: z, lo: lo, bo: bo}, tau, dist
		}
	}
	ex, ey, ez := earthHeliocentricXYZN(jde, n, inst)
	geoNow := planetGeocentricPositionWithEarthN(planetIndex, jde, ex, ey, ez, n, inst)
	distance := math.Sqrt(geoNow.x*geoNow.x + geoNow.y*geoNow.y + geoNow.z*geoNow.z)
	tau := 0.0057755183 * distance
	geo := planetGeocentricPositionWithEarthN(planetIndex, jde-tau, ex, ey, ez, n, inst)
	baseLo := geo.lo
	baseBo := geo.bo
	geo.lo = Limit360(baseLo + GXCLoWithInstance(baseLo, baseBo, jde, inst)/3600.0 + Nutation2000Bi(jde))
	geo.bo = baseBo + GXCBoWithInstance(baseLo, baseBo, jde, inst)/3600.0
	return geo, tau, distance
}

func planetTrueGeocentricPositionN(planetIndex int, jde float64, n int) (planetGeocentricPosition, float64) {
	ex, ey, ez := earthHeliocentricXYZN(jde, n)
	geoNow := planetGeocentricPositionWithEarthN(planetIndex, jde, ex, ey, ez, n)
	tau := 0.0057755183 * math.Sqrt(geoNow.x*geoNow.x+geoNow.y*geoNow.y+geoNow.z*geoNow.z)
	return planetGeocentricPositionWithEarthN(planetIndex, jde-tau, ex, ey, ez, n), tau
}

func planetEarthAwayExplicitN(planetIndex int, jd float64, n int) float64 {
	// 新模型全精度时与视位置管线同源，掩星等组合路径取到的距离逐位一致。
	if planetIndex >= 1 && planetIndex <= 7 {
		if _, _, _, dist, ok := vectorApparent(planetBody[planetIndex], jd, n); ok {
			return dist
		}
	}
	geoNow := planetGeocentricPositionN(planetIndex, jd, jd, n)
	return math.Sqrt(geoNow.x*geoNow.x + geoNow.y*geoNow.y + geoNow.z*geoNow.z)
}
