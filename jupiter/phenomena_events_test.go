package jupiter

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

type galileanPublicEventBaselineRecord struct {
	Label                string  `json:"label"`
	Satellite            int     `json:"satellite"`
	Type                 string  `json:"type"`
	StartUTC             string  `json:"start_utc"`
	StartDurationMinutes float64 `json:"start_duration_minutes"`
	EndUTC               string  `json:"end_utc"`
	EndDurationMinutes   float64 `json:"end_duration_minutes"`
}

func TestGalileanPhenomenonEventWrappersMatchBasic(t *testing.T) {
	records := loadGalileanPublicEventBaseline(t)
	loc := time.FixedZone("UTC+8", 8*3600)
	for _, record := range records {
		startUTC := mustParseGalileanEventTime(t, record.StartUTC)
		endUTC := mustParseGalileanEventTime(t, record.EndUTC)
		queryBefore := startUTC.Add(-12 * time.Hour).In(loc)
		queryAfter := endUTC.Add(12 * time.Hour).In(loc)
		queryMid := startUTC.Add(endUTC.Sub(startUTC) / 2).In(loc)
		phenomenonType := GalileanPhenomenonType(record.Type)

		assertGalileanWrapperMatchesBasic(
			t,
			record.Label+" next",
			NextGalileanPhenomenonEvent(queryBefore, record.Satellite, phenomenonType),
			basic.NextJupiterGalileanPhenomenonEvent(basic.Date2JDE(queryBefore.UTC()), record.Satellite, basic.JupiterGalileanPhenomenonType(phenomenonType)),
			loc,
		)
		assertGalileanWrapperMatchesBasic(
			t,
			record.Label+" last",
			LastGalileanPhenomenonEvent(queryAfter, record.Satellite, phenomenonType),
			basic.LastJupiterGalileanPhenomenonEvent(basic.Date2JDE(queryAfter.UTC()), record.Satellite, basic.JupiterGalileanPhenomenonType(phenomenonType)),
			loc,
		)
		assertGalileanWrapperMatchesBasic(
			t,
			record.Label+" closest",
			ClosestGalileanPhenomenonEvent(queryMid, record.Satellite, phenomenonType),
			basic.ClosestJupiterGalileanPhenomenonEvent(basic.Date2JDE(queryMid.UTC()), record.Satellite, basic.JupiterGalileanPhenomenonType(phenomenonType)),
			loc,
		)
	}
}

func assertGalileanWrapperMatchesBasic(
	t *testing.T,
	name string,
	got GalileanPhenomenonEvent,
	want basic.JupiterGalileanPhenomenonEvent,
	loc *time.Location,
) {
	t.Helper()
	if got.Valid != want.Valid {
		t.Fatalf("%s valid mismatch: got %v want %v", name, got.Valid, want.Valid)
	}
	if !got.Valid {
		return
	}
	if got.Start.Location() != loc || got.Greatest.Location() != loc || got.End.Location() != loc {
		t.Fatalf("%s timezone mismatch", name)
	}
	wantStart := basic.JDE2DateByZone(want.Start, loc, false)
	wantGreatest := basic.JDE2DateByZone(want.Greatest, loc, false)
	wantEnd := basic.JDE2DateByZone(want.End, loc, false)
	if !got.Start.Equal(wantStart) || !got.Greatest.Equal(wantGreatest) || !got.End.Equal(wantEnd) {
		t.Fatalf(
			"%s time mismatch: got [%s %s %s] want [%s %s %s]",
			name,
			got.Start.Format(time.RFC3339Nano),
			got.Greatest.Format(time.RFC3339Nano),
			got.End.Format(time.RFC3339Nano),
			wantStart.Format(time.RFC3339Nano),
			wantGreatest.Format(time.RFC3339Nano),
			wantEnd.Format(time.RFC3339Nano),
		)
	}
	if got.Duration != got.End.Sub(got.Start) {
		t.Fatalf("%s duration mismatch: got %s want %s", name, got.Duration, got.End.Sub(got.Start))
	}
	if got.Satellite != want.Satellite || string(got.Type) != string(want.Type) {
		t.Fatalf("%s id/type mismatch", name)
	}
	assertSameGalileanPhenomenon(t, name+" greatest", got.GreatestState, want.GreatestPhenomenon)
}

func assertSameGalileanPhenomenon(t *testing.T, name string, got GalileanSatellitePhenomenon, want basic.JupiterGalileanPhenomenon) {
	t.Helper()
	if got.Transit != want.Transit || got.Occultation != want.Occultation || got.Eclipse != want.Eclipse || got.ShadowTransit != want.ShadowTransit {
		t.Fatalf("%s flag mismatch", name)
	}
	gotFloats := []float64{got.ShadowOffsetXArcsec, got.ShadowOffsetYArcsec, got.ShadowOffsetXJupiterR, got.ShadowOffsetYJupiterR}
	wantFloats := []float64{want.ShadowOffsetXArcsec, want.ShadowOffsetYArcsec, want.ShadowOffsetXJupiterRadii, want.ShadowOffsetYJupiterRadii}
	for i := range gotFloats {
		if math.Float64bits(gotFloats[i]) != math.Float64bits(wantFloats[i]) {
			t.Fatalf("%s shadow field %d mismatch: got %.18f want %.18f", name, i, gotFloats[i], wantFloats[i])
		}
	}
}

func loadGalileanPublicEventBaseline(t *testing.T) []galileanPublicEventBaselineRecord {
	t.Helper()
	data, err := os.ReadFile("testdata/galilean_events_imcce_2026.json")
	if err != nil {
		t.Fatal(err)
	}
	var records []galileanPublicEventBaselineRecord
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatal(err)
	}
	if len(records) == 0 {
		t.Fatal("empty galilean public event baseline")
	}
	return records
}

func mustParseGalileanEventTime(t *testing.T, value string) time.Time {
	t.Helper()
	date, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		t.Fatalf("parse %q: %v", value, err)
	}
	return date
}
