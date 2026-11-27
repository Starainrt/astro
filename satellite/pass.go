package satellite

import (
	"errors"
	"math"
	"sort"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/satellite/internal/sgp4"
)

// ErrPassNotFound 搜索区间内没有过境 / no pass within the search interval.
var ErrPassNotFound = errors.New("satellite pass not found")

// PassKind 过境的可见性类别 / pass visibility category.
type PassKind int

const (
	// PassVisible 至少存在一个可见时段 / at least one visible interval.
	PassVisible PassKind = iota
	// PassDaylight 整场过境均在白昼阈值之上 / the entire pass is above the daylight threshold.
	PassDaylight
	// PassEclipsed 没有可见时段且不满足全程白昼 / no visible interval and not entirely in daylight.
	PassEclipsed
)

// PassOptions 过境选项，阈值与输出仰角遵循折射选项 / pass options; threshold and output elevations follow the refraction option.
type PassOptions struct {
	ObserveOptions
	MinElevation        float64
	TwilightSunAltitude *float64
}

// Pass 过境的接触、极大与可见性 / pass contacts, culmination and visibility.
type Pass struct {
	Rise, Culmination, Set   Look
	Duration                 time.Duration
	StartsAbove, EndsAbove   bool
	Kind                     PassKind
	VisibleStart, VisibleEnd time.Time // VisibleStart,VisibleEnd 首末可见时刻，无可见时段为零值 / first and last visible instants, zero without visibility.
	BrightestTime            time.Time // BrightestTime 在实际可见子时段内最亮的时刻 / brightest instant within the actual visible intervals.
	BrightestMagnitude       float64
}

type passSample struct {
	time           time.Time
	altitude, rate float64
}

type passSpan struct {
	start, peak, end       time.Time
	peakAltitude           float64
	startsAbove, endsAbove bool
}

type passSampler func(time.Time) (passSample, error)

// Passes 搜索闭区间内的过境，极大按几何仰角导数求根 / search passes in a closed interval, locating culmination by the geometric elevation derivative.
func (s *Satellite) Passes(e Elements, obs Observer, start, end time.Time, opt PassOptions) ([]Pass, basic.AltitudeCrossingState, error) {
	return s.passes(e, obs, start, end, opt, 0, false)
}

func (s *Satellite) passes(e Elements, obs Observer, start, end time.Time, opt PassOptions, step time.Duration, first bool) ([]Pass, basic.AltitudeCrossingState, error) {
	orbit, err := newSatelliteOrbit(e)
	if err != nil {
		return nil, basic.AltitudeAlwaysBelow, err
	}
	position, err := observerPosition(obs)
	if err != nil || !validPassOptions(opt) || start.IsZero() || end.IsZero() || end.Before(start) || end.Sub(start) > 30*24*time.Hour || !start.Add(end.Sub(start)).Equal(end) {
		return nil, basic.AltitudeAlwaysBelow, ErrInvalidInput
	}
	if _, err = orbit.stateAt(start); err != nil {
		return nil, basic.AltitudeAlwaysBelow, err
	}
	if !first {
		if _, err = orbit.stateAt(end); err != nil {
			return nil, basic.AltitudeAlwaysBelow, err
		}
	}
	threshold := opt.MinElevation
	if !opt.NoRefraction {
		pressure, temperature := opt.PressureHPa, opt.TemperatureC
		if pressure == 0 {
			pressure, temperature = 1010, 10
		}
		threshold = basic.TrueAltitude(threshold, pressure, temperature)
		if !finite(threshold) {
			return nil, basic.AltitudeAlwaysBelow, ErrInvalidInput
		}
	}
	if step <= 0 {
		step = passStep(e)
	}
	sample := func(ti time.Time) (passSample, error) { return orbit.passSampleAt(obs, position, ti) }
	spans, state, err := searchPasses(start, end, threshold, step, sample, first)
	if err != nil {
		return nil, state, err
	}
	out := make([]Pass, len(spans))
	for i, span := range spans {
		p := &out[i]
		p.Duration = span.end.Sub(span.start)
		p.StartsAbove, p.EndsAbove = span.startsAbove, span.endsAbove
		p.Kind = PassEclipsed
		p.BrightestMagnitude = math.NaN()
		fields := []*Look{&p.Rise, &p.Culmination, &p.Set}
		for i, ti := range []time.Time{span.start, span.peak, span.end} {
			*fields[i], err = s.lookAtOrbit(orbit, obs, position, ti, opt.ObserveOptions)
			if err != nil {
				return nil, state, err
			}
		}
		if err = s.completePass(p, orbit, obs, position, opt, step); err != nil {
			return nil, state, err
		}
	}
	return out, state, nil
}

// NextPass 搜索此刻起至多30天的下一过境，进行中的过境截于输入时刻 / find the next pass within 30 days, clipping an ongoing pass at the input time.
func (s *Satellite) NextPass(e Elements, obs Observer, after time.Time, opt PassOptions) (Pass, error) {
	passes, _, err := s.passes(e, obs, after, after.Add(30*24*time.Hour), opt, 0, true)
	if err != nil {
		return Pass{}, err
	}
	if len(passes) == 0 {
		return Pass{}, ErrPassNotFound
	}
	return passes[0], nil
}

func validPassOptions(opt PassOptions) bool {
	return validObserveOptions(opt.ObserveOptions) && finite(opt.MinElevation) && math.Abs(opt.MinElevation) <= 90 && (opt.TwilightSunAltitude == nil || finite(*opt.TwilightSunAltitude) && math.Abs(*opt.TwilightSunAltitude) <= 90)
}

func passStep(e Elements) time.Duration {
	seconds := math.Min(60, 86400/e.MeanMotion/120)
	return time.Duration(math.Floor(1000*seconds)) * time.Millisecond
}

func (o *satelliteOrbit) passSampleAt(obs Observer, position [3]float64, t time.Time) (passSample, error) {
	at, err := o.stateAt(t)
	if err != nil {
		return passSample{}, err
	}
	before, err := o.stateAt(t.Add(-500 * time.Millisecond))
	if err != nil {
		return passSample{}, err
	}
	after, err := o.stateAt(t.Add(500 * time.Millisecond))
	if err != nil {
		return passSample{}, err
	}
	theta := sgp4.GSTime(basic.UTC2UT1(basic.Date2JD(t.UTC())))
	r := rotateZ(theta, at.PositionKM)
	v := rotateZ(theta, subtractVector(after.PositionKM, before.PositionKM))
	v[0] += earthRotationRate * r[1]
	v[1] -= earthRotationRate * r[0]
	d := subtractVector(r, position)
	distance := vectorNorm(d)
	if distance <= 0 || !finite(distance) {
		return passSample{}, ErrInvalidInput
	}
	_, alt := horizontalDirection(d, obs)
	sinLat, cosLat := math.Sincos(obs.Latitude * degreesToRadians)
	sinLon, cosLon := math.Sincos(obs.Longitude * degreesToRadians)
	up := [3]float64{cosLat * cosLon, cosLat * sinLon, sinLat}
	rate := dotVector(v, up) - dotVector(d, up)*dotVector(d, v)/(distance*distance)
	return passSample{time: t, altitude: alt, rate: rate}, nil
}

func passRoot(a, b passSample, sample passSampler, threshold float64, derivative bool) (passSample, error) {
	value := func(x passSample) float64 {
		if derivative {
			return x.rate
		}
		return x.altitude - threshold
	}
	for b.time.Sub(a.time) > 500*time.Microsecond {
		mid, err := sample(a.time.Add(b.time.Sub(a.time) / 2))
		if err != nil {
			return passSample{}, err
		}
		if value(a)*value(mid) <= 0 {
			b = mid
		} else {
			a = mid
		}
	}
	return sample(a.time.Add(b.time.Sub(a.time) / 2))
}

func searchPasses(start, end time.Time, threshold float64, step time.Duration, sample passSampler, first bool) ([]passSpan, basic.AltitudeCrossingState, error) {
	if step <= 0 {
		return nil, basic.AltitudeAlwaysBelow, ErrInvalidInput
	}
	previous, err := sample(start)
	if err != nil {
		return nil, basic.AltitudeAlwaysBelow, err
	}
	var spans, tangents []passSpan
	active := previous.altitude >= threshold
	current := passSpan{start: start, peak: start, peakAltitude: previous.altitude, startsAbove: active}
	crosses := false
	consider := func(p passSample) {
		if p.altitude > current.peakAltitude {
			current.peak = p.time
			current.peakAltitude = p.altitude
		}
	}
	consume := func(a, b passSample) error {
		if !active && a.altitude <= threshold && b.altitude > threshold || active && a.altitude >= threshold && b.altitude < threshold {
			root, err := passRoot(a, b, sample, threshold, false)
			if err != nil {
				return err
			}
			crosses = true
			if b.altitude > threshold {
				active = true
				current = passSpan{start: root.time, peak: root.time, peakAltitude: root.altitude}
				consider(b)
			} else {
				consider(a)
				consider(root)
				current.end = root.time
				spans = append(spans, current)
				active = false
			}
		} else if active {
			consider(b)
		}
		return nil
	}
	for previous.time.Before(end) {
		nextTime := previous.time.Add(step)
		if nextTime.After(end) {
			nextTime = end
		}
		next, err := sample(nextTime)
		if err != nil {
			return nil, basic.AltitudeAlwaysBelow, err
		}
		if previous.rate > 0 && next.rate <= 0 || previous.rate < 0 && next.rate >= 0 {
			extreme, err := passRoot(previous, next, sample, threshold, true)
			if err != nil {
				return nil, basic.AltitudeAlwaysBelow, err
			}
			if previous.rate > 0 && math.Abs(extreme.altitude-threshold) < 1.0/3600 {
				tangents = append(tangents, passSpan{start: extreme.time, peak: extreme.time, end: extreme.time, peakAltitude: extreme.altitude})
			}
			if err = consume(previous, extreme); err != nil {
				return nil, basic.AltitudeAlwaysBelow, err
			}
			if err = consume(extreme, next); err != nil {
				return nil, basic.AltitudeAlwaysBelow, err
			}
		} else if err = consume(previous, next); err != nil {
			return nil, basic.AltitudeAlwaysBelow, err
		}
		previous = next
		if first && (len(spans) > 0 || len(tangents) > 0) {
			break
		}
	}
	if active {
		current.end = end
		current.endsAbove = true
		spans = append(spans, current)
	}
	for _, tangent := range tangents {
		covered := false
		for _, span := range spans {
			if !tangent.peak.Before(span.start) && !tangent.peak.After(span.end) {
				covered = true
				break
			}
		}
		if !covered {
			spans = append(spans, tangent)
		}
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].start.Before(spans[j].start) })
	state := basic.AltitudeAlwaysBelow
	if crosses {
		state = basic.AltitudeCrosses
	} else if active {
		state = basic.AltitudeAlwaysAbove
	} else if len(tangents) > 0 {
		state = basic.AltitudeTangent
	}
	if first && len(spans) > 1 {
		spans = spans[:1]
	}
	return spans, state, nil
}

type passLightSample struct {
	time                    time.Time
	night, light, magnitude float64
}

type passLightSampler func(time.Time) (passLightSample, error)

type passVisibleSpan struct{ start, end time.Time }

func (s *Satellite) passLightAt(orbit *satelliteOrbit, obs Observer, position [3]float64, t time.Time, opt PassOptions) (passLightSample, error) {
	state, err := orbit.stateAt(t)
	if err != nil {
		return passLightSample{}, err
	}
	r, _ := earthFixed(state)
	jdUTC := basic.Date2JD(t.UTC())
	jdTT := basic.UTC2TT(jdUTC)
	gast := basic.ApparentSiderealTime(basic.UTC2UT1(jdUTC)) * 15 * degreesToRadians
	var inst *ephem.Instance
	if s != nil {
		inst = s.inst
	}
	sun := apparentSunPosition(jdTT, gast, inst)
	_, solarAltitude := horizontalDirection(subtractVector(sun, position), obs)
	twilight := -6.0
	if opt.TwilightSunAltitude != nil {
		twilight = *opt.TwilightSunAltitude
	}
	toSun := subtractVector(sun, r)
	separation := vectorAngle([3]float64{-r[0], -r[1], -r[2]}, toSun)
	earthAngle := earthAngularRadius(r)
	magnitude := math.NaN()
	if opt.StandardMagnitude != nil {
		fraction := illuminatedFraction(earthAngle, math.Asin(695700/vectorNorm(toSun)), separation)
		phase := vectorAngle(toSun, subtractVector(position, r)) / degreesToRadians
		magnitude = satelliteMagnitude(vectorNorm(subtractVector(r, position)), phase*degreesToRadians, fraction, opt.StandardMagnitude)
	}
	return passLightSample{t, twilight - solarAltitude, (separation - earthAngle) / degreesToRadians, magnitude}, nil
}

func passLightExtreme(a, b time.Time, sample passLightSampler, field func(passLightSample) float64, maximize bool) (passLightSample, error) {
	const ratio = .6180339887498948482
	x := a.Add(time.Duration(float64(b.Sub(a)) * (1 - ratio)))
	y := a.Add(time.Duration(float64(b.Sub(a)) * ratio))
	fx, err := sample(x)
	if err != nil {
		return passLightSample{}, err
	}
	fy, err := sample(y)
	if err != nil {
		return passLightSample{}, err
	}
	for b.Sub(a) > 500*time.Microsecond {
		less := field(fx) < field(fy)
		if maximize {
			less = !less
		}
		if less {
			b = y
			y = x
			fy = fx
			x = a.Add(time.Duration(float64(b.Sub(a)) * (1 - ratio)))
			fx, err = sample(x)
		} else {
			a = x
			x = y
			fx = fy
			y = a.Add(time.Duration(float64(b.Sub(a)) * ratio))
			fy, err = sample(y)
		}
		if err != nil {
			return passLightSample{}, err
		}
	}
	if (field(fx) < field(fy)) != maximize {
		return fx, nil
	}
	return fy, nil
}

func passLightRoot(a, b passLightSample, sample passLightSampler, field func(passLightSample) float64) (passLightSample, error) {
	for b.time.Sub(a.time) > 500*time.Microsecond {
		mid, err := sample(a.time.Add(b.time.Sub(a.time) / 2))
		if err != nil {
			return passLightSample{}, err
		}
		if field(a)*field(mid) <= 0 {
			b = mid
		} else {
			a = mid
		}
	}
	if field(a) >= 0 {
		return a, nil
	}
	return b, nil
}

func passVisibility(start, end time.Time, step time.Duration, sample passLightSampler) ([]passVisibleSpan, bool, error) {
	var points []passLightSample
	if d := end.Sub(start); d > time.Nanosecond && step > d/2 {
		step = d / 2
	}
	for ti := start; ; {
		p, err := sample(ti)
		if err != nil {
			return nil, false, err
		}
		points = append(points, p)
		if !ti.Before(end) {
			break
		}
		ti = ti.Add(step)
		if ti.After(end) {
			ti = end
		}
	}
	fields := []func(passLightSample) float64{func(p passLightSample) float64 { return p.night }, func(p passLightSample) float64 { return p.light }}
	extras := []passLightSample{}
	for _, field := range fields {
		for i := 1; i+1 < len(points); i++ {
			a, b, c := field(points[i-1]), field(points[i]), field(points[i+1])
			if b >= a && b >= c && (b > a || b > c) || b <= a && b <= c && (b < a || b < c) {
				p, err := passLightExtreme(points[i-1].time, points[i+1].time, sample, field, b > a || b > c)
				if err != nil {
					return nil, false, err
				}
				extras = append(extras, p)
			}
		}
	}
	points = append(points, extras...)
	sort.Slice(points, func(i, j int) bool { return points[i].time.Before(points[j].time) })
	daylight := true
	for _, p := range points {
		if p.night >= 0 {
			daylight = false
		}
	}
	bounds := []time.Time{start, end}
	for i := 1; i < len(points); i++ {
		for _, field := range fields {
			if field(points[i-1])*field(points[i]) < 0 {
				root, err := passLightRoot(points[i-1], points[i], sample, field)
				if err != nil {
					return nil, false, err
				}
				bounds = append(bounds, root.time)
			}
		}
	}
	sort.Slice(bounds, func(i, j int) bool { return bounds[i].Before(bounds[j]) })
	spans := []passVisibleSpan{}
	if start.Equal(end) {
		p := points[0]
		if p.night >= 0 && p.light > 0 {
			spans = append(spans, passVisibleSpan{start, end})
		}
	}
	for i := 1; i < len(bounds); i++ {
		a, b := bounds[i-1], bounds[i]
		if a.Equal(b) {
			continue
		}
		p, err := sample(a.Add(b.Sub(a) / 2))
		if err != nil {
			return nil, false, err
		}
		if p.night >= 0 && p.light > 0 {
			if len(spans) > 0 && spans[len(spans)-1].end.Equal(a) {
				spans[len(spans)-1].end = b
			} else {
				spans = append(spans, passVisibleSpan{a, b})
			}
		}
	}
	return spans, daylight, nil
}

func passBrightest(spans []passVisibleSpan, step time.Duration, sample passLightSampler) (time.Time, float64, error) {
	var best time.Time
	minimum := math.NaN()
	consider := func(p passLightSample) {
		if p.night >= 0 && p.light > 0 && finite(p.magnitude) && (best.IsZero() || p.magnitude < minimum) {
			best = p.time
			minimum = p.magnitude
		}
	}
	for _, span := range spans {
		localStep := step
		if d := span.end.Sub(span.start); d > time.Nanosecond && localStep > d/2 {
			localStep = d / 2
		}
		var points []passLightSample
		for ti := span.start; ; {
			p, err := sample(ti)
			if err != nil {
				return time.Time{}, math.NaN(), err
			}
			consider(p)
			points = append(points, p)
			if !ti.Before(span.end) {
				break
			}
			ti = ti.Add(localStep)
			if ti.After(span.end) {
				ti = span.end
			}
		}
		field := func(p passLightSample) float64 { return p.magnitude }
		for i := 1; i+1 < len(points); i++ {
			if points[i].magnitude < points[i-1].magnitude && points[i].magnitude < points[i+1].magnitude {
				p, err := passLightExtreme(points[i-1].time, points[i+1].time, sample, field, false)
				if err != nil {
					return time.Time{}, math.NaN(), err
				}
				consider(p)
			}
		}
	}
	return best, minimum, nil
}

func (s *Satellite) completePass(p *Pass, orbit *satelliteOrbit, obs Observer, position [3]float64, opt PassOptions, step time.Duration) error {
	sample := func(ti time.Time) (passLightSample, error) { return s.passLightAt(orbit, obs, position, ti, opt) }
	spans, daylight, err := passVisibility(p.Rise.Time, p.Set.Time, step, sample)
	if err != nil {
		return err
	}
	if len(spans) == 0 {
		if daylight {
			p.Kind = PassDaylight
		}
		return nil
	}
	p.Kind = PassVisible
	p.VisibleStart = spans[0].start
	p.VisibleEnd = spans[len(spans)-1].end
	if opt.StandardMagnitude != nil {
		p.BrightestTime, p.BrightestMagnitude, err = passBrightest(spans, step, sample)
	}
	return err
}
