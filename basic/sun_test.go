package basic

import (
	"math"
	"testing"

	. "github.com/starainrt/astro/tools"
)

func TestNutationRegression(t *testing.T) {
	jde := 2452982.9872345612

	if got := Nutation2000Bi(jde); math.Abs(got-(-0.003747344689665249)) > 1e-9 {
		t.Fatalf("Nutation2000Bi mismatch: got %.18f", got)
	}
	if got := Nutation1980s(jde); math.Abs(got-0.001511763511795865) > 1e-9 {
		t.Fatalf("Nutation1980s mismatch: got %.18f", got)
	}
}

func Benchmark_SunRise(b *testing.B) {
	jde := GetNowJDE()
	for i := 0; i < b.N; i++ {
		_, _ = GetSunRiseTime(jde, 115, 32, 8, 0, 10)
	}
}

func Benchmark_SunLo(b *testing.B) {
	jde := GetNowJDE()
	for i := 0; i < b.N; i++ {
		HSunApparentLo(jde)
	}
}

// SunApparentDec 必须只用真黄赤交角（平交角 + 交角章动）一次。历史上这里又加了一项
// 0.00256*cos(Ω)，而那一项本身就是交角章动的近似值（《天文算法》24 章译者注：
// "ε 补上交角章动"），于是章动被计两次，视赤纬最大偏 ~12.3″。
func TestSunApparentDecCountsObliquityNutationOnce(t *testing.T) {
	for _, jd := range []float64{2460310.5, 2451545.0, 2448908.5, 2415020.5, 2500000.5} {
		want := ArcSin(Sin(TrueObliquity(jd)) * Sin(SunApparentLo(jd)))
		if got := SunApparentDec(jd); math.Float64bits(got) != math.Float64bits(want) {
			t.Fatalf("jd %.1f: SunApparentDec=%v, want %v (TrueObliquity 单次)", jd, got, want)
		}
	}
	// 与库内高精度视赤纬的全年偏差应回落到低精度视黄经决定的地板（~8″），而不是重复章动的 ~12″。
	maxDiff := 0.0
	for day := 0; day < 365; day++ {
		jd := 2460310.5 + float64(day)
		diff := math.Abs(SunApparentDec(jd)-HSunApparentDec(jd)) * 3600
		if diff > maxDiff {
			maxDiff = diff
		}
	}
	if maxDiff > 10 {
		t.Fatalf("全年 max|Δ| = %.2f″，仍高于低精度黄经地板（重复章动会到 ~12.3″）", maxDiff)
	}
}
