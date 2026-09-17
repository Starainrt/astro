package geojson_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

// A one-limit polar total eclipse has U1/U4 samples while the axis is still
// below the horizon. Those open samples must not flare the static central band
// beyond the central-line interval.
func TestMarshalSolarEclipse15220327ClipsOneLimitEndSweeps(t *testing.T) {
	date := time.Date(1522, time.March, 27, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, eclipse.SolarEclipsePartialFootprintOptions{
		Step: 2 * time.Minute, BoundaryPoints: 96, RiseSetStep: 2 * time.Minute,
	})
	if !ok {
		t.Fatal("expected solar eclipse footprints")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, eclipse.SolarEclipsePathOptions{
		Step: 2 * time.Minute, TargetSpacingKM: 150,
	})
	if !ok || central.Eclipse.Centrality != eclipse.SolarEclipseCentralOneLimit {
		t.Fatalf("expected one-limit central path, got ok=%v centrality=%s", ok, central.Eclipse.Centrality)
	}
	if len(central.CenterLine) < 2 || len(partial.CentralBandFootprints) == 0 {
		t.Fatal("missing central path or end footprints")
	}
	first := partial.CentralBandFootprints[0]
	if !first.Time.Before(central.CenterLine[0].Time) || len(first.Boundaries) == 0 || len(first.Boundaries[0]) == 0 {
		t.Fatal("fixture no longer exercises a pre-horizon end footprint")
	}
	data, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil {
		t.Fatalf("MarshalSolarEclipse: %v", err)
	}
	band := featureWithRole(t, decodeCollection(t, data), "central-band")
	for _, point := range first.Boundaries[0] {
		if geometryContainsPoint(t, band.Geometry, point.Longitude, point.Latitude) {
			t.Fatalf("central band includes pre-horizon footprint point %.6f, %.6f", point.Longitude, point.Latitude)
		}
	}
	var lines [][][]float64
	center := featureWithRole(t, decodeCollection(t, data), "center-line")
	if err := json.Unmarshal(center.Geometry.Coordinates, &lines); err != nil {
		t.Fatal(err)
	}
	for _, line := range lines {
		for _, point := range line {
			if !geometryContainsPoint(t, band.Geometry, point[0], point[1]) {
				t.Fatalf("central band omits center-line point %.6f, %.6f", point[0], point[1])
			}
		}
	}
}
