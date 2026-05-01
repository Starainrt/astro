package orbit

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// EarthDistance 地心距离 / Earth distance.
//
// 返回轨道目标在 date 对应绝对时刻到地球的几何距离，单位 AU。
// Returns the geometric distance from the orbiting target to Earth at the instant represented by date, in astronomical units.
func EarthDistance(date time.Time, elements Elements) float64 {
	return basic.OrbitEarthDistance(ttJulianDay(date), toBasicElements(elements))
}

// SunDistance 日心距离 / Sun distance.
//
// 返回轨道目标在 date 对应绝对时刻到太阳的几何距离，单位 AU。
// Returns the geometric distance from the orbiting target to the Sun at the instant represented by date, in astronomical units.
func SunDistance(date time.Time, elements Elements) float64 {
	return basic.OrbitSunDistance(ttJulianDay(date), toBasicElements(elements))
}

// Elongation 日距角 / elongation.
//
// 返回轨道目标与太阳在地心视方向上的角距，单位度。
// Returns the apparent geocentric angular separation between the target and the Sun, in degrees.
func Elongation(date time.Time, elements Elements) float64 {
	return basic.OrbitElongation(ttJulianDay(date), toBasicElements(elements))
}

// PhaseAngle 相位角 / phase angle.
//
// 返回轨道目标的相位角，单位度。
// Returns the phase angle of the orbiting target, in degrees.
func PhaseAngle(date time.Time, elements Elements) float64 {
	return basic.OrbitPhaseAngle(ttJulianDay(date), toBasicElements(elements))
}

// IlluminatedFraction 被照亮比例 / illuminated fraction.
//
// 返回轨道目标被太阳照亮的可见比例，范围通常为 [0, 1]。
// Returns the illuminated fraction of the target, typically in the range [0, 1].
func IlluminatedFraction(date time.Time, elements Elements) float64 {
	return basic.OrbitIlluminatedFraction(ttJulianDay(date), toBasicElements(elements))
}

// Phase 相位 / phase.
//
// 返回轨道目标被照亮比例，是 IlluminatedFraction 的别名。
// Returns the illuminated fraction of the target and is an alias of IlluminatedFraction.
func Phase(date time.Time, elements Elements) float64 {
	return IlluminatedFraction(date, elements)
}
