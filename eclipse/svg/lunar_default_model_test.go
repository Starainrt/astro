package svg

import (
	"strings"
	"testing"
	"time"

	eclipsecore "github.com/starainrt/astro/eclipse"
)

// 三个无后缀月食 SVG 入口共用与 LunarEclipseOnDate 一致的默认模型：Danjon 为主，
// 极浅半影回退 Chauvenet。2016-08-18 是 Danjon 判无食、Chauvenet 判极浅半影的个例。
func TestLunarEclipseSVGDefaultModelMatchesCore(t *testing.T) {
	date := time.Date(2016, time.August, 18, 0, 0, 0, 0, time.UTC)
	info, coreOK := eclipsecore.LunarEclipseOnDate(date)
	if !coreOK {
		t.Fatal("2016-08-18 is no longer a default-model eclipse date")
	}
	if info.ShadowModel != eclipsecore.LunarEclipseShadowModelChauvenet {
		t.Fatalf("default model = %v, want the Chauvenet fallback", info.ShadowModel)
	}
	if _, danjonOK := eclipsecore.LunarEclipseOnDateDanjon(date); danjonOK {
		t.Fatal("2016-08-18 no longer separates the Danjon and default models")
	}

	mapRendered, mapOK := LunarEclipseMapSVG(date, LunarEclipseMapSVGOptions{Width: 900, Height: 620, Location: time.UTC})
	plain, plainOK := LunarEclipseSVG(date, LunarEclipseSVGOptions{Width: 960, Height: 620, Location: time.UTC})
	detailed, detailedOK := LunarEclipseDetailedSVG(date, LunarEclipseDetailedSVGOptions{Width: 1000, Height: 1414, Location: time.UTC})
	if !mapOK || !plainOK || !detailedOK {
		t.Fatalf("default-model entries disagree: map=%v svg=%v detailed=%v (core=%v)", mapOK, plainOK, detailedOK, coreOK)
	}
	if mapRendered == "" || plain == "" || detailed == "" {
		t.Fatal("a default-model entry reported success but returned no document")
	}
	// 出图必须真的走 Chauvenet：详细版页脚写出实际模型，且与强制 Chauvenet 入口逐字节相同。
	if !strings.Contains(detailed, "Chauvenet") {
		t.Fatal("detailed chart does not name the Chauvenet fallback model")
	}
	chauvenet, chauvenetOK := LunarEclipseDetailedSVGChauvenet(date, LunarEclipseDetailedSVGOptions{Width: 1000, Height: 1414, Location: time.UTC})
	if !chauvenetOK || chauvenet != detailed {
		t.Fatal("default detailed chart differs from the forced Chauvenet chart")
	}
	forcedPlain, forcedPlainOK := LunarEclipseSVGChauvenet(date, LunarEclipseSVGOptions{Width: 960, Height: 620, Location: time.UTC})
	if !forcedPlainOK || forcedPlain != plain {
		t.Fatal("default lunar chart differs from the forced Chauvenet chart")
	}
	forcedMap, forcedMapOK := LunarEclipseMapSVGChauvenet(date, LunarEclipseMapSVGOptions{Width: 900, Height: 620, Location: time.UTC})
	if !forcedMapOK || forcedMap != mapRendered {
		t.Fatal("default lunar map differs from the forced Chauvenet map")
	}

	// 强制 Danjon 的两个入口必须继续判无食，后缀入口不得被默认口径带偏。
	if _, ok := LunarEclipseSVGDanjon(date, LunarEclipseSVGOptions{}); ok {
		t.Fatal("the Danjon-only chart accepted a date that Danjon rejects")
	}
	if _, ok := LunarEclipseDetailedSVGDanjon(date, LunarEclipseDetailedSVGOptions{}); ok {
		t.Fatal("the Danjon-only detailed chart accepted a date that Danjon rejects")
	}
	if _, ok := LunarEclipseMapSVGDanjon(date, LunarEclipseMapSVGOptions{}); ok {
		t.Fatal("the Danjon-only map accepted a date that Danjon rejects")
	}
}

// 默认模型在普通月食日期上仍走 Danjon，不能把全部入口改成回退口径。
func TestLunarEclipseSVGDefaultModelKeepsDanjon(t *testing.T) {
	date := time.Date(2025, time.March, 14, 0, 0, 0, 0, time.UTC)
	detailed, ok := LunarEclipseDetailedSVG(date, LunarEclipseDetailedSVGOptions{Width: 1000, Height: 1414, Location: time.UTC})
	if !ok {
		t.Fatal("missing 2025-03-14 detailed chart")
	}
	if !strings.Contains(detailed, "Danjon") || strings.Contains(detailed, "Chauvenet") {
		t.Fatal("2025-03-14 is a Danjon eclipse but the chart claims another model")
	}
	if _, ok := LunarEclipseSVG(date, LunarEclipseSVGOptions{}); !ok {
		t.Fatal("missing 2025-03-14 chart")
	}
}
