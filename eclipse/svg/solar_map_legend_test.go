package svg

import (
	"strings"
	"testing"
	"time"
)

// 图例只能画一次，横版渲染器与面板不得重复绘制。
func TestSolarEclipseMapLegendRendersOnce(t *testing.T) {
	for _, test := range []struct {
		name           string
		date           time.Time
		width, height  int
		hasCentralLine bool
	}{
		{name: "landscape default", date: time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC), width: 960, height: 640, hasCentralLine: true},
		{name: "landscape wide", date: time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC), width: 1200, height: 800, hasCentralLine: true},
		{name: "portrait", date: time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC), width: 900, height: 1400, hasCentralLine: true},
		{name: "landscape non-central", date: time.Date(2014, time.April, 29, 0, 0, 0, 0, time.UTC), width: 1200, height: 800},
	} {
		t.Run(test.name, func(t *testing.T) {
			rendered, ok := SolarEclipseMapSVG(test.date, SolarEclipseMapSVGOptions{
				Width: test.width, Height: test.height, Location: time.UTC, PartialStep: 10 * time.Minute,
			})
			if !ok {
				t.Fatal("missing map")
			}
			if got := strings.Count(rendered, `class="solar-map-legend"`); got != 1 {
				t.Fatalf("legend groups = %d, want exactly one", got)
			}
			for _, label := range []string{"偏食可见区", "P/U 影锥接触"} {
				if got := strings.Count(rendered, ">"+label+"</text>"); got != 1 {
					t.Fatalf("legend label %q appears %d times, want once", label, got)
				}
			}
			if got := strings.Count(rendered, ">中心线</text>"); (got == 1) != test.hasCentralLine {
				t.Fatalf("center-line legend label appears %d times, hasCentralLine=%v", got, test.hasCentralLine)
			}
		})
	}
}
