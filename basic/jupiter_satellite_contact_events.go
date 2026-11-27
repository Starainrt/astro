package basic

import (
	"math"
)

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
// 常规窗口的 Start/End 来自有限圆盘或有限影斑模型；ModelCrossing 是零半径参考点穿越边界的时刻，未穿越时为 NaN。
// For regular windows, Start/End come from the finite-disk or finite-shadow model. ModelCrossing is the zero-radius boundary crossing, or NaN if absent.
type JupiterGalileanPhenomenonContact struct {
	Valid bool
	Phase JupiterGalileanPhenomenonContactPhase

	// PenumbralFallback 备用半影时段，以事件 Greatest 为界 / penumbral segments split at the event Greatest.
	PenumbralFallback bool

	Start         float64
	ModelCrossing float64
	End           float64
}

// JupiterGalileanPhenomenonContactEvent IMCCE 风格的 D/F 接触事件 / IMCCE-style D/F contact event.
//
// 有限圆盘或有限影斑的接触窗口。影凌的部分相时长来自半影和本影边界，时刻中心在影轴过盘点上。
// Finite-disk or finite-shadow contact windows. Shadow-transit partial duration uses the penumbra and umbra edges, centered on the shadow-axis limb crossing.
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

// LastJupiterGalileanPhenomenonContactEventWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func LastJupiterGalileanPhenomenonContactEventWithInstance(jd float64, satellite int, phenomenonType JupiterGalileanPhenomenonType, inst *Instance) JupiterGalileanPhenomenonContactEvent {
	return searchJupiterGalileanContact(jd, satellite, phenomenonType, -1, inst)
}

// NextJupiterGalileanPhenomenonContactEventWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func NextJupiterGalileanPhenomenonContactEventWithInstance(jd float64, satellite int, phenomenonType JupiterGalileanPhenomenonType, inst *Instance) JupiterGalileanPhenomenonContactEvent {
	return searchJupiterGalileanContact(jd, satellite, phenomenonType, 1, inst)
}

// ClosestJupiterGalileanPhenomenonContactEventWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func ClosestJupiterGalileanPhenomenonContactEventWithInstance(jd float64, satellite int, phenomenonType JupiterGalileanPhenomenonType, inst *Instance) JupiterGalileanPhenomenonContactEvent {
	last := searchJupiterGalileanContact(jd, satellite, phenomenonType, -1, inst)
	if last.Valid && last.Disappearance.Start <= jd && jd <= last.Reappearance.End {
		return last
	}
	next := searchJupiterGalileanContact(jd, satellite, phenomenonType, 1, inst)
	if last.Valid && (!next.Valid || math.Abs(last.Greatest-jd) <= math.Abs(next.Greatest-jd)) {
		return last
	}
	return next
}

func searchJupiterGalileanContact(jd float64, satellite int, kind JupiterGalileanPhenomenonType, direction int, inst *Instance) JupiterGalileanPhenomenonContactEvent {
	if !isFinite(jd) || !isValidJupiterGalileanPhenomenonType(kind) {
		return invalidJupiterGalileanPhenomenonContactEvent()
	}
	searchKind := kind
	if kind == JupiterGalileanEclipse {
		searchKind = jupiterGalileanPenumbraContact
	}
	seed := jd
	if direction < 0 {
		seed += jupiterGalileanContactBracketSpanDays
	}
	for includeCurrent := true; math.Abs(seed-jd) < jupiterGalileanEventSearchSpanDays; includeCurrent = direction < 0 {
		whole, ok := searchJupiterGalileanPhenomenonEvent(seed, satellite, searchKind, direction, includeCurrent, inst)
		if !ok || math.Abs(whole.Greatest-jd) > jupiterGalileanEventSearchSpanDays {
			break
		}
		var contact JupiterGalileanPhenomenonContactEvent
		if searchKind != kind {
			center := findJupiterGalileanPhenomenonEventAround(whole.Greatest, satellite, kind, inst)
			contact = jupiterGalileanPhenomenonContactEventFromEvent(center, inst)
			if !contact.Valid {
				contact = partialJupiterGalileanEclipse(whole, center)
			}
		} else {
			contact = jupiterGalileanPhenomenonContactEventFromEvent(whole, inst)
		}
		started := contact.Disappearance.Start-jd <= jupiterGalileanEventEpsilonDays
		if contact.Valid && (direction < 0 && started || direction > 0 && !started) {
			return contact
		}
		if direction > 0 {
			seed = math.Max(seed, whole.End) + jupiterGalileanEventEpsilonDays
		} else {
			seed = math.Min(seed, whole.Start) - jupiterGalileanEventEpsilonDays
		}
	}
	return invalidJupiterGalileanPhenomenonContactEvent()
}

func partialJupiterGalileanEclipse(whole, center JupiterGalileanPhenomenonEvent) JupiterGalileanPhenomenonContactEvent {
	greatest := whole.Greatest
	state := whole.GreatestPhenomenon
	modelStart, modelEnd := math.NaN(), math.NaN()
	if center.Valid {
		greatest, state = center.Greatest, center.GreatestPhenomenon
		modelStart, modelEnd = center.Start, center.End
	}
	return JupiterGalileanPhenomenonContactEvent{
		Valid: true, Satellite: whole.Satellite, Type: JupiterGalileanEclipse, Greatest: greatest, GreatestPhenomenon: state,
		Disappearance: JupiterGalileanPhenomenonContact{
			Valid:             true,
			Phase:             JupiterGalileanDisappearanceContact,
			PenumbralFallback: true,
			Start:             whole.Start,
			ModelCrossing:     modelStart,
			End:               greatest,
		},
		Reappearance: JupiterGalileanPhenomenonContact{
			Valid:             true,
			Phase:             JupiterGalileanReappearanceContact,
			PenumbralFallback: true,
			Start:             greatest,
			ModelCrossing:     modelEnd,
			End:               whole.End,
		},
	}
}

func jupiterGalileanPhenomenonContactEventFromEvent(event JupiterGalileanPhenomenonEvent, instances ...*Instance) JupiterGalileanPhenomenonContactEvent {
	inst := instanceValue(instances)

	if !event.Valid {
		return invalidJupiterGalileanPhenomenonContactEvent()
	}
	var (
		disappearance JupiterGalileanPhenomenonContact
		reappearance  JupiterGalileanPhenomenonContact
		ok            bool
	)
	if event.Type == JupiterGalileanShadowTransit {
		disappearance, reappearance, ok = refineJupiterGalileanShadowContactPair(event, inst)
	} else if event.Type == JupiterGalileanEclipse {
		disappearance, reappearance, ok = refineJupiterGalileanEclipseContactPair(event, inst)
	} else {
		disappearance, reappearance, ok = refineJupiterGalileanContactPair(event.Greatest, event.Satellite, event.Type, inst)
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
	phenomenonType JupiterGalileanPhenomenonType, instances ...*Instance) (JupiterGalileanPhenomenonContact, JupiterGalileanPhenomenonContact, bool) {
	inst := instanceValue(instances)

	signedDistance := func(jd float64) float64 {
		geometry, ok := jupiterGalileanContactGeometryAt(jd, satellite, phenomenonType, inst)
		if !ok {
			return math.NaN()
		}
		return geometry.signedDistance
	}
	disappearanceStartTarget := func(jd float64) float64 {
		geometry, ok := jupiterGalileanContactGeometryAt(jd, satellite, phenomenonType, inst)
		if !ok {
			return math.NaN()
		}
		return geometry.signedDistance - geometry.effectiveRadius
	}
	insideTarget := func(jd float64) float64 {
		geometry, ok := jupiterGalileanContactGeometryAt(jd, satellite, phenomenonType, inst)
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
	disappearanceEnd, reappearanceStart := greatestJD, greatestJD
	if insideTarget(greatestJD) < 0 {
		disappearanceEnd, ok = refineJupiterGalileanContactRoot(disappearanceModel, 1, insideTarget)
		if !ok || disappearanceEnd <= disappearanceStart {
			return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
		}
		reappearanceStart, ok = refineJupiterGalileanContactRoot(reappearanceModel, -1, insideTarget)
		if !ok {
			return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
		}
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
	event JupiterGalileanPhenomenonEvent, instances ...*Instance) (JupiterGalileanPhenomenonContact, JupiterGalileanPhenomenonContact, bool) {
	inst := instanceValue(instances)

	satelliteRadius := jupiterGalileanSatelliteRadiusJupiterRadii(event.Satellite)
	if !isFinite(satelliteRadius) || satelliteRadius <= 0 {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	// D 相取半影入段，F 相取本影出段与半影出段的中点，对应目视消失和重现。
	penumbraOuterTarget := func(jd float64) float64 {
		signedDistance, ok := jupiterGalileanEclipseSignedDistanceAt(jd, event.Satellite, true, inst)
		if !ok {
			return math.NaN()
		}
		return signedDistance - satelliteRadius
	}
	penumbraInnerTarget := func(jd float64) float64 {
		signedDistance, ok := jupiterGalileanEclipseSignedDistanceAt(jd, event.Satellite, true, inst)
		if !ok {
			return math.NaN()
		}
		return signedDistance + satelliteRadius
	}
	umbraInnerTarget := func(jd float64) float64 {
		signedDistance, ok := jupiterGalileanEclipseSignedDistanceAt(jd, event.Satellite, false, inst)
		if !ok {
			return math.NaN()
		}
		return signedDistance + satelliteRadius
	}
	umbraOuterTarget := func(jd float64) float64 {
		signedDistance, ok := jupiterGalileanEclipseSignedDistanceAt(jd, event.Satellite, false, inst)
		if !ok {
			return math.NaN()
		}
		return signedDistance - satelliteRadius
	}

	disappearanceStart, ok := refineJupiterGalileanContactRoot(event.Start, -1, penumbraOuterTarget)
	if !ok {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	disappearanceEnd, ok := refineJupiterGalileanInnerContactRoot(event.Start, true, penumbraInnerTarget)
	if !ok || disappearanceEnd <= disappearanceStart {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	reappearanceUmbraStart, ok := refineJupiterGalileanContactRoot(event.End, -1, umbraInnerTarget)
	if !ok {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	reappearancePenumbraStart, ok := refineJupiterGalileanInnerContactRoot(event.End, false, penumbraInnerTarget)
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
	event JupiterGalileanPhenomenonEvent, instances ...*Instance) (JupiterGalileanPhenomenonContact, JupiterGalileanPhenomenonContact, bool) {
	contacts, ok := jupiterGalileanShadowContacts(event.Start, event.End, event.Satellite, instanceValue(instances))
	if !ok {
		return JupiterGalileanPhenomenonContact{}, JupiterGalileanPhenomenonContact{}, false
	}
	if !isFinite(contacts[1]) {
		contacts[1], contacts[2] = event.Greatest, event.Greatest
	} else {
		disappearanceHalf := (contacts[1] - contacts[0]) / 2
		reappearanceHalf := (contacts[3] - contacts[2]) / 2
		if event.Start+disappearanceHalf <= event.End-reappearanceHalf {
			contacts = [4]float64{event.Start - disappearanceHalf, event.Start + disappearanceHalf, event.End - reappearanceHalf, event.End + reappearanceHalf}
		}
	}
	return JupiterGalileanPhenomenonContact{
			Valid: true, Phase: JupiterGalileanDisappearanceContact,
			Start: contacts[0], ModelCrossing: event.Start, End: contacts[1],
		}, JupiterGalileanPhenomenonContact{
			Valid: true, Phase: JupiterGalileanReappearanceContact,
			Start: contacts[2], ModelCrossing: event.End, End: contacts[3],
		}, true
}

// JupiterGalileanShadowContactsWithInstance 求影凌四个几何接触，输入与输出均为 UTC 儒略日 / solve four geometric shadow contacts in UTC Julian days.
func JupiterGalileanShadowContactsWithInstance(start, end float64, satellite int, inst *Instance) ([4]float64, bool) {
	if !isFinite(start) || !isFinite(end) || end <= start || satellite < 1 || satellite > 4 {
		return [4]float64{}, false
	}
	contacts, ok := jupiterGalileanShadowContacts(start, end, satellite, inst)
	if !ok || !isFinite(contacts[1]) {
		return [4]float64{}, false
	}
	return contacts, true
}

func jupiterGalileanShadowContacts(start, end float64, satellite int, inst *Instance) ([4]float64, bool) {
	penumbraMetric := func(jd float64) float64 {
		value, ok := jupiterGalileanShadowLimbMetricAt(jd, satellite, true, inst)
		if !ok {
			return math.NaN()
		}
		return value
	}
	umbraMetric := func(jd float64) float64 {
		value, ok := jupiterGalileanShadowLimbMetricAt(jd, satellite, false, inst)
		if !ok {
			return math.NaN()
		}
		return value
	}

	var contacts [4]float64
	for index, item := range [...]struct {
		seed      float64
		direction int
		metric    func(float64) float64
	}{{start, -1, penumbraMetric}, {start, 1, umbraMetric}, {end, -1, umbraMetric}, {end, 1, penumbraMetric}} {
		seed, value, ok := findJupiterGalileanNegativeMetricSeed(item.seed, item.metric)
		if ok {
			contacts[index], ok = refineJupiterGalileanNegativeWindowRoot(seed, value, item.direction, item.metric)
		}
		if ok {
			continue
		}
		if index == 0 || index == 3 {
			return [4]float64{}, false
		}
		contacts[index] = math.NaN()
	}
	if contacts[0] >= contacts[3] {
		return [4]float64{}, false
	}
	if !(contacts[0] < contacts[1] && contacts[1] < contacts[2] && contacts[2] < contacts[3]) {
		contacts[1], contacts[2] = math.NaN(), math.NaN()
	}
	return contacts, true
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

// refineJupiterGalileanInnerContactRoot 半影内切：模型穿越时刻已全入影（值为负）时，入影侧向后、出影侧向前找；尚未全入则反向。
func refineJupiterGalileanInnerContactRoot(modelCrossingJD float64, ingress bool, target func(jd float64) float64) (float64, bool) {
	value := target(modelCrossingJD)
	if !isFinite(value) {
		return math.NaN(), false
	}
	inside := value < 0
	direction := 1
	if ingress == inside {
		direction = -1
	}
	return refineJupiterGalileanContactRoot(modelCrossingJD, direction, target)
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
	phenomenonType JupiterGalileanPhenomenonType, instances ...*Instance) (jupiterGalileanContactGeometry, bool) {
	inst := instanceValue(instances)

	if !isFinite(jd) || satellite < 1 || satellite > 4 || !isValidJupiterGalileanPhenomenonType(phenomenonType) {
		return jupiterGalileanContactGeometry{}, false
	}
	evaluationJDE := UTC2TT(jd)
	context := newJupiterGalileanObservationContext(evaluationJDE, inst)
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

func jupiterGalileanEclipseSignedDistanceAt(jd float64, satellite int, penumbra bool, instances ...*Instance) (float64, bool) {
	inst := instanceValue(instances)

	if !isFinite(jd) || satellite < 1 || satellite > 4 {
		return math.NaN(), false
	}
	evaluationJDE := UTC2TT(jd)
	context := newJupiterGalileanObservationContext(evaluationJDE, inst)
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

// jupiterGalileanSatelliteRadiusJupiterRadii 返回卫星球半径与木星赤道半径之比。
func jupiterGalileanSatelliteRadiusJupiterRadii(satellite int) float64 {
	switch satellite {
	case 1:
		return 1829.4 / jupiterGalileanEquatorialRadiusKM
	case 2:
		return 1562.6 / jupiterGalileanEquatorialRadiusKM
	case 3:
		return 2631.2 / jupiterGalileanEquatorialRadiusKM
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

func jupiterGalileanShadowLimbMetricAt(jd float64, satellite int, penumbra bool, instances ...*Instance) (float64, bool) {
	inst := instanceValue(instances)

	if !isFinite(jd) || satellite < 1 || satellite > 4 {
		return math.NaN(), false
	}
	evaluationJDE := UTC2TT(jd)
	context := newJupiterGalileanObservationContext(evaluationJDE, inst)
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
