package geodata

import (
	"errors"
	"fmt"
	"math"
)

// ErrOpenBoundaryEndpointTrackReversal 表示扫掠边界的端点轨迹发生了物理折返。
// ErrOpenBoundaryEndpointTrackReversal reports a physical reversal along a swept boundary endpoint track.
var ErrOpenBoundaryEndpointTrackReversal = errors.New("open boundary endpoint track reversal")

const (
	openBoundaryEndpointStepNoiseKM    = 1.0
	openBoundaryEndpointProgressBackKM = 5.0
)

// OpenBoundarySweepSample 是用于构造连续扫掠区域的一个边界样本。
// OpenBoundarySweepSample is one sampled boundary used to build a continuous
// sweep. Closed samples are ignored because they already describe an
// instantaneous footprint and are not part of an open non-central band.
type OpenBoundarySweepSample struct {
	Boundaries [][]GeoPoint
	Closed     bool
}

type openBoundarySweepSample struct {
	boundary []GeoPoint
	start    GeoPoint
	end      GeoPoint
}

// OpenBoundarySweep 将开放边界弧连接为连续扫掠多边形。
// OpenBoundarySweep joins open boundary arcs into continuous swept polygons.
// The first and last arcs close each group; the matching endpoint tracks form
// the two sides. A large endpoint jump starts a new group instead of creating
// a false bridge across a branch change.
func OpenBoundarySweep(samples []OpenBoundarySweepSample) ([][]GeoPoint, error) {
	groups, err := openBoundarySweepGroups(samples)
	if err != nil {
		return nil, err
	}

	polygons := make([][]GeoPoint, 0, len(groups))
	for _, group := range groups {
		groupPolygons, groupErr := sweepOpenBoundaryGroup(group)
		if groupErr != nil {
			return nil, groupErr
		}
		polygons = append(polygons, groupPolygons...)
	}
	return polygons, nil
}

// MonotoneOpenBoundarySweep 连接端点轨迹和中间弧已知单调的开放弧。
// MonotoneOpenBoundarySweep joins open arcs whose endpoint tracks and
// intermediate arcs are already known to be monotone. It avoids the general
// containment audit used by OpenBoundarySweep and is intended for tightly
// sampled, event-local contact caps.
func MonotoneOpenBoundarySweep(samples []OpenBoundarySweepSample) ([][]GeoPoint, error) {
	groups, err := openBoundarySweepGroups(samples)
	if err != nil {
		return nil, err
	}
	polygons := make([][]GeoPoint, 0, len(groups))
	for index, group := range groups {
		if !sweepEndpointTracksMonotone(group) {
			return nil, fmt.Errorf("%w in group %d", ErrOpenBoundaryEndpointTrackReversal, index)
		}
		outline := sweepOpenBoundaryOutline(group)
		if len(outline) < 3 {
			return nil, fmt.Errorf("monotone open boundary sweep group %d has no usable outline", index)
		}
		if !sweepOutlineContainsRepresentativeSamples(outline, group) {
			return nil, fmt.Errorf("monotone open boundary sweep group %d folds outside its endpoint tracks", index)
		}
		polygons = append(polygons, outline)
	}
	return polygons, nil
}

func sweepEndpointTracksMonotone(group []openBoundarySweepSample) bool {
	if len(group) < 3 {
		return true
	}
	for _, endpoint := range []func(openBoundarySweepSample) GeoPoint{
		func(sample openBoundarySweepSample) GeoPoint { return sample.start },
		func(sample openBoundarySweepSample) GeoPoint { return sample.end },
	} {
		first := geoPointUnitVector(endpoint(group[0]))
		last := geoPointUnitVector(endpoint(group[len(group)-1]))
		direction := geoPointTangentDirection(first, last)
		if vectorNorm(direction) <= 1e-12 {
			continue
		}
		previousProgress := 0.0
		previousVector := first
		var previousStep [3]float64
		for index := 1; index < len(group); index++ {
			current := geoPointUnitVector(endpoint(group[index]))
			step := [3]float64{
				current[0] - previousVector[0],
				current[1] - previousVector[1],
				current[2] - previousVector[2],
			}
			if index > 1 {
				previousLength := vectorNorm(previousStep)
				currentLength := vectorNorm(step)
				if math.Min(previousLength, currentLength)*EarthMeanRadiusKM > openBoundaryEndpointStepNoiseKM &&
					vectorDot(previousStep, step) < -0.1*previousLength*currentLength {
					return false
				}
			}
			progress := vectorDot(current, direction)
			// A long spherical track can deviate slightly from its first-to-last
			// tangent chord even while every local step remains forward. Reserve a
			// few kilometres for that projection effect; larger gradual reversals
			// still invalidate the endpoint envelope.
			if progress+openBoundaryEndpointProgressBackKM/EarthMeanRadiusKM < previousProgress {
				return false
			}
			previousProgress = progress
			previousVector = current
			previousStep = step
		}
	}
	return true
}

func geoPointUnitVector(point GeoPoint) [3]float64 {
	latitude := point.Latitude * math.Pi / 180
	longitude := point.Longitude * math.Pi / 180
	cosLatitude := math.Cos(latitude)
	return [3]float64{
		cosLatitude * math.Cos(longitude),
		cosLatitude * math.Sin(longitude),
		math.Sin(latitude),
	}
}

func geoPointTangentDirection(first, last [3]float64) [3]float64 {
	dot := vectorDot(first, last)
	return [3]float64{
		last[0] - dot*first[0],
		last[1] - dot*first[1],
		last[2] - dot*first[2],
	}
}

func vectorDot(first, second [3]float64) float64 {
	return first[0]*second[0] + first[1]*second[1] + first[2]*second[2]
}

func vectorNorm(value [3]float64) float64 {
	return math.Sqrt(vectorDot(value, value))
}

// OpenBoundaryEndpointOutlines 返回各开放边界分组的低成本端点轨迹轮廓。
// OpenBoundaryEndpointOutlines returns the inexpensive endpoint-track outline
// for each open-boundary group. The outlines intentionally omit intermediate
// arc bulges, so callers must combine them with the source footprint polygons;
// they are not a standalone geometric union.
func OpenBoundaryEndpointOutlines(samples []OpenBoundarySweepSample) ([][]GeoPoint, error) {
	groups, err := openBoundarySweepGroups(samples)
	if err != nil {
		return nil, err
	}
	polygons := make([][]GeoPoint, 0, len(groups))
	for index, group := range groups {
		if len(group) < 2 {
			continue
		}
		outline := sweepOpenBoundaryOutline(group)
		if len(outline) < 3 {
			return nil, fmt.Errorf("open boundary endpoint group %d has no usable outline", index)
		}
		polygons = append(polygons, outline)
	}
	if len(polygons) == 0 {
		return nil, fmt.Errorf("open boundary samples contain no endpoint track")
	}
	return polygons, nil
}

// DecimateOpenBoundarySweepSamples 在保留端点和分组的前提下限制通用带状 union 的采样成本。
// DecimateOpenBoundarySweepSamples bounds the cost of the general ribbon
// union while retaining every group endpoint and closed-sample separator.
func DecimateOpenBoundarySweepSamples(
	samples []OpenBoundarySweepSample,
	maximumGroupSamples int,
	targetSpacingKM float64,
) []OpenBoundarySweepSample {
	if maximumGroupSamples < 2 {
		maximumGroupSamples = 2
	}
	result := make([]OpenBoundarySweepSample, 0, len(samples))
	flush := func(group []OpenBoundarySweepSample) {
		if len(group) == 0 {
			return
		}
		selected := group
		if len(group) > maximumGroupSamples {
			selected = make([]OpenBoundarySweepSample, 0, maximumGroupSamples)
			last := len(group) - 1
			lastPosition := -1
			for index := 0; index < maximumGroupSamples; index++ {
				position := int(math.Round(float64(index*last) / float64(maximumGroupSamples-1)))
				if position != lastPosition {
					selected = append(selected, group[position])
					lastPosition = position
				}
			}
		}
		for _, sample := range selected {
			copySample := OpenBoundarySweepSample{Closed: sample.Closed}
			copySample.Boundaries = make([][]GeoPoint, len(sample.Boundaries))
			for index, boundary := range sample.Boundaries {
				copySample.Boundaries[index] = decimateOpenBoundaryPoints(boundary, targetSpacingKM)
			}
			result = append(result, copySample)
		}
	}
	group := make([]OpenBoundarySweepSample, 0, maximumGroupSamples)
	for _, sample := range samples {
		if sample.Closed {
			flush(group)
			group = group[:0]
			if len(result) == 0 || !result[len(result)-1].Closed {
				result = append(result, OpenBoundarySweepSample{Closed: true})
			}
			continue
		}
		group = append(group, sample)
	}
	flush(group)
	return result
}

// OpenBoundarySweepInnerCaps 闭合紧邻内部闭合阶段的小地平缺口。
// OpenBoundarySweepInnerCaps closes the small horizon gaps immediately next
// to a closed-sample separator. Those nearly closed arcs overlap the paired
// middle band; leaving their gap open can retain a narrow inward notch after
// polygon union.
func OpenBoundarySweepInnerCaps(
	samples []OpenBoundarySweepSample,
	maximumGapKM float64,
) [][]GeoPoint {
	if maximumGapKM <= 0 {
		return nil
	}
	indices := make([]int, 0, 2)
	for index := 0; index < len(samples); index++ {
		if !samples[index].Closed {
			continue
		}
		if index > 0 && !samples[index-1].Closed {
			indices = append(indices, index-1)
		}
		for index+1 < len(samples) && samples[index+1].Closed {
			index++
		}
		if index+1 < len(samples) && !samples[index+1].Closed {
			indices = append(indices, index+1)
		}
	}
	caps := make([][]GeoPoint, 0, len(indices))
	for _, index := range indices {
		boundary := openGeoRing(JoinPolylineSegments(samples[index].Boundaries))
		if len(boundary) < 3 || EquatorialDistanceKM(boundary[0], boundary[len(boundary)-1]) > maximumGapKM {
			continue
		}
		caps = append(caps, boundary)
	}
	return caps
}

func decimateOpenBoundaryPoints(points []GeoPoint, targetSpacingKM float64) []GeoPoint {
	if len(points) < 3 || targetSpacingKM <= 0 {
		return append([]GeoPoint(nil), points...)
	}
	result := make([]GeoPoint, 1, len(points))
	result[0] = points[0]
	for _, point := range points[1 : len(points)-1] {
		if EquatorialDistanceKM(result[len(result)-1], point) >= targetSpacingKM {
			result = append(result, point)
		}
	}
	if !SameGeoPoint(result[len(result)-1], points[len(points)-1]) {
		result = append(result, points[len(points)-1])
	}
	return result
}

func sweepOutlineContainsRepresentativeSamples(
	outline []GeoPoint,
	group []openBoundarySweepSample,
) bool {
	const checksPerArc = 5
	for _, sample := range group {
		for check := 0; check < checksPerArc; check++ {
			index := check * (len(sample.boundary) - 1) / (checksPerArc - 1)
			if !sweepPointInPolygon(outline, sample.boundary[index]) {
				return false
			}
		}
	}
	return true
}

func openBoundarySweepGroups(samples []OpenBoundarySweepSample) ([][]openBoundarySweepSample, error) {
	groups := make([][]openBoundarySweepSample, 0, 2)
	current := make([]openBoundarySweepSample, 0, len(samples))
	var previous openBoundarySweepSample
	for _, source := range samples {
		// A closed footprint is an instantaneous cap, not an open arc. It
		// terminates the current open branch so later samples cannot bridge
		// across a visibility/type transition.
		if source.Closed {
			if len(current) > 0 {
				groups = append(groups, current)
				current = nil
			}
			previous = openBoundarySweepSample{}
			continue
		}
		boundary := JoinPolylineSegments(source.Boundaries)
		boundary = openGeoRing(boundary)
		if len(boundary) < 2 {
			continue
		}
		value := openBoundarySweepSample{boundary: boundary, start: boundary[0], end: boundary[len(boundary)-1]}
		if len(current) > 0 {
			keep := EquatorialDistanceKM(previous.start, value.start) +
				EquatorialDistanceKM(previous.end, value.end)
			reverse := EquatorialDistanceKM(previous.start, value.end) +
				EquatorialDistanceKM(previous.end, value.start)
			if reverse < keep {
				reverseSweepGeoPoints(value.boundary)
				value.start, value.end = value.end, value.start
			}
			if math.Max(
				EquatorialDistanceKM(previous.start, value.start),
				EquatorialDistanceKM(previous.end, value.end),
			) > 2000 {
				groups = append(groups, current)
				current = nil
			}
		}
		current = append(current, value)
		previous = value
	}
	if len(current) > 0 {
		groups = append(groups, current)
	}
	if len(groups) == 0 {
		return nil, fmt.Errorf("open boundary samples contain no usable arcs")
	}
	return groups, nil
}

// sweepOpenBoundaryGroup constructs the union swept by a sequence of open
// arcs. Each adjacent pair forms a ribbon whose caps are the two sampled arcs
// and whose sides follow their endpoints. Unioning the ribbons retains any
// intermediate arc that becomes part of a non-monotone outer boundary; a
// convex hull would replace that boundary with visibly incorrect chords.
func sweepOpenBoundaryGroup(group []openBoundarySweepSample) ([][]GeoPoint, error) {
	if len(group) == 0 {
		return nil, fmt.Errorf("open boundary sweep group is empty")
	}
	if len(group) == 1 {
		return [][]GeoPoint{append([]GeoPoint(nil), group[0].boundary...)}, nil
	}
	if outline := sweepOpenBoundaryOutline(group); len(outline) >= 3 && sweepOutlineContainsSamples(outline, group) {
		return [][]GeoPoint{outline}, nil
	}
	ribbons := make([][]GeoPoint, 0, len(group)-1)
	for index := 1; index < len(group); index++ {
		first := group[index-1].boundary
		second := group[index].boundary
		ring := make([]GeoPoint, 0, len(first)+len(second))
		ring = append(ring, first...)
		for pointIndex := len(second) - 1; pointIndex >= 0; pointIndex-- {
			ring = append(ring, second[pointIndex])
		}
		ring = sweepDeduplicateAdjacent(ring)
		if len(ring) < 3 {
			return nil, fmt.Errorf("open boundary sweep ribbon %d has no usable boundary", index-1)
		}
		ribbons = append(ribbons, ring)
	}
	polygons, err := UnionPolygons(ribbons)
	if err != nil {
		return nil, fmt.Errorf("open boundary sweep ribbons: %w", err)
	}
	polygons, err = sweepBridgeTouchingPolygons(polygons)
	if err != nil {
		return nil, fmt.Errorf("open boundary sweep touching ribbons: %w", err)
	}
	return polygons, nil
}

// sweepOpenBoundaryOutline is the simple outer envelope for a non-crossing
// sequence of arcs. The two endpoint tracks are the lateral edges of the
// swept band; using them directly avoids retaining every overlapping ribbon
// edge as a false inward spike near a tangent horizon.
func sweepOpenBoundaryOutline(group []openBoundarySweepSample) []GeoPoint {
	if len(group) < 2 {
		return nil
	}
	first, last := group[0].boundary, group[len(group)-1].boundary
	if len(first) < 2 || len(last) < 2 {
		return nil
	}
	ring := make([]GeoPoint, 0, len(first)+len(last)+2*len(group))
	ring = appendUnwrappedSweepPoints(ring, first)
	for index := 1; index < len(group); index++ {
		ring = appendUnwrappedSweepPoint(ring, group[index].end)
	}
	for index := len(last) - 1; index >= 0; index-- {
		ring = appendUnwrappedSweepPoint(ring, last[index])
	}
	for index := len(group) - 2; index >= 0; index-- {
		ring = appendUnwrappedSweepPoint(ring, group[index].start)
	}
	return sweepDeduplicateAdjacent(ring)
}

func appendUnwrappedSweepPoints(result []GeoPoint, points []GeoPoint) []GeoPoint {
	for _, point := range points {
		result = appendUnwrappedSweepPoint(result, point)
	}
	return result
}

func appendUnwrappedSweepPoint(result []GeoPoint, point GeoPoint) []GeoPoint {
	if len(result) == 0 {
		return append(result, point)
	}
	longitude := point.Longitude
	for longitude-result[len(result)-1].Longitude > 180 {
		longitude -= 360
	}
	for longitude-result[len(result)-1].Longitude < -180 {
		longitude += 360
	}
	return append(result, GeoPoint{Longitude: longitude, Latitude: point.Latitude})
}

func sweepOutlineContainsSamples(outline []GeoPoint, group []openBoundarySweepSample) bool {
	if len(outline) < 3 {
		return false
	}
	for _, sample := range group {
		for _, point := range sample.boundary {
			if !sweepPointInPolygon(outline, point) {
				return false
			}
		}
	}
	return true
}

func sweepPointInPolygon(polygon []GeoPoint, point GeoPoint) bool {
	if len(polygon) < 3 {
		return false
	}
	longitude := point.Longitude
	for longitude-polygon[0].Longitude > 180 {
		longitude -= 360
	}
	for longitude-polygon[0].Longitude < -180 {
		longitude += 360
	}
	probe := GeoPoint{Longitude: longitude, Latitude: point.Latitude}
	for current, previous := 0, len(polygon)-1; current < len(polygon); previous, current = current, current+1 {
		if sweepPointOnSegment(probe, polygon[previous], polygon[current]) {
			return true
		}
	}
	inside := false
	for current, previous := 0, len(polygon)-1; current < len(polygon); previous, current = current, current+1 {
		first, second := polygon[previous], polygon[current]
		if (first.Latitude > point.Latitude) == (second.Latitude > point.Latitude) {
			continue
		}
		intersection := first.Longitude +
			(point.Latitude-first.Latitude)*(second.Longitude-first.Longitude)/(second.Latitude-first.Latitude)
		if intersection >= longitude {
			inside = !inside
		}
	}
	return inside
}

func sweepPointOnSegment(point, start, end GeoPoint) bool {
	deltaLongitude := end.Longitude - start.Longitude
	deltaLatitude := end.Latitude - start.Latitude
	length := math.Hypot(deltaLongitude, deltaLatitude)
	if length <= 1e-12 {
		return math.Hypot(point.Longitude-start.Longitude, point.Latitude-start.Latitude) <= 1e-9
	}
	cross := deltaLongitude*(point.Latitude-start.Latitude) -
		deltaLatitude*(point.Longitude-start.Longitude)
	if math.Abs(cross) > 1e-9*length {
		return false
	}
	dot := (point.Longitude-start.Longitude)*deltaLongitude +
		(point.Latitude-start.Latitude)*deltaLatitude
	return dot >= -1e-9*length && dot <= length*length+1e-9*length
}

func sweepBridgeTouchingPolygons(polygons [][]GeoPoint) ([][]GeoPoint, error) {
	const (
		touchingDistanceKM = 0.01
		bridgeHalfSizeDeg  = 1e-4
	)
	for len(polygons) > 1 {
		firstIndex, secondIndex, touchingPoint, ok := sweepTouchingPolygonPair(polygons, touchingDistanceKM)
		if !ok {
			return polygons, nil
		}
		bridge := []GeoPoint{
			{Longitude: touchingPoint.Longitude - bridgeHalfSizeDeg, Latitude: touchingPoint.Latitude - bridgeHalfSizeDeg},
			{Longitude: touchingPoint.Longitude + bridgeHalfSizeDeg, Latitude: touchingPoint.Latitude - bridgeHalfSizeDeg},
			{Longitude: touchingPoint.Longitude + bridgeHalfSizeDeg, Latitude: touchingPoint.Latitude + bridgeHalfSizeDeg},
			{Longitude: touchingPoint.Longitude - bridgeHalfSizeDeg, Latitude: touchingPoint.Latitude + bridgeHalfSizeDeg},
		}
		// 输入是并集输出，环之间互不重叠，因此只需合并这一对。
		merged, err := UnionPolygons([][]GeoPoint{polygons[firstIndex], polygons[secondIndex], bridge})
		if err != nil {
			return nil, err
		}
		if len(merged) >= 2 {
			return polygons, nil
		}
		polygons = sweepReplaceTouchingPolygonPair(polygons, firstIndex, secondIndex, merged)
	}
	return polygons, nil
}

const sweepTouchingCellSize = 2e-6

type sweepTouchingVertex struct {
	polygon int
	point   GeoPoint
}

// sweepTouchingPolygonPair 返回扫描顺序最前的一对近邻顶点：先按多边形下标，再按顶点下标。
func sweepTouchingPolygonPair(
	polygons [][]GeoPoint,
	thresholdKM float64,
) (int, int, GeoPoint, bool) {
	cells := make(map[int64][]sweepTouchingVertex)
	for polygonIndex, polygon := range polygons {
		for _, point := range polygon {
			key := sweepTouchingCellKey(geoPointVector(point))
			cells[key] = append(cells[key], sweepTouchingVertex{polygon: polygonIndex, point: point})
		}
	}
	pairs := make(map[[2]int]struct{})
	for polygonIndex, polygon := range polygons {
		for _, point := range polygon {
			x, y, z := sweepTouchingCellCoordinates(geoPointVector(point))
			for deltaX := int64(-1); deltaX <= 1; deltaX++ {
				for deltaY := int64(-1); deltaY <= 1; deltaY++ {
					for deltaZ := int64(-1); deltaZ <= 1; deltaZ++ {
						key := sweepTouchingCellKeyFromCoordinates(x+deltaX, y+deltaY, z+deltaZ)
						for _, candidate := range cells[key] {
							if candidate.polygon <= polygonIndex ||
								EquatorialDistanceKM(point, candidate.point) > thresholdKM {
								continue
							}
							pairs[[2]int{polygonIndex, candidate.polygon}] = struct{}{}
						}
					}
				}
			}
		}
	}
	firstIndex, secondIndex := -1, -1
	for pair := range pairs {
		if firstIndex < 0 || pair[0] < firstIndex || (pair[0] == firstIndex && pair[1] < secondIndex) {
			firstIndex, secondIndex = pair[0], pair[1]
		}
	}
	if firstIndex < 0 {
		return 0, 0, GeoPoint{}, false
	}
	for _, first := range polygons[firstIndex] {
		for _, second := range polygons[secondIndex] {
			if !(EquatorialDistanceKM(first, second) > thresholdKM) {
				return firstIndex, secondIndex, GeoPoint{
					Longitude: first.Longitude + math.Remainder(second.Longitude-first.Longitude, 360)/2,
					Latitude:  (first.Latitude + second.Latitude) / 2,
				}, true
			}
		}
	}
	return 0, 0, GeoPoint{}, false
}

// 网格边长 2e-6 弧度，大于 0.01 公里的球面弦长，近邻顶点必落在相邻格。
func sweepTouchingCellCoordinates(vector geoVector3) (int64, int64, int64) {
	const offset = 1 << 19
	return int64(math.Floor(vector.x/sweepTouchingCellSize)) + offset,
		int64(math.Floor(vector.y/sweepTouchingCellSize)) + offset,
		int64(math.Floor(vector.z/sweepTouchingCellSize)) + offset
}

func sweepTouchingCellKey(vector geoVector3) int64 {
	x, y, z := sweepTouchingCellCoordinates(vector)
	return sweepTouchingCellKeyFromCoordinates(x, y, z)
}

func sweepTouchingCellKeyFromCoordinates(x, y, z int64) int64 {
	return x<<40 | y<<20 | z
}

func sweepReplaceTouchingPolygonPair(
	polygons [][]GeoPoint,
	first, second int,
	merged [][]GeoPoint,
) [][]GeoPoint {
	if first > second {
		first, second = second, first
	}
	result := make([][]GeoPoint, 0, len(polygons)-1+len(merged))
	result = append(result, polygons[:first]...)
	result = append(result, merged...)
	result = append(result, polygons[first+1:second]...)
	result = append(result, polygons[second+1:]...)
	return result
}

func sweepDeduplicateAdjacent(points []GeoPoint) []GeoPoint {
	if len(points) < 2 {
		return points
	}
	result := make([]GeoPoint, 0, len(points))
	for _, point := range points {
		if len(result) == 0 || !SameGeoPoint(result[len(result)-1], point) {
			result = append(result, point)
		}
	}
	if len(result) > 1 && SameGeoPoint(result[0], result[len(result)-1]) {
		result = result[:len(result)-1]
	}
	return result
}

func openGeoRing(points []GeoPoint) []GeoPoint {
	if len(points) > 1 && SameGeoPoint(points[0], points[len(points)-1]) {
		return points[:len(points)-1]
	}
	return points
}

func reverseSweepGeoPoints(points []GeoPoint) {
	for left, right := 0, len(points)-1; left < right; left, right = left+1, right-1 {
		points[left], points[right] = points[right], points[left]
	}
}
