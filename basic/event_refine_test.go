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

// 48 次迭代用尽后括号仍宽于容差时不能谎报"已抛光"：调用方把 ok 当作精度保证。
func TestEventBracketSecantRootRejectsUnpolishedRoot(t *testing.T) {
	linear := func(x float64) float64 { return x - 0.5 }
	if _, ok := eventBracketSecantRoot(0, 1, -0.5, 0.5, 1e-9, linear); !ok {
		t.Fatalf("normal tolerance should return ok=true")
	}
	// 无理根（sqrt(2)）永远取不到精确零，容差又低于该量级的 ULP：必须报 false。
	irrational := func(x float64) float64 { return x*x - 2 }
	if _, ok := eventBracketSecantRoot(0, 2, -2, 2, 1e-25, irrational); ok {
		t.Fatalf("unreachable tolerance must not report a polished root")
	}
	if _, ok := eventBracketSecantRoot(0, 2, -2, 2, 1e-12, irrational); !ok {
		t.Fatalf("reachable tolerance should still report ok=true")
	}
}
