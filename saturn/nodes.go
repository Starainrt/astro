package saturn

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// AscendingNode 土星升交点黄经 / ascending node longitude of Saturn.
func (module *Saturn) AscendingNode(date time.Time) float64 {
	return module.AscendingNodeN(date, -1)
}

// AscendingNodeN 土星升交点黄经（n 变体） / n-parameterised ascending node longitude of Saturn.
func (module *Saturn) AscendingNodeN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.SaturnAscendingNodeNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// DescendingNode 土星降交点黄经 / descending node longitude of Saturn.
func (module *Saturn) DescendingNode(date time.Time) float64 {
	return module.DescendingNodeN(date, -1)
}

// DescendingNodeN 土星降交点黄经（n 变体） / n-parameterised descending node longitude of Saturn.
func (module *Saturn) DescendingNodeN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.SaturnDescendingNodeNWithInstance(basic.UTC2TT(jd), n, module.inst)
}
