package basic

import (
	"math"
	"testing"
)

func TestPlanetEventsRejectNonFiniteQuery(t *testing.T) {
	nan := math.NaN()
	events := []struct {
		name string
		fn   func(float64) float64
	}{
		{"Mars conjunction", NextMarsConjunction},
		{"Jupiter conjunction", NextJupiterConjunction},
		{"Saturn conjunction", NextSaturnConjunction},
		{"Uranus conjunction", NextUranusConjunction},
		{"Neptune conjunction", NextNeptuneConjunction},
		{"Mercury conjunction", NextMercuryConjunction},
		{"Mercury greatest elongation", NextMercuryGreatestElongation},
		{"Venus greatest elongation", NextVenusGreatestElongation},
		{"Venus station", NextVenusProgradeToRetrograde},
	}
	for _, tc := range events {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.fn(nan); !math.IsNaN(got) {
				t.Fatalf("got %v, want NaN", got)
			}
		})
	}
}

func TestPlanetObservationRejectsNonFiniteQuery(t *testing.T) {
	nan := math.NaN()
	culminations := []struct {
		name string
		fn   func(float64, float64, float64) float64
	}{
		{"Mercury", MercuryCulminationTime},
		{"Venus", VenusCulminationTime},
		{"Mars", MarsCulminationTime},
		{"Jupiter", JupiterCulminationTime},
		{"Saturn", SaturnCulminationTime},
		{"Uranus", UranusCulminationTime},
		{"Neptune", NeptuneCulminationTime},
	}
	for _, tc := range culminations {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.fn(nan, 0, 0); !math.IsNaN(got) {
				t.Fatalf("got %v, want NaN", got)
			}
		})
	}
	if got := OrbitCulminationTime(nan, 0, 0, 0, 0, OrbitElements{}); !math.IsNaN(got) {
		t.Fatalf("OrbitCulminationTime got %v, want NaN", got)
	}
	if got := MercuryCulminationTimeN(nan, 0, 0, 16); !math.IsNaN(got) {
		t.Fatalf("MercuryCulminationTimeN got %v, want NaN", got)
	}

	if _, err := MarsRiseTime(nan, 0, 0, 0, 0, 0); err != ErrInvalidObservationInput {
		t.Fatalf("MarsRiseTime error = %v, want ErrInvalidObservationInput", err)
	}
	if _, err := MercuryRiseTimeN(nan, 0, 0, 0, 0, 0, 16); err != ErrInvalidObservationInput {
		t.Fatalf("MercuryRiseTimeN error = %v, want ErrInvalidObservationInput", err)
	}
	if got := JupiterGalileanSatelliteState(nan, 1); !math.IsNaN(got.X) {
		t.Fatalf("JupiterGalileanSatelliteState X = %v, want NaN", got.X)
	}
}

func TestSharedObservationRejectsNonFiniteQuery(t *testing.T) {
	nan := math.NaN()
	if got := MoonCulminationTime(nan, 0, 0, 0); !math.IsNaN(got) {
		t.Fatalf("MoonCulminationTime got %v, want NaN", got)
	}
	if got := StarCulminationTime(nan, 0, 0, 0); !math.IsNaN(got) {
		t.Fatalf("StarCulminationTime got %v, want NaN", got)
	}
	if got := CalcMoonSHByJDE(nan, 0); !math.IsNaN(got) {
		t.Fatalf("CalcMoonSHByJDE got %v, want NaN", got)
	}

	if _, err := GetSunRiseTime(nan, 0, 0, 0, 0, 0); err != ErrInvalidObservationInput {
		t.Fatalf("GetSunRiseTime error = %v, want ErrInvalidObservationInput", err)
	}
	if _, err := GetMoonRiseTime(nan, 0, 0, 0, 0, 0); err != ErrInvalidObservationInput {
		t.Fatalf("GetMoonRiseTime error = %v, want ErrInvalidObservationInput", err)
	}
	if _, err := StarRiseTime(nan, 0, 0, 0, 0, 0, 0, false); err != ErrInvalidObservationInput {
		t.Fatalf("StarRiseTime error = %v, want ErrInvalidObservationInput", err)
	}
}
