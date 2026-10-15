// Package planets 行星半解析级数运行时（PLN1/PLN2/PLN3 表，日心，J2000 平黄道）。
package planets

import (
	"encoding/binary"
	"errors"
	"math"
	"sync"
)

const (
	j2000      = 2451545.0
	millennium = 365250.0
	blendDays  = 36525.0
	auKM       = 149597870.691
	tMid       = -2.0
	tHalf      = 10.0
	emrat      = 81.3005682214972
	gmSol      = 2.9591220836841438269e-04
	maxDeg     = 24
)

// Body 编号 / body id；3 为地月质心。
const (
	Mercury = 1
	Venus   = 2
	EMB     = 3
	Mars    = 4
	Jupiter = 5
	Saturn  = 6
	Uranus  = 7
	Neptune = 8
)

var ci0 = [17]float64{4.402608631669, 3.176134461576, 1.753470369433, 6.203500014141, 4.091360003050,
	1.713740719173, 5.598641292287, 2.805136360408, 2.326989734620, 0.5995461070350,
	0.8740185101070, 5.481225395663, 5.311897933164, 0.0, 5.19846640063,
	1.62790513602, 2.35555563875}

var ci1 = [17]float64{26087.90314068555, 10213.28554743445, 6283.075850353215, 3340.612434145457,
	1731.170452721855, 1704.450855027201, 1428.948917844273, 1364.756513629990,
	1361.923207632842, 529.6909615623250, 213.2990861084880, 74.78165903077800,
	38.13297222612500, 0.3595362285049309, 77713.7714481804, 84334.6615717837,
	83286.9142477147}

var gmp = [8]float64{4.9125474514508118699e-11, 7.2434524861627027000e-10, 8.9970116036316091182e-10,
	9.5495351057792580598e-11, 2.8253458420837780000e-07, 8.4597151856806587398e-08,
	1.2920249167819693900e-08, 1.5243589007842762800e-08}

// topDMu 必须逐步按 float64 舍入（与 Python 参考求值一致），不能写成常量表达式。
var topDMu = func() float64 {
	f0, f1, d := 529.6909622785881, 213.2990811942489, 880.0
	return (f0 - f1) / d
}()

// skyToVSOP skyfield ecliptic_J2000 → VSOP2013 动力学黄道；求值时用其转置。
var skyToVSOP = func() [3][3]float64 {
	as := 180 / math.Pi * 3600
	return mul(transpose(mul(rz(-0.05188/as), rx(84381.41136/as))), rx(84381.448/as))
}()

// element 的项按包络阶降序排列，第 d 阶只覆盖前 cnt[d] 项。
type element struct {
	secular []float64
	arg     []int32
	cnt     []int
	c, s    [][]float64
}

const (
	maxKeys = 8192
	maxTab  = 512
	topBits = 6
	// 超出此值的幅角乘数必然撑破幂表，先在 int64 下拒绝，保证转 int 不溢出。
	maxKeyAbs = maxTab << (2 * topBits)
)

var errBadKey = errors.New("planets: argument multiplier out of range")

// factor 幂表槽位；neg 表示取共轭。
type factor struct {
	slot int32
	neg  bool
}

type body struct {
	top   bool
	f, l0 float64
	rgm   float64
	// 幂表：VSOP 每个基本幅角一段，TOP 每个 6 位数字层一段
	tabMax  [17]int
	tabOff  [17]int
	keyOff  []int32
	factors []factor
	keyNeg  []bool
	el      [6]element
	// legDeg 长期项与包络用到的最高 Legendre 阶。
	legDeg int
}

// Model 已解析的 PLN1 表，只读、可并发求值 / parsed PLN1 table, safe for concurrent use.
type Model struct {
	bodies       [9]*body
	modernBodies [9]*body
	modernStart  float64
	modernEnd    float64
}

type reader struct {
	b   []byte
	p   int
	err error
}

func (r *reader) take(n int) []byte {
	if r.err != nil || r.p+n > len(r.b) {
		r.err = errors.New("planets: truncated table")
		return make([]byte, n)
	}
	s := r.b[r.p : r.p+n]
	r.p += n
	return s
}

func (r *reader) u8() uint8    { return r.take(1)[0] }
func (r *reader) f64() float64 { return math.Float64frombits(binary.LittleEndian.Uint64(r.take(8))) }
func (r *reader) f32() float32 { return math.Float32frombits(binary.LittleEndian.Uint32(r.take(4))) }

func (r *reader) uvar() uint64 {
	n, shift := uint64(0), uint(0)
	for {
		c := r.u8()
		n |= uint64(c&0x7F) << shift
		shift += 7
		if c < 0x80 || r.err != nil || shift > 56 {
			return n
		}
	}
}

func (r *reader) zvar() int64 {
	z := r.uvar()
	if z%2 == 0 {
		return int64(z >> 1)
	}
	return -int64(z>>1) - 1
}

// Parse 解析 PLN1 表 / parse a PLN1 table.
func Parse(b []byte) (*Model, error) {
	bodies, err := parseBodies(b)
	if err != nil {
		return nil, err
	}
	return &Model{bodies: bodies}, nil
}

// WithModernSegment 为长时段表叠加现代段表，[jdStart, jdEnd] 内优先使用 / add a modern-era segment used inside [jdStart, jdEnd].
func (m *Model) WithModernSegment(b []byte, jdStart, jdEnd float64) error {
	bodies, err := parseBodies(b)
	if err != nil {
		return err
	}
	m.modernBodies, m.modernStart, m.modernEnd = bodies, jdStart, jdEnd
	return nil
}

func parseBodies(b []byte) (bodies [9]*body, err error) {
	r := &reader{b: b}
	magic := string(r.take(4))
	if magic != "PLN1" && magic != "PLN2" && magic != "PLN3" {
		return bodies, errors.New("planets: bad magic")
	}
	quant, packed := magic == "PLN2", magic == "PLN3"
	for n := int(r.u8()); n > 0 && r.err == nil; n-- {
		id := int(r.u8())
		if id < 1 || id > 8 {
			return bodies, errors.New("planets: bad body id")
		}
		bd := &body{top: r.u8() != 0}
		secDeg := int(r.u8())
		bd.f, bd.l0 = r.f64(), r.f64()
		bd.rgm = math.Sqrt(gmp[id-1] + gmSol)
		nk := r.uvar()
		if nk > maxKeys {
			return bodies, errors.New("planets: too many arguments")
		}
		if err := readKeys(r, bd, int(nk)); err != nil {
			return bodies, err
		}
		bd.legDeg = secDeg
		sortedArgs := packed && r.u8() != 0
		for v := range bd.el {
			var err error
			if packed {
				err = readElement3(r, &bd.el[v], secDeg, int(nk), sortedArgs)
			} else {
				err = readElement(r, &bd.el[v], secDeg, int(nk), quant)
			}
			if err != nil {
				return bodies, err
			}
			if d := len(bd.el[v].cnt) - 1; d > bd.legDeg {
				bd.legDeg = d
			}
		}
		bodies[id] = bd
	}
	if r.err != nil {
		return bodies, r.err
	}
	if r.p != len(b) {
		return bodies, errors.New("planets: trailing bytes")
	}
	return bodies, nil
}

// readKeys 把整数幅角拆成幂表槽位：VSOP 按基本幅角升序，TOP 按 6 位数字层由低到高。
func readKeys(r *reader, bd *body, nk int) error {
	type raw struct{ fund, k int }
	keys := make([][]raw, nk)
	bd.keyNeg = make([]bool, nk)
	for j := 0; j < nk && r.err == nil; j++ {
		if bd.top {
			k := r.zvar()
			if k > maxKeyAbs || k < -maxKeyAbs {
				return errBadKey
			}
			a := int(k)
			if a < 0 {
				a, bd.keyNeg[j] = -a, true
			}
			for lvl := 0; lvl < 3; lvl++ {
				d := a >> uint(topBits*lvl)
				if lvl < 2 {
					d &= 1<<topBits - 1
				}
				keys[j] = append(keys[j], raw{lvl, d})
			}
			continue
		}
		mb := r.take(3)
		mask := int(mb[0]) | int(mb[1])<<8 | int(mb[2])<<16
		for i := 0; i < 17; i++ {
			if mask>>uint(i)&1 != 0 {
				k := r.zvar()
				if k > maxKeyAbs || k < -maxKeyAbs {
					return errBadKey
				}
				keys[j] = append(keys[j], raw{i, int(k)})
			}
		}
	}
	if r.err != nil {
		return r.err
	}
	for _, ks := range keys {
		for _, f := range ks {
			a := f.k
			if a < 0 {
				a = -a
			}
			if a > bd.tabMax[f.fund] {
				bd.tabMax[f.fund] = a
			}
		}
	}
	n := 0
	for i := range bd.tabMax {
		bd.tabOff[i] = n
		if bd.tabMax[i] > 0 || (bd.top && i < 3) {
			n += bd.tabMax[i] + 1
		}
	}
	if n > maxTab {
		return errors.New("planets: power table too large")
	}
	bd.keyOff = make([]int32, nk+1)
	for j, ks := range keys {
		for _, f := range ks {
			a, neg := f.k, false
			if a < 0 {
				a, neg = -a, true
			}
			if !bd.top && a == 0 {
				continue
			}
			bd.factors = append(bd.factors, factor{int32(bd.tabOff[f.fund] + a), neg})
		}
		bd.keyOff[j+1] = int32(len(bd.factors))
	}
	return nil
}

// readElement 把文件里按包络阶分块的列序重排为逐项连续；quant 为 PLN2 整数量化系数。
func readElement(r *reader, e *element, secDeg, nk int, quant bool) error {
	if secDeg > maxDeg {
		return errors.New("planets: secular degree too high")
	}
	n64 := r.uvar()
	if n64 > uint64(nk) {
		return errors.New("planets: too many terms")
	}
	nint := int(n64)
	e.secular = make([]float64, secDeg+1)
	for i := range e.secular {
		e.secular[i] = r.f64()
	}
	args, env := make([]int32, nint), make([]int, nint)
	top := 0
	for j := 0; j < nint; j++ {
		a := r.uvar()
		if a >= uint64(nk) {
			return errors.New("planets: argument index out of range")
		}
		args[j], env[j] = int32(a), int(r.u8())
		if env[j] > maxDeg {
			return errors.New("planets: envelope degree too high")
		}
		if env[j] > top {
			top = env[j]
		}
	}
	// 按包络阶降序的稳定排序：pos[j] 为文件第 j 项的新位置
	pos := make([]int, nint)
	e.arg = make([]int32, nint)
	k := 0
	for d := top; d >= 0; d-- {
		for j := 0; j < nint; j++ {
			if env[j] == d {
				pos[j], e.arg[k] = k, args[j]
				k++
			}
		}
	}
	coef := func() float64 { return float64(r.f32()) }
	if quant {
		step := r.f64()
		coef = func() float64 { return float64(r.zvar()) * step }
	}
	if nint == 0 {
		return r.err
	}
	e.cnt = make([]int, top+1)
	e.c, e.s = make([][]float64, top+1), make([][]float64, top+1)
	for d := 0; d <= top; d++ {
		for j := 0; j < nint; j++ {
			if env[j] >= d {
				e.cnt[d]++
			}
		}
		e.c[d], e.s[d] = make([]float64, e.cnt[d]), make([]float64, e.cnt[d])
		for _, dst := range [][]float64{e.c[d], e.s[d]} {
			for j := 0; j < nint; j++ {
				if env[j] >= d {
					dst[pos[j]] = coef()
				}
			}
		}
	}
	return r.err
}

// Heliocentric 日心位置，km，J2000 平黄道 / heliocentric position, km, mean ecliptic J2000.
func (m *Model) Heliocentric(id int, jdTDB float64) ([3]float64, error) {
	pos, _, err := m.state(id, jdTDB, false)
	return pos, err
}

// HeliocentricState 日心位置（km）与密切速度（km/日），J2000 平黄道 / heliocentric position and osculating velocity.
func (m *Model) HeliocentricState(id int, jdTDB float64) (pos, vel [3]float64, err error) {
	return m.state(id, jdTDB, true)
}

// state wantVel 为 false 时 vel 不计算、恒为零。
func (m *Model) state(id int, jdTDB float64, wantVel bool) (pos, vel [3]float64, err error) {
	if id < 1 || id > 8 {
		return pos, vel, errors.New("planets: body not in table")
	}
	bd := m.bodies[id]
	if bd == nil {
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
	if md == nil || jdTDB <= m.modernStart || jdTDB >= m.modernEnd {
		p, v := bd.ellxyz(bd.elements(t), wantVel)
		return sky(p, v)
	}
	w, dw := m.modernWeight(jdTDB)
	pm, vm := md.ellxyz(md.elements(t), wantVel)
	if w == 1 {
		return sky(pm, vm)
	}
	pl, vl := bd.ellxyz(bd.elements(t), wantVel)
	for i := range pos {
		pos[i] = pl[i] + w*(pm[i]-pl[i])
		if wantVel {
			vel[i] = vl[i] + w*(vm[i]-vl[i]) + dw*(pm[i]-pl[i])
		}
	}
	return sky(pos, vel)
}

// modernWeight 现代段权重及其对 JD 的导数；两端各 blendDays 内用 smoothstep 过渡，位置与速度均连续。
func (m *Model) modernWeight(jd float64) (w, dw float64) {
	u, sign := (jd-m.modernStart)/blendDays, 1.0
	if e := (m.modernEnd - jd) / blendDays; e < u {
		u, sign = e, -1
	}
	if u >= 1 {
		return 1, 0
	}
	return u * u * (3 - 2*u), sign * 6 * u * (1 - u) / blendDays
}

// toSky 动力学黄道 au → skyfield ecliptic_J2000 km。
func toSky(p [3]float64) (out [3]float64) {
	for i := range out {
		out[i] = (skyToVSOP[0][i]*p[0] + skyToVSOP[1][i]*p[1] + skyToVSOP[2][i]*p[2]) * auKM
	}
	return out
}

// cmul 显式转换阻止编译器融合乘加，保证与 Python 参考求值逐位一致。
func cmul(ar, ai, br, bi float64) (float64, float64) {
	return float64(ar*br) - float64(ai*bi), float64(ar*bi) + float64(ai*br)
}

// phasors 填充每个幅角的 (cos, sin)。
// tc、ts 为倍角表工作区，只读取本次写入的槽位，复用时不必清零。
func (bd *body) phasors(t float64, pc, ps, tc, ts []float64) {
	for i := range bd.tabMax {
		if bd.tabMax[i] == 0 && !(bd.top && i < 3) {
			continue
		}
		var s, c float64
		if bd.top {
			s, c = math.Sincos(topDMu * t * float64(int(1)<<uint(topBits*i)))
		} else {
			s, c = math.Sincos(ci0[i] + ci1[i]*t)
		}
		o := bd.tabOff[i]
		tc[o], ts[o] = 1, 0
		for j := 1; j <= bd.tabMax[i]; j++ {
			tc[o+j], ts[o+j] = cmul(tc[o+j-1], ts[o+j-1], c, s)
		}
		if bd.top && i == 2 {
			break
		}
	}
	for j := 0; j+1 < len(bd.keyOff); j++ {
		fs := bd.factors[bd.keyOff[j]:bd.keyOff[j+1]]
		r, im := 1.0, 0.0
		if len(fs) > 0 {
			r, im = tc[fs[0].slot], ts[fs[0].slot]
			if fs[0].neg {
				im = -im
			}
			fs = fs[1:]
		}
		for _, f := range fs {
			br, bi := tc[f.slot], ts[f.slot]
			if f.neg {
				bi = -bi
			}
			r, im = cmul(r, im, br, bi)
		}
		if bd.keyNeg[j] {
			im = -im
		}
		pc[j], ps[j] = r, im
	}
}

type scratch struct{ buf []float64 }

var scratchPool = sync.Pool{New: func() interface{} { return new(scratch) }}

// elements t 为 TDB 千年（J2000 起算）。
func (bd *body) elements(t float64) [6]float64 {
	x := (t - tMid) / tHalf
	var leg [maxDeg + 1]float64
	leg[0], leg[1] = 1, x
	for k := 2; k <= bd.legDeg; k++ {
		leg[k] = (leg[k-1]*x*float64(2*k-1) - leg[k-2]*float64(k-1)) / float64(k)
	}
	nk := len(bd.keyNeg)
	sc := scratchPool.Get().(*scratch)
	defer scratchPool.Put(sc)
	if cap(sc.buf) < 2*nk+2*maxTab {
		sc.buf = make([]float64, 2*nk+2*maxTab)
	}
	// 各段在读取前都会被完整写入，复用时不必清零。
	buf := sc.buf[:2*nk+2*maxTab]
	pc, ps := buf[:nk], buf[nk:2*nk]
	bd.phasors(t, pc, ps, buf[2*nk:2*nk+maxTab], buf[2*nk+maxTab:])
	var el [6]float64
	for v := range bd.el {
		e := &bd.el[v]
		sum := 0.0
		for i, c := range e.secular {
			sum += c * leg[i]
		}
		// 逐阶做点积再乘 Legendre 值，避免逐项累加包络系数的读改写。
		for d := range e.cnt {
			cd, sd := e.c[d], e.s[d]
			arg := e.arg[:len(cd)]
			s0, s1 := 0.0, 0.0
			for i, a := range arg {
				s0 += cd[i] * pc[a]
				s1 += sd[i] * ps[a]
			}
			sum += leg[d] * (s0 + s1)
		}
		el[v] = sum
	}
	// 先归约平经度再加周期和，与 Python 参考求值一致。
	lm := math.Mod(bd.l0+bd.f*t, 2*math.Pi)
	if lm < 0 {
		lm += 2 * math.Pi
	}
	el[1] += lm
	return el
}

// EarthFromEMB 由地月质心与地心月球合成物理地心 / Earth centre from EMB and geocentric Moon.
func EarthFromEMB(emb, moon [3]float64) [3]float64 {
	k := 1 / (1 + emrat)
	return [3]float64{emb[0] - moon[0]*k, emb[1] - moon[1]*k, emb[2] - moon[2]*k}
}

// ellxyz VSOP2013 椭圆元件 (a,l,k,h,q,p) → 动力学黄道位置（au）与速度（au/日）。
func (bd *body) ellxyz(el [6]float64, wantVel bool) (pos, vel [3]float64) {
	a, l, k, h, q, p := el[0], el[1], el[2], el[3], el[4], el[5]
	fi := math.Sqrt(1 - k*k - h*h)
	ki := math.Sqrt(1 - q*q - p*p)
	u := 1 / (1 + fi)
	ex := math.Hypot(k, h)
	gl := math.Mod(l, 2*math.Pi)
	if gl < 0 {
		gl += 2 * math.Pi
	}
	gm := gl - math.Atan2(h, k)
	e := gl + (ex-0.125*ex*ex*ex)*math.Sin(gm) + 0.5*ex*ex*math.Sin(2*gm) + 0.375*ex*ex*ex*math.Sin(3*gm)
	var zs, zc float64
	for it := 0; it < 20; it++ {
		zs, zc = math.Sincos(e)
		z3r, z3i := k*zc+h*zs, k*zs-h*zc
		dl := gl - e + z3i
		e += dl / (1 - z3r)
		if math.Abs(dl) < 1e-15 {
			break
		}
	}
	zs, zc = math.Sincos(e)
	z3r, z3i := k*zc+h*zs, k*zs-h*zc
	rsa := 1 - z3r
	z1r, z1i := u*k*z3i, u*h*z3i
	z2r, z2i := z1i, -z1r
	xcw := (-k + zc + z2r) / rsa
	xsw := (-h + zs + z2i) / rsa
	xm := p*xcw - q*xsw
	xr := a * rsa
	pos = [3]float64{xr * (xcw - 2*p*xm), xr * (xsw + 2*q*xm), -2 * xr * ki * xm}
	if !wantVel {
		return pos, vel
	}
	xms := a * (h + xsw) / fi
	xmc := a * (k + xcw) / fi
	xn := bd.rgm / math.Pow(a, 1.5)
	vel = [3]float64{xn * ((2*p*p-1)*xms + 2*p*q*xmc), xn * ((1-2*q*q)*xmc - 2*p*q*xms), 2 * xn * ki * (p*xms + q*xmc)}
	return pos, vel
}

func rx(a float64) [3][3]float64 {
	s, c := math.Sincos(a)
	return [3][3]float64{{1, 0, 0}, {0, c, -s}, {0, s, c}}
}

func rz(a float64) [3][3]float64 {
	s, c := math.Sincos(a)
	return [3][3]float64{{c, -s, 0}, {s, c, 0}, {0, 0, 1}}
}

func mul(a, b [3][3]float64) (out [3][3]float64) {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				out[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return out
}

func transpose(a [3][3]float64) (out [3][3]float64) {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			out[i][j] = a[j][i]
		}
	}
	return out
}
