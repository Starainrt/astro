package svg

import (
	"strconv"
	"strings"

	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/moon"
)

// occultationScreenPoint 是投影到 SVG 画布后的像素坐标。
type occultationScreenPoint struct {
	x float64
	y float64
}

// occultationGeometryTolerancePixels 是掩带几何写出前的折线简化容差（像素）；0 表示不简化。
var occultationGeometryTolerancePixels = 0.4

// projectOccultationRing 投影并简化一个地理环。
func (layout starOccultationSVGLayout) projectOccultationRing(points []starOccultationGeoPoint) []occultationScreenPoint {
	screen := make([]occultationScreenPoint, len(points))
	for index, point := range points {
		x, y := layout.project(point.longitude, point.latitude)
		screen[index] = occultationScreenPoint{x: x, y: y}
	}
	return simplifyOccultationRing(screen, occultationGeometryTolerancePixels)
}

// projectOccultationPolyline 投影并简化一条地理折线。
func (layout starOccultationSVGLayout) projectOccultationPolyline(points []starOccultationGeoPoint) []occultationScreenPoint {
	screen := make([]occultationScreenPoint, len(points))
	for index, point := range points {
		x, y := layout.project(point.longitude, point.latitude)
		screen[index] = occultationScreenPoint{x: x, y: y}
	}
	return simplifyOccultationPolyline(screen, occultationGeometryTolerancePixels)
}

// projectOccultationPathLine 投影并简化一条月掩路径折线。
func (layout starOccultationSVGLayout) projectOccultationPathLine(points []moon.OccultationPathPoint) []occultationScreenPoint {
	screen := make([]occultationScreenPoint, len(points))
	for index, point := range points {
		x, y := layout.project(point.Longitude, point.Latitude)
		screen[index] = occultationScreenPoint{x: x, y: y}
	}
	return simplifyOccultationPolyline(screen, occultationGeometryTolerancePixels)
}

// projectOccultationGeoRing 投影并简化一条地图多边形环。
func (layout starOccultationSVGLayout) projectOccultationGeoRing(points []geodata.GeoPoint) []occultationScreenPoint {
	screen := make([]occultationScreenPoint, len(points))
	for index, point := range points {
		x, y := layout.project(point.Longitude, point.Latitude)
		screen[index] = occultationScreenPoint{x: x, y: y}
	}
	return simplifyOccultationRing(screen, occultationGeometryTolerancePixels)
}

// simplifyOccultationRing 简化闭合环：首点与离首点最远的顶点保留为两个锚点，两段各自简化。
// 简化后不足三个顶点时返回原环，保证结果仍可填充。
func simplifyOccultationRing(points []occultationScreenPoint, tolerance float64) []occultationScreenPoint {
	if len(points) < 4 || !(tolerance > 0) {
		return points
	}
	ring := points
	closed := points[0] == points[len(points)-1]
	if closed {
		ring = points[:len(points)-1]
	}
	if len(ring) < 3 {
		return points
	}
	anchor := 1
	farthest := -1.0
	for index := 1; index < len(ring); index++ {
		distance := occultationSquaredDistance(ring[0], ring[index])
		if distance > farthest {
			farthest, anchor = distance, index
		}
	}
	first := simplifyOccultationPolyline(ring[:anchor+1], tolerance)
	second := make([]occultationScreenPoint, len(ring)-anchor+1)
	copy(second, ring[anchor:])
	second[len(second)-1] = ring[0]
	second = simplifyOccultationPolyline(second, tolerance)
	result := make([]occultationScreenPoint, 0, len(first)+len(second)-2)
	result = append(result, first[:len(first)-1]...)
	result = append(result, second[:len(second)-1]...)
	if len(result) < 3 {
		return points
	}
	if closed {
		result = append(result, result[0])
	}
	return result
}

// simplifyOccultationPolyline 用 Douglas-Peucker 简化折线，端点恒保留。
func simplifyOccultationPolyline(points []occultationScreenPoint, tolerance float64) []occultationScreenPoint {
	if len(points) < 3 || !(tolerance > 0) {
		return points
	}
	keep := make([]bool, len(points))
	keep[0], keep[len(points)-1] = true, true
	toleranceSquared := tolerance * tolerance
	type span struct{ first, last int }
	stack := make([]span, 0, 16)
	stack = append(stack, span{first: 0, last: len(points) - 1})
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if current.last <= current.first+1 {
			continue
		}
		index, distance := occultationFarthestVertex(points, current.first, current.last)
		if distance <= toleranceSquared {
			continue
		}
		keep[index] = true
		stack = append(stack, span{first: current.first, last: index}, span{first: index, last: current.last})
	}
	result := make([]occultationScreenPoint, 0, len(points))
	for index, point := range points {
		if keep[index] {
			result = append(result, point)
		}
	}
	return result
}

// occultationFarthestVertex 返回首末顶点之间离弦最远的顶点及其平方距离。
func occultationFarthestVertex(points []occultationScreenPoint, first, last int) (int, float64) {
	best, bestDistance := first, -1.0
	start, end := points[first], points[last]
	for index := first + 1; index < last; index++ {
		distance := occultationChordSquaredDistance(points[index], start, end)
		if distance > bestDistance {
			best, bestDistance = index, distance
		}
	}
	return best, bestDistance
}

// occultationChordSquaredDistance 返回点到线段的最短平方距离；线段退化时退化为点距。
func occultationChordSquaredDistance(point, start, end occultationScreenPoint) float64 {
	dx, dy := end.x-start.x, end.y-start.y
	lengthSquared := dx*dx + dy*dy
	if lengthSquared <= 0 {
		return occultationSquaredDistance(point, start)
	}
	position := ((point.x-start.x)*dx + (point.y-start.y)*dy) / lengthSquared
	if position < 0 {
		position = 0
	} else if position > 1 {
		position = 1
	}
	return occultationSquaredDistance(point, occultationScreenPoint{x: start.x + position*dx, y: start.y + position*dy})
}

func occultationSquaredDistance(first, second occultationScreenPoint) float64 {
	dx, dy := first.x-second.x, first.y-second.y
	return dx*dx + dy*dy
}

// writeOccultationPathCommands 写出 `M x y L x y ...` 命令序列，坐标保留三位小数。
func writeOccultationPathCommands(b *strings.Builder, points []occultationScreenPoint) {
	var xBuffer, yBuffer [32]byte
	for index, point := range points {
		if index == 0 {
			b.WriteString("M ")
		} else {
			b.WriteString("L ")
		}
		b.Write(strconv.AppendFloat(xBuffer[:0], point.x, 'f', 3, 64))
		b.WriteByte(' ')
		b.Write(strconv.AppendFloat(yBuffer[:0], point.y, 'f', 3, 64))
		b.WriteByte(' ')
	}
}
