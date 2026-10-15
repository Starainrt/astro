package basic

import (
	"fmt"
	"math"
	"sort"
	"time"
)

const (
	starOccultationLimbDefaultStep   = 200 * time.Millisecond
	starOccultationLimbMinimumStep   = 10 * time.Millisecond
	starOccultationLimbMaxSamples    = 20000
	starOccultationLimbMinSamples    = 64
	starOccultationLimbBoundStep     = 60.0 / 86400.0
	starOccultationLimbBoundSpan     = 3.5 / 24.0
	starOccultationLimbNodeDays      = 10.0 / 1440.0
	starOccultationLimbNodeSpan      = 4.0 / 24.0
	starOccultationLimbRootDays      = 1e-3 / 86400.0
	starOccultationLimbUnrangedPadKM = 1.0
)

// StarOccultationLimbProfileOptions 控制精密恒星月掩的星迹采样 / sampling of a precise stellar occultation.
type StarOccultationLimbProfileOptions struct {
	// Step 是星迹采样步长；零值 200 ms，小于 10 ms 提升到 10 ms；每段至少 64 个样本、总数不超过 20000，越界时自动调整。短于步长的遮挡或复现可能漏检 / track sampling step.
	Step time.Duration
}

// StarOccultationLimbSample 星迹上一个时刻的月缘几何，角量均为角秒、度 / limb geometry at one instant on the star track.
type StarOccultationLimbSample struct {
	Time time.Time
	// SeparationArcsec 恒星到站心月心的角距 / star–Moon-centre separation.
	SeparationArcsec float64
	// LimbRadiusArcsec 恒星方向上的月缘角半径 / limb radius toward the star.
	LimbRadiusArcsec float64
	// MeanRadiusArcsec 参考圆月面角半径 / circular reference radius.
	MeanRadiusArcsec float64
	// PositionAngleDeg 恒星相对月心的位置角，自天北向东 / position angle of the star.
	PositionAngleDeg float64
	// Hidden 恒星位于月缘以内 / star is behind the limb.
	Hidden bool
}

// OccultationInterval 一段连续遮挡 / one continuous hidden interval.
type OccultationInterval struct {
	Start time.Time
	End   time.Time
}

// StarOccultationLimbProfileResult 一次恒星月掩沿星迹的月缘剖面与全部遮挡区间 / limb profile and every hidden interval of a stellar occultation.
type StarOccultationLimbProfileResult struct {
	// Occultation 按同一剖面求得的标准结果，掩始、掩终为首次遮挡与最后复现 / standard result on the same profile.
	Occultation StarOccultationInfo
	// Segments 月缘起伏带内的采样段，按时间排序；深掩时为掩入侧与复现侧两段，掠掩时通常为一段 / samples inside the limb relief band.
	Segments [][]StarOccultationLimbSample
	// Hidden 全部遮挡区间，按时间排序 / every hidden interval.
	Hidden []OccultationInterval
	// ReferenceRadiusKM 样本 MeanRadiusArcsec 对应的参考月球半径 / reference lunar radius behind MeanRadiusArcsec.
	ReferenceRadiusKM float64
	// ReliefMinKM / ReliefMaxKM 判定所用的月缘半径范围 / limb radius range used for gating.
	ReliefMinKM float64
	ReliefMaxKM float64
	// Step 实际采用的采样步长 / effective step.
	Step time.Duration
}

type starLimbState struct {
	separation, limbRadius, meanRadius, positionAngle float64
	distanceKM                                        float64
}

// starLimbEphemeris 掩甚前后 ±4 h 的地心月球、恒星视位置与月球本体坐标架插值表；站心换算逐点精确计算。
type starLimbEphemeris struct {
	bodies      []localEphemerisVectorNode
	axes        []localEphemerisVectorNode
	base        lunarLimbFrame
	hasDistance bool
}

func newStarLimbEphemeris(greatestTT float64, star StarCoordinate, limb LunarLimb, inst *Instance) (*starLimbEphemeris, bool) {
	half := int(math.Round(starOccultationLimbNodeSpan / starOccultationLimbNodeDays))
	ephemeris := &starLimbEphemeris{
		bodies: make([]localEphemerisVectorNode, 2*half+1),
		axes:   make([]localEphemerisVectorNode, 2*half+1),
	}
	for index := range ephemeris.bodies {
		tt := greatestTT + float64(index-half)*starOccultationLimbNodeDays
		moonRA, moonDec := HMoonGeocentricApparentRaDecNWithInstance(tt, -1, inst)
		moon := lunarEclipseLimbVector(moonRA, moonDec, HMoonAwayNWithInstance(tt, -1, inst))
		starRA, starDec, distanceAU := starApparentRaDecDistanceGeocentric(tt, star, inst)
		if distanceAU <= 0 {
			distanceAU = 1
		} else {
			ephemeris.hasDistance = true
		}
		frame := newLunarLimbFrame(limb, tt, moon, inst)
		if !frame.valid || !finiteVector3(moon) {
			return nil, false
		}
		if index == half {
			ephemeris.base = frame
		}
		ephemeris.bodies[index] = localEphemerisVectorNode{tt: tt, first: moon, next: lunarEclipseLimbVector(starRA, starDec, distanceAU)}
		ephemeris.axes[index] = localEphemerisVectorNode{tt: tt, first: frame.axisX, next: frame.axisZ}
	}
	return ephemeris, true
}

func (ephemeris *starLimbEphemeris) stateAt(tt float64, observer Observer) (starLimbState, bool) {
	moon, starVector, ok := interpolateLocalEphemerisVectors(ephemeris.bodies, tt)
	if !ok {
		return starLimbState{}, false
	}
	axisX, axisZ, ok := interpolateLocalEphemerisVectors(ephemeris.axes, tt)
	if !ok {
		return starLimbState{}, false
	}
	geoDistance := math.Sqrt(lunarLimbDot(moon, moon))
	geoRA := normalizeRA(math.Atan2(moon[1], moon[0]) / rad)
	geoDec := math.Asin(clampUnit(moon[2]/geoDistance)) / rad
	ut := TT2UTC(tt)
	topoRA, topoDec := TopocentricRaDec(geoRA, geoDec, observer.Latitude, observer.Longitude, ut, geoDistance/149597870.7, observer.Height)
	topoRA = normalizeRA(topoRA)
	topoDistance := topocentricDistanceKM(geoRA, geoDec, geoDistance, observer, ut)
	starDistance := math.Sqrt(lunarLimbDot(starVector, starVector))
	starRA := normalizeRA(math.Atan2(starVector[1], starVector[0]) / rad)
	starDec := math.Asin(clampUnit(starVector[2]/starDistance)) / rad
	if ephemeris.hasDistance {
		starRA, starDec = TopocentricRaDec(starRA, starDec, observer.Latitude, observer.Longitude, ut, starDistance, observer.Height)
		starRA = normalizeRA(starRA)
	}
	if !finite(topoRA) || !finite(topoDec) || !finite(topoDistance) || topoDistance <= 0 || !finite(starRA) || !finite(starDec) {
		return starLimbState{}, false
	}
	frame := ephemeris.base
	if frame.axisX, ok = lunarLimbNormalize(axisX); !ok {
		return starLimbState{}, false
	}
	if frame.axisZ, ok = lunarLimbNormalize(axisZ); !ok {
		return starLimbState{}, false
	}
	// 两轴分别插值后不再严格正交，按 z 轴重新正交化 x 轴。
	if frame.axisX, ok = lunarLimbNormalize(lunarLimbReject(frame.axisX, frame.axisZ)); !ok {
		return starLimbState{}, false
	}
	frame.axisY = lunarLimbCross(frame.axisZ, frame.axisX)
	radiusKM, ok := frame.radiusKM(lunarEclipseLimbVector(topoRA, topoDec, topoDistance), lunarEclipseLimbVector(starRA, starDec, 1), false)
	if !ok {
		return starLimbState{}, false
	}
	return starLimbState{
		separation:    angularSeparationDegrees(topoRA, topoDec, starRA, starDec) * 3600,
		limbRadius:    angularSemidiameterArcsec(radiusKM, topoDistance),
		meanRadius:    angularSemidiameterArcsec(moonEquatorialRadiusKM, topoDistance),
		positionAngle: occultationPositionAngle(topoRA, topoDec, starRA, starDec),
		distanceKM:    topoDistance,
	}, true
}

// starOccultationLimbRange 返回剖面半径范围；剖面不报告范围时在掩甚扫一圈并外扩 1 km。
func starOccultationLimbRange(tt float64, star StarCoordinate, observer Observer, limb LunarLimb, inst *Instance) (float64, float64, bool) {
	geoRA, geoDec := HMoonGeocentricApparentRaDecNWithInstance(tt, -1, inst)
	geoDistance := HMoonAwayNWithInstance(tt, -1, inst)
	frame := newLunarLimbFrame(limb, tt, lunarEclipseLimbVector(geoRA, geoDec, geoDistance), inst)
	if !frame.valid {
		return 0, 0, false
	}
	if frame.ranged {
		return frame.minKM, frame.maxKM, true
	}
	topoRA, topoDec := moonTopocentricApparentRaDec(tt, observer, -1, inst)
	starRA, starDec := starApparentRaDec(tt, star, observer, inst)
	view, ok := frame.viewToward(lunarEclipseLimbVector(topoRA, topoDec, 1), lunarEclipseLimbVector(starRA, starDec, 1))
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
	return low - starOccultationLimbUnrangedPadKM, high + starOccultationLimbUnrangedPadKM, true
}

// starOccultationSeparationCrossing 自掩甚向 direction 一侧求角距升到 radiusKM 对应角半径的时刻；角距自掩甚单调增大。
func starOccultationSeparationCrossing(greatestTT float64, direction float64, radiusKM float64, ephemeris *starLimbEphemeris, observer Observer) (float64, bool) {
	value := func(tt float64) float64 {
		state, ok := ephemeris.stateAt(tt, observer)
		if !ok {
			return math.NaN()
		}
		return state.separation - angularSemidiameterArcsec(radiusKM, state.distanceKM)
	}
	near, nearValue := greatestTT, value(greatestTT)
	if !finite(nearValue) {
		return 0, false
	}
	if nearValue >= 0 {
		return greatestTT, true
	}
	for step := 1; float64(step)*starOccultationLimbBoundStep <= starOccultationLimbBoundSpan; step++ {
		far := greatestTT + direction*float64(step)*starOccultationLimbBoundStep
		farValue := value(far)
		if !finite(farValue) {
			return 0, false
		}
		if farValue < 0 {
			near, nearValue = far, farValue
			continue
		}
		for i := 0; i < 64 && math.Abs(far-near) > starOccultationLimbRootDays; i++ {
			mid := (near + far) / 2
			if v := value(mid); v < 0 {
				near = mid
			} else {
				far = mid
			}
		}
		return (near + far) / 2, true
	}
	return 0, false
}

type starLimbRoot struct {
	tt        float64
	disappear bool
}

type starLimbScan struct {
	segments [][]StarOccultationLimbSample
	hidden   [][2]float64
	lowKM    float64
	highKM   float64
	stepDays float64
	// grazing 为真表示星迹始终在月缘起伏带内（单段采样）。
	grazing bool
}

// starOccultationLimbScanAt 按剖面起伏带把星迹分必见、必掩与带内三段，带外两端遮挡状态由构造保证；ok 为假表示剖面无值或星历失败。
func starOccultationLimbScanAt(greatestTT float64, star StarCoordinate, observer Observer, limb LunarLimb, stepDays float64, location *time.Location, inst *Instance) (starLimbScan, bool) {
	lowKM, highKM, ok := starOccultationLimbRange(greatestTT, star, observer, limb, inst)
	if !ok {
		return starLimbScan{}, false
	}
	scan := starLimbScan{lowKM: lowKM, highKM: highKM, stepDays: stepDays}
	ephemeris, ok := newStarLimbEphemeris(greatestTT, star, limb, inst)
	if !ok {
		return starLimbScan{}, false
	}
	outerLeft, okLeft := starOccultationSeparationCrossing(greatestTT, -1, highKM, ephemeris, observer)
	outerRight, okRight := starOccultationSeparationCrossing(greatestTT, 1, highKM, ephemeris, observer)
	if !okLeft || !okRight {
		return starLimbScan{}, false
	}
	if outerLeft == greatestTT && outerRight == greatestTT {
		return scan, true
	}
	innerLeft, okLeft := starOccultationSeparationCrossing(greatestTT, -1, lowKM, ephemeris, observer)
	innerRight, okRight := starOccultationSeparationCrossing(greatestTT, 1, lowKM, ephemeris, observer)
	if !okLeft || !okRight {
		return starLimbScan{}, false
	}
	type span struct {
		start, end           float64
		hiddenAtStart        bool
		hiddenAtEnd          bool
		forceStart, forceEnd bool
	}
	spans := []span{{outerLeft, innerLeft, false, true, true, true}, {innerRight, outerRight, true, false, true, true}}
	if innerLeft == greatestTT && innerRight == greatestTT {
		spans = []span{{outerLeft, outerRight, false, false, true, true}}
	}
	total, shortest := 0.0, math.Inf(1)
	for _, item := range spans {
		total += item.end - item.start
		shortest = math.Min(shortest, item.end-item.start)
	}
	// 步长过大时掠掩两端都可见、中间的遮挡会整段漏掉，所以每段至少保留 64 个样本。
	if limit := shortest / starOccultationLimbMinSamples; shortest > 0 && stepDays > limit {
		stepDays = math.Max(limit, starOccultationLimbMinimumStep.Hours()/24)
	}
	if total/stepDays > starOccultationLimbMaxSamples {
		stepDays = total / starOccultationLimbMaxSamples
	}
	scan.stepDays = stepDays
	scan.grazing = len(spans) == 1
	gapAt := func(tt float64) (starLimbState, float64, bool) {
		state, ok := ephemeris.stateAt(tt, observer)
		return state, state.separation - state.limbRadius, ok
	}
	var roots []starLimbRoot
	for _, item := range spans {
		// 剖面范围零宽时带也零长，遮挡状态在该点翻转；不产样本，避免同刻样本。
		if item.end-item.start < starOccultationLimbRootDays {
			if item.hiddenAtStart != item.hiddenAtEnd {
				roots = append(roots, starLimbRoot{tt: item.start, disappear: item.hiddenAtEnd})
			}
			continue
		}
		count := int(math.Ceil((item.end - item.start) / stepDays))
		if count < 1 {
			count = 1
		}
		samples := make([]StarOccultationLimbSample, 0, count+1)
		previousTT, previousHidden := 0.0, false
		var previousGap float64
		for index := 0; index <= count; index++ {
			tt := item.start + (item.end-item.start)*float64(index)/float64(count)
			state, gap, ok := gapAt(tt)
			if !ok {
				return starLimbScan{}, false
			}
			hidden := gap < 0
			// 带两端的遮挡状态由角距保证，消除端点上的舍入抖动。
			if index == 0 && item.forceStart {
				hidden = item.hiddenAtStart
			}
			if index == count && item.forceEnd {
				hidden = item.hiddenAtEnd
			}
			samples = append(samples, StarOccultationLimbSample{
				Time:             occultationTTToLocation(tt, location),
				SeparationArcsec: state.separation,
				LimbRadiusArcsec: state.limbRadius,
				MeanRadiusArcsec: state.meanRadius,
				PositionAngleDeg: state.positionAngle,
				Hidden:           hidden,
			})
			if index > 0 && hidden != previousHidden {
				root := starOccultationLimbBisect(previousTT, tt, previousGap, gap, gapAt)
				roots = append(roots, starLimbRoot{tt: root, disappear: hidden})
			}
			previousTT, previousHidden, previousGap = tt, hidden, gap
		}
		scan.segments = append(scan.segments, samples)
	}
	sort.SliceStable(roots, func(i, j int) bool { return roots[i].tt < roots[j].tt })
	start := math.NaN()
	for _, root := range roots {
		if root.disappear {
			start = root.tt
		} else if !math.IsNaN(start) {
			scan.hidden = append(scan.hidden, [2]float64{start, root.tt})
			start = math.NaN()
		}
	}
	return scan, true
}

// starOccultationLimbBisect 在相邻样本间求月缘间隙零点；端点被强制定号时零点可能落在端点上。
func starOccultationLimbBisect(left, right, leftGap, rightGap float64, gapAt func(float64) (starLimbState, float64, bool)) float64 {
	if leftGap*rightGap > 0 {
		if math.Abs(leftGap) < math.Abs(rightGap) {
			return left
		}
		return right
	}
	for i := 0; i < 64 && right-left > starOccultationLimbRootDays; i++ {
		mid := (left + right) / 2
		_, gap, ok := gapAt(mid)
		if !ok {
			break
		}
		if (gap < 0) == (leftGap < 0) {
			left, leftGap = mid, gap
		} else {
			right = mid
		}
	}
	return (left + right) / 2
}

// starOccultationInfoWithLimb 有月缘时按剖面判定是否成掩，掩始、掩终取首次遮挡与最后复现；剖面无值时退回圆月面。
func starOccultationInfoWithLimb(greatestTT float64, star StarCoordinate, observer Observer, safetyMarginArcsec float64, location *time.Location, limb LunarLimb, inst *Instance) (StarOccultationInfo, bool) {
	if limb == nil {
		return starOccultationInfoAtGreatest(greatestTT, star, observer, safetyMarginArcsec, location, inst)
	}
	info, ok := starOccultationGreatestInfo(greatestTT, star, observer, safetyMarginArcsec, location, inst)
	if !ok {
		return StarOccultationInfo{}, false
	}
	scan, ok := starOccultationLimbScanAt(greatestTT, star, observer, limb, starOccultationLimbDefaultStep.Hours()/24, location, inst)
	if !ok {
		return starOccultationInfoAtGreatest(greatestTT, star, observer, safetyMarginArcsec, location, inst)
	}
	return starOccultationInfoFromScan(info, scan, location), len(scan.hidden) > 0
}

func starOccultationInfoFromScan(info StarOccultationInfo, scan starLimbScan, location *time.Location) StarOccultationInfo {
	if len(scan.hidden) == 0 {
		info.Type = OccultationNone
		info.Immersion, info.Emersion = time.Time{}, time.Time{}
		info.ContactsComplete = false
		info.LunarLimbApplied = true
		return info
	}
	info.Type = OccultationTotal
	if scan.grazing {
		info.Type = OccultationGrazing
	}
	info.Immersion = occultationTTToLocation(scan.hidden[0][0], location)
	info.Emersion = occultationTTToLocation(scan.hidden[len(scan.hidden)-1][1], location)
	info.ContactsComplete = true
	info.LunarLimbApplied = true
	return info
}

// StarOccultationLimbProfileWithInstance 沿星迹按月缘剖面精密求解一次恒星月掩，返回全部遮挡区间与带内采样；info 只取掩甚时刻与观测者，须为 UTC 口径结果（不能传 UT1 标注副本），恒星未被遮挡时 Occultation.Type 为 OccultationNone / precise stellar occultation along the track with a lunar limb profile.
func StarOccultationLimbProfileWithInstance(info StarOccultationInfo, star StarCoordinate, limb LunarLimb, options StarOccultationLimbProfileOptions, inst *Instance) (StarOccultationLimbProfileResult, error) {
	if limb == nil {
		return StarOccultationLimbProfileResult{}, fmt.Errorf("%w: lunar limb profile is required", ErrInvalidOccultationInput)
	}
	if err := star.Validate(); err != nil {
		return StarOccultationLimbProfileResult{}, err
	}
	if err := info.Observer.Validate(); err != nil {
		return StarOccultationLimbProfileResult{}, err
	}
	if info.Greatest.IsZero() {
		return StarOccultationLimbProfileResult{}, fmt.Errorf("%w: greatest occultation time is required", ErrInvalidOccultationInput)
	}
	if options.Step < 0 {
		return StarOccultationLimbProfileResult{}, fmt.Errorf("%w: profile step cannot be negative", ErrInvalidOccultationInput)
	}
	step := options.Step
	if step == 0 {
		step = starOccultationLimbDefaultStep
	}
	if step < starOccultationLimbMinimumStep {
		step = starOccultationLimbMinimumStep
	}
	location := info.Greatest.Location()
	greatestTT := occultationTimeToTT(info.Greatest)
	base, ok := starOccultationGreatestInfo(greatestTT, star, info.Observer, math.Inf(1), location, inst)
	if !ok {
		return StarOccultationLimbProfileResult{}, fmt.Errorf("%w: geometry is unavailable at greatest occultation", ErrInvalidOccultationInput)
	}
	base.TargetID = star.ID
	scan, ok := starOccultationLimbScanAt(greatestTT, star, info.Observer, limb, step.Hours()/24, location, inst)
	if !ok {
		return StarOccultationLimbProfileResult{}, fmt.Errorf("%w: lunar limb profile has no value along the track", ErrInvalidOccultationInput)
	}
	profile := StarOccultationLimbProfileResult{
		Occultation:       starOccultationInfoFromScan(base, scan, location),
		Segments:          scan.segments,
		ReferenceRadiusKM: moonEquatorialRadiusKM,
		ReliefMinKM:       scan.lowKM,
		ReliefMaxKM:       scan.highKM,
		Step:              time.Duration(math.Round(scan.stepDays * 86400 * float64(time.Second))),
	}
	for _, interval := range scan.hidden {
		profile.Hidden = append(profile.Hidden, OccultationInterval{
			Start: occultationTTToLocation(interval[0], location),
			End:   occultationTTToLocation(interval[1], location),
		})
	}
	return profile, nil
}
