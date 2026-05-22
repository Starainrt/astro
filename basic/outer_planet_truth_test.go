package basic

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

type outerTruthBaselineFile struct {
	Events []outerTruthBaselineEvent `json:"events"`
}

type outerTruthBaselineEvent struct {
	Planet          string `json:"planet"`
	Kind            string `json:"kind"`
	HintKind        string `json:"hint_kind"`
	NAOJHintJST     string `json:"naoj_hint_jst"`
	Precision       string `json:"precision"`
	CandidateJST    string `json:"candidate_jst"`
	VerifiedJST     string `json:"verified_jst"`
	CandidateSource string `json:"candidate_source"`
}

func loadOuterTruthBaseline(t *testing.T) outerTruthBaselineFile {
	t.Helper()

	paths := [][]string{
		{
			"testdata/jpl_outer_event_baseline.json",
			"basic/testdata/jpl_outer_event_baseline.json",
		},
		{
			"testdata/jpl_outer_event_baseline_21c_sample.json",
			"basic/testdata/jpl_outer_event_baseline_21c_sample.json",
		},
	}

	var merged outerTruthBaselineFile
	for index, candidates := range paths {
		var (
			data []byte
			err  error
		)
		for _, path := range candidates {
			data, err = os.ReadFile(path)
			if err == nil {
				var baseline outerTruthBaselineFile
				if err := json.Unmarshal(data, &baseline); err != nil {
					t.Fatal(err)
				}
				merged.Events = append(merged.Events, baseline.Events...)
				break
			}
		}
		if err != nil && index == 0 {
			t.Fatal(err)
		}
	}
	if len(merged.Events) == 0 {
		t.Fatal("empty outer truth baseline file")
	}
	return merged
}

func outerTruthTolerance(event outerTruthBaselineEvent) time.Duration {
	switch event.Kind {
	case "CONJ", "OPP", "EQE", "EQW":
		return 2 * time.Minute
	default:
		return 2 * time.Minute
	}
}

func outerTruthEventFuncs(t *testing.T, event outerTruthBaselineEvent) (func(float64) float64, func(float64) float64) {
	t.Helper()
	switch event.Planet + ":" + event.Kind {
	case "Jupiter:CONJ":
		return LastJupiterConjunction, NextJupiterConjunction
	case "Jupiter:OPP":
		return LastJupiterOpposition, NextJupiterOpposition
	case "Jupiter:EQE":
		return LastJupiterEasternQuadrature, NextJupiterEasternQuadrature
	case "Jupiter:EQW":
		return LastJupiterWesternQuadrature, NextJupiterWesternQuadrature
	case "Jupiter:P2R":
		return LastJupiterProgradeToRetrograde, NextJupiterProgradeToRetrograde
	case "Jupiter:R2P":
		return LastJupiterRetrogradeToPrograde, NextJupiterRetrogradeToPrograde
	case "Saturn:CONJ":
		return LastSaturnConjunction, NextSaturnConjunction
	case "Saturn:OPP":
		return LastSaturnOpposition, NextSaturnOpposition
	case "Saturn:EQE":
		return LastSaturnEasternQuadrature, NextSaturnEasternQuadrature
	case "Saturn:EQW":
		return LastSaturnWesternQuadrature, NextSaturnWesternQuadrature
	case "Saturn:P2R":
		return LastSaturnProgradeToRetrograde, NextSaturnProgradeToRetrograde
	case "Saturn:R2P":
		return LastSaturnRetrogradeToPrograde, NextSaturnRetrogradeToPrograde
	case "Uranus:CONJ":
		return LastUranusConjunction, NextUranusConjunction
	case "Uranus:OPP":
		return LastUranusOpposition, NextUranusOpposition
	case "Uranus:EQE":
		return LastUranusEasternQuadrature, NextUranusEasternQuadrature
	case "Uranus:EQW":
		return LastUranusWesternQuadrature, NextUranusWesternQuadrature
	case "Uranus:P2R":
		return LastUranusProgradeToRetrograde, NextUranusProgradeToRetrograde
	case "Uranus:R2P":
		return LastUranusRetrogradeToPrograde, NextUranusRetrogradeToPrograde
	case "Neptune:CONJ":
		return LastNeptuneConjunction, NextNeptuneConjunction
	case "Neptune:OPP":
		return LastNeptuneOpposition, NextNeptuneOpposition
	case "Neptune:EQE":
		return LastNeptuneEasternQuadrature, NextNeptuneEasternQuadrature
	case "Neptune:EQW":
		return LastNeptuneWesternQuadrature, NextNeptuneWesternQuadrature
	case "Neptune:P2R":
		return LastNeptuneProgradeToRetrograde, NextNeptuneProgradeToRetrograde
	case "Neptune:R2P":
		return LastNeptuneRetrogradeToPrograde, NextNeptuneRetrogradeToPrograde
	default:
		t.Fatalf("unsupported outer event %s:%s", event.Planet, event.Kind)
		return nil, nil
	}
}

func assertOuterTruthBaselineEvent(t *testing.T, event outerTruthBaselineEvent, lastFn, nextFn func(float64) float64) {
	t.Helper()
	when := parseInnerBaselineTime(t, event.VerifiedJST)
	before := when.Add(-7 * 24 * time.Hour)
	after := when.Add(7 * 24 * time.Hour)
	next := JDE2DateByZone(nextFn(toUTJD(before)), when.Location(), false)
	last := JDE2DateByZone(lastFn(toUTJD(after)), when.Location(), false)
	tolerance := outerTruthTolerance(event)

	if diff := next.Sub(when); diff < -tolerance || diff > tolerance {
		t.Fatalf("%s %s next mismatch: got %s want %s tol=%s hint=%s candidate=%s via=%s", event.Planet, event.Kind, next, when, tolerance, event.NAOJHintJST, event.CandidateJST, event.CandidateSource)
	}
	if diff := last.Sub(when); diff < -tolerance || diff > tolerance {
		t.Fatalf("%s %s last mismatch: got %s want %s tol=%s hint=%s candidate=%s via=%s", event.Planet, event.Kind, last, when, tolerance, event.NAOJHintJST, event.CandidateJST, event.CandidateSource)
	}
}

func TestOuterPlanetPhaseTruthAgainstJPL(t *testing.T) {
	baseline := loadOuterTruthBaseline(t)
	for _, event := range baseline.Events {
		event := event
		switch event.Kind {
		case "P2R", "R2P":
			// Station rows are retained as JPL apparent-RA reference data for
			// future refinement. Current station behavior is constrained by the
			// library's existing station baseline instead of these reference rows.
			continue
		}
		name := strings.Join([]string{event.Planet, event.Kind, event.VerifiedJST}, "_")
		t.Run(name, func(t *testing.T) {
			lastFn, nextFn := outerTruthEventFuncs(t, event)
			assertOuterTruthBaselineEvent(t, event, lastFn, nextFn)
		})
	}
}

func TestOuterPlanetStationJPLReferenceLoaded(t *testing.T) {
	baseline := loadOuterTruthBaseline(t)
	count := 0
	for _, event := range baseline.Events {
		switch event.Kind {
		case "P2R", "R2P":
			count++
		}
	}
	if count == 0 {
		t.Fatal("missing outer station JPL reference rows")
	}
}
