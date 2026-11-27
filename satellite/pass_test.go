package satellite

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
)

func TestPassSearchSynthetic(t *testing.T) {
	origin := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	for _, depth := range []float64{-3, -1, -0.72, -0.3, 0, 0.3, 0.72, 3} {
		sample := func(ti time.Time) (passSample, error) {
			x := ti.Sub(origin).Seconds() - 50
			return passSample{time: ti, altitude: depth/3600 - x*x/10000, rate: -x}, nil
		}
		for _, step := range []time.Duration{20 * time.Second, 2500 * time.Millisecond} {
			spans, state, err := searchPasses(origin, origin.Add(100*time.Second), 0, step, sample, false)
			if err != nil {
				t.Fatal(err)
			}
			if depth > 0 {
				if state != basic.AltitudeCrosses || len(spans) != 1 {
					t.Fatalf("depth=%g step=%v state=%v spans=%v", depth, step, state, spans)
				}
				root := math.Sqrt(depth / 3600 * 10000)
				if math.Abs(spans[0].start.Sub(origin).Seconds()-(50-root)) > 0.001 || math.Abs(spans[0].end.Sub(origin).Seconds()-(50+root)) > 0.001 {
					t.Fatalf("analytic roots depth=%g spans=%v", depth, spans)
				}
			} else if math.Abs(depth) < 1 {
				if state != basic.AltitudeTangent || len(spans) != 1 || !spans[0].start.Equal(spans[0].end) {
					t.Fatalf("tangent depth=%g state=%v spans=%v", depth, state, spans)
				}
			} else if state != basic.AltitudeAlwaysBelow || len(spans) != 0 {
				t.Fatalf("below depth=%g state=%v spans=%v", depth, state, spans)
			}
		}
	}
}

func TestPassDenseConsistency(t *testing.T) {
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	total := 0
	for _, g := range readLookFixtures(t).Groups {
		for _, lat := range []float64{0, 40, -70} {
			e := g.Elements
			o, err := newSatelliteOrbit(e)
			if err != nil {
				t.Fatal(err)
			}
			obs := Observer{Latitude: lat}
			position, err := observerPosition(obs)
			if err != nil {
				t.Fatal(err)
			}
			sample := func(ti time.Time) (passSample, error) { return o.passSampleAt(obs, position, ti) }
			coarse, state, err := searchPasses(start, start.Add(30*24*time.Hour), 0, passStep(e), sample, false)
			if err != nil {
				t.Fatal(err)
			}
			dense, denseState, err := searchPasses(start, start.Add(30*24*time.Hour), 0, passStep(e)/8, sample, false)
			if err != nil {
				t.Fatal(err)
			}
			if state != denseState || len(coarse) != len(dense) {
				t.Fatalf("%s lat=%g counts=%d/%d state=%v/%v", g.Label, lat, len(coarse), len(dense), state, denseState)
			}
			for i, a := range coarse {
				b := dense[i]
				if math.Abs(a.start.Sub(b.start).Seconds()) > 0.01 || math.Abs(a.end.Sub(b.end).Seconds()) > 0.01 || math.Abs(a.peakAltitude-b.peakAltitude)*3600 > 0.005 {
					t.Fatalf("%s lat=%g event=%d coarse=%+v dense=%+v", g.Label, lat, i, a, b)
				}
			}
			total += len(coarse)
			t.Logf("%s lat=%g passes=%d state=%v coarse/dense identical collection", g.Label, lat, len(coarse), state)
		}
	}
	if total != 1478 {
		t.Fatalf("total=%d want=1478", total)
	}
}

func TestPassExternal(t *testing.T) {
	const contactTolerance = 0.01
	const elevationTolerance = 0.1
	const peakLossTolerance = 0.005
	const selfcheckTolerance = 0.0002
	var fixture struct {
		Groups []struct {
			Label, Source string
			Elements      Elements
			Latitude      float64
			Start, End    time.Time
			Rows          [][]json.RawMessage
		}
	}
	b, err := os.ReadFile("testdata/skyfield_passes.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	var maxRise, maxSet, maxAltitude, maxLoss, maxCheck float64
	count := 0
	for _, g := range fixture.Groups {
		if g.Source == "" || g.Label == "" {
			t.Fatal("missing provenance")
		}
		passes, _, err := Passes(g.Elements, Observer{Latitude: g.Latitude}, g.Start, g.End, PassOptions{ObserveOptions: ObserveOptions{NoRefraction: true}})
		if err != nil || len(passes) != len(g.Rows) {
			t.Fatalf("%s passes=%d rows=%d err=%v", g.Label, len(passes), len(g.Rows), err)
		}
		for i, row := range g.Rows {
			if len(row) != 7 {
				t.Fatal("invalid row")
			}
			var rise, peak, set time.Time
			var alt, peakHeight, atPeak, check float64
			for j, dest := range []interface{}{&rise, &peak, &set, &alt, &peakHeight, &atPeak, &check} {
				if err = json.Unmarshal(row[j], dest); err != nil {
					t.Fatal(err)
				}
			}
			own := passes[i]
			dr, ds := math.Abs(own.Rise.Time.Sub(rise).Seconds()), math.Abs(own.Set.Time.Sub(set).Seconds())
			da := math.Abs(own.Culmination.Elevation-alt) * 3600
			loss := peakHeight - atPeak
			maxRise = math.Max(maxRise, dr)
			maxSet = math.Max(maxSet, ds)
			maxAltitude = math.Max(maxAltitude, da)
			maxLoss = math.Max(maxLoss, loss)
			maxCheck = math.Max(maxCheck, check)
			if !own.Culmination.Time.Equal(peak) {
				t.Fatalf("%s event=%d changed oracle query time: %s want=%s", g.Label, i, own.Culmination.Time, peak)
			}
			if dr > contactTolerance || ds > contactTolerance || da > elevationTolerance || loss > peakLossTolerance || check > selfcheckTolerance {
				t.Fatalf("%s event=%d rise=%.12g set=%.12g altitude=%.12g loss=%.12g check=%.12g", g.Label, i, dr, ds, da, loss, check)
			}
			count++
		}
	}
	t.Logf("passes=%d max rise=%.12g s set=%.12g s altitude=%.12g arcsec peak deficit=%.12g arcsec selfcheck=%.12g arcsec", count, maxRise, maxSet, maxAltitude, maxLoss, maxCheck)
}

func TestPassNextSeedsConsistency(t *testing.T) {
	e := readLookFixtures(t).Groups[0].Elements
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	obs := Observer{Latitude: 40}
	opt := PassOptions{ObserveOptions: ObserveOptions{NoRefraction: true}, MinElevation: 10}
	events, _, err := Passes(e, obs, start, start.Add(24*time.Hour), opt)
	if err != nil || len(events) < 2 {
		t.Fatalf("passes=%d err=%v", len(events), err)
	}
	p := events[0]
	for _, seed := range []time.Time{p.Rise.Time.Add(-time.Minute), p.Rise.Time.Add(p.Duration / 4), p.Culmination.Time, p.Set.Time.Add(-p.Duration / 4), p.Set.Time.Add(time.Minute)} {
		got, err := NextPass(e, obs, seed, opt)
		if err != nil {
			t.Fatal(err)
		}
		want, _, err := Passes(e, obs, seed, seed.Add(24*time.Hour), opt)
		if err != nil || len(want) == 0 || !equalPassBits(got, want[0]) {
			t.Fatalf("seed=%s got=%+v want=%+v err=%v", seed, got, want, err)
		}
		if seed.After(p.Rise.Time) && seed.Before(p.Set.Time) && (!got.StartsAbove || got.Rise.Time != seed) {
			t.Fatalf("ongoing pass not clipped at %s", seed)
		}
	}
}

func equalPassBits(a, b Pass) bool {
	return equalLookBits(a.Rise, b.Rise) && equalLookBits(a.Culmination, b.Culmination) && equalLookBits(a.Set, b.Set) && a.Duration == b.Duration && a.StartsAbove == b.StartsAbove && a.EndsAbove == b.EndsAbove && a.Kind == b.Kind && a.VisibleStart == b.VisibleStart && a.VisibleEnd == b.VisibleEnd && a.BrightestTime == b.BrightestTime && math.Float64bits(a.BrightestMagnitude) == math.Float64bits(b.BrightestMagnitude)
}

func TestPassRefractionConsistency(t *testing.T) {
	e := readLookFixtures(t).Groups[0].Elements
	obs := Observer{Latitude: 40}
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	apparent := PassOptions{MinElevation: 10}
	threshold := basic.TrueAltitude(10, 1010, 10)
	geometric := PassOptions{ObserveOptions: ObserveOptions{NoRefraction: true}, MinElevation: threshold}
	a, as, err := Passes(e, obs, start, start.Add(24*time.Hour), apparent)
	if err != nil {
		t.Fatal(err)
	}
	b, bs, err := Passes(e, obs, start, start.Add(24*time.Hour), geometric)
	if err != nil {
		t.Fatal(err)
	}
	if as != bs || len(a) != len(b) || len(a) == 0 {
		t.Fatalf("different threshold semantics %v %v %d %d", as, bs, len(a), len(b))
	}
	for i, p := range a {
		if p.Rise.Time != b[i].Rise.Time || p.Set.Time != b[i].Set.Time || p.Culmination.Time != b[i].Culmination.Time {
			t.Fatalf("event=%d refraction changed equivalent roots", i)
		}
		for _, look := range []Look{p.Rise, p.Culmination, p.Set} {
			want, err := LookAt(e, obs, look.Time, apparent.ObserveOptions)
			if err != nil || !equalLookBits(look, want) {
				t.Fatalf("output Look inconsistent %v", err)
			}
		}
	}
}

func TestPassInputSynthetic(t *testing.T) {
	e := readLookFixtures(t).Groups[0].Elements
	start := e.Epoch
	for _, interval := range [][2]time.Time{{{}, start}, {start, {}}, {start, start.Add(-time.Second)}, {start, start.Add(30*24*time.Hour + time.Nanosecond)}, {time.Date(2400, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2400, 1, 2, 0, 0, 0, 0, time.UTC)}} {
		_, _, err := Passes(e, Observer{}, interval[0], interval[1], PassOptions{})
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("invalid interval %v err=%v", interval, err)
		}
	}
	if _, _, err := Passes(Elements{}, Observer{}, start, start.Add(time.Hour), PassOptions{}); !errors.Is(err, ErrInvalidElements) {
		t.Fatalf("elements error %v", err)
	}
	for _, opt := range []PassOptions{{MinElevation: 91}, {MinElevation: math.NaN()}, {TwilightSunAltitude: floatPointer(math.Inf(1))}, {ObserveOptions: ObserveOptions{PressureHPa: -1}}} {
		if _, _, err := Passes(e, Observer{}, start, start.Add(time.Hour), opt); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("invalid options %+v err=%v", opt, err)
		}
	}
	for _, tc := range []struct {
		sample passSampler
		state  basic.AltitudeCrossingState
		n      int
	}{
		{func(ti time.Time) (passSample, error) { return passSample{time: ti, altitude: 10, rate: 1}, nil }, basic.AltitudeAlwaysAbove, 1},
		{func(ti time.Time) (passSample, error) { return passSample{time: ti, altitude: -10, rate: 1}, nil }, basic.AltitudeAlwaysBelow, 0},
	} {
		spans, state, err := searchPasses(start, start.Add(time.Hour), 0, time.Minute, tc.sample, false)
		if err != nil || state != tc.state || len(spans) != tc.n {
			t.Fatalf("constant state=%v spans=%v err=%v", state, spans, err)
		}
		if tc.n > 0 && (!spans[0].startsAbove || !spans[0].endsAbove) {
			t.Fatal("missing endpoint flags")
		}
	}
	geo := readLookFixtures(t).Groups[4].Elements
	_, state, err := Passes(geo, Observer{Longitude: readLookFixtures(t).Groups[4].Ground[0].Longitude}, start, start.Add(24*time.Hour), PassOptions{ObserveOptions: ObserveOptions{NoRefraction: true}})
	if err != nil || state != basic.AltitudeAlwaysAbove {
		t.Fatalf("stationary orbit state=%v err=%v", state, err)
	}
	_, err = NextPass(geo, Observer{Latitude: -90}, start, PassOptions{ObserveOptions: ObserveOptions{NoRefraction: true}, MinElevation: 89})
	if !errors.Is(err, ErrPassNotFound) {
		t.Fatalf("missing pass err=%v", err)
	}
}

func floatPointer(v float64) *float64 { return &v }

func TestPassInstanceConsistency(t *testing.T) {
	previous := ephem.Active()
	defer ephem.SetActive(previous)
	e := readLookFixtures(t).Groups[0].Elements
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	obs := Observer{Latitude: 40}
	opt := PassOptions{ObserveOptions: ObserveOptions{NoRefraction: true}}
	for _, model := range []ephem.Model{ephem.Legacy, ephem.Semianalytic} {
		ephem.SetActive(model)
		want, state, err := Passes(e, obs, start, start.Add(24*time.Hour), opt)
		if err != nil || len(want) == 0 {
			t.Fatalf("passes=%d err=%v", len(want), err)
		}
		for _, module := range []*Satellite{NewWithInstance(nil), NewWithInstance(ephem.NewModelInstance(model))} {
			got, gs, err := module.Passes(e, obs, start, start.Add(24*time.Hour), opt)
			if err != nil || state != gs || len(got) != len(want) {
				t.Fatalf("model=%v state=%v/%v err=%v", model, state, gs, err)
			}
			for i := range want {
				if !equalPassBits(want[i], got[i]) {
					t.Fatalf("model=%v pass=%d differs", model, i)
				}
			}
			next, err := module.NextPass(e, obs, start, opt)
			if err != nil || !equalPassBits(next, want[0]) {
				t.Fatalf("next differs model=%v err=%v", model, err)
			}
		}
	}
}

func TestPassVisibilityExternal(t *testing.T) {
	var data struct {
		Groups []struct {
			Label            string
			Elements         Elements
			Latitude         float64
			Start, End       time.Time
			VisibilitySource string `json:"visibility_source"`
			Visibility       []struct {
				Kind    PassKind
				Windows [][2]time.Time
			}
		}
	}
	b, err := os.ReadFile("testdata/skyfield_passes.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &data); err != nil {
		t.Fatal(err)
	}
	counts := [3]int{}
	windows := 0
	maximum := 0.0
	for _, g := range data.Groups {
		if g.VisibilitySource == "" {
			t.Fatal("missing provenance")
		}
		passes, _, err := Passes(g.Elements, Observer{Latitude: g.Latitude}, g.Start, g.End, PassOptions{ObserveOptions: ObserveOptions{NoRefraction: true}})
		if err != nil || len(passes) != len(g.Visibility) {
			t.Fatalf("%s count=%d/%d err=%v", g.Label, len(passes), len(g.Visibility), err)
		}
		for i, p := range passes {
			ref := g.Visibility[i]
			counts[ref.Kind]++
			if p.Kind != ref.Kind {
				t.Fatalf("%s event=%d kind=%v want=%v", g.Label, i, p.Kind, ref.Kind)
			}
			if len(ref.Windows) == 0 {
				if !p.VisibleStart.IsZero() || !p.VisibleEnd.IsZero() {
					t.Fatalf("unexpected visibility %s event=%d", g.Label, i)
				}
			} else {
				windows += len(ref.Windows)
				a, b := math.Abs(p.VisibleStart.Sub(ref.Windows[0][0]).Seconds()), math.Abs(p.VisibleEnd.Sub(ref.Windows[len(ref.Windows)-1][1]).Seconds())
				maximum = math.Max(maximum, math.Max(a, b))
				if a > .05 || b > .05 {
					t.Fatalf("%s event=%d visibility errors %g/%g s", g.Label, i, a, b)
				}
			}
			if !p.BrightestTime.IsZero() || !math.IsNaN(p.BrightestMagnitude) {
				t.Fatal("nil standard magnitude contract")
			}
		}
	}
	t.Logf("visibility counts visible/daylight/eclipsed=%v reference windows=%d max boundary error=%.12g s", counts, windows, maximum)
}

func TestPassVisibilitySynthetic(t *testing.T) {
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	for _, depth := range []float64{-.72, .3, .72, 3} {
		sample := func(ti time.Time) (passLightSample, error) {
			x := ti.Sub(start).Seconds() - 50
			return passLightSample{time: ti, night: 1, light: depth/3600 - x*x/10000, magnitude: x * x}, nil
		}
		for _, step := range []time.Duration{20 * time.Second, 2500 * time.Millisecond} {
			windows, daylight, err := passVisibility(start, start.Add(100*time.Second), step, sample)
			count := 0
			if depth > 0 {
				count = 1
			}
			if err != nil || daylight || len(windows) != count {
				t.Fatalf("depth=%g step=%v windows=%v daylight=%v err=%v", depth, step, windows, daylight, err)
			}
			if count > 0 {
				root := math.Sqrt(depth / 3600 * 10000)
				if math.Abs(windows[0].start.Sub(start).Seconds()-(50-root)) > .001 || math.Abs(windows[0].end.Sub(start).Seconds()-(50+root)) > .001 {
					t.Fatalf("analytic visibility roots depth=%g windows=%v", depth, windows)
				}
				when, mag, err := passBrightest(windows, step, sample)
				if err != nil || math.Abs(when.Sub(start).Seconds()-50) > .001 || math.Abs(mag) > 1e-6 {
					t.Fatalf("brightness when=%v mag=%g err=%v", when, mag, err)
				}
			}
		}
	}
}

func TestPassBrightnessDenseConsistency(t *testing.T) {
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	standard := 0.0
	opt := PassOptions{ObserveOptions: ObserveOptions{NoRefraction: true, StandardMagnitude: &standard}}
	visible, different := 0, 0
	maxExcess := 0.0
	for _, g := range readLookFixtures(t).Groups {
		for _, lat := range []float64{0, 40, -70} {
			obs := Observer{Latitude: lat}
			passes, _, err := Passes(g.Elements, obs, start, start.Add(48*time.Hour), opt)
			if err != nil {
				t.Fatal(err)
			}
			orbit, err := newSatelliteOrbit(g.Elements)
			if err != nil {
				t.Fatal(err)
			}
			position, err := observerPosition(obs)
			if err != nil {
				t.Fatal(err)
			}
			sample := func(ti time.Time) (passLightSample, error) {
				return defaultModule.passLightAt(orbit, obs, position, ti, opt)
			}
			for _, p := range passes {
				if p.Kind != PassVisible {
					if !p.BrightestTime.IsZero() || !math.IsNaN(p.BrightestMagnitude) {
						t.Fatal("non-visible pass has brightness")
					}
					continue
				}
				visible++
				if p.BrightestTime.Before(p.VisibleStart) || p.BrightestTime.After(p.VisibleEnd) {
					t.Fatal("brightest outside visible envelope")
				}
				look, err := LookAt(g.Elements, obs, p.BrightestTime, opt.ObserveOptions)
				light, lightErr := sample(p.BrightestTime)
				if err != nil || lightErr != nil || !look.SunCenterVisible || light.night < 0 || math.Float64bits(look.Magnitude) != math.Float64bits(p.BrightestMagnitude) {
					t.Fatalf("%s lat=%g brightness record=%g look=%g visible=%v err=%v/%v", g.Label, lat, p.BrightestMagnitude, look.Magnitude, look.SunCenterVisible, err, lightErr)
				}
				if math.Abs(p.BrightestTime.Sub(p.Culmination.Time).Seconds()) > 1 {
					different++
				}
				dense, _, err := passVisibility(p.Rise.Time, p.Set.Time, passStep(g.Elements)/8, sample)
				if err != nil || len(dense) == 0 {
					t.Fatalf("dense windows=%v err=%v", dense, err)
				}
				if math.Abs(p.VisibleStart.Sub(dense[0].start).Seconds()) > .05 || math.Abs(p.VisibleEnd.Sub(dense[len(dense)-1].end).Seconds()) > .05 {
					t.Fatal("coarse/dense visibility differs")
				}
				for _, span := range dense {
					for ti := span.start; ; {
						value, err := sample(ti)
						if err != nil {
							t.Fatal(err)
						}
						if value.night >= 0 && value.light > 0 && finite(value.magnitude) {
							excess := p.BrightestMagnitude - value.magnitude
							maxExcess = math.Max(maxExcess, excess)
							if excess > 1e-9 {
								t.Fatalf("%s lat=%g brightest=%g dense=%g at=%s", g.Label, lat, p.BrightestMagnitude, value.magnitude, ti)
							}
						}
						if !ti.Before(span.end) {
							break
						}
						ti = ti.Add(passStep(g.Elements) / 8)
						if ti.After(span.end) {
							ti = span.end
						}
					}
				}
			}
		}
	}
	if visible == 0 || different == 0 {
		t.Fatalf("missing visible/non-culmination cases %d/%d", visible, different)
	}
	t.Logf("visible passes=%d brightness differs from culmination by >1s=%d max brightness excess over dense=%.12g mag", visible, different, maxExcess)
}

func TestPassShortWindowSynthetic(t *testing.T) {
	g := readLookFixtures(t).Groups[0]
	obs := Observer{Longitude: g.Ground[0].Longitude, Latitude: g.Ground[0].Latitude}
	for _, duration := range []time.Duration{0, time.Nanosecond, time.Microsecond} {
		passes, state, err := Passes(g.Elements, obs, g.Elements.Epoch, g.Elements.Epoch.Add(duration), PassOptions{ObserveOptions: ObserveOptions{NoRefraction: true, StandardMagnitude: floatPointer(0)}})
		if err != nil || state != basic.AltitudeAlwaysAbove || len(passes) != 1 || passes[0].Duration != duration {
			t.Fatalf("duration=%v passes=%v state=%v err=%v", duration, passes, state, err)
		}
	}
}

func TestPassTwilightConsistency(t *testing.T) {
	g := readLookFixtures(t).Groups[0]
	start := g.Elements.Epoch
	obs := Observer{Latitude: 40}
	for _, twilight := range []float64{-90, 90} {
		opt := PassOptions{ObserveOptions: ObserveOptions{NoRefraction: true, StandardMagnitude: floatPointer(0)}, TwilightSunAltitude: &twilight}
		a, as, err := Passes(g.Elements, obs, start, start.Add(24*time.Hour), opt)
		if err != nil {
			t.Fatal(err)
		}
		b, bs, err := NewWithInstance(ephem.NewModelInstance(ephem.Active())).Passes(g.Elements, obs, start, start.Add(24*time.Hour), opt)
		if err != nil || as != bs || len(a) != len(b) {
			t.Fatalf("state/count differs %v", err)
		}
		visible := 0
		for i, p := range a {
			if !equalPassBits(p, b[i]) {
				t.Fatalf("pass=%d differs", i)
			}
			if twilight < 0 && p.Kind != PassDaylight {
				t.Fatal("unreachable night threshold must be daylight")
			}
			if p.Kind == PassVisible {
				visible++
				if !finite(p.BrightestMagnitude) || p.BrightestTime.IsZero() {
					t.Fatal("missing brightness")
				}
			}
		}
		if twilight > 0 && visible == 0 {
			t.Fatal("missing sunlit night-threshold cases")
		}
	}
}

func TestNextPassDecaySynthetic(t *testing.T) {
	e := readLookFixtures(t).Groups[0].Elements
	e.BStar = .02
	opt := PassOptions{ObserveOptions: ObserveOptions{NoRefraction: true}}
	obs := Observer{Latitude: 40}
	passes, _, err := Passes(e, obs, e.Epoch, e.Epoch.Add(24*time.Hour), opt)
	if err != nil || len(passes) == 0 {
		t.Fatalf("before decay passes=%d err=%v", len(passes), err)
	}
	got, err := NextPass(e, obs, e.Epoch, opt)
	if err != nil || !equalPassBits(got, passes[0]) {
		t.Fatalf("first pass before decay got=%+v want=%+v err=%v", got, passes[0], err)
	}
	_, err = NextPass(e, Observer{Latitude: 90}, e.Epoch, PassOptions{ObserveOptions: opt.ObserveOptions, MinElevation: 89})
	if !errors.Is(err, ErrDecayed) {
		t.Fatalf("no pass before decay: %v", err)
	}
	partial, _, err := Passes(e, obs, e.Epoch, e.Epoch.Add(30*24*time.Hour), opt)
	if !errors.Is(err, ErrDecayed) || partial != nil {
		t.Fatalf("full interval count=%d err=%v", len(partial), err)
	}
	t.Logf("before decay passes=%d first=%s; no-pass and full interval preserve ErrDecayed", len(passes), got.Culmination.Time)
}
