package basic

import (
	"math"
	"sort"
	"time"
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

// FindStarOccultations 搜索单颗点源恒星的月掩星。
//
// 输入坐标会从历元传播并转换到当日视坐标系；若提供视差，还会修正观测者的恒星视差。
// 搜索不会加载内嵌 9100 星表；需要全星表搜索时，调用者必须显式加载并选择恒星。经度东为正、纬度北为正，单位为度；高度为平均海平面以上米数。
// FindStarOccultations searches for lunar occultations of one point-source star.
// The input coordinate is propagated from its epoch, converted to the apparent frame of date, and corrected for the observer's stellar parallax when one is supplied.
// The search does not load the embedded 9100-star catalog; callers that need a catalog-wide search must load and select stars explicitly. Longitude is east-positive in degrees, latitude is north-positive in degrees, and height is the observer elevation above mean sea level in meters.
func FindStarOccultations(start, end time.Time, star StarCoordinate, longitude, latitude, height float64,
	options OccultationSearchOptions) ([]StarOccultationInfo, error) {
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
	for _, greatestTT := range starOccultationCandidateGreatestTimes(startTT, endTT, starOccultationCoarseStepDays(options), star, observer, options.SafetyMarginArcsec) {
		info, ok := starOccultationInfoAtGreatest(greatestTT, star, observer, options.SafetyMarginArcsec, resultLocation)
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

// FindBestStarOccultations 返回窗口内每次恒星月掩在地球上的全球几何掩甚点。
// 返回的 StarOccultationInfo.Observer 是海平面大地测量位置，由月掩几何选择，不使用地平线或可见性评分。
//
// 地心数据只用于月周期搜索初值；最终点是与 FindStarOccultationPaths 一致的标准全球掩甚路径点，然后在该处重新进行站心接触和可见性计算。查询端点 10 ms 内的掩甚时刻也会包含，与数值根精度一致。
// FindBestStarOccultations returns the global geometric greatest point on Earth for each stellar occultation in the window.
// The returned StarOccultationInfo.Observer is the geodetic location at sea level; it is selected from the occultation geometry, without a horizon or visibility score.
// Geocentric data only seeds each lunar-month search. The final point is the canonical global greatest-path point, matching FindStarOccultationPaths; event contacts and visibility are then recomputed topocentrically there. A greatest instant within 10 ms of either query endpoint is included, matching the numerical root precision.
func FindBestStarOccultations(start, end time.Time, star StarCoordinate, options OccultationSearchOptions) ([]StarOccultationInfo, error) {
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
	for _, seedTT := range starOccultationGeocentricCandidateGreatestTimes(candidateStartTT, candidateEndTT, starOccultationCoarseStepDays(options), star, options.SafetyMarginArcsec) {
		greatestTT, observer, _, observerOK := starOccultationBestObserver(seedTT, selectionStartTT, selectionEndTT, star)
		if !observerOK {
			continue
		}
		info, ok := starOccultationInfoAtGreatest(greatestTT, star, observer, options.SafetyMarginArcsec, resultLocation)
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

func starOccultationCandidateGreatestTimes(startTT, endTT, step float64, star StarCoordinate, observer Observer, safetyMarginArcsec float64) []float64 {
	results := make([]float64, 0)
	for cycleStart := startTT; cycleStart < endTT; cycleStart += starOccultationSiderealMonthDays {
		cycleEnd := math.Min(cycleStart+starOccultationSiderealMonthDays, endTT)
		scanStart := math.Max(startTT-step, cycleStart-step)
		scanEnd := math.Min(endTT+step, cycleEnd+step)
		if !starOccultationLatitudeEnvelopePass(scanStart, scanEnd, star, &observer, safetyMarginArcsec) {
			continue
		}
		results = append(results, starOccultationScanCandidates(
			scanStart, scanEnd, step,
			func(tt float64) float64 { return starMoonSeparationArcsec(tt, star, observer) },
			func(tt float64) bool {
				return starOccultationLatitudePass(tt, star, observer, safetyMarginArcsec)
			},
		)...)
	}
	return uniqueOccultationCandidateTimes(results, startTT, endTT)
}

func starOccultationGeocentricCandidateGreatestTimes(startTT, endTT, step float64, star StarCoordinate, safetyMarginArcsec float64) []float64 {
	results := make([]float64, 0)
	for cycleStart := startTT; cycleStart < endTT; cycleStart += starOccultationSiderealMonthDays {
		cycleEnd := math.Min(cycleStart+starOccultationSiderealMonthDays, endTT)
		scanStart := math.Max(startTT-step, cycleStart-step)
		scanEnd := math.Min(endTT+step, cycleEnd+step)
		if !starOccultationLatitudeEnvelopePass(scanStart, scanEnd, star, nil, safetyMarginArcsec) {
			continue
		}
		results = append(results, starOccultationScanCandidates(
			scanStart, scanEnd, step,
			func(tt float64) float64 { return starOccultationGeocentricSeparationArcsec(tt, star) },
			func(tt float64) bool {
				return starOccultationLatitudePassGeocentric(tt, star, safetyMarginArcsec)
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

func starOccultationLatitudeEnvelopePass(startTT, endTT float64, star StarCoordinate, observer *Observer, safetyMarginArcsec float64) bool {
	minimumLatitude := math.Inf(1)
	maximumLatitude := math.Inf(-1)
	maximumMoonRadiusDeg := 0.0
	for _, tt := range []float64{startTT, (startTT + endTT) / 2, endTT} {
		var ra, dec float64
		if observer == nil {
			ra, dec = starApparentRaDecGeocentric(tt, star)
		} else {
			ra, dec = starApparentRaDec(tt, star, *observer)
		}
		_, latitude := RaDecToLoBo(tt, ra, dec)
		minimumLatitude = math.Min(minimumLatitude, latitude)
		maximumLatitude = math.Max(maximumLatitude, latitude)
		moonRadius := MoonSemidiameter(tt)
		if observer != nil {
			moonRadius = moonTopocentricSemidiameterN(tt, *observer, -1)
		}
		maximumMoonRadiusDeg = math.Max(maximumMoonRadiusDeg, moonRadius/3600)
	}
	limit := starOccultationMoonLatitudeDeg + starOccultationLatitudeMarginAS/3600 + safetyMarginArcsec/3600 + maximumMoonRadiusDeg
	return minimumLatitude <= limit && maximumLatitude >= -limit
}

func starOccultationGeocentricLongitudeCandidate(startTT, endTT, step float64, star StarCoordinate) float64 {
	bestTT := math.NaN()
	bestDelta := math.Inf(1)
	for tt := startTT; tt <= endTT; tt += step {
		delta := math.Abs(signedAngleDifference(HMoonTrueLoN(tt, 8), starOccultationGeocentricStarLongitude(tt, star)))
		if delta < bestDelta {
			bestDelta = delta
			bestTT = tt
		}
	}
	if endTT > startTT {
		delta := math.Abs(signedAngleDifference(HMoonTrueLoN(endTT, 8), starOccultationGeocentricStarLongitude(endTT, star)))
		if delta < bestDelta {
			bestTT = endTT
		}
	}
	return bestTT
}

func starOccultationGeocentricStarLongitude(tt float64, star StarCoordinate) float64 {
	ra, dec := starApparentRaDecGeocentric(tt, star)
	longitude, _ := RaDecToLoBo(tt, ra, dec)
	return longitude
}

func starOccultationMinimizeGeocentricSeparation(seed, startTT, endTT float64, star StarCoordinate) float64 {
	halfWindow := 0.75
	left := math.Max(startTT, seed-halfWindow)
	right := math.Min(endTT, seed+halfWindow)
	return starOccultationMinimizeValue(left, right, func(tt float64) float64 {
		return starOccultationGeocentricSeparationArcsec(tt, star)
	})
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

func starOccultationGeocentricSeparationArcsec(tt float64, star StarCoordinate) float64 {
	moonRA, moonDec := HMoonGeocentricApparentRaDecN(tt, -1)
	starRA, starDec := starApparentRaDecGeocentric(tt, star)
	return angularSeparationDegrees(moonRA, moonDec, starRA, starDec) * 3600
}

func starOccultationBestObserver(seedTT, startTT, endTT float64, star StarCoordinate) (float64, Observer, float64, bool) {
	searchStart := seedTT - occultationPathSearchSpanDays
	searchEnd := seedTT + occultationPathSearchSpanDays
	outerStart, outerEnd, ok := starOccultationPathWindow(seedTT, searchStart, searchEnd, star, false)
	if !ok {
		return 0, Observer{}, 0, false
	}
	greatestTT := starOccultationPathGreatest(seedTT, outerStart, outerEnd, star)
	if greatestTT < startTT || greatestTT > endTT {
		return 0, Observer{}, 0, false
	}
	point, pointOK := starOccultationPathCenterPoint(greatestTT, star, time.UTC)
	if !pointOK {
		frameAt := func(tt float64) (occultationPathFrame, bool) {
			return starOccultationPathFrameAt(tt, star)
		}
		point, pointOK = occultationPathBoundaryPointForFrame(greatestTT, frameAt, time.UTC)
	}
	if !pointOK {
		return 0, Observer{}, 0, false
	}
	observer := Observer{Longitude: point.Longitude, Latitude: point.Latitude}
	position := starMoonPositionAt(greatestTT, star, observer)
	moonRadius := moonTopocentricSemidiameterN(greatestTT, observer, -1)
	if !position.valid || !finite(moonRadius) {
		return 0, Observer{}, 0, false
	}
	metric := angularSeparationDegrees(position.moonRA, position.moonDec, position.starRA, position.starDec)*3600 - moonRadius
	return greatestTT, observer, metric, finite(metric)
}

func normalizeLongitude180(longitude float64) float64 {
	longitude = math.Mod(longitude+180, 360)
	if longitude < 0 {
		longitude += 360
	}
	return longitude - 180
}

func starOccultationInfoAtGreatest(greatestTT float64, star StarCoordinate, observer Observer, safetyMarginArcsec float64, location *time.Location) (StarOccultationInfo, bool) {
	if !starOccultationLatitudePass(greatestTT, star, observer, safetyMarginArcsec) {
		return StarOccultationInfo{}, false
	}
	minimumSeparation := starMoonSeparationArcsec(greatestTT, star, observer)
	moonRadius := moonTopocentricSemidiameterN(greatestTT, observer, -1)
	if !finite(minimumSeparation) || !finite(moonRadius) || minimumSeparation > moonRadius {
		return StarOccultationInfo{}, false
	}
	greatestPosition := starMoonPositionAt(greatestTT, star, observer)
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

	minimumResidual := minimumSeparation - moonRadius
	if math.Abs(minimumResidual) <= starOccultationGrazingTolerance {
		info.Type = OccultationGrazing
		info.Immersion = info.Greatest
		info.Emersion = info.Greatest
		info.ContactsComplete = true
		return info, true
	}
	immersionTT, immersionOK := starOccultationContact(greatestTT, greatestTT-starOccultationContactSpanDays, -1, star, observer)
	emersionTT, emersionOK := starOccultationContact(greatestTT, greatestTT+starOccultationContactSpanDays, 1, star, observer)
	if !immersionOK || !emersionOK {
		return StarOccultationInfo{}, false
	}
	info.Immersion = occultationTTToLocation(immersionTT, location)
	info.Emersion = occultationTTToLocation(emersionTT, location)
	info.ContactsComplete = true
	return info, true
}

type starMoonPosition struct {
	moonRA, moonDec float64
	starRA, starDec float64
	valid           bool
}

func starMoonPositionAt(tt float64, star StarCoordinate, observer Observer) starMoonPosition {
	moonRA, moonDec := moonTopocentricApparentRaDec(tt, observer, -1)
	starRA, starDec := starApparentRaDec(tt, star, observer)
	return starMoonPosition{
		moonRA:  moonRA,
		moonDec: moonDec,
		starRA:  starRA,
		starDec: starDec,
		valid:   finite(moonRA) && finite(moonDec) && finite(starRA) && finite(starDec),
	}
}

func moonTopocentricApparentRaDec(tt float64, observer Observer, n int) (float64, float64) {
	ra, dec := HMoonGeocentricApparentRaDecN(tt, n)
	ut := TD2UT(tt, false)
	distanceAU := HMoonAwayN(tt, n) / 149597870.7
	ra, dec = TopocentricRaDec(ra, dec, observer.Latitude, observer.Longitude, ut, distanceAU, observer.Height)
	return normalizeRA(ra), dec
}

func starApparentRaDec(tt float64, star StarCoordinate, observer Observer) (float64, float64) {
	ra, dec := starApparentRaDecGeocentric(tt, star)
	if star.ParallaxMas > 0 {
		// 1 秒差距处 1 角秒对应 206264.806 AU。
		// One arcsecond at 1 pc corresponds to 206264.806 AU.
		distanceAU := 206264806.247 / star.ParallaxMas
		ra, dec = TopocentricRaDec(ra, dec, observer.Latitude, observer.Longitude, TD2UT(tt, false), distanceAU, observer.Height)
		ra = normalizeRA(ra)
	}
	return ra, dec
}

func starApparentRaDecGeocentric(tt float64, star StarCoordinate) (float64, float64) {
	epochJD := occultationTimeToTT(star.Epoch)
	years := (tt - epochJD) / 365.25
	ra := star.RA
	dec := star.Dec
	precessionEpoch := 2451545.0
	if star.Frame == CoordinateFrameICRS {
		ra, dec = starICRSToMeanJ2000RaDec(ra, dec)
	} else if star.Frame == CoordinateFrameApparentOfDate {
		ra, dec = starApparentToMeanRaDec(epochJD, ra, dec, star.ParallaxMas)
		precessionEpoch = epochJD
	}
	cosDec := math.Cos(dec * math.Pi / 180)
	if math.Abs(cosDec) > 1e-12 {
		ra += years * star.ProperMotionRACosDecMasPerYear / (3600000.0 * cosDec)
	}
	dec += years * star.ProperMotionDecMasPerYear / 3600000.0
	dec = math.Max(-90, math.Min(90, dec))

	ra, dec = Precess(ra, dec, precessionEpoch, tt)
	return starMeanToApparentRaDec(tt, ra, dec, star.ParallaxMas)
}

func starICRSToMeanJ2000RaDec(ra, dec float64) (float64, float64) {
	// IAU SOFA 框架偏差矩阵，将 GCRS/ICRS 向量转换为 J2000.0 平均赤道和春分点。
	// IAU SOFA frame-bias matrix, transforming a GCRS/ICRS vector to the mean equator and equinox of J2000.0.
	const (
		b00 = 0.9999999999999942
		b01 = -0.7078279744199197e-7
		b02 = 0.8056217146976134e-7
		b10 = 0.7078279477857337e-7
		b11 = 0.9999999999999969
		b12 = 0.3306041454222148e-7
		b20 = -0.8056217380986972e-7
		b21 = -0.3306040883980553e-7
		b22 = 0.9999999999999962
	)
	raRad := ra * math.Pi / 180
	decRad := dec * math.Pi / 180
	x := math.Cos(decRad) * math.Cos(raRad)
	y := math.Cos(decRad) * math.Sin(raRad)
	z := math.Sin(decRad)
	biasedX := b00*x + b01*y + b02*z
	biasedY := b10*x + b11*y + b12*z
	biasedZ := b20*x + b21*y + b22*z
	return normalizeRA(math.Atan2(biasedY, biasedX) * 180 / math.Pi),
		math.Atan2(biasedZ, math.Hypot(biasedX, biasedY)) * 180 / math.Pi
}

func starMeanToApparentRaDec(tt, ra, dec, parallaxMas float64) (float64, float64) {
	longitude, latitude := starMeanEquatorialToEcliptic(tt, ra, dec)
	if parallaxMas > 0 {
		longitude, latitude = starAnnualParallaxEcliptic(tt, longitude, latitude, parallaxMas)
	}
	meanLongitude, meanLatitude := longitude, latitude
	longitude = normalizeRA(meanLongitude + GXCLo(meanLongitude, meanLatitude, tt)/3600 + Nutation2000Bi(tt))
	latitude = meanLatitude + GXCBo(meanLongitude, meanLatitude, tt)/3600
	ra, dec = LoBoToRaDec(tt, longitude, latitude)
	return normalizeRA(ra), dec
}

func starApparentToMeanRaDec(tt, apparentRA, apparentDec, parallaxMas float64) (float64, float64) {
	meanRA, meanDec := apparentRA, apparentDec
	for i := 0; i < 8; i++ {
		computedRA, computedDec := starMeanToApparentRaDec(tt, meanRA, meanDec, parallaxMas)
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

func starAnnualParallaxEcliptic(tt, longitude, latitude, parallaxMas float64) (float64, float64) {
	distanceAU := 206264806.247 / parallaxMas
	longitudeRad := longitude * math.Pi / 180
	latitudeRad := latitude * math.Pi / 180
	cosLatitude := math.Cos(latitudeRad)
	starX := distanceAU * cosLatitude * math.Cos(longitudeRad)
	starY := distanceAU * cosLatitude * math.Sin(longitudeRad)
	starZ := distanceAU * math.Sin(latitudeRad)

	earthLongitude := normalizeRA(HSunTrueLoN(tt, -1)+180) * math.Pi / 180
	earthDistance := EarthAwayN(tt, -1)
	starX -= earthDistance * math.Cos(earthLongitude)
	starY -= earthDistance * math.Sin(earthLongitude)

	longitude = math.Atan2(starY, starX) * 180 / math.Pi
	latitude = math.Atan2(starZ, math.Hypot(starX, starY)) * 180 / math.Pi
	return normalizeRA(longitude), latitude
}

func starOccultationLongitudeCandidate(startTT, endTT, step float64, star StarCoordinate, observer Observer) float64 {
	bestTT := math.NaN()
	bestDelta := math.Inf(1)
	for tt := startTT; tt <= endTT; tt += step {
		delta := starOccultationLongitudeDistance(tt, star, observer)
		if delta < bestDelta {
			bestDelta = delta
			bestTT = tt
		}
	}
	if endTT > startTT {
		delta := starOccultationLongitudeDistance(endTT, star, observer)
		if delta < bestDelta {
			bestTT = endTT
		}
	}
	return bestTT
}

func starOccultationLongitudeDistance(tt float64, star StarCoordinate, observer Observer) float64 {
	moonLongitude := HMoonTrueLoN(tt, 8)
	starRA, starDec := starApparentRaDec(tt, star, observer)
	starLongitude, _ := RaDecToLoBo(tt, starRA, starDec)
	return math.Abs(signedAngleDifference(moonLongitude, starLongitude))
}

func starOccultationMinimizeSeparation(seed, startTT, endTT float64, star StarCoordinate, observer Observer) float64 {
	halfWindow := 0.75
	left := math.Max(startTT, seed-halfWindow)
	right := math.Min(endTT, seed+halfWindow)
	return starOccultationMinimizeValue(left, right, func(tt float64) float64 {
		return starMoonSeparationArcsec(tt, star, observer)
	})
}

func starMoonSeparationArcsec(tt float64, star StarCoordinate, observer Observer) float64 {
	position := starMoonPositionAt(tt, star, observer)
	if !position.valid {
		return math.Inf(1)
	}
	return angularSeparationDegrees(position.moonRA, position.moonDec, position.starRA, position.starDec) * 3600
}

func starOccultationLatitudePass(tt float64, star StarCoordinate, observer Observer, safetyMarginArcsec float64) bool {
	starRA, starDec := starApparentRaDec(tt, star, observer)
	_, starLatitude := RaDecToLoBo(tt, starRA, starDec)
	moonLatitude := HMoonTrueBoN(tt, 8)
	moonRadius := moonTopocentricSemidiameterN(tt, observer, -1)
	limit := moonRadius + starOccultationLatitudeMarginAS + safetyMarginArcsec
	return math.Abs(starLatitude-moonLatitude)*3600 <= limit
}

func starOccultationLatitudePassGeocentric(tt float64, star StarCoordinate, safetyMarginArcsec float64) bool {
	starRA, starDec := starApparentRaDecGeocentric(tt, star)
	_, starLatitude := RaDecToLoBo(tt, starRA, starDec)
	moonLatitude := HMoonTrueBoN(tt, 8)
	limit := MoonSemidiameter(tt) + starOccultationLatitudeMarginAS + safetyMarginArcsec
	return math.Abs(starLatitude-moonLatitude)*3600 <= limit
}

func starOccultationContact(greatestTT, boundaryTT float64, direction int, star StarCoordinate, observer Observer) (float64, bool) {
	valueAtGreatest := starMoonSeparationArcsec(greatestTT, star, observer) - moonTopocentricSemidiameterN(greatestTT, observer, -1)
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
		nextValue := starMoonSeparationArcsec(nextTT, star, observer) - moonTopocentricSemidiameterN(nextTT, observer, -1)
		if finite(nextValue) && nextValue >= 0 {
			return starOccultationRoot(currentTT, nextTT, currentValue, nextValue, star, observer)
		}
		if nextTT == boundaryTT {
			return math.NaN(), false
		}
		currentTT = nextTT
		currentValue = nextValue
	}
	return math.NaN(), false
}

func starOccultationRoot(leftTT, rightTT, leftValue, rightValue float64, star StarCoordinate, observer Observer) (float64, bool) {
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
		midValue := starMoonSeparationArcsec(midTT, star, observer) - moonTopocentricSemidiameterN(midTT, observer, -1)
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
	hourAngle := signedAngleDifference(ApparentSiderealTime(TD2UT(tt, false))*15+observer.Longitude, ra) * math.Pi / 180
	lat := observer.Latitude * math.Pi / 180
	declination := dec * math.Pi / 180
	sinAltitude := math.Sin(lat)*math.Sin(declination) + math.Cos(lat)*math.Cos(declination)*math.Cos(hourAngle)
	return math.Asin(math.Max(-1, math.Min(1, sinAltitude))) * 180 / math.Pi
}

func occultationAzimuth(tt float64, observer Observer, ra, dec float64) float64 {
	hourAngle := signedAngleDifference(ApparentSiderealTime(TD2UT(tt, false))*15+observer.Longitude, ra) * math.Pi / 180
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
	return TD2UT(Date2JDE(value.UTC()), true)
}

func occultationTTToLocation(tt float64, location *time.Location) time.Time {
	if location == nil {
		location = time.UTC
	}
	return JDE2DateByZone(TD2UT(tt, false), location, false)
}
