package geodata

import "math"

// SphericalCircle 返回球面小圆上的等间隔采样点 / SphericalCircle returns evenly spaced points on a small circle on the
// 球面小圆；方位角从地理北方顺时针采样 / sphere. Bearings are sampled clockwise from geographic north.
func SphericalCircle(center GeoPoint, radiusDegrees float64, points int) []GeoPoint {
	if points < 3 {
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
			result = append(segment, result...)
		} else {
			result = append(result, segment...)
		}
		used[bestIndex] = true
	}
	return result
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

// VisibleHemispherePolygons 返回以指定中心为中心的半球多边形 / VisibleHemispherePolygons returns polygons for the hemisphere centered on
// 中心的半球多边形，并裁剪到请求的地图投影 / center, clipped to the requested map projection.
func VisibleHemispherePolygons(center GeoPoint, projection Projection, samples int) [][]GeoPoint {
	if samples < 12 {
		samples = 12
	}
	if projection == ProjectionNorthPolar {
		return [][]GeoPoint{polarVisibleHemispherePolygon(center, 1, samples/2)}
	}
	if projection == ProjectionSouthPolar {
		return [][]GeoPoint{polarVisibleHemispherePolygon(center, -1, samples/2)}
	}
	return equirectangularVisibleHemispherePolygons(center, samples)
}

func equirectangularVisibleHemispherePolygons(center GeoPoint, samples int) [][]GeoPoint {
	if math.Abs(center.Latitude) < 1e-9 {
		return equirectangularLongitudeBand(center.Longitude)
	}
	polygon := make([]GeoPoint, 0, samples+3)
	for index := 0; index <= samples; index++ {
		longitude := -180 + 360*float64(index)/float64(samples)
		polygon = append(polygon, GeoPoint{
			Longitude: longitude,
			Latitude:  visibleHorizonLatitude(center, longitude),
		})
	}
	mapEdgeLatitude := math.Copysign(90, center.Latitude)
	return [][]GeoPoint{append(polygon,
		GeoPoint{Longitude: 180, Latitude: mapEdgeLatitude},
		GeoPoint{Longitude: -180, Latitude: mapEdgeLatitude},
	)}
}

func equirectangularLongitudeBand(centerLongitude float64) [][]GeoPoint {
	centerLongitude = normalizeLongitude(centerLongitude)
	start, end := centerLongitude-90, centerLongitude+90
	var polygons [][]GeoPoint
	for _, shift := range []float64{-360, 0, 360} {
		left := math.Max(-180, start+shift)
		right := math.Min(180, end+shift)
		if right-left <= 1e-9 {
			continue
		}
		polygons = append(polygons, []GeoPoint{
			{Longitude: left, Latitude: -90},
			{Longitude: right, Latitude: -90},
			{Longitude: right, Latitude: 90},
			{Longitude: left, Latitude: 90},
		})
	}
	return polygons
}

func polarVisibleHemispherePolygon(center GeoPoint, hemisphere float64, samples int) []GeoPoint {
	if samples < 6 {
		samples = 6
	}
	if math.Abs(center.Latitude) < 1e-9 {
		polygon := []GeoPoint{
			{Longitude: normalizeLongitude(center.Longitude - 90), Latitude: 0},
			{Longitude: normalizeLongitude(center.Longitude), Latitude: 90 * hemisphere},
			{Longitude: normalizeLongitude(center.Longitude + 90), Latitude: 0},
		}
		return appendPolarVisibilityRim(polygon, center.Longitude, false, samples)
	}

	centerLongitude := normalizeLongitude(center.Longitude)
	centerInsideProjection := center.Latitude*hemisphere > 0
	horizonMidpoint := centerLongitude
	if centerInsideProjection {
		horizonMidpoint += 180
	}
	polygon := make([]GeoPoint, 0, 2*samples+1)
	for index := 0; index <= samples; index++ {
		longitude := horizonMidpoint - 90 + 180*float64(index)/float64(samples)
		latitude := visibleHorizonLatitude(center, longitude)
		if latitude*hemisphere < 0 && math.Abs(latitude) < 1e-9 {
			latitude = 0
		}
		polygon = append(polygon, GeoPoint{
			Longitude: normalizeLongitude(longitude),
			Latitude:  latitude,
		})
	}
	return appendPolarVisibilityRim(polygon, centerLongitude, centerInsideProjection, samples)
}

func appendPolarVisibilityRim(
	polygon []GeoPoint,
	centerLongitude float64,
	centerInsideProjection bool,
	samples int,
) []GeoPoint {
	for index := 1; index <= samples; index++ {
		fraction := float64(index) / float64(samples)
		longitude := centerLongitude + 90 - 180*fraction
		if centerInsideProjection {
			longitude = centerLongitude - 90 + 180*fraction
		}
		polygon = append(polygon, GeoPoint{
			Longitude: normalizeLongitude(longitude),
			Latitude:  0,
		})
	}
	return polygon
}

func visibleHorizonLatitude(center GeoPoint, longitude float64) float64 {
	declination := center.Latitude * math.Pi / 180
	deltaLongitude := (longitude - center.Longitude) * math.Pi / 180
	return math.Atan(-math.Cos(declination)*math.Cos(deltaLongitude)/math.Sin(declination)) * 180 / math.Pi
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
