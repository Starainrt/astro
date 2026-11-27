package coord_test

import (
	"fmt"
	"math"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/coord"
)

func ExampleToBodyEquatorial() {
	date := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	body, err := coord.ToBodyEquatorial(basic.SkyMars, date, 101.28715533, -16.71611586, coord.BodyEquatorOriginEquinox)
	if err != nil {
		panic(err)
	}
	back, err := coord.FromBodyEquatorial(basic.SkyMars, date, body.RA, body.Dec, coord.BodyEquatorOriginEquinox)
	fmt.Println(err == nil, math.Abs(math.Remainder(back.RA-101.28715533, 360)) < 1e-10, math.Abs(back.Dec+16.71611586) < 1e-10)
	// Output: true true true
}
