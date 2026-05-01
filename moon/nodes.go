package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// AscendingNode 月球升交点黄经 / ascending node longitude of the Moon.
func AscendingNode(date time.Time) float64 {
	return AscendingNodeN(date, -1)
}

// AscendingNodeN 月球升交点黄经（截断版） / truncated ascending node longitude of the Moon.
func AscendingNodeN(date time.Time, n int) float64 {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.MoonAscendingNodeN(jde, n)
}

// DescendingNode 月球降交点黄经 / descending node longitude of the Moon.
func DescendingNode(date time.Time) float64 {
	return DescendingNodeN(date, -1)
}

// DescendingNodeN 月球降交点黄经（截断版） / truncated descending node longitude of the Moon.
func DescendingNodeN(date time.Time, n int) float64 {
	jde := basic.TD2UT(basic.Date2JDE(date.UTC()), true)
	return basic.MoonDescendingNodeN(jde, n)
}
