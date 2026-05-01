package coord

import (
	"math"
	"testing"
	"time"
)

func TestObliquityDrivenCoordinateConversions(t *testing.T) {
	date := time.Date(2026, 4, 27, 10, 30, 45, 0, time.FixedZone("CST", 8*3600))
	lon := 139.686111
	lat := 4.875278
	obliquity := EclipticObliquity(date, true)

	got := EclipticToEquatorialByObliquity(lon, lat, obliquity)
	want := EclipticToEquatorial(date, lon, lat)
	assertClose(t, "manual obliquity ra", got.RA, want.RA, 1e-12)
	assertClose(t, "manual obliquity dec", got.Dec, want.Dec, 1e-12)

	back := EquatorialToEclipticByObliquity(got.RA, got.Dec, obliquity)
	assertClose(t, "manual obliquity lon", back.Lon, lon, 1e-10)
	assertClose(t, "manual obliquity lat", back.Lat, lat, 1e-10)
}

func TestHourAngleDrivenHorizontalConversions(t *testing.T) {
	date := time.Date(2026, 4, 27, 2, 30, 45, 0, time.UTC)
	ra := 101.28715533
	dec := -16.71611586
	observerLon := 115.0
	observerLat := 40.0
	hourAngle := HourAngle(date, ra, observerLon)

	got := HourAngleDeclinationToHorizontal(hourAngle, dec, observerLat)
	want := EquatorialToHorizontal(date, ra, dec, observerLon, observerLat)
	assertClose(t, "manual hour angle azimuth", got.Azimuth, want.Azimuth, 1e-12)
	assertClose(t, "manual hour angle altitude", got.Altitude, want.Altitude, 1e-12)
	assertClose(t, "manual hour angle zenith", got.Zenith, want.Zenith, 1e-12)
	assertClose(t, "manual hour angle value", got.HourAngle, want.HourAngle, 1e-12)

	roundTripHourAngle, roundTripDeclination := HorizontalToHourAngleDeclination(got.Azimuth, got.Altitude, observerLat)
	assertClose(t, "round trip hour angle", roundTripHourAngle, normalize360(hourAngle), 1e-10)
	assertClose(t, "round trip declination", roundTripDeclination, dec, 1e-10)
}

func TestLocalSiderealTimeDrivenHorizontalConversions(t *testing.T) {
	date := time.Date(2026, 4, 27, 2, 30, 45, 0, time.UTC)
	ra := 101.28715533
	dec := -16.71611586
	observerLon := 115.0
	observerLat := 40.0
	localSiderealTimeHours := math.Mod(ApparentSiderealTime(date)+observerLon/15, 24)
	if localSiderealTimeHours < 0 {
		localSiderealTimeHours += 24
	}

	got := EquatorialToHorizontalByLocalSiderealTime(localSiderealTimeHours, ra, dec, observerLat)
	want := EquatorialToHorizontal(date, ra, dec, observerLon, observerLat)
	assertClose(t, "LST azimuth", got.Azimuth, want.Azimuth, 1e-12)
	assertClose(t, "LST altitude", got.Altitude, want.Altitude, 1e-12)
	assertClose(t, "LST zenith", got.Zenith, want.Zenith, 1e-12)
	assertClose(t, "LST hour angle", got.HourAngle, want.HourAngle, 1e-12)

	back := HorizontalToEquatorialByLocalSiderealTime(localSiderealTimeHours, got.Azimuth, got.Altitude, observerLat)
	assertClose(t, "LST round trip ra", back.RA, ra, 1e-10)
	assertClose(t, "LST round trip dec", back.Dec, dec, 1e-10)
}

func TestGalacticCoordinateConversions(t *testing.T) {
	galacticCenter := EquatorialToGalactic(266.4051, -28.936175)
	assertClose(t, "galactic center lon", galacticCenter.Lon, 0, 5e-4)
	assertClose(t, "galactic center lat", galacticCenter.Lat, 0, 5e-4)

	pole := GalacticToEquatorial(0, 90)
	assertClose(t, "north galactic pole ra", pole.RA, 192.85948, 1e-5)
	assertClose(t, "north galactic pole dec", pole.Dec, 27.12825, 1e-5)

	sample := EquatorialToGalactic(83.6331, 22.0145)
	back := GalacticToEquatorial(sample.Lon, sample.Lat)
	assertClose(t, "galactic round trip ra", back.RA, 83.6331, 1e-10)
	assertClose(t, "galactic round trip dec", back.Dec, 22.0145, 1e-10)
}

func TestHorizontalRoundTripAcrossQuadrants(t *testing.T) {
	samples := []struct {
		hourAngle   float64
		declination float64
		latitude    float64
	}{
		{15, 20, 35},
		{95, -10, 52},
		{210, 45, -20},
		{315, -35, 10},
	}

	for _, sample := range samples {
		hz := HourAngleDeclinationToHorizontal(sample.hourAngle, sample.declination, sample.latitude)
		hourAngle, declination := HorizontalToHourAngleDeclination(hz.Azimuth, hz.Altitude, sample.latitude)
		if math.Abs(hourAngle-normalize360(sample.hourAngle)) > 1e-10 {
			t.Fatalf("hour angle round trip mismatch: got %.15f want %.15f", hourAngle, normalize360(sample.hourAngle))
		}
		if math.Abs(declination-sample.declination) > 1e-10 {
			t.Fatalf("declination round trip mismatch: got %.15f want %.15f", declination, sample.declination)
		}
	}
}
