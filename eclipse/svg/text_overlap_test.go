package svg

import (
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/svgchart"
)

// 详细版式里任意两段文字都不许压叠：包围盒口径与 internal/svgchart 一致，
// 出图尺寸要么 0 压叠，要么明确拒绝，不得返回带压叠文字的图。
type overlapTextBox struct {
	text                string
	x, y, width, height float64
	fontSize            float64
}

func svgTextFloat(attrs map[string]string, name string) float64 {
	value := strings.TrimRight(strings.TrimRight(attrs[name], "pt"), "px")
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func svgTextAnchorBox(box overlapTextBox, anchor string) overlapTextBox {
	switch anchor {
	case "middle":
		box.x -= box.width / 2
	case "end":
		box.x -= box.width
	}
	return box
}

func canvasTextBoxes(t *testing.T, document string) []overlapTextBox {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(document))
	boxes := make([]overlapTextBox, 0, 128)
	var anchor string
	var pending *overlapTextBox
	var content strings.Builder
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
			if element.Name.Local != "text" {
				continue
			}
			attrs := make(map[string]string, len(element.Attr))
			for _, attribute := range element.Attr {
				attrs[attribute.Name.Local] = attribute.Value
			}
			fontSize := svgTextFloat(attrs, "font-size")
			above, below := svgchart.EstimatedTextExtents(fontSize)
			pending = &overlapTextBox{
				x: svgTextFloat(attrs, "x"), y: svgTextFloat(attrs, "y"),
				height: above + below, fontSize: fontSize,
			}
			pending.y -= above
			anchor = attrs["text-anchor"]
			content.Reset()
		case xml.CharData:
			if pending != nil {
				content.Write([]byte(element))
			}
		case xml.EndElement:
			if element.Name.Local != "text" || pending == nil {
				continue
			}
			pending.text = content.String()
			pending.width = svgchart.EstimatedTextWidth(pending.text, pending.fontSize)
			boxes = append(boxes, svgTextAnchorBox(*pending, anchor))
			pending = nil
		}
	}
	if len(boxes) == 0 {
		t.Fatal("chart carries no text")
	}
	return boxes
}

func overlapArea(first, second overlapTextBox) float64 {
	left, right := math.Max(first.x, second.x), math.Min(first.x+first.width, second.x+second.width)
	top, bottom := math.Max(first.y, second.y), math.Min(first.y+first.height, second.y+second.height)
	if right <= left || bottom <= top {
		return 0
	}
	return (right - left) * (bottom - top)
}

func assertNoTextOverlap(t *testing.T, label, document string) {
	t.Helper()
	boxes := canvasTextBoxes(t, document)
	for first := 0; first < len(boxes); first++ {
		for second := first + 1; second < len(boxes); second++ {
			area := overlapArea(boxes[first], boxes[second])
			if area <= 0.5 {
				continue
			}
			t.Errorf("%s: %q [%.1f,%.1f w=%.1f h=%.1f] overlaps %q [%.1f,%.1f w=%.1f h=%.1f] by %.1f px²",
				label,
				boxes[first].text, boxes[first].x, boxes[first].y, boxes[first].width, boxes[first].height,
				boxes[second].text, boxes[second].x, boxes[second].y, boxes[second].width, boxes[second].height,
				area)
		}
	}
}

var eclipseOverlapSizes = [][2]int{
	{800, 560}, {800, 600}, {960, 640}, {1000, 1414}, {1414, 1000}, {1200, 800},
}

type eclipseOverlapCase struct {
	name string
	date time.Time
}

var solarOverlapCases = []eclipseOverlapCase{
	{name: "2024-04-08-total", date: time.Date(2024, time.April, 8, 18, 0, 0, 0, time.UTC)},
	{name: "2010-01-15-annular", date: time.Date(2010, time.January, 15, 7, 0, 0, 0, time.UTC)},
	{name: "2014-04-29-noncentral", date: time.Date(2014, time.April, 29, 6, 0, 0, 0, time.UTC)},
	{name: "2009-07-22-total", date: time.Date(2009, time.July, 22, 2, 0, 0, 0, time.UTC)},
}

var lunarOverlapCases = []eclipseOverlapCase{
	{name: "2025-03-14-total", date: time.Date(2025, time.March, 14, 6, 0, 0, 0, time.UTC)},
	{name: "2016-08-18-penumbral", date: time.Date(2016, time.August, 18, 12, 0, 0, 0, time.UTC)},
}

func renderSolarOverlapCase(event eclipseOverlapCase, size [2]int, options SolarEclipseMapSVGOptions) (string, bool) {
	options.Width, options.Height = size[0], size[1]
	return SolarEclipseMapSVG(event.date, options)
}

func renderLunarOverlapCase(event eclipseOverlapCase, size [2]int, options LunarEclipseDetailedSVGOptions) (string, bool) {
	options.Width, options.Height = size[0], size[1]
	return LunarEclipseDetailedSVG(event.date, options)
}

// 事件 × 尺寸矩阵：返回 ok=true 的产物必须 0 压叠，被拒绝的必须返回空串。
func TestSolarEclipseMapTextOverlapMatrix(t *testing.T) {
	accepted := make(map[string]bool)
	for _, event := range solarOverlapCases {
		for _, size := range eclipseOverlapSizes {
			name := fmt.Sprintf("%s/%dx%d", event.name, size[0], size[1])
			document, ok := renderSolarOverlapCase(event, size, SolarEclipseMapSVGOptions{})
			accepted[name] = ok
			if !ok {
				if document != "" {
					t.Errorf("%s: rejected the canvas but returned %d bytes", name, len(document))
				}
				continue
			}
			assertNoTextOverlap(t, name, document)
		}
	}
	// 文档下限与默认尺寸必须继续出图：拒绝只允许出现在真的放不下的画布上。
	for _, name := range []string{
		"2024-04-08-total/800x560", "2024-04-08-total/800x600", "2024-04-08-total/960x640",
		"2010-01-15-annular/800x560", "2014-04-29-noncentral/800x560", "2009-07-22-total/1200x800",
	} {
		if !accepted[name] {
			t.Errorf("%s: documented size was rejected", name)
		}
	}
}

func TestLunarEclipseDetailedTextOverlapMatrix(t *testing.T) {
	accepted := make(map[string]bool)
	for _, event := range lunarOverlapCases {
		for _, size := range eclipseOverlapSizes {
			name := fmt.Sprintf("%s/%dx%d", event.name, size[0], size[1])
			document, ok := renderLunarOverlapCase(event, size, LunarEclipseDetailedSVGOptions{})
			accepted[name] = ok
			if !ok {
				if document != "" {
					t.Errorf("%s: rejected the canvas but returned %d bytes", name, len(document))
				}
				continue
			}
			assertNoTextOverlap(t, name, document)
		}
	}
	for _, name := range []string{
		"2025-03-14-total/1000x1414", "2025-03-14-total/1414x1000",
		"2016-08-18-penumbral/1000x1414", "2016-08-18-penumbral/1414x1000",
	} {
		if !accepted[name] {
			t.Errorf("%s: documented size was rejected", name)
		}
	}
}

// 英文文案更长、可选图层会撑出更多图例行，两种扩展都必须保持 0 压叠。
func TestEclipseChartTextOverlapWithLongLabels(t *testing.T) {
	for _, size := range [][2]int{{800, 560}, {960, 640}, {1000, 1414}} {
		name := fmt.Sprintf("solar-en/%dx%d", size[0], size[1])
		document, ok := renderSolarOverlapCase(solarOverlapCases[0], size, SolarEclipseMapSVGOptions{Language: "en"})
		if !ok {
			t.Errorf("%s: documented size was rejected", name)
			continue
		}
		assertNoTextOverlap(t, name, document)
	}
	for _, size := range [][2]int{{800, 560}, {960, 640}} {
		name := fmt.Sprintf("solar-all-layers/%dx%d", size[0], size[1])
		document, ok := renderSolarOverlapCase(solarOverlapCases[0], size, SolarEclipseMapSVGOptions{
			GreatestTimeStep: 30 * time.Minute, PenumbralOutlineStep: time.Hour, CentralShadowStep: 10 * time.Minute,
		})
		if !ok {
			continue
		}
		assertNoTextOverlap(t, name, document)
	}
	for _, event := range lunarOverlapCases {
		for _, size := range [][2]int{{1000, 1414}, {1414, 1000}} {
			name := fmt.Sprintf("lunar-en/%s/%dx%d", event.name, size[0], size[1])
			document, ok := renderLunarOverlapCase(event, size, LunarEclipseDetailedSVGOptions{Language: "en"})
			if !ok {
				t.Errorf("%s: documented size was rejected", name)
				continue
			}
			assertNoTextOverlap(t, name, document)
		}
	}
}
