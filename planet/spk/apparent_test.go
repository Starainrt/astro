package spk

import (
	"math"
	"os"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/planet/internal/semianalytic"
)

// SPK 只给几何状态，实例应走通用矢量视位置链；与按 DE441 拟合的内置表相比只剩表的拟合误差。
func TestSPKInstanceApparentMatchesSemianalytic(t *testing.T) {
	path := de441Path()
	if _, err := os.Stat(path); err != nil {
		t.Skip("DE441 file not available; set ASTRO_DE441_BSP")
	}
	file, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	inst := ephem.NewInstance(file)
	worst := 0.0
	for _, jd := range []float64{2451545.0, 2460409.25, 2470000.5} {
		for _, b := range []ephem.Body{ephem.Sun, ephem.Moon, ephem.Mercury, ephem.Venus, ephem.Mars, ephem.Jupiter, ephem.Saturn, ephem.Uranus, ephem.Neptune} {
			got, gd, _, err := inst.Apparent(b, jd)
			if err != nil {
				t.Fatalf("jd %.1f body %d: %v", jd, b, err)
			}
			want, wd, _, err := semianalytic.Ephemeris{}.Apparent(b, jd)
			if err != nil {
				t.Fatal(err)
			}
			c := got[0]*want[0] + got[1]*want[1] + got[2]*want[2]
			arcsec := math.Acos(math.Min(1, c)) * 206264.806
			worst = math.Max(worst, arcsec)
			// 内置表的几何距离取自快速档，外行星可差数万千米。
			if arcsec > 1.5 || math.Abs(gd-wd)/wd > 5e-5 {
				t.Errorf("jd %.1f body %d: %.4f\" dist rel %.2e", jd, b, arcsec, (gd-wd)/wd)
			}
		}
	}
	t.Logf("worst %.4f\"", worst)
}
