package basic

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
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
	// OccultationGrazing 表示两边缘相切，且没有正持续时间的重叠。
	// OccultationGrazing means the limbs are tangent without a positive-duration overlap.
	OccultationGrazing OccultationType = "grazing"
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
func moonTopocentricSemidiameterN(tt float64, observer Observer, n int) float64 {
	moonRA, moonDec := HMoonGeocentricApparentRaDecN(tt, n)
	moonDistanceKM := HMoonAwayN(tt, n)
	if !finite(moonRA) || !finite(moonDec) || !finite(moonDistanceKM) || moonDistanceKM <= 0 {
		return math.NaN()
	}
	distanceKM := topocentricDistanceKM(moonRA, moonDec, moonDistanceKM, observer, TD2UT(tt, false))
	if !finite(distanceKM) || distanceKM <= 0 {
		return math.NaN()
	}
	return angularSemidiameterArcsec(moonEquatorialRadiusKM, distanceKM)
}

// topocentricDistanceKM 使用与 TopocentricRaDec 相同的 WGS-84 风格站点因子计算观测者到目标的距离 /
// The target is supplied in apparent equatorial coordinates, and the sidereal angle uses UTC/UT like TopocentricRaDec.
func topocentricDistanceKM(ra, dec, distanceKM float64, observer Observer, ut float64) float64 {
	const earthEquatorialRadius = 6378.14
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
	theta := (ApparentSiderealTime(ut)*15 + observer.Longitude) * math.Pi / 180
	observerAU := earthEquatorialRadius / astronomicalUnitKM
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
	ParallaxMas                    float64
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
	return nil
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
}

// OccultationPathOptions 控制全球月掩路径采样。
//
// Step 为路径采样的基础时间步长，正值至少为 1 秒。TargetSpacingKM 要求相邻中心线点超过目标地面距离时进行自适应加密。
// 正的 TargetSpacingKM 至少为 1 km；超过中心线或有限盘面路径工作量预算时返回 ErrOccultationPathSamplingLimit，不会静默降低请求分辨率。行星瞬时足迹使用结果中说明的独立有界采样策略。
// OccultationPathOptions controls global occultation-path sampling.
// Step is the base time step used for path samples; positive values must be at least one second. TargetSpacingKM requests adaptive refinement when adjacent center-line points exceed the requested ground distance.
// Positive TargetSpacingKM values must be at least 1 km. Requests that exceed the center-line or aggregate finite-disk work budgets return ErrOccultationPathSamplingLimit instead of silently reducing resolution. Planetary instantaneous footprints have a separate bounded sampling policy documented on the result.
type OccultationPathOptions struct {
	Step            time.Duration
	TargetSpacingKM float64
}

// Validate 检查全球路径采样选项。
// Validate checks global path sampling options.
func (o OccultationPathOptions) Validate() error {
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

	MinimumSeparationArcsec  float64
	PositionAngleDeg         float64
	MoonSemidiameterArcsec   float64
	PlanetSemidiameterArcsec float64

	MoonAltitudeAtGreatest float64
	MoonAzimuthAtGreatest  float64
	VisibleAtGreatest      bool
}

// OccultationPathPoint 是全球月掩路径上的一个地理采样点。
// Start 和 End 描述月缘外接触掩带；WidthKM 是垂直地面轨迹方向的切平面宽度，仅对中心线采样点有意义。
// 基础采样直接求解，自适应插入点使用宽度插值并进行五米采样误差检查。
// OccultationPathPoint is a geographic sample of a global lunar-occultation path.
// Start and End describe the outer lunar-limb footprint. WidthKM is the local tangent-plane width perpendicular to the ground track and is meaningful only on center-line samples.
// Base samples are solved directly; adaptive samples use width interpolation and five-meter error checks.
type OccultationPathPoint struct {
	Time         time.Time
	Longitude    float64
	Latitude     float64
	MoonAltitude float64
	WidthKM      float64
}

// StarOccultationPath 包含点光源恒星月掩的全球掩带。
// 中心线是月心与恒星对齐的轨迹；NorthernLimit 和 SouthernLimit 是中心线两侧采样的月缘外边界。
// StarOccultationPath contains the global footprint of a point-source stellar occultation.
// The center line is the locus where the lunar center aligns with the star; NorthernLimit and SouthernLimit are the two outer lunar-limb boundaries sampled beside that line.
type StarOccultationPath struct {
	TargetID string

	Start    OccultationPathPoint
	Greatest OccultationPathPoint
	End      OccultationPathPoint
	// Complete 表示 Start 和 End 是全球月缘外接触点，而不是查询窗口裁剪点。
	// Complete is true when Start and End are the global outer-limb contacts rather than query-window clipping points.
	Complete bool

	CenterLine    []OccultationPathPoint
	NorthernLimit []OccultationPathPoint
	SouthernLimit []OccultationPathPoint

	Step            time.Duration
	TargetSpacingKM float64
}

// PlanetOccultationFootprint 是一个时刻的可见接触足迹。
// Polygons 包含接触锥圆弧；当锥面与椭球的交线在朝月半球开放时，沿月球地平线闭合。
// PlanetOccultationFootprint is one instantaneous visible contact footprint.
// Polygons contain contact-cone arcs closed along the lunar horizon when the cone/ellipsoid intersection is open on the Moon-facing hemisphere.
type PlanetOccultationFootprint struct {
	Time     time.Time
	Polygons [][]OccultationPathPoint
}

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

	CenterLine    []OccultationPathPoint
	NorthernLimit []OccultationPathPoint
	SouthernLimit []OccultationPathPoint
	// PartialFootprints 是时刻采样的可见外接触区域，其扫掠构成全球偏掩区域；为限制输出和运行时间，采样可能比 Step 更粗，
	// 每个足迹携带实际采样时刻。
	// PartialFootprints are instantaneous visible outer-contact regions whose sweep forms the global partial-occultation area. To bound output and runtime, sampling may be coarser than Step; each footprint carries its actual sample time.
	PartialFootprints []PlanetOccultationFootprint

	HasTotalBand bool
	// TotalStart 和 TotalEnd 是全球内接触的起止点。
	// TotalStart and TotalEnd are the first and last global inner contacts.
	TotalStart OccultationPathPoint
	TotalEnd   OccultationPathPoint
	// TotalComplete 表示 TotalStart 和 TotalEnd 未被内部搜索范围截断。
	// TotalComplete is true when TotalStart and TotalEnd are not clipped by the internal search span.
	TotalComplete bool

	NorthernTotalLimit []OccultationPathPoint
	SouthernTotalLimit []OccultationPathPoint
	// TotalFootprints 是时刻采样的可见内接触区域，其扫掠构成全球全掩区域；采样使用与 PartialFootprints 相同的有界策略。
	// TotalFootprints are instantaneous visible inner-contact regions whose sweep forms the global full-coverage area. Their sampling uses the same bounded policy as PartialFootprints.
	TotalFootprints []PlanetOccultationFootprint
	// GreatestTotalWidthKM 是全球掩甚时的全掩带宽度。
	// GreatestTotalWidthKM is the full-coverage band width at global greatest.
	GreatestTotalWidthKM float64

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
