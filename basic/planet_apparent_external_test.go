package basic

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

type planetApparentSample struct {
	Body              string  `json:"body"`
	InputUTC          string  `json:"input_utc"`
	RightAscension    float64 `json:"right_ascension"`
	Declination       float64 `json:"declination"`
	EclipticLongitude float64 `json:"ecliptic_longitude"`
	EclipticLatitude  float64 `json:"ecliptic_latitude"`
}

func TestPlanetApparentCoordinatesMatchHorizonsBaseline(t *testing.T) {
	data, err := os.ReadFile("testdata/planet_apparent_baseline.json")
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}

	var samples []planetApparentSample
	if err := json.Unmarshal(data, &samples); err != nil {
		t.Fatalf("decode baseline: %v", err)
	}

	type apparentCase struct {
		lo  func(float64) float64
		bo  func(float64) float64
		ra  func(float64) float64
		dec func(float64) float64
	}

	cases := map[string]apparentCase{
		"mercury": {lo: MercuryApparentLo, bo: MercuryApparentBo, ra: MercuryApparentRa, dec: MercuryApparentDec},
		"venus":   {lo: VenusApparentLo, bo: VenusApparentBo, ra: VenusApparentRa, dec: VenusApparentDec},
		"mars":    {lo: MarsApparentLo, bo: MarsApparentBo, ra: MarsApparentRa, dec: MarsApparentDec},
		"jupiter": {lo: JupiterApparentLo, bo: JupiterApparentBo, ra: JupiterApparentRa, dec: JupiterApparentDec},
		"saturn":  {lo: SaturnApparentLo, bo: SaturnApparentBo, ra: SaturnApparentRa, dec: SaturnApparentDec},
		"uranus":  {lo: UranusApparentLo, bo: UranusApparentBo, ra: UranusApparentRa, dec: UranusApparentDec},
		"neptune": {lo: NeptuneApparentLo, bo: NeptuneApparentBo, ra: NeptuneApparentRa, dec: NeptuneApparentDec},
	}

	seen := make(map[string]bool, len(cases))
	for _, sample := range samples {
		tc, ok := cases[sample.Body]
		if !ok {
			t.Fatalf("unknown body %q", sample.Body)
		}
		if seen[sample.Body] {
			t.Fatalf("duplicate body %q in apparent baseline", sample.Body)
		}
		seen[sample.Body] = true

		date, err := time.Parse(time.RFC3339, sample.InputUTC)
		if err != nil {
			t.Fatalf("parse sample time %q: %v", sample.InputUTC, err)
		}
		jd := TD2UT(Date2JDE(date.UTC()), true)
		prefix := sample.Body + "." + sample.InputUTC

		assertPlanetApparentAngleClose(t, prefix+".RightAscension", tc.ra(jd), sample.RightAscension, 0.001)
		assertPlanetPhaseClose(t, prefix+".Declination", tc.dec(jd), sample.Declination, 0.001)
		assertPlanetApparentAngleClose(t, prefix+".EclipticLongitude", tc.lo(jd), sample.EclipticLongitude, 0.001)
		assertPlanetPhaseClose(t, prefix+".EclipticLatitude", tc.bo(jd), sample.EclipticLatitude, 0.001)
	}

	for body := range cases {
		if !seen[body] {
			t.Fatalf("missing body %q in apparent baseline", body)
		}
	}
}

func assertPlanetApparentAngleClose(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if diff := angleDiffAbs(got, want); diff > tolerance {
		t.Fatalf("%s mismatch: got %.12f want %.12f diff %.12f tolerance %.12f", name, got, want, diff, tolerance)
	}
}
