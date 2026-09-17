package geojson

import (
	"fmt"
	"math"
	"time"

	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/internal/occultationgeo"
	"github.com/starainrt/astro/moon"
)

const lunarOccultationEvent = "lunar-occultation"

type occultationBandKind uint8

const (
	occultationSweepBand occultationBandKind = iota
	stellarOccultationBand
	partialOccultationBand
	totalOccultationBand
)

// MarshalStarOccultationFootprint 将指定时刻的精确恒星月掩足迹编码为 GeoJSON FeatureCollection；事件外返回空集合。
// MarshalStarOccultationFootprint encodes one exact stellar occultation footprint as a GeoJSON FeatureCollection; instants outside the event produce an empty collection.
func MarshalStarOccultationFootprint(instant moon.StarOccultationInstant) ([]byte, error) {
	if instant.Time.IsZero() {
		return nil, fmt.Errorf("geojson: stellar occultation footprint time is required")
	}
	if instant.Footprint == nil {
		return marshalEmptyFeatureCollection()
	}
	if !instant.Footprint.Time.Equal(instant.Time) {
		return nil, fmt.Errorf("geojson: stellar occultation footprint time must match the requested instant")
	}
	properties := map[string]interface{}{
		"target_type":      "star",
		"target_id":        instant.TargetID,
		"delta_t_seconds":  instant.DeltaTSeconds,
		"interp_signature": occultationFootprintSignature(instant.Footprint, "occultation"),
	}
	addOccultationClosureProperties(properties, instant.Footprint, instant.Time, instant.SublunarLongitude, instant.SublunarLatitude)
	features, err := appendOccultationFootprints(
		nil, "occultation-footprint", []moon.PlanetOccultationFootprint{*instant.Footprint}, properties,
	)
	if err != nil {
		return nil, err
	}
	return marshalFeatureCollection(features)
}

// MarshalPlanetOccultationFootprints 将指定时刻可用的精确行星外切和内切足迹编码为 GeoJSON FeatureCollection；事件外返回空集合。
// MarshalPlanetOccultationFootprints encodes the available exact outer- and inner-contact planetary footprints as a GeoJSON FeatureCollection; instants outside the event produce an empty collection.
func MarshalPlanetOccultationFootprints(instant moon.PlanetOccultationInstant) ([]byte, error) {
	if instant.Time.IsZero() {
		return nil, fmt.Errorf("geojson: planetary occultation footprint time is required")
	}
	if err := instant.Planet.Validate(); err != nil {
		return nil, fmt.Errorf("geojson: planetary occultation target: %w", err)
	}
	if instant.Partial == nil && instant.Total == nil {
		return marshalEmptyFeatureCollection()
	}
	properties := map[string]interface{}{
		"target_type":     "planet",
		"target_id":       instant.TargetID,
		"planet":          string(instant.Planet),
		"delta_t_seconds": instant.DeltaTSeconds,
	}
	features := make([]feature, 0, 2)
	for _, current := range []struct {
		role      string
		footprint *moon.PlanetOccultationFootprint
	}{
		{role: "partial-footprint", footprint: instant.Partial},
		{role: "total-footprint", footprint: instant.Total},
	} {
		if current.footprint == nil {
			continue
		}
		if !current.footprint.Time.Equal(instant.Time) {
			return nil, fmt.Errorf("geojson: %s time must match the requested instant", current.role)
		}
		currentProperties := cloneProperties(properties)
		currentProperties["interp_signature"] = occultationFootprintSignature(current.footprint, current.role)
		addOccultationClosureProperties(
			currentProperties, current.footprint, instant.Time,
			instant.SublunarLongitude, instant.SublunarLatitude,
		)
		var err error
		features, err = appendOccultationFootprints(
			features, current.role, []moon.PlanetOccultationFootprint{*current.footprint}, currentProperties,
		)
		if err != nil {
			return nil, err
		}
	}
	return marshalFeatureCollection(features)
}

// MarshalStarOccultation 将月掩恒星的全球掩带和中心线编码为 GeoJSON。
// MarshalStarOccultation encodes a global stellar occultation band and center line as GeoJSON.
func MarshalStarOccultation(path moon.StarOccultationPath) ([]byte, error) {
	return marshalStarOccultation(path, nil)
}

// MarshalStarOccultationWithTimeMarkers 编码恒星月掩，并沿中心线按固定间隔追加 Point 要素。
// MarshalStarOccultationWithTimeMarkers encodes a stellar occultation and adds Point Features at regular intervals along its center line.
func MarshalStarOccultationWithTimeMarkers(
	path moon.StarOccultationPath,
	options TimeMarkerOptions,
) ([]byte, error) {
	return marshalStarOccultation(path, &options)
}

func marshalStarOccultation(path moon.StarOccultationPath, markerOptions *TimeMarkerOptions) ([]byte, error) {
	if markerOptions != nil {
		if err := validateTimeMarkerOptions(*markerOptions); err != nil {
			return nil, err
		}
	}
	if err := validateStarOccultationPathData(path); err != nil {
		return nil, err
	}
	properties := map[string]interface{}{
		"target_type":                  "star",
		"target_id":                    path.TargetID,
		"complete":                     path.Complete,
		"compact_band":                 len(path.BandFootprints) > 0,
		"step_seconds":                 path.Step.Seconds(),
		"target_spacing_km":            path.TargetSpacingKM,
		"greatest_limit_separation_km": path.GreatestLimitSeparationKM,
	}
	var value geometry
	var err error
	authoritative := false
	// 掩带边界优先由解析接触/相位网络定义，两种足迹模式共用同一构造器，瞬时足迹只作
	// 覆盖见证与时间轴细节；只有解析边界完全缺失时才沿用纯足迹扫掠。默认的密集瞬时足迹
	// 走纯扫掠会截断非中心事件的极向部分，使月升可见性边界落在掩带之外。
	// The band boundary prefers the analytic contact/phase network, so both footprint
	// modes share one constructor and the footprints only witness coverage and carry
	// the time axis. The pure sweep stays only for the case with no analytic boundary:
	// applied to dense footprints it truncates the poleward part of non-central events
	// and leaves the moonrise visibility boundary outside the band.
	bandFootprints := path.BandFootprints
	if len(bandFootprints) == 0 && (len(path.BandContours) > 0 || len(path.RiseSetCurves) > 0) {
		bandFootprints = path.Footprints
	}
	switch {
	case len(bandFootprints) > 0:
		value, authoritative, err = occultationCompactBandGeometry(
			bandFootprints, path.BandContours, path.VisibilityContours,
			path.NorthernLimit, path.SouthernLimit, path.RiseSetCurves,
			stellarOccultationBand,
		)
	case len(path.Footprints) > 0:
		value, err = occultationFootprintSweepGeometry(path.Footprints)
	default:
		value, err = occultationBandGeometry(path.NorthernLimit, path.SouthernLimit)
	}
	if err != nil {
		return nil, fmt.Errorf("geojson: stellar occultation band: %w", err)
	}
	bandProperties := cloneProperties(properties)
	if len(bandFootprints) > 0 {
		applyOccultationBandSourceProperties(bandProperties, authoritative, len(path.BandContours))
	}
	features := []feature{
		newFeature(lunarOccultationEvent, "occultation-band", value, bandProperties),
	}
	features, err = appendOccultationBandOutline(features, "band-outline", value, bandProperties)
	if err != nil {
		return nil, err
	}
	if len(path.Footprints) > 0 {
		features, err = appendTimedOccultationFootprintFeatures(
			features, "occultation-footprint", path.Footprints, properties,
		)
		if err != nil {
			return nil, err
		}
	}
	features, err = appendOccultationRiseSetCurveFeatures(features, path.RiseSetCurves, properties, "partial")
	if err != nil {
		return nil, err
	}
	// 连接线必须与掩带取自同一足迹集合，否则掩带走解析回退而来、连接线却按空输入生成。
	// The connectors must use the same footprint set as the band; otherwise a band
	// built from the analytic fallback gets connectors generated from empty input.
	features, err = appendOccultationHorizonConnectorFeatures(
		features, bandFootprints, path.NorthernLimit, path.SouthernLimit,
		path.RiseSetCurves, properties, "partial", stellarOccultationBand,
	)
	if err != nil {
		return nil, err
	}
	if len(path.CenterLine) > 0 {
		features, err = appendOccultationPathLine(features, "center-line", path.CenterLine, properties)
		if err != nil {
			return nil, err
		}
	}
	features, err = appendOccultationBoundaryLine(features, "north-limit", path.NorthernLimit, properties)
	if err != nil {
		return nil, err
	}
	features, err = appendOccultationBoundaryLine(features, "south-limit", path.SouthernLimit, properties)
	if err != nil {
		return nil, err
	}
	if markerOptions != nil && len(path.CenterLine) > 0 {
		features, err = appendTimeMarkerFeatures(
			features,
			lunarOccultationEvent,
			"center-line",
			occultationPathSamples(path.CenterLine),
			*markerOptions,
		)
		if err != nil {
			return nil, err
		}
	}
	for _, marker := range []struct {
		role  string
		point moon.OccultationPathPoint
	}{
		{role: "start", point: path.Start},
		{role: "greatest", point: path.Greatest},
		{role: "end", point: path.End},
	} {
		features, err = appendOccultationPoint(features, marker.role, marker.point, properties)
		if err != nil {
			return nil, err
		}
	}
	return marshalFeatureCollection(features)
}

// MarshalPlanetOccultation 将月掩行星的部分掩、全掩和中心线编码为 GeoJSON。
// MarshalPlanetOccultation encodes partial, total, and center-line planetary occultation geometry as GeoJSON.
func MarshalPlanetOccultation(path moon.PlanetOccultationPath) ([]byte, error) {
	return marshalPlanetOccultation(path, nil)
}

// MarshalPlanetOccultationWithTimeMarkers 编码行星月掩，并沿中心线按固定间隔追加 Point 要素。
// MarshalPlanetOccultationWithTimeMarkers encodes a planetary occultation and adds Point Features at regular intervals along its center line.
func MarshalPlanetOccultationWithTimeMarkers(
	path moon.PlanetOccultationPath,
	options TimeMarkerOptions,
) ([]byte, error) {
	return marshalPlanetOccultation(path, &options)
}

func marshalPlanetOccultation(path moon.PlanetOccultationPath, markerOptions *TimeMarkerOptions) ([]byte, error) {
	if markerOptions != nil {
		if err := validateTimeMarkerOptions(*markerOptions); err != nil {
			return nil, err
		}
	}
	if err := path.Planet.Validate(); err != nil {
		return nil, fmt.Errorf("geojson: planetary occultation target: %w", err)
	}
	if err := validatePlanetOccultationPathData(path); err != nil {
		return nil, err
	}
	properties := map[string]interface{}{
		"target_type":                  "planet",
		"target_id":                    path.TargetID,
		"planet":                       string(path.Planet),
		"complete":                     path.Complete,
		"has_total_band":               path.HasTotalBand,
		"total_complete":               path.TotalComplete,
		"step_seconds":                 path.Step.Seconds(),
		"target_spacing_km":            path.TargetSpacingKM,
		"greatest_total_width_km":      path.GreatestTotalWidthKM,
		"greatest_limit_separation_km": path.GreatestLimitSeparationKM,
		"compact_band":                 len(path.PartialBandFootprints) > 0 || len(path.TotalBandFootprints) > 0,
	}
	features := make([]feature, 0, len(path.PartialFootprints)+len(path.TotalFootprints)+14)
	var err error
	// 偏掩带与全掩带共用同一套边界来源：解析接触/相位网络存在时优先由它定义边界，
	// 瞬时足迹只作覆盖见证与时间轴细节；两套解析边界都缺失时才保留纯足迹扫掠兼容路径。
	// Both bands share one boundary source: the analytic contact/phase network defines the
	// boundary when present and the footprints only witness coverage and carry the time
	// axis. The pure footprint sweep stays as the compatibility path used only when no
	// analytic boundary exists.
	partialFootprints := path.PartialBandFootprints
	if len(partialFootprints) == 0 && (len(path.PartialBandContours) > 0 || len(path.RiseSetCurves) > 0) {
		// 默认（密集瞬时足迹）模式改用解析边界，避免纯扫掠截断非中心事件的极向部分。
		// Dense-footprint mode switches to the analytic boundary so that a pure sweep
		// cannot truncate the poleward part of a non-central event.
		partialFootprints = path.PartialFootprints
	}
	totalFootprints := path.TotalBandFootprints
	if len(totalFootprints) == 0 && (len(path.TotalBandContours) > 0 || len(path.TotalRiseSetCurves) > 0) {
		// 全掩带沿用与偏掩带相同的边界来源选择。
		// The total band follows the same boundary-source selection as the partial band.
		totalFootprints = path.TotalFootprints
	}
	switch {
	case len(partialFootprints) > 0:
		features, err = appendOccultationFootprintBand(
			features, "partial-band", partialFootprints, path.PartialBandContours,
			path.PartialVisibilityContours, path.NorthernLimit, path.SouthernLimit,
			path.RiseSetCurves, properties, partialOccultationBand,
		)
		if err == nil && len(path.PartialFootprints) > 0 {
			features, err = appendTimedOccultationFootprintFeatures(
				features, "partial-footprint", path.PartialFootprints, properties,
			)
		}
	case len(path.PartialFootprints) > 0:
		bandGeometry, bandErr := occultationFootprintSweepGeometry(path.PartialFootprints)
		if bandErr != nil {
			return nil, fmt.Errorf("geojson: partial-band: %w", bandErr)
		}
		bandProperties := cloneProperties(properties)
		bandProperties["static_band"] = true
		features = append(features, newFeature(lunarOccultationEvent, "partial-band", bandGeometry, bandProperties))
		features, err = appendOccultationBandOutline(features, "band-outline", bandGeometry, bandProperties)
		if err != nil {
			return nil, err
		}
		features, err = appendOccultationFootprints(
			features, "partial-footprint", path.PartialFootprints, properties,
		)
	default:
		features, err = appendOccultationBand(
			features, "partial-band", path.NorthernLimit, path.SouthernLimit, properties,
		)
	}
	if err != nil {
		return nil, err
	}

	if path.HasTotalBand {
		// 全掩带的来源已在函数级解析完毕，这里只按已选集合出图。
		// The total-band source is already resolved at function scope; this block
		// only emits the geometry for the chosen set.
		switch {
		case len(totalFootprints) > 0:
			features, err = appendOccultationFootprintBand(
				features, "total-band", totalFootprints, path.TotalBandContours,
				path.TotalVisibilityContours, path.NorthernTotalLimit, path.SouthernTotalLimit,
				path.TotalRiseSetCurves, properties, totalOccultationBand,
			)
			if err == nil && len(path.TotalFootprints) > 0 {
				features, err = appendTimedOccultationFootprintFeatures(
					features, "total-footprint", path.TotalFootprints, properties,
				)
			}
		case len(path.TotalFootprints) > 0:
			bandGeometry, bandErr := occultationFootprintSweepGeometry(path.TotalFootprints)
			if bandErr != nil {
				return nil, fmt.Errorf("geojson: total-band: %w", bandErr)
			}
			bandProperties := cloneProperties(properties)
			bandProperties["static_band"] = true
			features = append(features, newFeature(lunarOccultationEvent, "total-band", bandGeometry, bandProperties))
			features, err = appendOccultationBandOutline(features, "total-band-outline", bandGeometry, bandProperties)
			if err != nil {
				return nil, err
			}
			features, err = appendOccultationFootprints(
				features, "total-footprint", path.TotalFootprints, properties,
			)
		default:
			features, err = appendOccultationBand(
				features, "total-band", path.NorthernTotalLimit, path.SouthernTotalLimit, properties,
			)
		}
		if err != nil {
			return nil, err
		}
	}
	// Analytic contour bands already carry the complete contact/visibility/
	// phase boundary network. Do not run the legacy footprint-junction and
	// outline-alignment passes over them: those passes intentionally rewrite
	// polygon vertices and can reintroduce a boundary that is not in the
	// analytic network. Legacy paths without visibility contours retain the
	// containment repair for compatibility.
	analyticBands := len(path.PartialBandContours) > 0 && len(path.RiseSetCurves) > 0 &&
		(!path.HasTotalBand || (len(path.TotalBandContours) > 0 && len(path.TotalRiseSetCurves) > 0))
	if !analyticBands {
		features, err = constrainPlanetTotalBandWithinPartial(features)
		if err != nil {
			return nil, err
		}
		features, err = roundAuthoritativePlanetBandJunctions(features)
		if err != nil {
			return nil, err
		}
		features, err = alignAuthoritativePlanetBandOutlines(
			features, path.RiseSetCurves, path.TotalRiseSetCurves,
		)
		if err != nil {
			return nil, err
		}
	}

	// Keep static band fills/outlines below the physical rise/set curves in
	// feature order. OpenLayers uses the GeoJSON feature order within a vector
	// source; appending the phase curves last prevents the static outline from
	// covering the visible moonrise/morning phase boundary.
	features, err = appendOccultationRiseSetCurveFeatures(features, path.RiseSetCurves, properties, "partial")
	if err != nil {
		return nil, err
	}
	// 与偏掩带同源：解析回退时掩带用了瞬时足迹，连接线也必须用同一集合。
	// Same source as the partial band: when the analytic fallback supplies the band
	// from instantaneous footprints, the connectors must use that same set.
	features, err = appendOccultationHorizonConnectorFeatures(
		features, partialFootprints, path.NorthernLimit, path.SouthernLimit,
		path.RiseSetCurves, properties, "partial", partialOccultationBand,
	)
	if err != nil {
		return nil, err
	}
	if path.HasTotalBand {
		// 与全掩带同源，规则同偏掩带。
		// Same source as the total band, following the partial-band rule.
		features, err = appendOccultationHorizonConnectorFeatures(
			features, totalFootprints, path.NorthernTotalLimit, path.SouthernTotalLimit,
			path.TotalRiseSetCurves, properties, "total", totalOccultationBand,
		)
		if err != nil {
			return nil, err
		}
	}

	if len(path.CenterLine) > 0 {
		features, err = appendOccultationPathLine(features, "center-line", path.CenterLine, properties)
		if err != nil {
			return nil, err
		}
	}
	features, err = appendOccultationBoundaryLine(features, "north-limit", path.NorthernLimit, properties)
	if err != nil {
		return nil, err
	}
	features, err = appendOccultationBoundaryLine(features, "south-limit", path.SouthernLimit, properties)
	if err != nil {
		return nil, err
	}
	if path.HasTotalBand {
		features, err = appendOccultationBoundaryLine(
			features, "north-total-limit", path.NorthernTotalLimit, properties,
		)
		if err != nil {
			return nil, err
		}
		features, err = appendOccultationBoundaryLine(
			features, "south-total-limit", path.SouthernTotalLimit, properties,
		)
		if err != nil {
			return nil, err
		}
	}
	if markerOptions != nil && len(path.CenterLine) > 0 {
		features, err = appendTimeMarkerFeatures(
			features,
			lunarOccultationEvent,
			"center-line",
			occultationPathSamples(path.CenterLine),
			*markerOptions,
		)
		if err != nil {
			return nil, err
		}
	}

	markers := []struct {
		role  string
		point moon.OccultationPathPoint
	}{
		{role: "start", point: path.Start},
	}
	if path.HasTotalBand {
		markers = append(markers, struct {
			role  string
			point moon.OccultationPathPoint
		}{role: "total-start", point: path.TotalStart})
	}
	markers = append(markers, struct {
		role  string
		point moon.OccultationPathPoint
	}{role: "greatest", point: path.Greatest})
	if path.HasTotalBand {
		markers = append(markers, struct {
			role  string
			point moon.OccultationPathPoint
		}{role: "total-end", point: path.TotalEnd})
	}
	markers = append(markers, struct {
		role  string
		point moon.OccultationPathPoint
	}{role: "end", point: path.End})
	for _, marker := range markers {
		features, err = appendOccultationPoint(features, marker.role, marker.point, properties)
		if err != nil {
			return nil, err
		}
	}
	return marshalFeatureCollection(features)
}

// roundAuthoritativePlanetBandJunctions runs after the partial/total
// containment pass. That pass may union the two bands and recreate a short
// polar sweep seam which was already removed from each source band.
func roundAuthoritativePlanetBandJunctions(features []feature) ([]feature, error) {
	for index := range features {
		role := features[index].Properties["role"]
		if role != "partial-band" && role != "total-band" {
			continue
		}
		authoritative, _ := features[index].Properties["static_band_authoritative"].(bool)
		if !authoritative || features[index].Properties["source"] != "visible-footprint-sweep" {
			continue
		}
		polygons, ok := geometryMultiPolygonPoints(features[index].Geometry)
		if !ok {
			continue
		}
		if role == "total-band" {
			polygons = occultationgeo.RoundAuthoritativeTotalBandJunctions(polygons)
		} else {
			polygons = occultationgeo.RoundAuthoritativeBandJunctions(polygons)
		}
		value, err := multiPolygonGeometry(polygons)
		if err != nil {
			return nil, fmt.Errorf("geojson: round %s: %w", role, err)
		}
		features[index].Geometry = value
		outlineRole := "band-outline"
		if role == "total-band" {
			outlineRole = "total-band-outline"
		}
		outline, outlineOK, outlineErr := occultationBandOutlineGeometry(value)
		if outlineErr != nil {
			return nil, fmt.Errorf("geojson: round %s outline: %w", role, outlineErr)
		}
		if !outlineOK {
			continue
		}
		for outlineIndex := range features {
			if features[outlineIndex].Properties["role"] == outlineRole {
				features[outlineIndex].Geometry = outline
			}
		}
	}
	return features, nil
}

// alignAuthoritativePlanetBandOutlines replaces only the portions of an
// authoritative static outline that are also an exterior start/end phase
// envelope. The fill remains the full horizon-visible time union; this is a
// display-only operation that prevents a polygon union seam from showing as a
// spike where the purple phase boundary is already the physical outer edge.
func alignAuthoritativePlanetBandOutlines(
	features []feature,
	partialCurves, totalCurves []moon.OccultationRiseSetCurve,
) ([]feature, error) {
	for index := range features {
		role := features[index].Properties["role"]
		var curves []moon.OccultationRiseSetCurve
		switch role {
		case "band-outline":
			curves = partialCurves
		case "total-band-outline":
			curves = totalCurves
		default:
			continue
		}
		if authoritative, ok := features[index].Properties["static_band_authoritative"].(bool); !ok || !authoritative {
			continue
		}
		aligned, changed, err := occultationBandOutlinePhaseOverlap(
			features[index].Geometry, curves,
		)
		if err != nil {
			return nil, fmt.Errorf("geojson: align %s: %w", role, err)
		}
		if changed {
			features[index].Geometry = aligned
		}
	}
	return features, nil
}

const (
	// The phase curve is sampled at about 35 km for display. Keep the match
	// radius below one rendered edge so a nearby inner branch cannot be selected.
	occultationPhaseOverlapDistanceKM     = 35.0
	occultationPhaseOverlapJoinDistanceKM = 25.0
	occultationPhaseOverlapMinimumArcKM   = 150.0
	occultationPhaseOverlapMaximumEdgeKM  = 55.0
)

func occultationBandOutlinePhaseOverlap(
	value geometry,
	curves []moon.OccultationRiseSetCurve,
) (geometry, bool, error) {
	if len(curves) == 0 {
		return value, false, nil
	}
	base := value
	var err error
	if base.Type != "MultiLineString" {
		var baseOK bool
		base, baseOK, err = occultationBandOutlineGeometry(value)
		if err != nil || !baseOK {
			return value, false, err
		}
	}
	if err != nil {
		return value, false, err
	}
	coordinates, ok := base.Coordinates.([][][]float64)
	if !ok {
		return value, false, fmt.Errorf("outline coordinates have type %T", base.Coordinates)
	}
	phases := occultationPhaseEnvelopeLines(curves)
	if len(phases) == 0 {
		return value, false, nil
	}
	changed := false
	for index, source := range coordinates {
		ring := make([]geodata.GeoPoint, len(source))
		for pointIndex, point := range source {
			if len(point) < 2 {
				return value, false, fmt.Errorf("outline point %d/%d is malformed", index, pointIndex)
			}
			ring[pointIndex] = geodata.GeoPoint{Longitude: point[0], Latitude: point[1]}
		}
		aligned, ringChanged := alignOccultationOutlineRingPhaseOverlap(ring, phases)
		if !ringChanged {
			continue
		}
		changed = true
		coordinates[index] = make([][]float64, len(aligned))
		for pointIndex, point := range aligned {
			coordinates[index][pointIndex] = []float64{point.Longitude, point.Latitude}
		}
	}
	if !changed {
		return value, false, nil
	}
	return geometry{Type: "MultiLineString", Coordinates: coordinates}, true, nil
}

func occultationPhaseEnvelopeLines(
	curves []moon.OccultationRiseSetCurve,
) [][]geodata.GeoPoint {
	densified := occultationgeo.DensifyRiseSetCurves(curves, 35)
	lines := make([][]geodata.GeoPoint, 0)
	for _, curve := range densified {
		if curve.Phase != moon.RiseSetPhaseStart && curve.Phase != moon.RiseSetPhaseEnd {
			continue
		}
		for _, segment := range curve.Segments {
			if len(segment) < 2 {
				continue
			}
			line := make([]geodata.GeoPoint, len(segment))
			for index, point := range segment {
				line[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			lines = append(lines, line)
		}
	}
	return lines
}

type occultationPhaseOverlapRun struct {
	start, end int
	arcKM      float64
}

func alignOccultationOutlineRingPhaseOverlap(
	ring []geodata.GeoPoint,
	phases [][]geodata.GeoPoint,
) ([]geodata.GeoPoint, bool) {
	if len(ring) < 5 {
		return ring, false
	}
	closed := geodata.SameGeoPoint(ring[0], ring[len(ring)-1])
	open := append([]geodata.GeoPoint(nil), ring...)
	if closed {
		open = open[:len(open)-1]
	}
	if len(open) < 4 {
		return ring, false
	}
	// Process longer overlaps first. A short branch that shares an endpoint
	// with the outer branch must not consume the same ring section first.
	ordered := append([][]geodata.GeoPoint(nil), phases...)
	for left := 0; left < len(ordered); left++ {
		for right := left + 1; right < len(ordered); right++ {
			if occultationGeoLineLengthKM(ordered[right]) > occultationGeoLineLengthKM(ordered[left]) {
				ordered[left], ordered[right] = ordered[right], ordered[left]
			}
		}
	}
	changed := false
	for _, phase := range ordered {
		if len(phase) < 2 {
			continue
		}
		if updated, ok := replaceOccultationOutlinePhaseRun(open, phase); ok {
			open = updated
			changed = true
			// A second branch can share the same horizon endpoint while lying
			// on the inner side of the band. Only the longest verified overlap
			// is the exterior envelope for this ring.
			break
		}
	}
	if !changed {
		return ring, false
	}
	open = append(open, open[0])
	return open, true
}

func replaceOccultationOutlinePhaseRun(
	ring, phase []geodata.GeoPoint,
) ([]geodata.GeoPoint, bool) {
	runs := occultationPhaseOverlapRuns(phase, ring)
	if len(runs) == 0 {
		return ring, false
	}
	best := runs[0]
	for _, run := range runs[1:] {
		if run.arcKM > best.arcKM {
			best = run
		}
	}
	if best.arcKM < occultationPhaseOverlapMinimumArcKM {
		return ring, false
	}
	startPoint := phase[best.start]
	endPoint := phase[best.end]
	startIndex, startDistance := occultationNearestRingVertex(startPoint, ring)
	endIndex, endDistance := occultationNearestRingVertex(endPoint, ring)
	if startIndex < 0 || endIndex < 0 ||
		startDistance > occultationPhaseOverlapJoinDistanceKM ||
		endDistance > occultationPhaseOverlapJoinDistanceKM || startIndex == endIndex {
		return ring, false
	}
	phaseRun := append([]geodata.GeoPoint(nil), phase[best.start:best.end+1]...)
	forwardArc := occultationRingPathLengthKM(ring, startIndex, endIndex, 1)
	backwardArc := occultationRingPathLengthKM(ring, startIndex, endIndex, -1)
	phaseArc := occultationGeoLineLengthKM(phaseRun)
	forward := math.Abs(forwardArc-phaseArc) <= math.Abs(backwardArc-phaseArc)
	if !forward {
		reverseOccultationGeoPoints(phaseRun)
		startIndex, endIndex = endIndex, startIndex
	}
	if occultationGeoLineLengthKM(phaseRun) < occultationPhaseOverlapMinimumArcKM {
		return ring, false
	}
	result := make([]geodata.GeoPoint, 0, len(ring)+len(phaseRun))
	result = append(result, phaseRun...)
	stepDirection := 1
	if !forward {
		stepDirection = -1
	}
	index := (endIndex + stepDirection + len(ring)) % len(ring)
	for index != startIndex {
		result = append(result, ring[index])
		if forward {
			index = (index + 1) % len(ring)
		} else {
			index = (index - 1 + len(ring)) % len(ring)
		}
	}
	if len(result) < 4 || !occultationRingEdgesWithinKM(result, occultationPhaseOverlapMaximumEdgeKM) {
		return ring, false
	}
	result = append(result, result[0])
	return result, true
}

func occultationPhaseOverlapRuns(
	phase, ring []geodata.GeoPoint,
) []occultationPhaseOverlapRun {
	const maximumGapPoints = 2
	runs := make([]occultationPhaseOverlapRun, 0, 2)
	start, gap := -1, 0
	for index, point := range phase {
		_, distance := occultationNearestRingVertex(point, ring)
		if distance <= occultationPhaseOverlapDistanceKM {
			if start < 0 {
				start = index
			}
			gap = 0
			continue
		}
		if start < 0 {
			continue
		}
		gap++
		if gap <= maximumGapPoints {
			continue
		}
		end := index - gap
		if end > start {
			arc := occultationGeoLineLengthKM(phase[start : end+1])
			if arc >= occultationPhaseOverlapMinimumArcKM {
				runs = append(runs, occultationPhaseOverlapRun{start: start, end: end, arcKM: arc})
			}
		}
		start, gap = -1, 0
	}
	if start >= 0 {
		end := len(phase) - 1
		if end > start {
			arc := occultationGeoLineLengthKM(phase[start : end+1])
			if arc >= occultationPhaseOverlapMinimumArcKM {
				runs = append(runs, occultationPhaseOverlapRun{start: start, end: end, arcKM: arc})
			}
		}
	}
	return runs
}

func occultationNearestRingVertex(
	point geodata.GeoPoint,
	ring []geodata.GeoPoint,
) (int, float64) {
	index := -1
	distance := math.Inf(1)
	for candidate, value := range ring {
		current := occultationGeoDistanceKM(point, value)
		if current < distance {
			index, distance = candidate, current
		}
	}
	return index, distance
}

func occultationRingPathLengthKM(ring []geodata.GeoPoint, start, end, direction int) float64 {
	if len(ring) == 0 || start < 0 || end < 0 || start >= len(ring) || end >= len(ring) {
		return math.Inf(1)
	}
	length := 0.0
	index := start
	for index != end {
		next := (index + direction + len(ring)) % len(ring)
		length += occultationGeoDistanceKM(ring[index], ring[next])
		index = next
		if length > 1e8 {
			return math.Inf(1)
		}
	}
	return length
}

func occultationRingEdgesWithinKM(ring []geodata.GeoPoint, maximum float64) bool {
	for index := 1; index < len(ring); index++ {
		if occultationGeoDistanceKM(ring[index-1], ring[index]) > maximum {
			return false
		}
	}
	return true
}

func occultationGeoLineLengthKM(points []geodata.GeoPoint) float64 {
	length := 0.0
	for index := 1; index < len(points); index++ {
		length += occultationGeoDistanceKM(points[index-1], points[index])
	}
	return length
}

func occultationGeoDistanceKM(first, second geodata.GeoPoint) float64 {
	const radiusKM = 6378.1366
	firstLatitude := first.Latitude * math.Pi / 180
	secondLatitude := second.Latitude * math.Pi / 180
	deltaLatitude := secondLatitude - firstLatitude
	deltaLongitude := math.Remainder(second.Longitude-first.Longitude, 360) * math.Pi / 180
	sineLatitude := math.Sin(deltaLatitude / 2)
	sineLongitude := math.Sin(deltaLongitude / 2)
	a := sineLatitude*sineLatitude + math.Cos(firstLatitude)*math.Cos(secondLatitude)*sineLongitude*sineLongitude
	return 2 * radiusKM * math.Atan2(math.Sqrt(math.Max(0, a)), math.Sqrt(math.Max(0, 1-a)))
}

func reverseOccultationGeoPoints(points []geodata.GeoPoint) {
	for left, right := 0, len(points)-1; left < right; left, right = left+1, right-1 {
		points[left], points[right] = points[right], points[left]
	}
}

func constrainPlanetTotalBandWithinPartial(features []feature) ([]feature, error) {
	partialIndex, totalIndex := -1, -1
	for index, current := range features {
		switch current.Properties["role"] {
		case "partial-band":
			partialIndex = index
		case "total-band":
			totalIndex = index
		}
	}
	if partialIndex < 0 || totalIndex < 0 {
		return features, nil
	}
	parent, parentOK := geometryMultiPolygonPoints(features[partialIndex].Geometry)
	child, childOK := geometryMultiPolygonPoints(features[totalIndex].Geometry)
	if !parentOK || !childOK {
		return features, nil
	}
	initialMiss := geodata.SphericalPolygonsPathMissDistanceKM(parent, child, true)
	partialSweepSource := features[partialIndex].Properties["source"] == "visible-footprint-sweep"
	if initialMiss > 0 && initialMiss <= 25 && !partialSweepSource {
		// A small total-vs-partial discrepancy is a sampling residual at the polar
		// junction. Do not snap total vertices onto the partial ring: that creates
		// a visible staircase made from the parent's unrelated samples. Expand the
		// parent once by the already smooth child face and keep the child boundary
		// intact. Larger discrepancies are core-geometry errors and are left for the
		// caller to diagnose rather than silently widening the serialized band.
		input := append(append([][]geodata.GeoPoint(nil), parent...), child...)
		expanded, unionErr := geodata.UnionPolygons(input)
		if unionErr == nil {
			expanded = occultationgeo.CleanAuthoritativeBandPolygons(expanded)
		}
		expandedMiss := geodata.SphericalPolygonsPathMissDistanceKM(expanded, child, true)
		if unionErr == nil && len(expanded) > 0 && expandedMiss <= 10 &&
			len(expanded) == 1 {
			parentValue, parentErr := multiPolygonGeometry(expanded)
			if parentErr != nil {
				return nil, fmt.Errorf("geojson: expand partial-band: %w", parentErr)
			}
			features[partialIndex].Geometry = parentValue
			for index := range features {
				if features[index].Properties["role"] != "band-outline" {
					continue
				}
				outline, ok, outlineErr := occultationBandOutlineGeometry(parentValue)
				if outlineErr != nil {
					return nil, fmt.Errorf("geojson: expand band-outline: %w", outlineErr)
				}
				if ok {
					features[index].Geometry = outline
				}
			}
			return features, nil
		}
	}
	repaired := occultationgeo.ConstrainPolygonsWithin(parent, child)
	if len(repaired) == 0 || geodata.SphericalPolygonsPathMissDistanceKM(parent, repaired, true) > 10 {
		return features, nil
	}
	value, err := multiPolygonGeometry(repaired)
	if err != nil {
		return nil, fmt.Errorf("geojson: constrain total-band: %w", err)
	}
	features[totalIndex].Geometry = value
	for index := range features {
		if features[index].Properties["role"] != "total-band-outline" {
			continue
		}
		outline, ok, outlineErr := occultationBandOutlineGeometry(value)
		if outlineErr != nil {
			return nil, fmt.Errorf("geojson: constrain total-band-outline: %w", outlineErr)
		}
		if ok {
			features[index].Geometry = outline
		}
	}
	return features, nil
}

func geometryMultiPolygonPoints(value geometry) ([][]geodata.GeoPoint, bool) {
	if value.Type != "MultiPolygon" {
		return nil, false
	}
	coordinates, ok := value.Coordinates.([][][][]float64)
	if !ok {
		return nil, false
	}
	polygons := make([][]geodata.GeoPoint, 0, len(coordinates))
	for _, polygon := range coordinates {
		if len(polygon) == 0 || len(polygon[0]) < 4 {
			continue
		}
		ring := make([]geodata.GeoPoint, len(polygon[0]))
		for index, point := range polygon[0] {
			if len(point) < 2 {
				return nil, false
			}
			ring[index] = geodata.GeoPoint{Longitude: point[0], Latitude: point[1]}
		}
		polygons = append(polygons, ring)
	}
	return polygons, len(polygons) > 0
}

func appendOccultationBand(
	features []feature,
	role string,
	northern, southern []moon.OccultationPathPoint,
	properties map[string]interface{},
) ([]feature, error) {
	value, err := occultationBandGeometry(northern, southern)
	if err != nil {
		return nil, fmt.Errorf("geojson: %s: %w", role, err)
	}
	features = append(features, newFeature(
		lunarOccultationEvent, role, value, cloneProperties(properties),
	))
	return appendOccultationBandOutline(features, occultationBandOutlineRole(role), value, properties)
}

func appendOccultationFootprints(
	features []feature,
	role string,
	footprints []moon.PlanetOccultationFootprint,
	properties map[string]interface{},
) ([]feature, error) {
	appended := 0
	for _, footprint := range footprints {
		if footprint.Time.IsZero() {
			return nil, fmt.Errorf("geojson: %s time is required", role)
		}
		polygons := make([][]geodata.GeoPoint, 0, len(footprint.Polygons))
		for _, source := range footprint.Polygons {
			polygon := make([]geodata.GeoPoint, len(source))
			for index, point := range source {
				polygon[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			polygons = append(polygons, polygon)
		}
		value, err := multiPolygonGeometry(polygons)
		if err != nil {
			return nil, fmt.Errorf("geojson: %s at %s: %w", role, formatTime(footprint.Time), err)
		}
		footprintProperties := cloneProperties(properties)
		footprintProperties["time"] = formatTime(footprint.Time)
		features = append(features, newFeature(
			lunarOccultationEvent, role, value, footprintProperties,
		))
		appended++
	}
	if appended == 0 {
		return nil, fmt.Errorf("geojson: %s has no valid polygons", role)
	}
	return features, nil
}

// appendTimedOccultationFootprintFeatures emits instantaneous footprint
// polygons with their sample time. Static band geometry remains separate.
func appendTimedOccultationFootprintFeatures(
	features []feature,
	role string,
	footprints []moon.PlanetOccultationFootprint,
	properties map[string]interface{},
) ([]feature, error) {
	return appendOccultationFootprints(features, role, footprints, properties)
}

// applyOccultationBandSourceProperties 写掩带来源标签：解析边界可用时足迹仍是覆盖见证。
func applyOccultationBandSourceProperties(
	bandProperties map[string]interface{},
	authoritative bool,
	contourCount int,
) {
	bandProperties["static_band_authoritative"] = authoritative
	if authoritative {
		bandProperties["source"] = "visible-footprint-sweep"
		bandProperties["boundary_source"] = "footprint-sweep+horizon-visible"
		return
	}
	bandProperties["source"] = "footprint-sweep-fallback"
	if contourCount > 0 {
		bandProperties["boundary_source"] = "contact-contours+horizon-boundary"
	}
}

func appendOccultationFootprintBand(
	features []feature,
	role string,
	footprints []moon.PlanetOccultationFootprint,
	contours [][]moon.OccultationPathPoint,
	visibilityContours [][]moon.OccultationPathPoint,
	northern, southern []moon.OccultationPathPoint,
	curves []moon.OccultationRiseSetCurve,
	properties map[string]interface{},
	kind occultationBandKind,
) ([]feature, error) {
	value, authoritative, err := occultationCompactBandGeometry(
		footprints, contours, visibilityContours, northern, southern, curves,
		kind,
	)
	if err != nil {
		return nil, fmt.Errorf("geojson: %s: %w", role, err)
	}
	bandProperties := cloneProperties(properties)
	applyOccultationBandSourceProperties(bandProperties, authoritative, len(contours))
	features = append(features, newFeature(
		lunarOccultationEvent, role, value, bandProperties,
	))
	return appendOccultationBandOutline(features, occultationBandOutlineRole(role), value, bandProperties)
}

func occultationBandOutlineRole(role string) string {
	if role == "total-band" {
		return "total-band-outline"
	}
	return "band-outline"
}

// appendOccultationBandOutline emits a closed line representation of a static
// band polygon. It deliberately does not turn discontinuous line fragments
// in a GeometryCollection into a fake closure.
func appendOccultationBandOutline(
	features []feature,
	role string,
	band geometry,
	properties map[string]interface{},
) ([]feature, error) {
	outline, ok, err := occultationBandOutlineGeometry(band)
	if err != nil {
		return nil, fmt.Errorf("geojson: %s: %w", role, err)
	}
	if !ok {
		return features, nil
	}
	outlineProperties := cloneProperties(properties)
	sourceRole := "occultation-band"
	if role == "total-band-outline" {
		sourceRole = "total-band"
	} else if properties["target_type"] == "planet" {
		sourceRole = "partial-band"
	}
	outlineProperties["source_role"] = sourceRole
	outlineProperties["closed"] = true
	return append(features, newFeature(lunarOccultationEvent, role, outline, outlineProperties)), nil
}

func occultationBandOutlineGeometry(value geometry) (geometry, bool, error) {
	lines := make([][][]float64, 0)
	var collect func(geometry) error
	collect = func(current geometry) error {
		switch current.Type {
		case "Polygon":
			rings, ok := current.Coordinates.([][][]float64)
			if !ok {
				return fmt.Errorf("polygon coordinates have type %T", current.Coordinates)
			}
			for _, ring := range rings {
				if len(ring) >= 4 {
					lines = append(lines, cloneGeoJSONLine(ring))
				}
			}
		case "MultiPolygon":
			polygons, ok := current.Coordinates.([][][][]float64)
			if !ok {
				return fmt.Errorf("multi-polygon coordinates have type %T", current.Coordinates)
			}
			for _, polygon := range polygons {
				for _, ring := range polygon {
					if len(ring) >= 4 {
						lines = append(lines, cloneGeoJSONLine(ring))
					}
				}
			}
		case "GeometryCollection":
			for _, child := range current.Geometries {
				if err := collect(child); err != nil {
					return err
				}
			}
		case "", "MultiLineString", "LineString":
			// A line-only fragment is intentionally not closed here.
		default:
			return fmt.Errorf("unsupported band geometry type %q", current.Type)
		}
		return nil
	}
	if err := collect(value); err != nil {
		return geometry{}, false, err
	}
	if len(lines) == 0 {
		return geometry{}, false, nil
	}
	return geometry{Type: "MultiLineString", Coordinates: lines}, true, nil
}

func cloneGeoJSONLine(source [][]float64) [][]float64 {
	result := make([][]float64, len(source))
	for index, point := range source {
		result[index] = append([]float64(nil), point...)
	}
	return result
}

func occultationFootprintSweepGeometry(footprints []moon.OccultationFootprint) (geometry, error) {
	value, _, err := occultationCompactBandGeometry(
		footprints, nil, nil, nil, nil, nil, occultationSweepBand,
	)
	return value, err
}

func occultationCompactBandGeometry(
	footprints []moon.OccultationFootprint,
	contours [][]moon.OccultationPathPoint,
	visibilityContours [][]moon.OccultationPathPoint,
	northern, southern []moon.OccultationPathPoint,
	curves []moon.OccultationRiseSetCurve,
	kind occultationBandKind,
) (geometry, bool, error) {
	var (
		merged        [][]geodata.GeoPoint
		authoritative bool
		err           error
	)
	analytic := len(contours) > 0 && len(curves) > 0
	switch kind {
	case stellarOccultationBand:
		if analytic {
			merged, authoritative, err = occultationgeo.VisibleStarBandPolygonsFromAnalyticContours(
				footprints, contours, visibilityContours, northern, southern, curves,
			)
		} else if len(contours) > 0 {
			merged, authoritative, err = occultationgeo.VisibleBandPolygonsFromContours(
				footprints, contours, northern, southern, curves,
			)
		} else {
			merged, authoritative, err = occultationgeo.VisibleBandPolygons(footprints, northern, southern, curves)
		}
	case totalOccultationBand:
		if analytic {
			merged, authoritative, err = occultationgeo.VisibleTotalBandPolygonsFromAnalyticContours(
				footprints, contours, visibilityContours, northern, southern, curves,
			)
		} else if len(contours) > 0 {
			merged, authoritative, err = occultationgeo.VisibleTotalBandPolygonsFromContours(
				footprints, contours, northern, southern, curves,
			)
		} else {
			merged, authoritative, err = occultationgeo.VisibleTotalBandPolygons(footprints, northern, southern, curves)
		}
	case partialOccultationBand, occultationSweepBand:
		if analytic {
			merged, authoritative, err = occultationgeo.VisibleBandPolygonsFromAnalyticContours(
				footprints, contours, visibilityContours, northern, southern, curves,
			)
		} else if len(contours) > 0 {
			merged, authoritative, err = occultationgeo.VisibleBandPolygonsFromContours(
				footprints, contours, northern, southern, curves,
			)
		} else {
			merged, authoritative, err = occultationgeo.VisibleBandPolygons(footprints, northern, southern, curves)
		}
	}
	if err != nil {
		return geometry{}, false, err
	}
	value, err := multiPolygonGeometry(merged)
	return value, authoritative, err
}

func occultationBandGeometry(
	northern, southern []moon.OccultationPathPoint,
) (geometry, error) {
	if len(northern) != len(southern) {
		return geometry{}, fmt.Errorf("paired limits must have the same sample count")
	}
	count := len(northern)
	if count < 2 {
		return geometry{}, fmt.Errorf("paired limits require at least two points per side")
	}
	for index := range northern {
		if northern[index].Time.IsZero() || southern[index].Time.IsZero() {
			return geometry{}, fmt.Errorf("paired limit sample %d time is required", index)
		}
		if !northern[index].Time.Equal(southern[index].Time) {
			return geometry{}, fmt.Errorf("paired limit sample %d times must match", index)
		}
	}
	polygons := make([][]geodata.GeoPoint, 0, count-1)
	sections := make([][]geodata.GeoPoint, 0, 2)
	for _, sampleRange := range occultationgeo.ContinuousPairedBoundaryRanges(northern, southern) {
		if sampleRange.End-sampleRange.Start == 1 {
			north, south := northern[sampleRange.Start], southern[sampleRange.Start]
			sections = append(sections, []geodata.GeoPoint{
				{Longitude: north.Longitude, Latitude: north.Latitude},
				{Longitude: south.Longitude, Latitude: south.Latitude},
			})
			continue
		}
		for index := sampleRange.Start + 1; index < sampleRange.End; index++ {
			previousNorth, north := northern[index-1], northern[index]
			previousSouth, south := southern[index-1], southern[index]
			polygons = append(polygons, []geodata.GeoPoint{
				{Longitude: previousNorth.Longitude, Latitude: previousNorth.Latitude},
				{Longitude: north.Longitude, Latitude: north.Latitude},
				{Longitude: south.Longitude, Latitude: south.Latitude},
				{Longitude: previousSouth.Longitude, Latitude: previousSouth.Latitude},
			})
		}
	}
	geometries := make([]geometry, 0, 2)
	if len(polygons) > 0 {
		merged, err := geodata.UnionPolygons(polygons)
		if err != nil {
			return geometry{}, fmt.Errorf("merge paired limit strips: %w", err)
		}
		value, err := multiPolygonGeometry(merged)
		if err != nil {
			return geometry{}, err
		}
		geometries = append(geometries, value)
	}
	if len(sections) > 0 {
		value, err := occultationBandSectionsGeometry(sections)
		if err != nil {
			return geometry{}, err
		}
		geometries = append(geometries, value)
	}
	if len(geometries) == 0 {
		return geometry{}, fmt.Errorf("paired limits have no continuous polygon segments")
	}
	if len(geometries) == 1 {
		return geometries[0], nil
	}
	return geometry{Type: "GeometryCollection", Geometries: geometries}, nil
}

func occultationBandSectionsGeometry(sections [][]geodata.GeoPoint) (geometry, error) {
	coordinates := make([][][]float64, 0, len(sections))
	for _, section := range sections {
		value, err := geoMultiLineGeometry(section, false)
		if err != nil {
			return geometry{}, err
		}
		lines, ok := value.Coordinates.([][][]float64)
		if !ok {
			return geometry{}, fmt.Errorf("unexpected band-section geometry %T", value.Coordinates)
		}
		coordinates = append(coordinates, lines...)
	}
	return geometry{Type: "MultiLineString", Coordinates: coordinates}, nil
}

func appendOccultationPathLine(
	features []feature,
	role string,
	points []moon.OccultationPathPoint,
	properties map[string]interface{},
) ([]feature, error) {
	samples := make([]pathSample, len(points))
	for index, point := range points {
		samples[index] = occultationPathSample(point)
	}
	return appendTimedLineFeature(features, lunarOccultationEvent, role, samples, properties)
}

func appendOccultationBoundaryLine(
	features []feature,
	role string,
	points []moon.OccultationPathPoint,
	properties map[string]interface{},
) ([]feature, error) {
	ranges := occultationgeo.ContinuousBoundaryRanges(points)
	segments := make([][]pathSample, len(ranges))
	for segmentIndex, sampleRange := range ranges {
		segments[segmentIndex] = occultationPathSamples(points[sampleRange.Start:sampleRange.End])
	}
	value, times, err := timedMultiLineGeometryFromSegments(segments)
	if err != nil {
		return nil, fmt.Errorf("geojson: %s: %w", role, err)
	}
	lineProperties := cloneProperties(properties)
	lineProperties["times"] = times
	return append(features, newFeature(lunarOccultationEvent, role, value, lineProperties)), nil
}

func appendOccultationRiseSetCurveFeatures(
	features []feature,
	curves []moon.OccultationRiseSetCurve,
	properties map[string]interface{},
	band string,
) ([]feature, error) {
	curves = occultationgeo.DensifyRiseSetCurves(curves, 35)
	for _, curve := range curves {
		segments := make([][]pathSample, len(curve.Segments))
		for index, segment := range curve.Segments {
			segments[index] = occultationPathSamples(segment)
		}
		value, times, err := timedMultiLineGeometryFromSegmentsWithTimeOrder(segments, true)
		if err != nil {
			return nil, fmt.Errorf("geojson: visibility-boundary: %w", err)
		}
		curveProperties := cloneProperties(properties)
		curveProperties["phase"] = string(curve.Phase)
		curveProperties["horizon"] = string(curve.Direction)
		curveProperties["body"] = "moon"
		if band != "" {
			curveProperties["band"] = band
		}
		curveProperties["times"] = times
		features = append(features, newFeature(
			lunarOccultationEvent, "visibility-boundary", value, curveProperties,
		))
	}
	return features, nil
}

func appendOccultationHorizonConnectorFeatures(
	features []feature,
	footprints []moon.OccultationFootprint,
	northern, southern []moon.OccultationPathPoint,
	curves []moon.OccultationRiseSetCurve,
	properties map[string]interface{},
	band string,
	kind occultationBandKind,
) ([]feature, error) {
	connectors := occultationgeo.HorizonConnectorSegments(footprints, curves, northern, southern)
	if kind == stellarOccultationBand {
		connectors = occultationgeo.StarHorizonConnectorSegments(footprints, curves, northern, southern)
	}
	if len(connectors) == 0 {
		return features, nil
	}
	segmentsByDirection := map[moon.RiseSetDirection][][]pathSample{
		moon.RiseSetDirectionRise: nil,
		moon.RiseSetDirectionSet:  nil,
	}
	for _, connector := range connectors {
		if len(connector.Points) < 2 {
			continue
		}
		connector.Points = occultationgeo.DensifyOccultationPathPoints(connector.Points, 35)
		segmentsByDirection[connector.Direction] = append(
			segmentsByDirection[connector.Direction],
			occultationPathSamples(connector.Points),
		)
	}
	for _, direction := range []moon.RiseSetDirection{moon.RiseSetDirectionRise, moon.RiseSetDirectionSet} {
		segments := segmentsByDirection[direction]
		if len(segments) == 0 {
			continue
		}
		value, times, err := timedMultiLineGeometryFromSegmentsWithTimeOrder(segments, false)
		if err != nil {
			return nil, fmt.Errorf("geojson: horizon-connector: %w", err)
		}
		connectorProperties := cloneProperties(properties)
		connectorProperties["phase"] = "horizon"
		connectorProperties["horizon"] = string(direction)
		connectorProperties["body"] = "moon"
		if band != "" {
			connectorProperties["band"] = band
		}
		connectorProperties["source"] = "footprint-horizon-connector"
		connectorProperties["times"] = times
		features = append(features, newFeature(
			lunarOccultationEvent, "horizon-connector", value, connectorProperties,
		))
	}
	return features, nil
}

func occultationPathSamples(points []moon.OccultationPathPoint) []pathSample {
	samples := make([]pathSample, len(points))
	for index, point := range points {
		samples[index] = occultationPathSample(point)
	}
	return samples
}

func appendOccultationPoint(
	features []feature,
	role string,
	point moon.OccultationPathPoint,
	properties map[string]interface{},
) ([]feature, error) {
	pointProperties := cloneProperties(properties)
	pointProperties["moon_altitude_deg"] = point.MoonAltitude
	pointProperties["width_km"] = point.WidthKM
	return appendPointFeature(
		features, lunarOccultationEvent, role, occultationPathSample(point), pointProperties,
	)
}

func occultationPathSample(point moon.OccultationPathPoint) pathSample {
	return pathSample{Time: point.Time, Longitude: point.Longitude, Latitude: point.Latitude}
}

func validateStarOccultationPathData(path moon.StarOccultationPath) error {
	if !path.Complete {
		return fmt.Errorf("geojson: occultation path is incomplete")
	}
	if err := (moon.OccultationPathOptions{Step: path.Step, TargetSpacingKM: path.TargetSpacingKM}).Validate(); err != nil {
		return fmt.Errorf("geojson: invalid occultation path sampling metadata: %w", err)
	}
	if err := validateOccultationPathPoint("start", path.Start); err != nil {
		return err
	}
	if err := validateOccultationPathPoint("greatest", path.Greatest); err != nil {
		return err
	}
	if err := validateOccultationPathPoint("end", path.End); err != nil {
		return err
	}
	if path.Greatest.Time.Before(path.Start.Time) || path.End.Time.Before(path.Greatest.Time) {
		return fmt.Errorf("geojson: occultation times must be ordered start, greatest, end")
	}
	if err := validateOccultationPathSeries("center line", path.CenterLine, false); err != nil {
		return err
	}
	if err := validateOccultationPathSeries("northern limit", path.NorthernLimit, true); err != nil {
		return err
	}
	if err := validateOccultationPathSeries("southern limit", path.SouthernLimit, true); err != nil {
		return err
	}
	if err := validateOccultationContours("band contours", path.BandContours, path.Start.Time, path.End.Time); err != nil {
		return err
	}
	if err := validateOccultationContours("visibility contours", path.VisibilityContours, path.Start.Time, path.End.Time); err != nil {
		return err
	}
	if len(path.NorthernLimit) != len(path.SouthernLimit) {
		return fmt.Errorf("geojson: occultation northern and southern limits must have the same sample count")
	}
	for index := range path.NorthernLimit {
		if !path.NorthernLimit[index].Time.Equal(path.SouthernLimit[index].Time) {
			return fmt.Errorf("geojson: occultation limit sample %d times must match", index)
		}
	}
	last := len(path.NorthernLimit) - 1
	if !path.NorthernLimit[0].Time.Equal(path.Start.Time) || !path.SouthernLimit[0].Time.Equal(path.Start.Time) ||
		!path.NorthernLimit[last].Time.Equal(path.End.Time) || !path.SouthernLimit[last].Time.Equal(path.End.Time) {
		return fmt.Errorf("geojson: occultation limits must span start through end")
	}
	if !finiteGeoJSON(path.GreatestLimitSeparationKM) || path.GreatestLimitSeparationKM < 0 {
		return fmt.Errorf("geojson: occultation greatest limit separation must be finite and non-negative")
	}
	if len(path.CenterLine) > 0 {
		if path.CenterLine[0].Time.Before(path.Start.Time) ||
			path.CenterLine[len(path.CenterLine)-1].Time.After(path.End.Time) {
			return fmt.Errorf("geojson: occultation center line must be inside start and end")
		}
		if path.Greatest.Time.Before(path.CenterLine[0].Time) ||
			path.Greatest.Time.After(path.CenterLine[len(path.CenterLine)-1].Time) {
			return fmt.Errorf("geojson: occultation greatest time is outside the center-line interval")
		}
	}
	if err := occultationgeo.ValidateRiseSetCurves(path.RiseSetCurves, path.Start.Time, path.End.Time); err != nil {
		return fmt.Errorf("geojson: invalid occultation rise/set curves: %w", err)
	}
	if err := validateOccultationFootprints("stellar", path.Footprints, path.Start.Time, path.End.Time); err != nil {
		return err
	}
	return validateOccultationFootprints("stellar compact band", path.BandFootprints, path.Start.Time, path.End.Time)
}

func validatePlanetOccultationPathData(path moon.PlanetOccultationPath) error {
	starPath := moon.StarOccultationPath{
		TargetID: path.TargetID, Start: path.Start, Greatest: path.Greatest, End: path.End,
		Complete: path.Complete, CenterLine: path.CenterLine,
		NorthernLimit: path.NorthernLimit, SouthernLimit: path.SouthernLimit,
		GreatestLimitSeparationKM: path.GreatestLimitSeparationKM,
		BandContours:              path.PartialBandContours, VisibilityContours: path.PartialVisibilityContours,
		RiseSetCurves: path.RiseSetCurves,
		Step:          path.Step, TargetSpacingKM: path.TargetSpacingKM,
	}
	if err := validateStarOccultationPathData(starPath); err != nil {
		return err
	}
	if !path.HasTotalBand {
		if path.TotalComplete || !path.TotalStart.Time.IsZero() || !path.TotalEnd.Time.IsZero() ||
			len(path.NorthernTotalLimit) != 0 || len(path.SouthernTotalLimit) != 0 ||
			len(path.TotalFootprints) != 0 || len(path.TotalBandFootprints) != 0 ||
			len(path.TotalBandContours) != 0 || len(path.TotalVisibilityContours) != 0 ||
			len(path.TotalRiseSetCurves) != 0 ||
			path.GreatestTotalWidthKM != 0 {
			return fmt.Errorf("geojson: total-band fields require HasTotalBand")
		}
		if err := validateOccultationFootprints("partial", path.PartialFootprints, path.Start.Time, path.End.Time); err != nil {
			return err
		}
		return validateOccultationFootprints("partial compact band", path.PartialBandFootprints, path.Start.Time, path.End.Time)
	}
	if !path.TotalComplete {
		return fmt.Errorf("geojson: total-occultation band is incomplete")
	}
	if err := validateOccultationPathPoint("total start", path.TotalStart); err != nil {
		return err
	}
	if err := validateOccultationPathPoint("total end", path.TotalEnd); err != nil {
		return err
	}
	if !path.Start.Time.Before(path.TotalStart.Time) || !path.TotalStart.Time.Before(path.Greatest.Time) ||
		!path.Greatest.Time.Before(path.TotalEnd.Time) || !path.TotalEnd.Time.Before(path.End.Time) {
		return fmt.Errorf("geojson: total-band times must be inside outer start, greatest, and end")
	}
	if !finiteGeoJSON(path.GreatestTotalWidthKM) || path.GreatestTotalWidthKM <= 0 ||
		!finiteGeoJSON(path.Greatest.WidthKM) || path.GreatestTotalWidthKM >= path.Greatest.WidthKM {
		return fmt.Errorf("geojson: total-band width must be positive and narrower than the outer band")
	}
	if err := validateOccultationPathSeries("northern total limit", path.NorthernTotalLimit, true); err != nil {
		return err
	}
	if err := validateOccultationPathSeries("southern total limit", path.SouthernTotalLimit, true); err != nil {
		return err
	}
	if len(path.NorthernTotalLimit) != len(path.SouthernTotalLimit) {
		return fmt.Errorf("geojson: total northern and southern limits must have the same sample count")
	}
	if err := validateOccultationContours(
		"total band contours", path.TotalBandContours, path.TotalStart.Time, path.TotalEnd.Time,
	); err != nil {
		return err
	}
	if err := validateOccultationContours(
		"total visibility contours", path.TotalVisibilityContours, path.TotalStart.Time, path.TotalEnd.Time,
	); err != nil {
		return err
	}
	for index := range path.NorthernTotalLimit {
		if !path.NorthernTotalLimit[index].Time.Equal(path.SouthernTotalLimit[index].Time) {
			return fmt.Errorf("geojson: total limit sample %d times must match", index)
		}
	}
	last := len(path.NorthernTotalLimit) - 1
	if !path.NorthernTotalLimit[0].Time.Equal(path.TotalStart.Time) ||
		!path.SouthernTotalLimit[0].Time.Equal(path.TotalStart.Time) ||
		!path.NorthernTotalLimit[last].Time.Equal(path.TotalEnd.Time) ||
		!path.SouthernTotalLimit[last].Time.Equal(path.TotalEnd.Time) {
		return fmt.Errorf("geojson: total limits must span total start through total end")
	}
	if err := occultationgeo.ValidateRiseSetCurves(path.TotalRiseSetCurves, path.TotalStart.Time, path.TotalEnd.Time); err != nil {
		return fmt.Errorf("geojson: invalid total occultation rise/set curves: %w", err)
	}
	if err := validateOccultationFootprints("partial", path.PartialFootprints, path.Start.Time, path.End.Time); err != nil {
		return err
	}
	if err := validateOccultationFootprints("partial compact band", path.PartialBandFootprints, path.Start.Time, path.End.Time); err != nil {
		return err
	}
	if err := validateOccultationFootprints("total", path.TotalFootprints, path.TotalStart.Time, path.TotalEnd.Time); err != nil {
		return err
	}
	return validateOccultationFootprints("total compact band", path.TotalBandFootprints, path.TotalStart.Time, path.TotalEnd.Time)
}

func validateOccultationContours(
	name string,
	contours [][]moon.OccultationPathPoint,
	start, end time.Time,
) error {
	for index, contour := range contours {
		if err := validateOccultationPathSeries(fmt.Sprintf("%s[%d]", name, index), contour, true); err != nil {
			return err
		}
		if contour[0].Time.Before(start) || contour[len(contour)-1].Time.After(end) {
			return fmt.Errorf("geojson: %s[%d] must stay inside its contact interval", name, index)
		}
	}
	return nil
}

func validateOccultationPathSeries(name string, points []moon.OccultationPathPoint, required bool) error {
	if required && len(points) < 2 {
		return fmt.Errorf("geojson: %s requires at least two points", name)
	}
	previous := time.Time{}
	for index, point := range points {
		if err := validateOccultationPathPoint(fmt.Sprintf("%s[%d]", name, index), point); err != nil {
			return err
		}
		if !previous.IsZero() && !point.Time.After(previous) {
			return fmt.Errorf("geojson: %s times must be strictly increasing", name)
		}
		previous = point.Time
	}
	return nil
}

func validateOccultationPathPoint(name string, point moon.OccultationPathPoint) error {
	if point.Time.IsZero() {
		return fmt.Errorf("geojson: %s time is required", name)
	}
	if err := validateCoordinate(point.Longitude, point.Latitude); err != nil {
		return fmt.Errorf("geojson: %s: %w", name, err)
	}
	if !finiteGeoJSON(point.MoonAltitude) || point.MoonAltitude < -90 || point.MoonAltitude > 90 {
		return fmt.Errorf("geojson: %s Moon altitude must be finite and within [-90, 90]", name)
	}
	if !finiteGeoJSON(point.WidthKM) || point.WidthKM < 0 {
		return fmt.Errorf("geojson: %s width must be finite and non-negative", name)
	}
	if !finiteGeoJSON(point.LimitSeparationKM) || point.LimitSeparationKM < 0 {
		return fmt.Errorf("geojson: %s limit separation must be finite and non-negative", name)
	}
	return nil
}

func validateOccultationFootprints(
	name string,
	footprints []moon.PlanetOccultationFootprint,
	start, end time.Time,
) error {
	if err := occultationgeo.ValidateFootprints(footprints, start, end); err != nil {
		return fmt.Errorf("geojson: %s footprints: %w", name, err)
	}
	return nil
}

// occultationFootprintSignature 与太阳侧的签名同构：由物理接触弧的顶点数、分段数与闭合标志给出。
func occultationFootprintSignature(
	footprint *moon.PlanetOccultationFootprint,
	prefix string,
) string {
	if footprint == nil || len(footprint.Boundaries) == 0 {
		return "empty"
	}
	vertices, pole := 0, false
	for _, segment := range footprint.Boundaries {
		vertices += len(segment)
		winding := 0.0
		for index := 1; index < len(segment); index++ {
			winding += math.Remainder(segment[index].Longitude-segment[index-1].Longitude, 360)
		}
		if len(footprint.Boundaries) == 1 && math.Abs(winding) >= 180 {
			pole = true
		}
	}
	state := "open"
	if footprint.Closed {
		state = "closed"
	}
	signature := fmt.Sprintf("%s-%s-seg%d-pt%d", prefix, state, len(footprint.Boundaries), vertices)
	if pole {
		signature += "-pole"
	}
	return signature
}

// addOccultationClosureProperties 声明掩星可见区的闭合弧：参照物是月球地平而非太阳。
func addOccultationClosureProperties(
	properties map[string]interface{},
	footprint *moon.PlanetOccultationFootprint,
	value time.Time,
	sublunarLongitude, sublunarLatitude float64,
) {
	if footprint == nil {
		return
	}
	properties["source_boundary_closed"] = footprint.Closed
	if footprint.Closed {
		return
	}
	properties["geometry_role"] = "horizon-closed-region"
	closure := map[string]interface{}{
		"kind": "target-horizon",
		"body": "moon",
		"time": formatTime(value),
	}
	if !math.IsNaN(sublunarLongitude) && !math.IsNaN(sublunarLatitude) {
		closure["sublunar"] = []float64{sublunarLongitude, sublunarLatitude}
	}
	properties["closure"] = closure
}
