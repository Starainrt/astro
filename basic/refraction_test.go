package basic

import (
	"math"
	"testing"
)

func TestRefractionFromApparentAltitudeStandardAtmosphere(t *testing.T) {
	assertClose(t, "Refraction@0deg", RefractionFromApparentAltitude(0, 1010, 10), 0.483032, 0.0001)
	assertClose(t, "Refraction@45deg", RefractionFromApparentAltitude(45, 1010, 10), 0.016878, 0.0001)
}

func TestRefractionRoundTrip(t *testing.T) {
	cases := []struct {
		altitude     float64
		pressureHPa  float64
		temperatureC float64
	}{
		{-1, 980, 25},
		{0, 1010, 10},
		{5, 1013.25, 15},
		{15, 900, -10},
		{45, 1010, 0},
		{80, 1030, 30},
	}

	for _, tc := range cases {
		apparentAltitude := ApparentAltitude(tc.altitude, tc.pressureHPa, tc.temperatureC)
		trueAltitude := TrueAltitude(apparentAltitude, tc.pressureHPa, tc.temperatureC)
		assertClose(t, "TrueAltitudeRoundTrip", trueAltitude, tc.altitude, 1e-9)

		trueFromApparent := TrueAltitude(tc.altitude, tc.pressureHPa, tc.temperatureC)
		apparentFromTrue := ApparentAltitude(trueFromApparent, tc.pressureHPa, tc.temperatureC)
		assertClose(t, "ApparentAltitudeRoundTrip", apparentFromTrue, tc.altitude, 1e-9)
	}
}

func TestRefractionScalingAndBounds(t *testing.T) {
	lowPressure := RefractionFromApparentAltitude(0, 980, 10)
	highPressure := RefractionFromApparentAltitude(0, 1030, 10)
	if !(lowPressure < highPressure) {
		t.Fatalf("pressure scaling mismatch: low %.12f high %.12f", lowPressure, highPressure)
	}

	cold := RefractionFromApparentAltitude(0, 1010, 0)
	hot := RefractionFromApparentAltitude(0, 1010, 30)
	if !(cold > hot) {
		t.Fatalf("temperature scaling mismatch: cold %.12f hot %.12f", cold, hot)
	}

	if RefractionFromApparentAltitude(-6, 1010, 10) != 0 {
		t.Fatalf("refraction below lower limit should be 0")
	}
	if RefractionFromApparentAltitude(95, 1010, 10) != 0 {
		t.Fatalf("refraction above upper limit should be 0")
	}
	if !math.IsNaN(RefractionFromApparentAltitude(0, 0, 10)) {
		t.Fatalf("invalid pressure should produce NaN")
	}
	if !math.IsNaN(RefractionFromApparentAltitude(0, 1010, -274)) {
		t.Fatalf("invalid temperature should produce NaN")
	}
}
