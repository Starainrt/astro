// 被测契约：内置表与参考向量一致（< 0.1 m），解析速度与差分一致，可并发求值。
package lunar

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

func TestMatchesReferenceVectors(t *testing.T) { checkVectors(t, Default(), "testdata/vectors.csv") }

func TestFastMatchesReferenceVectors(t *testing.T) { checkVectors(t, Fast(), "testdata/fast.csv") }

func checkVectors(t *testing.T, m *Model, path string) {
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	n, worst := 0, 0.0
	for sc.Scan() {
		parts := strings.Split(sc.Text(), ",")
		var v [4]float64
		for i := range v {
			if v[i], err = strconv.ParseFloat(parts[i], 64); err != nil {
				t.Fatal(err)
			}
		}
		got := m.PositionJ2000(v[0])
		if d := math.Sqrt(sq(got[0]-v[1]) + sq(got[1]-v[2]) + sq(got[2]-v[3])); d > worst {
			worst = d
		}
		n++
	}
	if n < 40 {
		t.Fatalf("only %d vectors", n)
	}
	if worst > 1e-4 {
		t.Fatalf("max diff %.3e km over %d vectors", worst, n)
	}
}

func TestVelocityMatchesFiniteDifference(t *testing.T) {
	m := Default()
	const h = 0.05
	worst := 0.0
	for _, jd := range []float64{990545.5, 1500000.25, 2299160.5, 2451545.0, 2460676.5, 2816787.5, 3912545.0} {
		p, v := m.PositionVelocityJ2000(jd)
		if q := m.PositionJ2000(jd); q != p {
			t.Fatalf("jd=%.1f position mismatch", jd)
		}
		a, b := m.PositionJ2000(jd+h), m.PositionJ2000(jd-h)
		a2, b2 := m.PositionJ2000(jd+2*h), m.PositionJ2000(jd-2*h)
		for i := range v {
			fd := (8*(a[i]-b[i]) - (a2[i] - b2[i])) / (12 * h)
			if d := math.Abs(fd - v[i]); d > worst {
				worst = d
			}
		}
	}
	if worst > 2e-3 {
		t.Fatalf("max velocity diff %.3e km/day", worst)
	}
}

func TestParseRejectsTruncated(t *testing.T) {
	for _, n := range []int{0, 3, 5, 40, len(defaultTable) / 2, len(defaultTable) - 1} {
		if _, err := Parse(defaultTable[:n]); err == nil {
			t.Errorf("len %d: expected error", n)
		}
	}
}

func TestConcurrentEvaluation(t *testing.T) {
	m := Default()
	const n = 64
	var want [n][3]float64
	for i := range want {
		want[i] = m.PositionJ2000(2451545.0 + float64(i)*13.7)
	}
	var wg sync.WaitGroup
	var bad [8]bool
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for k := 0; k < n; k++ {
				i := (k + g*7) % n
				if m.PositionJ2000(2451545.0+float64(i)*13.7) != want[i] {
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

func BenchmarkPositionJ2000(b *testing.B) {
	m := Default()
	for i := 0; i < b.N; i++ {
		_ = m.PositionJ2000(2460000.5 + float64(i%10000)*0.37)
	}
}
