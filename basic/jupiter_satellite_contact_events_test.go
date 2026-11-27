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

		event := ClosestJupiterGalileanPhenomenonContactEvent(Date2JD(queryMid.UTC()), record.Satellite, phenomenonType)
		if !event.Valid {
			t.Fatalf("%s invalid contact event", record.Label)
		}

		gotStart := JD2DateByZone(event.Disappearance.Start, time.UTC, false)
		gotEnd := JD2DateByZone(event.Reappearance.Start, time.UTC, false)
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
		fullEvent := ClosestJupiterGalileanPhenomenonEvent(Date2JD(queryMid.UTC()), record.Satellite, phenomenonType)
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

// 木影食 D/F 接触窗口的时间先后：慢速的木卫四半影内切可早于零半径穿越，D 窗口终点不得跑到出影一侧。
func TestJupiterGalileanEclipseContactConsistencyOrder(t *testing.T) {
	satellites := []int{1, 2, 3, 4}
	if testing.Short() {
		satellites = []int{4}
	}
	for _, satellite := range satellites {
		jd := Date2JD(JD2Date(JDCalc(2026, 1, 1)))
		checked := 0
		for checked < 12 {
			event := NextJupiterGalileanPhenomenonEventWithInstance(jd, satellite, JupiterGalileanEclipse, nil)
			if !event.Valid {
				t.Fatalf("satellite %d eclipse missing", satellite)
			}
			jd = event.End + 1.0/86400
			contact := jupiterGalileanPhenomenonContactEventFromEvent(event, nil)
			if !contact.Valid {
				t.Fatalf("satellite %d contact invalid at %s", satellite, JD2Date(event.Greatest).Format("2006-01-02"))
			}
			d, f := contact.Disappearance, contact.Reappearance
			if !(d.Start < d.End && d.End < contact.Greatest && contact.Greatest < f.Start && f.Start < f.End) {
				t.Fatalf("satellite %d %s contacts out of order: D %.6f..%.6f G %.6f F %.6f..%.6f", satellite, JD2Date(event.Greatest).Format("2006-01-02"), d.Start, d.End, contact.Greatest, f.Start, f.End)
			}
			radius := jupiterGalileanSatelliteRadiusJupiterRadii(satellite)
			before, _ := jupiterGalileanEclipseSignedDistanceAt(d.End-2.0/86400, satellite, true)
			after, _ := jupiterGalileanEclipseSignedDistanceAt(d.End+2.0/86400, satellite, true)
			if !(before+radius > 0 && after+radius < 0) {
				t.Fatalf("satellite %d D end not within 2 s of penumbral internal contact: %.6f / %.6f", satellite, before+radius, after+radius)
			}
			if (d.End-d.Start)*1440 > 20 {
				t.Fatalf("satellite %d D window %.1f min", satellite, (d.End-d.Start)*1440)
			}
			checked++
		}
	}
}
