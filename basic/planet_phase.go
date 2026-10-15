package basic

import (
	. "github.com/starainrt/astro/tools"
)

type planetRaDecNFunc func(jd float64, n int) (float64, float64)

// MercuryPhaseAngle 水星相位角 / phase angle of Mercury.
func MercuryPhaseAngle(jde float64) float64 {
	return MercuryPhaseAngleN(jde, -1)
}

// MercuryPhaseAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercuryPhaseAngleNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetPhaseAngleN(1, jd, n, inst)
}

// MercuryIlluminatedFraction 水星被照亮比例 / illuminated fraction of Mercury.
func MercuryIlluminatedFraction(jd float64) float64 {
	return MercuryIlluminatedFractionN(jd, -1)
}

// MercuryIlluminatedFractionNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercuryIlluminatedFractionNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetIlluminatedFractionN(1, jd, n, inst)
}

// MercuryBrightLimbPositionAngle 水星亮面中心位置角 / position angle of Mercury bright limb.
func MercuryBrightLimbPositionAngle(jde float64) float64 {
	return MercuryBrightLimbPositionAngleN(jde, -1)
}

// MercuryBrightLimbPositionAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercuryBrightLimbPositionAngleNWithInstance(jde float64, n int, inst *Instance) float64 {
	return planetBrightLimbPositionAngleN(jde, n, func(jde float64, n int) (float64, float64) { return MercuryApparentRaDecNWithInstance(jde, n, inst) }, inst)
}

// VenusPhaseAngle 金星相位角 / phase angle of Venus.
func VenusPhaseAngle(jde float64) float64 {
	return VenusPhaseAngleN(jde, -1)
}

// VenusPhaseAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func VenusPhaseAngleNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetPhaseAngleN(2, jd, n, inst)
}

// VenusIlluminatedFraction 金星被照亮比例 / illuminated fraction of Venus.
func VenusIlluminatedFraction(jd float64) float64 {
	return VenusIlluminatedFractionN(jd, -1)
}

// VenusIlluminatedFractionNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func VenusIlluminatedFractionNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetIlluminatedFractionN(2, jd, n, inst)
}

// VenusBrightLimbPositionAngle 金星亮面中心位置角 / position angle of Venus bright limb.
func VenusBrightLimbPositionAngle(jde float64) float64 {
	return VenusBrightLimbPositionAngleN(jde, -1)
}

// VenusBrightLimbPositionAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func VenusBrightLimbPositionAngleNWithInstance(jde float64, n int, inst *Instance) float64 {
	return planetBrightLimbPositionAngleN(jde, n, func(jd float64, n int) (float64, float64) { return VenusApparentRaDecNWithInstance(jd, n, inst) }, inst)
}

// MarsPhaseAngle 火星相位角 / phase angle of Mars.
func MarsPhaseAngle(jde float64) float64 {
	return MarsPhaseAngleN(jde, -1)
}

// MarsPhaseAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MarsPhaseAngleNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetPhaseAngleN(3, jd, n, inst)
}

// MarsIlluminatedFraction 火星被照亮比例 / illuminated fraction of Mars.
func MarsIlluminatedFraction(jd float64) float64 {
	return MarsIlluminatedFractionN(jd, -1)
}

// MarsIlluminatedFractionNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MarsIlluminatedFractionNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetIlluminatedFractionN(3, jd, n, inst)
}

// MarsBrightLimbPositionAngle 火星亮面中心位置角 / position angle of Mars bright limb.
func MarsBrightLimbPositionAngle(jde float64) float64 {
	return MarsBrightLimbPositionAngleN(jde, -1)
}

// MarsBrightLimbPositionAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MarsBrightLimbPositionAngleNWithInstance(jde float64, n int, inst *Instance) float64 {
	return planetBrightLimbPositionAngleN(jde, n, func(jd float64, n int) (float64, float64) { return MarsApparentRaDecNWithInstance(jd, n, inst) }, inst)
}

// JupiterPhaseAngle 木星相位角 / phase angle of Jupiter.
func JupiterPhaseAngle(jde float64) float64 {
	return JupiterPhaseAngleN(jde, -1)
}

// JupiterPhaseAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func JupiterPhaseAngleNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetPhaseAngleN(4, jd, n, inst)
}

// JupiterIlluminatedFraction 木星被照亮比例 / illuminated fraction of Jupiter.
func JupiterIlluminatedFraction(jd float64) float64 {
	return JupiterIlluminatedFractionN(jd, -1)
}

// JupiterIlluminatedFractionNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func JupiterIlluminatedFractionNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetIlluminatedFractionN(4, jd, n, inst)
}

// JupiterBrightLimbPositionAngle 木星亮面中心位置角 / position angle of Jupiter bright limb.
func JupiterBrightLimbPositionAngle(jde float64) float64 {
	return JupiterBrightLimbPositionAngleN(jde, -1)
}

// JupiterBrightLimbPositionAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func JupiterBrightLimbPositionAngleNWithInstance(jde float64, n int, inst *Instance) float64 {
	return planetBrightLimbPositionAngleN(jde, n, func(jd float64, n int) (float64, float64) { return JupiterApparentRaDecNWithInstance(jd, n, inst) }, inst)
}

// SaturnPhaseAngle 土星相位角 / phase angle of Saturn.
func SaturnPhaseAngle(jde float64) float64 {
	return SaturnPhaseAngleN(jde, -1)
}

// SaturnPhaseAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnPhaseAngleNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetPhaseAngleN(5, jd, n, inst)
}

// SaturnIlluminatedFraction 土星被照亮比例 / illuminated fraction of Saturn.
func SaturnIlluminatedFraction(jd float64) float64 {
	return SaturnIlluminatedFractionN(jd, -1)
}

// SaturnIlluminatedFractionNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnIlluminatedFractionNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetIlluminatedFractionN(5, jd, n, inst)
}

// SaturnBrightLimbPositionAngle 土星亮面中心位置角 / position angle of Saturn bright limb.
func SaturnBrightLimbPositionAngle(jde float64) float64 {
	return SaturnBrightLimbPositionAngleN(jde, -1)
}

// SaturnBrightLimbPositionAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnBrightLimbPositionAngleNWithInstance(jde float64, n int, inst *Instance) float64 {
	return planetBrightLimbPositionAngleN(jde, n, func(jd float64, n int) (float64, float64) { return SaturnApparentRaDecNWithInstance(jd, n, inst) }, inst)
}

// UranusPhaseAngle 天王星相位角 / phase angle of Uranus.
func UranusPhaseAngle(jde float64) float64 {
	return UranusPhaseAngleN(jde, -1)
}

// UranusPhaseAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusPhaseAngleNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetPhaseAngleN(6, jd, n, inst)
}

// UranusIlluminatedFraction 天王星被照亮比例 / illuminated fraction of Uranus.
func UranusIlluminatedFraction(jd float64) float64 {
	return UranusIlluminatedFractionN(jd, -1)
}

// UranusIlluminatedFractionNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusIlluminatedFractionNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetIlluminatedFractionN(6, jd, n, inst)
}

// UranusBrightLimbPositionAngle 天王星亮面中心位置角 / position angle of Uranus bright limb.
func UranusBrightLimbPositionAngle(jde float64) float64 {
	return UranusBrightLimbPositionAngleN(jde, -1)
}

// UranusBrightLimbPositionAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusBrightLimbPositionAngleNWithInstance(jde float64, n int, inst *Instance) float64 {
	return planetBrightLimbPositionAngleN(jde, n, func(jd float64, n int) (float64, float64) { return UranusApparentRaDecNWithInstance(jd, n, inst) }, inst)
}

// NeptunePhaseAngle 海王星相位角 / phase angle of Neptune.
func NeptunePhaseAngle(jde float64) float64 {
	return NeptunePhaseAngleN(jde, -1)
}

// NeptunePhaseAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NeptunePhaseAngleNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetPhaseAngleN(7, jd, n, inst)
}

// NeptuneIlluminatedFraction 海王星被照亮比例 / illuminated fraction of Neptune.
func NeptuneIlluminatedFraction(jd float64) float64 {
	return NeptuneIlluminatedFractionN(jd, -1)
}

// NeptuneIlluminatedFractionNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NeptuneIlluminatedFractionNWithInstance(jd float64, n int, inst *Instance) float64 {
	return planetIlluminatedFractionN(7, jd, n, inst)
}

// NeptuneBrightLimbPositionAngle 海王星亮面中心位置角 / position angle of Neptune bright limb.
func NeptuneBrightLimbPositionAngle(jde float64) float64 {
	return NeptuneBrightLimbPositionAngleN(jde, -1)
}

// NeptuneBrightLimbPositionAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NeptuneBrightLimbPositionAngleNWithInstance(jde float64, n int, inst *Instance) float64 {
	return planetBrightLimbPositionAngleN(jde, n, func(jd float64, n int) (float64, float64) { return NeptuneApparentRaDecNWithInstance(jd, n, inst) }, inst)
}

func planetPhaseAngleN(planetIndex int, jd float64, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	return ArcCos(planetPhaseCosineN(planetIndex, jd, n, inst))
}

func planetIlluminatedFractionN(planetIndex int, jd float64, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	return (1 + planetPhaseCosineN(planetIndex, jd, n, inst)) / 2
}

func planetPhaseCosineN(planetIndex int, jde float64, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	planetSunDistance := PlanetCoordinateNWithInstance(planetIndex, 2, jde, n, inst)
	planetEarthDistance := planetEarthAwayNWithInstance(planetIndex, jde, n, inst)
	earthSunDistance := EarthAwayNWithInstance(jde, n, inst)
	cosine := (planetSunDistance*planetSunDistance + planetEarthDistance*planetEarthDistance - earthSunDistance*earthSunDistance) / (2 * planetSunDistance * planetEarthDistance)
	return clampUnit(cosine)
}

func planetBrightLimbPositionAngleN(jde float64, n int, apparentRaDec planetRaDecNFunc, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	sunRa, sunDec := HSunApparentRaDecNWithInstance(jde, n, inst)
	planetRa, planetDec := apparentRaDec(jde, n)
	y := Cos(sunDec) * Sin(sunRa-planetRa)
	x := Sin(sunDec)*Cos(planetDec) - Cos(sunDec)*Sin(planetDec)*Cos(sunRa-planetRa)
	return ArcTan2(y, x)
}

func clampUnit(value float64) float64 {
	if value > 1 {
		return 1
	}
	if value < -1 {
		return -1
	}
	return value
}
