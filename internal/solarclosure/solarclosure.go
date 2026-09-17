// Package solarclosure 把被地平线切断的日食偏食足迹闭合到地平圈，并把零食分包络与
// 日升日落相位线并集成一块可见域。
//
// Package solarclosure closes horizon-cut solar-eclipse footprints onto the horizon and
// unions the zero-magnitude envelope with the rise/set phase lines into one visibility region.
package solarclosure

import (
	"math"

	"github.com/starainrt/astro/internal/geodata"
)

// Footprint 是一个瞬时足迹的闭合输入。
type Footprint struct {
	// Boundaries 是物理边界分段；反经线或无效投影会拆成多段。
	Boundaries [][]geodata.GeoPoint
	// HorizonEnds 是未闭合边界两端的地平擦地点，顺序任意，使用时按 Boundaries 走向排序。
	HorizonEnds []geodata.GeoPoint
	// Subsolar 是该时刻的太阳直射点，缺少精确擦地点时用它采样地平圈近似补口。
	Subsolar geodata.GeoPoint
	// Closed 表示 Boundaries 自身闭合，不需要补口。
	Closed bool
}

// Terminator 返回以太阳直射点为圆心的地平圈采样点。
func Terminator(subsolar geodata.GeoPoint) []geodata.GeoPoint {
	return geodata.SphericalCircle(subsolar, 90, 360)
}

// HorizonEnds 按 Boundaries 走向排序两个地平擦地点；点数不是 2 或首段为空时返回 nil。
func HorizonEnds(footprint Footprint) []geodata.GeoPoint {
	if len(footprint.HorizonEnds) != 2 || len(footprint.Boundaries) == 0 ||
		len(footprint.Boundaries[0]) == 0 {
		return nil
	}
	points := []geodata.GeoPoint{footprint.HorizonEnds[0], footprint.HorizonEnds[1]}
	if pointDistanceKM(points[0], footprint.Boundaries[0][0]) >
		pointDistanceKM(points[1], footprint.Boundaries[0][0]) {
		points[0], points[1] = points[1], points[0]
	}
	return points
}

// ExactHorizon 报告足迹能否用两个精确擦地点闭合。
func ExactHorizon(footprint Footprint) bool {
	return len(HorizonEnds(footprint)) == 2
}

// Curve 返回足迹的物理边界折线，重复的闭合点已去掉。
func Curve(footprint Footprint) []geodata.GeoPoint {
	return openRing(geodata.JoinPolylineSegments(footprint.Boundaries))
}

// Ring 返回足迹的填充环与补口后的物理边界折线；ok 为假表示边界点不足以成环。
// exact 为假或缺少擦地点时按 Subsolar 地平圈的最短弧近似补口。
func Ring(footprint Footprint, exact bool) (ring, boundary []geodata.GeoPoint, ok bool) {
	curve := Curve(footprint)
	// 单点开放边界既不能补口也不该报错：调用方按退化区域丢弃。
	if len(curve) == 1 && !footprint.Closed {
		return curve, curve, true
	}
	minimumPoints := 3
	if !footprint.Closed {
		minimumPoints = 2
	}
	if len(curve) < minimumPoints {
		return nil, nil, false
	}
	if footprint.Closed {
		return append([]geodata.GeoPoint(nil), curve...), curve, true
	}
	ring, boundary = closeOpen(footprint, curve, exact)
	if len(openRing(ring)) < 3 {
		return nil, nil, false
	}
	return ring, boundary, true
}

// HorizonRing 补出已拼接的开放边界 curve 的填充环与物理边界折线，不做点数校验。
func HorizonRing(footprint Footprint, curve []geodata.GeoPoint) (ring, boundary []geodata.GeoPoint) {
	return closeOpen(footprint, curve, true)
}

// SnapDistanceKM 是并集线网的节点吸附尺度：比这更近的交点按同一个物理节点处理。
const SnapDistanceKM = 25

// BandPolygons 以零食分连续包络和日升日落相位线为线网、瞬时足迹为覆盖面，
// 返回偏食可见域的并集；ok 为假表示线网无法成面。
// exact 为真时瞬时足迹按精确擦地点补口，为假时按 Subsolar 地平圈近似补口。
func BandPolygons(
	contours, phaseLines [][]geodata.GeoPoint,
	footprints []Footprint,
	exact bool,
	snapDistanceKM float64,
) ([][]geodata.GeoPoint, bool) {
	boundaryLines := make([][]geodata.GeoPoint, 0, len(contours)+len(phaseLines))
	boundaryLines = append(boundaryLines, contours...)
	boundaryLines = append(boundaryLines, phaseLines...)
	fillPolygons := make([][]geodata.GeoPoint, 0, len(footprints))
	coveragePaths := make([][]geodata.GeoPoint, 0, len(footprints)*2)
	for _, footprint := range footprints {
		for _, segment := range footprint.Boundaries {
			if len(segment) > 0 {
				coveragePaths = append(coveragePaths, segment)
			}
		}
		if ring, _, ok := Ring(footprint, exact); ok && len(ring) >= 3 {
			fillPolygons = append(fillPolygons, ring)
		}
	}
	polygons, err := geodata.VisibleLineworkPolygons(
		boundaryLines, fillPolygons, coveragePaths, snapDistanceKM,
	)
	if err != nil || len(polygons) == 0 {
		return nil, false
	}
	for polygonIndex := range polygons {
		for pointIndex := range polygons[polygonIndex] {
			polygons[polygonIndex][pointIndex].Longitude =
				normalizeLongitude(polygons[polygonIndex][pointIndex].Longitude)
		}
	}
	return polygons, true
}

// closeOpen 用精确擦地点闭合开放边界；擦地点缺失或 exact 为假时改用 Subsolar 地平圈的近似弧。
func closeOpen(footprint Footprint, curve []geodata.GeoPoint, exact bool) ([]geodata.GeoPoint, []geodata.GeoPoint) {
	if len(curve) == 0 {
		return nil, nil
	}
	ends := []geodata.GeoPoint(nil)
	if exact {
		ends = HorizonEnds(footprint)
	}
	if len(ends) != 2 {
		ring := append([]geodata.GeoPoint(nil), curve...)
		arc := geodata.ShortestCircleArc(Terminator(footprint.Subsolar), curve[len(curve)-1], curve[0])
		if len(arc) > 1 {
			ring = append(ring, arc[1:]...)
		}
		return ring, curve
	}
	boundary := make([]geodata.GeoPoint, 0, len(curve)+2)
	boundary = append(boundary, ends[0])
	boundary = append(boundary, curve...)
	boundary = append(boundary, ends[1])
	ring := make([]geodata.GeoPoint, 0, len(curve)+3)
	ring = append(ring, curve...)
	ring = append(ring, ends[1])
	arc := geodata.ShortestCircleArc(Terminator(footprint.Subsolar), ends[1], ends[0])
	if len(arc) > 1 {
		ring = append(ring, arc[1:]...)
	}
	return ring, boundary
}

func openRing(points []geodata.GeoPoint) []geodata.GeoPoint {
	if len(points) > 1 && geodata.SameGeoPoint(points[0], points[len(points)-1]) {
		return points[:len(points)-1]
	}
	return points
}

func normalizeLongitude(value float64) float64 {
	value = math.Mod(value+180, 360)
	if value < 0 {
		value += 360
	}
	return value - 180
}

func pointDistanceKM(first, second geodata.GeoPoint) float64 {
	lat1, lat2 := first.Latitude*math.Pi/180, second.Latitude*math.Pi/180
	dlat := lat2 - lat1
	dlon := math.Remainder((second.Longitude-first.Longitude)*math.Pi/180, 2*math.Pi)
	h := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dlon/2)*math.Sin(dlon/2)
	return 6371.0088 * 2 * math.Asin(math.Sqrt(math.Max(0, math.Min(1, h))))
}
