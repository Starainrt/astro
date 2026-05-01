package basic

import (
	"math"
	"testing"
	"time"
)

func TestJupiterGalileanPhenomenonContactEventsAgainstIMCCEBaseline(t *testing.T) {
	records := loadGalileanEventBaseline(t)
	maxStartDiff := 0.0
	maxEndDiff := 0.0
	maxStartDurationDiff := 0.0
	maxEndDurationDiff := 0.0
	for _, record := range records {
		startUTC := mustParseRFC3339Nano(t, record.StartUTC)
		endUTC := mustParseRFC3339Nano(t, record.EndUTC)
		queryMid := startUTC.Add(endUTC.Sub(startUTC) / 2)
		phenomenonType := parseBasicGalileanPhenomenonType(t, record.Type)

		event := ClosestJupiterGalileanPhenomenonContactEvent(Date2JDE(queryMid.UTC()), record.Satellite, phenomenonType)
		if !event.Valid {
			t.Fatalf("%s invalid contact event", record.Label)
		}

		gotStart := JDE2DateByZone(event.Disappearance.Start, time.UTC, false)
		gotEnd := JDE2DateByZone(event.Reappearance.Start, time.UTC, false)
		startDiff := math.Abs(gotStart.Sub(startUTC).Seconds())
		endDiff := math.Abs(gotEnd.Sub(endUTC).Seconds())
		startDurationDiff := math.Abs((event.Disappearance.End-event.Disappearance.Start)*86400 - record.StartDurationMinutes*60)
		endDurationDiff := math.Abs((event.Reappearance.End-event.Reappearance.Start)*86400 - record.EndDurationMinutes*60)

		if startDiff > maxStartDiff {
			maxStartDiff = startDiff
		}
		if endDiff > maxEndDiff {
			maxEndDiff = endDiff
		}
		if startDurationDiff > maxStartDurationDiff {
			maxStartDurationDiff = startDurationDiff
		}
		if endDurationDiff > maxEndDurationDiff {
			maxEndDurationDiff = endDurationDiff
		}

		if startDiff > galileanContactTimeToleranceSeconds(phenomenonType) {
			t.Fatalf("%s disappearance start mismatch: got %s want %s", record.Label, gotStart.Format(time.RFC3339Nano), startUTC.Format(time.RFC3339Nano))
		}
		if endDiff > galileanContactTimeToleranceSeconds(phenomenonType) {
			t.Fatalf("%s reappearance start mismatch: got %s want %s", record.Label, gotEnd.Format(time.RFC3339Nano), endUTC.Format(time.RFC3339Nano))
		}
		if startDurationDiff > galileanContactDurationToleranceSeconds(phenomenonType) {
			t.Fatalf("%s disappearance duration mismatch: got %.1fs want %.1fs", record.Label, (event.Disappearance.End-event.Disappearance.Start)*86400, record.StartDurationMinutes*60)
		}
		if endDurationDiff > galileanContactDurationToleranceSeconds(phenomenonType) {
			t.Fatalf("%s reappearance duration mismatch: got %.1fs want %.1fs", record.Label, (event.Reappearance.End-event.Reappearance.Start)*86400, record.EndDurationMinutes*60)
		}
		if !(event.Disappearance.Start <= event.Disappearance.ModelCrossing && event.Disappearance.ModelCrossing <= event.Disappearance.End) {
			t.Fatalf("%s contact ordering invalid", record.Label)
		}
		if !(event.Reappearance.Start <= event.Reappearance.ModelCrossing && event.Reappearance.ModelCrossing <= event.Reappearance.End) {
			t.Fatalf("%s reappearance ordering invalid", record.Label)
		}
		fullEvent := ClosestJupiterGalileanPhenomenonEvent(Date2JDE(queryMid.UTC()), record.Satellite, phenomenonType)
		if phenomenonType != JupiterGalileanShadowTransit {
			if math.Abs(event.Disappearance.ModelCrossing-fullEvent.Start)*86400 > 2 {
				t.Fatalf("%s disappearance model crossing mismatch", record.Label)
			}
			if math.Abs(event.Reappearance.ModelCrossing-fullEvent.End)*86400 > 2 {
				t.Fatalf("%s reappearance model crossing mismatch", record.Label)
			}
		}
	}
	t.Logf(
		"galilean contact baseline max diff: start=%.1fs end=%.1fs startDur=%.1fs endDur=%.1fs",
		maxStartDiff,
		maxEndDiff,
		maxStartDurationDiff,
		maxEndDurationDiff,
	)
}

func galileanContactTimeToleranceSeconds(phenomenonType JupiterGalileanPhenomenonType) float64 {
	switch phenomenonType {
	case JupiterGalileanShadowTransit:
		return 90.0
	case JupiterGalileanEclipse:
		return 180.0
	default:
		return 120.0
	}
}

func galileanContactDurationToleranceSeconds(phenomenonType JupiterGalileanPhenomenonType) float64 {
	switch phenomenonType {
	case JupiterGalileanShadowTransit:
		return 15.0
	default:
		return 90.0
	}
}
