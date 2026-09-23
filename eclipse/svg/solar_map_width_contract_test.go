package svg

import (
	"strings"
	"testing"
	"time"
)

// 带宽无定义时地图面板与摘要行都按目录口径留空，不得打印发散的解析值。

func TestSolarEclipseMapHidesUndefinedPathWidth(t *testing.T) {
	testCases := []struct {
		name       string
		date       time.Time
		panelValue string
		summary    string
		forbidden  []string
	}{
		{
			name:       "-1404-01-07 single-sided limit",
			date:       time.Date(-1404, 1, 7, 0, 0, 0, 0, time.UTC),
			panelValue: "—",
			forbidden:  []string{"20692", "中心食带宽 "},
		},
		{
			name:       "2003-05-31 single-sided limit with limit lines",
			date:       time.Date(2003, 5, 31, 0, 0, 0, 0, time.UTC),
			panelValue: "—",
			forbidden:  []string{"4454", "中心食带宽 "},
		},
		{
			name:       "2024-04-08 total",
			date:       time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC),
			panelValue: "198.6 km",
			summary:    "中心食带宽 198.6 km",
		},
		{
			name:       "2010-01-15 annular",
			date:       time.Date(2010, 1, 15, 0, 0, 0, 0, time.UTC),
			panelValue: "335.0 km",
			summary:    "中心食带宽 335.0 km",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			diagram, ok := SolarEclipseMapSVG(tc.date, SolarEclipseMapSVGOptions{
				Width: 900, Height: 620, Location: time.UTC, PartialStep: 10 * time.Minute,
			})
			if !ok {
				t.Fatalf("expected map")
			}
			value, found := solarEclipsePanelRowValue(diagram, "中心食带宽")
			if !found {
				t.Fatalf("panel row missing")
			}
			if value != tc.panelValue {
				t.Fatalf("panel row %q want %q", value, tc.panelValue)
			}
			if tc.summary != "" && !strings.Contains(diagram, tc.summary) {
				t.Fatalf("summary missing %q", tc.summary)
			}
			for _, unwanted := range tc.forbidden {
				if strings.Contains(diagram, unwanted) {
					t.Fatalf("diagram contains %q", unwanted)
				}
			}
			if err := validateEclipseMapXML(diagram); err != nil {
				t.Fatalf("map is not valid XML: %v", err)
			}
		})
	}
}

func solarEclipsePanelRowValue(diagram, label string) (string, bool) {
	index := strings.Index(diagram, ">"+label+"</text>")
	if index < 0 {
		return "", false
	}
	rest := diagram[index+len(label)+len("></text>"):]
	start := strings.Index(rest, ">")
	if start < 0 {
		return "", false
	}
	rest = rest[start+1:]
	end := strings.Index(rest, "</text>")
	if end < 0 {
		return "", false
	}
	return rest[:end], true
}
