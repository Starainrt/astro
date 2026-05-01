package mars

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// AscendingNode 火星升交点黄经 / ascending node longitude of Mars.
func AscendingNode(date time.Time) float64 {
	return AscendingNodeN(date, -1)
}

// AscendingNodeN 火星升交点黄经（截断版） / truncated ascending node longitude of Mars.
func AscendingNodeN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.MarsAscendingNodeN(basic.TD2UT(jde, true), n)
}

// DescendingNode 火星降交点黄经 / descending node longitude of Mars.
func DescendingNode(date time.Time) float64 {
	return DescendingNodeN(date, -1)
}

// DescendingNodeN 火星降交点黄经（截断版） / truncated descending node longitude of Mars.
func DescendingNodeN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.MarsDescendingNodeN(basic.TD2UT(jde, true), n)
}
