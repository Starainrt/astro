package basic

import "math"

const (
	jupiterGalileanMutualSearchSpanDays       = 7 * 365.25
	jupiterGalileanMutualMinimumIntervalDays  = 1.0 / 1440.0
	jupiterGalileanMutualOuterMinimumStepDays = 1.0 / 24.0
	jupiterGalileanMutualOuterMaximumStepDays = 1.0
	jupiterGalileanMutualSpeedSafetyFactor    = 1.25
	jupiterGalileanMutualBoundaryStepDays     = 1.0 / 1440.0
	jupiterGalileanMutualToleranceDays        = 0.1 / 86400.0
	jupiterGalileanMutualEventEpsilonDays     = 0.1 / 86400.0
	jupiterGalileanMutualSeasonJumpStepDays   = 1.0
)

// JupiterGalileanMutualPhenomenonType 伽利略卫星互现象类型 / Galilean mutual phenomenon type.
type JupiterGalileanMutualPhenomenonType string

const (
	// JupiterGalileanMutualOccultation 互掩 / mutual occultation.
	JupiterGalileanMutualOccultation JupiterGalileanMutualPhenomenonType = "mutual_occultation"
	// JupiterGalileanMutualEclipse 互食 / mutual eclipse.
	JupiterGalileanMutualEclipse JupiterGalileanMutualPhenomenonType = "mutual_eclipse"
)

// JupiterGalileanMutualEvent 伽利略卫星互现象整场事件 / full Galilean mutual phenomenon event.
type JupiterGalileanMutualEvent struct {
	Valid bool
	Type  JupiterGalileanMutualPhenomenonType

	Active  int
	Passive int

	Start    float64
	Greatest float64
	End      float64

	DurationDays float64

	MinimumSeparationArcsec float64
	ImpactParameter         float64
	Coverage                float64
	Annular                 bool

	GreatestState [4]JupiterGalileanPhenomenon
}

// LastJupiterGalileanMutualEventWithInstance 用指定星历实例求上一次互现象事件 / previous mutual phenomenon with an explicit instance.
func LastJupiterGalileanMutualEventWithInstance(jde float64, first, second int, phenomenonType JupiterGalileanMutualPhenomenonType, inst *Instance) JupiterGalileanMutualEvent {
	event, _ := searchJupiterGalileanMutualEvent(jde, first, second, phenomenonType, -1, true, inst)
	return event
}

// NextJupiterGalileanMutualEventWithInstance 用指定星历实例求下一次互现象事件 / next mutual phenomenon with an explicit instance.
func NextJupiterGalileanMutualEventWithInstance(jde float64, first, second int, phenomenonType JupiterGalileanMutualPhenomenonType, inst *Instance) JupiterGalileanMutualEvent {
	event, _ := searchJupiterGalileanMutualEvent(jde, first, second, phenomenonType, 1, false, inst)
	return event
}

// NextJupiterGalileanMutualEventBeforeWithInstance 下一场极大接收时刻早于 end 的互现象，时标为 TT / next mutual event with TT reception Greatest before end.
func NextJupiterGalileanMutualEventBeforeWithInstance(jde, end float64, first, second int, phenomenonType JupiterGalileanMutualPhenomenonType, inst *Instance) JupiterGalileanMutualEvent {
	if !isFinite(end) || end <= jde {
		return invalidJupiterGalileanMutualEvent()
	}
	event, _ := searchJupiterGalileanMutualEventBefore(jde, first, second, phenomenonType, 1, false, true, end, inst)
	return event
}

// ClosestJupiterGalileanMutualEventWithInstance 求离指定时刻最近的互现象事件 / closest mutual phenomenon with an explicit instance.
func ClosestJupiterGalileanMutualEventWithInstance(jde float64, first, second int, phenomenonType JupiterGalileanMutualPhenomenonType, inst *Instance) JupiterGalileanMutualEvent {
	last, hasLast := searchJupiterGalileanMutualEvent(jde, first, second, phenomenonType, -1, true, inst)
	next, hasNext := searchJupiterGalileanMutualEvent(jde, first, second, phenomenonType, 1, false, inst)
	switch {
	case hasLast && !hasNext:
		return last
	case !hasLast && hasNext:
		return next
	case !hasLast && !hasNext:
		return invalidJupiterGalileanMutualEvent()
	}
	if math.Abs(last.Greatest-jde) <= math.Abs(next.Greatest-jde) {
		return last
	}
	return next
}

func searchJupiterGalileanMutualEvent(
	jde float64,
	first, second int,
	phenomenonType JupiterGalileanMutualPhenomenonType,
	direction int,
	includeCurrent bool,
	instances ...*Instance,
) (JupiterGalileanMutualEvent, bool) {
	return searchJupiterGalileanMutualEventWithSeasonGate(jde, first, second, phenomenonType, direction, includeCurrent, true, instances...)
}

func searchJupiterGalileanMutualEventWithSeasonGate(
	jde float64,
	first, second int,
	phenomenonType JupiterGalileanMutualPhenomenonType,
	direction int,
	includeCurrent bool,
	seasonGate bool,
	instances ...*Instance,
) (JupiterGalileanMutualEvent, bool) {
	return searchJupiterGalileanMutualEventBefore(jde, first, second, phenomenonType, direction, includeCurrent, seasonGate, math.Inf(1), instances...)
}

func searchJupiterGalileanMutualEventBefore(
	jde float64,
	first, second int,
	phenomenonType JupiterGalileanMutualPhenomenonType,
	direction int,
	includeCurrent, seasonGate bool,
	end float64,
	instances ...*Instance,
) (JupiterGalileanMutualEvent, bool) {
	inst := instanceValue(instances)
	first, second, ok := normalizeJupiterGalileanMutualPair(first, second)
	if !ok || !isFinite(jde) || (direction != -1 && direction != 1) {
		return invalidJupiterGalileanMutualEvent(), false
	}
	if phenomenonType == JupiterGalileanMutualEclipse {
		search := jupiterGalileanEclipseSearch{first: first, second: second, inst: inst}
		return search.find(jde, direction, includeCurrent, seasonGate, end)
	}
	if phenomenonType != JupiterGalileanMutualOccultation {
		return invalidJupiterGalileanMutualEvent(), false
	}

	searchStart := jde
	if seasonGate && !jupiterGalileanMutualSeasonAllows(searchStart, first, second, inst) {
		searchStart, ok = seekJupiterGalileanMutualSeason(searchStart, direction, first, second, end, inst)
		if !ok {
			return invalidJupiterGalileanMutualEvent(), false
		}
	}

	if current, found := findJupiterGalileanMutualEventAround(searchStart, first, second, phenomenonType, inst); found {
		if mutualEventMatchesDirection(current.Greatest, jde, direction, includeCurrent) {
			if direction > 0 && current.Greatest >= end {
				return invalidJupiterGalileanMutualEvent(), false
			}
			return current, true
		}
		if direction > 0 {
			searchStart = current.End + jupiterGalileanMutualEventEpsilonDays
		} else {
			searchStart = current.Start - jupiterGalileanMutualEventEpsilonDays
		}
	}

	limit := jde + float64(direction)*jupiterGalileanMutualSearchSpanDays
	speedUpperBound := jupiterGalileanMutualRelativeSpeedUpperBound(first, second)
	polarSpeed := jupiterGalileanMutualPolarSpeedUpperBound(first, second)
	contactRadius := jupiterGalileanMutualContactRadius(first, second)
	if !isFinite(speedUpperBound) || speedUpperBound <= 0 || !isFinite(contactRadius) || contactRadius <= 0 {
		return invalidJupiterGalileanMutualEvent(), false
	}
	currentTime := searchStart
	memo := &jupiterGalileanMutualSplitMemo{}
	for {
		if (direction > 0 && (currentTime >= limit || currentTime >= end)) || (direction < 0 && currentTime <= limit) {
			break
		}
		if seasonGate && !jupiterGalileanMutualSeasonAllows(currentTime, first, second, inst) {
			allowedTime, found := seekJupiterGalileanMutualSeason(currentTime, direction, first, second, end, inst)
			if !found || (direction > 0 && allowedTime > limit) || (direction < 0 && allowedTime < limit) {
				break
			}
			currentTime = allowedTime
			continue
		}

		geometry, separationOK := memo.at(currentTime, first, second, inst)
		stepDays := jupiterGalileanMutualOuterMaximumStepDays
		if separationOK {
			stepDays = jupiterGalileanMutualOuterStep(geometry.Separation-contactRadius, geometry.Polar-contactRadius, speedUpperBound, polarSpeed, geometry.PolarAllowed)
		}
		nextTime := currentTime + float64(direction)*stepDays
		if separationOK && geometry.PolarAllowed {
			endpoint, endpointOK := memo.at(nextTime, first, second, inst)
			if !endpointOK || !endpoint.PolarAllowed {
				stepDays = jupiterGalileanMutualOuterStep(geometry.Separation-contactRadius, 0, speedUpperBound, polarSpeed, false)
				nextTime = currentTime + float64(direction)*stepDays
			}
		}
		if direction > 0 && nextTime > limit {
			nextTime = limit
		} else if direction < 0 && nextTime < limit {
			nextTime = limit
		}
		if nextTime == currentTime {
			break
		}
		if seasonGate && !jupiterGalileanMutualSeasonAllows(nextTime, first, second, inst) {
			allowedTime, found := seekJupiterGalileanMutualSeason(nextTime, direction, first, second, end, inst)
			if !found || (direction > 0 && allowedTime > limit) || (direction < 0 && allowedTime < limit) {
				break
			}
			currentTime = allowedTime
			continue
		}

		left, right := currentTime, nextTime
		if left > right {
			left, right = right, left
		}
		if hit, found := findJupiterGalileanMutualHitInInterval(left, right, direction, first, second, contactRadius, speedUpperBound, inst, memo); found {
			if event, eventFound := findJupiterGalileanMutualEventAround(hit, first, second, phenomenonType, inst); eventFound {
				if mutualEventMatchesDirection(event.Greatest, jde, direction, includeCurrent) {
					if direction > 0 && event.Greatest >= end {
						return invalidJupiterGalileanMutualEvent(), false
					}
					return event, true
				}
				if direction > 0 {
					currentTime = event.End + jupiterGalileanMutualEventEpsilonDays
				} else {
					currentTime = event.Start - jupiterGalileanMutualEventEpsilonDays
				}
				continue
			}
		}
		currentTime = nextTime
	}
	return invalidJupiterGalileanMutualEvent(), false
}

func normalizeJupiterGalileanMutualPair(first, second int) (int, int, bool) {
	if first < 1 || first > 4 || second < 1 || second > 4 || first == second {
		return 0, 0, false
	}
	if first > second {
		first, second = second, first
	}
	return first, second, true
}

func mutualEventMatchesDirection(greatest, seed float64, direction int, includeCurrent bool) bool {
	if direction > 0 {
		if includeCurrent {
			return greatest >= seed
		}
		return greatest > seed
	}
	if includeCurrent {
		return greatest <= seed
	}
	return greatest < seed
}

func seekJupiterGalileanMutualSeason(start float64, direction, first, second int, end float64, inst *Instance) (float64, bool) {
	limit := start + float64(direction)*jupiterGalileanMutualSearchSpanDays
	candidate := start
	for {
		candidate += float64(direction) * jupiterGalileanMutualSeasonJumpStepDays
		if (direction > 0 && (candidate > limit || candidate >= end)) || (direction < 0 && candidate < limit) {
			return 0, false
		}
		if jupiterGalileanMutualSeasonAllows(candidate, first, second, inst) {
			return candidate, true
		}
	}
}

func jupiterGalileanMutualMetricAt(jde float64, first, second int, inst *Instance) (float64, bool) {
	geometry, ok := jupiterGalileanMutualGeometryAt(jde, first, second, inst)
	if !ok || !isFinite(geometry.Separation) {
		return 0, false
	}
	return geometry.Separation * geometry.Separation, true
}

func jupiterGalileanMutualSeparationAt(jde float64, first, second int, inst *Instance) (float64, bool) {
	geometry, ok := jupiterGalileanMutualGeometryAt(jde, first, second, inst)
	if !ok || !isFinite(geometry.Separation) {
		return 0, false
	}
	return geometry.Separation, true
}

func jupiterGalileanMutualPolarSpeedUpperBound(first, second int) float64 {
	first, second, ok := normalizeJupiterGalileanMutualPair(first, second)
	if !ok {
		return math.NaN()
	}
	bound := 0.0
	for _, satellite := range []int{first, second} {
		axis := jupiterGalileanMutualSemimajorAxis(satellite)
		speed := 2 * math.Pi * axis / jupiterGalileanOrbitPeriodDays(satellite)
		bound += speed*(math.Sin(jupiterGalileanMutualMaximumInclination(satellite)*rad)+math.Sin(4*rad)) + 0.005*axis
	}
	return jupiterGalileanMutualSpeedSafetyFactor * bound
}

func jupiterGalileanMutualOuterStep(gap, polarGap, speed, polarSpeed float64, polarAllowed bool) float64 {
	step := gap / speed
	if polarAllowed {
		step = math.Max(step, polarGap/polarSpeed)
	}
	return math.Max(jupiterGalileanMutualOuterMinimumStepDays, math.Min(jupiterGalileanMutualOuterMaximumStepDays, step))
}

func jupiterGalileanMutualPolarExcludes(leftGap, rightGap, speed, span float64, leftAllowed, rightAllowed bool) bool {
	return leftAllowed && rightAllowed && (leftGap+rightGap-speed*span)/2 > 0
}

type jupiterGalileanMutualSplitMemo struct {
	entries [4]struct {
		bits     uint64
		geometry jupiterGalileanMutualGeometry
		filled   bool
		valid    bool
	}
	next int
}

func (memo *jupiterGalileanMutualSplitMemo) at(jde float64, first, second int, inst *Instance) (jupiterGalileanMutualGeometry, bool) {
	bits := math.Float64bits(jde)
	for _, entry := range memo.entries {
		if entry.filled && entry.bits == bits {
			return entry.geometry, entry.valid
		}
	}
	geometry, valid := jupiterGalileanMutualGeometryAt(jde, first, second, inst)
	entry := &memo.entries[memo.next]
	entry.bits, entry.geometry, entry.filled, entry.valid = bits, geometry, true, valid
	memo.next = (memo.next + 1) % len(memo.entries)
	return geometry, valid
}

func jupiterGalileanMutualContactRadius(first, second int) float64 {
	return jupiterGalileanSatelliteRadiusJupiterRadii(first) + jupiterGalileanSatelliteRadiusJupiterRadii(second)
}

func jupiterGalileanMutualRelativeSpeedUpperBound(first, second int) float64 {
	first, second, ok := normalizeJupiterGalileanMutualPair(first, second)
	if !ok {
		return math.NaN()
	}
	firstSpeed := 2 * math.Pi * jupiterGalileanMutualSemimajorAxis(first) / jupiterGalileanOrbitPeriodDays(first)
	secondSpeed := 2 * math.Pi * jupiterGalileanMutualSemimajorAxis(second) / jupiterGalileanOrbitPeriodDays(second)
	return jupiterGalileanMutualSpeedSafetyFactor * (firstSpeed + secondSpeed)
}

func findJupiterGalileanMutualHitInInterval(left, right float64, direction, first, second int, contactRadius, speedUpperBound float64, inst *Instance, memo *jupiterGalileanMutualSplitMemo) (float64, bool) {
	leftGeometry, leftOK := memo.at(left, first, second, inst)
	rightGeometry, rightOK := memo.at(right, first, second, inst)
	if !leftOK || !rightOK {
		return 0, false
	}
	leftSeparation, rightSeparation := leftGeometry.Separation, rightGeometry.Separation
	if direction > 0 {
		if leftSeparation < contactRadius {
			return left, true
		}
		if rightSeparation < contactRadius {
			return right, true
		}
	} else {
		if rightSeparation < contactRadius {
			return right, true
		}
		if leftSeparation < contactRadius {
			return left, true
		}
	}
	if (leftSeparation+rightSeparation-speedUpperBound*(right-left))/2 > contactRadius {
		return 0, false
	}
	if jupiterGalileanMutualPolarExcludes(leftGeometry.Polar-contactRadius, rightGeometry.Polar-contactRadius, jupiterGalileanMutualPolarSpeedUpperBound(first, second), right-left, leftGeometry.PolarAllowed, rightGeometry.PolarAllowed) {
		return 0, false
	}
	if right-left < jupiterGalileanMutualMinimumIntervalDays {
		greatest := refineJupiterGalileanMutualMinimum(left, right, first, second, inst)
		separation, ok := jupiterGalileanMutualSeparationAt(greatest, first, second, inst)
		return greatest, ok && separation < contactRadius
	}
	middle := (left + right) / 2
	if direction > 0 {
		if hit, found := findJupiterGalileanMutualHitInInterval(left, middle, direction, first, second, contactRadius, speedUpperBound, inst, memo); found {
			return hit, true
		}
		return findJupiterGalileanMutualHitInInterval(middle, right, direction, first, second, contactRadius, speedUpperBound, inst, memo)
	}
	if hit, found := findJupiterGalileanMutualHitInInterval(middle, right, direction, first, second, contactRadius, speedUpperBound, inst, memo); found {
		return hit, true
	}
	return findJupiterGalileanMutualHitInInterval(left, middle, direction, first, second, contactRadius, speedUpperBound, inst, memo)
}

func refineJupiterGalileanMutualMinimum(left, right float64, first, second int, inst *Instance) float64 {
	if right < left {
		left, right = right, left
	}
	const golden = 0.6180339887498948482
	if right == left {
		return (left + right) / 2
	}
	x1 := right - golden*(right-left)
	x2 := left + golden*(right-left)
	f1, ok1 := jupiterGalileanMutualMetricAt(x1, first, second, inst)
	f2, ok2 := jupiterGalileanMutualMetricAt(x2, first, second, inst)
	if !ok1 || !ok2 {
		return (left + right) / 2
	}
	for i := 0; i < 80; i++ {
		if f1 <= f2 {
			right, x2, f2 = x2, x1, f1
			x1 = right - golden*(right-left)
			f1, ok1 = jupiterGalileanMutualMetricAt(x1, first, second, inst)
		} else {
			left, x1, f1 = x1, x2, f2
			x2 = left + golden*(right-left)
			f2, ok2 = jupiterGalileanMutualMetricAt(x2, first, second, inst)
		}
		if !ok1 || !ok2 {
			return (left + right) / 2
		}
	}
	return (left + right) / 2
}

func findJupiterGalileanMutualEventAround(seed float64, first, second int, phenomenonType JupiterGalileanMutualPhenomenonType, inst *Instance) (JupiterGalileanMutualEvent, bool) {
	if phenomenonType != JupiterGalileanMutualOccultation || !isFinite(seed) {
		return invalidJupiterGalileanMutualEvent(), false
	}
	geometry, ok := jupiterGalileanMutualGeometryAt(seed, first, second, inst)
	if !ok {
		return invalidJupiterGalileanMutualEvent(), false
	}
	activeRadius := jupiterGalileanSatelliteRadiusJupiterRadii(geometry.Active)
	passiveRadius := jupiterGalileanSatelliteRadiusJupiterRadii(geometry.Passive)
	contactRadius := activeRadius + passiveRadius
	if !isFinite(contactRadius) || contactRadius <= 0 || geometry.Separation >= contactRadius {
		return invalidJupiterGalileanMutualEvent(), false
	}
	active := geometry.Active
	start, startOK := refineJupiterGalileanMutualBoundary(seed, -1, contactRadius*contactRadius, first, second, inst)
	end, endOK := refineJupiterGalileanMutualBoundary(seed, 1, contactRadius*contactRadius, first, second, inst)
	if !startOK || !endOK || !(start <= seed && seed <= end) || end <= start {
		return invalidJupiterGalileanMutualEvent(), false
	}
	greatest := refineJupiterGalileanMutualMinimum(start, end, first, second, inst)
	greatestMetric, ok := jupiterGalileanMutualMetricAt(greatest, first, second, inst)
	if !ok {
		return invalidJupiterGalileanMutualEvent(), false
	}
	// 单峰互掩沿用整段黄金分割以保持逐位一致；仅当 1 分钟粗扫发现更深极小时改取全局最深极小。
	if end-start > 2*jupiterGalileanMutualMinimumIntervalDays {
		bestTime, bestMetric := greatest, greatestMetric
		for sample := 0; ; sample++ {
			candidate := math.Min(end, start+float64(sample)*jupiterGalileanMutualMinimumIntervalDays)
			metric, metricOK := jupiterGalileanMutualMetricAt(candidate, first, second, inst)
			if !metricOK {
				return invalidJupiterGalileanMutualEvent(), false
			}
			if metric < bestMetric {
				bestTime, bestMetric = candidate, metric
			}
			if candidate == end {
				break
			}
		}
		if bestMetric < greatestMetric {
			refined := refineJupiterGalileanMutualMinimum(math.Max(start, bestTime-jupiterGalileanMutualMinimumIntervalDays), math.Min(end, bestTime+jupiterGalileanMutualMinimumIntervalDays), first, second, inst)
			if refinedMetric, refinedOK := jupiterGalileanMutualMetricAt(refined, first, second, inst); refinedOK && refinedMetric < greatestMetric {
				greatest = refined
			}
		}
	}
	geometry, ok = jupiterGalileanMutualGeometryAt(greatest, first, second, inst)
	if !ok || geometry.Active != active {
		return invalidJupiterGalileanMutualEvent(), false
	}
	for _, boundary := range []float64{start, end} {
		boundaryGeometry, boundaryOK := jupiterGalileanMutualGeometryAt(boundary, first, second, inst)
		if !boundaryOK || boundaryGeometry.Active != active {
			return invalidJupiterGalileanMutualEvent(), false
		}
	}
	return buildJupiterGalileanMutualEvent(greatest, start, end, geometry, phenomenonType, inst)
}

func refineJupiterGalileanMutualBoundary(greatest float64, direction int, contactMetric float64, first, second int, inst *Instance) (float64, bool) {
	inside := greatest
	for i := 1; i <= int(math.Round(6.0/(1.0/60.0))); i++ {
		outside := greatest + float64(direction)*float64(i)*jupiterGalileanMutualBoundaryStepDays
		metric, ok := jupiterGalileanMutualMetricAt(outside, first, second, inst)
		if !ok {
			return 0, false
		}
		if metric > contactMetric {
			return refineJupiterGalileanMutualContactRoot(inside, outside, contactMetric, first, second, inst), true
		}
		inside = outside
	}
	return 0, false
}

func refineJupiterGalileanMutualContactRoot(inside, outside, contactMetric float64, first, second int, inst *Instance) float64 {
	for i := 0; i < 80; i++ {
		middle := (inside + outside) / 2
		metric, ok := jupiterGalileanMutualMetricAt(middle, first, second, inst)
		if !ok {
			return middle
		}
		if metric > contactMetric {
			outside = middle
		} else {
			inside = middle
		}
	}
	return (inside + outside) / 2
}

func buildJupiterGalileanMutualEvent(greatest, start, end float64, geometry jupiterGalileanMutualGeometry, phenomenonType JupiterGalileanMutualPhenomenonType, inst *Instance) (JupiterGalileanMutualEvent, bool) {
	context := newJupiterGalileanObservationContext(greatest, inst)
	if context.jupiterDistance == 0 {
		return invalidJupiterGalileanMutualEvent(), false
	}
	radiusAU := jupiterGalileanEquatorialRadiusKM / astronomicalUnitKM
	semidiameterArcsec := math.Atan2(radiusAU, context.jupiterDistance) * deg * 3600
	if !isFinite(semidiameterArcsec) || semidiameterArcsec <= 0 {
		return invalidJupiterGalileanMutualEvent(), false
	}
	var greatestState [4]JupiterGalileanPhenomenon
	for index := range greatestState {
		greatestState[index] = context.phenomenonForSatellite(index)
	}
	event := JupiterGalileanMutualEvent{
		Valid:                   true,
		Type:                    phenomenonType,
		Active:                  geometry.Active,
		Passive:                 geometry.Passive,
		Start:                   start,
		Greatest:                greatest,
		End:                     end,
		DurationDays:            end - start,
		MinimumSeparationArcsec: geometry.Separation * semidiameterArcsec,
		ImpactParameter:         geometry.ImpactParameter,
		Coverage:                geometry.Coverage,
		Annular:                 geometry.Annular,
		GreatestState:           greatestState,
	}
	if !isFinite(event.Start) || !isFinite(event.Greatest) || !isFinite(event.End) || !isFinite(event.DurationDays) || !isFinite(event.MinimumSeparationArcsec) || !isFinite(event.ImpactParameter) || !isFinite(event.Coverage) {
		return invalidJupiterGalileanMutualEvent(), false
	}
	return event, true
}

type jupiterGalileanEclipseSearch struct {
	first, second int
	inst          *Instance
	memo          *jupiterGalileanEclipseMemo
}

type jupiterGalileanEclipseMemo struct {
	entries [4]struct {
		bits     uint64
		geometry jupiterGalileanMutualEclipseGeometry
		filled   bool
		valid    bool
	}
	next int
}

func (search jupiterGalileanEclipseSearch) geometry(jde float64) (jupiterGalileanMutualEclipseGeometry, bool) {
	if search.memo == nil {
		return jupiterGalileanMutualEclipseGeometryAt(jde, search.first, search.second, search.inst)
	}
	bits := math.Float64bits(jde)
	for _, entry := range search.memo.entries {
		if entry.filled && entry.bits == bits {
			return entry.geometry, entry.valid
		}
	}
	geometry, valid := jupiterGalileanMutualEclipseGeometryAt(jde, search.first, search.second, search.inst)
	entry := &search.memo.entries[search.memo.next]
	entry.bits, entry.geometry, entry.filled, entry.valid = bits, geometry, true, valid
	search.memo.next = (search.memo.next + 1) % len(search.memo.entries)
	return geometry, valid
}

func (search jupiterGalileanEclipseSearch) gap(jde float64) (float64, bool) {
	geometry, ok := search.geometry(jde)
	return geometry.SeparationKM - geometry.PenumbraKM - geometry.PassiveRadiusKM, ok
}

func (search jupiterGalileanEclipseSearch) seasonAllows(jde float64) bool {
	frame := newJupiterGalileanEclipseFrame(jde, search.inst)
	sine := math.Max(-1, math.Min(1, -vectorDot(normalizeVector(frame.jupiterHelio), frame.pole)))
	return math.Abs(math.Asin(sine)*deg) <= jupiterGalileanMutualSeasonLimit(search.first, search.second)
}

func (search jupiterGalileanEclipseSearch) seekSeason(jde float64, direction int, end float64) (float64, bool) {
	for day := 1; float64(day) <= jupiterGalileanMutualSearchSpanDays; day++ {
		candidate := jde + float64(direction*day)*jupiterGalileanMutualSeasonJumpStepDays
		if direction > 0 && candidate-jupiterGalileanMutualSeasonJumpStepDays >= end {
			break
		}
		if search.seasonAllows(candidate) {
			return candidate - float64(direction)*jupiterGalileanMutualSeasonJumpStepDays, true
		}
	}
	return 0, false
}

func (search jupiterGalileanEclipseSearch) find(reception float64, direction int, includeCurrent, seasonGate bool, end float64) (JupiterGalileanMutualEvent, bool) {
	search.memo = &jupiterGalileanEclipseMemo{}
	firstEmission := jupiterGalileanEclipseEmission(reception, search.first, search.inst)
	secondEmission := jupiterGalileanEclipseEmission(reception, search.second, search.inst)
	current := math.Min(firstEmission, secondEmission)
	if direction < 0 {
		current = math.Max(firstEmission, secondEmission)
	}
	limit := current + float64(direction)*jupiterGalileanMutualSearchSpanDays
	geometricEnd := end
	if direction > 0 && isFinite(end) {
		geometricEnd = math.Max(jupiterGalileanEclipseEmission(end, search.first, search.inst), jupiterGalileanEclipseEmission(end, search.second, search.inst))
	}
	if event, found := search.eventAround(current); found {
		if mutualEventMatchesDirection(event.Greatest, reception, direction, includeCurrent) {
			if direction > 0 && event.Greatest >= end {
				return invalidJupiterGalileanMutualEvent(), false
			}
			return event, true
		}
		boundary := event.End
		if direction < 0 {
			boundary = event.Start
		}
		current = jupiterGalileanEclipseEmission(boundary, event.Passive, search.inst) + float64(direction)*jupiterGalileanMutualEventEpsilonDays
	}
	speed := jupiterGalileanMutualRelativeSpeedUpperBound(search.first, search.second) * jupiterGalileanEquatorialRadiusKM
	polarSpeed := jupiterGalileanMutualPolarSpeedUpperBound(search.first, search.second) * jupiterGalileanEquatorialRadiusKM
	for float64(direction)*(limit-current) > 0 && (direction < 0 || current < geometricEnd) {
		if seasonGate && !search.seasonAllows(current) {
			allowed, found := search.seekSeason(current, direction, geometricEnd)
			if !found || float64(direction)*(allowed-limit) > 0 {
				break
			}
			current = allowed
		}
		geometry, ok := search.geometry(current)
		gap := geometry.SeparationKM - geometry.PenumbraKM - geometry.PassiveRadiusKM
		if !ok {
			gap = 0
		}
		polarGap := geometry.PolarKM - geometry.PenumbraKM - geometry.PassiveRadiusKM
		step := jupiterGalileanMutualOuterStep(gap, polarGap, speed, polarSpeed, ok && geometry.PolarAllowed)
		next := current + float64(direction)*step
		if ok && geometry.PolarAllowed {
			endpoint, endpointOK := search.geometry(next)
			if !endpointOK || !endpoint.PolarAllowed {
				step = jupiterGalileanMutualOuterStep(gap, 0, speed, polarSpeed, false)
				next = current + float64(direction)*step
			}
		}
		if float64(direction)*(next-limit) > 0 {
			next = limit
		}
		left, right := math.Min(current, next), math.Max(current, next)
		if hit, found := search.hit(left, right, direction); found {
			if event, valid := search.eventAround(hit); valid {
				if mutualEventMatchesDirection(event.Greatest, reception, direction, includeCurrent) {
					if direction > 0 && event.Greatest >= end {
						return invalidJupiterGalileanMutualEvent(), false
					}
					return event, true
				}
				boundary := event.End
				if direction < 0 {
					boundary = event.Start
				}
				current = jupiterGalileanEclipseEmission(boundary, event.Passive, search.inst) + float64(direction)*jupiterGalileanMutualEventEpsilonDays
				continue
			}
		}
		current = next
	}
	return invalidJupiterGalileanMutualEvent(), false
}

func (search jupiterGalileanEclipseSearch) hit(left, right float64, direction int) (float64, bool) {
	leftGeometry, leftOK := search.geometry(left)
	rightGeometry, rightOK := search.geometry(right)
	leftGap := leftGeometry.SeparationKM - leftGeometry.PenumbraKM - leftGeometry.PassiveRadiusKM
	rightGap := rightGeometry.SeparationKM - rightGeometry.PenumbraKM - rightGeometry.PassiveRadiusKM
	if !leftOK {
		leftGap = 0
	}
	if !rightOK {
		rightGap = 0
	}
	if direction > 0 {
		if leftGap < 0 {
			return left, true
		}
		if rightGap < 0 {
			return right, true
		}
	} else {
		if rightGap < 0 {
			return right, true
		}
		if leftGap < 0 {
			return left, true
		}
	}
	speed := jupiterGalileanMutualRelativeSpeedUpperBound(search.first, search.second) * jupiterGalileanEquatorialRadiusKM
	if (leftGap+rightGap-speed*(right-left))/2 > 0 {
		return 0, false
	}
	polarSpeed := jupiterGalileanMutualPolarSpeedUpperBound(search.first, search.second) * jupiterGalileanEquatorialRadiusKM
	if jupiterGalileanMutualPolarExcludes(leftGeometry.PolarKM-leftGeometry.PenumbraKM-leftGeometry.PassiveRadiusKM, rightGeometry.PolarKM-rightGeometry.PenumbraKM-rightGeometry.PassiveRadiusKM, polarSpeed, right-left, leftOK && leftGeometry.PolarAllowed, rightOK && rightGeometry.PolarAllowed) {
		return 0, false
	}
	if right-left <= jupiterGalileanMutualMinimumIntervalDays {
		minimum := search.minimum(left, right)
		gap, ok := search.gap(minimum)
		return minimum, ok && gap < 0
	}
	middle := (left + right) / 2
	if direction > 0 {
		if hit, found := search.hit(left, middle, direction); found {
			return hit, true
		}
		return search.hit(middle, right, direction)
	}
	if hit, found := search.hit(middle, right, direction); found {
		return hit, true
	}
	return search.hit(left, middle, direction)
}

func (search jupiterGalileanEclipseSearch) minimum(left, right float64) float64 {
	const golden = 0.6180339887498948482
	firstTime, secondTime := right-golden*(right-left), left+golden*(right-left)
	first, _ := search.geometry(firstTime)
	second, _ := search.geometry(secondTime)
	for iteration := 0; iteration < 80; iteration++ {
		if first.SeparationKM <= second.SeparationKM {
			right, secondTime, second = secondTime, firstTime, first
			firstTime = right - golden*(right-left)
			first, _ = search.geometry(firstTime)
		} else {
			left, firstTime, first = firstTime, secondTime, second
			secondTime = left + golden*(right-left)
			second, _ = search.geometry(secondTime)
		}
	}
	return (left + right) / 2
}

func (search jupiterGalileanEclipseSearch) boundary(seed float64, direction int) (float64, bool) {
	inside := seed
	for minute := 1; minute <= 360; minute++ {
		outside := seed + float64(direction*minute)*jupiterGalileanMutualBoundaryStepDays
		gap, ok := search.gap(outside)
		if !ok {
			return 0, false
		}
		if gap > 0 {
			for iteration := 0; iteration < 80; iteration++ {
				middle := (inside + outside) / 2
				value, valid := search.gap(middle)
				if !valid {
					return 0, false
				}
				if value > 0 {
					outside = middle
				} else {
					inside = middle
				}
			}
			return (inside + outside) / 2, true
		}
		inside = outside
	}
	return 0, false
}

func (search jupiterGalileanEclipseSearch) eventAround(seed float64) (JupiterGalileanMutualEvent, bool) {
	geometry, valid := search.geometry(seed)
	if !valid || geometry.SeparationKM >= geometry.PenumbraKM+geometry.PassiveRadiusKM {
		return invalidJupiterGalileanMutualEvent(), false
	}
	start, startOK := search.boundary(seed, -1)
	end, endOK := search.boundary(seed, 1)
	if !startOK || !endOK || !(start < end) {
		return invalidJupiterGalileanMutualEvent(), false
	}
	bestTime, bestSeparation := start, math.Inf(1)
	for sample := 0; ; sample++ {
		candidate := math.Min(end, start+float64(sample)*jupiterGalileanMutualMinimumIntervalDays)
		value, ok := search.geometry(candidate)
		if !ok || value.Active != geometry.Active {
			return invalidJupiterGalileanMutualEvent(), false
		}
		if value.SeparationKM < bestSeparation {
			bestTime, bestSeparation = candidate, value.SeparationKM
		}
		if candidate == end {
			break
		}
	}
	greatest := search.minimum(math.Max(start, bestTime-jupiterGalileanMutualMinimumIntervalDays), math.Min(end, bestTime+jupiterGalileanMutualMinimumIntervalDays))
	geometry, valid = search.geometry(greatest)
	startGeometry, _ := search.geometry(start)
	endGeometry, _ := search.geometry(end)
	if !valid || startGeometry.Active != geometry.Active || endGeometry.Active != geometry.Active {
		return invalidJupiterGalileanMutualEvent(), false
	}
	start = jupiterGalileanEclipseReception(start, geometry.Passive, search.inst)
	end = jupiterGalileanEclipseReception(end, geometry.Passive, search.inst)
	greatest = jupiterGalileanEclipseReception(greatest, geometry.Passive, search.inst)
	context := newJupiterGalileanObservationContext(greatest, search.inst)
	event := JupiterGalileanMutualEvent{
		Valid: true, Type: JupiterGalileanMutualEclipse, Active: geometry.Active, Passive: geometry.Passive,
		Start: start, Greatest: greatest, End: end, DurationDays: end - start,
		MinimumSeparationArcsec: geometry.MinimumSeparationArcsec, ImpactParameter: geometry.ImpactParameter,
		Coverage: geometry.Coverage, Annular: geometry.Annular,
	}
	for index := range event.GreatestState {
		event.GreatestState[index] = context.phenomenonForSatellite(index)
	}
	if !isFinite(start) || !isFinite(end) || !isFinite(greatest) || !isFinite(event.MinimumSeparationArcsec) || !isFinite(event.ImpactParameter) || !isFinite(event.Coverage) {
		return invalidJupiterGalileanMutualEvent(), false
	}
	return event, true
}

func invalidJupiterGalileanMutualEvent() JupiterGalileanMutualEvent {
	return JupiterGalileanMutualEvent{}
}
