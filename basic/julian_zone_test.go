package basic

import (
	"testing"
	"time"
)

func TestLocalJulianDayAcrossDST(t *testing.T) {
	for _, name := range []string{"America/New_York", "Australia/Lord_Howe"} {
		loc, err := time.LoadLocation(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, md := range [][2]int{{3, 8}, {11, 1}, {4, 5}, {10, 4}} {
			for _, hour := range []int{0, 3, 12, 23} {
				want := time.Date(2026, time.Month(md[0]), md[1], hour, 17, 23, 125000000, loc)
				got := JD2DateByZone(Date2JD(want), loc, true)
				if delta := got.Sub(want); delta < -100*time.Microsecond || delta > 100*time.Microsecond {
					t.Errorf("local %s: got %s, delta=%s", want, got, delta)
				}
				got = JD2DateByZone(Date2JD(want.UTC()), loc, false)
				if delta := got.Sub(want); delta < -100*time.Microsecond || delta > 100*time.Microsecond {
					t.Errorf("UTC %s: got %s, delta=%s", want, got, delta)
				}
			}
		}
	}
}
