package satellite

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
)

func TestTransitFramesConsistency(t *testing.T) {
	var maximumMoonDifference float64
	for _, g := range readTransitFixtures(t).Groups {
		events, err := Transits(g.Elements, g.Observer, g.Start, g.End, TransitOptions{Targets: []TransitTarget{g.Target}})
		if err != nil || len(events) != 1 {
			t.Fatal(err)
		}
		event := events[0]
		track, err := TransitFrames(g.Elements, g.Observer, event, 0)
		if err != nil {
			t.Fatal(err)
		}
		if len(track.Frames) < 200 || len(track.Frames) > 204 || !equalTransitBits(track.Transit, event) || track.Elements != g.Elements || track.Observer != g.Observer {
			t.Fatal(g.Label, len(track.Frames))
		}
		found := map[time.Time]bool{}
		for i, f := range track.Frames {
			for _, v := range []float64{f.XArcsec, f.YArcsec, f.TargetRadiusArcsec, f.SatelliteElevation, f.TargetAltitude} {
				if !finite(v) {
					t.Fatal(f)
				}
			}
			if i > 0 && f.Time.Sub(track.Frames[i-1].Time) < time.Microsecond {
				t.Fatal("unmerged grid", f)
			}
			for _, at := range []time.Time{event.Start, event.Center, event.End} {
				if f.Time == at {
					found[at] = true
				}
			}
			if f.Time == event.Center && (math.Float64bits(f.TargetRadiusArcsec) != math.Float64bits(event.TargetRadiusArcsec) || math.Float64bits(f.SatelliteElevation) != math.Float64bits(event.Satellite.Elevation) || math.Float64bits(f.TargetAltitude) != math.Float64bits(event.TargetAltitude)) {
				t.Fatal("center bits")
			}
		}
		if len(found) != 3 {
			t.Fatal(found)
		}
		inst := ephem.NewModelInstance(ephem.Semianalytic)
		b, err := NewWithInstance(inst).TransitFrames(g.Elements, g.Observer, event, 0)
		if err != nil || !reflect.DeepEqual(track.Frames, b.Frames) || track.MoonBrightLimbAngle != b.MoonBrightLimbAngle {
			for i := range track.Frames {
				if !reflect.DeepEqual(track.Frames[i], b.Frames[i]) {
					t.Logf("frame %d a=%+v b=%+v", i, track.Frames[i], b.Frames[i])
					break
				}
			}
			t.Fatal("instance", err, "limb", track.MoonBrightLimbAngle, b.MoonBrightLimbAngle)
		}
		if g.Target == TransitMoon {
			jd := basic.UTC2TT(basic.Date2JD(event.Center.UTC()))
			defaultMoon := basic.MoonPhaseWithInstance(jd, nil)
			explicitMoon := basic.MoonPhaseWithInstance(jd, inst)
			maximumMoonDifference = math.Max(maximumMoonDifference, math.Abs(defaultMoon-explicitMoon))
			if math.Float64bits(track.MoonIlluminatedFraction) != math.Float64bits(defaultMoon) || math.Float64bits(b.MoonIlluminatedFraction) != math.Float64bits(explicitMoon) || track.MoonBrightLimbAngle != basic.MoonTopocentricBrightLimbPositionAngleNWithInstance(jd, g.Observer.Longitude, g.Observer.Latitude, g.Observer.Height, -1, nil) {
				t.Fatal("moon bits")
			}
		} else if track.MoonIlluminatedFraction != 1 || track.MoonBrightLimbAngle != 0 {
			t.Fatal("sun")
		}
	}
	if maximumMoonDifference > 1e-7 {
		t.Fatal("moon fraction difference", maximumMoonDifference)
	}
	t.Logf("default/explicit MoonIlluminatedFraction maximum absolute difference=%g", maximumMoonDifference)
}

func TestTransitFramesExternal(t *testing.T) {
	data, err := os.ReadFile("testdata/skyfield_transits.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		TrackFrames []struct {
			GroupIndex int `json:"group_index"`
			Rows       [][]json.RawMessage
		} `json:"track_frames"`
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	groups := readTransitFixtures(t).Groups
	if len(fixture.TrackFrames) != len(groups) {
		t.Fatalf("reference groups=%d want=%d", len(fixture.TrackFrames), len(groups))
	}
	var maximum float64
	var maximumDelta time.Duration
	var matchedFrames, referenceFrames int
	for _, g := range fixture.TrackFrames {
		input := groups[g.GroupIndex]
		events, err := Transits(input.Elements, input.Observer, input.Start, input.End, TransitOptions{Targets: []TransitTarget{input.Target}})
		if err != nil || len(events) != 1 {
			t.Fatal(err)
		}
		track, err := TransitFrames(input.Elements, input.Observer, events[0], 0)
		if err != nil {
			t.Fatal(err)
		}
		matched := make([]bool, len(track.Frames))
		referenceFrames += len(g.Rows)
		for _, row := range g.Rows {
			var at time.Time
			var x, y float64
			if len(row) != 3 {
				t.Fatal("reference row")
			}
			if err = json.Unmarshal(row[0], &at); err != nil {
				t.Fatal(err)
			}
			json.Unmarshal(row[1], &x)
			json.Unmarshal(row[2], &y)
			nearest := -1
			var nearestDelta time.Duration
			tied := false
			for index, frame := range track.Frames {
				delta := frame.Time.Sub(at)
				if delta < 0 {
					delta = at.Sub(frame.Time)
				}
				if nearest < 0 || delta < nearestDelta {
					nearest, nearestDelta, tied = index, delta, false
				} else if delta == nearestDelta {
					tied = true
				}
			}
			if nearest < 0 || nearestDelta > time.Microsecond {
				t.Fatalf("%s %s missing frame within 1 µs (nearest index=%d delta=%s)", input.Label, at, nearest, nearestDelta)
			}
			if tied || matched[nearest] {
				t.Fatalf("%s %s nearest frame is tied or already matched (index=%d delta=%s)", input.Label, at, nearest, nearestDelta)
			}
			matched[nearest] = true
			matchedFrames++
			if nearestDelta > maximumDelta {
				maximumDelta = nearestDelta
			}
			frame := track.Frames[nearest]
			difference := math.Max(math.Abs(frame.XArcsec-x), math.Abs(frame.YArcsec-y))
			maximum = math.Max(maximum, difference)
			if difference > 0.5 {
				t.Fatalf("%s %s difference=%g arcsec", input.Label, at, difference)
			}
		}
	}
	if matchedFrames != referenceFrames {
		t.Fatalf("matched frames=%d want=%d", matchedFrames, referenceFrames)
	}
	t.Logf("%d independent frames maximum |Δt|=%s maximum X/Y difference=%.9g arcsec", matchedFrames, maximumDelta, maximum)
}

func BenchmarkTransitFramesISS(b *testing.B) {
	data, err := os.ReadFile("testdata/skyfield_transits.json")
	if err != nil {
		b.Fatal(err)
	}
	var fixture transitFixtures
	if err = json.Unmarshal(data, &fixture); err != nil {
		b.Fatal(err)
	}
	g := fixture.Groups[0]
	events, err := Transits(g.Elements, g.Observer, g.Start, g.End, TransitOptions{Targets: []TransitTarget{g.Target}})
	if err != nil || len(events) != 1 {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := TransitFrames(g.Elements, g.Observer, events[0], 0); err != nil {
			b.Fatal(err)
		}
	}
}
func TestTransitFramesInputSynthetic(t *testing.T) {
	g := readTransitFixtures(t).Groups[0]
	e, err := Transits(g.Elements, g.Observer, g.Start, g.End, TransitOptions{Targets: []TransitTarget{g.Target}})
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range []time.Duration{-1, time.Nanosecond} {
		if _, err := TransitFrames(g.Elements, g.Observer, e[0], step); !errors.Is(err, ErrInvalidInput) {
			t.Fatal(step, err)
		}
	}
	for _, tr := range []Transit{{}, func() Transit { tr := e[0]; tr.Target = TransitTarget(2); return tr }()} {
		if _, err := TransitFrames(g.Elements, g.Observer, tr, 0); !errors.Is(err, ErrInvalidInput) {
			t.Fatal(err)
		}
	}
	tr := e[0]
	tr.NearMiss = true
	tr.Start = time.Time{}
	tr.End = time.Time{}
	tr.Duration = 0
	got, err := TransitFrames(g.Elements, g.Observer, tr, 10*time.Millisecond)
	if err != nil || len(got.Frames) < 3 {
		t.Fatal(err)
	}
	found := false
	for _, f := range got.Frames {
		found = found || f.Time == tr.Center
	}
	if !found {
		t.Fatal("missing near-miss center")
	}
}
