package satellite

import (
	"github.com/starainrt/astro/internal/ephem"
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"os"
	"testing"
	"time"
)

func TestTransitEventPathConsistency(t *testing.T) {
	g := readTransitEventFixtures(t).Groups[0]
	evs, err := TransitEvents(g.Elements, g.Start, g.End, TransitSearchOptions{Targets: []TransitTarget{g.Target}, MinTargetAltitude: g.Minimum})
	if err != nil {
		t.Fatal(err)
	}
	ev := evs[0]
	path, err := TransitEventPath(g.Elements, ev, 60*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if path.CenterLine[0] != ev.StartPoint || path.CenterLine[len(path.CenterLine)-1] != ev.EndPoint {
		t.Fatal("endpoints changed")
	}
	if !equalTransitEventBits(ev, path.Event) || path.StartHorizon[0] != (GeoPoint{ev.StartPoint.Longitude, ev.StartPoint.Latitude}) || path.EndHorizon[0] != (GeoPoint{ev.EndPoint.Longitude, ev.EndPoint.Latitude}) {
		t.Fatal("horizon joins changed")
	}
	for _, ring := range [][]GeoPoint{path.StartHorizon, path.EndHorizon} {
		if len(ring) < 361 || ring[0] != ring[len(ring)-1] {
			t.Fatal("open horizon")
		}
	}
	for _, step := range []time.Duration{-1, time.Millisecond, 61 * time.Second} {
		if _, err := TransitEventPath(g.Elements, ev, step); !errors.Is(err, ErrInvalidInput) {
			t.Fatal(step, err)
		}
	}
}

func TestTransitEventPathExternal(t *testing.T) {
	b, err := os.ReadFile("testdata/skyfield_transit_events.json")
	if err != nil {
		t.Fatal(err)
	}
	var data struct {
		Paths []struct {
			Label              string
			Target             TransitTarget
			Time               time.Time
			Point, Left, Right [2]float64
			HalfWidth          float64 `json:"half_width"`
		}
	}
	if err = json.Unmarshal(b, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Paths) != 11 {
		t.Fatal("missing independent limits")
	}
	maxPoint, maxWidth := 0., 0.
	for _, q := range data.Paths {
		for _, g := range readLookFixtures(t).Groups {
			if g.Label != q.Label {
				continue
			}
			events, err := TransitEvents(g.Elements, q.Time.Add(-time.Second), q.Time.Add(time.Second), TransitSearchOptions{Targets: []TransitTarget{q.Target}})
			if err != nil || len(events) != 1 {
				t.Fatal(err)
			}
			path, err := TransitEventPath(g.Elements, events[0], time.Second)
			if err != nil {
				t.Fatal(err)
			}
			for i, p := range []TransitPathPoint{path.CenterLine[1], path.LeftLimit[1], path.RightLimit[1]} {
				r := [][2]float64{q.Point, q.Left, q.Right}[i]
				d, _ := transitInverse(r[0], r[1], p.Longitude, p.Latitude)
				maxPoint = math.Max(maxPoint, d*1000)
				if d > .01 || !finite(d) {
					t.Fatalf("%s side=%d point difference=%g m", q.Label, i, d*1000)
				}
			}
			dw := math.Abs(path.CenterLine[1].HalfWidthKM/q.HalfWidth - 1)
			maxWidth = math.Max(maxWidth, dw)
			if dw > .01 {
				t.Fatal(q.Label, "width", dw)
			}
		}
	}
	t.Logf("11 independent centre/left/right points max=%g m; half-width relative max=%g", maxPoint, maxWidth)
}

func TestTransitEventPathHorizonExternal(t *testing.T) {
	maxResidual := 0.
	seams := 0
	for _, index := range []int{0, 3, 6, 9, 36, 39, 42, 45, 48, 51, 54, 57, 60, 63} {
		g := readTransitEventFixtures(t).Groups[index]
		events, err := TransitEvents(g.Elements, g.Start, g.End, TransitSearchOptions{Targets: []TransitTarget{g.Target}, MinTargetAltitude: g.Minimum})
		if err != nil || len(events) != 1 {
			t.Fatal(g.Label, err)
		}
		path, err := TransitEventPath(g.Elements, events[0], time.Minute)
		if err != nil {
			t.Fatal(g.Label, err)
		}
		for i, ring := range [][]GeoPoint{path.StartHorizon, path.EndHorizon} {
			v := g.States[i*2].Target
			samples := append([]GeoPoint(nil), ring...)
			for j := 1; j < len(ring); j++ {
				a, b := ring[j-1], ring[j]
				for _, f := range []float64{.25, .5, .75} {
					samples = append(samples, GeoPoint{a.Longitude + f*math.Remainder(b.Longitude-a.Longitude, 360), a.Latitude + f*(b.Latitude-a.Latitude)})
				}
			}
			for j, p := range samples {
				phi, lam := p.Latitude*math.Pi/180, p.Longitude*math.Pi/180
				e2 := 1 - (1-1/298.257223563)*(1-1/298.257223563)
				n := 6378.137 / math.Sqrt(1-e2*math.Sin(phi)*math.Sin(phi))
				r := [3]float64{n * math.Cos(phi) * math.Cos(lam), n * math.Cos(phi) * math.Sin(lam), n * (1 - e2) * math.Sin(phi)}
				d := [3]float64{v[0] - r[0], v[1] - r[1], v[2] - r[2]}
				up := [3]float64{math.Cos(phi) * math.Cos(lam), math.Cos(phi) * math.Sin(lam), math.Sin(phi)}
				alt := math.Asin((d[0]*up[0]+d[1]*up[1]+d[2]*up[2])/math.Sqrt(d[0]*d[0]+d[1]*d[1]+d[2]*d[2])) * 180 / math.Pi
				res := math.Abs(alt-g.Minimum) * 3600
				maxResidual = math.Max(maxResidual, res)
				if res > 1 {
					t.Fatal(g.Label, p, res)
				}
				if j > 0 && j < len(ring) && math.Abs(p.Longitude-ring[j-1].Longitude) > 180 {
					seams++
				}
			}
		}
	}
	if seams == 0 {
		t.Fatal("no longitude seam covered")
	}
	t.Logf("horizon independent altitude residual max=%g arcsec; seam edges=%d", maxResidual, seams)
}

func TestTransitEventPathInstanceConsistency(t *testing.T) {
	g := readTransitEventFixtures(t).Groups[0]
	evs, err := TransitEvents(g.Elements, g.Start, g.End, TransitSearchOptions{Targets: []TransitTarget{g.Target}})
	if err != nil {
		t.Fatal(err)
	}
	a, err := TransitEventPath(g.Elements, evs[0], time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewWithInstance(ephem.NewModelInstance(ephem.Semianalytic)).TransitEventPath(g.Elements, evs[0], time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	ab, ae := json.Marshal(a)
	bb, be := json.Marshal(b)
	if ae != nil || be != nil || !bytes.Equal(ab, bb) {
		t.Fatal("instance or finite fields", ae, be)
	}
	ev := evs[0]
	ev.End = ev.Start.Add(time.Hour)
	ev.EndPoint.Time = ev.End
	if _, err = TransitEventPath(g.Elements, ev, 10*time.Millisecond); !errors.Is(err, ErrInvalidInput) {
		t.Fatal("point cap", err)
	}
	short, err := TransitEvents(g.Elements, evs[0].Greatest.Add(-time.Second), evs[0].Greatest.Add(time.Second), TransitSearchOptions{Targets: []TransitTarget{g.Target}})
	if err != nil {
		t.Fatal(err)
	}
	path, err := TransitEventPath(g.Elements, short[0], 0)
	if err != nil || len(path.CenterLine) != 3 {
		t.Fatal("default step", err)
	}
}

type transitPathLimitFixture struct {
	Label               string
	Index               int
	Side                int
	Target              TransitTarget
	Time                time.Time
	Center              [4]float64
	Heading             float64
	Reference           [2]float64
	ReferenceCenter     [2]float64 `json:"reference_center"`
	ReferenceHeading    float64    `json:"reference_heading"`
	ReferenceCrossTrack float64    `json:"reference_cross_track"`
	HalfWidth           float64    `json:"half_width"`
	Satellite           [3]float64
	TargetPosition      [3]float64 `json:"target_position"`
}

func readTransitPathLimitFixtures(t *testing.T) []transitPathLimitFixture {
	t.Helper()
	raw, err := os.ReadFile("testdata/skyfield_transit_events.json")
	if err != nil {
		t.Fatal(err)
	}
	var data struct {
		Limits struct {
			Points []transitPathLimitFixture
		} `json:"path_limits"`
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	if len(data.Limits.Points) != 148 {
		t.Fatal("missing grazing limits", len(data.Limits.Points))
	}
	return data.Limits.Points
}

func transitPathLimitResiduals(q transitPathLimitFixture, p TransitPathPoint, minimum float64) (edge, ring, objective, width float64) {
	phi, lambda := p.Latitude*math.Pi/180, p.Longitude*math.Pi/180
	sp, cp := math.Sincos(phi)
	sl, cl := math.Sincos(lambda)
	f := 1 / 298.257223563
	e2 := f * (2 - f)
	n := 6378.137 / math.Sqrt(1-e2*sp*sp)
	observer := [3]float64{n * cp * cl, n * cp * sl, n * (1 - e2) * sp}
	var satellite, target [3]float64
	for i := range observer {
		satellite[i] = q.Satellite[i] - observer[i]
		target[i] = q.TargetPosition[i] - observer[i]
	}
	cross := [3]float64{target[1]*satellite[2] - target[2]*satellite[1], target[2]*satellite[0] - target[0]*satellite[2], target[0]*satellite[1] - target[1]*satellite[0]}
	norm := math.Sqrt(target[0]*target[0] + target[1]*target[1] + target[2]*target[2])
	separation := math.Atan2(math.Sqrt(cross[0]*cross[0]+cross[1]*cross[1]+cross[2]*cross[2]), target[0]*satellite[0]+target[1]*satellite[1]+target[2]*satellite[2])
	radius := [2]float64{695700, 1737.4}[q.Target]
	edge = math.Abs(separation-math.Asin(radius/norm)) * 180 / math.Pi * 3600
	altitude := math.Asin((target[0]*cp*cl+target[1]*cp*sl+target[2]*sp)/norm) * 180 / math.Pi
	ring = math.Min(math.Abs(altitude), math.Abs(altitude-minimum)) * 3600
	d, az := transitInverse(q.ReferenceCenter[0], q.ReferenceCenter[1], p.Longitude, p.Latitude)
	objective = math.Abs(d*math.Sin((q.ReferenceHeading-az)*math.Pi/180)-q.ReferenceCrossTrack) * 1000
	width = math.Abs(p.HalfWidthKM/q.HalfWidth - 1)
	return
}

func TestTransitEventPathGrazingExternal(t *testing.T) {
	groups := readTransitEventFixtures(t).Groups
	sky := newTransitSky(nil)
	var maximum float64
	var worst string
	alternative := 0
	for _, q := range readTransitPathLimitFixtures(t) {
		t.Run(q.Label, func(t *testing.T) {
			orbit, err := newSatelliteOrbit(groups[q.Index].Elements)
			if err != nil {
				t.Fatal(err)
			}
			g, err := transitGeometryAt(orbit, sky, q.Target, q.Time, true)
			if err != nil {
				t.Fatal(err)
			}
			p := TransitPathPoint{q.Time, q.Center[0], q.Center[1], q.Center[2], q.Center[3]}
			left, right, width := transitPathLimits(g, q.Target, p, q.Heading)
			projector := newTransitWidthProjector(p.Longitude, p.Latitude, q.Heading)
			_, best := transitPathExtrema(newTransitPathCone(g, q.Target), projector.crossTrack)
			fastWidth := math.Max(0, math.Max(best[0], best[1]))
			if !finite(fastWidth) || math.Abs(fastWidth-width) > 1e-8 {
				t.Fatalf("optimized width %.12f km, want %.12f km", fastWidth, width)
			}
			v := [2]TransitPathPoint{left, right}[q.Side-1]
			d, _ := transitInverse(q.Reference[0], q.Reference[1], v.Longitude, v.Latitude)
			if d*1000 > maximum {
				maximum, worst = d*1000, q.Label
			}
			if !finite(d) {
				t.Fatal("non-finite independent distance", d)
			}
			if d <= .01 {
				return
			}
			edge, ring, objective, width := transitPathLimitResiduals(q, v, groups[q.Index].Minimum)
			t.Logf("criterion(ii): point=%g m edge=%g arcsec ring=%g arcsec objective=%g m width-relative=%g", d*1000, edge, ring, objective, width)
			if !finite(edge) || !finite(ring) || !finite(objective) || !finite(width) || (edge > 1e-4 && ring > 1e-4) || objective > 10 || width > .01 {
				t.Fatal("independent limit fails both criteria")
			}
			alternative++
		})
	}
	t.Logf("148 grazing limits: criterion(i)=%d criterion(ii)=%d; point-distance max=%g m at %s", 148-alternative, alternative, maximum, worst)
}

func TestTransitEventPathDirectionContinuityConsistency(t *testing.T) {
	groups := readTransitEventFixtures(t).Groups
	sky := newTransitSky(nil)
	maximum := 0.
	for _, q := range readTransitPathLimitFixtures(t) {
		t.Run(q.Label, func(t *testing.T) {
			orbit, err := newSatelliteOrbit(groups[q.Index].Elements)
			if err != nil {
				t.Fatal(err)
			}
			g, err := transitGeometryAt(orbit, sky, q.Target, q.Time, true)
			if err != nil {
				t.Fatal(err)
			}
			p := TransitPathPoint{q.Time, q.Center[0], q.Center[1], q.Center[2], q.Center[3]}
			left, right, _ := transitPathLimits(g, q.Target, p, q.Heading)
			base := [2]TransitPathPoint{left, right}
			r := vectorNorm(g.target)
			ra := math.Atan2(g.target[1], g.target[0])
			dec := math.Asin(g.target[2] / r)
			delta := 1e-4 / 3600 * degreesToRadians
			for _, shift := range [][2]float64{{delta, 0}, {-delta, 0}, {0, delta}, {0, -delta}} {
				moved := g
				a, d := ra+shift[0], dec+shift[1]
				moved.target = [3]float64{r * math.Cos(d) * math.Cos(a), r * math.Cos(d) * math.Sin(a), r * math.Sin(d)}
				l, rr, _ := transitPathLimits(moved, q.Target, p, q.Heading)
				for side, v := range [2]TransitPathPoint{l, rr} {
					d, _ := transitInverse(base[side].Longitude, base[side].Latitude, v.Longitude, v.Latitude)
					maximum = math.Max(maximum, d*1000)
					if !finite(d) || d > .01 {
						t.Fatalf("side=%d shift=%v limit jump=%g m", side, shift, d*1000)
					}
				}
			}
		})
	}
	t.Logf("target direction perturbation max limit displacement=%g m", maximum)
}
