package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// BrightLimbPositionAngle 月亮明亮边缘位置角，单位度 / position angle of the Moon's bright limb in degrees.
func (module *Moon) BrightLimbPositionAngle(date time.Time) float64 {
	return module.BrightLimbPositionAngleN(date, -1)
}

// BrightLimbPositionAngleN 月亮明亮边缘位置角（n 变体），单位度 / n-parameterised position angle of the Moon's bright limb in degrees.
func (module *Moon) BrightLimbPositionAngleN(date time.Time, n int) float64 {
	return basic.MoonBrightLimbPositionAngleNWithInstance(observationTT(date), n, module.inst)
}

// TopocentricBrightLimbPositionAngle 月亮站心明亮边缘位置角，单位度 / topocentric position angle of the Moon's bright limb in degrees.
func (module *Moon) TopocentricBrightLimbPositionAngle(date time.Time, observerLon, observerLat, height float64) float64 {
	return module.TopocentricBrightLimbPositionAngleN(date, observerLon, observerLat, height, -1)
}

// TopocentricBrightLimbPositionAngleN 月亮站心明亮边缘位置角（n 变体），单位度 / n-parameterised topocentric position angle of the Moon's bright limb in degrees.
func (module *Moon) TopocentricBrightLimbPositionAngleN(date time.Time, observerLon, observerLat, height float64, n int) float64 {
	return basic.MoonTopocentricBrightLimbPositionAngleNWithInstance(observationTT(date), observerLon, observerLat, height, n, module.inst)
}

func observationTT(date time.Time) float64 {
	return basic.UTC2TT(basic.Date2JD(date.UTC()))
}
