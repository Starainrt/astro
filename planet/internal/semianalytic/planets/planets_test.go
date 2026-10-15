// 被测契约：内置表与参考向量一致（< 10 cm），现代段区间外回落长时段表，接缝连续，可并发求值。
package planets

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func sq(x float64) float64 { return x * x }

func longOnly(t *testing.T) *Model {
	m, err := Parse(longTable)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func checkVectors(t *testing.T, m *Model, path string) {
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	n := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		parts := strings.Split(sc.Text(), ",")
		var v [8]float64
		for i := range v {
			if v[i], err = strconv.ParseFloat(parts[i], 64); err != nil {
				t.Fatal(err)
			}
		}
		id := int(v[0])
		p, vel, err := m.HeliocentricState(id, v[1])
		if err != nil {
			t.Fatal(err)
		}
		d := math.Sqrt(sq(p[0]-v[2]) + sq(p[1]-v[3]) + sq(p[2]-v[4]))
		dv := math.Sqrt(sq(vel[0]-v[5]) + sq(vel[1]-v[6]) + sq(vel[2]-v[7]))
		if !(d <= 1e-4 && dv <= 1e-6) {
			t.Errorf("body %d jd %.1f: %.3e km, %.3e km/day", id, v[1], d, dv)
		}
		n++
	}
	if n < 100 {
		t.Fatalf("only %d vectors", n)
	}
}

func TestLongTableMatchesReference(t *testing.T) {
	checkVectors(t, longOnly(t), "testdata/long.csv")
}

func TestModernSegmentMatchesReference(t *testing.T) {
	checkVectors(t, Default(), "testdata/modern.csv")
}

func TestFastTableMatchesReference(t *testing.T) {
	checkVectors(t, Fast(), "testdata/fast.csv")
}

func TestFallsBackOutsideModern(t *testing.T) {
	long, dual := longOnly(t), Default()
	for _, jd := range []float64{ModernStart, ModernEnd, ModernStart - 1, ModernEnd + 1, 625000.5, 5000000.5} {
		for id := 1; id <= 8; id++ {
			a, _ := long.Heliocentric(id, jd)
			b, _ := dual.Heliocentric(id, jd)
			if a != b {
				t.Errorf("body %d jd %.1f: did not fall back to long table", id, jd)
			}
		}
	}
}

func TestSeamContinuous(t *testing.T) {
	m := Default()
	const h = 0.5
	for _, jd := range []float64{ModernStart, ModernEnd, ModernStart + blendDays/2, ModernEnd - blendDays/2} {
		for id := 1; id <= 8; id++ {
			a, _ := m.Heliocentric(id, jd-h)
			b, _ := m.Heliocentric(id, jd+h)
			_, v, _ := m.HeliocentricState(id, jd)
			var d, vn float64
			for i := range v {
				d += sq((b[i]-a[i])/(2*h) - v[i])
				vn += sq(v[i])
			}
			if d, vn = math.Sqrt(d), math.Sqrt(vn); d > 1e-3*vn {
				t.Errorf("body %d jd %.1f: velocity differs from finite difference by %.3f of %.0f km/day", id, jd, d, vn)
			}
		}
	}
}

func TestParseRejectsTruncated(t *testing.T) {
	for _, b := range [][]byte{longTable, modernTable} {
		for _, n := range []int{0, 3, 5, 40, len(b) / 2, len(b) - 1} {
			if _, err := Parse(b[:n]); err == nil {
				t.Errorf("len %d: expected error", n)
			}
		}
	}
}

func TestConcurrentMatchesSerial(t *testing.T) {
	m := Default()
	const n = 64
	var want [n][3]float64
	for i := range want {
		want[i], _ = m.Heliocentric(1+i%8, 2451545.0+float64(i)*977.3)
	}
	var wg sync.WaitGroup
	var bad [8]bool
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < n; i++ {
				if got, _ := m.Heliocentric(1+i%8, 2451545.0+float64(i)*977.3); got != want[i] {
					bad[g] = true
				}
			}
		}(g)
	}
	wg.Wait()
	for g, b := range bad {
		if b {
			t.Fatalf("goroutine %d saw a mismatch", g)
		}
	}
}

func BenchmarkHeliocentricState(b *testing.B) {
	m := Default()
	for id := 1; id <= 8; id++ {
		id := id
		b.Run(strconv.Itoa(id), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				m.HeliocentricState(id, 2378496.0+float64(i%14610)*10)
			}
		})
	}
}

func TestZigzagDecodeBeyond32Bits(t *testing.T) {
	for _, want := range []int64{0, -1, 1, 1<<31 - 1, -1 << 31, 1 << 40, -(1 << 40) - 7, 1<<55 + 3} {
		z := uint64(want<<1) ^ uint64(want>>63)
		var b []byte
		for z >= 0x80 {
			b = append(b, byte(z)|0x80)
			z >>= 7
		}
		b = append(b, byte(z))
		r := &reader{b: b}
		if got := r.zvar(); got != want || r.err != nil || r.p != len(b) {
			t.Errorf("zvar(%d) = %d, err %v, read %d/%d", want, got, r.err, r.p, len(b))
		}
	}
}
