package basic

import "math"

// LunarLimb 月缘剖面，nil 为圆月面，内置数据表在 moon/limb / lunar limb profile; nil means a circular limb.
type LunarLimb interface {
	// LimbRadiusKM 按观测者总天平动经纬（度，东正北正）与自投影月北极起算的位置角（度）返回月心到月缘距离（km），无值时第二个返回 false / limb radius at a position angle for the given libration.
	LimbRadiusKM(librationLongitude, librationLatitude, positionAngle float64) (float64, bool)
}

// lunarLimbFrame 是某一瞬间的月球本体坐标架在真赤道坐标系下的表示，
// 由该时刻的地心天平动与自转轴位置角反解，可对任意站心方向复用。
type lunarLimbFrame struct {
	limb LunarLimb
	// stepDeg 是剖面位置角网格步长；ranged 为真时 minKM/maxKM 是剖面可能取到的半径范围。
	stepDeg      float64
	minKM, maxKM float64
	ranged       bool
	// axisX / axisY / axisZ 分别是月面经纬 (0°,0°)、(90°,0°) 与北极方向的单位矢量。
	axisX [3]float64
	axisY [3]float64
	axisZ [3]float64
	valid bool
}

// newLunarLimbFrame 用地心月球方向与该时刻的物理天平动构造本体坐标架。
// moonGeocentric 是真赤道坐标系下的地心月球矢量（km）。
func newLunarLimbFrame(limb LunarLimb, jdTT float64, moonGeocentric [3]float64, instances ...*Instance) lunarLimbFrame {
	if limb == nil {
		return lunarLimbFrame{}
	}
	inst := instanceValue(instances)

	toEarth, ok := lunarLimbNormalize([3]float64{-moonGeocentric[0], -moonGeocentric[1], -moonGeocentric[2]})
	if !ok {
		return lunarLimbFrame{}
	}
	physical := MoonPhysicalWithInstance(jdTT, inst)
	north, east, ok := lunarLimbSkyBasis([3]float64{-toEarth[0], -toEarth[1], -toEarth[2]})
	if !ok {
		return lunarLimbFrame{}
	}

	// 月球北极在天空平面上的投影自天北起向东量 P，于是极轴方向 = sinB·(月心→地心) + cosB·N。
	sinP, cosP := math.Sincos(physical.PositionAngle * rad)
	projected := lunarLimbCombine(cosP, north, sinP, east)
	sinB, cosB := math.Sincos(physical.LibrationLatitude * rad)
	axisZ, ok := lunarLimbNormalize(lunarLimbCombine(sinB, toEarth, cosB, projected))
	if !ok {
		return lunarLimbFrame{}
	}

	// 地心方向在赤道面内的分量对应本体系的 (cosL, sinL, 0)，据此旋回 x 轴。
	meridian, ok := lunarLimbNormalize(lunarLimbReject(toEarth, axisZ))
	if !ok {
		return lunarLimbFrame{}
	}
	sinL, cosL := math.Sincos(physical.LibrationLongitude * rad)
	axisX, ok := lunarLimbNormalize(lunarLimbCombine(cosL, meridian, -sinL, lunarLimbCross(axisZ, meridian)))
	if !ok {
		return lunarLimbFrame{}
	}
	frame := lunarLimbFrame{
		limb:    limb,
		stepDeg: lunarLimbDefaultStepDeg,
		axisX:   axisX,
		axisY:   lunarLimbCross(axisZ, axisX),
		axisZ:   axisZ,
		valid:   true,
	}
	if ranged, ok := limb.(lunarLimbRangeProvider); ok {
		low, high := ranged.RadiusRangeKM()
		if step := ranged.AngleStepDeg(); finite(step) && step >= lunarLimbMinStepDeg && step <= 1 {
			frame.stepDeg = step
		}
		if finite(low) && finite(high) && low > 0 && high >= low {
			frame.minKM, frame.maxKM, frame.ranged = low, high, true
		}
	}
	return frame
}

// radiusKM 返回站心所见月缘在指定方向上的月心距（km）；awayFromTarget 为真时取背向目标的一侧。
func (frame lunarLimbFrame) radiusKM(moonTopocentric, targetTopocentric [3]float64, awayFromTarget bool) (float64, bool) {
	if !frame.valid {
		return 0, false
	}
	moon, ok := lunarLimbNormalize(moonTopocentric)
	if !ok {
		return 0, false
	}
	target, ok := lunarLimbNormalize(targetTopocentric)
	if !ok {
		return 0, false
	}
	north, east, ok := lunarLimbSkyBasis(moon)
	if !ok {
		return 0, false
	}
	toObserver := [3]float64{-moon[0], -moon[1], -moon[2]}
	longitude := math.Atan2(lunarLimbDot(toObserver, frame.axisY), lunarLimbDot(toObserver, frame.axisX)) / rad
	latitude := math.Asin(clampUnit(lunarLimbDot(toObserver, frame.axisZ))) / rad

	// 自转轴位置角 P 与目标方向位置角 Q 都自天北向东量；剖面表的 θ 自月球北极向月面东经一侧
	// （从地球看是天西）量，所以 θ = P − Q。
	axisAngle := math.Atan2(lunarLimbDot(frame.axisZ, east), lunarLimbDot(frame.axisZ, north))
	targetPlane, ok := lunarLimbNormalize(lunarLimbReject(target, moon))
	if !ok {
		return 0, false
	}
	targetAngle := math.Atan2(lunarLimbDot(targetPlane, east), lunarLimbDot(targetPlane, north))
	positionAngle := (axisAngle - targetAngle) / rad
	if awayFromTarget {
		positionAngle += 180
	}
	positionAngle = math.Mod(positionAngle, 360)
	if positionAngle < 0 {
		positionAngle += 360
	}
	radius, ok := frame.limb.LimbRadiusKM(longitude, latitude, positionAngle)
	if !ok || !finite(radius) || radius <= 0 {
		return 0, false
	}
	return radius, true
}

// lunarLimbView 是站心所见的月面经纬与指向目标方向的剖面位置角（度）。
type lunarLimbView struct {
	longitude, latitude, angle float64
}

// viewToward 求站心所见天平动与目标方向的剖面位置角，约定与 radiusKM 相同。
func (frame lunarLimbFrame) viewToward(moonTopocentric, targetTopocentric [3]float64) (lunarLimbView, bool) {
	if !frame.valid {
		return lunarLimbView{}, false
	}
	moon, ok := lunarLimbNormalize(moonTopocentric)
	if !ok {
		return lunarLimbView{}, false
	}
	target, ok := lunarLimbNormalize(targetTopocentric)
	if !ok {
		return lunarLimbView{}, false
	}
	north, east, ok := lunarLimbSkyBasis(moon)
	if !ok {
		return lunarLimbView{}, false
	}
	toObserver := [3]float64{-moon[0], -moon[1], -moon[2]}
	axisAngle := math.Atan2(lunarLimbDot(frame.axisZ, east), lunarLimbDot(frame.axisZ, north))
	targetPlane, ok := lunarLimbNormalize(lunarLimbReject(target, moon))
	if !ok {
		return lunarLimbView{}, false
	}
	targetAngle := math.Atan2(lunarLimbDot(targetPlane, east), lunarLimbDot(targetPlane, north))
	return lunarLimbView{
		longitude: math.Atan2(lunarLimbDot(toObserver, frame.axisY), lunarLimbDot(toObserver, frame.axisX)) / rad,
		latitude:  math.Asin(clampUnit(lunarLimbDot(toObserver, frame.axisZ))) / rad,
		angle:     (axisAngle - targetAngle) / rad,
	}, true
}

// radiusAt 返回剖面位置角偏离目标方向 offset 度处的月缘月心距（km）。
func (frame lunarLimbFrame) radiusAt(view lunarLimbView, offset float64) (float64, bool) {
	angle := math.Mod(view.angle+offset, 360)
	if angle < 0 {
		angle += 360
	}
	radius, ok := frame.limb.LimbRadiusKM(view.longitude, view.latitude, angle)
	if !ok || !finite(radius) || radius <= 0 {
		return 0, false
	}
	return radius, true
}

// lunarLimbSkyBasis 返回视方向 direction（观测者指向天体）处的天球北、东单位矢量；
// 东向为赤经增加方向，e = n × direction。
func lunarLimbSkyBasis(direction [3]float64) ([3]float64, [3]float64, bool) {
	north, ok := lunarLimbNormalize(lunarLimbReject([3]float64{0, 0, 1}, direction))
	if !ok {
		return [3]float64{}, [3]float64{}, false
	}
	return north, lunarLimbCross(north, direction), true
}

func lunarLimbNormalize(vector [3]float64) ([3]float64, bool) {
	norm := math.Sqrt(vector[0]*vector[0] + vector[1]*vector[1] + vector[2]*vector[2])
	if !finite(norm) || norm <= 0 {
		return [3]float64{}, false
	}
	return [3]float64{vector[0] / norm, vector[1] / norm, vector[2] / norm}, true
}

// lunarLimbReject 返回 vector 中垂直于单位矢量 axis 的分量。
func lunarLimbReject(vector, axis [3]float64) [3]float64 {
	projection := lunarLimbDot(vector, axis)
	return [3]float64{
		vector[0] - projection*axis[0],
		vector[1] - projection*axis[1],
		vector[2] - projection*axis[2],
	}
}

func lunarLimbCombine(scaleA float64, a [3]float64, scaleB float64, b [3]float64) [3]float64 {
	return [3]float64{
		scaleA*a[0] + scaleB*b[0],
		scaleA*a[1] + scaleB*b[1],
		scaleA*a[2] + scaleB*b[2],
	}
}

func lunarLimbCross(a, b [3]float64) [3]float64 {
	return [3]float64{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}

func lunarLimbDot(a, b [3]float64) float64 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

// lunarLimbDefaultStepDeg 剖面不报告网格步长时的位置角采样步长（度）。
const lunarLimbDefaultStepDeg = 0.1

// lunarLimbMinStepDeg 剖面报告的步长低于此值时退回默认步长，防止全圈扫描次数失控。
const lunarLimbMinStepDeg = 0.001

// lunarLimbRangeProvider 可选接口：报告剖面半径范围与位置角网格步长，用于门控与剪枝；步长不在 [0.001°, 1°] 内时改用默认 0.1°。
type lunarLimbRangeProvider interface {
	RadiusRangeKM() (float64, float64)
	AngleStepDeg() float64
}

// lunarLimbContact 是某一站心时刻按全部位置角求日面接触余量的上下文，角量均为弧度，在天空平面上按平直几何计。
type lunarLimbContact struct {
	frame        *lunarLimbFrame
	view         lunarLimbView
	distanceKM   float64
	separation   float64
	targetRadius float64
	// anglePerKM 非零时月缘半径按线性比例换算成角量（月食影面坐标）；为零时按 asin(r/distanceKM)。
	anglePerKM float64
}

// radius 返回该方向月缘角半径；exact 为假时用 r/距离 的线性近似，只供扫描选方向，极值处再按 exact 重算。
func (contact lunarLimbContact) radius(offsetDeg float64, exact bool) (float64, bool) {
	radiusKM, ok := contact.frame.radiusAt(contact.view, offsetDeg)
	if !ok {
		return 0, false
	}
	if contact.anglePerKM != 0 {
		return radiusKM * contact.anglePerKM, true
	}
	if !exact {
		return radiusKM / contact.distanceKM, true
	}
	return math.Asin(localSolarEclipseClampUnit(radiusKM / contact.distanceKM)), true
}

// bounds 返回剖面半径的角量上下界；剖面不报告范围时 ok 为假，此时不门控、不剪枝。
func (contact lunarLimbContact) bounds() (float64, float64, bool) {
	if !contact.frame.ranged {
		return 0, 0, false
	}
	if contact.anglePerKM != 0 {
		return contact.frame.minKM * contact.anglePerKM, contact.frame.maxKM * contact.anglePerKM, true
	}
	return contact.frame.minKM / contact.distanceKM, contact.frame.maxKM / contact.distanceKM, true
}

// rayChord 返回自月心沿偏离日心方向 δ 的射线与日面边界的近、远交点距离；射线不交日面时 ok 为假。
func (contact lunarLimbContact) rayChord(offsetDeg float64) (float64, float64, bool) {
	sine, cosine := math.Sincos(offsetDeg * rad)
	d, s := contact.separation, contact.targetRadius
	discriminant := s*s - d*d*sine*sine
	if discriminant < 0 {
		return 0, 0, false
	}
	root := math.Sqrt(discriminant)
	return d*cosine - root, d*cosine + root, true
}

// scan 自 baseDeg 向两侧按剖面位置角网格节点求余量极值并在极值方向按精确半径重算；maximize 为真取最大。剖面在节点间按位置角线性插值，
// 只取节点与 baseDeg 即不会漏掉剖面峰谷。bound(δ) 给出该方向余量可能达到的最好值，越过它不可能再刷新极值即停止该侧。
func (contact lunarLimbContact) scan(
	baseDeg float64,
	maximize bool,
	margin func(offsetDeg float64, exact bool) (float64, bool, bool),
	bound func(offsetDeg float64) (float64, bool),
) (float64, bool) {
	best, bestOffset, found := 0.0, baseDeg, false
	visit := func(offset float64) bool {
		value, hit, ok := margin(offset, false)
		if !ok {
			return false
		}
		if hit && (!found || (maximize && value > best) || (!maximize && value < best)) {
			best, bestOffset, found = value, offset, true
		}
		return true
	}
	if !visit(baseDeg) {
		return 0, false
	}
	step := contact.frame.stepDeg
	absolute := contact.view.angle + baseDeg
	first := [2]float64{math.Floor(absolute/step)*step - absolute, math.Ceil(absolute/step)*step - absolute}
	steps := int(math.Ceil(180/step)) + 1
	for side, sign := range [2]float64{-1, 1} {
		for index := 0; index <= steps; index++ {
			delta := first[side] + sign*float64(index)*step
			if delta == 0 || math.Abs(delta) > 180 {
				continue
			}
			offset := baseDeg + delta
			if limit, ok := bound(offset); ok && found &&
				((maximize && limit <= best) || (!maximize && limit >= best)) {
				break
			}
			if !visit(offset) {
				return 0, false
			}
		}
	}
	if !found {
		return 0, false
	}
	exact, _, ok := margin(bestOffset, true)
	return exact, ok
}

// outerRadius 返回外切等效月半径（弧度）：取全部方向中月缘超出日面近侧边界最多者，使 d − R − s 为零即外切。
// 只有剖面范围保证间隙离零至少一个起伏幅度时才跳过扫描，免得门控两侧的间隙在接触根附近不连续。
func (contact lunarLimbContact) outerRadius() (float64, bool) {
	d, s := contact.separation, contact.targetRadius
	low, high, ranged := contact.bounds()
	if d <= s || (ranged && (high <= d-s-(high-low) || low >= d-s+(high-low))) {
		return contact.radius(0, true)
	}
	best, ok := contact.scan(0, true,
		func(offset float64, exact bool) (float64, bool, bool) {
			near, _, hit := contact.rayChord(offset)
			if !hit {
				return 0, false, true
			}
			radius, ok := contact.radius(offset, exact)
			return radius - near, true, ok
		},
		func(offset float64) (float64, bool) {
			near, _, hit := contact.rayChord(offset)
			if !hit {
				return math.Inf(-1), true
			}
			return high - near, ranged
		})
	if !ok {
		return 0, false
	}
	return d - s + best, true
}

// innerRadius 返回内切等效月半径（弧度），使内切间隙 d − |R − s| 等于 min(−余量, d)。total 为真按全食求月缘盖住日面
// 远侧边界的最小余量，否则按环食求日面远侧边界超出月缘的最小余量；门控条件同 outerRadius。
func (contact lunarLimbContact) innerRadius(total bool) (float64, bool) {
	d, s := contact.separation, contact.targetRadius
	low, high, ranged := contact.bounds()
	relief := high - low
	if total {
		if ranged && (low >= d+s+relief || high <= d+s-relief) {
			radius, ok := contact.radius(0, true)
			return s + math.Max(radius-s, 0), ok
		}
		best, ok := contact.scan(0, false,
			func(offset float64, exact bool) (float64, bool, bool) {
				_, far, hit := contact.rayChord(offset)
				if !hit {
					return 0, false, true
				}
				radius, ok := contact.radius(offset, exact)
				return radius - far, true, ok
			},
			func(offset float64) (float64, bool) {
				_, far, hit := contact.rayChord(offset)
				if !hit {
					return math.Inf(1), true
				}
				return low - far, ranged
			})
		if !ok {
			return 0, false
		}
		return s + math.Max(d+best, 0), true
	}
	// 月心在日面外时环食不可能，内切间隙恰为 d。
	if d >= s {
		return s, true
	}
	if ranged && (high <= s-d-relief || low >= s-d+relief) {
		radius, ok := contact.radius(180, true)
		return math.Min(radius, s), ok
	}
	best, ok := contact.scan(180, false,
		func(offset float64, exact bool) (float64, bool, bool) {
			_, far, hit := contact.rayChord(offset)
			if !hit {
				return 0, false, true
			}
			radius, ok := contact.radius(offset, exact)
			return far - radius, true, ok
		},
		func(offset float64) (float64, bool) {
			_, far, hit := contact.rayChord(offset)
			if !hit {
				return math.Inf(1), true
			}
			return far - high, ranged
		})
	if !ok {
		return 0, false
	}
	return s - math.Max(d+best, 0), true
}
