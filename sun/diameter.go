package sun

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// Semidiameter 太阳视半径，单位角秒 / apparent solar semidiameter in arcseconds.
func Semidiameter(date time.Time) float64 {
	return SemidiameterN(date, -1)
}

// SemidiameterN 太阳视半径（截断版），单位角秒 / truncated apparent solar semidiameter in arcseconds.
func SemidiameterN(date time.Time, n int) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.SunSemidiameterN(basic.UTC2TT(jd), n)
}

// Diameter 太阳视直径，单位角秒 / apparent solar diameter in arcseconds.
func Diameter(date time.Time) float64 {
	return DiameterN(date, -1)
}

// DiameterN 太阳视直径（截断版），单位角秒 / truncated apparent solar diameter in arcseconds.
func DiameterN(date time.Time, n int) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.SunDiameterN(basic.UTC2TT(jd), n)
}
