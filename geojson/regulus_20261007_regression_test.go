package geojson_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/moon"
)

func TestMarshalStarOccultationRegulus20261007UsesAnalyticVisibleEnvelope(t *testing.T) {
	if err := basic.LoadStarData(); err != nil {
		t.Fatal(err)
	}
	starData, err := basic.StarDataByHR(3982)
	if err != nil {
		t.Fatal(err)
	}
	star, err := moon.StarCoordinateFromStarData(starData)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.October, 7, 0, 0, 0, 0, time.UTC)
	paths, err := moon.FindStarOccultationPaths(start, start.Add(24*time.Hour), star, moon.OccultationPathOptions{})
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindStarOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	if len(path.Footprints) == 0 || len(path.BandFootprints) != 0 {
		t.Fatalf("dense path footprints=%d bandFootprints=%d, want dense mode", len(path.Footprints), len(path.BandFootprints))
	}
	data, err := geojson.MarshalStarOccultation(path)
	if err != nil {
		t.Fatalf("MarshalStarOccultation: %v", err)
	}
	collection := decodeCollection(t, data)
	band := featureWithRole(t, collection, "occultation-band")
	if authoritative, ok := band.Properties["static_band_authoritative"].(bool); !ok || !authoritative {
		t.Fatalf("occultation-band source=%v authoritative=%v, want analytic authoritative boundary",
			band.Properties["source"], band.Properties["static_band_authoritative"])
	}
	if source := band.Properties["source"]; source != "visible-footprint-sweep" {
		t.Fatalf("occultation-band source=%v, want visible-footprint-sweep", source)
	}
	assertClosedMultiPolygon(t, band)
	if rings := geoJSONMultiPolygonOuterRings(t, band); len(rings) != 1 {
		t.Fatalf("occultation-band polygons=%d, want one continuous visible band", len(rings))
	}
	assertOccultationBandMaximumEdge(t, band, 45)
	for _, feature := range featuresWithRole(collection, "visibility-boundary") {
		var lines [][][]float64
		if err := json.Unmarshal(feature.Geometry.Coordinates, &lines); err != nil {
			t.Fatalf("decode visibility-boundary: %v", err)
		}
		for lineIndex, line := range lines {
			for pointIndex, point := range line {
				if len(point) < 2 || !geometryContainsPointWithinKM(t, band.Geometry, point[0], point[1], 2) {
					t.Fatalf("visibility-boundary line %d point %d lies outside visible band at %.6f, %.6f", lineIndex, pointIndex, point[0], point[1])
				}
			}
		}
	}
	for index, point := range path.CenterLine {
		if !geometryContainsPoint(t, band.Geometry, point.Longitude, point.Latitude) {
			t.Fatalf("center-line sample %d lies outside visible band at %.6f, %.6f", index, point.Longitude, point.Latitude)
		}
	}
}
