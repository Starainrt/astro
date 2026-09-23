package civiltime

import (
	"errors"
	"testing"
	"time"
)

func TestEventFiltersByActualCivilDay(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	missing := errors.New("no event on requested day")
	cases := []struct {
		month, day, offset, hour int
		wantFound                bool
	}{
		{3, 8, -5, 23, false},
		{11, 1, -5, 23, true},
	}
	for _, tc := range cases {
		date := time.Date(2026, time.Month(tc.month), tc.day, 12, 0, 0, 0, loc)
		candidate := time.Date(2026, time.Month(tc.month), tc.day, tc.hour, 30, 0, 0, time.FixedZone("", tc.offset*3600))
		calls := 0
		got, err, handled := Event(date, missing, func(fixed time.Time) (time.Time, error) {
			calls++
			if fixed.Location() == loc {
				t.Fatal("calculation must use a fixed offset")
			}
			return candidate, nil
		})
		if !handled || calls != 2 {
			t.Fatalf("handled=%t calls=%d", handled, calls)
		}
		if tc.wantFound {
			if err != nil || !got.Equal(candidate) || got.Location() != loc {
				t.Errorf("got %s, %v; want %s", got, err, candidate)
			}
		} else if !got.IsZero() || !errors.Is(err, missing) {
			t.Errorf("outside-day result %s, %v", got, err)
		}
	}
}

func TestEventLeavesFixedOffsetToCaller(t *testing.T) {
	date := time.Date(2026, 11, 1, 12, 0, 0, 0, time.FixedZone("EST", -5*3600))
	_, _, handled := Event(date, nil, func(time.Time) (time.Time, error) {
		t.Fatal("ordinary day invoked fallback")
		return time.Time{}, nil
	})
	if handled {
		t.Fatal("ordinary day reported handled")
	}
}
