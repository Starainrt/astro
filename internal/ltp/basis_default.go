//go:build !astro_lowmem
// +build !astro_lowmem

package ltp

import "math"

var basisTable [basisSegs]basisSeg

// EclipticBasis 同 EclipticBasisExact，覆盖区间内用分段 Chebyshev 近似（差 < 1e-12 rad），区间外精确计算 / cached form of EclipticBasisExact.
func EclipticBasis(epj float64) (x, y, z Vector3) {
	u := (epj - basisStart) / basisSpan
	if !(u >= 0 && u < basisSegs) {
		return EclipticBasisExact(epj)
	}
	i := int(math.Floor(u))
	if i < 0 || i >= basisSegs {
		return EclipticBasisExact(epj)
	}
	seg := &basisTable[i]
	seg.once.Do(func() { seg.fit(basisStart + float64(i)*basisSpan) })
	s := 2*(u-float64(i)) - 1
	var v [6]float64
	for k := range v {
		v[k] = clenshaw(&seg.c[k], s)
	}
	x = Vector3{v[0], v[1], v[2]}
	z = Vector3{v[3], v[4], v[5]}
	return x, Cross(z, x), z
}
