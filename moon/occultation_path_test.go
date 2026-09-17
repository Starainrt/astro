package moon

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestStarOccultationBoundaryTimesRemainStrictAfterAdaptiveConvergence(t *testing.T) {
	data, err := basic.StarDataByChinese("进贤增九")
	if err != nil {
		t.Fatal(err)
	}
	star, err := StarCoordinateFromStarData(data)
	if err != nil {
		t.Fatal(err)
	}
	for _, date := range []time.Time{
		time.Date(2033, time.March, 17, 0, 0, 0, 0, time.UTC),
		time.Date(2050, time.December, 8, 0, 0, 0, 0, time.UTC),
	} {
		paths, err := FindStarOccultationPaths(date, date.Add(24*time.Hour), star,
			OccultationPathOptions{Step: 10 * time.Minute, DisableFootprints: true, DisableRiseSet: true})
		if err != nil {
			t.Fatalf("%s: FindStarOccultationPaths: %v", date.Format("2006-01-02"), err)
		}
		if len(paths) == 0 {
			t.Fatalf("%s: no occultation path", date.Format("2006-01-02"))
		}
		for _, points := range [][]OccultationPathPoint{paths[0].NorthernLimit, paths[0].SouthernLimit, paths[0].CenterLine} {
			for index := 1; index < len(points); index++ {
				if !points[index].Time.After(points[index-1].Time) {
					t.Fatalf("%s: non-increasing path time at %d: %s then %s", date.Format("2006-01-02"), index, points[index-1].Time, points[index].Time)
				}
			}
		}
	}
}

func TestFindStarOccultationPathsHR4799(t *testing.T) {
	location := time.FixedZone("CST", 8*3600)
	star := StarCoordinate{
		ID:                             "HR 4799",
		RA:                             189.1975,
		Dec:                            -5.831944444444,
		Epoch:                          time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		Frame:                          CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -28,
		ProperMotionDecMasPerYear:      -18,
	}
	paths, err := FindStarOccultationPaths(
		time.Date(2025, 6, 5, 0, 0, 0, 0, location),
		time.Date(2025, 6, 6, 0, 0, 0, 0, location),
		star, OccultationPathOptions{
			Step:            5 * time.Minute,
			TargetSpacingKM: 200,
		})
	if err != nil {
		t.Fatalf("FindStarOccultationPaths() error = %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("FindStarOccultationPaths() returned %d paths, want 1", len(paths))
	}
	path := paths[0]
	t.Logf("start=%s %.6f %.6f greatest=%s %.6f %.6f width=%.3f end=%s %.6f %.6f center=%d north=%d south=%d",
		path.Start.Time, path.Start.Longitude, path.Start.Latitude,
		path.Greatest.Time, path.Greatest.Longitude, path.Greatest.Latitude, path.Greatest.WidthKM,
		path.End.Time, path.End.Longitude, path.End.Latitude,
		len(path.CenterLine), len(path.NorthernLimit), len(path.SouthernLimit))
	if !path.Start.Time.Before(path.Greatest.Time) || !path.Greatest.Time.Before(path.End.Time) {
		t.Fatalf("path times are not ordered: start=%v greatest=%v end=%v", path.Start.Time, path.Greatest.Time, path.End.Time)
	}
	if !path.Complete {
		t.Fatal("global path must be marked complete")
	}
	if math.Abs(path.Start.Time.Sub(time.Date(2025, 6, 5, 17, 45, 26, 200000000, location)).Seconds()) > 5 {
		t.Fatalf("path start is too far from reference: got %v", path.Start.Time)
	}
	if math.Abs(path.End.Time.Sub(time.Date(2025, 6, 5, 22, 18, 53, 100000000, location)).Seconds()) > 5 {
		t.Fatalf("path end is too far from reference: got %v", path.End.Time)
	}
	if math.Abs(path.Start.Longitude-63.97388) > 0.5 || math.Abs(path.Start.Latitude-39.11772) > 0.5 {
		t.Fatalf("path start point is too far from reference: got %.6f %.6f", path.Start.Longitude, path.Start.Latitude)
	}
	if math.Abs(path.End.Longitude-172.33859) > 0.5 || math.Abs(path.End.Latitude-(-16.67159)) > 0.5 {
		t.Fatalf("path end point is too far from reference: got %.6f %.6f", path.End.Longitude, path.End.Latitude)
	}
	if math.Abs(path.Greatest.Longitude-121.55381) > 0.1 || math.Abs(path.Greatest.Latitude-6.79657) > 0.1 {
		t.Fatalf("greatest point is too far from reference: got %.6f %.6f", path.Greatest.Longitude, path.Greatest.Latitude)
	}
	if len(path.CenterLine) == 0 || len(path.NorthernLimit) == 0 || len(path.SouthernLimit) == 0 {
		t.Fatalf("path limits missing: center=%d north=%d south=%d", len(path.CenterLine), len(path.NorthernLimit), len(path.SouthernLimit))
	}
	if math.Abs(path.Greatest.Time.Sub(time.Date(2025, 6, 5, 20, 2, 7, 700000000, location)).Seconds()) > 5 {
		t.Fatalf("greatest time is too far from reference: got %v", path.Greatest.Time)
	}
	if math.Abs(path.Greatest.WidthKM-3571.9) > 20 {
		t.Fatalf("unexpected occultation width: got %.3f km", path.Greatest.WidthKM)
	}
	if len(path.NorthernLimit) != len(path.SouthernLimit) {
		t.Fatalf("limit counts differ: north=%d south=%d", len(path.NorthernLimit), len(path.SouthernLimit))
	}
	last := len(path.NorthernLimit) - 1
	if last < 1 || !path.NorthernLimit[0].Time.Equal(path.Start.Time) || !path.SouthernLimit[0].Time.Equal(path.Start.Time) ||
		!path.NorthernLimit[last].Time.Equal(path.End.Time) || !path.SouthernLimit[last].Time.Equal(path.End.Time) {
		t.Fatalf("limits do not cover the global start/end")
	}
	best, err := FindBestStarOccultations(
		time.Date(2025, 6, 5, 0, 0, 0, 0, location),
		time.Date(2025, 6, 6, 0, 0, 0, 0, location),
		star, OccultationSearchOptions{MaxEvents: 1},
	)
	if err != nil || len(best) != 1 {
		t.Fatalf("FindBestStarOccultations() = %d events, %v; want one event", len(best), err)
	}
	if math.Abs(best[0].Greatest.Sub(path.Greatest.Time).Seconds()) > 0.1 ||
		math.Abs(best[0].Observer.Longitude-path.Greatest.Longitude) > 1e-6 ||
		math.Abs(best[0].Observer.Latitude-path.Greatest.Latitude) > 1e-6 {
		t.Fatalf("best observer does not match global greatest: best=%v %+v path=%v %.9f %.9f",
			best[0].Greatest, best[0].Observer, path.Greatest.Time, path.Greatest.Longitude, path.Greatest.Latitude)
	}
}

func TestFindStarOccultationPathsHR4799LimitsRemainContinuous(t *testing.T) {
	location := time.FixedZone("CST", 8*3600)
	paths, err := FindStarOccultationPaths(
		time.Date(2025, 6, 5, 0, 0, 0, 0, location),
		time.Date(2025, 6, 6, 0, 0, 0, 0, location),
		hr4799CoordinateForTest(),
		OccultationPathOptions{Step: 30 * time.Second},
	)
	if err != nil {
		t.Fatalf("FindStarOccultationPaths() error = %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("FindStarOccultationPaths() returned %d paths, want 1", len(paths))
	}

	for _, limit := range []struct {
		name   string
		points []OccultationPathPoint
	}{
		{name: "northern", points: paths[0].NorthernLimit},
		{name: "southern", points: paths[0].SouthernLimit},
	} {
		for i := 1; i < len(limit.points); i++ {
			distance := occultationPathPointDistanceKM(limit.points[i-1], limit.points[i])
			if distance > 750 {
				t.Fatalf("%s limit jumps %.1f km between %v and %v", limit.name, distance,
					limit.points[i-1].Time, limit.points[i].Time)
			}
		}
		for i := 1; i+1 < len(limit.points); i++ {
			detour := occultationPathPointDistanceKM(limit.points[i-1], limit.points[i]) +
				occultationPathPointDistanceKM(limit.points[i], limit.points[i+1]) -
				occultationPathPointDistanceKM(limit.points[i-1], limit.points[i+1])
			if detour > 100 {
				t.Fatalf("%s limit has a %.1f km three-point detour at %v", limit.name, detour, limit.points[i].Time)
			}
		}
	}
}

func occultationPathPointDistanceKM(a, b OccultationPathPoint) float64 {
	lat1 := a.Latitude * math.Pi / 180
	lat2 := b.Latitude * math.Pi / 180
	dLat := lat2 - lat1
	dLon := (b.Longitude - a.Longitude) * math.Pi / 180
	dLon = math.Atan2(math.Sin(dLon), math.Cos(dLon))
	h := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * 6378.1366 * math.Asin(math.Sqrt(math.Min(1, h)))
}

func TestFindStarOccultationPathsReturnsFullPathOutsideQueryWindow(t *testing.T) {
	location := time.FixedZone("CST", 8*3600)
	greatest := time.Date(2025, 6, 5, 20, 2, 7, 700000000, location)
	start := greatest.Add(-time.Minute)
	end := greatest.Add(time.Minute)
	paths, err := FindStarOccultationPaths(start, end, hr4799CoordinateForTest(), OccultationPathOptions{Step: 5 * time.Minute})
	if err != nil {
		t.Fatalf("FindStarOccultationPaths() error = %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("FindStarOccultationPaths() returned %d paths, want 1", len(paths))
	}
	path := paths[0]
	if !path.Complete || !path.Start.Time.Before(start) || !path.End.Time.After(end) {
		t.Fatalf("path was clipped to query window: complete=%v start=%v window=%v..%v end=%v", path.Complete, path.Start.Time, start, end, path.End.Time)
	}
}

func TestGlobalStarOccultationQueriesSelectByGeometricGreatest(t *testing.T) {
	star := hr4799CoordinateForTest()
	dayStart := time.Date(2025, 6, 5, 0, 0, 0, 0, time.UTC)
	paths, err := FindStarOccultationPaths(
		dayStart, dayStart.Add(24*time.Hour), star,
		OccultationPathOptions{Step: 5 * time.Minute},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("full-day FindStarOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}

	start := paths[0].Greatest.Time.Add(-500 * time.Microsecond)
	end := start.Add(time.Millisecond)
	narrowPaths, err := FindStarOccultationPaths(start, end, star, OccultationPathOptions{Step: 5 * time.Minute})
	if err != nil || len(narrowPaths) != 1 {
		t.Fatalf("narrow FindStarOccultationPaths() paths=%d err=%v, want one", len(narrowPaths), err)
	}
	if !occultationTimeWithinNumericalWindowForTest(narrowPaths[0].Greatest.Time, start, end) {
		t.Fatalf("path greatest %v is outside query window %v..%v", narrowPaths[0].Greatest.Time, start, end)
	}

	best, err := FindBestStarOccultations(start, end, star, OccultationSearchOptions{})
	if err != nil || len(best) != 1 {
		t.Fatalf("narrow FindBestStarOccultations() events=%d err=%v, want one", len(best), err)
	}
	if !occultationTimeWithinNumericalWindowForTest(best[0].Greatest, start, end) {
		t.Fatalf("best greatest %v is outside query window %v..%v", best[0].Greatest, start, end)
	}
}

func occultationTimeWithinNumericalWindowForTest(value, start, end time.Time) bool {
	const tolerance = 10 * time.Millisecond
	return !value.Before(start.Add(-tolerance)) && !value.After(end.Add(tolerance))
}

func TestOccultationPathOptionsValidate(t *testing.T) {
	if err := (OccultationPathOptions{Step: -time.Second}).Validate(); err == nil {
		t.Fatal("negative path step should be rejected")
	}
	if err := (OccultationPathOptions{Step: 500 * time.Millisecond}).Validate(); err == nil {
		t.Fatal("subsecond positive path step should be rejected")
	}
	if err := (OccultationPathOptions{Step: time.Second}).Validate(); err != nil {
		t.Fatalf("one-second path step rejected: %v", err)
	}
	if err := (OccultationPathOptions{TargetSpacingKM: math.NaN()}).Validate(); err == nil {
		t.Fatal("non-finite target spacing should be rejected")
	}
	if err := (OccultationPathOptions{TargetSpacingKM: 0.5}).Validate(); err == nil {
		t.Fatal("sub-kilometer target spacing should be rejected")
	}
	if err := (OccultationPathOptions{TargetSpacingKM: 1}).Validate(); err != nil {
		t.Fatalf("one-kilometer target spacing rejected: %v", err)
	}
}

func BenchmarkFindStarOccultationPathTargetSpacing(b *testing.B) {
	location := time.FixedZone("CST", 8*3600)
	star := hr4799CoordinateForTest()
	for _, spacing := range []float64{1, 10, 200} {
		b.Run(fmt.Sprintf("%.0fkm", spacing), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				paths, err := FindStarOccultationPaths(
					time.Date(2025, 6, 5, 0, 0, 0, 0, location),
					time.Date(2025, 6, 6, 0, 0, 0, 0, location),
					star, OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: spacing},
				)
				if err != nil || len(paths) != 1 {
					b.Fatalf("FindStarOccultationPaths() paths=%d err=%v", len(paths), err)
				}
			}
		})
	}
}
