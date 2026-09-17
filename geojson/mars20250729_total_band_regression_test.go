package geojson_test

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/internal/occultationgeo"
)

func TestMars20250729TotalBandIsSingleMergedFace(t *testing.T) {
	collection := mars20250729TestFixture(t, time.Minute, true).collection
	totalBand := featureWithRole(t, collection, "total-band")
	rings := geoJSONMultiPolygonOuterRings(t, totalBand)
	if len(rings) != 1 {
		t.Fatalf("total-band polygons=%d, want one merged face", len(rings))
	}
	for _, test := range []struct {
		name   string
		probe  geodata.GeoPoint
		inside bool
	}{
		{name: "visible-sweep", probe: geodata.GeoPoint{Longitude: -129.8230, Latitude: -77.4030}, inside: true},
		{name: "selected-west", probe: geodata.GeoPoint{Longitude: -134.6060, Latitude: -78.5601}, inside: true},
		{name: "selected-center", probe: geodata.GeoPoint{Longitude: -134.2300, Latitude: -78.0725}, inside: true},
	} {
		inside := geodata.SphericalPolygonsContainPoints(rings, []geodata.GeoPoint{test.probe})[0]
		if inside != test.inside {
			t.Fatalf("total-band probe %s (%.4f, %.4f) inside=%v, want %v", test.name, test.probe.Longitude, test.probe.Latitude, inside, test.inside)
		}
	}
}

func TestMars20250729FinalTotalBandHasWidePolarShoulderRepair(t *testing.T) {
	collection := mars20250729TestFixture(t, time.Minute, true).collection
	partial := geoJSONMultiPolygonOuterRings(t, featureWithRole(t, collection, "partial-band"))
	total := geoJSONMultiPolygonOuterRings(t, featureWithRole(t, collection, "total-band"))

	// The final GeoJSON pass must already have applied the total-only broad
	// polar repair. Reapplying it should therefore be geometrically inert.
	rerounded := occultationgeo.RoundAuthoritativeTotalBandJunctions(total)
	remainingRepairKM := math.Max(
		geodata.SphericalPolygonsPathMissDistanceKM(total, rerounded, true),
		geodata.SphericalPolygonsPathMissDistanceKM(rerounded, total, true),
	)
	if remainingRepairKM > 1 {
		t.Fatalf("final total-band still has a %.1f km repairable polar shoulder", remainingRepairKM)
	}
	if miss := geodata.SphericalPolygonsPathMissDistanceKM(partial, total, true); miss > 1 {
		t.Fatalf("smoothed total-band lies %.1f km outside partial-band", miss)
	}

	roundedPartial := occultationgeo.RoundAuthoritativeBandJunctions(partial)
	partialChangeKM := math.Max(
		geodata.SphericalPolygonsPathMissDistanceKM(partial, roundedPartial, true),
		geodata.SphericalPolygonsPathMissDistanceKM(roundedPartial, partial, true),
	)
	if partialChangeKM > 1 {
		t.Fatalf("final partial-band unexpectedly needs %.1f km of further repair", partialChangeKM)
	}
}
