package mercury

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// Semidiameter 水星视半径，单位角秒 / apparent Mercury semidiameter in arcseconds.
func (module *Mercury) Semidiameter(date time.Time) float64 {
	return module.SemidiameterN(date, -1)
}

// SemidiameterN 水星视半径（n 变体），单位角秒 / n-parameterised apparent Mercury semidiameter in arcseconds.
func (module *Mercury) SemidiameterN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.MercurySemidiameterNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// Diameter 水星视直径，单位角秒 / apparent Mercury diameter in arcseconds.
func (module *Mercury) Diameter(date time.Time) float64 {
	return module.DiameterN(date, -1)
}

// DiameterN 水星视直径（n 变体），单位角秒 / n-parameterised apparent Mercury diameter in arcseconds.
func (module *Mercury) DiameterN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.MercuryDiameterNWithInstance(basic.UTC2TT(jd), n, module.inst)
}
