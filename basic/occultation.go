package basic

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro/tools"
)

// ErrInvalidOccultationInput 表示月掩输入契约无效。
// ErrInvalidOccultationInput reports invalid lunar-occultation contract input.
var ErrInvalidOccultationInput = errors.New("invalid lunar occultation input")

// ErrOccultationPathSamplingLimit 表示请求的时间步长、中心线间距或有限盘面采样超过确定性的工作量或输出预算。
// ErrOccultationPathSamplingLimit reports that the requested time step, center-line spacing, or aggregate finite-disk sampling would exceed the implementation's deterministic work or output budget.
var ErrOccultationPathSamplingLimit = errors.New("lunar occultation path sampling limit exceeded")

const (
	occultationSearchMinimumStep           = 250 * time.Millisecond
	occultationPathMinimumStep             = time.Second
	occultationPathMinimumTargetSpacingKM  = 1.0
	occultationEventSelectionTolerance     = 10 * time.Millisecond
	occultationEventSelectionToleranceDays = float64(occultationEventSelectionTolerance) / float64(24*time.Hour)
	// 银河系内恒星的径向速度上限，仅用于挡掉明显填错的输入。
	starRadialVelocityLimitKmPerSecond = 1000.0
)

// CoordinateFrame 标识恒星输入坐标使用的赤道坐标系。
// CoordinateFrame identifies the equatorial coordinate frame used by an input stellar coordinate.
type CoordinateFrame string

const (
	// CoordinateFrameICRS 表示 ICRS 星表坐标系。
	// CoordinateFrameICRS is the ICRS catalog frame.
	CoordinateFrameICRS CoordinateFrame = "icrs"
	// CoordinateFrameJ2000 表示 J2000 平均赤道坐标系。
	// CoordinateFrameJ2000 is the mean equatorial J2000 frame.
	CoordinateFrameJ2000 CoordinateFrame = "j2000"
	// CoordinateFrameApparentOfDate 表示历元时刻的视赤道坐标系。
	// CoordinateFrameApparentOfDate is the apparent equatorial frame of date.
	CoordinateFrameApparentOfDate CoordinateFrame = "apparent_of_date"
)

// OccultationType 标识月掩结果的几何类型。
// OccultationType identifies the result geometry.
type OccultationType string

const (
	// OccultationTotal 表示掩甚时目标盘面被完全覆盖。
	// OccultationTotal means the target disk is fully covered at greatest occultation.
	OccultationTotal OccultationType = "total"
	// OccultationPartial 表示掩甚时有限目标盘面只有部分被覆盖。
	// OccultationPartial means only part of a finite target disk is covered at greatest occultation.
	OccultationPartial OccultationType = "partial"
	// OccultationGrazing 表示掠掩：圆月面下两边缘相切、无正持续时间的重叠；按月缘剖面求解时指星迹始终在月缘起伏带内，可有多段遮挡。
	// OccultationGrazing is a graze: with a circular limb the limbs are tangent without a positive-duration overlap; with a limb profile the track stays inside the relief band and may be hidden several times.
	OccultationGrazing OccultationType = "grazing"
	// OccultationNone 表示月缘剖面下恒星未被遮挡，只出现在精密剖面结果中。
	// OccultationNone means the star is never hidden by the limb profile; it only appears in precise profile results.
	OccultationNone OccultationType = "none"
)

// OccultationPlanet 标识有限盘面的行星目标。
// OccultationPlanet identifies a finite-disk planetary target.
type OccultationPlanet string

const (
	// OccultationMercury 表示水星有限盘面目标。
	// OccultationMercury identifies Mercury as the finite-disk target.
	OccultationMercury OccultationPlanet = "mercury"
	// OccultationVenus 表示金星有限盘面目标。
	// OccultationVenus identifies Venus as the finite-disk target.
	OccultationVenus OccultationPlanet = "venus"
	// OccultationMars 表示火星有限盘面目标。
	// OccultationMars identifies Mars as the finite-disk target.
	OccultationMars OccultationPlanet = "mars"
	// OccultationJupiter 表示木星有限盘面目标。
	// OccultationJupiter identifies Jupiter as the finite-disk target.
	OccultationJupiter OccultationPlanet = "jupiter"
	// OccultationSaturn 表示土星有限盘面目标。
	// OccultationSaturn identifies Saturn as the finite-disk target.
	OccultationSaturn OccultationPlanet = "saturn"
	// OccultationUranus 表示天王星有限盘面目标。
	// OccultationUranus identifies Uranus as the finite-disk target.
	OccultationUranus OccultationPlanet = "uranus"
	// OccultationNeptune 表示海王星有限盘面目标。
	// OccultationNeptune identifies Neptune as the finite-disk target.
	OccultationNeptune OccultationPlanet = "neptune"
)

// String 返回结果标识中使用的英文目标名称。
// String returns the English target name used in result identifiers.
func (p OccultationPlanet) String() string {
	switch p {
	case OccultationMercury:
		return "Mercury"
	case OccultationVenus:
		return "Venus"
	case OccultationMars:
		return "Mars"
	case OccultationJupiter:
		return "Jupiter"
	case OccultationSaturn:
		return "Saturn"
	case OccultationUranus:
		return "Uranus"
	case OccultationNeptune:
		return "Neptune"
	default:
		return ""
	}
}

// Validate 检查行星目标是否受支持。
// Validate checks whether the planetary target is supported.
func (p OccultationPlanet) Validate() error {
	if p.String() == "" {
		return fmt.Errorf("%w: unsupported occultation planet %q", ErrInvalidOccultationInput, p)
	}
	return nil
}

func validateOccultationTimeRange(start, end time.Time) error {
	if start.IsZero() || end.IsZero() {
		return fmt.Errorf("%w: start and end are required", ErrInvalidOccultationInput)
	}
	if !end.After(start) {
		return fmt.Errorf("%w: end must be after start", ErrInvalidOccultationInput)
	}
	return nil
}

func occultationTimeInSelectionWindow(value, start, end time.Time) bool {
	return value.Sub(start) >= -occultationEventSelectionTolerance &&
		value.Sub(end) <= occultationEventSelectionTolerance
}

// Observer 描述站心观测地点。
// Observer describes the topocentric observing site.
type Observer struct {
	// Longitude 是经度，东经为正，单位为度。
	// Longitude is east-positive, in degrees.
	Longitude float64
	// Latitude 是纬度，北纬为正，单位为度。
	// Latitude is north-positive, in degrees.
	Latitude float64
	// Height 是观测者相对平均海平面的高度，单位为米。
	// Height is the observer elevation above mean sea level, in meters.
	Height float64
}

// moonTopocentricSemidiameterN 返回指定地点看到的月球角半径。
// 通用 MoonSemidiameterN 使用地心距离；月掩接触使用同一站心视差修正后的观测者到月球距离。
// moonTopocentricSemidiameterN returns the lunar angular radius as seen from the supplied site.
// The usual MoonSemidiameterN uses geocentric distance; occultation contacts use the observer-to-Moon distance after the same topocentric parallax correction as the direction.
func moonTopocentricSemidiameterN(tt float64, observer Observer, n int, instances ...*Instance) float64 {
	inst := instanceValue(instances)

	moonRA, moonDec := HMoonGeocentricApparentRaDecNWithInstance(tt, n, inst)
	moonDistanceKM := HMoonAwayNWithInstance(tt, n, inst)
	if !finite(moonRA) || !finite(moonDec) || !finite(moonDistanceKM) || moonDistanceKM <= 0 {
		return math.NaN()
	}
	distanceKM := topocentricDistanceKM(moonRA, moonDec, moonDistanceKM, observer, TT2UTC(tt))
	if !finite(distanceKM) || distanceKM <= 0 {
		return math.NaN()
	}
	return angularSemidiameterArcsec(moonEquatorialRadiusKM, distanceKM)
}

// topocentricDistanceKM 使用与 TopocentricRaDec 相同的 WGS-84 风格站点因子计算观测者到目标的距离。
// 目标采用视赤道坐标，恒星时与 TopocentricRaDec 一样基于 UTC/UT。
// topocentricDistanceKM uses the same WGS-84-style site factors as TopocentricRaDec.
// The target uses apparent equatorial coordinates, and the sidereal angle is based on UTC/UT.
func topocentricDistanceKM(ra, dec, distanceKM float64, observer Observer, ut float64) float64 {
	return topocentricDistanceKMWithSidereal(ra, dec, distanceKM, observer, ApparentSiderealTime(UTC2UT1(ut))*15)
}

func topocentricDistanceKMWithSidereal(
	ra, dec, distanceKM float64,
	observer Observer,
	siderealDegrees float64,
) float64 {
	const astronomicalUnitKM = angularDiameterAstronomicalUnitKM

	distanceAU := distanceKM / astronomicalUnitKM
	if distanceAU <= 0 {
		return math.NaN()
	}
	raRad := ra * math.Pi / 180
	decRad := dec * math.Pi / 180
	moon := [3]float64{
		distanceAU * math.Cos(decRad) * math.Cos(raRad),
		distanceAU * math.Cos(decRad) * math.Sin(raRad),
		distanceAU * math.Sin(decRad),
	}
	theta := (siderealDegrees + observer.Longitude) * math.Pi / 180
	observerAU := topocentricEarthRadiusKM / astronomicalUnitKM
	observerVector := [3]float64{
		observerAU * pcosi(observer.Latitude, observer.Height) * math.Cos(theta),
		observerAU * pcosi(observer.Latitude, observer.Height) * math.Sin(theta),
		observerAU * psini(observer.Latitude, observer.Height),
	}
	dx := moon[0] - observerVector[0]
	dy := moon[1] - observerVector[1]
	dz := moon[2] - observerVector[2]
	return math.Sqrt(dx*dx+dy*dy+dz*dz) * astronomicalUnitKM
}

// Validate 检查站心计算所需的地理范围。
// Validate checks the geographic bounds needed by topocentric calculations.
func (o Observer) Validate() error {
	if !finite(o.Longitude) || !finite(o.Latitude) || !finite(o.Height) {
		return fmt.Errorf("%w: observer values must be finite", ErrInvalidOccultationInput)
	}
	if o.Longitude < -180 || o.Longitude > 180 {
		return fmt.Errorf("%w: observer longitude must be in [-180, 180]", ErrInvalidOccultationInput)
	}
	if o.Latitude < -90 || o.Latitude > 90 {
		return fmt.Errorf("%w: observer latitude must be in [-90, 90]", ErrInvalidOccultationInput)
	}
	return nil
}

// StarCoordinate 是调用者为恒星提供的星表坐标或视位置坐标。
// RA 和 Dec 的单位为度；ProperMotionRACosDecMasPerYear 使用星表常见的 dRA*cos(Dec) 约定，单位为毫角秒/年。
// StarCoordinate is a catalog or apparent coordinate supplied for a star.
// RA and Dec are degrees; ProperMotionRACosDecMasPerYear uses the usual catalog convention of dRA*cos(Dec), in milliarcseconds per year.
type StarCoordinate struct {
	ID string

	RA    float64
	Dec   float64
	Epoch time.Time
	Frame CoordinateFrame

	ProperMotionRACosDecMasPerYear float64
	ProperMotionDecMasPerYear      float64
	// ParallaxMas 为周年视差，单位毫角秒；0 表示未提供距离，此时退回二维自行传播。
	// ParallaxMas is the annual parallax in milliarcseconds; 0 means no distance is given and the
	// two-dimensional proper-motion path is used.
	ParallaxMas float64
	// DistanceLightYear 是 ParallaxMas 的替代输入，单位光年；仅当 ParallaxMas 为 0 时生效，0 表示未提供。
	// DistanceLightYear is an alternative to ParallaxMas in light-years; it applies only when
	// ParallaxMas is 0, and 0 means it was not given.
	DistanceLightYear float64
	// RadialVelocityKmPerSecond 单位千米/秒，0 合法，只在距离已知时参与三维空间运动。
	// RadialVelocityKmPerSecond is in kilometres per second; 0 is valid and it enters the
	// three-dimensional motion only when a distance is known.
	RadialVelocityKmPerSecond float64
}

// Validate 在构造目标前检查恒星坐标契约。
// Validate checks the coordinate contract before a target is constructed.
func (s StarCoordinate) Validate() error {
	if !finite(s.RA) || s.RA < 0 || s.RA >= 360 {
		return fmt.Errorf("%w: star RA must be in [0, 360)", ErrInvalidOccultationInput)
	}
	if !finite(s.Dec) || s.Dec < -90 || s.Dec > 90 {
		return fmt.Errorf("%w: star Dec must be in [-90, 90]", ErrInvalidOccultationInput)
	}
	if s.Epoch.IsZero() {
		return fmt.Errorf("%w: star epoch is required", ErrInvalidOccultationInput)
	}
	if !validCoordinateFrame(s.Frame) {
		return fmt.Errorf("%w: unsupported star coordinate frame %q", ErrInvalidOccultationInput, s.Frame)
	}
	if !finite(s.ProperMotionRACosDecMasPerYear) || !finite(s.ProperMotionDecMasPerYear) {
		return fmt.Errorf("%w: star proper motion must be finite", ErrInvalidOccultationInput)
	}
	if !finite(s.ParallaxMas) || s.ParallaxMas < 0 {
		return fmt.Errorf("%w: star parallax must be finite and non-negative", ErrInvalidOccultationInput)
	}
	if !finite(s.DistanceLightYear) || s.DistanceLightYear < 0 {
		return fmt.Errorf("%w: star distance in light-years must be finite and non-negative", ErrInvalidOccultationInput)
	}
	if !finite(s.RadialVelocityKmPerSecond) || math.Abs(s.RadialVelocityKmPerSecond) > starRadialVelocityLimitKmPerSecond {
		return fmt.Errorf("%w: star radial velocity must be finite and within +/-%.0f km/s", ErrInvalidOccultationInput, starRadialVelocityLimitKmPerSecond)
	}
	return nil
}

// parallaxMas 返回生效的视差：显式视差优先，否则由光年距离折算。
func (s StarCoordinate) parallaxMas() float64 {
	if s.ParallaxMas > 0 {
		return s.ParallaxMas
	}
	if s.DistanceLightYear <= 0 {
		return 0
	}
	return 1000 / tools.DistanceToParsecs(s.DistanceLightYear, tools.DistanceLightYear)
}

// StarCoordinateFromStarData 将一条内嵌星表记录转换为月掩搜索使用的 J2000 坐标契约。
// 星表自行从角秒/年转换为毫角秒/年；正的秒差距距离转换为毫角秒年视差。本函数只转换传入值，不会加载星表。
// StarCoordinateFromStarData converts one embedded-catalog entry into the J2000 coordinate contract used by lunar-occultation searches.
// The catalog's proper motions are converted from arcseconds/year to milliarcseconds/year; a positive parsec distance is converted to annual parallax in milliarcseconds.
// This function only converts the supplied value and never loads the catalog.
func StarCoordinateFromStarData(star StarData) (StarCoordinate, error) {
	if star.HR == 0 {
		return StarCoordinate{}, fmt.Errorf("%w: star catalog HR number is required", ErrInvalidOccultationInput)
	}
	if !finite(star.Pc) || star.Pc < 0 {
		return StarCoordinate{}, fmt.Errorf("%w: star distance must be finite and non-negative", ErrInvalidOccultationInput)
	}

	parallaxMas := 0.0
	if star.Pc > 0 {
		parallaxMas = 1000 / star.Pc
	}
	coordinate := StarCoordinate{
		ID:                             starCoordinateIDFromStarData(star),
		RA:                             star.Ra,
		Dec:                            star.Dec,
		Epoch:                          time.Date(2000, time.January, 1, 12, 0, 0, 0, time.UTC),
		Frame:                          CoordinateFrameJ2000,
		ProperMotionRACosDecMasPerYear: star.PmRA * 1000,
		ProperMotionDecMasPerYear:      star.PmDec * 1000,
		ParallaxMas:                    parallaxMas,
		RadialVelocityKmPerSecond:      star.RadVel,
	}
	if err := coordinate.Validate(); err != nil {
		return StarCoordinate{}, fmt.Errorf("convert star catalog coordinate: %w", err)
	}
	return coordinate, nil
}

func starCoordinateIDFromStarData(star StarData) string {
	for _, name := range []string{star.ChineseName, star.ChineseAlias, star.CommonName, star.Name} {
		if name = strings.TrimSpace(name); name != "" {
			return name
		}
	}
	if star.HR > 0 {
		return fmt.Sprintf("HR %d", star.HR)
	}
	return ""
}

// OccultationSearchOptions 控制固定目标和行星月掩搜索。
// 零值使用实现默认值；MaxEvents == 0 表示不限制数量。
// OccultationSearchOptions controls fixed-target and planetary occultation searches.
// Zero values select implementation defaults; MaxEvents == 0 means unlimited.
type OccultationSearchOptions struct {
	// MaxStep 是粗略搜索的最大步长；小于 250ms 的正值会被拒绝，因为在支持的时间范围内无法可靠地用儒略日浮点数推进。
	// MaxStep is the maximum coarse-search step. Positive values below 250 ms are rejected because they cannot be advanced reliably in Julian-day floating-point arithmetic over the supported time span.
	MaxStep time.Duration
	// SafetyMarginArcsec 是加入粗略候选和黄纬预筛的安全余量，单位为角秒。
	// SafetyMarginArcsec is added to coarse candidate and latitude prefilters.
	SafetyMarginArcsec float64
	// MaxEvents 为正时限制返回事件数量。
	// MaxEvents limits the number of returned events when positive.
	MaxEvents int
	// LunarLimb 月缘剖面，nil 为圆月面。站心恒星、行星月掩按剖面重新判定是否成掩与食型（可新增或取消圆月面事件），接触取最外侧；
	// 恒星多段遮挡只报首次掩入与最后复现，全部区间见 StarOccultationLimbProfile。全球路径只修正界线 / lunar limb profile; local events are re-decided on the profile.
	LunarLimb LunarLimb
}

// OccultationPathAlgorithm 选择全球月掩路径的星历求解分支。
// OccultationPathAlgorithm selects the ephemeris branch for global occultation paths.
type OccultationPathAlgorithm string

const (
	// OccultationPathAlgorithmOptimized 使用经抽检的密集星历插值，保留站心方程与连续包络。
	// OccultationPathAlgorithmOptimized uses checked dense ephemeris interpolation with the same station equations and continuous envelopes.
	OccultationPathAlgorithmOptimized OccultationPathAlgorithm = "optimized"
	// OccultationPathAlgorithmExact 保留原分支：插值预测候选，最终求解使用全项星历。
	// OccultationPathAlgorithmExact retains the original branch: interpolated candidates and full-term ephemerides for final solving.
	OccultationPathAlgorithmExact OccultationPathAlgorithm = "exact"
)

// OccultationPathOptions 控制全球月掩路径采样。
//
// Step 为路径采样的基础时间步长，正值至少为 1 秒。TargetSpacingKM 要求相邻中心线点超过目标地面距离时进行自适应加密。
// 正的 TargetSpacingKM 至少为 1 km；超过中心线或有限盘面路径工作量预算时返回 ErrOccultationPathSamplingLimit，不会静默降低请求分辨率。恒星和行星瞬时足迹使用 1 分钟目标步长和独立样本上限。
// OccultationPathOptions controls global occultation-path sampling.
// Step is the base time step used for path samples; positive values must be at least one second. TargetSpacingKM requests adaptive refinement when adjacent center-line points exceed the requested ground distance.
// Positive TargetSpacingKM values must be at least 1 km. Requests that exceed the center-line or aggregate finite-disk work budgets return ErrOccultationPathSamplingLimit instead of silently reducing resolution. Stellar and planetary instantaneous footprints use a one-minute target step and a separate sample cap.
type OccultationPathOptions struct {
	// Algorithm 的零值等同 optimized；exact 可选择原有精确分支。此选项不影响独立的单时刻月影与事件查询接口。
	// Algorithm defaults to optimized; exact selects the original branch. Independent instant-footprint and event-search APIs are unaffected.
	Algorithm OccultationPathAlgorithm
	// LunarLimb 月缘剖面，nil 为圆月面；只修正外接触南北界（掠掩线），掩带填充与等时线仍按圆月面 / lunar limb profile for graze lines only.
	LunarLimb       LunarLimb
	Step            time.Duration
	TargetSpacingKM float64
	// RiseSetStep 独立控制六类升落阶段线的采样步长；零值使用 5 分钟。
	// RiseSetStep controls horizon-curve sampling independently from the path step; zero uses five minutes.
	RiseSetStep time.Duration
	// DisableRiseSet 在不需要时跳过六类初掩、掩甚、终掩月升/月落线。
	// DisableRiseSet skips the six local phase/horizon curves when they are not needed.
	DisableRiseSet bool
	// DisableFootprints 跳过密集的独立瞬时可见区，改用稀疏支撑样本构造紧凑掩带；中心线、边界和升落阶段线仍保留。
	// DisableFootprints skips dense standalone instantaneous visible regions and uses sparse support samples to construct compact bands; the center line, limits, and rise/set curves remain available.
	DisableFootprints bool
	// IncludeFootprintTimeline 在保留紧凑静态掩带的同时，增加独立采样的瞬时足迹时间线。
	// IncludeFootprintTimeline adds independently sampled instantaneous footprints while retaining the compact static band.
	IncludeFootprintTimeline bool
	// FootprintTimelineStep 控制瞬时足迹时间线的采样间隔；启用时零值使用 5 分钟。
	// FootprintTimelineStep controls the instantaneous timeline interval; zero uses five minutes when the timeline is enabled.
	FootprintTimelineStep time.Duration
	// GreatestTimeValues 是要计算的地方掩甚时刻等值线的时刻取值（力学时儒略日，最多 64 条，超出按时间截断）；空值时改用 GreatestTimeStep。
	// 只有确实存在该时刻掩甚轨迹的取值才会出现在结果里，所以返回条数可能少于请求条数。
	// 每条等时线用固定时刻的残差零集延拓，成本正比于曲线长度而不是可见域面积。
	// GreatestTimeValues requests local greatest-occultation time isolines as TT Julian ephemeris days;
	// when empty, GreatestTimeStep is used instead. Each isochrone is continued along the zero set of a
	// fixed-instant residual so the cost scales with curve length rather than with the visible area.
	GreatestTimeValues []float64
	// GreatestTimeStep 是等时线间隔；仅在 GreatestTimeValues 为空时生效，非正值不计算等时线。
	// GreatestTimeStep is the isochrone interval; it applies only when GreatestTimeValues is empty,
	// and non-positive values disable the isolines.
	GreatestTimeStep time.Duration
}

// Validate 检查全球路径采样选项。
// Validate checks global path sampling options.
func (o OccultationPathOptions) Validate() error {
	switch o.Algorithm {
	case "", OccultationPathAlgorithmOptimized, OccultationPathAlgorithmExact:
	default:
		return fmt.Errorf("%w: unsupported path algorithm %q", ErrInvalidOccultationInput, o.Algorithm)
	}
	if o.Step < 0 {
		return fmt.Errorf("%w: path step cannot be negative", ErrInvalidOccultationInput)
	}
	if o.Step > 0 && o.Step < occultationPathMinimumStep {
		return fmt.Errorf("%w: path step must be zero or at least %s", ErrInvalidOccultationInput, occultationPathMinimumStep)
	}
	if !finite(o.TargetSpacingKM) || o.TargetSpacingKM < 0 {
		return fmt.Errorf("%w: path target spacing must be finite and non-negative", ErrInvalidOccultationInput)
	}
	if o.TargetSpacingKM > 0 && o.TargetSpacingKM < occultationPathMinimumTargetSpacingKM {
		return fmt.Errorf("%w: path target spacing must be zero or at least %.0f km", ErrInvalidOccultationInput, occultationPathMinimumTargetSpacingKM)
	}
	if o.RiseSetStep < 0 {
		return fmt.Errorf("%w: rise/set step cannot be negative", ErrInvalidOccultationInput)
	}
	if o.RiseSetStep > 0 && o.RiseSetStep < occultationPathMinimumStep {
		return fmt.Errorf("%w: rise/set step must be zero or at least %s", ErrInvalidOccultationInput, occultationPathMinimumStep)
	}
	if o.FootprintTimelineStep < 0 {
		return fmt.Errorf("%w: footprint step cannot be negative", ErrInvalidOccultationInput)
	}
	if o.FootprintTimelineStep > 0 && o.FootprintTimelineStep < occultationPathMinimumStep {
		return fmt.Errorf("%w: footprint step must be zero or at least %s", ErrInvalidOccultationInput, occultationPathMinimumStep)
	}
	return nil
}

// Validate 检查选项值，但不选择算法专用默认值。
// Validate checks option values without selecting algorithm-specific defaults.
func (o OccultationSearchOptions) Validate() error {
	if o.MaxStep < 0 {
		return fmt.Errorf("%w: search max step cannot be negative", ErrInvalidOccultationInput)
	}
	if o.MaxStep > 0 && o.MaxStep < occultationSearchMinimumStep {
		return fmt.Errorf("%w: search max step must be zero or at least %s", ErrInvalidOccultationInput, occultationSearchMinimumStep)
	}
	if !finite(o.SafetyMarginArcsec) || o.SafetyMarginArcsec < 0 {
		return fmt.Errorf("%w: search safety margin must be finite and non-negative", ErrInvalidOccultationInput)
	}
	if o.MaxEvents < 0 {
		return fmt.Errorf("%w: search max events cannot be negative", ErrInvalidOccultationInput)
	}
	return nil
}

// StarOccultationInfo 描述点光源恒星月掩；掩始和掩终是月缘交点。
// StarOccultationInfo describes a point-source stellar occultation. The immersion and emersion times are the Moon-limb crossings.
type StarOccultationInfo struct {
	TargetID string
	Observer Observer
	Type     OccultationType

	Immersion time.Time
	Greatest  time.Time
	Emersion  time.Time
	// ContactsComplete 表示两个月缘接触时刻均已求解。
	// ContactsComplete is true when both lunar-limb contacts were solved.
	ContactsComplete bool
	// LunarLimbApplied 表示接触时刻已按月缘剖面精修，否则为圆月面 / contacts refined with a lunar limb profile.
	LunarLimbApplied bool

	MinimumSeparationArcsec float64
	PositionAngleDeg        float64
	MoonSemidiameterArcsec  float64

	MoonAltitudeAtGreatest float64
	MoonAzimuthAtGreatest  float64
	VisibleAtGreatest      bool
}

// PlanetOccultationInfo 描述有限盘面行星月掩。
// ExternalImmersion 和 ExternalEmersion 分别是 C1 和 C4；全掩事件的 InternalImmersion 和 InternalEmersion 分别是 C2 和 C3，偏掩和掠掩时为零。
// ContactsComplete 表示报告几何适用的所有接触均已求解；目标按赤道半径建模为圆盘，环、大气延伸和扁率不在模型内。
// PlanetOccultationInfo describes a finite-disk planetary occultation.
// ExternalImmersion and ExternalEmersion are C1 and C4. For a total event, InternalImmersion and InternalEmersion are C2 and C3; they are zero for partial and grazing events.
// ContactsComplete means every contact applicable to the reported geometry was solved. The target is modeled as a circular disk using its equatorial body radius; rings, atmospheric extensions, and oblateness are outside this contact model.
type PlanetOccultationInfo struct {
	Planet   OccultationPlanet
	TargetID string
	Observer Observer
	Type     OccultationType

	ExternalImmersion time.Time
	InternalImmersion time.Time
	Greatest          time.Time
	InternalEmersion  time.Time
	ExternalEmersion  time.Time

	HasInternalContacts bool
	ContactsComplete    bool
	// LunarLimbApplied 表示接触时刻已按月缘剖面精修，否则为圆月面 / contacts refined with a lunar limb profile.
	LunarLimbApplied bool

	MinimumSeparationArcsec  float64
	PositionAngleDeg         float64
	MoonSemidiameterArcsec   float64
	PlanetSemidiameterArcsec float64

	MoonAltitudeAtGreatest float64
	MoonAzimuthAtGreatest  float64
	VisibleAtGreatest      bool
}

// OccultationPathPoint 是全球月掩路径上的一个地理采样点。
// Start 和 End 描述月缘外接触掩带。
// WidthKM 是中心线采样处的地面横向宽度：接触锥可见弧上地面横向偏移的极差；投影折叠的退化事件改用同刻两条横切母线与椭球交点间的弦长。
// LimitSeparationKM 只对南北限采样点（含全掩限）非零，是同一时刻对侧限线的地面间距；它与 WidthKM 构造不同，不要互相换算。
// 基础采样直接求解，自适应插入点使用宽度插值并进行五米采样误差检查。
// OccultationPathPoint is a geographic sample of a global lunar-occultation path.
// Start and End describe the outer lunar-limb footprint.
// WidthKM is the ground cross-track width at a center-line sample: the spread of ground cross-track offsets over the visible contact-cone arc, falling back to the same-instant chord between the two cross-track generators when the projection folds.
// LimitSeparationKM is non-zero only on northern/southern limit samples (total limits included) and is the same-instant ground distance to the opposite limit; it is constructed differently from WidthKM and must not be converted into it.
// Base samples are solved directly; adaptive samples use width interpolation and five-meter error checks.
type OccultationPathPoint struct {
	// Time 是该点的时刻。
	// Time is the instant the point describes.
	Time time.Time
	// Longitude 与 Latitude 是地面坐标，东经、北纬为正。
	// Longitude and Latitude are ground coordinates, east and north positive.
	Longitude float64
	Latitude  float64
	// MoonAltitude 是月球几何高度角，单位度，不做蒙气差修正。
	// MoonAltitude is the geometric lunar altitude in degrees, without refraction.
	MoonAltitude float64
	// WidthKM 是可见掩带在该点的地面宽度。掩星可见带可以宽达数千公里，与日食中心带宽度不是同一口径，不要互相比较。
	// 擦边事件的路径可以没有中心线，此时它仍表示可见带宽度，与 LimitSeparationKM 口径不同。
	// WidthKM is the ground width of the visible occultation band at the point. An occultation band
	// can span thousands of kilometres and is not comparable to a solar central-path width. A grazing
	// path can have no center line while this field still describes the visible band, which is a
	// different construction from LimitSeparationKM.
	WidthKM float64
	// LimitSeparationKM 是该点南北限的地面间距，无论中心线是否存在都有定义。
	// LimitSeparationKM is the ground separation of the two limits at the point, defined whether or not a center line exists.
	LimitSeparationKM float64
}

// OccultationGreatestTimeContour 是一个固定地方掩甚时刻的等值线支路集合。
// OccultationGreatestTimeContour contains the continuous branches of one fixed local greatest-occultation time.
type OccultationGreatestTimeContour struct {
	// JDE 是该等值线表示的力学时儒略日，也就是各支路上地方掩甚发生的时刻。
	// JDE is the TT Julian ephemeris day represented by this contour, the local greatest-occultation instant along every branch.
	JDE float64
	// Time 是 JDE 对应的时刻；按步长请求时它是原始对齐时刻，避免 JDE 往返把整分取值截断成前一分钟。
	// Time is the instant matching JDE; for step-derived levels it is the original aligned instant,
	// so a JDE round trip cannot truncate a whole-minute level into the previous minute.
	Time time.Time
	// Segments 是该时刻的连续等时线支路；一条支路两端止于月平线（几何地平，无蒙气差修正）或掩可见域边界，
	// 纬度 ±88° 以上不再延拓，同一时刻可能有多条不相连的支路。
	// Segments are continuous isochrone branches; each branch ends at the lunar horizon or the occultation-visibility boundary.
	Segments [][]OccultationPathPoint
}

// StarOccultationPath 包含点光源恒星月掩的全球掩带。
// 中心线是月心与恒星对齐的轨迹；NorthernLimit 和 SouthernLimit 是月缘外接触锥的切点轨迹（掩星在该线上恰好退化为擦边），不是中心线的等距横向平移。
// StarOccultationPath contains the global footprint of a point-source stellar occultation.
// The center line is the locus where the lunar center aligns with the star; NorthernLimit and SouthernLimit are the tangency tracks of the outer-contact cone, where the occultation degenerates to a graze, so they are not a constant offset of the center line.
type StarOccultationPath struct {
	TargetID string

	Start    OccultationPathPoint
	Greatest OccultationPathPoint
	End      OccultationPathPoint
	// Complete 表示 Start 和 End 是全球月缘外接触点，而不是查询窗口裁剪点。
	// Complete is true when Start and End are the global outer-limb contacts rather than query-window clipping points.
	Complete bool

	// CenterLine 是掩星中心线；掠掩事件可以整条没有中心线（只有南北限），此时该切片为空而 Complete 仍可为真。
	// CenterLine is the occultation center line; a grazing event can have no center line at all (limits only), leaving this slice empty while Complete can still be true.
	CenterLine    []OccultationPathPoint
	NorthernLimit []OccultationPathPoint
	SouthernLimit []OccultationPathPoint
	// GreatestLimitSeparationKM 是掩甚处南北限的地面间距（取与掩甚时刻最近的限线采样对），与 Greatest.WidthKM 口径不同，不要互相换算。
	// GreatestLimitSeparationKM is the ground separation of the two limits at greatest, taken from the limit sample pair nearest to the greatest instant; it uses a different construction from Greatest.WidthKM and must not be converted into it.
	GreatestLimitSeparationKM float64
	// BandContours 是构造静态可见掩带的连续接触包络；点光源恒星等价于月缘外接触边界。
	// BandContours are the continuous contact envelopes used to construct the static visible band; for point-source stars they are the lunar-limb outer-contact boundaries.
	BandContours [][]OccultationPathPoint
	// VisibilityContours 是掩星有效期间月球可见性的连续时间外包络。
	// VisibilityContours are the continuous temporal envelopes of lunar visibility while the occultation is active.
	VisibilityContours [][]OccultationPathPoint
	// Footprints 是时刻采样的可见点源掩区，其扫掠构成全球掩带；采样口径与行星外接触足迹一致。
	// Footprints are sampled visible point-source regions whose sweep forms the global band, using the same sampling contract as planetary outer-contact footprints.
	Footprints []OccultationFootprint
	// BandFootprints 是关闭密集瞬时足迹时用于构造紧凑掩带的稀疏可见区样本；展示层应合并它们，不应逐个输出。
	// BandFootprints are sparse visible-region samples used to construct a compact band when dense instantaneous footprints are disabled. Renderers should merge rather than emit them individually.
	BandFootprints []OccultationFootprint
	// RiseSetCurves 是初掩、掩甚和终掩分别发生在月升或月落时的六类边界。
	// RiseSetCurves are the six boundaries where local start, greatest, or end occurs at moonrise or moonset.
	RiseSetCurves []OccultationRiseSetCurve
	// GreatestTimeContours 是按掩甚时刻采样的等时线。
	// GreatestTimeContours are sampled local greatest-occultation time isolines.
	GreatestTimeContours []OccultationGreatestTimeContour
	// LunarLimbApplied 表示南北限已按月缘剖面修正，否则为圆月面 / limits corrected with a lunar limb profile.
	LunarLimbApplied bool

	Step            time.Duration
	TargetSpacingKM float64
}

// OccultationFootprint 是一个时刻的可见接触足迹。
// Polygons 包含供独立时刻绘制的闭合可见区；Boundaries 保留未经地平线封口的接触锥弧，供稀疏静态样本连续扫掠。
// Closed 表示 Boundaries 本身是否为闭合交线。
// OccultationFootprint is one instantaneous visible contact footprint.
// Polygons contain closed visible regions for rendering one instant. Boundaries retain contact-cone arcs before horizon closure for continuously sweeping sparse static samples.
// Closed reports whether Boundaries themselves form a closed intersection.
type OccultationFootprint struct {
	// Time 是该瞬时足迹对应的事件时刻。
	// Time is the event time represented by this instantaneous footprint.
	Time time.Time
	// Polygons 是供独立时刻绘制的闭合可见区域。
	// Polygons are closed visible regions for rendering one instant.
	Polygons [][]OccultationPathPoint
	// InteriorPolygons 是为单时刻绘制保留的闭合可见修复面。静态掩带构建器会将其与连续扫掠的接触边界弧分开，避免数值中心修复扩大长时间地理掩带。
	// InteriorPolygons are closed visible repair faces retained for one-instant
	// rendering. Static band builders keep them separate from the continuously
	// swept contact-boundary arcs so a numerical center repair cannot enlarge a
	// long-lived geographic band.
	InteriorPolygons [][]OccultationPathPoint
	// Boundaries 是尚未经过地平线封口的接触边界弧。
	// Boundaries are contact-boundary arcs before horizon closure.
	Boundaries [][]OccultationPathPoint
	// Closed 表示 Boundaries 是否构成闭合交线。
	// Closed reports whether Boundaries form a closed intersection.
	Closed bool
}

// OccultationRiseSetCurve 是一种局部月掩阶段与月升/月落同时发生的边界。
// OccultationRiseSetCurve is one boundary where a local occultation phase coincides with moonrise or moonset.
type OccultationRiseSetCurve struct {
	// Phase 是与月升或月落同时发生的局部月掩阶段。
	// Phase is the local occultation phase coinciding with moonrise or moonset.
	Phase RiseSetPhase
	// Direction 标识月球正在升起还是落下。
	// Direction identifies whether the Moon is rising or setting.
	Direction RiseSetDirection
	// Segments 是反经线和支路跳变安全分段后的边界采样。
	// Segments are boundary samples split safely at the antimeridian and branch changes.
	Segments [][]OccultationPathPoint
}

// PlanetOccultationFootprint 保留有限盘面行星月掩足迹的兼容名称。
// PlanetOccultationFootprint retains the compatibility name for finite-disk planetary footprints.
type PlanetOccultationFootprint = OccultationFootprint

// PlanetOccultationPath 包含有限盘面行星月掩的全球掩带。
// NorthernLimit 和 SouthernLimit 是行星盘面任意部分被覆盖的外接触边界。
// HasTotalBand 为 true 时，NorthernTotalLimit 和 SouthernTotalLimit 是行星圆盘完全被月球覆盖的内接触边界；
// 环、大气延伸和扁率不在两种接触模型内。
// PlanetOccultationPath contains the global footprint of a finite-disk planetary occultation.
// NorthernLimit and SouthernLimit are the outer-contact boundaries where any part of the planet disk is covered.
// When HasTotalBand is true, NorthernTotalLimit and SouthernTotalLimit are the inner-contact boundaries where the complete circular planet disk is covered by the Moon. Rings, atmospheric extensions, and oblateness are outside both contact models.
type PlanetOccultationPath struct {
	Planet   OccultationPlanet
	TargetID string

	Start    OccultationPathPoint
	Greatest OccultationPathPoint
	End      OccultationPathPoint
	// Complete 表示 Start 和 End 是全球外接触点。
	// Complete is true when Start and End are the global outer contacts.
	Complete bool

	// CenterLine 是影轴与椭球交点的轨迹；非中心事件退化为最接近影轴的椭球点。
	// 掠掩事件的路径可以整条没有中心线（只有南北限），此时该切片为空而 Complete 仍可为真。
	// CenterLine is the track of the shadow-axis/ellipsoid intersection, degenerating to the ellipsoid point nearest the axis for non-central events.
	// A grazing event can have no center line at all (limits only); the slice is then empty while Complete can still be true.
	CenterLine []OccultationPathPoint
	// NorthernLimit 和 SouthernLimit 是外接触锥的切点轨迹：掩星在该线上恰好退化为擦边，不是中心线的等距横向平移。
	// NorthernLimit and SouthernLimit are the tangency tracks of the outer-contact cone, where the occultation degenerates to a graze, so they are not a constant-width offset of the center line.
	NorthernLimit []OccultationPathPoint
	SouthernLimit []OccultationPathPoint
	// GreatestLimitSeparationKM 是掩甚处南北限的地面间距（取与掩甚时刻最近的限线采样对），与 Greatest.WidthKM 口径不同，不要互相换算。
	// GreatestLimitSeparationKM is the ground separation of the two limits at greatest, taken from the limit sample pair nearest to the greatest instant; it uses a different construction from Greatest.WidthKM and must not be converted into it.
	GreatestLimitSeparationKM float64
	// PartialBandContours 是偏掩静态带的连续外接触包络；任意目标盘面重叠时取零。
	// PartialBandContours are the continuous outer-contact envelopes for the static partial band where any target-disk overlap begins.
	PartialBandContours [][]OccultationPathPoint
	// PartialVisibilityContours 是月球可见性在偏掩阶段内的连续时间外包络。
	// PartialVisibilityContours are the continuous temporal envelopes of lunar visibility while partial occultation is active.
	PartialVisibilityContours [][]OccultationPathPoint
	// PartialFootprints 是时刻采样的可见外接触区域，其扫掠构成全球偏掩区域；足迹以 1 分钟为目标步长，超出独立样本上限时会变粗，
	// 每个足迹携带实际采样时刻。
	// PartialFootprints are instantaneous visible outer-contact regions whose sweep forms the global partial-occultation area. Footprints target one-minute intervals and become coarser when their independent sample cap is reached; each footprint carries its actual sample time.
	PartialFootprints []PlanetOccultationFootprint
	// PartialBandFootprints 是关闭密集瞬时足迹时用于构造紧凑偏掩带的稀疏可见区样本。
	// PartialBandFootprints are sparse visible-region samples used to construct the compact partial band when dense instantaneous footprints are disabled.
	PartialBandFootprints []PlanetOccultationFootprint
	// RiseSetCurves 是初掩、掩甚和终掩分别发生在月升或月落时的六类边界。
	// RiseSetCurves are the six boundaries where local start, greatest, or end occurs at moonrise or moonset.
	RiseSetCurves []OccultationRiseSetCurve

	HasTotalBand bool
	// TotalStart 和 TotalEnd 是全球内接触的起止点。
	// TotalStart and TotalEnd are the first and last global inner contacts.
	TotalStart OccultationPathPoint
	TotalEnd   OccultationPathPoint
	// TotalComplete 表示 TotalStart 和 TotalEnd 未被内部搜索范围截断。
	// TotalComplete is true when TotalStart and TotalEnd are not clipped by the internal search span.
	TotalComplete bool

	// NorthernTotalLimit 和 SouthernTotalLimit 是内接触（全掩）锥的切点轨迹，同刻间距同样记录在 LimitSeparationKM 上。
	// NorthernTotalLimit and SouthernTotalLimit are the tangency tracks of the inner-contact cone; their same-instant separation is recorded in LimitSeparationKM as well.
	NorthernTotalLimit []OccultationPathPoint
	SouthernTotalLimit []OccultationPathPoint
	// TotalBandContours 是全掩静态带的连续内接触包络；仅在目标圆盘完全被月球覆盖时取零。
	// TotalBandContours are the continuous inner-contact envelopes for the static total band where the complete target disk is covered by the Moon.
	TotalBandContours [][]OccultationPathPoint
	// TotalVisibilityContours 是全掩阶段内月球可见性的连续时间外包络。
	// TotalVisibilityContours are the temporal envelopes of lunar visibility while the total-occultation contact condition is active.
	TotalVisibilityContours [][]OccultationPathPoint
	// TotalFootprints 是时刻采样的可见内接触区域，其扫掠构成全球全掩区域；采样使用与 PartialFootprints 相同的 1 分钟目标步长和样本上限。
	// TotalFootprints are instantaneous visible inner-contact regions whose sweep forms the global full-coverage area. Their sampling uses the same one-minute target step and sample cap as PartialFootprints.
	TotalFootprints []PlanetOccultationFootprint
	// TotalBandFootprints 是关闭密集瞬时足迹时用于构造紧凑全掩带的稀疏可见区样本。
	// TotalBandFootprints are sparse visible-region samples used to construct the compact total band when dense instantaneous footprints are disabled.
	TotalBandFootprints []PlanetOccultationFootprint
	// TotalRiseSetCurves 是全掩带使用的内接触月升/月落边界；其相位与 RiseSetCurves 相同，但 contact metric 采用内接触而非外接触。
	// TotalRiseSetCurves are the inner-contact moonrise/moonset boundaries used by the total band; they share the same phase labels as RiseSetCurves but evaluate the inner-contact metric instead of the outer-contact metric.
	TotalRiseSetCurves []OccultationRiseSetCurve
	// GreatestTotalWidthKM 是全球掩甚时的全掩带宽度，与 Greatest.WidthKM 同口径而作用于内接触锥，且恒小于它。
	// GreatestTotalWidthKM is the full-coverage band width at global greatest, using the same construction as Greatest.WidthKM on the inner-contact cone and always below it.
	GreatestTotalWidthKM float64
	// GreatestTimeContours 是按掩甚时刻采样的等时线，判据用外接触锥，与偏掩可见域一致。
	// GreatestTimeContours are sampled local greatest-occultation time isolines, gated by the outer-contact cone so they match the partial-occultation visibility area.
	GreatestTimeContours []OccultationGreatestTimeContour
	// LunarLimbApplied 表示南北限已按月缘剖面修正，否则为圆月面 / limits corrected with a lunar limb profile.
	LunarLimbApplied bool

	Step            time.Duration
	TargetSpacingKM float64
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func validCoordinateFrame(frame CoordinateFrame) bool {
	switch frame {
	case CoordinateFrameICRS, CoordinateFrameJ2000, CoordinateFrameApparentOfDate:
		return true
	default:
		return false
	}
}
