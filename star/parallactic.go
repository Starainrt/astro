package star

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ParallacticAngle 恒星视差角（天顶方向角） / stellar parallactic angle.
//
// ra/dec 为瞬时赤经赤纬，单位度；lon/lat 为观测者经纬度，东正西负、北正南负。
// 返回值为有符号视差角，单位度。
// ra/dec are apparent equatorial coordinates in degrees; lon/lat are east-positive and north-positive.
// Returns the signed parallactic angle in degrees.
func ParallacticAngle(date time.Time, ra, dec, lon, lat float64) float64 {
	jde := basic.Date2JDE(date)
	_, loc := date.Zone()
	timezone := float64(loc) / 3600.0
	return basic.StarParallacticAngle(jde, ra, dec, lon, lat, timezone)
}
