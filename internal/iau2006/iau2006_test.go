// 被测契约：章动、平黄赤交角、岁差与完整旋转矩阵和独立参考实现逐点一致。
package iau2006

import (
	"math"
	"testing"
)

type referenceCase struct {
	jd           float64
	dpsi, deps   float64
	obliquity    float64
	precession   Matrix
	fullRotation Matrix
}

func maxDiff(a, b Matrix) float64 {
	worst := 0.0
	for i := range a {
		for j := range a[i] {
			worst = math.Max(worst, math.Abs(a[i][j]-b[i][j]))
		}
	}
	return worst
}

func TestMatchesReference(t *testing.T) {
	if d := maxDiff(FrameBias(), referenceBias); d > 1e-15 {
		t.Errorf("frame bias differs by %.3g", d)
	}
	for _, c := range referenceCases {
		dpsi, deps := Nutation(c.jd)
		if math.Abs(dpsi-c.dpsi) > 1e-9 || math.Abs(deps-c.deps) > 1e-9 {
			t.Errorf("jd %.1f nutation %.12f %.12f want %.12f %.12f", c.jd, dpsi, deps, c.dpsi, c.deps)
		}
		if d := math.Abs(MeanObliquity(c.jd) - c.obliquity); d > 1e-9 {
			t.Errorf("jd %.1f obliquity differs by %.3g arcsec", c.jd, d)
		}
		if d := maxDiff(Precession(c.jd), c.precession); d > 1e-14 {
			t.Errorf("jd %.1f precession differs by %.3g", c.jd, d)
		}
		// 1e-12 rad ≈ 0.2 µas
		if d := maxDiff(BiasPrecessionNutation(c.jd), c.fullRotation); d > 1e-12 {
			t.Errorf("jd %.1f full rotation differs by %.3g", c.jd, d)
		}
	}
}

func TestMatrixIsRotation(t *testing.T) {
	m := BiasPrecessionNutation(2460676.5)
	v := m.Transpose().Apply(m.Apply([3]float64{0.3, -0.5, 0.8}))
	if math.Abs(v[0]-0.3)+math.Abs(v[1]+0.5)+math.Abs(v[2]-0.8) > 1e-14 {
		t.Errorf("round trip %v", v)
	}
}

func BenchmarkNutation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Nutation(2460676.5 + float64(i%100))
	}
}
