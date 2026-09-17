package basic

import (
	"math"
	"time"
)

// occultationStationBoundarySample is the result of correcting one geocentric
// contact seed with the station-centred contact equation.  Residuals follow
// occultationRiseSetContext.stateAt, which returns degrees for both the contact
// gap and the lunar altitude; the two residual fields are therefore degrees,
// not arcseconds.
type occultationStationBoundarySample struct {
	point              OccultationPathPoint
	contactResidualDeg float64
	horizonResidualDeg float64
	offsetKM           float64
	seedResidualDeg    float64
	valid              bool
}

const (
	occultationStationOracleInitialStepKM   = 25.0
	occultationStationOracleMaximumOffsetKM = 2000.0
	occultationStationOracleRootToleranceKM = 0.001
	// 站心接触残差与 stateAt 同单位（度）；1e-7 度 = 3.6e-4 角秒。
	// The station contact residual uses the stateAt unit (degrees); 1e-7 deg = 3.6e-4 arcsec.
	occultationStationOracleResidualToleranceDeg = 1e-7
	// 地面偏移每公里最多改变约 1/384400 弧度的月球视差方向，取 3e-4 度/公里作为
	// 接触残差的斜率上限，用来复核“括号已塌缩但残差没到容差”的解。
	occultationStationOracleResidualSlopeDegPerKM = 3e-4
	occultationStationEnvelopeMaximumOffsetKM     = 2500.0
	occultationStationHorizonMaximumOffsetKM      = 250.0
	occultationStationHorizonResidualToleranceDeg = 1e-7
	// 地平线接触求解的最终验收带；带内的负残差是数值噪声。
	occultationStationHorizonAcceptanceDeg = 1e-5
)

// occultationStationCorrectBoundaryPoint refines a geocentric contact point
// along the local ground cross-track direction.  The contact equation is
// evaluated by the existing station-centred vector context, so moon parallax,
// target parallax, apparent radii and the horizon all share one observer model.
// The function deliberately does not choose a polygon or join branches.
func occultationStationCorrectBoundaryPoint(
	tt float64,
	seed OccultationPathPoint,
	frameAt occultationPathFrameFunc,
	contextAt occultationRiseSetContextFunc,
	total bool,
	location *time.Location,
) (occultationStationBoundarySample, bool) {
	frame, ok := frameAt(tt)
	if !ok {
		return occultationStationBoundarySample{}, false
	}
	if contextAt == nil {
		return occultationStationBoundarySample{}, false
	}
	contextFactory := contextAt
	if total {
		contextFactory = func(value float64) occultationRiseSetContext {
			return contextAt(value).withInternalContact()
		}
	}
	context := contextFactory(tt)
	if !context.valid {
		return occultationStationBoundarySample{}, false
	}
	seedFixed := occultationStationSurfaceVector(seed.Longitude, seed.Latitude)
	if occultationPathNorm(seedFixed) <= 0 {
		return occultationStationBoundarySample{}, false
	}
	cross, directionOK := occultationStationCrossTrackAt(tt, frame, frameAt, seedFixed)
	if !directionOK {
		return occultationStationBoundarySample{}, false
	}

	evaluate := func(offsetKM float64) (float64, float64, float64, bool) {
		fixed := occultationStationOffsetSurfaceVector(seedFixed, cross, offsetKM)
		longitude, latitude := occultationStationGeodetic(fixed)
		state := context.stateAt(longitude, latitude)
		if !state.valid || !finite(state.contactMetric) || !finite(state.moonAltitude) {
			return 0, 0, 0, false
		}
		return state.contactMetric, state.moonAltitude, longitude, true
	}

	seedResidual, seedAltitude, _, seedOK := evaluate(0)
	if !seedOK {
		return occultationStationBoundarySample{}, false
	}
	if math.Abs(seedResidual) <= occultationStationOracleResidualToleranceDeg {
		return occultationStationOracleSampleAt(
			tt, seedFixed, 0, seedResidual, seedResidual, seedAltitude, frame, location,
		), true
	}

	leftOffset, rightOffset, bracketOK := occultationStationFindBracket(seedResidual, evaluate)
	if !bracketOK {
		return occultationStationBoundarySample{}, false
	}
	leftResidual, _, _, leftOK := evaluate(leftOffset)
	rightResidual, _, _, rightOK := evaluate(rightOffset)
	if !leftOK || !rightOK || leftResidual*rightResidual > 0 {
		return occultationStationBoundarySample{}, false
	}
	for iteration := 0; iteration < 64; iteration++ {
		middleOffset := (leftOffset + rightOffset) / 2
		middleResidual, _, _, middleOK := evaluate(middleOffset)
		if !middleOK {
			return occultationStationBoundarySample{}, false
		}
		if math.Abs(middleResidual) <= occultationStationOracleResidualToleranceDeg ||
			math.Abs(rightOffset-leftOffset) <= occultationStationOracleRootToleranceKM {
			leftOffset, rightOffset = middleOffset, middleOffset
			break
		}
		if leftResidual*middleResidual <= 0 {
			rightOffset, rightResidual = middleOffset, middleResidual
		} else {
			leftOffset, leftResidual = middleOffset, middleResidual
		}
	}
	offset := (leftOffset + rightOffset) / 2
	residual, altitude, _, solved := evaluate(offset)
	// 括号塌缩不等于解存在：偏离接触方程超过该宽度可解释范围的“收敛”点只是无效
	// 区间里的一个位置，不能当成已求解。
	if !solved || !finite(residual) ||
		math.Abs(residual) > occultationStationOracleResidualBound(rightOffset-leftOffset) {
		return occultationStationBoundarySample{}, false
	}
	pointFixed := occultationStationOffsetSurfaceVector(seedFixed, cross, offset)
	return occultationStationOracleSampleAt(
		tt, pointFixed, offset, seedResidual, residual, altitude, frame, location,
	), true
}

// occultationStationCorrectContours maps a geocentric contact-envelope seed to
// the station-centred temporal envelope without changing its sampling order.
// A static time union is bounded by contact=0 and d(contact)/dt=0; correcting
// contact alone moves a seed onto an instantaneous footprint but does not keep
// it on the outer envelope.
func occultationStationCorrectContours(
	contours [][]OccultationPathPoint,
	curves []OccultationRiseSetCurve,
	cache *occultationRiseSetEvaluationCache,
	location *time.Location,
) [][]OccultationPathPoint {
	if len(contours) == 0 || cache == nil {
		return contours
	}
	contextAt := cache.context
	corrected := make([][]OccultationPathPoint, 0, len(contours))
	for _, contour := range contours {
		if len(contour) == 0 {
			continue
		}
		samples, solved := occultationStationEnvelopeSamples(contour, contextAt, false, location)
		for start := 0; start < len(contour); {
			for start < len(contour) && !solved[start] {
				start++
			}
			if start == len(contour) {
				break
			}
			end := start
			for end < len(contour) && solved[end] {
				end++
			}
			if end-start >= 2 {
				segment := make([]OccultationPathPoint, end-start)
				for index := start; index < end; index++ {
					point := samples[index].point
					point.Time = contour[index].Time
					point.WidthKM = contour[index].WidthKM
					segment[index-start] = point
				}
				for _, sampleRange := range occultationContinuousBoundaryRanges(segment) {
					if sampleRange.end-sampleRange.start >= 2 {
						corrected = append(corrected, append(
							[]OccultationPathPoint(nil), segment[sampleRange.start:sampleRange.end]...,
						))
					}
				}
			}
			start = end
		}
	}
	if len(curves) == 0 {
		return occultationStationDensifyContours(corrected, cache, location)
	}
	visible := occultationStationVisibleEnvelopeContours(corrected, curves, cache, location)
	// Local corrected fragments do not establish a complete visible envelope.
	return occultationStationDensifyContours(visible, cache, location)
}

// occultationStationCorrectLimitSeries keeps the public north/south limit
// samples on the station contact curve. Static temporal envelopes are exposed
// separately through BandContours.
func occultationStationCorrectLimitSeries(
	points []OccultationPathPoint,
	frameAt occultationPathFrameFunc,
	contextAt occultationRiseSetContextFunc,
	total bool,
	location *time.Location,
) []OccultationPathPoint {
	if len(points) == 0 || frameAt == nil || contextAt == nil {
		return points
	}
	corrected := append([]OccultationPathPoint(nil), points...)
	for index, seed := range points {
		sample, ok := occultationStationCorrectBoundaryPoint(
			centerTimeTT(seed.Time), seed, frameAt, contextAt, total, location,
		)
		if !ok || !sample.valid {
			continue
		}
		point := sample.point
		// Public limit tracks retain their event sample times exactly; the
		// station solver may otherwise round-trip through civil time.
		point.Time = seed.Time
		point.WidthKM = seed.WidthKM
		corrected[index] = point
	}
	return corrected
}

func occultationStationEnvelopeSamples(
	points []OccultationPathPoint,
	contextAt occultationRiseSetContextFunc,
	total bool,
	location *time.Location,
) ([]occultationStationBoundarySample, []bool) {
	samples := make([]occultationStationBoundarySample, len(points))
	solved := make([]bool, len(points))
	try := func(index int, seed OccultationPathPoint) bool {
		sample, ok := occultationStationCorrectEnvelopePoint(
			centerTimeTT(points[index].Time), seed, contextAt, total, location,
		)
		if !ok || !sample.valid || occultationPathDistanceKM(points[index], sample.point) >
			occultationStationEnvelopeMaximumOffsetKM {
			return false
		}
		sample.offsetKM = occultationPathDistanceKM(points[index], sample.point)
		samples[index], solved[index] = sample, true
		return true
	}
	for index, seed := range points {
		// 延拓初值：轮廓上相邻点的解彼此接近，用上一个已收敛解做 Newton 初值通常能把
		// 迭代次数从十余次降到几次；若它落在原始点的容差之外或求解失败，再退回几何底点。
		// Continuation seed: neighbouring contour points solve to nearby stations, so the
		// previous converged solution is a much better Newton seed than the geometric point.
		// When it fails or lands outside the original point's tolerance, fall back to the
		// geometric seed.
		if index > 0 && solved[index-1] {
			neighbor := seed
			neighbor.Longitude = samples[index-1].point.Longitude
			neighbor.Latitude = samples[index-1].point.Latitude
			if try(index, neighbor) {
				continue
			}
		}
		try(index, seed)
	}
	for pass := 0; pass < 2; pass++ {
		for index := 1; index < len(points); index++ {
			if solved[index] || !solved[index-1] {
				continue
			}
			seed := points[index]
			seed.Longitude = samples[index-1].point.Longitude
			seed.Latitude = samples[index-1].point.Latitude
			try(index, seed)
		}
		for index := len(points) - 2; index >= 0; index-- {
			if solved[index] || !solved[index+1] {
				continue
			}
			seed := points[index]
			seed.Longitude = samples[index+1].point.Longitude
			seed.Latitude = samples[index+1].point.Latitude
			try(index, seed)
		}
	}
	return samples, solved
}

// occultationStationCorrectContactPoint refines an arbitrary contact-arc seed
// with a local two-dimensional Newton step.  The one-dimensional cross-track
// oracle is preferable for time-contour sides; this variant is for the two
// instantaneous contact arcs that close a direct static ring.
func occultationStationCorrectContactPoint(
	tt float64,
	seed OccultationPathPoint,
	contextAt occultationRiseSetContextFunc,
	total bool,
	location *time.Location,
) (occultationStationBoundarySample, bool) {
	if contextAt == nil {
		return occultationStationBoundarySample{}, false
	}
	context := contextAt(tt)
	if total {
		context = context.withInternalContact()
	}
	if !context.valid {
		return occultationStationBoundarySample{}, false
	}
	longitude, latitude := seed.Longitude, seed.Latitude
	seedResidual := math.NaN()
	for iteration := 0; iteration < 16; iteration++ {
		state := context.stateAt(longitude, latitude)
		if !state.valid || !finite(state.contactMetric) {
			return occultationStationBoundarySample{}, false
		}
		if iteration == 0 {
			seedResidual = state.contactMetric
		}
		if math.Abs(state.contactMetric) <= occultationStationOracleResidualToleranceDeg {
			return occultationStationSampleFromCoordinates(
				tt, seed, longitude, latitude, state.contactMetric, state.moonAltitude, seedResidual, location,
			), true
		}
		const coordinateStepDeg = 0.005
		plusLongitude := context.stateAt(normalizeLongitude(longitude+coordinateStepDeg), latitude)
		minusLongitude := context.stateAt(normalizeLongitude(longitude-coordinateStepDeg), latitude)
		plusLatitude := context.stateAt(longitude, math.Min(89.999999, latitude+coordinateStepDeg))
		minusLatitude := context.stateAt(longitude, math.Max(-89.999999, latitude-coordinateStepDeg))
		if !plusLongitude.valid || !minusLongitude.valid || !plusLatitude.valid || !minusLatitude.valid {
			return occultationStationBoundarySample{}, false
		}
		cosLatitude := math.Max(0.05, math.Cos(latitude*rad))
		gradientX := (plusLongitude.contactMetric - minusLongitude.contactMetric) / (2 * coordinateStepDeg * cosLatitude)
		gradientY := (plusLatitude.contactMetric - minusLatitude.contactMetric) / (2 * coordinateStepDeg)
		gradientSquared := gradientX*gradientX + gradientY*gradientY
		if !finite(gradientSquared) || gradientSquared <= 1e-18 {
			return occultationStationBoundarySample{}, false
		}
		deltaX := -state.contactMetric * gradientX / gradientSquared
		deltaY := -state.contactMetric * gradientY / gradientSquared
		length := math.Hypot(deltaX, deltaY)
		if length > 1.0 {
			scale := 1.0 / length
			deltaX *= scale
			deltaY *= scale
		}
		longitude = normalizeLongitude(longitude + deltaX/cosLatitude)
		latitude = math.Max(-89.999999, math.Min(89.999999, latitude+deltaY))
	}
	state := context.stateAt(longitude, latitude)
	if !state.valid || math.Abs(state.contactMetric) > 1e-5 {
		return occultationStationBoundarySample{}, false
	}
	return occultationStationSampleFromCoordinates(
		tt, seed, longitude, latitude, state.contactMetric, state.moonAltitude, seedResidual, location,
	), true
}

type occultationStationEnvelopeState struct {
	contactMetric     float64
	contactDerivative float64
	moonAltitude      float64
	valid             bool
}

// occultationStationCorrectEnvelopePoint solves the two necessary conditions
// for an interior boundary of the station-visible time union at a fixed time:
// the station contact gap is zero and stationary in time. The geocentric
// north/south limit is only a seed; both equations are evaluated with the same
// topocentric Moon/target vectors used by local occultation calculations.
func occultationStationCorrectEnvelopePoint(
	tt float64,
	seed OccultationPathPoint,
	contextAt occultationRiseSetContextFunc,
	total bool,
	location *time.Location,
) (occultationStationBoundarySample, bool) {
	if contextAt == nil {
		return occultationStationBoundarySample{}, false
	}
	contextFactory := contextAt
	if total {
		contextFactory = func(value float64) occultationRiseSetContext {
			return contextAt(value).withInternalContact()
		}
	}
	evaluation := occultationRiseSetEvaluation{
		tt:     tt,
		center: contextFactory(tt),
		before: contextFactory(tt - occultationRiseSetDerivativeStepDays),
		after:  contextFactory(tt + occultationRiseSetDerivativeStepDays),
	}
	if !evaluation.center.valid || !evaluation.before.valid || !evaluation.after.valid {
		return occultationStationBoundarySample{}, false
	}
	evaluate := func(longitude, latitude float64) occultationStationEnvelopeState {
		center := evaluation.center.stateAt(longitude, latitude)
		before := evaluation.before.stateAt(longitude, latitude)
		after := evaluation.after.stateAt(longitude, latitude)
		if !center.valid || !before.valid || !after.valid {
			return occultationStationEnvelopeState{}
		}
		derivative := (after.contactMetric - before.contactMetric) /
			(2 * occultationRiseSetDerivativeStepDays)
		return occultationStationEnvelopeState{
			contactMetric:     center.contactMetric,
			contactDerivative: derivative,
			moonAltitude:      center.moonAltitude,
			valid:             finite(center.contactMetric) && finite(derivative) && finite(center.moonAltitude),
		}
	}
	residualNorm := func(state occultationStationEnvelopeState) float64 {
		return math.Hypot(
			state.contactMetric,
			state.contactDerivative*occultationRiseSetDerivativeStepDays,
		)
	}

	longitude, latitude := seed.Longitude, seed.Latitude
	seedState := evaluate(longitude, latitude)
	if !seedState.valid {
		return occultationStationBoundarySample{}, false
	}
	for iteration := 0; iteration < 24; iteration++ {
		state := evaluate(longitude, latitude)
		if !state.valid {
			return occultationStationBoundarySample{}, false
		}
		if math.Abs(state.contactMetric) <= occultationStationOracleResidualToleranceDeg &&
			math.Abs(state.contactDerivative) <= occultationRiseSetJunctionDerivativeTolerance {
			return occultationStationSampleFromCoordinates(
				tt, seed, longitude, latitude, state.contactMetric, state.moonAltitude,
				seedState.contactMetric, location,
			), true
		}

		const coordinateStepDeg = 0.002
		cosLatitude := math.Max(0.05, math.Cos(latitude*rad))
		plusLongitude := evaluate(normalizeLongitude(longitude+coordinateStepDeg/cosLatitude), latitude)
		minusLongitude := evaluate(normalizeLongitude(longitude-coordinateStepDeg/cosLatitude), latitude)
		plusLatitude := evaluate(longitude, math.Min(89.999999, latitude+coordinateStepDeg))
		minusLatitude := evaluate(longitude, math.Max(-89.999999, latitude-coordinateStepDeg))
		if !plusLongitude.valid || !minusLongitude.valid || !plusLatitude.valid || !minusLatitude.valid {
			return occultationStationBoundarySample{}, false
		}
		contactX := (plusLongitude.contactMetric - minusLongitude.contactMetric) / (2 * coordinateStepDeg)
		contactY := (plusLatitude.contactMetric - minusLatitude.contactMetric) / (2 * coordinateStepDeg)
		derivativeX := (plusLongitude.contactDerivative - minusLongitude.contactDerivative) / (2 * coordinateStepDeg)
		derivativeY := (plusLatitude.contactDerivative - minusLatitude.contactDerivative) / (2 * coordinateStepDeg)
		determinant := contactX*derivativeY - contactY*derivativeX
		if !finite(determinant) || math.Abs(determinant) <= 1e-12 {
			return occultationStationBoundarySample{}, false
		}
		deltaX := (-state.contactMetric*derivativeY + contactY*state.contactDerivative) / determinant
		deltaY := (-contactX*state.contactDerivative + derivativeX*state.contactMetric) / determinant
		if !finite(deltaX) || !finite(deltaY) {
			return occultationStationBoundarySample{}, false
		}
		if length := math.Hypot(deltaX, deltaY); length > 6 {
			scale := 6 / length
			deltaX *= scale
			deltaY *= scale
		}

		currentNorm := residualNorm(state)
		accepted := false
		for damping := 1.0; damping >= 1.0/128; damping /= 2 {
			candidateLongitude := normalizeLongitude(longitude + damping*deltaX/cosLatitude)
			candidateLatitude := math.Max(-89.999999, math.Min(89.999999, latitude+damping*deltaY))
			if occultationPathDistanceKMValues(
				seed.Longitude, seed.Latitude, candidateLongitude, candidateLatitude,
			) > occultationStationEnvelopeMaximumOffsetKM {
				continue
			}
			candidate := evaluate(candidateLongitude, candidateLatitude)
			if !candidate.valid || residualNorm(candidate) >= currentNorm {
				continue
			}
			longitude, latitude = candidateLongitude, candidateLatitude
			accepted = true
			break
		}
		if !accepted {
			return occultationStationBoundarySample{}, false
		}
	}
	state := evaluate(longitude, latitude)
	if !state.valid || math.Abs(state.contactMetric) > 1e-5 ||
		math.Abs(state.contactDerivative) > 1e-4 {
		return occultationStationBoundarySample{}, false
	}
	return occultationStationSampleFromCoordinates(
		tt, seed, longitude, latitude, state.contactMetric, state.moonAltitude,
		seedState.contactMetric, location,
	), true
}

const (
	occultationStationEnvelopeTimeScale         = 360.0
	occultationStationEnvelopeArcStepDegrees    = 0.2
	occultationStationEnvelopeMinArcStepDegrees = 0.001
	occultationStationEnvelopeTargetSpacingKM   = 30.0
	occultationStationEnvelopeMaxArcSteps       = 8192
)

type occultationStationEnvelopeArcState struct {
	coordinates [3]float64
	tangent     [3]float64
	point       OccultationPathPoint
}

type occultationStationEnvelopeTrace struct {
	points   []OccultationPathPoint
	boundary bool
	closed   bool
}

type occultationStationEnvelopeKind uint8

const (
	occultationStationContactEnvelope occultationStationEnvelopeKind = iota
	occultationStationVisibilityEnvelope
)

type occultationStationEnvelopeModel struct {
	kind occultationStationEnvelopeKind
}

// occultationStationVisibleEnvelopeContours traces the station-centred static
// contact envelope as an implicit curve in longitude, latitude and time.  The
// pseudo-arclength parameter remains regular where fixed-time north/south roots
// meet, so a real temporal fold is preserved instead of becoming a branch jump.
func occultationStationVisibleEnvelopeContours(
	seeds [][]OccultationPathPoint,
	curves []OccultationRiseSetCurve,
	cache *occultationRiseSetEvaluationCache,
	location *time.Location,
) [][]OccultationPathPoint {
	model := occultationStationEnvelopeModel{kind: occultationStationContactEnvelope}
	minimumTT, maximumTT := math.Inf(1), math.Inf(-1)
	candidates := make([]OccultationPathPoint, 0, len(seeds))
	for _, segment := range seeds {
		if len(segment) < 2 {
			continue
		}
		best := OccultationPathPoint{MoonAltitude: math.Inf(-1)}
		for _, point := range segment {
			tt := centerTimeTT(point.Time)
			minimumTT = math.Min(minimumTT, tt)
			maximumTT = math.Max(maximumTT, tt)
			if point.MoonAltitude > best.MoonAltitude {
				best = point
			}
		}
		if best.MoonAltitude > occultationStationHorizonResidualToleranceDeg {
			candidates = append(candidates, best)
		}
	}
	// A geocentric north/south seed can locate only one component after station
	// parallax changes the topology. Every start/end phase junction is an exact
	// endpoint of a visible contact-envelope component, so use those junctions
	// to discover any remaining component without inventing a connector.
	for _, curve := range curves {
		if curve.Phase != RiseSetPhaseStart && curve.Phase != RiseSetPhaseEnd {
			continue
		}
		for _, segment := range curve.Segments {
			for _, point := range segment {
				tt := centerTimeTT(point.Time)
				minimumTT = math.Min(minimumTT, tt)
				maximumTT = math.Max(maximumTT, tt)
			}
			if len(segment) == 0 {
				continue
			}
			for _, endpoint := range []OccultationPathPoint{segment[0], segment[len(segment)-1]} {
				junction, ok := refineOccultationRiseSetPhaseJunction(endpoint, location, cache)
				if ok && !occultationStationEnvelopePointCovered(
					[][]OccultationPathPoint{candidates}, junction, 1,
				) {
					candidates = append(candidates, junction)
				}
			}
		}
	}
	if len(candidates) == 0 || !finite(minimumTT) || !finite(maximumTT) {
		return nil
	}
	// The geocentric seeds only locate the component. Station parallax can move a
	// temporal fold just beyond their fixed-time range, so leave a bounded margin.
	minimumTT -= 0.05
	maximumTT += 0.05
	result := make([][]OccultationPathPoint, 0, len(candidates))
	for _, candidate := range candidates {
		if occultationStationEnvelopePointCovered(result, candidate, occultationStationEnvelopeTargetSpacingKM) {
			continue
		}
		referenceTT := centerTimeTT(candidate.Time)
		if candidate.MoonAltitude <= occultationStationHorizonResidualToleranceDeg {
			for _, direction := range []int{-1, 1} {
				trace := occultationStationTraceEnvelope(
					candidate, direction, referenceTT, minimumTT, maximumTT, cache, model, location,
				)
				if !trace.boundary || len(trace.points) < 3 ||
					occultationPathDistanceKM(trace.points[0], trace.points[len(trace.points)-1]) <= 1 {
					continue
				}
				trace.points = occultationStationDeduplicateEnvelopePoints(trace.points)
				result = append(result, occultationStationSplitEnvelopeAtTimeFolds(trace.points)...)
				break
			}
			continue
		}
		backward := occultationStationTraceEnvelope(
			candidate, -1, referenceTT, minimumTT, maximumTT, cache, model, location,
		)
		if backward.closed {
			closed := append([]OccultationPathPoint(nil), backward.points...)
			if len(closed) >= 4 {
				closed[len(closed)-1] = closed[0]
				result = append(result, occultationStationSplitEnvelopeAtTimeFolds(closed)...)
			}
			continue
		}
		forward := occultationStationTraceEnvelope(
			candidate, 1, referenceTT, minimumTT, maximumTT, cache, model, location,
		)
		if !backward.boundary || !forward.boundary {
			continue
		}
		segment := make([]OccultationPathPoint, 0, len(backward.points)+len(forward.points)-1)
		for index := len(backward.points) - 1; index >= 0; index-- {
			segment = append(segment, backward.points[index])
		}
		segment = append(segment, forward.points[1:]...)
		segment = occultationStationDeduplicateEnvelopePoints(segment)
		if len(segment) >= 3 {
			result = append(result, occultationStationSplitEnvelopeAtTimeFolds(segment)...)
		}
	}
	return result
}

// occultationStationVisibilityEnvelopeContours traces the boundary of the
// time-union of Moon-above-horizon states. It is the implicit curve H=0,
// dH/dt=0 restricted to sites where the requested contact metric is negative
// and H has a temporal maximum. Endpoints are the exact [F,H,dH/dt]=0
// direction junctions shared by start/end moonrise and moonset phase curves.
func occultationStationVisibilityEnvelopeContours(
	curves []OccultationRiseSetCurve,
	cache *occultationRiseSetEvaluationCache,
	location *time.Location,
) [][]OccultationPathPoint {
	if len(curves) == 0 || cache == nil {
		return nil
	}
	minimumTT, maximumTT := math.Inf(1), math.Inf(-1)
	seeds := make([]OccultationPathPoint, 0, 8)
	for _, curve := range curves {
		if curve.Phase != RiseSetPhaseStart && curve.Phase != RiseSetPhaseEnd {
			continue
		}
		for _, segment := range curve.Segments {
			for _, point := range segment {
				tt := centerTimeTT(point.Time)
				minimumTT = math.Min(minimumTT, tt)
				maximumTT = math.Max(maximumTT, tt)
			}
			if len(segment) == 0 {
				continue
			}
			for _, candidate := range []OccultationPathPoint{segment[0], segment[len(segment)-1]} {
				junction, ok := refineOccultationRiseSetDirectionJunction(
					centerTimeTT(candidate.Time), candidate.Longitude, candidate.Latitude,
					false, location, cache,
				)
				if !ok {
					junction, ok = refineOccultationRiseSetDirectionJunctionOnHorizon(candidate, candidate, false, location, cache)
				}
				if ok && !occultationStationEnvelopePointCovered(
					[][]OccultationPathPoint{seeds}, junction, 1,
				) {
					seeds = append(seeds, junction)
				}
			}
		}
	}
	if len(seeds) == 0 || !finite(minimumTT) || !finite(maximumTT) {
		return nil
	}
	minimumTT -= 0.05
	maximumTT += 0.05
	model := occultationStationEnvelopeModel{kind: occultationStationVisibilityEnvelope}
	result := make([][]OccultationPathPoint, 0, len(seeds)/2+1)
	for _, seed := range seeds {
		if !model.isRequiredExtremum(cache.evaluation(centerTimeTT(seed.Time)), seed.Longitude, seed.Latitude) {
			continue
		}
		if occultationStationEnvelopePointCovered(result, seed, occultationStationEnvelopeTargetSpacingKM) {
			continue
		}
		referenceTT := centerTimeTT(seed.Time)
		for _, direction := range []int{-1, 1} {
			trace := occultationStationTraceEnvelope(
				seed, direction, referenceTT, minimumTT, maximumTT, cache, model, location,
			)
			if trace.closed && len(trace.points) >= 4 {
				closed := append([]OccultationPathPoint(nil), trace.points...)
				closed[len(closed)-1] = closed[0]
				result = append(result, occultationStationSplitEnvelopeAtTimeFolds(closed)...)
				break
			}
			if !trace.boundary || len(trace.points) < 3 ||
				occultationPathDistanceKM(trace.points[0], trace.points[len(trace.points)-1]) <= 1 {
				continue
			}
			trace.points = occultationStationDeduplicateEnvelopePoints(trace.points)
			result = append(result, occultationStationSplitEnvelopeAtTimeFolds(trace.points)...)
			break
		}
	}
	return result
}

func occultationStationTraceEnvelope(
	seed OccultationPathPoint,
	direction int,
	referenceTT, minimumTT, maximumTT float64,
	cache *occultationRiseSetEvaluationCache,
	model occultationStationEnvelopeModel,
	location *time.Location,
) occultationStationEnvelopeTrace {
	state, ok := occultationStationEnvelopeArcStateAt(seed, referenceTT, cache, model)
	if !ok {
		return occultationStationEnvelopeTrace{points: []OccultationPathPoint{seed}}
	}
	for index := range state.tangent {
		state.tangent[index] *= float64(direction)
	}
	result := occultationStationEnvelopeTrace{points: []OccultationPathPoint{state.point}}
	step := occultationStationEnvelopeArcStepDegrees
	travelKM := 0.0
	for count := 0; count < occultationStationEnvelopeMaxArcSteps; count++ {
		predictor := state.coordinates
		for index := range predictor {
			predictor[index] += step * state.tangent[index]
		}
		next, iterations, solved := occultationStationCorrectEnvelopeArc(
			predictor, state.tangent, referenceTT, cache, model, state.point.WidthKM, location,
		)
		if !solved {
			step /= 2
			if step < occultationStationEnvelopeMinArcStepDegrees {
				return result
			}
			continue
		}
		if dotSolarEclipse3(next.tangent, state.tangent) < 0 {
			for index := range next.tangent {
				next.tangent[index] = -next.tangent[index]
			}
		}
		distanceKM := occultationPathDistanceKM(state.point, next.point)
		if !finite(distanceKM) || distanceKM > occultationStationEnvelopeTargetSpacingKM {
			step /= 2
			if step < occultationStationEnvelopeMinArcStepDegrees {
				return result
			}
			continue
		}
		nextTT := centerTimeTT(next.point.Time)
		if nextTT < minimumTT || nextTT > maximumTT {
			return result
		}
		if model.activeMargin(cache.evaluation(nextTT), next.point.Longitude, next.point.Latitude) < 0 {
			junction, junctionOK := occultationStationEnvelopeActivityJunction(
				state.point, next.point, location, cache, model,
			)
			if !junctionOK {
				step /= 2
				if step < occultationStationEnvelopeMinArcStepDegrees {
					return result
				}
				continue
			}
			result.points = append(result.points, junction)
			result.boundary = true
			return result
		}
		travelKM += distanceKM
		if travelKM > 4*occultationStationEnvelopeTargetSpacingKM &&
			occultationPathDistanceKM(seed, next.point) <= occultationStationEnvelopeTargetSpacingKM/2 {
			result.points = append(result.points, seed)
			result.closed = true
			return result
		}
		result.points = append(result.points, next.point)
		state = next
		if model.activeMargin(cache.evaluation(nextTT), next.point.Longitude, next.point.Latitude) <=
			occultationStationHorizonResidualToleranceDeg {
			result.boundary = true
			return result
		}
		if distanceKM < occultationStationEnvelopeTargetSpacingKM/2 && iterations <= 4 {
			step = math.Min(occultationStationEnvelopeArcStepDegrees, step*1.5)
		}
	}
	return result
}

func occultationStationEnvelopeArcStateAt(
	point OccultationPathPoint,
	referenceTT float64,
	cache *occultationRiseSetEvaluationCache,
	model occultationStationEnvelopeModel,
) (occultationStationEnvelopeArcState, bool) {
	coordinates := [3]float64{
		point.Longitude,
		point.Latitude,
		(centerTimeTT(point.Time) - referenceTT) * occultationStationEnvelopeTimeScale,
	}
	_, jacobian, ok := occultationStationEnvelopeJacobian(coordinates, referenceTT, cache, model)
	if !ok {
		return occultationStationEnvelopeArcState{}, false
	}
	tangent, ok := occultationStationEnvelopeTangent(jacobian)
	if !ok {
		return occultationStationEnvelopeArcState{}, false
	}
	return occultationStationEnvelopeArcState{coordinates: coordinates, tangent: tangent, point: point}, true
}

func occultationStationCorrectEnvelopeArc(
	predictor, tangent [3]float64,
	referenceTT float64,
	cache *occultationRiseSetEvaluationCache,
	model occultationStationEnvelopeModel,
	widthKM float64,
	location *time.Location,
) (occultationStationEnvelopeArcState, int, bool) {
	coordinates := predictor
	iterationCache := cache.candidateOnly()
	jacobianCache := iterationCache
	for iteration := 0; iteration < 20; iteration++ {
		residual, jacobian, ok := occultationStationEnvelopeJacobian(coordinates, referenceTT, jacobianCache, model)
		if iterationCache == cache && ok {
			tt := referenceTT + coordinates[2]/occultationStationEnvelopeTimeScale
			residual, ok = model.residual(cache.evaluation(tt), normalizeLongitude(coordinates[0]), coordinates[1])
		}
		if !ok {
			if jacobianCache != cache {
				iterationCache = cache
				jacobianCache = cache
				coordinates = predictor
				continue
			}
			return occultationStationEnvelopeArcState{}, iteration, false
		}
		planeResidual := dotSolarEclipse3(subtractSolarEclipse3(coordinates, predictor), tangent)
		// Interpolation predicts the Newton root; accepted vertices must pass
		// the original exact residual and extremum checks. The candidate
		// Jacobian only predicts corrections and the next tracing direction.
		if iterationCache != cache && (iteration >= 5 ||
			math.Abs(residual[0]) <= model.valueTolerance() && math.Abs(residual[1]) <= model.derivativeTolerance()) {
			iterationCache = cache
			continue
		}
		// Retain a converged exact root before finite-difference noise can
		// move the derivative back outside the final acceptance tolerance.
		if math.Abs(residual[0]) <= 1e-9 && math.Abs(residual[1]) <= model.derivativeTolerance() && math.Abs(planeResidual) <= 1e-9 {
			return occultationStationValidEnvelopeArcState(
				coordinates, jacobian, referenceTT, cache, model, widthKM, location, iteration+1,
			)
		}
		matrix := [3][3]float64{jacobian[0], jacobian[1], tangent}
		delta, ok := solveSolarEclipse3x3(
			matrix, [3]float64{-residual[0], -residual[1], -planeResidual},
		)
		if !ok {
			return occultationStationEnvelopeArcState{}, iteration, false
		}
		if norm := math.Sqrt(dotSolarEclipse3(delta, delta)); norm > 1 {
			for index := range delta {
				delta[index] /= norm
			}
		}
		for index := range coordinates {
			coordinates[index] += delta[index]
		}
		coordinates[0] = predictor[0] + math.Remainder(coordinates[0]-predictor[0], 360)
		if coordinates[1] <= -89.999999 || coordinates[1] >= 89.999999 {
			return occultationStationEnvelopeArcState{}, iteration, false
		}
	}
	residual, jacobian, ok := occultationStationEnvelopeJacobian(coordinates, referenceTT, cache, model)
	planeResidual := dotSolarEclipse3(subtractSolarEclipse3(coordinates, predictor), tangent)
	if !ok || math.Abs(residual[0]) > model.valueTolerance() ||
		math.Abs(residual[1]) > model.derivativeTolerance() ||
		math.Abs(planeResidual) > 1e-7 {
		return occultationStationRefineEnvelopeAtFixedTime(
			coordinates, predictor, tangent, referenceTT, cache, model, widthKM, location,
		)
	}
	return occultationStationValidEnvelopeArcState(
		coordinates, jacobian, referenceTT, cache, model, widthKM, location, 20,
	)
}

// At the Julian-day rounding floor, changing time can oscillate between two
// derivative values. A bounded spatial correction keeps the physical equations
// exact; only the auxiliary arclength plane may move by one minimum trace step.
func occultationStationRefineEnvelopeAtFixedTime(
	coordinates, predictor, tangent [3]float64,
	referenceTT float64,
	cache *occultationRiseSetEvaluationCache,
	model occultationStationEnvelopeModel,
	widthKM float64,
	location *time.Location,
) (occultationStationEnvelopeArcState, int, bool) {
	origin := coordinates
	for iteration := 0; iteration < 4; iteration++ {
		residual, jacobian, ok := occultationStationEnvelopeJacobian(coordinates, referenceTT, cache, model)
		if !ok || coordinates[1] <= -89.999999 || coordinates[1] >= 89.999999 {
			break
		}
		planeResidual := dotSolarEclipse3(subtractSolarEclipse3(coordinates, predictor), tangent)
		if math.Abs(planeResidual) > occultationStationEnvelopeMinArcStepDegrees {
			break
		}
		if math.Abs(residual[0]) <= model.valueTolerance() && math.Abs(residual[1]) <= model.derivativeTolerance() {
			return occultationStationValidEnvelopeArcState(
				coordinates, jacobian, referenceTT, cache, model, widthKM, location, 20+iteration,
			)
		}
		determinant := jacobian[0][0]*jacobian[1][1] - jacobian[0][1]*jacobian[1][0]
		if !finite(determinant) || math.Abs(determinant) < 1e-12 {
			break
		}
		longitudeStep := (-residual[0]*jacobian[1][1] + residual[1]*jacobian[0][1]) / determinant
		latitudeStep := (-residual[1]*jacobian[0][0] + residual[0]*jacobian[1][0]) / determinant
		coordinates[0] += longitudeStep
		coordinates[1] += latitudeStep
		if !finite(coordinates[0]) || !finite(coordinates[1]) ||
			math.Hypot(coordinates[0]-origin[0], coordinates[1]-origin[1]) > occultationStationEnvelopeMinArcStepDegrees {
			break
		}
	}
	return occultationStationEnvelopeArcState{}, 24, false
}

func occultationStationValidEnvelopeArcState(
	coordinates [3]float64,
	jacobian [2][3]float64,
	referenceTT float64,
	cache *occultationRiseSetEvaluationCache,
	model occultationStationEnvelopeModel,
	widthKM float64,
	location *time.Location,
	iterations int,
) (occultationStationEnvelopeArcState, int, bool) {
	tt := referenceTT + coordinates[2]/occultationStationEnvelopeTimeScale
	longitude, latitude := normalizeLongitude(coordinates[0]), coordinates[1]
	evaluation := cache.evaluation(tt)
	state := evaluation.center.stateAt(longitude, latitude)
	residual, ok := model.residual(evaluation, longitude, latitude)
	if !ok || math.Abs(residual[0]) > model.valueTolerance() ||
		math.Abs(residual[1]) > model.derivativeTolerance() ||
		!model.isRequiredExtremum(evaluation, longitude, latitude) {
		return occultationStationEnvelopeArcState{}, iterations, false
	}
	tangent, ok := occultationStationEnvelopeTangent(jacobian)
	if !ok {
		return occultationStationEnvelopeArcState{}, iterations, false
	}
	return occultationStationEnvelopeArcState{
		coordinates: coordinates,
		tangent:     tangent,
		point: OccultationPathPoint{
			Time: occultationTTToLocation(tt, location), Longitude: longitude,
			Latitude: latitude, MoonAltitude: state.moonAltitude, WidthKM: widthKM,
		},
	}, iterations, true
}

func occultationStationEnvelopeJacobian(
	coordinates [3]float64,
	referenceTT float64,
	cache *occultationRiseSetEvaluationCache,
	model occultationStationEnvelopeModel,
) ([2]float64, [2][3]float64, bool) {
	tt := referenceTT + coordinates[2]/occultationStationEnvelopeTimeScale
	longitude, latitude := normalizeLongitude(coordinates[0]), coordinates[1]
	evaluation := cache.evaluation(tt)
	residual, ok := model.residual(evaluation, longitude, latitude)
	if !ok {
		return [2]float64{}, [2][3]float64{}, false
	}
	steps := [3]float64{1e-4, 1e-4, 5 * occultationStationEnvelopeTimeScale / 86400}
	jacobian := [2][3]float64{}
	for column := range steps {
		shifted := coordinates
		shifted[column] += steps[column]
		shiftedTT := referenceTT + shifted[2]/occultationStationEnvelopeTimeScale
		shiftedEvaluation := cache.evaluation(shiftedTT)
		shiftedResidual, shiftedOK := model.residual(
			shiftedEvaluation, normalizeLongitude(shifted[0]), shifted[1],
		)
		if !shiftedOK {
			return [2]float64{}, [2][3]float64{}, false
		}
		for row := range residual {
			jacobian[row][column] = (shiftedResidual[row] - residual[row]) / steps[column]
		}
	}
	return residual, jacobian, true
}

func (model occultationStationEnvelopeModel) residual(
	evaluation occultationRiseSetEvaluation,
	longitude, latitude float64,
) ([2]float64, bool) {
	state := evaluation.center.stateAt(longitude, latitude)
	if model.kind == occultationStationVisibilityEnvelope {
		derivative := evaluation.moonAltitudeDerivative(longitude, latitude)
		return [2]float64{state.moonAltitude, derivative}, state.valid && finite(derivative)
	}
	derivative := evaluation.contactDerivative(longitude, latitude)
	return [2]float64{state.contactMetric, derivative}, state.valid && finite(derivative)
}

func (model occultationStationEnvelopeModel) valueTolerance() float64 {
	if model.kind == occultationStationVisibilityEnvelope {
		return occultationStationHorizonResidualToleranceDeg
	}
	return 1e-5
}

func (model occultationStationEnvelopeModel) derivativeTolerance() float64 {
	return occultationRiseSetJunctionDerivativeTolerance
}

func (model occultationStationEnvelopeModel) isRequiredExtremum(
	evaluation occultationRiseSetEvaluation,
	longitude, latitude float64,
) bool {
	if model.kind == occultationStationVisibilityEnvelope {
		return evaluation.moonAltitudeSecondDerivative(longitude, latitude) < 0
	}
	return evaluation.contactSecondDerivative(longitude, latitude) > 0
}

func (model occultationStationEnvelopeModel) activeMargin(
	evaluation occultationRiseSetEvaluation,
	longitude, latitude float64,
) float64 {
	state := evaluation.center.stateAt(longitude, latitude)
	if !state.valid {
		return math.Inf(-1)
	}
	if model.kind == occultationStationVisibilityEnvelope {
		return -state.contactMetric
	}
	return state.moonAltitude
}

func occultationStationEnvelopeTangent(jacobian [2][3]float64) ([3]float64, bool) {
	tangent := [3]float64{
		jacobian[0][1]*jacobian[1][2] - jacobian[0][2]*jacobian[1][1],
		jacobian[0][2]*jacobian[1][0] - jacobian[0][0]*jacobian[1][2],
		jacobian[0][0]*jacobian[1][1] - jacobian[0][1]*jacobian[1][0],
	}
	norm := math.Sqrt(dotSolarEclipse3(tangent, tangent))
	if !finite(norm) || norm < 1e-14 {
		return [3]float64{}, false
	}
	for index := range tangent {
		tangent[index] /= norm
	}
	return tangent, true
}

func occultationStationEnvelopeActivityJunction(
	inside, outside OccultationPathPoint,
	location *time.Location,
	cache *occultationRiseSetEvaluationCache,
	model occultationStationEnvelopeModel,
) (OccultationPathPoint, bool) {
	seed := occultationRiseSetMidpoint(inside, outside)
	var junction OccultationPathPoint
	var ok bool
	if model.kind == occultationStationVisibilityEnvelope {
		junction, ok = refineOccultationRiseSetDirectionJunction(
			centerTimeTT(seed.Time), seed.Longitude, seed.Latitude, false, location, cache,
		)
		if !ok {
			junction, ok = refineOccultationRiseSetDirectionJunctionOnHorizon(inside, outside, false, location, cache)
		}
	} else {
		junction, ok = refineOccultationRiseSetPhaseJunction(seed, location, cache)
		if !ok {
			junction, ok = refineOccultationRiseSetPhaseJunctionOnHorizon(seed, location, cache)
		}
	}
	if !ok {
		return OccultationPathPoint{}, false
	}
	minimumTT := math.Min(centerTimeTT(inside.Time), centerTimeTT(outside.Time)) - 1.0/86400
	maximumTT := math.Max(centerTimeTT(inside.Time), centerTimeTT(outside.Time)) + 1.0/86400
	if tt := centerTimeTT(junction.Time); tt < minimumTT || tt > maximumTT {
		return OccultationPathPoint{}, false
	}
	maximumDistance := math.Max(10, 3*occultationPathDistanceKM(inside, outside))
	if occultationPathDistanceKM(junction, inside) > maximumDistance ||
		occultationPathDistanceKM(junction, outside) > maximumDistance {
		return OccultationPathPoint{}, false
	}
	return junction, true
}

func occultationStationEnvelopePointCovered(
	segments [][]OccultationPathPoint,
	point OccultationPathPoint,
	toleranceKM float64,
) bool {
	for _, segment := range segments {
		for _, existing := range segment {
			if occultationPathDistanceKM(existing, point) <= toleranceKM {
				return true
			}
		}
	}
	return false
}

func occultationStationDeduplicateEnvelopePoints(points []OccultationPathPoint) []OccultationPathPoint {
	if len(points) < 2 {
		return points
	}
	result := make([]OccultationPathPoint, 0, len(points))
	for _, point := range points {
		if len(result) == 0 || occultationPathDistanceKM(result[len(result)-1], point) > 0.001 {
			result = append(result, point)
		}
	}
	return result
}

func occultationStationSplitEnvelopeAtTimeFolds(
	points []OccultationPathPoint,
) [][]OccultationPathPoint {
	if len(points) < 2 {
		return nil
	}
	// Station correction can produce the same civil timestamp twice at a
	// temporal-envelope junction. Collapse coincident samples before splitting
	// folds; retaining both would violate the public strictly-increasing-time
	// contour contract while adding no geometry.
	normalized := make([]OccultationPathPoint, 0, len(points))
	for _, point := range points {
		if len(normalized) > 0 && point.Time.Equal(normalized[len(normalized)-1].Time) &&
			occultationPathDistanceKM(point, normalized[len(normalized)-1]) <= 0.01 {
			normalized[len(normalized)-1] = point
			continue
		}
		normalized = append(normalized, point)
	}
	points = normalized
	if len(points) < 2 {
		return nil
	}
	segments := make([][]OccultationPathPoint, 0, 4)
	start := 0
	direction := 0
	appendSegment := func(first, end, sign int) {
		if end-first < 2 {
			return
		}
		segment := append([]OccultationPathPoint(nil), points[first:end]...)
		if sign < 0 {
			for left, right := 0, len(segment)-1; left < right; left, right = left+1, right-1 {
				segment[left], segment[right] = segment[right], segment[left]
			}
		}
		segments = append(segments, segment)
	}
	// Public contour segments must be strictly monotonic. A sub-millisecond
	// reversal is still a real time fold when the traced points are spatially
	// distinct, so split it without the coarser rise/set junction tolerance.
	for index := 1; index < len(points); index++ {
		sign := 0
		if points[index].Time.After(points[index-1].Time) {
			sign = 1
		} else if points[index].Time.Before(points[index-1].Time) {
			sign = -1
		}
		if sign == 0 {
			// Distinct points at one timestamp are a genuine temporal fold. End
			// the preceding monotone segment and restart at the second branch.
			if direction != 0 {
				appendSegment(start, index, direction)
			}
			start = index
			direction = 0
			continue
		}
		if direction == 0 {
			direction = sign
			continue
		}
		if sign == direction {
			continue
		}
		appendSegment(start, index, direction)
		start = index - 1
		direction = sign
	}
	appendSegment(start, len(points), direction)
	return segments
}

// occultationStationCorrectHorizonContactPoint refines an open footprint
// endpoint onto the intersection of the fixed-time contact curve and the lunar
// horizon. Solving only the contact equation leaves one unconstrained surface
// direction and lets independently sampled endpoints drift along the contact
// curve, which turns their temporal outline into a scalloped static boundary.
func occultationStationCorrectHorizonContactPoint(
	tt float64,
	seed OccultationPathPoint,
	contextAt occultationRiseSetContextFunc,
	total bool,
	location *time.Location,
) (occultationStationBoundarySample, bool) {
	if contextAt == nil {
		return occultationStationBoundarySample{}, false
	}
	context := contextAt(tt)
	if total {
		context = context.withInternalContact()
	}
	if !context.valid {
		return occultationStationBoundarySample{}, false
	}
	longitude, latitude := seed.Longitude, seed.Latitude
	seedState := context.stateAt(longitude, latitude)
	if !seedState.valid {
		return occultationStationBoundarySample{}, false
	}
	seedResidual := seedState.contactMetric
	residualNorm := func(state occultationRiseSetState) float64 {
		return math.Hypot(state.contactMetric, state.moonAltitude)
	}
	for iteration := 0; iteration < 20; iteration++ {
		state := context.stateAt(longitude, latitude)
		if !state.valid || !finite(state.contactMetric) || !finite(state.moonAltitude) {
			return occultationStationBoundarySample{}, false
		}
		if math.Abs(state.contactMetric) <= occultationStationOracleResidualToleranceDeg &&
			math.Abs(state.moonAltitude) <= occultationStationHorizonResidualToleranceDeg {
			return occultationStationSampleFromCoordinates(
				tt, seed, longitude, latitude, state.contactMetric, state.moonAltitude, seedResidual, location,
			), true
		}

		const coordinateStepDeg = 0.002
		plusLongitude := context.stateAt(normalizeLongitude(longitude+coordinateStepDeg), latitude)
		minusLongitude := context.stateAt(normalizeLongitude(longitude-coordinateStepDeg), latitude)
		plusLatitude := context.stateAt(longitude, math.Min(89.999999, latitude+coordinateStepDeg))
		minusLatitude := context.stateAt(longitude, math.Max(-89.999999, latitude-coordinateStepDeg))
		if !plusLongitude.valid || !minusLongitude.valid || !plusLatitude.valid || !minusLatitude.valid {
			return occultationStationBoundarySample{}, false
		}
		cosLatitude := math.Max(0.05, math.Cos(latitude*rad))
		contactX := (plusLongitude.contactMetric - minusLongitude.contactMetric) / (2 * coordinateStepDeg * cosLatitude)
		contactY := (plusLatitude.contactMetric - minusLatitude.contactMetric) / (2 * coordinateStepDeg)
		altitudeX := (plusLongitude.moonAltitude - minusLongitude.moonAltitude) / (2 * coordinateStepDeg * cosLatitude)
		altitudeY := (plusLatitude.moonAltitude - minusLatitude.moonAltitude) / (2 * coordinateStepDeg)
		determinant := contactX*altitudeY - contactY*altitudeX
		if !finite(determinant) || math.Abs(determinant) <= 1e-12 {
			return occultationStationBoundarySample{}, false
		}
		deltaX := (-state.contactMetric*altitudeY + contactY*state.moonAltitude) / determinant
		deltaY := (-contactX*state.moonAltitude + altitudeX*state.contactMetric) / determinant
		if !finite(deltaX) || !finite(deltaY) {
			return occultationStationBoundarySample{}, false
		}
		if length := math.Hypot(deltaX, deltaY); length > 0.5 {
			scale := 0.5 / length
			deltaX *= scale
			deltaY *= scale
		}

		currentNorm := residualNorm(state)
		accepted := false
		for damping := 1.0; damping >= 1.0/64; damping /= 2 {
			candidateLongitude := normalizeLongitude(longitude + damping*deltaX/cosLatitude)
			candidateLatitude := math.Max(-89.999999, math.Min(89.999999, latitude+damping*deltaY))
			if occultationPathDistanceKMValues(
				seed.Longitude, seed.Latitude, candidateLongitude, candidateLatitude,
			) > occultationStationHorizonMaximumOffsetKM {
				continue
			}
			candidate := context.stateAt(candidateLongitude, candidateLatitude)
			if !candidate.valid || residualNorm(candidate) >= currentNorm {
				continue
			}
			longitude, latitude = candidateLongitude, candidateLatitude
			accepted = true
			break
		}
		if !accepted {
			return occultationStationBoundarySample{}, false
		}
	}
	state := context.stateAt(longitude, latitude)
	if !state.valid || math.Abs(state.contactMetric) > occultationStationHorizonAcceptanceDeg ||
		math.Abs(state.moonAltitude) > occultationStationHorizonAcceptanceDeg {
		return occultationStationBoundarySample{}, false
	}
	return occultationStationSampleFromCoordinates(
		tt, seed, longitude, latitude, state.contactMetric, state.moonAltitude, seedResidual, location,
	), true
}

func occultationStationSampleFromCoordinates(
	tt float64,
	seed OccultationPathPoint,
	longitude, latitude, residual, altitude, seedResidual float64,
	location *time.Location,
) occultationStationBoundarySample {
	offset := occultationPathDistanceKMValues(seed.Longitude, seed.Latitude, longitude, latitude)
	point := seed
	point.Time = occultationTTToLocation(tt, location)
	point.Longitude = longitude
	point.Latitude = latitude
	point.MoonAltitude = altitude
	return occultationStationBoundarySample{
		point:              point,
		contactResidualDeg: residual,
		horizonResidualDeg: altitude,
		offsetKM:           offset,
		seedResidualDeg:    seedResidual,
		valid:              finite(residual) && finite(altitude),
	}
}

// occultationStationCorrectFootprintEdges corrects only the two instantaneous
// contact arcs used by a direct ring and the endpoints that form their time
// tracks. Interior horizon samples remain untouched and retain their original
// footprint closure contract.
func occultationStationCorrectFootprintEdges(
	footprints []PlanetOccultationFootprint,
	frameAt occultationPathFrameFunc,
	contextAt occultationRiseSetContextFunc,
	total bool,
	location *time.Location,
) []PlanetOccultationFootprint {
	if len(footprints) == 0 || frameAt == nil || contextAt == nil {
		return footprints
	}
	result := make([]PlanetOccultationFootprint, len(footprints))
	for index, footprint := range footprints {
		result[index] = footprint
		result[index].Boundaries = make([][]OccultationPathPoint, len(footprint.Boundaries))
		for boundaryIndex, boundary := range footprint.Boundaries {
			result[index].Boundaries[boundaryIndex] = append([]OccultationPathPoint(nil), boundary...)
		}
	}
	for footprintIndex := range result {
		for boundaryIndex := range result[footprintIndex].Boundaries {
			boundary := result[footprintIndex].Boundaries[boundaryIndex]
			for pointIndex := range boundary {
				if pointIndex != 0 && pointIndex+1 != len(boundary) && footprintIndex != 0 && footprintIndex+1 != len(result) {
					continue
				}
				seed := boundary[pointIndex]
				correct := occultationStationCorrectContactPoint
				if !result[footprintIndex].Closed && (pointIndex == 0 || pointIndex+1 == len(boundary)) {
					correct = occultationStationCorrectHorizonContactPoint
				}
				sample, ok := correct(centerTimeTT(seed.Time), seed, contextAt, total, location)
				if ok && sample.valid {
					corrected := sample.point
					// Station solving may round-trip TT through a civil-time
					// conversion with sub-millisecond drift. Footprint contracts
					// require every boundary point to retain its parent sample time.
					corrected.Time = seed.Time
					// 验收带内的负高度是求解器噪声，导出契约要求边界切点高度非负。
					if corrected.MoonAltitude < 0 && corrected.MoonAltitude >= -occultationStationHorizonAcceptanceDeg {
						corrected.MoonAltitude = 0
					}
					boundary[pointIndex] = corrected
				}
			}
			result[footprintIndex].Boundaries[boundaryIndex] = boundary
		}
		result[footprintIndex].Polygons = occultationStationFootprintPolygons(
			result[footprintIndex], frameAt, location,
		)
		// Horizon-arc points are recomputed from TT and can differ from the
		// parent civil timestamp by a few microseconds after the station solve.
		// Polygon samples are instantaneous geometry, so normalize their time
		// metadata to the owning footprint before public validation/serialization.
		for polygonIndex := range result[footprintIndex].Polygons {
			for pointIndex := range result[footprintIndex].Polygons[polygonIndex] {
				result[footprintIndex].Polygons[polygonIndex][pointIndex].Time = result[footprintIndex].Time
			}
		}
	}
	return result
}

func occultationStationFootprintPolygons(
	footprint PlanetOccultationFootprint,
	frameAt occultationPathFrameFunc,
	location *time.Location,
) [][]OccultationPathPoint {
	if len(footprint.Boundaries) == 0 {
		return footprint.Polygons
	}
	if footprint.Closed {
		polygons := make([][]OccultationPathPoint, 0, len(footprint.Boundaries)+len(footprint.InteriorPolygons))
		for _, boundary := range footprint.Boundaries {
			if len(boundary) < 3 {
				continue
			}
			polygon := append([]OccultationPathPoint(nil), boundary...)
			if occultationPathDistanceKM(polygon[0], polygon[len(polygon)-1]) > 0.001 {
				polygon = append(polygon, polygon[0])
			}
			polygons = append(polygons, polygon)
		}
		for _, interior := range footprint.InteriorPolygons {
			polygons = append(polygons, append([]OccultationPathPoint(nil), interior...))
		}
		if len(polygons) > 0 {
			return polygons
		}
		return footprint.Polygons
	}
	tt := centerTimeTT(footprint.Time)
	frame, ok := frameAt(tt)
	if !ok {
		return footprint.Polygons
	}
	polygons := make([][]OccultationPathPoint, 0, len(footprint.Boundaries)+len(footprint.InteriorPolygons))
	for _, boundary := range footprint.Boundaries {
		if len(boundary) < 2 {
			continue
		}
		polygon := append([]OccultationPathPoint(nil), boundary...)
		polygon = append(polygon, planetOccultationHorizonArc(
			tt, frame, boundary, location, planetOccultationBandHorizonPoints,
		)...)
		if len(polygon) >= 4 {
			polygons = append(polygons, polygon)
		}
	}
	for _, interior := range footprint.InteriorPolygons {
		polygons = append(polygons, append([]OccultationPathPoint(nil), interior...))
	}
	if len(polygons) == 0 {
		return footprint.Polygons
	}
	return polygons
}

// occultationStationOracleProbeOffsets 返回由初始步长倍增得到的探测偏移序列，
// 末项恰好等于搜索半径上限；等比步长本身不会落在上限上。
func occultationStationOracleProbeOffsets() []float64 {
	offsets := make([]float64, 0, 8)
	for step := occultationStationOracleInitialStepKM; step < occultationStationOracleMaximumOffsetKM; step *= 2 {
		offsets = append(offsets, step)
	}
	return append(offsets, occultationStationOracleMaximumOffsetKM)
}

// occultationStationOracleResidualBound 返回某括号宽度下仍可接受的接触残差上限：
// 容差的 10 倍与斜率上限乘括号宽度取较大者，避免浮点噪声否定已收敛的解。
func occultationStationOracleResidualBound(bracketWidthKM float64) float64 {
	return math.Max(
		10*occultationStationOracleResidualToleranceDeg,
		occultationStationOracleResidualSlopeDegPerKM*math.Abs(bracketWidthKM),
	)
}

func occultationStationFindBracket(
	seedResidual float64,
	evaluate func(float64) (float64, float64, float64, bool),
) (float64, float64, bool) {
	if !finite(seedResidual) {
		return 0, 0, false
	}
	var brackets [][2]float64
	for _, sign := range []float64{1, -1} {
		previousOffset := 0.0
		previousResidual := seedResidual
		for _, step := range occultationStationOracleProbeOffsets() {
			currentOffset := sign * step
			currentResidual, _, _, ok := evaluate(currentOffset)
			if !ok {
				continue
			}
			if previousResidual*currentResidual <= 0 {
				brackets = append(brackets, [2]float64{previousOffset, currentOffset})
				break
			}
			previousOffset, previousResidual = currentOffset, currentResidual
		}
	}
	if len(brackets) == 0 {
		return 0, 0, false
	}
	best := brackets[0]
	for _, candidate := range brackets[1:] {
		bestDistance := math.Abs(best[0] + best[1])
		candidateDistance := math.Abs(candidate[0] + candidate[1])
		if candidateDistance < bestDistance {
			best = candidate
		}
	}
	return best[0], best[1], true
}

func occultationStationOracleSampleAt(
	tt float64,
	fixed occultationPathVector,
	offsetKM, seedResidual, residual, altitude float64,
	frame occultationPathFrame,
	location *time.Location,
) occultationStationBoundarySample {
	rotation := occultationPathEarthRotationAt(tt)
	inertial := occultationStationInverseEarthRotation(fixed, rotation)
	longitude, latitude := occultationPathGeodeticWithSidereal(inertial, ApparentSiderealTime(TD2UT(tt, false))*15)
	return occultationStationBoundarySample{
		point: OccultationPathPoint{
			Time:         occultationTTToLocation(tt, location),
			Longitude:    longitude,
			Latitude:     latitude,
			MoonAltitude: altitude,
		},
		contactResidualDeg: residual,
		horizonResidualDeg: altitude,
		offsetKM:           offsetKM,
		seedResidualDeg:    seedResidual,
		valid:              finite(residual) && finite(altitude) && occultationPathNorm(frame.moon) > 0,
	}
}

func occultationStationCrossTrackAt(
	tt float64,
	frame occultationPathFrame,
	frameAt occultationPathFrameFunc,
	seedFixed occultationPathVector,
) (occultationPathVector, bool) {
	before, beforeOK := frameAt(tt - occultationPathVelocityStepDays)
	after, afterOK := frameAt(tt + occultationPathVelocityStepDays)
	if !beforeOK || !afterOK {
		return occultationPathVector{}, false
	}
	vx := after.moonProjectionX() - before.moonProjectionX()
	vy := after.moonProjectionY() - before.moonProjectionY()
	if math.Hypot(vx, vy) <= 1e-12 {
		return occultationPathVector{}, false
	}
	planeCross := occultationPathUnit(occultationPathAdd(
		occultationPathScale(frame.first, -vy/math.Hypot(vx, vy)),
		occultationPathScale(frame.second, vx/math.Hypot(vx, vy)),
	))
	trackReference, trackOK := occultationPathTrackReference(frame)
	if !trackOK {
		return occultationPathVector{}, false
	}
	trackFixed := occultationPathSub(
		occultationPathEarthFixedVector(tt+occultationPathVelocityStepDays, trackReference),
		occultationPathEarthFixedVector(tt-occultationPathVelocityStepDays, trackReference),
	)
	polarRatioSquared := occultationPathEarthPolarRatio * occultationPathEarthPolarRatio
	normal := occultationPathUnit(occultationPathVector{
		x: seedFixed.x,
		y: seedFixed.y,
		z: seedFixed.z / polarRatioSquared,
	})
	trackFixed = occultationPathSub(trackFixed, occultationPathScale(normal, occultationPathDot(trackFixed, normal)))
	if occultationPathNorm(trackFixed) <= 1e-12 {
		trackFixed = occultationPathEarthFixedVectorWithRotation(planeCross, occultationPathEarthRotationAt(tt))
		trackFixed = occultationPathSub(trackFixed, occultationPathScale(normal, occultationPathDot(trackFixed, normal)))
	}
	trackFixed = occultationPathUnit(trackFixed)
	cross := occultationPathUnit(occultationPathCross(normal, trackFixed))
	if occultationPathNorm(cross) <= 1e-12 {
		return occultationPathVector{}, false
	}
	return cross, true
}

func occultationStationSurfaceVector(longitude, latitude float64) occultationPathVector {
	latitudeRad := latitude * rad
	longitudeRad := longitude * rad
	polarRatio := occultationPathEarthPolarRatio
	u := math.Atan(polarRatio * math.Tan(latitudeRad))
	sinU, cosU := math.Sincos(u)
	sinLongitude, cosLongitude := math.Sincos(longitudeRad)
	return occultationPathScale(occultationPathVector{
		x: cosU * cosLongitude,
		y: cosU * sinLongitude,
		z: polarRatio * sinU,
	}, occultationPathEarthEquatorialRadiusKM)
}

func occultationStationOffsetSurfaceVector(
	seedFixed, tangent occultationPathVector, offsetKM float64,
) occultationPathVector {
	point := occultationPathAdd(seedFixed, occultationPathScale(tangent, offsetKM))
	polarRatioSquared := occultationPathEarthPolarRatio * occultationPathEarthPolarRatio
	metric := math.Sqrt(point.x*point.x + point.y*point.y + point.z*point.z/polarRatioSquared)
	if metric <= 0 || !finite(metric) {
		return seedFixed
	}
	return occultationPathScale(point, occultationPathEarthEquatorialRadiusKM/metric)
}

func occultationStationGeodetic(fixed occultationPathVector) (float64, float64) {
	return normalizeLongitude(math.Atan2(fixed.y, fixed.x) / rad), occultationPathGeodeticLatitude(fixed)
}

func occultationStationInverseEarthRotation(
	fixed occultationPathVector,
	rotation occultationPathEarthRotation,
) occultationPathVector {
	return occultationPathVector{
		x: rotation.cosine*fixed.x - rotation.sine*fixed.y,
		y: rotation.sine*fixed.x + rotation.cosine*fixed.y,
		z: fixed.z,
	}
}
