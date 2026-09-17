package occultationgeo

// 本文件保存改动前的参考实现，仅用于差分对照：新实现必须在相同随机输入上给出逐点相同的输出。

import (
	"fmt"
	"math"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

func mergeTouchingVisiblePolygonsReference(polygons [][]geodata.GeoPoint) [][]geodata.GeoPoint {
	const (
		touchingDistanceKM   = 60.0
		minimumBridgeHalfDeg = 0.01
	)
	for len(polygons) > 1 {
		firstIndex, secondIndex := -1, -1
		var touching, leftTouch, rightTouch geodata.GeoPoint
		for first := 0; first < len(polygons) && firstIndex < 0; first++ {
			for second := first + 1; second < len(polygons) && firstIndex < 0; second++ {
				for _, left := range polygons[first] {
					for _, right := range polygons[second] {
						if geoDistanceKM(left, right) > touchingDistanceKM {
							continue
						}
						firstIndex, secondIndex = first, second
						leftTouch, rightTouch = left, right
						touching = geodata.GeoPoint{
							Longitude: left.Longitude + math.Remainder(right.Longitude-left.Longitude, 360)/2,
							Latitude:  (left.Latitude + right.Latitude) / 2,
						}
						break
					}
					if firstIndex >= 0 {
						break
					}
				}
			}
		}
		if firstIndex < 0 || secondIndex < 0 {
			break
		}
		bridgeHalfDeg := minimumBridgeHalfDeg
		if gap := geoDistanceKM(leftTouch, rightTouch) / EarthRadiusKM * 180 / math.Pi; gap/2+0.002 > bridgeHalfDeg {
			bridgeHalfDeg = gap/2 + 0.002
		}
		bridge := []geodata.GeoPoint{
			{Longitude: touching.Longitude - bridgeHalfDeg, Latitude: touching.Latitude - bridgeHalfDeg},
			{Longitude: touching.Longitude + bridgeHalfDeg, Latitude: touching.Latitude - bridgeHalfDeg},
			{Longitude: touching.Longitude + bridgeHalfDeg, Latitude: touching.Latitude + bridgeHalfDeg},
			{Longitude: touching.Longitude - bridgeHalfDeg, Latitude: touching.Latitude + bridgeHalfDeg},
		}
		leftPolygon := append([]geodata.GeoPoint(nil), polygons[firstIndex]...)
		rightPolygon := append([]geodata.GeoPoint(nil), polygons[secondIndex]...)
		for index, point := range leftPolygon {
			if point == leftTouch {
				leftPolygon[index] = touching
				break
			}
		}
		for index, point := range rightPolygon {
			if point == rightTouch {
				rightPolygon[index] = touching
				break
			}
		}
		pair, err := geodata.UnionPolygons([][]geodata.GeoPoint{
			leftPolygon, rightPolygon, bridge,
		})
		if err != nil || len(pair) != 1 {
			pair, err = geodata.UnionPolygons([][]geodata.GeoPoint{
				leftPolygon, rightPolygon,
			})
		}
		if err != nil || len(pair) != 1 {
			break
		}
		next := make([][]geodata.GeoPoint, 0, len(polygons)-1)
		for index, polygon := range polygons {
			if index == firstIndex {
				next = append(next, pair[0])
				continue
			}
			if index == secondIndex {
				continue
			}
			next = append(next, polygon)
		}
		polygons = next
	}
	return polygons
}

func removeTinyPolygonComponentsReference(polygons [][]geodata.GeoPoint) [][]geodata.GeoPoint {
	if len(polygons) < 2 {
		return polygons
	}
	areas := make([]float64, len(polygons))
	maximum := 0.0
	for index, polygon := range polygons {
		areas[index] = math.Abs(geoRingArea(polygon))
		if areas[index] > maximum {
			maximum = areas[index]
		}
	}
	if maximum <= 0 || !finiteGeo(maximum) {
		return polygons
	}
	threshold := maximum * 1e-4
	filtered := make([][]geodata.GeoPoint, 0, len(polygons))
	for index, polygon := range polygons {
		if areas[index] >= threshold {
			filtered = append(filtered, polygon)
		}
	}
	if len(filtered) == 0 {
		return polygons[:1]
	}
	return filtered
}

func pairedBoundaryPolygonsReference(
	first, second []basic.OccultationPathPoint,
) [][]geodata.GeoPoint {
	count := len(first)
	if len(second) < count {
		count = len(second)
	}
	polygons := make([][]geodata.GeoPoint, 0, count)
	pendingBranch := false
	for index := 1; index < count; index++ {
		firstChanged := BoundaryBranchChanged(first[index-1], first[index])
		secondChanged := BoundaryBranchChanged(second[index-1], second[index])
		if pendingBranch {
			if firstChanged || secondChanged {
				pendingBranch = false
			}
			if firstChanged || secondChanged || pendingBranch {
				continue
			}
		}
		if firstChanged != secondChanged {
			pendingBranch = true
			continue
		}
		if firstChanged { // both sides changed at the same transition
			continue
		}
		previousFirst, currentFirst := first[index-1], first[index]
		previousSecond, currentSecond := second[index-1], second[index]
		polygons = append(polygons, []geodata.GeoPoint{
			{Longitude: previousFirst.Longitude, Latitude: previousFirst.Latitude},
			{Longitude: currentFirst.Longitude, Latitude: currentFirst.Latitude},
			{Longitude: currentSecond.Longitude, Latitude: currentSecond.Latitude},
			{Longitude: previousSecond.Longitude, Latitude: previousSecond.Latitude},
		})
	}
	return polygons
}

func constrainPolygonsWithinReference(
	parent, child [][]geodata.GeoPoint,
) [][]geodata.GeoPoint {
	if len(parent) == 0 || len(child) == 0 {
		return child
	}
	initialMissDistance := geodata.SphericalPolygonsPathMissDistanceKM(parent, child, true)
	if initialMissDistance <= 0 {
		return child
	}
	const maximumRepairDistanceKM = 100.0
	const maximumResidualMissDistanceKM = 10.0
	if initialMissDistance > maximumRepairDistanceKM {
		return child
	}
	result := make([][]geodata.GeoPoint, len(child))
	for index, source := range child {
		if len(source) < 4 {
			if len(openFootprintRing(source)) >= 3 && math.Abs(geoRingArea(source)) > 1e-12 {
				result[index] = append([]geodata.GeoPoint(nil), source...)
			}
			continue
		}
		ring := append([]geodata.GeoPoint(nil), source...)
		closed := geodata.SameGeoPoint(ring[0], ring[len(ring)-1])
		limit := len(ring)
		if closed {
			limit--
		}
		containment := geodata.SphericalPolygonsContainPoints(parent, ring[:limit])
		for pointIndex := 0; pointIndex < limit; pointIndex++ {
			point := ring[pointIndex]
			if containment[pointIndex] {
				continue
			}
			nearest, distance := nearestPolygonBoundaryPoint(parent, point)
			if distance <= maximumRepairDistanceKM {
				ring[pointIndex] = nearest
			}
		}
		if closed {
			ring[len(ring)-1] = ring[0]
		}
		result[index] = ring
	}
	result = usableOccultationPolygons(result)
	if geodata.SphericalPolygonsPathMissDistanceKM(parent, result, true) <= 0 {
		return result
	}
	// Vertex-only repair cannot see a child edge whose endpoints are both
	// inside the parent while its great-circle midpoint crosses outside. Split
	// the repaired ring at the same projected spacing used by output geometry,
	// then apply the local vertex snap to those newly exposed edge probes.
	densified := densifyOccultationPolygons(result, 10)
	for pass := 0; pass < 6; pass++ {
		changed := false
		for index, source := range densified {
			ring := append([]geodata.GeoPoint(nil), source...)
			closed := len(ring) > 1 && geodata.SameGeoPoint(ring[0], ring[len(ring)-1])
			limit := len(ring)
			if closed {
				limit--
			}
			containment := geodata.SphericalPolygonsContainPoints(parent, ring[:limit])
			for pointIndex := 0; pointIndex < limit; pointIndex++ {
				point := ring[pointIndex]
				if containment[pointIndex] {
					continue
				}
				nearest, distance := nearestPolygonBoundaryPoint(parent, point)
				if distance <= maximumRepairDistanceKM {
					ring[pointIndex] = nearest
					changed = true
				}
			}
			if closed {
				ring[len(ring)-1] = ring[0]
			}
			for {
				cleaned := removeDirectProjectedSharpCorners(ring, 20, 30)
				cleaned = removeOccultationSharpCorners(cleaned, 20, 30)
				if len(cleaned) == len(ring) {
					break
				}
				ring = cleaned
			}
			if closed && len(ring) > 1 && !geodata.SameGeoPoint(ring[0], ring[len(ring)-1]) {
				ring = append(ring, ring[0])
			}
			densified[index] = ring
		}
		densified = usableOccultationPolygons(densified)
		if geodata.SphericalPolygonsPathMissDistanceKM(parent, densified, true) <= maximumResidualMissDistanceKM {
			return densified
		}
		if !changed {
			break
		}
	}
	return child
}

func footprintOpenSweepPolygonsWithTransitionsReference(
	footprints []basic.OccultationFootprint,
	includeTransitions bool,
) ([][]geodata.GeoPoint, error) {
	polygons := make([][]geodata.GeoPoint, 0, 2)
	for start := 0; start < len(footprints); {
		for start < len(footprints) && footprints[start].Closed {
			start++
		}
		if start == len(footprints) {
			break
		}
		end := start
		for end < len(footprints) && !footprints[end].Closed {
			end++
		}

		samples := make([]geodata.OpenBoundarySweepSample, 0, end-start+2)
		if includeTransitions && start > 0 {
			boundary, ok := footprintTransitionBoundary(footprints[start-1], footprints[start])
			if ok {
				samples = append(samples, geodata.OpenBoundarySweepSample{Boundaries: [][]geodata.GeoPoint{boundary}})
			}
		}
		for index := start; index < end; index++ {
			samples = append(samples, geodata.OpenBoundarySweepSample{
				Boundaries: footprintGeoBoundaries(footprints[index]),
			})
		}
		if includeTransitions && end < len(footprints) {
			boundary, ok := footprintTransitionBoundary(footprints[end], footprints[end-1])
			if ok {
				samples = append(samples, geodata.OpenBoundarySweepSample{Boundaries: [][]geodata.GeoPoint{boundary}})
			}
		}
		group, err := geodata.OpenBoundarySweep(samples)
		if err != nil && includeTransitions {
			// At an open/closed transition the exact transition arc can be
			// numerically coincident with the first ribbon edge. Retry the same
			// physical run without that synthetic endpoint; the sampled open arcs
			// still provide both endpoint tracks and avoid a false diagonal cap.
			bareSamples := make([]geodata.OpenBoundarySweepSample, 0, end-start)
			for index := start; index < end; index++ {
				bareSamples = append(bareSamples, geodata.OpenBoundarySweepSample{
					Boundaries: footprintGeoBoundaries(footprints[index]),
				})
			}
			group, err = geodata.OpenBoundarySweep(bareSamples)
		}
		if err != nil {
			return nil, err
		}
		polygons = append(polygons, group...)
		start = end
	}
	if len(polygons) == 0 {
		return nil, fmt.Errorf("open footprint samples contain no usable sweep")
	}
	return polygons, nil
}
