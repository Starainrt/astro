package planets

import (
	"encoding/binary"
	"errors"
	"math/bits"
)

// bitReader 按位读取，字节内高位在前。
type bitReader struct {
	b   []byte
	p   int // 位位置
	err error
}

// window 返回从当前位开始的 64 位窗口（高位对齐，越界补零），有效位至少 57 位。
func (br *bitReader) window() uint64 {
	i := br.p >> 3
	var w uint64
	if i+8 <= len(br.b) {
		w = binary.BigEndian.Uint64(br.b[i:])
	} else {
		for k := 0; k < 8; k++ {
			w <<= 8
			if i+k < len(br.b) {
				w |= uint64(br.b[i+k])
			}
		}
	}
	return w << uint(br.p&7)
}

func (br *bitReader) skip(n int) {
	br.p += n
	if br.p > 8*len(br.b) {
		br.err = errors.New("planets: truncated bit stream")
	}
}

func (br *bitReader) get(n int) uint64 {
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

// gamma Elias-gamma，编码值为 n+1。
func (br *bitReader) gamma() int {
	if br.err != nil {
		return 0
	}
	z := bits.LeadingZeros64(br.window())
	if z > 28 {
		br.err = errors.New("planets: bad gamma code")
		return 0
	}
	v := br.get(2*z + 1)
	return int(v) - 1
}

func unzigzag(z int) int {
	if z&1 == 0 {
		return z >> 1
	}
	return -(z >> 1) - 1
}

// readElement3 读取 PLN3 元件位流：包络游程、幅角号（定长或同阶升序间隔）、成对共享位宽的量化系数。
func readElement3(r *reader, e *element, secDeg, nk int, sortedArgs bool) error {
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
	step := r.f64()
	size := r.uvar()
	if r.err != nil {
		return r.err
	}
	if size > uint64(len(r.b)-r.p) {
		return errors.New("planets: truncated table")
	}
	br := &bitReader{b: r.take(int(size))}
	if r.err != nil {
		return r.err
	}
	env := make([]int, 0, nint)
	for runs := br.gamma(); runs > 0 && br.err == nil; runs-- {
		d := int(br.get(4))
		cnt := br.gamma() + 1
		if d > maxDeg || len(env)+cnt > nint {
			return errors.New("planets: bad envelope runs")
		}
		for ; cnt > 0; cnt-- {
			env = append(env, d)
		}
	}
	if br.err != nil {
		return br.err
	}
	if len(env) != nint {
		return errors.New("planets: envelope count mismatch")
	}
	args := make([]int32, nint)
	nbits := bits.Len(uint(nk - 1))
	if nbits == 0 {
		nbits = 1
	}
	prev, cur := -1, -1
	for j := 0; j < nint; j++ {
		a := 0
		if sortedArgs {
			if env[j] != cur {
				prev, cur = -1, env[j]
			}
			a = prev + 1 + br.gamma()
			prev = a
		} else {
			a = int(br.get(nbits))
		}
		if a >= nk {
			return errors.New("planets: argument index out of range")
		}
		args[j] = int32(a)
	}
	top := 0
	for _, d := range env {
		if d > top {
			top = d
		}
	}
	// 按包络阶降序的稳定排序：pos[j] 为位流第 j 项的新位置
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
	if nint == 0 {
		return br.err
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
	}
	width := 0
	coef := func() float64 {
		neg := br.get(1) != 0
		v := float64(br.get(width))
		if neg {
			v = -v
		}
		return v * step
	}
	for d := 0; d <= top; d++ {
		for j := 0; j < nint; j++ {
			if env[j] < d {
				continue
			}
			width += unzigzag(br.gamma())
			if width < 0 || width > 63 {
				return errors.New("planets: bad coefficient width")
			}
			e.c[d][pos[j]] = coef()
			e.s[d][pos[j]] = coef()
		}
	}
	return br.err
}
