package venus

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// AscendingNode 金星升交点黄经 / ascending node longitude of Venus.
func AscendingNode(date time.Time) float64 {
	return AscendingNodeN(date, -1)
}

// AscendingNodeN 金星升交点黄经（截断版） / truncated ascending node longitude of Venus.
func AscendingNodeN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.VenusAscendingNodeN(basic.TD2UT(jde, true), n)
}

// DescendingNode 金星降交点黄经 / descending node longitude of Venus.
func DescendingNode(date time.Time) float64 {
	return DescendingNodeN(date, -1)
}

// DescendingNodeN 金星降交点黄经（截断版） / truncated descending node longitude of Venus.
func DescendingNodeN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.VenusDescendingNodeN(basic.TD2UT(jde, true), n)
}
