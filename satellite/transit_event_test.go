package satellite

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/ephem"
)

func TestTransitEventsShallowSynthetic(t *testing.T) {
	base := time.Unix(1794614400, 0)
	for _, depth := range []float64{-.001, .001, 1} {
		sample := func(at time.Time, exact bool) (float64, error) {
			x := at.Sub(base).Seconds() - 27.13
			return depth - x*x/1000, nil
		}
		for _, step := range []time.Duration{time.Minute, 7500 * time.Millisecond} {
			spans, err := findTransitSpans(base, base.Add(time.Minute), step, sample)
			want := btoi(depth > 0)
			if err != nil || len(spans) != want {
				t.Fatalf("depth=%g step=%s spans=%v err=%v", depth, step, spans, err)
			}
			if want == 0 {
				continue
			}
			width := math.Sqrt(depth * 1000)
			if math.Abs(spans[0].start.Sub(base).Seconds()-(math.Max(0, 27.13-width))) > 1e-5 || math.Abs(spans[0].end.Sub(base).Seconds()-(27.13+width)) > 1e-5 {
				t.Fatal(spans)
			}
		}
	}
}

func TestTransitEventsMaxHalfWidthNearHorizonConsistency(t *testing.T) {
	elements := readLookFixtures(t).Groups[0].Elements
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	events, err := TransitEvents(elements, start, start.Add(24*time.Hour), TransitSearchOptions{Targets: []TransitTarget{TransitSun}})
	if err != nil || len(events) == 0 {
		t.Fatalf("missing ISS solar transit: %v", err)
	}
	event := events[0]
	path, err := TransitEventPath(elements, event, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	orbit, err := newSatelliteOrbit(elements)
	if err != nil {
		t.Fatal(err)
	}
	sky := newTransitSky(&defaultModule)
	for index, point := range path.CenterLine {
		if point.HalfWidthKM > event.MaxHalfWidthKM {
			t.Fatalf("event max %.6f km is below path point %.6f km at %s", event.MaxHalfWidthKM, point.HalfWidthKM, point.Time)
		}
		if index > 8 && index+8 < len(path.CenterLine) {
			continue
		}
		heading, err := transitEventHeading(orbit, sky, event.Target, point.Time, event.Start, event.End)
		if err != nil {
			t.Fatal(err)
		}
		width := 0.0
		for _, limit := range [...]TransitPathPoint{path.LeftLimit[index], path.RightLimit[index]} {
			distance, bearing := transitInverse(point.Longitude, point.Latitude, limit.Longitude, limit.Latitude)
			width = math.Max(width, distance*math.Abs(math.Sin((heading-bearing)*degreesToRadians)))
		}
		if point.HalfWidthKM < width*0.99 {
			t.Fatalf("path half-width %.6f km is below its limit geometry %.6f km at %s", point.HalfWidthKM, width, point.Time)
		}
	}
}

func TestTransitEventsCountConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("30-day dense census")
	}
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	end := start.Add(30 * 24 * time.Hour)
	counts := [6][4]int{{389, 369, 404, 393}, {469, 469, 468, 468}, {460, 460, 459, 459}, {0, 0, 8, 8}, {13, 12, 9, 9}, {74, 74, 56, 56}}
	for i, g := range readLookFixtures(t).Groups {
		for j, target := range []TransitTarget{TransitSun, TransitMoon} {
			for k, minimum := range []float64{0, 10} {
				opt := TransitSearchOptions{Targets: []TransitTarget{target}, MinTargetAltitude: minimum}
				got, err := TransitEvents(g.Elements, start, end, opt)
				if err != nil || len(got) != counts[i][j*2+k] {
					t.Fatalf("%s %d %g count=%d want=%d err=%v", g.Label, target, minimum, len(got), counts[i][j*2+k], err)
				}
				dense, err := defaultModule.transitEvents(g.Elements, start, end, opt, 7500*time.Millisecond, newTransitSky(&defaultModule))
				if err != nil || len(dense) != len(got) {
					t.Fatalf("%s target=%d minimum=%g dense=%d coarse=%d err=%v", g.Label, target, minimum, len(dense), len(got), err)
				}
				for n, a := range got {
					b := dense[n]
					if a.Complete != b.Complete || math.Abs(a.Start.Sub(b.Start).Seconds()) > .01 || math.Abs(a.End.Sub(b.End).Seconds()) > .01 {
						t.Fatalf("%s %d coarse=%+v dense=%+v", g.Label, n, a, b)
					}
				}
				t.Logf("%s target=%d minimum=%.0f coarse=dense=%d", g.Label, target, minimum, len(got))
			}
		}
	}
}

type transitOracleState struct {
	Time              time.Time
	Point             [2]float64
	Satellite, Target [3]float64
	Altitude, DeltaT  float64 `json:"-"`
}
type transitEventFixture struct {
	Source, Version, Retrieved string
	Groups                     []struct {
		Label      string
		Elements   Elements
		Target     TransitTarget
		Minimum    float64
		Start, End time.Time
		Roots      [2]time.Time
		States     []transitOracleState
	}
	Shadows []struct {
		Label     string
		Kind      TransitTarget
		State     transitOracleState
		HalfWidth float64 `json:"half_width"`
	}
}

func readTransitEventFixtures(t *testing.T) transitEventFixture {
	t.Helper()
	b, err := os.ReadFile("testdata/skyfield_transit_events.json")
	if err != nil {
		t.Fatal(err)
	}
	var data transitEventFixture
	if err = json.Unmarshal(b, &data); err != nil {
		t.Fatal(err)
	}
	if data.Source == "" || data.Version == "" || data.Retrieved == "" {
		t.Fatal("missing provenance")
	}
	return data
}
func TestTransitEventsTangentExternal(t *testing.T) {
	for _, g := range readTransitEventFixtures(t).Groups {
		if g.Label != "geo-tle" || g.Target != TransitMoon || g.Minimum != 0 || g.States[0].Time.Day() != 18 {
			continue
		}
		events, err := TransitEvents(g.Elements, g.Start, g.End, TransitSearchOptions{Targets: []TransitTarget{TransitMoon}})
		if err != nil || len(events) != 1 {
			t.Fatal(events, err)
		}
		want := [][2]float64{{-127.23044733785623, 37.6252944138834}, {25.874415140740812, 61.39183368494252}}
		for i, p := range []TransitPathPoint{events[0].StartPoint, events[0].EndPoint} {
			d, _ := transitInverse(want[i][0], want[i][1], p.Longitude, p.Latitude)
			if d > .01 {
				t.Errorf("natural tangent endpoint %d difference=%g m", i, d*1000)
			}
		}
		return
	}
	t.Fatal("missing tangent regression")
}

func TestTransitEventsPartialLimbWidthExternal(t *testing.T) {
	for _, fixture := range readLookFixtures(t).Groups {
		if fixture.Label != "geo-tle" {
			continue
		}
		at := time.Date(2026, 11, 15, 14, 43, 18, 598405439, time.UTC)
		events, err := TransitEvents(fixture.Elements, at.Add(-time.Minute), at.Add(time.Minute), TransitSearchOptions{Targets: []TransitTarget{TransitSun}})
		if err != nil || len(events) != 1 {
			t.Fatalf("missing GEO transit: %v", err)
		}
		event := events[0]
		orbit, err := newSatelliteOrbit(fixture.Elements)
		if err != nil {
			t.Fatal(err)
		}
		point, err := transitEventPoint(orbit, newTransitSky(&defaultModule), TransitSun, at, event.Start, event.End, false)
		if err != nil || math.Abs(point.HalfWidthKM-199.291091900) > .01 {
			t.Fatalf("half-width = %.9f km, want 199.291091900 km: %v", point.HalfWidthKM, err)
		}
		return
	}
	t.Fatal("missing GEO fixture")
}
func TestTransitEventsExternal(t *testing.T) {
	var maxRoot, maxPoint float64
	count := 0
	for _, g := range readTransitEventFixtures(t).Groups {
		got, err := TransitEvents(g.Elements, g.Start, g.End, TransitSearchOptions{Targets: []TransitTarget{g.Target}, MinTargetAltitude: g.Minimum})
		if err != nil || len(got) != 1 {
			t.Fatal(g.Label, len(got), err)
		}
		event := got[0]
		orbit, err := newSatelliteOrbit(g.Elements)
		if err != nil {
			t.Fatal(err)
		}
		sky := newTransitSky(nil)
		for i, at := range []time.Time{event.Start, event.End} {
			d := math.Abs(at.Sub(g.Roots[i]).Seconds())
			maxRoot = math.Max(maxRoot, d)
			if d > .01 {
				t.Fatal(g.Label, "root", d)
			}
		}
		if event.MinTargetAltitude != g.Minimum || !event.Complete {
			t.Fatal(event)
		}
		for i, q := range g.States {
			p := []TransitPathPoint{event.StartPoint, event.GreatestPoint, event.EndPoint}[i]
			for _, v := range []float64{p.Longitude, p.Latitude, p.HalfWidthKM, p.TargetAltitude, event.MaxHalfWidthKM} {
				if !finite(v) {
					t.Fatal(g.Label, "nonfinite field", p)
				}
			}
			tangent := g.Minimum == 0 && i != 1
			recomputed, err := transitEventPoint(orbit, sky, g.Target, p.Time, event.Start, event.End, tangent)
			if err != nil || recomputed != p {
				t.Fatal("event point recomputation changed", g.Label, i, err)
			}
			p, err = transitEventPoint(orbit, sky, g.Target, q.Time, event.Start, event.End, tangent)
			if err != nil {
				t.Fatal(g.Label, i, err)
			}
			d, _ := transitInverse(q.Point[0], q.Point[1], p.Longitude, p.Latitude)
			maxPoint = math.Max(maxPoint, d*1000)
			if !finite(d) || d > .01 {
				t.Fatalf("%s point %s difference=%g m", g.Label, q.Time, d*1000)
			}
			count++
		}
	}
	t.Logf("%d same-time independent points max=%g m; natural roots max=%g s", count, maxPoint, maxRoot)
}
func TestTransitEventsInstanceConsistency(t *testing.T) {
	g := readTransitEventFixtures(t).Groups[0]
	opt := TransitSearchOptions{Targets: []TransitTarget{g.Target}, MinTargetAltitude: g.Minimum}
	a, err := TransitEvents(g.Elements, g.Start, g.End, opt)
	if err != nil || len(a) != 1 {
		t.Fatal(a, err)
	}
	b, err := NewWithInstance(ephem.NewModelInstance(ephem.Semianalytic)).TransitEvents(g.Elements, g.Start, g.End, opt)
	if err != nil || len(b) != 1 || !equalTransitEventBits(a[0], b[0]) {
		t.Fatal(a, b, err)
	}
}
func equalTransitEventBits(a, b TransitEvent) bool {
	if a.Target != b.Target || a.Start != b.Start || a.Greatest != b.Greatest || a.End != b.End || a.Complete != b.Complete || math.Float64bits(a.MinTargetAltitude) != math.Float64bits(b.MinTargetAltitude) || math.Float64bits(a.MaxHalfWidthKM) != math.Float64bits(b.MaxHalfWidthKM) {
		return false
	}
	for i, p := range []TransitPathPoint{a.StartPoint, a.GreatestPoint, a.EndPoint} {
		q := []TransitPathPoint{b.StartPoint, b.GreatestPoint, b.EndPoint}[i]
		if p.Time != q.Time {
			return false
		}
		for j, x := range []float64{p.Longitude, p.Latitude, p.HalfWidthKM, p.TargetAltitude} {
			y := []float64{q.Longitude, q.Latitude, q.HalfWidthKM, q.TargetAltitude}[j]
			if math.Float64bits(x) != math.Float64bits(y) {
				return false
			}
		}
	}
	return true
}
func TestTransitEventsSeedConsistency(t *testing.T) {
	data := readTransitEventFixtures(t)
	for i, g := range data.Groups {
		if testing.Short() && i > 1 {
			break
		}
		opt := TransitSearchOptions{Targets: []TransitTarget{g.Target}, MinTargetAltitude: g.Minimum}
		events, err := TransitEvents(g.Elements, g.Start, g.End, opt)
		if err != nil || len(events) != 1 {
			t.Fatal(g.Label, err)
		}
		want := events[0]
		for _, start := range []time.Time{want.Start.Add(-time.Hour), want.Start.Add(-3 * time.Hour), alignedTransitTime(want.Start, time.Minute).Add(time.Nanosecond)} {
			got, err := TransitEvents(g.Elements, start, want.End.Add(time.Hour), opt)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, a := range got {
				if a.Greatest.Equal(want.Greatest) {
					found = true
					if !equalTransitEventBits(a, want) {
						t.Fatal(g.Label, "window changed fields")
					}
				}
			}
			if !found {
				t.Fatal(g.Label, "missing seed event")
			}
			for _, seed := range []time.Time{want.Start.Add(-time.Minute), want.Start, want.Greatest, want.End, want.End.Add(time.Minute)} {
				got, err := ClosestTransitEvent(g.Elements, g.Target, seed, opt)
				if err != nil || !equalTransitEventBits(got, want) {
					t.Fatalf("%s seed=%s err=%v want=%+v got=%+v", g.Label, seed, err, want, got)
				}
			}
		}

		clipped, err := TransitEvents(g.Elements, want.Greatest.Add(-time.Second), want.Greatest.Add(time.Second), opt)
		if err != nil || len(clipped) != 1 || clipped[0].Complete || clipped[0].Start != want.Greatest.Add(-time.Second) || clipped[0].End != want.Greatest.Add(time.Second) || clipped[0].MinTargetAltitude != g.Minimum {
			t.Fatal(g.Label, clipped, err)
		}
	}
}
func TestTransitEventsInputSynthetic(t *testing.T) {
	g := readTransitEventFixtures(t).Groups[0]
	for _, opt := range []TransitSearchOptions{{MinTargetAltitude: -1}, {MinTargetAltitude: 91}, {MinTargetAltitude: math.NaN()}, {Targets: []TransitTarget{3}}} {
		if _, err := TransitEvents(g.Elements, g.Start, g.End, opt); !errors.Is(err, ErrInvalidInput) {
			t.Fatal(opt, err)
		}
	}
	for _, w := range [][2]time.Time{{{}, g.End}, {g.End, g.Start}, {g.Start, g.Start.Add(31 * 24 * time.Hour)}} {
		if _, err := TransitEvents(g.Elements, w[0], w[1], TransitSearchOptions{}); !errors.Is(err, ErrInvalidInput) {
			t.Fatal(w, err)
		}
	}
}
