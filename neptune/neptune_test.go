package neptune

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/ephemtest"
)

func sameUnixSecond(got time.Time, want int64) bool {
	return math.Abs(float64(got.Unix()-want)) <= 1
}

func TestNeptune(t *testing.T) {
	tz := time.FixedZone("CST", 8*3600)
	date := time.Date(2022, 01, 20, 00, 00, 00, 00, tz)
	values := ephemtest.NewValues(t, "neptune")
	want := func(key string, legacy int64, got time.Time) int64 {
		return int64(values.Want(key, float64(legacy), float64(got.Unix())))
	}
	conjunction := NextConjunction(date)
	if !sameUnixSecond(conjunction, want("NextConjunction", 1647171800, conjunction)) {
		t.Fatal(conjunction.Unix())
	}
	culmination := CulminationTime(date, 115)
	if culmination.Unix() != want("CulminationTime", 1642665021, culmination) {
		t.Fatal(culmination.Unix())
	}
	date, err := (RiseTime(date, 115, 40, 0, false))
	if err != nil {
		t.Fatal(err)
	}
	if date.Unix() != want("RiseTime", 1642644399, date) {
		t.Fatal(date.Unix())
	}
}
