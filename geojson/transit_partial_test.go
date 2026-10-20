package geojson_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/venus"
)

func TestMarshalVenusTransitGlobalPartial(tester *testing.T) {
	info := venus.NextTransit(time.Date(2854, 1, 1, 0, 0, 0, 0, time.UTC))
	data, err := geojson.MarshalVenusTransit(info, geojson.PlanetTransitOptions{IsochroneStep: 30 * time.Second})
	if err != nil {
		tester.Fatal(err)
	}
	collection, roles := decodeTransit(tester, data)
	for _, role := range []string{"transit-partial", "greatest", "c1-isochrone", "c4-isochrone"} {
		if roles[role] != 1 {
			tester.Fatalf("missing %s: %v", role, roles)
		}
	}
	for _, role := range []string{"c1-horizon", "c4-horizon", "visible-at-c1", "visible-at-c4", "c2-isochrone", "c3-isochrone", "transit-entire"} {
		if roles[role] != 0 {
			tester.Fatalf("fabricated %s", role)
		}
	}
	for _, item := range collection.Features {
		if item.Properties["role"] != "transit-partial" {
			continue
		}
		var polygons [][][][]float64
		if err := json.Unmarshal(item.Geometry.Coordinates, &polygons); err != nil {
			tester.Fatal(err)
		}
		if !geoJSONMultiPolygonContains(polygons, -135, -60) || geoJSONMultiPolygonContains(polygons, 0, 45) {
			tester.Fatal("partial-transit footprint does not match the topocentric geometry")
		}
		if item.Properties["has_geocentric_contacts"] != false || item.Properties["series"] != float64(4) || item.Properties["gamma"] == nil {
			tester.Fatalf("missing global-transit metadata: %v", item.Properties)
		}
	}
}
