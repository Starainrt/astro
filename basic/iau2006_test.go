// 被测契约：2000A 与默认 2000B 章动差在 1 mas 量级，单位换算正确。
package basic

import (
	"math"
	"testing"
)

func TestNutation2000ANearDefault(t *testing.T) {
	for _, jde := range []float64{2451545.0, 2460676.5, 2415020.0, 2488069.5} {
		a1, a2 := Nutation2000A(jde)
		b1, b2 := Nutation2000B(jde)
		if d := math.Max(math.Abs(a1-b1), math.Abs(a2-b2)) * 3600e3; d > 2 {
			t.Errorf("jde %.1f: 2000A vs 2000B differ by %.3f mas", jde, d)
		}
	}
	if d := math.Abs(MeanObliquity2006(2451545.0)*3600 - 84381.406); d > 1e-9 {
		t.Errorf("J2000 obliquity off by %g arcsec", d)
	}
	m := BiasPrecessionNutation2006(2460676.5)
	v := m.Apply([3]float64{1, 0, 0})
	if n := math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2]); math.Abs(n-1) > 1e-14 {
		t.Errorf("rotation not orthonormal: %g", n)
	}
}
