package sundial

import (
	"math"
	"time"

	"github.com/starainrt/astro/sun"
)

// TrueSolarTime 真太阳时 / apparent solar time.
//
// 返回 date 在经度 lon 处对应的真太阳时，口径沿用 `sun.ApparentSolarTime`。
// Returns the apparent solar time for date at longitude lon.
func TrueSolarTime(date time.Time, lon float64) time.Time {
	return sun.ApparentSolarTime(date, lon)
}

// MeanSolarTime 地方平太阳时 / local mean solar time.
//
// 返回 date 在经度 lon 处对应的地方平太阳时，结果时区为按经度换算的地方平太阳时区。
// 该实现直接按“真太阳时 - 均时差”构造，以与本仓库的真太阳时和均时差口径保持一致。
// Returns the local mean solar time for date at longitude lon. The result uses
// a synthetic local-mean-solar time zone derived from longitude and is built as
// apparent solar time minus equation of time to keep the package's conventions aligned.
func MeanSolarTime(date time.Time, lon float64) time.Time {
	return TrueSolarTime(date, lon).Add(time.Duration(-sun.EquationTime(date) * float64(time.Hour)))
}

// HourAngle 太阳时角 / solar hour angle.
//
// 返回经度 lon 处的有符号太阳时角，单位度。上午为负，下午为正，中午为 0°。
// Returns the signed apparent-solar hour angle at longitude lon, in degrees.
func HourAngle(date time.Time, lon float64) float64 {
	return normalizeSigned180(sun.HourAngle(date, lon, 0))
}

// MeanSolarHourAngle 平太阳时对应的太阳时角 / hour angle for local mean solar time.
//
// date 负责提供地方平太阳时日期与时区，原有钟面时间会被 meanSolarHours 替换；
// meanSolarHours 为地方平太阳时钟面读数，单位小时，例如 9.5 表示地方平太阳时 09:30。
// 返回对应的视太阳时角，单位度，上午为负，下午为正。
func MeanSolarHourAngle(date time.Time, meanSolarHours float64) float64 {
	if !isFinite(meanSolarHours) {
		return math.NaN()
	}
	sampleTime := dateWithClockHours(date, meanSolarHours)
	return normalizeSigned180(HourAngle(sampleTime, longitudeFromTimeZone(sampleTime)))
}

// ZoneTimeHourAngle 区时对应的太阳时角 / hour angle for zone time.
//
// zoneTimeHours 为 date 所在时区下的钟面读数，单位小时；lon 为当地经度，东正西负。
// 返回该区时在给定经度上对应的视太阳时角，单位度，上午为负，下午为正。
// date 提供民用日期和时区；其原有钟面时间会被 zoneTimeHours 替换。
func ZoneTimeHourAngle(date time.Time, lon, zoneTimeHours float64) float64 {
	if !isFinite(zoneTimeHours) || !isFinite(lon) {
		return math.NaN()
	}
	return normalizeSigned180(HourAngle(dateWithClockHours(date, zoneTimeHours), lon))
}

// HorizontalHourLineAngle 水平日晷时线角 / horizontal sundial hour-line angle.
//
// lat 为纬度（北正南负），hourAngle 为有符号太阳时角，单位度。返回值是相对午线的时线角，
// 上午在东侧为负，下午在西侧为正。
// lat is the observer latitude in degrees and hourAngle is the signed solar
// hour angle in degrees. The result is the hour-line angle from the noon line:
// east/morning is negative and west/afternoon is positive.
func HorizontalHourLineAngle(lat, hourAngle float64) float64 {
	if !isFinite(hourAngle) || !isFinite(lat) {
		return math.NaN()
	}
	hourAngle = normalizeSigned180(hourAngle)
	latRad := lat * math.Pi / 180
	hourAngleRad := hourAngle * math.Pi / 180
	return math.Atan2(math.Sin(latRad)*math.Sin(hourAngleRad), math.Cos(hourAngleRad)) * 180 / math.Pi
}

// HorizontalHourLineAngleAt 水平日晷时线角 / horizontal sundial hour-line angle.
//
// 先按给定时刻和经度求瞬时视太阳时角，再结合纬度返回水平日晷的时线角。
func HorizontalHourLineAngleAt(date time.Time, lon, lat float64) float64 {
	return HorizontalHourLineAngle(lat, HourAngle(date, lon))
}

func dateWithClockHours(date time.Time, hours float64) time.Time {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	return startOfDay.Add(time.Duration(hours * float64(time.Hour)))
}

func longitudeFromTimeZone(date time.Time) float64 {
	_, offsetSeconds := date.Zone()
	return float64(offsetSeconds) / 240.0
}

func clockHours(date time.Time) float64 {
	return float64(date.Hour()) +
		float64(date.Minute())/60 +
		float64(date.Second())/3600 +
		float64(date.Nanosecond())/3.6e12
}

func normalizeSigned180(value float64) float64 {
	value = math.Mod(value, 360)
	if value < -180 {
		value += 360
	}
	if value >= 180 {
		value -= 360
	}
	return value
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
