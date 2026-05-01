package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

const (
	// 月球和内行星轨道角速度更快，取更小的中心差分步长。
	orbitalNodeStepFastBodyDays = 0.005
	// 外行星保留更稳健的 0.01 day 步长，避免无意义地放大数值噪声。
	orbitalNodeStepSlowBodyDays = 0.01
)

// MoonAscendingNode 月球升交点黄经 / ascending node longitude of the Moon.
func MoonAscendingNode(jd float64) float64 {
	return MoonAscendingNodeN(jd, -1)
}

// MoonAscendingNodeN 月球升交点黄经（截断版） / truncated ascending node longitude of the Moon.
func MoonAscendingNodeN(jd float64, n int) float64 {
	return orbitalAscendingNodeLongitude(jd, n, orbitalNodeStepFastBodyDays, moonGeocentricNodePositionN)
}

// MoonDescendingNode 月球降交点黄经 / descending node longitude of the Moon.
func MoonDescendingNode(jd float64) float64 {
	return MoonDescendingNodeN(jd, -1)
}

// MoonDescendingNodeN 月球降交点黄经（截断版） / truncated descending node longitude of the Moon.
func MoonDescendingNodeN(jd float64, n int) float64 {
	return Limit360(MoonAscendingNodeN(jd, n) + 180)
}

// MercuryAscendingNode 水星升交点黄经 / ascending node longitude of Mercury.
func MercuryAscendingNode(jd float64) float64 {
	return MercuryAscendingNodeN(jd, -1)
}

// MercuryAscendingNodeN 水星升交点黄经（截断版） / truncated ascending node longitude of Mercury.
func MercuryAscendingNodeN(jd float64, n int) float64 {
	return planetAscendingNodeLongitudeN(1, jd, n)
}

// MercuryDescendingNode 水星降交点黄经 / descending node longitude of Mercury.
func MercuryDescendingNode(jd float64) float64 {
	return MercuryDescendingNodeN(jd, -1)
}

// MercuryDescendingNodeN 水星降交点黄经（截断版） / truncated descending node longitude of Mercury.
func MercuryDescendingNodeN(jd float64, n int) float64 {
	return Limit360(MercuryAscendingNodeN(jd, n) + 180)
}

// VenusAscendingNode 金星升交点黄经 / ascending node longitude of Venus.
func VenusAscendingNode(jd float64) float64 {
	return VenusAscendingNodeN(jd, -1)
}

// VenusAscendingNodeN 金星升交点黄经（截断版） / truncated ascending node longitude of Venus.
func VenusAscendingNodeN(jd float64, n int) float64 {
	return planetAscendingNodeLongitudeN(2, jd, n)
}

// VenusDescendingNode 金星降交点黄经 / descending node longitude of Venus.
func VenusDescendingNode(jd float64) float64 {
	return VenusDescendingNodeN(jd, -1)
}

// VenusDescendingNodeN 金星降交点黄经（截断版） / truncated descending node longitude of Venus.
func VenusDescendingNodeN(jd float64, n int) float64 {
	return Limit360(VenusAscendingNodeN(jd, n) + 180)
}

// MarsAscendingNode 火星升交点黄经 / ascending node longitude of Mars.
func MarsAscendingNode(jd float64) float64 {
	return MarsAscendingNodeN(jd, -1)
}

// MarsAscendingNodeN 火星升交点黄经（截断版） / truncated ascending node longitude of Mars.
func MarsAscendingNodeN(jd float64, n int) float64 {
	return planetAscendingNodeLongitudeN(3, jd, n)
}

// MarsDescendingNode 火星降交点黄经 / descending node longitude of Mars.
func MarsDescendingNode(jd float64) float64 {
	return MarsDescendingNodeN(jd, -1)
}

// MarsDescendingNodeN 火星降交点黄经（截断版） / truncated descending node longitude of Mars.
func MarsDescendingNodeN(jd float64, n int) float64 {
	return Limit360(MarsAscendingNodeN(jd, n) + 180)
}

// JupiterAscendingNode 木星升交点黄经 / ascending node longitude of Jupiter.
func JupiterAscendingNode(jd float64) float64 {
	return JupiterAscendingNodeN(jd, -1)
}

// JupiterAscendingNodeN 木星升交点黄经（截断版） / truncated ascending node longitude of Jupiter.
func JupiterAscendingNodeN(jd float64, n int) float64 {
	return planetAscendingNodeLongitudeN(4, jd, n)
}

// JupiterDescendingNode 木星降交点黄经 / descending node longitude of Jupiter.
func JupiterDescendingNode(jd float64) float64 {
	return JupiterDescendingNodeN(jd, -1)
}

// JupiterDescendingNodeN 木星降交点黄经（截断版） / truncated descending node longitude of Jupiter.
func JupiterDescendingNodeN(jd float64, n int) float64 {
	return Limit360(JupiterAscendingNodeN(jd, n) + 180)
}

// SaturnAscendingNode 土星升交点黄经 / ascending node longitude of Saturn.
func SaturnAscendingNode(jd float64) float64 {
	return SaturnAscendingNodeN(jd, -1)
}

// SaturnAscendingNodeN 土星升交点黄经（截断版） / truncated ascending node longitude of Saturn.
func SaturnAscendingNodeN(jd float64, n int) float64 {
	return planetAscendingNodeLongitudeN(5, jd, n)
}

// SaturnDescendingNode 土星降交点黄经 / descending node longitude of Saturn.
func SaturnDescendingNode(jd float64) float64 {
	return SaturnDescendingNodeN(jd, -1)
}

// SaturnDescendingNodeN 土星降交点黄经（截断版） / truncated descending node longitude of Saturn.
func SaturnDescendingNodeN(jd float64, n int) float64 {
	return Limit360(SaturnAscendingNodeN(jd, n) + 180)
}

// UranusAscendingNode 天王星升交点黄经 / ascending node longitude of Uranus.
func UranusAscendingNode(jd float64) float64 {
	return UranusAscendingNodeN(jd, -1)
}

// UranusAscendingNodeN 天王星升交点黄经（截断版） / truncated ascending node longitude of Uranus.
func UranusAscendingNodeN(jd float64, n int) float64 {
	return planetAscendingNodeLongitudeN(6, jd, n)
}

// UranusDescendingNode 天王星降交点黄经 / descending node longitude of Uranus.
func UranusDescendingNode(jd float64) float64 {
	return UranusDescendingNodeN(jd, -1)
}

// UranusDescendingNodeN 天王星降交点黄经（截断版） / truncated descending node longitude of Uranus.
func UranusDescendingNodeN(jd float64, n int) float64 {
	return Limit360(UranusAscendingNodeN(jd, n) + 180)
}

// NeptuneAscendingNode 海王星升交点黄经 / ascending node longitude of Neptune.
func NeptuneAscendingNode(jd float64) float64 {
	return NeptuneAscendingNodeN(jd, -1)
}

// NeptuneAscendingNodeN 海王星升交点黄经（截断版） / truncated ascending node longitude of Neptune.
func NeptuneAscendingNodeN(jd float64, n int) float64 {
	return planetAscendingNodeLongitudeN(7, jd, n)
}

// NeptuneDescendingNode 海王星降交点黄经 / descending node longitude of Neptune.
func NeptuneDescendingNode(jd float64) float64 {
	return NeptuneDescendingNodeN(jd, -1)
}

// NeptuneDescendingNodeN 海王星降交点黄经（截断版） / truncated descending node longitude of Neptune.
func NeptuneDescendingNodeN(jd float64, n int) float64 {
	return Limit360(NeptuneAscendingNodeN(jd, n) + 180)
}

func planetAscendingNodeLongitudeN(planetIndex int, jd float64, n int) float64 {
	step := orbitalNodeStepSlowBodyDays
	if planetIndex <= 3 {
		step = orbitalNodeStepFastBodyDays
	}
	return orbitalAscendingNodeLongitude(jd, n, step, func(sampleJD float64, seriesTerms int) Vector3 {
		return planetHeliocentricNodePositionN(planetIndex, sampleJD, seriesTerms)
	})
}

func orbitalAscendingNodeLongitude(jd float64, n int, step float64, position func(float64, int) Vector3) float64 {
	current := eclipticVectorAtReferenceEpoch(position(jd, n), jd, jd)
	previous := eclipticVectorAtReferenceEpoch(position(jd-step, n), jd-step, jd)
	next := eclipticVectorAtReferenceEpoch(position(jd+step, n), jd+step, jd)

	velocity := Vector3{
		(next[0] - previous[0]) / (2 * step),
		(next[1] - previous[1]) / (2 * step),
		(next[2] - previous[2]) / (2 * step),
	}
	angularMomentum := pxp(current, velocity)
	nodeVector, magnitude := pn(Vector3{-angularMomentum[1], angularMomentum[0], 0})
	if magnitude == 0 {
		return 0
	}
	return Limit360(math.Atan2(nodeVector[1], nodeVector[0]) * deg)
}

func eclipticVectorAtReferenceEpoch(vector Vector3, sampleJD, referenceJD float64) Vector3 {
	if sampleJD == referenceJD {
		return vector
	}

	sampleEquatorial := rotateEclipticToEquatorial(vector, EclipticObliquity(sampleJD, false))
	precessedEquatorial := applyMatrix3(precessionMatrix(sampleJD, referenceJD), sampleEquatorial)
	return rotateEquatorialToEcliptic(precessedEquatorial, EclipticObliquity(referenceJD, false))
}

func rotateEclipticToEquatorial(vector Vector3, obliquity float64) Vector3 {
	epsilon := obliquity * rad
	cosEpsilon := math.Cos(epsilon)
	sinEpsilon := math.Sin(epsilon)
	return Vector3{
		vector[0],
		vector[1]*cosEpsilon - vector[2]*sinEpsilon,
		vector[1]*sinEpsilon + vector[2]*cosEpsilon,
	}
}

func rotateEquatorialToEcliptic(vector Vector3, obliquity float64) Vector3 {
	epsilon := obliquity * rad
	cosEpsilon := math.Cos(epsilon)
	sinEpsilon := math.Sin(epsilon)
	return Vector3{
		vector[0],
		vector[1]*cosEpsilon + vector[2]*sinEpsilon,
		-vector[1]*sinEpsilon + vector[2]*cosEpsilon,
	}
}

func precessionMatrix(jdFrom, jdTo float64) Matrix3 {
	epjFrom := 2000.0 + (jdFrom-2451545.0)/365.25
	epjTo := 2000.0 + (jdTo-2451545.0)/365.25

	rpFrom := ltpPMAT(epjFrom)
	rpFromInv := Matrix3{
		{rpFrom[0][0], rpFrom[1][0], rpFrom[2][0]},
		{rpFrom[0][1], rpFrom[1][1], rpFrom[2][1]},
		{rpFrom[0][2], rpFrom[1][2], rpFrom[2][2]},
	}
	rpTo := ltpPMAT(epjTo)

	var result Matrix3
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				result[i][j] += rpTo[i][k] * rpFromInv[k][j]
			}
		}
	}
	return result
}

func applyMatrix3(matrix Matrix3, vector Vector3) Vector3 {
	return Vector3{
		matrix[0][0]*vector[0] + matrix[0][1]*vector[1] + matrix[0][2]*vector[2],
		matrix[1][0]*vector[0] + matrix[1][1]*vector[1] + matrix[1][2]*vector[2],
		matrix[2][0]*vector[0] + matrix[2][1]*vector[1] + matrix[2][2]*vector[2],
	}
}

func planetHeliocentricNodePositionN(planetIndex int, jd float64, n int) Vector3 {
	longitude := planet.WherePlanetN(planetIndex, 0, jd, n)
	latitude := planet.WherePlanetN(planetIndex, 1, jd, n)
	radius := planet.WherePlanetN(planetIndex, 2, jd, n)
	return eclipticCartesian(longitude, latitude, radius)
}

func moonGeocentricNodePositionN(jd float64, n int) Vector3 {
	longitude := HMoonTrueLoN(jd, n)
	latitude := HMoonTrueBoN(jd, n)
	radius := HMoonAwayN(jd, n)
	return eclipticCartesian(longitude, latitude, radius)
}

func eclipticCartesian(longitude, latitude, radius float64) Vector3 {
	cosLatitude := Cos(latitude)
	return Vector3{
		radius * cosLatitude * Cos(longitude),
		radius * cosLatitude * Sin(longitude),
		radius * Sin(latitude),
	}
}
