package basic

import (
	"math"
	"testing"
	"time"
)

func TestOccultationStationCacheRecyclingPreservesContactModel(t *testing.T) {
	context := newOccultationRiseSetContext(2451545, 10, 5, 384000, 10.5, 5.25, 1e9, 60000)
	uncached := context
	uncached.states = nil
	for pass := 0; pass < 2; pass++ {
		for index := 0; index < 4*occultationRiseSetStateCacheMaximumEntries; index++ {
			lon, lat := float64(index*7-110), float64(index-40)
			for _, total := range []bool{false, true} {
				actual, reference := context, uncached
				if total {
					actual, reference = actual.withInternalContact(), reference.withInternalContact()
				}
				if got, want := actual.stateAt(lon, lat), reference.stateAt(lon, lat); got != want {
					t.Fatalf("cached station differs: got=%+v want=%+v", got, want)
				}
			}
			if len(context.states) > occultationRiseSetStateCacheMaximumEntries {
				t.Fatal("station cache exceeded bound")
			}
		}
	}
}

func TestOccultationEnvelopePredictionRequiresExactCorrection(t *testing.T) {
	start := time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC)
	events, err := FindBestPlanetOccultations(start, start.Add(48*time.Hour), OccultationSaturn, OccultationSearchOptions{})
	if err != nil || len(events) != 1 {
		t.Fatalf("events=%d err=%v", len(events), err)
	}
	tt := occultationTimeToTT(events[0].Greatest)
	config, _ := planetOccultationConfigFor(OccultationSaturn)
	event := newPlanetOccultationEventCache(config)
	event.prepareLocalEphemeris(tt)
	north, _, width, ok := occultationPathLimitsAndWidthForFrame(tt, event.outerFrameAt)
	if !ok {
		t.Fatal("missing limit seed")
	}
	seed := occultationPathPointFromVector(tt, north, width, time.UTC)
	sample, ok := occultationStationCorrectEnvelopePoint(tt, seed, event.riseSetContextAt, false, time.UTC)
	if !ok {
		t.Fatal("missing exact envelope seed")
	}
	model := occultationStationEnvelopeModel{kind: occultationStationContactEnvelope}
	exact := newOccultationRiseSetEvaluationCache(event.riseSetContextAt)
	state, ok := occultationStationEnvelopeArcStateAt(sample.point, tt, exact, model)
	if !ok {
		t.Fatal("missing envelope tangent")
	}
	predictor := state.coordinates
	for index := range predictor {
		predictor[index] += 0.2 * state.tangent[index]
	}
	want, _, ok := occultationStationCorrectEnvelopeArc(predictor, state.tangent, tt, exact, model, width, time.UTC)
	if !ok {
		t.Fatal("exact correction failed")
	}
	for _, invalid := range []bool{false, true} {
		candidateCalls := 0
		cache := newOccultationRiseSetEvaluationCacheWithCandidate(event.riseSetContextAt, func(at float64) occultationRiseSetContext {
			candidateCalls++
			if invalid {
				return occultationRiseSetContext{}
			}
			return event.candidateRiseSetContextAt(at)
		})
		got, _, ok := occultationStationCorrectEnvelopeArc(predictor, state.tangent, tt, cache, model, width, time.UTC)
		if !ok || candidateCalls == 0 {
			t.Fatalf("candidate correction ok=%v calls=%d invalid=%v", ok, candidateCalls, invalid)
		}
		if distance := occultationPathDistanceKM(want.point, got.point); distance > 0.005 {
			t.Fatalf("candidate displaced exact root by %v km", distance)
		}
		evaluation := exact.evaluation(centerTimeTT(got.point.Time))
		residual, ok := model.residual(evaluation, got.point.Longitude, got.point.Latitude)
		if !ok || math.Abs(residual[0]) > model.valueTolerance() || math.Abs(residual[1]) > model.derivativeTolerance() {
			t.Fatalf("accepted inexact root %v", residual)
		}
	}
}
