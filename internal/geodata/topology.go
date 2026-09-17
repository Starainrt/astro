package geodata

import "math"

// PolylineSegments 将地理折线裁剪到选定投影并 / PolylineSegments clips a geographic polyline to the selected projection and
// 在等经纬投影中按日界线拆分路径 / splits equirectangular paths at the antimeridian.
func PolylineSegments(points []GeoPoint, view ClipView) [][]GeoPoint {
	prepared, ok := prepareTopologyPoints(points)
	if !ok {
		return nil
	}
	points = prepared
	if view.Orthographic() {
		return clipPolylineOrthographic(points, view.Center)
	}
	projection := view.Projection
	if projection == ProjectionNorthPolar {
		return clipPolylineHemisphere(points, 1)
	}
	if projection == ProjectionSouthPolar {
		return clipPolylineHemisphere(points, -1)
	}
	if shift := equirectangularSeamShift(view); shift != 0 {
		return splitPolylineAtSeam(points, shift)
	}
	return splitPolylineAntimeridian(points)
}

// PolygonFragments 将地理多边形裁剪到选定地图范围 / PolygonFragments clips a geographic polygon to the selected map extent.
func PolygonFragments(points []GeoPoint, view ClipView) [][]GeoPoint {
	if len(points) < 3 {
		return nil
	}
	prepared, ok := prepareTopologyPoints(points)
	if !ok {
		return nil
	}
	points = prepared
	if view.Orthographic() {
		return polygonFragmentsOrthographic(points, view.Center)
	}
	projection := view.Projection
	if projection == ProjectionNorthPolar {
		if clipped := clipPolygonHemisphere(points, 1); len(clipped) >= 3 {
			return [][]GeoPoint{clipped}
		}
		return nil
	}
	if projection == ProjectionSouthPolar {
		if clipped := clipPolygonHemisphere(points, -1); len(clipped) >= 3 {
			return [][]GeoPoint{clipped}
		}
		return nil
	}
	if shift := equirectangularSeamShift(view); shift != 0 {
		return splitPolygonAtSeam(points, shift)
	}
	return splitPolygonAntimeridian(points)
}

// prepareTopologyPoints 校验裁剪输入：NaN 与 ±Inf 直接拒绝，经度超出 ±180 时按 360 取模归一化
// （同一子午线的等价表示），只有真的越界才复制，正常输入保持零分配。
func prepareTopologyPoints(points []GeoPoint) ([]GeoPoint, bool) {
	for _, point := range points {
		if math.IsNaN(point.Longitude) || math.IsNaN(point.Latitude) ||
			math.IsInf(point.Longitude, 0) || math.IsInf(point.Latitude, 0) {
			return nil, false
		}
	}
	shifted := false
	for _, point := range points {
		if point.Longitude < -180 || point.Longitude > 180 {
			shifted = true
			break
		}
	}
	if !shifted {
		return points, true
	}
	normalized := make([]GeoPoint, len(points))
	copy(normalized, points)
	for index := range normalized {
		normalized[index].Longitude = normalizeLongitude(normalized[index].Longitude)
	}
	return normalized, true
}

func splitPolylineAntimeridian(points []GeoPoint) [][]GeoPoint {
	if len(points) == 0 {
		return nil
	}
	segments := make([][]GeoPoint, 0, 2)
	current := []GeoPoint{points[0]}
	for index := 1; index < len(points); index++ {
		a, b := points[index-1], points[index]
		if (a.Longitude == -180 && b.Longitude == 180) || (a.Longitude == 180 && b.Longitude == -180) {
			// -180/+180 的精确端点属于同一子午线，不要 / Exact -180/+180 endpoints are the same meridian; do not
			// 不要让它们进入分母为零的交叉插值 / feed them into the crossing interpolation with a zero denominator.
			b.Longitude = a.Longitude
			current = append(current, b)
			continue
		}
		if math.Abs(b.Longitude-a.Longitude) <= 180 {
			current = append(current, b)
			continue
		}
		boundary := 180.0
		adjustedLongitude := b.Longitude
		if a.Longitude < 0 {
			boundary = -180
			adjustedLongitude -= 360
		} else {
			adjustedLongitude += 360
		}
		crossing := sphericalLongitudeIntersection(a, b, boundary, adjustedLongitude)
		current = append(current, crossing)
		if len(current) >= 2 {
			segments = append(segments, current)
		}
		crossing.Longitude = -boundary
		current = []GeoPoint{crossing, b}
	}
	if len(current) >= 2 {
		segments = append(segments, current)
	}
	return segments
}

func clipPolylineHemisphere(points []GeoPoint, hemisphere float64) [][]GeoPoint {
	if len(points) == 0 {
		return nil
	}
	inside := func(value GeoPoint) bool { return value.Latitude*hemisphere >= 0 }
	var segments [][]GeoPoint
	var current []GeoPoint
	for index, point := range points {
		pointInside := inside(point)
		if index == 0 {
			if pointInside {
				current = append(current, point)
			}
			continue
		}
		previous := points[index-1]
		previousInside := inside(previous)
		if previousInside != pointInside {
			crossing := hemisphereIntersection(previous, point)
			if previousInside {
				current = append(current, crossing)
				if len(current) >= 2 {
					segments = append(segments, current)
				}
				current = nil
			} else {
				current = []GeoPoint{crossing}
			}
		}
		if pointInside {
			current = append(current, point)
		}
	}
	if len(current) >= 2 {
		segments = append(segments, current)
	}
	return segments
}

func clipPolygonHemisphere(points []GeoPoint, hemisphere float64) []GeoPoint {
	inside := func(value GeoPoint) bool { return value.Latitude*hemisphere >= 0 }
	return clipPolygon(points, inside, hemisphereIntersection)
}

func splitPolygonAntimeridian(points []GeoPoint) [][]GeoPoint {
	unwrapped := make([]GeoPoint, len(points))
	unwrapped[0] = points[0]
	for index := 1; index < len(points); index++ {
		point := points[index]
		previous := unwrapped[index-1].Longitude
		// 一次取整到最近的 360 倍数，等价于反复加减 360 但不随偏移量增长。
		if offset := point.Longitude - previous; offset > 180 || offset < -180 {
			point.Longitude -= 360 * math.Round(offset/360)
		}
		unwrapped[index] = point
	}
	unwrapped = closePoleEnclosingPolygon(unwrapped)
	minimum, maximum := unwrapped[0].Longitude, unwrapped[0].Longitude
	for _, point := range unwrapped[1:] {
		minimum = math.Min(minimum, point.Longitude)
		maximum = math.Max(maximum, point.Longitude)
	}
	if span := maximum - minimum; !(span >= 0 && span <= 3*360) {
		// 经度归一化后每个点最多偏离 ±180 再加一次 360 的展开，跨度不可能超过三个世界。
		return nil
	}
	firstWorld := int(math.Floor((minimum + 180) / 360))
	lastWorld := int(math.Floor((maximum + 180) / 360))
	var fragments [][]GeoPoint
	for world := firstWorld; world <= lastWorld; world++ {
		left := -180.0 + 360*float64(world)
		right := 180.0 + 360*float64(world)
		clipped := clipPolygonLongitude(unwrapped, left, true)
		clipped = clipPolygonLongitude(clipped, right, false)
		if len(clipped) < 3 {
			continue
		}
		for index := range clipped {
			clipped[index].Longitude -= 360 * float64(world)
		}
		if math.Abs(signedPolygonArea(clipped)) < 1e-12 {
			// 边恰好落在日界线上的多边形可能在相邻世界各输出一次 / A polygon whose edge lies exactly on the antimeridian can be
			// 可能在相邻世界各输出一次；丢弃重复的 / emitted once for each adjacent world. Drop the duplicate
			// 零面积片段后再做 GeoJSON 环验证 / zero-area fragment before GeoJSON ring validation.
			continue
		}
		fragments = append(fragments, clipped)
	}
	return fragments
}

// closePoleEnclosingPolygon adds the equirectangular map-edge closure for a
// simple spherical ring that winds once around a pole. Without this edge, the
// implicit last-to-first segment cuts across the map instead of representing
// the cap at +90 or -90 degrees.
func closePoleEnclosingPolygon(points []GeoPoint) []GeoPoint {
	if len(points) < 3 {
		return points
	}
	first := points[0].Longitude
	closure := first
	last := points[len(points)-1].Longitude
	for closure-last > 180 {
		closure -= 360
	}
	for closure-last < -180 {
		closure += 360
	}
	winding := math.Round((closure - first) / 360)
	if math.Abs(winding) != 1 {
		return points
	}
	// Choose the smaller map-edge closure from edge geometry. Unlike a vertex
	// average, its result is unchanged when a straight boundary edge is resampled.
	north := appendPoleClosure(points, closure, first, 90)
	south := appendPoleClosure(points, closure, first, -90)
	northInside := sphericalPolygonContainsOrTouches(points, GeoPoint{Longitude: first, Latitude: 90})
	southInside := sphericalPolygonContainsOrTouches(points, GeoPoint{Longitude: first, Latitude: -90})
	northArea := math.Abs(signedPolygonArea(north))
	southArea := math.Abs(signedPolygonArea(south))
	if northInside == southInside && math.Abs(northArea-southArea) <= 1e-12 {
		return points
	}
	pole := -90.0
	if (northInside != southInside && northInside) || (northInside == southInside && northArea < southArea) {
		pole = 90
	}
	// The closure must not cross the physical boundary again. For a concave
	// polar ring, only the poleward-most seam crossing has a clear path to
	// the pole; the first crossing can turn an excluded pocket into a fill.
	points = rotatePoleRingToMapEdge(points, pole)
	first = points[0].Longitude
	return appendPoleClosure(points, first+360*winding, first, pole)
}

func rotatePoleRingToMapEdge(points []GeoPoint, poleLatitude float64) []GeoPoint {
	if len(points) < 3 {
		return points
	}
	bestIndex := -1
	bestScore := math.Inf(-1)
	var crossing GeoPoint
	for index := 0; index < len(points); index++ {
		next := (index + 1) % len(points)
		a, b := points[index], points[next]
		bLongitude := b.Longitude
		for bLongitude-a.Longitude > 180 {
			bLongitude -= 360
		}
		for bLongitude-a.Longitude < -180 {
			bLongitude += 360
		}
		if math.Abs(a.Longitude-bLongitude) <= 1e-12 {
			continue
		}
		target := 180 + 360*math.Ceil((math.Min(a.Longitude, bLongitude)-180-1e-12)/360)
		for ; target <= math.Max(a.Longitude, bLongitude)+1e-12; target += 360 {
			candidate := sphericalLongitudeIntersection(a, b, target, bLongitude)
			if score := candidate.Latitude * poleLatitude; score > bestScore {
				bestIndex, bestScore, crossing = index, score, candidate
			}
		}
	}
	if bestIndex < 0 {
		return points
	}
	rotated := make([]GeoPoint, 1, len(points)+1)
	rotated[0] = crossing
	for offset := 1; offset <= len(points); offset++ {
		point := points[(bestIndex+offset)%len(points)]
		previous := rotated[len(rotated)-1].Longitude
		for point.Longitude-previous > 180 {
			point.Longitude -= 360
		}
		for point.Longitude-previous < -180 {
			point.Longitude += 360
		}
		rotated = append(rotated, point)
	}
	return sweepDeduplicateAdjacent(rotated)
}

func appendPoleClosure(points []GeoPoint, closure, first, poleLatitude float64) []GeoPoint {
	result := append([]GeoPoint(nil), points...)
	// The unwrapped ring ends in the world adjacent to its first vertex. Repeat
	// that vertex in the adjacent world before climbing to the map edge;
	// otherwise the last boundary point is connected diagonally to the pole and
	// a triangular gap is cut out after antimeridian clipping.
	result = append(result, GeoPoint{
		Longitude: closure,
		Latitude:  points[0].Latitude,
	})
	return append(result,
		GeoPoint{Longitude: closure, Latitude: poleLatitude},
		GeoPoint{Longitude: first, Latitude: poleLatitude},
	)
}

func signedPolygonArea(points []GeoPoint) float64 {
	if len(points) < 3 {
		return 0
	}
	area := 0.0
	for index, point := range points {
		next := points[(index+1)%len(points)]
		area += point.Longitude*next.Latitude - next.Longitude*point.Latitude
	}
	return area / 2
}

func clipPolygonLongitude(points []GeoPoint, boundary float64, keepGreater bool) []GeoPoint {
	inside := func(value GeoPoint) bool {
		if keepGreater {
			return value.Longitude >= boundary
		}
		return value.Longitude <= boundary
	}
	intersection := func(a, b GeoPoint) GeoPoint {
		return sphericalLongitudeIntersection(a, b, boundary, b.Longitude)
	}
	return clipPolygon(points, inside, intersection)
}

func clipPolygon(
	points []GeoPoint,
	inside func(GeoPoint) bool,
	intersection func(GeoPoint, GeoPoint) GeoPoint,
) []GeoPoint {
	if len(points) == 0 {
		return nil
	}
	result := make([]GeoPoint, 0, len(points)+2)
	previous := points[len(points)-1]
	previousInside := inside(previous)
	for _, current := range points {
		currentInside := inside(current)
		if currentInside != previousInside {
			result = append(result, intersection(previous, current))
		}
		if currentInside {
			result = append(result, current)
		}
		previous = current
		previousInside = currentInside
	}
	return result
}

func hemisphereIntersection(a, b GeoPoint) GeoPoint {
	first := a
	second := b
	firstVector := geoPointVector(first)
	secondVector := geoPointVector(second)
	firstSign := firstVector.z
	secondSign := secondVector.z
	if math.Abs(firstSign) <= 1e-15 {
		return GeoPoint{Longitude: normalizeLongitude(first.Longitude), Latitude: 0}
	}
	if math.Abs(secondSign) <= 1e-15 {
		return GeoPoint{Longitude: normalizeLongitude(second.Longitude), Latitude: 0}
	}
	left, right := 0.0, 1.0
	for iteration := 0; iteration < 64; iteration++ {
		middle := (left + right) / 2
		point := sphericalInterpolate(first, second, middle)
		if point.Latitude == 0 || right-left <= 1e-13 {
			return GeoPoint{Longitude: normalizeLongitude(point.Longitude), Latitude: 0}
		}
		if point.Latitude*first.Latitude > 0 {
			left = middle
		} else {
			right = middle
		}
	}
	point := sphericalInterpolate(first, second, (left+right)/2)
	return GeoPoint{Longitude: normalizeLongitude(point.Longitude), Latitude: 0}
}

func sphericalLongitudeIntersection(a, b GeoPoint, boundary, adjustedLongitude float64) GeoPoint {
	firstLongitude := a.Longitude
	secondLongitude := adjustedLongitude
	if math.Abs(secondLongitude-firstLongitude) <= 1e-14 {
		return GeoPoint{Longitude: boundary, Latitude: a.Latitude}
	}
	left, right := 0.0, 1.0
	for iteration := 0; iteration < 64; iteration++ {
		middle := (left + right) / 2
		point := sphericalInterpolate(a, b, middle)
		middleLongitude := point.Longitude
		for middleLongitude-firstLongitude > 180 {
			middleLongitude -= 360
		}
		for middleLongitude-firstLongitude < -180 {
			middleLongitude += 360
		}
		if math.Abs(middleLongitude-boundary) <= 1e-12 || right-left <= 1e-13 {
			return GeoPoint{Longitude: boundary, Latitude: point.Latitude}
		}
		if (firstLongitude-boundary)*(middleLongitude-boundary) <= 0 {
			right = middle
		} else {
			left = middle
		}
	}
	point := sphericalInterpolate(a, b, (left+right)/2)
	return GeoPoint{Longitude: boundary, Latitude: point.Latitude}
}

func normalizeLongitude(value float64) float64 {
	value = math.Mod(value+180, 360)
	if value < 0 {
		value += 360
	}
	return value - 180
}

// equirectangularSeamShift 返回把等经纬接缝从 ±180 搬到视图中心对面所需的经度旋转量。
// 视图中心为 0 时返回 0，调用方即可沿用未旋转的原有路径，保证既有输出逐字节不变。
func equirectangularSeamShift(view ClipView) float64 {
	if view.Projection != ProjectionEquirectangular || view.Center.Longitude == 0 {
		return 0
	}
	return view.Center.Longitude
}

// splitPolylineAtSeam 先把经度旋转到接缝落在 ±180 的坐标系，交给原有分割逻辑，再旋转回来。
func splitPolylineAtSeam(points []GeoPoint, shift float64) [][]GeoPoint {
	segments := splitPolylineAntimeridian(rotateLongitudes(points, -shift))
	for _, segment := range segments {
		rotateLongitudesInPlace(segment, shift)
	}
	return segments
}

// splitPolygonAtSeam 与 splitPolylineAtSeam 同理，供多边形使用。
func splitPolygonAtSeam(points []GeoPoint, shift float64) [][]GeoPoint {
	fragments := splitPolygonAntimeridian(rotateLongitudes(points, -shift))
	for _, fragment := range fragments {
		rotateLongitudesInPlace(fragment, shift)
	}
	return fragments
}

func rotateLongitudes(points []GeoPoint, shift float64) []GeoPoint {
	rotated := make([]GeoPoint, len(points))
	copy(rotated, points)
	rotateLongitudesInPlace(rotated, shift)
	return rotated
}

func rotateLongitudesInPlace(points []GeoPoint, shift float64) {
	for index := range points {
		points[index].Longitude = normalizeLongitude180(points[index].Longitude + shift)
	}
}

func normalizeLongitude180(longitude float64) float64 {
	value := math.Mod(longitude+180, 360)
	if value < 0 {
		value += 360
	}
	return value - 180
}
