package svg

// 各日月食出图入口都接受月缘剖面，并在图注写明所用口径。

import (
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/moon/limb"
)

func TestEclipseSVGsDeclareLunarLimbConvention(t *testing.T) {
	solarDate := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	lunarDate := time.Date(2025, time.September, 7, 0, 0, 0, 0, time.UTC)
	profile := limb.Default()
	render := map[string]func(applied bool) (string, bool){
		"local solar": func(applied bool) (string, bool) {
			options := LocalSolarEclipseSVGOptions{Location: time.UTC}
			if applied {
				options.LunarLimb = profile
			}
			return LocalSolarEclipseSVG(solarDate, -96.8, 32.78, 0, options)
		},
		"solar map": func(applied bool) (string, bool) {
			options := SolarEclipseMapSVGOptions{Location: time.UTC}
			if applied {
				options.LunarLimb = profile
			}
			return SolarEclipseMapSVG(solarDate, options)
		},
		"lunar": func(applied bool) (string, bool) {
			options := LunarEclipseSVGOptions{Location: time.UTC}
			if applied {
				options.LunarLimb = profile
			}
			return LunarEclipseSVG(lunarDate, options)
		},
		"lunar map": func(applied bool) (string, bool) {
			options := LunarEclipseMapSVGOptions{Location: time.UTC}
			if applied {
				options.LunarLimb = profile
			}
			return LunarEclipseMapSVG(lunarDate, options)
		},
		"lunar detailed": func(applied bool) (string, bool) {
			options := LunarEclipseDetailedSVGOptions{Location: time.UTC}
			if applied {
				options.LunarLimb = profile
			}
			return LunarEclipseDetailedSVG(lunarDate, options)
		},
	}
	for name, draw := range render {
		circular, ok := draw(false)
		if !ok {
			t.Fatalf("%s: circular chart missing", name)
		}
		corrected, ok := draw(true)
		if !ok {
			t.Fatalf("%s: limb chart missing", name)
		}
		if !strings.Contains(circular, "圆月面") {
			t.Fatalf("%s: circular chart does not declare the circular limb", name)
		}
		if !strings.Contains(corrected, "月缘剖面") {
			t.Fatalf("%s: limb chart does not declare the limb profile", name)
		}
	}
}
