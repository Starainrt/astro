package svg

import (
	"github.com/starainrt/astro/satellite"
	"crypto/md5"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

type svgNode struct {
	Name string
	Attr map[string]string
	Text string
}

func parseDiagram(t *testing.T, value string) []svgNode {
	t.Helper()
	if strings.Contains(value, "NaN") || strings.Contains(value, "Inf") {
		t.Fatal("nonfinite diagram")
	}
	d := xml.NewDecoder(strings.NewReader(value))
	var nodes []svgNode
	for {
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		switch x := token.(type) {
		case xml.StartElement:
			a := map[string]string{}
			for _, v := range x.Attr {
				a[v.Name.Local] = v.Value
			}
			nodes = append(nodes, svgNode{Name: x.Name.Local, Attr: a})
		case xml.CharData:
			if len(nodes) > 0 {
				nodes[len(nodes)-1].Text += string(x)
			}
		}
	}
	return nodes
}
func mapSynthetic() satellite.TransitPath {
	at := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	p := satellite.TransitPath{Elements: satellite.Elements{Name: "ISS", CatalogNumber: 25544, Epoch: at}}
	for i := 0; i < 3; i++ {
		q := satellite.TransitPathPoint{Time: at.Add(time.Duration(i) * time.Second), Longitude: 179 + float64(i), Latitude: 10 + float64(i), HalfWidthKM: 1, TargetAltitude: 30}
		if q.Longitude > 180 {
			q.Longitude -= 360
		}
		p.CenterLine = append(p.CenterLine, q)
		q.Latitude += .1
		p.LeftLimit = append(p.LeftLimit, q)
		q.Latitude -= .2
		p.RightLimit = append(p.RightLimit, q)
	}
	p.Event = satellite.TransitEvent{Target: satellite.TransitSun, Start: at, Greatest: at.Add(time.Second), End: at.Add(2 * time.Second), StartPoint: p.CenterLine[0], GreatestPoint: p.CenterLine[1], EndPoint: p.CenterLine[2], Complete: true}
	p.StartHorizon = []satellite.GeoPoint{{Longitude: 179, Latitude: 10}, {Longitude: 180, Latitude: 20}, {Longitude: -179, Latitude: 10}}
	p.EndHorizon = p.StartHorizon
	return p
}
func number(s string) float64 { x, _ := strconv.ParseFloat(s, 64); return x }
func TestTransitMapSVGSynthetic(t *testing.T) {
	p := mapSynthetic()
	for _, projection := range []MapProjection{MapProjectionAuto, MapProjectionEquirectangular, MapProjectionNorthPolar, MapProjectionSouthPolar, MapProjectionOrthographic} {
		value, err := TransitMapSVG(p, TransitMapSVGOptions{Projection: projection})
		if err != nil {
			t.Fatal(err)
		}
		nodes := parseDiagram(t, value)
		roles := map[string]int{}
		for _, n := range nodes {
			roles[n.Attr["data-role"]]++
		}
		for _, role := range []string{"transit-band", "center-line", "left-limit", "right-limit", "target-horizon", "event-start", "event-greatest", "event-end"} {
			if roles[role] == 0 && projection != MapProjectionSouthPolar {
				t.Fatal(projection, role)
			}
		}
	}
	for _, o := range []TransitMapSVGOptions{{Width: 639}, {Height: 479}, {Projection: "bad"}} {
		if _, err := TransitMapSVG(p, o); !errors.Is(err, satellite.ErrInvalidInput) {
			t.Fatal(err)
		}
	}
	if _, err := TransitMapSVG(satellite.TransitPath{}, TransitMapSVGOptions{}); !errors.Is(err, satellite.ErrInvalidInput) {
		t.Fatal(err)
	}
}
func TestTransitMapSVGProjectionConsistency(t *testing.T) {
	p := mapSynthetic()
	p.CenterLine[0].Longitude = 20
	p.CenterLine[1].Longitude = 21
	p.CenterLine[2].Longitude = 22
	for _, proj := range []MapProjection{MapProjectionEquirectangular, MapProjectionNorthPolar, MapProjectionSouthPolar, MapProjectionOrthographic} {
		if proj == MapProjectionSouthPolar {
			for i := range p.CenterLine {
				p.CenterLine[i].Latitude = -10 - float64(i)
			}
		} else {
			for i := range p.CenterLine {
				p.CenterLine[i].Latitude = 10 + float64(i)
			}
		}
		p.Event.GreatestPoint = p.CenterLine[1]
		value, err := TransitMapSVG(p, TransitMapSVGOptions{Projection: proj})
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
		count := 0
		for _, n := range nodes {
			if n.Attr["data-role"] != "center-line" {
				continue
			}
			for _, pair := range strings.Fields(n.Attr["points"]) {
				v := strings.Split(pair, ",")
				x, y := number(v[0]), number(v[1])
				lon, lat := inverseMap(frame, x, y)
				found := false
				for _, q := range p.CenterLine {
					if math.Abs(math.Remainder(lon-q.Longitude, 360)) < 1e-3 && math.Abs(lat-q.Latitude) < 1e-3 {
						found = true
					}
				}
				if !found {
					t.Fatal(proj, lon, lat)
				}
				count++
			}
		}
		if count < 3 {
			t.Fatal("missing points", proj, count)
		}
	}
}
func inverseMap(f map[string]string, x, y float64) (float64, float64) {
	fx, fy, w, h := number(f["data-x"]), number(f["data-y"]), number(f["data-width"]), number(f["data-height"])
	cl, cp := number(f["data-longitude"]), number(f["data-latitude"])
	r := math.Min(w, h) / 2
	dx, dy := (x-fx-w/2)/r, -(y-fy-h/2)/r
	rad := math.Pi / 180
	switch f["data-projection"] {
	case "north-polar":
		return cl + 180 - math.Atan2(dx, dy)/rad, 90 - math.Hypot(dx, dy)*90
	case "south-polar":
		return cl + math.Atan2(dx, dy)/rad, -90 + math.Hypot(dx, dy)*90
	case "orthographic":
		rho := math.Hypot(dx, dy)
		if rho == 0 {
			return cl, cp
		}
		c := math.Asin(math.Min(1, rho))
		p := cp * rad
		return cl + math.Atan2(dx*math.Sin(c), rho*math.Cos(p)*math.Cos(c)-dy*math.Sin(p)*math.Sin(c))/rad, math.Asin(math.Cos(c)*math.Sin(p)+dy*math.Sin(c)*math.Cos(p)/rho) / rad
	}
	return cl + (x-fx)/w*360 - 180, (90 - (y-fy)/h*180)
}
func actualMapPaths(t testing.TB) []satellite.TransitPath {
	t.Helper()
	data, err := os.ReadFile("../testdata/skyfield_transit_events.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Groups []struct {
			Elements   satellite.Elements
			Target     satellite.TransitTarget
			Minimum    float64
			Start, End time.Time
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	var paths []satellite.TransitPath
	for _, i := range []int{0, 3, 42, 48, 60} {
		g := fixture.Groups[i]
		events, err := satellite.TransitEvents(g.Elements, g.Start, g.End, satellite.TransitSearchOptions{Targets: []satellite.TransitTarget{g.Target}, MinTargetAltitude: g.Minimum})
		if err != nil || len(events) != 1 {
			t.Fatal(err)
		}
		p, err := satellite.TransitEventPath(g.Elements, events[0], time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, p)
	}
	return paths
}

func checkTextBoxes(t *testing.T, nodes []svgNode, width, height float64) {
	t.Helper()
	type box struct {
		x, y, w, h float64
		text       string
	}
	var boxes []box
	for _, n := range nodes {
		if n.Name != "text" {
			continue
		}
		x, y, size := number(n.Attr["x"]), number(n.Attr["y"]), number(n.Attr["font-size"])
		var w float64
		for _, r := range n.Text {
			if r < 128 {
				w += .62 * size
			} else {
				w += 1.05 * size
			}
		}
		switch n.Attr["text-anchor"] {
		case "middle":
			x -= w / 2
		case "end":
			x -= w
		}
		b := box{x, y - size*.82, w, size * 1.05, n.Text}
		if b.x < 0 || b.y < 0 || b.x+b.w > width || b.y+b.h > height {
			t.Fatalf("text out of bounds: %+v canvas %g %g", b, width, height)
		}
		for _, a := range boxes {
			if b.x < a.x+a.w && a.x < b.x+b.w && b.y < a.y+a.h && a.y < b.y+b.h {
				t.Fatalf("text overlap %q %q", a.text, b.text)
			}
		}
		boxes = append(boxes, b)
	}
}
func TestTransitMapSVGLayoutConsistency(t *testing.T) {
	paths := actualMapPaths(t)
	for i, p := range paths {
		for _, language := range []string{"zh", "en"} {
			for _, size := range [][2]int{{1200, 800}, {640, 480}} {
				o := TransitMapSVGOptions{Language: language, Width: size[0], Height: size[1]}
				if i == 0 {
					q := p.Event.GreatestPoint
					obs := satellite.Observer{Longitude: q.Longitude + .01, Latitude: q.Latitude}
					d, a := mapInverse(obs.Longitude, obs.Latitude, q.Longitude, q.Latitude)
					o.Observer = &obs
					o.Nearby = &satellite.NearbyTransit{Event: p.Event, Nearest: q, NearestDistanceKM: d, DistanceKM: -d / 2, BearingDeg: a}
				}
				value, err := TransitMapSVG(p, o)
				if err != nil {
					t.Fatal(i, err)
				}
				checkTextBoxes(t, parseDiagram(t, value), float64(size[0]), float64(size[1]))
			}
		}
	}
}
func TestTransitMapSVGInsetConsistency(t *testing.T) {
	p := mapSynthetic()
	q := p.CenterLine[1]
	obs := satellite.Observer{Longitude: 179.99, Latitude: 11.01}
	d, a := mapInverse(obs.Longitude, obs.Latitude, q.Longitude, q.Latitude)
	near := satellite.NearbyTransit{Event: p.Event, Nearest: q, NearestDistanceKM: d, DistanceKM: .3, BearingDeg: a}
	value, err := TransitMapSVG(p, TransitMapSVGOptions{Observer: &obs, Nearby: &near})
	if err != nil {
		t.Fatal(err)
	}
	nodes := parseDiagram(t, value)
	var scale, cx, cy float64
	var distance float64
	for _, n := range nodes {
		switch n.Attr["data-role"] {
		case "inset":
			scale = number(n.Attr["data-scale"])
			cx = number(n.Attr["data-cx"])
			cy = number(n.Attr["data-cy"])
		case "inset-nearest":
			distance = math.Hypot(number(n.Attr["cx"])-cx, number(n.Attr["cy"])-cy) / scale
		}
	}
	if math.Abs(distance-d) > 1e-6 {
		t.Fatal(distance, d)
	}
	near.Event.Complete = false
	if _, err := TransitMapSVG(p, TransitMapSVGOptions{Observer: &obs, Nearby: &near}); !errors.Is(err, satellite.ErrInvalidInput) {
		t.Fatal(err)
	}
}

func mapBaselineCases(t *testing.T) []struct {
	name string
	p    satellite.TransitPath
	o    TransitMapSVGOptions
} {
	p := actualMapPaths(t)
	q := p[0].Event.GreatestPoint
	obs := satellite.Observer{Longitude: q.Longitude + .01, Latitude: q.Latitude}
	d, a := mapInverse(obs.Longitude, obs.Latitude, q.Longitude, q.Latitude)
	near := satellite.NearbyTransit{Event: p[0].Event, Nearest: q, NearestDistanceKM: d, DistanceKM: -d / 2, BearingDeg: a}
	polar := MapProjectionNorthPolar
	if p[4].Event.GreatestPoint.Latitude < 0 {
		polar = MapProjectionSouthPolar
	}
	return []struct {
		name string
		p    satellite.TransitPath
		o    TransitMapSVGOptions
	}{
		{"map-equirect-zh-nearby", p[0], TransitMapSVGOptions{Observer: &obs, Nearby: &near, Projection: MapProjectionEquirectangular}},
		{"map-equirect-en", p[0], TransitMapSVGOptions{Language: "en", Projection: MapProjectionEquirectangular}},
		{"map-polar-zh", p[4], TransitMapSVGOptions{Projection: polar}},
		{"map-seam-zh", p[1], TransitMapSVGOptions{Projection: MapProjectionEquirectangular}},
		{"map-geo-long-en", p[3], TransitMapSVGOptions{Language: "en", Projection: MapProjectionEquirectangular}},
	}
}
func TestTransitMapSVGBaselineConsistency(t *testing.T) {
	if runtime.GOARCH == "386" && strings.HasPrefix(runtime.Version(), "go1.16") {
		t.Skip("platform byte baseline")
	}
	expected := map[string]struct {
		length int
		digest string
	}{
		"map-equirect-zh-nearby": {336891, "2e217ffd26ae65f625e8c025207a1930"},
		"map-equirect-en":        {328990, "a7162227622b331c4cac33483408a759"},
		"map-polar-zh":           {120598, "415ad70281e4fb100c4fa40b91062dff"},
		"map-seam-zh":            {309828, "2b73be05dc9eefbc982aa5b6a4e6a163"},
		"map-geo-long-en":        {314288, "ef6504bbebdd0fa6a3a017eacba49f7d"},
	}
	if runtime.GOARCH == "386" {
		// math.Pow 所用的 Exp 在 386 上有平台舍入差异。
		baseline := expected["map-geo-long-en"]
		baseline.digest = "abd010e61f81a2aa0a3809c23fb6615c"
		expected["map-geo-long-en"] = baseline
	}
	for _, v := range mapBaselineCases(t) {
		s, err := TransitMapSVG(v.p, v.o)
		if err != nil {
			t.Fatal(err)
		}
		sum := fmt.Sprintf("%x", md5.Sum([]byte(s)))
		t.Log(v.name, len(s), sum)
		if want := expected[v.name]; len(s) != want.length || sum != want.digest {
			t.Fatalf("%s byte baseline: %d %s", v.name, len(s), sum)
		}
		if dir := os.Getenv("ASTRO_C6_SVG_DIR"); dir != "" {
			if err := os.WriteFile(filepath.Join(dir, v.name+".svg"), []byte(s), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
}
func BenchmarkTransitMapSVG(b *testing.B) {
	p := actualMapPaths(b)[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := TransitMapSVG(p, TransitMapSVGOptions{Projection: MapProjectionEquirectangular}); err != nil {
			b.Fatal(err)
		}
	}
}

func TestTransitSVGTypographyConsistency(t *testing.T) {
	local := localTracks(t)[0]
	localSVG, err := TransitLocalSVG(local, TransitLocalSVGOptions{})
	if err != nil {
		t.Fatal(err)
	}
	groundSVG, err := TransitMapSVG(mapSynthetic(), TransitMapSVGOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{localSVG, groundSVG} {
		for _, n := range parseDiagram(t, value) {
			role := n.Attr["data-role"]
			size := number(n.Attr["font-size"])
			switch role {
			case "title":
				if size != 26 || n.Attr["font-weight"] != "700" || n.Attr["font-family"] != "Georgia, 'Times New Roman', serif" {
					t.Fatalf("title typography: %+v", n.Attr)
				}
			case "header":
				if size < 14 {
					t.Fatalf("small header: %g", size)
				}
			case "contact", "statistic", "panel-label", "panel-value":
				if size < 12 {
					t.Fatalf("small body %s: %g", role, size)
				}
			case "track-label", "time-label", "event-label", "footer", "stage-time":
				if size < 10 {
					t.Fatalf("small annotation %s: %g", role, size)
				}
			}
		}
	}
}

func TestTransitMapSVGValidationAndCompactLayoutSynthetic(t *testing.T) {
	base := mapSynthetic()
	badPoint := base
	badPoint.Event.GreatestPoint.HalfWidthKM = math.NaN()
	if _, err := TransitMapSVG(badPoint, TransitMapSVGOptions{}); err == nil {
		t.Fatal("expected error for NaN GreatestPoint.HalfWidthKM")
	}
	if _, err := TransitMapSVG(base, TransitMapSVGOptions{TimeLabelStep: time.Nanosecond}); err == nil {
		t.Fatal("expected error for excessive TimeLabelStep")
	}
	obs := satellite.Observer{Longitude: 11, Latitude: 20}
	badNear := satellite.NearbyTransit{Event: base.Event, Nearest: base.CenterLine[1], DistanceKM: 1, NearestDistanceKM: 2}
	badNear.Nearest.HalfWidthKM = math.NaN()
	if _, err := TransitMapSVG(base, TransitMapSVGOptions{Observer: &obs, Nearby: &badNear}); err == nil {
		t.Fatal("expected error for NaN Nearby.Nearest.HalfWidthKM")
	}
	goodNear := satellite.NearbyTransit{Event: base.Event, Nearest: base.CenterLine[1], DistanceKM: 1, NearestDistanceKM: 2}
	clipped := base
	clipped.Event.Complete = false
	goodNear.Event = clipped.Event
	loc := time.FixedZone("America/Los_Angeles", -8*3600)
	got, err := TransitMapSVG(clipped, TransitMapSVGOptions{
		Width: 640, Height: 480, Language: "EN", Location: loc,
		Title: strings.Repeat("Long Wrapped Transit Title ", 6), Observer: &obs, Nearby: &goodNear,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "Target altitude") || strings.Contains(got, `width="-`) || strings.Contains(got, "NaN") {
		t.Fatal("compact 640x480 transit map invalid")
	}
}
