package sun

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ParallacticAngle 太阳视差角（天顶方向角） / solar parallactic angle.
//
// lon/lat 为观测者经纬度，东正西负、北正南负；返回值为有符号视差角，单位度。
func ParallacticAngle(date time.Time, lon, lat float64) float64 {
	return basic.ParallacticAngleByHourAngle(HourAngle(date, lon, lat), ApparentDec(date), lat)
}

// ParallacticAngleN 截断项太阳视差角（天顶方向角） / truncated solar parallactic angle.
func ParallacticAngleN(date time.Time, lon, lat float64, n int) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.ParallacticAngleByHourAngle(
		HourAngleN(date, lon, lat, n),
		basic.HSunApparentDecN(basic.TD2UT(jde, true), n),
		lat,
	)
}
