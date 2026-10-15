package geodata

import (
	"fmt"
	"math"
	"sort"
)

type visibleLineworkNode struct {
	point    GeoPoint
	outgoing []visibleLineworkHalfEdge
}

type visibleLineworkEdge struct {
	points     []GeoPoint
	start, end int
}

type visibleLineworkHalfEdge struct {
	edge    int
	reverse bool
}

type visibleLineworkBounds struct {
	minLongitude float64
	maxLongitude float64
	minLatitude  float64
	maxLatitude  float64
}

// visibleLineworkArcIndex caches the endpoint vectors and great-circle
// normals used by the source-boundary audit. The audit may visit thousands of
// polygon vertices, but its source segments are fixed for the whole pass.
// Caching only these derived values preserves the exact spherical-distance
// calculation while avoiding repeated trigonometric conversions in WASM.
type visibleLineworkArcIndex struct {
	arcs     []visibleLineworkArc
	vertices map[visibleLineworkPointKey]GeoPoint
	edges    map[visibleLineworkEdgeKey][]int
}

type visibleLineworkArc struct {
	start, end  GeoPoint
	startVector geoVector3
	endVector   geoVector3
	arcRadians  float64
	normal      geoVector3
	normalValid bool
}

type visibleLineworkPointKey struct {
	longitude int64
	latitude  int64
}

type visibleLineworkEdgeKey struct {
	first  visibleLineworkPointKey
	second visibleLineworkPointKey
}

const visibleLineworkIndexCoordinateScale = 1e7

// Coarse historical solar footprints can leave a bounded spherical residual;
// larger residuals indicate a planar-union shortcut rather than sampling noise.
const visibleLineworkBoundaryMinimumAuditToleranceKM = 100

func newVisibleLineworkArcIndex(lines [][]GeoPoint) visibleLineworkArcIndex {
	count := 0
	for _, line := range lines {
		if len(line) > 1 {
			count += len(line) - 1
		}
	}
	index := visibleLineworkArcIndex{
		arcs:     make([]visibleLineworkArc, 0, count),
		vertices: make(map[visibleLineworkPointKey]GeoPoint, count+1),
		edges:    make(map[visibleLineworkEdgeKey][]int, count),
	}
	for _, line := range lines {
		for segment := 1; segment < len(line); segment++ {
			start, end := line[segment-1], line[segment]
			startVector := geoPointVector(start)
			endVector := geoPointVector(end)
			arcRadians := math.Acos(math.Max(-1, math.Min(1, geoVectorDot(startVector, endVector))))
			normal, normalValid := geoVectorNormalize(geoVectorCross(startVector, endVector))
			arcIndex := len(index.arcs)
			index.arcs = append(index.arcs, visibleLineworkArc{
				start: start, end: end,
				startVector: startVector, endVector: endVector,
				arcRadians: arcRadians,
				normal:     normal, normalValid: normalValid,
			})
			startKey := visibleLineworkIndexPointKey(start)
			endKey := visibleLineworkIndexPointKey(end)
			index.vertices[startKey] = start
			index.vertices[endKey] = end
			edgeKey := visibleLineworkIndexEdgeKey(startKey, endKey)
			index.edges[edgeKey] = append(index.edges[edgeKey], arcIndex)
		}
	}
	return index
}

func (index visibleLineworkArcIndex) pointDistanceKM(point GeoPoint) float64 {
	if source, ok := index.vertices[visibleLineworkIndexPointKey(point)]; ok {
		return EquatorialDistanceKM(point, source)
	}
	pointVector := geoPointVector(point)
	minimum := math.Inf(1)
	for _, arc := range index.arcs {
		minimum = math.Min(minimum, arc.pointDistanceKM(point, pointVector))
		if minimum == 0 {
			return 0
		}
	}
	return minimum
}

func (index visibleLineworkArcIndex) edgePointDistanceKM(
	first, second, point GeoPoint,
) (float64, bool) {
	firstKey := visibleLineworkIndexPointKey(first)
	secondKey := visibleLineworkIndexPointKey(second)
	arcIndices := index.edges[visibleLineworkIndexEdgeKey(firstKey, secondKey)]
	if len(arcIndices) == 0 {
		return 0, false
	}
	pointVector := geoPointVector(point)
	minimum := math.Inf(1)
	matched := false
	for _, arcIndex := range arcIndices {
		arc := index.arcs[arcIndex]
		direct := EquatorialDistanceKM(first, arc.start) <= 0.001 &&
			EquatorialDistanceKM(second, arc.end) <= 0.001
		reverse := EquatorialDistanceKM(first, arc.end) <= 0.001 &&
			EquatorialDistanceKM(second, arc.start) <= 0.001
		if !direct && !reverse {
			continue
		}
		matched = true
		minimum = math.Min(minimum, arc.pointDistanceKM(point, pointVector))
	}
	return minimum, matched
}

func (arc visibleLineworkArc) pointDistanceKM(point GeoPoint, pointVector geoVector3) float64 {
	if arc.normalValid {
		projection := geoVectorAdd(pointVector,
			geoVectorScale(arc.normal, -geoVectorDot(pointVector, arc.normal)))
		if projected, projectedOK := geoVectorNormalize(projection); projectedOK {
			for _, candidate := range []geoVector3{projected, geoVectorScale(projected, -1)} {
				if sphericalPointOnArcVectors(candidate, arc.startVector, arc.endVector, arc.arcRadians) {
					angle := math.Acos(math.Max(-1, math.Min(1, geoVectorDot(pointVector, candidate))))
					return angle * EarthEquatorialRadiusKM
				}
			}
		}
	}
	return math.Min(EquatorialDistanceKM(point, arc.start), EquatorialDistanceKM(point, arc.end))
}

func visibleLineworkIndexPointKey(point GeoPoint) visibleLineworkPointKey {
	return visibleLineworkPointKey{
		longitude: int64(math.Round(point.Longitude * visibleLineworkIndexCoordinateScale)),
		latitude:  int64(math.Round(point.Latitude * visibleLineworkIndexCoordinateScale)),
	}
}

func visibleLineworkIndexEdgeKey(
	first, second visibleLineworkPointKey,
) visibleLineworkEdgeKey {
	if first.longitude > second.longitude ||
		(first.longitude == second.longitude && first.latitude > second.latitude) {
		first, second = second, first
	}
	return visibleLineworkEdgeKey{first: first, second: second}
}

func sphericalPointOnArcVectors(point, first, second geoVector3, arc float64) bool {
	if arc <= 1e-14 {
		return math.Acos(math.Max(-1, math.Min(1, geoVectorDot(first, point)))) <= 1e-9
	}
	firstDistance := math.Acos(math.Max(-1, math.Min(1, geoVectorDot(first, point))))
	secondDistance := math.Acos(math.Max(-1, math.Min(1, geoVectorDot(point, second))))
	return math.Abs(firstDistance+secondDistance-arc) <= 1e-9
}

// VisibleLineworkPolygons 将球面边界线网多边形化，并仅保留由给定填充多边形覆盖的面。
// VisibleLineworkPolygons polygonizes a spherical boundary network and keeps
// only faces covered by at least one supplied fill polygon. Boundary line
// endpoints within snapDistanceKM are treated as the same physical junction.
func VisibleLineworkPolygons(
	boundaryLines [][]GeoPoint,
	fillPolygons [][]GeoPoint,
	coveragePaths [][]GeoPoint,
	snapDistanceKM float64,
) ([][]GeoPoint, error) {
	return visibleLineworkPolygons(
		boundaryLines, fillPolygons, coveragePaths, snapDistanceKM, 1.5*snapDistanceKM,
	)
}

// VisibleLineworkPolygonsWithAuditTolerance 将边界图节点吸附尺度与源填充审计容差分开。
// VisibleLineworkPolygonsWithAuditTolerance keeps graph junction snapping and
// source-fill auditing on independent scales. This is needed when a boundary
// is sampled at a fine spatial cadence but the interior witness is a coarser
// sweep: increasing the snap distance would merge nearby physical branches,
// while a bounded audit tolerance does not alter the boundary graph.
func VisibleLineworkPolygonsWithAuditTolerance(
	boundaryLines [][]GeoPoint,
	fillPolygons [][]GeoPoint,
	coveragePaths [][]GeoPoint,
	snapDistanceKM, sourceFillToleranceKM float64,
) ([][]GeoPoint, error) {
	return visibleLineworkPolygons(
		boundaryLines, fillPolygons, coveragePaths, snapDistanceKM, sourceFillToleranceKM,
	)
}

func visibleLineworkPolygons(
	boundaryLines [][]GeoPoint,
	fillPolygons [][]GeoPoint,
	coveragePaths [][]GeoPoint,
	snapDistanceKM, sourceFillToleranceKM float64,
) ([][]GeoPoint, error) {
	if len(boundaryLines) == 0 {
		return nil, fmt.Errorf("visible linework has no boundary lines")
	}
	if !(snapDistanceKM >= 0) || !(sourceFillToleranceKM >= 0) ||
		math.IsInf(snapDistanceKM, 0) || math.IsInf(sourceFillToleranceKM, 0) {
		return nil, fmt.Errorf("visible linework tolerances must be finite and non-negative")
	}
	chartInputs := make([][]GeoPoint, 0, len(boundaryLines)+len(fillPolygons))
	chartInputs = append(chartInputs, boundaryLines...)
	chartInputs = append(chartInputs, fillPolygons...)
	chart, ok := newPolygonUnionChart(chartInputs)
	if !ok {
		return nil, fmt.Errorf("visible linework has no stable spherical chart")
	}

	projectedLines := make([][]GeoPoint, 0, len(boundaryLines))
	for _, source := range boundaryLines {
		line := projectVisibleLine(chart, source)
		if len(line) >= 2 {
			projectedLines = append(projectedLines, line)
		}
	}
	projectedFill := make([][]GeoPoint, 0, len(fillPolygons))
	for _, source := range fillPolygons {
		polygon := projectVisibleLine(chart, openGeoRing(source))
		if len(polygon) >= 3 {
			projectedFill = append(projectedFill, polygon)
		}
	}
	if len(projectedLines) == 0 || len(projectedFill) == 0 {
		return nil, fmt.Errorf("visible linework has no usable boundary or fill geometry")
	}
	projectedLines = splitVisibleLineworkIntersections(projectedLines)
	if len(coveragePaths) == 0 {
		coveragePaths = fillPolygons
	}
	projectedCoverage := make([][]GeoPoint, 0, len(coveragePaths))
	for _, source := range coveragePaths {
		path := projectVisibleLine(chart, source)
		if len(path) > 0 {
			projectedCoverage = append(projectedCoverage, path)
		}
	}
	for _, probe := range visibleLineworkFilledEdgeProbes(projectedLines, projectedFill) {
		projectedCoverage = append(projectedCoverage, []GeoPoint{probe})
	}

	nodes, edges, loops := buildVisibleLineworkGraph(projectedLines, chart, snapDistanceKM)
	cycles := enumerateVisibleLineworkCycles(nodes, edges)
	cycles = append(cycles, loops...)
	// Endpoint snapping and face membership have different numerical scales.
	// A tens-of-kilometres junction snap is useful for joining independently sampled phase
	// curves, but using that same radius for point-in-face tests makes adjacent
	// polar faces overlap and selects the wrong side of a narrow band.
	faceToleranceKM := math.Min(5, math.Max(0.5, snapDistanceKM/10))
	projectedTolerance := faceToleranceKM / EarthKMPerDegree
	cycleProbes := visibleLineworkFilledCycleProbes(cycles, projectedFill, projectedTolerance)
	if len(cycleProbes) > 0 {
		for _, probe := range cycleProbes {
			projectedCoverage = append(projectedCoverage, []GeoPoint{probe})
		}
	}
	// Audit the selected boundary against the complete source fill, not just
	// the sparse coverage probes used to choose faces. Otherwise a face can
	// contain the probes while its outline still cuts across an unsampled part
	// of the instantaneous footprint.
	projectedFillSamples := sampleVisibleLineworkFillPaths(projectedFill, 16)
	fillSamples := visibleLineworkPointPaths(
		unprojectVisibleLineworkPaths(chart, projectedFillSamples),
	)
	selectedCycles := make([][]GeoPoint, 0)
	if len(cycleProbes) > 0 {
		priorityCoverage := make([][]GeoPoint, len(cycleProbes))
		for index, probe := range cycleProbes {
			priorityCoverage[index] = []GeoPoint{probe}
		}
		prioritySelected := selectVisibleLineworkCycles(
			cycles, priorityCoverage, projectedTolerance,
		)
		if visibleLineworkCyclesCoverCoverage(prioritySelected, projectedCoverage, projectedTolerance) {
			selectedCycles = prioritySelected
		}
	}
	if len(selectedCycles) == 0 {
		selectedCycles = selectVisibleLineworkCycles(
			cycles, projectedCoverage, projectedTolerance,
		)
	}
	selected := make([][]GeoPoint, 0, len(selectedCycles))
	for _, cycle := range selectedCycles {
		polygon := make([]GeoPoint, len(cycle))
		for index, point := range cycle {
			polygon[index] = chart.unproject(point)
		}
		selected = append(selected, polygon)
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("visible linework produced no covering boundary from %d cycles (coverage=%d cycle_probes=%d misses=%d)",
			len(cycles), len(projectedCoverage), len(cycleProbes),
			visibleLineworkCoverageMissCount(cycles, projectedCoverage, projectedTolerance))
	}
	// A single simple face is already a polygon. Repeating a boolean union
	// can lose tiny edges where tangential branches were split into the graph.
	merged := selected
	if len(selected) > 1 {
		var err error
		merged, err = UnionPolygons(selected)
		if err != nil {
			return nil, fmt.Errorf("visible linework face union: %w", err)
		}
	}
	// The common output already consists of short projected edges. Running the
	// exact spherical point-to-arc audit on every edge is disproportionately
	// expensive in TinyGo/WASM, while an artificial polar chord is necessarily
	// visible as a long projected edge. Keep the precise repair/audit only for
	// that suspicious case.
	suspiciousBoundary := visibleLineworkHasLongProjectedEdge(merged, 45) || visibleLineworkHasProjectedPolarBacktrack(merged)
	if suspiciousBoundary {
		merged = repairVisibleLineworkBoundaryChords(merged, boundaryLines, 5, 30)
		// Re-run the exact source audit only if a suspicious edge remains after
		// repair. Normal authoritative rings have no long edge or polar reversal
		// at this point, so a full point-to-arc scan would be redundant.
		if visibleLineworkHasLongProjectedEdge(merged, 45) || visibleLineworkHasProjectedPolarBacktrack(merged) {
			if sourceMiss := visibleLineworkBoundarySourceMissDistanceKM(merged, boundaryLines); sourceMiss > math.Max(visibleLineworkBoundaryMinimumAuditToleranceKM, 2*snapDistanceKM) {
				// A planar union can introduce a shortcut while each selected
				// source cycle is still a valid physical boundary. Preserve those
				// source cycles as separate polygons rather than falling through
				// to an endpoint-only sweep that closes through a pole.
				if selectedMiss := visibleLineworkBoundarySourceMissDistanceKM(selected, boundaryLines); selectedMiss <= math.Max(visibleLineworkBoundaryMinimumAuditToleranceKM, 2*snapDistanceKM) {
					return selected, nil
				}
				return nil, fmt.Errorf("visible linework boundary leaves source linework by %.1f km", sourceMiss)
			}
		}
	}
	if miss := SphericalPolygonsPathMissDistanceKM(merged, fillSamples, true); miss > sourceFillToleranceKM {
		return nil, fmt.Errorf("visible linework boundary misses source fill by %.1f km", miss)
	}
	return merged, nil
}

func visibleLineworkHasProjectedPolarBacktrack(polygons [][]GeoPoint) bool {
	for _, polygon := range polygons {
		for index := 1; index+1 < len(polygon); index++ {
			first, middle, last := polygon[index-1], polygon[index], polygon[index+1]
			if math.Abs(middle.Latitude) < 70 ||
				visibleLineworkProjectedEdgeDistanceKM(first, middle) > 40 ||
				visibleLineworkProjectedEdgeDistanceKM(middle, last) > 40 ||
				visibleLineworkProjectedEdgeDistanceKM(first, last) > 55 {
				continue
			}
			longitudeReversal := math.Remainder(middle.Longitude-first.Longitude, 360)*
				math.Remainder(last.Longitude-middle.Longitude, 360) < 0
			latitudeReversal := (middle.Latitude-first.Latitude)*(last.Latitude-middle.Latitude) < 0
			if longitudeReversal || latitudeReversal {
				return true
			}
		}
	}
	return false
}

func visibleLineworkHasLongProjectedEdge(polygons [][]GeoPoint, maximumKM float64) bool {
	if maximumKM <= 0 {
		return true
	}
	for _, polygon := range polygons {
		for index := 1; index < len(polygon); index++ {
			if visibleLineworkProjectedEdgeDistanceKM(polygon[index-1], polygon[index]) > maximumKM {
				return true
			}
		}
	}
	return false
}

func visibleLineworkProjectedEdgeDistanceKM(first, second GeoPoint) float64 {
	latitude := (first.Latitude + second.Latitude) * math.Pi / 360
	deltaLongitude := math.Remainder(second.Longitude-first.Longitude, 360) * math.Cos(latitude)
	deltaLatitude := second.Latitude - first.Latitude
	return math.Hypot(deltaLongitude, deltaLatitude) * EarthKMPerDegree
}

func visibleLineworkBoundarySourceMissDistanceKM(
	polygons, boundaryLines [][]GeoPoint,
) float64 {
	sourceArcs := newVisibleLineworkArcIndex(boundaryLines)
	maximumDistance := 0.0
	for _, polygon := range polygons {
		for index, point := range polygon {
			// Endpoints of a planar shortcut often coincide with real source
			// vertices. They must not mask a midpoint that leaves the source
			// boundary by hundreds or thousands of kilometres.
			edgeMaximum := sourceArcs.pointDistanceKM(point)
			if index > 0 {
				midpoint := sphericalInterpolate(polygon[index-1], point, 0.5)
				midpointDistance, matched := sourceArcs.edgePointDistanceKM(
					polygon[index-1], point, midpoint,
				)
				if !matched {
					midpointDistance = sourceArcs.pointDistanceKM(midpoint)
				}
				edgeMaximum = math.Max(edgeMaximum, midpointDistance)
			}
			maximumDistance = math.Max(maximumDistance, edgeMaximum)
		}
	}
	return maximumDistance
}

// repairVisibleLineworkBoundaryChords restores a source-boundary run when a
// planar union closes a polar face with a shortcut between two vertices on the
// same physical curve. The shortcut is topologically valid in the chart but
// is not part of the supplied linework and renders as an artificial straight
// chord on the map.
func repairVisibleLineworkBoundaryChords(
	polygons, boundaryLines [][]GeoPoint,
	maxDeviationKM, endpointToleranceKM float64,
) [][]GeoPoint {
	if len(polygons) == 0 || len(boundaryLines) == 0 || maxDeviationKM <= 0 {
		return polygons
	}
	sourceArcs := newVisibleLineworkArcIndex(boundaryLines)
	result := make([][]GeoPoint, len(polygons))
	for polygonIndex, ring := range polygons {
		if len(ring) < 4 {
			result[polygonIndex] = ring
			continue
		}
		repaired := make([]GeoPoint, 0, len(ring))
		repaired = append(repaired, ring[0])
		for index := 1; index < len(ring); index++ {
			start, end := ring[index-1], ring[index]
			midpoint := sphericalInterpolate(start, end, 0.5)
			deviation, matched := sourceArcs.edgePointDistanceKM(start, end, midpoint)
			if !matched {
				deviation = sourceArcs.pointDistanceKM(midpoint)
			}
			if deviation <= maxDeviationKM {
				repaired = append(repaired, end)
				continue
			}
			bridge, ok := visibleLineworkBoundaryBridge(
				start, end, boundaryLines, endpointToleranceKM,
			)
			if !ok {
				repaired = append(repaired, end)
				continue
			}
			repaired = append(repaired, bridge[1:]...)
		}
		result[polygonIndex] = sweepDeduplicateAdjacent(repaired)
	}
	return result
}

func visibleLineworkPointToLinesDistanceKM(
	point GeoPoint,
	lines [][]GeoPoint,
) float64 {
	return newVisibleLineworkArcIndex(lines).pointDistanceKM(point)
}

func visibleLineworkBoundaryBridge(
	start, end GeoPoint,
	lines [][]GeoPoint,
	toleranceKM float64,
) ([]GeoPoint, bool) {
	originalStart := start
	var best []GeoPoint
	bestLength := math.Inf(1)
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		startIndex, startDistance := visibleLineworkNearestVertex(start, line)
		endIndex, endDistance := visibleLineworkNearestVertex(end, line)
		if startDistance > toleranceKM || endDistance > toleranceKM ||
			startIndex == endIndex || absInt(startIndex-endIndex) < 2 {
			continue
		}
		if startIndex > endIndex {
			startIndex, endIndex = endIndex, startIndex
			start, end = end, start
		}
		bridge := make([]GeoPoint, 0, endIndex-startIndex+1)
		bridge = append(bridge, start)
		bridge = append(bridge, line[startIndex+1:endIndex]...)
		bridge = append(bridge, end)
		if len(bridge) >= 3 {
			if !visibleLineworkBridgeIsSimple(bridge, toleranceKM) {
				continue
			}
			if EquatorialDistanceKM(originalStart, bridge[0]) > EquatorialDistanceKM(originalStart, bridge[len(bridge)-1]) {
				reverseGeoPoints(bridge)
			}
			length := visibleLineworkBridgeLength(bridge)
			if length < bestLength {
				best, bestLength = bridge, length
			}
		}
	}
	if len(best) >= 3 {
		return best, true
	}
	return visibleLineworkCrossLineBoundaryBridge(start, end, lines, toleranceKM)
}

// visibleLineworkCrossLineBoundaryBridge restores a source-boundary run whose
// endpoints lie on different line strings. Polar phase envelopes commonly
// meet a horizon connector at an endpoint; the planar graph can collapse that
// junction and leave a long shortcut between the two strings. Stitch only
// existing source vertices that are within the same endpoint tolerance, and
// choose the shortest resulting source-path so unrelated lines cannot become a
// new artificial edge.
func visibleLineworkCrossLineBoundaryBridge(
	start, end GeoPoint,
	lines [][]GeoPoint,
	toleranceKM float64,
) ([]GeoPoint, bool) {
	type location struct {
		lineIndex, pointIndex int
		distance              float64
	}
	startLocations := make([]location, 0)
	endLocations := make([]location, 0)
	for lineIndex, line := range lines {
		if len(line) < 2 {
			continue
		}
		for pointIndex, point := range line {
			if distance := EquatorialDistanceKM(start, point); distance <= toleranceKM {
				startLocations = append(startLocations, location{lineIndex, pointIndex, distance})
			}
			if distance := EquatorialDistanceKM(end, point); distance <= toleranceKM {
				endLocations = append(endLocations, location{lineIndex, pointIndex, distance})
			}
		}
	}
	var best []GeoPoint
	bestLength := math.Inf(1)
	for _, first := range startLocations {
		for _, second := range endLocations {
			if first.lineIndex == second.lineIndex {
				continue
			}
			firstLine, secondLine := lines[first.lineIndex], lines[second.lineIndex]
			firstIndices := []int{0, len(firstLine) - 1}
			secondIndices := []int{0, len(secondLine) - 1}
			for _, firstIndex := range firstIndices {
				for _, secondIndex := range secondIndices {
					if EquatorialDistanceKM(firstLine[firstIndex], secondLine[secondIndex]) > toleranceKM {
						continue
					}
					firstPath := visibleLineworkLinePath(firstLine, first.pointIndex, firstIndex, start, firstLine[firstIndex])
					secondPath := visibleLineworkLinePath(secondLine, secondIndex, second.pointIndex, secondLine[secondIndex], end)
					if len(firstPath) < 2 || len(secondPath) < 2 {
						continue
					}
					bridge := append(append([]GeoPoint(nil), firstPath...), secondPath[1:]...)
					if len(bridge) < 3 {
						continue
					}
					if !visibleLineworkBridgeIsSimple(bridge, toleranceKM) {
						continue
					}
					length := visibleLineworkBridgeLength(bridge)
					if length < bestLength {
						best, bestLength = bridge, length
					}
				}
			}
		}
	}
	if len(best) < 3 {
		return nil, false
	}
	return best, true
}

func visibleLineworkBridgeIsSimple(points []GeoPoint, toleranceKM float64) bool {
	if len(points) < 3 {
		return false
	}
	endpointRepeatTolerance := math.Min(1, toleranceKM/10)
	for index := 1; index+1 < len(points); index++ {
		if EquatorialDistanceKM(points[index], points[0]) <= endpointRepeatTolerance ||
			EquatorialDistanceKM(points[index], points[len(points)-1]) <= endpointRepeatTolerance {
			return false
		}
	}
	return true
}

func visibleLineworkLinePath(
	line []GeoPoint,
	startIndex, endIndex int,
	start, end GeoPoint,
) []GeoPoint {
	if startIndex < 0 || startIndex >= len(line) || endIndex < 0 || endIndex >= len(line) {
		return nil
	}
	path := []GeoPoint{start}
	if startIndex <= endIndex {
		path = append(path, line[startIndex+1:endIndex+1]...)
	} else {
		for index := startIndex - 1; index >= endIndex; index-- {
			path = append(path, line[index])
		}
	}
	if len(path) == 0 || EquatorialDistanceKM(path[len(path)-1], end) > 1e-9 {
		path = append(path, end)
	}
	return path
}

func visibleLineworkBridgeLength(points []GeoPoint) float64 {
	length := 0.0
	for index := 1; index < len(points); index++ {
		length += EquatorialDistanceKM(points[index-1], points[index])
	}
	return length
}

func visibleLineworkNearestVertex(
	point GeoPoint,
	line []GeoPoint,
) (int, float64) {
	bestIndex := -1
	bestDistance := math.Inf(1)
	for index, candidate := range line {
		distance := EquatorialDistanceKM(point, candidate)
		if distance < bestDistance {
			bestIndex = index
			bestDistance = distance
		}
	}
	return bestIndex, bestDistance
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func visibleLineworkFilledEdgeProbes(lines, fillPolygons [][]GeoPoint) []GeoPoint {
	fillBounds := make([]visibleLineworkBounds, len(fillPolygons))
	for index, polygon := range fillPolygons {
		fillBounds[index] = visibleLineworkPolygonBounds(polygon, 0)
	}
	probes := make([]GeoPoint, 0, len(lines))
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		segment := (len(line) - 1) / 2
		start, end := line[segment], line[segment+1]
		dx := end.Longitude - start.Longitude
		dy := end.Latitude - start.Latitude
		length := math.Hypot(dx, dy)
		if length <= 1e-12 {
			continue
		}
		midpoint := GeoPoint{
			Longitude: (start.Longitude + end.Longitude) / 2,
			Latitude:  (start.Latitude + end.Latitude) / 2,
		}
		offset := math.Min(0.0025, 0.02*length)
		for _, side := range []float64{-1, 1} {
			probe := GeoPoint{
				Longitude: midpoint.Longitude - side*dy*offset/length,
				Latitude:  midpoint.Latitude + side*dx*offset/length,
			}
			for fillIndex, polygon := range fillPolygons {
				if fillBounds[fillIndex].contains(polygon, probe) &&
					visibleLineworkContainsWithin(polygon, probe, 0) {
					probes = append(probes, probe)
					break
				}
			}
		}
	}
	return probes
}

func visibleLineworkFilledCycleProbes(
	cycles, fillPolygons [][]GeoPoint,
	tolerance float64,
) []GeoPoint {
	type candidate struct {
		polygon []GeoPoint
		probe   GeoPoint
		area    float64
		bounds  visibleLineworkBounds
	}
	fillBounds := make([]visibleLineworkBounds, len(fillPolygons))
	for index, polygon := range fillPolygons {
		fillBounds[index] = visibleLineworkPolygonBounds(polygon, tolerance)
	}
	candidates := make([]candidate, 0, len(cycles))
	for _, cycle := range cycles {
		probe, ok := visibleLineworkPolygonCentroid(cycle)
		if !ok || !visibleLineworkContainsWithin(cycle, probe, tolerance) {
			continue
		}
		for fillIndex, polygon := range fillPolygons {
			if fillBounds[fillIndex].contains(polygon, probe) &&
				visibleLineworkContainsWithin(polygon, probe, tolerance) {
				candidates = append(candidates, candidate{
					polygon: cycle,
					probe:   probe,
					area:    math.Abs(visibleLineworkSignedArea(cycle)),
					bounds:  visibleLineworkPolygonBounds(cycle, tolerance),
				})
				break
			}
		}
	}
	sort.Slice(candidates, func(first, second int) bool {
		return candidates[first].area < candidates[second].area
	})
	selected := make([]candidate, 0, len(candidates))
	probes := make([]GeoPoint, 0, len(candidates))
	for _, value := range candidates {
		coveredBySmallerFace := false
		for _, smaller := range selected {
			if smaller.bounds.contains(smaller.polygon, value.probe) &&
				visibleLineworkContainsWithin(smaller.polygon, value.probe, tolerance) {
				coveredBySmallerFace = true
				break
			}
		}
		if coveredBySmallerFace {
			continue
		}
		selected = append(selected, value)
		probes = append(probes, value.probe)
	}
	return probes
}

func visibleLineworkPolygonCentroid(polygon []GeoPoint) (GeoPoint, bool) {
	if len(polygon) < 3 {
		return GeoPoint{}, false
	}
	area, longitude, latitude := 0.0, 0.0, 0.0
	for current, previous := 0, len(polygon)-1; current < len(polygon); previous, current = current, current+1 {
		cross := polygon[previous].Longitude*polygon[current].Latitude -
			polygon[current].Longitude*polygon[previous].Latitude
		area += cross
		longitude += (polygon[previous].Longitude + polygon[current].Longitude) * cross
		latitude += (polygon[previous].Latitude + polygon[current].Latitude) * cross
	}
	if math.Abs(area) <= 1e-12 {
		return GeoPoint{}, false
	}
	return GeoPoint{
		Longitude: longitude / (3 * area),
		Latitude:  latitude / (3 * area),
	}, true
}

type visibleLineworkSegmentSplit struct {
	fraction float64
	point    GeoPoint
}

// splitVisibleLineworkIntersections turns interior crossings into graph
// nodes. Lunar rise/set curves can cross at a real polar horizon-grazing
// site, whereas solar-eclipse linework usually meets only at endpoints.
func splitVisibleLineworkIntersections(lines [][]GeoPoint) [][]GeoPoint {
	if len(lines) < 2 {
		return lines
	}
	splits := make([][][]visibleLineworkSegmentSplit, len(lines))
	junctions := make([][]bool, len(lines))
	for lineIndex, line := range lines {
		if len(line) < 2 {
			continue
		}
		splits[lineIndex] = make([][]visibleLineworkSegmentSplit, len(line)-1)
		junctions[lineIndex] = make([]bool, len(line))
	}
	for firstLine := 0; firstLine < len(lines); firstLine++ {
		for firstSegment := 0; firstSegment+1 < len(lines[firstLine]); firstSegment++ {
			firstStart := lines[firstLine][firstSegment]
			firstEnd := lines[firstLine][firstSegment+1]
			for secondLine := firstLine; secondLine < len(lines); secondLine++ {
				secondStartSegment := 0
				if secondLine == firstLine {
					secondStartSegment = firstSegment + 2
				}
				for secondSegment := secondStartSegment; secondSegment+1 < len(lines[secondLine]); secondSegment++ {
					secondStart := lines[secondLine][secondSegment]
					secondEnd := lines[secondLine][secondSegment+1]
					longitudeShift := 360 * math.Round(
						((firstStart.Longitude+firstEnd.Longitude)-(secondStart.Longitude+secondEnd.Longitude))/720,
					)
					shiftedStart := secondStart
					shiftedEnd := secondEnd
					shiftedStart.Longitude += longitudeShift
					shiftedEnd.Longitude += longitudeShift
					firstFraction, secondFraction, point, ok := visibleLineworkSegmentIntersection(
						firstStart, firstEnd, shiftedStart, shiftedEnd,
					)
					if !ok {
						continue
					}
					visibleLineworkRecordSplit(
						splits[firstLine], junctions[firstLine], firstSegment, firstFraction, point,
					)
					point.Longitude -= longitudeShift
					visibleLineworkRecordSplit(
						splits[secondLine], junctions[secondLine], secondSegment, secondFraction, point,
					)
				}
			}
		}
	}

	result := make([][]GeoPoint, 0, len(lines)*2)
	for lineIndex, line := range lines {
		if len(line) < 2 {
			continue
		}
		current := []GeoPoint{line[0]}
		for segmentIndex := 0; segmentIndex+1 < len(line); segmentIndex++ {
			if segmentIndex > 0 && junctions[lineIndex][segmentIndex] {
				if len(current) >= 2 {
					result = append(result, current)
				}
				current = []GeoPoint{line[segmentIndex]}
			}
			segmentSplits := splits[lineIndex][segmentIndex]
			sort.Slice(segmentSplits, func(first, second int) bool {
				return segmentSplits[first].fraction < segmentSplits[second].fraction
			})
			for _, split := range segmentSplits {
				if len(current) == 0 || !SameGeoPoint(current[len(current)-1], split.point) {
					current = append(current, split.point)
				}
				if len(current) >= 2 {
					result = append(result, current)
				}
				current = []GeoPoint{split.point}
			}
			if !SameGeoPoint(current[len(current)-1], line[segmentIndex+1]) {
				current = append(current, line[segmentIndex+1])
			}
		}
		if len(current) >= 2 {
			result = append(result, current)
		}
	}
	return result
}

func visibleLineworkRecordSplit(
	splits [][]visibleLineworkSegmentSplit,
	junctions []bool,
	segment int,
	fraction float64,
	point GeoPoint,
) {
	const endpointTolerance = 1e-8
	if fraction <= endpointTolerance {
		junctions[segment] = true
		return
	}
	if fraction >= 1-endpointTolerance {
		junctions[segment+1] = true
		return
	}
	for _, existing := range splits[segment] {
		if math.Abs(existing.fraction-fraction) <= endpointTolerance {
			return
		}
	}
	splits[segment] = append(splits[segment], visibleLineworkSegmentSplit{
		fraction: fraction,
		point:    point,
	})
}

func visibleLineworkSegmentIntersection(
	firstStart, firstEnd, secondStart, secondEnd GeoPoint,
) (float64, float64, GeoPoint, bool) {
	firstX := firstEnd.Longitude - firstStart.Longitude
	firstY := firstEnd.Latitude - firstStart.Latitude
	secondX := secondEnd.Longitude - secondStart.Longitude
	secondY := secondEnd.Latitude - secondStart.Latitude
	denominator := firstX*secondY - firstY*secondX
	if math.Abs(denominator) <= 1e-12 {
		return 0, 0, GeoPoint{}, false
	}
	offsetX := secondStart.Longitude - firstStart.Longitude
	offsetY := secondStart.Latitude - firstStart.Latitude
	firstFraction := (offsetX*secondY - offsetY*secondX) / denominator
	secondFraction := (offsetX*firstY - offsetY*firstX) / denominator
	const intersectionTolerance = 1e-9
	if firstFraction < -intersectionTolerance || firstFraction > 1+intersectionTolerance ||
		secondFraction < -intersectionTolerance || secondFraction > 1+intersectionTolerance {
		// 参数判定只容忍 1e-9，旋转图幅下 T 型交点会被拒绝，这里按垂距回退判定。
		if fraction, ok := visibleLineworkEndpointOnSegment(secondStart, firstStart, firstEnd); ok {
			return fraction, 0, secondStart, true
		}
		if fraction, ok := visibleLineworkEndpointOnSegment(secondEnd, firstStart, firstEnd); ok {
			return fraction, 1, secondEnd, true
		}
		if fraction, ok := visibleLineworkEndpointOnSegment(firstStart, secondStart, secondEnd); ok {
			return 0, fraction, firstStart, true
		}
		if fraction, ok := visibleLineworkEndpointOnSegment(firstEnd, secondStart, secondEnd); ok {
			return 1, fraction, firstEnd, true
		}
		return 0, 0, GeoPoint{}, false
	}
	firstFraction = math.Max(0, math.Min(1, firstFraction))
	secondFraction = math.Max(0, math.Min(1, secondFraction))
	return firstFraction, secondFraction, GeoPoint{
		Longitude: firstStart.Longitude + firstFraction*firstX,
		Latitude:  firstStart.Latitude + firstFraction*firstY,
	}, true
}

// visibleLineworkTJunctionToleranceDeg 是端点落在另一段内部的垂距容差（度）。
const visibleLineworkTJunctionToleranceDeg = 1e-5

// visibleLineworkEndpointOnSegment 返回端点落在段内部时的投影参数。
func visibleLineworkEndpointOnSegment(point, start, end GeoPoint) (float64, bool) {
	deltaX := end.Longitude - start.Longitude
	deltaY := end.Latitude - start.Latitude
	lengthSquared := deltaX*deltaX + deltaY*deltaY
	if !(lengthSquared > 0) {
		return 0, false
	}
	offsetX := point.Longitude - start.Longitude
	offsetY := point.Latitude - start.Latitude
	fraction := (offsetX*deltaX + offsetY*deltaY) / lengthSquared
	if !(fraction > 0 && fraction < 1) {
		return 0, false
	}
	distance := math.Hypot(offsetX-fraction*deltaX, offsetY-fraction*deltaY)
	if !(distance <= visibleLineworkTJunctionToleranceDeg) {
		return 0, false
	}
	return fraction, true
}

func selectVisibleLineworkCycles(
	cycles, coveragePaths [][]GeoPoint,
	boundaryTolerance float64,
) [][]GeoPoint {
	return selectVisibleLineworkCyclesWithProbeLimit(cycles, coveragePaths, boundaryTolerance, 768)
}

func selectVisibleLineworkCyclesWithProbeLimit(
	cycles, coveragePaths [][]GeoPoint,
	boundaryTolerance float64,
	maximumProbes int,
) [][]GeoPoint {
	type candidate struct {
		polygon []GeoPoint
		area    float64
		bounds  visibleLineworkBounds
		index   visibleLineworkPolygonIndex
		indexed bool
		covers  []bool
		count   int
	}
	probes := visibleLineworkCoverageProbes(coveragePaths, 8)
	if len(probes) == 0 {
		return nil
	}
	probes = limitVisibleLineworkCoverageProbes(probes, maximumProbes)
	candidates := make([]candidate, 0, len(cycles))
	for _, source := range cycles {
		polygon := append([]GeoPoint(nil), source...)
		area := visibleLineworkSignedArea(polygon)
		if math.Abs(area) <= 1e-10 {
			continue
		}
		if area < 0 {
			reverseSweepGeoPoints(polygon)
			area = -area
		}
		candidates = append(candidates, candidate{
			polygon: polygon,
			area:    area,
			bounds:  visibleLineworkPolygonBounds(polygon, boundaryTolerance),
		})
	}
	sort.SliceStable(candidates, func(first, second int) bool {
		return candidates[first].area < candidates[second].area
	})

	// Most physical line networks have one face containing every footprint
	// probe. A small subset cheaply rejects impossible faces; candidates that
	// survive still have to cover the original complete probe set.
	sentinels := sampleVisibleLineworkProbes(probes, 16)
	best := -1
	for candidateIndex := range candidates {
		value := &candidates[candidateIndex]
		if best >= 0 && value.area >= candidates[best].area {
			continue
		}
		if !visibleLineworkBoundsContainAll(value.bounds, value.polygon, sentinels) ||
			!visibleLineworkBoundsContainAll(value.bounds, value.polygon, probes) {
			continue
		}
		if !value.indexed {
			value.index = newVisibleLineworkPolygonIndex(value.polygon)
			value.indexed = true
		}
		if !visibleLineworkContainsAllIndexed(
			&value.index, value.bounds, sentinels, boundaryTolerance,
		) {
			continue
		}
		if visibleLineworkContainsAllIndexed(
			&value.index, value.bounds, probes, boundaryTolerance,
		) {
			best = candidateIndex
		}
	}
	if best >= 0 {
		return [][]GeoPoint{candidates[best].polygon}
	}

	for candidateIndex := range candidates {
		value := &candidates[candidateIndex]
		value.covers = make([]bool, len(probes))
		for probeIndex, probe := range probes {
			if !value.bounds.contains(value.polygon, probe) {
				continue
			}
			if !value.indexed {
				value.index = newVisibleLineworkPolygonIndex(value.polygon)
				value.indexed = true
			}
			if visibleLineworkContainsWithinIndexed(&value.index, probe, boundaryTolerance) {
				value.covers[probeIndex] = true
				value.count++
			}
		}
	}
	covered := make([]bool, len(probes))
	remaining := len(probes)
	selected := make([][]GeoPoint, 0, len(candidates))
	used := make([]bool, len(candidates))
	for remaining > 0 {
		best, bestGain := -1, 0
		for index, value := range candidates {
			if used[index] {
				continue
			}
			gain := 0
			for probeIndex, contains := range value.covers {
				if contains && !covered[probeIndex] {
					gain++
				}
			}
			if gain > bestGain || gain == bestGain && gain > 0 &&
				(best < 0 || value.area < candidates[best].area) {
				best, bestGain = index, gain
			}
		}
		if best < 0 || bestGain == 0 {
			// A handful of probes can land exactly on a numerically ambiguous
			// junction after projection. Keep the already selected faces when
			// they cover the overwhelming majority of source probes; the miss
			// audit below still rejects a materially incomplete outline.
			allowedUncovered := len(probes) / 100
			if allowedUncovered < 1 {
				allowedUncovered = 1
			}
			if allowedUncovered > 8 {
				allowedUncovered = 8
			}
			if remaining <= allowedUncovered && len(selected) > 0 {
				break
			}
			return nil
		}
		used[best] = true
		selected = append(selected, candidates[best].polygon)
		for probeIndex, contains := range candidates[best].covers {
			if contains && !covered[probeIndex] {
				covered[probeIndex] = true
				remaining--
			}
		}
	}
	return selected
}

func visibleLineworkBoundsContainAll(
	bounds visibleLineworkBounds,
	polygon, probes []GeoPoint,
) bool {
	for _, probe := range probes {
		if !bounds.contains(polygon, probe) {
			return false
		}
	}
	return true
}

func sampleVisibleLineworkProbes(probes []GeoPoint, maximumPoints int) []GeoPoint {
	if maximumPoints < 2 || len(probes) <= maximumPoints {
		return probes
	}
	result := make([]GeoPoint, maximumPoints)
	for index := range result {
		probeIndex := index * (len(probes) - 1) / (maximumPoints - 1)
		result[index] = probes[probeIndex]
	}
	return result
}

func limitVisibleLineworkCoverageProbes(probes []GeoPoint, maximum int) []GeoPoint {
	if maximum < 1 || len(probes) <= maximum {
		return probes
	}
	selected := make([]bool, len(probes))
	reserve := maximum / 4
	if reserve < 1 {
		reserve = 1
	}
	indices := make([]int, len(probes))
	for index := range indices {
		indices[index] = index
	}
	sort.SliceStable(indices, func(first, second int) bool {
		return probes[indices[first]].Latitude > probes[indices[second]].Latitude
	})
	for index := 0; index < reserve && index < len(indices); index++ {
		selected[indices[index]] = true
		selected[indices[len(indices)-1-index]] = true
	}
	remaining := maximum
	for _, value := range selected {
		if value {
			remaining--
		}
	}
	if remaining < 1 {
		remaining = 1
	}
	stride := float64(len(probes)) / float64(remaining)
	result := make([]GeoPoint, 0, maximum)
	for index, value := range selected {
		if value {
			result = append(result, probes[index])
		}
	}
	for cursor := 0.0; len(result) < maximum && int(cursor) < len(probes); cursor += stride {
		index := int(cursor)
		if selected[index] {
			continue
		}
		selected[index] = true
		result = append(result, probes[index])
	}
	return result
}

func visibleLineworkContainsAll(
	polygon []GeoPoint,
	bounds visibleLineworkBounds,
	probes []GeoPoint,
	tolerance float64,
) bool {
	for _, probe := range probes {
		if !bounds.contains(polygon, probe) ||
			!visibleLineworkContainsWithin(polygon, probe, tolerance) {
			return false
		}
	}
	return true
}

const visibleLineworkLatitudeBins = 64

type visibleLineworkPolygonIndex struct {
	polygon     []GeoPoint
	minLatitude float64
	maxLatitude float64
	bins        [][]int
	marks       []uint32
	generation  uint32
}

func newVisibleLineworkPolygonIndex(polygon []GeoPoint) visibleLineworkPolygonIndex {
	index := visibleLineworkPolygonIndex{polygon: polygon}
	if len(polygon) == 0 {
		return index
	}
	index.minLatitude, index.maxLatitude = polygon[0].Latitude, polygon[0].Latitude
	for _, point := range polygon[1:] {
		index.minLatitude = math.Min(index.minLatitude, point.Latitude)
		index.maxLatitude = math.Max(index.maxLatitude, point.Latitude)
	}
	index.bins = make([][]int, visibleLineworkLatitudeBins)
	index.marks = make([]uint32, len(polygon))
	for edge := range polygon {
		previous := (edge + len(polygon) - 1) % len(polygon)
		first, second := polygon[previous], polygon[edge]
		firstBin := index.latitudeBin(math.Min(first.Latitude, second.Latitude))
		lastBin := index.latitudeBin(math.Max(first.Latitude, second.Latitude))
		for bin := firstBin; bin <= lastBin; bin++ {
			index.bins[bin] = append(index.bins[bin], edge)
		}
	}
	return index
}

func (index visibleLineworkPolygonIndex) latitudeBin(latitude float64) int {
	if index.maxLatitude <= index.minLatitude {
		return 0
	}
	fraction := (latitude - index.minLatitude) / (index.maxLatitude - index.minLatitude)
	if fraction <= 0 {
		return 0
	}
	if fraction >= 1 {
		return visibleLineworkLatitudeBins - 1
	}
	return int(fraction * visibleLineworkLatitudeBins)
}

func (index *visibleLineworkPolygonIndex) nextGeneration() uint32 {
	index.generation++
	if index.generation == 0 {
		for mark := range index.marks {
			index.marks[mark] = 0
		}
		index.generation = 1
	}
	return index.generation
}

func visibleLineworkContainsAllIndexed(
	index *visibleLineworkPolygonIndex,
	bounds visibleLineworkBounds,
	probes []GeoPoint,
	tolerance float64,
) bool {
	for _, probe := range probes {
		if !bounds.contains(index.polygon, probe) ||
			!visibleLineworkContainsWithinIndexed(index, probe, tolerance) {
			return false
		}
	}
	return true
}

func visibleLineworkContainsWithinIndexed(
	index *visibleLineworkPolygonIndex,
	point GeoPoint,
	tolerance float64,
) bool {
	if index == nil {
		return visibleLineworkContainsWithin(nil, point, tolerance)
	}
	polygon := index.polygon
	if len(polygon) < 3 || tolerance <= 0 {
		return visibleLineworkContainsWithin(polygon, point, tolerance)
	}
	point = visibleLineworkPointNearPolygon(polygon, point)
	if point.Latitude < index.minLatitude-tolerance || point.Latitude > index.maxLatitude+tolerance {
		return false
	}
	firstBin := index.latitudeBin(point.Latitude-tolerance) - 2
	lastBin := index.latitudeBin(point.Latitude+tolerance) + 2
	if firstBin < 0 {
		firstBin = 0
	}
	if lastBin >= visibleLineworkLatitudeBins {
		lastBin = visibleLineworkLatitudeBins - 1
	}
	generation := index.nextGeneration()
	inside := false
	for bin := firstBin; bin <= lastBin; bin++ {
		for _, edge := range index.bins[bin] {
			if index.marks[edge] == generation {
				continue
			}
			index.marks[edge] = generation
			current, previous := edge, (edge+len(polygon)-1)%len(polygon)
			first, second := polygon[previous], polygon[current]
			if (first.Latitude > point.Latitude) == (second.Latitude > point.Latitude) {
				continue
			}
			intersection := first.Longitude +
				(point.Latitude-first.Latitude)*(second.Longitude-first.Longitude)/(second.Latitude-first.Latitude)
			if intersection >= point.Longitude {
				inside = !inside
			}
		}
	}
	if inside {
		return true
	}

	generation = index.nextGeneration()
	toleranceSquared := tolerance * tolerance
	for bin := firstBin; bin <= lastBin; bin++ {
		for _, edge := range index.bins[bin] {
			if index.marks[edge] == generation {
				continue
			}
			index.marks[edge] = generation
			current, previous := edge, (edge+len(polygon)-1)%len(polygon)
			start, end := polygon[previous], polygon[current]
			if point.Longitude < math.Min(start.Longitude, end.Longitude)-tolerance ||
				point.Longitude > math.Max(start.Longitude, end.Longitude)+tolerance ||
				point.Latitude < math.Min(start.Latitude, end.Latitude)-tolerance ||
				point.Latitude > math.Max(start.Latitude, end.Latitude)+tolerance {
				continue
			}
			if visibleLineworkPointSegmentDistanceSquared(point, start, end) <= toleranceSquared {
				return true
			}
		}
	}
	return false
}

func visibleLineworkCyclesCoverCoverage(
	cycles, coveragePaths [][]GeoPoint,
	tolerance float64,
) bool {
	return visibleLineworkCoverageMissCount(cycles, coveragePaths, tolerance) == 0
}

func visibleLineworkCoverageMissCount(
	cycles, coveragePaths [][]GeoPoint,
	tolerance float64,
) int {
	if len(cycles) == 0 {
		return -1
	}
	probes := visibleLineworkCoverageProbes(coveragePaths, 8)
	probes = limitVisibleLineworkCoverageProbes(probes, 768)
	if len(probes) == 0 {
		return -1
	}
	bounds := make([]visibleLineworkBounds, len(cycles))
	for index, cycle := range cycles {
		bounds[index] = visibleLineworkPolygonBounds(cycle, tolerance)
	}
	misses := 0
	for _, probe := range probes {
		covered := false
		for index, cycle := range cycles {
			if bounds[index].contains(cycle, probe) &&
				visibleLineworkContainsWithin(cycle, probe, tolerance) {
				covered = true
				break
			}
		}
		if !covered {
			misses++
		}
	}
	return misses
}

func visibleLineworkPolygonBounds(polygon []GeoPoint, tolerance float64) visibleLineworkBounds {
	if len(polygon) == 0 {
		return visibleLineworkBounds{}
	}
	bounds := visibleLineworkBounds{
		minLongitude: polygon[0].Longitude,
		maxLongitude: polygon[0].Longitude,
		minLatitude:  polygon[0].Latitude,
		maxLatitude:  polygon[0].Latitude,
	}
	for _, point := range polygon[1:] {
		bounds.minLongitude = math.Min(bounds.minLongitude, point.Longitude)
		bounds.maxLongitude = math.Max(bounds.maxLongitude, point.Longitude)
		bounds.minLatitude = math.Min(bounds.minLatitude, point.Latitude)
		bounds.maxLatitude = math.Max(bounds.maxLatitude, point.Latitude)
	}
	return visibleLineworkBounds{
		minLongitude: bounds.minLongitude - tolerance,
		maxLongitude: bounds.maxLongitude + tolerance,
		minLatitude:  bounds.minLatitude - tolerance,
		maxLatitude:  bounds.maxLatitude + tolerance,
	}
}

func (bounds visibleLineworkBounds) contains(polygon []GeoPoint, point GeoPoint) bool {
	point = visibleLineworkPointNearPolygon(polygon, point)
	return point.Longitude >= bounds.minLongitude && point.Longitude <= bounds.maxLongitude &&
		point.Latitude >= bounds.minLatitude && point.Latitude <= bounds.maxLatitude
}

func visibleLineworkCoverageProbes(paths [][]GeoPoint, maximumPoints int) []GeoPoint {
	if maximumPoints < 2 {
		maximumPoints = 2
	}
	probes := make([]GeoPoint, 0, len(paths)*maximumPoints)
	for _, path := range paths {
		step := (len(path) + maximumPoints - 1) / maximumPoints
		if step < 1 {
			step = 1
		}
		for index := 0; index < len(path); index += step {
			probes = append(probes, path[index])
		}
		if len(path) > 0 && (len(path)-1)%step != 0 {
			probes = append(probes, path[len(path)-1])
		}
	}
	return probes
}

func visibleLineworkContainsWithin(polygon []GeoPoint, point GeoPoint, tolerance float64) bool {
	if len(polygon) < 3 {
		return false
	}
	point = visibleLineworkPointNearPolygon(polygon, point)
	if tolerance <= 0 {
		return sweepPointInPolygon(polygon, point)
	}
	inside := false
	for current, previous := 0, len(polygon)-1; current < len(polygon); previous, current = current, current+1 {
		first, second := polygon[previous], polygon[current]
		if (first.Latitude > point.Latitude) == (second.Latitude > point.Latitude) {
			continue
		}
		intersection := first.Longitude +
			(point.Latitude-first.Latitude)*(second.Longitude-first.Longitude)/(second.Latitude-first.Latitude)
		if intersection >= point.Longitude {
			inside = !inside
		}
	}
	if inside {
		return true
	}
	toleranceSquared := tolerance * tolerance
	for current, previous := 0, len(polygon)-1; current < len(polygon); previous, current = current, current+1 {
		start, end := polygon[previous], polygon[current]
		if point.Longitude < math.Min(start.Longitude, end.Longitude)-tolerance ||
			point.Longitude > math.Max(start.Longitude, end.Longitude)+tolerance ||
			point.Latitude < math.Min(start.Latitude, end.Latitude)-tolerance ||
			point.Latitude > math.Max(start.Latitude, end.Latitude)+tolerance {
			continue
		}
		if visibleLineworkPointSegmentDistanceSquared(point, start, end) <= toleranceSquared {
			return true
		}
	}
	return false
}

func visibleLineworkPointNearPolygon(polygon []GeoPoint, point GeoPoint) GeoPoint {
	if len(polygon) == 0 {
		return point
	}
	longitude := point.Longitude
	for longitude-polygon[0].Longitude > 180 {
		longitude -= 360
	}
	for longitude-polygon[0].Longitude < -180 {
		longitude += 360
	}
	point.Longitude = longitude
	return point
}

func visibleLineworkPointSegmentDistanceSquared(point, start, end GeoPoint) float64 {
	dx := end.Longitude - start.Longitude
	dy := end.Latitude - start.Latitude
	denominator := dx*dx + dy*dy
	if denominator <= 1e-20 {
		dx = point.Longitude - start.Longitude
		dy = point.Latitude - start.Latitude
		return dx*dx + dy*dy
	}
	fraction := ((point.Longitude-start.Longitude)*dx + (point.Latitude-start.Latitude)*dy) / denominator
	fraction = math.Max(0, math.Min(1, fraction))
	dx = point.Longitude - (start.Longitude + fraction*dx)
	dy = point.Latitude - (start.Latitude + fraction*dy)
	return dx*dx + dy*dy
}

func sampleVisibleLineworkFillPaths(paths [][]GeoPoint, maximumPoints int) [][]GeoPoint {
	if maximumPoints < 3 {
		maximumPoints = 3
	}
	result := make([][]GeoPoint, 0, len(paths))
	for _, path := range paths {
		path = openGeoRing(path)
		if len(path) <= maximumPoints {
			result = append(result, path)
			continue
		}
		indices := make([]int, 0, maximumPoints)
		seen := make(map[int]bool, maximumPoints)
		for direction := 0; direction < maximumPoints; direction++ {
			angle := 2 * math.Pi * float64(direction) / float64(maximumPoints)
			x, y := math.Cos(angle), math.Sin(angle)
			bestIndex := 0
			bestValue := path[0].Longitude*x + path[0].Latitude*y
			for index := 1; index < len(path); index++ {
				value := path[index].Longitude*x + path[index].Latitude*y
				if value > bestValue {
					bestIndex, bestValue = index, value
				}
			}
			if !seen[bestIndex] {
				seen[bestIndex] = true
				indices = append(indices, bestIndex)
			}
		}
		sort.Ints(indices)
		sampled := make([]GeoPoint, 0, len(indices))
		for _, index := range indices {
			sampled = append(sampled, path[index])
		}
		result = append(result, sampled)
	}
	return result
}

func unprojectVisibleLineworkPaths(chart polygonUnionChart, paths [][]GeoPoint) [][]GeoPoint {
	result := make([][]GeoPoint, len(paths))
	for pathIndex, path := range paths {
		result[pathIndex] = make([]GeoPoint, len(path))
		for pointIndex, point := range path {
			result[pathIndex][pointIndex] = chart.unproject(point)
		}
	}
	return result
}

func visibleLineworkPointPaths(paths [][]GeoPoint) [][]GeoPoint {
	result := make([][]GeoPoint, 0)
	for _, path := range paths {
		for _, point := range path {
			result = append(result, []GeoPoint{point})
		}
	}
	return result
}

type visibleLineworkCycleStep struct {
	edge    int
	reverse bool
}

func enumerateVisibleLineworkCycles(
	nodes []visibleLineworkNode,
	edges []visibleLineworkEdge,
) [][]GeoPoint {
	// Walk the face on each side of every edge once. The exterior face of
	// each component is retained as a candidate for its complete outer band.
	// Enumerating arbitrary combinations of interior faces grows exponentially.
	order := make([][]int, len(nodes))
	position := make([]int, 2*len(edges))
	halfAngle := make([]float64, 2*len(edges))
	for index, edge := range edges {
		// 正向半边的方向取"首点到相邻的第二点"，与反向半边的"末点到倒数第二点"对称；
		// 取首末两点会在折线上给出不同的排序角度，进而改变环枚举结果。
		first, next := edge.points[0], edge.points[1]
		previous, last := edge.points[len(edge.points)-2], edge.points[len(edge.points)-1]
		halfAngle[2*index] = math.Atan2(next.Latitude-first.Latitude, next.Longitude-first.Longitude)
		halfAngle[2*index+1] = math.Atan2(previous.Latitude-last.Latitude, previous.Longitude-last.Longitude)
	}
	for node, value := range nodes {
		for _, half := range value.outgoing {
			id := 2 * half.edge
			if half.reverse {
				id++
			}
			order[node] = append(order[node], id)
		}
		sort.SliceStable(order[node], func(i, j int) bool {
			return halfAngle[order[node][i]] < halfAngle[order[node][j]]
		})
		for i, half := range order[node] {
			position[half] = i
		}
	}
	visited := make([]bool, 2*len(edges))
	cycles := make([][]GeoPoint, 0, len(edges))
	seen := make(map[uint64][][]int32)
	for start := range visited {
		if visited[start] {
			continue
		}
		var steps []visibleLineworkCycleStep
		for half := start; !visited[half]; {
			visited[half] = true
			step := visibleLineworkCycleStep{edge: half / 2, reverse: half%2 != 0}
			if len(steps) > 0 && steps[len(steps)-1].edge == step.edge && steps[len(steps)-1].reverse != step.reverse {
				steps = steps[:len(steps)-1]
			} else {
				steps = append(steps, step)
			}
			next := edges[step.edge].end
			if step.reverse {
				next = edges[step.edge].start
			}
			out := order[next]
			half = out[(position[half^1]+len(out)-1)%len(out)]
		}
		for len(steps) > 1 && steps[0].edge == steps[len(steps)-1].edge && steps[0].reverse != steps[len(steps)-1].reverse {
			steps = steps[1 : len(steps)-1]
		}
		if len(steps) < 1 {
			continue
		}
		for _, simple := range visibleLineworkSimpleCycles(edges, steps) {
			if visibleLineworkCycleSeen(seen, simple) {
				continue
			}
			if ring := visibleLineworkCyclePoints(edges, simple); len(ring) >= 3 && math.Abs(visibleLineworkSignedArea(ring)) > 1e-12 {
				cycles = append(cycles, ring)
			}
		}
	}
	return cycles
}

// The exterior face can revisit an articulation node. Separate its lobes so
// each candidate is a simple ring, including components joined by a bridge.
func visibleLineworkSimpleCycles(edges []visibleLineworkEdge, steps []visibleLineworkCycleStep) [][]visibleLineworkCycleStep {
	var cycles [][]visibleLineworkCycleStep
	var path []visibleLineworkCycleStep
	positions := make(map[int]int)
	for _, step := range steps {
		start, end := edges[step.edge].start, edges[step.edge].end
		if step.reverse {
			start, end = end, start
		}
		positions[start] = len(path)
		path = append(path, step)
		if at, ok := positions[end]; ok {
			cycles = append(cycles, append([]visibleLineworkCycleStep(nil), path[at:]...))
			for _, removed := range path[at:] {
				node := edges[removed.edge].start
				if removed.reverse {
					node = edges[removed.edge].end
				}
				delete(positions, node)
			}
			path = path[:at]
		}
	}
	return cycles
}

// visibleLineworkCycleSeen 按与方向无关的边集合去重，返回该环是否已出现。
func visibleLineworkCycleSeen(seen map[uint64][][]int32, steps []visibleLineworkCycleStep) bool {
	key := make([]int32, len(steps))
	for index, step := range steps {
		key[index] = int32(step.edge)
	}
	sort.Slice(key, func(first, second int) bool { return key[first] < key[second] })
	hash := uint64(14695981039346656037)
	for _, edge := range key {
		hash = (hash ^ uint64(uint32(edge))) * 1099511628211
	}
	for _, existing := range seen[hash] {
		if len(existing) != len(key) {
			continue
		}
		same := true
		for index, edge := range key {
			if existing[index] != edge {
				same = false
				break
			}
		}
		if same {
			return true
		}
	}
	seen[hash] = append(seen[hash], key)
	return false
}

func visibleLineworkCyclePoints(
	edges []visibleLineworkEdge,
	steps []visibleLineworkCycleStep,
) []GeoPoint {
	ring := make([]GeoPoint, 0, len(steps)*4)
	for _, step := range steps {
		points := visibleLineworkDirectedPoints(edges[step.edge], step.reverse)
		if len(ring) == 0 {
			ring = append(ring, points...)
		} else {
			ring = append(ring, points[1:]...)
		}
	}
	return sweepDeduplicateAdjacent(openGeoRing(ring))
}

func projectVisibleLine(chart polygonUnionChart, source []GeoPoint) []GeoPoint {
	if len(source) == 0 {
		return nil
	}
	line := make([]GeoPoint, 0, len(source))
	for sourceIndex, point := range source {
		projected := chart.project(point)
		if len(line) > 0 {
			for projected.Longitude-line[len(line)-1].Longitude > 180 {
				projected.Longitude -= 360
			}
			for projected.Longitude-line[len(line)-1].Longitude < -180 {
				projected.Longitude += 360
			}
		}
		if len(line) > 0 {
			if math.Hypot(projected.Longitude-line[len(line)-1].Longitude, projected.Latitude-line[len(line)-1].Latitude) > 1e-10 {
				line = appendVisibleProjectedArc(line, chart, source[sourceIndex-1], point, line[len(line)-1], projected, 0)
			}
			continue
		}
		if len(line) == 0 || math.Hypot(
			projected.Longitude-line[len(line)-1].Longitude,
			projected.Latitude-line[len(line)-1].Latitude,
		) > 1e-10 {
			line = append(line, projected)
		}
	}
	return line
}

// A great-circle arc becomes curved after rotation into the topology chart.
// Bound its projected chord error before intersections and face containment.
func appendVisibleProjectedArc(
	points []GeoPoint,
	chart polygonUnionChart,
	first, last, start, end GeoPoint,
	depth int,
) []GeoPoint {
	if depth < 12 && (math.Abs(end.Longitude-start.Longitude) > 0.25 ||
		math.Abs(end.Latitude-start.Latitude) > 0.25 ||
		math.Max(math.Abs(start.Latitude), math.Abs(end.Latitude)) > 75) {
		middle := sphericalInterpolate(first, last, 0.5)
		projected := chart.project(middle)
		projected.Longitude = start.Longitude + math.Remainder(projected.Longitude-start.Longitude, 360)
		if visibleLineworkPointSegmentDistanceSquared(projected, start, end) > 0.002*0.002 {
			points = appendVisibleProjectedArc(points, chart, first, middle, start, projected, depth+1)
			return appendVisibleProjectedArc(points, chart, middle, last, projected, end, depth+1)
		}
	}
	return append(points, end)
}

func buildVisibleLineworkGraph(
	lines [][]GeoPoint,
	chart polygonUnionChart,
	snapDistanceKM float64,
) ([]visibleLineworkNode, []visibleLineworkEdge, [][]GeoPoint) {
	const commonJunctionKM = 1e-5
	nodes, edges, loops := buildVisibleLineworkGraphAtTolerance(lines, chart, commonJunctionKM)
	if snapDistanceKM <= 0 {
		snapDistanceKM = 1
	}
	changed := false
	for {
		first, second := -1, -1
		minimum := snapDistanceKM
		// Connected intersections are distinct even when physically close.
		// Only an open endpoint can require the wider sampling snap radius.
		for index, node := range nodes {
			if len(node.outgoing) != 1 {
				continue
			}
			point := chart.unproject(node.point)
			for otherIndex, other := range nodes {
				if index == otherIndex || len(other.outgoing) == 0 {
					continue
				}
				if distance := EquatorialDistanceKM(point, chart.unproject(other.point)); distance < minimum {
					first, second, minimum = index, otherIndex, distance
				}
			}
		}
		if first < 0 {
			break
		}
		changed = true
		for _, half := range nodes[first].outgoing {
			edge := &edges[half.edge]
			if edge.start == first {
				edge.start, edge.points[0] = second, nodes[second].point
			}
			if edge.end == first {
				edge.end, edge.points[len(edge.points)-1] = second, nodes[second].point
			}
			nodes[second].outgoing = append(nodes[second].outgoing, half)
		}
		nodes[first].outgoing = nil
	}
	if !changed {
		return nodes, edges, loops
	}
	joined := make([][]GeoPoint, 0, len(edges)+len(loops))
	for _, edge := range edges {
		joined = append(joined, edge.points)
	}
	for _, loop := range loops {
		joined = append(joined, append(append([]GeoPoint(nil), loop...), loop[0]))
	}
	// Moving an endpoint can create a crossing on its adjacent segment.
	// Re-node those intersections before ordering directed edges into faces.
	return buildVisibleLineworkGraphAtTolerance(splitVisibleLineworkIntersections(joined), chart, commonJunctionKM)
}

func buildVisibleLineworkGraphAtTolerance(
	lines [][]GeoPoint,
	chart polygonUnionChart,
	snapDistanceKM float64,
) ([]visibleLineworkNode, []visibleLineworkEdge, [][]GeoPoint) {
	if snapDistanceKM <= 0 {
		snapDistanceKM = 1
	}
	nodes := make([]visibleLineworkNode, 0, len(lines))
	edges := make([]visibleLineworkEdge, 0, len(lines))
	loops := make([][]GeoPoint, 0, 2)
	for _, source := range lines {
		line := append([]GeoPoint(nil), source...)
		start := visibleLineworkNodeIndex(nodes, chart, line[0], snapDistanceKM)
		if start < 0 {
			start = len(nodes)
			nodes = append(nodes, visibleLineworkNode{point: line[0]})
		}
		end := visibleLineworkNodeIndex(nodes, chart, line[len(line)-1], snapDistanceKM)
		if end < 0 {
			end = len(nodes)
			nodes = append(nodes, visibleLineworkNode{point: line[len(line)-1]})
		}
		line[0] = nodes[start].point
		line[len(line)-1] = nodes[end].point
		line = sweepDeduplicateAdjacent(line)
		if len(line) < 2 {
			continue
		}
		if start == end && len(openGeoRing(line)) < 3 {
			continue
		}
		duplicate := false
		for _, edge := range edges {
			if len(edge.points) != len(line) {
				continue
			}
			reverse := edge.start == end && edge.end == start
			if !reverse && (edge.start != start || edge.end != end) {
				continue
			}
			same := true
			for i, point := range line {
				j := i
				if reverse {
					j = len(line) - 1 - i
				}
				if !SameGeoPoint(point, edge.points[j]) {
					same = false
					break
				}
			}
			if same {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}
		edgeIndex := len(edges)
		edges = append(edges, visibleLineworkEdge{points: line, start: start, end: end})
		nodes[start].outgoing = append(nodes[start].outgoing, visibleLineworkHalfEdge{
			edge: edgeIndex,
		})
		nodes[end].outgoing = append(nodes[end].outgoing, visibleLineworkHalfEdge{
			edge: edgeIndex, reverse: true,
		})
	}
	return nodes, edges, loops
}

func visibleLineworkNodeIndex(
	nodes []visibleLineworkNode,
	chart polygonUnionChart,
	projected GeoPoint,
	snapDistanceKM float64,
) int {
	target := chart.unproject(projected)
	bestIndex := -1
	bestDistance := math.Inf(1)
	for index, node := range nodes {
		distance := EquatorialDistanceKM(chart.unproject(node.point), target)
		if distance <= snapDistanceKM && distance < bestDistance {
			bestIndex, bestDistance = index, distance
		}
	}
	return bestIndex
}

func visibleLineworkDirectedPoints(edge visibleLineworkEdge, reverse bool) []GeoPoint {
	if !reverse {
		return edge.points
	}
	points := make([]GeoPoint, len(edge.points))
	for index := range edge.points {
		points[index] = edge.points[len(edge.points)-1-index]
	}
	return points
}

func visibleLineworkSignedArea(polygon []GeoPoint) float64 {
	if len(polygon) < 3 {
		return 0
	}
	area := 0.0
	for current, previous := 0, len(polygon)-1; current < len(polygon); previous, current = current, current+1 {
		area += polygon[previous].Longitude*polygon[current].Latitude -
			polygon[current].Longitude*polygon[previous].Latitude
	}
	return area / 2
}
