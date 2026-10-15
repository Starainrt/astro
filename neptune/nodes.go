package neptune

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// AscendingNode 海王星升交点黄经 / ascending node longitude of Neptune.
func (module *Neptune) AscendingNode(date time.Time) float64 {
	return module.AscendingNodeN(date, -1)
}

// AscendingNodeN 海王星升交点黄经（n 变体） / n-parameterised ascending node longitude of Neptune.
func (module *Neptune) AscendingNodeN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.NeptuneAscendingNodeNWithInstance(basic.UTC2TT(jd), n, module.inst)
}

// DescendingNode 海王星降交点黄经 / descending node longitude of Neptune.
func (module *Neptune) DescendingNode(date time.Time) float64 {
	return module.DescendingNodeN(date, -1)
}

// DescendingNodeN 海王星降交点黄经（n 变体） / n-parameterised descending node longitude of Neptune.
func (module *Neptune) DescendingNodeN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.NeptuneDescendingNodeNWithInstance(basic.UTC2TT(jd), n, module.inst)
}
