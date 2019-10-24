package moon

import (
	"errors"

	"github.com/starainrt/astro/basic"
)

/*
	月亮真黄经
	jde，世界时UTC
*/
func TrueLo(jde float64) float64 {
	return basic.HMoonTrueLo(basic.TD2UT(jde, true))
}

/*
	月亮视黄经
	jde，世界时UTC
*/
func SeeLo(jde float64) float64 {
	return basic.HMoonSeeLo(basic.TD2UT(jde, true))
}

/*
	月亮视赤经
	jde，世界时UTC
	lon, 经度
	lat, 纬度
	timezone, 时区
	返回站心坐标
*/
func SeeRa(jde, lon, lat, timezone float64) float64 {
	return basic.HMoonSeeRa(basic.TD2UT(jde, true), lon, lat, timezone)
}

/*
	月亮视赤纬
	jde，世界时UTC
	lon, 经度
	lat, 纬度
	timezone, 时区
	返回站心坐标
*/
func SeeDec(jde, lon, lat, timezone float64) float64 {
	return basic.HMoonSeeDec(basic.TD2UT(jde, true), lon, lat, timezone)
}

/*
	月亮视赤经赤纬
	jde，世界时UTC
	lon, 经度
	lat, 纬度
	timezone, 时区
	返回站心坐标
*/
func SeeRaDec(jde, lon, lat, timezone float64) (float64, float64) {
	return basic.HMoonSeeRaDec(basic.TD2UT(jde, true), lon, lat, timezone)
}

/*
	月亮真赤经
	jde，世界时UTC
*/
func TrueRa(jde float64) float64 {
	return basic.HMoonTrueRa(basic.TD2UT(jde, true))
}

/*
	月亮真赤纬
	jde，世界时UTC
*/
func TrueDec(jde float64) float64 {
	return basic.HMoonTrueDec(basic.TD2UT(jde, true))
}

/*
  月亮时角
  jde，世界时当地时间
  lon，经度，东正西负
  lat，纬度，北正南负
  timezone，时区，东正西负
*/
func HourAngle(jde, lon, lat, timezone float64) float64 {
	return basic.MoonTimeAngle(jde, lon, lat, timezone)
}

/*
  月亮方位角
  jde，世界时当地时间
  lon，经度，东正西负
  lat，纬度，北正南负
  timezone，时区，东正西负
*/
func Azimuth(jde, lon, lat, timezone float64) float64 {
	return basic.HMoonAngle(jde, lon, lat, timezone)
}

/*
  月亮高度角
  jde，世界时当地时间
  lon，经度，东正西负
  lat，纬度，北正南负
  timezone，时区，东正西负
*/
func Zenith(jde, lon, lat, timezone float64) float64 {
	return basic.HMoonHeight(jde, lon, lat, timezone)
}

/*
  月亮中天时间
  jde，世界时当地0时
  lon，经度，东正西负
	lat，纬度，北正南负
  timezone，时区，东正西负
*/
func CulminationTime(jde, lon, lat, timezone float64) float64 {
	return basic.GetMoonTZTime(jde, lon, lat, timezone)
}

/*
  月亮升起时间
  jde，世界时当地0时 JDE
  lon，经度，东正西负
  lat，纬度，北正南负
  timezone，时区，东正西负
  aero，true时进行大气修正
*/
func RiseTime(jde, lon, lat, timezone float64, aero bool) (float64, error) {
	var err error = nil
	tz := 0.00
	if aero {
		tz = 1
	}
	tm := basic.GetMoonRiseTime(jde, lon, lat, timezone, tz)
	if tm == -2 {
		err = errors.New("极夜")
	}
	if tm == -1 {
		err = errors.New("极昼")
	}
	return tm, err
}

/*
  月亮落下时间
  jde，世界时当地0时 JDE
  lon，经度，东正西负
  lat，纬度，北正南负
  timezone，时区，东正西负
  aero，true时进行大气修正
*/
func DownTime(jde, lon, lat, timezone float64, aero bool) (float64, error) {
	var err error = nil
	tz := 0.00
	if aero {
		tz = 1
	}
	tm := basic.GetMoonDownTime(jde, lon, lat, timezone, tz)
	if tm == -2 {
		err = errors.New("极夜")
	}
	if tm == -1 {
		err = errors.New("极昼")
	}
	return tm, err
}

/*
月相
jde，世界时UTC JDE
*/
func Phase(jde float64) float64 {
	return basic.MoonLight(basic.TD2UT(jde, true))
}

/*
朔月
*/
func ShuoYue(year float64) float64 {
	return basic.CalcMoonSH(year, 0)
}

/*
望月
*/
func WangYue(year float64) float64 {
	return basic.CalcMoonSH(year, 1)
}

/*
上弦月
*/
func ShangXianYue(year float64) float64 {
	return basic.CalcMoonXH(year, 0)
}

/*
下弦月
*/
func XiaXianYue(year float64) float64 {
	return basic.CalcMoonXH(year, 1)
}
