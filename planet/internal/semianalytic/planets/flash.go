//go:build astro_lowmem
// +build astro_lowmem

package planets

import (
	_ "embed"
	"errors"
	"math"
	"sync"
)

// FlashModel 与 Model 逐位一致，系数与幅角在求值时从只读表现场解码，常驻内存只有索引 / bit-identical to Model, decodes the read-only table on each evaluation.
type FlashModel struct {
	bodies       [9]*flashBody
	modernBodies [9]*flashBody
	window       Model
}

// flashBody 复用 body 的幂表尺寸与 ellxyz；body.el、factors 等切片保持为空。
type flashBody struct {
	body
	keys string
	nk   int
	fel  [6]flashElement
}

// flashElement 的项按包络阶降序排列（解析时校验），第 d 阶覆盖前 cnt[d] 项。
type flashElement struct {
	secular []float64
	cnt     []int32
	step    float64
	src     string
	packed  bool
	sorted  bool
	nbits   int
	argPos  int
	coefPos int
	// blocks PLN2 第 d 阶 cos、sin 系数块在 src 中的起点。
	blocks   []int32
	expanded expandedData
}

type expandedData struct {
	args string
	coef string
}

// ParseFlash 为 PLN2/PLN3 表建立索引，表必须在模型存活期间保持不变 / index a PLN2/PLN3 table in place.
func ParseFlash(s string) (*FlashModel, error) {
	bodies, err := parseFlashBodies(s)
	if err != nil {
		return nil, err
	}
	return &FlashModel{bodies: bodies}, nil
}

// WithModernSegment 叠加现代段表，[jdStart, jdEnd] 内优先使用 / add a modern-era segment used inside [jdStart, jdEnd].
func (m *FlashModel) WithModernSegment(s string, jdStart, jdEnd float64) error {
	bodies, err := parseFlashBodies(s)
	if err != nil {
		return err
	}
	m.modernBodies = bodies
	m.window.modernStart, m.window.modernEnd = jdStart, jdEnd
	return nil
}

// Heliocentric 日心位置，km，J2000 平黄道 / heliocentric position, km, mean ecliptic J2000.
func (m *FlashModel) Heliocentric(id int, jdTDB float64) ([3]float64, error) {
	pos, _, err := m.state(id, jdTDB, false)
	return pos, err
}

// HeliocentricState 日心位置（km）与密切速度（km/日），J2000 平黄道 / heliocentric position and osculating velocity.
func (m *FlashModel) HeliocentricState(id int, jdTDB float64) (pos, vel [3]float64, err error) {
	return m.state(id, jdTDB, true)
}

func (m *FlashModel) state(id int, jdTDB float64, wantVel bool) (pos, vel [3]float64, err error) {
	if id < 1 || id > 8 {
		return pos, vel, errors.New("planets: body not in table")
	}
	fb := m.bodies[id]
	if fb == nil {
		return pos, vel, errors.New("planets: body not in table")
	}
	t := (jdTDB - j2000) / millennium
	md := m.modernBodies[id]
	sky := func(p, v [3]float64) ([3]float64, [3]float64, error) {
		if !wantVel {
			return toSky(p), v, nil
		}
		return toSky(p), toSky(v), nil
	}
	if md == nil || jdTDB <= m.window.modernStart || jdTDB >= m.window.modernEnd {
		p, v := fb.ellxyz(fb.flashElements(t), wantVel)
		return sky(p, v)
	}
	w, dw := m.window.modernWeight(jdTDB)
	pm, vm := md.ellxyz(md.flashElements(t), wantVel)
	if w == 1 {
		return sky(pm, vm)
	}
	pl, vl := fb.ellxyz(fb.flashElements(t), wantVel)
	for i := range pos {
		pos[i] = pl[i] + w*(pm[i]-pl[i])
		if wantVel {
			vel[i] = vl[i] + w*(vm[i]-vl[i]) + dw*(pm[i]-pl[i])
		}
	}
	return sky(pos, vel)
}

type flashScratch struct {
	buf  []float64
	args []uint16
	leg  [maxDeg + 1]float64
	el   [6]float64
}

var flashScratchPool = sync.Pool{New: func() interface{} { return new(flashScratch) }}

// flashPhasors 与 phasors 同序同式，因子从键区现场解码。
func (fb *flashBody) flashPhasors(t float64, pc, ps, tc, ts []float64) {
	for i := range fb.tabMax {
		if fb.tabMax[i] == 0 && !(fb.top && i < 3) {
			continue
		}
		var s, c float64
		if fb.top {
			s, c = math.Sincos(topDMu * t * float64(int(1)<<uint(topBits*i)))
		} else {
			s, c = math.Sincos(ci0[i] + ci1[i]*t)
		}
		o := fb.tabOff[i]
		tc[o], ts[o] = 1, 0
		for j := 1; j <= fb.tabMax[i]; j++ {
			tc[o+j], ts[o+j] = cmul(tc[o+j-1], ts[o+j-1], c, s)
		}
		if fb.top && i == 2 {
			break
		}
	}
	r := &strReader{s: fb.keys}
	for j := 0; j < fb.nk; j++ {
		re, im, first := 1.0, 0.0, true
		keyNeg := false
		if fb.top {
			a := int(r.zvar())
			if a < 0 {
				a, keyNeg = -a, true
			}
			for lvl := 0; lvl < 3; lvl++ {
				d := a >> uint(topBits*lvl)
				if lvl < 2 {
					d &= 1<<topBits - 1
				}
				slot := fb.tabOff[lvl] + d
				if first {
					re, im, first = tc[slot], ts[slot], false
				} else {
					re, im = cmul(re, im, tc[slot], ts[slot])
				}
			}
		} else {
			mask := int(r.u8()) | int(r.u8())<<8 | int(r.u8())<<16
			for i := 0; i < 17; i++ {
				if mask>>uint(i)&1 == 0 {
					continue
				}
				a, neg := int(r.zvar()), false
				if a < 0 {
					a, neg = -a, true
				}
				if a == 0 {
					continue
				}
				br, bi := tc[fb.tabOff[i]+a], ts[fb.tabOff[i]+a]
				if neg {
					bi = -bi
				}
				if first {
					re, im, first = br, bi, false
				} else {
					re, im = cmul(re, im, br, bi)
				}
			}
		}
		if keyNeg {
			im = -im
		}
		pc[j], ps[j] = re, im
	}
}

// flashElements 与 elements 同序同式。
func (fb *flashBody) flashElements(t float64) [6]float64 {
	x := (t - tMid) / tHalf
	sc := flashScratchPool.Get().(*flashScratch)
	defer flashScratchPool.Put(sc)
	leg := &sc.leg
	leg[0], leg[1] = 1, x
	for k := 2; k <= fb.legDeg; k++ {
		leg[k] = (leg[k-1]*x*float64(2*k-1) - leg[k-2]*float64(k-1)) / float64(k)
	}
	nk := fb.nk
	if cap(sc.buf) < 2*nk+2*maxTab {
		sc.buf = make([]float64, 2*nk+2*maxTab)
	}
	if cap(sc.args) < nk {
		sc.args = make([]uint16, nk)
	}
	buf := sc.buf[:2*nk+2*maxTab]
	pc, ps := buf[:nk], buf[nk:2*nk]
	fb.flashPhasors(t, pc, ps, buf[2*nk:2*nk+maxTab], buf[2*nk+maxTab:])
	el := &sc.el
	for v := range fb.fel {
		e := &fb.fel[v]
		sum := 0.0
		for i, c := range e.secular {
			sum += c * leg[i]
		}
		if len(e.cnt) > 0 {
			args := sc.args[:e.cnt[0]]
			if len(e.expanded.args) != 0 {
				for i := range args {
					args[i] = uint16(e.expanded.args[2*i]) | uint16(e.expanded.args[2*i+1])<<8
				}
			} else {
				e.readArgs(args)
			}
			if len(e.expanded.coef) != 0 {
				term := 0
				for d, cnt := range e.cnt {
					s0, s1 := 0.0, 0.0
					for _, a := range args[:cnt] {
						s0 += expandedFloat64(e.expanded.coef, 16*term) * pc[a]
						s1 += expandedFloat64(e.expanded.coef, 16*term+8) * ps[a]
						term++
					}
					sum += leg[d] * (s0 + s1)
				}
			} else if e.packed {
				br := &strBits{s: e.src, p: e.coefPos}
				width := 0
				for d, cnt := range e.cnt {
					s0, s1 := 0.0, 0.0
					for _, a := range args[:cnt] {
						width += unzigzag(br.gamma())
						s0 += br.coef(width, e.step) * pc[a]
						s1 += br.coef(width, e.step) * ps[a]
					}
					sum += leg[d] * (s0 + s1)
				}
			} else {
				for d, cnt := range e.cnt {
					rc := &strReader{s: e.src, p: int(e.blocks[2*d])}
					rs := &strReader{s: e.src, p: int(e.blocks[2*d+1])}
					s0, s1 := 0.0, 0.0
					for _, a := range args[:cnt] {
						s0 += float64(float64(rc.zvar())*e.step) * pc[a]
					}
					for _, a := range args[:cnt] {
						s1 += float64(float64(rs.zvar())*e.step) * ps[a]
					}
					sum += leg[d] * (s0 + s1)
				}
			}
		}
		el[v] = sum
	}
	lm := math.Mod(fb.l0+fb.f*t, 2*math.Pi)
	if lm < 0 {
		lm += 2 * math.Pi
	}
	el[1] += lm
	return *el
}

// expandedFloat64 按小端逐字节拼装，不依赖地址对齐与宿主字节序。
func expandedFloat64(s string, offset int) float64 {
	s = s[offset : offset+8]
	return math.Float64frombits(uint64(s[0]) |
		uint64(s[1])<<8 |
		uint64(s[2])<<16 |
		uint64(s[3])<<24 |
		uint64(s[4])<<32 |
		uint64(s[5])<<40 |
		uint64(s[6])<<48 |
		uint64(s[7])<<56)
}

// readArgs 解码全部项的幅角号；各阶只取前缀，所以只解一次。
func (e *flashElement) readArgs(out []uint16) []uint16 {
	if !e.packed {
		r := &strReader{s: e.src, p: e.argPos}
		for j := range out {
			out[j] = uint16(r.uvar())
			r.u8()
		}
		return out
	}
	br := &strBits{s: e.src, p: e.argPos}
	d, last, cur := len(e.cnt)-1, -1, -1
	for j := range out {
		if e.sorted {
			for int(e.cnt[d]) <= j {
				d--
			}
			if d != cur {
				last, cur = -1, d
			}
			last += 1 + br.gamma()
			out[j] = uint16(last)
		} else {
			out[j] = uint16(br.get(e.nbits))
		}
	}
	return out
}

var (
	//go:embed long.bin
	flashLong string
	//go:embed modern.bin
	flashModern string
	//go:embed fast.bin
	flashFast string
)

var (
	flashDefaultOnce  sync.Once
	flashDefaultModel *FlashModel
	flashFastOnce     sync.Once
	flashFastModel    *FlashModel
)

// FlashDefault 与 Default 同表的只读常驻版本 / read-only resident counterpart of Default.
func FlashDefault() *FlashModel {
	flashDefaultOnce.Do(func() {
		m, err := ParseFlash(flashLong)
		if err == nil {
			err = m.WithModernSegment(flashModern, ModernStart, ModernEnd)
		}
		if err != nil {
			panic(err)
		}
		flashDefaultModel = m
		attachExpanded(flashDefaultModel.bodies, flashExpandedLong, expandedLong)
		attachExpanded(flashDefaultModel.modernBodies, flashExpandedModern, expandedModern)
	})
	return flashDefaultModel
}

// FlashFast 与 Fast 同表的只读常驻版本 / read-only resident counterpart of Fast.
func FlashFast() *FlashModel {
	flashFastOnce.Do(func() {
		m, err := ParseFlash(flashFast)
		if err != nil {
			panic(err)
		}
		flashFastModel = m
		attachExpanded(flashFastModel.bodies, flashExpandedFast, expandedFast)
	})
	return flashFastModel
}
