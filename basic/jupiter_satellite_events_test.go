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

func TestJupiterGalileanSearchFindsShortGrazingTransit(t *testing.T) {
	// 2463144 附近 Callisto 有一次擦边凌日：Start 2463143.9993 / Greatest 2463144.0015 /
	// End 2463144.0038，全场只有 6.4 分钟，整段落在 2 小时粗采样之间。旧实现看不到度量
	// 凹陷，会把 16.85 天后的下一次凌日当成 "next"（Valid 仍为 true，属静默跳事件）。
	// Callisto has a grazing transit near 2463144: Start 2463143.9993, Greatest
	// 2463144.0015, End 2463144.0038 — only 6.4 minutes, entirely between two two-hour
	// coarse samples. The old search saw no metric dip and reported the next transit 16.85
	// days later as "next", still with Valid=true (a silent skip).
	closest := ClosestJupiterGalileanPhenomenonEvent(2463144.0, 4, JupiterGalileanTransit)
	if !closest.Valid {
		t.Fatal("no Callisto transit found at 2463144")
	}
	if duration := closest.End - closest.Start; duration > 0.01 {
		t.Fatalf("fixture changed: transit duration %.4f d, want the 6.4-minute grazing event", duration)
	}

	next := NextJupiterGalileanPhenomenonEvent(2463143.9646, 4, JupiterGalileanTransit)
	if !next.Valid {
		t.Fatal("NextJupiterGalileanPhenomenonEvent returned nothing")
	}
	if offset := next.Greatest - 2463143.9646; offset > 0.1 {
		t.Fatalf("next transit is %.3f d away, want the 53-minute one (the search skipped it)", offset)
	}
	if math.Abs(next.Greatest-closest.Greatest) > 1e-5 {
		t.Fatalf("next greatest %.9f, want the same event as closest %.9f", next.Greatest, closest.Greatest)
	}

	last := LastJupiterGalileanPhenomenonEvent(2463144.05, 4, JupiterGalileanTransit)
	if !last.Valid {
		t.Fatal("LastJupiterGalileanPhenomenonEvent returned nothing")
	}
	if math.Abs(last.Greatest-closest.Greatest) > 1e-5 {
		t.Fatalf("last greatest %.9f, want the same event as closest %.9f", last.Greatest, closest.Greatest)
	}
}
