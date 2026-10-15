package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

func UranusL(jde float64) float64 {
	return planet.WherePlanet(6, 0, jde)
}

func UranusB(jde float64) float64 {
	return planet.WherePlanet(6, 1, jde)
}
func UranusR(jde float64) float64 {
	return planet.WherePlanet(6, 2, jde)
}
func AUranusX(jde float64) float64 {
	l := UranusL(jde)
	b := UranusB(jde)
	r := UranusR(jde)
	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	return x
}

func AUranusY(jde float64) float64 {

	l := UranusL(jde)
	b := UranusB(jde)
	r := UranusR(jde)
	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	return y
}
func AUranusZ(jde float64) float64 {
	//l := UranusL(jde)
	b := UranusB(jde)
	r := UranusR(jde)
	//	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	z := r*Sin(b) - er*Sin(eb)
	return z
}

func AUranusXYZ(jde float64) (float64, float64, float64) {
	l := UranusL(jde)
	b := UranusB(jde)
	r := UranusR(jde)
	el := planet.WherePlanet(-1, 0, jde)
	eb := planet.WherePlanet(-1, 1, jde)
	er := planet.WherePlanet(-1, 2, jde)
	x := r*Cos(b)*Cos(l) - er*Cos(eb)*Cos(el)
	y := r*Cos(b)*Sin(l) - er*Cos(eb)*Sin(el)
	z := r*Sin(b) - er*Sin(eb)
	return x, y, z
}

func UranusApparentRa(jde float64) float64 {
	lo, bo := UranusApparentLoBo(jde)
	eps := TrueObliquity(jde)
	ra := math.Atan2((Sin(lo)*Cos(eps) - Tan(bo)*Sin(eps)), Cos(lo))
	ra = ra * 180 / math.Pi
	return Limit360(ra)
}
func UranusApparentDec(jde float64) float64 {
	lo, bo := UranusApparentLoBo(jde)
	eps := TrueObliquity(jde)
	dec := ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
	return dec
}

func UranusApparentRaDec(jde float64) (float64, float64) {
	lo, bo := UranusApparentLoBo(jde)
	eps := TrueObliquity(jde)
	ra := math.Atan2((Sin(lo)*Cos(eps) - Tan(bo)*Sin(eps)), Cos(lo))
	ra = ra * 180 / math.Pi
	dec := ArcSin(Sin(bo)*Cos(eps) + Cos(bo)*Sin(eps)*Sin(lo))
	return Limit360(ra), dec
}

func EarthUranusAway(jd float64) float64 {
	return planetEarthAwayExplicitN(6, jd, -1)
}

func UranusApparentLo(jd float64) float64 {
	geo, _ := planetApparentGeocentricPositionN(6, jd, -1)
	return geo.lo
}

func UranusApparentBo(jd float64) float64 {
	geo, _ := planetApparentGeocentricPositionN(6, jd, -1)
	return geo.bo
}

func UranusApparentLoBo(jd float64) (float64, float64) {
	geo, _ := planetApparentGeocentricPositionN(6, jd, -1)
	return geo.lo, geo.bo
}

func UranusMag(jde float64) float64 {
	sunDistance := UranusR(jde)
	earthDistance := EarthUranusAway(jde)
	earthSunDistance := planet.WherePlanet(-1, 2, jde)
	i := (sunDistance*sunDistance + earthDistance*earthDistance - earthSunDistance*earthSunDistance) / (2 * sunDistance * earthDistance)
	i = ArcCos(i)
	mag := -7.19 + 5*math.Log10(sunDistance*earthDistance) + 0.016*i
	return FloatRound(mag, 2)
}

// UranusHeightWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusHeightWithInstance(localJD, lon, lat, timezone float64, inst *Instance) float64 {
	return planetHeightN(6, localJD, lon, lat, timezone, -1, inst)
}

// UranusAzimuthWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusAzimuthWithInstance(localJD, lon, lat, timezone float64, inst *Instance) float64 {
	return planetAzimuthN(6, localJD, lon, lat, timezone, -1, inst)
}

// UranusHourAngleWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusHourAngleWithInstance(jd, lon, timezone float64, inst *Instance) float64 {
	siderealLongitude := Limit360(ApparentSiderealTime(UTC2UT1(jd-timezone/24))*15 + lon)
	hourAngle := siderealLongitude - UranusApparentRaWithInstance(UTC2TT(jd-timezone/24.0), inst)
	if hourAngle < 0 {
		hourAngle += 360
	}
	return hourAngle
}

// UranusCulminationTimeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusCulminationTimeWithInstance(localJD, lon, timezone float64, inst *Instance) float64 {
	// localJD 是本地民用日锚点（当地 0 时），不是力学时。
	//ra,dec 瞬时天球座标，非J2000等时间天球坐标
	localJD = math.Floor(localJD) + 0.5
	estimateJD := localJD + Limit360(360-UranusHourAngleWithInstance(localJD, lon, timezone, inst))/15.0/24.0*0.99726851851851851851
	normalizedHourAngle := func(localJD, lon, timezone float64) float64 {
		currentHourAngle := UranusHourAngleWithInstance(localJD, lon, timezone, inst)
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

// UranusRiseTimeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusRiseTimeWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, inst *Instance) (float64, error) {
	return uranusRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight, true, inst)
}

// UranusSetTimeWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func UranusSetTimeWithInstance(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, inst *Instance) (float64, error) {
	return uranusRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight, false, inst)
}

func uranusRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight float64, isRise bool, instances ...*Instance) (float64, error) {
	inst := instanceValue(instances)

	return planetRiseDown(jd, lon, lat, timezone, aeroCorrection, observerHeight, isRise, func(localJD, lon, timezone float64) float64 {
		return UranusCulminationTimeWithInstance(localJD, lon, timezone, inst)
	}, func(localJD, lon, lat, timezone float64) float64 {
		return UranusHeightWithInstance(localJD, lon, lat, timezone, inst)
	}, func(jde float64) float64 { return UranusApparentDecWithInstance(jde, inst) })
}
