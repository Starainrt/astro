package basic

import (
	"math"
	"testing"
	"time"
)

func TestGetNowJDEUsesSingleTimestamp(t *testing.T) {
	oldTimeNow := timeNow
	defer func() {
		timeNow = oldTimeNow
	}()

	calls := 0
	first := time.Date(2026, 4, 29, 23, 59, 59, 0, time.FixedZone("CST", 8*3600))
	second := first.Add(2 * time.Second)
	timeNow = func() time.Time {
		calls++
		if calls == 1 {
			return first
		}
		return second
	}

	got := GetNowJDE()
	want := Date2JDE(first)
	if calls != 1 {
		t.Fatalf("GetNowJDE should read current time once, got %d calls", calls)
	}
	if math.Float64bits(got) != math.Float64bits(want) {
		t.Fatalf("GetNowJDE mismatch: got %.15f want %.15f", got, want)
	}
}
