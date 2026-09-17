package solarclosure

import (
	"math"
	"testing"

	"github.com/starainrt/astro/internal/geodata"
)

var testSubsolar = geodata.GeoPoint{Longitude: 12, Latitude: 34}

func testTerminator() []geodata.GeoPoint {
	return Terminator(testSubsolar)
}

// testFootprint 取地平圈上一段采样点当开放边界，两端擦地点各外扩两个采样点。
func testFootprint() Footprint {
	circle := testTerminator()
	return Footprint{
		Boundaries:  [][]geodata.GeoPoint{append([]geodata.GeoPoint(nil), circle[100:131]...)},
		HorizonEnds: []geodata.GeoPoint{circle[98], circle[133]},
		Subsolar:    testSubsolar,
	}
}

func angleFromSubsolar(point geodata.GeoPoint) float64 {
	cosine := math.Sin(testSubsolar.Latitude*math.Pi/180)*math.Sin(point.Latitude*math.Pi/180) +
		math.Cos(testSubsolar.Latitude*math.Pi/180)*math.Cos(point.Latitude*math.Pi/180)*
			math.Cos((point.Longitude-testSubsolar.Longitude)*math.Pi/180)
	return math.Acos(math.Max(-1, math.Min(1, cosine))) * 180 / math.Pi
}

func TestTerminatorPointsStayOnHorizonCircle(t *testing.T) {
	circle := testTerminator()
	if len(circle) != 360 {
		t.Fatalf("terminator points=%d, want 360", len(circle))
	}
	for index, point := range circle {
		if angle := angleFromSubsolar(point); math.Abs(angle-90) > 1e-9 {
			t.Fatalf("point %d is %.12f degrees from the subsolar point, want 90", index, angle)
		}
	}
}

func TestHorizonEndsPairsBothEndsByBoundaryStart(t *testing.T) {
	circle := testTerminator()
	for _, test := range []struct {
		name string
		ends []geodata.GeoPoint
		want []geodata.GeoPoint
	}{
		{name: "in order", ends: []geodata.GeoPoint{circle[98], circle[133]},
			want: []geodata.GeoPoint{circle[98], circle[133]}},
		{name: "reversed", ends: []geodata.GeoPoint{circle[133], circle[98]},
			want: []geodata.GeoPoint{circle[98], circle[133]}},
	} {
		footprint := testFootprint()
		footprint.HorizonEnds = test.ends
		got := HorizonEnds(footprint)
		if len(got) != 2 {
			t.Fatalf("%s: ends=%d, want 2", test.name, len(got))
		}
		for index := range got {
			if got[index] != test.want[index] {
				t.Fatalf("%s: ends[%d]=%v, want %v", test.name, index, got[index], test.want[index])
			}
		}
	}
}

func TestHorizonEndsRejectsIncompleteInput(t *testing.T) {
	circle := testTerminator()
	base := testFootprint()
	for _, test := range []struct {
		name      string
		footprint Footprint
	}{
		{name: "no ends", footprint: Footprint{Boundaries: base.Boundaries}},
		{name: "one end", footprint: Footprint{Boundaries: base.Boundaries,
			HorizonEnds: []geodata.GeoPoint{circle[98]}}},
		{name: "three ends", footprint: Footprint{Boundaries: base.Boundaries,
			HorizonEnds: []geodata.GeoPoint{circle[98], circle[133], circle[134]}}},
		{name: "no boundaries", footprint: Footprint{HorizonEnds: base.HorizonEnds}},
		{name: "empty first segment", footprint: Footprint{
			Boundaries: [][]geodata.GeoPoint{{}}, HorizonEnds: base.HorizonEnds}},
	} {
		if ends := HorizonEnds(test.footprint); ends != nil {
			t.Fatalf("%s: ends=%v, want nil", test.name, ends)
		}
		if ExactHorizon(test.footprint) {
			t.Fatalf("%s: ExactHorizon=true, want false", test.name)
		}
	}
	if !ExactHorizon(base) {
		t.Fatal("ExactHorizon=false for a footprint carrying two grazing points")
	}
}

func TestRingClosesOpenBoundaryAtGrazingPoints(t *testing.T) {
	footprint := testFootprint()
	curve := Curve(footprint)
	if len(curve) != 31 {
		t.Fatalf("curve points=%d, want 31", len(curve))
	}
	ends := HorizonEnds(footprint)
	ring, boundary, ok := Ring(footprint, true)
	if !ok {
		t.Fatal("Ring reported an unusable boundary")
	}
	if len(ring) <= len(curve) {
		t.Fatalf("ring points=%d, want more than the %d boundary points", len(ring), len(curve))
	}
	for index, point := range curve {
		if ring[index] != point {
			t.Fatalf("ring[%d]=%v, want the boundary point %v", index, ring[index], point)
		}
	}
	if ring[len(curve)] != ends[1] {
		t.Fatalf("ring[%d]=%v, want the trailing grazing point %v", len(curve), ring[len(curve)], ends[1])
	}
	if ring[len(ring)-1] != ends[0] {
		t.Fatalf("ring ends at %v, want the leading grazing point %v", ring[len(ring)-1], ends[0])
	}
	if len(boundary) != len(curve)+2 || boundary[0] != ends[0] ||
		boundary[len(boundary)-1] != ends[1] {
		t.Fatalf("boundary=%d points starting %v ending %v", len(boundary), boundary[0],
			boundary[len(boundary)-1])
	}
	for index, point := range boundary[1 : len(boundary)-1] {
		if point != curve[index] {
			t.Fatalf("boundary[%d]=%v, want %v", index+1, point, curve[index])
		}
	}
	for index, point := range ring {
		if angle := angleFromSubsolar(point); math.Abs(angle-90) > 1e-9 {
			t.Fatalf("ring[%d] is %.12f degrees from the subsolar point, want 90", index, angle)
		}
	}
}

func TestRingFallsBackToSampledTerminatorArc(t *testing.T) {
	footprint := testFootprint()
	curve := Curve(footprint)
	for _, test := range []struct {
		name  string
		input Footprint
		exact bool
	}{
		{name: "no grazing points", input: Footprint{
			Boundaries: footprint.Boundaries, Subsolar: testSubsolar}, exact: true},
		{name: "approximate requested", input: footprint, exact: false},
	} {
		ring, boundary, ok := Ring(test.input, test.exact)
		if !ok {
			t.Fatalf("%s: Ring reported an unusable boundary", test.name)
		}
		if len(ring) <= len(curve) || ring[len(ring)-1] != curve[0] {
			t.Fatalf("%s: ring closes at %v, want the boundary start %v",
				test.name, ring[len(ring)-1], curve[0])
		}
		for index, point := range curve {
			if ring[index] != point {
				t.Fatalf("%s: ring[%d]=%v, want %v", test.name, index, ring[index], point)
			}
		}
		if len(boundary) != len(curve) {
			t.Fatalf("%s: boundary points=%d, want the %d boundary points", test.name, len(boundary), len(curve))
		}
		for index, point := range ring {
			if angle := angleFromSubsolar(point); math.Abs(angle-90) > 1e-9 {
				t.Fatalf("%s: ring[%d] is %.12f degrees from the subsolar point, want 90",
					test.name, index, angle)
			}
		}
	}
}

func TestRingDegenerateInputs(t *testing.T) {
	circle := testTerminator()
	closed := []geodata.GeoPoint{circle[10], circle[40], circle[70], circle[10]}
	for _, test := range []struct {
		name      string
		footprint Footprint
		wantOK    bool
		wantRing  int
	}{
		{name: "no boundary", footprint: Footprint{Closed: true}, wantOK: false},
		{name: "closed triangle", footprint: Footprint{Boundaries: [][]geodata.GeoPoint{
			{circle[10], circle[40], circle[70]}}, Closed: true}, wantOK: true, wantRing: 3},
		{name: "closed with repeated point", footprint: Footprint{
			Boundaries: [][]geodata.GeoPoint{closed}, Closed: true}, wantOK: true, wantRing: 3},
		{name: "closed segment", footprint: Footprint{Boundaries: [][]geodata.GeoPoint{
			{circle[10], circle[40]}}, Closed: true}, wantOK: false},
		{name: "single open point", footprint: Footprint{Boundaries: [][]geodata.GeoPoint{
			{circle[10]}}}, wantOK: true, wantRing: 1},
		{name: "open segment without grazing points", footprint: Footprint{
			Boundaries: [][]geodata.GeoPoint{{circle[10], circle[40]}},
			Subsolar:   testSubsolar}, wantOK: true},
	} {
		ring, _, ok := Ring(test.footprint, true)
		if ok != test.wantOK {
			t.Fatalf("%s: ok=%v, want %v", test.name, ok, test.wantOK)
		}
		if test.wantRing > 0 && len(ring) != test.wantRing {
			t.Fatalf("%s: ring points=%d, want %d", test.name, len(ring), test.wantRing)
		}
	}
}

func TestHorizonRingAcceptsPrejoinedCurve(t *testing.T) {
	footprint := testFootprint()
	curve := Curve(footprint)
	ring, boundary := HorizonRing(footprint, curve)
	ends := HorizonEnds(footprint)
	if len(boundary) != len(curve)+2 || boundary[0] != ends[0] ||
		boundary[len(boundary)-1] != ends[1] {
		t.Fatalf("boundary=%d points, want the leading and trailing grazing points around %d points",
			len(boundary), len(curve))
	}
	if len(ring) < len(curve)+2 || ring[len(curve)] != ends[1] || ring[len(ring)-1] != ends[0] {
		t.Fatalf("ring=%d points, want the grazing points closing %d boundary points",
			len(ring), len(curve))
	}
}

func TestBandPolygonsSelectsFaceCoveredByFootprint(t *testing.T) {
	box := [][2]geodata.GeoPoint{
		{{Longitude: 0, Latitude: 0}, {Longitude: 10, Latitude: 0}},
		{{Longitude: 10, Latitude: 0}, {Longitude: 10, Latitude: 10}},
		{{Longitude: 10, Latitude: 10}, {Longitude: 0, Latitude: 10}},
		{{Longitude: 0, Latitude: 10}, {Longitude: 0, Latitude: 0}},
	}
	contours := make([][]geodata.GeoPoint, 0, len(box))
	for _, edge := range box {
		contours = append(contours, []geodata.GeoPoint{edge[0], edge[1]})
	}
	inside := []geodata.GeoPoint{
		{Longitude: 4, Latitude: 4}, {Longitude: 6, Latitude: 4},
		{Longitude: 6, Latitude: 6}, {Longitude: 4, Latitude: 6},
	}
	footprints := []Footprint{{Boundaries: [][]geodata.GeoPoint{inside}, Closed: true}}
	polygons, ok := BandPolygons(contours, nil, footprints, true, SnapDistanceKM)
	if !ok || len(polygons) == 0 {
		t.Fatalf("band polygons unavailable: ok=%v count=%d", ok, len(polygons))
	}
	probes := []geodata.GeoPoint{{Longitude: 5, Latitude: 5}, {Longitude: 1, Latitude: 9}}
	contained := geodata.SphericalPolygonsContainPoints(polygons, probes)
	if len(contained) != 2 || !contained[0] || !contained[1] {
		t.Fatalf("containment=%v, want the whole box face selected", contained)
	}
	outside := geodata.SphericalPolygonsContainPoints(polygons,
		[]geodata.GeoPoint{{Longitude: 15, Latitude: 5}})
	if outside[0] {
		t.Fatal("a point outside the boundary network was selected")
	}
}

func TestBandPolygonsRejectsEmptyBoundaryNetwork(t *testing.T) {
	if polygons, ok := BandPolygons(nil, nil, nil, true, SnapDistanceKM); ok || polygons != nil {
		t.Fatalf("polygons=%d ok=%v, want no polygons", len(polygons), ok)
	}
}
