package basic

import (
	"math"

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

// MoonAscendingNodeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MoonAscendingNodeNWithInstance(jd float64, n int, inst *Instance) float64 {
	return orbitalAscendingNodeLongitude(jd, n, orbitalNodeStepFastBodyDays, func(jde float64, n int) Vector3 { return moonGeocentricNodePositionN(jde, n, inst) })
}

// MoonDescendingNode 月球降交点黄经 / descending node longitude of the Moon.
func MoonDescendingNode(jd float64) float64 {
	return MoonDescendingNodeN(jd, -1)
}

// MoonDescendingNodeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MoonDescendingNodeNWithInstance(jd float64, n int, inst *Instance) float64 {
	return Limit360(MoonAscendingNodeNWithInstance(jd, n, inst) + 180)
}

// MercuryAscendingNode 水星升交点黄经 / ascending node longitude of Mercury.
func MercuryAscendingNode(jd float64) float64 {
	return MercuryAscendingNodeN(jd, -1)
}

// MercuryAscendingNodeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercuryAscendingNodeNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetAscendingNodeLongitudeN(1, jd, n, inst)
}

// MercuryDescendingNode 水星降交点黄经 / descending node longitude of Mercury.
func MercuryDescendingNode(jd float64) float64 {
	return MercuryDescendingNodeN(jd, -1)
}

// MercuryDescendingNodeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercuryDescendingNodeNWithInstance(jd float64, n int, inst *Instance) float64 {
	return Limit360(MercuryAscendingNodeNWithInstance(jd, n, inst) + 180)
}

// VenusAscendingNode 金星升交点黄经 / ascending node longitude of Venus.
func VenusAscendingNode(jd float64) float64 {
	return VenusAscendingNodeN(jd, -1)
}

// VenusAscendingNodeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func VenusAscendingNodeNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetAscendingNodeLongitudeN(2, jd, n, inst)
}

// VenusDescendingNode 金星降交点黄经 / descending node longitude of Venus.
func VenusDescendingNode(jd float64) float64 {
	return VenusDescendingNodeN(jd, -1)
}

// VenusDescendingNodeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func VenusDescendingNodeNWithInstance(jd float64, n int, inst *Instance) float64 {
	return Limit360(VenusAscendingNodeNWithInstance(jd, n, inst) + 180)
}

// MarsAscendingNode 火星升交点黄经 / ascending node longitude of Mars.
func MarsAscendingNode(jd float64) float64 {
	return MarsAscendingNodeN(jd, -1)
}

// MarsAscendingNodeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MarsAscendingNodeNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetAscendingNodeLongitudeN(3, jd, n, inst)
}

// MarsDescendingNode 火星降交点黄经 / descending node longitude of Mars.
func MarsDescendingNode(jd float64) float64 {
	return MarsDescendingNodeN(jd, -1)
}

// MarsDescendingNodeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MarsDescendingNodeNWithInstance(jd float64, n int, inst *Instance) float64 {
	return Limit360(MarsAscendingNodeNWithInstance(jd, n, inst) + 180)
}

// JupiterAscendingNode 木星升交点黄经 / ascending node longitude of Jupiter.
func JupiterAscendingNode(jd float64) float64 {
	return JupiterAscendingNodeN(jd, -1)
}

// JupiterAscendingNodeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func JupiterAscendingNodeNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetAscendingNodeLongitudeN(4, jd, n, inst)
}

// JupiterDescendingNode 木星降交点黄经 / descending node longitude of Jupiter.
func JupiterDescendingNode(jd float64) float64 {
	return JupiterDescendingNodeN(jd, -1)
}

// JupiterDescendingNodeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func JupiterDescendingNodeNWithInstance(jd float64, n int, inst *Instance) float64 {
	return Limit360(JupiterAscendingNodeNWithInstance(jd, n, inst) + 180)
}

// SaturnAscendingNode 土星升交点黄经 / ascending node longitude of Saturn.
func SaturnAscendingNode(jd float64) float64 {
	return SaturnAscendingNodeN(jd, -1)
}

// SaturnAscendingNodeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnAscendingNodeNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetAscendingNodeLongitudeN(5, jd, n, inst)
}

// SaturnDescendingNode 土星降交点黄经 / descending node longitude of Saturn.
func SaturnDescendingNode(jd float64) float64 {
	return SaturnDescendingNodeN(jd, -1)
}

// SaturnDescendingNodeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnDescendingNodeNWithInstance(jd float64, n int, inst *Instance) float64 {
	return Limit360(SaturnAscendingNodeNWithInstance(jd, n, inst) + 180)
}

// UranusAscendingNode 天王星升交点黄经 / ascending node longitude of Uranus.
func UranusAscendingNode(jd float64) float64 {
	return UranusAscendingNodeN(jd, -1)
}

// UranusAscendingNodeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusAscendingNodeNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetAscendingNodeLongitudeN(6, jd, n, inst)
}

// UranusDescendingNode 天王星降交点黄经 / descending node longitude of Uranus.
func UranusDescendingNode(jd float64) float64 {
	return UranusDescendingNodeN(jd, -1)
}

// UranusDescendingNodeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusDescendingNodeNWithInstance(jd float64, n int, inst *Instance) float64 {
	return Limit360(UranusAscendingNodeNWithInstance(jd, n, inst) + 180)
}

// NeptuneAscendingNode 海王星升交点黄经 / ascending node longitude of Neptune.
func NeptuneAscendingNode(jd float64) float64 {
	return NeptuneAscendingNodeN(jd, -1)
}

// NeptuneAscendingNodeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NeptuneAscendingNodeNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetAscendingNodeLongitudeN(7, jd, n, inst)
}

// NeptuneDescendingNode 海王星降交点黄经 / descending node longitude of Neptune.
func NeptuneDescendingNode(jd float64) float64 {
	return NeptuneDescendingNodeN(jd, -1)
}

// NeptuneDescendingNodeNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NeptuneDescendingNodeNWithInstance(jd float64, n int, inst *Instance) float64 {
	return Limit360(NeptuneAscendingNodeNWithInstance(jd, n, inst) + 180)
}

func planetAscendingNodeLongitudeN(planetIndex int, jd float64, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	step := orbitalNodeStepSlowBodyDays
	if planetIndex <= 3 {
		step = orbitalNodeStepFastBodyDays
	}
	return orbitalAscendingNodeLongitude(jd, n, step, func(sampleJD float64, seriesTerms int) Vector3 {
		return planetHeliocentricNodePositionN(planetIndex, sampleJD, seriesTerms, inst)
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

func eclipticVectorAtReferenceEpoch(vector Vector3, sampleJDE, referenceJDE float64) Vector3 {
	if sampleJDE == referenceJDE {
		return vector
	}

	sampleEquatorial := rotateEclipticToEquatorial(vector, EclipticObliquity(sampleJDE, false))
	precessedEquatorial := applyMatrix3(precessionMatrix(sampleJDE, referenceJDE), sampleEquatorial)
	return rotateEquatorialToEcliptic(precessedEquatorial, EclipticObliquity(referenceJDE, false))
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

func planetHeliocentricNodePositionN(planetIndex int, jde float64, n int, instances ...*Instance) Vector3 {
	inst := instanceValue(instances)

	longitude := PlanetCoordinateNWithInstance(planetIndex, 0, jde, n, inst)
	latitude := PlanetCoordinateNWithInstance(planetIndex, 1, jde, n, inst)
	radius := PlanetCoordinateNWithInstance(planetIndex, 2, jde, n, inst)
	return eclipticCartesian(longitude, latitude, radius)
}

func moonGeocentricNodePositionN(jde float64, n int, instances ...*Instance) Vector3 {
	inst := instanceValue(instances)

	longitude := HMoonTrueLoNWithInstance(jde, n, inst)
	latitude := HMoonTrueBoNWithInstance(jde, n, inst)
	radius := HMoonAwayNWithInstance(jde, n, inst)
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
