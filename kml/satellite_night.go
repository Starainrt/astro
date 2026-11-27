package kml

import (
	"fmt"

	"github.com/starainrt/astro/internal/geodata"
)

func satelliteNightGeometry(elements []element) (geometry, error) {
	outline, err := geometryFor(elements, false)
	if err != nil {
		return geometry{}, err
	}
	multi := outline.MultiGeometry
	if multi == nil {
		multi = &multiGeometry{}
		if outline.LineString != nil {
			multi.Lines = append(multi.Lines, *outline.LineString)
		}
	}
	for _, item := range elements {
		if item.kind != "Polygon" {
			continue
		}
		if len(item.rings) == 0 {
			return geometry{}, fmt.Errorf("night polygon has no rings")
		}
		var region geodata.Region
		for index, ring := range item.rings {
			positions, err := normalizeRing(ring, index == 0)
			if err != nil {
				return geometry{}, err
			}
			points := make([]geodata.GeoPoint, len(positions))
			for vertex, position := range positions {
				points[vertex] = geodata.GeoPoint{Longitude: position[0], Latitude: position[1]}
			}
			if index == 0 {
				region.Outer = points
			} else {
				region.Holes = append(region.Holes, points)
			}
		}
		pieces, err := geodata.RegionOctants(region)
		if err != nil {
			return geometry{}, fmt.Errorf("clip night polygon: %w", err)
		}
		for _, piece := range pieces {
			shape := polygon{Tessellate: 1, AltitudeMode: "clampToGround", OuterBoundaryIs: nightBoundary(piece.Outer)}
			for _, hole := range piece.Holes {
				shape.InnerBoundaries = append(shape.InnerBoundaries, *nightBoundary(hole))
			}
			multi.Polygons = append(multi.Polygons, shape)
		}
	}
	return geometry{MultiGeometry: multi}, nil
}

func nightBoundary(points []geodata.GeoPoint) *boundary {
	positions := make([][]float64, 0, len(points)+1)
	for _, point := range points {
		positions = append(positions, []float64{point.Longitude, point.Latitude})
	}
	positions = append(positions, positions[0])
	return &boundary{Ring: linearRing{Coordinates: textForPositions(positions)}}
}
