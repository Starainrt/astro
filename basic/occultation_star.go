package basic

import (
	"math"
	"sort"
	"time"

	"github.com/starainrt/astro/internal/iau2006"
)

const (
	starOccultationSiderealMonthDays = 27.321661
	starOccultationDefaultStepDays   = 0.25
	starOccultationContactStepDays   = 10.0 / 1440.0
	starOccultationContactSpanDays   = 2.0
	starOccultationLatitudeMarginAS  = 3600.0
	starOccultationMoonLatitudeDeg   = 6.0
	starOccultationGrazingTolerance  = 0.01
	starOccultationRootToleranceDays = occultationEventSelectionToleranceDays
	starOccultationMaxContactSteps   = 10000
)

// FindStarOccultationsWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func FindStarOccultationsWithInstance(start, end time.Time, star StarCoordinate, longitude, latitude, height float64,
	options OccultationSearchOptions, inst *Instance) ([]StarOccultationInfo, error) {
	if err := validateOccultationTimeRange(start, end); err != nil {
		return nil, err
	}
	if err := star.Validate(); err != nil {
		return nil, err
	}
	observer := Observer{Longitude: longitude, Latitude: latitude, Height: height}
	if err := observer.Validate(); err != nil {
		return nil, err
	}
	if err := options.Validate(); err != nil {
		return nil, err
	}

	startTT := occultationTimeToTT(start)
	endTT := occultationTimeToTT(end)
	resultLocation := start.Location()
	results := make([]StarOccultationInfo, 0)
	for _, greatestTT := range starOccultationCandidateGreatestTimes(startTT, endTT, starOccultationCoarseStepDays(options), star, observer, options.SafetyMarginArcsec, inst) {
		info, ok := starOccultationInfoWithLimb(greatestTT, star, observer, options.SafetyMarginArcsec, resultLocation, options.LunarLimb, inst)
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

// FindBestStarOccultationsWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func FindBestStarOccultationsWithInstance(start, end time.Time, star StarCoordinate, options OccultationSearchOptions, inst *Instance) ([]StarOccultationInfo, error) {
	if err := validateOccultationTimeRange(start, end); err != nil {
		return nil, err
	}
	if err := star.Validate(); err != nil {
		return nil, err
	}
	if err := options.Validate(); err != nil {
		return nil, err
	}

	startTT := occultationTimeToTT(start)
	endTT := occultationTimeToTT(end)
	selectionStartTT := startTT - occultationEventSelectionToleranceDays
	selectionEndTT := endTT + occultationEventSelectionToleranceDays
	candidateStartTT := startTT - occultationPathSearchSpanDays
	candidateEndTT := endTT + occultationPathSearchSpanDays
	resultLocation := start.Location()
	results := make([]StarOccultationInfo, 0)
	for _, seedTT := range starOccultationGeocentricCandidateGreatestTimes(candidateStartTT, candidateEndTT, starOccultationCoarseStepDays(options), star, options.SafetyMarginArcsec, inst) {
		greatestTT, observer, _, observerOK := starOccultationBestObserver(seedTT, selectionStartTT, selectionEndTT, star, inst)
		if !observerOK {
			continue
		}
		info, ok := starOccultationInfoWithLimb(greatestTT, star, observer, options.SafetyMarginArcsec, resultLocation, options.LunarLimb, inst)
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

func starOccultationCandidateGreatestTimes(startTT, endTT, step float64, star StarCoordinate, observer Observer, safetyMarginArcsec float64, instances ...*Instance) []float64 {
	inst := instanceValue(instances)

	results := make([]float64, 0)
	for cycleStart := startTT; cycleStart < endTT; cycleStart += starOccultationSiderealMonthDays {
		cycleEnd := math.Min(cycleStart+starOccultationSiderealMonthDays, endTT)
		scanStart := math.Max(startTT-step, cycleStart-step)
		scanEnd := math.Min(endTT+step, cycleEnd+step)
		if !starOccultationLatitudeEnvelopePass(scanStart, scanEnd, star, &observer, safetyMarginArcsec, inst) {
			continue
		}
		results = append(results, starOccultationScanCandidates(
			scanStart, scanEnd, step,
			func(tt float64) float64 { return starMoonSeparationArcsec(tt, star, observer, inst) },
			func(tt float64) bool {
				return starOccultationLatitudePass(tt, star, observer, safetyMarginArcsec, inst)
			},
		)...)
	}
	return uniqueOccultationCandidateTimes(results, startTT, endTT)
}

func starOccultationGeocentricCandidateGreatestTimes(startTT, endTT, step float64, star StarCoordinate, safetyMarginArcsec float64, instances ...*Instance) []float64 {
	inst := instanceValue(instances)

	results := make([]float64, 0)
	for cycleStart := startTT; cycleStart < endTT; cycleStart += starOccultationSiderealMonthDays {
		cycleEnd := math.Min(cycleStart+starOccultationSiderealMonthDays, endTT)
		scanStart := math.Max(startTT-step, cycleStart-step)
		scanEnd := math.Min(endTT+step, cycleEnd+step)
		if !starOccultationLatitudeEnvelopePass(scanStart, scanEnd, star, nil, safetyMarginArcsec, inst) {
			continue
		}
		results = append(results, starOccultationScanCandidates(
			scanStart, scanEnd, step,
			func(tt float64) float64 { return starOccultationGeocentricSeparationArcsec(tt, star, inst) },
			func(tt float64) bool {
				return starOccultationLatitudePassGeocentric(tt, star, safetyMarginArcsec, inst)
			},
		)...)
	}
	return uniqueOccultationCandidateTimes(results, startTT, endTT)
}

// starOccultationScanCandidates 找出粗扫描中的所有局部最小值。
// 恒星月仍适合作为黄纬预筛桶，但不能假定每个桶恰好只有一个最近接近。
// starOccultationScanCandidates finds every local minimum in a coarse scan.
// A lunar month remains a useful latitude-prefilter bucket, but it must not be treated as a promise that exactly one closest approach exists in that bucket.
func starOccultationScanCandidates(
	startTT, endTT, step float64,
	value func(float64) float64,
	accept func(float64) bool,
) []float64 {
	if endTT < startTT {
		return nil
	}
	if endTT == startTT {
		if accept(startTT) && finite(value(startTT)) {
			return []float64{startTT}
		}
		return nil
	}
	if step <= 0 || !finite(step) {
		step = starOccultationDefaultStepDays
	}
	if step > (endTT-startTT)/2 {
		step = (endTT - startTT) / 2
	}
	appendCandidate := func(results *[]float64, tt float64) {
		if tt < startTT || tt > endTT || !finite(tt) || !finite(value(tt)) || !accept(tt) {
			return
		}
		if len(*results) == 0 || math.Abs(tt-(*results)[len(*results)-1]) > 60.0/86400.0 {
			*results = append(*results, tt)
		}
	}

	results := make([]float64, 0, 2)
	leftTT, leftValue := startTT, value(startTT)
	centerTT := math.Min(startTT+step, endTT)
	centerValue := value(centerTT)
	for centerTT < endTT {
		rightTT := math.Min(centerTT+step, endTT)
		rightValue := value(rightTT)
		if finite(leftValue) && finite(centerValue) && finite(rightValue) &&
			centerValue <= leftValue && centerValue <= rightValue {
			candidate := starOccultationMinimizeValue(leftTT, rightTT, value)
			appendCandidate(&results, candidate)
		}
		leftTT, leftValue = centerTT, centerValue
		centerTT, centerValue = rightTT, rightValue
	}
	return results
}

func uniqueOccultationCandidateTimes(times []float64, startTT, endTT float64) []float64 {
	if len(times) == 0 {
		return nil
	}
	sort.Float64s(times)
	unique := times[:0]
	for _, tt := range times {
		if tt < startTT || tt > endTT {
			continue
		}
		if len(unique) == 0 || math.Abs(tt-unique[len(unique)-1]) > 60.0/86400.0 {
			unique = append(unique, tt)
		}
	}
	return unique
}

func starOccultationLatitudeEnvelopePass(startTT, endTT float64, star StarCoordinate, observer *Observer, safetyMarginArcsec float64, instances ...*Instance) bool {
	inst := instanceValue(instances)

	minimumLatitude := math.Inf(1)
	maximumLatitude := math.Inf(-1)
	maximumMoonRadiusDeg := 0.0
	for _, tt := range []float64{startTT, (startTT + endTT) / 2, endTT} {
		var ra, dec float64
		if observer == nil {
			ra, dec = starApparentRaDecGeocentric(tt, star, inst)
		} else {
			ra, dec = starApparentRaDec(tt, star, *observer, inst)
		}
		_, latitude := RaDecToLoBo(tt, ra, dec)
		minimumLatitude = math.Min(minimumLatitude, latitude)
		maximumLatitude = math.Max(maximumLatitude, latitude)
		moonRadius := MoonSemidiameterWithInstance(tt, inst)
		if observer != nil {
			moonRadius = moonTopocentricSemidiameterN(tt, *observer, -1, inst)
		}
		maximumMoonRadiusDeg = math.Max(maximumMoonRadiusDeg, moonRadius/3600)
	}
	limit := starOccultationMoonLatitudeDeg + starOccultationLatitudeMarginAS/3600 + safetyMarginArcsec/3600 + maximumMoonRadiusDeg
	return minimumLatitude <= limit && maximumLatitude >= -limit
}

func starOccultationGeocentricStarLongitude(tt float64, star StarCoordinate) float64 {
	ra, dec := starApparentRaDecGeocentric(tt, star)
	longitude, _ := RaDecToLoBo(tt, ra, dec)
	return longitude
}

func starOccultationMinimizeValue(left, right float64, value func(float64) float64) float64 {
	if right <= left {
		return left
	}
	const goldenRatio = 0.6180339887498949
	x1 := right - goldenRatio*(right-left)
	x2 := left + goldenRatio*(right-left)
	f1 := value(x1)
	f2 := value(x2)
	for i := 0; i < 64 && right-left > starOccultationRootToleranceDays; i++ {
		if f1 > f2 {
			left = x1
			x1, f1 = x2, f2
			x2 = left + goldenRatio*(right-left)
			f2 = value(x2)
		} else {
			right = x2
			x2, f2 = x1, f1
			x1 = right - goldenRatio*(right-left)
			f1 = value(x1)
		}
	}
	return (left + right) / 2
}

func starOccultationGeocentricSeparationArcsec(tt float64, star StarCoordinate, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	moonRA, moonDec := HMoonGeocentricApparentRaDecNWithInstance(tt, -1, inst)
	starRA, starDec, _ := starApparentRaDecDistanceGeocentric(tt, star, inst)
	return angularSeparationDegrees(moonRA, moonDec, starRA, starDec) * 3600
}

func starOccultationBestObserver(seedTT, startTT, endTT float64, star StarCoordinate, instances ...*Instance) (float64, Observer, float64, bool) {
	inst := instanceValue(instances)

	searchStart := seedTT - occultationPathSearchSpanDays
	searchEnd := seedTT + occultationPathSearchSpanDays
	outerStart, outerEnd, ok := starOccultationPathWindow(seedTT, searchStart, searchEnd, star, false, inst)
	if !ok {
		return 0, Observer{}, 0, false
	}
	greatestTT := starOccultationPathGreatest(seedTT, outerStart, outerEnd, star, inst)
	if greatestTT < startTT || greatestTT > endTT {
		return 0, Observer{}, 0, false
	}
	point, pointOK := starOccultationPathCenterPoint(greatestTT, star, time.UTC, inst)
	if !pointOK {
		frameAt := func(tt float64) (occultationPathFrame, bool) {
			return starOccultationPathFrameAt(tt, star, inst)
		}
		// 同行星最佳事件搜索：非中心路径的最佳站心取离影轴最近的椭球点，
		// 外接触切点位于掩带边缘，只作最后回退。
		// Same as the planetary best-event search: for a non-central path the
		// best station is the ellipsoid point nearest to the shadow axis, while
		// the outer contact tangent on the band edge stays the last resort.
		point, pointOK = occultationPathTrackPointForFrame(greatestTT, frameAt, time.UTC)
		if !pointOK {
			point, pointOK = occultationPathBoundaryPointForFrame(greatestTT, frameAt, time.UTC, inst)
		}
	}
	if !pointOK {
		return 0, Observer{}, 0, false
	}
	observer := Observer{Longitude: point.Longitude, Latitude: point.Latitude}
	position := starMoonPositionAt(greatestTT, star, observer, inst)
	moonRadius := moonTopocentricSemidiameterN(greatestTT, observer, -1, inst)
	if !position.valid || !finite(moonRadius) {
		return 0, Observer{}, 0, false
	}
	metric := angularSeparationDegrees(position.moonRA, position.moonDec, position.starRA, position.starDec)*3600 - moonRadius
	return greatestTT, observer, metric, finite(metric)
}

func starOccultationInfoAtGreatest(greatestTT float64, star StarCoordinate, observer Observer, safetyMarginArcsec float64, location *time.Location, instances ...*Instance) (StarOccultationInfo, bool) {
	inst := instanceValue(instances)

	info, ok := starOccultationGreatestInfo(greatestTT, star, observer, safetyMarginArcsec, location, inst)
	if !ok || info.MinimumSeparationArcsec > info.MoonSemidiameterArcsec {
		return StarOccultationInfo{}, false
	}
	minimumResidual := info.MinimumSeparationArcsec - info.MoonSemidiameterArcsec
	if math.Abs(minimumResidual) <= starOccultationGrazingTolerance {
		info.Type = OccultationGrazing
		info.Immersion = info.Greatest
		info.Emersion = info.Greatest
		info.ContactsComplete = true
		return info, true
	}
	immersionTT, immersionOK := starOccultationContact(greatestTT, greatestTT-starOccultationContactSpanDays, -1, star, observer, inst)
	emersionTT, emersionOK := starOccultationContact(greatestTT, greatestTT+starOccultationContactSpanDays, 1, star, observer, inst)
	if !immersionOK || !emersionOK {
		return StarOccultationInfo{}, false
	}
	info.Immersion = occultationTTToLocation(immersionTT, location)
	info.Emersion = occultationTTToLocation(emersionTT, location)
	info.ContactsComplete = true
	return info, true
}

// starOccultationGreatestInfo 只填掩甚几何，不判断圆月面是否成掩、不求接触。
func starOccultationGreatestInfo(greatestTT float64, star StarCoordinate, observer Observer, safetyMarginArcsec float64, location *time.Location, inst *Instance) (StarOccultationInfo, bool) {
	if !starOccultationLatitudePass(greatestTT, star, observer, safetyMarginArcsec, inst) {
		return StarOccultationInfo{}, false
	}
	minimumSeparation := starMoonSeparationArcsec(greatestTT, star, observer, inst)
	moonRadius := moonTopocentricSemidiameterN(greatestTT, observer, -1, inst)
	if !finite(minimumSeparation) || !finite(moonRadius) {
		return StarOccultationInfo{}, false
	}
	greatestPosition := starMoonPositionAt(greatestTT, star, observer, inst)
	if !greatestPosition.valid {
		return StarOccultationInfo{}, false
	}
	info := StarOccultationInfo{
		TargetID:                star.ID,
		Observer:                observer,
		Type:                    OccultationTotal,
		Greatest:                occultationTTToLocation(greatestTT, location),
		MinimumSeparationArcsec: minimumSeparation,
		PositionAngleDeg:        occultationPositionAngle(greatestPosition.moonRA, greatestPosition.moonDec, greatestPosition.starRA, greatestPosition.starDec),
		MoonSemidiameterArcsec:  moonRadius,
		MoonAltitudeAtGreatest:  occultationAltitude(greatestTT, observer, greatestPosition.moonRA, greatestPosition.moonDec),
		MoonAzimuthAtGreatest:   occultationAzimuth(greatestTT, observer, greatestPosition.moonRA, greatestPosition.moonDec),
	}
	info.VisibleAtGreatest = info.MoonAltitudeAtGreatest >= 0
	return info, true
}

type starMoonPosition struct {
	moonRA, moonDec float64
	starRA, starDec float64
	valid           bool
}

func starMoonPositionAt(tt float64, star StarCoordinate, observer Observer, instances ...*Instance) starMoonPosition {
	inst := instanceValue(instances)

	moonRA, moonDec := moonTopocentricApparentRaDec(tt, observer, -1, inst)
	starRA, starDec := starApparentRaDec(tt, star, observer, inst)
	return starMoonPosition{
		moonRA:  moonRA,
		moonDec: moonDec,
		starRA:  starRA,
		starDec: starDec,
		valid:   finite(moonRA) && finite(moonDec) && finite(starRA) && finite(starDec),
	}
}

func moonTopocentricApparentRaDec(tt float64, observer Observer, n int, instances ...*Instance) (float64, float64) {
	inst := instanceValue(instances)

	ra, dec := HMoonGeocentricApparentRaDecNWithInstance(tt, n, inst)
	ut := TT2UTC(tt)
	distanceAU := HMoonAwayNWithInstance(tt, n, inst) / 149597870.7
	ra, dec = TopocentricRaDec(ra, dec, observer.Latitude, observer.Longitude, ut, distanceAU, observer.Height)
	return normalizeRA(ra), dec
}

func starApparentRaDec(tt float64, star StarCoordinate, observer Observer, instances ...*Instance) (float64, float64) {
	inst := instanceValue(instances)

	ra, dec, distanceAU := starApparentRaDecDistanceGeocentric(tt, star, inst)
	if distanceAU > 0 {
		ra, dec = TopocentricRaDec(ra, dec, observer.Latitude, observer.Longitude, TT2UTC(tt), distanceAU, observer.Height)
		ra = normalizeRA(ra)
	}
	return ra, dec
}

func starApparentRaDecGeocentric(tt float64, star StarCoordinate, instances ...*Instance) (float64, float64) {
	inst := instanceValue(instances)

	ra, dec, _ := starApparentRaDecDistanceGeocentric(tt, star, inst)
	return ra, dec
}

// starApparentRaDecDistanceGeocentric 同时给出视位置与推进后的距离（天文单位，0 表示距离未知）。
func starApparentRaDecDistanceGeocentric(tt float64, star StarCoordinate, instances ...*Instance) (float64, float64, float64) {
	inst := instanceValue(instances)

	epochJD := occultationTimeToTT(star.Epoch)
	parallaxMas := star.parallaxMas()
	ra := star.RA
	dec := star.Dec
	precessionEpoch := 2451545.0
	if star.Frame == CoordinateFrameICRS {
		ra, dec = starICRSToMeanJ2000RaDec(ra, dec)
	} else if star.Frame == CoordinateFrameApparentOfDate {
		ra, dec = starApparentToMeanRaDec(epochJD, ra, dec, parallaxMas, inst)
		precessionEpoch = epochJD
	}
	epochDistanceAU := starEpochDistanceAU(parallaxMas)
	ra, dec, distanceAU := starProperMotionRaDec(tt, star, ra, dec, epochDistanceAU)
	ra, dec = Precess(ra, dec, precessionEpoch, tt)
	ra, dec = starMeanToApparentRaDec(tt, ra, dec, parallaxMasAtDistance(parallaxMas, distanceAU), inst)
	return ra, dec, distanceAU
}

// starEpochDistanceAU 由历元视差给出距离，非正表示距离未知。
func starEpochDistanceAU(parallaxMas float64) float64 {
	if parallaxMas <= 0 {
		return 0
	}
	return 206264806.247 / parallaxMas
}

// starPropagatedPositionAU 把历元位置矢量按三维匀速直线运动推进一个历元差。
// 切向速度取自行乘历元距离，视向分量取径向速度；返回推进后的矢量，其模长即当日距离。
func starPropagatedPositionAU(star StarCoordinate, ra, dec, epochDistanceAU, years float64) ([3]float64, float64) {
	if epochDistanceAU <= 0 {
		return [3]float64{}, 0
	}
	raRad := ra * math.Pi / 180
	decRad := dec * math.Pi / 180
	cosDec, sinDec := math.Cos(decRad), math.Sin(decRad)
	cosRA, sinRA := math.Cos(raRad), math.Sin(raRad)
	pmRA := star.ProperMotionRACosDecMasPerYear / 1000 * math.Pi / (180 * 3600) * epochDistanceAU
	pmDec := star.ProperMotionDecMasPerYear / 1000 * math.Pi / (180 * 3600) * epochDistanceAU
	radial := star.RadialVelocityKmPerSecond * 365.25 * 86400 / 149597870.7
	position := [3]float64{
		epochDistanceAU*cosDec*cosRA + years*(pmDec*(-sinDec*cosRA)-pmRA*sinRA+radial*cosDec*cosRA),
		epochDistanceAU*cosDec*sinRA + years*(pmDec*(-sinDec*sinRA)+pmRA*cosRA+radial*cosDec*sinRA),
		epochDistanceAU*sinDec + years*(pmDec*cosDec+radial*sinDec),
	}
	norm := math.Sqrt(position[0]*position[0] + position[1]*position[1] + position[2]*position[2])
	return position, norm
}

// parallaxMasAtDistance 把推进后的距离折回周年视差，视差修正必须跟着距离一起变。
func parallaxMasAtDistance(parallaxMas, distanceAU float64) float64 {
	if parallaxMas <= 0 || distanceAU <= 0 {
		return 0
	}
	return 206264806.247 / distanceAU
}

// starProperMotionRaDec 把历元输入坐标推进到 tt，并给出推进后的距离。
// 距离已知走三维、否则只推进两个角分量（此时距离返回 0）。
func starProperMotionRaDec(tt float64, star StarCoordinate, ra, dec, epochDistanceAU float64) (float64, float64, float64) {
	years := (tt - occultationTimeToTT(star.Epoch)) / 365.25
	if epochDistanceAU > 0 {
		return starProperMotionRaDec3D(ra, dec, years, epochDistanceAU, star)
	}
	cosDec := math.Cos(dec * math.Pi / 180)
	if math.Abs(cosDec) > 1e-12 {
		ra += years * star.ProperMotionRACosDecMasPerYear / (3600000.0 * cosDec)
	}
	dec += years * star.ProperMotionDecMasPerYear / 3600000.0
	return ra, math.Max(-90, math.Min(90, dec)), 0
}

// starProperMotionRaDec3D 按三维匀速直线运动推进：赤经赤纬只是位置矢量的方向。
func starProperMotionRaDec3D(ra, dec, years, epochDistanceAU float64, star StarCoordinate) (float64, float64, float64) {
	position, norm := starPropagatedPositionAU(star, ra, dec, epochDistanceAU, years)
	outRA := math.Atan2(position[1], position[0]) * 180 / math.Pi
	if outRA < 0 {
		outRA += 360
	}
	return outRA, math.Asin(position[2]/norm) * 180 / math.Pi, norm
}

func starICRSToMeanJ2000RaDec(ra, dec float64) (float64, float64) {
	b := iau2006.FrameBias()
	raRad := ra * math.Pi / 180
	decRad := dec * math.Pi / 180
	x := math.Cos(decRad) * math.Cos(raRad)
	y := math.Cos(decRad) * math.Sin(raRad)
	z := math.Sin(decRad)
	biasedX := b[0][0]*x + b[0][1]*y + b[0][2]*z
	biasedY := b[1][0]*x + b[1][1]*y + b[1][2]*z
	biasedZ := b[2][0]*x + b[2][1]*y + b[2][2]*z
	return normalizeRA(math.Atan2(biasedY, biasedX) * 180 / math.Pi),
		math.Atan2(biasedZ, math.Hypot(biasedX, biasedY)) * 180 / math.Pi
}

func starMeanToApparentRaDec(tt, ra, dec, parallaxMas float64, instances ...*Instance) (float64, float64) {
	inst := instanceValue(instances)

	longitude, latitude := starMeanEquatorialToEcliptic(tt, ra, dec)
	if parallaxMas > 0 {
		longitude, latitude = starAnnualParallaxEcliptic(tt, longitude, latitude, parallaxMas, inst)
	}
	meanLongitude, meanLatitude := longitude, latitude
	longitude = normalizeRA(meanLongitude + GXCLoWithInstance(meanLongitude, meanLatitude, tt, inst)/3600 + Nutation2000Bi(tt))
	latitude = meanLatitude + GXCBoWithInstance(meanLongitude, meanLatitude, tt, inst)/3600
	ra, dec = LoBoToRaDec(tt, longitude, latitude)
	return normalizeRA(ra), dec
}

func starApparentToMeanRaDec(tt, apparentRA, apparentDec, parallaxMas float64, instances ...*Instance) (float64, float64) {
	inst := instanceValue(instances)

	meanRA, meanDec := apparentRA, apparentDec
	for i := 0; i < 8; i++ {
		computedRA, computedDec := starMeanToApparentRaDec(tt, meanRA, meanDec, parallaxMas, inst)
		meanRA = normalizeRA(meanRA - signedAngleDifference(computedRA, apparentRA))
		meanDec -= computedDec - apparentDec
	}
	return meanRA, math.Max(-90, math.Min(90, meanDec))
}

func starMeanEquatorialToEcliptic(tt, ra, dec float64) (float64, float64) {
	obliquity := EclipticObliquity(tt, false) * math.Pi / 180
	ra *= math.Pi / 180
	dec *= math.Pi / 180
	longitude := math.Atan2(
		math.Sin(ra)*math.Cos(obliquity)+math.Tan(dec)*math.Sin(obliquity),
		math.Cos(ra),
	) * 180 / math.Pi
	latitude := math.Asin(
		math.Sin(dec)*math.Cos(obliquity)-math.Cos(dec)*math.Sin(obliquity)*math.Sin(ra),
	) * 180 / math.Pi
	return normalizeRA(longitude), latitude
}

func starAnnualParallaxEcliptic(tt, longitude, latitude, parallaxMas float64, instances ...*Instance) (float64, float64) {
	inst := instanceValue(instances)

	distanceAU := 206264806.247 / parallaxMas
	longitudeRad := longitude * math.Pi / 180
	latitudeRad := latitude * math.Pi / 180
	cosLatitude := math.Cos(latitudeRad)
	starX := distanceAU * cosLatitude * math.Cos(longitudeRad)
	starY := distanceAU * cosLatitude * math.Sin(longitudeRad)
	starZ := distanceAU * math.Sin(latitudeRad)

	earthLongitude := normalizeRA(HSunTrueLoNWithInstance(tt, -1, inst)+180) * math.Pi / 180
	earthDistance := EarthAwayNWithInstance(tt, -1, inst)
	starX -= earthDistance * math.Cos(earthLongitude)
	starY -= earthDistance * math.Sin(earthLongitude)

	longitude = math.Atan2(starY, starX) * 180 / math.Pi
	latitude = math.Atan2(starZ, math.Hypot(starX, starY)) * 180 / math.Pi
	return normalizeRA(longitude), latitude
}

func starOccultationLongitudeDistance(tt float64, star StarCoordinate, observer Observer) float64 {
	moonLongitude := HMoonTrueLoN(tt, 8)
	starRA, starDec := starApparentRaDec(tt, star, observer)
	starLongitude, _ := RaDecToLoBo(tt, starRA, starDec)
	return math.Abs(signedAngleDifference(moonLongitude, starLongitude))
}

func starMoonSeparationArcsec(tt float64, star StarCoordinate, observer Observer, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	position := starMoonPositionAt(tt, star, observer, inst)
	if !position.valid {
		return math.Inf(1)
	}
	return angularSeparationDegrees(position.moonRA, position.moonDec, position.starRA, position.starDec) * 3600
}

func starOccultationLatitudePass(tt float64, star StarCoordinate, observer Observer, safetyMarginArcsec float64, instances ...*Instance) bool {
	inst := instanceValue(instances)

	starRA, starDec := starApparentRaDec(tt, star, observer, inst)
	_, starLatitude := RaDecToLoBo(tt, starRA, starDec)
	moonLatitude := HMoonTrueBoNWithInstance(tt, 8, inst)
	moonRadius := moonTopocentricSemidiameterN(tt, observer, -1, inst)
	limit := moonRadius + starOccultationLatitudeMarginAS + safetyMarginArcsec
	return math.Abs(starLatitude-moonLatitude)*3600 <= limit
}

func starOccultationLatitudePassGeocentric(tt float64, star StarCoordinate, safetyMarginArcsec float64, instances ...*Instance) bool {
	inst := instanceValue(instances)

	starRA, starDec := starApparentRaDecGeocentric(tt, star, inst)
	_, starLatitude := RaDecToLoBo(tt, starRA, starDec)
	moonLatitude := HMoonTrueBoNWithInstance(tt, 8, inst)
	limit := MoonSemidiameterWithInstance(tt, inst) + starOccultationLatitudeMarginAS + safetyMarginArcsec
	return math.Abs(starLatitude-moonLatitude)*3600 <= limit
}

func starOccultationContact(greatestTT, boundaryTT float64, direction int, star StarCoordinate, observer Observer, instances ...*Instance) (float64, bool) {
	inst := instanceValue(instances)

	valueAtGreatest := starMoonSeparationArcsec(greatestTT, star, observer, inst) - moonTopocentricSemidiameterN(greatestTT, observer, -1, inst)
	if !finite(valueAtGreatest) || valueAtGreatest > 0 {
		return math.NaN(), false
	}
	currentTT := greatestTT
	currentValue := valueAtGreatest
	step := starOccultationContactStepDays
	for i := 0; i < starOccultationMaxContactSteps; i++ {
		nextTT := currentTT + float64(direction)*step
		if direction < 0 && nextTT < boundaryTT {
			nextTT = boundaryTT
		}
		if direction > 0 && nextTT > boundaryTT {
			nextTT = boundaryTT
		}
		nextValue := starMoonSeparationArcsec(nextTT, star, observer, inst) - moonTopocentricSemidiameterN(nextTT, observer, -1, inst)
		if finite(nextValue) && nextValue >= 0 {
			return starOccultationRoot(currentTT, nextTT, currentValue, nextValue, star, observer, inst)
		}
		if nextTT == boundaryTT {
			return math.NaN(), false
		}
		currentTT = nextTT
		currentValue = nextValue
	}
	return math.NaN(), false
}

func starOccultationRoot(leftTT, rightTT, leftValue, rightValue float64, star StarCoordinate, observer Observer, instances ...*Instance) (float64, bool) {
	inst := instanceValue(instances)

	if !finite(leftValue) || !finite(rightValue) || leftValue*rightValue > 0 {
		return math.NaN(), false
	}
	if leftValue == 0 {
		return leftTT, true
	}
	if rightValue == 0 {
		return rightTT, true
	}
	for i := 0; i < 64 && math.Abs(rightTT-leftTT) > starOccultationRootToleranceDays; i++ {
		midTT := (leftTT + rightTT) / 2
		midValue := starMoonSeparationArcsec(midTT, star, observer, inst) - moonTopocentricSemidiameterN(midTT, observer, -1, inst)
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

func starOccultationCoarseStepDays(options OccultationSearchOptions) float64 {
	step := starOccultationDefaultStepDays
	if options.MaxStep > 0 {
		requested := options.MaxStep.Hours() / 24
		if requested > 0 && requested < step {
			step = requested
		}
	}
	return math.Max(step, occultationSearchMinimumStep.Hours()/24)
}

func angularSeparationDegrees(ra1, dec1, ra2, dec2 float64) float64 {
	ra1 *= math.Pi / 180
	ra2 *= math.Pi / 180
	dec1 *= math.Pi / 180
	dec2 *= math.Pi / 180
	cosSeparation := math.Sin(dec1)*math.Sin(dec2) + math.Cos(dec1)*math.Cos(dec2)*math.Cos(ra1-ra2)
	return math.Acos(math.Max(-1, math.Min(1, cosSeparation))) * 180 / math.Pi
}

func occultationPositionAngle(moonRA, moonDec, starRA, starDec float64) float64 {
	deltaRA := (starRA - moonRA) * math.Pi / 180
	moonDecRad := moonDec * math.Pi / 180
	starDecRad := starDec * math.Pi / 180
	y := math.Sin(deltaRA) * math.Cos(starDecRad)
	x := math.Cos(moonDecRad)*math.Sin(starDecRad) - math.Sin(moonDecRad)*math.Cos(starDecRad)*math.Cos(deltaRA)
	return normalizeRA(math.Atan2(y, x) * 180 / math.Pi)
}

func occultationAltitude(tt float64, observer Observer, ra, dec float64) float64 {
	return occultationAltitudeWithSidereal(
		ApparentSiderealTime(TT2UT1(tt))*15, observer, ra, dec,
	)
}

func occultationAltitudeWithSidereal(siderealDegrees float64, observer Observer, ra, dec float64) float64 {
	hourAngle := signedAngleDifference(siderealDegrees+observer.Longitude, ra) * math.Pi / 180
	lat := observer.Latitude * math.Pi / 180
	declination := dec * math.Pi / 180
	sinAltitude := math.Sin(lat)*math.Sin(declination) + math.Cos(lat)*math.Cos(declination)*math.Cos(hourAngle)
	return math.Asin(math.Max(-1, math.Min(1, sinAltitude))) * 180 / math.Pi
}

func occultationAzimuth(tt float64, observer Observer, ra, dec float64) float64 {
	hourAngle := signedAngleDifference(ApparentSiderealTime(TT2UT1(tt))*15+observer.Longitude, ra) * math.Pi / 180
	lat := observer.Latitude * math.Pi / 180
	declination := dec * math.Pi / 180
	y := math.Sin(hourAngle)
	x := math.Cos(hourAngle)*math.Sin(lat) - math.Tan(declination)*math.Cos(lat)
	return normalizeRA(math.Atan2(y, x)*180/math.Pi + 180)
}

func signedAngleDifference(a, b float64) float64 {
	difference := math.Mod(a-b+180, 360)
	if difference < 0 {
		difference += 360
	}
	return difference - 180
}

func normalizeRA(ra float64) float64 {
	ra = math.Mod(ra, 360)
	if ra < 0 {
		ra += 360
	}
	return ra
}

func occultationTimeToTT(value time.Time) float64 {
	return UTC2TT(Date2JD(value.UTC()))
}

func occultationTTToLocation(tt float64, location *time.Location) time.Time {
	if location == nil {
		location = time.UTC
	}
	return JD2DateByZone(TT2UTC(tt), location, false)
}
