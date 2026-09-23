package basic

import . "github.com/starainrt/astro/tools"

// OrbitSunDistance 返回轨道目标在给定 TT/TDB 儒略日的日心距离，单位 AU。
func OrbitSunDistance(jde float64, elements OrbitElements) float64 {
	_, _, distance := OrbitHeliocentricEclipticJ2000(jde, elements)
	return distance
}

// OrbitEarthDistance 返回轨道目标在给定 TT/TDB 儒略日的地心距离，单位 AU。
func OrbitEarthDistance(jde float64, elements OrbitElements) float64 {
	_, _, distance := OrbitGeocentricEclipticJ2000(jde, elements)
	return distance
}

// OrbitPhaseAngle 返回轨道目标的相位角，单位度。
func OrbitPhaseAngle(jde float64, elements OrbitElements) float64 {
	return ArcCos(orbitPhaseCosine(jde, elements))
}

// OrbitIlluminatedFraction 返回轨道目标的被照亮比例。
func OrbitIlluminatedFraction(jde float64, elements OrbitElements) float64 {
	return (1 + orbitPhaseCosine(jde, elements)) / 2
}

// OrbitElongation 返回轨道目标相对于太阳的地心视角距，单位度。
func OrbitElongation(jde float64, elements OrbitElements) float64 {
	lon, lat, _ := OrbitApparentGeocentricEcliptic(jde, elements)
	return StarAngularSeparation(lon, lat, HSunApparentLo(jde), HSunTrueBo(jde))
}

func orbitPhaseCosine(jde float64, elements OrbitElements) float64 {
	sunDistance := OrbitSunDistance(jde, elements)
	earthDistance := OrbitEarthDistance(jde, elements)
	earthSunDistance := EarthAway(jde)
	cosine := (sunDistance*sunDistance + earthDistance*earthDistance - earthSunDistance*earthSunDistance) / (2 * sunDistance * earthDistance)
	return clampUnit(cosine)
}
