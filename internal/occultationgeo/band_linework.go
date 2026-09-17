package occultationgeo

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

func occultationVisibleBoundaryLines(
	fallbackPolygons [][]geodata.GeoPoint,
	curves []basic.OccultationRiseSetCurve,
	extraLines [][]geodata.GeoPoint,
) [][]geodata.GeoPoint {
	return occultationVisibleBoundaryLinesFromBase(
		occultationFallbackPolygonBoundaryLines(fallbackPolygons), curves, extraLines,
	)
}

func occultationFallbackPolygonBoundaryLines(
	fallbackPolygons [][]geodata.GeoPoint,
) [][]geodata.GeoPoint {
	boundaryLines := make([][]geodata.GeoPoint, 0, len(fallbackPolygons))
	for _, polygon := range fallbackPolygons {
		if len(polygon) < 3 {
			continue
		}
		line := append([]geodata.GeoPoint(nil), polygon...)
		if !geodata.SameGeoPoint(line[0], line[len(line)-1]) {
			line = append(line, line[0])
		}
		boundaryLines = append(boundaryLines, line)
	}
	return boundaryLines
}

func occultationContactContourBoundaryLines(
	contours [][]basic.OccultationPathPoint,
) [][]geodata.GeoPoint {
	boundaryLines := make([][]geodata.GeoPoint, 0, len(contours))
	for _, contour := range contours {
		for _, sampleRange := range ContinuousBoundaryRanges(contour) {
			if sampleRange.End-sampleRange.Start < 2 {
				continue
			}
			line := occultationPathGeoLine(contour[sampleRange.Start:sampleRange.End])
			if len(line) >= 2 {
				boundaryLines = append(boundaryLines, line)
			}
		}
	}
	return boundaryLines
}

func occultationStaticBandCurves(
	curves []basic.OccultationRiseSetCurve,
) []basic.OccultationRiseSetCurve {
	if len(curves) == 0 {
		return nil
	}
	result := make([]basic.OccultationRiseSetCurve, 0, len(curves))
	for _, curve := range curves {
		if curve.Phase == basic.RiseSetPhaseGreatest {
			continue
		}
		result = append(result, curve)
	}
	return result
}

func occultationVisibleBoundaryLinesFromBase(
	baseLines [][]geodata.GeoPoint,
	curves []basic.OccultationRiseSetCurve,
	extraLines [][]geodata.GeoPoint,
) [][]geodata.GeoPoint {
	boundaryLines := make([][]geodata.GeoPoint, 0, len(baseLines)+len(curves)*2+len(extraLines))
	for _, source := range baseLines {
		if len(source) < 2 {
			continue
		}
		line := append([]geodata.GeoPoint(nil), source...)
		boundaryLines = append(boundaryLines, line)
	}
	for _, curve := range curves {
		boundaryLines = append(boundaryLines, occultationCurveBoundaryLines(curve)...)
	}
	boundaryLines = append(boundaryLines, extraLines...)
	return boundaryLines
}

func occultationHorizonConnectorBoundaryLines(
	connectors []HorizonConnector,
) [][]geodata.GeoPoint {
	lines := make([][]geodata.GeoPoint, 0, len(connectors))
	for _, connector := range connectors {
		if len(connector.Points) < 2 {
			continue
		}
		line := make([]geodata.GeoPoint, len(connector.Points))
		for index, point := range connector.Points {
			line[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		lines = append(lines, line)
	}
	return lines
}

type occultationPhaseBoundaryEdge struct {
	points []geodata.GeoPoint
}

type occultationPhaseBoundaryNode struct {
	point geodata.GeoPoint
	edges []int
}

// occultationPhaseBoundaryPolygons constructs the visible outer band from
// start/end phase curves and horizon closures. Greatest is intentionally
// excluded: it is a diagnostic stage curve inside the visible region, not an
// exterior boundary. The returned rings are accepted only when the cycle
// covers the continuous contact envelope and its own edges remain close to
// that envelope.
func occultationPhaseBoundaryPolygons(
	curves []basic.OccultationRiseSetCurve,
	connectors []HorizonConnector,
	fillPolygons [][]geodata.GeoPoint,
) ([][]geodata.GeoPoint, bool) {
	if len(fillPolygons) == 0 {
		return nil, false
	}
	edgesByDirection := make(map[basic.RiseSetDirection][]occultationPhaseBoundaryEdge)
	for _, curve := range curves {
		if curve.Phase == basic.RiseSetPhaseGreatest {
			continue
		}
		if curve.Phase != basic.RiseSetPhaseStart &&
			curve.Phase != basic.RiseSetPhaseEnd {
			continue
		}
		for _, line := range occultationCurveBoundaryLines(curve) {
			if len(line) >= 2 {
				edgesByDirection[curve.Direction] = append(
					edgesByDirection[curve.Direction],
					occultationPhaseBoundaryEdge{points: line},
				)
			}
		}
	}
	for _, connector := range connectors {
		if len(connector.Points) < 2 {
			continue
		}
		edgesByDirection[connector.Direction] = append(
			edgesByDirection[connector.Direction],
			occultationPhaseBoundaryEdge{
				points: occultationPathGeoLine(connector.Points),
			},
		)
	}
	if len(edgesByDirection) == 0 {
		return nil, false
	}

	// 每个通过门槛的环都是一条独立相位支路，并集结果与顺序无关；按分数只取"最优环"会
	// 丢掉真实支路，因此这里保留全部环，只按方向键固定生成顺序以保证可复现。
	directions := make([]basic.RiseSetDirection, 0, len(edgesByDirection))
	for direction := range edgesByDirection {
		directions = append(directions, direction)
	}
	sort.Slice(directions, func(first, second int) bool { return directions[first] < directions[second] })
	rings := make([][]geodata.GeoPoint, 0, len(directions))
	for _, direction := range directions {
		for _, ring := range occultationPhaseBoundaryCyclesForEdges(edgesByDirection[direction]) {
			if len(ring) < 4 {
				continue
			}
			fillMiss := geodata.SphericalPolygonsPathMissDistanceKM(
				[][]geodata.GeoPoint{ring}, fillPolygons, true,
			)
			if fillMiss > 150 {
				continue
			}
			edgeMiss := geodata.SphericalPolygonsPathMissDistanceKM(
				fillPolygons, [][]geodata.GeoPoint{ring}, true,
			)
			if edgeMiss > 200 {
				continue
			}
			rings = append(rings, ring)
		}
	}
	if len(rings) == 0 {
		return nil, false
	}
	if len(rings) == 1 {
		return rings, true
	}
	merged, err := geodata.UnionPolygons(rings)
	if err != nil || len(merged) == 0 {
		return nil, false
	}
	return merged, true
}

func occultationPhaseBoundaryCyclesForEdges(
	edges []occultationPhaseBoundaryEdge,
) [][]geodata.GeoPoint {
	if len(edges) < 2 {
		return nil
	}
	nodes := make([]occultationPhaseBoundaryNode, 0, len(edges)*2)
	edgeNodes := make([][2]int, len(edges))
	validEdges := make([]bool, len(edges))
	nodeFor := func(point geodata.GeoPoint) int {
		for index := range nodes {
			if geoDistanceKM(nodes[index].point, point) <= curveBoundaryJoinDistanceKM {
				return index
			}
		}
		nodes = append(nodes, occultationPhaseBoundaryNode{point: point})
		return len(nodes) - 1
	}
	for edgeIndex, edge := range edges {
		if len(edge.points) < 2 {
			continue
		}
		start := nodeFor(edge.points[0])
		end := nodeFor(edge.points[len(edge.points)-1])
		if start == end {
			continue
		}
		validEdges[edgeIndex] = true
		edgeNodes[edgeIndex] = [2]int{start, end}
		nodes[start].edges = append(nodes[start].edges, edgeIndex)
		nodes[end].edges = append(nodes[end].edges, edgeIndex)
	}

	type cycleResult struct {
		ring      []geodata.GeoPoint
		edgeIDs   []int
		signature string
	}
	results := make([]cycleResult, 0)
	seen := make(map[string]struct{})
	for startEdge, edge := range edges {
		if len(edge.points) < 2 || !validEdges[startEdge] {
			continue
		}
		for _, forward := range []bool{true, false} {
			startNode := edgeNodes[startEdge][0]
			currentNode := edgeNodes[startEdge][1]
			points := append([]geodata.GeoPoint(nil), edge.points...)
			if !forward {
				startNode, currentNode = currentNode, startNode
				reverseGeoPointRing(points)
			}
			used := map[int]bool{startEdge: true}
			edgeIDs := []int{startEdge}
			var walk func(int, []geodata.GeoPoint, map[int]bool, []int)
			walk = func(node int, ring []geodata.GeoPoint, used map[int]bool, usedEdges []int) {
				if node == startNode {
					if len(usedEdges) < 2 || len(ring) < 3 {
						return
					}
					closed := append([]geodata.GeoPoint(nil), ring...)
					closed = append(closed, closed[0])
					ids := append([]int(nil), usedEdges...)
					sort.Ints(ids)
					signature := fmt.Sprint(ids)
					if _, exists := seen[signature]; exists {
						return
					}
					seen[signature] = struct{}{}
					results = append(results, cycleResult{
						ring: closed, edgeIDs: ids, signature: signature,
					})
					return
				}
				if len(usedEdges) >= len(edges) {
					return
				}
				for _, nextEdge := range nodes[node].edges {
					if used[nextEdge] || !validEdges[nextEdge] || len(edges[nextEdge].points) < 2 {
						continue
					}
					next := edges[nextEdge].points
					nextNode := edgeNodes[nextEdge][1]
					if edgeNodes[nextEdge][1] == node {
						nextNode = edgeNodes[nextEdge][0]
						next = append([]geodata.GeoPoint(nil), next...)
						reverseGeoPointRing(next)
					}
					nextRing := append([]geodata.GeoPoint(nil), ring...)
					if len(nextRing) > 0 && geoDistanceKM(nextRing[len(nextRing)-1], next[0]) <= curveBoundaryJoinDistanceKM {
						next[0] = nextRing[len(nextRing)-1]
					}
					nextRing = append(nextRing, next[1:]...)
					nextUsed := make(map[int]bool, len(used)+1)
					for key, value := range used {
						nextUsed[key] = value
					}
					nextUsed[nextEdge] = true
					nextEdges := append(append([]int(nil), usedEdges...), nextEdge)
					walk(nextNode, nextRing, nextUsed, nextEdges)
				}
			}
			walk(currentNode, points, used, edgeIDs)
		}
	}
	rings := make([][]geodata.GeoPoint, 0, len(results))
	for _, result := range results {
		rings = append(rings, result.ring)
	}
	return rings
}

// occultationCurveBoundaryAlternatives returns boundary line sets that replace
// one folded multi-branch curve with a single raw branch. The alternatives are
// intentionally bounded: they are only evaluated after the normal topology
// fails, and the shortest branch is tried first because it is the usual polar
// fold connector rather than the long interior branch.
func occultationCurveBoundaryAlternatives(
	fallbackPolygons [][]geodata.GeoPoint,
	curves []basic.OccultationRiseSetCurve,
	extraLines [][]geodata.GeoPoint,
) [][][]geodata.GeoPoint {
	return occultationCurveBoundaryAlternativesFromBase(
		occultationFallbackPolygonBoundaryLines(fallbackPolygons), curves, extraLines,
	)
}

func occultationCurveBoundaryAlternativesFromBase(
	baseLines [][]geodata.GeoPoint,
	curves []basic.OccultationRiseSetCurve,
	extraLines [][]geodata.GeoPoint,
) [][][]geodata.GeoPoint {
	base := occultationVisibleBoundaryLinesFromBase(baseLines, nil, extraLines)
	curveLines := make([][][]geodata.GeoPoint, len(curves))
	for index, curve := range curves {
		curveLines[index] = occultationCurveBoundaryLines(curve)
	}
	var alternatives [][][]geodata.GeoPoint
	for curveIndex, curve := range curves {
		branches := make([][]geodata.GeoPoint, 0, len(curve.Segments))
		for _, segment := range curve.Segments {
			if len(segment) < 2 {
				continue
			}
			branch := make([]geodata.GeoPoint, len(segment))
			for pointIndex, point := range segment {
				branch[pointIndex] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			branches = append(branches, branch)
		}
		if len(branches) < 2 {
			continue
		}
		sort.SliceStable(branches, func(first, second int) bool {
			return len(branches[first]) < len(branches[second])
		})
		for _, branch := range branches {
			candidate := make([][]geodata.GeoPoint, 0, len(base)+len(curves)*2)
			candidate = append(candidate, base...)
			for otherIndex, lines := range curveLines {
				if otherIndex == curveIndex {
					candidate = append(candidate, branch)
					continue
				}
				candidate = append(candidate, lines...)
			}
			alternatives = append(alternatives, candidate)
			if len(alternatives) >= maximumOccultationBoundaryAlternatives {
				return alternatives
			}
		}
	}
	return alternatives
}

const (
	// CurveBoundaryJoinDistanceKM is deliberately much smaller than the
	// polygonizer snap radius. It only joins endpoints that represent the same
	// physical fold, rather than nearby polar branches that merely converge in
	// longitude.
	curveBoundaryJoinDistanceKM = 5.0
	curveBoundaryJoinTime       = 30 * time.Second
)

// occultationCurveBoundaryLines stitches same-curve branch segments that meet
// at a common physical endpoint. Public rise/set data keeps the original
// strictly time-ordered segments; the static fill topology needs a spatial
// edge, so a fold such as C->D and B->D becomes B->D->C.
func occultationCurveBoundaryLines(curve basic.OccultationRiseSetCurve) [][]geodata.GeoPoint {
	segments := make([][]geodata.GeoPoint, 0, len(curve.Segments))
	segmentTimes := make([][]time.Time, 0, len(curve.Segments))
	for _, source := range curve.Segments {
		if len(source) < 2 {
			continue
		}
		line := occultationPathGeoLine(source)
		times := make([]time.Time, len(source))
		for index, point := range source {
			times[index] = point.Time
		}
		segments = append(segments, line)
		segmentTimes = append(segmentTimes, times)
	}
	lines := make([][]geodata.GeoPoint, 0, len(segments))
	for len(segments) > 0 {
		line := append([]geodata.GeoPoint(nil), segments[0]...)
		lineTimes := append([]time.Time(nil), segmentTimes[0]...)
		segments = segments[1:]
		segmentTimes = segmentTimes[1:]
		for {
			joined := false
			for index := range segments {
				if !curveBoundaryEndpointsMatch(line, lineTimes, segments[index], segmentTimes[index]) {
					continue
				}
				line, lineTimes = joinCurveBoundarySegments(line, lineTimes, segments[index], segmentTimes[index])
				segments = append(segments[:index], segments[index+1:]...)
				segmentTimes = append(segmentTimes[:index], segmentTimes[index+1:]...)
				joined = true
				break
			}
			if !joined {
				break
			}
		}
		if len(line) >= 2 {
			lines = append(lines, line)
		}
	}
	return lines
}

func occultationPathGeoLine(source []basic.OccultationPathPoint) []geodata.GeoPoint {
	line := make([]geodata.GeoPoint, len(source))
	for index, point := range source {
		line[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
	}
	return line
}

// StitchedRiseSetCurveSegments 返回面向显示的升落曲线分段。
// StitchedRiseSetCurveSegments returns display-oriented rise/set curve
// segments. Solver output keeps branch segments time-ordered; map strokes need
// the spatial continuation at shared fold endpoints so a pair such as A->D and
// B->D renders as one continuous A->D->B polyline.
func StitchedRiseSetCurveSegments(curve basic.OccultationRiseSetCurve) [][]basic.OccultationPathPoint {
	segments := make([][]basic.OccultationPathPoint, 0, len(curve.Segments))
	for _, source := range curve.Segments {
		if len(source) < 2 {
			continue
		}
		segments = append(segments, append([]basic.OccultationPathPoint(nil), source...))
	}
	lines := make([][]basic.OccultationPathPoint, 0, len(segments))
	for len(segments) > 0 {
		line := append([]basic.OccultationPathPoint(nil), segments[0]...)
		segments = segments[1:]
		for {
			joined := false
			for index := range segments {
				if !riseSetCurveEndpointsMatch(line, segments[index]) {
					continue
				}
				line = joinRiseSetCurveSegments(line, segments[index])
				segments = append(segments[:index], segments[index+1:]...)
				joined = true
				break
			}
			if !joined {
				break
			}
		}
		if len(line) >= 2 {
			lines = append(lines, line)
		}
	}
	return lines
}

func curveBoundaryEndpointsMatch(
	first []geodata.GeoPoint,
	firstTimes []time.Time,
	second []geodata.GeoPoint,
	secondTimes []time.Time,
) bool {
	if len(first) < 2 || len(second) < 2 || len(firstTimes) != len(first) || len(secondTimes) != len(second) {
		return false
	}
	return (curveBoundaryEndpointMatch(first[0], firstTimes[0], second[0], secondTimes[0]) ||
		curveBoundaryEndpointMatch(first[0], firstTimes[0], second[len(second)-1], secondTimes[len(second)-1]) ||
		curveBoundaryEndpointMatch(first[len(first)-1], firstTimes[len(first)-1], second[0], secondTimes[0]) ||
		curveBoundaryEndpointMatch(first[len(first)-1], firstTimes[len(first)-1], second[len(second)-1], secondTimes[len(second)-1]))
}

func curveBoundaryEndpointMatch(first geodata.GeoPoint, firstTime time.Time, second geodata.GeoPoint, secondTime time.Time) bool {
	return geoDistanceKM(first, second) <= curveBoundaryJoinDistanceKM &&
		!firstTime.IsZero() && !secondTime.IsZero() && absDuration(firstTime.Sub(secondTime)) <= curveBoundaryJoinTime
}

func riseSetCurveEndpointsMatch(first, second []basic.OccultationPathPoint) bool {
	if len(first) < 2 || len(second) < 2 {
		return false
	}
	return riseSetCurveEndpointMatch(first[0], second[0]) ||
		riseSetCurveEndpointMatch(first[0], second[len(second)-1]) ||
		riseSetCurveEndpointMatch(first[len(first)-1], second[0]) ||
		riseSetCurveEndpointMatch(first[len(first)-1], second[len(second)-1])
}

func riseSetCurveEndpointMatch(first, second basic.OccultationPathPoint) bool {
	return DistanceKM(first, second) <= curveBoundaryJoinDistanceKM &&
		!first.Time.IsZero() && !second.Time.IsZero() &&
		absDuration(first.Time.Sub(second.Time)) <= curveBoundaryJoinTime
}

func joinRiseSetCurveSegments(
	first, second []basic.OccultationPathPoint,
) []basic.OccultationPathPoint {
	if riseSetCurveEndpointMatch(first[len(first)-1], second[0]) {
		return append(first, second[1:]...)
	}
	if riseSetCurveEndpointMatch(first[len(first)-1], second[len(second)-1]) {
		reverseOccultationPathPoints(second)
		return append(first, second[1:]...)
	}
	if riseSetCurveEndpointMatch(first[0], second[len(second)-1]) {
		return append(second[:len(second)-1], first...)
	}
	reverseOccultationPathPoints(second)
	return append(second[:len(second)-1], first...)
}

func reverseOccultationPathPoints(points []basic.OccultationPathPoint) {
	for left, right := 0, len(points)-1; left < right; left, right = left+1, right-1 {
		points[left], points[right] = points[right], points[left]
	}
}

func joinCurveBoundarySegments(
	first []geodata.GeoPoint,
	firstTimes []time.Time,
	second []geodata.GeoPoint,
	secondTimes []time.Time,
) ([]geodata.GeoPoint, []time.Time) {
	if curveBoundaryEndpointMatch(first[len(first)-1], firstTimes[len(first)-1], second[0], secondTimes[0]) {
		return append(first, second[1:]...), append(firstTimes, secondTimes[1:]...)
	}
	if curveBoundaryEndpointMatch(first[len(first)-1], firstTimes[len(first)-1], second[len(second)-1], secondTimes[len(second)-1]) {
		for left, right := 0, len(second)-1; left < right; left, right = left+1, right-1 {
			second[left], second[right] = second[right], second[left]
			secondTimes[left], secondTimes[right] = secondTimes[right], secondTimes[left]
		}
		return append(first, second[1:]...), append(firstTimes, secondTimes[1:]...)
	}
	if curveBoundaryEndpointMatch(first[0], firstTimes[0], second[len(second)-1], secondTimes[len(second)-1]) {
		return append(second[:len(second)-1], first...), append(secondTimes[:len(secondTimes)-1], firstTimes...)
	}
	for left, right := 0, len(second)-1; left < right; left, right = left+1, right-1 {
		second[left], second[right] = second[right], second[left]
		secondTimes[left], secondTimes[right] = secondTimes[right], secondTimes[left]
	}
	return append(second[:len(second)-1], first...), append(secondTimes[:len(secondTimes)-1], firstTimes...)
}

func absDuration(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}
	return value
}

func densifyOccultationPolygons(
	polygons [][]geodata.GeoPoint,
	maximumEdgeKM float64,
) [][]geodata.GeoPoint {
	if maximumEdgeKM <= 0 {
		return polygons
	}
	result := make([][]geodata.GeoPoint, len(polygons))
	for polygonIndex, polygon := range polygons {
		if len(polygon) < 2 {
			result[polygonIndex] = polygon
			continue
		}
		ring := make([]geodata.GeoPoint, 0, len(polygon)*2)
		for index, point := range polygon {
			ring = append(ring, point)
			if index+1 >= len(polygon) {
				continue
			}
			next := polygon[index+1]
			distance := occultationProjectedEdgeDistanceKM(point, next)
			steps := int(math.Ceil(distance / maximumEdgeKM))
			if steps < 2 {
				continue
			}
			for step := 1; step < steps; step++ {
				ring = append(ring, interpolateOccultationGeoPoint(point, next, float64(step)/float64(steps)))
			}
		}
		// GeoJSON closes polygon rings by connecting the final point back to
		// the first. Densify that implicit edge as well, otherwise a smooth
		// fallback can still render one long closing chord.
		if len(polygon) > 2 && !geodata.SameGeoPoint(polygon[0], polygon[len(polygon)-1]) {
			first, last := polygon[0], polygon[len(polygon)-1]
			distance := occultationProjectedEdgeDistanceKM(last, first)
			steps := int(math.Ceil(distance / maximumEdgeKM))
			if steps >= 2 {
				for step := 1; step < steps; step++ {
					ring = append(ring, interpolateOccultationGeoPoint(last, first, float64(step)/float64(steps)))
				}
			}
		}
		result[polygonIndex] = ring
	}
	return result
}

// occultationProjectedEdgeDistanceKM measures an edge in the Web Mercator
// chart used by the GeoJSON/OpenLayers consumer, while retaining the shortest
// wrapped longitude. At high latitude a small spherical edge expands strongly
// in this chart; using only the great-circle distance leaves visible polygon
// chords even though the source contact contour is densely sampled.
func occultationProjectedEdgeDistanceKM(first, second geodata.GeoPoint) float64 {
	const maxLatitude = 85.05112878
	clampLatitude := func(value float64) float64 {
		return math.Max(-maxLatitude, math.Min(maxLatitude, value))
	}
	longitude := math.Remainder(second.Longitude-first.Longitude, 360) * math.Pi / 180
	firstLatitude := clampLatitude(first.Latitude) * math.Pi / 180
	secondLatitude := clampLatitude(second.Latitude) * math.Pi / 180
	firstY := math.Log(math.Tan(math.Pi/4 + firstLatitude/2))
	secondY := math.Log(math.Tan(math.Pi/4 + secondLatitude/2))
	return EarthRadiusKM * math.Hypot(longitude, secondY-firstY)
}

func interpolateOccultationGeoPoint(
	first, second geodata.GeoPoint,
	fraction float64,
) geodata.GeoPoint {
	firstLongitude, firstLatitude := first.Longitude*math.Pi/180, first.Latitude*math.Pi/180
	secondLongitude, secondLatitude := second.Longitude*math.Pi/180, second.Latitude*math.Pi/180
	firstCos := math.Cos(firstLatitude)
	secondCos := math.Cos(secondLatitude)
	firstVector := [3]float64{firstCos * math.Cos(firstLongitude), firstCos * math.Sin(firstLongitude), math.Sin(firstLatitude)}
	secondVector := [3]float64{secondCos * math.Cos(secondLongitude), secondCos * math.Sin(secondLongitude), math.Sin(secondLatitude)}
	dot := firstVector[0]*secondVector[0] + firstVector[1]*secondVector[1] + firstVector[2]*secondVector[2]
	dot = math.Max(-1, math.Min(1, dot))
	angle := math.Acos(dot)
	var vector [3]float64
	if angle < 1e-12 {
		vector = [3]float64{
			(1-fraction)*firstVector[0] + fraction*secondVector[0],
			(1-fraction)*firstVector[1] + fraction*secondVector[1],
			(1-fraction)*firstVector[2] + fraction*secondVector[2],
		}
	} else {
		firstWeight := math.Sin((1-fraction)*angle) / math.Sin(angle)
		secondWeight := math.Sin(fraction*angle) / math.Sin(angle)
		vector = [3]float64{
			firstWeight*firstVector[0] + secondWeight*secondVector[0],
			firstWeight*firstVector[1] + secondWeight*secondVector[1],
			firstWeight*firstVector[2] + secondWeight*secondVector[2],
		}
	}
	length := math.Sqrt(vector[0]*vector[0] + vector[1]*vector[1] + vector[2]*vector[2])
	return geodata.GeoPoint{
		Longitude: math.Atan2(vector[1], vector[0]) * 180 / math.Pi,
		Latitude:  math.Asin(math.Max(-1, math.Min(1, vector[2]/length))) * 180 / math.Pi,
	}
}

func occultationPolygonNeedsInteriorProbes(
	polygon []basic.OccultationPathPoint,
) bool {
	for _, point := range polygon {
		if math.Abs(point.Latitude) >= 70 {
			return true
		}
	}
	return false
}

// clipOccultationPolygonToHorizon removes the below-horizon part of an
// instantaneous contact footprint. Footprint construction retains the full
// contact-cone arc so that its horizon closure can be swept continuously;
// that arc is not itself a visible area. Linear interpolation is sufficient
// at the sub-degree horizon crossing because the source samples are already
// spatially dense and the resulting linework is subsequently projected.
func clipOccultationPolygonToHorizon(
	source []basic.OccultationPathPoint,
) []basic.OccultationPathPoint {
	if len(source) < 3 {
		return nil
	}
	points := source
	if sameOccultationPoint(points[0], points[len(points)-1]) {
		points = points[:len(points)-1]
	}
	if len(points) < 3 {
		return nil
	}
	inside := func(point basic.OccultationPathPoint) bool {
		return finiteGeo(point.MoonAltitude) && point.MoonAltitude >= -1e-9
	}
	result := make([]basic.OccultationPathPoint, 0, len(points)+2)
	previous := points[len(points)-1]
	previousInside := inside(previous)
	for _, current := range points {
		currentInside := inside(current)
		if currentInside != previousInside {
			result = append(result, interpolateOccultationHorizonPoint(previous, current))
		}
		if currentInside {
			result = append(result, current)
		}
		previous, previousInside = current, currentInside
	}
	if len(result) < 3 {
		return nil
	}
	return result
}

func interpolateOccultationHorizonPoint(
	first, second basic.OccultationPathPoint,
) basic.OccultationPathPoint {
	denominator := first.MoonAltitude - second.MoonAltitude
	fraction := 0.5
	if math.Abs(denominator) > 1e-12 {
		fraction = first.MoonAltitude / denominator
	}
	fraction = math.Max(0, math.Min(1, fraction))
	deltaLongitude := second.Longitude - first.Longitude
	for deltaLongitude > 180 {
		deltaLongitude -= 360
	}
	for deltaLongitude < -180 {
		deltaLongitude += 360
	}
	point := first
	point.Time = first.Time.Add(time.Duration(float64(second.Time.Sub(first.Time)) * fraction))
	point.Longitude = normalizeGeoLongitude(first.Longitude + fraction*deltaLongitude)
	point.Latitude = first.Latitude + fraction*(second.Latitude-first.Latitude)
	point.MoonAltitude = 0
	point.WidthKM = first.WidthKM + fraction*(second.WidthKM-first.WidthKM)
	return point
}

func occultationVisibleFootprintProbes(
	source []basic.OccultationPathPoint,
	polygon []geodata.GeoPoint,
) []geodata.GeoPoint {
	index := geodata.NewSphericalPolygonIndex([][]geodata.GeoPoint{polygon})
	probes := occultationPolarFootprintProbes(source, polygon)
	if len(source) < 3 || len(source) != len(polygon) {
		return probes
	}
	x, y, z := 0.0, 0.0, 0.0
	for _, point := range polygon {
		longitude, latitude := point.Longitude*math.Pi/180, point.Latitude*math.Pi/180
		cosLatitude := math.Cos(latitude)
		x += cosLatitude * math.Cos(longitude)
		y += cosLatitude * math.Sin(longitude)
		z += math.Sin(latitude)
	}
	centerLength := math.Sqrt(x*x + y*y + z*z)
	if centerLength <= 1e-12 {
		return probes
	}
	center := [3]float64{x / centerLength, y / centerLength, z / centerLength}
	centerProbe := geodata.GeoPoint{
		Longitude: math.Atan2(center[1], center[0]) * 180 / math.Pi,
		Latitude:  math.Asin(math.Max(-1, math.Min(1, center[2]))) * 180 / math.Pi,
	}
	if index.ContainsPoints([]geodata.GeoPoint{centerProbe})[0] {
		probes = append(probes, centerProbe)
	}
	const maximumVertices = 8
	step := (len(polygon) + maximumVertices - 1) / maximumVertices
	type candidateRange struct{ start, end int }
	vertexCandidates := make([]geodata.GeoPoint, 0, 2*maximumVertices)
	vertexRanges := make([]candidateRange, 0, maximumVertices)
	for index := 0; index < len(polygon); index += step {
		start := len(vertexCandidates)
		point := polygon[index]
		longitude, latitude := point.Longitude*math.Pi/180, point.Latitude*math.Pi/180
		cosLatitude := math.Cos(latitude)
		vertex := [3]float64{cosLatitude * math.Cos(longitude), cosLatitude * math.Sin(longitude), math.Sin(latitude)}
		for _, inward := range []float64{0.0005, 0.005, 0.02, 0.05} {
			candidateVector := [3]float64{
				(1-inward)*vertex[0] + inward*center[0],
				(1-inward)*vertex[1] + inward*center[1],
				(1-inward)*vertex[2] + inward*center[2],
			}
			length := math.Sqrt(candidateVector[0]*candidateVector[0] + candidateVector[1]*candidateVector[1] + candidateVector[2]*candidateVector[2])
			vertexCandidates = append(vertexCandidates, geodata.GeoPoint{
				Longitude: math.Atan2(candidateVector[1], candidateVector[0]) * 180 / math.Pi,
				Latitude:  math.Asin(math.Max(-1, math.Min(1, candidateVector[2]/length))) * 180 / math.Pi,
			})
		}
		vertexRanges = append(vertexRanges, candidateRange{start: start, end: len(vertexCandidates)})
	}
	vertexInside := index.ContainsPoints(vertexCandidates)
	for _, value := range vertexRanges {
		for index := value.start; index < value.end; index++ {
			if vertexInside[index] {
				probes = append(probes, vertexCandidates[index])
				break
			}
		}
	}
	// A polar visible sliver can be much narrower than the vector from a
	// vertex to the polygon centroid. Probe both sides of sampled edges as
	// well, retaining only points that are actually inside the clipped source.
	edgeCandidates := make([]geodata.GeoPoint, 0, 4*maximumVertices)
	for index := 0; index < len(polygon); index += step {
		first := polygon[index]
		second := polygon[(index+1)%len(polygon)]
		deltaLongitude := second.Longitude - first.Longitude
		for deltaLongitude > 180 {
			deltaLongitude -= 360
		}
		for deltaLongitude < -180 {
			deltaLongitude += 360
		}
		deltaLatitude := second.Latitude - first.Latitude
		length := math.Hypot(deltaLongitude, deltaLatitude)
		if length <= 1e-12 {
			continue
		}
		midpoint := geodata.GeoPoint{
			Longitude: normalizeGeoLongitude(first.Longitude + deltaLongitude/2),
			Latitude:  first.Latitude + deltaLatitude/2,
		}
		for _, offset := range []float64{0.0002, 0.002, 0.01, 0.03} {
			for _, side := range []float64{-1, 1} {
				edgeCandidates = append(edgeCandidates, geodata.GeoPoint{
					Longitude: normalizeGeoLongitude(midpoint.Longitude - side*deltaLatitude*offset/length),
					Latitude:  midpoint.Latitude + side*deltaLongitude*offset/length,
				})
			}
		}
	}
	edgeInside := index.ContainsPoints(edgeCandidates)
	for index, candidate := range edgeCandidates {
		if edgeInside[index] {
			probes = append(probes, candidate)
		}
	}
	return probes
}

func limitOccultationCoveragePaths(
	paths [][]geodata.GeoPoint,
	maximum int,
) [][]geodata.GeoPoint {
	if maximum < 1 || len(paths) <= maximum {
		return paths
	}
	type indexedPoint struct {
		index    int
		latitude float64
	}
	ordered := make([]indexedPoint, len(paths))
	for index, path := range paths {
		latitude := 0.0
		if len(path) > 0 {
			latitude = path[0].Latitude
		}
		ordered[index] = indexedPoint{index: index, latitude: latitude}
	}
	sort.SliceStable(ordered, func(first, second int) bool {
		return ordered[first].latitude > ordered[second].latitude
	})
	keep := make([]bool, len(paths))
	reserve := maximum / 8
	if reserve < 1 {
		reserve = 1
	}
	for index := 0; index < reserve && index < len(ordered); index++ {
		keep[ordered[index].index] = true
	}
	for index := 0; index < reserve && index < len(ordered); index++ {
		keep[ordered[len(ordered)-1-index].index] = true
	}
	remaining := maximum
	for _, value := range keep {
		if value {
			remaining--
		}
	}
	if remaining < 0 {
		remaining = 0
	}
	stride := float64(len(paths)) / float64(remaining)
	result := make([][]geodata.GeoPoint, 0, maximum)
	for index, path := range paths {
		if keep[index] {
			result = append(result, path)
		}
	}
	if remaining > 0 {
		for cursor := 0.0; len(result) < maximum && int(cursor) < len(paths); cursor += stride {
			index := int(cursor)
			if keep[index] {
				continue
			}
			keep[index] = true
			result = append(result, paths[index])
		}
	}
	return result
}

func sameOccultationPoint(first, second basic.OccultationPathPoint) bool {
	return math.Abs(first.Latitude-second.Latitude) <= 1e-10 &&
		math.Abs(normalizeGeoLongitude(first.Longitude-second.Longitude)) <= 1e-10
}

func normalizeGeoLongitude(value float64) float64 {
	value = math.Mod(value+180, 360)
	if value < 0 {
		value += 360
	}
	return value - 180
}

func occultationPolarFootprintProbes(
	source []basic.OccultationPathPoint,
	polygon []geodata.GeoPoint,
) []geodata.GeoPoint {
	if len(source) < 3 || len(source) != len(polygon) {
		return nil
	}
	x, y, z := 0.0, 0.0, 0.0
	polarIndices := make([]int, 0, len(source))
	for index, point := range source {
		longitude, latitude := point.Longitude*math.Pi/180, point.Latitude*math.Pi/180
		cosLatitude := math.Cos(latitude)
		x += cosLatitude * math.Cos(longitude)
		y += cosLatitude * math.Sin(longitude)
		z += math.Sin(latitude)
		if math.Abs(point.Latitude) >= 70 {
			polarIndices = append(polarIndices, index)
		}
	}
	if len(polarIndices) == 0 {
		return nil
	}
	centerLength := math.Sqrt(x*x + y*y + z*z)
	if centerLength <= 1e-12 {
		return nil
	}
	const maximumProbes = 16
	step := (len(polarIndices) + maximumProbes - 1) / maximumProbes
	candidates := make([]geodata.GeoPoint, 0, maximumProbes)
	for position := 0; position < len(polarIndices); position += step {
		point := source[polarIndices[position]]
		longitude, latitude := point.Longitude*math.Pi/180, point.Latitude*math.Pi/180
		cosLatitude := math.Cos(latitude)
		const inward = 0.002
		candidateVector := [3]float64{
			(1-inward)*cosLatitude*math.Cos(longitude) + inward*x/centerLength,
			(1-inward)*cosLatitude*math.Sin(longitude) + inward*y/centerLength,
			(1-inward)*math.Sin(latitude) + inward*z/centerLength,
		}
		candidateLength := math.Sqrt(
			candidateVector[0]*candidateVector[0] +
				candidateVector[1]*candidateVector[1] +
				candidateVector[2]*candidateVector[2],
		)
		candidates = append(candidates, geodata.GeoPoint{
			Longitude: math.Atan2(candidateVector[1], candidateVector[0]) * 180 / math.Pi,
			Latitude:  math.Asin(math.Max(-1, math.Min(1, candidateVector[2]/candidateLength))) * 180 / math.Pi,
		})
	}
	inside := geodata.SphericalPolygonsContainPoints([][]geodata.GeoPoint{polygon}, candidates)
	probes := make([]geodata.GeoPoint, 0, len(candidates))
	for index, candidate := range candidates {
		if inside[index] {
			probes = append(probes, candidate)
		}
	}
	return probes
}

func removeOccultationHairpins(points []geodata.GeoPoint, maximumClosureKM, minimumDetourKM float64, maximumSpan int) []geodata.GeoPoint {
	if len(points) < 5 || maximumSpan < 3 {
		return points
	}
	result := append([]geodata.GeoPoint(nil), points...)
	for pass := 0; pass < 8; pass++ {
		changed := false
		for start := 0; start+3 < len(result); start++ {
			limit := start + maximumSpan
			if limit >= len(result) {
				limit = len(result) - 1
			}
			arcLength := 0.0
			best := -1
			for end := start + 1; end <= limit; end++ {
				arcLength += geoDistanceKM(result[end-1], result[end])
				if end < start+3 {
					continue
				}
				closure := geoDistanceKM(result[start], result[end])
				if closure <= maximumClosureKM && arcLength-closure >= minimumDetourKM {
					best = end
				}
			}
			if best > start+1 {
				result = append(result[:start+1], result[best:]...)
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	return result
}

// smoothOccultationHairpins replaces a numerical return with a short,
// densified great-circle chord. Deleting the return vertices outright can
// leave a long straight edge in a rendered fallback band, so the replacement
// preserves display sampling while removing only the detected detour.
func smoothOccultationHairpins(
	points []geodata.GeoPoint,
	maximumClosureKM, minimumDetourKM float64,
	maximumSpan int,
	maximumEdgeKM float64,
) []geodata.GeoPoint {
	if len(points) < 5 || maximumSpan < 3 || maximumEdgeKM <= 0 {
		return points
	}
	result := append([]geodata.GeoPoint(nil), points...)
	for pass := 0; pass < 8; pass++ {
		changed := false
		for start := 0; start+3 < len(result); start++ {
			limit := start + maximumSpan
			if limit >= len(result) {
				limit = len(result) - 1
			}
			arcLength := 0.0
			best := -1
			bestDetour := minimumDetourKM
			for end := start + 1; end <= limit; end++ {
				arcLength += geoDistanceKM(result[end-1], result[end])
				if end < start+3 {
					continue
				}
				closure := geoDistanceKM(result[start], result[end])
				detour := arcLength - closure
				if closure <= maximumClosureKM && detour >= bestDetour {
					best = end
					bestDetour = detour
				}
			}
			if best <= start+1 {
				continue
			}
			first, last := result[start], result[best]
			steps := int(math.Ceil(geoDistanceKM(first, last) / maximumEdgeKM))
			if steps < 1 {
				steps = 1
			}
			replacement := make([]geodata.GeoPoint, 0, steps)
			for step := 1; step < steps; step++ {
				replacement = append(replacement,
					interpolateOccultationGeoPoint(first, last, float64(step)/float64(steps)))
			}
			next := make([]geodata.GeoPoint, 0, len(result)-best+start+1+len(replacement))
			next = append(next, result[:start+1]...)
			next = append(next, replacement...)
			next = append(next, result[best:]...)
			result = next
			changed = true
			break
		}
		if !changed {
			break
		}
	}
	return result
}

// removeOccultationSharpCorners drops tiny numerical backtracks left at a
// shared polar fold. These are not physical phase vertices: the neighbouring
// points are only a few kilometres apart while the rendered edge reverses
// direction, which produces a visible corner in a filled map polygon.
func removeOccultationSharpCorners(
	points []geodata.GeoPoint,
	maximumChordKM, minimumAngleDegrees float64,
) []geodata.GeoPoint {
	if len(points) < 4 || maximumChordKM <= 0 || minimumAngleDegrees <= 0 {
		return points
	}
	result := append([]geodata.GeoPoint(nil), points...)
	for pass := 0; pass < 128; pass++ {
		changed := false
		for index := 1; index+1 < len(result); index++ {
			first, middle, last := result[index-1], result[index], result[index+1]
			if geoDistanceKM(first, last) > maximumChordKM {
				continue
			}
			if occultationTurnAngleDegrees(first, middle, last) >= minimumAngleDegrees {
				continue
			}
			result = append(result[:index], result[index+1:]...)
			changed = true
			break
		}
		if !changed {
			break
		}
	}
	return result
}

// removeOccultationPolarSharpCorners removes a short, high-latitude change of
// direction even when both coordinate axes remain monotonic. Such vertices
// are not reversals, but are the polygonizer's seam between two sampled polar
// branches and render as a visible notch in Web Mercator.
func removeOccultationPolarSharpCorners(
	points []geodata.GeoPoint,
	maximumChordKM, minimumAngleDegrees, minimumLatitudeDegrees float64,
) []geodata.GeoPoint {
	if len(points) < 4 || maximumChordKM <= 0 || minimumAngleDegrees <= 0 {
		return points
	}
	result := append([]geodata.GeoPoint(nil), points...)
	for pass := 0; pass < 128; pass++ {
		changed := false
		for index := 1; index+1 < len(result); index++ {
			first, middle, last := result[index-1], result[index], result[index+1]
			if math.Abs(middle.Latitude) < minimumLatitudeDegrees ||
				geoDistanceKM(first, last) > maximumChordKM ||
				occultationTurnAngleDegrees(first, middle, last) >= minimumAngleDegrees {
				continue
			}
			result = append(result[:index], result[index+1:]...)
			changed = true
			break
		}
		if !changed {
			break
		}
	}
	return result
}

// smoothOccultationPolarCorners replaces a short polar backtrack whose direct
// endpoint chord is too long for deletion. Interpolating that chord keeps the
// rendered spacing bounded while removing the sampled branch's angular seam.
func smoothOccultationPolarCorners(points []geodata.GeoPoint) []geodata.GeoPoint {
	if len(points) < 4 {
		return points
	}
	result := append([]geodata.GeoPoint(nil), points...)
	for pass := 0; pass < 128; pass++ {
		changed := false
		for index := 1; index+1 < len(result); index++ {
			first, middle, last := result[index-1], result[index], result[index+1]
			if math.Abs(middle.Latitude) < 70 ||
				occultationProjectedEdgeDistanceKM(first, last) > 140 ||
				occultationTurnAngleDegrees(first, middle, last) >= 165 {
				continue
			}
			longitudeReversal := math.Remainder(middle.Longitude-first.Longitude, 360)*
				math.Remainder(last.Longitude-middle.Longitude, 360) < 0
			latitudeReversal := (middle.Latitude-first.Latitude)*(last.Latitude-middle.Latitude) < 0
			if !longitudeReversal && !latitudeReversal {
				continue
			}
			steps := int(math.Ceil(occultationProjectedEdgeDistanceKM(first, last) / 40))
			if steps < 2 {
				steps = 2
			}
			replacement := make([]geodata.GeoPoint, 0, steps-1)
			for step := 1; step < steps; step++ {
				replacement = append(replacement,
					interpolateOccultationGeoPoint(first, last, float64(step)/float64(steps)))
			}
			next := make([]geodata.GeoPoint, 0, len(result)-1+len(replacement))
			next = append(next, result[:index]...)
			next = append(next, replacement...)
			next = append(next, result[index+1:]...)
			result = next
			changed = true
			break
		}
		if !changed {
			break
		}
	}
	return result
}

// smoothOccultationPolarWobbles removes a short latitude oscillation on an
// otherwise monotone polar edge. Polygonizer seams can alternate north/south
// by a few kilometres while spanning a longer longitude interval, which is
// too wide for the point-deletion cleaners above. The replacement is bounded
// to a four-point window, a 35 km deviation, and a 400 km endpoint span; real
// branch folds with a longitude reversal or larger curvature remain intact.
func smoothOccultationPolarWobbles(points []geodata.GeoPoint) []geodata.GeoPoint {
	const (
		minimumPolarLatitudeDegrees = 75.0
		maximumDeviationKM          = 35.0
		maximumEndpointSpanKM       = 400.0
		windowPoints                = 3
	)
	if len(points) < windowPoints+1 {
		return points
	}
	result := append([]geodata.GeoPoint(nil), points...)
	closed := len(result) > 1 && geodata.SameGeoPoint(result[0], result[len(result)-1])
	limit := len(result)
	if closed {
		limit--
	}
	for pass := 0; pass < 8; pass++ {
		changed := false
		for start := 0; start+windowPoints < limit; start++ {
			end := start + windowPoints
			first, last := result[start], result[end]
			if math.Abs(first.Latitude) < minimumPolarLatitudeDegrees ||
				math.Abs(last.Latitude) < minimumPolarLatitudeDegrees ||
				occultationProjectedEdgeDistanceKM(first, last) > maximumEndpointSpanKM {
				continue
			}
			longitudeDirection := 0.0
			latitudeReversal := false
			previousLatitudeDelta := 0.0
			valid := true
			for index := start + 1; index <= end; index++ {
				point := result[index]
				if math.Abs(point.Latitude) < minimumPolarLatitudeDegrees {
					valid = false
					break
				}
				longitudeDelta := math.Remainder(point.Longitude-result[index-1].Longitude, 360)
				if math.Abs(longitudeDelta) <= 1e-7 {
					valid = false
					break
				}
				if longitudeDirection == 0 {
					longitudeDirection = math.Copysign(1, longitudeDelta)
				} else if longitudeDelta*longitudeDirection <= 0 {
					valid = false
					break
				}
				latitudeDelta := point.Latitude - result[index-1].Latitude
				if previousLatitudeDelta != 0 && latitudeDelta*previousLatitudeDelta < 0 {
					latitudeReversal = true
				}
				if latitudeDelta != 0 {
					previousLatitudeDelta = latitudeDelta
				}
			}
			if !valid || !latitudeReversal {
				continue
			}
			for index := start + 1; index < end; index++ {
				fraction := float64(index-start) / float64(windowPoints)
				baseline := interpolateOccultationGeoPoint(first, last, fraction)
				if geoDistanceKM(result[index], baseline) > maximumDeviationKM {
					valid = false
					break
				}
			}
			if !valid {
				continue
			}
			for index := start + 1; index < end; index++ {
				fraction := float64(index-start) / float64(windowPoints)
				result[index] = interpolateOccultationGeoPoint(first, last, fraction)
			}
			changed = true
			start = end - 1
		}
		if !changed {
			break
		}
	}
	if closed && len(result) > 1 {
		result[len(result)-1] = result[0]
	}
	return result
}

func occultationTurnAngleDegrees(
	first, middle, last geodata.GeoPoint,
) float64 {
	latitude := middle.Latitude * math.Pi / 180
	scale := math.Cos(latitude)
	firstX := math.Remainder(first.Longitude-middle.Longitude, 360) * scale
	firstY := first.Latitude - middle.Latitude
	lastX := math.Remainder(last.Longitude-middle.Longitude, 360) * scale
	lastY := last.Latitude - middle.Latitude
	firstLength := math.Hypot(firstX, firstY)
	lastLength := math.Hypot(lastX, lastY)
	if firstLength <= 1e-12 || lastLength <= 1e-12 {
		return 180
	}
	cosine := (firstX*lastX + firstY*lastY) / (firstLength * lastLength)
	cosine = math.Max(-1, math.Min(1, cosine))
	return math.Acos(cosine) * 180 / math.Pi
}
