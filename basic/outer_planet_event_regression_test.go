package basic

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

type outerPlanetEventBaselineSample struct {
	InputUTC string            `json:"input_utc"`
	TTJDBits uint64            `json:"tt_jd_bits"`
	Events   map[string]uint64 `json:"events"`
}

type outerPlanetEventBaseline struct {
	Samples []outerPlanetEventBaselineSample `json:"samples"`
}

func loadOuterPlanetEventBaseline(t *testing.T, path string) outerPlanetEventBaseline {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var baseline outerPlanetEventBaseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatal(err)
	}
	if len(baseline.Samples) == 0 {
		t.Fatalf("empty baseline: %s", path)
	}
	return baseline
}

func TestOuterPlanetEventBaselineRegression(t *testing.T) {
	for _, plan := range outerPlanetEventPlans() {
		t.Run(plan.planet, func(t *testing.T) {
			baseline := loadOuterPlanetEventBaseline(t, plan.baselineFile)
			cases := plan.allCases()

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
		})
	}
}
