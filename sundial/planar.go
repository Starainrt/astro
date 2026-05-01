package sundial

import (
	"math"
	"time"

	"github.com/starainrt/astro/sun"
)

// PlanarDial 平面日晷参数 / planar-sundial parameters.
//
// Latitude 为地理纬度（北正南负）；PlaneNormalAzimuth 为日晷盘面法线的方位角，
// 按正北为 0°、向东增加；PlaneNormalZenithDistance 为盘面法线的天顶距，0° 表示水平日晷，
// 90° 表示垂直日晷；StylusLength 为垂直于盘面的直晷针长度。
//
// 坐标系沿本章通用约定：x 轴位于盘面内且保持水平，y 轴沿盘面最大坡度向上，
// x 正向为右侧，y 正向为上坡方向。
type PlanarDial struct {
	Latitude                  float64
	PlaneNormalAzimuth        float64
	PlaneNormalZenithDistance float64
	StylusLength              float64
}

// PlanarShadowPoint 平面日晷影尖位置 / shadow-tip position on a planar sundial.
//
// X/Y 为影尖在盘面坐标系中的坐标；DenominatorQ 对应书中公式里的 Q；
// SunAboveHorizon 表示太阳在地平线上方；PlaneIlluminated 表示盘面被太阳照亮；
// Illuminated 为前两者同时满足。
type PlanarShadowPoint struct {
	X                float64
	Y                float64
	DenominatorQ     float64
	SunAboveHorizon  bool
	PlaneIlluminated bool
	Illuminated      bool
}

// PlanarGeometry 平面日晷几何量 / derived planar-sundial geometry.
//
// CenterX/CenterY 为日晷中心（极轴晷针固定点）坐标；PolarStylusLength 为极轴晷针长度；
// PolarStylusPlaneAngle 为极轴晷针与盘面的夹角。HasFiniteCenter 为 false 时，
// 表示极轴晷针与盘面平行，中心退化到无穷远处。
type PlanarGeometry struct {
	CenterX               float64
	CenterY               float64
	PolarStylusLength     float64
	PolarStylusPlaneAngle float64
	HasFiniteCenter       bool
}

// HourAngleInterval 时角区间 / hour-angle interval.
//
// Start/End 均为有符号太阳时角，单位度，满足 Start <= End。
// 约定取值范围为 [-180, 180]，用于表达一天中的一段连续时角。
type HourAngleInterval struct {
	Start float64
	End   float64
}

// DeclinationCurveSample 赤纬曲线采样点 / sampled point on a declination curve.
//
// HourAngle 为采样点的太阳时角；Point 为该时角下的影尖位置与照明状态。
type DeclinationCurveSample struct {
	HourAngle float64
	Point     PlanarShadowPoint
}

// DeclinationCurveSegment 赤纬曲线分段 / one illuminated segment of a declination curve.
//
// Interval 给出该段对应的连续受光时角范围；Samples 为该段内部的采样点。
type DeclinationCurveSegment struct {
	Declination float64
	Interval    HourAngleInterval
	Samples     []DeclinationCurveSample
}

// TimeLineSample 时间线采样点 / sampled point on a mean-time or zone-time line.
//
// Date 为该采样点对应的绝对时刻；其日期来自输入 date，钟面时间由调用参数指定。
// Declination 为该采样瞬时的太阳赤纬；HourAngle 为换算后的视太阳时角；
// Point 为该时角下的影尖位置与照明状态。
type TimeLineSample struct {
	Date        time.Time
	Declination float64
	HourAngle   float64
	Point       PlanarShadowPoint
}

// EquatorialNorthDial 北面赤道日晷 / north-face equatorial dial.
//
// 北半球时用于春夏半年（太阳赤纬为正）；南半球也可直接按公式使用。
func EquatorialNorthDial(latitude, stylusLength float64) PlanarDial {
	return PlanarDial{
		Latitude:                  latitude,
		PlaneNormalAzimuth:        0,
		PlaneNormalZenithDistance: 90 - latitude,
		StylusLength:              stylusLength,
	}
}

// EquatorialSouthDial 南面赤道日晷 / south-face equatorial dial.
//
// 北半球时用于秋冬半年（太阳赤纬为负）；南半球也可直接按公式使用。
func EquatorialSouthDial(latitude, stylusLength float64) PlanarDial {
	return PlanarDial{
		Latitude:                  latitude,
		PlaneNormalAzimuth:        180,
		PlaneNormalZenithDistance: 90 + latitude,
		StylusLength:              stylusLength,
	}
}

// HorizontalDial 水平日晷 / horizontal dial.
//
// 该构造器采用经典水平日晷的坐标约定：x 轴向东，y 轴向北。
func HorizontalDial(latitude, stylusLength float64) PlanarDial {
	return PlanarDial{
		Latitude:                  latitude,
		PlaneNormalAzimuth:        180,
		PlaneNormalZenithDistance: 0,
		StylusLength:              stylusLength,
	}
}

// VerticalDial 垂直日晷 / vertical dial.
//
// planeNormalAzimuth 为盘面法线方位角，按正北为 0°、向东增加。
// 例如：朝南墙面取 180°，朝东墙面取 90°。
func VerticalDial(latitude, planeNormalAzimuth, stylusLength float64) PlanarDial {
	return PlanarDial{
		Latitude:                  latitude,
		PlaneNormalAzimuth:        normalize360(planeNormalAzimuth),
		PlaneNormalZenithDistance: 90,
		StylusLength:              stylusLength,
	}
}

// Geometry 返回平面日晷的中心与极轴晷针几何量 / returns the derived planar geometry.
func (dial PlanarDial) Geometry() PlanarGeometry {
	geometry := PlanarGeometry{
		CenterX:               math.NaN(),
		CenterY:               math.NaN(),
		PolarStylusLength:     math.NaN(),
		PolarStylusPlaneAngle: math.NaN(),
	}
	if !dial.isFinite() {
		return geometry
	}

	P, _, _, _ := dial.baseCoefficients(0, 0)
	geometry.PolarStylusPlaneAngle = math.Asin(clampUnit(math.Abs(P))) * 180 / math.Pi
	if nearZero(P) {
		return geometry
	}

	latSin, latCos := sinCosDeg(dial.Latitude)
	zedSin, zedCos := sinCosDeg(dial.PlaneNormalZenithDistance)
	declinationSin, declinationCos := sinCosDeg(dial.bookPlaneNormalAzimuth())
	geometry.CenterX = dial.StylusLength / P * latCos * declinationSin
	geometry.CenterY = -dial.StylusLength / P * (latSin*zedSin + latCos*zedCos*declinationCos)
	geometry.PolarStylusLength = dial.StylusLength / math.Abs(P)
	geometry.HasFiniteCenter = true
	return geometry
}

// ShadowPointByHourAngleDeclination 影尖坐标（按时角与赤纬） / shadow point from hour angle and declination.
//
// hourAngle 为有符号太阳时角，上午为负，下午为正；declination 为太阳赤纬，单位度。
func (dial PlanarDial) ShadowPointByHourAngleDeclination(hourAngle, declination float64) PlanarShadowPoint {
	point := PlanarShadowPoint{
		X:            math.NaN(),
		Y:            math.NaN(),
		DenominatorQ: math.NaN(),
	}
	if !dial.isFinite() || !isFinite(hourAngle) || !isFinite(declination) {
		return point
	}

	_, Q, Nx, Ny := dial.baseCoefficients(hourAngle, declination)
	point.DenominatorQ = Q
	point.SunAboveHorizon = sunAboveHorizon(hourAngle, declination, dial.Latitude)
	point.PlaneIlluminated = Q > 0
	point.Illuminated = point.SunAboveHorizon && point.PlaneIlluminated
	if nearZero(Q) {
		return point
	}

	point.X = dial.StylusLength * Nx / Q
	point.Y = dial.StylusLength * Ny / Q
	return point
}

// ShadowPointAt 影尖坐标（按绝对时刻） / shadow point at an instant.
//
// 直接读取该时刻对应的视太阳时角和瞬时太阳赤纬，并返回平面日晷上的影尖位置。
func (dial PlanarDial) ShadowPointAt(date time.Time, lon float64) PlanarShadowPoint {
	return dial.ShadowPointByHourAngleDeclination(HourAngle(date, lon), sun.ApparentDec(date))
}

// MeanSolarTimePoint 平太阳时影尖位置 / shadow point for local mean solar time.
//
// date 应处于目标地点的地方平太阳时区，例如 `MeanSolarTime` 的返回值；其原有钟面时间会被忽略。
// meanSolarHours 为地方平太阳时钟面读数，单位小时，例如 9.5 表示 09:30。
func (dial PlanarDial) MeanSolarTimePoint(date time.Time, meanSolarHours float64) PlanarShadowPoint {
	sampleTime := dateWithClockHours(date, meanSolarHours)
	declination := sun.ApparentDec(sampleTime)
	return dial.ShadowPointByHourAngleDeclination(MeanSolarHourAngle(sampleTime, meanSolarHours), declination)
}

// ZoneTimePoint 区时影尖位置 / shadow point for zone time.
//
// date 提供民用日期和时区，原有钟面时间会被忽略；zoneTimeHours 为该时区下的区时钟面读数。
func (dial PlanarDial) ZoneTimePoint(date time.Time, lon, zoneTimeHours float64) PlanarShadowPoint {
	sampleTime := dateWithClockHours(date, zoneTimeHours)
	declination := sun.ApparentDec(sampleTime)
	return dial.ShadowPointByHourAngleDeclination(HourAngle(sampleTime, lon), declination)
}

// MeanSolarTimeLine 平太阳时时间线 / local mean solar time line.
//
// dates 由调用者自行决定取样日期密度，例如每月或每日；每个 date 都应处于目标地点的地方平太阳时区，
// 例如先通过 `MeanSolarTime` 得到对应地点的地方平太阳时再取其年月日。meanSolarHours 为地方平太阳时钟面读数。
func (dial PlanarDial) MeanSolarTimeLine(dates []time.Time, meanSolarHours float64) []TimeLineSample {
	if !isFinite(meanSolarHours) {
		return nil
	}
	samples := make([]TimeLineSample, 0, len(dates))
	for _, date := range dates {
		sampleTime := dateWithClockHours(date, meanSolarHours)
		declination := sun.ApparentDec(sampleTime)
		hourAngle := MeanSolarHourAngle(sampleTime, meanSolarHours)
		samples = append(samples, TimeLineSample{
			Date:        sampleTime,
			Declination: declination,
			HourAngle:   hourAngle,
			Point:       dial.ShadowPointByHourAngleDeclination(hourAngle, declination),
		})
	}
	return samples
}

// ZoneTimeLine 区时时间线 / zone-time line.
//
// dates 由调用者自行决定取样日期密度；zoneTimeHours 为 date 所在时区的区时钟面读数。
// 每个 date 的原有钟面时间都会被 zoneTimeHours 替换。
func (dial PlanarDial) ZoneTimeLine(dates []time.Time, lon, zoneTimeHours float64) []TimeLineSample {
	if !isFinite(zoneTimeHours) || !isFinite(lon) {
		return nil
	}
	samples := make([]TimeLineSample, 0, len(dates))
	for _, date := range dates {
		sampleTime := dateWithClockHours(date, zoneTimeHours)
		declination := sun.ApparentDec(sampleTime)
		hourAngle := HourAngle(sampleTime, lon)
		samples = append(samples, TimeLineSample{
			Date:        sampleTime,
			Declination: declination,
			HourAngle:   hourAngle,
			Point:       dial.ShadowPointByHourAngleDeclination(hourAngle, declination),
		})
	}
	return samples
}

// PlaneIlluminatedHourAngleIntervals 盘面受光时角区间 / plane-illuminated hour-angle intervals.
//
// declination 为太阳赤纬，单位度。返回的区间只考虑盘面受光，不判断太阳是否在地平线上方。
func (dial PlanarDial) PlaneIlluminatedHourAngleIntervals(declination float64) []HourAngleInterval {
	if !dial.isFinite() || !isFinite(declination) {
		return nil
	}
	sinCoeff, cosCoeff, constant := dial.qCoefficients(declination)
	return positiveHourAngleIntervals(sinCoeff, cosCoeff, constant)
}

// IlluminatedHourAngleIntervals 可见且受光时角区间 / illuminated hour-angle intervals.
//
// declination 为太阳赤纬，单位度。结果可直接用于日晷绘图时筛掉无效时线。
func (dial PlanarDial) IlluminatedHourAngleIntervals(declination float64) []HourAngleInterval {
	aboveHorizon := SunAboveHorizonHourAngleIntervals(dial.Latitude, declination)
	planeLit := dial.PlaneIlluminatedHourAngleIntervals(declination)
	return intersectHourAngleIntervals(aboveHorizon, planeLit)
}

// DeclinationCurve 赤纬曲线采样 / declination-curve samples.
//
// declination 为太阳赤纬，单位度；hourAngleStep 为采样步长，单位度，常用值是 15°（每小时一格）。
// 返回值按受光区间分段，每段都带有精确的时角范围；Samples 只包含区间内部的有效采样点。
func (dial PlanarDial) DeclinationCurve(declination, hourAngleStep float64) []DeclinationCurveSegment {
	if !dial.isFinite() || !isFinite(declination) || !isFinite(hourAngleStep) || hourAngleStep <= 0 {
		return nil
	}

	intervals := dial.IlluminatedHourAngleIntervals(declination)
	segments := make([]DeclinationCurveSegment, 0, len(intervals))
	for _, interval := range intervals {
		sampleAngles := intervalInteriorSampleAngles(interval, hourAngleStep)
		samples := make([]DeclinationCurveSample, 0, len(sampleAngles))
		for _, hourAngle := range sampleAngles {
			samples = append(samples, DeclinationCurveSample{
				HourAngle: hourAngle,
				Point:     dial.ShadowPointByHourAngleDeclination(hourAngle, declination),
			})
		}
		segments = append(segments, DeclinationCurveSegment{
			Declination: declination,
			Interval:    interval,
			Samples:     samples,
		})
	}
	return segments
}

// DeclinationCurveAt 瞬时赤纬曲线采样 / declination-curve samples at an instant.
//
// 用 date 对应瞬时太阳赤纬生成日晷分段曲线采样。
func (dial PlanarDial) DeclinationCurveAt(date time.Time, hourAngleStep float64) []DeclinationCurveSegment {
	return dial.DeclinationCurve(sun.ApparentDec(date), hourAngleStep)
}

// SunAboveHorizonHourAngleIntervals 地平线上方时角区间 / above-horizon hour-angle intervals.
//
// latitude 为地理纬度，declination 为太阳赤纬，单位度。结果只反映太阳是否升到地平线上方，
// 不包含盘面朝向的影响。
func SunAboveHorizonHourAngleIntervals(latitude, declination float64) []HourAngleInterval {
	if !isFinite(latitude) || !isFinite(declination) {
		return nil
	}

	latRad := latitude * math.Pi / 180
	declinationRad := declination * math.Pi / 180
	threshold := -math.Tan(latRad) * math.Tan(declinationRad)
	if threshold <= -1 {
		return fullDayHourAngleIntervals()
	}
	if threshold >= 1 {
		return nil
	}
	halfWidth := math.Acos(clampUnit(threshold)) * 180 / math.Pi
	if nearZero(halfWidth) {
		return nil
	}
	return []HourAngleInterval{{Start: -halfWidth, End: halfWidth}}
}

func (dial PlanarDial) baseCoefficients(hourAngle, declination float64) (P, Q, Nx, Ny float64) {
	latSin, latCos := sinCosDeg(dial.Latitude)
	zedSin, zedCos := sinCosDeg(dial.PlaneNormalZenithDistance)
	declinationSin, declinationCos := sinCosDeg(dial.bookPlaneNormalAzimuth())
	hourAngleSin, hourAngleCos := sinCosDeg(hourAngle)
	declinationTan := math.Tan(declination * math.Pi / 180)

	P = latSin*zedCos - latCos*zedSin*declinationCos
	Q = declinationSin*zedSin*hourAngleSin +
		(latCos*zedCos+latSin*zedSin*declinationCos)*hourAngleCos +
		P*declinationTan
	Nx = declinationCos*hourAngleSin -
		declinationSin*(latSin*hourAngleCos-latCos*declinationTan)
	Ny = zedCos*declinationSin*hourAngleSin -
		(latCos*zedSin-latSin*zedCos*declinationCos)*hourAngleCos -
		(latSin*zedSin+latCos*zedCos*declinationCos)*declinationTan
	return P, Q, Nx, Ny
}

func (dial PlanarDial) bookPlaneNormalAzimuth() float64 {
	return normalize360(dial.PlaneNormalAzimuth - 180)
}

func (dial PlanarDial) qCoefficients(declination float64) (sinCoeff, cosCoeff, constant float64) {
	latSin, latCos := sinCosDeg(dial.Latitude)
	zedSin, zedCos := sinCosDeg(dial.PlaneNormalZenithDistance)
	declinationSin, declinationCos := sinCosDeg(dial.bookPlaneNormalAzimuth())
	P := latSin*zedCos - latCos*zedSin*declinationCos
	return declinationSin * zedSin,
		latCos*zedCos + latSin*zedSin*declinationCos,
		P * math.Tan(declination*math.Pi/180)
}

func (dial PlanarDial) isFinite() bool {
	return isFinite(dial.Latitude) &&
		isFinite(dial.PlaneNormalAzimuth) &&
		isFinite(dial.PlaneNormalZenithDistance) &&
		isFinite(dial.StylusLength)
}

func sunAboveHorizon(hourAngle, declination, latitude float64) bool {
	latSin, latCos := sinCosDeg(latitude)
	decSin, decCos := sinCosDeg(declination)
	_, hourAngleCos := sinCosDeg(hourAngle)
	return latSin*decSin+latCos*decCos*hourAngleCos > 0
}

func sinCosDeg(value float64) (sinValue, cosValue float64) {
	rad := value * math.Pi / 180
	return math.Sin(rad), math.Cos(rad)
}

func normalize360(value float64) float64 {
	value = math.Mod(value, 360)
	if value < 0 {
		value += 360
	}
	return value
}

func clampUnit(value float64) float64 {
	if value > 1 {
		return 1
	}
	if value < -1 {
		return -1
	}
	return value
}

func nearZero(value float64) bool {
	return math.Abs(value) <= 1e-15
}

func positiveHourAngleIntervals(sinCoeff, cosCoeff, constant float64) []HourAngleInterval {
	radius := math.Hypot(sinCoeff, cosCoeff)
	if nearZero(radius) {
		if constant > 0 {
			return fullDayHourAngleIntervals()
		}
		return nil
	}

	threshold := -constant / radius
	if threshold <= -1 {
		return fullDayHourAngleIntervals()
	}
	if threshold >= 1 {
		return nil
	}

	center := math.Atan2(sinCoeff, cosCoeff) * 180 / math.Pi
	halfWidth := math.Acos(clampUnit(threshold)) * 180 / math.Pi
	if nearZero(halfWidth) {
		return nil
	}
	return splitWrappedSignedInterval(center-halfWidth, center+halfWidth)
}

func splitWrappedSignedInterval(start, end float64) []HourAngleInterval {
	if end-start >= 360-negligibleHourAngle {
		return fullDayHourAngleIntervals()
	}
	for start < -180 {
		start += 360
		end += 360
	}
	for start >= 180 {
		start -= 360
		end -= 360
	}
	if end <= 180 {
		return []HourAngleInterval{{Start: start, End: end}}
	}
	return []HourAngleInterval{
		{Start: -180, End: end - 360},
		{Start: start, End: 180},
	}
}

func intersectHourAngleIntervals(a, b []HourAngleInterval) []HourAngleInterval {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}
	intersections := make([]HourAngleInterval, 0, len(a)+len(b))
	for i, j := 0, 0; i < len(a) && j < len(b); {
		start := math.Max(a[i].Start, b[j].Start)
		end := math.Min(a[i].End, b[j].End)
		if end-start > negligibleHourAngle {
			intersections = append(intersections, HourAngleInterval{Start: start, End: end})
		}
		switch {
		case a[i].End < b[j].End-negligibleHourAngle:
			i++
		case b[j].End < a[i].End-negligibleHourAngle:
			j++
		default:
			i++
			j++
		}
	}
	return intersections
}

func intervalInteriorSampleAngles(interval HourAngleInterval, step float64) []float64 {
	if !isFinite(interval.Start) || !isFinite(interval.End) || !isFinite(step) || step <= 0 {
		return nil
	}
	if interval.End-interval.Start <= negligibleHourAngle {
		return nil
	}

	samples := make([]float64, 0, int(math.Ceil((interval.End-interval.Start)/step)))
	first := math.Ceil((interval.Start+negligibleHourAngle)/step) * step
	for hourAngle := first; hourAngle < interval.End-negligibleHourAngle; hourAngle += step {
		samples = append(samples, hourAngle)
	}
	if len(samples) == 0 {
		samples = append(samples, (interval.Start+interval.End)/2)
	}
	return samples
}

func fullDayHourAngleIntervals() []HourAngleInterval {
	return []HourAngleInterval{{Start: -180, End: 180}}
}

const negligibleHourAngle = 1e-12
