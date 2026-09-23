package svg

import (
	"encoding/xml"
	"io"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/svgchart"
)

// 站心日食页脚契约：方向说明与补充说明（含时标声明）同字号排成三行，行宽接近、不超可用宽度，
// 且显示时区的偏移只写一次。

const (
	solarEclipseLocalFooterDirectionFill = "#333"
	solarEclipseLocalFooterNoteFill      = "#555"
	// 页脚与页眉同色，靠字号区分。
	solarEclipseLocalFooterFontSize = 12
)

type solarEclipseLocalFooterText struct {
	fill     string
	fontSize float64
	value    string
}

func solarEclipseLocalFooterTexts(t *testing.T, diagram string) []solarEclipseLocalFooterText {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(diagram))
	texts := []solarEclipseLocalFooterText{}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("decode SVG: %v", err)
		}
		element, ok := token.(xml.StartElement)
		if !ok || element.Name.Local != "text" {
			continue
		}
		text := solarEclipseLocalFooterText{}
		for _, attribute := range element.Attr {
			switch attribute.Name.Local {
			case "fill":
				text.fill = attribute.Value
			case "font-size":
				text.fontSize, _ = strconv.ParseFloat(attribute.Value, 64)
			}
		}
		if text.fontSize != solarEclipseLocalFooterFontSize {
			continue
		}
		if text.fill != solarEclipseLocalFooterDirectionFill && text.fill != solarEclipseLocalFooterNoteFill {
			continue
		}
		var content string
		if err := decoder.DecodeElement(&content, &element); err != nil {
			t.Fatalf("decode text element: %v", err)
		}
		text.value = content
		texts = append(texts, text)
	}
	return texts
}

func solarEclipseLocalFooterDiagram(t *testing.T, language string, location *time.Location) string {
	t.Helper()
	diagram, ok := LocalSolarEclipseSVG(
		time.Date(2012, time.May, 21, 6, 10, 0, 0, time.UTC),
		118.0894, 24.4798, 0,
		LocalSolarEclipseSVGOptions{Width: 920, Height: 720, Language: language, Location: location},
	)
	if !ok {
		t.Fatalf("%s %s: no local solar eclipse diagram", language, location)
	}
	return diagram
}

func TestLocalSolarEclipseFooterIsEvenAndNamesTheZoneOnce(t *testing.T) {
	const width = 920
	for _, location := range []*time.Location{
		time.FixedZone("UTC+08:00", 8*3600),
		time.FixedZone("CST", 8*3600),
	} {
		for _, language := range []string{"zh", "en"} {
			diagram := solarEclipseLocalFooterDiagram(t, language, location)
			lines := solarEclipseLocalFooterTexts(t, diagram)
			if len(lines) != 3 {
				t.Fatalf("%s %s: footer lines = %d, want the direction line plus two note lines (%#v)",
					language, location, len(lines), lines)
			}
			minimum, maximum := math.Inf(1), 0.0
			note := []string{}
			for _, line := range lines {
				lineWidth := svgchart.EstimatedTextWidth(line.value, line.fontSize)
				minimum, maximum = math.Min(minimum, lineWidth), math.Max(maximum, lineWidth)
				if lineWidth > width-80 {
					t.Fatalf("%s %s: footer line %q width %.1f exceeds %d", language, location, line.value, lineWidth, width-80)
				}
				if strings.Contains(line.value, "…") {
					t.Fatalf("%s %s: default footer text is truncated: %q", language, location, line.value)
				}
				if line.fill == solarEclipseLocalFooterNoteFill {
					note = append(note, line.value)
				}
			}
			if spread := maximum / minimum; spread > 1.15 {
				t.Fatalf("%s %s: footer spread = %.3f, want <= 1.15 (%#v)", language, location, spread, lines)
			}
			noteText := strings.Join(note, " ")
			if language == "en" {
				if !strings.Contains(noteText, "shown in") {
					t.Fatalf("en %s: footer note lost the time zone declaration: %q", location, noteText)
				}
			} else if !strings.Contains(noteText, "显示时区") {
				t.Fatalf("zh %s: footer note lost the time zone declaration: %q", location, noteText)
			}
			if got := strings.Count(noteText, "UTC+08:00"); got != 1 {
				t.Fatalf("%s %s: footer note names UTC+08:00 %d times, want once: %q", language, location, got, noteText)
			}
		}
	}
}
