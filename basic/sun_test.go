package basic

import (
	"math"
	"testing"
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
