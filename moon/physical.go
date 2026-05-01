package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// PhysicalInfo 月球物理观测参数 / physical observing parameters of the Moon.
type PhysicalInfo struct {
	// OpticalLongitude 光学经度天平动，单位度 / optical libration in longitude, degrees.
	OpticalLongitude float64
	// OpticalLatitude 光学纬度天平动，单位度 / optical libration in latitude, degrees.
	OpticalLatitude float64
	// PhysicalLongitude 物理经度天平动，单位度 / physical libration in longitude, degrees.
	PhysicalLongitude float64
	// PhysicalLatitude 物理纬度天平动，单位度 / physical libration in latitude, degrees.
	PhysicalLatitude float64
	// LibrationLongitude 总经度天平动，单位度 / total libration in longitude, degrees.
	LibrationLongitude float64
	// LibrationLatitude 总纬度天平动，单位度 / total libration in latitude, degrees.
	LibrationLatitude float64
	// PositionAngle 月球自转轴位置角，单位度 / position angle of the lunar rotation axis, degrees.
	PositionAngle float64
}

// Physical 月球物理观测参数 / physical observing parameters of the Moon.
func Physical(date time.Time) PhysicalInfo {
	return PhysicalN(date, -1)
}

// PhysicalN 月球物理观测参数（截断版） / truncated physical observing parameters of the Moon.
func PhysicalN(date time.Time, n int) PhysicalInfo {
	jde := basic.Date2JDE(date.UTC())
	info := basic.MoonPhysicalN(basic.TD2UT(jde, true), n)
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
