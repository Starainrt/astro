package sun

import (
	"errors"
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

var (
	ERR_SUN_NEVER_RISE = errors.New("ERROR:极夜，太阳在今日永远在地平线下！")
	ERR_SUN_NEVER_SET  = errors.New("ERROR:极昼，太阳在今日永远在地平线上！")
	// ERR_SUN_NEVER_DOWN deprecated: -- use ERR_SUN_NEVER_RISE instead
	ERR_SUN_NEVER_DOWN      = ERR_SUN_NEVER_SET
	ERR_TWILIGHT_NOT_EXISTS = errors.New("ERROR:今日晨昏朦影不存在！")
)

func riseSetResult(date time.Time, jde float64, err error) (time.Time, error) {
	if err != nil {
		switch {
		case errors.Is(err, basic.ErrNeverRise):
			return time.Time{}, ERR_SUN_NEVER_RISE
		case errors.Is(err, basic.ErrNeverSet):
			return time.Time{}, ERR_SUN_NEVER_SET
		default:
			return time.Time{}, err
		}
	}
	return basic.JDE2DateByZone(jde, date.Location(), true), nil
}

func twilightResult(date time.Time, jde float64, err error) (time.Time, error) {
	if err != nil {
		return time.Time{}, ERR_TWILIGHT_NOT_EXISTS
	}
	return basic.JDE2DateByZone(jde, date.Location(), true), nil
}

/*
太阳
视星等	−26.74
绝对星等	4.839
光谱类型	G2V
金属量	Z = 0.0122
角直径	31.6′ – 32.7′
*/

// RiseTime 日出时刻 / sunrise time.
//
// date 取其所在时区的当地日期，返回值保持相同时区；lon/lat 为观测者经纬度，东正西负、北正南负；
// height 为海拔高度，单位米；aero 为 true 时加入标准大气折射。
// date is interpreted on its local civil day and the result keeps the same time zone. lon/lat are east-positive and north-positive;
// height is observer elevation in meters, and aero enables standard atmospheric refraction.
func RiseTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	// 以 date 的当地日期为锚点，并读取其时区偏移参与地方时计算。
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	// 返回值保持与输入 date 一致的时区。
	riseJde, err := basic.GetSunRiseTime(jde, lon, lat, timezone, aeroFloat, height)
	return riseSetResult(date, riseJde, err)
}

// RiseTimeN 截断项日出时刻 / truncated sunrise time.
//
// 参数与 RiseTime 相同；n<0 使用当前仓库内嵌的全部 VSOP 项，其余值用于截断太阳位置级数。
// Uses the same inputs as RiseTime. n<0 keeps all embedded VSOP terms in this repository; other values truncate the solar series.
func RiseTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error) {
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
	riseJde, err := basic.GetSunRiseTimeN(jde, lon, lat, timezone, aeroFloat, height, n)
	return riseSetResult(date, riseJde, err)
}

// DownTime 日落时刻别名 / deprecated sunset alias.
//
// Deprecated: use SetTime instead.
//
// 参数与 SetTime 相同，仅为兼容旧接口保留。
// Same as SetTime and kept only for backward compatibility.
func DownTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return SetTime(date, lon, lat, height, aero)
}

// DownTimeN 截断项日落时刻别名 / deprecated truncated sunset alias.
//
// Deprecated: use SetTimeN instead.
//
// 参数与 SetTimeN 相同，仅为兼容旧接口保留。
// Same as SetTimeN and kept only for backward compatibility.
func DownTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error) {
	return SetTimeN(date, lon, lat, height, aero, n)
}

// SetTime 日落时刻 / sunset time.
//
// 参数与 RiseTime 相同，返回给定当地日期内的日落时刻。
// Uses the same inputs as RiseTime and returns the sunset time on the corresponding local civil day.
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
	downJde, err := basic.GetSunSetTime(jde, lon, lat, timezone, aeroFloat, height)
	return riseSetResult(date, downJde, err)
}

// SetTimeN 截断项日落时刻 / truncated sunset time.
//
// 参数与 RiseTimeN 相同，返回给定当地日期内的日落时刻。
// Uses the same inputs as RiseTimeN and returns the sunset time on the corresponding local civil day.
func SetTimeN(date time.Time, lon, lat, height float64, aero bool, n int) (time.Time, error) {
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
	downJde, err := basic.GetSunSetTimeN(jde, lon, lat, timezone, aeroFloat, height, n)
	return riseSetResult(date, downJde, err)
}

// MorningTwilight 晨光始时 / morning twilight.
//
// date 取其所在时区的当地日期，返回值保持相同时区；lon/lat 为观测者经纬度，东正西负、北正南负；
// angle 为目标太阳高度角，常用 -6/-12/-18 度，分别对应民用、航海、天文朦影。
// date is interpreted on its local civil day and the result keeps the same time zone. lon/lat are east-positive and north-positive;
// angle is the target solar altitude in degrees, typically -6, -12, or -18 for civil, nautical, and astronomical twilight.
func MorningTwilight(date time.Time, lon, lat, angle float64) (time.Time, error) {
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJde, err := basic.MorningTwilight(jde, lon, lat, timezone, angle)
	return twilightResult(date, calcJde, err)
}

// MorningTwilightN 截断项晨光始时 / truncated morning twilight.
//
// 参数与 MorningTwilight 相同；n<0 使用当前仓库内嵌的全部 VSOP 项，其余值用于截断太阳位置级数。
// Uses the same inputs as MorningTwilight. n<0 keeps all embedded VSOP terms in this repository; other values truncate the solar series.
func MorningTwilightN(date time.Time, lon, lat, angle float64, n int) (time.Time, error) {
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJde, err := basic.MorningTwilightN(jde, lon, lat, timezone, angle, n)
	return twilightResult(date, calcJde, err)
}

// EveningTwilight 暮光终时 / evening twilight.
//
// 参数与 MorningTwilight 相同，返回对应当地日期的暮光结束时刻。
// Uses the same inputs as MorningTwilight and returns the evening-twilight time on the corresponding local civil day.
func EveningTwilight(date time.Time, lon, lat, angle float64) (time.Time, error) {
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	//不需要进行力学时转换，会在GetBanTime中转换
	calcJde, err := basic.EveningTwilight(jde, lon, lat, timezone, angle)
	return twilightResult(date, calcJde, err)
}

// EveningTwilightN 截断项暮光终时 / truncated evening twilight.
//
// 参数与 MorningTwilightN 相同，返回对应当地日期的暮光结束时刻。
// Uses the same inputs as MorningTwilightN and returns the evening-twilight time on the corresponding local civil day.
func EveningTwilightN(date time.Time, lon, lat, angle float64, n int) (time.Time, error) {
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJde, err := basic.EveningTwilightN(jde, lon, lat, timezone, angle, n)
	return twilightResult(date, calcJde, err)
}

// EclipticObliquity 黄赤交角 / ecliptic obliquity.
//
// 返回 date 对应绝对时刻的黄赤交角，单位度；nutation 为 true 时加入交角章动。
// Returns the obliquity of the ecliptic at the instant represented by date, in degrees. When nutation is true, obliquity nutation is included.
func EclipticObliquity(date time.Time, nutation bool) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	//进行力学时转换
	jde = basic.TD2UT(jde, true)
	//黄赤交角计算
	return basic.EclipticObliquity(jde, nutation)
}

// EclipticNutation 黄经章动（IAU 2000B） / nutation in longitude, IAU 2000B.
//
// 返回 date 对应绝对时刻的黄经章动，单位度。
// Returns nutation in longitude at the instant represented by date, in degrees.
func EclipticNutation(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	//进行力学时转换与章动计算
	return basic.Nutation2000Bi(basic.TD2UT(jde, true))
}

// EclipticNutation1980 黄经章动（IAU 1980） / nutation in longitude, IAU 1980.
//
// 返回 date 对应绝对时刻的黄经章动，单位度。
// Returns nutation in longitude at the instant represented by date, in degrees.
func EclipticNutation1980(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	//进行力学时转换与章动计算
	return basic.Nutation1980i(basic.TD2UT(jde, true))
}

// AxialtiltNutation 交角章动（IAU 2000B） / nutation in obliquity, IAU 2000B.
//
// 返回 date 对应绝对时刻的交角章动，单位度。
// Returns nutation in obliquity at the instant represented by date, in degrees.
func AxialtiltNutation(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	//进行力学时转换与章动计算
	return basic.Nutation2000Bs(basic.TD2UT(jde, true))
}

// AxialtiltNutation1980 交角章动（IAU 1980） / nutation in obliquity, IAU 1980.
//
// 返回 date 对应绝对时刻的交角章动，单位度。
// Returns nutation in obliquity at the instant represented by date, in degrees.
func AxialtiltNutation1980(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	//进行力学时转换与章动计算
	return basic.Nutation1980s(basic.TD2UT(jde, true))
}

// GeometricLo 太阳几何黄经 / geometric ecliptic longitude.
//
// 返回 date 对应绝对时刻的太阳几何黄经，单位度。
// Returns the Sun's geometric ecliptic longitude at the instant represented by date, in degrees.
func GeometricLo(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.SunLo(basic.TD2UT(jde, true))
}

// TrueLo 太阳真黄经 / true ecliptic longitude.
//
// 返回 date 对应绝对时刻的太阳真黄经，单位度。
// Returns the Sun's true ecliptic longitude at the instant represented by date, in degrees.
func TrueLo(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.HSunTrueLo(basic.TD2UT(jde, true))
}

// TrueLoN 截断项太阳真黄经 / truncated true ecliptic longitude.
//
// 参数与 TrueLo 相同；n<0 使用当前仓库内嵌的全部 VSOP 项，其余值用于截断太阳位置级数。
// Uses the same inputs as TrueLo. n<0 keeps all embedded VSOP terms in this repository; other values truncate the solar series.
func TrueLoN(date time.Time, n int) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.HSunTrueLoN(basic.TD2UT(jde, true), n)
}

// TrueBo 太阳真黄纬 / true ecliptic latitude.
//
// 返回 date 对应绝对时刻的太阳真黄纬，单位度。
// Returns the Sun's true ecliptic latitude at the instant represented by date, in degrees.
func TrueBo(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.HSunTrueBo(basic.TD2UT(jde, true))
}

// TrueBoN 截断项太阳真黄纬 / truncated true ecliptic latitude.
//
// 参数与 TrueBo 相同；n<0 使用当前仓库内嵌的全部 VSOP 项，其余值用于截断太阳位置级数。
// Uses the same inputs as TrueBo. n<0 keeps all embedded VSOP terms in this repository; other values truncate the solar series.
func TrueBoN(date time.Time, n int) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.HSunTrueBoN(basic.TD2UT(jde, true), n)
}

// ApparentLo 太阳视黄经 / apparent ecliptic longitude.
//
// 返回 date 对应绝对时刻的太阳视黄经，单位度。
// Returns the Sun's apparent ecliptic longitude at the instant represented by date, in degrees.
func ApparentLo(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.HSunApparentLo(basic.TD2UT(jde, true))
}

// ApparentRa 太阳地心视赤经 / apparent geocentric right ascension.
//
// 返回 date 对应绝对时刻的太阳地心视赤经，单位度。
// Returns the Sun's apparent geocentric right ascension at the instant represented by date, in degrees.
func ApparentRa(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.HSunApparentRa(basic.TD2UT(jde, true))
}

// ApparentDec 太阳地心视赤纬 / apparent geocentric declination.
//
// 返回 date 对应绝对时刻的太阳地心视赤纬，单位度。
// Returns the Sun's apparent geocentric declination at the instant represented by date, in degrees.
func ApparentDec(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.HSunApparentDec(basic.TD2UT(jde, true))
}

// ApparentRaDec 太阳地心视赤经、视赤纬 / apparent geocentric right ascension and declination.
//
// 返回 date 对应绝对时刻的太阳地心视赤经与视赤纬，单位度。
// Returns the Sun's apparent geocentric right ascension and declination at the instant represented by date, in degrees.
func ApparentRaDec(date time.Time) (float64, float64) {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.HSunApparentRaDec(basic.TD2UT(jde, true))
}

// MidFunc 太阳中心差 / solar equation of center.
//
// 返回 date 对应绝对时刻的太阳中心差，单位度。
// Returns the Sun's equation of center at the instant represented by date, in degrees.
func MidFunc(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.SunMidFun(basic.TD2UT(jde, true))
}

// EquationTime 均时差 / equation of time.
//
// 返回 date 对应绝对时刻的均时差，单位小时。
// Returns the equation of time at the instant represented by date, in hours.
func EquationTime(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.SunTime(basic.TD2UT(jde, true))
}

// HourAngle 太阳时角 / hour angle.
//
// date 为观测时刻，会读取其时区参与地方时计算；lon 为观测者经度，东正西负；返回值单位度。
// lat 目前不参与计算，仅为与其他观测接口保持参数形状一致。
// date is the observing instant and its zone offset participates in local-time calculations. lon is east-positive longitude and the result is in degrees.
// lat is currently unused and kept only for API symmetry with the other observation helpers.
func HourAngle(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SunTimeAngle(jde, lon, lat, timezone)
}

// HourAngleN 截断项太阳时角 / truncated hour angle.
//
// 参数与 HourAngle 相同；n<0 使用当前仓库内嵌的全部 VSOP 项，其余值用于截断太阳位置级数。
// Uses the same inputs as HourAngle. n<0 keeps all embedded VSOP terms in this repository; other values truncate the solar series.
func HourAngleN(date time.Time, lon, lat float64, n int) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SunTimeAngleN(jde, lon, lat, timezone, n)
}

// Azimuth 太阳方位角 / azimuth.
//
// date 为观测时刻，会读取其时区参与地方时计算；lon/lat 为观测者经纬度，东正西负、北正南负；返回值按正北为 0°、向东增加。
// date is the observing instant and its zone offset participates in local-time calculations. lon/lat are east-positive and north-positive; azimuth is measured from north toward east.
func Azimuth(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SunAzimuth(jde, lon, lat, timezone)
}

// AzimuthN 截断项太阳方位角 / truncated azimuth.
//
// 参数与 Azimuth 相同；n<0 使用当前仓库内嵌的全部 VSOP 项，其余值用于截断太阳位置级数。
// Uses the same inputs as Azimuth. n<0 keeps all embedded VSOP terms in this repository; other values truncate the solar series.
func AzimuthN(date time.Time, lon, lat float64, n int) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SunAzimuthN(jde, lon, lat, timezone, n)
}

// Altitude 太阳高度角 / solar altitude.
//
// date 为观测时刻，会读取其时区参与地方时计算；lon/lat 为观测者经纬度，东正西负、北正南负；返回值单位度。
// date is the observing instant and its zone offset participates in local-time calculations. lon/lat are east-positive and north-positive; the result is in degrees.
func Altitude(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SunHeight(jde, lon, lat, timezone)
}

// Zenith 太阳天顶距 / solar zenith distance.
//
// 参数与 Altitude 相同，返回值为对应时刻的天顶距，单位度。
// Uses the same inputs as Altitude and returns the zenith distance in degrees.
func Zenith(date time.Time, lon, lat float64) float64 {
	return 90 - Altitude(date, lon, lat)
}

// AltitudeN 截断项太阳高度角 / truncated solar altitude.
//
// 参数与 Altitude 相同；n<0 使用当前仓库内嵌的全部 VSOP 项，其余值用于截断太阳位置级数。
// Uses the same inputs as Altitude. n<0 keeps all embedded VSOP terms in this repository; other values truncate the solar series.
func AltitudeN(date time.Time, lon, lat float64, n int) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SunHeightN(jde, lon, lat, timezone, n)
}

// ZenithN 截断项太阳天顶距 / truncated solar zenith distance.
//
// 参数与 AltitudeN 相同，返回值为对应时刻的天顶距，单位度。
// Uses the same inputs as AltitudeN and returns the zenith distance in degrees.
func ZenithN(date time.Time, lon, lat float64, n int) float64 {
	return 90 - AltitudeN(date, lon, lat, n)
}

// CulminationTime 太阳中天时刻 / culmination time.
//
// date 取其所在时区的当地日期，返回值保持相同时区；lon 为观测者经度，东正西负。
// date is interpreted on its local civil day and the result keeps the same time zone. lon is east-positive longitude.
func CulminationTime(date time.Time, lon float64) time.Time {
	jde := basic.Date2JDE(date.Add(time.Duration(-1*date.Hour())*time.Hour)) + 0.5
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJde := basic.CulminationTime(jde, lon, timezone) - timezone/24.00
	return basic.JDE2DateByZone(calcJde, date.Location(), false)
}

// CulminationTimeN 截断项太阳中天时刻 / truncated culmination time.
//
// 参数与 CulminationTime 相同；n<0 使用当前仓库内嵌的全部 VSOP 项，其余值用于截断太阳位置级数。
// Uses the same inputs as CulminationTime. n<0 keeps all embedded VSOP terms in this repository; other values truncate the solar series.
func CulminationTimeN(date time.Time, lon float64, n int) time.Time {
	jde := basic.Date2JDE(date.Add(time.Duration(-1*date.Hour())*time.Hour)) + 0.5
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJde := basic.CulminationTimeN(jde, lon, timezone, n) - timezone/24.00
	return basic.JDE2DateByZone(calcJde, date.Location(), false)
}

// EarthDistance 日地距离 / Earth-Sun distance.
//
// 返回 date 对应绝对时刻的日地距离，单位 AU。
// Returns the Earth-Sun distance at the instant represented by date, in astronomical units.
func EarthDistance(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	jde = basic.TD2UT(jde, true)
	return basic.EarthAway(jde)
}

// ApparentSolarTime 真太阳时 / apparent solar time.
//
// 返回 date 这一绝对时刻在给定经度 lon 处对应的真太阳时，结果时区为按经度换算的地方平太阳时区。
// Returns the apparent solar time for the instant represented by date at longitude lon. The result uses a synthetic local-solar time zone derived from longitude.
func ApparentSolarTime(date time.Time, lon float64) time.Time {
	//真太阳时=太阳时角+12小时
	trueTime := (HourAngle(date, lon, 0) + 180) / 15
	if trueTime > 24 {
		trueTime -= 24
	}
	//真太阳时的分
	minute := (trueTime - math.Floor(trueTime)) * 60
	//真太阳时的秒
	second := (minute - math.Floor(minute)) * 60
	//当地经度下的本地时区
	trueSunTime := date.In(time.FixedZone("LTZ", int(lon*3600.00/15.0)))
	if trueSunTime.Hour()-int(trueTime) > 12 {
		trueSunTime = trueSunTime.Add(time.Hour * 24)
	} else if int(trueTime)-trueSunTime.Hour() > 12 {
		trueSunTime = trueSunTime.Add(-time.Hour * 24)
	}
	return time.Date(trueSunTime.Year(), trueSunTime.Month(), trueSunTime.Day(),
		int(trueTime), int(minute), int(second), int((second-math.Floor(second))*1000000000),
		time.FixedZone("LTZ", int(lon*3600.00/15.0)))
}

// ApparentSolarTimeN 截断项真太阳时 / truncated apparent solar time.
//
// 参数与 ApparentSolarTime 相同；n<0 使用当前仓库内嵌的全部 VSOP 项，其余值用于截断太阳位置级数。
// Uses the same inputs as ApparentSolarTime. n<0 keeps all embedded VSOP terms in this repository; other values truncate the solar series.
func ApparentSolarTimeN(date time.Time, lon float64, n int) time.Time {
	trueTime := (HourAngleN(date, lon, 0, n) + 180) / 15
	if trueTime > 24 {
		trueTime -= 24
	}
	minute := (trueTime - math.Floor(trueTime)) * 60
	second := (minute - math.Floor(minute)) * 60
	trueSunTime := date.In(time.FixedZone("LTZ", int(lon*3600.00/15.0)))
	if trueSunTime.Hour()-int(trueTime) > 12 {
		trueSunTime = trueSunTime.Add(time.Hour * 24)
	} else if int(trueTime)-trueSunTime.Hour() > 12 {
		trueSunTime = trueSunTime.Add(-time.Hour * 24)
	}
	return time.Date(trueSunTime.Year(), trueSunTime.Month(), trueSunTime.Day(),
		int(trueTime), int(minute), int(second), int((second-math.Floor(second))*1000000000),
		time.FixedZone("LTZ", int(lon*3600.00/15.0)))
}
