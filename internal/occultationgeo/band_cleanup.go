package occultationgeo

import (
	"fmt"
	"math"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

func cleanupOccultationVisibleBandPolygons(
	polygons [][]geodata.GeoPoint,
	strongPolarSmoothing bool,
) [][]geodata.GeoPoint {
	polygons = RemoveTinyPolygonComponents(polygons)
	maximumDisplayEdgeKM := 150.0
	if strongPolarSmoothing {
		// Inner-contact Saturn fallbacks retain the sampled sweep rather than a
		// selected linework face. Keep their rendered chords at the same scale as
		// the authoritative path so the polar edge cannot appear stepped.
		maximumDisplayEdgeKM = 40
	}
	densified := densifyOccultationPolygons(polygons, maximumDisplayEdgeKM)
	for index := range densified {
		// Densification can turn one long, thin numerical return into more than
		// a dozen short edges. Inspect a wider local window, but require a much
		// larger detour before removing it so ordinary rounded cusps survive.
		densified[index] = removeOccultationHairpins(densified[index], 35, 25, 12)
		densified[index] = removeOccultationHairpins(densified[index], 100, 25, 32)
	}
	if strongPolarSmoothing {
		for index := range densified {
			densified[index] = smoothOccultationHairpins(densified[index], 180, 35, 16, 50)
			densified[index] = removeOccultationPolarKinks(densified[index])
			densified[index] = smoothOccultationPolarCorners(densified[index])
		}
	}
	for index := range densified {
		densified[index] = removeOccultationSharpCorners(densified[index], 20, 30)
	}
	for index := range densified {
		densified[index] = smoothOccultationPolarWobbles(densified[index])
		densified[index] = smoothOccultationOrdinaryWobbles(densified[index])
	}
	return densified
}

// smoothOccultationOrdinaryWobbles removes the small alternating envelope
// error left when adjacent temporal footprints contribute different sampled
// limb vertices. A projected five-point filter is used so the correction
// follows the Web Mercator chart seen by map clients. The displacement and
// local-deviation gates remain deliberately small, so this also handles the
// high-latitude portion of a smooth arc without flattening a real horizon fold.
func smoothOccultationOrdinaryWobbles(points []geodata.GeoPoint) []geodata.GeoPoint {
	if len(points) < 7 {
		return points
	}
	closed := len(points) > 1 && geodata.SameGeoPoint(points[0], points[len(points)-1])
	limit := len(points)
	if closed {
		limit--
	}
	if limit < 7 {
		return points
	}
	result := append([]geodata.GeoPoint(nil), points...)
	original := append([]geodata.GeoPoint(nil), points...)
	const (
		minimumLatitude    = 0.5
		maximumLatitude    = 85.0
		maximumMoveKM      = 20.0
		maximumDeviationKM = 40.0
	)
	for pass := 0; pass < 12; pass++ {
		updated := append([]geodata.GeoPoint(nil), result...)
		for index := 2; index+2 < limit; index++ {
			point := result[index]
			if math.Abs(point.Latitude) < minimumLatitude || math.Abs(point.Latitude) > maximumLatitude {
				continue
			}
			weights := [...]float64{-3, 12, 17, 12, -3}
			var x, y float64
			for offset := -2; offset <= 2; offset++ {
				projected := occultationProjectedPoint(result[index+offset])
				x += weights[offset+2] * projected[0]
				y += weights[offset+2] * projected[1]
			}
			candidate := occultationUnprojectedPoint(x/35, y/35)
			move := geoDistanceKM(original[index], candidate)
			if move > maximumMoveKM {
				continue
			}
			first := result[index-2]
			last := result[index+2]
			deviation := occultationProjectedPointLineDistanceKM(point, first, last)
			if deviation > maximumDeviationKM {
				continue
			}
			updated[index] = candidate
		}
		result = updated
	}
	if closed {
		result[len(result)-1] = result[0]
	}
	return result
}

func occultationProjectedPoint(point geodata.GeoPoint) [2]float64 {
	latitude := math.Max(-85.05112878, math.Min(85.05112878, point.Latitude)) * math.Pi / 180
	return [2]float64{
		EarthRadiusKM * point.Longitude * math.Pi / 180,
		EarthRadiusKM * math.Log(math.Tan(math.Pi/4+latitude/2)),
	}
}

func occultationUnprojectedPoint(pointX, pointY float64) geodata.GeoPoint {
	longitude := pointX / EarthRadiusKM * 180 / math.Pi
	longitude = math.Mod(longitude+180, 360)
	if longitude < 0 {
		longitude += 360
	}
	return geodata.GeoPoint{
		Longitude: longitude - 180,
		Latitude:  (2*math.Atan(math.Exp(pointY/EarthRadiusKM)) - math.Pi/2) * 180 / math.Pi,
	}
}

func occultationProjectedPointLineDistanceKM(point, first, last geodata.GeoPoint) float64 {
	p := occultationProjectedPoint(point)
	a := occultationProjectedPoint(first)
	b := occultationProjectedPoint(last)
	dx, dy := b[0]-a[0], b[1]-a[1]
	if lengthSquared := dx*dx + dy*dy; lengthSquared > 0 {
		fraction := ((p[0]-a[0])*dx + (p[1]-a[1])*dy) / lengthSquared
		fraction = math.Max(0, math.Min(1, fraction))
		return math.Hypot(p[0]-(a[0]+fraction*dx), p[1]-(a[1]+fraction*dy))
	}
	return math.Hypot(p[0]-a[0], p[1]-a[1])
}

func occultationProjectedTurnAngleDegrees(
	first, middle, last geodata.GeoPoint,
) float64 {
	firstProjected := occultationProjectedPoint(first)
	middleProjected := occultationProjectedPoint(middle)
	lastProjected := occultationProjectedPoint(last)
	firstX, firstY := firstProjected[0]-middleProjected[0], firstProjected[1]-middleProjected[1]
	lastX, lastY := lastProjected[0]-middleProjected[0], lastProjected[1]-middleProjected[1]
	firstLength := math.Hypot(firstX, firstY)
	lastLength := math.Hypot(lastX, lastY)
	if firstLength <= 1e-12 || lastLength <= 1e-12 {
		return 180
	}
	cosine := (firstX*lastX + firstY*lastY) / (firstLength * lastLength)
	cosine = math.Max(-1, math.Min(1, cosine))
	return math.Acos(cosine) * 180 / math.Pi
}

// cleanupOccultationFootprintUnionPolygons keeps every component of the
// horizon-closed instantaneous union. Unlike the compact-band cleaner it does
// not discard small components: at a horizon transition a narrow component is
// still a real visible portion of the time union, not a polygonizer sliver.
func cleanupOccultationFootprintUnionPolygons(
	polygons [][]geodata.GeoPoint,
	strongPolarSmoothing bool,
) [][]geodata.GeoPoint {
	result := make([][]geodata.GeoPoint, 0, len(polygons))
	// A disconnected horizon union is already the authoritative topology. In
	// particular, Saturn's polar samples contain many narrow real components;
	// moving their vertices can bridge a below-horizon cap. Keep those rings
	// fixed and only densify their projected edges for rendering.
	for _, polygon := range polygons {
		if len(openFootprintRing(polygon)) < 3 || math.Abs(geoRingArea(polygon)) <= 1e-12 {
			continue
		}
		ring := densifyOccultationPolygons([][]geodata.GeoPoint{polygon}, 40)[0]
		if len(polygons) == 1 {
			if strongPolarSmoothing {
				ring = removeOccultationHairpins(ring, 190, 25, 32)
			} else {
				ring = removeOccultationHairpins(ring, 100, 25, 32)
			}
		}
		ring = removeOccultationSharpCorners(ring, 20, 30)
		ring = densifyOccultationPolygons([][]geodata.GeoPoint{ring}, 40)[0]
		if len(openFootprintRing(ring)) >= 3 && math.Abs(geoRingArea(ring)) > 1e-12 {
			result = append(result, ring)
		}
	}
	return result
}

// Contact-contour output is already a physical continuous envelope. Clean
// only the short, high-latitude reversals that are numerical polygonizer
// vertices, then densify for display; long physical phase branches remain
// untouched.
func cleanupOccultationAuthoritativeBandPolygons(
	polygons [][]geodata.GeoPoint,
) [][]geodata.GeoPoint {
	polygons = RemoveTinyPolygonComponents(polygons)
	densified := densifyOccultationPolygons(polygons, 25)
	usable := make([][]geodata.GeoPoint, 0, len(densified))
	for _, polygon := range densified {
		open := openFootprintRing(polygon)
		if len(open) < 3 || math.Abs(geoRingArea(open)) <= 1e-12 {
			continue
		}
		usable = append(usable, polygon)
	}
	densified = usable
	for index := range densified {
		densified[index] = removeOccultationPolarKinks(densified[index])
		densified[index] = removeOccultationSharpCorners(densified[index], 60, 45)
		densified[index] = removeOccultationPolarSharpCorners(densified[index], 20, 70, 70)
		densified[index] = smoothOccultationPolarCorners(densified[index])
		densified[index] = smoothOccultationPolarWobbles(densified[index])
		densified[index] = smoothOccultationOrdinaryWobbles(densified[index])
	}
	// The first cleanup can expose a short reversal at the interpolation seam
	// of an otherwise longer edge. Run the local cleanup once more before the
	// final spacing pass so the GeoJSON consumer receives both smooth turns and
	// bounded projected edges.
	densified = densifyOccultationPolygons(densified, 25)
	for index := range densified {
		densified[index] = removeOccultationPolarKinks(densified[index])
		densified[index] = removeOccultationSharpCorners(densified[index], 60, 45)
		densified[index] = removeOccultationPolarSharpCorners(densified[index], 20, 70, 70)
		densified[index] = smoothOccultationPolarCorners(densified[index])
		densified[index] = smoothOccultationPolarWobbles(densified[index])
		densified[index] = smoothOccultationOrdinaryWobbles(densified[index])
	}
	result := make([][]geodata.GeoPoint, 0, len(densified))
	for _, polygon := range densified {
		open := openFootprintRing(polygon)
		if len(open) >= 3 && math.Abs(geoRingArea(open)) > 1e-12 {
			result = append(result, polygon)
		}
	}
	return result
}

// preserveOccultationPhaseBoundaryEnvelope restores only the part of an
// accepted physical phase cycle that bounded smoothing moved outside the
// cleaned polygon. The source cycle is already topology- and witness-checked;
// unioning it back cannot invent visibility, while leaving unrelated portions
// of the cleaned ring unchanged.
func preserveOccultationPhaseBoundaryEnvelope(
	cleaned, physicalBoundary [][]geodata.GeoPoint,
) [][]geodata.GeoPoint {
	if len(cleaned) == 0 || len(physicalBoundary) == 0 {
		return cleaned
	}
	// The phase cycle can close two branches at a shared polar extremum. Unioning
	// that raw closure back into the cleaned footprint restores the very cusp the
	// cleanup removed, so round only that junction before the merge. The rounded
	// candidate is accepted only when it still contains the complete source cycle.
	roundedBoundary := roundOccultationPhaseBoundaryJunctions(physicalBoundary)
	input := append([][]geodata.GeoPoint(nil), cleaned...)
	input = append(input, roundedBoundary...)
	merged, err := geodata.UnionPolygons(input)
	if err != nil || len(merged) == 0 {
		return cleaned
	}
	merged = RemoveTinyPolygonComponents(merged)
	merged = densifyOccultationPolygons(merged, 40)
	return roundOccultationPhaseBoundaryJunctions(merged)
}

// roundOccultationPhaseBoundaryJunctions replaces only a short polar phase
// junction with a projected C1 curve. The replaced source window remains the
// acceptance witness: if the outward fillet cannot contain it within the
// source-specific numerical projection tolerance, the original ring is
// returned unchanged rather than silently shrinking visibility.
func roundOccultationPhaseBoundaryJunctions(
	polygons [][]geodata.GeoPoint,
) [][]geodata.GeoPoint {
	result := make([][]geodata.GeoPoint, len(polygons))
	for index, polygon := range polygons {
		result[index] = roundOccultationPhaseBoundaryRing(polygon)
	}
	return result
}

func roundOccultationAuthoritativeBandJunctions(
	polygons [][]geodata.GeoPoint,
) [][]geodata.GeoPoint {
	result := make([][]geodata.GeoPoint, len(polygons))
	for index, polygon := range polygons {
		rounded := roundOccultationAuthoritativeBandJunctionRing(polygon)
		result[index] = rounded
	}
	return result
}

func roundOccultationPhaseBoundaryRing(points []geodata.GeoPoint) []geodata.GeoPoint {
	return roundOccultationBandJunctionRing(points, 74, 75, 155, false)
}

// roundOccultationAuthoritativeBandJunctionRing rounds a compact high-latitude
// latitude return left by a temporal sweep. Unlike a normal sampled arc, these
// seam vertices reverse latitude over a short chord and are shared by the
// partial and total bands. The candidate remains subject to the source-window
// containment gate in roundOccultationBandJunctionRing, so smoothing preserves
// the local visible envelope within the numerical projection tolerance.
func roundOccultationAuthoritativeBandJunctionRing(points []geodata.GeoPoint) []geodata.GeoPoint {
	return roundOccultationBandJunctionRing(points, 60, 80, 125, true)
}

func roundOccultationTotalBandJunctions(
	polygons [][]geodata.GeoPoint,
) [][]geodata.GeoPoint {
	result := roundOccultationAuthoritativeBandJunctions(polygons)
	for index := range result {
		result[index] = roundOccultationWidePolarShoulder(result[index])
	}
	return result
}

// roundOccultationWidePolarShoulder repairs a broad concave shoulder where a
// horizon phase cap joins the continuously swept inner-contact envelope. The
// ordinary junction rounder above deliberately handles only short chords; a
// several-hundred-kilometre shoulder therefore survives as a visible notch in
// Web Mercator even though every individual vertex has a benign local turn.
//
// This pass is used only by total bands. It requires a high-latitude global
// extremum, a sharp drop in secant slope on one adjacent arc, and a replacement
// which contains the complete source ring. Ordinary convex polar arcs never
// meet that slope signature and are returned unchanged.
func roundOccultationWidePolarShoulder(points []geodata.GeoPoint) []geodata.GeoPoint {
	if len(points) < 16 {
		return points
	}
	closed := geodata.SameGeoPoint(points[0], points[len(points)-1])
	open := append([]geodata.GeoPoint(nil), points...)
	if closed {
		open = open[:len(open)-1]
	}
	if len(open) < 15 {
		return points
	}
	extreme := 0
	for index := 1; index < len(open); index++ {
		if math.Abs(open[index].Latitude) > math.Abs(open[extreme].Latitude) {
			extreme = index
		}
	}
	if math.Abs(open[extreme].Latitude) < 75 {
		return points
	}
	type shoulderCandidate struct {
		ring []geodata.GeoPoint
		gap  float64
	}
	var best shoulderCandidate
	for _, direction := range []int{1, -1} {
		candidate, gap, ok := occultationWidePolarShoulderCandidate(open, extreme, direction)
		if ok && gap > best.gap {
			best = shoulderCandidate{ring: candidate, gap: gap}
		}
	}
	if len(best.ring) == 0 {
		return points
	}
	if closed {
		best.ring = append(best.ring, best.ring[0])
	}
	return best.ring
}

func occultationWidePolarShoulderCandidate(
	open []geodata.GeoPoint,
	extreme, direction int,
) ([]geodata.GeoPoint, float64, bool) {
	const (
		minimumSpanKM        = 150.0
		maximumSpanKM        = 1200.0
		minimumSlopeDrop     = 0.20
		minimumInwardGapKM   = 20.0
		maximumInwardGapKM   = 100.0
		containmentTolerance = 1.0
	)
	n := len(open)
	oriented := make([]geodata.GeoPoint, n)
	for index := range oriented {
		source := (extreme - direction + index*direction) % n
		if source < 0 {
			source += n
		}
		oriented[index] = open[source]
	}
	start := occultationProjectedPoint(oriented[1])
	polarSign := 1.0
	if oriented[1].Latitude > 0 {
		polarSign = -1
	}
	xSign := 0.0
	minimumSlope := math.Inf(1)
	maximumEarlySlope := 0.0
	anchor := -1
	for index := 2; index+1 < len(oriented); index++ {
		point := occultationProjectedPoint(oriented[index])
		dx := point[0] - start[0]
		if xSign == 0 && math.Abs(dx) > 1 {
			xSign = math.Copysign(1, dx)
		}
		if xSign == 0 || dx*xSign <= 0 {
			break
		}
		span := math.Abs(dx)
		if span > maximumSpanKM {
			break
		}
		inward := (point[1] - start[1]) * polarSign
		if inward < -containmentTolerance {
			continue
		}
		slope := inward / span
		if span < minimumSpanKM {
			maximumEarlySlope = math.Max(maximumEarlySlope, slope)
			continue
		}
		if slope < minimumSlope {
			minimumSlope = slope
			anchor = index
		}
	}
	if anchor < 4 || maximumEarlySlope-minimumSlope < minimumSlopeDrop {
		return nil, 0, false
	}
	anchorPoint := occultationProjectedPoint(oriented[anchor])
	anchorSpan := math.Abs(anchorPoint[0] - start[0])
	maximumGap := 0.0
	for index := 2; index < anchor; index++ {
		point := occultationProjectedPoint(oriented[index])
		span := math.Abs(point[0] - start[0])
		if span > anchorSpan {
			return nil, 0, false
		}
		inward := (point[1] - start[1]) * polarSign
		maximumGap = math.Max(maximumGap, inward-minimumSlope*span)
	}
	if maximumGap < minimumInwardGapKM || maximumGap > maximumInwardGapKM {
		return nil, 0, false
	}
	area := geoRingArea(oriented)
	if !finiteGeo(area) || math.Abs(area) <= 1e-12 {
		return nil, 0, false
	}
	outwardSign := 1.0
	if area < 0 {
		outwardSign = -1
	}
	for offset := 0.0; offset <= maximumInwardGapKM; offset += 2.0 {
		candidate := roundOccultationPhaseBoundaryWindow(
			oriented, 1, anchor, outwardSign, offset,
		)
		if len(candidate) < 4 || !geodata.SphericalPolygonsContainPathsWithinKM(
			[][]geodata.GeoPoint{candidate}, [][]geodata.GeoPoint{oriented}, true, containmentTolerance,
		) {
			continue
		}
		if direction < 0 {
			reverseGeoPointRing(candidate)
		}
		return candidate, maximumGap, true
	}
	return nil, 0, false
}

func roundOccultationBandJunctionRing(
	points []geodata.GeoPoint,
	minimumLatitudeDegrees, maximumChordKM, minimumTurnDegrees float64,
	requireLatitudeReturn bool,
) []geodata.GeoPoint {
	if len(points) < 9 {
		return points
	}
	closed := geodata.SameGeoPoint(points[0], points[len(points)-1])
	open := append([]geodata.GeoPoint(nil), points...)
	if closed {
		open = open[:len(open)-1]
		// The polygonizer can place a short polar return across the ring's
		// closing edge. Rotate the closed ring so that the original last
		// vertex becomes an ordinary interior node and is examined by the
		// same junction scan as every other vertex.
		if requireLatitudeReturn && len(open) > 3 {
			// Keep two successors after the original closing pair. The rounder
			// replaces a window and needs one neighbour beyond that window; a
			// one-vertex rotation would still leave a seam candidate at the
			// final index and make it ineligible for smoothing.
			rotated := make([]geodata.GeoPoint, len(open))
			copy(rotated, open[3:])
			copy(rotated[len(open)-3:], open[:3])
			open = rotated
		}
	}
	if len(open) < 9 {
		return points
	}
	area := geoRingArea(open)
	if !finiteGeo(area) || math.Abs(area) <= 1e-12 {
		return points
	}
	// A positive ring area has its interior on the left of traversal; the right
	// normal is therefore outward. Reverse it for a clockwise ring.
	outwardSign := 1.0
	if area < 0 {
		outwardSign = -1
	}
	result := open
	changed := false
	sourceWindowContainmentToleranceKM := 0.05
	if requireLatitudeReturn {
		sourceWindowContainmentToleranceKM = 1.0
	}
	for pass := 0; pass < 8; pass++ {
		// A locally sharp candidate can fail the source-window containment gate
		// even when a neighbouring return is safely roundable. Keep trying the
		// remaining candidates in this pass instead of abandoning the whole ring.
		rejected := make(map[int]bool)
		passAccepted := false
		for {
			candidate := -1
			bestLatitude := 0.0
			bestTurn := 180.0
			for index := 1; index+1 < len(result); index++ {
				if rejected[index] {
					continue
				}
				point := result[index]
				if math.Abs(point.Latitude) < minimumLatitudeDegrees {
					continue
				}
				previous, next := result[index-1], result[index+1]
				if occultationProjectedEdgeDistanceKM(previous, next) > maximumChordKM {
					continue
				}
				turn := occultationTurnAngleDegrees(previous, point, next)
				if turn >= minimumTurnDegrees {
					continue
				}
				if requireLatitudeReturn &&
					((point.Latitude-previous.Latitude)*(next.Latitude-point.Latitude) >= 0 ||
						occultationProjectedTurnAngleDegrees(previous, point, next) >= minimumTurnDegrees) {
					continue
				}
				extreme := math.Abs(point.Latitude)
				if candidate < 0 || turn < bestTurn ||
					(turn == bestTurn && extreme > bestLatitude) {
					candidate = index
					bestLatitude = extreme
					bestTurn = turn
				}
			}
			if candidate < 0 {
				break
			}
			plateauStart, plateauEnd := candidate, candidate
			for plateauStart > 1 && math.Abs(result[plateauStart-1].Latitude-result[candidate].Latitude) <= 1e-5 {
				plateauStart--
			}
			for plateauEnd+1 < len(result)-1 && math.Abs(result[plateauEnd+1].Latitude-result[candidate].Latitude) <= 1e-5 {
				plateauEnd++
			}
			left, right := plateauStart-1, plateauEnd+1
			if left < 1 || right >= len(result)-1 || right-left < 2 {
				rejected[candidate] = true
				continue
			}
			accepted := false
			sourceWindow := append([]geodata.GeoPoint(nil), result[left:right+1]...)
			for offset := 0.0; offset <= 80.0; offset += 2.0 {
				candidateRing := roundOccultationPhaseBoundaryWindow(
					result, left, right, outwardSign, offset,
				)
				if len(candidateRing) < 4 || !geodata.SphericalPolygonsContainPathsWithinKM(
					[][]geodata.GeoPoint{candidateRing}, [][]geodata.GeoPoint{sourceWindow}, false, sourceWindowContainmentToleranceKM,
				) {
					continue
				}
				result = candidateRing
				changed = true
				accepted = true
				passAccepted = true
				break
			}
			if accepted {
				break
			}
			rejected[candidate] = true
		}
		if !passAccepted {
			break
		}
	}
	if !changed {
		return points
	}
	if closed {
		result = append(result, result[0])
	}
	return result
}

func roundOccultationPhaseBoundaryWindow(
	points []geodata.GeoPoint,
	left, right int,
	outwardSign, offsetKM float64,
) []geodata.GeoPoint {
	first := occultationProjectedPoint(points[left])
	last := occultationProjectedPoint(points[right])
	firstNeighbour := occultationProjectedPoint(points[left-1])
	lastNeighbour := occultationProjectedPoint(points[right+1])
	startDX, startDY := first[0]-firstNeighbour[0], first[1]-firstNeighbour[1]
	endDX, endDY := lastNeighbour[0]-last[0], lastNeighbour[1]-last[1]
	startLength, endLength := math.Hypot(startDX, startDY), math.Hypot(endDX, endDY)
	chordLength := math.Hypot(last[0]-first[0], last[1]-first[1])
	if startLength <= 1e-9 || endLength <= 1e-9 || chordLength <= 1e-9 {
		return nil
	}
	startDX, startDY = startDX/startLength, startDY/startLength
	endDX, endDY = endDX/endLength, endDY/endLength
	tangentLength := chordLength / 3
	controlFirst := [2]float64{first[0] + startDX*tangentLength, first[1] + startDY*tangentLength}
	controlLast := [2]float64{last[0] - endDX*tangentLength, last[1] - endDY*tangentLength}
	chordDX, chordDY := last[0]-first[0], last[1]-first[1]
	chordScale := math.Hypot(chordDX, chordDY)
	if chordScale <= 1e-9 {
		return nil
	}
	outwardX, outwardY := chordDY/chordScale, -chordDX/chordScale
	outwardX *= outwardSign
	outwardY *= outwardSign
	steps := int(math.Ceil(chordLength / 5))
	if steps < 8 {
		steps = 8
	}
	if steps > 96 {
		steps = 96
	}
	replacement := make([]geodata.GeoPoint, 0, steps-1)
	for step := 1; step < steps; step++ {
		t := float64(step) / float64(steps)
		u := 1 - t
		weightFirst := u * u * u
		weightControlFirst := 3 * u * u * t
		weightControlLast := 3 * u * t * t
		weightLast := t * t * t
		x := weightFirst*first[0] + weightControlFirst*controlFirst[0] +
			weightControlLast*controlLast[0] + weightLast*last[0]
		y := weightFirst*first[1] + weightControlFirst*controlFirst[1] +
			weightControlLast*controlLast[1] + weightLast*last[1]
		bulge := offsetKM * math.Sin(math.Pi*t) * math.Sin(math.Pi*t)
		x += outwardX * bulge
		y += outwardY * bulge
		replacement = append(replacement, occultationUnprojectedPoint(x, y))
	}
	result := make([]geodata.GeoPoint, 0, len(points)+len(replacement)-right+left)
	result = append(result, points[:left+1]...)
	result = append(result, replacement...)
	result = append(result, points[right:]...)
	return result
}

// CleanAuthoritativeBandPolygons 在调用方执行父子包含 union 等拓扑操作后，重新应用权威掩带的显示清理。
// CleanAuthoritativeBandPolygons reapplies the authoritative display cleanup
// after a caller performs a topology operation such as a parent/child union.
// Those operations can reintroduce the short polar seams removed from the
// original linework.
func CleanAuthoritativeBandPolygons(polygons [][]geodata.GeoPoint) [][]geodata.GeoPoint {
	return cleanupOccultationAuthoritativeBandPolygons(polygons)
}

// RoundAuthoritativeBandJunctions 对已组装的权威掩带应用最终的仅显示接缝平滑。
// RoundAuthoritativeBandJunctions applies the final display-only smoothing to
// an already assembled authoritative band. Callers should invoke it after
// topology operations such as parent/child containment unions, since those
// operations can reintroduce the short polar sweep junction.
func RoundAuthoritativeBandJunctions(polygons [][]geodata.GeoPoint) [][]geodata.GeoPoint {
	return roundOccultationAuthoritativeBandJunctions(polygons)
}

// RoundAuthoritativeTotalBandJunctions 应用常规局部接缝清理和仅限全掩带的宽极区肩部修复。
// RoundAuthoritativeTotalBandJunctions applies the ordinary local seam cleanup
// plus the total-band-only broad polar shoulder repair.
func RoundAuthoritativeTotalBandJunctions(polygons [][]geodata.GeoPoint) [][]geodata.GeoPoint {
	return roundOccultationTotalBandJunctions(polygons)
}

// removeOccultationPolarKinks removes short high-latitude direction reversals
// that are numerical polygonizer vertices rather than physical boundary folds.
// The endpoint/chord limits keep this local: real phase branches and horizon
// connectors span much farther than this pattern, even when they turn sharply.
func removeOccultationPolarKinks(points []geodata.GeoPoint) []geodata.GeoPoint {
	if len(points) < 4 {
		return points
	}
	const (
		// Web Mercator magnifies a 5-10 km spherical return to tens of
		// kilometres at the south-polar latitudes used by occultation maps.
		// Use a wider local window here than the generic corner cleaner, but
		// keep it restricted to the genuinely polar branch where these returns
		// are numerical polygonizer junctions rather than physical curves.
		minimumPolarLatitudeDegrees = 70.0
		maximumAdjacentEdgeKM       = 120.0
		// Never delete a vertex when the resulting rendered chord would itself
		// exceed this local 75 km cleanup budget; the final densification pass
		// restores the output spacing target after the seam is removed.
		maximumEndpointChordKM  = 75.0
		maximumTurnAngleDegrees = 165.0
	)
	result := append([]geodata.GeoPoint(nil), points...)
	for pass := 0; pass < 128; pass++ {
		changed := false
		for index := 1; index+1 < len(result); index++ {
			first, middle, last := result[index-1], result[index], result[index+1]
			if math.Abs(middle.Latitude) < minimumPolarLatitudeDegrees ||
				occultationProjectedEdgeDistanceKM(first, middle) > maximumAdjacentEdgeKM ||
				occultationProjectedEdgeDistanceKM(middle, last) > maximumAdjacentEdgeKM ||
				occultationProjectedEdgeDistanceKM(first, last) > maximumEndpointChordKM {
				continue
			}
			// A reversal in either projected axis is required. A merely curved
			// sample with a small turn angle is still a valid physical boundary.
			longitudeReversal := math.Remainder(middle.Longitude-first.Longitude, 360)*
				math.Remainder(last.Longitude-middle.Longitude, 360) < 0
			latitudeReversal := (middle.Latitude-first.Latitude)*(last.Latitude-middle.Latitude) < 0
			if !longitudeReversal && !latitudeReversal {
				continue
			}
			if occultationTurnAngleDegrees(first, middle, last) >= maximumTurnAngleDegrees {
				continue
			}
			result = append(result[:index], result[index+1:]...)
			changed = true
			break
		}
		if !changed {
			break
		}
	}
	return result
}

func validateOccultationVisibleBandWitnesses(
	visible [][]geodata.GeoPoint,
	useContactContours bool,
	visibleFill, contourFill, visibleFillCoveragePaths [][]geodata.GeoPoint,
) error {
	if !useContactContours || len(visibleFill) == 0 {
		return nil
	}
	if len(visibleFillCoveragePaths) > 0 {
		// Footprint probes are sparse, but they are the cheapest witness set that
		// still tracks the actual visible region rather than only the continuous
		// contact envelope. Catching a wrong-but-valid face here is much cheaper
		// than re-running the full linework search.
		if !geodata.SphericalPolygonsContainPathsWithinKM(visible, visibleFillCoveragePaths, true, 25) {
			return fmt.Errorf("visible contact-contour face misses footprint witnesses by more than 25 km")
		}
	}
	if !geodata.SphericalPolygonsContainPathsWithinKM(visible, visibleFill, true, 50) {
		return fmt.Errorf("visible contact-contour face misses footprint fill by more than 50 km")
	}
	auditFill := contourFill
	if len(auditFill) == 0 {
		auditFill = visibleFill
	}
	if !geodata.SphericalPolygonsContainPathsWithinKM(visible, auditFill, true, 150) {
		return fmt.Errorf("visible contact-contour face misses authoritative fill by more than 150 km")
	}
	return nil
}

func occultationInteriorPolygon(
	source []basic.OccultationPathPoint,
	interior [][]basic.OccultationPathPoint,
) bool {
	for _, candidate := range interior {
		if len(source) != len(candidate) {
			continue
		}
		match := true
		for index := range source {
			if !sameOccultationPoint(source[index], candidate[index]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func occultationStaticInteriorPolygon(
	source []basic.OccultationPathPoint,
	interior [][]basic.OccultationPathPoint,
) bool {
	if !occultationInteriorPolygon(source, interior) {
		return false
	}
	// Center repairs contain a triangular bridge from the physical center to
	// the nearest sampled contact edge plus a small circular cap around the
	// center. Static bands may keep the bridge because it only closes a sampled
	// seam; the radius-based cap is for one-instant rendering and would
	// artificially enlarge the long-lived band.
	open := source
	if len(open) > 1 && sameOccultationPoint(open[0], open[len(open)-1]) {
		open = open[:len(open)-1]
	}
	return len(open) <= 3
}

func mergeStaticFootprintRepairs(
	polygons [][]geodata.GeoPoint,
	footprints []basic.OccultationFootprint,
) [][]geodata.GeoPoint {
	repairs := footprintStaticInteriorPolygons(footprints)
	if len(repairs) == 0 {
		return polygons
	}
	input := append([][]geodata.GeoPoint(nil), polygons...)
	input = append(input, repairs...)
	merged, err := geodata.UnionPolygons(input)
	if err != nil {
		return polygons
	}
	return densifyOccultationPolygons(merged, 50)
}

func occultationLimitVisibleFillPolygons(
	northern, southern []basic.OccultationPathPoint,
) [][]geodata.GeoPoint {
	polygons := make([][]geodata.GeoPoint, 0)
	for _, sampleRange := range ContinuousPairedBoundaryRanges(northern, southern) {
		for index := sampleRange.Start + 1; index < sampleRange.End; index++ {
			source := []basic.OccultationPathPoint{
				northern[index-1],
				northern[index],
				southern[index],
				southern[index-1],
			}
			visibleSource := clipOccultationPolygonToHorizon(source)
			if len(visibleSource) < 3 {
				continue
			}
			polygon := make([]geodata.GeoPoint, len(visibleSource))
			for pointIndex, point := range visibleSource {
				polygon[pointIndex] = geodata.GeoPoint{
					Longitude: point.Longitude,
					Latitude:  point.Latitude,
				}
			}
			polygons = append(polygons, polygon)
		}
	}
	if len(polygons) == 0 {
		return nil
	}
	merged, err := geodata.UnionPolygons(polygons)
	if err != nil {
		return polygons
	}
	return RemoveTinyPolygonComponents(merged)
}

// occultationCurveCoverageProbes adds small interior probes on both sides of
// every physical phase curve. A narrow face between two nearly coincident
// curves can contain real visible footprints while missing all footprint
// centroids; sampling the curve sides lets the linework selector retain that
// face without admitting below-horizon regions.
func occultationCurveCoverageProbes(
	curves []basic.OccultationRiseSetCurve,
	fillPolygons [][]geodata.GeoPoint,
) [][]geodata.GeoPoint {
	if len(curves) == 0 || len(fillPolygons) == 0 {
		return nil
	}
	const maximumSamplesPerSegment = 64
	candidates := make([]geodata.GeoPoint, 0, len(curves)*maximumSamplesPerSegment)
	for _, curve := range curves {
		for _, segment := range curve.Segments {
			if len(segment) < 2 {
				continue
			}
			step := (len(segment) + maximumSamplesPerSegment - 1) / maximumSamplesPerSegment
			if step < 1 {
				step = 1
			}
			for index := 0; index+1 < len(segment); index += step {
				next := index + step
				if next >= len(segment) {
					next = len(segment) - 1
				}
				first, second := segment[index], segment[next]
				deltaLongitude := normalizeGeoLongitude(second.Longitude - first.Longitude)
				deltaLatitude := second.Latitude - first.Latitude
				length := math.Hypot(deltaLongitude, deltaLatitude)
				if length <= 1e-12 {
					continue
				}
				midpoint := geodata.GeoPoint{
					Longitude: normalizeGeoLongitude(first.Longitude + deltaLongitude/2),
					Latitude:  first.Latitude + deltaLatitude/2,
				}
				for _, offset := range []float64{0.00005, 0.0005, 0.002} {
					for _, side := range []float64{-1, 1} {
						candidates = append(candidates, geodata.GeoPoint{
							Longitude: normalizeGeoLongitude(midpoint.Longitude - side*deltaLatitude*offset/length),
							Latitude:  midpoint.Latitude + side*deltaLongitude*offset/length,
						})
					}
				}
			}
		}
	}
	inside := geodata.SphericalPolygonsContainPoints(fillPolygons, candidates)
	probes := make([][]geodata.GeoPoint, 0, len(candidates))
	for index, probe := range candidates {
		if !inside[index] {
			continue
		}
		duplicate := false
		for _, existing := range probes {
			if len(existing) > 0 && geoDistanceKM(existing[0], probe) < 1 {
				duplicate = true
				break
			}
		}
		if !duplicate {
			probes = append(probes, []geodata.GeoPoint{probe})
		}
	}
	return probes
}

const maximumOccultationBoundaryAlternatives = 8
