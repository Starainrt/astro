package mercury

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// AscendingNode 水星升交点黄经 / ascending node longitude of Mercury.
func (module *Mercury) AscendingNode(date time.Time) float64 {
	return module.AscendingNodeN(date, -1)
}

// AscendingNodeN 水星升交点黄经（n 变体） / n-parameterised ascending node longitude of Mercury.
func (module *Mercury) AscendingNodeN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.MercuryAscendingNodeNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// DescendingNode 水星降交点黄经 / descending node longitude of Mercury.
func (module *Mercury) DescendingNode(date time.Time) float64 {
	return module.DescendingNodeN(date, -1)
}

// DescendingNodeN 水星降交点黄经（n 变体） / n-parameterised descending node longitude of Mercury.
func (module *Mercury) DescendingNodeN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.MercuryDescendingNodeNWithInstance(basic.UTC2TT(jd), n, module.inst)
}
