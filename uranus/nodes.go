package uranus

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// AscendingNode 天王星升交点黄经 / ascending node longitude of Uranus.
func (module *Uranus) AscendingNode(date time.Time) float64 {
	return module.AscendingNodeN(date, -1)
}

// AscendingNodeN 天王星升交点黄经（n 变体） / n-parameterised ascending node longitude of Uranus.
func (module *Uranus) AscendingNodeN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.UranusAscendingNodeNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// DescendingNode 天王星降交点黄经 / descending node longitude of Uranus.
func (module *Uranus) DescendingNode(date time.Time) float64 {
	return module.DescendingNodeN(date, -1)
}

// DescendingNodeN 天王星降交点黄经（n 变体） / n-parameterised descending node longitude of Uranus.
func (module *Uranus) DescendingNodeN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.UranusDescendingNodeNWithInstance(basic.UTC2TT(jd), n, module.inst)
}
