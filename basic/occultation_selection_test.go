package basic

import (
	"testing"
	"time"
)

func TestOccultationPathsSkipOutsideWindowBeforeSampling(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*3600)
	options := OccultationPathOptions{
		Step: time.Second, TargetSpacingKM: 900, RiseSetStep: time.Minute, DisableFootprints: true,
	}
	start := time.Date(2025, 1, 6, 0, 0, 0, 0, zone)
	planets, err := FindPlanetOccultationPaths(start, start.Add(24*time.Hour), OccultationSaturn, options)
	if err != nil || len(planets) != 0 {
		t.Fatalf("outside-window planet paths=%d, err=%v", len(planets), err)
	}
	start = time.Date(2024, 3, 4, 0, 0, 0, 0, zone)
	star := StarCoordinate{ID: "Antares", RA: 247.35166667, Dec: -26.43194444,
		Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameJ2000}
	stars, err := FindStarOccultationPaths(start, start.Add(24*time.Hour), star, options)
	if err != nil || len(stars) != 0 {
		t.Fatalf("outside-window star paths=%d, err=%v", len(stars), err)
	}
}
