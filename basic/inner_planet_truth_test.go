package basic

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

type innerBaselineFile struct {
	Events []innerBaselineEvent `json:"events"`
}

type innerBaselineEvent struct {
	Planet          string `json:"planet"`
	Kind            string `json:"kind"`
	NAOJHintJST     string `json:"naoj_hint_jst"`
	Precision       string `json:"precision"`
	CandidateJST    string `json:"candidate_jst"`
	VerifiedJST     string `json:"verified_jst"`
	CandidateSource string `json:"candidate_source"`
}

func loadInnerBaseline(t *testing.T) innerBaselineFile {
	t.Helper()

	paths := [][]string{
		{
			"testdata/jpl_inner_event_baseline.json",
			"basic/testdata/jpl_inner_event_baseline.json",
		},
		{
			"testdata/jpl_inner_event_baseline_21c_sample.json",
			"basic/testdata/jpl_inner_event_baseline_21c_sample.json",
		},
		{
			"testdata/jpl_inner_event_baseline_20c_sample.json",
			"basic/testdata/jpl_inner_event_baseline_20c_sample.json",
		},
		{
			"testdata/jpl_inner_event_baseline_22c_sample.json",
			"basic/testdata/jpl_inner_event_baseline_22c_sample.json",
		},
	}
	var merged innerBaselineFile
	for index, candidates := range paths {
		var (
			data []byte
			err  error
		)
		for _, path := range candidates {
			data, err = os.ReadFile(path)
			if err == nil {
				var baseline innerBaselineFile
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
		t.Fatal("empty inner baseline file")
	}
	return merged
}

func parseInnerBaselineTime(t *testing.T, value string) time.Time {
	t.Helper()
	loc := time.FixedZone("JST", 9*3600)
	layouts := []string{
		"2006-01-02 15:04:05 MST",
		"2006-01-02 15:04 MST",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
	}
	var err error
	for _, layout := range layouts {
		when, parseErr := time.ParseInLocation(layout, value, loc)
		if parseErr == nil {
			return when
		}
		err = parseErr
	}
	t.Fatalf("parse baseline time %q: %v", value, err)
	return time.Time{}
}

func innerBaselineTolerance(event innerBaselineEvent) time.Duration {
	switch event.Kind {
	case "IC", "SC", "P2R", "R2P":
		return 2 * time.Minute
	case "GEE", "GEW":
		return 90 * time.Minute
	default:
		return 2 * time.Minute
	}
}

func innerEventFuncs(t *testing.T, event innerBaselineEvent) (func(float64) float64, func(float64) float64) {
	t.Helper()
	switch event.Planet + ":" + event.Kind {
	case "Mercury:IC":
		return LastMercuryInferiorConjunctionInclusive, NextMercuryInferiorConjunctionInclusive
	case "Mercury:SC":
		return LastMercurySuperiorConjunctionInclusive, NextMercurySuperiorConjunctionInclusive
	case "Mercury:P2R":
		return LastMercuryProgradeToRetrogradeInclusive, NextMercuryProgradeToRetrogradeInclusive
	case "Mercury:R2P":
		return LastMercuryRetrogradeToProgradeInclusive, NextMercuryRetrogradeToProgradeInclusive
	case "Mercury:GEE":
		return LastMercuryGreatestElongationEastInclusive, NextMercuryGreatestElongationEastInclusive
	case "Mercury:GEW":
		return LastMercuryGreatestElongationWestInclusive, NextMercuryGreatestElongationWestInclusive
	case "Venus:IC":
		return LastVenusInferiorConjunctionInclusive, NextVenusInferiorConjunctionInclusive
	case "Venus:SC":
		return LastVenusSuperiorConjunctionInclusive, NextVenusSuperiorConjunctionInclusive
	case "Venus:P2R":
		return LastVenusProgradeToRetrogradeInclusive, NextVenusProgradeToRetrogradeInclusive
	case "Venus:R2P":
		return LastVenusRetrogradeToProgradeInclusive, NextVenusRetrogradeToProgradeInclusive
	case "Venus:GEE":
		return LastVenusGreatestElongationEastInclusive, NextVenusGreatestElongationEastInclusive
	case "Venus:GEW":
		return LastVenusGreatestElongationWestInclusive, NextVenusGreatestElongationWestInclusive
	default:
		t.Fatalf("unsupported event %s:%s", event.Planet, event.Kind)
		return nil, nil
	}
}

func assertInnerBaselineEvent(t *testing.T, event innerBaselineEvent, lastFn, nextFn func(float64) float64) {
	t.Helper()
	when := parseInnerBaselineTime(t, event.VerifiedJST)
	before := when.Add(-24 * time.Hour)
	after := when.Add(24 * time.Hour)
	next := JDE2DateByZone(nextFn(toUTJD(before)), when.Location(), false)
	last := JDE2DateByZone(lastFn(toUTJD(after)), when.Location(), false)
	tolerance := innerBaselineTolerance(event)

	if diff := next.Sub(when); diff < -tolerance || diff > tolerance {
		t.Fatalf("%s %s next mismatch: got %s want %s tol=%s hint=%s candidate=%s via=%s", event.Planet, event.Kind, next, when, tolerance, event.NAOJHintJST, event.CandidateJST, event.CandidateSource)
	}
	if diff := last.Sub(when); diff < -tolerance || diff > tolerance {
		t.Fatalf("%s %s last mismatch: got %s want %s tol=%s hint=%s candidate=%s via=%s", event.Planet, event.Kind, last, when, tolerance, event.NAOJHintJST, event.CandidateJST, event.CandidateSource)
	}
}

func TestInnerPlanetTruthAgainstJPL(t *testing.T) {
	baseline := loadInnerBaseline(t)
	for _, event := range baseline.Events {
		event := event
		name := strings.Join([]string{event.Planet, event.Kind, event.VerifiedJST}, "_")
		t.Run(name, func(t *testing.T) {
			lastFn, nextFn := innerEventFuncs(t, event)
			assertInnerBaselineEvent(t, event, lastFn, nextFn)
		})
	}
}
