package svg

import (
	"encoding/xml"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/starainrt/astro/moon"
)

// 本文件提供解析 SVG 产物、按保守字宽估算文本占位的测试工具。

type occultationSVGText struct {
	class       string
	parent      string
	parentClass string
	value       string
	x, y        float64
	fontSize    float64
	anchor      string
}

func occultationSVGTexts(t *testing.T, diagram string) []occultationSVGText {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(diagram))
	stack := []string{}
	classes := []string{}
	texts := []occultationSVGText{}
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("decode SVG: %v", err)
		}
		switch element := token.(type) {
		case xml.StartElement:
			stack = append(stack, element.Name.Local)
			class := ""
			for _, attribute := range element.Attr {
				if attribute.Name.Local == "class" {
					class = attribute.Value
				}
			}
			classes = append(classes, class)
			if element.Name.Local != "text" {
				continue
			}
			text := occultationSVGText{class: class, fontSize: 10}
			if len(classes) >= 2 {
				text.parent = stack[len(stack)-2]
				text.parentClass = classes[len(classes)-2]
			}
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
			if len(stack) == 0 {
				t.Fatal("unbalanced SVG element")
			}
			stack = stack[:len(stack)-1]
			classes = classes[:len(classes)-1]
		case xml.CharData:
			if len(stack) == 0 || stack[len(stack)-1] != "text" || len(texts) == 0 {
				continue
			}
			texts[len(texts)-1].value += string(element)
		}
	}
	return texts
}

// occultationSVGTextBox 用偏大的字宽估算占位，避免真实字体比测试度量更宽时漏判重叠。
func occultationSVGTextBox(text occultationSVGText) [4]float64 {
	fontSize := text.fontSize
	if fontSize <= 0 {
		fontSize = 10
	}
	units := 0.0
	for _, current := range text.value {
		if current < 0x80 {
			units += 0.62
			continue
		}
		units++
		units += 0.05
	}
	width := units * fontSize
	left := text.x
	switch text.anchor {
	case "middle":
		left = text.x - width/2
	case "end":
		left = text.x - width
	}
	return [4]float64{left, text.y - fontSize*0.85, left + width, text.y + fontSize*0.3}
}

func occultationSVGTextsOutsideCanvas(texts []occultationSVGText, width, height float64) []string {
	outside := []string{}
	for _, text := range texts {
		box := occultationSVGTextBox(text)
		if box[0] < 0 || box[2] > width || box[1] < 0 || box[3] > height {
			outside = append(outside, text.value)
		}
	}
	return outside
}

func occultationSVGTextOverlapPairs(texts []occultationSVGText) int {
	pairs := 0
	for first := 0; first < len(texts); first++ {
		for second := first + 1; second < len(texts); second++ {
			a, b := occultationSVGTextBox(texts[first]), occultationSVGTextBox(texts[second])
			if a[0] < b[2] && b[0] < a[2] && a[1] < b[3] && b[1] < a[3] {
				pairs++
			}
		}
	}
	return pairs
}

// occultationSVGLabelTexts 取三族地图标签：等时线标注、中心线时刻标注、事件标记。
func occultationSVGLabelTexts(texts []occultationSVGText) []occultationSVGText {
	labels := []occultationSVGText{}
	for _, text := range texts {
		if strings.Contains(text.class, "greatest-time-isoline-label") ||
			strings.Contains(text.parentClass, "occultation-time-marker") ||
			strings.Contains(text.parentClass, "event-marker") {
			labels = append(labels, text)
		}
	}
	return labels
}

// occultationSVGDetailedPanelRows 解析详细版式数据块：返回"块标题 → 行标签 → 值"。
func occultationSVGDetailedPanelRows(t *testing.T, diagram string) map[string]map[string]string {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(diagram))
	panels := map[string]map[string]string{}
	inPanel := false
	texts := []occultationSVGText{}
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("decode SVG: %v", err)
		}
		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local == "g" {
				for _, attribute := range element.Attr {
					if attribute.Name.Local == "class" && attribute.Value == "occultation-detailed-panel" {
						inPanel, texts = true, nil
					}
				}
				continue
			}
			if !inPanel || element.Name.Local != "text" {
				continue
			}
			text := occultationSVGText{}
			for _, attribute := range element.Attr {
				if attribute.Name.Local == "text-anchor" {
					text.anchor = attribute.Value
				}
			}
			texts = append(texts, text)
		case xml.EndElement:
			if !inPanel || element.Name.Local != "g" {
				continue
			}
			inPanel = false
			rows := map[string]string{}
			title := ""
			for index := 0; index < len(texts); {
				if texts[index].anchor == "end" {
					index++
					continue
				}
				if index+1 < len(texts) && texts[index+1].anchor == "end" {
					rows[texts[index].value] = texts[index+1].value
					index += 2
					continue
				}
				if title == "" {
					title = texts[index].value
				}
				index++
			}
			panels[title] = rows
		case xml.CharData:
			if inPanel && len(texts) > 0 {
				texts[len(texts)-1].value += string(element)
			}
		}
	}
	return panels
}

func occultationSVGPanelValue(t *testing.T, panels map[string]map[string]string, title, label string) string {
	t.Helper()
	rows, ok := panels[title]
	if !ok {
		titles := make([]string, 0, len(panels))
		for current := range panels {
			titles = append(titles, current)
		}
		sort.Strings(titles)
		t.Fatalf("detailed SVG is missing panel %q, have %v", title, titles)
	}
	value, ok := rows[label]
	if !ok {
		labels := make([]string, 0, len(rows))
		for current := range rows {
			labels = append(labels, current)
		}
		sort.Strings(labels)
		t.Fatalf("panel %q is missing row %q, have %v", title, label, labels)
	}
	return value
}

// occultationSVGNumericPrefix 取 "69.2 s" / "-5.86°" 这类文本开头的数值。
func occultationSVGNumericPrefix(t *testing.T, value string) float64 {
	t.Helper()
	trimmed := strings.TrimSpace(value)
	end := 0
	for end < len(trimmed) && (trimmed[end] == '+' || trimmed[end] == '-' || trimmed[end] == '.' ||
		(trimmed[end] >= '0' && trimmed[end] <= '9')) {
		end++
	}
	parsed, err := strconv.ParseFloat(trimmed[:end], 64)
	if err != nil {
		t.Fatalf("parse number from %q: %v", value, err)
	}
	return parsed
}

// occultationSVGSexagesimal 解析 "±DD°MM'SS.S\"" 与 "HHhMMmSS.Ss" 两种六十进制文本，返回度。
func occultationSVGSexagesimal(t *testing.T, value string) float64 {
	t.Helper()
	fields := strings.FieldsFunc(value, func(current rune) bool {
		return current != '.' && (current < '0' || current > '9')
	})
	if len(fields) != 3 {
		t.Fatalf("parse sexagesimal %q: got %d fields", value, len(fields))
	}
	numbers := make([]float64, 3)
	for index, field := range fields {
		parsed, err := strconv.ParseFloat(field, 64)
		if err != nil {
			t.Fatalf("parse sexagesimal field %q of %q: %v", field, value, err)
		}
		numbers[index] = parsed
	}
	degrees := numbers[0] + numbers[1]/60 + numbers[2]/3600
	if strings.HasSuffix(strings.TrimSpace(value), "s") {
		degrees *= 15
	}
	if strings.HasPrefix(strings.TrimSpace(value), "-") {
		degrees = -degrees
	}
	return degrees
}

// occultationSVGPanelRects 取详细版式数据块自身的矩形。
func occultationSVGPanelRects(t *testing.T, diagram string) [][4]float64 {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(diagram))
	rects := [][4]float64{}
	inPanel := false
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("decode SVG: %v", err)
		}
		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local == "g" {
				inPanel = false
				for _, attribute := range element.Attr {
					if attribute.Name.Local == "class" && attribute.Value == "occultation-detailed-panel" {
						inPanel = true
					}
				}
				continue
			}
			if !inPanel || element.Name.Local != "rect" {
				continue
			}
			rect := [4]float64{}
			for _, attribute := range element.Attr {
				value, parseErr := strconv.ParseFloat(attribute.Value, 64)
				if parseErr != nil {
					continue
				}
				switch attribute.Name.Local {
				case "x":
					rect[0] = value
				case "y":
					rect[1] = value
				case "width":
					rect[2] = value
				case "height":
					rect[3] = value
				}
			}
			rects = append(rects, rect)
		case xml.EndElement:
			if element.Name.Local == "g" {
				inPanel = false
			}
		}
	}
	return rects
}

var (
	occultationTestStarOnce   sync.Once
	occultationTestStarEvent  moon.StarOccultationPath
	occultationTestSaturnOnce sync.Once
	occultationTestSaturnData moon.PlanetOccultationPath
	occultationTestMarsOnce   sync.Once
	occultationTestMarsData   moon.PlanetOccultationPath
)

// occultationTestStarPath 取一个带等时线的真实恒星掩事件。
func occultationTestStarPath(t *testing.T) moon.StarOccultationPath {
	t.Helper()
	occultationTestStarOnce.Do(func() {
		location := time.FixedZone("CST", 8*3600)
		paths, err := moon.FindStarOccultationPaths(
			time.Date(2025, 6, 5, 0, 0, 0, 0, location),
			time.Date(2025, 6, 6, 0, 0, 0, 0, location),
			hr4799StarCoordinate(),
			moon.OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200, GreatestTimeStep: 30 * time.Minute},
		)
		if err != nil || len(paths) != 1 {
			t.Fatalf("FindStarOccultationPaths() paths=%d err=%v, want one", len(paths), err)
		}
		occultationTestStarEvent = paths[0]
	})
	return occultationTestStarEvent
}

// occultationTestSaturnPath 取一个关闭密集足迹、带等时线的真实行星掩事件。
func occultationTestSaturnPath(t *testing.T) moon.PlanetOccultationPath {
	t.Helper()
	occultationTestSaturnOnce.Do(func() {
		start := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
		paths, err := moon.FindPlanetOccultationPaths(start, start.Add(24*time.Hour), moon.OccultationSaturn,
			moon.OccultationPathOptions{Step: 2 * time.Minute, DisableFootprints: true, GreatestTimeStep: 30 * time.Minute})
		if err != nil || len(paths) != 1 {
			t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
		}
		occultationTestSaturnData = paths[0]
	})
	return occultationTestSaturnData
}

// occultationTestMarsPath 取一个同时带偏掩带与全掩带的真实行星掩事件。
func occultationTestMarsPath(t *testing.T) moon.PlanetOccultationPath {
	t.Helper()
	occultationTestMarsOnce.Do(func() {
		start := time.Date(2025, time.July, 29, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*3600))
		paths, err := moon.FindPlanetOccultationPaths(start, start.Add(24*time.Hour), moon.OccultationMars,
			moon.OccultationPathOptions{
				Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
				DisableFootprints: true, IncludeFootprintTimeline: true,
				FootprintTimelineStep: 5 * time.Minute, GreatestTimeStep: 30 * time.Minute,
			})
		if err != nil || len(paths) != 1 {
			t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
		}
		occultationTestMarsData = paths[0]
	})
	return occultationTestMarsData
}
