package svg

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/svgchart"
)

// 调用方给图题、面板标题与页脚传超长文本时，产物必须留在画布内并在截断处给出省略号；
// 判定口径与 internal/svgchart 的 EstimatedTextWidth/EstimatedTextExtents 一致。

type footerFitText struct {
	value    string
	x, y     float64
	fontSize float64
	anchor   string
}

func footerFitLongText(prefix string) string {
	piece := prefix + "long custom text 自定义长文本，用于验证折行与截断；"
	value := ""
	for len([]rune(value)) < 640 {
		value += piece
	}
	return value
}

func footerFitTexts(t *testing.T, document string) []footerFitText {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(document))
	texts := []footerFitText{}
	stack := []string{}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("chart is not valid XML: %v", err)
		}
		switch element := token.(type) {
		case xml.StartElement:
			stack = append(stack, element.Name.Local)
			if element.Name.Local != "text" {
				continue
			}
			text := footerFitText{fontSize: 12}
			for _, attribute := range element.Attr {
				switch attribute.Name.Local {
				case "x":
					text.x, _ = strconv.ParseFloat(attribute.Value, 64)
				case "y":
					text.y, _ = strconv.ParseFloat(attribute.Value, 64)
				case "font-size":
					text.fontSize, _ = strconv.ParseFloat(attribute.Value, 64)
				case "text-anchor":
					text.anchor = attribute.Value
				}
			}
			texts = append(texts, text)
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if len(stack) > 0 && stack[len(stack)-1] == "text" && len(texts) > 0 {
				texts[len(texts)-1].value += string(element)
			}
		}
	}
	return texts
}

func footerFitViolations(t *testing.T, document string, width, height int) []string {
	t.Helper()
	violations := []string{}
	for _, text := range footerFitTexts(t, document) {
		textWidth := svgchart.EstimatedTextWidth(text.value, text.fontSize)
		left := text.x
		switch text.anchor {
		case "middle":
			left = text.x - textWidth/2
		case "end":
			left = text.x - textWidth
		}
		above, below := svgchart.EstimatedTextExtents(text.fontSize)
		if text.x < 0 || text.x > float64(width) || text.y < 0 || text.y > float64(height) {
			violations = append(violations, text.value)
			continue
		}
		if left < 0 || left+textWidth > float64(width) || text.y-above < 0 || text.y+below > float64(height) {
			violations = append(violations, text.value)
		}
	}
	return violations
}

func TestSolarEclipseMapSVGCustomLongTextStaysInsideCanvas(t *testing.T) {
	long := footerFitLongText("solar ")
	if runes := len([]rune(long)); runes < 300 || runes > 800 {
		t.Fatalf("long text runes = %d, want 300..800", runes)
	}
	date := time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC)
	for _, size := range [][2]int{{800, 560}, {960, 640}, {1000, 1414}} {
		diagram, ok := SolarEclipseMapSVG(date, SolarEclipseMapSVGOptions{
			Width: size[0], Height: size[1], Title: long, MapTitle: long, FooterNote: long,
		})
		if !ok {
			t.Fatalf("%dx%d: missing chart", size[0], size[1])
		}
		if violations := footerFitViolations(t, diagram, size[0], size[1]); len(violations) > 0 {
			t.Fatalf("%dx%d: %d texts outside the canvas, first = %q", size[0], size[1], len(violations), violations[0])
		}
		if !strings.Contains(diagram, "…") {
			t.Fatalf("%dx%d: truncated custom text is not marked", size[0], size[1])
		}
	}
}

func TestLunarEclipseDetailedSVGCustomLongTextStaysInsideCanvas(t *testing.T) {
	long := footerFitLongText("lunar ")
	date := time.Date(2025, 3, 14, 0, 0, 0, 0, time.UTC)
	// 1000x1000 是该版式能装下地图与数据块的最小高度，1000x1414 是整页开本。
	for _, size := range [][2]int{{1000, 1000}, {1000, 1414}} {
		diagram, ok := LunarEclipseDetailedSVG(date, LunarEclipseDetailedSVGOptions{
			Width: size[0], Height: size[1], Title: long, FooterNote: long,
		})
		if !ok {
			t.Fatalf("%dx%d: missing chart", size[0], size[1])
		}
		if violations := footerFitViolations(t, diagram, size[0], size[1]); len(violations) > 0 {
			t.Fatalf("%dx%d: %d texts outside the canvas, first = %q", size[0], size[1], len(violations), violations[0])
		}
		if !strings.Contains(diagram, "…") {
			t.Fatalf("%dx%d: truncated custom text is not marked", size[0], size[1])
		}
	}
}

func TestLunarEclipseMapAndSVGCustomLongTextStaysInsideCanvas(t *testing.T) {
	long := footerFitLongText("moon ")
	date := time.Date(2025, 3, 14, 0, 0, 0, 0, time.UTC)
	mapDiagram, ok := LunarEclipseMapSVG(date, LunarEclipseMapSVGOptions{
		Width: 960, Height: 640, Title: long, FooterNote: long,
	})
	if !ok {
		t.Fatal("missing lunar-eclipse map")
	}
	if violations := footerFitViolations(t, mapDiagram, 960, 640); len(violations) > 0 {
		t.Fatalf("map: %d texts outside the canvas, first = %q", len(violations), violations[0])
	}
	if !strings.Contains(mapDiagram, "…") {
		t.Fatal("map: truncated custom text is not marked")
	}
	diagram, ok := LunarEclipseSVG(date, LunarEclipseSVGOptions{Title: long, FooterNote: long})
	if !ok {
		t.Fatal("missing lunar-eclipse diagram")
	}
	if violations := footerFitViolations(t, diagram, 960, 620); len(violations) > 0 {
		t.Fatalf("diagram: %d texts outside the canvas, first = %q", len(violations), violations[0])
	}
	if !strings.Contains(diagram, "…") {
		t.Fatal("diagram: truncated custom text is not marked")
	}
}

func TestLocalSolarEclipseSVGCustomLongTextStaysInsideCanvas(t *testing.T) {
	long := footerFitLongText("local ")
	date := time.Date(2024, 4, 8, 18, 0, 0, 0, time.UTC)
	diagram, ok := LocalSolarEclipseSVG(date, -90, 35, 0, LocalSolarEclipseSVGOptions{
		Title: long, SummaryText: long, GreatestText: long, MetaText: long,
		DirectionText: long, FooterNote: long, Location: time.UTC,
	})
	if !ok {
		t.Fatal("missing local solar-eclipse chart")
	}
	if violations := footerFitViolations(t, diagram, 920, 720); len(violations) > 0 {
		t.Fatalf("%d texts outside the canvas, first = %q", len(violations), violations[0])
	}
	if !strings.Contains(diagram, "…") {
		t.Fatal("truncated custom text is not marked")
	}
}

// 自动生成的默认文案不得进入截断分支：这些产物在改动前后逐字节一致，出现省略号即说明分流写错。
func TestEclipseSVGDefaultTextIsNotTruncated(t *testing.T) {
	date := time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC)
	for _, language := range []string{"zh", "en"} {
		solar, ok := SolarEclipseMapSVG(date, SolarEclipseMapSVGOptions{Language: language})
		if !ok || strings.Contains(solar, "…") {
			t.Fatalf("solar map %s: ok=%v, default text must stay unmarked", language, ok)
		}
		lunar, ok := LunarEclipseDetailedSVG(time.Date(2025, 3, 14, 0, 0, 0, 0, time.UTC),
			LunarEclipseDetailedSVGOptions{Language: language})
		if !ok || strings.Contains(lunar, "…") {
			t.Fatalf("lunar detailed %s: ok=%v, default text must stay unmarked", language, ok)
		}
	}
}
