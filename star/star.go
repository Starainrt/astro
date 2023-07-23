package star

import (
	"errors"
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

var (
	ERR_STAR_NEVER_RISE = errors.New("ERROR:极夜，星星在今日永远在地平线下！")
	ERR_STAR_NEVER_DOWN = errors.New("ERROR:极昼，星星在今日永远在地平线上！")
)

// Constellation
// 计算date对应UTC世界时给定Date坐标赤经、赤纬所在的星座
func Constellation(ra, dec float64, date time.Time) string {
	jde := basic.Date2JDE(date.UTC())
	return basic.WhichCst(ra, dec, jde)
}

// MeanSiderealTime UTC 平恒星时
func MeanSiderealTime(date time.Time) float64 {
	return basic.MeanSiderealTime(basic.Date2JDE(date.UTC()))
}

// ApparentSiderealTime UTC真恒星时
func ApparentSiderealTime(date time.Time) float64 {
	return basic.ApparentSiderealTime(basic.Date2JDE(date.UTC()))
}

// RiseTime 星星升起时间
//
//	date, 世界时（忽略此处时区）
//	ra，Date瞬时赤经
//	dec，Date瞬时赤纬
//	lon，经度，东正西负
//	lat，纬度，北正南负
//	height，高度
//	aero,是否进行大气修正
func RiseTime(date time.Time, ra, dec, lon, lat, height float64, aero bool) (time.Time, error) {
	var err error
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJde := basic.StarRiseTime(jde, ra, dec, lon, lat, height, timezone, aero)
	if riseJde == -2 {
		err = ERR_STAR_NEVER_RISE
	}
	if riseJde == -1 {
		err = ERR_STAR_NEVER_DOWN
	}
	return basic.JDE2DateByZone(riseJde, date.Location(), true), err
}

// DownTime 星星升起时间
//
//	date, 世界时（忽略此处时区）
//	ra，Date瞬时赤经
//	dec，Date瞬时赤纬
//	lon，经度，东正西负
//	lat，纬度，北正南负
//	height，高度
//	aero,是否进行大气修正
func DownTime(date time.Time, ra, dec, lon, lat, height float64, aero bool) (time.Time, error) {
	var err error
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJde := basic.StarDownTime(jde, ra, dec, lon, lat, height, timezone, aero)
	if riseJde == -2 {
		err = ERR_STAR_NEVER_RISE
	}
	if riseJde == -1 {
		err = ERR_STAR_NEVER_DOWN
	}
	return basic.JDE2DateByZone(riseJde, date.Location(), true), err
}

// HourAngle 恒星时角
// 返回给定Date赤经、经度、对应date时区date时刻的太阳时角（
func HourAngle(date time.Time, ra, lon float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.StarHourAngle(jde, ra, lon, timezone)
}

// Azimuth 恒星方位角
// 返回给定Date赤经赤纬、经纬度、对应date时区date时刻的恒星方位角（正北为0，向东增加）
func Azimuth(date time.Time, ra, dec, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.StarAzimuth(jde, ra, dec, lon, lat, timezone)
}

// Zenith 恒星高度角
// 返回给定赤经赤纬、经纬度、对应date时区date时刻的太阳高度角
func Zenith(date time.Time, ra, dec, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.StarHeight(jde, ra, dec, lon, lat, timezone)
}

// CulminationTime 恒星中天时间
// 返回给定赤经赤纬、经纬度、对应date时区date时刻的太阳中天日期
func CulminationTime(date time.Time, ra, lon float64) time.Time {
	jde := basic.Date2JDE(date)
	if jde-math.Floor(jde) < 0.5 {
		jde--
	}
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJde := basic.StarCulminationTime(jde, ra, lon, timezone) - timezone/24.00
	return basic.JDE2DateByZone(calcJde, date.Location(), false)
}

// InitStarDatabase 初始化恒星数据库
func InitStarDatabase() error {
	return basic.LoadStarData()
}

// 通过恒星HR编号获取恒星参数
func GetStarDataByHR(hr int) (basic.StarData, error) {
	return basic.StarDataByHR(hr)
}
