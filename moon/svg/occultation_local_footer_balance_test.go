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

// 本地月掩页脚契约：方向说明与补充说明（含时标声明）都排在可用宽度内，同一块各行宽度接近，
// 不留孤立短尾行，且显示时区的偏移只写一次。

const (
	localOccultationFooterDirectionFill = "#4f595b"
	localOccultationFooterNoteFill      = "#747c7d"
)

type localOccultationFooterText struct {
	fill     string
	fontSize float64
	value    string
}

func localOccultationFooterTexts(t *testing.T, diagram string) []localOccultationFooterText {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(diagram))
	texts := []localOccultationFooterText{}
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
		text := localOccultationFooterText{}
		for _, attribute := range element.Attr {
			switch attribute.Name.Local {
			case "fill":
				text.fill = attribute.Value
			case "font-size":
				text.fontSize, _ = strconv.ParseFloat(attribute.Value, 64)
			}
		}
		if text.fill != localOccultationFooterDirectionFill && text.fill != localOccultationFooterNoteFill {
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

func localOccultationFooterWidths(lines []localOccultationFooterText) ([]string, []float64) {
	values := make([]string, 0, len(lines))
	widths := make([]float64, 0, len(lines))
	for _, line := range lines {
		values = append(values, line.value)
		widths = append(widths, svgchart.EstimatedTextWidth(line.value, line.fontSize))
	}
	return values, widths
}

func localOccultationFooterSpread(lines []localOccultationFooterText) float64 {
	_, widths := localOccultationFooterWidths(lines)
	minimum, maximum := widths[0], widths[0]
	for _, width := range widths {
		minimum = math.Min(minimum, width)
		maximum = math.Max(maximum, width)
	}
	return maximum / minimum
}

func localOccultationFooterGroup(lines []localOccultationFooterText, fill string) []localOccultationFooterText {
	group := []localOccultationFooterText{}
	for _, line := range lines {
		if line.fill == fill {
			group = append(group, line)
		}
	}
	return group
}

// 说明块各行的最大最小宽度比不得超过 15%，整块最短行不得短于最长行的 60%（旧实现里时标声明独占的短尾行只有 25%~47%）。
func TestLocalOccultationFooterLinesAreEven(t *testing.T) {
	location := time.FixedZone("UTC+08:00", 8*3600)
	const width, height = 920, 720
	for _, language := range []string{"zh", "en"} {
		star, err := LocalStarOccultationSVG(localHR4799Occultation(t), hr4799StarCoordinate(), LocalStarOccultationSVGOptions{
			Language: language, Location: location, Width: width, Height: height,
		})
		if err != nil {
			t.Fatalf("star %s: %v", language, err)
		}
		planet, err := LocalPlanetOccultationSVG(localSaturnOccultation(t), LocalPlanetOccultationSVGOptions{
			Language: language, Location: location, Width: width, Height: height,
		})
		if err != nil {
			t.Fatalf("planet %s: %v", language, err)
		}
		for name, diagram := range map[string]string{"star": star, "planet": planet} {
			lines := localOccultationFooterTexts(t, diagram)
			if len(lines) < 2 {
				t.Fatalf("%s %s: footer lines = %#v, want the direction and note blocks", name, language, lines)
			}
			values, widths := localOccultationFooterWidths(lines)
			for index, line := range lines {
				if limit := float64(width) - 80; widths[index] > limit {
					t.Fatalf("%s %s: footer line %q width %.1f exceeds %.1f", name, language, line.value, widths[index], limit)
				}
				if strings.Contains(line.value, "…") {
					t.Fatalf("%s %s: default footer text is truncated: %q", name, language, line.value)
				}
			}
			note := localOccultationFooterGroup(lines, localOccultationFooterNoteFill)
			if len(note) == 0 {
				t.Fatalf("%s %s: missing the footer note block", name, language)
			}
			// 说明块与方向说明行字号不同（fs9 / fs10），故按块判 1.15；整块只兜"孤尾短行"（1.60）。
			if spread := localOccultationFooterSpread(note); spread > 1.15 {
				t.Fatalf("%s %s: note block spread = %.3f, want <= 1.15 (%q)", name, language, spread, values)
			}
			if spread := localOccultationFooterSpread(lines); spread > 1.60 {
				t.Fatalf("%s %s: footer spread = %.3f, want no orphan short line (%q)", name, language, spread, values)
			}
			joined := strings.Join(values, " ")
			if zone := "显示时区"; language == "en" {
				if !strings.Contains(joined, "shown in") {
					t.Fatalf("%s %s: footer note lost the time zone declaration: %q", name, language, joined)
				}
			} else if !strings.Contains(joined, zone) {
				t.Fatalf("%s %s: footer note lost the time zone declaration: %q", name, language, joined)
			}
		}
	}
}

// 偏移型时区名（UTC+08:00）不得让时标声明写成“显示时区 UTC+08:00，UTC+08:00”。
func TestLocalOccultationFooterNamesTheZoneOnce(t *testing.T) {
	location := time.FixedZone("UTC+08:00", 8*3600)
	for _, language := range []string{"zh", "en"} {
		star, err := LocalStarOccultationSVG(localHR4799Occultation(t), hr4799StarCoordinate(), LocalStarOccultationSVGOptions{
			Language: language, Location: location,
		})
		if err != nil {
			t.Fatalf("star %s: %v", language, err)
		}
		planet, err := LocalPlanetOccultationSVG(localSaturnOccultation(t), LocalPlanetOccultationSVGOptions{
			Language: language, Location: location,
		})
		if err != nil {
			t.Fatalf("planet %s: %v", language, err)
		}
		for name, diagram := range map[string]string{"star": star, "planet": planet} {
			note := localOccultationFooterGroup(localOccultationFooterTexts(t, diagram), localOccultationFooterNoteFill)
			values, _ := localOccultationFooterWidths(note)
			joined := strings.Join(values, " ")
			if got := strings.Count(joined, "UTC+08:00"); got != 1 {
				t.Fatalf("%s %s: footer note names UTC+08:00 %d times, want once: %q", name, language, got, joined)
			}
		}
	}
}
