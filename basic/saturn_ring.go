package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

// SaturnRingParameters 土星环参数 / Saturn ring parameters.
func SaturnRingParameters(jd float64) (earthLatitude, sunLatitude, positionAngle, deltaU, majorAxis, minorAxis float64) {
	return SaturnRingParametersN(jd, -1)
}

// SaturnRingParametersN 土星环参数（截断版） / truncated Saturn ring parameters.
func SaturnRingParametersN(jd float64, n int) (earthLatitude, sunLatitude, positionAngle, deltaU, majorAxis, minorAxis float64) {
	inclination, node := saturnRingPlane(jd)
	earthLon, earthLat := SaturnApparentLoBoN(jd, n)
	sunLon, sunLat := saturnRingSunLoBo(jd, n, node)

	earthLatitude = saturnRingLatitude(inclination, node, earthLon, earthLat)
	sunLatitude = saturnRingLatitude(inclination, node, sunLon, sunLat)
	positionAngle = saturnRingPositionAngle(jd, inclination, node, earthLon, earthLat)
	earthU := saturnRingLongitude(inclination, node, earthLon, earthLat)
	sunU := saturnRingLongitude(inclination, node, sunLon, sunLat)
	deltaU = saturnRingLongitudeDelta(earthU, sunU)

	earthDistance := EarthSaturnAwayN(jd, n)
	majorAxis = 375.35 / earthDistance
	minorAxis = majorAxis * math.Abs(Sin(earthLatitude))
	return
}

// SaturnRingB 土星环张角 B / Saturn ring opening angle B.
func SaturnRingB(jd float64) float64 {
	return SaturnRingBN(jd, -1)
}

// SaturnRingBN 土星环张角 B（截断版） / truncated Saturn ring opening angle B.
func SaturnRingBN(jd float64, n int) float64 {
	earthLatitude, _, _, _, _, _ := SaturnRingParametersN(jd, n)
	return earthLatitude
}

// SaturnRingSunB 土星环太阳侧张角 B' / Sun-side Saturn ring opening angle B'.
func SaturnRingSunB(jd float64) float64 {
	return SaturnRingSunBN(jd, -1)
}

// SaturnRingSunBN 土星环太阳侧张角 B'（截断版） / truncated Sun-side Saturn ring opening angle B'.
func SaturnRingSunBN(jd float64, n int) float64 {
	_, sunLatitude, _, _, _, _ := SaturnRingParametersN(jd, n)
	return sunLatitude
}

// SaturnRingPositionAngle 土星环北半短轴位置角 / position angle of Saturn ring northern semiminor axis.
func SaturnRingPositionAngle(jd float64) float64 {
	return SaturnRingPositionAngleN(jd, -1)
}

// SaturnRingPositionAngleN 土星环北半短轴位置角（截断版） / truncated position angle of Saturn ring northern semiminor axis.
func SaturnRingPositionAngleN(jd float64, n int) float64 {
	_, _, positionAngle, _, _, _ := SaturnRingParametersN(jd, n)
	return positionAngle
}

// SaturnRingDeltaU 太阳和地球在环面内的土星心黄经差 / difference of Saturnicentric ring longitudes.
func SaturnRingDeltaU(jd float64) float64 {
	return SaturnRingDeltaUN(jd, -1)
}

// SaturnRingDeltaUN 太阳和地球在环面内的土星心黄经差（截断版） / truncated Saturnicentric ring longitude difference.
func SaturnRingDeltaUN(jd float64, n int) float64 {
	_, _, _, deltaU, _, _ := SaturnRingParametersN(jd, n)
	return deltaU
}

// SaturnRingAxis 土星环外缘长短轴，单位角秒 / outer ring axes in arcseconds.
func SaturnRingAxis(jd float64) (majorAxis, minorAxis float64) {
	return SaturnRingAxisN(jd, -1)
}

// SaturnRingAxisN 土星环外缘长短轴（截断版），单位角秒 / truncated outer ring axes in arcseconds.
func SaturnRingAxisN(jd float64, n int) (majorAxis, minorAxis float64) {
	_, _, _, _, majorAxis, minorAxis = SaturnRingParametersN(jd, n)
	return
}

func saturnRingPlane(jd float64) (inclination, node float64) {
	t := (jd - 2451545.0) / 36525.0
	inclination = 28.075216 - 0.012998*t + 0.000004*t*t
	node = 169.508470 + 1.394681*t + 0.000412*t*t
	return
}

func saturnRingSunLoBo(jd float64, n int, node float64) (lon, lat float64) {
	lon = planet.WherePlanetN(5, 0, jd, n)
	lat = planet.WherePlanetN(5, 1, jd, n)
	distance := planet.WherePlanetN(5, 2, jd, n)
	lat -= 0.000764 * Cos(lon-node) / distance
	lon -= 0.01759 / distance
	return lon, lat
}

func saturnRingLatitude(inclination, node, lon, lat float64) float64 {
	return ArcSin(Sin(inclination)*Cos(lat)*Sin(lon-node) - Cos(inclination)*Sin(lat))
}

func saturnRingLongitude(inclination, node, lon, lat float64) float64 {
	y := Sin(inclination)*Sin(lat) + Cos(inclination)*Cos(lat)*Sin(lon-node)
	x := Cos(lat) * Cos(lon-node)
	return ArcTan2(y, x)
}

func saturnRingLongitudeDelta(a, b float64) float64 {
	delta := math.Abs(Limit360(a - b))
	if delta > 180 {
		return 360 - delta
	}
	return delta
}

func saturnRingPositionAngle(jd, inclination, node, lon, lat float64) float64 {
	poleLon := node - 90 + Nutation2000Bi(jd)
	poleLat := 90 - inclination
	eps := TrueObliquity(jd)
	poleRa, poleDec := saturnRingEclipticToEquatorial(poleLon, poleLat, eps)
	saturnRa, saturnDec := saturnRingEclipticToEquatorial(lon, lat, eps)

	y := Cos(poleDec) * Sin(poleRa-saturnRa)
	x := Sin(poleDec)*Cos(saturnDec) - Cos(poleDec)*Sin(saturnDec)*Cos(poleRa-saturnRa)
	return ArcTan2(y, x)
}

func saturnRingEclipticToEquatorial(lon, lat, eps float64) (ra, dec float64) {
	ra = ArcTan2(Sin(lon)*Cos(eps)-Tan(lat)*Sin(eps), Cos(lon))
	dec = ArcSin(Sin(lat)*Cos(eps) + Cos(lat)*Sin(eps)*Sin(lon))
	return
}
