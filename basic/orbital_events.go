package basic

import "math"

// OrbitalEventKind 轨道事件类型 / orbital event kind.
type OrbitalEventKind int

const (
	OrbitalPerihelion     OrbitalEventKind = iota // 近日点 / perihelion
	OrbitalAphelion                               // 远日点 / aphelion
	OrbitalAscendingNode                          // 升交点 / ascending node
	OrbitalDescendingNode                         // 降交点 / descending node
)

// OrbitalEvent 轨道事件：TT 儒略日与日心距（AU）/ orbital event at TT JDE with heliocentric distance in AU.
type OrbitalEvent struct {
	JDE        float64
	Kind       OrbitalEventKind
	DistanceAU float64
}

// 恒星周期（日），只用于定采样步长和支配窗口。
var orbitalPeriodDays = map[SkyBody]float64{
	SkyMercury: 87.969, SkyVenus: 224.701, SkyMars: 686.98, SkyJupiter: 4332.59,
	SkySaturn: 10759.22, SkyUranus: 30688.5, SkyNeptune: 60182,
}

// SkyEarth 地球，只用于近日点、远日点查询 / Earth, valid only for apsis queries.
const SkyEarth SkyBody = -1

func orbitalIndex(body SkyBody) (int, float64, bool) {
	if body == SkyEarth {
		return -1, 365.256, true
	}
	period, ok := orbitalPeriodDays[body]
	if !ok {
		return 0, 0, false
	}
	return int(body-SkyMercury) + 1, period, true
}

// PlanetApsidesWithInstance 行星或地球（SkyEarth）在 [startJDE, endJDE] 内的近日点与远日点（TT）/ perihelia and aphelia in the TT window.
//
// 极值取日心距本身；太阳绕质心的反射运动会让天王星、海王星的日心距在近日点附近出现多个小极值，
// 这里只保留前后 0.3 个周期内最深的一个，与吻切轨道近日点可相差数年。
func PlanetApsidesWithInstance(body SkyBody, startJDE, endJDE float64, inst *Instance) []OrbitalEvent {
	index, period, ok := orbitalIndex(body)
	if !ok || !finite(startJDE) || !finite(endJDE) || endJDE < startJDE {
		return nil
	}
	radius := func(t float64) float64 { return PlanetCoordinateNWithInstance(index, 2, t, -1, inst) }
	step := period / 200
	from, to := startJDE-period/2, endJDE+period/2
	// 与 skyGrid 同一上限，同时挡住 to-from 溢出为 Inf。
	ratio := math.Ceil((to - from) / step)
	if !(ratio < skyGridMaxIntervals) {
		return nil
	}
	n := int(ratio) + 1
	window := int(math.Ceil(0.3 * period / step))
	ring := make([]float64, 2*window+1)
	sample := func(j int) float64 { return ring[j%len(ring)] }
	var out []OrbitalEvent
	check := func(i int) {
		for _, aphelion := range []bool{false, true} {
			lo, hi := i-window, i+window
			if lo < 0 {
				lo = 0
			}
			if hi > n-1 {
				hi = n - 1
			}
			if !orbitalDominant(sample, i, lo, hi, aphelion) {
				continue
			}
			center := from + float64(i)*step
			t, r := orbitalGoldenExtremum(radius, center-step, center+step, aphelion)
			if t < startJDE || t > endJDE {
				continue
			}
			kind := OrbitalPerihelion
			if aphelion {
				kind = OrbitalAphelion
			}
			out = append(out, OrbitalEvent{JDE: t, Kind: kind, DistanceAU: r})
		}
	}
	for k := 0; k < n; k++ {
		ring[k%len(ring)] = radius(from + float64(k)*step)
		if i := k - window; i >= 1 && i <= n-2 {
			check(i)
		}
	}
	tail := n - window
	if tail < 1 {
		tail = 1
	}
	for i := tail; i <= n-2; i++ {
		check(i)
	}
	return out
}

// orbitalDominant 报告第 i 个样本是否为 [lo, hi] 内的唯一最值（并列时只取最左侧）。
func orbitalDominant(value func(int) float64, i, lo, hi int, maximize bool) bool {
	better := func(a, b float64) bool {
		if maximize {
			return a > b
		}
		return a < b
	}
	vi := value(i)
	for j := lo; j <= hi; j++ {
		if j == i {
			continue
		}
		if vj := value(j); better(vj, vi) || (vj == vi && j < i) {
			return false
		}
	}
	return !math.IsNaN(vi)
}

func orbitalGoldenExtremum(f func(float64) float64, a, b float64, maximize bool) (float64, float64) {
	sign := 1.0
	if maximize {
		sign = -1
	}
	const ratio = 0.6180339887498949
	c, d := b-ratio*(b-a), a+ratio*(b-a)
	fc, fd := sign*f(c), sign*f(d)
	for i := 0; i < 100 && b-a > 1e-5; i++ {
		if fc < fd {
			b, d, fd = d, c, fc
			c = b - ratio*(b-a)
			fc = sign * f(c)
		} else {
			a, c, fc = c, d, fd
			d = a + ratio*(b-a)
			fd = sign * f(d)
		}
	}
	t := (a + b) / 2
	return t, f(t)
}

// PlanetNodesWithInstance 行星在 [startJDE, endJDE] 内穿过历元平黄道的升交点与降交点时刻（TT）/ ascending and descending node passages.
func PlanetNodesWithInstance(body SkyBody, startJDE, endJDE float64, inst *Instance) []OrbitalEvent {
	index, period, ok := orbitalIndex(body)
	if !ok || index < 0 || !finite(startJDE) || !finite(endJDE) || endJDE < startJDE {
		return nil
	}
	latitude := func(t float64) float64 { return PlanetCoordinateNWithInstance(index, 1, t, -1, inst) }
	count, at, ok := skyGrid(startJDE, endJDE, period/200, 1)
	if !ok {
		return nil
	}
	var out []OrbitalEvent
	last := map[OrbitalEventKind]float64{}
	prevT, prevB := startJDE, latitude(startJDE)
	for step := 1; step <= count; step++ {
		t := at(step)
		b := latitude(t)
		if (prevB < 0) != (b < 0) && !math.IsNaN(b) && !math.IsNaN(prevB) {
			kind := OrbitalAscendingNode
			if b < 0 {
				kind = OrbitalDescendingNode
			}
			root := orbitalBisect(latitude, prevT, t, prevB)
			if seen, ok := last[kind]; !ok || root-seen > 0.3*period {
				last[kind] = root
				out = append(out, OrbitalEvent{JDE: root, Kind: kind, DistanceAU: PlanetCoordinateNWithInstance(index, 2, root, -1, inst)})
			}
		}
		prevT, prevB = t, b
	}
	return out
}

func orbitalBisect(f func(float64) float64, a, b, fa float64) float64 {
	for i := 0; i < 80 && b-a > 1e-6; i++ {
		m := (a + b) / 2
		fm := f(m)
		if (fm < 0) == (fa < 0) {
			a, fa = m, fm
		} else {
			b = m
		}
	}
	return (a + b) / 2
}
