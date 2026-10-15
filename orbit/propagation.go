package orbit

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
)

const (
	propagationAU        = 149597870.7
	propagationGaussianK = 0.01720209895
	propagationSunGM     = propagationGaussianK * propagationGaussianK * propagationAU * propagationAU * propagationAU
)

var (
	// ErrInvalidPropagationElements 根数不能用于传播 / orbital elements cannot be propagated.
	ErrInvalidPropagationElements = errors.New("orbit: invalid propagation elements")
	// ErrPerturbationEngine 摄动传播缺少天体星历 / perturbation propagation requires an ephemeris.
	ErrPerturbationEngine = errors.New("orbit: perturbation ephemeris is not configured")
	// ErrUnsupportedPerturber 摄动天体未知或重复计入 / perturbing body is unknown or counted twice.
	ErrUnsupportedPerturber = errors.New("orbit: unsupported perturbing body")
)

// maxPropagationNodes 单次传播允许的网格步数上限 / upper bound on grid steps of one propagation.
const maxPropagationNodes = 1 << 30

// Body NAIF 天体编号 / NAIF body identifier.
type Body = ephem.Body

// StateProvider 提供摄动天体状态 / supplies perturbing-body states.
type StateProvider = ephem.Engine

// PropagationOptions 可选摄动传播参数；零值保持纯二体 / optional perturbation propagation parameters; zero preserves two-body propagation.
type PropagationOptions struct {
	Perturbers []Body        // 摄动天体 / perturbing bodies.
	Engine     StateProvider // 摄动星历 / perturbation ephemeris.
	Context    StateProvider // 摄动星历别名 / alternate perturbation ephemeris.
	Step       float64       // 最大步长，单位日 / maximum step in days.
	Tolerance  float64       // 步长加倍的相对误差阈值 / relative step-doubling error threshold.
}

// State 计算日心 J2000 平黄道根数的二体状态，输出 ICRF km 与 km/日 / computes a two-body state from heliocentric J2000 ecliptic elements, returning ICRF km and km/day.
func State(elements Elements, jdTDB float64) (pos, vel [3]float64, err error) {
	return twoBodyState(elements, jdTDB)
}

// StateWithOptions 按选项传播根数；零值与 State 完全一致 / propagates elements with optional perturbations; zero options exactly match State.
func StateWithOptions(elements Elements, jdTDB float64, options PropagationOptions) (pos, vel [3]float64, err error) {
	if !finite(jdTDB) {
		return pos, vel, ErrInvalidPropagationElements
	}
	perturbers, err := normalizePerturbers(options.Perturbers)
	if err != nil {
		return pos, vel, err
	}
	if len(perturbers) == 0 {
		return twoBodyState(elements, jdTDB)
	}
	run, err := newPropagationRun(elements, perturbers, options)
	if err != nil {
		return pos, vel, err
	}
	return run.state(jdTDB, nil)
}

// propagationRun 固定一次传播的根数、天体与步长网格。
type propagationRun struct {
	elements   Elements
	perturbers []Body
	provider   StateProvider
	step       float64
	tolerance  float64
	epoch      float64
	start      propagationCheckpoint
}

type propagationCheckpoint struct {
	pos [3]float64
	vel [3]float64
}

// propagationCheckpointStride 每隔多少个网格点缓存一次状态。
const propagationCheckpointStride = 32

// propagationCheckpoints 引擎缓存的网格状态，下标 j 对应 epoch±j·stride·step。
type propagationCheckpoints struct {
	forward  []propagationCheckpoint
	backward []propagationCheckpoint
}

func newPropagationRun(elements Elements, perturbers []Body, options PropagationOptions) (*propagationRun, error) {
	provider := options.Engine
	if provider == nil {
		provider = options.Context
	}
	if provider == nil {
		return nil, ErrPerturbationEngine
	}
	epoch := propagationEpoch(elements)
	// 经典根数漏填 EpochJD 时为 0，会从 JD 0 起积分数千年，按非法根数拒绝。
	if !finite(epoch) || epoch == 0 {
		return nil, ErrInvalidPropagationElements
	}
	pos, vel, err := twoBodyState(elements, epoch)
	if err != nil {
		return nil, err
	}
	step := options.Step
	if !(step > 0) || !finite(step) {
		step = 0.5
	}
	tolerance := options.Tolerance
	if !(tolerance > 0) || !finite(tolerance) {
		tolerance = 1e-6
	}
	return &propagationRun{
		elements:   elements,
		perturbers: perturbers,
		provider:   provider,
		step:       step,
		tolerance:  tolerance,
		epoch:      epoch,
		start:      propagationCheckpoint{pos: pos, vel: vel},
	}, nil
}

// propagationEpoch 积分起点：近日点形式给出 EpochJD 时取其为密切历元，否则用 TpJD / integration start: the osculating EpochJD when a perihelion-form set provides one, otherwise TpJD.
func propagationEpoch(elements Elements) float64 {
	if usesPerihelionForm(elements) && (elements.EpochJD == 0 || !finite(elements.EpochJD)) {
		return elements.TpJD
	}
	return elements.EpochJD
}

// state 先沿固定网格积分到不越过目标的最后一个格点，再积分余量；结果只取决于目标历元，与 cache 内容无关 / integrates along a fixed grid to the last node short of the target, then the remainder; the result depends only on the target, never on cache contents.
func (run *propagationRun) state(target float64, cache *propagationCheckpoints) (pos, vel [3]float64, err error) {
	if target == run.epoch {
		return run.start.pos, run.start.vel, nil
	}
	direction := 1.0
	if target < run.epoch {
		direction = -1
	}
	span := math.Abs(target - run.epoch)
	count := math.Floor(span / run.step)
	if count > maxPropagationNodes {
		return pos, vel, ErrInvalidPropagationElements
	}
	index := int(count)
	for index > 0 && math.Abs(run.node(index, direction)-run.epoch) > span {
		index--
	}
	current, known := run.start, 0
	var stored *[]propagationCheckpoint
	if cache != nil {
		stored = &cache.forward
		if direction < 0 {
			stored = &cache.backward
		}
		if len(*stored) == 0 {
			*stored = append(*stored, run.start)
		}
		slot := index / propagationCheckpointStride
		if slot > len(*stored)-1 {
			slot = len(*stored) - 1
		}
		known = slot * propagationCheckpointStride
		current = (*stored)[slot]
	}
	session := newPropagationSession(run.provider)
	for k := known; k < index; k++ {
		current.pos, current.vel, err = run.integrate(current.pos, current.vel, run.node(k, direction), run.node(k+1, direction), session)
		if err != nil {
			return current.pos, current.vel, err
		}
		if stored != nil && (k+1)%propagationCheckpointStride == 0 && len(*stored) == (k+1)/propagationCheckpointStride {
			*stored = append(*stored, current)
		}
	}
	from := run.node(index, direction)
	if from == target {
		return current.pos, current.vel, nil
	}
	return run.integrate(current.pos, current.vel, from, target, session)
}

func (run *propagationRun) node(index int, direction float64) float64 {
	return run.epoch + direction*float64(index)*run.step
}

// integrate 自适应子步积分 [from, to]，终点精确落在 to / adaptively integrates [from, to], landing exactly on to.
func (run *propagationRun) integrate(pos, vel [3]float64, from, to float64, session *propagationSession) (outPos, outVel [3]float64, err error) {
	outPos, outVel = pos, vel
	current := from
	for current != to {
		dt := to - current
		var used float64
		outPos, outVel, used, err = propagationAdaptiveStep(outPos, outVel, current, dt, run.elements, run.perturbers, session, run.tolerance)
		if err != nil {
			return outPos, outVel, err
		}
		if used == dt {
			current = to
		} else {
			current += used
		}
	}
	return outPos, outVel, nil
}

func propagationAdaptiveStep(pos, vel [3]float64, jd, dt float64, elements Elements, perturbers []Body, session *propagationSession, tolerance float64) (outPos, outVel [3]float64, used float64, err error) {
	trial := dt
	lastError := math.Inf(1)
	for attempts := 0; attempts < 20; attempts++ {
		fullPos, fullVel, fullErr := propagationRK4(pos, vel, jd, trial, elements, perturbers, session)
		if fullErr != nil {
			return pos, vel, trial, fullErr
		}
		halfPos, halfVel, halfErr := propagationRK4(pos, vel, jd, trial/2, elements, perturbers, session)
		if halfErr != nil {
			return pos, vel, trial, halfErr
		}
		halfPos, halfVel, halfErr = propagationRK4(halfPos, halfVel, jd+trial/2, trial/2, elements, perturbers, session)
		if halfErr != nil {
			return pos, vel, trial, halfErr
		}
		error := 0.0
		for index := range fullPos {
			positionScale := math.Max(1, math.Max(math.Abs(fullPos[index]), math.Abs(halfPos[index])))
			velocityScale := math.Max(1, math.Max(math.Abs(fullVel[index]), math.Abs(halfVel[index])))
			error = math.Max(error, math.Abs(fullPos[index]-halfPos[index])/(15*positionScale))
			error = math.Max(error, math.Abs(fullVel[index]-halfVel[index])/(15*velocityScale))
		}
		lastError = error
		if error <= tolerance || math.Abs(trial) <= 1e-8 {
			return halfPos, halfVel, trial, nil
		}
		trial /= 2
	}
	return pos, vel, trial, fmt.Errorf("orbit: propagation step did not converge (dt=%g error=%g tolerance=%g)", trial, lastError, tolerance)
}

type propagationSession struct {
	provider StateProvider
	states   map[propagationStateKey]propagationSample
}

type propagationStateKey struct {
	body Body
	jd   uint64
}

type propagationSample struct {
	pos [3]float64
	vel [3]float64
	err error
}

func newPropagationSession(provider StateProvider) *propagationSession {
	return &propagationSession{
		provider: provider,
		states:   make(map[propagationStateKey]propagationSample),
	}
}

func (session *propagationSession) state(body Body, jd float64) (pos, vel [3]float64, err error) {
	key := propagationStateKey{body: body, jd: math.Float64bits(jd)}
	if sample, ok := session.states[key]; ok {
		return sample.pos, sample.vel, sample.err
	}
	pos, vel, err = session.provider.State(body, ephem.SSB, jd)
	session.states[key] = propagationSample{pos: pos, vel: vel, err: err}
	return pos, vel, err
}

// EphemerisEngine 根数传播的只读星历引擎，结果与 StateWithOptions 逐位一致且不受查询顺序影响 / read-only engine over propagated elements.
type EphemerisEngine struct {
	target   Body
	elements Elements
	options  PropagationOptions
	setupErr error
	run      *propagationRun
	mu       sync.Mutex
	cache    propagationCheckpoints
}

// NewEphemerisEngine 创建根数传播星历引擎；选项无效时 State 返回对应错误 / creates an ephemeris engine backed by propagated elements; invalid options surface from State.
func NewEphemerisEngine(target Body, elements Elements, options PropagationOptions) *EphemerisEngine {
	engine := &EphemerisEngine{target: target, elements: elements}
	options.Perturbers, engine.setupErr = normalizePerturbers(options.Perturbers)
	engine.options = options
	if engine.setupErr == nil && len(options.Perturbers) != 0 {
		engine.run, engine.setupErr = newPropagationRun(elements, options.Perturbers, options)
	}
	return engine
}

// Target 返回传播目标 / returns the propagated target.
func (engine *EphemerisEngine) Target() Body { return engine.target }

// Elements 返回传播根数 / returns the propagation elements.
func (engine *EphemerisEngine) Elements() Elements { return engine.elements }

// State 返回传播目标相对指定中心的状态 / returns propagated target state relative to a center.
func (engine *EphemerisEngine) State(target, center Body, jdTDB float64) (pos, vel [3]float64, err error) {
	if engine == nil || target != engine.target {
		if target == center {
			return pos, vel, nil
		}
		return pos, vel, ephem.ErrUnsupportedBody
	}
	if target == center {
		return pos, vel, nil
	}
	if center != ephem.Sun && center != ephem.SSB {
		return pos, vel, ephem.ErrUnsupportedBody
	}
	if engine.setupErr != nil {
		return pos, vel, engine.setupErr
	}
	if !finite(jdTDB) {
		return pos, vel, ErrInvalidPropagationElements
	}
	if engine.run == nil {
		pos, vel, err = twoBodyState(engine.elements, jdTDB)
	} else {
		engine.mu.Lock()
		pos, vel, err = engine.run.state(jdTDB, &engine.cache)
		engine.mu.Unlock()
	}
	if err != nil {
		return pos, vel, err
	}
	if center == ephem.Sun {
		return pos, vel, nil
	}
	provider := engine.options.Engine
	if provider == nil {
		provider = engine.options.Context
	}
	if provider == nil {
		return pos, vel, ErrPerturbationEngine
	}
	sunPos, sunVel, sunErr := provider.State(ephem.Sun, ephem.SSB, jdTDB)
	if sunErr != nil {
		return pos, vel, sunErr
	}
	for index := range pos {
		pos[index] += sunPos[index]
		vel[index] += sunVel[index]
	}
	return pos, vel, nil
}

// Coverage 返回传播引擎的数学覆盖范围 / returns the mathematical coverage of the propagator.
func (engine *EphemerisEngine) Coverage(target Body) (start, end float64, ok bool) {
	if engine == nil || target != engine.target {
		return 0, 0, false
	}
	return math.Inf(-1), math.Inf(1), true
}

func propagationRK4(pos, vel [3]float64, jd, dt float64, elements Elements, perturbers []Body, session *propagationSession) ([3]float64, [3]float64, error) {
	p1, v1, err := propagationDerivative(pos, vel, jd, elements, perturbers, session)
	if err != nil {
		return pos, vel, err
	}
	p2 := vectorAdd(pos, vectorScale(p1, dt/2))
	v2 := vectorAdd(vel, vectorScale(v1, dt/2))
	p2d, v2d, err := propagationDerivative(p2, v2, jd+dt/2, elements, perturbers, session)
	if err != nil {
		return pos, vel, err
	}
	p3 := vectorAdd(pos, vectorScale(p2d, dt/2))
	v3 := vectorAdd(vel, vectorScale(v2d, dt/2))
	p3d, v3d, err := propagationDerivative(p3, v3, jd+dt/2, elements, perturbers, session)
	if err != nil {
		return pos, vel, err
	}
	p4 := vectorAdd(pos, vectorScale(p3d, dt))
	v4 := vectorAdd(vel, vectorScale(v3d, dt))
	p4d, v4d, err := propagationDerivative(p4, v4, jd+dt, elements, perturbers, session)
	if err != nil {
		return pos, vel, err
	}
	return vectorAdd(pos, vectorScale(vectorAdd(vectorAdd(p1, vectorScale(p2d, 2)), vectorAdd(vectorScale(p3d, 2), p4d)), dt/6)), vectorAdd(vel, vectorScale(vectorAdd(vectorAdd(v1, vectorScale(v2d, 2)), vectorAdd(vectorScale(v3d, 2), v4d)), dt/6)), nil
}

func propagationDerivative(pos, vel [3]float64, jd float64, elements Elements, perturbers []Body, session *propagationSession) ([3]float64, [3]float64, error) {
	radius := vectorNorm(pos)
	if !(radius > 0) || !finite(radius) {
		return vel, [3]float64{}, ErrInvalidPropagationElements
	}
	acc := vectorScale(pos, -propagationSunGM/math.Pow(radius, 3))
	sunPos, _, err := session.state(ephem.Sun, jd)
	if err != nil {
		return vel, acc, err
	}
	for _, body := range perturbers {
		if body == ephem.Sun || body == ephem.SSB {
			continue
		}
		bodyPos, _, bodyErr := session.state(body, jd)
		if bodyErr != nil {
			return vel, acc, bodyErr
		}
		relBody := vectorSub(bodyPos, sunPos)
		toBody := vectorSub(relBody, pos)
		mu := perturbingGM(body)
		if vectorNorm(relBody) == 0 || vectorNorm(toBody) == 0 {
			return vel, acc, ErrInvalidPropagationElements
		}
		acc = vectorAdd(acc, vectorScale(toBody, mu/math.Pow(vectorNorm(toBody), 3)))
		acc = vectorAdd(acc, vectorScale(relBody, -mu/math.Pow(vectorNorm(relBody), 3)))
	}
	return vel, acc, nil
}

func perturbingGM(body Body) float64 {
	const secondPerDay = 86400.0
	var gm float64
	switch body {
	case ephem.Mercury:
		gm = 22031.86855
	case ephem.Venus:
		gm = 324858.592
	case ephem.Earth:
		gm = 398600.435436
	case ephem.Moon:
		gm = 4902.800066
	case ephem.EMB:
		gm = 398600.435436 + 4902.800066
	case ephem.Mars:
		gm = 42828.375214
	case ephem.Jupiter:
		gm = 126686534.911
	case ephem.Saturn:
		gm = 37931207.8
	case ephem.Uranus:
		gm = 5793966.0
	case ephem.Neptune:
		gm = 6835100.0
	}
	return gm * secondPerDay * secondPerDay
}

// normalizePerturbers 去重并跳过太阳与 SSB；未知天体或 EMB 与地球/月球同时出现时报错 / dedups and skips Sun and SSB; rejects unknown bodies and EMB combined with Earth or Moon.
func normalizePerturbers(perturbers []Body) ([]Body, error) {
	if len(perturbers) == 0 {
		return nil, nil
	}
	seen := make(map[Body]struct{}, len(perturbers))
	result := make([]Body, 0, len(perturbers))
	for _, body := range perturbers {
		if body == ephem.Sun || body == ephem.SSB {
			continue
		}
		if perturbingGM(body) == 0 {
			return nil, fmt.Errorf("%w: %d", ErrUnsupportedPerturber, body)
		}
		if _, ok := seen[body]; ok {
			continue
		}
		seen[body] = struct{}{}
		result = append(result, body)
	}
	_, hasEMB := seen[ephem.EMB]
	_, hasEarth := seen[ephem.Earth]
	_, hasMoon := seen[ephem.Moon]
	if hasEMB && (hasEarth || hasMoon) {
		return nil, fmt.Errorf("%w: EMB already includes Earth and Moon", ErrUnsupportedPerturber)
	}
	return result, nil
}

func twoBodyState(elements Elements, jd float64) (pos, vel [3]float64, err error) {
	if !finite(jd) {
		return pos, vel, ErrInvalidPropagationElements
	}
	basicElements := toBasicElements(elements)
	trueAnomaly := basic.OrbitTrueAnomaly(jd, basicElements)
	if math.IsNaN(trueAnomaly) {
		return pos, vel, ErrInvalidPropagationElements
	}
	resolved := elementsAt(elements, jd)
	eccentricity := resolved.E
	semiLatus := resolved.A * (1 - eccentricity*eccentricity)
	if usesPerihelionForm(elements) {
		semiLatus = resolved.Q * (1 + eccentricity)
	}
	if !(semiLatus > 0) || !finite(semiLatus) {
		return pos, vel, ErrInvalidPropagationElements
	}
	nu := trueAnomaly * math.Pi / 180
	mu := propagationGaussianK * propagationGaussianK
	radius := semiLatus / (1 + eccentricity*math.Cos(nu))
	if !(radius > 0) || !finite(radius) {
		return pos, vel, ErrInvalidPropagationElements
	}
	x := radius * math.Cos(nu)
	y := radius * math.Sin(nu)
	speedScale := math.Sqrt(mu / semiLatus)
	vx := -speedScale * math.Sin(nu)
	vy := speedScale * (eccentricity + math.Cos(nu))
	posAU, velAU := rotatePerifocal(x, y, vx, vy, resolved)
	return rotateEquatorial(posAU, propagationAU), rotateEquatorial(velAU, propagationAU), nil
}

func elementsAt(elements Elements, jd float64) Elements {
	if usesPerihelionForm(elements) || !finite(elements.EpochJD) {
		return elements
	}
	delta := jd - elements.EpochJD
	elements.A += elements.ADot * delta
	elements.E += elements.EDot * delta
	elements.I += elements.IDot * delta
	elements.Omega += elements.OmegaDot * delta
	elements.W += elements.WDot * delta
	return elements
}

func usesPerihelionForm(elements Elements) bool { return elements.Q > 0 && finite(elements.TpJD) }

func rotatePerifocal(x, y, vx, vy float64, elements Elements) ([3]float64, [3]float64) {
	omega := elements.Omega * math.Pi / 180
	argument := elements.W * math.Pi / 180
	inclination := elements.I * math.Pi / 180
	sinOmega, cosOmega := math.Sincos(omega)
	sinArgument, cosArgument := math.Sincos(argument)
	sinInclination, cosInclination := math.Sincos(inclination)
	basisX := [3]float64{cosOmega*cosArgument - sinOmega*sinArgument*cosInclination, sinOmega*cosArgument + cosOmega*sinArgument*cosInclination, sinArgument * sinInclination}
	basisY := [3]float64{-cosOmega*sinArgument - sinOmega*cosArgument*cosInclination, -sinOmega*sinArgument + cosOmega*cosArgument*cosInclination, cosArgument * sinInclination}
	return vectorAdd(vectorScale(basisX, x), vectorScale(basisY, y)), vectorAdd(vectorScale(basisX, vx), vectorScale(basisY, vy))
}

var sinJ2000Obliquity, cosJ2000Obliquity = math.Sincos(basic.EclipticObliquity(2451545.0, false) * math.Pi / 180)

func rotateEquatorial(value [3]float64, scale float64) [3]float64 {
	sinEps, cosEps := sinJ2000Obliquity, cosJ2000Obliquity
	return [3]float64{value[0] * scale, (value[1]*cosEps - value[2]*sinEps) * scale, (value[1]*sinEps + value[2]*cosEps) * scale}
}

func vectorAdd(left, right [3]float64) [3]float64 {
	return [3]float64{left[0] + right[0], left[1] + right[1], left[2] + right[2]}
}

func vectorSub(left, right [3]float64) [3]float64 {
	return [3]float64{left[0] - right[0], left[1] - right[1], left[2] - right[2]}
}

func vectorScale(value [3]float64, scale float64) [3]float64 {
	return [3]float64{value[0] * scale, value[1] * scale, value[2] * scale}
}

func vectorNorm(value [3]float64) float64 {
	return math.Sqrt(value[0]*value[0] + value[1]*value[1] + value[2]*value[2])
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

var _ StateProvider = (*EphemerisEngine)(nil)
