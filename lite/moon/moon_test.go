package moon

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	lite "github.com/starainrt/astro/lite/internal"
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

// 四个地平入口须只求一次站心位置：参考实现按旧口径对同一时刻重复求值，逐位对照。
func TestHorizontalEntriesMatchDuplicatedEvaluation(t *testing.T) {
	type site struct {
		name     string
		lon, lat float64
	}
	sites := []site{
		{"shanghai", 121.4737, 31.2304},
		{"sydney", 151.2093, -33.8688},
		{"north-pole", 0, 89.9999},
		{"south-pole", 0, -89.9999},
		{"dateline-west", -179.99, 12},
		{"dateline-east", 179.99, -12},
		{"equator", 0, 0},
	}
	zones := []*time.Location{
		time.UTC,
		time.FixedZone("CST", 8*3600),
		time.FixedZone("EST", -5*3600),
		time.FixedZone("LMT", -7*3600-52*60-58),
	}
	years := []int{-500, 1000, 1582, 2025, 2100, 3000, 4000}
	hours := []int{0, 5, 12, 20, 23}

	type sample struct {
		label string
		date  time.Time
		lon   float64
		lat   float64
	}
	samples := make([]sample, 0, len(sites)*len(zones)*len(years)*len(hours)+200)
	for _, st := range sites {
		for _, zone := range zones {
			for _, year := range years {
				for _, hour := range hours {
					samples = append(samples, sample{
						label: st.name + "/" + zone.String(),
						date:  time.Date(year, 3, 17, hour, 43, 21, 123456789, zone),
						lon:   st.lon,
						lat:   st.lat,
					})
				}
			}
		}
	}
	rng := rand.New(rand.NewSource(20260915))
	for i := 0; i < 200; i++ {
		zone := time.FixedZone("random", (rng.Intn(97)-48)*1800)
		samples = append(samples, sample{
			label: "random",
			date: time.Date(
				rng.Intn(8000)-2000, time.Month(1+rng.Intn(12)), 1+rng.Intn(28),
				rng.Intn(24), rng.Intn(60), rng.Intn(60), rng.Intn(1000000000), zone,
			),
			lon: rng.Float64()*360 - 180,
			lat: rng.Float64()*179.8 - 89.9,
		})
	}

	for _, s := range samples {
		jd := basic.Date2JD(s.date.UTC())
		altitude, azimuth, hourAngle := lite.HorizontalCoordinates(ApparentRa(s.date, s.lon, s.lat), ApparentDec(s.date, s.lon, s.lat), jd, s.lon, s.lat)
		checks := []struct {
			name string
			got  float64
			want float64
		}{
			{"HourAngle", HourAngle(s.date, s.lon, s.lat), hourAngle},
			{"Azimuth", Azimuth(s.date, s.lon, s.lat), azimuth},
			{"Altitude", Altitude(s.date, s.lon, s.lat), altitude},
			{"Zenith", Zenith(s.date, s.lon, s.lat), 90 - altitude},
		}
		for _, check := range checks {
			if check.got != check.want {
				t.Fatalf("%s %s %s: got %.17g want %.17g", check.name, s.label, s.date.Format(time.RFC3339Nano), check.got, check.want)
			}
		}
	}
}
