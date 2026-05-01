package coord

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ParallacticAngleByHourAngle 由时角计算视差角（天顶方向角） / parallactic angle from hour angle.
//
// hourAngle/dec/observerLat 单位均为度，返回值通常落在 [-180, 180] 度。
func ParallacticAngleByHourAngle(hourAngle, dec, observerLat float64) float64 {
	return basic.ParallacticAngleByHourAngle(hourAngle, dec, observerLat)
}

// ParallacticAngle 由赤经赤纬计算视差角（天顶方向角） / parallactic angle from right ascension and declination.
//
// ra/dec 为瞬时赤经赤纬；observerLon/observerLat 为观测者经纬度，东正西负、北正南负。
// Returns the signed parallactic angle for the apparent equatorial coordinates
// at the observing instant.
func ParallacticAngle(date time.Time, ra, dec, observerLon, observerLat float64) float64 {
	return basic.StarParallacticAngle(jdeUTC(date), ra, dec, observerLon, observerLat, 0)
}
