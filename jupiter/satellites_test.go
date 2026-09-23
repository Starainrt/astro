package jupiter

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

// 基线是 UTC 瞬时地心赤道量的派生偏移，容差取实测最大残差（0.054″）的约三倍。
const galileanHorizonsToleranceArcsec = 0.15

type galileanHorizonsEquatorialRecord struct {
	RA    float64 `json:"ra_deg"`
	Dec   float64 `json:"dec_deg"`
	Delta float64 `json:"delta_au"`
}

type galileanHorizonsOffsetRecord struct {
	RA            float64 `json:"ra_deg"`
	Dec           float64 `json:"dec_deg"`
	Delta         float64 `json:"delta_au"`
	OffsetXArcsec float64 `json:"offset_x_arcsec"`
	OffsetYArcsec float64 `json:"offset_y_arcsec"`
}

type galileanHorizonsSample struct {
	UTC        string                                  `json:"utc"`
	Jupiter    galileanHorizonsEquatorialRecord        `json:"jupiter"`
	Satellites map[string]galileanHorizonsOffsetRecord `json:"satellites"`
}

func TestGalileanSatellitesAgainstHorizonsRelativeOffsets(t *testing.T) {
	samples := loadGalileanHorizonsBaseline(t)
	maxXDiff := 0.0
	maxYDiff := 0.0
	for _, sample := range samples {
		date, err := time.Parse(time.RFC3339, sample.UTC)
		if err != nil {
			t.Fatalf("parse %s: %v", sample.UTC, err)
		}
		got := Satellites(date)
		for name, want := range sample.Satellites {
			position := selectGalileanSatellite(got, name)
			xDiff := math.Abs(position.OffsetXArcsec - want.OffsetXArcsec)
			yDiff := math.Abs(position.OffsetYArcsec - want.OffsetYArcsec)
			if xDiff > maxXDiff {
				maxXDiff = xDiff
			}
			if yDiff > maxYDiff {
				maxYDiff = yDiff
			}
			if xDiff > galileanHorizonsToleranceArcsec {
				t.Fatalf("%s X mismatch at %s: got %.6f want %.6f", name, sample.UTC, position.OffsetXArcsec, want.OffsetXArcsec)
			}
			if yDiff > galileanHorizonsToleranceArcsec {
				t.Fatalf("%s Y mismatch at %s: got %.6f want %.6f", name, sample.UTC, position.OffsetYArcsec, want.OffsetYArcsec)
			}
			wantFront := want.Delta < sample.Jupiter.Delta
			if position.InFrontOfJupiter != wantFront {
				t.Fatalf("%s front/back mismatch at %s: got %v want %v", name, sample.UTC, position.InFrontOfJupiter, wantFront)
			}
		}
	}
	t.Logf("galilean Horizons max diff: X=%.3f arcsec Y=%.3f arcsec", maxXDiff, maxYDiff)
}

func loadGalileanHorizonsBaseline(t *testing.T) []galileanHorizonsSample {
	t.Helper()
	data, err := os.ReadFile("testdata/galilean_satellites_horizons.json")
	if err != nil {
		t.Fatal(err)
	}
	var samples []galileanHorizonsSample
	if err := json.Unmarshal(data, &samples); err != nil {
		t.Fatal(err)
	}
	if len(samples) == 0 {
		t.Fatal("empty Galilean baseline")
	}
	return samples
}

func selectGalileanSatellite(info GalileanSatellitesInfo, name string) GalileanSatellitePosition {
	switch name {
	case "io":
		return info.Io
	case "europa":
		return info.Europa
	case "ganymede":
		return info.Ganymede
	case "callisto":
		return info.Callisto
	default:
		panic("unknown satellite: " + name)
	}
}

func TestGalileanSatellitesUseTTFrame(t *testing.T) {
	dates := []time.Time{
		time.Date(1973, 4, 15, 0, 0, 0, 0, time.UTC),
		time.Date(1999, 1, 1, 2, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 2, 18, 0, 0, 0, time.UTC),
	}
	for _, date := range dates {
		want := basic.JupiterGalileanSatelliteObservations(basic.UTC2TT(basic.Date2JD(date.UTC())))
		got := Satellites(date)
		positions := []GalileanSatellitePosition{got.Io, got.Europa, got.Ganymede, got.Callisto}
		for i, position := range positions {
			expected := galileanSatellitePositionFromBasic(want[i])
			if position != expected {
				t.Fatalf("date=%s satellite=%d TT frame mismatch: got %+v want %+v", date, i+1, position, expected)
			}
		}
	}
}
