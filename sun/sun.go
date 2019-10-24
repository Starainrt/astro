package sun

import (
	"errors"

	"github.com/starainrt/astro/basic"
)

/*
   黄赤交角
	jde，世界时UTC
	nutation，为true时计算交角章动
*/
func EclipticObliquity(jde float64, nutation bool) float64 {
	jde = basic.TD2UT(jde, true)
	return basic.EclipticObliquity(jde, nutation)
}

/*
    黄经章动
	jde，世界时UTC
*/
func EclipticNutation(jde float64) float64 {
	return basic.HJZD(basic.TD2UT(jde, true))
}

/*
    交角章动
	jde，世界时UTC
*/
func AxialtiltNutation(jde float64) float64 {
	return basic.JJZD(basic.TD2UT(jde, true))
}

/*
	太阳几何黄经
	jde，世界时UTC
*/
func GeometricLo(jde float64) float64 {
	return basic.SunLo(basic.TD2UT(jde, true))
}

/*
	太阳真黄经
	jde，世界时UTC
*/
func TrueLo(jde float64) float64 {
	return basic.HSunTrueLo(basic.TD2UT(jde, true))
}

/*
	太阳视黄经
	jde，世界时UTC
*/
func SeeLo(jde float64) float64 {
	return basic.HSunSeeLo(basic.TD2UT(jde, true))
}

/*
	太阳视赤经
	jde，世界时UTC
*/
func SeeRa(jde float64) float64 {
	return basic.HSunSeeRa(basic.TD2UT(jde, true))
}

/*
	太阳视赤纬
	jde，世界时UTC
*/
func SeeDec(jde float64) float64 {
	return basic.HSunSeeDec(basic.TD2UT(jde, true))
}

/*
	太阳视赤经赤纬
	jde，世界时UTC
*/
func SeeRaDec(jde float64) (float64, float64) {
	return basic.HSunSeeRaDec(basic.TD2UT(jde, true))
}

/*
	太阳真赤经
	jde，世界时UTC
*/
func TrueRa(jde float64) float64 {
	return basic.HSunTrueRa(basic.TD2UT(jde, true))
}

/*
	太阳真赤纬
	jde，世界时UTC
*/
func TrueDec(jde float64) float64 {
	return basic.HSunTrueDec(basic.TD2UT(jde, true))
}

/*
	太阳中间方程
	jde，世界时UTC
*/
func MidFunc(jde float64) float64 {
	return basic.SunMidFun(basic.TD2UT(jde, true))
}

/*
	均时差
	jde，世界时UTC
*/
func EquationTime(jde float64) float64 {
	return basic.SunTime(basic.TD2UT(jde, true))
}

/*
  太阳时角
  jde，世界时当地时间
  lon，经度，东正西负
  lat，纬度，北正南负
  timezone，时区，东正西负
*/
func HourAngle(jde, lon, lat, timezone float64) float64 {
	return basic.SunTimeAngle(jde, lon, lat, timezone)
}

/*
  太阳方位角
  jde，世界时当地时间
  lon，经度，东正西负
  lat，纬度，北正南负
  timezone，时区，东正西负
*/
func Azimuth(jde, lon, lat, timezone float64) float64 {
	return basic.SunAngle(jde, lon, lat, timezone)
}

/*
  太阳高度角
  jde，世界时当地时间
  lon，经度，东正西负
  lat，纬度，北正南负
  timezone，时区，东正西负
*/
func Zenith(jde, lon, lat, timezone float64) float64 {
	return basic.SunHeight(jde, lon, lat, timezone)
}

/*
  太阳中天时间
  jde，世界时当地0时
  lon，经度，东正西负
  timezone，时区，东正西负
*/
func CulminationTime(jde, lon, timezone float64) float64 {
	return basic.GetSunTZTime(jde, lon, timezone)
}

/*
  太阳升起时间
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
	tm := basic.GetSunRiseTime(jde, lon, lat, timezone, tz)
	if tm == -2 {
		err = errors.New("极夜")
	}
	if tm == -1 {
		err = errors.New("极昼")
	}
	return tm, err
}

/*
  太阳落下时间
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
	tm := basic.GetSunDownTime(jde, lon, lat, timezone, tz)
	if tm == -2 {
		err = errors.New("极夜")
	}
	if tm == -1 {
		err = errors.New("极昼")
	}
	return tm, err
}

/*
  晨朦影
  jde，世界时当地0时 JDE
  lon，经度，东正西负
  lat，纬度，北正南负
  timezone，时区，东正西负
  angle，朦影角度：可选-6 -12 -18
*/
func MorningTwilightTime(jde, lon, lat, timezone, angle float64) (float64, error) {
	var err error = nil
	tm := basic.GetAsaTime(jde, lon, lat, timezone, angle)
	if tm == -2 {
		err = errors.New("不存在")
	}
	if tm == -1 {
		err = errors.New("不存在")
	}
	return tm, err
}

/*
  昏朦影
  jde，世界时当地0时 JDE
  lon，经度，东正西负
  lat，纬度，北正南负
  timezone，时区，东正西负
  angle，朦影角度：可选-6 -12 -18
*/
func EveningTwilightTime(jde, lon, lat, timezone, angle float64) (float64, error) {
	var err error = nil
	tm := basic.GetBanTime(jde, lon, lat, timezone, angle)
	if tm == -2 {
		err = errors.New("不存在")
	}
	if tm == -1 {
		err = errors.New("不存在")
	}
	return tm, err
}
