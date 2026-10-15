package geodata

import "math"

// 地球尺度常量 / Earth size constants.
const (
	// EarthEquatorialRadiusKM 地球赤道半径 / Earth equatorial radius.
	EarthEquatorialRadiusKM = 6378.1366
	// EarthMeanRadiusKM 地球平均半径 / Earth mean radius.
	EarthMeanRadiusKM = 6371.0088
	// EarthWGS84RadiusKM WGS84 椭球长半轴 / WGS84 semi-major axis.
	EarthWGS84RadiusKM = 6378.137
	// EarthKMPerDegree 大圆每度的近似长度 / approximate great-circle length per degree.
	EarthKMPerDegree = 111.195
	// WebMercatorMaxLatitude Web Mercator 可投影的纬度上限 / Web Mercator latitude limit.
	WebMercatorMaxLatitude = 85.05112878
)

// NormalizeLongitude 把经度归一化到 [-180, 180) / wraps a longitude into [-180, 180).
func NormalizeLongitude(value float64) float64 {
	value = math.Mod(value+180, 360)
	if value < 0 {
		value += 360
	}
	return value - 180
}

// CentralAngle 两点间的球心角（弧度），参数为度 / haversine central angle in radians.
func CentralAngle(longitude1, latitude1, longitude2, latitude2 float64) float64 {
	firstLatitude := latitude1 * math.Pi / 180
	secondLatitude := latitude2 * math.Pi / 180
	deltaLatitude := secondLatitude - firstLatitude
	deltaLongitude := math.Remainder((longitude2-longitude1)*math.Pi/180, 2*math.Pi)
	haversine := math.Sin(deltaLatitude/2)*math.Sin(deltaLatitude/2) +
		math.Cos(firstLatitude)*math.Cos(secondLatitude)*math.Sin(deltaLongitude/2)*math.Sin(deltaLongitude/2)
	return 2 * math.Asin(math.Sqrt(math.Max(0, math.Min(1, haversine))))
}

// EquatorialDistanceKM 以赤道半径计的大圆距离 / great-circle distance on the equatorial-radius sphere.
func EquatorialDistanceKM(first, second GeoPoint) float64 {
	return EarthEquatorialRadiusKM * CentralAngle(first.Longitude, first.Latitude, second.Longitude, second.Latitude)
}

// MeanDistanceKM 以平均半径计的大圆距离 / great-circle distance on the mean-radius sphere.
func MeanDistanceKM(first, second GeoPoint) float64 {
	return EarthMeanRadiusKM * CentralAngle(first.Longitude, first.Latitude, second.Longitude, second.Latitude)
}

// WebMercatorDistanceKM 两点在 Web Mercator 平面上的距离，经差取短弧 / Web Mercator planar distance.
func WebMercatorDistanceKM(first, second GeoPoint) float64 {
	clampLatitude := func(value float64) float64 {
		return math.Max(-WebMercatorMaxLatitude, math.Min(WebMercatorMaxLatitude, value))
	}
	longitude := math.Remainder(second.Longitude-first.Longitude, 360) * math.Pi / 180
	firstLatitude := clampLatitude(first.Latitude) * math.Pi / 180
	secondLatitude := clampLatitude(second.Latitude) * math.Pi / 180
	firstY := math.Log(math.Tan(math.Pi/4 + firstLatitude/2))
	secondY := math.Log(math.Tan(math.Pi/4 + secondLatitude/2))
	return EarthEquatorialRadiusKM * math.Hypot(longitude, secondY-firstY)
}
