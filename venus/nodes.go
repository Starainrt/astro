package venus

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// AscendingNode 金星升交点黄经 / ascending node longitude of Venus.
func (module *Venus) AscendingNode(date time.Time) float64 {
	return module.AscendingNodeN(date, -1)
}

// AscendingNodeN 金星升交点黄经（n 变体） / n-parameterised ascending node longitude of Venus.
func (module *Venus) AscendingNodeN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.VenusAscendingNodeNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// DescendingNode 金星降交点黄经 / descending node longitude of Venus.
func (module *Venus) DescendingNode(date time.Time) float64 {
	return module.DescendingNodeN(date, -1)
}

// DescendingNodeN 金星降交点黄经（n 变体） / n-parameterised descending node longitude of Venus.
func (module *Venus) DescendingNodeN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.VenusDescendingNodeNWithInstance(basic.UTC2TT(jd), n, module.inst)
}
