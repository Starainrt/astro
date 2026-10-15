package mars

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// AscendingNode 火星升交点黄经 / ascending node longitude of Mars.
func (module *Mars) AscendingNode(date time.Time) float64 {
	return module.AscendingNodeN(date, -1)
}

// AscendingNodeN 火星升交点黄经（n 变体） / n-parameterised ascending node longitude of Mars.
func (module *Mars) AscendingNodeN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.MarsAscendingNodeNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// DescendingNode 火星降交点黄经 / descending node longitude of Mars.
func (module *Mars) DescendingNode(date time.Time) float64 {
	return module.DescendingNodeN(date, -1)
}

// DescendingNodeN 火星降交点黄经（n 变体） / n-parameterised descending node longitude of Mars.
func (module *Mars) DescendingNodeN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.MarsDescendingNodeNWithInstance(basic.UTC2TT(jd), n, module.inst)
}
