package uranus

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// Semidiameter 天王星视半径，单位角秒 / apparent Uranus semidiameter in arcseconds.
func (module *Uranus) Semidiameter(date time.Time) float64 {
	return module.SemidiameterN(date, -1)
}

// SemidiameterN 天王星视半径（n 变体），单位角秒 / n-parameterised apparent Uranus semidiameter in arcseconds.
func (module *Uranus) SemidiameterN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.UranusSemidiameterNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// Diameter 天王星视直径，单位角秒 / apparent Uranus diameter in arcseconds.
func (module *Uranus) Diameter(date time.Time) float64 {
	return module.DiameterN(date, -1)
}

// DiameterN 天王星视直径（n 变体），单位角秒 / n-parameterised apparent Uranus diameter in arcseconds.
func (module *Uranus) DiameterN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.UranusDiameterNWithInstance(basic.UTC2TT(jd), n, module.inst)
}
