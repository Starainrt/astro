package astro

import (
	"testing"
	"time"
)

func TestEarthUTCDateSyntheticContract(t *testing.T) {
	plusFourteen := time.FixedZone("UTC+14", 14*60*60)
	minusFourteen := time.FixedZone("UTC-14", -14*60*60)

	tests := []struct {
		name string
		in   time.Time
		want time.Time
	}{
		{name: "zero", in: time.Time{}, want: time.Time{}},
		{
			name: "UTC+14 crosses to previous UTC date",
			in:   time.Date(2026, time.January, 2, 0, 30, 0, 123, plusFourteen),
			want: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "UTC-14 crosses to next UTC date",
			in:   time.Date(2026, time.January, 1, 23, 30, 0, 456, minusFourteen),
			want: time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "before Unix epoch",
			in:   time.Date(1969, time.December, 31, 23, 59, 59, 999999999, minusFourteen),
			want: time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "after 2100",
			in:   time.Date(2101, time.March, 1, 0, 15, 0, 789, plusFourteen),
			want: time.Date(2101, time.February, 28, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := EarthUTCDate(test.in)
			if got != test.want {
				t.Fatalf("EarthUTCDate(%s) = %s, want %s", test.in, got, test.want)
			}
			if !got.IsZero() && got.Location() != time.UTC {
				t.Errorf("EarthUTCDate location = %v, want time.UTC", got.Location())
			}
			if !got.IsZero() && (got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0) {
				t.Errorf("EarthUTCDate time = %02d:%02d:%02d.%09d, want UTC midnight", got.Hour(), got.Minute(), got.Second(), got.Nanosecond())
			}
		})
	}
}
