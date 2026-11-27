package basic

import (
	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/tools"
	"errors"
	"math"
	"sort"
)

// SurfacePosition 无折射的参考层站心视位置 / airless apparent position on a planetary reference layer.
type SurfacePosition struct {
	// Altitude 目标中心高度，度 / altitude of the target center in degrees.
	Altitude float64
	// Azimuth 正北起顺时针方位角，度 / azimuth clockwise from north in degrees.
	Azimuth float64
	// HourAngle 随自转推进的时角，度，范围 (−180,180] / rotation-directed hour angle in degrees within (−180,180].
	HourAngle float64
	// BodyRA 行星赤道升交点零点赤经，度 / planetary equatorial right ascension from the ascending node in degrees.
	BodyRA float64
	// BodyDec 行星赤道赤纬，度 / planetary equatorial declination in degrees.
	BodyDec float64
	// DistanceAU 几何距离，AU，恒星距离未知时为 0 / geometric distance in AU, zero for stars without distance.
	DistanceAU float64
	// AngularRadius 太阳视半径，度，其他目标为 0 / solar angular radius in degrees, zero for other targets.
	AngularRadius float64
}

func surfacePosition(state surfaceDirection, observer BodyObserver, solar bool) SurfacePosition {
	d, f := state.direction, state.frame
	equator := bodyEquatorialNodeFrame(f.orientation)
	ra := tools.Limit360(math.Atan2(vectorDot(d, equator.east), vectorDot(d, equator.node)) * 180 / math.Pi)
	longitude := tools.Limit360(observer.EastLongitude)
	h := tools.Limit360(math.Copysign(1, f.orientation.RotationRate) * (f.orientation.PrimeMeridianAngle + longitude - ra))
	if h > 180 {
		h -= 360
	}
	result := SurfacePosition{
		Altitude:  math.Atan2(vectorDot(d, f.up), math.Hypot(vectorDot(d, f.east), vectorDot(d, f.north))) * 180 / math.Pi,
		Azimuth:   tools.Limit360(math.Atan2(vectorDot(d, f.east), vectorDot(d, f.north)) * 180 / math.Pi),
		HourAngle: h, BodyRA: ra,
		BodyDec:    math.Atan2(vectorDot(d, equator.pole), math.Hypot(vectorDot(d, equator.node), vectorDot(d, equator.east))) * 180 / math.Pi,
		DistanceAU: state.distanceKM / astronomicalUnitKM,
	}
	if solar {
		result.AngularRadius = 695700 / state.distanceKM * 180 / math.Pi
	}
	return result
}

// SurfacePositionAtWithInstance 行星表面无折射站心视位置，时刻为 TT 儒略日 / airless planetary surface apparent position at a TT Julian day.
func SurfacePositionAtWithInstance(target SkyBody, observer BodyObserver, jdTDB float64, inst *Instance) (SurfacePosition, error) {
	state, err := surfaceDirectionAt(target, observer, jdTDB, inst)
	if err != nil {
		return SurfacePosition{}, err
	}
	return surfacePosition(state, observer, target == SkySun), nil
}

// StarSurfacePositionAtWithInstance 行星表面的恒星无折射视位置，时刻为 TT 儒略日 / airless stellar apparent position on a planet at a TT Julian day.
func StarSurfacePositionAtWithInstance(star StarCoordinate, observer BodyObserver, jdTDB float64, inst *Instance) (SurfacePosition, error) {
	state, err := surfaceStarDirectionAt(star, observer, jdTDB, inst)
	if err != nil {
		return SurfacePosition{}, err
	}
	return surfacePosition(state, observer, false), nil
}

// SurfaceOptions 无折射高度事件选项，零值取太阳上缘与零度地平 / airless altitude event options, defaulting to the solar upper limb at zero degrees.
type SurfaceOptions struct {
	// Limb 太阳盘面参考点，其他目标忽略 / solar limb, ignored for other targets.
	Limb AltitudeLimb
	// TargetAltitude 参考点目标高度，度 / target altitude of the reference point in degrees.
	TargetAltitude float64
}

// SurfaceEventKind 行星表面瞬时事件种类 / kind of instantaneous planetary surface event.
type SurfaceEventKind int

const (
	// SurfaceRise 参考点升过目标高度 / reference point rising through the target altitude.
	SurfaceRise SurfaceEventKind = iota
	// SurfaceSet 参考点落过目标高度 / reference point setting through the target altitude.
	SurfaceSet
	// SurfaceUpperCulmination 上中天 / upper culmination.
	SurfaceUpperCulmination
	// SurfaceLowerCulmination 下中天 / lower culmination.
	SurfaceLowerCulmination
)

// SurfaceEvent 行星参考层上的无折射事件，时刻为 TT 儒略日 / airless planetary reference-layer event at a TT Julian day.
type SurfaceEvent struct {
	Kind     SurfaceEventKind
	JDE      float64
	Position SurfacePosition
}

type surfaceSampler func(float64) (SurfacePosition, error)
type surfaceGapFinder func(float64, float64) (float64, bool)
type surfaceSearchConfig struct {
	rotationDays, solarDays, latitude float64
	options                           SurfaceOptions
	steps                             int
	disableSparse                     bool
	envelope                          bool
	onlyKind                          bool
	kind                              SurfaceEventKind
}
type surfaceSearchPoint struct {
	jd       float64
	position SurfacePosition
	residual float64
}

func surfaceResidual(p SurfacePosition, options SurfaceOptions) float64 {
	offset := 0.0
	if options.Limb == AltitudeLimbUpper {
		offset = p.AngularRadius
	} else if options.Limb == AltitudeLimbLower {
		offset = -p.AngularRadius
	}
	return p.Altitude + offset - options.TargetAltitude
}
func surfacePointAt(sample surfaceSampler, jd float64, options SurfaceOptions) (surfaceSearchPoint, error) {
	p, err := sample(jd)
	if err != nil {
		return surfaceSearchPoint{}, err
	}
	for _, v := range []float64{p.Altitude, p.HourAngle, p.BodyDec, p.AngularRadius} {
		if !isFiniteFloat(v) {
			return surfaceSearchPoint{}, ErrInvalidObservationInput
		}
	}
	return surfaceSearchPoint{jd, p, surfaceResidual(p, options)}, nil
}

func surfaceCulminationResidual(hourAngle float64) float64 {
	if math.Mod(hourAngle, 180) == 0 {
		return 0
	}
	return math.Sin(hourAngle * math.Pi / 180)
}

func surfaceRefineRoot(sample surfaceSampler, options SurfaceOptions, a, b surfaceSearchPoint, culmination bool) (SurfaceEvent, error) {
	value := func(p surfaceSearchPoint) float64 {
		if culmination {
			return surfaceCulminationResidual(p.position.HourAngle)
		}
		return p.residual
	}
	fa, fb := value(a), value(b)
	for i := 0; i < 64 && b.jd-a.jd > 0.1/86400; i++ {
		jd := (a.jd + b.jd) / 2
		secant := (a.jd*fb - b.jd*fa) / (fb - fa)
		if secant > a.jd+0.1*(b.jd-a.jd) && secant < b.jd-0.1*(b.jd-a.jd) {
			jd = secant
		}
		p, err := surfacePointAt(sample, jd, options)
		if err != nil {
			return SurfaceEvent{}, err
		}
		f := value(p)
		if fa*f <= 0 {
			b, fb = p, f
		} else {
			a, fa = p, f
		}
	}
	jd := (a.jd + b.jd) / 2
	if math.Abs(fb-fa) > 0 {
		jd = a.jd - fa*(b.jd-a.jd)/(fb-fa)
	}
	if jd < a.jd || jd > b.jd {
		jd = (a.jd + b.jd) / 2
	}
	p, err := sample(jd)
	if err != nil {
		return SurfaceEvent{}, err
	}
	kind := SurfaceRise
	if culmination {
		kind = SurfaceUpperCulmination
		if math.Cos(p.HourAngle*math.Pi/180) < 0 {
			kind = SurfaceLowerCulmination
		}
	} else if a.residual > b.residual {
		kind = SurfaceSet
	}
	return SurfaceEvent{Kind: kind, JDE: jd, Position: p}, nil
}
func surfaceExtremum(sample surfaceSampler, options SurfaceOptions, a, b float64, maximize bool) (surfaceSearchPoint, error) {
	const ratio = 0.6180339887498948482
	x, y := b-ratio*(b-a), a+ratio*(b-a)
	p, err := surfacePointAt(sample, x, options)
	if err != nil {
		return p, err
	}
	q, err := surfacePointAt(sample, y, options)
	if err != nil {
		return q, err
	}
	for i := 0; i < 64 && b-a > 0.1/86400; i++ {
		chooseLeft := p.residual < q.residual
		if maximize {
			chooseLeft = p.residual > q.residual
		}
		if chooseLeft {
			b, y, q = y, x, p
			x = b - ratio*(b-a)
			p, err = surfacePointAt(sample, x, options)
		} else {
			a, x, p = x, y, q
			y = a + ratio*(b-a)
			q, err = surfacePointAt(sample, y, options)
		}
		if err != nil {
			return surfaceSearchPoint{}, err
		}
	}
	return surfacePointAt(sample, (a+b)/2, options)
}
func surfaceEnvelope(a, b surfaceSearchPoint, c surfaceSearchConfig) bool {
	margin := 1 + 2*math.Abs(b.position.BodyDec-a.position.BodyDec)
	for _, p := range []surfaceSearchPoint{a, b} {
		maxAlt := 90 - math.Abs(c.latitude-p.position.BodyDec)
		minAlt := math.Abs(c.latitude+p.position.BodyDec) - 90
		h0 := p.position.Altitude - p.residual
		if !(minAlt > h0+margin || maxAlt < h0-margin) {
			return false
		}
	}
	return a.residual*b.residual > 0
}
func surfaceSearch(sample surfaceSampler, start, end float64, c surfaceSearchConfig) ([]SurfaceEvent, AltitudeCrossingState, error) {
	events, state, err := surfaceSearchGrid(sample, start, end, c, false)
	if errors.Is(err, ephem.ErrOutOfRange) {
		return surfaceSearchGrid(sample, start, end, c, true)
	}
	return events, state, err
}

func (c surfaceSearchConfig) grid() (segment float64, steps int) {
	segment, steps = c.rotationDays, c.steps
	if steps <= 0 {
		steps = 48
		if c.rotationDays >= 3 {
			steps = 96
		}
	}
	if c.rotationDays >= 3 {
		segment = math.Min(c.rotationDays, c.solarDays)
	}
	return segment, steps
}

type surfaceGridSearcher struct {
	sample       surfaceSampler
	start, end   float64
	c            surfaceSearchConfig
	fast         bool
	segment      float64
	steps        int
	sampleStart  float64
	sampleEnd    float64
	events       []SurfaceEvent
	minValue     float64
	maxValue     float64
	previous     surfaceSearchPoint
	havePrevious bool
	tangent      bool
}

func (s *surfaceGridSearcher) update(p surfaceSearchPoint) {
	if p.jd >= s.start && p.jd <= s.end {
		s.minValue = math.Min(s.minValue, p.residual)
		s.maxValue = math.Max(s.maxValue, p.residual)
	}
}

func (s *surfaceGridSearcher) recordEvent(e SurfaceEvent) {
	if e.JDE >= s.start && e.JDE <= s.end && (!s.c.onlyKind || e.Kind == s.c.kind) {
		s.events = append(s.events, e)
	}
}

func (s *surfaceGridSearcher) add(a, b surfaceSearchPoint, culmination bool) error {
	event, err := surfaceRefineRoot(s.sample, s.c.options, a, b, culmination)
	if err != nil {
		return err
	}
	s.recordEvent(event)
	return nil
}

func (s *surfaceGridSearcher) checkAltitudeCrossing(a, b surfaceSearchPoint, inPair, lastStep bool) error {
	if inPair {
		crossing := a.residual*b.residual < 0 ||
			(a.residual == 0 && b.residual != 0 && !s.havePrevious) ||
			(b.residual == 0 && a.residual != 0 && lastStep && b.jd == s.sampleEnd)
		if crossing {
			if err := s.add(a, b, false); err != nil {
				return err
			}
		}
	}
	if s.havePrevious && a.residual == 0 && b.jd >= s.start && s.previous.jd <= s.end {
		if s.previous.residual*b.residual < 0 {
			return s.add(s.previous, b, false)
		}
		if a.jd >= s.start && a.jd <= s.end {
			s.tangent = true
		}
	}
	return nil
}

func (s *surfaceGridSearcher) checkCulmination(a, b surfaceSearchPoint) error {
	ha, hb := surfaceCulminationResidual(a.position.HourAngle), surfaceCulminationResidual(b.position.HourAngle)
	if ha*hb <= 0 && math.Abs(ha-hb) > 0 {
		return s.add(a, b, true)
	}
	return nil
}

func (s *surfaceGridSearcher) checkExtremum(a, b surfaceSearchPoint) error {
	d1, d2 := a.residual-s.previous.residual, b.residual-a.residual
	if !(d1*d2 < 0 || (d1 != 0 && d2 == 0) || (d1 == 0 && d2 != 0)) {
		return nil
	}
	maximize := d1 > 0
	if d1 == 0 {
		maximize = d2 < 0
	}
	extreme, err := surfaceExtremum(s.sample, s.c.options, s.previous.jd, b.jd, maximize)
	if err != nil {
		return err
	}
	s.update(extreme)
	if s.previous.residual*extreme.residual < 0 {
		if err = s.add(s.previous, extreme, false); err != nil {
			return err
		}
	}
	if extreme.residual*b.residual < 0 {
		if err = s.add(extreme, b, false); err != nil {
			return err
		}
	}
	if math.Abs(extreme.residual) < altitudeTangentDegrees && extreme.jd >= s.start && extreme.jd <= s.end {
		s.tangent = true
	}
	return nil
}

func (s *surfaceGridSearcher) scanSteps(gridLeft, leftJD, rightJD float64, clipped, skipAltitude bool, a, z surfaceSearchPoint) error {
	checkCulm := !s.c.onlyKind || s.c.kind >= SurfaceUpperCulmination
	for step := 1; step <= s.steps; step++ {
		b := z
		jd := gridLeft + float64(step)*s.segment/float64(s.steps)
		if clipped && jd <= leftJD && step < s.steps {
			continue
		}
		lastStep := step == s.steps || (clipped && jd >= rightJD)
		if !lastStep {
			var err error
			b, err = surfacePointAt(s.sample, jd, s.c.options)
			if err != nil {
				return err
			}
		}
		s.update(b)
		inPair := b.jd >= s.start && a.jd <= s.end
		if !skipAltitude {
			if err := s.checkAltitudeCrossing(a, b, inPair, lastStep); err != nil {
				return err
			}
		}
		if checkCulm && inPair {
			if err := s.checkCulmination(a, b); err != nil {
				return err
			}
		}
		if !skipAltitude && s.havePrevious && b.jd >= s.start && s.previous.jd <= s.end {
			if err := s.checkExtremum(a, b); err != nil {
				return err
			}
		}
		s.previous, a, s.havePrevious = a, b, true
		if lastStep {
			break
		}
	}
	return nil
}

func (s *surfaceGridSearcher) scanSegment(index, stepDays float64) error {
	gridLeft := 2451545 + index*s.segment
	gridRight := 2451545 + (index+1)*s.segment
	if gridRight+stepDays < s.start || gridLeft-stepDays > s.end {
		s.havePrevious = false
		return nil
	}
	leftJD, rightJD := math.Max(gridLeft, s.sampleStart), math.Min(gridRight, s.sampleEnd)
	if leftJD >= rightJD {
		return nil
	}
	clipped := leftJD != gridLeft || rightJD != gridRight
	a, err := surfacePointAt(s.sample, leftJD, s.c.options)
	if err != nil {
		return err
	}
	z, err := surfacePointAt(s.sample, rightJD, s.c.options)
	if err != nil {
		return err
	}
	skipAltitude := s.fast && s.c.envelope && surfaceEnvelope(a, z, s.c)
	if skipAltitude && s.c.onlyKind && s.c.kind <= SurfaceSet {
		s.update(a)
		s.update(z)
		s.havePrevious = false
		return nil
	}
	s.update(a)
	if s.fast && skipAltitude && !s.c.disableSparse && !clipped {
		if rightJD < s.start || leftJD > s.end {
			s.havePrevious = false
			return nil
		}
		ok, err := surfaceSparseCulminations(s.sample, s.c, a, z, s.recordEvent, s.update)
		if err != nil {
			return err
		}
		if ok {
			s.havePrevious = false
			return nil
		}
	}
	return s.scanSteps(gridLeft, leftJD, rightJD, clipped, skipAltitude, a, z)
}

func (s *surfaceGridSearcher) finish() ([]SurfaceEvent, AltitudeCrossingState) {
	sort.Slice(s.events, func(i, j int) bool {
		if math.Abs(s.events[i].JDE-s.events[j].JDE) > 0 {
			return s.events[i].JDE < s.events[j].JDE
		}
		return s.events[i].Kind < s.events[j].Kind
	})
	unique := s.events[:0]
	for _, e := range s.events {
		duplicate := false
		for k := len(unique) - 1; k >= 0 && e.JDE-unique[k].JDE <= 0.1/86400; k-- {
			if e.Kind == unique[k].Kind {
				duplicate = true
				break
			}
		}
		if !duplicate {
			unique = append(unique, e)
		}
	}
	crossing := false
	for _, e := range unique {
		if e.Kind <= SurfaceSet {
			crossing = true
			break
		}
	}
	state := AltitudeAlwaysBelow
	if crossing {
		state = AltitudeCrosses
	} else if s.tangent {
		state = AltitudeTangent
	} else if s.minValue > 0 {
		state = AltitudeAlwaysAbove
	} else if s.minValue == 0 && s.maxValue == 0 {
		state = AltitudeTangent
	} else if s.maxValue >= 0 {
		state = AltitudeCrosses
	}
	return unique, state
}

func surfaceSearchGrid(sample surfaceSampler, start, end float64, c surfaceSearchConfig, bounded bool) ([]SurfaceEvent, AltitudeCrossingState, error) {
	segment, steps := c.grid()
	first := math.Floor((start-2451545)/segment) - 1
	last := math.Floor((end-2451545)/segment) + 1
	searcher := surfaceGridSearcher{
		sample:      sample,
		start:       start,
		end:         end,
		c:           c,
		fast:        c.rotationDays < 3,
		segment:     segment,
		steps:       steps,
		sampleStart: 2451545 + first*segment,
		sampleEnd:   2451545 + (last+1)*segment,
		events:      make([]SurfaceEvent, 0),
		minValue:    math.Inf(1),
		maxValue:    math.Inf(-1),
	}
	for _, jd := range []float64{start, end} {
		p, err := surfacePointAt(sample, jd, c.options)
		if err != nil {
			return nil, 0, err
		}
		searcher.update(p)
	}
	if bounded {
		var err error
		searcher.sampleStart, err = surfaceCoveredLimit(sample, start, searcher.sampleStart, segment, steps)
		if err != nil {
			return nil, 0, err
		}
		searcher.sampleEnd, err = surfaceCoveredLimit(sample, end, searcher.sampleEnd, segment, steps)
		if err != nil {
			return nil, 0, err
		}
	}
	stepDays := segment / float64(steps)
	for index := first; index <= last; index++ {
		if !(index+1 > index) {
			return nil, 0, ErrInvalidObservationInput
		}
		if err := searcher.scanSegment(index, stepDays); err != nil {
			return nil, 0, err
		}
	}
	unique, state := searcher.finish()
	return unique, state, nil
}

func surfaceConfigForBody(observer BodyObserver, options SurfaceOptions) surfaceSearchConfig {
	_, rate, _, _, _ := bodyOrientationModel(observer.Body)
	config := surfaceSearchConfig{rotationDays: 360 / math.Abs(rate), solarDays: 1 / math.Abs(rate/360-1/orbitalPeriodDays[observer.Body]), latitude: observer.Latitude, options: options, envelope: true, steps: 48}
	if config.rotationDays >= 3 {
		config.steps = 96
	}
	return config
}

// ErrSurfaceEventNotFound 搜索上限内未找到指定事件 / requested event was not found within the search horizon.
var ErrSurfaceEventNotFound = errors.New("surface event not found within search horizon")

func validateSurfaceQuery(observer BodyObserver, start, end float64, options SurfaceOptions) error {
	if _, err := validateBodyObserver(observer); err != nil {
		return err
	}
	const maxSurfaceJDDelta = 1e11
	if !isFiniteFloat(start) || !isFiniteFloat(end) || math.Abs(start-2451545) > maxSurfaceJDDelta || math.Abs(end-2451545) > maxSurfaceJDDelta || end < start || end-start > 10*365.25 || !isFiniteFloat(options.TargetAltitude) || math.Abs(options.TargetAltitude) > 90 || options.Limb < AltitudeLimbUpper || options.Limb > AltitudeLimbLower {
		return ErrInvalidObservationInput
	}
	return nil
}
func surfaceEventsForSampler(sample surfaceSampler, observer BodyObserver, start, end float64, options SurfaceOptions, findGap surfaceGapFinder) ([]SurfaceEvent, AltitudeCrossingState, error) {
	if err := validateSurfaceQuery(observer, start, end, options); err != nil {
		return nil, 0, err
	}
	if findGap != nil {
		if _, gap := findGap(start, end); gap {
			if _, err := sample(start); err != nil {
				return nil, 0, err
			}
			return nil, 0, ephem.ErrOutOfRange
		}
	}
	return surfaceSearch(sample, start, end, surfaceConfigForBody(observer, options))
}
func nextSurfaceForSampler(sample surfaceSampler, observer BodyObserver, kind SurfaceEventKind, after float64, options SurfaceOptions, findGap surfaceGapFinder) (SurfaceEvent, error) {
	if err := validateSurfaceQuery(observer, after, after, options); err != nil {
		return SurfaceEvent{}, err
	}
	if kind < SurfaceRise || kind > SurfaceLowerCulmination {
		return SurfaceEvent{}, ErrInvalidObservationInput
	}
	c := surfaceConfigForBody(observer, options)
	c.onlyKind = true
	c.kind = kind
	cycle := math.Max(orbitalPeriodDays[observer.Body], c.solarDays)
	limit := after + math.Min(1.1*cycle, 30*365.25)
	step, gridSteps := c.grid()
	for start := after; start < limit; {
		end := math.Min(start+step, limit)
		if !(end > start) {
			return SurfaceEvent{}, ErrInvalidObservationInput
		}
		windowSample := sample
		if findGap != nil {
			if gap, ok := findGap(start, end); ok {
				windowSample = func(at float64) (SurfacePosition, error) {
					if at >= gap {
						return SurfacePosition{}, ephem.ErrOutOfRange
					}
					return sample(at)
				}
			}
		}
		events, _, err := surfaceSearch(windowSample, start, end, c)
		if errors.Is(err, ephem.ErrOutOfRange) {
			if coveredEnd, ok := surfaceCoveredEnd(windowSample, start, end, step, gridSteps); ok {
				covered, _, coveredErr := surfaceSearchGrid(windowSample, start, coveredEnd, c, true)
				if coveredErr != nil {
					return SurfaceEvent{}, coveredErr
				}
				if len(covered) > 0 {
					return covered[0], nil
				}
			}
		}
		if err != nil {
			return SurfaceEvent{}, err
		}
		if len(events) > 0 {
			return events[0], nil
		}
		start = end
	}
	return SurfaceEvent{}, ErrSurfaceEventNotFound
}

type surfaceEpochEngine struct {
	ephem.Engine
	target ephem.Body
	epochs [4]float64
	count  int
}

func (engine *surfaceEpochEngine) State(target, center ephem.Body, jd float64) ([3]float64, [3]float64, error) {
	if target == engine.target && engine.count < len(engine.epochs) {
		engine.epochs[engine.count] = jd
		engine.count++
	}
	return engine.Engine.State(target, center, jd)
}

func surfaceTargetEpochs(engine ephem.Engine, observer BodyObserver, target ephem.Body, jd float64) (surfaceEpochEngine, error) {
	recorded := surfaceEpochEngine{Engine: engine, target: target}
	frame, err := bodyObserverFrameAt(observer, jd)
	if err != nil {
		return recorded, err
	}
	_, err = surfaceEngineDirection(&recorded, observer.Body, target, frame, jd)
	return recorded, err
}

func surfaceSourceGaps(inst *Instance, observer BodyObserver, target ...SkyBody) surfaceGapFinder {
	engine := inst.Engine()
	if inst.Model() != ephem.Custom || engine == nil {
		return nil
	}
	observerID, ok := surfaceBodyID(observer.Body)
	if !ok {
		return nil
	}
	bodies := []ephem.Body{observerID}
	targetID := ephem.Sun
	for _, body := range target {
		id, ok := surfaceBodyID(body)
		if !ok || id == observerID {
			return nil
		}
		bodies = append(bodies, id)
		targetID = id
	}
	return func(start, end float64) (float64, bool) {
		if targetID == ephem.Sun {
			return ephem.FirstStateGap(engine, bodies, ephem.Sun, start, end)
		}
		left, err := surfaceTargetEpochs(engine, observer, targetID, start)
		if err != nil {
			return start, errors.Is(err, ephem.ErrOutOfRange)
		}
		crossesGap := func(until float64) bool {
			if _, gap := ephem.FirstStateGap(engine, bodies, ephem.Sun, start, until); gap {
				return true
			}
			right, err := surfaceTargetEpochs(engine, observer, targetID, until)
			if errors.Is(err, ephem.ErrOutOfRange) {
				return true
			}
			for stage := 1; stage < right.count && stage < left.count; stage++ {
				first, last := left.epochs[stage], right.epochs[stage]
				if _, gap := ephem.FirstStateGap(engine, bodies[1:], ephem.Sun, math.Min(first, last), math.Max(first, last)); gap {
					return true
				}
			}
			return false
		}
		if !crossesGap(end) {
			return 0, false
		}
		covered, missing := start, end
		for {
			middle := covered + (missing-covered)/2
			if middle == covered || middle == missing {
				return missing, true
			}
			if crossesGap(middle) {
				missing = middle
			} else {
				covered = middle
			}
		}
	}
}

func surfaceCoveredEnd(sample surfaceSampler, start, end, segment float64, steps int) (float64, bool) {
	if _, err := sample(start); err != nil {
		return 0, false
	}
	limit, err := surfaceCoveredLimit(sample, start, end, segment, steps)
	return limit, err == nil && limit < end
}

func surfaceCoveredLimit(sample surfaceSampler, available, limit, segment float64, steps int) (float64, error) {
	count := float64(steps)
	direction := math.Copysign(1, limit-available)
	first := math.Floor((available-2451545)/segment*count) - direction
	for index := first; available != limit; index += direction {
		if index+direction == index {
			return 0, ErrInvalidObservationInput
		}
		turn := math.Floor(index / count)
		probe := (2451545 + turn*segment) + (index-turn*count)*segment/count
		if (probe-available)*direction <= 0 {
			continue
		}
		if (probe-limit)*direction > 0 {
			probe = limit
		}
		_, err := sample(probe)
		if errors.Is(err, ephem.ErrOutOfRange) {
			return surfaceRefineCoveredLimit(sample, available, probe)
		}
		if err != nil {
			return 0, err
		}
		available = probe
	}
	return available, nil
}

func surfaceRefineCoveredLimit(sample surfaceSampler, available, limit float64) (float64, error) {
	for {
		middle := (available + limit) / 2
		if middle == available || middle == limit {
			return available, nil
		}
		_, err := sample(middle)
		if err == nil {
			available = middle
		} else if errors.Is(err, ephem.ErrOutOfRange) {
			limit = middle
		} else {
			return 0, err
		}
	}
}

// SurfaceEventsWithInstance 查询 TT 区间内全部无折射表面事件与区间穿越状态，最多十年 / all airless surface events and interval crossing state in a TT window of at most ten years.
func SurfaceEventsWithInstance(target SkyBody, observer BodyObserver, startJDE, endJDE float64, options SurfaceOptions, inst *Instance) ([]SurfaceEvent, AltitudeCrossingState, error) {
	if inst == nil {
		inst = ephem.Global()
	}
	if target != SkySun {
		options.Limb = AltitudeLimbCenter
	}
	sample := func(jd float64) (SurfacePosition, error) {
		return SurfacePositionAtWithInstance(target, observer, jd, inst)
	}
	return surfaceEventsForSampler(sample, observer, startJDE, endJDE, options, surfaceSourceGaps(inst, observer, target))
}

// StarSurfaceEventsWithInstance 查询 TT 区间内全部恒星表面事件与区间穿越状态，最多十年 / all stellar surface events and interval crossing state in a TT window of at most ten years.
func StarSurfaceEventsWithInstance(star StarCoordinate, observer BodyObserver, startJDE, endJDE float64, options SurfaceOptions, inst *Instance) ([]SurfaceEvent, AltitudeCrossingState, error) {
	if inst == nil {
		inst = ephem.Global()
	}
	options.Limb = AltitudeLimbCenter
	sample := func(jd float64) (SurfacePosition, error) {
		return StarSurfacePositionAtWithInstance(star, observer, jd, inst)
	}
	return surfaceEventsForSampler(sample, observer, startJDE, endJDE, options, surfaceSourceGaps(inst, observer))
}

// NextSurfaceEventWithInstance 查找不早于 TT 时刻的下一次无折射表面事件 / next airless planetary surface event at or after a TT Julian day.
func NextSurfaceEventWithInstance(target SkyBody, observer BodyObserver, kind SurfaceEventKind, afterJDE float64, options SurfaceOptions, inst *Instance) (SurfaceEvent, error) {
	if inst == nil {
		inst = ephem.Global()
	}
	if target != SkySun {
		options.Limb = AltitudeLimbCenter
	}
	sample := func(jd float64) (SurfacePosition, error) {
		return SurfacePositionAtWithInstance(target, observer, jd, inst)
	}
	return nextSurfaceForSampler(sample, observer, kind, afterJDE, options, surfaceSourceGaps(inst, observer, target))
}

// NextStarSurfaceEventWithInstance 查找不早于 TT 时刻的下一次恒星表面事件 / next stellar surface event at or after a TT Julian day.
func NextStarSurfaceEventWithInstance(star StarCoordinate, observer BodyObserver, kind SurfaceEventKind, afterJDE float64, options SurfaceOptions, inst *Instance) (SurfaceEvent, error) {
	if inst == nil {
		inst = ephem.Global()
	}
	options.Limb = AltitudeLimbCenter
	sample := func(jd float64) (SurfacePosition, error) {
		return StarSurfacePositionAtWithInstance(star, observer, jd, inst)
	}
	return nextSurfaceForSampler(sample, observer, kind, afterJDE, options, surfaceSourceGaps(inst, observer))
}

func surfaceForwardAngle(from, to float64) float64 {
	d := math.Mod(to-from, 360)
	if d < 0 {
		d += 360
	}
	return d
}

func surfaceSparseCulminations(sample surfaceSampler, c surfaceSearchConfig, a, z surfaceSearchPoint, emit func(SurfaceEvent), update func(surfaceSearchPoint)) (bool, error) {
	const n = 4
	pts := [n + 1]surfaceSearchPoint{a}
	for k := 1; k < n; k++ {
		p, err := surfacePointAt(sample, a.jd+float64(k)*(z.jd-a.jd)/n, c.options)
		if err != nil {
			return false, err
		}
		pts[k] = p
	}
	pts[n] = z
	drift := math.Abs(math.Remainder(z.position.BodyRA-a.position.BodyRA, 360))
	if drift > 10 {
		return false, nil
	}
	for k := 1; k <= n; k++ {
		d := surfaceForwardAngle(pts[k-1].position.HourAngle, pts[k].position.HourAngle)
		if !(d > 0 && d < 170) {
			return false, nil
		}
	}
	for k := 1; k <= n; k++ {
		update(pts[k])
		lo, hi := pts[k-1], pts[k]
		h0 := lo.position.HourAngle
		span := surfaceForwardAngle(h0, hi.position.HourAngle)
		target := math.Ceil(h0/180) * 180
		if target == h0 {
			target += 180
		}
		if target > h0+span {
			continue
		}
		f := func(p surfaceSearchPoint) float64 { return h0 + surfaceForwardAngle(h0, p.position.HourAngle) - target }
		fa, fb := f(lo), f(hi)
		jd := lo.jd
		for i := 0; i < 64; i++ {
			jd = lo.jd - fa*(hi.jd-lo.jd)/(fb-fa)
			if !(jd > lo.jd && jd < hi.jd) {
				jd = (lo.jd + hi.jd) / 2
			}
			p, err := surfacePointAt(sample, jd, c.options)
			if err != nil {
				return false, err
			}
			fp := f(p)
			rate := (fb - fa) / (hi.jd - lo.jd)
			if math.Abs(fp/rate) < 0.05/86400 || hi.jd-lo.jd <= 0.1/86400 {
				jd = p.jd - fp/rate
				break
			}
			if fp < 0 {
				lo, fa = p, fp
			} else {
				hi, fb = p, fp
			}
		}
		pos, err := sample(jd)
		if err != nil {
			return false, err
		}
		kind := SurfaceUpperCulmination
		if math.Cos(pos.HourAngle*math.Pi/180) < 0 {
			kind = SurfaceLowerCulmination
		}
		emit(SurfaceEvent{Kind: kind, JDE: jd, Position: pos})
	}
	return true, nil
}
