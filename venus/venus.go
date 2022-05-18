package venus

import (
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
	"github.com/starainrt/astro/planet"
	"errors"
	"time"
)

var (
	ERR_VENUS_NEVER_RISE = errors.New("ERROR:极夜，金星今日永远在地平线下！")
	ERR_VENUS_NEVER_DOWN = errors.New("ERROR:极昼，金星今日永远在地平线上！")
)

// ApparentLo 视黄经
func ApparentLo(date time.Time) float64 {
	jde := calendar.Date2JDE(date)
	return basic.VenusApparentLo(basic.TD2UT(jde, true))
}

// ApparentBo 视黄纬
func ApparentBo(date time.Time) float64 {
	jde := calendar.Date2JDE(date)
	return basic.VenusApparentBo(basic.TD2UT(jde, true))
}

// ApparentRa 视赤经
func ApparentRa(date time.Time) float64 {
	jde := calendar.Date2JDE(date)
	return basic.VenusApparentRa(basic.TD2UT(jde, true))
}

// ApparentDec 视赤纬
func ApparentDec(date time.Time) float64 {
	jde := calendar.Date2JDE(date)
	return basic.VenusApparentDec(basic.TD2UT(jde, true))
}

// ApparentRaDec 视赤经赤纬
func ApparentRaDec(date time.Time) (float64, float64) {
	jde := calendar.Date2JDE(date)
	return basic.VenusApparentRaDec(basic.TD2UT(jde, true))
}

// ApparentMagnitude 视星等
func ApparentMagnitude(date time.Time) float64 {
	jde := calendar.Date2JDE(date)
	return basic.VenusMag(basic.TD2UT(jde, true))
}

// EarthDistance 与地球距离（天文单位）
func EarthDistance(date time.Time) float64 {
	jde := calendar.Date2JDE(date)
	return basic.EarthVenusAway(basic.TD2UT(jde, true))
}

// EarthDistance 与太阳距离（天文单位）
func SunDistance(date time.Time) float64 {
	jde := calendar.Date2JDE(date)
	return planet.WherePlanet(2, 2, basic.TD2UT(jde, true))
}

// Zenith 高度角
func Zenith(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.VenusHeight(jde, lon, lat, timezone)
}

// Azimuth 方位角
func Azimuth(date time.Time, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.VenusAzimuth(jde, lon, lat, timezone)
}

// HourAngle 时角
// 返回给定经纬度、对应date时区date时刻的时角（
func HourAngle(date time.Time, lon float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.VenusHourAngle(jde, lon, timezone)
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
	calcJde := basic.VenusCulminationTime(jde, lon, timezone) - timezone/24.00
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
	riseJde := basic.VenusRiseTime(jde, lon, lat, timezone, aeroFloat, height)
	if riseJde == -2 {
		err = ERR_VENUS_NEVER_RISE
	}
	if riseJde == -1 {
		err = ERR_VENUS_NEVER_DOWN
	}
	return basic.JDE2DateByZone(riseJde, date.Location(), true), err
}

// DownTime 落下时间
// date，取日期，时区忽略
// lon，经度，东正西负
// lat，纬度，北正南负
// height，高度
// aero，true时进行大气修正
func DownTime(date time.Time, lon, lat, height float64, aero bool) (time.Time, error) {
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
	riseJde := basic.VenusDownTime(jde, lon, lat, timezone, aeroFloat, height)
	if riseJde == -2 {
		err = ERR_VENUS_NEVER_RISE
	}
	if riseJde == -1 {
		err = ERR_VENUS_NEVER_DOWN
	}
	return basic.JDE2DateByZone(riseJde, date.Location(), true), err
}

// LastConjunction 上次合日时间
// 返回上次合日时间，不区分上合下合
func LastConjunction(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastVenusConjunction(jde), date.Location(), false)
}

// NextConjunction 下次合日时间
// 返回下次合日时间，不区分上合下合
func NextConjunction(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextVenusConjunction(jde), date.Location(), false)
}

// LastInferiorConjunction 上次下合时间
// 返回上次下合日时间
func LastInferiorConjunction(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastVenusInferiorConjunction(jde), date.Location(), false)
}

// NextInferiorConjunction 下次下合时间
// 返回下次合日时间
func NextInferiorConjunction(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextVenusInferiorConjunction(jde), date.Location(), false)
}

// LastSuperiorConjunction 上次上合时间
// 返回上次下合时间
func LastSuperiorConjunction(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastVenusSuperiorConjunction(jde), date.Location(), false)
}

// NextSuperiorConjunction 下次上合时间
// 返回下次上合时间
func NextSuperiorConjunction(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextVenusSuperiorConjunction(jde), date.Location(), false)
}

// LastRetrograde 上次留的时间
// 返回上次留时间，不区分顺逆
func LastRetrograde(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastVenusRetrograde(jde), date.Location(), false)
}

// NextRetrograde 下次留时间
// 返回下次留的时间，不区分顺逆
func NextRetrograde(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextVenusRetrograde(jde), date.Location(), false)
}

// LastProgradeToRetrograde 上次留（顺转逆）
// 返回上次顺转逆留的时间
func LastProgradeToRetrograde(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastVenusProgradeToRetrograde(jde), date.Location(), false)
}

// NextProgradeToRetrograde 下次留（顺转逆）
// 返回下次顺转逆留的时间
func NextProgradeToRetrograde(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextVenusProgradeToRetrograde(jde), date.Location(), false)
}

// LastRetrogradeToPrograde 上次留（逆转瞬）
// 返回上次逆转瞬留的时间
func LastRetrogradeToPrograde(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastVenusRetrogradeToPrograde(jde), date.Location(), false)
}

// NextRetrogradeToPrograde 上次留（逆转瞬）
//// 返回上次逆转瞬留的时间
func NextRetrogradeToPrograde(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextVenusRetrogradeToPrograde(jde), date.Location(), false)
}

// LastGreatestElongation 上次大距时间
// 返回上次大距时间，不区分东西大距
func LastGreatestElongation(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastVenusGreatestElongation(jde), date.Location(), false)
}

// NextGreatestElongation  下次大距时间
// 返回下次大距时间，不区分东西大距
func NextGreatestElongation(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextVenusGreatestElongation(jde), date.Location(), false)
}

// LastGreatestElongationEast 上次东大距时间
// 返回上次东大距时间
func LastGreatestElongationEast(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastVenusGreatestElongationEast(jde), date.Location(), false)
}

// NextGreatestElongationEast 下次东大距时间
// 返回下次东大距时间
func NextGreatestElongationEast(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextVenusGreatestElongationEast(jde), date.Location(), false)
}

// LastGreatestElongationWest 上次西大距时间
// 返回上次西大距时间
func LastGreatestElongationWest(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.LastVenusGreatestElongationWest(jde), date.Location(), false)
}

// NextGreatestElongationWest 下次西大距时间
// 返回下次西大距时间
func NextGreatestElongationWest(date time.Time) time.Time {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.JDE2DateByZone(basic.NextVenusGreatestElongationWest(jde), date.Location(), false)
}
