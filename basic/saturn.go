package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func SaturnL(jde float64) float64 {
	return planet.WherePlanet(5, 0, jde)
}

func SaturnB(jde float64) float64 {
	return planet.WherePlanet(5, 1, jde)
}
func SaturnR(jde float64) float64 {
	return planet.WherePlanet(5, 2, jde)
}
func ASaturnX(jde float64) float64 {
	l := SaturnL(jde)
	b := SaturnB(jde)
	r := SaturnR(jde)
	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	return x
}

func ASaturnY(jde float64) float64 {

	l := SaturnL(jde)
	b := SaturnB(jde)
	r := SaturnR(jde)
	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	return y
}
func ASaturnZ(jde float64) float64 {
	//l := SaturnL(jde)
	b := SaturnB(jde)
	r := SaturnR(jde)
	//	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	z := r*Sin(b) - er*Sin(eb)
	return z
}

func ASaturnXYZ(jde float64) (float64, float64, float64) {
	l := SaturnL(jde)
	b := SaturnB(jde)
	r := SaturnR(jde)
	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func SaturnApparentRa(jde float64) float64 {
	lo, bo := SaturnApparentLoBo(jde)
	eps := TrueObliquity(jde)
	ra := math.Atan2((Sin(lo)*Cos(eps) - Tan(bo)*Sin(eps)), Cos(lo))
	ra = ra * 180 / math.Pi
	return Limit360(ra)
}
func SaturnApparentDec(jde float64) float64 {
	lo, bo := SaturnApparentLoBo(jde)
	eps := TrueObliquity(jde)
	dec := ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
	return dec
}

func SaturnApparentRaDec(jde float64) (float64, float64) {
	lo, bo := SaturnApparentLoBo(jde)
	eps := TrueObliquity(jde)
	ra := math.Atan2((Sin(lo)*Cos(eps) - Tan(bo)*Sin(eps)), Cos(lo))
	ra = ra * 180 / math.Pi
	dec := ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
	return Limit360(ra), dec
}

func EarthSaturnAway(jd float64) float64 {
	return planetEarthAwayExplicitN(5, jd, -1)
}

func SaturnApparentLo(jd float64) float64 {
	geo, _ := planetApparentGeocentricPositionN(5, jd, -1)
	return geo.lo
}

func SaturnApparentBo(jd float64) float64 {
	geo, _ := planetApparentGeocentricPositionN(5, jd, -1)
	return geo.bo
}

func SaturnApparentLoBo(jd float64) (float64, float64) {
	geo, _ := planetApparentGeocentricPositionN(5, jd, -1)
	return geo.lo, geo.bo
}

func SaturnMag(jde float64) float64 {
	return SaturnMagN(jde, -1)
}

// SaturnHeightWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnHeightWithInstance(localJD, lon, lat, timezone float64, inst *Instance) float64 {
	return planetHeightN(5, localJD, lon, lat, timezone, -1, inst)
}

// SaturnAzimuthWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnAzimuthWithInstance(localJD, lon, lat, timezone float64, inst *Instance) float64 {
	return planetAzimuthN(5, localJD, lon, lat, timezone, -1, inst)
}

// SaturnHourAngleWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnHourAngleWithInstance(jd, lon, timezone float64, inst *Instance) float64 {
	siderealLongitude := Limit360(ApparentSiderealTime(UTC2UT1(jd-timezone/24))*15 + lon)
	hourAngle := siderealLongitude - SaturnApparentRaWithInstance(UTC2TT(jd-timezone/24.0), inst)
	if hourAngle < 0 {
		hourAngle += 360
	}
	return hourAngle
}

// SaturnCulminationTimeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnCulminationTimeWithInstance(localJD, lon, timezone float64, inst *Instance) float64 {
	// localJD 是本地民用日锚点（当地 0 时），不是力学时。
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	localJD = math.Floor(localJD) + 0.5
	estimateJD := localJD + Limit360(360-SaturnHourAngleWithInstance(localJD, lon, timezone, inst))/15.0/24.0*0.99726851851851851851
	normalizedHourAngle := func(localJD, lon, timezone float64) float64 {
		currentHourAngle := SaturnHourAngleWithInstance(localJD, lon, timezone, inst)
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

// SaturnRiseTimeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnRiseTimeWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, inst *Instance) (float64, error) {
	return saturnRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight, true, inst)
}

// SaturnSetTimeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SaturnSetTimeWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, inst *Instance) (float64, error) {
	return saturnRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight, false, inst)
}

func saturnRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, isRise bool, instances ...*Instance) (float64, error) {
	inst := instanceValue(instances)

	return planetRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight, isRise, func(localJD, lon, timezone float64) float64 {
		return SaturnCulminationTimeWithInstance(localJD, lon, timezone, inst)
	}, func(localJD, lon, lat, timezone float64) float64 {
		return SaturnHeightWithInstance(localJD, lon, lat, timezone, inst)
	}, func(jde float64) float64 { return SaturnApparentDecWithInstance(jde, inst) })
}
