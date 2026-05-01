package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// Semidiameter 月亮视半径，单位角秒 / apparent lunar semidiameter in arcseconds.
func Semidiameter(date time.Time) float64 {
	return SemidiameterN(date, -1)
}

// SemidiameterN 月亮视半径（截断版），单位角秒 / truncated apparent lunar semidiameter in arcseconds.
func SemidiameterN(date time.Time, n int) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.MoonSemidiameterN(basic.TD2UT(jde, true), n)
}

// Diameter 月亮视直径，单位角秒 / apparent lunar diameter in arcseconds.
func Diameter(date time.Time) float64 {
	return DiameterN(date, -1)
}

// DiameterN 月亮视直径（截断版），单位角秒 / truncated apparent lunar diameter in arcseconds.
func DiameterN(date time.Time, n int) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.MoonDiameterN(basic.TD2UT(jde, true), n)
}
