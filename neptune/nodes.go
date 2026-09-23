package neptune

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/calendar"
)

// AscendingNode 海王星升交点黄经 / ascending node longitude of Neptune.
func AscendingNode(date time.Time) float64 {
	return AscendingNodeN(date, -1)
}

// AscendingNodeN 海王星升交点黄经（截断版） / truncated ascending node longitude of Neptune.
func AscendingNodeN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.NeptuneAscendingNodeN(basic.UTC2TT(jd), n)
}

// DescendingNode 海王星降交点黄经 / descending node longitude of Neptune.
func DescendingNode(date time.Time) float64 {
	return DescendingNodeN(date, -1)
}

// DescendingNodeN 海王星降交点黄经（截断版） / truncated descending node longitude of Neptune.
func DescendingNodeN(date time.Time, n int) float64 {
	jd := calendar.Date2JD(date.UTC())
	return basic.NeptuneDescendingNodeN(basic.UTC2TT(jd), n)
}
