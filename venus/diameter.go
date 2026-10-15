package venus

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// Semidiameter 金星视半径，单位角秒 / apparent Venus semidiameter in arcseconds.
func (module *Venus) Semidiameter(date time.Time) float64 {
	return module.SemidiameterN(date, -1)
}

// SemidiameterN 金星视半径（n 变体），单位角秒 / n-parameterised apparent Venus semidiameter in arcseconds.
func (module *Venus) SemidiameterN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.VenusSemidiameterNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// Diameter 金星视直径，单位角秒 / apparent Venus diameter in arcseconds.
func (module *Venus) Diameter(date time.Time) float64 {
	return module.DiameterN(date, -1)
}

// DiameterN 金星视直径（n 变体），单位角秒 / n-parameterised apparent Venus diameter in arcseconds.
func (module *Venus) DiameterN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.VenusDiameterNWithInstance(basic.UTC2TT(jd), n, module.inst)
}
