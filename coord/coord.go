// Package coord 坐标工具包 / coordinate utility package.
//
// 本包面向用户提供常用坐标变换、恒星时、岁差、站心和地平坐标封装。
// 所有角度输入和输出默认使用度；恒星时输出使用小时。
// date 按绝对时刻使用，内部会转换为 UTC 后计算。
package coord

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// Ecliptic 黄道坐标 / ecliptic coordinates.
type Ecliptic struct {
	Lon float64 // 黄经，单位度 / ecliptic longitude in degrees.
	Lat float64 // 黄纬，单位度 / ecliptic latitude in degrees.
}

// Equatorial 赤道坐标 / equatorial coordinates.
type Equatorial struct {
	RA  float64 // 赤经，单位度 / right ascension in degrees.
	Dec float64 // 赤纬，单位度 / declination in degrees.
}

// Horizontal 地平坐标 / horizontal coordinates.
type Horizontal struct {
	Azimuth   float64 // 方位角，正北为0，顺时针增加 / azimuth from north clockwise, degrees.
	Altitude  float64 // 高度角，单位度 / altitude in degrees.
	Zenith    float64 // 天顶距，单位度 / zenith distance in degrees.
	HourAngle float64 // 时角，单位度 / hour angle in degrees.
}

func jdeUTC(date time.Time) float64 {
	return basic.Date2JDE(date.UTC())
}

// EclipticToEquatorial 黄道坐标转赤道坐标 / converts ecliptic to equatorial coordinates.
func EclipticToEquatorial(date time.Time, lon, lat float64) Equatorial {
	ra, dec := basic.LoBoToRaDec(jdeUTC(date), lon, lat)
	return Equatorial{RA: ra, Dec: dec}
}

// EquatorialToEcliptic 赤道坐标转黄道坐标 / converts equatorial to ecliptic coordinates.
func EquatorialToEcliptic(date time.Time, ra, dec float64) Ecliptic {
	lon, lat := basic.RaDecToLoBo(jdeUTC(date), ra, dec)
	return Ecliptic{Lon: lon, Lat: lat}
}

// Precess 岁差修正 / precesses equatorial coordinates from one date to another.
func Precess(from, to time.Time, ra, dec float64) Equatorial {
	nextRA, nextDec := basic.Precess(ra, dec, jdeUTC(from), jdeUTC(to))
	return Equatorial{RA: nextRA, Dec: nextDec}
}

// EclipticObliquity 黄赤交角 / ecliptic obliquity.
func EclipticObliquity(date time.Time, nutation bool) float64 {
	return basic.EclipticObliquity(jdeUTC(date), nutation)
}

// Nutation2000B IAU 2000B 章动 / IAU 2000B nutation.
func Nutation2000B(date time.Time) (longitude, obliquity float64) {
	return basic.Nutation2000B(jdeUTC(date))
}

// Nutation1980 IAU 1980 章动 / IAU 1980 nutation.
func Nutation1980(date time.Time) (longitude, obliquity float64) {
	return basic.Nutation1980(jdeUTC(date))
}

// MeanSiderealTime 平恒星时，单位小时 / mean sidereal time in hours.
func MeanSiderealTime(date time.Time) float64 {
	return basic.MeanSiderealTime(jdeUTC(date))
}

// ApparentSiderealTime 真恒星时，单位小时 / apparent sidereal time in hours.
func ApparentSiderealTime(date time.Time) float64 {
	return basic.ApparentSiderealTime(jdeUTC(date))
}

// HourAngle 时角 / hour angle.
//
// ra 为瞬时赤经；observerLon 为观测者经度，东正西负。
// ra is apparent right ascension; observerLon is east-positive longitude.
func HourAngle(date time.Time, ra, observerLon float64) float64 {
	return basic.StarHourAngle(jdeUTC(date), ra, observerLon, 0)
}

// EquatorialToHorizontal 赤道坐标转地平坐标 / converts equatorial to horizontal coordinates.
//
// ra/dec 为瞬时赤经赤纬；observerLon/observerLat 为观测者经纬度，东正西负、北正南负。
// ra/dec are apparent coordinates; observerLon/observerLat are east-positive and north-positive.
func EquatorialToHorizontal(date time.Time, ra, dec, observerLon, observerLat float64) Horizontal {
	jde := jdeUTC(date)
	altitude := basic.StarHeight(jde, ra, dec, observerLon, observerLat, 0)
	return Horizontal{
		Azimuth:   basic.StarAzimuth(jde, ra, dec, observerLon, observerLat, 0),
		Altitude:  altitude,
		Zenith:    90 - altitude,
		HourAngle: basic.StarHourAngle(jde, ra, observerLon, 0),
	}
}

// TopocentricEquatorial 地心赤道坐标转站心赤道坐标 / converts geocentric to topocentric equatorial coordinates.
//
// distanceAU 为目标天体到地心距离，单位 AU；height 为观测者海拔，单位米。
// distanceAU is geocentric distance in AU; height is observer elevation in meters.
func TopocentricEquatorial(date time.Time, ra, dec, observerLon, observerLat, distanceAU, height float64) Equatorial {
	topRA, topDec := basic.TopocentricRaDec(ra, dec, observerLat, observerLon, jdeUTC(date), distanceAU, height)
	return Equatorial{RA: topRA, Dec: topDec}
}

// TopocentricEcliptic 地心黄道坐标转站心黄道坐标 / converts geocentric to topocentric ecliptic coordinates.
//
// distanceAU 为目标天体到地心距离，单位 AU；height 为观测者海拔，单位米。
// distanceAU is geocentric distance in AU; height is observer elevation in meters.
func TopocentricEcliptic(date time.Time, lon, lat, observerLon, observerLat, distanceAU, height float64) Ecliptic {
	topLon := basic.TopocentricLo(lon, lat, observerLat, observerLon, jdeUTC(date), distanceAU, height)
	topLat := basic.TopocentricBo(lon, lat, observerLat, observerLon, jdeUTC(date), distanceAU, height)
	return Ecliptic{Lon: topLon, Lat: topLat}
}

// AngularSeparation 角距离 / angular separation.
//
// 输入为两组赤道坐标，单位度；返回角距离，单位度。
// Inputs are two equatorial coordinates in degrees; return value is in degrees.
func AngularSeparation(ra1, dec1, ra2, dec2 float64) float64 {
	return basic.StarAngularSeparation(ra1, dec1, ra2, dec2)
}
