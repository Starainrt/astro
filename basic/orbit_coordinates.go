package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

var orbitJ2000Obliquity = EclipticObliquity(orbitReferenceJD, false)

// OrbitHeliocentricXYZJ2000 返回日心 J2000 平黄道直角坐标，单位 AU。
func OrbitHeliocentricXYZJ2000(jde float64, elements OrbitElements) Vector3 {
	trueAnomaly, radius, resolved, ok := orbitTrueAnomalyAndRadius(jde, elements)
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
func OrbitHeliocentricEclipticJ2000(jde float64, elements OrbitElements) (lon, lat, distance float64) {
	return orbitVectorToEcliptic(OrbitHeliocentricXYZJ2000(jde, elements))
}

// OrbitHeliocentricXYZ 返回日心历元黄道直角坐标，单位 AU。
func OrbitHeliocentricXYZ(jde float64, elements OrbitElements) Vector3 {
	return eclipticVectorAtReferenceEpoch(OrbitHeliocentricXYZJ2000(jde, elements), orbitReferenceJD, jde)
}

// OrbitHeliocentricEcliptic 返回日心历元黄道球坐标，单位度/AU。
func OrbitHeliocentricEcliptic(jde float64, elements OrbitElements) (lon, lat, distance float64) {
	return orbitVectorToEcliptic(OrbitHeliocentricXYZ(jde, elements))
}

// OrbitGeocentricXYZJ2000 返回地心 J2000 平黄道直角坐标，单位 AU。
func OrbitGeocentricXYZJ2000(jde float64, elements OrbitElements) Vector3 {
	objectVector := OrbitHeliocentricXYZJ2000(jde, elements)
	earthVector := earthHeliocentricVectorJ2000(jde)
	return Vector3{
		objectVector[0] - earthVector[0],
		objectVector[1] - earthVector[1],
		objectVector[2] - earthVector[2],
	}
}

// OrbitGeocentricEclipticJ2000 返回地心 J2000 平黄道球坐标，单位度/AU。
func OrbitGeocentricEclipticJ2000(jde float64, elements OrbitElements) (lon, lat, distance float64) {
	return orbitVectorToEcliptic(OrbitGeocentricXYZJ2000(jde, elements))
}

// OrbitGeocentricXYZ 返回地心历元黄道直角坐标，单位 AU。
func OrbitGeocentricXYZ(jde float64, elements OrbitElements) Vector3 {
	objectVector := OrbitHeliocentricXYZ(jde, elements)
	earthVector := earthHeliocentricVectorOfDate(jde)
	return Vector3{
		objectVector[0] - earthVector[0],
		objectVector[1] - earthVector[1],
		objectVector[2] - earthVector[2],
	}
}

// OrbitGeocentricEcliptic 返回地心历元黄道球坐标，单位度/AU。
func OrbitGeocentricEcliptic(jde float64, elements OrbitElements) (lon, lat, distance float64) {
	return orbitVectorToEcliptic(OrbitGeocentricXYZ(jde, elements))
}

// OrbitGeocentricEquatorialJ2000 返回地心 J2000 平赤道球坐标，单位度/AU。
func OrbitGeocentricEquatorialJ2000(jde float64, elements OrbitElements) (ra, dec, distance float64) {
	vector := rotateEclipticToEquatorial(OrbitGeocentricXYZJ2000(jde, elements), orbitJ2000Obliquity)
	return orbitVectorToEquatorial(vector)
}

// OrbitGeocentricEquatorial 返回地心历元平赤道球坐标，单位度/AU。
func OrbitGeocentricEquatorial(jde float64, elements OrbitElements) (ra, dec, distance float64) {
	vector := rotateEclipticToEquatorial(OrbitGeocentricXYZ(jde, elements), EclipticObliquity(jde, false))
	return orbitVectorToEquatorial(vector)
}

// OrbitAstrometricGeocentricXYZJ2000 返回光行时修正后的地心 J2000 平黄道直角坐标，单位 AU。
func OrbitAstrometricGeocentricXYZJ2000(jde float64, elements OrbitElements) Vector3 {
	if !isFinite(jde) {
		nan := math.NaN()
		return Vector3{nan, nan, nan}
	}
	earthVector := earthHeliocentricVectorJ2000(jde)
	lightTime := 0.0
	result := Vector3{}
	for i := 0; i < 8; i++ {
		objectVector := OrbitHeliocentricXYZJ2000(jde-lightTime, elements)
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
func OrbitAstrometricGeocentricEquatorialJ2000(jde float64, elements OrbitElements) (ra, dec, distance float64) {
	vector := rotateEclipticToEquatorial(OrbitAstrometricGeocentricXYZJ2000(jde, elements), orbitJ2000Obliquity)
	return orbitVectorToEquatorial(vector)
}

// OrbitApparentGeocentricEcliptic 返回光行时、周年光行差与章动修正后的地心视黄道坐标，单位度/AU。
func OrbitApparentGeocentricEcliptic(jde float64, elements OrbitElements) (lon, lat, distance float64) {
	vectorDate := eclipticVectorAtReferenceEpoch(OrbitAstrometricGeocentricXYZJ2000(jde, elements), orbitReferenceJD, jde)
	lon, lat, distance = orbitVectorToEcliptic(vectorDate)
	if math.IsNaN(lon) {
		return math.NaN(), math.NaN(), math.NaN()
	}
	aberrationLon, aberrationLat := GXCLo(lon, lat, jde), GXCBo(lon, lat, jde)
	lon = Limit360(lon + aberrationLon/3600 + Nutation2000Bi(jde))
	return lon, lat + aberrationLat/3600, distance
}

// OrbitApparentGeocentricEquatorial 返回光行时、周年光行差与章动修正后的地心视赤道坐标，单位度/AU。
func OrbitApparentGeocentricEquatorial(jde float64, elements OrbitElements) (ra, dec, distance float64) {
	lon, lat, distance := OrbitApparentGeocentricEcliptic(jde, elements)
	if math.IsNaN(lon) {
		return math.NaN(), math.NaN(), math.NaN()
	}
	ra, dec = LoBoToRaDec(jde, lon, lat)
	return ra, dec, distance
}

// OrbitApparentTopocentricEquatorial 返回光行时、周年光行差、章动与站心修正后的视赤道坐标，单位度/AU。
func OrbitApparentTopocentricEquatorial(jde, observerLon, observerLat, observerHeight float64, elements OrbitElements) (ra, dec, distance float64) {
	geocentricRA, geocentricDec, geocentricDistance := OrbitApparentGeocentricEquatorial(jde, elements)
	if math.IsNaN(geocentricRA) {
		return math.NaN(), math.NaN(), math.NaN()
	}
	geocentricVector := orbitEquatorialVector(geocentricRA, geocentricDec, geocentricDistance)
	observerVector := orbitObserverEquatorialVectorOfDate(TT2UTC(jde), observerLon, observerLat, observerHeight)
	topocentricVector := Vector3{
		geocentricVector[0] - observerVector[0],
		geocentricVector[1] - observerVector[1],
		geocentricVector[2] - observerVector[2],
	}
	return orbitVectorToEquatorial(topocentricVector)
}

func earthHeliocentricVectorOfDate(jde float64, instances ...*Instance) Vector3 {
	inst := instanceValue(instances)

	return eclipticCartesian(
		PlanetCoordinateWithInstance(-1, 0, jde, inst),
		PlanetCoordinateWithInstance(-1, 1, jde, inst),
		PlanetCoordinateWithInstance(-1, 2, jde, inst),
	)
}

func earthHeliocentricVectorJ2000(jde float64, instances ...*Instance) Vector3 {
	inst := instanceValue(instances)

	return eclipticVectorAtReferenceEpoch(earthHeliocentricVectorOfDate(jde, inst), jde, orbitReferenceJD)
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
	localApparentSiderealLongitude := Limit360(ApparentSiderealTime(UTC2UT1(jdUT))*15 + observerLon)
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
