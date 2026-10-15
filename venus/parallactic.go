package venus

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ParallacticAngle 金星视差角（天顶方向角） / Venus parallactic angle.
func (module *Venus) ParallacticAngle(date time.Time, lon, lat float64) float64 {
	return basic.ParallacticAngleByHourAngle(module.HourAngle(date, lon), module.ApparentDec(date), lat)
}

// ParallacticAngleN 金星视差角（天顶方向角）（n 变体） / n-parameterised Venus parallactic angle.
func (module *Venus) ParallacticAngleN(date time.Time, lon, lat float64, n int) float64 {
	return basic.ParallacticAngleByHourAngle(module.HourAngleN(date, lon, n), module.ApparentDecN(date, n), lat)
}
