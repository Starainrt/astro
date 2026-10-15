// Package lunar 月球半解析级数运行时（地心，J2000 平黄道）。
package lunar

import (
	"encoding/binary"
	"errors"
	"math"
	"sync"
)

const (
	j2000       = 2451545.0
	century     = 36525.0
	arcsec      = math.Pi / 648000.0
	fundCount   = 13
	maxEnvelope = 8
)

type coordinate struct {
	args  []uint16
	nodes []int32
	// 包络阶数为 0 的项单独成段：flat0 为 [cos,sin] 交错
	nodes0 []int32
	flat0  []float32
	// 其余项（阶数 ≥1）
	nodesH  []int32
	envH    []uint8
	coefH   []float32
	env     []uint8
	secular []float64
	coef    []float32
}

// Model 已解析的月球级数表 / parsed lunar series table.
type Model struct {
	s2      [5][]float64
	args    [fundCount][5]float64
	lonPoly [5]float64
	p, q    [5]float64
	maxMult [fundCount]int
	powOff  [fundCount]int
	coords  [3]coordinate
	maxEnv  int
	// 前缀树：节点 i 的复相位 = 节点 parent[i] × pow[factor[i]]，parent 恒小于 i（-1 为根）
	nodeParent []int32
	nodeFactor []int32
	argNode    []int32
	// pow 槽位 → (基本幅角, 乘数)，用于速度计算中的相位变化率
	slotFund []int8
	slotMult []float64
	powLen   int
	legLen   int
	// Parse 之后 Model 只读；每次求值的缓冲从池中取，可并发调用
	pool sync.Pool
}

// scratch 单次求值的工作区。
type scratch struct {
	pow      []complex128
	node     []complex128
	nodeRate []float64
	leg      []float64
	dleg     []float64
	env      [maxEnvelope + 1]float64
	denv     [maxEnvelope + 1]float64
}

func (m *Model) getScratch() *scratch {
	if s, ok := m.pool.Get().(*scratch); ok {
		return s
	}
	return &scratch{
		pow:      make([]complex128, m.powLen),
		node:     make([]complex128, len(m.nodeParent)),
		nodeRate: make([]float64, len(m.nodeParent)),
		leg:      make([]float64, m.legLen),
		dleg:     make([]float64, m.legLen),
	}
}

type reader struct {
	b   []byte
	off int
	err error
}

func (r *reader) take(n int) []byte {
	if r.err != nil || r.off+n > len(r.b) {
		r.err = errors.New("lunar: truncated table")
		return make([]byte, n)
	}
	s := r.b[r.off : r.off+n]
	r.off += n
	return s
}

func (r *reader) u8() uint8    { return r.take(1)[0] }
func (r *reader) u16() uint16  { return binary.LittleEndian.Uint16(r.take(2)) }
func (r *reader) f64() float64 { return math.Float64frombits(binary.LittleEndian.Uint64(r.take(8))) }
func (r *reader) f32() float32 { return math.Float32frombits(binary.LittleEndian.Uint32(r.take(4))) }

// Parse 解析 LM7 v2 二进制表 / parse an LM7 v2 table.
func Parse(b []byte) (*Model, error) {
	r := &reader{b: b}
	if string(r.take(4)) != "LM7\x02" {
		return nil, errors.New("lunar: bad magic")
	}
	m := &Model{}
	nargs := int(r.u16())
	s2n := int(r.u16())
	secDeg := int(r.u8())
	for j := 0; j < 5; j++ {
		m.s2[j] = make([]float64, s2n)
		for k := range m.s2[j] {
			m.s2[j][k] = r.f64()
		}
	}
	for i := 0; i < fundCount; i++ {
		for k := 0; k < 5; k++ {
			m.args[i][k] = r.f64()
		}
	}
	for k := range m.lonPoly {
		m.lonPoly[k] = r.f64()
	}
	for k := range m.p {
		m.p[k] = r.f64()
	}
	for k := range m.q {
		m.q[k] = r.f64()
	}
	r.f64() // 距离比例已折入拟合系数
	mults := make([][fundCount]int8, nargs)
	for a := range mults {
		mask := r.u16()
		for i := 0; i < 16; i++ {
			if mask&(1<<uint(i)) == 0 {
				continue
			}
			if i >= fundCount {
				return nil, errors.New("lunar: multiplier index out of range")
			}
			k := int8(r.u8())
			mults[a][i] = k
			if v := int(math.Abs(float64(k))); v > m.maxMult[i] {
				m.maxMult[i] = v
			}
		}
	}
	m.buildTrie(mults)
	for c := 0; c < 3; c++ {
		n := int(r.u16())
		co := &m.coords[c]
		co.args = make([]uint16, n)
		for j := range co.args {
			co.args[j] = r.u16()
			if int(co.args[j]) >= nargs {
				return nil, errors.New("lunar: argument index out of range")
			}
		}
		co.env = r.take(n)
		ncoef := 0
		for _, e := range co.env {
			if int(e) > maxEnvelope {
				return nil, errors.New("lunar: envelope degree too high")
			}
			if int(e) > m.maxEnv {
				m.maxEnv = int(e)
			}
			ncoef += 2 * (int(e) + 1)
		}
		co.secular = make([]float64, secDeg+1)
		for k := range co.secular {
			co.secular[k] = r.f64()
		}
		raw := make([]float32, ncoef)
		for k := range raw {
			raw[k] = r.f32()
		}
		co.coef = interleave(co.env, raw)
	}
	if r.err != nil {
		return nil, r.err
	}
	if r.off != len(b) {
		return nil, errors.New("lunar: trailing bytes")
	}
	for c := range m.coords {
		co := &m.coords[c]
		co.nodes = make([]int32, len(co.args))
		pos := 0
		for j, a := range co.args {
			co.nodes[j] = m.argNode[a]
			e := int(co.env[j])
			if e == 0 {
				co.nodes0 = append(co.nodes0, co.nodes[j])
				co.flat0 = append(co.flat0, co.coef[pos], co.coef[pos+1])
			} else {
				co.nodesH = append(co.nodesH, co.nodes[j])
				co.envH = append(co.envH, co.env[j])
				co.coefH = append(co.coefH, co.coef[pos:pos+2*(e+1)]...)
			}
			pos += 2 * (e + 1)
		}
	}
	m.legLen = max3(s2n, secDeg+1, m.maxEnv+1)
	return m, nil
}

func (m *Model) buildTrie(mults [][fundCount]int8) {
	total := 0
	for i := 0; i < fundCount; i++ {
		m.powOff[i] = total + m.maxMult[i]
		total += 2*m.maxMult[i] + 1
	}
	m.powLen = total
	m.slotFund = make([]int8, total)
	m.slotMult = make([]float64, total)
	for i := 0; i < fundCount; i++ {
		for k := -m.maxMult[i]; k <= m.maxMult[i]; k++ {
			m.slotFund[m.powOff[i]+k] = int8(i)
			m.slotMult[m.powOff[i]+k] = float64(k)
		}
	}
	type edge struct{ parent, factor int32 }
	index := map[edge]int32{}
	m.argNode = make([]int32, len(mults))
	for a, mv := range mults {
		cur := int32(-1)
		for i := 0; i < fundCount; i++ {
			if mv[i] == 0 {
				continue
			}
			e := edge{cur, int32(m.powOff[i] + int(mv[i]))}
			n, ok := index[e]
			if !ok {
				n = int32(len(m.nodeParent))
				index[e] = n
				m.nodeParent = append(m.nodeParent, e.parent)
				m.nodeFactor = append(m.nodeFactor, e.factor)
			}
			cur = n
		}
		m.argNode[a] = cur
	}
}

// cmul 绕开内建复数乘法的 Inf/NaN 分支。
func cmul(a, b complex128) complex128 {
	ar, ai, br, bi := real(a), imag(a), real(b), imag(b)
	return complex(ar*br-ai*bi, ar*bi+ai*br)
}

func max3(a, b, c int) int {
	if b > a {
		a = b
	}
	if c > a {
		a = c
	}
	return a
}

func poly(c *[5]float64, t float64) float64 {
	return c[0] + t*(c[1]+t*(c[2]+t*(c[3]+t*c[4])))
}

// phase 与 poly 同值，但线性项单独做补偿归约：c1·t 在 ±1 万年达 1e6 rad，直接相加会丢掉约 1e-10 rad。
func phase(c *[5]float64, t, tLow float64) float64 {
	const twoPiLow = 2.4492935982947064e-16
	product := c[1] * t
	roundoff := math.FMA(c[1], t, -product) + c[1]*tLow
	turns := math.Round(product / (2 * math.Pi))
	lin := math.FMA(-turns, 2*math.Pi, product) + (roundoff - turns*twoPiLow)
	return lin + (c[0] + t*t*(c[2]+t*(c[3]+t*c[4])))
}

// dpoly 返回 poly 的一阶导数。
func dpoly(c *[5]float64, t float64) float64 {
	return c[1] + t*(2*c[2]+t*(3*c[3]+t*4*c[4]))
}

// legendreD 写入 P0..P(n-1)(x)；withD 时同时写入导数 P'k(x) 到 d。
func legendreD(x float64, out, d []float64, withD bool) {
	out[0] = 1
	if len(out) > 1 {
		out[1] = x
	}
	for k := 2; k < len(out); k++ {
		out[k] = ((2*float64(k)-1)*x*out[k-1] - (float64(k)-1)*out[k-2]) / float64(k)
	}
	if !withD {
		return
	}
	d[0] = 0
	if len(d) > 1 {
		d[1] = 1
	}
	// P'k = P'(k-2) + (2k-1)·P(k-1)
	for k := 2; k < len(d); k++ {
		d[k] = d[k-2] + (2*float64(k)-1)*out[k-1]
	}
}

func dot(c []float64, l []float64) float64 {
	s := 0.0
	for k, v := range c {
		s += v * l[k]
	}
	return s
}

// PositionJ2000 地心月球 J2000 平黄道几何位置（km）/ geocentric Moon, J2000 mean ecliptic, km.
// 可并发调用。
func (m *Model) PositionJ2000(jdTT float64) [3]float64 {
	p, _ := m.eval(jdTT, false)
	return p
}

// PositionVelocityJ2000 位置（km）与解析速度（km/日，对 TT）/ position (km) and analytic velocity (km/day).
// 可并发调用。
func (m *Model) PositionVelocityJ2000(jdTT float64) (pos, vel [3]float64) {
	return m.eval(jdTT, true)
}

// eval 所有量先对 t（儒略世纪）求导，最后换算到每日。
func (m *Model) eval(jdTT float64, wantVel bool) (pos, vel [3]float64) {
	s := m.getScratch()
	defer m.pool.Put(s)
	days := jdTT - j2000
	origin := days - jdTT
	daysLow := (jdTT - (days - origin)) - (j2000 + origin)
	t := days / century
	// tLow 保留历元相减与世纪除法的舍入，远离 J2000 时线性相位才不丢精度。
	tLow := (math.FMA(-t, century, days) + daysLow) / century
	x := (t + 20) / 100
	const dxdt = 0.01
	leg, dleg := s.leg, s.dleg
	// 递推前缀稳定，一次算到最长，S2 修正与各坐标长期项共用。
	legendreD(x, leg[:m.legLen], dleg[:m.legLen], wantVel)
	var dc, ddc [5]float64
	for j := range dc {
		dc[j] = dot(m.s2[j], leg)
		if wantVel {
			ddc[j] = dot(m.s2[j], dleg) * dxdt
		}
	}
	// ζ 随 W1 修正，D/F/l/l′ 各自修正
	var corr, dcorr, rate [fundCount]float64
	corr[0], corr[1], corr[2], corr[3], corr[12] = dc[1], dc[2], dc[3], dc[4], dc[0]
	dcorr[0], dcorr[1], dcorr[2], dcorr[3], dcorr[12] = ddc[1], ddc[2], ddc[3], ddc[4], ddc[0]
	for i := 0; i < fundCount; i++ {
		k := m.maxMult[i]
		if k == 0 {
			continue
		}
		rate[i] = dpoly(&m.args[i], t) + dcorr[i]
		sn, cs := math.Sincos(phase(&m.args[i], t, tLow) + corr[i])
		e := complex(cs, sn)
		p := s.pow[m.powOff[i]-k : m.powOff[i]+k+1]
		p[k] = 1
		for n := 1; n <= k; n++ {
			p[k+n] = cmul(p[k+n-1], e)
			p[k-n] = complex(real(p[k+n]), -imag(p[k+n]))
		}
	}
	// 热循环按 wantVel 拆开，纯位置路径不写 nodeRate
	node, factor := s.node, m.nodeFactor
	if wantVel {
		nr := s.nodeRate
		for n, parent := range m.nodeParent {
			fac := factor[n]
			f := s.pow[fac]
			r := m.slotMult[fac] * rate[m.slotFund[fac]]
			if parent >= 0 {
				f = cmul(node[parent], f)
				r += nr[parent]
			}
			node[n] = f
			nr[n] = r
		}
	} else {
		for n, parent := range m.nodeParent {
			f := s.pow[factor[n]]
			if parent >= 0 {
				f = cmul(node[parent], f)
			}
			node[n] = f
		}
	}
	env, denv := s.env[:m.maxEnv+1], s.denv[:m.maxEnv+1]
	legendreD(x, env, denv, wantVel)
	var v, dv [3]float64
	for c := 0; c < 3; c++ {
		co := &m.coords[c]
		sum := dot(co.secular, leg)
		dsum := 0.0
		if wantVel {
			dsum = dot(co.secular, dleg) * dxdt
		}
		f0 := co.flat0
		if wantVel {
			nr := s.nodeRate
			for j, n := range co.nodes0 {
				z := node[n]
				a, b := float64(f0[2*j]), float64(f0[2*j+1])
				sum += a*real(z) + b*imag(z)
				dsum += nr[n] * (b*real(z) - a*imag(z))
			}
			pos := 0
			for j, e := range co.envH {
				n := co.nodesH[j]
				z := node[n]
				var ac, as, dac, das float64
				for d := 0; d <= int(e); d++ {
					cc, cn := float64(co.coefH[pos]), float64(co.coefH[pos+1])
					ac += env[d] * cc
					as += env[d] * cn
					dac += denv[d] * cc
					das += denv[d] * cn
					pos += 2
				}
				sum += ac*real(z) + as*imag(z)
				dsum += dxdt*(dac*real(z)+das*imag(z)) + nr[n]*(as*real(z)-ac*imag(z))
			}
		} else {
			for j, n := range co.nodes0 {
				z := node[n]
				sum += float64(f0[2*j])*real(z) + float64(f0[2*j+1])*imag(z)
			}
			pos := 0
			for j, e := range co.envH {
				z := node[co.nodesH[j]]
				var ac, as float64
				for d := 0; d <= int(e); d++ {
					ac += env[d] * float64(co.coefH[pos])
					as += env[d] * float64(co.coefH[pos+1])
					pos += 2
				}
				sum += ac*real(z) + as*imag(z)
			}
		}
		v[c], dv[c] = sum, dsum
	}
	lon := v[0]*arcsec + phase(&m.lonPoly, t, tLow) + dc[0]
	lat := v[1] * arcsec
	rad := v[2]
	sl, cl := math.Sincos(lon)
	sb, cb := math.Sincos(lat)
	x0, y0, z0 := rad*cb*cl, rad*cb*sl, rad*sb
	p := poly(&m.p, t) * t
	q := poly(&m.q, t) * t
	ra := 2 * math.Sqrt(1-p*p-q*q)
	pos = [3]float64{
		(1-2*p*p)*x0 + 2*p*q*y0 + p*ra*z0,
		2*p*q*x0 + (1-2*q*q)*y0 - q*ra*z0,
		-p*ra*x0 + q*ra*y0 + (1-2*p*p-2*q*q)*z0,
	}
	if !wantVel {
		return pos, vel
	}
	dlon := dv[0]*arcsec + dpoly(&m.lonPoly, t) + ddc[0]
	dlat := dv[1] * arcsec
	drad := dv[2]
	// 球坐标 → 直角坐标的时间导数
	dx0 := drad*cb*cl - rad*sb*cl*dlat - rad*cb*sl*dlon
	dy0 := drad*cb*sl - rad*sb*sl*dlat + rad*cb*cl*dlon
	dz0 := drad*sb + rad*cb*dlat
	// 历元黄道 → J2000 旋转矩阵本身随 P、Q 变化
	dp := dpoly(&m.p, t)*t + poly(&m.p, t)
	dq := dpoly(&m.q, t)*t + poly(&m.q, t)
	dra := -4 * (p*dp + q*dq) / ra
	dpq := 2 * (dp*q + p*dq)
	vel = [3]float64{
		(1-2*p*p)*dx0 + 2*p*q*dy0 + p*ra*dz0 +
			-4*p*dp*x0 + dpq*y0 + (dp*ra+p*dra)*z0,
		2*p*q*dx0 + (1-2*q*q)*dy0 - q*ra*dz0 +
			dpq*x0 - 4*q*dq*y0 - (dq*ra+q*dra)*z0,
		-p*ra*dx0 + q*ra*dy0 + (1-2*p*p-2*q*q)*dz0 +
			-(dp*ra+p*dra)*x0 + (dq*ra+q*dra)*y0 - 4*(p*dp+q*dq)*z0,
	}
	for i := range vel {
		vel[i] /= century
	}
	return pos, vel
}

// interleave 把按阶数分块（每阶先 cos 块后 sin 块）的系数重排为逐项 [d0cos,d0sin,d1cos,…]。
func interleave(env []uint8, raw []float32) []float32 {
	out := make([]float32, len(raw))
	start := make([]int, len(env))
	for j := 1; j < len(env); j++ {
		start[j] = start[j-1] + 2*(int(env[j-1])+1)
	}
	pos := 0
	for d := 0; ; d++ {
		var idx []int
		for j, e := range env {
			if int(e) >= d {
				idx = append(idx, j)
			}
		}
		if len(idx) == 0 {
			break
		}
		for n, j := range idx {
			out[start[j]+2*d] = raw[pos+n]
			out[start[j]+2*d+1] = raw[pos+len(idx)+n]
		}
		pos += 2 * len(idx)
	}
	return out
}
