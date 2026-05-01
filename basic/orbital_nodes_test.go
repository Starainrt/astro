package basic

import (
	"math"
	"testing"
)

func TestOrbitalNodesDescendingOpposesAscending(t *testing.T) {
	jde := 2461157.5
	testCases := []struct {
		name       string
		ascending  func(float64) float64
		descending func(float64) float64
	}{
		{name: "moon", ascending: MoonAscendingNode, descending: MoonDescendingNode},
		{name: "mercury", ascending: MercuryAscendingNode, descending: MercuryDescendingNode},
		{name: "venus", ascending: VenusAscendingNode, descending: VenusDescendingNode},
		{name: "mars", ascending: MarsAscendingNode, descending: MarsDescendingNode},
		{name: "jupiter", ascending: JupiterAscendingNode, descending: JupiterDescendingNode},
		{name: "saturn", ascending: SaturnAscendingNode, descending: SaturnDescendingNode},
		{name: "uranus", ascending: UranusAscendingNode, descending: UranusDescendingNode},
		{name: "neptune", ascending: NeptuneAscendingNode, descending: NeptuneDescendingNode},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ascending := tc.ascending(jde)
			descending := tc.descending(jde)
			want := ascending + 180
			if want >= 360 {
				want -= 360
			}
			if diff := angularDifference(descending, ascending+180); diff > 1e-10 {
				t.Fatalf("descending node mismatch: got %.12f want %.12f diff=%.12g", descending, want, diff)
			}
		})
	}
}

func TestOrbitalNodesNFullMatchesDefault(t *testing.T) {
	jde := 2461157.5
	testCases := []struct {
		name       string
		defaultFn  func(float64) float64
		truncatedN func(float64, int) float64
	}{
		{name: "moon", defaultFn: MoonAscendingNode, truncatedN: MoonAscendingNodeN},
		{name: "mercury", defaultFn: MercuryAscendingNode, truncatedN: MercuryAscendingNodeN},
		{name: "venus", defaultFn: VenusAscendingNode, truncatedN: VenusAscendingNodeN},
		{name: "mars", defaultFn: MarsAscendingNode, truncatedN: MarsAscendingNodeN},
		{name: "jupiter", defaultFn: JupiterAscendingNode, truncatedN: JupiterAscendingNodeN},
		{name: "saturn", defaultFn: SaturnAscendingNode, truncatedN: SaturnAscendingNodeN},
		{name: "uranus", defaultFn: UranusAscendingNode, truncatedN: UranusAscendingNodeN},
		{name: "neptune", defaultFn: NeptuneAscendingNode, truncatedN: NeptuneAscendingNodeN},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.defaultFn(jde)
			gotN := tc.truncatedN(jde, -1)
			if diff := angularDifference(got, gotN); diff > 1e-10 {
				t.Fatalf("full-series N mismatch: got %.12f want %.12f diff=%.12g", gotN, got, diff)
			}
		})
	}
}

func angularDifference(a, b float64) float64 {
	diff := math.Mod(a-b, 360)
	if diff < -180 {
		diff += 360
	}
	if diff > 180 {
		diff -= 360
	}
	return math.Abs(diff)
}
