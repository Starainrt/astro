package ephem

import "math"

const (
	genericCKMPerDay = 299792.458 * 86400
	genericAUKM      = 149597870.7
	// 太阳史瓦西半径 2GM☉/c²，AU。
	genericSunSchwarzschildAU = 1.97412574336e-8
)

// genericApparent 用 Engine.State 求地心视方向：迭代光行时、太阳引力偏折、相对论光行差，均在 ICRF 中进行。
// 引擎不能给出相对太阳系质心的地球、太阳或目标状态时返回 ErrUnsupportedApparent。
func genericApparent(engine Engine, target Body, jdTT float64) (dir [3]float64, distKM, tauDays float64, err error) {
	earth, earthVel, err := engine.State(Earth, SSB, jdTT)
	if err != nil {
		return dir, 0, 0, ErrUnsupportedApparent
	}
	now, _, err := engine.State(target, SSB, jdTT)
	if err != nil {
		return dir, 0, 0, ErrUnsupportedApparent
	}
	g := genericSub(now, earth)
	distKM = genericLen(g)
	if !(distKM > 0) || math.IsInf(distKM, 0) {
		return dir, 0, 0, ErrUnsupportedApparent
	}
	tauDays = distKM / genericCKMPerDay
	var retarded [3]float64
	for k := 0; k < 3; k++ {
		if retarded, _, err = engine.State(target, SSB, jdTT-tauDays); err != nil {
			return dir, 0, 0, ErrUnsupportedApparent
		}
		g = genericSub(retarded, earth)
		tauDays = genericLen(g) / genericCKMPerDay
	}
	q := genericScale(g, 1/genericLen(g))
	if target != Sun && target != Moon && target != Earth {
		sunThen, _, sunErr := engine.State(Sun, SSB, jdTT-tauDays)
		sunNow, _, sunNowErr := engine.State(Sun, SSB, jdTT)
		if sunErr == nil && sunNowErr == nil {
			q = genericDeflect(q, genericSub(retarded, sunThen), genericSub(earth, sunNow))
		}
	}
	return genericAberrate(q, earthVel), distKM, tauDays, nil
}

// genericDeflect 太阳引力偏折；p 目标方向，t 目标日心位置，e 地球日心位置。
func genericDeflect(p, t, e [3]float64) [3]float64 {
	el, tl := genericLen(e), genericLen(t)
	if el <= 0 || tl <= 0 {
		return p
	}
	eu, qu := genericScale(e, 1/el), genericScale(t, 1/tl)
	qpe := genericDot(qu, eu) + 1
	if qpe < 1e-9 {
		return p
	}
	w := genericSunSchwarzschildAU / (el / genericAUKM) / qpe
	pe, pq := genericDot(p, eu), genericDot(p, qu)
	for i := range p {
		p[i] += w * (eu[i]*pq - qu[i]*pe)
	}
	return genericScale(p, 1/genericLen(p))
}

// genericAberrate 相对论光行差；v 观测者相对质心速度 km/日。
func genericAberrate(p, v [3]float64) [3]float64 {
	b := genericScale(v, 1/genericCKMPerDay)
	bm := 1 / math.Sqrt(1-genericDot(b, b))
	pdv := genericDot(p, b)
	w1 := 1 + pdv/(1+1/bm)
	w2 := 1 / (1 + pdv)
	var r [3]float64
	for i := range p {
		r[i] = (p[i]/bm + w1*b[i]) * w2
	}
	return genericScale(r, 1/genericLen(r))
}

func genericDot(a, b [3]float64) float64 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }
func genericLen(a [3]float64) float64    { return math.Sqrt(genericDot(a, a)) }
func genericSub(a, b [3]float64) [3]float64 {
	return [3]float64{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
}
func genericScale(a [3]float64, k float64) [3]float64 {
	return [3]float64{a[0] * k, a[1] * k, a[2] * k}
}
