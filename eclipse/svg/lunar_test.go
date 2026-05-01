package svg

import (
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestLunarEclipseSVG(t *testing.T) {
	svg, ok := LunarEclipseSVG(
		time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
		LunarEclipseSVGOptions{Width: 720, Height: 480, Step: 10 * time.Minute},
	)
	if !ok {
		t.Fatalf("expected lunar eclipse SVG")
	}
	for _, want := range []string{"<svg", "月全食", "黄道", "赤经", "黄经", "狮子座", "方位", "P1", "U1", "U2", "食甚", "U3", "U4", "P4", "UTC+8", "沙罗"} {
		if !strings.Contains(svg, want) {
			t.Fatalf("SVG missing %q", want)
		}
	}
	if got := strings.Count(svg, `class="event-moon"`); got != 7 {
		t.Fatalf("event moon count = %d, want 7", got)
	}
}

func TestLunarEclipseSVGEnglishOption(t *testing.T) {
	svg, ok := LunarEclipseSVG(
		time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
		LunarEclipseSVGOptions{Language: "en", Location: time.UTC},
	)
	if !ok {
		t.Fatalf("expected lunar eclipse SVG")
	}
	for _, want := range []string{"Total Lunar Eclipse", "Ecliptic", "Moon: RA", "ecl.lon", "PA", "Greatest", "Contacts (UTC)", "Lunar Saros"} {
		if !strings.Contains(svg, want) {
			t.Fatalf("SVG missing %q", want)
		}
	}
}

func TestLunarEclipseSVGCustomText(t *testing.T) {
	svg, ok := LunarEclipseSVG(
		time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC),
		LunarEclipseSVGOptions{
			Title:           "Custom lunar title",
			SummaryText:     "Custom lunar summary",
			MaximumText:     "Custom lunar maximum",
			CoordinatesText: "Custom lunar coordinates",
			DurationText:    "Custom lunar duration",
			MetaText:        "Custom lunar meta",
			ContactsTitle:   "Custom lunar contacts",
			DirectionText:   "Custom lunar direction",
			FooterNote:      "Custom lunar footer",
		},
	)
	if !ok {
		t.Fatalf("expected lunar eclipse SVG")
	}
	for _, want := range []string{
		"Custom lunar title",
		"Custom lunar summary",
		"Custom lunar maximum",
		"Custom lunar coordinates",
		"Custom lunar duration",
		"Custom lunar meta",
		"Custom lunar contacts",
		"Custom lunar direction",
		"Custom lunar footer",
	} {
		if !strings.Contains(svg, want) {
			t.Fatalf("SVG missing custom text %q", want)
		}
	}
}

func TestLunarEclipseSVGNoEvent(t *testing.T) {
	_, ok := LunarEclipseSVG(time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC), LunarEclipseSVGOptions{})
	if ok {
		t.Fatalf("unexpected lunar eclipse SVG for no-event date")
	}
}

func TestLunarEclipseSVGEventPointsExpandMergedLabels(t *testing.T) {
	events := lunarEclipseSVGEventPoints([]lunarEclipseSVGPoint{
		{
			LunarEclipseDiagramPoint: basic.LunarEclipseDiagramPoint{
				Label:  "Greatest",
				Labels: []string{"U2", "Greatest", "U3"},
			},
		},
	})
	if got, want := len(events), 3; got != want {
		t.Fatalf("event point count = %d, want %d", got, want)
	}
	want := []string{"U2", "Greatest", "U3"}
	for i, label := range want {
		if events[i].Label != label {
			t.Fatalf("event labels = %#v, want %v", []string{events[0].Label, events[1].Label, events[2].Label}, want)
		}
	}
}
