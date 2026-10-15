package svg

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/internal/svgchart"
)

func TestTransitLocalSVG(t *testing.T) {
	venus := time.Date(2012, 6, 6, 0, 0, 0, 0, time.UTC)
	chart, ok := TransitLocalSVG(TransitVenus, venus, 116.4, 39.9, 50, TransitLocalSVGOptions{Location: time.FixedZone("UTC+8", 8*3600)})
	if !ok {
		t.Fatal("Beijing 2012 render failed")
	}
	for _, want := range []string{"金星凌日站心图", `class="transit-track"`, "transit-planet-c1", "transit-planet-c2",
		"transit-planet-g", "transit-planet-c3", "transit-planet-c4", "一触 06:10:13", "四触 12:49:17"} {
		if !strings.Contains(chart, want) {
			t.Fatalf("Beijing: missing %s", want)
		}
	}
	if strings.Contains(chart, "transit-track-below-horizon") || strings.Contains(chart, "NaN") {
		t.Fatal("Beijing sees the whole transit in daylight")
	}

	// 伦敦：日出时凌日已在进行，日出前的轨迹画成地平下虚线。
	chart, ok = TransitLocalSVG(TransitVenus, venus, -0.1, 51.5, 20, TransitLocalSVGOptions{Language: "en"})
	if !ok || !strings.Contains(chart, "transit-track-below-horizon") || !strings.Contains(chart, "Sunrise during transit") ||
		!strings.Contains(chart, "Local Circumstances") {
		t.Fatal("London must show the below-horizon track and sunrise row in English")
	}

	// 1999 水星掠凌，南纬 60° 只有一触、凌甚、四触。
	chart, ok = TransitLocalSVG(TransitMercury, time.Date(1999, 11, 15, 0, 0, 0, 0, time.UTC), 150, -60, 0, TransitLocalSVGOptions{})
	if !ok || !strings.Contains(chart, "掠凌") || strings.Contains(chart, "transit-planet-c2") || strings.Contains(chart, "transit-planet-c3") {
		t.Fatal("grazing site must omit internal contacts")
	}
	if !strings.Contains(chart, "1999-11-15") {
		t.Fatal("same-day query must find the 1999-11-15 transit")
	}

	// 一、二触相隔约 2 分钟，二触标签不得被避让丢弃。
	chart, ok = TransitLocalSVG(TransitMercury, time.Date(2032, 11, 13, 0, 0, 0, 0, time.UTC), 116.4074, 39.9042, 0,
		TransitLocalSVGOptions{Location: time.FixedZone("UTC+8", 8*3600)})
	if !ok {
		t.Fatal("Beijing 2032 render failed")
	}
	for _, want := range []string{">一触 ", ">二触 ", ">凌甚 ", ">三触 ", ">四触 "} {
		if strings.Count(chart, want) != 1 {
			t.Fatalf("Beijing 2032: label %s must appear exactly once on the disk", want)
		}
	}
	for _, match := range regexp.MustCompile(`class="transit-event-label" x="([0-9.-]+)" y="([0-9.-]+)"[^>]*text-anchor="(\w+)">([^<]*)<`).FindAllStringSubmatch(chart, -1) {
		x, _ := strconv.ParseFloat(match[1], 64)
		y, _ := strconv.ParseFloat(match[2], 64)
		if box := svgchart.LabelTextBox(x, y, 11, match[4], match[3]); box.X < 30 || box.X+box.Width > 960-344 {
			t.Fatalf("Beijing 2032: label %q leaves the disk frame (box x %.1f..%.1f)", match[4], box.X, box.X+box.Width)
		}
	}
}

func TestTransitLocalSVGOptions(t *testing.T) {
	venus := time.Date(2012, 6, 6, 0, 0, 0, 0, time.UTC)
	if _, ok := TransitLocalSVG(3, venus, 0, 0, 0, TransitLocalSVGOptions{}); ok {
		t.Fatal("unsupported planet must fail")
	}
	if _, ok := TransitLocalSVG(TransitVenus, venus, 0, 0, 0, TransitLocalSVGOptions{
		TimeScale: astro.TimeScaleUT1, Location: time.FixedZone("CST", 8*3600)}); ok {
		t.Fatal("UT1 with non-UTC location must fail")
	}
	if _, ok := TransitLocalSVG(TransitVenus, venus, 0, 95, 0, TransitLocalSVGOptions{}); ok {
		t.Fatal("invalid latitude must fail")
	}
	chart, ok := TransitLocalSVG(TransitVenus, venus, 116.4, 39.9, 0, TransitLocalSVGOptions{TimeScale: astro.TimeScaleUT1, Title: "X"})
	if !ok || !strings.Contains(chart, ">X<") || !strings.Contains(chart, "UT1") {
		t.Fatal("UT1 labels and custom title")
	}
}
