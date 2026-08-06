package moon

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func TestFindPlanetOccultationPathsSaturnFiniteDisk(t *testing.T) {
	start := time.Date(2024, time.August, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, time.August, 22, 0, 0, 0, 0, time.UTC)
	paths, err := FindPlanetOccultationPaths(
		start, end, OccultationSaturn,
		OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200},
	)
	if err != nil {
		t.Fatalf("FindPlanetOccultationPaths() error = %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() returned %d paths, want 1", len(paths))
	}
	path := paths[0]
	if path.Planet != OccultationSaturn || path.TargetID != "Saturn" {
		t.Fatalf("target = %v/%q, want Saturn", path.Planet, path.TargetID)
	}
	if !path.Complete {
		t.Fatal("planetary global path must be complete")
	}
	if !(path.Start.Time.Before(path.Greatest.Time) && path.Greatest.Time.Before(path.End.Time)) {
		t.Fatalf("global event order is invalid: start=%v greatest=%v end=%v", path.Start.Time, path.Greatest.Time, path.End.Time)
	}
	wantGreatest := time.Date(2024, time.August, 21, 2, 41, 36, 0, time.UTC)
	if delta := path.Greatest.Time.Sub(wantGreatest); delta < -2*time.Second || delta > 2*time.Second {
		t.Fatalf("global greatest differs from regression baseline by %v: got %v want %v", delta, path.Greatest.Time, wantGreatest)
	}
	if path.Greatest.WidthKM < 3000 || path.Greatest.WidthKM > 5000 {
		t.Fatalf("global outer width = %.3f km, want a physical lunar occultation band", path.Greatest.WidthKM)
	}
	if !path.HasTotalBand || !path.TotalComplete {
		t.Fatalf("global Saturn path total band = %v complete=%v, want a complete total-occultation band", path.HasTotalBand, path.TotalComplete)
	}
	if !(path.Start.Time.Before(path.TotalStart.Time) && path.TotalStart.Time.Before(path.Greatest.Time) &&
		path.Greatest.Time.Before(path.TotalEnd.Time) && path.TotalEnd.Time.Before(path.End.Time)) {
		t.Fatalf("global total-band order is invalid: outer=%v..%v total=%v..%v greatest=%v",
			path.Start.Time, path.End.Time, path.TotalStart.Time, path.TotalEnd.Time, path.Greatest.Time)
	}
	if path.GreatestTotalWidthKM <= 0 || path.GreatestTotalWidthKM >= path.Greatest.WidthKM {
		t.Fatalf("global total width = %.3f km, want positive and narrower than outer width %.3f km",
			path.GreatestTotalWidthKM, path.Greatest.WidthKM)
	}
	if len(path.CenterLine) < 3 || len(path.NorthernLimit) < 3 || len(path.NorthernLimit) != len(path.SouthernLimit) {
		t.Fatalf("incomplete global samples: center=%d north=%d south=%d", len(path.CenterLine), len(path.NorthernLimit), len(path.SouthernLimit))
	}
	last := len(path.NorthernLimit) - 1
	if !path.NorthernLimit[0].Time.Equal(path.Start.Time) || !path.SouthernLimit[0].Time.Equal(path.Start.Time) ||
		!path.NorthernLimit[last].Time.Equal(path.End.Time) || !path.SouthernLimit[last].Time.Equal(path.End.Time) {
		t.Fatal("planetary path limits do not span global start through end")
	}
	if len(path.NorthernTotalLimit) < 3 || len(path.NorthernTotalLimit) != len(path.SouthernTotalLimit) {
		t.Fatalf("incomplete global total samples: north=%d south=%d", len(path.NorthernTotalLimit), len(path.SouthernTotalLimit))
	}
	totalLast := len(path.NorthernTotalLimit) - 1
	if !path.NorthernTotalLimit[0].Time.Equal(path.TotalStart.Time) || !path.SouthernTotalLimit[0].Time.Equal(path.TotalStart.Time) ||
		!path.NorthernTotalLimit[totalLast].Time.Equal(path.TotalEnd.Time) || !path.SouthernTotalLimit[totalLast].Time.Equal(path.TotalEnd.Time) {
		t.Fatal("planetary total limits do not span global total start through end")
	}
	if len(path.PartialFootprints) < 3 || len(path.TotalFootprints) < 3 {
		t.Fatalf("incomplete planetary footprint sweep: partial=%d total=%d",
			len(path.PartialFootprints), len(path.TotalFootprints))
	}
	for _, footprints := range [][]PlanetOccultationFootprint{path.PartialFootprints, path.TotalFootprints} {
		for index, footprint := range footprints {
			if len(footprint.Polygons) == 0 {
				t.Fatalf("planetary footprint[%d] contains no visible polygons", index)
			}
		}
	}
	best, err := FindBestPlanetOccultations(start, end, OccultationSaturn, OccultationSearchOptions{MaxEvents: 1})
	if err != nil || len(best) != 1 {
		t.Fatalf("FindBestPlanetOccultations() = %d events, %v; want one event", len(best), err)
	}
	if math.Abs(best[0].Greatest.Sub(path.Greatest.Time).Seconds()) > 0.1 ||
		math.Abs(best[0].Observer.Longitude-path.Greatest.Longitude) > 1e-6 ||
		math.Abs(best[0].Observer.Latitude-path.Greatest.Latitude) > 1e-6 {
		t.Fatalf("best observer does not match global greatest: best=%v %+v path=%v %.9f %.9f",
			best[0].Greatest, best[0].Observer, path.Greatest.Time, path.Greatest.Longitude, path.Greatest.Latitude)
	}
}

func TestGlobalPlanetOccultationQueriesSelectByGeometricGreatest(t *testing.T) {
	dayStart := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
	paths, err := FindPlanetOccultationPaths(
		dayStart, dayStart.Add(24*time.Hour), OccultationSaturn,
		OccultationPathOptions{Step: 5 * time.Minute},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("full-day FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}

	start := paths[0].Greatest.Time.Add(-500 * time.Microsecond)
	end := start.Add(time.Millisecond)
	narrowPaths, err := FindPlanetOccultationPaths(start, end, OccultationSaturn, OccultationPathOptions{Step: 5 * time.Minute})
	if err != nil || len(narrowPaths) != 1 {
		t.Fatalf("narrow FindPlanetOccultationPaths() paths=%d err=%v, want one", len(narrowPaths), err)
	}
	if !occultationTimeWithinNumericalWindowForTest(narrowPaths[0].Greatest.Time, start, end) {
		t.Fatalf("path greatest %v is outside query window %v..%v", narrowPaths[0].Greatest.Time, start, end)
	}

	best, err := FindBestPlanetOccultations(start, end, OccultationSaturn, OccultationSearchOptions{})
	if err != nil || len(best) != 1 {
		t.Fatalf("narrow FindBestPlanetOccultations() events=%d err=%v, want one", len(best), err)
	}
	if !occultationTimeWithinNumericalWindowForTest(best[0].Greatest, start, end) {
		t.Fatalf("best greatest %v is outside query window %v..%v", best[0].Greatest, start, end)
	}
}

func BenchmarkFindPlanetOccultationPathTargetSpacing(b *testing.B) {
	for _, spacing := range []float64{1, 10, 200} {
		b.Run(fmt.Sprintf("%.0fkm", spacing), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				paths, err := FindPlanetOccultationPaths(
					time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 2, 2, 0, 0, 0, 0, time.UTC),
					OccultationSaturn,
					OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: spacing},
				)
				if err != nil || len(paths) != 1 {
					b.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v", len(paths), err)
				}
			}
		})
	}
}
