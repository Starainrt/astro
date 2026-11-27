package geodata

import (
	"math"
	"sort"
)

type mapEdge struct {
	start, end polygonUnionPoint
}

func (edge mapEdge) parameter(point polygonUnionPoint) float64 {
	if edge.start.x != edge.end.x {
		return (point.x - edge.start.x) / (edge.end.x - edge.start.x)
	}
	return (point.y - edge.start.y) / (edge.end.y - edge.start.y)
}

func (edge mapEdge) contains(from, to polygonUnionPoint) bool {
	if edge.start.x == edge.end.x {
		return from.x == edge.start.x && to.x == edge.start.x
	}
	return from.y == edge.start.y && to.y == edge.start.y
}

// EquirectangularComplement 从世界矩形扣除已切缝的多边形 / subtracts seam-split polygons from the equirectangular world.
func EquirectangularComplement(fragments [][]GeoPoint) ([]Region, error) {
	world := [...]mapEdge{
		{polygonUnionPoint{-180, -90}, polygonUnionPoint{180, -90}},
		{polygonUnionPoint{180, -90}, polygonUnionPoint{180, 90}},
		{polygonUnionPoint{180, 90}, polygonUnionPoint{-180, 90}},
		{polygonUnionPoint{-180, 90}, polygonUnionPoint{-180, -90}},
	}
	var cuts [len(world)][]mapEdge
	var edges []polygonUnionEdge
	for _, fragment := range fragments {
		reverse := stablePolygonArea(fragment) > 0
		for index, point := range fragment {
			next := fragment[(index+1)%len(fragment)]
			if point == next {
				continue
			}
			from, to := polygonUnionPoint{point.Longitude, point.Latitude}, polygonUnionPoint{next.Longitude, next.Latitude}
			if reverse {
				from, to = to, from
			}
			side := 0
			for ; side < len(world); side++ {
				if world[side].contains(from, to) {
					cuts[side] = append(cuts[side], mapEdge{to, from})
					break
				}
			}
			if side == len(world) {
				edges = append(edges, polygonUnionEdge{start: from, end: to})
			}
		}
	}
	for side, boundary := range world {
		removed := cuts[side]
		sort.Slice(removed, func(left, right int) bool {
			return boundary.parameter(removed[left].start) < boundary.parameter(removed[right].start)
		})
		cursor := boundary.start
		for _, cut := range removed {
			if boundary.parameter(cursor) < boundary.parameter(cut.start) {
				edges = appendMapEdge(edges, cursor, cut.start)
			}
			if boundary.parameter(cursor) < boundary.parameter(cut.end) {
				cursor = cut.end
			}
		}
		edges = appendMapEdge(edges, cursor, boundary.end)
	}
	rings, err := polygonUnionTraceEdges(edges, polygonOverlayMinArea)
	if err != nil {
		return nil, err
	}
	for index := range rings {
		for vertex, point := range rings[index].chart {
			rings[index].points[vertex] = GeoPoint{Longitude: point.x, Latitude: point.y}
		}
	}
	return polygonUnionClassifyRegions(rings), nil
}

func appendMapEdge(edges []polygonUnionEdge, from, to polygonUnionPoint) []polygonUnionEdge {
	if math.Abs(to.x-from.x) > 180 {
		middle := polygonUnionPoint{(from.x + to.x) / 2, from.y}
		edges = append(edges, polygonUnionEdge{start: from, end: middle})
		from = middle
	}
	if polygonUnionNodeForPoint(from) != polygonUnionNodeForPoint(to) {
		edges = append(edges, polygonUnionEdge{start: from, end: to})
	}
	return edges
}
