package eclipse

import (
	"testing"
	"time"
)

func TestLocalLunarEclipseOnDateByLocalDay(t *testing.T) {
	loc := time.FixedZone("CDT", -5*3600)
	lon, lat, height := -87.65, 41.85, 0.0

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
			info, ok := LocalLunarEclipseOnDate(tc.date, lon, lat, height)
			if ok != tc.want {
				t.Fatalf("LocalLunarEclipseOnDate(%v) got %v want %v", tc.date, ok, tc.want)
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

func TestLocalLunarEclipseVisibilityFilter(t *testing.T) {
	chicagoLoc := time.FixedZone("CDT", -5*3600)
	chicagoDate := time.Date(2023, 10, 28, 12, 0, 0, 0, chicagoLoc)
	chicagoLon, chicagoLat, chicagoHeight := -87.65, 41.85, 0.0

	geometricInfo, geometricOK := GeometricLocalLunarEclipseOnDate(chicagoDate, chicagoLon, chicagoLat, chicagoHeight)
	if !geometricOK {
		t.Fatalf("expected geometric local eclipse on date")
	}
	if geometricInfo.Type != LunarEclipsePartial {
		t.Fatalf("unexpected geometric eclipse type: got %s want %s", geometricInfo.Type, LunarEclipsePartial)
	}
	if geometricInfo.VisibleAtMaximum {
		t.Fatalf("expected geometric eclipse to be below horizon at maximum: %+v", geometricInfo)
	}

	visibleInfo, visibleOK := LocalLunarEclipseOnDate(chicagoDate, chicagoLon, chicagoLat, chicagoHeight)
	if visibleOK {
		t.Fatalf("expected visible filter to reject invisible eclipse, got %+v", visibleInfo)
	}

	londonDate := time.Date(2025, 3, 14, 12, 0, 0, 0, time.UTC)
	londonInfo, londonOK := LocalLunarEclipseOnDate(londonDate, -0.1278, 51.5074, 0)
	if !londonOK {
		t.Fatalf("expected visible local eclipse in London")
	}
	if londonInfo.VisibleAtMaximum {
		t.Fatalf("expected London maximum to be below horizon, got %+v", londonInfo)
	}
}

func TestLocalLunarEclipseSearchSemantics(t *testing.T) {
	loc := time.UTC
	lon, lat, height := -0.1278, 51.5074, 0.0
	current := ClosestLocalLunarEclipseDanjon(time.Date(2025, 3, 14, 12, 0, 0, 0, loc), lon, lat, height)
	if current.Type != LunarEclipseTotal {
		t.Fatalf("unexpected current eclipse type: got %s want %s", current.Type, LunarEclipseTotal)
	}
	assertSameLocalLunarEclipse(t, "ClosestLocalLunarEclipse(default)", ClosestLocalLunarEclipse(current.Maximum, lon, lat, height), current, time.Second)

	last := LastLocalLunarEclipseDanjon(current.Maximum, lon, lat, height)
	assertSameLocalLunarEclipse(t, "LastLocalLunarEclipseDanjon(current.Maximum)", last, current, time.Second)

	closest := ClosestLocalLunarEclipseDanjon(current.Maximum, lon, lat, height)
	assertSameLocalLunarEclipse(t, "ClosestLocalLunarEclipseDanjon(current.Maximum)", closest, current, time.Second)

	next := NextLocalLunarEclipseDanjon(current.Maximum, lon, lat, height)
	if !next.Maximum.After(current.Maximum) {
		t.Fatalf("NextLocalLunarEclipseDanjon should be strictly future: current=%v next=%v", current.Maximum, next.Maximum)
	}
	if next.Type != LunarEclipseTotal {
		t.Fatalf("unexpected next eclipse type: got %s want %s", next.Type, LunarEclipseTotal)
	}

	wantNextMax := time.Date(2025, 9, 7, 18, 11, 49, 0, loc)
	assertTimeClose(t, "next.Maximum", next.Maximum, wantNextMax, 2*time.Minute)
}

func TestLocalLunarEclipseSearchBeyondFiveYears(t *testing.T) {
	loc := time.FixedZone("CDT", -5*3600)
	lon, lat, height := -87.65, 41.85, 0.0
	current := ClosestLocalLunarEclipseDanjon(time.Date(2025, 3, 14, 12, 0, 0, 0, loc), lon, lat, height)
	if current.Type != LunarEclipseTotal {
		t.Fatalf("unexpected current eclipse type: got %s want %s", current.Type, LunarEclipseTotal)
	}

	next := NextLocalLunarEclipseDanjon(current.Maximum, lon, lat, height)
	if next.Type == LunarEclipseNone || next.Maximum.IsZero() {
		t.Fatalf("expected a future visible local lunar eclipse beyond the old 60-lunation window")
	}
	if !next.Maximum.After(current.Maximum) {
		t.Fatalf("expected strictly future local lunar eclipse: current=%v next=%v", current.Maximum, next.Maximum)
	}
}

func TestLocalTotalLunarEclipseSearch(t *testing.T) {
	loc := time.FixedZone("CDT", -5*3600)
	lon, lat, height := -87.65, 41.85, 0.0
	date := time.Date(2025, 3, 13, 0, 0, 0, 0, loc)

	next, ok := NextLocalTotalLunarEclipse(date, lon, lat, height)
	if !ok {
		t.Fatal("expected to find next local total lunar eclipse")
	}
	if next.Type != LunarEclipseTotal || !next.HasTotal {
		t.Fatalf("unexpected next total lunar eclipse: %+v", next)
	}
	assertTimeClose(t, "NextLocalTotalLunarEclipse", next.Maximum, time.Date(2025, 3, 14, 1, 58, 47, 0, loc), 2*time.Minute)

	last, ok := LastLocalTotalLunarEclipse(next.Maximum, lon, lat, height)
	if !ok {
		t.Fatal("expected to find previous local total lunar eclipse")
	}
	if last.Type != LunarEclipseTotal || !last.HasTotal {
		t.Fatalf("unexpected last total lunar eclipse: %+v", last)
	}
	assertTimeClose(t, "LastLocalTotalLunarEclipse", last.Maximum, next.Maximum, time.Second)
}

func TestLocalTotalLunarEclipseClosest(t *testing.T) {
	loc := time.FixedZone("CDT", -5*3600)
	lon, lat, height := -87.65, 41.85, 0.0
	date := time.Date(2025, 3, 14, 0, 0, 0, 0, loc)

	info, ok := ClosestLocalTotalLunarEclipse(date, lon, lat, height)
	if !ok {
		t.Fatal("expected to find closest local total lunar eclipse")
	}
	if info.Type != LunarEclipseTotal || !info.HasTotal {
		t.Fatalf("unexpected closest total lunar eclipse: %+v", info)
	}
	assertTimeClose(t, "ClosestLocalTotalLunarEclipse", info.Maximum, time.Date(2025, 3, 14, 1, 58, 47, 0, loc), 2*time.Minute)
}

func TestLocalTotalLunarEclipseVisibleRequiresTotalPhaseVisibility(t *testing.T) {
	info, ok := LocalLunarEclipseOnDate(time.Date(2025, 3, 14, 12, 0, 0, 0, time.UTC), -0.1278, 51.5074, 0)
	if !ok {
		t.Fatalf("expected visible local eclipse in London")
	}
	if info.Type != LunarEclipseTotal || !info.HasTotal {
		t.Fatalf("unexpected eclipse type: %+v", info)
	}
	if !localLunarEclipseVisible(info) {
		t.Fatalf("expected some phase to be visible")
	}
	if localTotalLunarEclipseVisible(info) {
		t.Fatalf("expected total phase below horizon to be rejected")
	}
}

func TestLocalLunarEclipseInfoKeepsLocation(t *testing.T) {
	loc := time.FixedZone("JST", 9*3600)
	lon, lat, height := 139.6917, 35.6895, 1234.0
	testCases := []struct {
		name string
		calc func(time.Time, float64, float64, float64) LocalLunarEclipseInfo
	}{
		{name: "danjon", calc: ClosestLocalLunarEclipseDanjon},
		{name: "chauvenet", calc: ClosestLocalLunarEclipseChauvenet},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info := tc.calc(time.Date(2023, 10, 29, 12, 0, 0, 0, loc), lon, lat, height)

			if info.Type != LunarEclipsePartial {
				t.Fatalf("unexpected eclipse type: got %s want %s", info.Type, LunarEclipsePartial)
			}
			if info.Longitude != lon || info.Latitude != lat || info.Height != height {
				t.Fatalf("observer metadata mismatch: got (%f,%f,%f)", info.Longitude, info.Latitude, info.Height)
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
		})
	}
}

func TestLocalLunarEclipseChauvenetRemainsAvailable(t *testing.T) {
	date := time.Date(2025, 3, 14, 12, 0, 0, 0, time.FixedZone("CDT", -5*3600))
	lon, lat, height := -87.65, 41.85, 0.0
	defaultInfo := ClosestLocalLunarEclipse(date, lon, lat, height)
	chauvenetInfo := ClosestLocalLunarEclipseChauvenet(date, lon, lat, height)

	assertFloatClose(t, "Chauvenet.PenumbralMagnitude", chauvenetInfo.PenumbralMagnitude, 2.285436461, 1e-6)
	assertFloatClose(t, "Chauvenet.UmbralMagnitude", chauvenetInfo.UmbralMagnitude, 1.182806541, 1e-6)

	if !(chauvenetInfo.PenumbralMagnitude > defaultInfo.PenumbralMagnitude) {
		t.Fatalf("expected Chauvenet penumbral magnitude > Danjon: chauvenet=%.6f danjon=%.6f", chauvenetInfo.PenumbralMagnitude, defaultInfo.PenumbralMagnitude)
	}
	if !(chauvenetInfo.PenumbralStart.Before(defaultInfo.PenumbralStart) && chauvenetInfo.PenumbralEnd.After(defaultInfo.PenumbralEnd)) {
		t.Fatalf("expected Chauvenet penumbral span to be wider: chauvenet=(%v,%v) danjon=(%v,%v)", chauvenetInfo.PenumbralStart, chauvenetInfo.PenumbralEnd, defaultInfo.PenumbralStart, defaultInfo.PenumbralEnd)
	}
}

func TestLocalLunarEclipseAgainstNASABaseline(t *testing.T) {
	testCases := []struct {
		name               string
		date               time.Time
		lon                float64
		lat                float64
		height             float64
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
		wantMoonAltitude   float64
	}{
		{
			name:               "2023-10-29 tokyo partial",
			date:               time.Date(2023, 10, 29, 12, 0, 0, 0, time.FixedZone("JST", 9*3600)),
			lon:                139.6917,
			lat:                35.6895,
			height:             0,
			wantType:           LunarEclipsePartial,
			wantPenumbralMag:   1.1181,
			wantUmbralMag:      0.122,
			wantPenumbralStart: time.Date(2023, 10, 29, 03, 01, 43, 0, time.FixedZone("JST", 9*3600)),
			wantPartialStart:   time.Date(2023, 10, 29, 04, 35, 18, 0, time.FixedZone("JST", 9*3600)),
			wantMaximum:        time.Date(2023, 10, 29, 05, 14, 06, 0, time.FixedZone("JST", 9*3600)),
			wantPartialEnd:     time.Date(2023, 10, 29, 05, 52, 53, 0, time.FixedZone("JST", 9*3600)),
			wantPenumbralEnd:   time.Date(2023, 10, 29, 07, 26, 19, 0, time.FixedZone("JST", 9*3600)),
			wantMoonAltitude:   9.1,
		},
		{
			name:               "2025-03-14 chicago total",
			date:               time.Date(2025, 3, 14, 12, 0, 0, 0, time.FixedZone("CDT", -5*3600)),
			lon:                -87.65,
			lat:                41.85,
			height:             0,
			wantType:           LunarEclipseTotal,
			wantPenumbralMag:   2.2595,
			wantUmbralMag:      1.1784,
			wantPenumbralStart: time.Date(2025, 3, 13, 22, 57, 28, 0, time.FixedZone("CDT", -5*3600)),
			wantPartialStart:   time.Date(2025, 3, 14, 0, 9, 40, 0, time.FixedZone("CDT", -5*3600)),
			wantTotalStart:     time.Date(2025, 3, 14, 1, 26, 6, 0, time.FixedZone("CDT", -5*3600)),
			wantMaximum:        time.Date(2025, 3, 14, 1, 58, 41, 0, time.FixedZone("CDT", -5*3600)),
			wantTotalEnd:       time.Date(2025, 3, 14, 2, 31, 26, 0, time.FixedZone("CDT", -5*3600)),
			wantPartialEnd:     time.Date(2025, 3, 14, 3, 47, 56, 0, time.FixedZone("CDT", -5*3600)),
			wantPenumbralEnd:   time.Date(2025, 3, 14, 5, 0, 9, 0, time.FixedZone("CDT", -5*3600)),
			wantMoonAltitude:   48.2,
		},
	}

	const timeTolerance = 2 * time.Minute
	const umbralMagnitudeTolerance = 0.02
	const penumbralMagnitudeTolerance = 0.1
	const altitudeTolerance = 1.5

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info := ClosestLocalLunarEclipse(tc.date, tc.lon, tc.lat, tc.height)
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
			assertFloatClose(t, "MoonAltitude", info.MoonAltitude, tc.wantMoonAltitude, altitudeTolerance)
		})
	}
}

func TestLocalPenumbralLunarEclipseKeepsNegativeUmbralMagnitude(t *testing.T) {
	cdt := time.FixedZone("CDT", -5*3600)
	info := ClosestLocalLunarEclipse(time.Date(2024, 3, 25, 2, 0, 0, 0, cdt), -87.65, 41.85, 0)
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

// TestLocalLunarEclipseVisibilityClass 固定站点自己的食内可见性八分类；rise-and-set 与 interrupted
// 是静态地平圈几何推不出来的，penumbra-moonrise / penumbra-moonset 是本影整段在地平线下的细分。
func TestLocalLunarEclipseVisibilityClass(t *testing.T) {
	date := time.Date(2029, 1, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	previousDay := time.Date(2028, 12, 31, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	testCases := []struct {
		name string
		date time.Time
		lon  float64
		lat  float64
		want LocalLunarEclipseVisibility
	}{
		{name: "mid latitude stays up", lon: 107.25, lat: 30.25, want: LocalLunarEclipseFull},
		{name: "moonset inside eclipse", lon: -161.25, lat: 3.25, want: LocalLunarEclipseMoonset},
		{name: "moonrise inside eclipse", lon: 20.25, lat: -2.75, want: LocalLunarEclipseMoonrise},
		{name: "south polar lens", lon: 108.75, lat: -62.75, want: LocalLunarEclipseRiseAndSet},
		{name: "reported station", lon: 108.729001, lat: -59.937452, want: LocalLunarEclipseRiseAndSet},
		{name: "arctic midday dip", lon: -71.25, lat: 64.75, want: LocalLunarEclipseInterrupted},
		{name: "far side never up", lon: -178.5, lat: -86.5, want: LocalLunarEclipseInvisible},
		// 只看到半影：本影阶段整段在地平线下，月出/月落仍落在食内。
		{name: "penumbra only moonrise", lon: -69.25, lat: 61.25, want: LocalLunarEclipsePenumbraMoonrise},
		{name: "penumbra only moonset", date: previousDay, lon: -179.25, lat: -61.25, want: LocalLunarEclipsePenumbraMoonset},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			at := testCase.date
			if at.IsZero() {
				at = date
			}
			geometric, geometricOK := GeometricLocalLunarEclipseOnDate(at, testCase.lon, testCase.lat, 0)
			if !geometricOK {
				t.Fatal("expected a geometric lunar eclipse on the local date")
			}
			if geometric.Visibility != testCase.want {
				t.Fatalf("visibility=%s want %s", geometric.Visibility, testCase.want)
			}
			_, visibleOK := LocalLunarEclipseOnDate(at, testCase.lon, testCase.lat, 0)
			if wantVisible := testCase.want != LocalLunarEclipseInvisible; visibleOK != wantVisible {
				t.Fatalf("visible filter ok=%v want %v", visibleOK, wantVisible)
			}
		})
	}
}

func TestLocalLunarEclipsePolarWindowScansBetweenContacts(t *testing.T) {
	date := time.Date(1800, 4, 9, 0, 0, 0, 0, time.UTC)
	info, ok := GeometricLocalLunarEclipseOnDate(date, 121, 82, 0)
	if !ok {
		t.Fatal("expected a geometric lunar eclipse")
	}
	if info.Visibility != LocalLunarEclipseRiseAndSet {
		t.Fatalf("visibility=%s, want %s", info.Visibility, LocalLunarEclipseRiseAndSet)
	}
	if _, ok := LocalLunarEclipseOnDate(date, 121, 82, 0); !ok {
		t.Fatal("the short above-horizon interval must make the eclipse visible")
	}
}

func assertSameLocalLunarEclipse(t *testing.T, name string, got, want LocalLunarEclipseInfo, tolerance time.Duration) {
	t.Helper()
	if got.Type != want.Type {
		t.Fatalf("%s type mismatch: got %s want %s", name, got.Type, want.Type)
	}
	assertTimeClose(t, name+".Maximum", got.Maximum, want.Maximum, tolerance)
}

// 纯半影月食没有本影接触，不把任何站点降级成 penumbra-moonrise / penumbra-moonset。
func TestLocalLunarEclipsePenumbralOnlyKeepsOldClasses(t *testing.T) {
	date := time.Date(2020, 1, 11, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
	info, ok := LunarEclipseOnDate(date)
	if !ok || info.HasPartial {
		t.Fatalf("expected a purely penumbral eclipse, HasPartial=%v ok=%v", info.HasPartial, ok)
	}
	seen := map[LocalLunarEclipseVisibility]int{}
	for longitude := -179.5; longitude < 180; longitude += 5 {
		for latitude := -89.5; latitude < 90; latitude += 5 {
			local, found := GeometricLocalLunarEclipseOnDate(date, longitude, latitude, 0)
			if !found {
				continue
			}
			seen[local.Visibility]++
		}
	}
	if seen[LocalLunarEclipsePenumbraMoonrise] != 0 || seen[LocalLunarEclipsePenumbraMoonset] != 0 {
		t.Fatalf("purely penumbral eclipse produced penumbra-only classes: %v", seen)
	}
	if seen[LocalLunarEclipseMoonrise] == 0 || seen[LocalLunarEclipseMoonset] == 0 {
		t.Fatalf("purely penumbral eclipse lost its moonrise/moonset sites: %v", seen)
	}
}

// 极区掠射：U1、食甚、U4 三点都可能在地平下，但本影区间中途仍有短暂窗口高于地平；这类站点确实
// 见到本影，必须留在 moonrise/moonset，只有本影区间极值也在地平下才算"仅见半影"。
func TestLocalLunarEclipseKeepsUmbralGrazingWindow(t *testing.T) {
	date := time.Date(1932, 3, 22, 0, 0, 0, 0, time.UTC)
	for _, witness := range []struct {
		name     string
		lon, lat float64
		want     LocalLunarEclipseVisibility
	}{
		{"本影窗口 +0.0006 度", -91.8, -88.7675, LocalLunarEclipseMoonrise},
		{"U1 已在地平上", -91.75, -88.75, LocalLunarEclipseMoonrise},
		{"U1 高度 +0.21 度（陡边界）", -75.5, 6.5, LocalLunarEclipseMoonset},
		{"本影区间极值 -0.005 度", -90.0, -88.65, LocalLunarEclipsePenumbraMoonset},
	} {
		local, ok := GeometricLocalLunarEclipseOnDate(date, witness.lon, witness.lat, 0)
		if !ok {
			t.Fatalf("%s: 本影几何月食缺失", witness.name)
		}
		if local.Visibility != witness.want {
			t.Fatalf("%s (%.4f,%.4f) visibility=%s want %s",
				witness.name, witness.lon, witness.lat, local.Visibility, witness.want)
		}
	}
}
