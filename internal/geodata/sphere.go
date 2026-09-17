package geodata

import "math"

// SphericalCircle 返回球面小圆上的等间隔采样点 / SphericalCircle returns evenly spaced points on a small circle on the
// 球面小圆；方位角从地理北方顺时针采样 / sphere. Bearings are sampled clockwise from geographic north.
func SphericalCircle(center GeoPoint, radiusDegrees float64, points int) []GeoPoint {
	// r=0 时采样点重合、r=180 时只剩对跖点副本，都不构成环。
	if points < 3 || !(radiusDegrees > 0 && radiusDegrees < 180) {
		return nil
	}
	latitude := center.Latitude * math.Pi / 180
	longitude := center.Longitude * math.Pi / 180
	radius := radiusDegrees * math.Pi / 180
	result := make([]GeoPoint, points)
	for index := range result {
		bearing := 2 * math.Pi * float64(index) / float64(points)
		lat := math.Asin(math.Sin(latitude)*math.Cos(radius) +
			math.Cos(latitude)*math.Sin(radius)*math.Cos(bearing))
		lon := longitude + math.Atan2(
			math.Sin(bearing)*math.Sin(radius)*math.Cos(latitude),
			math.Cos(radius)-math.Sin(latitude)*math.Sin(lat),
		)
		result[index] = GeoPoint{
			Longitude: normalizeLongitude(lon * 180 / math.Pi),
			Latitude:  lat * 180 / math.Pi,
		}
	}
	return result
}

type geoVector3 struct {
	x, y, z float64
}

func geoPointVector(point GeoPoint) geoVector3 {
	latitude := point.Latitude * math.Pi / 180
	longitude := point.Longitude * math.Pi / 180
	cosLatitude := math.Cos(latitude)
	return geoVector3{
		x: cosLatitude * math.Cos(longitude),
		y: cosLatitude * math.Sin(longitude),
		z: math.Sin(latitude),
	}
}

func geoVectorPoint(vector geoVector3) GeoPoint {
	length := math.Sqrt(vector.x*vector.x + vector.y*vector.y + vector.z*vector.z)
	if length == 0 || math.IsNaN(length) || math.IsInf(length, 0) {
		return GeoPoint{Longitude: math.NaN(), Latitude: math.NaN()}
	}
	vector.x /= length
	vector.y /= length
	vector.z /= length
	return GeoPoint{
		Longitude: normalizeLongitude(math.Atan2(vector.y, vector.x) * 180 / math.Pi),
		Latitude:  math.Asin(math.Max(-1, math.Min(1, vector.z))) * 180 / math.Pi,
	}
}

func geoVectorDot(first, second geoVector3) float64 {
	return first.x*second.x + first.y*second.y + first.z*second.z
}

func geoVectorCross(first, second geoVector3) geoVector3 {
	return geoVector3{
		x: first.y*second.z - first.z*second.y,
		y: first.z*second.x - first.x*second.z,
		z: first.x*second.y - first.y*second.x,
	}
}

func geoVectorScale(vector geoVector3, scale float64) geoVector3 {
	return geoVector3{x: vector.x * scale, y: vector.y * scale, z: vector.z * scale}
}

func geoVectorAdd(first, second geoVector3) geoVector3 {
	return geoVector3{x: first.x + second.x, y: first.y + second.y, z: first.z + second.z}
}

func geoVectorNormalize(vector geoVector3) (geoVector3, bool) {
	length := math.Sqrt(geoVectorDot(vector, vector))
	if length <= 1e-15 || math.IsNaN(length) || math.IsInf(length, 0) {
		return geoVector3{}, false
	}
	return geoVectorScale(vector, 1/length), true
}

// InterpolateGreatCircle 在较短大圆弧上按分数插值，分数超出 [0,1] 时沿弧延长 / interpolates the shorter great-circle arc, extending it beyond an endpoint.
func InterpolateGreatCircle(first, second GeoPoint, fraction float64) GeoPoint {
	return sphericalInterpolate(first, second, fraction)
}

func sphericalInterpolate(first, second GeoPoint, fraction float64) GeoPoint {
	a := geoPointVector(first)
	b := geoPointVector(second)
	dot := math.Max(-1, math.Min(1, geoVectorDot(a, b)))
	if dot > 1-1e-14 {
		return geoVectorPoint(geoVectorAdd(geoVectorScale(a, 1-fraction), geoVectorScale(b, fraction)))
	}
	if dot < -1+1e-14 {
		// Antipodal endpoints have no unique great circle. The path samplers
		// never intentionally create one, but retain a finite fallback for
		// malformed caller input.
		return GeoPoint{
			Longitude: normalizeLongitude(first.Longitude + fraction*normalizeLongitude(second.Longitude-first.Longitude)),
			Latitude:  first.Latitude + fraction*(second.Latitude-first.Latitude),
		}
	}
	angle := math.Acos(dot)
	sine := math.Sin(angle)
	value := geoVectorAdd(
		geoVectorScale(a, math.Sin((1-fraction)*angle)/sine),
		geoVectorScale(b, math.Sin(fraction*angle)/sine),
	)
	return geoVectorPoint(value)
}

func sphericalPointOnArc(point, first, second GeoPoint) bool {
	firstVector := geoPointVector(first)
	secondVector := geoPointVector(second)
	pointVector := geoPointVector(point)
	arc := math.Acos(math.Max(-1, math.Min(1, geoVectorDot(firstVector, secondVector))))
	if arc <= 1e-14 {
		return math.Acos(math.Max(-1, math.Min(1, geoVectorDot(firstVector, pointVector)))) <= 1e-9
	}
	firstDistance := math.Acos(math.Max(-1, math.Min(1, geoVectorDot(firstVector, pointVector))))
	secondDistance := math.Acos(math.Max(-1, math.Min(1, geoVectorDot(pointVector, secondVector))))
	return math.Abs(firstDistance+secondDistance-arc) <= 1e-9
}

// sphericalPolygonContainsOrTouches tests a simple ring using its minor great
// circle edges. It is intentionally internal: map encoders still own the
// projection-specific winding rules, while topology code needs a seam-free
// containment predicate for pole and antimeridian decisions.
func sphericalPolygonContainsOrTouches(polygon []GeoPoint, point GeoPoint) bool {
	if len(polygon) < 3 {
		return false
	}
	if SameGeoPoint(polygon[0], polygon[len(polygon)-1]) {
		polygon = polygon[:len(polygon)-1]
	}
	target := geoPointVector(point)
	angleSum := 0.0
	area := 0.0
	origin := geoPointVector(polygon[0])
	for index, currentPoint := range polygon {
		previousPoint := polygon[(index+len(polygon)-1)%len(polygon)]
		if sphericalPointOnArc(point, previousPoint, currentPoint) {
			return true
		}
		previous := geoPointVector(previousPoint)
		current := geoPointVector(currentPoint)
		area += 2 * math.Atan2(geoVectorDot(origin, geoVectorCross(previous, current)),
			1+geoVectorDot(origin, previous)+geoVectorDot(previous, current)+geoVectorDot(current, origin))
		previousTangent := geoVectorAdd(previous, geoVectorScale(target, -geoVectorDot(previous, target)))
		currentTangent := geoVectorAdd(current, geoVectorScale(target, -geoVectorDot(current, target)))
		previousTangent, previousOK := geoVectorNormalize(previousTangent)
		currentTangent, currentOK := geoVectorNormalize(currentTangent)
		if !previousOK || !currentOK {
			// A vertex and its antipode both have no tangent. Only the vertex
			// belongs to the ring; acos roundoff can miss it in the arc check.
			return (!previousOK && geoVectorDot(previous, target) > 0) ||
				(!currentOK && geoVectorDot(current, target) > 0)
		}
		cross := geoVectorCross(previousTangent, currentTangent)
		angleSum += math.Atan2(geoVectorDot(target, cross), geoVectorDot(previousTangent, currentTangent))
	}
	// Tangent winding has opposite signs inside the polygon and its antipodal
	// image. Match the signed minor area instead of accepting both images.
	return math.Abs(angleSum) > math.Pi && angleSum*math.Remainder(area, 4*math.Pi) > 0
}

// SphericalPolygonsContainPaths 判断每个路径顶点和加密边中点是否都位于球面多边形内。
// SphericalPolygonsContainPaths reports whether every path vertex and every
// minor-great-circle edge midpoint lies in or on at least one polygon. When
// closePaths is true, the last vertex of each path is also joined to its first.
func SphericalPolygonsContainPaths(polygons, paths [][]GeoPoint, closePaths bool) bool {
	return SphericalPolygonsContainPathsWithinKM(polygons, paths, closePaths, 0)
}

// SphericalPolygonIndex 可复用的球面多边形包含索引 / a reusable containment index for one polygon set.
type SphericalPolygonIndex struct {
	containment sphericalPolygonContainment
}

// NewSphericalPolygonIndex 为多边形集合构建包含索引 / builds a containment index for the polygon set.
func NewSphericalPolygonIndex(polygons [][]GeoPoint) *SphericalPolygonIndex {
	return &SphericalPolygonIndex{containment: newSphericalPolygonContainment(polygons)}
}

// ContainsPoints 返回各点是否位于多边形内，结果与 points 对齐 / reports per-point containment aligned with points.
func (index *SphericalPolygonIndex) ContainsPoints(points []GeoPoint) []bool {
	result := make([]bool, len(points))
	if index == nil {
		return result
	}
	for pointIndex, point := range points {
		result[pointIndex] = index.containment.contains(point)
	}
	return result
}

// SphericalPolygonsContainPoints 用共享索引批量测试各个独立点 / tests independent points against one shared spherical polygon index.
func SphericalPolygonsContainPoints(polygons [][]GeoPoint, points []GeoPoint) []bool {
	return NewSphericalPolygonIndex(polygons).ContainsPoints(points)
}

// SphericalPolygonsContainPathsWithinKM 是 SphericalPolygonsContainPaths 的容差感知形式。
// SphericalPolygonsContainPathsWithinKM is the tolerance-aware form of
// SphericalPolygonsContainPaths. It stops at the first probe farther than the
// requested distance from every polygon.
func SphericalPolygonsContainPathsWithinKM(
	polygons, paths [][]GeoPoint,
	closePaths bool,
	toleranceKM float64,
) bool {
	containment := newSphericalPolygonContainment(polygons)
	return visitSphericalPathProbes(paths, closePaths, func(point GeoPoint) bool {
		return containment.missDistanceKM(point) <= toleranceKM
	})
}

// SphericalPolygonsPathMissDistanceKM 返回路径样本到多边形内部或边界的最大偏离距离。
// SphericalPolygonsPathMissDistanceKM returns the greatest distance from a
// path probe outside all polygons to the nearest polygon edge. Vertices and
// minor-great-circle edge midpoints are probed; a fully contained path returns
// zero.
func SphericalPolygonsPathMissDistanceKM(polygons, paths [][]GeoPoint, closePaths bool) float64 {
	containment := newSphericalPolygonContainment(polygons)
	maximumMiss := 0.0
	visitSphericalPathProbes(paths, closePaths, func(point GeoPoint) bool {
		maximumMiss = math.Max(maximumMiss, containment.missDistanceKM(point))
		return true
	})
	return maximumMiss
}

func visitSphericalPathProbes(
	paths [][]GeoPoint,
	closePaths bool,
	visit func(GeoPoint) bool,
) bool {
	for _, path := range paths {
		path = openGeoRing(path)
		for _, point := range path {
			if !visit(point) {
				return false
			}
		}
		edgeCount := len(path) - 1
		if closePaths && len(path) > 1 {
			edgeCount = len(path)
		}
		for index := 0; index < edgeCount; index++ {
			if !visit(sphericalInterpolate(path[index], path[(index+1)%len(path)], 0.5)) {
				return false
			}
		}
	}
	return true
}

type sphericalPolygonContainment struct {
	polygons  [][]GeoPoint
	projected []polygonUnionRing
	center    geoVector3
	xAxis     geoVector3
	yAxis     geoVector3
}

func newSphericalPolygonContainment(polygons [][]GeoPoint) sphericalPolygonContainment {
	result := sphericalPolygonContainment{polygons: polygons}
	center := geoVector3{}
	for _, polygon := range polygons {
		for _, point := range openGeoRing(polygon) {
			center = geoVectorAdd(center, geoPointVector(point))
		}
	}
	var ok bool
	result.center, ok = geoVectorNormalize(center)
	if !ok {
		return result
	}
	reference := geoVector3{z: 1}
	if math.Abs(geoVectorDot(reference, result.center)) > 0.9 {
		reference = geoVector3{x: 1}
	}
	result.xAxis, ok = geoVectorNormalize(geoVectorCross(reference, result.center))
	if !ok {
		return result
	}
	result.yAxis, ok = geoVectorNormalize(geoVectorCross(result.center, result.xAxis))
	if !ok {
		return result
	}
	result.projected = make([]polygonUnionRing, len(polygons))
	for polygonIndex, polygon := range polygons {
		polygon = openGeoRing(polygon)
		points := make([]polygonUnionPoint, len(polygon))
		for pointIndex, point := range polygon {
			projected, projectedOK := result.project(point)
			if !projectedOK {
				result.projected = nil
				return result
			}
			points[pointIndex] = projected
		}
		result.projected[polygonIndex] = polygonUnionRingForPoints(points)
	}
	return result
}

func (containment sphericalPolygonContainment) project(point GeoPoint) (polygonUnionPoint, bool) {
	vector := geoPointVector(point)
	denominator := geoVectorDot(vector, containment.center)
	if denominator <= 1e-12 {
		return polygonUnionPoint{}, false
	}
	return polygonUnionPoint{
		x: geoVectorDot(vector, containment.xAxis) / denominator,
		y: geoVectorDot(vector, containment.yAxis) / denominator,
	}, true
}

func (containment sphericalPolygonContainment) contains(point GeoPoint) bool {
	if len(containment.projected) > 0 {
		if projected, ok := containment.project(point); ok {
			for _, polygon := range containment.projected {
				if polygonUnionRingContainsPoint(projected, polygon) {
					return true
				}
			}
			return false
		}
		// Every ring is inside this gnomonic hemisphere. A point beyond its
		// horizon cannot be inside and must not use antipodal tangent winding.
		return false
	}
	for _, polygon := range containment.polygons {
		if sphericalPolygonContainsOrTouches(polygon, point) {
			return true
		}
	}
	return false
}

func (containment sphericalPolygonContainment) missDistanceKM(point GeoPoint) float64 {
	if containment.contains(point) {
		return 0
	}
	return containment.sphericalMissDistanceKM(point)
}

func (containment sphericalPolygonContainment) sphericalMissDistanceKM(point GeoPoint) float64 {
	nearest := math.Inf(1)
	for _, polygon := range containment.polygons {
		polygon = openGeoRing(polygon)
		for index, start := range polygon {
			end := polygon[(index+1)%len(polygon)]
			nearest = math.Min(nearest, sphericalPointArcDistanceKM(point, start, end))
		}
	}
	return nearest
}

func sphericalPointArcDistanceKM(point, start, end GeoPoint) float64 {
	pointVector := geoPointVector(point)
	startVector := geoPointVector(start)
	endVector := geoPointVector(end)
	normal, ok := geoVectorNormalize(geoVectorCross(startVector, endVector))
	if ok {
		projection := geoVectorAdd(pointVector, geoVectorScale(normal, -geoVectorDot(pointVector, normal)))
		if projected, projectedOK := geoVectorNormalize(projection); projectedOK {
			for _, candidate := range []geoVector3{projected, geoVectorScale(projected, -1)} {
				candidatePoint := geoVectorPoint(candidate)
				if sphericalPointOnArc(candidatePoint, start, end) {
					angle := math.Acos(math.Max(-1, math.Min(1, geoVectorDot(pointVector, candidate))))
					return angle * 6378.1366
				}
			}
		}
	}
	return math.Min(geoPointDistanceKM(point, start), geoPointDistanceKM(point, end))
}

// JoinPolylineSegments 按最近端点连接无序边界线段 / JoinPolylineSegments joins unordered boundary segments by their nearest
// 端点连接；输入线段不会被修改 / endpoints. The input segments are not modified.
func JoinPolylineSegments(segments [][]GeoPoint) []GeoPoint {
	filtered := make([][]GeoPoint, 0, len(segments))
	for _, segment := range segments {
		if len(segment) == 0 {
			continue
		}
		filtered = append(filtered, append([]GeoPoint(nil), segment...))
	}
	if len(filtered) == 0 {
		return nil
	}

	result := append([]GeoPoint(nil), filtered[0]...)
	used := make([]bool, len(filtered))
	used[0] = true
	for joined := 1; joined < len(filtered); joined++ {
		bestIndex := -1
		bestReverse := false
		bestPrepend := false
		bestDistance := math.Inf(1)
		start := result[0]
		end := result[len(result)-1]
		for index, segment := range filtered {
			if used[index] {
				continue
			}
			if distance := angularDistanceDegrees(end, segment[0]); distance < bestDistance {
				bestIndex, bestReverse, bestPrepend, bestDistance = index, false, false, distance
			}
			if distance := angularDistanceDegrees(end, segment[len(segment)-1]); distance < bestDistance {
				bestIndex, bestReverse, bestPrepend, bestDistance = index, true, false, distance
			}
			if distance := angularDistanceDegrees(start, segment[len(segment)-1]); distance < bestDistance {
				bestIndex, bestReverse, bestPrepend, bestDistance = index, false, true, distance
			}
			if distance := angularDistanceDegrees(start, segment[0]); distance < bestDistance {
				bestIndex, bestReverse, bestPrepend, bestDistance = index, true, true, distance
			}
		}
		if bestIndex < 0 {
			break
		}
		segment := filtered[bestIndex]
		if bestReverse {
			reverseGeoPoints(segment)
		}
		if bestPrepend {
			result = appendJoinedGeoPoints(segment, result)
		} else {
			result = appendJoinedGeoPoints(result, segment)
		}
		used[bestIndex] = true
	}
	return result
}

// appendJoinedGeoPoints 拼接两段并丢掉衔接处重合的顶点。
func appendJoinedGeoPoints(first, second []GeoPoint) []GeoPoint {
	result := append([]GeoPoint(nil), first...)
	start := 0
	for start < len(second) && len(result) > 0 && SameGeoPoint(result[len(result)-1], second[start]) {
		start++
	}
	return append(result, second[start:]...)
}

// ShortestCircleArc 返回两点之间较短的采样圆弧 / ShortestCircleArc returns the shorter sampled arc from one point to another.
func ShortestCircleArc(circle []GeoPoint, from, to GeoPoint) []GeoPoint {
	if len(circle) == 0 {
		return nil
	}
	fromIndex := nearestGeoPointIndex(circle, from)
	toIndex := nearestGeoPointIndex(circle, to)
	forwardSteps := (toIndex - fromIndex + len(circle)) % len(circle)
	backwardSteps := (fromIndex - toIndex + len(circle)) % len(circle)
	direction := 1
	steps := forwardSteps
	if backwardSteps < forwardSteps {
		direction = -1
		steps = backwardSteps
	}
	result := make([]GeoPoint, 0, steps+2)
	result = append(result, from)
	for step := 1; step < steps; step++ {
		index := (fromIndex + direction*step) % len(circle)
		if index < 0 {
			index += len(circle)
		}
		result = append(result, circle[index])
	}
	return append(result, to)
}

// SameGeoPoint 判断两个经纬度点是否在拓扑所需精度内相等 / SameGeoPoint reports whether two longitude/latitude points are equal within
// 地图拓扑辅助函数所需的精度内相等 / the precision needed by the map topology helpers.
func SameGeoPoint(a, b GeoPoint) bool {
	return math.Abs(normalizeLongitude(a.Longitude-b.Longitude)) < 1e-9 &&
		math.Abs(a.Latitude-b.Latitude) < 1e-9
}

func nearestGeoPointIndex(points []GeoPoint, target GeoPoint) int {
	bestIndex := 0
	bestDistance := math.Inf(1)
	for index, point := range points {
		if distance := angularDistanceDegrees(point, target); distance < bestDistance {
			bestIndex, bestDistance = index, distance
		}
	}
	return bestIndex
}

func angularDistanceDegrees(a, b GeoPoint) float64 {
	lat1 := a.Latitude * math.Pi / 180
	lat2 := b.Latitude * math.Pi / 180
	dLongitude := normalizeLongitude(b.Longitude-a.Longitude) * math.Pi / 180
	cosine := math.Sin(lat1)*math.Sin(lat2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Cos(dLongitude)
	return math.Acos(math.Max(-1, math.Min(1, cosine))) * 180 / math.Pi
}

func reverseGeoPoints(points []GeoPoint) {
	for left, right := 0, len(points)-1; left < right; left, right = left+1, right-1 {
		points[left], points[right] = points[right], points[left]
	}
}
