package moon

import (
	"errors"
	"time"

	"github.com/starainrt/astro/basic"
)

var (
	ERR_MOON_NEVER_RISE = errors.New("ERROR:极夜，月亮在今日永远在地平线下！")
	ERR_MOON_NEVER_DOWN = errors.New("ERROR:极昼，月亮在今日永远在地平线上！")
	ERR_NOT_TODAY       = errors.New("ERROR:月亮已在（昨日/明日）（升起/降下）")
)

// TrueLo 月亮真黄经
func TrueLo(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.HMoonTrueLo(basic.TD2UT(jde, true))
}

// TrueBo 月亮真黄纬
func TrueBo(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.HMoonTrueBo(basic.TD2UT(jde, true))
}

// SeeLo 月亮视黄经（地心）
// 传入UTC对应的儒略日时间
func SeeLo(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.HMoonSeeLo(basic.TD2UT(jde, true))
}

// SeeRa 月亮视赤经（站心）
// date, 时间
// lon, 经度
// lat, 纬度
// 返回站心坐标
func SeeRa(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.HMoonSeeRa(jde, lon, lat, float64(loc)/3600.0)
}

// SeeDec 月亮视赤纬（站心）
// date, 时间
// lon, 经度
// lat, 纬度
// 返回站心坐标
func SeeDec(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.HMoonSeeDec(jde, lon, lat, float64(loc)/3600.0)
}

// SeeRaDec 月亮视赤纬（站心）
// date, 本地时间
// lon, 经度
// lat, 纬度
// 返回站心坐标
func SeeRaDec(date time.Time, lon, lat float64) (float64, float64) {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.HMoonSeeRaDec(jde, lon, lat, float64(loc)/3600.0)
}

// HourAngle 月亮时角
//  date, 世界时（忽略此处时区）
//  lon，经度，东正西负
//  lat，纬度，北正南负
func HourAngle(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.MoonTimeAngle(jde, lon, lat, float64(loc)/3600.0)
}

// Azimuth 月亮方位角
//  date, 世界时（忽略此处时区）
//  lon，经度，东正西负
//  lat，纬度，北正南负
func Azimuth(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.HMoonAngle(jde, lon, lat, float64(loc)/3600.0)
}

// Zenith 月亮高度角
//  date, 世界时（忽略此处时区）
//  lon，经度，东正西负
//  lat，纬度，北正南负
func Zenith(date time.Time, lon, lat, timezone float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.HMoonHeight(jde, lon, lat, float64(loc)/3600.0)
}

// CulminationTime 月亮中天时间
//  date, 世界时（忽略此处时区）
//  lon，经度，东正西负
//  lat，纬度，北正南负
//  timezone，时区，东正西负
func CulminationTime(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.GetMoonTZTime(jde, lon, lat, float64(loc)/3600.0)
}

// RiseTime 月亮升起时间
//  date, 世界时（忽略此处时区）
//  lon，经度，东正西负
//  lat，纬度，北正南负
//  timezone，时区，东正西负
func RiseTime(date time.Time, lon, lat float64, aero bool) (time.Time, error) {
	var err error
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	aeroFloat := 0.00
	if aero {
		aeroFloat = 1
	}
	riseJde := basic.GetMoonRiseTime(jde, lon, lat, timezone, aeroFloat)
	if riseJde == -3 {
		err = ERR_NOT_TODAY
	}
	if riseJde == -2 {
		err = ERR_MOON_NEVER_RISE
	}
	if riseJde == -1 {
		err = ERR_MOON_NEVER_DOWN
	}
	return basic.JDE2Date(riseJde), err
}

// DownTime 月亮降下时间
//  date, 世界时（忽略此处时区）
//  lon，经度，东正西负
//  lat，纬度，北正南负
//  timezone，时区，东正西负
func DownTime(date time.Time, lon, lat float64, aero bool) (time.Time, error) {
	var err error
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	aeroFloat := 0.00
	if aero {
		aeroFloat = 1
	}
	downJde := basic.GetMoonDownTime(jde, lon, lat, timezone, aeroFloat)
	if downJde == -3 {
		err = ERR_NOT_TODAY
	}
	if downJde == -2 {
		err = ERR_MOON_NEVER_RISE
	}
	if downJde == -1 {
		err = ERR_MOON_NEVER_DOWN
	}
	return basic.JDE2Date(downJde), err
}

// Phase 月相
// 返回Date对应UTC世界时的月相大小
func Phase(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.MoonLight(basic.TD2UT(jde, true))
}

// ShuoYue 朔月
func ShuoYue(year float64) time.Time {
	jde := basic.TD2UT(basic.CalcMoonSH(year, 0), false)
	return basic.JDE2DateByZone(jde, time.UTC)
}

// WangYue 望月
func WangYue(year float64) time.Time {
	jde := basic.TD2UT(basic.CalcMoonSH(year, 1), false)
	return basic.JDE2DateByZone(jde, time.UTC)
}

// ShangXianYue 上弦月
func ShangXianYue(year float64) time.Time {
	jde := basic.TD2UT(basic.CalcMoonXH(year, 0), false)
	return basic.JDE2DateByZone(jde, time.UTC)
}

// XiaXianYue 下弦月
func XiaXianYue(year float64) time.Time {
	jde := basic.TD2UT(basic.CalcMoonXH(year, 1), false)
	return basic.JDE2DateByZone(jde, time.UTC)
}

// EarthDistance 日地距离
// 返回date对应UTC世界时日地距离
func EarthDistance(date time.Time) float64 {
	jde := basic.Date2JDE(date)
	jde = basic.TD2UT(jde, true)
	return basic.MoonAway(jde)
}
