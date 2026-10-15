package basic

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"
)

type moonMaxDeclinationSample struct {
	Kind           string  `json:"kind"`
	Year           int     `json:"year"`
	Month          int     `json:"month"`
	TimeUTC        string  `json:"time_utc"`
	DeclinationDeg float64 `json:"declination_deg"`
}

type moonMaxDeclinationMonthState struct {
	north  []DeclinationEvent
	south  []DeclinationEvent
	northI int
	southI int
}

func TestMoonMaximumDeclinationsMatchHorizonsBaseline(t *testing.T) {
	data, err := os.ReadFile("testdata/moon_max_declination_baseline.json")
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}

	var samples []moonMaxDeclinationSample
	if err := json.Unmarshal(data, &samples); err != nil {
		t.Fatalf("decode baseline: %v", err)
	}
	if len(samples) == 0 {
		t.Fatal("empty moon maximum declination baseline")
	}

	const timeTolerance = 15 * time.Second
	const declinationToleranceDeg = 0.0002

	states := make(map[int]*moonMaxDeclinationMonthState)
	var maxTimeDiff time.Duration
	var maxDeclinationDiff float64

	for _, sample := range samples {
		wantTime, err := time.Parse(time.RFC3339Nano, sample.TimeUTC)
		if err != nil {
			t.Fatalf("parse sample time %q: %v", sample.TimeUTC, err)
		}

		key := sample.Year*100 + sample.Month
		state := states[key]
		if state == nil {
			state = &moonMaxDeclinationMonthState{
				north: MoonMaximumNorthDeclinations(sample.Year, time.Month(sample.Month)),
				south: MoonMaximumSouthDeclinations(sample.Year, time.Month(sample.Month)),
			}
			states[key] = state
		}

		var got DeclinationEvent
		switch sample.Kind {
		case "north":
			if state.northI >= len(state.north) {
				t.Fatalf("%04d-%02d missing north declination event #%d", sample.Year, sample.Month, state.northI+1)
			}
			got = state.north[state.northI]
			state.northI++
		case "south":
			if state.southI >= len(state.south) {
				t.Fatalf("%04d-%02d missing south declination event #%d", sample.Year, sample.Month, state.southI+1)
			}
			got = state.south[state.southI]
			state.southI++
		default:
			t.Fatalf("unknown declination kind %q", sample.Kind)
		}

		gotTime := JD2DateByZone(got.JD, time.UTC, false)
		timeDiff := gotTime.Sub(wantTime)
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}
		if timeDiff > maxTimeDiff {
			maxTimeDiff = timeDiff
		}
		if timeDiff > timeTolerance {
			t.Fatalf("%s %04d-%02d time mismatch: got %s want %s tolerance %v", sample.Kind, sample.Year, sample.Month, gotTime.Format(time.RFC3339Nano), sample.TimeUTC, timeTolerance)
		}

		declinationDiff := math.Abs(got.Declination - sample.DeclinationDeg)
		if declinationDiff > maxDeclinationDiff {
			maxDeclinationDiff = declinationDiff
		}
		if declinationDiff > declinationToleranceDeg {
			t.Fatalf("%s %04d-%02d declination mismatch: got %.8f want %.8f tolerance %.8f", sample.Kind, sample.Year, sample.Month, got.Declination, sample.DeclinationDeg, declinationToleranceDeg)
		}
	}

	for key, state := range states {
		year := key / 100
		month := key % 100
		if state.northI != len(state.north) {
			t.Fatalf("%04d-%02d unconsumed north events: got %d of %d", year, month, state.northI, len(state.north))
		}
		if state.southI != len(state.south) {
			t.Fatalf("%04d-%02d unconsumed south events: got %d of %d", year, month, state.southI, len(state.south))
		}
	}

	t.Logf("moon maximum declination max diff: time=%v declination=%.8f deg", maxTimeDiff, maxDeclinationDiff)
}

func TestMoonMaximumDeclinationSignsAndOrder(t *testing.T) {
	north := MoonMaximumNorthDeclinations(2026, time.January)
	south := MoonMaximumSouthDeclinations(2026, time.January)
	if len(north) == 0 || len(south) == 0 {
		t.Fatalf("expected both north and south events in 2026-01, got north=%d south=%d", len(north), len(south))
	}
	for i, event := range north {
		if event.Declination <= 0 {
			t.Fatalf("north event #%d should be positive, got %.8f", i+1, event.Declination)
		}
		if i > 0 && !(north[i-1].JD < event.JD) {
			t.Fatalf("north events not strictly increasing: %.12f then %.12f", north[i-1].JD, event.JD)
		}
	}
	for i, event := range south {
		if event.Declination >= 0 {
			t.Fatalf("south event #%d should be negative, got %.8f", i+1, event.Declination)
		}
		if i > 0 && !(south[i-1].JD < event.JD) {
			t.Fatalf("south events not strictly increasing: %.12f then %.12f", south[i-1].JD, event.JD)
		}
	}
}

func TestMoonMaximumDeclinationSearchMatchesMonthlyEvents(t *testing.T) {
	query := time.Date(2026, time.January, 10, 0, 0, 0, 0, time.UTC)
	queryJD := Date2JD(query)

	northEvents := append([]DeclinationEvent{}, MoonMaximumNorthDeclinations(2025, time.December)...)
	northEvents = append(northEvents, MoonMaximumNorthDeclinations(2026, time.January)...)
	northEvents = append(northEvents, MoonMaximumNorthDeclinations(2026, time.February)...)

	southEvents := append([]DeclinationEvent{}, MoonMaximumSouthDeclinations(2025, time.December)...)
	southEvents = append(southEvents, MoonMaximumSouthDeclinations(2026, time.January)...)
	southEvents = append(southEvents, MoonMaximumSouthDeclinations(2026, time.February)...)

	assertSameDeclinationEvent(t, "last north", LastMoonMaximumNorthDeclination(queryJD), expectedDirectionalDeclinationEvent(northEvents, queryJD, -1, true))
	assertSameDeclinationEvent(t, "next north", NextMoonMaximumNorthDeclination(queryJD), expectedDirectionalDeclinationEvent(northEvents, queryJD, 1, false))
	assertSameDeclinationEvent(t, "closest north", ClosestMoonMaximumNorthDeclination(queryJD), expectedClosestDeclinationEvent(northEvents, queryJD))

	assertSameDeclinationEvent(t, "last south", LastMoonMaximumSouthDeclination(queryJD), expectedDirectionalDeclinationEvent(southEvents, queryJD, -1, true))
	assertSameDeclinationEvent(t, "next south", NextMoonMaximumSouthDeclination(queryJD), expectedDirectionalDeclinationEvent(southEvents, queryJD, 1, false))
	assertSameDeclinationEvent(t, "closest south", ClosestMoonMaximumSouthDeclination(queryJD), expectedClosestDeclinationEvent(southEvents, queryJD))
}

func TestMoonMaximumDeclinationSearchAtExactEventTime(t *testing.T) {
	north := MoonMaximumNorthDeclinations(2026, time.January)
	if len(north) < 2 {
		t.Fatalf("expected at least two north events spanning Jan 2026 search window, got %d", len(north))
	}

	exactJD := north[0].JD
	assertSameDeclinationEvent(t, "exact last north", LastMoonMaximumNorthDeclination(exactJD), north[0])
	assertSameDeclinationEvent(t, "exact closest north", ClosestMoonMaximumNorthDeclination(exactJD), north[0])
	assertSameDeclinationEvent(t, "exact next north", NextMoonMaximumNorthDeclination(exactJD), north[1])
}

func assertSameDeclinationEvent(t *testing.T, name string, got, want DeclinationEvent) {
	t.Helper()
	if math.Abs(got.JD-want.JD) > 1e-12 {
		t.Fatalf("%s JD mismatch: got %.12f want %.12f", name, got.JD, want.JD)
	}
	if math.Float64bits(got.Declination) != math.Float64bits(want.Declination) {
		t.Fatalf("%s declination mismatch: got %.12f want %.12f", name, got.Declination, want.Declination)
	}
}

func expectedDirectionalDeclinationEvent(events []DeclinationEvent, queryJD float64, direction int, includeCurrent bool) DeclinationEvent {
	var (
		found bool
		best  DeclinationEvent
	)
	for _, event := range events {
		delta := event.JD - queryJD
		if !moonMaximumDeclinationMatchesDirection(delta, direction, includeCurrent) {
			continue
		}
		if !found {
			best = event
			found = true
			continue
		}
		if math.Abs(delta) < math.Abs(best.JD-queryJD) || (math.Abs(delta) == math.Abs(best.JD-queryJD) && event.JD < best.JD) {
			best = event
		}
	}
	return best
}

func expectedClosestDeclinationEvent(events []DeclinationEvent, queryJD float64) DeclinationEvent {
	last := expectedDirectionalDeclinationEvent(events, queryJD, -1, true)
	next := expectedDirectionalDeclinationEvent(events, queryJD, 1, false)
	if math.Abs(queryJD-last.JD) <= math.Abs(next.JD-queryJD) {
		return last
	}
	return next
}
