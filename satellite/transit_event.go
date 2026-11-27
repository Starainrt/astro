package satellite

import (
	"errors"
	"math"
	"sort"
	"time"
)

// ErrTransitNotFound 搜索范围内没有凌现象带 / no transit band within the search range.
var ErrTransitNotFound = errors.New("satellite transit not found")

// GeoPoint WGS84大地经纬度，东经为正 / WGS84 geodetic coordinates with east-positive longitude.
type GeoPoint struct{ Longitude, Latitude float64 }

// TransitSearchOptions 全球凌现象的目标与几何高度阈值 / targets and geometric altitude threshold for global transit searches.
type TransitSearchOptions struct {
	Targets           []TransitTarget
	MinTargetAltitude float64
}

// TransitPathPoint 中心线或带限上的时刻、位置与半宽 / timed centre-line or band-limit position with half-width.
type TransitPathPoint struct {
	Time                                             time.Time
	Longitude, Latitude, HalfWidthKM, TargetAltitude float64
}

// TransitEvent 连续凌现象带及其原始高度阈值 / continuous transit band and its original altitude threshold.
type TransitEvent struct {
	Target                              TransitTarget
	Start, Greatest, End                time.Time
	StartPoint, GreatestPoint, EndPoint TransitPathPoint
	Complete                            bool
	MaxHalfWidthKM                      float64
	MinTargetAltitude                   float64
}

type transitSpan struct {
	start, end               time.Time
	complete                 bool
	startNatural, endNatural bool
}
type transitMargin func(time.Time, bool) (float64, error)
type transitGeometry struct {
	satellite, target, point, direction [3]float64
	discriminant, altitude              float64
	valid                               bool
}

func validTransitSearch(opt TransitSearchOptions) bool {
	return finite(opt.MinTargetAltitude) && opt.MinTargetAltitude >= 0 && opt.MinTargetAltitude <= 90
}
func transitGeometryAt(orbit *satelliteOrbit, sky *transitSky, target TransitTarget, t time.Time, exact bool) (transitGeometry, error) {
	state, err := orbit.stateAt(t)
	if err != nil {
		return transitGeometry{}, err
	}
	s, _ := earthFixed(state)
	v := sky.position(target, t, exact)
	k := transitUnit(subtractVector(s, v))
	g, disc, valid := transitRay(s, k)
	result := transitGeometry{satellite: s, target: v, point: g, direction: k, discriminant: disc, valid: valid}
	if valid {
		lon, lat, _ := geographicPosition(g)
		_, result.altitude = horizontalDirection(subtractVector(v, g), Observer{Longitude: lon, Latitude: lat})
	}
	return result, nil
}
func globalTransitMargin(orbit *satelliteOrbit, sky *transitSky, target TransitTarget, minimum float64) transitMargin {
	return func(t time.Time, exact bool) (float64, error) {
		g, err := transitGeometryAt(orbit, sky, target, t, exact)
		if err != nil {
			return 0, err
		}
		if g.valid {
			return g.altitude - minimum, nil
		}
		return -math.Sqrt(math.Max(0, -g.discriminant))*earthRadiusKM - 1 - minimum, nil
	}
}
func findTransitSpans(start, end time.Time, step time.Duration, sample transitMargin) ([]transitSpan, error) {
	first := alignedTransitTime(start, step).Add(-step)
	last := alignedTransitTime(end, step).Add(2 * step)
	var roots []time.Time
	exact := func(t time.Time) (float64, error) { return sample(t, true) }
	add := func(a, b time.Time) error {
		fa, err := exact(a)
		if err != nil {
			return err
		}
		fb, err := exact(b)
		if err != nil {
			return err
		}
		if fa*fb > 0 {
			return nil
		}
		root, err := transitBisect(a, b, exact, 500*time.Nanosecond)
		if err != nil {
			return err
		}
		roots = append(roots, root)
		return nil
	}
	a, err := sample(first, false)
	if err != nil {
		return nil, err
	}
	b, err := sample(first.Add(step), false)
	if err != nil {
		return nil, err
	}
	if a*b <= 0 {
		if err = add(first, first.Add(step)); err != nil {
			return nil, err
		}
	}
	for t := first.Add(2 * step); !t.After(last); t = t.Add(step) {
		c, err := sample(t, false)
		if err != nil {
			return nil, err
		}
		if b*c <= 0 {
			if err = add(t.Add(-step), t); err != nil {
				return nil, err
			}
		}
		if b <= 0 && b >= a && b >= c && (b > a || b > c) {
			peak, err := transitGolden(t.Add(-2*step), t, exact, true, 500*time.Microsecond)
			if err != nil {
				return nil, err
			}
			v, err := exact(peak)
			if err != nil {
				return nil, err
			}
			if v > 0 {
				if err = add(t.Add(-2*step), peak); err != nil {
					return nil, err
				}
				if err = add(peak, t); err != nil {
					return nil, err
				}
			}
		}
		a, b = b, c
	}
	sort.Slice(roots, func(i, j int) bool { return roots[i].Before(roots[j]) })
	unique := roots[:0]
	for _, root := range roots {
		if len(unique) == 0 || root.Sub(unique[len(unique)-1]) > time.Microsecond {
			unique = append(unique, root)
		}
	}
	boundaries := append([]time.Time{first}, unique...)
	boundaries = append(boundaries, last)
	var out []transitSpan
	for i := 0; i+1 < len(boundaries); i++ {
		lo, hi := boundaries[i], boundaries[i+1]
		if hi.Before(start) || lo.After(end) {
			continue
		}
		v, err := exact(lo.Add(hi.Sub(lo) / 2))
		if err != nil {
			return nil, err
		}
		if v < 0 {
			continue
		}
		startNatural, endNatural := !lo.Before(start), !hi.After(end)
		complete := true
		if lo.Before(start) {
			lo = start
			complete = false
		}
		if hi.After(end) {
			hi = end
			complete = false
		}
		out = append(out, transitSpan{lo, hi, complete, startNatural, endNatural})
	}
	return out, nil
}
func transitGreatest(span transitSpan, step time.Duration, fn func(time.Time) (float64, error)) (time.Time, error) {
	times := []time.Time{span.start}
	for t := alignedTransitTime(span.start, step).Add(step); t.Before(span.end); t = t.Add(step) {
		times = append(times, t)
	}
	if !span.end.Equal(span.start) {
		times = append(times, span.end)
	}
	best, _, err := transitMaximum(times, fn)
	return best, err
}

func transitMaximum(times []time.Time, fn func(time.Time) (float64, error)) (time.Time, float64, error) {
	values := make([]float64, len(times))
	best := times[0]
	height := math.Inf(-1)
	for i, t := range times {
		v, err := fn(t)
		if err != nil {
			return time.Time{}, 0, err
		}
		values[i] = v
		if v > height {
			best, height = t, v
		}
	}
	for i := 1; i+1 < len(times); i++ {
		if values[i] >= values[i-1] && values[i] >= values[i+1] {
			at, err := transitGolden(times[i-1], times[i+1], fn, true, 500*time.Microsecond)
			if err != nil {
				return time.Time{}, 0, err
			}
			v, err := fn(at)
			if err != nil {
				return time.Time{}, 0, err
			}
			if v > height || v == height && at.Before(best) {
				best, height = at, v
			}
		}
	}
	return best, height, nil
}
func transitEventHeading(orbit *satelliteOrbit, sky *transitSky, target TransitTarget, t, start, end time.Time) (float64, error) {
	lo, hi := t.Add(-100*time.Millisecond), t.Add(100*time.Millisecond)
	if lo.Before(start.Add(100 * time.Microsecond)) {
		lo = start.Add(100 * time.Microsecond)
	}
	if hi.After(end.Add(-100 * time.Microsecond)) {
		hi = end.Add(-100 * time.Microsecond)
	}
	if !lo.Before(hi) {
		lo = start
		hi = end
	}
	if lo.Equal(hi) {
		return 0, nil
	}
	a, err := transitGeometryAt(orbit, sky, target, lo, true)
	if err != nil {
		return 0, err
	}
	b, err := transitGeometryAt(orbit, sky, target, hi, true)
	if err != nil {
		return 0, err
	}
	if !a.valid || !b.valid {
		mid, err := transitGeometryAt(orbit, sky, target, lo.Add(hi.Sub(lo)/2), true)
		if err != nil {
			return 0, err
		}
		if !a.valid && mid.valid {
			a = mid
		}
		if !b.valid && mid.valid {
			b = mid
		}
		if !a.valid || !b.valid || a.point == b.point {
			return 0, nil
		}
	}
	al, ap, _ := geographicPosition(a.point)
	bl, bp, _ := geographicPosition(b.point)
	_, heading := transitInverse(al, ap, bl, bp)
	return heading, nil
}
func transitCross(a, b [3]float64) [3]float64 {
	return [3]float64{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
}

type transitLimbProjection struct {
	satellite, axis, east, north [3]float64
	sine, cosine                 float64
}

func newTransitLimbProjection(g transitGeometry, target TransitTarget) transitLimbProjection {
	u := transitUnit(subtractVector(g.target, g.satellite))
	east := transitUnit(transitCross([3]float64{0, 0, 1}, u))
	north := transitCross(u, east)
	alpha := math.Asin(transitRadius(target) / vectorNorm(subtractVector(g.target, g.satellite)))
	sa, ca := math.Sincos(alpha)
	return transitLimbProjection{g.satellite, u, east, north, sa, ca}
}

func (limb transitLimbProjection) point(pa float64) ([3]float64, bool) {
	sp, cp := math.Sincos(pa)
	var k [3]float64
	for i := range k {
		k[i] = -limb.cosine*limb.axis[i] - limb.sine*(cp*limb.east[i]+sp*limb.north[i])
	}
	p, _, ok := transitRay(limb.satellite, k)
	return p, ok
}
func transitAngleMaximum(lo, hi float64, fn func(float64) float64) float64 {
	q := (math.Sqrt(5) - 1) / 2
	a, b := hi-q*(hi-lo), lo+q*(hi-lo)
	fa, fb := fn(a), fn(b)
	for hi-lo > 1e-8 {
		if fa >= fb {
			hi = b
			b = a
			fb = fa
			a = hi - q*(hi-lo)
			fa = fn(a)
		} else {
			lo = a
			a = b
			fa = fb
			b = lo + q*(hi-lo)
			fb = fn(b)
		}
	}
	return (hi + lo) / 2
}
func transitWidth(g transitGeometry, target TransitTarget, heading float64) (float64, bool) {
	lon, lat, _ := geographicPosition(g.point)
	limb := newTransitLimbProjection(g, target)
	projector := newTransitWidthProjector(lon, lat, heading)
	var points [72][3]float64
	delta := 2 * math.Pi / float64(len(points))
	for index := range points {
		point, ok := limb.point(float64(index) * delta)
		if !ok {
			return 0, false
		}
		points[index] = point
	}
	var values [72]float64
	maximum := 0.0
	for index, point := range points {
		value := math.Abs(projector.crossTrack(point))
		if !finite(value) {
			return 0, false
		}
		values[index] = value
		maximum = math.Max(maximum, value)
	}
	for index, value := range values {
		if value < values[(index+len(values)-1)%len(values)] || value < values[(index+1)%len(values)] {
			continue
		}
		score := func(angle float64) float64 {
			point, ok := limb.point(angle)
			if !ok {
				return math.Inf(-1)
			}
			value := projector.crossTrack(point)
			if !finite(value) {
				return math.Inf(-1)
			}
			return math.Abs(value)
		}
		angle := transitAngleMaximum(float64(index-1)*delta, float64(index+1)*delta, score)
		refined := score(angle)
		if !finite(refined) {
			return 0, false
		}
		maximum = math.Max(maximum, refined)
	}
	return maximum, true
}
func transitEventPoint(orbit *satelliteOrbit, sky *transitSky, target TransitTarget, t, start, end time.Time, tangent bool) (TransitPathPoint, error) {
	g, err := transitGeometryAt(orbit, sky, target, t, true)
	if err != nil {
		return TransitPathPoint{}, err
	}
	if tangent {
		b := earthRadiusKM * (1 - earthFlattening)
		q := [3]float64{1 / (earthRadiusKM * earthRadiusKM), 1 / (earthRadiusKM * earthRadiusKM), 1 / (b * b)}
		num, den := 0., 0.
		for i := range q {
			num += g.satellite[i] * g.direction[i] * q[i]
			den += g.direction[i] * g.direction[i] * q[i]
		}
		for i := range q {
			g.point[i] = g.satellite[i] - num/den*g.direction[i]
		}
		lon, lat, _ := geographicPosition(g.point)
		_, g.altitude = horizontalDirection(subtractVector(g.target, g.point), Observer{Longitude: lon, Latitude: lat})
	} else if !g.valid {
		return TransitPathPoint{}, ErrInvalidInput
	}
	heading, err := transitEventHeading(orbit, sky, target, t, start, end)
	if err != nil {
		return TransitPathPoint{}, err
	}
	lon, lat, _ := geographicPosition(g.point)
	point := TransitPathPoint{Time: t, Longitude: lon, Latitude: lat, TargetAltitude: g.altitude}
	width, fullLimb := transitWidth(g, target, heading)
	if !fullLimb {
		projector := newTransitWidthProjector(lon, lat, heading)
		_, best := transitPathExtrema(newTransitPathCone(g, target), projector.crossTrack)
		width = math.Max(0, math.Max(best[0], best[1]))
	}
	point.HalfWidthKM = width
	return point, nil
}

func transitEventMaxWidth(orbit *satelliteOrbit, sky *transitSky, event TransitEvent) (float64, error) {
	times := []time.Time{event.Start, event.Greatest, event.End}
	for at := alignedTransitTime(event.Start, time.Minute).Add(time.Minute); at.Before(event.End); at = at.Add(time.Minute) {
		times = append(times, at)
	}
	span := event.End.Sub(event.Start)
	first := span / 64
	if first > time.Second {
		first = time.Second
	}
	for offset := first; offset > 0 && offset < span/2 && offset < time.Minute; offset *= 2 {
		times = append(times, event.Start.Add(offset), event.End.Add(-offset))
	}
	sort.Slice(times, func(left, right int) bool { return times[left].Before(times[right]) })
	unique := times[:1]
	for _, at := range times[1:] {
		if !at.Equal(unique[len(unique)-1]) {
			unique = append(unique, at)
		}
	}
	_, width, err := transitMaximum(unique, func(at time.Time) (float64, error) {
		for _, point := range [...]TransitPathPoint{event.StartPoint, event.GreatestPoint, event.EndPoint} {
			if at.Equal(point.Time) {
				return point.HalfWidthKM, nil
			}
		}
		point, err := transitEventPoint(orbit, sky, event.Target, at, event.Start, event.End, false)
		return point.HalfWidthKM, err
	})
	return width, err
}

// TransitEvents 搜索不超过30天的全球凌日凌月带 / search global solar and lunar transit bands over at most 30 days.
func (s *Satellite) TransitEvents(e Elements, start, end time.Time, opt TransitSearchOptions) ([]TransitEvent, error) {
	return s.transitEvents(e, start, end, opt, time.Minute, newTransitSky(s))
}
func (s *Satellite) transitEvents(e Elements, start, end time.Time, opt TransitSearchOptions, step time.Duration, sky *transitSky) ([]TransitEvent, error) {
	targets, err := transitTargets(opt.Targets)
	if err != nil || !validTransitSearch(opt) || !transitInterval(start, end) {
		return nil, ErrInvalidInput
	}
	orbit, err := newSatelliteOrbit(e)
	if err != nil {
		return nil, err
	}
	if _, err = orbit.stateAt(start); err != nil {
		return nil, err
	}
	if _, err = orbit.stateAt(end); err != nil {
		return nil, err
	}
	var out []TransitEvent
	for _, target := range targets {
		spans, err := findTransitSpans(start, end, step, globalTransitMargin(orbit, sky, target, opt.MinTargetAltitude))
		if err != nil {
			return nil, err
		}
		for _, span := range spans {
			height := func(t time.Time) (float64, error) {
				g, err := transitGeometryAt(orbit, sky, target, t, true)
				return g.altitude, err
			}
			greatest, err := transitGreatest(span, time.Minute, height)
			if err != nil {
				return nil, err
			}
			event := TransitEvent{Target: target, Start: span.start, Greatest: greatest, End: span.end, Complete: span.complete, MinTargetAltitude: opt.MinTargetAltitude}
			points := []*TransitPathPoint{&event.StartPoint, &event.GreatestPoint, &event.EndPoint}
			for i, t := range []time.Time{event.Start, event.Greatest, event.End} {
				tangent := opt.MinTargetAltitude == 0 && (i == 0 && span.startNatural || i == 2 && span.endNatural)
				*points[i], err = transitEventPoint(orbit, sky, target, t, event.Start, event.End, tangent)
				if err != nil {
					return nil, err
				}
			}
			event.MaxHalfWidthKM, err = transitEventMaxWidth(orbit, sky, event)
			if err != nil {
				return nil, err
			}
			out = append(out, event)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Greatest.Equal(out[j].Greatest) {
			return out[i].Target < out[j].Target
		}
		return out[i].Greatest.Before(out[j].Greatest)
	})
	return out, nil
}

// ClosestTransitEvent 搜索包含此刻或前后各30天内最近的凌现象带 / find the containing band or the closest within 30 days on either side.
func (s *Satellite) ClosestTransitEvent(e Elements, target TransitTarget, t time.Time, opt TransitSearchOptions) (TransitEvent, error) {
	if !validTransitTarget(target) || t.IsZero() {
		return TransitEvent{}, ErrInvalidInput
	}
	opt.Targets = []TransitTarget{target}
	sky := newTransitSky(s)
	var events []TransitEvent
	distance := func(e TransitEvent) time.Duration {
		if t.Before(e.Start) {
			return e.Start.Sub(t)
		}
		if t.After(e.End) {
			return t.Sub(e.End)
		}
		return 0
	}
	for reach := 24 * time.Hour; reach <= 30*24*time.Hour; reach += 24 * time.Hour {
		lo, hi := t.Add(-reach), t.Add(reach)
		var windows [][2]time.Time
		if reach <= 15*24*time.Hour {
			windows = append(windows, [2]time.Time{lo, hi})
		} else {
			windows = append(windows, [2]time.Time{lo, t}, [2]time.Time{t, hi})
		}
		events = events[:0]
		for _, w := range windows {
			found, err := s.transitEvents(e, w[0], w[1], opt, time.Minute, sky)
			if err != nil {
				return TransitEvent{}, err
			}
			events = append(events, found...)
		}
		if len(events) == 0 {
			continue
		}
		best := events[0]
		for _, event := range events[1:] {
			if distance(event) < distance(best) || distance(event) == distance(best) && event.Greatest.Before(best.Greatest) {
				best = event
			}
		}
		if best.Complete {
			return best, nil
		}
		if reach == 30*24*time.Hour {
			return best, nil
		}
	}
	return TransitEvent{}, ErrTransitNotFound
}
