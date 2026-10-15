package semianalytic

import (
	"errors"
	"math"
	"sync"

	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/planet/internal/semianalytic/lunar"
	"github.com/starainrt/astro/planet/internal/semianalytic/planets"
)

const (
	cKMPerDay = 299792.458 * 86400
	// 太阳史瓦西半径 2GM☉/c²，AU。
	srsAU = 1.97412574336e-8
)

var errInvalidObserver = errors.New("semianalytic: non-finite observer state")

type observer struct {
	jd     float64
	p, v   [3]float64 // 地球日心位置 km、相对质心速度 km/日
	sunVel [3]float64 // 太阳相对质心速度 km/日
}

// 同一时刻常连续求多个天体，观测者状态缓存最近一次。
var (
	observerMu    sync.RWMutex
	observerCache observer
	observerSet   bool
)

func observerAt(jd float64) (observer, error) {
	observerMu.RLock()
	c, set := observerCache, observerSet
	observerMu.RUnlock()
	if set && c.jd == jd {
		return c, nil
	}
	e, err := helio(ephem.Earth, jd)
	if err != nil {
		return observer{}, err
	}
	// 太阳质心速度只进光行差与光行时修正：快速档、只计四颗巨行星，内行星贡献 < 0.1 mas。
	var sv [3]float64
	_, total := planets.GM(1)
	for id := 1; id <= 4; id++ {
		gm, _ := planets.GM(id)
		total += gm
	}
	for id := 5; id <= 8; id++ {
		gm, _ := planets.GM(id)
		_, v, err := planets.Fast().HeliocentricState(id, jd)
		if err != nil {
			return observer{}, err
		}
		for i := range sv {
			sv[i] -= gm * v[i]
		}
		total += gm
	}
	o := observer{jd: jd, p: e.p}
	for i := range sv {
		sv[i] /= total
		o.sunVel[i] = sv[i]
		o.v[i] = e.v[i] + sv[i]
		// 坏表读出的非有限速度不得进缓存，否则同一时刻会一直返回 NaN 方向而 err 为 nil。
		if !finite(o.v[i]) || !finite(o.p[i]) {
			return observer{}, errInvalidObserver
		}
	}
	observerMu.Lock()
	observerCache, observerSet = o, true
	observerMu.Unlock()
	return o, nil
}

// targetHelio 目标日心位置；fast 用快速档，只作光行时初值与几何距离。
func targetHelio(b ephem.Body, jd float64, fast bool) ([3]float64, error) {
	switch {
	case b == ephem.Sun:
		return [3]float64{}, nil
	case b == ephem.Moon:
		s, err := helio(ephem.Moon, jd)
		return s.p, err
	case b >= ephem.Mercury && b <= ephem.Neptune && b != ephem.EMB:
		if fast {
			return planets.Fast().Heliocentric(int(b), jd)
		}
		return planets.Default().Heliocentric(int(b), jd)
	}
	return [3]float64{}, ephem.ErrUnsupportedBody
}

type apparentEntry struct {
	jd        float64
	dir       [3]float64
	dist, tau float64
	set       bool
}

var (
	apparentMu    sync.RWMutex
	apparentCache [10]apparentEntry
)

func apparentStore(slot int, e apparentEntry) {
	e.set = true
	apparentMu.Lock()
	apparentCache[slot] = e
	apparentMu.Unlock()
}

// cacheSlot 天体对应的缓存槽；-1 表示不支持该天体。
func cacheSlot(b ephem.Body) int {
	switch {
	case b == ephem.Sun:
		return 0
	case b == ephem.Moon:
		return 9
	case b >= ephem.Mercury && b <= ephem.Neptune && b != ephem.EMB:
		return int(b)
	}
	return -1
}

// Apparent 地心视方向（ICRF 单位矢量）、几何地心距 km、光行时 日 / geocentric apparent direction.
func Apparent(target ephem.Body, jd float64) (dir [3]float64, distKM, tauDays float64, err error) {
	slot := cacheSlot(target)
	if slot < 0 {
		return dir, 0, 0, ephem.ErrUnsupportedBody
	}
	if !finite(jd) {
		return dir, 0, 0, ephem.ErrOutOfRange
	}
	apparentMu.RLock()
	c := apparentCache[slot]
	apparentMu.RUnlock()
	if c.set && c.jd == jd {
		return c.dir, c.dist, c.tau, nil
	}
	o, err := observerAt(jd)
	if err != nil {
		return dir, 0, 0, err
	}
	if target == ephem.Moon {
		e, err := apparentMoon(jd, o)
		if err != nil {
			return dir, 0, 0, err
		}
		apparentStore(slot, e)
		return e.dir, e.dist, e.tau, nil
	}
	now, err := targetHelio(target, jd, true)
	if err != nil {
		return dir, 0, 0, err
	}
	g := vsub(now, o.p)
	dist := vlen(g)
	tau := dist / cKMPerDay
	// 迭代光行时，每次误差缩小约 v/c 倍：先用快速档收敛 τ，末次用全表；太阳在 τ 内相对质心的位移按线性计入。
	var r [3]float64
	for k := 0; k < 2; k++ {
		if r, err = targetHelio(target, jd-tau, k == 0); err != nil {
			return dir, 0, 0, err
		}
		for i := range g {
			g[i] = r[i] - o.p[i] - o.sunVel[i]*tau
		}
		tau = vlen(g) / cKMPerDay
	}
	q := vscale(g, 1/vlen(g))
	if target != ephem.Sun && target != ephem.Moon {
		q = deflect(q, r, o.p)
	}
	q = aberrate(q, o.v)
	e := apparentEntry{jd: jd, dir: toICRF(q), dist: dist, tau: tau}
	apparentStore(slot, e)
	return e.dir, e.dist, e.tau, nil
}

// apparentMoon 月球地心视方向。光行时仅约 1.3 s，若像行星那样在 jd 与 jd−τ 各求一次地月质心
// 日心位置（~1.5e8 km）再与地球相减，远历元表求值的舍入噪声（2981 年约 2e-4 km）会原样进入
// 地心矢量，超过下游牛顿迭代的残差容差。这里全程用地心量：
// 月球日心(jd−τ) − 地球日心(jd) = EMB(jd−τ) − EMB(jd) + f·m(jd−τ) + m(jd)/(1+μ)，
// 其中 EMB 在 τ 内的位移按 jd 时刻速度线性外推，二阶项 ½aτ² < 1e-5 km。
// 月球地心位置同样按 jd 时刻解析速度外推（½aτ² ≈ 2e-6 km）：直接求 m(jd−τ) 时，jd−τ 的舍入
// （~4.7e-10 日）乘月球地心速度（~8.8e4 km/日）给出 ~1e-5 km 的逐点抖动，有限差分导数会把它放大。
func apparentMoon(jd float64, o observer) (apparentEntry, error) {
	emb, err := planetHelio(int(ephem.EMB), jd)
	if err != nil {
		return apparentEntry{}, err
	}
	now, vel := lunar.Default().PositionVelocityJ2000(jd)
	dist := vlen(now)
	tau := dist / cKMPerDay
	f := emrat / (1 + emrat)
	g := now
	for k := 0; k < 2; k++ {
		for i := range g {
			g[i] = f*(now[i]-vel[i]*tau) + now[i]/(1+emrat) - (emb.v[i]+o.sunVel[i])*tau
		}
		tau = vlen(g) / cKMPerDay
	}
	q := aberrate(vscale(g, 1/vlen(g)), o.v)
	return apparentEntry{jd: jd, dir: toICRF(q), dist: dist, tau: tau}, nil
}

// deflect 太阳引力偏折；p 目标方向，t 目标日心位置，e 地球日心位置。
func deflect(p, t, e [3]float64) [3]float64 {
	em := vlen(e) / auKM
	eu := vscale(e, 1/vlen(e))
	q := vscale(t, 1/vlen(t))
	qpe := vdot(q, eu) + 1
	if qpe < 1e-9 {
		return p
	}
	w := srsAU / em / qpe
	pe, pq := vdot(p, eu), vdot(p, q)
	// p × (e × q) = e(p·q) − q(p·e)
	for i := range p {
		p[i] += w * (eu[i]*pq - q[i]*pe)
	}
	return p
}

// aberrate 相对论光行差；v 观测者相对质心速度 km/日。
func aberrate(p, v [3]float64) [3]float64 {
	b := vscale(v, 1/cKMPerDay)
	bm := 1 / math.Sqrt(1-vdot(b, b))
	pdv := vdot(p, b)
	w1 := 1 + pdv/(1+1/bm)
	w2 := 1 / (1 + pdv)
	var r [3]float64
	for i := range p {
		r[i] = (p[i]/bm + w1*b[i]) * w2
	}
	return vscale(r, 1/vlen(r))
}

func vdot(a, b [3]float64) float64 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }
func vlen(a [3]float64) float64    { return math.Sqrt(vdot(a, a)) }
func vsub(a, b [3]float64) [3]float64 {
	return [3]float64{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
}
func vscale(a [3]float64, k float64) [3]float64 { return [3]float64{a[0] * k, a[1] * k, a[2] * k} }
