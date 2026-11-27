package svg

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/satellite"
)

type localFixtureGroup struct {
	Elements   satellite.Elements
	Observer   satellite.Observer
	Target     satellite.TransitTarget
	Start, End time.Time
}

func localGroups(t testing.TB) []localFixtureGroup {
	t.Helper()
	data, err := os.ReadFile("../testdata/skyfield_transits.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct{ Groups []localFixtureGroup }
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture.Groups
}
func localTracks(t testing.TB) []satellite.TransitTrack {
	t.Helper()
	groups := localGroups(t)
	var result []satellite.TransitTrack
	for _, i := range []int{0, 1, 0} {
		g := groups[i]
		opt := satellite.TransitOptions{Targets: []satellite.TransitTarget{g.Target}}
		if len(result) == 2 {
			g.Observer.Latitude += .05
			opt.NearMissArcsec = 7200
		}
		events, err := satellite.Transits(g.Elements, g.Observer, g.Start, g.End, opt)
		if err != nil || len(events) != 1 {
			t.Fatalf("events: %d %v", len(events), err)
		}
		track, err := satellite.TransitFrames(g.Elements, g.Observer, events[0], 0)
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, track)
	}
	if !result[2].Transit.NearMiss {
		t.Fatal("expected a physical near miss")
	}
	return result
}

func TestTransitLocalSVGSynthetic(t *testing.T) {
	for i, track := range localTracks(t) {
		value, err := TransitLocalSVG(track, TransitLocalSVGOptions{})
		if err != nil {
			t.Fatal(err)
		}
		nodes := parseDiagram(t, value)
		roles := map[string]int{}
		var scale, cx, cy float64
		for _, n := range nodes {
			role := n.Attr["data-role"]
			roles[role]++
			switch role {
			case "event-start", "event-center", "event-end", "stage-satellite":
				if n.Attr["stroke"] != "#ffffff" || number(n.Attr["stroke-width"]) <= 0 {
					t.Fatalf("%s missing contrast outline: %+v", role, n.Attr)
				}
			}
			if role == "overview" {
				scale = number(n.Attr["data-scale"])
				cx = number(n.Attr["data-cx"])
				cy = number(n.Attr["data-cy"])
			}
			if role != "track" {
				continue
			}
			pairs := strings.Fields(n.Attr["points"])
			if len(pairs) != len(track.Frames) {
				t.Fatal(len(pairs))
			}
			for j, pair := range pairs {
				v := strings.Split(pair, ",")
				x, y := (cx-number(v[0]))/scale, (cy-number(v[1]))/scale
				if math.Abs(x-track.Frames[j].XArcsec) > .01 || math.Abs(y-track.Frames[j].YArcsec) > .01 {
					t.Fatal("projection", j, x, y)
				}
			}
		}
		for _, role := range []string{"target-disk", "track", "tick", "event-center", "stage-panel"} {
			if roles[role] == 0 {
				t.Fatal(role)
			}
		}
		if i == 2 {
			if roles["stage-panel"] != 1 || roles["event-start"] != 0 || roles["event-end"] != 0 {
				t.Fatal(roles)
			}
		} else if roles["stage-panel"] != 3 || roles["event-start"] == 0 || roles["event-end"] == 0 {
			t.Fatal(roles)
		}
	}
}

func TestTransitLocalSVGLayoutConsistency(t *testing.T) {
	for i, track := range localTracks(t) {
		for _, lang := range []string{"zh", "en"} {
			for _, s := range [][2]int{{920, 720}, {640, 520}} {
				t.Run(fmt.Sprintf("%d/%s/%d", i, lang, s[0]), func(t *testing.T) {
					value, err := TransitLocalSVG(track, TransitLocalSVGOptions{Language: lang, Width: s[0], Height: s[1]})
					if err != nil {
						t.Fatal(err)
					}
					checkTextBoxes(t, parseDiagram(t, value), float64(s[0]), float64(s[1]))
				})
			}
		}
	}
}

func TestTransitLocalSVGExpandsForWrappedTitle(t *testing.T) {
	track := localTracks(t)[1]
	title := strings.TrimSpace(strings.Repeat("ISS lunar transit, observing station in New Zealand.\n", 8))
	document, err := TransitLocalSVG(track, TransitLocalSVGOptions{Width: 640, Height: 520, Language: "en", Title: title})
	if err != nil {
		t.Fatal(err)
	}
	nodes := parseDiagram(t, document)
	height := number(nodes[0].Attr["height"])
	checkTextBoxes(t, nodes, 640, height)
	var radius, dataBottom float64
	for _, node := range nodes {
		switch node.Attr["data-role"] {
		case "target-disk":
			if radius == 0 {
				radius = number(node.Attr["data-radius"])
			}
		case "contact", "statistic":
			dataBottom = math.Max(dataBottom, number(node.Attr["y"]))
		}
	}
	if height <= 520 || radius < 50 || dataBottom > height-50 {
		t.Fatalf("clipped local chart: height %.1f, radius %.1f, last data baseline %.1f", height, radius, dataBottom)
	}
}

func TestTransitLocalSVGTrackLabelsConsistency(t *testing.T) {
	for i, original := range localTracks(t) {
		for _, angle := range []float64{0, math.Pi / 3, math.Pi / 2, math.Pi} {
			track := original
			track.Frames = append([]satellite.TransitFrame(nil), original.Frames...)
			for j, f := range track.Frames {
				track.Frames[j].XArcsec = f.XArcsec*math.Cos(angle) - f.YArcsec*math.Sin(angle)
				track.Frames[j].YArcsec = f.XArcsec*math.Sin(angle) + f.YArcsec*math.Cos(angle)
			}
			for _, size := range [][2]int{{920, 720}, {640, 520}} {
				for _, language := range []string{"zh", "en"} {
					value, err := TransitLocalSVG(track, TransitLocalSVGOptions{Width: size[0], Height: size[1], Language: language})
					if err != nil {
						t.Fatal(err)
					}
					nodes := parseDiagram(t, value)
					checkTextBoxes(t, nodes, float64(size[0]), float64(size[1]))
					var x0, y0, dx, dy float64
					for _, n := range nodes {
						if n.Attr["data-role"] == "track" {
							pairs := strings.Fields(n.Attr["points"])
							a, b := strings.Split(pairs[0], ","), strings.Split(pairs[len(pairs)-1], ",")
							x0, y0 = number(a[0]), number(a[1])
							dx, dy = number(b[0])-x0, number(b[1])-y0
						}
					}
					length := math.Hypot(dx, dy)
					timeSide, eventSide, count, eventCount := 0., 0., 0, 0
					for _, n := range nodes {
						if n.Attr["data-role"] != "track-label" {
							continue
						}
						x, y := number(n.Attr["x"]), number(n.Attr["y"])-number(n.Attr["font-size"])*.3
						side := ((x-x0)*dy - (y-y0)*dx) / length
						if strings.HasSuffix(n.Text, " s") {
							count++
							if math.Abs(side) < 12 || math.Abs(side) > 40 || timeSide*side < 0 {
								t.Fatalf("track %d angle %.2f: scattered time label %s (offset %.2f)", i, angle, n.Text, side)
							}
							timeSide = side
						} else {
							eventCount++
							if math.Abs(side) < 12 || eventSide*side < 0 {
								t.Fatalf("scattered event label %s: %.2f", n.Text, side)
							}
							eventSide = side
						}
					}
					expectedEvents := 3
					if track.Transit.NearMiss {
						expectedEvents = 1
					}
					minimumTicks := 1
					if size[0] == 920 && !track.Transit.NearMiss {
						minimumTicks = 2
					}
					if count < minimumTicks || count > 6 || eventCount != expectedEvents || timeSide*eventSide >= 0 {
						t.Fatalf("track %d angle %.2f size %d %s: time/event labels %d/%d; sides %.2f/%.2f", i, angle, size[0], language, count, eventCount, timeSide, eventSide)
					}
				}
			}
		}
	}
}

func TestTransitLocalSVGContactsConsistency(t *testing.T) {
	for _, g := range localGroups(t) {
		events, err := satellite.Transits(g.Elements, g.Observer, g.Start, g.End, satellite.TransitOptions{Targets: []satellite.TransitTarget{g.Target}})
		if err != nil || len(events) != 1 {
			t.Fatal(err)
		}
		track, err := satellite.TransitFrames(g.Elements, g.Observer, events[0], 0)
		if err != nil {
			t.Fatal(err)
		}
		value, err := TransitLocalSVG(track, TransitLocalSVGOptions{Language: "en"})
		if err != nil {
			t.Fatal(err)
		}
		j := 0
		for _, n := range parseDiagram(t, value) {
			if !strings.HasPrefix(n.Text, "Az ") {
				continue
			}
			fields := strings.Fields(n.Text)
			az, err := strconv.ParseFloat(strings.TrimSuffix(fields[1], "°"), 64)
			if err != nil {
				t.Fatal(err)
			}
			at := []time.Time{track.Transit.Start, track.Transit.Center, track.Transit.End}[j]
			look, err := satellite.LookAt(track.Elements, track.Observer, at, satellite.ObserveOptions{NoRefraction: true})
			if err != nil {
				t.Fatal(err)
			}
			if d := math.Abs(math.Remainder(az-look.Azimuth, 360)); d > .05001 {
				t.Fatalf("azimuth %d: %.9f vs %.9f", j, az, look.Azimuth)
			}
			j++
		}
		if j != 3 {
			t.Fatal("missing contacts", j)
		}
	}
}

func TestTransitLocalSVGInvalidSynthetic(t *testing.T) {
	track := localTracks(t)[0]
	for _, o := range []TransitLocalSVGOptions{{Width: 639}, {Height: 519}} {
		if _, err := TransitLocalSVG(track, o); !errors.Is(err, satellite.ErrInvalidInput) {
			t.Fatal(err)
		}
	}
	if _, err := TransitLocalSVG(satellite.TransitTrack{}, TransitLocalSVGOptions{}); !errors.Is(err, satellite.ErrInvalidInput) {
		t.Fatal(err)
	}
}

func TestTransitLocalSVGMoonPhaseSynthetic(t *testing.T) {
	track := localTracks(t)[1]
	for _, illumination := range []float64{0, .25, .5, .75, 1} {
		for _, angle := range []float64{0, 90, 180, 270} {
			track.MoonIlluminatedFraction = illumination
			track.MoonBrightLimbAngle = angle
			value, err := TransitLocalSVG(track, TransitLocalSVGOptions{})
			if err != nil {
				t.Fatal(err)
			}
			found := 0
			for _, node := range parseDiagram(t, value) {
				if node.Name != "path" || !strings.HasPrefix(node.Attr["transform"], "translate(") {
					continue
				}
				var cy, rx, ry, end, tx, ty, last float64
				var sweep int
				if _, err := fmt.Sscanf(node.Attr["d"], "M 0 %f A %f %f 0 0 1 0 %f A %f %f 0 0 %d 0 %f Z", &cy, &rx, &ry, &end, &tx, &ty, &sweep, &last); err != nil {
					t.Fatal(err)
				}
				radius := ry
				phase := (2*illumination - 1)
				if math.Abs(tx-radius*math.Abs(phase)) > .0011 || math.Abs(cy+radius) > .0011 || math.Abs(end-radius) > .0011 || math.Abs(last+radius) > .0011 || math.Abs(ty-radius) > .0011 {
					t.Fatal("terminator geometry", node.Attr)
				}
				wantSweep := 1
				if phase < 0 {
					wantSweep = 0
				}
				if sweep != wantSweep {
					t.Fatal("wrong bright side")
				}
				var x, y, rotation float64
				if _, err := fmt.Sscanf(node.Attr["transform"], "translate(%f %f) rotate(%f)", &x, &y, &rotation); err != nil {
					t.Fatal(err)
				}
				if math.Abs(rotation+angle+90) > 1e-9 {
					t.Fatal("limb orientation", rotation, angle)
				}
				found++
			}
			if found != 4 {
				t.Fatal("missing phase disks", found)
			}
		}
	}
}

func TestTransitLocalSVGNonfiniteSynthetic(t *testing.T) {
	track := localTracks(t)[0]
	for _, change := range []func(*satellite.TransitTrack){
		func(v *satellite.TransitTrack) { v.Transit.MinSeparationArcsec = math.NaN() },
		func(v *satellite.TransitTrack) { v.Transit.Satellite.RangeKM = math.Inf(1) },
		func(v *satellite.TransitTrack) { v.Transit.Satellite.Azimuth = math.NaN() },
	} {
		v := track
		change(&v)
		if _, err := TransitLocalSVG(v, TransitLocalSVGOptions{}); !errors.Is(err, satellite.ErrInvalidInput) {
			t.Fatal("nonfinite input accepted", err)
		}
	}
	track.Transit.CenterlineOffsetKM = math.NaN()
	value, err := TransitLocalSVG(track, TransitLocalSVGOptions{})
	if err != nil {
		t.Fatal(err)
	}
	parseDiagram(t, value)
}

func TestTransitLocalSVGStageConsistency(t *testing.T) {
	var maximum float64
	for _, track := range localTracks(t) {
		value, err := TransitLocalSVG(track, TransitLocalSVGOptions{})
		if err != nil {
			t.Fatal(err)
		}
		var cx, cy, scale float64
		var at time.Time
		for _, n := range parseDiagram(t, value) {
			switch n.Attr["data-role"] {
			case "stage-panel":
				cx, cy, scale = number(n.Attr["data-cx"]), number(n.Attr["data-cy"]), number(n.Attr["data-scale"])
				at, err = time.Parse(time.RFC3339Nano, n.Attr["data-time"])
				if err != nil {
					t.Fatal(err)
				}
			case "stage-satellite":
				var want satellite.TransitFrame
				for _, f := range track.Frames {
					if f.Time.Equal(at) {
						want = f
						break
					}
				}
				d := math.Max(math.Abs((cx-number(n.Attr["cx"]))/scale-want.XArcsec), math.Abs((cy-number(n.Attr["cy"]))/scale-want.YArcsec))
				maximum = math.Max(maximum, d)
				if d > .01 {
					t.Fatal("stage projection", d)
				}
			}
		}
	}
	t.Logf("maximum independent stage inversion %.9g arcsec", maximum)
}

func TestTransitLocalSVGBaselineConsistency(t *testing.T) {
	if runtime.GOARCH == "386" && strings.HasPrefix(runtime.Version(), "go1.16") {
		t.Skip("platform byte baseline")
	}
	expected := []struct {
		length int
		digest string
	}{{15359, "8f021d2811d8e308c432233ef94c6e98"}, {28382, "bd3a29f4ff41e77b63a5abaf3bd2e8cd"}, {11451, "33f7cb2871ab8cc306bd7e39f16c9bac"}}
	tracks := localTracks(t)
	names := []string{"local-sun-zh", "local-moon-en", "local-nearmiss-zh"}
	for i, track := range tracks {
		o := TransitLocalSVGOptions{}
		if i == 1 {
			o.Language = "en"
		}
		value, err := TransitLocalSVG(track, o)
		if err != nil {
			t.Fatal(err)
		}
		sum := fmt.Sprintf("%x", md5.Sum([]byte(value)))
		t.Log(names[i], len(value), sum)
		if len(value) != expected[i].length || sum != expected[i].digest {
			t.Fatalf("%s byte baseline: %d %s", names[i], len(value), sum)
		}
		if dir := os.Getenv("ASTRO_C6_SVG_DIR"); dir != "" {
			if err := os.WriteFile(filepath.Join(dir, names[i]+".svg"), []byte(value), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func BenchmarkTransitLocalSVG(b *testing.B) {
	track := localTracks(b)[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := TransitLocalSVG(track, TransitLocalSVGOptions{}); err != nil {
			b.Fatal(err)
		}
	}
}

func TestLocalAzimuthDegenerateEquatorialHorizonSynthetic(t *testing.T) {
	track := satellite.TransitTrack{
		Observer: satellite.Observer{Longitude: 0, Latitude: 0},
		Transit:  satellite.Transit{Satellite: satellite.Look{Azimuth: 90, Elevation: 0}},
	}
	frame := satellite.TransitFrame{XArcsec: 0, YArcsec: 10, TargetAltitude: 0, SatelliteElevation: 0}
	az := localAzimuth(track, frame)
	if math.IsNaN(az) || math.IsInf(az, 0) {
		t.Fatalf("degenerate localAzimuth = %v", az)
	}
}
