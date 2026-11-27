package satellite

import (
	"github.com/starainrt/astro/internal/ephem"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"sync"
	"testing"
	"time"
)

func TestNearbyTransitsWindowConsistency(t *testing.T) {
	g := readTransitEventFixtures(t).Groups[0]
	evs, err := TransitEvents(g.Elements, g.Start, g.End, TransitSearchOptions{Targets: []TransitTarget{g.Target}})
	if err != nil {
		t.Fatal(err)
	}
	ev := evs[0]
	obs := Observer{Longitude: ev.GreatestPoint.Longitude, Latitude: ev.GreatestPoint.Latitude}
	start, end := ev.Greatest.Add(-20*time.Second), ev.Greatest.Add(20*time.Second)
	got, err := NearbyTransits([]Elements{g.Elements}, obs, start, end, NearbyTransitOptions{TransitSearchOptions: TransitSearchOptions{Targets: []TransitTarget{g.Target}}, RadiusKM: 50})
	if err != nil || len(got) != 1 {
		t.Fatal(got, err)
	}
	site, err := Transits(g.Elements, obs, start, end, TransitOptions{Targets: []TransitTarget{g.Target}})
	if err != nil || len(site) != 1 || got[0].Site == nil || !equalTransitBits(site[0], *got[0].Site) || !got[0].InBand {
		t.Fatal(got, site, err)
	}
}

type nearbyOracle struct {
	label      string
	target     TransitTarget
	minimum    float64
	obs        Observer
	radius     float64
	start, end time.Time
	rows       [][]float64
}

func readNearbyOracle(t *testing.T, key string) []nearbyOracle {
	t.Helper()
	b, err := os.ReadFile("testdata/skyfield_transit_events.json")
	if err != nil {
		t.Fatal(err)
	}
	var data map[string]json.RawMessage
	if err = json.Unmarshal(b, &data); err != nil {
		t.Fatal(err)
	}
	var out []nearbyOracle
	base := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	var records [][]json.RawMessage
	if err = json.Unmarshal(data[key], &records); err != nil {
		t.Fatal(err)
	}
	for _, r := range records {
		var a nearbyOracle
		read := func(i int, v interface{}) {
			if err = json.Unmarshal(r[i], v); err != nil {
				t.Fatal(err)
			}
		}
		read(0, &a.label)
		read(1, &a.target)
		var site [2]float64
		if key == "nearby_full" {
			read(2, &a.minimum)
			read(3, &site)
			read(4, &a.radius)
			read(5, &a.rows)
			a.start = base
			a.end = base.Add(30 * 24 * time.Hour)
		} else {
			read(2, &site)
			read(3, &a.radius)
			var lo, hi float64
			read(4, &lo)
			read(5, &hi)
			var q []float64
			read(6, &q)
			a.rows = [][]float64{q}
			a.start = base.Add(time.Duration(lo * 1e9))
			a.end = base.Add(time.Duration(hi * 1e9))
		}
		a.obs = Observer{Longitude: site[0], Latitude: site[1]}
		out = append(out, a)
	}
	return out
}
func nearbyElements(t *testing.T, label string) Elements {
	t.Helper()
	for _, g := range readLookFixtures(t).Groups {
		if g.Label == label {
			return g.Elements
		}
	}
	t.Fatal(label)
	return Elements{}
}
func auditNearby(t *testing.T, cases []nearbyOracle, dense bool) [5]float64 {
	t.Helper()
	base := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	mp, md, ms, mb := 0., 0., 0., 0.
	count := 0
	for ci, q := range cases {
		e := nearbyElements(t, q.label)
		opt := NearbyTransitOptions{TransitSearchOptions: TransitSearchOptions{Targets: []TransitTarget{q.target}, MinTargetAltitude: q.minimum}, RadiusKM: q.radius}
		got, err := NearbyTransits([]Elements{e}, q.obs, q.start, q.end, opt)
		if err != nil || len(got) != len(q.rows) {
			t.Fatalf("case=%d %s target=%d min=%g R=%g count=%d want=%d err=%v", ci, q.label, q.target, q.minimum, q.radius, len(got), len(q.rows), err)
		}
		for i, a := range got {
			r := q.rows[i]
			d, _ := transitInverse(r[1], r[2], a.Nearest.Longitude, a.Nearest.Latitude)
			dd, ds := math.Abs(a.NearestDistanceKM-r[3]), math.Abs(a.DistanceKM-r[4])
			bearing := math.Abs(math.Remainder(a.BearingDeg-r[5], 360))
			mp = math.Max(mp, d*1000)
			md = math.Max(md, dd*1000)
			ms = math.Max(ms, ds*1000)
			mb = math.Max(mb, bearing)
			if !finite(d) || d > .01 || dd > .01 || ds > .01 {
				t.Fatalf("%s case=%d row=%d nearest=%g m distance=%g m cross=%g m at=%s reference=%s", q.label, ci, i, d*1000, dd*1000, ds*1000, a.Nearest.Time, base.Add(time.Duration(r[0]*1e9)))
			}
			if a.InBand != (a.Site != nil && !a.Site.NearMiss) {
				t.Fatal("InBand")
			}
			count++
		}
		if dense {
			fine, err := defaultModule.nearbyTransits([]Elements{e}, q.obs, q.start, q.end, opt, 7500*time.Millisecond, 250*time.Millisecond)
			if err != nil || len(fine) != len(got) {
				t.Fatal("dense set", ci, err, len(fine), len(got))
			}
			for i, a := range got {
				d, _ := transitInverse(a.Nearest.Longitude, a.Nearest.Latitude, fine[i].Nearest.Longitude, fine[i].Nearest.Latitude)
				if d > .01 || !equalTransitEventBits(a.Event, fine[i].Event) {
					t.Fatal("dense mismatch", ci, d)
				}
			}
		}
		t.Logf("case=%d %s target=%d minimum=%g R=%g rows=%d", ci, q.label, q.target, q.minimum, q.radius, len(got))
	}
	t.Logf("%d independent nearby entries nearest=%g m distance=%g m cross=%g m bearing=%g deg", count, mp, md, ms, mb)
	return [5]float64{mp, md, ms, mb, float64(count)}
}
func TestNearbyTransitsOffsetsExternal(t *testing.T) {
	auditNearby(t, readNearbyOracle(t, "nearby_offsets"), true)
}
func TestNearbyTransitsCensusExternal(t *testing.T) {
	if testing.Short() {
		t.Skip("30-day nearby dense census")
	}
	cases := readNearbyOracle(t, "nearby_full")
	var maximum [5]float64
	var mu sync.Mutex
	for begin := 0; begin < len(cases); begin += 6 {
		end := begin + 6
		if end > len(cases) {
			end = len(cases)
		}
		t.Run(fmt.Sprintf("batch-%d", begin/6), func(t *testing.T) {
			for i := begin; i < end; i++ {
				q := cases[i]
				t.Run(fmt.Sprintf("case-%d", i), func(t *testing.T) {
					t.Parallel()
					got := auditNearby(t, []nearbyOracle{q}, true)
					mu.Lock()
					defer mu.Unlock()
					for j := 0; j < 4; j++ {
						maximum[j] = math.Max(maximum[j], got[j])
					}
					maximum[4] += got[4]
				})
			}
		})
		if t.Failed() {
			return
		}
	}
	if len(cases) != 216 || maximum[4] != 454 {
		t.Fatal("census totals", len(cases), maximum)
	}
	t.Logf("216 combinations/454 entries nearest=%g m distance=%g m cross=%g m bearing=%g deg", maximum[0], maximum[1], maximum[2], maximum[3])

}

func TestNearbyTransitsBoundaryConsistency(t *testing.T) {
	for _, q := range readNearbyOracle(t, "nearby_offsets") {
		e := nearbyElements(t, q.label)
		for _, delta := range []float64{-1, 1} {
			opt := NearbyTransitOptions{TransitSearchOptions: TransitSearchOptions{Targets: []TransitTarget{q.target}}, RadiusKM: q.rows[0][3] + delta}
			got, err := NearbyTransits([]Elements{e}, q.obs, q.start, q.end, opt)
			if err != nil || len(got) != btoi(delta > 0) {
				t.Fatal(q.label, delta, len(got), err)
			}
		}
	}
}
func TestNearbyTransitsBatchConsistency(t *testing.T) {
	q := readNearbyOracle(t, "nearby_offsets")[0]
	e := nearbyElements(t, q.label)
	opt := NearbyTransitOptions{TransitSearchOptions: TransitSearchOptions{Targets: []TransitTarget{q.target}}, RadiusKM: q.radius}
	a, err := NearbyTransits([]Elements{e}, q.obs, q.start, q.end, opt)
	if err != nil || len(a) != 1 {
		t.Fatal(a, err)
	}
	batch, err := NearbyTransits([]Elements{e, e}, q.obs, q.start, q.end, opt)
	if err != nil || len(batch) != 2 {
		t.Fatal(batch, err)
	}
	for i, row := range batch {
		if row.ElementsIndex != i {
			t.Fatal("index", row.ElementsIndex)
		}
		row.ElementsIndex = 0
		ab, _ := json.Marshal(a[0])
		bb, err := json.Marshal(row)
		if err != nil || !bytes.Equal(ab, bb) {
			t.Fatal("batch changed fields", err)
		}
	}
	inst, err := NewWithInstance(ephem.NewModelInstance(ephem.Semianalytic)).NearbyTransits([]Elements{e}, q.obs, q.start, q.end, opt)
	ab, _ := json.Marshal(a)
	bb, be := json.Marshal(inst)
	if err != nil || be != nil || !bytes.Equal(ab, bb) {
		t.Fatal("instance", err, be)
	}
	opt.NearMissArcsec = 10000
	near, err := NearbyTransits([]Elements{e}, q.obs, q.start, q.end, opt)
	if err != nil || len(near) != 1 || near[0].Site == nil || !near[0].Site.NearMiss || near[0].InBand {
		t.Fatal("near miss", near, err)
	}
}
func TestNearbyTransitsSiteWindowConsistency(t *testing.T) {
	for _, idx := range []int{0, 14} {
		g := readTransitFixtures(t).Groups[idx]
		opt := NearbyTransitOptions{TransitSearchOptions: TransitSearchOptions{Targets: []TransitTarget{g.Target}}, RadiusKM: 500}
		site, err := Transits(g.Elements, g.Observer, g.Start, g.End, TransitOptions{Targets: []TransitTarget{g.Target}})
		if err != nil || len(site) != 1 {
			t.Fatal(err)
		}
		center := site[0].Center
		for _, w := range [][2]time.Time{{center, center}, {center.Add(-time.Millisecond), center}, {center, center.Add(time.Millisecond)}, {center.Add(-time.Second), center.Add(-time.Millisecond)}, {center.Add(time.Millisecond), center.Add(time.Second)}} {
			near, err := NearbyTransits([]Elements{g.Elements}, g.Observer, w[0], w[1], opt)
			if err != nil || len(near) != 1 {
				t.Fatal(idx, w, near, err)
			}
			a := near[0]
			events, err := TransitEvents(g.Elements, w[0], w[1], opt.TransitSearchOptions)
			if err != nil || len(events) != 1 || !equalTransitEventBits(a.Event, events[0]) {
				t.Fatal("event query differs", err)
			}
			want, err := Transits(g.Elements, g.Observer, a.Event.Start, a.Event.End, TransitOptions{Targets: []TransitTarget{g.Target}})
			if err != nil || len(want) > 1 || len(want) == 0 && a.Site != nil || len(want) == 1 && (a.Site == nil || !equalTransitBits(want[0], *a.Site)) {
				t.Fatal("Site window", idx, w, err)
			}
		}
	}
}
func TestNearbyTransitsInputSynthetic(t *testing.T) {
	q := readNearbyOracle(t, "nearby_offsets")[0]
	e := nearbyElements(t, q.label)
	for _, opt := range []NearbyTransitOptions{{}, {RadiusKM: 501}, {RadiusKM: math.NaN()}, {RadiusKM: 50, NearMissArcsec: -1}, {RadiusKM: 50, TransitSearchOptions: TransitSearchOptions{MinTargetAltitude: -1}}} {
		if _, err := NearbyTransits([]Elements{e}, q.obs, q.start, q.end, opt); !errors.Is(err, ErrInvalidInput) {
			t.Fatal(opt, err)
		}
	}
	for _, es := range [][]Elements{nil, make([]Elements, 1001)} {
		if _, err := NearbyTransits(es, q.obs, q.start, q.end, NearbyTransitOptions{RadiusKM: 50}); !errors.Is(err, ErrInvalidInput) {
			t.Fatal(len(es), err)
		}
	}
}

func TestNearbyAntipodalSynthetic(t *testing.T) {
	for _, q := range [][6]float64{{179.35506364742187, .1367491829800372, 0, 0, 19958.893387256114, 0}, {0, 0, 180, 0, 20003.931458625448, 0}, {100, 40, -80, -40, 20003.931458625448, 0}} {
		d, _ := transitNearbyInverse(q[0], q[1], q[2], q[3])
		if !finite(d) || math.Abs(d-q[4]) > .01 {
			t.Fatal(q, d)
		}
	}
}

func TestNearbyTransitsEndpointDistanceConsistency(t *testing.T) {
	g := readTransitEventFixtures(t).Groups[0]
	at := g.States[0].Time.Add(300 * time.Second)
	opt := TransitSearchOptions{Targets: []TransitTarget{g.Target}}
	evs, err := TransitEvents(g.Elements, at, at.Add(time.Second), opt)
	if err != nil || len(evs) != 1 {
		t.Fatal(err)
	}
	ev := evs[0]
	o, _ := newSatelliteOrbit(g.Elements)
	heading, err := transitEventHeading(o, newTransitSky(&defaultModule), g.Target, at, ev.Start, ev.End)
	if err != nil {
		t.Fatal(err)
	}
	site := transitDirect(ev.StartPoint.Longitude, ev.StartPoint.Latitude, heading+180, 40)
	obs := Observer{Longitude: site.Longitude, Latitude: site.Latitude}
	for _, radius := range []float64{39, 41} {
		got, err := NearbyTransits([]Elements{g.Elements}, obs, ev.Start, ev.End, NearbyTransitOptions{TransitSearchOptions: opt, RadiusKM: radius})
		if err != nil || len(got) != btoi(radius > 40) {
			t.Fatal(radius, got, err)
		}
		if len(got) > 0 {
			a := got[0]
			if a.Nearest != a.Event.StartPoint || !equalTransitEventBits(a.Event, ev) || math.Abs(a.NearestDistanceKM-40) > .01 || math.Abs(a.DistanceKM) > .01 {
				t.Fatal("endpoint distance semantics", a)
			}
		}
	}
}
