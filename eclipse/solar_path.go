package eclipse

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

// SolarEclipsePathOptions 控制日食中心路径采样。
// SolarEclipsePathOptions controls central solar eclipse path sampling.
type SolarEclipsePathOptions struct {
	// Step 是基础时间采样步长；<=0 时使用 1 分钟。
	// Step is the base time step; values <= 0 use one minute.
	Step time.Duration
	// TargetSpacingKM 是相邻中心线点的最大目标地表距离；<=0 时不按距离加密。
	// TargetSpacingKM is the target maximum ground spacing between centerline points; values <= 0 disable spacing refinement.
	TargetSpacingKM float64
	// SkipCentralBand 表示调用方已经持有同一场日食的完整足迹结果（其中包含
	// CentralBandSegments），本次不再重复重建中心食带。
	// SkipCentralBand reports that the caller already holds the full-footprint
	// result for the same eclipse, so the central-band envelope is not rebuilt.
	SkipCentralBand bool
}

// SolarEclipsePathPoint 表示日食路径上的一个地理点。
// SolarEclipsePathPoint is one geographic point on a solar eclipse path.
type SolarEclipsePathPoint struct {
	// Time 时刻，保持用户输入时区, time in the input location.
	Time time.Time
	// Longitude 经度，东正西负, longitude in degrees, east positive.
	Longitude float64
	// Latitude 纬度，北正南负, latitude in degrees, north positive.
	Latitude float64
	// SunAltitude 太阳高度角，单位度, Sun altitude in degrees.
	SunAltitude float64
	// WidthKM 中心食带宽度，单位千米；仅中心线点有意义。
	// WidthKM is the central path width in kilometers; meaningful for centerline points.
	WidthKM float64
}

// SolarEclipsePath 表示一次中心日食的路径数据。
// SolarEclipsePath contains central solar eclipse path data.
type SolarEclipsePath struct {
	// Eclipse 是对应的全局日食信息, related global solar eclipse information.
	Eclipse SolarEclipseInfo
	// PathWidthDefined 表示本结果的带宽是否有定义：两限存在且上下两条限界线都非空时才为 true；
	// 为 false 时 Eclipse.PathWidthKM 与 Greatest.WidthKM 都是 0。
	// PathWidthDefined reports whether this result has a defined band width: both
	// limits must exist and both limit lines must be non-empty. When it is false,
	// Eclipse.PathWidthKM and Greatest.WidthKM are 0.
	PathWidthDefined bool
	// Greatest 是食甚点/最佳观测点, greatest eclipse point.
	Greatest SolarEclipsePathPoint
	// CenterLine 是中心线, central line.
	CenterLine []SolarEclipsePathPoint
	// NorthernLimit 是中心食带北界近似线, approximate northern limit of the central path.
	NorthernLimit []SolarEclipsePathPoint
	// SouthernLimit 是中心食带南界近似线, approximate southern limit of the central path.
	SouthernLimit []SolarEclipsePathPoint
	// MaxCentralDuration 是中心线上最长的中心食时长，0 表示没有中心食样本可用。
	// 与目录口径的区别：目录取食甚点，这里取整条中心线的最大值。
	// MaxCentralDuration is the longest central phase on the center line, and 0
	// when no usable sample exists. Catalogues publish the value at greatest
	// eclipse; this is the maximum along the track.
	MaxCentralDuration time.Duration
	// MaxCentralDurationLongitude 与 MaxCentralDurationLatitude 给出该最长时长的位置。
	// MaxCentralDurationLongitude and MaxCentralDurationLatitude locate it.
	MaxCentralDurationLongitude float64
	MaxCentralDurationLatitude  float64
	// CentralBandSegments is the authoritative continuous central-band envelope.
	CentralBandSegments [][]SolarEclipsePathPoint
	// CentralBandSampled reports that the envelope was reconstructed from sampled
	// instantaneous footprints rather than an analytic envelope.
	CentralBandSampled bool
	// Step 是实际采用的基础时间采样步长, effective base time step.
	Step time.Duration
	// TargetSpacingKM 是实际采用的目标空间采样距离，单位千米。
	// TargetSpacingKM is the effective target spacing in kilometers.
	TargetSpacingKM float64
}

// SolarEclipsePartialFootprintOptions 控制日食偏食半影足迹采样。
// SolarEclipsePartialFootprintOptions controls solar eclipse penumbral footprint sampling.
type SolarEclipsePartialFootprintOptions struct {
	// Step 是基础时间采样步长；<=0 时使用 5 分钟。
	// Step is the base time step; values <= 0 use five minutes.
	Step time.Duration
	// BoundaryPoints 是每个瞬时半影边界的角向采样点数；<=0 时使用 180。
	// BoundaryPoints is the angular sample count for each instantaneous penumbral boundary; values <= 0 use 180.
	BoundaryPoints int
	// CentralShadowStep 是本影/反本影瞬时足迹的采样步长；<=0 时不计算。
	// CentralShadowStep is the umbral/antumbral footprint step; values <= 0 disable it.
	CentralShadowStep time.Duration
	// RiseSetStep 独立于 Step 控制日升日落边界的采样步长；零值使用 2 分钟。
	// RiseSetStep controls sunrise/sunset boundary sampling independently from Step; zero uses two minutes.
	RiseSetStep time.Duration
	// DisableRiseSet 跳过六类日升日落阶段边界曲线。
	// DisableRiseSet skips the six sunrise/sunset boundary curves.
	DisableRiseSet bool
	// MagnitudeValues 是要计算的地方最大食分等值线；空值不计算，线条使用独立的自适应空间采样。
	// MagnitudeValues requests local maximum-magnitude contours; empty disables them, and contours use independent adaptive spatial sampling.
	MagnitudeValues []float64
	// GreatestTimeValues 是要计算的地方食甚时刻等值线取值，按绝对时刻使用，Location 不参与换算；
	// 空值时改用 GreatestTimeStep。最多 64 条，超出按时间截断。
	// GreatestTimeValues requests local greatest-eclipse time isolines as absolute instants; their
	// Location does not affect the computation. When empty, GreatestTimeStep is used instead. At
	// most 64 are kept, truncated in time order.
	GreatestTimeValues []time.Time
	// GreatestTimeStep 是等时线间隔；仅在 GreatestTimeValues 为空时生效，非正值不绘制。
	// 取值对齐到 UTC 整刻度（显示层要按展示时区对齐时请自行生成时刻并传给 GreatestTimeValues）。
	// GreatestTimeStep is the isochrone interval; it applies only when GreatestTimeValues is empty,
	// and non-positive values disable the isolines. Levels align to UTC ticks; display layers that
	// need the viewing timezone grid should generate the instants themselves.
	GreatestTimeStep time.Duration
}

// SolarEclipsePartialAreaOptions 是 SolarEclipsePartialFootprintOptions 的兼容别名。
// SolarEclipsePartialAreaOptions is a compatibility alias for SolarEclipsePartialFootprintOptions.
type SolarEclipsePartialAreaOptions = SolarEclipsePartialFootprintOptions

// SolarEclipsePartialFootprint 表示某一时刻的半影足迹边界。
// SolarEclipsePartialFootprint is the penumbral footprint boundary at one instant.
type SolarEclipsePartialFootprint struct {
	// Time 时刻，保持用户输入时区, time in the input location.
	Time time.Time
	// Boundaries 是半影边界分段；反经线或无效投影会拆成多段。
	// Boundaries are segmented penumbral boundary polylines, split at invalid projections or the antimeridian.
	Boundaries [][]SolarEclipsePathPoint
	// Closed 表示 Boundaries 是否构成一个闭合边界。
	// Closed indicates whether Boundaries form one closed boundary.
	Closed bool
	// HorizonEnds 是未闭合边界两端延伸到地平圈的擦地点，顺序与 Boundaries 的走向一致；
	// 边界自身闭合时为空。被地平线切断的瞬时阴影区域由「物理边界 + 两个擦地点之间的
	// 地平弧」闭合，擦地点的太阳高度为 0，因此闭合弧是精确结果而非启发式。
	// HorizonEnds are the two limb-grazing points where an open boundary reaches the
	// horizon, ordered like Boundaries; empty when the boundary closes on itself. A
	// region cut by the horizon is closed by the physical boundary plus the horizon
	// arc between these grazing points, whose solar altitude is zero, so the closing
	// arc is exact rather than a heuristic.
	HorizonEnds []SolarEclipsePathPoint
}

// SolarEclipsePartialFootprintsInfo 表示一次日食的偏食半影足迹序列。
// SolarEclipsePartialFootprintsInfo contains penumbral footprint samples for a solar eclipse.
type SolarEclipsePartialFootprintsInfo struct {
	// Eclipse 是对应的全局日食信息, related global solar eclipse information.
	Eclipse SolarEclipseInfo
	// Footprints 是按时间采样的瞬时半影足迹, sampled instantaneous penumbral footprints.
	Footprints []SolarEclipsePartialFootprint
	// CentralShadowFootprints 是按时间采样的本影/反本影足迹。
	// CentralShadowFootprints are sampled umbral/antumbral footprints.
	CentralShadowFootprints []SolarEclipsePartialFootprint
	// CentralBandFootprints 是始终计算的低成本本影/反本影端部样本，用于闭合中心食带。
	// CentralBandFootprints are always-computed lightweight umbral/antumbral end samples used to close the central band.
	CentralBandFootprints []SolarEclipsePartialFootprint
	// CentralBandSegments 是地方中心食条件与日出/日落食甚边界组成的连续闭合包络。
	// CentralBandSegments are continuous closed envelopes bounded by local centrality and greatest-at-horizon conditions.
	CentralBandSegments [][]SolarEclipsePathPoint
	// CentralBandSampled 表示上面的包络是用采样瞬时足迹重建的（解析包络不适用）。
	// CentralBandSampled reports that the envelope was reconstructed from the
	// sampled instantaneous footprints because no analytic envelope applied.
	CentralBandSampled bool
	// CentralBandHorizonClosures 是两限界掠地事件中分别连接首尾两侧限界的食甚地平线弧。
	// CentralBandHorizonClosures are the greatest-at-horizon arcs joining both ends of a grazing two-limit event.
	CentralBandHorizonClosures [][]SolarEclipsePathPoint
	// PartialBandContours 是地方最大食分等于零的连续可见包络，用于闭合偏食可见域。
	// PartialBandContours are the continuous zero local-maximum-magnitude envelopes used to close the partial-eclipse visibility region.
	PartialBandContours [][]SolarEclipsePathPoint
	// MagnitudeContours 是按食分值采样的两侧等值线。
	// MagnitudeContours are sampled two-sided local maximum-magnitude contours.
	MagnitudeContours []SolarEclipseMagnitudeContour
	// GreatestTimeContours 是按食甚时刻采样的等时线。
	// GreatestTimeContours are sampled local greatest-eclipse time isolines.
	GreatestTimeContours []SolarEclipseGreatestTimeContour
	// RiseSetCurves 是初亏、食甚和复圆分别发生在日出或日落时的六类边界。
	// RiseSetCurves are the six boundaries where local start, greatest, or end occurs at sunrise or sunset.
	RiseSetCurves []SolarEclipseRiseSetCurve
	// P1-P4 是半影与地球的外切/内切接触点；不存在的内切点保持零值。
	// P1-P4 are external/internal penumbral contacts; absent internal contacts remain zero.
	P1 SolarEclipsePathPoint
	P2 SolarEclipsePathPoint
	P3 SolarEclipsePathPoint
	P4 SolarEclipsePathPoint
	// U1-U4 是本影/反本影与地球的外切/内切接触点；不存在时保持零值。
	// U1-U4 are external/internal umbral/antumbral contacts; absent contacts remain zero.
	U1 SolarEclipsePathPoint
	U2 SolarEclipsePathPoint
	U3 SolarEclipsePathPoint
	U4 SolarEclipsePathPoint
	// Step 是实际采用的基础时间采样步长, effective base time step.
	Step time.Duration
	// BoundaryPoints 是实际采用的边界角向采样点数。
	// BoundaryPoints is the effective angular sample count for each boundary.
	BoundaryPoints int
	// CentralShadowStep 是本影/反本影足迹的实际采样步长；0 表示未计算。
	// CentralShadowStep is the effective umbral/antumbral footprint step; zero means disabled.
	CentralShadowStep time.Duration
	// CentralBandStep 是中心食带足迹的最细实际采样步长。
	// CentralBandStep is the finest effective sampling step for central-band footprints.
	CentralBandStep time.Duration
}

// SolarEclipseGreatestTimeContour 是一个固定地方食甚时刻的等值线支路集合。
// SolarEclipseGreatestTimeContour contains the continuous branches of one fixed local greatest-eclipse time.
type SolarEclipseGreatestTimeContour struct {
	// JDE 是该等值线对应的力学时儒略日，也就是支路上地方食甚发生的时刻。
	// JDE is the TT Julian ephemeris day represented by this contour, the local greatest-eclipse instant along every branch.
	JDE float64
	// Time 是该等值线表示的地方食甚时刻，保持用户输入时区；按步长请求时它是原始对齐时刻，
	// 避免 JDE 往返把整分取值截断成前一分钟。
	// Time is the local greatest-eclipse instant represented by this contour, in the input timezone;
	// for step-derived levels it is the original aligned instant, so a JDE round trip cannot truncate
	// a whole-minute level into the previous minute.
	Time time.Time
	// Segments 是该时刻的连续等时线支路；一条支路两端止于地平线（几何地平，无蒙气差修正）或偏食可见域边界，
	// 纬度 ±88° 以上不再延拓，同一时刻可能有多条不相连的支路。
	// Segments are continuous isochrone branches; each branch ends at the horizon or the partial-visibility boundary.
	Segments [][]SolarEclipsePathPoint
}

// SolarEclipseMagnitudeContour 是一条地方最大食分等值线的连续支路集合。
// SolarEclipseMagnitudeContour contains the continuous branches of one local maximum-magnitude contour.
type SolarEclipseMagnitudeContour struct {
	// Magnitude 是该等值线表示的地方最大食分。
	// Magnitude is the local maximum eclipse magnitude represented by this contour.
	Magnitude float64
	// Segments 是等食分线的连续支路；临近地平线时局部食甚时刻可以沿空间支路折返。
	// Segments are continuous contour branches; local greatest times may fold along a spatial branch near the horizon.
	Segments [][]SolarEclipsePathPoint
	// NorthernLimit 和 SouthernLimit 保留两侧中心食等值线的兼容视图。
	// NorthernLimit and SouthernLimit retain the compatibility view for two-sided central-eclipse contours.
	NorthernLimit []SolarEclipsePathPoint
	SouthernLimit []SolarEclipsePathPoint
}

// RiseSetPhase 标识局部日食阶段。
// RiseSetPhase identifies a local eclipse phase.
type RiseSetPhase = basic.RiseSetPhase

// RiseSetDirection 标识太阳正在升起还是落下。
// RiseSetDirection identifies sunrise or sunset.
type RiseSetDirection = basic.RiseSetDirection

// 与 basic 同名的阶段与方向常量 / the phase and direction constants re-exported from basic.
const (
	RiseSetPhaseStart    = basic.RiseSetPhaseStart
	RiseSetPhaseGreatest = basic.RiseSetPhaseGreatest
	RiseSetPhaseEnd      = basic.RiseSetPhaseEnd
	RiseSetDirectionRise = basic.RiseSetDirectionRise
	RiseSetDirectionSet  = basic.RiseSetDirectionSet
)

// SolarEclipseRiseSetCurve 是一种局部阶段与日出/日落同时发生的边界。
// SolarEclipseRiseSetCurve is one boundary where a local phase coincides with sunrise or sunset.
type SolarEclipseRiseSetCurve struct {
	// Phase 是与日出或日落同时发生的局部日食阶段。
	// Phase is the local eclipse phase coinciding with sunrise or sunset.
	Phase RiseSetPhase
	// Direction 标识太阳正在升起还是落下。
	// Direction identifies whether the Sun is rising or setting.
	Direction RiseSetDirection
	// Segments 是反经线和支路跳变安全分段后的边界采样。
	// Segments are boundary samples split safely at the antimeridian and branch changes.
	Segments [][]SolarEclipsePathPoint
}

// SolarEclipsePartialAreaInfo 是 SolarEclipsePartialFootprintsInfo 的兼容别名。
// SolarEclipsePartialAreaInfo is a compatibility alias for SolarEclipsePartialFootprintsInfo.
type SolarEclipsePartialAreaInfo = SolarEclipsePartialFootprintsInfo

type solarEclipsePathCalculator func(float64, basic.SolarEclipsePathOptions) basic.SolarEclipsePathResult
type solarEclipsePartialFootprintsCalculator func(float64, basic.SolarEclipsePartialFootprintOptions) basic.SolarEclipsePartialFootprintsResult

// SolarEclipseCentralPath 计算指定日期附近的日食中心路径，默认使用 NASA bulletin Split-K 模型。
// SolarEclipseCentralPath computes the central path near the given date, using NASA bulletin Split-K by default.
func SolarEclipseCentralPath(date time.Time, options SolarEclipsePathOptions) (SolarEclipsePath, bool) {
	return SolarEclipseCentralPathNASABulletinSplitK(date, options)
}

// SolarEclipseCentralPathNASABulletinSplitK 使用 NASA bulletin Split-K 模型计算日食中心路径。
// SolarEclipseCentralPathNASABulletinSplitK computes the central path with the NASA bulletin Split-K model.
func SolarEclipseCentralPathNASABulletinSplitK(date time.Time, options SolarEclipsePathOptions) (SolarEclipsePath, bool) {
	return solarEclipseCentralPath(date, options, basic.SolarEclipseCentralPathNASABulletinSplitK)
}

// SolarEclipseCentralPathIAUSingleK 使用 IAU Single-K 模型计算日食中心路径。
// SolarEclipseCentralPathIAUSingleK computes the central path with the IAU Single-K model.
func SolarEclipseCentralPathIAUSingleK(date time.Time, options SolarEclipsePathOptions) (SolarEclipsePath, bool) {
	return solarEclipseCentralPath(date, options, basic.SolarEclipseCentralPathIAUSingleK)
}

// SolarEclipsePartialFootprints 计算指定日期附近的日食半影足迹，默认使用 NASA bulletin Split-K 模型。
// SolarEclipsePartialFootprints computes penumbral footprint samples near the given date, using NASA bulletin Split-K by default.
func SolarEclipsePartialFootprints(date time.Time, options SolarEclipsePartialFootprintOptions) (SolarEclipsePartialFootprintsInfo, bool) {
	return SolarEclipsePartialFootprintsNASABulletinSplitK(date, options)
}

// SolarEclipsePartialFootprintsNASABulletinSplitK 使用 NASA bulletin Split-K 模型计算日食半影足迹。
// SolarEclipsePartialFootprintsNASABulletinSplitK computes penumbral footprint samples with the NASA bulletin Split-K model.
func SolarEclipsePartialFootprintsNASABulletinSplitK(date time.Time, options SolarEclipsePartialFootprintOptions) (SolarEclipsePartialFootprintsInfo, bool) {
	return solarEclipsePartialFootprints(date, options, basic.SolarEclipsePartialFootprintsNASABulletinSplitK)
}

// SolarEclipsePartialFootprintsIAUSingleK 使用 IAU Single-K 模型计算日食半影足迹。
// SolarEclipsePartialFootprintsIAUSingleK computes penumbral footprint samples with the IAU Single-K model.
func SolarEclipsePartialFootprintsIAUSingleK(date time.Time, options SolarEclipsePartialFootprintOptions) (SolarEclipsePartialFootprintsInfo, bool) {
	return solarEclipsePartialFootprints(date, options, basic.SolarEclipsePartialFootprintsIAUSingleK)
}

// SolarEclipsePartialArea 计算半影足迹，是 SolarEclipsePartialFootprints 的兼容包装。
// SolarEclipsePartialArea computes penumbral footprint samples and is a compatibility wrapper for SolarEclipsePartialFootprints.
func SolarEclipsePartialArea(date time.Time, options SolarEclipsePartialAreaOptions) (SolarEclipsePartialAreaInfo, bool) {
	return SolarEclipsePartialFootprints(date, options)
}

// SolarEclipsePartialAreaNASABulletinSplitK 是 SolarEclipsePartialFootprintsNASABulletinSplitK 的兼容包装。
// SolarEclipsePartialAreaNASABulletinSplitK is a compatibility wrapper for SolarEclipsePartialFootprintsNASABulletinSplitK.
func SolarEclipsePartialAreaNASABulletinSplitK(date time.Time, options SolarEclipsePartialAreaOptions) (SolarEclipsePartialAreaInfo, bool) {
	return SolarEclipsePartialFootprintsNASABulletinSplitK(date, options)
}

// SolarEclipsePartialAreaIAUSingleK 是 SolarEclipsePartialFootprintsIAUSingleK 的兼容包装。
// SolarEclipsePartialAreaIAUSingleK is a compatibility wrapper for SolarEclipsePartialFootprintsIAUSingleK.
func SolarEclipsePartialAreaIAUSingleK(date time.Time, options SolarEclipsePartialAreaOptions) (SolarEclipsePartialAreaInfo, bool) {
	return SolarEclipsePartialFootprintsIAUSingleK(date, options)
}

func solarEclipseCentralPath(
	date time.Time,
	options SolarEclipsePathOptions,
	calculator solarEclipsePathCalculator,
) (SolarEclipsePath, bool) {
	location := date.Location()
	result := calculator(solarEclipseTimeToTTJDE(date), basicSolarEclipsePathOptions(options))
	if !result.Eclipse.HasCentral || len(result.CenterLine) == 0 {
		return SolarEclipsePath{}, false
	}

	path := SolarEclipsePath{
		Eclipse:                     solarEclipseInfoFromBasic(result.Eclipse, location),
		PathWidthDefined:            result.PathWidthDefined,
		Greatest:                    solarEclipsePathPointFromBasic(result.Greatest, location),
		CenterLine:                  solarEclipsePathPointsFromBasic(result.CenterLine, location),
		NorthernLimit:               solarEclipsePathPointsFromBasic(result.NorthernLimit, location),
		SouthernLimit:               solarEclipsePathPointsFromBasic(result.SouthernLimit, location),
		MaxCentralDuration:          solarEclipseDurationFromDays(result.MaxCentralDurationDays),
		MaxCentralDurationLongitude: result.MaxCentralDurationLongitude,
		MaxCentralDurationLatitude:  result.MaxCentralDurationLatitude,
		CentralBandSegments:         solarEclipsePathSegmentsFromBasic(result.CentralBandSegments, location),
		CentralBandSampled:          result.CentralBandSampled,
		Step:                        solarEclipsePathStepDuration(result.StepDays),
		TargetSpacingKM:             result.TargetSpacingKM,
	}
	return path, true
}

type solarEclipseGreatestTimeLevel struct {
	jde float64
	at  time.Time
}

func solarEclipseGreatestTimeLevels(values []time.Time) []solarEclipseGreatestTimeLevel {
	levels := make([]solarEclipseGreatestTimeLevel, 0, len(values))
	for _, value := range values {
		if value.IsZero() {
			continue
		}
		levels = append(levels, solarEclipseGreatestTimeLevel{jde: solarEclipseTimeToTTJDE(value), at: value})
	}
	return levels
}

func solarEclipseGreatestTimeLevelJDEs(levels []solarEclipseGreatestTimeLevel) []float64 {
	if len(levels) == 0 {
		return nil
	}
	values := make([]float64, len(levels))
	for index, level := range levels {
		values[index] = level.jde
	}
	return values
}

func solarEclipseGreatestTimeContoursFromBasic(
	contours []basic.SolarEclipseGreatestTimeContour,
	levels []solarEclipseGreatestTimeLevel,
	location *time.Location,
) []SolarEclipseGreatestTimeContour {
	if len(contours) == 0 {
		return nil
	}
	result := make([]SolarEclipseGreatestTimeContour, len(contours))
	for index, contour := range contours {
		// 核心原样回显请求的时刻取值，优先还原调用方给的时刻：JDE 往返只有微秒级误差，
		// 但整分取值会因此落到 59.999 秒，按分钟格式化时被截断成前一分钟。
		value := time.Time{}
		for _, level := range levels {
			if level.jde == contour.JDE {
				value = level.at
				break
			}
		}
		if value.IsZero() {
			value = solarEclipseTTJDEToTime(contour.JDE, location)
			// 按步长请求时核心只回显 TT 儒略日；对齐刻度的往返误差在微秒级，抹到毫秒
			// 才能保证 13:00:00 不会被格式化成 12:59。
			value = value.Round(time.Millisecond)
		}
		result[index] = SolarEclipseGreatestTimeContour{
			JDE:      contour.JDE,
			Time:     value,
			Segments: solarEclipsePathSegmentsFromBasic(contour.Segments, location),
		}
	}
	return result
}

func solarEclipsePartialFootprints(
	date time.Time,
	options SolarEclipsePartialFootprintOptions,
	calculator solarEclipsePartialFootprintsCalculator,
) (SolarEclipsePartialFootprintsInfo, bool) {
	location := date.Location()
	basicOptions := basicSolarEclipsePartialFootprintOptions(options)
	greatestTimeLevels := solarEclipseGreatestTimeLevels(options.GreatestTimeValues)
	basicOptions.GreatestTimeValues = solarEclipseGreatestTimeLevelJDEs(greatestTimeLevels)
	basicOptions.GreatestTimeStep = options.GreatestTimeStep
	result := calculator(solarEclipseTimeToTTJDE(date), basicOptions)
	if !result.Eclipse.HasPartial || len(result.Footprints) == 0 {
		return SolarEclipsePartialFootprintsInfo{}, false
	}

	footprints := SolarEclipsePartialFootprintsInfo{
		Eclipse:                    solarEclipseInfoFromBasic(result.Eclipse, location),
		Footprints:                 solarEclipsePartialFootprintsFromBasic(result.Footprints, location),
		CentralShadowFootprints:    solarEclipsePartialFootprintsFromBasic(result.CentralShadowFootprints, location),
		CentralBandFootprints:      solarEclipsePartialFootprintsFromBasic(result.CentralBandFootprints, location),
		CentralBandSegments:        solarEclipsePathSegmentsFromBasic(result.CentralBandSegments, location),
		CentralBandSampled:         result.CentralBandSampled,
		CentralBandHorizonClosures: solarEclipsePathSegmentsFromBasic(result.CentralBandHorizonClosures, location),
		PartialBandContours:        solarEclipsePathSegmentsFromBasic(result.PartialBandContours, location),
		MagnitudeContours:          solarEclipseMagnitudeContoursFromBasic(result.MagnitudeContours, location),
		GreatestTimeContours:       solarEclipseGreatestTimeContoursFromBasic(result.GreatestTimeContours, greatestTimeLevels, location),
		RiseSetCurves:              solarEclipseRiseSetCurvesFromBasic(result.RiseSetCurves, location),
		P1:                         solarEclipseOptionalPathPointFromBasic(result.P1, location),
		P2:                         solarEclipseOptionalPathPointFromBasic(result.P2, location),
		P3:                         solarEclipseOptionalPathPointFromBasic(result.P3, location),
		P4:                         solarEclipseOptionalPathPointFromBasic(result.P4, location),
		U1:                         solarEclipseOptionalPathPointFromBasic(result.U1, location),
		U2:                         solarEclipseOptionalPathPointFromBasic(result.U2, location),
		U3:                         solarEclipseOptionalPathPointFromBasic(result.U3, location),
		U4:                         solarEclipseOptionalPathPointFromBasic(result.U4, location),
		Step:                       solarEclipsePathStepDuration(result.StepDays),
		BoundaryPoints:             result.BoundaryPoints,
		CentralShadowStep:          solarEclipsePathStepDuration(result.CentralShadowStepDays),
		CentralBandStep:            solarEclipsePathStepDuration(result.CentralBandStepDays),
	}
	return footprints, true
}

func solarEclipseRiseSetCurvesFromBasic(
	curves []basic.SolarEclipseRiseSetCurve,
	location *time.Location,
) []SolarEclipseRiseSetCurve {
	if len(curves) == 0 {
		return nil
	}
	result := make([]SolarEclipseRiseSetCurve, len(curves))
	for index, curve := range curves {
		segments := make([][]SolarEclipsePathPoint, len(curve.Segments))
		for segmentIndex, segment := range curve.Segments {
			segments[segmentIndex] = solarEclipsePathPointsFromBasic(segment, location)
		}
		result[index] = SolarEclipseRiseSetCurve{
			Phase: curve.Phase, Direction: curve.Direction, Segments: segments,
		}
	}
	return result
}

func basicSolarEclipsePathOptions(options SolarEclipsePathOptions) basic.SolarEclipsePathOptions {
	basicOptions := basic.SolarEclipsePathOptions{
		TargetSpacingKM: options.TargetSpacingKM,
		SkipCentralBand: options.SkipCentralBand,
	}
	if options.Step > 0 {
		basicOptions.StepDays = options.Step.Hours() / 24
	}
	return basicOptions
}

func basicSolarEclipsePartialFootprintOptions(options SolarEclipsePartialFootprintOptions) basic.SolarEclipsePartialFootprintOptions {
	basicOptions := basic.SolarEclipsePartialFootprintOptions{
		BoundaryPoints:       options.BoundaryPoints,
		MagnitudeValues:      append([]float64(nil), options.MagnitudeValues...),
		DisableRiseSetCurves: options.DisableRiseSet,
	}
	if options.Step > 0 {
		basicOptions.StepDays = options.Step.Hours() / 24
	}
	if options.CentralShadowStep > 0 {
		basicOptions.CentralShadowStepDays = options.CentralShadowStep.Hours() / 24
	}
	if options.RiseSetStep > 0 {
		basicOptions.RiseSetStepDays = options.RiseSetStep.Hours() / 24
	}
	return basicOptions
}

func solarEclipseMagnitudeContoursFromBasic(
	contours []basic.SolarEclipseMagnitudeContour,
	location *time.Location,
) []SolarEclipseMagnitudeContour {
	if len(contours) == 0 {
		return nil
	}
	result := make([]SolarEclipseMagnitudeContour, len(contours))
	for index, contour := range contours {
		segments := make([][]SolarEclipsePathPoint, len(contour.Segments))
		for segmentIndex, segment := range contour.Segments {
			segments[segmentIndex] = solarEclipsePathPointsFromBasic(segment, location)
		}
		result[index] = SolarEclipseMagnitudeContour{
			Magnitude:     contour.Magnitude,
			Segments:      segments,
			NorthernLimit: solarEclipsePathPointsFromBasic(contour.NorthernLimit, location),
			SouthernLimit: solarEclipsePathPointsFromBasic(contour.SouthernLimit, location),
		}
	}
	return result
}

// solarEclipseDurationFromDays converts a JDE-day span into a duration.
func solarEclipseDurationFromDays(days float64) time.Duration {
	if days <= 0 || math.IsNaN(days) || math.IsInf(days, 0) {
		return 0
	}
	return time.Duration(math.Round(days * 86400 * float64(time.Second)))
}

func solarEclipsePathStepDuration(stepDays float64) time.Duration {
	return time.Duration(math.Round(stepDays * 24 * float64(time.Hour)))
}

func solarEclipsePathPointsFromBasic(points []basic.SolarEclipsePathPoint, location *time.Location) []SolarEclipsePathPoint {
	if len(points) == 0 {
		return nil
	}
	result := make([]SolarEclipsePathPoint, len(points))
	for i, point := range points {
		result[i] = solarEclipsePathPointFromBasic(point, location)
	}
	return result
}

func solarEclipsePathSegmentsFromBasic(
	segments [][]basic.SolarEclipsePathPoint,
	location *time.Location,
) [][]SolarEclipsePathPoint {
	if len(segments) == 0 {
		return nil
	}
	result := make([][]SolarEclipsePathPoint, len(segments))
	for index, segment := range segments {
		result[index] = solarEclipsePathPointsFromBasic(segment, location)
	}
	return result
}

func solarEclipsePathPointFromBasic(point basic.SolarEclipsePathPoint, location *time.Location) SolarEclipsePathPoint {
	return SolarEclipsePathPoint{
		Time:        solarEclipseTTJDEToTime(point.JDE, location),
		Longitude:   point.Longitude,
		Latitude:    point.Latitude,
		SunAltitude: point.SunAltitude,
		WidthKM:     point.WidthKM,
	}
}

func solarEclipseOptionalPathPointFromBasic(point basic.SolarEclipsePathPoint, location *time.Location) SolarEclipsePathPoint {
	if point.JDE == 0 {
		return SolarEclipsePathPoint{}
	}
	return solarEclipsePathPointFromBasic(point, location)
}

func solarEclipsePartialFootprintsFromBasic(
	footprints []basic.SolarEclipsePartialFootprint,
	location *time.Location,
) []SolarEclipsePartialFootprint {
	if len(footprints) == 0 {
		return nil
	}
	result := make([]SolarEclipsePartialFootprint, len(footprints))
	for i, footprint := range footprints {
		result[i] = SolarEclipsePartialFootprint{
			Time:       solarEclipseTTJDEToTime(footprint.JDE, location),
			Boundaries: solarEclipsePartialBoundariesFromBasic(footprint.Boundaries, location),
			Closed:     footprint.Closed,
			HorizonEnds: solarEclipsePathPointsFromBasic(
				footprint.HorizonEnds, location,
			),
		}
	}
	return result
}

func solarEclipsePartialBoundariesFromBasic(
	boundaries [][]basic.SolarEclipsePathPoint,
	location *time.Location,
) [][]SolarEclipsePathPoint {
	if len(boundaries) == 0 {
		return nil
	}
	result := make([][]SolarEclipsePathPoint, len(boundaries))
	for i, boundary := range boundaries {
		result[i] = solarEclipsePathPointsFromBasic(boundary, location)
	}
	return result
}
