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
			basic.Date2JD(queryMid.UTC()),
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
	wantGreatest := basic.JD2DateByZone(want.Greatest, loc, false)
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
	wantStart := basic.JD2DateByZone(want.Start, loc, false)
	wantModel := basic.JD2DateByZone(want.ModelCrossing, loc, false)
	wantEnd := basic.JD2DateByZone(want.End, loc, false)
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

// IMCCE 年表每个事件给 4 个时刻：D1、D2 = D1+时长、F1、F2 = F1+时长。
func TestGalileanPhenomenonContactEventMatchesIMCCEBaseline(t *testing.T) {
	records := loadGalileanPublicEventBaseline(t)
	const (
		contactTolerance  = 120 * time.Second
		durationTolerance = 25 * time.Second
	)
	var maxContact, maxDuration time.Duration
	for _, record := range records {
		startUTC := mustParseGalileanEventTime(t, record.StartUTC)
		endUTC := mustParseGalileanEventTime(t, record.EndUTC)
		startDuration := time.Duration(record.StartDurationMinutes * float64(time.Minute))
		endDuration := time.Duration(record.EndDurationMinutes * float64(time.Minute))
		endOfEvent := endUTC.Add(endDuration)
		midpoint := startUTC.Add(endOfEvent.Sub(startUTC) / 2)

		got := ClosestGalileanPhenomenonContactEvent(midpoint, record.Satellite, GalileanPhenomenonType(record.Type))
		if !got.Valid {
			t.Fatalf("%s returned no contact event", record.Label)
		}
		contacts := []struct {
			name string
			got  time.Time
			want time.Time
		}{
			{"D1", got.Disappearance.Start, startUTC},
			{"D2", got.Disappearance.End, startUTC.Add(startDuration)},
			{"F1", got.Reappearance.Start, endUTC},
			{"F2", got.Reappearance.End, endOfEvent},
		}
		for _, contact := range contacts {
			delta := contact.got.Sub(contact.want)
			if delta < 0 {
				delta = -delta
			}
			if delta > maxContact {
				maxContact = delta
			}
			if delta > contactTolerance {
				t.Fatalf("%s %s offset %v exceeds %v", record.Label, contact.name, contact.got.Sub(contact.want), contactTolerance)
			}
		}
		durations := []struct {
			name string
			got  time.Duration
			want time.Duration
		}{
			{"D", got.Disappearance.Duration, startDuration},
			{"F", got.Reappearance.Duration, endDuration},
		}
		for _, duration := range durations {
			delta := duration.got - duration.want
			if delta < 0 {
				delta = -delta
			}
			if delta > maxDuration {
				maxDuration = delta
			}
			if delta > durationTolerance {
				t.Fatalf("%s %s-phase duration offset %v exceeds %v", record.Label, duration.name, duration.got-duration.want, durationTolerance)
			}
		}
	}
	t.Logf("contact offset max %v, duration offset max %v", maxContact, maxDuration)
}
