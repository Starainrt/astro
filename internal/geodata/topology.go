package geodata

import "math"

// PolylineSegments 将地理折线裁剪到选定投影并 / PolylineSegments clips a geographic polyline to the selected projection and
// 在等经纬投影中按日界线拆分路径 / splits equirectangular paths at the antimeridian.
func PolylineSegments(points []GeoPoint, projection Projection) [][]GeoPoint {
	if projection == ProjectionNorthPolar {
		return clipPolylineHemisphere(points, 1)
	}
	if projection == ProjectionSouthPolar {
		return clipPolylineHemisphere(points, -1)
	}
	return splitPolylineAntimeridian(points)
}

// PolygonFragments 将地理多边形裁剪到选定地图范围 / PolygonFragments clips a geographic polygon to the selected map extent.
func PolygonFragments(points []GeoPoint, projection Projection) [][]GeoPoint {
	if len(points) < 3 {
		return nil
	}
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
	return splitPolygonAntimeridian(points)
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
		fraction := (boundary - a.Longitude) / (adjustedLongitude - a.Longitude)
		crossing := GeoPoint{Longitude: boundary, Latitude: a.Latitude + fraction*(b.Latitude-a.Latitude)}
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
		for point.Longitude-previous > 180 {
			point.Longitude -= 360
		}
		for point.Longitude-previous < -180 {
			point.Longitude += 360
		}
		unwrapped[index] = point
	}
	minimum, maximum := unwrapped[0].Longitude, unwrapped[0].Longitude
	for _, point := range unwrapped[1:] {
		minimum = math.Min(minimum, point.Longitude)
		maximum = math.Max(maximum, point.Longitude)
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
		fraction := (boundary - a.Longitude) / (b.Longitude - a.Longitude)
		return GeoPoint{Longitude: boundary, Latitude: a.Latitude + fraction*(b.Latitude-a.Latitude)}
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
	dLongitude := normalizeLongitude(b.Longitude - a.Longitude)
	fraction := -a.Latitude / (b.Latitude - a.Latitude)
	return GeoPoint{Longitude: normalizeLongitude(a.Longitude + fraction*dLongitude), Latitude: 0}
}

func normalizeLongitude(value float64) float64 {
	value = math.Mod(value+180, 360)
	if value < 0 {
		value += 360
	}
	return value - 180
}
