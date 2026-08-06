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

// FindPlanetOccultations 搜索固定观测点的有限盘面行星月掩。
// 经度东为正、纬度北为正，单位为度；高度为平均海平面以上米数。目标位置、视差和视半径会在每次候选、掩甚和接触计算时重新计算。
// FindPlanetOccultations searches one finite-disk planet at a fixed observing site.
// Longitude is east-positive in degrees, latitude is north-positive in degrees, and height is the observer elevation above mean sea level in meters. The target position, parallax, and semidiameter are recomputed at every candidate, greatest, and contact evaluation.
func FindPlanetOccultations(start, end time.Time, planet OccultationPlanet, longitude, latitude, height float64,
	options OccultationSearchOptions) ([]PlanetOccultationInfo, error) {
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

	config, _ := planetOccultationConfigFor(planet)
	startTT := occultationTimeToTT(start)
	endTT := occultationTimeToTT(end)
	resultLocation := start.Location()
	candidates := planetOccultationCandidateGreatestTimes(
		startTT, endTT, planetOccultationCoarseStepDays(options), config, &observer, options.SafetyMarginArcsec,
	)
	results := make([]PlanetOccultationInfo, 0, len(candidates))
	for _, greatestTT := range candidates {
		info, ok := planetOccultationInfoAtGreatest(greatestTT, config, observer, options.SafetyMarginArcsec, resultLocation)
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

// FindBestPlanetOccultations 返回窗口内每次有限盘面行星月掩的全球海平面几何掩甚点。
// 地心数据只用于搜索初值；返回位置与 FindPlanetOccultationPaths 一致，地平线可见性只报告、不参与点选择。距离查询端点 10 ms 内的掩甚时刻也会包含，与数值根精度一致。
// FindBestPlanetOccultations returns the global geometric greatest point at sea level for every finite-disk planetary occultation in the window.
// Geocentric data only seeds the search. The returned location matches FindPlanetOccultationPaths; horizon visibility is reported but does not select the point. A greatest instant within 10 ms of either query endpoint is included, matching the numerical root precision.
func FindBestPlanetOccultations(start, end time.Time, planet OccultationPlanet,
	options OccultationSearchOptions) ([]PlanetOccultationInfo, error) {
	if err := validateOccultationTimeRange(start, end); err != nil {
		return nil, err
	}
	if err := planet.Validate(); err != nil {
		return nil, err
	}
	if err := options.Validate(); err != nil {
		return nil, err
	}

	config, _ := planetOccultationConfigFor(planet)
	startTT := occultationTimeToTT(start)
	endTT := occultationTimeToTT(end)
	selectionStartTT := startTT - occultationEventSelectionToleranceDays
	selectionEndTT := endTT + occultationEventSelectionToleranceDays
	candidateStartTT := startTT - occultationPathSearchSpanDays
	candidateEndTT := endTT + occultationPathSearchSpanDays
	resultLocation := start.Location()
	candidates := planetOccultationCandidateGreatestTimes(
		candidateStartTT, candidateEndTT, planetOccultationCoarseStepDays(options), config, nil, options.SafetyMarginArcsec,
	)
	results := make([]PlanetOccultationInfo, 0, len(candidates))
	for _, seedTT := range candidates {
		greatestTT, observer, _, observerOK := planetOccultationBestObserver(seedTT, selectionStartTT, selectionEndTT, config)
		if !observerOK {
			continue
		}
		info, ok := planetOccultationInfoAtGreatest(greatestTT, config, observer, options.SafetyMarginArcsec, resultLocation)
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

func planetOccultationConfigFor(planet OccultationPlanet) (planetOccultationConfig, bool) {
	config := planetOccultationConfig{planet: planet}
	switch planet {
	case OccultationMercury:
		config.equatorialRadiusKM = mercuryEquatorialRadiusKM
		config.apparentRaDecN = MercuryApparentRaDecN
		config.earthDistanceN = EarthMercuryAwayN
		config.semidiameterN = MercurySemidiameterN
	case OccultationVenus:
		config.equatorialRadiusKM = venusEquatorialRadiusKM
		config.apparentRaDecN = VenusApparentRaDecN
		config.earthDistanceN = EarthVenusAwayN
		config.semidiameterN = VenusSemidiameterN
	case OccultationMars:
		config.equatorialRadiusKM = marsEquatorialRadiusKM
		config.apparentRaDecN = MarsApparentRaDecN
		config.earthDistanceN = EarthMarsAwayN
		config.semidiameterN = MarsSemidiameterN
	case OccultationJupiter:
		config.equatorialRadiusKM = jupiterEquatorialRadiusKM
		config.apparentRaDecN = JupiterApparentRaDecN
		config.earthDistanceN = EarthJupiterAwayN
		config.semidiameterN = JupiterSemidiameterN
	case OccultationSaturn:
		config.equatorialRadiusKM = saturnEquatorialRadiusKM
		config.apparentRaDecN = SaturnApparentRaDecN
		config.earthDistanceN = EarthSaturnAwayN
		config.semidiameterN = SaturnSemidiameterN
	case OccultationUranus:
		config.equatorialRadiusKM = uranusEquatorialRadiusKM
		config.apparentRaDecN = UranusApparentRaDecN
		config.earthDistanceN = EarthUranusAwayN
		config.semidiameterN = UranusSemidiameterN
	case OccultationNeptune:
		config.equatorialRadiusKM = neptuneEquatorialRadiusKM
		config.apparentRaDecN = NeptuneApparentRaDecN
		config.earthDistanceN = EarthNeptuneAwayN
		config.semidiameterN = NeptuneSemidiameterN
	default:
		return planetOccultationConfig{}, false
	}
	return config, true
}

func planetOccultationCandidateGreatestTimes(
	startTT, endTT, step float64,
	config planetOccultationConfig,
	observer *Observer,
	safetyMarginArcsec float64,
) []float64 {
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
	leftValue := planetOccultationExternalContactMetric(leftTT, config, observer, 8)
	centerValue := planetOccultationExternalContactMetric(centerTT, config, observer, 8)
	results := make([]float64, 0)
	for centerTT < scanEnd {
		rightTT := math.Min(centerTT+step, scanEnd)
		rightValue := planetOccultationExternalContactMetric(rightTT, config, observer, 8)
		candidateLimit := planetOccultationCandidateLimitArcsec + safetyMarginArcsec
		if finite(leftValue) && finite(centerValue) && finite(rightValue) &&
			centerValue <= leftValue && centerValue <= rightValue && centerValue <= candidateLimit {
			// 最小外接触度量同时包含动态月面和行星盘面，因此定义事件是否存在以及报告的掩甚时刻。
			// The minimum outer-contact metric includes both dynamic disks and therefore defines event existence and the reported greatest instant.
			greatestTT := planetOccultationMinimizeExternalMetric(leftTT, rightTT, config, observer)
			if greatestTT >= startTT && greatestTT <= endTT &&
				planetOccultationLatitudePass(greatestTT, config, observer, safetyMarginArcsec) {
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

func planetOccultationMinimizeExternalMetric(left, right float64, config planetOccultationConfig, observer *Observer) float64 {
	if right <= left {
		return left
	}
	const goldenRatio = 0.6180339887498949
	x1 := right - goldenRatio*(right-left)
	x2 := left + goldenRatio*(right-left)
	f1 := planetOccultationExternalContactMetric(x1, config, observer, -1)
	f2 := planetOccultationExternalContactMetric(x2, config, observer, -1)
	for i := 0; i < 64 && right-left > planetOccultationRootToleranceDays; i++ {
		if f1 > f2 {
			left = x1
			x1, f1 = x2, f2
			x2 = left + goldenRatio*(right-left)
			f2 = planetOccultationExternalContactMetric(x2, config, observer, -1)
		} else {
			right = x2
			x2, f2 = x1, f1
			x1 = right - goldenRatio*(right-left)
			f1 = planetOccultationExternalContactMetric(x1, config, observer, -1)
		}
	}
	return (left + right) / 2
}

func planetOccultationBestObserver(seedTT, startTT, endTT float64, config planetOccultationConfig) (float64, Observer, float64, bool) {
	frameAt := func(tt float64) (occultationPathFrame, bool) {
		return planetOccultationPathFrameAt(tt, config)
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
		point, pointOK = occultationPathBoundaryPointForFrame(greatestTT, frameAt, time.UTC)
	}
	if !pointOK {
		return 0, Observer{}, 0, false
	}
	observer := Observer{Longitude: point.Longitude, Latitude: point.Latitude}
	state := planetOccultationStateAt(greatestTT, config, &observer, -1)
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
	location *time.Location,
) (PlanetOccultationInfo, bool) {
	if !planetOccultationLatitudePass(greatestTT, config, &observer, safetyMarginArcsec) {
		return PlanetOccultationInfo{}, false
	}
	state := planetOccultationStateAt(greatestTT, config, &observer, -1)
	if !state.valid || state.externalContactMetric > 0 {
		return PlanetOccultationInfo{}, false
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

	if math.Abs(state.externalContactMetric) <= planetOccultationGrazingToleranceArcsec {
		info.Type = OccultationGrazing
		info.ExternalImmersion = info.Greatest
		info.ExternalEmersion = info.Greatest
		info.ContactsComplete = true
		return info, true
	}

	externalImmersionTT, externalImmersionOK := planetOccultationContact(greatestTT, -1, false, config, observer)
	externalEmersionTT, externalEmersionOK := planetOccultationContact(greatestTT, 1, false, config, observer)
	if !externalImmersionOK || !externalEmersionOK || externalEmersionTT <= externalImmersionTT {
		return PlanetOccultationInfo{}, false
	}
	info.ExternalImmersion = occultationTTToLocation(externalImmersionTT, location)
	info.ExternalEmersion = occultationTTToLocation(externalEmersionTT, location)

	if state.internalContactMetric < -planetOccultationGrazingToleranceArcsec {
		internalImmersionTT, internalImmersionOK := planetOccultationContact(greatestTT, -1, true, config, observer)
		internalEmersionTT, internalEmersionOK := planetOccultationContact(greatestTT, 1, true, config, observer)
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

func planetMoonPositionAt(tt float64, config planetOccultationConfig, observer *Observer, n int) planetMoonPosition {
	moonRA, moonDec := HMoonGeocentricApparentRaDecN(tt, n)
	planetRA, planetDec := config.apparentRaDecN(tt, n)
	if observer != nil {
		ut := TD2UT(tt, false)
		moonDistanceAU := HMoonAwayN(tt, n) / angularDiameterAstronomicalUnitKM
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

func planetOccultationStateAt(tt float64, config planetOccultationConfig, observer *Observer, n int) planetOccultationState {
	position := planetMoonPositionAt(tt, config, observer, n)
	moonRadius := MoonSemidiameterN(tt, n)
	planetRadius := config.semidiameterN(tt, n)
	if observer != nil {
		moonRadius = moonTopocentricSemidiameterN(tt, *observer, n)
		planetRadius = planetTopocentricSemidiameterN(tt, config, *observer, n)
	}
	if !position.valid || !finite(moonRadius) || !finite(planetRadius) || moonRadius <= planetRadius || planetRadius <= 0 {
		return planetOccultationState{}
	}
	separation := angularSeparationDegrees(position.moonRA, position.moonDec, position.planetRA, position.planetDec) * 3600
	return planetOccultationState{
		position:              position,
		separationArcsec:      separation,
		moonSemidiameter:      moonRadius,
		planetSemidiameter:    planetRadius,
		externalContactMetric: separation - (moonRadius + planetRadius),
		internalContactMetric: separation - (moonRadius - planetRadius),
		valid:                 finite(separation),
	}
}

func planetTopocentricSemidiameterN(tt float64, config planetOccultationConfig, observer Observer, n int) float64 {
	ra, dec := config.apparentRaDecN(tt, n)
	distanceKM := config.earthDistanceN(tt, n) * angularDiameterAstronomicalUnitKM
	if !finite(ra) || !finite(dec) || !finite(distanceKM) || distanceKM <= 0 {
		return math.NaN()
	}
	distanceKM = topocentricDistanceKM(ra, dec, distanceKM, observer, TD2UT(tt, false))
	if !finite(distanceKM) || distanceKM <= config.equatorialRadiusKM {
		return math.NaN()
	}
	return angularSemidiameterArcsec(config.equatorialRadiusKM, distanceKM)
}

func planetOccultationExternalContactMetric(tt float64, config planetOccultationConfig, observer *Observer, n int) float64 {
	state := planetOccultationStateAt(tt, config, observer, n)
	if !state.valid {
		return math.Inf(1)
	}
	return state.externalContactMetric
}

func planetMoonSeparationArcsec(tt float64, config planetOccultationConfig, observer *Observer, n int) float64 {
	position := planetMoonPositionAt(tt, config, observer, n)
	if !position.valid {
		return math.Inf(1)
	}
	return angularSeparationDegrees(position.moonRA, position.moonDec, position.planetRA, position.planetDec) * 3600
}

func planetOccultationLatitudePass(tt float64, config planetOccultationConfig, observer *Observer, safetyMarginArcsec float64) bool {
	state := planetOccultationStateAt(tt, config, observer, -1)
	if !state.valid {
		return false
	}
	_, moonLatitude := RaDecToLoBo(tt, state.position.moonRA, state.position.moonDec)
	_, planetLatitude := RaDecToLoBo(tt, state.position.planetRA, state.position.planetDec)
	limit := state.moonSemidiameter + state.planetSemidiameter + planetOccultationLatitudeMarginArcsec + safetyMarginArcsec
	return math.Abs(moonLatitude-planetLatitude)*3600 <= limit
}

func planetOccultationContact(
	greatestTT float64,
	direction int,
	internal bool,
	config planetOccultationConfig,
	observer Observer,
) (float64, bool) {
	if direction != -1 && direction != 1 {
		return math.NaN(), false
	}
	metric := func(tt float64) float64 {
		state := planetOccultationStateAt(tt, config, &observer, -1)
		if !state.valid {
			return math.NaN()
		}
		if internal {
			return state.internalContactMetric
		}
		return state.externalContactMetric
	}
	nearTT := greatestTT
	nearValue := metric(nearTT)
	if !finite(nearValue) || nearValue > 0 {
		return math.NaN(), false
	}
	maxSteps := int(math.Ceil(planetOccultationContactSpanDays / planetOccultationContactStepDays))
	if maxSteps > planetOccultationMaxContactSteps {
		maxSteps = planetOccultationMaxContactSteps
	}
	for i := 1; i <= maxSteps; i++ {
		farTT := greatestTT + float64(direction*i)*planetOccultationContactStepDays
		farValue := metric(farTT)
		if !finite(farValue) {
			continue
		}
		if farValue >= 0 {
			return planetOccultationRoot(nearTT, farTT, nearValue, farValue, metric)
		}
		nearTT, nearValue = farTT, farValue
	}
	return math.NaN(), false
}

func planetOccultationRoot(
	leftTT, rightTT, leftValue, rightValue float64,
	metric func(float64) float64,
) (float64, bool) {
	if leftTT > rightTT {
		leftTT, rightTT = rightTT, leftTT
		leftValue, rightValue = rightValue, leftValue
	}
	if !finite(leftValue) || !finite(rightValue) || leftValue*rightValue > 0 {
		return math.NaN(), false
	}
	for i := 0; i < 64 && math.Abs(rightTT-leftTT) > planetOccultationRootToleranceDays; i++ {
		midTT := (leftTT + rightTT) / 2
		midValue := metric(midTT)
		if !finite(midValue) {
			return math.NaN(), false
		}
		if leftValue*midValue <= 0 {
			rightTT, rightValue = midTT, midValue
		} else {
			leftTT, leftValue = midTT, midValue
		}
	}
	return (leftTT + rightTT) / 2, true
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
