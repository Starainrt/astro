package basic

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"
)

type moonPlanetConjunctionBaselineSample struct {
	Planet  string `json:"planet"`
	Year    int    `json:"year"`
	Month   int    `json:"month"`
	TimeUTC string `json:"time_utc"`
}

type moonPlanetConjunctionBaseline struct {
	Samples []moonPlanetConjunctionBaselineSample `json:"samples"`
}

func loadMoonPlanetConjunctionBaseline(t *testing.T) moonPlanetConjunctionBaseline {
	t.Helper()

	paths := [][]string{
		{
			"testdata/moon_planet_conjunction_baseline.json",
			"basic/testdata/moon_planet_conjunction_baseline.json",
		},
		{
			"testdata/moon_planet_conjunction_baseline_samples.json",
			"basic/testdata/moon_planet_conjunction_baseline_samples.json",
		},
	}

	var merged moonPlanetConjunctionBaseline
	for index, candidates := range paths {
		var (
			data []byte
			err  error
		)
		for _, path := range candidates {
			data, err = os.ReadFile(path)
			if err == nil {
				var baseline moonPlanetConjunctionBaseline
				if err := json.Unmarshal(data, &baseline); err != nil {
					t.Fatalf("decode baseline %s: %v", path, err)
				}
				merged.Samples = append(merged.Samples, baseline.Samples...)
				break
			}
		}
		if err != nil && index == 0 {
			t.Fatalf("read baseline: %v", err)
		}
	}
	if len(merged.Samples) == 0 {
		t.Fatal("empty moon-planet conjunction baseline")
	}
	return merged
}

func TestMoonPlanetConjunctionsMatchHorizonsBaseline(t *testing.T) {
	baseline := loadMoonPlanetConjunctionBaseline(t)

	type conjunctionCase struct {
		planet MoonPlanetConjunctionPlanet
		next   func(float64, MoonPlanetConjunctionPlanet) float64
	}

	cases := map[string]conjunctionCase{
		"mercury": {planet: MoonPlanetConjunctionMercury, next: NextMoonPlanetConjunction},
		"venus":   {planet: MoonPlanetConjunctionVenus, next: NextMoonPlanetConjunction},
		"mars":    {planet: MoonPlanetConjunctionMars, next: NextMoonPlanetConjunction},
		"jupiter": {planet: MoonPlanetConjunctionJupiter, next: NextMoonPlanetConjunction},
		"saturn":  {planet: MoonPlanetConjunctionSaturn, next: NextMoonPlanetConjunction},
		"uranus":  {planet: MoonPlanetConjunctionUranus, next: NextMoonPlanetConjunction},
		"neptune": {planet: MoonPlanetConjunctionNeptune, next: NextMoonPlanetConjunction},
	}

	const tolerance = 20 * time.Second
	var maxDiff time.Duration

	seen := make(map[string]int, len(cases))
	for _, sample := range baseline.Samples {
		tc, ok := cases[sample.Planet]
		if !ok {
			t.Fatalf("unknown planet %q", sample.Planet)
		}

		wantTime, err := time.Parse(time.RFC3339Nano, sample.TimeUTC)
		if err != nil {
			t.Fatalf("parse sample time %q: %v", sample.TimeUTC, err)
		}
		queryTT := TD2UT(Date2JDE(wantTime.Add(-12*time.Hour).UTC()), true)
		gotUT := tc.next(queryTT, tc.planet)
		gotTime := JDE2DateByZone(gotUT, time.UTC, false)
		diff := gotTime.Sub(wantTime)
		if diff < 0 {
			diff = -diff
		}
		if diff > maxDiff {
			maxDiff = diff
		}
		if diff > tolerance {
			t.Fatalf("%s %04d-%02d time mismatch: got %s want %s tolerance %v", sample.Planet, sample.Year, sample.Month, gotTime.Format(time.RFC3339Nano), sample.TimeUTC, tolerance)
		}

		delta := math.Abs(moonPlanetConjunctionDeltaAt(TD2UT(gotUT, true), tc.planet, -1))
		if delta > 0.01 {
			t.Fatalf("%s %04d-%02d event not near conjunction: delta=%.8f deg", sample.Planet, sample.Year, sample.Month, delta)
		}
		seen[sample.Planet]++
	}

	for planet := range cases {
		if seen[planet] == 0 {
			t.Fatalf("missing baseline samples for %s", planet)
		}
	}

	t.Logf("moon-planet conjunction max diff: time=%v", maxDiff)
}

func TestMoonPlanetConjunctionDirectionalConsistencyAroundBaseline(t *testing.T) {
	baseline := loadMoonPlanetConjunctionBaseline(t)

	planets := map[string]MoonPlanetConjunctionPlanet{
		"mercury": MoonPlanetConjunctionMercury,
		"venus":   MoonPlanetConjunctionVenus,
		"mars":    MoonPlanetConjunctionMars,
		"jupiter": MoonPlanetConjunctionJupiter,
		"saturn":  MoonPlanetConjunctionSaturn,
		"uranus":  MoonPlanetConjunctionUranus,
		"neptune": MoonPlanetConjunctionNeptune,
	}

	for _, sample := range baseline.Samples {
		planet, ok := planets[sample.Planet]
		if !ok {
			t.Fatalf("unknown planet %q", sample.Planet)
		}
		wantTime, err := time.Parse(time.RFC3339Nano, sample.TimeUTC)
		if err != nil {
			t.Fatalf("parse sample time %q: %v", sample.TimeUTC, err)
		}
		queryAtTT := TD2UT(Date2JDE(wantTime.UTC()), true)
		queryAfterTT := TD2UT(Date2JDE(wantTime.Add(time.Hour).UTC()), true)

		exactNext := NextMoonPlanetConjunction(queryAtTT, planet)
		exactClosest := ClosestMoonPlanetConjunction(queryAtTT, planet)
		exactLastAfter := LastMoonPlanetConjunction(queryAfterTT, planet)

		wantUT := Date2JDE(wantTime.UTC())
		for name, gotUT := range map[string]float64{
			"exactNext":      exactNext,
			"exactClosest":   exactClosest,
			"lastAfterEvent": exactLastAfter,
		} {
			gotTime := JDE2DateByZone(gotUT, time.UTC, false)
			if diff := math.Abs(gotUT - wantUT); diff > 5.0/86400.0 {
				t.Fatalf("%s %s mismatch: got %s want %s diff=%v", sample.Planet, name, gotTime.Format(time.RFC3339Nano), sample.TimeUTC, diff*86400)
			}
		}
	}
}

func TestMoonPlanetConjunctionRejectsOppositionBranchJump(t *testing.T) {
	query := time.Date(1900, 11, 10, 12, 0, 0, 0, time.UTC)
	queryTT := TD2UT(Date2JDE(query), true)

	lastUT := LastMoonPlanetConjunction(queryTT, MoonPlanetConjunctionSaturn)
	nextUT := NextMoonPlanetConjunction(queryTT, MoonPlanetConjunctionSaturn)

	if math.Abs(lastUT-Date2JDE(query)) <= 5.0/86400.0 {
		t.Fatalf("last returned query time on branch jump: got %s", JDE2DateByZone(lastUT, time.UTC, false).Format(time.RFC3339Nano))
	}
	if math.Abs(nextUT-Date2JDE(query)) <= 5.0/86400.0 {
		t.Fatalf("next returned query time on branch jump: got %s", JDE2DateByZone(nextUT, time.UTC, false).Format(time.RFC3339Nano))
	}

	for name, gotUT := range map[string]float64{
		"last": lastUT,
		"next": nextUT,
	} {
		delta := math.Abs(moonPlanetConjunctionDeltaAt(TD2UT(gotUT, true), MoonPlanetConjunctionSaturn, -1))
		if delta > moonPlanetConjunctionEventTolerance {
			t.Fatalf("%s returned non-event candidate: delta=%.8f event=%s", name, delta, JDE2DateByZone(gotUT, time.UTC, false).Format(time.RFC3339Nano))
		}
	}
}

func TestMoonPlanetConjunctionDirectionalOrderingOnSampleQueries(t *testing.T) {
	samples := []struct {
		planet MoonPlanetConjunctionPlanet
		query  time.Time
	}{
		{planet: MoonPlanetConjunctionSaturn, query: time.Date(1700, 4, 15, 12, 0, 0, 0, time.UTC)},
		{planet: MoonPlanetConjunctionMercury, query: time.Date(1900, 1, 14, 12, 0, 0, 0, time.UTC)},
		{planet: MoonPlanetConjunctionVenus, query: time.Date(1950, 6, 3, 12, 0, 0, 0, time.UTC)},
		{planet: MoonPlanetConjunctionMars, query: time.Date(2000, 2, 29, 18, 0, 0, 0, time.UTC)},
		{planet: MoonPlanetConjunctionJupiter, query: time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)},
		{planet: MoonPlanetConjunctionSaturn, query: time.Date(2100, 8, 17, 6, 0, 0, 0, time.UTC)},
		{planet: MoonPlanetConjunctionUranus, query: time.Date(2200, 11, 2, 9, 0, 0, 0, time.UTC)},
		{planet: MoonPlanetConjunctionNeptune, query: time.Date(2300, 4, 24, 3, 0, 0, 0, time.UTC)},
	}

	for _, sample := range samples {
		queryTT := TD2UT(Date2JDE(sample.query.UTC()), true)
		lastUT := LastMoonPlanetConjunction(queryTT, sample.planet)
		nextUT := NextMoonPlanetConjunction(queryTT, sample.planet)
		closestUT := ClosestMoonPlanetConjunction(queryTT, sample.planet)

		if math.IsNaN(lastUT) || math.IsNaN(nextUT) || math.IsNaN(closestUT) {
			t.Fatalf("planet=%v query=%s returned NaN event(s): last=%v next=%v closest=%v", sample.planet, sample.query.Format(time.RFC3339), lastUT, nextUT, closestUT)
		}
		if !eventUTQueryBeforeOrEqual(lastUT, queryTT) {
			t.Fatalf("planet=%v last after query: last=%s query=%s", sample.planet, JDE2DateByZone(lastUT, time.UTC, false).Format(time.RFC3339Nano), sample.query.Format(time.RFC3339Nano))
		}
		if !eventUTQueryAfterOrEqual(nextUT, queryTT) {
			t.Fatalf("planet=%v next before query: next=%s query=%s", sample.planet, JDE2DateByZone(nextUT, time.UTC, false).Format(time.RFC3339Nano), sample.query.Format(time.RFC3339Nano))
		}
		if closestUT != closestEventUTToQueryTT(queryTT, lastUT, nextUT) {
			t.Fatalf("planet=%v closest mismatch: got=%s want=%s", sample.planet, JDE2DateByZone(closestUT, time.UTC, false).Format(time.RFC3339Nano), JDE2DateByZone(closestEventUTToQueryTT(queryTT, lastUT, nextUT), time.UTC, false).Format(time.RFC3339Nano))
		}
		for name, gotUT := range map[string]float64{
			"last":    lastUT,
			"next":    nextUT,
			"closest": closestUT,
		} {
			delta := math.Abs(moonPlanetConjunctionDeltaAt(TD2UT(gotUT, true), sample.planet, -1))
			if delta > moonPlanetConjunctionEventTolerance {
				t.Fatalf("planet=%v %s returned non-event candidate: delta=%.8f event=%s", sample.planet, name, delta, JDE2DateByZone(gotUT, time.UTC, false).Format(time.RFC3339Nano))
			}
		}
	}
}

func TestMoonPlanetConjunctionKeepsImmediateNeighborEvents(t *testing.T) {
	query := time.Date(1700, 4, 15, 12, 0, 0, 0, time.UTC)
	queryTT := TD2UT(Date2JDE(query.UTC()), true)

	lastUT := LastMoonPlanetConjunction(queryTT, MoonPlanetConjunctionSaturn)
	nextUT := NextMoonPlanetConjunction(queryTT, MoonPlanetConjunctionSaturn)
	closestUT := ClosestMoonPlanetConjunction(queryTT, MoonPlanetConjunctionSaturn)

	wantLast := time.Date(1700, 4, 15, 11, 55, 59, 115569293, time.UTC)
	wantNext := time.Date(1700, 5, 13, 0, 35, 5, 981616675, time.UTC)
	const tolerance = 5.0 / 86400.0

	if diff := math.Abs(lastUT - Date2JDE(wantLast)); diff > tolerance {
		t.Fatalf("last mismatch: got=%s want=%s diff=%.3fs", JDE2DateByZone(lastUT, time.UTC, false).Format(time.RFC3339Nano), wantLast.Format(time.RFC3339Nano), diff*86400)
	}
	if diff := math.Abs(nextUT - Date2JDE(wantNext)); diff > tolerance {
		t.Fatalf("next mismatch: got=%s want=%s diff=%.3fs", JDE2DateByZone(nextUT, time.UTC, false).Format(time.RFC3339Nano), wantNext.Format(time.RFC3339Nano), diff*86400)
	}
	if !sameEventJD(closestUT, lastUT) {
		t.Fatalf("closest should keep immediate previous event: closest=%s last=%s", JDE2DateByZone(closestUT, time.UTC, false).Format(time.RFC3339Nano), JDE2DateByZone(lastUT, time.UTC, false).Format(time.RFC3339Nano))
	}
}
