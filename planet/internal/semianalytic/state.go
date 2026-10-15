// Package semianalytic 内置半解析模型：VSOP2013/TOP2013 与 MPP02 结构、按 DE441 重拟的表，提供历元 L/B/R 与 ICRF 几何状态。
package semianalytic

import (
	"math"

	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/planet/internal/semianalytic/lunar"
	"github.com/starainrt/astro/planet/internal/semianalytic/planets"
)

// 表以 TT 为自变量拟合；TDB−TT 小于 1.7 ms，直接当作 TDB 使用。
const (
	coverStart = -1931455.0
	coverEnd   = 5373545.0
	eps0       = 84381.448 / 3600 * math.Pi / 180
	emrat      = 81.3005682214972
)

// Ephemeris 内置表；越过覆盖范围仍外推求值，只在 Coverage 中报告范围 / built-in tables; extrapolates beyond coverage.
type Ephemeris struct{}

var _ ephem.Engine = Ephemeris{}

// Apparent 地心视方向（ICRF 单位矢量）、几何地心距 km、光行时日 / apparent geocentric direction.
func (Ephemeris) Apparent(target ephem.Body, jdTT float64) (dir [3]float64, distKM, tauDays float64, err error) {
	return Apparent(target, jdTT)
}

func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }

type state struct{ p, v [3]float64 }

func (a state) sub(b state) state {
	for i := range a.p {
		a.p[i] -= b.p[i]
		a.v[i] -= b.v[i]
	}
	return a
}

// State 目标相对中心的 ICRF 几何状态 / geometric ICRF state of target relative to center.
func (Ephemeris) State(target, center ephem.Body, jdTDB float64) (pos, vel [3]float64, err error) {
	if !finite(jdTDB) {
		return pos, vel, ephem.ErrOutOfRange
	}
	t, err := helio(target, jdTDB)
	if err != nil {
		return pos, vel, err
	}
	c, err := helio(center, jdTDB)
	if err != nil {
		return pos, vel, err
	}
	d := t.sub(c)
	return toICRF(d.p), toICRF(d.v), nil
}

// Coverage 可信区间 / trusted time span.
func (Ephemeris) Coverage(target ephem.Body) (start, end float64, ok bool) {
	if _, err := helio(target, 2451545); err != nil {
		return 0, 0, false
	}
	return coverStart, coverEnd, true
}

// helio 日心 J2000 平黄道状态（km、km/日）。
func helio(b ephem.Body, jd float64) (state, error) {
	switch b {
	case ephem.Sun:
		return state{}, nil
	case ephem.SSB:
		return ssbHelio(jd), nil
	case ephem.Earth, ephem.Moon:
		emb, err := planetHelio(int(ephem.EMB), jd)
		if err != nil {
			return state{}, err
		}
		mp, mv := lunar.Default().PositionVelocityJ2000(jd)
		k := 1.0 / (1 + emrat)
		if b == ephem.Moon {
			k -= 1
		}
		for i := range mp {
			emb.p[i] -= mp[i] * k
			emb.v[i] -= mv[i] * k
		}
		return emb, nil
	}
	if b >= ephem.Mercury && b <= ephem.Neptune {
		return planetHelio(int(b), jd)
	}
	return state{}, ephem.ErrUnsupportedBody
}

func planetHelio(id int, jd float64) (state, error) {
	p, v, err := planets.Default().HeliocentricState(id, jd)
	return state{p, v}, err
}

// ssbHelio 太阳系质心的日心状态：Σ GMᵢ·rᵢ / (GM☉ + Σ GMᵢ)，冥王星等小天体不计。
func ssbHelio(jd float64) state {
	var s state
	_, gmSun := planets.GM(1)
	total := gmSun
	for id := 1; id <= 8; id++ {
		gm, _ := planets.GM(id)
		h, _ := planetHelio(id, jd)
		for i := range s.p {
			s.p[i] += gm * h.p[i]
			s.v[i] += gm * h.v[i]
		}
		total += gm
	}
	for i := range s.p {
		s.p[i] /= total
		s.v[i] /= total
	}
	return s
}

// toICRF J2000 平黄道 → ICRF 赤道（绕 x 轴转 −ε0）。
func toICRF(e [3]float64) [3]float64 {
	s, c := math.Sincos(eps0)
	return [3]float64{e[0], e[1]*c - e[2]*s, e[1]*s + e[2]*c}
}
