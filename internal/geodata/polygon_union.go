package geodata

import (
	"fmt"
	"math"
	"sort"
)

const polygonUnionEpsilon = 1e-9
const polygonUnionProbeOffset = 1e-7
const polygonUnionSnapGrid = 1e-7
const polygonUnionNodeGrid = 1e-7
const polygonUnionContainmentToleranceKM = 1.0
const polygonUnionLocationBinCount = 64

type polygonUnionPoint struct {
	x float64
	y float64
}

type polygonUnionRing struct {
	points        []polygonUnionPoint
	edges         []polygonUnionSourceEdge
	locationEdges []polygonUnionSourceEdge
	locationBins  [][]polygonUnionSourceEdge
	params        [][]float64
	minX          float64
	maxX          float64
	minY          float64
	maxY          float64
}

type polygonUnionSourceEdge struct {
	start polygonUnionPoint
	end   polygonUnionPoint
	minX  float64
	maxX  float64
	minY  float64
	maxY  float64
}

type polygonUnionSourceEdgeRef struct {
	ringIndex int
	edgeIndex int
	edge      polygonUnionSourceEdge
}

type polygonUnionEdge struct {
	start polygonUnionPoint
	end   polygonUnionPoint
}

type polygonUnionNode struct {
	x int64
	y int64
}

type polygonUnionEdgeKey struct {
	start polygonUnionNode
	end   polygonUnionNode
}

// UnionPolygons 合并相互重叠的地理多边形环，不相交的输入保持分离。
// UnionPolygons merges overlapping geographic polygon rings. Ordinary small
// regions use the fast equirectangular branch; polar or antimeridian regions
// are rotated to a local spherical chart first, so the planar edge splitter
// never sees a coordinate singularity. Disjoint inputs remain disjoint.
func UnionPolygons(polygons [][]GeoPoint) ([][]GeoPoint, error) {
	if polygonUnionNeedsSphericalChart(polygons) {
		// Preserve the numerically stable legacy result when it already retains
		// every input boundary and pole cap. The rotated chart is a recovery
		// path for the singular cases; applying it to an ordinary antimeridian
		// band can unnecessarily change the displayed equirectangular chord.
		// 绕极环在平面图幅里无法闭合，平面并集必然失败，不必先算一遍。
		if !polygonUnionWindsAroundPole(polygons) {
			if planar, err := unionPolygonsPlanar(polygons); err == nil &&
				polygonUnionContainsInputs(planar, polygons) {
				return planar, nil
			}
		}
		result, err := unionPolygonsSphericalChart(polygons)
		if err == nil {
			return result, nil
		}
		if snapped := snapPolygonUnionInputs(polygons); snapped != nil {
			if result, retryErr := unionPolygonsSphericalChart(snapped); retryErr == nil {
				return result, nil
			}
		}
		return nil, err
	}
	result, err := unionPolygonsPlanar(polygons)
	if err == nil {
		return result, nil
	}
	if snapped := snapPolygonUnionInputs(polygons); snapped != nil {
		if result, retryErr := unionPolygonsPlanar(snapped); retryErr == nil {
			return result, nil
		}
	}
	return nil, err
}

// snapPolygonUnionInputs retries a failed boolean join after quantizing input
// vertices to a centimetre-scale angular grid. Intersections generated from
// adjacent temporal footprints can differ by a few nanodegrees; the primary
// union keeps full precision, while this bounded retry only closes that
// numerical seam when the exact join cannot form a ring.
func snapPolygonUnionInputs(polygons [][]GeoPoint) [][]GeoPoint {
	if len(polygons) == 0 {
		return nil
	}
	result := make([][]GeoPoint, len(polygons))
	changed := false
	for polygonIndex, polygon := range polygons {
		if len(polygon) < 3 {
			return nil
		}
		result[polygonIndex] = make([]GeoPoint, len(polygon))
		for pointIndex, point := range polygon {
			longitude := math.Round(point.Longitude/polygonUnionSnapGrid) * polygonUnionSnapGrid
			latitude := math.Round(point.Latitude/polygonUnionSnapGrid) * polygonUnionSnapGrid
			result[polygonIndex][pointIndex] = GeoPoint{Longitude: longitude, Latitude: latitude}
			changed = changed || longitude != point.Longitude || latitude != point.Latitude
		}
	}
	if !changed {
		return nil
	}
	return result
}

func polygonUnionContainsInputs(result, inputs [][]GeoPoint) bool {
	if !SphericalPolygonsContainPathsWithinKM(
		result, inputs, true, polygonUnionContainmentToleranceKM,
	) {
		return false
	}
	for _, input := range inputs {
		open := openGeoRing(input)
		if !polygonUnionRingWindsAroundPole(open) {
			continue
		}
		for _, pole := range []GeoPoint{{Longitude: 0, Latitude: 90}, {Longitude: 0, Latitude: -90}} {
			if sphericalPolygonContainsOrTouches(open, pole) &&
				!SphericalPolygonsContainPaths(result, [][]GeoPoint{{pole}}, false) {
				return false
			}
		}
	}
	return true
}

// polygonUnionWindsAroundPole 报告是否有环绕极点一圈。
func polygonUnionWindsAroundPole(polygons [][]GeoPoint) bool {
	for _, polygon := range polygons {
		if polygonUnionRingWindsAroundPole(openGeoRing(polygon)) {
			return true
		}
	}
	return false
}

// polygonUnionRingWindsAroundPole 用与 polygonUnionRings 相同的展开闭合判定识别绕极环。
func polygonUnionRingWindsAroundPole(ring []GeoPoint) bool {
	if len(ring) < 3 {
		return false
	}
	last := ring[0].Longitude
	for _, point := range ring[1:] {
		last += math.Remainder(point.Longitude-last, 360)
	}
	closure := last + math.Remainder(ring[0].Longitude-last, 360)
	return math.Abs(closure-ring[0].Longitude) > 180
}

func unionPolygonsPlanar(polygons [][]GeoPoint) ([][]GeoPoint, error) {
	rings, err := polygonUnionRings(polygons)
	if err != nil {
		return nil, err
	}
	polygonUnionAddIntersections(rings)
	edges := polygonUnionOuterEdges(rings)
	if len(edges) == 0 {
		return nil, fmt.Errorf("polygon union has no outer edges")
	}
	result, err := polygonUnionJoinEdges(edges)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("polygon union has no usable rings")
	}
	return result, nil
}

type polygonUnionChart struct {
	xAxis geoVector3
	yAxis geoVector3
	zAxis geoVector3
}

func polygonUnionNeedsSphericalChart(polygons [][]GeoPoint) bool {
	for _, polygon := range polygons {
		if len(polygon) < 3 {
			continue
		}
		open := openGeoRing(polygon)
		for index, point := range open {
			if math.Abs(point.Latitude) >= 70 {
				return true
			}
			next := open[(index+1)%len(open)]
			if math.Abs(next.Longitude-point.Longitude) > 180 {
				return true
			}
		}
	}
	return false
}

func unionPolygonsSphericalChart(polygons [][]GeoPoint) ([][]GeoPoint, error) {
	chart, ok := newPolygonUnionChart(polygons)
	if !ok {
		return unionPolygonsPlanar(polygons)
	}
	projected := make([][]GeoPoint, len(polygons))
	for polygonIndex, polygon := range polygons {
		projected[polygonIndex] = make([]GeoPoint, len(polygon))
		for pointIndex, point := range polygon {
			projected[polygonIndex][pointIndex] = chart.project(point)
		}
	}
	merged, err := unionPolygonsPlanar(projected)
	if err != nil {
		return nil, err
	}
	result := make([][]GeoPoint, len(merged))
	for polygonIndex, polygon := range merged {
		result[polygonIndex] = make([]GeoPoint, len(polygon))
		for pointIndex, point := range polygon {
			result[polygonIndex][pointIndex] = chart.unproject(point)
		}
	}
	return result, nil
}

func newPolygonUnionChart(polygons [][]GeoPoint) (polygonUnionChart, bool) {
	center := geoVector3{}
	var first geoVector3
	haveFirst := false
	for _, polygon := range polygons {
		for _, point := range openGeoRing(polygon) {
			vector := geoPointVector(point)
			center = geoVectorAdd(center, vector)
			if !haveFirst {
				first = vector
				haveFirst = true
			}
		}
	}
	center, ok := geoVectorNormalize(center)
	if !ok {
		center, ok = geoVectorNormalize(first)
		if !ok {
			return polygonUnionChart{}, false
		}
	}
	globalNorth := geoVector3{z: 1}
	zAxis, ok := geoVectorNormalize(geoVectorAdd(globalNorth, geoVectorScale(center, -geoVectorDot(globalNorth, center))))
	if !ok {
		zAxis, ok = geoVectorNormalize(geoVectorAdd(geoVector3{x: 1}, geoVectorScale(center, -center.x)))
		if !ok {
			return polygonUnionChart{}, false
		}
	}
	yAxis, ok := geoVectorNormalize(geoVectorCross(zAxis, center))
	if !ok {
		return polygonUnionChart{}, false
	}
	return polygonUnionChart{xAxis: center, yAxis: yAxis, zAxis: zAxis}, true
}

func (chart polygonUnionChart) project(point GeoPoint) GeoPoint {
	vector := geoPointVector(point)
	return GeoPoint{
		Longitude: math.Atan2(geoVectorDot(vector, chart.yAxis), geoVectorDot(vector, chart.xAxis)) * 180 / math.Pi,
		Latitude:  math.Asin(math.Max(-1, math.Min(1, geoVectorDot(vector, chart.zAxis)))) * 180 / math.Pi,
	}
}

func (chart polygonUnionChart) unproject(point GeoPoint) GeoPoint {
	latitude := point.Latitude * math.Pi / 180
	longitude := point.Longitude * math.Pi / 180
	cosLatitude := math.Cos(latitude)
	vector := geoVectorAdd(
		geoVectorScale(chart.xAxis, cosLatitude*math.Cos(longitude)),
		geoVectorAdd(
			geoVectorScale(chart.yAxis, cosLatitude*math.Sin(longitude)),
			geoVectorScale(chart.zAxis, math.Sin(latitude)),
		),
	)
	return geoVectorPoint(vector)
}

func polygonUnionRings(polygons [][]GeoPoint) ([]polygonUnionRing, error) {
	rings := make([]polygonUnionRing, 0, len(polygons))
	reference := 0.0
	haveReference := false
	for polygonIndex, source := range polygons {
		source = openGeoRing(source)
		if len(source) < 3 {
			return nil, fmt.Errorf("polygon %d requires at least three points", polygonIndex)
		}
		points := make([]polygonUnionPoint, len(source))
		points[0] = polygonUnionPoint{x: source[0].Longitude, y: source[0].Latitude}
		for index := 1; index < len(source); index++ {
			points[index] = polygonUnionPoint{
				x: points[index-1].x + math.Remainder(source[index].Longitude-points[index-1].x, 360),
				y: source[index].Latitude,
			}
		}
		// A pole-winding ring closes one full longitude turn away. Joining it
		// directly in this chart creates an artificial chord through its interior;
		// let UnionPolygons retry in the rotated spherical chart instead.
		closure := points[len(points)-1].x + math.Remainder(points[0].x-points[len(points)-1].x, 360)
		if math.Abs(closure-points[0].x) > 180 {
			return nil, fmt.Errorf("polygon %d winds around a chart pole", polygonIndex)
		}
		mean := polygonUnionMeanLongitude(points)
		if !haveReference {
			reference = mean
			haveReference = true
		} else {
			shift := math.Round((reference-mean)/360) * 360
			for index := range points {
				points[index].x += shift
			}
		}
		ring := polygonUnionRingForPoints(points)
		ring.params = make([][]float64, len(points))
		for index := range ring.params {
			ring.params[index] = []float64{0, 1}
		}
		rings = append(rings, ring)
	}
	if len(rings) == 0 {
		return nil, fmt.Errorf("polygon union requires at least one polygon")
	}
	return rings, nil
}

func polygonUnionRingForPoints(points []polygonUnionPoint) polygonUnionRing {
	if len(points) == 0 {
		return polygonUnionRing{points: points}
	}
	edges := make([]polygonUnionSourceEdge, len(points))
	for index, start := range points {
		end := points[(index+1)%len(points)]
		edges[index] = polygonUnionSourceEdge{
			start: start, end: end,
			minX: math.Min(start.x, end.x), maxX: math.Max(start.x, end.x),
			minY: math.Min(start.y, end.y), maxY: math.Max(start.y, end.y),
		}
	}
	minX, maxX := points[0].x, points[0].x
	minY, maxY := points[0].y, points[0].y
	for _, point := range points[1:] {
		minX, maxX = math.Min(minX, point.x), math.Max(maxX, point.x)
		minY, maxY = math.Min(minY, point.y), math.Max(maxY, point.y)
	}
	locationEdges := append([]polygonUnionSourceEdge(nil), edges...)
	sort.Sort(polygonUnionEdgesByY(locationEdges))
	return polygonUnionRing{
		points: points, edges: edges, locationEdges: locationEdges,
		locationBins: polygonUnionBuildLocationBins(locationEdges, minY, maxY),
		minX:         minX, maxX: maxX, minY: minY, maxY: maxY,
	}
}

func polygonUnionAddIntersections(rings []polygonUnionRing) {
	edges := make([]polygonUnionSourceEdgeRef, 0)
	for ringIndex, ring := range rings {
		for edgeIndex, edge := range ring.edges {
			edges = append(edges, polygonUnionSourceEdgeRef{
				ringIndex: ringIndex,
				edgeIndex: edgeIndex,
				edge:      edge,
			})
		}
	}
	sort.Sort(polygonUnionEdgeRefsByX(edges))
	for firstIndex, first := range edges {
		for secondIndex := firstIndex + 1; secondIndex < len(edges); secondIndex++ {
			second := edges[secondIndex]
			if second.edge.minX > first.edge.maxX+polygonUnionEpsilon {
				break
			}
			if first.ringIndex == second.ringIndex && polygonUnionEdgesAdjacent(
				first.edgeIndex, second.edgeIndex, len(rings[first.ringIndex].points),
			) {
				continue
			}
			if !polygonUnionEdgeBoundsOverlap(first.edge, second.edge) {
				continue
			}
			firstParams, firstCount, secondParams, secondCount := polygonUnionSegmentIntersections(
				first.edge.start, first.edge.end, second.edge.start, second.edge.end,
			)
			firstRing := &rings[first.ringIndex]
			secondRing := &rings[second.ringIndex]
			firstRing.params[first.edgeIndex] = append(firstRing.params[first.edgeIndex], firstParams[:firstCount]...)
			secondRing.params[second.edgeIndex] = append(secondRing.params[second.edgeIndex], secondParams[:secondCount]...)
		}
	}
}

func polygonUnionEdgeBoundsOverlap(first, second polygonUnionSourceEdge) bool {
	return first.minX <= second.maxX+polygonUnionEpsilon && second.minX <= first.maxX+polygonUnionEpsilon &&
		first.minY <= second.maxY+polygonUnionEpsilon && second.minY <= first.maxY+polygonUnionEpsilon
}

func polygonUnionEdgesAdjacent(first, second, count int) bool {
	return (first+1)%count == second || (second+1)%count == first
}

func polygonUnionSegmentIntersections(
	firstStart, firstEnd, secondStart, secondEnd polygonUnionPoint,
) (firstParams [2]float64, firstCount int, secondParams [2]float64, secondCount int) {
	firstDelta := polygonUnionSubtract(firstEnd, firstStart)
	secondDelta := polygonUnionSubtract(secondEnd, secondStart)
	offset := polygonUnionSubtract(secondStart, firstStart)
	denominator := polygonUnionCross(firstDelta, secondDelta)
	if math.Abs(denominator) > polygonUnionEpsilon {
		firstParam := polygonUnionCross(offset, secondDelta) / denominator
		secondParam := polygonUnionCross(offset, firstDelta) / denominator
		if firstParam < -polygonUnionEpsilon || firstParam > 1+polygonUnionEpsilon ||
			secondParam < -polygonUnionEpsilon || secondParam > 1+polygonUnionEpsilon {
			return firstParams, 0, secondParams, 0
		}
		firstParams[0] = polygonUnionClampParam(firstParam)
		secondParams[0] = polygonUnionClampParam(secondParam)
		return firstParams, 1, secondParams, 1
	}
	if math.Abs(polygonUnionCross(offset, firstDelta)) > polygonUnionEpsilon {
		return firstParams, 0, secondParams, 0
	}

	for _, point := range [2]polygonUnionPoint{secondStart, secondEnd} {
		if value, ok := polygonUnionPointSegmentParam(point, firstStart, firstEnd); ok {
			firstParams[firstCount] = value
			firstCount++
		}
	}
	for _, point := range [2]polygonUnionPoint{firstStart, firstEnd} {
		if value, ok := polygonUnionPointSegmentParam(point, secondStart, secondEnd); ok {
			secondParams[secondCount] = value
			secondCount++
		}
	}
	return firstParams, firstCount, secondParams, secondCount
}

func polygonUnionPointSegmentParam(point, start, end polygonUnionPoint) (float64, bool) {
	delta := polygonUnionSubtract(end, start)
	lengthSquared := delta.x*delta.x + delta.y*delta.y
	if lengthSquared <= polygonUnionEpsilon*polygonUnionEpsilon {
		return 0, false
	}
	value := ((point.x-start.x)*delta.x + (point.y-start.y)*delta.y) / lengthSquared
	if value < -polygonUnionEpsilon || value > 1+polygonUnionEpsilon {
		return 0, false
	}
	projected := polygonUnionInterpolate(start, end, value)
	if math.Hypot(projected.x-point.x, projected.y-point.y) > polygonUnionEpsilon {
		return 0, false
	}
	return polygonUnionClampParam(value), true
}

func polygonUnionOuterEdges(rings []polygonUnionRing) []polygonUnionEdge {
	edges := make(map[polygonUnionEdgeKey]polygonUnionEdge)
	for ringIndex, ring := range rings {
		for edgeIndex, start := range ring.points {
			end := ring.points[(edgeIndex+1)%len(ring.points)]
			params := polygonUnionUniqueParams(ring.params[edgeIndex])
			for index := 1; index < len(params); index++ {
				firstParam, secondParam := params[index-1], params[index]
				if secondParam-firstParam <= polygonUnionEpsilon {
					continue
				}
				pieceStart := polygonUnionInterpolate(start, end, firstParam)
				pieceEnd := polygonUnionInterpolate(start, end, secondParam)
				delta := polygonUnionSubtract(pieceEnd, pieceStart)
				length := math.Hypot(delta.x, delta.y)
				if length <= polygonUnionEpsilon {
					continue
				}
				middle := polygonUnionInterpolate(pieceStart, pieceEnd, 0.5)
				normal := polygonUnionPoint{
					x: -delta.y / length * polygonUnionProbeOffset,
					y: delta.x / length * polygonUnionProbeOffset,
				}
				leftInside := polygonUnionInsideAnyRing(polygonUnionAdd(middle, normal), rings, ringIndex)
				rightInside := polygonUnionInsideAnyRing(polygonUnionSubtract(middle, normal), rings, ringIndex)
				if leftInside == rightInside {
					continue
				}
				if rightInside {
					pieceStart, pieceEnd = pieceEnd, pieceStart
				}
				edge := polygonUnionEdge{start: pieceStart, end: pieceEnd}
				key := polygonUnionKey(edge)
				edges[key] = edge
			}
		}
	}
	result := make([]polygonUnionEdge, 0, len(edges))
	for _, edge := range edges {
		result = append(result, edge)
	}
	sort.Slice(result, func(first, second int) bool {
		a, b := result[first], result[second]
		if a.start.x != b.start.x {
			return a.start.x < b.start.x
		}
		if a.start.y != b.start.y {
			return a.start.y < b.start.y
		}
		if a.end.x != b.end.x {
			return a.end.x < b.end.x
		}
		return a.end.y < b.end.y
	})
	return result
}

func polygonUnionInsideAnyRing(point polygonUnionPoint, rings []polygonUnionRing, preferred int) bool {
	if preferred >= 0 && preferred < len(rings) && polygonUnionRingRefContainsPoint(point, &rings[preferred]) {
		return true
	}
	for index := range rings {
		if index != preferred && polygonUnionRingRefContainsPoint(point, &rings[index]) {
			return true
		}
	}
	return false
}

func polygonUnionRingContainsPoint(point polygonUnionPoint, ring polygonUnionRing) bool {
	return polygonUnionRingRefContainsPoint(point, &ring)
}

func polygonUnionRingRefContainsPoint(point polygonUnionPoint, ring *polygonUnionRing) bool {
	margin := polygonUnionProbeOffset + polygonUnionEpsilon
	if point.x < ring.minX-margin || point.x > ring.maxX+margin ||
		point.y < ring.minY-margin || point.y > ring.maxY+margin {
		return false
	}
	return polygonUnionPointLocationInRingRef(point, ring) >= 0
}

func polygonUnionJoinEdges(edges []polygonUnionEdge) ([][]GeoPoint, error) {
	outgoing := make(map[polygonUnionNode][]int, len(edges))
	for index, edge := range edges {
		outgoing[polygonUnionNodeForPoint(edge.start)] = append(
			outgoing[polygonUnionNodeForPoint(edge.start)], index,
		)
	}
	used := make([]bool, len(edges))
	type resultRing struct {
		points []GeoPoint
		area   float64
	}
	results := make([]resultRing, 0, 2)
	// chartRings 保留图幅坐标，用于并集后检测孔洞。
	chartRings := make([][]polygonUnionPoint, 0, 2)
	for firstEdgeIndex := range edges {
		if used[firstEdgeIndex] {
			continue
		}
		firstNode := polygonUnionNodeForPoint(edges[firstEdgeIndex].start)
		currentEdgeIndex := firstEdgeIndex
		points := make([]polygonUnionPoint, 0, len(edges))
		for step := 0; step <= len(edges); step++ {
			if used[currentEdgeIndex] {
				return nil, fmt.Errorf("polygon union outer edges form a repeated branch")
			}
			current := edges[currentEdgeIndex]
			used[currentEdgeIndex] = true
			points = append(points, current.start)
			nextNode := polygonUnionNodeForPoint(current.end)
			if nextNode == firstNode {
				break
			}
			candidates := outgoing[nextNode]
			nextEdgeIndex := -1
			for _, candidate := range candidates {
				if used[candidate] {
					continue
				}
				if nextEdgeIndex < 0 || polygonUnionClockwiseTurn(
					current, edges[candidate],
				) < polygonUnionClockwiseTurn(current, edges[nextEdgeIndex]) {
					nextEdgeIndex = candidate
				}
			}
			if nextEdgeIndex < 0 {
				return nil, fmt.Errorf("polygon union outer edges are open at %.9f, %.9f", current.end.x, current.end.y)
			}
			currentEdgeIndex = nextEdgeIndex
		}
		points = polygonUnionDeduplicatePoints(points)
		area := polygonUnionSignedArea(points)
		if len(points) < 3 || math.Abs(area) <= polygonUnionEpsilon {
			continue
		}
		if area < 0 {
			polygonUnionReversePoints(points)
			area = -area
		}
		geographic := make([]GeoPoint, len(points))
		for index, point := range points {
			geographic[index] = GeoPoint{
				Longitude: NormalizeLongitude(point.x),
				Latitude:  point.y,
			}
		}
		results = append(results, resultRing{points: geographic, area: area})
		chartRings = append(chartRings, points)
	}
	// 环列表无法表达孔洞（内外边界都会当成实体面），检测到严格包含就报错。
	if inner, outer, ratio, found := polygonUnionNestedRing(chartRings); found && ratio >= polygonUnionHoleAreaFraction {
		return nil, fmt.Errorf(
			"polygon union produced a hole (ring %d, area %.6g lies strictly inside ring %d, area %.6g); a ring list cannot represent it",
			inner, math.Abs(polygonUnionSignedArea(chartRings[inner])),
			outer, math.Abs(polygonUnionSignedArea(chartRings[outer])),
		)
	}
	sort.Slice(results, func(first, second int) bool { return results[first].area > results[second].area })
	polygons := make([][]GeoPoint, len(results))
	for index, result := range results {
		polygons[index] = result.points
	}
	return polygons, nil
}

// polygonUnionHoleAreaFraction 是判定真孔洞的相对面积下限：低于它按数值细条处理，达到或超过则报错。
const polygonUnionHoleAreaFraction = 0.01

// polygonUnionNestedRing 找出被另一个环严格包含的环（孔洞）；贴边不算包含。
func polygonUnionNestedRing(rings [][]polygonUnionPoint) (inner, outer int, ratio float64, found bool) {
	bestRatio := 0.0
	for candidate := range rings {
		candidateArea := math.Abs(polygonUnionSignedArea(rings[candidate]))
		for container := range rings {
			if candidate == container {
				continue
			}
			strictlyInside := false
			contained := true
			for _, vertex := range rings[candidate] {
				switch polygonUnionPointLocation(vertex, rings[container]) {
				case 1:
					strictlyInside = true
				case 0:
					// 贴边：不改变判定
				default:
					contained = false
				}
				if !contained {
					break
				}
			}
			if !contained || !strictlyInside {
				continue
			}
			containerArea := math.Abs(polygonUnionSignedArea(rings[container]))
			if containerArea <= 0 {
				continue
			}
			if candidateRatio := candidateArea / containerArea; candidateRatio > bestRatio {
				bestRatio, inner, outer, found = candidateRatio, candidate, container, true
			}
		}
	}
	return inner, outer, bestRatio, found
}

func polygonUnionClockwiseTurn(incoming, outgoing polygonUnionEdge) float64 {
	incomingAngle := math.Atan2(incoming.end.y-incoming.start.y, incoming.end.x-incoming.start.x)
	reverseAngle := incomingAngle + math.Pi
	outgoingAngle := math.Atan2(outgoing.end.y-outgoing.start.y, outgoing.end.x-outgoing.start.x)
	turn := math.Mod(reverseAngle-outgoingAngle, 2*math.Pi)
	if turn < 0 {
		turn += 2 * math.Pi
	}
	return turn
}

// polygonUnionPointLocation returns 1 inside, 0 on the boundary, and -1 outside.
func polygonUnionPointLocation(point polygonUnionPoint, ring []polygonUnionPoint) int {
	inside := false
	for index, start := range ring {
		end := ring[(index+1)%len(ring)]
		if _, ok := polygonUnionPointSegmentParam(point, start, end); ok {
			return 0
		}
		if (start.y > point.y) != (end.y > point.y) {
			intersectionX := start.x + (end.x-start.x)*(point.y-start.y)/(end.y-start.y)
			if point.x < intersectionX {
				inside = !inside
			}
		}
	}
	if inside {
		return 1
	}
	return -1
}

func polygonUnionPointLocationOnEdges(
	point polygonUnionPoint,
	edges []polygonUnionSourceEdge,
) int {
	inside := false
	for _, edge := range edges {
		if edge.minY > point.y+polygonUnionEpsilon {
			break
		}
		if edge.maxY < point.y-polygonUnionEpsilon {
			continue
		}
		if point.x >= edge.minX-polygonUnionEpsilon && point.x <= edge.maxX+polygonUnionEpsilon &&
			point.y >= edge.minY-polygonUnionEpsilon && point.y <= edge.maxY+polygonUnionEpsilon {
			if _, ok := polygonUnionPointSegmentParam(point, edge.start, edge.end); ok {
				return 0
			}
		}
		start := edge.start
		end := edge.end
		if (start.y > point.y) != (end.y > point.y) {
			intersectionX := start.x + (end.x-start.x)*(point.y-start.y)/(end.y-start.y)
			if point.x < intersectionX {
				inside = !inside
			}
		}
	}
	if inside {
		return 1
	}
	return -1
}

// polygonUnionBuildLocationBins 要求 sortedEdges 已按 (minY, maxY) 升序，各分箱沿用该顺序。
func polygonUnionBuildLocationBins(
	sortedEdges []polygonUnionSourceEdge,
	minimumY, maximumY float64,
) [][]polygonUnionSourceEdge {
	var counts [polygonUnionLocationBinCount]int
	total := 0
	for _, edge := range sortedEdges {
		first := polygonUnionLocationBin(edge.minY-polygonUnionEpsilon, minimumY, maximumY)
		last := polygonUnionLocationBin(edge.maxY+polygonUnionEpsilon, minimumY, maximumY)
		for index := first; index <= last; index++ {
			counts[index]++
		}
		total += last - first + 1
	}
	backing := make([]polygonUnionSourceEdge, total)
	bins := make([][]polygonUnionSourceEdge, polygonUnionLocationBinCount)
	offset := 0
	for index, count := range counts {
		bins[index] = backing[offset : offset : offset+count]
		offset += count
	}
	for _, edge := range sortedEdges {
		first := polygonUnionLocationBin(edge.minY-polygonUnionEpsilon, minimumY, maximumY)
		last := polygonUnionLocationBin(edge.maxY+polygonUnionEpsilon, minimumY, maximumY)
		for index := first; index <= last; index++ {
			bins[index] = append(bins[index], edge)
		}
	}
	return bins
}

func polygonUnionPointLocationInRing(point polygonUnionPoint, ring polygonUnionRing) int {
	return polygonUnionPointLocationInRingRef(point, &ring)
}

func polygonUnionPointLocationInRingRef(point polygonUnionPoint, ring *polygonUnionRing) int {
	if len(ring.locationBins) != polygonUnionLocationBinCount {
		return polygonUnionPointLocationOnEdges(point, ring.locationEdges)
	}
	index := polygonUnionLocationBin(point.y, ring.minY, ring.maxY)
	return polygonUnionPointLocationOnEdges(point, ring.locationBins[index])
}

func polygonUnionLocationBin(value, minimum, maximum float64) int {
	if maximum <= minimum {
		return 0
	}
	index := int((value - minimum) / (maximum - minimum) * polygonUnionLocationBinCount)
	if index < 0 {
		return 0
	}
	if index >= polygonUnionLocationBinCount {
		return polygonUnionLocationBinCount - 1
	}
	return index
}

func polygonUnionUniqueParams(values []float64) []float64 {
	sort.Float64s(values)
	result := values[:0]
	for _, value := range values {
		value = polygonUnionClampParam(value)
		if len(result) == 0 || value-result[len(result)-1] > polygonUnionEpsilon {
			result = append(result, value)
		}
	}
	return result
}

func polygonUnionDeduplicatePoints(points []polygonUnionPoint) []polygonUnionPoint {
	result := points[:0]
	for _, point := range points {
		if len(result) == 0 || polygonUnionNodeForPoint(result[len(result)-1]) != polygonUnionNodeForPoint(point) {
			result = append(result, point)
		}
	}
	if len(result) > 1 && polygonUnionNodeForPoint(result[0]) == polygonUnionNodeForPoint(result[len(result)-1]) {
		result = result[:len(result)-1]
	}
	return result
}

func polygonUnionSignedArea(points []polygonUnionPoint) float64 {
	area := 0.0
	for index, current := range points {
		next := points[(index+1)%len(points)]
		area += current.x*next.y - next.x*current.y
	}
	return area / 2
}

func polygonUnionMeanLongitude(points []polygonUnionPoint) float64 {
	value := 0.0
	for _, point := range points {
		value += point.x
	}
	return value / float64(len(points))
}

func polygonUnionReversePoints(points []polygonUnionPoint) {
	for left, right := 0, len(points)-1; left < right; left, right = left+1, right-1 {
		points[left], points[right] = points[right], points[left]
	}
}

func polygonUnionInterpolate(start, end polygonUnionPoint, fraction float64) polygonUnionPoint {
	return polygonUnionPoint{
		x: start.x + (end.x-start.x)*fraction,
		y: start.y + (end.y-start.y)*fraction,
	}
}

func polygonUnionSubtract(first, second polygonUnionPoint) polygonUnionPoint {
	return polygonUnionPoint{x: first.x - second.x, y: first.y - second.y}
}

func polygonUnionAdd(first, second polygonUnionPoint) polygonUnionPoint {
	return polygonUnionPoint{x: first.x + second.x, y: first.y + second.y}
}

func polygonUnionCross(first, second polygonUnionPoint) float64 {
	return first.x*second.y - first.y*second.x
}

func polygonUnionClampParam(value float64) float64 {
	return math.Max(0, math.Min(1, value))
}

func polygonUnionNodeForPoint(point polygonUnionPoint) polygonUnionNode {
	return polygonUnionNode{
		x: int64(math.Round(point.x / polygonUnionNodeGrid)),
		y: int64(math.Round(point.y / polygonUnionNodeGrid)),
	}
}

func polygonUnionKey(edge polygonUnionEdge) polygonUnionEdgeKey {
	return polygonUnionEdgeKey{
		start: polygonUnionNodeForPoint(edge.start),
		end:   polygonUnionNodeForPoint(edge.end),
	}
}

type polygonUnionEdgesByY []polygonUnionSourceEdge

func (edges polygonUnionEdgesByY) Len() int      { return len(edges) }
func (edges polygonUnionEdgesByY) Swap(i, j int) { edges[i], edges[j] = edges[j], edges[i] }
func (edges polygonUnionEdgesByY) Less(i, j int) bool {
	if edges[i].minY != edges[j].minY {
		return edges[i].minY < edges[j].minY
	}
	return edges[i].maxY < edges[j].maxY
}

type polygonUnionEdgeRefsByX []polygonUnionSourceEdgeRef

func (edges polygonUnionEdgeRefsByX) Len() int      { return len(edges) }
func (edges polygonUnionEdgeRefsByX) Swap(i, j int) { edges[i], edges[j] = edges[j], edges[i] }
func (edges polygonUnionEdgeRefsByX) Less(i, j int) bool {
	a, b := &edges[i], &edges[j]
	if a.edge.minX != b.edge.minX {
		return a.edge.minX < b.edge.minX
	}
	if a.edge.maxX != b.edge.maxX {
		return a.edge.maxX < b.edge.maxX
	}
	if a.ringIndex != b.ringIndex {
		return a.ringIndex < b.ringIndex
	}
	return a.edgeIndex < b.edgeIndex
}
