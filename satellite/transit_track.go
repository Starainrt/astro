package satellite

import (
	"math"
	"sort"
	"time"

	"github.com/starainrt/astro/basic"
)

// TransitTrack 站心凌现象的盘面轨迹与月相 / local transit track and lunar appearance.
type TransitTrack struct {
	Elements                Elements
	Observer                Observer
	Transit                 Transit
	MoonIlluminatedFraction float64
	MoonBrightLimbAngle     float64
	Frames                  []TransitFrame
}

// TransitFrame 目标中心切平面上的东、北偏移及几何高度 / east and north offsets on the target tangent plane with geometric altitudes.
type TransitFrame struct {
	Time               time.Time
	XArcsec, YArcsec   float64
	TargetRadiusArcsec float64
	SatelliteElevation float64
	TargetAltitude     float64
}

func transitFrameAt(orbit *satelliteOrbit, sky *transitSky, target TransitTarget, obs Observer, position [3]float64, t time.Time) (TransitFrame, error) {
	state, err := orbit.stateAt(t)
	if err != nil {
		return TransitFrame{}, err
	}
	r, _ := earthFixed(state)
	sat := subtractVector(r, position)
	body := subtractVector(sky.position(target, t, true), position)
	distance := vectorNorm(body)
	angle := -sky.sidereal(t)
	s := rotateZ(angle, sat)
	v := transitUnit(rotateZ(angle, body))
	h := math.Hypot(v[0], v[1])
	if h == 0 || distance <= transitRadius(target) {
		return TransitFrame{}, ErrInvalidInput
	}
	east := [3]float64{-v[1] / h, v[0] / h, 0}
	north := [3]float64{-v[2] * v[0] / h, -v[2] * v[1] / h, h}
	den := dotVector(s, v)
	if den <= 0 {
		return TransitFrame{}, ErrInvalidInput
	}
	_, sa := horizontalDirection(sat, obs)
	_, ta := horizontalDirection(body, obs)
	return TransitFrame{Time: t, XArcsec: dotVector(s, east) / den / degreesToRadians * 3600, YArcsec: dotVector(s, north) / den / degreesToRadians * 3600, TargetRadiusArcsec: math.Asin(transitRadius(target)/distance) / degreesToRadians * 3600, SatelliteElevation: sa, TargetAltitude: ta}, nil
}

// TransitFrames 采样站心凌现象轨迹，保留接触帧与输入时区 / sample a local transit, preserving contact frames and the input time zone.
func (s *Satellite) TransitFrames(e Elements, obs Observer, tr Transit, step time.Duration) (TransitTrack, error) {
	if !validTransitTarget(tr.Target) || tr.Center.IsZero() || step < 0 || !finite(tr.TargetRadiusArcsec) || tr.TargetRadiusArcsec <= 0 {
		return TransitTrack{}, ErrInvalidInput
	}
	if !tr.NearMiss && (tr.Start.IsZero() || tr.End.IsZero() || tr.Start.After(tr.Center) || tr.End.Before(tr.Center)) {
		return TransitTrack{}, ErrInvalidInput
	}
	position, err := observerPosition(obs)
	if err != nil {
		return TransitTrack{}, ErrInvalidInput
	}
	orbit, err := newSatelliteOrbit(e)
	if err != nil {
		return TransitTrack{}, err
	}
	sky := newTransitSky(s)
	sample := func(at time.Time) (TransitFrame, error) {
		return transitFrameAt(orbit, sky, tr.Target, obs, position, at)
	}
	a, err := sample(tr.Center.Add(-time.Millisecond))
	if err != nil {
		return TransitTrack{}, err
	}
	b, err := sample(tr.Center.Add(time.Millisecond))
	if err != nil {
		return TransitTrack{}, err
	}
	omega := math.Hypot(b.XArcsec-a.XArcsec, b.YArcsec-a.YArcsec) / .002
	span := 1.5 * tr.TargetRadiusArcsec / omega
	if !tr.NearMiss {
		span = math.Max(span, 1.1*math.Max(tr.Center.Sub(tr.Start).Seconds(), tr.End.Sub(tr.Center).Seconds()))
	}
	if !finite(span) || span <= 0 || span >= float64(1<<62)/1e9 {
		return TransitTrack{}, ErrInvalidInput
	}
	half := time.Duration(span * 1e9)
	times, err := transitFrameTimes(tr, half, step)
	if err != nil {
		return TransitTrack{}, err
	}
	track := TransitTrack{Elements: e, Observer: obs, Transit: tr, MoonIlluminatedFraction: 1, Frames: make([]TransitFrame, 0, len(times))}
	for _, at := range times {
		f, err := sample(at)
		if err != nil {
			return TransitTrack{}, err
		}
		if at.Equal(tr.Center) {
			f.TargetRadiusArcsec = tr.TargetRadiusArcsec
			f.SatelliteElevation = tr.Satellite.Elevation
			f.TargetAltitude = tr.TargetAltitude
		}
		track.Frames = append(track.Frames, f)
	}
	if tr.Target == TransitMoon {
		jd := basic.UTC2TT(basic.Date2JD(tr.Center.UTC()))
		track.MoonIlluminatedFraction = basic.MoonPhaseWithInstance(jd, sky.module.inst)
		track.MoonBrightLimbAngle = basic.MoonTopocentricBrightLimbPositionAngleNWithInstance(jd, obs.Longitude, obs.Latitude, obs.Height, -1, sky.module.inst)
	}
	return track, nil
}

func transitFrameTimes(tr Transit, half, step time.Duration) ([]time.Time, error) {
	if step == 0 {
		step = half / 100
	}
	if step <= 0 {
		return nil, ErrInvalidInput
	}
	start, end := tr.Center.Add(-half), tr.Center.Add(half)
	if end.Sub(start)/step+1 > 2001 {
		return nil, ErrInvalidInput
	}
	type instant struct {
		at    time.Time
		event bool
	}
	nodes := make([]instant, 0, 205)
	for at := start; !at.After(end); at = at.Add(step) {
		nodes = append(nodes, instant{at, false})
	}
	nodes = append(nodes, instant{tr.Center, true})
	if !tr.NearMiss {
		nodes = append(nodes, instant{tr.Start, true}, instant{tr.End, true})
	}
	sort.SliceStable(nodes, func(i, j int) bool { return nodes[i].at.Before(nodes[j].at) })
	tol := time.Microsecond
	if step/2 < tol {
		tol = step / 2
	}
	merged := nodes[:0]
	for _, n := range nodes {
		if len(merged) > 0 {
			prev := merged[len(merged)-1]
			if n.at.Equal(prev.at) {
				if n.event {
					merged[len(merged)-1] = n
				}
				continue
			}
			if (!n.event || !prev.event) && n.at.Sub(prev.at) < tol {
				if n.event {
					merged[len(merged)-1] = n
				}
				continue
			}
		}
		merged = append(merged, n)
	}
	if len(merged) > 2001 {
		return nil, ErrInvalidInput
	}
	at := make([]time.Time, len(merged))
	for i, n := range merged {
		at[i] = n.at
	}
	return at, nil
}
