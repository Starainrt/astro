package basic

import (
	"math"
	"testing"
)

func TestHMoonGeocentricApparentRaDecComponentsMatch(t *testing.T) {
	jd := TD2UT(JDECalc(2026, 1, 1.25), true)

	ra, dec := HMoonGeocentricApparentRaDec(jd)
	if diff := math.Abs(ra - HMoonGeocentricApparentRa(jd)); diff > 1e-12 {
		t.Fatalf("RA pair mismatch: got %.15f want %.15f", ra, HMoonGeocentricApparentRa(jd))
	}
	if diff := math.Abs(dec - HMoonGeocentricApparentDec(jd)); diff > 1e-12 {
		t.Fatalf("Dec pair mismatch: got %.15f want %.15f", dec, HMoonGeocentricApparentDec(jd))
	}
}

func TestHMoonGeocentricTrueRaDecComponentsMatch(t *testing.T) {
	jd := TD2UT(JDECalc(2026, 1, 1.25), true)

	ra, dec := HMoonGeocentricTrueRaDec(jd)
	if diff := math.Abs(ra - HMoonGeocentricTrueRa(jd)); diff > 1e-12 {
		t.Fatalf("RA pair mismatch: got %.15f want %.15f", ra, HMoonGeocentricTrueRa(jd))
	}
	if diff := math.Abs(dec - HMoonGeocentricTrueDec(jd)); diff > 1e-12 {
		t.Fatalf("Dec pair mismatch: got %.15f want %.15f", dec, HMoonGeocentricTrueDec(jd))
	}
}
