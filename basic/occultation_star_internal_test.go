package basic

import (
	"math"
	"testing"
	"time"
)

func TestStarOccultationLatitudeEnvelopePrefilter(t *testing.T) {
	start := occultationTimeToTT(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	end := start + starOccultationSiderealMonthDays
	polar := StarCoordinate{RA: 0, Dec: 89, Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameICRS}
	if starOccultationLatitudeEnvelopePass(start, end, polar, nil, 0) {
		t.Fatal("polar star should be rejected by global ecliptic-latitude envelope")
	}
	if !starOccultationLatitudeEnvelopePass(start, end, hr4799OccultationCoordinateForTest(), nil, 0) {
		t.Fatal("HR 4799 should pass the global ecliptic-latitude envelope")
	}
}

func TestStarOccultationScanCandidatesFindsMultipleLocalMinima(t *testing.T) {
	value := func(tt float64) float64 {
		return (tt-1)*(tt-1)*(tt-3)*(tt-3) + 0.001
	}
	got := starOccultationScanCandidates(0, 4, 0.5, value, func(float64) bool { return true })
	if len(got) != 2 {
		t.Fatalf("candidate count = %d, want 2: %v", len(got), got)
	}
	if math.Abs(got[0]-1) > 1e-6 || math.Abs(got[1]-3) > 1e-6 {
		t.Fatalf("candidate minima = %v, want [1 3]", got)
	}
}

func TestStarOccultationScanCandidatesFindsNarrowWindowMidpoint(t *testing.T) {
	const (
		start = 10.0
		end   = 10.1
		want  = (start + end) / 2
	)
	value := func(tt float64) float64 { return (tt - want) * (tt - want) }
	got := starOccultationScanCandidates(start, end, starOccultationDefaultStepDays, value, func(float64) bool { return true })
	if len(got) != 1 {
		t.Fatalf("candidate count = %d, want 1: %v", len(got), got)
	}
	if math.Abs(got[0]-want) > 1e-7 {
		t.Fatalf("candidate midpoint = %.12f, want %.12f", got[0], want)
	}
}

func TestUniqueOccultationCandidateTimesFiltersSingleOutsideCandidate(t *testing.T) {
	if got := uniqueOccultationCandidateTimes([]float64{9.9}, 10, 11); len(got) != 0 {
		t.Fatalf("outside candidate was not filtered: %v", got)
	}
	got := uniqueOccultationCandidateTimes([]float64{10.5}, 10, 11)
	if len(got) != 1 || got[0] != 10.5 {
		t.Fatalf("inside candidate = %v, want [10.5]", got)
	}
}

func TestStarOccultationPropagatesApparentCoordinateProperMotion(t *testing.T) {
	epoch := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	target := epoch.Add(365*24*time.Hour + 6*time.Hour)
	star := StarCoordinate{
		RA:                             10,
		Dec:                            20,
		Epoch:                          epoch,
		Frame:                          CoordinateFrameApparentOfDate,
		ProperMotionRACosDecMasPerYear: 360000,
		ProperMotionDecMasPerYear:      -720000,
	}
	baseline := star
	baseline.ProperMotionRACosDecMasPerYear = 0
	baseline.ProperMotionDecMasPerYear = 0
	baselineRA, baselineDec := starApparentRaDec(occultationTimeToTT(target), baseline, Observer{})
	ra, dec := starApparentRaDec(occultationTimeToTT(target), star, Observer{})
	raMotion := signedAngleDifference(ra, baselineRA) * math.Cos(baselineDec*math.Pi/180)
	decMotion := dec - baselineDec
	if math.Abs(raMotion-0.1) > 0.001 {
		t.Fatalf("propagated RA*cos(Dec) motion = %.10f deg, want 0.1", raMotion)
	}
	if math.Abs(decMotion-(-0.2)) > 0.001 {
		t.Fatalf("propagated Dec motion = %.10f deg, want -0.2", decMotion)
	}
}

func TestStarOccultationApparentOfDateRoundTripsAtEpoch(t *testing.T) {
	epoch := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	star := StarCoordinate{RA: 189.5, Dec: -6.2, Epoch: epoch, Frame: CoordinateFrameApparentOfDate, ParallaxMas: 100}
	ra, dec := starApparentRaDecGeocentric(occultationTimeToTT(epoch), star)
	if math.Abs(signedAngleDifference(ra, star.RA))*3600 > 1e-5 || math.Abs(dec-star.Dec)*3600 > 1e-5 {
		t.Fatalf("apparent coordinate did not round-trip at epoch: got %.12f %.12f", ra, dec)
	}
}

func TestStarOccultationApparentPlaceCorrections(t *testing.T) {
	location := time.FixedZone("CST", 8*3600)
	tt := occultationTimeToTT(time.Date(2025, 6, 5, 20, 2, 7, 700000000, location))
	star := hr4799OccultationCoordinateForTest()

	gotRA, gotDec := starApparentRaDecGeocentric(tt, star)
	if math.Abs(signedAngleDifference(gotRA, 189.527817)) > 0.0002 || math.Abs(gotDec-(-5.973401)) > 0.0002 {
		t.Fatalf("apparent place = %.9f %.9f, want near 189.527817 -5.973401", gotRA, gotDec)
	}

	years := (tt - Date2JDE(star.Epoch.UTC())) / 365.25
	meanRA := star.RA + years*star.ProperMotionRACosDecMasPerYear/(3600000*math.Cos(star.Dec*math.Pi/180))
	meanDec := star.Dec + years*star.ProperMotionDecMasPerYear/3600000
	meanRA, meanDec = Precess(meanRA, meanDec, 2451545, tt)
	correction := angularSeparationDegrees(meanRA, meanDec, gotRA, gotDec) * 3600
	if correction < 5 || correction > 30 {
		t.Fatalf("apparent-place correction = %.6f arcsec, want a plausible annual correction", correction)
	}
}

func TestStarOccultationApparentPlaceAppliesAnnualParallax(t *testing.T) {
	star := hr4799OccultationCoordinateForTest()
	tt := occultationTimeToTT(time.Date(2025, 6, 5, 12, 0, 0, 0, time.UTC))
	withoutRA, withoutDec := starApparentRaDecGeocentric(tt, star)
	star.ParallaxMas = 1000
	withRA, withDec := starApparentRaDecGeocentric(tt, star)
	shift := angularSeparationDegrees(withoutRA, withoutDec, withRA, withDec) * 3600
	if shift < 0.05 || shift > 1.1 {
		t.Fatalf("annual parallax shift = %.6f arcsec, want (0.05, 1.1]", shift)
	}
}

func TestStarOccultationICRSAppliesJ2000FrameBias(t *testing.T) {
	epoch := time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC)
	icrs := StarCoordinate{RA: 0, Dec: 0, Epoch: epoch, Frame: CoordinateFrameICRS}
	j2000 := icrs
	j2000.Frame = CoordinateFrameJ2000
	tt := occultationTimeToTT(epoch)
	icrsRA, icrsDec := starApparentRaDecGeocentric(tt, icrs)
	j2000RA, j2000Dec := starApparentRaDecGeocentric(tt, j2000)
	raBiasMas := signedAngleDifference(icrsRA, j2000RA) * 3600000
	decBiasMas := (icrsDec - j2000Dec) * 3600000
	if math.Abs(raBiasMas-14.6) > 0.1 || math.Abs(decBiasMas-(-16.617)) > 0.1 {
		t.Fatalf("ICRS frame bias = %.6f %.6f mas, want about 14.6 -16.617", raBiasMas, decBiasMas)
	}
}

func TestRefinedStarOccultationCenterLineRespectsWidthTolerance(t *testing.T) {
	star := hr4799OccultationCoordinateForTest()
	start := time.Date(2025, time.June, 5, 0, 0, 0, 0, time.UTC)
	paths, err := FindStarOccultationPaths(
		start, start.Add(24*time.Hour), star,
		OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 50},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindStarOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	for index, point := range paths[0].CenterLine {
		exact, ok := starOccultationPathCenterPoint(centerTimeTT(point.Time), star, time.UTC)
		if !ok {
			t.Fatalf("exact center point %d is unavailable", index)
		}
		if difference := math.Abs(point.WidthKM - exact.WidthKM); difference > occultationPathWidthToleranceKM {
			t.Fatalf("center point %d width differs from exact value by %.9f km: got %.9f want %.9f",
				index, difference, point.WidthKM, exact.WidthKM)
		}
	}
}

func TestPolarStarOccultationFootprintsContainCenterLine(t *testing.T) {
	star := StarCoordinate{
		ID: "Antares", RA: 247.3516666666667, Dec: -26.431944444444444,
		Epoch: time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC), Frame: CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -10, ProperMotionDecMasPerYear: -20, ParallaxMas: 24,
	}
	start := time.Date(2026, time.February, 11, 0, 0, 0, 0, time.UTC)
	paths, err := FindStarOccultationPaths(
		start, start.Add(24*time.Hour), star,
		OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 200},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindStarOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	if len(paths[0].Footprints) == 0 {
		t.Fatal("polar stellar path has no instantaneous footprints")
	}
	checked := 0
	for _, center := range paths[0].CenterLine {
		if center.MoonAltitude <= 0 {
			continue
		}
		contained := false
		for _, footprint := range paths[0].Footprints {
			if planetOccultationFootprintContains(footprint, center.Longitude, center.Latitude) {
				contained = true
				break
			}
		}
		if !contained {
			t.Fatalf("footprint sweep does not contain visible center line at %v, %.4f, %.4f",
				center.Time, center.Longitude, center.Latitude)
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("no visible center-line sample was checked")
	}
}

func TestRefineOccultationPathWidthsBoundsSmoothInterpolationError(t *testing.T) {
	start := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	startTT := occultationTimeToTT(start)
	widthAt := func(tt float64) (float64, bool) {
		seconds := (tt - startTT) * 86400
		return 3500 + 0.0002*(seconds-50)*(seconds-50), true
	}
	points := make([]OccultationPathPoint, 101)
	for index := range points {
		points[index].Time = start.Add(time.Duration(index) * time.Second)
	}
	points[0].WidthKM, _ = widthAt(centerTimeTT(points[0].Time))
	points[len(points)-1].WidthKM, _ = widthAt(centerTimeTT(points[len(points)-1].Time))
	refineOccultationPathWidths(points, widthAt)
	for index, point := range points {
		exact, _ := widthAt(centerTimeTT(point.Time))
		if difference := math.Abs(point.WidthKM - exact); difference > occultationPathWidthToleranceKM {
			t.Fatalf("interpolated width %d differs by %.9f km, tolerance %.9f", index, difference, occultationPathWidthToleranceKM)
		}
	}
}

func TestOccultationPathCachedEarthRotationMatchesDirectGeometry(t *testing.T) {
	tt := occultationTimeToTT(time.Date(2025, time.June, 5, 12, 2, 6, 0, time.UTC))
	vector := occultationPathVector{x: 4123.5, y: -2789.25, z: 3950.75}
	angle := ApparentSiderealTime(TD2UT(tt, false)) * 15 * math.Pi / 180
	want := occultationPathVector{
		x: math.Cos(angle)*vector.x + math.Sin(angle)*vector.y,
		y: -math.Sin(angle)*vector.x + math.Cos(angle)*vector.y,
		z: vector.z,
	}
	got := occultationPathEarthFixedVectorWithRotation(vector, occultationPathEarthRotationAt(tt))
	if difference := occultationPathNorm(occultationPathSub(got, want)); difference > 1e-12 {
		t.Fatalf("cached Earth rotation differs by %.15g km", difference)
	}

	_, geodeticLatitude := occultationPathGeodetic(tt, vector)
	if difference := math.Abs(occultationPathGeodeticLatitude(vector) - geodeticLatitude); difference > 1e-12 {
		t.Fatalf("cached geodetic latitude differs by %.15g degrees", difference)
	}
}

func TestOccultationPathCachedMoonAndSiderealMatchDirectPoint(t *testing.T) {
	star := hr4799OccultationCoordinateForTest()
	tt := occultationTimeToTT(time.Date(2025, time.June, 5, 12, 2, 6, 0, time.UTC))
	frame, ok := starOccultationPathFrameAt(tt, star)
	if !ok {
		t.Fatal("stellar occultation frame is unavailable")
	}
	vector, _, ok := occultationEarthLineIntersection(frame.moon, frame.axis)
	if !ok {
		t.Fatal("stellar occultation center point is unavailable")
	}

	direct := occultationPathPointFromVector(tt, vector, 1234.5, time.UTC)
	cached := occultationPathPointFromVectorWithMoonSidereal(
		tt, vector, 1234.5, frame.moon, ApparentSiderealTime(TD2UT(tt, false))*15, time.UTC,
	)
	if !cached.Time.Equal(direct.Time) {
		t.Fatalf("cached point time = %v, want %v", cached.Time, direct.Time)
	}
	for name, difference := range map[string]float64{
		"longitude":     math.Abs(cached.Longitude - direct.Longitude),
		"latitude":      math.Abs(cached.Latitude - direct.Latitude),
		"moon altitude": math.Abs(cached.MoonAltitude - direct.MoonAltitude),
		"width":         math.Abs(cached.WidthKM - direct.WidthKM),
	} {
		if difference > 1e-10 {
			t.Fatalf("cached point %s differs by %.15g degrees or km", name, difference)
		}
	}
}

func hr4799OccultationCoordinateForTest() StarCoordinate {
	return StarCoordinate{
		ID:                             "HR 4799",
		RA:                             189.1975,
		Dec:                            -5.831944444444,
		Epoch:                          time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		Frame:                          CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: -28,
		ProperMotionDecMasPerYear:      -18,
	}
}
