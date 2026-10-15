package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// TopocentricPhysical 月球站心物理观测参数 / topocentric physical observing parameters of the Moon.
func (module *Moon) TopocentricPhysical(date time.Time, observerLon, observerLat, height float64) PhysicalInfo {
	return module.TopocentricPhysicalN(date, observerLon, observerLat, height, -1)
}

// TopocentricPhysicalN 月球站心物理观测参数（n 变体） / n-parameterised topocentric physical observing parameters of the Moon.
func (module *Moon) TopocentricPhysicalN(date time.Time, observerLon, observerLat, height float64, n int) PhysicalInfo {
	info := basic.MoonTopocentricPhysicalNWithInstance(observationTT(date), observerLon, observerLat, height, n, module.inst)
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
