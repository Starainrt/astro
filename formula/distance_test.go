package formula

import (
	"math"
	"testing"
)

// formula 只做对外转发，换算基元在 tools，这里只锁别名与转发不脱钩。
func TestDistanceDelegatesToTools(t *testing.T) {
	assertFormulaClose(t, "1 ly in pc", Distance(1, DistanceLightYear), 0.306601393786, 1e-12)
	assertFormulaClose(t, "1 pc identity", Distance(2.5, DistanceParsec), 2.5, 1e-15)
	assertFormulaClose(t, "aliases agree", float64(DistanceParsec), 0, 1e-15)
	if got := Distance(0, DistanceParsec); !math.IsNaN(got) {
		t.Fatalf("Distance(0) = %v, want NaN", got)
	}
	if got := Distance(1, DistanceUnit(200)); !math.IsNaN(got) {
		t.Fatalf("Distance(1, unknown unit) = %v, want NaN", got)
	}
}
