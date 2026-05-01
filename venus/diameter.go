package venus

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// Semidiameter 金星视半径，单位角秒 / apparent Venus semidiameter in arcseconds.
func Semidiameter(date time.Time) float64 {
	return SemidiameterN(date, -1)
}

// SemidiameterN 金星视半径（截断版），单位角秒 / truncated apparent Venus semidiameter in arcseconds.
func SemidiameterN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.VenusSemidiameterN(basic.TD2UT(jde, true), n)
}

// Diameter 金星视直径，单位角秒 / apparent Venus diameter in arcseconds.
func Diameter(date time.Time) float64 {
	return DiameterN(date, -1)
}

// DiameterN 金星视直径（截断版），单位角秒 / truncated apparent Venus diameter in arcseconds.
func DiameterN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.VenusDiameterN(basic.TD2UT(jde, true), n)
}
