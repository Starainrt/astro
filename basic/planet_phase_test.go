package basic

import (
	"math"
	"testing"
	"time"
)

func TestVenusIlluminatedFractionMeeusExample(t *testing.T) {
	jd := Date2JDE(time.Date(1992, 12, 20, 0, 0, 0, 0, time.UTC))

	assertPlanetPhaseClose(t, "VenusPhaseAngle", VenusPhaseAngle(jd), 72.96, 0.01)
	assertPlanetPhaseClose(t, "VenusIlluminatedFraction", VenusIlluminatedFraction(jd), 0.647, 0.001)
}

func TestPlanetIlluminatedFractionRanges(t *testing.T) {
	jd := TD2UT(Date2JDE(time.Date(2026, 4, 26, 9, 30, 45, 0, time.UTC)), true)
	cases := []struct {
		name          string
		phaseAngle    func(float64) float64
		fraction      func(float64) float64
		positionAngle func(float64) float64
	}{
		{"Mercury", MercuryPhaseAngle, MercuryIlluminatedFraction, MercuryBrightLimbPositionAngle},
		{"Venus", VenusPhaseAngle, VenusIlluminatedFraction, VenusBrightLimbPositionAngle},
		{"Mars", MarsPhaseAngle, MarsIlluminatedFraction, MarsBrightLimbPositionAngle},
		{"Jupiter", JupiterPhaseAngle, JupiterIlluminatedFraction, JupiterBrightLimbPositionAngle},
		{"Saturn", SaturnPhaseAngle, SaturnIlluminatedFraction, SaturnBrightLimbPositionAngle},
		{"Uranus", UranusPhaseAngle, UranusIlluminatedFraction, UranusBrightLimbPositionAngle},
		{"Neptune", NeptunePhaseAngle, NeptuneIlluminatedFraction, NeptuneBrightLimbPositionAngle},
	}

	for _, tc := range cases {
		phaseAngle := tc.phaseAngle(jd)
		if phaseAngle < 0 || phaseAngle > 180 {
			t.Fatalf("%s phase angle out of range: %.12f", tc.name, phaseAngle)
		}
		fraction := tc.fraction(jd)
		if fraction < 0 || fraction > 1 {
			t.Fatalf("%s illuminated fraction out of range: %.12f", tc.name, fraction)
		}
		positionAngle := tc.positionAngle(jd)
		if positionAngle < 0 || positionAngle >= 360 {
			t.Fatalf("%s bright limb position angle out of range: %.12f", tc.name, positionAngle)
		}
	}
}

func assertPlanetPhaseClose(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s mismatch: got %.12f want %.12f tolerance %.12f", name, got, want, tolerance)
	}
}
