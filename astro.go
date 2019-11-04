package astro

import (
	"errors"
	"time"

	"github.com/starainrt/astro/basic"
)

/*
获取当前时刻（本地时间）对应的儒略日时间
*/
func NowJDE() float64 {
	return basic.GetNowJDE()
}

/*
日期转儒略日
*/
func Date2JDE(date time.Time) float64 {
	day := float64(date.Day()) + float64(date.Hour())/24.0 + float64(date.Minute())/24.0/60.0 + float64(date.Second())/24.0/3600.0 + float64(date.Nanosecond())/1000000000.0/3600.0/24.0
	return basic.JDECalc(date.Year(), int(date.Month()), day)
}

/*
儒略日转日期
*/
func JDE2Date(jde float64) time.Time {
	return basic.JDE2Date(jde)
}

/*
公历转农历
返回：月，日，是否闰月，文字描述
*/
func Lunar(year, month, day int) (int, int, bool, string) {
	return basic.GetLunar(year, month, day)
}

/*
农历转公历
*/
func Solar(year, month, day int, leap bool) time.Time {
	jde := basic.GetSolar(year, month, day, leap)
	return JDE2Date(jde)
}

/*
  太阳升起时间
  jde，世界时当地0时 JDE
  lon，经度，东正西负
  lat，纬度，北正南负
  timezone，时区，东正西负
  aero，true时进行大气修正
*/
func SunRiseTime(jde, lon, lat, timezone float64, aero bool) (time.Time, error) {
	var err error = nil
	tz := 0.00
	if aero {
		tz = 1
	}
	tm := basic.GetSunRiseTime(jde, lon, lat, timezone, tz)
	if tm == -2 {
		err = errors.New("极夜")
	}
	if tm == -1 {
		err = errors.New("极昼")
	}
	return JDE2Date(tm), err
}

/*
  太阳落下时间
  jde，世界时当地0时 JDE
  lon，经度，东正西负
  lat，纬度，北正南负
  timezone，时区，东正西负
  aero，true时进行大气修正
*/
func SunDownTime(jde, lon, lat, timezone float64, aero bool) (time.Time, error) {
	var err error = nil
	tz := 0.00
	if aero {
		tz = 1
	}
	tm := basic.GetSunDownTime(jde, lon, lat, timezone, tz)
	if tm == -2 {
		err = errors.New("极夜")
	}
	if tm == -1 {
		err = errors.New("极昼")
	}
	return JDE2Date(tm), err
}

/*
  晨朦影
  jde，世界时当地0时 JDE
  lon，经度，东正西负
  lat，纬度，北正南负
  timezone，时区，东正西负
  angle，朦影角度：可选-6 -12 -18
*/
func MorningTwilightTime(jde, lon, lat, timezone, angle float64) (time.Time, error) {
	var err error = nil
	tm := basic.GetAsaTime(jde, lon, lat, timezone, angle)
	if tm == -2 {
		err = errors.New("不存在")
	}
	if tm == -1 {
		err = errors.New("不存在")
	}
	return JDE2Date(tm), err
}

/*
  昏朦影
  jde，世界时当地0时 JDE
  lon，经度，东正西负
  lat，纬度，北正南负
  timezone，时区，东正西负
  angle，朦影角度：可选-6 -12 -18
*/
func EveningTwilightTime(jde, lon, lat, timezone, angle float64) (time.Time, error) {
	var err error = nil
	tm := basic.GetBanTime(jde, lon, lat, timezone, angle)
	if tm == -2 {
		err = errors.New("不存在")
	}
	if tm == -1 {
		err = errors.New("不存在")
	}
	return JDE2Date(tm), err
}

/*
  月亮升起时间
  jde，世界时当地0时 JDE
  lon，经度，东正西负
  lat，纬度，北正南负
  timezone，时区，东正西负
  aero，true时进行大气修正
*/
func MoonRiseTime(jde, lon, lat, timezone float64, aero bool) (time.Time, error) {
	var err error = nil
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

/*
  月亮落下时间
  jde，世界时当地0时 JDE
  lon，经度，东正西负
  lat，纬度，北正南负
  timezone，时区，东正西负
  aero，true时进行大气修正
*/
func MoonDownTime(jde, lon, lat, timezone float64, aero bool) (time.Time, error) {
	var err error = nil
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

/*
月相
jde，世界时UTC JDE
*/
func Phase(jde float64) float64 {
	return basic.MoonLight(basic.TD2UT(jde, true))
}
