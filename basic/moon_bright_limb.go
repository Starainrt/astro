package basic

import . "github.com/starainrt/astro/tools"

// MoonBrightLimbPositionAngle 月亮明亮边缘位置角 / position angle of the Moon's bright limb.
func MoonBrightLimbPositionAngle(jd float64) float64 {
	return MoonBrightLimbPositionAngleN(jd, -1)
}

// MoonBrightLimbPositionAngleN 月亮明亮边缘位置角（截断版） / truncated position angle of the Moon's bright limb.
func MoonBrightLimbPositionAngleN(jd float64, n int) float64 {
	sunRA, sunDec := HSunApparentRaDecN(jd, n)
	moonRA, moonDec := HMoonTrueRaDecN(jd, n)
	return brightLimbPositionAngleFromRaDec(sunRA, sunDec, moonRA, moonDec)
}

// MoonTopocentricBrightLimbPositionAngle 月亮站心明亮边缘位置角 / topocentric position angle of the Moon's bright limb.
func MoonTopocentricBrightLimbPositionAngle(jd, observerLon, observerLat, height float64) float64 {
	return MoonTopocentricBrightLimbPositionAngleN(jd, observerLon, observerLat, height, -1)
}

// MoonTopocentricBrightLimbPositionAngleN 月亮站心明亮边缘位置角（截断版） / truncated topocentric position angle of the Moon's bright limb.
func MoonTopocentricBrightLimbPositionAngleN(jd, observerLon, observerLat, height float64, n int) float64 {
	sunRA, sunDec := sunTopocentricApparentRaDecN(jd, observerLon, observerLat, height, n)
	moonRA, moonDec := moonTopocentricApparentRaDecN(jd, observerLon, observerLat, height, n)
	return brightLimbPositionAngleFromRaDec(sunRA, sunDec, moonRA, moonDec)
}

func moonTopocentricApparentRaDecN(jd, observerLon, observerLat, height float64, n int) (float64, float64) {
	geocentricRA := HMoonTrueRaN(jd, n)
	geocentricDec := HMoonTrueDecN(jd, n)
	distanceAU := HMoonAwayN(jd, n) / moonPhysicalAstronomicalUnitKM
	return TopocentricRaDec(geocentricRA, geocentricDec, observerLat, observerLon, TD2UT(jd, false), distanceAU, height)
}

func sunTopocentricApparentRaDecN(jd, observerLon, observerLat, height float64, n int) (float64, float64) {
	geocentricRA, geocentricDec := HSunApparentRaDecN(jd, n)
	return TopocentricRaDec(geocentricRA, geocentricDec, observerLat, observerLon, TD2UT(jd, false), EarthAwayN(jd, n), height)
}

func brightLimbPositionAngleFromRaDec(sunRA, sunDec, bodyRA, bodyDec float64) float64 {
	y := Cos(sunDec) * Sin(sunRA-bodyRA)
	x := Sin(sunDec)*Cos(bodyDec) - Cos(sunDec)*Sin(bodyDec)*Cos(sunRA-bodyRA)
	return ArcTan2(y, x)
}
