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
// height，高度
// aero，true时进行大气修正
func RiseTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	var err error
	var aeroFloat float64
	if aero {
		aeroFloat = 1
	}
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	//忽略时区的字面量时间
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	//risedate 时区修正后的时间，转换应当包括时区
	riseJde := basic.GetSunRiseTime(jde, lon, lat, timezone, aeroFloat, height)
	if riseJde == -2 {
		err = ERR_SUN_NEVER_RISE
	}
	if riseJde == -1 {
		err = ERR_SUN_NEVER_SET
	}
	return basic.JDE2DateByZone(riseJde, date.Location(), true), err
}

// deprecated: -- use SetTime instead
// DownTime 太阳落下时间
// date，当地时区日期，务必做时区修正
// lon，经度，东正西负
// lat，纬度，北正南负
// height，高度
// aero，true时进行大气修正
func DownTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return SetTime(date, lon, lat, height, aero)
}

// SetTime 太阳落下时间
// date，当地时区日期，务必做时区修正
// lon，经度，东正西负
// lat，纬度，北正南负
// height，高度
// aero，true时进行大气修正
func SetTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	var err error
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
	downJde := basic.GetSunSetTime(jde, lon, lat, timezone, aeroFloat, height)
	if downJde == -2 {
		err = ERR_SUN_NEVER_RISE
	}
	if downJde == -1 {
		err = ERR_SUN_NEVER_SET
	}
	return basic.JDE2DateByZone(downJde, date.Location(), true), err
}

// MorningTwilight 晨朦影
// date，当地时区日期，返回的时间时区与此参数中的时区一致
// lon，经度，东正西负
// lat，纬度，北正南负
// angle，朦影角度：可选-6 -12 -18(民用、航海、天文)
func MorningTwilight(date time.Time, lon, lat, angle float64) (time.Time, error) {
	var err error
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJde := basic.MorningTwilight(jde, lon, lat, timezone, angle)
	if calcJde == -2 {
		err = ERR_TWILIGHT_NOT_EXISTS
	}
	if calcJde == -1 {
		err = ERR_TWILIGHT_NOT_EXISTS
	}
	return basic.JDE2DateByZone(calcJde, date.Location(), true), err
}

// EveningTwilight 昏朦影
// date，当地时区日期，返回的时间时区与此参数中的时区一致
// lon，经度，东正西负
// lat，纬度，北正南负
// angle，朦影角度：可选-6 -12 -18(民用、航海、天文)
func EveningTwilight(date time.Time, lon, lat, angle float64) (time.Time, error) {
	var err error
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	//不需要进行力学时转换，会在GetBanTime中转换
	calcJde := basic.EveningTwilight(jde, lon, lat, timezone, angle)
	if calcJde == -2 {
		err = ERR_TWILIGHT_NOT_EXISTS
	}
	if calcJde == -1 {
		err = ERR_TWILIGHT_NOT_EXISTS
	}
	return basic.JDE2DateByZone(calcJde, date.Location(), true), err
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

// EclipticNutation 黄经章动(2000b)
// 返回date对应UTC世界时的黄经章动
func EclipticNutation(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	//进行力学时转换与章动计算
	return basic.Nutation2000Bi(basic.TD2UT(jde, true))
}

// EclipticNutation1980 黄经章动(iau 1980)
// 返回date对应UTC世界时的黄经章动
func EclipticNutation1980(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	//进行力学时转换与章动计算
	return basic.Nutation2000Bi(basic.TD2UT(jde, true))
}

// AxialtiltNutation 交角章动(2000b)
// 返回date对应UTC世界时的交角章动
func AxialtiltNutation(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	//进行力学时转换与章动计算
	return basic.Nutation2000Bs(basic.TD2UT(jde, true))
}

// AxialtiltNutation1980 交角章动(1980)
// 返回date对应UTC世界时的交角章动
func AxialtiltNutation1980(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	//进行力学时转换与章动计算
	return basic.Nutation1980s(basic.TD2UT(jde, true))
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
	return basic.HSunTrueBo(basic.TD2UT(jde, true))
}

// ApparentLo 太阳视黄经
// 返回date对应UTC世界时的太阳视黄经
func ApparentLo(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.HSunApparentLo(basic.TD2UT(jde, true))
}

// ApparentRa 太阳地心视赤经
// 返回date对应UTC世界时的太阳视赤经（使用黄道坐标转换，且默认忽略黄纬）
func ApparentRa(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.HSunApparentRa(basic.TD2UT(jde, true))
}

// ApparentDec 太阳地心视赤纬
// 返回date对应UTC世界时的太阳视赤纬（使用黄道坐标转换，且默认忽略黄纬）
func ApparentDec(date time.Time) float64 {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.HSunApparentDec(basic.TD2UT(jde, true))
}

// ApparentRaDec 太阳地心视赤经和赤纬
// 返回date对应UTC世界时的太阳视赤纬（使用黄道坐标转换，且默认忽略黄纬）
func ApparentRaDec(date time.Time) (float64, float64) {
	//转换为UTC时间
	jde := basic.Date2JDE(date.UTC())
	return basic.HSunApparentRaDec(basic.TD2UT(jde, true))
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
// 返回给定经纬度、对应date时区date时刻的太阳时角（
func HourAngle(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SunTimeAngle(jde, lon, lat, timezone)
}

// Azimuth 太阳方位角
// 返回给定经纬度、对应date时区date时刻的太阳方位角（正北为0，向东增加）
func Azimuth(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SunAngle(jde, lon, lat, timezone)
}

// Zenith 太阳高度角
// 返回给定经纬度、对应date时区date时刻的太阳高度角
func Zenith(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.SunHeight(jde, lon, lat, timezone)
}

// CulminationTime 太阳中天时间
// 返回给定经纬度、对应date时区date时刻的太阳中天日期
func CulminationTime(date time.Time, lon float64) time.Time {
	jde := basic.Date2JDE(date.Add(time.Duration(-1*date.Hour())*time.Hour)) + 0.5
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJde := basic.CulminationTime(jde, lon, timezone) - timezone/24.00
	return basic.JDE2DateByZone(calcJde, date.Location(), false)
}

// EarthDistance 日地距离
// 返回date对应UTC世界时日地距离
func EarthDistance(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	jde = basic.TD2UT(jde, true)
	return basic.EarthAway(jde)
}

// ApparentSolarTime 真太阳时
// 返回给定经度lon的真太阳时
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
