package basic

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"
)

type sunPhysicalSample struct {
	InputUTC string  `json:"input_utc"`
	P        float64 `json:"p"`
	B0       float64 `json:"b0"`
	L0       float64 `json:"l0"`
}

func TestSunPhysicalMatchesHorizonsBaseline(t *testing.T) {
	// Baseline is generated from JPL Horizons by scripts/generate_sun_physical_baseline.sh.
	data, err := os.ReadFile("testdata/sun_physical_baseline.json")
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}

	var samples []sunPhysicalSample
	if err := json.Unmarshal(data, &samples); err != nil {
		t.Fatalf("decode baseline: %v", err)
	}

	var maxPDiff float64
	var maxB0Diff float64
	var maxL0Diff float64
	for _, sample := range samples {
		date, err := time.Parse(time.RFC3339, sample.InputUTC)
		if err != nil {
			t.Fatalf("parse sample time %q: %v", sample.InputUTC, err)
		}
		jd := TD2UT(Date2JDE(date.UTC()), true)
		got := SunPhysical(jd)

		pDiff := angleDiffAbs(got.P, sample.P)
		if pDiff > maxPDiff {
			maxPDiff = pDiff
		}
		b0Diff := math.Abs(got.B0 - sample.B0)
		if b0Diff > maxB0Diff {
			maxB0Diff = b0Diff
		}
		l0Diff := angleDiffAbs(got.L0, sample.L0)
		if l0Diff > maxL0Diff {
			maxL0Diff = l0Diff
		}

		assertPlanetPhaseClose(t, sample.InputUTC+".P", got.P, sample.P, 0.01)
		assertPlanetPhaseClose(t, sample.InputUTC+".B0", got.B0, sample.B0, 0.01)
		assertPlanetPhaseClose(t, sample.InputUTC+".L0", got.L0, sample.L0, 0.05)
	}

	t.Logf("sun physical max diff: P=%.6fdeg B0=%.6fdeg L0=%.6fdeg", maxPDiff, maxB0Diff, maxL0Diff)
}

func TestSunPhysicalNFullMatchesDefault(t *testing.T) {
	jd := TD2UT(Date2JDE(time.Date(2026, 4, 28, 9, 30, 45, 0, time.UTC)), true)

	got := SunPhysical(jd)
	gotN := SunPhysicalN(jd, -1)
	if math.Float64bits(got.P) != math.Float64bits(gotN.P) ||
		math.Float64bits(got.B0) != math.Float64bits(gotN.B0) ||
		math.Float64bits(got.L0) != math.Float64bits(gotN.L0) {
		t.Fatalf("SunPhysical full-n mismatch: got %+v want %+v", got, gotN)
	}
}

func TestSunPhysicalSampleSweepFiniteAndInRange(t *testing.T) {
	dates := []time.Time{
		time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1969, 7, 20, 20, 17, 40, 0, time.UTC),
		time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 28, 9, 30, 45, 0, time.UTC),
		time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	for _, date := range dates {
		jd := TD2UT(Date2JDE(date.UTC()), true)
		got := SunPhysical(jd)
		prefix := date.Format(time.RFC3339)

		assertFiniteRange(t, prefix+".P", got.P, 0, 360, true)
		assertFiniteRange(t, prefix+".B0", got.B0, -90, 90, false)
		assertFiniteRange(t, prefix+".L0", got.L0, 0, 360, true)
	}
}

func angleDiffAbs(got, want float64) float64 {
	diff := math.Abs(got - want)
	if diff > 180 {
		diff = 360 - diff
	}
	return diff
}
