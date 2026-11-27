package satellite

import (
	"github.com/starainrt/astro/internal/ephem"
	"encoding/json"
	"errors"
	"math"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestSphericalFootprintRadiusKM(t *testing.T) {
	for _, fixture := range []struct {
		name         string
		heightKM     float64
		elevationDeg float64
		radiusKM     float64
	}{
		{"horizon", 386.6, 0, 2165.458369593287},
		{"ten degrees", 386.6, 10, 1312.4190491265822},
		{"surface", 0, 0, 0},
		{"zenith", 386.6, 90, 0},
		{"tiny height at horizon", 1e-12, 0, math.Sqrt(2 * 6371 * 1e-12)},
		{"tiny height at 45 degrees", 1e-12, 45, 1e-12},
		{"very large height", math.MaxFloat64, 10, 6371 * 80 * math.Pi / 180},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			radiusKM, err := SphericalFootprintRadiusKM(fixture.heightKM, fixture.elevationDeg)
			if err != nil || !finite(radiusKM) || math.Abs(radiusKM-fixture.radiusKM) > 1e-12*fixture.radiusKM {
				t.Fatalf("radius = %.17g km, want %.17g km; error = %v", radiusKM, fixture.radiusKM, err)
			}
		})
	}
	for _, input := range [][2]float64{{-1, 0}, {0, -1}, {0, 91}, {math.NaN(), 0}, {0, math.NaN()}, {math.Inf(1), 0}, {0, math.Inf(1)}} {
		if _, err := SphericalFootprintRadiusKM(input[0], input[1]); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("input %v returned %v, want ErrInvalidInput", input, err)
		}
	}
}

func TestCoverageAtConsistency(t *testing.T) {
	e := readLookFixtures(t).Groups[0].Elements
	when := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	got, err := CoverageAt(e, when, CoverageOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Time != when || len(got.Footprint) < 361 || got.Footprint[0] != got.Footprint[len(got.Footprint)-1] || len(got.Night) < 361 {
		t.Fatal("invalid closed boundaries")
	}
	gp, err := GroundPointAt(e, when)
	if err != nil || gp != got.SubPoint {
		t.Fatal(got.SubPoint, gp, err)
	}
	look, err := LookAt(e, Observer{}, when, ObserveOptions{NoRefraction: true})
	if err != nil || look.Illumination != got.Illumination || look.SunCenterVisible != got.SunCenterVisible {
		t.Fatal("illumination", look, got, err)
	}
}

func TestCoverageExternal(t *testing.T) {
	data, err := os.ReadFile("testdata/skyfield_coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	var source struct {
		Groups []struct {
			Label    string
			Elements Elements
			Rows     []struct {
				Time           time.Time
				Satellite, Sun [3]float64
				CommonSun      [3]float64 `json:"common_sun"`
			}
		}
	}
	if err = json.Unmarshal(data, &source); err != nil {
		t.Fatal(err)
	}
	maximum := 0.0
	edges := 0
	for _, g := range source.Groups {
		for _, row := range g.Rows {
			for _, minimum := range []float64{0, 10, 80} {
				got, err := CoverageAt(g.Elements, row.Time, CoverageOptions{MinElevation: minimum})
				if err != nil {
					t.Fatalf("%s %s min%g: %v", g.Label, row.Time, minimum, err)
				}
				for k, ring := range [][]GeoPoint{got.Footprint, got.Night} {
					target, threshold := row.Satellite, minimum
					if k == 1 {
						target, threshold = row.Sun, -6
					}
					for i := 0; i < len(ring)-1; i++ {
						for _, fraction := range []float64{0, .25, .5, .75} {
							a, b := ring[i], ring[i+1]
							p := GeoPoint{Longitude: a.Longitude + fraction*math.Remainder(b.Longitude-a.Longitude, 360), Latitude: a.Latitude + fraction*(b.Latitude-a.Latitude)}
							alt := coverageReferenceAltitude(target, p)
							difference := math.Abs(alt-threshold) * 3600
							maximum = math.Max(maximum, difference)
							edges++
							if difference > 1 {
								t.Fatalf("%s %s min%g ring%d edge%d fraction%g residual=%g arcsec", g.Label, row.Time, minimum, k, i, fraction, difference)
							}
						}
					}
				}
				if got.Illumination == IlluminationUmbra && len(got.Visible) != 0 {
					t.Fatal("umbra visible")
				}
				for lat := -85.0; lat <= 85; lat += 10 {
					for lon := -175.0; lon < 180; lon += 10 {
						p := GeoPoint{Longitude: lon, Latitude: lat}
						sa := coverageReferenceAltitude(row.Satellite, p)
						ta := coverageReferenceAltitude(row.CommonSun, p)
						if math.Abs(sa-minimum)*3600 <= 1 || math.Abs(ta+6)*3600 <= 1 {
							continue
						}
						want := sa >= minimum && ta <= -6 && coverageReferenceSunVisible(row.Satellite, row.CommonSun)
						inside := false
						for _, r := range got.Visible {
							inside = inside || coverageReferenceContains(r, p)
						}
						if inside != want {
							t.Fatalf("%s %s min%g point%+v alt%g sun%g inside%v want%v", g.Label, row.Time, minimum, p, sa, ta, inside, want)
						}
					}
				}
			}
		}
	}
	t.Logf("independent boundary and edge samples=%d max error=%g arcsec", edges, maximum)
}

func coverageReferenceAltitude(vector [3]float64, p GeoPoint) float64 {
	phi, lambda := p.Latitude*math.Pi/180, p.Longitude*math.Pi/180
	up := [3]float64{math.Cos(phi) * math.Cos(lambda), math.Cos(phi) * math.Sin(lambda), math.Sin(phi)}
	e2 := (1 / 298.257223563) * (2 - 1/298.257223563)
	n := 6378.137 / math.Sqrt(1-e2*up[2]*up[2])
	r := [3]float64{n * up[0], n * up[1], n * (1 - e2) * up[2]}
	dot, norm := 0.0, 0.0
	for i := 0; i < 3; i++ {
		d := vector[i] - r[i]
		dot += d * up[i]
		norm += d * d
	}
	return math.Asin(math.Max(-1, math.Min(1, dot/math.Sqrt(norm)))) * 180 / math.Pi
}

func coverageReferenceContains(ring []GeoPoint, p GeoPoint) bool {
	if len(ring) < 3 {
		return false
	}
	points := make([][2]float64, len(ring))
	points[0] = [2]float64{ring[0].Longitude, ring[0].Latitude}
	for i := 1; i < len(ring); i++ {
		points[i] = [2]float64{points[i-1][0] + math.Remainder(ring[i].Longitude-points[i-1][0], 360), ring[i].Latitude}
	}
	if math.Abs(points[len(points)-1][0]-points[0][0]) > 180 {
		pole := 90.0
		mean := 0.0
		for _, q := range points {
			mean += q[1]
		}
		if mean < 0 {
			pole = -90
		}
		points = append(points, [2]float64{points[len(points)-1][0], pole}, [2]float64{points[0][0], pole})
	}
	for _, turn := range []float64{-720, -360, 0, 360, 720} {
		x, y := p.Longitude+turn, p.Latitude
		inside := false
		for i, a := range points {
			b := points[(i+1)%len(points)]
			if (a[1] > y) != (b[1] > y) && x < a[0]+(b[0]-a[0])*(y-a[1])/(b[1]-a[1]) {
				inside = !inside
			}
		}
		if inside {
			return true
		}
	}
	return false
}

func TestCoverageBetweenConsistency(t *testing.T) {
	e := readLookFixtures(t).Groups[0].Elements
	start := time.Date(2026, 11, 14, 0, 0, 3, 123, time.FixedZone("test", 8*3600))
	end := start.Add(125 * time.Second)
	got, err := CoverageBetween(e, start, end, time.Minute, CoverageOptions{})
	if err != nil {
		t.Fatal(err)
	}
	times := []time.Time{start, alignedTransitTime(start, time.Minute).Add(time.Minute), alignedTransitTime(start, time.Minute).Add(2 * time.Minute), end}
	if len(got.Instants) != len(times) {
		t.Fatal(len(got.Instants), times)
	}
	for i, when := range times {
		want, err := CoverageAt(e, when, CoverageOptions{})
		if err != nil || !reflect.DeepEqual(want, got.Instants[i]) || got.Track[i] != want.SubPoint || got.TrackIllumination[i] != want.Illumination {
			t.Fatal(i, err)
		}
	}
}

func TestCoverageBetweenUnixGridConsistency(t *testing.T) {
	elements := readLookFixtures(t).Groups[0].Elements
	start := time.Date(2026, 11, 14, 0, 0, 3, 123, time.FixedZone("station", 8*3600))
	end := start.Add(45 * time.Second)
	for _, step := range []time.Duration{11 * time.Second, 13 * time.Second, 23 * time.Second, 11500 * time.Millisecond} {
		got, err := CoverageBetween(elements, start, end, step, CoverageOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if got.Instants[0].Time != start || got.Instants[len(got.Instants)-1].Time != end {
			t.Fatal("coverage lost its original endpoints")
		}
		next := (start.UnixNano()/int64(step) + 1) * int64(step)
		for _, instant := range got.Instants[1 : len(got.Instants)-1] {
			if instant.Time.UnixNano() != next || instant.Time.Location() != start.Location() {
				t.Fatalf("step %s: got %s, want %s", step, instant.Time, time.Unix(0, next).In(start.Location()))
			}
			next += int64(step)
		}
		if next < end.UnixNano() {
			t.Fatalf("step %s: missing interior instant %s", step, time.Unix(0, next))
		}
	}
}

func BenchmarkCoverageBetweenISS(b *testing.B) {
	var data struct{ Groups []struct{ Elements Elements } }
	raw, err := os.ReadFile("testdata/skyfield_look.json")
	if err != nil {
		b.Fatal(err)
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		b.Fatal(err)
	}
	e := data.Groups[0].Elements
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := CoverageBetween(e, start, start.Add(90*time.Minute), time.Minute, CoverageOptions{}); err != nil {
			b.Fatal(err)
		}
	}
}

func TestCoverageModulesConsistency(t *testing.T) {
	before := ephem.Active()
	defer ephem.SetActive(before)
	for _, model := range []ephem.Model{ephem.Semianalytic, ephem.Legacy} {
		ephem.SetActive(model)
		inst := ephem.NewModelInstance(model)
		e := readLookFixtures(t).Groups[0].Elements
		when := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
		a, err := CoverageAt(e, when, CoverageOptions{})
		if err != nil {
			t.Fatal(err)
		}
		b, err := NewWithInstance(inst).CoverageAt(e, when, CoverageOptions{})
		if err != nil || !reflect.DeepEqual(a, b) {
			t.Fatalf("model%v paths differ %v", model, err)
		}
		av, err := CoverageBetween(e, when, when.Add(time.Minute), time.Minute, CoverageOptions{})
		if err != nil {
			t.Fatal(err)
		}
		bv, err := NewWithInstance(inst).CoverageBetween(e, when, when.Add(time.Minute), time.Minute, CoverageOptions{})
		if err != nil || !reflect.DeepEqual(av, bv) {
			t.Fatal("sweep paths", err)
		}
	}
}

func TestCoverageInvalidSynthetic(t *testing.T) {
	e := readLookFixtures(t).Groups[0].Elements
	t0 := e.Epoch
	nan, inf := math.NaN(), math.Inf(1)
	for _, opt := range []CoverageOptions{{MinElevation: nan}, {MinElevation: inf}, {MinElevation: -1}, {MinElevation: 91}, {TwilightSunAltitude: &nan}, {TwilightSunAltitude: &inf}} {
		if _, err := CoverageAt(e, t0, opt); !errors.Is(err, ErrInvalidInput) {
			t.Fatal(opt, err)
		}
	}
	for _, when := range []time.Time{{}, t0.AddDate(400, 0, 0)} {
		if _, err := CoverageAt(e, when, CoverageOptions{}); !errors.Is(err, ErrInvalidInput) {
			t.Fatal(when, err)
		}
	}
	if _, err := CoverageAt(Elements{}, t0, CoverageOptions{}); !errors.Is(err, ErrInvalidElements) {
		t.Fatal(err)
	}
	for _, q := range []struct {
		start, end time.Time
		step       time.Duration
	}{
		{time.Time{}, t0, time.Minute}, {t0, time.Time{}, time.Minute}, {t0, t0.Add(-time.Second), time.Minute},
		{t0, t0.Add(24*time.Hour + time.Nanosecond), time.Minute}, {t0, t0, 9 * time.Second}, {t0, t0, 10*time.Minute + time.Nanosecond},
		{t0, t0.Add(24 * time.Hour), 10 * time.Second},
	} {
		if _, err := CoverageBetween(e, q.start, q.end, q.step, CoverageOptions{}); !errors.Is(err, ErrInvalidInput) {
			t.Fatal(q, err)
		}
	}
}

func TestCoverageSweepExternal(t *testing.T) {
	data, err := os.ReadFile("testdata/skyfield_coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	var source struct {
		Groups []struct {
			Label    string
			Elements Elements
			Rows     []struct {
				Time           time.Time
				Satellite, Sun [3]float64
				CommonSun      [3]float64 `json:"common_sun"`
			}
		}
	}
	if err = json.Unmarshal(data, &source); err != nil {
		t.Fatal(err)
	}
	checks := 0
	for _, g := range source.Groups {
		start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
		end := start.Add(10 * time.Minute)
		got, err := CoverageBetween(g.Elements, start, end, time.Minute, CoverageOptions{})
		if err != nil {
			t.Fatal(g.Label, err)
		}
		for lat := -85.0; lat <= 85; lat += 10 {
			for lon := -175.0; lon < 180; lon += 10 {
				p := GeoPoint{Longitude: lon, Latitude: lat}
				want, wantVisible := false, false
				border := false
				for _, row := range g.Rows {
					if row.Time.Before(start) || row.Time.After(end) || row.Time.Second() != 0 {
						continue
					}
					alt := coverageReferenceAltitude(row.Satellite, p)
					sa := coverageReferenceAltitude(row.CommonSun, p)
					eNorm := 0.0
					sNorm := 0.0
					dot := 0.0
					for i := 0; i < 3; i++ {
						d := row.CommonSun[i] - row.Satellite[i]
						eNorm += row.Satellite[i] * row.Satellite[i]
						sNorm += d * d
						dot -= row.Satellite[i] * d
					}
					visible := math.Acos(math.Max(-1, math.Min(1, dot/math.Sqrt(eNorm*sNorm)))) > math.Asin(6378.137/math.Sqrt(eNorm))
					if math.Abs(alt)*3600 <= 1 || math.Abs(sa+6)*3600 <= 1 {
						border = true
					}
					want = want || alt >= 0
					wantVisible = wantVisible || alt >= 0 && sa <= -6 && visible
				}
				if border {
					continue
				}
				contains := func(rs []Region) bool {
					for _, region := range rs {
						if !coverageReferenceChartContains(region.Outer, p) {
							continue
						}
						hole := false
						for _, h := range region.Holes {
							hole = hole || coverageReferenceChartContains(h, p)
						}
						if !hole {
							return true
						}
					}
					return false
				}
				if contains(got.Swept) != want || contains(got.VisibleSwept) != wantVisible {
					t.Fatalf("%s %+v inside=%v/%v want=%v/%v", g.Label, p, contains(got.Swept), contains(got.VisibleSwept), want, wantVisible)
				}
				checks++
			}
		}
	}
	t.Logf("independent swept/visible-swept membership checks=%d", checks)
}

func coverageReferenceChartContains(ring []GeoPoint, p GeoPoint) bool {
	inside := false
	for i, a := range ring {
		b := ring[(i+1)%len(ring)]
		if (a.Latitude > p.Latitude) != (b.Latitude > p.Latitude) && p.Longitude < a.Longitude+(b.Longitude-a.Longitude)*(p.Latitude-a.Latitude)/(b.Latitude-a.Latitude) {
			inside = !inside
		}
	}
	return inside
}

func TestCoverageAdaptiveSynthetic(t *testing.T) {
	if testing.Short() {
		t.Skip("large adaptive sweep")
	}
	e := readLookFixtures(t).Groups[0].Elements
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		minimum        float64
		duration       time.Duration
		regions, holes int
	}{{89, 10 * time.Minute, 1, 0}, {80, 6 * time.Hour, 2, 6}} {
		before := time.Now()
		sweep, err := CoverageBetween(e, start, start.Add(tc.duration), time.Minute, CoverageOptions{MinElevation: tc.minimum})
		if err != nil {
			t.Fatal(tc.minimum, err)
		}
		holes := 0
		for _, r := range sweep.Swept {
			holes += len(r.Holes)
		}
		if len(sweep.Swept) != tc.regions || holes != tc.holes {
			t.Fatalf("min%g regions=%d holes=%d", tc.minimum, len(sweep.Swept), holes)
		}
		data, err := os.ReadFile("testdata/skyfield_coverage.json")
		if err != nil {
			t.Fatal(err)
		}
		var source struct {
			Adaptive []struct {
				Minimum, DenseAreaKM2 float64
				Witnesses             []GeoPoint
			}
		}
		if err = json.Unmarshal(data, &source); err != nil {
			t.Fatal(err)
		}
		for _, ref := range source.Adaptive {
			if ref.Minimum == tc.minimum {
				area := coverageTestArea(sweep.Swept)
				t.Logf("min%g area=%g dense=%g ratio=%g", tc.minimum, area, ref.DenseAreaKM2, area/ref.DenseAreaKM2)
				if area/ref.DenseAreaKM2 < .99 {
					t.Fatal("area below independent dense scan")
				}
				for _, p := range ref.Witnesses {
					if coverageIndex(sweep.Swept).contains(p) {
						t.Fatal("hole witness covered", p)
					}
				}
			}
		}
		index := coverageIndex(sweep.Swept)
		for _, q := range sweep.Instants {
			for _, p := range q.Footprint {
				if !index.contains(p) {
					t.Fatalf("min%g missing %+v", tc.minimum, p)
				}
			}
		}
		t.Logf("min%g output=%d regions=%d holes=%d elapsed=%s", tc.minimum, len(sweep.Instants), len(sweep.Swept), holes, time.Since(before))
	}
}

func TestCoverageSweepRetainsTinyNightCap(t *testing.T) {
	at := time.Date(2026, 11, 14, 1, 0, 0, 0, time.UTC)
	threshold := -89.999999
	for _, group := range readLookFixtures(t).Groups {
		if group.Label != "gps-tle" {
			continue
		}
		for _, end := range []time.Time{at, at.Add(time.Minute)} {
			sweep, err := CoverageBetween(group.Elements, at, end, time.Minute, CoverageOptions{TwilightSunAltitude: &threshold})
			if err != nil {
				t.Fatal(err)
			}
			if len(sweep.Swept) == 0 || len(sweep.VisibleSwept) != len(sweep.Instants) {
				t.Fatalf("end %s: got %d small swept caps, want %d", end, len(sweep.VisibleSwept), len(sweep.Instants))
			}
		}
		return
	}
	t.Fatal("missing GPS coverage fixture")
}

func TestCoverageExternalGrazingTwilight(t *testing.T) {
	for _, group := range readLookFixtures(t).Groups {
		switch group.Label {
		case "iss-tle":
			when := time.Date(2026, 11, 14, 19, 27, 20, 0, time.UTC)
			threshold := 89.99
			coverage, err := CoverageAt(group.Elements, when, CoverageOptions{MinElevation: 1.030449263063, TwilightSunAltitude: &threshold})
			point := GeoPoint{Longitude: -115.7102237485, Latitude: -18.3852806440}
			if err != nil || !coverage.SunCenterVisible || !coverageReferenceContains(coverage.Footprint, point) {
				t.Fatalf("ISS probe must be inside the geometric footprint: %v", err)
			}
			for _, ring := range coverage.Visible {
				if coverageReferenceContains(ring, point) {
					t.Fatal("ISS probe is in daylight, not visual coverage")
				}
			}
			if coverageIndex(coverage.VisibleRegions).contains(point) {
				t.Fatal("ISS probe is in daylight, not visual coverage")
			}
		case "gps-tle":
			when := time.Date(2026, 11, 14, 18, 23, 0, 0, time.UTC)
			threshold := -57.8437694586
			coverage, err := CoverageAt(group.Elements, when, CoverageOptions{TwilightSunAltitude: &threshold})
			if err != nil || !coverage.SunCenterVisible || len(coverage.Visible)+len(coverage.VisibleRegions) == 0 {
				t.Fatalf("GPS footprint must retain a nonempty sliver of night: %v", err)
			}
		}
	}
}

func TestCoverageZenithHasNoArea(t *testing.T) {
	elements := readLookFixtures(t).Groups[0].Elements
	at := time.Date(2026, 11, 14, 19, 27, 20, 0, time.UTC)
	for _, minimum := range []float64{89.999999, 90} {
		for _, end := range []time.Time{at, at.Add(time.Minute)} {
			sweep, err := CoverageBetween(elements, at, end, time.Minute, CoverageOptions{MinElevation: minimum})
			if err != nil || len(sweep.Track) == 0 || len(sweep.Track) != len(sweep.Instants) || len(sweep.Swept)+len(sweep.VisibleSwept) != 0 {
				t.Fatalf("%g°: expected track without swept area: %v", minimum, err)
			}
			for _, instant := range sweep.Instants {
				if instant.SubPoint.Time.IsZero() || len(instant.Footprint)+len(instant.Visible)+len(instant.VisibleRegions) != 0 {
					t.Fatalf("%g°: expected subpoint without footprint", minimum)
				}
			}
		}
	}
}

func TestCoverageSweepChecksNearZenithAreaTransition(t *testing.T) {
	threshold := 90.0
	opt := CoverageOptions{MinElevation: 89.99999139265660, TwilightSunAltitude: &threshold}
	start := time.Date(2026, 11, 14, 4, 0, 0, 0, time.UTC)
	for _, group := range readLookFixtures(t).Groups {
		if group.Label != "molniya7276-tle" {
			continue
		}
		first, err := CoverageAt(group.Elements, start, opt)
		if err != nil || len(first.Footprint) != 0 {
			t.Fatalf("expected a start without area: %v", err)
		}
		middle, err := CoverageAt(group.Elements, start.Add(5*time.Minute), opt)
		if err != nil || len(middle.Footprint) == 0 || len(middle.Visible) == 0 {
			t.Fatalf("expected intermediate geometric and visual area: %v", err)
		}
		_, err = CoverageBetween(group.Elements, start, start.Add(10*time.Minute), 10*time.Minute, opt)
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected the sweep to hit its resolution limit, got %v", err)
		}
		return
	}
	t.Fatal("missing Molniya coverage fixture")
}

func TestCoveragePolarDaylightExclusion(t *testing.T) {
	at := time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC)
	var elements Elements
	for _, group := range readLookFixtures(t).Groups {
		if group.Label == "molniya7276-tle" {
			elements = group.Elements
			break
		}
	}
	jd := basic.Date2JD(at)
	sun := apparentSunPosition(basic.UTC2TT(jd), basic.ApparentSiderealTime(basic.UTC2UT1(jd))*15*degreesToRadians, nil)
	sunLongitude, sunLatitude, _ := geographicPosition(sun)
	day := GeoPoint{sunLongitude, sunLatitude}
	for _, threshold := range []float64{79.9, 80, 80.1} {
		coverage, err := CoverageAt(elements, at, CoverageOptions{TwilightSunAltitude: &threshold})
		if err != nil {
			t.Fatalf("threshold %g: %v", threshold, err)
		}
		if !coverage.SunCenterVisible || len(coverage.VisibleRegions) == 0 || len(coverage.Visible) != 0 || !coverageReferenceContains(coverage.Footprint, day) {
			t.Fatalf("threshold %g: missing daylight exclusion inside the geometric footprint", threshold)
		}
		index := coverageIndex(coverage.VisibleRegions)
		if index.contains(day) {
			t.Fatalf("threshold %g: subsolar point filled", threshold)
		}
		for _, longitude := range []float64{-179.9, 179.9} {
			if !index.contains(GeoPoint{longitude, 89.9}) {
				t.Fatalf("threshold %g: lost polar night coverage at longitude %g", threshold, longitude)
			}
		}
	}
}

func TestCoveragePolarSeamSynthetic(t *testing.T) {
	fixture := readLookFixtures(t)
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	polar, seam, umbra := false, false, false
	for _, g := range fixture.Groups {
		for k := 0; k < 24; k++ {
			c, err := CoverageAt(g.Elements, start.Add(time.Duration(k)*time.Hour), CoverageOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if c.Illumination == IlluminationUmbra {
				umbra = true
				if len(c.Visible) != 0 {
					t.Fatal("umbra visible")
				}
			}
			for i, p := range c.Footprint {
				if i > 0 && math.Abs(p.Longitude-c.Footprint[i-1].Longitude) > 180 {
					seam = true
				}
			}
			polar = polar || coverageReferenceContains(c.Footprint, GeoPoint{Latitude: 89.999}) || coverageReferenceContains(c.Footprint, GeoPoint{Latitude: -89.999})
		}
	}
	if !polar || !seam || !umbra {
		t.Fatal("missing cases", polar, seam, umbra)
	}
}

func coverageReferenceSunVisible(r, sun [3]float64) bool {
	dot, a, b := 0.0, 0.0, 0.0
	for i := range r {
		d := sun[i] - r[i]
		dot -= r[i] * d
		a += r[i] * r[i]
		b += d * d
	}
	return math.Acos(math.Max(-1, math.Min(1, dot/math.Sqrt(a*b)))) > math.Asin(6378.137/math.Sqrt(a))
}

type coverageTestEdge struct{ a, b GeoPoint }
type coverageTestIndex [][]coverageTestEdge

func coverageIndex(regions []Region) coverageTestIndex {
	out := make(coverageTestIndex, 1801)
	for _, r := range regions {
		rings := append([][]GeoPoint{r.Outer}, r.Holes...)
		for _, ring := range rings {
			for i, a := range ring {
				b := ring[(i+1)%len(ring)]
				low := int(math.Floor((math.Min(a.Latitude, b.Latitude) + 90) * 10))
				high := int(math.Floor((math.Max(a.Latitude, b.Latitude) + 90) * 10))
				for k := low; k <= high; k++ {
					out[k] = append(out[k], coverageTestEdge{a, b})
				}
			}
		}
	}
	return out
}
func (idx coverageTestIndex) contains(p GeoPoint) bool {
	inside := false
	for _, edge := range idx[int(math.Floor((p.Latitude+90)*10))] {
		a, b := edge.a, edge.b
		cross := (p.Longitude-a.Longitude)*(b.Latitude-a.Latitude) - (p.Latitude-a.Latitude)*(b.Longitude-a.Longitude)
		if math.Abs(cross) < 1e-9 && p.Longitude >= math.Min(a.Longitude, b.Longitude)-1e-9 && p.Longitude <= math.Max(a.Longitude, b.Longitude)+1e-9 && p.Latitude >= math.Min(a.Latitude, b.Latitude)-1e-9 && p.Latitude <= math.Max(a.Latitude, b.Latitude)+1e-9 {
			return true
		}
		if (a.Latitude > p.Latitude) != (b.Latitude > p.Latitude) && p.Longitude < a.Longitude+(b.Longitude-a.Longitude)*(p.Latitude-a.Latitude)/(b.Latitude-a.Latitude) {
			inside = !inside
		}
	}
	return inside
}

func coverageTestArea(regions []Region) float64 {
	const a = 6378.137
	e2 := (1 / 298.257223563) * (2 - 1/298.257223563)
	e := math.Sqrt(e2)
	primitive := func(phi float64) float64 {
		sn := math.Sin(phi)
		return a * a * (1 - e2) / 2 * (sn/(1-e2*sn*sn) + math.Atanh(e*sn)/e)
	}
	area := 0.0
	for _, r := range regions {
		for _, ring := range append([][]GeoPoint{r.Outer}, r.Holes...) {
			for i, p := range ring {
				q := ring[(i+1)%len(ring)]
				for k := 0; k < 16; k++ {
					f := (float64(k) + .5) / 16
					area -= primitive((p.Latitude+f*(q.Latitude-p.Latitude))*math.Pi/180) * (q.Longitude - p.Longitude) * math.Pi / 180 / 16
				}
			}
		}
	}
	return area
}

func TestCoverageResourceSynthetic(t *testing.T) {
	if testing.Short() {
		t.Skip("internal sampling limit")
	}
	e := readLookFixtures(t).Groups[0].Elements
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	if _, err := CoverageBetween(e, start, start.Add(2*time.Hour), 10*time.Minute, CoverageOptions{MinElevation: 89}); !errors.Is(err, ErrInvalidInput) {
		t.Fatal("internal sample limit", err)
	}
}

func TestCoverageSweepElementsConsistency(t *testing.T) {
	e := readLookFixtures(t).Groups[0].Elements
	e.Name = "原样名称"
	e.Epoch = e.Epoch.In(time.FixedZone("epoch", 9*3600))
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.FixedZone("input", -5*3600))
	sweep, err := CoverageBetween(e, start, start.Add(time.Minute), time.Minute, CoverageOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if sweep.Elements != e {
		t.Fatalf("elements changed: %+v != %+v", sweep.Elements, e)
	}
	inst := ephem.NewModelInstance(ephem.Semianalytic)
	explicit, err := NewWithInstance(inst).CoverageBetween(e, start, start.Add(time.Minute), time.Minute, CoverageOptions{})
	if err != nil || explicit.Elements != e {
		t.Fatal("explicit elements changed", err)
	}
}

func TestCoverageOblateFourCrossingsSynthetic(t *testing.T) {
	sat := [3]float64{42164.0, 0, 0}
	foot, err := newCoverageBoundary(sat, GeoPoint{Longitude: 0, Latitude: 0}, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	sun := [3]float64{1.496e8, 0, 0}
	// 赤道 GEO 在约 +8.69° 晨昏角下，扁球足迹与恒定太阳高度圆产生 4 个交点。
	night, err := newTwilightBoundary(sun, 8.69)
	if err != nil {
		t.Fatal(err)
	}
	region, overlay, err := coverageIntersection(&foot, &night)
	if err != nil {
		t.Fatalf("4-crossing oblate intersection failed: %v", err)
	}
	if len(overlay) == 0 && len(region.Outer) == 0 {
		t.Fatal("expected non-empty visual coverage for 4-crossing oblate case")
	}
}
