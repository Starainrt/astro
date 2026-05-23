package moon

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestConjunctionPlanetWrappersMatchBasic(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	query := time.Date(2026, 1, 15, 20, 0, 0, 0, loc)
	queryTT := basic.TD2UT(basic.Date2JDE(query.UTC()), true)

	cases := []struct {
		name   string
		planet ConjunctionPlanet
		basic  basic.MoonPlanetConjunctionPlanet
	}{
		{name: "Mercury", planet: ConjunctionMercury, basic: basic.MoonPlanetConjunctionMercury},
		{name: "Venus", planet: ConjunctionVenus, basic: basic.MoonPlanetConjunctionVenus},
		{name: "Mars", planet: ConjunctionMars, basic: basic.MoonPlanetConjunctionMars},
		{name: "Jupiter", planet: ConjunctionJupiter, basic: basic.MoonPlanetConjunctionJupiter},
		{name: "Saturn", planet: ConjunctionSaturn, basic: basic.MoonPlanetConjunctionSaturn},
		{name: "Uranus", planet: ConjunctionUranus, basic: basic.MoonPlanetConjunctionUranus},
		{name: "Neptune", planet: ConjunctionNeptune, basic: basic.MoonPlanetConjunctionNeptune},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertSameConjunctionTime(t, "last", LastConjunctionWithPlanet(query, tc.planet), basic.LastMoonPlanetConjunction(queryTT, tc.basic), loc)
			assertSameConjunctionTime(t, "next", NextConjunctionWithPlanet(query, tc.planet), basic.NextMoonPlanetConjunction(queryTT, tc.basic), loc)
			assertSameConjunctionTime(t, "closest", ClosestConjunctionWithPlanet(query, tc.planet), basic.ClosestMoonPlanetConjunction(queryTT, tc.basic), loc)
		})
	}
}

func assertSameConjunctionTime(t *testing.T, name string, got time.Time, wantJDE float64, loc *time.Location) {
	t.Helper()
	want := basic.JDE2DateByZone(wantJDE, loc, false)
	if got.Location() != loc {
		t.Fatalf("%s location mismatch: got %q want %q", name, got.Location().String(), loc.String())
	}
	if !got.Equal(want) {
		t.Fatalf("%s time mismatch: got %s want %s", name, got.Format(time.RFC3339Nano), want.Format(time.RFC3339Nano))
	}
}

func TestClosestConjunctionReturnsNearestCandidate(t *testing.T) {
	query := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	last := LastConjunctionWithPlanet(query, ConjunctionMercury)
	next := NextConjunctionWithPlanet(query, ConjunctionMercury)
	got := ClosestConjunctionWithPlanet(query, ConjunctionMercury)

	lastDiff := math.Abs(query.Sub(last).Seconds())
	nextDiff := math.Abs(next.Sub(query).Seconds())
	if lastDiff <= nextDiff {
		if !got.Equal(last) {
			t.Fatalf("closest should match last: got %s want %s", got.Format(time.RFC3339Nano), last.Format(time.RFC3339Nano))
		}
		return
	}
	if !got.Equal(next) {
		t.Fatalf("closest should match next: got %s want %s", got.Format(time.RFC3339Nano), next.Format(time.RFC3339Nano))
	}
}

func TestInvalidConjunctionPlanetReturnsZeroTime(t *testing.T) {
	query := time.Date(2026, 1, 15, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	invalid := ConjunctionPlanet("pluto")

	for name, fn := range map[string]func(time.Time, ConjunctionPlanet) time.Time{
		"last":    LastConjunctionWithPlanet,
		"next":    NextConjunctionWithPlanet,
		"closest": ClosestConjunctionWithPlanet,
	} {
		if got := fn(query, invalid); !got.IsZero() {
			t.Fatalf("%s should return zero time for invalid planet, got %s", name, got.Format(time.RFC3339Nano))
		}
	}
}

func TestNextConjunctionAdvancesPastReturnedEvent(t *testing.T) {
	cursor := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	first := NextConjunctionWithPlanet(cursor, ConjunctionMercury)
	query := first.Add(time.Second)
	next := NextConjunctionWithPlanet(query, ConjunctionMercury)

	if !next.After(query) {
		t.Fatalf("expected next conjunction after query: query=%s next=%s delta=%v",
			query.Format(time.RFC3339Nano),
			next.Format(time.RFC3339Nano),
			next.Sub(query),
		)
	}
	if next.Equal(first) {
		t.Fatalf("expected next conjunction to advance: first=%s next=%s",
			first.Format(time.RFC3339Nano),
			next.Format(time.RFC3339Nano),
		)
	}
}
