package svg

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro"
	eclipsecore "github.com/starainrt/astro/eclipse"
)

// UT1 出图：换算不能碰零值（零值表示"该阶段不存在"），且非 UTC 时区应明确失败而不是画错图。
func TestLunarEclipseUT1LabelsPreserveZeroPhases(t *testing.T) {
	info := eclipsecore.LunarEclipseInfo{
		Maximum: time.Date(2025, 9, 7, 18, 11, 0, 0, time.UTC),
	}
	got := eclipsecore.LunarEclipseInfoInUT1(info)
	if !got.TotalStart.IsZero() {
		t.Fatalf("不存在的阶段应保持零值, got %v", got.TotalStart)
	}
	if delta := got.Maximum.Sub(info.Maximum).Seconds(); delta <= 0 {
		t.Fatalf("UT1 时刻应领先民用时刻, got %g 秒", delta)
	}
}

// assertUT1Declaration 钉住 UT1 声明必须落在 <text> 元素内，并写出 DUT1 差值。
func assertUT1Declaration(t *testing.T, name, svg string) {
	t.Helper()
	declaration := "DUT1 = UT1−UTC = "
	idx := strings.Index(svg, declaration)
	if idx < 0 {
		t.Errorf("%s 应在图内声明 UT1 并写出 DUT1 差值", name)
		return
	}
	rest := svg[idx:]
	end := strings.Index(rest, "</text>")
	if end < 0 || strings.Contains(rest[:end], "<") {
		t.Errorf("%s 的 UT1 声明必须落在 <text> 元素内", name)
	}
}

func TestLunarEclipseMapSVGTimeScaleOption(t *testing.T) {
	date := time.Date(2025, time.September, 7, 12, 0, 0, 0, time.UTC)
	utcSVG, ok := LunarEclipseMapSVG(date, LunarEclipseMapSVGOptions{})
	if !ok {
		t.Fatal("UTC 渲染应成功")
	}
	ut1SVG, ok := LunarEclipseMapSVG(date, LunarEclipseMapSVGOptions{TimeScale: astro.TimeScaleUT1})
	if !ok {
		t.Fatal("UT1 渲染应成功")
	}
	assertUT1Declaration(t, "月食全球图", ut1SVG)
	if strings.Contains(utcSVG, "世界时") {
		t.Fatal("UTC 图不应带 UT1 图注")
	}
	cst := time.FixedZone("CST", 8*3600)
	if _, ok := LunarEclipseMapSVG(date, LunarEclipseMapSVGOptions{TimeScale: astro.TimeScaleUT1, Location: cst}); ok {
		t.Fatal("UT1 配非 UTC 时区应返回 false")
	}
	if _, ok := LunarEclipseMapSVG(date, LunarEclipseMapSVGOptions{TimeScale: astro.TimeScaleUT1, Location: time.UTC}); !ok {
		t.Fatal("UT1 配 UTC 时区应成功")
	}
}

// 三个月食入口共用同一个展示副本：不存在的阶段（零值）不得被换算成真实时刻，
// 且展示副本只影响文字，几何仍走原 UTC 数据。
func TestLunarEclipseUT1LabelHelperPreservesZero(t *testing.T) {
	real := time.Date(2025, 9, 7, 18, 11, 0, 0, time.UTC)
	info := LunarEclipseInfo{Maximum: real, PenumbralStart: real}
	display := lunarEclipseDisplayInfo(info, astro.TimeScaleUT1)
	if !display.TotalStart.IsZero() {
		t.Fatalf("零值阶段应原样保留, got %v", display.TotalStart)
	}
	if !display.Maximum.After(real) {
		t.Fatalf("UT1 时刻应领先民用时刻: %v vs %v", display.Maximum, real)
	}
	if utc := lunarEclipseDisplayInfo(info, astro.TimeScaleUTC); !utc.Maximum.Equal(real) {
		t.Fatalf("UTC 口径不应换算: %v vs %v", utc.Maximum, real)
	}
}

func TestLunarEclipseChartTimeScaleGuard(t *testing.T) {
	date := time.Date(2025, time.September, 7, 12, 0, 0, 0, time.UTC)
	cst := time.FixedZone("CST", 8*3600)
	renderers := map[string]func(time.Time, *time.Location) (string, bool){
		"diagram": func(d time.Time, loc *time.Location) (string, bool) {
			return LunarEclipseSVG(d, LunarEclipseSVGOptions{TimeScale: astro.TimeScaleUT1, Location: loc})
		},
		"detailed": func(d time.Time, loc *time.Location) (string, bool) {
			return LunarEclipseDetailedSVG(d, LunarEclipseDetailedSVGOptions{TimeScale: astro.TimeScaleUT1, Location: loc})
		},
		"map": func(d time.Time, loc *time.Location) (string, bool) {
			return LunarEclipseMapSVG(d, LunarEclipseMapSVGOptions{TimeScale: astro.TimeScaleUT1, Location: loc})
		},
	}
	for name, render := range renderers {
		if _, ok := render(date, cst); ok {
			t.Errorf("%s: UT1 配非 UTC 时区应返回 false", name)
		}
		if _, ok := render(date, time.UTC); !ok {
			t.Errorf("%s: UT1 配 UTC 时区应成功", name)
		}
		if _, ok := render(date, nil); !ok {
			t.Errorf("%s: UT1 下 nil Location 应被强制为 UTC 并成功", name)
		}
	}
}

// 日食两族：UT1 选项必须成功、在图注里声明尺度，并且拒绝非 UTC 时区。
func TestSolarEclipseChartTimeScaleOption(t *testing.T) {
	date := time.Date(2024, time.April, 8, 12, 0, 0, 0, time.UTC)
	cst := time.FixedZone("CST", 8*3600)

	mapSVG, ok := SolarEclipseMapSVG(date, SolarEclipseMapSVGOptions{TimeScale: astro.TimeScaleUT1})
	if !ok {
		t.Fatal("日食地图 UT1 渲染应成功")
	}
	assertUT1Declaration(t, "日食地图", mapSVG)
	if _, ok := SolarEclipseMapSVG(date, SolarEclipseMapSVGOptions{TimeScale: astro.TimeScaleUT1, Location: cst}); ok {
		t.Error("日食地图 UT1 配非 UTC 时区应返回 false")
	}

	localSVG, ok := LocalSolarEclipseSVG(date, -96.8, 32.8, 0, LocalSolarEclipseSVGOptions{TimeScale: astro.TimeScaleUT1})
	if !ok {
		t.Fatal("站心日食 UT1 渲染应成功")
	}
	assertUT1Declaration(t, "站心日食图", localSVG)
	if _, ok := LocalSolarEclipseSVG(date, -96.8, 32.8, 0, LocalSolarEclipseSVGOptions{TimeScale: astro.TimeScaleUT1, Location: cst}); ok {
		t.Error("站心日食 UT1 配非 UTC 时区应返回 false")
	}
}

// 月食组合图与详细版的图注也要声明尺度。
func TestLunarEclipseChartsDeclareUT1(t *testing.T) {
	date := time.Date(2025, time.September, 7, 12, 0, 0, 0, time.UTC)
	for name, svg := range map[string]string{
		"diagram": func() string {
			text, _ := LunarEclipseSVG(date, LunarEclipseSVGOptions{TimeScale: astro.TimeScaleUT1})
			return text
		}(),
		"detailed": func() string {
			text, _ := LunarEclipseDetailedSVG(date, LunarEclipseDetailedSVGOptions{TimeScale: astro.TimeScaleUT1})
			return text
		}(),
	} {
		if svg == "" {
			t.Fatalf("%s: UT1 渲染应成功", name)
		}
		assertUT1Declaration(t, name, svg)
	}
}

// UT1 输出只允许改文字：把 <text> 元素整体去掉后，UTC 与 UT1 的同名图必须逐字节相同。
// 几何一旦用了 UT1 读数（把 UT1 当民用时刻），地平线、月下点与地心坐标块都会平移。
func TestLunarEclipseUT1ChartsKeepGeometry(t *testing.T) {
	date := time.Date(2011, 6, 15, 12, 0, 0, 0, time.UTC)
	renderers := map[string]func(time.Time, astro.TimeScale) (string, bool){
		"map": func(at time.Time, scale astro.TimeScale) (string, bool) {
			return LunarEclipseMapSVG(at, LunarEclipseMapSVGOptions{
				Width: 960, Height: 640, Location: time.UTC, TimeScale: scale})
		},
		"detailed": func(at time.Time, scale astro.TimeScale) (string, bool) {
			return LunarEclipseDetailedSVG(at, LunarEclipseDetailedSVGOptions{
				Width: 1000, Height: 1414, Location: time.UTC, TimeScale: scale})
		},
		"diagram": func(at time.Time, scale astro.TimeScale) (string, bool) {
			return LunarEclipseSVG(at, LunarEclipseSVGOptions{
				Width: 920, Height: 720, Location: time.UTC, TimeScale: scale})
		},
	}
	textElements := regexp.MustCompile(`(?s)<text[^>]*>.*?</text>`)
	for name, render := range renderers {
		utc, ok := render(date, astro.TimeScaleUTC)
		if !ok {
			t.Fatalf("%s: UTC render failed", name)
		}
		ut1, ok := render(date, astro.TimeScaleUT1)
		if !ok {
			t.Fatalf("%s: UT1 render failed", name)
		}
		if strings.Contains(ut1, "UT1") == false {
			t.Errorf("%s: UT1 chart does not declare the scale", name)
		}
		utcGeometry := textElements.ReplaceAllString(utc, "")
		ut1Geometry := textElements.ReplaceAllString(ut1, "")
		if utcGeometry != ut1Geometry {
			t.Errorf("%s: switching to UT1 changed chart geometry", name)
		}
	}
}
