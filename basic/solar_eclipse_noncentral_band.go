package basic

import (
	"math"

	"github.com/starainrt/astro/internal/geodata"
)

const (
	solarEclipseNonCentralBandTimeScale              = 360.0
	solarEclipseNonCentralBandContainmentToleranceKM = solarEclipseCentralBandTargetSpacingKM / 2
	solarEclipseNonCentralBandDerivativeTolerance    = 5e-7
)

type solarEclipseNonCentralBandState struct {
	coordinates [3]float64
	tangent     [3]float64
	point       SolarEclipsePathPoint
}

func solarEclipseNonCentralBandContainsFootprints(
	segments [][]SolarEclipsePathPoint,
	footprints []SolarEclipsePartialFootprint,
) bool {
	return solarEclipseBandContainsFootprintsWithinKM(
		segments, footprints, solarEclipseNonCentralBandContainmentToleranceKM,
	)
}

// solarEclipseBandContainsFootprintsWithinKM is the tolerance-aware form used
// by the sampled central-band reconstruction, whose decimated rings are not an
// analytic envelope and may cut inside the sharpest grazing tips.
func solarEclipseBandContainsFootprintsWithinKM(
	segments [][]SolarEclipsePathPoint,
	footprints []SolarEclipsePartialFootprint,
	toleranceKM float64,
) bool {
	polygons, paths := solarEclipseBandFootprintGeometry(segments, footprints)
	if len(polygons) == 0 || len(paths) == 0 {
		return false
	}
	return geodata.SphericalPolygonsContainPathsWithinKM(
		polygons, paths, false, toleranceKM,
	)
}

// The sweep audit samples the footprint series instead of probing every vertex
// of every instantaneous footprint: the reconstruction residual is a systematic
// gap, not one stray vertex, and a full probe costs an order of magnitude more
// than building the band (1136-06-01: 109k probes, 1.3 s in Go).
const (
	solarEclipseCentralBandSweepProbePaths  = 64
	solarEclipseCentralBandSweepProbePoints = 32
)

// solarEclipseBandContainsSampledFootprintsWithinKM validates a candidate
// against a bounded, evenly spread probe set of the reference footprints.
func solarEclipseBandContainsSampledFootprintsWithinKM(
	segments [][]SolarEclipsePathPoint,
	footprints []SolarEclipsePartialFootprint,
	toleranceKM float64,
) bool {
	polygons, _ := solarEclipseBandFootprintGeometry(segments, footprints)
	if len(polygons) == 0 || len(footprints) == 0 {
		return false
	}
	paths := make([][]geodata.GeoPoint, 0, solarEclipseCentralBandSweepProbePaths)
	for _, index := range solarEclipseBandProbeIndices(len(footprints), solarEclipseCentralBandSweepProbePaths) {
		for _, boundary := range footprints[index].Boundaries {
			if len(boundary) < 3 {
				continue
			}
			path := make([]geodata.GeoPoint, 0, solarEclipseCentralBandSweepProbePoints)
			for _, pointIndex := range solarEclipseBandProbeIndices(len(boundary), solarEclipseCentralBandSweepProbePoints) {
				point := boundary[pointIndex]
				path = append(path, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
			}
			if len(path) >= 3 {
				paths = append(paths, path)
			}
		}
	}
	if len(paths) == 0 {
		return false
	}
	return geodata.SphericalPolygonsContainPathsWithinKM(polygons, paths, false, toleranceKM)
}

// solarEclipseBandProbeIndices returns at most limit evenly spread indices over
// count items, always including the first and the last one.
func solarEclipseBandProbeIndices(count, limit int) []int {
	if count <= 0 {
		return nil
	}
	if limit < 2 {
		limit = 2
	}
	if count <= limit {
		indices := make([]int, count)
		for index := range indices {
			indices[index] = index
		}
		return indices
	}
	indices := make([]int, 0, limit)
	for index := 0; index < limit; index++ {
		indices = append(indices, index*(count-1)/(limit-1))
	}
	return indices
}

// solarEclipseBandFootprintGeometry converts one band candidate and its
// reference footprints into the spherical form the audits share.
func solarEclipseBandFootprintGeometry(
	segments [][]SolarEclipsePathPoint,
	footprints []SolarEclipsePartialFootprint,
) ([][]geodata.GeoPoint, [][]geodata.GeoPoint) {
	if len(segments) == 0 || len(footprints) == 0 {
		return nil, nil
	}
	polygons := make([][]geodata.GeoPoint, 0, len(segments))
	for _, segment := range segments {
		polygon := make([]geodata.GeoPoint, len(segment))
		for index, point := range segment {
			polygon[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		polygons = append(polygons, polygon)
	}
	paths := make([][]geodata.GeoPoint, 0, len(footprints))
	for _, footprint := range footprints {
		for _, boundary := range footprint.Boundaries {
			path := make([]geodata.GeoPoint, len(boundary))
			for index, point := range boundary {
				path[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			paths = append(paths, path)
		}
	}
	return polygons, paths
}

// nonCentralBandRegion returns one closed annular region for a non-central
// eclipse. The sampled critical envelope supplies the visible outer arc; the
// horizon arc supplies the degenerate side where the band reaches sunset or
// sunrise.
func (solver solarEclipseSolver) nonCentralBandRegion(
	samples []solarEclipseCentralBandSweepSample,
	riseSetCurves []SolarEclipseRiseSetCurve,
	referenceJDE float64,
) ([]SolarEclipsePathPoint, []SolarEclipsePathPoint) {
	boundary := solver.correctNonCentralBandEnvelopeSamples(
		samples, referenceJDE,
	)
	return solver.closeNonCentralBandBoundary(boundary, riseSetCurves)
}

func (solver solarEclipseSolver) closeNonCentralBandBoundary(
	boundary []SolarEclipsePathPoint,
	riseSetCurves []SolarEclipseRiseSetCurve,
) ([]SolarEclipsePathPoint, []SolarEclipsePathPoint) {
	horizon := solver.nonCentralBandHorizonSegment(riseSetCurves)
	if len(horizon) < 3 {
		return nil, nil
	}
	if len(boundary) < 2 {
		return nil, nil
	}
	keep := solarEclipsePathDistanceKM(boundary[0], horizon[0]) +
		solarEclipsePathDistanceKM(boundary[len(boundary)-1], horizon[len(horizon)-1])
	reverse := solarEclipsePathDistanceKM(boundary[0], horizon[len(horizon)-1]) +
		solarEclipsePathDistanceKM(boundary[len(boundary)-1], horizon[0])
	if reverse < keep {
		for left, right := 0, len(horizon)-1; left < right; left, right = left+1, right-1 {
			horizon[left], horizon[right] = horizon[right], horizon[left]
		}
	}
	polygon := append([]SolarEclipsePathPoint(nil), boundary...)
	polygon = appendNonCentralBandInterpolatedSegment(polygon, boundary[len(boundary)-1], horizon[len(horizon)-1])
	for index := len(horizon) - 2; index >= 0; index-- {
		polygon = append(polygon, horizon[index])
	}
	polygon = appendNonCentralBandInterpolatedSegment(polygon, horizon[0], boundary[0])
	polygon = deduplicateSolarEclipsePathPoints(polygon)
	if len(polygon) < 4 || solarEclipsePathDistanceKM(polygon[0], polygon[len(polygon)-1]) > 0.01 {
		return nil, nil
	}
	polygon[len(polygon)-1] = polygon[0]
	return polygon, horizon
}

// alignNonCentralBandHorizon replaces the coarse public greatest-at-horizon
// samples with the exact horizon side used to close the non-central band. The
// shared vertices keep GeoJSON and SVG renderers from drawing a chord through
// the narrow band between otherwise identical roots.
func alignNonCentralBandHorizon(
	curves []SolarEclipseRiseSetCurve,
	horizon []SolarEclipsePathPoint,
) {
	if len(horizon) < 2 {
		return
	}
	start, end := horizon[0], horizon[len(horizon)-1]
	for curveIndex := range curves {
		curve := &curves[curveIndex]
		if curve.Phase != RiseSetPhaseGreatest {
			continue
		}
		for segmentIndex, segment := range curve.Segments {
			if len(segment) < 2 || segment[0].JDE >= end.JDE || segment[len(segment)-1].JDE <= start.JDE {
				continue
			}
			if !nonCentralBandHorizonMatchesSegment(segment, start, end) {
				continue
			}
			joined := make([]SolarEclipsePathPoint, 0, len(segment)+len(horizon))
			for _, point := range segment {
				if point.JDE < start.JDE-solarEclipseRiseSetTimeEpsilonDays {
					joined = append(joined, point)
				}
			}
			joined = append(joined, horizon...)
			for _, point := range segment {
				if point.JDE > end.JDE+solarEclipseRiseSetTimeEpsilonDays {
					joined = append(joined, point)
				}
			}
			curve.Segments[segmentIndex] = joined
			// The exact horizon arc may fold in time at high latitude. It is
			// attached after the normal rise/set topology pass, so normalize
			// here as well to split that newly introduced fold into branches.
			normalizeSolarEclipseRiseSetCurveSegments(curve)
			return
		}
	}
}

func nonCentralBandHorizonMatchesSegment(
	segment []SolarEclipsePathPoint,
	start, end SolarEclipsePathPoint,
) bool {
	const maximumAttachmentDistanceKM = 100.0
	nearestDistance := func(target SolarEclipsePathPoint) float64 {
		best := math.Inf(1)
		for _, point := range segment {
			if math.Abs(point.JDE-target.JDE) > 10.0/1440.0 {
				continue
			}
			best = math.Min(best, solarEclipsePathDistanceKM(point, target))
		}
		return best
	}
	return nearestDistance(start) <= maximumAttachmentDistanceKM &&
		nearestDistance(end) <= maximumAttachmentDistanceKM
}

func (solver solarEclipseSolver) correctNonCentralBandEnvelopeSamples(
	samples []solarEclipseCentralBandSweepSample,
	referenceJDE float64,
) []SolarEclipsePathPoint {
	points := make([]SolarEclipsePathPoint, 0, len(samples))
	for _, sample := range samples {
		state, stateOK := solver.nonCentralBandStateAt(sample.envelope, referenceJDE)
		if !stateOK {
			continue
		}
		corrected, _, correctedOK := solver.correctNonCentralBandBoundary(
			state.coordinates, state.tangent, referenceJDE,
		)
		// A near-grazing Newton solve can converge to another critical branch.
		// Keep the correction only when it remains close to this sample;
		// otherwise validate and use the local predictor below.
		if correctedOK && solarEclipsePathDistanceKM(corrected.point, sample.envelope) <=
			2*solarEclipseCentralBandTargetSpacingKM {
			state = corrected
		} else {
			evaluation := solver.magnitudeEvaluationAt(sample.envelope.JDE)
			approximate := evaluation.center.stateAt(sample.envelope.Longitude*rad, sample.envelope.Latitude*rad, 0)
			if math.Abs(solarEclipseCentralContactGap(approximate)) > 1e-6 ||
				evaluation.centralContactSecondDerivative(sample.envelope.Longitude, sample.envelope.Latitude) <= 0 {
				continue
			}
			state.point = SolarEclipsePathPoint{
				JDE: sample.envelope.JDE, Longitude: sample.envelope.Longitude,
				Latitude: sample.envelope.Latitude, SunAltitude: approximate.sunAltitudeRad / rad,
			}
		}
		if len(points) > 0 && solarEclipsePathDistanceKM(points[len(points)-1], state.point) >
			2*solarEclipseCentralBandTargetSpacingKM {
			// Prefer the continuous predictor if a corrected branch jumped.
			evaluation := solver.magnitudeEvaluationAt(sample.envelope.JDE)
			approximate := evaluation.center.stateAt(sample.envelope.Longitude*rad, sample.envelope.Latitude*rad, 0)
			if math.Abs(solarEclipseCentralContactGap(approximate)) > 1e-6 ||
				evaluation.centralContactSecondDerivative(sample.envelope.Longitude, sample.envelope.Latitude) <= 0 ||
				solarEclipsePathDistanceKM(points[len(points)-1], sample.envelope) >
					2*solarEclipseCentralBandTargetSpacingKM {
				return nil
			}
			state.point = sample.envelope
		}
		points = append(points, state.point)
	}
	return deduplicateSolarEclipsePathPoints(points)
}

func appendNonCentralBandInterpolatedSegment(
	points []SolarEclipsePathPoint,
	start, end SolarEclipsePathPoint,
) []SolarEclipsePathPoint {
	distance := solarEclipsePathDistanceKM(start, end)
	steps := int(math.Ceil(distance / solarEclipseCentralBandTargetSpacingKM))
	if steps < 1 {
		steps = 1
	}
	deltaLongitude := math.Remainder(end.Longitude-start.Longitude, 360)
	for step := 1; step <= steps; step++ {
		fraction := float64(step) / float64(steps)
		points = append(points, SolarEclipsePathPoint{
			JDE:         start.JDE + fraction*(end.JDE-start.JDE),
			Longitude:   normalizeLongitude(start.Longitude + fraction*deltaLongitude),
			Latitude:    start.Latitude + fraction*(end.Latitude-start.Latitude),
			SunAltitude: start.SunAltitude + fraction*(end.SunAltitude-start.SunAltitude),
		})
	}
	return points
}

func (solver solarEclipseSolver) nonCentralBandHorizonSegment(
	curves []SolarEclipseRiseSetCurve,
) []SolarEclipsePathPoint {
	var best []SolarEclipsePathPoint
	for _, curve := range curves {
		if curve.Phase != RiseSetPhaseGreatest {
			continue
		}
		for _, segment := range curve.Segments {
			segment = solver.correctNonCentralBandHorizonSamples(segment)
			if len(segment) < 2 {
				continue
			}
			var current []SolarEclipsePathPoint
			for index := 1; index < len(segment); index++ {
				first, second := segment[index-1], segment[index]
				firstGap, firstOK := solver.nonCentralBandGapAt(first)
				secondGap, secondOK := solver.nonCentralBandGapAt(second)
				if !firstOK || !secondOK {
					current = nil
					continue
				}
				if len(current) == 0 {
					switch {
					case firstGap > 0 && secondGap <= 0:
						junction, ok := solver.refineNonCentralBandHorizonGapCrossing(first, second)
						if !ok {
							continue
						}
						current = append(current, junction)
					case firstGap <= 0:
						current = append(current, first)
					default:
						continue
					}
				}
				if secondGap <= 0 {
					if solarEclipsePathDistanceKM(current[len(current)-1], second) > 0.001 {
						current = append(current, second)
					}
					continue
				}
				if firstGap <= 0 {
					junction, ok := solver.refineNonCentralBandHorizonGapCrossing(first, second)
					if ok {
						current = append(current, junction)
					}
				}
				if solver.nonCentralBandHorizonCandidateValid(current) && len(current) > len(best) {
					best = append([]SolarEclipsePathPoint(nil), current...)
				}
				current = nil
			}
			if solver.nonCentralBandHorizonCandidateValid(current) && len(current) > len(best) {
				best = append([]SolarEclipsePathPoint(nil), current...)
			}
		}
	}
	if len(best) < 2 {
		return nil
	}
	refined := make([]SolarEclipsePathPoint, 1, len(best))
	refined[0] = best[0]
	for index := 1; index < len(best); index++ {
		refined = solver.appendRefinedSolarEclipseCentralHorizonSegment(
			refined, best[index-1], best[index], 0,
		)
	}
	return deduplicateSolarEclipsePathPoints(refined)
}

func (solver solarEclipseSolver) correctNonCentralBandHorizonSamples(
	segment []SolarEclipsePathPoint,
) []SolarEclipsePathPoint {
	corrected := make([]SolarEclipsePathPoint, 0, len(segment))
	for _, point := range segment {
		evaluation := solver.magnitudeEvaluationAt(point.JDE)
		longitude, latitude, ok := riseSetRefineGeographicRoot(
			point.Longitude,
			point.Latitude,
			func(lon, lat float64) (float64, float64, bool) {
				state := evaluation.center.stateAt(lon*rad, lat*rad, 0)
				phase := evaluation.separationDerivative(lon, lat)
				return phase, state.sunAltitudeRad, finite(phase) && finite(state.sunAltitudeRad)
			},
		)
		if !ok || evaluation.separationSecondDerivative(longitude, latitude) <= 0 {
			continue
		}
		state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
		corrected = append(corrected, SolarEclipsePathPoint{
			JDE: point.JDE, Longitude: longitude, Latitude: latitude, SunAltitude: state.sunAltitudeRad / rad,
		})
	}
	return corrected
}

func (solver solarEclipseSolver) nonCentralBandHorizonCandidateValid(
	points []SolarEclipsePathPoint,
) bool {
	if len(points) < 2 {
		return false
	}
	firstGap, firstOK := solver.nonCentralBandGapAt(points[0])
	lastGap, lastOK := solver.nonCentralBandGapAt(points[len(points)-1])
	return firstOK && lastOK && math.Abs(firstGap) <= 1e-7 && math.Abs(lastGap) <= 1e-7
}

func (solver solarEclipseSolver) appendRefinedSolarEclipseCentralHorizonSegment(
	points []SolarEclipsePathPoint,
	start, end SolarEclipsePathPoint,
	depth int,
) []SolarEclipsePathPoint {
	if solarEclipsePathDistanceKM(start, end) <= solarEclipseCentralBandTargetSpacingKM ||
		depth >= 16 || end.JDE-start.JDE <= solarEclipsePathMinStepDays {
		return append(points, end)
	}
	jd := (start.JDE + end.JDE) / 2
	longitude := normalizeLongitude(
		start.Longitude + math.Remainder(end.Longitude-start.Longitude, 360)/2,
	)
	latitude := (start.Latitude + end.Latitude) / 2
	evaluation := solver.magnitudeEvaluationAt(jd)
	longitude, latitude, ok := riseSetRefineGeographicRoot(
		longitude,
		latitude,
		func(lon, lat float64) (float64, float64, bool) {
			state := evaluation.center.stateAt(lon*rad, lat*rad, 0)
			phase := evaluation.separationDerivative(lon, lat)
			return phase, state.sunAltitudeRad, finite(phase) && finite(state.sunAltitudeRad)
		},
	)
	if !ok {
		return append(points, end)
	}
	state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
	if solarEclipseCentralContactGap(state) > 1e-7 ||
		evaluation.separationSecondDerivative(longitude, latitude) <= 0 {
		return append(points, end)
	}
	middle := SolarEclipsePathPoint{
		JDE: jd, Longitude: longitude, Latitude: latitude, SunAltitude: state.sunAltitudeRad / rad,
	}
	points = solver.appendRefinedSolarEclipseCentralHorizonSegment(points, start, middle, depth+1)
	return solver.appendRefinedSolarEclipseCentralHorizonSegment(points, middle, end, depth+1)
}

func (solver solarEclipseSolver) nonCentralBandStateAt(
	point SolarEclipsePathPoint,
	referenceJDE float64,
) (solarEclipseNonCentralBandState, bool) {
	coordinates := [3]float64{
		point.Longitude,
		point.Latitude,
		(point.JDE - referenceJDE) * solarEclipseNonCentralBandTimeScale,
	}
	_, jacobian, ok := solver.nonCentralBandBoundaryJacobian(coordinates, referenceJDE)
	if !ok {
		return solarEclipseNonCentralBandState{}, false
	}
	tangent, ok := solarEclipseMagnitudeArcTangent(jacobian)
	return solarEclipseNonCentralBandState{coordinates: coordinates, tangent: tangent, point: point}, ok
}

func (solver solarEclipseSolver) correctNonCentralBandBoundary(
	predictor, tangent [3]float64,
	referenceJDE float64,
) (solarEclipseNonCentralBandState, int, bool) {
	coordinates := predictor
	for iteration := 0; iteration < 16; iteration++ {
		residual, jacobian, ok := solver.nonCentralBandBoundaryJacobian(coordinates, referenceJDE)
		if !ok {
			return solarEclipseNonCentralBandState{}, iteration, false
		}
		planeResidual := dotSolarEclipse3(subtractSolarEclipse3(coordinates, predictor), tangent)
		if math.Abs(residual[0]) <= 1e-10 &&
			math.Abs(residual[1]) <= solarEclipseNonCentralBandDerivativeTolerance &&
			math.Abs(planeResidual) <= 1e-9 {
			return solver.validNonCentralBandState(coordinates, jacobian, referenceJDE, iteration+1)
		}
		matrix := [3][3]float64{jacobian[0], jacobian[1], tangent}
		delta, ok := solveSolarEclipse3x3(matrix, [3]float64{-residual[0], -residual[1], -planeResidual})
		if !ok {
			return solarEclipseNonCentralBandState{}, iteration, false
		}
		norm := math.Sqrt(dotSolarEclipse3(delta, delta))
		if norm > 2 {
			for index := range delta {
				delta[index] *= 2 / norm
			}
		}
		for index := range coordinates {
			coordinates[index] += delta[index]
		}
		coordinates[0], coordinates[1] = normalizeSolarEclipseSphericalCoordinates(coordinates[0], coordinates[1])
	}
	residual, jacobian, ok := solver.nonCentralBandBoundaryJacobian(coordinates, referenceJDE)
	planeResidual := dotSolarEclipse3(subtractSolarEclipse3(coordinates, predictor), tangent)
	if !ok || math.Abs(residual[0]) > 1e-7 ||
		math.Abs(residual[1]) > solarEclipseNonCentralBandDerivativeTolerance ||
		math.Abs(planeResidual) > 1e-7 {
		return solarEclipseNonCentralBandState{}, 16, false
	}
	return solver.validNonCentralBandState(coordinates, jacobian, referenceJDE, 16)
}

// normalizeSolarEclipseSphericalCoordinates keeps continuation coordinates on
// the sphere when a polar branch crosses a geographic pole. Reflecting the
// latitude and shifting longitude by 180 degrees preserves the same point and
// avoids the artificial singularity at +/-90 degrees.
func normalizeSolarEclipseSphericalCoordinates(longitude, latitude float64) (float64, float64) {
	crossedPole := false
	for latitude > 90 || latitude < -90 {
		crossedPole = true
		if latitude > 90 {
			latitude = 180 - latitude
			longitude += 180
			continue
		}
		latitude = -180 - latitude
		longitude += 180
	}
	if crossedPole {
		longitude = normalizeLongitude(longitude)
	}
	return longitude, latitude
}

func (solver solarEclipseSolver) validNonCentralBandState(
	coordinates [3]float64,
	jacobian [2][3]float64,
	referenceJDE float64,
	iterations int,
) (solarEclipseNonCentralBandState, int, bool) {
	jd := referenceJDE + coordinates[2]/solarEclipseNonCentralBandTimeScale
	longitude, latitude := normalizeLongitude(coordinates[0]), coordinates[1]
	evaluation := solver.magnitudeEvaluationAt(jd)
	state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
	if math.Abs(solarEclipseCentralContactGap(state)) > 1e-6 ||
		math.Abs(evaluation.centralContactDerivative(longitude, latitude)) >
			solarEclipseNonCentralBandDerivativeTolerance ||
		evaluation.centralContactSecondDerivative(longitude, latitude) <= 0 {
		return solarEclipseNonCentralBandState{}, iterations, false
	}
	tangent, ok := solarEclipseMagnitudeArcTangent(jacobian)
	if !ok {
		return solarEclipseNonCentralBandState{}, iterations, false
	}
	return solarEclipseNonCentralBandState{
		coordinates: coordinates,
		tangent:     tangent,
		point: SolarEclipsePathPoint{
			JDE: jd, Longitude: longitude, Latitude: latitude, SunAltitude: state.sunAltitudeRad / rad,
		},
	}, iterations, true
}

func (solver solarEclipseSolver) nonCentralBandBoundaryJacobian(
	coordinates [3]float64,
	referenceJDE float64,
) ([2]float64, [2][3]float64, bool) {
	jd := referenceJDE + coordinates[2]/solarEclipseNonCentralBandTimeScale
	longitude, latitude := normalizeLongitude(coordinates[0]), coordinates[1]
	evaluation := solver.magnitudeEvaluationAt(jd)
	residual, ok := solarEclipseNonCentralBandBoundaryResidualAt(evaluation, longitude, latitude)
	if !ok {
		return [2]float64{}, [2][3]float64{}, false
	}
	steps := [3]float64{1e-4, 1e-4, 5.0 * solarEclipseNonCentralBandTimeScale / 86400.0}
	jacobian := [2][3]float64{}
	for column, shifted := range [][2]float64{{longitude + steps[0], latitude}, {longitude, latitude + steps[1]}} {
		shiftedResidual, shiftedOK := solarEclipseNonCentralBandBoundaryResidualAt(
			evaluation, shifted[0], shifted[1],
		)
		if !shiftedOK {
			return [2]float64{}, [2][3]float64{}, false
		}
		for row := 0; row < 2; row++ {
			jacobian[row][column] = (shiftedResidual[row] - residual[row]) / steps[column]
		}
	}
	timeEvaluation := solver.magnitudeEvaluationAt(jd + steps[2]/solarEclipseNonCentralBandTimeScale)
	timeResidual, timeOK := solarEclipseNonCentralBandBoundaryResidualAt(
		timeEvaluation, longitude, latitude,
	)
	if !timeOK {
		return [2]float64{}, [2][3]float64{}, false
	}
	for row := 0; row < 2; row++ {
		jacobian[row][2] = (timeResidual[row] - residual[row]) / steps[2]
	}
	return residual, jacobian, true
}

func solarEclipseNonCentralBandBoundaryResidualAt(
	evaluation solarEclipseRiseSetEvaluation,
	longitude, latitude float64,
) ([2]float64, bool) {
	state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
	residual := [2]float64{
		solarEclipseCentralContactGap(state),
		evaluation.centralContactDerivative(longitude, latitude),
	}
	return residual, finite(residual[0]) && finite(residual[1])
}

func solarEclipseCentralContactGap(state localSolarEclipseState) float64 {
	return state.movingDiskContactState().internalContactGap()
}

func (evaluation solarEclipseRiseSetEvaluation) centralContactDerivative(longitude, latitude float64) float64 {
	before := evaluation.before.stateAt(longitude*rad, latitude*rad, 0)
	after := evaluation.after.stateAt(longitude*rad, latitude*rad, 0)
	return (solarEclipseCentralContactGap(after) - solarEclipseCentralContactGap(before)) /
		(2 * solarEclipseRiseSetDerivativeStepDays)
}

func (evaluation solarEclipseRiseSetEvaluation) centralContactSecondDerivative(longitude, latitude float64) float64 {
	before := evaluation.before.stateAt(longitude*rad, latitude*rad, 0)
	center := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
	after := evaluation.after.stateAt(longitude*rad, latitude*rad, 0)
	stepSquared := solarEclipseRiseSetDerivativeStepDays * solarEclipseRiseSetDerivativeStepDays
	return (solarEclipseCentralContactGap(after) - 2*solarEclipseCentralContactGap(center) +
		solarEclipseCentralContactGap(before)) / stepSquared
}

func (solver solarEclipseSolver) refineNonCentralBandHorizonGapCrossing(
	first, second SolarEclipsePathPoint,
) (SolarEclipsePathPoint, bool) {
	firstGap, firstOK := solver.nonCentralBandGapAt(first)
	secondGap, secondOK := solver.nonCentralBandGapAt(second)
	if !firstOK || !secondOK || firstGap*secondGap > 0 {
		return SolarEclipsePathPoint{}, false
	}
	if first.JDE > second.JDE {
		first, second = second, first
		firstGap, secondGap = secondGap, firstGap
	}
	best := first
	bestGap := math.Abs(firstGap)
	if math.Abs(secondGap) < bestGap {
		best, bestGap = second, math.Abs(secondGap)
	}
	for iteration := 0; iteration < 64; iteration++ {
		jd := (first.JDE + second.JDE) / 2
		fraction := (jd - first.JDE) / (second.JDE - first.JDE)
		longitude := normalizeLongitude(
			first.Longitude + fraction*math.Remainder(second.Longitude-first.Longitude, 360),
		)
		latitude := first.Latitude + fraction*(second.Latitude-first.Latitude)
		evaluation := solver.magnitudeEvaluationAt(jd)
		longitude, latitude, ok := riseSetRefineGeographicRoot(
			longitude,
			latitude,
			func(lon, lat float64) (float64, float64, bool) {
				state := evaluation.center.stateAt(lon*rad, lat*rad, 0)
				phase := evaluation.separationDerivative(lon, lat)
				return phase, state.sunAltitudeRad, finite(phase) && finite(state.sunAltitudeRad)
			},
		)
		if !ok {
			return SolarEclipsePathPoint{}, false
		}
		state := evaluation.center.stateAt(longitude*rad, latitude*rad, 0)
		middleGap := solarEclipseCentralContactGap(state)
		middle := SolarEclipsePathPoint{
			JDE: jd, Longitude: longitude, Latitude: latitude, SunAltitude: state.sunAltitudeRad / rad,
		}
		if math.Abs(middleGap) < bestGap {
			best, bestGap = middle, math.Abs(middleGap)
		}
		if bestGap <= 1e-10 || second.JDE-first.JDE <= 1e-10 {
			break
		}
		if firstGap*middleGap <= 0 {
			second, secondGap = middle, middleGap
		} else {
			first, firstGap = middle, middleGap
		}
	}
	if bestGap > 1e-7 {
		return SolarEclipsePathPoint{}, false
	}
	evaluation := solver.magnitudeEvaluationAt(best.JDE)
	if math.Abs(evaluation.separationDerivative(best.Longitude, best.Latitude)) > 1e-8 ||
		math.Abs(best.SunAltitude) > 1e-5 ||
		evaluation.separationSecondDerivative(best.Longitude, best.Latitude) <= 0 {
		return SolarEclipsePathPoint{}, false
	}
	return best, true
}

func (solver solarEclipseSolver) nonCentralBandGapAt(
	point SolarEclipsePathPoint,
) (float64, bool) {
	evaluation := solver.magnitudeEvaluationAt(point.JDE)
	state := evaluation.center.stateAt(point.Longitude*rad, point.Latitude*rad, 0)
	gap := solarEclipseCentralContactGap(state)
	return gap, finite(gap)
}
