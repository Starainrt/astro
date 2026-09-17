package occultationgeo

import (
	"testing"
	"time"

	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/moon"
)

func TestMars20250729TotalBandAvoidsDetachedSliver(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationMars,
		moon.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
			DisableFootprints: true, IncludeFootprintTimeline: true,
			FootprintTimelineStep: 5 * time.Minute,
		},
	)
	if err != nil {
		t.Fatalf("FindPlanetOccultationPaths: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("paths=%d", len(paths))
	}
	path := paths[0]
	polygons, authoritative, err := VisibleTotalBandPolygonsFromContours(
		path.TotalBandFootprints, path.TotalBandContours,
		path.NorthernTotalLimit, path.SouthernTotalLimit, path.TotalRiseSetCurves,
	)
	if err != nil {
		t.Fatalf("VisibleTotalBandPolygonsFromContours: %v", err)
	}
	if !authoritative {
		t.Fatal("total band should be authoritative for contour-driven Mars path")
	}
	if len(polygons) != 1 {
		t.Fatalf("total-band polygons=%d, want one merged face", len(polygons))
	}
	for _, test := range []struct {
		name   string
		probe  geodata.GeoPoint
		inside bool
	}{
		// This point was previously rejected by the phase-cycle envelope, but it
		// lies inside the horizon-closed total footprint during the event. The
		// static band now represents the full time union, so it must remain.
		{name: "visible-sweep", probe: geodata.GeoPoint{Longitude: -129.8230, Latitude: -77.4030}, inside: true},
		{name: "selected-west", probe: geodata.GeoPoint{Longitude: -134.6060, Latitude: -78.5601}, inside: true},
		{name: "selected-center", probe: geodata.GeoPoint{Longitude: -134.2300, Latitude: -78.0725}, inside: true},
	} {
		inside := geodata.SphericalPolygonsContainPoints(polygons, []geodata.GeoPoint{test.probe})[0]
		if inside != test.inside {
			t.Fatalf("merged total-band probe %s (%.4f, %.4f) inside=%v, want %v", test.name, test.probe.Longitude, test.probe.Latitude, inside, test.inside)
		}
	}
}
