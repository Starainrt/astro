package basic

import (
	"math"

	. "github.com/starainrt/astro/tools"
)

// ParallacticAngleByHourAngle 返回视差角（天顶方向角）/ parallactic angle.
//
// hourAngle 为目标时角，dec 为赤纬，lat 为观测者纬度，单位均为度。
// 返回值是 atan2 公式给出的有符号结果，范围通常为 [-180, 180] 度。
// hourAngle may be signed or normalized to [0, 360); the trigonometric
// formula handles either representation.
func ParallacticAngleByHourAngle(hourAngle, dec, lat float64) float64 {
	return math.Atan2(
		Sin(hourAngle),
		Tan(lat)*Cos(dec)-Sin(dec)*Cos(hourAngle),
	) * 180 / math.Pi
}

// StarParallacticAngle 返回星体在给定观测条件下的视差角（天顶方向角）/ parallactic angle.
func StarParallacticAngle(jde, ra, dec, lon, lat, timezone float64) float64 {
	return ParallacticAngleByHourAngle(StarHourAngle(jde, ra, lon, timezone), dec, lat)
}
