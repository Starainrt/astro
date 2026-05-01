package saturn

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// RingInfo 土星环观测参数 / Saturn ring observing parameters.
type RingInfo struct {
	// EarthLatitude 地球在土星环面上的土星心纬度 B，单位度 / Saturnicentric latitude of Earth, degrees.
	EarthLatitude float64
	// SunLatitude 太阳在土星环面上的土星心纬度 B'，单位度 / Saturnicentric latitude of Sun, degrees.
	SunLatitude float64
	// PositionAngle 土星环北半短轴位置角，单位度 / position angle of the northern semiminor axis, degrees.
	PositionAngle float64
	// DeltaU 太阳和地球在环面内的土星心黄经差，单位度 / difference of Saturnicentric ring longitudes, degrees.
	DeltaU float64
	// MajorAxis 土星环外缘长轴，单位角秒 / outer ring major axis, arcseconds.
	MajorAxis float64
	// MinorAxis 土星环外缘短轴，单位角秒 / outer ring minor axis, arcseconds.
	MinorAxis float64
}

// Ring 土星环观测参数 / Saturn ring observing parameters.
func Ring(date time.Time) RingInfo {
	return RingN(date, -1)
}

// RingN 土星环观测参数（截断版） / truncated Saturn ring observing parameters.
func RingN(date time.Time, n int) RingInfo {
	jde := calendar.Date2JDE(date.UTC())
	earthLatitude, sunLatitude, positionAngle, deltaU, majorAxis, minorAxis := basic.SaturnRingParametersN(basic.TD2UT(jde, true), n)
	return RingInfo{
		EarthLatitude: earthLatitude,
		SunLatitude:   sunLatitude,
		PositionAngle: positionAngle,
		DeltaU:        deltaU,
		MajorAxis:     majorAxis,
		MinorAxis:     minorAxis,
	}
}
