package basic

import (
	"math"
	"testing"
)

func TestEventZeroRefineFindsNearbyRoot(t *testing.T) {
	root := 0.123456789
	got := eventZeroRefine(root+0.00002, 0.0003, 1e-7, func(x float64) float64 {
		return x - root
	})
	if math.Abs(got-root) > 1e-12 {
		t.Fatalf("got %.15f want %.15f", got, root)
	}
}

func TestEventZeroRefineFallsBackToFixedScan(t *testing.T) {
	got := eventZeroRefine(0.1, 1, 0.1, func(x float64) float64 {
		return x*x + 1
	})
	if math.Abs(got) > 1e-12 {
		t.Fatalf("got %.15f want 0", got)
	}
}
