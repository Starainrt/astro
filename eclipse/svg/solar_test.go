package svg

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestLocalSolarEclipseSVG(t *testing.T) {
	svg, ok := LocalSolarEclipseSVG(
		time.Date(2024, 4, 8, 12, 0, 0, 0, time.UTC),
		-96.7970,
		32.7767,
		0,
		LocalSolarEclipseSVGOptions{Width: 640, Height: 480, Step: 5 * time.Minute},
	)
	if !ok {
		t.Fatalf("expected local solar eclipse SVG")
	}
	for _, want := range []string{"<svg", "站心日全食", "全局路径", "阶段视圆图", "黄道", "C1", "食既", "食甚", "生光", "C4", "方位", "左东右西", "UTC+8", "太阳位于", "沙罗", "全食历时", `fill="#efefed"`, `class="contact-point"`} {
		if !strings.Contains(svg, want) {
			t.Fatalf("SVG missing %q", want)
		}
	}
	for _, notWant := range []string{"中心食始", "中心食终"} {
		if strings.Contains(svg, notWant) {
			t.Fatalf("SVG should not contain %q for total eclipse", notWant)
		}
	}
	if got := strings.Count(svg, `class="event-moon"`); got != 3 {
		t.Fatalf("event moon count = %d, want 3", got)
	}
	if got := strings.Count(svg, `class="stage-moon"`); got != 5 {
		t.Fatalf("stage moon count = %d, want 5", got)
	}
	if got := strings.Count(svg, `class="event-center"`); got != 5 {
		t.Fatalf("event center count = %d, want 5", got)
	}
}

func TestLocalSolarEclipseSVGEnglishOption(t *testing.T) {
	svg, ok := LocalSolarEclipseSVG(
		time.Date(2024, 4, 8, 12, 0, 0, 0, time.UTC),
		-96.7970,
		32.7767,
		0,
		LocalSolarEclipseSVGOptions{Language: "en", Location: time.UTC},
	)
	if !ok {
		t.Fatalf("expected local solar eclipse SVG")
	}
	for _, want := range []string{"Local Solar Eclipse", "Greatest", "PA", "East is left", "Ecliptic", "Contacts (UTC)", "Sun in", "Solar Saros", "Totality"} {
		if !strings.Contains(svg, want) {
			t.Fatalf("SVG missing %q", want)
		}
	}
}

func TestLocalSolarEclipseSVGCustomText(t *testing.T) {
	svg, ok := LocalSolarEclipseSVG(
		time.Date(2024, 4, 8, 12, 0, 0, 0, time.UTC),
		-96.7970,
		32.7767,
		0,
		LocalSolarEclipseSVGOptions{
			Title:            "Custom solar title",
			SummaryText:      "Custom solar summary",
			GreatestText:     "Custom solar greatest",
			MetaText:         "Custom solar meta",
			OverviewTitle:    "Custom solar overview",
			PhasePanelsTitle: "Custom solar phases",
			ContactsTitle:    "Custom solar contacts",
			DirectionText:    "Custom solar direction",
			FooterNote:       "Custom solar footer",
		},
	)
	if !ok {
		t.Fatalf("expected local solar eclipse SVG")
	}
	for _, want := range []string{
		"Custom solar title",
		"Custom solar summary",
		"Custom solar greatest",
		"Custom solar meta",
		"Custom solar overview",
		"Custom solar phases",
		"Custom solar contacts",
		"Custom solar direction",
		"Custom solar footer",
	} {
		if !strings.Contains(svg, want) {
			t.Fatalf("SVG missing custom text %q", want)
		}
	}
}

func TestLocalSolarEclipseSVGStagePanelsShareScale(t *testing.T) {
	svg, ok := LocalSolarEclipseSVG(
		time.Date(2024, 4, 8, 12, 0, 0, 0, time.UTC),
		-96.7970,
		32.7767,
		0,
		LocalSolarEclipseSVGOptions{Width: 640, Height: 480, Step: 5 * time.Minute},
	)
	if !ok {
		t.Fatalf("expected local solar eclipse SVG")
	}
	re := regexp.MustCompile(`<circle cx="[-0-9.]+" cy="[-0-9.]+" r="([0-9.]+)" fill="url\(#se-sun\)" stroke="#c78211" stroke-width="1"/>`)
	matches := re.FindAllStringSubmatch(svg, -1)
	if got, want := len(matches), 5; got != want {
		t.Fatalf("stage sun count = %d, want %d", got, want)
	}
	first, err := strconv.ParseFloat(matches[0][1], 64)
	if err != nil {
		t.Fatalf("parse first radius: %v", err)
	}
	for i := 1; i < len(matches); i++ {
		radius, err := strconv.ParseFloat(matches[i][1], 64)
		if err != nil {
			t.Fatalf("parse radius %d: %v", i, err)
		}
		if math.Abs(radius-first) > 1e-9 {
			t.Fatalf("stage panel radii differ: first=%f current=%f", first, radius)
		}
	}
}

func TestLocalSolarEclipseSVGNoEvent(t *testing.T) {
	_, ok := LocalSolarEclipseSVG(
		time.Date(2024, 5, 15, 12, 0, 0, 0, time.UTC),
		-96.7970,
		32.7767,
		0,
		LocalSolarEclipseSVGOptions{},
	)
	if ok {
		t.Fatalf("unexpected local solar eclipse SVG for no-event date")
	}
}
