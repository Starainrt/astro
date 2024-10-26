package moon

import (
	"errors"
	"math"
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

// ApparentLo 月亮视黄经（地心）
// 传入UTC对应的儒略日时间
func ApparentLo(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.HMoonApparentLo(basic.TD2UT(jde, true))
}

// TrueRa 月亮视赤经（地心）
// date, 时间
// 返回地心坐标
func TrueRa(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.HMoonTrueRa(basic.TD2UT(jde, true))
}

// TrueDec 月亮视赤纬（地心）
// date, 时间
// 返回地心坐标
func TrueDec(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.HMoonTrueDec(basic.TD2UT(jde, true))
}

// TrueRaDec 月亮视赤纬赤纬（地心）
// date, 时间
// 返回地心坐标
func TrueRaDec(date time.Time) (float64, float64) {
	jde := basic.Date2JDE(date.UTC())
	return basic.HMoonTrueRaDec(basic.TD2UT(jde, true))
}

// ApparentRa 月亮视赤经（站心）
// date, 时间
// lon, 经度
// lat, 纬度
// 返回站心坐标
func ApparentRa(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.HMoonApparentRa(jde, lon, lat, float64(loc)/3600.0)
}

// ApparentDec 月亮视赤纬（站心）
// date, 时间
// lon, 经度
// lat, 纬度
// 返回站心坐标
func ApparentDec(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.HMoonApparentDec(jde, lon, lat, float64(loc)/3600.0)
}

// ApparentRaDec 月亮视赤纬（站心）
// date, 本地时间
// lon, 经度
// lat, 纬度
// 返回站心坐标
func ApparentRaDec(date time.Time, lon, lat float64) (float64, float64) {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.HMoonApparentRaDec(jde, lon, lat, float64(loc)/3600.0)
}

// HourAngle 月亮时角
//
//	date, 世界时（忽略此处时区）
//	lon，经度，东正西负
//	lat，纬度，北正南负
func HourAngle(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.MoonTimeAngle(jde, lon, lat, float64(loc)/3600.0)
}

// Azimuth 月亮方位角
//
//	date, 世界时（忽略此处时区）
//	lon，经度，东正西负
//	lat，纬度，北正南负
func Azimuth(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.HMoonAngle(jde, lon, lat, float64(loc)/3600.0)
}

// Zenith 月亮高度角
//
//	date, 世界时（忽略此处时区）
//	lon，经度，东正西负
//	lat，纬度，北正南负
func Zenith(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.HMoonHeight(jde, lon, lat, float64(loc)/3600.0)
}

// CulminationTime 月亮中天时间
//
//	date, 世界时（忽略此处时区）
//	lon，经度，东正西负
//	lat，纬度，北正南负
func CulminationTime(date time.Time, lon, lat float64) time.Time {
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	return basic.JDE2DateByZone(basic.MoonCulminationTime(jde, lon, lat, float64(loc)/3600.0), date.Location(), true)
}

// RiseTime 月亮升起时间
//
//	date, 世界时（忽略此处时区）
//	lon，经度，东正西负
//	lat，纬度，北正南负
//	height，高度
//	aero,是否进行大气修正
func RiseTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	var err error
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	aeroFloat := 0.00
	if aero {
		aeroFloat = 1
	}
	riseJde := basic.GetMoonRiseTime(jde, lon, lat, timezone, aeroFloat, height)
	if riseJde == -3 {
		err = ERR_NOT_TODAY
	}
	if riseJde == -2 {
		err = ERR_MOON_NEVER_RISE
	}
	if riseJde == -1 {
		err = ERR_MOON_NEVER_DOWN
	}
	return basic.JDE2DateByZone(riseJde, date.Location(), true), err
}

// DownTime 月亮降下时间
//
//	date, 世界时（忽略此处时区）
//	lon，经度，东正西负
//	lat，纬度，北正南负
//	height，高度
//	aero，大气修正
func DownTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	var err error
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	aeroFloat := 0.00
	if aero {
		aeroFloat = 1
	}
	downJde := basic.GetMoonDownTime(jde, lon, lat, timezone, aeroFloat, height)
	if downJde == -3 {
		err = ERR_NOT_TODAY
	}
	if downJde == -2 {
		err = ERR_MOON_NEVER_RISE
	}
	if downJde == -1 {
		err = ERR_MOON_NEVER_DOWN
	}
	return basic.JDE2DateByZone(downJde, date.Location(), true), err
}

// Phase 月相
// 返回Date对应UTC世界时的月相大小
func Phase(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.MoonLight(basic.TD2UT(jde, true))
}

// ShuoYue 朔月
// 返回Date对应UTC世界时的月相大小
func ShuoYue(year float64) time.Time {
	jde := basic.TD2UT(basic.CalcMoonSH(year, 0), false)
	return basic.JDE2DateByZone(jde, time.UTC, false)
}

// NextShuoYue 下次朔月时间
// 返回date之后的下一个朔月时间（UTC时间）
func NextShuoYue(date time.Time) time.Time {
	return nextMoonPhase(date, 0)
}

func LastShuoYue(date time.Time) time.Time {
	return lastMoonPhase(date, 0)
}

func ClosestShuoYue(date time.Time) time.Time {
	return closestMoonPhase(date, 0)
}

func closestMoonPhase(date time.Time, typed int) time.Time {
	//0=shuo 1=wang 2=shangxian 3=xiaxian
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	if typed < 2 {
		return basic.JDE2DateByZone(basic.TD2UT(basic.CalcMoonSHByJDE(jde, typed), false), date.Location(), false)
	}
	return basic.JDE2DateByZone(basic.TD2UT(basic.CalcMoonXHByJDE(jde, typed-2), false), date.Location(), false)
}

func nextMoonPhase(date time.Time, typed int) time.Time {
	//0=shuo 1=wang 2=shangxian 3=xiaxian
	diffCode := 0.00
	switch typed {
	case 1:
		diffCode = 180
	case 2:
		diffCode = 90
	case 3:
		diffCode = 270
	}
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	cost := basic.HMoonApparentLo(jde) - basic.HSunApparentLo(jde) - float64(diffCode)
	for cost < 0 {
		cost += 360
	}
	if cost < 0 && math.Floor(math.Abs(cost)*10000) == 0 {
		cost = 0
	}
	if cost < 240 {
		jde += (240 - cost) / 11.19
	}
	if typed < 2 {
		return basic.JDE2DateByZone(basic.TD2UT(basic.CalcMoonSHByJDE(jde, typed), false), date.Location(), false)
	}
	return basic.JDE2DateByZone(basic.TD2UT(basic.CalcMoonXHByJDE(jde, typed-2), false), date.Location(), false)
}

func lastMoonPhase(date time.Time, typed int) time.Time {
	//0=shuo 1=wang 2=shangxian 3=xiaxian
	diffCode := 0.00
	switch typed {
	case 1:
		diffCode = 180
	case 2:
		diffCode = 90
	case 3:
		diffCode = 270
	}
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	cost := basic.HMoonApparentLo(jde) - basic.HSunApparentLo(jde) - float64(diffCode)
	for cost < 0 {
		cost += 360
	}
	if cost > 0 && math.Floor(math.Abs(cost)*10000) == 0 {
		cost = 360
	}
	if cost > 120 {
		jde -= (cost - 120) / 11.19
	}
	if typed < 2 {
		return basic.JDE2DateByZone(basic.TD2UT(basic.CalcMoonSHByJDE(jde, typed), false), date.Location(), false)
	}
	return basic.JDE2DateByZone(basic.TD2UT(basic.CalcMoonXHByJDE(jde, typed-2), false), date.Location(), false)
}

// WangYue 望月
func WangYue(year float64) time.Time {
	jde := basic.TD2UT(basic.CalcMoonSH(year, 1), false)
	return basic.JDE2DateByZone(jde, time.UTC, false)
}

func NextWangYue(date time.Time) time.Time {
	return nextMoonPhase(date, 1)
}

func LastWangYue(date time.Time) time.Time {
	return lastMoonPhase(date, 1)
}

func ClosestWangYue(date time.Time) time.Time {
	return closestMoonPhase(date, 1)
}

// ShangXianYue 上弦月
func ShangXianYue(year float64) time.Time {
	jde := basic.TD2UT(basic.CalcMoonXH(year, 0), false)
	return basic.JDE2DateByZone(jde, time.UTC, false)
}

func NextShangXianYue(date time.Time) time.Time {
	return nextMoonPhase(date, 2)
}

func LastShangXianYue(date time.Time) time.Time {
	return lastMoonPhase(date, 2)
}

func ClosestShangXianYue(date time.Time) time.Time {
	return closestMoonPhase(date, 2)
}

// XiaXianYue 下弦月
func XiaXianYue(year float64) time.Time {
	jde := basic.TD2UT(basic.CalcMoonXH(year, 1), false)
	return basic.JDE2DateByZone(jde, time.UTC, false)
}

func NextXiaXianYue(date time.Time) time.Time {
	return nextMoonPhase(date, 3)
}

func LastXiaXianYue(date time.Time) time.Time {
	return lastMoonPhase(date, 3)
}

func ClosestXiaXianYue(date time.Time) time.Time {
	return closestMoonPhase(date, 3)
}

// EarthDistance 日地距离
// 返回date对应UTC世界时日地距离
func EarthDistance(date time.Time) float64 {
	jde := basic.Date2JDE(date)
	jde = basic.TD2UT(jde, true)
	return basic.MoonAway(jde)
}
