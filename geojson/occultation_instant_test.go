package geojson_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/moon"
)

func TestMarshalPlanetOccultationFootprintsUsesTimelineRoles(t *testing.T) {
	start := time.Date(2025, time.January, 5, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))
	events, err := moon.FindBestPlanetOccultations(
		start, start.Add(24*time.Hour), moon.OccultationSaturn, moon.OccultationSearchOptions{MaxEvents: 1},
	)
	if err != nil || len(events) != 1 {
		t.Fatalf("FindBestPlanetOccultations events=%d err=%v, want one", len(events), err)
	}
	at := events[0].Greatest
	instant, err := moon.PlanetOccultationFootprintsAt(at, moon.OccultationSaturn)
	if err != nil {
		t.Fatalf("PlanetOccultationFootprintsAt: %v", err)
	}
	raw, err := geojson.MarshalPlanetOccultationFootprints(instant)
	if err != nil {
		t.Fatalf("MarshalPlanetOccultationFootprints: %v", err)
	}
	roles, times := instantFootprintRoles(t, raw)
	if roles["partial-footprint"] != 1 || roles["total-footprint"] != 1 {
		t.Fatalf("roles=%v, want one partial and one total footprint", roles)
	}
	for role, value := range times {
		if value != at.UTC().Format(time.RFC3339Nano) {
			t.Fatalf("%s time=%q, want exact requested instant", role, value)
		}
	}
}

func TestMarshalStarOccultationFootprintUsesTimelineRole(t *testing.T) {
	star := moon.StarCoordinate{
		ID: "Antares", RA: 247.35166667, Dec: -26.43194444,
		Epoch: time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC),
		Frame: moon.CoordinateFrameJ2000,
	}
	start := time.Date(2024, time.March, 3, 0, 0, 0, 0, time.UTC)
	events, err := moon.FindBestStarOccultations(
		start, start.Add(24*time.Hour), star, moon.OccultationSearchOptions{MaxEvents: 1},
	)
	if err != nil || len(events) != 1 {
		t.Fatalf("FindBestStarOccultations events=%d err=%v, want one", len(events), err)
	}
	instant, err := moon.StarOccultationFootprintAt(events[0].Greatest, star)
	if err != nil {
		t.Fatalf("StarOccultationFootprintAt: %v", err)
	}
	if instant.Footprint == nil {
		t.Fatal("stellar footprint is nil at the event instant")
	}
	raw, err := geojson.MarshalStarOccultationFootprint(instant)
	if err != nil {
		t.Fatalf("MarshalStarOccultationFootprint: %v", err)
	}
	roles, _ := instantFootprintRoles(t, raw)
	if roles["occultation-footprint"] != 1 {
		t.Fatalf("roles=%v, want one stellar footprint", roles)
	}
}

func TestMarshalOccultationInstantOutsideEventReturnsEmptyCollection(t *testing.T) {
	at := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		marshal func() ([]byte, error)
	}{
		{
			name: "star",
			marshal: func() ([]byte, error) {
				return geojson.MarshalStarOccultationFootprint(moon.StarOccultationInstant{
					Time: at, TargetID: "outside-event",
				})
			},
		},
		{
			name: "planet",
			marshal: func() ([]byte, error) {
				return geojson.MarshalPlanetOccultationFootprints(moon.PlanetOccultationInstant{
					Time: at, Planet: moon.OccultationSaturn, TargetID: "Saturn",
				})
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			raw, err := test.marshal()
			if err != nil {
				t.Fatalf("marshal outside-event footprint: %v", err)
			}
			roles, _ := instantFootprintRoles(t, raw)
			if len(roles) != 0 {
				t.Fatalf("roles=%v, want empty FeatureCollection", roles)
			}
			var collection struct {
				Features []json.RawMessage `json:"features"`
			}
			if err := json.Unmarshal(raw, &collection); err != nil {
				t.Fatalf("decode empty FeatureCollection: %v", err)
			}
			if collection.Features == nil {
				t.Fatal("features is null, want an empty JSON array")
			}
		})
	}
}

func instantFootprintRoles(t *testing.T, raw []byte) (map[string]int, map[string]string) {
	t.Helper()
	var collection struct {
		Features []struct {
			Properties map[string]interface{} `json:"properties"`
		} `json:"features"`
	}
	if err := json.Unmarshal(raw, &collection); err != nil {
		t.Fatalf("decode GeoJSON: %v", err)
	}
	roles := make(map[string]int)
	times := make(map[string]string)
	for _, feature := range collection.Features {
		role, _ := feature.Properties["role"].(string)
		roles[role]++
		times[role], _ = feature.Properties["time"].(string)
	}
	return roles, times
}

func TestMarshalPlanetOccultationFootprintsCarriesInstantMetadata(t *testing.T) {
	start := time.Date(2025, time.January, 5, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))
	events, err := moon.FindBestPlanetOccultations(
		start, start.Add(24*time.Hour), moon.OccultationSaturn, moon.OccultationSearchOptions{MaxEvents: 1},
	)
	if err != nil || len(events) != 1 {
		t.Fatalf("FindBestPlanetOccultations events=%d err=%v, want one", len(events), err)
	}
	instant, err := moon.PlanetOccultationFootprintsAt(events[0].Greatest, moon.OccultationSaturn)
	if err != nil {
		t.Fatalf("PlanetOccultationFootprintsAt: %v", err)
	}
	raw, err := geojson.MarshalPlanetOccultationFootprints(instant)
	if err != nil {
		t.Fatalf("MarshalPlanetOccultationFootprints: %v", err)
	}
	var collection struct {
		Features []struct {
			Properties map[string]interface{} `json:"properties"`
		} `json:"features"`
	}
	if err := json.Unmarshal(raw, &collection); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(collection.Features) == 0 {
		t.Fatal("no instant features")
	}
	if instant.DeltaTSeconds <= 0 || instant.SublunarLongitude == 0 && instant.SublunarLatitude == 0 {
		t.Fatalf("instant metadata missing: ΔT=%.3f sublunar=(%.3f,%.3f)",
			instant.DeltaTSeconds, instant.SublunarLongitude, instant.SublunarLatitude)
	}
	for _, feature := range collection.Features {
		properties := feature.Properties
		deltaT, ok := properties["delta_t_seconds"].(float64)
		if !ok || deltaT <= 0 {
			t.Fatalf("delta_t_seconds=%v, want a positive value", properties["delta_t_seconds"])
		}
		if signature, _ := properties["interp_signature"].(string); signature == "" || signature == "empty" {
			t.Fatalf("interp_signature=%v, want a footprint signature", properties["interp_signature"])
		}
		closed, present := properties["source_boundary_closed"].(bool)
		if !present {
			t.Fatal("source_boundary_closed is missing")
		}
		closure, hasClosure := properties["closure"].(map[string]interface{})
		if closed {
			if hasClosure {
				t.Fatal("self-closed occultation footprint must not carry a closure")
			}
			continue
		}
		if !hasClosure || closure["kind"] != "target-horizon" || closure["body"] != "moon" {
			t.Fatalf("closure=%v, want kind=target-horizon body=moon", properties["closure"])
		}
		if _, present := closure["sublunar"]; !present {
			t.Fatal("target-horizon closure is missing the sublunar reference")
		}
		if properties["geometry_role"] != "horizon-closed-region" {
			t.Fatalf("geometry_role=%v, want horizon-closed-region", properties["geometry_role"])
		}
	}
}

func TestMarshalStarOccultationFootprintHorizonCutCarriesClosure(t *testing.T) {
	star := moon.StarCoordinate{
		ID: "Antares", RA: 247.35166667, Dec: -26.43194444,
		Epoch: time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC),
		Frame: moon.CoordinateFrameJ2000,
	}
	start := time.Date(2024, time.March, 3, 0, 0, 0, 0, time.UTC)
	events, err := moon.FindBestStarOccultations(start, start.Add(24*time.Hour), star, moon.OccultationSearchOptions{MaxEvents: 1})
	if err != nil || len(events) != 1 {
		t.Fatalf("FindBestStarOccultations events=%d err=%v, want one", len(events), err)
	}
	instant, err := moon.StarOccultationFootprintAt(events[0].Greatest.Add(-120*time.Minute), star)
	if err != nil {
		t.Fatalf("StarOccultationFootprintAt: %v", err)
	}
	if instant.Footprint == nil {
		t.Fatal("expected a footprint inside the event window")
	}
	if instant.Footprint.Closed {
		t.Skip("this instant is self-closed; the horizon-cut branch is covered elsewhere")
	}
	raw, err := geojson.MarshalStarOccultationFootprint(instant)
	if err != nil {
		t.Fatalf("MarshalStarOccultationFootprint: %v", err)
	}
	var collection struct {
		Features []struct {
			Properties map[string]interface{} `json:"properties"`
		} `json:"features"`
	}
	if err := json.Unmarshal(raw, &collection); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(collection.Features) != 1 {
		t.Fatalf("features=%d, want one", len(collection.Features))
	}
	closure, ok := collection.Features[0].Properties["closure"].(map[string]interface{})
	if !ok || closure["kind"] != "target-horizon" || closure["body"] != "moon" {
		t.Fatalf("closure=%v, want kind=target-horizon body=moon", collection.Features[0].Properties["closure"])
	}
}
