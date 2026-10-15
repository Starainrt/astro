// 被测契约：缓存基与精确基在覆盖区间内相差 < 1e-12，区间外逐位相同。
package ltp

import (
	"math"
	"testing"
)

func TestEclipticBasisMatchesExact(t *testing.T) {
	worst := 0.0
	for epj := -10199.9; epj < 10199.9; epj += 3.137 {
		ax, ay, az := EclipticBasis(epj)
		ex, ey, ez := EclipticBasisExact(epj)
		for i := 0; i < 3; i++ {
			worst = math.Max(worst, math.Abs(ax[i]-ex[i]))
			worst = math.Max(worst, math.Abs(ay[i]-ey[i]))
			worst = math.Max(worst, math.Abs(az[i]-ez[i]))
		}
	}
	t.Logf("max component diff %.3g", worst)
	if worst > 1e-12 {
		t.Fatalf("cached basis differs by %.3g", worst)
	}
	for _, epj := range []float64{-10300, 10300} {
		ax, _, az := EclipticBasis(epj)
		ex, _, ez := EclipticBasisExact(epj)
		if ax != ex || az != ez {
			t.Fatalf("epj %v: outside the cache the exact basis must be used", epj)
		}
	}
}

func BenchmarkEclipticBasis(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EclipticBasis(2000 + float64(i%7300)*0.27)
	}
}

func BenchmarkEclipticBasisExact(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EclipticBasisExact(2000 + float64(i%7300)*0.27)
	}
}
