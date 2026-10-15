// 被测契约：内置表经 ICRF 适配后与 DE441 几何状态一致到表本身的精度；质心推算误差不影响光行差。
package semianalytic

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
)

type sample struct {
	target, center ephem.Body
	jd             float64
	p, v           [3]float64
}

func loadSamples(t *testing.T) []sample {
	f, err := os.Open("testdata/de441-states.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var out []sample
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		parts := strings.Split(sc.Text(), ",")
		var v [9]float64
		for i := range v {
			if v[i], err = strconv.ParseFloat(parts[i], 64); err != nil {
				t.Fatal(err)
			}
		}
		out = append(out, sample{ephem.Body(v[0]), ephem.Body(v[1]), v[2], [3]float64{v[3], v[4], v[5]}, [3]float64{v[6], v[7], v[8]}})
	}
	if len(out) < 500 {
		t.Fatalf("only %d samples", len(out))
	}
	return out
}

func norm(a, b [3]float64) float64 {
	return math.Sqrt((a[0]-b[0])*(a[0]-b[0]) + (a[1]-b[1])*(a[1]-b[1]) + (a[2]-b[2])*(a[2]-b[2]))
}

func modern(jd float64) bool { return jd >= 2086302.5 && jd <= 2816787.5 }

func TestStatesMatchDE441(t *testing.T) {
	type key struct {
		target, center ephem.Body
		modern         bool
	}
	type acc struct{ n, dp, dv, maxDp, maxDv, maxAng float64 }
	stats := map[key]*acc{}
	e := Ephemeris{}
	for _, s := range loadSamples(t) {
		p, v, err := e.State(s.target, s.center, s.jd)
		if err != nil {
			t.Fatal(err)
		}
		k := key{s.target, s.center, modern(s.jd)}
		if stats[k] == nil {
			stats[k] = &acc{}
		}
		a := stats[k]
		dp, dv := norm(p, s.p), norm(v, s.v)/86.4
		r := math.Sqrt(s.p[0]*s.p[0] + s.p[1]*s.p[1] + s.p[2]*s.p[2])
		a.n++
		a.dp += dp * dp
		a.dv += dv * dv
		a.maxDp = math.Max(a.maxDp, dp)
		a.maxDv = math.Max(a.maxDv, dv)
		a.maxAng = math.Max(a.maxAng, dp/r*206264.806)
	}
	// 上限：最大位置差 km、最大速度差 m/s；太阳相对质心的残差来自表外天体（冥王星、阋神星等）。
	type limit struct{ km, ms float64 }
	limits := map[key]limit{
		{ephem.Sun, ephem.SSB, true}: {300, 1e-3}, {ephem.Sun, ephem.SSB, false}: {300, 1e-3},
		{ephem.Earth, ephem.SSB, true}: {300, 0.01}, {ephem.Earth, ephem.SSB, false}: {500, 0.05},
		{ephem.Earth, ephem.Sun, true}: {40, 0.01}, {ephem.Earth, ephem.Sun, false}: {250, 0.05},
		{ephem.Moon, ephem.Earth, true}: {0.6, 0.005}, {ephem.Moon, ephem.Earth, false}: {1.2, 0.008},
		{ephem.Mars, ephem.SSB, true}: {400, 0.04}, {ephem.Mars, ephem.SSB, false}: {1200, 0.1},
		{ephem.Jupiter, ephem.SSB, true}: {2000, 0.03}, {ephem.Jupiter, ephem.SSB, false}: {4000, 0.05},
		{ephem.Saturn, ephem.SSB, true}: {2500, 0.02}, {ephem.Saturn, ephem.SSB, false}: {7000, 0.04},
	}
	for k, a := range stats {
		t.Logf("%3d-%3d modern=%-5v n=%2.0f pos rms %.3g km max %.3g km (%.4f″)  vel rms %.3g m/s max %.3g m/s",
			k.target, k.center, k.modern, a.n, math.Sqrt(a.dp/a.n), a.maxDp, a.maxAng, math.Sqrt(a.dv/a.n), a.maxDv)
		l, ok := limits[k]
		if !ok {
			t.Errorf("%d-%d modern=%v: no limit defined", k.target, k.center, k.modern)
			continue
		}
		if a.maxDp > l.km || a.maxDv > l.ms {
			t.Errorf("%d-%d modern=%v: max %.3g km / %.3g m/s exceeds %.3g km / %.3g m/s",
				k.target, k.center, k.modern, a.maxDp, a.maxDv, l.km, l.ms)
		}
	}
}

func TestStateAntisymmetricAndComposable(t *testing.T) {
	e := Ephemeris{}
	bodies := []ephem.Body{ephem.SSB, ephem.Sun, ephem.Mercury, ephem.EMB, ephem.Earth, ephem.Moon, ephem.Neptune}
	for _, jd := range []float64{2451545.0, 2460000.5, 1000000.5} {
		for _, a := range bodies {
			for _, b := range bodies {
				p1, v1, err1 := e.State(a, b, jd)
				p2, v2, err2 := e.State(b, a, jd)
				if err1 != nil || err2 != nil {
					t.Fatal(err1, err2)
				}
				for i := range p1 {
					if math.Abs(p1[i]+p2[i]) > 1e-6 || math.Abs(v1[i]+v2[i]) > 1e-9 {
						t.Fatalf("%d-%d jd %.1f: not antisymmetric", a, b, jd)
					}
				}
			}
			// 经质心中转与直接求差一致
			ps, _, _ := e.State(a, ephem.SSB, jd)
			pe, _, _ := e.State(ephem.Earth, ephem.SSB, jd)
			pd, _, _ := e.State(a, ephem.Earth, jd)
			for i := range pd {
				if math.Abs(ps[i]-pe[i]-pd[i]) > 1e-4 {
					t.Fatalf("%d via SSB differs from direct by %.3g km", a, ps[i]-pe[i]-pd[i])
				}
			}
		}
	}
}

func TestUnsupportedBodyAndCoverage(t *testing.T) {
	e := Ephemeris{}
	if _, _, err := e.State(ephem.Body(999), ephem.Sun, 2451545); err != ephem.ErrUnsupportedBody {
		t.Fatalf("want ErrUnsupportedBody, got %v", err)
	}
	if _, _, ok := e.Coverage(ephem.Body(999)); ok {
		t.Fatal("unsupported body reported coverage")
	}
	s, end, ok := e.Coverage(ephem.Moon)
	if !ok || s >= 2086302.5 || end <= 2816787.5 {
		t.Fatalf("coverage %v..%v ok=%v", s, end, ok)
	}
}

func BenchmarkEarthBarycentric(b *testing.B) {
	e := Ephemeris{}
	for i := 0; i < b.N; i++ {
		e.State(ephem.Earth, ephem.SSB, 2451545.0+float64(i%7300)*10)
	}
}
