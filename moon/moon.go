package moon

import (
	"github.com/starainrt/astro/tools"
	"errors"
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

var (
	ERR_MOON_NEVER_RISE = errors.New("ERROR:极夜，月亮在今日永远在地平线下！")
	ERR_MOON_NEVER_SET  = errors.New("ERROR:极昼，月亮在今日永远在地平线上！")
	// ERR_MOON_NEVER_DOWN deprecated: -- use ERR_MOON_NEVER_SET instead
	ERR_MOON_NEVER_DOWN = ERR_MOON_NEVER_SET
	ERR_NOT_TODAY       = errors.New("ERROR:月亮已在（昨日/明日）（升起/降下）")
)

func riseSetResult(date time.Time, jde float64, err error) (time.Time, error) {
	if err != nil {
		switch {
		case errors.Is(err, basic.ErrNotOnThisDate):
			return time.Time{}, ERR_NOT_TODAY
		case errors.Is(err, basic.ErrNeverRise):
			return time.Time{}, ERR_MOON_NEVER_RISE
		case errors.Is(err, basic.ErrNeverSet):
			return time.Time{}, ERR_MOON_NEVER_SET
		default:
			return time.Time{}, err
		}
	}
	return basic.JDE2DateByZone(jde, date.Location(), true), nil
}

// TrueLo 月亮真黄经 / true ecliptic longitude.
//
// 返回月亮在 date 对应绝对时刻的地心真黄经，单位度。
// Returns the Moon's geocentric true ecliptic longitude at the instant represented by date, in degrees.
func TrueLo(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.HMoonTrueLo(basic.TD2UT(jde, true))
}

// TrueLoN 截断项月亮真黄经 / truncated true ecliptic longitude.
//
// 参数与 TrueLo 相同；n<0 使用当前仓库内嵌的全部 ELP 项，其余值用于截断月球级数。
// Uses the same inputs as TrueLo. n<0 keeps all embedded ELP terms in this repository; other values truncate the lunar series.
func TrueLoN(date time.Time, n int) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.HMoonTrueLoN(basic.TD2UT(jde, true), n)
}

// TrueBo 月亮真黄纬 / true ecliptic latitude.
//
// 返回月亮在 date 对应绝对时刻的地心真黄纬，单位度。
// Returns the Moon's geocentric true ecliptic latitude at the instant represented by date, in degrees.
func TrueBo(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.HMoonTrueBo(basic.TD2UT(jde, true))
}

// TrueBoN 截断项月亮真黄纬 / truncated true ecliptic latitude.
//
// 参数与 TrueBo 相同；n<0 使用当前仓库内嵌的全部 ELP 项，其余值用于截断月球级数。
// Uses the same inputs as TrueBo. n<0 keeps all embedded ELP terms in this repository; other values truncate the lunar series.
func TrueBoN(date time.Time, n int) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.HMoonTrueBoN(basic.TD2UT(jde, true), n)
}

// ApparentLo 月亮地心视黄经 / apparent geocentric ecliptic longitude.
//
// 返回月亮在 date 对应绝对时刻的地心视黄经，单位度。
// Returns the Moon's apparent geocentric ecliptic longitude at the instant represented by date, in degrees.
func ApparentLo(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.HMoonApparentLo(basic.TD2UT(jde, true))
}

// TrueRa 月亮地心真赤经 / true geocentric right ascension.
//
// 返回月亮在 date 对应绝对时刻的地心真赤经，单位度。
// Returns the Moon's geocentric true right ascension at the instant represented by date, in degrees.
func TrueRa(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.HMoonGeocentricTrueRa(basic.TD2UT(jde, true))
}

// TrueDec 月亮地心真赤纬 / true geocentric declination.
//
// 返回月亮在 date 对应绝对时刻的地心真赤纬，单位度。
// Returns the Moon's geocentric true declination at the instant represented by date, in degrees.
func TrueDec(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.HMoonGeocentricTrueDec(basic.TD2UT(jde, true))
}

// TrueRaDec 月亮地心真赤经、真赤纬 / true geocentric right ascension and declination.
//
// 返回月亮在 date 对应绝对时刻的地心真赤经与真赤纬，单位度。
// Returns the Moon's geocentric true right ascension and declination at the instant represented by date, in degrees.
func TrueRaDec(date time.Time) (float64, float64) {
	jde := basic.Date2JDE(date.UTC())
	return basic.HMoonGeocentricTrueRaDec(basic.TD2UT(jde, true))
}

// GeocentricApparentRa 月亮地心视赤经 / apparent geocentric right ascension.
//
// 返回月亮在 date 对应绝对时刻的地心视赤经，单位度。
// Returns the Moon's apparent geocentric right ascension at the instant represented by date, in degrees.
func GeocentricApparentRa(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.HMoonGeocentricApparentRa(basic.TD2UT(jde, true))
}

// GeocentricApparentDec 月亮地心视赤纬 / apparent geocentric declination.
//
// 返回月亮在 date 对应绝对时刻的地心视赤纬，单位度。
// Returns the Moon's apparent geocentric declination at the instant represented by date, in degrees.
func GeocentricApparentDec(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.HMoonGeocentricApparentDec(basic.TD2UT(jde, true))
}

// GeocentricApparentRaDec 月亮地心视赤经、视赤纬 / apparent geocentric right ascension and declination.
//
// 返回月亮在 date 对应绝对时刻的地心视赤经与视赤纬，单位度。
// Returns the Moon's apparent geocentric right ascension and declination at the instant represented by date, in degrees.
func GeocentricApparentRaDec(date time.Time) (float64, float64) {
	jde := basic.Date2JDE(date.UTC())
	return basic.HMoonGeocentricApparentRaDec(basic.TD2UT(jde, true))
}

// ApparentRa 月亮站心视赤经 / apparent topocentric right ascension.
//
// date 为观测时刻，会读取其时区参与地方时计算；lon/lat 为观测者经纬度，东正西负、北正南负；返回值单位度。
// date is the observing instant and its zone offset participates in local-time calculations. lon/lat are east-positive and north-positive; the result is in degrees.
func ApparentRa(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.HMoonApparentRa(jde, lon, lat, float64(loc)/3600.0)
}

// ApparentDec 月亮站心视赤纬 / apparent topocentric declination.
//
// 参数与 ApparentRa 相同，返回月亮站心视赤纬，单位度。
// Uses the same inputs as ApparentRa and returns the Moon's apparent topocentric declination in degrees.
func ApparentDec(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.HMoonApparentDec(jde, lon, lat, float64(loc)/3600.0)
}

// ApparentRaDec 月亮站心视赤经、视赤纬 / apparent topocentric right ascension and declination.
//
// 参数与 ApparentRa 相同，返回月亮站心视赤经与视赤纬，单位度。
// Uses the same inputs as ApparentRa and returns the Moon's apparent topocentric right ascension and declination in degrees.
func ApparentRaDec(date time.Time, lon, lat float64) (float64, float64) {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.HMoonApparentRaDec(jde, lon, lat, float64(loc)/3600.0)
}

// HourAngle 月亮时角 / hour angle.
//
// date 为观测时刻，会读取其时区参与地方时计算；lon/lat 为观测者经纬度，东正西负、北正南负；返回值单位度。
// date is the observing instant and its zone offset participates in local-time calculations. lon/lat are east-positive and north-positive; the result is in degrees.
func HourAngle(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.MoonTimeAngle(jde, lon, lat, float64(loc)/3600.0)
}

// Azimuth 月亮方位角 / azimuth.
//
// date 为观测时刻，会读取其时区参与地方时计算；lon/lat 为观测者经纬度，东正西负、北正南负；返回值按正北为 0°、向东增加。
// date is the observing instant and its zone offset participates in local-time calculations. lon/lat are east-positive and north-positive; azimuth is measured from north toward east.
func Azimuth(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.HMoonAzimuth(jde, lon, lat, float64(loc)/3600.0)
}

// Altitude 月亮高度角 / lunar altitude.
//
// date 为观测时刻，会读取其时区参与地方时计算；lon/lat 为观测者经纬度，东正西负、北正南负；返回值单位度。
// date is the observing instant and its zone offset participates in local-time calculations. lon/lat are east-positive and north-positive; the result is in degrees.
func Altitude(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.HMoonHeight(jde, lon, lat, float64(loc)/3600.0)
}

// Zenith 月亮天顶距 / lunar zenith distance.
//
// 参数与 Altitude 相同，返回值为对应时刻的天顶距，单位度。
// Uses the same inputs as Altitude and returns the zenith distance in degrees.
func Zenith(date time.Time, lon, lat float64) float64 {
	return 90 - Altitude(date, lon, lat)
}

// CulminationTime 月亮中天时刻 / culmination time.
//
// date 取其所在时区的当地日期，返回值保持相同时区；lon/lat 为观测者经纬度，东正西负、北正南负。
// date is interpreted on its local civil day and the result keeps the same time zone. lon/lat are east-positive and north-positive.
func CulminationTime(date time.Time, lon, lat float64) time.Time {
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.JDE2DateByZone(basic.MoonCulminationTime(jde, lon, lat, float64(loc)/3600.0), date.Location(), true)
}

// RiseTime 月出时刻 / moonrise time.
//
// date 取其所在时区的当地日期，返回值保持相同时区；lon/lat 为观测者经纬度，东正西负、北正南负；
// height 为海拔高度，单位米；aero 为 true 时加入标准大气折射。
// date is interpreted on its local civil day and the result keeps the same time zone. lon/lat are east-positive and north-positive;
// height is observer elevation in meters, and aero enables standard atmospheric refraction.
func RiseTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	aeroFloat := 0.00
	if aero {
		aeroFloat = 1
	}
	riseJde, err := basic.GetMoonRiseTime(jde, lon, lat, timezone, aeroFloat, height)
	return riseSetResult(date, riseJde, err)
}

// DownTime 月落时刻别名 / deprecated moonset alias.
//
// Deprecated: use SetTime instead.
//
// 参数与 SetTime 相同，仅为兼容旧接口保留。
// Same as SetTime and kept only for backward compatibility.
func DownTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return SetTime(date, lon, lat, height, aero)
}

// SetTime 月落时刻 / moonset time.
//
// 参数与 RiseTime 相同，返回给定当地日期内的月落时刻。
// Uses the same inputs as RiseTime and returns the moonset time on the corresponding local civil day.
func SetTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	aeroFloat := 0.00
	if aero {
		aeroFloat = 1
	}
	downJde, err := basic.GetMoonSetTime(jde, lon, lat, timezone, aeroFloat, height)
	return riseSetResult(date, downJde, err)
}

// SunMoonLoDiff 日月黄经差 / Moon-Sun longitude difference.
//
// 返回月亮视黄经减去太阳视黄经的结果，单位度，取值范围 [0, 360)；新月附近接近 0°，满月附近接近 180°。
// Returns apparent lunar longitude minus apparent solar longitude in degrees, normalized to [0, 360). It is near 0° at new moon and near 180° at full moon.
func SunMoonLoDiff(date time.Time) float64 {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	sunLo := basic.HSunApparentLo(jde)
	moonLo := basic.HMoonApparentLo(jde)
	return tools.Limit360(moonLo - sunLo)
}

// PhaseDesc 月相文字描述 / textual lunar phase description.
//
// 基于 SunMoonLoDiff 的分段结果返回中文月相名称。
// Returns a Chinese phase name derived from the segmented Moon-Sun longitude difference.
func PhaseDesc(date time.Time) string {
	moonSunLoDiff := SunMoonLoDiff(date)
	if moonSunLoDiff >= 0 && moonSunLoDiff <= 30 {
		return "新月"
	} else if moonSunLoDiff > 30 && moonSunLoDiff <= 75 {
		return "上峨眉月"
	} else if moonSunLoDiff > 75 && moonSunLoDiff <= 135 {
		return "上弦月"
	} else if moonSunLoDiff > 135 && moonSunLoDiff < 170 {
		return "盈凸月"
	} else if moonSunLoDiff >= 170 && moonSunLoDiff <= 190 {
		return "满月"
	} else if moonSunLoDiff > 190 && moonSunLoDiff < 225 {
		return "亏凸月"
	} else if moonSunLoDiff >= 225 && moonSunLoDiff < 285 {
		return "下弦月"
	} else if moonSunLoDiff >= 285 && moonSunLoDiff < 330 {
		return "下峨眉月"
	} else {
		return "残月"
	}
}

// Phase 月面受照比例 / illuminated fraction.
//
// 返回月亮在 date 对应绝对时刻的受照比例，范围 [0, 1]。
// Returns the Moon's illuminated fraction at the instant represented by date, in the range [0, 1].
func Phase(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.MoonPhase(basic.TD2UT(jde, true))
}

// ShuoYue 朔月锚点解 / new-moon solution near a decimal year anchor.
//
// year 为公历小数年锚点，例如 2025.0 或 2025.5；返回以该锚点求得的一次朔月时刻，结果为 UTC。
// year is a decimal Gregorian-year anchor such as 2025.0 or 2025.5. The returned time is one new moon solved near that anchor, in UTC.
func ShuoYue(year float64) time.Time {
	jde := basic.TD2UT(basic.CalcMoonSH(year, 0), false)
	return basic.JDE2DateByZone(jde, time.UTC, false)
}

// NextShuoYue 下一次朔月 / next new moon.
//
// 返回 date 之后最近一次朔月时刻，结果保持 date 的时区。
// Returns the next new moon after date, keeping date's time zone.
func NextShuoYue(date time.Time) time.Time {
	return nextMoonPhase(date, 0)
}

// LastShuoYue 上一次朔月 / previous new moon.
//
// 返回 date 之前最近一次朔月时刻，结果保持 date 的时区。
// Returns the previous new moon before date, keeping date's time zone.
func LastShuoYue(date time.Time) time.Time {
	return lastMoonPhase(date, 0)
}

// ClosestShuoYue 最近朔月 / closest new moon.
//
// 返回离 date 最近的朔月时刻，结果保持 date 的时区。
// Returns the new moon nearest to date, keeping date's time zone.
func ClosestShuoYue(date time.Time) time.Time {
	return closestMoonPhase(date, 0)
}

// NewMoon 朔月英文别名 / English alias for ShuoYue.
func NewMoon(year float64) time.Time {
	return ShuoYue(year)
}

// NextNewMoon 下一次朔月英文别名 / English alias for NextShuoYue.
func NextNewMoon(date time.Time) time.Time {
	return NextShuoYue(date)
}

// LastNewMoon 上一次朔月英文别名 / English alias for LastShuoYue.
func LastNewMoon(date time.Time) time.Time {
	return LastShuoYue(date)
}

// ClosestNewMoon 最近朔月英文别名 / English alias for ClosestShuoYue.
func ClosestNewMoon(date time.Time) time.Time {
	return ClosestShuoYue(date)
}

func closestMoonPhase(date time.Time, typed int) time.Time {
	//0=shuo 1=wang 2=shangxian 3=xiaxian
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	if typed < 2 {
		return basic.JDE2DateByZone(basic.TD2UT(basic.CalcMoonSHByJDE(jde, typed), false), date.Location(), false)
	}
	return basic.JDE2DateByZone(basic.TD2UT(basic.CalcMoonXHByJDE(jde, typed-2), false), date.Location(), false)
}

func nextMoonPhase(date time.Time, typed int) time.Time {
	//0=shuo 1=wang 2=shangxian 3=xiaxian
	diffCode := 0.00
	switch typed {
	case 1:
		diffCode = 180
	case 2:
		diffCode = 90
	case 3:
		diffCode = 270
	}
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	cost := basic.HMoonApparentLo(jde) - basic.HSunApparentLo(jde) - float64(diffCode)
	for cost < 0 {
		cost += 360
	}
	if cost < 0 && math.Floor(math.Abs(cost)*10000) == 0 {
		cost = 0
	}
	if cost < 240 {
		jde += (240 - cost) / 11.19
	}
	if typed < 2 {
		return basic.JDE2DateByZone(basic.TD2UT(basic.CalcMoonSHByJDE(jde, typed), false), date.Location(), false)
	}
	return basic.JDE2DateByZone(basic.TD2UT(basic.CalcMoonXHByJDE(jde, typed-2), false), date.Location(), false)
}

func lastMoonPhase(date time.Time, typed int) time.Time {
	//0=shuo 1=wang 2=shangxian 3=xiaxian
	diffCode := 0.00
	switch typed {
	case 1:
		diffCode = 180
	case 2:
		diffCode = 90
	case 3:
		diffCode = 270
	}
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	cost := basic.HMoonApparentLo(jde) - basic.HSunApparentLo(jde) - float64(diffCode)
	for cost < 0 {
		cost += 360
	}
	if cost > 0 && math.Floor(math.Abs(cost)*10000) == 0 {
		cost = 360
	}
	if cost > 120 {
		jde -= (cost - 120) / 11.19
	}
	if typed < 2 {
		return basic.JDE2DateByZone(basic.TD2UT(basic.CalcMoonSHByJDE(jde, typed), false), date.Location(), false)
	}
	return basic.JDE2DateByZone(basic.TD2UT(basic.CalcMoonXHByJDE(jde, typed-2), false), date.Location(), false)
}

// WangYue 望月锚点解 / full-moon solution near a decimal year anchor.
//
// year 为公历小数年锚点，例如 2025.0 或 2025.5；返回以该锚点求得的一次望月时刻，结果为 UTC。
// year is a decimal Gregorian-year anchor such as 2025.0 or 2025.5. The returned time is one full moon solved near that anchor, in UTC.
func WangYue(year float64) time.Time {
	jde := basic.TD2UT(basic.CalcMoonSH(year, 1), false)
	return basic.JDE2DateByZone(jde, time.UTC, false)
}

// NextWangYue 下一次望月 / next full moon.
//
// 返回 date 之后最近一次望月时刻，结果保持 date 的时区。
// Returns the next full moon after date, keeping date's time zone.
func NextWangYue(date time.Time) time.Time {
	return nextMoonPhase(date, 1)
}

// LastWangYue 上一次望月 / previous full moon.
//
// 返回 date 之前最近一次望月时刻，结果保持 date 的时区。
// Returns the previous full moon before date, keeping date's time zone.
func LastWangYue(date time.Time) time.Time {
	return lastMoonPhase(date, 1)
}

// ClosestWangYue 最近望月 / closest full moon.
//
// 返回离 date 最近的望月时刻，结果保持 date 的时区。
// Returns the full moon nearest to date, keeping date's time zone.
func ClosestWangYue(date time.Time) time.Time {
	return closestMoonPhase(date, 1)
}

// FullMoon 望月英文别名 / English alias for WangYue.
func FullMoon(year float64) time.Time {
	return WangYue(year)
}

// NextFullMoon 下一次望月英文别名 / English alias for NextWangYue.
func NextFullMoon(date time.Time) time.Time {
	return NextWangYue(date)
}

// LastFullMoon 上一次望月英文别名 / English alias for LastWangYue.
func LastFullMoon(date time.Time) time.Time {
	return LastWangYue(date)
}

// ClosestFullMoon 最近望月英文别名 / English alias for ClosestWangYue.
func ClosestFullMoon(date time.Time) time.Time {
	return ClosestWangYue(date)
}

// ShangXianYue 上弦锚点解 / first-quarter solution near a decimal year anchor.
//
// year 为公历小数年锚点，例如 2025.0 或 2025.5；返回以该锚点求得的一次上弦时刻，结果为 UTC。
// year is a decimal Gregorian-year anchor such as 2025.0 or 2025.5. The returned time is one first-quarter solution near that anchor, in UTC.
func ShangXianYue(year float64) time.Time {
	jde := basic.TD2UT(basic.CalcMoonXH(year, 0), false)
	return basic.JDE2DateByZone(jde, time.UTC, false)
}

// NextShangXianYue 下一次上弦 / next first quarter.
//
// 返回 date 之后最近一次上弦时刻，结果保持 date 的时区。
// Returns the next first quarter after date, keeping date's time zone.
func NextShangXianYue(date time.Time) time.Time {
	return nextMoonPhase(date, 2)
}

// LastShangXianYue 上一次上弦 / previous first quarter.
//
// 返回 date 之前最近一次上弦时刻，结果保持 date 的时区。
// Returns the previous first quarter before date, keeping date's time zone.
func LastShangXianYue(date time.Time) time.Time {
	return lastMoonPhase(date, 2)
}

// ClosestShangXianYue 最近上弦 / closest first quarter.
//
// 返回离 date 最近的上弦时刻，结果保持 date 的时区。
// Returns the first quarter nearest to date, keeping date's time zone.
func ClosestShangXianYue(date time.Time) time.Time {
	return closestMoonPhase(date, 2)
}

// FirstQuarter 上弦英文别名 / English alias for ShangXianYue.
func FirstQuarter(year float64) time.Time {
	return ShangXianYue(year)
}

// NextFirstQuarter 下一次上弦英文别名 / English alias for NextShangXianYue.
func NextFirstQuarter(date time.Time) time.Time {
	return NextShangXianYue(date)
}

// LastFirstQuarter 上一次上弦英文别名 / English alias for LastShangXianYue.
func LastFirstQuarter(date time.Time) time.Time {
	return LastShangXianYue(date)
}

// ClosestFirstQuarter 最近上弦英文别名 / English alias for ClosestShangXianYue.
func ClosestFirstQuarter(date time.Time) time.Time {
	return ClosestShangXianYue(date)
}

// XiaXianYue 下弦锚点解 / last-quarter solution near a decimal year anchor.
//
// year 为公历小数年锚点，例如 2025.0 或 2025.5；返回以该锚点求得的一次下弦时刻，结果为 UTC。
// year is a decimal Gregorian-year anchor such as 2025.0 or 2025.5. The returned time is one last-quarter solution near that anchor, in UTC.
func XiaXianYue(year float64) time.Time {
	jde := basic.TD2UT(basic.CalcMoonXH(year, 1), false)
	return basic.JDE2DateByZone(jde, time.UTC, false)
}

// NextXiaXianYue 下一次下弦 / next last quarter.
//
// 返回 date 之后最近一次下弦时刻，结果保持 date 的时区。
// Returns the next last quarter after date, keeping date's time zone.
func NextXiaXianYue(date time.Time) time.Time {
	return nextMoonPhase(date, 3)
}

// LastXiaXianYue 上一次下弦 / previous last quarter.
//
// 返回 date 之前最近一次下弦时刻，结果保持 date 的时区。
// Returns the previous last quarter before date, keeping date's time zone.
func LastXiaXianYue(date time.Time) time.Time {
	return lastMoonPhase(date, 3)
}

// ClosestXiaXianYue 最近下弦 / closest last quarter.
//
// 返回离 date 最近的下弦时刻，结果保持 date 的时区。
// Returns the last quarter nearest to date, keeping date's time zone.
func ClosestXiaXianYue(date time.Time) time.Time {
	return closestMoonPhase(date, 3)
}

// LastQuarter 下弦英文别名 / English alias for XiaXianYue.
func LastQuarter(year float64) time.Time {
	return XiaXianYue(year)
}

// NextLastQuarter 下一次下弦英文别名 / English alias for NextXiaXianYue.
func NextLastQuarter(date time.Time) time.Time {
	return NextXiaXianYue(date)
}

// LastLastQuarter 上一次下弦英文别名 / English alias for LastXiaXianYue.
func LastLastQuarter(date time.Time) time.Time {
	return LastXiaXianYue(date)
}

// ClosestLastQuarter 最近下弦英文别名 / English alias for ClosestXiaXianYue.
func ClosestLastQuarter(date time.Time) time.Time {
	return ClosestXiaXianYue(date)
}

// EarthDistance 地月距离 / Earth-Moon distance.
//
// 返回月亮在 date 对应绝对时刻到地球质心的距离，单位千米。
// Returns the distance from the Moon to Earth's center at the instant represented by date, in kilometers.
func EarthDistance(date time.Time) float64 {
	jde := basic.Date2JDE(date)
	jde = basic.TD2UT(jde, true)
	return basic.MoonAway(jde)
}
