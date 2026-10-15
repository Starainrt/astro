package saturn

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ParallacticAngle 土星视差角（天顶方向角） / Saturn parallactic angle.
func (module *Saturn) ParallacticAngle(date time.Time, lon, lat float64) float64 {
	return basic.ParallacticAngleByHourAngle(module.HourAngle(date, lon), module.ApparentDec(date), lat)
}

// ParallacticAngleN 土星视差角（天顶方向角）（n 变体） / n-parameterised Saturn parallactic angle.
func (module *Saturn) ParallacticAngleN(date time.Time, lon, lat float64, n int) float64 {
	return basic.ParallacticAngleByHourAngle(module.HourAngleN(date, lon, n), module.ApparentDecN(date, n), lat)
}
