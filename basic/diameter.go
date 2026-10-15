package basic

import (
	"math"
)

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

// SunSemidiameterNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SunSemidiameterNWithInstance(jde float64, n int, inst *Instance) float64 {
	return angularSemidiameterFromAU(sunEquatorialRadiusKM, EarthAwayNWithInstance(jde, n, inst))
}

// SunDiameter 太阳视直径，单位角秒 / apparent solar diameter in arcseconds.
func SunDiameter(jde float64) float64 {
	return SunDiameterN(jde, -1)
}

// SunDiameterNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SunDiameterNWithInstance(jde float64, n int, inst *Instance) float64 {
	return 2 * SunSemidiameterNWithInstance(jde, n, inst)
}

// MoonSemidiameterWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MoonSemidiameterWithInstance(jde float64, inst *Instance) float64 {
	return MoonSemidiameterNWithInstance(jde, -1, inst)
}

// MoonSemidiameterNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MoonSemidiameterNWithInstance(jde float64, n int, inst *Instance) float64 {
	return angularSemidiameterArcsec(moonEquatorialRadiusKM, HMoonAwayNWithInstance(jde, n, inst))
}

// MoonDiameter 月亮视直径，单位角秒 / apparent lunar diameter in arcseconds.
func MoonDiameter(jde float64) float64 {
	return MoonDiameterN(jde, -1)
}

// MoonDiameterNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MoonDiameterNWithInstance(jde float64, n int, inst *Instance) float64 {
	return 2 * MoonSemidiameterNWithInstance(jde, n, inst)
}

// MercurySemidiameter 水星视半径，单位角秒 / apparent Mercury semidiameter in arcseconds.
func MercurySemidiameter(jde float64) float64 {
	return MercurySemidiameterN(jde, -1)
}

// MercurySemidiameterNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercurySemidiameterNWithInstance(jde float64, n int, inst *Instance) float64 {
	return angularSemidiameterFromAU(mercuryEquatorialRadiusKM, EarthMercuryAwayNWithInstance(jde, n, inst))
}

// MercuryDiameter 水星视直径，单位角秒 / apparent Mercury diameter in arcseconds.
func MercuryDiameter(jde float64) float64 {
	return MercuryDiameterN(jde, -1)
}

// MercuryDiameterNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercuryDiameterNWithInstance(jde float64, n int, inst *Instance) float64 {
	return 2 * MercurySemidiameterNWithInstance(jde, n, inst)
}

// VenusSemidiameter 金星视半径，单位角秒 / apparent Venus semidiameter in arcseconds.
func VenusSemidiameter(jde float64) float64 {
	return VenusSemidiameterN(jde, -1)
}

// VenusSemidiameterNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func VenusSemidiameterNWithInstance(jde float64, n int, inst *Instance) float64 {
	return angularSemidiameterFromAU(venusEquatorialRadiusKM, EarthVenusAwayNWithInstance(jde, n, inst))
}

// VenusDiameter 金星视直径，单位角秒 / apparent Venus diameter in arcseconds.
func VenusDiameter(jde float64) float64 {
	return VenusDiameterN(jde, -1)
}

// VenusDiameterNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func VenusDiameterNWithInstance(jde float64, n int, inst *Instance) float64 {
	return 2 * VenusSemidiameterNWithInstance(jde, n, inst)
}

// MarsSemidiameter 火星视半径，单位角秒 / apparent Mars semidiameter in arcseconds.
func MarsSemidiameter(jde float64) float64 {
	return MarsSemidiameterN(jde, -1)
}

// MarsSemidiameterNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MarsSemidiameterNWithInstance(jde float64, n int, inst *Instance) float64 {
	return angularSemidiameterFromAU(marsEquatorialRadiusKM, EarthMarsAwayNWithInstance(jde, n, inst))
}

// MarsDiameter 火星视直径，单位角秒 / apparent Mars diameter in arcseconds.
func MarsDiameter(jde float64) float64 {
	return MarsDiameterN(jde, -1)
}

// MarsDiameterNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MarsDiameterNWithInstance(jde float64, n int, inst *Instance) float64 {
	return 2 * MarsSemidiameterNWithInstance(jde, n, inst)
}

// JupiterSemidiameter 木星视半径，单位角秒 / apparent Jupiter semidiameter in arcseconds.
func JupiterSemidiameter(jde float64) float64 {
	return JupiterSemidiameterN(jde, -1)
}

// JupiterSemidiameterNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func JupiterSemidiameterNWithInstance(jde float64, n int, inst *Instance) float64 {
	return angularSemidiameterFromAU(jupiterEquatorialRadiusKM, EarthJupiterAwayNWithInstance(jde, n, inst))
}

// JupiterDiameter 木星视直径，单位角秒 / apparent Jupiter diameter in arcseconds.
func JupiterDiameter(jde float64) float64 {
	return JupiterDiameterN(jde, -1)
}

// JupiterDiameterNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func JupiterDiameterNWithInstance(jde float64, n int, inst *Instance) float64 {
	return 2 * JupiterSemidiameterNWithInstance(jde, n, inst)
}

// SaturnSemidiameter 土星视半径，单位角秒 / apparent Saturn semidiameter in arcseconds.
func SaturnSemidiameter(jde float64) float64 {
	return SaturnSemidiameterN(jde, -1)
}

// SaturnSemidiameterNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnSemidiameterNWithInstance(jde float64, n int, inst *Instance) float64 {
	return angularSemidiameterFromAU(saturnEquatorialRadiusKM, EarthSaturnAwayNWithInstance(jde, n, inst))
}

// SaturnDiameter 土星视直径，单位角秒 / apparent Saturn diameter in arcseconds.
func SaturnDiameter(jde float64) float64 {
	return SaturnDiameterN(jde, -1)
}

// SaturnDiameterNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnDiameterNWithInstance(jde float64, n int, inst *Instance) float64 {
	return 2 * SaturnSemidiameterNWithInstance(jde, n, inst)
}

// UranusSemidiameter 天王星视半径，单位角秒 / apparent Uranus semidiameter in arcseconds.
func UranusSemidiameter(jde float64) float64 {
	return UranusSemidiameterN(jde, -1)
}

// UranusSemidiameterNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusSemidiameterNWithInstance(jde float64, n int, inst *Instance) float64 {
	return angularSemidiameterFromAU(uranusEquatorialRadiusKM, EarthUranusAwayNWithInstance(jde, n, inst))
}

// UranusDiameter 天王星视直径，单位角秒 / apparent Uranus diameter in arcseconds.
func UranusDiameter(jde float64) float64 {
	return UranusDiameterN(jde, -1)
}

// UranusDiameterNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusDiameterNWithInstance(jde float64, n int, inst *Instance) float64 {
	return 2 * UranusSemidiameterNWithInstance(jde, n, inst)
}

// NeptuneSemidiameter 海王星视半径，单位角秒 / apparent Neptune semidiameter in arcseconds.
func NeptuneSemidiameter(jde float64) float64 {
	return NeptuneSemidiameterN(jde, -1)
}

// NeptuneSemidiameterNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NeptuneSemidiameterNWithInstance(jde float64, n int, inst *Instance) float64 {
	return angularSemidiameterFromAU(neptuneEquatorialRadiusKM, EarthNeptuneAwayNWithInstance(jde, n, inst))
}

// NeptuneDiameter 海王星视直径，单位角秒 / apparent Neptune diameter in arcseconds.
func NeptuneDiameter(jde float64) float64 {
	return NeptuneDiameterN(jde, -1)
}

// NeptuneDiameterNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NeptuneDiameterNWithInstance(jde float64, n int, inst *Instance) float64 {
	return 2 * NeptuneSemidiameterNWithInstance(jde, n, inst)
}
