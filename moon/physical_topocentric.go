package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// TopocentricPhysical 月球站心物理观测参数 / topocentric physical observing parameters of the Moon.
//
// date 为观测时刻；observerLon/observerLat 为观测者经纬度，东正西负、北正南负；height 为海拔高度，单位米。
// date is the observing instant; observerLon/observerLat are east-positive and north-positive; height is observer elevation in meters.
func TopocentricPhysical(date time.Time, observerLon, observerLat, height float64) PhysicalInfo {
	return TopocentricPhysicalN(date, observerLon, observerLat, height, -1)
}

// TopocentricPhysicalN 月球站心物理观测参数（截断版） / truncated topocentric physical observing parameters of the Moon.
func TopocentricPhysicalN(date time.Time, observerLon, observerLat, height float64, n int) PhysicalInfo {
	info := basic.MoonTopocentricPhysicalN(observationTT(date), observerLon, observerLat, height, n)
	return PhysicalInfo{
		OpticalLongitude:   info.OpticalLongitude,
		OpticalLatitude:    info.OpticalLatitude,
		PhysicalLongitude:  info.PhysicalLongitude,
		PhysicalLatitude:   info.PhysicalLatitude,
		LibrationLongitude: info.LibrationLongitude,
		LibrationLatitude:  info.LibrationLatitude,
		PositionAngle:      info.PositionAngle,
	}
}
