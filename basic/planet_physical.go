package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

const astronomicalUnitLightTimeDays = 0.0057755183

type planetPhysicalModel struct {
	planetIndex      int
	positiveEast     bool
	equatorialRadius float64
	polarRadius      float64
	poleRotation     func(float64) (ra, dec, rotationEast float64)
}

type phaseAngleModel struct {
	base      float64
	rate      float64
	accelRate float64
}

// PlanetPhysicalInfo 行星物理观测参数 / planetary physical observing parameters.
type PlanetPhysicalInfo struct {
	// SubEarthLongitude 子地经度，单位度；正方向遵循该天体当前 IAU/Horizons 制图约定。
	SubEarthLongitude float64
	// SubEarthLatitude 子地纬度，单位度。
	SubEarthLatitude float64
	// SubSolarLongitude 子日经度，单位度；正方向遵循该天体当前 IAU/Horizons 制图约定。
	SubSolarLongitude float64
	// SubSolarLatitude 子日纬度，单位度。
	SubSolarLatitude float64
	// NorthPolePositionAngle 天体北极位置角，单位度。
	NorthPolePositionAngle float64
}

var (
	mercuryPhysicalModel = planetPhysicalModel{
		planetIndex:      1,
		positiveEast:     false,
		equatorialRadius: 2440.53,
		polarRadius:      2438.26,
		poleRotation:     mercuryPoleRotation,
	}
	venusPhysicalModel = planetPhysicalModel{
		planetIndex:      2,
		positiveEast:     true,
		equatorialRadius: 6051.8,
		polarRadius:      6051.8,
		poleRotation:     venusPoleRotation,
	}
	marsPhysicalModel = planetPhysicalModel{
		planetIndex:      3,
		positiveEast:     false,
		equatorialRadius: 3396.19,
		polarRadius:      3376.20,
		poleRotation:     marsPoleRotation,
	}
	jupiterPhysicalModel = planetPhysicalModel{
		planetIndex:      4,
		positiveEast:     false,
		equatorialRadius: 71492.0,
		polarRadius:      66854.0,
		poleRotation:     jupiterPoleRotation,
	}
	saturnPhysicalModel = planetPhysicalModel{
		planetIndex:      5,
		positiveEast:     false,
		equatorialRadius: 60268.0,
		polarRadius:      54364.0,
		poleRotation:     saturnPoleRotation,
	}
	uranusPhysicalModel = planetPhysicalModel{
		planetIndex:      6,
		positiveEast:     true,
		equatorialRadius: 25559.0,
		polarRadius:      24973.0,
		poleRotation:     uranusPoleRotation,
	}
	neptunePhysicalModel = planetPhysicalModel{
		planetIndex:      7,
		positiveEast:     false,
		equatorialRadius: 24764.0,
		polarRadius:      24341.0,
		poleRotation:     neptunePoleRotation,
	}
)

// MercuryPhysical 水星物理观测参数 / physical observing parameters of Mercury.
func MercuryPhysical(jd float64) PlanetPhysicalInfo {
	return MercuryPhysicalN(jd, -1)
}

// MercuryPhysicalN 水星物理观测参数（截断版） / truncated physical observing parameters of Mercury.
func MercuryPhysicalN(jd float64, n int) PlanetPhysicalInfo {
	return planetPhysicalN(jd, n, mercuryPhysicalModel)
}

// VenusPhysical 金星物理观测参数 / physical observing parameters of Venus.
func VenusPhysical(jd float64) PlanetPhysicalInfo {
	return VenusPhysicalN(jd, -1)
}

// VenusPhysicalN 金星物理观测参数（截断版） / truncated physical observing parameters of Venus.
func VenusPhysicalN(jd float64, n int) PlanetPhysicalInfo {
	return planetPhysicalN(jd, n, venusPhysicalModel)
}

// MarsPhysical 火星物理观测参数 / physical observing parameters of Mars.
func MarsPhysical(jd float64) PlanetPhysicalInfo {
	return MarsPhysicalN(jd, -1)
}

// MarsPhysicalN 火星物理观测参数（截断版） / truncated physical observing parameters of Mars.
func MarsPhysicalN(jd float64, n int) PlanetPhysicalInfo {
	return planetPhysicalN(jd, n, marsPhysicalModel)
}

// JupiterPhysical 木星物理观测参数 / physical observing parameters of Jupiter.
func JupiterPhysical(jd float64) PlanetPhysicalInfo {
	return JupiterPhysicalN(jd, -1)
}

// JupiterPhysicalN 木星物理观测参数（截断版） / truncated physical observing parameters of Jupiter.
func JupiterPhysicalN(jd float64, n int) PlanetPhysicalInfo {
	return planetPhysicalN(jd, n, jupiterPhysicalModel)
}

// SaturnPhysical 土星物理观测参数 / physical observing parameters of Saturn.
func SaturnPhysical(jd float64) PlanetPhysicalInfo {
	return SaturnPhysicalN(jd, -1)
}

// SaturnPhysicalN 土星物理观测参数（截断版） / truncated physical observing parameters of Saturn.
func SaturnPhysicalN(jd float64, n int) PlanetPhysicalInfo {
	return planetPhysicalN(jd, n, saturnPhysicalModel)
}

// UranusPhysical 天王星物理观测参数 / physical observing parameters of Uranus.
func UranusPhysical(jd float64) PlanetPhysicalInfo {
	return UranusPhysicalN(jd, -1)
}

// UranusPhysicalN 天王星物理观测参数（截断版） / truncated physical observing parameters of Uranus.
func UranusPhysicalN(jd float64, n int) PlanetPhysicalInfo {
	return planetPhysicalN(jd, n, uranusPhysicalModel)
}

// NeptunePhysical 海王星物理观测参数 / physical observing parameters of Neptune.
func NeptunePhysical(jd float64) PlanetPhysicalInfo {
	return NeptunePhysicalN(jd, -1)
}

// NeptunePhysicalN 海王星物理观测参数（截断版） / truncated physical observing parameters of Neptune.
func NeptunePhysicalN(jd float64, n int) PlanetPhysicalInfo {
	return planetPhysicalN(jd, n, neptunePhysicalModel)
}

func planetPhysicalN(jd float64, n int, model planetPhysicalModel) PlanetPhysicalInfo {
	initialX, initialY, initialZ := planetXYZN(model.planetIndex, jd, n)
	targetVector := Vector3{initialX, initialY, initialZ}
	lightTimeDays := astronomicalUnitLightTimeDays * vectorMagnitude(targetVector)
	targetJD := jd - lightTimeDays

	geoX, geoY, geoZ := planetXYZN(model.planetIndex, targetJD, n)
	geocentricVector := Vector3{geoX, geoY, geoZ}
	observerDirection := normalizeVector(Vector3{-geocentricVector[0], -geocentricVector[1], -geocentricVector[2]})

	heliocentricLongitude := planet.WherePlanetN(model.planetIndex, 0, targetJD, n)
	heliocentricLatitude := planet.WherePlanetN(model.planetIndex, 1, targetJD, n)
	solarDirection := normalizeVector(eclipticCartesian(heliocentricLongitude+180, -heliocentricLatitude, 1))

	obliquity := EclipticObliquity(targetJD, false)
	observerEquatorial := normalizeVector(rotateEclipticToEquatorial(observerDirection, obliquity))
	solarEquatorial := normalizeVector(rotateEclipticToEquatorial(solarDirection, obliquity))

	poleRA, poleDec, rotationEast := model.poleRotation(targetJD)
	poleJ2000 := raDecToVector(poleRA, poleDec)
	nodeJ2000 := Vector3{-math.Sin(poleRA * rad), math.Cos(poleRA * rad), 0}
	eastJ2000 := normalizeVector(pxp(poleJ2000, nodeJ2000))
	primeMeridianJ2000 := normalizeVector(Vector3{
		nodeJ2000[0]*Cos(rotationEast) + eastJ2000[0]*Sin(rotationEast),
		nodeJ2000[1]*Cos(rotationEast) + eastJ2000[1]*Sin(rotationEast),
		nodeJ2000[2]*Cos(rotationEast) + eastJ2000[2]*Sin(rotationEast),
	})

	j2000ToDate := precessionMatrix(2451545.0, targetJD)
	poleDate := normalizeVector(applyMatrix3(j2000ToDate, poleJ2000))
	primeMeridianDate := normalizeVector(applyMatrix3(j2000ToDate, primeMeridianJ2000))
	eastDate := normalizeVector(pxp(poleDate, primeMeridianDate))

	poleRAOfDate, poleDecOfDate := vectorToRaDec(poleDate)
	planetRA, planetDec := vectorToRaDec(observerEquatorial)

	return PlanetPhysicalInfo{
		SubEarthLongitude:      bodyLongitude(observerEquatorial, primeMeridianDate, eastDate, model.positiveEast),
		SubEarthLatitude:       bodyLatitude(observerEquatorial, poleDate, model.equatorialRadius, model.polarRadius),
		SubSolarLongitude:      bodyLongitude(solarEquatorial, primeMeridianDate, eastDate, model.positiveEast),
		SubSolarLatitude:       bodyLatitude(solarEquatorial, poleDate, model.equatorialRadius, model.polarRadius),
		NorthPolePositionAngle: northPolePositionAngle(planetRA, planetDec, poleRAOfDate, poleDecOfDate),
	}
}

func bodyLongitude(direction, primeMeridian, eastAxis Vector3, positiveEast bool) float64 {
	eastLongitude := Limit360(math.Atan2(vectorDot(direction, eastAxis), vectorDot(direction, primeMeridian)) * deg)
	if positiveEast {
		return eastLongitude
	}
	return Limit360(360 - eastLongitude)
}

func bodyLatitude(direction, pole Vector3, equatorialRadius, polarRadius float64) float64 {
	geocentricLatitude := ArcSin(vectorDot(direction, pole))
	if equatorialRadius == polarRadius {
		return geocentricLatitude
	}
	ratio := (polarRadius * polarRadius) / (equatorialRadius * equatorialRadius)
	return math.Atan2(math.Sin(geocentricLatitude*rad), math.Cos(geocentricLatitude*rad)*ratio) * deg
}

func northPolePositionAngle(planetRA, planetDec, poleRA, poleDec float64) float64 {
	y := math.Cos(poleDec*rad) * math.Sin((poleRA-planetRA)*rad)
	x := math.Sin(poleDec*rad)*math.Cos(planetDec*rad) -
		math.Cos(poleDec*rad)*math.Sin(planetDec*rad)*math.Cos((poleRA-planetRA)*rad)
	return Limit360(-math.Atan2(y, x) * deg)
}

func vectorDot(a, b Vector3) float64 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

func vectorMagnitude(vector Vector3) float64 {
	return math.Sqrt(vectorDot(vector, vector))
}

func normalizeVector(vector Vector3) Vector3 {
	normalized, magnitude := pn(vector)
	if magnitude == 0 {
		return Vector3{}
	}
	return normalized
}

func mercuryPoleRotation(jd float64) (ra, dec, rotationEast float64) {
	days := jd - 2451545.0
	julianCentury := days / 36525.0
	m1 := Limit360(174.7910857 + 4.092335*days)
	m2 := Limit360(349.5821714 + 8.184670*days)
	m3 := Limit360(164.3732571 + 12.277005*days)
	m4 := Limit360(339.1643429 + 16.369340*days)
	m5 := Limit360(153.9554286 + 20.461675*days)
	ra = 281.0103 - 0.0328*julianCentury
	dec = 61.4155 - 0.0049*julianCentury
	rotationEast = 329.5988 + 6.1385108*days +
		0.01067257*Sin(m1) -
		0.00112309*Sin(m2) -
		0.00011040*Sin(m3) -
		0.00002539*Sin(m4) -
		0.00000571*Sin(m5)
	return
}

func venusPoleRotation(jd float64) (ra, dec, rotationEast float64) {
	days := jd - 2451545.0
	return 272.76, 67.16, 160.20 - 1.4813688*days
}

// Mars/Jupiter orientation terms follow the official NAIF text PCK `pck00011.tpc`.
var (
	marsPoleRAAngles = [...]phaseAngleModel{
		{198.991226, 19139.4819985, 0},
		{226.292679, 38280.8511281, 0},
		{249.663391, 57420.7251593, 0},
		{266.183510, 76560.6367950, 0},
		{79.398797, 0.5042615, 0},
	}
	marsPoleRACoefficients = [...]float64{0.000068, 0.000238, 0.000052, 0.000009, 0.419057}
	marsPoleDECAngles      = [...]phaseAngleModel{
		{122.433576, 19139.9407476, 0},
		{43.058401, 38280.8753272, 0},
		{57.663379, 57420.7517205, 0},
		{79.476401, 76560.6495004, 0},
		{166.325722, 0.5042615, 0},
	}
	marsPoleDECCoefficients = [...]float64{0.000051, 0.000141, 0.000031, 0.000005, 1.591274}
	marsPrimeMeridianAngles = [...]phaseAngleModel{
		{129.071773, 19140.0328244, 0},
		{36.352167, 38281.0473591, 0},
		{56.668646, 57420.9295360, 0},
		{67.364003, 76560.2552215, 0},
		{104.792680, 95700.4387578, 0},
		{95.391654, 0.5042615, 0},
	}
	marsPrimeMeridianCoefficients = [...]float64{0.000145, 0.000157, 0.000040, 0.000001, 0.000001, 0.584542}
	jupiterPoleAngles             = [...]phaseAngleModel{
		{99.360714, 4850.4046, 0},
		{175.895369, 1191.9605, 0},
		{300.323162, 262.5475, 0},
		{114.012305, 6070.2476, 0},
		{49.511251, 64.3000, 0},
	}
	jupiterPoleRACoefficients  = [...]float64{0.000117, 0.000938, 0.001432, 0.000030, 0.002150}
	jupiterPoleDECCoefficients = [...]float64{0.000050, 0.000404, 0.000617, -0.000013, 0.000926}
)

func marsPoleRotation(jd float64) (ra, dec, rotationEast float64) {
	days := jd - 2451545.0
	julianCentury := days / 36525.0
	ra = 317.269202 - 0.10927547*julianCentury + sumSinOrientationTerms(marsPoleRAAngles[:], marsPoleRACoefficients[:], julianCentury)
	dec = 54.432516 - 0.05827105*julianCentury + sumCosOrientationTerms(marsPoleDECAngles[:], marsPoleDECCoefficients[:], julianCentury)
	rotationEast = 176.049863 + 350.891982443297*days +
		sumSinOrientationTerms(marsPrimeMeridianAngles[:], marsPrimeMeridianCoefficients[:], julianCentury)
	return
}

func jupiterPoleRotation(jd float64) (ra, dec, rotationEast float64) {
	days := jd - 2451545.0
	julianCentury := days / 36525.0
	ra = 268.056595 - 0.006499*julianCentury + sumSinOrientationTerms(jupiterPoleAngles[:], jupiterPoleRACoefficients[:], julianCentury)
	dec = 64.495303 + 0.002413*julianCentury + sumCosOrientationTerms(jupiterPoleAngles[:], jupiterPoleDECCoefficients[:], julianCentury)
	rotationEast = 284.95 + 870.5360000*days
	return
}

func saturnPoleRotation(jd float64) (ra, dec, rotationEast float64) {
	days := jd - 2451545.0
	julianCentury := days / 36525.0
	ra = 40.589 - 0.036*julianCentury
	dec = 83.537 - 0.004*julianCentury
	rotationEast = 38.90 + 810.7939024*days
	return
}

func uranusPoleRotation(jd float64) (ra, dec, rotationEast float64) {
	days := jd - 2451545.0
	return 257.311, -15.175, 203.81 - 501.1600928*days
}

func neptunePoleRotation(jd float64) (ra, dec, rotationEast float64) {
	days := jd - 2451545.0
	julianCentury := days / 36525.0
	n := Limit360(357.85 + 52.316*julianCentury)
	ra = 299.36 + 0.70*Sin(n)
	dec = 43.46 - 0.51*Cos(n)
	rotationEast = 249.978 + 541.1397757*days - 0.48*Sin(n)
	return
}

func (model phaseAngleModel) at(julianCentury float64) float64 {
	return Limit360(model.base + model.rate*julianCentury + model.accelRate*julianCentury*julianCentury)
}

func sumSinOrientationTerms(angles []phaseAngleModel, coefficients []float64, julianCentury float64) float64 {
	sum := 0.0
	for i, angle := range angles {
		sum += coefficients[i] * Sin(angle.at(julianCentury))
	}
	return sum
}

func sumCosOrientationTerms(angles []phaseAngleModel, coefficients []float64, julianCentury float64) float64 {
	sum := 0.0
	for i, angle := range angles {
		sum += coefficients[i] * Cos(angle.at(julianCentury))
	}
	return sum
}
