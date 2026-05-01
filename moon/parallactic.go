package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ParallacticAngle 月亮视差角（天顶方向角） / lunar parallactic angle.
//
// 对月亮使用现有站心视时角和站心视赤纬链路，因此会显式依赖观测者经纬度。
func ParallacticAngle(date time.Time, lon, lat float64) float64 {
	return basic.ParallacticAngleByHourAngle(HourAngle(date, lon, lat), ApparentDec(date, lon, lat), lat)
}
