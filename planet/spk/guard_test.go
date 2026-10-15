// 被测契约：畸形阶数与非有限系数报 ErrInvalidFile 而不是 panic 或返回 NaN；无关段的覆盖缺口不阻断有效路径。
package spk

import (
	"encoding/binary"
	"errors"
	"math"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
)

func type21Record(maxdim, kqmax1, kq int) []float64 {
	record := make([]float64, 4*maxdim+12)
	record[0] = float64(maxdim)
	for index := 0; index < maxdim; index++ {
		record[2+index] = 1
	}
	record[4*maxdim+8] = float64(kqmax1)
	for component := 0; component < 3; component++ {
		record[4*maxdim+9+component] = float64(kq)
	}
	return record
}

func TestType21RejectsInconsistentOrders(t *testing.T) {
	cases := []struct{ maxdim, kqmax1, kq int }{
		{15, 3, 15},
		{15, 27, 1},
		{15, 17, 1},
		{15, 5, 5},
	}
	for _, c := range cases {
		_, _, err := evaluateType21(0, type21Record(c.maxdim, c.kqmax1, c.kq))
		if !errors.Is(err, ErrInvalidFile) {
			t.Errorf("maxdim=%d kqmax1=%d kq=%d: err = %v, want %v", c.maxdim, c.kqmax1, c.kq, err, ErrInvalidFile)
		}
	}
	if _, _, err := evaluateType21(0, type21Record(15, 16, 15)); err != nil {
		t.Errorf("maximal valid orders: err = %v", err)
	}
}

func TestType21RejectsNonFiniteRecord(t *testing.T) {
	record := type21Record(15, 3, 1)
	record[15+2] = math.NaN()
	if _, _, err := evaluateType21(0, record); !errors.Is(err, ErrInvalidFile) {
		t.Errorf("err = %v, want %v", err, ErrInvalidFile)
	}
}

func TestType2RejectsNonFiniteCoefficient(t *testing.T) {
	path := writeTestSPK(t, binary.LittleEndian, "LTL-IEEE",
		testSPKSegment{target: Mars, center: Sun, frame: 1, typ: 2, start: 0, end: 10, words: type2Words(0, 10, []float64{5, 5, math.NaN(), 20, 30})},
	)
	file, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if _, _, err := file.State(Mars, Sun, spkJ2000+5/spkSecondsDay); !errors.Is(err, ErrInvalidFile) {
		t.Errorf("err = %v, want %v", err, ErrInvalidFile)
	}
}

func TestUnrelatedSegmentCoverageDoesNotBlockDirectPath(t *testing.T) {
	path := writeTestSPK(t, binary.LittleEndian, "LTL-IEEE",
		testSPKSegment{target: Mars, center: Sun, frame: 1, typ: 2, start: 0, end: 10, words: type2Words(0, 10, []float64{5, 5, 10, 20, 30})},
		testSPKSegment{target: Sun, center: SSB, frame: 1, typ: 2, start: 100, end: 110, words: type2Words(100, 10, []float64{105, 5, 1, 2, 3})},
	)
	file, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	pos, _, err := file.State(Mars, Sun, spkJ2000+5/spkSecondsDay)
	if err != nil {
		t.Fatalf("Mars→Sun err = %v", err)
	}
	assertVector(t, pos, [3]float64{10, 20, 30})
	if _, _, err := file.State(Mars, SSB, spkJ2000+5/spkSecondsDay); err != ephem.ErrOutOfRange {
		t.Errorf("Mars→SSB err = %v, want %v", err, ephem.ErrOutOfRange)
	}
}
