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
func SunSemidiameter(jde float64) float64 {
	return SunSemidiameterN(jde, -1)
}

// SunSemidiameterN 太阳视半径（截断版），单位角秒 / truncated apparent solar semidiameter in arcseconds.
func SunSemidiameterN(jde float64, n int) float64 {
	return angularSemidiameterFromAU(sunEquatorialRadiusKM, EarthAwayN(jde, n))
}

// SunDiameter 太阳视直径，单位角秒 / apparent solar diameter in arcseconds.
func SunDiameter(jde float64) float64 {
	return SunDiameterN(jde, -1)
}

// SunDiameterN 太阳视直径（截断版），单位角秒 / truncated apparent solar diameter in arcseconds.
func SunDiameterN(jde float64, n int) float64 {
	return 2 * SunSemidiameterN(jde, n)
}

// MoonSemidiameter 月亮视半径，单位角秒 / apparent lunar semidiameter in arcseconds.
func MoonSemidiameter(jde float64) float64 {
	return MoonSemidiameterN(jde, -1)
}

// MoonSemidiameterN 月亮视半径（截断版），单位角秒 / truncated apparent lunar semidiameter in arcseconds.
func MoonSemidiameterN(jde float64, n int) float64 {
	return angularSemidiameterArcsec(moonEquatorialRadiusKM, HMoonAwayN(jde, n))
}

// MoonDiameter 月亮视直径，单位角秒 / apparent lunar diameter in arcseconds.
func MoonDiameter(jde float64) float64 {
	return MoonDiameterN(jde, -1)
}

// MoonDiameterN 月亮视直径（截断版），单位角秒 / truncated apparent lunar diameter in arcseconds.
func MoonDiameterN(jde float64, n int) float64 {
	return 2 * MoonSemidiameterN(jde, n)
}

// MercurySemidiameter 水星视半径，单位角秒 / apparent Mercury semidiameter in arcseconds.
func MercurySemidiameter(jde float64) float64 {
	return MercurySemidiameterN(jde, -1)
}

// MercurySemidiameterN 水星视半径（截断版），单位角秒 / truncated apparent Mercury semidiameter in arcseconds.
func MercurySemidiameterN(jde float64, n int) float64 {
	return angularSemidiameterFromAU(mercuryEquatorialRadiusKM, EarthMercuryAwayN(jde, n))
}

// MercuryDiameter 水星视直径，单位角秒 / apparent Mercury diameter in arcseconds.
func MercuryDiameter(jde float64) float64 {
	return MercuryDiameterN(jde, -1)
}

// MercuryDiameterN 水星视直径（截断版），单位角秒 / truncated apparent Mercury diameter in arcseconds.
func MercuryDiameterN(jde float64, n int) float64 {
	return 2 * MercurySemidiameterN(jde, n)
}

// VenusSemidiameter 金星视半径，单位角秒 / apparent Venus semidiameter in arcseconds.
func VenusSemidiameter(jde float64) float64 {
	return VenusSemidiameterN(jde, -1)
}

// VenusSemidiameterN 金星视半径（截断版），单位角秒 / truncated apparent Venus semidiameter in arcseconds.
func VenusSemidiameterN(jde float64, n int) float64 {
	return angularSemidiameterFromAU(venusEquatorialRadiusKM, EarthVenusAwayN(jde, n))
}

// VenusDiameter 金星视直径，单位角秒 / apparent Venus diameter in arcseconds.
func VenusDiameter(jde float64) float64 {
	return VenusDiameterN(jde, -1)
}

// VenusDiameterN 金星视直径（截断版），单位角秒 / truncated apparent Venus diameter in arcseconds.
func VenusDiameterN(jde float64, n int) float64 {
	return 2 * VenusSemidiameterN(jde, n)
}

// MarsSemidiameter 火星视半径，单位角秒 / apparent Mars semidiameter in arcseconds.
func MarsSemidiameter(jde float64) float64 {
	return MarsSemidiameterN(jde, -1)
}

// MarsSemidiameterN 火星视半径（截断版），单位角秒 / truncated apparent Mars semidiameter in arcseconds.
func MarsSemidiameterN(jde float64, n int) float64 {
	return angularSemidiameterFromAU(marsEquatorialRadiusKM, EarthMarsAwayN(jde, n))
}

// MarsDiameter 火星视直径，单位角秒 / apparent Mars diameter in arcseconds.
func MarsDiameter(jde float64) float64 {
	return MarsDiameterN(jde, -1)
}

// MarsDiameterN 火星视直径（截断版），单位角秒 / truncated apparent Mars diameter in arcseconds.
func MarsDiameterN(jde float64, n int) float64 {
	return 2 * MarsSemidiameterN(jde, n)
}

// JupiterSemidiameter 木星视半径，单位角秒 / apparent Jupiter semidiameter in arcseconds.
func JupiterSemidiameter(jde float64) float64 {
	return JupiterSemidiameterN(jde, -1)
}

// JupiterSemidiameterN 木星视半径（截断版），单位角秒 / truncated apparent Jupiter semidiameter in arcseconds.
func JupiterSemidiameterN(jde float64, n int) float64 {
	return angularSemidiameterFromAU(jupiterEquatorialRadiusKM, EarthJupiterAwayN(jde, n))
}

// JupiterDiameter 木星视直径，单位角秒 / apparent Jupiter diameter in arcseconds.
func JupiterDiameter(jde float64) float64 {
	return JupiterDiameterN(jde, -1)
}

// JupiterDiameterN 木星视直径（截断版），单位角秒 / truncated apparent Jupiter diameter in arcseconds.
func JupiterDiameterN(jde float64, n int) float64 {
	return 2 * JupiterSemidiameterN(jde, n)
}

// SaturnSemidiameter 土星视半径，单位角秒 / apparent Saturn semidiameter in arcseconds.
func SaturnSemidiameter(jde float64) float64 {
	return SaturnSemidiameterN(jde, -1)
}

// SaturnSemidiameterN 土星视半径（截断版），单位角秒 / truncated apparent Saturn semidiameter in arcseconds.
func SaturnSemidiameterN(jde float64, n int) float64 {
	return angularSemidiameterFromAU(saturnEquatorialRadiusKM, EarthSaturnAwayN(jde, n))
}

// SaturnDiameter 土星视直径，单位角秒 / apparent Saturn diameter in arcseconds.
func SaturnDiameter(jde float64) float64 {
	return SaturnDiameterN(jde, -1)
}

// SaturnDiameterN 土星视直径（截断版），单位角秒 / truncated apparent Saturn diameter in arcseconds.
func SaturnDiameterN(jde float64, n int) float64 {
	return 2 * SaturnSemidiameterN(jde, n)
}

// UranusSemidiameter 天王星视半径，单位角秒 / apparent Uranus semidiameter in arcseconds.
func UranusSemidiameter(jde float64) float64 {
	return UranusSemidiameterN(jde, -1)
}

// UranusSemidiameterN 天王星视半径（截断版），单位角秒 / truncated apparent Uranus semidiameter in arcseconds.
func UranusSemidiameterN(jde float64, n int) float64 {
	return angularSemidiameterFromAU(uranusEquatorialRadiusKM, EarthUranusAwayN(jde, n))
}

// UranusDiameter 天王星视直径，单位角秒 / apparent Uranus diameter in arcseconds.
func UranusDiameter(jde float64) float64 {
	return UranusDiameterN(jde, -1)
}

// UranusDiameterN 天王星视直径（截断版），单位角秒 / truncated apparent Uranus diameter in arcseconds.
func UranusDiameterN(jde float64, n int) float64 {
	return 2 * UranusSemidiameterN(jde, n)
}

// NeptuneSemidiameter 海王星视半径，单位角秒 / apparent Neptune semidiameter in arcseconds.
func NeptuneSemidiameter(jde float64) float64 {
	return NeptuneSemidiameterN(jde, -1)
}

// NeptuneSemidiameterN 海王星视半径（截断版），单位角秒 / truncated apparent Neptune semidiameter in arcseconds.
func NeptuneSemidiameterN(jde float64, n int) float64 {
	return angularSemidiameterFromAU(neptuneEquatorialRadiusKM, EarthNeptuneAwayN(jde, n))
}

// NeptuneDiameter 海王星视直径，单位角秒 / apparent Neptune diameter in arcseconds.
func NeptuneDiameter(jde float64) float64 {
	return NeptuneDiameterN(jde, -1)
}

// NeptuneDiameterN 海王星视直径（截断版），单位角秒 / truncated apparent Neptune diameter in arcseconds.
func NeptuneDiameterN(jde float64, n int) float64 {
	return 2 * NeptuneSemidiameterN(jde, n)
}
