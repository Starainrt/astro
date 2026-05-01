package basic

import "math"

const (
	angularDiameterAstronomicalUnitKM = 149597870.7
	angularDiameterArcsecPerRadian    = 180.0 * 3600.0 / math.Pi

	sunEquatorialRadiusKM     = 695700.0
	moonEquatorialRadiusKM    = 1737.4
	mercuryEquatorialRadiusKM = 2440.53
	venusEquatorialRadiusKM   = 6051.8
	marsEquatorialRadiusKM    = 3396.19
	jupiterEquatorialRadiusKM = 71492.0
	saturnEquatorialRadiusKM  = 60268.0
	uranusEquatorialRadiusKM  = 25559.0
	neptuneEquatorialRadiusKM = 24764.0
)

func angularSemidiameterArcsec(radiusKM, distanceKM float64) float64 {
	return math.Asin(radiusKM/distanceKM) * angularDiameterArcsecPerRadian
}

func angularSemidiameterFromAU(radiusKM, distanceAU float64) float64 {
	return angularSemidiameterArcsec(radiusKM, distanceAU*angularDiameterAstronomicalUnitKM)
}

// SunSemidiameter 太阳视半径，单位角秒 / apparent solar semidiameter in arcseconds.
func SunSemidiameter(jd float64) float64 {
	return SunSemidiameterN(jd, -1)
}

// SunSemidiameterN 太阳视半径（截断版），单位角秒 / truncated apparent solar semidiameter in arcseconds.
func SunSemidiameterN(jd float64, n int) float64 {
	return angularSemidiameterFromAU(sunEquatorialRadiusKM, EarthAwayN(jd, n))
}

// SunDiameter 太阳视直径，单位角秒 / apparent solar diameter in arcseconds.
func SunDiameter(jd float64) float64 {
	return SunDiameterN(jd, -1)
}

// SunDiameterN 太阳视直径（截断版），单位角秒 / truncated apparent solar diameter in arcseconds.
func SunDiameterN(jd float64, n int) float64 {
	return 2 * SunSemidiameterN(jd, n)
}

// MoonSemidiameter 月亮视半径，单位角秒 / apparent lunar semidiameter in arcseconds.
func MoonSemidiameter(jd float64) float64 {
	return MoonSemidiameterN(jd, -1)
}

// MoonSemidiameterN 月亮视半径（截断版），单位角秒 / truncated apparent lunar semidiameter in arcseconds.
func MoonSemidiameterN(jd float64, n int) float64 {
	return angularSemidiameterArcsec(moonEquatorialRadiusKM, HMoonAwayN(jd, n))
}

// MoonDiameter 月亮视直径，单位角秒 / apparent lunar diameter in arcseconds.
func MoonDiameter(jd float64) float64 {
	return MoonDiameterN(jd, -1)
}

// MoonDiameterN 月亮视直径（截断版），单位角秒 / truncated apparent lunar diameter in arcseconds.
func MoonDiameterN(jd float64, n int) float64 {
	return 2 * MoonSemidiameterN(jd, n)
}

// MercurySemidiameter 水星视半径，单位角秒 / apparent Mercury semidiameter in arcseconds.
func MercurySemidiameter(jd float64) float64 {
	return MercurySemidiameterN(jd, -1)
}

// MercurySemidiameterN 水星视半径（截断版），单位角秒 / truncated apparent Mercury semidiameter in arcseconds.
func MercurySemidiameterN(jd float64, n int) float64 {
	return angularSemidiameterFromAU(mercuryEquatorialRadiusKM, EarthMercuryAwayN(jd, n))
}

// MercuryDiameter 水星视直径，单位角秒 / apparent Mercury diameter in arcseconds.
func MercuryDiameter(jd float64) float64 {
	return MercuryDiameterN(jd, -1)
}

// MercuryDiameterN 水星视直径（截断版），单位角秒 / truncated apparent Mercury diameter in arcseconds.
func MercuryDiameterN(jd float64, n int) float64 {
	return 2 * MercurySemidiameterN(jd, n)
}

// VenusSemidiameter 金星视半径，单位角秒 / apparent Venus semidiameter in arcseconds.
func VenusSemidiameter(jd float64) float64 {
	return VenusSemidiameterN(jd, -1)
}

// VenusSemidiameterN 金星视半径（截断版），单位角秒 / truncated apparent Venus semidiameter in arcseconds.
func VenusSemidiameterN(jd float64, n int) float64 {
	return angularSemidiameterFromAU(venusEquatorialRadiusKM, EarthVenusAwayN(jd, n))
}

// VenusDiameter 金星视直径，单位角秒 / apparent Venus diameter in arcseconds.
func VenusDiameter(jd float64) float64 {
	return VenusDiameterN(jd, -1)
}

// VenusDiameterN 金星视直径（截断版），单位角秒 / truncated apparent Venus diameter in arcseconds.
func VenusDiameterN(jd float64, n int) float64 {
	return 2 * VenusSemidiameterN(jd, n)
}

// MarsSemidiameter 火星视半径，单位角秒 / apparent Mars semidiameter in arcseconds.
func MarsSemidiameter(jd float64) float64 {
	return MarsSemidiameterN(jd, -1)
}

// MarsSemidiameterN 火星视半径（截断版），单位角秒 / truncated apparent Mars semidiameter in arcseconds.
func MarsSemidiameterN(jd float64, n int) float64 {
	return angularSemidiameterFromAU(marsEquatorialRadiusKM, EarthMarsAwayN(jd, n))
}

// MarsDiameter 火星视直径，单位角秒 / apparent Mars diameter in arcseconds.
func MarsDiameter(jd float64) float64 {
	return MarsDiameterN(jd, -1)
}

// MarsDiameterN 火星视直径（截断版），单位角秒 / truncated apparent Mars diameter in arcseconds.
func MarsDiameterN(jd float64, n int) float64 {
	return 2 * MarsSemidiameterN(jd, n)
}

// JupiterSemidiameter 木星视半径，单位角秒 / apparent Jupiter semidiameter in arcseconds.
func JupiterSemidiameter(jd float64) float64 {
	return JupiterSemidiameterN(jd, -1)
}

// JupiterSemidiameterN 木星视半径（截断版），单位角秒 / truncated apparent Jupiter semidiameter in arcseconds.
func JupiterSemidiameterN(jd float64, n int) float64 {
	return angularSemidiameterFromAU(jupiterEquatorialRadiusKM, EarthJupiterAwayN(jd, n))
}

// JupiterDiameter 木星视直径，单位角秒 / apparent Jupiter diameter in arcseconds.
func JupiterDiameter(jd float64) float64 {
	return JupiterDiameterN(jd, -1)
}

// JupiterDiameterN 木星视直径（截断版），单位角秒 / truncated apparent Jupiter diameter in arcseconds.
func JupiterDiameterN(jd float64, n int) float64 {
	return 2 * JupiterSemidiameterN(jd, n)
}

// SaturnSemidiameter 土星视半径，单位角秒 / apparent Saturn semidiameter in arcseconds.
func SaturnSemidiameter(jd float64) float64 {
	return SaturnSemidiameterN(jd, -1)
}

// SaturnSemidiameterN 土星视半径（截断版），单位角秒 / truncated apparent Saturn semidiameter in arcseconds.
func SaturnSemidiameterN(jd float64, n int) float64 {
	return angularSemidiameterFromAU(saturnEquatorialRadiusKM, EarthSaturnAwayN(jd, n))
}

// SaturnDiameter 土星视直径，单位角秒 / apparent Saturn diameter in arcseconds.
func SaturnDiameter(jd float64) float64 {
	return SaturnDiameterN(jd, -1)
}

// SaturnDiameterN 土星视直径（截断版），单位角秒 / truncated apparent Saturn diameter in arcseconds.
func SaturnDiameterN(jd float64, n int) float64 {
	return 2 * SaturnSemidiameterN(jd, n)
}

// UranusSemidiameter 天王星视半径，单位角秒 / apparent Uranus semidiameter in arcseconds.
func UranusSemidiameter(jd float64) float64 {
	return UranusSemidiameterN(jd, -1)
}

// UranusSemidiameterN 天王星视半径（截断版），单位角秒 / truncated apparent Uranus semidiameter in arcseconds.
func UranusSemidiameterN(jd float64, n int) float64 {
	return angularSemidiameterFromAU(uranusEquatorialRadiusKM, EarthUranusAwayN(jd, n))
}

// UranusDiameter 天王星视直径，单位角秒 / apparent Uranus diameter in arcseconds.
func UranusDiameter(jd float64) float64 {
	return UranusDiameterN(jd, -1)
}

// UranusDiameterN 天王星视直径（截断版），单位角秒 / truncated apparent Uranus diameter in arcseconds.
func UranusDiameterN(jd float64, n int) float64 {
	return 2 * UranusSemidiameterN(jd, n)
}

// NeptuneSemidiameter 海王星视半径，单位角秒 / apparent Neptune semidiameter in arcseconds.
func NeptuneSemidiameter(jd float64) float64 {
	return NeptuneSemidiameterN(jd, -1)
}

// NeptuneSemidiameterN 海王星视半径（截断版），单位角秒 / truncated apparent Neptune semidiameter in arcseconds.
func NeptuneSemidiameterN(jd float64, n int) float64 {
	return angularSemidiameterFromAU(neptuneEquatorialRadiusKM, EarthNeptuneAwayN(jd, n))
}

// NeptuneDiameter 海王星视直径，单位角秒 / apparent Neptune diameter in arcseconds.
func NeptuneDiameter(jd float64) float64 {
	return NeptuneDiameterN(jd, -1)
}

// NeptuneDiameterN 海王星视直径（截断版），单位角秒 / truncated apparent Neptune diameter in arcseconds.
func NeptuneDiameterN(jd float64, n int) float64 {
	return 2 * NeptuneSemidiameterN(jd, n)
}
