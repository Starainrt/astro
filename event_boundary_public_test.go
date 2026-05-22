package astro_test

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/jupiter"
	"github.com/starainrt/astro/mars"
	"github.com/starainrt/astro/mercury"
	"github.com/starainrt/astro/neptune"
	"github.com/starainrt/astro/saturn"
	"github.com/starainrt/astro/uranus"
	"github.com/starainrt/astro/venus"
)

func TestPublicPlanetEventBoundaryIncludesCurrent(t *testing.T) {
	type eventFuncs struct {
		last func(time.Time) time.Time
		next func(time.Time) time.Time
	}

	cases := []struct {
		name    string
		eventUT float64
		funcs   eventFuncs
	}{
		{name: "MercuryConjunction", eventUT: basic.NextMercuryConjunction(eventBoundaryTT(2026)), funcs: eventFuncs{last: mercury.LastConjunction, next: mercury.NextConjunction}},
		{name: "MercuryInferior", eventUT: basic.NextMercuryInferiorConjunctionInclusive(eventBoundaryTT(2026)), funcs: eventFuncs{last: mercury.LastInferiorConjunction, next: mercury.NextInferiorConjunction}},
		{name: "MercurySuperior", eventUT: basic.NextMercurySuperiorConjunctionInclusive(eventBoundaryTT(2026)), funcs: eventFuncs{last: mercury.LastSuperiorConjunction, next: mercury.NextSuperiorConjunction}},
		{name: "MercuryRetrograde", eventUT: basic.NextMercuryRetrogradeInclusive(eventBoundaryTT(2026)), funcs: eventFuncs{last: mercury.LastRetrograde, next: mercury.NextRetrograde}},
		{name: "MercuryP2R", eventUT: basic.NextMercuryProgradeToRetrogradeInclusive(eventBoundaryTT(2026)), funcs: eventFuncs{last: mercury.LastProgradeToRetrograde, next: mercury.NextProgradeToRetrograde}},
		{name: "MercuryR2P", eventUT: basic.NextMercuryRetrogradeToProgradeInclusive(eventBoundaryTT(2026)), funcs: eventFuncs{last: mercury.LastRetrogradeToPrograde, next: mercury.NextRetrogradeToPrograde}},
		{name: "MercuryGreatestElongation", eventUT: basic.NextMercuryGreatestElongationInclusive(eventBoundaryTT(2026)), funcs: eventFuncs{last: mercury.LastGreatestElongation, next: mercury.NextGreatestElongation}},
		{name: "MercuryGreatestElongationEast", eventUT: basic.NextMercuryGreatestElongationEastInclusive(eventBoundaryTT(2026)), funcs: eventFuncs{last: mercury.LastGreatestElongationEast, next: mercury.NextGreatestElongationEast}},
		{name: "MercuryGreatestElongationWest", eventUT: basic.NextMercuryGreatestElongationWestInclusive(eventBoundaryTT(2026)), funcs: eventFuncs{last: mercury.LastGreatestElongationWest, next: mercury.NextGreatestElongationWest}},

		{name: "VenusConjunction", eventUT: basic.NextVenusConjunction(eventBoundaryTT(2026)), funcs: eventFuncs{last: venus.LastConjunction, next: venus.NextConjunction}},
		{name: "VenusInferior", eventUT: basic.NextVenusInferiorConjunctionInclusive(eventBoundaryTT(2026)), funcs: eventFuncs{last: venus.LastInferiorConjunction, next: venus.NextInferiorConjunction}},
		{name: "VenusSuperior", eventUT: basic.NextVenusSuperiorConjunctionInclusive(eventBoundaryTT(2026)), funcs: eventFuncs{last: venus.LastSuperiorConjunction, next: venus.NextSuperiorConjunction}},
		{name: "VenusRetrograde", eventUT: basic.NextVenusRetrogradeInclusive(eventBoundaryTT(2026)), funcs: eventFuncs{last: venus.LastRetrograde, next: venus.NextRetrograde}},
		{name: "VenusP2R", eventUT: basic.NextVenusProgradeToRetrogradeInclusive(eventBoundaryTT(2026)), funcs: eventFuncs{last: venus.LastProgradeToRetrograde, next: venus.NextProgradeToRetrograde}},
		{name: "VenusR2P", eventUT: basic.NextVenusRetrogradeToProgradeInclusive(eventBoundaryTT(2026)), funcs: eventFuncs{last: venus.LastRetrogradeToPrograde, next: venus.NextRetrogradeToPrograde}},
		{name: "VenusGreatestElongation", eventUT: basic.NextVenusGreatestElongationInclusive(eventBoundaryTT(2026)), funcs: eventFuncs{last: venus.LastGreatestElongation, next: venus.NextGreatestElongation}},
		{name: "VenusGreatestElongationEast", eventUT: basic.NextVenusGreatestElongationEastInclusive(eventBoundaryTT(2026)), funcs: eventFuncs{last: venus.LastGreatestElongationEast, next: venus.NextGreatestElongationEast}},
		{name: "VenusGreatestElongationWest", eventUT: basic.NextVenusGreatestElongationWestInclusive(eventBoundaryTT(2026)), funcs: eventFuncs{last: venus.LastGreatestElongationWest, next: venus.NextGreatestElongationWest}},

		{name: "MarsConjunction", eventUT: basic.NextMarsConjunction(eventBoundaryTT(2026)), funcs: eventFuncs{last: mars.LastConjunction, next: mars.NextConjunction}},
		{name: "MarsOpposition", eventUT: basic.NextMarsOpposition(eventBoundaryTT(2026)), funcs: eventFuncs{last: mars.LastOpposition, next: mars.NextOpposition}},
		{name: "MarsP2R", eventUT: basic.NextMarsProgradeToRetrograde(eventBoundaryTT(2026)), funcs: eventFuncs{last: mars.LastProgradeToRetrograde, next: mars.NextProgradeToRetrograde}},
		{name: "MarsR2P", eventUT: basic.NextMarsRetrogradeToPrograde(eventBoundaryTT(2025)), funcs: eventFuncs{last: mars.LastRetrogradeToPrograde, next: mars.NextRetrogradeToPrograde}},
		{name: "MarsEasternQuadrature", eventUT: basic.NextMarsEasternQuadrature(eventBoundaryTT(2026)), funcs: eventFuncs{last: mars.LastEasternQuadrature, next: mars.NextEasternQuadrature}},
		{name: "MarsWesternQuadrature", eventUT: basic.NextMarsWesternQuadrature(eventBoundaryTT(2026)), funcs: eventFuncs{last: mars.LastWesternQuadrature, next: mars.NextWesternQuadrature}},

		{name: "JupiterConjunction", eventUT: basic.NextJupiterConjunction(eventBoundaryTT(2026)), funcs: eventFuncs{last: jupiter.LastConjunction, next: jupiter.NextConjunction}},
		{name: "JupiterOpposition", eventUT: basic.NextJupiterOpposition(eventBoundaryTT(2026)), funcs: eventFuncs{last: jupiter.LastOpposition, next: jupiter.NextOpposition}},
		{name: "JupiterP2R", eventUT: basic.NextJupiterProgradeToRetrograde(eventBoundaryTT(2026)), funcs: eventFuncs{last: jupiter.LastProgradeToRetrograde, next: jupiter.NextProgradeToRetrograde}},
		{name: "JupiterR2P", eventUT: basic.NextJupiterRetrogradeToPrograde(eventBoundaryTT(2026)), funcs: eventFuncs{last: jupiter.LastRetrogradeToPrograde, next: jupiter.NextRetrogradeToPrograde}},
		{name: "JupiterEasternQuadrature", eventUT: basic.NextJupiterEasternQuadrature(eventBoundaryTT(2026)), funcs: eventFuncs{last: jupiter.LastEasternQuadrature, next: jupiter.NextEasternQuadrature}},
		{name: "JupiterWesternQuadrature", eventUT: basic.NextJupiterWesternQuadrature(eventBoundaryTT(2026)), funcs: eventFuncs{last: jupiter.LastWesternQuadrature, next: jupiter.NextWesternQuadrature}},

		{name: "SaturnOpposition", eventUT: basic.NextSaturnOpposition(eventBoundaryTT(2025)), funcs: eventFuncs{last: saturn.LastOpposition, next: saturn.NextOpposition}},
		{name: "SaturnP2R", eventUT: basic.NextSaturnProgradeToRetrograde(eventBoundaryTT(2025)), funcs: eventFuncs{last: saturn.LastProgradeToRetrograde, next: saturn.NextProgradeToRetrograde}},
		{name: "SaturnR2P", eventUT: basic.NextSaturnRetrogradeToPrograde(eventBoundaryTT(2025)), funcs: eventFuncs{last: saturn.LastRetrogradeToPrograde, next: saturn.NextRetrogradeToPrograde}},
		{name: "SaturnEasternQuadrature", eventUT: basic.NextSaturnEasternQuadrature(eventBoundaryTT(2025)), funcs: eventFuncs{last: saturn.LastEasternQuadrature, next: saturn.NextEasternQuadrature}},
		{name: "SaturnWesternQuadrature", eventUT: basic.NextSaturnWesternQuadrature(eventBoundaryTT(2025)), funcs: eventFuncs{last: saturn.LastWesternQuadrature, next: saturn.NextWesternQuadrature}},

		{name: "UranusOpposition", eventUT: basic.NextUranusOpposition(eventBoundaryTT(2025)), funcs: eventFuncs{last: uranus.LastOpposition, next: uranus.NextOpposition}},
		{name: "UranusP2R", eventUT: basic.NextUranusProgradeToRetrograde(eventBoundaryTT(2025)), funcs: eventFuncs{last: uranus.LastProgradeToRetrograde, next: uranus.NextProgradeToRetrograde}},
		{name: "UranusR2P", eventUT: basic.NextUranusRetrogradeToPrograde(eventBoundaryTT(2025)), funcs: eventFuncs{last: uranus.LastRetrogradeToPrograde, next: uranus.NextRetrogradeToPrograde}},
		{name: "UranusEasternQuadrature", eventUT: basic.NextUranusEasternQuadrature(eventBoundaryTT(2025)), funcs: eventFuncs{last: uranus.LastEasternQuadrature, next: uranus.NextEasternQuadrature}},
		{name: "UranusWesternQuadrature", eventUT: basic.NextUranusWesternQuadrature(eventBoundaryTT(2025)), funcs: eventFuncs{last: uranus.LastWesternQuadrature, next: uranus.NextWesternQuadrature}},

		{name: "NeptuneOpposition", eventUT: basic.NextNeptuneOpposition(eventBoundaryTT(2026)), funcs: eventFuncs{last: neptune.LastOpposition, next: neptune.NextOpposition}},
		{name: "NeptuneP2R", eventUT: basic.NextNeptuneProgradeToRetrograde(eventBoundaryTT(2026)), funcs: eventFuncs{last: neptune.LastProgradeToRetrograde, next: neptune.NextProgradeToRetrograde}},
		{name: "NeptuneR2P", eventUT: basic.NextNeptuneRetrogradeToPrograde(eventBoundaryTT(2026)), funcs: eventFuncs{last: neptune.LastRetrogradeToPrograde, next: neptune.NextRetrogradeToPrograde}},
		{name: "NeptuneEasternQuadrature", eventUT: basic.NextNeptuneEasternQuadrature(eventBoundaryTT(2026)), funcs: eventFuncs{last: neptune.LastEasternQuadrature, next: neptune.NextEasternQuadrature}},
		{name: "NeptuneWesternQuadrature", eventUT: basic.NextNeptuneWesternQuadrature(eventBoundaryTT(2026)), funcs: eventFuncs{last: neptune.LastWesternQuadrature, next: neptune.NextWesternQuadrature}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eventTime := basic.JDE2DateByZone(tc.eventUT, time.UTC, false)
			assertSameEventTime(t, "last", tc.funcs.last(eventTime), eventTime)
			assertSameEventTime(t, "next", tc.funcs.next(eventTime), eventTime)
		})
	}
}

func eventBoundaryTT(year int) float64 {
	return basic.TD2UT(basic.Date2JDE(time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)), true)
}

func assertSameEventTime(t *testing.T, name string, got, want time.Time) {
	t.Helper()
	if math.Abs(got.Sub(want).Seconds()) > 1.0 {
		t.Fatalf("%s boundary mismatch: got %s want %s delta %.3fs", name, got, want, got.Sub(want).Seconds())
	}
}
