package internal

import (
	"math"
	"testing"

	"github.com/starainrt/astro/basic"
)

func TestHeightDegreeByLatTakesDegrees(t *testing.T) {
	for _, lat := range []float64{0, 30, 60, -45, 89} {
		got := HeightDegreeByLat(10000, lat)
		want := basic.HeightDegreeByLat(10000, lat)
		if math.Abs(got-want) > 1e-9 {
			t.Fatalf("lat %v: lite %.9f, basic %.9f", lat, got, want)
		}
	}
}
