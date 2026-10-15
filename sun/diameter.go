package sun

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// Semidiameter 太阳视半径，单位角秒 / apparent solar semidiameter in arcseconds.
func (module *Sun) Semidiameter(date time.Time) float64 {
	return module.SemidiameterN(date, -1)
}

// SemidiameterN 太阳视半径（n 变体），单位角秒 / n-parameterised apparent solar semidiameter in arcseconds.
func (module *Sun) SemidiameterN(date time.Time, n int) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.SunSemidiameterNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// Diameter 太阳视直径，单位角秒 / apparent solar diameter in arcseconds.
func (module *Sun) Diameter(date time.Time) float64 {
	return module.DiameterN(date, -1)
}

// DiameterN 太阳视直径（n 变体），单位角秒 / n-parameterised apparent solar diameter in arcseconds.
func (module *Sun) DiameterN(date time.Time, n int) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.SunDiameterNWithInstance(basic.UTC2TT(jd), n, module.inst)
}
