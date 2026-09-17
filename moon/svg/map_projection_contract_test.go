package svg

import (
	"testing"

	eclipsesvg "github.com/starainrt/astro/eclipse/svg"
	"github.com/starainrt/astro/internal/svgmap"
)

// 两包为同一投影概念各留一套名字：取值字符串与内部 svgmap 映射必须逐对一致。
func TestMapProjectionMatchesEclipseSVGContract(t *testing.T) {
	pairs := []struct {
		name    string
		value   MapProjection
		eclipse eclipsesvg.EclipseMapProjection
		text    string
	}{
		{name: "auto", value: MapProjectionAuto, eclipse: eclipsesvg.EclipseMapProjectionAuto, text: ""},
		{name: "equirectangular", value: MapProjectionEquirectangular, eclipse: eclipsesvg.EclipseMapProjectionEquirectangular, text: "equirectangular"},
		{name: "north-polar", value: MapProjectionNorthPolar, eclipse: eclipsesvg.EclipseMapProjectionNorthPolar, text: "north-polar"},
		{name: "south-polar", value: MapProjectionSouthPolar, eclipse: eclipsesvg.EclipseMapProjectionSouthPolar, text: "south-polar"},
		{name: "orthographic", value: MapProjectionOrthographic, eclipse: eclipsesvg.EclipseMapProjectionOrthographic, text: "orthographic"},
	}
	seen := map[svgmap.Projection]string{}
	for _, pair := range pairs {
		if got := string(pair.value); got != pair.text {
			t.Fatalf("moon/svg %s = %q, want %q", pair.name, got, pair.text)
		}
		if got := string(pair.eclipse); got != pair.text {
			t.Fatalf("eclipse/svg %s = %q, want %q", pair.name, got, pair.text)
		}
		moonProjection := internalMapProjection(pair.value)
		if moonProjection != svgmap.Projection(pair.text) {
			t.Fatalf("moon/svg %s maps to %q, want %q", pair.name, moonProjection, svgmap.Projection(pair.text))
		}
		if eclipseProjection := svgmap.Projection(pair.eclipse); eclipseProjection != moonProjection {
			t.Fatalf("%s maps to %q in moon/svg and %q in eclipse/svg", pair.name, moonProjection, eclipseProjection)
		}
		if previous, ok := seen[moonProjection]; ok {
			t.Fatalf("%s and %s share the internal projection %q", previous, pair.name, moonProjection)
		}
		seen[moonProjection] = pair.name
	}
}
