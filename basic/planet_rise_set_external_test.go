package basic

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

type planetRiseSetBaselineSample struct {
	Body       string  `json:"body"`
	Site       string  `json:"site"`
	InputUTC   string  `json:"input_utc"`
	Longitude  float64 `json:"longitude"`
	Latitude   float64 `json:"latitude"`
	RiseUTC    string  `json:"rise_utc"`
	TransitUTC string  `json:"transit_utc"`
	SetUTC     string  `json:"set_utc"`
}

func TestPlanetRiseSetMatchesHorizonsBaseline(t *testing.T) {
	data, err := os.ReadFile("testdata/planet_rise_set_baseline.json")
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}

	var samples []planetRiseSetBaselineSample
	if err := json.Unmarshal(data, &samples); err != nil {
		t.Fatalf("decode baseline: %v", err)
	}

	type observationCase struct {
		rise    func(float64, float64, float64, float64, float64, float64) (float64, error)
		transit func(float64, float64, float64) float64
		set     func(float64, float64, float64, float64, float64, float64) (float64, error)
	}

	cases := map[string]observationCase{
		"mercury": {rise: MercuryRiseTime, transit: MercuryCulminationTime, set: MercurySetTime},
		"venus":   {rise: VenusRiseTime, transit: VenusCulminationTime, set: VenusSetTime},
		"mars":    {rise: MarsRiseTime, transit: MarsCulminationTime, set: MarsSetTime},
		"jupiter": {rise: JupiterRiseTime, transit: JupiterCulminationTime, set: JupiterSetTime},
		"saturn":  {rise: SaturnRiseTime, transit: SaturnCulminationTime, set: SaturnSetTime},
		"uranus":  {rise: UranusRiseTime, transit: UranusCulminationTime, set: UranusSetTime},
		"neptune": {rise: NeptuneRiseTime, transit: NeptuneCulminationTime, set: NeptuneSetTime},
	}

	const tolerance = 2 * time.Minute
	var maxRiseDiff time.Duration
	var maxTransitDiff time.Duration
	var maxSetDiff time.Duration

	for _, sample := range samples {
		tc, ok := cases[sample.Body]
		if !ok {
			t.Fatalf("unknown body %q", sample.Body)
		}

		inputTime, err := time.Parse(time.RFC3339, sample.InputUTC)
		if err != nil {
			t.Fatalf("parse input time %q: %v", sample.InputUTC, err)
		}
		jd := Date2JDE(inputTime.UTC())

		riseJD, err := tc.rise(jd, sample.Longitude, sample.Latitude, 0, 1, 0)
		if err != nil {
			t.Fatalf("%s %s rise error: %v", sample.Body, sample.Site, err)
		}
		riseDiff := assertEventTimeClose(t, sample.Body+"."+sample.Site+".rise", riseJD, sample.RiseUTC, tolerance)
		if riseDiff > maxRiseDiff {
			maxRiseDiff = riseDiff
		}

		transitJD := tc.transit(jd, sample.Longitude, 0)
		transitDiff := assertEventTimeClose(t, sample.Body+"."+sample.Site+".transit", transitJD, sample.TransitUTC, tolerance)
		if transitDiff > maxTransitDiff {
			maxTransitDiff = transitDiff
		}

		setJD, err := tc.set(jd, sample.Longitude, sample.Latitude, 0, 1, 0)
		if err != nil {
			t.Fatalf("%s %s set error: %v", sample.Body, sample.Site, err)
		}
		setDiff := assertEventTimeClose(t, sample.Body+"."+sample.Site+".set", setJD, sample.SetUTC, tolerance)
		if setDiff > maxSetDiff {
			maxSetDiff = setDiff
		}
	}

	t.Logf("planet rise/set max diff: rise=%v transit=%v set=%v", maxRiseDiff, maxTransitDiff, maxSetDiff)
}

func assertEventTimeClose(t *testing.T, name string, gotJD float64, wantUTC string, tolerance time.Duration) time.Duration {
	t.Helper()

	wantTime, err := time.Parse(time.RFC3339, wantUTC)
	if err != nil {
		t.Fatalf("parse %s baseline time %q: %v", name, wantUTC, err)
	}

	gotTime := JDE2DateByZone(gotJD, time.UTC, false)
	diff := gotTime.Sub(wantTime)
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Fatalf("%s mismatch: got %s want %s tolerance %v", name, gotTime.Format(time.RFC3339), wantUTC, tolerance)
	}
	return diff
}
