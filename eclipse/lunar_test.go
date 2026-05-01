package eclipse

import (
	"testing"
	"time"
)

func TestLunarEclipseLocalDayBoundsRespectDST(t *testing.T) {
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
			dayStart, dayMid, dayEnd := lunarEclipseLocalDayBounds(tc.date)

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

func TestLunarEclipseOnDateByLocalDay(t *testing.T) {
	loc := time.FixedZone("UTC-05", -5*3600)

	testCases := []struct {
		name string
		date time.Time
		want bool
	}{
		{
			name: "day before no eclipse",
			date: time.Date(2025, 3, 12, 12, 0, 0, 0, loc),
			want: false,
		},
		{
			name: "local start day overlaps",
			date: time.Date(2025, 3, 13, 12, 0, 0, 0, loc),
			want: true,
		},
		{
			name: "local end day overlaps",
			date: time.Date(2025, 3, 14, 12, 0, 0, 0, loc),
			want: true,
		},
		{
			name: "day after no eclipse",
			date: time.Date(2025, 3, 15, 12, 0, 0, 0, loc),
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info, ok := LunarEclipseOnDate(tc.date)
			if ok != tc.want {
				t.Fatalf("LunarEclipseOnDate(%v) got %v want %v", tc.date, ok, tc.want)
			}
			if !ok {
				return
			}
			if info.Type != LunarEclipseTotal {
				t.Fatalf("unexpected eclipse type: got %s want %s", info.Type, LunarEclipseTotal)
			}
			if info.Maximum.Location() != loc {
				t.Fatalf("maximum location mismatch: got %q want %q", info.Maximum.Location(), loc)
			}
			if info.PenumbralStart.Day() != 13 || info.PenumbralEnd.Day() != 14 {
				t.Fatalf("unexpected local date span: start=%v end=%v", info.PenumbralStart, info.PenumbralEnd)
			}
		})
	}
}

func TestLunarEclipseSearchSemantics(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	current := ClosestLunarEclipseDanjon(time.Date(2025, 3, 14, 12, 0, 0, 0, loc))
	if current.Type != LunarEclipseTotal {
		t.Fatalf("unexpected current eclipse type: got %s want %s", current.Type, LunarEclipseTotal)
	}
	assertSameEclipse(t, "ClosestLunarEclipse(default)", ClosestLunarEclipse(current.Maximum), current, time.Second)

	last := LastLunarEclipseDanjon(current.Maximum)
	assertSameEclipse(t, "LastLunarEclipseDanjon(current.Maximum)", last, current, time.Second)

	closest := ClosestLunarEclipseDanjon(current.Maximum)
	assertSameEclipse(t, "ClosestLunarEclipseDanjon(current.Maximum)", closest, current, time.Second)

	next := NextLunarEclipseDanjon(current.Maximum)
	if !next.Maximum.After(current.Maximum) {
		t.Fatalf("NextLunarEclipseDanjon should be strictly future: current=%v next=%v", current.Maximum, next.Maximum)
	}
	if next.Type != LunarEclipseTotal {
		t.Fatalf("unexpected next eclipse type: got %s want %s", next.Type, LunarEclipseTotal)
	}

	wantNextMax := time.Date(2025, 9, 8, 2, 12, 58, 0, loc)
	assertTimeClose(t, "next.Maximum", next.Maximum, wantNextMax, 2*time.Minute)
}

func TestLunarEclipseInfoKeepsLocation(t *testing.T) {
	loc := time.FixedZone("UTC+08", 8*3600)
	testCases := []struct {
		name string
		calc func(time.Time) LunarEclipseInfo
	}{
		{name: "danjon", calc: ClosestLunarEclipseDanjon},
		{name: "chauvenet", calc: ClosestLunarEclipseChauvenet},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info := tc.calc(time.Date(2023, 10, 29, 12, 0, 0, 0, loc))

			if info.Type != LunarEclipsePartial {
				t.Fatalf("unexpected eclipse type: got %s want %s", info.Type, LunarEclipsePartial)
			}

			for _, item := range []struct {
				name string
				tm   time.Time
			}{
				{name: "PenumbralStart", tm: info.PenumbralStart},
				{name: "PartialStart", tm: info.PartialStart},
				{name: "Maximum", tm: info.Maximum},
				{name: "PartialEnd", tm: info.PartialEnd},
				{name: "PenumbralEnd", tm: info.PenumbralEnd},
			} {
				if item.tm.Location() != loc {
					t.Fatalf("%s location mismatch: got %q want %q", item.name, item.tm.Location(), loc)
				}
			}
			for _, point := range info.ContactPoints {
				if point.Time.Location() != loc {
					t.Fatalf("contact %s location mismatch: got %q want %q", point.Label, point.Time.Location(), loc)
				}
			}
		})
	}
}

func TestLunarEclipseContactPoints(t *testing.T) {
	loc := time.FixedZone("UTC+08", 8*3600)
	info := ClosestLunarEclipse(time.Date(2026, 3, 3, 12, 0, 0, 0, loc))
	if info.Type != LunarEclipseTotal {
		t.Fatalf("unexpected eclipse type: got %s want %s", info.Type, LunarEclipseTotal)
	}
	if got, want := len(info.ContactPoints), 6; got != want {
		t.Fatalf("contact point count = %d, want %d", got, want)
	}

	points := make(map[string]LunarEclipseContactPoint, len(info.ContactPoints))
	for _, point := range info.ContactPoints {
		points[point.Label] = point
	}
	u1 := points["U1"]
	assertFloatClose(t, "U1.ContactPositionAngle", u1.ContactPositionAngle, 96.181711, 1e-3)
	assertFloatClose(t, "U1.ContactClockwiseAngle", u1.ContactClockwiseAngle, 263.818289, 1e-3)
	u2 := points["U2"]
	assertFloatClose(t, "U2.ContactPositionAngle", u2.ContactPositionAngle, 243.025171, 1e-3)
	assertFloatClose(t, "U2.MoonCenterPositionAngle", u2.MoonCenterPositionAngle, 243.025171, 1e-3)
}

func TestLunarEclipseChauvenetRemainsAvailable(t *testing.T) {
	date := time.Date(2025, 3, 14, 12, 0, 0, 0, time.UTC)
	defaultInfo := ClosestLunarEclipse(date)
	chauvenetInfo := ClosestLunarEclipseChauvenet(date)

	assertFloatClose(t, "Chauvenet.PenumbralMagnitude", chauvenetInfo.PenumbralMagnitude, 2.285431290, 1e-6)
	assertFloatClose(t, "Chauvenet.UmbralMagnitude", chauvenetInfo.UmbralMagnitude, 1.182811712, 1e-6)

	if !(chauvenetInfo.PenumbralMagnitude > defaultInfo.PenumbralMagnitude) {
		t.Fatalf("expected Chauvenet penumbral magnitude > Danjon: chauvenet=%.6f danjon=%.6f", chauvenetInfo.PenumbralMagnitude, defaultInfo.PenumbralMagnitude)
	}
	if !(chauvenetInfo.PenumbralStart.Before(defaultInfo.PenumbralStart) && chauvenetInfo.PenumbralEnd.After(defaultInfo.PenumbralEnd)) {
		t.Fatalf("expected Chauvenet penumbral span to be wider: chauvenet=(%v,%v) danjon=(%v,%v)", chauvenetInfo.PenumbralStart, chauvenetInfo.PenumbralEnd, defaultInfo.PenumbralStart, defaultInfo.PenumbralEnd)
	}
}

func TestLunarEclipseDefaultFallsBackForUltraShallowPenumbralEdge(t *testing.T) {
	date := time.Date(-780, 12, 13, 12, 0, 0, 0, time.UTC)

	defaultInfo := ClosestLunarEclipse(date)
	danjonInfo := ClosestLunarEclipseDanjon(date)
	chauvenetInfo := ClosestLunarEclipseChauvenet(date)

	if defaultInfo.Type != LunarEclipsePenumbral {
		t.Fatalf("default type mismatch: got %s want %s", defaultInfo.Type, LunarEclipsePenumbral)
	}
	if !defaultInfo.HasSaros || defaultInfo.Saros.Series != 61 || defaultInfo.Saros.Member != 1 || defaultInfo.Saros.Count != 78 {
		t.Fatalf("default shallow Saros mismatch: got has=%v saros=%+v", defaultInfo.HasSaros, defaultInfo.Saros)
	}
	if danjonInfo.Maximum.Equal(defaultInfo.Maximum) {
		t.Fatalf("default fallback should differ from explicit Danjon in this edge case: default=%v danjon=%v", defaultInfo.Maximum, danjonInfo.Maximum)
	}
	if !defaultInfo.Maximum.Equal(chauvenetInfo.Maximum) {
		t.Fatalf("default fallback should reuse Chauvenet edge event timing: default=%v chauvenet=%v", defaultInfo.Maximum, chauvenetInfo.Maximum)
	}
	if !(defaultInfo.PenumbralMagnitude > 0 && defaultInfo.PenumbralMagnitude <= lunarEclipseDefaultFallbackMaxPenumbralMagnitude) {
		t.Fatalf("default fallback penumbral magnitude out of narrow edge range: %.9f", defaultInfo.PenumbralMagnitude)
	}
}

func TestLunarEclipseAgainstNASABaseline(t *testing.T) {
	// NASA GSFC lunar eclipse catalog / plot pages:
	// - 2023 Oct 28 partial: LE2023Oct28P.pdf
	// - 2025 Mar 14 total: LE2025Mar14T.pdf
	testCases := []struct {
		name               string
		date               time.Time
		wantType           LunarEclipseType
		wantPenumbralMag   float64
		wantUmbralMag      float64
		wantPenumbralStart time.Time
		wantPartialStart   time.Time
		wantTotalStart     time.Time
		wantMaximum        time.Time
		wantTotalEnd       time.Time
		wantPartialEnd     time.Time
		wantPenumbralEnd   time.Time
	}{
		{
			name:               "2023-10-28 partial",
			date:               time.Date(2023, 10, 28, 0, 0, 0, 0, time.UTC),
			wantType:           LunarEclipsePartial,
			wantPenumbralMag:   1.1181,
			wantUmbralMag:      0.122,
			wantPenumbralStart: time.Date(2023, 10, 28, 18, 1, 43, 0, time.UTC),
			wantPartialStart:   time.Date(2023, 10, 28, 19, 35, 18, 0, time.UTC),
			wantMaximum:        time.Date(2023, 10, 28, 20, 14, 6, 0, time.UTC),
			wantPartialEnd:     time.Date(2023, 10, 28, 20, 52, 53, 0, time.UTC),
			wantPenumbralEnd:   time.Date(2023, 10, 28, 22, 26, 19, 0, time.UTC),
		},
		{
			name:               "2025-03-14 total",
			date:               time.Date(2025, 3, 14, 0, 0, 0, 0, time.UTC),
			wantType:           LunarEclipseTotal,
			wantPenumbralMag:   2.2595,
			wantUmbralMag:      1.1784,
			wantPenumbralStart: time.Date(2025, 3, 14, 3, 57, 28, 0, time.UTC),
			wantPartialStart:   time.Date(2025, 3, 14, 5, 9, 40, 0, time.UTC),
			wantTotalStart:     time.Date(2025, 3, 14, 6, 26, 6, 0, time.UTC),
			wantMaximum:        time.Date(2025, 3, 14, 6, 58, 41, 0, time.UTC),
			wantTotalEnd:       time.Date(2025, 3, 14, 7, 31, 26, 0, time.UTC),
			wantPartialEnd:     time.Date(2025, 3, 14, 8, 47, 56, 0, time.UTC),
			wantPenumbralEnd:   time.Date(2025, 3, 14, 10, 0, 9, 0, time.UTC),
		},
	}

	const timeTolerance = 2 * time.Minute
	const umbralMagnitudeTolerance = 0.02
	const penumbralMagnitudeTolerance = 0.1

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assertSameEclipse(t, "ClosestLunarEclipse(default)", ClosestLunarEclipse(tc.date), ClosestLunarEclipseDanjon(tc.date), time.Second)
			info := ClosestLunarEclipse(tc.date)
			if info.Type != tc.wantType {
				t.Fatalf("type mismatch: got %s want %s", info.Type, tc.wantType)
			}

			assertFloatClose(t, "PenumbralMagnitude", info.PenumbralMagnitude, tc.wantPenumbralMag, penumbralMagnitudeTolerance)
			assertFloatClose(t, "UmbralMagnitude", info.UmbralMagnitude, tc.wantUmbralMag, umbralMagnitudeTolerance)

			assertTimeClose(t, "PenumbralStart", info.PenumbralStart, tc.wantPenumbralStart, timeTolerance)
			assertTimeClose(t, "PartialStart", info.PartialStart, tc.wantPartialStart, timeTolerance)
			assertTimeClose(t, "TotalStart", info.TotalStart, tc.wantTotalStart, timeTolerance)
			assertTimeClose(t, "Maximum", info.Maximum, tc.wantMaximum, timeTolerance)
			assertTimeClose(t, "TotalEnd", info.TotalEnd, tc.wantTotalEnd, timeTolerance)
			assertTimeClose(t, "PartialEnd", info.PartialEnd, tc.wantPartialEnd, timeTolerance)
			assertTimeClose(t, "PenumbralEnd", info.PenumbralEnd, tc.wantPenumbralEnd, timeTolerance)
		})
	}
}

func TestPenumbralLunarEclipseKeepsNegativeUmbralMagnitude(t *testing.T) {
	info := ClosestLunarEclipse(time.Date(2024, 3, 25, 0, 0, 0, 0, time.UTC))
	if info.Type != LunarEclipsePenumbral {
		t.Fatalf("type mismatch: got %s want %s", info.Type, LunarEclipsePenumbral)
	}
	if !(info.UmbralMagnitude < 0) {
		t.Fatalf("expected negative umbral magnitude for penumbral eclipse, got %.12f", info.UmbralMagnitude)
	}
	if !(info.PenumbralMagnitude > 0) {
		t.Fatalf("expected positive penumbral magnitude, got %.12f", info.PenumbralMagnitude)
	}
}

func assertSameEclipse(t *testing.T, name string, got, want LunarEclipseInfo, tolerance time.Duration) {
	t.Helper()
	if got.Type != want.Type {
		t.Fatalf("%s type mismatch: got %s want %s", name, got.Type, want.Type)
	}
	assertTimeClose(t, name+".Maximum", got.Maximum, want.Maximum, tolerance)
}

func assertTimeClose(t *testing.T, name string, got, want time.Time, tolerance time.Duration) {
	t.Helper()
	diff := got.Sub(want)
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Fatalf("%s mismatch: got %v want %v diff=%v", name, got, want, diff)
	}
}

func assertFloatClose(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	diff := got - want
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Fatalf("%s mismatch: got %.6f want %.6f diff=%.6f", name, got, want, diff)
	}
}
