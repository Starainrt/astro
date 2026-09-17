package svg

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

// 详细版式组合图必须把摘要、穿影示意图与世界可见性底图放在同一页，并带日月地心坐标块与三栏面板。
func TestLunarEclipseDetailedSVGCombinesBothCharts(t *testing.T) {
	date := time.Date(2025, time.March, 14, 0, 0, 0, 0, time.UTC)
	for _, language := range []string{"zh", "en"} {
		rendered, ok := LunarEclipseDetailedSVG(date, LunarEclipseDetailedSVGOptions{
			Language: language, Location: time.UTC,
		})
		if !ok {
			t.Fatalf("%s: missing chart", language)
		}
		shared := []string{
			`class="lunar-geocentric-block"`,   // 日月地心坐标块
			`class="lunar-detailed-panel"`,     // 历时/接触/……面板
			`class="lunar-scale-bar"`,          // 弧分比例尺
			`class="lunar-visibility-regions"`, // 可见性区域
			`class="entire-eclipse-region"`,    // 全程可见区
			`<path class="land"`,               // 陆地底图
		}
		// 底图图例按语言取词。
		if language == "en" {
			shared = append(shared, "Entire eclipse", "Moonrise during eclipse", "Not visible")
		} else {
			shared = append(shared, "全程可见", "带食月出", "带食月落", "不可见")
		}
		for _, want := range shared {
			if !strings.Contains(rendered, want) {
				t.Fatalf("%s: combined chart missing %q", language, want)
			}
		}
		if err := validateEclipseMapXML(rendered); err != nil {
			t.Fatalf("%s: chart is not valid XML: %v", language, err)
		}
	}
}

// 地影几何要与 NASA 月食图逐项吻合：2009 之外的 2025-03-14 全食，Gamma 0.3481、P./U. Radius 1.1899/0.6537。
func TestLunarEclipseShadowGeometryMatchesNASABulletin(t *testing.T) {
	date := time.Date(2025, time.March, 14, 0, 0, 0, 0, time.UTC)
	result := basic.LunarEclipse(basic.TD2UT(basic.Date2JDE(date), true))
	geometry := basic.LunarEclipseShadowGeometryAt(result.Maximum)
	t.Logf("gamma=%.4f P.Radius=%.4f U.Radius=%.4f umbral=%.4f penumbral=%.4f",
		geometry.Gamma, geometry.PenumbralRadiusDegrees, geometry.UmbralRadiusDegrees,
		result.Magnitude, result.PenumbralMagnitude)
	if math.Abs(geometry.Gamma-0.3481) > 0.001 {
		t.Fatalf("gamma = %.4f, want 0.3481", geometry.Gamma)
	}
	if math.Abs(geometry.PenumbralRadiusDegrees-1.1899) > 0.001 {
		t.Fatalf("penumbral radius = %.4f, want 1.1899", geometry.PenumbralRadiusDegrees)
	}
	if math.Abs(geometry.UmbralRadiusDegrees-0.6537) > 0.001 {
		t.Fatalf("umbral radius = %.4f, want 0.6537", geometry.UmbralRadiusDegrees)
	}
	if math.Abs(result.Magnitude-1.1784) > 0.001 {
		t.Fatalf("umbral magnitude = %.4f, want 1.1784", result.Magnitude)
	}
	if math.Abs(result.PenumbralMagnitude-2.2595) > 0.001 {
		t.Fatalf("penumbral magnitude = %.4f, want 2.2595", result.PenumbralMagnitude)
	}
}
