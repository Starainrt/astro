package svg

import (
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/svgchart"
)

// 页脚排版契约：任何文字的下沿都要离画布底边留出余量，站心图的页脚各行必须等行距。
// 这两条都是排版不变量，与具体事件无关，因此在这里按图种各取一个样本固定下来。

const chartBottomMarginMinimum = 6.0

type chartTextElement struct {
	x        float64
	y        float64
	fontSize float64
	fill     string
	value    string
}

func chartTextElements(t *testing.T, document string) []chartTextElement {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(document))
	elements := []chartTextElement{}
	stack := []string{}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("chart is not valid XML: %v", err)
		}
		switch typed := token.(type) {
		case xml.StartElement:
			stack = append(stack, typed.Name.Local)
			if typed.Name.Local != "text" {
				continue
			}
			element := chartTextElement{}
			for _, attribute := range typed.Attr {
				switch attribute.Name.Local {
				case "x":
					element.x, _ = strconv.ParseFloat(attribute.Value, 64)
				case "y":
					element.y, _ = strconv.ParseFloat(attribute.Value, 64)
				case "font-size":
					element.fontSize, _ = strconv.ParseFloat(attribute.Value, 64)
				case "fill":
					element.fill = attribute.Value
				}
			}
			elements = append(elements, element)
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if len(stack) > 0 && stack[len(stack)-1] == "text" && len(elements) > 0 {
				elements[len(elements)-1].value += string(typed)
			}
		}
	}
	return elements
}

func chartBottomMarginViolations(t *testing.T, document string, height int) []string {
	t.Helper()
	violations := []string{}
	for _, element := range chartTextElements(t, document) {
		if strings.TrimSpace(element.value) == "" {
			continue
		}
		_, below := svgchart.EstimatedTextExtents(element.fontSize)
		if margin := float64(height) - (element.y + below); margin < chartBottomMarginMinimum {
			violations = append(violations, fmt.Sprintf("%q 距底边 %.2f px（下限 %.1f）", element.value, margin, chartBottomMarginMinimum))
		}
	}
	return violations
}

func chartFooterCases(t *testing.T) []struct {
	name     string
	height   int
	document string
} {
	t.Helper()
	cst := time.FixedZone("CST", 8*3600)
	lunarDate := time.Date(2026, 3, 3, 0, 0, 0, 0, cst)
	long := footerFitLongText("footer ")
	cases := []struct {
		name     string
		height   int
		document string
	}{}
	add := func(name string, height int, document string, ok bool) {
		if !ok {
			t.Fatalf("%s: chart missing", name)
		}
		cases = append(cases, struct {
			name     string
			height   int
			document string
		}{name, height, document})
	}
	solar, ok := LocalSolarEclipseSVG(time.Date(2009, 7, 22, 12, 0, 0, 0, cst), 121.9850, 30.6167, 0,
		LocalSolarEclipseSVGOptions{Width: 923, Height: 692, Location: cst})
	add("站心日食", 692, solar, ok)
	solarLong, ok := LocalSolarEclipseSVG(time.Date(2009, 7, 22, 12, 0, 0, 0, cst), 121.9850, 30.6167, 0,
		LocalSolarEclipseSVGOptions{Width: 640, Height: 520, Location: cst, FooterNote: long})
	add("站心日食 窄画布长说明", 520, solarLong, ok)
	lunar, ok := LunarEclipseSVG(lunarDate, LunarEclipseSVGOptions{Width: 960, Height: 640, Language: "zh", Location: cst})
	add("站心月食", 640, lunar, ok)
	lunarLong, ok := LunarEclipseSVG(lunarDate, LunarEclipseSVGOptions{Width: 640, Height: 520, Language: "zh", Location: cst, FooterNote: long})
	add("站心月食 窄画布长说明", 520, lunarLong, ok)
	detailed, ok := LunarEclipseDetailedSVG(lunarDate, LunarEclipseDetailedSVGOptions{Width: 1000, Height: 1000, Language: "zh", Location: cst, FooterNote: long})
	add("月食详图 长说明", 1000, detailed, ok)
	mapChart, ok := LunarEclipseMapSVG(lunarDate, LunarEclipseMapSVGOptions{Width: 960, Height: 640, Language: "zh", Location: cst, FooterNote: long})
	add("月食地图 长说明", 640, mapChart, ok)
	solarMap, ok := SolarEclipseMapSVG(time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC),
		SolarEclipseMapSVGOptions{Width: 960, Height: 640, Location: time.UTC, FooterNote: long})
	add("日食地图 长说明", 640, solarMap, ok)
	return cases
}

func TestChartTextKeepsBottomMargin(t *testing.T) {
	for _, tc := range chartFooterCases(t) {
		if violations := chartBottomMarginViolations(t, tc.document, tc.height); len(violations) > 0 {
			t.Errorf("%s：%d 行文字贴到画布底边，首条 %s", tc.name, len(violations), violations[0])
		}
	}
}

func TestLocalDiagramFooterLinesAreEvenlySpaced(t *testing.T) {
	cst := time.FixedZone("CST", 8*3600)
	cases := []struct {
		name     string
		height   int
		document string
	}{}
	solar, ok := LocalSolarEclipseSVG(time.Date(2009, 7, 22, 12, 0, 0, 0, cst), 121.9850, 30.6167, 0,
		LocalSolarEclipseSVGOptions{Width: 923, Height: 692, Location: cst})
	if !ok {
		t.Fatal("站心日食：缺图")
	}
	cases = append(cases, struct {
		name     string
		height   int
		document string
	}{"站心日食", 692, solar})
	lunar, ok := LunarEclipseSVG(time.Date(2026, 3, 3, 0, 0, 0, 0, cst), LunarEclipseSVGOptions{Width: 960, Height: 640, Language: "zh", Location: cst})
	if !ok {
		t.Fatal("站心月食：缺图")
	}
	cases = append(cases, struct {
		name     string
		height   int
		document string
	}{"站心月食", 640, lunar})

	for _, tc := range cases {
		baselines := []float64{}
		for _, element := range chartTextElements(t, tc.document) {
			if element.fontSize != 12 || (element.fill != "#333" && element.fill != "#555") {
				continue
			}
			if strings.TrimSpace(element.value) == "" {
				continue
			}
			baselines = append(baselines, element.y)
		}
		if len(baselines) < 3 {
			t.Fatalf("%s：页脚只有 %d 行，样本太少", tc.name, len(baselines))
		}
		sort.Float64s(baselines)
		step := svgchart.FooterLineHeight(12)
		for index := 1; index < len(baselines); index++ {
			if gap := baselines[index] - baselines[index-1]; math.Abs(gap-step) > 0.01 {
				t.Fatalf("%s：页脚第 %d 行与上一行间距 %.3f，应为 %.1f（基线 %v）", tc.name, index+1, gap, step, baselines)
			}
		}
		if margin := float64(tc.height) - baselines[len(baselines)-1]; math.Abs(margin-localDiagramFooterBottomPadding) > 0.01 {
			t.Fatalf("%s：页脚末行基线距底边 %.3f，应为 %.1f（须落在图框内）", tc.name, margin, localDiagramFooterBottomPadding)
		}
	}
}

// 页脚时标声明契约：调用方给超长自定义说明时，时标声明仍须留在页脚里（截断可以吃掉说明，不能吃掉声明）。
func TestChartFooterKeepsTimeScaleDeclaration(t *testing.T) {
	long := footerFitLongText("scale ")
	cst := time.FixedZone("CST", 8*3600)
	lunarDate := time.Date(2026, 3, 3, 0, 0, 0, 0, cst)
	cases := []struct {
		name     string
		document string
		ok       bool
	}{}
	solar, ok := LocalSolarEclipseSVG(time.Date(2012, 5, 21, 12, 0, 0, 0, cst), 118.0894, 24.4798, 0,
		LocalSolarEclipseSVGOptions{Width: 920, Height: 720, Location: cst, FooterNote: long})
	cases = append(cases, struct {
		name     string
		document string
		ok       bool
	}{"站心日食", solar, ok})
	lunar, ok := LunarEclipseSVG(lunarDate, LunarEclipseSVGOptions{Width: 960, Height: 640, Language: "zh", Location: cst, FooterNote: long})
	cases = append(cases, struct {
		name     string
		document string
		ok       bool
	}{"站心月食", lunar, ok})
	mapChart, ok := LunarEclipseMapSVG(lunarDate, LunarEclipseMapSVGOptions{Width: 960, Height: 640, Language: "zh", Location: cst, FooterNote: long})
	cases = append(cases, struct {
		name     string
		document string
		ok       bool
	}{"月食地图", mapChart, ok})
	for _, tc := range cases {
		if !tc.ok {
			t.Fatalf("%s：缺图", tc.name)
		}
		if !strings.Contains(tc.document, "显示时区") && !strings.Contains(tc.document, "图中时刻为 UTC") {
			t.Errorf("%s：超长说明把时标声明挤掉了", tc.name)
		}
	}
}
