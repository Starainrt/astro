package planet

import "math"

// legacy.bin 格式：魔数 "LGC1"，随后依次是 VSOP 表、ELP 表；每张表先写 uvarint 维数与各级长度，再写小端 float64。

// legacyTableError 用零尺寸类型而非 errors.New，避免给包初始化增加代码。
type legacyTableError struct{}

func (legacyTableError) Error() string { return "planet: bad legacy table" }

var errLegacyTable error = legacyTableError{}

type legacyReader struct {
	s   string
	p   int
	err error
}

func (r *legacyReader) uvar() int {
	if r.err != nil {
		return 0
	}
	n, shift := uint64(0), uint(0)
	for r.p < len(r.s) && shift <= 56 {
		c := r.s[r.p]
		r.p++
		n |= uint64(c&0x7F) << shift
		if c < 0x80 {
			if n > uint64(len(r.s)) {
				r.err = errLegacyTable
				return 0
			}
			return int(n)
		}
		shift += 7
	}
	r.err = errLegacyTable
	return 0
}

func (r *legacyReader) floats(n int) []float64 {
	if r.err != nil || n > (len(r.s)-r.p)/8 {
		r.err = errLegacyTable
		return nil
	}
	out := make([]float64, n)
	for i := range out {
		var u uint64
		for k := 7; k >= 0; k-- {
			u = u<<8 | uint64(r.s[r.p+k])
		}
		r.p += 8
		out[i] = math.Float64frombits(u)
	}
	return out
}

func decodeLegacyTables(s string) (raw [][]float64, cir [][][]float64, err error) {
	r := &legacyReader{s: s}
	if len(s) < 4 || s[:4] != "LGC1" {
		return nil, nil, errLegacyTable
	}
	r.p = 4
	raw = make([][]float64, r.uvar())
	for i := range raw {
		raw[i] = r.floats(r.uvar())
	}
	cir = make([][][]float64, r.uvar())
	for i := range cir {
		cir[i] = make([][]float64, r.uvar())
		for j := range cir[i] {
			cir[i][j] = r.floats(r.uvar())
		}
	}
	if r.err == nil && r.p != len(s) {
		r.err = errLegacyTable
	}
	if r.err != nil {
		return nil, nil, r.err
	}
	return raw, cir, nil
}
