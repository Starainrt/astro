package semianalytic

import (
	"math"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
)

// 视月球方向在远历元也必须是光滑函数：下游中心带牛顿迭代的残差容差约 2e-10，
// 求值噪声需留出一个量级。二阶差分 RMS 按月地距折算成 km，h=1e-7 日时真实曲率项可忽略。
func TestApparentMoonSmoothAtDistantEpochs(t *testing.T) {
	const h, n, limitKM = 1e-7, 200, 6e-5
	for _, jd0 := range []float64{2451545.0, 2460680.5, 2288000.5, 2810139.6097} {
		var prev [2][3]float64
		var sum float64
		for i := 0; i < n; i++ {
			d, _, _, err := Apparent(ephem.Moon, jd0+float64(i)*h)
			if err != nil {
				t.Fatal(err)
			}
			p := [3]float64{d[0] * 3.844e5, d[1] * 3.844e5, d[2] * 3.844e5}
			if i >= 2 {
				for k := range p {
					dd := p[k] - 2*prev[1][k] + prev[0][k]
					sum += dd * dd
				}
			}
			prev[0], prev[1] = prev[1], p
		}
		if rms := math.Sqrt(sum / (n - 2)); rms > limitKM {
			t.Errorf("jd %.1f: apparent Moon second-difference rms %.2e km, want <= %.0e", jd0, rms, limitKM)
		}
	}
}

// 地心化的光行时链与"两次日心相减"的旧链在数学上等价，只差 EMB 的二阶外推项（< 0.01 mas）。
// 旧链自身带有 jd−τ 的舍入（~1e-10 日）乘地月质心速度（2.6e6 km/日）的误差，最大约 0.35 mas，
// 因此容差取 0.5 mas；两者差值反映的是旧链误差，不是新链误差。
func TestApparentMoonMatchesHeliocentricChain(t *testing.T) {
	for _, jd := range []float64{2451545.0, 2460409.25, 2433282.5, 2810139.6097, 2086000.5} {
		got, gd, gt, err := Apparent(ephem.Moon, jd)
		if err != nil {
			t.Fatal(err)
		}
		o, err := observerAt(jd)
		if err != nil {
			t.Fatal(err)
		}
		tau := gt
		r, err := targetHelio(ephem.Moon, jd-tau, false)
		if err != nil {
			t.Fatal(err)
		}
		var g [3]float64
		for i := range g {
			g[i] = r[i] - o.p[i] - o.sunVel[i]*tau
		}
		want := toICRF(aberrate(vscale(g, 1/vlen(g)), o.v))
		cx := got[1]*want[2] - got[2]*want[1]
		cy := got[2]*want[0] - got[0]*want[2]
		cz := got[0]*want[1] - got[1]*want[0]
		c := got[0]*want[0] + got[1]*want[1] + got[2]*want[2]
		mas := math.Atan2(math.Sqrt(cx*cx+cy*cy+cz*cz), c) * 206264806.2
		if mas > 0.5 {
			t.Errorf("jd %.1f: geocentric chain differs by %.4f mas", jd, mas)
		}
		now, err := targetHelio(ephem.Moon, jd, true)
		if err != nil {
			t.Fatal(err)
		}
		if wd := vlen(vsub(now, o.p)); math.Abs(gd-wd) > 1 {
			t.Errorf("jd %.1f: geometric distance %.3f km, heliocentric chain %.3f km", jd, gd, wd)
		}
	}
}
