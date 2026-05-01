package eclipse

import (
	"math"
	"testing"
	"time"
)

func TestSolarEclipseLocalDayBoundsRespectDST(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skipf("tzdata unavailable: %v", err)
	}

	testCases := []struct {
		name         string
		date         time.Time
		wantDuration time.Duration
	}{
		{
			name:         "spring forward 2025-03-09",
			date:         time.Date(2025, 3, 9, 8, 0, 0, 0, loc),
			wantDuration: 23 * time.Hour,
		},
		{
			name:         "fall back 2025-11-02",
			date:         time.Date(2025, 11, 2, 8, 0, 0, 0, loc),
			wantDuration: 25 * time.Hour,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dayStart, dayMid, dayEnd := solarEclipseLocalDayBounds(tc.date)

			if dayStart.Hour() != 0 || dayStart.Minute() != 0 || dayStart.Second() != 0 {
				t.Fatalf("dayStart should be local midnight, got %v", dayStart)
			}
			if dayMid.Hour() != 12 || dayMid.Minute() != 0 || dayMid.Second() != 0 {
				t.Fatalf("dayMid should be local noon, got %v", dayMid)
			}
			if dayEnd.Hour() != 0 || dayEnd.Minute() != 0 || dayEnd.Second() != 0 {
				t.Fatalf("dayEnd should be next local midnight, got %v", dayEnd)
			}
			if got := dayEnd.Sub(dayStart); got != tc.wantDuration {
				t.Fatalf("day length mismatch: got %v want %v", got, tc.wantDuration)
			}
		})
	}
}

func TestSolarEclipseOnDateByLocalDay(t *testing.T) {
	loc := time.FixedZone("UTC+14", 14*3600)

	testCases := []struct {
		name string
		date time.Time
		want bool
	}{
		{
			name: "day before no eclipse",
			date: time.Date(2024, 4, 8, 12, 0, 0, 0, loc),
			want: false,
		},
		{
			name: "local event day overlaps",
			date: time.Date(2024, 4, 9, 12, 0, 0, 0, loc),
			want: true,
		},
		{
			name: "day after no eclipse",
			date: time.Date(2024, 4, 10, 12, 0, 0, 0, loc),
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info, ok := SolarEclipseOnDate(tc.date)
			if ok != tc.want {
				t.Fatalf("SolarEclipseOnDate(%v) got %v want %v", tc.date, ok, tc.want)
			}
			if !ok {
				return
			}
			if info.Type != SolarEclipseTotal {
				t.Fatalf("unexpected eclipse type: got %s want %s", info.Type, SolarEclipseTotal)
			}
			if info.GreatestEclipse.Location() != loc {
				t.Fatalf("greatest eclipse location mismatch: got %q want %q", info.GreatestEclipse.Location(), loc)
			}
			if info.PartialBeginOnEarth.Day() != 9 || info.PartialEndOnEarth.Day() != 9 {
				t.Fatalf("unexpected local date span: begin=%v end=%v", info.PartialBeginOnEarth, info.PartialEndOnEarth)
			}
		})
	}
}

func TestSolarEclipseSearchSemantics(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	current := ClosestSolarEclipseNASABulletinSplitK(time.Date(2024, 4, 8, 12, 0, 0, 0, loc))
	if current.Type != SolarEclipseTotal {
		t.Fatalf("unexpected current eclipse type: got %s want %s", current.Type, SolarEclipseTotal)
	}

	assertSameSolarEclipse(t, "ClosestSolarEclipse(default)", ClosestSolarEclipse(current.GreatestEclipse), current, time.Second)

	last := LastSolarEclipseNASABulletinSplitK(current.GreatestEclipse)
	assertSameSolarEclipse(t, "LastSolarEclipseNASABulletinSplitK(current.GreatestEclipse)", last, current, time.Second)

	closest := ClosestSolarEclipseNASABulletinSplitK(current.GreatestEclipse)
	assertSameSolarEclipse(t, "ClosestSolarEclipseNASABulletinSplitK(current.GreatestEclipse)", closest, current, time.Second)

	next := NextSolarEclipseNASABulletinSplitK(current.GreatestEclipse)
	if !next.GreatestEclipse.After(current.GreatestEclipse) {
		t.Fatalf("NextSolarEclipseNASABulletinSplitK should be strictly future: current=%v next=%v", current.GreatestEclipse, next.GreatestEclipse)
	}
	if next.Type != SolarEclipseAnnular {
		t.Fatalf("unexpected next eclipse type: got %s want %s", next.Type, SolarEclipseAnnular)
	}

	wantNext := ClosestSolarEclipseNASABulletinSplitK(time.Date(2024, 10, 2, 12, 0, 0, 0, loc))
	assertSameSolarEclipse(t, "NextSolarEclipseNASABulletinSplitK(current.GreatestEclipse)", next, wantNext, time.Second)
}

func TestSolarEclipseInfoKeepsLocation(t *testing.T) {
	loc := time.FixedZone("UTC+08", 8*3600)
	testCases := []struct {
		name string
		calc func(time.Time) SolarEclipseInfo
	}{
		{name: "nasa", calc: ClosestSolarEclipseNASABulletinSplitK},
		{name: "iau", calc: ClosestSolarEclipseIAUSingleK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info := tc.calc(time.Date(2024, 10, 2, 12, 0, 0, 0, loc))

			if info.Type != SolarEclipseAnnular {
				t.Fatalf("unexpected eclipse type: got %s want %s", info.Type, SolarEclipseAnnular)
			}

			for _, item := range []struct {
				name string
				tm   time.Time
			}{
				{name: "GreatestEclipse", tm: info.GreatestEclipse},
				{name: "PartialBeginOnEarth", tm: info.PartialBeginOnEarth},
				{name: "PartialEndOnEarth", tm: info.PartialEndOnEarth},
				{name: "CentralBeginOnEarth", tm: info.CentralBeginOnEarth},
				{name: "CentralEndOnEarth", tm: info.CentralEndOnEarth},
			} {
				if item.tm.Location() != loc {
					t.Fatalf("%s location mismatch: got %q want %q", item.name, item.tm.Location(), loc)
				}
			}
		})
	}
}

func TestSolarEclipseIAUSingleKRemainsAvailable(t *testing.T) {
	date := time.Date(2024, 4, 8, 12, 0, 0, 0, time.UTC)
	defaultInfo := ClosestSolarEclipse(date)
	iauInfo := ClosestSolarEclipseIAUSingleK(date)

	if defaultInfo.Type != SolarEclipseTotal || iauInfo.Type != SolarEclipseTotal {
		t.Fatalf("unexpected eclipse types: default=%s iau=%s", defaultInfo.Type, iauInfo.Type)
	}
	assertSolarTimeClose(t, "GreatestEclipse", iauInfo.GreatestEclipse, defaultInfo.GreatestEclipse, time.Second)

	if !(iauInfo.PathWidthKM > defaultInfo.PathWidthKM) {
		t.Fatalf("expected IAU path width > NASA path width: iau=%.6f nasa=%.6f", iauInfo.PathWidthKM, defaultInfo.PathWidthKM)
	}
	if !(iauInfo.Magnitude > defaultInfo.Magnitude) {
		t.Fatalf("expected IAU magnitude > NASA magnitude: iau=%.9f nasa=%.9f", iauInfo.Magnitude, defaultInfo.Magnitude)
	}
}

func TestSolarEclipseAgainstNASAUTBaseline(t *testing.T) {
	testCases := []struct {
		name            string
		date            time.Time
		wantType        SolarEclipseType
		wantGreatest    time.Time
		wantGamma       float64
		wantMagnitude   float64
		wantLongitude   float64
		wantLatitude    float64
		wantPathWidthKM float64
		wantCentrality  SolarEclipseCentrality
	}{
		{
			name:            "2023-04-20 hybrid",
			date:            time.Date(2023, 4, 20, 0, 0, 0, 0, time.UTC),
			wantType:        SolarEclipseHybrid,
			wantGreatest:    time.Date(2023, 4, 20, 4, 16, 43, 0, time.UTC),
			wantGamma:       -0.3952,
			wantMagnitude:   1.0132,
			wantLongitude:   125.8,
			wantLatitude:    -9.6,
			wantPathWidthKM: 49.0,
			wantCentrality:  SolarEclipseCentralTwoLimits,
		},
		{
			name:            "2024-04-08 total",
			date:            time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC),
			wantType:        SolarEclipseTotal,
			wantGreatest:    time.Date(2024, 4, 8, 18, 17, 15, 0, time.UTC),
			wantGamma:       0.3431,
			wantMagnitude:   1.0566,
			wantLongitude:   -104.1,
			wantLatitude:    25.3,
			wantPathWidthKM: 197.5,
			wantCentrality:  SolarEclipseCentralTwoLimits,
		},
		{
			name:            "2024-10-02 annular",
			date:            time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC),
			wantType:        SolarEclipseAnnular,
			wantGreatest:    time.Date(2024, 10, 2, 18, 44, 59, 0, time.UTC),
			wantGamma:       -0.3509,
			wantMagnitude:   0.9326,
			wantLongitude:   -114.5,
			wantLatitude:    -22.0,
			wantPathWidthKM: 266.5,
			wantCentrality:  SolarEclipseCentralTwoLimits,
		},
		{
			name:            "2025-03-29 partial",
			date:            time.Date(2025, 3, 29, 0, 0, 0, 0, time.UTC),
			wantType:        SolarEclipsePartial,
			wantGreatest:    time.Date(2025, 3, 29, 10, 47, 21, 0, time.UTC),
			wantGamma:       1.0405,
			wantMagnitude:   0.9376,
			wantLongitude:   -77.1,
			wantLatitude:    61.1,
			wantPathWidthKM: 0.0,
			wantCentrality:  SolarEclipseNonCentral,
		},
	}

	const (
		// 这里对 UT 使用稍宽容差，因为 `sun` 层会把 TT 转成 UT，
		// 与 NASA 页面公开的 ΔT 口径存在数秒级差异；几何本身已在 basic 层用 TT 锁定。
		timeTolerance       = 8 * time.Second
		gammaTolerance      = 5e-4
		magnitudeTolerance  = 5e-4
		coordinateTolerance = 0.1
		pathWidthTolerance  = 5.0
	)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assertSameSolarEclipse(
				t,
				"ClosestSolarEclipse(default)",
				ClosestSolarEclipse(tc.date),
				ClosestSolarEclipseNASABulletinSplitK(tc.date),
				time.Second,
			)
			info := ClosestSolarEclipse(tc.date)
			if info.Type != tc.wantType {
				t.Fatalf("type mismatch: got %s want %s", info.Type, tc.wantType)
			}
			if info.Centrality != tc.wantCentrality {
				t.Fatalf("centrality mismatch: got %s want %s", info.Centrality, tc.wantCentrality)
			}

			assertSolarTimeClose(t, "GreatestEclipse", info.GreatestEclipse, tc.wantGreatest, timeTolerance)
			assertSolarFloatClose(t, "Gamma", info.Gamma, tc.wantGamma, gammaTolerance)
			assertSolarFloatClose(t, "Magnitude", info.Magnitude, tc.wantMagnitude, magnitudeTolerance)
			assertSolarFloatClose(t, "GreatestLongitude", info.GreatestLongitude, tc.wantLongitude, coordinateTolerance)
			assertSolarFloatClose(t, "GreatestLatitude", info.GreatestLatitude, tc.wantLatitude, coordinateTolerance)
			assertSolarFloatClose(t, "PathWidthKM", info.PathWidthKM, tc.wantPathWidthKM, pathWidthTolerance)
		})
	}
}

func assertSameSolarEclipse(t *testing.T, name string, got, want SolarEclipseInfo, tolerance time.Duration) {
	t.Helper()
	if got.Type != want.Type {
		t.Fatalf("%s type mismatch: got %s want %s", name, got.Type, want.Type)
	}
	assertSolarTimeClose(t, name+".GreatestEclipse", got.GreatestEclipse, want.GreatestEclipse, tolerance)
}

func assertSolarTimeClose(t *testing.T, name string, got, want time.Time, tolerance time.Duration) {
	t.Helper()
	diff := got.Sub(want)
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Fatalf("%s mismatch: got %v want %v diff=%v", name, got, want, diff)
	}
}

func assertSolarFloatClose(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s mismatch: got %.6f want %.6f diff=%.6f", name, got, want, math.Abs(got-want))
	}
}
