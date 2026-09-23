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

func TestMoonMaximumDeclinationEventsRejectNonFiniteQuery(t *testing.T) {
	nonFinite := []struct {
		name string
		jd   float64
	}{
		{"NaN", math.NaN()},
		{"+Inf", math.Inf(1)},
		{"-Inf", math.Inf(-1)},
	}
	events := []struct {
		name string
		fn   func(float64) DeclinationEvent
	}{
		{"LastMoonMaximumNorthDeclination", LastMoonMaximumNorthDeclination},
		{"NextMoonMaximumNorthDeclination", NextMoonMaximumNorthDeclination},
		{"ClosestMoonMaximumNorthDeclination", ClosestMoonMaximumNorthDeclination},
		{"LastMoonMaximumSouthDeclination", LastMoonMaximumSouthDeclination},
		{"NextMoonMaximumSouthDeclination", NextMoonMaximumSouthDeclination},
		{"ClosestMoonMaximumSouthDeclination", ClosestMoonMaximumSouthDeclination},
	}
	for _, event := range events {
		for _, query := range nonFinite {
			// A non-finite query used to seed an out-of-range cycle index and spin in the
			// sampling sweep forever; it must return the zero event instead.
			got := event.fn(query.jd)
			if got.JD != 0 || got.Declination != 0 {
				t.Fatalf("%s(%s) = %+v, want zero event", event.name, query.name, got)
			}
		}
	}

	// The finite path must keep working.
	const jd = 2460310.5
	north := NextMoonMaximumNorthDeclination(jd)
	if !(north.JD > jd) || north.Declination == 0 {
		t.Fatalf("NextMoonMaximumNorthDeclination(%v) = %+v, want a later event", jd, north)
	}
	south := LastMoonMaximumSouthDeclination(jd)
	if !(south.JD <= jd) || south.Declination == 0 {
		t.Fatalf("LastMoonMaximumSouthDeclination(%v) = %+v, want an earlier event", jd, south)
	}
}

func TestOuterPlanetStationEventsRejectNonFiniteQuery(t *testing.T) {
	nonFinite := []struct {
		name string
		jd   float64
	}{
		{"NaN", math.NaN()},
		{"+Inf", math.Inf(1)},
		{"-Inf", math.Inf(-1)},
	}
	events := []struct {
		name string
		fn   func(float64) float64
	}{
		{"Mercury", NextMercuryProgradeToRetrograde},
		{"Mercury retrograde", NextMercuryRetrograde},
		{"Venus", NextVenusProgradeToRetrograde},
		{"Venus retrograde", NextVenusRetrograde},
		{"Mars", NextMarsProgradeToRetrograde},
		{"Jupiter", NextJupiterProgradeToRetrograde},
		{"Saturn", NextSaturnProgradeToRetrograde},
		{"Uranus", NextUranusProgradeToRetrograde},
		{"Neptune", NextNeptuneProgradeToRetrograde},
		{"Mars opposition", NextMarsOpposition},
		{"Jupiter opposition", NextJupiterOpposition},
		{"Saturn opposition", NextSaturnOpposition},
		{"Uranus opposition", NextUranusOpposition},
		{"Neptune opposition", NextNeptuneOpposition},
		{"Mars eastern quadrature", NextMarsEasternQuadrature},
		{"Jupiter eastern quadrature", NextJupiterEasternQuadrature},
		{"Saturn eastern quadrature", NextSaturnEasternQuadrature},
		{"Uranus eastern quadrature", NextUranusEasternQuadrature},
		{"Neptune eastern quadrature", NextNeptuneEasternQuadrature},
	}
	for _, event := range events {
		for _, query := range nonFinite {
			if got := event.fn(query.jd); !math.IsNaN(got) {
				t.Fatalf("%s(%s) = %v, want NaN", event.name, query.name, got)
			}
		}
	}
}

func TestOuterPlanetLastStationEventsRejectNonFiniteQuery(t *testing.T) {
	events := []struct {
		name string
		fn   func(float64) float64
	}{
		{"LastMarsProgradeToRetrograde", LastMarsProgradeToRetrograde},
		{"LastMarsRetrogradeToPrograde", LastMarsRetrogradeToPrograde},
		{"LastJupiterProgradeToRetrograde", LastJupiterProgradeToRetrograde},
		{"LastJupiterRetrogradeToPrograde", LastJupiterRetrogradeToPrograde},
		{"LastSaturnProgradeToRetrograde", LastSaturnProgradeToRetrograde},
		{"LastSaturnRetrogradeToPrograde", LastSaturnRetrogradeToPrograde},
		{"LastUranusProgradeToRetrograde", LastUranusProgradeToRetrograde},
		{"LastUranusRetrogradeToPrograde", LastUranusRetrogradeToPrograde},
		{"LastNeptuneProgradeToRetrograde", LastNeptuneProgradeToRetrograde},
		{"LastNeptuneRetrogradeToPrograde", LastNeptuneRetrogradeToPrograde},
	}
	for _, event := range events {
		if got := event.fn(math.NaN()); !math.IsNaN(got) {
			t.Fatalf("%s(NaN) = %v, want NaN", event.name, got)
		}
	}
}

// 查询落在事件之后 1 秒：首候选必须被侧向不变量拒绝，第二候选也必须通过同一不变量。
func TestOuterPlanetStationSecondCandidateKeepsSideInvariant(t *testing.T) {
	const offsetSeconds = 1.0
	cases := []struct {
		name string
		fn   func(float64) float64
	}{
		{"Mars", NextMarsProgradeToRetrograde},
		{"Jupiter", NextJupiterProgradeToRetrograde},
		{"Saturn", NextSaturnProgradeToRetrograde},
		{"Uranus", NextUranusProgradeToRetrograde},
		{"Neptune", NextNeptuneProgradeToRetrograde},
	}
	for _, tc := range cases {
		event := tc.fn(JDCalc(2025, 1, 1))
		if math.IsNaN(event) {
			t.Fatalf("%s: no station event found", tc.name)
		}
		query := event + offsetSeconds/86400.0
		got := tc.fn(UTC2TT(query))
		if math.IsNaN(got) {
			t.Fatalf("%s: second candidate rejected as NaN at %s", tc.name, JD2Date(query))
		}
		if got < query-stationQueryToleranceUT {
			t.Fatalf("%s: returned %.9f before the query %.9f", tc.name, got, query)
		}
	}
}
