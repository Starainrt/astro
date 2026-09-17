package geojson_test

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/moon"
)

// 导出属性必须同时给出 WidthKM（width_km）与限线间距（greatest_limit_separation_km），且两者口径不同。

func occultationLimitSeparationJSONValue(t *testing.T, feature decodedFeature, key string) float64 {
	t.Helper()
	raw, ok := feature.Properties[key]
	if !ok {
		t.Fatalf("feature role=%v has no %s property", feature.Properties["role"], key)
	}
	value, ok := raw.(float64)
	if !ok || math.IsNaN(value) || math.IsInf(value, 0) {
		t.Fatalf("feature role=%v %s=%v, want a finite number", feature.Properties["role"], key, raw)
	}
	return value
}

func TestMarshalPlanetOccultationCarriesGreatestLimitSeparation(t *testing.T) {
	start := time.Date(1962, time.October, 10, 0, 0, 0, 0, time.UTC)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationJupiter,
		moon.OccultationPathOptions{
			Step: 10 * time.Minute, TargetSpacingKM: 900,
			DisableFootprints: true, DisableRiseSet: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	data, err := geojson.MarshalPlanetOccultation(paths[0])
	if err != nil {
		t.Fatalf("MarshalPlanetOccultation: %v", err)
	}
	collection := decodeCollection(t, data)
	var greatestWidth, limitSeparation float64
	foundSeparation, foundGreatest := false, false
	for _, feature := range collection.Features {
		if value := occultationLimitSeparationJSONValue(t, feature, "greatest_limit_separation_km"); !foundSeparation {
			limitSeparation = value
			foundSeparation = true
		}
		if feature.Properties["role"] == "greatest" {
			greatestWidth = occultationLimitSeparationJSONValue(t, feature, "width_km")
			foundGreatest = true
		}
	}
	if !foundSeparation || !foundGreatest {
		t.Fatalf("limit separation=%v greatest width=%v, want both properties present", foundSeparation, foundGreatest)
	}
	if math.Abs(limitSeparation-3593.197135) > 0.5 {
		t.Fatalf("greatest_limit_separation_km=%.6f, want 3593.197135 within 0.5 km", limitSeparation)
	}
	if math.Abs(greatestWidth-5319.263900) > 0.5 {
		t.Fatalf("greatest width_km=%.6f, want 5319.263900 within 0.5 km", greatestWidth)
	}
	if math.Abs(greatestWidth-limitSeparation) < 1000 {
		t.Fatalf("width_km=%.3f and greatest_limit_separation_km=%.3f are too close, want the two distinct constructions",
			greatestWidth, limitSeparation)
	}
}

func TestMarshalPlanetOccultationRejectsNonFiniteLimitSeparation(t *testing.T) {
	start := time.Date(1962, time.October, 10, 0, 0, 0, 0, time.UTC)
	paths, err := moon.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), moon.OccultationJupiter,
		moon.OccultationPathOptions{
			Step: 10 * time.Minute, TargetSpacingKM: 900,
			DisableFootprints: true, DisableRiseSet: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	broken := paths[0]
	broken.NorthernLimit = append([]moon.OccultationPathPoint(nil), paths[0].NorthernLimit...)
	broken.NorthernLimit[1].LimitSeparationKM = math.NaN()
	if _, err := geojson.MarshalPlanetOccultation(broken); err == nil {
		t.Fatal("MarshalPlanetOccultation accepted a NaN limit separation on a limit point")
	}
	broken = paths[0]
	broken.GreatestLimitSeparationKM = math.Inf(1)
	if _, err := geojson.MarshalPlanetOccultation(broken); err == nil {
		t.Fatal("MarshalPlanetOccultation accepted a non-finite greatest limit separation")
	}
}

func TestMarshalStarOccultationCarriesGreatestLimitSeparation(t *testing.T) {
	start := time.Date(2025, time.June, 5, 0, 0, 0, 0, time.UTC)
	paths, err := moon.FindStarOccultationPaths(
		start, start.Add(24*time.Hour), hr4799OccultationCoordinateForGeoJSONLimitTest(),
		moon.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900,
			DisableFootprints: true, DisableRiseSet: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindStarOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	data, err := geojson.MarshalStarOccultation(paths[0])
	if err != nil {
		t.Fatalf("MarshalStarOccultation: %v", err)
	}
	collection := decodeCollection(t, data)
	value := occultationLimitSeparationJSONValue(t, collection.Features[0], "greatest_limit_separation_km")
	if math.Abs(value-3666.576690) > 0.5 {
		t.Fatalf("greatest_limit_separation_km=%.6f, want 3666.576690 within 0.5 km", value)
	}
}

func hr4799OccultationCoordinateForGeoJSONLimitTest() moon.StarCoordinate {
	return moon.StarCoordinate{
		ID: "HR 4799", RA: 189.1975, Dec: -5.831944444444,
		Epoch:                          time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC),
		Frame:                          moon.CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -28,
		ProperMotionDecMasPerYear:      -18,
	}
}
