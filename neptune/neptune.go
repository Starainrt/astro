package neptune

import (
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
	"github.com/starainrt/astro/planet"
	"errors"
	"time"
)

var (
	ERR_NEPTUNE_NEVER_RISE = errors.New("ERROR:极夜，海王星今日永远在地平线下！")
	ERR_NEPTUNE_NEVER_SET  = errors.New("ERROR:极昼，海王星今日永远在地平线上！")
	// ERR_NEPTUNE_NEVER_DOWN deprecated: -- use ERR_NEPTUNE_NEVER_SET instead
	ERR_NEPTUNE_NEVER_DOWN = ERR_NEPTUNE_NEVER_SET
)

func riseSetResult(date time.Time, jde float64, err error) (time.Time, error) {
	if err != nil {
		switch {
		case errors.Is(err, basic.ErrNeverRise):
			return time.Time{}, ERR_NEPTUNE_NEVER_RISE
		case errors.Is(err, basic.ErrNeverSet):
			return time.Time{}, ERR_NEPTUNE_NEVER_SET
		default:
			return time.Time{}, err
		}
	}
	return basic.JDE2DateByZone(jde, date.Location(), true), nil
}

// ApparentLo 视黄经 / apparent ecliptic longitude.
//
// 返回海王星在 date 对应绝对时刻的瞬时视黄经，单位度。
// Returns the apparent ecliptic longitude of Neptune at the instant represented by date, in degrees.
func ApparentLo(date time.Time) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.NeptuneApparentLo(basic.TD2UT(jde, true))
}

// ApparentBo 视黄纬 / apparent ecliptic latitude.
//
// 返回海王星在 date 对应绝对时刻的瞬时视黄纬，单位度。
// Returns the apparent ecliptic latitude of Neptune at the instant represented by date, in degrees.
func ApparentBo(date time.Time) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.NeptuneApparentBo(basic.TD2UT(jde, true))
}

// ApparentRa 视赤经 / apparent right ascension.
//
// 返回海王星在 date 对应绝对时刻的瞬时视赤经，单位度。
// Returns the apparent right ascension of Neptune at the instant represented by date, in degrees.
func ApparentRa(date time.Time) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.NeptuneApparentRa(basic.TD2UT(jde, true))
}

// ApparentDec 视赤纬 / apparent declination.
//
// 返回海王星在 date 对应绝对时刻的瞬时视赤纬，单位度。
// Returns the apparent declination of Neptune at the instant represented by date, in degrees.
func ApparentDec(date time.Time) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.NeptuneApparentDec(basic.TD2UT(jde, true))
}

// ApparentRaDec 视赤经、视赤纬 / apparent right ascension and declination.
//
// 返回海王星在 date 对应绝对时刻的瞬时视赤经与视赤纬，单位度。
// Returns the apparent right ascension and declination of Neptune at the instant represented by date, in degrees.
func ApparentRaDec(date time.Time) (float64, float64) {
	jde := calendar.Date2JDE(date.UTC())
	return basic.NeptuneApparentRaDec(basic.TD2UT(jde, true))
}

// ApparentMagnitude 视星等 / apparent magnitude.
//
// 返回海王星在 date 对应绝对时刻的视星等。
// Returns the apparent visual magnitude of Neptune at the instant represented by date.
func ApparentMagnitude(date time.Time) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.NeptuneMag(basic.TD2UT(jde, true))
}

// EarthDistance 地心距离 / Earth distance.
//
// 返回海王星在 date 对应绝对时刻到地球的距离，单位 AU。
// Returns the distance from Neptune to Earth at the instant represented by date, in astronomical units.
func EarthDistance(date time.Time) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.EarthNeptuneAway(basic.TD2UT(jde, true))
}

// SunDistance 日心距离 / Sun distance.
//
// 返回海王星在 date 对应绝对时刻到太阳的距离，单位 AU。
// Returns the distance from Neptune to the Sun at the instant represented by date, in astronomical units.
func SunDistance(date time.Time) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return planet.WherePlanet(7, 2, basic.TD2UT(jde, true))
}

// Altitude 高度角 / altitude.
//
// date 表示观测时刻，会读取其时区参与地方时计算；lon 为观测者经度，东正西负；lat 为观测者纬度，北正南负。返回值单位度。
// date is the observing instant and its zone offset participates in local-time calculations. lon is east-positive longitude, lat is north-positive latitude, and the result is in degrees.
func Altitude(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.NeptuneHeight(jde, lon, lat, timezone)
}

// Zenith 天顶距 / zenith distance.
//
// 参数与 Altitude 相同，返回值为对应时刻的天顶距，单位度。
// Uses the same inputs as Altitude and returns the zenith distance in degrees.
func Zenith(date time.Time, lon, lat float64) float64 {
	return 90 - Altitude(date, lon, lat)
}

// Azimuth 方位角 / azimuth.
//
// date 表示观测时刻，会读取其时区参与地方时计算；lon 为观测者经度，东正西负；lat 为观测者纬度，北正南负。返回值按正北为 0°、向东增加。
// date is the observing instant and its zone offset participates in local-time calculations. lon is east-positive longitude, lat is north-positive latitude, and azimuth is measured from north toward east.
func Azimuth(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.NeptuneAzimuth(jde, lon, lat, timezone)
}

// HourAngle 时角 / hour angle.
//
// date 表示观测时刻，会读取其时区参与地方时计算；lon 为观测者经度，东正西负。返回值单位度。
// date is the observing instant and its zone offset participates in local-time calculations. lon is east-positive longitude and the returned hour angle is in degrees.
func HourAngle(date time.Time, lon float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.NeptuneHourAngle(jde, lon, timezone)
}

// CulminationTime 中天时刻 / culmination time.
//
// date 取其所在时区的当地日期，返回值保持相同时区；lon 为观测者经度，东正西负。
// date is interpreted on its local civil day and the result keeps the same time zone. lon is east-positive longitude.
func CulminationTime(date time.Time, lon float64) time.Time {
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJde := basic.NeptuneCulminationTime(jde, lon, timezone) - timezone/24.00
	return basic.JDE2DateByZone(calcJde, date.Location(), false)
}

// RiseTime 升起时间 / rise time.
//
// date 取其所在时区的当地日期，返回值保持相同时区；lon 为东正西负经度，lat 为北正南负纬度；height 为观测点海拔高度（米）；aero 为 true 时加入标准大气折射。
// date is interpreted on its local civil day and the result keeps the same time zone. lon is east-positive longitude, lat is north-positive latitude, height is observer elevation in meters, and aero enables standard atmospheric refraction.
func RiseTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJde, err := basic.NeptuneRiseTime(jde, lon, lat, timezone, aeroFloat, height)
	return riseSetResult(date, riseJde, err)
}

// DownTime 落下时间别名 / deprecated set-time alias.
//
// Deprecated: use SetTime instead.
//
// 参数与 SetTime 相同，仅为兼容旧接口保留。
// Same as SetTime and kept only for backward compatibility.
func DownTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return SetTime(date, lon, lat, height, aero)
}

// SetTime 落下时间 / set time.
//
// 参数与 RiseTime 相同，返回给定当地日期内的落下时刻。
// Uses the same inputs as RiseTime and returns the set time on the corresponding local civil day.
func SetTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJde, err := basic.NeptuneSetTime(jde, lon, lat, timezone, aeroFloat, height)
	return riseSetResult(date, riseJde, err)
}

// LastConjunction 上一次合日 / previous conjunction with the Sun.
//
// 返回 date 当前或之前最近一次与太阳的合日时刻，结果保持 date 的时区。
func LastConjunction(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastNeptuneConjunction(jde), date.Location(), false)
}

// NextConjunction 下一次合日 / next conjunction with the Sun.
//
// 返回 date 当前或之后最近一次与太阳的合日时刻，结果保持 date 的时区。
func NextConjunction(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextNeptuneConjunction(jde), date.Location(), false)
}

// LastOpposition 上一次冲日 / previous opposition.
//
// 返回 date 当前或之前最近一次冲日时刻，结果保持 date 的时区。
func LastOpposition(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastNeptuneOpposition(jde), date.Location(), false)
}

// NextOpposition 下一次冲日 / next opposition.
//
// 返回 date 当前或之后最近一次冲日时刻，结果保持 date 的时区。
func NextOpposition(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextNeptuneOpposition(jde), date.Location(), false)
}

// LastProgradeToRetrograde 上一次顺行转逆行留 / previous station from prograde to retrograde.
//
// 返回 date 当前或之前最近一次由顺行转为逆行的留时刻，结果保持 date 的时区。
func LastProgradeToRetrograde(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastNeptuneProgradeToRetrograde(jde), date.Location(), false)
}

// NextProgradeToRetrograde 下一次顺行转逆行留 / next station from prograde to retrograde.
//
// 返回 date 当前或之后最近一次由顺行转为逆行的留时刻，结果保持 date 的时区。
func NextProgradeToRetrograde(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextNeptuneProgradeToRetrograde(jde), date.Location(), false)
}

// LastRetrogradeToPrograde 上一次逆行转顺行留 / previous station from retrograde to prograde.
//
// 返回 date 当前或之前最近一次由逆行转为顺行的留时刻，结果保持 date 的时区。
func LastRetrogradeToPrograde(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastNeptuneRetrogradeToPrograde(jde), date.Location(), false)
}

// NextRetrogradeToPrograde 下一次逆行转顺行留 / next station from retrograde to prograde.
//
// 返回 date 当前或之后最近一次由逆行转为顺行的留时刻，结果保持 date 的时区。
func NextRetrogradeToPrograde(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextNeptuneRetrogradeToPrograde(jde), date.Location(), false)
}

// LastEasternQuadrature 上一次东方照 / previous eastern quadrature.
//
// 返回 date 当前或之前最近一次东方照时刻，结果保持 date 的时区。
func LastEasternQuadrature(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastNeptuneEasternQuadrature(jde), date.Location(), false)
}

// NextEasternQuadrature 下一次东方照 / next eastern quadrature.
//
// 返回 date 当前或之后最近一次东方照时刻，结果保持 date 的时区。
func NextEasternQuadrature(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextNeptuneEasternQuadrature(jde), date.Location(), false)
}

// LastWesternQuadrature 上一次西方照 / previous western quadrature.
//
// 返回 date 当前或之前最近一次西方照时刻，结果保持 date 的时区。
func LastWesternQuadrature(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastNeptuneWesternQuadrature(jde), date.Location(), false)
}

// NextWesternQuadrature 下一次西方照 / next western quadrature.
//
// 返回 date 当前或之后最近一次西方照时刻，结果保持 date 的时区。
func NextWesternQuadrature(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextNeptuneWesternQuadrature(jde), date.Location(), false)
}
