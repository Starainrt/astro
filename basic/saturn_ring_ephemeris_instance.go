package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

func saturnRingSunLoBoWithInstance(jde float64, n int, node float64, inst *Instance) (lon, lat float64) {
	lon = PlanetCoordinateNWithInstance(5, 0, jde, n, inst)
	lat = PlanetCoordinateNWithInstance(5, 1, jde, n, inst)
	distance := PlanetCoordinateNWithInstance(5, 2, jde, n, inst)
	lat -= 0.000764 * Cos(lon-node) / distance
	lon -= 0.01759 / distance
	return lon, lat
}

func saturnRingParametersNWithInstance(jde float64, n int, inst *Instance) (earthLatitude, sunLatitude, positionAngle, deltaU, majorAxis, minorAxis float64) {
	inclination, node := saturnRingPlane(jde)
	earthLon, earthLat := SaturnApparentLoBoNWithInstance(jde, n, inst)
	sunLon, sunLat := saturnRingSunLoBoWithInstance(jde, n, node, inst)
	earthLatitude = saturnRingLatitude(inclination, node, earthLon, earthLat)
	sunLatitude = saturnRingLatitude(inclination, node, sunLon, sunLat)
	positionAngle = saturnRingPositionAngle(jde, inclination, node, earthLon, earthLat)
	earthU := saturnRingLongitude(inclination, node, earthLon, earthLat)
	sunU := saturnRingLongitude(inclination, node, sunLon, sunLat)
	deltaU = saturnRingLongitudeDelta(earthU, sunU)
	earthDistance := EarthSaturnAwayNWithInstance(jde, n, inst)
	majorAxis = 375.35 / earthDistance
	minorAxis = majorAxis * math.Abs(Sin(earthLatitude))
	return
}

func saturnRingParametersWithInstance(jde float64, inst *Instance) (earthLatitude, sunLatitude, positionAngle, deltaU, majorAxis, minorAxis float64) {
	return saturnRingParametersNWithInstance(jde, -1, inst)
}

// SaturnRingParametersWithInstance 土星环参数 / Saturn ring parameters with an explicit ephemeris instance.
func SaturnRingParametersWithInstance(jde float64, inst *Instance) (earthLatitude, sunLatitude, positionAngle, deltaU, majorAxis, minorAxis float64) {
	if inst == nil {
		return SaturnRingParameters(jde)
	}
	return saturnRingParametersWithInstance(jde, inst)
}

// SaturnRingParametersNWithInstance 土星环参数（n 变体） / n-parameterised Saturn ring parameters with an explicit ephemeris instance.
func SaturnRingParametersNWithInstance(jde float64, n int, inst *Instance) (earthLatitude, sunLatitude, positionAngle, deltaU, majorAxis, minorAxis float64) {
	if inst == nil {
		return SaturnRingParametersN(jde, n)
	}
	return saturnRingParametersNWithInstance(jde, n, inst)
}

// SaturnRingBWithInstance 土星环张角 / Saturn ring opening angle with an explicit ephemeris instance.
func SaturnRingBWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return SaturnRingB(jde)
	}
	earthLatitude, _, _, _, _, _ := saturnRingParametersWithInstance(jde, inst)
	return earthLatitude
}

// SaturnRingSunBWithInstance 土星环太阳侧张角 / Sun-side Saturn ring opening angle with an explicit ephemeris instance.
func SaturnRingSunBWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return SaturnRingSunB(jde)
	}
	_, sunLatitude, _, _, _, _ := saturnRingParametersWithInstance(jde, inst)
	return sunLatitude
}

// SaturnRingPositionAngleWithInstance 土星环位置角 / Saturn ring position angle with an explicit ephemeris instance.
func SaturnRingPositionAngleWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return SaturnRingPositionAngle(jde)
	}
	_, _, positionAngle, _, _, _ := saturnRingParametersWithInstance(jde, inst)
	return positionAngle
}

// SaturnRingDeltaUWithInstance 土星环面经度差 / Saturn ring-plane longitude difference with an explicit ephemeris instance.
func SaturnRingDeltaUWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return SaturnRingDeltaU(jde)
	}
	_, _, _, deltaU, _, _ := saturnRingParametersWithInstance(jde, inst)
	return deltaU
}

// SaturnRingAxisWithInstance 土星环外缘轴 / outer Saturn ring axes with an explicit ephemeris instance.
func SaturnRingAxisWithInstance(jde float64, inst *Instance) (float64, float64) {
	if inst == nil {
		return SaturnRingAxis(jde)
	}
	_, _, _, _, majorAxis, minorAxis := saturnRingParametersWithInstance(jde, inst)
	return majorAxis, minorAxis
}

// SaturnMagWithInstance 土星视星等 / apparent magnitude of Saturn with an explicit ephemeris instance.
func SaturnMagWithInstance(jde float64, inst *Instance) float64 {
	if inst == nil {
		return SaturnMag(jde)
	}
	awaySun := SaturnRWithInstance(jde, inst)
	awayEarth := EarthSaturnAwayWithInstance(jde, inst)
	ringB, _, _, deltaU, _, _ := saturnRingParametersWithInstance(jde, inst)
	ringB = math.Abs(ringB)
	mag := -8.68 + 5*math.Log10(awaySun*awayEarth) + 0.044*deltaU - 2.6*Sin(ringB) + 1.25*Sin(ringB)*Sin(ringB)
	return FloatRound(mag, 2)
}
