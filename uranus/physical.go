package uranus

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// PhysicalInfo 天王星物理观测参数 / physical observing parameters of Uranus.
type PhysicalInfo struct {
	// SubEarthLongitude 子地经度，单位度；采用 Uranus 当前 IAU/Horizons 东经为正约定。
	SubEarthLongitude float64
	// SubEarthLatitude 子地纬度，单位度。
	SubEarthLatitude float64
	// SubSolarLongitude 子日经度，单位度；采用 Uranus 当前 IAU/Horizons 东经为正约定。
	SubSolarLongitude float64
	// SubSolarLatitude 子日纬度，单位度。
	SubSolarLatitude float64
	// NorthPolePositionAngle 天王星北极位置角，单位度。
	NorthPolePositionAngle float64
}

// Physical 天王星物理观测参数 / physical observing parameters of Uranus.
func Physical(date time.Time) PhysicalInfo {
	return PhysicalN(date, -1)
}

// PhysicalN 天王星物理观测参数（截断版） / truncated physical observing parameters of Uranus.
func PhysicalN(date time.Time, n int) PhysicalInfo {
	jde := basic.Date2JDE(date.UTC())
	info := basic.UranusPhysicalN(basic.TD2UT(jde, true), n)
	return PhysicalInfo{
		SubEarthLongitude:      info.SubEarthLongitude,
		SubEarthLatitude:       info.SubEarthLatitude,
		SubSolarLongitude:      info.SubSolarLongitude,
		SubSolarLatitude:       info.SubSolarLatitude,
		NorthPolePositionAngle: info.NorthPolePositionAngle,
	}
}

// PhysicalSystemIII 天王星 System III 物理观测参数 / Uranus System III physical observing parameters.
func PhysicalSystemIII(date time.Time) PhysicalInfo {
	return Physical(date)
}

// PhysicalSystemIIIN 天王星 System III 物理观测参数（截断版） / truncated Uranus System III physical observing parameters.
func PhysicalSystemIIIN(date time.Time, n int) PhysicalInfo {
	return PhysicalN(date, n)
}
