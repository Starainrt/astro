package coord

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func assertClose(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s mismatch: got %.15f want %.15f", name, got, want)
	}
}

func TestEclipticEquatorialWrappers(t *testing.T) {
	date := time.Date(2026, 4, 27, 10, 30, 45, 0, time.FixedZone("CST", 8*3600))
	jde := basic.Date2JDE(date.UTC())
	lon := 139.686111
	lat := 4.875278

	got := EclipticToEquatorial(date, lon, lat)
	wantRA, wantDec := basic.LoBoToRaDec(jde, lon, lat)
	assertClose(t, "ra", got.RA, wantRA, 1e-12)
	assertClose(t, "dec", got.Dec, wantDec, 1e-12)

	back := EquatorialToEcliptic(date, got.RA, got.Dec)
	assertClose(t, "lon", back.Lon, lon, 1e-10)
	assertClose(t, "lat", back.Lat, lat, 1e-10)
}

func TestTimeAndPrecessionWrappers(t *testing.T) {
	date := time.Date(2026, 4, 27, 2, 30, 45, 0, time.UTC)
	to := time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC)
	jde := basic.Date2JDE(date.UTC())

	assertClose(t, "mean sidereal time", MeanSiderealTime(date), basic.MeanSiderealTime(jde), 1e-12)
	assertClose(t, "apparent sidereal time", ApparentSiderealTime(date), basic.ApparentSiderealTime(jde), 1e-12)
	assertClose(t, "obliquity", EclipticObliquity(date, true), basic.EclipticObliquity(jde, true), 1e-12)

	gotLon, gotObl := Nutation2000B(date)
	wantLon, wantObl := basic.Nutation2000B(jde)
	assertClose(t, "nutation longitude", gotLon, wantLon, 1e-12)
	assertClose(t, "nutation obliquity", gotObl, wantObl, 1e-12)

	got := Precess(date, to, 101.28715533, -16.71611586)
	wantRA, wantDec := basic.Precess(101.28715533, -16.71611586, jde, basic.Date2JDE(to.UTC()))
	assertClose(t, "precess ra", got.RA, wantRA, 1e-12)
	assertClose(t, "precess dec", got.Dec, wantDec, 1e-12)
}

func TestHorizontalAndTopocentricWrappers(t *testing.T) {
	date := time.Date(2026, 4, 27, 2, 30, 45, 0, time.UTC)
	jde := basic.Date2JDE(date.UTC())
	ra := 101.28715533
	dec := -16.71611586
	observerLon := 115.0
	observerLat := 40.0

	hz := EquatorialToHorizontal(date, ra, dec, observerLon, observerLat)
	wantAltitude := basic.StarHeight(jde, ra, dec, observerLon, observerLat, 0)
	assertClose(t, "altitude", hz.Altitude, wantAltitude, 1e-12)
	assertClose(t, "zenith", hz.Zenith, 90-wantAltitude, 1e-12)
	assertClose(t, "azimuth", hz.Azimuth, basic.StarAzimuth(jde, ra, dec, observerLon, observerLat, 0), 1e-12)
	assertClose(t, "hour angle", hz.HourAngle, basic.StarHourAngle(jde, ra, observerLon, 0), 1e-12)
	assertClose(t, "hour angle func", HourAngle(date, ra, observerLon), hz.HourAngle, 1e-12)

	top := TopocentricEquatorial(date, ra, dec, observerLon, observerLat, 0.00257, 53)
	wantRA, wantDec := basic.TopocentricRaDec(ra, dec, observerLat, observerLon, jde, 0.00257, 53)
	assertClose(t, "topocentric ra", top.RA, wantRA, 1e-12)
	assertClose(t, "topocentric dec", top.Dec, wantDec, 1e-12)

	ecl := TopocentricEcliptic(date, 139.686111, 4.875278, observerLon, observerLat, 0.00257, 53)
	assertClose(t, "topocentric lon", ecl.Lon, basic.TopocentricLo(139.686111, 4.875278, observerLat, observerLon, jde, 0.00257, 53), 1e-12)
	assertClose(t, "topocentric lat", ecl.Lat, basic.TopocentricBo(139.686111, 4.875278, observerLat, observerLon, jde, 0.00257, 53), 1e-12)
}

func TestAngularSeparationWrapper(t *testing.T) {
	got := AngularSeparation(101.28715533, -16.71611586, 95.9879578, -52.6956611)
	want := basic.StarAngularSeparation(101.28715533, -16.71611586, 95.9879578, -52.6956611)
	assertClose(t, "angular separation", got, want, 1e-12)
}
