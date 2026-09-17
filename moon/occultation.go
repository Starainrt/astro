package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ErrInvalidOccultationInput 表示月掩输入无效。
// ErrInvalidOccultationInput reports invalid lunar-occultation input.
var ErrInvalidOccultationInput = basic.ErrInvalidOccultationInput

// ErrOccultationPathSamplingLimit 表示请求的路径分辨率超过确定性采样预算。
// ErrOccultationPathSamplingLimit reports that the requested path resolution exceeds the deterministic sampling budget.
var ErrOccultationPathSamplingLimit = basic.ErrOccultationPathSamplingLimit

// CoordinateFrame 标识恒星坐标的赤道参考系。
// CoordinateFrame identifies the equatorial frame of a stellar coordinate.
type CoordinateFrame = basic.CoordinateFrame

const (
	// CoordinateFrameICRS 表示 ICRS 星表坐标系。
	// CoordinateFrameICRS is the ICRS catalog frame.
	CoordinateFrameICRS = basic.CoordinateFrameICRS
	// CoordinateFrameJ2000 表示 J2000 平均赤道坐标系。
	// CoordinateFrameJ2000 is the mean equatorial J2000 frame.
	CoordinateFrameJ2000 = basic.CoordinateFrameJ2000
	// CoordinateFrameApparentOfDate 表示历元时刻的视赤道坐标系。
	// CoordinateFrameApparentOfDate is the apparent equatorial frame of date.
	CoordinateFrameApparentOfDate = basic.CoordinateFrameApparentOfDate
)

// OccultationType 标识点光源或有限盘面月掩几何。
// OccultationType identifies point-source or finite-disk occultation geometry.
type OccultationType = basic.OccultationType

const (
	// OccultationTotal 表示掩甚时目标盘面被完全覆盖。
	// OccultationTotal means the target disk is fully covered at greatest occultation.
	OccultationTotal = basic.OccultationTotal
	// OccultationPartial 表示掩甚时有限目标盘面只有部分被覆盖。
	// OccultationPartial means only part of a finite target disk is covered at greatest occultation.
	OccultationPartial = basic.OccultationPartial
	// OccultationGrazing 表示两边缘相切，且没有正持续时间的重叠。
	// OccultationGrazing means the limbs are tangent without a positive-duration overlap.
	OccultationGrazing = basic.OccultationGrazing
)

// OccultationPlanet 标识有限盘面行星目标。
// OccultationPlanet identifies a finite-disk planetary target.
type OccultationPlanet = basic.OccultationPlanet

const (
	// OccultationMercury 表示水星有限盘面目标。
	// OccultationMercury identifies Mercury as the finite-disk target.
	OccultationMercury = basic.OccultationMercury
	// OccultationVenus 表示金星有限盘面目标。
	// OccultationVenus identifies Venus as the finite-disk target.
	OccultationVenus = basic.OccultationVenus
	// OccultationMars 表示火星有限盘面目标。
	// OccultationMars identifies Mars as the finite-disk target.
	OccultationMars = basic.OccultationMars
	// OccultationJupiter 表示木星有限盘面目标。
	// OccultationJupiter identifies Jupiter as the finite-disk target.
	OccultationJupiter = basic.OccultationJupiter
	// OccultationSaturn 表示土星有限盘面目标。
	// OccultationSaturn identifies Saturn as the finite-disk target.
	OccultationSaturn = basic.OccultationSaturn
	// OccultationUranus 表示天王星有限盘面目标。
	// OccultationUranus identifies Uranus as the finite-disk target.
	OccultationUranus = basic.OccultationUranus
	// OccultationNeptune 表示海王星有限盘面目标。
	// OccultationNeptune identifies Neptune as the finite-disk target.
	OccultationNeptune = basic.OccultationNeptune
)

// Observer 描述站心观测地点。
// Observer describes a topocentric observing site.
type Observer = basic.Observer

// StarCoordinate 是调用者提供的恒星星表坐标或视位置坐标。
// StarCoordinate is a catalog or apparent coordinate supplied for a star.
type StarCoordinate = basic.StarCoordinate

// StarData 是内嵌 9100 星表中的一条记录。
// StarData is one entry from the embedded 9100-star catalog.
type StarData = basic.StarData

// StarCoordinateFromStarData 将星表记录转换为恒星月掩搜索接受的 J2000 输入；不会加载星表。
// StarCoordinateFromStarData converts a catalog entry into the J2000 input accepted by stellar lunar-occultation searches. It does not load the catalog.
func StarCoordinateFromStarData(star StarData) (StarCoordinate, error) {
	return basic.StarCoordinateFromStarData(star)
}

// OccultationSearchOptions 控制恒星和行星月掩搜索。
// OccultationSearchOptions controls stellar and planetary occultation searches.
type OccultationSearchOptions = basic.OccultationSearchOptions

// OccultationPathOptions 控制全球月掩路径采样。
// OccultationPathOptions controls global occultation-path sampling.
type OccultationPathOptions = basic.OccultationPathOptions

// OccultationPathAlgorithm 选择全球月掩路径的优化或原有精确分支。
// OccultationPathAlgorithm selects the optimized or original exact global-path branch.
type OccultationPathAlgorithm = basic.OccultationPathAlgorithm

// 全球月掩路径的两条算法分支取值 / the two global occultation-path algorithm values.
const (
	OccultationPathAlgorithmOptimized = basic.OccultationPathAlgorithmOptimized
	OccultationPathAlgorithmExact     = basic.OccultationPathAlgorithmExact
)

// StarOccultationInfo 描述一次点光源恒星月掩。
// StarOccultationInfo describes one point-source stellar occultation.
type StarOccultationInfo = basic.StarOccultationInfo

// PlanetOccultationInfo 描述一次有限盘面行星月掩。
// PlanetOccultationInfo describes one finite-disk planetary occultation.
type PlanetOccultationInfo = basic.PlanetOccultationInfo

// OccultationPathPoint 是全球月掩路径上的地理采样点。
// OccultationPathPoint is a geographic sample of a global occultation path.
type OccultationPathPoint = basic.OccultationPathPoint

// RiseSetPhase 标识局部月掩阶段。
// RiseSetPhase identifies a local occultation phase.
type RiseSetPhase = basic.RiseSetPhase

// RiseSetDirection 标识月升或月落方向。
// RiseSetDirection identifies moonrise or moonset.
type RiseSetDirection = basic.RiseSetDirection

// 与 basic 同名的阶段与方向常量 / the phase and direction constants re-exported from basic.
const (
	RiseSetPhaseStart    = basic.RiseSetPhaseStart
	RiseSetPhaseGreatest = basic.RiseSetPhaseGreatest
	RiseSetPhaseEnd      = basic.RiseSetPhaseEnd
	RiseSetDirectionRise = basic.RiseSetDirectionRise
	RiseSetDirectionSet  = basic.RiseSetDirectionSet
)

// OccultationRiseSetCurve 是局部阶段发生在月升或月落时的边界。
// OccultationRiseSetCurve is a boundary where a local phase occurs at moonrise or moonset.
type OccultationRiseSetCurve = basic.OccultationRiseSetCurve

// OccultationGreatestTimeContour 是一个固定地方掩甚时刻的等值线支路集合。
// OccultationGreatestTimeContour contains the continuous branches of one fixed local greatest-occultation time.
type OccultationGreatestTimeContour = basic.OccultationGreatestTimeContour

// StarOccultationPath 包含点光源恒星月掩掩带。
// StarOccultationPath contains a point-source stellar occultation footprint.
type StarOccultationPath = basic.StarOccultationPath

// OccultationFootprint 是一个时刻的可见接触区域。
// OccultationFootprint is one instantaneous visible contact region.
type OccultationFootprint = basic.OccultationFootprint

// PlanetOccultationPath 包含有限盘面行星月掩掩带。
// PlanetOccultationPath contains a finite-disk planetary occultation footprint.
type PlanetOccultationPath = basic.PlanetOccultationPath

// PlanetOccultationFootprint 是一个时刻的可见行星接触区域。
// PlanetOccultationFootprint is one instantaneous visible planetary-contact region.
type PlanetOccultationFootprint = basic.PlanetOccultationFootprint

// StarOccultationInstant 是一个请求时刻的点源恒星月掩可见足迹。
// StarOccultationInstant is the visible point-source footprint at one requested instant.
type StarOccultationInstant = basic.StarOccultationInstant

// PlanetOccultationInstant 是一个请求时刻的行星偏掩和全掩可见足迹。
// PlanetOccultationInstant contains the visible partial and total planetary footprints at one requested instant.
type PlanetOccultationInstant = basic.PlanetOccultationInstant

// StarOccultationFootprintAt 计算指定时刻的精确恒星月掩可见足迹。
// StarOccultationFootprintAt calculates the exact visible stellar-occultation footprint at one instant.
func StarOccultationFootprintAt(at time.Time, star StarCoordinate) (StarOccultationInstant, error) {
	return basic.StarOccultationFootprintAt(at, star)
}

// PlanetOccultationFootprintsAt 计算指定时刻的精确行星偏掩和全掩可见足迹。
// PlanetOccultationFootprintsAt calculates the exact visible partial and total planetary footprints at one instant.
func PlanetOccultationFootprintsAt(at time.Time, planet OccultationPlanet) (PlanetOccultationInstant, error) {
	return basic.PlanetOccultationFootprintsAt(at, planet)
}
