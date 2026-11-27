package satellite

import (
	"math"
	"time"
)

// TransitShadow 凌现象带的瞬时中心点、边界与星下点 / instantaneous transit centre, footprint, and subsatellite point.
type TransitShadow struct {
	Time      time.Time
	Valid     bool
	Target    TransitTarget
	Point     TransitPathPoint
	Footprint []GeoPoint
	SubPoint  GroundPoint
}

// TransitShadowAt 计算任意时刻的凌现象快照，边缘未全部落地时无效 / compute a transit snapshot, invalid unless the entire limb reaches Earth.
func (s *Satellite) TransitShadowAt(e Elements, target TransitTarget, t time.Time, opt TransitSearchOptions) (TransitShadow, error) {
	if !validTransitTarget(target) || !validTransitSearch(opt) || t.IsZero() {
		return TransitShadow{}, ErrInvalidInput
	}
	orbit, err := newSatelliteOrbit(e)
	if err != nil {
		return TransitShadow{}, err
	}
	sky := newTransitSky(s)
	sky.direct = true
	g, err := transitGeometryAt(orbit, sky, target, t, true)
	if err != nil {
		return TransitShadow{}, err
	}
	invalid := TransitShadow{Time: t}
	if !g.valid || g.altitude < opt.MinTargetAltitude {
		return invalid, nil
	}
	footprint := make([]GeoPoint, 0, 73)
	limb := newTransitLimbProjection(g, target)
	for i := 0; i < 72; i++ {
		p, ok := limb.point(float64(i) * 2 * math.Pi / 72)
		if !ok {
			return invalid, nil
		}
		lon, lat, _ := geographicPosition(p)
		footprint = append(footprint, GeoPoint{lon, lat})
	}
	footprint = append(footprint, footprint[0])
	heading, ok, err := transitHeading(orbit, sky, target, t)
	if err != nil {
		return TransitShadow{}, err
	}
	if !ok {
		return invalid, nil
	}
	width, ok := transitWidth(g, target, heading)
	if !ok {
		return invalid, nil
	}
	lon, lat, _ := geographicPosition(g.point)
	sub, err := orbit.groundAt(t)
	if err != nil {
		return TransitShadow{}, err
	}
	return TransitShadow{
		Time: t, Valid: true, Target: target,
		Point:     TransitPathPoint{Time: t, Longitude: lon, Latitude: lat, HalfWidthKM: width, TargetAltitude: g.altitude},
		Footprint: footprint, SubPoint: sub,
	}, nil
}
