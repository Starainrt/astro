package satellite

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/ephem"
)

func TestTransitShadowExternal(t *testing.T) {
	maxPoint, maxContact, maxWidth := 0., 0., 0.
	for _, q := range readTransitEventFixtures(t).Shadows {
		e := nearbyElements(t, q.Label)
		got, err := TransitShadowAt(e, q.Kind, q.State.Time, TransitSearchOptions{})
		if err != nil || !got.Valid {
			t.Fatal(q.Label, got, err)
		}
		if got.Time != q.State.Time || got.Target != q.Kind || len(got.Footprint) != 73 || got.Footprint[0] != got.Footprint[72] {
			t.Fatal(got)
		}
		phi, lam := q.State.Point[1]*math.Pi/180, q.State.Point[0]*math.Pi/180
		e2 := (2 - 1/298.257223563) / 298.257223563
		n := 6378.137 / math.Sqrt(1-e2*math.Sin(phi)*math.Sin(phi))
		r := [3]float64{n * math.Cos(phi) * math.Cos(lam), n * math.Cos(phi) * math.Sin(lam), n * (1 - e2) * math.Sin(phi)}
		u := [3]float64{math.Cos(phi) * math.Cos(lam), math.Cos(phi) * math.Sin(lam), math.Sin(phi)}
		dot, length := 0., 0.
		for i := range r {
			d := q.State.Target[i] - r[i]
			dot += d * u[i]
			length += d * d
		}
		alt := math.Asin(dot/math.Sqrt(length)) * 180 / math.Pi
		if got.Point.Time != got.Time || math.Abs(got.Point.TargetAltitude-alt)*3600 > 1 {
			t.Fatal(q.Label, "point time or altitude", got.Point, alt)
		}
		d, _ := transitInverse(got.Point.Longitude, got.Point.Latitude, q.State.Point[0], q.State.Point[1])
		maxPoint = math.Max(maxPoint, d*1000)
		if d > .01 {
			t.Fatal(q.Label, d)
		}
		dw := math.Abs(got.Point.HalfWidthKM/q.HalfWidth - 1)
		maxWidth = math.Max(maxWidth, dw)
		if dw > .01 {
			t.Fatal(q.Label, "width", dw)
		}
		for _, p := range got.Footprint {
			phi, lam := p.Latitude*math.Pi/180, p.Longitude*math.Pi/180
			f := 1 / 298.257223563
			e2 := f * (2 - f)
			n := 6378.137 / math.Sqrt(1-e2*math.Sin(phi)*math.Sin(phi))
			r := [3]float64{n * math.Cos(phi) * math.Cos(lam), n * math.Cos(phi) * math.Sin(lam), n * (1 - e2) * math.Sin(phi)}
			a, b := [3]float64{}, [3]float64{}
			for i := range a {
				a[i] = q.State.Satellite[i] - r[i]
				b[i] = q.State.Target[i] - r[i]
			}
			cross := [3]float64{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
			dot := a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
			sep := math.Atan2(math.Sqrt(cross[0]*cross[0]+cross[1]*cross[1]+cross[2]*cross[2]), dot)
			radius := 1737.4
			if q.Kind == TransitSun {
				radius = 695700
			}
			angular := math.Asin(radius / math.Sqrt(b[0]*b[0]+b[1]*b[1]+b[2]*b[2]))
			res := math.Abs(sep-angular) * 180 / math.Pi * 3600
			maxContact = math.Max(maxContact, res)
			if res > .5 {
				t.Fatal(q.Label, p, res)
			}
		}
		sub, err := GroundPointAt(e, q.State.Time)
		if err != nil || got.SubPoint != sub {
			t.Fatal("SubPoint", err)
		}
		threshold := (alt + 90) / 2
		hidden, err := TransitShadowAt(e, q.Kind, q.State.Time, TransitSearchOptions{MinTargetAltitude: threshold})
		if err != nil || !reflect.DeepEqual(hidden, TransitShadow{Time: q.State.Time}) {
			t.Fatal("altitude threshold", hidden, err)
		}
		if _, err = json.Marshal(got); err != nil {
			t.Fatal("non-finite snapshot", err)
		}
	}
	t.Logf("11 independent snapshots point=%g m contact=%g arcsec width-relative=%g", maxPoint, maxContact, maxWidth)
}

func TestTransitShadowInstanceConsistency(t *testing.T) {
	module := NewWithInstance(ephem.NewModelInstance(ephem.Semianalytic))
	for _, q := range readTransitEventFixtures(t).Shadows {
		e := nearbyElements(t, q.Label)
		at := q.State.Time.In(time.FixedZone("snapshot", 9*3600))
		a, err := TransitShadowAt(e, q.Kind, at, TransitSearchOptions{})
		if err != nil {
			t.Fatal(err)
		}
		b, err := module.TransitShadowAt(e, q.Kind, at, TransitSearchOptions{})
		if err != nil || !reflect.DeepEqual(a, b) || a.Time.Location() != at.Location() || a.Point.Time != at || a.SubPoint.Time != at {
			t.Fatal("instance or timezone", err)
		}
	}
}

func TestTransitShadowInputConsistency(t *testing.T) {
	g := readTransitEventFixtures(t).Groups[0]
	for _, q := range []struct {
		e      Elements
		target TransitTarget
		at     time.Time
		opt    TransitSearchOptions
		want   error
	}{
		{g.Elements, -1, g.Start, TransitSearchOptions{}, ErrInvalidInput},
		{g.Elements, TransitTarget(2), g.Start, TransitSearchOptions{}, ErrInvalidInput},
		{g.Elements, g.Target, time.Time{}, TransitSearchOptions{}, ErrInvalidInput},
		{g.Elements, g.Target, time.Date(2400, 1, 1, 0, 0, 0, 0, time.UTC), TransitSearchOptions{}, ErrInvalidInput},
		{Elements{}, g.Target, g.Start, TransitSearchOptions{}, ErrInvalidElements},
		{g.Elements, g.Target, g.Start, TransitSearchOptions{MinTargetAltitude: -1}, ErrInvalidInput},
		{g.Elements, g.Target, g.Start, TransitSearchOptions{MinTargetAltitude: 91}, ErrInvalidInput},
		{g.Elements, g.Target, g.Start, TransitSearchOptions{MinTargetAltitude: math.NaN()}, ErrInvalidInput},
		{g.Elements, g.Target, g.Start, TransitSearchOptions{MinTargetAltitude: math.Inf(1)}, ErrInvalidInput},
	} {
		if _, err := TransitShadowAt(q.e, q.target, q.at, q.opt); !errors.Is(err, q.want) {
			t.Fatal(q, err)
		}
	}
}
func TestTransitShadowInvalidConsistency(t *testing.T) {
	g := readTransitEventFixtures(t).Groups[0]
	at := g.States[0].Time
	got, err := TransitShadowAt(g.Elements, g.Target, at, TransitSearchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, TransitShadow{Time: at}) {
		t.Fatal("natural event endpoint is not an arbitrary snapshot tangent", got)
	}
}

func TestTransitShadowPartialLimbExternal(t *testing.T) {
	count, centres := 0, 0
	for _, g := range readTransitEventFixtures(t).Groups {
		if g.Minimum != 0 {
			continue
		}
		for _, q := range []transitOracleState{g.States[0], g.States[2]} {
			u := [3]float64{}
			distance := 0.
			for i := range u {
				u[i] = q.Target[i] - q.Satellite[i]
				distance += u[i] * u[i]
			}
			distance = math.Sqrt(distance)
			for i := range u {
				u[i] /= distance
			}
			p := math.Hypot(u[0], u[1])
			east := [3]float64{-u[1] / p, u[0] / p, 0}
			north := [3]float64{-u[2] * east[1], u[2] * east[0], u[0]*east[1] - u[1]*east[0]}
			radius := 1737.4
			if g.Target == TransitSun {
				radius = 695700
			}
			angle := math.Asin(radius / distance)
			hits := func(k [3]float64) bool {
				a, b, c := 0., 0., -1.
				for i, axis := range [3]float64{6378.137, 6378.137, 6378.137 * (1 - 1/298.257223563)} {
					x, v := q.Satellite[i]/axis, k[i]/axis
					a += v * v
					b += x * v
					c += x * x
				}
				return b < 0 && b*b-a*c >= 0
			}
			if hits([3]float64{-u[0], -u[1], -u[2]}) {
				centres++
			}
			landed := 0
			for i := 0; i < 72; i++ {
				pa := float64(i) * 2 * math.Pi / 72
				k := [3]float64{}
				for j := range k {
					k[j] = -math.Cos(angle)*u[j] - math.Sin(angle)*(math.Cos(pa)*east[j]+math.Sin(pa)*north[j])
				}
				if hits(k) {
					landed++
				}
			}
			if landed == 0 || landed == 72 {
				t.Fatal("fixture does not have a partially grounded limb", g.Label, landed)
			}
			got, err := TransitShadowAt(g.Elements, g.Target, q.Time, TransitSearchOptions{})
			if err != nil || !reflect.DeepEqual(got, TransitShadow{Time: q.Time}) {
				t.Fatal(g.Label, q.Time, landed, got, err)
			}
			count++
		}
	}
	if count == 0 || centres == 0 {
		t.Fatal("missing horizon or grounded-centre cases", count, centres)
	}
	t.Logf("%d independent partial-limb snapshots invalid, %d with a grounded centre", count, centres)
}
