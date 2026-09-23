package uranus

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// Semidiameter 天王星视半径，单位角秒 / apparent Uranus semidiameter in arcseconds.
func Semidiameter(date time.Time) float64 {
	return SemidiameterN(date, -1)
}

// SemidiameterN 天王星视半径（截断版），单位角秒 / truncated apparent Uranus semidiameter in arcseconds.
func SemidiameterN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.UranusSemidiameterN(basic.UTC2TT(jd), n)
}

// Diameter 天王星视直径，单位角秒 / apparent Uranus diameter in arcseconds.
func Diameter(date time.Time) float64 {
	return DiameterN(date, -1)
}

// DiameterN 天王星视直径（截断版），单位角秒 / truncated apparent Uranus diameter in arcseconds.
func DiameterN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.UranusDiameterN(basic.UTC2TT(jd), n)
}
