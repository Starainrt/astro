package svg

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"
	"testing"
)

// svgElement 是一个已解析的 SVG 起始标签及其属性。
type svgElement struct {
	tag   string
	attrs map[string]string
}

func parseSVGElements(t *testing.T, value string) []svgElement {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(value))
	elements := make([]svgElement, 0)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return elements
		}
		if err != nil {
			t.Fatalf("chart is not valid XML: %v", err)
		}
		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		attrs := make(map[string]string, len(start.Attr))
		for _, attr := range start.Attr {
			attrs[attr.Name.Local] = attr.Value
		}
		elements = append(elements, svgElement{tag: start.Name.Local, attrs: attrs})
	}
}

func svgAttributeFloat(attrs map[string]string, name string) (float64, bool) {
	value, ok := attrs[name]
	if !ok {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

// assertSVGElementsInCanvas 断言全部文本与矩形都落在画布内：版面一旦按硬编码坐标排版，
// 尺寸不足时元素会静默跑到画布外而不是报错。
func assertSVGElementsInCanvas(t *testing.T, value string, width, height int) {
	t.Helper()
	texts, rects := 0, 0
	for _, element := range parseSVGElements(t, value) {
		switch element.tag {
		case "text":
			texts++
			x, hasX := svgAttributeFloat(element.attrs, "x")
			y, hasY := svgAttributeFloat(element.attrs, "y")
			if hasX && (x < 0 || x > float64(width)) {
				t.Fatalf("text x=%.3f outside the %d-wide canvas: %v", x, width, element.attrs)
			}
			if hasY && (y < 0 || y > float64(height)) {
				t.Fatalf("text y=%.3f outside the %d-high canvas: %v", y, height, element.attrs)
			}
		case "rect":
			x, hasX := svgAttributeFloat(element.attrs, "x")
			y, hasY := svgAttributeFloat(element.attrs, "y")
			w, hasW := svgAttributeFloat(element.attrs, "width")
			h, hasH := svgAttributeFloat(element.attrs, "height")
			if !hasX || !hasY || !hasW || !hasH {
				continue
			}
			rects++
			if x < -0.01 || y < -0.01 || x+w > float64(width)+0.01 || y+h > float64(height)+0.01 {
				t.Fatalf("rect (%.3f,%.3f,%.3f,%.3f) outside the %dx%d canvas", x, y, w, h, width, height)
			}
		}
	}
	if texts == 0 || rects == 0 {
		t.Fatalf("chart carries %d texts and %d rects, want a full page", texts, rects)
	}
}
