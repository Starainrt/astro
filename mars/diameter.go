package mars

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// Semidiameter 火星视半径，单位角秒 / apparent Mars semidiameter in arcseconds.
func (module *Mars) Semidiameter(date time.Time) float64 {
	return module.SemidiameterN(date, -1)
}

// SemidiameterN 火星视半径（n 变体），单位角秒 / n-parameterised apparent Mars semidiameter in arcseconds.
func (module *Mars) SemidiameterN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.MarsSemidiameterNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// Diameter 火星视直径，单位角秒 / apparent Mars diameter in arcseconds.
func (module *Mars) Diameter(date time.Time) float64 {
	return module.DiameterN(date, -1)
}

// DiameterN 火星视直径（n 变体），单位角秒 / n-parameterised apparent Mars diameter in arcseconds.
func (module *Mars) DiameterN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.MarsDiameterNWithInstance(basic.UTC2TT(jd), n, module.inst)
}
