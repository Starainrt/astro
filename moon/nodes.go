package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// AscendingNode 月球升交点黄经 / ascending node longitude of the Moon.
func (module *Moon) AscendingNode(date time.Time) float64 {
	return module.AscendingNodeN(date, -1)
}

// AscendingNodeN 月球升交点黄经（n 变体） / n-parameterised ascending node longitude of the Moon.
func (module *Moon) AscendingNodeN(date time.Time, n int) float64 {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.MoonAscendingNodeNWithInstance(jde, n, module.inst)
}

// DescendingNode 月球降交点黄经 / descending node longitude of the Moon.
func (module *Moon) DescendingNode(date time.Time) float64 {
	return module.DescendingNodeN(date, -1)
}

// DescendingNodeN 月球降交点黄经（n 变体） / n-parameterised descending node longitude of the Moon.
func (module *Moon) DescendingNodeN(date time.Time, n int) float64 {
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.MoonDescendingNodeNWithInstance(jde, n, module.inst)
}
