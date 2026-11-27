package satellite

import (
	"github.com/starainrt/astro/internal/ephem"
	"encoding/json"
	"errors"
	"math"
	"os"
	"testing"
	"time"
)

type siteFixture struct {
	Groups []struct {
		Label            string
		Elements         Elements
		Source           string `json:"look_source"`
		Look             [][]json.RawMessage
		Transitions      [][]json.RawMessage
		TransitionSource string `json:"transition_source"`
	}
}

func readSiteFixtures(t *testing.T) siteFixture {
	t.Helper()
	b, err := os.ReadFile("testdata/skyfield_look.json")
	if err != nil {
		t.Fatal(err)
	}
	var data siteFixture
	if err = json.Unmarshal(b, &data); err != nil {
		t.Fatal(err)
	}
	return data
}

func TestLookExternal(t *testing.T) {
	var maximum [7]float64
	count := 0
	for _, g := range readSiteFixtures(t).Groups {
		if g.Source == "" {
			t.Fatal("missing provenance")
		}
		for _, r := range g.Look {
			if len(r) != 10 {
				t.Fatalf("invalid row %v", r)
			}
			var when time.Time
			var site [3]float64
			var want [7]float64
			var visible bool
			if err := json.Unmarshal(r[0], &when); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(r[1], &site); err != nil {
				t.Fatal(err)
			}
			for i := range want {
				if err := json.Unmarshal(r[i+2], &want[i]); err != nil {
					t.Fatal(err)
				}
			}
			if err := json.Unmarshal(r[9], &visible); err != nil {
				t.Fatal(err)
			}
			got, err := LookAt(g.Elements, Observer{Longitude: site[0], Latitude: site[1], Height: site[2]}, when, ObserveOptions{NoRefraction: true})
			if err != nil {
				t.Fatal(err)
			}
			actual := [7]float64{got.Azimuth, got.Elevation, got.RangeKM, got.RangeRateKMS, got.RA, got.Dec, got.PhaseAngle}
			limits := [7]float64{0.1 / 3600, 0.1 / 3600, 1e-4, 1e-6, 1.0 / 3600, 1.0 / 3600, 1.0 / 3600}
			for i := range actual {
				diff := actual[i] - want[i]
				if i == 0 || i == 4 {
					diff = math.Remainder(diff, 360)
				}
				diff = math.Abs(diff)
				maximum[i] = math.Max(maximum[i], diff)
				if math.IsNaN(diff) || diff > limits[i] {
					t.Fatalf("%s %s site=%v field=%d got=%.15g want=%.15g difference=%.12g limit=%.12g", g.Label, when, site, i, actual[i], want[i], diff, limits[i])
				}
			}
			if got.Time != when || got.SunCenterVisible != visible || !math.IsNaN(got.Magnitude) {
				t.Fatalf("%s invalid output %+v visible=%v", g.Label, got, visible)
			}
			count++
		}
	}
	if count != 180 {
		t.Fatalf("count=%d", count)
	}
	t.Logf("looks=%d max azimuth=%.12g arcsec elevation=%.12g arcsec range=%.12g km rate=%.12g km/s RA=%.12g arcsec Dec=%.12g arcsec phase=%.12g arcsec", count, maximum[0]*3600, maximum[1]*3600, maximum[2], maximum[3], maximum[4]*3600, maximum[5]*3600, maximum[6]*3600)
}

func equalLookBits(a, b Look) bool {
	if a.Time != b.Time || a.Illumination != b.Illumination || a.SunCenterVisible != b.SunCenterVisible {
		return false
	}
	left := []float64{a.Azimuth, a.Elevation, a.RangeKM, a.RangeRateKMS, a.RA, a.Dec, a.PhaseAngle, a.Magnitude}
	right := []float64{b.Azimuth, b.Elevation, b.RangeKM, b.RangeRateKMS, b.RA, b.Dec, b.PhaseAngle, b.Magnitude}
	for i := range left {
		if math.Float64bits(left[i]) != math.Float64bits(right[i]) {
			return false
		}
	}
	return true
}

func TestLookInstanceConsistency(t *testing.T) {
	previous := ephem.Active()
	defer ephem.SetActive(previous)
	cases := readLookFixtures(t).Groups
	count := 0
	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		ephem.SetActive(model)
		explicit := NewWithInstance(ephem.NewModelInstance(model))
		implicit := NewWithInstance(nil)
		for _, g := range cases {
			orbit, orbitErr := newSatelliteOrbit(g.Elements)
			if orbitErr != nil {
				t.Fatal(orbitErr)
			}
			for _, row := range g.Ground {
				site := Observer{Longitude: 121.5, Latitude: 31.2, Height: 1234}
				opt := ObserveOptions{NoRefraction: true}
				a, err := LookAt(g.Elements, site, row.Time, opt)
				if err != nil {
					t.Fatal(err)
				}
				b, err := explicit.LookAt(g.Elements, site, row.Time, opt)
				if err != nil {
					t.Fatal(err)
				}
				c, err := implicit.LookAt(g.Elements, site, row.Time, opt)
				if err != nil {
					t.Fatal(err)
				}
				position, positionErr := observerPosition(site)
				reused, reusedErr := explicit.lookAtOrbit(orbit, site, position, row.Time, opt)
				if positionErr != nil || reusedErr != nil || !equalLookBits(b, reused) {
					t.Fatalf("reused initialization changed observation: %v %v", positionErr, reusedErr)
				}
				if !equalLookBits(a, b) || !equalLookBits(a, c) {
					t.Fatalf("model=%v %s default=%+v explicit=%+v nil=%+v", model, g.Label, a, b, c)
				}
				other := ephem.Legacy
				if model == other {
					other = ephem.Semianalytic
				}
				ephem.SetActive(other)
				isolated, err := explicit.LookAt(g.Elements, site, row.Time, opt)
				ephem.SetActive(model)
				if err != nil || !equalLookBits(b, isolated) {
					t.Fatalf("explicit instance follows global model %v", err)
				}
				count++
			}
		}
	}
	t.Logf("two models, %d observations: default / explicit / nil-instance and global-model isolation match bit-for-bit", count)
}

func TestLookRefractionSynthetic(t *testing.T) {
	e := readLookFixtures(t).Groups[0].Elements
	when, err := time.Parse(time.RFC3339Nano, "2026-12-13T21:59:13.395264Z")
	if err != nil {
		t.Fatal(err)
	}
	site := Observer{Longitude: 123.4, Latitude: 51.5, Height: 2345}
	raw, err := LookAt(e, site, when, ObserveOptions{NoRefraction: true})
	if err != nil {
		t.Fatal(err)
	}
	if raw.Elevation < 0 || raw.Elevation > 60 {
		t.Fatalf("unusable synthetic geometry %+v", raw)
	}
	for _, tc := range []struct {
		opt     ObserveOptions
		p, temp float64
	}{
		{ObserveOptions{}, 1010, 10},
		{ObserveOptions{PressureHPa: 1010, TemperatureC: 10}, 1010, 10},
		{ObserveOptions{PressureHPa: 850, TemperatureC: -5}, 850, -5},
	} {
		got, err := LookAt(e, site, when, tc.opt)
		if err != nil {
			t.Fatal(err)
		}
		correction := (tc.p / 1010) * (283.15 / (273.15 + tc.temp)) * 1.02 / math.Tan((raw.Elevation+10.3/(raw.Elevation+5.11))*math.Pi/180) / 60
		if math.Abs(got.Elevation-raw.Elevation-correction) > 1e-14 {
			t.Fatalf("refraction got=%.15g want=%.15g", got.Elevation-raw.Elevation, correction)
		}
		got.Elevation = raw.Elevation
		if !equalLookBits(got, raw) {
			t.Fatal("refraction changed other fields")
		}
	}
	local := when.In(time.FixedZone("local", -5*3600))
	shifted, err := LookAt(e, site, local, ObserveOptions{NoRefraction: true})
	if err != nil {
		t.Fatal(err)
	}
	if shifted.Time != local {
		t.Fatal("lost time zone")
	}
	shifted.Time = when
	if !equalLookBits(shifted, raw) {
		t.Fatal("time zone changed geometry")
	}
}

func TestLookInvalidSynthetic(t *testing.T) {
	e := readLookFixtures(t).Groups[0].Elements
	for _, site := range []Observer{
		{Latitude: 90.1}, {Latitude: -90.1}, {Longitude: math.NaN()},
		{Latitude: math.Inf(1)}, {Height: math.Inf(-1)},
	} {
		got, err := LookAt(e, site, e.Epoch, ObserveOptions{})
		if !errors.Is(err, ErrInvalidInput) || got != (Look{}) {
			t.Fatalf("accepted observer %+v: %+v err=%v", site, got, err)
		}
	}
	invalid := math.Inf(1)
	for _, opt := range []ObserveOptions{
		{PressureHPa: math.NaN()}, {TemperatureC: math.NaN()},
		{PressureHPa: -1}, {PressureHPa: 1000, TemperatureC: -273.15},
		{StandardMagnitude: &invalid},
	} {
		got, err := LookAt(e, Observer{}, e.Epoch, opt)
		if !errors.Is(err, ErrInvalidInput) || got != (Look{}) {
			t.Fatalf("accepted options %+v: %+v err=%v", opt, got, err)
		}
	}
	for _, tc := range []struct {
		elements Elements
		when     time.Time
		want     error
	}{
		{Elements{}, e.Epoch, ErrInvalidElements}, {e, time.Time{}, ErrInvalidInput}, {e, time.Date(2400, 1, 1, 0, 0, 0, 0, time.UTC), ErrInvalidInput},
	} {
		got, err := LookAt(tc.elements, Observer{}, tc.when, ObserveOptions{})
		if !errors.Is(err, tc.want) || got != (Look{}) {
			t.Fatalf("invalid propagation accepted %+v %v", got, err)
		}
	}
	got, err := NewWithInstance(ephem.NewInstance(nil)).LookAt(e, Observer{}, e.Epoch, ObserveOptions{})
	if !errors.Is(err, ErrInvalidInput) || got != (Look{}) {
		t.Fatalf("invalid engine accepted %+v %v", got, err)
	}
}

func TestLookMagnitudeConsistency(t *testing.T) {
	previous := ephem.Active()
	defer ephem.SetActive(previous)
	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		ephem.SetActive(model)
		module := NewWithInstance(ephem.NewModelInstance(model))
		for _, g := range readLookFixtures(t).Groups {
			for _, row := range g.Ground {
				opt := ObserveOptions{NoRefraction: true, StandardMagnitude: floatPointer(-1.3)}
				a, err := LookAt(g.Elements, Observer{Latitude: 40}, row.Time, opt)
				if err != nil {
					t.Fatal(err)
				}
				b, err := module.LookAt(g.Elements, Observer{Latitude: 40}, row.Time, opt)
				if err != nil || !equalLookBits(a, b) {
					t.Fatalf("magnitude instance differs model=%v err=%v", model, err)
				}
				c, err := LookAt(g.Elements, Observer{Latitude: 40}, row.Time, ObserveOptions{NoRefraction: true})
				if err != nil {
					t.Fatal(err)
				}
				if a.Illumination == IlluminationUmbra {
					if !math.IsNaN(a.Magnitude) {
						t.Fatal("umbra must have NaN magnitude")
					}
				} else if !finite(a.Magnitude) {
					t.Fatalf("unlit magnitude %g", a.Magnitude)
				}
				a.Magnitude = c.Magnitude
				if !equalLookBits(a, c) {
					t.Fatal("magnitude option changed geometry")
				}
			}
		}
	}
}
