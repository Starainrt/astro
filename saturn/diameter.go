package saturn

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// Semidiameter 土星视半径，单位角秒 / apparent Saturn semidiameter in arcseconds.
func (module *Saturn) Semidiameter(date time.Time) float64 {
	return module.SemidiameterN(date, -1)
}

// SemidiameterN 土星视半径（n 变体），单位角秒 / n-parameterised apparent Saturn semidiameter in arcseconds.
func (module *Saturn) SemidiameterN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.SaturnSemidiameterNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// Diameter 土星视直径，单位角秒 / apparent Saturn diameter in arcseconds.
func (module *Saturn) Diameter(date time.Time) float64 {
	return module.DiameterN(date, -1)
}

// DiameterN 土星视直径（n 变体），单位角秒 / n-parameterised apparent Saturn diameter in arcseconds.
func (module *Saturn) DiameterN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.SaturnDiameterNWithInstance(basic.UTC2TT(jd), n, module.inst)
}
