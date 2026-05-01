package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// BrightLimbPositionAngle 月亮明亮边缘位置角，单位度 / position angle of the Moon's bright limb in degrees.
func BrightLimbPositionAngle(date time.Time) float64 {
	return BrightLimbPositionAngleN(date, -1)
}

// BrightLimbPositionAngleN 月亮明亮边缘位置角（截断版），单位度 / truncated position angle of the Moon's bright limb in degrees.
func BrightLimbPositionAngleN(date time.Time, n int) float64 {
	return basic.MoonBrightLimbPositionAngleN(observationTT(date), n)
}

// TopocentricBrightLimbPositionAngle 月亮站心明亮边缘位置角，单位度 / topocentric position angle of the Moon's bright limb in degrees.
//
// date 为观测时刻；observerLon/observerLat 为观测者经纬度，东正西负、北正南负；height 为海拔高度，单位米。
// date is the observing instant; observerLon/observerLat are east-positive and north-positive; height is observer elevation in meters.
func TopocentricBrightLimbPositionAngle(date time.Time, observerLon, observerLat, height float64) float64 {
	return TopocentricBrightLimbPositionAngleN(date, observerLon, observerLat, height, -1)
}

// TopocentricBrightLimbPositionAngleN 月亮站心明亮边缘位置角（截断版），单位度 / truncated topocentric position angle of the Moon's bright limb in degrees.
func TopocentricBrightLimbPositionAngleN(date time.Time, observerLon, observerLat, height float64, n int) float64 {
	return basic.MoonTopocentricBrightLimbPositionAngleN(observationTT(date), observerLon, observerLat, height, n)
}

func observationTT(date time.Time) float64 {
	return basic.TD2UT(basic.Date2JDE(date.UTC()), true)
}
