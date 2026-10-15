package sun

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ParallacticAngle 太阳视差角（天顶方向角） / solar parallactic angle.
func (module *Sun) ParallacticAngle(date time.Time, lon, lat float64) float64 {
	return basic.ParallacticAngleByHourAngle(module.HourAngle(date, lon, lat), module.ApparentDec(date), lat)
}

// ParallacticAngleN 太阳视差角（天顶方向角）（n 变体） / n-parameterised solar parallactic angle.
func (module *Sun) ParallacticAngleN(date time.Time, lon, lat float64, n int) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.ParallacticAngleByHourAngle(
		module.HourAngleN(date, lon, lat, n),
		basic.HSunApparentDecNWithInstance(basic.UTC2TT(jd), n, module.inst),
		lat,
	)
}
