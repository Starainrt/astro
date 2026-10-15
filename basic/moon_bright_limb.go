package basic

import (
	. "github.com/starainrt/astro/tools"
)

// MoonBrightLimbPositionAngle 月亮明亮边缘位置角 / position angle of the Moon's bright limb.
func MoonBrightLimbPositionAngle(jde float64) float64 {
	return MoonBrightLimbPositionAngleN(jde, -1)
}

// MoonBrightLimbPositionAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MoonBrightLimbPositionAngleNWithInstance(jde float64, n int, inst *Instance) float64 {
	sunRA, sunDec := HSunApparentRaDecNWithInstance(jde, n, inst)
	moonRA, moonDec := HMoonTrueRaDecNWithInstance(jde, n, inst)
	return brightLimbPositionAngleFromRaDec(sunRA, sunDec, moonRA, moonDec)
}

// MoonTopocentricBrightLimbPositionAngle 月亮站心明亮边缘位置角 / topocentric position angle of the Moon's bright limb.
func MoonTopocentricBrightLimbPositionAngle(jde, observerLon, observerLat, height float64) float64 {
	return MoonTopocentricBrightLimbPositionAngleN(jde, observerLon, observerLat, height, -1)
}

// MoonTopocentricBrightLimbPositionAngleNWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MoonTopocentricBrightLimbPositionAngleNWithInstance(jde, observerLon, observerLat, height float64, n int, inst *Instance) float64 {
	sunRA, sunDec := sunTopocentricApparentRaDecN(jde, observerLon, observerLat, height, n, inst)
	moonRA, moonDec := moonTopocentricApparentRaDecN(jde, observerLon, observerLat, height, n, inst)
	return brightLimbPositionAngleFromRaDec(sunRA, sunDec, moonRA, moonDec)
}

func moonTopocentricApparentRaDecN(jde, observerLon, observerLat, height float64, n int, instances ...*Instance) (float64, float64) {
	inst := instanceValue(instances)

	geocentricRA := HMoonTrueRaNWithInstance(jde, n, inst)
	geocentricDec := HMoonTrueDecNWithInstance(jde, n, inst)
	distanceAU := HMoonAwayNWithInstance(jde, n, inst) / moonPhysicalAstronomicalUnitKM
	return TopocentricRaDec(geocentricRA, geocentricDec, observerLat, observerLon, TT2UTC(jde), distanceAU, height)
}

func sunTopocentricApparentRaDecN(jde, observerLon, observerLat, height float64, n int, instances ...*Instance) (float64, float64) {
	inst := instanceValue(instances)

	geocentricRA, geocentricDec := HSunApparentRaDecNWithInstance(jde, n, inst)
	return TopocentricRaDec(geocentricRA, geocentricDec, observerLat, observerLon, TT2UTC(jde), EarthAwayNWithInstance(jde, n, inst), height)
}

func brightLimbPositionAngleFromRaDec(sunRA, sunDec, bodyRA, bodyDec float64) float64 {
	y := Cos(sunDec) * Sin(sunRA-bodyRA)
	x := Sin(sunDec)*Cos(bodyDec) - Cos(sunDec)*Sin(bodyDec)*Cos(sunRA-bodyRA)
	return ArcTan2(y, x)
}
