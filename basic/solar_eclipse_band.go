package basic

import (
	"math"
	"sort"

	"github.com/starainrt/astro/internal/geodata"
)

type solarEclipseCentralBandSweepSample struct {
	jde       float64
	envelope  SolarEclipsePathPoint
	first     SolarEclipsePathPoint
	second    SolarEclipsePathPoint
	firstCap  []SolarEclipsePathPoint
	secondCap []SolarEclipsePathPoint
}

type solarEclipseCentralBandGeometry struct {
	jde  float64
	moon [3]float64
	axis solarEclipseAxis
	sun  [3]float64
}

// Horizon roots use topocentric local-centrality and sunrise/sunset, while
// U1/U4 are geocentric shadow contacts.  Near a polar grazing event the two
// definitions can differ by a few seconds; keep that physical seam from
// rejecting an otherwise converged horizon root.
const solarEclipseCentralLimitHorizonContactMarginDays = 30.0 / 86400.0

// centralLimitHorizonRootsNearAxisContact finds the two physical intersections
// between the central limits and the local-greatest horizon curve. Seeds are
// taken from the on-Earth side of the central-axis contact, independently of
// the caller's global path sampling step.
func (solver solarEclipseSolver) centralLimitHorizonRootsNearAxisContact(
	axisContactJDE, shadowContactJDE, innerContactJDE, direction float64,
) (SolarEclipsePathPoint, SolarEclipsePathPoint, bool) {
	const (
		seedWindowDays  = 300.0 / 86400.0
		seedStepDays    = 15.0 / 86400.0
		rootTimeEpsilon = 0.25 / 86400.0
		rootDistanceKM  = 0.1
	)
	// At a grazing contact, the two horizon roots can be many minutes apart.
	// Both belong to the interval while the shadow intersects the Earth limb.
	rootWindowDays := direction * (innerContactJDE - shadowContactJDE)
	if rootWindowDays <= 0 {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	centers := make([]SolarEclipsePathPoint, 0, int(seedWindowDays/seedStepDays))
	for offset := seedStepDays; offset <= seedWindowDays+seedStepDays/2; offset += seedStepDays {
		if center, ok := solver.centralPathPointAt(axisContactJDE + direction*offset); ok {
			centers = append(centers, center)
		}
	}
	sort.Slice(centers, func(first, second int) bool { return centers[first].JDE < centers[second].JDE })
	northern, southern := solver.centralPathLimits(centers)
	roots := make([]SolarEclipsePathPoint, 0, 2)
	for index := range northern {
		for _, seed := range []SolarEclipsePathPoint{northern[index], southern[index]} {
			root, ok := solveSolarEclipseCentralLimitHorizonRoot(
				solver,
				[3]float64{seed.Longitude, seed.Latitude, seed.JDE},
			)
			insideOffset := direction * (root.JDE - shadowContactJDE)
			if !ok ||
				insideOffset < -solarEclipseCentralLimitHorizonContactMarginDays ||
				insideOffset > rootWindowDays+solarEclipseCentralLimitHorizonContactMarginDays {
				continue
			}
			duplicate := false
			for _, existing := range roots {
				if math.Abs(root.JDE-existing.JDE) <= rootTimeEpsilon &&
					solarEclipsePathDistanceKM(root, existing) <= rootDistanceKM {
					duplicate = true
					break
				}
			}
			if !duplicate {
				roots = append(roots, root)
			}
		}
	}
	if len(roots) != 2 || solarEclipsePathDistanceKM(roots[0], roots[1]) <= 0.01 {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	sort.Slice(roots, func(first, second int) bool { return roots[first].JDE < roots[second].JDE })
	return roots[0], roots[1], true
}

// centralLimitHorizonRootsFromCurves brackets missing grazing roots on the
// already sampled greatest-at-horizon branches, then corrects C and dC/dt.
// Axis-adjacent Newton seeds alone can converge to the opposite end of a
// polar event even when the two limits are well separated.
func (solver solarEclipseSolver) centralLimitHorizonRootsFromCurves(
	curves []SolarEclipseRiseSetCurve,
	firstContactJDE, lastContactJDE float64,
) ([]SolarEclipsePathPoint, RiseSetDirection) {
	startJDE, endJDE := math.Min(firstContactJDE, lastContactJDE), math.Max(firstContactJDE, lastContactJDE)
	if startJDE <= 0 || endJDE <= startJDE {
		return nil, ""
	}
	roots := make([]SolarEclipsePathPoint, 0, 2)
	seededDirection := RiseSetDirection("")
	for _, curve := range curves {
		if curve.Phase != RiseSetPhaseGreatest {
			continue
		}
		for _, segment := range curve.Segments {
			for index := 1; index < len(segment); index++ {
				first, second := segment[index-1], segment[index]
				if second.JDE < startJDE || first.JDE > endJDE {
					continue
				}
				seed, ok := solver.refineNonCentralBandHorizonGapCrossing(first, second)
				if !ok {
					continue
				}
				root, ok := solveSolarEclipseCentralLimitHorizonRoot(
					solver, [3]float64{seed.Longitude, seed.Latitude, seed.JDE},
				)
				if !ok || root.JDE < startJDE-solarEclipseCentralLimitHorizonContactMarginDays ||
					root.JDE > endJDE+solarEclipseCentralLimitHorizonContactMarginDays || solarEclipseRiseSetPointExists(roots, root) {
					continue
				}
				if seededDirection == "" {
					seededDirection = curve.Direction
				}
				roots = append(roots, root)
			}
		}
	}
	sort.Slice(roots, func(first, second int) bool { return roots[first].JDE < roots[second].JDE })
	return roots, seededDirection
}

// centralLimitHorizonRootsFromSampledBoundary recovers a horizon root the
// analytic seeds miss, by intersecting the sampled umbral sweep with the
// greatest-at-horizon curves. A grazing closure arc ends where the swept region
// crosses that curve, and the sweep is already available, so the crossing is a
// far better seed than the axis-adjacent Newton start that fails on these
// events.
func (solver solarEclipseSolver) centralLimitHorizonRootsFromSampledBoundary(
	startJDE, endJDE, greatestJDE float64,
	footprints []SolarEclipsePartialFootprint,
	curves []SolarEclipseRiseSetCurve,
) ([]SolarEclipsePathPoint, RiseSetDirection) {
	low, high := math.Min(startJDE, endJDE), math.Max(startJDE, endJDE)
	if low <= 0 || high <= low || len(footprints) == 0 {
		return nil, ""
	}
	// Sample slightly beyond the contact window: a closure arc can put one of
	// its roots just outside it, and the sweep is cheapest here (the caller's
	// footprints are reused).
	margin := solarEclipseCentralLimitHorizonContactMarginDays
	if span := high - low; span > 0 {
		margin = math.Max(margin, span*0.15)
	}
	rings := solver.centralBandSampledFootprintUnionOverRange(
		low-margin, high+margin, greatestJDE, footprints,
	)
	if len(rings) == 0 {
		return nil, ""
	}
	// A grazing closure arc meets the sweep almost tangentially, so an exact
	// segment crossing is not reliable; every seed below is refined by the same
	// Newton correction and then judged by the contact window.
	seeds := solarEclipseHorizonCurveSeeds(curves, low, high)
	seeds = append(seeds, solarEclipseHorizonBoundarySeeds(rings, curves, greatestJDE)...)
	roots := make([]SolarEclipsePathPoint, 0, 2)
	seededDirection := RiseSetDirection("")
	for _, seed := range seeds {
		root, ok := solveSolarEclipseCentralLimitHorizonRoot(
			solver, [3]float64{seed.point.Longitude, seed.point.Latitude, seed.point.JDE},
		)
		if !ok ||
			root.JDE < low-solarEclipseCentralLimitHorizonContactMarginDays ||
			root.JDE > high+solarEclipseCentralLimitHorizonContactMarginDays ||
			solarEclipseRiseSetPointExists(roots, root) {
			continue
		}
		if seededDirection == "" {
			seededDirection = seed.direction
		}
		roots = append(roots, root)
	}
	sort.Slice(roots, func(first, second int) bool { return roots[first].JDE < roots[second].JDE })
	return roots, seededDirection
}

// solarEclipseHorizonSeed is a starting point for the closure-root correction,
// carrying the horizon direction of the curve it came from.
type solarEclipseHorizonSeed struct {
	point     SolarEclipsePathPoint
	direction RiseSetDirection
}

// solarEclipseHorizonCurveSeeds seeds from the sampled greatest-at-horizon
// branches that lie inside the window, including their endpoints: a sampled
// branch can stop right at the closure root (1136-06-01 loses its second root
// that way).
func solarEclipseHorizonCurveSeeds(
	curves []SolarEclipseRiseSetCurve,
	low, high float64,
) []solarEclipseHorizonSeed {
	var seeds []solarEclipseHorizonSeed
	for _, curve := range curves {
		if curve.Phase != RiseSetPhaseGreatest {
			continue
		}
		for _, segment := range curve.Segments {
			if len(segment) == 0 {
				continue
			}
			for _, endpoint := range []SolarEclipsePathPoint{segment[0], segment[len(segment)-1]} {
				if endpoint.JDE >= low && endpoint.JDE <= high {
					seeds = append(seeds, solarEclipseHorizonSeed{point: endpoint, direction: curve.Direction})
				}
			}
		}
	}
	return seeds
}

// solarEclipseHorizonBoundarySeeds seeds from the sampled sweep boundary: the
// closest approach between a boundary segment and a horizon segment, and the
// ends of every stretch of boundary that hugs the curve. A shallow polar band
// can run along the curve for hundreds of kilometres, and its closure endpoints
// are where it enters and leaves that stretch rather than a closest approach
// along it.
func solarEclipseHorizonBoundarySeeds(
	rings [][]SolarEclipsePathPoint,
	curves []SolarEclipseRiseSetCurve,
	greatestJDE float64,
) []solarEclipseHorizonSeed {
	var seeds []solarEclipseHorizonSeed
	for _, ring := range rings {
		if len(ring) < 3 {
			continue
		}
		near, directions := solarEclipseBoundaryNearRuns(ring, curves, greatestJDE)
		for index := range ring {
			previous := (index + len(ring) - 1) % len(ring)
			if near[index] != near[previous] {
				entry := index
				if !near[index] {
					entry = previous
				}
				seeds = append(seeds, solarEclipseHorizonSeed{
					point: SolarEclipsePathPoint{
						JDE:       greatestJDE,
						Longitude: ring[entry].Longitude,
						Latitude:  ring[entry].Latitude,
					},
					direction: directions[entry],
				})
			}
		}
		for index := 0; index+1 < len(ring); index++ {
			first, second := ring[index], ring[index+1]
			for _, curve := range curves {
				if curve.Phase != RiseSetPhaseGreatest {
					continue
				}
				for _, segment := range curve.Segments {
					for position := 0; position+1 < len(segment); position++ {
						seed, ok := closestSolarEclipseBandApproach(
							first, second, segment[position], segment[position+1],
							solarEclipseCentralLimitHorizonContactSeedKM,
						)
						if ok {
							seeds = append(seeds, solarEclipseHorizonSeed{point: seed, direction: curve.Direction})
						}
					}
				}
			}
		}
	}
	return seeds
}

// solarEclipseBoundaryNearRuns marks the boundary vertices that sit within the
// seed tolerance of a greatest-at-horizon curve, with that curve's direction.
func solarEclipseBoundaryNearRuns(
	ring []SolarEclipsePathPoint,
	curves []SolarEclipseRiseSetCurve,
	greatestJDE float64,
) ([]bool, []RiseSetDirection) {
	near := make([]bool, len(ring))
	directions := make([]RiseSetDirection, len(ring))
	for index, vertex := range ring {
		direction, ok := solarEclipseGreatestCurveWithin(curves, SolarEclipsePathPoint{
			JDE: greatestJDE, Longitude: vertex.Longitude, Latitude: vertex.Latitude,
		}, solarEclipseCentralLimitHorizonContactSeedKM)
		if !ok {
			continue
		}
		near[index] = true
		directions[index] = direction
	}
	return near, directions
}

// solarEclipseGreatestCurveWithin reports the horizon direction of the sampled
// greatest-at-horizon point closest to one location, when it lies within the
// tolerance. A grazing closure root lies on one of those curves by
// construction, so the nearest sample already knows whether the Sun is rising
// or setting there, which the local classification cannot always reproduce.
func solarEclipseGreatestCurveWithin(
	curves []SolarEclipseRiseSetCurve,
	point SolarEclipsePathPoint,
	toleranceKM float64,
) (RiseSetDirection, bool) {
	bestDistance := toleranceKM
	direction := RiseSetDirection("")
	for _, curve := range curves {
		if curve.Phase != RiseSetPhaseGreatest || curve.Direction == "" {
			continue
		}
		for _, segment := range curve.Segments {
			for _, sample := range segment {
				distance := solarEclipsePathDistanceKM(point, sample)
				if distance <= bestDistance {
					bestDistance = distance
					direction = curve.Direction
				}
			}
		}
	}
	return direction, direction != ""
}

// solarEclipseNearestGreatestDirection reports the horizon direction of the
// sampled greatest-at-horizon curve closest to one closure root, at any
// distance.
func solarEclipseNearestGreatestDirection(
	curves []SolarEclipseRiseSetCurve,
	point SolarEclipsePathPoint,
) RiseSetDirection {
	direction, _ := solarEclipseGreatestCurveWithin(curves, point, math.Inf(1))
	return direction
}

// solarEclipseCentralLimitHorizonContactSeedKM bounds how far a sampled sweep
// boundary may sit from a horizon curve and still seed a closure root.
const solarEclipseCentralLimitHorizonContactSeedKM = 40.0

// closestSolarEclipseBandApproach returns the point of the horizon segment that
// comes closest to one band segment, when the two nearly touch. A grazing
// closure arc leaves the swept region almost tangentially, so requiring an
// exact crossing would miss it.
func closestSolarEclipseBandApproach(
	first, second, horizonFirst, horizonSecond SolarEclipsePathPoint,
	toleranceKM float64,
) (SolarEclipsePathPoint, bool) {
	scale := math.Cos(horizonFirst.Latitude * math.Pi / 180)
	ax := math.Remainder(first.Longitude-horizonFirst.Longitude, 360) * scale
	ay := first.Latitude - horizonFirst.Latitude
	bx := math.Remainder(second.Longitude-horizonFirst.Longitude, 360) * scale
	by := second.Latitude - horizonFirst.Latitude
	dx := math.Remainder(horizonSecond.Longitude-horizonFirst.Longitude, 360) * scale
	dy := horizonSecond.Latitude - horizonFirst.Latitude
	length := dx*dx + dy*dy
	if length <= 0 {
		return SolarEclipsePathPoint{}, false
	}
	bestDistance := math.Inf(1)
	bestFraction := 0.0
	for step := 0; step <= 8; step++ {
		bandFraction := float64(step) / 8
		pointX := ax + bandFraction*(bx-ax)
		pointY := ay + bandFraction*(by-ay)
		fraction := math.Max(0, math.Min(1, (pointX*dx+pointY*dy)/length))
		distance := math.Hypot(pointX-fraction*dx, pointY-fraction*dy)
		if distance < bestDistance {
			bestDistance = distance
			bestFraction = fraction
		}
	}
	if bestDistance*111.32 > toleranceKM {
		return SolarEclipsePathPoint{}, false
	}
	return SolarEclipsePathPoint{
		JDE:       horizonFirst.JDE + bestFraction*(horizonSecond.JDE-horizonFirst.JDE),
		Longitude: geodata.NormalizeLongitude(horizonFirst.Longitude + bestFraction*math.Remainder(horizonSecond.Longitude-horizonFirst.Longitude, 360)),
		Latitude:  horizonFirst.Latitude + bestFraction*(horizonSecond.Latitude-horizonFirst.Latitude),
	}, true
}

// magnitudeOneHorizonRoots pairs the early or late endpoints of the two
// complete totality branches. Approximate shadow-contact times must not clip
// these independently solved topocentric horizon roots.
func magnitudeOneHorizonRoots(
	segments [][]SolarEclipsePathPoint,
	direction float64,
) (SolarEclipsePathPoint, SolarEclipsePathPoint, bool) {
	if len(segments) != 2 || (direction != 1 && direction != -1) {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	var roots [2]SolarEclipsePathPoint
	horizon := [2]SolarEclipsePathPoint{}
	horizonAvailable := [2]bool{}
	for index, segment := range segments {
		if len(segment) < 2 {
			return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
		}
		first, last := segment[0], segment[len(segment)-1]
		if !finite(first.JDE) || !finite(first.Longitude) || !finite(first.Latitude) || !finite(first.SunAltitude) ||
			!finite(last.JDE) || !finite(last.Longitude) || !finite(last.Latitude) || !finite(last.SunAltitude) {
			return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
		}
		if first.JDE > last.JDE {
			first, last = last, first
		}
		if last.JDE-first.JDE <= solarEclipsePathDuplicateTimeDays {
			return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
		}
		if math.Abs(first.SunAltitude) <= 1e-5 {
			horizon[index], horizonAvailable[index] = first, true
		} else if math.Abs(last.SunAltitude) <= 1e-5 {
			horizon[index], horizonAvailable[index] = last, true
		}
		candidate := first
		if direction < 0 {
			candidate = last
		}
		// A very shallow polar branch can terminate at the contour's
		// interior sampling limit instead of the horizon. In that case the
		// opposite endpoint is the physical horizon root.
		if math.Abs(candidate.SunAltitude) > 1e-5 {
			alternate := last
			if candidate.JDE == last.JDE {
				alternate = first
			}
			if math.Abs(alternate.SunAltitude) > 1e-5 {
				return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
			}
			candidate = alternate
		}
		roots[index] = candidate
	}
	// Extremely shallow two-limit events can expose only one horizon endpoint
	// on each complete magnitude-one branch. Those two endpoints are the
	// physical pair for both temporal closures.
	if horizonAvailable[0] && horizonAvailable[1] {
		if math.Abs(roots[0].SunAltitude) > 1e-5 || math.Abs(roots[1].SunAltitude) > 1e-5 {
			roots = horizon
		}
	}
	if solarEclipsePathDistanceKM(roots[0], roots[1]) <= 0.01 {
		return SolarEclipsePathPoint{}, SolarEclipsePathPoint{}, false
	}
	if roots[0].JDE > roots[1].JDE {
		roots[0], roots[1] = roots[1], roots[0]
	}
	return roots[0], roots[1], true
}

func (solver solarEclipseSolver) centralBandHorizonClosure(
	firstRoot, lastRoot SolarEclipsePathPoint,
	direction RiseSetDirection,
	curves []SolarEclipseRiseSetCurve,
) []SolarEclipsePathPoint {
	closure := []SolarEclipsePathPoint{firstRoot}
	var selected []SolarEclipsePathPoint
	bestScore := math.Inf(1)
	for _, curve := range curves {
		if curve.Phase != RiseSetPhaseGreatest || curve.Direction != direction {
			continue
		}
		for _, segment := range curve.Segments {
			if len(segment) < 2 {
				continue
			}
			first, firstOK := solarEclipseHorizonSegmentPointAt(segment, firstRoot.JDE)
			last, lastOK := solarEclipseHorizonSegmentPointAt(segment, lastRoot.JDE)
			if !firstOK || !lastOK {
				continue
			}
			// A phase can have several disconnected horizon branches with the
			// same time range. Select the branch that actually joins both solved
			// central-limit roots; concatenating every time-overlapping branch
			// creates the large 2010/2056 endpoint folds seen on the map.
			score := solarEclipsePathDistanceKM(first, firstRoot) +
				solarEclipsePathDistanceKM(last, lastRoot)
			if score >= bestScore {
				continue
			}
			bestScore = score
			selected = segment
		}
	}
	if len(selected) > 0 {
		for _, point := range selected {
			if point.JDE > firstRoot.JDE && point.JDE < lastRoot.JDE {
				closure = append(closure, point)
			}
		}
	}
	closure = append(closure, lastRoot)
	closure = deduplicateSolarEclipsePathPoints(closure)
	refined := make([]SolarEclipsePathPoint, 1, len(closure))
	refined[0] = closure[0]
	for index := 1; index < len(closure); index++ {
		refined = solver.appendRefinedSolarEclipseCentralHorizonSegment(
			refined, closure[index-1], closure[index], 0,
		)
	}
	return deduplicateSolarEclipsePathPoints(refined)
}

// alignSolarEclipseCentralBandHorizonClosures makes the public greatest-at-
// horizon curves share the exact arcs used to close a two-limit central band.
// Without this replacement, the band uses refined roots while the rendered
// greatest curve keeps its coarser samples, leaving a visible seam at both
// endpoints even though the two calculations describe the same boundary.
func (solver solarEclipseSolver) alignSolarEclipseCentralBandHorizonClosures(
	curves []SolarEclipseRiseSetCurve,
	closures [][]SolarEclipsePathPoint,
) {
	for _, closure := range closures {
		if len(closure) < 2 {
			continue
		}
		root := closure[0]
		_, key, ok := solver.magnitudeEvaluationAt(root.JDE).classify(root.Longitude, root.Latitude, true)
		if ok {
			alignSolarEclipseCentralBandHorizonClosure(curves, closure, key.direction)
		}
	}
}

func alignSolarEclipseCentralBandHorizonClosure(
	curves []SolarEclipseRiseSetCurve,
	closure []SolarEclipsePathPoint,
	direction RiseSetDirection,
) {
	if len(closure) < 2 {
		return
	}
	ordered := append([]SolarEclipsePathPoint(nil), closure...)
	if ordered[0].JDE > ordered[len(ordered)-1].JDE {
		for left, right := 0, len(ordered)-1; left < right; left, right = left+1, right-1 {
			ordered[left], ordered[right] = ordered[right], ordered[left]
		}
	}
	start, end := ordered[0], ordered[len(ordered)-1]
	bestCurve, bestSegment := -1, -1
	bestScore := math.Inf(1)
	for curveIndex := range curves {
		curve := &curves[curveIndex]
		if curve.Phase != RiseSetPhaseGreatest || curve.Direction != direction {
			continue
		}
		for segmentIndex, segment := range curve.Segments {
			if len(segment) < 2 ||
				start.JDE < segment[0].JDE-solarEclipseRiseSetTimeEpsilonDays ||
				end.JDE > segment[len(segment)-1].JDE+solarEclipseRiseSetTimeEpsilonDays {
				continue
			}
			segmentStart, startOK := solarEclipseHorizonSegmentPointAt(segment, start.JDE)
			segmentEnd, endOK := solarEclipseHorizonSegmentPointAt(segment, end.JDE)
			if !startOK || !endOK {
				continue
			}
			score := solarEclipsePathDistanceKM(segmentStart, start) +
				solarEclipsePathDistanceKM(segmentEnd, end)
			if score < bestScore {
				bestCurve, bestSegment, bestScore = curveIndex, segmentIndex, score
			}
		}
	}
	if bestCurve < 0 || bestScore > 2*solarEclipseRiseSetTargetSpacingKM {
		return
	}
	segment := curves[bestCurve].Segments[bestSegment]
	joined := make([]SolarEclipsePathPoint, 0, len(segment)+len(ordered))
	for _, point := range segment {
		if point.JDE < start.JDE-solarEclipseRiseSetTimeEpsilonDays {
			joined = append(joined, point)
		}
	}
	joined = append(joined, ordered...)
	for _, point := range segment {
		if point.JDE > end.JDE+solarEclipseRiseSetTimeEpsilonDays {
			joined = append(joined, point)
		}
	}
	curves[bestCurve].Segments[bestSegment] = deduplicateSolarEclipsePathPoints(joined)
}

func solarEclipseHorizonSegmentPointAt(
	segment []SolarEclipsePathPoint, jde float64,
) (SolarEclipsePathPoint, bool) {
	if len(segment) < 2 {
		return SolarEclipsePathPoint{}, false
	}
	best := SolarEclipsePathPoint{}
	bestDistance := math.Inf(1)
	for index := 1; index < len(segment); index++ {
		first, second := segment[index-1], segment[index]
		if (jde < first.JDE && jde < second.JDE) || (jde > first.JDE && jde > second.JDE) {
			continue
		}
		fraction := 0.0
		if second.JDE != first.JDE {
			fraction = (jde - first.JDE) / (second.JDE - first.JDE)
		}
		fraction = math.Max(0, math.Min(1, fraction))
		candidate := solarEclipsePathSphericalInterpolate(first, second, fraction)
		candidate.JDE = jde
		return candidate, true
	}
	for _, point := range segment {
		if distance := math.Abs(point.JDE - jde); distance < bestDistance {
			best, bestDistance = point, distance
		}
	}
	if bestDistance > 10.0/1440.0 {
		return SolarEclipsePathPoint{}, false
	}
	best.JDE = jde
	return best, true
}

func solveSolarEclipseCentralLimitHorizonRoot(
	solver solarEclipseSolver, coordinates [3]float64,
) (SolarEclipsePathPoint, bool) {
	for iteration := 0; iteration < 24; iteration++ {
		residual, jacobian, ok := solarEclipseCentralLimitHorizonJacobian(
			solver, coordinates, solver.magnitudeEvaluationAt,
		)
		if !ok {
			return SolarEclipsePathPoint{}, false
		}
		if solarEclipseCentralLimitHorizonConverged(residual) {
			if !finite(coordinates[0]) || !finite(coordinates[1]) || !finite(coordinates[2]) ||
				coordinates[1] < -90 || coordinates[1] > 90 {
				return SolarEclipsePathPoint{}, false
			}
			evaluation := solver.magnitudeEvaluationAt(coordinates[2])
			state := evaluation.center.stateAt(coordinates[0]*rad, coordinates[1]*rad, 0)
			if evaluation.centralContactSecondDerivative(coordinates[0], coordinates[1]) <= 0 {
				return SolarEclipsePathPoint{}, false
			}
			return SolarEclipsePathPoint{JDE: coordinates[2], Longitude: geodata.NormalizeLongitude(coordinates[0]), Latitude: coordinates[1], SunAltitude: state.sunAltitudeRad / rad}, true
		}
		delta, ok := solveSolarEclipse3x3(jacobian, [3]float64{-residual[0], -residual[1], -residual[2]})
		if !ok {
			return SolarEclipsePathPoint{}, false
		}
		for index := range coordinates {
			coordinates[index] += delta[index]
		}
		coordinates[0] = geodata.NormalizeLongitude(coordinates[0])
	}
	return SolarEclipsePathPoint{}, false
}

func solarEclipseCentralLimitHorizonConverged(residual [3]float64) bool {
	return math.Abs(residual[0]) < 1e-9 && math.Abs(residual[1]) < 1e-7 && math.Abs(residual[2]) < 1e-9
}

func solarEclipseCentralLimitHorizonJacobian(
	solver solarEclipseSolver, coordinates [3]float64,
	evaluate func(float64) solarEclipseRiseSetEvaluation,
) ([3]float64, [3][3]float64, bool) {
	steps := [3]float64{1e-5, 1e-5, 1.0 / 86400.0}
	// 经度、纬度两列的差分不改变时刻，三处残差共用同一时刻的星历态；只有时间列要换时刻。
	centerEvaluation := evaluate(coordinates[2])
	evaluations := [4]solarEclipseRiseSetEvaluation{
		centerEvaluation, centerEvaluation, centerEvaluation, evaluate(coordinates[2] + steps[2]),
	}
	valueAt := func(evaluation solarEclipseRiseSetEvaluation, longitude, latitude float64) ([3]float64, bool) {
		state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
		return [3]float64{solarEclipseCentralContactGap(state), evaluation.centralContactDerivative(longitude, latitude), state.sunAltitudeRad}, finite(state.sunAltitudeRad)
	}
	residual, ok := valueAt(evaluations[0], coordinates[0], coordinates[1])
	if !ok {
		return [3]float64{}, [3][3]float64{}, false
	}
	jacobian := [3][3]float64{}
	for column, step := range steps {
		shifted := coordinates
		shifted[column] += step
		value, ok := valueAt(evaluations[column+1], shifted[0], shifted[1])
		if !ok {
			return [3]float64{}, [3][3]float64{}, false
		}
		for row := range residual {
			jacobian[row][column] = (value[row] - residual[row]) / step
		}
	}
	return residual, jacobian, true
}

// centralBandSweepPolygons combines the critical envelope with the part of the
// local-greatest horizon arc that lies inside the central-contact condition.
func (solver solarEclipseSolver) centralBandSweepPolygons(
	startJDE, endJDE, greatestJDE float64,
	riseSetCurves []SolarEclipseRiseSetCurve,
) ([][]SolarEclipsePathPoint, []SolarEclipsePathPoint, bool) {
	return solver.centralBandSweepPolygonsReusing(startJDE, endJDE, greatestJDE, riseSetCurves, nil)
}

// centralBandSweepPolygonsReusing is centralBandSweepPolygons with the caller's
// already solved instantaneous footprints, so the sampled reconstruction does
// not solve them a second time when the analytic region fails.
func (solver solarEclipseSolver) centralBandSweepPolygonsReusing(
	startJDE, endJDE, greatestJDE float64,
	riseSetCurves []SolarEclipseRiseSetCurve,
	existing []SolarEclipsePartialFootprint,
) ([][]SolarEclipsePathPoint, []SolarEclipsePathPoint, bool) {
	times, _ := solarEclipsePathSampleTimes(
		startJDE, endJDE, greatestJDE, solarEclipseCentralBandStepDays,
	)
	times = solarEclipseCentralBandContactSampleTimes(times, startJDE, endJDE)
	samples := make([]solarEclipseCentralBandSweepSample, 0, len(times))
	for _, jd := range times {
		sample, ok := solver.centralBandSweepSampleAt(jd)
		if ok {
			samples = append(samples, sample)
		}
	}
	if len(samples) >= 2 {
		refined := solver.refineCentralBandSweepSamples(samples)
		polygon, horizon := solver.nonCentralBandRegion(refined, riseSetCurves, greatestJDE)
		if len(polygon) >= 4 {
			return [][]SolarEclipsePathPoint{polygon}, horizon, false
		}
	}
	return solver.centralBandSampledFootprintUnion(times, existing, greatestJDE), nil, true
}

// centralBandSampledFootprintUnionOverRange rebuilds the sampled-footprint
// union for the complete central-contact interval of one event.
func (solver solarEclipseSolver) centralBandSampledFootprintUnionOverRange(
	startJDE, endJDE, greatestJDE float64,
	existing []SolarEclipsePartialFootprint,
) [][]SolarEclipsePathPoint {
	times, _ := solarEclipsePathSampleTimes(
		startJDE, endJDE, greatestJDE, solarEclipseCentralBandStepDays,
	)
	times = solarEclipseCentralBandContactSampleTimes(times, startJDE, endJDE)
	return solver.centralBandSampledFootprintUnion(times, existing, greatestJDE)
}

// A grazing central event can keep part of the umbral/antumbral rim off the
// Earth over the whole path: every instantaneous footprint then has an open
// arc over the outer samples and a fully closed rim over the middle. The
// analytic envelope and the open-arc sweep both describe only part of that
// shape (and the closed middle samples carry no open arc at all), so the static
// band is rebuilt from the instantaneous footprints themselves. The union of
// every sampled footprint is the central band by definition; a vertex spacing
// keeps the spherical union affordable.
const (
	// The union is a coverage reconstruction, not a display-resolution trace:
	// its rings are only ever unioned and exported, so they are sampled well
	// below the footprint resolution. Neighbouring grazing footprints each
	// extend for a thousand kilometres, so the temporal stride only has to be
	// shorter than the along-track overlap; the contact ends keep every sample.
	solarEclipseCentralBandUnionBoundaryPoints = 90
	solarEclipseCentralBandUnionMinStepSeconds = 10.0
	// The union needs a new footprint only once the previous one has moved far
	// enough for the envelope between them to stay resolved, so the stride
	// follows the shadow's ground speed instead of the clock: a grazing event
	// spends most of its contact interval creeping across the terminator, and a
	// fixed stride over-samples exactly there. The contact anchors below keep
	// the flared ends at full resolution.
	// The advance target matches the union spacing: the caller already samples
	// the contact interval at roughly this step, and a coarser stride measurably
	// opens the sweep (1042-06-20 leaves its umbral sweep by 86 km at 20 km).
	// The rule therefore only bites when a caller passes a sparse grid.
	solarEclipseCentralBandUnionAdvanceKM      = 4.0
	solarEclipseCentralBandUnionMaxStepSeconds = 60.0
	solarEclipseCentralBandUnionAnchorSamples  = 3
	solarEclipseCentralBandUnionSpacingKM      = 10.0
	// The caller already solved the contact-interval footprints; a sample within
	// this tolerance of one of them is reused instead of solved again, so only
	// the interval the caller skipped is actually computed here.
	solarEclipseCentralBandUnionReuseDays = 6.0 / 86400.0
	// A decimated union is not an analytic envelope, so the reconstructed band
	// may cut inside the sharpest tip by the union spacing instead of the
	// footprint tolerance. The caller validates it with this bound.
	solarEclipseCentralBandUnionContainmentToleranceKM = 40.0
	// Each pair of neighbouring instantaneous rims meets in a shallow cusp, and
	// where the rims run nearly parallel the union alternates between them and
	// reads as a staircase even though the physical envelope is smooth. Many
	// lambda|mu passes flatten that sampling sweep without the systematic inward
	// shrink a plain Laplacian or a wide quadratic fit would introduce.
	solarEclipseCentralBandUnionSmoothPasses = 30
	solarEclipseCentralBandUnionSmoothLambda = 0.5
	solarEclipseCentralBandUnionSmoothMu     = -0.53
	// A spherical union leaves nodes where two rims nearly touch: vertices a few
	// hundred metres apart that carry no geometry but read as 150-degree spikes.
	solarEclipseCentralBandUnionMinVertexSpacingKM = 1.0
	// The relaxation can pull two nearby vertices onto opposite sides of the
	// underlying curve, which turns a harmless pair into a spike. A second,
	// coarser cleanup after the filter removes those pairs.
	solarEclipseCentralBandUnionPostSmoothSpacingKM = 3.0
)

// solarEclipseRingCenter returns the mean position of one closed ring.
func solarEclipseRingCenter(ring []geodata.GeoPoint) geodata.GeoPoint {
	if len(ring) == 0 {
		return geodata.GeoPoint{}
	}
	longitude, latitude := 0.0, 0.0
	reference := ring[0].Longitude
	for _, point := range ring {
		longitude += reference + math.Remainder(point.Longitude-reference, 360)
		latitude += point.Latitude
	}
	total := float64(len(ring))
	return geodata.GeoPoint{
		Longitude: geodata.NormalizeLongitude(longitude / total),
		Latitude:  latitude / total,
	}
}

// centralBandSampledFootprintUnion samples the instantaneous central-shadow
// footprint over the complete central-contact interval and returns their
// spherical union as closed rings.
func (solver solarEclipseSolver) centralBandSampledFootprintUnion(
	times []float64,
	existing []SolarEclipsePartialFootprint,
	greatestJDE float64,
) [][]SolarEclipsePathPoint {
	rings := make([][]geodata.GeoPoint, 0, len(times))
	ringJDE := make(map[geodata.GeoPoint]float64, len(times)*8)
	solved := append([]SolarEclipsePartialFootprint(nil), existing...)
	sort.Slice(solved, func(first, second int) bool { return solved[first].JDE < solved[second].JDE })
	existingIndex := 0
	lastKept := 0.0
	var lastCenter geodata.GeoPoint
	speedKMperSecond := 0.0
	for index, jde := range times {
		anchored := index < solarEclipseCentralBandUnionAnchorSamples ||
			index >= len(times)-solarEclipseCentralBandUnionAnchorSamples
		if !anchored && lastKept > 0 {
			elapsed := (jde - lastKept) * 86400
			if elapsed < solarEclipseCentralBandUnionMinStepSeconds {
				continue
			}
			// Skip ahead while the previous footprint has barely moved: the
			// swept envelope between two samples this close is already covered.
			if speedKMperSecond > 0 &&
				elapsed*speedKMperSecond < solarEclipseCentralBandUnionAdvanceKM {
				continue
			}
			if elapsed > solarEclipseCentralBandUnionMaxStepSeconds {
				// Never let the stride stretch past the cap, so a stalled centre
				// cannot leave a gap in the sweep.
				if lastKept+solarEclipseCentralBandUnionMaxStepSeconds/86400 < jde {
					jde = lastKept + solarEclipseCentralBandUnionMaxStepSeconds/86400
				}
			}
		}
		for existingIndex < len(solved) &&
			solved[existingIndex].JDE < jde-solarEclipseCentralBandUnionReuseDays {
			existingIndex++
		}
		var ring []SolarEclipsePathPoint
		if existingIndex < len(solved) &&
			math.Abs(solved[existingIndex].JDE-jde) <= solarEclipseCentralBandUnionReuseDays {
			ring = solarEclipseFootprintBoundaryRing(solved[existingIndex])
		} else {
			ring = solver.centralBandSampledFootprintRingAt(jde)
		}
		if len(ring) < 4 {
			continue
		}
		ring = decimateSolarEclipseClosedRing(ring, solarEclipseCentralBandUnionSpacingKM)
		if len(ring) < 4 {
			continue
		}
		polygon := make([]geodata.GeoPoint, len(ring))
		for position, point := range ring {
			polygon[position] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			// 同一坐标可能同时落在相邻两个足迹的环上，取它最早出现的那次，
			// 也就是本影边缘真正扫过它的时刻。
			if _, exists := ringJDE[polygon[position]]; !exists {
				ringJDE[polygon[position]] = point.JDE
			}
		}
		rings = append(rings, polygon)
		if lastKept > 0 && jde > lastKept {
			center := solarEclipseRingCenter(polygon)
			if speedKMperSecond <= 0 {
				speedKMperSecond = solarEclipsePathDistanceKM(
					SolarEclipsePathPoint{Longitude: lastCenter.Longitude, Latitude: lastCenter.Latitude},
					SolarEclipsePathPoint{Longitude: center.Longitude, Latitude: center.Latitude},
				) / ((jde - lastKept) * 86400)
			}
			lastCenter = center
		} else {
			lastCenter = solarEclipseRingCenter(polygon)
		}
		lastKept = jde
	}
	if len(rings) == 0 {
		return nil
	}
	merged := unionSolarEclipseCentralBandRings(rings)
	if len(merged) == 0 {
		return nil
	}
	segments := make([][]SolarEclipsePathPoint, 0, len(merged))
	for _, polygon := range merged {
		if len(polygon) < 4 {
			continue
		}
		times := solarEclipseRingVertexTimes(polygon, ringJDE, greatestJDE)
		polygon, times = dedupeSolarEclipseClosedRing(polygon, times, solarEclipseCentralBandUnionMinVertexSpacingKM)
		smoothed, smoothedTimes := smoothSolarEclipseClosedRing(
			polygon, times, solarEclipseCentralBandUnionSmoothPasses,
		)
		if len(smoothed) >= 4 {
			polygon, times = dedupeSolarEclipseClosedRing(
				smoothed, smoothedTimes, solarEclipseCentralBandUnionPostSmoothSpacingKM,
			)
		}
		segment := make([]SolarEclipsePathPoint, 0, len(polygon)+1)
		for index, point := range polygon {
			segment = append(segment, SolarEclipsePathPoint{
				JDE: times[index], Longitude: point.Longitude, Latitude: point.Latitude,
			})
		}
		if solarEclipsePathDistanceKM(segment[0], segment[len(segment)-1]) > 0.01 {
			segment = append(segment, segment[0])
		}
		segments = append(segments, segment)
	}
	if len(segments) == 0 {
		return nil
	}
	return segments
}

// unionSolarEclipseCentralBandRings merges the sampled footprints one at a
// time. A single boolean join of a few hundred grazing footprints leaves a
// handful of nanodegree seams between temporally adjacent rims, and the
// all-at-once join then reports the outer ring as open; folding the rings into
// a growing accumulator keeps every join local, so the few rings that still
// fail are skipped instead of discarding the whole band.
func unionSolarEclipseCentralBandRings(rings [][]geodata.GeoPoint) [][]geodata.GeoPoint {
	if len(rings) == 0 {
		return nil
	}
	accumulator, err := geodata.UnionPolygons(rings[:1])
	if err != nil || len(accumulator) == 0 {
		return nil
	}
	for _, ring := range rings[1:] {
		merged, mergeErr := geodata.UnionPolygons(append(accumulator, ring))
		if mergeErr != nil || len(merged) == 0 {
			continue
		}
		accumulator = merged
	}
	return accumulator
}

// centralBandSampledFootprintRingAt returns one instantaneous central-shadow
// footprint as a closed ring. The traced boundary is used as-is, so every
// sampled footprint vertex stays on the reconstructed band and the coverage
// audit keeps its meaning; an open (horizon-cut) footprint is closed by the
// chord between its two rim ends, and the neighbouring samples cover the
// horizon side.
func (solver solarEclipseSolver) centralBandSampledFootprintRingAt(
	jde float64,
) []SolarEclipsePathPoint {
	return solarEclipseFootprintBoundaryRing(solver.shadowFootprintAtWithSpacing(
		jde, solarEclipseCentralBandUnionBoundaryPoints, solarEclipseCentralShadow,
		solarEclipseCentralBandUnionSpacingKM,
	))
}

// solarEclipseFootprintBoundaryRing joins one instantaneous footprint into a
// closed ring. An antimeridian-crossing rim is split into two segments; they are
// joined the same way the renderers do instead of adding a false chord across
// 180, and an open (horizon-cut) rim is closed by the chord between its two
// ends while the neighbouring samples cover the horizon side.
func solarEclipseFootprintBoundaryRing(
	footprint SolarEclipsePartialFootprint,
) []SolarEclipsePathPoint {
	if len(footprint.Boundaries) == 0 {
		return nil
	}
	segments := make([][]geodata.GeoPoint, 0, len(footprint.Boundaries))
	for _, boundary := range footprint.Boundaries {
		segment := make([]geodata.GeoPoint, len(boundary))
		for index, point := range boundary {
			segment[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		segments = append(segments, segment)
	}
	joined := geodata.JoinPolylineSegments(segments)
	if len(joined) < 3 {
		return nil
	}
	ring := make([]SolarEclipsePathPoint, 0, len(joined)+1)
	for _, point := range joined {
		ring = append(ring, SolarEclipsePathPoint{
			JDE: footprint.JDE, Longitude: point.Longitude, Latitude: point.Latitude,
		})
	}
	if solarEclipsePathDistanceKM(ring[0], ring[len(ring)-1]) > 0.01 {
		ring = append(ring, ring[0])
	}
	return ring
}

// dedupeSolarEclipseClosedRing drops vertices closer together than the given
// spacing, returns a ring that closes on its first vertex, and keeps the
// parallel time slice aligned so every retained vertex keeps its own time.
func dedupeSolarEclipseClosedRing(
	points []geodata.GeoPoint,
	times []float64,
	minimumKM float64,
) ([]geodata.GeoPoint, []float64) {
	if len(points) < 4 || minimumKM <= 0 || len(times) != len(points) {
		return points, times
	}
	open, openTimes := points, times
	if geodata.EquatorialDistanceKM(points[len(points)-1], points[0]) < minimumKM {
		open, openTimes = points[:len(points)-1], times[:len(times)-1]
	}
	if len(open) < 3 {
		return points, times
	}
	result := make([]geodata.GeoPoint, 0, len(open)+1)
	resultTimes := make([]float64, 0, len(open)+1)
	result = append(result, open[0])
	resultTimes = append(resultTimes, openTimes[0])
	for index, point := range open[1:] {
		if geodata.EquatorialDistanceKM(result[len(result)-1], point) >= minimumKM {
			result = append(result, point)
			resultTimes = append(resultTimes, openTimes[index+1])
		}
	}
	if len(result) < 3 {
		return points, times
	}
	result = append(result, result[0])
	resultTimes = append(resultTimes, resultTimes[0])
	return result, resultTimes
}

// solarEclipseRingVertexTimes 给并集环的每个顶点配回它自己的时间：并集输出保留了
// 原足迹顶点，交点顶点没有直接时间，用环上前后两个已知时间按沿环距离插值。
func solarEclipseRingVertexTimes(
	ring []geodata.GeoPoint,
	ringJDE map[geodata.GeoPoint]float64,
	fallbackJDE float64,
) []float64 {
	times := make([]float64, len(ring))
	known := make([]bool, len(ring))
	found := false
	for index, point := range ring {
		if jde, ok := ringJDE[point]; ok {
			times[index], known[index], found = jde, true, true
		}
	}
	if !found {
		for index := range times {
			times[index] = fallbackJDE
		}
		return times
	}
	if len(ring) > 1 && ring[0] == ring[len(ring)-1] {
		if known[0] {
			known[len(ring)-1], times[len(ring)-1] = true, times[0]
		}
		if known[len(ring)-1] {
			known[0], times[0] = true, times[len(ring)-1]
		}
	}
	count := len(ring)
	for offset := 0; offset < count; offset++ {
		if known[offset] {
			continue
		}
		backIndex, backDistance := offset, 0.0
		for steps := 0; steps < count; steps++ {
			previous := (backIndex - 1 + count) % count
			backDistance += geodata.EquatorialDistanceKM(ring[previous], ring[backIndex])
			backIndex = previous
			if known[backIndex] {
				break
			}
		}
		forwardIndex, forwardDistance := offset, 0.0
		for steps := 0; steps < count; steps++ {
			next := (forwardIndex + 1) % count
			forwardDistance += geodata.EquatorialDistanceKM(ring[forwardIndex], ring[next])
			forwardIndex = next
			if known[forwardIndex] {
				break
			}
		}
		fraction := 0.5
		if total := backDistance + forwardDistance; total > 0 {
			fraction = backDistance / total
		}
		times[offset] = times[backIndex] + fraction*(times[forwardIndex]-times[backIndex])
	}
	return times
}

// smoothSolarEclipseClosedRing filters one closed ring with Taubin smoothing.
// A bare Laplacian pass removes short-wavelength ripple but also shrinks a
// convex boundary — which a grazing band cannot afford, because its coverage
// check allows only tens of kilometres — and a wide quadratic fit can fold a
// tight corner. The lambda|mu pair alternates a shrinking with a slightly
// larger inflating step, so a long run of iterations flattens the sampling
// sweep of the footprint union without pulling the band inside the umbra.
// Longitudes are unwrapped first so the filter never averages across the
// antimeridian. The same weights are applied to the parallel time slice, so a
// moved vertex carries the time of the neighbourhood it was moved into instead
// of falling back to the event's greatest eclipse.
func smoothSolarEclipseClosedRing(
	points []geodata.GeoPoint,
	times []float64,
	passes int,
) ([]geodata.GeoPoint, []float64) {
	count := len(points) - 1
	if count < 8 || passes <= 0 || len(times) != len(points) {
		return nil, nil
	}
	longitudes := make([]float64, count)
	latitudes := make([]float64, count)
	vertexTimes := make([]float64, count)
	longitudes[0] = points[0].Longitude
	latitudes[0] = points[0].Latitude
	vertexTimes[0] = times[0]
	for index := 1; index < count; index++ {
		delta := math.Remainder((points[index].Longitude-points[index-1].Longitude)*rad, 2*math.Pi) / rad
		longitudes[index] = longitudes[index-1] + delta
		latitudes[index] = points[index].Latitude
		vertexTimes[index] = times[index]
	}
	nextLongitudes := make([]float64, count)
	nextLatitudes := make([]float64, count)
	nextTimes := make([]float64, count)
	for pass := 0; pass < passes; pass++ {
		weight := solarEclipseCentralBandUnionSmoothLambda
		if pass%2 == 1 {
			weight = solarEclipseCentralBandUnionSmoothMu
		}
		for index := 0; index < count; index++ {
			previous := (index + count - 1) % count
			following := (index + 1) % count
			nextLongitudes[index] = longitudes[index] +
				weight*(0.5*(longitudes[previous]+longitudes[following])-longitudes[index])
			nextLatitudes[index] = latitudes[index] +
				weight*(0.5*(latitudes[previous]+latitudes[following])-latitudes[index])
			nextTimes[index] = vertexTimes[index] +
				weight*(0.5*(vertexTimes[previous]+vertexTimes[following])-vertexTimes[index])
		}
		copy(longitudes, nextLongitudes)
		copy(latitudes, nextLatitudes)
		copy(vertexTimes, nextTimes)
	}
	ring := make([]geodata.GeoPoint, 0, count+1)
	ringTimes := make([]float64, 0, count+1)
	for index := 0; index < count; index++ {
		ring = append(ring, geodata.GeoPoint{
			Longitude: geodata.NormalizeLongitude(longitudes[index]),
			Latitude:  latitudes[index],
		})
		ringTimes = append(ringTimes, vertexTimes[index])
	}
	ring = append(ring, ring[0])
	ringTimes = append(ringTimes, ringTimes[0])
	return ring, ringTimes
}

// decimateSolarEclipseClosedRing thins one closed ring to the requested
// spacing while keeping its first vertex and its closure.
func decimateSolarEclipseClosedRing(
	points []SolarEclipsePathPoint,
	spacingKM float64,
) []SolarEclipsePathPoint {
	if len(points) < 4 || spacingKM <= 0 {
		return points
	}
	result := make([]SolarEclipsePathPoint, 0, len(points))
	result = append(result, points[0])
	for _, point := range points[1:] {
		if solarEclipsePathDistanceKM(result[len(result)-1], point) >= spacingKM {
			result = append(result, point)
		}
	}
	if len(result) < 3 {
		return points
	}
	first := result[0]
	if solarEclipsePathDistanceKM(result[len(result)-1], first) > 0.01 {
		result = append(result, first)
	}
	return result
}

func (solver solarEclipseSolver) nonCentralTotalBandPolygons(
	segments [][]SolarEclipsePathPoint,
	firstContact, lastContact SolarEclipsePathPoint,
	referenceJDE float64,
	riseSetCurves []SolarEclipseRiseSetCurve,
) ([][]SolarEclipsePathPoint, []SolarEclipsePathPoint) {
	contacts := []SolarEclipsePathPoint{firstContact, lastContact}
	for _, segment := range segments {
		boundary, ok := solver.completeNonCentralTotalBandBoundary(
			segment, contacts, referenceJDE,
		)
		if !ok {
			continue
		}
		polygon, horizon := solver.closeNonCentralBandBoundary(boundary, riseSetCurves)
		if len(polygon) >= 4 {
			return [][]SolarEclipsePathPoint{polygon}, horizon
		}
	}
	return nil, nil
}

func (solver solarEclipseSolver) completeNonCentralTotalBandBoundary(
	segment []SolarEclipsePathPoint,
	contacts []SolarEclipsePathPoint,
	referenceJDE float64,
) ([]SolarEclipsePathPoint, bool) {
	if len(segment) < 2 {
		return nil, false
	}
	boundary := append([]SolarEclipsePathPoint(nil), segment...)
	for _, atStart := range []bool{true, false} {
		endpointIndex := len(boundary) - 1
		if atStart {
			endpointIndex = 0
		}
		endpoint := boundary[endpointIndex]
		if math.Abs(endpoint.SunAltitude) <= 1e-5 {
			continue
		}
		contact, ok := nearestNonCentralBandContact(endpoint, contacts)
		if !ok {
			return nil, false
		}
		extension, ok := solver.nonCentralBandContactExtension(
			contact, endpoint, referenceJDE,
		)
		if !ok {
			return nil, false
		}
		if atStart {
			if solarEclipsePathDistanceKM(extension[len(extension)-1], endpoint) > 0.01 {
				for left, right := 0, len(extension)-1; left < right; left, right = left+1, right-1 {
					extension[left], extension[right] = extension[right], extension[left]
				}
			}
			boundary = append(extension[:len(extension)-1], boundary...)
		} else {
			if solarEclipsePathDistanceKM(extension[0], endpoint) > 0.01 {
				for left, right := 0, len(extension)-1; left < right; left, right = left+1, right-1 {
					extension[left], extension[right] = extension[right], extension[left]
				}
			}
			boundary = append(boundary, extension[1:]...)
		}
	}
	return deduplicateSolarEclipsePathPoints(boundary), true
}

func nearestNonCentralBandContact(
	endpoint SolarEclipsePathPoint,
	contacts []SolarEclipsePathPoint,
) (SolarEclipsePathPoint, bool) {
	const maximumTimeGapDays = 5.0 / 1440.0
	best := SolarEclipsePathPoint{}
	bestTimeGap := math.Inf(1)
	for _, contact := range contacts {
		if contact.JDE == 0 {
			continue
		}
		timeGap := math.Abs(contact.JDE - endpoint.JDE)
		if timeGap < bestTimeGap {
			best, bestTimeGap = contact, timeGap
		}
	}
	if bestTimeGap > maximumTimeGapDays ||
		solarEclipsePathDistanceKM(best, endpoint) > 3000 {
		return SolarEclipsePathPoint{}, false
	}
	return best, true
}

func (solver solarEclipseSolver) nonCentralBandContactExtension(
	contact, endpoint SolarEclipsePathPoint,
	referenceJDE float64,
) ([]SolarEclipsePathPoint, bool) {
	startJDE, endJDE := contact.JDE, endpoint.JDE
	if startJDE > endJDE {
		startJDE, endJDE = endJDE, startJDE
	}
	if endJDE-startJDE <= solarEclipsePathDuplicateTimeDays {
		return []SolarEclipsePathPoint{endpoint}, true
	}
	const stepDays = 5.0 / 86400.0
	times := make([]float64, 0, int((endJDE-startJDE)/stepDays)+20)
	for jde := startJDE + stepDays; jde < endJDE-solarEclipsePathDuplicateTimeDays; jde += stepDays {
		times = append(times, jde)
	}
	for second := 1.0; second <= 10; second++ {
		offset := second / 86400.0
		if startJDE+offset < endJDE-solarEclipsePathDuplicateTimeDays {
			times = append(times, startJDE+offset)
		}
		if endJDE-offset > startJDE+solarEclipsePathDuplicateTimeDays {
			times = append(times, endJDE-offset)
		}
	}
	sort.Float64s(times)
	times = uniqueSolarEclipsePathTimes(times)
	samples := make([]solarEclipseCentralBandSweepSample, 0, len(times))
	for _, jde := range times {
		if sample, ok := solver.centralBandSweepSampleAt(jde); ok {
			samples = append(samples, sample)
		}
	}
	samples = solver.refineCentralBandSweepSamples(samples)
	corrected := solver.correctNonCentralBandEnvelopeSamples(samples, referenceJDE)
	if len(corrected) == 0 {
		return nil, false
	}
	result := make([]SolarEclipsePathPoint, 0, len(corrected)+1)
	if contact.JDE <= endpoint.JDE {
		result = append(result, corrected...)
		result = append(result, endpoint)
	} else {
		result = append(result, endpoint)
		result = append(result, corrected...)
	}
	result = deduplicateSolarEclipsePathPoints(result)
	if len(result) < 2 {
		return nil, false
	}
	for index := 1; index < len(result); index++ {
		if solarEclipsePathDistanceKM(result[index-1], result[index]) >
			2*solarEclipseCentralBandTargetSpacingKM {
			return nil, false
		}
	}
	return result, true
}

func solarEclipseCentralBandContactSampleTimes(times []float64, startJDE, endJDE float64) []float64 {
	window := math.Min(solarEclipseCentralBandContactFineWindowDays, (endJDE-startJDE)/4)
	if window <= solarEclipseCentralBandContactFineStepDays {
		return times
	}
	for jd := startJDE + solarEclipseCentralBandContactFineStepDays; jd < startJDE+window; jd += solarEclipseCentralBandContactFineStepDays {
		times = append(times, jd)
	}
	for jd := endJDE - window + solarEclipseCentralBandContactFineStepDays; jd < endJDE; jd += solarEclipseCentralBandContactFineStepDays {
		times = append(times, jd)
	}
	sort.Float64s(times)
	return uniqueSolarEclipsePathTimes(times)
}

func (solver solarEclipseSolver) centralBandSweepSampleAt(
	jde float64,
) (solarEclipseCentralBandSweepSample, bool) {
	moon, axis, sun := solver.besselGeometryAt(jde)
	samples := make([]solarEclipsePartialBoundarySample, solarEclipseCentralBandBoundaryPoints)
	for index := range samples {
		angle := 2 * math.Pi * float64(index) / float64(len(samples))
		point, ok := solver.shadowFootprintPointAt(
			jde, moon, axis, sun, angle, solarEclipseCentralShadow,
		)
		samples[index] = solarEclipsePartialBoundarySample{point: point, ok: ok, angle: angle}
	}
	samples = solver.refineShadowFootprintTransitions(
		jde, moon, axis, sun, samples, solarEclipseCentralShadow,
	)
	samples = solver.refineShadowFootprintSpacing(
		jde, moon, axis, sun, samples, solarEclipseCentralShadow,
		solarEclipseCentralBandTargetSpacingKM,
	)
	arc := solarEclipseLongestOpenShadowArc(samples)
	if len(arc) < 2 {
		return solarEclipseCentralBandSweepSample{}, false
	}
	geometry := solarEclipseCentralBandGeometry{jde: jde, moon: moon, axis: axis, sun: sun}
	angle, ok := solver.centralBandEnvelopeAngle(geometry, arc)
	if !ok {
		return solarEclipseCentralBandSweepSample{}, false
	}
	envelope, ok := solver.centralShadowPointAt(jde, angle)
	if !ok {
		return solarEclipseCentralBandSweepSample{}, false
	}
	// 弧角沿环单向展开，可以超过 2π；包络角取 [0,2π) 支，比较前必须映射到同一支，
	// 否则展开段的样本会被两端同时丢弃。
	envelopeAngle := angle
	if len(arc) > 0 {
		envelopeAngle = solarEclipseArcBranchAngle(angle, arc[0].angle)
	}
	firstCap := []SolarEclipsePathPoint{envelope}
	secondCap := []SolarEclipsePathPoint{envelope}
	for index := len(arc) - 1; index >= 0; index-- {
		if arc[index].angle >= envelopeAngle {
			continue
		}
		firstCap = append(firstCap, arc[index].point)
	}
	for _, sample := range arc {
		if sample.angle <= envelopeAngle {
			continue
		}
		secondCap = append(secondCap, sample.point)
	}
	firstCap = deduplicateSolarEclipsePathPoints(firstCap)
	secondCap = deduplicateSolarEclipsePathPoints(secondCap)
	if len(firstCap) < 2 || len(secondCap) < 2 {
		return solarEclipseCentralBandSweepSample{}, false
	}
	return solarEclipseCentralBandSweepSample{
		jde: jde, envelope: envelope,
		first: firstCap[len(firstCap)-1], second: secondCap[len(secondCap)-1],
		firstCap: firstCap, secondCap: secondCap,
	}, true
}

// solarEclipseArcBranchAngle 把 [0,2π) 的角度映射到以 reference 为起点的那条展开支上。
func solarEclipseArcBranchAngle(angle, reference float64) float64 {
	return reference + math.Remainder(angle-reference, 2*math.Pi)
}

func solarEclipseLongestOpenShadowArc(
	samples []solarEclipsePartialBoundarySample,
) []solarEclipsePartialBoundarySample {
	if len(samples) == 0 {
		return nil
	}
	invalid := -1
	for index, sample := range samples {
		if !sample.ok {
			invalid = index
			break
		}
	}
	if invalid < 0 {
		return nil
	}
	var longest, current []solarEclipsePartialBoundarySample
	for offset := 1; offset <= len(samples); offset++ {
		sample := samples[(invalid+offset)%len(samples)]
		if !sample.ok {
			if len(current) > len(longest) {
				longest = append([]solarEclipsePartialBoundarySample(nil), current...)
			}
			current = nil
			continue
		}
		if len(current) > 0 {
			for sample.angle <= current[len(current)-1].angle {
				sample.angle += 2 * math.Pi
			}
		}
		current = append(current, sample)
	}
	if len(current) > len(longest) {
		longest = current
	}
	return longest
}

func (solver solarEclipseSolver) centralBandEnvelopeAngle(
	geometry solarEclipseCentralBandGeometry,
	arc []solarEclipsePartialBoundarySample,
) (float64, bool) {
	const timeStep = 0.5 / 86400.0
	beforeMoon, beforeAxis, beforeSun := solver.besselGeometryAt(geometry.jde - timeStep)
	afterMoon, afterAxis, afterSun := solver.besselGeometryAt(geometry.jde + timeStep)
	before := solarEclipseCentralBandGeometry{
		jde: geometry.jde - timeStep, moon: beforeMoon, axis: beforeAxis, sun: beforeSun,
	}
	after := solarEclipseCentralBandGeometry{
		jde: geometry.jde + timeStep, moon: afterMoon, axis: afterAxis, sun: afterSun,
	}
	previousAngle := arc[0].angle
	previous, previousOK := solver.centralBandSweepJacobian(geometry, before, after, previousAngle)
	bestAngle, bestValue := previousAngle, math.Abs(previous)
	for index := 1; index < len(arc); index++ {
		angle := arc[index].angle
		value, ok := solver.centralBandSweepJacobian(geometry, before, after, angle)
		if ok && math.Abs(value) < bestValue {
			bestAngle, bestValue = angle, math.Abs(value)
		}
		if previousOK && ok && previous*value <= 0 {
			left, right := previousAngle, angle
			leftValue := previous
			for iteration := 0; iteration < 48 && right-left > 1e-11; iteration++ {
				middle := (left + right) / 2
				middleValue, middleOK := solver.centralBandSweepJacobian(geometry, before, after, middle)
				if !middleOK {
					return 0, false
				}
				if leftValue*middleValue <= 0 {
					right = middle
				} else {
					left, leftValue = middle, middleValue
				}
			}
			return math.Mod((left+right)/2, 2*math.Pi), true
		}
		previousAngle, previous, previousOK = angle, value, ok
	}
	if bestValue <= 1e-5 {
		return math.Mod(bestAngle, 2*math.Pi), true
	}
	return 0, false
}

func (solver solarEclipseSolver) centralBandSweepJacobian(
	geometry, before, after solarEclipseCentralBandGeometry,
	angle float64,
) (float64, bool) {
	const angleStep = 1e-5
	center, centerOK := solver.centralShadowPointAtGeometry(geometry, angle)
	angleBefore, angleBeforeOK := solver.centralShadowPointAtGeometry(geometry, angle-angleStep)
	angleAfter, angleAfterOK := solver.centralShadowPointAtGeometry(geometry, angle+angleStep)
	timeBefore, timeBeforeOK := solver.centralShadowPointAtGeometry(before, angle)
	timeAfter, timeAfterOK := solver.centralShadowPointAtGeometry(after, angle)
	if !centerOK || !angleBeforeOK || !angleAfterOK || !timeBeforeOK || !timeAfterOK {
		return 0, false
	}
	latitudeScale := math.Cos(center.Latitude * rad)
	angleX := math.Remainder(angleAfter.Longitude-angleBefore.Longitude, 360) * latitudeScale
	angleY := angleAfter.Latitude - angleBefore.Latitude
	timeX := math.Remainder(timeAfter.Longitude-timeBefore.Longitude, 360) * latitudeScale
	timeY := timeAfter.Latitude - timeBefore.Latitude
	value := angleX*timeY - angleY*timeX
	return value, finite(value)
}

func (solver solarEclipseSolver) centralShadowPointAtGeometry(
	geometry solarEclipseCentralBandGeometry,
	angle float64,
) (SolarEclipsePathPoint, bool) {
	return solver.shadowFootprintPointAt(
		geometry.jde, geometry.moon, geometry.axis, geometry.sun,
		math.Mod(angle+2*math.Pi, 2*math.Pi), solarEclipseCentralShadow,
	)
}

func (solver solarEclipseSolver) centralShadowPointAt(jde, angle float64) (SolarEclipsePathPoint, bool) {
	moon, axis, sun := solver.besselGeometryAt(jde)
	return solver.shadowFootprintPointAt(
		jde, moon, axis, sun, math.Mod(angle+2*math.Pi, 2*math.Pi), solarEclipseCentralShadow,
	)
}

func (solver solarEclipseSolver) refineCentralBandSweepSamples(
	samples []solarEclipseCentralBandSweepSample,
) []solarEclipseCentralBandSweepSample {
	if len(samples) < 2 {
		return samples
	}
	result := make([]solarEclipseCentralBandSweepSample, 0, len(samples))
	result = append(result, samples[0])
	for index := 1; index < len(samples); index++ {
		result = solver.appendRefinedCentralBandSweepSample(result, samples[index-1], samples[index], 0)
	}
	return result
}

func (solver solarEclipseSolver) appendRefinedCentralBandSweepSample(
	result []solarEclipseCentralBandSweepSample,
	start, end solarEclipseCentralBandSweepSample,
	depth int,
) []solarEclipseCentralBandSweepSample {
	maximumDistance := math.Max(
		solarEclipsePathDistanceKM(start.envelope, end.envelope),
		math.Max(
			solarEclipsePathDistanceKM(start.first, end.first),
			solarEclipsePathDistanceKM(start.second, end.second),
		),
	)
	if depth >= solarEclipseShadowFootprintAdaptiveMaxDepth ||
		maximumDistance <= solarEclipseCentralBandTargetSpacingKM {
		return append(result, end)
	}
	middle, ok := solver.centralBandSweepSampleAt((start.jde + end.jde) / 2)
	if !ok {
		return append(result, end)
	}
	keep := solarEclipsePathDistanceKM(start.first, middle.first) +
		solarEclipsePathDistanceKM(start.second, middle.second)
	reverse := solarEclipsePathDistanceKM(start.first, middle.second) +
		solarEclipsePathDistanceKM(start.second, middle.first)
	if reverse < keep {
		middle.first, middle.second = middle.second, middle.first
		middle.firstCap, middle.secondCap = middle.secondCap, middle.firstCap
	}
	result = solver.appendRefinedCentralBandSweepSample(result, start, middle, depth+1)
	return solver.appendRefinedCentralBandSweepSample(result, middle, end, depth+1)
}

func deduplicateSolarEclipsePathPoints(points []SolarEclipsePathPoint) []SolarEclipsePathPoint {
	if len(points) < 2 {
		return points
	}
	result := make([]SolarEclipsePathPoint, 0, len(points))
	for _, point := range points {
		if len(result) == 0 || solarEclipsePathDistanceKM(result[len(result)-1], point) > 0.001 {
			result = append(result, point)
		}
	}
	return result
}

func normalizeSolarEclipsePartialFootprintOptions(options SolarEclipsePartialFootprintOptions) SolarEclipsePartialFootprintOptions {
	if options.StepDays <= 0 || math.IsNaN(options.StepDays) || math.IsInf(options.StepDays, 0) {
		options.StepDays = solarEclipsePartialFootprintDefaultStepDays
	}
	if options.StepDays < solarEclipsePathMinStepDays {
		options.StepDays = solarEclipsePathMinStepDays
	}
	if options.BoundaryPoints <= 0 {
		options.BoundaryPoints = solarEclipsePartialFootprintDefaultBoundaryPoints
	}
	if options.BoundaryPoints < solarEclipsePartialFootprintMinBoundaryPoints {
		options.BoundaryPoints = solarEclipsePartialFootprintMinBoundaryPoints
	}
	if options.BoundaryPoints > solarEclipsePartialFootprintMaxBoundaryPoints {
		options.BoundaryPoints = solarEclipsePartialFootprintMaxBoundaryPoints
	}
	if options.CentralShadowStepDays <= 0 || math.IsNaN(options.CentralShadowStepDays) || math.IsInf(options.CentralShadowStepDays, 0) {
		options.CentralShadowStepDays = 0
	} else if options.CentralShadowStepDays < solarEclipsePathMinStepDays {
		options.CentralShadowStepDays = solarEclipsePathMinStepDays
	}
	if options.RiseSetStepDays > 0 && options.RiseSetStepDays < solarEclipsePathMinStepDays {
		options.RiseSetStepDays = solarEclipsePathMinStepDays
	}
	if len(options.MagnitudeValues) > 0 {
		values := make([]float64, 0, len(options.MagnitudeValues))
		for _, value := range options.MagnitudeValues {
			if !isFinite(value) || value <= 0 {
				continue
			}
			duplicate := false
			for _, existing := range values {
				if math.Abs(existing-value) <= 1e-12 {
					duplicate = true
					break
				}
			}
			if !duplicate {
				values = append(values, value)
			}
		}
		sort.Float64s(values)
		if len(values) > solarEclipseMagnitudeContourMaxValues {
			values = values[:solarEclipseMagnitudeContourMaxValues]
		}
		options.MagnitudeValues = values
	}
	if len(options.GreatestTimeValues) > 0 {
		values := make([]float64, 0, len(options.GreatestTimeValues))
		for _, value := range options.GreatestTimeValues {
			if !isFinite(value) {
				continue
			}
			duplicate := false
			for _, existing := range values {
				if math.Abs(existing-value) <= 1e-9 {
					duplicate = true
					break
				}
			}
			if !duplicate {
				values = append(values, value)
			}
		}
		sort.Float64s(values)
		if len(values) > greatestTimeContourMaxLevels {
			values = values[:greatestTimeContourMaxLevels]
		}
		options.GreatestTimeValues = values
	}
	return options
}

func (solver solarEclipseSolver) centralPathPoints(
	startJDE, endJDE, greatestJDE float64,
	options SolarEclipsePathOptions,
) ([]SolarEclipsePathPoint, float64) {
	if endJDE < startJDE {
		startJDE, endJDE = endJDE, startJDE
	}
	if startJDE == 0 || endJDE == 0 || endJDE <= startJDE {
		return nil, options.StepDays
	}

	times, stepDays := solarEclipseMovingDiskEngine().sampleTimes(
		startJDE, endJDE, greatestJDE, options.StepDays,
	)

	points := make([]SolarEclipsePathPoint, 0, len(times))
	for _, jde := range times {
		point, ok := solver.centralPathPointAt(jde)
		if ok {
			points = append(points, point)
		}
	}
	// A coarse caller step can place both contact endpoints exactly on the
	// numerical horizon where the Earth intersection is rejected.  Keep the
	// public center line usable by making one bounded fallback pass through the
	// event interval instead of returning a single greatest-point sample.
	if len(points) < 2 {
		fallbackCount := 32
		for index := 0; index <= fallbackCount; index++ {
			jde := startJDE + (endJDE-startJDE)*float64(index)/float64(fallbackCount)
			point, ok := solver.centralPathPointAt(jde)
			if !ok {
				continue
			}
			duplicate := false
			for _, existing := range points {
				if math.Abs(existing.JDE-jde) <= solarEclipsePathDuplicateTimeDays {
					duplicate = true
					break
				}
			}
			if !duplicate {
				points = append(points, point)
			}
		}
		sort.Slice(points, func(i, j int) bool { return points[i].JDE < points[j].JDE })
		if fallbackStep := (endJDE - startJDE) / float64(fallbackCount); fallbackStep > 0 && fallbackStep < stepDays {
			stepDays = fallbackStep
		}
	}
	sort.Slice(points, func(i, j int) bool { return points[i].JDE < points[j].JDE })
	if options.TargetSpacingKM > 0 {
		points = solver.refineCentralPathSpacing(points, options.TargetSpacingKM)
	}
	points = normalizeSolarEclipsePathPointSeries(points)
	if len(points) < 2 {
		return nil, stepDays
	}
	return points, stepDays
}

func (solver solarEclipseSolver) partialFootprints(
	startJDE, endJDE, greatestJDE float64,
	options SolarEclipsePartialFootprintOptions,
) ([]SolarEclipsePartialFootprint, float64, int) {
	footprints, stepDays, boundaryPoints := solver.shadowFootprintsWithSpacing(
		startJDE,
		endJDE,
		greatestJDE,
		options.StepDays,
		options.BoundaryPoints,
		solarEclipsePenumbralShadow,
		solarEclipsePartialFootprintTargetSpacingKM,
	)
	return footprints, stepDays, boundaryPoints
}

func (solver solarEclipseSolver) shadowFootprints(
	startJDE, endJDE, greatestJDE, requestedStepDays float64,
	boundaryPoints int,
	kind solarEclipseShadowKind,
) ([]SolarEclipsePartialFootprint, float64, int) {
	return solver.shadowFootprintsWithSpacing(
		startJDE, endJDE, greatestJDE, requestedStepDays,
		boundaryPoints, kind, 0,
	)
}

func (solver solarEclipseSolver) shadowFootprintsWithSpacing(
	startJDE, endJDE, greatestJDE, requestedStepDays float64,
	boundaryPoints int,
	kind solarEclipseShadowKind,
	targetSpacingKM float64,
) ([]SolarEclipsePartialFootprint, float64, int) {
	if endJDE < startJDE {
		startJDE, endJDE = endJDE, startJDE
	}
	if startJDE == 0 || endJDE == 0 || endJDE <= startJDE {
		return nil, requestedStepDays, boundaryPoints
	}

	times, stepDays := solarEclipseMovingDiskEngine().sampleTimes(
		startJDE, endJDE, greatestJDE, requestedStepDays,
	)

	// Very small, near-grazing footprints can occupy less than one angular
	// sample at a caller-requested low boundary resolution.  Retry the whole
	// sequence at a bounded finer resolution only when the first pass found no
	// usable footprint; ordinary events keep the requested cost unchanged.
	maximumBoundaryPoints := solarEclipseMaximumBoundaryPoints(len(times))
	retryBoundaryPoints := maximumBoundaryPoints
	if retryBoundaryPoints > 360 {
		retryBoundaryPoints = 360
	}
	effectiveBoundaryPoints := solarEclipseEffectiveBoundaryPoints(len(times), boundaryPoints)
	// With very dense temporal sampling, the fixed spatial refinement target
	// can add many more vertices than the requested angular resolution. The
	// shared aggregate budget above already limits the two main sweeps; avoid
	// defeating that limit by repeating spatial refinement at every second.
	if targetSpacingKM > 0 && len(times)*effectiveBoundaryPoints > solarEclipsePartialFootprintMaxPointCount/3 {
		targetSpacingKM = 0
	}
	var footprints []SolarEclipsePartialFootprint
	for {
		footprints = make([]SolarEclipsePartialFootprint, 0, len(times))
		for _, jd := range times {
			footprint := solver.shadowFootprintAtWithSpacing(
				jd, effectiveBoundaryPoints, kind, targetSpacingKM,
			)
			if len(footprint.Boundaries) > 0 {
				footprints = append(footprints, footprint)
			}
		}
		if len(footprints) > 0 || effectiveBoundaryPoints >= retryBoundaryPoints {
			break
		}
		effectiveBoundaryPoints *= 4
		if effectiveBoundaryPoints < 96 {
			effectiveBoundaryPoints = 96
		}
		if effectiveBoundaryPoints > retryBoundaryPoints {
			effectiveBoundaryPoints = retryBoundaryPoints
		}
	}
	return footprints, stepDays, effectiveBoundaryPoints
}

func solarEclipseEffectiveBoundaryPoints(sampleCount, requested int) int {
	if requested < solarEclipsePartialFootprintMinBoundaryPoints {
		requested = solarEclipsePartialFootprintMinBoundaryPoints
	}
	if sampleCount < 1 {
		return requested
	}
	maximum := solarEclipseMaximumBoundaryPoints(sampleCount)
	if requested > maximum {
		return maximum
	}
	return requested
}

func solarEclipseMaximumBoundaryPoints(sampleCount int) int {
	if sampleCount < 1 {
		return solarEclipsePartialFootprintMaxBoundaryPoints
	}
	maximum := solarEclipsePartialFootprintMaxPointCount / sampleCount
	if maximum < solarEclipsePartialFootprintMinBoundaryPoints {
		maximum = solarEclipsePartialFootprintMinBoundaryPoints
	}
	if maximum > solarEclipsePartialFootprintMaxBoundaryPoints {
		maximum = solarEclipsePartialFootprintMaxBoundaryPoints
	}
	return maximum
}

func solarEclipseSharedBoundaryPoints(requested, partialSamples, shadowSamples, reserved int) int {
	if requested < solarEclipsePartialFootprintMinBoundaryPoints {
		requested = solarEclipsePartialFootprintMinBoundaryPoints
	}
	if partialSamples < 0 {
		partialSamples = 0
	}
	if shadowSamples < 0 {
		shadowSamples = 0
	}
	if reserved < 0 {
		reserved = 0
	}
	remaining := solarEclipsePartialFootprintMaxPointCount - reserved
	if remaining <= 0 || partialSamples+shadowSamples <= 0 {
		return requested
	}
	shared := remaining / (partialSamples + shadowSamples)
	if shared < solarEclipsePartialFootprintMinBoundaryPoints {
		shared = solarEclipsePartialFootprintMinBoundaryPoints
	}
	if shared < requested {
		return shared
	}
	return requested
}

func solarEclipseFootprintPointCount(footprints []SolarEclipsePartialFootprint) int {
	total := 0
	for _, footprint := range footprints {
		for _, boundary := range footprint.Boundaries {
			if len(boundary) > solarEclipsePartialFootprintMaxPointCount-total {
				return solarEclipsePartialFootprintMaxPointCount
			}
			total += len(boundary)
		}
	}
	return total
}

func (solver solarEclipseSolver) magnitudeContours(
	startJDE, endJDE, centralStartJDE, centralEndJDE, greatestJDE float64,
	options SolarEclipsePartialFootprintOptions,
	maximumMagnitude float64,
	hybrid bool,
	precomputedMagnitudeOne ...[][]SolarEclipsePathPoint,
) []SolarEclipseMagnitudeContour {
	if len(options.MagnitudeValues) == 0 || startJDE == 0 || endJDE == 0 || endJDE <= startJDE {
		return nil
	}
	contours := make([]SolarEclipseMagnitudeContour, 0, len(options.MagnitudeValues))
	for _, magnitude := range options.MagnitudeValues {
		if magnitude > maximumMagnitude+1e-9 {
			continue
		}
		var segments [][]SolarEclipsePathPoint
		if math.Abs(magnitude-1) <= 1e-12 && len(precomputedMagnitudeOne) > 0 &&
			len(precomputedMagnitudeOne[0]) > 0 {
			segments = precomputedMagnitudeOne[0]
		} else {
			segments = solver.magnitudeContourSegments(
				startJDE, endJDE, centralStartJDE, centralEndJDE,
				greatestJDE, magnitude, options.StepDays, hybrid,
			)
		}
		if len(segments) == 0 {
			continue
		}
		northern, southern := solarEclipseMagnitudeContourCompatibilitySides(segments)
		contours = append(contours, SolarEclipseMagnitudeContour{
			Magnitude:     magnitude,
			Segments:      segments,
			NorthernLimit: northern,
			SouthernLimit: southern,
		})
	}
	return contours
}
