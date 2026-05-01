package moon

import (
	"math"
	"testing"
	"time"

	fullmoon "github.com/starainrt/astro/moon"
)

func TestLiteMoonGeocentricAgainstFullPrecision(t *testing.T) {
	samples := []time.Time{
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 14, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 1, 6, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 21, 18, 0, 0, 0, time.UTC),
		time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 10, 5, 6, 0, 0, 0, time.UTC),
		time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC),
	}

	for _, sample := range samples {
		if got, want := TrueLo(sample), fullmoon.TrueLo(sample); math.Abs(angleDiff(got, want)) > 0.20 {
			t.Fatalf("TrueLo(%s) = %.6f, want %.6f", sample.Format(time.RFC3339), got, want)
		}
		if got, want := TrueBo(sample), fullmoon.TrueBo(sample); math.Abs(got-want) > 0.06 {
			t.Fatalf("TrueBo(%s) = %.6f, want %.6f", sample.Format(time.RFC3339), got, want)
		}
	}
}

func TestLiteMoonPhaseAgainstFullPrecision(t *testing.T) {
	samples := []time.Time{
		time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 18, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 25, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
	}

	for _, sample := range samples {
		if got, want := Phase(sample), fullmoon.Phase(sample); math.Abs(got-want) > 0.03 {
			t.Fatalf("Phase(%s) = %.6f, want %.6f", sample.Format(time.RFC3339), got, want)
		}
	}
}

func TestLiteMoonRiseSetAgainstFullPrecision(t *testing.T) {
	cases := []struct {
		name string
		date time.Time
		lon  float64
		lat  float64
	}{
		{"Shanghai", time.Date(2026, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)), 121.4737, 31.2304},
		{"London", time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC), -0.1278, 51.5074},
		{"NewYork", time.Date(2026, 4, 16, 0, 0, 0, 0, time.FixedZone("EST", -5*3600)), -74.0060, 40.7128},
		{"Sydney", time.Date(2026, 8, 14, 0, 0, 0, 0, time.FixedZone("AEST", 10*3600)), 151.2093, -33.8688},
		{"Reykjavik", time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC), -21.8174, 64.1265},
	}

	for _, tc := range cases {
		gotRise, gotErr := RiseTime(tc.date, tc.lon, tc.lat, 0, true)
		wantRise, wantErr := fullmoon.RiseTime(tc.date, tc.lon, tc.lat, 0, true)
		if gotErr != nil || wantErr != nil {
			t.Fatalf("%s rise unexpected error: got=%v want=%v", tc.name, gotErr, wantErr)
		}
		assertTimeWithinMinutes(t, tc.name+" rise", gotRise, wantRise, 3.0)

		gotSet, gotSetErr := SetTime(tc.date, tc.lon, tc.lat, 0, true)
		wantSet, wantSetErr := fullmoon.SetTime(tc.date, tc.lon, tc.lat, 0, true)
		if gotSetErr != nil || wantSetErr != nil {
			t.Fatalf("%s set unexpected error: got=%v want=%v", tc.name, gotSetErr, wantSetErr)
		}
		assertTimeWithinMinutes(t, tc.name+" set", gotSet, wantSet, 3.0)
	}
}

func angleDiff(a, b float64) float64 {
	diff := math.Mod(a-b, 360)
	if diff > 180 {
		diff -= 360
	}
	if diff < -180 {
		diff += 360
	}
	return diff
}

func assertTimeWithinMinutes(t *testing.T, name string, got, want time.Time, limitMinutes float64) {
	t.Helper()
	if math.Abs(got.Sub(want).Minutes()) > limitMinutes {
		t.Fatalf("%s = %s, want %s", name, got, want)
	}
}
