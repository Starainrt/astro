package moon

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestGeocentricApparentRaDecComponentsMatch(t *testing.T) {
	date := time.Date(2026, 1, 1, 6, 0, 0, 0, time.UTC)

	ra, dec := GeocentricApparentRaDec(date)
	if diff := math.Abs(ra - GeocentricApparentRa(date)); diff > 1e-12 {
		t.Fatalf("RA pair mismatch: got %.15f want %.15f", ra, GeocentricApparentRa(date))
	}
	if diff := math.Abs(dec - GeocentricApparentDec(date)); diff > 1e-12 {
		t.Fatalf("Dec pair mismatch: got %.15f want %.15f", dec, GeocentricApparentDec(date))
	}
}

func TestGeocentricApparentRaDecDiffersFromTopocentricAtSite(t *testing.T) {
	date := time.Date(2026, 1, 1, 6, 0, 0, 0, time.FixedZone("CST", 8*3600))

	geoRA, geoDec := GeocentricApparentRaDec(date)
	topoRA, topoDec := ApparentRaDec(date, 121.4737, 31.2304)

	if math.Abs(geoRA-topoRA) < 1e-6 && math.Abs(geoDec-topoDec) < 1e-6 {
		t.Fatalf("geocentric apparent RA/Dec unexpectedly matches topocentric values: geo=(%.12f, %.12f) topo=(%.12f, %.12f)",
			geoRA, geoDec, topoRA, topoDec)
	}
}

func TestTrueRaDecUsesBasicGeocentricTrue(t *testing.T) {
	date := time.Date(2026, 1, 1, 6, 0, 0, 0, time.UTC)

	wantRA, wantDec := basic.HMoonGeocentricTrueRaDec(basic.TD2UT(basic.Date2JDE(date.UTC()), true))
	gotRA, gotDec := TrueRaDec(date)
	if math.Abs(gotRA-wantRA) > 1e-12 || math.Abs(gotDec-wantDec) > 1e-12 {
		t.Fatalf("TrueRaDec mismatch: got (%.15f, %.15f) want (%.15f, %.15f)", gotRA, gotDec, wantRA, wantDec)
	}
}
