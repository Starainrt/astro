package satellite

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"testing"
	"time"
)

const (
	propagatePositionToleranceKM  = 1e-6
	propagateVelocityToleranceKMS = 1e-9
)

type propagationFixture struct {
	Label, Source string
	Elements      Elements
	Rows          []json.RawMessage
}

func readPropagationFixtures(t *testing.T) []propagationFixture {
	t.Helper()
	b, err := os.ReadFile("testdata/sgp4_python.json")
	if err != nil {
		t.Fatal(err)
	}
	var f []propagationFixture
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestPropagateExternal(t *testing.T) {
	var maxR, maxV float64
	rows, valid, failed := 0, 0, 0
	for _, g := range readPropagationFixtures(t) {
		if g.Label == "" || g.Source == "" {
			t.Fatal("missing provenance")
		}
		for _, raw := range g.Rows {
			var row []json.RawMessage
			if err := json.Unmarshal(raw, &row); err != nil || len(row) != 5 {
				t.Fatalf("invalid row: %s", raw)
			}
			var when time.Time
			var minutes float64
			var code int
			var wantR, wantV [3]float64
			for i, v := range []interface{}{&when, &minutes, &code, &wantR, &wantV} {
				if err := json.Unmarshal(row[i], v); err != nil {
					t.Fatal(err)
				}
			}
			if got := float64(when.Sub(g.Elements.Epoch).Nanoseconds()) / 6e10; got != minutes {
				t.Fatalf("%s inconsistent tsince %.17g want %.17g", g.Label, got, minutes)
			}
			got, err := Propagate(g.Elements, when)
			rows++
			if code != 0 {
				failed++
				want := map[int]error{1: ErrMeanEccentricity, 2: ErrMeanMotion, 3: ErrPerturbedEccentricity, 4: ErrSemiLatusRectum, 6: ErrDecayed}[code]
				if want == nil || !errors.Is(err, want) || got != (State{}) {
					t.Fatalf("%s t=%s got=%+v err=%v want=%v", g.Label, when, got, err, want)
				}
				continue
			}
			valid++
			if err != nil {
				t.Fatalf("%s %s: %v", g.Label, when, err)
			}
			if got.Time != when {
				t.Fatalf("%s output time changed", g.Label)
			}
			var dr, dv float64
			for i := range got.PositionKM {
				dr = math.Hypot(dr, got.PositionKM[i]-wantR[i])
				dv = math.Hypot(dv, got.VelocityKMS[i]-wantV[i])
			}
			maxR = math.Max(maxR, dr)
			maxV = math.Max(maxV, dv)
			if math.IsNaN(dr) || math.IsNaN(dv) || dr > propagatePositionToleranceKM || dv > propagateVelocityToleranceKMS {
				t.Fatalf("%s time=%s position=%.17g velocity=%.17g", g.Label, when, dr, dv)
			}
		}
	}
	if rows != 387 || valid != 380 || failed != 7 {
		t.Fatalf("rows=%d valid=%d errors=%d", rows, valid, failed)
	}
	t.Logf("rows=%d valid=%d errors=%d max_position=%.12g km max_velocity=%.12g km/s", rows, valid, failed, maxR, maxV)
}

func TestPropagationEpochPartsSynthetic(t *testing.T) {
	for _, tc := range []struct {
		when          time.Time
		day, fraction float64
	}{
		{time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), 2440587.5, 0},
		{time.Date(1969, 12, 31, 12, 0, 0, 0, time.UTC), 2440586.5, 0.5},
		{time.Date(2000, 1, 1, 12, 0, 0, 123456789, time.UTC), 2451544.5, float64(43200123456789) / 86400e9},
		{time.Date(2000, 1, 2, 7, 59, 59, 999999999, time.FixedZone("east", 8*3600)), 2451544.5, float64(86399999999999) / 86400e9},
	} {
		day, fraction := propagationEpochParts(tc.when)
		if day != tc.day || fraction != tc.fraction {
			t.Fatalf("%s: day=%.17g fraction=%.17g", tc.when, day, fraction)
		}
	}
}

func TestPropagateTimezoneAndOrderConsistency(t *testing.T) {
	fixtures := readPropagationFixtures(t)
	zones := []*time.Location{time.UTC, time.FixedZone("east", 14*3600), time.FixedZone("west", -12*3600)}
	for _, f := range fixtures[:8] {
		original := f.Elements
		for _, offset := range []time.Duration{24*time.Hour + 123456789, -24*time.Hour - 987654321, 0, 123 * time.Nanosecond} {
			when := original.Epoch.Add(offset)
			want, err := Propagate(original, when)
			if err != nil {
				t.Fatal(err)
			}
			for _, zone := range zones {
				e := original
				e.Epoch = e.Epoch.In(zone)
				date := when.In(zone)
				got, err := Propagate(e, date)
				if err != nil {
					t.Fatal(err)
				}
				if got.Time != date || got.Time.Location() != zone || got.PositionKM != want.PositionKM || got.VelocityKMS != want.VelocityKMS {
					t.Fatalf("%s time/zone/order changed: %+v", f.Label, got)
				}
				if e.Epoch.Location() != zone {
					t.Fatal("mutated elements")
				}
			}
		}
		if f.Elements != original {
			t.Fatal("mutated fixture")
		}
	}
}

func TestPropagateErrorsSynthetic(t *testing.T) {
	for _, tc := range []struct {
		code int
		want error
	}{
		{0, nil}, {1, ErrMeanEccentricity}, {2, ErrMeanMotion}, {3, ErrPerturbedEccentricity}, {4, ErrSemiLatusRectum}, {6, ErrDecayed}, {5, ErrInvalidElements}, {99, ErrInvalidElements},
	} {
		if got := propagationError(tc.code); !errors.Is(got, tc.want) {
			t.Fatalf("code=%d got=%v want=%v", tc.code, got, tc.want)
		}
	}
}

func TestPropagateInvalidSynthetic(t *testing.T) {
	base := readPropagationFixtures(t)[0].Elements
	when := base.Epoch.Add(time.Hour)
	for _, tc := range []struct {
		label  string
		modify func(*Elements)
	}{
		{"epoch", func(e *Elements) { e.Epoch = time.Time{} }},
		{"catalog", func(e *Elements) { e.CatalogNumber = 0 }},
		{"inclination", func(e *Elements) { e.Inclination = 181 }},
		{"right-ascension", func(e *Elements) { e.RightAscension = 360 }},
		{"eccentricity", func(e *Elements) { e.Eccentricity = 1 }},
		{"argument", func(e *Elements) { e.ArgumentOfPerigee = -1 }},
		{"anomaly", func(e *Elements) { e.MeanAnomaly = 360 }},
		{"mean-motion", func(e *Elements) { e.MeanMotion = 0 }},
	} {
		e := base
		tc.modify(&e)
		got, err := Propagate(e, when)
		if !errors.Is(err, ErrInvalidElements) || got != (State{}) {
			t.Fatalf("%s got=%+v err=%v", tc.label, got, err)
		}
	}
	for _, set := range []func(*Elements, float64){
		func(e *Elements, x float64) { e.MeanMotionDot = x }, func(e *Elements, x float64) { e.MeanMotionDDot = x },
		func(e *Elements, x float64) { e.BStar = x }, func(e *Elements, x float64) { e.Inclination = x },
		func(e *Elements, x float64) { e.RightAscension = x }, func(e *Elements, x float64) { e.Eccentricity = x },
		func(e *Elements, x float64) { e.ArgumentOfPerigee = x }, func(e *Elements, x float64) { e.MeanAnomaly = x }, func(e *Elements, x float64) { e.MeanMotion = x },
	} {
		for _, x := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
			e := base
			set(&e, x)
			got, err := Propagate(e, when)
			if !errors.Is(err, ErrInvalidElements) || got != (State{}) {
				t.Fatalf("nonfinite input got=%+v err=%v", got, err)
			}
		}
	}
	if got, err := Propagate(base, time.Time{}); !errors.Is(err, ErrInvalidElements) || got != (State{}) {
		t.Fatalf("zero time got=%+v err=%v", got, err)
	}
	for _, motion := range []float64{math.SmallestNonzeroFloat64, math.MaxFloat64} {
		e := base
		e.MeanMotion = motion
		got, err := Propagate(e, when)
		if err == nil || got != (State{}) {
			t.Fatalf("nonfinite intermediate motion=%g got=%+v err=%v", motion, got, err)
		}
	}
}

func TestPropagateDurationOverflowSynthetic(t *testing.T) {
	var e Elements
	for _, f := range readElementFixtures(t) {
		if f.Label == "saramago-json" {
			e = f.Expected[0]
		}
	}
	if e.Epoch.IsZero() {
		t.Fatal("missing elements")
	}
	for _, year := range []int{2400, 1600} {
		t.Run(time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006"), func(t *testing.T) {
			when := time.Date(year, 1, 1, 0, 0, 0, 123456789, time.UTC)
			got, err := Propagate(e, when)
			if !errors.Is(err, ErrInvalidElements) || got != (State{}) {
				t.Fatalf("saturated interval accepted: requested=%s duration=%v state=%+v err=%v", when, when.Sub(e.Epoch), got, err)
			}
		})
	}
}
