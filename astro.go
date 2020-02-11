// Package astro is a Go Package with Astronomy
package astro

import (
	"errors"
	"time"

	"github.com/starainrt/astro/basic"
)

// NowJDE 获取当前时刻（本地时间）对应的儒略日时间
func NowJDE() float64 {
	return basic.GetNowJDE()
}

// Date2JDE 日期转儒略日
func Date2JDE(date time.Time) float64 {
	day := float64(date.Day()) + float64(date.Hour())/24.0 + float64(date.Minute())/24.0/60.0 + float64(date.Second())/24.0/3600.0 + float64(date.Nanosecond())/1000000000.0/3600.0/24.0
	return basic.JDECalc(date.Year(), int(date.Month()), day)
}

// JDE2Date 儒略日转日期
func JDE2Date(jde float64) time.Time {
	return basic.JDE2Date(jde)
}

// Lunar 公历转农历
// 传入 公历年月日
// 返回 农历月，日，是否闰月以及文字描述
func Lunar(year, month, day int) (int, int, bool, string) {
	return basic.GetLunar(year, month, day)
}

// Solar 农历转公历
// 传入 农历年份，月，日，是否闰月
// 传出 公历时间
// 农历年份用公历年份代替，但是岁首需要使用农历岁首
// 例：计算己亥猪年腊月三十日对应的公历（即2020年1月24日）
// 由于农历还未到鼠年，故应当传入Solar(2019,12,30,false)
func Solar(year, month, day int, leap bool) time.Time {
	jde := basic.GetSolar(year, month, day, leap)
	return JDE2Date(jde)
}

// MoonRiseTime 月亮升起时间
// jde，世界时当地0时 JDE
// lon，经度，东正西负
// lat，纬度，北正南负
// timezone，时区，东正西负
// aero，true时进行大气修正
func MoonRiseTime(jde, lon, lat, timezone float64, aero bool) (time.Time, error) {
	var err error
	tz := 0.00
	if aero {
		tz = 1
	}
	tm := basic.GetMoonRiseTime(jde, lon, lat, timezone, tz)
	if tm == -3 {
		err = errors.New("非今日")
	}
	if tm == -2 {
		err = errors.New("极夜")
	}
	if tm == -1 {
		err = errors.New("极昼")
	}
	return JDE2Date(tm), err
}

// MoonDownTime 月亮落下时间
// jde，世界时当地0时 JDE
// lon，经度，东正西负
// lat，纬度，北正南负
// timezone，时区，东正西负
//aero，true时进行大气修正
func MoonDownTime(jde, lon, lat, timezone float64, aero bool) (time.Time, error) {
	var err error
	tz := 0.00
	if aero {
		tz = 1
	}
	tm := basic.GetMoonDownTime(jde, lon, lat, timezone, tz)
	if tm == -3 {
		err = errors.New("非今日")
	}
	if tm == -2 {
		err = errors.New("极夜")
	}
	if tm == -1 {
		err = errors.New("极昼")
	}
	return JDE2Date(tm), err
}

// Phase 月相
// jde，世界时UTC对应的儒略日
func Phase(jde float64) float64 {
	return basic.MoonLight(basic.TD2UT(jde, true))
}
