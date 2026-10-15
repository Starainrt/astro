package basic

import (
	"math"
	"sort"
	"time"
)

const (
	planetOccultationDefaultStepDays        = 0.25
	planetOccultationCandidateLimitArcsec   = 10 * 3600.0
	planetOccultationLatitudeMarginArcsec   = 3600.0
	planetOccultationContactStepDays        = 10.0 / 1440.0
	planetOccultationContactSpanDays        = 2.0
	planetOccultationGrazingToleranceArcsec = 0.01
	planetOccultationRootToleranceDays      = occultationEventSelectionToleranceDays
	planetOccultationMaxContactSteps        = 10000
)

type planetOccultationConfig struct {
	planet             OccultationPlanet
	planetIndex        int
	equatorialRadiusKM float64
	apparentRaDecN     func(float64, int) (float64, float64)
	earthDistanceN     func(float64, int) float64
	semidiameterN      func(float64, int) float64
}

type planetMoonPosition struct {
	moonRA, moonDec     float64
	planetRA, planetDec float64
	valid               bool
}

type planetOccultationState struct {
	position              planetMoonPosition
	separationArcsec      float64
	moonSemidiameter      float64
	planetSemidiameter    float64
	externalContactMetric float64
	internalContactMetric float64
	valid                 bool
}

func (state planetOccultationState) movingDiskContactState() movingDiskContactState {
	return movingDiskContactState{
		separation:           state.separationArcsec,
		occultingOuterRadius: state.moonSemidiameter,
		occultingInnerRadius: state.moonSemidiameter,
		targetRadius:         state.planetSemidiameter,
		valid:                state.valid,
	}
}

// FindPlanetOccultationsWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func FindPlanetOccultationsWithInstance(start, end time.Time, planet OccultationPlanet, longitude, latitude, height float64,
	options OccultationSearchOptions, inst *Instance) ([]PlanetOccultationInfo, error) {
	if err := validateOccultationTimeRange(start, end); err != nil {
		return nil, err
	}
	if err := planet.Validate(); err != nil {
		return nil, err
	}
	observer := Observer{Longitude: longitude, Latitude: latitude, Height: height}
	if err := observer.Validate(); err != nil {
		return nil, err
	}
	if err := options.Validate(); err != nil {
		return nil, err
	}

	config, _ := planetOccultationConfigFor(planet, inst)
	startTT := occultationTimeToTT(start)
	endTT := occultationTimeToTT(end)
	resultLocation := start.Location()
	candidates := planetOccultationCandidateGreatestTimes(
		startTT, endTT, planetOccultationCoarseStepDays(options), config, &observer, options.SafetyMarginArcsec, inst,
	)
	results := make([]PlanetOccultationInfo, 0, len(candidates))
	for _, greatestTT := range candidates {
		info, ok := planetOccultationInfoWithLimb(greatestTT, config, observer, options.SafetyMarginArcsec, resultLocation, options.LunarLimb, inst)
		if !ok {
			continue
		}
		if len(results) == 0 || math.Abs(results[len(results)-1].Greatest.Sub(info.Greatest).Seconds()) > 60 {
			results = append(results, info)
			if options.MaxEvents > 0 && len(results) >= options.MaxEvents {
				break
			}
		}
	}
	sort.SliceStable(results, func(i, j int) bool { return results[i].Greatest.Before(results[j].Greatest) })
	return results, nil
}

// FindBestPlanetOccultationsWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func FindBestPlanetOccultationsWithInstance(start, end time.Time, planet OccultationPlanet,
	options OccultationSearchOptions, inst *Instance) ([]PlanetOccultationInfo, error) {
	if err := validateOccultationTimeRange(start, end); err != nil {
		return nil, err
	}
	if err := planet.Validate(); err != nil {
		return nil, err
	}
	if err := options.Validate(); err != nil {
		return nil, err
	}

	config, _ := planetOccultationConfigFor(planet, inst)
	startTT := occultationTimeToTT(start)
	endTT := occultationTimeToTT(end)
	selectionStartTT := startTT - occultationEventSelectionToleranceDays
	selectionEndTT := endTT + occultationEventSelectionToleranceDays
	candidateStartTT := startTT - occultationPathSearchSpanDays
	candidateEndTT := endTT + occultationPathSearchSpanDays
	resultLocation := start.Location()
	candidates := planetOccultationCandidateGreatestTimes(
		candidateStartTT, candidateEndTT, planetOccultationCoarseStepDays(options), config, nil, options.SafetyMarginArcsec, inst,
	)
	results := make([]PlanetOccultationInfo, 0, len(candidates))
	for _, seedTT := range candidates {
		greatestTT, observer, _, observerOK := planetOccultationBestObserver(seedTT, selectionStartTT, selectionEndTT, config, inst)
		if !observerOK {
			continue
		}
		info, ok := planetOccultationInfoWithLimb(greatestTT, config, observer, options.SafetyMarginArcsec, resultLocation, options.LunarLimb, inst)
		if !ok {
			continue
		}
		if len(results) == 0 || math.Abs(results[len(results)-1].Greatest.Sub(info.Greatest).Seconds()) > 60 {
			results = append(results, info)
			if options.MaxEvents > 0 && len(results) >= options.MaxEvents {
				break
			}
		}
	}
	sort.SliceStable(results, func(i, j int) bool { return results[i].Greatest.Before(results[j].Greatest) })
	return results, nil
}

func planetOccultationConfigFor(planet OccultationPlanet, instances ...*Instance) (planetOccultationConfig, bool) {
	inst := instanceValue(instances)

	config := planetOccultationConfig{planet: planet}
	switch planet {
	case OccultationMercury:
		config.planetIndex = 1
		config.equatorialRadiusKM = mercuryEquatorialRadiusKM
		config.apparentRaDecN = func(jde float64, n int) (float64, float64) { return MercuryApparentRaDecNWithInstance(jde, n, inst) }
		config.earthDistanceN = func(jd float64, n int) float64 { return EarthMercuryAwayNWithInstance(jd, n, inst) }
		config.semidiameterN = func(jde float64, n int) float64 { return MercurySemidiameterNWithInstance(jde, n, inst) }
	case OccultationVenus:
		config.planetIndex = 2
		config.equatorialRadiusKM = venusEquatorialRadiusKM
		config.apparentRaDecN = func(jd float64, n int) (float64, float64) { return VenusApparentRaDecNWithInstance(jd, n, inst) }
		config.earthDistanceN = func(jd float64, n int) float64 { return EarthVenusAwayNWithInstance(jd, n, inst) }
		config.semidiameterN = func(jde float64, n int) float64 { return VenusSemidiameterNWithInstance(jde, n, inst) }
	case OccultationMars:
		config.planetIndex = 3
		config.equatorialRadiusKM = marsEquatorialRadiusKM
		config.apparentRaDecN = func(jd float64, n int) (float64, float64) { return MarsApparentRaDecNWithInstance(jd, n, inst) }
		config.earthDistanceN = func(jd float64, n int) float64 { return EarthMarsAwayNWithInstance(jd, n, inst) }
		config.semidiameterN = func(jde float64, n int) float64 { return MarsSemidiameterNWithInstance(jde, n, inst) }
	case OccultationJupiter:
		config.planetIndex = 4
		config.equatorialRadiusKM = jupiterEquatorialRadiusKM
		config.apparentRaDecN = func(jd float64, n int) (float64, float64) { return JupiterApparentRaDecNWithInstance(jd, n, inst) }
		config.earthDistanceN = func(jd float64, n int) float64 { return EarthJupiterAwayNWithInstance(jd, n, inst) }
		config.semidiameterN = func(jde float64, n int) float64 { return JupiterSemidiameterNWithInstance(jde, n, inst) }
	case OccultationSaturn:
		config.planetIndex = 5
		config.equatorialRadiusKM = saturnEquatorialRadiusKM
		config.apparentRaDecN = func(jd float64, n int) (float64, float64) { return SaturnApparentRaDecNWithInstance(jd, n, inst) }
		config.earthDistanceN = func(jd float64, n int) float64 { return EarthSaturnAwayNWithInstance(jd, n, inst) }
		config.semidiameterN = func(jde float64, n int) float64 { return SaturnSemidiameterNWithInstance(jde, n, inst) }
	case OccultationUranus:
		config.planetIndex = 6
		config.equatorialRadiusKM = uranusEquatorialRadiusKM
		config.apparentRaDecN = func(jd float64, n int) (float64, float64) { return UranusApparentRaDecNWithInstance(jd, n, inst) }
		config.earthDistanceN = func(jd float64, n int) float64 { return EarthUranusAwayNWithInstance(jd, n, inst) }
		config.semidiameterN = func(jde float64, n int) float64 { return UranusSemidiameterNWithInstance(jde, n, inst) }
	case OccultationNeptune:
		config.planetIndex = 7
		config.equatorialRadiusKM = neptuneEquatorialRadiusKM
		config.apparentRaDecN = func(jd float64, n int) (float64, float64) { return NeptuneApparentRaDecNWithInstance(jd, n, inst) }
		config.earthDistanceN = func(jd float64, n int) float64 { return EarthNeptuneAwayNWithInstance(jd, n, inst) }
		config.semidiameterN = func(jde float64, n int) float64 { return NeptuneSemidiameterNWithInstance(jde, n, inst) }
	default:
		return planetOccultationConfig{}, false
	}
	return config, true
}

func planetOccultationCandidateGreatestTimes(
	startTT, endTT, step float64,
	config planetOccultationConfig,
	observer *Observer,
	safetyMarginArcsec float64, instances ...*Instance) []float64 {
	inst := instanceValue(instances)

	if endTT <= startTT {
		return nil
	}
	duration := endTT - startTT
	if step > duration/4 {
		step = math.Max(duration/4, 0.25/86400.0)
	}
	step = math.Max(step, 0.25/86400.0)
	scanStart := startTT - step
	scanEnd := endTT + step

	leftTT := scanStart
	centerTT := math.Min(leftTT+step, scanEnd)
	leftValue := planetOccultationExternalContactMetric(leftTT, config, observer, 8, inst)
	centerValue := planetOccultationExternalContactMetric(centerTT, config, observer, 8, inst)
	results := make([]float64, 0)
	for centerTT < scanEnd {
		rightTT := math.Min(centerTT+step, scanEnd)
		rightValue := planetOccultationExternalContactMetric(rightTT, config, observer, 8, inst)
		candidateLimit := planetOccultationCandidateLimitArcsec + safetyMarginArcsec
		if finite(leftValue) && finite(centerValue) && finite(rightValue) &&
			centerValue <= leftValue && centerValue <= rightValue && centerValue <= candidateLimit {
			// 最小外接触度量同时包含动态月面和行星盘面，因此定义事件是否存在以及报告的掩甚时刻。
			// The minimum outer-contact metric includes both dynamic disks and therefore defines event existence and the reported greatest instant.
			greatestTT := planetOccultationMinimizeExternalMetric(leftTT, rightTT, config, observer, inst)
			if greatestTT >= startTT && greatestTT <= endTT &&
				planetOccultationLatitudePass(greatestTT, config, observer, safetyMarginArcsec, inst) {
				if len(results) == 0 || math.Abs(greatestTT-results[len(results)-1]) > 60.0/86400.0 {
					results = append(results, greatestTT)
				}
			}
		}
		leftTT, leftValue = centerTT, centerValue
		centerTT, centerValue = rightTT, rightValue
	}
	return results
}

func planetOccultationMinimizeExternalMetric(left, right float64, config planetOccultationConfig, observer *Observer, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	if right <= left {
		return left
	}
	const goldenRatio = 0.6180339887498949
	x1 := right - goldenRatio*(right-left)
	x2 := left + goldenRatio*(right-left)
	f1 := planetOccultationExternalContactMetric(x1, config, observer, -1, inst)
	f2 := planetOccultationExternalContactMetric(x2, config, observer, -1, inst)
	for i := 0; i < 64 && right-left > planetOccultationRootToleranceDays; i++ {
		if f1 > f2 {
			left = x1
			x1, f1 = x2, f2
			x2 = left + goldenRatio*(right-left)
			f2 = planetOccultationExternalContactMetric(x2, config, observer, -1, inst)
		} else {
			right = x2
			x2, f2 = x1, f1
			x1 = right - goldenRatio*(right-left)
			f1 = planetOccultationExternalContactMetric(x1, config, observer, -1, inst)
		}
	}
	return (left + right) / 2
}

func planetOccultationBestObserver(seedTT, startTT, endTT float64, config planetOccultationConfig, instances ...*Instance) (float64, Observer, float64, bool) {
	inst := instanceValue(instances)

	frameAt := func(tt float64) (occultationPathFrame, bool) {
		return planetOccultationPathFrameAt(tt, config, inst)
	}
	searchStart := seedTT - occultationPathSearchSpanDays
	searchEnd := seedTT + occultationPathSearchSpanDays
	outerStart, outerEnd, ok := occultationPathWindowForFrame(seedTT, searchStart, searchEnd, frameAt, false)
	if !ok {
		return 0, Observer{}, 0, false
	}
	greatestTT := occultationPathGreatestForFrame(seedTT, outerStart, outerEnd, frameAt)
	if greatestTT < startTT || greatestTT > endTT {
		return 0, Observer{}, 0, false
	}
	point, pointOK := occultationPathCenterPointForFrame(greatestTT, frameAt, time.UTC)
	if !pointOK {
		// 非中心事件的月影轴不与地球椭球相交，最佳站心应取离影轴最近的椭球点；
		// 外接触切点位于掩带边缘，会把掩甚站心推出掩食之外，只作最后回退。
		// For a non-central event the shadow axis misses the ellipsoid, so the
		// best station is the ellipsoid point nearest to that axis. The outer
		// contact tangent sits on the band edge and moves the greatest station
		// outside the occultation, so it remains the last resort only.
		point, pointOK = occultationPathTrackPointForFrame(greatestTT, frameAt, time.UTC)
		if !pointOK {
			point, pointOK = occultationPathBoundaryPointForFrame(greatestTT, frameAt, time.UTC, inst)
		}
	}
	if !pointOK {
		return 0, Observer{}, 0, false
	}
	observer := Observer{Longitude: point.Longitude, Latitude: point.Latitude}
	state := planetOccultationStateAt(greatestTT, config, &observer, -1, inst)
	if !state.valid {
		return 0, Observer{}, 0, false
	}
	return greatestTT, observer, state.externalContactMetric, true
}

func planetOccultationInfoAtGreatest(
	greatestTT float64,
	config planetOccultationConfig,
	observer Observer,
	safetyMarginArcsec float64,
	location *time.Location, instances ...*Instance) (PlanetOccultationInfo, bool) {
	inst := instanceValue(instances)

	info, state, ok := planetOccultationGreatestInfo(greatestTT, config, observer, safetyMarginArcsec, location, inst)
	if !ok || state.externalContactMetric > 0 {
		return PlanetOccultationInfo{}, false
	}

	if math.Abs(state.externalContactMetric) <= planetOccultationGrazingToleranceArcsec {
		info.Type = OccultationGrazing
		info.ExternalImmersion = info.Greatest
		info.ExternalEmersion = info.Greatest
		info.ContactsComplete = true
		return info, true
	}
	return planetOccultationCircularContacts(info, state, greatestTT, config, observer, location, inst)
}

// planetOccultationGreatestInfo 只填掩甚几何，不判断圆月面是否成掩、不求接触。
func planetOccultationGreatestInfo(greatestTT float64, config planetOccultationConfig, observer Observer, safetyMarginArcsec float64, location *time.Location, inst *Instance) (PlanetOccultationInfo, planetOccultationState, bool) {
	if !planetOccultationLatitudePass(greatestTT, config, &observer, safetyMarginArcsec, inst) {
		return PlanetOccultationInfo{}, planetOccultationState{}, false
	}
	state := planetOccultationStateAt(greatestTT, config, &observer, -1, inst)
	if !state.valid {
		return PlanetOccultationInfo{}, planetOccultationState{}, false
	}
	info := PlanetOccultationInfo{
		Planet:                   config.planet,
		TargetID:                 config.planet.String(),
		Observer:                 observer,
		Type:                     OccultationPartial,
		Greatest:                 occultationTTToLocation(greatestTT, location),
		MinimumSeparationArcsec:  state.separationArcsec,
		PositionAngleDeg:         occultationPositionAngle(state.position.moonRA, state.position.moonDec, state.position.planetRA, state.position.planetDec),
		MoonSemidiameterArcsec:   state.moonSemidiameter,
		PlanetSemidiameterArcsec: state.planetSemidiameter,
		MoonAltitudeAtGreatest:   occultationAltitude(greatestTT, observer, state.position.moonRA, state.position.moonDec),
		MoonAzimuthAtGreatest:    occultationAzimuth(greatestTT, observer, state.position.moonRA, state.position.moonDec),
	}
	info.VisibleAtGreatest = info.MoonAltitudeAtGreatest >= 0
	return info, state, true
}

func planetOccultationCircularContacts(info PlanetOccultationInfo, state planetOccultationState, greatestTT float64, config planetOccultationConfig, observer Observer, location *time.Location, inst *Instance) (PlanetOccultationInfo, bool) {
	contactEvaluator := newPlanetOccultationContactEvaluator(config, observer, inst)
	contactEvaluator.prime(greatestTT, state.movingDiskContactState(), state.valid)
	externalImmersionTT, externalImmersionOK := planetOccultationContactWithEvaluator(greatestTT, -1, false, contactEvaluator)
	externalEmersionTT, externalEmersionOK := planetOccultationContactWithEvaluator(greatestTT, 1, false, contactEvaluator)
	if !externalImmersionOK || !externalEmersionOK || externalEmersionTT <= externalImmersionTT {
		return PlanetOccultationInfo{}, false
	}
	info.ExternalImmersion = occultationTTToLocation(externalImmersionTT, location)
	info.ExternalEmersion = occultationTTToLocation(externalEmersionTT, location)

	if state.internalContactMetric < -planetOccultationGrazingToleranceArcsec {
		internalImmersionTT, internalImmersionOK := planetOccultationContactWithEvaluator(greatestTT, -1, true, contactEvaluator)
		internalEmersionTT, internalEmersionOK := planetOccultationContactWithEvaluator(greatestTT, 1, true, contactEvaluator)
		if !internalImmersionOK || !internalEmersionOK ||
			internalImmersionTT <= externalImmersionTT || internalEmersionTT >= externalEmersionTT ||
			internalImmersionTT >= greatestTT || internalEmersionTT <= greatestTT {
			return PlanetOccultationInfo{}, false
		}
		info.Type = OccultationTotal
		info.InternalImmersion = occultationTTToLocation(internalImmersionTT, location)
		info.InternalEmersion = occultationTTToLocation(internalEmersionTT, location)
		info.HasInternalContacts = true
	}
	info.ContactsComplete = true
	return info, true
}

func planetMoonPositionAt(tt float64, config planetOccultationConfig, observer *Observer, n int, instances ...*Instance) planetMoonPosition {
	inst := instanceValue(instances)

	moonRA, moonDec := HMoonGeocentricApparentRaDecNWithInstance(tt, n, inst)
	planetRA, planetDec := config.apparentRaDecN(tt, n)
	if observer != nil {
		ut := TT2UTC(tt)
		moonDistanceAU := HMoonAwayNWithInstance(tt, n, inst) / angularDiameterAstronomicalUnitKM
		planetDistanceAU := config.earthDistanceN(tt, n)
		moonRA, moonDec = TopocentricRaDec(moonRA, moonDec, observer.Latitude, observer.Longitude, ut, moonDistanceAU, observer.Height)
		planetRA, planetDec = TopocentricRaDec(planetRA, planetDec, observer.Latitude, observer.Longitude, ut, planetDistanceAU, observer.Height)
		moonRA = normalizeRA(moonRA)
		planetRA = normalizeRA(planetRA)
	}
	return planetMoonPosition{
		moonRA:    moonRA,
		moonDec:   moonDec,
		planetRA:  planetRA,
		planetDec: planetDec,
		valid:     finite(moonRA) && finite(moonDec) && finite(planetRA) && finite(planetDec),
	}
}

func planetOccultationStateAt(tt float64, config planetOccultationConfig, observer *Observer, n int, instances ...*Instance) planetOccultationState {
	inst := instanceValue(instances)

	position := planetMoonPositionAt(tt, config, observer, n, inst)
	moonRadius := MoonSemidiameterNWithInstance(tt, n, inst)
	planetRadius := config.semidiameterN(tt, n)
	if observer != nil {
		moonRadius = moonTopocentricSemidiameterN(tt, *observer, n, inst)
		planetRadius = planetTopocentricSemidiameterN(tt, config, *observer, n)
	}
	if !position.valid || !finite(moonRadius) || !finite(planetRadius) || moonRadius <= planetRadius || planetRadius <= 0 {
		return planetOccultationState{}
	}
	separation := angularSeparationDegrees(position.moonRA, position.moonDec, position.planetRA, position.planetDec) * 3600
	contactState := movingDiskContactState{
		separation:           separation,
		occultingOuterRadius: moonRadius,
		occultingInnerRadius: moonRadius,
		targetRadius:         planetRadius,
		valid: movingDiskContactStateValid(
			separation, moonRadius, moonRadius, planetRadius,
		),
	}
	return planetOccultationState{
		position:              position,
		separationArcsec:      separation,
		moonSemidiameter:      moonRadius,
		planetSemidiameter:    planetRadius,
		externalContactMetric: contactState.externalContactGap(),
		internalContactMetric: contactState.internalContactGap(),
		valid:                 contactState.valid,
	}
}

func planetTopocentricSemidiameterN(tt float64, config planetOccultationConfig, observer Observer, n int) float64 {
	ra, dec := config.apparentRaDecN(tt, n)
	distanceKM := config.earthDistanceN(tt, n) * angularDiameterAstronomicalUnitKM
	if !finite(ra) || !finite(dec) || !finite(distanceKM) || distanceKM <= 0 {
		return math.NaN()
	}
	distanceKM = topocentricDistanceKM(ra, dec, distanceKM, observer, TT2UTC(tt))
	if !finite(distanceKM) || distanceKM <= config.equatorialRadiusKM {
		return math.NaN()
	}
	return angularSemidiameterArcsec(config.equatorialRadiusKM, distanceKM)
}

func planetOccultationExternalContactMetric(tt float64, config planetOccultationConfig, observer *Observer, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	state := planetOccultationStateAt(tt, config, observer, n, inst)
	if !state.valid {
		return math.Inf(1)
	}
	return state.externalContactMetric
}

func planetOccultationLatitudePass(tt float64, config planetOccultationConfig, observer *Observer, safetyMarginArcsec float64, instances ...*Instance) bool {
	inst := instanceValue(instances)

	state := planetOccultationStateAt(tt, config, observer, -1, inst)
	if !state.valid {
		return false
	}
	_, moonLatitude := RaDecToLoBo(tt, state.position.moonRA, state.position.moonDec)
	_, planetLatitude := RaDecToLoBo(tt, state.position.planetRA, state.position.planetDec)
	limit := state.moonSemidiameter + state.planetSemidiameter + planetOccultationLatitudeMarginArcsec + safetyMarginArcsec
	return math.Abs(moonLatitude-planetLatitude)*3600 <= limit
}

func newPlanetOccultationContactEvaluator(
	config planetOccultationConfig,
	observer Observer, instances ...*Instance) *movingDiskContactEvaluator {
	inst := instanceValue(instances)

	return newMovingDiskContactEvaluator(func(tt float64) (movingDiskContactState, bool) {
		state := planetOccultationStateAt(tt, config, &observer, -1, inst)
		return state.movingDiskContactState(), state.valid
	})
}

func planetOccultationContactWithEvaluator(
	greatestTT float64,
	direction int,
	internal bool,
	evaluator *movingDiskContactEvaluator,
) (float64, bool) {
	if direction != -1 && direction != 1 {
		return math.NaN(), false
	}
	metric := func(tt float64) (float64, bool) {
		return evaluator.gap(tt, internal)
	}
	root, ok := occultationMovingDiskEngine().contactRoot(
		greatestTT,
		float64(direction),
		planetOccultationContactStepDays,
		planetOccultationContactSpanDays,
		planetOccultationRootToleranceDays,
		metric,
		64,
	)
	if !ok {
		return math.NaN(), false
	}
	return root, true
}

func planetOccultationCoarseStepDays(options OccultationSearchOptions) float64 {
	step := planetOccultationDefaultStepDays
	if options.MaxStep > 0 {
		requested := options.MaxStep.Hours() / 24
		if requested > 0 && requested < step {
			step = requested
		}
	}
	return math.Max(step, occultationSearchMinimumStep.Hours()/24)
}
