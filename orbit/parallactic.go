package orbit

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ParallacticAngle 轨道目标视差角（天顶方向角） / orbit-target parallactic angle.
//
// 返回轨道目标在观测者所在地的视差角，单位度；`observerLon` 东经为正，`observerLat` 北纬为正，`observerHeight` 单位米。
// Returns the parallactic angle of the orbital target for the observing site, in degrees. `observerLon` is east-positive, `observerLat` is north-positive, and `observerHeight` is in meters.
func ParallacticAngle(date time.Time, elements Elements, observerLon, observerLat, observerHeight float64) float64 {
	// 时角与赤纬须取自同一次站心求解：分属两条儒略日路径（相差 1 ULP ≈ 40 µs）会引入 ≤2e-9 度漂移，
	// 因此统一到同一时刻后的亚纳度量级输出变化是有意为之，不是纯性能改动。
	_, dec, hourAngle := basic.OrbitHourAngleWithTopocentric(
		basic.Date2JD(date),
		observerLon,
		observerLat,
		observationTimezone(date),
		observerHeight,
		toBasicElements(elements),
	)
	return basic.ParallacticAngleByHourAngle(hourAngle, dec, observerLat)
}
