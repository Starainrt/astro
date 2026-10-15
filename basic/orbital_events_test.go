// 被测契约：近日点与远日点交替出现且日心距为局部极值；过交点处日心黄纬为零、升降交点交替，并带该时刻的日心距。
package basic

import (
	"math"
	"testing"
)

func TestPlanetApsidesKnownDates(t *testing.T) {
	cases := []struct {
		body    SkyBody
		kind    OrbitalEventKind
		startTT float64
		wantTT  float64
		tolDays float64
	}{
		{SkyEarth, OrbitalPerihelion, 2451545, 2451547.2, 1.5},
		{SkyMars, OrbitalPerihelion, 2451545, 2452195, 3},
		{SkyJupiter, OrbitalPerihelion, 2459000, 2459965, 15},
	}
	for _, c := range cases {
		_, period, _ := orbitalIndex(c.body)
		var got float64
		for _, e := range PlanetApsidesWithInstance(c.body, c.startTT, c.startTT+1.1*period, nil) {
			if e.Kind == c.kind {
				got = e.JDE
				break
			}
		}
		if math.Abs(got-c.wantTT) > c.tolDays {
			t.Errorf("body %d kind %d: got %.2f want %.2f", c.body, c.kind, got, c.wantTT)
		}
	}
}

func TestPlanetApsidesAlternateAndNodesAreZeroLatitude(t *testing.T) {
	for _, body := range []SkyBody{SkyMercury, SkyVenus, SkyMars, SkyJupiter, SkySaturn} {
		index, period, _ := orbitalIndex(body)
		ap := PlanetApsidesWithInstance(body, 2451545, 2451545+3*period, nil)
		if len(ap) < 5 {
			t.Fatalf("body %d: only %d apsides", body, len(ap))
		}
		for i := 1; i < len(ap); i++ {
			if ap[i].Kind == ap[i-1].Kind {
				t.Errorf("body %d: apsides do not alternate at %d", body, i)
			}
		}
		nodes := PlanetNodesWithInstance(body, 2451545, 2451545+3*period, nil)
		if len(nodes) < 5 {
			t.Fatalf("body %d: only %d nodes", body, len(nodes))
		}
		for i, e := range nodes {
			if b := PlanetCoordinateNWithInstance(index, 1, e.JDE, -1, nil); math.Abs(b) > 1e-6 {
				t.Errorf("body %d node %d: latitude %.3g", body, i, b)
			}
			if want := PlanetCoordinateNWithInstance(index, 2, e.JDE, -1, nil); math.IsNaN(e.DistanceAU) || math.Abs(e.DistanceAU-want) > 1e-12 {
				t.Errorf("body %d node %d: distance %.9f want %.9f", body, i, e.DistanceAU, want)
			}
			if i > 0 && e.Kind == nodes[i-1].Kind {
				t.Errorf("body %d: nodes do not alternate at %d", body, i)
			}
		}
	}
	if PlanetNodesWithInstance(SkyEarth, 2451545, 2452545, nil) != nil || PlanetApsidesWithInstance(SkySun, 2451545, 2452545, nil) != nil {
		t.Error("unsupported body must return nil")
	}
}

func TestPlanetApsidesRejectsUnscannableRange(t *testing.T) {
	if got := PlanetApsidesWithInstance(SkyMars, -1e308, 1e308, nil); got != nil {
		t.Errorf("huge range: %d events", len(got))
	}
	if got := PlanetApsidesWithInstance(SkyMars, -1e15, 1e15, nil); got != nil {
		t.Errorf("over grid limit: %d events", len(got))
	}
	if got := PlanetApsidesWithInstance(SkyEarth, 0, 1e12, nil); got != nil {
		t.Errorf("over 32-bit grid limit: %d events", len(got))
	}
	if got := PlanetNodesWithInstance(SkyMars, 0, 1e12, nil); got != nil {
		t.Errorf("nodes over grid limit: %d events", len(got))
	}
	if got := PlanetNodesWithInstance(SkyMars, 0, 686.98/200*float64(skyGridMaxIntervals), nil); got != nil {
		t.Errorf("nodes at grid limit: %d events", len(got))
	}
}
