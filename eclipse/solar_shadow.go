package eclipse

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

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
	// DeltaTSeconds 显式 ΔT（秒），<=0 用进程级模型，只改变地球自转相位 / explicit ΔT in seconds.
	DeltaTSeconds float64
	// BoundaryPoints 边界角向采样点数，<=0 用 96 / boundary sample count.
	BoundaryPoints int
	// Kind 本影或半影，零值为本影 / umbra or penumbra, zero is the umbra.
	Kind SolarEclipseShadowKind
	// TargetSpacingKM 边界加密目标间距（千米），0 用该类型的整包采样默认值，负值表示不加密 / boundary refinement spacing in km.
	TargetSpacingKM float64
}

// SolarEclipseShadowTopology 瞬时足迹拓扑签名，用于插值门控 / footprint topology signature.
type SolarEclipseShadowTopology struct {
	// Kind 阴影类型 / shadow kind.
	Kind SolarEclipseShadowKind
	// Vertices、Segments 与 Closed 描述边界形状 / boundary shape.
	Vertices int
	Segments int
	Closed   bool
	// EnclosesPole 环绕过极点，经度已展开，不能按下标插值 / pole-winding ring.
	EnclosesPole bool
}

// Signature 稳定签名字符串，可直接作为能否插值的比较键 / stable interpolation key.
func (topology SolarEclipseShadowTopology) Signature() string {
	return basic.SolarEclipseShadowTopology{
		Kind:         basic.SolarEclipseShadowKind(topology.Kind),
		Vertices:     topology.Vertices,
		Segments:     topology.Segments,
		Closed:       topology.Closed,
		EnclosesPole: topology.EnclosesPole,
	}.Signature()
}

// SolarEclipseShadowInstant 某瞬时的全球本影或半影足迹 / instantaneous shadow footprint.
type SolarEclipseShadowInstant struct {
	// Time 记录对应的时刻：UTC 入口为调用方所给，TT 入口按句柄 ΔT 换算 / instant this record describes.
	Time time.Time
	// JDE 几何使用的力学时儒略日，ΔT 不改变它 / TT instant used by the geometry.
	JDE float64
	// DeltaTSeconds 实际使用的 ΔT（秒），来自句柄显式值或进程级模型 / ΔT actually used.
	DeltaTSeconds float64
	// Model 本次使用的月亮半径模型 / lunar radius model used.
	Model SolarEclipseRadiusModel
	// Kind 本次计算的阴影类型 / shadow kind of this computation.
	Kind SolarEclipseShadowKind
	// Closed 边界由阴影自身闭合，false 表示被地平线切断 / self-closed or horizon-cut.
	Closed bool
	// Boundaries 物理边界分段，HorizonEnds 是两端的地平擦地点 / segments and grazing points.
	Boundaries  [][]SolarEclipsePathPoint
	HorizonEnds []SolarEclipsePathPoint
	// Topology 插值判定用的拓扑签名 / interpolation signature.
	Topology SolarEclipseShadowTopology
}

// Empty 该时刻阴影未落在地球表面 / no footprint on the Earth.
func (instant SolarEclipseShadowInstant) Empty() bool {
	return len(instant.Boundaries) == 0
}

// SolarEclipseStationState 某瞬时的站心日月几何 / topocentric geometry at one instant.
type SolarEclipseStationState struct {
	Time time.Time
	// JDE 本次计算使用的力学时儒略日 / TT instant used.
	JDE float64
	// DeltaTSeconds 实际使用的 ΔT（秒） / ΔT actually used.
	DeltaTSeconds float64
	// SeparationDeg 与 SeparationArcsec 是日月中心的站心角距 / topocentric separation.
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
	// 中心食标志：位于本影或反本影内，以及全食或环食 / central phase flags.
	InCentralPhase  bool
	HasTotalPhase   bool
	HasAnnularPhase bool
	// CentralPhaseType 中心食类型，非中心食为 SolarEclipseNone / central phase kind.
	CentralPhaseType SolarEclipseType
	// Visible 太阳中心在地平线上，海拔用俯仰角修正 / Sun center above the horizon.
	Visible bool
}

// SolarEclipseShadowSolver 可复用的单时刻求解器；非并发安全，宿主每条 lane 各持一个 / reusable solver, one per lane.
type SolarEclipseShadowSolver struct {
	options SolarEclipseShadowSolverOptions
	inner   *basic.SolarEclipseShadowSolver
}

// NewSolarEclipseShadowSolver 构造单时刻求解器 / builds a single-instant solver.
func NewSolarEclipseShadowSolver(options SolarEclipseShadowSolverOptions) *SolarEclipseShadowSolver {
	return &SolarEclipseShadowSolver{
		options: options,
		inner: basic.NewSolarEclipseShadowSolver(basic.SolarEclipseShadowSolverOptions{
			Model:           basic.SolarEclipseRadiusModel(options.Model),
			DeltaTSeconds:   options.DeltaTSeconds,
			BoundaryPoints:  options.BoundaryPoints,
			Kind:            basic.SolarEclipseShadowKind(options.Kind),
			TargetSpacingKM: options.TargetSpacingKM,
		}),
	}
}

func (solver *SolarEclipseShadowSolver) ttJDE(value time.Time) float64 {
	utJDE := basic.Date2JDE(value.UTC())
	if solver.options.DeltaTSeconds > 0 {
		return utJDE + solver.options.DeltaTSeconds/86400
	}
	return basic.TD2UT(utJDE, true)
}

// ShadowAt 给定 UTC 时刻的阴影足迹；不在地球上时返回 (零值, false) / footprint at one UTC instant.
func (solver *SolarEclipseShadowSolver) ShadowAt(value time.Time) (SolarEclipseShadowInstant, bool) {
	return solver.shadowAt(value, solver.ttJDE(value))
}

// ShadowAtJDE 以 TT 时刻取足迹，ΔT 只改变地面位置 / footprint at one TT instant.
func (solver *SolarEclipseShadowSolver) ShadowAtJDE(jdeTT float64) (SolarEclipseShadowInstant, bool) {
	if !finiteEclipseValue(jdeTT) {
		return SolarEclipseShadowInstant{}, false
	}
	return solver.shadowAt(time.Time{}, jdeTT)
}

func (solver *SolarEclipseShadowSolver) shadowAt(
	value time.Time,
	jdeTT float64,
) (SolarEclipseShadowInstant, bool) {
	instant, ok := solver.inner.ShadowAtJDE(jdeTT)
	return solarEclipseShadowInstantFromBasic(value, instant), ok
}

func solarEclipseShadowInstantFromBasic(
	value time.Time, instant basic.SolarEclipseShadowInstant,
) SolarEclipseShadowInstant {
	location := value.Location()
	if value.IsZero() {
		value = solarEclipseTTJDEToTimeWithDeltaT(instant.JDE, instant.DeltaTSeconds, time.UTC)
		location = time.UTC
	}
	return SolarEclipseShadowInstant{
		Time:          value,
		JDE:           instant.JDE,
		DeltaTSeconds: instant.DeltaTSeconds,
		Model:         mapBasicSolarEclipseModel(instant.Model),
		Kind:          SolarEclipseShadowKind(instant.Kind),
		Closed:        instant.Closed,
		Boundaries:    solarEclipsePathSegmentsFromBasic(instant.Boundaries, location),
		HorizonEnds:   solarEclipsePathPointsFromBasic(instant.HorizonEnds, location),
		Topology: SolarEclipseShadowTopology{
			Kind:         SolarEclipseShadowKind(instant.Topology.Kind),
			Vertices:     instant.Topology.Vertices,
			Segments:     instant.Topology.Segments,
			Closed:       instant.Topology.Closed,
			EnclosesPole: instant.Topology.EnclosesPole,
		},
	}
}

// StationStateAt 给定 UTC 时刻与站点的站心情形，任何时刻可调用 / topocentric state at one instant.
func (solver *SolarEclipseShadowSolver) StationStateAt(
	value time.Time, lon, lat, height float64,
) SolarEclipseStationState {
	return solver.stationState(value, solver.ttJDE(value), lon, lat, height)
}

// StationStateAtJDE 以 TT 时刻取站心情形，ΔT 只影响地球自转相位 / topocentric state at one TT instant.
func (solver *SolarEclipseShadowSolver) StationStateAtJDE(
	jdeTT, lon, lat, height float64,
) SolarEclipseStationState {
	if !finiteEclipseValue(jdeTT) {
		return SolarEclipseStationState{}
	}
	state := solver.inner.StationStateAtJDE(jdeTT, lon, lat, height)
	result := solarEclipseStationStateFromBasic(time.Time{}, state)
	result.Time = solarEclipseTTJDEToTimeWithDeltaT(jdeTT, state.DeltaTSeconds, time.UTC)
	return result
}

func (solver *SolarEclipseShadowSolver) stationState(
	value time.Time, jdeTT, lon, lat, height float64,
) SolarEclipseStationState {
	return solarEclipseStationStateFromBasic(value, solver.inner.StationStateAtJDE(jdeTT, lon, lat, height))
}

func solarEclipseStationStateFromBasic(
	value time.Time, state basic.SolarEclipseStationState,
) SolarEclipseStationState {
	return SolarEclipseStationState{
		Time:               value,
		JDE:                state.JDE,
		DeltaTSeconds:      state.DeltaTSeconds,
		SeparationDeg:      state.SeparationDeg,
		SeparationArcsec:   state.SeparationArcsec,
		SunRadiusDeg:       state.SunRadiusDeg,
		MoonOuterRadiusDeg: state.MoonOuterRadiusDeg,
		MoonInnerRadiusDeg: state.MoonInnerRadiusDeg,
		SunAltitudeDeg:     state.SunAltitudeDeg,
		SunAzimuthDeg:      state.SunAzimuthDeg,
		Magnitude:          state.Magnitude,
		Obscuration:        state.Obscuration,
		InCentralPhase:     state.InCentralPhase,
		HasTotalPhase:      state.HasTotalPhase,
		HasAnnularPhase:    state.HasAnnularPhase,
		CentralPhaseType:   mapBasicSolarEclipseType(state.CentralPhaseType),
		Visible:            state.Visible,
	}
}

// SolarEclipseShadowAt 无状态便捷入口，可并发调用 / stateless convenience entry point.
func SolarEclipseShadowAt(value time.Time) (SolarEclipseShadowInstant, bool) {
	return NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{}).ShadowAt(value)
}

// SolarEclipseStationStateAt 无状态便捷入口，可并发调用 / stateless convenience entry point.
func SolarEclipseStationStateAt(value time.Time, lon, lat, height float64) SolarEclipseStationState {
	return NewSolarEclipseShadowSolver(SolarEclipseShadowSolverOptions{}).StationStateAt(value, lon, lat, height)
}

func finiteEclipseValue(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

// ShadowBetween 区间内等步长逐时刻的阴影足迹，超过内部上限时返回 nil / per-step footprints over a range.
func (solver *SolarEclipseShadowSolver) ShadowBetween(
	start, end time.Time, step time.Duration,
) []SolarEclipseShadowInstant {
	if step <= 0 || !end.After(start) {
		return nil
	}
	startTT := solver.ttJDE(start)
	endTT := solver.ttJDE(end)
	stepDays := step.Hours() / 24
	if stepDays <= 0 {
		return nil
	}
	inner := solver.inner.ShadowBetweenJDE(startTT, endTT, stepDays)
	result := make([]SolarEclipseShadowInstant, 0, len(inner))
	// 逐条累加而非 index*step：Duration 是 int64 纳秒，长窗口下乘法会溢出。
	instantTime := start
	for _, instant := range inner {
		// 直接映射，避免为每个时刻重复解算一次足迹。
		result = append(result, solarEclipseShadowInstantFromBasic(instantTime, instant))
		instantTime = instantTime.Add(step)
	}
	return result
}

// StationStatesBetween 区间内等步长逐时刻的站心情形 / per-step station states.
func (solver *SolarEclipseShadowSolver) StationStatesBetween(
	start, end time.Time, step time.Duration, lon, lat, height float64,
) []SolarEclipseStationState {
	if step <= 0 || !end.After(start) {
		return nil
	}
	startTT := solver.ttJDE(start)
	endTT := solver.ttJDE(end)
	stepDays := step.Hours() / 24
	inner := solver.inner.StationStatesBetweenJDE(startTT, endTT, stepDays, lon, lat, height)
	result := make([]SolarEclipseStationState, 0, len(inner))
	// 同上：累加避免长窗口下的纳秒溢出。
	stateTime := start
	for _, state := range inner {
		result = append(result, solarEclipseStationStateFromBasic(stateTime, state))
		stateTime = stateTime.Add(step)
	}
	return result
}

// SolarEclipseShadowBetween 无状态批量入口，使用进程级 ΔT / stateless batch entry point.
func SolarEclipseShadowBetween(
	start, end time.Time, step time.Duration,
) []SolarEclipseShadowInstant {
	return SolarEclipseShadowBetweenWithOptions(start, end, step, SolarEclipseShadowSolverOptions{})
}

// SolarEclipseShadowBetweenWithOptions 可显式指定 ΔT 与模型 / with explicit ΔT and model.
func SolarEclipseShadowBetweenWithOptions(
	start, end time.Time, step time.Duration, options SolarEclipseShadowSolverOptions,
) []SolarEclipseShadowInstant {
	return NewSolarEclipseShadowSolver(options).ShadowBetween(start, end, step)
}
