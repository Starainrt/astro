package occultationgeo

import (
	"math"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

const directBandRingMaximumEdgeKM = 25.0

// DirectVisibleBandPolygons 直接从一系列连续开放接触弧组装静态环。
// DirectVisibleBandPolygons assembles a static ring directly from one
// continuous sequence of open contact arcs.  The cycle is:
//
//	first contact arc -> endpoint-B track -> last contact arc (reverse) -> endpoint-A track (reverse)
//
// This is the authoritative construction for grazing finite-disk events whose
// instantaneous visible footprints are all open.  Events containing closed
// footprints or multiple simultaneous arcs return ok=false and remain on the
// existing coverage/polygonization fallback.
func DirectVisibleBandPolygons(
	footprints []basic.OccultationFootprint,
) (polygons [][]geodata.GeoPoint, ok bool) {
	if len(footprints) < 2 {
		return nil, false
	}
	samples := make([]directOpenBoundarySample, 0, len(footprints))
	for _, footprint := range footprints {
		if footprint.Closed || len(footprint.Boundaries) != 1 || len(footprint.Boundaries[0]) < 2 {
			return nil, false
		}
		boundary := make([]geodata.GeoPoint, 0, len(footprint.Boundaries[0]))
		for _, point := range footprint.Boundaries[0] {
			boundary = append(boundary, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
		}
		if len(boundary) < 2 {
			return nil, false
		}
		if len(samples) > 0 {
			previous := samples[len(samples)-1]
			direct := geoDistanceKM(previous.a, boundary[0]) + geoDistanceKM(previous.b, boundary[len(boundary)-1])
			swapped := geoDistanceKM(previous.a, boundary[len(boundary)-1]) + geoDistanceKM(previous.b, boundary[0])
			if swapped < direct {
				reverseGeoPoints(boundary)
			}
		}
		samples = append(samples, directOpenBoundarySample{
			arc: boundary, a: boundary[0], b: boundary[len(boundary)-1],
		})
	}
	if len(samples) < 2 {
		return nil, false
	}
	// For a monotone open-arc sequence the endpoint-track outline is already
	// the intended exterior envelope. It avoids constructing and unioning one
	// quadrilateral per resampled arc segment; retain the older swept union as
	// a fallback for folded or otherwise ambiguous tracks.
	monotoneSamples := make([]geodata.OpenBoundarySweepSample, len(samples))
	for index, sample := range samples {
		monotoneSamples[index] = geodata.OpenBoundarySweepSample{
			Boundaries: [][]geodata.GeoPoint{sample.arc},
		}
	}
	if monotone, monotoneErr := geodata.MonotoneOpenBoundarySweep(monotoneSamples); monotoneErr == nil {
		candidate := normalizeDirectBandPolygons(monotone)
		if directBandCoversFootprints(candidate, footprints) {
			return candidate, true
		}
	}
	if swept, sweepOK := directOpenBoundarySweep(samples); sweepOK {
		candidate := normalizeDirectBandPolygons(directBandAugmentVisibleFill(swept, footprints))
		if directBandCoversFootprints(candidate, footprints) {
			return candidate, true
		}
	}
	first, last := samples[0], samples[len(samples)-1]
	ring := make([]geodata.GeoPoint, 0, len(first.arc)+len(last.arc)+2*len(samples)+2)
	for _, point := range first.arc {
		appendDirectRingPoint(&ring, point)
	}
	for index := 1; index < len(samples); index++ {
		appendDirectRingPoint(&ring, samples[index].b)
	}
	for index := len(last.arc) - 1; index >= 0; index-- {
		appendDirectRingPoint(&ring, last.arc[index])
	}
	for index := len(samples) - 2; index >= 0; index-- {
		appendDirectRingPoint(&ring, samples[index].a)
	}
	if len(ring) < 4 {
		return nil, false
	}
	appendDirectRingPoint(&ring, ring[0])
	ring = densifyDirectRing(ring, directBandRingMaximumEdgeKM)
	ring = removeOccultationHairpins(ring, 35, 25, 12)
	ring = removeOccultationHairpins(ring, 100, 25, 32)
	ring = removeOccultationSharpCorners(ring, 20, 30)
	ring = densifyDirectRing(ring, directBandRingMaximumEdgeKM)
	if len(ring) < 4 || math.Abs(directRingArea(ring)) <= 1e-9 {
		return nil, false
	}
	polygons = [][]geodata.GeoPoint{ring}
	polygons = normalizeDirectBandPolygons(directBandAugmentVisibleFill(polygons, footprints))
	if !directBandCoversFootprints(polygons, footprints) {
		return nil, false
	}
	return polygons, true
}

func normalizeDirectBandPolygons(polygons [][]geodata.GeoPoint) [][]geodata.GeoPoint {
	result := make([][]geodata.GeoPoint, 0, len(polygons))
	for _, source := range polygons {
		if len(source) < 3 {
			continue
		}
		ring := append([]geodata.GeoPoint(nil), source...)
		if len(ring) > 1 && !geodata.SameGeoPoint(ring[0], ring[len(ring)-1]) {
			ring = append(ring, ring[0])
		}
		ring = densifyDirectRing(ring, directBandRingMaximumEdgeKM)
		ring = removeDirectProjectedSharpCorners(ring, 20, 30)
		ring = densifyDirectRing(ring, directBandRingMaximumEdgeKM)
		if len(ring) >= 3 && math.Abs(directRingArea(ring)) > 1e-9 {
			result = append(result, ring)
		}
	}
	return result
}

func removeDirectProjectedSharpCorners(
	points []geodata.GeoPoint,
	maximumChordKM, minimumAngleDegrees float64,
) []geodata.GeoPoint {
	if len(points) < 4 || maximumChordKM <= 0 || minimumAngleDegrees <= 0 {
		return points
	}
	result := append([]geodata.GeoPoint(nil), points...)
	for {
		changed := false
		for index := 1; index+1 < len(result); index++ {
			first, middle, last := result[index-1], result[index], result[index+1]
			if geoDistanceKM(first, last) > maximumChordKM {
				continue
			}
			scale := math.Cos(middle.Latitude * math.Pi / 180)
			firstX := math.Remainder(first.Longitude-middle.Longitude, 360) * scale
			firstY := first.Latitude - middle.Latitude
			lastX := math.Remainder(last.Longitude-middle.Longitude, 360) * scale
			lastY := last.Latitude - middle.Latitude
			firstLength := math.Hypot(firstX, firstY)
			lastLength := math.Hypot(lastX, lastY)
			if firstLength <= 1e-12 || lastLength <= 1e-12 {
				continue
			}
			cosine := (firstX*lastX + firstY*lastY) / (firstLength * lastLength)
			cosine = math.Max(-1, math.Min(1, cosine))
			if math.Acos(cosine)*180/math.Pi >= minimumAngleDegrees {
				continue
			}
			result = append(result[:index], result[index+1:]...)
			changed = true
			index--
		}
		if !changed {
			break
		}
	}
	return result
}

// directBandAugmentVisibleFill adds the horizon-clipped instantaneous faces
// only when the open contact-arc envelope is not enough to witness the source
// footprints. The common case remains a single continuous sweep; polar
// horizon lobes get a bounded union of already-visible source faces instead
// of falling through to the global linework polygonizer.
func directBandAugmentVisibleFill(
	polygons [][]geodata.GeoPoint,
	footprints []basic.OccultationFootprint,
) [][]geodata.GeoPoint {
	if len(polygons) == 0 {
		return polygons
	}
	if directBandCoversFootprints(polygons, footprints) {
		return polygons
	}
	visibleFill, _ := occultationVisibleFillAndCoverage(footprints)
	if len(visibleFill) == 0 {
		return polygons
	}
	input := append([][]geodata.GeoPoint(nil), polygons...)
	input = append(input, visibleFill...)
	merged, err := geodata.UnionPolygons(input)
	if err != nil || len(merged) == 0 {
		return polygons
	}
	return merged
}

// directBandCoversFootprints is a postcondition for the fast open-arc ring.
// Boundary continuity alone is insufficient near a polar fold: a correctly
// paired pair of arcs can still leave an interior lobe outside the assembled
// ring. Reuse the same boundary probes as the ordinary sweep and add the
// polar/interior witnesses used by the authoritative linework validator.
func directBandCoversFootprints(
	polygons [][]geodata.GeoPoint,
	footprints []basic.OccultationFootprint,
) bool {
	if len(polygons) == 0 || !footprintSweepCoversSamples(polygons, footprints) {
		return false
	}
	_, coveragePaths := occultationVisibleFillAndCoverage(footprints)
	if len(coveragePaths) == 0 {
		return true
	}
	return geodata.SphericalPolygonsContainPathsWithinKM(polygons, coveragePaths, true, 25)
}

func directOpenBoundarySweep(samples []directOpenBoundarySample) ([][]geodata.GeoPoint, bool) {
	// Source contact arcs are already sampled at roughly 35 km. Sixty-four
	// cross-arc samples keeps the rendered chord below the 25 km display target
	// after densification while reducing the union input for long timelines.
	const arcSamples = 64
	quads := make([][]geodata.GeoPoint, 0, (len(samples)-1)*arcSamples)
	previous := resampleDirectArc(samples[0].arc, arcSamples)
	for index := 1; index < len(samples); index++ {
		current := resampleDirectArc(samples[index].arc, arcSamples)
		if len(current) != len(previous) {
			return nil, false
		}
		for pointIndex := 1; pointIndex < len(current); pointIndex++ {
			quad := []geodata.GeoPoint{
				previous[pointIndex-1], current[pointIndex-1], current[pointIndex], previous[pointIndex],
			}
			if math.Abs(directRingArea(quad)) > 1e-10 {
				quads = append(quads, quad)
			}
		}
		previous = current
	}
	if len(quads) == 0 {
		return nil, false
	}
	merged, err := geodata.UnionPolygons(quads)
	if err != nil || len(merged) == 0 {
		return nil, false
	}
	for index := range merged {
		merged[index] = densifyDirectRing(merged[index], directBandRingMaximumEdgeKM)
	}
	return merged, true
}

func resampleDirectArc(arc []geodata.GeoPoint, count int) []geodata.GeoPoint {
	if len(arc) < 2 || count < 2 {
		return nil
	}
	result := make([]geodata.GeoPoint, count)
	for index := range result {
		position := float64(index) * float64(len(arc)-1) / float64(count-1)
		left := int(math.Floor(position))
		if left >= len(arc)-1 {
			result[index] = arc[len(arc)-1]
			continue
		}
		result[index] = interpolateOccultationGeoPoint(arc[left], arc[left+1], position-float64(left))
	}
	return result
}

type directOpenBoundarySample struct {
	arc []geodata.GeoPoint
	a   geodata.GeoPoint
	b   geodata.GeoPoint
}

func reverseGeoPoints(points []geodata.GeoPoint) {
	for left, right := 0, len(points)-1; left < right; left, right = left+1, right-1 {
		points[left], points[right] = points[right], points[left]
	}
}

func appendDirectRingPoint(ring *[]geodata.GeoPoint, point geodata.GeoPoint) {
	if len(*ring) > 0 && geoDistanceKM((*ring)[len(*ring)-1], point) <= 1e-6 {
		return
	}
	*ring = append(*ring, point)
}

func densifyDirectRing(ring []geodata.GeoPoint, maximumEdgeKM float64) []geodata.GeoPoint {
	if len(ring) < 2 || maximumEdgeKM <= 0 {
		return ring
	}
	result := make([]geodata.GeoPoint, 0, len(ring)*2)
	for index, point := range ring {
		result = append(result, point)
		if index+1 >= len(ring) {
			continue
		}
		next := ring[index+1]
		steps := int(math.Ceil(math.Max(
			geoDistanceKM(point, next), occultationProjectedEdgeDistanceKM(point, next),
		) / maximumEdgeKM))
		if steps < 2 {
			continue
		}
		for step := 1; step < steps; step++ {
			fraction := float64(step) / float64(steps)
			result = append(result, interpolateOccultationGeoPoint(point, next, fraction))
		}
	}
	return result
}

func directRingArea(ring []geodata.GeoPoint) float64 {
	if len(ring) < 3 {
		return 0
	}
	area := 0.0
	for index := 1; index < len(ring); index++ {
		previous := ring[index-1]
		current := ring[index]
		latitude := (previous.Latitude + current.Latitude) * math.Pi / 360
		area += math.Remainder(current.Longitude-previous.Longitude, 360) * math.Cos(latitude)
	}
	return area
}
