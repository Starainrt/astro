package svg

import (
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/moon"
)

// 月掩五族的 UT1 选项：渲染成功、详细图声明尺度、非 UTC 时区报错，且绝不改写调用方传进来的 path。
func TestOccultationSVGTimeScaleUT1(t *testing.T) {
	cst := time.FixedZone("CST", 8*3600)
	path := occultationTestStarPath(t)
	before := path.Start.Time

	starSVG, err := StarOccultationDetailedSVG(path, hr4799StarCoordinate(), OccultationDetailedSVGOptions{TimeScale: astro.TimeScaleUT1})
	if err != nil {
		t.Fatalf("恒星月掩 UT1 详细图渲染失败: %v", err)
	}
	assertOccultationUT1Declaration(t, "恒星月掩详细图", starSVG)
	if !path.Start.Time.Equal(before) {
		t.Error("渲染器不得改写调用方传入的 path")
	}

	starPathSVG, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{TimeScale: astro.TimeScaleUT1})
	if err != nil {
		t.Fatalf("恒星月掩 UT1 路径图渲染失败: %v", err)
	}
	assertOccultationUT1Declaration(t, "恒星月掩全球路径图", starPathSVG)
	if _, err := StarOccultationPathSVG(path, StarOccultationSVGOptions{TimeScale: astro.TimeScaleUT1, Location: cst}); err == nil {
		t.Error("UT1 配非 UTC 时区应报错")
	}
	if _, err := StarOccultationDetailedSVG(path, hr4799StarCoordinate(), OccultationDetailedSVGOptions{TimeScale: astro.TimeScaleUT1, Location: cst}); err == nil {
		t.Error("UT1 配非 UTC 时区应报错（详细图）")
	}

	planetPath := samplePlanetOccultationPath()
	if _, err := PlanetOccultationDetailedSVG(planetPath, OccultationDetailedSVGOptions{TimeScale: astro.TimeScaleUT1}); err != nil {
		t.Fatalf("行星月掩 UT1 详细图渲染失败: %v", err)
	}
	if _, err := PlanetOccultationDetailedSVG(planetPath, OccultationDetailedSVGOptions{TimeScale: astro.TimeScaleUT1, Location: cst}); err == nil {
		t.Error("行星月掩 UT1 配非 UTC 时区应报错")
	}
	// 行星全球路径图与恒星侧同契约：UT1 配非 UTC 必须报错，配 UTC 必须声明时标。
	if _, err := PlanetOccultationPathSVG(planetPath, PlanetOccultationSVGOptions{TimeScale: astro.TimeScaleUT1, Location: cst}); err == nil {
		t.Error("行星月掩全球路径图 UT1 配非 UTC 时区应报错")
	}
	planetPathSVG, err := PlanetOccultationPathSVG(planetPath, PlanetOccultationSVGOptions{TimeScale: astro.TimeScaleUT1})
	if err != nil {
		t.Fatalf("行星月掩 UT1 全球路径图渲染失败: %v", err)
	}
	assertOccultationUT1Declaration(t, "行星月掩全球路径图", planetPathSVG)

	info := localHR4799Occultation(t)
	if _, err := LocalStarOccultationSVG(info, hr4799StarCoordinate(), LocalStarOccultationSVGOptions{TimeScale: astro.TimeScaleUT1, Location: time.UTC}); err != nil {
		t.Fatalf("地方月掩 UT1 渲染失败: %v", err)
	}
	if _, err := LocalStarOccultationSVG(info, hr4799StarCoordinate(), LocalStarOccultationSVGOptions{TimeScale: astro.TimeScaleUT1, Location: cst}); err == nil {
		t.Error("地方月掩 UT1 配非 UTC 时区应报错")
	}
}

// assertOccultationUT1Declaration 钉住 UT1 声明是可解析的 <text> 元素并写出 DUT1：曾经的裸文本写入不会渲染。
func assertOccultationUT1Declaration(t *testing.T, name, diagram string) {
	t.Helper()
	for _, text := range occultationSVGTexts(t, diagram) {
		if strings.Contains(text.value, "DUT1 = UT1−UTC = ") {
			return
		}
	}
	t.Errorf("%s 应在图内的 <text> 元素里声明 UT1 与 DUT1 差值", name)
}

// 换算器必须保留"该阶段不存在"的零值语义。
func TestOccultationInfoConvertersPreserveZero(t *testing.T) {
	info := moon.StarOccultationInfo{Greatest: time.Date(2025, 1, 5, 17, 0, 0, 0, time.UTC)}
	got := moon.StarOccultationInfoInUT1(info)
	if !got.Immersion.IsZero() || !got.Emersion.IsZero() {
		t.Error("不存在的阶段应保持零值")
	}
	if !got.Greatest.After(info.Greatest) {
		t.Error("UT1 时刻应领先民用时刻")
	}
}

// 详细版历表面板（ΔT、月距、天平动、月亮地心坐标）必须与输出时标无关：注入一个偏移 1 小时的
// TT−UTC 覆盖，让"把 UT1 读数当民用时刻"的写法必然算出不同历表值；修复后两次渲染的
// 非时刻文字必须逐个相同。
func TestOccultationDetailedPanelIgnoresTimeScale(t *testing.T) {
	astro.SetTTMinusUTC(func(jd float64) float64 {
		return basic.TTMinusUTCSecondsDefault(jd) + 3600
	})
	defer astro.SetTTMinusUTC(nil)

	path := occultationTestStarPath(t)
	utc, err := StarOccultationDetailedSVG(path, hr4799StarCoordinate(), OccultationDetailedSVGOptions{})
	if err != nil {
		t.Fatalf("UTC 详细图渲染失败: %v", err)
	}
	ut1, err := StarOccultationDetailedSVG(path, hr4799StarCoordinate(), OccultationDetailedSVGOptions{TimeScale: astro.TimeScaleUT1})
	if err != nil {
		t.Fatalf("UT1 详细图渲染失败: %v", err)
	}
	utcValues := occultationPanelValues(t, utc)
	ut1Values := occultationPanelValues(t, ut1)
	if len(utcValues) == 0 {
		t.Fatal("未解析到面板文字")
	}
	if len(utcValues) != len(ut1Values) {
		t.Fatalf("两图文字数不同: %d vs %d", len(utcValues), len(ut1Values))
	}
	for index := range utcValues {
		if utcValues[index] != ut1Values[index] {
			t.Errorf("第 %d 条文字随时标改变: %q vs %q", index, utcValues[index], ut1Values[index])
		}
	}
}

// occultationPanelValues 取出不随时标变化的文字：跳过时刻、时区声明与包含它们的行。
func occultationPanelValues(t *testing.T, diagram string) []string {
	t.Helper()
	values := make([]string, 0, 32)
	for _, text := range occultationSVGTexts(t, diagram) {
		value := strings.TrimSpace(text.value)
		if value == "" || strings.Contains(value, "UT1") || strings.Contains(value, "DUT1") ||
			strings.Contains(value, "UTC") || strings.Contains(value, "时区") || strings.Contains(value, "图中时刻") {
			continue
		}
		if occultationTimeLike(value) {
			continue
		}
		values = append(values, value)
	}
	return values
}

func occultationTimeLike(value string) bool {
	for _, field := range strings.Fields(value) {
		trimmed := strings.Trim(field, "(),")
		if len(trimmed) >= 5 && trimmed[2] == ':' && trimmed[0] >= '0' && trimmed[0] <= '9' {
			return true
		}
		if len(trimmed) >= 8 && trimmed[4] == '-' && trimmed[7] == '-' {
			return true
		}
	}
	return false
}
