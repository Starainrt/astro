//go:build astro_lowmem
// +build astro_lowmem

package lunar

import (
	_ "embed"
	"errors"
	"math"
	"sync"
)

// FlashModel 与 Model 逐位一致，各坐标的幅角号、包络阶与系数在求值时从只读表现场读取 / bit-identical to Model, reads per-term data from the read-only table.
type FlashModel struct {
	// shell 只持有 S2、基本幅角与 pow 布局，coords 与前缀树为空，不能直接求值。
	shell  Model
	coords [3]flashCoordinate
	// mults 为表内乘数区原文；幅角按因子路径前缀有序，求值时逐个重建。
	mults string
	nargs int
	pool  sync.Pool
}

type flashScratch struct {
	pow                  []complex128
	val                  []complex128
	rate                 []float64
	leg                  []float64
	dleg                 []float64
	env, denv            [maxEnvelope + 1]float64
	stackVal             [fundCount]complex128
	stackRate            [fundCount]float64
	stackFac             [fundCount]int32
	corr, dcorr, argRate [fundCount]float64
	dc, ddc              [5]float64
	v, dv                [3]float64
	rank                 [maxEnvelope + 1]int32
}

func (fm *FlashModel) getScratch() *flashScratch {
	if s, ok := fm.pool.Get().(*flashScratch); ok {
		return s
	}
	m := &fm.shell
	return &flashScratch{
		pow:  make([]complex128, m.powLen),
		val:  make([]complex128, fm.nargs),
		rate: make([]float64, fm.nargs),
		leg:  make([]float64, m.legLen),
		dleg: make([]float64, m.legLen),
	}
}

// flashCoordinate 系数区按阶分块，每阶先 cos 块后 sin 块，块内按项序。
type flashCoordinate struct {
	n       int
	args    string
	env     string
	secular []float64
	coef    string
	// blockStart 第 d 阶 cos 块在 coef 中的 float32 下标；sin 块紧随其后，长度均为 count[d]。
	blockStart [maxEnvelope + 1]int32
	count      [maxEnvelope + 1]int32
}

type lunarStrReader struct {
	s   string
	off int
	err error
}

func (r *lunarStrReader) str(n int) string {
	if r.err != nil || r.off+n > len(r.s) {
		r.err = errors.New("lunar: truncated table")
		return ""
	}
	s := r.s[r.off : r.off+n]
	r.off += n
	return s
}

func (r *lunarStrReader) u8() uint8 {
	if s := r.str(1); s != "" {
		return s[0]
	}
	return 0
}

func (r *lunarStrReader) u16() uint16 {
	if s := r.str(2); s != "" {
		return uint16(s[0]) | uint16(s[1])<<8
	}
	return 0
}

func (r *lunarStrReader) f64() float64 {
	s := r.str(8)
	if s == "" {
		return 0
	}
	var u uint64
	for k := 7; k >= 0; k-- {
		u = u<<8 | uint64(s[k])
	}
	return math.Float64frombits(u)
}

// phasors 求各幅角的复相位与相位变化率，与 Model 前缀树逐节点同式：
// 栈第 d 层 = 第 d-1 层 × pow[因子]，只重算与上一幅角公共前缀之后的层。
func (fm *FlashModel) phasors(s *flashScratch, wantVel bool) {
	m := &fm.shell
	depth, p := 0, 0
	for a := 0; a < fm.nargs; a++ {
		mask := uint16(fm.mults[p]) | uint16(fm.mults[p+1])<<8
		p += 2
		d, same := 0, true
		for i := 0; i < fundCount; i++ {
			if mask&(1<<uint(i)) == 0 {
				continue
			}
			k := int8(fm.mults[p])
			p++
			if k == 0 {
				continue
			}
			fac := int32(m.powOff[i] + int(k))
			if same && d < depth && s.stackFac[d] == fac {
				d++
				continue
			}
			same = false
			f := s.pow[fac]
			r := float64(k) * s.argRate[i]
			if d > 0 {
				f = cmul(s.stackVal[d-1], f)
				r += s.stackRate[d-1]
			}
			s.stackVal[d], s.stackRate[d], s.stackFac[d] = f, r, fac
			d++
		}
		depth = d
		s.val[a] = s.stackVal[d-1]
		if wantVel {
			s.rate[a] = s.stackRate[d-1]
		}
	}
}

func le16(s string, i int) uint16 { return uint16(s[2*i]) | uint16(s[2*i+1])<<8 }

func lef32(s string, i int) float64 {
	s = s[4*i : 4*i+4]
	return float64(math.Float32frombits(uint32(s[0]) | uint32(s[1])<<8 | uint32(s[2])<<16 | uint32(s[3])<<24))
}

// ParseFlash 为 LM7 v2 表建立索引，表必须在模型存活期间保持不变 / index an LM7 v2 table in place.
func ParseFlash(s string) (*FlashModel, error) {
	r := &lunarStrReader{s: s}
	if r.str(4) != "LM7\x02" {
		return nil, errors.New("lunar: bad magic")
	}
	fm := &FlashModel{}
	m := &fm.shell
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
	r.f64()
	multStart := r.off
	for a := 0; a < nargs && r.err == nil; a++ {
		mask := r.u16()
		nonzero := false
		for i := 0; i < 16; i++ {
			if mask&(1<<uint(i)) == 0 {
				continue
			}
			if i >= fundCount {
				return nil, errors.New("lunar: multiplier index out of range")
			}
			k := int8(r.u8())
			nonzero = nonzero || k != 0
			if v := int(math.Abs(float64(k))); v > m.maxMult[i] {
				m.maxMult[i] = v
			}
		}
		if !nonzero && r.err == nil {
			return nil, errors.New("lunar: empty argument")
		}
	}
	if r.err != nil {
		return nil, r.err
	}
	fm.mults, fm.nargs = s[multStart:r.off], nargs
	for i := 0; i < fundCount; i++ {
		m.powOff[i] = m.powLen + m.maxMult[i]
		m.powLen += 2*m.maxMult[i] + 1
	}
	for c := 0; c < 3; c++ {
		co := &fm.coords[c]
		co.n = int(r.u16())
		co.args = r.str(2 * co.n)
		for j := 0; j < co.n && r.err == nil; j++ {
			if int(le16(co.args, j)) >= nargs {
				return nil, errors.New("lunar: argument index out of range")
			}
		}
		co.env = r.str(co.n)
		ncoef := 0
		for j := 0; j < len(co.env); j++ {
			e := int(co.env[j])
			if e > maxEnvelope {
				return nil, errors.New("lunar: envelope degree too high")
			}
			if e > m.maxEnv {
				m.maxEnv = e
			}
			for d := 0; d <= e; d++ {
				co.count[d]++
			}
			ncoef += 2 * (e + 1)
		}
		co.secular = make([]float64, secDeg+1)
		for k := range co.secular {
			co.secular[k] = r.f64()
		}
		pos := int32(0)
		for d := range co.count {
			co.blockStart[d] = pos
			pos += 2 * co.count[d]
		}
		co.coef = r.str(4 * ncoef)
	}
	if r.err != nil {
		return nil, r.err
	}
	if r.off != len(s) {
		return nil, errors.New("lunar: trailing bytes")
	}
	m.legLen = max3(s2n, secDeg+1, m.maxEnv+1)
	return fm, nil
}

// PositionJ2000 地心月球 J2000 平黄道几何位置（km）/ geocentric Moon, J2000 mean ecliptic, km.
// 可并发调用。
func (fm *FlashModel) PositionJ2000(jdTT float64) [3]float64 {
	p, _ := fm.eval(jdTT, false)
	return p
}

// PositionVelocityJ2000 位置（km）与解析速度（km/日，对 TT）/ position (km) and analytic velocity (km/day).
// 可并发调用。
func (fm *FlashModel) PositionVelocityJ2000(jdTT float64) (pos, vel [3]float64) {
	return fm.eval(jdTT, true)
}

// eval 与 Model.eval 同序同式，只有坐标求和改为按项现场取系数。
func (fm *FlashModel) eval(jdTT float64, wantVel bool) (pos, vel [3]float64) {
	m := &fm.shell
	s := fm.getScratch()
	defer fm.pool.Put(s)
	days := jdTT - j2000
	origin := days - jdTT
	daysLow := (jdTT - (days - origin)) - (j2000 + origin)
	t := days / century
	tLow := (math.FMA(-t, century, days) + daysLow) / century
	x := (t + 20) / 100
	const dxdt = 0.01
	leg, dleg := s.leg, s.dleg
	legendreD(x, leg[:m.legLen], dleg[:m.legLen], wantVel)
	dc, ddc := &s.dc, &s.ddc
	*ddc = [5]float64{}
	for j := range dc {
		dc[j] = dot(m.s2[j], leg)
		if wantVel {
			ddc[j] = dot(m.s2[j], dleg) * dxdt
		}
	}
	corr, dcorr, rate := &s.corr, &s.dcorr, &s.argRate
	*corr, *dcorr, *rate = [fundCount]float64{}, [fundCount]float64{}, [fundCount]float64{}
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
	fm.phasors(s, wantVel)
	env, denv := s.env[:m.maxEnv+1], s.denv[:m.maxEnv+1]
	legendreD(x, env, denv, wantVel)
	v, dv := &s.v, &s.dv
	for c := 0; c < 3; c++ {
		v[c], dv[c] = fm.coords[c].sum(s, leg, dleg, env, denv, wantVel)
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
	dx0 := drad*cb*cl - rad*sb*cl*dlat - rad*cb*sl*dlon
	dy0 := drad*cb*sl - rad*sb*sl*dlat + rad*cb*cl*dlon
	dz0 := drad*sb + rad*cb*dlat
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

// sum 先累加包络阶为 0 的项，再累加其余项，与 Model 的 flat0/coefH 两段同序。
func (co *flashCoordinate) sum(s *flashScratch, leg, dleg, env, denv []float64, wantVel bool) (sum, dsum float64) {
	sum = dot(co.secular, leg)
	if wantVel {
		dsum = dot(co.secular, dleg) * dxdtLunar
	}
	val, nr := s.val, s.rate
	rank0 := int32(0)
	for j := 0; j < co.n; j++ {
		if co.env[j] != 0 {
			rank0++
			continue
		}
		n := le16(co.args, j)
		z := val[n]
		a := lef32(co.coef, int(co.blockStart[0]+rank0))
		b := lef32(co.coef, int(co.blockStart[0]+co.count[0]+rank0))
		rank0++
		if wantVel {
			sum += a*real(z) + b*imag(z)
			dsum += nr[n] * (b*real(z) - a*imag(z))
		} else {
			sum += a*real(z) + b*imag(z)
		}
	}
	rank := &s.rank
	*rank = [maxEnvelope + 1]int32{}
	for j := 0; j < co.n; j++ {
		e := int(co.env[j])
		if e == 0 {
			rank[0]++
			continue
		}
		n := le16(co.args, j)
		z := val[n]
		var ac, as, dac, das float64
		for d := 0; d <= e; d++ {
			cc := lef32(co.coef, int(co.blockStart[d]+rank[d]))
			cn := lef32(co.coef, int(co.blockStart[d]+co.count[d]+rank[d]))
			rank[d]++
			ac += env[d] * cc
			as += env[d] * cn
			if wantVel {
				dac += denv[d] * cc
				das += denv[d] * cn
			}
		}
		sum += ac*real(z) + as*imag(z)
		if wantVel {
			dsum += dxdtLunar*(dac*real(z)+das*imag(z)) + nr[n]*(as*real(z)-ac*imag(z))
		}
	}
	return sum, dsum
}

const dxdtLunar = 0.01

//go:embed lunar.bin
var flashDefaultTable string

//go:embed fast.bin
var flashFastTable string

var (
	flashDefaultOnce  sync.Once
	flashDefaultModel *FlashModel
	flashFastOnce     sync.Once
	flashFastModel    *FlashModel
)

// FlashDefault 与 Default 同表的只读常驻版本 / read-only resident counterpart of Default.
func FlashDefault() *FlashModel {
	flashDefaultOnce.Do(func() {
		m, err := ParseFlash(flashDefaultTable)
		if err != nil {
			panic(err)
		}
		flashDefaultModel = m
	})
	return flashDefaultModel
}

// FlashFast 与 Fast 同表的只读常驻版本 / read-only resident counterpart of Fast.
func FlashFast() *FlashModel {
	flashFastOnce.Do(func() {
		m, err := ParseFlash(flashFastTable)
		if err != nil {
			panic(err)
		}
		flashFastModel = m
	})
	return flashFastModel
}
