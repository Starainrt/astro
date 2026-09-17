package basic

import (
	"testing"
)

// 这些基准是"性能不回退"的度量口径：Nutation2000B 有界记忆表、升落上下文 stateAt 的分配、
// 以及两条真实调用链（月球升落、掩星升落回归）。改动这些热路径后请对比 benchmark 数字。
// These benchmarks are the regression yardstick for the hot paths changed here: the bounded
// Nutation2000B memo, the allocation behaviour of the rise/set context stateAt, and two real call
// chains (moon rise/set and the occultation rise/set regression case). Compare before/after numbers
// whenever those paths change.
func BenchmarkNutation2000B(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = Nutation2000B(2460310.5 + float64(i%97)*0.37)
	}
}

func BenchmarkTrueObliquity(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = TrueObliquity(2460310.5 + float64(i%97)*0.37)
	}
}

var benchMoonRise, benchMoonSet float64

func BenchmarkMoonRiseSetChain(b *testing.B) {
	jd := JDECalc(2023, 6, 21)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchMoonRise, _ = GetMoonRiseTime(jd, 116.4074, 39.9042, 8, 1, 0)
		benchMoonSet, _ = GetMoonSetTime(jd, 116.4074, 39.9042, 8, 1, 0)
	}
}

// Nutation2000B 有界记忆表：重复瞬时必须命中且数值逐位不变。
func TestNutationMemoHitsRepeatedInstants(t *testing.T) {
	resetNutationMemo()
	const jd = 2460310.5
	firstPsi, firstEps := Nutation2000B(jd)
	secondPsi, secondEps := Nutation2000B(jd)
	if firstPsi != secondPsi || firstEps != secondEps {
		t.Fatalf("memo changed the value: (%v,%v) vs (%v,%v)", firstPsi, firstEps, secondPsi, secondEps)
	}
	hits, misses := nutationMemoStats()
	if hits == 0 || misses == 0 {
		t.Fatalf("expected one miss and at least one hit, got hits=%d misses=%d", hits, misses)
	}
	// 与未走记忆表的实现逐位一致。
	directPsi, directEps := nutation2000BCompute(jd)
	if firstPsi != directPsi || firstEps != directEps {
		t.Fatalf("memoized value differs from direct computation")
	}
}
