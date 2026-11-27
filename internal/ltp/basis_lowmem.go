//go:build astro_lowmem
// +build astro_lowmem

package ltp

import (
	"math"
	"sync"
)

// lowmem 下只保留最近用到的几段；段一经拟合不再改写，换出时读者手里的旧段仍可安全使用。
const basisCacheSlots = 4

var (
	basisCacheMu sync.Mutex
	basisCache   [basisCacheSlots]struct {
		index int
		seg   *basisSeg
	}
	basisCacheNext int
)

func basisSegment(i int) *basisSeg {
	basisCacheMu.Lock()
	defer basisCacheMu.Unlock()
	for _, slot := range basisCache {
		if slot.seg != nil && slot.index == i {
			return slot.seg
		}
	}
	seg := new(basisSeg)
	seg.fit(basisStart + float64(i)*basisSpan)
	basisCache[basisCacheNext].index, basisCache[basisCacheNext].seg = i, seg
	basisCacheNext = (basisCacheNext + 1) % basisCacheSlots
	return seg
}

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
	seg := basisSegment(i)
	s := 2*(u-float64(i)) - 1
	var v [6]float64
	for k := range v {
		v[k] = clenshaw(&seg.c[k], s)
	}
	x = Vector3{v[0], v[1], v[2]}
	z = Vector3{v[3], v[4], v[5]}
	return x, Cross(z, x), z
}
