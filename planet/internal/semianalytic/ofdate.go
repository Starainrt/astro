package semianalytic

import (
	"math"
	"sync"

	"github.com/starainrt/astro/internal/ltp"
	"github.com/starainrt/astro/planet/internal/semianalytic/lunar"
	"github.com/starainrt/astro/planet/internal/semianalytic/planets"
	"github.com/starainrt/astro/tools"
)

const (
	auKM      = 149597870.7
	earthSlot = 0
	moonSlot  = 8
	fastBase  = 9
)

var sinEps0, cosEps0 = math.Sincos(eps0)

// xt 1..7 对应的表内天体号。
var bodyID = [8]int{0, planets.Mercury, planets.Venus, planets.Mars, planets.Jupiter, planets.Saturn, planets.Uranus, planets.Neptune}

type lbr struct{ jd, l, b, r float64 }

type lbrSlot struct {
	v   lbr
	set bool
}

// L、B、R 通常在同一时刻连续取用，每个天体每档缓存最近一次的结果；0..8 全表，9..17 快速档。
var (
	cacheMu sync.RWMutex
	cache   [18]lbrSlot
)

// tier n < 0 用全表，n ≥ 0 用快速档；新模型的快速档不随 n 细分。
func tier(n int) (*planets.Model, *lunar.Model, int) {
	if n < 0 {
		return planets.Default(), lunar.Default(), 0
	}
	return planets.Fast(), lunar.Fast(), fastBase
}

func cached(slot int, jd float64) (lbr, bool) {
	cacheMu.RLock()
	c := cache[slot]
	cacheMu.RUnlock()
	if c.set && c.v.jd == jd {
		return c.v, true
	}
	return lbr{}, false
}

func store(slot int, jd float64, v [3]float64) lbr {
	l, b, r := ofDate(v, jd)
	c := lbr{jd, l, b, r}
	cacheMu.Lock()
	cache[slot] = lbrSlot{v: c, set: true}
	cacheMu.Unlock()
	return c
}

type moonVec struct {
	jd  float64
	p   [3]float64
	set bool
}

// 地球合成与月球本身在同一时刻都要月球向量，按档缓存最近一次。
var (
	moonVecMu    sync.RWMutex
	moonVecCache [2]moonVec
)

func moonJ2000(lm *lunar.Model, slot int, jde float64) [3]float64 {
	moonVecMu.RLock()
	c := moonVecCache[slot]
	moonVecMu.RUnlock()
	if c.set && c.jd == jde {
		return c.p
	}
	p := lm.PositionJ2000(jde)
	moonVecMu.Lock()
	moonVecCache[slot] = moonVec{jd: jde, p: p, set: true}
	moonVecMu.Unlock()
	return p
}

// ofDate J2000 平黄道向量 → 历元平黄道经纬（弧度，经度在 [0, 2π)）与模长。
func ofDate(v [3]float64, jde float64) (l, b, r float64) {
	q0, q1, q2 := v[0], v[1]*cosEps0-v[2]*sinEps0, v[1]*sinEps0+v[2]*cosEps0
	x, y, z := ltp.EclipticBasis(2000 + (jde-2451545)/365.25)
	ex := x[0]*q0 + x[1]*q1 + x[2]*q2
	ey := y[0]*q0 + y[1]*q1 + y[2]*q2
	ez := z[0]*q0 + z[1]*q1 + z[2]*q2
	l = math.Atan2(ey, ex)
	if l < 0 {
		l += 2 * math.Pi
	}
	return l, math.Atan2(ez, math.Hypot(ex, ey)), math.Sqrt(ex*ex + ey*ey + ez*ez)
}

// Planet 与 planet.WherePlanetN 同语义（xt 取 -1..7，zn 取 0..2），调用方负责越界检查 / same contract as planet.WherePlanetN.
func Planet(xt, zn int, jde float64, n int) float64 {
	pm, lm, base := tier(n)
	slot := base + earthSlot
	if xt > 0 {
		slot = base + xt
	}
	c, ok := cached(slot, jde)
	if !ok {
		var p [3]float64
		if xt <= 0 {
			emb, _ := pm.Heliocentric(planets.EMB, jde)
			p = planets.EarthFromEMB(emb, moonJ2000(lm, base/fastBase, jde))
		} else {
			p, _ = pm.Heliocentric(bodyID[xt], jde)
		}
		c = store(slot, jde, p)
	}
	switch zn {
	case 0:
		l := c.l * 180 / math.Pi
		if xt == 0 {
			l += 180
		}
		return tools.Limit360(l)
	case 1:
		if xt == 0 {
			return -c.b * 180 / math.Pi
		}
		return c.b * 180 / math.Pi
	}
	return c.r / auKM
}

// Moon 地心月球历元平黄道：coord 0 黄经、1 黄纬（弧度）、2 距离（km），n ≥ 0 用快速档，调用方负责越界检查 / geocentric Moon of date.
func Moon(coord int, jde float64, n int) float64 {
	_, lm, base := tier(n)
	slot := base + moonSlot
	c, ok := cached(slot, jde)
	if !ok {
		c = store(slot, jde, moonJ2000(lm, base/fastBase, jde))
	}
	switch coord {
	case 0:
		return c.l
	case 1:
		return c.b
	}
	return c.r
}
