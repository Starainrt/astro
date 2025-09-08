package neptune

import (
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
	"github.com/starainrt/astro/planet"
	"errors"
	"time"
)

var (
	ERR_NEPTUNE_NEVER_RISE = errors.New("ERROR:极夜，海王星今日永远在地平线下！")
	ERR_NEPTUNE_NEVER_SET  = errors.New("ERROR:极昼，海王星今日永远在地平线上！")
	// ERR_NEPTUNE_NEVER_DOWN deprecated: -- use ERR_NEPTUNE_NEVER_SET instead
	ERR_NEPTUNE_NEVER_DOWN = ERR_NEPTUNE_NEVER_SET
)

// ApparentLo 视黄经
func ApparentLo(date time.Time) float64 {
	jde := calendar.Date2JDE(date)
	return basic.NeptuneApparentLo(basic.TD2UT(jde, true))
}

// ApparentBo 视黄纬
func ApparentBo(date time.Time) float64 {
	jde := calendar.Date2JDE(date)
	return basic.NeptuneApparentBo(basic.TD2UT(jde, true))
}

// ApparentRa 视赤经
func ApparentRa(date time.Time) float64 {
	jde := calendar.Date2JDE(date)
	return basic.NeptuneApparentRa(basic.TD2UT(jde, true))
}

// ApparentDec 视赤纬
func ApparentDec(date time.Time) float64 {
	jde := calendar.Date2JDE(date)
	return basic.NeptuneApparentDec(basic.TD2UT(jde, true))
}

// ApparentRaDec 视赤经赤纬
func ApparentRaDec(date time.Time) (float64, float64) {
	jde := calendar.Date2JDE(date)
	return basic.NeptuneApparentRaDec(basic.TD2UT(jde, true))
}

// ApparentMagnitude 视星等
func ApparentMagnitude(date time.Time) float64 {
	jde := calendar.Date2JDE(date)
	return basic.NeptuneMag(basic.TD2UT(jde, true))
}

// EarthDistance 与地球距离（天文单位）
func EarthDistance(date time.Time) float64 {
	jde := calendar.Date2JDE(date)
	return basic.EarthNeptuneAway(basic.TD2UT(jde, true))
}

// EarthDistance 与太阳距离（天文单位）
func SunDistance(date time.Time) float64 {
	jde := calendar.Date2JDE(date)
	return planet.WherePlanet(7, 2, basic.TD2UT(jde, true))
}

// Zenith 高度角
func Zenith(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.NeptuneHeight(jde, lon, lat, timezone)
}

// Azimuth 方位角
func Azimuth(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.NeptuneAzimuth(jde, lon, lat, timezone)
}

// HourAngle 时角
// 返回给定经纬度、对应date时区date时刻的时角（
func HourAngle(date time.Time, lon float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.NeptuneHourAngle(jde, lon, timezone)
}

// CulminationTime 中天时间
// 返回给定经纬度、对应date时区date时刻的中天日期
func CulminationTime(date time.Time, lon float64) time.Time {
	if date.Hour() > 12 {
		date = date.Add(time.Hour * -12)
	}
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	calcJde := basic.NeptuneCulminationTime(jde, lon, timezone) - timezone/24.00
	return basic.JDE2DateByZone(calcJde, date.Location(), false)
}

// RiseTime 升起时间
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
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	riseJde := basic.NeptuneRiseTime(jde, lon, lat, timezone, aeroFloat, height)
	if riseJde == -2 {
		err = ERR_NEPTUNE_NEVER_RISE
	}
	if riseJde == -1 {
		err = ERR_NEPTUNE_NEVER_SET
	}
	return basic.JDE2DateByZone(riseJde, date.Location(), true), err
}

// deprecated: -- use SetTime instead
// DownTime 落下时间
// date，取日期，时区忽略
// lon，经度，东正西负
// lat，纬度，北正南负
// height，高度
// aero，true时进行大气修正
func DownTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
	return SetTime(date, lon, lat, height, aero)
}

// SetTime 落下时间
// date，取日期，时区忽略
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
	riseJde := basic.NeptuneDownTime(jde, lon, lat, timezone, aeroFloat, height)
	if riseJde == -2 {
		err = ERR_NEPTUNE_NEVER_RISE
	}
	if riseJde == -1 {
		err = ERR_NEPTUNE_NEVER_SET
	}
	return basic.JDE2DateByZone(riseJde, date.Location(), true), err
}

// LastConjunction 上次合日时间
// 返回上次合日时间
func LastConjunction(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastNeptuneConjunction(jde), date.Location(), false)
}

// NextConjunction 下次合日时间
// 返回下次合日时间
func NextConjunction(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextNeptuneConjunction(jde), date.Location(), false)
}

// LastOpposition 上次冲日时间
// 返回上次冲日时间
func LastOpposition(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastNeptuneOpposition(jde), date.Location(), false)
}

// NextOpposition 下次冲日时间
// 返回下次冲日时间
func NextOpposition(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextNeptuneOpposition(jde), date.Location(), false)
}

// LastProgradeToRetrograde 上次留（顺转逆）
// 返回上次顺转逆留的时间
func LastProgradeToRetrograde(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastNeptuneProgradeToRetrograde(jde), date.Location(), false)
}

// NextProgradeToRetrograde 下次留（顺转逆）
// 返回下次顺转逆留的时间
func NextProgradeToRetrograde(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextNeptuneProgradeToRetrograde(jde), date.Location(), false)
}

// LastRetrogradeToPrograde 上次留（逆转瞬）
// 返回上次逆转瞬留的时间
func LastRetrogradeToPrograde(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastNeptuneRetrogradeToPrograde(jde), date.Location(), false)
}

// NextRetrogradeToPrograde 上次留（逆转瞬）
// // 返回上次逆转瞬留的时间
func NextRetrogradeToPrograde(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextNeptuneRetrogradeToPrograde(jde), date.Location(), false)
}

// LastEasternQuadrature 上次东方照时间
// 返回上次东方照时间
func LastEasternQuadrature(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastNeptuneEasternQuadrature(jde), date.Location(), false)
}

// NextEasternQuadrature 下次东方照时间
// 返回下次东方照时间
func NextEasternQuadrature(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextNeptuneEasternQuadrature(jde), date.Location(), false)
}

// LastWesternQuadrature 上次西方照时间
// 返回上次西方照时间
func LastWesternQuadrature(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastNeptuneWesternQuadrature(jde), date.Location(), false)
}

// NextWesternQuadrature 下次西方照时间
// 返回下次西方照时间
func NextWesternQuadrature(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextNeptuneWesternQuadrature(jde), date.Location(), false)
}
