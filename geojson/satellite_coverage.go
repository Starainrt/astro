package geojson

import (
	"math"
	"time"

	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/satellite"
)

func coveragePolygonFeatures(features []feature, role string, ring []satellite.GeoPoint, at time.Time) ([]feature, error) {
	for _, p := range ring {
		if validateCoordinate(p.Longitude, p.Latitude) != nil {
			return nil, satellite.ErrInvalidInput
		}
	}
	points := satelliteGeoPoints(ring)
	fragments := geodata.PolygonFragments(points, geodata.ClipView{Projection: geodata.ProjectionEquirectangular})
	var value geometry
	var err error
	if role == "night-region" && len(fragments) > 0 && geodata.ClockwiseRegionIsComplement(points) {
		value, err = coverageComplementGeometry(fragments)
	} else {
		value, err = multiPolygonGeometryFromFragments(fragments)
	}
	if err != nil {
		return nil, satellite.ErrInvalidInput
	}
	props := map[string]interface{}{}
	if !at.IsZero() {
		props["time"] = formatTime(at)
	}
	return append(features, newFeature("satellite-coverage", role, value, props)), nil
}

func coverageComplementGeometry(fragments [][]geodata.GeoPoint) (geometry, error) {
	regions, err := geodata.EquirectangularComplement(fragments)
	if err != nil || len(regions) == 0 {
		return geometry{}, satellite.ErrInvalidInput
	}
	coordinates := make([][][][]float64, 0, len(regions))
	for _, region := range regions {
		rings, err := coverageRegionCoordinates(region.Outer, region.Holes)
		if err != nil {
			return geometry{}, err
		}
		coordinates = append(coordinates, rings)
	}
	return geometry{Type: "MultiPolygon", Coordinates: coordinates}, nil
}

// MarshalSatelliteCoverage 编码瞬时卫星可见圈、夜区与目视可见区 / encode an instantaneous footprint, night region, and visual coverage.
func (g *Generator) MarshalSatelliteCoverage(coverage satellite.Coverage) ([]byte, error) {
	if coverage.Time.IsZero() {
		return nil, satellite.ErrInvalidInput
	}
	features, err := appendPointFeature(nil, "satellite-coverage", "subsatellite-point", pathSample{coverage.Time, coverage.SubPoint.Longitude, coverage.SubPoint.Latitude}, nil)
	if err != nil {
		return nil, satellite.ErrInvalidInput
	}
	if len(coverage.Footprint) != 0 {
		features, err = coveragePolygonFeatures(features, "satellite-footprint", coverage.Footprint, coverage.Time)
		if err != nil {
			return nil, err
		}
	}
	switch {
	case coverage.NightIsGlobal:
		if len(coverage.Night) != 0 {
			return nil, satellite.ErrInvalidInput
		}
		value, err := coverageComplementGeometry(nil)
		if err != nil {
			return nil, err
		}
		features = append(features, newFeature("satellite-coverage", "night-region", value, map[string]interface{}{"time": formatTime(coverage.Time)}))
	case len(coverage.Night) != 0:
		features, err = coveragePolygonFeatures(features, "night-region", coverage.Night, coverage.Time)
		if err != nil {
			return nil, err
		}
	}
	if len(coverage.VisibleRegions) != 0 {
		features, err = coverageRegionFeaturesAt(features, "satellite-visible", coverage.VisibleRegions, coverage.Time)
	} else {
		for _, ring := range coverage.Visible {
			features, err = coveragePolygonFeatures(features, "satellite-visible", ring, coverage.Time)
			if err != nil {
				return nil, err
			}
		}
	}
	if err != nil {
		return nil, err
	}
	return satelliteEncode(features, map[string]interface{}{"event": "satellite-coverage", "time": formatTime(coverage.Time)})
}

func groundTrackSamples(points []satellite.GroundPoint) ([]pathSample, error) {
	if len(points) < 1 {
		return nil, satellite.ErrInvalidInput
	}
	out := make([]pathSample, len(points))
	for i, p := range points {
		if p.Time.IsZero() || (i > 0 && !p.Time.After(points[i-1].Time)) || !finiteCoverageValue(p.HeightKM) || !finiteCoverageValue(p.SpeedKMS) || p.SpeedKMS < 0 || validateCoordinate(p.Longitude, p.Latitude) != nil {
			return nil, satellite.ErrInvalidInput
		}
		out[i] = pathSample{p.Time, p.Longitude, p.Latitude}
	}
	return out, nil
}

func finiteCoverageValue(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }

func coverageTrackOptions(features []feature, points []pathSample, options SatellitePathOptions) ([]feature, error) {
	if options.TimeMarkerStep > 0 {
		index := 1
		for at := points[0].Time.Add(options.TimeMarkerStep); at.Before(points[len(points)-1].Time); at = at.Add(options.TimeMarkerStep) {
			for index < len(points)-1 && points[index].Time.Before(at) {
				index++
			}
			a, b := points[index-1], points[index]
			p := interpolateTimeMarker(a, b, float64(at.Sub(a.Time))/float64(b.Time.Sub(a.Time)), at)
			var err error
			features, err = appendPointFeature(features, "satellite-coverage", "time-marker", p, map[string]interface{}{"label": at.Format("15:04:05")})
			if err != nil {
				return nil, satellite.ErrInvalidInput
			}
		}
	}
	if obs := options.Observer; obs != nil {
		if !finiteCoverageValue(obs.Height) {
			return nil, satellite.ErrInvalidInput
		}
		value, err := pointGeometry(obs.Longitude, obs.Latitude)
		if err != nil {
			return nil, satellite.ErrInvalidInput
		}
		features = append(features, newFeature("satellite-coverage", "observer", value, nil))
	}
	return features, nil
}

func coverageTrackFeature(features []feature, event string, samples []pathSample, props map[string]interface{}) ([]feature, error) {
	if len(samples) == 1 {
		return appendPointFeature(features, event, "ground-track", samples[0], props)
	}
	return appendTimedLineFeature(features, event, "ground-track", samples, props)
}

// MarshalSatelliteGroundTrack 编码带逐点时间的卫星地面轨迹 / encode a satellite ground track with per-point times.
func (g *Generator) MarshalSatelliteGroundTrack(points []satellite.GroundPoint, options SatellitePathOptions) ([]byte, error) {
	if len(points) == 0 || options.Nearby != nil || !validSatelliteMarkerStep(points[0].Time, points[len(points)-1].Time, options.TimeMarkerStep) {
		return nil, satellite.ErrInvalidInput
	}
	samples, err := groundTrackSamples(points)
	if err != nil {
		return nil, err
	}
	features, err := coverageTrackFeature(nil, "satellite-coverage", samples, nil)
	if err != nil {
		return nil, satellite.ErrInvalidInput
	}
	features, err = coverageTrackOptions(features, samples, options)
	if err != nil {
		return nil, err
	}
	return satelliteEncode(features, map[string]interface{}{"event": "satellite-coverage"})
}

func coverageRegionFeatures(features []feature, role string, regions []satellite.Region) ([]feature, error) {
	return coverageRegionFeaturesAt(features, role, regions, time.Time{})
}

func coverageRegionFeaturesAt(features []feature, role string, regions []satellite.Region, at time.Time) ([]feature, error) {
	for _, r := range regions {
		holes := make([][]geodata.GeoPoint, len(r.Holes))
		for index, hole := range r.Holes {
			holes[index] = satelliteGeoPoints(hole)
		}
		rings, err := coverageRegionCoordinates(satelliteGeoPoints(r.Outer), holes)
		if err != nil {
			return nil, err
		}
		var props map[string]interface{}
		if !at.IsZero() {
			props = map[string]interface{}{"time": formatTime(at)}
		}
		features = append(features, newFeature("satellite-coverage", role, geometry{Type: "Polygon", Coordinates: rings}, props))
	}
	return features, nil
}

func coverageRegionCoordinates(outer []geodata.GeoPoint, holes [][]geodata.GeoPoint) ([][][]float64, error) {
	rings := make([][][]float64, 0, 1+len(holes))
	for index, ring := range append([][]geodata.GeoPoint{outer}, holes...) {
		coordinates, err := geoJSONRing(ring)
		if err != nil {
			return nil, satellite.ErrInvalidInput
		}
		if index > 0 {
			for left, right := 0, len(coordinates)-1; left < right; left, right = left+1, right-1 {
				coordinates[left], coordinates[right] = coordinates[right], coordinates[left]
			}
		}
		rings = append(rings, coordinates)
	}
	return rings, nil
}

func coverageIllumination(kind satellite.Illumination) string {
	switch kind {
	case satellite.IlluminationSunlit:
		return "sunlit"
	case satellite.IlluminationPenumbra:
		return "penumbra"
	case satellite.IlluminationUmbra:
		return "umbra"
	}
	return ""
}

// MarshalSatelliteCoverageSweep 编码带孔扫掠区、光照轨迹与输出时刻可见圈 / encode swept regions with holes, an illuminated track, and output footprints.
func (g *Generator) MarshalSatelliteCoverageSweep(sweep satellite.CoverageSweep, options SatellitePathOptions) ([]byte, error) {
	if options.Nearby != nil || !validSatelliteMarkerStep(sweep.Start, sweep.End, options.TimeMarkerStep) {
		return nil, satellite.ErrInvalidInput
	}
	samples, err := groundTrackSamples(sweep.Track)
	if err != nil || sweep.Start.IsZero() || sweep.End.Before(sweep.Start) || !sweep.Track[0].Time.Equal(sweep.Start) || !sweep.Track[len(sweep.Track)-1].Time.Equal(sweep.End) || len(sweep.TrackIllumination) != len(samples) || len(sweep.Instants) != len(samples) {
		return nil, satellite.ErrInvalidInput
	}
	for _, kind := range sweep.TrackIllumination {
		if coverageIllumination(kind) == "" {
			return nil, satellite.ErrInvalidInput
		}
	}
	features := make([]feature, 0)
	if len(samples) == 1 {
		name := coverageIllumination(sweep.TrackIllumination[0])
		features, err = coverageTrackFeature(features, "satellite-"+name, samples, map[string]interface{}{"illumination": name})
		if err != nil {
			return nil, satellite.ErrInvalidInput
		}
	}
	for i := 0; i < len(samples)-1; {
		j := i + 1
		for j < len(samples)-1 && sweep.TrackIllumination[j] == sweep.TrackIllumination[i] {
			j++
		}
		illumination := coverageIllumination(sweep.TrackIllumination[i])
		features, err = appendTimedLineFeature(features, "satellite-"+illumination, "ground-track", samples[i:j+1], map[string]interface{}{"illumination": illumination})
		if err != nil {
			return nil, satellite.ErrInvalidInput
		}
		i = j
	}
	for _, item := range []struct {
		role    string
		regions []satellite.Region
	}{{"coverage-sweep", sweep.Swept}, {"visible-sweep", sweep.VisibleSwept}} {
		features, err = coverageRegionFeatures(features, item.role, item.regions)
		if err != nil {
			return nil, err
		}
	}
	for i, instant := range sweep.Instants {
		if !instant.Time.Equal(samples[i].Time) {
			return nil, satellite.ErrInvalidInput
		}
		if len(instant.Footprint) != 0 {
			features, err = coveragePolygonFeatures(features, "current-satellite-footprint", instant.Footprint, instant.Time)
			if err != nil {
				return nil, err
			}
		}
	}
	features, err = coverageTrackOptions(features, samples, options)
	if err != nil {
		return nil, err
	}
	return satelliteEncode(features, map[string]interface{}{"event": "satellite-coverage", "start": formatTime(sweep.Start), "end": formatTime(sweep.End)})
}
