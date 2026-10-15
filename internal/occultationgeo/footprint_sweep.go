package occultationgeo

import (
	"fmt"
	"math"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

func footprintClosedSweepPolygons(
	footprints []basic.OccultationFootprint,
) [][]geodata.GeoPoint {
	polygons := make([][]geodata.GeoPoint, 0, len(footprints))
	var previous []geodata.GeoPoint
	for _, footprint := range footprints {
		if !footprint.Closed {
			previous = nil
			continue
		}
		ring, ok := footprintClosedSweepRing(footprint)
		if !ok {
			previous = nil
			continue
		}
		ring = resampleClosedFootprintRing(ring, closedFootprintSweepPoints)
		if len(ring) < 4 {
			previous = nil
			continue
		}
		if len(previous) == 0 {
			previous = ring
			continue
		}
		ring = alignClosedFootprintRing(previous, ring)
		if footprintClosedSweepMaximumStep(previous, ring) > closedFootprintSweepMaxStepKM {
			previous = ring
			continue
		}
		polygons = append(polygons, footprintClosedSweepCells(previous, ring)...)
		previous = ring
	}
	return polygons
}

func footprintClosedSweepRing(
	footprint basic.OccultationFootprint,
) ([]geodata.GeoPoint, bool) {
	boundary := openFootprintRing(geodata.JoinPolylineSegments(footprintGeoBoundaries(footprint)))
	if len(boundary) >= 4 {
		return boundary, true
	}
	for _, source := range footprint.Polygons {
		if len(source) < 4 || occultationInteriorPolygon(source, footprint.InteriorPolygons) {
			continue
		}
		ring := make([]geodata.GeoPoint, len(source))
		for index, point := range source {
			ring[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		ring = openFootprintRing(ring)
		if len(ring) >= 4 {
			return ring, true
		}
	}
	return nil, false
}

func resampleClosedFootprintRing(
	ring []geodata.GeoPoint,
	count int,
) []geodata.GeoPoint {
	ring = openFootprintRing(ring)
	if count <= 0 || len(ring) <= count {
		return append([]geodata.GeoPoint(nil), ring...)
	}
	result := make([]geodata.GeoPoint, count)
	for index := range result {
		result[index] = ring[index*len(ring)/count]
	}
	return result
}

func alignClosedFootprintRing(
	previous, current []geodata.GeoPoint,
) []geodata.GeoPoint {
	if len(previous) == 0 || len(current) == 0 || len(previous) != len(current) {
		return append([]geodata.GeoPoint(nil), current...)
	}
	bestShift := 0
	bestReversed := false
	bestScore := math.Inf(1)
	for _, reversed := range []bool{false, true} {
		candidate := append([]geodata.GeoPoint(nil), current...)
		if reversed {
			reverseGeoPointRing(candidate)
		}
		for shift := 0; shift < len(candidate); shift++ {
			score := closedFootprintRingAlignmentScore(previous, candidate, shift)
			if score < bestScore {
				bestScore = score
				bestShift = shift
				bestReversed = reversed
			}
		}
	}
	aligned := append([]geodata.GeoPoint(nil), current...)
	if bestReversed {
		reverseGeoPointRing(aligned)
	}
	return rotateClosedFootprintRing(aligned, bestShift)
}

func reverseGeoPointRing(points []geodata.GeoPoint) {
	for left, right := 0, len(points)-1; left < right; left, right = left+1, right-1 {
		points[left], points[right] = points[right], points[left]
	}
}

func closedFootprintRingAlignmentScore(
	previous, current []geodata.GeoPoint,
	shift int,
) float64 {
	samples := 16
	if len(previous) < samples {
		samples = len(previous)
	}
	score := 0.0
	for sample := 0; sample < samples; sample++ {
		index := sample * len(previous) / samples
		score += geodata.EquatorialDistanceKM(previous[index], current[(index+shift)%len(current)])
	}
	return score
}

func rotateClosedFootprintRing(
	points []geodata.GeoPoint,
	shift int,
) []geodata.GeoPoint {
	result := make([]geodata.GeoPoint, len(points))
	for index := range result {
		result[index] = points[(index+shift)%len(points)]
	}
	return result
}

func footprintClosedSweepMaximumStep(
	first, second []geodata.GeoPoint,
) float64 {
	count := len(first)
	if len(second) < count {
		count = len(second)
	}
	maximum := 0.0
	for index := 0; index < count; index++ {
		maximum = math.Max(maximum, geodata.EquatorialDistanceKM(first[index], second[index]))
	}
	return maximum
}

func footprintClosedSweepCells(
	first, second []geodata.GeoPoint,
) [][]geodata.GeoPoint {
	count := len(first)
	if len(second) < count {
		count = len(second)
	}
	polygons := make([][]geodata.GeoPoint, 0, count)
	for index := 0; index < count; index++ {
		next := (index + 1) % count
		polygon := []geodata.GeoPoint{
			first[index],
			second[index],
			second[next],
			first[next],
			first[index],
		}
		if math.Abs(geoRingArea(polygon)) <= 1e-10 {
			continue
		}
		polygons = append(polygons, polygon)
	}
	return polygons
}

// footprintSweepSampleProbes 采样开放接触弧上的可见边界点作为覆盖见证探针。
func footprintSweepSampleProbes(footprints []basic.OccultationFootprint) [][]geodata.GeoPoint {
	probes := make([][]geodata.GeoPoint, 0, len(footprints)*16)
	for _, footprint := range footprints {
		// Closed footprints are added separately as instantaneous caps. They do
		// not belong to the open-boundary sweep, so checking them here would
		// reject every valid sweep that contains a horizon-closed sample.
		if footprint.Closed {
			continue
		}
		for _, boundary := range footprint.Boundaries {
			step := len(boundary) / 16
			if step < 1 {
				step = 1
			}
			for index := 0; index < len(boundary); index += step {
				point := geodata.GeoPoint{Longitude: boundary[index].Longitude, Latitude: boundary[index].Latitude}
				probes = append(probes, []geodata.GeoPoint{point})
			}
		}
	}
	return probes
}

func footprintSweepCoversSamples(
	polygons [][]geodata.GeoPoint,
	footprints []basic.OccultationFootprint,
) bool {
	if len(polygons) == 0 {
		return false
	}
	probes := footprintSweepSampleProbes(footprints)
	// At a high-latitude open/closed transition the sweep edge can coincide
	// with a source tangent sample within floating-point error. A sub-kilometre
	// tolerance accepts that shared physical edge without admitting a real
	// branch gap. Ordinary-latitude stellar bands retain the strict check so a
	// genuinely missing center-line segment cannot be hidden by the tolerance.
	toleranceKM := 0.0
	minimumAbsoluteLatitude := 90.0
	allNorthern, allSouthern := true, true
	for _, footprint := range footprints {
		for _, boundary := range footprint.Boundaries {
			for _, point := range boundary {
				minimumAbsoluteLatitude = math.Min(minimumAbsoluteLatitude, math.Abs(point.Latitude))
				allNorthern = allNorthern && point.Latitude >= 0
				allSouthern = allSouthern && point.Latitude <= 0
			}
		}
	}
	if minimumAbsoluteLatitude >= 40 && (allNorthern || allSouthern) {
		toleranceKM = 1
	}
	return geodata.SphericalPolygonsContainPathsWithinKM(polygons, probes, true, toleranceKM)
}

// footprintVisibleUnionPolygons returns the sampled, horizon-closed visible
// area without replacing it by open contact-arc ribbons. It is used as a mask
// audit for polar sweeps: the source polygons are the only representation that
// carries the instantaneous Moon-above-horizon closure.
func footprintVisibleUnionPolygons(
	footprints []basic.OccultationFootprint,
) [][]geodata.GeoPoint {
	visibleFill := occultationVisibleFootprintFillOnly(footprints)
	inputs := visibleFill
	inputs = append(inputs, footprintStaticInteriorPolygons(footprints)...)
	inputs = usableOccultationPolygons(inputs)
	if len(inputs) == 0 {
		return nil
	}
	merged, err := geodata.UnionPolygons(inputs)
	if err != nil || len(merged) == 0 {
		// The spherical union can reject one numerically open edge at a horizon
		// transition even though the input footprints form a single temporal
		// ribbon. Run the bounded touching merge on the usable source rings so a
		// sub-60 km seam does not leak out as one polygon per time sample.
		return mergeTouchingVisiblePolygons(inputs)
	}
	return mergeTouchingVisiblePolygons(merged)
}

func usableOccultationPolygons(polygons [][]geodata.GeoPoint) [][]geodata.GeoPoint {
	result := make([][]geodata.GeoPoint, 0, len(polygons))
	for _, polygon := range polygons {
		open := openFootprintRing(polygon)
		if len(open) < 3 || math.Abs(geoRingArea(open)) <= 1e-12 {
			continue
		}
		result = append(result, polygon)
	}
	return result
}

// occultationUnitBounds 是环的三维单位向量包围盒（已含大圆弧外凸余量）。
type occultationUnitBounds struct {
	minX, minY, minZ float64
	maxX, maxY, maxZ float64
}

func occultationUnitVector(point geodata.GeoPoint) (float64, float64, float64) {
	latitude := point.Latitude * math.Pi / 180
	longitude := point.Longitude * math.Pi / 180
	cosLatitude := math.Cos(latitude)
	return cosLatitude * math.Cos(longitude), cosLatitude * math.Sin(longitude), math.Sin(latitude)
}

func occultationPolygonUnitBounds(polygon []geodata.GeoPoint) occultationUnitBounds {
	bounds := occultationUnitBounds{
		minX: math.Inf(1), minY: math.Inf(1), minZ: math.Inf(1),
		maxX: math.Inf(-1), maxY: math.Inf(-1), maxZ: math.Inf(-1),
	}
	vectors := make([][3]float64, 0, len(polygon))
	minimumCosine := 1.0
	for _, point := range polygon {
		x, y, z := occultationUnitVector(point)
		vectors = append(vectors, [3]float64{x, y, z})
	}
	for index, vector := range vectors {
		next := vectors[(index+1)%len(vectors)]
		if cosine := vector[0]*next[0] + vector[1]*next[1] + vector[2]*next[2]; cosine < minimumCosine {
			minimumCosine = cosine
		}
	}
	// 大圆弧中点会凸出端点坐标，按最长弧的半角放大包围盒，保证下界仍然成立。
	scale := 1.0
	if half := math.Sqrt(math.Max(0, (1+minimumCosine)/2)); half > 1e-9 {
		scale = 1 / half
	}
	for _, vector := range vectors {
		for axis, value := range vector {
			high, low := value, value
			if value > 0 {
				high = value * scale
			} else {
				low = value * scale
			}
			switch axis {
			case 0:
				bounds.minX, bounds.maxX = math.Min(bounds.minX, low), math.Max(bounds.maxX, high)
			case 1:
				bounds.minY, bounds.maxY = math.Min(bounds.minY, low), math.Max(bounds.maxY, high)
			default:
				bounds.minZ, bounds.maxZ = math.Min(bounds.minZ, low), math.Max(bounds.maxZ, high)
			}
		}
	}
	return bounds
}

// occultationUnitBoundsNear 报告两包围盒的弦距下界是否可能小于 limitKM。
func occultationUnitBoundsNear(first, second occultationUnitBounds, limitKM float64) bool {
	dx := math.Max(0, math.Max(first.minX-second.maxX, second.minX-first.maxX))
	dy := math.Max(0, math.Max(first.minY-second.maxY, second.minY-first.maxY))
	dz := math.Max(0, math.Max(first.minZ-second.maxZ, second.minZ-first.maxZ))
	// 大圆距离不小于弦长，因此弦距下界可以作为 60 km 近邻判定的必要条件。
	return geodata.EarthEquatorialRadiusKM*math.Sqrt(dx*dx+dy*dy+dz*dz) <= limitKM
}

func mergeTouchingVisiblePolygons(polygons [][]geodata.GeoPoint) [][]geodata.GeoPoint {
	const (
		// Boolean operations on adjacent five-minute footprints can leave
		// several-kilometre numerical gaps even though the temporal samples
		// overlap. Keep a modest 60 km ceiling for the rare last-gap transition;
		// larger separations remain disconnected physical components and are not
		// bridged.
		touchingDistanceKM   = 60.0
		minimumBridgeHalfDeg = 0.01
	)
	if len(polygons) < 2 {
		return polygons
	}
	bounds := make([]occultationUnitBounds, len(polygons))
	for index, polygon := range polygons {
		bounds[index] = occultationPolygonUnitBounds(polygon)
	}
	for len(polygons) > 1 {
		firstIndex, secondIndex := -1, -1
		var touching, leftTouch, rightTouch geodata.GeoPoint
		for first := 0; first < len(polygons) && firstIndex < 0; first++ {
			for second := first + 1; second < len(polygons) && firstIndex < 0; second++ {
				if !occultationUnitBoundsNear(bounds[first], bounds[second], touchingDistanceKM) {
					continue
				}
				for _, left := range polygons[first] {
					for _, right := range polygons[second] {
						if geodata.EquatorialDistanceKM(left, right) > touchingDistanceKM {
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
		if gap := geodata.EquatorialDistanceKM(leftTouch, rightTouch) / geodata.EarthEquatorialRadiusKM * 180 / math.Pi; gap/2+0.002 > bridgeHalfDeg {
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
		nextBounds := make([]occultationUnitBounds, 0, len(polygons)-1)
		for index, polygon := range polygons {
			if index == firstIndex {
				next = append(next, pair[0])
				nextBounds = append(nextBounds, occultationPolygonUnitBounds(pair[0]))
				continue
			}
			if index == secondIndex {
				continue
			}
			next = append(next, polygon)
			nextBounds = append(nextBounds, bounds[index])
		}
		polygons, bounds = next, nextBounds
	}
	return polygons
}

func footprintSweepNeedsHorizonClipping(
	swept, visibleUnion [][]geodata.GeoPoint,
	footprints []basic.OccultationFootprint,
) bool {
	if len(swept) == 0 || len(visibleUnion) == 0 {
		return false
	}
	// A single continuous open sweep against a fragmented horizon union is a
	// topological disagreement, not a small metric residual. The caller must
	// use the continuous contact/phase linework in this case; avoid measuring
	// every edge of all sampled fragments against the sweep because that exact
	// spherical distance check dominates compact polar events.
	if len(swept) == 1 && len(visibleUnion) > 1 {
		return true
	}
	// A disconnected open sweep is the common polar-fold failure mode: its
	// endpoint ribbons can bridge across a cap even though the horizon-closed
	// source remains split into separate visible faces.
	if len(swept) > 1 && len(visibleUnion) > len(swept) {
		return true
	}
	if len(swept) > 1 {
		for _, footprint := range footprints {
			for _, boundary := range footprint.Boundaries {
				for _, point := range boundary {
					if math.Abs(point.Latitude) >= 70 {
						return true
					}
				}
			}
		}
	}
	// Probe vertices and edge midpoints of the smooth sweep against the
	// horizon-closed union. A miss larger than a small numerical tolerance means
	// the ribbon has crossed into the below-horizon complement.
	if !geodata.SphericalPolygonsContainPathsWithinKM(visibleUnion, swept, true, 20) {
		return true
	}
	probes := make([][]geodata.GeoPoint, 0, 2048)
	for _, footprint := range footprints {
		for _, source := range footprint.Polygons {
			if len(source) < 3 {
				continue
			}
			ring := make([]geodata.GeoPoint, len(source))
			for index, point := range source {
				ring[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			for _, probe := range occultationVisibleFootprintProbes(source, ring) {
				probes = append(probes, []geodata.GeoPoint{probe})
			}
		}
	}
	probes = limitOccultationCoveragePaths(probes, 2048)
	if len(probes) > 0 && !geodata.SphericalPolygonsContainPathsWithinKM(swept, probes, false, 25) {
		return true
	}
	// The opposite miss means the open ribbon dropped a legitimate visible
	// footprint lobe. Keep the horizon-closed union in that case as well; a
	// static band must contain every sampled instantaneous witness.
	return !geodata.SphericalPolygonsContainPathsWithinKM(swept, visibleUnion, true, 25)
}

func footprintBoundariesAvailable(footprints []basic.OccultationFootprint) bool {
	if len(footprints) == 0 {
		return false
	}
	for _, footprint := range footprints {
		if len(footprint.Boundaries) == 0 {
			return false
		}
	}
	return true
}

func footprintOpenSweepPolygons(footprints []basic.OccultationFootprint) ([][]geodata.GeoPoint, error) {
	return footprintOpenSweepPolygonsWithTransitions(footprints, true)
}

func footprintOpenSweepPolygonsWithoutTransitions(
	footprints []basic.OccultationFootprint,
) ([][]geodata.GeoPoint, error) {
	return footprintOpenSweepPolygonsWithTransitions(footprints, false)
}

func footprintOpenSweepPolygonsWithTransitions(
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

func footprintGeoBoundaries(footprint basic.OccultationFootprint) [][]geodata.GeoPoint {
	boundaries := make([][]geodata.GeoPoint, len(footprint.Boundaries))
	for boundaryIndex, source := range footprint.Boundaries {
		boundary := make([]geodata.GeoPoint, len(source))
		for pointIndex, point := range source {
			boundary[pointIndex] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		boundaries[boundaryIndex] = boundary
	}
	return boundaries
}

func footprintTransitionBoundary(
	closed, adjacent basic.OccultationFootprint,
) ([]geodata.GeoPoint, bool) {
	closedRing := openFootprintRing(geodata.JoinPolylineSegments(footprintGeoBoundaries(closed)))
	adjacentArc := openFootprintRing(geodata.JoinPolylineSegments(footprintGeoBoundaries(adjacent)))
	if len(closedRing) < 3 || len(adjacentArc) < 2 {
		return nil, false
	}
	start := nearestFootprintPointIndex(closedRing, adjacentArc[0])
	end := nearestFootprintPointIndex(closedRing, adjacentArc[len(adjacentArc)-1])
	if start == end {
		forward := footprintOpenedRing(closedRing, start, 1)
		backward := footprintOpenedRing(closedRing, start, -1)
		if footprintArcMatchScore(backward, adjacentArc) < footprintArcMatchScore(forward, adjacentArc) {
			return backward, true
		}
		return forward, true
	}
	forward := footprintRingArc(closedRing, start, end, 1)
	backward := footprintRingArc(closedRing, start, end, -1)
	if footprintArcMatchScore(backward, adjacentArc) < footprintArcMatchScore(forward, adjacentArc) {
		return backward, true
	}
	return forward, true
}

func nearestFootprintPointIndex(points []geodata.GeoPoint, target geodata.GeoPoint) int {
	nearest := 0
	distance := math.Inf(1)
	for index, point := range points {
		candidate := geodata.EquatorialDistanceKM(point, target)
		if candidate < distance {
			nearest, distance = index, candidate
		}
	}
	return nearest
}

func footprintOpenedRing(points []geodata.GeoPoint, start, direction int) []geodata.GeoPoint {
	result := make([]geodata.GeoPoint, len(points))
	for index := range result {
		position := (start + direction*index) % len(points)
		if position < 0 {
			position += len(points)
		}
		result[index] = points[position]
	}
	return result
}

func footprintRingArc(points []geodata.GeoPoint, start, end, direction int) []geodata.GeoPoint {
	arc := make([]geodata.GeoPoint, 1, len(points)+1)
	arc[0] = points[start]
	index := start
	for step := 1; step <= len(points); step++ {
		index = (index + direction + len(points)) % len(points)
		arc = append(arc, points[index])
		if index == end {
			return arc
		}
	}
	return arc
}

func footprintArcMatchScore(candidate, reference []geodata.GeoPoint) float64 {
	if len(candidate) < 2 || len(reference) < 2 {
		return math.Inf(1)
	}
	const samples = 17
	score := 0.0
	for index := 0; index < samples; index++ {
		candidateIndex := index * (len(candidate) - 1) / (samples - 1)
		referenceIndex := index * (len(reference) - 1) / (samples - 1)
		score += geodata.EquatorialDistanceKM(candidate[candidateIndex], reference[referenceIndex])
	}
	return score
}

func openFootprintRing(points []geodata.GeoPoint) []geodata.GeoPoint {
	if len(points) > 1 && geodata.SameGeoPoint(points[0], points[len(points)-1]) {
		return points[:len(points)-1]
	}
	return points
}

func footprintClosedPolygons(footprints []basic.OccultationFootprint) [][]geodata.GeoPoint {
	var polygons [][]geodata.GeoPoint
	for _, footprint := range footprints {
		if !footprint.Closed || len(footprint.Boundaries) == 0 {
			continue
		}
		polygons = append(polygons, footprintPolygons([]basic.OccultationFootprint{footprint})...)
	}
	return polygons
}

func footprintStaticInteriorPolygons(footprints []basic.OccultationFootprint) [][]geodata.GeoPoint {
	var polygons [][]geodata.GeoPoint
	for _, footprint := range footprints {
		if footprint.Closed {
			continue
		}
		for _, source := range footprint.InteriorPolygons {
			if len(source) < 3 || !occultationStaticInteriorPolygon(source, footprint.InteriorPolygons) {
				continue
			}
			center := geodata.GeoPoint{Longitude: source[0].Longitude, Latitude: source[0].Latitude}
			polygon, ok := footprintStaticBoundaryBridge(center, footprint.Boundaries)
			if !ok {
				polygon = make([]geodata.GeoPoint, len(source))
				for index, point := range source {
					polygon[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
				}
			}
			// A footprint can briefly collapse to a tangent point at a
			// visibility transition. Its bridge then repeats the same vertex and
			// is not a polygon; passing it to UnionPolygons aborts the entire band.
			if len(polygon) >= 3 && math.Abs(geoRingArea(polygon)) > 1e-12 {
				polygons = append(polygons, polygon)
			}
			centerCap := geodata.SphericalCircle(
				center,
				staticCenterCapRadiusKM/geodata.EarthEquatorialRadiusKM*180/math.Pi,
				staticCenterCapPoints,
			)
			if len(centerCap) >= 3 {
				centerCap = append(centerCap, centerCap[0])
				polygons = append(polygons, centerCap)
			}
		}
	}
	return polygons
}

func footprintStaticBoundaryBridge(
	center geodata.GeoPoint,
	boundaries [][]basic.OccultationPathPoint,
) ([]geodata.GeoPoint, bool) {
	nearestDistance := math.Inf(1)
	var nearestStart, nearestEnd geodata.GeoPoint
	for _, boundary := range boundaries {
		for index := 1; index < len(boundary); index++ {
			start := geodata.GeoPoint{
				Longitude: boundary[index-1].Longitude,
				Latitude:  boundary[index-1].Latitude,
			}
			end := geodata.GeoPoint{
				Longitude: boundary[index].Longitude,
				Latitude:  boundary[index].Latitude,
			}
			distance := geoPointSegmentDistanceKM(center, start, end)
			if distance < nearestDistance {
				nearestDistance = distance
				nearestStart = start
				nearestEnd = end
			}
		}
	}
	if !finiteGeo(nearestDistance) {
		return nil, false
	}
	return []geodata.GeoPoint{center, nearestStart, nearestEnd, center}, true
}

func footprintPolygons(footprints []basic.OccultationFootprint) [][]geodata.GeoPoint {
	var polygons [][]geodata.GeoPoint
	for _, footprint := range footprints {
		for _, source := range footprint.Polygons {
			if occultationInteriorPolygon(source, footprint.InteriorPolygons) {
				continue
			}
			polygon := make([]geodata.GeoPoint, len(source))
			for index, point := range source {
				polygon[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			polygons = append(polygons, polygon)
		}
	}
	return polygons
}

func geoPointSegmentDistanceKM(point, start, end geodata.GeoPoint) float64 {
	latitude := point.Latitude * math.Pi / 180
	scaleX := math.Cos(latitude) * geodata.EarthEquatorialRadiusKM * math.Pi / 180
	scaleY := geodata.EarthEquatorialRadiusKM * math.Pi / 180
	x := func(value geodata.GeoPoint) float64 {
		return math.Remainder(value.Longitude-point.Longitude, 360) * scaleX
	}
	y := func(value geodata.GeoPoint) float64 {
		return (value.Latitude - point.Latitude) * scaleY
	}
	startX, startY := x(start), y(start)
	endX, endY := x(end), y(end)
	deltaX, deltaY := endX-startX, endY-startY
	fraction := 0.0
	if lengthSquared := deltaX*deltaX + deltaY*deltaY; lengthSquared > 0 {
		fraction = math.Max(0, math.Min(1,
			-(startX*deltaX+startY*deltaY)/lengthSquared,
		))
	}
	return math.Hypot(startX+fraction*deltaX, startY+fraction*deltaY)
}

func geoRingArea(polygon []geodata.GeoPoint) float64 {
	if len(polygon) < 3 {
		return 0
	}
	longitudes := make([]float64, len(polygon))
	longitudes[0] = polygon[0].Longitude
	for index := 1; index < len(polygon); index++ {
		longitudes[index] = longitudes[index-1] + math.Remainder(
			polygon[index].Longitude-longitudes[index-1], 360,
		)
	}
	area := 0.0
	for index, point := range polygon {
		next := polygon[(index+1)%len(polygon)]
		area += longitudes[index]*next.Latitude - longitudes[(index+1)%len(polygon)]*point.Latitude
	}
	return area / 2
}

func finiteGeo(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func continuousRanges(count int, changed func(int) bool) []SampleRange {
	if count == 0 {
		return nil
	}
	ranges := make([]SampleRange, 0, 2)
	start := 0
	for index := 1; index < count; index++ {
		if !changed(index) {
			continue
		}
		ranges = append(ranges, SampleRange{Start: start, End: index})
		start = index
	}
	return append(ranges, SampleRange{Start: start, End: count})
}

// BoundaryBranchChanged 判断两个相邻样本是否距离过大，无法属于同一物理支路。
// BoundaryBranchChanged reports whether two adjacent samples are too far apart to be one physical branch.
func BoundaryBranchChanged(first, second basic.OccultationPathPoint) bool {
	distance := DistanceKM(first, second)
	if distance <= BoundaryBranchJumpKM {
		return false
	}
	duration := math.Abs(second.Time.Sub(first.Time).Seconds())
	return duration == 0 || distance/duration > BoundaryBranchSpeedKMPerSecond
}

// DistanceKM 返回两个边界样本之间的最短球面距离。
// DistanceKM returns the shortest spherical surface distance between two boundary samples.
func DistanceKM(first, second basic.OccultationPathPoint) float64 {
	return geodata.EquatorialDistanceKM(
		geodata.GeoPoint{Longitude: first.Longitude, Latitude: first.Latitude},
		geodata.GeoPoint{Longitude: second.Longitude, Latitude: second.Latitude},
	)
}
