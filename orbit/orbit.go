package orbit

import (
	"errors"
	"time"

	"github.com/starainrt/astro/basic"
)

var (
	ERR_ORBIT_NEVER_RISE = errors.New("ERROR:轨道目标今日永远在地平线下！")
	ERR_ORBIT_NEVER_SET  = errors.New("ERROR:轨道目标今日永远在地平线上！")
)

// Elements 日心二体圆锥曲线根数，参考系为 J2000 平黄道/平春分点。
// EpochJD 与 TpJD 使用 TT/TDB 对应的儒略日。
//
// 经典椭圆根数：A/E/I/Omega/W/M0
// 近日点形式：Q/E/I/Omega/W/TpJD
//
// 线性 rates 仅作用于经典椭圆根数，单位均为每天变化量。
type Elements struct {
	EpochJD float64 // 历元儒略日（TT/TDB） / epoch Julian day in TT/TDB.
	A       float64 // 半长径，单位 AU / semi-major axis in AU.
	E       float64 // 离心率 / eccentricity.
	I       float64 // 轨道倾角，单位度 / inclination in degrees.
	Omega   float64 // 升交点黄经，单位度 / longitude of ascending node in degrees.
	W       float64 // 近日点幅角，单位度 / argument of perihelion in degrees.
	M0      float64 // 历元平近点角，单位度 / mean anomaly at epoch in degrees.
	Q       float64 // 近日点距离，单位 AU / perihelion distance in AU.
	TpJD    float64 // 近日点通过时刻（TT/TDB JD） / perihelion passage time in TT/TDB Julian day.

	ADot     float64 // 半长径日变化，单位 AU/day / daily rate of A.
	EDot     float64 // 离心率日变化，单位 1/day / daily rate of E.
	IDot     float64 // 倾角日变化，单位 deg/day / daily rate of I.
	OmegaDot float64 // 升交点黄经日变化，单位 deg/day / daily rate of Omega.
	WDot     float64 // 近日点幅角日变化，单位 deg/day / daily rate of W.
	MDot     float64 // 平近点角日变化，单位 deg/day / daily rate of M.
}

// EclipticPosition 黄道球坐标结果，Lon/Lat 单位度，Distance 单位 AU。
type EclipticPosition struct {
	Lon      float64
	Lat      float64
	Distance float64
}

// EquatorialPosition 赤道球坐标结果，RA/Dec 单位度，Distance 单位 AU。
type EquatorialPosition struct {
	RA       float64
	Dec      float64
	Distance float64
}

// MeanMotion 平均角速度 / mean motion.
//
// 返回平均角速度，单位度/日；对抛物线和双曲线轨道返回 `NaN`。
func MeanMotion(elements Elements) float64 {
	return basic.OrbitMeanMotion(toBasicElements(elements))
}

// MeanAnomaly 平近点角 / mean anomaly.
//
// 返回给定时刻的平近点角，单位度；对抛物线和双曲线轨道返回 `NaN`。
func MeanAnomaly(date time.Time, elements Elements) float64 {
	return basic.OrbitMeanAnomaly(ttJulianDay(date), toBasicElements(elements))
}

// TrueAnomaly 真近点角 / true anomaly.
//
// 返回给定时刻的真近点角，单位度。
func TrueAnomaly(date time.Time, elements Elements) float64 {
	return basic.OrbitTrueAnomaly(ttJulianDay(date), toBasicElements(elements))
}

// HeliocentricEclipticJ2000 日心 J2000 平黄道坐标 / heliocentric J2000 ecliptic coordinates.
//
// 返回黄经、黄纬和距离；角度单位度，距离单位 AU。
func HeliocentricEclipticJ2000(date time.Time, elements Elements) EclipticPosition {
	lon, lat, distance := basic.OrbitHeliocentricEclipticJ2000(ttJulianDay(date), toBasicElements(elements))
	return EclipticPosition{Lon: lon, Lat: lat, Distance: distance}
}

// HeliocentricEcliptic 日心历元黄道坐标 / heliocentric ecliptic coordinates of date.
//
// 返回历元黄经、黄纬和距离；角度单位度，距离单位 AU。
func HeliocentricEcliptic(date time.Time, elements Elements) EclipticPosition {
	lon, lat, distance := basic.OrbitHeliocentricEcliptic(ttJulianDay(date), toBasicElements(elements))
	return EclipticPosition{Lon: lon, Lat: lat, Distance: distance}
}

// GeocentricEclipticJ2000 地心 J2000 平黄道坐标 / geocentric J2000 ecliptic coordinates.
//
// 返回黄经、黄纬和距离；角度单位度，距离单位 AU。
func GeocentricEclipticJ2000(date time.Time, elements Elements) EclipticPosition {
	lon, lat, distance := basic.OrbitGeocentricEclipticJ2000(ttJulianDay(date), toBasicElements(elements))
	return EclipticPosition{Lon: lon, Lat: lat, Distance: distance}
}

// GeocentricEcliptic 地心历元黄道坐标 / geocentric ecliptic coordinates of date.
//
// 返回历元黄经、黄纬和距离；角度单位度，距离单位 AU。
func GeocentricEcliptic(date time.Time, elements Elements) EclipticPosition {
	lon, lat, distance := basic.OrbitGeocentricEcliptic(ttJulianDay(date), toBasicElements(elements))
	return EclipticPosition{Lon: lon, Lat: lat, Distance: distance}
}

// GeocentricEquatorialJ2000 地心 J2000 平赤道坐标 / geocentric J2000 equatorial coordinates.
//
// 返回赤经、赤纬和距离；角度单位度，距离单位 AU。
func GeocentricEquatorialJ2000(date time.Time, elements Elements) EquatorialPosition {
	ra, dec, distance := basic.OrbitGeocentricEquatorialJ2000(ttJulianDay(date), toBasicElements(elements))
	return EquatorialPosition{RA: ra, Dec: dec, Distance: distance}
}

// GeocentricEquatorial 地心历元平赤道坐标 / geocentric equatorial coordinates of date.
//
// 返回历元赤经、赤纬和距离；角度单位度，距离单位 AU。
func GeocentricEquatorial(date time.Time, elements Elements) EquatorialPosition {
	ra, dec, distance := basic.OrbitGeocentricEquatorial(ttJulianDay(date), toBasicElements(elements))
	return EquatorialPosition{RA: ra, Dec: dec, Distance: distance}
}

// AstrometricGeocentricEquatorialJ2000 地心测算 J2000 赤道坐标 / astrometric geocentric J2000 equatorial coordinates.
//
// 返回加入光行时修正后的地心 J2000 赤经、赤纬和距离；角度单位度，距离单位 AU。
func AstrometricGeocentricEquatorialJ2000(date time.Time, elements Elements) EquatorialPosition {
	ra, dec, distance := basic.OrbitAstrometricGeocentricEquatorialJ2000(ttJulianDay(date), toBasicElements(elements))
	return EquatorialPosition{RA: ra, Dec: dec, Distance: distance}
}

// ApparentGeocentricEcliptic 地心视黄道坐标 / apparent geocentric ecliptic coordinates.
//
// 返回加入光行时与章动修正后的地心视黄经、黄纬和距离；角度单位度，距离单位 AU。
func ApparentGeocentricEcliptic(date time.Time, elements Elements) EclipticPosition {
	lon, lat, distance := basic.OrbitApparentGeocentricEcliptic(ttJulianDay(date), toBasicElements(elements))
	return EclipticPosition{Lon: lon, Lat: lat, Distance: distance}
}

// ApparentGeocentricEquatorial 地心视赤道坐标 / apparent geocentric equatorial coordinates.
//
// 返回加入光行时与章动修正后的地心视赤经、赤纬和距离；角度单位度，距离单位 AU。
func ApparentGeocentricEquatorial(date time.Time, elements Elements) EquatorialPosition {
	ra, dec, distance := basic.OrbitApparentGeocentricEquatorial(ttJulianDay(date), toBasicElements(elements))
	return EquatorialPosition{RA: ra, Dec: dec, Distance: distance}
}

// ApparentTopocentricEquatorial 站心视赤道坐标 / apparent topocentric equatorial coordinates.
//
// 返回加入光行时、章动和站心修正后的视赤经、赤纬和距离；
// `observerLon` 东经为正，`observerLat` 北纬为正，`observerHeight` 单位米。
func ApparentTopocentricEquatorial(date time.Time, elements Elements, observerLon, observerLat, observerHeight float64) EquatorialPosition {
	ra, dec, distance := basic.OrbitApparentTopocentricEquatorial(ttJulianDay(date), observerLon, observerLat, observerHeight, toBasicElements(elements))
	return EquatorialPosition{RA: ra, Dec: dec, Distance: distance}
}

// Altitude 视高度角 / apparent altitude.
//
// 返回目标在观测者所在地的视高度角，单位度；经度东正西负，纬度北正南负，海拔单位米。
func Altitude(date time.Time, elements Elements, observerLon, observerLat, observerHeight float64) float64 {
	jde := basic.Date2JDE(date)
	return basic.OrbitHeight(jde, observerLon, observerLat, observationTimezone(date), observerHeight, toBasicElements(elements))
}

// Zenith 天顶距 / zenith distance.
//
// 返回目标在观测者所在地的天顶距，单位度。
func Zenith(date time.Time, elements Elements, observerLon, observerLat, observerHeight float64) float64 {
	return 90 - Altitude(date, elements, observerLon, observerLat, observerHeight)
}

// Azimuth 视方位角 / apparent azimuth.
//
// 返回目标在观测者所在地的视方位角，按正北为 0°、向东增加。
func Azimuth(date time.Time, elements Elements, observerLon, observerLat, observerHeight float64) float64 {
	jde := basic.Date2JDE(date)
	return basic.OrbitAzimuth(jde, observerLon, observerLat, observationTimezone(date), observerHeight, toBasicElements(elements))
}

// HourAngle 站心视时角 / topocentric hour angle.
//
// 返回目标在观测者所在地的站心视时角，单位度。
func HourAngle(date time.Time, elements Elements, observerLon, observerLat, observerHeight float64) float64 {
	jde := basic.Date2JDE(date)
	return basic.OrbitHourAngle(jde, observerLon, observerLat, observationTimezone(date), observerHeight, toBasicElements(elements))
}

// CulminationTime 中天时刻 / culmination time.
//
// 返回目标在给定当地日期内的中天时刻，结果保持输入 `date` 的时区。
func CulminationTime(date time.Time, elements Elements, observerLon, observerLat, observerHeight float64) time.Time {
	if date.Hour() > 12 {
		date = date.Add(-12 * time.Hour)
	}
	timezone := observationTimezone(date)
	jde := basic.Date2JDE(date)
	calcJde := basic.OrbitCulminationTime(jde, observerLon, observerLat, timezone, observerHeight, toBasicElements(elements)) - timezone/24.0
	return basic.JDE2DateByZone(calcJde, date.Location(), false)
}

// RiseTime 升起时刻 / rise time.
//
// 返回目标在给定当地日期内的升起时刻；`aero=true` 时加入标准大气折射修正。
func RiseTime(date time.Time, elements Elements, observerLon, observerLat, observerHeight float64, aero bool) (time.Time, error) {
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	if date.Hour() > 12 {
		date = date.Add(-12 * time.Hour)
	}
	timezone := observationTimezone(date)
	jde := basic.Date2JDE(date)
	calcJde, err := basic.OrbitRiseTime(jde, observerLon, observerLat, timezone, aeroFloat, observerHeight, toBasicElements(elements))
	return orbitRiseSetResult(date, calcJde, err)
}

// SetTime 落下时刻 / set time.
//
// 返回目标在给定当地日期内的落下时刻；`aero=true` 时加入标准大气折射修正。
func SetTime(date time.Time, elements Elements, observerLon, observerLat, observerHeight float64, aero bool) (time.Time, error) {
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	if date.Hour() > 12 {
		date = date.Add(-12 * time.Hour)
	}
	timezone := observationTimezone(date)
	jde := basic.Date2JDE(date)
	calcJde, err := basic.OrbitSetTime(jde, observerLon, observerLat, timezone, aeroFloat, observerHeight, toBasicElements(elements))
	return orbitRiseSetResult(date, calcJde, err)
}

func orbitRiseSetResult(date time.Time, jde float64, err error) (time.Time, error) {
	if err != nil {
		switch {
		case errors.Is(err, basic.ErrNeverRise):
			return time.Time{}, ERR_ORBIT_NEVER_RISE
		case errors.Is(err, basic.ErrNeverSet):
			return time.Time{}, ERR_ORBIT_NEVER_SET
		default:
			return time.Time{}, err
		}
	}
	return basic.JDE2DateByZone(jde, date.Location(), true), nil
}

func observationTimezone(date time.Time) float64 {
	_, loc := date.Zone()
	return float64(loc) / 3600.0
}

func ttJulianDay(date time.Time) float64 {
	jdeUTC := basic.Date2JDE(date.UTC())
	return basic.TD2UT(jdeUTC, true)
}

func toBasicElements(elements Elements) basic.OrbitElements {
	return basic.OrbitElements{
		EpochJD:  elements.EpochJD,
		A:        elements.A,
		E:        elements.E,
		I:        elements.I,
		Omega:    elements.Omega,
		W:        elements.W,
		M0:       elements.M0,
		Q:        elements.Q,
		TpJD:     elements.TpJD,
		ADot:     elements.ADot,
		EDot:     elements.EDot,
		IDot:     elements.IDot,
		OmegaDot: elements.OmegaDot,
		WDot:     elements.WDot,
		MDot:     elements.MDot,
	}
}
