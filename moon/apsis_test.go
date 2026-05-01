package moon

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestApsisWrappersMatchBasic(t *testing.T) {
	perigees := basic.MoonPerigees(2026, time.January)
	perigeesWrapped := PerigeesInMonth(2026, time.January)
	if len(perigeesWrapped) != len(perigees) {
		t.Fatalf("perigee count mismatch: got %d want %d", len(perigeesWrapped), len(perigees))
	}
	for i, event := range perigees {
		wantTime := basic.JDE2DateByZone(event.JDE, time.UTC, false)
		if !perigeesWrapped[i].Time.Equal(wantTime) {
			t.Fatalf("perigee #%d time mismatch: got %s want %s", i+1, perigeesWrapped[i].Time.Format(time.RFC3339Nano), wantTime.Format(time.RFC3339Nano))
		}
		if math.Float64bits(perigeesWrapped[i].Distance) != math.Float64bits(event.Distance) {
			t.Fatalf("perigee #%d distance mismatch: got %.6f want %.6f", i+1, perigeesWrapped[i].Distance, event.Distance)
		}
	}

	apogees := basic.MoonApogees(2026, time.June)
	apogeesWrapped := ApogeesInMonth(2026, time.June)
	if len(apogeesWrapped) != len(apogees) {
		t.Fatalf("apogee count mismatch: got %d want %d", len(apogeesWrapped), len(apogees))
	}
	for i, event := range apogees {
		wantTime := basic.JDE2DateByZone(event.JDE, time.UTC, false)
		if !apogeesWrapped[i].Time.Equal(wantTime) {
			t.Fatalf("apogee #%d time mismatch: got %s want %s", i+1, apogeesWrapped[i].Time.Format(time.RFC3339Nano), wantTime.Format(time.RFC3339Nano))
		}
		if math.Float64bits(apogeesWrapped[i].Distance) != math.Float64bits(event.Distance) {
			t.Fatalf("apogee #%d distance mismatch: got %.6f want %.6f", i+1, apogeesWrapped[i].Distance, event.Distance)
		}
	}
}
