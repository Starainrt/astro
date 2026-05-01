package basic

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

type mercuryEventBaselineSample struct {
	InputUTC string            `json:"input_utc"`
	TTJDBits uint64            `json:"tt_jd_bits"`
	Events   map[string]uint64 `json:"events"`
}

type mercuryEventBaseline struct {
	Samples []mercuryEventBaselineSample `json:"samples"`
}

func loadMercuryEventBaseline(t *testing.T) mercuryEventBaseline {
	t.Helper()

	data, err := os.ReadFile("testdata/mercury_event_baseline.json")
	if err != nil {
		t.Fatal(err)
	}

	var baseline mercuryEventBaseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatal(err)
	}
	if len(baseline.Samples) == 0 {
		t.Fatal("empty mercury event baseline")
	}
	return baseline
}

func TestMercuryEventBaselineRegression(t *testing.T) {
	baseline := loadMercuryEventBaseline(t)
	cases := mercuryEventCases()

	for _, sample := range baseline.Samples {
		jd := math.Float64frombits(sample.TTJDBits)
		for _, event := range cases {
			wantBits, ok := sample.Events[event.name]
			if !ok {
				t.Fatalf("%s missing baseline event %s", sample.InputUTC, event.name)
			}
			want := math.Float64frombits(wantBits)
			got := event.fn(jd)
			diff := math.Abs(got - want)
			if diff > event.tolerance {
				t.Fatalf("%s %s diff %.12f > tolerance %.12f", sample.InputUTC, event.name, diff, event.tolerance)
			}
		}
	}
}
