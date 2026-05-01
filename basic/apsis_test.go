package basic

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"
)

type earthApsisSample struct {
	Kind       string  `json:"kind"`
	Year       int     `json:"year"`
	TimeUTC    string  `json:"time_utc"`
	DistanceAU float64 `json:"distance_au"`
}

type moonApsisSample struct {
	Kind       string  `json:"kind"`
	Year       int     `json:"year"`
	Month      int     `json:"month"`
	TimeUTC    string  `json:"time_utc"`
	DistanceKM float64 `json:"distance_km"`
}

type moonApsisMonthState struct {
	perigees []ApsisEvent
	apogees  []ApsisEvent
	perigeeI int
	apogeeI  int
}

func TestEarthApsisMatchesHorizonsBaseline(t *testing.T) {
	data, err := os.ReadFile("testdata/earth_apsis_baseline.json")
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}

	var samples []earthApsisSample
	if err := json.Unmarshal(data, &samples); err != nil {
		t.Fatalf("decode baseline: %v", err)
	}

	const timeTolerance = 2 * time.Minute
	const distanceToleranceAU = 5e-8

	var maxTimeDiff time.Duration
	var maxDistanceDiff float64
	for _, sample := range samples {
		wantTime, err := time.Parse(time.RFC3339Nano, sample.TimeUTC)
		if err != nil {
			t.Fatalf("parse sample time %q: %v", sample.TimeUTC, err)
		}

		var got ApsisEvent
		switch sample.Kind {
		case "perihelion":
			got = EarthPerihelion(sample.Year)
		case "aphelion":
			got = EarthAphelion(sample.Year)
		default:
			t.Fatalf("unknown earth apsis kind %q", sample.Kind)
		}

		gotTime := JDE2DateByZone(got.JDE, time.UTC, false)
		timeDiff := gotTime.Sub(wantTime)
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}
		if timeDiff > maxTimeDiff {
			maxTimeDiff = timeDiff
		}
		if timeDiff > timeTolerance {
			t.Fatalf("%s %d time mismatch: got %s want %s tolerance %v", sample.Kind, sample.Year, gotTime.Format(time.RFC3339Nano), sample.TimeUTC, timeTolerance)
		}

		distanceDiff := math.Abs(got.Distance - sample.DistanceAU)
		if distanceDiff > maxDistanceDiff {
			maxDistanceDiff = distanceDiff
		}
		if distanceDiff > distanceToleranceAU {
			t.Fatalf("%s %d distance mismatch: got %.12f want %.12f tolerance %.12f", sample.Kind, sample.Year, got.Distance, sample.DistanceAU, distanceToleranceAU)
		}
	}

	t.Logf("earth apsis max diff: time=%v distance=%.12f AU", maxTimeDiff, maxDistanceDiff)
}

func TestMoonApsisMatchesHorizonsBaseline(t *testing.T) {
	// Baseline is generated from JPL Horizons by scripts/generate_moon_apsis_baseline.sh.
	data, err := os.ReadFile("testdata/moon_apsis_baseline.json")
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}

	var samples []moonApsisSample
	if err := json.Unmarshal(data, &samples); err != nil {
		t.Fatalf("decode baseline: %v", err)
	}

	const timeTolerance = 20 * time.Minute
	const distanceToleranceKM = 50.0

	states := make(map[int]*moonApsisMonthState)
	var maxTimeDiff time.Duration
	var maxDistanceDiff float64

	for _, sample := range samples {
		wantTime, err := time.Parse(time.RFC3339Nano, sample.TimeUTC)
		if err != nil {
			t.Fatalf("parse sample time %q: %v", sample.TimeUTC, err)
		}

		key := sample.Year*100 + sample.Month
		state := states[key]
		if state == nil {
			state = &moonApsisMonthState{
				perigees: MoonPerigees(sample.Year, time.Month(sample.Month)),
				apogees:  MoonApogees(sample.Year, time.Month(sample.Month)),
			}
			states[key] = state
		}

		var got ApsisEvent
		switch sample.Kind {
		case "perigee":
			if state.perigeeI >= len(state.perigees) {
				t.Fatalf("%04d-%02d missing perigee #%d", sample.Year, sample.Month, state.perigeeI+1)
			}
			got = state.perigees[state.perigeeI]
			state.perigeeI++
		case "apogee":
			if state.apogeeI >= len(state.apogees) {
				t.Fatalf("%04d-%02d missing apogee #%d", sample.Year, sample.Month, state.apogeeI+1)
			}
			got = state.apogees[state.apogeeI]
			state.apogeeI++
		default:
			t.Fatalf("unknown moon apsis kind %q", sample.Kind)
		}

		gotTime := JDE2DateByZone(got.JDE, time.UTC, false)
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

		distanceDiff := math.Abs(got.Distance - sample.DistanceKM)
		if distanceDiff > maxDistanceDiff {
			maxDistanceDiff = distanceDiff
		}
		if distanceDiff > distanceToleranceKM {
			t.Fatalf("%s %04d-%02d distance mismatch: got %.6f want %.6f tolerance %.6f", sample.Kind, sample.Year, sample.Month, got.Distance, sample.DistanceKM, distanceToleranceKM)
		}
	}

	for key, state := range states {
		year := key / 100
		month := key % 100
		if state.perigeeI != len(state.perigees) {
			t.Fatalf("%04d-%02d unconsumed perigees: got %d of %d", year, month, state.perigeeI, len(state.perigees))
		}
		if state.apogeeI != len(state.apogees) {
			t.Fatalf("%04d-%02d unconsumed apogees: got %d of %d", year, month, state.apogeeI, len(state.apogees))
		}
	}

	t.Logf("moon apsis max diff: time=%v distance=%.6f km", maxTimeDiff, maxDistanceDiff)
}
