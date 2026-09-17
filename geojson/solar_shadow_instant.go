package geojson

import (
	"fmt"

	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/geodata"
)

// MarshalSolarEclipseShadowInstant 单时刻阴影足迹的最小 GeoJSON：只含该时刻的区域与物理边界 / one instant as the smallest useful GeoJSON.
func MarshalSolarEclipseShadowInstant(
	instant eclipsecore.SolarEclipseShadowInstant,
) ([]byte, error) {
	if instant.Empty() {
		return marshalEmptyFeatureCollection()
	}
	curve, err := solarShadowFootprintCurveFromSegments(instant.Boundaries)
	if err != nil {
		return nil, fmt.Errorf("geojson: solar shadow instant at %s: %w", formatTime(instant.Time), err)
	}
	if len(curve) < 2 {
		return marshalEmptyFeatureCollection()
	}
	properties := map[string]interface{}{
		"model":            string(instant.Model),
		"delta_t_seconds":  instant.DeltaTSeconds,
		"interp_signature": instant.Topology.Signature(),
	}
	regionRole, boundaryRole := solarCentralShadowFootprintRole, solarCentralShadowBoundaryRole
	if instant.Kind == eclipsecore.SolarEclipseShadowPenumbra {
		regionRole, boundaryRole = solarPartialFootprintRole, solarPartialFootprintBoundaryRole
	}
	features := make([]feature, 0, 2)
	regionProperties := cloneProperties(properties)
	regionProperties["time"] = formatTime(instant.Time)
	if instant.Closed {
		// 与整包导出同构：自身闭合的区域不带 geometry_role，也不带 closure。
		regionProperties["source_boundary_closed"] = true
		ring := append([]geodata.GeoPoint(nil), curve...)
		if solarShadowRegionDegenerate(curve, ring) {
			// 首末接触附近区域收缩到零：整条缺省，不退化成线，也不让近零面积的环触发导出错误。
			return marshalEmptyFeatureCollection()
		}
		value, err := multiPolygonGeometry([][]geodata.GeoPoint{ring})
		if err != nil {
			return nil, fmt.Errorf("geojson: solar shadow instant at %s: %w", formatTime(instant.Time), err)
		}
		return marshalFeatureCollection([]feature{
			newFeature(solarEclipseEvent, regionRole, value, regionProperties),
		})
	}
	regionProperties["source_boundary_closed"] = false
	regionProperties["geometry_role"] = "horizon-closed-region"
	regionProperties["closure"] = solarHorizonClosureProperties(
		instant.Time, solarHorizonClosureExact(instant.Boundaries, instant.HorizonEnds),
	)
	ring, boundary := solarShadowFootprintHorizonRing(
		instant.Time, instant.Boundaries, instant.HorizonEnds, curve,
	)
	if solarShadowRegionDegenerate(curve, ring) {
		return marshalEmptyFeatureCollection()
	}
	value, err := multiPolygonGeometry([][]geodata.GeoPoint{ring})
	if err != nil {
		return nil, fmt.Errorf("geojson: solar shadow instant at %s: %w", formatTime(instant.Time), err)
	}
	features = append(features, newFeature(
		solarEclipseEvent, regionRole, value, regionProperties,
	))
	boundaryValue, err := geoMultiLineGeometry(boundary, false)
	if err != nil {
		return nil, fmt.Errorf("geojson: solar shadow instant at %s: %w", formatTime(instant.Time), err)
	}
	boundaryProperties := cloneProperties(properties)
	boundaryProperties["time"] = formatTime(instant.Time)
	boundaryProperties["source_boundary_closed"] = false
	boundaryProperties["geometry_role"] = "open-boundary"
	features = append(features, newFeature(
		solarEclipseEvent, boundaryRole, boundaryValue, boundaryProperties,
	))
	return marshalFeatureCollection(features)
}
