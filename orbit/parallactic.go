package orbit

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ParallacticAngle 轨道目标视差角（天顶方向角） / orbit-target parallactic angle.
//
// 返回轨道目标在观测者所在地的视差角，单位度；`observerLon` 东经为正，`observerLat` 北纬为正，`observerHeight` 单位米。
func ParallacticAngle(date time.Time, elements Elements, observerLon, observerLat, observerHeight float64) float64 {
	position := ApparentTopocentricEquatorial(date, elements, observerLon, observerLat, observerHeight)
	return basic.ParallacticAngleByHourAngle(
		HourAngle(date, elements, observerLon, observerLat, observerHeight),
		position.Dec,
		observerLat,
	)
}
