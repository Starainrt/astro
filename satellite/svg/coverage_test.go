package svg

import (
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/satellite"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

func coverageSynthetic() satellite.CoverageSweep {
	at := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	box := func(a, b, c, d float64) []satellite.GeoPoint {
		return []satellite.GeoPoint{{Longitude: a, Latitude: b}, {Longitude: c, Latitude: b}, {Longitude: c, Latitude: d}, {Longitude: a, Latitude: d}, {Longitude: a, Latitude: b}}
	}
	s := satellite.CoverageSweep{Elements: satellite.Elements{Name: "ISS", CatalogNumber: 25544, Epoch: at.Add(-time.Hour)}, Start: at, End: at.Add(2 * time.Minute), Swept: []satellite.Region{{Outer: box(-20, -20, 40, 40), Holes: [][]satellite.GeoPoint{box(0, 0, 5, 5)}}}, VisibleSwept: []satellite.Region{{Outer: box(10, 5, 25, 20)}}}
	for i := 0; i < 3; i++ {
		p := satellite.GroundPoint{Time: at.Add(time.Duration(i) * time.Minute), Longitude: 20 + float64(i), Latitude: 10 + float64(i), HeightKM: 400, SpeedKMS: 7}
		s.Track = append(s.Track, p)
		s.TrackIllumination = append(s.TrackIllumination, satellite.Illumination(i))
		night := []satellite.GeoPoint{{Longitude: -80, Latitude: -60}, {Longitude: -80, Latitude: 60}, {Longitude: 80, Latitude: 60}, {Longitude: 80, Latitude: -60}, {Longitude: -80, Latitude: -60}}
		s.Instants = append(s.Instants, satellite.Coverage{Time: p.Time, SubPoint: p, Footprint: box(p.Longitude-2, p.Latitude-2, p.Longitude+2, p.Latitude+2), Night: night, Visible: [][]satellite.GeoPoint{box(p.Longitude-1, p.Latitude-1, p.Longitude+1, p.Latitude+1)}})
	}
	return s
}
func TestCoverageSVGSynthetic(t *testing.T) {
	s := coverageSynthetic()
	value, err := CoverageSVG(s, CoverageSVGOptions{Highlight: s.Instants[1].Time})
	if err != nil {
		t.Fatal(err)
	}
	nodes := parseDiagram(t, value)
	roles := map[string]int{}
	for _, n := range nodes {
		roles[n.Attr["data-role"]]++
	}
	for _, r := range []string{"coverage-sweep", "visible-sweep", "ground-track", "night-region", "satellite-footprint", "satellite-visible", "subsatellite-point", "panel-value", "footer"} {
		if roles[r] == 0 {
			t.Fatal(r)
		}
	}
}

func TestCoverageSVGPaletteSynthetic(t *testing.T) {
	s := coverageSynthetic()
	observer := satellite.Observer{Longitude: 21, Latitude: 11}
	value, err := CoverageSVG(s, CoverageSVGOptions{Highlight: s.Instants[1].Time, Observer: &observer})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"coverage-sweep": "0.28", "visible-sweep": "0.45", "satellite-visible": "0.55", "coverage-sweep-legend": "0.28", "visible-sweep-legend": "0.45", "satellite-visible-legend": "0.55"}
	seen := map[string]bool{}
	strokes := map[string]string{}
	for _, n := range parseDiagram(t, value) {
		role := n.Attr["data-role"]
		if n.Attr["stroke"] != "" {
			strokes[role] = n.Attr["stroke"]
		}
		if opacity, ok := want[role]; ok {
			if n.Attr["fill"] != "#e0ae43" || n.Attr["fill-opacity"] != opacity {
				t.Errorf("%s fill=%s opacity=%s", role, n.Attr["fill"], n.Attr["fill-opacity"])
			}
			seen[role] = true
		}
		if role == "coverage-sweep-boundary" && (n.Attr["stroke"] != "#a66f18" || number(n.Attr["stroke-width"]) < 1.2) {
			t.Errorf("sweep boundary=%v", n.Attr)
		}
		if role == "observer" || role == "subsatellite-point" {
			if n.Name != "circle" || n.Attr["stroke"] != "white" || n.Attr["fill"] != "#2d3e5a" {
				t.Errorf("marker=%v", n)
			}
			seen[role] = true
		}
		if n.Attr["fill"] == "#4080cf" || n.Attr["fill"] == "#329b88" {
			t.Errorf("blue/teal coverage fill=%v", n.Attr)
		}
	}
	for role := range want {
		if !seen[role] {
			t.Error("missing", role)
		}
	}
	if current := strokes["satellite-visible-boundary"]; current == "" || current == strokes["visible-sweep-boundary"] || current != strokes["satellite-visible-legend"] {
		t.Errorf("current visual coverage must have a distinct stroke shared by its legend: %v", strokes)
	}
	if !seen["observer"] || !seen["subsatellite-point"] || !strings.Contains(value, `stroke-linecap="round" stroke-linejoin="round"`) {
		t.Error("missing outlined markers or rounded track joins")
	}
}

func TestCoverageSVGWithoutArea(t *testing.T) {
	sweep := coverageSynthetic()
	sweep.Swept, sweep.VisibleSwept = nil, nil
	for index := range sweep.Instants {
		sweep.Instants[index].Footprint, sweep.Instants[index].Visible = nil, nil
	}
	document, err := CoverageSVG(sweep, CoverageSVGOptions{Highlight: sweep.Instants[1].Time})
	if err != nil {
		t.Fatal(err)
	}
	roles := map[string]int{}
	for _, node := range parseDiagram(t, document) {
		roles[node.Attr["data-role"]]++
	}
	if roles["coverage-sweep"]+roles["visible-sweep"]+roles["satellite-footprint"]+roles["satellite-visible"] != 0 || roles["subsatellite-point"] == 0 || roles["ground-track"] == 0 {
		t.Fatalf("expected track and subpoint without polygons: %v", roles)
	}
}

func TestCoverageSVGCurrentFootprintSynthetic(t *testing.T) {
	s := coverageSynthetic()
	s.Instants[1].Visible = nil
	s.TrackIllumination[1] = satellite.IlluminationUmbra
	for _, language := range []string{"zh", "en"} {
		for _, size := range [][2]int{{1200, 800}, {640, 480}} {
			value, err := CoverageSVG(s, CoverageSVGOptions{Width: size[0], Height: size[1], Language: language, Highlight: s.Track[1].Time})
			if err != nil {
				t.Fatal(err)
			}
			seen := map[string]bool{}
			for _, n := range parseDiagram(t, value) {
				role := n.Attr["data-role"]
				switch role {
				case "satellite-footprint":
					if n.Attr["fill"] != "none" {
						t.Errorf("current geometric footprint must be unfilled: %v", n.Attr)
					}
				case "satellite-footprint-boundary", "satellite-footprint-legend":
					if n.Attr["fill"] != "none" || n.Attr["stroke-dasharray"] == "" || n.Attr["stroke"] != "#a66f18" {
						t.Errorf("current geometric footprint must use matching dashed outlines: %v", n.Attr)
					}
				case "satellite-visible":
					t.Error("eclipsed satellite has no current visual coverage")
				}
				seen[role] = true
			}
			label := "当前几何可见圈"
			if language == "en" {
				label = "Current geometric footprint"
			}
			if !seen["satellite-footprint"] || !seen["satellite-footprint-boundary"] || !seen["satellite-footprint-legend"] || !strings.Contains(value, label) {
				t.Errorf("language=%s size=%v missing footprint or legend", language, size)
			}
			checkTextBoxes(t, parseDiagram(t, value), float64(size[0]), float64(size[1]))
		}
	}
}

func TestCoverageSVGReferencePoint(t *testing.T) {
	s := coverageSynthetic()
	point := s.Track[1]
	for _, highlight := range []time.Time{{}, point.Time} {
		value, err := CoverageSVG(s, CoverageSVGOptions{Projection: MapProjectionOrthographic, Highlight: highlight})
		if err != nil {
			t.Fatal(err)
		}
		var frame, marker, leader map[string]string
		var label string
		for _, node := range parseDiagram(t, value) {
			switch node.Attr["data-role"] {
			case "map":
				frame = node.Attr
			case "subsatellite-point":
				marker = node.Attr
			case "time-label-leader":
				if node.Attr["data-time"] == point.Time.Format(time.RFC3339Nano) {
					leader = node.Attr
				}
			case "time-label":
				if node.Text == point.Time.Format("15:04:05") {
					label = node.Text
				}
			case "observer", "satellite-footprint":
				if highlight.IsZero() {
					t.Fatalf("unexpected layer %s", node.Attr["data-role"])
				}
			}
		}
		if marker["data-time"] != point.Time.Format(time.RFC3339Nano) || label == "" || leader["x1"] != marker["cx"] || leader["y1"] != marker["cy"] {
			t.Fatalf("reference point and time not linked: marker=%v label=%q leader=%v", marker, label, leader)
		}
		lon, lat := inverseMap(frame, number(marker["cx"]), number(marker["cy"]))
		if math.Hypot(lon-point.Longitude, lat-point.Latitude) > 1e-3 {
			t.Fatalf("reference position = %.6f°, %.6f°", lon, lat)
		}
	}
}

func TestCoverageSVGTimeHeaderConsistency(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		location, inputLocation *time.Location
	}{
		{time.UTC, time.UTC},
		{time.FixedZone("UTC+08:00", 8*3600), time.UTC},
		{ny, time.UTC},
		{nil, time.FixedZone("UTC+14:00", 14*3600)},
	} {
		for _, language := range []string{"zh", "en"} {
			for _, size := range [][2]int{{1200, 800}, {640, 480}} {
				s := coverageSynthetic()
				at := time.Date(2026, 11, 14, 23, 59, 30, 0, time.UTC).In(tc.inputLocation)
				delta := at.Sub(s.Start)
				s.Start, s.End = at, s.End.Add(delta).In(tc.inputLocation)
				for i := range s.Track {
					s.Track[i].Time = s.Track[i].Time.Add(delta).In(tc.inputLocation)
					s.Instants[i].Time = s.Track[i].Time
					s.Instants[i].SubPoint.Time = s.Track[i].Time
				}
				loc := tc.location
				if loc == nil {
					loc = tc.inputLocation
				}
				value, err := CoverageSVG(s, CoverageSVGOptions{Width: size[0], Height: size[1], Language: language, Location: tc.location, Highlight: s.Track[1].Time, TimeLabelStep: time.Minute, Title: "Coverage demo"})
				if err != nil {
					t.Fatal(err)
				}
				nodes := parseDiagram(t, value)
				var header, panel strings.Builder
				labels := 0
				for _, n := range nodes {
					switch n.Attr["data-role"] {
					case "header":
						header.WriteString(n.Text + " ")
					case "panel-label", "panel-value":
						panel.WriteString(n.Text + " ")
					case "time-label":
						if _, err := time.Parse("15:04:05", n.Text); err != nil {
							t.Errorf("time zone must be in table header, label=%q", n.Text)
						}
						labels++
					}
				}
				interval, heading, current := "扫掠时段", "时刻", "当前圈"
				if language == "en" {
					interval, heading, current = "Sweep interval", "Times", "Current footprint"
				}
				for _, want := range []string{interval, s.Start.In(loc).Format("2006-01-02 15:04:05"), s.End.In(loc).Format("15:04:05"), "(" + loc.String() + ")"} {
					if !strings.Contains(header.String(), want) {
						t.Errorf("header missing %q: %s", want, header.String())
					}
				}
				for _, want := range []string{heading + " (" + loc.String() + ")", current + " " + s.Track[1].Time.In(loc).Format("01-02 15:04:05")} {
					if !strings.Contains(panel.String(), want) {
						t.Errorf("time table missing %q: %s", want, panel.String())
					}
				}
				if s.Start.In(loc).Format("2006-01-02") != s.End.In(loc).Format("2006-01-02") && !strings.Contains(header.String(), s.End.In(loc).Format("2006-01-02")) {
					t.Error("missing interval end date")
				}
				if labels == 0 {
					t.Error("missing time labels")
				}
				checkTextBoxes(t, nodes, float64(size[0]), float64(size[1]))
			}
		}
	}
}

func TestCoverageSVGSweepTimeOffsetChange(t *testing.T) {
	paris, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Fatal(err)
	}
	sweep := coverageSynthetic()
	sweep.Start = time.Date(2026, 10, 25, 0, 45, 0, 0, time.UTC)
	sweep.End = sweep.Start.Add(30 * time.Minute)
	for index := range sweep.Track {
		at := sweep.Start.Add(time.Duration(index) * 15 * time.Minute)
		sweep.Track[index].Time = at
		sweep.Instants[index].Time = at
		sweep.Instants[index].SubPoint.Time = at
	}
	document, err := CoverageSVG(sweep, CoverageSVGOptions{Language: "en", Location: paris, TimeLabelStep: 15 * time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	matches := map[string]int{}
	for _, node := range parseDiagram(t, document) {
		switch node.Attr["data-role"] {
		case "header", "panel-value":
			for _, want := range []string{"02:45:00 +02:00", "02:15:00 +01:00"} {
				if strings.Contains(node.Text, want) {
					matches[want]++
				}
			}
		}
	}
	if matches["02:45:00 +02:00"] != 2 || matches["02:15:00 +01:00"] != 2 {
		t.Fatalf("header and panel must both disambiguate the offset: %v", matches)
	}
}

func TestCoverageSVGLabelClearanceSynthetic(t *testing.T) {
	for _, projection := range []MapProjection{MapProjectionEquirectangular, MapProjectionOrthographic} {
		s := coverageSynthetic()
		value, err := CoverageSVG(s, CoverageSVGOptions{Projection: projection, Highlight: s.Track[1].Time, TimeLabelStep: time.Minute})
		if err != nil {
			t.Fatal(err)
		}
		nodes := parseDiagram(t, value)
		count := 0
		for _, label := range nodes {
			if label.Attr["data-role"] != "time-label" {
				continue
			}
			count++
			box := svgchart.LabelTextBox(number(label.Attr["x"]), number(label.Attr["y"]), number(label.Attr["font-size"]), label.Text, label.Attr["text-anchor"])
			for _, line := range nodes {
				role := line.Attr["data-role"]
				if role != "ground-track" && !strings.HasSuffix(role, "-boundary") {
					continue
				}
				points := strings.Fields(line.Attr["points"])
				for i := 1; i < len(points); i++ {
					a, b := strings.Split(points[i-1], ","), strings.Split(points[i], ",")
					x, y := number(a[0]), number(a[1])
					dx, dy := number(b[0])-x, number(b[1])-y
					steps := int(math.Ceil(math.Hypot(dx, dy))) + 1
					for j := 0; j <= steps; j++ {
						px, py := x+dx*float64(j)/float64(steps), y+dy*float64(j)/float64(steps)
						if px >= box.X-2 && px <= box.X+box.Width+2 && py >= box.Y-2 && py <= box.Y+box.Height+2 {
							t.Fatalf("projection=%v time label %s overlaps %s", projection, label.Text, role)
						}
					}
				}
			}
		}
		if count == 0 {
			t.Fatal(projection, "missing time labels")
		}
	}
}

func TestCoverageSVGSeamSynthetic(t *testing.T) {
	s := coverageSynthetic()
	east := []satellite.GeoPoint{{Longitude: 170, Latitude: 0}, {Longitude: 180, Latitude: 0}, {Longitude: 180, Latitude: 20}, {Longitude: 170, Latitude: 20}, {Longitude: 170, Latitude: 0}}
	west := []satellite.GeoPoint{{Longitude: -180, Latitude: 0}, {Longitude: -170, Latitude: 0}, {Longitude: -170, Latitude: 10}, {Longitude: -180, Latitude: 10}, {Longitude: -180, Latitude: 0}}
	s.Swept = []satellite.Region{{Outer: east}, {Outer: west}}
	s.VisibleSwept = nil
	value, err := CoverageSVG(s, CoverageSVGOptions{Projection: MapProjectionEquirectangular})
	if err != nil {
		t.Fatal(err)
	}
	nodes := parseDiagram(t, value)
	var frame map[string]string
	for _, n := range nodes {
		if n.Attr["data-role"] == "map" {
			frame = n.Attr
		}
	}
	length := 0.0
	for _, n := range nodes {
		if n.Attr["data-role"] != "coverage-sweep-boundary" {
			continue
		}
		pts := strings.Fields(n.Attr["points"])
		for i := 1; i < len(pts); i++ {
			a, b := strings.Split(pts[i-1], ","), strings.Split(pts[i], ",")
			lo, la := inverseMap(frame, number(a[0]), number(a[1]))
			qo, qa := inverseMap(frame, number(b[0]), number(b[1]))
			if math.Abs(math.Abs(lo)-180) < 1e-5 && math.Abs(math.Abs(qo)-180) < 1e-5 {
				length += math.Abs(qa - la)
			}
		}
	}
	if math.Abs(length-10) > 1e-5 {
		t.Fatalf("seam stroke length=%g want10; matched artificial edges must disappear", length)
	}
}

func TestCoverageSVGProjectionConsistency(t *testing.T) {
	const tolerance = 1e-3
	maximum := 0.0
	for _, proj := range []MapProjection{MapProjectionAuto, MapProjectionEquirectangular, MapProjectionNorthPolar, MapProjectionSouthPolar, MapProjectionOrthographic} {
		s := coverageSynthetic()
		if proj == MapProjectionSouthPolar {
			for i := range s.Track {
				s.Track[i].Latitude = -s.Track[i].Latitude
				s.Instants[i].SubPoint = s.Track[i]
			}
		}
		opt := CoverageSVGOptions{Projection: proj, Highlight: s.Track[1].Time, TimeLabelStep: time.Minute}
		value, err := CoverageSVG(s, opt)
		if err != nil {
			t.Fatal(err)
		}
		nodes := parseDiagram(t, value)
		var frame map[string]string
		for _, n := range nodes {
			if n.Attr["data-role"] == "map" {
				frame = n.Attr
			}
		}
		checks := 0
		for _, n := range nodes {
			if n.Attr["data-role"] != "ground-track" {
				continue
			}
			for _, p := range strings.Fields(n.Attr["points"]) {
				xy := strings.Split(p, ",")
				lo, la := inverseMap(frame, number(xy[0]), number(xy[1]))
				best := math.Inf(1)
				for _, q := range s.Track {
					best = math.Min(best, math.Hypot(math.Remainder(lo-q.Longitude, 360), la-q.Latitude))
				}
				maximum = math.Max(maximum, best)
				if best > tolerance {
					t.Fatal(proj, best)
				}
				checks++
			}
		}
		if checks < 3 {
			t.Fatal(proj, "missing track", checks)
		}
		roles := map[string]int{}
		for _, n := range nodes {
			roles[n.Attr["data-role"]]++
		}
		if roles["panel-value"] < 10 || roles["subsatellite-point"] != 1 {
			t.Fatal(proj, "missing information", roles)
		}
	}
	t.Logf("independent projection inverse maximum error=%g degrees", maximum)
}

func TestCoverageSVGObserverConsistency(t *testing.T) {
	s := coverageSynthetic()
	for _, tc := range []struct {
		lon, lat  float64
		highlight bool
		want      bool
	}{{1, 1, false, false}, {15, 15, false, true}, {100, 0, false, false}, {21, 11, true, true}, {1, 1, true, false}} {
		o := CoverageSVGOptions{Observer: &satellite.Observer{Longitude: tc.lon, Latitude: tc.lat}}
		if tc.highlight {
			o.Highlight = s.Track[1].Time
		}
		value, err := CoverageSVG(s, o)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, n := range parseDiagram(t, value) {
			if n.Attr["data-role"] == "observer-membership" {
				found = true
				if n.Attr["data-inside"] != strconv.FormatBool(tc.want) {
					t.Fatal(tc, n.Attr)
				}
			}
		}
		if !found {
			t.Fatal("missing observer membership")
		}
	}
}

func TestCoverageSVGStationPasses(t *testing.T) {
	s := coverageSynthetic()
	observer := &satellite.Observer{Longitude: 21, Latitude: 11}
	look := func(second int, azimuth, elevation, magnitude float64) satellite.Look {
		return satellite.Look{Time: s.Start.Add(time.Duration(second) * time.Second), Azimuth: azimuth, Elevation: elevation, RangeKM: 1000, Magnitude: magnitude}
	}
	rise, peak, set := look(5, 15, 10, 3), look(25, 90, 70, -2), look(50, 210, 10, 4)
	first := CoveragePass{Pass: satellite.Pass{Rise: rise, Culmination: peak, Set: set, Kind: satellite.PassVisible, VisibleStart: rise.Time, VisibleEnd: set.Time, BrightestTime: peak.Time, BrightestMagnitude: peak.Magnitude}, Samples: []satellite.Look{rise, peak, look(30, 110, 65, math.NaN()), look(35, 130, 50, 1), set}}
	second := CoveragePass{Pass: satellite.Pass{Rise: look(70, 120, 10, math.NaN()), Culmination: look(100, 160, 50, math.NaN()), Set: look(120, 250, 30, math.NaN()), EndsAbove: true, Kind: satellite.PassDaylight}}
	for _, projection := range []MapProjection{MapProjectionOrthographic, MapProjectionEquirectangular} {
		value, err := CoverageSVG(s, CoverageSVGOptions{Projection: projection, Width: 640, Height: 480, Language: "en", Location: time.FixedZone("UTC+08:00", 8*3600), Observer: observer, Passes: []CoveragePass{first, second}})
		if err != nil {
			t.Fatal(err)
		}
		nodes := parseDiagram(t, value)
		count, magnitudeCurves, elevationCurves, timeTicks, polarCharts := 0, 0, 0, 0, 0
		var polarPaths []string
		for _, node := range nodes {
			switch node.Attr["data-role"] {
			case "coverage-pass":
				count++
			case "coverage-magnitude-curve":
				magnitudeCurves++
				if strings.Count(node.Attr["d"], "M ") != 2 || strings.Contains(node.Attr["d"], "NaN") {
					t.Fatal("curve must break across missing magnitudes", node.Attr["d"])
				}
			case "coverage-elevation-curve":
				elevationCurves++
				if strings.Count(node.Attr["d"], "M ") != 1 || strings.Contains(node.Attr["d"], "NaN") {
					t.Fatal("elevation curve must remain continuous", node.Attr["d"])
				}
			case "coverage-time-tick":
				timeTicks++
			case "coverage-pass-polar":
				polarCharts++
			case "coverage-pass-polar-track":
				polarPaths = append(polarPaths, node.Attr["d"])
			}
		}
		if count != 2 || magnitudeCurves != 1 || elevationCurves != 1 || timeTicks < 5 {
			t.Fatalf("passes=%d magnitude curves=%d elevation curves=%d time ticks=%d", count, magnitudeCurves, elevationCurves, timeTicks)
		}
		if polarCharts != count || len(polarPaths) != count || strings.Count(polarPaths[0], " M ") != 1 || strings.Count(polarPaths[0], " L ") != 4 || strings.Count(polarPaths[1], " L ") != 2 || !strings.Contains(value, "Window end · 20:02:00") {
			t.Fatalf("expected one polar plot per pass, including a three-point track for the truncated pass: charts=%d paths=%v", polarCharts, polarPaths)
		}
		if !strings.Contains(value, `data-role="coverage-elevation-title"`) || !strings.Contains(value, `data-role="coverage-elevation-tick"`) {
			t.Fatal("missing right elevation axis")
		}
		for _, want := range []string{"Station passes (UTC+08:00)", "20:00:05", "20:00:25", "20:00:50", "70.0°", "Brightest", "-2.0", "Window end", "Daylight"} {
			if !strings.Contains(value, want) {
				t.Errorf("missing %q", want)
			}
		}
		checkTextBoxes(t, nodes, 640, number(nodes[0].Attr["height"]))
	}
	value, err := CoverageSVG(s, CoverageSVGOptions{Observer: observer, Passes: []CoveragePass{}})
	if err != nil || !strings.Contains(value, "本时段无过境") {
		t.Fatalf("missing empty station result: %v", err)
	}
	for _, options := range []CoverageSVGOptions{
		{Passes: []CoveragePass{first}},
		{Observer: observer, Passes: []CoveragePass{second, first}},
		{Observer: observer, Passes: []CoveragePass{{Pass: first.Pass, Samples: []satellite.Look{second.Rise}}}},
	} {
		if _, err := CoverageSVG(s, options); !errors.Is(err, satellite.ErrInvalidInput) {
			t.Errorf("invalid station data returned %v", err)
		}
	}
}

func TestCoverageSVGPassChartPeakAndSubsecondTicks(t *testing.T) {
	sweep := coverageSynthetic()
	look := func(offset time.Duration, elevation float64) satellite.Look {
		return satellite.Look{Time: sweep.Start.Add(offset), Elevation: elevation, RangeKM: 1000, Magnitude: math.NaN()}
	}
	rise, peak, set := look(100*time.Millisecond, 2.89e-6), look(250*time.Millisecond, 90), look(475*time.Millisecond, -2.89e-6)
	pass := CoveragePass{
		Pass:    satellite.Pass{Rise: rise, Culmination: peak, Set: set},
		Samples: []satellite.Look{rise, look(350*time.Millisecond, 80.3), set},
	}
	document, err := CoverageSVG(sweep, CoverageSVGOptions{Width: 640, Height: 480, Observer: &satellite.Observer{}, Passes: []CoveragePass{pass}})
	if err != nil {
		t.Fatal(err)
	}
	var elevationTicks, timeTicks []string
	var curve string
	for _, node := range parseDiagram(t, document) {
		switch node.Attr["data-role"] {
		case "coverage-elevation-tick":
			elevationTicks = append(elevationTicks, node.Text)
		case "coverage-time-tick":
			timeTicks = append(timeTicks, node.Text)
		case "coverage-elevation-curve":
			curve = node.Attr["d"]
		}
	}
	if strings.Join(elevationTicks, ",") != "90.0°,0°" || strings.Count(curve, "L ") != 3 {
		t.Fatalf("want a curve including the unsampled peak, got ticks %v and path %q", elevationTicks, curve)
	}
	if len(timeTicks) < 3 || timeTicks[0] != "12:00:00.100" || timeTicks[len(timeTicks)-1] != "12:00:00.475" {
		t.Fatalf("expected distinct subsecond time ticks, got %v", timeTicks)
	}
}

func TestCoveragePassPolarPosition(t *testing.T) {
	const centerX, centerY, radius = 100.0, 120.0, 80.0
	for _, tc := range []struct {
		name      string
		azimuth   float64
		elevation float64
		wantX     float64
		wantY     float64
	}{
		{name: "north horizon", azimuth: 0, elevation: 0, wantX: centerX, wantY: centerY - radius},
		{name: "east horizon", azimuth: 90, elevation: 0, wantX: centerX + radius, wantY: centerY},
		{name: "south horizon", azimuth: 180, elevation: 0, wantX: centerX, wantY: centerY + radius},
		{name: "west horizon", azimuth: 270, elevation: 0, wantX: centerX - radius, wantY: centerY},
		{name: "wrapped north", azimuth: 360, elevation: 30, wantX: centerX, wantY: centerY - 2*radius/3},
		{name: "zenith", azimuth: 35, elevation: 90, wantX: centerX, wantY: centerY},
	} {
		t.Run(tc.name, func(t *testing.T) {
			positionX, positionY, ok := coveragePassPolarPosition(centerX, centerY, radius, tc.azimuth, tc.elevation)
			if !ok || math.Hypot(positionX-tc.wantX, positionY-tc.wantY) > 1e-9 {
				t.Fatalf("position=(%.9f, %.9f), want=(%.9f, %.9f), ok=%t", positionX, positionY, tc.wantX, tc.wantY, ok)
			}
		})
	}
	if _, _, ok := coveragePassPolarPosition(centerX, centerY, radius, 0, -10); ok {
		t.Fatal("point below the horizon was plotted")
	}
}

func TestCoverageSVGStationPolarCharts(t *testing.T) {
	s := coverageSynthetic()
	look := func(second int, azimuth, elevation float64) satellite.Look {
		return satellite.Look{Time: s.Start.Add(time.Duration(second) * time.Second), Azimuth: azimuth, Elevation: elevation, RangeKM: 1000, Magnitude: 1}
	}
	pass := CoveragePass{
		Pass:    satellite.Pass{Rise: look(0, 0, 0), Culmination: look(30, 90, 60), Set: look(60, 180, 0)},
		Samples: []satellite.Look{look(15, math.NaN(), 30), look(30, math.NaN(), 60), look(45, 135, 30)},
	}
	for _, language := range []string{"zh", "en"} {
		for _, size := range [][2]int{{640, 480}, {1200, 800}} {
			value, err := CoverageSVG(s, CoverageSVGOptions{Width: size[0], Height: size[1], Language: language, Observer: &satellite.Observer{}, Passes: []CoveragePass{pass}})
			if err != nil {
				t.Fatal(err)
			}
			nodes := parseDiagram(t, value)
			roles := map[string]int{}
			markers := map[string]map[string]string{}
			rings := map[string]map[string]string{}
			var track string
			for _, node := range nodes {
				roles[node.Attr["data-role"]]++
				switch node.Attr["data-role"] {
				case "coverage-pass-polar-marker":
					markers[node.Attr["data-stage"]] = node.Attr
				case "coverage-pass-polar-ring":
					rings[node.Attr["data-elevation"]] = node.Attr
				case "coverage-pass-polar-track":
					track = node.Attr["d"]
				}
			}
			if roles["coverage-pass-polar"] != 1 || roles["coverage-pass-polar-track"] != 1 || roles["coverage-pass-polar-arrow"] != 1 || len(rings) != 3 || roles["coverage-pass-polar-direction"] != 4 || len(markers) != 3 {
				t.Fatalf("language=%s size=%v roles=%v markers=%d", language, size, roles, len(markers))
			}
			if markers["rise"]["fill"] != "white" || markers["rise"]["stroke"] == markers["set"]["fill"] || markers["set"]["fill"] == "white" {
				t.Fatal("rise and set markers need distinct fill and color")
			}
			for _, stage := range []string{"rise", "culmination", "set"} {
				marker := markers[stage]
				if marker == nil || !strings.Contains(track, marker["cx"]+" "+marker["cy"]) {
					t.Fatalf("%s must lie on the track: marker=%v track=%s", stage, marker, track)
				}
			}
			if strings.Count(track, " M ") != 2 || strings.Count(track, " L ") != 2 {
				t.Fatalf("track must break at the missing sample and include the highest point: %s", track)
			}
			horizon := rings["0.0"]
			centerX, centerY, radius := number(horizon["cx"]), number(horizon["cy"]), number(horizon["r"])
			if centerX != float64(size[0])/2 || radius <= 0 || centerY+radius >= number(nodes[0].Attr["height"]) || math.Abs(number(rings["60.0"]["r"])-radius/3) > .001 {
				t.Fatalf("invalid polar plot bounds or ring spacing: %v", rings)
			}
			checkTextBoxes(t, nodes, float64(size[0]), number(nodes[0].Attr["height"]))
		}
	}
}

func TestCoveragePassPolarArrowFollowsTrack(t *testing.T) {
	for _, finish := range []float64{100, -100} {
		arrow := coveragePassPolarArrow([][2]float64{{0, 0}, {finish, 0}}, 55, 0)
		coords := strings.Fields(arrow)
		if len(coords) != 7 {
			t.Fatalf("incomplete arrowhead: %q", arrow)
		}
		tip := number(strings.TrimPrefix(coords[0], "M"))
		base := number(strings.TrimPrefix(coords[2], "L"))
		if tip != math.Copysign(55, finish) || (tip-base)*finish <= 0 {
			t.Fatalf("arrowhead %q points away from the endpoint %g", arrow, finish)
		}
	}
}

func TestCoveragePassTimeStepUsesAvailableWidth(t *testing.T) {
	for _, tc := range []struct {
		name     string
		duration time.Duration
		width    float64
		format   string
		want     time.Duration
	}{
		{name: "wide short pass", duration: 365 * time.Second, width: 1012, format: "15:04:05", want: 30 * time.Second},
		{name: "narrow short pass", duration: 365 * time.Second, width: 452, format: "15:04:05", want: time.Minute},
		{name: "long pass", duration: 45 * time.Minute, width: 1012, format: "15:04:05", want: 5 * time.Minute},
		{name: "cross day labels", duration: 2 * time.Hour, width: 1012, format: "01-02 15:04:05", want: 15 * time.Minute},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := coveragePassTimeStep(tc.duration, tc.width, tc.format); got != tc.want {
				t.Fatalf("step=%s, want %s", got, tc.want)
			}
		})
	}
}

func TestCoverageSVGShortPassArrow(t *testing.T) {
	sweep := coverageSynthetic()
	rise := satellite.Look{Time: sweep.Start, Azimuth: 90, Elevation: 10}
	peak := satellite.Look{Time: sweep.Start.Add(2 * time.Second), Azimuth: 92.5, Elevation: 11}
	set := satellite.Look{Time: sweep.Start.Add(5 * time.Second), Azimuth: 95, Elevation: 10}
	pass := CoveragePass{Pass: satellite.Pass{Rise: rise, Culmination: peak, Set: set}}
	image, err := CoverageSVG(sweep, CoverageSVGOptions{Observer: &satellite.Observer{}, Passes: []CoveragePass{pass}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(image, `data-role="coverage-pass-polar-arrow"`) {
		t.Fatal("short moving pass has no direction arrow")
	}
}

func TestCoverageSVGStationTimeOffsetChange(t *testing.T) {
	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	sweep := coverageSynthetic()
	sweep.Start = time.Date(2026, 11, 1, 5, 59, 0, 0, time.UTC)
	sweep.End = sweep.Start.Add(2 * time.Minute)
	for index := range sweep.Track {
		sweep.Track[index].Time = sweep.Start.Add(time.Duration(index) * time.Minute)
		sweep.Instants[index].Time = sweep.Track[index].Time
		sweep.Instants[index].SubPoint.Time = sweep.Track[index].Time
	}
	samples := []satellite.Look{
		{Time: sweep.Start, Azimuth: 270, Elevation: 10},
		{Time: sweep.Start.Add(time.Minute), Azimuth: 265, Elevation: 15},
		{Time: sweep.End, Azimuth: 260, Elevation: 10},
	}
	pass := CoveragePass{Pass: satellite.Pass{Rise: samples[0], Culmination: samples[1], Set: samples[2]}, Samples: samples}
	image, err := CoverageSVG(sweep, CoverageSVGOptions{Width: 640, Height: 480, Language: "en", Location: zone, Observer: &satellite.Observer{}, Passes: []CoveragePass{pass}})
	if err != nil {
		t.Fatal(err)
	}
	var labels []string
	nodes := parseDiagram(t, image)
	for _, node := range nodes {
		if node.Attr["data-role"] == "coverage-time-tick" {
			labels = append(labels, node.Text)
		}
	}
	if len(labels) < 2 || labels[0] != "01:59:00 -04:00" || labels[len(labels)-1] != "01:01:00 -05:00" {
		t.Fatalf("local tick labels do not distinguish both UTC offsets: %q", labels)
	}
	checkTextBoxes(t, nodes, 640, number(nodes[0].Attr["height"]))
}

func TestCoverageSVGStationSpacingConsistency(t *testing.T) {
	sweep := coverageSynthetic()
	observer := &satellite.Observer{Longitude: 21, Latitude: 11}
	zone := time.FixedZone("America/Argentina/Buenos_Aires", -3*3600)
	for _, size := range [][2]int{{1200, 800}, {640, 600}, {640, 480}} {
		for _, language := range []string{"zh", "en"} {
			for _, projection := range []MapProjection{MapProjectionEquirectangular, MapProjectionOrthographic, MapProjectionNorthPolar, MapProjectionSouthPolar} {
				t.Run(fmt.Sprintf("%dx%d/%s/%s", size[0], size[1], language, projection), func(t *testing.T) {
					value, err := CoverageSVG(sweep, CoverageSVGOptions{Width: size[0], Height: size[1], Language: language, Location: zone, Projection: projection, Observer: observer, Highlight: sweep.Track[1].Time, Passes: []CoveragePass{}})
					if err != nil {
						t.Fatal(err)
					}
					nodes := parseDiagram(t, value)
					var contentBottom, separator float64
					for _, node := range nodes {
						switch node.Attr["data-role"] {
						case "map":
							contentBottom = math.Max(contentBottom, number(node.Attr["data-y"])+number(node.Attr["data-height"]))
						case "panel-value", "legend":
							contentBottom = math.Max(contentBottom, number(node.Attr["y"])+.23*number(node.Attr["font-size"]))
						case "coverage-station-title":
							if separator == 0 {
								separator = number(node.Attr["y"]) - 25
							}
						}
					}
					if gap := separator - contentBottom; gap < 12 || gap > 28 {
						t.Errorf("separator %.1f, content bottom %.1f, gap %.1f", separator, contentBottom, gap)
					}
					checkTextBoxes(t, nodes, float64(size[0]), number(nodes[0].Attr["height"]))
				})
			}
		}
	}
}

func TestCoverageSVGInvalidSynthetic(t *testing.T) {
	s := coverageSynthetic()
	for _, opt := range []CoverageSVGOptions{{Width: 639}, {Height: 479}, {Projection: "bad"}, {Highlight: s.Start.Add(time.Second)}, {Observer: &satellite.Observer{Latitude: 91}}, {Observer: &satellite.Observer{Height: math.NaN()}}} {
		if _, err := CoverageSVG(s, opt); !errors.Is(err, satellite.ErrInvalidInput) {
			t.Fatal(opt, err)
		}
	}
	for _, q := range []satellite.CoverageSweep{{}, func() satellite.CoverageSweep { q := s; q.TrackIllumination = nil; return q }(), func() satellite.CoverageSweep { q := s; q.Elements.Epoch = time.Time{}; return q }()} {
		if _, err := CoverageSVG(q, CoverageSVGOptions{}); !errors.Is(err, satellite.ErrInvalidInput) {
			t.Fatal(err)
		}
	}
}
func TestCoverageSVGLayoutConsistency(t *testing.T) {
	s := coverageSynthetic()
	for _, language := range []string{"zh", "en"} {
		for _, size := range [][2]int{{1200, 800}, {640, 480}} {
			for _, highlight := range []bool{false, true} {
				for _, projection := range []MapProjection{MapProjectionEquirectangular, MapProjectionOrthographic} {
					o := CoverageSVGOptions{Width: size[0], Height: size[1], Language: language, Projection: projection, Observer: &satellite.Observer{Longitude: 21, Latitude: 11}, TimeLabelStep: time.Minute}
					if highlight {
						o.Highlight = s.Track[1].Time
					}
					v, err := CoverageSVG(s, o)
					if err != nil {
						t.Fatal(err)
					}
					checkTextBoxes(t, parseDiagram(t, v), float64(size[0]), float64(size[1]))
					for _, role := range []string{"coverage-sweep-legend", "visible-sweep-legend"} {
						if !strings.Contains(v, `data-role="`+role+`"`) {
							t.Errorf("%s %v %s highlight=%t: missing %s", language, size, projection, highlight, role)
						}
					}
					if !strings.Contains(v, "Sunlit / penumbra / umbra track") && !strings.Contains(v, "向阳 / 半影 / 本影轨迹") {
						t.Errorf("%s %v %s highlight=%t: missing illumination legend", language, size, projection, highlight)
					}
				}
			}
		}
	}
}

func TestCoverageSVGWrappedPanelSynthetic(t *testing.T) {
	sweep := coverageSynthetic()
	zone := time.FixedZone("America/Argentina/Buenos_Aires", -3*3600)
	for _, height := range []int{599, 600, 601, 800} {
		for _, language := range []string{"zh", "en"} {
			for _, projection := range []MapProjection{MapProjectionEquirectangular, MapProjectionOrthographic, MapProjectionNorthPolar, MapProjectionSouthPolar} {
				t.Run(fmt.Sprintf("%d/%s/%s", height, language, projection), func(t *testing.T) {
					value, err := CoverageSVG(sweep, CoverageSVGOptions{Width: 640, Height: height, Location: zone, Language: language, Projection: projection, Highlight: sweep.Track[1].Time, Observer: &satellite.Observer{Longitude: 21, Latitude: 11}})
					if err != nil {
						t.Fatal(err)
					}
					checkTextBoxes(t, parseDiagram(t, value), 640, float64(height))
				})
			}
		}
	}
}

func TestCoverageSVGPolarSeamSynthetic(t *testing.T) {
	s := coverageSynthetic()
	s.Swept = []satellite.Region{{Outer: []satellite.GeoPoint{{Longitude: -180, Latitude: 70}, {Longitude: -90, Latitude: 75}, {Longitude: 0, Latitude: 70}, {Longitude: 90, Latitude: 75}, {Longitude: 180, Latitude: 70}, {Longitude: 180, Latitude: 90}, {Longitude: -180, Latitude: 90}, {Longitude: -180, Latitude: 70}}}}
	v, err := CoverageSVG(s, CoverageSVGOptions{Projection: MapProjectionNorthPolar})
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range parseDiagram(t, v) {
		if n.Attr["data-role"] == "coverage-sweep-boundary" {
			pts := strings.Fields(n.Attr["points"])
			for i := 1; i < len(pts); i++ {
				if pts[i] == pts[i-1] {
					continue
				}
			}
		}
	}
	for _, path := range coverageStrokePaths(s.Swept) {
		for i := 1; i < len(path); i++ {
			if math.Abs(path[i].Latitude) == 90 && path[i].Latitude == path[i-1].Latitude {
				t.Fatal("polar closure stroked")
			}
		}
	}
}

func coverageActualSweeps(t testing.TB, visual bool) []satellite.CoverageSweep {
	t.Helper()
	raw, err := os.ReadFile("../testdata/skyfield_coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Groups []struct{ Elements satellite.Elements }
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	var sweeps []satellite.CoverageSweep
	cases := []struct {
		index            int
		offset, duration time.Duration
	}{{0, 0, 90 * time.Minute}, {5, 0, 10 * time.Minute}}
	if visual {
		cases[0].offset, cases[0].duration = 75*time.Minute, 35*time.Minute
		cases = append(cases, struct {
			index            int
			offset, duration time.Duration
		}{0, 45 * time.Minute, 20 * time.Minute})
	}
	for _, q := range cases {
		at := start.Add(q.offset)
		s, err := satellite.CoverageBetween(fixture.Groups[q.index].Elements, at, at.Add(q.duration), time.Minute, satellite.CoverageOptions{})
		if err != nil {
			t.Fatal(err)
		}
		sweeps = append(sweeps, s)
	}
	return sweeps
}
func coverageBaselineCases(t testing.TB) []struct {
	name string
	s    satellite.CoverageSweep
	o    CoverageSVGOptions
} {
	s := coverageActualSweeps(t, true)
	observer := satellite.Observer{Longitude: s[0].Track[20].Longitude, Latitude: s[0].Track[20].Latitude}
	return []struct {
		name string
		s    satellite.CoverageSweep
		o    CoverageSVGOptions
	}{
		{"coverage-interval-zh", s[0], CoverageSVGOptions{Observer: &observer, TimeLabelStep: 10 * time.Minute}},
		{"coverage-interval-en", s[0], CoverageSVGOptions{Language: "en", TimeLabelStep: 10 * time.Minute}},
		{"coverage-highlight-zh", s[0], CoverageSVGOptions{Highlight: s[0].Instants[20].Time, Observer: &observer, TimeLabelStep: 10 * time.Minute}},
		{"coverage-highlight-en", s[0], CoverageSVGOptions{Highlight: s[0].Instants[20].Time, Language: "en", TimeLabelStep: 10 * time.Minute}},
		{"coverage-polar-zh", s[1], CoverageSVGOptions{Projection: MapProjectionNorthPolar, Highlight: s[1].Instants[5].Time}},
		{"coverage-seam-zh", s[2], CoverageSVGOptions{Projection: MapProjectionEquirectangular, Highlight: s[2].Instants[10].Time, TimeLabelStep: 5 * time.Minute}},
	}
}
func TestCoverageSVGActualConsistency(t *testing.T) {
	for _, tc := range coverageBaselineCases(t) {
		for _, language := range []string{"zh", "en"} {
			for _, size := range [][2]int{{1200, 800}, {640, 480}} {
				o := tc.o
				o.Language = language
				o.Width = size[0]
				o.Height = size[1]
				value, err := CoverageSVG(tc.s, o)
				if err != nil {
					t.Fatal(tc.name, err)
				}
				checkTextBoxes(t, parseDiagram(t, value), float64(size[0]), float64(size[1]))
			}
		}
	}
}
func BenchmarkCoverageSVG(b *testing.B) {
	s := coverageActualSweeps(b, false)[0]
	o := CoverageSVGOptions{Projection: MapProjectionEquirectangular}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := CoverageSVG(s, o); err != nil {
			b.Fatal(err)
		}
	}
}

func TestCoverageSVGShapeConsistency(t *testing.T) {
	const tolerance = 1e-3
	s := coverageSynthetic()
	s.Swept[0].Holes = nil
	opt := CoverageSVGOptions{Projection: MapProjectionEquirectangular, Highlight: s.Track[1].Time}
	value, err := CoverageSVG(s, opt)
	if err != nil {
		t.Fatal(err)
	}
	nodes := parseDiagram(t, value)
	var frame map[string]string
	for _, n := range nodes {
		if n.Attr["data-role"] == "map" {
			frame = n.Attr
		}
	}
	want := map[string][][]satellite.GeoPoint{"coverage-sweep": {s.Swept[0].Outer}, "visible-sweep": {s.VisibleSwept[0].Outer}, "satellite-footprint": {s.Instants[1].Footprint}, "satellite-visible": s.Instants[1].Visible, "night-region": {s.Instants[1].Night}}
	maximum := 0.0
	checks := 0
	for _, n := range nodes {
		rings, ok := want[n.Attr["data-role"]]
		if !ok {
			continue
		}
		tokens := strings.Fields(n.Attr["d"])
		for i := 0; i < len(tokens); i++ {
			if len(tokens[i]) == 0 || tokens[i] == "Z" {
				continue
			}
			x := number(strings.TrimLeft(tokens[i], "ML"))
			i++
			if i >= len(tokens) {
				t.Fatal("malformed path")
			}
			y := number(tokens[i])
			lo, la := inverseMap(frame, x, y)
			best := math.Inf(1)
			for _, ring := range rings {
				for _, p := range ring {
					best = math.Min(best, math.Hypot(math.Remainder(lo-p.Longitude, 360), la-p.Latitude))
				}
			}
			maximum = math.Max(maximum, best)
			checks++
			if best > tolerance {
				t.Fatal(n.Attr["data-role"], lo, la, best)
			}
		}
	}
	if checks < 20 {
		t.Fatal("missing layers", checks)
	}
	t.Logf("independent polygon inverse checks=%d maximum=%g degrees", checks, maximum)
}

func TestCoverageSVGNightRegionTwilightMembership(t *testing.T) {
	raw, err := os.ReadFile("../testdata/skyfield_coverage.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Groups []struct{ Elements satellite.Elements }
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	for _, threshold := range []float64{-6, 0, 6, 30, -89.999999, 89.999999, math.Nextafter(-90, 0), math.Nextafter(90, 0)} {
		sweep, err := satellite.CoverageBetween(fixture.Groups[0].Elements, at, at, time.Minute, satellite.CoverageOptions{TwilightSunAltitude: &threshold})
		if err != nil {
			t.Fatal(err)
		}
		for _, projection := range []MapProjection{MapProjectionEquirectangular, MapProjectionNorthPolar, MapProjectionSouthPolar, MapProjectionOrthographic} {
			value, err := CoverageSVG(sweep, CoverageSVGOptions{Projection: projection, Highlight: at})
			if err != nil {
				t.Fatal(err)
			}
			var frame map[string]string
			var path string
			for _, node := range parseDiagram(t, value) {
				switch node.Attr["data-role"] {
				case "map":
					frame = node.Attr
				case "night-region":
					path = node.Attr["d"]
				}
			}
			for row := 1; row < 16; row++ {
				for column := 1; column < 16; column++ {
					x, y := float64(column)/16, float64(row)/16
					if projection != MapProjectionEquirectangular && math.Hypot(x-.5, y-.5) > .48 {
						continue
					}
					x = number(frame["data-x"]) + x*number(frame["data-width"])
					y = number(frame["data-y"]) + y*number(frame["data-height"])
					lon, lat := inverseMap(frame, x, y)
					radians := math.Pi / 180
					altitude := math.Asin(math.Sin(lat*radians)*math.Sin(-18*radians)+math.Cos(lat*radians)*math.Cos(-18*radians)*math.Cos((lon-176)*radians)) / radians
					if math.Abs(altitude-threshold) < 2 {
						continue
					}
					if got := coverageSVGFillContains(path, x, y); got != (altitude < threshold) {
						t.Fatalf("%s threshold %g: filled=%t at (%.2f, %.2f), Sun alt %.2f", projection, threshold, got, lon, lat, altitude)
					}
				}
			}
		}
	}
}

func TestCoverageSVGOrthographicSweepFill(t *testing.T) {
	elements := coverageActualSweeps(t, true)[0].Elements
	start := time.Date(2026, 11, 14, 18, 0, 0, 0, time.UTC)
	sweep, err := satellite.CoverageBetween(elements, start, start.Add(time.Hour), time.Minute, satellite.CoverageOptions{MinElevation: 10})
	if err != nil {
		t.Fatal(err)
	}
	value, err := CoverageSVG(sweep, CoverageSVGOptions{Projection: MapProjectionOrthographic, Highlight: start.Add(23 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	var frame map[string]string
	paths := map[string]string{}
	for _, node := range parseDiagram(t, value) {
		role := node.Attr["data-role"]
		if role == "map" {
			frame = node.Attr
		}
		if node.Name == "path" && node.Attr["fill-rule"] == "evenodd" {
			paths[role] = node.Attr["d"]
		}
	}
	width := number(frame["data-width"])
	centerX := number(frame["data-x"]) + width/2
	centerY := number(frame["data-y"]) + width/2
	for role, regions := range map[string][]satellite.Region{"coverage-sweep": sweep.Swept, "visible-sweep": sweep.VisibleSwept} {
		for _, offset := range [][2]float64{{0, 0}, {-.72, -.36}, {.18, -.60}, {-.18, .60}, {.54, .18}} {
			x, y := centerX+offset[0]*width/2, centerY+offset[1]*width/2
			lon, lat := inverseMap(frame, x, y)
			want := coverageInside(regions, satellite.GeoPoint{Longitude: math.Remainder(lon, 360), Latitude: lat})
			if got := coverageSVGFillContains(paths[role], x, y); got != want {
				t.Errorf("%s at %.4f°, %.4f°: filled=%t, want %t", role, lon, lat, got, want)
			}
		}
	}
}

func coverageSVGFillContains(path string, x, y float64) bool {
	inside := false
	for _, part := range strings.Split(path, "Z") {
		fields := strings.Fields(part)
		if len(fields) < 6 {
			continue
		}
		previousX, previousY := number(strings.TrimLeft(fields[len(fields)-2], "ML")), number(fields[len(fields)-1])
		for index := 0; index < len(fields); index += 2 {
			currentX, currentY := number(strings.TrimLeft(fields[index], "ML")), number(fields[index+1])
			if (currentY > y) != (previousY > y) && x < currentX+(previousX-currentX)*(y-currentY)/(previousY-currentY) {
				inside = !inside
			}
			previousX, previousY = currentX, currentY
		}
	}
	return inside
}

func TestCoverageSVGBaselineConsistency(t *testing.T) {
	if runtime.GOARCH == "386" && strings.HasPrefix(runtime.Version(), "go1.16") {
		t.Skip("platform byte baseline")
	}
	expected := map[string]struct {
		size int
		md5  string
	}{
		"coverage-interval-zh":  {498537, "e1abc28ae6edfbdf924420510f2361b0"},
		"coverage-interval-en":  {498186, "cf02e959ae3c32c7da4cbefa60d2aff7"},
		"coverage-highlight-zh": {701289, "b392158404130049ed0691bc103959c1"},
		"coverage-highlight-en": {700962, "e57b506ef86c85b91036bdec47e68439"},
		"coverage-polar-zh":     {413572, "595b03b29a060093214dde1bff4091d7"},
		"coverage-seam-zh":      {551275, "7460d91e3de1a2e525a0e6ddd25c58ce"},
	}
	missing := false
	for _, tc := range coverageBaselineCases(t) {
		value, err := CoverageSVG(tc.s, tc.o)
		if err != nil {
			t.Fatal(err)
		}
		sum := fmt.Sprintf("%x", md5.Sum([]byte(value)))
		t.Logf("%s bytes=%d md5=%s", tc.name, len(value), sum)
		want, ok := expected[tc.name]
		if !ok {
			missing = true
			continue
		}
		if len(value) != want.size || sum != want.md5 {
			t.Errorf("%s baseline mismatch", tc.name)
		}
	}
	if missing {
		t.Fatal("missing byte baseline")
	}
}

func TestCoverageSVGEquirectangularCentersOnTrack(t *testing.T) {
	elements, err := satellite.ParseTLEText([]byte(`DFH 1
1 04382U 70034A   26324.08985725  .00003017  00000+0  41599-3 0  9993
2 04382  68.4229 103.5770 1029888 199.8637 155.9389 13.14766370666176`))
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 11, 20, 7, 3, 0, 0, time.UTC)
	sweep, err := satellite.CoverageBetween(elements[0], start, start.Add(time.Hour), 5*time.Minute, satellite.CoverageOptions{MinElevation: 10})
	if err != nil {
		t.Fatal(err)
	}
	image, err := CoverageSVG(sweep, CoverageSVGOptions{Projection: MapProjectionEquirectangular, Width: 1200, Height: 800})
	if err != nil {
		t.Fatal(err)
	}
	focus := sweep.Track[len(sweep.Track)/2]
	foundMap := false
	for _, node := range parseDiagram(t, image) {
		if node.Attr["data-role"] != "map" {
			continue
		}
		foundMap = true
		if got := number(node.Attr["data-longitude"]); math.Abs(got-focus.Longitude) > 1e-6 {
			t.Fatalf("map central meridian %.6f, want focus longitude %.6f", got, focus.Longitude)
		}
		frameCenter := number(node.Attr["data-x"]) + number(node.Attr["data-width"])/2
		projectedCenter := number(node.Attr["data-x"]) + 180.0/360.0*number(node.Attr["data-width"])
		if math.Abs(frameCenter-projectedCenter) > 1e-6 {
			t.Fatalf("focus point projects to x=%.6f, map center is %.6f", projectedCenter, frameCenter)
		}
	}
	if !foundMap {
		t.Fatal("missing equirectangular map")
	}
}

func TestCoverageSVGEquirectangularKeepsSeamSides(t *testing.T) {
	sweep := coverageSynthetic()
	for index, longitude := range []float64{120, 123, -58.29200738758804} {
		sweep.Track[index].Longitude = longitude
		sweep.Instants[index].SubPoint.Longitude = longitude
		sweep.TrackIllumination[index] = satellite.IlluminationSunlit
	}
	image, err := CoverageSVG(sweep, CoverageSVGOptions{Width: 640, Height: 480, Projection: MapProjectionEquirectangular, Highlight: sweep.End})
	if err != nil {
		t.Fatal(err)
	}
	var left, width float64
	var seamLeft, seamRight bool
	for _, node := range parseDiagram(t, image) {
		switch node.Attr["data-role"] {
		case "map":
			left, width = number(node.Attr["data-x"]), number(node.Attr["data-width"])
		case "ground-track":
			var previousX float64
			for index, pair := range strings.Fields(node.Attr["points"]) {
				x := number(strings.SplitN(pair, ",", 2)[0])
				seamLeft = seamLeft || math.Abs(x-left) < 1e-5
				seamRight = seamRight || math.Abs(x-left-width) < 1e-5
				if index > 0 && math.Abs(x-previousX) > width/2+1e-5 {
					t.Fatalf("track segment jumps across the map: %g to %g", previousX, x)
				}
				previousX = x
			}
		}
	}
	if !seamLeft || !seamRight {
		t.Fatalf("track must reach both sides of the seam: left=%t right=%t", seamLeft, seamRight)
	}
}
