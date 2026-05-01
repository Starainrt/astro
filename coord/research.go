package coord

import "math"

// Galactic 银道坐标 / galactic coordinates.
type Galactic struct {
	Lon float64 // 银经，单位度 / galactic longitude in degrees.
	Lat float64 // 银纬，单位度 / galactic latitude in degrees.
}

// SOFA ICRS2G/G2ICRS fixed rotation matrix.
// Source: IAU SOFA 2023-10-11, iauIcrs2g/iauG2icrs.
var icrsToGalacticMatrix = [3][3]float64{
	{-0.054875560416215368492398900454, -0.873437090234885048760383168409, -0.483835015548713226831774175116},
	{+0.494109427875583673525222371358, -0.444829629960011178146614061616, +0.746982244497218890527388004556},
	{-0.867666149019004701181616534570, -0.198076373431201528180486091412, +0.455983776175066922272100478348},
}

// EclipticToEquatorialByObliquity 黄道转赤道坐标（指定黄赤交角） / ecliptic to equatorial by obliquity.
//
//	lon: 黄经，单位度
//	lat: 黄纬，单位度
//	obliquity: 黄赤交角，单位度
//
// 返回：
//
//	赤经 RA，单位度；赤纬 Dec，单位度
func EclipticToEquatorialByObliquity(lon, lat, obliquity float64) Equatorial {
	sinLon, cosLon := sinCosDeg(lon)
	sinLat, cosLat := sinCosDeg(lat)
	sinObliquity, cosObliquity := sinCosDeg(obliquity)

	ra := normalize360(math.Atan2(sinLon*cosObliquity-math.Tan(lat*math.Pi/180)*sinObliquity, cosLon) * 180 / math.Pi)
	dec := math.Asin(clampUnit(sinLat*cosObliquity+cosLat*sinObliquity*sinLon)) * 180 / math.Pi
	return Equatorial{RA: ra, Dec: dec}
}

// EquatorialToEclipticByObliquity 赤道转黄道坐标（指定黄赤交角） / equatorial to ecliptic by obliquity.
//
//	ra: 赤经，单位度
//	dec: 赤纬，单位度
//	obliquity: 黄赤交角，单位度
//
// 返回：
//
//	黄经 Lon，单位度；黄纬 Lat，单位度
func EquatorialToEclipticByObliquity(ra, dec, obliquity float64) Ecliptic {
	sinRA, cosRA := sinCosDeg(ra)
	sinDec, cosDec := sinCosDeg(dec)
	sinObliquity, cosObliquity := sinCosDeg(obliquity)

	lon := normalize360(math.Atan2(sinRA*cosObliquity+math.Tan(dec*math.Pi/180)*sinObliquity, cosRA) * 180 / math.Pi)
	lat := math.Asin(clampUnit(sinDec*cosObliquity-cosDec*sinObliquity*sinRA)) * 180 / math.Pi
	return Ecliptic{Lon: lon, Lat: lat}
}

// HourAngleDeclinationToHorizontal 时角赤纬转地平坐标 / horizontal coordinates from hour angle and declination.
//
//	hourAngle: 时角，单位度
//	declination: 赤纬，单位度
//	latitude: 观测者地理纬度，单位度，北正南负
//
// 返回：
//
//	方位角 Azimuth（正北为0，顺时针增加）、高度角 Altitude、天顶距 Zenith，均为度
func HourAngleDeclinationToHorizontal(hourAngle, declination, latitude float64) Horizontal {
	sinLatitude, cosLatitude := sinCosDeg(latitude)
	sinDeclination, cosDeclination := sinCosDeg(declination)
	sinHourAngle, cosHourAngle := sinCosDeg(hourAngle)

	altitude := math.Asin(clampUnit(sinLatitude*sinDeclination+cosLatitude*cosDeclination*cosHourAngle)) * 180 / math.Pi
	azimuth := normalize360(math.Atan2(-cosDeclination*sinHourAngle, cosLatitude*sinDeclination-sinLatitude*cosDeclination*cosHourAngle) * 180 / math.Pi)
	return Horizontal{
		Azimuth:   azimuth,
		Altitude:  altitude,
		Zenith:    90 - altitude,
		HourAngle: normalize360(hourAngle),
	}
}

// HorizontalToHourAngleDeclination 地平坐标转时角赤纬 / hour angle and declination from horizontal coordinates.
//
//	azimuth: 方位角，单位度，正北为0，顺时针增加
//	altitude: 高度角，单位度
//	latitude: 观测者地理纬度，单位度，北正南负
//
// 返回：
//
//	时角 HourAngle，单位度；赤纬 Declination，单位度
func HorizontalToHourAngleDeclination(azimuth, altitude, latitude float64) (hourAngle, declination float64) {
	sinLatitude, cosLatitude := sinCosDeg(latitude)
	sinAltitude, cosAltitude := sinCosDeg(altitude)
	sinAzimuth, cosAzimuth := sinCosDeg(azimuth)

	declination = math.Asin(clampUnit(sinLatitude*sinAltitude+cosLatitude*cosAltitude*cosAzimuth)) * 180 / math.Pi
	sinHourAngle := -cosAltitude * sinAzimuth
	cosHourAngle := sinAltitude*cosLatitude - cosAltitude*sinLatitude*cosAzimuth
	hourAngle = normalize360(math.Atan2(sinHourAngle, cosHourAngle) * 180 / math.Pi)
	return hourAngle, declination
}

// EquatorialToHorizontalByLocalSiderealTime 赤道转地平坐标（指定地方恒星时） / equatorial to horizontal by local sidereal time.
//
//	localSiderealTimeHours: 站点本地恒星时，单位小时
//	ra: 赤经，单位度
//	dec: 赤纬，单位度
//	latitude: 观测者地理纬度，单位度，北正南负
//
// 返回：
//
//	方位角 Azimuth（正北为0，顺时针增加）、高度角 Altitude、天顶距 Zenith，均为度；
//	附带返回对应的时角 HourAngle，单位度
//
// 例：
//
//	hz := coord.EquatorialToHorizontalByLocalSiderealTime(10.5, 83.6331, 22.0145, 31.2)
func EquatorialToHorizontalByLocalSiderealTime(localSiderealTimeHours, ra, dec, latitude float64) Horizontal {
	hourAngle := normalize360(localSiderealTimeHours*15 - ra)
	return HourAngleDeclinationToHorizontal(hourAngle, dec, latitude)
}

// HorizontalToEquatorialByLocalSiderealTime 地平转赤道坐标（指定地方恒星时） / horizontal to equatorial by local sidereal time.
//
//	localSiderealTimeHours: 站点本地恒星时，单位小时
//	azimuth: 方位角，单位度，正北为0，顺时针增加
//	altitude: 高度角，单位度
//	latitude: 观测者地理纬度，单位度，北正南负
//
// 返回：
//
//	赤经 RA，单位度；赤纬 Dec，单位度
//
// 例：
//
//	eq := coord.HorizontalToEquatorialByLocalSiderealTime(10.5, 128.2, 37.6, 31.2)
func HorizontalToEquatorialByLocalSiderealTime(localSiderealTimeHours, azimuth, altitude, latitude float64) Equatorial {
	hourAngle, declination := HorizontalToHourAngleDeclination(azimuth, altitude, latitude)
	ra := normalize360(localSiderealTimeHours*15 - hourAngle)
	return Equatorial{RA: ra, Dec: declination}
}

// EquatorialToGalactic 赤道转银道坐标 / equatorial to galactic coordinates.
//
//	ra: ICRS 赤经，单位度
//	dec: ICRS 赤纬，单位度
//
// 返回：
//
//	银经 Lon，单位度；银纬 Lat，单位度
func EquatorialToGalactic(ra, dec float64) Galactic {
	vector := sphericalToVector(ra, dec)
	rotated := matrixVectorMul(icrsToGalacticMatrix, vector)
	lon, lat := vectorToSpherical(rotated)
	return Galactic{Lon: lon, Lat: lat}
}

// GalacticToEquatorial 银道转赤道坐标 / galactic to equatorial coordinates.
//
//	lon: 银经，单位度
//	lat: 银纬，单位度
//
// 返回：
//
//	ICRS 赤经 RA，单位度；ICRS 赤纬 Dec，单位度
func GalacticToEquatorial(lon, lat float64) Equatorial {
	vector := sphericalToVector(lon, lat)
	rotated := matrixTransposeVectorMul(icrsToGalacticMatrix, vector)
	ra, dec := vectorToSpherical(rotated)
	return Equatorial{RA: ra, Dec: dec}
}

func sinCosDeg(angle float64) (sinValue, cosValue float64) {
	return math.Sincos(angle * math.Pi / 180)
}

func normalize360(angle float64) float64 {
	angle = math.Mod(angle, 360)
	if angle < 0 {
		angle += 360
	}
	return angle
}

func clampUnit(value float64) float64 {
	if value > 1 {
		return 1
	}
	if value < -1 {
		return -1
	}
	return value
}

func sphericalToVector(lon, lat float64) [3]float64 {
	sinLon, cosLon := sinCosDeg(lon)
	sinLat, cosLat := sinCosDeg(lat)
	return [3]float64{cosLat * cosLon, cosLat * sinLon, sinLat}
}

func vectorToSpherical(vector [3]float64) (lon, lat float64) {
	lon = normalize360(math.Atan2(vector[1], vector[0]) * 180 / math.Pi)
	lat = math.Asin(clampUnit(vector[2])) * 180 / math.Pi
	return lon, lat
}

func matrixVectorMul(matrix [3][3]float64, vector [3]float64) [3]float64 {
	return [3]float64{
		matrix[0][0]*vector[0] + matrix[0][1]*vector[1] + matrix[0][2]*vector[2],
		matrix[1][0]*vector[0] + matrix[1][1]*vector[1] + matrix[1][2]*vector[2],
		matrix[2][0]*vector[0] + matrix[2][1]*vector[1] + matrix[2][2]*vector[2],
	}
}

func matrixTransposeVectorMul(matrix [3][3]float64, vector [3]float64) [3]float64 {
	return [3]float64{
		matrix[0][0]*vector[0] + matrix[1][0]*vector[1] + matrix[2][0]*vector[2],
		matrix[0][1]*vector[0] + matrix[1][1]*vector[1] + matrix[2][1]*vector[2],
		matrix[0][2]*vector[0] + matrix[1][2]*vector[1] + matrix[2][2]*vector[2],
	}
}
