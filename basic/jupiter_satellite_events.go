package basic

import "math"

const (
	jupiterGalileanEventSearchSpanDays = 8 * 365.25
	jupiterGalileanEventEpsilonDays    = 1.0 / 86400.0
	jupiterGalileanBoundaryStepDays    = 1.0 / 24.0
)

// JupiterGalileanPhenomenonType 伽利略卫星现象类型 / Galilean-satellite phenomenon type.
type JupiterGalileanPhenomenonType string

const (
	// JupiterGalileanTransit 凌日 / satellite transit across Jupiter.
	JupiterGalileanTransit JupiterGalileanPhenomenonType = "transit"
	// JupiterGalileanOccultation 掩蔽 / occultation behind Jupiter.
	JupiterGalileanOccultation JupiterGalileanPhenomenonType = "occultation"
	// JupiterGalileanEclipse 食 / eclipse in Jupiter's shadow.
	JupiterGalileanEclipse JupiterGalileanPhenomenonType = "eclipse"
	// JupiterGalileanShadowTransit 影凌 / shadow transit across Jupiter.
	JupiterGalileanShadowTransit JupiterGalileanPhenomenonType = "shadow_transit"
)

// JupiterGalileanPhenomenonEvent 伽利略卫星现象整场事件 / full Galilean-satellite phenomenon event.
//
// Start、Greatest、End 都使用 UTC/UT 对应的儒略日。
// Start, Greatest, and End are UTC/UT Julian days.
type JupiterGalileanPhenomenonEvent struct {
	Valid     bool
	Satellite int
	Type      JupiterGalileanPhenomenonType

	Start    float64
	Greatest float64
	End      float64

	GreatestPhenomenon JupiterGalileanPhenomenon
}

type jupiterGalileanMetricSample struct {
	active bool
	metric float64
	// signed 是该现象相对木星视面中心线的有符号偏移（木星半径），每次过零对应一次现象；未定义时为 NaN。
	signed     float64
	phenomenon JupiterGalileanPhenomenon
}

type jupiterGalileanShadowPoint struct {
	hasIntersection bool
	visible         bool
	pathLengthAU    float64
	xAU             float64
	yAU             float64
	xJupiterRadii   float64
	yJupiterRadii   float64
}

// LastJupiterGalileanPhenomenonEvent 上一次伽利略卫星现象 / previous Galilean-satellite event.
//
// jd 与返回时刻都是民用儒略日；需要瞬时状态时先用 UTC2TT(jd) 换成 TT。
func LastJupiterGalileanPhenomenonEvent(jd float64, satellite int, phenomenonType JupiterGalileanPhenomenonType) JupiterGalileanPhenomenonEvent {
	event, _ := searchJupiterGalileanPhenomenonEvent(jd, satellite, phenomenonType, -1, true)
	return event
}

// NextJupiterGalileanPhenomenonEvent 下一次伽利略卫星现象 / next Galilean-satellite event.
//
// jd 与返回时刻都是民用儒略日；需要瞬时状态时先用 UTC2TT(jd) 换成 TT。
func NextJupiterGalileanPhenomenonEvent(jd float64, satellite int, phenomenonType JupiterGalileanPhenomenonType) JupiterGalileanPhenomenonEvent {
	event, _ := searchJupiterGalileanPhenomenonEvent(jd, satellite, phenomenonType, 1, false)
	return event
}

// ClosestJupiterGalileanPhenomenonEvent 最近一次伽利略卫星现象 / closest Galilean-satellite event.
//
// jd 与返回时刻都是民用儒略日；需要瞬时状态时先用 UTC2TT(jd) 换成 TT。
func ClosestJupiterGalileanPhenomenonEvent(jd float64, satellite int, phenomenonType JupiterGalileanPhenomenonType) JupiterGalileanPhenomenonEvent {
	last, hasLast := searchJupiterGalileanPhenomenonEvent(jd, satellite, phenomenonType, -1, true)
	next, hasNext := searchJupiterGalileanPhenomenonEvent(jd, satellite, phenomenonType, 1, false)
	switch {
	case hasLast && !hasNext:
		return last
	case !hasLast && hasNext:
		return next
	case !hasLast && !hasNext:
		return invalidJupiterGalileanPhenomenonEvent()
	}
	if math.Abs(last.Greatest-jd) <= math.Abs(next.Greatest-jd) {
		return last
	}
	return next
}

func searchJupiterGalileanPhenomenonEvent(
	jd float64,
	satellite int,
	phenomenonType JupiterGalileanPhenomenonType,
	direction int,
	includeCurrent bool,
) (JupiterGalileanPhenomenonEvent, bool) {
	if !isFinite(jd) || direction == 0 || satellite < 1 || satellite > 4 || !isValidJupiterGalileanPhenomenonType(phenomenonType) {
		return invalidJupiterGalileanPhenomenonEvent(), false
	}

	if sample := jupiterGalileanPhenomenonMetricAt(jd, satellite, phenomenonType); sample.active {
		current := findJupiterGalileanPhenomenonEventAround(jd, satellite, phenomenonType)
		if current.Valid && includeCurrent {
			return current, true
		}
		if current.Valid {
			if direction > 0 {
				jd = current.End + jupiterGalileanEventEpsilonDays
			} else {
				jd = current.Start - jupiterGalileanEventEpsilonDays
			}
		}
	}

	stepDays := jupiterGalileanCoarseStepDays(satellite)
	maxSteps := int(math.Ceil(jupiterGalileanEventSearchSpanDays / stepDays))
	sign := float64(direction)

	prevTime := jd
	prevSample := jupiterGalileanPhenomenonMetricAt(prevTime, satellite, phenomenonType)
	midTime := jd + sign*stepDays
	midSample := jupiterGalileanPhenomenonMetricAt(midTime, satellite, phenomenonType)

	for i := 2; i <= maxSteps; i++ {
		nextTime := jd + sign*float64(i)*stepDays
		nextSample := jupiterGalileanPhenomenonMetricAt(nextTime, satellite, phenomenonType)
		if event, ok := jupiterGalileanIntervalCandidate(
			prevTime, nextTime, prevSample, midSample, nextSample,
			jd, satellite, phenomenonType, direction, includeCurrent,
		); ok {
			return event, true
		}
		prevTime, prevSample = midTime, midSample
		midTime, midSample = nextTime, nextSample
	}

	return invalidJupiterGalileanPhenomenonEvent(), false
}

// jupiterGalileanIntervalCandidate 在粗扫区间内确认现象：度量局部极小，或有符号偏移过零。
func jupiterGalileanIntervalCandidate(
	prevTime, nextTime float64,
	prevSample, midSample, nextSample jupiterGalileanMetricSample,
	jd float64,
	satellite int,
	phenomenonType JupiterGalileanPhenomenonType,
	direction int,
	includeCurrent bool,
) (JupiterGalileanPhenomenonEvent, bool) {
	var best JupiterGalileanPhenomenonEvent
	found := false
	consider := func(event JupiterGalileanPhenomenonEvent) {
		if !event.Valid ||
			!jupiterGalileanEventMatchesDirection(event.Greatest, jd, direction, includeCurrent) {
			return
		}
		if !found ||
			(direction > 0 && event.Greatest < best.Greatest) ||
			(direction < 0 && event.Greatest > best.Greatest) {
			best, found = event, true
		}
	}
	if isFinite(midSample.metric) &&
		midSample.metric <= prevSample.metric &&
		midSample.metric <= nextSample.metric {
		consider(findJupiterGalileanPhenomenonEventAround(
			refineJupiterGalileanMetricMinimum(prevTime, nextTime, satellite, phenomenonType),
			satellite, phenomenonType,
		))
	}
	if jupiterGalileanSignedCrosses(prevSample.signed, nextSample.signed) &&
		jupiterGalileanNearDisk(prevSample, midSample, nextSample) {
		crossing := refineJupiterGalileanSignedCrossing(prevTime, nextTime, satellite, phenomenonType)
		// 过零点只说明"靠近木星"，只在它已贴近视面时按事件尺度细扫一次。
		if sample := jupiterGalileanPhenomenonMetricAt(crossing, satellite, phenomenonType); isFinite(sample.metric) &&
			sample.metric <= jupiterGalileanCrossingProbeMetric {
			if event, ok := jupiterGalileanFineScanAround(crossing, satellite, phenomenonType); ok {
				consider(event)
			}
		}
	}
	return best, found
}

const (
	// jupiterGalileanCrossingProbeMetric 是细扫门限（椭球度量，1 为视面边缘）。
	jupiterGalileanCrossingProbeMetric = 2.5
	// jupiterGalileanCrossingProbeDays 是过零点两侧的细扫半宽。
	jupiterGalileanCrossingProbeDays = 15.0 / 1440.0
	// jupiterGalileanCrossingProbeStepDays 是细扫步长（30 秒），短于最短的擦边事件。
	jupiterGalileanCrossingProbeStepDays = 0.5 / 1440.0
)

// jupiterGalileanNearDisk 报告区间内是否已有贴近木星视面的采样点。
func jupiterGalileanNearDisk(prevSample, midSample, nextSample jupiterGalileanMetricSample) bool {
	for _, metric := range [3]float64{prevSample.metric, midSample.metric, nextSample.metric} {
		if isFinite(metric) && metric <= jupiterGalileanCrossingProbeMetric {
			return true
		}
	}
	return false
}

// jupiterGalileanFineScanAround 在过零点邻域按事件尺度取样，命中后交给事件求解。
func jupiterGalileanFineScanAround(
	jd float64,
	satellite int,
	phenomenonType JupiterGalileanPhenomenonType,
) (JupiterGalileanPhenomenonEvent, bool) {
	// 从过零点向两侧交替外扩，通常几步内就能落进事件窗口。
	steps := int(math.Round(jupiterGalileanCrossingProbeDays / jupiterGalileanCrossingProbeStepDays))
	for step := 0; step <= steps; step++ {
		for _, sign := range [2]float64{1, -1} {
			if step == 0 && sign < 0 {
				continue
			}
			candidate := jd + sign*float64(step)*jupiterGalileanCrossingProbeStepDays
			if !jupiterGalileanPhenomenonMetricAt(candidate, satellite, phenomenonType).active {
				continue
			}
			if event := findJupiterGalileanPhenomenonEventAround(candidate, satellite, phenomenonType); event.Valid {
				return event, true
			}
		}
	}
	return JupiterGalileanPhenomenonEvent{}, false
}

// jupiterGalileanSignedCrosses 判断两次采样的有符号偏移是否变号（端点为 0 也算）。
func jupiterGalileanSignedCrosses(first, second float64) bool {
	if !isFinite(first) || !isFinite(second) {
		return false
	}
	if first == 0 || second == 0 {
		return true
	}
	return (first < 0) != (second < 0)
}

// refineJupiterGalileanSignedCrossing 用二分把过零时刻收敛到事件容差内。
func refineJupiterGalileanSignedCrossing(
	left, right float64,
	satellite int,
	phenomenonType JupiterGalileanPhenomenonType,
) float64 {
	leftValue := jupiterGalileanPhenomenonMetricAt(left, satellite, phenomenonType).signed
	rightValue := jupiterGalileanPhenomenonMetricAt(right, satellite, phenomenonType).signed
	if !isFinite(leftValue) || !isFinite(rightValue) || (leftValue < 0) == (rightValue < 0) {
		return (left + right) / 2
	}
	for i := 0; i < 80 && right-left > jupiterGalileanEventEpsilonDays; i++ {
		middle := (left + right) / 2
		value := jupiterGalileanPhenomenonMetricAt(middle, satellite, phenomenonType).signed
		if !isFinite(value) {
			return middle
		}
		if (value < 0) == (leftValue < 0) {
			left, leftValue = middle, value
		} else {
			right, rightValue = middle, value
		}
	}
	return (left + right) / 2
}

func findJupiterGalileanPhenomenonEventAround(jd float64, satellite int, phenomenonType JupiterGalileanPhenomenonType) JupiterGalileanPhenomenonEvent {
	sample := jupiterGalileanPhenomenonMetricAt(jd, satellite, phenomenonType)
	if !sample.active {
		return invalidJupiterGalileanPhenomenonEvent()
	}

	stepDays := jupiterGalileanBoundaryStep(satellite)
	maxBoundarySteps := int(math.Ceil(jupiterGalileanOrbitPeriodDays(satellite)/stepDays)) + 4

	activeStart := jd
	inactiveStart := math.NaN()
	for i := 0; i < maxBoundarySteps; i++ {
		candidate := activeStart - stepDays
		if !jupiterGalileanPhenomenonMetricAt(candidate, satellite, phenomenonType).active {
			inactiveStart = candidate
			break
		}
		activeStart = candidate
	}
	if !isFinite(inactiveStart) {
		return invalidJupiterGalileanPhenomenonEvent()
	}

	activeEnd := jd
	inactiveEnd := math.NaN()
	for i := 0; i < maxBoundarySteps; i++ {
		candidate := activeEnd + stepDays
		if !jupiterGalileanPhenomenonMetricAt(candidate, satellite, phenomenonType).active {
			inactiveEnd = candidate
			break
		}
		activeEnd = candidate
	}
	if !isFinite(inactiveEnd) {
		return invalidJupiterGalileanPhenomenonEvent()
	}

	start := refineJupiterGalileanEventStart(inactiveStart, activeStart, satellite, phenomenonType)
	end := refineJupiterGalileanEventEnd(activeEnd, inactiveEnd, satellite, phenomenonType)
	if !isFinite(start) || !isFinite(end) || end <= start {
		return invalidJupiterGalileanPhenomenonEvent()
	}

	greatest := refineJupiterGalileanMetricMinimum(start, end, satellite, phenomenonType)
	greatestSample := jupiterGalileanPhenomenonMetricAt(greatest, satellite, phenomenonType)
	if !greatestSample.active {
		return invalidJupiterGalileanPhenomenonEvent()
	}

	return JupiterGalileanPhenomenonEvent{
		Valid:              true,
		Satellite:          satellite,
		Type:               phenomenonType,
		Start:              start,
		Greatest:           greatest,
		End:                end,
		GreatestPhenomenon: greatestSample.phenomenon,
	}
}

func refineJupiterGalileanEventStart(
	outsideJD, insideJD float64,
	satellite int,
	phenomenonType JupiterGalileanPhenomenonType,
) float64 {
	if insideJD < outsideJD {
		outsideJD, insideJD = insideJD, outsideJD
	}
	if jupiterGalileanPhenomenonMetricAt(outsideJD, satellite, phenomenonType).active {
		return math.NaN()
	}
	if !jupiterGalileanPhenomenonMetricAt(insideJD, satellite, phenomenonType).active {
		return math.NaN()
	}
	left := outsideJD
	right := insideJD
	for i := 0; i < 80 && right-left > jupiterGalileanEventEpsilonDays; i++ {
		mid := (left + right) / 2
		if jupiterGalileanPhenomenonMetricAt(mid, satellite, phenomenonType).active {
			right = mid
		} else {
			left = mid
		}
	}
	return right
}

func refineJupiterGalileanEventEnd(
	insideJD, outsideJD float64,
	satellite int,
	phenomenonType JupiterGalileanPhenomenonType,
) float64 {
	if outsideJD < insideJD {
		insideJD, outsideJD = outsideJD, insideJD
	}
	if !jupiterGalileanPhenomenonMetricAt(insideJD, satellite, phenomenonType).active {
		return math.NaN()
	}
	if jupiterGalileanPhenomenonMetricAt(outsideJD, satellite, phenomenonType).active {
		return math.NaN()
	}
	left := insideJD
	right := outsideJD
	for i := 0; i < 80 && right-left > jupiterGalileanEventEpsilonDays; i++ {
		mid := (left + right) / 2
		if jupiterGalileanPhenomenonMetricAt(mid, satellite, phenomenonType).active {
			left = mid
		} else {
			right = mid
		}
	}
	return left
}

func refineJupiterGalileanMetricMinimum(
	jd1, jd2 float64,
	satellite int,
	phenomenonType JupiterGalileanPhenomenonType,
) float64 {
	left := math.Min(jd1, jd2)
	right := math.Max(jd1, jd2)
	if right-left <= jupiterGalileanEventEpsilonDays {
		return (left + right) / 2
	}
	const phi = 0.6180339887498948482
	x1 := right - phi*(right-left)
	x2 := left + phi*(right-left)
	f1 := jupiterGalileanPhenomenonMetricAt(x1, satellite, phenomenonType).metric
	f2 := jupiterGalileanPhenomenonMetricAt(x2, satellite, phenomenonType).metric
	for i := 0; i < 80 && right-left > jupiterGalileanEventEpsilonDays; i++ {
		if f1 <= f2 {
			right = x2
			x2 = x1
			f2 = f1
			x1 = right - phi*(right-left)
			f1 = jupiterGalileanPhenomenonMetricAt(x1, satellite, phenomenonType).metric
		} else {
			left = x1
			x1 = x2
			f1 = f2
			x2 = left + phi*(right-left)
			f2 = jupiterGalileanPhenomenonMetricAt(x2, satellite, phenomenonType).metric
		}
	}
	return (left + right) / 2
}

func jupiterGalileanPhenomenonMetricAt(
	jd float64,
	satellite int,
	phenomenonType JupiterGalileanPhenomenonType,
) jupiterGalileanMetricSample {
	if !isFinite(jd) || satellite < 1 || satellite > 4 || !isValidJupiterGalileanPhenomenonType(phenomenonType) {
		return jupiterGalileanMetricSample{
			metric:     math.Inf(1),
			signed:     math.NaN(),
			phenomenon: invalidJupiterGalileanPhenomenon(),
		}
	}

	evaluationJDE := UTC2TT(jd)
	context := newJupiterGalileanObservationContext(evaluationJDE)
	if context.jupiterDistance == 0 {
		return jupiterGalileanMetricSample{
			metric:     math.Inf(1),
			signed:     math.NaN(),
			phenomenon: invalidJupiterGalileanPhenomenon(),
		}
	}

	index := satellite - 1
	observation := context.observationForSatellite(index)
	stateVector := Vector3{observation.State.X, observation.State.Y, observation.State.Z}
	radiusAU := jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM

	xEarth := observation.OffsetXJupiterRadii
	yEarth := observation.OffsetYJupiterRadii
	earthMetric := ellipseMetric(xEarth, yEarth, 1, context.earthMinorRadius)
	onEarthDisk := ellipseInside(xEarth, yEarth, 1, context.earthMinorRadius)

	xSunAU := vectorDot(stateVector, context.sunEast)
	ySunAU := vectorDot(stateVector, context.sunNorth)
	zSunAU := vectorDot(stateVector, context.sunLineOfSight)
	xSun := xSunAU / radiusAU
	ySun := ySunAU / radiusAU
	umbraScale := jupiterUmbraScale(zSunAU, context.sunDistanceAU)
	sunMetric := math.Inf(1)
	eclipse := false
	if umbraScale > 0 {
		sunMetric = ellipseMetric(xSun, ySun, umbraScale, context.sunMinorRadius*umbraScale)
		eclipse = zSunAU > 0 && sunMetric <= 1+1e-12
	}

	shadowPoint := context.shadowPointFor(stateVector)
	shadowMetric := math.Inf(1)
	shadowTransit := false
	if shadowPoint.hasIntersection {
		shadowMetric = ellipseMetric(shadowPoint.xJupiterRadii, shadowPoint.yJupiterRadii, 1, context.earthMinorRadius)
		shadowTransit = shadowPoint.visible && shadowMetric <= 1+1e-12
	}

	phenomenon := JupiterGalileanPhenomenon{
		Transit:       onEarthDisk && observation.InFrontOfJupiter,
		Occultation:   onEarthDisk && !observation.InFrontOfJupiter,
		Eclipse:       eclipse,
		ShadowTransit: shadowTransit,

		ShadowOffsetXArcsec:       math.NaN(),
		ShadowOffsetYArcsec:       math.NaN(),
		ShadowOffsetXJupiterRadii: math.NaN(),
		ShadowOffsetYJupiterRadii: math.NaN(),
	}
	if shadowTransit {
		phenomenon.ShadowOffsetXArcsec = math.Atan2(shadowPoint.xAU, context.jupiterDistance) * deg * 3600
		phenomenon.ShadowOffsetYArcsec = math.Atan2(shadowPoint.yAU, context.jupiterDistance) * deg * 3600
		phenomenon.ShadowOffsetXJupiterRadii = shadowPoint.xJupiterRadii
		phenomenon.ShadowOffsetYJupiterRadii = shadowPoint.yJupiterRadii
	}

	switch phenomenonType {
	case JupiterGalileanTransit:
		metric := earthMetric
		if !observation.InFrontOfJupiter {
			metric += 4
		}
		return jupiterGalileanMetricSample{
			active: phenomenon.Transit, metric: metric, signed: xEarth, phenomenon: phenomenon,
		}
	case JupiterGalileanOccultation:
		metric := earthMetric
		if observation.InFrontOfJupiter {
			metric += 4
		}
		return jupiterGalileanMetricSample{
			active: phenomenon.Occultation, metric: metric, signed: xEarth, phenomenon: phenomenon,
		}
	case JupiterGalileanEclipse:
		metric := sunMetric
		if zSunAU <= 0 {
			metric += 4
		}
		return jupiterGalileanMetricSample{
			active: phenomenon.Eclipse, metric: metric, signed: xSun, phenomenon: phenomenon,
		}
	case JupiterGalileanShadowTransit:
		metric := shadowMetric
		if shadowPoint.hasIntersection && !shadowPoint.visible {
			metric += 4
		}
		signed := math.NaN()
		if shadowPoint.hasIntersection {
			signed = shadowPoint.xJupiterRadii
		}
		return jupiterGalileanMetricSample{
			active: phenomenon.ShadowTransit, metric: metric, signed: signed, phenomenon: phenomenon,
		}
	default:
		return jupiterGalileanMetricSample{
			metric: math.Inf(1), signed: math.NaN(), phenomenon: invalidJupiterGalileanPhenomenon(),
		}
	}
}

func (context jupiterGalileanObservationContext) shadowPointFor(stateVector Vector3) jupiterGalileanShadowPoint {
	radiusAU := jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM
	satelliteBody := context.toBodyCoordinates(stateVector)
	satelliteBody = Vector3{
		satelliteBody[0] / radiusAU,
		satelliteBody[1] / radiusAU,
		satelliteBody[2] / radiusAU,
	}
	directionBody := context.toBodyCoordinates(context.sunLineOfSight)
	intersectionBody, ok := ellipsoidRayIntersection(satelliteBody, directionBody, jupiterPolarRadiusRatio())
	if !ok {
		return jupiterGalileanShadowPoint{}
	}
	normalBody := Vector3{
		intersectionBody[0],
		intersectionBody[1],
		intersectionBody[2] / (jupiterPolarRadiusRatio() * jupiterPolarRadiusRatio()),
	}
	earthBody := context.toBodyCoordinates(context.earthDirection)
	intersection := context.fromBodyCoordinates(Vector3{
		intersectionBody[0] * radiusAU,
		intersectionBody[1] * radiusAU,
		intersectionBody[2] * radiusAU,
	})
	dx := intersection[0] - stateVector[0]
	dy := intersection[1] - stateVector[1]
	dz := intersection[2] - stateVector[2]
	xAU := vectorDot(intersection, context.east)
	yAU := vectorDot(intersection, context.north)
	xR := xAU / radiusAU
	yR := yAU / radiusAU
	return jupiterGalileanShadowPoint{
		hasIntersection: true,
		visible:         vectorDot(normalBody, earthBody) > 0,
		pathLengthAU:    math.Sqrt(dx*dx + dy*dy + dz*dz),
		xAU:             xAU,
		yAU:             yAU,
		xJupiterRadii:   xR,
		yJupiterRadii:   yR,
	}
}

func jupiterGalileanOrbitPeriodDays(satellite int) float64 {
	switch satellite {
	case 1:
		return 1.769137786
	case 2:
		return 3.551181
	case 3:
		return 7.154553
	case 4:
		return 16.689018
	default:
		return math.NaN()
	}
}

func jupiterGalileanCoarseStepDays(satellite int) float64 {
	step := jupiterGalileanOrbitPeriodDays(satellite) / 16
	maxStep := 2.0 / 24.0
	if step > maxStep {
		return maxStep
	}
	return step
}

func jupiterGalileanBoundaryStep(satellite int) float64 {
	step := jupiterGalileanOrbitPeriodDays(satellite) / 32
	if step > jupiterGalileanBoundaryStepDays {
		return jupiterGalileanBoundaryStepDays
	}
	return step
}

func jupiterGalileanEventMatchesDirection(eventJD, targetJD float64, direction int, includeCurrent bool) bool {
	diff := eventJD - targetJD
	switch {
	case direction < 0 && includeCurrent:
		return diff <= jupiterGalileanEventEpsilonDays
	case direction < 0:
		return diff < -jupiterGalileanEventEpsilonDays
	case includeCurrent:
		return diff >= -jupiterGalileanEventEpsilonDays
	default:
		return diff > jupiterGalileanEventEpsilonDays
	}
}

func ellipseMetric(x, y, major, minor float64) float64 {
	if major <= 0 || minor <= 0 {
		return math.Inf(1)
	}
	return (x*x)/(major*major) + (y*y)/(minor*minor)
}

func isValidJupiterGalileanPhenomenonType(phenomenonType JupiterGalileanPhenomenonType) bool {
	switch phenomenonType {
	case JupiterGalileanTransit, JupiterGalileanOccultation, JupiterGalileanEclipse, JupiterGalileanShadowTransit:
		return true
	default:
		return false
	}
}

func invalidJupiterGalileanPhenomenonEvent() JupiterGalileanPhenomenonEvent {
	return JupiterGalileanPhenomenonEvent{
		Start:              math.NaN(),
		Greatest:           math.NaN(),
		End:                math.NaN(),
		GreatestPhenomenon: invalidJupiterGalileanPhenomenon(),
	}
}
