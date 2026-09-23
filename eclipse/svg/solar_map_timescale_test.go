package svg

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro"
)

// 太阳能图的 UT1 输出只改时刻文字：直射点、闭合弧与等时线几何必须逐字节相同。
// 时刻标记按输出时标的整点取点（UT1 整点是另一个物理时刻），与文字一起从比较里剔除。
func TestSolarEclipseMapUT1KeepsGeometry(t *testing.T) {
	for _, date := range []time.Time{
		time.Date(2021, time.June, 10, 12, 0, 0, 0, time.UTC),    // DUT1 较大，直射点位移可测
		time.Date(1995, time.October, 24, 12, 0, 0, 0, time.UTC), // 双向极限事件
		time.Date(2024, time.April, 8, 12, 0, 0, 0, time.UTC),
	} {
		options := SolarEclipseMapSVGOptions{
			Width: 1200, Height: 800, Location: time.UTC, TimeLabelStep: 30 * time.Minute,
		}
		utc, ok := SolarEclipseMapSVG(date, options)
		if !ok {
			t.Fatalf("%s: UTC render failed", date.Format("2006-01-02"))
		}
		ut1Options := options
		ut1Options.TimeScale = astro.TimeScaleUT1
		ut1, ok := SolarEclipseMapSVG(date, ut1Options)
		if !ok {
			t.Fatalf("%s: UT1 render failed", date.Format("2006-01-02"))
		}
		if !strings.Contains(ut1, "DUT1 = UT1−UTC") {
			t.Errorf("%s: UT1 chart does not declare the DUT1 offset", date.Format("2006-01-02"))
		}
		cst := time.FixedZone("CST", 8*3600)
		badOptions := ut1Options
		badOptions.Location = cst
		if _, ok := SolarEclipseMapSVG(date, badOptions); ok {
			t.Errorf("%s: UT1 with a non-UTC location must fail", date.Format("2006-01-02"))
		}
		if stripSolarTimescaleText(utc) != stripSolarTimescaleText(ut1) {
			t.Errorf("%s: switching to UT1 changed chart geometry", date.Format("2006-01-02"))
		}
	}
}

var (
	solarTimescaleText    = regexp.MustCompile(`(?s)<text[^>]*>.*?</text>`)
	solarTimescaleMarkers = regexp.MustCompile(`(?s)<g class="solar-time-marker">.*?</g>`)
)

func stripSolarTimescaleText(doc string) string {
	return solarTimescaleText.ReplaceAllString(solarTimescaleMarkers.ReplaceAllString(doc, ""), "")
}
