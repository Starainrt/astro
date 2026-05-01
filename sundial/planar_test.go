package sundial

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/sun"
)

func TestPlanarShadowPointMatchesBookExamples(t *testing.T) {
	dial := PlanarDial{
		Latitude:                  40,
		PlaneNormalAzimuth:        250,
		PlaneNormalZenithDistance: 50,
		StylusLength:              1,
	}

	point := dial.ShadowPointByHourAngleDeclination(30, 23.44)
	assertClose(t, "58a x", point.X, -0.0390, 1e-4)
	assertClose(t, "58a y", point.Y, -0.3615, 1e-4)
	if !point.Illuminated {
		t.Fatalf("58a point should be illuminated")
	}

	point = dial.ShadowPointByHourAngleDeclination(-15, -11.47)
	assertClose(t, "58a x second", point.X, -2.0007, 1e-4)
	assertClose(t, "58a y second", point.Y, -1.1069, 1e-4)
	if !point.Illuminated {
		t.Fatalf("58a second point should be illuminated")
	}
}

func TestPlanarGeometryMatchesBookExamples(t *testing.T) {
	dial := PlanarDial{
		Latitude:                  40,
		PlaneNormalAzimuth:        250,
		PlaneNormalZenithDistance: 50,
		StylusLength:              1,
	}
	geometry := dial.Geometry()
	if !geometry.HasFiniteCenter {
		t.Fatalf("58a geometry should have finite center")
	}
	assertClose(t, "58a center x", geometry.CenterX, 3.3880, 1e-4)
	assertClose(t, "58a center y", geometry.CenterY, -3.1102, 1e-4)
	assertClose(t, "58a polar angle", geometry.PolarStylusPlaneAngle, 12.2672, 5e-4)

	vertical := VerticalDial(-35, 340, 1)
	point := vertical.ShadowPointByHourAngleDeclination(45, 0)
	assertClose(t, "58b x first", point.X, -0.8439, 1e-4)
	assertClose(t, "58b y first", point.Y, -0.9298, 1e-4)

	point = vertical.ShadowPointByHourAngleDeclination(0, 20.15)
	assertClose(t, "58b x second", point.X, 0.3640, 1e-4)
	assertClose(t, "58b y second", point.Y, -0.7410, 1e-4)

	geometry = vertical.Geometry()
	if !geometry.HasFiniteCenter {
		t.Fatalf("58b geometry should have finite center")
	}
	assertClose(t, "58b center x", geometry.CenterX, 0.3640, 1e-4)
	assertClose(t, "58b center y", geometry.CenterY, 0.7451, 1e-4)
	assertClose(t, "58b polar angle", geometry.PolarStylusPlaneAngle, 50.3315, 5e-4)
}

func TestPlanarIlluminationMatchesBookExample58c(t *testing.T) {
	dial := PlanarDial{
		Latitude:                  40,
		PlaneNormalAzimuth:        340,
		PlaneNormalZenithDistance: 75,
		StylusLength:              1,
	}
	declination := 23.44

	if !dial.ShadowPointByHourAngleDeclination(-105, declination).Illuminated {
		t.Fatalf("58c -105° should be illuminated")
	}
	if !dial.ShadowPointByHourAngleDeclination(-90, declination).Illuminated {
		t.Fatalf("58c -90° should be illuminated")
	}
	if dial.ShadowPointByHourAngleDeclination(-75, declination).PlaneIlluminated {
		t.Fatalf("58c -75° should leave the plane unilluminated")
	}
	if dial.ShadowPointByHourAngleDeclination(0, declination).PlaneIlluminated {
		t.Fatalf("58c 0° should remain unilluminated on the plane")
	}
	if !dial.ShadowPointByHourAngleDeclination(15, declination).Illuminated {
		t.Fatalf("58c +15° should be illuminated again")
	}
	if !dial.ShadowPointByHourAngleDeclination(105, declination).Illuminated {
		t.Fatalf("58c +105° should be illuminated")
	}
}

func TestSunAboveHorizonHourAngleIntervals(t *testing.T) {
	intervals := SunAboveHorizonHourAngleIntervals(40, 0)
	if len(intervals) != 1 {
		t.Fatalf("expected one horizon interval, got %d", len(intervals))
	}
	assertClose(t, "equinox rise", intervals[0].Start, -90, 1e-12)
	assertClose(t, "equinox set", intervals[0].End, 90, 1e-12)

	if len(SunAboveHorizonHourAngleIntervals(80, -23.44)) != 0 {
		t.Fatalf("polar night should have no above-horizon interval")
	}

	intervals = SunAboveHorizonHourAngleIntervals(80, 23.44)
	if len(intervals) != 1 || intervals[0].Start != -180 || intervals[0].End != 180 {
		t.Fatalf("polar day should cover the full day, got %+v", intervals)
	}
}

func TestIlluminatedHourAngleIntervalsMatchBookExample58c(t *testing.T) {
	dial := PlanarDial{
		Latitude:                  40,
		PlaneNormalAzimuth:        340,
		PlaneNormalZenithDistance: 75,
		StylusLength:              1,
	}

	intervals := dial.IlluminatedHourAngleIntervals(23.44)
	if len(intervals) != 2 {
		t.Fatalf("expected two illuminated intervals, got %d", len(intervals))
	}
	assertClose(t, "58c first start", intervals[0].Start, -111.33, 0.05)
	assertClose(t, "58c first end", intervals[0].End, -84, 1.0)
	assertClose(t, "58c second start", intervals[1].Start, 2, 1.0)
	assertClose(t, "58c second end", intervals[1].End, 111, 1.0)
}

func TestDeclinationCurveSplitsByIlluminationIntervals(t *testing.T) {
	dial := PlanarDial{
		Latitude:                  40,
		PlaneNormalAzimuth:        340,
		PlaneNormalZenithDistance: 75,
		StylusLength:              1,
	}

	segments := dial.DeclinationCurve(23.44, 15)
	if len(segments) != 2 {
		t.Fatalf("expected two curve segments, got %d", len(segments))
	}
	assertClose(t, "segment first start", segments[0].Interval.Start, -111.33, 0.05)
	assertClose(t, "segment first end", segments[0].Interval.End, -84, 1.0)
	assertClose(t, "segment second start", segments[1].Interval.Start, 2, 1.0)
	assertClose(t, "segment second end", segments[1].Interval.End, 111, 1.0)

	firstHours := []float64{-105, -90}
	secondHours := []float64{15, 30, 45, 60, 75, 90, 105}
	if len(segments[0].Samples) != len(firstHours) {
		t.Fatalf("unexpected first segment sample count: got %d want %d", len(segments[0].Samples), len(firstHours))
	}
	if len(segments[1].Samples) != len(secondHours) {
		t.Fatalf("unexpected second segment sample count: got %d want %d", len(segments[1].Samples), len(secondHours))
	}
	for index, hourAngle := range firstHours {
		assertClose(t, "first segment hour angle", segments[0].Samples[index].HourAngle, hourAngle, 1e-12)
		if !segments[0].Samples[index].Point.Illuminated {
			t.Fatalf("first segment sample %d should be illuminated", index)
		}
	}
	for index, hourAngle := range secondHours {
		assertClose(t, "second segment hour angle", segments[1].Samples[index].HourAngle, hourAngle, 1e-12)
		if !segments[1].Samples[index].Point.Illuminated {
			t.Fatalf("second segment sample %d should be illuminated", index)
		}
	}
}

func TestDeclinationCurveAtMatchesDeclinationChain(t *testing.T) {
	dial := HorizontalDial(31.2304, 1)
	date := time.Date(2026, 6, 21, 9, 30, 0, 0, time.FixedZone("CST", 8*3600))

	got := dial.DeclinationCurveAt(date, 15)
	want := dial.DeclinationCurve(sunDeclination(date), 15)
	if len(got) != len(want) {
		t.Fatalf("segment count mismatch: got %d want %d", len(got), len(want))
	}
	for segmentIndex := range got {
		assertClose(t, "curve start", got[segmentIndex].Interval.Start, want[segmentIndex].Interval.Start, 1e-12)
		assertClose(t, "curve end", got[segmentIndex].Interval.End, want[segmentIndex].Interval.End, 1e-12)
		if len(got[segmentIndex].Samples) != len(want[segmentIndex].Samples) {
			t.Fatalf("sample count mismatch in segment %d: got %d want %d", segmentIndex, len(got[segmentIndex].Samples), len(want[segmentIndex].Samples))
		}
	}
}

func TestMeanSolarTimePointMatchesHourAngleDeclinationChain(t *testing.T) {
	dial := HorizontalDial(31.2304, 1)
	lon := 121.4737
	date := MeanSolarTime(time.Date(2026, 6, 21, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), lon)
	meanSolarHours := 9.5
	sampleTime := dateWithClockHours(date, meanSolarHours)

	got := dial.MeanSolarTimePoint(date, meanSolarHours)
	want := dial.ShadowPointByHourAngleDeclination(HourAngle(sampleTime, longitudeFromTimeZone(sampleTime)), sunDeclination(sampleTime))
	assertClose(t, "mean solar point x", got.X, want.X, 1e-12)
	assertClose(t, "mean solar point y", got.Y, want.Y, 1e-12)
	if got.Illuminated != want.Illuminated {
		t.Fatalf("mean solar point illumination mismatch: got %v want %v", got.Illuminated, want.Illuminated)
	}
}

func TestZoneTimePointMatchesHourAngleDeclinationChain(t *testing.T) {
	dial := HorizontalDial(31.2304, 1)
	date := time.Date(2026, 6, 21, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	lon := 121.4737
	zoneTimeHours := 9.5
	sampleTime := dateWithClockHours(date, zoneTimeHours)

	got := dial.ZoneTimePoint(date, lon, zoneTimeHours)
	want := dial.ShadowPointByHourAngleDeclination(HourAngle(sampleTime, lon), sunDeclination(sampleTime))
	assertClose(t, "zone time point x", got.X, want.X, 1e-12)
	assertClose(t, "zone time point y", got.Y, want.Y, 1e-12)
	if got.Illuminated != want.Illuminated {
		t.Fatalf("zone time point illumination mismatch: got %v want %v", got.Illuminated, want.Illuminated)
	}
}

func TestMeanSolarAndZoneTimeLinesMatchPointHelpers(t *testing.T) {
	dial := HorizontalDial(31.2304, 1)
	lon := 121.4737
	dates := []time.Time{
		MeanSolarTime(time.Date(2026, 3, 21, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), lon),
		MeanSolarTime(time.Date(2026, 6, 21, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), lon),
		MeanSolarTime(time.Date(2026, 9, 23, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), lon),
	}
	meanSolarHours := 10.0
	zoneTimeHours := 10.0

	meanSamples := dial.MeanSolarTimeLine(dates, meanSolarHours)
	if len(meanSamples) != len(dates) {
		t.Fatalf("mean-solar line sample count mismatch: got %d want %d", len(meanSamples), len(dates))
	}
	for index, sample := range meanSamples {
		expectedTime := dateWithClockHours(dates[index], meanSolarHours)
		if !sample.Date.Equal(expectedTime) {
			t.Fatalf("mean-solar line sample instant mismatch at %d: got %s want %s", index, sample.Date, expectedTime)
		}
		assertClose(t, "mean-solar line hour angle", sample.HourAngle, HourAngle(expectedTime, longitudeFromTimeZone(expectedTime)), 1e-12)
		assertClose(t, "mean-solar line declination", sample.Declination, sunDeclination(expectedTime), 1e-12)
		assertClose(t, "mean-solar line x", sample.Point.X, dial.MeanSolarTimePoint(dates[index], meanSolarHours).X, 1e-12)
		assertClose(t, "mean-solar line y", sample.Point.Y, dial.MeanSolarTimePoint(dates[index], meanSolarHours).Y, 1e-12)
	}

	zoneDates := []time.Time{
		time.Date(2026, 3, 21, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		time.Date(2026, 6, 21, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		time.Date(2026, 9, 23, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
	}
	zoneSamples := dial.ZoneTimeLine(zoneDates, lon, zoneTimeHours)
	if len(zoneSamples) != len(dates) {
		t.Fatalf("zone-time line sample count mismatch: got %d want %d", len(zoneSamples), len(dates))
	}
	for index, sample := range zoneSamples {
		expectedTime := dateWithClockHours(zoneDates[index], zoneTimeHours)
		if !sample.Date.Equal(expectedTime) {
			t.Fatalf("zone-time line sample instant mismatch at %d: got %s want %s", index, sample.Date, expectedTime)
		}
		assertClose(t, "zone-time line hour angle", sample.HourAngle, HourAngle(expectedTime, lon), 1e-12)
		assertClose(t, "zone-time line declination", sample.Declination, sunDeclination(expectedTime), 1e-12)
		assertClose(t, "zone-time line x", sample.Point.X, dial.ZoneTimePoint(zoneDates[index], lon, zoneTimeHours).X, 1e-12)
		assertClose(t, "zone-time line y", sample.Point.Y, dial.ZoneTimePoint(zoneDates[index], lon, zoneTimeHours).Y, 1e-12)
	}
}

func TestSpecialDialConstructorsMatchKnownGeometry(t *testing.T) {
	horizontal := HorizontalDial(45, 1)
	geometry := horizontal.Geometry()
	if !geometry.HasFiniteCenter {
		t.Fatalf("horizontal dial should have finite center")
	}
	assertClose(t, "horizontal center x", geometry.CenterX, 0, 1e-12)
	assertClose(t, "horizontal center y", geometry.CenterY, -1, 1e-12)
	assertClose(t, "horizontal polar length", geometry.PolarStylusLength, math.Sqrt2, 1e-12)
	assertClose(t, "horizontal polar angle", geometry.PolarStylusPlaneAngle, 45, 1e-12)

	equatorialNorth := EquatorialNorthDial(40, 1)
	geometry = equatorialNorth.Geometry()
	if !geometry.HasFiniteCenter {
		t.Fatalf("equatorial north dial should have finite center")
	}
	assertClose(t, "equatorial north center x", geometry.CenterX, 0, 1e-12)
	assertClose(t, "equatorial north center y", geometry.CenterY, 0, 1e-12)
	assertClose(t, "equatorial north polar length", geometry.PolarStylusLength, 1, 1e-12)
	assertClose(t, "equatorial north polar angle", geometry.PolarStylusPlaneAngle, 90, 1e-12)

	equatorialSouth := EquatorialSouthDial(40, 1)
	geometry = equatorialSouth.Geometry()
	if !geometry.HasFiniteCenter {
		t.Fatalf("equatorial south dial should have finite center")
	}
	assertClose(t, "equatorial south center x", geometry.CenterX, 0, 1e-12)
	assertClose(t, "equatorial south center y", geometry.CenterY, 0, 1e-12)
	assertClose(t, "equatorial south polar length", geometry.PolarStylusLength, 1, 1e-12)
	assertClose(t, "equatorial south polar angle", geometry.PolarStylusPlaneAngle, 90, 1e-12)
}

func TestShadowPointAtMatchesHourAngleDeclinationChain(t *testing.T) {
	dial := HorizontalDial(31.2304, 1)
	date := time.Date(2026, 6, 21, 9, 30, 0, 0, time.FixedZone("CST", 8*3600))
	lon := 121.4737

	got := dial.ShadowPointAt(date, lon)
	want := dial.ShadowPointByHourAngleDeclination(HourAngle(date, lon), sunDeclination(date))
	assertClose(t, "point at x", got.X, want.X, 1e-12)
	assertClose(t, "point at y", got.Y, want.Y, 1e-12)
	if got.Illuminated != want.Illuminated {
		t.Fatalf("illumination chain mismatch: got %v want %v", got.Illuminated, want.Illuminated)
	}
}

func TestPlanarGeometryDegeneratesWhenPolarStylusParallelToPlane(t *testing.T) {
	dial := PlanarDial{
		Latitude:                  45,
		PlaneNormalAzimuth:        180,
		PlaneNormalZenithDistance: 45,
		StylusLength:              1,
	}
	geometry := dial.Geometry()
	if geometry.HasFiniteCenter {
		t.Fatalf("degenerate geometry should not have a finite center")
	}
	if !math.IsNaN(geometry.CenterX) || !math.IsNaN(geometry.CenterY) || !math.IsNaN(geometry.PolarStylusLength) {
		t.Fatalf("degenerate geometry should return NaN finite quantities")
	}
	assertClose(t, "degenerate polar angle", geometry.PolarStylusPlaneAngle, 0, 1e-12)
}

func sunDeclination(date time.Time) float64 {
	return sun.ApparentDec(date)
}

func assertClose(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.IsNaN(got) || math.Abs(got-want) > tol {
		t.Fatalf("%s mismatch: got %.12f want %.12f tol %.12f", name, got, want, tol)
	}
}
