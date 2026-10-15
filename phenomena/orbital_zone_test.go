package phenomena

import (
	"testing"
	"time"
)

func TestApsidesIndependentOfHostZone(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	saved := time.Local
	defer func() { time.Local = saved }()

	time.Local = time.UTC
	want := Apsides(Earth, start, end)
	newYork, err := time.LoadLocation("America/New_York")
	if err != nil {
		newYork = time.FixedZone("EST", -5*3600)
	}
	time.Local = newYork
	got := Apsides(Earth, start, end)
	if len(got) != len(want) || len(got) == 0 {
		t.Fatalf("len got %d, want %d", len(got), len(want))
	}
	for i := range got {
		if !got[i].Time.Equal(want[i].Time) || got[i].Time.Location() != time.UTC {
			t.Fatalf("event %d: got %v, want %v", i, got[i].Time, want[i].Time)
		}
	}
}
