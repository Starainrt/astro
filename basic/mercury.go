package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func MercuryL(jde float64) float64 {
	return planet.WherePlanet(1, 0, jde)
}

func MercuryB(jde float64) float64 {
	return planet.WherePlanet(1, 1, jde)
}
func MercuryR(jde float64) float64 {
	return planet.WherePlanet(1, 2, jde)
}
func AMercuryX(jde float64) float64 {
	l := MercuryL(jde)
	b := MercuryB(jde)
	r := MercuryR(jde)
	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	return x
}

func AMercuryY(jde float64) float64 {

	l := MercuryL(jde)
	b := MercuryB(jde)
	r := MercuryR(jde)
	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	return y
}
func AMercuryZ(jde float64) float64 {
	//l := MercuryL(jde)
	b := MercuryB(jde)
	r := MercuryR(jde)
	//	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	z := r*Sin(b) - er*Sin(eb)
	return z
}

func AMercuryXYZ(jde float64) (float64, float64, float64) {
	l := MercuryL(jde)
	b := MercuryB(jde)
	r := MercuryR(jde)
	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func MercuryApparentRa(jde float64) float64 {
	lo, bo := MercuryApparentLoBo(jde)
	return LoToRa(jde, lo, bo)
}
func MercuryApparentDec(jde float64) float64 {
	lo, bo := MercuryApparentLoBo(jde)
	eps := TrueObliquity(jde)
	dec := ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
	return dec
}

func MercuryApparentRaDec(jde float64) (float64, float64) {
	lo, bo := MercuryApparentLoBo(jde)
	return LoBoToRaDec(jde, lo, bo)
}

func EarthMercuryAway(jd float64) float64 {
	return planetEarthAwayExplicitN(1, jd, -1)
}

func MercuryApparentLo(jd float64) float64 {
	geo, _ := planetApparentGeocentricPositionN(1, jd, -1)
	return geo.lo
}

func MercuryApparentBo(jd float64) float64 {
	geo, _ := planetApparentGeocentricPositionN(1, jd, -1)
	return geo.bo
}

func MercuryApparentLoBo(jd float64) (float64, float64) {
	geo, _ := planetApparentGeocentricPositionN(1, jd, -1)
	return geo.lo, geo.bo
}

func MercuryMag(jde float64) float64 {
	sunDistance := MercuryR(jde)
	earthDistance := EarthMercuryAway(jde)
	earthSunDistance := planet.WherePlanet(-1, 2, jde)
	i := (sunDistance*sunDistance + earthDistance*earthDistance - earthSunDistance*earthSunDistance) / (2 * sunDistance * earthDistance)
	i = ArcCos(i)
	mag := -0.42 + 5*math.Log10(sunDistance*earthDistance) + 0.0380*i - 0.000273*i*i + 0.000002*i*i*i
	return FloatRound(mag, 2)
}

// MercuryHeightWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercuryHeightWithInstance(localJD, lon, lat, timezone float64, inst *Instance) float64 {
	return planetHeightN(1, localJD, lon, lat, timezone, -1, inst)
}

// MercuryAzimuthWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercuryAzimuthWithInstance(localJD, lon, lat, timezone float64, inst *Instance) float64 {
	return planetAzimuthN(1, localJD, lon, lat, timezone, -1, inst)
}

// MercuryHourAngleWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercuryHourAngleWithInstance(jd, lon, timezone float64, inst *Instance) float64 {
	siderealLongitude := Limit360(ApparentSiderealTime(UTC2UT1(jd-timezone/24))*15 + lon)
	hourAngle := siderealLongitude - MercuryApparentRaWithInstance(UTC2TT(jd-timezone/24.0), inst)
	if hourAngle < 0 {
		hourAngle += 360
	}
	return hourAngle
}

// MercuryCulminationTimeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercuryCulminationTimeWithInstance(localJD, lon, timezone float64, inst *Instance) float64 {
	// localJD 是本地民用日锚点（当地 0 时），不是力学时。
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	localJD = math.Floor(localJD) + 0.5
	estimateJD := localJD + Limit360(360-MercuryHourAngleWithInstance(localJD, lon, timezone, inst))/15.0/24.0*0.99726851851851851851
	normalizedHourAngle := func(localJD, lon, timezone float64) float64 {
		currentHourAngle := MercuryHourAngleWithInstance(localJD, lon, timezone, inst)
		if currentHourAngle < 180 {
			currentHourAngle += 360
		}
		return currentHourAngle
	}
	var ok bool
	estimateJD, ok = eventNewtonRefine(estimateJD, 0.00001, func(prevJD float64) float64 {
		hourAngleDelta := normalizedHourAngle(prevJD, lon, timezone) - 360
		hourAngleSlope := (normalizedHourAngle(prevJD+0.000005, lon, timezone) - normalizedHourAngle(prevJD-0.000005, lon, timezone)) / 0.00001
		return hourAngleDelta / hourAngleSlope
	})
	if !ok {
		return math.NaN()
	}
	return estimateJD
}

// MercuryRiseTimeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercuryRiseTimeWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, inst *Instance) (float64, error) {
	return mercuryRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight, true, inst)
}

// MercurySetTimeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MercurySetTimeWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, inst *Instance) (float64, error) {
	return mercuryRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight, false, inst)
}

func mercuryRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, isRise bool, instances ...*Instance) (float64, error) {
	inst := instanceValue(instances)

	return planetRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight, isRise, func(localJD, lon, timezone float64) float64 {
		return MercuryCulminationTimeWithInstance(localJD, lon, timezone, inst)
	}, func(localJD, lon, lat, timezone float64) float64 {
		return MercuryHeightWithInstance(localJD, lon, lat, timezone, inst)
	}, func(jde float64) float64 { return MercuryApparentDecWithInstance(jde, inst) })
}
