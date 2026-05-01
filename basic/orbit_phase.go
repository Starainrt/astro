package basic

import . "github.com/starainrt/astro/tools"

// OrbitSunDistance 返回轨道目标在给定 TT/TDB 儒略日的日心距离，单位 AU。
func OrbitSunDistance(jd float64, elements OrbitElements) float64 {
	_, _, distance := OrbitHeliocentricEclipticJ2000(jd, elements)
	return distance
}

// OrbitEarthDistance 返回轨道目标在给定 TT/TDB 儒略日的地心距离，单位 AU。
func OrbitEarthDistance(jd float64, elements OrbitElements) float64 {
	_, _, distance := OrbitGeocentricEclipticJ2000(jd, elements)
	return distance
}

// OrbitPhaseAngle 返回轨道目标的相位角，单位度。
func OrbitPhaseAngle(jd float64, elements OrbitElements) float64 {
	return ArcCos(orbitPhaseCosine(jd, elements))
}

// OrbitIlluminatedFraction 返回轨道目标的被照亮比例。
func OrbitIlluminatedFraction(jd float64, elements OrbitElements) float64 {
	return (1 + orbitPhaseCosine(jd, elements)) / 2
}

// OrbitElongation 返回轨道目标相对于太阳的地心视角距，单位度。
func OrbitElongation(jd float64, elements OrbitElements) float64 {
	lon, lat, _ := OrbitApparentGeocentricEcliptic(jd, elements)
	return StarAngularSeparation(lon, lat, HSunApparentLo(jd), HSunTrueBo(jd))
}

func orbitPhaseCosine(jd float64, elements OrbitElements) float64 {
	sunDistance := OrbitSunDistance(jd, elements)
	earthDistance := OrbitEarthDistance(jd, elements)
	earthSunDistance := EarthAway(jd)
	cosine := (sunDistance*sunDistance + earthDistance*earthDistance - earthSunDistance*earthSunDistance) / (2 * sunDistance * earthDistance)
	return clampUnit(cosine)
}
