package svg

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestTransitSVGGlobalPartial(tester *testing.T) {
	date := time.Date(2854, 12, 14, 0, 0, 0, 0, time.UTC)
	chart, ok := TransitMapSVG(TransitVenus, date, TransitMapSVGOptions{Isochrones: TransitIsochroneAll, IsochroneStep: 30 * time.Second})
	if !ok {
		tester.Fatal("global partial transit render failed")
	}
	for _, expected := range []string{"2854-12-14", "transit-partial-region", "地心无接触", "周期族 4", "γ ", "c1-isochrone", "c4-isochrone"} {
		if !strings.Contains(chart, expected) {
			tester.Fatalf("missing %s", expected)
		}
	}
	for _, absent := range []string{"NaN", "Inf", "transit-sun-c1-shape", "c2-isochrone", "c3-isochrone"} {
		if strings.Contains(chart, absent) {
			tester.Fatalf("invalid SVG contains %s", absent)
		}
	}
	local, ok := TransitLocalSVG(TransitVenus, date, -135, -60, 0, TransitLocalSVGOptions{})
	if !ok || !strings.Contains(local, "transit-planet-c1") || !strings.Contains(local, "transit-planet-c4") || strings.Contains(local, "transit-planet-c2") {
		tester.Fatal("local partial-transit diagram has invalid contacts")
	}
	labels := regexp.MustCompile(`class="transit-event-label" x="[0-9.-]+" y="([0-9.-]+)"`).FindAllStringSubmatch(local, -1)
	if len(labels) != 3 {
		tester.Fatalf("grazing contact labels=%d, want 3", len(labels))
	}
	footer := regexp.MustCompile(`<text x="40" y="([0-9.-]+)" fill="#555"`).FindAllStringSubmatch(local, -1)
	if len(footer) == 0 {
		tester.Fatal("local partial-transit diagram has no footer")
	}
	firstFooterBaseline, _ := strconv.ParseFloat(footer[0][1], 64)
	for _, label := range labels {
		baseline, _ := strconv.ParseFloat(label[1], 64)
		if baseline > firstFooterBaseline-8 {
			tester.Fatal("contact label overlaps the footer")
		}
	}
	if _, ok := TransitLocalSVG(TransitVenus, date, 0, 45, 0, TransitLocalSVGOptions{}); ok {
		tester.Fatal("a northern site must miss the transit")
	}
}

func TestTransitMapLegendListsDrawnIsochrones(tester *testing.T) {
	for _, test := range []struct {
		planet TransitPlanet
		date   time.Time
		inner  bool
	}{
		{TransitVenus, time.Date(554, 5, 22, 0, 0, 0, 0, time.UTC), true},
		{TransitMercury, time.Date(-513, 10, 22, 0, 0, 0, 0, time.UTC), false},
	} {
		chart, ok := TransitMapSVG(test.planet, test.date, TransitMapSVGOptions{Isochrones: TransitIsochroneAll, Language: "en"})
		if !ok {
			tester.Fatalf("%v: render failed", test.date)
		}
		for _, marker := range []string{"c2-isochrone", "C2 isochrones", "C3 isochrones"} {
			if strings.Contains(chart, marker) != test.inner {
				tester.Fatalf("%v: %s present=%v, want %v", test.date, marker, !test.inner, test.inner)
			}
		}
	}
}

// 地心无接触的部分凌日也要在图例里列出实际画出的一触、四触等时线；二触、三触不存在。
func TestTransitPartialLegendListsDrawnIsochrones(tester *testing.T) {
	for _, language := range []string{"zh", "en"} {
		chart, ok := TransitMapSVG(TransitVenus, time.Date(2854, 12, 14, 0, 0, 0, 0, time.UTC), TransitMapSVGOptions{
			Isochrones: TransitIsochroneAll, Language: language,
		})
		if !ok {
			tester.Fatalf("%s: render failed", language)
		}
		partial, c1, c4 := "部分凌日可见区", "一触等时线", "四触等时线"
		if language == "en" {
			partial, c1, c4 = "Partial transit visible", "C1 isochrones", "C4 isochrones"
		}
		for _, expected := range []string{"c1-isochrone", "c4-isochrone", partial, c1, c4} {
			if !strings.Contains(chart, expected) {
				tester.Fatalf("%s: partial chart is missing %q", language, expected)
			}
		}
		for _, absent := range []string{"c2-isochrone", "c3-isochrone", "二触等时线", "三触等时线", "C2 isochrones", "C3 isochrones"} {
			if strings.Contains(chart, absent) {
				tester.Fatalf("%s: partial chart must not contain %q", language, absent)
			}
		}
	}
	// 默认不画等时线时图例只留可见区色块，不凭空加等时线条目。
	chart, ok := TransitMapSVG(TransitVenus, time.Date(2854, 12, 14, 0, 0, 0, 0, time.UTC), TransitMapSVGOptions{Language: "en"})
	if !ok {
		tester.Fatal("default render failed")
	}
	for _, absent := range []string{"c1-isochrone", "C1 isochrones", "C4 isochrones"} {
		if strings.Contains(chart, absent) {
			tester.Fatalf("default chart must not contain %q", absent)
		}
	}
}
