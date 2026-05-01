package mercury

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ParallacticAngle 水星视差角（天顶方向角） / Mercury parallactic angle.
func ParallacticAngle(date time.Time, lon, lat float64) float64 {
	return basic.ParallacticAngleByHourAngle(HourAngle(date, lon), ApparentDec(date), lat)
}

// ParallacticAngleN 截断项水星视差角（天顶方向角） / truncated Mercury parallactic angle.
func ParallacticAngleN(date time.Time, lon, lat float64, n int) float64 {
	return basic.ParallacticAngleByHourAngle(HourAngleN(date, lon, n), ApparentDecN(date, n), lat)
}
