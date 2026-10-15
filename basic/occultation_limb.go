package basic

import (
	"math"
	"time"
)

// occultationLimbTarget 给出某时刻目标的地心视赤经、赤纬（度）与地心距（km，恒星为 0）及站心角半径（角秒）。
type occultationLimbTarget func(tt float64) (ra, dec, distanceKM, radiusArcsec float64, ok bool)

func planetOccultationLimbTarget(config planetOccultationConfig, observer Observer) occultationLimbTarget {
	return func(tt float64) (float64, float64, float64, float64, bool) {
		ra, dec := config.apparentRaDecN(tt, -1)
		distanceKM := config.earthDistanceN(tt, -1) * angularDiameterAstronomicalUnitKM
		radius := planetTopocentricSemidiameterN(tt, config, observer, -1)
		return ra, dec, distanceKM, radius, finite(ra) && finite(dec) && finite(distanceKM) && finite(radius)
	}
}

const (
	occultationLimbCrossingStepDays = 60.0 / 86400.0
	occultationLimbCrossingSpanDays = 3.5 / 24.0
	occultationLimbRootDays         = 1e-3 / 86400.0
	occultationLimbMinSamples       = 64
	occultationLimbMaxSamples       = 4000
	occultationLimbBandPadArcsec    = 0.05
	occultationLimbUnrangedPadKM    = 1.0
	planetOccultationLimbStepDays   = 0.25 / 86400.0
)

// occultationLimbRangeKM 返回剖面半径范围；剖面不报告范围时在 tt 时刻朝目标扫一圈并外扩 1 km。
func occultationLimbRangeKM(tt float64, limb LunarLimb, observer Observer, target occultationLimbTarget, inst *Instance) (float64, float64, bool) {
	geoRA, geoDec := HMoonGeocentricApparentRaDecNWithInstance(tt, -1, inst)
	geoDistance := HMoonAwayNWithInstance(tt, -1, inst)
	frame := newLunarLimbFrame(limb, tt, lunarEclipseLimbVector(geoRA, geoDec, geoDistance), inst)
	if !frame.valid {
		return 0, 0, false
	}
	if frame.ranged {
		return frame.minKM, frame.maxKM, true
	}
	ra, dec, _, _, ok := target(tt)
	if !ok {
		return 0, 0, false
	}
	topoRA, topoDec := moonTopocentricApparentRaDec(tt, observer, -1, inst)
	view, ok := frame.viewToward(lunarEclipseLimbVector(topoRA, topoDec, 1), lunarEclipseLimbVector(ra, dec, 1))
	if !ok {
		return 0, 0, false
	}
	low, high := math.Inf(1), math.Inf(-1)
	for offset := 0.0; offset < 360; offset += frame.stepDeg {
		radius, ok := frame.radiusAt(view, offset)
		if !ok {
			return 0, 0, false
		}
		low, high = math.Min(low, radius), math.Max(high, radius)
	}
	return low - occultationLimbUnrangedPadKM, high + occultationLimbUnrangedPadKM, true
}

// occultationLimbCrossing 自 greatestTT 向 direction 一侧求圆月面间隙升到 level 的时刻；间隙自掩甚单调增大，起点已不低于 level 时返回起点。
func occultationLimbCrossing(greatestTT, direction, level float64, circular func(float64) (float64, bool)) (float64, bool) {
	near := greatestTT
	value, ok := circular(near)
	if !ok {
		return 0, false
	}
	if value >= level {
		return greatestTT, true
	}
	for step := 1; float64(step)*occultationLimbCrossingStepDays <= occultationLimbCrossingSpanDays; step++ {
		far := greatestTT + direction*float64(step)*occultationLimbCrossingStepDays
		farValue, ok := circular(far)
		if !ok {
			return 0, false
		}
		if farValue < level {
			near = far
			continue
		}
		for i := 0; i < 64 && math.Abs(far-near) > occultationLimbRootDays; i++ {
			mid := (near + far) / 2
			midValue, ok := circular(mid)
			if !ok {
				return 0, false
			}
			if midValue < level {
				near = mid
			} else {
				far = mid
			}
		}
		return (near + far) / 2, true
	}
	return 0, false
}

// occultationLimbOuterRoots 在 [start, end] 上采样月缘间隙，返回首个由正转负与最后一个由负转正的零点；
// 端点状态由调用方保证（hiddenAtStart、hiddenAtEnd），ok 为假表示剖面无值或星历失败。
func occultationLimbOuterRoots(start, end float64, hiddenAtStart, hiddenAtEnd bool, gap func(float64) (float64, bool)) (first, last float64, found, ok bool) {
	if end <= start {
		return 0, 0, false, true
	}
	count := int(math.Ceil((end - start) / planetOccultationLimbStepDays))
	if count < occultationLimbMinSamples {
		count = occultationLimbMinSamples
	}
	if count > occultationLimbMaxSamples {
		count = occultationLimbMaxSamples
	}
	first, last = math.NaN(), math.NaN()
	previousTT, previousHidden := start, hiddenAtStart
	for index := 0; index <= count; index++ {
		tt := start + (end-start)*float64(index)/float64(count)
		value, valueOK := gap(tt)
		if !valueOK {
			return 0, 0, false, false
		}
		hidden := value < 0
		if index == 0 {
			hidden = hiddenAtStart
		} else if index == count {
			hidden = hiddenAtEnd
		}
		if index > 0 && hidden != previousHidden {
			root := occultationLimbBisect(previousTT, tt, previousHidden, gap)
			if hidden && math.IsNaN(first) {
				first = root
			}
			if !hidden {
				last = root
			}
		}
		previousTT, previousHidden = tt, hidden
	}
	return first, last, !math.IsNaN(first) || !math.IsNaN(last), true
}

// occultationLimbBisect 在相邻样本间求遮挡状态翻转的时刻；leftHidden 是左端的遮挡状态。
func occultationLimbBisect(left, right float64, leftHidden bool, gap func(float64) (float64, bool)) float64 {
	for i := 0; i < 64 && right-left > occultationLimbRootDays; i++ {
		mid := (left + right) / 2
		value, ok := gap(mid)
		if !ok {
			break
		}
		if (value < 0) == leftHidden {
			left = mid
		} else {
			right = mid
		}
	}
	return (left + right) / 2
}

// planetOccultationInfoWithLimb 有月缘时按剖面判定行星月掩：只在圆月面间隙落进月缘起伏带的时段采样全方向月缘间隙，
// 取最外侧的外切与内切；月缘可新增或取消圆月面的掠掩与全掩，有正持续时间的接触一律按偏掩或全掩上报。剖面无值时退回圆月面结果。
func planetOccultationInfoWithLimb(greatestTT float64, config planetOccultationConfig, observer Observer, safetyMarginArcsec float64, location *time.Location, limb LunarLimb, inst *Instance) (PlanetOccultationInfo, bool) {
	circularInfo, circularOK := planetOccultationInfoAtGreatest(greatestTT, config, observer, safetyMarginArcsec, location, inst)
	if limb == nil {
		return circularInfo, circularOK
	}
	info, state, ok := planetOccultationGreatestInfo(greatestTT, config, observer, safetyMarginArcsec, location, inst)
	if !ok {
		return circularInfo, circularOK
	}
	target := planetOccultationLimbTarget(config, observer)
	lowKM, highKM, ok := occultationLimbRangeKM(greatestTT, limb, observer, target, inst)
	if !ok {
		return circularInfo, circularOK
	}
	ephemeris, ok := newPlanetLimbEphemeris(greatestTT, config, limb, inst)
	if !ok {
		return circularInfo, circularOK
	}
	// 月缘间隙 ≈ 圆月面间隙 −（月缘半径 − 参考半径），所以圆月面间隙不低于 bandHigh 必不接触、不高于 bandLow 必已接触。
	// 起伏带以角秒计，外扩一点容差覆盖月心距在事件内的变化。
	scale := state.moonSemidiameter / moonEquatorialRadiusKM
	bandLow := (lowKM-moonEquatorialRadiusKM)*scale - occultationLimbBandPadArcsec
	bandHigh := (highKM-moonEquatorialRadiusKM)*scale + occultationLimbBandPadArcsec
	circular := func(internal bool) func(float64) (float64, bool) {
		return func(tt float64) (float64, bool) { return ephemeris.circularGap(tt, observer, internal) }
	}
	limbGap := func(internal bool) func(float64) (float64, bool) {
		return func(tt float64) (float64, bool) { return ephemeris.limbGap(tt, observer, internal) }
	}
	// solve 求一种接触的最外侧掩入、复现：圆月面间隙在 [bandLow, bandHigh] 之外的两端遮挡状态已知，只在带内采样。
	solve := func(internal bool, circularMinimum float64) (float64, float64, bool, bool) {
		if circularMinimum >= bandHigh {
			return 0, 0, false, true
		}
		outerLeft, okLeft := occultationLimbCrossing(greatestTT, -1, bandHigh, circular(internal))
		outerRight, okRight := occultationLimbCrossing(greatestTT, 1, bandHigh, circular(internal))
		if !okLeft || !okRight {
			return 0, 0, false, false
		}
		if circularMinimum > bandLow {
			first, last, found, ok := occultationLimbOuterRoots(outerLeft, outerRight, false, false, limbGap(internal))
			return first, last, found && !math.IsNaN(first) && !math.IsNaN(last), ok
		}
		innerLeft, okLeft := occultationLimbCrossing(greatestTT, -1, bandLow, circular(internal))
		innerRight, okRight := occultationLimbCrossing(greatestTT, 1, bandLow, circular(internal))
		if !okLeft || !okRight {
			return 0, 0, false, false
		}
		first, _, foundLeft, okLeftRoots := occultationLimbOuterRoots(outerLeft, innerLeft, false, true, limbGap(internal))
		_, last, foundRight, okRightRoots := occultationLimbOuterRoots(innerRight, outerRight, true, false, limbGap(internal))
		return first, last, foundLeft && foundRight && !math.IsNaN(first) && !math.IsNaN(last), okLeftRoots && okRightRoots
	}
	externalImmersion, externalEmersion, external, ok := solve(false, state.externalContactMetric)
	if !ok {
		return circularInfo, circularOK
	}
	if !external {
		return PlanetOccultationInfo{}, false
	}
	internalImmersion, internalEmersion, internal, ok := solve(true, state.internalContactMetric)
	if !ok {
		return circularInfo, circularOK
	}
	info.Type = OccultationPartial
	info.ExternalImmersion = occultationTTToLocation(externalImmersion, location)
	info.ExternalEmersion = occultationTTToLocation(externalEmersion, location)
	if internal && internalImmersion > externalImmersion && internalEmersion < externalEmersion && internalImmersion < internalEmersion {
		info.Type = OccultationTotal
		info.HasInternalContacts = true
		info.InternalImmersion = occultationTTToLocation(internalImmersion, location)
		info.InternalEmersion = occultationTTToLocation(internalEmersion, location)
	}
	info.ContactsComplete = true
	info.LunarLimbApplied = true
	return info, true
}

// planetLimbEphemeris 掩甚前后 ±4 h 的地心月球、行星视位置与月球本体坐标架插值表；站心换算逐点精确计算。
type planetLimbEphemeris struct {
	bodies   []localEphemerisVectorNode
	axes     []localEphemerisVectorNode
	base     lunarLimbFrame
	radiusKM float64
}

func newPlanetLimbEphemeris(greatestTT float64, config planetOccultationConfig, limb LunarLimb, inst *Instance) (*planetLimbEphemeris, bool) {
	half := int(math.Round(starOccultationLimbNodeSpan / starOccultationLimbNodeDays))
	ephemeris := &planetLimbEphemeris{
		bodies:   make([]localEphemerisVectorNode, 2*half+1),
		axes:     make([]localEphemerisVectorNode, 2*half+1),
		radiusKM: config.equatorialRadiusKM,
	}
	for index := range ephemeris.bodies {
		tt := greatestTT + float64(index-half)*starOccultationLimbNodeDays
		moonRA, moonDec := HMoonGeocentricApparentRaDecNWithInstance(tt, -1, inst)
		moon := lunarEclipseLimbVector(moonRA, moonDec, HMoonAwayNWithInstance(tt, -1, inst))
		planetRA, planetDec := config.apparentRaDecN(tt, -1)
		planet := lunarEclipseLimbVector(planetRA, planetDec, config.earthDistanceN(tt, -1)*angularDiameterAstronomicalUnitKM)
		frame := newLunarLimbFrame(limb, tt, moon, inst)
		if !frame.valid || !finiteVector3(moon) || !finiteVector3(planet) {
			return nil, false
		}
		if index == half {
			ephemeris.base = frame
		}
		ephemeris.bodies[index] = localEphemerisVectorNode{tt: tt, first: moon, next: planet}
		ephemeris.axes[index] = localEphemerisVectorNode{tt: tt, first: frame.axisX, next: frame.axisZ}
	}
	return ephemeris, true
}

// planetLimbTopocentric 是插值得到的站心月球、行星方向与距离（km），以及该时刻的月球本体坐标架。
type planetLimbTopocentric struct {
	moon, planet               [3]float64
	moonDistance, planetRadius float64
	frame                      lunarLimbFrame
}

func (ephemeris *planetLimbEphemeris) topocentricAt(tt float64, observer Observer) (planetLimbTopocentric, bool) {
	moon, planet, ok := interpolateLocalEphemerisVectors(ephemeris.bodies, tt)
	if !ok {
		return planetLimbTopocentric{}, false
	}
	axisX, axisZ, ok := interpolateLocalEphemerisVectors(ephemeris.axes, tt)
	if !ok {
		return planetLimbTopocentric{}, false
	}
	ut := TT2UTC(tt)
	toTopocentric := func(vector [3]float64) (float64, float64, float64, bool) {
		distance := math.Sqrt(lunarLimbDot(vector, vector))
		ra := normalizeRA(math.Atan2(vector[1], vector[0]) / rad)
		dec := math.Asin(clampUnit(vector[2]/distance)) / rad
		topoRA, topoDec := TopocentricRaDec(ra, dec, observer.Latitude, observer.Longitude, ut, distance/angularDiameterAstronomicalUnitKM, observer.Height)
		topoDistance := topocentricDistanceKM(ra, dec, distance, observer, ut)
		return normalizeRA(topoRA), topoDec, topoDistance, finite(topoRA) && finite(topoDec) && finite(topoDistance) && topoDistance > 0
	}
	moonRA, moonDec, moonDistance, okMoon := toTopocentric(moon)
	planetRA, planetDec, planetDistance, okPlanet := toTopocentric(planet)
	if !okMoon || !okPlanet || planetDistance <= ephemeris.radiusKM {
		return planetLimbTopocentric{}, false
	}
	frame := ephemeris.base
	if frame.axisZ, ok = lunarLimbNormalize(axisZ); !ok {
		return planetLimbTopocentric{}, false
	}
	if frame.axisX, ok = lunarLimbNormalize(lunarLimbReject(axisX, frame.axisZ)); !ok {
		return planetLimbTopocentric{}, false
	}
	frame.axisY = lunarLimbCross(frame.axisZ, frame.axisX)
	return planetLimbTopocentric{
		moon:         lunarEclipseLimbVector(moonRA, moonDec, moonDistance),
		planet:       lunarEclipseLimbVector(planetRA, planetDec, 1),
		moonDistance: moonDistance,
		planetRadius: math.Asin(ephemeris.radiusKM / planetDistance),
		frame:        frame,
	}, true
}

// circularGap 圆月面外切或内切间隙（角秒），与 planetOccultationStateAt 同口径。
func (ephemeris *planetLimbEphemeris) circularGap(tt float64, observer Observer, internal bool) (float64, bool) {
	state, ok := ephemeris.topocentricAt(tt, observer)
	if !ok {
		return 0, false
	}
	separation := math.Acos(clampUnit(lunarLimbDot(state.moon, state.planet) / state.moonDistance))
	moonRadius := math.Asin(moonEquatorialRadiusKM / state.moonDistance)
	if internal {
		return (separation - moonRadius + state.planetRadius) * angularDiameterArcsecPerRadian, true
	}
	return (separation - moonRadius - state.planetRadius) * angularDiameterArcsecPerRadian, true
}

// limbGap 全方向月缘外切或内切间隙（角秒）。
func (ephemeris *planetLimbEphemeris) limbGap(tt float64, observer Observer, internal bool) (float64, bool) {
	state, ok := ephemeris.topocentricAt(tt, observer)
	if !ok {
		return 0, false
	}
	view, ok := state.frame.viewToward(state.moon, state.planet)
	if !ok {
		return 0, false
	}
	contact := lunarLimbContact{
		frame:        &state.frame,
		view:         view,
		distanceKM:   state.moonDistance,
		separation:   math.Acos(clampUnit(lunarLimbDot(state.moon, state.planet) / state.moonDistance)),
		targetRadius: state.planetRadius,
	}
	if internal {
		inner, ok := contact.innerRadius(true)
		if !ok {
			return 0, false
		}
		return (contact.separation - math.Abs(inner-contact.targetRadius)) * angularDiameterArcsecPerRadian, true
	}
	outer, ok := contact.outerRadius()
	if !ok {
		return 0, false
	}
	return (contact.separation - outer - contact.targetRadius) * angularDiameterArcsecPerRadian, true
}
