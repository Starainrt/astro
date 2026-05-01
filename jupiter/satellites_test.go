package jupiter

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"
)

const galileanHorizonsToleranceArcsec = 1.0

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
