package svg

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/moon"
)

// 掩甚时刻等时线由路径选项请求，渲染器只在路径携带时绘制。
func TestPlanetOccultationSVGDrawsGreatestTimeIsochrones(t *testing.T) {
	start := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, time.February, 2, 0, 0, 0, 0, time.UTC)
	render := func(step time.Duration) string {
		t.Helper()
		paths, err := moon.FindPlanetOccultationPaths(start, end, moon.OccultationSaturn, moon.OccultationPathOptions{
			Step: 2 * time.Minute, DisableFootprints: true, GreatestTimeStep: step,
		})
		if err != nil || len(paths) == 0 {
			t.Fatalf("no occultation path: %v", err)
		}
		if step > 0 && len(paths[0].GreatestTimeContours) == 0 {
			t.Fatal("path carries no isochrones")
		}
		rendered, renderErr := PlanetOccultationPathSVG(paths[0], StarOccultationSVGOptions{
			Width: 1200, Height: 800, Location: time.UTC,
		})
		if renderErr != nil {
			t.Fatalf("render: %v", renderErr)
		}
		return rendered
	}

	withIsochrones := render(30 * time.Minute)
	if !strings.Contains(withIsochrones, `class="greatest-time-isoline"`) {
		t.Fatal("isochrone paths missing")
	}
	if !strings.Contains(withIsochrones, `class="greatest-time-isoline-label"`) {
		t.Fatal("isochrone labels missing")
	}
	if strings.Contains(render(0), "greatest-time-isoline") {
		t.Fatal("a zero step must not draw isochrones")
	}
}

// 等时线也要走路径校验：坏支路必须在渲染前被拦住，而不是画出 NaN 折线。
func TestPlanetOccultationSVGRejectsBrokenGreatestTimeIsochrones(t *testing.T) {
	paths, err := moon.FindPlanetOccultationPaths(
		time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, time.February, 2, 0, 0, 0, 0, time.UTC), moon.OccultationSaturn,
		moon.OccultationPathOptions{Step: 2 * time.Minute, DisableFootprints: true, GreatestTimeStep: 30 * time.Minute},
	)
	if err != nil || len(paths) == 0 || len(paths[0].GreatestTimeContours) == 0 {
		t.Fatalf("no isochrone path: %v", err)
	}
	options := StarOccultationSVGOptions{Width: 1200, Height: 800, Location: time.UTC}
	render := func(mutate func(path *moon.PlanetOccultationPath)) error {
		path := paths[0]
		path.GreatestTimeContours = append([]moon.OccultationGreatestTimeContour(nil), path.GreatestTimeContours...)
		mutate(&path)
		_, renderErr := PlanetOccultationPathSVG(path, options)
		return renderErr
	}
	if err := render(func(path *moon.PlanetOccultationPath) {}); err != nil {
		t.Fatalf("valid isochrones rejected: %v", err)
	}
	if err := render(func(path *moon.PlanetOccultationPath) {
		path.GreatestTimeContours[0].JDE = 0
	}); err == nil {
		t.Fatal("zero JDE accepted")
	}
	if err := render(func(path *moon.PlanetOccultationPath) {
		path.GreatestTimeContours[0].Segments[0][0].Longitude = math.NaN()
	}); err == nil {
		t.Fatal("NaN longitude accepted")
	}
	if err := render(func(path *moon.PlanetOccultationPath) {
		path.GreatestTimeContours[0].Segments[0] = path.GreatestTimeContours[0].Segments[0][:1]
	}); err == nil {
		t.Fatal("single-point branch accepted")
	}
	if err := render(func(path *moon.PlanetOccultationPath) {
		path.GreatestTimeContours[0].Segments[0][0].Time = path.GreatestTimeContours[0].Segments[0][0].Time.Add(2 * time.Second)
	}); err == nil {
		t.Fatal("branch point with a mismatched instant accepted")
	}
}
