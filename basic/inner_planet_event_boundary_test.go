package basic

import "testing"

func TestInnerPlanetExactEventBoundaryIncludesCurrent(t *testing.T) {
	cases := []struct {
		name   string
		seed   float64
		lastFn func(float64) float64
		nextFn func(float64) float64
	}{
		{name: "MercuryConjunction", seed: NextMercuryConjunction(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastMercuryConjunction, nextFn: NextMercuryConjunction},
		{name: "MercuryInferior", seed: NextMercuryInferiorConjunctionInclusive(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastMercuryInferiorConjunctionInclusive, nextFn: NextMercuryInferiorConjunctionInclusive},
		{name: "MercurySuperior", seed: NextMercurySuperiorConjunctionInclusive(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastMercurySuperiorConjunctionInclusive, nextFn: NextMercurySuperiorConjunctionInclusive},
		{name: "MercuryRetrograde", seed: NextMercuryRetrogradeInclusive(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastMercuryRetrogradeInclusive, nextFn: NextMercuryRetrogradeInclusive},
		{name: "MercuryP2R", seed: NextMercuryProgradeToRetrogradeInclusive(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastMercuryProgradeToRetrogradeInclusive, nextFn: NextMercuryProgradeToRetrogradeInclusive},
		{name: "MercuryR2P", seed: NextMercuryRetrogradeToProgradeInclusive(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastMercuryRetrogradeToProgradeInclusive, nextFn: NextMercuryRetrogradeToProgradeInclusive},
		{name: "MercuryGreatestElongation", seed: NextMercuryGreatestElongationInclusive(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastMercuryGreatestElongationInclusive, nextFn: NextMercuryGreatestElongationInclusive},
		{name: "MercuryEastElongation", seed: NextMercuryGreatestElongationEastInclusive(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastMercuryGreatestElongationEastInclusive, nextFn: NextMercuryGreatestElongationEastInclusive},
		{name: "MercuryWestElongation", seed: NextMercuryGreatestElongationWestInclusive(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastMercuryGreatestElongationWestInclusive, nextFn: NextMercuryGreatestElongationWestInclusive},
		{name: "VenusConjunction", seed: NextVenusConjunction(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastVenusConjunction, nextFn: NextVenusConjunction},
		{name: "VenusInferior", seed: NextVenusInferiorConjunctionInclusive(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastVenusInferiorConjunctionInclusive, nextFn: NextVenusInferiorConjunctionInclusive},
		{name: "VenusSuperior", seed: NextVenusSuperiorConjunctionInclusive(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastVenusSuperiorConjunctionInclusive, nextFn: NextVenusSuperiorConjunctionInclusive},
		{name: "VenusRetrograde", seed: NextVenusRetrogradeInclusive(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastVenusRetrogradeInclusive, nextFn: NextVenusRetrogradeInclusive},
		{name: "VenusP2R", seed: NextVenusProgradeToRetrogradeInclusive(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastVenusProgradeToRetrogradeInclusive, nextFn: NextVenusProgradeToRetrogradeInclusive},
		{name: "VenusR2P", seed: NextVenusRetrogradeToProgradeInclusive(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastVenusRetrogradeToProgradeInclusive, nextFn: NextVenusRetrogradeToProgradeInclusive},
		{name: "VenusGreatestElongation", seed: NextVenusGreatestElongationInclusive(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastVenusGreatestElongationInclusive, nextFn: NextVenusGreatestElongationInclusive},
		{name: "VenusEastElongation", seed: NextVenusGreatestElongationEastInclusive(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastVenusGreatestElongationEastInclusive, nextFn: NextVenusGreatestElongationEastInclusive},
		{name: "VenusWestElongation", seed: NextVenusGreatestElongationWestInclusive(ttjdUTC(2026, 1, 1, 0, 0, 0)), lastFn: LastVenusGreatestElongationWestInclusive, nextFn: NextVenusGreatestElongationWestInclusive},
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
