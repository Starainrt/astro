package star

import (
	"math"
	"testing"
	"time"
)

func TestStarRiseSetWithAtmosphere(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	date := time.Date(2025, 1, 1, 0, 0, 0, 0, loc)
	for _, rise := range []bool{true, false} {
		plainFn, atmosFn := SetTime, SetTimeWithAtmosphere
		if rise {
			plainFn, atmosFn = RiseTime, RiseTimeWithAtmosphere
		}
		plain, err := plainFn(date, 101.29, -16.72, 116.4, 39.9, 0, true)
		if err != nil {
			t.Fatal(err)
		}
		std, err := atmosFn(date, 101.29, -16.72, 116.4, 39.9, 0, 1010, 10)
		if err != nil {
			t.Fatal(err)
		}
		if d := math.Abs(std.Sub(plain).Seconds()); d > 5 {
			t.Errorf("rise=%v: standard weather differs by %.2f s", rise, d)
		}
	}
}
