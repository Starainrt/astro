package svg

import (
	"testing"

	"github.com/starainrt/astro/internal/svgmap"
)

// 投影常量的取值字符串是与 moon/svg.MapProjection 共享的稳定契约。
func TestEclipseMapProjectionValuesContract(t *testing.T) {
	equivalents := []struct {
		name  string
		value EclipseMapProjection
		text  string
	}{
		{name: "auto", value: EclipseMapProjectionAuto, text: ""},
		{name: "equirectangular", value: EclipseMapProjectionEquirectangular, text: "equirectangular"},
		{name: "north-polar", value: EclipseMapProjectionNorthPolar, text: "north-polar"},
		{name: "south-polar", value: EclipseMapProjectionSouthPolar, text: "south-polar"},
		{name: "orthographic", value: EclipseMapProjectionOrthographic, text: "orthographic"},
	}
	for _, testCase := range equivalents {
		if got := string(testCase.value); got != testCase.text {
			t.Fatalf("%s = %q, want %q", testCase.name, got, testCase.text)
		}
		if got := internalEclipseMapProjection(testCase.value); got != svgmap.Projection(testCase.text) {
			t.Fatalf("%s maps to %q, want %q", testCase.name, got, svgmap.Projection(testCase.text))
		}
		if !validEclipseMapProjection(testCase.value) {
			t.Fatalf("%s must pass validation", testCase.name)
		}
	}
	// 边界：只有上表登记的字面量可以过校验，大小写与分隔符都不放宽。
	for _, invalid := range []EclipseMapProjection{"auto", "Orthographic", "north_polar", "polar", "equirectangular "} {
		if validEclipseMapProjection(invalid) {
			t.Fatalf("validator accepted unsupported projection %q", invalid)
		}
	}
}
