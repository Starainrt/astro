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

		next := NextJupiterGalileanPhenomenonEvent(Date2JDE(queryBefore.UTC()), record.Satellite, phenomenonType)
		last := LastJupiterGalileanPhenomenonEvent(Date2JDE(queryAfter.UTC()), record.Satellite, phenomenonType)
		closest := ClosestJupiterGalileanPhenomenonEvent(Date2JDE(queryMid.UTC()), record.Satellite, phenomenonType)

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
	gotStart := JDE2DateByZone(event.Start, time.UTC, false)
	gotEnd := JDE2DateByZone(event.End, time.UTC, false)
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
