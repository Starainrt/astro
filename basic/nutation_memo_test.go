package basic

import (
	"math"
	"testing"
)

// 记忆表必须逐位精确：命中返回值与重新展开 77 项级数完全一致，且相邻 1 ULP 的输入互不干扰。
// The memo must be bit-exact: a hit returns exactly what the 77-term series would recompute, and
// inputs one ULP apart keep independent entries.
func TestNutationMemoIsBitExact(t *testing.T) {
	values := []float64{
		2451545.0, 2460310.5, 0, -1000000.5, -4713.5, 5373484.5,
		math.NaN(), math.Inf(1), math.Inf(-1), math.SmallestNonzeroFloat64,
	}
	for index := 0; index < 5000; index++ {
		values = append(values, 1000000.0+float64(index)*0.37)
	}
	for _, jd := range values {
		memoPsi, memoEps := Nutation2000B(jd)
		directPsi, directEps := nutation2000BCompute(jd)
		if math.Float64bits(memoPsi) != math.Float64bits(directPsi) ||
			math.Float64bits(memoEps) != math.Float64bits(directEps) {
			t.Fatalf("jd=%v: memo (%v,%v) != direct (%v,%v)", jd, memoPsi, memoEps, directPsi, directEps)
		}
		againPsi, againEps := Nutation2000B(jd)
		if math.Float64bits(againPsi) != math.Float64bits(memoPsi) ||
			math.Float64bits(againEps) != math.Float64bits(memoEps) {
			t.Fatalf("jd=%v: repeated call changed the value", jd)
		}
		// 相邻 1 ULP 的输入必须各自独立（不能被同一槽位合并）。
		neighbour := math.Nextafter(jd, math.Inf(1))
		if math.IsNaN(jd) || math.IsInf(jd, 0) {
			continue
		}
		nearPsi, _ := Nutation2000B(neighbour)
		nearDirect, _ := nutation2000BCompute(neighbour)
		if math.Float64bits(nearPsi) != math.Float64bits(nearDirect) {
			t.Fatalf("jd=%v neighbour: memo (%v) != direct (%v)", jd, nearPsi, nearDirect)
		}
	}
}

// 章动只依赖传入的 jd：ΔT 覆盖不改变它，因此记忆表不需要世代戳（与视恒星时不同）。
func TestNutationMemoIgnoresDeltaTGeneration(t *testing.T) {
	original := GetDeltaTFn()
	defer SetDeltaTFn(original)
	jd := 2460310.5
	beforePsi, beforeEps := Nutation2000B(jd)
	SetDeltaTFn(func(date float64, isJd bool) float64 { return 6000 })
	afterPsi, afterEps := Nutation2000B(jd)
	if math.Float64bits(beforePsi) != math.Float64bits(afterPsi) ||
		math.Float64bits(beforeEps) != math.Float64bits(afterEps) {
		t.Fatalf("nutation changed with the ΔT generation: (%v,%v) vs (%v,%v)", beforePsi, beforeEps, afterPsi, afterEps)
	}
}
