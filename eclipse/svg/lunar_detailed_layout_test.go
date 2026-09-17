package svg

import (
	"testing"
	"time"
)

// 详细版式必须按 Height 推导版面：放得下的尺寸不得有任何 text/rect 越界，放不下的尺寸必须明确拒绝。
func TestLunarEclipseDetailedSVGLayoutFitsCanvas(t *testing.T) {
	date := time.Date(2025, time.March, 14, 0, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name          string
		width, height int
		accepted      bool
	}{
		{name: "small landscape", width: 640, height: 420},
		{name: "small landscape 800", width: 800, height: 600},
		{name: "nasa page", width: 1000, height: 1414, accepted: true},
		{name: "wide page", width: 1414, height: 1000, accepted: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			rendered, ok := LunarEclipseDetailedSVG(date, LunarEclipseDetailedSVGOptions{
				Width: test.width, Height: test.height, Location: time.UTC,
			})
			if ok != test.accepted {
				t.Fatalf("%dx%d accepted = %v, want %v", test.width, test.height, ok, test.accepted)
			}
			if !test.accepted {
				if rendered != "" {
					t.Fatalf("%dx%d rejected the canvas but returned %d bytes", test.width, test.height, len(rendered))
				}
				return
			}
			assertSVGElementsInCanvas(t, rendered, test.width, test.height)
		})
	}

	// 零值走文档默认开本，版式必须落在画布内。
	defaulted, ok := LunarEclipseDetailedSVG(date, LunarEclipseDetailedSVGOptions{Location: time.UTC})
	if !ok {
		t.Fatal("default canvas was rejected")
	}
	assertSVGElementsInCanvas(t, defaulted, 1000, 1414)
}
