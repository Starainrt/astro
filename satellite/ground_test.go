package satellite

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"testing"
	"time"
)

type lookFixtures struct {
	Groups []struct {
		Label, Source string
		Elements      Elements
		Ground        []groundFixtureRow
	}
}

type groundFixtureRow struct {
	Time                               time.Time
	Position, Velocity                 [3]float64
	Longitude, Latitude, Height, Speed float64
}

func (r *groundFixtureRow) UnmarshalJSON(b []byte) error {
	var v []json.RawMessage
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	if len(v) != 7 {
		return errors.New("invalid ground row")
	}
	for i, dest := range []interface{}{&r.Time, &r.Position, &r.Velocity, &r.Longitude, &r.Latitude, &r.Height, &r.Speed} {
		if err := json.Unmarshal(v[i], dest); err != nil {
			return err
		}
	}
	return nil
}

func readLookFixtures(t *testing.T) lookFixtures {
	t.Helper()
	b, err := os.ReadFile("testdata/skyfield_look.json")
	if err != nil {
		t.Fatal(err)
	}
	var f lookFixtures
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestGroundPointExternal(t *testing.T) {
	var maximum [4]float64
	count := 0
	for _, g := range readLookFixtures(t).Groups {
		if g.Label == "" || g.Source == "" {
			t.Fatal("missing provenance")
		}
		for _, r := range g.Ground {
			got, err := GroundPointAt(g.Elements, r.Time)
			if err != nil {
				t.Fatal(err)
			}
			differences := [4]float64{math.Abs(math.Remainder(got.Longitude-r.Longitude, 360)), math.Abs(got.Latitude - r.Latitude), math.Abs(got.HeightKM - r.Height), math.Abs(got.SpeedKMS - r.Speed)}
			for i, limit := range [4]float64{1e-7, 1e-7, 1e-4, 1e-6} {
				maximum[i] = math.Max(maximum[i], differences[i])
				if math.IsNaN(differences[i]) || differences[i] > limit {
					t.Fatalf("%s %s field=%d difference=%.12g", g.Label, r.Time, i, differences[i])
				}
			}
			if got.Time != r.Time || got.Longitude <= -180 || got.Longitude > 180 {
				t.Fatalf("invalid output %+v", got)
			}
			count++
		}
	}
	if count != 36 {
		t.Fatalf("count=%d", count)
	}
	t.Logf("points=%d max longitude=%.12g deg latitude=%.12g deg height=%.12g km speed=%.12g km/s", count, maximum[0], maximum[1], maximum[2], maximum[3])
}

func TestGroundTrackConsistency(t *testing.T) {
	e := readLookFixtures(t).Groups[0].Elements
	start := e.Epoch.In(time.FixedZone("test", 8*3600))
	end := start.Add(125*time.Second + time.Nanosecond)
	got, err := GroundTrack(e, start, end, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	times := []time.Time{start, start.Add(time.Minute), start.Add(2 * time.Minute), end}
	if len(got) != len(times) {
		t.Fatalf("points=%d", len(got))
	}
	for i, when := range times {
		want, err := GroundPointAt(e, when)
		if err != nil || want != got[i] {
			t.Fatalf("point=%d got=%+v want=%+v err=%v", i, got[i], want, err)
		}
	}
	for _, span := range []time.Duration{0, 2 * time.Minute} {
		got, err := GroundTrack(e, start, start.Add(span), time.Minute)
		if err != nil || len(got) != int(span/time.Minute)+1 || got[len(got)-1].Time != start.Add(span) {
			t.Fatalf("closed interval %v points=%v err=%v", span, got, err)
		}
	}
}

func TestGroundTrackLimitsSynthetic(t *testing.T) {
	e := readLookFixtures(t).Groups[0].Elements
	start := e.Epoch
	for _, tc := range []struct {
		start, end time.Time
		step       time.Duration
	}{
		{time.Time{}, start, time.Second}, {start, time.Time{}, time.Second},
		{start, start.Add(-time.Second), time.Second}, {start, start.Add(time.Second), 0},
		{start, start.Add(time.Second), -time.Second},
		{start, start.Add(100000 * time.Millisecond), time.Millisecond},
		{start, start.Add(time.Duration(1<<63 - 1)), time.Nanosecond},
		{start, time.Date(2400, 1, 1, 0, 0, 0, 0, time.UTC), time.Hour},
	} {
		got, err := GroundTrack(e, tc.start, tc.end, tc.step)
		if !errors.Is(err, ErrInvalidInput) || got != nil {
			t.Fatalf("invalid interval accepted %+v: len=%d err=%v", tc, len(got), err)
		}
	}
	got, err := GroundTrack(e, start, start.Add(99999*time.Millisecond), time.Millisecond)
	if err != nil || len(got) != 100000 {
		t.Fatalf("point limit len=%d err=%v", len(got), err)
	}
	if got, err := GroundPointAt(Elements{}, start); !errors.Is(err, ErrInvalidElements) || got != (GroundPoint{}) {
		t.Fatalf("invalid elements %+v %v", got, err)
	}
	if got, err := GroundPointAt(e, time.Date(2400, 1, 1, 0, 0, 0, 0, time.UTC)); !errors.Is(err, ErrInvalidInput) || got != (GroundPoint{}) {
		t.Fatalf("overflow accepted %+v %v", got, err)
	}
}

func TestObservationTimeInvalidSynthetic(t *testing.T) {
	e := readLookFixtures(t).Groups[0].Elements
	for _, when := range []time.Time{{}, time.Date(2400, 1, 1, 0, 0, 0, 0, time.UTC)} {
		ground, err := GroundPointAt(e, when)
		if !errors.Is(err, ErrInvalidInput) || ground != (GroundPoint{}) {
			t.Errorf("GroundPointAt time=%s got=%+v err=%v", when, ground, err)
		}
		look, err := LookAt(e, Observer{}, when, ObserveOptions{})
		if !errors.Is(err, ErrInvalidInput) || look != (Look{}) {
			t.Errorf("LookAt time=%s got=%+v err=%v", when, look, err)
		}
		look, err = NewWithInstance(nil).LookAt(e, Observer{}, when, ObserveOptions{})
		if !errors.Is(err, ErrInvalidInput) || look != (Look{}) {
			t.Errorf("method LookAt time=%s got=%+v err=%v", when, look, err)
		}
	}
}
