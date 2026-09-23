package basic

import . "github.com/starainrt/astro/tools"

// MoonBrightLimbPositionAngle 月亮明亮边缘位置角 / position angle of the Moon's bright limb.
func MoonBrightLimbPositionAngle(jde float64) float64 {
	return MoonBrightLimbPositionAngleN(jde, -1)
}

// MoonBrightLimbPositionAngleN 月亮明亮边缘位置角（截断版） / truncated position angle of the Moon's bright limb.
func MoonBrightLimbPositionAngleN(jde float64, n int) float64 {
	sunRA, sunDec := HSunApparentRaDecN(jde, n)
	moonRA, moonDec := HMoonTrueRaDecN(jde, n)
	return brightLimbPositionAngleFromRaDec(sunRA, sunDec, moonRA, moonDec)
}

// MoonTopocentricBrightLimbPositionAngle 月亮站心明亮边缘位置角 / topocentric position angle of the Moon's bright limb.
func MoonTopocentricBrightLimbPositionAngle(jde, observerLon, observerLat, height float64) float64 {
	return MoonTopocentricBrightLimbPositionAngleN(jde, observerLon, observerLat, height, -1)
}

// MoonTopocentricBrightLimbPositionAngleN 月亮站心明亮边缘位置角（截断版） / truncated topocentric position angle of the Moon's bright limb.
func MoonTopocentricBrightLimbPositionAngleN(jde, observerLon, observerLat, height float64, n int) float64 {
	sunRA, sunDec := sunTopocentricApparentRaDecN(jde, observerLon, observerLat, height, n)
	moonRA, moonDec := moonTopocentricApparentRaDecN(jde, observerLon, observerLat, height, n)
	return brightLimbPositionAngleFromRaDec(sunRA, sunDec, moonRA, moonDec)
}

func moonTopocentricApparentRaDecN(jde, observerLon, observerLat, height float64, n int) (float64, float64) {
	geocentricRA := HMoonTrueRaN(jde, n)
	geocentricDec := HMoonTrueDecN(jde, n)
	distanceAU := HMoonAwayN(jde, n) / moonPhysicalAstronomicalUnitKM
	return TopocentricRaDec(geocentricRA, geocentricDec, observerLat, observerLon, TT2UTC(jde), distanceAU, height)
}

func sunTopocentricApparentRaDecN(jde, observerLon, observerLat, height float64, n int) (float64, float64) {
	geocentricRA, geocentricDec := HSunApparentRaDecN(jde, n)
	return TopocentricRaDec(geocentricRA, geocentricDec, observerLat, observerLon, TT2UTC(jde), EarthAwayN(jde, n), height)
}

func brightLimbPositionAngleFromRaDec(sunRA, sunDec, bodyRA, bodyDec float64) float64 {
	y := Cos(sunDec) * Sin(sunRA-bodyRA)
	x := Sin(sunDec)*Cos(bodyDec) - Cos(sunDec)*Sin(bodyDec)*Cos(sunRA-bodyRA)
	return ArcTan2(y, x)
}
