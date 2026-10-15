package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// Semidiameter 月亮视半径，单位角秒 / apparent lunar semidiameter in arcseconds.
func (module *Moon) Semidiameter(date time.Time) float64 {
	return module.SemidiameterN(date, -1)
}

// SemidiameterN 月亮视半径（n 变体），单位角秒 / n-parameterised apparent lunar semidiameter in arcseconds.
func (module *Moon) SemidiameterN(date time.Time, n int) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.MoonSemidiameterNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// Diameter 月亮视直径，单位角秒 / apparent lunar diameter in arcseconds.
func (module *Moon) Diameter(date time.Time) float64 {
	return module.DiameterN(date, -1)
}

// DiameterN 月亮视直径（n 变体），单位角秒 / n-parameterised apparent lunar diameter in arcseconds.
func (module *Moon) DiameterN(date time.Time, n int) float64 {
	jd := basic.Date2JD(date.UTC())
	return basic.MoonDiameterNWithInstance(basic.UTC2TT(jd), n, module.inst)
}
