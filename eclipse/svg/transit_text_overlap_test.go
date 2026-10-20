package svg

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"
)

var transitLegendSwatchPattern = regexp.MustCompile(`<rect [^>]*width="18" height="9"[^>]*fill-opacity="([0-9.]+)"`)

type transitOverlapCase struct {
	name   string
	render func(width, height int) (string, bool)
}

func transitOverlapCases() []transitOverlapCase {
	partial := time.Date(1937, time.May, 11, 0, 0, 0, 0, time.UTC)
	ordinary := time.Date(2012, time.June, 6, 0, 0, 0, 0, time.UTC)
	return []transitOverlapCase{
		{"local-partial-zh", func(width, height int) (string, bool) {
			return TransitLocalSVG(TransitMercury, partial, 0, -60, 0, TransitLocalSVGOptions{Width: width, Height: height})
		}},
		{"local-partial-en", func(width, height int) (string, bool) {
			return TransitLocalSVG(TransitMercury, partial, 0, -60, 0, TransitLocalSVGOptions{Width: width, Height: height, Language: "en"})
		}},
		{"local-full-zh", func(width, height int) (string, bool) {
			return TransitLocalSVG(TransitVenus, ordinary, 116.4, 39.9, 0, TransitLocalSVGOptions{Width: width, Height: height})
		}},
		{"local-full-en", func(width, height int) (string, bool) {
			return TransitLocalSVG(TransitVenus, ordinary, 116.4, 39.9, 0, TransitLocalSVGOptions{Width: width, Height: height, Language: "en"})
		}},
		{"map-partial", func(width, height int) (string, bool) {
			return TransitMapSVG(TransitMercury, partial, TransitMapSVGOptions{Width: width, Height: height, Isochrones: TransitIsochroneAll})
		}},
		{"map-partial-orthographic", func(width, height int) (string, bool) {
			return TransitMapSVG(TransitMercury, partial, TransitMapSVGOptions{
				Width: width, Height: height, Isochrones: TransitIsochroneAll, Projection: EclipseMapProjectionOrthographic,
			})
		}},
		{"map-ordinary", func(width, height int) (string, bool) {
			return TransitMapSVG(TransitVenus, ordinary, TransitMapSVGOptions{Width: width, Height: height, Isochrones: TransitIsochroneAll})
		}},
	}
}

// 凌日图与日月食图同一口径：任意两个尺寸下返回的图都必须 0 文字压叠。
func TestTransitChartTextOverlapMatrix(t *testing.T) {
	for _, size := range [][2]int{{640, 420}, {640, 480}, {800, 600}, {960, 640}, {960, 680}, {1200, 800}} {
		for _, test := range transitOverlapCases() {
			name := fmt.Sprintf("%s/%dx%d", test.name, size[0], size[1])
			document, ok := test.render(size[0], size[1])
			if !ok {
				continue
			}
			assertNoTextOverlap(t, name, document)
		}
	}
	// 文档下限尺寸必须继续出图，拒绝只允许出现在真放不下的画布上。
	if _, ok := TransitLocalSVG(TransitMercury, time.Date(1937, time.May, 11, 0, 0, 0, 0, time.UTC), 0, -60, 0,
		TransitLocalSVGOptions{Width: 640, Height: 480}); !ok {
		t.Error("documented 640x480 local transit canvas was rejected")
	}
}

// 地心无接触的可见区填充必须落在 data-source 词表内，消费方才能判断该层来源。
func TestTransitPartialRegionsUseDocumentedDataSource(t *testing.T) {
	chart, ok := TransitMapSVG(TransitMercury, time.Date(1937, time.May, 11, 0, 0, 0, 0, time.UTC), TransitMapSVGOptions{})
	if !ok {
		t.Fatal("partial transit render failed")
	}
	if !strings.Contains(chart, `data-source="`+eclipseMapSourceTransitPartialVisibility+`"`) {
		t.Fatalf("partial regions carry no documented data-source marker")
	}
	for _, match := range eclipseMapDataSourcePattern.FindAllStringSubmatch(chart, -1) {
		known := false
		for _, source := range eclipseMapDataSources {
			known = known || source == match[1]
		}
		if !known {
			t.Fatalf("undocumented data-source %q", match[1])
		}
	}
}

// 图例色块画在白底上取月食图同值；部分凌日单块可见区比普通分区更浓。
func TestTransitMapFillOpacities(t *testing.T) {
	for _, test := range []struct {
		planet  TransitPlanet
		date    time.Time
		pattern *regexp.Regexp
		region  float64
	}{
		{TransitVenus, time.Date(2012, time.June, 6, 0, 0, 0, 0, time.UTC),
			regexp.MustCompile(`class="transit-[a-z-]+-region"[^>]*fill-opacity="([0-9.]+)"`), transitMapRegionFillOpacity},
		{TransitMercury, time.Date(1937, time.May, 11, 0, 0, 0, 0, time.UTC),
			regexp.MustCompile(`class="transit-partial-region"[^>]*fill-opacity="([0-9.]+)"`), transitMapPartialFillOpacity},
	} {
		chart, ok := TransitMapSVG(test.planet, test.date, TransitMapSVGOptions{})
		if !ok {
			t.Fatalf("%v: render failed", test.date)
		}
		regions := test.pattern.FindAllStringSubmatch(chart, -1)
		if len(regions) == 0 {
			t.Fatalf("%v: no region fill found", test.date)
		}
		for _, match := range regions {
			if match[1] != fmt.Sprintf("%.2f", test.region) {
				t.Fatalf("%v: region fill-opacity %s, want %.2f", test.date, match[1], test.region)
			}
		}
		legends := transitLegendSwatchPattern.FindAllStringSubmatch(chart, -1)
		if len(legends) == 0 {
			t.Fatalf("%v: no legend swatch found", test.date)
		}
		for _, match := range legends {
			if match[1] != fmt.Sprintf("%.2f", transitMapLegendFillOpacity) {
				t.Fatalf("%v: legend fill-opacity %s, want %.2f", test.date, match[1], transitMapLegendFillOpacity)
			}
		}
	}
	if transitMapPartialFillOpacity <= transitMapRegionFillOpacity {
		t.Fatal("partial-transit footprint must be denser than ordinary regions")
	}
}
