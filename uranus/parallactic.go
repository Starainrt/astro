package uranus

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ParallacticAngle 天王星视差角（天顶方向角） / Uranus parallactic angle.
func (module *Uranus) ParallacticAngle(date time.Time, lon, lat float64) float64 {
	return basic.ParallacticAngleByHourAngle(module.HourAngle(date, lon), module.ApparentDec(date), lat)
}

// ParallacticAngleN 天王星视差角（天顶方向角）（n 变体） / n-parameterised Uranus parallactic angle.
func (module *Uranus) ParallacticAngleN(date time.Time, lon, lat float64, n int) float64 {
	return basic.ParallacticAngleByHourAngle(module.HourAngleN(date, lon, n), module.ApparentDecN(date, n), lat)
}
