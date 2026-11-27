package geojson

import (
	"encoding/json"
	"math"
	"reflect"
	"time"

	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/satellite"
)

// SatellitePathOptions 凌现象地理输出的时间标记与站点 / time markers and observing site for transit geography.
type SatellitePathOptions struct {
	TimeMarkerStep time.Duration
	Observer       *satellite.Observer
	Nearby         *satellite.NearbyTransit
}

func validSatelliteMarkerStep(start, end time.Time, step time.Duration) bool {
	if step <= 0 {
		return step == 0
	}
	marker := start.Add(step)
	batch := int64(math.MaxInt64) / int64(step)
	for remaining := int64(maximumTimeMarkerCount); remaining > 0 && marker.Before(end); {
		count := batch
		if remaining < count {
			count = remaining
		}
		marker = marker.Add(time.Duration(count) * step)
		remaining -= count
	}
	return !marker.Before(end)
}

type satelliteCollection struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Features   []feature              `json:"features"`
}

func satelliteTarget(target satellite.TransitTarget) string {
	if target == satellite.TransitSun {
		return "sun"
	}
	if target == satellite.TransitMoon {
		return "moon"
	}
	return ""
}
func satellitePathSamples(points []satellite.TransitPathPoint) []pathSample {
	out := make([]pathSample, len(points))
	for i, p := range points {
		out[i] = pathSample{p.Time, p.Longitude, p.Latitude}
	}
	return out
}
func satelliteGeoPoints(points []satellite.GeoPoint) []geodata.GeoPoint {
	out := make([]geodata.GeoPoint, len(points))
	for i, p := range points {
		out[i] = geodata.GeoPoint{Longitude: p.Longitude, Latitude: p.Latitude}
	}
	return out
}
func satelliteSide(distance float64) string {
	if distance > 0 {
		return "left"
	}
	if distance < 0 {
		return "right"
	}
	return "none"
}
func satelliteEncode(features []feature, properties map[string]interface{}) ([]byte, error) {
	out, err := json.Marshal(satelliteCollection{featureCollectionType, properties, features})
	if err != nil {
		return nil, satellite.ErrInvalidInput
	}
	return out, nil
}

// MarshalSatelliteTransitPath 编码凌现象带与可选站点图层 / encode a transit band and optional observing-site layers.
func (g *Generator) MarshalSatelliteTransitPath(path satellite.TransitPath, options SatellitePathOptions) ([]byte, error) {
	event := path.Event
	if len(path.CenterLine) < 2 || len(path.LeftLimit) < 2 || len(path.RightLimit) < 2 || event.Start.IsZero() || event.Greatest.Before(event.Start) || event.End.Before(event.Greatest) || satelliteTarget(event.Target) == "" || !validSatelliteMarkerStep(event.Start, event.End, options.TimeMarkerStep) {
		return nil, satellite.ErrInvalidInput
	}
	if options.Nearby != nil && (options.Observer == nil || !reflect.DeepEqual(options.Nearby.Event, event)) {
		return nil, satellite.ErrInvalidInput
	}
	features := make([]feature, 0, 16)
	for _, line := range []struct {
		role   string
		points []satellite.TransitPathPoint
	}{{"center-line", path.CenterLine}, {"left-limit", path.LeftLimit}, {"right-limit", path.RightLimit}} {
		var err error
		features, err = appendTimedLineFeature(features, "satellite-transit", line.role, satellitePathSamples(line.points), nil)
		if err != nil {
			return nil, satellite.ErrInvalidInput
		}
	}
	ring := make([]geodata.GeoPoint, 0, len(path.LeftLimit)+len(path.RightLimit))
	for _, p := range path.LeftLimit {
		ring = append(ring, geodata.GeoPoint{Longitude: p.Longitude, Latitude: p.Latitude})
	}
	for i := len(path.RightLimit) - 1; i >= 0; i-- {
		p := path.RightLimit[i]
		ring = append(ring, geodata.GeoPoint{Longitude: p.Longitude, Latitude: p.Latitude})
	}
	band, err := multiPolygonGeometryFromFragments(geodata.PolygonFragments(ring, geodata.ClipView{Projection: geodata.ProjectionEquirectangular}))
	if err != nil {
		return nil, satellite.ErrInvalidInput
	}
	features = append([]feature{newFeature("satellite-transit", "transit-band", band, nil)}, features...)
	for _, edge := range []struct {
		name   string
		at     time.Time
		points []satellite.GeoPoint
	}{{"start", event.Start, path.StartHorizon}, {"end", event.End, path.EndHorizon}} {
		value, err := geoMultiLineGeometry(satelliteGeoPoints(edge.points), true)
		if err != nil {
			return nil, satellite.ErrInvalidInput
		}
		features = append(features, newFeature("satellite-transit", "target-horizon", value, map[string]interface{}{"edge": edge.name, "time": formatTime(edge.at)}))
	}
	for _, p := range []struct {
		role  string
		point satellite.TransitPathPoint
	}{{"start", event.StartPoint}, {"greatest", event.GreatestPoint}, {"end", event.EndPoint}} {
		var err error
		features, err = appendPointFeature(features, "satellite-transit", p.role, pathSample{p.point.Time, p.point.Longitude, p.point.Latitude}, nil)
		if err != nil {
			return nil, satellite.ErrInvalidInput
		}
	}
	if options.TimeMarkerStep > 0 {
		points := satellitePathSamples(path.CenterLine)
		index := 1
		for at := event.Start.Add(options.TimeMarkerStep); at.Before(event.End); at = at.Add(options.TimeMarkerStep) {
			for index < len(points)-1 && points[index].Time.Before(at) {
				index++
			}
			a, b := points[index-1], points[index]
			fraction := float64(at.Sub(a.Time)) / float64(b.Time.Sub(a.Time))
			p := interpolateTimeMarker(a, b, fraction, at)
			features, err = appendPointFeature(features, "satellite-transit", "time-marker", p, map[string]interface{}{"label": at.Format("15:04:05")})
			if err != nil {
				return nil, satellite.ErrInvalidInput
			}
		}
	}
	if options.Observer != nil {
		obs := options.Observer
		if math.IsNaN(obs.Height) || math.IsInf(obs.Height, 0) {
			return nil, satellite.ErrInvalidInput
		}
		value, err := pointGeometry(obs.Longitude, obs.Latitude)
		if err != nil {
			return nil, satellite.ErrInvalidInput
		}
		features = append(features, newFeature("satellite-transit", "observer", value, nil))
	}
	if n := options.Nearby; n != nil {
		if n.NearestDistanceKM < 0 || math.IsNaN(n.DistanceKM) || math.IsInf(n.DistanceKM, 0) || math.IsNaN(n.NearestDistanceKM) || math.IsInf(n.NearestDistanceKM, 0) {
			return nil, satellite.ErrInvalidInput
		}
		prop := map[string]interface{}{"distance_km": n.NearestDistanceKM, "side": satelliteSide(n.DistanceKM)}
		features, err = appendPointFeature(features, "satellite-transit", "nearest-point", pathSample{n.Nearest.Time, n.Nearest.Longitude, n.Nearest.Latitude}, prop)
		if err != nil {
			return nil, satellite.ErrInvalidInput
		}
		value, err := geoMultiLineGeometry([]geodata.GeoPoint{{Longitude: options.Observer.Longitude, Latitude: options.Observer.Latitude}, {Longitude: n.Nearest.Longitude, Latitude: n.Nearest.Latitude}}, false)
		if err != nil {
			return nil, satellite.ErrInvalidInput
		}
		features = append(features, newFeature("satellite-transit", "nearest-connector", value, map[string]interface{}{"distance_km": n.NearestDistanceKM, "side": satelliteSide(n.DistanceKM), "time": formatTime(n.Nearest.Time)}))
	}
	return satelliteEncode(features, map[string]interface{}{"event": "satellite-transit", "norad": path.Elements.CatalogNumber, "name": path.Elements.Name, "elements_epoch": formatTime(path.Elements.Epoch), "target": satelliteTarget(event.Target), "complete": event.Complete, "greatest": formatTime(event.Greatest), "min_target_altitude": event.MinTargetAltitude})
}

// MarshalSatelliteTransitShadow 编码瞬时凌现象足迹，无效快照为空集合 / encode an instantaneous transit footprint, empty for an invalid snapshot.
func (g *Generator) MarshalSatelliteTransitShadow(shadow satellite.TransitShadow) ([]byte, error) {
	if shadow.Time.IsZero() || satelliteTarget(shadow.Target) == "" {
		return nil, satellite.ErrInvalidInput
	}
	features := make([]feature, 0, 3)
	if shadow.Valid {
		for _, p := range shadow.Footprint {
			if validateCoordinate(p.Longitude, p.Latitude) != nil {
				return nil, satellite.ErrInvalidInput
			}
		}
		shape, err := multiPolygonGeometryFromFragments(geodata.PolygonFragments(satelliteGeoPoints(shadow.Footprint), geodata.ClipView{Projection: geodata.ProjectionEquirectangular}))
		if err != nil {
			return nil, satellite.ErrInvalidInput
		}
		features = append(features, newFeature("satellite-transit", "current-transit-footprint", shape, map[string]interface{}{"time": formatTime(shadow.Time)}))
		for _, p := range []struct {
			role     string
			lon, lat float64
		}{{"current-transit-point", shadow.Point.Longitude, shadow.Point.Latitude}, {"subsatellite-point", shadow.SubPoint.Longitude, shadow.SubPoint.Latitude}} {
			features, err = appendPointFeature(features, "satellite-transit", p.role, pathSample{shadow.Time, p.lon, p.lat}, nil)
			if err != nil {
				return nil, satellite.ErrInvalidInput
			}
		}
	}
	return satelliteEncode(features, map[string]interface{}{"event": "satellite-transit", "target": satelliteTarget(shadow.Target), "time": formatTime(shadow.Time)})
}
