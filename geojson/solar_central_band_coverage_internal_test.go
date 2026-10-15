package geojson

import (
	"math"
	"testing"
	"time"

	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/geodata"
)

// 本文件钉住两条导出契约：覆盖判据必须逐点成立（不能靠抽样），
// 中心线走廊必须让每个探针都落在容差内，且只修补越界的那一段。
const centralBandToleranceKM = solarCentralBandCoverageToleranceKM

func centralBandSquareRing(minLongitude, minLatitude, maxLongitude, maxLatitude float64) []geodata.GeoPoint {
	return []geodata.GeoPoint{
		{Longitude: minLongitude, Latitude: minLatitude},
		{Longitude: maxLongitude, Latitude: minLatitude},
		{Longitude: maxLongitude, Latitude: maxLatitude},
		{Longitude: minLongitude, Latitude: maxLatitude},
	}
}

func centralBandCenterLine(points ...[2]float64) []eclipsecore.SolarEclipsePathPoint {
	base := time.Date(2024, time.April, 8, 18, 0, 0, 0, time.UTC)
	line := make([]eclipsecore.SolarEclipsePathPoint, len(points))
	for index, point := range points {
		line[index] = eclipsecore.SolarEclipsePathPoint{
			Time:      base.Add(time.Duration(index) * time.Minute),
			Longitude: point[0],
			Latitude:  point[1],
		}
	}
	return line
}

func TestSolarCentralBandPointsCoverProbesEveryVertex(t *testing.T) {
	polygons := [][]geodata.GeoPoint{centralBandSquareRing(0, 0, 1, 1)}
	points := make([]geodata.GeoPoint, 0, 257)
	for index := 0; index < 257; index++ {
		points = append(points, geodata.GeoPoint{
			Longitude: 0.1 + 0.8*float64(index)/256,
			Latitude:  0.1 + 0.8*float64(index%97)/96,
		})
	}
	if !solarCentralBandPointsCover(polygons, points, centralBandToleranceKM) {
		t.Fatal("points inside the ring must be covered")
	}
	for index := range points {
		moved := append([]geodata.GeoPoint(nil), points...)
		moved[index] = geodata.GeoPoint{Longitude: 40, Latitude: 40}
		if solarCentralBandPointsCover(polygons, moved, centralBandToleranceKM) {
			t.Fatalf("vertex %d outside the tolerance was accepted; every vertex must be probed", index)
		}
	}
}

func TestSolarCentralBandPointsCoverKeepsMacroTolerance(t *testing.T) {
	polygons := [][]geodata.GeoPoint{centralBandSquareRing(0, 0, 1, 1)}
	near := []geodata.GeoPoint{{Longitude: 0.5, Latitude: 1.4}}
	far := []geodata.GeoPoint{{Longitude: 0.5, Latitude: 2.0}}
	if !solarCentralBandPointsCover(polygons, near, centralBandToleranceKM) {
		t.Fatal("a point 45 km outside must stay inside the 100 km macro tolerance")
	}
	if solarCentralBandPointsCover(polygons, far, centralBandToleranceKM) {
		t.Fatal("a point 111 km outside must exceed the 100 km macro tolerance")
	}
}

func TestSolarCentralBandRingsCoverProbesEveryFootprintVertex(t *testing.T) {
	rings := [][]geodata.GeoPoint{centralBandSquareRing(0, 0, 1, 1)}
	stamp := time.Date(2024, time.April, 8, 18, 0, 0, 0, time.UTC)
	boundary := make([]eclipsecore.SolarEclipsePathPoint, 0, 64)
	for index := 0; index < 64; index++ {
		boundary = append(boundary, eclipsecore.SolarEclipsePathPoint{
			Time:      stamp,
			Longitude: 0.2 + 0.6*float64(index)/63,
			Latitude:  0.3,
		})
	}
	footprints := []eclipsecore.SolarEclipsePartialFootprint{{
		Time:       stamp,
		Boundaries: [][]eclipsecore.SolarEclipsePathPoint{boundary},
	}}
	if !solarCentralBandRingsCover(rings, centralBandCenterLine([2]float64{0.5, 0.5}), footprints) {
		t.Fatal("a covered footprint must pass the coverage check")
	}
	for index := range boundary {
		moved := append([]eclipsecore.SolarEclipsePathPoint(nil), boundary...)
		moved[index] = eclipsecore.SolarEclipsePathPoint{Time: stamp, Longitude: 40, Latitude: 40}
		probe := []eclipsecore.SolarEclipsePartialFootprint{{
			Time:       stamp,
			Boundaries: [][]eclipsecore.SolarEclipsePathPoint{moved},
		}}
		if solarCentralBandRingsCover(rings, centralBandCenterLine([2]float64{0.5, 0.5}), probe) {
			t.Fatalf("footprint vertex %d outside the tolerance was accepted", index)
		}
	}
}

func TestSolarCentralBandRingsCoverProbesEveryCenterLineVertex(t *testing.T) {
	rings := [][]geodata.GeoPoint{centralBandSquareRing(0, 0, 1, 1)}
	line := centralBandCenterLine([2]float64{0.5, 0.5}, [2]float64{0.5, 0.6}, [2]float64{0.5, 0.7})
	if !solarCentralBandRingsCover(rings, line, nil) {
		t.Fatal("a covered center line must pass the coverage check")
	}
	for index := range line {
		moved := append([]eclipsecore.SolarEclipsePathPoint(nil), line...)
		moved[index].Longitude = 40
		moved[index].Latitude = 40
		if solarCentralBandRingsCover(rings, moved, nil) {
			t.Fatalf("center-line vertex %d outside the tolerance was accepted", index)
		}
	}
}

func TestSolarCentralBandCorridorCoversEveryCenterlineProbe(t *testing.T) {
	polygons := [][]geodata.GeoPoint{centralBandSquareRing(0, 0, 1, 1)}
	line := centralBandCenterLine([2]float64{0.5, 1.2}, [2]float64{0.5, 1.8})
	repaired := solarCentralBandWithCenterlineCorridor(polygons, line)
	probes := solarCentralBandCenterlineProbes(line)
	miss := geodata.SphericalPolygonsPathMissDistanceKM(
		repaired, [][]geodata.GeoPoint{probes}, false,
	)
	if miss > solarCentralBandCenterlineToleranceKM {
		t.Fatalf("center-line probe sits %.1f km from the repaired band, want <= %.1f km",
			miss, solarCentralBandCenterlineToleranceKM)
	}
}

func TestSolarCentralBandCorridorPatchesOnlyTheClippedSegment(t *testing.T) {
	polygons := [][]geodata.GeoPoint{centralBandSquareRing(-1, -0.05, 1, 0.05)}
	line := centralBandCenterLine([2]float64{-0.5, 0.08}, [2]float64{0.5, 0.6})
	repaired := solarCentralBandWithCenterlineCorridor(polygons, line)
	witness := []geodata.GeoPoint{{Longitude: -0.5, Latitude: 0.35}}
	if covered := geodata.SphericalPolygonsContainPoints(repaired, witness); covered[0] {
		t.Fatal("the repaired band inflated the shallow end to the deepest probe's radius")
	}
}

func TestSolarCentralBandCorridorLeavesUnboundedMissAlone(t *testing.T) {
	polygons := [][]geodata.GeoPoint{centralBandSquareRing(0, 0, 1, 1)}
	line := centralBandCenterLine([2]float64{0.5, 0.5}, [2]float64{0.5, 20})
	repaired := solarCentralBandWithCenterlineCorridor(polygons, line)
	if len(repaired) != len(polygons) {
		t.Fatalf("polygons=%d, want the input unchanged when the miss exceeds the corridor cap", len(repaired))
	}
	for index := range polygons {
		if len(repaired[index]) != len(polygons[index]) {
			t.Fatalf("ring %d has %d vertices, want %d", index, len(repaired[index]), len(polygons[index]))
		}
	}
}

func TestSolarShadowRegionDegenerateSkipsOnlyDegenerateRings(t *testing.T) {
	normal := centralBandSquareRing(0, 0, 0.1, 0.1)
	if solarShadowRegionDegenerate(normal, normal) {
		t.Fatal("an 11 km square must not be degenerate")
	}
	sliver := []geodata.GeoPoint{
		{Longitude: 0, Latitude: 0},
		{Longitude: 0.01, Latitude: 0},
		{Longitude: 0.01, Latitude: 1e-11},
		{Longitude: 0, Latitude: 1e-11},
	}
	if !solarShadowRegionDegenerate(sliver, sliver) {
		t.Fatal("a 2.2 km long sliver with 1e-13 square degrees of area must be degenerate")
	}
	short := []geodata.GeoPoint{{Longitude: 0, Latitude: 0}, {Longitude: 0.001, Latitude: 0}}
	if !solarShadowRegionDegenerate(short, short) {
		t.Fatal("a 0.1 km boundary must be degenerate")
	}
	crossing := []geodata.GeoPoint{
		{Longitude: 179.9, Latitude: -0.05},
		{Longitude: -179.9, Latitude: -0.05},
		{Longitude: -179.9, Latitude: 0.05},
		{Longitude: 179.9, Latitude: 0.05},
	}
	if solarShadowRegionDegenerate(crossing, crossing) {
		t.Fatal("a region crossing the antimeridian must not be degenerate")
	}
}

func TestSolarShadowSegmentClosedUsesBasicLayerCaliber(t *testing.T) {
	start := eclipsecore.SolarEclipsePathPoint{Longitude: 12, Latitude: 30}
	nearClosed := []eclipsecore.SolarEclipsePathPoint{
		start, {Longitude: 12.5, Latitude: 30.5}, {Longitude: 12 + 5e-9, Latitude: 30 + 5e-9},
	}
	if !solarShadowSegmentClosed(nearClosed) {
		t.Fatal("a 0.8 mm gap counts as closed for the basic layer signature")
	}
	open := []eclipsecore.SolarEclipsePathPoint{
		start, {Longitude: 12.5, Latitude: 30.5}, {Longitude: 12.001, Latitude: 30.001},
	}
	if solarShadowSegmentClosed(open) {
		t.Fatal("a 150 m gap is not a closed ring")
	}
	if solarShadowSegmentClosed(nearClosed[:2]) {
		t.Fatal("a two-point segment is not a closed ring")
	}
}

func TestOccultationBandSourcePropertiesCoverBothPaths(t *testing.T) {
	for _, testCase := range []struct {
		authoritative bool
		contours      int
		source        string
		boundary      string
	}{
		{authoritative: true, contours: 2, source: "visible-footprint-sweep", boundary: "footprint-sweep+horizon-visible"},
		{authoritative: false, contours: 2, source: "footprint-sweep-fallback", boundary: "contact-contours+horizon-boundary"},
		{authoritative: false, contours: 0, source: "footprint-sweep-fallback", boundary: ""},
	} {
		properties := map[string]interface{}{}
		applyOccultationBandSourceProperties(properties, testCase.authoritative, testCase.contours)
		if properties["source"] != testCase.source {
			t.Fatalf("source=%v, want %s", properties["source"], testCase.source)
		}
		boundary, present := properties["boundary_source"]
		if testCase.boundary == "" {
			if present {
				t.Fatalf("boundary_source=%v, want absent", boundary)
			}
			continue
		}
		if boundary != testCase.boundary {
			t.Fatalf("boundary_source=%v, want %s", boundary, testCase.boundary)
		}
	}
}

func TestAppendSolarFootprintFeaturesSkipsDegenerateFootprint(t *testing.T) {
	stamp := time.Date(2024, time.April, 8, 18, 0, 0, 0, time.UTC)
	degenerate := eclipsecore.SolarEclipsePartialFootprint{
		Time:   stamp,
		Closed: true,
		Boundaries: [][]eclipsecore.SolarEclipsePathPoint{{
			{Time: stamp, Longitude: 0, Latitude: 0},
			{Time: stamp, Longitude: 0.01, Latitude: 0},
			{Time: stamp, Longitude: 0.01, Latitude: 1e-11},
			{Time: stamp, Longitude: 0, Latitude: 1e-11},
		}},
	}
	features, err := appendSolarFootprintFeatures(
		nil, solarCentralShadowFootprintRole,
		[]eclipsecore.SolarEclipsePartialFootprint{degenerate},
		map[string]interface{}{},
	)
	if err != nil {
		t.Fatalf("degenerate footprint must be omitted, not rejected: %v", err)
	}
	if len(features) != 0 {
		t.Fatalf("features=%d, want the degenerate footprint omitted", len(features))
	}
}

// centralBandProjectTime 独立复算顶点在中心线上的投影时刻。
func centralBandProjectTime(
	point eclipsecore.SolarEclipsePathPoint,
	centerLine []eclipsecore.SolarEclipsePathPoint,
) time.Time {
	bestDistance := math.Inf(1)
	bestTime := centerLine[0].Time
	scale := math.Cos(point.Latitude * math.Pi / 180)
	for index := 0; index+1 < len(centerLine); index++ {
		first, second := centerLine[index], centerLine[index+1]
		ax := math.Remainder(first.Longitude-point.Longitude, 360) * scale
		ay := first.Latitude - point.Latitude
		bx := math.Remainder(second.Longitude-point.Longitude, 360) * scale
		by := second.Latitude - point.Latitude
		dx, dy := bx-ax, by-ay
		length := dx*dx + dy*dy
		fraction := 0.0
		if length > 0 {
			fraction = math.Max(0, math.Min(1, -(ax*dx+ay*dy)/length))
		}
		distance := math.Hypot(ax+fraction*dx, ay+fraction*dy)
		if distance >= bestDistance {
			continue
		}
		bestDistance = distance
		bestTime = first.Time.Add(time.Duration(float64(second.Time.Sub(first.Time)) * fraction))
	}
	return bestTime
}

func TestSolarCentralBandLimitSidesKeepProjectionTimes(t *testing.T) {
	base := time.Date(2024, time.April, 8, 18, 0, 0, 0, time.UTC)
	centerLine := []eclipsecore.SolarEclipsePathPoint{
		{Time: base, Longitude: 0, Latitude: 0},
		{Time: base.Add(time.Minute), Longitude: 0.3, Latitude: 0},
		{Time: base.Add(20 * time.Minute), Longitude: 1, Latitude: 0},
	}
	ring := []eclipsecore.SolarEclipsePathPoint{
		{Longitude: 0, Latitude: 0.1},
		{Longitude: 0.3, Latitude: 0.1},
		{Longitude: 1, Latitude: 0.1},
		{Longitude: 1.05, Latitude: -0.05},
		{Longitude: 1, Latitude: -0.1},
		{Longitude: 0.3, Latitude: -0.1},
		{Longitude: 0, Latitude: -0.1},
		{Longitude: -0.05, Latitude: -0.05},
	}
	north, south, ok := solarCentralBandLimitSidesFromRings(
		[][]eclipsecore.SolarEclipsePathPoint{ring}, centerLine,
	)
	if !ok {
		t.Fatal("synthetic band must produce two limit sides")
	}
	if len(north) != 3 || len(south) != 5 {
		t.Fatalf("north=%d south=%d, want the 3 north and 5 south vertices without the end caps",
			len(north), len(south))
	}
	for _, side := range []struct {
		name   string
		north  bool
		points []eclipsecore.SolarEclipsePathPoint
	}{{"north-limit", true, north}, {"south-limit", false, south}} {
		for index, point := range side.points {
			if side.north && point.Latitude <= 0 {
				t.Fatalf("%s vertex %d latitude=%v, want north of the center line", side.name, index, point.Latitude)
			}
			if !side.north && point.Latitude >= 0 {
				t.Fatalf("%s vertex %d latitude=%v, want south of the center line", side.name, index, point.Latitude)
			}
			projected := centralBandProjectTime(point, centerLine)
			if delta := point.Time.Sub(projected); delta > time.Millisecond || delta < -time.Millisecond {
				t.Fatalf("%s vertex %d time=%v, projected=%v", side.name, index, point.Time, projected)
			}
		}
	}
}

// 掠食端中心线回折时，界线顶点的最近点投影会落到回折支上（1552-07-21）。
// 投影沿环连续追踪后，顶点不得被丢弃，界线也不得跳点抄近路。
func TestSolarCentralBandLimitSidesSurviveFoldedCenterLine(t *testing.T) {
	base := time.Date(1552, time.July, 21, 6, 0, 0, 0, time.UTC)
	centerLine := []eclipsecore.SolarEclipsePathPoint{
		{Time: base, Longitude: 0, Latitude: 0},
		{Time: base.Add(5 * time.Minute), Longitude: 0.5, Latitude: 0},
		{Time: base.Add(10 * time.Minute), Longitude: 1, Latitude: 0},
		{Time: base.Add(12 * time.Minute), Longitude: 1, Latitude: 0.25},
		{Time: base.Add(15 * time.Minute), Longitude: 0.6, Latitude: 0.25},
	}
	northVertices := []eclipsecore.SolarEclipsePathPoint{
		{Longitude: 0, Latitude: 0.15},
		{Longitude: 0.3, Latitude: 0.15},
		// 这两点离回折支（0.10°）比离去程支（0.15°）更近，仍须判到北侧。
		{Longitude: 0.65, Latitude: 0.15},
		{Longitude: 0.8, Latitude: 0.15},
		{Longitude: 0.95, Latitude: 0.15},
	}
	southVertices := []eclipsecore.SolarEclipsePathPoint{
		{Longitude: 0.95, Latitude: -0.15},
		{Longitude: 0.5, Latitude: -0.15},
		{Longitude: 0, Latitude: -0.15},
	}
	ring := append([]eclipsecore.SolarEclipsePathPoint(nil), northVertices...)
	ring = append(ring, eclipsecore.SolarEclipsePathPoint{Longitude: 1.05, Latitude: 0})
	ring = append(ring, southVertices...)
	ring = append(ring, eclipsecore.SolarEclipsePathPoint{Longitude: -0.05, Latitude: 0})
	north, south, ok := solarCentralBandLimitSidesFromRings(
		[][]eclipsecore.SolarEclipsePathPoint{ring}, centerLine,
	)
	if !ok {
		t.Fatal("folded center line must still produce two limit sides")
	}
	for _, side := range []struct {
		name   string
		want   []eclipsecore.SolarEclipsePathPoint
		points []eclipsecore.SolarEclipsePathPoint
	}{{"north-limit", northVertices, north}, {"south-limit", southVertices, south}} {
		for _, vertex := range side.want {
			found := false
			for _, point := range side.points {
				found = found || point.Longitude == vertex.Longitude && point.Latitude == vertex.Latitude
			}
			if !found {
				t.Fatalf("%s lost boundary vertex (%v,%v): %+v", side.name, vertex.Longitude, vertex.Latitude, side.points)
			}
		}
		for index := 1; index < len(side.points); index++ {
			if !side.points[index].Time.After(side.points[index-1].Time) {
				t.Fatalf("%s times not increasing at %d: %+v", side.name, index, side.points)
			}
		}
	}
	for _, point := range north {
		if point.Latitude < 0 {
			t.Fatalf("north-limit took southern vertex (%v,%v)", point.Longitude, point.Latitude)
		}
	}
	for _, point := range south {
		if point.Latitude > 0 {
			t.Fatalf("south-limit took northern vertex (%v,%v)", point.Longitude, point.Latitude)
		}
	}
}
