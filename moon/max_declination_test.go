package moon

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestMaximumDeclinationWrappersMatchBasic(t *testing.T) {
	north := basic.MoonMaximumNorthDeclinations(2026, time.January)
	northWrapped := MaximumNorthDeclinationsInMonth(2026, time.January)
	if len(northWrapped) != len(north) {
		t.Fatalf("north count mismatch: got %d want %d", len(northWrapped), len(north))
	}
	for i, event := range north {
		wantTime := basic.JDE2DateByZone(event.JDE, time.UTC, false)
		if !northWrapped[i].Time.Equal(wantTime) {
			t.Fatalf("north #%d time mismatch: got %s want %s", i+1, northWrapped[i].Time.Format(time.RFC3339Nano), wantTime.Format(time.RFC3339Nano))
		}
		if math.Float64bits(northWrapped[i].Declination) != math.Float64bits(event.Declination) {
			t.Fatalf("north #%d declination mismatch: got %.8f want %.8f", i+1, northWrapped[i].Declination, event.Declination)
		}
	}

	south := basic.MoonMaximumSouthDeclinations(2026, time.June)
	southWrapped := MaximumSouthDeclinationsInMonth(2026, time.June)
	if len(southWrapped) != len(south) {
		t.Fatalf("south count mismatch: got %d want %d", len(southWrapped), len(south))
	}
	for i, event := range south {
		wantTime := basic.JDE2DateByZone(event.JDE, time.UTC, false)
		if !southWrapped[i].Time.Equal(wantTime) {
			t.Fatalf("south #%d time mismatch: got %s want %s", i+1, southWrapped[i].Time.Format(time.RFC3339Nano), wantTime.Format(time.RFC3339Nano))
		}
		if math.Float64bits(southWrapped[i].Declination) != math.Float64bits(event.Declination) {
			t.Fatalf("south #%d declination mismatch: got %.8f want %.8f", i+1, southWrapped[i].Declination, event.Declination)
		}
	}
}

func TestMaximumDeclinationSearchWrappersMatchBasic(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	query := time.Date(2026, time.January, 10, 18, 30, 0, 0, loc)
	queryJDE := basic.Date2JDE(query.UTC())

	assertMaximumDeclinationInfoMatchesBasic(t, "last north", LastMaximumNorthDeclination(query), basic.LastMoonMaximumNorthDeclination(queryJDE), loc)
	assertMaximumDeclinationInfoMatchesBasic(t, "next north", NextMaximumNorthDeclination(query), basic.NextMoonMaximumNorthDeclination(queryJDE), loc)
	assertMaximumDeclinationInfoMatchesBasic(t, "closest north", ClosestMaximumNorthDeclination(query), basic.ClosestMoonMaximumNorthDeclination(queryJDE), loc)

	assertMaximumDeclinationInfoMatchesBasic(t, "last south", LastMaximumSouthDeclination(query), basic.LastMoonMaximumSouthDeclination(queryJDE), loc)
	assertMaximumDeclinationInfoMatchesBasic(t, "next south", NextMaximumSouthDeclination(query), basic.NextMoonMaximumSouthDeclination(queryJDE), loc)
	assertMaximumDeclinationInfoMatchesBasic(t, "closest south", ClosestMaximumSouthDeclination(query), basic.ClosestMoonMaximumSouthDeclination(queryJDE), loc)
}

func assertMaximumDeclinationInfoMatchesBasic(t *testing.T, name string, got MaximumDeclinationInfo, want basic.DeclinationEvent, loc *time.Location) {
	t.Helper()
	wantTime := basic.JDE2DateByZone(want.JDE, loc, false)
	if got.Time.Location() != loc {
		t.Fatalf("%s location mismatch: got %q want %q", name, got.Time.Location().String(), loc.String())
	}
	if !got.Time.Equal(wantTime) {
		t.Fatalf("%s time mismatch: got %s want %s", name, got.Time.Format(time.RFC3339Nano), wantTime.Format(time.RFC3339Nano))
	}
	if math.Float64bits(got.Declination) != math.Float64bits(want.Declination) {
		t.Fatalf("%s declination mismatch: got %.8f want %.8f", name, got.Declination, want.Declination)
	}
}
