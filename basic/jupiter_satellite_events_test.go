package basic

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const galileanEventToleranceSeconds = 480.0

type galileanEventBaselineRecord struct {
	Label                string  `json:"label"`
	Satellite            int     `json:"satellite"`
	Type                 string  `json:"type"`
	StartUTC             string  `json:"start_utc"`
	StartDurationMinutes float64 `json:"start_duration_minutes"`
	EndUTC               string  `json:"end_utc"`
	EndDurationMinutes   float64 `json:"end_duration_minutes"`
}

func TestJupiterGalileanPhenomenonEventsAgainstIMCCEBaseline(t *testing.T) {
	records := loadGalileanEventBaseline(t)
	maxStartDiff := 0.0
	maxEndDiff := 0.0
	for _, record := range records {
		startUTC := mustParseRFC3339Nano(t, record.StartUTC)
		endUTC := mustParseRFC3339Nano(t, record.EndUTC)
		queryBefore := startUTC.Add(-12 * time.Hour)
		queryAfter := endUTC.Add(12 * time.Hour)
		queryMid := startUTC.Add(endUTC.Sub(startUTC) / 2)
		phenomenonType := parseBasicGalileanPhenomenonType(t, record.Type)

		next := NextJupiterGalileanPhenomenonEvent(Date2JD(queryBefore.UTC()), record.Satellite, phenomenonType)
		last := LastJupiterGalileanPhenomenonEvent(Date2JD(queryAfter.UTC()), record.Satellite, phenomenonType)
		closest := ClosestJupiterGalileanPhenomenonEvent(Date2JD(queryMid.UTC()), record.Satellite, phenomenonType)

		assertGalileanEventMatchesBaseline(t, record.Label+" next", next, record, startUTC, endUTC, &maxStartDiff, &maxEndDiff)
		assertGalileanEventMatchesBaseline(t, record.Label+" last", last, record, startUTC, endUTC, &maxStartDiff, &maxEndDiff)
		assertGalileanEventMatchesBaseline(t, record.Label+" closest", closest, record, startUTC, endUTC, &maxStartDiff, &maxEndDiff)
	}
	t.Logf("galilean event baseline max diff: start=%.1fs end=%.1fs", maxStartDiff, maxEndDiff)
}

func assertGalileanEventMatchesBaseline(
	t *testing.T,
	name string,
	event JupiterGalileanPhenomenonEvent,
	record galileanEventBaselineRecord,
	startUTC, endUTC time.Time,
	maxStartDiff, maxEndDiff *float64,
) {
	t.Helper()
	if !event.Valid {
		t.Fatalf("%s invalid event", name)
	}
	if event.Satellite != record.Satellite {
		t.Fatalf("%s satellite mismatch: got %d want %d", name, event.Satellite, record.Satellite)
	}
	if string(event.Type) != record.Type {
		t.Fatalf("%s type mismatch: got %q want %q", name, event.Type, record.Type)
	}
	gotStart := JD2DateByZone(event.Start, time.UTC, false)
	gotEnd := JD2DateByZone(event.End, time.UTC, false)
	startDiff := math.Abs(gotStart.Sub(startUTC).Seconds())
	endDiff := math.Abs(gotEnd.Sub(endUTC).Seconds())
	if startDiff > *maxStartDiff {
		*maxStartDiff = startDiff
	}
	if endDiff > *maxEndDiff {
		*maxEndDiff = endDiff
	}
	if startDiff > galileanEventToleranceSeconds {
		t.Fatalf("%s start mismatch: got %s want %s", name, gotStart.Format(time.RFC3339Nano), startUTC.Format(time.RFC3339Nano))
	}
	if endDiff > galileanEventToleranceSeconds {
		t.Fatalf("%s end mismatch: got %s want %s", name, gotEnd.Format(time.RFC3339Nano), endUTC.Format(time.RFC3339Nano))
	}
	if !(event.Start <= event.Greatest && event.Greatest <= event.End) {
		t.Fatalf("%s greatest not inside event: start=%.9f greatest=%.9f end=%.9f", name, event.Start, event.Greatest, event.End)
	}
	if !galileanPhenomenonFlag(event.GreatestPhenomenon, event.Type) {
		t.Fatalf("%s greatest state is not active", name)
	}
	if jupiterGalileanPhenomenonMetricAt(event.Start-5.0/86400.0, event.Satellite, event.Type).active {
		t.Fatalf("%s still active 5s before start", name)
	}
	if jupiterGalileanPhenomenonMetricAt(event.End+5.0/86400.0, event.Satellite, event.Type).active {
		t.Fatalf("%s still active 5s after end", name)
	}
}

func galileanPhenomenonFlag(phenomenon JupiterGalileanPhenomenon, phenomenonType JupiterGalileanPhenomenonType) bool {
	switch phenomenonType {
	case JupiterGalileanTransit:
		return phenomenon.Transit
	case JupiterGalileanOccultation:
		return phenomenon.Occultation
	case JupiterGalileanEclipse:
		return phenomenon.Eclipse
	case JupiterGalileanShadowTransit:
		return phenomenon.ShadowTransit
	default:
		return false
	}
}

func parseBasicGalileanPhenomenonType(t *testing.T, value string) JupiterGalileanPhenomenonType {
	t.Helper()
	switch JupiterGalileanPhenomenonType(value) {
	case JupiterGalileanTransit, JupiterGalileanOccultation, JupiterGalileanEclipse, JupiterGalileanShadowTransit:
		return JupiterGalileanPhenomenonType(value)
	default:
		t.Fatalf("unknown galilean phenomenon type %q", value)
		return ""
	}
}

func loadGalileanEventBaseline(t *testing.T) []galileanEventBaselineRecord {
	t.Helper()
	path := filepath.Join("..", "jupiter", "testdata", "galilean_events_imcce_2026.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var records []galileanEventBaselineRecord
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatal(err)
	}
	if len(records) == 0 {
		t.Fatal("empty galilean event baseline")
	}
	return records
}

func mustParseRFC3339Nano(t *testing.T, value string) time.Time {
	t.Helper()
	date, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		t.Fatalf("parse %q: %v", value, err)
	}
	return date
}

func TestJupiterGalileanSearchUsesJUP365GrazingGeometry(t *testing.T) {
	const seed = 2463143.9646
	next := NextJupiterGalileanPhenomenonEvent(seed, 4, JupiterGalileanTransit)
	if !next.Valid {
		t.Fatal("NextJupiterGalileanPhenomenonEvent returned nothing")
	}
	if offset := next.Greatest - seed; offset < 0.1 {
		t.Fatalf("next transit is %.3f d away, but JUP365 has no transit in the legacy grazing window", offset)
	}

	closest := ClosestJupiterGalileanPhenomenonEvent(2463144.0, 4, JupiterGalileanTransit)
	if !closest.Valid {
		t.Fatal("ClosestJupiterGalileanPhenomenonEvent returned nothing")
	}
	if math.Abs(next.Greatest-closest.Greatest) > 1e-5 {
		t.Fatalf("next greatest %.9f, want the same event as closest %.9f", next.Greatest, closest.Greatest)
	}

	last := LastJupiterGalileanPhenomenonEvent(next.End+0.05, 4, JupiterGalileanTransit)
	if !last.Valid {
		t.Fatal("LastJupiterGalileanPhenomenonEvent returned nothing")
	}
	if math.Abs(last.Greatest-closest.Greatest) > 1e-5 {
		t.Fatalf("last greatest %.9f, want the same event as closest %.9f", last.Greatest, closest.Greatest)
	}
}

func TestJupiterGalileanSearchConsistencyShortGrazingTransit(t *testing.T) {
	const satellite = 4
	const firstCoarse = 2441548.7104382128
	const denseStep = 5.0 / 86400
	const timeTolerance = 1.0 / 86400
	coarseStep := jupiterGalileanCoarseStepDays(satellite)
	secondCoarse := firstCoarse + coarseStep
	for _, jd := range []float64{firstCoarse - coarseStep, firstCoarse, secondCoarse, secondCoarse + coarseStep} {
		sample := jupiterGalileanPhenomenonMetricAt(jd, satellite, JupiterGalileanTransit)
		if !isFinite(sample.metric) || sample.active {
			t.Fatalf("fixture coarse point %.12f: active=%v metric=%.12g", jd, sample.active, sample.metric)
		}
	}
	firstActive, lastActive, denseGreatest := math.NaN(), math.NaN(), math.NaN()
	minimumMetric := math.Inf(1)
	ended := false
	steps := int(math.Round(coarseStep / denseStep))
	for sampleIndex := 0; sampleIndex <= steps; sampleIndex++ {
		jd := firstCoarse + float64(sampleIndex)*denseStep
		sample := jupiterGalileanPhenomenonMetricAt(jd, satellite, JupiterGalileanTransit)
		if !isFinite(sample.metric) {
			t.Fatalf("invalid dense metric at %.12f", jd)
		}
		if sample.metric < minimumMetric {
			minimumMetric, denseGreatest = sample.metric, jd
		}
		if sample.active {
			if ended {
				t.Fatal("fixture has more than one active interval between coarse points")
			}
			if math.IsNaN(firstActive) {
				firstActive = jd
			}
			lastActive = jd
		} else if !math.IsNaN(firstActive) {
			ended = true
		}
	}
	if math.IsNaN(firstActive) || !ended || firstActive-denseStep <= firstCoarse || lastActive+denseStep >= secondCoarse {
		t.Fatalf("fixture does not contain a complete hidden transit: coarse=[%.12f, %.12f] active=[%.12f, %.12f]", firstCoarse, secondCoarse, firstActive, lastActive)
	}
	if maximumDuration := lastActive - firstActive + 2*denseStep; maximumDuration >= 5.0/1440 {
		t.Fatalf("fixture is no longer short: dense upper duration %.3f min", maximumDuration*1440)
	}
	t.Logf("hidden transit coarse=[%.12f, %.12f] start bracket=[%.12f, %.12f] end bracket=[%.12f, %.12f] dense minimum=%.12g at %.12f", firstCoarse, secondCoarse, firstActive-denseStep, firstActive, lastActive, lastActive+denseStep, minimumMetric, denseGreatest)
	var firstEvent JupiterGalileanPhenomenonEvent
	for _, result := range []struct {
		name  string
		event JupiterGalileanPhenomenonEvent
	}{
		{"Next", NextJupiterGalileanPhenomenonEvent(firstCoarse, satellite, JupiterGalileanTransit)},
		{"Last", LastJupiterGalileanPhenomenonEvent(secondCoarse, satellite, JupiterGalileanTransit)},
		{"Closest", ClosestJupiterGalileanPhenomenonEvent(firstCoarse, satellite, JupiterGalileanTransit)},
	} {
		event := result.event
		if !event.Valid || !isFinite(event.Start) || !isFinite(event.Greatest) || !isFinite(event.End) {
			t.Errorf("%s returned no short grazing transit", result.name)
			continue
		}
		t.Logf("%s start=%.12f greatest=%.12f end=%.12f duration=%.6f min", result.name, event.Start, event.Greatest, event.End, (event.End-event.Start)*1440)
		if event.Start < firstActive-denseStep-timeTolerance || event.Start > firstActive+timeTolerance || event.End < lastActive-timeTolerance || event.End > lastActive+denseStep+timeTolerance || math.Abs(event.Greatest-denseGreatest) > denseStep+timeTolerance {
			t.Errorf("%s skipped the hidden transit: start=%.12f greatest=%.12f end=%.12f", result.name, event.Start, event.Greatest, event.End)
		}
		if event.Start >= event.Greatest || event.Greatest >= event.End || event.Start <= firstCoarse || event.End >= secondCoarse {
			t.Errorf("%s event does not lie completely between coarse points: %+v", result.name, event)
		}
		if firstEvent.Valid && (math.Abs(event.Start-firstEvent.Start) > timeTolerance || math.Abs(event.Greatest-firstEvent.Greatest) > timeTolerance || math.Abs(event.End-firstEvent.End) > timeTolerance) {
			t.Errorf("%s disagrees with Next by more than one second", result.name)
		}
		if !firstEvent.Valid {
			firstEvent = event
		}
	}
}

func TestJupiterGalileanShadowTransitSearchConsistency(t *testing.T) {
	seed := time.Date(2022, 6, 13, 6, 11, 35, 510000000, time.UTC)
	first := NextJupiterGalileanPhenomenonEvent(Date2JD(seed), 4, JupiterGalileanShadowTransit)
	if !first.Valid {
		t.Fatal("Next missed the June 2022 Callisto shadow transit")
	}
	second := NextJupiterGalileanPhenomenonEvent(first.End+0.1/86400, 4, JupiterGalileanShadowTransit)
	if !second.Valid {
		t.Fatal("Next missed the following Callisto shadow transit")
	}

	firstDate := JD2Date(first.Greatest).UTC().Format("2006-01-02")
	secondDate := JD2Date(second.Greatest).UTC().Format("2006-01-02")
	if firstDate != "2022-06-29" || secondDate != "2022-07-16" {
		t.Fatalf("shadow-transit dates=%s, %s; want 2022-06-29 and 2022-07-16", firstDate, secondDate)
	}
	for _, event := range []JupiterGalileanPhenomenonEvent{first, second} {
		if !jupiterGalileanPhenomenonMetricAt(event.Greatest, 4, JupiterGalileanShadowTransit).active {
			t.Fatalf("greatest is not active: %+v", event)
		}
		if jupiterGalileanPhenomenonMetricAt(event.Start-1.0/1440, 4, JupiterGalileanShadowTransit).active ||
			jupiterGalileanPhenomenonMetricAt(event.End+1.0/1440, 4, JupiterGalileanShadowTransit).active {
			t.Fatalf("event boundaries do not enclose one active interval: %+v", event)
		}
		for name, result := range map[string]JupiterGalileanPhenomenonEvent{
			"Last":    LastJupiterGalileanPhenomenonEvent(event.Greatest, 4, JupiterGalileanShadowTransit),
			"Closest": ClosestJupiterGalileanPhenomenonEvent(event.Greatest, 4, JupiterGalileanShadowTransit),
		} {
			if !result.Valid || math.Abs(result.Greatest-event.Greatest) > 1.0/86400 {
				t.Errorf("%s disagrees with Next event on %s: %+v", name, JD2Date(event.Greatest).UTC().Format("2006-01-02"), result)
			}
		}
	}
}
