package sun

import (
	"errors"
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

var (
	ERR_SUN_NEVER_RISE      = errors.New("ERROR:极夜，太阳在今日永远在地平线下！")
	ERR_SUN_NEVER_DOWN      = errors.New("ERROR:极昼，太阳在今日永远在地平线上！")
	ERR_TWILIGHT_NOT_EXISTS = errors.New("ERROR:今日晨昏朦影不存在！")
)

/*
太阳
视星等	−26.74
绝对星等	4.839
光谱类型	G2V
金属量	Z = 0.0122
角直径	31.6′ – 32.7′
*/

// RiseTime 太阳升起时间
// date，取日期，时区忽略
// lon，经度，东正西负
// lat，纬度，北正南负
// timezone，时区，东正西负
// aero，true时进行大气修正
func RiseTime(date time.Time, lon, lat, timezone float64, aero bool) (time.Time, error) {
	var err error
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	jde := basic.Date2JDE(date)
	riseJde := basic.GetSunRiseTime(jde, lon, lat, timezone, aeroFloat)
	if riseJde == -2 {
		err = ERR_SUN_NEVER_RISE
	}
	if riseJde == -1 {
		err = ERR_SUN_NEVER_DOWN
	}
	return basic.JDE2Date(riseJde), err
}

// SunDownTime 太阳落下时间
// date，取日期，时区忽略
// lon，经度，东正西负
// lat，纬度，北正南负
// timezone，时区，东正西负
// aero，true时进行大气修正
func DownTime(date time.Time, lon, lat, timezone float64, aero bool) (time.Time, error) {
	var err error
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	jde := basic.Date2JDE(date)
	downJde := basic.GetSunDownTime(jde, lon, lat, timezone, aeroFloat)
	if downJde == -2 {
		err = ERR_SUN_NEVER_RISE
	}
	if downJde == -1 {
		err = ERR_SUN_NEVER_DOWN
	}
	return basic.JDE2Date(downJde), err
}

// MorningTwilight 晨朦影
// date，取日期，时区忽略
// lon，经度，东正西负
// lat，纬度，北正南负
// timezone，时区，东正西负
// angle，朦影角度：可选-6 -12 -18(民用、航海、天文)
func MorningTwilight(date time.Time, lon, lat, timezone, angle float64) (time.Time, error) {
	var err error
	jde := basic.Date2JDE(date)
	calcJde := basic.GetAsaTime(jde, lon, lat, timezone, angle)
	if calcJde == -2 {
		err = ERR_TWILIGHT_NOT_EXISTS
	}
	if calcJde == -1 {
		err = ERR_TWILIGHT_NOT_EXISTS
	}
	return basic.JDE2Date(calcJde), err
}

// EveningTwilight 昏朦影
// date，取日期，时区忽略
// lon，经度，东正西负
// lat，纬度，北正南负
// timezone，时区，东正西负
// angle，朦影角度：可选-6 -12 -18(民用、航海、天文)
func EveningTwilight(date time.Time, lon, lat, timezone, angle float64) (time.Time, error) {
	var err error
	jde := basic.Date2JDE(date)
	//不需要进行力学时转换，会在GetBanTime中转换
	calcJde := basic.GetBanTime(jde, lon, lat, timezone, angle)
	if calcJde == -2 {
		err = ERR_TWILIGHT_NOT_EXISTS
	}
	if calcJde == -1 {
		err = ERR_TWILIGHT_NOT_EXISTS
	}
	return basic.JDE2Date(calcJde), err
}

// EclipticObliquity 黄赤交角
// 返回date对应UTC世界时的黄赤交角，nutation为true时，计算交角章动
func EclipticObliquity(date time.Time, nutation bool) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	//进行力学时转换
	jde = basic.TD2UT(jde, true)
	//黄赤交角计算
	return basic.EclipticObliquity(jde, nutation)
}

// EclipticNutation 黄经章动
// 返回date对应UTC世界时的黄经章动
func EclipticNutation(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	//进行力学时转换与章动计算
	return basic.HJZD(basic.TD2UT(jde, true))
}

// AxialtiltNutation 交角章动
// 返回date对应UTC世界时的交角章动
func AxialtiltNutation(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	//进行力学时转换与章动计算
	return basic.JJZD(basic.TD2UT(jde, true))
}

// GeometricLo 太阳几何黄经
// 返回date对应UTC世界时的太阳几何黄经
func GeometricLo(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.SunLo(basic.TD2UT(jde, true))
}

// TrueLo 太阳真黄经
// 返回date对应UTC世界时的太阳真黄经
func TrueLo(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.HSunTrueLo(basic.TD2UT(jde, true))
}

// TrueBo 太阳真黄纬
// 返回date对应UTC世界时的太阳真黄纬
func TrueBo(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.HSunTrueLo(basic.TD2UT(jde, true))
}

// SeeLo 太阳视黄经
// 返回date对应UTC世界时的太阳视黄经
func SeeLo(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.HSunSeeLo(basic.TD2UT(jde, true))
}

// SeeRa 太阳地心视赤经
// 返回date对应UTC世界时的太阳视赤经（使用黄道坐标转换，且默认忽略黄纬）
func SeeRa(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.HSunSeeRa(basic.TD2UT(jde, true))
}

// SeeDec 太阳地心视赤纬
// 返回date对应UTC世界时的太阳视赤纬（使用黄道坐标转换，且默认忽略黄纬）
func SeeDec(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.HSunSeeDec(basic.TD2UT(jde, true))
}

// SeeRaDec 太阳地心视赤经和赤纬
// 返回date对应UTC世界时的太阳视赤纬（使用黄道坐标转换，且默认忽略黄纬）
func SeeRaDec(date time.Time) (float64, float64) {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.HSunSeeRaDec(basic.TD2UT(jde, true))
}

// MidFunc 太阳中间方程
func MidFunc(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.SunMidFun(basic.TD2UT(jde, true))
}

// EquationTime 均时差
// 返回date对应UTC世界时的均时差
func EquationTime(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.SunTime(basic.TD2UT(jde, true))
}

// HourAngle 太阳时角
// 返回给定经纬度、对应timezone时区date时刻的太阳时角（注意，date本身的时区将默认舍去）
func HourAngle(date time.Time, lon, lat, timezone float64) float64 {
	jde := basic.Date2JDE(date)
	return basic.SunTimeAngle(jde, lon, lat, timezone)
}

// Azimuth 太阳方位角
// 返回给定经纬度、对应timezone时区date时刻的太阳方位角（正北为0，向东增加）
//（注意，date本身的时区将默认舍去）
func Azimuth(date time.Time, lon, lat, timezone float64) float64 {
	jde := basic.Date2JDE(date)
	return basic.SunAngle(jde, lon, lat, timezone)
}

// Zenith 太阳高度角
// 返回给定经纬度、对应timezone时区date时刻的太阳高度角
//（注意，date本身的时区将默认舍去）
func Zenith(date time.Time, lon, lat, timezone float64) float64 {
	jde := basic.Date2JDE(date)
	return basic.SunHeight(jde, lon, lat, timezone)
}

// CulminationTime 太阳中天时间
// 返回给定经纬度、对应timezone时区date时刻的太阳中天日期
//（注意，date本身的时区将默认舍去，返回的时间时区应当为传入的timezone）
func CulminationTime(date time.Time, lon, timezone float64) time.Time {
	jde := basic.Date2JDE(date)
	if jde-math.Floor(jde) > 0.5 {
		jde++
	}
	calcJde := basic.GetSunTZTime(jde, lon, timezone)
	return basic.JDE2Date(calcJde)
}

// EarthDistance 日地距离
// 返回date对应UTC世界时日地距离
func EarthDistance(date time.Time) float64 {
	jde := basic.Date2JDE(date)
	jde = basic.TD2UT(jde, true)
	return basic.EarthAway(jde)
}
