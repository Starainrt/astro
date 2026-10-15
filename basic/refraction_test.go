package basic

import (
	"math"
	"testing"
)

func TestRefractionFromTrueAltitudeStandardAtmosphere(t *testing.T) {
	assertClose(t, "RefractionFromTrue@0deg", RefractionFromTrueAltitude(0, 1010, 10), 0.483032, 0.0001)
	assertClose(t, "RefractionFromTrue@45deg", RefractionFromTrueAltitude(45, 1010, 10), 0.016878, 0.0001)
	assertClose(t, "ApparentAltitude@0deg", ApparentAltitude(0, 1010, 10), 0.483032, 0.0001)
}

func TestRefractionFromApparentAltitudeUsesApparentInput(t *testing.T) {
	// 视地平线处修正约为 34.5 角分，而不是真高度角公式在 0 度返回的 29 角分 / At the apparent horizon the correction is about 34.5 arcminutes, not
	// 29 角分 / the 29 arcminutes returned by the true-altitude formula at 0 degrees.
	assertClose(t, "RefractionFromApparent@0deg", RefractionFromApparentAltitude(0, 1010, 10), 0.574, 0.001)
	assertClose(t, "TrueAltitude@0deg", TrueAltitude(0, 1010, 10), -0.574, 0.001)
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

func TestRefractionRoundTripNearZenith(t *testing.T) {
	for _, trueAltitude := range []float64{89.90, 89.95, 89.99} {
		apparentAltitude := ApparentAltitude(trueAltitude, 1010, 10)
		if math.IsNaN(apparentAltitude) || math.IsInf(apparentAltitude, 0) {
			t.Fatalf("true altitude %.2f produced invalid apparent altitude %v", trueAltitude, apparentAltitude)
		}
		got := TrueAltitude(apparentAltitude, 1010, 10)
		if math.IsNaN(got) || math.IsInf(got, 0) {
			t.Fatalf("apparent altitude %.12f produced invalid true altitude %v", apparentAltitude, got)
		}
		assertClose(t, "NearZenithRoundTrip", got, trueAltitude, 1e-9)
		if refraction := RefractionFromApparentAltitude(apparentAltitude, 1010, 10); math.IsNaN(refraction) || math.IsInf(refraction, 0) {
			t.Fatalf("apparent altitude %.12f produced invalid refraction %v", apparentAltitude, refraction)
		}
	}
}

func TestRefractionExtremeTemperatureInverse(t *testing.T) {
	const (
		apparentAltitude = 0.0
		pressureHPa      = 1010.0
	)
	trueAltitude := TrueAltitude(apparentAltitude, pressureHPa, -273)
	if math.IsNaN(trueAltitude) || math.IsInf(trueAltitude, 0) {
		t.Fatalf("-273 C inverse returned invalid true altitude %v", trueAltitude)
	}
	assertClose(t, "ExtremeTemperatureRoundTrip",
		ApparentAltitude(trueAltitude, pressureHPa, -273), apparentAltitude, 1e-9)

	for _, temperatureC := range []float64{-273.14, math.Nextafter(-273.15, math.Inf(1))} {
		if got := TrueAltitude(apparentAltitude, pressureHPa, temperatureC); !math.IsNaN(got) {
			t.Errorf("temperature %.17g C inverse = %v, want NaN when the model has no root", temperatureC, got)
		}
		if got := RefractionFromApparentAltitude(apparentAltitude, pressureHPa, temperatureC); !math.IsNaN(got) {
			t.Errorf("temperature %.17g C apparent refraction = %v, want NaN when the model has no root", temperatureC, got)
		}
	}
}

func TestRefractionScalingAndBounds(t *testing.T) {
	lowPressure := RefractionFromTrueAltitude(0, 980, 10)
	highPressure := RefractionFromTrueAltitude(0, 1030, 10)
	if !(lowPressure < highPressure) {
		t.Fatalf("pressure scaling mismatch: low %.12f high %.12f", lowPressure, highPressure)
	}

	cold := RefractionFromTrueAltitude(0, 1010, 0)
	hot := RefractionFromTrueAltitude(0, 1010, 30)
	if !(cold > hot) {
		t.Fatalf("temperature scaling mismatch: cold %.12f hot %.12f", cold, hot)
	}

	if RefractionFromTrueAltitude(-6, 1010, 10) != 0 {
		t.Fatalf("refraction below lower limit should be 0")
	}
	if RefractionFromTrueAltitude(95, 1010, 10) != 0 {
		t.Fatalf("refraction above upper limit should be 0")
	}
	if !math.IsNaN(RefractionFromTrueAltitude(0, 0, 10)) {
		t.Fatalf("invalid pressure should produce NaN")
	}
	if !math.IsNaN(RefractionFromTrueAltitude(0, 1010, -274)) {
		t.Fatalf("invalid temperature should produce NaN")
	}
	if !math.IsNaN(RefractionFromTrueAltitude(0, 1010, -273.15)) {
		t.Fatalf("absolute zero should produce NaN")
	}
	for _, temperatureC := range []float64{-273, -273.14} {
		refraction := RefractionFromTrueAltitude(0, 1010, temperatureC)
		if math.IsNaN(refraction) || math.IsInf(refraction, 0) || refraction <= 0 {
			t.Fatalf("temperature %.2f C produced invalid refraction %v", temperatureC, refraction)
		}
	}
}
