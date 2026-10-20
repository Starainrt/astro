package geojson

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/basic"
)

func marshalPlanetTransitPartial(input planetTransitInput, options PlanetTransitOptions, scale basic.TimeScale, withBase func(map[string]interface{}) map[string]interface{}, module *Generator) ([]byte, error) {
	result := planetTransitResultOf(input)
	label := func(value time.Time) string { return formatTime(basic.LabelIn(scale, value.UTC())) }
	features := make([]feature, 0, 4)
	if !skippedLunarRole(options.SkipRoles, "transit-partial") {
		columns := basic.PlanetTransitVisibilityColumnsWithInstance(result, module.inst)
		longitudes := make([]float64, len(columns))
		intervals := make([][]lunarLatitudeInterval, len(columns))
		for index, column := range columns {
			longitudes[index] = column.Longitude
			for _, interval := range column.Intervals {
				intervals[index] = append(intervals[index], lunarLatitudeInterval{low: interval[0], high: interval[1]})
			}
		}
		value, err := planetTransitRegionGeometry(longitudes, intervals)
		if err != nil {
			return nil, fmt.Errorf("geojson: partial transit: %w", err)
		}
		features = append(features, newFeature(planetTransitEvent, "transit-partial", value,
			withBase(map[string]interface{}{"longitude_step_degrees": 2, "latitude_sample_degrees": 1})))
	}
	if options.IsochroneStep > 0 {
		lines, err := planetTransitIsochroneFeatures(input, options.IsochroneStep, scale, options.SkipRoles, withBase, module)
		if err != nil {
			return nil, err
		}
		features = append(features, lines...)
	}
	point := solarSubsolarPoint(input.greatest, module)
	var err error
	features, err = appendPointFeature(features, planetTransitEvent, "greatest",
		pathSample{Time: basic.LabelIn(scale, input.greatest.UTC()), Longitude: point.Longitude, Latitude: point.Latitude},
		withBase(map[string]interface{}{"greatest": label(input.greatest), "window_start": label(input.windowStart), "window_end": label(input.windowEnd),
			"minimum_separation_arcsec": input.separation, "sun_semidiameter_arcsec": input.sunRadius, "planet_semidiameter_arcsec": input.planetRad}))
	if err != nil {
		return nil, err
	}
	if options.TimeMarkers != nil {
		markers, markerErr := planetTransitTimeMarkerSamples(input.windowStart, input.windowEnd, scale, *options.TimeMarkers, module)
		if markerErr != nil {
			return nil, markerErr
		}
		features, err = appendTimeMarkerPointFeatures(features, planetTransitEvent, "subsolar-track", markers, options.TimeMarkers.Location)
		if err != nil {
			return nil, err
		}
	}
	return marshalFeatureCollectionWithTimeScale(dropFeaturesByRole(features, options.SkipRoles), scale)
}
