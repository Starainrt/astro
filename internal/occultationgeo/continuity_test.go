package occultationgeo

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

func TestOccultationVisibleFootprintFillOnlyMatchesFullFill(t *testing.T) {
	visible := []basic.OccultationPathPoint{
		{Longitude: 0, Latitude: 0, MoonAltitude: -1},
		{Longitude: 2, Latitude: 0, MoonAltitude: 1},
		{Longitude: 2, Latitude: 2, MoonAltitude: 1},
		{Longitude: 0, Latitude: 2, MoonAltitude: -1},
	}
	interior := []basic.OccultationPathPoint{
		{Longitude: 1.2, Latitude: 0.5, MoonAltitude: 1},
		{Longitude: 1.8, Latitude: 0.5, MoonAltitude: 1},
		{Longitude: 1.5, Latitude: 1.5, MoonAltitude: 1},
	}
	polar := []basic.OccultationPathPoint{
		{Longitude: 0, Latitude: 80, MoonAltitude: 1},
		{Longitude: 120, Latitude: 80, MoonAltitude: 1},
		{Longitude: -120, Latitude: 80, MoonAltitude: 1},
	}
	for _, footprints := range [][]basic.OccultationFootprint{
		nil,
		{{Polygons: [][]basic.OccultationPathPoint{visible, interior, visible[:2]}, InteriorPolygons: [][]basic.OccultationPathPoint{interior}}},
		{{Polygons: [][]basic.OccultationPathPoint{polar}}},
	} {
		want, _ := occultationVisibleFillAndCoverage(footprints)
		if got := occultationVisibleFootprintFillOnly(footprints); !reflect.DeepEqual(got, want) {
			t.Fatalf("fill-only changed the horizon-clipped source: got=%v want=%v", got, want)
		}
	}
}

func TestContinuousBoundaryRangesRetainsEndpointSingletons(t *testing.T) {
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	points := []basic.OccultationPathPoint{
		{Time: start, Longitude: 0, Latitude: 10},
		{Time: start.Add(time.Second), Longitude: 30, Latitude: 10},
		{Time: start.Add(2 * time.Second), Longitude: 30.1, Latitude: 10},
		{Time: start.Add(3 * time.Second), Longitude: 0, Latitude: 10},
	}
	ranges := ContinuousBoundaryRanges(points)
	want := []SampleRange{{Start: 0, End: 1}, {Start: 1, End: 3}, {Start: 3, End: 4}}
	if len(ranges) != len(want) {
		t.Fatalf("range count = %d, want %d: %#v", len(ranges), len(want), ranges)
	}
	for index := range want {
		if ranges[index] != want[index] {
			t.Fatalf("range %d = %#v, want %#v", index, ranges[index], want[index])
		}
	}
}

func TestStitchedRiseSetCurveSegmentsJoinsSharedFoldEndpoint(t *testing.T) {
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, time.UTC)
	fold := basic.OccultationPathPoint{Time: start.Add(2 * time.Minute), Longitude: 1, Latitude: 1}
	curve := basic.OccultationRiseSetCurve{
		Phase:     basic.RiseSetPhaseEnd,
		Direction: basic.RiseSetDirectionRise,
		Segments: [][]basic.OccultationPathPoint{
			{
				{Time: start, Longitude: 0, Latitude: 0},
				fold,
			},
			{
				{Time: start.Add(time.Minute), Longitude: 0, Latitude: 2},
				fold,
			},
		},
	}
	segments := StitchedRiseSetCurveSegments(curve)
	if len(segments) != 1 {
		t.Fatalf("stitched segment count=%d, want one continuous display segment", len(segments))
	}
	line := segments[0]
	if len(line) != 3 {
		t.Fatalf("stitched point count=%d, want the shared fold once", len(line))
	}
	if line[1] != fold {
		t.Fatalf("middle point=%+v, want shared fold %+v", line[1], fold)
	}
}

func TestDistanceKMUsesShortestAntimeridianArc(t *testing.T) {
	first := basic.OccultationPathPoint{Longitude: 179.9}
	second := basic.OccultationPathPoint{Longitude: -179.9}
	if distance := DistanceKM(first, second); distance > 25 {
		t.Fatalf("antimeridian distance = %.1f km, want shortest arc", distance)
	}
}

func TestValidateRiseSetCurvesRejectsPolarSubsecondBranchJump(t *testing.T) {
	start := time.Date(2025, time.January, 5, 0, 0, 0, 0, time.UTC)
	curve := basic.OccultationRiseSetCurve{
		Phase:     basic.RiseSetPhaseStart,
		Direction: basic.RiseSetDirectionSet,
		Segments: [][]basic.OccultationPathPoint{{
			{Time: start, Longitude: -12.591916, Latitude: 81.864617},
			{Time: start.Add(40 * time.Millisecond), Longitude: 22.206242, Latitude: 79.125287},
		}},
	}
	if err := ValidateRiseSetCurves([]basic.OccultationRiseSetCurve{curve}, start, start.Add(time.Hour)); err == nil {
		t.Fatal("ValidateRiseSetCurves accepted a polar subsecond branch jump")
	}
}

func TestPairedBoundaryPolygonsDoNotBridgeBranchChanges(t *testing.T) {
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	first := []basic.OccultationPathPoint{
		{Time: start, Longitude: 0, Latitude: 10},
		{Time: start.Add(time.Minute), Longitude: 1, Latitude: 10},
		{Time: start.Add(2 * time.Minute), Longitude: 90, Latitude: 10},
		{Time: start.Add(3 * time.Minute), Longitude: 91, Latitude: 10},
	}
	second := []basic.OccultationPathPoint{
		{Time: start, Longitude: 0, Latitude: -10},
		{Time: start.Add(time.Minute), Longitude: 1, Latitude: -10},
		{Time: start.Add(2 * time.Minute), Longitude: 90, Latitude: -10},
		{Time: start.Add(3 * time.Minute), Longitude: 91, Latitude: -10},
	}
	polygons := PairedBoundaryPolygons(first, second)
	if len(polygons) != 2 {
		t.Fatalf("polygon count = %d, want two cells separated at the branch change", len(polygons))
	}
	if polygons[0][0].Longitude != 0 || polygons[0][1].Longitude != 1 ||
		polygons[1][0].Longitude != 90 || polygons[1][1].Longitude != 91 {
		t.Fatalf("unexpected paired cells: %#v", polygons)
	}
}

func TestPairedBoundaryPolygonsSuppressOneSidedBranchInterval(t *testing.T) {
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	first := []basic.OccultationPathPoint{
		{Time: start, Longitude: 0, Latitude: 10},
		{Time: start.Add(time.Minute), Longitude: 1, Latitude: 10},
		{Time: start.Add(2 * time.Minute), Longitude: 2, Latitude: 10},
		{Time: start.Add(3 * time.Minute), Longitude: 90, Latitude: 10},
		{Time: start.Add(4 * time.Minute), Longitude: 91, Latitude: 10},
	}
	second := []basic.OccultationPathPoint{
		{Time: start, Longitude: 0, Latitude: -10},
		{Time: start.Add(time.Minute), Longitude: 1, Latitude: -10},
		{Time: start.Add(2 * time.Minute), Longitude: 90, Latitude: -10},
		{Time: start.Add(3 * time.Minute), Longitude: 91, Latitude: -10},
		{Time: start.Add(4 * time.Minute), Longitude: 92, Latitude: -10},
	}
	polygons := PairedBoundaryPolygons(first, second)
	if len(polygons) != 2 {
		t.Fatalf("polygon count = %d, want two valid cells around the one-sided branch interval", len(polygons))
	}
	if polygons[0][0].Longitude != 0 || polygons[0][1].Longitude != 1 ||
		polygons[1][0].Longitude != 90 || polygons[1][1].Longitude != 91 {
		t.Fatalf("unexpected polygons: %#v", polygons)
	}
}

func TestRemoveTinyPolygonComponentsKeepsComparableBranches(t *testing.T) {
	polygons := [][]geodata.GeoPoint{
		{{Longitude: 0, Latitude: 0}, {Longitude: 10, Latitude: 0}, {Longitude: 10, Latitude: 10}, {Longitude: 0, Latitude: 10}},
		{{Longitude: 20, Latitude: 0}, {Longitude: 30, Latitude: 0}, {Longitude: 30, Latitude: 10}, {Longitude: 20, Latitude: 10}},
		{{Longitude: 179.99, Latitude: 0}, {Longitude: -179.99, Latitude: 0}, {Longitude: -179.99, Latitude: 1}, {Longitude: 179.99, Latitude: 1}},
		{{Longitude: 0, Latitude: 0}, {Longitude: 0.01, Latitude: 0}, {Longitude: 0.01, Latitude: 0.01}},
	}
	filtered := RemoveTinyPolygonComponents(polygons)
	if len(filtered) != 3 {
		t.Fatalf("filtered component count = %d, want two large and one dateline component", len(filtered))
	}
}

func TestFootprintTransitionBoundaryHandlesCoincidentNearestSample(t *testing.T) {
	when := time.Date(2025, time.January, 5, 0, 12, 17, 0, time.UTC)
	closed := basic.OccultationFootprint{
		Time:   when,
		Closed: true,
		Boundaries: [][]basic.OccultationPathPoint{{
			{Longitude: -86.16, Latitude: 30.50},
			{Longitude: -86.10, Latitude: 30.56},
			{Longitude: -86.04, Latitude: 30.50},
			{Longitude: -86.10, Latitude: 30.44},
		}},
	}
	open := basic.OccultationFootprint{
		Time: when.Add(-50 * time.Millisecond),
		Boundaries: [][]basic.OccultationPathPoint{{
			{Longitude: -86.12, Latitude: 30.51},
			{Longitude: -86.10, Latitude: 30.56},
			{Longitude: -86.04, Latitude: 30.50},
			{Longitude: -86.10, Latitude: 30.44},
			{Longitude: -86.12, Latitude: 30.49},
		}},
	}
	boundary, ok := footprintTransitionBoundary(closed, open)
	if !ok {
		t.Fatal("coincident nearest sample did not produce a transition boundary")
	}
	if len(boundary) < len(closed.Boundaries[0])-1 {
		t.Fatalf("transition boundary points=%d, want the long physical arc", len(boundary))
	}
}

func TestRemoveOccultationHairpinsKeepsMainOutline(t *testing.T) {
	points := []geodata.GeoPoint{
		{Longitude: -86.3, Latitude: 31.0},
		{Longitude: -86.16, Latitude: 30.61},
		{Longitude: -85.97, Latitude: 30.55},
		{Longitude: -85.74, Latitude: 30.24},
		{Longitude: -86.07, Latitude: 30.43},
		{Longitude: -86.18, Latitude: 30.47},
		{Longitude: -86.4, Latitude: 29.1},
	}
	cleaned := removeOccultationHairpins(points, 35, 25, 12)
	if len(cleaned) != 4 {
		t.Fatalf("cleaned point count=%d, want the main outline and hairpin join: %#v", len(cleaned), cleaned)
	}
	if cleaned[0] != points[0] || cleaned[len(cleaned)-1] != points[len(points)-1] {
		t.Fatalf("hairpin cleanup changed main outline endpoints: %#v", cleaned)
	}
}

func TestSmoothOccultationHairpinsKeepsShortDisplayEdges(t *testing.T) {
	points := []geodata.GeoPoint{
		{Longitude: -137.238657, Latitude: -75.841965},
		{Longitude: -135.756780, Latitude: -76.038909},
		{Longitude: -134.234342, Latitude: -76.226631},
		{Longitude: -133.715901, Latitude: -76.650415},
		{Longitude: -133.164148, Latitude: -77.073078},
		{Longitude: -134.361125, Latitude: -77.215854},
	}
	smoothed := smoothOccultationHairpins(points, 180, 25, 16, 50)
	maximumEdge := 0.0
	maximumDetour := 0.0
	for index := 1; index < len(smoothed); index++ {
		maximumEdge = math.Max(maximumEdge, geodata.EquatorialDistanceKM(smoothed[index-1], smoothed[index]))
	}
	for index := 1; index+1 < len(smoothed); index++ {
		maximumDetour = math.Max(maximumDetour,
			geodata.EquatorialDistanceKM(smoothed[index-1], smoothed[index])+
				geodata.EquatorialDistanceKM(smoothed[index], smoothed[index+1])-
				geodata.EquatorialDistanceKM(smoothed[index-1], smoothed[index+1]))
	}
	if maximumEdge > 50.01 {
		t.Fatalf("smoothed maximum edge=%.1f km, want <=50 km", maximumEdge)
	}
	if maximumDetour > 25 {
		t.Fatalf("smoothed maximum local detour=%.1f km, want <=25 km", maximumDetour)
	}
}

func TestSmoothOccultationPolarWobblesRemovesShortLatitudeSeam(t *testing.T) {
	points := []geodata.GeoPoint{
		{Longitude: 10.19691, Latitude: 80.96484},
		{Longitude: 9.18809, Latitude: 81.05666},
		{Longitude: 8.61181, Latitude: 80.98198},
		{Longitude: 7.76426, Latitude: 81.01962},
	}
	smoothed := smoothOccultationPolarWobbles(points)
	if len(smoothed) != len(points) || smoothed[0] != points[0] || smoothed[len(smoothed)-1] != points[len(points)-1] {
		t.Fatalf("polar wobble smoothing changed endpoints or point count: %#v", smoothed)
	}
	for index := 1; index+1 < len(smoothed); index++ {
		firstDelta := math.Remainder(smoothed[index].Longitude-smoothed[index-1].Longitude, 360)
		secondDelta := math.Remainder(smoothed[index+1].Longitude-smoothed[index].Longitude, 360)
		if firstDelta*secondDelta <= 0 {
			t.Fatalf("polar seam retains longitude reversal at %d: %#v", index, smoothed)
		}
		firstLatitudeDelta := smoothed[index].Latitude - smoothed[index-1].Latitude
		secondLatitudeDelta := smoothed[index+1].Latitude - smoothed[index].Latitude
		if firstLatitudeDelta*secondLatitudeDelta < 0 {
			t.Fatalf("polar seam retains latitude reversal at %d: %#v", index, smoothed)
		}
	}
}

func TestSmoothOccultationOrdinaryWobblesHandlesHighLatitudeArc(t *testing.T) {
	points := []geodata.GeoPoint{
		{Longitude: -30.0, Latitude: 80.000},
		{Longitude: -29.8, Latitude: 80.020},
		{Longitude: -29.6, Latitude: 80.011},
		{Longitude: -29.4, Latitude: 80.031},
		{Longitude: -29.2, Latitude: 80.021},
		{Longitude: -29.0, Latitude: 80.041},
		{Longitude: -28.8, Latitude: 80.032},
		{Longitude: -28.6, Latitude: 80.052},
		{Longitude: -28.4, Latitude: 80.043},
	}
	smoothed := smoothOccultationOrdinaryWobbles(points)
	if len(smoothed) != len(points) {
		t.Fatalf("high-latitude smoothing changed point count: got %d want %d", len(smoothed), len(points))
	}
	if smoothed[0] != points[0] || smoothed[len(smoothed)-1] != points[len(points)-1] {
		t.Fatal("high-latitude smoothing changed arc endpoints")
	}
	changed := false
	for index := 1; index+1 < len(points); index++ {
		if geodata.EquatorialDistanceKM(points[index], smoothed[index]) > 0.001 {
			changed = true
			if geodata.EquatorialDistanceKM(points[index], smoothed[index]) > 20 {
				t.Fatalf("high-latitude smoothing moved point %d by %.2f km", index, geodata.EquatorialDistanceKM(points[index], smoothed[index]))
			}
		}
	}
	if !changed {
		t.Fatal("high-latitude smoothing did not remove the synthetic local wobble")
	}
}

func TestDensifyOccultationPolygonsIncludesClosingEdge(t *testing.T) {
	points := []geodata.GeoPoint{
		{Longitude: 0, Latitude: 0},
		{Longitude: 1, Latitude: 0},
		{Longitude: 1, Latitude: 1},
		{Longitude: 0, Latitude: 1},
	}
	densified := densifyOccultationPolygons([][]geodata.GeoPoint{points}, 50)
	ring := densified[0]
	maximumEdge := geodata.EquatorialDistanceKM(ring[len(ring)-1], ring[0])
	if maximumEdge > 50.01 {
		t.Fatalf("closing edge=%.1f km, want <=50 km", maximumEdge)
	}
}

func TestVisibleBandPolygons20250105SaturnExcludesInvisiblePolarCap(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.January, 5, 0, 0, 0, 0, zone)
	paths, err := basic.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), basic.OccultationSaturn,
		basic.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	t.Logf("total footprints=%d contours=%d northern=%d southern=%d curves=%d", len(path.TotalBandFootprints), len(path.TotalBandContours), len(path.NorthernTotalLimit), len(path.SouthernTotalLimit), len(path.TotalRiseSetCurves))
	for index, contour := range path.TotalBandContours {
		if len(contour) == 0 {
			continue
		}
		t.Logf("contour %d len=%d first=(%.4f,%.4f) last=(%.4f,%.4f)", index, len(contour), contour[0].Longitude, contour[0].Latitude, contour[len(contour)-1].Longitude, contour[len(contour)-1].Latitude)
	}
	visibleSite := []geodata.GeoPoint{{Longitude: -20, Latitude: 81.95}}
	visibleFootprints := 0
	for _, footprint := range path.PartialBandFootprints {
		for _, source := range footprint.Polygons {
			polygon := make([]geodata.GeoPoint, len(source))
			for index, point := range source {
				polygon[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			if geodata.SphericalPolygonsContainPaths([][]geodata.GeoPoint{polygon}, [][]geodata.GeoPoint{visibleSite}, false) {
				visibleFootprints++
			}
		}
	}
	if visibleFootprints == 0 {
		t.Fatal("visible polar site is absent from every instantaneous footprint")
	}
	polygons, authoritative, err := VisibleBandPolygons(
		path.PartialBandFootprints, path.NorthernLimit, path.SouthernLimit, path.RiseSetCurves,
	)
	if err != nil {
		t.Fatalf("VisibleBandPolygons: %v", err)
	}
	if !authoritative {
		t.Fatal("VisibleBandPolygons fell back to the footprint sweep")
	}
	if geodata.SphericalPolygonsContainPaths(polygons, [][]geodata.GeoPoint{{
		{Longitude: -20, Latitude: 82.10},
	}}, false) {
		t.Fatal("visible band contains the polar site whose occultation remains below the lunar horizon")
	}
	for _, site := range []geodata.GeoPoint{
		{Longitude: -20, Latitude: 81.95},
		{Longitude: -20, Latitude: 81.70},
	} {
		if !geodata.SphericalPolygonsContainPaths(polygons, [][]geodata.GeoPoint{{site}}, false) {
			t.Fatalf("visible band excludes the site %.2f, %.2f whose occultation is above the lunar horizon", site.Longitude, site.Latitude)
		}
	}
}

func TestVisibleBandPolygons20250105SaturnEnvelopeCandidates(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.January, 5, 0, 0, 0, 0, zone)
	paths, err := basic.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), basic.OccultationSaturn,
		basic.OccultationPathOptions{
			// Keep this legacy fallback regression on its original exact input.
			// The current analytic output is exercised for both branches below.
			Algorithm: basic.OccultationPathAlgorithmExact,
			Step:      20 * time.Minute, TargetSpacingKM: 900,
			DisableFootprints: true, IncludeFootprintTimeline: true,
			FootprintTimelineStep: 5 * time.Minute, RiseSetStep: time.Minute,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	final, _, err := VisibleBandPolygonsFromContours(
		path.PartialBandFootprints, path.PartialBandContours,
		path.NorthernLimit, path.SouthernLimit, path.RiseSetCurves,
	)
	if err != nil {
		t.Fatalf("VisibleBandPolygonsFromContours: %v", err)
	}
	sweep, err := footprintSweepPolygons(
		path.PartialBandFootprints, path.NorthernLimit, path.SouthernLimit,
	)
	if err != nil {
		t.Fatalf("footprintSweepPolygons: %v", err)
	}
	visibleUnion := footprintVisibleUnionPolygons(path.PartialBandFootprints)
	direct, _ := DirectVisibleBandPolygons(path.PartialBandFootprints)
	contourLines := occultationContactContourBoundaryLines(path.PartialBandContours)
	phase, _ := occultationPreferredPhaseBand(
		true, path.PartialBandFootprints, path.RiseSetCurves,
		path.NorthernLimit, path.SouthernLimit, PairedBoundaryPolygons(path.NorthernLimit, path.SouthernLimit),
	)
	for name, polygons := range map[string][][]geodata.GeoPoint{
		"final": final, "sweep": sweep, "visible-union": visibleUnion, "direct": direct,
		"contours": contourLines, "phase": phase,
	} {
		t.Logf("%s polygons=%d points=%d roughness=%v", name, len(polygons),
			occultationTestPolygonPointCount(polygons), occultationTestEnvelopeRoughness(polygons))
	}
	// 首尾端帽由瞬时足迹的端帽/桥接构造而来，本就有折角，不属于普通纬度平滑的检查范围。
	// 接触弧端点（地平线切点）邻域同样按构造排除：可见带在那里转入月球地平线闭合弧，
	// 度量到的是切点转折的采样弦高，不是普通纬度抖动。
	horizons := occultationTestFootprintHorizonPoints(path.PartialBandFootprints)
	roughness := occultationTestEnvelopeRoughnessExcludingHorizon(
		occultationTestTrimRingTail(final, occultationTestEnvelopeCapPoints),
		horizons, occultationTestEnvelopeHorizonWindowKM,
	)
	t.Logf("asserted final roughness=%v horizons=%d", roughness, len(horizons))
	if roughness["bottom"] > 6 || roughness["back"] > 10 {
		t.Fatalf("smoothed Saturn ordinary-latitude envelope remains rough: %v", roughness)
	}
	if below, checked := occultationTestBelowHorizonVertices(path.PartialBandFootprints); below != 0 || checked < 1000 {
		t.Fatalf("band footprints hold %d of %d vertices below the lunar horizon", below, checked)
	}
	for _, polygon := range final {
		for index := 1; index < len(polygon); index++ {
			if distance := geodata.EquatorialDistanceKM(polygon[index-1], polygon[index]); distance > BoundaryBranchJumpKM {
				t.Fatalf("band ring spans a %.1f km edge", distance)
			}
		}
	}
}

// occultationTestEnvelopeCapPoints 是平滑度断言从环尾剔除的端帽顶点数。
const occultationTestEnvelopeCapPoints = 48

// occultationTestTrimRingTail 返回去掉环尾端帽窗口后的副本。
func occultationTestTrimRingTail(polygons [][]geodata.GeoPoint, capPoints int) [][]geodata.GeoPoint {
	trimmed := make([][]geodata.GeoPoint, 0, len(polygons))
	for _, polygon := range polygons {
		if len(polygon) <= 2*capPoints+4 {
			continue
		}
		trimmed = append(trimmed, polygon[:len(polygon)-capPoints])
	}
	return trimmed
}

func TestVisibleBandPolygonsSaturnAnalyticAlgorithmBranches(t *testing.T) {
	start := time.Date(2025, time.January, 5, 0, 0, 0, 0, time.FixedZone("UTC+8", 8*3600))
	for _, algorithm := range []basic.OccultationPathAlgorithm{basic.OccultationPathAlgorithmExact, basic.OccultationPathAlgorithmOptimized} {
		t.Run(string(algorithm), func(t *testing.T) {
			paths, err := basic.FindPlanetOccultationPaths(start, start.Add(24*time.Hour), basic.OccultationSaturn,
				basic.OccultationPathOptions{
					Algorithm: algorithm, Step: 20 * time.Minute, TargetSpacingKM: 900,
					DisableFootprints: true, IncludeFootprintTimeline: true,
					FootprintTimelineStep: 5 * time.Minute, RiseSetStep: time.Minute,
				})
			if err != nil || len(paths) != 1 {
				t.Fatalf("paths=%d err=%v, want one", len(paths), err)
			}
			path := paths[0]
			if len(path.PartialVisibilityContours) == 0 {
				t.Fatal("missing analytic visibility contours")
			}
			polygons, authoritative, err := VisibleBandPolygonsFromAnalyticContours(
				path.PartialBandFootprints, path.PartialBandContours, path.PartialVisibilityContours,
				path.NorthernLimit, path.SouthernLimit, path.RiseSetCurves,
			)
			if err != nil || !authoritative || len(polygons) != 1 {
				t.Fatalf("analytic polygons=%d authoritative=%v err=%v", len(polygons), authoritative, err)
			}
			roughness := occultationTestEnvelopeRoughness(polygons)
			if roughness["bottom"] > 6 || roughness["back"] > 10 {
				t.Fatalf("analytic Saturn envelope remains rough: %v", roughness)
			}
		})
	}
}

func TestSmoothOccultationOrdinaryWobblesPreservesEndpointsAndBoundedMotion(t *testing.T) {
	ring := []geodata.GeoPoint{
		{Longitude: -82.0, Latitude: -4.0},
		{Longitude: -81.0, Latitude: -5.5},
		{Longitude: -80.0, Latitude: -7.0},
		{Longitude: -79.0, Latitude: -6.6},
		{Longitude: -78.0, Latitude: -7.5},
		{Longitude: -77.0, Latitude: -8.0},
		{Longitude: -76.0, Latitude: -7.8},
		{Longitude: -75.0, Latitude: -7.0},
	}
	cleaned := smoothOccultationOrdinaryWobbles(ring)
	if cleaned[0] != ring[0] || cleaned[len(cleaned)-1] != ring[len(ring)-1] {
		t.Fatalf("ordinary smoothing changed open endpoints: %#v", cleaned)
	}
	for index, point := range cleaned {
		if move := geodata.EquatorialDistanceKM(point, ring[index]); move > 20.001 {
			t.Fatalf("point %d moved %.2f km, want <=20 km", index, move)
		}
	}
}

func occultationTestPolygonPointCount(polygons [][]geodata.GeoPoint) int {
	count := 0
	for _, polygon := range polygons {
		count += len(polygon)
	}
	return count
}

func occultationTestEnvelopeRoughness(polygons [][]geodata.GeoPoint) map[string]float64 {
	return occultationTestEnvelopeRoughnessExcludingHorizon(polygons, nil, 0)
}

// occultationTestEnvelopeHorizonWindowKM 是粗糙度度量跳过地平线切点邻域的半径：
// 可见带在切点处从接触包络转入月球地平线闭合弧，那里的折角是构造边界而不是普通纬度抖动。
const occultationTestEnvelopeHorizonWindowKM = 20.0

// occultationTestEnvelopeRoughnessExcludingHorizon 跳过起点或终点落在切点邻域内的五采样窗口。
func occultationTestEnvelopeRoughnessExcludingHorizon(
	polygons [][]geodata.GeoPoint,
	horizons []geodata.GeoPoint,
	windowKM float64,
) map[string]float64 {
	result := map[string]float64{"bottom": 0, "back": 0, "polar": 0}
	for _, polygon := range polygons {
		for index := 2; index+2 < len(polygon); index++ {
			point := polygon[index]
			region := ""
			switch {
			case point.Latitude < 5:
				region = "bottom"
			case point.Longitude < -75 && point.Latitude >= 25 && point.Latitude <= 45:
				region = "back"
			case point.Latitude > 75:
				region = "polar"
			default:
				continue
			}
			if occultationTestWindowTouchesHorizon(polygon, index, horizons, windowKM) {
				continue
			}
			deviation := occultationTestProjectedPointLineDistanceKM(
				point, polygon[index-2], polygon[index+2],
			)
			result[region] = math.Max(result[region], deviation)
		}
	}
	return result
}

func occultationTestWindowTouchesHorizon(
	polygon []geodata.GeoPoint,
	index int,
	horizons []geodata.GeoPoint,
	windowKM float64,
) bool {
	if windowKM <= 0 || len(horizons) == 0 {
		return false
	}
	for offset := -2; offset <= 2; offset++ {
		for _, horizon := range horizons {
			if geodata.EquatorialDistanceKM(polygon[index+offset], horizon) <= windowKM {
				return true
			}
		}
	}
	return false
}

// occultationTestFootprintHorizonPoints 返回接触弧两端的地平线切点，作为"地平线切点邻域"的构造锚点。
func occultationTestFootprintHorizonPoints(footprints []basic.OccultationFootprint) []geodata.GeoPoint {
	points := make([]geodata.GeoPoint, 0, len(footprints)*2)
	for _, footprint := range footprints {
		if footprint.Closed {
			continue
		}
		for _, boundary := range footprint.Boundaries {
			if len(boundary) == 0 {
				continue
			}
			for _, index := range []int{0, len(boundary) - 1} {
				points = append(points, geodata.GeoPoint{
					Longitude: boundary[index].Longitude,
					Latitude:  boundary[index].Latitude,
				})
			}
		}
	}
	return points
}

// occultationTestBelowHorizonVertices 统计足迹边界、闭合多边形与修复面中位于月球地平线以下的顶点数。
func occultationTestBelowHorizonVertices(footprints []basic.OccultationFootprint) (below, checked int) {
	for _, footprint := range footprints {
		for _, rings := range [][][]basic.OccultationPathPoint{
			footprint.Boundaries, footprint.Polygons, footprint.InteriorPolygons,
		} {
			for _, ring := range rings {
				for _, point := range ring {
					checked++
					if point.MoonAltitude < 0 {
						below++
					}
				}
			}
		}
	}
	return below, checked
}

func occultationTestProjectedPointLineDistanceKM(point, first, last geodata.GeoPoint) float64 {
	project := func(value geodata.GeoPoint) (float64, float64) {
		const radiusKM = 6378.1366
		latitude := math.Max(-85.05112878, math.Min(85.05112878, value.Latitude)) * math.Pi / 180
		return radiusKM * value.Longitude * math.Pi / 180,
			radiusKM * math.Log(math.Tan(math.Pi/4+latitude/2))
	}
	px, py := project(point)
	ax, ay := project(first)
	bx, by := project(last)
	dx, dy := bx-ax, by-ay
	if lengthSquared := dx*dx + dy*dy; lengthSquared > 0 {
		fraction := ((px-ax)*dx + (py-ay)*dy) / lengthSquared
		fraction = math.Max(0, math.Min(1, fraction))
		return math.Hypot(px-(ax+fraction*dx), py-(ay+fraction*dy))
	}
	return math.Hypot(px-ax, py-ay)
}

func TestVisibleBandPolygons20250729MarsStitchesEndRiseFold(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	paths, err := basic.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), basic.OccultationMars,
		basic.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: 5 * time.Minute,
			DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	if len(path.CenterLine) != 0 {
		t.Fatalf("center-line points=%d, want non-central path", len(path.CenterLine))
	}
	if len(path.RiseSetCurves) != 3 {
		t.Fatalf("rise/set curve count=%d, want three rising phase curves", len(path.RiseSetCurves))
	}
	endCurve := path.RiseSetCurves[2]
	if len(endCurve.Segments) < 2 {
		t.Fatalf("end-rise segment count=%d, want the polar fold branches", len(endCurve.Segments))
	}
	polygons, authoritative, err := VisibleBandPolygons(
		path.PartialBandFootprints, path.NorthernLimit, path.SouthernLimit, path.RiseSetCurves,
	)
	if err != nil {
		t.Fatalf("VisibleBandPolygons: %v", err)
	}
	if !authoritative {
		t.Fatal("VisibleBandPolygons fell back to the footprint sweep")
	}
	if len(polygons) != 1 {
		t.Fatalf("visible band polygon count=%d, want one continuous polar band", len(polygons))
	}
	for footprintIndex, footprint := range path.PartialBandFootprints {
		for polygonIndex, source := range footprint.Polygons {
			for pointIndex, point := range source {
				if point.MoonAltitude < -1e-6 && geodata.SphericalPolygonsContainPaths(
					polygons, [][]geodata.GeoPoint{{{Longitude: point.Longitude, Latitude: point.Latitude}}}, false,
				) {
					t.Fatalf("visible band contains below-horizon footprint %d polygon %d point %d", footprintIndex, polygonIndex, pointIndex)
				}
			}
		}
	}
	if miss := geodata.SphericalPolygonsPathMissDistanceKM(polygons, [][]geodata.GeoPoint{{
		{Longitude: path.PartialBandFootprints[len(path.PartialBandFootprints)/2].Polygons[0][0].Longitude,
			Latitude: path.PartialBandFootprints[len(path.PartialBandFootprints)/2].Polygons[0][0].Latitude},
	}}, true); miss > 112.5 {
		t.Fatalf("visible band misses a source footprint vertex by %.1f km", miss)
	}
	if !geodata.SphericalPolygonsContainPaths(polygons, [][]geodata.GeoPoint{{
		{Longitude: -110, Latitude: -75},
	}}, false) {
		t.Fatal("visible band misses the eastern footprint-sweep region")
	}
	for polygonIndex, polygon := range polygons {
		maximumDetour := 0.0
		minimumTurn := 180.0
		for index := 1; index+1 < len(polygon); index++ {
			first, middle, last := polygon[index-1], polygon[index], polygon[index+1]
			detour := geodata.EquatorialDistanceKM(first, middle) + geodata.EquatorialDistanceKM(middle, last) - geodata.EquatorialDistanceKM(first, last)
			if detour > maximumDetour {
				maximumDetour = detour
			}
			firstLongitude := first.Longitude - middle.Longitude
			firstLatitude := first.Latitude - middle.Latitude
			lastLongitude := last.Longitude - middle.Longitude
			lastLatitude := last.Latitude - middle.Latitude
			firstLength := math.Hypot(firstLongitude, firstLatitude)
			lastLength := math.Hypot(lastLongitude, lastLatitude)
			if firstLength <= 1e-12 || lastLength <= 1e-12 {
				continue
			}
			cosine := (firstLongitude*lastLongitude + firstLatitude*lastLatitude) /
				(firstLength * lastLength)
			cosine = math.Max(-1, math.Min(1, cosine))
			minimumTurn = math.Min(minimumTurn, math.Acos(cosine)*180/math.Pi)
		}
		if maximumDetour > 35 {
			t.Fatalf("visible band polygon %d retains %.1f km local hairpin", polygonIndex, maximumDetour)
		}
		if minimumTurn < 30 {
			t.Fatalf("visible band polygon %d retains a %.1f degree numerical corner", polygonIndex, minimumTurn)
		}
	}
}

func TestVisibleBandPolygonsFromContours20250729MarsUsesAnalyticBoundaryNetwork(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	paths, err := basic.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), basic.OccultationMars,
		basic.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
			DisableFootprints: true, IncludeFootprintTimeline: true,
			FootprintTimelineStep: 5 * time.Minute,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	polygons, authoritative, err := VisibleBandPolygonsFromAnalyticContours(
		path.PartialBandFootprints, path.PartialBandContours, path.PartialVisibilityContours,
		path.NorthernLimit, path.SouthernLimit, path.RiseSetCurves,
	)
	if err != nil {
		t.Fatalf("VisibleBandPolygonsFromContours: %v", err)
	}
	if !authoritative || len(polygons) != 1 {
		t.Fatalf("visible band authoritative=%v rings=%d, want one authoritative ring", authoritative, len(polygons))
	}
	partialBoundary := occultationVisibleBoundaryLinesFromBase(
		occultationContactContourBoundaryLines(path.PartialBandContours),
		occultationStaticBandCurves(path.RiseSetCurves),
		occultationContactContourBoundaryLines(path.PartialVisibilityContours),
	)
	partialBoundary = append(partialBoundary, occultationHorizonConnectorBoundaryLines(
		HorizonConnectorSegments(path.PartialBandFootprints, path.RiseSetCurves,
			path.NorthernLimit, path.SouthernLimit),
	)...)
	assertOccultationPolygonUsesOnlyBoundaryNetwork(t, "partial", polygons, partialBoundary, 1)
	phaseLines := make([][]geodata.GeoPoint, 0, len(path.RiseSetCurves)*2)
	for _, curve := range path.RiseSetCurves {
		phaseLines = append(phaseLines, occultationCurveBoundaryLines(curve)...)
	}
	if !geodata.SphericalPolygonsContainPathsWithinKM(polygons, phaseLines, false, 1) {
		miss := geodata.SphericalPolygonsPathMissDistanceKM(polygons, phaseLines, false)
		t.Fatalf("static partial band misses a physical rise/set phase curve by %.3f km, want <=1 km", miss)
	}
	assertOccultationPolarExtremeSmooth(t, "partial", polygons)
	if !geodata.SphericalPolygonsContainPaths(polygons, [][]geodata.GeoPoint{{
		{Longitude: -110, Latitude: -75},
	}}, false) {
		t.Fatal("static band misses the eastern event-time footprint")
	}
	totalPolygons, totalAuthoritative, err := VisibleTotalBandPolygonsFromAnalyticContours(
		path.TotalBandFootprints, path.TotalBandContours, path.TotalVisibilityContours,
		path.NorthernTotalLimit, path.SouthernTotalLimit, path.TotalRiseSetCurves,
	)
	if err != nil {
		t.Fatalf("VisibleTotalBandPolygonsFromContours: %v", err)
	}
	if !totalAuthoritative || len(totalPolygons) != 1 {
		t.Fatalf("total visible band authoritative=%v rings=%d, want one authoritative ring", totalAuthoritative, len(totalPolygons))
	}
	totalBoundary := occultationVisibleBoundaryLinesFromBase(
		occultationContactContourBoundaryLines(path.TotalBandContours),
		occultationStaticBandCurves(path.TotalRiseSetCurves),
		occultationContactContourBoundaryLines(path.TotalVisibilityContours),
	)
	totalBoundary = append(totalBoundary, occultationHorizonConnectorBoundaryLines(
		HorizonConnectorSegments(path.TotalBandFootprints, path.TotalRiseSetCurves,
			path.NorthernTotalLimit, path.SouthernTotalLimit),
	)...)
	assertOccultationPolygonUsesOnlyBoundaryNetwork(t, "total", totalPolygons, totalBoundary, 1)
	totalPhaseLines := make([][]geodata.GeoPoint, 0, len(path.TotalRiseSetCurves)*2)
	for _, curve := range path.TotalRiseSetCurves {
		totalPhaseLines = append(totalPhaseLines, occultationCurveBoundaryLines(curve)...)
	}
	if !geodata.SphericalPolygonsContainPathsWithinKM(totalPolygons, totalPhaseLines, false, 1) {
		miss := geodata.SphericalPolygonsPathMissDistanceKM(totalPolygons, totalPhaseLines, false)
		t.Fatalf("static total band misses a physical rise/set phase curve by %.3f km, want <=1 km", miss)
	}
	assertOccultationPolarExtremeSmooth(t, "total", totalPolygons)
	if miss := geodata.SphericalPolygonsPathMissDistanceKM(polygons, totalPolygons, true); miss > 10 {
		t.Fatalf("total band extends %.1f km beyond the partial band", miss)
	}
}

func TestVisibleBandPolygons20250105SaturnUsesCompleteAnalyticBoundaryNetwork(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.January, 5, 0, 0, 0, 0, zone)
	paths, err := basic.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), basic.OccultationSaturn,
		basic.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
			DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	partial, partialAuthoritative, err := VisibleBandPolygonsFromAnalyticContours(
		path.PartialBandFootprints, path.PartialBandContours, path.PartialVisibilityContours,
		path.NorthernLimit, path.SouthernLimit, path.RiseSetCurves,
	)
	if err != nil || !partialAuthoritative || len(partial) != 1 {
		t.Fatalf("partial polygons=%d authoritative=%v err=%v, want one analytic polygon", len(partial), partialAuthoritative, err)
	}
	total, totalAuthoritative, err := VisibleTotalBandPolygonsFromAnalyticContours(
		path.TotalBandFootprints, path.TotalBandContours, path.TotalVisibilityContours,
		path.NorthernTotalLimit, path.SouthernTotalLimit, path.TotalRiseSetCurves,
	)
	if err != nil || !totalAuthoritative || len(total) != 1 {
		t.Fatalf("total polygons=%d authoritative=%v err=%v, want one analytic polygon", len(total), totalAuthoritative, err)
	}
	for _, band := range []struct {
		name       string
		polygons   [][]geodata.GeoPoint
		contacts   [][]basic.OccultationPathPoint
		visibility [][]basic.OccultationPathPoint
		curves     []basic.OccultationRiseSetCurve
	}{
		{name: "partial", polygons: partial, contacts: path.PartialBandContours, visibility: path.PartialVisibilityContours, curves: path.RiseSetCurves},
		{name: "total", polygons: total, contacts: path.TotalBandContours, visibility: path.TotalVisibilityContours, curves: path.TotalRiseSetCurves},
	} {
		boundary := occultationVisibleBoundaryLinesFromBase(
			occultationContactContourBoundaryLines(band.contacts),
			occultationStaticBandCurves(band.curves),
			occultationContactContourBoundaryLines(band.visibility),
		)
		if band.name == "partial" {
			boundary = append(boundary, occultationHorizonConnectorBoundaryLines(
				HorizonConnectorSegments(path.PartialBandFootprints, band.curves,
					path.NorthernLimit, path.SouthernLimit),
			)...)
		} else {
			boundary = append(boundary, occultationHorizonConnectorBoundaryLines(
				HorizonConnectorSegments(path.TotalBandFootprints, band.curves,
					path.NorthernTotalLimit, path.SouthernTotalLimit),
			)...)
		}
		assertOccultationPolygonUsesOnlyBoundaryNetwork(t, band.name, band.polygons, boundary, 1)
		phaseLines := make([][]geodata.GeoPoint, 0, len(band.curves)*2)
		for _, curve := range band.curves {
			phaseLines = append(phaseLines, occultationCurveBoundaryLines(curve)...)
		}
		if !geodata.SphericalPolygonsContainPathsWithinKM(band.polygons, phaseLines, false, 1) {
			t.Fatalf("%s misses a rise/set phase by %.3f km", band.name,
				geodata.SphericalPolygonsPathMissDistanceKM(band.polygons, phaseLines, false))
		}
	}
	if miss := geodata.SphericalPolygonsPathMissDistanceKM(partial, total, true); miss > 10 {
		t.Fatalf("total band extends %.1f km beyond the partial band", miss)
	}
}

func assertOccultationPolygonUsesOnlyBoundaryNetwork(
	t *testing.T,
	name string,
	polygons, lines [][]geodata.GeoPoint,
	maximumDistanceKM float64,
) {
	t.Helper()
	maximum := 0.0
	maximumPoint := geodata.GeoPoint{}
	for _, polygon := range polygons {
		for _, point := range polygon {
			distance := math.Inf(1)
			for _, line := range lines {
				for index := 1; index < len(line); index++ {
					distance = math.Min(distance, occultationTestProjectedPointLineDistanceKM(
						point, line[index-1], line[index],
					))
				}
			}
			if distance > maximum {
				maximum, maximumPoint = distance, point
			}
		}
	}
	if maximum > maximumDistanceKM {
		t.Fatalf("%s boundary leaves the analytic contact/phase network by %.3f km at %+v, want <=%.1f km",
			name, maximum, maximumPoint, maximumDistanceKM)
	}
}

func assertOccultationPolarExtremeSmooth(
	t *testing.T,
	name string,
	polygons [][]geodata.GeoPoint,
) {
	t.Helper()
	minimumLatitude := math.Inf(1)
	for _, polygon := range polygons {
		for _, point := range polygon {
			minimumLatitude = math.Min(minimumLatitude, point.Latitude)
		}
	}
	if minimumLatitude > -75 {
		return
	}
	minimumTurn := 180.0
	maximumDeviation := 0.0
	minimumTurnPoint := geodata.GeoPoint{}
	for _, polygon := range polygons {
		for index := 2; index+2 < len(polygon); index++ {
			if polygon[index].Latitude > minimumLatitude+0.15 {
				continue
			}
			turn := occultationTurnAngleDegrees(
				polygon[index-1], polygon[index], polygon[index+1],
			)
			if turn < minimumTurn {
				minimumTurn = turn
				minimumTurnPoint = polygon[index]
			}
			maximumDeviation = math.Max(maximumDeviation, occultationProjectedPointLineDistanceKM(
				polygon[index], polygon[index-2], polygon[index+2],
			))
		}
	}
	if minimumTurn < 155 {
		t.Fatalf("%s polar phase junction turn=%.2f degrees at %+v, want >=155", name, minimumTurn, minimumTurnPoint)
	}
	if maximumDeviation > 10 {
		t.Fatalf("%s polar phase junction deviation=%.2f km, want <=10 km", name, maximumDeviation)
	}
}

func TestOccultationPhaseBoundaryCycleExcludesGreatestInteriorLine(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	paths, err := basic.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), basic.OccultationMars,
		basic.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
			DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	contourFill, _ := occultationContourFillAndCoverage(
		true, path.NorthernLimit, path.SouthernLimit, nil,
	)
	cycles, ok := occultationPhaseBoundaryPolygons(
		path.RiseSetCurves,
		HorizonConnectorSegments(
			path.PartialBandFootprints, path.RiseSetCurves,
			path.NorthernLimit, path.SouthernLimit,
		),
		contourFill,
	)
	if !ok || len(cycles) != 1 {
		t.Fatalf("phase boundary cycles=%d ok=%v, want one closed Mars outer cycle", len(cycles), ok)
	}
	greatest := [][]geodata.GeoPoint{{
		{Longitude: path.Greatest.Longitude, Latitude: path.Greatest.Latitude},
	}}
	if !geodata.SphericalPolygonsContainPaths(cycles, greatest, false) {
		t.Fatal("greatest point is outside the explicit start/end phase cycle")
	}
	if below := geodata.SphericalPolygonsPathMissDistanceKM(
		cycles, contourFill, true,
	); below > 150 {
		t.Fatalf("phase cycle misses contact envelope by %.1f km", below)
	}
	phaseLines := make([][]geodata.GeoPoint, 0, len(path.RiseSetCurves)*2)
	for _, curve := range path.RiseSetCurves {
		if curve.Phase == basic.RiseSetPhaseGreatest {
			continue
		}
		phaseLines = append(phaseLines, occultationCurveBoundaryLines(curve)...)
	}
	for _, connector := range HorizonConnectorSegments(
		path.PartialBandFootprints, path.RiseSetCurves,
		path.NorthernLimit, path.SouthernLimit,
	) {
		phaseLines = append(phaseLines, occultationPathGeoLine(connector.Points))
	}
	for pointIndex, point := range cycles[0] {
		if pointIndex == len(cycles[0])-1 && geodata.SameGeoPoint(point, cycles[0][0]) {
			continue
		}
		minimum := math.Inf(1)
		for _, line := range phaseLines {
			for _, candidate := range line {
				minimum = math.Min(minimum, geodata.EquatorialDistanceKM(point, candidate))
			}
		}
		if minimum > 1 {
			t.Fatalf("cycle vertex %d is %.1f km away from start/end/connector phase data", pointIndex, minimum)
		}
	}
}

func TestRemoveOccultationPolarKinksRemovesShortAxisReversals(t *testing.T) {
	points := []geodata.GeoPoint{
		{Longitude: -135.15725570344884, Latitude: -78.65055730057036},
		{Longitude: -135.02235091247215, Latitude: -78.70666016247198},
		{Longitude: -134.7363589035614, Latitude: -78.69259306299674},
		{Longitude: -134.45107516883687, Latitude: -78.67825217368357},
	}
	cleaned := removeOccultationPolarKinks(points)
	if len(cleaned) >= len(points) {
		t.Fatalf("cleaned point count=%d, want fewer than %d", len(cleaned), len(points))
	}
	if cleaned[0] != points[0] || cleaned[len(cleaned)-1] != points[len(points)-1] {
		t.Fatalf("cleaned endpoints changed: %#v", cleaned)
	}
}

func TestConstrainPolygonsWithinRepairsSmallChildBreach(t *testing.T) {
	parent := [][]geodata.GeoPoint{{
		{Longitude: 0, Latitude: 0},
		{Longitude: 2, Latitude: 0},
		{Longitude: 2, Latitude: 2},
		{Longitude: 0, Latitude: 2},
		{Longitude: 0, Latitude: 0},
	}}
	child := [][]geodata.GeoPoint{{
		{Longitude: 0.5, Latitude: 0.5},
		{Longitude: 1.5, Latitude: 0.5},
		{Longitude: 2.0005, Latitude: 0.0005},
		{Longitude: 1.5, Latitude: 1.5},
		{Longitude: 0.5, Latitude: 1.5},
		{Longitude: 0.5, Latitude: 0.5},
	}}
	repaired := ConstrainPolygonsWithin(parent, child)
	if geodata.SphericalPolygonsPathMissDistanceKM(parent, repaired, true) > 0 {
		t.Fatal("small child breach remains outside parent")
	}
	if len(repaired[0]) != len(child[0]) {
		t.Fatalf("repair changed ring point count from %d to %d", len(child[0]), len(repaired[0]))
	}
}

func TestConstrainPolygonsWithinLeavesLargeBreachUntouched(t *testing.T) {
	parent := [][]geodata.GeoPoint{{
		{Longitude: 0, Latitude: 0},
		{Longitude: 1, Latitude: 0},
		{Longitude: 1, Latitude: 1},
		{Longitude: 0, Latitude: 1},
		{Longitude: 0, Latitude: 0},
	}}
	child := [][]geodata.GeoPoint{{
		{Longitude: 0.2, Latitude: 0.2},
		{Longitude: 4, Latitude: 0.2},
		{Longitude: 0.2, Latitude: 0.8},
		{Longitude: 0.2, Latitude: 0.2},
	}}
	repaired := ConstrainPolygonsWithin(parent, child)
	if miss := geodata.SphericalPolygonsPathMissDistanceKM(parent, repaired, true); miss <= 0 {
		t.Fatal("large child breach was unexpectedly constrained")
	}
}

func TestVisibleBandPolygons20250729MarsTotalRiseSetRemovesPolarBacktrack(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	paths, err := basic.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), basic.OccultationMars,
		basic.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
			DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	if len(path.TotalRiseSetCurves) == 0 {
		t.Fatal("Mars path is missing inner-contact total rise/set curves")
	}
	polygons, authoritative, err := VisibleTotalBandPolygons(
		path.TotalBandFootprints, path.NorthernTotalLimit, path.SouthernTotalLimit, path.TotalRiseSetCurves,
	)
	if err != nil {
		t.Fatalf("VisibleBandPolygons: %v", err)
	}
	if !authoritative {
		t.Fatal("total band fell back to footprint-sweep geometry instead of inner-contact rise/set boundaries")
	}
	if len(polygons) != 1 {
		t.Fatalf("total polygon count=%d, want one continuous band", len(polygons))
	}
	for polygonIndex, polygon := range polygons {
		maximumEdge := 0.0
		for startIndex := 0; startIndex < len(polygon); startIndex++ {
			arcLength := 0.0
			limit := startIndex + 16
			if limit >= len(polygon) {
				limit = len(polygon) - 1
			}
			for endIndex := startIndex + 1; endIndex <= limit; endIndex++ {
				edge := geodata.EquatorialDistanceKM(polygon[endIndex-1], polygon[endIndex])
				arcLength += edge
				if edge > maximumEdge {
					maximumEdge = edge
				}
				if endIndex < startIndex+3 {
					continue
				}
				closure := geodata.EquatorialDistanceKM(polygon[startIndex], polygon[endIndex])
				if closure <= 180 && arcLength-closure >= 35 {
					t.Fatalf("total polygon %d retains a large polar backtrack: start=%d end=%d closure=%.1f km detour=%.1f km",
						polygonIndex, startIndex, endIndex, closure, arcLength-closure)
				}
			}
		}
		if maximumEdge > 175 {
			t.Fatalf("total polygon %d retains a %.1f km display chord", polygonIndex, maximumEdge)
		}
	}
}

func TestVisibleBandPolygons20250729MarsTotalSmoothingPreservesGeometry(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	paths, err := basic.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), basic.OccultationMars,
		basic.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
			DisableFootprints: true, IncludeFootprintTimeline: true,
			FootprintTimelineStep: 5 * time.Minute,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	rawTotal, authoritative, err := visibleBandPolygons(
		path.TotalBandFootprints, path.NorthernTotalLimit, path.SouthernTotalLimit,
		path.TotalBandContours, path.TotalRiseSetCurves, true,
	)
	if err != nil || !authoritative {
		t.Fatalf("raw total band authoritative=%v err=%v, want authoritative geometry", authoritative, err)
	}
	rawTotal = normalizeOccultationBandOutput(rawTotal)
	smoothedTotal := roundOccultationTotalBandJunctions(rawTotal)
	if shrink := geodata.SphericalPolygonsPathMissDistanceKM(smoothedTotal, rawTotal, true); shrink > 1 {
		t.Fatalf("wide-shoulder repair excludes %.1f km of the source total ring", shrink)
	}
	secondPass := roundOccultationTotalBandJunctions(smoothedTotal)
	if change := math.Max(
		geodata.SphericalPolygonsPathMissDistanceKM(smoothedTotal, secondPass, true),
		geodata.SphericalPolygonsPathMissDistanceKM(secondPass, smoothedTotal, true),
	); change > 0.05 {
		t.Fatalf("total-band smoothing is not idempotent: second pass changed %.3f km", change)
	}
	for polygonIndex, polygon := range smoothedTotal {
		for pointIndex := 2; pointIndex+2 < len(polygon); pointIndex++ {
			if deviation := occultationTestProjectedPointLineDistanceKM(
				polygon[pointIndex], polygon[pointIndex-2], polygon[pointIndex+2],
			); deviation > 35 {
				t.Fatalf("total polygon %d retains %.1f km local roughness at point %d",
					polygonIndex, deviation, pointIndex)
			}
		}
	}

	partial, partialAuthoritative, err := VisibleBandPolygonsFromContours(
		path.PartialBandFootprints, path.PartialBandContours,
		path.NorthernLimit, path.SouthernLimit, path.RiseSetCurves,
	)
	if err != nil || !partialAuthoritative {
		t.Fatalf("partial band authoritative=%v err=%v, want authoritative geometry", partialAuthoritative, err)
	}
	rawPartial, _, err := visibleBandPolygons(
		path.PartialBandFootprints, path.NorthernLimit, path.SouthernLimit,
		path.PartialBandContours, path.RiseSetCurves, false,
	)
	if err != nil {
		t.Fatalf("raw partial band: %v", err)
	}
	rawPartial = normalizeOccultationBandOutput(rawPartial)
	partialChangeKM := math.Max(
		geodata.SphericalPolygonsPathMissDistanceKM(rawPartial, partial, true),
		geodata.SphericalPolygonsPathMissDistanceKM(partial, rawPartial, true),
	)
	if partialChangeKM > 0.05 {
		t.Fatalf("total-only wide-shoulder path changed partial geometry by %.3f km", partialChangeKM)
	}
}

func TestVisibleBandPolygons20250729MarsTotalRiseSetStaysInsidePartialBand(t *testing.T) {
	zone := time.FixedZone("UTC+8", 8*60*60)
	start := time.Date(2025, time.July, 29, 0, 0, 0, 0, zone)
	paths, err := basic.FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), basic.OccultationMars,
		basic.OccultationPathOptions{
			Step: 20 * time.Minute, TargetSpacingKM: 900, RiseSetStep: time.Minute,
			DisableFootprints: true,
		},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	path := paths[0]
	partial, authoritative, err := VisibleBandPolygons(
		path.PartialBandFootprints, path.NorthernLimit, path.SouthernLimit, path.RiseSetCurves,
	)
	if err != nil {
		t.Fatalf("partial VisibleBandPolygons: %v", err)
	}
	if !authoritative {
		t.Fatal("partial visible band fell back to the footprint sweep")
	}
	total, authoritative, err := VisibleTotalBandPolygons(
		path.TotalBandFootprints, path.NorthernTotalLimit, path.SouthernTotalLimit, path.TotalRiseSetCurves,
	)
	if err != nil {
		t.Fatalf("total VisibleBandPolygons: %v", err)
	}
	if !authoritative {
		t.Fatal("total visible band fell back to the footprint sweep")
	}
	greatest := [][]geodata.GeoPoint{{
		{Longitude: path.Greatest.Longitude, Latitude: path.Greatest.Latitude},
	}}
	if !geodata.SphericalPolygonsContainPaths(total, greatest, false) {
		miss := geodata.SphericalPolygonsPathMissDistanceKM(total, greatest, false)
		t.Fatalf("total band misses greatest point by %.1f km", miss)
	}
	if miss := geodata.SphericalPolygonsPathMissDistanceKM(partial, total, true); miss > 10 {
		t.Fatalf("total band extends %.1f km outside the partial band", miss)
	}
}

func TestOccultationVisibleBandLineworkInputsPreferContoursWhenAvailable(t *testing.T) {
	visibleFill := [][]geodata.GeoPoint{{
		{Longitude: -20, Latitude: 81.5},
		{Longitude: -19, Latitude: 81.5},
		{Longitude: -19, Latitude: 82.0},
		{Longitude: -20, Latitude: 82.0},
	}}
	contourFill := [][]geodata.GeoPoint{{
		{Longitude: -20.5, Latitude: 81.6},
		{Longitude: -19.5, Latitude: 81.6},
		{Longitude: -19.5, Latitude: 81.9},
		{Longitude: -20.5, Latitude: 81.9},
	}}
	coveragePaths := [][]geodata.GeoPoint{{
		{Longitude: -20.25, Latitude: 81.75},
	}}

	for _, strongPolarSmoothing := range []bool{false, true} {
		label := "partial"
		if strongPolarSmoothing {
			label = "total"
		}
		selectionFill, selectedCoverage, curveCoveragePaths, visibleFillCoveragePaths :=
			occultationVisibleBandLineworkInputs(
				true, strongPolarSmoothing, visibleFill, contourFill, coveragePaths, nil,
			)
		expectedSelection := append(append([][]geodata.GeoPoint(nil), contourFill...), visibleFill...)
		if !geoPointCollectionsEqual(selectionFill, expectedSelection) {
			t.Fatalf("%s selection fill = %#v, want contour plus footprint witnesses %#v", label, selectionFill, expectedSelection)
		}
		expectedCoverage := coveragePaths
		if !geoPointCollectionsEqual(selectedCoverage, expectedCoverage) {
			t.Fatalf("%s coverage paths = %#v, want original footprint probes %#v", label, selectedCoverage, expectedCoverage)
		}
		if len(curveCoveragePaths) != 0 {
			t.Fatalf("%s curve coverage count=%d, want zero with no curves", label, len(curveCoveragePaths))
		}
		if !geoPointCollectionsEqual(visibleFillCoveragePaths, coveragePaths) {
			t.Fatalf("%s visible-fill coverage = %#v, want original footprint probes %#v",
				label, visibleFillCoveragePaths, coveragePaths)
		}
	}
}

func geoPointCollectionsEqual(first, second [][]geodata.GeoPoint) bool {
	if len(first) != len(second) {
		return false
	}
	for index := range first {
		if len(first[index]) != len(second[index]) {
			return false
		}
		for pointIndex := range first[index] {
			if first[index][pointIndex] != second[index][pointIndex] {
				return false
			}
		}
	}
	return true
}
