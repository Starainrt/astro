package basic

import (
	"math"
	"sort"
)

// RiseSetPhase 标识升落边界对应的局部事件阶段。
// RiseSetPhase identifies the local event phase represented by a horizon curve.
type RiseSetPhase string

// 升落边界上局部事件的三个阶段取值 / the three local event phase values on a rise/set boundary.
const (
	RiseSetPhaseStart    RiseSetPhase = "start"
	RiseSetPhaseGreatest RiseSetPhase = "greatest"
	RiseSetPhaseEnd      RiseSetPhase = "end"
)

// RiseSetDirection 标识边界上的目标天体正在升起还是落下。
// RiseSetDirection identifies whether the occulted body is rising or setting.
type RiseSetDirection string

// 边界上目标天体正在升起或落下 / whether the body is rising or setting along the boundary.
const (
	RiseSetDirectionRise RiseSetDirection = "rise"
	RiseSetDirectionSet  RiseSetDirection = "set"
)

type riseSetCyclicValueFunc func(float64) (float64, bool)

// 折点处升落残差与零相切而不变号，纯符号扫描会整圈找不到根；极区相位曲线正是在
// 这里断开。补根由调用方只在“分支内部空洞”上启用，因此不会改变分支端点语义。
// A rise/set fold makes the residual touch zero without changing sign, so a pure
// sign scan can return no root at all and a polar phase curve breaks apart there.
// The caller enables fold recovery only for interior branch holes, so branch
// endpoints keep their existing semantics.
const riseSetFoldRootResidualToleranceDeg = 5e-4

func riseSetCyclicRoots(samples int, valueAt riseSetCyclicValueFunc) []float64 {
	return riseSetSignChangeRoots(samples, valueAt)
}

// riseSetCyclicRootsWithFoldTolerance 在符号扫描为空时补出相切折点根。
// foldTolerance 非正时与历史符号扫描完全一致；正值为折点根的残差极小值上限。
// riseSetCyclicRootsWithFoldTolerance falls back to tangent fold roots when the
// sign scan stays empty. A non-positive foldTolerance reproduces the historical
// sign scan exactly; a positive value bounds the residual minimum accepted as a
// fold root.
func riseSetCyclicRootsWithFoldTolerance(samples int, foldTolerance float64, valueAt riseSetCyclicValueFunc) []float64 {
	roots := riseSetSignChangeRoots(samples, valueAt)
	if len(roots) > 0 || foldTolerance <= 0 {
		return roots
	}
	return riseSetFoldRoots(samples, foldTolerance, valueAt)
}

// riseSetFoldRoots 取相邻采样 |残差| 的严格极小值为候选，由黄金分割核对区间极小值
// 是否进入折点容差。
// riseSetFoldRoots takes adjacent |residual| samples forming a strict local minimum
// as candidates and lets the golden-section minimum decide whether the fold
// tolerance is met.
func riseSetFoldRoots(samples int, foldTolerance float64, valueAt riseSetCyclicValueFunc) []float64 {
	if samples < 12 {
		samples = 12
	}
	step := 2 * math.Pi / float64(samples)
	values := make([]float64, samples)
	valid := make([]bool, samples)
	for index := range values {
		values[index], valid[index] = valueAt(step * float64(index))
		valid[index] = valid[index] && finite(values[index])
	}
	roots := make([]float64, 0, 2)
	for index := range values {
		next := (index + 1) % samples
		previous := (index - 1 + samples) % samples
		following := (next + 1) % samples
		if !valid[index] || !valid[next] || !valid[previous] || !valid[following] {
			continue
		}
		leftValue, rightValue := math.Abs(values[index]), math.Abs(values[next])
		if math.Abs(values[previous]) <= leftValue || math.Abs(values[following]) <= rightValue {
			continue
		}
		if angle, ok := riseSetFoldRoot(step*float64(index), step*float64(next), foldTolerance, valueAt); ok {
			roots = append(roots, riseSetNormalizeRadians(angle))
		}
	}
	sort.Float64s(roots)
	return roots
}

func riseSetSignChangeRoots(samples int, valueAt riseSetCyclicValueFunc) []float64 {
	if samples < 12 {
		samples = 12
	}
	step := 2 * math.Pi / float64(samples)
	values := make([]float64, samples)
	valid := make([]bool, samples)
	for index := range values {
		values[index], valid[index] = valueAt(step * float64(index))
		valid[index] = valid[index] && finite(values[index])
	}

	roots := make([]float64, 0, 4)
	for index := range values {
		next := (index + 1) % samples
		if !valid[index] || !valid[next] {
			continue
		}
		left := step * float64(index)
		right := step * float64(index+1)
		leftValue, rightValue := values[index], values[next]
		if leftValue == 0 {
			roots = append(roots, riseSetNormalizeRadians(left))
			continue
		}
		if leftValue*rightValue > 0 {
			continue
		}
		for iteration := 0; iteration < 48 && right-left > 1e-11; iteration++ {
			middle := (left + right) / 2
			middleValue, ok := valueAt(riseSetNormalizeRadians(middle))
			if !ok || !finite(middleValue) {
				break
			}
			if leftValue*middleValue <= 0 {
				right, rightValue = middle, middleValue
			} else {
				left, leftValue = middle, middleValue
			}
		}
		roots = append(roots, riseSetNormalizeRadians((left+right)/2))
	}
	sort.Float64s(roots)
	unique := roots[:0]
	for _, root := range roots {
		if len(unique) == 0 || riseSetAngularDistance(root, unique[len(unique)-1]) > 1e-7 {
			unique = append(unique, root)
		}
	}
	if len(unique) > 1 && riseSetAngularDistance(unique[0], unique[len(unique)-1]) <= 1e-7 {
		unique = unique[:len(unique)-1]
	}
	return unique
}

// riseSetFoldRoot 用黄金分割在区间内最小化 |残差|，极小值进入容差时返回折点根。
// riseSetFoldRoot minimizes |residual| inside the interval by golden section and
// returns the fold root when the minimum stays inside the tolerance.
func riseSetFoldRoot(left, right, tolerance float64, valueAt riseSetCyclicValueFunc) (float64, bool) {
	const goldenRatio = 0.6180339887498949
	valueAtAbs := func(angle float64) (float64, bool) {
		value, ok := valueAt(riseSetNormalizeRadians(angle))
		if !ok || !finite(value) {
			return 0, false
		}
		return math.Abs(value), true
	}
	x1 := right - goldenRatio*(right-left)
	x2 := left + goldenRatio*(right-left)
	f1, ok1 := valueAtAbs(x1)
	f2, ok2 := valueAtAbs(x2)
	if !ok1 || !ok2 {
		return 0, false
	}
	for iteration := 0; iteration < 48 && right-left > 1e-9; iteration++ {
		if f1 > f2 {
			left, x1, f1 = x1, x2, f2
			x2 = left + goldenRatio*(right-left)
			if f2, ok2 = valueAtAbs(x2); !ok2 {
				return 0, false
			}
			continue
		}
		right, x2, f2 = x2, x1, f1
		x1 = right - goldenRatio*(right-left)
		if f1, ok1 = valueAtAbs(x1); !ok1 {
			return 0, false
		}
	}
	angle, minimum := (left+right)/2, math.Min(f1, f2)
	if minimum > tolerance {
		return 0, false
	}
	return angle, true
}

type riseSetGeographicResidualFunc func(longitude, latitude float64) (float64, float64, bool)

func riseSetRefineGeographicRoot(
	longitude, latitude float64,
	residualAt riseSetGeographicResidualFunc,
) (float64, float64, bool) {
	const finiteDifferenceDegrees = 1e-4
	for iteration := 0; iteration < 16; iteration++ {
		first, second, ok := residualAt(longitude, latitude)
		if !ok || !finite(first) || !finite(second) {
			return 0, 0, false
		}
		if math.Abs(first) <= 1e-11 && math.Abs(second) <= 1e-11 {
			return normalizeLongitude(longitude), latitude, true
		}
		firstLon, secondLon, lonOK := residualAt(longitude+finiteDifferenceDegrees, latitude)
		firstLat, secondLat, latOK := residualAt(longitude, latitude+finiteDifferenceDegrees)
		if !lonOK || !latOK {
			return 0, 0, false
		}
		a := (firstLon - first) / finiteDifferenceDegrees
		b := (firstLat - first) / finiteDifferenceDegrees
		c := (secondLon - second) / finiteDifferenceDegrees
		d := (secondLat - second) / finiteDifferenceDegrees
		determinant := a*d - b*c
		if !finite(determinant) || math.Abs(determinant) < 1e-18 {
			return 0, 0, false
		}
		deltaLongitude := (-first*d + b*second) / determinant
		deltaLatitude := (c*first - a*second) / determinant
		scale := math.Max(math.Abs(deltaLongitude), math.Abs(deltaLatitude))
		if scale > 5 {
			deltaLongitude *= 5 / scale
			deltaLatitude *= 5 / scale
		}
		longitude = normalizeLongitude(longitude + deltaLongitude)
		latitude += deltaLatitude
		if latitude <= -89.999999 || latitude >= 89.999999 || !finite(latitude) {
			return 0, 0, false
		}
	}
	first, second, ok := residualAt(longitude, latitude)
	return normalizeLongitude(longitude), latitude,
		ok && finite(first) && finite(second) && math.Abs(first) <= 1e-8 && math.Abs(second) <= 1e-8
}

func riseSetHorizonPoint(centerLongitude, centerLatitude, angle float64) (float64, float64) {
	longitude := centerLongitude * math.Pi / 180
	latitude := centerLatitude * math.Pi / 180
	center := [3]float64{
		math.Cos(latitude) * math.Cos(longitude),
		math.Cos(latitude) * math.Sin(longitude),
		math.Sin(latitude),
	}
	reference := [3]float64{0, 0, 1}
	if math.Abs(center[2]) > 0.9 {
		reference = [3]float64{1, 0, 0}
	}
	first := riseSetUnitVector(riseSetCross(reference, center))
	second := riseSetUnitVector(riseSetCross(center, first))
	point := [3]float64{
		first[0]*math.Cos(angle) + second[0]*math.Sin(angle),
		first[1]*math.Cos(angle) + second[1]*math.Sin(angle),
		first[2]*math.Cos(angle) + second[2]*math.Sin(angle),
	}
	return normalizeLongitude(math.Atan2(point[1], point[0]) * 180 / math.Pi),
		math.Asin(math.Max(-1, math.Min(1, point[2]))) * 180 / math.Pi
}

func riseSetCross(first, second [3]float64) [3]float64 {
	return [3]float64{
		first[1]*second[2] - first[2]*second[1],
		first[2]*second[0] - first[0]*second[2],
		first[0]*second[1] - first[1]*second[0],
	}
}

func riseSetUnitVector(value [3]float64) [3]float64 {
	norm := math.Sqrt(value[0]*value[0] + value[1]*value[1] + value[2]*value[2])
	return [3]float64{value[0] / norm, value[1] / norm, value[2] / norm}
}

func riseSetNormalizeRadians(value float64) float64 {
	value = math.Mod(value, 2*math.Pi)
	if value < 0 {
		value += 2 * math.Pi
	}
	return value
}

func riseSetAngularDistance(first, second float64) float64 {
	return math.Abs(math.Remainder(first-second, 2*math.Pi))
}
