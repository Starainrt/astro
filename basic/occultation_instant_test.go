package basic

import (
	"errors"
	"testing"
	"time"
)

func TestPlanetOccultationFootprintsAtReturnsExactRequestedInstant(t *testing.T) {
	start := time.Date(2025, time.January, 5, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))
	events, err := FindBestPlanetOccultations(
		start, start.Add(24*time.Hour), OccultationSaturn, OccultationSearchOptions{MaxEvents: 1},
	)
	if err != nil || len(events) != 1 {
		t.Fatalf("FindBestPlanetOccultations events=%d err=%v, want one", len(events), err)
	}
	at := events[0].Greatest.Add(123 * time.Nanosecond)
	instant, err := PlanetOccultationFootprintsAt(at, OccultationSaturn)
	if err != nil {
		t.Fatalf("PlanetOccultationFootprintsAt: %v", err)
	}
	if instant.Partial == nil || instant.Total == nil {
		t.Fatalf("instant footprints partial=%v total=%v, want both", instant.Partial != nil, instant.Total != nil)
	}
	for name, footprint := range map[string]*OccultationFootprint{
		"partial": instant.Partial,
		"total":   instant.Total,
	} {
		if !footprint.Time.Equal(at) {
			t.Fatalf("%s footprint time=%v, want %v", name, footprint.Time, at)
		}
		if len(footprint.Polygons) == 0 || len(footprint.Boundaries) == 0 {
			t.Fatalf("%s footprint has no visible geometry", name)
		}
		for _, boundary := range footprint.Boundaries {
			for _, point := range boundary {
				if !point.Time.Equal(at) {
					t.Fatalf("%s boundary time=%v, want %v", name, point.Time, at)
				}
			}
		}
	}
}

func TestStarOccultationFootprintAtReturnsEmptyOutsideEvent(t *testing.T) {
	star := StarCoordinate{
		ID: "Antares", RA: 247.35166667, Dec: -26.43194444,
		Epoch: time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC),
		Frame: CoordinateFrameJ2000,
	}
	instant, err := StarOccultationFootprintAt(
		time.Date(2024, time.March, 1, 17, 0, 0, 0, time.UTC), star,
	)
	if err != nil {
		t.Fatalf("StarOccultationFootprintAt: %v", err)
	}
	if instant.Footprint != nil {
		t.Fatal("outside-event stellar footprint is not nil")
	}
}

func TestOccultationFootprintAtValidatesInputs(t *testing.T) {
	if _, err := PlanetOccultationFootprintsAt(time.Time{}, OccultationSaturn); !errors.Is(err, ErrInvalidOccultationInput) {
		t.Fatalf("zero time error=%v, want ErrInvalidOccultationInput", err)
	}
	if _, err := PlanetOccultationFootprintsAt(time.Now(), OccultationPlanet("earth")); !errors.Is(err, ErrInvalidOccultationInput) {
		t.Fatalf("invalid planet error=%v, want ErrInvalidOccultationInput", err)
	}
	if _, err := StarOccultationFootprintAt(time.Now(), StarCoordinate{}); !errors.Is(err, ErrInvalidOccultationInput) {
		t.Fatalf("invalid star error=%v, want ErrInvalidOccultationInput", err)
	}
}
