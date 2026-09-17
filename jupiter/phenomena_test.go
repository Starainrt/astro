package jupiter

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

// 基线 shadow 值按 TDB/TT 逆行历元（UTC 观测时刻换算 ΔT 后再减光行时）标定；
// 按 UTC 瞬时求值会让整条曲线平移 ΔT（Io 约 0.46″）并被本容差拒绝。
const galileanShadowToleranceArcsec = 0.15

type galileanPhenomenaSample struct {
	UTC       string                                   `json:"utc"`
	Phenomena map[string]galileanPhenomenonExpectation `json:"phenomena"`
}

type galileanPhenomenonExpectation struct {
	Transit       bool     `json:"transit"`
	Occultation   bool     `json:"occultation"`
	Eclipse       bool     `json:"eclipse"`
	ShadowTransit bool     `json:"shadow_transit"`
	ShadowXArcsec *float64 `json:"shadow_x_arcsec,omitempty"`
	ShadowYArcsec *float64 `json:"shadow_y_arcsec,omitempty"`
}

func TestGalileanPhenomenaAgainstHorizonsBaseline(t *testing.T) {
	samples := loadGalileanPhenomenaBaseline(t)
	maxShadowX := 0.0
	maxShadowY := 0.0
	for _, sample := range samples {
		date, err := time.Parse(time.RFC3339, sample.UTC)
		if err != nil {
			t.Fatalf("parse %s: %v", sample.UTC, err)
		}
		got := SatellitePhenomena(date)
		for name, want := range sample.Phenomena {
			phenomenon := selectGalileanPhenomenon(got, name)
			if phenomenon.Transit != want.Transit {
				t.Fatalf("%s transit mismatch at %s: got %v want %v", name, sample.UTC, phenomenon.Transit, want.Transit)
			}
			if phenomenon.Occultation != want.Occultation {
				t.Fatalf("%s occultation mismatch at %s: got %v want %v", name, sample.UTC, phenomenon.Occultation, want.Occultation)
			}
			if phenomenon.Eclipse != want.Eclipse {
				t.Fatalf("%s eclipse mismatch at %s: got %v want %v", name, sample.UTC, phenomenon.Eclipse, want.Eclipse)
			}
			if phenomenon.ShadowTransit != want.ShadowTransit {
				t.Fatalf("%s shadow-transit mismatch at %s: got %v want %v", name, sample.UTC, phenomenon.ShadowTransit, want.ShadowTransit)
			}
			if !want.ShadowTransit {
				continue
			}
			if want.ShadowXArcsec == nil || want.ShadowYArcsec == nil {
				t.Fatalf("%s shadow baseline incomplete at %s", name, sample.UTC)
			}
			xDiff := math.Abs(phenomenon.ShadowOffsetXArcsec - *want.ShadowXArcsec)
			yDiff := math.Abs(phenomenon.ShadowOffsetYArcsec - *want.ShadowYArcsec)
			if xDiff > maxShadowX {
				maxShadowX = xDiff
			}
			if yDiff > maxShadowY {
				maxShadowY = yDiff
			}
			if xDiff > galileanShadowToleranceArcsec {
				t.Fatalf("%s shadow X mismatch at %s: got %.6f want %.6f", name, sample.UTC, phenomenon.ShadowOffsetXArcsec, *want.ShadowXArcsec)
			}
			if yDiff > galileanShadowToleranceArcsec {
				t.Fatalf("%s shadow Y mismatch at %s: got %.6f want %.6f", name, sample.UTC, phenomenon.ShadowOffsetYArcsec, *want.ShadowYArcsec)
			}
		}
	}
	t.Logf("galilean phenomena shadow max diff: X=%.3f arcsec Y=%.3f arcsec", maxShadowX, maxShadowY)
}

func loadGalileanPhenomenaBaseline(t *testing.T) []galileanPhenomenaSample {
	t.Helper()
	data, err := os.ReadFile("testdata/galilean_phenomena_horizons.json")
	if err != nil {
		t.Fatal(err)
	}
	var samples []galileanPhenomenaSample
	if err := json.Unmarshal(data, &samples); err != nil {
		t.Fatal(err)
	}
	if len(samples) == 0 {
		t.Fatal("empty phenomena baseline")
	}
	return samples
}

func selectGalileanPhenomenon(info GalileanPhenomenaInfo, name string) GalileanSatellitePhenomenon {
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

func TestGalileanPhenomenaUseTTFrame(t *testing.T) {
	dates := []time.Time{
		time.Date(1999, 1, 1, 2, 0, 0, 0, time.UTC),
		time.Date(2000, 1, 1, 14, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 2, 20, 0, 0, 0, time.UTC),
	}
	for _, date := range dates {
		want := basic.JupiterGalileanSatellitePhenomena(basic.TD2UT(basic.Date2JDE(date.UTC()), true))
		got := SatellitePhenomena(date)
		phenomena := []GalileanSatellitePhenomenon{got.Io, got.Europa, got.Ganymede, got.Callisto}
		for i, phenomenon := range phenomena {
			assertSameGalileanPhenomenon(t, fmt.Sprintf("date=%s satellite=%d TT frame", date, i+1), phenomenon, want[i])
		}
	}
}
