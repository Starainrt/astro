package basic

import (
	"encoding/json"
	"math"
	"os"
	"testing"
	"time"
)

type planetPhysicalSample struct {
	Body                   string  `json:"body"`
	InputUTC               string  `json:"input_utc"`
	SubEarthLongitude      float64 `json:"sub_earth_longitude"`
	SubEarthLatitude       float64 `json:"sub_earth_latitude"`
	SubSolarLongitude      float64 `json:"sub_solar_longitude"`
	SubSolarLatitude       float64 `json:"sub_solar_latitude"`
	NorthPolePositionAngle float64 `json:"north_pole_position_angle"`
}

func TestPlanetPhysicalMatchesHorizonsBaseline(t *testing.T) {
	data, err := os.ReadFile("testdata/planet_physical_baseline.json")
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}

	var samples []planetPhysicalSample
	if err := json.Unmarshal(data, &samples); err != nil {
		t.Fatalf("decode baseline: %v", err)
	}

	cases := map[string]func(float64) PlanetPhysicalInfo{
		"mercury": MercuryPhysical,
		"venus":   VenusPhysical,
		"mars":    MarsPhysical,
		"jupiter": JupiterPhysical,
		"saturn":  SaturnPhysical,
		"uranus":  UranusPhysical,
		"neptune": NeptunePhysical,
	}

	for _, sample := range samples {
		physical := cases[sample.Body]
		if physical == nil {
			t.Fatalf("missing body case %q", sample.Body)
		}

		date, err := time.Parse(time.RFC3339, sample.InputUTC)
		if err != nil {
			t.Fatalf("parse sample time %q: %v", sample.InputUTC, err)
		}
		jd := TD2UT(Date2JDE(date.UTC()), true)
		got := physical(jd)

		assertPlanetPhaseClose(t, sample.Body+"."+sample.InputUTC+".SubEarthLongitude", got.SubEarthLongitude, sample.SubEarthLongitude, 0.02)
		assertPlanetPhaseClose(t, sample.Body+"."+sample.InputUTC+".SubEarthLatitude", got.SubEarthLatitude, sample.SubEarthLatitude, 0.02)
		assertPlanetPhaseClose(t, sample.Body+"."+sample.InputUTC+".SubSolarLongitude", got.SubSolarLongitude, sample.SubSolarLongitude, 0.02)
		assertPlanetPhaseClose(t, sample.Body+"."+sample.InputUTC+".SubSolarLatitude", got.SubSolarLatitude, sample.SubSolarLatitude, 0.02)
		assertPlanetPhaseClose(t, sample.Body+"."+sample.InputUTC+".NorthPolePositionAngle", got.NorthPolePositionAngle, sample.NorthPolePositionAngle, 0.02)
	}
}

func TestPlanetPhysicalNFullMatchesDefault(t *testing.T) {
	jd := TD2UT(Date2JDE(time.Date(2026, 4, 28, 9, 30, 45, 0, time.UTC)), true)

	cases := []struct {
		name      string
		physical  func(float64) PlanetPhysicalInfo
		physicalN func(float64, int) PlanetPhysicalInfo
	}{
		{"Mercury", MercuryPhysical, MercuryPhysicalN},
		{"Venus", VenusPhysical, VenusPhysicalN},
		{"Mars", MarsPhysical, MarsPhysicalN},
		{"Jupiter", JupiterPhysical, JupiterPhysicalN},
		{"Saturn", SaturnPhysical, SaturnPhysicalN},
		{"Uranus", UranusPhysical, UranusPhysicalN},
		{"Neptune", NeptunePhysical, NeptunePhysicalN},
	}

	for _, tc := range cases {
		got := tc.physical(jd)
		gotN := tc.physicalN(jd, -1)
		assertSameFloat(t, tc.name+".SubEarthLongitude", got.SubEarthLongitude, gotN.SubEarthLongitude)
		assertSameFloat(t, tc.name+".SubEarthLatitude", got.SubEarthLatitude, gotN.SubEarthLatitude)
		assertSameFloat(t, tc.name+".SubSolarLongitude", got.SubSolarLongitude, gotN.SubSolarLongitude)
		assertSameFloat(t, tc.name+".SubSolarLatitude", got.SubSolarLatitude, gotN.SubSolarLatitude)
		assertSameFloat(t, tc.name+".NorthPolePositionAngle", got.NorthPolePositionAngle, gotN.NorthPolePositionAngle)
	}
}

func TestPlanetPhysicalSampleSweepFiniteAndInRange(t *testing.T) {
	dates := []time.Time{
		time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1969, 7, 20, 20, 17, 40, 0, time.UTC),
		time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 28, 9, 30, 45, 0, time.UTC),
		time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	cases := []struct {
		name     string
		physical func(float64) PlanetPhysicalInfo
	}{
		{"Mercury", MercuryPhysical},
		{"Venus", VenusPhysical},
		{"Mars", MarsPhysical},
		{"Jupiter", JupiterPhysical},
		{"Saturn", SaturnPhysical},
		{"Uranus", UranusPhysical},
		{"Neptune", NeptunePhysical},
	}

	for _, date := range dates {
		jd := TD2UT(Date2JDE(date.UTC()), true)
		for _, tc := range cases {
			info := tc.physical(jd)
			prefix := tc.name + "." + date.Format(time.RFC3339)
			assertFiniteRange(t, prefix+".SubEarthLongitude", info.SubEarthLongitude, 0, 360, true)
			assertFiniteRange(t, prefix+".SubEarthLatitude", info.SubEarthLatitude, -90, 90, false)
			assertFiniteRange(t, prefix+".SubSolarLongitude", info.SubSolarLongitude, 0, 360, true)
			assertFiniteRange(t, prefix+".SubSolarLatitude", info.SubSolarLatitude, -90, 90, false)
			assertFiniteRange(t, prefix+".NorthPolePositionAngle", info.NorthPolePositionAngle, 0, 360, true)
		}
	}
}

func assertFiniteRange(t *testing.T, name string, got, min, max float64, upperExclusive bool) {
	t.Helper()
	if math.IsNaN(got) || math.IsInf(got, 0) {
		t.Fatalf("%s is not finite: %.18f", name, got)
	}
	if upperExclusive {
		if got < min || got >= max {
			t.Fatalf("%s out of range: %.18f not in [%.18f, %.18f)", name, got, min, max)
		}
		return
	}
	if got < min || got > max {
		t.Fatalf("%s out of range: %.18f not in [%.18f, %.18f]", name, got, min, max)
	}
}
