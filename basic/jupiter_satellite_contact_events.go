package basic

import "math"

const (
	jupiterGalileanContactBracketStepDays = 2.0 / 1440.0
	jupiterGalileanContactBracketSpanDays = 10.0 / 24.0
)

// JupiterGalileanPhenomenonContactPhase 接触阶段 / contact phase.
type JupiterGalileanPhenomenonContactPhase string

const (
	// JupiterGalileanDisappearanceContact 初亏/初入接触阶段 / disappearance ingress contact.
	JupiterGalileanDisappearanceContact JupiterGalileanPhenomenonContactPhase = "disappearance"
	// JupiterGalileanReappearanceContact 复圆/复出接触阶段 / reappearance egress contact.
	JupiterGalileanReappearanceContact JupiterGalileanPhenomenonContactPhase = "reappearance"
)

// JupiterGalileanPhenomenonContact 伽利略卫星接触窗口 / Galilean-satellite contact window.
//
// Start/End 表示有限圆盘或有限影斑开始/结束接触的时刻；ModelCrossing 表示这套连续接触模型下，
// 零半径参考点穿越边界的时刻。
// Start/End mark the beginning/end of the finite-disk or finite-shadow contact interval.
// ModelCrossing is the zero-radius boundary crossing in this continuous contact model.
type JupiterGalileanPhenomenonContact struct {
	Valid bool
	Phase JupiterGalileanPhenomenonContactPhase

	Start         float64
	ModelCrossing float64
	End           float64
}

// JupiterGalileanPhenomenonContactEvent IMCCE 风格的 D/F 接触事件 / IMCCE-style D/F contact event.
//
// 与 `JupiterGalileanPhenomenonEvent` 不同，这里返回的是有限圆盘/有限影斑的初亏与复圆接触窗口；
// 现有整场事件 API 返回的则是零半径几何模型处于 active 状态的整段区间。
// 对 `shadow_transit`，这里按 IMCCE 的影凌语义处理：先用半影/本影边界求出部分相持续时间，
// 再把这段持续时间中心放在旧 `shadow_transit` API 的影轴过盘时刻上。
// Unlike `JupiterGalileanPhenomenonEvent`, this returns the finite-disk / finite-shadow D/F contact windows.
// The existing full-event API returns the whole active interval of the zero-radius geometric model.
// For `shadow_transit`, the partial-phase duration comes from penumbra/umbra boundaries,
// while the reported D/F time is centered on the shadow-axis limb crossing from the existing full-event model.
type JupiterGalileanPhenomenonContactEvent struct {
	Valid     bool
	Satellite int
	Type      JupiterGalileanPhenomenonType

	Disappearance JupiterGalileanPhenomenonContact
	Greatest      float64
	Reappearance  JupiterGalileanPhenomenonContact

	GreatestPhenomenon JupiterGalileanPhenomenon
}

type jupiterGalileanContactGeometry struct {
	signedDistance  float64
	effectiveRadius float64
}

// LastJupiterGalileanPhenomenonContactEvent 上一次 IMCCE 风格接触事件 / previous IMCCE-style contact event.
func LastJupiterGalileanPhenomenonContactEvent(jd float64, satellite int, phenomenonType JupiterGalileanPhenomenonType) JupiterGalileanPhenomenonContactEvent {
	return jupiterGalileanPhenomenonContactEventFromEvent(
		LastJupiterGalileanPhenomenonEvent(jd, satellite, phenomenonType),
	)
}

// NextJupiterGalileanPhenomenonContactEvent 下一次 IMCCE 风格接触事件 / next IMCCE-style contact event.
func NextJupiterGalileanPhenomenonContactEvent(jd float64, satellite int, phenomenonType JupiterGalileanPhenomenonType) JupiterGalileanPhenomenonContactEvent {
	return jupiterGalileanPhenomenonContactEventFromEvent(
		NextJupiterGalileanPhenomenonEvent(jd, satellite, phenomenonType),
	)
}

// ClosestJupiterGalileanPhenomenonContactEvent 最近一次 IMCCE 风格接触事件 / closest IMCCE-style contact event.
func ClosestJupiterGalileanPhenomenonContactEvent(jd float64, satellite int, phenomenonType JupiterGalileanPhenomenonType) JupiterGalileanPhenomenonContactEvent {
	return jupiterGalileanPhenomenonContactEventFromEvent(
		ClosestJupiterGalileanPhenomenonEvent(jd, satellite, phenomenonType),
	)
}

func jupiterGalileanPhenomenonContactEventFromEvent(event JupiterGalileanPhenomenonEvent) JupiterGalileanPhenomenonContactEvent {
	if !event.Valid {
		return invalidJupiterGalileanPhenomenonContactEvent()
	}
	var (
		disappearance JupiterGalileanPhenomenonContact
		reappearance  JupiterGalileanPhenomenonContact
		ok            bool
	)
	if event.Type == JupiterGalileanShadowTransit {
		disappearance, reappearance, ok = refineJupiterGalileanShadowContactPair(event)
	} else if event.Type == JupiterGalileanEclipse {
		disappearance, reappearance, ok = refineJupiterGalileanEclipseContactPair(event)
	} else {
		disappearance, reappearance, ok = refineJupiterGalileanContactPair(event.Greatest, event.Satellite, event.Type)
	}
	if !ok {
		return invalidJupiterGalileanPhenomenonContactEvent()
	}
	return JupiterGalileanPhenomenonContactEvent{
		Valid:              true,
		Satellite:          event.Satellite,
		Type:               event.Type,
		Disappearance:      disappearance,
		Greatest:           event.Greatest,
		Reappearance:       reappearance,
		GreatestPhenomenon: event.GreatestPhenomenon,
	}
}

func refineJupiterGalileanContactPair(
	greatestJD float64,
	satellite int,
	phenomenonType JupiterGalileanPhenomenonType,
) (JupiterGalileanPhenomenonContact, JupiterGalileanPhenomenonContact, bool) {
	signedDistance := func(jd float64) float64 {
		geometry, ok := jupiterGalileanContactGeometryAt(jd, satellite, phenomenonType)
		if !ok {
			return math.NaN()
		}
		return geometry.signedDistance
	}
	disappearanceStartTarget := func(jd float64) float64 {
		geometry, ok := jupiterGalileanContactGeometryAt(jd, satellite, phenomenonType)
		if !ok {
			return math.NaN()
		}
		return geometry.signedDistance - geometry.effectiveRadius
	}
	insideTarget := func(jd float64) float64 {
		geometry, ok := jupiterGalileanContactGeometryAt(jd, satellite, phenomenonType)
		if !ok {
			return math.NaN()
		}
		return geometry.signedDistance + geometry.effectiveRadius
	}

	disappearanceModel, ok := refineJupiterGalileanContactRoot(greatestJD, -1, signedDistance)
	if !ok {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	reappearanceModel, ok := refineJupiterGalileanContactRoot(greatestJD, 1, signedDistance)
	if !ok || reappearanceModel <= disappearanceModel {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}

	disappearanceStart, ok := refineJupiterGalileanContactRoot(disappearanceModel, -1, disappearanceStartTarget)
	if !ok {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	disappearanceEnd, ok := refineJupiterGalileanContactRoot(disappearanceModel, 1, insideTarget)
	if !ok || disappearanceEnd <= disappearanceStart {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	reappearanceStart, ok := refineJupiterGalileanContactRoot(reappearanceModel, -1, insideTarget)
	if !ok {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	reappearanceEnd, ok := refineJupiterGalileanContactRoot(reappearanceModel, 1, disappearanceStartTarget)
	if !ok || reappearanceEnd <= reappearanceStart {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}

	return JupiterGalileanPhenomenonContact{
			Valid:         true,
			Phase:         JupiterGalileanDisappearanceContact,
			Start:         disappearanceStart,
			ModelCrossing: disappearanceModel,
			End:           disappearanceEnd,
		}, JupiterGalileanPhenomenonContact{
			Valid:         true,
			Phase:         JupiterGalileanReappearanceContact,
			Start:         reappearanceStart,
			ModelCrossing: reappearanceModel,
			End:           reappearanceEnd,
		}, true
}

func refineJupiterGalileanEclipseContactPair(
	event JupiterGalileanPhenomenonEvent,
) (JupiterGalileanPhenomenonContact, JupiterGalileanPhenomenonContact, bool) {
	satelliteRadius := jupiterGalileanSatelliteRadiusJupiterRadii(event.Satellite)
	if !isFinite(satelliteRadius) || satelliteRadius <= 0 {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	// IMCCE 的 EC.D/EC.F 更接近“目视消失/重现”而不是纯几何接触：
	// D 相用半影入段近似，F 相用本影/半影出段的中点近似。
	penumbraOuterTarget := func(jd float64) float64 {
		signedDistance, ok := jupiterGalileanEclipseSignedDistanceAt(jd, event.Satellite, true)
		if !ok {
			return math.NaN()
		}
		return signedDistance - satelliteRadius
	}
	penumbraInnerTarget := func(jd float64) float64 {
		signedDistance, ok := jupiterGalileanEclipseSignedDistanceAt(jd, event.Satellite, true)
		if !ok {
			return math.NaN()
		}
		return signedDistance + satelliteRadius
	}
	umbraInnerTarget := func(jd float64) float64 {
		signedDistance, ok := jupiterGalileanEclipseSignedDistanceAt(jd, event.Satellite, false)
		if !ok {
			return math.NaN()
		}
		return signedDistance + satelliteRadius
	}
	umbraOuterTarget := func(jd float64) float64 {
		signedDistance, ok := jupiterGalileanEclipseSignedDistanceAt(jd, event.Satellite, false)
		if !ok {
			return math.NaN()
		}
		return signedDistance - satelliteRadius
	}

	disappearanceStart, ok := refineJupiterGalileanContactRoot(event.Start, -1, penumbraOuterTarget)
	if !ok {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	disappearanceEnd, ok := refineJupiterGalileanContactRoot(event.Start, 1, penumbraInnerTarget)
	if !ok || disappearanceEnd <= disappearanceStart {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	reappearanceUmbraStart, ok := refineJupiterGalileanContactRoot(event.End, -1, umbraInnerTarget)
	if !ok {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	reappearancePenumbraStart, ok := refineJupiterGalileanContactRoot(event.End, -1, penumbraInnerTarget)
	if !ok {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	reappearanceStart := (reappearanceUmbraStart + reappearancePenumbraStart) / 2
	reappearanceUmbraEnd, ok := refineJupiterGalileanContactRoot(event.End, 1, umbraOuterTarget)
	if !ok {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	reappearancePenumbraEnd, ok := refineJupiterGalileanContactRoot(event.End, 1, penumbraOuterTarget)
	if !ok {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	reappearanceEnd := (reappearanceUmbraEnd + reappearancePenumbraEnd) / 2
	if !ok || reappearanceEnd <= reappearanceStart {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}

	return JupiterGalileanPhenomenonContact{
			Valid:         true,
			Phase:         JupiterGalileanDisappearanceContact,
			Start:         disappearanceStart,
			ModelCrossing: event.Start,
			End:           disappearanceEnd,
		}, JupiterGalileanPhenomenonContact{
			Valid:         true,
			Phase:         JupiterGalileanReappearanceContact,
			Start:         reappearanceStart,
			ModelCrossing: event.End,
			End:           reappearanceEnd,
		}, true
}

func refineJupiterGalileanShadowContactPair(
	event JupiterGalileanPhenomenonEvent,
) (JupiterGalileanPhenomenonContact, JupiterGalileanPhenomenonContact, bool) {
	penumbraMetric := func(jd float64) float64 {
		value, ok := jupiterGalileanShadowLimbMetricAt(jd, event.Satellite, true)
		if !ok {
			return math.NaN()
		}
		return value
	}
	umbraMetric := func(jd float64) float64 {
		value, ok := jupiterGalileanShadowLimbMetricAt(jd, event.Satellite, false)
		if !ok {
			return math.NaN()
		}
		return value
	}

	penumbraSeedStartJD, penumbraSeedStartValue, ok := findJupiterGalileanNegativeMetricSeed(event.Start, penumbraMetric)
	if !ok {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	disappearanceStart, ok := refineJupiterGalileanNegativeWindowRoot(penumbraSeedStartJD, penumbraSeedStartValue, -1, penumbraMetric)
	if !ok {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	umbraSeedStartJD, umbraSeedStartValue, ok := findJupiterGalileanNegativeMetricSeed(event.Start, umbraMetric)
	if !ok {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	disappearanceUmbraIn, ok := refineJupiterGalileanNegativeWindowRoot(umbraSeedStartJD, umbraSeedStartValue, 1, umbraMetric)
	if !ok || disappearanceUmbraIn <= disappearanceStart {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}

	umbraSeedEndJD, umbraSeedEndValue, ok := findJupiterGalileanNegativeMetricSeed(event.End, umbraMetric)
	if !ok {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	reappearanceUmbraOut, ok := refineJupiterGalileanNegativeWindowRoot(umbraSeedEndJD, umbraSeedEndValue, -1, umbraMetric)
	if !ok {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	penumbraSeedEndJD, penumbraSeedEndValue, ok := findJupiterGalileanNegativeMetricSeed(event.End, penumbraMetric)
	if !ok {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	reappearancePenumbraOut, ok := refineJupiterGalileanNegativeWindowRoot(penumbraSeedEndJD, penumbraSeedEndValue, 1, penumbraMetric)
	if !ok || reappearancePenumbraOut <= reappearanceUmbraOut {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	disappearanceDuration := disappearanceUmbraIn - disappearanceStart
	reappearanceDuration := reappearancePenumbraOut - reappearanceUmbraOut
	if disappearanceDuration <= 0 || reappearanceDuration <= 0 {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	disappearanceCenteredStart := event.Start - disappearanceDuration/2
	disappearanceCenteredEnd := event.Start + disappearanceDuration/2
	reappearanceCenteredStart := event.End - reappearanceDuration/2
	reappearanceCenteredEnd := event.End + reappearanceDuration/2
	return JupiterGalileanPhenomenonContact{
			Valid:         true,
			Phase:         JupiterGalileanDisappearanceContact,
			Start:         disappearanceCenteredStart,
			ModelCrossing: event.Start,
			End:           disappearanceCenteredEnd,
		}, JupiterGalileanPhenomenonContact{
			Valid:         true,
			Phase:         JupiterGalileanReappearanceContact,
			Start:         reappearanceCenteredStart,
			ModelCrossing: event.End,
			End:           reappearanceCenteredEnd,
		}, true
}

func refineJupiterGalileanNegativeMetricWindow(
	seedJD float64,
	metric func(jd float64) float64,
) (float64, float64, bool) {
	activeJD, activeValue, ok := findJupiterGalileanNegativeMetricSeed(seedJD, metric)
	if !ok {
		return math.NaN(), math.NaN(), false
	}
	start, ok := refineJupiterGalileanNegativeWindowRoot(activeJD, activeValue, -1, metric)
	if !ok {
		return math.NaN(), math.NaN(), false
	}
	end, ok := refineJupiterGalileanNegativeWindowRoot(activeJD, activeValue, 1, metric)
	if !ok {
		return math.NaN(), math.NaN(), false
	}
	return start, end, true
}

func findJupiterGalileanNegativeMetricSeed(
	seedJD float64,
	metric func(jd float64) float64,
) (float64, float64, bool) {
	value := metric(seedJD)
	if isFinite(value) && value < 0 {
		return seedJD, value, true
	}
	step := jupiterGalileanContactBracketStepDays
	maxSteps := int(math.Ceil(jupiterGalileanContactBracketSpanDays / step))
	for i := 1; i <= maxSteps; i++ {
		for _, direction := range []float64{-1, 1} {
			candidateJD := seedJD + direction*step*float64(i)
			candidateValue := metric(candidateJD)
			if isFinite(candidateValue) && candidateValue < 0 {
				return candidateJD, candidateValue, true
			}
		}
	}
	return math.NaN(), math.NaN(), false
}

func refineJupiterGalileanNegativeWindowRoot(
	activeJD, activeValue float64,
	direction int,
	metric func(jd float64) float64,
) (float64, bool) {
	currentJD := activeJD
	currentValue := activeValue
	step := jupiterGalileanContactBracketStepDays
	maxSteps := int(math.Ceil(jupiterGalileanContactBracketSpanDays / step))
	for i := 1; i <= maxSteps; i++ {
		candidateJD := currentJD + float64(direction)*step
		candidateValue := metric(candidateJD)
		if !isFinite(candidateValue) {
			currentJD = candidateJD
			currentValue = math.Inf(1)
			continue
		}
		if candidateValue >= 0 {
			return bisectJupiterGalileanContactRoot(currentJD, currentValue, candidateJD, candidateValue, metric)
		}
		currentJD = candidateJD
		currentValue = candidateValue
	}
	return math.NaN(), false
}

func refineJupiterGalileanContactRoot(
	modelCrossingJD float64,
	direction int,
	target func(jd float64) float64,
) (float64, bool) {
	if direction != -1 && direction != 1 {
		return math.NaN(), false
	}
	modelValue := target(modelCrossingJD)
	if !isFinite(modelValue) {
		return math.NaN(), false
	}
	step := jupiterGalileanContactBracketStepDays
	maxSteps := int(math.Ceil(jupiterGalileanContactBracketSpanDays / step))
	nearJD := modelCrossingJD
	nearValue := modelValue
	for i := 1; i <= maxSteps; i++ {
		farJD := modelCrossingJD + float64(direction)*step*float64(i)
		farValue := target(farJD)
		if !isFinite(farValue) {
			continue
		}
		if nearValue == 0 {
			return nearJD, true
		}
		if farValue == 0 {
			return farJD, true
		}
		if nearValue*farValue < 0 {
			return bisectJupiterGalileanContactRoot(nearJD, nearValue, farJD, farValue, target)
		}
		nearJD = farJD
		nearValue = farValue
	}
	return math.NaN(), false
}

func bisectJupiterGalileanContactRoot(
	jd1, value1, jd2, value2 float64,
	target func(jd float64) float64,
) (float64, bool) {
	leftJD := jd1
	rightJD := jd2
	leftValue := value1
	rightValue := value2
	if rightJD < leftJD {
		leftJD, rightJD = rightJD, leftJD
		leftValue, rightValue = rightValue, leftValue
	}
	if !isFinite(leftValue) || !isFinite(rightValue) || leftValue*rightValue > 0 {
		return math.NaN(), false
	}
	for i := 0; i < 80 && rightJD-leftJD > jupiterGalileanEventEpsilonDays; i++ {
		midJD := (leftJD + rightJD) / 2
		midValue := target(midJD)
		if !isFinite(midValue) {
			return math.NaN(), false
		}
		if midValue == 0 {
			return midJD, true
		}
		if leftValue*midValue <= 0 {
			rightJD = midJD
			rightValue = midValue
		} else {
			leftJD = midJD
			leftValue = midValue
		}
	}
	return (leftJD + rightJD) / 2, true
}

func jupiterGalileanContactGeometryAt(
	jd float64,
	satellite int,
	phenomenonType JupiterGalileanPhenomenonType,
) (jupiterGalileanContactGeometry, bool) {
	if !isFinite(jd) || satellite < 1 || satellite > 4 || !isValidJupiterGalileanPhenomenonType(phenomenonType) {
		return jupiterGalileanContactGeometry{}, false
	}
	evaluationJD := TD2UT(jd, true)
	context := newJupiterGalileanObservationContext(evaluationJD)
	if context.jupiterDistance == 0 {
		return jupiterGalileanContactGeometry{}, false
	}

	index := satellite - 1
	observation := context.observationForSatellite(index)
	stateVector := Vector3{observation.State.X, observation.State.Y, observation.State.Z}
	satelliteRadius := jupiterGalileanSatelliteRadiusJupiterRadii(satellite)
	if !isFinite(satelliteRadius) || satelliteRadius <= 0 {
		return jupiterGalileanContactGeometry{}, false
	}

	switch phenomenonType {
	case JupiterGalileanTransit, JupiterGalileanOccultation:
		return jupiterGalileanContactGeometry{
			signedDistance:  ellipseSignedDistance(observation.OffsetXJupiterRadii, observation.OffsetYJupiterRadii, 1, context.earthMinorRadius),
			effectiveRadius: satelliteRadius,
		}, true
	case JupiterGalileanEclipse:
		xSunAU := vectorDot(stateVector, context.sunEast)
		ySunAU := vectorDot(stateVector, context.sunNorth)
		zSunAU := vectorDot(stateVector, context.sunLineOfSight)
		umbraScale := jupiterUmbraScale(zSunAU, context.sunDistanceAU)
		if zSunAU <= 0 || umbraScale <= 0 {
			return jupiterGalileanContactGeometry{}, false
		}
		radiusAU := jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM
		return jupiterGalileanContactGeometry{
			signedDistance:  ellipseSignedDistance(xSunAU/radiusAU, ySunAU/radiusAU, umbraScale, context.sunMinorRadius*umbraScale),
			effectiveRadius: satelliteRadius,
		}, true
	case JupiterGalileanShadowTransit:
		radiusAU := jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM
		axisDenominator := vectorDot(context.sunLineOfSight, context.lineOfSight)
		if math.Abs(axisDenominator) < 1e-12 {
			return jupiterGalileanContactGeometry{}, false
		}
		axisDistanceAU := -vectorDot(stateVector, context.lineOfSight) / axisDenominator
		if axisDistanceAU <= 0 {
			return jupiterGalileanContactGeometry{}, false
		}
		axisPoint := Vector3{
			stateVector[0] + axisDistanceAU*context.sunLineOfSight[0],
			stateVector[1] + axisDistanceAU*context.sunLineOfSight[1],
			stateVector[2] + axisDistanceAU*context.sunLineOfSight[2],
		}
		xAU := vectorDot(axisPoint, context.east)
		yAU := vectorDot(axisPoint, context.north)
		return jupiterGalileanContactGeometry{
			signedDistance:  ellipseSignedDistance(xAU/radiusAU, yAU/radiusAU, 1, context.earthMinorRadius),
			effectiveRadius: jupiterGalileanPenumbraRadiusJupiterRadii(satellite, axisDistanceAU, context.sunDistanceAU),
		}, true
	default:
		return jupiterGalileanContactGeometry{}, false
	}
}

func jupiterGalileanEclipseSignedDistanceAt(jd float64, satellite int, penumbra bool) (float64, bool) {
	if !isFinite(jd) || satellite < 1 || satellite > 4 {
		return math.NaN(), false
	}
	evaluationJD := TD2UT(jd, true)
	context := newJupiterGalileanObservationContext(evaluationJD)
	if context.jupiterDistance == 0 {
		return math.NaN(), false
	}
	state := context.observationForSatellite(satellite - 1).State
	stateVector := Vector3{state.X, state.Y, state.Z}
	xSunAU := vectorDot(stateVector, context.sunEast)
	ySunAU := vectorDot(stateVector, context.sunNorth)
	zSunAU := vectorDot(stateVector, context.sunLineOfSight)
	if zSunAU <= 0 {
		return math.NaN(), false
	}
	scale := jupiterUmbraScale(zSunAU, context.sunDistanceAU)
	if penumbra {
		scale = jupiterPenumbraScale(zSunAU, context.sunDistanceAU)
	}
	if scale <= 0 {
		return math.NaN(), false
	}
	radiusAU := jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM
	return ellipseSignedDistance(xSunAU/radiusAU, ySunAU/radiusAU, scale, context.sunMinorRadius*scale), true
}

func jupiterGalileanSatelliteRadiusJupiterRadii(satellite int) float64 {
	switch satellite {
	case 1:
		return 1821.6 / jupiterGalileanEquatorialRadiusKM
	case 2:
		return 1560.8 / jupiterGalileanEquatorialRadiusKM
	case 3:
		return 2634.1 / jupiterGalileanEquatorialRadiusKM
	case 4:
		return 2410.3 / jupiterGalileanEquatorialRadiusKM
	default:
		return math.NaN()
	}
}

func jupiterPenumbraScale(distanceBehindAU, sunDistanceAU float64) float64 {
	if distanceBehindAU <= 0 || sunDistanceAU <= 0 {
		return 0
	}
	jupiterRadiusAU := jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM
	return 1 + distanceBehindAU*(solarRadiusAU+jupiterRadiusAU)/(sunDistanceAU*jupiterRadiusAU)
}

func jupiterGalileanUmbraRadiusJupiterRadii(satellite int, pathLengthAU, sunDistanceAU float64) float64 {
	if pathLengthAU <= 0 || sunDistanceAU <= 0 {
		return math.NaN()
	}
	satelliteRadiusAU := jupiterGalileanSatelliteRadiusJupiterRadii(satellite) * jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM
	umbraRadiusAU := satelliteRadiusAU - pathLengthAU*(solarRadiusAU-satelliteRadiusAU)/sunDistanceAU
	if umbraRadiusAU <= 0 {
		return math.NaN()
	}
	return umbraRadiusAU / (jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM)
}

func jupiterGalileanPenumbraRadiusJupiterRadii(satellite int, pathLengthAU, sunDistanceAU float64) float64 {
	if pathLengthAU <= 0 || sunDistanceAU <= 0 {
		return math.NaN()
	}
	satelliteRadiusAU := jupiterGalileanSatelliteRadiusJupiterRadii(satellite) * jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM
	penumbraRadiusAU := satelliteRadiusAU + pathLengthAU*(solarRadiusAU+satelliteRadiusAU)/sunDistanceAU
	if penumbraRadiusAU <= 0 {
		return math.NaN()
	}
	return penumbraRadiusAU / (jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM)
}

func jupiterGalileanShadowLimbMetricAt(jd float64, satellite int, penumbra bool) (float64, bool) {
	if !isFinite(jd) || satellite < 1 || satellite > 4 {
		return math.NaN(), false
	}
	evaluationJD := TD2UT(jd, true)
	context := newJupiterGalileanObservationContext(evaluationJD)
	if context.jupiterDistance == 0 {
		return math.NaN(), false
	}
	state := context.observationForSatellite(satellite - 1).State
	stateVector := Vector3{state.X, state.Y, state.Z}
	satelliteBody := context.toBodyCoordinates(stateVector)
	axisBody := normalizeVector(context.toBodyCoordinates(context.sunLineOfSight))
	if vectorMagnitude(axisBody) == 0 {
		return math.NaN(), false
	}
	radiusAU := jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM
	satelliteRadiusAU := jupiterGalileanSatelliteRadiusJupiterRadii(satellite) * radiusAU
	limbU, limbV, ok := jupiterGalileanVisibleLimbBasis(context)
	if !ok {
		return math.NaN(), false
	}
	metricAtAngle := func(angle float64) float64 {
		limbPointBody := jupiterGalileanVisibleLimbPoint(angle, limbU, limbV, jupiterPolarRadiusRatio())
		limbPointAU := Vector3{
			limbPointBody[0] * radiusAU,
			limbPointBody[1] * radiusAU,
			limbPointBody[2] * radiusAU,
		}
		return jupiterGalileanShadowConeMetricForPoint(limbPointAU, satelliteBody, axisBody, satelliteRadiusAU, context.sunDistanceAU, penumbra)
	}
	return minimizeJupiterGalileanPeriodicMetric(metricAtAngle)
}

func jupiterGalileanVisibleLimbBasis(context jupiterGalileanObservationContext) (Vector3, Vector3, bool) {
	polar := jupiterPolarRadiusRatio()
	earthBody := context.toBodyCoordinates(context.earthDirection)
	planeNormal := Vector3{earthBody[0], earthBody[1], earthBody[2] / polar}
	planeNormal = normalizeVector(planeNormal)
	if vectorMagnitude(planeNormal) == 0 {
		return Vector3{}, Vector3{}, false
	}
	reference := Vector3{0, 0, 1}
	if math.Abs(vectorDot(reference, planeNormal)) > 0.9 {
		reference = Vector3{1, 0, 0}
	}
	u := normalizeVector(pxp(planeNormal, reference))
	if vectorMagnitude(u) == 0 {
		reference = Vector3{0, 1, 0}
		u = normalizeVector(pxp(planeNormal, reference))
		if vectorMagnitude(u) == 0 {
			return Vector3{}, Vector3{}, false
		}
	}
	v := normalizeVector(pxp(planeNormal, u))
	if vectorMagnitude(v) == 0 {
		return Vector3{}, Vector3{}, false
	}
	return u, v, true
}

func jupiterGalileanVisibleLimbPoint(angle float64, u, v Vector3, polar float64) Vector3 {
	sinAngle := math.Sin(angle)
	cosAngle := math.Cos(angle)
	q := Vector3{
		u[0]*cosAngle + v[0]*sinAngle,
		u[1]*cosAngle + v[1]*sinAngle,
		u[2]*cosAngle + v[2]*sinAngle,
	}
	return Vector3{q[0], q[1], polar * q[2]}
}

func jupiterGalileanShadowConeMetricForPoint(
	pointAU, satelliteAU, axisUnit Vector3,
	satelliteRadiusAU, sunDistanceAU float64,
	penumbra bool,
) float64 {
	vector := Vector3{
		pointAU[0] - satelliteAU[0],
		pointAU[1] - satelliteAU[1],
		pointAU[2] - satelliteAU[2],
	}
	axisDistanceAU := vectorDot(vector, axisUnit)
	if axisDistanceAU <= 0 {
		return math.Inf(1)
	}
	perpendicular := Vector3{
		vector[0] - axisDistanceAU*axisUnit[0],
		vector[1] - axisDistanceAU*axisUnit[1],
		vector[2] - axisDistanceAU*axisUnit[2],
	}
	perpendicularDistanceAU := vectorMagnitude(perpendicular)
	if penumbra {
		penumbraRadiusAU := satelliteRadiusAU + axisDistanceAU*(solarRadiusAU+satelliteRadiusAU)/sunDistanceAU
		return perpendicularDistanceAU - penumbraRadiusAU
	}
	umbraRadiusAU := satelliteRadiusAU - axisDistanceAU*(solarRadiusAU-satelliteRadiusAU)/sunDistanceAU
	return perpendicularDistanceAU - umbraRadiusAU
}

func minimizeJupiterGalileanPeriodicMetric(metric func(angle float64) float64) (float64, bool) {
	const (
		samples = 144
		phi     = 0.6180339887498948482
	)
	step := 2 * math.Pi / float64(samples)
	bestAngle := 0.0
	bestValue := math.Inf(1)
	for i := 0; i < samples; i++ {
		angle := float64(i) * step
		value := metric(angle)
		if value < bestValue {
			bestValue = value
			bestAngle = angle
		}
	}
	if !isFinite(bestValue) {
		return math.NaN(), false
	}
	left := bestAngle - step
	right := bestAngle + step
	x1 := right - phi*(right-left)
	x2 := left + phi*(right-left)
	f1 := metric(x1)
	f2 := metric(x2)
	for i := 0; i < 80; i++ {
		if f1 <= f2 {
			right = x2
			x2 = x1
			f2 = f1
			x1 = right - phi*(right-left)
			f1 = metric(x1)
		} else {
			left = x1
			x1 = x2
			f1 = f2
			x2 = left + phi*(right-left)
			f2 = metric(x2)
		}
	}
	return math.Min(bestValue, metric((left+right)/2)), true
}

func ellipseSignedDistance(x, y, major, minor float64) float64 {
	if major <= 0 || minor <= 0 {
		return math.NaN()
	}
	targetX := math.Abs(x)
	targetY := math.Abs(y)
	if targetX == 0 && targetY == 0 {
		if minor < major {
			return -minor
		}
		return -major
	}
	left := 0.0
	right := math.Pi / 2
	const phi = 0.6180339887498948482
	x1 := right - phi*(right-left)
	x2 := left + phi*(right-left)
	f1 := ellipseDistanceSquaredAtAngle(targetX, targetY, major, minor, x1)
	f2 := ellipseDistanceSquaredAtAngle(targetX, targetY, major, minor, x2)
	for i := 0; i < 80; i++ {
		if f1 <= f2 {
			right = x2
			x2 = x1
			f2 = f1
			x1 = right - phi*(right-left)
			f1 = ellipseDistanceSquaredAtAngle(targetX, targetY, major, minor, x1)
		} else {
			left = x1
			x1 = x2
			f1 = f2
			x2 = left + phi*(right-left)
			f2 = ellipseDistanceSquaredAtAngle(targetX, targetY, major, minor, x2)
		}
	}
	distance := math.Sqrt(ellipseDistanceSquaredAtAngle(targetX, targetY, major, minor, (left+right)/2))
	if ellipseInside(x, y, major, minor) {
		return -distance
	}
	return distance
}

func ellipseDistanceSquaredAtAngle(x, y, major, minor, angle float64) float64 {
	ellipseX := major * math.Cos(angle)
	ellipseY := minor * math.Sin(angle)
	dx := ellipseX - x
	dy := ellipseY - y
	return dx*dx + dy*dy
}

func invalidJupiterGalileanPhenomenonContactEvent() JupiterGalileanPhenomenonContactEvent {
	return JupiterGalileanPhenomenonContactEvent{
		Disappearance: JupiterGalileanPhenomenonContact{
			Start:         math.NaN(),
			ModelCrossing: math.NaN(),
			End:           math.NaN(),
		},
		Greatest: math.NaN(),
		Reappearance: JupiterGalileanPhenomenonContact{
			Start:         math.NaN(),
			ModelCrossing: math.NaN(),
			End:           math.NaN(),
		},
		GreatestPhenomenon: invalidJupiterGalileanPhenomenon(),
	}
}
