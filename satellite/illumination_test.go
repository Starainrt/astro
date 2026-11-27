package satellite

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

func TestIlluminationSynthetic(t *testing.T) {
	const earthAngle = 0.5
	const sunAngle = 0.02
	radius := 6378.137 / math.Sin(earthAngle)
	distance := 695700 / math.Sin(sunAngle)
	r := [3]float64{radius, 0, 0}
	for _, tc := range []struct {
		s       float64
		kind    Illumination
		visible bool
	}{
		{0, IlluminationUmbra, false},
		{earthAngle - sunAngle - 0.01, IlluminationUmbra, false},
		{earthAngle - 0.01, IlluminationPenumbra, false},
		{earthAngle + 0.01, IlluminationPenumbra, true},
		{earthAngle + sunAngle + 0.01, IlluminationSunlit, true},
		{math.Pi, IlluminationSunlit, true},
	} {
		sun := [3]float64{radius - distance*math.Cos(tc.s), distance * math.Sin(tc.s), 0}
		kind, visible := illuminationAt(r, sun)
		if kind != tc.kind || visible != tc.visible {
			t.Fatalf("angle=%g kind=%v visible=%v want=%v/%v", tc.s, kind, visible, tc.kind, tc.visible)
		}
	}
	if kind, visible := illuminationAt([3]float64{0, 0, 6378.1355}, [3]float64{0, 0, 1.5e8}); kind != IlluminationSunlit || !visible {
		t.Fatalf("low polar radius gave %v/%t, want sunlit/true", kind, visible)
	}
}

func TestIlluminationExternal(t *testing.T) {
	var maximum float64
	count := 0
	for _, g := range readSiteFixtures(t).Groups {
		if g.TransitionSource == "" {
			t.Fatal("missing provenance")
		}
		for _, r := range g.Transitions {
			if len(r) != 2 {
				t.Fatalf("invalid row %v", r)
			}
			var when time.Time
			var after bool
			if err := json.Unmarshal(r[0], &when); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(r[1], &after); err != nil {
				t.Fatal(err)
			}
			value := func(ti time.Time) bool {
				look, err := LookAt(g.Elements, Observer{}, ti, ObserveOptions{NoRefraction: true})
				if err != nil {
					t.Fatal(err)
				}
				return look.SunCenterVisible
			}
			lo, hi := when.Add(-time.Second), when.Add(time.Second)
			if value(lo) == after || value(hi) != after {
				t.Fatalf("%s missing illumination transition at %s", g.Label, when)
			}
			for hi.Sub(lo) > time.Millisecond {
				mid := lo.Add(hi.Sub(lo) / 2)
				if value(mid) == after {
					hi = mid
				} else {
					lo = mid
				}
			}
			own := lo.Add(hi.Sub(lo) / 2)
			diff := math.Abs(own.Sub(when).Seconds())
			maximum = math.Max(maximum, diff)
			if diff > 0.05 {
				t.Fatalf("%s transition=%s difference=%.12g s", g.Label, when, diff)
			}
			count++
		}
	}
	if count != 95 {
		t.Fatalf("transitions=%d", count)
	}
	t.Logf("SunCenterVisible transitions=%d maximum difference=%.12g s", count, maximum)
}

func TestMagnitudeSynthetic(t *testing.T) {
	const tolerance = 1e-12
	standard := 0.0
	for _, tc := range []struct{ phase, rangeKM, fraction, want float64 }{
		{0, 1000, 1, -1.2428746817353347},
		{math.Pi / 2, 1000, 1, 0},
		{math.Pi, 1000, 1, 39.77997287069612},
		{math.Pi / 2, 2000, 1, 1.505149978319906},
		{math.Pi / 2, 1000, .5, .752574989159953},
		{math.Pi / 2, 1000, .25, 1.505149978319906},
	} {
		got := satelliteMagnitude(tc.rangeKM, tc.phase, tc.fraction, &standard)
		if math.Abs(got-tc.want) > tolerance || !finite(got) {
			t.Fatalf("phase=%g fraction=%g magnitude=%.17g want=%.17g", tc.phase, tc.fraction, got, tc.want)
		}
	}
	if !math.IsNaN(satelliteMagnitude(1000, 0, 1, nil)) || !math.IsNaN(satelliteMagnitude(1000, 0, 0, &standard)) {
		t.Fatal("missing standard/umbra must be NaN")
	}
	for _, tc := range []struct{ earth, sun, separation, want float64 }{
		{1, 1, 3, 1}, {1, 1, 0, 0}, {2, 1, .5, 0}, {1, 2, .5, .75},
		{1, 1, 1, 0.6089977810442293},
	} {
		got := illuminatedFraction(tc.earth, tc.sun, tc.separation)
		if math.Abs(got-tc.want) > tolerance {
			t.Fatalf("fraction=%g want=%g", got, tc.want)
		}
	}
}
