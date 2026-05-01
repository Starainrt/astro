package basic

import (
	"math"

	"github.com/starainrt/astro/planet"
	. "github.com/starainrt/astro/tools"
)

var orbitJ2000Obliquity = EclipticObliquity(orbitReferenceJD, false)

// OrbitHeliocentricXYZJ2000 返回日心 J2000 平黄道直角坐标，单位 AU。
func OrbitHeliocentricXYZJ2000(jd float64, elements OrbitElements) Vector3 {
	trueAnomaly, radius, resolved, ok := orbitTrueAnomalyAndRadius(jd, elements)
	if !ok {
		nan := math.NaN()
		return Vector3{nan, nan, nan}
	}

	ascendingNode := resolved.Omega * rad
	argumentLatitude := resolved.W*rad + trueAnomaly
	inclination := resolved.I * rad

	sinAscendingNode, cosAscendingNode := math.Sincos(ascendingNode)
	sinArgumentLatitude, cosArgumentLatitude := math.Sincos(argumentLatitude)
	sinInclination, cosInclination := math.Sincos(inclination)

	return Vector3{
		radius * (cosAscendingNode*cosArgumentLatitude - sinAscendingNode*sinArgumentLatitude*cosInclination),
		radius * (sinAscendingNode*cosArgumentLatitude + cosAscendingNode*sinArgumentLatitude*cosInclination),
		radius * sinArgumentLatitude * sinInclination,
	}
}

// OrbitHeliocentricEclipticJ2000 返回日心 J2000 平黄道球坐标，单位度/AU。
func OrbitHeliocentricEclipticJ2000(jd float64, elements OrbitElements) (lon, lat, distance float64) {
	return orbitVectorToEcliptic(OrbitHeliocentricXYZJ2000(jd, elements))
}

// OrbitHeliocentricXYZ 返回日心历元黄道直角坐标，单位 AU。
func OrbitHeliocentricXYZ(jd float64, elements OrbitElements) Vector3 {
	return eclipticVectorAtReferenceEpoch(OrbitHeliocentricXYZJ2000(jd, elements), orbitReferenceJD, jd)
}

// OrbitHeliocentricEcliptic 返回日心历元黄道球坐标，单位度/AU。
func OrbitHeliocentricEcliptic(jd float64, elements OrbitElements) (lon, lat, distance float64) {
	return orbitVectorToEcliptic(OrbitHeliocentricXYZ(jd, elements))
}

// OrbitGeocentricXYZJ2000 返回地心 J2000 平黄道直角坐标，单位 AU。
func OrbitGeocentricXYZJ2000(jd float64, elements OrbitElements) Vector3 {
	objectVector := OrbitHeliocentricXYZJ2000(jd, elements)
	earthVector := earthHeliocentricVectorJ2000(jd)
	return Vector3{
		objectVector[0] - earthVector[0],
		objectVector[1] - earthVector[1],
		objectVector[2] - earthVector[2],
	}
}

// OrbitGeocentricEclipticJ2000 返回地心 J2000 平黄道球坐标，单位度/AU。
func OrbitGeocentricEclipticJ2000(jd float64, elements OrbitElements) (lon, lat, distance float64) {
	return orbitVectorToEcliptic(OrbitGeocentricXYZJ2000(jd, elements))
}

// OrbitGeocentricXYZ 返回地心历元黄道直角坐标，单位 AU。
func OrbitGeocentricXYZ(jd float64, elements OrbitElements) Vector3 {
	objectVector := OrbitHeliocentricXYZ(jd, elements)
	earthVector := earthHeliocentricVectorOfDate(jd)
	return Vector3{
		objectVector[0] - earthVector[0],
		objectVector[1] - earthVector[1],
		objectVector[2] - earthVector[2],
	}
}

// OrbitGeocentricEcliptic 返回地心历元黄道球坐标，单位度/AU。
func OrbitGeocentricEcliptic(jd float64, elements OrbitElements) (lon, lat, distance float64) {
	return orbitVectorToEcliptic(OrbitGeocentricXYZ(jd, elements))
}

// OrbitGeocentricEquatorialJ2000 返回地心 J2000 平赤道球坐标，单位度/AU。
func OrbitGeocentricEquatorialJ2000(jd float64, elements OrbitElements) (ra, dec, distance float64) {
	vector := rotateEclipticToEquatorial(OrbitGeocentricXYZJ2000(jd, elements), orbitJ2000Obliquity)
	return orbitVectorToEquatorial(vector)
}

// OrbitGeocentricEquatorial 返回地心历元平赤道球坐标，单位度/AU。
func OrbitGeocentricEquatorial(jd float64, elements OrbitElements) (ra, dec, distance float64) {
	vector := rotateEclipticToEquatorial(OrbitGeocentricXYZ(jd, elements), EclipticObliquity(jd, false))
	return orbitVectorToEquatorial(vector)
}

// OrbitAstrometricGeocentricXYZJ2000 返回光行时修正后的地心 J2000 平黄道直角坐标，单位 AU。
func OrbitAstrometricGeocentricXYZJ2000(jd float64, elements OrbitElements) Vector3 {
	if !isFinite(jd) {
		nan := math.NaN()
		return Vector3{nan, nan, nan}
	}
	earthVector := earthHeliocentricVectorJ2000(jd)
	lightTime := 0.0
	result := Vector3{}
	for i := 0; i < 8; i++ {
		objectVector := OrbitHeliocentricXYZJ2000(jd-lightTime, elements)
		result = Vector3{
			objectVector[0] - earthVector[0],
			objectVector[1] - earthVector[1],
			objectVector[2] - earthVector[2],
		}
		nextLightTime := lightTimeDaysPerAU * orbitVectorNorm(result)
		if math.Abs(nextLightTime-lightTime) < 1e-12 {
			break
		}
		lightTime = nextLightTime
	}
	return result
}

// OrbitAstrometricGeocentricEquatorialJ2000 返回光行时修正后的地心 J2000 赤道坐标，单位度/AU。
func OrbitAstrometricGeocentricEquatorialJ2000(jd float64, elements OrbitElements) (ra, dec, distance float64) {
	vector := rotateEclipticToEquatorial(OrbitAstrometricGeocentricXYZJ2000(jd, elements), orbitJ2000Obliquity)
	return orbitVectorToEquatorial(vector)
}

// OrbitApparentGeocentricEcliptic 返回光行时与章动修正后的地心视黄道坐标，单位度/AU。
func OrbitApparentGeocentricEcliptic(jd float64, elements OrbitElements) (lon, lat, distance float64) {
	vectorDate := eclipticVectorAtReferenceEpoch(OrbitAstrometricGeocentricXYZJ2000(jd, elements), orbitReferenceJD, jd)
	lon, lat, distance = orbitVectorToEcliptic(vectorDate)
	if math.IsNaN(lon) {
		return math.NaN(), math.NaN(), math.NaN()
	}
	lon = Limit360(lon + Nutation2000Bi(jd))
	return lon, lat, distance
}

// OrbitApparentGeocentricEquatorial 返回光行时与章动修正后的地心视赤道坐标，单位度/AU。
func OrbitApparentGeocentricEquatorial(jd float64, elements OrbitElements) (ra, dec, distance float64) {
	lon, lat, distance := OrbitApparentGeocentricEcliptic(jd, elements)
	if math.IsNaN(lon) {
		return math.NaN(), math.NaN(), math.NaN()
	}
	ra, dec = LoBoToRaDec(jd, lon, lat)
	return ra, dec, distance
}

// OrbitApparentTopocentricEquatorial 返回光行时、章动与站心修正后的视赤道坐标，单位度/AU。
func OrbitApparentTopocentricEquatorial(jd, observerLon, observerLat, observerHeight float64, elements OrbitElements) (ra, dec, distance float64) {
	geocentricRA, geocentricDec, geocentricDistance := OrbitApparentGeocentricEquatorial(jd, elements)
	if math.IsNaN(geocentricRA) {
		return math.NaN(), math.NaN(), math.NaN()
	}
	geocentricVector := orbitEquatorialVector(geocentricRA, geocentricDec, geocentricDistance)
	observerVector := orbitObserverEquatorialVectorOfDate(TD2UT(jd, false), observerLon, observerLat, observerHeight)
	topocentricVector := Vector3{
		geocentricVector[0] - observerVector[0],
		geocentricVector[1] - observerVector[1],
		geocentricVector[2] - observerVector[2],
	}
	return orbitVectorToEquatorial(topocentricVector)
}

func earthHeliocentricVectorOfDate(jd float64) Vector3 {
	return eclipticCartesian(
		planet.WherePlanet(-1, 0, jd),
		planet.WherePlanet(-1, 1, jd),
		planet.WherePlanet(-1, 2, jd),
	)
}

func earthHeliocentricVectorJ2000(jd float64) Vector3 {
	return eclipticVectorAtReferenceEpoch(earthHeliocentricVectorOfDate(jd), jd, orbitReferenceJD)
}

func orbitVectorToEcliptic(vector Vector3) (lon, lat, distance float64) {
	distance = orbitVectorNorm(vector)
	if math.IsNaN(distance) || math.IsInf(distance, 0) {
		return math.NaN(), math.NaN(), math.NaN()
	}
	if distance == 0 {
		return 0, 0, 0
	}
	lon = Limit360(math.Atan2(vector[1], vector[0]) * deg)
	lat = math.Asin(orbitClampUnit(vector[2]/distance)) * deg
	return lon, lat, distance
}

func orbitVectorToEquatorial(vector Vector3) (ra, dec, distance float64) {
	distance = orbitVectorNorm(vector)
	if math.IsNaN(distance) || math.IsInf(distance, 0) {
		return math.NaN(), math.NaN(), math.NaN()
	}
	if distance == 0 {
		return 0, 0, 0
	}
	ra = Limit360(math.Atan2(vector[1], vector[0]) * deg)
	dec = math.Asin(orbitClampUnit(vector[2]/distance)) * deg
	return ra, dec, distance
}

func orbitEquatorialVector(ra, dec, distance float64) Vector3 {
	cosDec := Cos(dec)
	return Vector3{
		distance * cosDec * Cos(ra),
		distance * cosDec * Sin(ra),
		distance * Sin(dec),
	}
}

func orbitObserverEquatorialVectorOfDate(jdUT, observerLon, observerLat, observerHeight float64) Vector3 {
	localApparentSiderealLongitude := Limit360(ApparentSiderealTime(jdUT)*15 + observerLon)
	observerScaleAU := Sin(0.0024427777777)
	rhoCosPhiPrime := pcosi(observerLat, observerHeight)
	rhoSinPhiPrime := psini(observerLat, observerHeight)
	return Vector3{
		observerScaleAU * rhoCosPhiPrime * Cos(localApparentSiderealLongitude),
		observerScaleAU * rhoCosPhiPrime * Sin(localApparentSiderealLongitude),
		observerScaleAU * rhoSinPhiPrime,
	}
}

func orbitVectorNorm(vector Vector3) float64 {
	return math.Sqrt(vector[0]*vector[0] + vector[1]*vector[1] + vector[2]*vector[2])
}

func orbitClampUnit(value float64) float64 {
	if value > 1 {
		return 1
	}
	if value < -1 {
		return -1
	}
	return value
}
