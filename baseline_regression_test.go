package astro_test

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/planet"
	"github.com/starainrt/astro/sun"
)

type baselinePlanetSnapshot struct {
	Name    string `json:"name"`
	XT      int    `json:"xt"`
	LonBits uint64 `json:"lon_bits"`
	LatBits uint64 `json:"lat_bits"`
	RadBits uint64 `json:"rad_bits"`
}

type baselineMoonSnapshot struct {
	LonBits uint64 `json:"lon_bits"`
	LatBits uint64 `json:"lat_bits"`
	DisBits uint64 `json:"dis_bits"`
}

type baselineSample struct {
	UTC     string                   `json:"utc"`
	TTJD    float64                  `json:"tt_jd"`
	Planets []baselinePlanetSnapshot `json:"planets"`
	Moon    baselineMoonSnapshot     `json:"moon"`
}

func loadBaselineSamples(t *testing.T) []baselineSample {
	t.Helper()

	data, err := os.ReadFile("testdata/planet_moon_baseline.json")
	if err != nil {
		t.Fatal(err)
	}

	var samples []baselineSample
	if err := json.Unmarshal(data, &samples); err != nil {
		t.Fatal(err)
	}
	if len(samples) == 0 {
		t.Fatal("empty baseline samples")
	}
	return samples
}

func TestPlanetMoonBaselineRegression(t *testing.T) {
	samples := loadBaselineSamples(t)
	for _, sample := range samples {
		for _, body := range sample.Planets {
			gotLon := planet.WherePlanet(body.XT, 0, sample.TTJD)
			if math.Float64bits(gotLon) != body.LonBits {
				t.Fatalf("%s lon regression at %s", body.Name, sample.UTC)
			}
			gotLonN := planet.WherePlanetN(body.XT, 0, sample.TTJD, -1)
			if math.Float64bits(gotLonN) != body.LonBits {
				t.Fatalf("%s lon full-n regression at %s", body.Name, sample.UTC)
			}

			gotLat := planet.WherePlanet(body.XT, 1, sample.TTJD)
			if math.Float64bits(gotLat) != body.LatBits {
				t.Fatalf("%s lat regression at %s", body.Name, sample.UTC)
			}
			gotLatN := planet.WherePlanetN(body.XT, 1, sample.TTJD, -1)
			if math.Float64bits(gotLatN) != body.LatBits {
				t.Fatalf("%s lat full-n regression at %s", body.Name, sample.UTC)
			}

			gotRad := planet.WherePlanet(body.XT, 2, sample.TTJD)
			if math.Float64bits(gotRad) != body.RadBits {
				t.Fatalf("%s rad regression at %s", body.Name, sample.UTC)
			}
			gotRadN := planet.WherePlanetN(body.XT, 2, sample.TTJD, -1)
			if math.Float64bits(gotRadN) != body.RadBits {
				t.Fatalf("%s rad full-n regression at %s", body.Name, sample.UTC)
			}
		}

		if math.Float64bits(basic.HMoonTrueLo(sample.TTJD)) != sample.Moon.LonBits {
			t.Fatalf("moon lon regression at %s", sample.UTC)
		}
		if math.Float64bits(basic.HMoonTrueLoN(sample.TTJD, -1)) != sample.Moon.LonBits {
			t.Fatalf("moon lon full-n regression at %s", sample.UTC)
		}
		if math.Float64bits(basic.HMoonTrueBo(sample.TTJD)) != sample.Moon.LatBits {
			t.Fatalf("moon lat regression at %s", sample.UTC)
		}
		if math.Float64bits(basic.HMoonTrueBoN(sample.TTJD, -1)) != sample.Moon.LatBits {
			t.Fatalf("moon lat full-n regression at %s", sample.UTC)
		}
		if math.Float64bits(basic.HMoonAway(sample.TTJD)) != sample.Moon.DisBits {
			t.Fatalf("moon distance regression at %s", sample.UTC)
		}
		if math.Float64bits(basic.HMoonAwayN(sample.TTJD, -1)) != sample.Moon.DisBits {
			t.Fatalf("moon distance full-n regression at %s", sample.UTC)
		}
	}
}

func TestPublicTruncationFullMatchesDefault(t *testing.T) {
	date := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.UTC)

	if math.Float64bits(sun.TrueLo(date)) != math.Float64bits(sun.TrueLoN(date, -1)) {
		t.Fatal("sun.TrueLoN(-1) should match default")
	}
	if math.Float64bits(sun.TrueBo(date)) != math.Float64bits(sun.TrueBoN(date, -1)) {
		t.Fatal("sun.TrueBoN(-1) should match default")
	}
	if math.Float64bits(moon.TrueLo(date)) != math.Float64bits(moon.TrueLoN(date, -1)) {
		t.Fatal("moon.TrueLoN(-1) should match default")
	}
	if math.Float64bits(moon.TrueBo(date)) != math.Float64bits(moon.TrueBoN(date, -1)) {
		t.Fatal("moon.TrueBoN(-1) should match default")
	}
}
