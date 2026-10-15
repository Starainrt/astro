package basic

import (
	"math"

	"github.com/starainrt/astro/internal/geodata"
)

// SkyBody 通用天象搜索支持的天体 / body for generic sky event searches.
type SkyBody int

const (
	// SkySun 太阳 / the Sun.
	SkySun SkyBody = iota
	// SkyMoon 月球 / the Moon.
	SkyMoon
	// SkyMercury 水星 / Mercury.
	SkyMercury
	// SkyVenus 金星 / Venus.
	SkyVenus
	// SkyMars 火星 / Mars.
	SkyMars
	// SkyJupiter 木星 / Jupiter.
	SkyJupiter
	// SkySaturn 土星 / Saturn.
	SkySaturn
	// SkyUranus 天王星 / Uranus.
	SkyUranus
	// SkyNeptune 海王星 / Neptune.
	SkyNeptune
)

// SkyEvent 通用天象事件：TT 儒略日与事件时的角量（度）/ generic sky event at TT JDE with its angle in degrees.
type SkyEvent struct {
	JDE   float64
	Value float64
}

func (body SkyBody) valid() bool { return body >= SkySun && body <= SkyNeptune }

// skyScanStep 保证相邻采样间视黄经变化远小于 180°，逆行回转也不会漏根。
func skyScanStep(bodies ...SkyBody) float64 {
	step := 2.0
	for _, b := range bodies {
		switch b {
		case SkyMoon:
			step = math.Min(step, 0.25)
		case SkyMercury, SkyVenus, SkySun:
			step = math.Min(step, 1.0)
		}
	}
	return step
}

// SkyBodyApparentLoBoWithInstance 地心视黄经、视黄纬（度）/ geocentric apparent ecliptic longitude and latitude.
func SkyBodyApparentLoBoWithInstance(body SkyBody, jde float64, inst *Instance) (float64, float64) {
	switch body {
	case SkySun:
		return sunApparentLoBoN(jde, -1, inst)
	case SkyMoon:
		return moonApparentLoBoN(jde, -1, inst)
	case SkyMercury, SkyVenus, SkyMars, SkyJupiter, SkySaturn, SkyUranus, SkyNeptune:
		return planetApparentLoBoWithInstance(int(body-SkyMercury)+1, jde, inst)
	}
	return math.NaN(), math.NaN()
}

// skyGridMaxIntervals 网格段数上限；须能放进 32 位 int，超出时扫描本身也不可能完成。
const skyGridMaxIntervals = 1 << 30

// skyGrid 把 [start, end] 均分为至少 minIntervals 段、每段不超过 step；按下标取点，浮点下必然前进且末点恰为 end。
func skyGrid(start, end, step float64, minIntervals int) (count int, at func(int) float64, ok bool) {
	span := end - start
	if !finite(start) || !finite(end) || !(span > 0) || !(step > 0) {
		return 0, nil, false
	}
	ratio := math.Ceil(span / step)
	if !(ratio < skyGridMaxIntervals) {
		return 0, nil, false
	}
	count = minIntervals
	if ratio > float64(count) {
		count = int(ratio)
	}
	return count, func(index int) float64 {
		if index == count {
			return end
		}
		return start + span*float64(index)/float64(count)
	}, true
}

// skyRoots 在闭区间 [start, end] 内按 step 扫描 f 的零点并二分到约 0.1 秒；f 须连续（已处理 360° 回绕）。
func skyRoots(start, end, step float64, f func(float64) float64) []float64 {
	count, at, ok := skyGrid(start, end, step, 1)
	if !ok {
		return nil
	}
	var roots []float64
	prevT, prev := start, f(start)
	if prev == 0 {
		roots = append(roots, start)
	}
	for index := 1; index <= count; index++ {
		t := at(index)
		cur := f(t)
		switch {
		case cur == 0:
			roots = append(roots, t)
		// 回绕跳变时 |差| 接近 360，不是真根；prev 为 0 的根已在上一点记过。
		case prev != 0 && !math.IsNaN(prev) && !math.IsNaN(cur) && (prev < 0) != (cur < 0) && math.Abs(cur-prev) < 180:
			lo, hi, flo := prevT, t, prev
			// 迭代上限防止大 JDE 下浮点间隔大于容差时不收敛。
			for i := 0; i < 100 && hi-lo > 1e-6; i++ {
				mid := (lo + hi) / 2
				fm := f(mid)
				if fm == 0 {
					lo, hi = mid, mid
					break
				}
				if (fm < 0) == (flo < 0) {
					lo, flo = mid, fm
				} else {
					hi = mid
				}
			}
			roots = append(roots, (lo+hi)/2)
		}
		prevT, prev = t, cur
	}
	return roots
}

// SkyLongitudeCrossingsWithInstance body 视黄经经过 longitude 的全部时刻 / times when the apparent longitude equals longitude.
func SkyLongitudeCrossingsWithInstance(body SkyBody, longitude, startJDE, endJDE float64, inst *Instance) []SkyEvent {
	if !body.valid() || !(endJDE > startJDE) {
		return nil
	}
	f := func(t float64) float64 {
		lo, _ := SkyBodyApparentLoBoWithInstance(body, t, inst)
		return geodata.NormalizeLongitude(lo - longitude)
	}
	var out []SkyEvent
	for _, t := range skyRoots(startJDE, endJDE, skyScanStep(body), f) {
		out = append(out, SkyEvent{JDE: t, Value: longitude})
	}
	return out
}

// SkyStationsWithInstance 视黄经留（速度过零）的时刻，Value 为留时视黄经 / longitude stations; Value is the longitude.
func SkyStationsWithInstance(body SkyBody, startJDE, endJDE float64, inst *Instance) []SkyEvent {
	if !body.valid() || body == SkySun || body == SkyMoon || !(endJDE > startJDE) {
		return nil
	}
	const h = 0.05
	rate := func(t float64) float64 {
		a, _ := SkyBodyApparentLoBoWithInstance(body, t-h, inst)
		b, _ := SkyBodyApparentLoBoWithInstance(body, t+h, inst)
		return geodata.NormalizeLongitude(b - a)
	}
	var out []SkyEvent
	for _, t := range skyRoots(startJDE, endJDE, skyScanStep(body), rate) {
		lo, _ := SkyBodyApparentLoBoWithInstance(body, t, inst)
		out = append(out, SkyEvent{JDE: t, Value: lo})
	}
	return out
}

// SkyAspectCrossingsWithInstance a 与 b 视黄经差（a − b）等于 angle 的时刻；angle 0 为黄经合 / times when lon(a) − lon(b) equals angle.
func SkyAspectCrossingsWithInstance(a, b SkyBody, angle, startJDE, endJDE float64, inst *Instance) []SkyEvent {
	if !a.valid() || !b.valid() || a == b || !(endJDE > startJDE) {
		return nil
	}
	f := func(t float64) float64 {
		la, _ := SkyBodyApparentLoBoWithInstance(a, t, inst)
		lb, _ := SkyBodyApparentLoBoWithInstance(b, t, inst)
		return geodata.NormalizeLongitude(la - lb - angle)
	}
	var out []SkyEvent
	for _, t := range skyRoots(startJDE, endJDE, skyScanStep(a, b), f) {
		out = append(out, SkyEvent{JDE: t, Value: angle})
	}
	return out
}

func skyAngularDistance(lon1, lat1, lon2, lat2 float64) float64 {
	v1, v2 := sphericalUnit(lon1, lat1), sphericalUnit(lon2, lat2)
	cross := [3]float64{v1[1]*v2[2] - v1[2]*v2[1], v1[2]*v2[0] - v1[0]*v2[2], v1[0]*v2[1] - v1[1]*v2[0]}
	return math.Atan2(math.Sqrt(cross[0]*cross[0]+cross[1]*cross[1]+cross[2]*cross[2]), v1[0]*v2[0]+v1[1]*v2[1]+v1[2]*v2[2]) / rad
}

// skyGoldenMin 黄金分割求 [lo, hi] 内单峰 f 的极小点。
func skyGoldenMin(lo, hi float64, f func(float64) float64) (float64, float64) {
	const g = 0.6180339887498949
	x1, x2 := hi-g*(hi-lo), lo+g*(hi-lo)
	fx1, fx2 := f(x1), f(x2)
	for i := 0; i < 200 && hi-lo > 1e-6; i++ {
		if fx1 < fx2 {
			hi, x2, fx2 = x2, x1, fx1
			x1 = hi - g*(hi-lo)
			fx1 = f(x1)
		} else {
			lo, x1, fx1 = x1, x2, fx2
			x2 = lo + g*(hi-lo)
			fx2 = f(x2)
		}
	}
	t := (lo + hi) / 2
	return t, f(t)
}

// skyMinima 扫描 f 在 [start, end] 内的局部极小并用黄金分割细化；只保留 ≤ maxValue 的极小。
func skyMinima(start, end, step, maxValue float64, f func(float64) float64) []SkyEvent {
	count, at, ok := skyGrid(start, end, step, 2)
	if !ok {
		return nil
	}
	var out []SkyEvent
	add := func(lo, hi, boundary float64) {
		// 端点段细化值须低于边界采样值，否则极小在窗口外，不是窗口内的局部极小。
		if t, v := skyGoldenMin(lo, hi, f); v <= maxValue && v < boundary {
			out = append(out, SkyEvent{JDE: t, Value: v})
		}
	}
	t0, t1 := at(0), at(1)
	f0, f1 := f(t0), f(t1)
	if f0 < f1 {
		add(t0, t1, f0)
	}
	for index := 2; index <= count; index++ {
		t2 := at(index)
		f2 := f(t2)
		if f1 <= f0 && f1 < f2 {
			add(t0, t2, math.Inf(1))
		}
		t0, t1, f0, f1 = t1, t2, f1, f2
	}
	if f1 < f0 {
		add(t0, t1, f1)
	}
	return out
}

// SkySeparationMinimaWithInstance a、b 地心视角距的局部极小（度），只返回 ≤ maxSeparation 的 / closest approaches within maxSeparation degrees.
func SkySeparationMinimaWithInstance(a, b SkyBody, maxSeparation, startJDE, endJDE float64, inst *Instance) []SkyEvent {
	if !a.valid() || !b.valid() || a == b || !(endJDE > startJDE) {
		return nil
	}
	f := func(t float64) float64 {
		la, ba := SkyBodyApparentLoBoWithInstance(a, t, inst)
		lb, bb := SkyBodyApparentLoBoWithInstance(b, t, inst)
		return skyAngularDistance(la, ba, lb, bb)
	}
	return skyMinima(startJDE, endJDE, skyScanStep(a, b)/2, maxSeparation, f)
}

// SkyStarSeparationMinimaWithInstance 天体与恒星地心视角距的局部极小（度）/ closest approaches of a body to a star within maxSeparation degrees.
func SkyStarSeparationMinimaWithInstance(body SkyBody, star StarCoordinate, maxSeparation, startJDE, endJDE float64, inst *Instance) []SkyEvent {
	if !body.valid() || !(endJDE > startJDE) {
		return nil
	}
	f := func(t float64) float64 {
		ra, dec := starApparentRaDecGeocentric(t, star, inst)
		sl, sb := RaDecToLoBo(t, ra, dec)
		lo, bo := SkyBodyApparentLoBoWithInstance(body, t, inst)
		return skyAngularDistance(lo, bo, sl, sb)
	}
	return skyMinima(startJDE, endJDE, skyScanStep(body)/2, maxSeparation, f)
}
