package uranus

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// AscendingNode 天王星升交点黄经 / ascending node longitude of Uranus.
func AscendingNode(date time.Time) float64 {
	return AscendingNodeN(date, -1)
}

// AscendingNodeN 天王星升交点黄经（截断版） / truncated ascending node longitude of Uranus.
func AscendingNodeN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.UranusAscendingNodeN(basic.TD2UT(jde, true), n)
}

// DescendingNode 天王星降交点黄经 / descending node longitude of Uranus.
func DescendingNode(date time.Time) float64 {
	return DescendingNodeN(date, -1)
}

// DescendingNodeN 天王星降交点黄经（截断版） / truncated descending node longitude of Uranus.
func DescendingNodeN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.UranusDescendingNodeN(basic.TD2UT(jde, true), n)
}
