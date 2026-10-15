package semianalytic

import (
	"math"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
)

// stateOnly 只暴露几何状态，迫使实例走通用矢量链。
type stateOnly struct{ Ephemeris }

func (e stateOnly) State(t, c ephem.Body, jd float64) ([3]float64, [3]float64, error) {
	return e.Ephemeris.State(t, c, jd)
}
func (e stateOnly) Coverage(t ephem.Body) (float64, float64, bool) { return e.Ephemeris.Coverage(t) }

func TestGenericApparentMatchesNative(t *testing.T) {
	inst := ephem.NewInstance(stateOnly{})
	worst := 0.0
	for _, jd := range []float64{2451545.0, 2460409.25, 2433282.5, 2488069.5} {
		for _, b := range []ephem.Body{ephem.Sun, ephem.Moon, ephem.Mercury, ephem.Venus, ephem.Mars, ephem.Jupiter, ephem.Saturn, ephem.Uranus, ephem.Neptune} {
			want, wd, _, err := Ephemeris{}.Apparent(b, jd)
			if err != nil {
				t.Fatal(err)
			}
			got, gd, _, err := inst.Apparent(b, jd)
			if err != nil {
				t.Fatalf("body %d: %v", b, err)
			}
			cx := got[1]*want[2] - got[2]*want[1]
			cy := got[2]*want[0] - got[0]*want[2]
			cz := got[0]*want[1] - got[1]*want[0]
			c := got[0]*want[0] + got[1]*want[1] + got[2]*want[2]
			// acos 在 1 附近每 ulp 约 3 mas，用 atan2 避开这一噪声底。
			arcsec := math.Atan2(math.Sqrt(cx*cx+cy*cy+cz*cz), c) * 206264.806
			worst = math.Max(worst, arcsec)
			// 原生链的太阳质心速度只计四颗巨行星，与通用链差到数 mas。
			if arcsec > 0.005 || math.Abs(gd-wd)/wd > 1e-9 {
				t.Errorf("jd %.1f body %d: direction %.5f\" dist %.3f km", jd, b, arcsec, gd-wd)
			}
		}
	}
	t.Logf("worst %.6f\"", worst)
}
