package satellite

import (
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/satellite/internal/sgp4"
	"math"
	"sort"
	"time"
)

// NearbyTransitOptions 站点附近搜索半径与近掠选项 / search radius and near-miss options around a site.
type NearbyTransitOptions struct {
	TransitSearchOptions
	RadiusKM, NearMissArcsec float64
}

// NearbyTransit 站点附近的带、最近距离与正左负右的横向偏移 / nearby band, nearest distance, and cross-track offset positive to the left.
type NearbyTransit struct {
	ElementsIndex                             int
	Event                                     TransitEvent
	Nearest                                   TransitPathPoint
	DistanceKM, NearestDistanceKM, BearingDeg float64
	InBand                                    bool
	Site                                      *Transit
}

type transitNearest struct {
	at       time.Time
	distance float64
}

func transitCandidateMargin(orbit *satelliteOrbit, sky *transitSky, target TransitTarget, obs Observer, radius, minimum float64) transitMargin {
	position, _ := observerPosition(obs)
	return func(t time.Time, exact bool) (float64, error) {
		state, err := orbit.stateAt(t)
		if err != nil {
			return 0, err
		}
		r, _ := earthFixed(state)
		h := vectorNorm(r) - earthRadiusKM
		m := 180.
		b := earthRadiusKM * (1 - earthFlattening)
		if radius < h {
			m = math.Min(180, (radius/(b*b/earthRadiusKM)+math.Asin(radius/h))/degreesToRadians)
		}
		_, sa := horizontalDirection(subtractVector(r, position), obs)
		_, ta := horizontalDirection(subtractVector(sky.position(target, t, exact), position), obs)
		return math.Min(sa, ta) - minimum + m, nil
	}
}
func transitNearestGeometry(orbit *satelliteOrbit, sky *transitSky, event TransitEvent, t time.Time) (transitGeometry, error) {
	g, err := transitGeometryAt(orbit, sky, event.Target, t, true)
	if err != nil {
		return g, err
	}
	var p *TransitPathPoint
	if t.Equal(event.Start) {
		p = &event.StartPoint
	} else if t.Equal(event.End) {
		p = &event.EndPoint
	}
	if p != nil && event.MinTargetAltitude == 0 && math.Abs(p.TargetAltitude) < 1e-8 {
		b := earthRadiusKM * (1 - earthFlattening)
		q := [3]float64{1 / (earthRadiusKM * earthRadiusKM), 1 / (earthRadiusKM * earthRadiusKM), 1 / (b * b)}
		num, den := 0., 0.
		for i := range q {
			num += g.satellite[i] * g.direction[i] * q[i]
			den += g.direction[i] * g.direction[i] * q[i]
		}
		for i := range g.point {
			g.point[i] = g.satellite[i] - num/den*g.direction[i]
		}
		g.discriminant = 0
		g.valid = true
	}
	if !g.valid {
		return g, ErrInvalidInput
	}
	return g, nil
}
func transitNearestRate(orbit *satelliteOrbit, sky *transitSky, event TransitEvent, obs Observer, t time.Time) (float64, error) {
	g, err := transitNearestGeometry(orbit, sky, event, t)
	if err != nil {
		return 0, err
	}
	a, err := orbit.stateAt(t.Add(-500 * time.Millisecond))
	if err != nil {
		return 0, err
	}
	b, err := orbit.stateAt(t.Add(500 * time.Millisecond))
	if err != nil {
		return 0, err
	}
	theta := sgp4.GSTime(basic.UTC2UT1(basic.Date2JD(t.UTC())))
	vs := rotateZ(theta, subtractVector(b.PositionKM, a.PositionKM))
	vs[0] += earthRotationRate * g.satellite[1]
	vs[1] -= earthRotationRate * g.satellite[0]
	va := sky.equatorial(event.Target, t.Add(-500*time.Millisecond))
	vb := sky.equatorial(event.Target, t.Add(500*time.Millisecond))
	gast := sky.sidereal(t)
	vt := rotateZ(gast, subtractVector(vb, va))
	vt[0] += earthRotationRate * g.target[1]
	vt[1] -= earthRotationRate * g.target[0]
	delta := subtractVector(vs, vt)
	along := dotVector(g.direction, delta)
	length := vectorNorm(subtractVector(g.satellite, g.target))
	var dk [3]float64
	for i := range dk {
		dk[i] = (delta[i] - along*g.direction[i]) / length
	}
	ell := dotVector(subtractVector(g.point, g.satellite), g.direction)
	shape := [3]float64{earthRadiusKM, earthRadiusKM, earthRadiusKM * (1 - earthFlattening)}
	var normal, w [3]float64
	for i := range w {
		normal[i] = g.point[i] / (shape[i] * shape[i])
		w[i] = vs[i] + ell*dk[i]
	}
	h := math.Sqrt(math.Max(0, g.discriminant))
	nw := dotVector(normal, w)
	var velocity [3]float64
	for i := range velocity {
		velocity[i] = h*w[i] + g.direction[i]*nw
	}
	lon, lat, _ := geographicPosition(g.point)
	d, az := transitNearbyInverse(lon, lat, obs.Longitude, obs.Latitude)
	if !finite(d) {
		return 0, ErrInvalidInput
	}
	sl, cl := math.Sincos(lon * degreesToRadians)
	sp, cp := math.Sincos(lat * degreesToRadians)
	sa, ca := math.Sincos(az * degreesToRadians)
	toward := [3]float64{-sa*sl - ca*sp*cl, sa*cl - ca*sp*sl, ca * cp}
	return -d * dotVector(toward, velocity), nil
}
func transitNearestInSegments(orbit *satelliteOrbit, sky *transitSky, event TransitEvent, obs Observer, segments []transitSpan, step time.Duration) (transitNearest, error) {
	best := transitNearest{distance: math.Inf(1)}
	distance := func(t time.Time, exact bool) (float64, error) {
		var g transitGeometry
		var err error
		if exact {
			g, err = transitNearestGeometry(orbit, sky, event, t)
		} else {
			g, err = transitGeometryAt(orbit, sky, event.Target, t, false)
		}
		if err != nil {
			return 0, err
		}
		if !g.valid {
			return math.Inf(1), nil
		}
		lon, lat, _ := geographicPosition(g.point)
		d, _ := transitNearbyInverse(obs.Longitude, obs.Latitude, lon, lat)
		if !finite(d) {
			return math.Inf(1), nil
		}
		return d, nil
	}
	take := func(t time.Time) error {
		d, err := distance(t, true)
		if err != nil {
			return err
		}
		if d < best.distance || d == best.distance && (best.at.IsZero() || t.Before(best.at)) {
			best = transitNearest{t, d}
		}
		return nil
	}
	for _, span := range segments {
		lo, hi := span.start, span.end
		if lo.Before(event.Start) {
			lo = event.Start
		}
		if hi.After(event.End) {
			hi = event.End
		}
		if hi.Before(lo) {
			continue
		}
		if err := take(lo); err != nil {
			return best, err
		}
		if err := take(hi); err != nil {
			return best, err
		}
		ts := []time.Time{lo}
		for t := alignedTransitTime(lo, step).Add(step); t.Before(hi); t = t.Add(step) {
			ts = append(ts, t)
		}
		if hi.After(lo) {
			ts = append(ts, hi)
		}
		if len(ts) < 2 {
			continue
		}
		values := make([]float64, len(ts))
		for i, t := range ts {
			v, err := distance(t, false)
			if err != nil {
				return best, err
			}
			values[i] = v
		}
		pairs := [][2]int{{0, 1}, {len(ts) - 2, len(ts) - 1}}
		for i := 1; i+1 < len(ts); i++ {
			if values[i] <= values[i-1] && values[i] <= values[i+1] {
				pairs = append(pairs, [2]int{i - 1, i + 1})
			}
		}
		for _, pair := range pairs {
			a, b := ts[pair[0]], ts[pair[1]]
			fa, err := transitNearestRate(orbit, sky, event, obs, a)
			if err != nil {
				return best, err
			}
			fb, err := transitNearestRate(orbit, sky, event, obs, b)
			if err != nil {
				return best, err
			}
			if fa > 0 || fb < 0 {
				continue
			}
			root, err := transitBisect(a, b, func(t time.Time) (float64, error) { return transitNearestRate(orbit, sky, event, obs, t) }, 10*time.Microsecond)
			if err != nil {
				return best, err
			}
			if err = take(root); err != nil {
				return best, err
			}
		}
	}
	return best, nil
}
func (s *Satellite) nearbyTransits(elements []Elements, obs Observer, start, end time.Time, opt NearbyTransitOptions, candidateStep, distanceStep time.Duration) ([]NearbyTransit, error) {
	if len(elements) < 1 || len(elements) > 1000 || !transitInterval(start, end) || !validTransitSearch(opt.TransitSearchOptions) || !finite(opt.RadiusKM) || opt.RadiusKM <= 0 || opt.RadiusKM > 500 || !finite(opt.NearMissArcsec) || opt.NearMissArcsec < 0 {
		return nil, ErrInvalidInput
	}
	if _, err := observerPosition(obs); err != nil {
		return nil, err
	}
	targets, err := transitTargets(opt.Targets)
	if err != nil {
		return nil, err
	}
	sky := newTransitSky(s)
	var out []NearbyTransit
	for index, e := range elements {
		orbit, err := newSatelliteOrbit(e)
		if err != nil {
			return nil, err
		}
		for _, target := range targets {
			candidate, err := findTransitSpans(start, end, candidateStep, transitCandidateMargin(orbit, sky, target, obs, opt.RadiusKM, opt.MinTargetAltitude))
			if err != nil {
				return nil, err
			}
			if len(candidate) == 0 {
				continue
			}
			spans, err := findTransitSpans(start, end, time.Minute, globalTransitMargin(orbit, sky, target, opt.MinTargetAltitude))
			if err != nil {
				return nil, err
			}
			for _, span := range spans {
				event := TransitEvent{Target: target, Start: span.start, End: span.end, MinTargetAltitude: opt.MinTargetAltitude}
				for i, at := range []time.Time{span.start, span.end} {
					p := TransitPathPoint{Time: at}
					if opt.MinTargetAltitude != 0 || !(i == 0 && span.startNatural || i == 1 && span.endNatural) {
						g, err := transitGeometryAt(orbit, sky, target, at, true)
						if err != nil {
							return nil, err
						}
						p.TargetAltitude = g.altitude
					}
					if i == 0 {
						event.StartPoint = p
					} else {
						event.EndPoint = p
					}
				}
				best, err := transitNearestInSegments(orbit, sky, event, obs, candidate, distanceStep)
				if err != nil {
					return nil, err
				}
				if best.at.IsZero() || best.distance > opt.RadiusKM {
					continue
				}
				lo, hi := span.start.Add(-time.Minute), span.end.Add(time.Minute)
				if lo.Before(start) {
					lo = start
				}
				if hi.After(end) {
					hi = end
				}
				events, err := s.transitEvents(e, lo, hi, TransitSearchOptions{Targets: []TransitTarget{target}, MinTargetAltitude: opt.MinTargetAltitude}, time.Minute, sky)
				if err != nil {
					return nil, err
				}
				found := false
				for _, full := range events {
					if full.Start.Equal(span.start) && full.End.Equal(span.end) {
						event = full
						found = true
						break
					}
				}
				if !found {
					return nil, ErrInvalidInput
				}
				var point TransitPathPoint
				if best.at.Equal(event.Start) {
					point = event.StartPoint
				} else if best.at.Equal(event.End) {
					point = event.EndPoint
				} else {
					point, err = transitEventPoint(orbit, sky, target, best.at, event.Start, event.End, false)
					if err != nil {
						return nil, err
					}
				}
				heading, err := transitEventHeading(orbit, sky, target, best.at, event.Start, event.End)
				if err != nil {
					return nil, err
				}
				d, az := transitNearbyInverse(point.Longitude, point.Latitude, obs.Longitude, obs.Latitude)
				_, bearing := transitNearbyInverse(obs.Longitude, obs.Latitude, point.Longitude, point.Latitude)
				row := NearbyTransit{ElementsIndex: index, Event: event, Nearest: point, NearestDistanceKM: best.distance, DistanceKM: d * math.Sin((heading-az)*degreesToRadians), BearingDeg: bearing}
				siteStart, siteEnd := event.Start, event.End
				pad := 30 * time.Second
				if span := event.End.Sub(event.Start) / 10; span > pad {
					pad = span
				}
				if pad > 2*time.Minute {
					pad = 2 * time.Minute
				}
				if event.Complete || event.Start.After(start) {
					candidate := event.Start.Add(-pad)
					if !event.Complete && candidate.Before(start) {
						candidate = start
					}
					if _, err := orbit.stateAt(candidate); err == nil {
						siteStart = candidate
					}
				}
				if event.Complete || event.End.Before(end) {
					candidate := event.End.Add(pad)
					if !event.Complete && candidate.After(end) {
						candidate = end
					}
					if _, err := orbit.stateAt(candidate); err == nil {
						siteEnd = candidate
					}
				}
				site, err := s.transits(e, obs, siteStart, siteEnd, TransitOptions{Targets: []TransitTarget{target}, NearMissArcsec: opt.NearMissArcsec, MinTargetAltitude: opt.MinTargetAltitude}, 2*time.Second, sky)
				if err != nil {
					return nil, err
				}
				if len(site) > 1 {
					return nil, ErrInvalidInput
				}
				if len(site) == 1 {
					copy := site[0]
					row.Site = &copy
					row.InBand = !copy.NearMiss
				}
				out = append(out, row)
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Nearest.Time.Equal(out[j].Nearest.Time) {
			return out[i].ElementsIndex < out[j].ElementsIndex
		}
		return out[i].Nearest.Time.Before(out[j].Nearest.Time)
	})
	return out, nil
}

// NearbyTransits 搜索站点半径内的凌现象带，距离筛选用真实最近距离 / search transit bands by their true nearest distance to a site.
func (s *Satellite) NearbyTransits(elements []Elements, obs Observer, start, end time.Time, opt NearbyTransitOptions) ([]NearbyTransit, error) {
	return s.nearbyTransits(elements, obs, start, end, opt, time.Minute, 2*time.Second)
}

func transitNearbyInverse(lon1, lat1, lon2, lat2 float64) (float64, float64) {
	d, az := transitInverse(lon1, lat1, lon2, lat2)
	if finite(d) && (d > 0 || math.Abs(lat1-lat2) < 1e-12 && (math.Abs(lat1) == 90 || math.Abs(math.Remainder(lon2-lon1, 360)) < 1e-12)) {
		return d, az
	}
	best, bearing := math.Inf(1), 0.
	for _, seed := range []float64{0, 45, 90, 135, 180, 225, 270, 315} {
		angle, length := seed, 20000.
		for i := 0; i < 32; i++ {
			p := transitDirect(lon1, lat1, angle, length)
			x, y := math.Remainder(lon2-p.Longitude, 360), lat2-p.Latitude
			if math.Hypot(x, y) < 1e-9 {
				if length >= 0 && length < best {
					best = length
					bearing = math.Mod(angle+720, 360)
				}
				break
			}
			a := transitDirect(lon1, lat1, angle+.001, length)
			b := transitDirect(lon1, lat1, angle, length+.01)
			ax, ay := math.Remainder(a.Longitude-p.Longitude, 360)/.001, (a.Latitude-p.Latitude)/.001
			bx, by := math.Remainder(b.Longitude-p.Longitude, 360)/.01, (b.Latitude-p.Latitude)/.01
			det := ax*by - ay*bx
			if math.Abs(det) < 1e-15 {
				break
			}
			da, dd := (x*by-y*bx)/det, (ax*y-ay*x)/det
			scale := math.Max(1, math.Max(math.Abs(da)/30, math.Abs(dd)/1000))
			angle += da / scale
			length += dd / scale
			if length < 0 || length > 22000 {
				break
			}
		}
	}
	if math.IsInf(best, 1) {
		return math.NaN(), math.NaN()
	}
	return best, bearing
}
