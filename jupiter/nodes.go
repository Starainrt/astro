package jupiter

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// AscendingNode 木星升交点黄经 / ascending node longitude of Jupiter.
func (module *Jupiter) AscendingNode(date time.Time) float64 {
	return module.AscendingNodeN(date, -1)
}

// AscendingNodeN 木星升交点黄经（n 变体） / n-parameterised ascending node longitude of Jupiter.
func (module *Jupiter) AscendingNodeN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.JupiterAscendingNodeNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// DescendingNode 木星降交点黄经 / descending node longitude of Jupiter.
func (module *Jupiter) DescendingNode(date time.Time) float64 {
	return module.DescendingNodeN(date, -1)
}

// DescendingNodeN 木星降交点黄经（n 变体） / n-parameterised descending node longitude of Jupiter.
func (module *Jupiter) DescendingNodeN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.JupiterDescendingNodeNWithInstance(basic.UTC2TT(jd), n, module.inst)
}
