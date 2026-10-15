package jupiter

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ParallacticAngle 木星视差角（天顶方向角） / Jupiter parallactic angle.
func (module *Jupiter) ParallacticAngle(date time.Time, lon, lat float64) float64 {
	return basic.ParallacticAngleByHourAngle(module.HourAngle(date, lon), module.ApparentDec(date), lat)
}

// ParallacticAngleN 木星视差角（天顶方向角）（n 变体） / n-parameterised Jupiter parallactic angle.
func (module *Jupiter) ParallacticAngleN(date time.Time, lon, lat float64, n int) float64 {
	return basic.ParallacticAngleByHourAngle(module.HourAngleN(date, lon, n), module.ApparentDecN(date, n), lat)
}
