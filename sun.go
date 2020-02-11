package astro

import (
	"errors"
	"time"

	"github.com/starainrt/astro/basic"
)

// Here Are Sun Functions!

/*
太阳
视星等	−26.74
绝对星等	4.839
光谱类型	G2V
金属量	Z = 0.0122
角直径	31.6′ – 32.7′
*/

// SunRiseTime 太阳升起时间
// jde，计算日期0时时刻对应的jde
// lon，经度，东正西负
// lat，纬度，北正南负
// timezone，时区，东正西负
// aero，true时进行大气修正
func SunRiseTime(jde, lon, lat, timezone float64, aero bool) (time.Time, error) {
	var err error
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

// SunDownTime 太阳落下时间
// jde，世界时当地0时 JDE
// lon，经度，东正西负
// lat，纬度，北正南负
// timezone，时区，东正西负
// aero，true时进行大气修正
func SunDownTime(jde, lon, lat, timezone float64, aero bool) (time.Time, error) {
	var err error
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

// MorningTwilightTime 晨朦影
// jde，世界时当地0时 JDE
// lon，经度，东正西负
// lat，纬度，北正南负
// timezone，时区，东正西负
// angle，朦影角度：可选-6 -12 -18(民用、航海、天文)
func MorningTwilightTime(jde, lon, lat, timezone, angle float64) (time.Time, error) {
	var err error
	tm := basic.GetAsaTime(jde, lon, lat, timezone, angle)
	if tm == -2 {
		err = errors.New("不存在")
	}
	if tm == -1 {
		err = errors.New("不存在")
	}
	return JDE2Date(tm), err
}

// EveningTwilightTime 昏朦影
// jde，世界时当地0时 JDE
// lon，经度，东正西负
// lat，纬度，北正南负
// timezone，时区，东正西负
// angle，朦影角度：可选-6 -12 -18(民用、航海、天文)
func EveningTwilightTime(jde, lon, lat, timezone, angle float64) (time.Time, error) {
	var err error
	tm := basic.GetBanTime(jde, lon, lat, timezone, angle)
	if tm == -2 {
		err = errors.New("不存在")
	}
	if tm == -1 {
		err = errors.New("不存在")
	}
	return JDE2Date(tm), err
}
