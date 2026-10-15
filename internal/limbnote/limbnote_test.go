package limbnote

import "testing"

func TestSolarMapListsOnlyAppliedLayers(t *testing.T) {
	cases := []struct {
		layers   SolarMapLayers
		language string
		want     string
	}{
		{SolarMapLayers{}, "zh", "月面按圆月面计算。"},
		{SolarMapLayers{}, english, "Circular lunar limb."},
		{SolarMapLayers{PartialBoundary: true}, "zh", "按月缘剖面修正可见域边界（食分线与等时线仍按圆月面）。"},
		{SolarMapLayers{GlobalContacts: true, CentralLimits: true}, "zh", "按月缘剖面修正全球始末、中心带界线（食分线与等时线仍按圆月面）。"},
		{SolarMapLayers{LocalPanel: true, PartialBoundary: true}, english,
			"Lunar limb profile applied to local values at greatest eclipse, visibility boundary (isolines circular)."},
	}
	for _, c := range cases {
		if got := SolarMap(c.layers, c.language); got != c.want {
			t.Errorf("SolarMap(%+v, %s) = %q, want %q", c.layers, c.language, got, c.want)
		}
	}
}
