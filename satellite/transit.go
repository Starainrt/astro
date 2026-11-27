package satellite

import (
	"math"
	"sort"
	"time"

	"github.com/starainrt/astro/basic"
)

// TransitTarget 凌现象的背景天体 / background body for a transit.
type TransitTarget int

const (
	// TransitSun 凌日 / solar transit.
	TransitSun TransitTarget = iota
	// TransitMoon 凌月，包含暗月 / lunar transit, including the unlit Moon.
	TransitMoon
)

// TransitOptions 站点凌现象与近掠选项，仰角为几何值 / site transit and near-miss options with geometric altitudes.
type TransitOptions struct {
	Targets           []TransitTarget
	NearMissArcsec    float64
	MinTargetAltitude float64
}

// Transit 站点凌现象，接触完整且以Center归属查询闭区间 / site transit with full contacts and Center membership in the closed query interval.
type Transit struct {
	Target                                  TransitTarget
	Start, Center, End                      time.Time
	Duration                                time.Duration
	NearMiss                                bool
	MinSeparationArcsec, TargetRadiusArcsec float64
	Satellite                               Look
	TargetAltitude                          float64
	CenterlineOffsetKM                      float64 // CenterlineOffsetKM 正左负右，中心射线不交WGS84椭球时为NaN / signed cross-track distance, positive left, NaN when the centre ray misses WGS84.
}

type transitSample struct{ separation, radius, satelliteAltitude, targetAltitude float64 }
type siteTransitSpan struct {
	start, center, end time.Time
	nearMiss           bool
}
type transitSampler func(time.Time, bool) (transitSample, error)
type transitSky struct {
	module              *Satellite
	values              map[transitSkyKey][3]float64
	siderealValues      map[int64]float64
	direct              bool
	disableSiteScanSkip bool
}
type transitSkyKey struct {
	seconds int64
	target  TransitTarget
}

func validTransitTarget(target TransitTarget) bool {
	return target == TransitSun || target == TransitMoon
}
func transitTargets(targets []TransitTarget) ([]TransitTarget, error) {
	if len(targets) == 0 {
		return []TransitTarget{TransitSun, TransitMoon}, nil
	}
	out := make([]TransitTarget, 0, 2)
	for _, target := range targets {
		if !validTransitTarget(target) {
			return nil, ErrInvalidInput
		}
		found := false
		for _, x := range out {
			found = found || x == target
		}
		if !found {
			out = append(out, target)
		}
	}
	return out, nil
}
func transitInterval(start, end time.Time) bool {
	d := end.Sub(start)
	return !start.IsZero() && !end.IsZero() && !end.Before(start) && d <= 30*24*time.Hour && start.Add(d).Equal(end)
}
func newTransitSky(s *Satellite) *transitSky {
	if s == nil {
		s = &defaultModule
	}
	return &transitSky{module: s, values: make(map[transitSkyKey][3]float64), siderealValues: make(map[int64]float64)}
}
func (sky *transitSky) directEquatorial(target TransitTarget, t time.Time) [3]float64 {
	jd := basic.UTC2TT(basic.Date2JD(t.UTC()))
	var ra, dec, dist float64
	if target == TransitSun {
		ra, dec = basic.HSunApparentRaDecNWithInstance(jd, -1, sky.module.inst)
		dist = basic.EarthAwayNWithInstance(jd, -1, sky.module.inst) * 149597870.7
	} else {
		ra, dec = basic.HMoonGeocentricApparentRaDecNWithInstance(jd, -1, sky.module.inst)
		dist = basic.HMoonAwayNWithInstance(jd, -1, sky.module.inst)
	}
	sr, cr := math.Sincos(ra * degreesToRadians)
	sd, cd := math.Sincos(dec * degreesToRadians)
	return [3]float64{dist * cd * cr, dist * cd * sr, dist * sd}
}
func transitInterpolation(t time.Time) (int64, [5]float64) {
	anchor := t.Unix() - modInt64(t.Unix(), 600)
	x := float64(t.Unix()-anchor)/600 + float64(t.Nanosecond())/600e9
	var weights [5]float64
	for i := -2; i <= 2; i++ {
		w := 1.
		for j := -2; j <= 2; j++ {
			if j != i {
				w *= (x - float64(j)) / float64(i-j)
			}
		}
		weights[i+2] = w
	}
	return anchor, weights
}
func (sky *transitSky) equatorial(target TransitTarget, t time.Time) [3]float64 {
	anchor, weights := transitInterpolation(t)
	var eq [3]float64
	for i, w := range weights {
		key := transitSkyKey{anchor + int64(i-2)*600, target}
		v, ok := sky.values[key]
		if !ok {
			v = sky.directEquatorial(target, time.Unix(key.seconds, 0))
			sky.values[key] = v
		}
		for k := range eq {
			eq[k] += w * v[k]
		}
	}
	return eq
}
func (sky *transitSky) sidereal(t time.Time) float64 {
	jd := basic.UTC2UT1(basic.Date2JD(t.UTC()))
	if sky.direct {
		return basic.ApparentSiderealTime(jd) * 15 * degreesToRadians
	}
	anchor, weights := transitInterpolation(t)
	correction := 0.
	for i, w := range weights {
		seconds := anchor + int64(i-2)*600
		v, ok := sky.siderealValues[seconds]
		if !ok {
			node := basic.UTC2UT1(basic.Date2JD(time.Unix(seconds, 0).UTC()))
			v = basic.ApparentSiderealTime(node) - basic.MeanSiderealTime(node)
			sky.siderealValues[seconds] = v
		}
		correction += w * v
	}
	return (basic.MeanSiderealTime(jd) + correction) * 15 * degreesToRadians
}
func (sky *transitSky) position(target TransitTarget, t time.Time, _ bool) [3]float64 {
	var eq [3]float64
	if sky.direct {
		eq = sky.directEquatorial(target, t)
	} else {
		eq = sky.equatorial(target, t)
	}
	return rotateZ(sky.sidereal(t), eq)
}
func modInt64(x, m int64) int64 {
	r := x % m
	if r < 0 {
		r += m
	}
	return r
}
func transitRadius(target TransitTarget) float64 {
	if target == TransitSun {
		return 695700
	}
	return 1737.4
}
func alignedTransitTime(t time.Time, step time.Duration) time.Time {
	epoch := time.Unix(0, 0)
	offset := epoch.Sub(epoch.Truncate(step))
	return t.Add(-offset).Truncate(step).Add(offset)
}

func siteTransitSample(orbit *satelliteOrbit, sky *transitSky, target TransitTarget, obs Observer, position [3]float64, t time.Time, exact bool) (transitSample, error) {
	state, err := orbit.stateAt(t)
	if err != nil {
		return transitSample{}, err
	}
	r, _ := earthFixed(state)
	d := subtractVector(r, position)
	v := subtractVector(sky.position(target, t, exact), position)
	_, sa := horizontalDirection(d, obs)
	_, ta := horizontalDirection(v, obs)
	return transitSample{vectorAngle(d, v) / degreesToRadians * 3600, math.Asin(transitRadius(target)/vectorNorm(v)) / degreesToRadians * 3600, sa, ta}, nil
}

type siteTransitScanner struct {
	orbit       *satelliteOrbit
	sky         *transitSky
	observer    Observer
	position    [3]float64
	step        time.Duration
	at          time.Time
	direction   [3]float64
	altitude    float64
	unavailable bool
}

func (scan *siteTransitScanner) sample(target TransitTarget, t time.Time, exact bool) (transitSample, error) {
	if exact {
		return siteTransitSample(scan.orbit, scan.sky, target, scan.observer, scan.position, t, true)
	}
	if !scan.at.Equal(t) {
		state, err := scan.orbit.stateAt(t)
		if err != nil {
			return transitSample{}, err
		}
		r, velocity := earthFixed(state)
		scan.direction = subtractVector(r, scan.position)
		_, scan.altitude = horizontalDirection(scan.direction, scan.observer)
		margin := 2 * scan.step.Seconds() * vectorNorm(velocity) / vectorNorm(scan.direction) / degreesToRadians
		scan.unavailable = scan.altitude+margin+1e-6 < 0
		scan.at = t
	}
	if scan.unavailable && !scan.sky.disableSiteScanSkip {
		return transitSample{separation: math.NaN()}, nil
	}
	v := subtractVector(scan.sky.position(target, t, false), scan.position)
	_, ta := horizontalDirection(v, scan.observer)
	return transitSample{vectorAngle(scan.direction, v) / degreesToRadians * 3600, math.Asin(transitRadius(target)/vectorNorm(v)) / degreesToRadians * 3600, scan.altitude, ta}, nil
}

func transitGolden(lo, hi time.Time, fn func(time.Time) (float64, error), maximum bool, tolerance time.Duration) (time.Time, error) {
	origin := lo
	left, right := 0., hi.Sub(lo).Seconds()
	q := (math.Sqrt(5) - 1) / 2
	at := func(x float64) (float64, error) {
		v, err := fn(origin.Add(time.Duration(x * 1e9)))
		if maximum {
			v = -v
		}
		return v, err
	}
	a, b := right-q*(right-left), left+q*(right-left)
	fa, err := at(a)
	if err != nil {
		return time.Time{}, err
	}
	fb, err := at(b)
	if err != nil {
		return time.Time{}, err
	}
	for right-left > tolerance.Seconds() {
		if fa <= fb {
			right = b
			b = a
			fb = fa
			a = right - q*(right-left)
			fa, err = at(a)
		} else {
			left = a
			a = b
			fa = fb
			b = left + q*(right-left)
			fb, err = at(b)
		}
		if err != nil {
			return time.Time{}, err
		}
	}
	return origin.Add(time.Duration((left + right) * .5e9)), nil
}
func transitBisect(lo, hi time.Time, fn func(time.Time) (float64, error), tolerance time.Duration) (time.Time, error) {
	a, err := fn(lo)
	if err != nil {
		return time.Time{}, err
	}
	if a == 0 {
		return lo, nil
	}
	b, err := fn(hi)
	if err != nil {
		return time.Time{}, err
	}
	if b == 0 {
		return hi, nil
	}
	if a*b > 0 {
		return time.Time{}, ErrInvalidInput
	}
	for hi.Sub(lo) > tolerance {
		mid := lo.Add(hi.Sub(lo) / 2)
		v, err := fn(mid)
		if err != nil {
			return time.Time{}, err
		}
		if v == 0 {
			return mid, nil
		}
		if (v > 0) == (a > 0) {
			lo = mid
			a = v
		} else {
			hi = mid
		}
	}
	return lo.Add(hi.Sub(lo) / 2), nil
}
func findSiteTransits(start, end time.Time, step time.Duration, near, minimum float64, sample transitSampler) ([]siteTransitSpan, error) {
	out, err := findSiteTransitTargets(start, end, step, near, minimum, []transitSampler{sample})
	if err != nil {
		return nil, err
	}
	return out[0], nil
}
func findSiteTransitTargets(start, end time.Time, step time.Duration, near, minimum float64, samples []transitSampler) ([][]siteTransitSpan, error) {
	first := alignedTransitTime(start, step).Add(-step)
	last := alignedTransitTime(end, step).Add(2 * step)
	a, b := make([]transitSample, len(samples)), make([]transitSample, len(samples))
	for n, at := range []time.Time{first, first.Add(step)} {
		for i, sample := range samples {
			v, err := sample(at, false)
			if err != nil {
				return nil, err
			}
			if n == 0 {
				a[i] = v
			} else {
				b[i] = v
			}
		}
	}
	out := make([][]siteTransitSpan, len(samples))
	for t := first.Add(2 * step); !t.After(last); t = t.Add(step) {
		for i, sample := range samples {
			c, err := sample(t, false)
			if err != nil {
				return nil, err
			}
			if b[i].separation <= a[i].separation && b[i].separation <= c.separation && (b[i].separation < a[i].separation || b[i].separation < c.separation) {
				span, valid, err := refineSiteTransit(start, end, t.Add(-2*step), t, step, near, minimum, sample)
				if err != nil {
					return nil, err
				}
				if valid && (len(out[i]) == 0 || span.center.Sub(out[i][len(out[i])-1].center) > time.Millisecond) {
					out[i] = append(out[i], span)
				}
			}
			a[i], b[i] = b[i], c
		}
	}
	return out, nil
}
func refineSiteTransit(start, end, lo, hi time.Time, step time.Duration, near, minimum float64, sample transitSampler) (siteTransitSpan, bool, error) {
	center, err := transitGolden(lo, hi, func(x time.Time) (float64, error) { v, e := sample(x, true); return v.separation * v.separation, e }, false, 500*time.Microsecond)
	if err != nil {
		return siteTransitSpan{}, false, err
	}
	if !center.Before(start) && !center.After(end) {
		v, err := sample(center, true)
		if err != nil {
			return siteTransitSpan{}, false, err
		}
		if v.satelliteAltitude > 0 && v.targetAltitude >= minimum && v.separation <= v.radius+near {
			span := siteTransitSpan{center: center, nearMiss: v.separation > v.radius}
			if !span.nearMiss {
				contact := func(x time.Time) (float64, error) { v, e := sample(x, true); return v.separation - v.radius, e }
				for _, direction := range []time.Duration{-1, 1} {
					inside := center
					for n := 0; ; n++ {
						if n > int(30*24*time.Hour/step) {
							return siteTransitSpan{}, false, ErrInvalidInput
						}
						outside := inside.Add(direction * step)
						margin, err := contact(outside)
						if err != nil {
							return siteTransitSpan{}, false, err
						}
						if margin >= 0 {
							lo, hi := inside, outside
							if hi.Before(lo) {
								lo, hi = hi, lo
							}
							root, err := transitBisect(lo, hi, contact, 500*time.Microsecond)
							if err != nil {
								return siteTransitSpan{}, false, err
							}
							if direction < 0 {
								span.start = root
							} else {
								span.end = root
							}
							break
						}
						inside = outside
					}
				}
			}
			return span, true, nil
		}
	}
	return siteTransitSpan{}, false, nil
}

// Transits 搜索以极小时刻归属闭区间的站点凌日凌月 / search site transits whose minimum times belong to the closed interval.
func (s *Satellite) Transits(e Elements, obs Observer, start, end time.Time, opt TransitOptions) ([]Transit, error) {
	return s.transits(e, obs, start, end, opt, 2*time.Second, newTransitSky(s))
}
func (s *Satellite) transits(e Elements, obs Observer, start, end time.Time, opt TransitOptions, step time.Duration, sky *transitSky) ([]Transit, error) {
	targets, err := transitTargets(opt.Targets)
	if err != nil || !transitInterval(start, end) || !finite(opt.MinTargetAltitude) || math.Abs(opt.MinTargetAltitude) > 90 || !finite(opt.NearMissArcsec) || opt.NearMissArcsec < 0 {
		return nil, ErrInvalidInput
	}
	position, err := observerPosition(obs)
	if err != nil {
		return nil, err
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
	scan := siteTransitScanner{orbit: orbit, sky: sky, observer: obs, position: position, step: step}
	samples := make([]transitSampler, len(targets))
	for i, target := range targets {
		target := target
		samples[i] = func(t time.Time, exact bool) (transitSample, error) { return scan.sample(target, t, exact) }
	}
	spans, err := findSiteTransitTargets(start, end, step, opt.NearMissArcsec, opt.MinTargetAltitude, samples)
	if err != nil {
		return nil, err
	}
	var out []Transit
	for i, target := range targets {
		sample := samples[i]
		for _, span := range spans[i] {
			v, err := sample(span.center, true)
			if err != nil {
				return nil, err
			}
			look, err := s.lookAtOrbit(orbit, obs, position, span.center, ObserveOptions{NoRefraction: true})
			if err != nil {
				return nil, err
			}
			offset, err := transitOffsetAt(orbit, sky, target, span.center, obs)
			if err != nil {
				return nil, err
			}
			out = append(out, Transit{Target: target, Start: span.start, Center: span.center, End: span.end, Duration: span.end.Sub(span.start), NearMiss: span.nearMiss, MinSeparationArcsec: v.separation, TargetRadiusArcsec: v.radius, Satellite: look, TargetAltitude: v.targetAltitude, CenterlineOffsetKM: offset})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Center.Before(out[j].Center) })
	return out, nil
}

func transitUnit(v [3]float64) [3]float64 {
	n := vectorNorm(v)
	return [3]float64{v[0] / n, v[1] / n, v[2] / n}
}
func transitRay(s, k [3]float64) (point [3]float64, disc float64, ok bool) {
	b := earthRadiusKM * (1 - earthFlattening)
	shape := [3]float64{earthRadiusKM, earthRadiusKM, b}
	a, c, d := 0., 0., -1.
	for i := range s {
		u, v := s[i]/shape[i], k[i]/shape[i]
		a += v * v
		c += u * v
		d += u * u
	}
	disc = c*c - a*d
	if disc < 0 || c >= 0 {
		return point, disc, false
	}
	l := (-c - math.Sqrt(disc)) / a
	if l <= 0 {
		return point, disc, false
	}
	for i := range point {
		point[i] = s[i] + l*k[i]
	}
	return point, disc, true
}
func transitGroundAt(orbit *satelliteOrbit, sky *transitSky, target TransitTarget, t time.Time) ([3]float64, error) {
	state, err := orbit.stateAt(t)
	if err != nil {
		return [3]float64{}, err
	}
	r, _ := earthFixed(state)
	v := sky.position(target, t, true)
	g, _, ok := transitRay(r, transitUnit(subtractVector(r, v)))
	if !ok {
		return [3]float64{}, ErrInvalidInput
	}
	return g, nil
}
func transitHeading(orbit *satelliteOrbit, sky *transitSky, target TransitTarget, t time.Time) (float64, bool, error) {
	a, err := transitGeometryAt(orbit, sky, target, t.Add(-100*time.Millisecond), true)
	if err != nil {
		return 0, false, err
	}
	b, err := transitGeometryAt(orbit, sky, target, t.Add(100*time.Millisecond), true)
	if err != nil {
		return 0, false, err
	}
	if a.valid && b.valid {
		al, ap, _ := geographicPosition(a.point)
		bl, bp, _ := geographicPosition(b.point)
		_, heading := transitInverse(al, ap, bl, bp)
		return heading, true, nil
	}
	c, err := transitGeometryAt(orbit, sky, target, t, true)
	if err != nil {
		return 0, false, err
	}
	if !c.valid {
		return 0, false, nil
	}
	for _, dt := range []time.Duration{100 * time.Millisecond, 10 * time.Millisecond, time.Millisecond} {
		if dt != 100*time.Millisecond {
			a, err = transitGeometryAt(orbit, sky, target, t.Add(-dt), true)
			if err != nil {
				return 0, false, err
			}
			b, err = transitGeometryAt(orbit, sky, target, t.Add(dt), true)
			if err != nil {
				return 0, false, err
			}
		}
		if a.valid && b.valid {
			al, ap, _ := geographicPosition(a.point)
			bl, bp, _ := geographicPosition(b.point)
			_, heading := transitInverse(al, ap, bl, bp)
			return heading, true, nil
		}
		if a.valid {
			al, ap, _ := geographicPosition(a.point)
			cl, cp, _ := geographicPosition(c.point)
			_, heading := transitInverse(al, ap, cl, cp)
			return heading, true, nil
		}
		if b.valid {
			cl, cp, _ := geographicPosition(c.point)
			bl, bp, _ := geographicPosition(b.point)
			_, heading := transitInverse(cl, cp, bl, bp)
			return heading, true, nil
		}
	}
	return 0, false, nil
}
func transitOffsetAt(orbit *satelliteOrbit, sky *transitSky, target TransitTarget, t time.Time, obs Observer) (float64, error) {
	g, err := transitGeometryAt(orbit, sky, target, t, true)
	if err != nil {
		return 0, err
	}
	if !g.valid {
		return math.NaN(), nil
	}
	heading, ok, err := transitHeading(orbit, sky, target, t)
	if err != nil {
		return 0, err
	}
	if !ok {
		return math.NaN(), nil
	}
	lon, lat, _ := geographicPosition(g.point)
	distance, azimuth := transitInverse(lon, lat, obs.Longitude, obs.Latitude)
	return distance * math.Sin((heading-azimuth)*degreesToRadians), nil
}
func transitInverse(lon1, lat1, lon2, lat2 float64) (distance, bearing float64) {
	u1, u2 := math.Atan((1-earthFlattening)*math.Tan(lat1*degreesToRadians)), math.Atan((1-earthFlattening)*math.Tan(lat2*degreesToRadians))
	s1, c1 := math.Sincos(u1)
	s2, c2 := math.Sincos(u2)
	distance, east, north := transitInverseReduced(math.Remainder(lon2-lon1, 360)*degreesToRadians, s1, c1, s2, c2, false)
	bearing = math.Atan2(east, north) / degreesToRadians
	if bearing < 0 {
		bearing += 360
	}
	return
}

func transitInverseReduced(l, s1, c1, s2, c2 float64, widthOnly bool) (distance, east, north float64) {
	lambda := l
	var ss, cc, sigma, sa, ca2, cos2 float64
	for i := 0; i < 100; i++ {
		sl, cl := math.Sincos(lambda)
		eastward, northward := c2*sl, c1*s2-s1*c2*cl
		if widthOnly {
			ss = math.Sqrt(eastward*eastward + northward*northward)
		} else {
			ss = math.Hypot(eastward, northward)
		}
		if ss < 1e-15 {
			return 0, 0, 1
		}
		cc = s1*s2 + c1*c2*cl
		sigma = math.Atan2(ss, cc)
		sa = c1 * c2 * sl / ss
		ca2 = 1 - sa*sa
		cos2 = 0
		if ca2 > 1e-15 {
			cos2 = cc - 2*s1*s2/ca2
		}
		c := earthFlattening / 16 * ca2 * (4 + earthFlattening*(4-3*ca2))
		next := l + (1-c)*earthFlattening*sa*(sigma+c*ss*(cos2+c*cc*(-1+2*cos2*cos2)))
		if math.Abs(next-lambda) < 1e-13 {
			lambda = next
			break
		}
		lambda = next
		if i == 99 {
			return math.NaN(), math.NaN(), math.NaN()
		}
	}
	b := earthRadiusKM * (1 - earthFlattening)
	u := ca2 * (earthRadiusKM*earthRadiusKM - b*b) / (b * b)
	a := 1 + u/16384*(4096+u*(-768+u*(320-175*u)))
	bb := u / 1024 * (256 + u*(-128+u*(74-47*u)))
	delta := bb * ss * (cos2 + bb/4*(cc*(-1+2*cos2*cos2)-bb/6*cos2*(-3+4*ss*ss)*(-3+4*cos2*cos2)))
	distance = b * a * (sigma - delta)
	return distance, c2 * math.Sin(lambda), c1*s2 - s1*c2*math.Cos(lambda)
}
