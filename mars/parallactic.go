package mars

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ParallacticAngle 火星视差角（天顶方向角） / Mars parallactic angle.
func (module *Mars) ParallacticAngle(date time.Time, lon, lat float64) float64 {
	return basic.ParallacticAngleByHourAngle(module.HourAngle(date, lon), module.ApparentDec(date), lat)
}

// ParallacticAngleN 火星视差角（天顶方向角）（n 变体） / n-parameterised Mars parallactic angle.
func (module *Mars) ParallacticAngleN(date time.Time, lon, lat float64, n int) float64 {
	return basic.ParallacticAngleByHourAngle(module.HourAngleN(date, lon, n), module.ApparentDecN(date, n), lat)
}
