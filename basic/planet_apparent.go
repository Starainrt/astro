package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

type planetGeocentricPosition struct {
	x  float64
	y  float64
	z  float64
	lo float64
	bo float64
}

func planetHeliocentricXYZN(planetIndex int, jd float64, n int) (float64, float64, float64) {
	l := planet.WherePlanetN(planetIndex, 0, jd, n)
	b := planet.WherePlanetN(planetIndex, 1, jd, n)
	r := planet.WherePlanetN(planetIndex, 2, jd, n)
	return sphericalToRectangular(l, b, r)
}

func earthHeliocentricXYZN(jd float64, n int) (float64, float64, float64) {
	l := planet.WherePlanetN(-1, 0, jd, n)
	b := planet.WherePlanetN(-1, 1, jd, n)
	r := planet.WherePlanetN(-1, 2, jd, n)
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

func planetGeocentricPositionN(planetIndex int, planetJD, earthJD float64, n int) planetGeocentricPosition {
	px, py, pz := planetHeliocentricXYZN(planetIndex, planetJD, n)
	ex, ey, ez := earthHeliocentricXYZN(earthJD, n)
	return geocentricPositionFromRectangular(px-ex, py-ey, pz-ez)
}

func planetGeocentricPositionWithEarthN(planetIndex int, planetJD float64, ex, ey, ez float64, n int) planetGeocentricPosition {
	px, py, pz := planetHeliocentricXYZN(planetIndex, planetJD, n)
	return geocentricPositionFromRectangular(px-ex, py-ey, pz-ez)
}

func planetApparentGeocentricPositionN(planetIndex int, jd float64, n int) (planetGeocentricPosition, float64) {
	ex, ey, ez := earthHeliocentricXYZN(jd, n)
	geoNow := planetGeocentricPositionWithEarthN(planetIndex, jd, ex, ey, ez, n)
	tau := 0.0057755183 * math.Sqrt(geoNow.x*geoNow.x+geoNow.y*geoNow.y+geoNow.z*geoNow.z)
	geo := planetGeocentricPositionWithEarthN(planetIndex, jd-tau, ex, ey, ez, n)
	baseLo := geo.lo
	baseBo := geo.bo
	geo.lo = Limit360(baseLo + GXCLo(baseLo, baseBo, jd)/3600.0 + Nutation2000Bi(jd))
	geo.bo = baseBo + GXCBo(baseLo, baseBo, jd)/3600.0
	return geo, tau
}

func planetTrueGeocentricPositionN(planetIndex int, jd float64, n int) (planetGeocentricPosition, float64) {
	ex, ey, ez := earthHeliocentricXYZN(jd, n)
	geoNow := planetGeocentricPositionWithEarthN(planetIndex, jd, ex, ey, ez, n)
	tau := 0.0057755183 * math.Sqrt(geoNow.x*geoNow.x+geoNow.y*geoNow.y+geoNow.z*geoNow.z)
	return planetGeocentricPositionWithEarthN(planetIndex, jd-tau, ex, ey, ez, n), tau
}

func planetEarthAwayExplicitN(planetIndex int, jd float64, n int) float64 {
	geoNow := planetGeocentricPositionN(planetIndex, jd, jd, n)
	return math.Sqrt(geoNow.x*geoNow.x + geoNow.y*geoNow.y + geoNow.z*geoNow.z)
}
