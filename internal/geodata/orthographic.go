package geodata

import "math"

// rad 是度到弧度的换算因子 / rad converts degrees to radians.
const rad = math.Pi / 180

const (
	// orthographicRimSteps 是视界闭合弧与整盘回退环的加密段数。
	orthographicRimSteps = 180
	// orthographicCrossingIterations 是视界交点的二分次数，1e-12 弧度量级足够。
	orthographicCrossingIterations = 48
	// orthographicRunCapacity 是单个可见段的初始容量：段长与环长无关，
	// 按环长预分配会让反复穿越视界的环退化成 O(段数×环长) 的内存。
	orthographicRunCapacity = 8
)

// OrthographicDiskPoint 把点正射投影到可见半球的单位圆盘，x 向东、y 向北。
// 第二个返回值是深度余弦；false 表示点落在背面，不与可见半球构成一一映射。
// OrthographicDiskPoint projects a point onto the unit disk of the visible hemisphere, x east and y north.
func OrthographicDiskPoint(point, center GeoPoint) (float64, float64, bool) {
	longitude := (point.Longitude - center.Longitude) * rad
	latitude := point.Latitude * rad
	centerLatitude := center.Latitude * rad
	cosLatitude, sinLatitude := math.Cos(latitude), math.Sin(latitude)
	sinCenter, cosCenter := math.Sin(centerLatitude), math.Cos(centerLatitude)
	cosine := sinCenter*sinLatitude + cosCenter*cosLatitude*math.Cos(longitude)
	// 视界本身（余弦为 0）映射到圆盘边界，必须可投影；只有严格背面才折叠到盘内。
	if cosine < -1e-9 {
		return 0, 0, false
	}
	return cosLatitude * math.Sin(longitude),
		cosCenter*sinLatitude - sinCenter*cosLatitude*math.Cos(longitude), true
}

// orthographicDepth 返回点相对视点的深度余弦，正值表示在可见半球上。
func orthographicDepth(point, center GeoPoint) float64 {
	return geoVectorDot(geoPointVector(point), geoPointVector(center))
}

// orthographicCrossing 二分求线段与视界大圆的交点；两端同侧时返回 false。
func orthographicCrossing(first, second GeoPoint, center GeoPoint) (GeoPoint, bool) {
	firstDepth := orthographicDepth(first, center)
	secondDepth := orthographicDepth(second, center)
	if firstDepth == 0 {
		return first, true
	}
	if secondDepth == 0 {
		return second, true
	}
	if (firstDepth > 0) == (secondDepth > 0) {
		return GeoPoint{}, false
	}
	// 收敛到起点那一侧的边界：可见性仍与起点相同就往后挪，翻转了就往前收。
	firstVisible := firstDepth > 0
	low, high := 0.0, 1.0
	for iteration := 0; iteration < orthographicCrossingIterations; iteration++ {
		middle := (low + high) / 2
		if (orthographicDepth(InterpolateGreatCircle(first, second, middle), center) > 0) == firstVisible {
			low = middle
		} else {
			high = middle
		}
	}
	return InterpolateGreatCircle(first, second, (low+high)/2), true
}

// clipPolylineOrthographic 把折线裁到可见半球，并在视界处插入精确交点。
func clipPolylineOrthographic(points []GeoPoint, center GeoPoint) [][]GeoPoint {
	if len(points) == 0 {
		return nil
	}
	segments := make([][]GeoPoint, 0, 2)
	current := make([]GeoPoint, 0, len(points))
	for index, point := range points {
		if orthographicDepth(point, center) > 0 {
			if len(current) == 0 && index > 0 {
				if crossing, ok := orthographicCrossing(points[index-1], point, center); ok {
					current = append(current, crossing)
				}
			}
			current = append(current, point)
			continue
		}
		if len(current) > 0 {
			if crossing, ok := orthographicCrossing(points[index-1], point, center); ok {
				current = append(current, crossing)
			}
			if len(current) >= 2 {
				segments = append(segments, current)
			}
			current = nil
		}
	}
	if len(current) >= 2 {
		segments = append(segments, current)
	}
	return segments
}

// orthographicRimArc 沿视界大圆从起点加密到终点；long 为 true 时走另一侧的长弧。
// 视界大圆的法线就是视点方向，必须绕它旋转：两端接近对径时 cross(起点, 终点) 会退化成零向量，
// 那样闭合弧会塌成一条横穿圆盘的直线弦。
func orthographicRimArc(from, to GeoPoint, center GeoPoint, long bool) []GeoPoint {
	axis := geoPointVector(center)
	startVector, ok := geoVectorNormalize(geoVectorAdd(
		geoPointVector(from),
		geoVectorScale(axis, -geoVectorDot(geoPointVector(from), axis)),
	))
	if !ok {
		return []GeoPoint{from, to}
	}
	tangent := geoVectorCross(axis, startVector)
	endVector := geoPointVector(to)
	signed := math.Atan2(geoVectorDot(endVector, tangent), geoVectorDot(endVector, startVector))
	begin, span := 0.0, signed
	if long {
		turn := 2 * math.Pi
		if signed < 0 {
			turn = -2 * math.Pi
		}
		begin, span = signed, turn-signed
	}
	arc := make([]GeoPoint, 0, orthographicRimSteps+1)
	for step := 0; step <= orthographicRimSteps; step++ {
		angle := begin + span*float64(step)/orthographicRimSteps
		arc = append(arc, geoVectorPoint(geoVectorAdd(
			geoVectorScale(startVector, math.Cos(angle)),
			geoVectorScale(tangent, math.Sin(angle)),
		)))
	}
	return arc
}

// orthographicRimInside 判断某段视界弧是否紧邻环的内部：把弧中点朝可见半球内侧挪一点再看它落在哪一侧。
func orthographicRimInside(arc []GeoPoint, ring []GeoPoint, center GeoPoint) bool {
	if len(arc) == 0 {
		return false
	}
	middle := geoPointVector(arc[len(arc)/2])
	inside := geoVectorAdd(middle, geoVectorScale(geoPointVector(center), 1e-3))
	probe, ok := geoVectorNormalize(inside)
	if !ok {
		return false
	}
	return sphericalPolygonContainsOrTouches(ring, geoVectorPoint(probe))
}

// closeOrthographicRun 把一段可见折线沿视界大圆闭合回起点，闭合弧取紧邻环内部的那一侧。
func closeOrthographicRun(run, ring []GeoPoint, center GeoPoint) []GeoPoint {
	if len(run) < 2 {
		return nil
	}
	exit, entry := run[len(run)-1], run[0]
	shortArc := orthographicRimArc(exit, entry, center, false)
	longArc := orthographicRimArc(exit, entry, center, true)
	arc := shortArc
	switch {
	case orthographicRimInside(shortArc, ring, center):
	case orthographicRimInside(longArc, ring, center):
		arc = longArc
	}
	closed := make([]GeoPoint, 0, len(run)+len(arc))
	closed = append(closed, run...)
	closed = append(closed, arc[1:len(arc)-1]...)
	return closed
}

// polygonFragmentsOrthographic 把环裁到可见半球，并沿视界大圆闭合被切断的部分。
func polygonFragmentsOrthographic(points []GeoPoint, center GeoPoint) [][]GeoPoint {
	if len(points) < 3 {
		return nil
	}
	visible := 0
	for _, point := range points {
		if orthographicDepth(point, center) > 0 {
			visible++
		}
	}
	if visible == len(points) {
		return [][]GeoPoint{points}
	}
	if visible == 0 {
		// 整环都在背面：只有把视点包在环内的环，其内部才会覆盖整个可见半球——否则可见部分为空。
		if sphericalPolygonContainsOrTouches(points, center) {
			return [][]GeoPoint{SphericalCircle(center, 90, orthographicRimSteps)}
		}
		return nil
	}
	// 逐边展开成"顶点 + 视界交点"序列，再按可见性切段；闭合环首尾相接，所以按环遍历。
	type rimNode struct {
		point   GeoPoint
		visible bool
	}
	nodes := make([]rimNode, 0, 2*len(points))
	for index := 0; index < len(points); index++ {
		first := points[index]
		second := points[(index+1)%len(points)]
		firstVisible := orthographicDepth(first, center) > 0
		secondVisible := orthographicDepth(second, center) > 0
		nodes = append(nodes, rimNode{point: first, visible: firstVisible})
		if firstVisible != secondVisible {
			if crossing, ok := orthographicCrossing(first, second, center); ok {
				// 交点落在视界上，两侧的可见段都要以它收尾/起头，所以它恒属于可见段。
				nodes = append(nodes, rimNode{point: crossing, visible: true})
			}
		}
	}
	runs := make([][]GeoPoint, 0, 4)
	current := make([]GeoPoint, 0, orthographicRunCapacity)
	for _, node := range nodes {
		if node.visible {
			current = append(current, node.point)
			continue
		}
		if len(current) >= 2 {
			runs = append(runs, current)
		}
		current = make([]GeoPoint, 0, orthographicRunCapacity)
	}
	if len(current) >= 2 {
		runs = append(runs, current)
	}
	// 环首尾相接：起点本身可见时，同一段可见区间会被切成首尾两段，必须先接回来再闭合。
	if len(runs) >= 2 && nodes[0].visible && nodes[len(nodes)-1].visible {
		merged := make([]GeoPoint, 0, len(runs[0])+len(runs[len(runs)-1]))
		merged = append(merged, runs[len(runs)-1]...)
		merged = append(merged, runs[0]...)
		runs[0] = merged
		runs = runs[:len(runs)-1]
	}
	fragments := make([][]GeoPoint, 0, len(runs))
	for _, run := range runs {
		if closed := closeOrthographicRun(run, points, center); len(closed) >= 3 {
			fragments = append(fragments, closed)
		}
	}
	return fragments
}
