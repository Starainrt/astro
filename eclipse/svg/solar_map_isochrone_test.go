package svg

import (
	"strings"
	"testing"
	"time"

	eclipsecore "github.com/starainrt/astro/eclipse"
)

func solarEclipseIsochroneMapOptions() SolarEclipseMapSVGOptions {
	return SolarEclipseMapSVGOptions{
		Width: 900, Height: 620, Location: time.UTC, PartialStep: 10 * time.Minute,
		GreatestTimeStep: solarEclipseMapDefaultGreatestTimeStep,
	}
}

// 等时线是可选图层：不请求时不得出现在输出里，请求时路径与标注都必须绘制。
func TestSolarEclipseMapGreatestTimeIsochrones(t *testing.T) {
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC)
	rendered, ok := SolarEclipseMapSVG(date, solarEclipseIsochroneMapOptions())
	if !ok {
		t.Fatal("missing map")
	}
	if !strings.Contains(rendered, `class="solar-greatest-time-isoline"`) {
		t.Fatal("isochrone paths missing from a map that requested them")
	}
	if !strings.Contains(rendered, `class="solar-greatest-time-label"`) {
		t.Fatal("isochrone labels missing from a map that requested them")
	}

	// 等时线是 opt-in：零值和负值都不画，与核心层和 moon/svg 一致；30 分钟只是 NASA 全球图的推荐间隔。
	if got := normalizeSolarEclipseMapSVGOptions(date, SolarEclipseMapSVGOptions{}).GreatestTimeStep; got != 0 {
		t.Fatalf("default isochrone step = %s, want them disabled", got)
	}
	if solarEclipseMapDefaultGreatestTimeStep != 30*time.Minute {
		t.Fatalf("recommended NASA isochrone step = %s, want 30m", solarEclipseMapDefaultGreatestTimeStep)
	}
	for _, step := range []time.Duration{0, -1} {
		disabled, ok := SolarEclipseMapSVG(date, SolarEclipseMapSVGOptions{
			Width: 900, Height: 620, Location: time.UTC, PartialStep: 10 * time.Minute,
			GreatestTimeStep: step,
		})
		if !ok {
			t.Fatal("missing map")
		}
		if strings.Contains(disabled, "solar-greatest-time-isoline") ||
			strings.Contains(disabled, "solar-greatest-time-label") {
			t.Fatalf("step %s must not draw isochrones", step)
		}
	}
	// 正值小于一分钟时夹取到一分钟，而不是关闭。
	clamped := normalizeSolarEclipseMapSVGOptions(date, SolarEclipseMapSVGOptions{GreatestTimeStep: time.Second})
	if clamped.GreatestTimeStep != time.Minute {
		t.Fatalf("sub-minute isochrone step = %s, want it clamped to one minute", clamped.GreatestTimeStep)
	}
}

// 电平必须落在偏食可见窗口内并按步长对齐，避免 59.999 秒那类截断。
func TestSolarEclipseMapGreatestTimeLevelsAlignToStep(t *testing.T) {
	date := time.Date(2009, time.July, 22, 0, 0, 0, 0, time.UTC)
	options := normalizeSolarEclipseMapSVGOptions(date, solarEclipseIsochroneMapOptions())
	info, ok := eclipsecore.SolarEclipseOnDateNASABulletinSplitK(date)
	if !ok {
		t.Fatal("missing eclipse")
	}
	levels := solarEclipseGreatestTimeLevels(info, options)
	if len(levels) < 5 {
		t.Fatalf("expected a full isochrone fan, got %d levels", len(levels))
	}
	for _, level := range levels {
		if level.Second() != 0 || level.Nanosecond() != 0 || level.Minute()%30 != 0 {
			t.Fatalf("level %s is not aligned to the requested step", level.Format("15:04:05.000"))
		}
		if level.Before(info.PartialBeginOnEarth) || level.After(info.PartialEndOnEarth) {
			t.Fatalf("level %s is outside the partial-eclipse window", level.Format("15:04:05"))
		}
	}
}
