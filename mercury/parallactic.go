package mercury

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ParallacticAngle 水星视差角（天顶方向角） / Mercury parallactic angle.
func (module *Mercury) ParallacticAngle(date time.Time, lon, lat float64) float64 {
	return basic.ParallacticAngleByHourAngle(module.HourAngle(date, lon), module.ApparentDec(date), lat)
}

// ParallacticAngleN 水星视差角（天顶方向角）（n 变体） / n-parameterised Mercury parallactic angle.
func (module *Mercury) ParallacticAngleN(date time.Time, lon, lat float64, n int) float64 {
	return basic.ParallacticAngleByHourAngle(module.HourAngleN(date, lon, n), module.ApparentDecN(date, n), lat)
}
