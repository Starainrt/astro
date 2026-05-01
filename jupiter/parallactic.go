package jupiter

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ParallacticAngle 木星视差角（天顶方向角） / Jupiter parallactic angle.
func ParallacticAngle(date time.Time, lon, lat float64) float64 {
	return basic.ParallacticAngleByHourAngle(HourAngle(date, lon), ApparentDec(date), lat)
}

// ParallacticAngleN 截断项木星视差角（天顶方向角） / truncated Jupiter parallactic angle.
func ParallacticAngleN(date time.Time, lon, lat float64, n int) float64 {
	return basic.ParallacticAngleByHourAngle(HourAngleN(date, lon, n), ApparentDecN(date, n), lat)
}
