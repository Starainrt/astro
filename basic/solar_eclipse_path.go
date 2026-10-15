package basic

import (
	"math"
	"sort"
	"time"

	"github.com/starainrt/astro/internal/geodata"
)

//哦~我是一颗小地球~~

const (
	solarEclipsePathDefaultStepDays  = 1.0 / 1440.0
	solarEclipsePathMinStepDays      = 1.0 / 86400.0
	solarEclipsePathMaxSampleCount   = 30000
	solarEclipsePathMaxAdaptiveDepth = 20
	solarEclipsePathVelocityStepDays = 1.0 / 1440.0
	// A sample may sit exactly on the horizon; only clearly below it is the
	// central phase unobservable and its duration meaningless.
	solarEclipsePathDurationHorizonToleranceDegrees = 0.1
	// 地球表面上的影锥足迹不可能比地球本身更宽。
	// No shadow footprint on the Earth can be wider than the Earth itself.
	solarEclipsePathMaxPossibleWidthKM = 2 * geodata.EarthMeanRadiusKM
	solarEclipsePathDuplicateTimeDays  = 1e-10
	// A short segment can still hide a large directional change at high
	// latitude.  Use a small relative sagitta tolerance so refinement follows
	// geometry without forcing every path to the finest possible spacing.
	solarEclipsePathAdaptiveCurvatureFraction = 0.02
	solarEclipsePathMinimumCurvatureKM        = 2.0

	solarEclipsePartialFootprintDefaultStepDays       = 5.0 / 1440.0
	solarEclipseRiseSetDefaultStepDays                = 2.0 / 1440.0
	solarEclipsePartialFootprintDefaultBoundaryPoints = 180
	solarEclipsePartialFootprintMinBoundaryPoints     = 12
	solarEclipsePartialFootprintMaxBoundaryPoints     = 1440
	solarEclipsePartialFootprintPointTolerance        = 1e-12
	solarEclipsePartialFootprintIterationLimit        = 10
	solarEclipsePartialFootprintMaxPointCount         = 2000000
	solarEclipsePartialFootprintTargetSpacingKM       = 200.0
	solarEclipseMagnitudeContourMaxValues             = 16
	solarEclipsePartialFootprintTransitionIterations  = 48
	// An open footprint ends where the axis-parallel line through the shadow
	// radius stops meeting the ellipsoid.  The bracket is searched around the
	// converged sampled radius, so a few doublings cover both the outside cut of
	// an on-disc shadow and the inside cut of a shadow whose axis misses the
	// Earth; the bisection below then lands on the tangency point, which sits on
	// the horizon and makes the closing arc exact.
	solarEclipseHorizonEndSearchStep      = 1e-5
	solarEclipseHorizonEndSearchLimit     = 0.25
	solarEclipseHorizonEndBisectionRounds = 40
	// The band footprints only have to resolve the swept envelope, not the
	// instantaneous rim: consecutive samples may advance far as long as their
	// union still follows the boundary. A grazing event spends most of its
	// contact interval creeping across the terminator, so a 10 s stride there
	// costs hundreds of solves for a few kilometres of travel.
	solarEclipseCentralBandStepDays                  = 10.0 / 86400.0
	solarEclipseNonCentralBandFallbackStepDays       = 60.0 / 86400.0
	solarEclipseNonCentralBandPreciseMaxDurationDays = 20.0 / 1440.0
	solarEclipseCentralBandContactFineStepDays       = 1.0 / 86400.0
	solarEclipseCentralBandContactFineWindowDays     = 45.0 / 86400.0
	solarEclipseCentralBandBoundaryPoints            = 90
	// The union decimates every footprint ring to the union spacing before it
	// merges them, and the result is relaxed afterwards, so solving the band
	// footprints finer than this only buys trigonometry.
	solarEclipseCentralBandTargetSpacingKM           = 10.0
	solarEclipseNonCentralTotalBandTargetSpacingKM   = 25.0
	solarEclipseCentralLimitTargetSpacingKM          = 200.0
	solarEclipseShadowFootprintAdaptiveMaxDepth      = 12
	solarEclipseShadowContactSearchStepDays          = 10.0 / 1440.0
	solarEclipseShadowContactSearchSpanDays          = 0.75
	solarEclipseShadowContactToleranceDays           = 1e-9
	solarEclipseMagnitudeContourBoundaryPoints       = 180
	solarEclipseMagnitudeContourTargetSpacingKM      = 500.0
	solarEclipseMagnitudeContourArcStepDegrees       = 4.0
	solarEclipseMagnitudeContourMinArcStepDegrees    = 0.01
	solarEclipseMagnitudeContourMaxArcSteps          = 2000
	solarEclipseMagnitudeContourTimeScale            = 360.0
	solarEclipseMagnitudeContourFallbackBearings     = 72
	solarEclipseMagnitudeContourFallbackDistances    = 12
	solarEclipseRiseSetBoundaryPoints                = 180
	solarEclipseRiseSetCriticalStepDays              = 1.0 / 1440.0
	solarEclipseRiseSetDerivativeStepDays            = 5.0 / 86400.0
	solarEclipseRiseSetTargetSpacingKM               = 500.0
	solarEclipseRiseSetAttachmentDistanceToleranceKM = 5.0
	solarEclipseRiseSetAttachmentTimeToleranceDays   = 2.0 / 86400.0
	solarEclipseRiseSetPhaseConnectionLimitKM        = 3000.0
	solarEclipseRiseSetTimeEpsilonDays               = 1e-8
)

type solarEclipseShadowKind uint8

const (
	solarEclipsePenumbralShadow solarEclipseShadowKind = iota
	solarEclipseCentralShadow
)

// SolarEclipsePathOptions 控制日食中心路径采样。
// SolarEclipsePathOptions controls central solar eclipse path sampling.
type SolarEclipsePathOptions struct {
	// StepDays 是基础时间采样步长，单位为日；<=0 时使用 1 分钟。
	// StepDays is the base time step in days; values <= 0 use one minute.
	StepDays float64
	// TargetSpacingKM 是相邻中心线点的最大目标地表距离；<=0 时不按距离加密。
	// TargetSpacingKM is the target maximum ground spacing between centerline points; values <= 0 disable spacing refinement.
	TargetSpacingKM float64
	// DeltaTSeconds 显式 ΔT（秒），<=0 用进程级模型；与单时刻阴影层同口径，避免同一张图上
	// 路径几何与瞬时足迹各自使用不同的 ΔT（两者会沿经度错开 0.4651·|ΔΔT|·cosφ 千米）。
	// DeltaTSeconds is an explicit ΔT in seconds; values <= 0 use the process-wide
	// model. The path and the instantaneous footprints of one figure must share it.
	DeltaTSeconds float64
	// SunRadiusModel 太阳半径口径，零值为标准档。
	// SunRadiusModel is the solar radius convention; the zero value is the standard one.
	SunRadiusModel SolarEclipseSunRadiusModel
	// SkipCentralBand 表示调用方已经持有同一场日食的完整足迹结果（其中包含
	// CentralBandSegments），本次只求解中心线、南北限界与地平线端点，不重复重建中心食带。
	// 单场日食的中心带足迹是整条链路里最贵的一段，同时取足迹与路径时重复计算会翻倍。
	// SkipCentralBand reports that the caller already holds the full-footprint
	// result for the same eclipse, whose CentralBandSegments are authoritative,
	// so this call only solves the center line, the limits and the contact
	// points instead of rebuilding the central band. The band footprints are the
	// most expensive stage of the pipeline and would otherwise be computed twice
	// by callers that ask for both products.
	SkipCentralBand bool
	// LunarLimb 月缘剖面，nil 为圆月面；修正南北界、带宽、中心带足迹、最长中心食时长与全球始末，中心线不变 / lunar limb profile for limits, widths, band, longest duration and P1/P4.
	LunarLimb LunarLimb
}

// SolarEclipsePathPoint 表示日食路径上的一个地理点。
// SolarEclipsePathPoint is one geographic point on a solar eclipse path.
type SolarEclipsePathPoint struct {
	// JDE 是力学时儒略日, TT Julian ephemeris day.
	JDE float64
	// Longitude 经度，东正西负, longitude in degrees, east positive.
	Longitude float64
	// Latitude 纬度，北正南负, latitude in degrees, north positive.
	Latitude float64
	// SunAltitude 太阳高度角，单位度, Sun altitude in degrees.
	SunAltitude float64
	// WidthKM 中心食带宽度，单位千米；仅中心线点有意义；退化地平线切点无法形成稳定成对横截面时为 0。
	// WidthKM is the central path width in kilometers; it is meaningful for centerline points and is 0 at a degenerate horizon contact without a stable paired cross-section.
	WidthKM float64
}

// SolarEclipsePathResult 表示一次中心日食的路径数据。
// SolarEclipsePathResult contains central solar eclipse path data.
type SolarEclipsePathResult struct {
	// Eclipse 是对应的全局日食结果, related global solar eclipse result.
	Eclipse SolarEclipseResult
	// PathWidthDefined 表示本结果的带宽是否有定义：两限存在且上下两条限界线都非空时才为 true；
	// 为 false 时 Eclipse.PathWidthKM 与 Greatest.WidthKM 都是 0，调用方引用带宽前必须先看这里。
	// PathWidthDefined reports whether this result has a defined band width: both
	// limits must exist and both limit lines must be non-empty. When it is false,
	// Eclipse.PathWidthKM and Greatest.WidthKM are 0 and callers must check this
	// flag before quoting a width.
	PathWidthDefined bool
	// Greatest 是食甚点/最佳观测点, greatest eclipse point.
	Greatest SolarEclipsePathPoint
	// MaxCentralDurationDays 是中心线上最长的中心食时长（单位为日），并给出其发生位置。
	// 与目录（NASA）口径不同：目录值取食甚点，这里是整条中心线上的最大值。
	// MaxCentralDurationDays is the longest central phase on the center line, in
	// days, with the location where it occurs. The catalogued "central duration"
	// is the value at greatest eclipse; this is the maximum along the track.
	MaxCentralDurationDays      float64
	MaxCentralDurationLongitude float64
	MaxCentralDurationLatitude  float64
	// CenterLine 是中心线, central line.
	CenterLine []SolarEclipsePathPoint
	// NorthernLimit 是中心食带北界近似线, approximate northern limit of the central path.
	NorthernLimit []SolarEclipsePathPoint
	// SouthernLimit 是中心食带南界近似线, approximate southern limit of the central path.
	SouthernLimit []SolarEclipsePathPoint
	// CentralBandSegments is the authoritative continuous central-band envelope.
	CentralBandSegments [][]SolarEclipsePathPoint
	// CentralBandSampled 表示上面的包络是用采样瞬时足迹重建的（解析包络不适用）。
	// CentralBandSampled reports that the envelope above was reconstructed from
	// the sampled instantaneous footprints because no analytic envelope
	// described this event.
	CentralBandSampled bool
	// CentralLimitsLimbApplied 南北界至少一点按月缘剖面平移，带宽随之修正 / the central limits were moved by the lunar limb.
	CentralLimitsLimbApplied bool
	// StepDays 是实际采用的基础时间采样步长，单位为日。
	// StepDays is the effective base time step in days.
	StepDays float64
	// TargetSpacingKM 是实际采用的目标空间采样距离，单位千米。
	// TargetSpacingKM is the effective target spacing in kilometers.
	TargetSpacingKM float64
}

// SolarEclipsePartialFootprintOptions 控制日食偏食半影足迹采样。
// SolarEclipsePartialFootprintOptions controls solar eclipse penumbral footprint sampling.
type SolarEclipsePartialFootprintOptions struct {
	// StepDays 是基础时间采样步长，单位为日；<=0 时使用 5 分钟。
	// StepDays is the base time step in days; values <= 0 use five minutes.
	StepDays float64
	// BoundaryPoints 是每个瞬时半影边界的角向采样点数；<=0 时使用 180。
	// BoundaryPoints is the angular sample count for each instantaneous penumbral boundary; values <= 0 use 180.
	BoundaryPoints int
	// CentralShadowStepDays 是本影/反本影瞬时足迹的时间步长，单位为日；<=0 时不计算。
	// CentralShadowStepDays is the umbral/antumbral footprint step in days; values <= 0 disable it.
	CentralShadowStepDays float64
	// RiseSetStepDays 独立采样地平阶段曲线；<=0 时使用 2 分钟。
	// RiseSetStepDays samples horizon curves independently; values <=0 use two minutes.
	RiseSetStepDays float64
	// DisableRiseSetCurves 禁用六类日升日落阶段边界。
	// DisableRiseSetCurves disables the six sunrise/sunset boundaries.
	DisableRiseSetCurves bool
	// MagnitudeValues 是要计算的地方最大食分等值线；空值不计算，线条使用独立的自适应空间采样。
	// MagnitudeValues requests local maximum-magnitude contours; empty disables them, and contours use independent adaptive spatial sampling.
	MagnitudeValues []float64
	// GreatestTimeValues 是要计算的地方食甚时刻等值线取值（TT 儒略日，最多 64 条，超出按时间截断）；
	// 只有确实存在该时刻食甚轨迹的取值才会出现在结果里，所以返回条数可能少于请求条数。
	// 空值时改用 GreatestTimeStep。每条等时线用固定时刻的残差零集延拓，成本正比于曲线长度而不是可见域面积。
	// GreatestTimeValues requests local greatest-eclipse time isolines as TT Julian ephemeris days
	// (at most 64, truncated in time order); when empty, GreatestTimeStep is used instead. Each
	// isochrone is continued along the zero set of a fixed-instant residual, so the cost scales
	// with curve length rather than with the visible area.
	GreatestTimeValues []float64
	// GreatestTimeStep 是等时线间隔；仅在 GreatestTimeValues 为空时生效，非正值不计算等时线。
	// 取值对齐到 UTC 整刻度并覆盖地球范围的偏食窗口，最多 64 条；显示层若需要按展示时区对齐，
	// 应自行生成时刻后改用 GreatestTimeValues。
	// GreatestTimeStep is the isochrone interval; it applies only when GreatestTimeValues is empty,
	// and non-positive values disable the isolines. Levels align to UTC ticks across the global
	// partial-eclipse window, at most 64; display layers that need the viewing timezone grid should
	// generate the instants themselves and pass GreatestTimeValues instead.
	GreatestTimeStep time.Duration
	// DeltaTSeconds 显式 ΔT（秒），<=0 用进程级模型；与单时刻阴影层同口径。
	// DeltaTSeconds is an explicit ΔT in seconds; values <= 0 use the process-wide model.
	DeltaTSeconds float64
	// SunRadiusModel 太阳半径口径，零值为标准档。
	// SunRadiusModel is the solar radius convention; the zero value is the standard one.
	SunRadiusModel SolarEclipseSunRadiusModel
	// LunarLimb 月缘剖面，nil 为圆月面；修正偏食可见域外边界（零食分包络与初亏/复圆升落曲线），
	// 内嵌全球结果与 P1/P4 也与全球日食接口同源地修正；食分线与等时线仍按圆月面 / lunar limb profile for the outer partial boundary, embedded global result and P1/P4.
	LunarLimb LunarLimb
}

// SolarEclipsePartialAreaOptions 是 SolarEclipsePartialFootprintOptions 的兼容别名。
// SolarEclipsePartialAreaOptions is a compatibility alias for SolarEclipsePartialFootprintOptions.
type SolarEclipsePartialAreaOptions = SolarEclipsePartialFootprintOptions

// SolarEclipsePartialFootprint 表示某一时刻的半影足迹边界。
// SolarEclipsePartialFootprint is the penumbral footprint boundary at one instant.
type SolarEclipsePartialFootprint struct {
	// JDE 是力学时儒略日, TT Julian ephemeris day.
	JDE float64
	// Boundaries 是半影边界分段；反经线或无效投影会拆成多段。
	// Boundaries are segmented penumbral boundary polylines, split at invalid projections or the antimeridian.
	Boundaries [][]SolarEclipsePathPoint
	// Closed 表示 Boundaries 是否构成一个闭合边界。
	// Closed indicates whether Boundaries form one closed boundary.
	Closed bool
	// HorizonEnds 是未闭合边界两端延伸到地平圈的擦地点，顺序与 Boundaries 的走向一致
	// （HorizonEnds[0] 贴近边界起点，HorizonEnds[1] 贴近边界终点）；边界自身闭合时为空。
	// 被地平线切断的瞬时阴影区域由「物理边界 + 两个擦地点之间的地平弧」闭合；擦地点是
	// 阴影锥面与地表的切点，其太阳高度为 0，所以该闭合弧是精确结果而不是启发式。
	// HorizonEnds are the two limb-grazing points where an open boundary reaches the
	// horizon, ordered like Boundaries (HorizonEnds[0] near the boundary start,
	// HorizonEnds[1] near its end); empty when the boundary closes on itself. A region
	// cut by the horizon is closed by the physical boundary plus the horizon arc
	// between these grazing points, which are exact cone-surface tangency points with
	// zero solar altitude rather than a heuristic.
	HorizonEnds []SolarEclipsePathPoint
}

// SolarEclipsePartialFootprintsResult 表示一次日食的偏食半影足迹序列。
// SolarEclipsePartialFootprintsResult contains penumbral footprint samples for a solar eclipse.
type SolarEclipsePartialFootprintsResult struct {
	// Eclipse 是对应的全局日食结果, related global solar eclipse result.
	Eclipse SolarEclipseResult
	// Footprints 是按时间采样的瞬时半影足迹, sampled instantaneous penumbral footprints.
	Footprints []SolarEclipsePartialFootprint
	// CentralShadowFootprints 是按时间采样的本影/反本影足迹。
	// CentralShadowFootprints are sampled umbral/antumbral footprints.
	CentralShadowFootprints []SolarEclipsePartialFootprint
	// CentralBandFootprints 是始终计算的低成本本影/反本影端部样本，用于闭合中心食带。
	// CentralBandFootprints are always-computed lightweight umbral/antumbral end samples used to close the central band.
	CentralBandFootprints []SolarEclipsePartialFootprint
	// CentralBandSegments 是地方中心食条件的连续闭合包络。
	// CentralBandSegments are continuous closed envelopes of the local central-eclipse condition.
	CentralBandSegments [][]SolarEclipsePathPoint
	// CentralBandSampled 表示上面的包络是用采样瞬时足迹重建的（解析包络不适用）。
	// CentralBandSampled reports that CentralBandSegments were reconstructed from
	// the sampled instantaneous footprints because no analytic envelope applied.
	CentralBandSampled bool
	// CentralBandHorizonClosures 是两限界掠地事件中分别连接首尾两侧限界的食甚地平线弧。
	// CentralBandHorizonClosures are the greatest-at-horizon arcs joining both ends of a grazing two-limit event.
	CentralBandHorizonClosures [][]SolarEclipsePathPoint
	// PartialBandContours 是地方最大食分等于零的连续可见包络，用于与日出日落边界闭合偏食可见域。
	// PartialBandContours are the continuous zero local-maximum-magnitude envelopes used with horizon boundaries to close the partial-eclipse visibility region.
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
	// RiseSetTopologyDegraded 表示请求了升落曲线但六类边界未通过拓扑校验（每曲线 16 段、
	// 总计 32 段的段数上限，端点共享与时间单调），RiseSetCurves 是被截断的降级结果。
	// RiseSetTopologyDegraded reports that the requested rise/set curves failed the
	// topology audit (the 16-segment per-curve and 32-segment total caps, shared
	// endpoints and monotonic times), so RiseSetCurves is a truncated result.
	RiseSetTopologyDegraded bool
	// PartialBoundaryLimbApplied 表示偏食可见域外边界至少有一点按月缘剖面平移 / the outer partial boundary was moved by the lunar limb.
	PartialBoundaryLimbApplied bool
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
	// StepDays 是实际采用的基础时间采样步长，单位为日。
	// StepDays is the effective base time step in days.
	StepDays float64
	// BoundaryPoints 是实际采用的边界角向采样点数。
	// BoundaryPoints is the effective angular sample count for each boundary.
	BoundaryPoints int
	// CentralShadowStepDays 是本影/反本影足迹的实际采样步长；0 表示未计算。
	// CentralShadowStepDays is the effective umbral/antumbral footprint step; zero means disabled.
	CentralShadowStepDays float64
	// CentralBandStepDays 是中心食带足迹的最细实际采样步长。
	// CentralBandStepDays is the finest effective sampling step for central-band footprints.
	CentralBandStepDays float64
}

// SolarEclipseGreatestTimeContour 是一个固定地方食甚时刻的等值线支路集合。
// SolarEclipseGreatestTimeContour contains the continuous branches of one fixed local greatest-eclipse time.
type SolarEclipseGreatestTimeContour struct {
	// JDE 是该等值线表示的力学时儒略日，也就是各支路上地方食甚发生的时刻。
	// JDE is the TT Julian ephemeris day represented by this contour, the local greatest-eclipse instant along every branch.
	JDE float64
	// Segments 是该时刻的连续等时线支路；一条支路两端止于地平线或偏食可见域边界。
	// Segments are continuous isochrone branches; each branch ends at the horizon or the partial-visibility boundary.
	Segments [][]SolarEclipsePathPoint
}

// SolarEclipseMagnitudeContour 是一条地方最大食分等值线的连续支路集合。
// SolarEclipseMagnitudeContour contains the continuous branches of one local maximum-magnitude contour.
type SolarEclipseMagnitudeContour struct {
	// Magnitude 是该等值线表示的地方最大食分。
	// Magnitude is the local maximum eclipse magnitude represented by this contour.
	Magnitude float64
	// Segments 是地方最大食分等值线的连续支路；临近地平线时局部食甚时刻可以沿空间支路折返。
	// Segments are continuous local maximum-magnitude branches; local greatest times may fold along a spatial branch near the horizon.
	Segments [][]SolarEclipsePathPoint
	// NorthernLimit 和 SouthernLimit 保留两侧中心食等值线的兼容视图。
	// NorthernLimit and SouthernLimit retain the compatibility view for two-sided central-eclipse contours.
	NorthernLimit []SolarEclipsePathPoint
	SouthernLimit []SolarEclipsePathPoint
}

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

// SolarEclipsePartialAreaResult 是 SolarEclipsePartialFootprintsResult 的兼容别名。
// SolarEclipsePartialAreaResult is a compatibility alias for SolarEclipsePartialFootprintsResult.
type SolarEclipsePartialAreaResult = SolarEclipsePartialFootprintsResult

// SolarEclipseCentralPath 计算给定近朔时刻附近的日食中心路径，默认使用 NASA bulletin Split-K 模型。
// SolarEclipseCentralPath computes the central path near the given new-moon seed, using NASA bulletin Split-K by default.
func SolarEclipseCentralPath(seedJDE float64, options SolarEclipsePathOptions) SolarEclipsePathResult {
	return SolarEclipseCentralPathNASABulletinSplitK(seedJDE, options)
}

// SolarEclipseCentralPathIAUSingleKWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SolarEclipseCentralPathIAUSingleKWithInstance(seedJDE float64, options SolarEclipsePathOptions, inst *Instance) SolarEclipsePathResult {
	return solarEclipseCentralPath(seedJDE, SolarEclipseModelIAUSingleK, options, inst)
}

// SolarEclipseCentralPathNASABulletinSplitKWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SolarEclipseCentralPathNASABulletinSplitKWithInstance(seedJDE float64, options SolarEclipsePathOptions, inst *Instance) SolarEclipsePathResult {
	return solarEclipseCentralPath(seedJDE, SolarEclipseModelNASABulletinSplitK, options, inst)
}

// SolarEclipsePartialFootprints 计算给定近朔时刻附近的日食偏食半影足迹序列，默认使用 NASA bulletin Split-K 模型。
// SolarEclipsePartialFootprints computes penumbral footprint samples near the given new-moon seed, using NASA bulletin Split-K by default.
func SolarEclipsePartialFootprints(seedJDE float64, options SolarEclipsePartialFootprintOptions) SolarEclipsePartialFootprintsResult {
	return SolarEclipsePartialFootprintsNASABulletinSplitK(seedJDE, options)
}

// SolarEclipsePartialFootprintsIAUSingleKWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SolarEclipsePartialFootprintsIAUSingleKWithInstance(seedJDE float64, options SolarEclipsePartialFootprintOptions, inst *Instance) SolarEclipsePartialFootprintsResult {
	return solarEclipsePartialFootprints(seedJDE, SolarEclipseModelIAUSingleK, options, inst)
}

// SolarEclipsePartialFootprintsNASABulletinSplitKWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func SolarEclipsePartialFootprintsNASABulletinSplitKWithInstance(seedJDE float64, options SolarEclipsePartialFootprintOptions, inst *Instance) SolarEclipsePartialFootprintsResult {
	return solarEclipsePartialFootprints(seedJDE, SolarEclipseModelNASABulletinSplitK, options, inst)
}

// SolarEclipsePartialArea 计算日食偏食半影足迹序列，是 SolarEclipsePartialFootprints 的兼容包装。
// SolarEclipsePartialArea computes penumbral footprint samples and is a compatibility wrapper for SolarEclipsePartialFootprints.
func SolarEclipsePartialArea(seedJDE float64, options SolarEclipsePartialAreaOptions) SolarEclipsePartialAreaResult {
	return SolarEclipsePartialFootprints(seedJDE, options)
}

// SolarEclipsePartialAreaIAUSingleK 计算日食偏食半影足迹序列，是 SolarEclipsePartialFootprintsIAUSingleK 的兼容包装。
// SolarEclipsePartialAreaIAUSingleK is a compatibility wrapper for SolarEclipsePartialFootprintsIAUSingleK.
func SolarEclipsePartialAreaIAUSingleK(seedJDE float64, options SolarEclipsePartialAreaOptions) SolarEclipsePartialAreaResult {
	return SolarEclipsePartialFootprintsIAUSingleK(seedJDE, options)
}

// SolarEclipsePartialAreaNASABulletinSplitK 计算日食偏食半影足迹序列，是 SolarEclipsePartialFootprintsNASABulletinSplitK 的兼容包装。
// SolarEclipsePartialAreaNASABulletinSplitK is a compatibility wrapper for SolarEclipsePartialFootprintsNASABulletinSplitK.
func SolarEclipsePartialAreaNASABulletinSplitK(seedJDE float64, options SolarEclipsePartialAreaOptions) SolarEclipsePartialAreaResult {
	return SolarEclipsePartialFootprintsNASABulletinSplitK(seedJDE, options)
}

func solarEclipseCentralPath(seedJDE float64, model SolarEclipseRadiusModel, options SolarEclipsePathOptions, instances ...*Instance) SolarEclipsePathResult {
	inst := instanceValue(instances)

	options = normalizeSolarEclipsePathOptions(options)
	newMoonJDE := solarEclipseAnchorNewMoon(seedJDE, inst)
	solver := newSolarEclipseSolverWithOptions(newMoonJDE, SolarEclipseOptions{
		RadiusModel:    model,
		SunRadiusModel: options.SunRadiusModel,
	}, inst).withDeltaTSeconds(options.DeltaTSeconds)
	result := solver.eclipseResult()
	path := SolarEclipsePathResult{
		Eclipse:         result,
		StepDays:        options.StepDays,
		TargetSpacingKM: options.TargetSpacingKM,
	}
	if !result.HasCentral {
		return path
	}

	greatest, ok := solver.centralPathPointAt(result.GreatestEclipse)
	if !ok {
		greatest = SolarEclipsePathPoint{
			JDE:       result.GreatestEclipse,
			Longitude: result.GreatestLongitude,
			Latitude:  result.GreatestLatitude,
			WidthKM:   result.PathWidthKM,
			SunAltitude: solarEclipseSunAltitudeAtGreatest(
				result.GreatestEclipse,
				result.GreatestLongitude,
				result.GreatestLatitude,
				solver.besselAxisAt(result.GreatestEclipse).gst, inst,
			) / rad,
		}
	}
	greatest.Longitude = result.GreatestLongitude
	greatest.Latitude = result.GreatestLatitude
	greatest.WidthKM = result.PathWidthKM
	path.Greatest = greatest

	samplingOptions := options
	samplingOptions.TargetSpacingKM = 0
	centerLine, stepDays := solver.centralPathPoints(
		result.CentralBeginOnEarth,
		result.CentralEndOnEarth,
		result.GreatestEclipse,
		samplingOptions,
	)
	centerLine = solver.attachCentralAxisContactPoints(
		centerLine, result.CentralBeginOnEarth, result.CentralEndOnEarth,
	)
	if result.Type == SolarEclipseHybrid {
		centerLine = solver.attachCentralMagnitudeOneTransitions(centerLine)
	}
	if options.TargetSpacingKM > 0 {
		centerLine = solver.refineCentralPathSpacing(centerLine, options.TargetSpacingKM)
	}
	path.StepDays = stepDays
	// 每对横截面只解一次：限界线与宽度都取自同一份配对结果，重复求解是这条链上最大的
	// 单点分配与耗时来源。
	northernLimit, southernLimit, pairedLimits := solver.centralPathLimitPairs(centerLine)
	path.NorthernLimit, path.SouthernLimit = solarEclipseFilterCentralPathLimits(
		northernLimit, southernLimit, pairedLimits,
	)
	// The paired-limit width runs first, then the contact pass restores the exact
	// U1/U4 samples, whose cross-section is solved from the contact pair rather
	// than from the interior samples.
	solver.setCenterLinePairWidths(
		centerLine, result.Centrality, northernLimit, southernLimit, pairedLimits,
	)
	limb := options.LunarLimb
	if limb != nil {
		path.Eclipse = solver.withLunarLimb(path.Eclipse, limb)
		greatestIndex, circularWidth := solarEclipseNearestCenterWidth(centerLine, result.GreatestEclipse)
		northernLimb, northWidthApplied := solver.limbCorrectCentralLimits(northernLimit, limb)
		southernLimb, southWidthApplied := solver.limbCorrectCentralLimits(southernLimit, limb)
		path.CentralLimitsLimbApplied = northWidthApplied || southWidthApplied
		solver.setCenterLinePairWidths(centerLine, result.Centrality, northernLimb, southernLimb, pairedLimits)
		// 目录带宽沿用解析口径，只叠加食甚附近成对横截面上的月缘增量。
		if greatestIndex >= 0 && circularWidth > 0 && centerLine[greatestIndex].WidthKM > 0 {
			delta := centerLine[greatestIndex].WidthKM - circularWidth
			path.Eclipse.PathWidthKM += delta
			path.Greatest.WidthKM += delta
		}
	}
	if result.Centrality == SolarEclipseCentralTwoLimits {
		// A one-limit event has no paired cross-section at all, so every sample
		// keeps the 0 that the catalogues also print for it.
		setSolarEclipseCentralContactWidths(centerLine, path.NorthernLimit, path.SouthernLimit)
	}
	path.CenterLine = centerLine
	durationSolver := solver
	if limb != nil {
		durationSolver = solver.withLimbParams(limb)
	}
	path.MaxCentralDurationDays, path.MaxCentralDurationLongitude, path.MaxCentralDurationLatitude =
		durationSolver.centralPathMaxCentralDuration(centerLine)
	// A two-limit central eclipse starts and ends at the external umbral
	// contacts, where the northern and southern limits meet.  The center line
	// begins later because its Earth-axis intersection is undefined before the
	// axis reaches the surface.  Add those exact contact points so consumers do
	// not close the band across the first/last coarse center-line samples.
	if result.Type != SolarEclipseHybrid &&
		result.Centrality == SolarEclipseCentralTwoLimits && len(path.NorthernLimit) > 0 {
		first, last, contactsOK := solver.shadowContactPair(
			result.GreatestEclipse, solarEclipseCentralShadow, false,
		)
		if contactsOK {
			// The open footprint endpoints before the axis reaches the Earth are
			// horizon branches, not the two side limits.  Appending them here
			// makes one side turn back and self-intersect near U1/U4.  The exact
			// contact points and the first/last cross-sections form the required
			// short end caps without introducing that false branch.
			path.NorthernLimit = append([]SolarEclipsePathPoint{first}, path.NorthernLimit...)
			path.SouthernLimit = append([]SolarEclipsePathPoint{first}, path.SouthernLimit...)
			path.NorthernLimit = append(path.NorthernLimit, last)
			path.SouthernLimit = append(path.SouthernLimit, last)
		}
	}
	path.NorthernLimit, path.SouthernLimit = solver.refineCentralPathLimitSpacing(
		path.NorthernLimit,
		path.SouthernLimit,
		result.CentralBeginOnEarth,
		result.CentralEndOnEarth,
		solarEclipseCentralLimitTargetSpacingKM,
	)
	var northApplied, southApplied bool
	path.NorthernLimit, northApplied = solver.limbCorrectCentralLimits(path.NorthernLimit, limb)
	path.SouthernLimit, southApplied = solver.limbCorrectCentralLimits(path.SouthernLimit, limb)
	path.CentralLimitsLimbApplied = path.CentralLimitsLimbApplied || northApplied || southApplied
	path.PathWidthDefined = result.Centrality == SolarEclipseCentralTwoLimits &&
		len(path.NorthernLimit) > 0 && len(path.SouthernLimit) > 0
	if !path.PathWidthDefined {
		// 泛化到路径层：单侧极限只解出一侧限界，成对横截面凑不齐的事件连限界线都为空，
		// 两者的解析带宽同样无定义，与中心线逐点宽度一并置 0，避免发散值从路径接口外泄。
		// 后一种情形在当前扫描范围内不可达（1000–3000 年 4773 场、1800–2200 年 584 场 two_limits 均未触发），只是防御。
		path.Eclipse.PathWidthKM = 0
		path.Eclipse.PathWidthDefined = false
		path.Greatest.WidthKM = 0
	}
	if options.SkipCentralBand {
		return path
	}
	// Generate the same continuous envelope exposed by the full-footprint API
	// so path-only callers cannot reconstruct a different polar ribbon.
	band, _ := solver.centralBandWithResult(result, normalizeSolarEclipsePartialFootprintOptions(SolarEclipsePartialFootprintOptions{
		StepDays: options.StepDays, BoundaryPoints: 96, DisableRiseSetCurves: true,
	}))
	path.CentralBandSegments = band.CentralBandSegments
	if limb != nil {
		for index, segment := range path.CentralBandSegments {
			path.CentralBandSegments[index], _ = solver.limbCorrectCentralLimits(segment, limb)
		}
	}
	path.CentralBandSampled = band.CentralBandSampled
	return path
}

// solarEclipseNearestCenterWidth 返回最接近 jde 的中心线样本序号与其带宽，没有样本时序号为 −1。
func solarEclipseNearestCenterWidth(centerLine []SolarEclipsePathPoint, jde float64) (int, float64) {
	best, bestDelta := -1, math.Inf(1)
	for index, point := range centerLine {
		if delta := math.Abs(point.JDE - jde); delta < bestDelta {
			best, bestDelta = index, delta
		}
	}
	if best < 0 {
		return -1, 0
	}
	return best, centerLine[best].WidthKM
}

func (solver solarEclipseSolver) attachCentralMagnitudeOneTransitions(
	points []SolarEclipsePathPoint,
) []SolarEclipsePathPoint {
	for _, transition := range solver.centralMagnitudeOneTransitions(points) {
		duplicate := false
		for _, point := range points {
			if math.Abs(point.JDE-transition.JDE) <= solarEclipsePathDuplicateTimeDays {
				duplicate = true
				break
			}
		}
		if !duplicate {
			points = append(points, transition)
		}
	}
	sort.Slice(points, func(first, second int) bool { return points[first].JDE < points[second].JDE })
	return normalizeSolarEclipsePathPointSeries(points)
}

func normalizeSolarEclipsePathOptions(options SolarEclipsePathOptions) SolarEclipsePathOptions {
	if options.StepDays <= 0 || math.IsNaN(options.StepDays) || math.IsInf(options.StepDays, 0) {
		options.StepDays = solarEclipsePathDefaultStepDays
	}
	if options.StepDays < solarEclipsePathMinStepDays {
		options.StepDays = solarEclipsePathMinStepDays
	}
	if options.TargetSpacingKM <= 0 || math.IsNaN(options.TargetSpacingKM) || math.IsInf(options.TargetSpacingKM, 0) {
		options.TargetSpacingKM = 0
	}
	return options
}

func solarEclipsePartialFootprints(
	seedJDE float64,
	model SolarEclipseRadiusModel,
	options SolarEclipsePartialFootprintOptions, instances ...*Instance) SolarEclipsePartialFootprintsResult {
	inst := instanceValue(instances)

	return solarEclipsePartialFootprintsWithResult(
		seedJDE, model, options, solarEclipseWithDeltaT(seedJDE, SolarEclipseOptions{
			RadiusModel:    model,
			SunRadiusModel: options.SunRadiusModel,
		}, options.DeltaTSeconds, inst), inst,
	)
}

func solarEclipsePartialFootprintsWithResult(
	seedJDE float64,
	model SolarEclipseRadiusModel,
	options SolarEclipsePartialFootprintOptions,
	result SolarEclipseResult, instances ...*Instance) SolarEclipsePartialFootprintsResult {
	inst := instanceValue(instances)

	options = normalizeSolarEclipsePartialFootprintOptions(options)
	footprintsResult := SolarEclipsePartialFootprintsResult{
		Eclipse:               result,
		StepDays:              options.StepDays,
		BoundaryPoints:        options.BoundaryPoints,
		CentralShadowStepDays: options.CentralShadowStepDays,
	}
	if !result.HasPartial {
		return footprintsResult
	}

	newMoonJDE := solarEclipseAnchorNewMoon(seedJDE, inst)
	solver := newSolarEclipseSolverWithOptions(newMoonJDE, SolarEclipseOptions{
		RadiusModel:    model,
		SunRadiusModel: options.SunRadiusModel,
	}, inst).withDeltaTSeconds(options.DeltaTSeconds)
	footprintsResult, totalMagnitudeOneSegments := solver.centralBandWithResult(result, options)
	// 有月缘时先求全球月缘结果：偏食足迹与食甚等时线必须落在最终报告的 P1–P4 之内。
	global, window := result, result
	if options.LunarLimb != nil {
		global = solver.withLunarLimb(result, options.LunarLimb)
		if global.PartialBeginOnEarth < result.GreatestEclipse && result.GreatestEclipse < global.PartialEndOnEarth {
			window = global
		}
	}
	// Partial and central-shadow sweeps used to enforce the point budget
	// independently. A high-resolution request could therefore allocate
	// nearly two budgets for each sweep before the caller saw the result.
	// Reserve the already-built central-band samples and share the remaining
	// budget between the main time series before either sweep is generated.
	partialTimes, _ := solarEclipseMovingDiskEngine().sampleTimes(
		window.PartialBeginOnEarth,
		window.PartialEndOnEarth,
		result.GreatestEclipse,
		options.StepDays,
	)
	var shadowTimes []float64
	if options.CentralShadowStepDays > 0 &&
		footprintsResult.U1.JDE != 0 && footprintsResult.U4.JDE != 0 {
		shadowTimes, _ = solarEclipseMovingDiskEngine().sampleTimes(
			footprintsResult.U1.JDE,
			footprintsResult.U4.JDE,
			result.GreatestEclipse,
			options.CentralShadowStepDays,
		)
	}
	options.BoundaryPoints = solarEclipseSharedBoundaryPoints(
		options.BoundaryPoints,
		len(partialTimes),
		len(shadowTimes),
		solarEclipseFootprintPointCount(footprintsResult.CentralBandFootprints),
	)
	footprintsResult.BoundaryPoints = options.BoundaryPoints
	if !options.DisableRiseSetCurves {
		footprintsResult.PartialBandContours = solver.magnitudeContourSegments(
			result.PartialBeginOnEarth,
			result.PartialEndOnEarth,
			result.CentralBeginOnEarth,
			result.CentralEndOnEarth,
			result.GreatestEclipse,
			0,
			options.StepDays,
			false,
		)
	}
	if options.LunarLimb != nil && !options.DisableRiseSetCurves {
		footprintsResult.PartialBoundaryLimbApplied = solver.limbCorrectPartialBoundary(&footprintsResult, options.LunarLimb)
	}
	footprints, stepDays, effectiveBoundaryPoints := solver.partialFootprints(
		window.PartialBeginOnEarth,
		window.PartialEndOnEarth,
		result.GreatestEclipse,
		options,
	)
	footprintsResult.StepDays = stepDays
	footprintsResult.Footprints = footprints
	if effectiveBoundaryPoints > footprintsResult.BoundaryPoints {
		footprintsResult.BoundaryPoints = effectiveBoundaryPoints
	}
	if options.CentralShadowStepDays > 0 &&
		footprintsResult.U1.JDE != 0 && footprintsResult.U4.JDE != 0 {
		var effectiveShadowBoundaryPoints int
		footprintsResult.CentralShadowFootprints, footprintsResult.CentralShadowStepDays, effectiveShadowBoundaryPoints = solver.shadowFootprints(
			footprintsResult.U1.JDE,
			footprintsResult.U4.JDE,
			result.GreatestEclipse,
			options.CentralShadowStepDays,
			options.BoundaryPoints,
			solarEclipseCentralShadow,
		)
		if effectiveShadowBoundaryPoints > footprintsResult.BoundaryPoints {
			footprintsResult.BoundaryPoints = effectiveShadowBoundaryPoints
		}
	}
	footprintsResult.MagnitudeContours = solver.magnitudeContours(
		result.PartialBeginOnEarth,
		result.PartialEndOnEarth,
		result.CentralBeginOnEarth,
		result.CentralEndOnEarth,
		result.GreatestEclipse,
		options,
		result.Magnitude,
		result.Type == SolarEclipseHybrid,
		totalMagnitudeOneSegments,
	)
	greatestTimeOptions := options
	if len(greatestTimeOptions.GreatestTimeValues) == 0 && greatestTimeOptions.GreatestTimeStep > 0 {
		levels := greatestTimeContourAlignedLevels(
			window.PartialBeginOnEarth, window.PartialEndOnEarth,
			greatestTimeOptions.GreatestTimeStep, greatestTimeContourMaxLevels,
		)
		greatestTimeOptions.GreatestTimeValues = make([]float64, 0, len(levels))
		for _, level := range levels {
			greatestTimeOptions.GreatestTimeValues = append(greatestTimeOptions.GreatestTimeValues, level.tt)
		}
	}
	footprintsResult.GreatestTimeContours = solver.greatestTimeContours(
		window.PartialBeginOnEarth,
		window.PartialEndOnEarth,
		greatestTimeOptions,
	)
	if options.LunarLimb != nil {
		applyFootprintsGlobalLimb(&footprintsResult, global)
	}
	return footprintsResult
}

// applyFootprintsGlobalLimb 让足迹里的全球结果与 P1/P4 和全球日食接口同源：换上月缘修正后的偏食始末、
// 食甚点中心食时长与接触点位置；偏食足迹与食甚等时线须已按同一时间窗采样。
func applyFootprintsGlobalLimb(footprintsResult *SolarEclipsePartialFootprintsResult, global SolarEclipseResult) {
	footprintsResult.Eclipse = global
	if !global.LunarLimbApplied {
		return
	}
	// 时刻与位置必须同源：月缘解给出接触点时一并替换，否则保留整套圆月面接触点。
	if footprintsResult.P1.JDE != 0 && global.PartialBeginContact.Defined {
		footprintsResult.P1.JDE = global.PartialBeginOnEarth
		footprintsResult.P1.Longitude = global.PartialBeginContact.Longitude
		footprintsResult.P1.Latitude = global.PartialBeginContact.Latitude
	}
	if footprintsResult.P4.JDE != 0 && global.PartialEndContact.Defined {
		footprintsResult.P4.JDE = global.PartialEndOnEarth
		footprintsResult.P4.Longitude = global.PartialEndContact.Longitude
		footprintsResult.P4.Latitude = global.PartialEndContact.Latitude
	}
}

func (solver solarEclipseSolver) centralBandWithResult(
	result SolarEclipseResult,
	options SolarEclipsePartialFootprintOptions,
) (SolarEclipsePartialFootprintsResult, [][]SolarEclipsePathPoint) {
	footprintsResult := SolarEclipsePartialFootprintsResult{
		Eclipse:               result,
		StepDays:              options.StepDays,
		BoundaryPoints:        options.BoundaryPoints,
		CentralShadowStepDays: options.CentralShadowStepDays,
	}
	if !result.HasPartial {
		return footprintsResult, nil
	}

	var totalMagnitudeOneSegments [][]SolarEclipsePathPoint
	var nonCentralRiseSetCurves []SolarEclipseRiseSetCurve
	nonCentralRiseSetComplete := true
	if solver.attachCentralBandContacts(&footprintsResult, result) {
		solver.attachCentralBandFootprints(&footprintsResult, result)
		if result.Centrality == SolarEclipseNonCentral {
			totalMagnitudeOneSegments, nonCentralRiseSetCurves, nonCentralRiseSetComplete =
				solver.nonCentralCentralBandEnvelope(result, options, &footprintsResult)
		}
	}
	if result.Type == SolarEclipseTotal && result.Centrality == SolarEclipseCentralTwoLimits {
		totalMagnitudeOneSegments = solver.magnitudeContourSegmentsWithSpacing(
			result.PartialBeginOnEarth, result.PartialEndOnEarth,
			result.CentralBeginOnEarth, result.CentralEndOnEarth,
			result.GreatestEclipse, 1, options.StepDays, false,
			solarEclipseTotalEnvelopeTargetSpacingKM,
		)
	}
	// The rise/set curves are a presentation option, but the precise central
	// envelope still needs their roots to close grazing annular/total paths.
	// Compute the small internal set unconditionally and only suppress it from
	// the returned result when requested.
	riseSetCurves, riseSetComplete := solver.centralBandRiseSetCurves(
		result, options, nonCentralRiseSetCurves, nonCentralRiseSetComplete,
	)
	if !options.DisableRiseSetCurves {
		footprintsResult.RiseSetCurves = riseSetCurves
		// 请求了升落曲线但拓扑校验未通过：返回的是被截断的降级结果，必须显式标记。
		footprintsResult.RiseSetTopologyDegraded = len(riseSetCurves) > 0 && !riseSetComplete
	}
	solver.attachCentralBandClosureEnvelope(
		result, &footprintsResult, riseSetCurves, totalMagnitudeOneSegments,
	)
	solver.attachCentralBandFallbackSweep(result, &footprintsResult, riseSetCurves)
	return footprintsResult, totalMagnitudeOneSegments
}

// attachCentralBandContacts 求半影与本影/反本影的外切/内切接触点，并报告本影接触对能否
// 支撑中心带采样：本影不存在或退化的抛物接触会让整个中心带支路没有意义。
func (solver solarEclipseSolver) attachCentralBandContacts(
	footprintsResult *SolarEclipsePartialFootprintsResult,
	result SolarEclipseResult,
) bool {
	contactSolver := solver
	contactSolver.exactCentralContact = result.Centrality == SolarEclipseNonCentral
	footprintsResult.P1, footprintsResult.P4, _ = contactSolver.shadowContactPair(
		result.GreatestEclipse, solarEclipsePenumbralShadow, false,
	)
	footprintsResult.P2, footprintsResult.P3, _ = contactSolver.shadowContactPair(
		result.GreatestEclipse, solarEclipsePenumbralShadow, true,
	)
	if result.Type != SolarEclipsePartial {
		footprintsResult.U1, footprintsResult.U4, _ = contactSolver.shadowContactPair(
			result.GreatestEclipse, solarEclipseCentralShadow, false,
		)
		footprintsResult.U2, footprintsResult.U3, _ = contactSolver.shadowContactPair(
			result.GreatestEclipse, solarEclipseCentralShadow, true,
		)
		return footprintsResult.U1.JDE != 0 && footprintsResult.U4.JDE != 0
	}
	return false
}

// attachCentralBandFootprints 采样本影/反本影端帽足迹，非中心食的时长决定用哪一档步长。
func (solver solarEclipseSolver) attachCentralBandFootprints(
	footprintsResult *SolarEclipsePartialFootprintsResult,
	result SolarEclipseResult,
) {
	bandStepDays := solarEclipseCentralBandStepDays
	if result.Centrality == SolarEclipseNonCentral {
		centralDuration := footprintsResult.U4.JDE - footprintsResult.U1.JDE
		if centralDuration > solarEclipseNonCentralBandPreciseMaxDurationDays {
			bandStepDays = solarEclipseNonCentralBandFallbackStepDays
		}
	}
	footprintsResult.CentralBandFootprints, footprintsResult.CentralBandStepDays = solver.centralBandFootprintsWithStep(
		footprintsResult.U1.JDE,
		footprintsResult.U2.JDE,
		footprintsResult.U3.JDE,
		footprintsResult.U4.JDE,
		result.GreatestEclipse,
		bandStepDays,
	)
}

// nonCentralCentralBandEnvelope 求解轴不入地的中心食带：先试解析包络，覆盖不足时退回采样
// 瞬时足迹并集；同时返回该支路使用的升落曲线与 magnitude-one 等值线。
func (solver solarEclipseSolver) nonCentralCentralBandEnvelope(
	result SolarEclipseResult,
	options SolarEclipsePartialFootprintOptions,
	footprintsResult *SolarEclipsePartialFootprintsResult,
) ([][]SolarEclipsePathPoint, []SolarEclipseRiseSetCurve, bool) {
	var totalMagnitudeOneSegments [][]SolarEclipsePathPoint
	var nonCentralRiseSetCurves []SolarEclipseRiseSetCurve
	nonCentralRiseSetComplete := true
	if result.Centrality == SolarEclipseNonCentral {
		riseSetStepDays := options.RiseSetStepDays
		if riseSetStepDays <= 0 || math.IsNaN(riseSetStepDays) || math.IsInf(riseSetStepDays, 0) {
			riseSetStepDays = solarEclipseRiseSetDefaultStepDays
		}
		nonCentralRiseSetCurves, nonCentralRiseSetComplete = solver.riseSetCurvesWithStatus(
			result.PartialBeginOnEarth,
			result.PartialEndOnEarth,
			result.GreatestEclipse,
			riseSetStepDays,
		)
		var bandHorizon []SolarEclipsePathPoint
		if result.Type == SolarEclipseTotal {
			totalMagnitudeOneSegments = solver.magnitudeContourSegmentsWithSpacing(
				result.PartialBeginOnEarth,
				result.PartialEndOnEarth,
				result.CentralBeginOnEarth,
				result.CentralEndOnEarth,
				result.GreatestEclipse,
				1,
				options.StepDays,
				false,
				solarEclipseNonCentralTotalBandTargetSpacingKM,
			)
			footprintsResult.CentralBandSegments, bandHorizon = solver.nonCentralTotalBandPolygons(
				totalMagnitudeOneSegments,
				footprintsResult.U1,
				footprintsResult.U4,
				result.GreatestEclipse,
				nonCentralRiseSetCurves,
			)
		}
		if !solarEclipseNonCentralBandContainsFootprints(
			footprintsResult.CentralBandSegments, footprintsResult.CentralBandFootprints,
		) {
			footprintsResult.CentralBandSegments = nil
			bandHorizon = nil
		}
		centralDuration := footprintsResult.U4.JDE - footprintsResult.U1.JDE
		if len(footprintsResult.CentralBandSegments) == 0 &&
			centralDuration <= solarEclipseNonCentralBandPreciseMaxDurationDays {
			sweep, sweepHorizon, sampled := solver.centralBandSweepPolygons(
				footprintsResult.U1.JDE,
				footprintsResult.U4.JDE,
				result.GreatestEclipse,
				nonCentralRiseSetCurves,
			)
			if centralBandSweepCoversFootprints(
				sweep, footprintsResult.CentralBandFootprints, sampled,
			) {
				footprintsResult.CentralBandSegments, bandHorizon = sweep, sweepHorizon
			}
		}
		if !centralBandSweepCoversFootprints(
			footprintsResult.CentralBandSegments, footprintsResult.CentralBandFootprints, false,
		) {
			footprintsResult.CentralBandSegments = nil
			bandHorizon = nil
		}
		alignNonCentralBandHorizon(nonCentralRiseSetCurves, bandHorizon)
	}
	return totalMagnitudeOneSegments, nonCentralRiseSetCurves, nonCentralRiseSetComplete
}

// centralBandRiseSetCurves 求解中心带闭合需要的升落曲线；DisableRiseSetCurves 只影响是否
// 对外返回，两限中心食与非中心全食的端帽闭合始终需要它们。
func (solver solarEclipseSolver) centralBandRiseSetCurves(
	result SolarEclipseResult,
	options SolarEclipsePartialFootprintOptions,
	nonCentralRiseSetCurves []SolarEclipseRiseSetCurve,
	nonCentralRiseSetComplete bool,
) ([]SolarEclipseRiseSetCurve, bool) {
	needRiseSetInternals := !options.DisableRiseSetCurves
	if result.Centrality == SolarEclipseCentralTwoLimits && result.Type != SolarEclipsePartial {
		needRiseSetInternals = true
	}
	if result.Centrality == SolarEclipseNonCentral && result.Type == SolarEclipseTotal {
		needRiseSetInternals = true
	}
	if !needRiseSetInternals {
		return nil, true
	}
	var riseSetCurves []SolarEclipseRiseSetCurve
	riseSetComplete := true
	if len(nonCentralRiseSetCurves) > 0 {
		riseSetCurves, riseSetComplete = nonCentralRiseSetCurves, nonCentralRiseSetComplete
	} else {
		riseSetStepDays := options.RiseSetStepDays
		if riseSetStepDays <= 0 || math.IsNaN(riseSetStepDays) || math.IsInf(riseSetStepDays, 0) {
			riseSetStepDays = solarEclipseRiseSetDefaultStepDays
		}
		riseSetCurves, riseSetComplete = solver.riseSetCurvesWithStatus(
			result.PartialBeginOnEarth,
			result.PartialEndOnEarth,
			result.GreatestEclipse,
			riseSetStepDays,
		)
	}
	if result.HasCentral {
		solver.attachCentralAxisContactsToRiseSetCurves(
			riseSetCurves,
			result.CentralBeginOnEarth,
			result.CentralEndOnEarth,
		)
	}
	return riseSetCurves, riseSetComplete
}

// attachCentralBandClosureEnvelope 用两端的地平闭合弧构造解析中心带；没有得到完整的两条
// 闭合弧时不改动结果，交给采样兜底。
func (solver solarEclipseSolver) attachCentralBandClosureEnvelope(
	result SolarEclipseResult,
	footprintsResult *SolarEclipsePartialFootprintsResult,
	riseSetCurves []SolarEclipseRiseSetCurve,
	totalMagnitudeOneSegments [][]SolarEclipsePathPoint,
) {
	var totalEnvelopeSegments [][]SolarEclipsePathPoint
	if (result.Type == SolarEclipseAnnular || result.Type == SolarEclipseTotal || result.Type == SolarEclipseHybrid) &&
		result.Centrality == SolarEclipseCentralTwoLimits &&
		(len(riseSetCurves) > 0 || result.Type == SolarEclipseTotal) {
		closures := make([][]SolarEclipsePathPoint, 0, 2)
		for _, side := range solarEclipseCentralBandClosureSides(*footprintsResult, result) {
			closure, ok := solver.centralBandHorizonClosureForSide(
				result, *footprintsResult, riseSetCurves, totalMagnitudeOneSegments, side,
			)
			if !ok {
				return
			}
			closures = append(closures, closure)
		}
		if len(closures) == 2 {
			footprintsResult.CentralBandHorizonClosures = closures
			solver.alignSolarEclipseCentralBandHorizonClosures(
				riseSetCurves,
				footprintsResult.CentralBandHorizonClosures,
			)
			if result.Type == SolarEclipseAnnular {
				// The continuous critical envelope is the authoritative static
				// central band. Instantaneous low-altitude footprint edges are
				// open slices and must not veto this closed envelope: requiring
				// every sampled slice to lie inside it recreates the comb-shaped
				// union that this path replaces.
				footprintsResult.CentralBandSegments = solver.centralTwoLimitBandEnvelope(
					closures, result.GreatestEclipse,
				)
				if len(footprintsResult.CentralBandSegments) == 0 {
					// At an extremely shallow polar annular event the analytic
					// continuation can start exactly on the horizon and have no
					// numerically visible first step. The sampled central sweep
					// still carries the same contact and horizon topology.
					solver.attachCentralBandSweep(
						footprintsResult,
						footprintsResult.U1.JDE, footprintsResult.U4.JDE,
						result.GreatestEclipse, riseSetCurves,
					)
				}
			} else if result.Type == SolarEclipseHybrid {
				footprintsResult.CentralBandSegments = solver.hybridCentralBandEnvelope(closures, result)
			} else {
				// Share the signed physical branches with magnitude-one output;
				// independently traced scalar contours may stop before the horizon.
				// Keep the exact magnitude-one contours when they form a valid
				// closed band. Only replace a scalar branch when its own sampled
				// shadow slices prove that it terminates early or misses the band.
				totalEnvelopeSegments = totalMagnitudeOneSegments
				if len(solarEclipseTotalBandEnvelope(totalEnvelopeSegments, closures)) == 0 {
					// Rebuild the closures from the same signed roots as the
					// vector branches. Scalar magnitude contours may have valid
					// samples but terminate on different horizon branches.
					vectorClosures := make([][]SolarEclipsePathPoint, 0, 2)
					for _, side := range []struct {
						axis, first, last, direction float64
					}{
						{result.CentralBeginOnEarth, footprintsResult.U1.JDE, footprintsResult.U2.JDE, 1},
						{result.CentralEndOnEarth, footprintsResult.U4.JDE, footprintsResult.U3.JDE, -1},
					} {
						first, last, rootsOK := solver.centralBandVectorHorizonRoots(side.axis, side.direction, side.first, side.last)
						if !rootsOK {
							vectorClosures = nil
							break
						}
						_, key, classified := solver.magnitudeEvaluationAt(first.JDE).classify(first.Longitude, first.Latitude, true)
						if !classified {
							vectorClosures = nil
							break
						}
						vectorClosures = append(vectorClosures, solver.centralBandHorizonClosure(first, last, key.direction, riseSetCurves))
					}
					if len(vectorClosures) == 2 {
						closures = vectorClosures
						footprintsResult.CentralBandHorizonClosures = closures
						totalEnvelopeSegments = solver.centralBandVectorBranches(closures, result.GreatestEclipse)
					}
				}
				footprintsResult.CentralBandSegments = solarEclipseTotalBandEnvelope(totalEnvelopeSegments, closures)
				if len(footprintsResult.CentralBandSegments) == 0 && len(riseSetCurves) > 0 {
					// Very shallow polar totalities may expose only one horizon
					// endpoint on each magnitude-one branch. The analytic join then
					// has no distinct start/end caps; use the sampled central sweep
					// as a topology-preserving fallback.
					solver.attachCentralBandSweep(
						footprintsResult,
						footprintsResult.U1.JDE, footprintsResult.U4.JDE,
						result.GreatestEclipse, riseSetCurves,
					)
				}
			}
		}
	}
}

// attachCentralBandSweep 用扫掠兜底填充分段：解析扫掠与采样并集都可能返回，采样标记
// 必须跟着返回值走，否则导出层会把采样并集当成精确解析包络处理。
func (solver solarEclipseSolver) attachCentralBandSweep(
	footprintsResult *SolarEclipsePartialFootprintsResult,
	startJDE, endJDE, greatestJDE float64,
	riseSetCurves []SolarEclipseRiseSetCurve,
) {
	segments, _, sampled := solver.centralBandSweepPolygons(startJDE, endJDE, greatestJDE, riseSetCurves)
	footprintsResult.CentralBandSegments = segments
	footprintsResult.CentralBandSampled = sampled
}

// attachCentralBandFallbackSweep 是解析包络完全缺席时的最后兜底：只要采样重建覆盖了瞬时足迹
// 就采用它，并如实标记这是采样包络。
func (solver solarEclipseSolver) attachCentralBandFallbackSweep(
	result SolarEclipseResult,
	footprintsResult *SolarEclipsePartialFootprintsResult,
	riseSetCurves []SolarEclipseRiseSetCurve,
) {
	centrality := footprintsResult.Eclipse.Centrality
	if len(footprintsResult.CentralBandSegments) == 0 && len(riseSetCurves) > 0 &&
		footprintsResult.U1.JDE != 0 && footprintsResult.U4.JDE != 0 &&
		(centrality == SolarEclipseCentralTwoLimits ||
			centrality == SolarEclipseCentralOneLimit && result.Type == SolarEclipseAnnular) {
		sweep, _, sweepSampled, covers := solver.centralBandBestSweep(
			footprintsResult.U1.JDE, footprintsResult.U4.JDE,
			result.GreatestEclipse, riseSetCurves,
			footprintsResult.CentralBandFootprints,
		)
		if covers {
			footprintsResult.CentralBandSegments = sweep
			footprintsResult.CentralBandSampled = sweepSampled
		}
	}
}

type solarEclipseCentralBandClosureSide struct {
	axisContactJDE   float64
	shadowContactJDE float64
	innerContactJDE  float64
	direction        float64
}

func solarEclipseCentralBandClosureSides(
	footprints SolarEclipsePartialFootprintsResult,
	result SolarEclipseResult,
) []solarEclipseCentralBandClosureSide {
	innerStartJDE := result.CentralBeginOnEarth
	if footprints.U2.JDE != 0 {
		innerStartJDE = footprints.U2.JDE
	}
	innerEndJDE := result.CentralEndOnEarth
	if footprints.U3.JDE != 0 {
		innerEndJDE = footprints.U3.JDE
	}
	return []solarEclipseCentralBandClosureSide{
		{result.CentralBeginOnEarth, footprints.U1.JDE, innerStartJDE, 1},
		{result.CentralEndOnEarth, footprints.U4.JDE, innerEndJDE, -1},
	}
}

// centralBandHorizonClosureForSide 解出中心带一端的地平闭合弧；三种食型各有一条取根链，
// 取不到唯一一对根就放弃该端（整条解析包络随之放弃）。
// 扫描快路径让 4862-09-28 从采样带改判为解析带：闭包是几何真解，且解析带已通过采样足迹包含审计。
func (solver solarEclipseSolver) centralBandHorizonClosureForSide(
	result SolarEclipseResult,
	footprintsResult SolarEclipsePartialFootprintsResult,
	riseSetCurves []SolarEclipseRiseSetCurve,
	totalMagnitudeOneSegments [][]SolarEclipsePathPoint,
	side solarEclipseCentralBandClosureSide,
) ([]SolarEclipsePathPoint, bool) {
	var firstRoot, lastRoot SolarEclipsePathPoint
	var ok bool
	seededDirection := RiseSetDirection("")
	if result.Type == SolarEclipseTotal {
		firstRoot, lastRoot, ok = magnitudeOneHorizonRoots(totalMagnitudeOneSegments, side.direction)
		if !ok {
			firstRoot, lastRoot, ok = solver.centralBandVectorHorizonRoots(
				side.axisContactJDE, side.direction, side.shadowContactJDE, side.innerContactJDE,
			)
		}
	} else if result.Type == SolarEclipseHybrid {
		firstRoot, lastRoot, ok = solver.centralBandVectorHorizonRoots(
			side.axisContactJDE, side.direction, side.shadowContactJDE, side.innerContactJDE,
		)
	} else if scanned := solver.centralLimitHorizonRootsByScan(
		side.shadowContactJDE, side.innerContactJDE,
	); len(scanned) == 2 {
		// 扫描式枚举直接从闭包条件解出这一对根：它不依赖种子，因此采样分支恰好终止在
		// 根上的掠地事件（1136-06-01）也不会漏根，代价是常数次星历求值。三种子牛顿链
		// 只在扫描凑不齐一对时兜底——那种情形（1552-07-21）本来就该退回采样带。
		firstRoot, lastRoot, ok = scanned[0], scanned[1], true
		seededDirection = solarEclipseNearestGreatestDirection(riseSetCurves, firstRoot)
		if seededDirection == "" {
			seededDirection = solarEclipseNearestGreatestDirection(riseSetCurves, lastRoot)
		}
	} else {
		// A grazing closure arc routinely splits its two endpoints
		// between the seeding paths: one sits outside the sampled
		// horizon branches, the other outside the axis-adjacent Newton
		// basin. Collect what every solver can offer and accept the pair
		// only when exactly two distinct roots come out.
		candidates := make([]SolarEclipsePathPoint, 0, 4)
		directions := make([]RiseSetDirection, 0, 4)
		curveDirection := RiseSetDirection("")
		unionDirection := RiseSetDirection("")
		appendRoots := func(
			first, last SolarEclipsePathPoint,
			direction RiseSetDirection,
			found bool,
		) {
			if !found {
				return
			}
			for _, root := range []SolarEclipsePathPoint{first, last} {
				if root.JDE == 0 || solarEclipseRiseSetPointExists(candidates, root) {
					continue
				}
				candidates = append(candidates, root)
				directions = append(directions, direction)
			}
		}
		firstRoot, lastRoot, ok = solver.centralLimitHorizonRootsNearAxisContact(
			side.axisContactJDE, side.shadowContactJDE, side.innerContactJDE, side.direction,
		)
		appendRoots(firstRoot, lastRoot, "", ok)
		curveRoots, curveDirection := solver.centralLimitHorizonRootsFromCurves(
			riseSetCurves, side.shadowContactJDE, side.innerContactJDE,
		)
		for _, root := range curveRoots {
			appendRoots(root, SolarEclipsePathPoint{}, curveDirection, true)
		}
		if len(candidates) < 2 {
			// The sampled sweep is the most expensive seeding, so it
			// only runs while the pair is still incomplete. Some grazing
			// events have no two-root closure at all, and then the
			// sampled band is the right representation rather than a
			// fallback: 4862-09-28 ends 40 km inside the horizon (the
			// cap is bounded by the umbral rim), and 1552-07-21 has its
			// boundary running along the horizon (measured +0.004 then
			// -0.000 degrees), so its closure arc is degenerate.
			unionRoots, unionDirection := solver.centralLimitHorizonRootsFromSampledBoundary(
				side.shadowContactJDE, side.innerContactJDE, result.GreatestEclipse,
				footprintsResult.CentralBandFootprints, riseSetCurves,
			)
			for _, root := range unionRoots {
				appendRoots(root, SolarEclipsePathPoint{}, unionDirection, true)
			}
		}
		if seededDirection == "" {
			seededDirection = curveDirection
		}
		if seededDirection == "" {
			seededDirection = unionDirection
		}
		ok = len(candidates) == 2
		if ok {
			sort.Slice(candidates, func(first, second int) bool {
				return candidates[first].JDE < candidates[second].JDE
			})
			firstRoot, lastRoot = candidates[0], candidates[1]
			for _, direction := range directions {
				if direction != "" {
					seededDirection = direction
					break
				}
			}
		}
	}
	if !ok {
		return nil, false
	}
	// Near a pole, the first and last contacts can both be sunrise or
	// both sunset. Classify the solved root instead of the temporal end.
	_, key, classified := solver.magnitudeEvaluationAt(firstRoot.JDE).classify(
		firstRoot.Longitude, firstRoot.Latitude, true,
	)
	if !classified && seededDirection == "" {
		seededDirection = solarEclipseNearestGreatestDirection(riseSetCurves, firstRoot)
	}
	if !classified && seededDirection != "" {
		// The local classification needs the eclipse to be exactly at
		// its greatest there, which a shallow polar closure root does
		// not always satisfy. The horizon curve that seeded the root
		// already knows whether the Sun is rising or setting.
		key = solarEclipseRiseSetCurveKey{
			phase: RiseSetPhaseGreatest, direction: seededDirection,
		}
	}
	if !classified {
		return nil, false
	}
	return solver.centralBandHorizonClosure(
		firstRoot, lastRoot, key.direction, riseSetCurves,
	), true
}

// centralBandBestSweep returns the sweep candidate that contains the sampled
// umbra. The analytic sweep is smooth and stays the first choice whenever it
// really contains it; a grazing path whose envelope terminates early is rebuilt
// from the sampled footprints instead.
// The returned flag reports whether the candidate already passed the coverage
// audit, so the caller does not repeat the probe.
func (solver solarEclipseSolver) centralBandBestSweep(
	startJDE, endJDE, greatestJDE float64,
	riseSetCurves []SolarEclipseRiseSetCurve,
	footprints []SolarEclipsePartialFootprint,
) ([][]SolarEclipsePathPoint, []SolarEclipsePathPoint, bool, bool) {
	candidate, horizon, sampled := solver.centralBandSweepPolygonsReusing(
		startJDE, endJDE, greatestJDE, riseSetCurves, footprints,
	)
	if solarEclipseBandContainsSampledFootprintsWithinKM(
		candidate, footprints, centralBandSweepToleranceKM(sampled),
	) {
		return candidate, horizon, sampled, true
	}
	// The sampled reconstruction is only worth building when the sweep had an
	// analytic region to reject; when the sweep itself already fell back to the
	// sampled union, the audit above has judged that exact reconstruction.
	if !sampled {
		union := solver.centralBandSampledFootprintUnionOverRange(startJDE, endJDE, greatestJDE, footprints)
		if len(union) > 0 && solarEclipseBandContainsSampledFootprintsWithinKM(
			union, footprints, centralBandSweepToleranceKM(true),
		) {
			return union, nil, true, true
		}
	}
	return candidate, horizon, sampled, false
}

// centralBandSweepToleranceKM returns the coverage bound of a sweep candidate.
func centralBandSweepToleranceKM(sampled bool) float64 {
	if sampled {
		return solarEclipseCentralBandUnionContainmentToleranceKM
	}
	return solarEclipseNonCentralBandContainmentToleranceKM
}

// centralBandSweepCoversFootprints validates a sweep candidate. The analytic
// sweep must contain every sampled footprint within the usual footprint
// tolerance; the sampled footprint union is decimated for tractability and is
// validated against the wider union bound instead.
func centralBandSweepCoversFootprints(
	segments [][]SolarEclipsePathPoint,
	footprints []SolarEclipsePartialFootprint,
	sampled bool,
) bool {
	tolerance := solarEclipseNonCentralBandContainmentToleranceKM
	if sampled {
		tolerance = solarEclipseCentralBandUnionContainmentToleranceKM
	}
	return solarEclipseBandContainsFootprintsWithinKM(segments, footprints, tolerance)
}

// setCenterLinePairWidths replaces the analytic 2r/sin(altitude) width with the
// distance between the paired limits, which is the width the tables publish.
// The analytic form is a small-angle approximation that diverges towards the
// horizon (2017-08-21 reports 10,062 km near its ends instead of ~90 km), so it
// is only kept as the greatest-eclipse figure that the catalogues quote. A
// sample whose cross-section does not close has no width at all, and a one-limit
// event has none anywhere - which is why catalogues print no width for it (NASA
// lists 1874-10-10 with "-").
func (solver solarEclipseSolver) setCenterLinePairWidths(
	centerLine []SolarEclipsePathPoint,
	centrality SolarEclipseCentrality,
	northern, southern []SolarEclipsePathPoint,
	paired []bool,
) {
	if len(centerLine) == 0 {
		return
	}
	if centrality != SolarEclipseCentralTwoLimits {
		for index := range centerLine {
			centerLine[index].WidthKM = 0
		}
		return
	}
	if len(northern) != len(centerLine) || len(southern) != len(centerLine) || len(paired) != len(centerLine) {
		return
	}
	for index := range centerLine {
		if !paired[index] {
			centerLine[index].WidthKM = 0
			continue
		}
		width := solarEclipsePathDistanceKM(northern[index], southern[index])
		if width <= 0 || width > solarEclipsePathMaxPossibleWidthKM {
			// A cross-section wider than the Earth cannot be a footprint edge.
			width = 0
		}
		centerLine[index].WidthKM = width
	}
}

// centralPathMaxCentralDuration returns the longest central phase on the center
// line and where it happens. The local solution reuses the event-local
// interpolated ephemeris, which is what keeps a per-sample scan affordable.
// Catalogues publish the duration at greatest eclipse,
// which for a shallow event is not the longest one: 2020-06-21 lasts 38 s there
// but 1m22s near the sunrise end of its track. Samples below the horizon are
// skipped because the central phase is not observable from them; the width is
// deliberately not used as a filter, because a one-limit event has no width at
// all yet still has a published duration.
func (solver solarEclipseSolver) centralPathMaxCentralDuration(
	centerLine []SolarEclipsePathPoint,
) (float64, float64, float64) {
	bestDays, bestLongitude, bestLatitude := 0.0, 0.0, 0.0
	// 局部星历只建一次；逐点各建一次会让整段扫描的星历开销翻上百倍。
	solver = solver.withLocalEphemeris()
	for _, point := range centerLine {
		if point.SunAltitude < -solarEclipsePathDurationHorizonToleranceDegrees {
			continue
		}
		span := solver.centralPhaseDurationDaysAt(point.JDE, point.Longitude, point.Latitude)
		if span <= 0 {
			continue
		}
		if span > bestDays {
			bestDays, bestLongitude, bestLatitude = span, point.Longitude, point.Latitude
		}
	}
	return bestDays, bestLongitude, bestLatitude
}

func (solver solarEclipseSolver) attachCentralAxisContactPoints(
	points []SolarEclipsePathPoint,
	startJDE, endJDE float64,
) []SolarEclipsePathPoint {
	if len(points) == 0 {
		return points
	}
	const supportOffsetDays = 0.1 / 86400.0
	contacts := make([]SolarEclipsePathPoint, 0, 2)
	for _, jd := range []float64{startJDE, endJDE} {
		if point, ok := solver.centralAxisContactPointAt(jd); ok {
			contacts = append(contacts, point)
		}
	}
	if len(contacts) != 2 {
		return points
	}
	supports := make([]SolarEclipsePathPoint, 0, 2)
	if endJDE-startJDE > 2*supportOffsetDays {
		if point, ok := solver.centralPathPointAt(startJDE + supportOffsetDays); ok {
			supports = append(supports, point)
		}
		if point, ok := solver.centralPathPointAt(endJDE - supportOffsetDays); ok {
			supports = append(supports, point)
		}
	}
	filtered := points[:0]
	for _, point := range points {
		if math.Abs(point.JDE-startJDE) <= solarEclipsePathDuplicateTimeDays ||
			math.Abs(point.JDE-endJDE) <= solarEclipsePathDuplicateTimeDays {
			continue
		}
		filtered = append(filtered, point)
	}
	points = append(filtered, supports...)
	points = append(points, contacts...)
	sort.Slice(points, func(first, second int) bool { return points[first].JDE < points[second].JDE })
	return normalizeSolarEclipsePathPointSeries(points)
}

func setSolarEclipseCentralContactWidths(
	centerLine, northern, southern []SolarEclipsePathPoint,
) {
	if len(centerLine) < 2 || len(northern) != len(southern) {
		return
	}
	for _, centerIndex := range []int{0, len(centerLine) - 1} {
		bestIndex := -1
		bestDelta := math.Inf(1)
		for index := range northern {
			delta := math.Abs(northern[index].JDE - centerLine[centerIndex].JDE)
			if delta < bestDelta {
				bestIndex, bestDelta = index, delta
			}
		}
		if bestIndex >= 0 && bestDelta <= 0.2/86400.0 {
			centerLine[centerIndex].WidthKM = solarEclipsePathDistanceKM(
				northern[bestIndex], southern[bestIndex],
			)
		}
	}
}

func (solver solarEclipseSolver) attachCentralAxisContactsToRiseSetCurves(
	curves []SolarEclipseRiseSetCurve,
	startJDE, endJDE float64,
) {
	for _, endpoint := range []struct {
		jde       float64
		direction RiseSetDirection
	}{
		{startJDE, RiseSetDirectionRise},
		{endJDE, RiseSetDirectionSet},
	} {
		contact, ok := solver.centralAxisContactPointAt(endpoint.jde)
		if !ok {
			continue
		}
		insertSolarEclipseRiseSetContact(curves, contact, endpoint.direction)
	}
}

func insertSolarEclipseRiseSetContact(
	curves []SolarEclipseRiseSetCurve,
	contact SolarEclipsePathPoint,
	direction RiseSetDirection,
) {
	bestCurve, bestSegment, bestPoint := -1, -1, -1
	bestDistance := math.Inf(1)
	for curveIndex := range curves {
		curve := &curves[curveIndex]
		if curve.Phase != RiseSetPhaseGreatest || curve.Direction != direction {
			continue
		}
		for segmentIndex, segment := range curve.Segments {
			if len(segment) < 2 ||
				contact.JDE < segment[0].JDE-solarEclipseRiseSetTimeEpsilonDays ||
				contact.JDE > segment[len(segment)-1].JDE+solarEclipseRiseSetTimeEpsilonDays {
				continue
			}
			pointIndex := sort.Search(len(segment), func(index int) bool {
				return segment[index].JDE >= contact.JDE
			})
			if pointIndex == 0 {
				pointIndex = 1
			} else if pointIndex >= len(segment) {
				pointIndex = len(segment) - 1
			}
			first, second := segment[pointIndex-1], segment[pointIndex]
			fraction := 0.0
			if second.JDE > first.JDE {
				fraction = (contact.JDE - first.JDE) / (second.JDE - first.JDE)
			}
			fraction = math.Max(0, math.Min(1, fraction))
			candidate := solarEclipsePathSphericalInterpolate(first, second, fraction)
			distance := solarEclipsePathDistanceKM(candidate, contact)
			if distance < bestDistance {
				bestCurve, bestSegment, bestPoint = curveIndex, segmentIndex, pointIndex
				bestDistance = distance
			}
		}
	}
	if bestCurve < 0 || bestDistance > solarEclipseRiseSetTargetSpacingKM {
		return
	}
	segment := curves[bestCurve].Segments[bestSegment]
	if bestPoint > 0 && math.Abs(segment[bestPoint-1].JDE-contact.JDE) <= solarEclipseRiseSetTimeEpsilonDays {
		segment[bestPoint-1] = contact
	} else if bestPoint < len(segment) && math.Abs(segment[bestPoint].JDE-contact.JDE) <= solarEclipseRiseSetTimeEpsilonDays {
		segment[bestPoint] = contact
	} else {
		segment = append(segment, SolarEclipsePathPoint{})
		copy(segment[bestPoint+1:], segment[bestPoint:])
		segment[bestPoint] = contact
	}
	curves[bestCurve].Segments[bestSegment] = segment
}

func (solver solarEclipseSolver) centralBandFootprintsWithStep(
	outerStartJDE, innerStartJDE, innerEndJDE, outerEndJDE, greatestJDE, stepDays float64,
) ([]SolarEclipsePartialFootprint, float64) {
	if stepDays <= 0 || math.IsNaN(stepDays) || math.IsInf(stepDays, 0) {
		stepDays = solarEclipseCentralBandStepDays
	}
	if innerStartJDE != 0 && innerEndJDE != 0 &&
		outerStartJDE < innerStartJDE && innerStartJDE < innerEndJDE && innerEndJDE < outerEndJDE {
		start, actualStartStep, _ := solver.shadowFootprintsWithSpacing(
			outerStartJDE, innerStartJDE, innerStartJDE, stepDays,
			solarEclipseCentralBandBoundaryPoints, solarEclipseCentralShadow,
			solarEclipseCentralBandTargetSpacingKM,
		)
		end, actualEndStep, _ := solver.shadowFootprintsWithSpacing(
			innerEndJDE, outerEndJDE, innerEndJDE, stepDays,
			solarEclipseCentralBandBoundaryPoints, solarEclipseCentralShadow,
			solarEclipseCentralBandTargetSpacingKM,
		)
		return append(start, end...), math.Max(actualStartStep, actualEndStep)
	}
	if greatestJDE <= outerStartJDE || greatestJDE >= outerEndJDE {
		greatestJDE = (outerStartJDE + outerEndJDE) / 2
	}
	footprints, actualStep, _ := solver.shadowFootprintsWithSpacing(
		outerStartJDE, outerEndJDE, greatestJDE, stepDays,
		solarEclipseCentralBandBoundaryPoints, solarEclipseCentralShadow,
		solarEclipseCentralBandTargetSpacingKM,
	)
	return footprints, actualStep
}
