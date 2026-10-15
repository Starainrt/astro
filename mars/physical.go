package mars

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// PhysicalInfo 火星物理观测参数 / physical observing parameters of Mars.
type PhysicalInfo struct {
	// SubEarthLongitude 子地经度，单位度；采用 Mars 当前 IAU/Horizons 西经为正约定。
	SubEarthLongitude float64
	// SubEarthLatitude 子地纬度，单位度。
	SubEarthLatitude float64
	// SubSolarLongitude 子日经度，单位度；采用 Mars 当前 IAU/Horizons 西经为正约定。
	SubSolarLongitude float64
	// SubSolarLatitude 子日纬度，单位度。
	SubSolarLatitude float64
	// NorthPolePositionAngle 火星北极位置角，单位度。
	NorthPolePositionAngle float64
}

// Physical 火星物理观测参数 / physical observing parameters of Mars.
func (module *Mars) Physical(date time.Time) PhysicalInfo {
	return module.PhysicalN(date, -1)
}

// PhysicalN 火星物理观测参数（n 变体） / n-parameterised physical observing parameters of Mars.
func (module *Mars) PhysicalN(date time.Time, n int) PhysicalInfo {
	jd := basic.Date2JD(date.UTC())
	info := basic.MarsPhysicalNWithInstance(basic.UTC2TT(jd), n, module.inst)
	return PhysicalInfo{
		SubEarthLongitude:      info.SubEarthLongitude,
		SubEarthLatitude:       info.SubEarthLatitude,
		SubSolarLongitude:      info.SubSolarLongitude,
		SubSolarLatitude:       info.SubSolarLatitude,
		NorthPolePositionAngle: info.NorthPolePositionAngle,
	}
}
