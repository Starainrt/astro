package basic

import (
	"testing"
	"time"
)

func TestOuterPlanetExactEventBoundaryIncludesCurrent(t *testing.T) {
	cases := []struct {
		name   string
		seed   float64
		lastFn func(float64) float64
		nextFn func(float64) float64
	}{
		{name: "JupiterConjunction", seed: NextJupiterConjunction(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastJupiterConjunction, nextFn: NextJupiterConjunction},
		{name: "JupiterOpposition", seed: NextJupiterOpposition(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastJupiterOpposition, nextFn: NextJupiterOpposition},
		{name: "JupiterEasternQuadrature", seed: NextJupiterEasternQuadrature(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastJupiterEasternQuadrature, nextFn: NextJupiterEasternQuadrature},
		{name: "JupiterWesternQuadrature", seed: NextJupiterWesternQuadrature(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastJupiterWesternQuadrature, nextFn: NextJupiterWesternQuadrature},
		{name: "JupiterP2R", seed: NextJupiterProgradeToRetrograde(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastJupiterProgradeToRetrograde, nextFn: NextJupiterProgradeToRetrograde},
		{name: "JupiterR2P", seed: NextJupiterRetrogradeToPrograde(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastJupiterRetrogradeToPrograde, nextFn: NextJupiterRetrogradeToPrograde},
		{name: "SaturnOpposition", seed: NextSaturnOpposition(ttjdUTC(2025, 1, 1, 0, 0, 0)), lastFn: LastSaturnOpposition, nextFn: NextSaturnOpposition},
		{name: "SaturnP2R", seed: NextSaturnProgradeToRetrograde(ttjdUTC(2025, 1, 1, 0, 0, 0)), lastFn: LastSaturnProgradeToRetrograde, nextFn: NextSaturnProgradeToRetrograde},
		{name: "SaturnR2P", seed: NextSaturnRetrogradeToPrograde(ttjdUTC(2025, 1, 1, 0, 0, 0)), lastFn: LastSaturnRetrogradeToPrograde, nextFn: NextSaturnRetrogradeToPrograde},
		{name: "UranusOpposition", seed: NextUranusOpposition(ttjdUTC(2025, 1, 1, 0, 0, 0)), lastFn: LastUranusOpposition, nextFn: NextUranusOpposition},
		{name: "UranusP2R", seed: NextUranusProgradeToRetrograde(ttjdUTC(2025, 1, 1, 0, 0, 0)), lastFn: LastUranusProgradeToRetrograde, nextFn: NextUranusProgradeToRetrograde},
		{name: "UranusR2P", seed: NextUranusRetrogradeToPrograde(ttjdUTC(2025, 1, 1, 0, 0, 0)), lastFn: LastUranusRetrogradeToPrograde, nextFn: NextUranusRetrogradeToPrograde},
		{name: "NeptuneOpposition", seed: NextNeptuneOpposition(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastNeptuneOpposition, nextFn: NextNeptuneOpposition},
		{name: "NeptuneP2R", seed: NextNeptuneProgradeToRetrograde(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastNeptuneProgradeToRetrograde, nextFn: NextNeptuneProgradeToRetrograde},
		{name: "NeptuneR2P", seed: NextNeptuneRetrogradeToPrograde(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastNeptuneRetrogradeToPrograde, nextFn: NextNeptuneRetrogradeToPrograde},
		{name: "MarsConjunction", seed: NextMarsConjunction(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastMarsConjunction, nextFn: NextMarsConjunction},
		{name: "MarsOpposition", seed: NextMarsOpposition(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastMarsOpposition, nextFn: NextMarsOpposition},
		{name: "MarsEasternQuadrature", seed: NextMarsEasternQuadrature(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastMarsEasternQuadrature, nextFn: NextMarsEasternQuadrature},
		{name: "MarsWesternQuadrature", seed: NextMarsWesternQuadrature(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastMarsWesternQuadrature, nextFn: NextMarsWesternQuadrature},
		{name: "MarsP2R", seed: NextMarsProgradeToRetrograde(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastMarsProgradeToRetrograde, nextFn: NextMarsProgradeToRetrograde},
		{name: "MarsR2P", seed: NextMarsRetrogradeToPrograde(ttjdUTC(2025, 1, 1, 0, 0, 0)), lastFn: LastMarsRetrogradeToPrograde, nextFn: NextMarsRetrogradeToPrograde},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			queryTT := TD2UT(tc.seed, true)
			last := tc.lastFn(queryTT)
			next := tc.nextFn(queryTT)
			if !sameEventJD(last, tc.seed) {
				t.Fatalf("last exact boundary mismatch: got %.12f want %.12f", last, tc.seed)
			}
			if !sameEventJD(next, tc.seed) {
				t.Fatalf("next exact boundary mismatch: got %.12f want %.12f", next, tc.seed)
			}
		})
	}
}

func TestOuterPlanetNextEventAdvancesPastReturnedEvent(t *testing.T) {
	cases := []struct {
		name string
		seed float64
		next func(float64) float64
	}{
		{name: "MarsOpposition", seed: ttjdUTC(2026, 1, 1, 0, 0, 0), next: NextMarsOpposition},
		{name: "MarsP2R", seed: ttjdUTC(2026, 1, 1, 0, 0, 0), next: NextMarsProgradeToRetrograde},
		{name: "JupiterOpposition", seed: ttjdUTC(2026, 1, 1, 0, 0, 0), next: NextJupiterOpposition},
		{name: "SaturnConjunction", seed: ttjdUTC(2026, 1, 1, 0, 0, 0), next: NextSaturnConjunction},
		{name: "UranusP2R", seed: ttjdUTC(2026, 1, 1, 0, 0, 0), next: NextUranusProgradeToRetrograde},
		{name: "NeptuneR2P", seed: ttjdUTC(2026, 1, 1, 0, 0, 0), next: NextNeptuneRetrogradeToPrograde},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			first := tc.next(tc.seed)
			query := TD2UT(Date2JDE(JDE2DateByZone(first, time.UTC, false).Add(time.Second)), true)
			next := tc.next(query)
			if !eventUTQueryAfterOrEqual(next, query) {
				t.Fatalf("next should be after query: first=%.12f query=%.12f next=%.12f", first, query, next)
			}
			if sameEventJD(next, first) {
				t.Fatalf("next should advance past first event: first=%.12f next=%.12f", first, next)
			}
		})
	}
}

func ttjdUTC(year, month, day, hour, min, sec int) float64 {
	return TD2UT(Date2JDE(time.Date(year, time.Month(month), day, hour, min, sec, 0, time.UTC)), true)
}
