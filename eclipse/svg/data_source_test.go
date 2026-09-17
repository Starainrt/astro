package svg

import (
	"regexp"
	"testing"
	"time"
)

var eclipseMapDataSourcePattern = regexp.MustCompile(`data-source="([^"]+)"`)

// 降级标记必须收敛到包注释里的词表：拼错的来源会让消费方静默漏掉降级图层。
func TestEclipseMapDataSourcesUseDocumentedVocabulary(t *testing.T) {
	// 词表必须与包注释逐项一致：常量改名或拼错都会让消费方的匹配静默失效。
	documented := []string{
		"partial-band-union", "sampled-footprint-sweep", "partial-band-contours",
		"rise-set-phase-lines", "magnitude-contours", "greatest-time-isochrones",
		"besselian-critical-envelope", "paired-limit-chords", "sampled-open-sweep",
		"central-path-limits", "penumbral-outlines", "central-shadow-outlines",
		"p1-p4-visibility-regions", "p1-p4-horizon-boundaries",
	}
	if len(documented) != len(eclipseMapDataSources) {
		t.Fatalf("vocabulary has %d entries, want %d", len(eclipseMapDataSources), len(documented))
	}
	known := make(map[string]bool, len(documented))
	for index, source := range documented {
		if eclipseMapDataSources[index] != source {
			t.Fatalf("vocabulary entry %d = %q, want %q", index, eclipseMapDataSources[index], source)
		}
		known[source] = true
	}

	solarOptions := SolarEclipseMapSVGOptions{
		Width: 960, Height: 640, Location: time.UTC, PartialStep: 10 * time.Minute,
		GreatestTimeStep: 30 * time.Minute, PenumbralOutlineStep: time.Hour, CentralShadowStep: 10 * time.Minute,
	}
	charts := map[string]string{}
	if rendered, ok := SolarEclipseMapSVG(time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC), solarOptions); ok {
		charts["2024-04-08 total"] = rendered
	} else {
		t.Fatal("missing 2024-04-08 map")
	}
	if rendered, ok := SolarEclipseMapSVG(time.Date(2014, time.April, 29, 0, 0, 0, 0, time.UTC), solarOptions); ok {
		charts["2014-04-29 non-central"] = rendered
	} else {
		t.Fatal("missing 2014-04-29 map")
	}
	if rendered, ok := LunarEclipseMapSVG(time.Date(2026, time.March, 3, 0, 0, 0, 0, time.UTC),
		LunarEclipseMapSVGOptions{Width: 960, Height: 640, Location: time.UTC}); ok {
		charts["2026-03-03 lunar"] = rendered
	} else {
		t.Fatal("missing 2026-03-03 lunar map")
	}

	seen := map[string]bool{}
	for name, chart := range charts {
		for _, match := range eclipseMapDataSourcePattern.FindAllStringSubmatch(chart, -1) {
			if !known[match[1]] {
				t.Fatalf("%s: undocumented data-source %q", name, match[1])
			}
			seen[match[1]] = true
		}
	}
	// 主要图层必须带标记，否则消费方无法判断这一层是否降级。
	for _, source := range []string{
		eclipseMapSourcePartialBandUnion,
		eclipseMapSourcePartialBandContours,
		eclipseMapSourceRiseSetPhaseLines,
		eclipseMapSourceMagnitudeContours,
		eclipseMapSourceGreatestTimeIsochrones,
		eclipseMapSourceBesselianEnvelope,
		eclipseMapSourceCentralPathLimits,
		eclipseMapSourcePenumbralOutlines,
		eclipseMapSourceCentralShadowOutlines,
		eclipseMapSourceLunarVisibilityRegions,
		eclipseMapSourceLunarHorizonBoundaries,
	} {
		if !seen[source] {
			t.Fatalf("main layer source %q carries no data-source marker", source)
		}
	}
}
