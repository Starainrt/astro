package neptune

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// Semidiameter 海王星视半径，单位角秒 / apparent Neptune semidiameter in arcseconds.
func (module *Neptune) Semidiameter(date time.Time) float64 {
	return module.SemidiameterN(date, -1)
}

// SemidiameterN 海王星视半径（n 变体），单位角秒 / n-parameterised apparent Neptune semidiameter in arcseconds.
func (module *Neptune) SemidiameterN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.NeptuneSemidiameterNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// Diameter 海王星视直径，单位角秒 / apparent Neptune diameter in arcseconds.
func (module *Neptune) Diameter(date time.Time) float64 {
	return module.DiameterN(date, -1)
}

// DiameterN 海王星视直径（n 变体），单位角秒 / n-parameterised apparent Neptune diameter in arcseconds.
func (module *Neptune) DiameterN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.NeptuneDiameterNWithInstance(basic.UTC2TT(jd), n, module.inst)
}
