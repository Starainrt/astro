package saturn

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// Semidiameter 土星视半径，单位角秒 / apparent Saturn semidiameter in arcseconds.
func Semidiameter(date time.Time) float64 {
	return SemidiameterN(date, -1)
}

// SemidiameterN 土星视半径（截断版），单位角秒 / truncated apparent Saturn semidiameter in arcseconds.
func SemidiameterN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.SaturnSemidiameterN(basic.TD2UT(jde, true), n)
}

// Diameter 土星视直径，单位角秒 / apparent Saturn diameter in arcseconds.
func Diameter(date time.Time) float64 {
	return DiameterN(date, -1)
}

// DiameterN 土星视直径（截断版），单位角秒 / truncated apparent Saturn diameter in arcseconds.
func DiameterN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.SaturnDiameterN(basic.TD2UT(jde, true), n)
}
