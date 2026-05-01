package mercury

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// AscendingNode 水星升交点黄经 / ascending node longitude of Mercury.
func AscendingNode(date time.Time) float64 {
	return AscendingNodeN(date, -1)
}

// AscendingNodeN 水星升交点黄经（截断版） / truncated ascending node longitude of Mercury.
func AscendingNodeN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.MercuryAscendingNodeN(basic.TD2UT(jde, true), n)
}

// DescendingNode 水星降交点黄经 / descending node longitude of Mercury.
func DescendingNode(date time.Time) float64 {
	return DescendingNodeN(date, -1)
}

// DescendingNodeN 水星降交点黄经（截断版） / truncated descending node longitude of Mercury.
func DescendingNodeN(date time.Time, n int) float64 {
	jde := calendar.Date2JDE(date.UTC())
	return basic.MercuryDescendingNodeN(basic.TD2UT(jde, true), n)
}
