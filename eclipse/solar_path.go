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
	// Greatest 是食甚点/最佳观测点, greatest eclipse point.
	Greatest SolarEclipsePathPoint
	// CenterLine 是中心线, central line.
	CenterLine []SolarEclipsePathPoint
	// NorthernLimit 是中心食带北界近似线, approximate northern limit of the central path.
	NorthernLimit []SolarEclipsePathPoint
	// SouthernLimit 是中心食带南界近似线, approximate southern limit of the central path.
	SouthernLimit []SolarEclipsePathPoint
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
		Eclipse:         solarEclipseInfoFromBasic(result.Eclipse, location),
		Greatest:        solarEclipsePathPointFromBasic(result.Greatest, location),
		CenterLine:      solarEclipsePathPointsFromBasic(result.CenterLine, location),
		NorthernLimit:   solarEclipsePathPointsFromBasic(result.NorthernLimit, location),
		SouthernLimit:   solarEclipsePathPointsFromBasic(result.SouthernLimit, location),
		Step:            solarEclipsePathStepDuration(result.StepDays),
		TargetSpacingKM: result.TargetSpacingKM,
	}
	return path, true
}

func solarEclipsePartialFootprints(
	date time.Time,
	options SolarEclipsePartialFootprintOptions,
	calculator solarEclipsePartialFootprintsCalculator,
) (SolarEclipsePartialFootprintsInfo, bool) {
	location := date.Location()
	result := calculator(solarEclipseTimeToTTJDE(date), basicSolarEclipsePartialFootprintOptions(options))
	if !result.Eclipse.HasPartial || len(result.Footprints) == 0 {
		return SolarEclipsePartialFootprintsInfo{}, false
	}

	footprints := SolarEclipsePartialFootprintsInfo{
		Eclipse:                 solarEclipseInfoFromBasic(result.Eclipse, location),
		Footprints:              solarEclipsePartialFootprintsFromBasic(result.Footprints, location),
		CentralShadowFootprints: solarEclipsePartialFootprintsFromBasic(result.CentralShadowFootprints, location),
		P1:                      solarEclipseOptionalPathPointFromBasic(result.P1, location),
		P2:                      solarEclipseOptionalPathPointFromBasic(result.P2, location),
		P3:                      solarEclipseOptionalPathPointFromBasic(result.P3, location),
		P4:                      solarEclipseOptionalPathPointFromBasic(result.P4, location),
		U1:                      solarEclipseOptionalPathPointFromBasic(result.U1, location),
		U2:                      solarEclipseOptionalPathPointFromBasic(result.U2, location),
		U3:                      solarEclipseOptionalPathPointFromBasic(result.U3, location),
		U4:                      solarEclipseOptionalPathPointFromBasic(result.U4, location),
		Step:                    solarEclipsePathStepDuration(result.StepDays),
		BoundaryPoints:          result.BoundaryPoints,
		CentralShadowStep:       solarEclipsePathStepDuration(result.CentralShadowStepDays),
	}
	return footprints, true
}

func basicSolarEclipsePathOptions(options SolarEclipsePathOptions) basic.SolarEclipsePathOptions {
	basicOptions := basic.SolarEclipsePathOptions{
		TargetSpacingKM: options.TargetSpacingKM,
	}
	if options.Step > 0 {
		basicOptions.StepDays = options.Step.Hours() / 24
	}
	return basicOptions
}

func basicSolarEclipsePartialFootprintOptions(options SolarEclipsePartialFootprintOptions) basic.SolarEclipsePartialFootprintOptions {
	basicOptions := basic.SolarEclipsePartialFootprintOptions{
		BoundaryPoints: options.BoundaryPoints,
	}
	if options.Step > 0 {
		basicOptions.StepDays = options.Step.Hours() / 24
	}
	if options.CentralShadowStep > 0 {
		basicOptions.CentralShadowStepDays = options.CentralShadowStep.Hours() / 24
	}
	return basicOptions
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
