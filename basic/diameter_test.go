package basic

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"
)

type angularDiameterSample struct {
	InputUTC string             `json:"input_utc"`
	Values   map[string]float64 `json:"values"`
}

func TestAngularDiametersMatchHorizonsBaseline(t *testing.T) {
	// Baseline is generated from JPL Horizons by scripts/generate_angular_diameter_baseline.sh.
	data, err := os.ReadFile("testdata/angular_diameter_baseline.json")
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}

	var samples []angularDiameterSample
	if err := json.Unmarshal(data, &samples); err != nil {
		t.Fatalf("decode baseline: %v", err)
	}

	type diameterCase struct {
		name            string
		diameter        func(float64) float64
		semidiameter    func(float64) float64
		baselineKey     string
		toleranceArcsec float64
	}

	cases := []diameterCase{
		{name: "Sun", diameter: SunDiameter, semidiameter: SunSemidiameter, baselineKey: "sun", toleranceArcsec: 0.01},
		{name: "Moon", diameter: MoonDiameter, semidiameter: MoonSemidiameter, baselineKey: "moon", toleranceArcsec: 0.2},
		{name: "Mercury", diameter: MercuryDiameter, semidiameter: MercurySemidiameter, baselineKey: "mercury", toleranceArcsec: 0.01},
		{name: "Venus", diameter: VenusDiameter, semidiameter: VenusSemidiameter, baselineKey: "venus", toleranceArcsec: 0.01},
		{name: "Mars", diameter: MarsDiameter, semidiameter: MarsSemidiameter, baselineKey: "mars", toleranceArcsec: 0.01},
		{name: "Jupiter", diameter: JupiterDiameter, semidiameter: JupiterSemidiameter, baselineKey: "jupiter", toleranceArcsec: 0.01},
		{name: "Saturn", diameter: SaturnDiameter, semidiameter: SaturnSemidiameter, baselineKey: "saturn", toleranceArcsec: 0.01},
		{name: "Uranus", diameter: UranusDiameter, semidiameter: UranusSemidiameter, baselineKey: "uranus", toleranceArcsec: 0.01},
		{name: "Neptune", diameter: NeptuneDiameter, semidiameter: NeptuneSemidiameter, baselineKey: "neptune", toleranceArcsec: 0.01},
	}

	maxDiffs := make(map[string]float64, len(cases))
	for _, sample := range samples {
		date, err := time.Parse(time.RFC3339, sample.InputUTC)
		if err != nil {
			t.Fatalf("parse sample time %q: %v", sample.InputUTC, err)
		}
		jd := TD2UT(Date2JDE(date.UTC()), true)
		for _, tc := range cases {
			want := sample.Values[tc.baselineKey]
			got := tc.diameter(jd)
			diff := math.Abs(got - want)
			if diff > maxDiffs[tc.name] {
				maxDiffs[tc.name] = diff
			}
			if math.Abs(got-want) > tc.toleranceArcsec {
				t.Fatalf("%s diameter mismatch at %s: got %.9f want %.9f tolerance %.9f", tc.name, sample.InputUTC, got, want, tc.toleranceArcsec)
			}
			semi := tc.semidiameter(jd)
			if math.Abs(got-2*semi) > 1e-12 {
				t.Fatalf("%s diameter/semidiameter mismatch at %s: diameter %.12f semidiameter %.12f", tc.name, sample.InputUTC, got, semi)
			}
		}
	}

	for _, tc := range cases {
		t.Logf("%s diameter max diff: %.6f arcsec", tc.name, maxDiffs[tc.name])
	}
}

func TestAngularDiameterNFullMatchesDefault(t *testing.T) {
	jd := TD2UT(Date2JDE(time.Date(2026, 4, 28, 9, 30, 45, 0, time.UTC)), true)

	cases := []struct {
		name          string
		diameter      func(float64) float64
		diameterN     func(float64, int) float64
		semidiameter  func(float64) float64
		semidiameterN func(float64, int) float64
	}{
		{"Sun", SunDiameter, SunDiameterN, SunSemidiameter, SunSemidiameterN},
		{"Moon", MoonDiameter, MoonDiameterN, MoonSemidiameter, MoonSemidiameterN},
		{"Mercury", MercuryDiameter, MercuryDiameterN, MercurySemidiameter, MercurySemidiameterN},
		{"Venus", VenusDiameter, VenusDiameterN, VenusSemidiameter, VenusSemidiameterN},
		{"Mars", MarsDiameter, MarsDiameterN, MarsSemidiameter, MarsSemidiameterN},
		{"Jupiter", JupiterDiameter, JupiterDiameterN, JupiterSemidiameter, JupiterSemidiameterN},
		{"Saturn", SaturnDiameter, SaturnDiameterN, SaturnSemidiameter, SaturnSemidiameterN},
		{"Uranus", UranusDiameter, UranusDiameterN, UranusSemidiameter, UranusSemidiameterN},
		{"Neptune", NeptuneDiameter, NeptuneDiameterN, NeptuneSemidiameter, NeptuneSemidiameterN},
	}

	for _, tc := range cases {
		assertSameFloat(t, tc.name+".Diameter", tc.diameter(jd), tc.diameterN(jd, -1))
		assertSameFloat(t, tc.name+".Semidiameter", tc.semidiameter(jd), tc.semidiameterN(jd, -1))
	}
}
