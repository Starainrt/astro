package basic

import (
	"testing"
)

func TestConstellationNameZH(t *testing.T) {
	for _, jde := range []float64{2451545.0, 2461328.5} {
		for i := 0.00; i <= 360.00; i += 0.25 {
			for j := -90.00; j <= 90.00; j += 0.25 {
				if ConstellationNameZH(i, j, jde) == "" {
					t.Fatalf("ConstellationNameZH(%v, %v, %v) is empty", i, j, jde)
				}
			}
		}
	}
}

func TestConstellationBoundarySliver(t *testing.T) {
	for _, dec := range []float64{3.25, 3.5, 3.75, 4.0} {
		if got := ConstellationNameZH(278.25, dec, 2461328.5); got != "巨蛇座" && got != "蛇夫座" {
			t.Fatalf("ConstellationNameZH(278.25, %v) = %q, want 巨蛇座 or 蛇夫座", dec, got)
		}
	}
}

func BenchmarkConstellationNameZH(b *testing.B) {
	jde := GetNowJD()
	for i := 0; i < b.N; i++ {
		//GetNowJD()
		ConstellationNameZH(11.11, 12.12, jde)
	}

}
