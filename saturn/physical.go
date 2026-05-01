package saturn

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// PhysicalInfo 土星物理观测参数 / physical observing parameters of Saturn.
type PhysicalInfo struct {
	// SubEarthLongitude 子地经度，单位度；采用 Saturn 当前 IAU/Horizons 西经为正约定。
	SubEarthLongitude float64
	// SubEarthLatitude 子地纬度，单位度。
	SubEarthLatitude float64
	// SubSolarLongitude 子日经度，单位度；采用 Saturn 当前 IAU/Horizons 西经为正约定。
	SubSolarLongitude float64
	// SubSolarLatitude 子日纬度，单位度。
	SubSolarLatitude float64
	// NorthPolePositionAngle 土星北极位置角，单位度。
	NorthPolePositionAngle float64
}

// Physical 土星物理观测参数 / physical observing parameters of Saturn.
func Physical(date time.Time) PhysicalInfo {
	return PhysicalN(date, -1)
}

// PhysicalN 土星物理观测参数（截断版） / truncated physical observing parameters of Saturn.
func PhysicalN(date time.Time, n int) PhysicalInfo {
	jde := basic.Date2JDE(date.UTC())
	info := basic.SaturnPhysicalN(basic.TD2UT(jde, true), n)
	return PhysicalInfo{
		SubEarthLongitude:      info.SubEarthLongitude,
		SubEarthLatitude:       info.SubEarthLatitude,
		SubSolarLongitude:      info.SubSolarLongitude,
		SubSolarLatitude:       info.SubSolarLatitude,
		NorthPolePositionAngle: info.NorthPolePositionAngle,
	}
}

// PhysicalSystemIII 土星 System III 物理观测参数 / Saturn System III physical observing parameters.
func PhysicalSystemIII(date time.Time) PhysicalInfo {
	return Physical(date)
}

// PhysicalSystemIIIN 土星 System III 物理观测参数（截断版） / truncated Saturn System III physical observing parameters.
func PhysicalSystemIIIN(date time.Time, n int) PhysicalInfo {
	return PhysicalN(date, n)
}
