package ltp

import (
	"math"
	"sync"
)

// 分段 Chebyshev 缓存的覆盖与精度：最短周期项 157.87 年，200 年段 16 阶误差远小于 1e-12 rad。

const (
	basisStart  = -10200.0
	basisSpan   = 200.0
	basisSegs   = 102
	basisDegree = 16
)

type basisSeg struct {
	once sync.Once
	// 6 个分量（x 的 3 个、z 的 3 个）各 basisDegree+1 个系数，c0 已减半。
	c [6][basisDegree + 1]float64
}

var basisTable [basisSegs]basisSeg

// EclipticBasisExact 历元平黄道在 J2000 平赤道下的正交基：x 指向平春分点，z 为黄道极 / exact basis of the mean ecliptic of date.
func EclipticBasisExact(epj float64) (x, y, z Vector3) {
	z = EclipticPole(epj)
	x, _ = Normalize(Cross(EquatorPole(epj), z))
	return x, Cross(z, x), z
}

// EclipticBasis 同 EclipticBasisExact，覆盖区间内用分段 Chebyshev 近似（差 < 1e-12 rad），区间外精确计算 / cached form of EclipticBasisExact.
func EclipticBasis(epj float64) (x, y, z Vector3) {
	u := (epj - basisStart) / basisSpan
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

func (seg *basisSeg) fit(start float64) {
	const n = basisDegree + 1
	var f [n][6]float64
	for j := 0; j < n; j++ {
		s := math.Cos(math.Pi * (float64(j) + 0.5) / n)
		x, _, z := EclipticBasisExact(start + (s+1)/2*basisSpan)
		f[j] = [6]float64{x[0], x[1], x[2], z[0], z[1], z[2]}
	}
	for k := 0; k < 6; k++ {
		for m := 0; m < n; m++ {
			sum := 0.0
			for j := 0; j < n; j++ {
				sum += f[j][k] * math.Cos(math.Pi*float64(m)*(float64(j)+0.5)/n)
			}
			seg.c[k][m] = 2 * sum / n
		}
		seg.c[k][0] /= 2
	}
}

func clenshaw(c *[basisDegree + 1]float64, s float64) float64 {
	b1, b2 := 0.0, 0.0
	for m := basisDegree; m >= 1; m-- {
		b1, b2 = 2*s*b1-b2+c[m], b1
	}
	return s*b1 - b2 + c[0]
}
