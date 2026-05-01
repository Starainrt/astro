package jupiter

import (
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestGalileanPhenomenonContactEventWrappersMatchBasic(t *testing.T) {
	records := loadGalileanPublicEventBaseline(t)
	loc := time.FixedZone("UTC+8", 8*3600)
	for _, record := range records {
		startUTC := mustParseGalileanEventTime(t, record.StartUTC)
		endUTC := mustParseGalileanEventTime(t, record.EndUTC)
		queryMid := startUTC.Add(endUTC.Sub(startUTC) / 2).In(loc)
		phenomenonType := GalileanPhenomenonType(record.Type)

		got := ClosestGalileanPhenomenonContactEvent(queryMid, record.Satellite, phenomenonType)
		want := basic.ClosestJupiterGalileanPhenomenonContactEvent(
			basic.Date2JDE(queryMid.UTC()),
			record.Satellite,
			basic.JupiterGalileanPhenomenonType(phenomenonType),
		)
		assertGalileanContactWrapperMatchesBasic(t, record.Label, got, want, loc)
	}
}

func assertGalileanContactWrapperMatchesBasic(
	t *testing.T,
	name string,
	got GalileanPhenomenonContactEvent,
	want basic.JupiterGalileanPhenomenonContactEvent,
	loc *time.Location,
) {
	t.Helper()
	if got.Valid != want.Valid {
		t.Fatalf("%s valid mismatch: got %v want %v", name, got.Valid, want.Valid)
	}
	if !got.Valid {
		return
	}
	if got.Greatest.Location() != loc {
		t.Fatalf("%s greatest timezone mismatch", name)
	}
	wantGreatest := basic.JDE2DateByZone(want.Greatest, loc, false)
	if !got.Greatest.Equal(wantGreatest) {
		t.Fatalf("%s greatest mismatch: got %s want %s", name, got.Greatest.Format(time.RFC3339Nano), wantGreatest.Format(time.RFC3339Nano))
	}
	if got.Satellite != want.Satellite || string(got.Type) != string(want.Type) {
		t.Fatalf("%s id/type mismatch", name)
	}
	assertGalileanContactMatchesBasic(t, name+" disappearance", got.Disappearance, want.Disappearance, loc)
	assertGalileanContactMatchesBasic(t, name+" reappearance", got.Reappearance, want.Reappearance, loc)
	assertSameGalileanPhenomenon(t, name+" greatest", got.GreatestState, want.GreatestPhenomenon)
}

func assertGalileanContactMatchesBasic(
	t *testing.T,
	name string,
	got GalileanPhenomenonContact,
	want basic.JupiterGalileanPhenomenonContact,
	loc *time.Location,
) {
	t.Helper()
	if got.Valid != want.Valid {
		t.Fatalf("%s valid mismatch", name)
	}
	if !got.Valid {
		return
	}
	wantStart := basic.JDE2DateByZone(want.Start, loc, false)
	wantModel := basic.JDE2DateByZone(want.ModelCrossing, loc, false)
	wantEnd := basic.JDE2DateByZone(want.End, loc, false)
	if got.Start.Location() != loc || got.ModelCrossing.Location() != loc || got.End.Location() != loc {
		t.Fatalf("%s timezone mismatch", name)
	}
	if !got.Start.Equal(wantStart) || !got.ModelCrossing.Equal(wantModel) || !got.End.Equal(wantEnd) {
		t.Fatalf(
			"%s time mismatch: got [%s %s %s] want [%s %s %s]",
			name,
			got.Start.Format(time.RFC3339Nano),
			got.ModelCrossing.Format(time.RFC3339Nano),
			got.End.Format(time.RFC3339Nano),
			wantStart.Format(time.RFC3339Nano),
			wantModel.Format(time.RFC3339Nano),
			wantEnd.Format(time.RFC3339Nano),
		)
	}
	if got.Duration != got.End.Sub(got.Start) {
		t.Fatalf("%s duration mismatch", name)
	}
	if string(got.Phase) != string(want.Phase) {
		t.Fatalf("%s phase mismatch: got %q want %q", name, got.Phase, want.Phase)
	}
}
