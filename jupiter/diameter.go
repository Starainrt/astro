package jupiter

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// Semidiameter 木星视半径，单位角秒 / apparent Jupiter semidiameter in arcseconds.
func (module *Jupiter) Semidiameter(date time.Time) float64 {
	return module.SemidiameterN(date, -1)
}

// SemidiameterN 木星视半径（n 变体），单位角秒 / n-parameterised apparent Jupiter semidiameter in arcseconds.
func (module *Jupiter) SemidiameterN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.JupiterSemidiameterNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// Diameter 木星视直径，单位角秒 / apparent Jupiter diameter in arcseconds.
func (module *Jupiter) Diameter(date time.Time) float64 {
	return module.DiameterN(date, -1)
}

// DiameterN 木星视直径（n 变体），单位角秒 / n-parameterised apparent Jupiter diameter in arcseconds.
func (module *Jupiter) DiameterN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.JupiterDiameterNWithInstance(basic.UTC2TT(jd), n, module.inst)
}
