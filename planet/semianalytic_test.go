// 被测契约：半解析模型在 1800–2200 年与旧级数相差不超过旧级数自身误差量级；n ≥ 0 走快速档且不劣于旧级数 n=16；越界返回 NaN，切回后逐位还原。
package planet

import (
	"math"
	"strconv"
	"testing"

	"github.com/starainrt/astro/internal/ephem"

	"github.com/starainrt/astro/internal/ephemtest"
)

func useSemianalytic(t testing.TB) {
	prev := ephem.Active()
	ephem.SetActive(ephem.Semianalytic)
	t.Cleanup(func() { ephem.SetActive(prev) })
}

func angleDiffArcsec(a, b float64) float64 {
	return math.Abs(math.Mod(a-b+540, 360)-180) * 3600
}

type lbrDeg struct{ l, b, r float64 }

func sampleAll(jds []float64) map[[2]int]lbrDeg {
	out := map[[2]int]lbrDeg{}
	for i, jd := range jds {
		for xt := -1; xt <= 7; xt++ {
			out[[2]int{xt, i}] = lbrDeg{WherePlanet(xt, 0, jd), WherePlanet(xt, 1, jd), WherePlanet(xt, 2, jd)}
		}
		out[[2]int{-9, i}] = lbrDeg{
			math.Mod(WhereMoon(0, jd)*180/math.Pi, 360), WhereMoon(1, jd) * 180 / math.Pi, WhereMoon(2, jd),
		}
	}
	return out
}

func TestSemianalyticNearLegacyModernEra(t *testing.T) {
	var jds []float64
	for jd := 2378496.5; jd <= 2524593.5; jd += 1461.37 {
		jds = append(jds, jd)
	}
	ephemtest.PinLegacy(t)
	legacy := sampleAll(jds)
	useSemianalytic(t)
	semi := sampleAll(jds)
	// 角秒；距离 AU（行星）或 km（月球，-9）。
	tolA := map[int]float64{-1: 0.5, 0: 0.5, 1: 1, 2: 1, 3: 2, 4: 2, 5: 3, 6: 6, 7: 10, -9: 10}
	tolR := map[int]float64{-1: 2e-7, 0: 2e-7, 1: 2e-7, 2: 3e-7, 3: 1e-6, 4: 4e-6, 5: 1e-5, 6: 2e-4, 7: 2e-4, -9: 5}
	worst := map[int][3]float64{}
	for k, want := range legacy {
		got := semi[k]
		w := worst[k[0]]
		w[0] = math.Max(w[0], angleDiffArcsec(got.l, want.l)*math.Cos(want.b*math.Pi/180))
		w[1] = math.Max(w[1], math.Abs(got.b-want.b)*3600)
		w[2] = math.Max(w[2], math.Abs(got.r-want.r))
		worst[k[0]] = w
	}
	for xt, w := range worst {
		t.Logf("xt %d: max ΔL %.3f″ ΔB %.3f″ ΔR %.3g", xt, w[0], w[1], w[2])
		if w[0] > tolA[xt] || w[1] > tolA[xt] || w[2] > tolR[xt] {
			t.Errorf("xt %d differs from legacy beyond expected truncation error", xt)
		}
	}
}

func TestSemianalyticDomainAndRestore(t *testing.T) {
	ephemtest.PinLegacy(t)
	jd := 2460000.5
	before := WherePlanetN(3, 0, jd, 16)
	moonBefore := WhereMoonN(0, jd, 8)
	useSemianalytic(t)
	for _, c := range [][2]int{{-2, 0}, {8, 0}, {0, -1}, {0, 3}} {
		if v := WherePlanet(c[0], c[1], jd); !math.IsNaN(v) {
			t.Errorf("xt %d zn %d: want NaN, got %v", c[0], c[1], v)
		}
	}
	if WherePlanetN(3, 0, jd, 16) != WherePlanetN(3, 0, jd, 8) || WhereMoonN(0, jd, 16) != WhereMoonN(0, jd, 0) {
		t.Error("every n >= 0 must use the same fast tier")
	}
	if WherePlanetN(3, 0, jd, 16) == WherePlanetN(3, 0, jd, -1) {
		t.Error("n >= 0 did not switch to the fast tier")
	}
	ephem.SetActive(ephem.Legacy)
	if WherePlanetN(3, 0, jd, 16) != before || WhereMoonN(0, jd, 8) != moonBefore {
		t.Error("switching back did not restore legacy values bit for bit")
	}
}

// 快速档相对全表的最大差（角秒；月球另有 km），全表本身对 DE441 已在 0.02″ 内。
func TestSemianalyticFastTierNearFull(t *testing.T) {
	useSemianalytic(t)
	var worst [9]float64
	worstKM := 0.0
	for jd := -1931455.0; jd <= 5373545.0; jd += 3652.5 * 7.3 {
		for xt := 0; xt <= 7; xt++ {
			d := angleDiffArcsec(WherePlanetN(xt, 0, jd, 0), WherePlanetN(xt, 0, jd, -1))*math.Cos(WherePlanetN(xt, 1, jd, -1)*math.Pi/180) +
				math.Abs(WherePlanetN(xt, 1, jd, 0)-WherePlanetN(xt, 1, jd, -1))*3600
			worst[xt] = math.Max(worst[xt], d)
		}
		ml := angleDiffArcsec(WhereMoonN(0, jd, 0)*180/math.Pi, WhereMoonN(0, jd, -1)*180/math.Pi) +
			math.Abs(WhereMoonN(1, jd, 0)-WhereMoonN(1, jd, -1))*206264.806
		worst[8] = math.Max(worst[8], ml)
		worstKM = math.Max(worstKM, math.Abs(WhereMoonN(2, jd, 0)-WhereMoonN(2, jd, -1)))
	}
	limits := [9]float64{2, 0.6, 1, 8, 15, 60, 75, 20, 60}
	for i, w := range worst {
		t.Logf("slot %d: max %.2f″", i, w)
		if w > limits[i] {
			t.Errorf("slot %d: fast tier differs from full table by %.2f″ (limit %.0f″)", i, w, limits[i])
		}
	}
	if worstKM > 20 {
		t.Errorf("moon distance: fast tier differs by %.1f km", worstKM)
	}
}

func BenchmarkWherePlanetSemianalytic(b *testing.B) {
	useSemianalytic(b)
	for _, n := range []int{-1, 16} {
		n := n
		b.Run("n="+strconv.Itoa(n), func(b *testing.B) {
			for xt := 0; xt <= 7; xt++ {
				xt := xt
				b.Run(strconv.Itoa(xt), func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						jd := 2451545.0 + float64(i%7300)*100
						_ = WherePlanetN(xt, 0, jd, n) + WherePlanetN(xt, 1, jd, n) + WherePlanetN(xt, 2, jd, n)
					}
				})
			}
			b.Run("moon", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					jd := 2451545.0 + float64(i%7300)*10.3
					_ = WhereMoonN(0, jd, n) + WhereMoonN(1, jd, n) + WhereMoonN(2, jd, n)
				}
			})
		})
	}
}
