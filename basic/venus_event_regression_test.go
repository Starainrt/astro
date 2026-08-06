package basic

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"
)

type venusEventBaselineSample struct {
	InputUTC string            `json:"input_utc"`
	TTJDBits uint64            `json:"tt_jd_bits"`
	Events   map[string]uint64 `json:"events"`
}

type venusEventBaseline struct {
	Samples []venusEventBaselineSample `json:"samples"`
}

func loadVenusEventBaseline(t *testing.T) venusEventBaseline {
	t.Helper()

	data, err := os.ReadFile("testdata/venus_event_baseline.json")
	if err != nil {
		t.Fatal(err)
	}

	var baseline venusEventBaseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatal(err)
	}
	if len(baseline.Samples) == 0 {
		t.Fatal("empty venus event baseline")
	}
	return baseline
}

func TestVenusEventBaselineRegression(t *testing.T) {
	baseline := loadVenusEventBaseline(t)
	cases := venusEventCases()

	for _, sample := range baseline.Samples {
		jd := math.Float64frombits(sample.TTJDBits)
		for _, event := range cases {
			wantBits, ok := sample.Events[event.name]
			if !ok {
				t.Fatalf("%s missing baseline event %s", sample.InputUTC, event.name)
			}
			want := math.Float64frombits(wantBits)
			if math.IsNaN(want) || math.IsInf(want, 0) {
				t.Fatalf("%s %s baseline is non-finite %v", sample.InputUTC, event.name, want)
			}
			got := event.fn(jd)
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("%s %s returned non-finite result %v", sample.InputUTC, event.name, got)
			}
			diff := math.Abs(got - want)
			if diff > event.tolerance {
				t.Fatalf("%s %s diff %.12f > tolerance %.12f", sample.InputUTC, event.name, diff, event.tolerance)
			}
		}
	}
}

func TestVenusConjunctionExtremeInputIsBounded(t *testing.T) {
	started := time.Now()
	for _, next := range []uint8{0, 1} {
		got := venusConjunction(1e8, next)
		if math.IsInf(got, 0) {
			t.Fatalf("venusConjunction(1e8, %d) returned infinite result %v", next, got)
		}
	}
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Fatalf("venusConjunction extreme input took %s, want <= 2s", elapsed)
	}
}
