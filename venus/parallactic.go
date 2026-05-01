package venus

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ParallacticAngle 金星视差角（天顶方向角） / Venus parallactic angle.
func ParallacticAngle(date time.Time, lon, lat float64) float64 {
	return basic.ParallacticAngleByHourAngle(HourAngle(date, lon), ApparentDec(date), lat)
}

// ParallacticAngleN 截断项金星视差角（天顶方向角） / truncated Venus parallactic angle.
func ParallacticAngleN(date time.Time, lon, lat float64, n int) float64 {
	return basic.ParallacticAngleByHourAngle(HourAngleN(date, lon, n), ApparentDecN(date, n), lat)
}
