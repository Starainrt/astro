package svg

import (
	"testing"
	"time"

	"github.com/starainrt/astro/moon"
)

// 总览区接触标签契约：掠掩时四次接触只差几十角秒，标签必须自动避让排开，既不互相压叠、也不压到方位字、
// 月球圆盘与右侧接触时刻栏；被推远的标签要用引线连回接触点。2026-09-14 保山月掩金星就是这种掠掩样本。

const localOccultationOverviewLabelFill = "#792d28"

func localPlanetOccultationOverviewLabels(t *testing.T, diagram string) []occultationSVGText {
	t.Helper()
	labels := []occultationSVGText{}
	for _, text := range occultationSVGTexts(t, diagram) {
		if text.fill == localOccultationOverviewLabelFill {
			labels = append(labels, text)
		}
	}
	return labels
}

func localOccultationOverviewLabelOverlaps(labels []occultationSVGText) [][2]string {
	overlaps := [][2]string{}
	for first := 0; first < len(labels); first++ {
		for second := first + 1; second < len(labels); second++ {
			a := occultationSVGTextBox(labels[first])
			b := occultationSVGTextBox(labels[second])
			if a[0] < b[2] && b[0] < a[2] && a[1] < b[3] && b[1] < a[3] {
				overlaps = append(overlaps, [2]string{labels[first].value, labels[second].value})
			}
		}
	}
	return overlaps
}

func grazingVenusOccultation(t *testing.T) moon.PlanetOccultationInfo {
	t.Helper()
	location := time.FixedZone("CST", 8*3600)
	events, err := moon.FindPlanetOccultations(
		time.Date(2026, time.September, 14, 0, 0, 0, 0, location),
		time.Date(2026, time.September, 15, 0, 0, 0, 0, location),
		moon.OccultationVenus, 99.1611, 25.1120, 1650, moon.OccultationSearchOptions{},
	)
	if err != nil {
		t.Fatalf("FindPlanetOccultations() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("FindPlanetOccultations() returned %d events, want 1", len(events))
	}
	return events[0]
}

func TestLocalPlanetOccultationOverviewLabelsAvoidEachOther(t *testing.T) {
	location := time.FixedZone("CST", 8*3600)
	for _, tc := range []struct {
		name  string
		info  moon.PlanetOccultationInfo
		width int
		heigh int
	}{
		{"保山掠掩金星", grazingVenusOccultation(t), 935, 683},
		{"土星全掩", localSaturnOccultation(t), 920, 720},
	} {
		diagram, err := LocalPlanetOccultationSVG(tc.info, LocalPlanetOccultationSVGOptions{
			Width: tc.width, Height: tc.heigh, Location: location,
		})
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		labels := localPlanetOccultationOverviewLabels(t, diagram)
		if len(labels) < 2 {
			t.Fatalf("%s：总览区只有 %d 个接触标签", tc.name, len(labels))
		}
		if overlaps := localOccultationOverviewLabelOverlaps(labels); len(overlaps) > 0 {
			t.Fatalf("%s：接触标签互相压叠 %v", tc.name, overlaps)
		}
		for _, text := range labels {
			box := occultationSVGTextBox(text)
			if box[0] < 0 || box[2] > float64(tc.width) || box[1] < 0 || box[3] > float64(tc.heigh) {
				t.Fatalf("%s：标签「%s」越出画布 %v", tc.name, text.value, box)
			}
		}
	}
}
