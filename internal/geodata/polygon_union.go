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
const polygonOverlayMinArea = 1e-12
const polygonUnionContainmentToleranceKM = 1.0
const polygonUnionLocationBinCount = 64
const polygonUnionBatchThreshold = 5000000
const polygonUnionBatchVertices = 250000

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
	result, err := unionPolygonsLegacy(polygons)
	if err == nil {
		return result, nil
	}
	rings, buildErr := polygonUnionRings(polygons)
	if buildErr != nil {
		return nil, err
	}
	polygonUnionAddIntersections(rings, 0)
	result, recoveryErr := polygonUnionJoinEdges(polygonUnionLimitOuterEdges(rings, 0, regionUnion))
	if recoveryErr != nil || len(result) == 0 {
		return nil, err
	}
	return result, nil
}

func unionPolygonsLegacy(polygons [][]GeoPoint) ([][]GeoPoint, error) {
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
	polygonUnionAddIntersections(rings, 0)
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

func polygonUnionAddIntersections(rings []polygonUnionRing, split int) {
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
			if split != 0 && (first.ringIndex < split) == (second.ringIndex < split) {
				continue
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

type polygonUnionJoinedRing struct {
	points []GeoPoint
	chart  []polygonUnionPoint
	area   float64
}

func polygonUnionJoinEdges(edges []polygonUnionEdge) ([][]GeoPoint, error) {
	results, err := polygonUnionTraceEdges(edges, polygonUnionEpsilon)
	if err != nil {
		return nil, err
	}
	chartRings := make([][]polygonUnionPoint, len(results))
	for i := range results {
		chartRings[i] = results[i].chart
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

func polygonUnionTraceEdges(edges []polygonUnionEdge, minArea float64) ([]polygonUnionJoinedRing, error) {
	outgoing := make(map[polygonUnionNode][]int, len(edges))
	for index, edge := range edges {
		outgoing[polygonUnionNodeForPoint(edge.start)] = append(
			outgoing[polygonUnionNodeForPoint(edge.start)], index,
		)
	}
	used := make([]bool, len(edges))
	results := make([]polygonUnionJoinedRing, 0, 2)
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
		if len(points) < 3 || math.Abs(area) <= minArea {
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
		results = append(results, polygonUnionJoinedRing{points: geographic, chart: points, area: area})
	}
	return results, nil
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
	if len(points) < 3 {
		return 0
	}
	area := 0.0
	origin := points[0]
	for index, current := range points {
		next := points[(index+1)%len(points)]
		area += (current.x-origin.x)*(next.y-origin.y) - (next.x-origin.x)*(current.y-origin.y)
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

type regionOperation uint8

const (
	regionUnion regionOperation = iota
	regionSubtract
	regionIntersect
)

func polygonUnionLimitOuterEdges(rings []polygonUnionRing, split int, operation regionOperation) []polygonUnionEdge {
	edges := make(map[polygonUnionEdgeKey]polygonUnionEdge)
	neighbors := polygonUnionRingNeighbors(rings)
	for ringIndex, ring := range rings {
		interiorLeft := split != 0 || polygonUnionSignedArea(ring.points) > 0
		var leftInside, rightInside bool
		previousIntersection := true
		evaluated := false
		for edgeIndex, start := range ring.points {
			end := ring.points[(edgeIndex+1)%len(ring.points)]
			intersection := len(ring.params[edgeIndex]) > 2
			uncached := !evaluated || previousIntersection || intersection
			evaluatedEdge := false
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
				if split == 0 {
					leftInside, rightInside = polygonUnionLimitSides(middle, delta, rings, interiorLeft, neighbors[ringIndex])
				} else if uncached {
					left, right := polygonUnionPoint{-delta.y, delta.x}, polygonUnionPoint{delta.y, -delta.x}
					if operation == regionUnion {
						leftInside, rightInside = true, polygonUnionOtherRegionInside(middle, right, rings, neighbors[ringIndex], ringIndex, split)
					} else {
						leftInside = polygonOverlayContains(middle, left, rings, neighbors[ringIndex], ringIndex, split, operation)
						rightInside = polygonOverlayContains(middle, right, rings, neighbors[ringIndex], ringIndex, split, operation)
					}
					evaluated = true
				}
				evaluatedEdge = true
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
			if evaluatedEdge {
				previousIntersection = intersection
			} else {
				previousIntersection = previousIntersection || intersection
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

// Region 带孔的地理区域，外环逆时针、孔环顺时针 / geographic region with a counterclockwise outer ring and clockwise holes.
type Region struct {
	Outer []GeoPoint
	Holes [][]GeoPoint
}

// UnionRegions 在固定经度图幅中合并地理区域并保留孔洞 / union geographic regions in a fixed longitude chart, retaining holes.
func UnionRegions(polygons [][]GeoPoint) ([]Region, error) {
	if polygonUnionVertexCount(polygons) > polygonUnionBatchThreshold {
		return polygonUnionRegionBatches(polygons, polygonUnionBatchVertices)
	}
	return polygonUnionRegions(polygons)
}

func polygonUnionVertexCount(polygons [][]GeoPoint) int64 {
	var count int64
	for _, ring := range polygons {
		count += int64(len(ring))
	}
	return count
}

func polygonUnionRegionBatches(polygons [][]GeoPoint, limit int64) ([]Region, error) {
	if len(polygons) < 2 || polygonUnionVertexCount(polygons) <= limit {
		return polygonUnionRegions(polygons)
	}
	split := len(polygons) / 2
	left, err := polygonUnionRegionBatches(polygons[:split], limit)
	if err != nil {
		return nil, err
	}
	right, err := polygonUnionRegionBatches(polygons[split:], limit)
	if err != nil {
		return nil, err
	}
	return polygonUnionMergeRegions(left, right)
}

func polygonUnionMergeRegions(left, right []Region) ([]Region, error) {
	return overlayCanonicalRegions(left, right, regionUnion)
}

func overlayCanonicalRegions(left, right []Region, operation regionOperation) ([]Region, error) {
	var rings []polygonUnionRing
	add := func(regions []Region) {
		for _, region := range regions {
			rings = append(rings, polygonUnionCanonicalRing(region.Outer))
			for _, hole := range region.Holes {
				rings = append(rings, polygonUnionCanonicalRing(hole))
			}
		}
	}
	add(left)
	split := len(rings)
	add(right)
	if operation == regionUnion {
		polygonUnionAddIntersections(rings, split)
	} else {
		polygonUnionRegionIntersections(rings)
	}
	return polygonUnionFinishRegions(rings, split, operation)
}

func polygonOverlayContains(mid, normal polygonUnionPoint, rings []polygonUnionRing, neighbors []int, index, split int, operation regionOperation) bool {
	inside := [2]bool{}
	update := func(index int) {
		ring := &rings[index]
		if mid.x >= ring.minX-polygonUnionEpsilon && mid.x <= ring.maxX+polygonUnionEpsilon &&
			mid.y >= ring.minY-polygonUnionEpsilon && mid.y <= ring.maxY+polygonUnionEpsilon && polygonUnionLimitInside(mid, normal, ring) {
			side := 0
			if index >= split {
				side = 1
			}
			inside[side] = !inside[side]
		}
	}
	update(index)
	for _, neighbor := range neighbors {
		update(neighbor)
	}
	return inside[0] && (inside[1] == (operation == regionIntersect))
}

func polygonUnionOtherRegionInside(mid, normal polygonUnionPoint, rings []polygonUnionRing, neighbors []int, index, split int) bool {
	inside := false
	for _, j := range neighbors {
		if (index < split) == (j < split) {
			continue
		}
		ring := &rings[j]
		if mid.x < ring.minX-polygonUnionEpsilon || mid.x > ring.maxX+polygonUnionEpsilon || mid.y < ring.minY-polygonUnionEpsilon || mid.y > ring.maxY+polygonUnionEpsilon {
			continue
		}
		if polygonUnionLimitInside(mid, normal, ring) {
			inside = !inside
		}
	}
	return inside
}

func polygonUnionCanonicalRing(piece []GeoPoint) polygonUnionRing {
	if len(piece) > 1 && piece[0] == piece[len(piece)-1] {
		piece = piece[:len(piece)-1]
	}
	points := make([]polygonUnionPoint, len(piece))
	for j, p := range piece {
		points[j] = polygonUnionPoint{x: p.Longitude, y: p.Latitude}
	}
	ring := polygonUnionRingForPoints(points)
	ring.locationEdges = nil
	ring.params = make([][]float64, len(points))
	params := make([]float64, 2*len(points))
	for j := range ring.params {
		params[2*j+1] = 1
		ring.params[j] = params[2*j : 2*j+2 : 2*j+2]
	}
	return ring
}

func polygonUnionRegions(polygons [][]GeoPoint) ([]Region, error) {
	rings := make([]polygonUnionRing, 0, len(polygons))
	for k, polygon := range polygons {
		if len(openGeoRing(polygon)) < 3 {
			return nil, fmt.Errorf("polygon %d requires three points", k)
		}
		pieces := PolygonFragments(openGeoRing(polygon), ClipView{Projection: ProjectionEquirectangular})
		if len(pieces) == 0 {
			return nil, fmt.Errorf("polygon %d has no canonical pieces", k)
		}
		for _, piece := range pieces {
			rings = append(rings, polygonUnionCanonicalRing(piece))
		}
	}
	if len(rings) == 0 {
		return nil, fmt.Errorf("no input polygons")
	}
	polygonUnionRegionIntersections(rings)
	return polygonUnionFinishRegions(rings, 0, regionUnion)
}

func polygonUnionFinishRegions(rings []polygonUnionRing, split int, operation regionOperation) ([]Region, error) {
	edges := polygonUnionLimitOuterEdges(rings, split, operation)
	if len(edges) == 0 && operation != regionUnion {
		return nil, nil
	}
	joined, err := polygonUnionTraceEdges(edges, polygonOverlayMinArea)
	if err != nil {
		return nil, err
	}
	if len(joined) == 0 && operation == regionUnion {
		return nil, fmt.Errorf("no joined rings")
	}
	for i := range joined {
		for j, p := range joined[i].chart {
			joined[i].points[j] = GeoPoint{Longitude: p.x, Latitude: p.y}
		}
	}
	return polygonUnionClassifyRegions(joined), nil
}

func polygonUnionClassifyRegions(rings []polygonUnionJoinedRing) []Region {
	parent := make([]int, len(rings))
	depth := make([]int, len(rings))
	for i := range rings {
		parent[i] = -1
		for j := range rings {
			if i == j || rings[j].area <= rings[i].area {
				continue
			}
			// 求交追踪后的边界不互穿，一个非贴边点即可确定嵌套关系；全顶点贴边时再探测边中点。
			inside := false
			found := false
			for _, p := range rings[i].chart {
				position := polygonUnionPointLocation(p, rings[j].chart)
				if position != 0 {
					inside = position > 0
					found = true
					break
				}
			}
			if !found {
				for k, p := range rings[i].chart {
					next := rings[i].chart[(k+1)%len(rings[i].chart)]
					mid := polygonUnionInterpolate(p, next, 0.5)
					position := polygonUnionPointLocation(mid, rings[j].chart)
					if position != 0 {
						inside = position > 0
						break
					}
				}
			}
			if inside {
				depth[i]++
				if parent[i] < 0 || rings[j].area < rings[parent[i]].area {
					parent[i] = j
				}
			}
		}
	}
	order := make([]int, len(rings))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool { return rings[order[i]].area > rings[order[j]].area })
	result := make([]Region, 0)
	indexes := make(map[int]int)
	for _, i := range order {
		if depth[i]%2 == 0 {
			indexes[i] = len(result)
			result = append(result, Region{Outer: rings[i].points})
		}
	}
	for _, i := range order {
		if depth[i]%2 == 0 {
			continue
		}
		points := append([]GeoPoint(nil), rings[i].points...)
		for a, b := 0, len(points)-1; a < b; a, b = a+1, b-1 {
			points[a], points[b] = points[b], points[a]
		}
		ancestor := parent[i]
		for depth[ancestor]%2 != 0 {
			ancestor = parent[ancestor]
		}
		k := indexes[ancestor]
		result[k].Holes = append(result[k].Holes, points)
	}
	return result
}

func polygonUnionRingNeighbors(rings []polygonUnionRing) [][]int {
	order := make([]int, len(rings))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool { return rings[order[i]].minX < rings[order[j]].minX })
	result := make([][]int, len(rings))
	for k, i := range order {
		a := rings[i]
		for _, j := range order[k+1:] {
			b := rings[j]
			if b.minX > a.maxX+polygonUnionEpsilon {
				break
			}
			if a.minY > b.maxY+polygonUnionEpsilon || b.minY > a.maxY+polygonUnionEpsilon {
				continue
			}
			result[i] = append(result[i], j)
			result[j] = append(result[j], i)
		}
	}
	return result
}

func polygonUnionLimitSides(mid, delta polygonUnionPoint, rings []polygonUnionRing, interiorLeft bool, neighbors []int) (bool, bool) {
	left := interiorLeft
	right := !left
	normal := polygonUnionPoint{-delta.y, delta.x}
	for _, i := range neighbors {
		ring := &rings[i]
		if mid.x < ring.minX-polygonUnionEpsilon || mid.x > ring.maxX+polygonUnionEpsilon || mid.y < ring.minY-polygonUnionEpsilon || mid.y > ring.maxY+polygonUnionEpsilon {
			continue
		}
		if !left {
			left = polygonUnionLimitInside(mid, normal, ring)
		}
		if !right {
			right = polygonUnionLimitInside(mid, polygonUnionPoint{-normal.x, -normal.y}, ring)
		}
		if left && right {
			break
		}
	}
	return left, right
}

func polygonUnionLimitInside(p, n polygonUnionPoint, ring *polygonUnionRing) bool {
	const eps = 0x1p-52
	inside := false
	for _, edge := range ring.locationBins[polygonUnionLocationBin(p.y, ring.minY, ring.maxY)] {
		a, b := edge.start, edge.end
		ay := polygonUnionLimitSign(a.y-p.y, 4*eps*(math.Abs(a.y)+math.Abs(p.y)), -n.y)
		by := polygonUnionLimitSign(b.y-p.y, 4*eps*(math.Abs(b.y)+math.Abs(p.y)), -n.y)
		if (ay > 0) == (by > 0) {
			continue
		}
		d := polygonUnionSubtract(b, a)
		q := polygonUnionSubtract(p, a)
		det := polygonUnionCross(d, q)
		errorBound := 8 * eps * (math.Abs(d.x)*(math.Abs(p.y)+math.Abs(a.y)) + math.Abs(d.y)*(math.Abs(p.x)+math.Abs(a.x)) + math.Abs(d.x*q.y) + math.Abs(d.y*q.x))
		sign := polygonUnionLimitSign(det, errorBound, polygonUnionCross(d, n))
		if (d.y > 0 && sign > 0) || (d.y < 0 && sign < 0) {
			inside = !inside
		}
	}
	return inside
}

func polygonUnionLimitSign(value, roundoff, derivative float64) int {
	if value > roundoff {
		return 1
	}
	if value < -roundoff {
		return -1
	}
	if derivative > 0 {
		return 1
	}
	if derivative < 0 {
		return -1
	}
	return 0
}

func polygonUnionRegionIntersections(rings []polygonUnionRing) {
	neighbors := polygonUnionRingNeighbors(rings)
	for i := range rings {
		a := &rings[i]
		bins := make([][]int, polygonUnionLocationBinCount)
		for j, e := range a.edges {
			first := polygonUnionLocationBin(e.minY-polygonUnionEpsilon, a.minY, a.maxY)
			last := polygonUnionLocationBin(e.maxY+polygonUnionEpsilon, a.minY, a.maxY)
			for k := first; k <= last; k++ {
				bins[k] = append(bins[k], j)
			}
		}
		candidates := append([]int{i}, neighbors[i]...)
		for _, j := range candidates {
			if j < i {
				continue
			}
			b := &rings[j]
			for bi, second := range b.edges {
				if second.maxY < a.minY-polygonUnionEpsilon || second.minY > a.maxY+polygonUnionEpsilon {
					continue
				}
				firstBin := polygonUnionLocationBin(second.minY-polygonUnionEpsilon, a.minY, a.maxY)
				lastBin := polygonUnionLocationBin(second.maxY+polygonUnionEpsilon, a.minY, a.maxY)
				for k := firstBin; k <= lastBin; k++ {
					for _, ai := range bins[k] {
						first := a.edges[ai]
						startBin := polygonUnionLocationBin(first.minY-polygonUnionEpsilon, a.minY, a.maxY)
						if k > firstBin && k > startBin {
							continue
						}
						if i == j && (ai >= bi || polygonUnionEdgesAdjacent(ai, bi, len(a.points))) {
							continue
						}
						if !polygonUnionEdgeBoundsOverlap(first, second) {
							continue
						}
						ap, an, bp, bn := polygonUnionSegmentIntersections(first.start, first.end, second.start, second.end)
						a.params[ai] = append(a.params[ai], ap[:an]...)
						b.params[bi] = append(b.params[bi], bp[:bn]...)
					}
				}
			}
		}
	}
}
