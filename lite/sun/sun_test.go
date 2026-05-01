package sun

import (
	"math"
	"testing"
	"time"

	fullsun "github.com/starainrt/astro/sun"
)

func TestLiteSunPositionAgainstFullPrecision(t *testing.T) {
	samples := []time.Time{
		time.Date(2026, 1, 5, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 20, 6, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 21, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 22, 18, 0, 0, 0, time.UTC),
		time.Date(2026, 12, 21, 12, 0, 0, 0, time.UTC),
	}

	for _, sample := range samples {
		if got, want := ApparentLo(sample), fullsun.ApparentLo(sample); math.Abs(angleDiff(got, want)) > 0.02 {
			t.Fatalf("ApparentLo(%s) = %.6f, want %.6f", sample.Format(time.RFC3339), got, want)
		}
		gotRA, gotDec := ApparentRaDec(sample)
		wantRA, wantDec := fullsun.ApparentRaDec(sample)
		if math.Abs(angleDiff(gotRA, wantRA)) > 0.03 {
			t.Fatalf("ApparentRa(%s) = %.6f, want %.6f", sample.Format(time.RFC3339), gotRA, wantRA)
		}
		if math.Abs(gotDec-wantDec) > 0.03 {
			t.Fatalf("ApparentDec(%s) = %.6f, want %.6f", sample.Format(time.RFC3339), gotDec, wantDec)
		}
	}
}

func TestLiteSunRiseSetAgainstFullPrecision(t *testing.T) {
	cases := []struct {
		name string
		date time.Time
		lon  float64
		lat  float64
	}{
		{"Shanghai", time.Date(2026, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)), 121.4737, 31.2304},
		{"London", time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC), -0.1278, 51.5074},
		{"NewYork", time.Date(2026, 6, 21, 0, 0, 0, 0, time.FixedZone("EST", -5*3600)), -74.0060, 40.7128},
		{"Sydney", time.Date(2026, 9, 23, 0, 0, 0, 0, time.FixedZone("AEST", 10*3600)), 151.2093, -33.8688},
	}

	for _, tc := range cases {
		gotRise, gotErr := RiseTime(tc.date, tc.lon, tc.lat, 0, true)
		wantRise, wantErr := fullsun.RiseTime(tc.date, tc.lon, tc.lat, 0, true)
		if gotErr != nil || wantErr != nil {
			t.Fatalf("%s rise unexpected error: got=%v want=%v", tc.name, gotErr, wantErr)
		}
		assertTimeWithinMinutes(t, tc.name+" rise", gotRise, wantRise, 2.0)

		gotSet, gotSetErr := SetTime(tc.date, tc.lon, tc.lat, 0, true)
		wantSet, wantSetErr := fullsun.SetTime(tc.date, tc.lon, tc.lat, 0, true)
		if gotSetErr != nil || wantSetErr != nil {
			t.Fatalf("%s set unexpected error: got=%v want=%v", tc.name, gotSetErr, wantSetErr)
		}
		assertTimeWithinMinutes(t, tc.name+" set", gotSet, wantSet, 2.0)
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
