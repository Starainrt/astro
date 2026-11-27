package satellite

import (
	"github.com/starainrt/astro/internal/ephem"
	"encoding/json"
	"errors"
	"math"
	"os"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestTransitSkyUnifiedConsistency(t *testing.T) {
	at := time.Date(2026, 11, 14, 0, 2, 3, 456789123, time.UTC)
	for _, target := range []TransitTarget{TransitSun, TransitMoon} {
		sky := newTransitSky(nil)
		coarse := sky.position(target, at, false)
		refined := sky.position(target, at, true)
		for i := range coarse {
			if math.Float64bits(coarse[i]) != math.Float64bits(refined[i]) {
				t.Fatalf("target=%d component=%d coarse and refined differ", target, i)
			}
		}
		if len(sky.values) != 5 {
			t.Fatal("five deterministic nodes required", len(sky.values))
		}
		for key := range sky.values {
			if modInt64(key.seconds, 600) != 0 {
				t.Fatal("node not on Unix 600-second grid", key)
			}
		}
	}
}

func directTransitEquatorial(target TransitTarget, at time.Time, inst *ephem.Instance) [3]float64 {
	jd := basic.UTC2TT(basic.Date2JD(at.UTC()))
	var ra, dec, distance float64
	if target == TransitSun {
		ra, dec = basic.HSunApparentRaDecNWithInstance(jd, -1, inst)
		distance = basic.EarthAwayNWithInstance(jd, -1, inst) * 149597870.7
	} else {
		ra, dec = basic.HMoonGeocentricApparentRaDecNWithInstance(jd, -1, inst)
		distance = basic.HMoonAwayNWithInstance(jd, -1, inst)
	}
	sr, cr := math.Sincos(ra * (math.Pi / 180))
	sd, cd := math.Sincos(dec * (math.Pi / 180))
	return [3]float64{distance * cd * cr, distance * cd * sr, distance * sd}
}

func TestTransitSkyInterpolationConsistency(t *testing.T) {
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	var maximum [2]float64
	siderealMaximum := 0.
	sky := newTransitSky(nil)
	for i := 0; i < 4096; i++ {
		at := start.Add(time.Duration((int64(i)*791900123457)%int64(30*24*time.Hour)) - 13*time.Minute)
		jdUT1 := basic.UTC2UT1(basic.Date2JD(at.UTC()))
		want := basic.ApparentSiderealTime(jdUT1) * 15 * (math.Pi / 180)
		difference := math.Abs(math.Remainder(sky.sidereal(at)-want, 2*math.Pi)) * 180 / math.Pi * 3600
		siderealMaximum = math.Max(siderealMaximum, difference)
		for _, target := range []TransitTarget{TransitSun, TransitMoon} {
			a := sky.equatorial(target, at)
			b := directTransitEquatorial(target, at, nil)
			na, nb := vectorNorm(a), vectorNorm(b)
			for k := range a {
				a[k] /= na
				b[k] /= nb
			}
			cross := [3]float64{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
			difference := math.Atan2(vectorNorm(cross), dotVector(a, b)) * 180 / math.Pi * 3600
			maximum[target] = math.Max(maximum[target], difference)
		}
	}
	t.Logf("4096 phases, interpolation/direct max Sun=%g Moon=%g GAST=%g arcsec", maximum[0], maximum[1], siderealMaximum)
	if siderealMaximum > 1e-6 {
		t.Fatal("interpolated GAST exceeds 1e-6 arcsec", siderealMaximum)
	}
}

func TestTransitMinimumSynthetic(t *testing.T) {
	base := time.Unix(1794614400, 0)
	for _, gap := range []float64{-0.1, 0, 0.1, 1.9, 2.1} {
		sample := func(at time.Time, exact bool) (transitSample, error) {
			x := at.Sub(base).Seconds() - .713
			return transitSample{separation: math.Hypot(x, 10+gap), radius: 10, satelliteAltitude: 40, targetAltitude: 40}, nil
		}
		for _, step := range []time.Duration{2 * time.Second, 250 * time.Millisecond} {
			got, err := findSiteTransits(base.Add(-3*time.Second), base.Add(3*time.Second), step, 2, 0, sample)
			if err != nil || len(got) != btoi(gap <= 2) {
				t.Fatalf("gap=%g step=%s events=%v err=%v", gap, step, got, err)
			}
			if len(got) == 0 {
				continue
			}
			if math.Abs(got[0].center.Sub(base).Seconds()-.713) > .001 || gap != 0 && got[0].nearMiss != (gap > 0) {
				t.Fatalf("gap=%g event=%+v", gap, got[0])
			}
			if gap < 0 {
				want := math.Sqrt(100 - (10+gap)*(10+gap))
				if math.Abs(got[0].start.Sub(base).Seconds()-(.713-want)) > .001 || math.Abs(got[0].end.Sub(base).Seconds()-(.713+want)) > .001 {
					t.Fatal(got[0])
				}
			}
		}
	}
}

func btoi(v bool) int {
	if v {
		return 1
	}
	return 0
}

type transitFixtures struct {
	Source, Version, Retrieved string
	Groups                     []struct {
		Label      string
		Elements   Elements
		Target     TransitTarget
		Observer   Observer
		Start, End time.Time
		Rows       [][]json.RawMessage
	}
}

func readTransitFixtures(t *testing.T) transitFixtures {
	t.Helper()
	b, err := os.ReadFile("testdata/skyfield_transits.json")
	if err != nil {
		t.Fatal(err)
	}
	var data transitFixtures
	if err = json.Unmarshal(b, &data); err != nil {
		t.Fatal(err)
	}
	if data.Source == "" || data.Version == "" || data.Retrieved == "" {
		t.Fatal("missing provenance")
	}
	return data
}
func TestTransitsExternal(t *testing.T) {
	var maximum [7]float64
	count := 0
	for _, g := range readTransitFixtures(t).Groups {
		got, err := Transits(g.Elements, g.Observer, g.Start, g.End, TransitOptions{Targets: []TransitTarget{g.Target}})
		if err != nil || len(got) != len(g.Rows) {
			t.Fatalf("%s count=%d want=%d err=%v", g.Label, len(got), len(g.Rows), err)
		}
		for i, r := range g.Rows {
			var times [3]time.Time
			var numbers [4]float64
			for j := range times {
				if err = json.Unmarshal(r[j], &times[j]); err != nil {
					t.Fatal(err)
				}
			}
			for j := range numbers {
				if err = json.Unmarshal(r[j+3], &numbers[j]); err != nil {
					t.Fatal(err)
				}
			}
			a := got[i]
			diff := [7]float64{a.Start.Sub(times[0]).Seconds(), a.Center.Sub(times[1]).Seconds(), a.End.Sub(times[2]).Seconds(), a.MinSeparationArcsec - numbers[0], a.TargetRadiusArcsec - numbers[1], (a.TargetAltitude - numbers[2]) * 3600, (a.CenterlineOffsetKM - numbers[3]) * 1000}
			limits := [7]float64{.01, .01, .01, .5, .01, 1, 10}
			for j, d := range diff {
				maximum[j] = math.Max(maximum[j], math.Abs(d))
				if !finite(d) || math.Abs(d) > limits[j] {
					t.Fatalf("%s field=%d signed=%g limit=%g", g.Label, j, d, limits[j])
				}
			}
			if a.Duration != a.End.Sub(a.Start) || a.NearMiss || a.Target != g.Target {
				t.Fatal(a)
			}
			count++
		}
	}
	t.Logf("%d events max |Start/Center/End| s, separation/radius/altitude arcsec, offset m: %v", count, maximum)
}
func equalTransitBits(a, b Transit) bool {
	return a.Target == b.Target && a.Start == b.Start && a.Center == b.Center && a.End == b.End && a.Duration == b.Duration && a.NearMiss == b.NearMiss && equalLookBits(a.Satellite, b.Satellite) && math.Float64bits(a.MinSeparationArcsec) == math.Float64bits(b.MinSeparationArcsec) && math.Float64bits(a.TargetRadiusArcsec) == math.Float64bits(b.TargetRadiusArcsec) && math.Float64bits(a.TargetAltitude) == math.Float64bits(b.TargetAltitude) && math.Float64bits(a.CenterlineOffsetKM) == math.Float64bits(b.CenterlineOffsetKM)
}

func TestTransitsMissingCenterlineConsistency(t *testing.T) {
	e := readLookFixtures(t).Groups[0].Elements
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	got, err := Transits(e, Observer{Latitude: 40}, start, start.Add(30*24*time.Hour), TransitOptions{NearMissArcsec: 3600})
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 12, 9, 16, 9, 6, 824000000, time.UTC)
	for _, event := range got {
		if event.Target != TransitSun || math.Abs(event.Center.Sub(want).Seconds()) > .01 {
			continue
		}
		if !event.NearMiss || !math.IsNaN(event.CenterlineOffsetKM) || event.Satellite.Elevation <= 0 {
			t.Fatalf("missing centre line: %+v", event)
		}
		assertTransitRayMiss(t, e, event)
		t.Logf("near miss retained: %+v", event)
		return
	}
	t.Fatal("missing solar near miss")
}

func assertTransitRayMiss(t *testing.T, e Elements, event Transit) {
	t.Helper()
	state, err := Propagate(e, event.Center)
	if err != nil {
		t.Fatal(err)
	}
	s, _ := earthFixed(state)
	v := newTransitSky(nil).position(event.Target, event.Center, true)
	a, b, c := 0., 0., -1.
	for i, radius := range [3]float64{6378.137, 6378.137, 6378.137 * (1 - 1/298.257223563)} {
		q, d := s[i]/radius, (s[i]-v[i])/radius
		a += d * d
		b += q * d
		c += q * q
	}
	if b*b-a*c >= 0 {
		t.Fatal("independent WGS84 discriminant does not miss")
	}
}

func TestTransitsLowTrueMissingCenterlineConsistency(t *testing.T) {
	e := readLookFixtures(t).Groups[0].Elements
	at := time.Date(2026, 11, 14, 0, 35, 2, 300000000, time.UTC)
	obs := Observer{Longitude: 61.56581150830282, Latitude: -39.63608224251665}
	got, err := Transits(e, obs, at.Add(-time.Minute), at.Add(time.Minute), TransitOptions{Targets: []TransitTarget{TransitSun}})
	if err != nil || len(got) != 1 {
		t.Fatalf("count=%d err=%v", len(got), err)
	}
	event := got[0]
	if event.NearMiss || !math.IsNaN(event.CenterlineOffsetKM) || event.Satellite.Elevation <= 0 || event.Satellite.Elevation >= .02 || event.Start.IsZero() || !event.Start.Before(event.Center) || !event.End.After(event.Center) {
		t.Fatalf("low true transit: %+v", event)
	}
	assertTransitRayMiss(t, e, event)
	t.Logf("low true transit retained: %+v", event)
}
func TestTransitsConsistency(t *testing.T) {
	for _, g := range readTransitFixtures(t).Groups {
		opt := TransitOptions{Targets: []TransitTarget{g.Target}}
		a, err := Transits(g.Elements, g.Observer, g.Start, g.End, opt)
		if err != nil || len(a) != 1 {
			t.Fatal(g.Label, a, err)
		}
		event := a[0]
		dense, err := defaultModule.transits(g.Elements, g.Observer, g.Start, g.End, opt, 250*time.Millisecond, newTransitSky(&defaultModule))
		if err != nil || len(dense) != 1 {
			t.Fatal(g.Label, dense, err)
		}
		for _, d := range []time.Duration{dense[0].Center.Sub(event.Center), dense[0].Start.Sub(event.Start), dense[0].End.Sub(event.End)} {
			if math.Abs(d.Seconds()) > .01 {
				t.Fatal(g.Label, d)
			}
		}
		for _, window := range [][2]time.Time{{event.Center, event.Center}, {event.Center.Add(-time.Millisecond), event.Center}, {event.Center, event.Center.Add(time.Millisecond)}, {g.Start, event.Center.Add(-time.Millisecond)}, {event.Center.Add(time.Millisecond), g.End}} {
			b, err := Transits(g.Elements, g.Observer, window[0], window[1], opt)
			want := !event.Center.Before(window[0]) && !event.Center.After(window[1])
			if err != nil || len(b) != btoi(want) || want && !equalTransitBits(event, b[0]) {
				t.Fatalf("%s window=%v count=%d want=%v err=%v", g.Label, window, len(b), want, err)
			}
		}
		inst := ephem.NewModelInstance(ephem.Semianalytic)
		b, err := NewWithInstance(inst).Transits(g.Elements, g.Observer, g.Start, g.End, opt)
		if err != nil || len(b) != 1 || !equalTransitBits(event, b[0]) {
			t.Fatal(g.Label, "instance", err)
		}
	}
}
func TestTransitsInputSynthetic(t *testing.T) {
	g := readTransitFixtures(t).Groups[0]
	for _, opt := range []TransitOptions{{Targets: []TransitTarget{TransitTarget(2)}}, {NearMissArcsec: -1}, {NearMissArcsec: math.NaN()}, {MinTargetAltitude: 91}, {MinTargetAltitude: math.Inf(1)}} {
		if _, err := Transits(g.Elements, g.Observer, g.Start, g.End, opt); !errors.Is(err, ErrInvalidInput) {
			t.Fatal(opt, err)
		}
	}
	for _, window := range [][2]time.Time{{{}, g.End}, {g.End, g.Start}, {g.Start, g.Start.Add(31 * 24 * time.Hour)}} {
		if _, err := Transits(g.Elements, g.Observer, window[0], window[1], TransitOptions{}); !errors.Is(err, ErrInvalidInput) {
			t.Fatal(window, err)
		}
	}
	base := time.Unix(1794614400, 0)
	sample := func(at time.Time, exact bool) (transitSample, error) {
		x := at.Sub(base).Seconds()
		return transitSample{separation: math.Abs(x), radius: 10, satelliteAltitude: 30, targetAltitude: 40}, nil
	}
	got, err := findSiteTransits(base.Add(-time.Second), base.Add(time.Second), 2*time.Second, 0, 0, sample)
	if err != nil || len(got) != 1 || got[0].start.Before(base.Add(-11*time.Second)) || !got[0].start.Before(base.Add(-time.Second)) || !got[0].end.After(base.Add(time.Second)) {
		t.Fatal(got, err)
	}
}

func TestTransitsDarkMoonSynthetic(t *testing.T) {
	e := readLookFixtures(t).Groups[0].Elements
	sky := newTransitSky(&defaultModule)
	orbit, err := newSatelliteOrbit(e)
	if err != nil {
		t.Fatal(err)
	}
	base := time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)
	for d := time.Duration(0); d < 24*time.Hour; d += 5 * time.Minute {
		at := base.Add(d)
		g, err := transitGroundAt(orbit, sky, TransitMoon, at)
		if err != nil {
			continue
		}
		lon, lat, _ := geographicPosition(g)
		obs := Observer{Longitude: lon, Latitude: lat}
		_, alt := horizontalDirection(subtractVector(sky.position(TransitMoon, at, true), g), obs)
		if alt < 30 {
			continue
		}
		got, err := Transits(e, obs, at.Add(-time.Minute), at.Add(time.Minute), TransitOptions{Targets: []TransitTarget{TransitMoon}})
		if err != nil || len(got) != 1 {
			t.Fatalf("dark Moon=%v err=%v", got, err)
		}
		t.Logf("dark Moon transit retained at %s", got[0].Center)
		return
	}
	t.Fatal("no dark Moon fixture")
}

func TestTransitsScanSkipConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("30-day scan equivalence matrix")
	}
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	total := 0
	for _, label := range []string{"iss-tle", "geo-tle", "molniya7276-tle"} {
		e := nearbyElements(t, label)
		latitudes := []float64{0, 40, -33.9, 51.5}
		if label != "iss-tle" {
			latitudes = []float64{40}
		}
		for _, lat := range latitudes {
			for _, near := range []float64{0, 1800} {
				for _, minimum := range []float64{0, -5} {
					obs := Observer{Latitude: lat, Longitude: lat * 2}
					opt := TransitOptions{NearMissArcsec: near, MinTargetAltitude: minimum}
					ref := newTransitSky(nil)
					ref.disableSiteScanSkip = true
					want, err := defaultModule.transits(e, obs, start, start.Add(30*24*time.Hour), opt, 2*time.Second, ref)
					if err != nil {
						t.Fatal(label, lat, near, minimum, err)
					}
					got, err := Transits(e, obs, start, start.Add(30*24*time.Hour), opt)
					if err != nil || len(got) != len(want) {
						t.Fatalf("%s lat=%g near=%g min=%g count=%d want=%d err=%v", label, lat, near, minimum, len(got), len(want), err)
					}
					for i := range got {
						if !equalTransitBits(got[i], want[i]) {
							t.Fatalf("%s lat=%g near=%g min=%g event=%d changed bits", label, lat, near, minimum, i)
						}
					}
					total += len(got)
					t.Logf("%s lat=%g near=%g min=%g: %d identical events", label, lat, near, minimum, len(got))
				}
			}
		}
	}
	t.Logf("24 thirty-day combinations: %d identical events", total)
}

func TestTransitScanUnavailableSynthetic(t *testing.T) {
	base := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	for _, missing := range []int{-1, 0, 1} {
		exactCalls := 0
		sample := func(at time.Time, exact bool) (transitSample, error) {
			x := at.Sub(base).Seconds()
			if exact {
				exactCalls++
			} else if x == float64(missing*2) {
				return transitSample{separation: math.NaN()}, nil
			}
			return transitSample{separation: 100 + x*x, radius: 100, satelliteAltitude: 40, targetAltitude: 40}, nil
		}
		got, err := findSiteTransits(base, base, 2*time.Second, 0, 0, sample)
		if err != nil || len(got) != 0 || exactCalls != 0 {
			t.Fatal(missing, got, exactCalls, err)
		}
	}
}

func TestTransitsDistantScanSkipConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("30-day distant scan equivalence")
	}
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	for _, label := range []string{"geo-Sun", "molniya7276-Sun"} {
		for _, g := range readTransitFixtures(t).Groups {
			if g.Label != label {
				continue
			}
			for _, near := range []float64{0, 1800} {
				for _, minimum := range []float64{0, -5} {
					ref := newTransitSky(nil)
					ref.disableSiteScanSkip = true
					opt := TransitOptions{NearMissArcsec: near, MinTargetAltitude: minimum}
					want, err := defaultModule.transits(g.Elements, g.Observer, start, start.Add(30*24*time.Hour), opt, 2*time.Second, ref)
					if err != nil || len(want) == 0 {
						t.Fatal(label, "missing reference transit", err)
					}
					got, err := Transits(g.Elements, g.Observer, start, start.Add(30*24*time.Hour), opt)
					if err != nil || len(got) != len(want) {
						t.Fatal(label, near, minimum, len(got), len(want), err)
					}
					for i := range got {
						if !equalTransitBits(got[i], want[i]) {
							t.Fatal(label, near, minimum, i, "changed bits")
						}
					}
					t.Logf("%s near=%g min=%g: %d identical events", label, near, minimum, len(got))
				}
			}
			break
		}
	}
}

func TestTransitScanStepConsistency(t *testing.T) {
	e := readLookFixtures(t).Groups[0].Elements
	orbit, err := newSatelliteOrbit(e)
	if err != nil {
		t.Fatal(err)
	}
	obs := Observer{Longitude: 61.56581150830282, Latitude: -39.63608224251665}
	position, err := observerPosition(obs)
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 11, 14, 0, 35, 5, 300000000, time.UTC)
	state, err := Propagate(e, at)
	if err != nil {
		t.Fatal(err)
	}
	r, velocity := earthFixed(state)
	d := subtractVector(r, position)
	_, altitude := horizontalDirection(d, obs)
	skipped := 0
	for _, step := range []time.Duration{250 * time.Millisecond, 2 * time.Second} {
		sky := newTransitSky(nil)
		scan := siteTransitScanner{orbit: orbit, sky: sky, observer: obs, position: position, step: step}
		wantSkip := altitude+2*step.Seconds()*vectorNorm(velocity)/vectorNorm(d)*180/math.Pi+1e-6 < 0
		got, err := scan.sample(TransitSun, at, false)
		if err != nil || math.IsNaN(got.separation) != wantSkip || (len(sky.values) == 0) != wantSkip {
			t.Fatalf("step=%s skip=%v sample=%+v nodes=%d err=%v", step, wantSkip, got, len(sky.values), err)
		}
		if wantSkip {
			skipped++
		}
		exact, err := scan.sample(TransitSun, at, true)
		if err != nil || !finite(exact.separation) {
			t.Fatal("exact sample skipped", exact, err)
		}
	}
	if skipped != 1 {
		t.Fatal("step-dependent guard not exercised", skipped)
	}
}
