package basic

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

type marsEventBaselineSample struct {
	InputUTC string            `json:"input_utc"`
	TTJDBits uint64            `json:"tt_jd_bits"`
	Events   map[string]uint64 `json:"events"`
}

type marsEventBaseline struct {
	Samples []marsEventBaselineSample `json:"samples"`
}

func loadMarsEventBaseline(t *testing.T) marsEventBaseline {
	t.Helper()

	data, err := os.ReadFile("testdata/mars_event_baseline.json")
	if err != nil {
		t.Fatal(err)
	}

	var baseline marsEventBaseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatal(err)
	}
	if len(baseline.Samples) == 0 {
		t.Fatal("empty mars event baseline")
	}
	return baseline
}

func TestMarsEventBaselineRegression(t *testing.T) {
	baseline := loadMarsEventBaseline(t)
	cases := marsEventCases()

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
