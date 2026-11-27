package satellite

import (
	"math"
	"sort"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

// CoverageOptions 可见范围的几何仰角与夜区阈值 / geometric elevation and twilight limits for coverage.
type CoverageOptions struct {
	MinElevation        float64
	TwilightSunAltitude *float64
}

// Coverage 单时刻的可见圈、夜区与目视可见区 / instantaneous footprint, night region, and visual coverage.
type Coverage struct {
	Time             time.Time
	SubPoint         GroundPoint
	Illumination     Illumination
	SunCenterVisible bool
	Footprint        []GeoPoint   // 足迹无可绘制面积时为空 / empty when the footprint has no drawable area.
	Night            []GeoPoint   // 顺时针闭合，夜区在边界右侧 / clockwise closed ring with night on the right.
	NightIsGlobal    bool         // Night 为空时区分全球夜区与无面积夜区 / distinguishes global from zero-area night when Night is empty.
	Visible          [][]GeoPoint // 无孔时的兼容表示 / legacy rings, populated for hole-free intersections.
	VisibleRegions   []Region     // 昼区内孔交集的切缝结果，非空时优先于 Visible / seam-split result for enclosed daylight, preferred over Visible when present.
}

// Region 带孔的地理区域，外环逆时针、孔环顺时针 / geographic region with a counterclockwise outer ring and clockwise holes.
type Region struct {
	Outer []GeoPoint
	Holes [][]GeoPoint
}

// CoverageSweep 区间内的轨迹、输出时刻与带孔扫掠区域 / interval track, output instants, and swept regions with holes.
type CoverageSweep struct {
	Elements          Elements
	Start, End        time.Time
	Track             []GroundPoint
	TrackIllumination []Illumination
	Swept             []Region
	VisibleSwept      []Region
	Instants          []Coverage
}

type coverageBoundary struct {
	vector        [3]float64
	origin        GeoPoint
	threshold     float64
	positive      bool
	bearings      []float64
	points        []GeoPoint
	minimumRadius float64
}

type coverageSnapshot struct {
	coverage      Coverage
	minimumRadius float64
}

const twilightCapResolution = 8e-7
const coverageRadialResolutionKM = 0.0005

// SphericalFootprintRadiusKM 计算半径为6371 km的球形地球上几何可见圈的地表弧长 / compute the geometric footprint radius along a 6371 km Earth sphere.
func SphericalFootprintRadiusKM(heightKM, minElevationDeg float64) (float64, error) {
	if !finite(heightKM) || heightKM < 0 || !finite(minElevationDeg) || minElevationDeg < 0 || minElevationDeg > 90 {
		return 0, ErrInvalidInput
	}
	if heightKM == 0 || minElevationDeg == 90 {
		return 0, nil
	}
	const sphericalEarthRadiusKM = 6371.0
	distanceFromCenter := sphericalEarthRadiusKM + heightKM
	horizonHalfAngleTangent := math.Sqrt(heightKM) / math.Sqrt(distanceFromCenter+sphericalEarthRadiusKM)
	scaledSine := math.Sin(minElevationDeg*degreesToRadians) * (distanceFromCenter / (distanceFromCenter + sphericalEarthRadiusKM))
	scaledCosine := math.Sin((90-minElevationDeg)*degreesToRadians) * horizonHalfAngleTangent
	halfAngleTangent := horizonHalfAngleTangent * (scaledCosine / (scaledSine + math.Hypot(scaledSine, scaledCosine)))
	return 2 * sphericalEarthRadiusKM * math.Atan(halfAngleTangent), nil
}

func validCoverageOptions(opt CoverageOptions) bool {
	return finite(opt.MinElevation) && opt.MinElevation >= 0 && opt.MinElevation <= 90 && (opt.TwilightSunAltitude == nil || finite(*opt.TwilightSunAltitude) && math.Abs(*opt.TwilightSunAltitude) <= 90)
}

func coverageAltitude(vector [3]float64, point GeoPoint) float64 {
	obs := Observer{Longitude: point.Longitude, Latitude: point.Latitude}
	p, _ := observerPosition(obs)
	_, alt := horizontalDirection(subtractVector(vector, p), obs)
	return alt
}

func (b *coverageBoundary) radial(az float64) (GeoPoint, float64, error) {
	sign := 1.0
	if !b.positive {
		sign = -1
	}
	value := func(distance float64) (GeoPoint, float64) {
		p := transitDirect(b.origin.Longitude, b.origin.Latitude, az, distance)
		return p, sign * (coverageAltitude(b.vector, p) - b.threshold)
	}
	lo, hi := 0.0, 20000.0
	_, flo := value(lo)
	_, fhi := value(hi)
	if flo <= 0 || fhi >= 0 {
		return GeoPoint{}, 0, ErrInvalidInput
	}
	x := earthRadiusKM * (math.Acos(math.Min(1, earthRadiusKM/vectorNorm(b.vector)*math.Cos(b.threshold*degreesToRadians))) - b.threshold*degreesToRadians)
	if !b.positive {
		x = earthRadiusKM * (90 + b.threshold) * degreesToRadians
	}
	if x <= lo || x >= hi {
		x = (lo + hi) / 2
	}
	previous, fp := lo, flo
	for i := 0; i < 64; i++ {
		point, f := value(x)
		if math.Abs(f) < 1e-8 {
			_, left := value(math.Max(lo, x-.00025))
			_, right := value(math.Min(hi, x+.00025))
			if left >= 0 && right <= 0 {
				return point, x, nil
			}
		}
		if f > 0 {
			lo = x
			flo = f
		} else {
			hi = x
			fhi = f
		}
		if hi-lo <= coverageRadialResolutionKM {
			x = (lo + hi) / 2
			point, _ = value(x)
			return point, x, nil
		}
		next := x - f*(x-previous)/(f-fp)
		if !finite(next) || next <= lo || next >= hi {
			next = (lo + hi) / 2
		}
		previous, fp = x, f
		x = next
	}
	return GeoPoint{}, 0, ErrInvalidInput
}

func (b *coverageBoundary) edge(a, c float64, p, q GeoPoint, depth int) error {
	dl := math.Remainder(q.Longitude-p.Longitude, 360)
	split := false
	for _, f := range []float64{.25, .5, .75} {
		r := GeoPoint{Longitude: p.Longitude + f*dl, Latitude: p.Latitude + f*(q.Latitude-p.Latitude)}
		if math.Abs(coverageAltitude(b.vector, r)-b.threshold)*3600 > .25 {
			split = true
		}
	}
	if split {
		if depth >= 24 {
			return ErrInvalidInput
		}
		mid := (a + c) / 2
		r, _, err := b.radial(mid)
		if err != nil {
			return err
		}
		if err = b.edge(a, mid, p, r, depth+1); err != nil {
			return err
		}
		return b.edge(mid, c, r, q, depth+1)
	}
	b.bearings = append(b.bearings, c)
	b.points = append(b.points, q)
	return nil
}

func newCoverageBoundary(vector [3]float64, origin GeoPoint, threshold float64, positive bool) (coverageBoundary, error) {
	b := coverageBoundary{vector: vector, origin: origin, threshold: threshold, positive: positive, minimumRadius: math.Inf(1)}
	seeds := make([]GeoPoint, 361)
	for i := 0; i < 360; i++ {
		p, r, err := b.radial(float64(i))
		if err != nil {
			return b, err
		}
		seeds[i] = p
		b.minimumRadius = math.Min(b.minimumRadius, r)
	}
	seeds[360] = seeds[0]
	b.bearings = []float64{0}
	b.points = []GeoPoint{seeds[0]}
	if err := b.edge(0, 1, seeds[0], seeds[1], 0); err != nil {
		return b, err
	}
	capacity := 1 + 360*(len(b.points)-1)
	if capacity > 5761 {
		capacity = 5761
	}
	b.bearings = append(make([]float64, 0, capacity), b.bearings...)
	b.points = append(make([]GeoPoint, 0, capacity), b.points...)
	for i := 1; i < 360; i++ {
		if err := b.edge(float64(i), float64(i+1), seeds[i], seeds[i+1], 0); err != nil {
			return b, err
		}
	}
	return b, nil
}

// CoverageAt 计算单时刻的几何与目视可见范围 / compute instantaneous geometric and visual coverage.
func (s *Satellite) CoverageAt(e Elements, t time.Time, opt CoverageOptions) (Coverage, error) {
	if !validCoverageOptions(opt) {
		return Coverage{}, ErrInvalidInput
	}
	orbit, err := newSatelliteOrbit(e)
	if err != nil {
		return Coverage{}, err
	}
	snap, err := s.coverageAt(orbit, t, opt)
	return snap.coverage, err
}

func (s *Satellite) coverageAt(orbit *satelliteOrbit, t time.Time, opt CoverageOptions) (coverageSnapshot, error) {
	state, err := orbit.stateAt(t)
	if err != nil {
		return coverageSnapshot{}, err
	}
	r, v := earthFixed(state)
	lon, lat, h := geographicPosition(r)
	var footprint coverageBoundary
	if h <= 0 {
		return coverageSnapshot{}, ErrInvalidInput
	}
	if opt.MinElevation < 90 && (90-opt.MinElevation > 1 || h*math.Tan((90-opt.MinElevation)*degreesToRadians) > coverageRadialResolutionKM/2) {
		footprint, err = newCoverageBoundary(r, GeoPoint{Longitude: lon, Latitude: lat}, opt.MinElevation, true)
		if err != nil {
			return coverageSnapshot{}, err
		}
	}
	jd := basic.Date2JD(t.UTC())
	gast := basic.ApparentSiderealTime(basic.UTC2UT1(jd)) * 15 * degreesToRadians
	var sun [3]float64
	if s == nil {
		sun = apparentSunPosition(basic.UTC2TT(jd), gast, nil)
	} else {
		sun = apparentSunPosition(basic.UTC2TT(jd), gast, s.inst)
	}
	twilight := -6.0
	if opt.TwilightSunAltitude != nil {
		twilight = *opt.TwilightSunAltitude
	}
	kind, visible := illuminationAt(r, sun)
	result := Coverage{Time: t, SubPoint: GroundPoint{Time: t, Longitude: lon, Latitude: lat, HeightKM: h, SpeedKMS: vectorNorm(v)}, Illumination: kind, SunCenterVisible: visible, Footprint: footprint.points}
	if 90-math.Abs(twilight) < twilightCapResolution {
		result.NightIsGlobal = twilight > 0
		if result.NightIsGlobal && visible && len(footprint.points) != 0 {
			result.Visible = [][]GeoPoint{footprint.points}
		}
		return coverageSnapshot{coverage: result, minimumRadius: footprint.minimumRadius}, nil
	}
	night, err := newTwilightBoundary(sun, twilight)
	if err != nil {
		return coverageSnapshot{}, err
	}
	result.Night = night.points
	if night.positive {
		result.Night = make([]GeoPoint, len(night.points))
		for index, point := range night.points {
			result.Night[len(night.points)-1-index] = point
		}
	}
	if visible && len(footprint.points) != 0 {
		region, overlay, err := coverageIntersection(&footprint, &night)
		if err != nil {
			return coverageSnapshot{}, err
		}
		if len(overlay) > 0 {
			result.VisibleRegions = overlay
		} else if len(region.Holes) > 0 {
			pieces, err := geodata.SubtractPolygon(coverageGeoRing(region.Outer), coverageGeoRing(region.Holes[0]))
			if err != nil {
				return coverageSnapshot{}, err
			}
			result.VisibleRegions = coverageSatelliteRegions(pieces)
		} else if len(region.Outer) > 0 {
			result.Visible = [][]GeoPoint{region.Outer}
		}
	}
	return coverageSnapshot{coverage: result, minimumRadius: footprint.minimumRadius}, nil
}

func coverageAntipode(point GeoPoint) GeoPoint {
	return GeoPoint{Longitude: math.Remainder(point.Longitude+180, 360), Latitude: -point.Latitude}
}

func newTwilightBoundary(sun [3]float64, twilight float64) (coverageBoundary, error) {
	if twilight > 0 {
		lon, lat, _ := geographicPosition(sun)
		return newCoverageBoundary(sun, GeoPoint{Longitude: lon, Latitude: lat}, twilight, true)
	}
	horizontal := math.Hypot(sun[0], sun[1])
	lon := math.Atan2(sun[1], sun[0]) / degreesToRadians
	origin := coverageAntipode(GeoPoint{Longitude: lon, Latitude: math.Atan2(sun[2], horizontal) / degreesToRadians})
	if twilight < -89 {
		latitude := origin.Latitude * degreesToRadians
		eccentricitySquared := earthFlattening * (2 - earthFlattening)
		for iteration := 0; iteration < 4; iteration++ {
			sine := math.Sin(latitude)
			radius := earthRadiusKM / math.Sqrt(1-eccentricitySquared*sine*sine)
			latitude = math.Atan2(-sun[2]-eccentricitySquared*radius*sine, horizontal)
		}
		origin.Latitude = latitude / degreesToRadians
	}
	return newCoverageBoundary(sun, origin, twilight, false)
}

func coverageIntersection(foot, night *coverageBoundary) (Region, []Region, error) {
	type crossing struct {
		az    float64
		p     GeoPoint
		entry bool
	}
	var crossings []crossing
	margin := func(point GeoPoint) float64 { return night.threshold - coverageAltitude(night.vector, point) }
	values := make([]float64, len(foot.points)-1)
	for index, point := range foot.points[:len(values)] {
		values[index] = margin(point)
	}
	type sample struct{ az, value float64 }
	samples := make([]sample, 0, len(values)+2)
	for index, value := range values {
		samples = append(samples, sample{foot.bearings[index], value})
		previous := (index + len(values) - 1) % len(values)
		next := (index + 1) % len(values)
		if (value >= 0) != (values[previous] >= 0) || (value >= 0) != (values[next] >= 0) {
			continue
		}
		sign := 1.
		if value >= 0 {
			sign = -1
		}
		if sign*value < sign*values[previous] || sign*value < sign*values[next] || value == values[previous] && value == values[next] {
			continue
		}
		lo, hi := foot.bearings[previous], foot.bearings[index+1]
		if index == 0 {
			lo -= 360
		}
		var evalErr error
		at := transitAngleMaximum(lo, hi, func(az float64) float64 {
			point, _, err := foot.radial(az)
			if err != nil {
				evalErr = err
				return math.Inf(-1)
			}
			return sign * margin(point)
		})
		if evalErr != nil {
			return Region{}, nil, evalErr
		}
		point, _, err := foot.radial(at)
		if err != nil {
			return Region{}, nil, err
		}
		if refined := margin(point); (refined >= 0) != (value >= 0) {
			samples = append(samples, sample{math.Mod(at+360, 360), refined})
		}
	}
	sort.Slice(samples, func(first, second int) bool { return samples[first].az < samples[second].az })
	insideCount := 0
	for index, left := range samples {
		right := samples[(index+1)%len(samples)]
		fp, fq := left.value, right.value
		if fp >= 0 {
			insideCount++
		}
		if (fp >= 0) == (fq >= 0) {
			continue
		}
		lo, hi := left.az, right.az
		if hi < lo {
			hi += 360
		}
		for j := 0; j < 40; j++ {
			mid := (lo + hi) / 2
			r, _, err := foot.radial(mid)
			if err != nil {
				return Region{}, nil, err
			}
			value := margin(r)
			if (value >= 0) == (fp >= 0) {
				lo = mid
			} else {
				hi = mid
			}
		}
		az := math.Mod((lo+hi)/2, 360)
		p, _, err := foot.radial(az)
		if err != nil {
			return Region{}, nil, err
		}
		crossings = append(crossings, crossing{az: az, p: p, entry: fp < 0})
	}
	if len(crossings) == 0 {
		if insideCount > 0 {
			region := Region{Outer: foot.points}
			day := night.origin
			if !night.positive {
				day = coverageAntipode(day)
			}
			if coverageAltitude(night.vector, day) > night.threshold && coverageAltitude(foot.vector, day) >= foot.threshold {
				region.Holes = [][]GeoPoint{night.points}
			}
			return region, nil, nil
		}
		if !night.positive && coverageAltitude(foot.vector, night.origin) >= foot.threshold {
			return Region{Outer: night.points}, nil, nil
		}
		return Region{}, nil, nil
	}
	if len(crossings) != 2 {
		if len(crossings)%2 != 0 {
			return Region{}, nil, ErrInvalidInput
		}
		regions, err := coverageOverlayIntersection(foot, night)
		if err != nil {
			return Region{}, nil, err
		}
		return Region{}, regions, nil
	}
	a, b := crossings[0], crossings[1]
	if !a.entry {
		a, b = b, a
	}
	extent := math.Mod(b.az-a.az+360, 360)
	ring := []GeoPoint{a.p}
	for i, az := range foot.bearings[:len(foot.bearings)-1] {
		d := math.Mod(az-a.az+360, 360)
		if d > 0 && d < extent {
			ring = append(ring, foot.points[i])
		}
	}
	// 交集的可见圈弧可能跨越方位零点。
	if b.az < a.az {
		ring = ring[:1]
		for i, az := range foot.bearings[:len(foot.bearings)-1] {
			if az > a.az {
				ring = append(ring, foot.points[i])
			}
		}
		for i, az := range foot.bearings[:len(foot.bearings)-1] {
			if az < b.az {
				ring = append(ring, foot.points[i])
			}
		}
	}
	ring = append(ring, b.p)
	_, begin := transitInverse(night.origin.Longitude, night.origin.Latitude, b.p.Longitude, b.p.Latitude)
	_, finish := transitInverse(night.origin.Longitude, night.origin.Latitude, a.p.Longitude, a.p.Latitude)
	begin = math.Mod(begin+360, 360)
	finish = math.Mod(finish+360, 360)
	span := math.Mod(finish-begin+360, 360)
	mid, _, err := night.radial(begin + span/2)
	if err != nil {
		return Region{}, nil, err
	}
	forward := coverageAltitude(foot.vector, mid) >= foot.threshold
	if forward {
		for _, shift := range []float64{0, 360} {
			for i, az := range night.bearings[:len(night.bearings)-1] {
				if az+shift > begin && az+shift < begin+span {
					ring = append(ring, night.points[i])
				}
			}
		}
	} else {
		for _, shift := range []float64{0, -360} {
			for i := len(night.bearings) - 2; i >= 0; i-- {
				az := night.bearings[i] + shift
				if az < begin && az > begin-(360-span) {
					ring = append(ring, night.points[i])
				}
			}
		}
	}
	ring = append(ring, a.p)
	return Region{Outer: ring}, nil, nil
}

// coverageOverlayIntersection 处理扁球足迹与晨昏圈多交点（≥4）的球面求交。
func coverageOverlayIntersection(foot, night *coverageBoundary) ([]Region, error) {
	footRing := coverageGeoRing(foot.points)
	nightRing := coverageGeoRing(night.points)
	subtractDay := night.positive
	if !subtractDay {
		// 晨昏门限接近 0° 时，视差可能让夜区略超半球，此时小侧为昼区。
		subtractDay = geodata.ClockwiseRegionIsComplement(nightRing)
	}
	var pieces []geodata.Region
	var err error
	if subtractDay {
		pieces, err = geodata.SubtractPolygon(footRing, nightRing)
	} else {
		pieces, err = geodata.IntersectPolygon(footRing, nightRing)
	}
	if err != nil {
		return nil, err
	}
	return coverageSatelliteRegions(pieces), nil
}

func coverageGeoRing(ring []GeoPoint) []geodata.GeoPoint {
	points := make([]geodata.GeoPoint, len(ring))
	for index, point := range ring {
		points[index] = geodata.GeoPoint(point)
	}
	return points
}

func coverageGeoRegions(regions []Region) []geodata.Region {
	result := make([]geodata.Region, len(regions))
	for index, region := range regions {
		result[index].Outer = coverageGeoRing(region.Outer)
		for _, hole := range region.Holes {
			result[index].Holes = append(result[index].Holes, coverageGeoRing(hole))
		}
	}
	return result
}

func coverageRegions(polygons [][]GeoPoint) ([]geodata.Region, error) {
	if len(polygons) == 0 {
		return nil, nil
	}
	inputs := make([][]geodata.GeoPoint, len(polygons))
	for i, ring := range polygons {
		inputs[i] = coverageGeoRing(ring)
		polygons[i] = nil
	}
	return geodata.UnionRegions(inputs)
}

func coverageSatelliteRegions(rs []geodata.Region) []Region {
	if rs == nil {
		return nil
	}
	result := make([]Region, len(rs))
	convert := func(r []geodata.GeoPoint) []GeoPoint {
		v := make([]GeoPoint, len(r))
		for i, p := range r {
			v[i] = GeoPoint{Longitude: p.Longitude, Latitude: p.Latitude}
		}
		return v
	}
	for i, r := range rs {
		result[i].Outer = convert(r.Outer)
		for _, h := range r.Holes {
			result[i].Holes = append(result[i].Holes, convert(h))
		}
	}
	return result
}

// CoverageBetween 采样闭区间并自动加密内部扫掠网格，至多24小时 / sample a closed interval up to 24 hours with internal swept-grid refinement.
func (s *Satellite) CoverageBetween(e Elements, start, end time.Time, step time.Duration, opt CoverageOptions) (CoverageSweep, error) {
	if !validCoverageOptions(opt) || start.IsZero() || end.IsZero() || end.Before(start) || end.Sub(start) > 24*time.Hour || step < 10*time.Second || step > 10*time.Minute {
		return CoverageSweep{}, ErrInvalidInput
	}
	orbit, err := newSatelliteOrbit(e)
	if err != nil {
		return CoverageSweep{}, err
	}
	times := []time.Time{start}
	for t := alignedTransitTime(start, step).Add(step); t.Before(end); t = t.Add(step) {
		times = append(times, t)
		if len(times) > 2000 {
			return CoverageSweep{}, ErrInvalidInput
		}
	}
	if !end.Equal(start) {
		times = append(times, end)
	}
	if len(times) > 2000 {
		return CoverageSweep{}, ErrInvalidInput
	}
	sweep := CoverageSweep{Elements: e, Start: start, End: end}
	var footprints, visibles [][]GeoPoint
	var visibleSets [][]geodata.Region
	count := len(times)
	add := func(snap coverageSnapshot) {
		if len(snap.coverage.Footprint) == 0 {
			return
		}
		footprints = append(footprints, snap.coverage.Footprint)
		if len(snap.coverage.VisibleRegions) != 0 {
			visibleSets = append(visibleSets, coverageGeoRegions(snap.coverage.VisibleRegions))
		} else {
			visibles = append(visibles, snap.coverage.Visible...)
		}
	}
	var refine func(coverageSnapshot, coverageSnapshot) error
	refine = func(a, b coverageSnapshot) error {
		if opt.MinElevation == 90 {
			return nil
		}
		minimumRadius := math.Min(a.minimumRadius, b.minimumRadius)
		if minimumRadius == 0 {
			minimumRadius = math.Max(a.minimumRadius, b.minimumRadius)
		}
		p, q := a.coverage.SubPoint, b.coverage.SubPoint
		d, _ := transitInverse(p.Longitude, p.Latitude, q.Longitude, q.Latitude)
		if minimumRadius > 0 && d <= .35*minimumRadius {
			add(b)
			return nil
		}
		if b.coverage.Time.Sub(a.coverage.Time) <= time.Millisecond || count >= 20000 {
			return ErrInvalidInput
		}
		count++
		m, err := s.coverageAt(orbit, a.coverage.Time.Add(b.coverage.Time.Sub(a.coverage.Time)/2), opt)
		if err != nil {
			return err
		}
		if minimumRadius == 0 && m.minimumRadius == 0 {
			return nil
		}
		if err = refine(a, m); err != nil {
			return err
		}
		return refine(m, b)
	}
	var previous coverageSnapshot
	for i, t := range times {
		snap, err := s.coverageAt(orbit, t, opt)
		if err != nil {
			return CoverageSweep{}, err
		}
		sweep.Instants = append(sweep.Instants, snap.coverage)
		sweep.Track = append(sweep.Track, snap.coverage.SubPoint)
		sweep.TrackIllumination = append(sweep.TrackIllumination, snap.coverage.Illumination)
		if i == 0 {
			add(snap)
		} else if err = refine(previous, snap); err != nil {
			return CoverageSweep{}, err
		}
		previous = snap
	}
	swept, err := coverageRegions(footprints)
	if err != nil {
		return CoverageSweep{}, err
	}
	sweep.Swept = coverageSatelliteRegions(swept)
	visual, err := coverageRegions(visibles)
	if err != nil {
		return CoverageSweep{}, err
	}
	if len(visibleSets) != 0 {
		visual, err = geodata.UnionRegionSets(append(visibleSets, visual))
		if err != nil {
			return CoverageSweep{}, err
		}
	}
	sweep.VisibleSwept = coverageSatelliteRegions(visual)
	return sweep, nil
}
