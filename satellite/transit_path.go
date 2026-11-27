package satellite

import (
	"math"
	"sort"
	"time"
)

// TransitPath 凌现象带的中心线、左右限与起止目标高度边界 / transit centre line, lateral limits, and endpoint target-altitude boundaries.
type TransitPath struct {
	Elements                          Elements
	Event                             TransitEvent
	CenterLine, LeftLimit, RightLimit []TransitPathPoint
	StartHorizon, EndHorizon          []GeoPoint
}

func transitDirect(lon, lat, bearing, distance float64) GeoPoint {
	alpha := bearing * degreesToRadians
	phi := lat * degreesToRadians
	u := math.Atan((1 - earthFlattening) * math.Tan(phi))
	su, cu := math.Sincos(u)
	sa := cu * math.Sin(alpha)
	ca2 := 1 - sa*sa
	b := earthRadiusKM * (1 - earthFlattening)
	sigma1 := math.Atan2(math.Tan(u), math.Cos(alpha))
	usq := ca2 * (earthRadiusKM*earthRadiusKM - b*b) / (b * b)
	aa := 1 + usq/16384*(4096+usq*(-768+usq*(320-175*usq)))
	bb := usq / 1024 * (256 + usq*(-128+usq*(74-47*usq)))
	sigma := distance / (b * aa)
	for i := 0; i < 100; i++ {
		cs := math.Cos(2*sigma1 + sigma)
		ss, cc := math.Sincos(sigma)
		delta := bb * ss * (cs + bb/4*(cc*(-1+2*cs*cs)-bb/6*cs*(-3+4*ss*ss)*(-3+4*cs*cs)))
		next := distance/(b*aa) + delta
		if math.Abs(next-sigma) < 1e-13 {
			sigma = next
			break
		}
		sigma = next
	}
	ss, cc := math.Sincos(sigma)
	tmp := su*ss - cu*cc*math.Cos(alpha)
	p := math.Atan2(su*cc+cu*ss*math.Cos(alpha), (1-earthFlattening)*math.Sqrt(sa*sa+tmp*tmp))
	lambda := math.Atan2(ss*math.Sin(alpha), cu*cc-su*ss*math.Cos(alpha))
	c := earthFlattening / 16 * ca2 * (4 + earthFlattening*(4-3*ca2))
	cs := math.Cos(2*sigma1 + sigma)
	l := lambda - (1-c)*earthFlattening*sa*(sigma+c*ss*(cs+c*cc*(-1+2*cs*cs)))
	longitude := math.Remainder(lon+l/degreesToRadians, 360)
	if longitude <= -180 {
		longitude = 180
	}
	return GeoPoint{longitude, p / degreesToRadians}
}
func transitHorizon(target [3]float64, minimum float64, join TransitPathPoint) ([]GeoPoint, error) {
	lon, lat, _ := geographicPosition(target)
	bearing := 0.
	joined := math.Abs(join.TargetAltitude-minimum) < 1./3600
	if joined {
		_, bearing = transitInverse(lon, lat, join.Longitude, join.Latitude)
	}
	point := func(az float64) GeoPoint {
		lo, hi := 0., 20000.
		for i := 0; i < 32; i++ {
			mid := (lo + hi) / 2
			p := transitDirect(lon, lat, az, mid)
			r, _ := observerPosition(Observer{Longitude: p.Longitude, Latitude: p.Latitude})
			_, alt := horizontalDirection(subtractVector(target, r), Observer{Longitude: p.Longitude, Latitude: p.Latitude})
			if alt > minimum {
				lo = mid
			} else {
				hi = mid
			}
		}
		return transitDirect(lon, lat, az, (lo+hi)/2)
	}
	residual := func(p GeoPoint) float64 {
		r, _ := observerPosition(Observer{Longitude: p.Longitude, Latitude: p.Latitude})
		_, alt := horizontalDirection(subtractVector(target, r), Observer{Longitude: p.Longitude, Latitude: p.Latitude})
		return math.Abs(alt-minimum) * 3600
	}
	first := point(bearing)
	if joined {
		first = GeoPoint{join.Longitude, join.Latitude}
	}
	out := []GeoPoint{first}
	var edge func(float64, float64, GeoPoint, GeoPoint, int) error
	edge = func(a, b float64, p, q GeoPoint, depth int) error {
		dl := math.Remainder(q.Longitude-p.Longitude, 360)
		split := false
		for _, f := range []float64{.25, .5, .75} {
			x := GeoPoint{p.Longitude + f*dl, p.Latitude + f*(q.Latitude-p.Latitude)}
			if residual(x) > .25 {
				split = true
			}
		}
		if split {
			if depth >= 24 {
				return ErrInvalidInput
			}
			mid := (a + b) / 2
			r := point(mid)
			if err := edge(a, mid, p, r, depth+1); err != nil {
				return err
			}
			return edge(mid, b, r, q, depth+1)
		}
		out = append(out, q)
		return nil
	}
	for i := 0; i < 360; i++ {
		a := bearing + float64(i)
		b := a + 1
		p := out[len(out)-1]
		q := point(b)
		if i == 359 {
			q = first
		}
		if err := edge(a, b, p, q, 0); err != nil {
			return nil, err
		}
	}
	return out, nil
}
func transitPathPolynomialRoots(c []float64) []float64 {
	for len(c) > 1 && c[len(c)-1] == 0 {
		c = c[:len(c)-1]
	}
	if len(c) < 2 {
		return nil
	}
	value := func(x float64) float64 {
		v := c[len(c)-1]
		for i := len(c) - 2; i >= 0; i-- {
			v = v*x + c[i]
		}
		return v
	}
	d := make([]float64, len(c)-1)
	for i := range d {
		d[i] = float64(i+1) * c[i+1]
	}
	cuts := append([]float64{-1}, transitPathPolynomialRoots(d)...)
	cuts = append(cuts, 1)
	scale := 0.
	for _, v := range c {
		scale += math.Abs(v)
	}
	var roots []float64
	for i, a := range cuts {
		fa := value(a)
		if math.Abs(fa) <= 2e-15*scale {
			roots = append(roots, a)
		}
		if i+1 == len(cuts) {
			break
		}
		b := cuts[i+1]
		fb := value(b)
		if fa*fb >= 0 {
			continue
		}
		for b-a > 2e-14 {
			mid := (a + b) / 2
			fm := value(mid)
			if (fa < 0) == (fm < 0) {
				a, fa = mid, fm
			} else {
				b = mid
			}
		}
		roots = append(roots, (a+b)/2)
	}
	sort.Float64s(roots)
	return roots
}

type transitPathCone struct {
	satellite, q, axis, cosine, sine [3]float64
	outside                          float64
}

func newTransitPathCone(g transitGeometry, target TransitTarget) transitPathCone {
	u := transitUnit(subtractVector(g.target, g.satellite))
	east := transitUnit(transitCross([3]float64{0, 0, 1}, u))
	north := transitCross(u, east)
	sa := transitRadius(target) / vectorNorm(subtractVector(g.target, g.satellite))
	ca := math.Sqrt(1 - sa*sa)
	cone := transitPathCone{satellite: g.satellite, outside: -1}
	for i := range u {
		scale := earthRadiusKM
		if i == 2 {
			scale *= 1 - earthFlattening
		}
		cone.q[i] = g.satellite[i] / scale
		cone.outside += cone.q[i] * cone.q[i]
		cone.axis[i], cone.cosine[i], cone.sine[i] = -ca*u[i], -sa*east[i], -sa*north[i]
	}
	return cone
}

func (c transitPathCone) tangencies() []float64 {
	var angles []float64
	for _, sign := range []float64{1, -1} {
		var v [3][3]float64
		var dot [3]float64
		for i := range c.q {
			scale := 1.
			if i == 2 {
				scale /= 1 - earthFlattening
			}
			v[0][i] = (c.axis[i] + sign*c.cosine[i]) * scale
			v[1][i] = 2 * sign * c.sine[i] * scale
			v[2][i] = (c.axis[i] - sign*c.cosine[i]) * scale
			for j := range dot {
				dot[j] += c.q[i] * v[j][i]
			}
		}
		polynomial := make([]float64, 5)
		for j := range v {
			for k := range v {
				x := dot[j] * dot[k]
				for i := range c.q {
					x -= c.outside * v[j][i] * v[k][i]
				}
				polynomial[j+k] += x
			}
		}
		for _, root := range transitPathPolynomialRoots(polynomial) {
			angle := 2 * math.Atan(root)
			if sign < 0 {
				angle += math.Pi
			}
			if angle < 0 {
				angle += 2 * math.Pi
			}
			angles = append(angles, angle)
		}
	}
	sort.Float64s(angles)
	return angles
}

func (c transitPathCone) point(pa float64, tangent bool) (point [3]float64, ok bool) {
	s, co := math.Sincos(pa)
	var k [3]float64
	a, b := 0., 0.
	for i := range k {
		k[i] = c.axis[i] + co*c.cosine[i] + s*c.sine[i]
		v := k[i]
		if i == 2 {
			v /= 1 - earthFlattening
		}
		a += v * v
		b += c.q[i] * v
	}
	disc := math.FMA(b, b, -a*c.outside)
	if b >= 0 || (!tangent && disc < 0) {
		return point, false
	}
	l := -b / a
	if !tangent {
		l -= math.Sqrt(disc) / a
	}
	for i := range point {
		point[i] = c.satellite[i] + earthRadiusKM*l*k[i]
	}
	return point, true
}

type transitWidthProjector struct {
	longitude, sinReducedLat, cosReducedLat, sinHeading, cosHeading float64
}

func newTransitWidthProjector(longitude, latitude, heading float64) transitWidthProjector {
	sinLat, cosLat := math.Sincos(math.Atan((1 - earthFlattening) * math.Tan(latitude*degreesToRadians)))
	sinHeading, cosHeading := math.Sincos(heading * degreesToRadians)
	return transitWidthProjector{longitude * degreesToRadians, sinLat, cosLat, sinHeading, cosHeading}
}

func (projector transitWidthProjector) crossTrack(point [3]float64) float64 {
	longitude := math.Atan2(point[1], point[0])
	radial := math.Sqrt(point[0]*point[0] + point[1]*point[1])
	polar := point[2] / (1 - earthFlattening)
	scale := math.Sqrt(radial*radial + polar*polar)
	deltaLon := longitude - projector.longitude
	if deltaLon > math.Pi {
		deltaLon -= 2 * math.Pi
	} else if deltaLon < -math.Pi {
		deltaLon += 2 * math.Pi
	}
	distance, east, north := transitInverseReduced(deltaLon, projector.sinReducedLat, projector.cosReducedLat, polar/scale, radial/scale, true)
	return distance * (projector.sinHeading*north - projector.cosHeading*east) / math.Sqrt(east*east+north*north)
}

func transitPathLimits(g transitGeometry, target TransitTarget, p TransitPathPoint, heading float64) (TransitPathPoint, TransitPathPoint, float64) {
	points, best := transitPathExtrema(newTransitPathCone(g, target), func(point [3]float64) float64 {
		lon, lat, _ := geographicPosition(point)
		distance, bearing := transitInverse(p.Longitude, p.Latitude, lon, lat)
		return distance * math.Sin((heading-bearing)*degreesToRadians)
	})
	limits := [2]TransitPathPoint{p, p}
	for side, value := range best {
		if value != math.Inf(-1) {
			point := points[side]
			lon, lat, _ := geographicPosition(point)
			_, alt := horizontalDirection(subtractVector(g.target, point), Observer{Longitude: lon, Latitude: lat})
			limits[side] = TransitPathPoint{Time: p.Time, Longitude: lon, Latitude: lat, HalfWidthKM: p.HalfWidthKM, TargetAltitude: alt}
		}
	}
	return limits[0], limits[1], math.Max(0, math.Max(best[0], best[1]))
}

func transitPathExtrema(cone transitPathCone, project func([3]float64) float64) ([2][3]float64, [2]float64) {
	roots := cone.tangencies()
	cuts := append([]float64{0}, roots...)
	cuts = append(cuts, 2*math.Pi)
	score := func(pa float64, tangent bool) (float64, [3]float64, bool) {
		v, ok := cone.point(pa, tangent)
		if !ok {
			return 0, v, false
		}
		return project(v), v, true
	}
	var extrema [2][3]float64
	best := [2]float64{math.Inf(-1), math.Inf(-1)}
	take := func(pa float64, tangent bool) float64 {
		x, r, ok := score(pa, tangent)
		if !ok {
			return math.NaN()
		}
		for side, sign := range []float64{1, -1} {
			if sign*x > best[side] {
				best[side] = sign * x
				extrema[side] = r
			}
		}
		return x
	}
	for _, root := range roots {
		take(root, true)
	}
	for i := 1; i < len(cuts); i++ {
		a, b := cuts[i-1], cuts[i]
		if _, ok := cone.point((a+b)/2, false); !ok {
			continue
		}
		n := int(math.Ceil((b - a) / (2 * math.Pi / 24)))
		if n < 4 {
			n = 4
		}
		values := make([]float64, n+1)
		for j := range values {
			angle := a + (b-a)*float64(j)/float64(n)
			values[j] = take(angle, (j == 0 && i > 1) || (j == n && i < len(cuts)-1))
		}
		for j := 0; j <= n; j++ {
			for _, sign := range []float64{1, -1} {
				if j == 0 || j == n || (sign*values[j] >= sign*values[j-1] && sign*values[j] >= sign*values[j+1]) {
					first, last := j-1, j+1
					if first < 0 {
						first = 0
					}
					if last > n {
						last = n
					}
					lo, hi := a+(b-a)*float64(first)/float64(n), a+(b-a)*float64(last)/float64(n)
					opt := transitAngleMaximum(lo, hi, func(z float64) float64 {
						x, _, ok := score(z, false)
						if !ok {
							return math.Inf(-1)
						}
						return sign * x
					})
					take(opt, false)
				}
			}
		}
	}
	return extrema, best
}

// TransitEventPath 采样凌现象带，零步长为1秒 / sample a transit band, using one second for a zero step.
func (s *Satellite) TransitEventPath(e Elements, event TransitEvent, step time.Duration) (TransitPath, error) {
	if step == 0 {
		step = time.Second
	}
	if step < 10*time.Millisecond || step > time.Minute || !transitInterval(event.Start, event.End) || event.Greatest.Before(event.Start) || event.Greatest.After(event.End) || !validTransitTarget(event.Target) || !validTransitSearch(TransitSearchOptions{MinTargetAltitude: event.MinTargetAltitude}) {
		return TransitPath{}, ErrInvalidInput
	}
	count := event.End.Sub(event.Start)/step + 1
	if event.End.Sub(event.Start)%step != 0 {
		count++
	}
	if event.StartPoint.Time != event.Start || event.EndPoint.Time != event.End || event.GreatestPoint.Time != event.Greatest || count > 200000 {
		return TransitPath{}, ErrInvalidInput
	}
	orbit, err := newSatelliteOrbit(e)
	if err != nil {
		return TransitPath{}, err
	}
	sky := newTransitSky(s)
	out := TransitPath{Elements: e, Event: event}
	for t := event.Start; ; {
		var p TransitPathPoint
		if t.Equal(event.Start) {
			p = event.StartPoint
		} else if t.Equal(event.End) {
			p = event.EndPoint
		} else {
			p, err = transitEventPoint(orbit, sky, event.Target, t, event.Start, event.End, false)
			if err != nil {
				return TransitPath{}, err
			}
		}
		g, err := transitGeometryAt(orbit, sky, event.Target, t, true)
		if err != nil {
			return TransitPath{}, err
		}
		heading, err := transitEventHeading(orbit, sky, event.Target, t, event.Start, event.End)
		if err != nil {
			return TransitPath{}, err
		}
		left, right, _ := transitPathLimits(g, event.Target, p, heading)
		out.CenterLine = append(out.CenterLine, p)
		out.LeftLimit = append(out.LeftLimit, left)
		out.RightLimit = append(out.RightLimit, right)
		if t.Equal(event.End) {
			break
		}
		t = t.Add(step)
		if t.After(event.End) {
			t = event.End
		}
	}
	out.StartHorizon, err = transitHorizon(sky.position(event.Target, event.Start, true), event.MinTargetAltitude, event.StartPoint)
	if err != nil {
		return TransitPath{}, err
	}
	out.EndHorizon, err = transitHorizon(sky.position(event.Target, event.End, true), event.MinTargetAltitude, event.EndPoint)
	if err != nil {
		return TransitPath{}, err
	}
	return out, nil
}
