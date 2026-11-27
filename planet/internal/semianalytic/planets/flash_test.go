//go:build astro_lowmem
// +build astro_lowmem

// 被测契约：FlashModel 与同表 Model 的位置、速度逐位一致，含现代段接缝与 ±1 万年两端。
package planets

import (
	"math"
	"reflect"
	"testing"
	"unsafe"
)

func flashSampleJDs() []float64 {
	var jds []float64
	for jd := j2000 - 3652500.0; jd <= j2000+3652500.0; jd += 36524.25 * 0.97 {
		jds = append(jds, jd)
	}
	for _, edge := range []float64{ModernStart, ModernEnd} {
		for k := -3; k <= 3; k++ {
			jds = append(jds, edge+float64(k)*blendDays/3)
		}
	}
	for jd := 2451545.0 - 50000; jd <= 2451545.0+50000; jd += 997.3 {
		jds = append(jds, jd)
	}
	return jds
}

type heliocentricStater interface {
	HeliocentricState(id int, jdTDB float64) (pos, vel [3]float64, err error)
}

func assertSameStates(t *testing.T, name string, want, got heliocentricStater) {
	n := 0
	for id := 0; id <= 9; id++ {
		for _, jd := range flashSampleJDs() {
			wp, wv, werr := want.HeliocentricState(id, jd)
			gp, gv, gerr := got.HeliocentricState(id, jd)
			if (werr == nil) != (gerr == nil) {
				t.Fatalf("%s id=%d jd=%.3f: err %v vs %v", name, id, jd, werr, gerr)
			}
			for i := 0; i < 3; i++ {
				if math.Float64bits(wp[i]) != math.Float64bits(gp[i]) || math.Float64bits(wv[i]) != math.Float64bits(gv[i]) {
					t.Fatalf("%s id=%d jd=%.3f axis %d: pos %v vs %v, vel %v vs %v", name, id, jd, i, wp[i], gp[i], wv[i], gv[i])
				}
			}
			n++
		}
	}
	if n < 1000 {
		t.Fatalf("%s: only %d samples", name, n)
	}
}

func TestFlashDefaultMatchesDefault(t *testing.T) {
	assertSameStates(t, "default", Default(), FlashDefault())
}

func TestFlashFastMatchesFast(t *testing.T) {
	assertSameStates(t, "fast", Fast(), FlashFast())
}

func TestFlashLongOnlyMatchesLongOnly(t *testing.T) {
	m, err := ParseFlash(flashLong)
	if err != nil {
		t.Fatal(err)
	}
	assertSameStates(t, "long", longOnly(t), m)
}

func TestFlashExpandedAddressIndependent(t *testing.T) {
	want := Default()
	jds := []float64{
		j2000 - 3652500,
		ModernStart - blendDays,
		ModernStart,
		ModernStart + blendDays,
		j2000,
		ModernEnd - blendDays,
		ModernEnd,
		ModernEnd + blendDays,
		j2000 + 3652500,
	}
	for offset := 0; offset < 8; offset++ {
		got, err := ParseFlash(flashLong)
		if err != nil {
			t.Fatal(err)
		}
		if err := got.WithModernSegment(flashModern, ModernStart, ModernEnd); err != nil {
			t.Fatal(err)
		}
		long := shiftedExpandedTable(flashExpandedLong, offset)
		modern := shiftedExpandedTable(flashExpandedModern, offset)
		attachExpanded(got.bodies, long, expandedLong)
		attachExpanded(got.modernBodies, modern, expandedModern)
		for id := 1; id <= 8; id++ {
			for _, jd := range jds {
				wp, wv, werr := want.HeliocentricState(id, jd)
				gp, gv, gerr := got.HeliocentricState(id, jd)
				if werr != nil || gerr != nil {
					t.Fatalf("offset %d id=%d jd=%.3f: %v / %v", offset, id, jd, werr, gerr)
				}
				for axis := 0; axis < 3; axis++ {
					if math.Float64bits(wp[axis]) != math.Float64bits(gp[axis]) || math.Float64bits(wv[axis]) != math.Float64bits(gv[axis]) {
						t.Fatalf("offset %d id=%d jd=%.3f axis=%d: pos %v / %v vel %v / %v", offset, id, jd, axis, wp[axis], gp[axis], wv[axis], gv[axis])
					}
				}
			}
		}
	}
}

func shiftedExpandedTable(table string, offset int) string {
	backing := make([]byte, len(table)+offset)
	copy(backing[offset:], table)
	whole := string(backing)
	return whole[offset:]
}

func TestShiftedExpandedTableUsesRequestedOffsets(t *testing.T) {
	for offset := 0; offset < 8; offset++ {
		table := shiftedExpandedTable(flashExpandedLong, offset)
		address := (*reflect.StringHeader)(unsafe.Pointer(&table)).Data
		if len(table) != len(flashExpandedLong) {
			t.Fatalf("offset %d: length %d", offset, len(table))
		}
		if address&7 != uintptr(offset) {
			t.Fatalf("offset %d: address alignment %d", offset, address&7)
		}
	}
}

func TestParseFlashRejectsBadInput(t *testing.T) {
	for _, s := range []string{"", "PLN1\x00", "PLN3", flashFast[:len(flashFast)/2], flashFast + "x"} {
		if _, err := ParseFlash(s); err == nil {
			t.Errorf("ParseFlash(len %d) accepted bad input", len(s))
		}
	}
}
