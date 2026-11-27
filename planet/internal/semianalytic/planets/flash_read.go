//go:build astro_lowmem
// +build astro_lowmem

package planets

import (
	"errors"
	"math"
	"math/bits"
)

var errTruncated = errors.New("planets: truncated table")

// strReader 与 reader 同格式，直接读 string，不复制表。
type strReader struct {
	s   string
	p   int
	err error
}

func (r *strReader) need(n int) bool {
	if r.err != nil || r.p+n > len(r.s) {
		r.err = errTruncated
		return false
	}
	return true
}

func (r *strReader) str(n int) string {
	if !r.need(n) {
		return ""
	}
	s := r.s[r.p : r.p+n]
	r.p += n
	return s
}

func (r *strReader) u8() uint8 {
	if !r.need(1) {
		return 0
	}
	c := r.s[r.p]
	r.p++
	return c
}

func (r *strReader) f64() float64 {
	if !r.need(8) {
		return 0
	}
	var u uint64
	for k := 7; k >= 0; k-- {
		u = u<<8 | uint64(r.s[r.p+k])
	}
	r.p += 8
	return math.Float64frombits(u)
}

func (r *strReader) uvar() uint64 {
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

func (r *strReader) zvar() int64 {
	z := r.uvar()
	if z%2 == 0 {
		return int64(z >> 1)
	}
	return -int64(z>>1) - 1
}

// strBits 与 bitReader 同语义；s 必须是单个元件的子串，越界窗口才会和 bitReader 一样补零。
type strBits struct {
	s   string
	p   int
	err error
}

func (br *strBits) window() uint64 {
	offset := br.p >> 3
	var window uint64
	if offset+8 <= len(br.s) {
		data := br.s[offset : offset+8]
		window = uint64(data[0])<<56 | uint64(data[1])<<48 |
			uint64(data[2])<<40 | uint64(data[3])<<32 |
			uint64(data[4])<<24 | uint64(data[5])<<16 |
			uint64(data[6])<<8 | uint64(data[7])
	} else {
		for index := 0; index < 8; index++ {
			window <<= 8
			if offset+index < len(br.s) {
				window |= uint64(br.s[offset+index])
			}
		}
	}
	return window << uint(br.p&7)
}

func (br *strBits) skip(n int) {
	br.p += n
	if br.p > 8*len(br.s) {
		br.err = errors.New("planets: truncated bit stream")
	}
}

func (br *strBits) get(n int) uint64 {
	if n == 0 || br.err != nil {
		return 0
	}
	if n > 56 {
		hi := br.get(n - 32)
		return hi<<32 | br.get(32)
	}
	v := br.window() >> uint(64-n)
	br.skip(n)
	return v
}

func (br *strBits) gamma() int {
	if br.err != nil {
		return 0
	}
	window := br.window()
	zeros := bits.LeadingZeros64(window)
	if zeros > 28 {
		br.err = errors.New("planets: bad gamma code")
		return 0
	}
	width := 2*zeros + 1
	if width > 56 {
		return int(br.get(width)) - 1
	}
	value := window >> uint(64-width)
	br.skip(width)
	return int(value) - 1
}

func (br *strBits) coef(width int, step float64) float64 {
	var sign uint64
	var magnitude uint64
	if width < 56 && br.err == nil && br.p < 8*len(br.s) {
		encoded := br.window() >> uint(63-width)
		br.skip(width + 1)
		sign = encoded >> uint(width)
		magnitude = encoded & (uint64(1)<<uint(width) - 1)
	} else {
		sign = br.get(1)
		magnitude = br.get(width)
	}
	value := math.Float64frombits(math.Float64bits(float64(magnitude)) | sign<<63)
	return float64(value * step)
}

func parseFlashBodies(s string) (bodies [9]*flashBody, err error) {
	r := &strReader{s: s}
	magic := r.str(4)
	if magic != "PLN2" && magic != "PLN3" {
		return bodies, errors.New("planets: flash model needs a PLN2 or PLN3 table")
	}
	packed := magic == "PLN3"
	for n := int(r.u8()); n > 0 && r.err == nil; n-- {
		id := int(r.u8())
		if id < 1 || id > 8 {
			return bodies, errors.New("planets: bad body id")
		}
		fb := &flashBody{}
		fb.top = r.u8() != 0
		secDeg := int(r.u8())
		fb.f, fb.l0 = r.f64(), r.f64()
		fb.rgm = math.Sqrt(gmp[id-1] + gmSol)
		nk := r.uvar()
		if nk > maxKeys {
			return bodies, errors.New("planets: too many arguments")
		}
		if err := fb.scanKeys(r, int(nk)); err != nil {
			return bodies, err
		}
		fb.legDeg = secDeg
		sorted := packed && r.u8() != 0
		for v := range fb.fel {
			var err error
			if packed {
				err = fb.fel[v].parse3(r, secDeg, int(nk), sorted)
			} else {
				err = fb.fel[v].parse2(r, secDeg, int(nk))
			}
			if err != nil {
				return bodies, err
			}
			if d := len(fb.fel[v].cnt) - 1; d > fb.legDeg {
				fb.legDeg = d
			}
		}
		bodies[id] = fb
	}
	if r.err != nil {
		return bodies, r.err
	}
	if r.p != len(s) {
		return bodies, errors.New("planets: trailing bytes")
	}
	return bodies, nil
}

// scanKeys 只求幂表尺寸并记下键区，因子在求值时现场解码。
func (fb *flashBody) scanKeys(r *strReader, nk int) error {
	start := r.p
	for j := 0; j < nk && r.err == nil; j++ {
		if fb.top {
			k := r.zvar()
			if k > maxKeyAbs || k < -maxKeyAbs {
				return errBadKey
			}
			a := int(k)
			if a < 0 {
				a = -a
			}
			for lvl := 0; lvl < 3; lvl++ {
				d := a >> uint(topBits*lvl)
				if lvl < 2 {
					d &= 1<<topBits - 1
				}
				if d > fb.tabMax[lvl] {
					fb.tabMax[lvl] = d
				}
			}
			continue
		}
		mask := int(r.u8()) | int(r.u8())<<8 | int(r.u8())<<16
		for i := 0; i < 17; i++ {
			if mask>>uint(i)&1 == 0 {
				continue
			}
			k := r.zvar()
			if k > maxKeyAbs || k < -maxKeyAbs {
				return errBadKey
			}
			a := int(k)
			if a < 0 {
				a = -a
			}
			if a > fb.tabMax[i] {
				fb.tabMax[i] = a
			}
		}
	}
	if r.err != nil {
		return r.err
	}
	n := 0
	for i := range fb.tabMax {
		fb.tabOff[i] = n
		if fb.tabMax[i] > 0 || (fb.top && i < 3) {
			n += fb.tabMax[i] + 1
		}
	}
	if n > maxTab {
		return errors.New("planets: power table too large")
	}
	fb.keys, fb.nk = r.s[start:r.p], nk
	return nil
}

var errUnsortedEnvelope = errors.New("planets: flash model needs terms sorted by envelope degree")

// envelopeCounts 由降序包络阶序列求 cnt[d]（包络阶 ≥ d 的项数）。
func envelopeCounts(perDeg *[maxDeg + 1]int32, top int) []int32 {
	cnt := make([]int32, top+1)
	acc := int32(0)
	for d := top; d >= 0; d-- {
		acc += perDeg[d]
		cnt[d] = acc
	}
	return cnt
}

func (e *flashElement) parse3(r *strReader, secDeg, nk int, sorted bool) error {
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
	e.step = r.f64()
	size := r.uvar()
	if r.err != nil {
		return r.err
	}
	if size > uint64(len(r.s)-r.p) {
		return errTruncated
	}
	e.src = r.str(int(size))
	e.packed, e.sorted = true, sorted
	br := &strBits{s: e.src}
	var perDeg [maxDeg + 1]int32
	total, top, prev := 0, -1, maxDeg+1
	for runs := br.gamma(); runs > 0 && br.err == nil; runs-- {
		d := int(br.get(4))
		cnt := br.gamma() + 1
		if d > maxDeg || total+cnt > nint {
			return errors.New("planets: bad envelope runs")
		}
		if d > prev {
			return errUnsortedEnvelope
		}
		if d > top {
			top = d
		}
		prev, total = d, total+cnt
		perDeg[d] += int32(cnt)
	}
	if br.err != nil {
		return br.err
	}
	if total != nint {
		return errors.New("planets: envelope count mismatch")
	}
	e.nbits = bits.Len(uint(nk - 1))
	if e.nbits == 0 {
		e.nbits = 1
	}
	e.argPos = br.p
	if nint == 0 {
		return br.err
	}
	e.cnt = envelopeCounts(&perDeg, top)
	d, last, cur := top, -1, -1
	for j := 0; j < nint; j++ {
		a := 0
		if sorted {
			for int(e.cnt[d]) <= j {
				d--
			}
			if d != cur {
				last, cur = -1, d
			}
			a = last + 1 + br.gamma()
			last = a
		} else {
			a = int(br.get(e.nbits))
		}
		if a >= nk {
			return errors.New("planets: argument index out of range")
		}
	}
	e.coefPos = br.p
	width := 0
	for _, cnt := range e.cnt {
		for i := int32(0); i < cnt; i++ {
			width += unzigzag(br.gamma())
			if width < 0 || width > 63 {
				return errors.New("planets: bad coefficient width")
			}
			br.get(1)
			br.get(width)
			br.get(1)
			br.get(width)
		}
	}
	return br.err
}

func (e *flashElement) parse2(r *strReader, secDeg, nk int) error {
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
	e.src, e.argPos = r.s, r.p
	var perDeg [maxDeg + 1]int32
	top, prev := -1, maxDeg+1
	for j := 0; j < nint; j++ {
		if r.uvar() >= uint64(nk) {
			return errors.New("planets: argument index out of range")
		}
		d := int(r.u8())
		if d > maxDeg {
			return errors.New("planets: envelope degree too high")
		}
		if d > prev {
			return errUnsortedEnvelope
		}
		if d > top {
			top = d
		}
		prev = d
		perDeg[d]++
	}
	e.step = r.f64()
	if nint == 0 {
		return r.err
	}
	e.cnt = envelopeCounts(&perDeg, top)
	e.blocks = make([]int32, 2*len(e.cnt))
	for d, cnt := range e.cnt {
		for half := 0; half < 2; half++ {
			e.blocks[2*d+half] = int32(r.p)
			for i := int32(0); i < cnt; i++ {
				r.zvar()
			}
		}
	}
	return r.err
}
