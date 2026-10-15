package neptune

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ParallacticAngle 海王星视差角（天顶方向角） / Neptune parallactic angle.
func (module *Neptune) ParallacticAngle(date time.Time, lon, lat float64) float64 {
	return basic.ParallacticAngleByHourAngle(module.HourAngle(date, lon), module.ApparentDec(date), lat)
}

// ParallacticAngleN 海王星视差角（天顶方向角）（n 变体） / n-parameterised Neptune parallactic angle.
func (module *Neptune) ParallacticAngleN(date time.Time, lon, lat float64, n int) float64 {
	return basic.ParallacticAngleByHourAngle(module.HourAngleN(date, lon, n), module.ApparentDecN(date, n), lat)
}
