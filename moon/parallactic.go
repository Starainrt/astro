package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ParallacticAngle 月亮视差角（天顶方向角） / lunar parallactic angle.
func (module *Moon) ParallacticAngle(date time.Time, lon, lat float64) float64 {
	return basic.ParallacticAngleByHourAngle(module.HourAngle(date, lon, lat), module.ApparentDec(date, lon, lat), lat)
}
