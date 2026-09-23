package basic

import "math"

// SolarEclipseShadowKind 阴影类型 / shadow kind.
type SolarEclipseShadowKind int

const (
	// SolarEclipseShadowUmbra 本影与反本影 / umbra and antumbra.
	SolarEclipseShadowUmbra SolarEclipseShadowKind = iota
	// SolarEclipseShadowPenumbra 半影，即偏食区 / penumbra, the partial-eclipse region.
	SolarEclipseShadowPenumbra
)

// SolarEclipseShadowSolverOptions 单时刻阴影求解器配置 / single-instant shadow solver options.
type SolarEclipseShadowSolverOptions struct {
	// Model 月亮半径模型，零值为 NASA bulletin Split-K / lunar radius model.
	Model SolarEclipseRadiusModel
	// SunRadiusModel 太阳半径口径，零值为标准档 / solar radius convention, standard when zero.
	SunRadiusModel SolarEclipseSunRadiusModel
	// DeltaTSeconds 显式 ΔT（秒），<=0 用进程级模型，只改变地球自转相位 / explicit ΔT in seconds.
	DeltaTSeconds float64
	// BoundaryPoints 边界角向采样点数，<=0 用 96 / boundary sample count.
	BoundaryPoints int
	// Kind 本影或半影，零值为本影 / umbra or penumbra, zero is the umbra.
	Kind SolarEclipseShadowKind
	// TargetSpacingKM 边界加密目标间距（千米），0 用该类型的整包采样默认值，负值表示不加密 / boundary refinement spacing in km.
	TargetSpacingKM float64
}

const (
	solarEclipseShadowDefaultBoundaryPoints = 96
	// 开放足迹的物理边界含擦地点，点数上限沿用整包口径。
	solarEclipseShadowMaximumBoundaryPoints = 1440
	// 半个朔望月：窗口内缓存锚点必为最近朔月。
	solarEclipseShadowAnchorHalfWindowDays = 14.7652944265
	// 批量条数上限，超出返回 nil。
	solarEclipseShadowMaximumBatchCount = 262144
)

// SolarEclipseShadowTopology 瞬时足迹拓扑签名，用于插值门控 / footprint topology signature.
type SolarEclipseShadowTopology struct {
	// Kind 阴影类型，签名以 umbra 或 penumbra 开头 / shadow kind, the signature prefix.
	Kind SolarEclipseShadowKind
	// Vertices 物理边界顶点总数 / boundary vertex count.
	Vertices int
	// Segments 边界分段数，换日线会拆分 / segment count.
	Segments int
	// Closed 边界是否由阴影自身闭合 / self-closed boundary.
	Closed bool
	// EnclosesPole 环绕过极点，经度已展开，不能按下标插值 / pole-winding ring.
	EnclosesPole bool
}

// Signature 稳定签名字符串，可直接作为能否插值的比较键 / stable interpolation key.
func (topology SolarEclipseShadowTopology) Signature() string {
	if topology.Vertices == 0 {
		return "empty"
	}
	builder := make([]byte, 0, 32)
	if topology.Kind == SolarEclipseShadowPenumbra {
		builder = append(builder, "penumbra"...)
	} else {
		builder = append(builder, "umbra"...)
	}
	if topology.Closed {
		builder = append(builder, "-closed"...)
	} else {
		builder = append(builder, "-horizon"...)
	}
	builder = append(builder, "-seg"...)
	builder = appendInt(builder, topology.Segments)
	builder = append(builder, "-pt"...)
	builder = appendInt(builder, topology.Vertices)
	if topology.EnclosesPole {
		builder = append(builder, "-pole"...)
	}
	return string(builder)
}

func appendInt(target []byte, value int) []byte {
	if value == 0 {
		return append(target, '0')
	}
	if value < 0 {
		target = append(target, '-')
		value = -value
	}
	var digits [20]byte
	position := len(digits)
	for value > 0 {
		position--
		digits[position] = byte('0' + value%10)
		value /= 10
	}
	return append(target, digits[position:]...)
}

// SolarEclipseShadowInstant 某瞬时（TT）的全球本影或半影足迹 / instantaneous shadow footprint.
type SolarEclipseShadowInstant struct {
	// JDE 输入的力学时儒略日 / requested TT instant.
	JDE float64
	// Model 本次使用的月亮半径模型 / lunar radius model used.
	Model SolarEclipseRadiusModel
	// SunRadiusModel 本次使用的太阳半径口径 / solar radius convention used.
	SunRadiusModel SolarEclipseSunRadiusModel
	// DeltaTSeconds 实际使用的 ΔT / ΔT actually used.
	DeltaTSeconds float64
	// Kind 本次计算的阴影类型 / shadow kind of this computation.
	Kind SolarEclipseShadowKind
	// Closed 边界由阴影自身闭合，false 表示被地平线切断 / self-closed or horizon-cut.
	Closed bool
	// Boundaries 物理边界分段，反经线会拆分 / physical boundary segments.
	Boundaries [][]SolarEclipsePathPoint
	// HorizonEnds 两端在地平圈上的擦地点，顺序与 Boundaries 一致 / horizon grazing points.
	HorizonEnds []SolarEclipsePathPoint
	// Topology 插值判定用的拓扑签名 / interpolation signature.
	Topology SolarEclipseShadowTopology
}

// Empty 该时刻阴影未落在地球表面 / no footprint on the Earth.
func (instant SolarEclipseShadowInstant) Empty() bool {
	return len(instant.Boundaries) == 0
}

// SolarEclipseShadowSolver 可复用的单时刻求解器；非并发安全，宿主每条 lane 各持一个 / reusable solver, one per lane.
type SolarEclipseShadowSolver struct {
	options SolarEclipseShadowSolverOptions

	anchorSet    bool
	anchorJDE    float64
	anchorSolver solarEclipseSolver
}

// NewSolarEclipseShadowSolver 构造单时刻求解器 / builds a single-instant solver.
func NewSolarEclipseShadowSolver(options SolarEclipseShadowSolverOptions) *SolarEclipseShadowSolver {
	options.Model = normalizeSolarEclipseRadiusModel(options.Model)
	options.SunRadiusModel = normalizeSolarEclipseSunRadiusModel(options.SunRadiusModel)
	if options.BoundaryPoints <= 0 {
		options.BoundaryPoints = solarEclipseShadowDefaultBoundaryPoints
	}
	// 下界与整包采样同口径，个位数采样点会产出无意义的足迹。
	if options.BoundaryPoints < solarEclipsePartialFootprintMinBoundaryPoints {
		options.BoundaryPoints = solarEclipsePartialFootprintMinBoundaryPoints
	}
	if options.BoundaryPoints > solarEclipseShadowMaximumBoundaryPoints {
		options.BoundaryPoints = solarEclipseShadowMaximumBoundaryPoints
	}
	if options.Kind != SolarEclipseShadowPenumbra {
		options.Kind = SolarEclipseShadowUmbra
	}
	if options.TargetSpacingKM == 0 && options.Kind == SolarEclipseShadowPenumbra {
		// 半影的整包采样默认带空间加密，单时刻要用同一默认值才能与采样逐点一致。
		options.TargetSpacingKM = solarEclipsePartialFootprintTargetSpacingKM
	}
	if options.TargetSpacingKM < 0 {
		options.TargetSpacingKM = 0
	}
	return &SolarEclipseShadowSolver{options: options}
}

// ShadowAtJDE 给定 TT 时刻的阴影足迹；不在地球上时返回 (零值, false) / footprint at one TT instant.
func (solver *SolarEclipseShadowSolver) ShadowAtJDE(jdeTT float64) (SolarEclipseShadowInstant, bool) {
	if !finite(jdeTT) {
		return SolarEclipseShadowInstant{}, false
	}
	deltaT := solver.effectiveDeltaT(jdeTT)
	inner := solver.solverFor(jdeTT)
	moon, axis, sun := solver.besselGeometryAt(jdeTT, inner, deltaT)
	footprint := inner.shadowFootprintAtWithGeometry(
		jdeTT, moon, axis, sun, solver.options.BoundaryPoints,
		solver.options.shadowKind(), solver.options.TargetSpacingKM,
	)
	if len(footprint.Boundaries) == 0 {
		return SolarEclipseShadowInstant{
			JDE: jdeTT, Model: solver.options.Model, SunRadiusModel: solver.options.SunRadiusModel,
			Kind: solver.options.Kind, DeltaTSeconds: deltaT,
		}, false
	}
	return SolarEclipseShadowInstant{
		JDE:            jdeTT,
		Model:          solver.options.Model,
		SunRadiusModel: solver.options.SunRadiusModel,
		Kind:           solver.options.Kind,
		DeltaTSeconds:  deltaT,
		Closed:         footprint.Closed,
		Boundaries:     footprint.Boundaries,
		HorizonEnds:    footprint.HorizonEnds,
		Topology:       solarEclipseShadowFootprintTopology(footprint, solver.options.Kind),
	}, true
}

func (options SolarEclipseShadowSolverOptions) shadowKind() solarEclipseShadowKind {
	if options.Kind == SolarEclipseShadowPenumbra {
		return solarEclipsePenumbralShadow
	}
	return solarEclipseCentralShadow
}

func (solver *SolarEclipseShadowSolver) effectiveDeltaT(jdeTT float64) float64 {
	override := solver.options.DeltaTSeconds
	if override <= 0 {
		// ΔT 模型的自变量是 UT1，须先从输入的 TT 反解。
		return ut1ToTTOffsetSeconds(ttToUT1JDE(jdeTT))
	}
	return DeltaTSecondsAt(jdeTT, override)
}

func (solver *SolarEclipseShadowSolver) solverFor(jdeTT float64) solarEclipseSolver {
	if solver.anchorSet && math.Abs(jdeTT-solver.anchorJDE) <= solarEclipseShadowAnchorHalfWindowDays {
		return solver.anchorSolver
	}
	anchor := CalcMoonSHByJDE(jdeTT, 0)
	solver.anchorSet = true
	solver.anchorJDE = anchor
	solver.anchorSolver = newSolarEclipseSolverWithOptions(anchor, SolarEclipseOptions{
		RadiusModel:    solver.options.Model,
		SunRadiusModel: solver.options.SunRadiusModel,
	})
	return solver.anchorSolver
}

func (solver *SolarEclipseShadowSolver) besselGeometryAt(
	jdeTT float64,
	inner solarEclipseSolver,
	deltaTSeconds float64,
) ([3]float64, solarEclipseAxis, [3]float64) {
	if solver.options.DeltaTSeconds <= 0 {
		return inner.besselGeometryAt(jdeTT)
	}
	sunEquatorial, moonEquatorial := solarEclipseSunMoonEquatorial(jdeTT)
	axis := solarEclipseBesselAxisFromEquatorialWithDeltaT(
		jdeTT, sunEquatorial, moonEquatorial, deltaTSeconds,
	)
	return solarEclipseBesselMoonFromEquatorial(moonEquatorial, axis), axis, sunEquatorial
}

// SolarEclipseShadowAtJDE 是无状态版本：自建一次内部状态后求单时刻足迹，任意并发安全。
// SolarEclipseShadowAtJDE is the stateless variant: it builds its own state per call and
// is therefore safe for concurrent use.
func SolarEclipseShadowAtJDE(
	jdeTT float64,
	options SolarEclipseShadowSolverOptions,
) (SolarEclipseShadowInstant, bool) {
	return NewSolarEclipseShadowSolver(options).ShadowAtJDE(jdeTT)
}

func solarEclipseShadowFootprintTopology(
	footprint SolarEclipsePartialFootprint,
	kind SolarEclipseShadowKind,
) SolarEclipseShadowTopology {
	topology := SolarEclipseShadowTopology{
		Kind:     kind,
		Segments: len(footprint.Boundaries),
		Closed:   footprint.Closed,
	}
	winding := 0.0
	for _, segment := range footprint.Boundaries {
		topology.Vertices += len(segment)
		for index := 1; index < len(segment); index++ {
			winding += math.Remainder(segment[index].Longitude-segment[index-1].Longitude, 360)
		}
		// 只有真正闭合的分段才把"末点回到首点"计入绕极判定；开放分段的收口边是虚拟的。
		if len(segment) > 2 && solarEclipsePathDistanceKM(segment[0], segment[len(segment)-1]) < 1e-6 {
			winding += math.Remainder(segment[0].Longitude-segment[len(segment)-1].Longitude, 360)
		}
	}
	if topology.Segments == 1 && math.Abs(winding) >= 180 {
		topology.EnclosesPole = true
	}
	return topology
}

// SolarEclipseStationState 某瞬时的站心日月几何 / topocentric geometry at one instant.
type SolarEclipseStationState struct {
	// JDE 输入的 TT 时刻，DeltaTSeconds 实际使用的 ΔT / requested instant and ΔT used.
	JDE           float64
	DeltaTSeconds float64
	// SeparationDeg 与 SeparationArcsec 是日月中心的站心角距，极小化用的连续量 / topocentric separation.
	SeparationDeg    float64
	SeparationArcsec float64
	// 三个半径字段是日月站心视半径 / topocentric apparent radii.
	SunRadiusDeg       float64
	MoonOuterRadiusDeg float64
	MoonInnerRadiusDeg float64
	// SunAltitudeDeg 与 SunAzimuthDeg 是站心太阳高度角与方位角，方位角自北向东 / solar altitude and azimuth.
	SunAltitudeDeg float64
	SunAzimuthDeg  float64
	// Magnitude 瞬时食分，Obscuration 太阳视面积遮蔽率 / instantaneous magnitude and obscuration.
	Magnitude   float64
	Obscuration float64
	// InCentralPhase 该瞬时站点位于本影或反本影内 / station inside the central shadow now.
	InCentralPhase bool
	// CentralPhaseType 中心食类型，非中心食为 SolarEclipseNone / central phase kind.
	CentralPhaseType SolarEclipseType
	// HasTotalPhase 与 HasAnnularPhase 表示该瞬时是否处于全食或环食 / total or annular now.
	HasTotalPhase   bool
	HasAnnularPhase bool
	// Visible 太阳中心高于几何地平，高度用俯仰角修正阈值 / Sun center above the horizon.
	Visible bool
}

// StationStateAtJDE 给定 TT 时刻与站点的站心情形，任何时刻可调用且不报错 / topocentric state at one instant.
func (solver *SolarEclipseShadowSolver) StationStateAtJDE(
	jdeTT, lonDeg, latDeg, heightMeters float64,
) SolarEclipseStationState {
	if !finite(jdeTT) || !finite(lonDeg) || !finite(latDeg) {
		return SolarEclipseStationState{JDE: jdeTT}
	}
	deltaT := solver.effectiveDeltaT(jdeTT)
	heightKM := heightMeters / 1000
	state := localSolarEclipseStateAtWithDeltaT(
		jdeTT, deltaT, lonDeg*rad, latDeg*rad, heightKM,
		solarEclipseModelParams(solver.options.Model, solver.options.SunRadiusModel),
	)
	contact := state.movingDiskContactState()
	central := contact.internalContactGap() <= 0
	result := SolarEclipseStationState{
		JDE:                jdeTT,
		DeltaTSeconds:      deltaT,
		SeparationDeg:      state.separationRad / rad,
		SeparationArcsec:   state.separationRad / rad * 3600,
		SunRadiusDeg:       state.sunRadiusRad / rad,
		MoonOuterRadiusDeg: state.moonOuterRadiusRad / rad,
		MoonInnerRadiusDeg: state.moonInnerRadiusRad / rad,
		SunAltitudeDeg:     state.sunAltitudeRad / rad,
		SunAzimuthDeg:      state.sunAzimuthRad / rad,
		Obscuration:        localSolarEclipseObscuration(state.sunRadiusRad, state.moonOuterRadiusRad, state.separationRad),
		InCentralPhase:     central,
	}
	visibleThreshold := 0.0
	if heightMeters > 0 {
		visibleThreshold = -HeightDegreeByLat(heightMeters, latDeg) * rad
	}
	result.Visible = state.sunAltitudeRad > visibleThreshold
	if central {
		result.Magnitude = state.moonInnerRadiusRad / state.sunRadiusRad
		if state.moonInnerRadiusRad >= state.sunRadiusRad {
			result.CentralPhaseType = SolarEclipseTotal
			result.HasTotalPhase = true
		} else {
			result.CentralPhaseType = SolarEclipseAnnular
			result.HasAnnularPhase = true
		}
		return result
	}
	result.Magnitude = (state.moonOuterRadiusRad + state.sunRadiusRad - state.separationRad) / (2 * state.sunRadiusRad)
	if result.Magnitude < 0 {
		result.Magnitude = 0
	}
	return result
}

// SolarEclipseStationStateAtJDE 无状态版本，可并发调用 / stateless, concurrency-safe variant.
func SolarEclipseStationStateAtJDE(
	jdeTT, lonDeg, latDeg, heightMeters float64,
	options SolarEclipseShadowSolverOptions,
) SolarEclipseStationState {
	return NewSolarEclipseShadowSolver(options).StationStateAtJDE(jdeTT, lonDeg, latDeg, heightMeters)
}

func localSolarEclipseStateAtWithDeltaT(
	jdTT, deltaTSeconds, lonRad, latRad, heightKM float64,
	params solarEclipseModelParameters,
) localSolarEclipseState {
	context := newLocalSolarEclipseStateContextWithDeltaT(jdTT, deltaTSeconds, params)
	return context.stateAt(lonRad, latRad, heightKM)
}

func newLocalSolarEclipseStateContextWithDeltaT(
	jdTT, deltaTSeconds float64,
	params solarEclipseModelParameters,
) localSolarEclipseStateContext {
	sunEquatorial, moonEquatorial := solarEclipseSunMoonEquatorial(jdTT)
	utJDE := jdTT - deltaTSeconds/86400
	return localSolarEclipseStateContext{
		sunXYZ:  solarEclipseLLRToXYZ(sunEquatorial[0], sunEquatorial[1], sunEquatorial[2]),
		moonXYZ: solarEclipseLLRToXYZ(moonEquatorial[0], moonEquatorial[1], moonEquatorial[2]),
		gst:     ApparentSiderealTime(utJDE) * 15 * rad,
		params:  params,
	}
}

// ShadowBetweenJDE 区间内等步长逐时刻的阴影足迹，无阴影时刻为空条目 / per-step footprints.
func (solver *SolarEclipseShadowSolver) ShadowBetweenJDE(
	startJDE, endJDE, stepDays float64,
) []SolarEclipseShadowInstant {
	if !finite(startJDE) || !finite(endJDE) || !finite(stepDays) || stepDays <= 0 || endJDE < startJDE {
		return nil
	}
	count := solarEclipseShadowBatchCount(startJDE, endJDE, stepDays)
	if count <= 0 || count > solarEclipseShadowMaximumBatchCount {
		return nil
	}
	result := make([]SolarEclipseShadowInstant, 0, count)
	for index := 0; index < count; index++ {
		instant, _ := solver.ShadowAtJDE(startJDE + float64(index)*stepDays)
		result = append(result, instant)
	}
	return result
}

// StationStatesBetweenJDE 区间内等步长逐时刻的站心情形 / per-step station states.
func (solver *SolarEclipseShadowSolver) StationStatesBetweenJDE(
	startJDE, endJDE, stepDays, lonDeg, latDeg, heightMeters float64,
) []SolarEclipseStationState {
	if !finite(startJDE) || !finite(endJDE) || !finite(stepDays) || stepDays <= 0 || endJDE < startJDE {
		return nil
	}
	count := solarEclipseShadowBatchCount(startJDE, endJDE, stepDays)
	if count <= 0 || count > solarEclipseShadowMaximumBatchCount {
		return nil
	}
	result := make([]SolarEclipseStationState, 0, count)
	for index := 0; index < count; index++ {
		result = append(result, solver.StationStateAtJDE(
			startJDE+float64(index)*stepDays, lonDeg, latDeg, heightMeters,
		))
	}
	return result
}

// solarEclipseShadowBatchTimeTolerance 是批量采样末点的时刻容差（天，约 0.9 ms）：
// 起止点由浮点运算给出时末点会落在 end 之外若干个 ULP，容差内仍算入，避免丢掉本应包含的采样。
const solarEclipseShadowBatchTimeTolerance = 1e-8

// solarEclipseShadowBatchCount 返回闭区间上按 stepDays 采样的格点数。
func solarEclipseShadowBatchCount(startJDE, endJDE, stepDays float64) int {
	return int(math.Floor((endJDE-startJDE+solarEclipseShadowBatchTimeTolerance)/stepDays)) + 1
}

// SolarEclipseShadowBetweenJDE 无状态批量版本 / stateless batch variant.
func SolarEclipseShadowBetweenJDE(
	startJDE, endJDE, stepDays float64, options SolarEclipseShadowSolverOptions,
) []SolarEclipseShadowInstant {
	return NewSolarEclipseShadowSolver(options).ShadowBetweenJDE(startJDE, endJDE, stepDays)
}
