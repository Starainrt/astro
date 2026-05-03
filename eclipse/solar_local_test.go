package eclipse

import (
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestLocalSolarEclipseOnDateByLocalDay(t *testing.T) {
	loc := time.FixedZone("CDT", -5*3600)
	lon, lat, height := -87.65, 41.85, 0.0

	testCases := []struct {
		name string
		date time.Time
		want bool
	}{
		{
			name: "day before no eclipse",
			date: time.Date(2024, 4, 7, 12, 0, 0, 0, loc),
			want: false,
		},
		{
			name: "local event day overlaps",
			date: time.Date(2024, 4, 8, 12, 0, 0, 0, loc),
			want: true,
		},
		{
			name: "day after no eclipse",
			date: time.Date(2024, 4, 9, 12, 0, 0, 0, loc),
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info, ok := LocalSolarEclipseOnDate(tc.date, lon, lat, height)
			if ok != tc.want {
				t.Fatalf("LocalSolarEclipseOnDate(%v) got %v want %v", tc.date, ok, tc.want)
			}
			if !ok {
				return
			}
			if info.Type != SolarEclipsePartial {
				t.Fatalf("unexpected eclipse type: got %s want %s", info.Type, SolarEclipsePartial)
			}
			if info.GreatestEclipse.Location() != loc {
				t.Fatalf("greatest eclipse location mismatch: got %q want %q", info.GreatestEclipse.Location(), loc)
			}
			if info.PartialStart.Day() != 8 || info.PartialEnd.Day() != 8 {
				t.Fatalf("unexpected local date span: begin=%v end=%v", info.PartialStart, info.PartialEnd)
			}
		})
	}
}

func TestLocalSolarEclipseVisibilityFilter(t *testing.T) {
	loc := time.FixedZone("JST", 9*3600)
	date := time.Date(2024, 4, 9, 12, 0, 0, 0, loc)
	lon, lat, height := 139.6917, 35.6895, 0.0

	geometricInfo, geometricOK := GeometricLocalSolarEclipseOnDate(date, lon, lat, height)
	if !geometricOK {
		t.Fatalf("expected geometric local eclipse on date")
	}
	if geometricInfo.Type != SolarEclipsePartial {
		t.Fatalf("unexpected geometric eclipse type: got %s want %s", geometricInfo.Type, SolarEclipsePartial)
	}
	if geometricInfo.VisibleAtGreatest {
		t.Fatalf("expected geometric eclipse to be below horizon at greatest: %+v", geometricInfo)
	}

	visibleInfo, visibleOK := LocalSolarEclipseOnDate(date, lon, lat, height)
	if visibleOK {
		t.Fatalf("expected visible filter to reject invisible eclipse, got %+v", visibleInfo)
	}
}

func TestLocalSolarEclipseSearchSemantics(t *testing.T) {
	loc := time.UTC
	lon, lat, height := -0.1278, 51.5074, 0.0
	current := ClosestLocalSolarEclipseNASABulletinSplitK(time.Date(2025, 3, 29, 12, 0, 0, 0, loc), lon, lat, height)
	if current.Type != SolarEclipsePartial {
		t.Fatalf("unexpected current eclipse type: got %s want %s", current.Type, SolarEclipsePartial)
	}

	assertSameLocalSolarEclipse(
		t,
		"ClosestLocalSolarEclipse(default)",
		ClosestLocalSolarEclipse(current.GreatestEclipse, lon, lat, height),
		current,
		time.Second,
	)

	last := LastLocalSolarEclipseNASABulletinSplitK(current.GreatestEclipse, lon, lat, height)
	assertSameLocalSolarEclipse(t, "LastLocalSolarEclipseNASABulletinSplitK(current.GreatestEclipse)", last, current, time.Second)

	closest := ClosestLocalSolarEclipseNASABulletinSplitK(current.GreatestEclipse, lon, lat, height)
	assertSameLocalSolarEclipse(t, "ClosestLocalSolarEclipseNASABulletinSplitK(current.GreatestEclipse)", closest, current, time.Second)

	next := NextLocalSolarEclipseNASABulletinSplitK(current.GreatestEclipse, lon, lat, height)
	if !next.GreatestEclipse.After(current.GreatestEclipse) {
		t.Fatalf("NextLocalSolarEclipseNASABulletinSplitK should be strictly future: current=%v next=%v", current.GreatestEclipse, next.GreatestEclipse)
	}
	if next.Type != SolarEclipsePartial {
		t.Fatalf("unexpected next eclipse type: got %s want %s", next.Type, SolarEclipsePartial)
	}
	assertSolarTimeClose(t, "NextLocalSolarEclipseNASABulletinSplitK", next.GreatestEclipse, time.Date(2026, 8, 12, 18, 13, 21, 0, time.UTC), 2*time.Minute)
}

func TestLocalSolarEclipseSearchSkipsInvisibleCurrentCandidate(t *testing.T) {
	loc := time.FixedZone("JST", 9*3600)
	lon, lat, height := 139.6917, 35.6895, 0.0
	current, ok := GeometricLocalSolarEclipseOnDate(time.Date(2024, 4, 9, 12, 0, 0, 0, loc), lon, lat, height)
	if !ok {
		t.Fatalf("expected geometric local eclipse on date")
	}
	if current.VisibleAtGreatest {
		t.Fatalf("expected current geometric eclipse to be below horizon: %+v", current)
	}

	next := NextLocalSolarEclipseNASABulletinSplitK(current.GreatestEclipse, lon, lat, height)
	if next.Type == SolarEclipseNone || next.GreatestEclipse.IsZero() {
		t.Fatalf("expected search to skip the invisible current candidate and find a future visible eclipse")
	}
	if !next.GreatestEclipse.After(current.GreatestEclipse) {
		t.Fatalf("expected strictly future local solar eclipse: current=%v next=%v", current.GreatestEclipse, next.GreatestEclipse)
	}
}

func TestLocalTotalSolarEclipseSearch(t *testing.T) {
	loc := time.UTC
	lon, lat, height := -104.1, 25.3, 0.0
	date := time.Date(2024, 4, 7, 0, 0, 0, 0, loc)

	next, ok := NextLocalTotalSolarEclipse(date, lon, lat, height)
	if !ok {
		t.Fatal("expected to find next local total solar eclipse")
	}
	if next.Type != SolarEclipseTotal || !next.HasTotal {
		t.Fatalf("unexpected next total eclipse: %+v", next)
	}
	assertSolarTimeClose(t, "NextLocalTotalSolarEclipse", next.GreatestEclipse, time.Date(2024, 4, 8, 18, 17, 15, 0, loc), time.Minute)
	assertSolarDurationClose(t, "NextLocalTotalSolarEclipse duration", next.CentralEnd.Sub(next.CentralStart), 4*time.Minute+28*time.Second, 5*time.Second)

	last, ok := LastLocalTotalSolarEclipse(next.GreatestEclipse, lon, lat, height)
	if !ok {
		t.Fatal("expected to find previous local total solar eclipse")
	}
	if last.Type != SolarEclipseTotal || !last.HasTotal {
		t.Fatalf("unexpected last total eclipse: %+v", last)
	}
	assertSolarTimeClose(t, "LastLocalTotalSolarEclipse", last.GreatestEclipse, next.GreatestEclipse, time.Second)
	assertSolarDurationClose(t, "LastLocalTotalSolarEclipse duration", last.CentralEnd.Sub(last.CentralStart), 4*time.Minute+28*time.Second, 5*time.Second)
}

func TestLocalTotalSolarEclipseClosest(t *testing.T) {
	loc := time.UTC
	lon, lat, height := -104.1, 25.3, 0.0
	date := time.Date(2024, 4, 8, 12, 0, 0, 0, loc)

	info, ok := ClosestLocalTotalSolarEclipse(date, lon, lat, height)
	if !ok {
		t.Fatal("expected to find closest local total solar eclipse")
	}
	if info.Type != SolarEclipseTotal || !info.HasTotal {
		t.Fatalf("unexpected closest total eclipse: %+v", info)
	}
	assertSolarTimeClose(t, "ClosestLocalTotalSolarEclipse", info.GreatestEclipse, time.Date(2024, 4, 8, 18, 17, 15, 0, loc), time.Minute)
}

func TestLocalAnnularSolarEclipseSearch(t *testing.T) {
	loc := time.UTC
	lon, lat, height := -114.5, -22.0, 0.0
	date := time.Date(2024, 10, 1, 0, 0, 0, 0, loc)

	next, ok := NextLocalAnnularSolarEclipse(date, lon, lat, height)
	if !ok {
		t.Fatal("expected to find next local annular solar eclipse")
	}
	if next.Type != SolarEclipseAnnular || !next.HasAnnular || next.HasTotal {
		t.Fatalf("unexpected next annular eclipse: %+v", next)
	}
	assertSolarTimeClose(t, "NextLocalAnnularSolarEclipse", next.GreatestEclipse, time.Date(2024, 10, 2, 18, 44, 59, 0, loc), time.Minute)
	assertSolarDurationClose(t, "NextLocalAnnularSolarEclipse duration", next.CentralEnd.Sub(next.CentralStart), 7*time.Minute+25*time.Second, 5*time.Second)

	last, ok := LastLocalAnnularSolarEclipse(next.GreatestEclipse, lon, lat, height)
	if !ok {
		t.Fatal("expected to find previous local annular solar eclipse")
	}
	if last.Type != SolarEclipseAnnular || !last.HasAnnular || last.HasTotal {
		t.Fatalf("unexpected last annular eclipse: %+v", last)
	}
	assertSolarTimeClose(t, "LastLocalAnnularSolarEclipse", last.GreatestEclipse, next.GreatestEclipse, time.Second)
}

func TestLocalAnnularSolarEclipseClosest(t *testing.T) {
	loc := time.UTC
	lon, lat, height := -114.5, -22.0, 0.0
	date := time.Date(2024, 10, 2, 12, 0, 0, 0, loc)

	info, ok := ClosestLocalAnnularSolarEclipse(date, lon, lat, height)
	if !ok {
		t.Fatal("expected to find closest local annular solar eclipse")
	}
	if info.Type != SolarEclipseAnnular || !info.HasAnnular || info.HasTotal {
		t.Fatalf("unexpected closest annular eclipse: %+v", info)
	}
	assertSolarTimeClose(t, "ClosestLocalAnnularSolarEclipse", info.GreatestEclipse, time.Date(2024, 10, 2, 18, 44, 59, 0, loc), time.Minute)
}

func TestLocalCentralSolarEclipseVisibleRequiresCentralPhaseVisibility(t *testing.T) {
	info := LocalSolarEclipseInfo{
		Type:              SolarEclipseTotal,
		Longitude:         0,
		Latitude:          0,
		PartialStart:      time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		PartialEnd:        time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
		CentralStart:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		CentralEnd:        time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
		HasPartial:        true,
		HasCentral:        true,
		HasTotal:          true,
		VisibleAtGreatest: false,
	}
	if !localSolarEclipseVisible(info) {
		t.Fatalf("expected partial phase to be visible")
	}
	if localCentralSolarEclipseVisible(info) {
		t.Fatalf("expected central phase below horizon to be rejected")
	}

	info.Type = SolarEclipseAnnular
	info.HasTotal = false
	info.HasAnnular = true
	if localCentralSolarEclipseVisible(info) {
		t.Fatalf("expected annular central phase below horizon to be rejected")
	}
}

func TestLocalSolarEclipseInfoKeepsLocation(t *testing.T) {
	loc := time.FixedZone("UTC+08", 8*3600)
	lon, lat, height := -104.1, 25.3, 1234.0
	testCases := []struct {
		name string
		calc func(time.Time, float64, float64, float64) LocalSolarEclipseInfo
	}{
		{name: "nasa", calc: ClosestLocalSolarEclipseNASABulletinSplitK},
		{name: "iau", calc: ClosestLocalSolarEclipseIAUSingleK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info := tc.calc(time.Date(2024, 4, 8, 12, 0, 0, 0, loc), lon, lat, height)
			if info.Type != SolarEclipseTotal {
				t.Fatalf("unexpected eclipse type: got %s want %s", info.Type, SolarEclipseTotal)
			}
			if info.Longitude != lon || info.Latitude != lat || info.Height != height {
				t.Fatalf("observer metadata mismatch: got (%f,%f,%f)", info.Longitude, info.Latitude, info.Height)
			}

			for _, item := range []struct {
				name string
				tm   time.Time
			}{
				{name: "GreatestEclipse", tm: info.GreatestEclipse},
				{name: "PartialStart", tm: info.PartialStart},
				{name: "PartialEnd", tm: info.PartialEnd},
				{name: "CentralStart", tm: info.CentralStart},
				{name: "CentralEnd", tm: info.CentralEnd},
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

func TestLocalSolarEclipseContactPoints(t *testing.T) {
	loc := time.FixedZone("UTC+08", 8*3600)
	info := ClosestLocalSolarEclipse(
		time.Date(2024, 4, 8, 12, 0, 0, 0, loc),
		-96.7970,
		32.7767,
		0,
	)
	if info.Type != SolarEclipseTotal {
		t.Fatalf("unexpected eclipse type: got %s want %s", info.Type, SolarEclipseTotal)
	}
	if got, want := len(info.ContactPoints), 4; got != want {
		t.Fatalf("contact point count = %d, want %d", got, want)
	}

	points := make(map[string]LocalSolarEclipseContactPoint, len(info.ContactPoints))
	for _, point := range info.ContactPoints {
		points[point.Label] = point
	}
	assertSolarFloatClose(t, "C1.ContactPositionAngle", points["C1"].ContactPositionAngle, 226.219228, 1e-3)
	assertSolarFloatClose(t, "C2.ContactPositionAngle", points["C2"].ContactPositionAngle, 19.137089, 1e-3)
	assertSolarFloatClose(t, "C2.MoonCenterPositionAngle", points["C2"].MoonCenterPositionAngle, 199.137089, 1e-3)
	assertSolarFloatClose(t, "C4.ContactClockwiseAngle", points["C4"].ContactClockwiseAngle, 310.781438, 1e-3)
}

func TestLocalSolarEclipseContactPointsFromMergedFrameLabels(t *testing.T) {
	frames := []basic.LocalSolarEclipseDiagramFrame{
		{
			JDE:           2460409.25,
			SunRadius:     1,
			MoonRadius:    1.05,
			PositionAngle: 199.137089,
			Label:         "Greatest",
			Labels:        []string{"C2", "Greatest", "C3"},
		},
	}
	points := localSolarEclipseContactPointsFromFrames(frames, time.UTC)
	if got, want := len(points), 2; got != want {
		t.Fatalf("contact point count = %d, want %d", got, want)
	}
	if points[0].Label != "C2" || points[1].Label != "C3" {
		t.Fatalf("labels = %#v, want [C2 C3]", []string{points[0].Label, points[1].Label})
	}
	assertSolarFloatClose(t, "C2.ContactPositionAngle", points[0].ContactPositionAngle, 19.137089, 1e-6)
	assertSolarFloatClose(t, "C3.ContactPositionAngle", points[1].ContactPositionAngle, 19.137089, 1e-6)
}

func TestLocalSolarEclipseIAUSingleKRemainsAvailable(t *testing.T) {
	date := time.Date(2024, 4, 8, 12, 0, 0, 0, time.UTC)
	lon, lat, height := -104.1, 25.3, 0.0
	defaultInfo := ClosestLocalSolarEclipse(date, lon, lat, height)
	iauInfo := ClosestLocalSolarEclipseIAUSingleK(date, lon, lat, height)

	if defaultInfo.Type != SolarEclipseTotal || iauInfo.Type != SolarEclipseTotal {
		t.Fatalf("unexpected eclipse types: default=%s iau=%s", defaultInfo.Type, iauInfo.Type)
	}
	assertSolarTimeClose(t, "GreatestEclipse", iauInfo.GreatestEclipse, defaultInfo.GreatestEclipse, time.Second)

	if !(iauInfo.CentralEnd.Sub(iauInfo.CentralStart) > defaultInfo.CentralEnd.Sub(defaultInfo.CentralStart)) {
		t.Fatalf("expected IAU central duration > NASA duration: iau=%v nasa=%v", iauInfo.CentralEnd.Sub(iauInfo.CentralStart), defaultInfo.CentralEnd.Sub(defaultInfo.CentralStart))
	}
	if !(iauInfo.Magnitude > defaultInfo.Magnitude) {
		t.Fatalf("expected IAU magnitude > NASA magnitude: iau=%.9f nasa=%.9f", iauInfo.Magnitude, defaultInfo.Magnitude)
	}
}

func TestLocalSolarEclipseAgainstNASABaseline(t *testing.T) {
	chicagoLoc := time.FixedZone("CDT", -5*3600)

	testCases := []struct {
		name                string
		date                time.Time
		lon                 float64
		lat                 float64
		height              float64
		wantType            SolarEclipseType
		wantGreatest        time.Time
		wantPartialStart    time.Time
		wantPartialEnd      time.Time
		wantMagnitude       float64
		wantObscuration     float64
		wantSunAltitude     float64
		wantSunAzimuth      float64
		wantCentralDuration time.Duration
	}{
		{
			name:             "2024-04-08 chicago partial",
			date:             time.Date(2024, 4, 8, 12, 0, 0, 0, chicagoLoc),
			lon:              -87.65,
			lat:              41.85,
			height:           0,
			wantType:         SolarEclipsePartial,
			wantGreatest:     time.Date(2024, 4, 8, 14, 7, 0, 0, chicagoLoc),
			wantPartialStart: time.Date(2024, 4, 8, 12, 51, 0, 0, chicagoLoc),
			wantPartialEnd:   time.Date(2024, 4, 8, 15, 22, 0, 0, chicagoLoc),
			wantMagnitude:    0.942,
			wantObscuration:  0.938,
			wantSunAltitude:  52,
			wantSunAzimuth:   211,
		},
		{
			name:                "2024-04-08 greatest total",
			date:                time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC),
			lon:                 -104.1,
			lat:                 25.3,
			height:              0,
			wantType:            SolarEclipseTotal,
			wantGreatest:        time.Date(2024, 4, 8, 18, 17, 15, 0, time.UTC),
			wantPartialStart:    time.Date(2024, 4, 8, 16, 58, 0, 0, time.UTC),
			wantPartialEnd:      time.Date(2024, 4, 8, 19, 40, 0, 0, time.UTC),
			wantMagnitude:       1.0566,
			wantObscuration:     1.0,
			wantSunAltitude:     70,
			wantSunAzimuth:      150,
			wantCentralDuration: 4*time.Minute + 28*time.Second,
		},
		{
			name:                "2024-10-02 greatest annular",
			date:                time.Date(2024, 10, 2, 0, 0, 0, 0, time.UTC),
			lon:                 -114.5,
			lat:                 -22.0,
			height:              0,
			wantType:            SolarEclipseAnnular,
			wantGreatest:        time.Date(2024, 10, 2, 18, 44, 59, 0, time.UTC),
			wantPartialStart:    time.Date(2024, 10, 2, 17, 3, 0, 0, time.UTC),
			wantPartialEnd:      time.Date(2024, 10, 2, 20, 33, 0, 0, time.UTC),
			wantMagnitude:       0.9326,
			wantObscuration:     0.871,
			wantSunAltitude:     69,
			wantSunAzimuth:      31,
			wantCentralDuration: 7*time.Minute + 25*time.Second,
		},
	}

	const (
		partialTimeTolerance  = 90 * time.Second
		greatestTimeTolerance = 45 * time.Second
		floatTolerance        = 0.01
		altitudeTolerance     = 1.0
		azimuthTolerance      = 1.5
		durationTolerance     = 5 * time.Second
	)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info := ClosestLocalSolarEclipse(tc.date, tc.lon, tc.lat, tc.height)
			if info.Type != tc.wantType {
				t.Fatalf("type mismatch: got %s want %s", info.Type, tc.wantType)
			}

			assertSolarTimeClose(t, "GreatestEclipse", info.GreatestEclipse, tc.wantGreatest, greatestTimeTolerance)
			assertSolarTimeClose(t, "PartialStart", info.PartialStart, tc.wantPartialStart, partialTimeTolerance)
			assertSolarTimeClose(t, "PartialEnd", info.PartialEnd, tc.wantPartialEnd, partialTimeTolerance)
			assertSolarFloatClose(t, "Magnitude", info.Magnitude, tc.wantMagnitude, floatTolerance)
			assertSolarFloatClose(t, "Obscuration", info.Obscuration, tc.wantObscuration, floatTolerance)
			assertSolarFloatClose(t, "SunAltitude", info.SunAltitude, tc.wantSunAltitude, altitudeTolerance)
			assertSolarFloatClose(t, "SunAzimuth", info.SunAzimuth, tc.wantSunAzimuth, azimuthTolerance)

			if tc.wantCentralDuration > 0 {
				duration := info.CentralEnd.Sub(info.CentralStart)
				assertSolarTimeClose(
					t,
					"CentralDuration",
					time.Unix(0, int64(duration)),
					time.Unix(0, int64(tc.wantCentralDuration)),
					durationTolerance,
				)
			} else if info.HasCentral || !info.CentralStart.IsZero() || !info.CentralEnd.IsZero() {
				t.Fatalf("expected no central phase, got %+v", info)
			}
		})
	}
}

func assertSameLocalSolarEclipse(t *testing.T, name string, got, want LocalSolarEclipseInfo, tolerance time.Duration) {
	t.Helper()
	if got.Type != want.Type {
		t.Fatalf("%s type mismatch: got %s want %s", name, got.Type, want.Type)
	}
	assertSolarTimeClose(t, name+".GreatestEclipse", got.GreatestEclipse, want.GreatestEclipse, tolerance)
}

func assertSolarDurationClose(t *testing.T, name string, got, want, tolerance time.Duration) {
	t.Helper()
	diff := got - want
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Fatalf("%s mismatch: got %v want %v diff=%v", name, got, want, diff)
	}
}
