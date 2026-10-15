package geojson

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/starainrt/astro/basic"
	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/internal/lunarhorizon"
	"github.com/starainrt/astro/internal/solarclosure"
)

func validateSolarEclipseInput(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	central *eclipsecore.SolarEclipsePath,
) error {
	info := partial.Eclipse
	if info.GreatestEclipse.IsZero() {
		return fmt.Errorf("geojson: solar eclipse greatest time is required")
	}
	if !info.HasPartial {
		return fmt.Errorf("geojson: solar eclipse must contain a partial phase")
	}
	if info.Type == eclipsecore.SolarEclipsePartial && len(partial.CentralBandFootprints) > 0 {
		return fmt.Errorf("geojson: partial solar eclipse cannot contain central band footprints")
	}
	if info.PartialBeginOnEarth.IsZero() || info.PartialEndOnEarth.IsZero() {
		return fmt.Errorf("geojson: solar eclipse partial contact times are required")
	}
	if !info.PartialBeginOnEarth.Before(info.GreatestEclipse) ||
		!info.GreatestEclipse.Before(info.PartialEndOnEarth) {
		return fmt.Errorf("geojson: solar eclipse times must be ordered partial begin, greatest, partial end")
	}
	if !finiteGeoJSON(info.Magnitude) || info.Magnitude <= 0 {
		return fmt.Errorf("geojson: solar eclipse magnitude must be positive and finite")
	}
	if err := validateSolarPathPoint("solar greatest", eclipsecore.SolarEclipsePathPoint{
		Time: info.GreatestEclipse, Longitude: info.GreatestLongitude, Latitude: info.GreatestLatitude,
	}); err != nil {
		return err
	}
	if err := validateSolarFootprints(
		"partial", partial.Footprints, info.PartialBeginOnEarth, info.PartialEndOnEarth,
	); err != nil {
		return err
	}
	for _, contact := range []struct {
		name  string
		point eclipsecore.SolarEclipsePathPoint
	}{
		{"P1", partial.P1}, {"P2", partial.P2}, {"P3", partial.P3}, {"P4", partial.P4},
		{"U1", partial.U1}, {"U2", partial.U2}, {"U3", partial.U3}, {"U4", partial.U4},
	} {
		if contact.point.Time.IsZero() {
			continue
		}
		if err := validateSolarPathPoint("solar "+contact.name, contact.point); err != nil {
			return err
		}
		if !solarEclipseTimeInsideInterval(
			contact.point.Time, info.PartialBeginOnEarth, info.PartialEndOnEarth,
		) {
			return fmt.Errorf("geojson: solar %s time is outside the partial interval", contact.name)
		}
	}
	if err := validateSolarContactSequence(
		"penumbral", partial.P1, partial.P2, partial.P3, partial.P4,
	); err != nil {
		return err
	}
	if err := validateSolarContactSequence(
		"central-shadow", partial.U1, partial.U2, partial.U3, partial.U4,
	); err != nil {
		return err
	}
	centralShadowStart, centralShadowEnd := partial.U1.Time, partial.U4.Time
	if len(partial.CentralShadowFootprints) > 0 || len(partial.CentralBandFootprints) > 0 {
		if centralShadowStart.IsZero() || centralShadowEnd.IsZero() || !centralShadowStart.Before(centralShadowEnd) {
			return fmt.Errorf("geojson: solar central-shadow footprints require ordered U1 and U4 contacts")
		}
	}
	if err := validateSolarFootprints(
		"central-shadow", partial.CentralShadowFootprints, centralShadowStart, centralShadowEnd,
	); err != nil {
		return err
	}
	if err := validateSolarFootprints(
		"central-band", partial.CentralBandFootprints, centralShadowStart, centralShadowEnd,
	); err != nil {
		return err
	}
	if len(partial.CentralBandHorizonClosures) != 0 && len(partial.CentralBandHorizonClosures) != 2 {
		return fmt.Errorf("geojson: solar central-band horizon closures require start and end arcs")
	}
	for index, closure := range partial.CentralBandHorizonClosures {
		if len(closure) < 2 {
			return fmt.Errorf("geojson: solar central-band horizon closure %d requires at least two points", index)
		}
		for pointIndex, point := range closure {
			if err := validateSolarPathPoint(
				fmt.Sprintf("solar central-band horizon closure %d point %d", index, pointIndex), point,
			); err != nil {
				return err
			}
		}
		for _, point := range closure {
			if !solarEclipseTimeInsideInterval(point.Time, centralShadowStart, centralShadowEnd) {
				return fmt.Errorf("geojson: solar central-band horizon closure %d is outside U1-U4", index)
			}
		}
	}
	for segmentIndex, segment := range partial.PartialBandContours {
		if err := validateSolarMagnitudeContourSeries(
			fmt.Sprintf("solar partial-band contour %d", segmentIndex), segment, true,
		); err != nil {
			return err
		}
	}
	for index, contour := range partial.MagnitudeContours {
		if !finiteGeoJSON(contour.Magnitude) || contour.Magnitude <= 0 || contour.Magnitude > info.Magnitude+1e-9 {
			return fmt.Errorf("geojson: solar magnitude contour %d must be positive and no greater than the eclipse magnitude", index)
		}
		if len(contour.Segments) > 0 {
			for segmentIndex, segment := range contour.Segments {
				if err := validateSolarMagnitudeContourSeries(
					fmt.Sprintf("solar magnitude contour %d segment %d", index, segmentIndex), segment, true,
				); err != nil {
					return err
				}
			}
			continue
		}
		if err := validateSolarPathSeries("solar northern magnitude contour", contour.NorthernLimit, true); err != nil {
			return err
		}
		if err := validateSolarPathSeries("solar southern magnitude contour", contour.SouthernLimit, true); err != nil {
			return err
		}
	}
	if err := validateSolarGreatestTimeContours(partial.GreatestTimeContours); err != nil {
		return err
	}
	if err := validateSolarRiseSetCurves(
		partial.RiseSetCurves, info.PartialBeginOnEarth, info.PartialEndOnEarth,
	); err != nil {
		return err
	}
	if central == nil {
		return nil
	}
	if !central.Eclipse.GreatestEclipse.Equal(info.GreatestEclipse) ||
		central.Eclipse.Type != info.Type || central.Eclipse.Model != info.Model {
		return fmt.Errorf("geojson: partial footprints and central path describe different eclipses")
	}
	if central.Eclipse.CentralBeginOnEarth.IsZero() || central.Eclipse.CentralEndOnEarth.IsZero() ||
		!central.Eclipse.CentralBeginOnEarth.Before(central.Eclipse.GreatestEclipse) ||
		!central.Eclipse.GreatestEclipse.Before(central.Eclipse.CentralEndOnEarth) {
		return fmt.Errorf("geojson: solar central path contact times are invalid")
	}
	if err := validateSolarPathPoint("solar central greatest", central.Greatest); err != nil {
		return err
	}
	if !central.Greatest.Time.Equal(central.Eclipse.GreatestEclipse) {
		return fmt.Errorf("geojson: solar central greatest time does not match eclipse greatest")
	}
	if err := validateSolarPathSeries("solar center line", central.CenterLine, true); err != nil {
		return err
	}
	if central.Greatest.Time.Before(central.CenterLine[0].Time) ||
		central.Greatest.Time.After(central.CenterLine[len(central.CenterLine)-1].Time) {
		return fmt.Errorf("geojson: solar greatest time is outside the center-line interval")
	}
	if central.CenterLine[0].Time.Before(central.Eclipse.CentralBeginOnEarth) ||
		central.CenterLine[len(central.CenterLine)-1].Time.After(central.Eclipse.CentralEndOnEarth) {
		return fmt.Errorf("geojson: solar center line is outside the central interval")
	}
	if len(central.NorthernLimit) != len(central.SouthernLimit) {
		return fmt.Errorf("geojson: solar central limits must have the same sample count")
	}
	if len(central.NorthernLimit) > 0 {
		if err := validateSolarPathSeries("solar northern limit", central.NorthernLimit, true); err != nil {
			return err
		}
		if err := validateSolarPathSeries("solar southern limit", central.SouthernLimit, true); err != nil {
			return err
		}
		for index := range central.NorthernLimit {
			if !central.NorthernLimit[index].Time.Equal(central.SouthernLimit[index].Time) {
				return fmt.Errorf("geojson: solar central limit sample %d times must match", index)
			}
		}
	}
	return nil
}

func validateSolarContactSequence(
	name string,
	contacts ...eclipsecore.SolarEclipsePathPoint,
) error {
	previous := time.Time{}
	for index, contact := range contacts {
		if contact.Time.IsZero() {
			continue
		}
		if !previous.IsZero() && !previous.Before(contact.Time) {
			return fmt.Errorf("geojson: solar %s contacts must be strictly ordered at %d", name, index)
		}
		previous = contact.Time
	}
	return nil
}

func validateSolarFootprints(
	name string,
	footprints []eclipsecore.SolarEclipsePartialFootprint,
	start, end time.Time,
) error {
	previous := time.Time{}
	for footprintIndex, footprint := range footprints {
		if footprint.Time.IsZero() {
			return fmt.Errorf("geojson: solar %s footprint %d time is required", name, footprintIndex)
		}
		if !previous.IsZero() && !footprint.Time.After(previous) {
			return fmt.Errorf("geojson: solar %s footprint times must be strictly increasing", name)
		}
		if footprint.Time.Before(start) || footprint.Time.After(end) {
			return fmt.Errorf("geojson: solar %s footprint %d time is outside its event interval", name, footprintIndex)
		}
		if len(footprint.Boundaries) == 0 {
			return fmt.Errorf("geojson: solar %s footprint %d has no boundary", name, footprintIndex)
		}
		for segmentIndex, segment := range footprint.Boundaries {
			if len(segment) == 0 {
				return fmt.Errorf("geojson: solar %s footprint %d boundary %d is empty", name, footprintIndex, segmentIndex)
			}
			for pointIndex, point := range segment {
				if err := validateSolarPathPoint(fmt.Sprintf(
					"solar %s footprint %d boundary %d point %d", name, footprintIndex, segmentIndex, pointIndex,
				), point); err != nil {
					return err
				}
				if !point.Time.Equal(footprint.Time) {
					return fmt.Errorf("geojson: solar %s footprint point time must match its footprint", name)
				}
			}
		}
		previous = footprint.Time
	}
	return nil
}

func validateSolarRiseSetCurves(
	curves []eclipsecore.SolarEclipseRiseSetCurve,
	start, end time.Time,
) error {
	seen := make(map[[2]string]bool, len(curves))
	for curveIndex, curve := range curves {
		if curve.Phase != eclipsecore.RiseSetPhaseStart &&
			curve.Phase != eclipsecore.RiseSetPhaseGreatest && curve.Phase != eclipsecore.RiseSetPhaseEnd {
			return fmt.Errorf("geojson: solar rise/set curve %d has unsupported phase %q", curveIndex, curve.Phase)
		}
		if curve.Direction != eclipsecore.RiseSetDirectionRise && curve.Direction != eclipsecore.RiseSetDirectionSet {
			return fmt.Errorf("geojson: solar rise/set curve %d has unsupported direction %q", curveIndex, curve.Direction)
		}
		key := [2]string{string(curve.Phase), string(curve.Direction)}
		if seen[key] {
			return fmt.Errorf("geojson: solar rise/set curve %d duplicates phase %q and direction %q", curveIndex, curve.Phase, curve.Direction)
		}
		seen[key] = true
		if len(curve.Segments) == 0 {
			return fmt.Errorf("geojson: solar rise/set curve %d has no segments", curveIndex)
		}
		for segmentIndex, segment := range curve.Segments {
			if len(segment) < 2 {
				return fmt.Errorf("geojson: solar rise/set curve %d segment %d requires at least two points", curveIndex, segmentIndex)
			}
			previous := time.Time{}
			for pointIndex, point := range segment {
				if err := validateSolarPathPoint(fmt.Sprintf(
					"solar rise/set curve %d segment %d point %d", curveIndex, segmentIndex, pointIndex,
				), point); err != nil {
					return err
				}
				if !solarEclipseTimeInsideInterval(point.Time, start, end) {
					return fmt.Errorf("geojson: solar rise/set curve point is outside the partial interval")
				}
				if !previous.IsZero() && !point.Time.After(previous) {
					return fmt.Errorf("geojson: solar rise/set curve segment times must be strictly increasing")
				}
				previous = point.Time
			}
		}
	}
	return nil
}

func solarEclipseTimeInsideInterval(value, start, end time.Time) bool {
	return !value.Before(start.Add(-solarEclipseValidationTimeTolerance)) &&
		!value.After(end.Add(solarEclipseValidationTimeTolerance))
}

func validateSolarMagnitudeContourSeries(
	name string,
	points []eclipsecore.SolarEclipsePathPoint,
	required bool,
) error {
	if required && len(points) < 2 {
		return fmt.Errorf("geojson: %s requires at least two points", name)
	}
	for index, point := range points {
		if err := validateSolarPathPoint(fmt.Sprintf("%s[%d]", name, index), point); err != nil {
			return err
		}
	}
	return nil
}

func validateSolarGreatestTimeContours(
	contours []eclipsecore.SolarEclipseGreatestTimeContour,
) error {
	for index, contour := range contours {
		if !finiteGeoJSON(contour.JDE) || contour.JDE == 0 {
			return fmt.Errorf("geojson: solar greatest-time contour %d JDE must be finite and non-zero", index)
		}
		if contour.Time.IsZero() {
			return fmt.Errorf("geojson: solar greatest-time contour %d time is required", index)
		}
		if len(contour.Segments) == 0 {
			return fmt.Errorf("geojson: solar greatest-time contour %d has no branches", index)
		}
		for segmentIndex, segment := range contour.Segments {
			if err := validateSolarMagnitudeContourSeries(
				fmt.Sprintf("solar greatest-time contour %d branch %d", index, segmentIndex), segment, true,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateSolarPathSeries(name string, points []eclipsecore.SolarEclipsePathPoint, required bool) error {
	if required && len(points) < 2 {
		return fmt.Errorf("geojson: %s requires at least two points", name)
	}
	previous := time.Time{}
	for index, point := range points {
		if err := validateSolarPathPoint(fmt.Sprintf("%s[%d]", name, index), point); err != nil {
			return err
		}
		if !previous.IsZero() && !point.Time.After(previous) {
			return fmt.Errorf("geojson: %s times must be strictly increasing", name)
		}
		previous = point.Time
	}
	return nil
}

func validateSolarPathPoint(name string, point eclipsecore.SolarEclipsePathPoint) error {
	if point.Time.IsZero() {
		return fmt.Errorf("geojson: %s time is required", name)
	}
	if err := validateCoordinate(point.Longitude, point.Latitude); err != nil {
		return fmt.Errorf("geojson: %s: %w", name, err)
	}
	if !finiteGeoJSON(point.SunAltitude) || point.SunAltitude < -90 || point.SunAltitude > 90 {
		return fmt.Errorf("geojson: %s sun altitude must be finite and within [-90, 90]", name)
	}
	if !finiteGeoJSON(point.WidthKM) || point.WidthKM < 0 {
		return fmt.Errorf("geojson: %s width must be finite and non-negative", name)
	}
	return nil
}

func validateLunarEclipseInfo(info eclipsecore.LunarEclipseInfo) error {
	if !info.HasPenumbral || info.PenumbralStart.IsZero() || info.PenumbralEnd.IsZero() {
		return fmt.Errorf("geojson: lunar eclipse penumbral contact times are required")
	}
	if info.Maximum.IsZero() {
		return fmt.Errorf("geojson: lunar eclipse greatest time is required")
	}
	if info.Type != eclipsecore.LunarEclipsePenumbral && info.Type != eclipsecore.LunarEclipsePartial &&
		info.Type != eclipsecore.LunarEclipseTotal {
		return fmt.Errorf("geojson: lunar eclipse type is invalid")
	}
	switch info.Type {
	case eclipsecore.LunarEclipsePenumbral:
		if info.HasPartial || info.HasTotal {
			return fmt.Errorf("geojson: penumbral eclipse cannot contain partial or total phases")
		}
	case eclipsecore.LunarEclipsePartial:
		if !info.HasPartial || info.HasTotal {
			return fmt.Errorf("geojson: partial eclipse must contain only a partial phase")
		}
	case eclipsecore.LunarEclipseTotal:
		if !info.HasPartial || !info.HasTotal {
			return fmt.Errorf("geojson: total eclipse must contain partial and total phases")
		}
	}
	if !info.HasPartial && (!info.PartialStart.IsZero() || !info.PartialEnd.IsZero()) {
		return fmt.Errorf("geojson: partial contact times require a partial phase")
	}
	if !info.HasTotal && (!info.TotalStart.IsZero() || !info.TotalEnd.IsZero()) {
		return fmt.Errorf("geojson: total contact times require a total phase")
	}
	ordered := []time.Time{info.PenumbralStart}
	if info.HasPartial {
		if info.PartialStart.IsZero() || info.PartialEnd.IsZero() {
			return fmt.Errorf("geojson: lunar eclipse partial contact times are required")
		}
		ordered = append(ordered, info.PartialStart)
	}
	if info.HasTotal {
		if info.TotalStart.IsZero() || info.TotalEnd.IsZero() {
			return fmt.Errorf("geojson: lunar eclipse total contact times are required")
		}
		ordered = append(ordered, info.TotalStart)
	}
	ordered = append(ordered, info.Maximum)
	if info.HasTotal {
		ordered = append(ordered, info.TotalEnd)
	}
	if info.HasPartial {
		ordered = append(ordered, info.PartialEnd)
	}
	ordered = append(ordered, info.PenumbralEnd)
	for index := 1; index < len(ordered); index++ {
		if !ordered[index-1].Before(ordered[index]) {
			return fmt.Errorf("geojson: lunar eclipse contact times are not strictly ordered")
		}
	}
	return nil
}

const (
	solarEclipseEvent                            = "solar-eclipse"
	lunarEclipseEvent                            = "lunar-eclipse"
	solarEclipseValidationTimeTolerance          = 3 * time.Minute
	solarEclipseCentralBandMinimumBoundaryPoints = 90

	defaultLunarBoundaryPoints = 360
	minimumLunarBoundaryPoints = 12
	maximumLunarBoundaryPoints = 1440
)

// MarshalSolarEclipse 将日食半影足迹和可选中心食带编码为 GeoJSON / encode solar eclipse footprints and the central band as GeoJSON.
func (module *Generator) MarshalSolarEclipse(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	central *eclipsecore.SolarEclipsePath,
) ([]byte, error) {
	return marshalSolarEclipse(partial, central, SolarEclipseOptions{}, module)
}

// SolarEclipseOptions 控制日食 GeoJSON 的输出内容。
// SolarEclipseOptions controls the solar-eclipse GeoJSON content.
type SolarEclipseOptions struct {
	// TimeMarkers 非空时沿中心线追加时间标记 Point 要素，等价于 MarshalSolarEclipseWithTimeMarkers。
	// TimeMarkers adds time-marker Point Features along the center line when non-nil.
	TimeMarkers *TimeMarkerOptions
	// SkipRoles 列出不写进输出的 role，例如 partial-footprint（瞬时半影轮廓）、partial-band、
	// magnitude-line、visibility-boundary。只丢要素，不改变几何：偏食域包络仍用完整采样闭合，
	// 因此跳过 partial-footprint 不会降低 partial-band 的精度。全部要素都被丢掉时返回错误。
	// SkipRoles lists roles to leave out of the output, such as partial-footprint, partial-band,
	// magnitude-line or visibility-boundary. It drops Features only and does not change geometry:
	// the partial band is still closed from the complete sampling, so skipping the instantaneous
	// penumbral outlines costs no accuracy. Returns an error when every Feature is dropped.
	SkipRoles []string
}

// MarshalSolarEclipseWithOptions 编码日食，输出内容由 options 选择 / encodes a solar eclipse with the content selected by options.
func (module *Generator) MarshalSolarEclipseWithOptions(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	central *eclipsecore.SolarEclipsePath,
	options SolarEclipseOptions,
) ([]byte, error) {
	return marshalSolarEclipse(partial, central, options, module)
}

// MarshalSolarEclipseWithTimeMarkers 编码日食，并沿中心线按固定间隔追加 Point 要素；已有要素不变，标记标签使用 options.Location，时间值保持 UTC / encode a solar eclipse with time markers along the centre line.
func (module *Generator) MarshalSolarEclipseWithTimeMarkers(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	central *eclipsecore.SolarEclipsePath,
	options TimeMarkerOptions,
) ([]byte, error) {
	return marshalSolarEclipse(partial, central, SolarEclipseOptions{TimeMarkers: &options}, module)
}

func marshalSolarEclipse(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	central *eclipsecore.SolarEclipsePath,
	options SolarEclipseOptions, modules ...*Generator) ([]byte, error) {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	markerOptions := options.TimeMarkers
	if markerOptions != nil {
		if err := validateTimeMarkerOptions(*markerOptions); err != nil {
			return nil, err
		}
	}
	if len(partial.Footprints) == 0 {
		return nil, fmt.Errorf("geojson: solar eclipse has no partial footprints")
	}
	if err := validateSolarEclipseInput(partial, central); err != nil {
		return nil, err
	}
	timeScale, scaleErr := timeScaleForMarkers(markerOptions)
	if scaleErr != nil {
		return nil, scaleErr
	}
	// 几何（月下点、地平闭合弧、带宽限界）必须用民用时刻算：只有写进属性的时刻换时标。
	geometryPartial := partial
	civilCentral := central
	if timeScale == basic.TimeScaleUT1 {
		partial = eclipsecore.SolarEclipsePartialFootprintsInUT1(partial)
		if central != nil {
			converted := eclipsecore.SolarEclipsePathInUT1(*central)
			central = &converted
		}
	}

	properties := map[string]interface{}{
		"eclipse_type": string(partial.Eclipse.Type),
		"model":        string(partial.Eclipse.Model),
	}
	features := make([]feature, 0, len(partial.Footprints)+9)
	for footprintIndex, footprint := range partial.Footprints {
		// 闭合弧由月下点决定，必须用未换时标的同一足迹算几何。
		polygon, err := solarPartialFootprintPolygon(geometryPartial.Footprints[footprintIndex], true, module)
		if err != nil {
			return nil, err
		}
		curve, err := solarShadowFootprintCurveFromSegments(footprint.Boundaries)
		if err != nil {
			return nil, err
		}
		if solarShadowRegionDegenerate(curve, polygon) {
			// 与单时刻导出同口径：退化区域整条缺省，不退化成点或零面积环。
			continue
		}
		footprintProperties := cloneProperties(properties)
		footprintProperties["time"] = formatTime(footprint.Time)
		footprintProperties["source_boundary_closed"] = footprint.Closed
		footprintProperties["interp_signature"] = solarShadowFootprintSignature(
			footprint.Boundaries, footprint.Closed, eclipsecore.SolarEclipseShadowPenumbra,
		)
		if !footprint.Closed {
			footprintProperties["geometry_role"] = "horizon-closed-region"
			footprintProperties["closure"] = solarHorizonClosureProperties(
				geometryPartial.Footprints[footprintIndex].Time, footprint.Time,
				solarHorizonClosureExact(footprint.Boundaries, footprint.HorizonEnds), module,
			)
		}
		if len(polygon) == 1 {
			value, pointErr := pointGeometry(polygon[0].Longitude, polygon[0].Latitude)
			if pointErr != nil {
				return nil, fmt.Errorf("geojson: solar partial footprint at %s: %w", formatTime(footprint.Time), pointErr)
			}
			features = append(features, newFeature(
				solarEclipseEvent, "partial-footprint", value, footprintProperties,
			))
			continue
		}
		value, err := multiPolygonFillGeometry([][]geodata.GeoPoint{polygon})
		if err != nil {
			return nil, fmt.Errorf("geojson: solar partial footprint at %s: %w", formatTime(footprint.Time), err)
		}
		features = append(features, newFeature(
			solarEclipseEvent, "partial-footprint", value, footprintProperties,
		))
	}

	if value, source, ok, err := solarPartialBandGeometry(geometryPartial, module); err != nil {
		return nil, fmt.Errorf("geojson: solar partial band: %w", err)
	} else if ok {
		bandProperties := cloneProperties(properties)
		bandProperties["source"] = source
		features = append(features, newFeature(
			solarEclipseEvent, "partial-band", value, bandProperties,
		))
	}

	var err error
	features, err = appendSolarRiseSetCurveFeatures(features, partial.RiseSetCurves, properties)
	if err != nil {
		return nil, err
	}
	features, err = appendSolarFootprintFeatures(
		features, "central-shadow-footprint", partial.CentralShadowFootprints, properties, module,
	)
	if err != nil {
		return nil, err
	}
	bandFootprints := solarCentralBandFootprints(partial)
	// Keep the densest shadow footprints selected by solarCentralBandFootprints.
	// The lightweight companion is sufficient for ordinary closed envelopes, but
	// polar two-limit fallback needs the exact U1/U4 endpoint sweep.
	if central == nil && len(bandFootprints) > 0 {
		band, bandSource, bandErr := solarCentralBandEnvelopeGeometry(partial.CentralBandSegments)
		if bandErr != nil {
			band, bandErr = solarCentralShadowSweepGeometry(bandFootprints)
			bandSource = "central-shadow-sweep"
			if bandErr != nil && len(partial.CentralBandFootprints) > 0 &&
				!sameSolarFootprintSlice(bandFootprints, partial.CentralBandFootprints) {
				band, bandErr = solarCentralShadowSweepGeometry(partial.CentralBandFootprints)
			}
		}
		if bandErr != nil {
			return nil, fmt.Errorf("geojson: solar central band: %w", bandErr)
		}
		bandProperties := cloneProperties(properties)
		bandProperties["centrality"] = string(partial.Eclipse.Centrality)
		bandProperties["source"] = bandSource
		features = append(features, newFeature(
			solarEclipseEvent, "central-band", band, bandProperties,
		))
	}
	for _, contour := range partial.MagnitudeContours {
		if len(contour.Segments) > 0 {
			contourProperties := cloneProperties(properties)
			contourProperties["magnitude"] = contour.Magnitude
			features, err = appendSolarSegmentedPathLine(
				features, "magnitude-line", contour.Segments, contourProperties, false,
			)
			if err != nil {
				return nil, err
			}
			continue
		}
		for _, side := range []struct {
			name   string
			points []eclipsecore.SolarEclipsePathPoint
		}{
			{name: "north", points: contour.NorthernLimit},
			{name: "south", points: contour.SouthernLimit},
		} {
			contourProperties := cloneProperties(properties)
			contourProperties["magnitude"] = contour.Magnitude
			contourProperties["side"] = side.name
			features, err = appendSolarPathLine(
				features, "magnitude-line", side.points, contourProperties,
			)
			if err != nil {
				return nil, err
			}
		}
	}
	for _, contour := range partial.GreatestTimeContours {
		for _, segment := range contour.Segments {
			contourProperties := cloneProperties(properties)
			contourProperties["time"] = formatTime(contour.Time)
			contourProperties["jde"] = contour.JDE
			// 支路各点同为该时刻，逐点时间不是递增序列。
			features, err = appendSolarSegmentedPathLine(
				features, "greatest-time-line", [][]eclipsecore.SolarEclipsePathPoint{segment},
				contourProperties, false,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	if central != nil {
		// bandFootprints is the presentation subset the ribbon and the closed
		// envelopes are validated against; one-limit events trim the U1/U4 tails
		// out of it. sweepFootprints keeps the complete umbral sweep, because a
		// grazing one-limit path really does extend over that whole interval
		// (NASA's path table lists its limits from U1 to U4), so a band built or
		// validated only against the trimmed subset silently loses the flared
		// ends of the real annular/total region.
		// 带宽与限界几何按民用时刻构造；写出的时刻仍取换过时标的 central。
		geometryCentral := central
		if timeScale == basic.TimeScaleUT1 && civilCentral != nil {
			geometryCentral = civilCentral
		}
		bandFootprints = solarCentralBandFootprintsForPath(geometryPartial, geometryCentral)
		sweepFootprints := solarCentralBandFootprints(geometryPartial)
		// The exported limit lines are trimmed to the center-line interval for
		// ordinary maps, but the static band must be built from the complete
		// U1/U4 paired limits: for a shallow two-limit event the axis interval is
		// a fraction of the umbral window, and a band built from the trimmed
		// limits drops hundreds of kilometres of real annular area.
		presentationNorthernLimit := central.NorthernLimit
		presentationSouthernLimit := central.SouthernLimit
		geometryNorthernLimit := geometryCentral.NorthernLimit
		geometrySouthernLimit := geometryCentral.SouthernLimit
		if partial.Eclipse.Centrality == eclipsecore.SolarEclipseCentralTwoLimits {
			if north, south, ok := solarCentralTwoLimitPresentationLimits(
				geometryNorthernLimit, geometrySouthernLimit, geometryCentral.CenterLine,
			); ok {
				presentationNorthernLimit, presentationSouthernLimit = north, south
				geometryNorthernLimit, geometrySouthernLimit = north, south
			}
		}
		bandNorthernLimit := geometryNorthernLimit
		bandSouthernLimit := geometrySouthernLimit
		// A grazing band is not bounded by the instantaneous cross-section
		// limits: those stop describing the region and can sit hundreds of
		// kilometres inside it (1136-06-01: 456 km for the northern limit).
		// Whenever the analytic limits no longer follow the band boundary, the
		// exported lines are taken from the band ring itself, so the dashed
		// limits and the filled band describe the same region.
		var derivedNorthernLimit, derivedSouthernLimit []eclipsecore.SolarEclipsePathPoint
		if len(geometryNorthernLimit) > 0 {
			var value geometry
			var source string
			var usedMagnitudeOne bool
			centralEnvelope := geometryPartial.CentralBandSegments
			if len(geometryCentral.CentralBandSegments) > 0 {
				centralEnvelope = geometryCentral.CentralBandSegments
			}
			useCriticalEnvelope := len(centralEnvelope) > 0
			// Check the shadow axis, not the instantaneous cross-section limits:
			// near the horizon those samples can have their local greatest below
			// the horizon and need not belong to the visible central band.
			coveragePath := *geometryCentral
			coveragePath.NorthernLimit = geometryNorthernLimit
			coveragePath.SouthernLimit = geometrySouthernLimit
			if useCriticalEnvelope && !solarCentralBandEnvelopeCoversPath(
				centralEnvelope, &coveragePath,
			) {
				useCriticalEnvelope = false
			}
			if useCriticalEnvelope && !solarCentralBandEnvelopeCoversFootprints(
				centralEnvelope, bandFootprints,
			) {
				useCriticalEnvelope = false
			}
			if useCriticalEnvelope {
				value, source, err = solarCentralBandEnvelopeGeometry(centralEnvelope)
				if partial.Eclipse.Type == eclipsecore.SolarEclipseTotal {
					source = "magnitude-one-envelope"
				}
			} else {
				value, source, usedMagnitudeOne, err = solarCentralMagnitudeOneBandGeometry(
					partial.Eclipse.Type,
					partial.MagnitudeContours,
					geometryCentral.CenterLine,
					geometryPartial.CentralBandHorizonClosures,
				)
			}
			if !useCriticalEnvelope && !usedMagnitudeOne {
				value, source, err = solarCentralBandGeometry(
					bandNorthernLimit,
					bandSouthernLimit,
					geometryCentral.CenterLine,
					partial.Eclipse.Type,
					partial.Eclipse.Centrality,
					sweepFootprints,
					geometryPartial.CentralBandHorizonClosures,
				)
			}
			if err != nil && len(geometryPartial.CentralBandFootprints) > 0 &&
				!sameSolarFootprintSlice(sweepFootprints, geometryPartial.CentralBandFootprints) {
				// A caller may request dense central-shadow samples. Near
				// grazing contacts, the planar sweep can become numerically
				// open; the always-available end-cap samples provide a stable
				// equivalent band without rejecting the whole export.
				value, source, err = solarCentralBandGeometry(
					bandNorthernLimit,
					bandSouthernLimit,
					geometryCentral.CenterLine,
					partial.Eclipse.Type,
					partial.Eclipse.Centrality,
					geometryPartial.CentralBandFootprints,
					partial.CentralBandHorizonClosures,
				)
			}
			if err != nil {
				return nil, fmt.Errorf("geojson: solar central band: %w", err)
			}
			// Only a band rebuilt from sampled footprints carries the sampling
			// ripple the snap removes; an analytic envelope is already the exact
			// boundary and must keep its own end caps.
			if partial.CentralBandSampled || central.CentralBandSampled {
				value = snapSolarBandGeometryToHorizonCurves(value, partial.RiseSetCurves)
			}
			bandProperties := cloneProperties(properties)
			bandProperties["source"] = source
			features = append(features, newFeature(
				solarEclipseEvent, "central-band", value, bandProperties,
			))
		} else if len(sweepFootprints) > 0 {
			// A one-limit event may publish no paired limits at all. Fall back to
			// the complete umbral sweep and make sure the exported band still
			// contains its own center line.
			polygons, sweepErr := solarCentralShadowSweepPolygons(sweepFootprints)
			if sweepErr != nil {
				return nil, fmt.Errorf("geojson: solar central band: %w", sweepErr)
			}
			value, geometryErr := multiPolygonGeometry(
				solarCentralBandWithCenterlineCorridor(polygons, geometryCentral.CenterLine),
			)
			if geometryErr != nil {
				return nil, fmt.Errorf("geojson: solar central band: %w", geometryErr)
			}
			bandProperties := cloneProperties(properties)
			bandProperties["source"] = "central-shadow-sweep"
			features = append(features, newFeature(
				solarEclipseEvent, "central-band", value, bandProperties,
			))
		}

		// Derive the exported limits from whichever band was built above: a
		// grazing band is not bounded by the instantaneous cross-section limits,
		// which stop describing the region and can sit hundreds of kilometres
		// inside it (1136-06-01: 456 km for the northern limit). A band split at
		// the antimeridian is rejoined first; when its fragments do not pair up,
		// each ring is cut into runs that stay on one side of the center line.
		if bandGeometry, ok := solarEclipseBandGeometry(features); ok && len(central.CenterLine) >= 2 {
			geometryRings := solarBandGeometryRings(bandGeometry)
			stitched := stitchSolarBandRings(geometryRings)
			// 单环先接缝再切侧；两条分支共用同一逐点投影侧判据，标签不会互相矛盾。
			if len(stitched) == 1 {
				geometryRings = stitched
			}
			north, south, derived := solarCentralBandLimitSidesFromRings(geometryRings, central.CenterLine)
			if derived && (solarCentralLimitSeparationKM(
				presentationNorthernLimit, north,
			) > solarCentralBandLimitSidesSplitKM || solarCentralLimitSeparationKM(
				presentationSouthernLimit, south,
			) > solarCentralBandLimitSidesSplitKM) {
				derivedNorthernLimit, derivedSouthernLimit = north, south
			}
		}
		if len(derivedNorthernLimit) > 0 {
			presentationNorthernLimit, presentationSouthernLimit = derivedNorthernLimit, derivedSouthernLimit
		}
		features, err = appendSolarPathLine(features, "center-line", central.CenterLine, properties)
		if err != nil {
			return nil, err
		}
		if len(presentationNorthernLimit) > 0 {
			features, err = appendSolarPathLine(features, "north-limit", presentationNorthernLimit, properties)
			if err != nil {
				return nil, err
			}
			features, err = appendSolarPathLine(features, "south-limit", presentationSouthernLimit, properties)
			if err != nil {
				return nil, err
			}
		}
		if markerOptions != nil {
			features, err = appendTimeMarkerFeatures(
				features,
				solarEclipseEvent,
				"center-line",
				solarPathSamples(central.CenterLine),
				*markerOptions,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	greatest := pathSample{
		Time:      partial.Eclipse.GreatestEclipse,
		Longitude: partial.Eclipse.GreatestLongitude,
		Latitude:  partial.Eclipse.GreatestLatitude,
	}
	greatestProperties := solarEclipseMetadata(partial.Eclipse)
	if central != nil {
		greatest = solarPathSample(central.Greatest)
		greatestProperties["width_km"] = central.Greatest.WidthKM
		greatestProperties["sun_altitude_deg"] = central.Greatest.SunAltitude
		if central.MaxCentralDuration > 0 {
			// The longest central phase anywhere on the track, which for a
			// shallow event exceeds the value at greatest eclipse.
			greatestProperties["max_central_duration_seconds"] = central.MaxCentralDuration.Seconds()
			greatestProperties["max_central_duration"] = central.MaxCentralDuration.String()
			greatestProperties["max_central_duration_longitude"] = central.MaxCentralDurationLongitude
			greatestProperties["max_central_duration_latitude"] = central.MaxCentralDurationLatitude
		}
	}
	features, err = appendPointFeature(
		features, solarEclipseEvent, "greatest", greatest, greatestProperties,
	)
	if err != nil {
		return nil, err
	}
	return marshalFeatureCollectionWithTimeScale(dropFeaturesByRole(features, options.SkipRoles), timeScale)
}

// solarCentralBandSeamEpsilonKM is the seam tolerance used when rejoining the
// fragments the antimeridian split left in one band boundary.
const solarCentralBandSeamEpsilonKM = 0.5

// solarCentralBandRunConnectKM is the gap below which two boundary runs are
// treated as consecutive pieces of one limit line.
const solarCentralBandRunConnectKM = 25.0

// stitchSolarBandRings rejoins the fragments an antimeridian split produced, so
// the northern and southern sides can be derived from a single loop. Each
// fragment carries meridian edges at the seam; dropping them leaves open chains
// whose endpoints are rejoined at matching latitudes.
func stitchSolarBandRings(
	rings [][]eclipsecore.SolarEclipsePathPoint,
) [][]eclipsecore.SolarEclipsePathPoint {
	if len(rings) < 2 {
		return rings
	}
	type bandChain struct {
		points []eclipsecore.SolarEclipsePathPoint
	}
	var chains []bandChain
	for _, ring := range rings {
		points := openSolarPathRing(ring)
		count := len(points)
		if count < 3 || !solarBandRingTouchesSeam(points) {
			chains = append(chains, bandChain{points: points})
			continue
		}
		seam := make([]bool, count)
		start := -1
		for index := 0; index < count; index++ {
			first, second := points[index], points[(index+1)%count]
			seam[index] = solarBandSeamLongitude(first.Longitude) &&
				solarBandSeamLongitude(second.Longitude) &&
				math.Abs(first.Latitude-second.Latitude) > 1e-9
			if seam[index] && start < 0 {
				start = (index + 1) % count
			}
		}
		if start < 0 {
			chains = append(chains, bandChain{points: points})
			continue
		}
		current := make([]eclipsecore.SolarEclipsePathPoint, 0, count)
		for step := 0; step < count; step++ {
			index := (start + step) % count
			current = append(current, points[index])
			if seam[index] {
				chains = append(chains, bandChain{points: current})
				current = make([]eclipsecore.SolarEclipsePathPoint, 0, count)
			}
		}
		if len(current) > 0 {
			chains = append(chains, bandChain{points: current})
		}
	}
	used := make([]bool, len(chains))
	merged := make([][]eclipsecore.SolarEclipsePathPoint, 0, len(chains))
	for index := range chains {
		if used[index] {
			continue
		}
		used[index] = true
		current := chains[index].points
		for {
			joined := false
			for next := range chains {
				if used[next] {
					continue
				}
				if value, ok := joinSolarBandChains(current, chains[next].points); ok {
					current = value
					used[next] = true
					joined = true
					break
				}
			}
			if !joined {
				break
			}
		}
		if len(current) >= 3 {
			merged = append(merged, current)
		}
	}
	// Close every rejoined loop so downstream code sees whole rings again.
	for index, ring := range merged {
		if len(ring) > 1 && !solarBandPointsCoincide(ring[0], ring[len(ring)-1]) {
			merged[index] = append(ring, ring[0])
		}
	}
	return merged
}

// solarBandRingTouchesSeam reports whether any vertex sits on the antimeridian.
func solarBandRingTouchesSeam(points []eclipsecore.SolarEclipsePathPoint) bool {
	for _, point := range points {
		if solarBandSeamLongitude(point.Longitude) {
			return true
		}
	}
	return false
}

// solarBandSeamLongitude reports whether one longitude lies on the export seam.
func solarBandSeamLongitude(longitude float64) bool {
	return math.Abs(math.Abs(longitude)-180) <= 1e-6
}

// solarBandPointsCoincide compares two path points, wrapping longitudes.
func solarBandPointsCoincide(first, second eclipsecore.SolarEclipsePathPoint) bool {
	if math.Abs(first.Latitude-second.Latitude) > 1e-9 {
		return false
	}
	delta := math.Abs(math.Remainder(first.Longitude-second.Longitude, 360))
	return delta <= 1e-9 || math.Abs(delta-360) <= 1e-9
}

// joinSolarBandChains appends one open chain to another when their seam
// endpoints describe the same latitude on opposite sides of the antimeridian.
func joinSolarBandChains(
	first, second []eclipsecore.SolarEclipsePathPoint,
) ([]eclipsecore.SolarEclipsePathPoint, bool) {
	if len(first) == 0 || len(second) == 0 {
		return nil, false
	}
	reversed := make([]eclipsecore.SolarEclipsePathPoint, len(second))
	for index := range second {
		reversed[index] = second[len(second)-1-index]
	}
	switch {
	case solarBandSeamMatch(first[len(first)-1], second[0]):
		return append(append([]eclipsecore.SolarEclipsePathPoint{}, first...), second[1:]...), true
	case solarBandSeamMatch(first[len(first)-1], second[len(second)-1]):
		return append(append([]eclipsecore.SolarEclipsePathPoint{}, first...), reversed[1:]...), true
	case solarBandSeamMatch(first[0], second[len(second)-1]):
		return append(append([]eclipsecore.SolarEclipsePathPoint{}, second...), first[1:]...), true
	case solarBandSeamMatch(first[0], second[0]):
		return append(append([]eclipsecore.SolarEclipsePathPoint{}, reversed...), first[1:]...), true
	}
	return nil, false
}

// solarBandSeamMatch reports whether two chain ends meet across the seam.
func solarBandSeamMatch(first, second eclipsecore.SolarEclipsePathPoint) bool {
	if math.Abs(first.Latitude-second.Latitude) > 1e-6 {
		return false
	}
	delta := math.Abs(math.Abs(first.Longitude) - math.Abs(second.Longitude))
	if delta > 1e-6 {
		return false
	}
	// Opposite sides of the seam, or the very same meridian point.
	return math.Signbit(first.Longitude) != math.Signbit(second.Longitude) ||
		math.Abs(first.Longitude-second.Longitude) <= 1e-6
}

// solarCentralBandLimitSidesSplitKM is how far the analytic limits may sit from
// the band boundary before the export replaces them with the band's own sides.
// Ordinary events agree to a few kilometres; a grazing band is hundreds of
// kilometres away, because there the instantaneous cross-section limits stop
// describing the boundary of the region at all.
const solarCentralBandLimitSidesSplitKM = 25.0

// solarBandSideRun is one boundary stretch that stays on a single side of the
// center line, with the projected position and time of each of its vertices.
type solarBandSideRun struct {
	points   []eclipsecore.SolarEclipsePathPoint
	times    []time.Time
	progress []float64
	north    bool
}

// solarCentralBandLimitSidesFromRings 由食带边界派生南北限：环顶点按自身投影侧别切段，同侧最长连通组按路径序拼接。
func solarCentralBandLimitSidesFromRings(
	rings [][]eclipsecore.SolarEclipsePathPoint,
	centerLine []eclipsecore.SolarEclipsePathPoint,
) ([]eclipsecore.SolarEclipsePathPoint, []eclipsecore.SolarEclipsePathPoint, bool) {
	if len(centerLine) < 2 {
		return nil, nil, false
	}
	runs, ok := solarBandSideRuns(rings, centerLine)
	if !ok || len(runs) == 0 {
		return nil, nil, false
	}
	northern := solarBandSidePoints(runs, true)
	southern := solarBandSidePoints(runs, false)
	if len(northern) < 3 || len(southern) < 3 {
		return nil, nil, false
	}
	return northern, southern, true
}

// solarBandSideRuns 把每个环按逐点投影侧别切成同侧段；投影沿环连续追踪，中心线回折处不跳支。
func solarBandSideRuns(
	rings [][]eclipsecore.SolarEclipsePathPoint,
	centerLine []eclipsecore.SolarEclipsePathPoint,
) ([]solarBandSideRun, bool) {
	var runs []solarBandSideRun
	for _, ring := range rings {
		points := openSolarPathRing(ring)
		count := len(points)
		if count < 3 {
			continue
		}
		projections, start, ok := solarBandTrackRingProjections(points, centerLine)
		if !ok {
			return nil, false
		}
		// 从起点之后第一个侧别翻转处开环，使首段不在环的接缝处被截断。

		origin := start
		for offset := 1; offset < count; offset++ {
			index := (start + offset) % count
			if projections[index].side != projections[(index+count-1)%count].side {
				origin = index
				break
			}
		}
		current := solarBandSideRun{}
		flush := func() {
			if len(current.points) >= 3 {
				runs = append(runs, current)
			}
			current = solarBandSideRun{}
		}
		for offset := 0; offset < count; offset++ {
			index := (origin + offset) % count
			next := (index + 1) % count
			projection := projections[index]
			current.points = append(current.points, points[index])
			current.progress = append(current.progress, projection.progress)
			current.times = append(current.times, solarBandCenterLineTime(centerLine, projection.progress))
			current.north = projection.side > 0
			seamJump := math.Abs(math.Remainder(points[next].Longitude-points[index].Longitude, 360)) > 180
			if projection.side != projections[next].side || seamJump {
				flush()
			}
		}
		flush()
	}
	return runs, true
}

// solarBandTrackRingProjections 为环上每个顶点在中心线的局部最近点中选一个，使"距离 + 沿线跳变超额"之和最小（Viterbi），回折与端点处不跳支。
func solarBandTrackRingProjections(
	points []eclipsecore.SolarEclipsePathPoint,
	centerLine []eclipsecore.SolarEclipsePathPoint,
) ([]solarBandSideProjection, int, bool) {
	count := len(points)
	if len(centerLine) < 2 {
		return nil, 0, false
	}
	middle := float64(len(centerLine)-1) / 2
	along := solarBandCenterLineAlongKM(centerLine)
	ringCandidates := make([][]solarBandTrackCandidate, count)
	start, startGap := 0, math.Inf(1)
	for index, point := range points {
		ringCandidates[index] = solarBandTrackCandidates(point, centerLine, along)
		// 全局最近点就是距离最小的局部极小。
		nearest := ringCandidates[index][0]
		for _, candidate := range ringCandidates[index][1:] {
			if candidate.distanceKM < nearest.distanceKM {
				nearest = candidate
			}
		}
		if gap := math.Abs(nearest.projection.progress - middle); gap < startGap {
			start, startGap = index, gap
		}
	}
	candidates := make([][]solarBandTrackCandidate, count)
	for offset := range candidates {
		candidates[offset] = ringCandidates[(start+offset)%count]
	}
	cost := make([][]float64, count)
	from := make([][]int, count)
	cost[0] = make([]float64, len(candidates[0]))
	for state, candidate := range candidates[0] {
		cost[0][state] = candidate.distanceKM
	}
	for offset := 1; offset < count; offset++ {
		step := solarCentralBandPointDistanceKM(points[(start+offset-1)%count], points[(start+offset)%count])
		cost[offset] = make([]float64, len(candidates[offset]))
		from[offset] = make([]int, len(candidates[offset]))
		for state, candidate := range candidates[offset] {
			best, bestFrom := math.Inf(1), 0
			for previousState, previous := range candidates[offset-1] {
				// 沿中心线的移动超过环上步长两倍的部分才计罚。
				jump := math.Max(0, math.Abs(candidate.alongKM-previous.alongKM)-2*step)
				if value := cost[offset-1][previousState] + jump; value < best {
					best, bestFrom = value, previousState
				}
			}
			cost[offset][state] = best + candidate.distanceKM
			from[offset][state] = bestFrom
		}
	}
	state := 0
	for candidate := range cost[count-1] {
		if cost[count-1][candidate] < cost[count-1][state] {
			state = candidate
		}
	}
	projections := make([]solarBandSideProjection, count)
	for offset := count - 1; offset >= 0; offset-- {
		projections[(start+offset)%count] = candidates[offset][state].projection
		if offset > 0 {
			state = from[offset][state]
		}
	}
	return projections, start, true
}

type solarBandTrackCandidate struct {
	projection solarBandSideProjection
	distanceKM float64
	alongKM    float64
}

// solarBandTrackCandidates 返回顶点到中心线距离沿线的全部局部极小（含平台与两端）。
func solarBandTrackCandidates(
	point eclipsecore.SolarEclipsePathPoint,
	centerLine []eclipsecore.SolarEclipsePathPoint,
	along []float64,
) []solarBandTrackCandidate {
	segments := len(centerLine) - 1
	projections := make([]solarBandSideProjection, segments)
	distances := make([]float64, segments)
	for position := range projections {
		projections[position], distances[position] = solarBandProjectOnSegment(point, centerLine, position)
	}
	var candidates []solarBandTrackCandidate
	for position := range projections {
		if position > 0 && distances[position-1] < distances[position] ||
			position+1 < segments && distances[position+1] < distances[position] {
			continue
		}
		// 平台只取首段，避免同一极小重复入选。
		if position > 0 && distances[position-1] == distances[position] {
			continue
		}
		progress := projections[position].progress
		segment := solarBandProgressSegment(progress, len(centerLine))
		fraction := progress - float64(segment)
		candidates = append(candidates, solarBandTrackCandidate{
			projection: projections[position],
			distanceKM: 111.32 * distances[position],
			alongKM:    along[segment] + fraction*(along[segment+1]-along[segment]),
		})
	}
	return candidates
}

// solarBandCenterLineAlongKM 是中心线各顶点的累计沿线距离。
func solarBandCenterLineAlongKM(centerLine []eclipsecore.SolarEclipsePathPoint) []float64 {
	along := make([]float64, len(centerLine))
	for index := 1; index < len(centerLine); index++ {
		along[index] = along[index-1] + solarCentralBandPointDistanceKM(centerLine[index-1], centerLine[index])
	}
	return along
}

func solarBandProgressSegment(progress float64, size int) int {
	segment := int(math.Floor(progress))
	if segment > size-2 {
		segment = size - 2
	}
	if segment < 0 {
		segment = 0
	}
	return segment
}

// solarBandCenterLineTime is the center-line time at one fractional position.
func solarBandCenterLineTime(centerLine []eclipsecore.SolarEclipsePathPoint, progress float64) time.Time {
	position := solarBandProgressSegment(progress, len(centerLine))
	fraction := math.Max(0, math.Min(1, progress-float64(position)))
	first, second := centerLine[position], centerLine[position+1]
	return first.Time.Add(time.Duration(float64(second.Time.Sub(first.Time)) * fraction))
}

// solarBandSidePoints concatenates the runs of one side in path order and
// spreads their times evenly, because the export requires strictly increasing
// times. Runs that do not touch each other belong to different boundary
// fragments (the union leaves small islands behind); concatenating them would
// draw a limit line straight across the map, so only the longest connected
// group is kept.
func solarBandSidePoints(runs []solarBandSideRun, north bool) []eclipsecore.SolarEclipsePathPoint {
	chosen := make([]solarBandSideRun, 0, len(runs))
	for _, run := range runs {
		if run.north == north {
			chosen = append(chosen, run)
		}
	}
	if len(chosen) == 0 {
		return nil
	}
	sort.SliceStable(chosen, func(first, second int) bool {
		return meanProgress(chosen[first].progress) < meanProgress(chosen[second].progress)
	})
	var groups [][]solarBandSideRun
	for _, run := range chosen {
		if len(groups) > 0 {
			last := groups[len(groups)-1]
			if solarBandRunsConnect(last[len(last)-1].points, run.points) {
				groups[len(groups)-1] = append(last, run)
				continue
			}
		}
		groups = append(groups, []solarBandSideRun{run})
	}
	countPoints := func(group []solarBandSideRun) int {
		total := 0
		for _, run := range group {
			total += len(run.points)
		}
		return total
	}
	best := groups[0]
	for _, group := range groups[1:] {
		if countPoints(group) > countPoints(best) {
			best = group
		}
	}
	side := make([]eclipsecore.SolarEclipsePathPoint, 0, countPoints(best))
	for _, run := range best {
		for index, point := range run.points {
			if index < len(run.times) {
				point.Time = run.times[index]
			}
			side = append(side, point)
		}
	}
	if len(side) < 3 {
		return nil
	}
	if !side[len(side)-1].Time.After(side[0].Time) {
		for left, right := 0, len(side)-1; left < right; left, right = left+1, right-1 {
			side[left], side[right] = side[right], side[left]
		}
	}
	return enforceSolarBandSideTimes(side)
}

// enforceSolarBandSideTimes 保留逐点投影时间，只把投影时间回退的顶点抬到前一点之后。
func enforceSolarBandSideTimes(side []eclipsecore.SolarEclipsePathPoint) []eclipsecore.SolarEclipsePathPoint {
	if len(side) < 2 || !side[len(side)-1].Time.After(side[0].Time) {
		return nil
	}
	for index := 1; index < len(side); index++ {
		if !side[index].Time.After(side[index-1].Time) {
			side[index].Time = side[index-1].Time.Add(time.Millisecond)
		}
	}
	return side
}

// solarBandRunsConnect reports whether two runs share an endpoint, wrapping
// longitudes so a seam crossing still counts as connected.
func solarBandRunsConnect(first, second []eclipsecore.SolarEclipsePathPoint) bool {
	if len(first) == 0 || len(second) == 0 {
		return false
	}
	scale := math.Cos(first[len(first)-1].Latitude * math.Pi / 180)
	deltaLongitude := math.Remainder(first[len(first)-1].Longitude-second[0].Longitude, 360) * scale
	deltaLatitude := first[len(first)-1].Latitude - second[0].Latitude
	return 111.32*math.Hypot(deltaLongitude, deltaLatitude) <= solarCentralBandRunConnectKM
}

// meanProgress averages the projected positions of one run.
func meanProgress(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

// solarBandSideProjection is the nearest center-line position of one band vertex.
type solarBandSideProjection struct {
	// progress is the segment index plus the fraction along that segment.
	progress float64
	// side is +1 when the vertex lies at or north of the center line there, -1 south.
	side int
}

// solarBandProjectOnSegment 返回顶点在中心线第 position 段上的投影与局部平面距离（度）。
func solarBandProjectOnSegment(
	point eclipsecore.SolarEclipsePathPoint,
	centerLine []eclipsecore.SolarEclipsePathPoint,
	position int,
) (solarBandSideProjection, float64) {
	scale := math.Cos(point.Latitude * math.Pi / 180)
	first, second := centerLine[position], centerLine[position+1]
	ax := math.Remainder(first.Longitude-point.Longitude, 360) * scale
	ay := first.Latitude - point.Latitude
	bx := math.Remainder(second.Longitude-point.Longitude, 360) * scale
	by := second.Latitude - point.Latitude
	dx, dy := bx-ax, by-ay
	length := dx*dx + dy*dy
	fraction := 0.0
	if length > 0 {
		fraction = math.Max(0, math.Min(1, -(ax*dx+ay*dy)/length))
	}
	offset := point.Latitude - (first.Latitude + fraction*(second.Latitude-first.Latitude))
	side := -1
	if offset >= 0 {
		side = 1
	}
	return solarBandSideProjection{
		progress: float64(position) + fraction,
		side:     side,
	}, math.Hypot(ax+fraction*dx, ay+fraction*dy)
}

// solarCentralLimitSeparationKM returns the greatest distance from one exported
// limit curve to the matching side of the band.
func solarCentralLimitSeparationKM(
	line []eclipsecore.SolarEclipsePathPoint,
	side []eclipsecore.SolarEclipsePathPoint,
) float64 {
	if len(line) < 2 || len(side) < 2 {
		return math.Inf(1)
	}
	maximum := 0.0
	for _, point := range line {
		best := math.Inf(1)
		for index := 0; index+1 < len(side); index++ {
			best = math.Min(best, solarCentralBandPointSegmentKM(point, side[index], side[index+1]))
		}
		maximum = math.Max(maximum, best)
	}
	return maximum
}

// solarCentralBandPointSegmentKM is the distance from a point to one great-circle
// segment, evaluated on a local equirectangular chart.
func solarCentralBandPointSegmentKM(
	point, first, second eclipsecore.SolarEclipsePathPoint,
) float64 {
	scale := math.Cos(point.Latitude * math.Pi / 180)
	// 经度差必须先归约到 ±180°：跨换日线的限线用裸差值会得到数万公里的假距离
	// （同一文件其它点-段投影都先做 math.Remainder）。
	// Longitude differences must be wrapped to ±180°: a limit line crossing the
	// antimeridian otherwise measures tens of thousands of kilometres away, while every
	// other point-to-segment projection in this file wraps first.
	ax := math.Remainder(first.Longitude-point.Longitude, 360) * scale
	ay := first.Latitude - point.Latitude
	bx := math.Remainder(second.Longitude-point.Longitude, 360) * scale
	by := second.Latitude - point.Latitude
	dx, dy := bx-ax, by-ay
	length := dx*dx + dy*dy
	fraction := 0.0
	if length > 0 {
		fraction = math.Max(0, math.Min(1, -(ax*dx+ay*dy)/length))
	}
	return 111.32 * math.Hypot(ax+fraction*dx, ay+fraction*dy)
}

// solarCentralBandSnapToleranceKM is how close an exported band vertex must be
// to a greatest-at-horizon curve before it is moved onto it. A grazing band is
// rebuilt from sampled footprints, so its horizon-bounded edge carries a few
// kilometres of sampling ripple; the curve itself is the exact boundary there,
// and the map draws both, so the ripple reads as two lines weaving instead of
// one boundary.
const solarCentralBandSnapToleranceKM = 25.0

// snapSolarBandGeometryToHorizonCurves replaces the band boundary runs that
// already follow a greatest-at-horizon curve with that curve's own vertices, so
// the filled band and the exported visibility line share one boundary. Runs are
// only replaced while their projection onto the curve stays monotone, which
// keeps the substitution from folding the ring; every other edge (the
// shadow-bounded parts) is left untouched.
func snapSolarBandGeometryToHorizonCurves(
	value geometry,
	curves []eclipsecore.SolarEclipseRiseSetCurve,
) geometry {
	polygons, ok := value.Coordinates.([][][][]float64)
	if !ok || len(polygons) == 0 {
		return value
	}
	paths := make([][]eclipsecore.SolarEclipsePathPoint, 0, 2)
	for _, curve := range curves {
		if curve.Phase != eclipsecore.RiseSetPhaseGreatest {
			continue
		}
		for _, segment := range curve.Segments {
			if len(segment) >= 2 {
				paths = append(paths, segment)
			}
		}
	}
	if len(paths) == 0 {
		return value
	}
	snapped := make([][][][]float64, len(polygons))
	for polygonIndex, polygon := range polygons {
		snapped[polygonIndex] = make([][][]float64, len(polygon))
		for ringIndex, ring := range polygon {
			snapped[polygonIndex][ringIndex] = snapSolarBandRingToHorizonPaths(ring, paths)
		}
	}
	return geometry{Type: value.Type, Coordinates: snapped}
}

// solarBandProjection is the closest point of one greatest-at-horizon path to a
// band vertex, with the parameter that locates it along that path.
type solarBandProjection struct {
	pathIndex int
	parameter float64
	longitude float64
	latitude  float64
	distance  float64
}

func solarBandProjectionAt(
	longitude, latitude float64,
	paths [][]eclipsecore.SolarEclipsePathPoint,
) (solarBandProjection, bool) {
	best := solarBandProjection{distance: solarCentralBandSnapToleranceKM}
	found := false
	scale := math.Cos(latitude * math.Pi / 180)
	for pathIndex, path := range paths {
		for index := 0; index+1 < len(path); index++ {
			first, second := path[index], path[index+1]
			ax := math.Remainder(first.Longitude-longitude, 360) * scale
			ay := first.Latitude - latitude
			bx := math.Remainder(second.Longitude-longitude, 360) * scale
			by := second.Latitude - latitude
			dx, dy := bx-ax, by-ay
			length := dx*dx + dy*dy
			fraction := 0.0
			if length > 0 {
				fraction = math.Max(0, math.Min(1, -(ax*dx+ay*dy)/length))
			}
			distance := 111.32 * math.Hypot(ax+fraction*dx, ay+fraction*dy)
			if distance >= best.distance {
				continue
			}
			candidate := longitude + (ax+fraction*dx)/scale
			if candidate < -180 || candidate > 180 {
				// A projection that leaves the export window would have to be
				// wrapped, which moves the vertex across the seam. Keep the
				// sampled position instead of rewriting the fragment topology.
				continue
			}
			best = solarBandProjection{
				pathIndex: pathIndex,
				parameter: float64(index) + fraction,
				longitude: candidate,
				latitude:  latitude + (ay + fraction*dy),
				distance:  distance,
			}
			found = true
		}
	}
	return best, found
}

// snapSolarBandRingToHorizonPaths substitutes the monotone runs of one ring.
func snapSolarBandRingToHorizonPaths(
	ring [][]float64,
	paths [][]eclipsecore.SolarEclipsePathPoint,
) [][]float64 {
	if len(ring) < 4 {
		return ring
	}
	type projected struct {
		point      []float64
		projection solarBandProjection
		matched    bool
	}
	points := make([]projected, len(ring))
	for index, point := range ring {
		points[index] = projected{point: point}
		if len(point) < 2 {
			continue
		}
		if projection, ok := solarBandProjectionAt(point[0], point[1], paths); ok {
			points[index] = projected{
				point:      []float64{projection.longitude, projection.latitude},
				projection: projection, matched: true,
			}
		}
	}
	result := make([][]float64, 0, len(ring))
	for index := 0; index < len(points); {
		if !points[index].matched {
			result = append(result, points[index].point)
			index++
			continue
		}
		end := index
		for end+1 < len(points) && points[end+1].matched &&
			points[end+1].projection.pathIndex == points[index].projection.pathIndex &&
			points[end+1].projection.parameter > points[end].projection.parameter {
			end++
		}
		if end == index {
			result = append(result, points[index].point)
			index++
			continue
		}
		path := paths[points[index].projection.pathIndex]
		startParameter := points[index].projection.parameter
		endParameter := points[end].projection.parameter
		result = append(result, []float64{points[index].projection.longitude, points[index].projection.latitude})
		for position := int(math.Ceil(startParameter)); position < len(path); position++ {
			if float64(position) <= startParameter {
				continue
			}
			if float64(position) >= endParameter {
				break
			}
			result = append(result, []float64{path[position].Longitude, path[position].Latitude})
		}
		result = append(result, []float64{points[end].projection.longitude, points[end].projection.latitude})
		index = end + 1
	}
	if len(result) > 1 {
		result[len(result)-1] = result[0]
	}
	if len(result) < 4 {
		return ring
	}
	return result
}

// solarEclipseBandGeometry returns the geometry of the exported central band.
func solarEclipseBandGeometry(features []feature) (geometry, bool) {
	for index := len(features) - 1; index >= 0; index-- {
		if features[index].Properties["role"] != "central-band" {
			continue
		}
		return features[index].Geometry, true
	}
	return geometry{}, false
}

// solarBandGeometryRings returns the outer rings of one exported band geometry
// as path points without times; the limit split re-times them from the center
// line, so any band construction can be split the same way.
func solarBandGeometryRings(value geometry) [][]eclipsecore.SolarEclipsePathPoint {
	polygons, ok := value.Coordinates.([][][][]float64)
	if !ok {
		return nil
	}
	rings := make([][]eclipsecore.SolarEclipsePathPoint, 0, len(polygons))
	for _, polygon := range polygons {
		if len(polygon) == 0 {
			continue
		}
		ring := make([]eclipsecore.SolarEclipsePathPoint, 0, len(polygon[0]))
		for _, position := range polygon[0] {
			if len(position) < 2 {
				continue
			}
			ring = append(ring, eclipsecore.SolarEclipsePathPoint{
				Longitude: position[0], Latitude: position[1],
			})
		}
		if len(ring) >= 3 {
			rings = append(rings, ring)
		}
	}
	return rings
}

func solarCentralBandEnvelopeCoversFootprints(
	segments [][]eclipsecore.SolarEclipsePathPoint,
	footprints []eclipsecore.SolarEclipsePartialFootprint,
) bool {
	if len(segments) == 0 || len(footprints) == 0 {
		return true
	}
	polygons := make([][]geodata.GeoPoint, 0, len(segments))
	for _, segment := range segments {
		polygon := make([]geodata.GeoPoint, len(segment))
		for index, point := range segment {
			polygon[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		polygons = append(polygons, polygon)
	}
	// Endpoint footprints are sampled independently from the analytic
	// envelope; small numerical gaps are expected. Only a macroscopic miss
	// indicates that the envelope selected the wrong polar branch.
	points := make([]geodata.GeoPoint, 0, len(footprints))
	for _, footprint := range footprints {
		for _, boundary := range footprint.Boundaries {
			if len(boundary) < 2 {
				continue
			}
			converted, ok := solarCentralBandCoveragePoints(boundary)
			if !ok {
				return false
			}
			points = append(points, converted...)
		}
	}
	if len(points) == 0 {
		return true
	}
	return solarCentralBandPointsCover(polygons, points, solarCentralBandCoverageToleranceKM)
}

func solarCentralBandEnvelopeCoversPath(
	segments [][]eclipsecore.SolarEclipsePathPoint,
	central *eclipsecore.SolarEclipsePath,
) bool {
	if len(segments) == 0 || central == nil {
		return false
	}
	// Away from the poles the gnomonic containment check is well conditioned;
	// retain the critical envelope there to avoid changing ordinary output.
	// Validate the same spherical path containment at every latitude. A
	// latitude-based bypass hid ordinary grazing endpoint errors in addition
	// to the polar cases it was originally meant to protect.
	polygons := make([][]geodata.GeoPoint, 0, len(segments))
	for _, segment := range segments {
		if len(openSolarPathRing(segment)) < 3 {
			return false
		}
		polygon := make([]geodata.GeoPoint, len(segment))
		for index, point := range segment {
			polygon[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		polygons = append(polygons, polygon)
	}
	if len(central.CenterLine) < 2 {
		return false
	}
	path := make([]geodata.GeoPoint, len(central.CenterLine))
	for index, point := range central.CenterLine {
		path[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
	}
	const maximumMissKM = 6.0
	miss := geodata.SphericalPolygonsPathMissDistanceKM(polygons, [][]geodata.GeoPoint{path}, false)
	if miss > maximumMissKM {
		// A narrow band can have a few-kilometre spherical edge sag at a
		// closure. Larger misses still select the physical fallback geometry.
		if miss > maximumMissKM {
			return false
		}
	}
	// The raw spherical ring can still lose a seam when converted to
	// RFC-7946 fragments at the antimeridian. Validate the same fragments
	// used by multiPolygonGeometry before accepting this envelope.
	fragments := make([][]geodata.GeoPoint, 0, len(polygons))
	for _, polygon := range polygons {
		fragments = append(fragments,
			geodata.PolygonFragments(sampleSphericalMapRing(polygon), geodata.ClipView{Projection: geodata.ProjectionEquirectangular})...,
		)
	}
	if len(fragments) == 0 || geodata.SphericalPolygonsPathMissDistanceKM(fragments, [][]geodata.GeoPoint{path}, false) > maximumMissKM {
		return false
	}
	// The projected fragments above are the same RFC-7946 pieces used for
	// export and already contain the path-containment check.  A separate
	// latitude-only seam heuristic rejects valid thin polar rings when the
	// axis and boundary cross the antimeridian at different local curvatures.
	return true
}

func solarCentralBandAntimeridianSeamMatchesPath(
	polygons [][]geodata.GeoPoint, path []geodata.GeoPoint,
) bool {
	const maximumSeamLatitudeGap = 5.0
	for index := 1; index < len(path); index++ {
		first, second := path[index-1], path[index]
		if math.Abs(first.Longitude-second.Longitude) <= 180 {
			continue
		}
		secondLongitude := second.Longitude
		if secondLongitude < first.Longitude {
			secondLongitude += 360
		}
		firstLongitude := first.Longitude
		if firstLongitude < second.Longitude {
			firstLongitude += 360
		}
		fraction := (180 - firstLongitude) / (secondLongitude - firstLongitude)
		if fraction < 0 || fraction > 1 {
			fraction = (-180 - firstLongitude) / (secondLongitude - firstLongitude)
		}
		seamLatitude := first.Latitude + fraction*(second.Latitude-first.Latitude)
		bestGap := math.Inf(1)
		for _, polygon := range polygons {
			for pointIndex := 1; pointIndex < len(polygon); pointIndex++ {
				firstPoint, secondPoint := polygon[pointIndex-1], polygon[pointIndex]
				if math.Abs(firstPoint.Longitude-secondPoint.Longitude) > 180 {
					secondPointLongitude := secondPoint.Longitude
					if secondPointLongitude < firstPoint.Longitude {
						secondPointLongitude += 360
					}
					firstPointLongitude := firstPoint.Longitude
					if firstPointLongitude < secondPoint.Longitude {
						firstPointLongitude += 360
					}
					fraction := (180 - firstPointLongitude) / (secondPointLongitude - firstPointLongitude)
					if fraction >= 0 && fraction <= 1 {
						candidate := firstPoint.Latitude + fraction*(secondPoint.Latitude-firstPoint.Latitude)
						bestGap = math.Min(bestGap, math.Abs(candidate-seamLatitude))
					}
				}
			}
		}
		if bestGap > maximumSeamLatitudeGap {
			return false
		}
	}
	return true
}

func solarCentralBandEnvelopeGeometry(
	segments [][]eclipsecore.SolarEclipsePathPoint,
) (geometry, string, error) {
	if len(segments) == 0 {
		return geometry{}, "", fmt.Errorf("central-band envelope is unavailable")
	}
	polygons := make([][]geodata.GeoPoint, 0, len(segments))
	for segmentIndex, segment := range segments {
		if len(openSolarPathRing(segment)) < 3 {
			return geometry{}, "", fmt.Errorf("central-band envelope segment %d has fewer than three points", segmentIndex)
		}
		polygon := make([]geodata.GeoPoint, len(segment))
		for pointIndex, point := range segment {
			if err := validateCoordinate(point.Longitude, point.Latitude); err != nil {
				return geometry{}, "", fmt.Errorf("central-band envelope segment %d point %d: %w", segmentIndex, pointIndex, err)
			}
			polygon[pointIndex] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		polygons = append(polygons, polygon)
	}
	merged := polygons
	if len(polygons) > 1 {
		var err error
		merged, err = geodata.UnionPolygons(polygons)
		if err != nil {
			// Hybrid envelopes can contain annular/total/annular components
			// that meet only at a zero-width transition. Their individual
			// spherical rings are valid, while forcing a planar union creates
			// an open seam at the transition. Preserve those physical components
			// as a MultiPolygon instead of rejecting the whole eclipse.
			merged = polygons
			value, geometryErr := multiPolygonGeometry(merged)
			if geometryErr != nil {
				return geometry{}, "", fmt.Errorf("central-band envelope union: %w", err)
			}
			return value, "besselian-critical-envelope-components", nil
		}
	}
	value, err := multiPolygonGeometry(merged)
	if err != nil {
		return geometry{}, "", err
	}
	return value, "besselian-critical-envelope", nil
}

func openSolarPathRing(points []eclipsecore.SolarEclipsePathPoint) []eclipsecore.SolarEclipsePathPoint {
	if len(points) > 1 && points[0].Longitude == points[len(points)-1].Longitude &&
		points[0].Latitude == points[len(points)-1].Latitude {
		return points[:len(points)-1]
	}
	return points
}

func sameSolarFootprintSlice(
	first, second []eclipsecore.SolarEclipsePartialFootprint,
) bool {
	if len(first) != len(second) {
		return false
	}
	if len(first) == 0 {
		return true
	}
	return &first[0] == &second[0]
}

func solarCentralBandFootprints(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
) []eclipsecore.SolarEclipsePartialFootprint {
	if len(partial.CentralShadowFootprints) > 0 &&
		(len(partial.CentralBandFootprints) == 0 ||
			partial.CentralShadowStep > 0 && partial.CentralShadowStep <= partial.CentralBandStep &&
				partial.BoundaryPoints >= solarEclipseCentralBandMinimumBoundaryPoints) {
		return partial.CentralShadowFootprints
	}
	return partial.CentralBandFootprints
}

// solarCentralBandFootprintsForPath 将开放端部足迹限制在中心轴位于地平线以上的时段。
// One-limit polar eclipses have U1/U4 contacts before/after that interval;
// sweeping those open footprints into the static band creates artificial flared ends.
func solarCentralBandFootprintsForPath(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	central *eclipsecore.SolarEclipsePath,
) []eclipsecore.SolarEclipsePartialFootprint {
	footprints := solarCentralBandFootprints(partial)
	if central == nil || partial.Eclipse.Centrality != eclipsecore.SolarEclipseCentralOneLimit ||
		len(central.CenterLine) < 2 {
		return footprints
	}
	start := central.CenterLine[0].Time
	end := central.CenterLine[len(central.CenterLine)-1].Time
	if start.IsZero() || !start.Before(end) {
		return footprints
	}
	filtered := make([]eclipsecore.SolarEclipsePartialFootprint, 0, len(footprints))
	for _, footprint := range footprints {
		if footprint.Time.IsZero() || footprint.Time.Before(start) || footprint.Time.After(end) {
			continue
		}
		filtered = append(filtered, footprint)
	}
	if len(filtered) >= 2 {
		return filtered
	}
	return footprints
}

// LunarEclipseOptions 月食 GeoJSON 导出选项；全部要素都被丢掉时导出返回错误。
// LunarEclipseOptions are the lunar-eclipse GeoJSON export options; the export fails when every
// Feature is dropped.
type LunarEclipseOptions struct {
	// TimeMarkers 非空时沿月下点轨迹追加时间标记 Point 要素，等价于 MarshalLunarEclipseWithTimeMarkers。
	// TimeMarkers adds time-marker Point Features along the sublunar track when non-nil.
	TimeMarkers *TimeMarkerOptions
	// SkipRoles 列出不写进输出的 role。写包络的 role 被跳过时不再为它成环，两块都跳过则连整段
	// 扫掠都不做——时间包络是本入口的主要开销，其余要素只占很小一部分。
	// SkipRoles lists roles to leave out. A skipped envelope role is not polygonized, and skipping
	// both skips the whole sweep, which dominates this entry point.
	SkipRoles []string
	// EnvelopeSweepSamples 是时间包络在 P1-P4 上的采样段数；0 或负值用默认 48，正值收敛到 [2, 192]。
	// 段数越少越快，包络边界处的采样误差量级见手册。
	// EnvelopeSweepSamples is the number of P1-P4 sampling steps for the time envelopes: 0 or a
	// negative value keeps the default of 48, positive values are clamped to [2, 192]. Fewer steps
	// are faster; the manual lists the sampling error at the envelope boundary.
	EnvelopeSweepSamples int
	// EnvelopeLongitudePoints 是时间包络的经度列数；0 或负值取 max(360, boundaryPoints)，正值收敛到 [12, 720]。
	// 列距就是区域边缘的固有误差量级，减小它同时变快、变粗；实际列数由包络要素的 longitude_points 给出，
	// 瞬时半球的 boundary_points 不受它影响。
	// EnvelopeLongitudePoints is the time-envelope longitude column count: 0 or a negative value uses
	// max(360, boundaryPoints), positive values are clamped to [12, 720]. The column spacing sets the
	// inherent edge error, so lowering it is faster and coarser; the effective count is reported as the
	// envelope's own longitude_points and the instantaneous hemispheres' boundary_points is unaffected.
	EnvelopeLongitudePoints int
}

// MarshalLunarEclipseWithOptions 编码月食，输出内容与采样精度由 options 选择 / encode a lunar eclipse with selectable content and sampling.
func (module *Generator) MarshalLunarEclipseWithOptions(
	info eclipsecore.LunarEclipseInfo,
	boundaryPoints int,
	options LunarEclipseOptions,
) ([]byte, error) {
	return marshalLunarEclipse(info, boundaryPoints, options, module)
}

// MarshalLunarEclipse 将月食 P1/P4 站心月心可见区、几何地平线以及 P1-P4 的时间包络编码为 GeoJSON，不含折射 / encode lunar eclipse visibility regions and contact envelopes as GeoJSON.
func (module *Generator) MarshalLunarEclipse(info eclipsecore.LunarEclipseInfo, boundaryPoints int) ([]byte, error) {
	return marshalLunarEclipse(info, boundaryPoints, LunarEclipseOptions{}, module)
}

// MarshalLunarEclipseWithTimeMarkers 编码月食，并沿半影开始到结束的月下点轨迹追加 Point 要素 / encode a lunar eclipse with time markers along the sublunar track.
func (module *Generator) MarshalLunarEclipseWithTimeMarkers(
	info eclipsecore.LunarEclipseInfo,
	boundaryPoints int,
	options TimeMarkerOptions,
) ([]byte, error) {
	return marshalLunarEclipse(info, boundaryPoints, LunarEclipseOptions{TimeMarkers: &options}, module)
}

type lunarLatitudeInterval struct {
	low  float64
	high float64
}

type lunarHorizonSeries struct {
	points     []geodata.GeoPoint
	longitudes []float64
}

const (
	lunarVisibilitySweepSamples    = 48
	lunarVisibilitySweepSamplesMin = 2
	lunarVisibilitySweepSamplesMax = 192
	lunarVisibilityLongitudeMin    = 360
	// 包络按 1° 经度分列，交点纬度只需比列距小两个数量级；用 GeoJSON 默认的 0.002° 会多插一倍以上顶点。
	lunarVisibilityHorizonToleranceDegrees = 0.02
)

func lunarVisibilityLongitudePoints(boundaryPoints int, options LunarEclipseOptions) int {
	if options.EnvelopeLongitudePoints > 0 {
		points := options.EnvelopeLongitudePoints
		if points < 12 {
			points = 12
		}
		if points > 720 {
			points = 720
		}
		return points
	}
	points := lunarVisibilityLongitudeMin
	if boundaryPoints > points {
		points = boundaryPoints
	}
	if points > 720 {
		points = 720
	}
	return points
}

func lunarVisibilitySamples(options LunarEclipseOptions) int {
	if options.EnvelopeSweepSamples <= 0 {
		return lunarVisibilitySweepSamples
	}
	samples := options.EnvelopeSweepSamples
	if samples < lunarVisibilitySweepSamplesMin {
		samples = lunarVisibilitySweepSamplesMin
	}
	if samples > lunarVisibilitySweepSamplesMax {
		samples = lunarVisibilitySweepSamplesMax
	}
	return samples
}

func skippedLunarRole(skip []string, role string) bool {
	for _, value := range skip {
		if value == role {
			return true
		}
	}
	return false
}

func lunarVisibilityEnvelopeProperties(info eclipsecore.LunarEclipseInfo, aggregation string, longitudePoints int) map[string]interface{} {
	return map[string]interface{}{
		"eclipse_type":     string(info.Type),
		"longitude_points": longitudePoints,
		"time_start":       formatTime(info.PenumbralStart),
		"time_end":         formatTime(info.PenumbralEnd),
		"aggregation":      aggregation,
	}
}

// lunarSubtractInterval 从一段纬度区间里扣掉另一段。
func lunarSubtractInterval(piece, second lunarLatitudeInterval) []lunarLatitudeInterval {
	if second.high <= piece.low || second.low >= piece.high {
		return []lunarLatitudeInterval{piece}
	}
	pieces := make([]lunarLatitudeInterval, 0, 2)
	if second.low > piece.low {
		pieces = append(pieces, lunarLatitudeInterval{low: piece.low, high: second.low})
	}
	if second.high < piece.high {
		pieces = append(pieces, lunarLatitudeInterval{low: second.high, high: piece.high})
	}
	return pieces
}

// lunarSubtractLatitudeIntervals 从 minuend 里扣掉 subtrahend 覆盖的纬度区间。
func lunarSubtractLatitudeIntervals(minuend, subtrahend []lunarLatitudeInterval) []lunarLatitudeInterval {
	if len(minuend) == 0 || len(subtrahend) == 0 {
		return minuend
	}
	result := make([]lunarLatitudeInterval, 0, len(minuend))
	for _, first := range minuend {
		pieces := []lunarLatitudeInterval{first}
		for _, second := range subtrahend {
			next := make([]lunarLatitudeInterval, 0, len(pieces)+1)
			for _, piece := range pieces {
				next = append(next, lunarSubtractInterval(piece, second)...)
			}
			pieces = next
		}
		result = append(result, pieces...)
	}
	return lunarUnionLatitudeIntervals(result...)
}

// lunarHorizonColumns 是某时刻地平线在每个经度列上的纬度交点，与那一刻的月面状态。
type lunarHorizonColumns struct {
	state basic.MoonState
	roots [][]float64
}

func lunarHorizonColumnsAt(at time.Time, longitudePoints int, longitudes []float64, modules ...*Generator) (lunarHorizonColumns, error) {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	state := basic.MoonStateAtWithInstance(basic.Date2JD(at.UTC()), module.inst)
	points := state.MoonHorizon(longitudePoints)
	if len(points) < 3 {
		return lunarHorizonColumns{}, fmt.Errorf("geojson: lunar visibility horizon is unavailable")
	}
	return lunarHorizonColumns{
		state: state,
		roots: lunarHorizonSeriesFor(lunarHorizonRefinedPoints(points, at, module)).rootsByLongitude(longitudes),
	}, nil
}

func lunarVisibilityLongitudes(boundaryPoints int, options LunarEclipseOptions) []float64 {
	longitudePoints := lunarVisibilityLongitudePoints(boundaryPoints, options)
	longitudes := make([]float64, longitudePoints+1)
	for index := range longitudes {
		longitudes[index] = -180 + 360*float64(index)/float64(longitudePoints)
	}
	return longitudes
}

// lunarBandColumn 取一列上"主端可见区依次扣掉若干可见区"后的纬度区间。
func lunarBandColumn(primary lunarHorizonColumns, longitude float64, index int, exclude ...[]lunarLatitudeInterval) []lunarLatitudeInterval {
	intervals := lunarVisibleLatitudeIntervals(primary.state, longitude, primary.roots[index])
	for _, other := range exclude {
		intervals = lunarSubtractLatitudeIntervals(intervals, other)
	}
	return intervals
}

// lunarVisibilityUnionColumns 逐经度取 [start, end] 上"月亮在地平上"可见区的并集：
// 只比区间端点会漏掉极区掠射时两刻之间短暂露出的窗口，必须按与包络同一档位采样后求并。
func lunarVisibilityUnionColumns(
	start, end time.Time, samples, longitudePoints int, longitudes []float64, modules ...*Generator) ([][]lunarLatitudeInterval, error) {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	columns := make([][]lunarLatitudeInterval, len(longitudes))
	if samples < 1 {
		samples = 1
	}
	for index := 0; index <= samples; index++ {
		at := start.Add(end.Sub(start) * time.Duration(index) / time.Duration(samples))
		instant, err := lunarHorizonColumnsAt(at, longitudePoints, longitudes, module)
		if err != nil {
			return nil, err
		}
		for longitudeIndex, longitude := range longitudes {
			intervals := lunarVisibleLatitudeIntervals(instant.state, longitude, instant.roots[longitudeIndex])
			columns[longitudeIndex] = lunarUnionLatitudeIntervals(append(columns[longitudeIndex], intervals...)...)
		}
	}
	return columns, nil
}

// lunarUmbralBandSamples 让本影区间的采样步长与整场包络一致：区间更短就按比例少采，下限 2 段。
func lunarUmbralBandSamples(info eclipsecore.LunarEclipseInfo, options LunarEclipseOptions) int {
	total := info.PenumbralEnd.Sub(info.PenumbralStart)
	umbral := info.PartialEnd.Sub(info.PartialStart)
	if total <= 0 || umbral <= 0 {
		return lunarVisibilitySamples(options)
	}
	samples := int(math.Round(float64(lunarVisibilitySamples(options)) * float64(umbral) / float64(total)))
	if samples < 2 {
		return 2
	}
	if limit := lunarVisibilitySamples(options); samples > limit {
		return limit
	}
	return samples
}

// lunarPenumbraBandGeometries 生成"仅见半影"的两条带：月落侧是食始在地平上、本影阶段整段在地平下，
// 月出侧是食终在地平上、本影阶段整段在地平下；各自再扣掉另一端的可见区（极区下中天会让两端同时可见）。
func lunarPenumbraBandGeometries(
	info eclipsecore.LunarEclipseInfo,
	boundaryPoints int,
	options LunarEclipseOptions,
	wantMoonset, wantMoonrise bool, modules ...*Generator) ([2]geometry, error) {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	var result [2]geometry
	longitudes := lunarVisibilityLongitudes(boundaryPoints, options)
	longitudePoints := len(longitudes) - 1
	columns := make([]lunarHorizonColumns, 2)
	for index, at := range []time.Time{info.PenumbralStart, info.PenumbralEnd} {
		value, err := lunarHorizonColumnsAt(at, longitudePoints, longitudes, module)
		if err != nil {
			return result, err
		}
		columns[index] = value
	}
	umbralColumns, err := lunarVisibilityUnionColumns(
		info.PartialStart, info.PartialEnd, lunarUmbralBandSamples(info, options), longitudePoints, longitudes, module,
	)
	if err != nil {
		return result, err
	}
	bandColumns := [2][][]lunarLatitudeInterval{
		make([][]lunarLatitudeInterval, len(longitudes)),
		make([][]lunarLatitudeInterval, len(longitudes)),
	}
	for longitudeIndex, longitude := range longitudes {
		p1 := lunarVisibleLatitudeIntervals(columns[0].state, longitude, columns[0].roots[longitudeIndex])
		p4 := lunarVisibleLatitudeIntervals(columns[1].state, longitude, columns[1].roots[longitudeIndex])
		if wantMoonset {
			bandColumns[0][longitudeIndex] = lunarBandColumn(columns[0], longitude, longitudeIndex, umbralColumns[longitudeIndex], p4)
		}
		if wantMoonrise {
			bandColumns[1][longitudeIndex] = lunarBandColumn(columns[1], longitude, longitudeIndex, umbralColumns[longitudeIndex], p1)
		}
	}
	for index, role := range []string{"penumbra-moonset", "penumbra-moonrise"} {
		if (index == 0 && !wantMoonset) || (index == 1 && !wantMoonrise) {
			continue
		}
		value, err := lunarVisibilityEnvelopeGeometry(longitudes, bandColumns[index])
		if err != nil {
			return result, fmt.Errorf("geojson: %s: %w", role, err)
		}
		result[index] = dropDegenerateMultiPolygonRings(value)
	}
	return result, nil
}

func lunarPenumbraBandProperties(info eclipsecore.LunarEclipseInfo, start, end time.Time, boundaryPoints int) map[string]interface{} {
	return map[string]interface{}{
		"eclipse_type": string(info.Type),
		"time_start":   formatTime(start),
		"time_end":     formatTime(end),
		"phase":        "penumbral-only",
	}
}

// lunarVisibilityEnvelopeGeometries 只对 wantUnion / wantIntersection 指定的包络成环；被跳过的
// 那一块连逐列并/交都不做。
func lunarVisibilityEnvelopeGeometries(
	info eclipsecore.LunarEclipseInfo,
	boundaryPoints int,
	options LunarEclipseOptions,
	wantUnion, wantIntersection bool, modules ...*Generator) (geometry, geometry, error) {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	start, end := info.PenumbralStart, info.PenumbralEnd
	if !start.Before(end) {
		return geometry{}, geometry{}, fmt.Errorf("geojson: lunar eclipse penumbral interval is invalid")
	}
	longitudePoints := lunarVisibilityLongitudePoints(boundaryPoints, options)
	sweepSamples := lunarVisibilitySamples(options)
	times := make([]time.Time, sweepSamples+1)
	for index := range times {
		times[index] = start.Add(end.Sub(start) * time.Duration(index) / time.Duration(sweepSamples))
	}
	longitudes := lunarVisibilityLongitudes(boundaryPoints, options)
	states := make([]basic.MoonState, len(times))
	roots := make([][][]float64, len(times))
	for index, at := range times {
		columns, err := lunarHorizonColumnsAt(at, longitudePoints, longitudes, module)
		if err != nil {
			return geometry{}, geometry{}, err
		}
		states[index], roots[index] = columns.state, columns.roots
	}
	unionColumns := make([][]lunarLatitudeInterval, len(longitudes))
	intersectionColumns := make([][]lunarLatitudeInterval, len(longitudes))
	for longitudeIndex, longitude := range longitudes {
		var union, intersection []lunarLatitudeInterval
		for timeIndex := range times {
			intervals := lunarVisibleLatitudeIntervals(
				states[timeIndex], longitude, roots[timeIndex][longitudeIndex],
			)
			if wantIntersection {
				if timeIndex == 0 {
					intersection = append(intersection, intervals...)
				} else {
					intersection = lunarIntersectLatitudeIntervals(intersection, intervals)
				}
			}
			if wantUnion {
				union = lunarUnionLatitudeIntervals(append(union, intervals...)...)
			}
		}
		if wantUnion {
			unionColumns[longitudeIndex] = union
		}
		if wantIntersection {
			intersectionColumns[longitudeIndex] = intersection
		}
	}
	var unionGeometry, intersectionGeometry geometry
	if wantUnion {
		value, err := lunarVisibilityEnvelopeGeometry(longitudes, unionColumns)
		if err != nil {
			return geometry{}, geometry{}, fmt.Errorf("geojson: visible-during-eclipse: %w", err)
		}
		unionGeometry = dropDegenerateMultiPolygonRings(value)
	}
	if wantIntersection {
		value, err := lunarVisibilityEnvelopeGeometry(longitudes, intersectionColumns)
		if err != nil {
			return geometry{}, geometry{}, fmt.Errorf("geojson: visible-throughout-eclipse: %w", err)
		}
		intersectionGeometry = dropDegenerateMultiPolygonRings(value)
	}
	return unionGeometry, intersectionGeometry, nil
}

func lunarHorizonRefinedPoints(points [][2]float64, at time.Time, modules ...*Generator) [][2]float64 {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	horizon := make([]geodata.GeoPoint, len(points))
	for index, point := range points {
		horizon[index] = geodata.GeoPoint{Longitude: point[0], Latitude: point[1]}
	}
	horizon = lunarhorizon.RefineWithinWithInstance(horizon, at, lunarVisibilityHorizonToleranceDegrees, module.inst)
	refined := make([][2]float64, len(horizon))
	for index, point := range horizon {
		refined[index] = [2]float64{point.Longitude, point.Latitude}
	}
	return refined
}

func lunarHorizonSeriesFor(points [][2]float64) lunarHorizonSeries {
	series := lunarHorizonSeries{
		points:     make([]geodata.GeoPoint, len(points)),
		longitudes: make([]float64, len(points)),
	}
	for index, point := range points {
		series.points[index] = geodata.GeoPoint{Longitude: point[0], Latitude: point[1]}
		longitude := point[0]
		if index > 0 {
			previous := series.longitudes[index-1]
			for longitude-previous > 180 {
				longitude -= 360
			}
			for longitude-previous < -180 {
				longitude += 360
			}
		}
		series.longitudes[index] = longitude
	}
	return series
}

// rootsByLongitude 一次遍历地平圈，给出每个经度列上的交点纬度。
func (series lunarHorizonSeries) rootsByLongitude(longitudes []float64) [][]float64 {
	roots := make([][]float64, len(longitudes))
	if len(series.points) < 3 || len(longitudes) < 2 {
		return roots
	}
	origin := longitudes[0]
	step := (longitudes[len(longitudes)-1] - origin) / float64(len(longitudes)-1)
	if !(step > 0) {
		return roots
	}
	for index := range series.points {
		next := (index + 1) % len(series.points)
		first, second := series.longitudes[index], series.longitudes[next]
		for second-first > 180 {
			second -= 360
		}
		for second-first < -180 {
			second += 360
		}
		span := second - first
		if math.Abs(span) < 1e-12 {
			continue
		}
		minimum, maximum := math.Min(first, second), math.Max(first, second)
		firstWorld := int(math.Floor((minimum - origin) / 360))
		lastWorld := int(math.Floor((maximum - origin) / 360))
		for world := firstWorld; world <= lastWorld; world++ {
			offset := 360 * float64(world)
			lowIndex := int(math.Ceil((minimum - offset - origin) / step))
			highIndex := int(math.Floor((maximum - offset - origin) / step))
			if lowIndex < 0 {
				lowIndex = 0
			}
			if highIndex >= len(longitudes) {
				highIndex = len(longitudes) - 1
			}
			for column := lowIndex; column <= highIndex; column++ {
				target := origin + step*float64(column) + offset
				fraction := (target - first) / span
				if fraction < 0 || fraction > 1 {
					continue
				}
				latitude := series.points[index].Latitude +
					(series.points[next].Latitude-series.points[index].Latitude)*fraction
				duplicate := false
				for _, root := range roots[column] {
					if math.Abs(root-latitude) < 1e-9 {
						duplicate = true
						break
					}
				}
				if !duplicate {
					roots[column] = append(roots[column], latitude)
				}
			}
		}
	}
	for column := range roots {
		sort.Float64s(roots[column])
	}
	return roots
}

func lunarVisibleLatitudeIntervals(state basic.MoonState, longitude float64, roots []float64) []lunarLatitudeInterval {
	boundaries := make([]float64, 0, len(roots)+2)
	boundaries = append(boundaries, -90)
	for _, root := range roots {
		if root > -90 && root < 90 {
			boundaries = append(boundaries, root)
		}
	}
	boundaries = append(boundaries, 90)
	intervals := make([]lunarLatitudeInterval, 0, len(boundaries)-1)
	for index := 0; index+1 < len(boundaries); index++ {
		low, high := boundaries[index], boundaries[index+1]
		if high-low <= 1e-9 {
			continue
		}
		if state.HMoonHeight(longitude, (low+high)/2) > 0 {
			intervals = append(intervals, lunarLatitudeInterval{low: low, high: high})
		}
	}
	return intervals
}

func lunarUnionLatitudeIntervals(intervals ...lunarLatitudeInterval) []lunarLatitudeInterval {
	if len(intervals) == 0 {
		return nil
	}
	sort.Slice(intervals, func(left, right int) bool { return intervals[left].low < intervals[right].low })
	merged := make([]lunarLatitudeInterval, 0, len(intervals))
	for _, interval := range intervals {
		if interval.high <= interval.low {
			continue
		}
		if len(merged) == 0 || interval.low > merged[len(merged)-1].high+1e-9 {
			merged = append(merged, interval)
			continue
		}
		if interval.high > merged[len(merged)-1].high {
			merged[len(merged)-1].high = interval.high
		}
	}
	return merged
}

func lunarIntersectLatitudeIntervals(left, right []lunarLatitudeInterval) []lunarLatitudeInterval {
	if len(left) == 0 || len(right) == 0 {
		return nil
	}
	result := make([]lunarLatitudeInterval, 0, len(left))
	for _, first := range left {
		for _, second := range right {
			low := math.Max(first.low, second.low)
			high := math.Min(first.high, second.high)
			if high > low {
				result = append(result, lunarLatitudeInterval{low: low, high: high})
			}
		}
	}
	return lunarUnionLatitudeIntervals(result...)
}

// lunarEnvelopeBand 是一条沿经度连续延伸的可见带，收口时下边界正向、上边界反向拼成环。
type lunarEnvelopeBand struct {
	low  []geodata.GeoPoint
	high []geodata.GeoPoint
	last lunarLatitudeInterval
}

// lunarVisibilityEnvelopeGeometry 按纬度重叠把每列的区间串成带，同一列的主带与极冠各走各的。
func lunarVisibilityEnvelopeGeometry(longitudes []float64, columns [][]lunarLatitudeInterval) (geometry, error) {
	polygons := make([][]geodata.GeoPoint, 0, 2)
	bands := make([]lunarEnvelopeBand, 0, 2)
	closeBand := func(band lunarEnvelopeBand) {
		if len(band.low) < 2 {
			return
		}
		ring := make([]geodata.GeoPoint, 0, len(band.low)+len(band.high))
		ring = append(ring, band.low...)
		for index := len(band.high) - 1; index >= 0; index-- {
			ring = append(ring, band.high[index])
		}
		if len(ring) >= 3 {
			polygons = append(polygons, ring)
		}
	}
	for columnIndex, longitude := range longitudes {
		column := columns[columnIndex]
		assigned := make([]int, len(bands))
		used := make([]bool, len(column))
		for index := range assigned {
			assigned[index] = -1
		}
		for {
			bestBand, bestInterval, bestOverlap := -1, -1, 0.0
			for bandIndex := range bands {
				if assigned[bandIndex] >= 0 {
					continue
				}
				for intervalIndex, interval := range column {
					if used[intervalIndex] {
						continue
					}
					overlap := math.Min(interval.high, bands[bandIndex].last.high) -
						math.Max(interval.low, bands[bandIndex].last.low)
					if overlap > bestOverlap {
						bestBand, bestInterval, bestOverlap = bandIndex, intervalIndex, overlap
					}
				}
			}
			if bestBand < 0 {
				break
			}
			assigned[bestBand] = bestInterval
			used[bestInterval] = true
			bands[bestBand].low = append(bands[bestBand].low, geodata.GeoPoint{Longitude: longitude, Latitude: column[bestInterval].low})
			bands[bestBand].high = append(bands[bestBand].high, geodata.GeoPoint{Longitude: longitude, Latitude: column[bestInterval].high})
			bands[bestBand].last = column[bestInterval]
		}
		alive := bands[:0]
		for bandIndex := range bands {
			if assigned[bandIndex] < 0 {
				closeBand(bands[bandIndex])
				continue
			}
			alive = append(alive, bands[bandIndex])
		}
		bands = alive
		for intervalIndex, interval := range column {
			if used[intervalIndex] {
				continue
			}
			bands = append(bands, lunarEnvelopeBand{
				low:  []geodata.GeoPoint{{Longitude: longitude, Latitude: interval.low}},
				high: []geodata.GeoPoint{{Longitude: longitude, Latitude: interval.high}},
				last: interval,
			})
		}
	}
	for _, band := range bands {
		closeBand(band)
	}
	if len(polygons) == 0 {
		return geometry{Type: "MultiPolygon", Coordinates: [][][][]float64{}}, nil
	}
	return multiPolygonGeometry(polygons)
}

func marshalLunarEclipse(
	info eclipsecore.LunarEclipseInfo,
	boundaryPoints int,
	options LunarEclipseOptions, modules ...*Generator) ([]byte, error) {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	markerOptions := options.TimeMarkers
	if markerOptions != nil {
		if err := validateTimeMarkerOptions(*markerOptions); err != nil {
			return nil, err
		}
	}
	if err := validateLunarEclipseInfo(info); err != nil {
		return nil, err
	}
	timeScale, scaleErr := timeScaleForMarkers(markerOptions)
	if scaleErr != nil {
		return nil, scaleErr
	}
	// 几何一律用民用时刻，只有写进属性的时刻换时标，否则把 UT1 读数当民用时刻会平移月下点。
	geometry := info
	if timeScale == basic.TimeScaleUT1 {
		info = eclipsecore.LunarEclipseInfoInUT1(info)
	}
	boundaryPoints = normalizeLunarBoundaryPoints(boundaryPoints)
	properties := map[string]interface{}{
		"eclipse_type":    string(info.Type),
		"boundary_points": boundaryPoints,
	}
	features := make([]feature, 0, 7)

	contacts := []struct {
		role         string
		horizonRole  string
		geometryTime time.Time
		labelTime    time.Time
	}{
		{role: "visible-at-p1", horizonRole: "p1-horizon", geometryTime: geometry.PenumbralStart, labelTime: info.PenumbralStart},
		{role: "visible-at-p4", horizonRole: "p4-horizon", geometryTime: geometry.PenumbralEnd, labelTime: info.PenumbralEnd},
	}
	for _, contact := range contacts {
		points := basic.MoonHorizonWithInstance(basic.Date2JD(contact.geometryTime.UTC()), boundaryPoints, module.inst)
		horizon := make([]geodata.GeoPoint, len(points))
		for index, point := range points {
			horizon[index] = geodata.GeoPoint{Longitude: point[0], Latitude: point[1]}
		}
		horizon = lunarhorizon.RefineWithInstance(horizon, contact.geometryTime, module.inst)
		value, err := multiPolygonGeometry([][]geodata.GeoPoint{horizon})
		if err != nil {
			return nil, fmt.Errorf("geojson: %s: %w", contact.role, err)
		}
		// 日界线剪裁会在相邻世界各输出一次零宽薄片：顶点全落在同一条子午线上、平面面积只剩
		// 浮点噪声（可能越过共享剪裁器 1e-12 的零面积阈值）。只在本月食
		// 可见区导出里丢弃它——共享剪裁器的输出被掩星拓扑依赖，不能在那里过滤。
		// The antimeridian split can emit one zero-width sliver per adjacent world: every vertex
		// on one meridian and a planar area of pure floating-point noise (possibly past the
		// shared splitter's 1e-12 zero-area floor). Drop it here, in the lunar
		// visibility export only; the shared splitter's output feeds occultation topology and must
		// not be filtered.
		value = dropDegenerateMultiPolygonRings(value)
		contactProperties := cloneProperties(properties)
		contactProperties["time"] = formatTime(contact.labelTime)
		features = append(features, newFeature(
			lunarEclipseEvent, contact.role, value, contactProperties,
		))

		horizonValue, err := geoMultiLineGeometry(horizon, true)
		if err != nil {
			return nil, fmt.Errorf("geojson: %s: %w", contact.horizonRole, err)
		}
		features = append(features, newFeature(
			lunarEclipseEvent,
			contact.horizonRole,
			horizonValue,
			map[string]interface{}{
				"eclipse_type": string(info.Type),
				"time":         formatTime(contact.labelTime),
			},
		))
	}

	wantPenumbraMoonset := !skippedLunarRole(options.SkipRoles, "penumbra-moonset")
	wantPenumbraMoonrise := !skippedLunarRole(options.SkipRoles, "penumbra-moonrise")
	if info.HasPartial && (wantPenumbraMoonset || wantPenumbraMoonrise) {
		bandGeometries, bandErr := lunarPenumbraBandGeometries(
			geometry, boundaryPoints, options, wantPenumbraMoonset, wantPenumbraMoonrise, module,
		)
		if bandErr != nil {
			return nil, bandErr
		}
		if wantPenumbraMoonset {
			features = append(features, newFeature(lunarEclipseEvent, "penumbra-moonset", bandGeometries[0],
				lunarPenumbraBandProperties(info, info.PenumbralStart, info.PartialStart, boundaryPoints)))
		}
		if wantPenumbraMoonrise {
			features = append(features, newFeature(lunarEclipseEvent, "penumbra-moonrise", bandGeometries[1],
				lunarPenumbraBandProperties(info, info.PartialEnd, info.PenumbralEnd, boundaryPoints)))
		}
	}

	wantDuring := !skippedLunarRole(options.SkipRoles, "visible-during-eclipse")
	wantThroughout := !skippedLunarRole(options.SkipRoles, "visible-throughout-eclipse")
	if wantDuring || wantThroughout {
		unionGeometry, intersectionGeometry, envelopeErr := lunarVisibilityEnvelopeGeometries(
			geometry, boundaryPoints, options, wantDuring, wantThroughout, module,
		)
		if envelopeErr != nil {
			return nil, envelopeErr
		}
		longitudePoints := lunarVisibilityLongitudePoints(boundaryPoints, options)
		if wantDuring {
			features = append(features, newFeature(lunarEclipseEvent, "visible-during-eclipse", unionGeometry,
				lunarVisibilityEnvelopeProperties(info, "union", longitudePoints)))
		}
		if wantThroughout {
			features = append(features, newFeature(lunarEclipseEvent, "visible-throughout-eclipse", intersectionGeometry,
				lunarVisibilityEnvelopeProperties(info, "intersection", longitudePoints)))
		}
	}

	maximum := lunarSubpoint(geometry.Maximum, module)
	features, err := appendPointFeature(
		features,
		lunarEclipseEvent,
		"greatest",
		pathSample{Time: info.Maximum, Longitude: maximum.Longitude, Latitude: maximum.Latitude},
		lunarEclipseMetadata(info),
	)
	if err != nil {
		return nil, err
	}
	if markerOptions != nil {
		markers, markerErr := lunarEclipseTimeMarkerSamples(geometry, timeScale, *markerOptions, module)
		if markerErr != nil {
			return nil, markerErr
		}
		features, err = appendTimeMarkerPointFeatures(
			features,
			lunarEclipseEvent,
			"sublunar-track",
			markers,
			markerOptions.Location,
		)
		if err != nil {
			return nil, err
		}
	}
	return marshalFeatureCollectionWithTimeScale(dropFeaturesByRole(features, options.SkipRoles), timeScale)
}

// horizonExact 为 true 时把开放边界补到地平圈擦地点（导出用）；为 false 时沿用旧封口，
// 因为掩带的面选择依赖这些填充提示，端点外扩会改变极区边缘的面归属。
func solarPartialFootprintPolygon(
	footprint eclipsecore.SolarEclipsePartialFootprint,
	horizonExact bool, modules ...*Generator) ([]geodata.GeoPoint, error) {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	if footprint.Time.IsZero() {
		return nil, fmt.Errorf("geojson: solar partial footprint time is required")
	}
	input, err := solarClosureFootprint(footprint, module)
	if err != nil {
		return nil, err
	}
	polygon, _, ok := solarclosure.Ring(input, horizonExact)
	if !ok {
		return nil, fmt.Errorf("geojson: solar partial footprint boundary is incomplete")
	}
	return polygon, nil
}

// solarPartialBandGeometry builds the authoritative static visibility region
// from the continuous zero-magnitude envelope and the horizon endpoint tracks.
// Instantaneous footprints remain available for selecting the current shadow,
// but are not part of this time-independent outline.
func solarPartialBandGeometry(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo, modules ...*Generator) (geometry, string, bool, error) {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	if len(partial.PartialBandContours) == 0 || len(partial.RiseSetCurves) == 0 {
		return solarPartialBandOverlayGeometry(partial.Footprints, partial.RiseSetCurves)
	}
	contours := make([][]geodata.GeoPoint, 0, len(partial.PartialBandContours))
	for _, contour := range partial.PartialBandContours {
		contours = append(contours, solarCentralBandGeoPoints(contour))
	}
	riseSetLines := make([][]geodata.GeoPoint, 0, len(partial.RiseSetCurves)*2)
	for _, curve := range partial.RiseSetCurves {
		for _, segment := range curve.Segments {
			riseSetLines = append(riseSetLines, solarCentralBandGeoPoints(segment))
		}
	}
	footprints := make([]solarclosure.Footprint, 0, len(partial.Footprints))
	for _, footprint := range partial.Footprints {
		input, err := solarClosureFootprint(footprint, module)
		if err != nil {
			return geometry{}, "", false, err
		}
		footprints = append(footprints, input)
	}
	// 面选择沿用采样端点的近似补口：端点外扩会改变极区边缘的面归属。
	polygons, ok := solarclosure.BandPolygons(
		contours, riseSetLines, footprints, false, solarclosure.SnapDistanceKM,
	)
	if !ok {
		return solarPartialBandOverlayGeometry(partial.Footprints, partial.RiseSetCurves)
	}
	source := "zero-magnitude-envelope+horizon-boundary"
	phaseLines := solarPartialBandPhaseLines(partial.RiseSetCurves)
	// The linework polygonizer selects faces using sampled instantaneous
	// footprints. In an extremely shallow non-central eclipse a phase branch
	// can lie in a neighbouring face that no sampled footprint reaches, even
	// though it belongs to the same visible envelope. Repair only that proven
	// containment miss; ordinary events keep the exact polygonizer result.
	if geodata.SphericalPolygonsPathMissDistanceKM(polygons, phaseLines, false) > 2 {
		if bridged, bridgedOK := solarPartialBandBridgePhaseLines(polygons, phaseLines); bridgedOK {
			polygons = bridged
			source += "+phase-bridge"
		}
	}
	for polygonIndex, polygon := range polygons {
		for pointIndex, point := range polygon {
			polygons[polygonIndex][pointIndex].Longitude = geodata.NormalizeLongitude(point.Longitude)
		}
	}
	value, err := multiPolygonGeometry(polygons)
	if err != nil {
		return geometry{}, "", false, err
	}
	return value, source, true, nil
}

// solarPartialBandOverlayGeometry fills the time-sampling gaps along the
// sunrise/sunset side when the physical line network cannot form a complete
// visibility boundary. The original footprint sequence remains part of the
// static fill for this explicitly non-authoritative fallback.
func solarPartialBandOverlayGeometry(
	footprints []eclipsecore.SolarEclipsePartialFootprint,
	curves []eclipsecore.SolarEclipseRiseSetCurve,
) (geometry, string, bool, error) {
	openCount := 0
	for _, footprint := range footprints {
		if !footprint.Closed && len(footprint.Boundaries) > 0 {
			openCount++
		}
	}
	if openCount < 2 {
		return geometry{}, "", false, nil
	}
	samples, err := solarCentralShadowSweepSamples(footprints)
	if err != nil {
		return geometry{}, "", false, err
	}
	polygons, err := geodata.OpenBoundaryEndpointOutlines(samples)
	if err != nil {
		return geometry{}, "", false, nil
	}
	source := "open-boundary-endpoint-outlines"
	if bridged, ok := solarPartialBandBridgePhaseLines(polygons, solarPartialBandPhaseLines(curves)); ok {
		polygons = bridged
		source += "+horizon-boundary"
	}
	for polygonIndex, polygon := range polygons {
		for pointIndex, point := range polygon {
			polygons[polygonIndex][pointIndex].Longitude = geodata.NormalizeLongitude(point.Longitude)
		}
	}
	value, err := multiPolygonGeometry(polygons)
	if err != nil {
		return geometry{}, "", false, err
	}
	return value, source, true, nil
}

func solarPartialBandPhaseLines(
	curves []eclipsecore.SolarEclipseRiseSetCurve,
) [][]geodata.GeoPoint {
	lines := make([][]geodata.GeoPoint, 0, len(curves)*2)
	for _, curve := range curves {
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

// solarPartialBandBridgePhaseLines closes the open-footprint fallback with
// the supplied rise/set tracks. Extremely shallow non-central eclipses may
// have no continuous zero-magnitude contour, while their phase tracks still
// extend beyond the few sampled open footprints. Connect each track to the
// nearest base-ring vertices, then union the local bridge faces. The union
// removes the artificial endpoint crossing and leaves every source phase line
// inside the returned visible envelope.
func solarPartialBandBridgePhaseLines(
	base, curves [][]geodata.GeoPoint,
) ([][]geodata.GeoPoint, bool) {
	if len(base) == 0 || len(curves) == 0 {
		return base, false
	}
	inputs := append([][]geodata.GeoPoint(nil), base...)
	bridged := false
	for _, line := range curves {
		if len(line) < 2 {
			continue
		}
		bestRing := -1
		bestDistance := math.Inf(1)
		for ringIndex, ring := range base {
			open := openRing(ring)
			if len(open) < 3 {
				continue
			}
			distance := solarPartialBandGeoPointDistanceToRing(line[0], open) +
				solarPartialBandGeoPointDistanceToRing(line[len(line)-1], open)
			if distance < bestDistance {
				bestRing, bestDistance = ringIndex, distance
			}
		}
		if bestRing < 0 || bestDistance > 4000 {
			return base, false
		}
		bridge := solarPartialBandBridgeLineToRing(line, openRing(base[bestRing]))
		if len(bridge) < 4 {
			continue
		}
		inputs = append(inputs, bridge)
		bridged = true
	}
	if !bridged {
		return base, false
	}
	merged, err := geodata.UnionPolygons(inputs)
	if err != nil || len(merged) == 0 ||
		!geodata.SphericalPolygonsContainPathsWithinKM(merged, curves, false, 2) {
		return base, false
	}
	return merged, true
}

func solarPartialBandGeoPointDistanceToRing(
	point geodata.GeoPoint,
	ring []geodata.GeoPoint,
) float64 {
	minimum := math.Inf(1)
	for index := range ring {
		minimum = math.Min(minimum, geodata.MeanDistanceKM(point, ring[index]))
	}
	return minimum
}

func solarPartialBandBridgeLineToRing(
	line, ring []geodata.GeoPoint,
) []geodata.GeoPoint {
	if len(line) < 2 || len(ring) < 3 {
		return nil
	}
	open := openRing(ring)
	if len(open) < 3 {
		return nil
	}
	nearest := func(point geodata.GeoPoint) (int, float64) {
		bestIndex := 0
		bestDistance := geodata.MeanDistanceKM(point, open[0])
		for index := 1; index < len(open); index++ {
			if distance := geodata.MeanDistanceKM(point, open[index]); distance < bestDistance {
				bestIndex, bestDistance = index, distance
			}
		}
		return bestIndex, bestDistance
	}
	startIndex, _ := nearest(line[0])
	endIndex, _ := nearest(line[len(line)-1])
	path := func(step int) []geodata.GeoPoint {
		result := []geodata.GeoPoint{open[endIndex]}
		index := endIndex
		for index != startIndex {
			index = (index + step + len(open)) % len(open)
			result = append(result, open[index])
		}
		return result
	}
	forward, reverse := path(1), path(-1)
	pathLength := func(points []geodata.GeoPoint) float64 {
		length := 0.0
		for index := 1; index < len(points); index++ {
			length += geodata.MeanDistanceKM(points[index-1], points[index])
		}
		return length
	}
	boundary := forward
	if pathLength(reverse) < pathLength(forward) {
		boundary = reverse
	}
	result := append([]geodata.GeoPoint(nil), line...)
	result = append(result, boundary...)
	return result
}

func appendSolarFootprintFeatures(
	features []feature,
	role string,
	footprints []eclipsecore.SolarEclipsePartialFootprint,
	properties map[string]interface{}, modules ...*Generator) ([]feature, error) {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	for _, footprint := range footprints {
		if role == solarCentralShadowFootprintRole && !footprint.Closed {
			appended, err := appendSolarHorizonClosedShadowFootprint(features, footprint, properties, module)
			if err != nil {
				return nil, err
			}
			features = appended
			continue
		}
		polygon, err := solarPartialFootprintPolygon(footprint, false, module)
		if err != nil {
			return nil, fmt.Errorf("geojson: solar %s at %s: %w", role, formatTime(footprint.Time), err)
		}
		curve, err := solarShadowFootprintCurveFromSegments(footprint.Boundaries)
		if err != nil {
			return nil, fmt.Errorf("geojson: solar %s at %s: %w", role, formatTime(footprint.Time), err)
		}
		if solarShadowRegionDegenerate(curve, polygon) {
			// 与单时刻导出同口径：退化区域整条缺省，不退化成点或零面积环。
			continue
		}
		footprintProperties := cloneProperties(properties)
		footprintProperties["time"] = formatTime(footprint.Time)
		footprintProperties["source_boundary_closed"] = footprint.Closed
		footprintProperties["interp_signature"] = solarShadowFootprintSignature(
			footprint.Boundaries, footprint.Closed, eclipsecore.SolarEclipseShadowUmbra,
		)
		if len(polygon) == 1 {
			value, pointErr := pointGeometry(polygon[0].Longitude, polygon[0].Latitude)
			if pointErr != nil {
				return nil, fmt.Errorf("geojson: solar %s at %s: %w", role, formatTime(footprint.Time), pointErr)
			}
			features = append(features, newFeature(solarEclipseEvent, role, value, footprintProperties))
			continue
		}
		value, geometryErr := multiPolygonGeometry([][]geodata.GeoPoint{polygon})
		if geometryErr != nil {
			return nil, fmt.Errorf("geojson: solar %s at %s: %w", role, formatTime(footprint.Time), geometryErr)
		}
		features = append(features, newFeature(solarEclipseEvent, role, value, footprintProperties))
	}
	return features, nil
}

// solarCentralShadowSweepGeometry closes the open antumbral arcs as one swept
// region. A non-central eclipse has no axis/earth intersection, so the normal
// paired central limits cannot describe this half-band. The first and last
// shadow arcs form the end caps; their two endpoint tracks form the sides.
func solarCentralShadowSweepGeometry(
	footprints []eclipsecore.SolarEclipsePartialFootprint,
) (geometry, error) {
	polygons, err := solarCentralShadowSweepPolygons(footprints)
	if err != nil {
		return geometry{}, err
	}
	return multiPolygonGeometry(polygons)
}

func solarCentralShadowSweepPolygons(
	footprints []eclipsecore.SolarEclipsePartialFootprint,
) ([][]geodata.GeoPoint, error) {
	samples, err := solarCentralShadowSweepSamples(footprints)
	if err != nil {
		return nil, err
	}
	polygons, err := geodata.OpenBoundarySweep(samples)
	if err != nil {
		return nil, fmt.Errorf("central-shadow footprints: %w", err)
	}
	return usableSolarCentralShadowSweepPolygons(polygons)
}

func solarCentralMonotoneEndSweepPolygons(
	footprints []eclipsecore.SolarEclipsePartialFootprint,
) ([][]geodata.GeoPoint, error) {
	samples, err := solarCentralShadowSweepSamples(footprints)
	if err != nil {
		return nil, err
	}
	polygons, err := geodata.MonotoneOpenBoundarySweep(samples)
	if err != nil {
		polygons, err = geodata.OpenBoundarySweep(
			geodata.DecimateOpenBoundarySweepSamples(samples, 24, 40),
		)
		if err != nil {
			return nil, fmt.Errorf("central-shadow endpoint footprints: %w", err)
		}
	}
	return usableSolarCentralShadowSweepPolygons(polygons)
}

func solarCentralShadowSweepSamples(
	footprints []eclipsecore.SolarEclipsePartialFootprint,
) ([]geodata.OpenBoundarySweepSample, error) {
	samples := make([]geodata.OpenBoundarySweepSample, 0, len(footprints))
	for _, footprint := range footprints {
		boundary := make([][]geodata.GeoPoint, 0, len(footprint.Boundaries))
		for _, source := range footprint.Boundaries {
			segment := make([]geodata.GeoPoint, len(source))
			for index, point := range source {
				if err := validateCoordinate(point.Longitude, point.Latitude); err != nil {
					return nil, err
				}
				segment[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			boundary = append(boundary, segment)
		}
		samples = append(samples, geodata.OpenBoundarySweepSample{
			Boundaries: boundary,
			Closed:     footprint.Closed,
		})
	}
	return samples, nil
}

func usableSolarCentralShadowSweepPolygons(
	polygons [][]geodata.GeoPoint,
) ([][]geodata.GeoPoint, error) {
	usable := make([][]geodata.GeoPoint, 0, len(polygons))
	for _, polygon := range polygons {
		if len(openRing(polygon)) >= 3 {
			usable = append(usable, polygon)
		}
	}
	if len(usable) == 0 {
		return nil, fmt.Errorf("central-shadow footprints contain no usable swept region")
	}
	if len(usable) == 1 {
		return usable, nil
	}
	return geodata.UnionPolygons(usable)
}

// solarCentralBandCoverageToleranceKM is the macro-leak threshold used to
// reject a central-band candidate that leaves real umbral area uncovered.
// Ordinary events stay within roughly 40 km at the 90th percentile, while a
// grazing one-limit ribbon or a failed two-limit ribbon misses by hundreds of
// kilometres.
const solarCentralBandCoverageToleranceKM = 100.0

// solarCentralBandCoveragePoints 把路径点转成球面点，坐标非法时返回 false。
func solarCentralBandCoveragePoints(
	points []eclipsecore.SolarEclipsePathPoint,
) ([]geodata.GeoPoint, bool) {
	result := make([]geodata.GeoPoint, len(points))
	for index, point := range points {
		if err := validateCoordinate(point.Longitude, point.Latitude); err != nil {
			return nil, false
		}
		result[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
	}
	return result, true
}

// solarCentralBandVertexGridDegrees 是顶点网格的边长：一格的纬度跨度已超过任何容差。
const solarCentralBandVertexGridDegrees = 1.0

// solarCentralBandVertexGrid 按固定网格索引环顶点，用于快速确认探针落在环附近。
type solarCentralBandVertexGrid map[int][]geodata.GeoPoint

func solarCentralBandVertexGridKey(latitude, longitude float64) int {
	latitudeCell := int(math.Floor(latitude/solarCentralBandVertexGridDegrees)) + 90
	longitudeCell := int(math.Floor(geodata.NormalizeLongitude(longitude)/solarCentralBandVertexGridDegrees)) + 180
	return latitudeCell*360 + longitudeCell
}

func newSolarCentralBandVertexGrid(polygons [][]geodata.GeoPoint) solarCentralBandVertexGrid {
	grid := make(solarCentralBandVertexGrid)
	for _, polygon := range polygons {
		for _, point := range openRing(polygon) {
			key := solarCentralBandVertexGridKey(point.Latitude, point.Longitude)
			grid[key] = append(grid[key], point)
		}
	}
	return grid
}

// vertexWithinKM 报告网格邻域内是否存在容差范围内的环顶点；网格给不出结论不代表真的超限。
func (grid solarCentralBandVertexGrid) vertexWithinKM(
	point geodata.GeoPoint,
	toleranceKM float64,
) bool {
	deltaLatitude := toleranceKM/111.0 + solarCentralBandVertexGridDegrees
	scale := math.Abs(math.Cos(point.Latitude * math.Pi / 180))
	if scale < 1e-6 {
		scale = 1e-6
	}
	deltaLongitude := deltaLatitude/scale + solarCentralBandVertexGridDegrees
	minimumLatitudeCell := int(math.Floor((point.Latitude-deltaLatitude)/solarCentralBandVertexGridDegrees)) + 90
	maximumLatitudeCell := int(math.Floor((point.Latitude+deltaLatitude)/solarCentralBandVertexGridDegrees)) + 90
	minimumLongitudeCell := int(math.Floor((point.Longitude-deltaLongitude)/solarCentralBandVertexGridDegrees)) + 180
	maximumLongitudeCell := int(math.Floor((point.Longitude+deltaLongitude)/solarCentralBandVertexGridDegrees)) + 180
	for latitudeCell := minimumLatitudeCell; latitudeCell <= maximumLatitudeCell; latitudeCell++ {
		for longitudeCell := minimumLongitudeCell; longitudeCell <= maximumLongitudeCell; longitudeCell++ {
			for _, vertex := range grid[latitudeCell*360+longitudeCell] {
				if geodata.MeanDistanceKM(point, vertex) <= toleranceKM {
					return true
				}
			}
		}
	}
	return false
}

// solarCentralBandPointsCover 报告每个点是否都在容差内落在多边形里：先做一次球面包含判定，
// 环外的点先用顶点网格确认附近有环顶点，只有网格给不出结论时才做精确的球面偏离计算。
func solarCentralBandPointsCover(
	polygons [][]geodata.GeoPoint,
	points []geodata.GeoPoint,
	toleranceKM float64,
) bool {
	if len(polygons) == 0 {
		return false
	}
	if len(points) == 0 {
		return true
	}
	grid := newSolarCentralBandVertexGrid(polygons)
	index := geodata.NewSphericalPolygonIndex(polygons)
	for position, inside := range index.ContainsPoints(points) {
		if inside {
			continue
		}
		// 环顶点到多边形的距离不小于到环顶点的距离，邻域内有顶点即已满足容差。
		if grid.vertexWithinKM(points[position], toleranceKM) {
			continue
		}
		if geodata.SphericalPolygonsPathMissDistanceKM(
			polygons, [][]geodata.GeoPoint{{points[position]}}, false,
		) > toleranceKM {
			return false
		}
	}
	return true
}

// solarCentralBandRingsCover reports whether the candidate rings contain the
// complete center line and the umbral sweep. The center line and the swept
// footprints are the ground truth the static band must describe, so a candidate
// that leaves either outside is not acceptable while a better alternative
// remains; every vertex is probed because this decision selects the exported
// candidate and a subsample can step over a real gap.
func solarCentralBandRingsCover(
	rings [][]geodata.GeoPoint,
	centerLine []eclipsecore.SolarEclipsePathPoint,
	footprints []eclipsecore.SolarEclipsePartialFootprint,
) bool {
	if len(rings) == 0 {
		return false
	}
	points := make([]geodata.GeoPoint, 0, len(centerLine))
	if len(centerLine) >= 2 {
		converted, ok := solarCentralBandCoveragePoints(centerLine)
		if !ok {
			return false
		}
		points = append(points, converted...)
	}
	for _, footprint := range footprints {
		for _, boundary := range footprint.Boundaries {
			if len(boundary) < 2 {
				continue
			}
			converted, ok := solarCentralBandCoveragePoints(boundary)
			if !ok {
				return false
			}
			points = append(points, converted...)
		}
	}
	if len(points) == 0 {
		return true
	}
	return solarCentralBandPointsCover(rings, points, solarCentralBandCoverageToleranceKM)
}

// The center line is the spine of the band: a candidate that clips it is not a
// valid envelope even when its outer edge follows the umbral sweep, and the
// union of a paired ribbon with end sweeps can re-orient a polar ring just
// enough to push a few line vertices outside. The corridor below widens such a
// candidate locally instead of rejecting the whole band and falling back to a
// ribbon that loses hundreds of kilometres of real umbral area.
const (
	solarCentralBandCenterlineToleranceKM = 2.0
	// 探针偏离超过该上限就不补走廊：半径会把食带撑成一个覆盖半个地球的圆盘。
	solarCentralBandCenterlineCorridorMaxKM = 200.0
)

// solarCentralBandCenterlineProbes 把采样中心线展开成顶点与边中点，作为包含判据的探针集合。
func solarCentralBandCenterlineProbes(
	centerLine []eclipsecore.SolarEclipsePathPoint,
) []geodata.GeoPoint {
	probes := make([]geodata.GeoPoint, 0, 2*len(centerLine))
	for index, point := range centerLine {
		if err := validateCoordinate(point.Longitude, point.Latitude); err != nil {
			return nil
		}
		probes = append(probes, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
		if index+1 >= len(centerLine) {
			continue
		}
		next := centerLine[index+1]
		if err := validateCoordinate(next.Longitude, next.Latitude); err != nil {
			return nil
		}
		probes = append(probes, geodata.GeoPoint{
			Longitude: geodata.NormalizeLongitude(
				point.Longitude + math.Remainder(next.Longitude-point.Longitude, 360)/2,
			),
			Latitude: (point.Latitude + next.Latitude) / 2,
		})
	}
	return probes
}

// solarCentralBandCenterlineMissesKM 逐个探针量到多边形的偏离，落在多边形内的探针为 0。
func solarCentralBandCenterlineMissesKM(
	polygons [][]geodata.GeoPoint,
	probes []geodata.GeoPoint,
) []float64 {
	misses := make([]float64, len(probes))
	if len(polygons) == 0 || len(probes) == 0 {
		return misses
	}
	index := geodata.NewSphericalPolygonIndex(polygons)
	contained := index.ContainsPoints(probes)
	for position, inside := range contained {
		if inside {
			continue
		}
		misses[position] = geodata.SphericalPolygonsPathMissDistanceKM(
			polygons, [][]geodata.GeoPoint{{probes[position]}}, false,
		)
	}
	return misses
}

// solarCentralBandWithCenterlineCorridor 保证每个中心线探针都在容差内落在返回的多边形里：
// 只给越界的探针补一个半径等于它自身偏离加容差的圆盘，偏离超过上限时原样返回。
func solarCentralBandWithCenterlineCorridor(
	polygons [][]geodata.GeoPoint,
	centerLine []eclipsecore.SolarEclipsePathPoint,
) [][]geodata.GeoPoint {
	if len(polygons) == 0 || len(centerLine) < 2 {
		return polygons
	}
	probes := solarCentralBandCenterlineProbes(centerLine)
	if len(probes) == 0 {
		return polygons
	}
	misses := solarCentralBandCenterlineMissesKM(polygons, probes)
	inputs := append([][]geodata.GeoPoint{}, polygons...)
	patched := 0
	for position, probe := range probes {
		miss := misses[position]
		if miss <= solarCentralBandCenterlineToleranceKM {
			continue
		}
		if miss > solarCentralBandCenterlineCorridorMaxKM {
			return polygons
		}
		circle := geodata.SphericalCircle(
			probe, (miss+solarCentralBandCenterlineToleranceKM)/111.32, 12,
		)
		if len(circle) < 3 {
			continue
		}
		inputs = append(inputs, append(circle, circle[0]))
		patched++
	}
	if patched == 0 {
		return polygons
	}
	merged, err := geodata.UnionPolygons(inputs)
	if err != nil || len(merged) == 0 {
		return polygons
	}
	return merged
}

// solarCentralMagnitudeOneBandGeometry builds the static totality envelope
// from the local-maximum magnitude-one contour. The old central limits are
// instantaneous cross-sections perpendicular to the moving shadow; near a
// low-altitude path those cross-sections can be narrower than the spatial
// envelope swept by the shadow. A magnitude-one contour is already that
// envelope. This fallback supports caller-supplied results without the core
// CentralBandSegments; normal calculations provide the closed band directly.
func solarCentralMagnitudeOneBandGeometry(
	eclipseType eclipsecore.SolarEclipseType,
	contours []eclipsecore.SolarEclipseMagnitudeContour,
	centerLine []eclipsecore.SolarEclipsePathPoint,
	horizonClosures [][]eclipsecore.SolarEclipsePathPoint,
) (geometry, string, bool, error) {
	// Only a total eclipse has a local magnitude-one contour: an annular eclipse
	// stays below one everywhere (the ring is the whole point), so its band is
	// bounded by the antumbral limits instead.
	if eclipseType != eclipsecore.SolarEclipseTotal || len(centerLine) < 2 {
		return geometry{}, "", false, nil
	}
	var segments [][]eclipsecore.SolarEclipsePathPoint
	for _, contour := range contours {
		if math.Abs(contour.Magnitude-1) > 1e-12 || len(contour.Segments) != 2 {
			continue
		}
		for _, segment := range contour.Segments {
			if len(segment) >= 2 {
				segments = append(segments, segment)
			}
		}
		if len(segments) == 2 {
			break
		}
		segments = nil
	}
	if len(segments) != 2 {
		return geometry{}, "", false, nil
	}
	first := append([]eclipsecore.SolarEclipsePathPoint(nil), segments[0]...)
	second := append([]eclipsecore.SolarEclipsePathPoint(nil), segments[1]...)
	ring, closedAtHorizon := solarCentralMagnitudeOneHorizonRing(first, second, horizonClosures)
	if !closedAtHorizon {
		forwardGap := solarCentralBandPathDistanceKM(first[len(first)-1], second[0])
		reverseGap := solarCentralBandPathDistanceKM(first[len(first)-1], second[len(second)-1])
		if reverseGap < forwardGap {
			for left, right := 0, len(second)-1; left < right; left, right = left+1, right-1 {
				second[left], second[right] = second[right], second[left]
			}
			forwardGap = reverseGap
		}
		closingGap := solarCentralBandPathDistanceKM(second[len(second)-1], first[0])
		if forwardGap > 2000 || closingGap > 2000 {
			return geometry{}, "", false, nil
		}
		ring = make([]geodata.GeoPoint, 0, len(first)+len(second))
		for _, point := range first {
			ring = append(ring, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
		}
		for _, point := range second {
			ring = append(ring, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
		}
	}
	ring = solarCentralBandRefineRingSpacing(ring, 200)
	centerPath := make([]geodata.GeoPoint, len(centerLine))
	for index, point := range centerLine {
		centerPath[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
	}
	if !geodata.SphericalPolygonsContainPathsWithinKM(
		[][]geodata.GeoPoint{ring}, [][]geodata.GeoPoint{centerPath}, false, 5,
	) {
		return geometry{}, "", false, nil
	}
	value, err := multiPolygonGeometry([][]geodata.GeoPoint{ring})
	if err != nil {
		return geometry{}, "", false, err
	}
	return value, "magnitude-one-envelope", true, nil
}

// solarCentralMagnitudeOneHorizonRing replaces both straight endpoint chords
// with the exact greatest-at-horizon arcs shared by the public rise/set lines.
func solarCentralMagnitudeOneHorizonRing(
	first, second []eclipsecore.SolarEclipsePathPoint,
	closures [][]eclipsecore.SolarEclipsePathPoint,
) ([]geodata.GeoPoint, bool) {
	if len(first) < 2 || len(second) < 2 || len(closures) != 2 ||
		len(closures[0]) < 2 || len(closures[1]) < 2 {
		return nil, false
	}
	type candidate struct {
		first  []eclipsecore.SolarEclipsePathPoint
		second []eclipsecore.SolarEclipsePathPoint
		score  float64
	}
	best := candidate{score: math.Inf(1)}
	for _, reverseFirst := range []bool{false, true} {
		for _, reverseSecond := range []bool{false, true} {
			firstCandidate := solarCentralBandOrientedPath(first, reverseFirst)
			secondCandidate := solarCentralBandOrientedPath(second, reverseSecond)
			score := solarCentralMagnitudeOneClosurePairDistance(
				closures[0], firstCandidate[0], secondCandidate[0],
			) + solarCentralMagnitudeOneClosurePairDistance(
				closures[1], firstCandidate[len(firstCandidate)-1], secondCandidate[len(secondCandidate)-1],
			)
			if score < best.score {
				best = candidate{first: firstCandidate, second: secondCandidate, score: score}
			}
		}
	}
	startClosure, startOK := solarCentralMagnitudeOneOrientedClosure(
		closures[0], best.second[0], best.first[0],
	)
	endClosure, endOK := solarCentralMagnitudeOneOrientedClosure(
		closures[1], best.first[len(best.first)-1], best.second[len(best.second)-1],
	)
	if !startOK || !endOK {
		return nil, false
	}
	best.first[0] = startClosure[len(startClosure)-1]
	best.first[len(best.first)-1] = endClosure[0]
	best.second[0] = startClosure[0]
	best.second[len(best.second)-1] = endClosure[len(endClosure)-1]

	points := make([]eclipsecore.SolarEclipsePathPoint, 0,
		len(best.first)+len(best.second)+len(startClosure)+len(endClosure),
	)
	points = append(points, best.first...)
	points = append(points, endClosure[1:]...)
	for index := len(best.second) - 2; index >= 0; index-- {
		points = append(points, best.second[index])
	}
	points = append(points, startClosure[1:]...)
	return solarCentralBandGeoPoints(points), true
}

func solarCentralBandOrientedPath(
	points []eclipsecore.SolarEclipsePathPoint,
	reverse bool,
) []eclipsecore.SolarEclipsePathPoint {
	result := append([]eclipsecore.SolarEclipsePathPoint(nil), points...)
	if reverse {
		for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
			result[left], result[right] = result[right], result[left]
		}
	}
	return result
}

func solarCentralMagnitudeOneClosurePairDistance(
	closure []eclipsecore.SolarEclipsePathPoint,
	first, second eclipsecore.SolarEclipsePathPoint,
) float64 {
	direct := solarCentralBandPathDistanceKM(closure[0], first) +
		solarCentralBandPathDistanceKM(closure[len(closure)-1], second)
	reverse := solarCentralBandPathDistanceKM(closure[len(closure)-1], first) +
		solarCentralBandPathDistanceKM(closure[0], second)
	return math.Min(direct, reverse)
}

func solarCentralMagnitudeOneOrientedClosure(
	source []eclipsecore.SolarEclipsePathPoint,
	start, end eclipsecore.SolarEclipsePathPoint,
) ([]eclipsecore.SolarEclipsePathPoint, bool) {
	const maximumRootDistanceKM = 1.0
	closure := solarCentralBandOrientedPath(source, false)
	direct := solarCentralBandPathDistanceKM(start, closure[0]) +
		solarCentralBandPathDistanceKM(end, closure[len(closure)-1])
	reverse := solarCentralBandPathDistanceKM(start, closure[len(closure)-1]) +
		solarCentralBandPathDistanceKM(end, closure[0])
	if reverse < direct {
		closure = solarCentralBandOrientedPath(closure, true)
	}
	if solarCentralBandPathDistanceKM(start, closure[0]) > maximumRootDistanceKM ||
		solarCentralBandPathDistanceKM(end, closure[len(closure)-1]) > maximumRootDistanceKM {
		return nil, false
	}
	return closure, true
}

func solarCentralBandPathDistanceKM(
	first, second eclipsecore.SolarEclipsePathPoint,
) float64 {
	return geodata.MeanDistanceKM(
		geodata.GeoPoint{Longitude: first.Longitude, Latitude: first.Latitude},
		geodata.GeoPoint{Longitude: second.Longitude, Latitude: second.Latitude},
	)
}

func solarCentralBandGeometry(
	northern, southern []eclipsecore.SolarEclipsePathPoint,
	centerLine []eclipsecore.SolarEclipsePathPoint,
	eclipseType eclipsecore.SolarEclipseType,
	centrality eclipsecore.SolarEclipseCentrality,
	footprints []eclipsecore.SolarEclipsePartialFootprint,
	horizonClosures [][]eclipsecore.SolarEclipsePathPoint,
) (geometry, string, error) {
	if eclipseType != eclipsecore.SolarEclipseHybrid {
		// A one-limit central event traces its paired limits only over the
		// shadow-axis interval, which for a grazing event (|gamma| ~ 0.98-0.997)
		// is a fraction of the U1..U4 umbral window. Reuse the two-limit
		// candidate chain, whose end-sweep and complete-contact alternatives are
		// coverage-validated, before falling back to the continuous ribbon.
		if centrality == eclipsecore.SolarEclipseCentralOneLimit {
			// The chain's last-resort candidates are returned without coverage
			// validation, so re-check here: a validated alternative is only worth
			// taking when it actually covers the umbral sweep, otherwise the
			// continuous ribbon below stays the better rendering.
			if polygons, source, ok := solarCentralTwoLimitBandPolygons(
				northern, southern, centerLine, footprints, horizonClosures,
			); ok && solarCentralBandRingsCover(polygons, centerLine, footprints) {
				value, geometryErr := multiPolygonGeometry(
					solarCentralBandWithCenterlineCorridor(polygons, centerLine),
				)
				if geometryErr != nil {
					return geometry{}, "", geometryErr
				}
				return value, source, nil
			}
			band, err := pairedLimitPolygon(northern, southern)
			if err != nil {
				return geometry{}, "", err
			}
			inputs := append([][]geodata.GeoPoint{band}, solarCentralBandEndpointCaps(northern, southern, centerLine)...)
			if merged, mergeErr := geodata.UnionPolygons(inputs); mergeErr == nil {
				inputs = merged
			}
			value, geometryErr := multiPolygonGeometry(inputs)
			if geometryErr != nil {
				return geometry{}, "", geometryErr
			}
			return value, "paired-limits-one-limit", nil
		}
		if centrality == eclipsecore.SolarEclipseCentralTwoLimits {
			polygons, source, ok := solarCentralTwoLimitBandPolygons(
				northern, southern, centerLine, footprints, horizonClosures,
			)
			if ok {
				value, geometryErr := multiPolygonGeometry(
					solarCentralBandWithCenterlineCorridor(polygons, centerLine),
				)
				if geometryErr != nil {
					return geometry{}, "", geometryErr
				}
				return value, source, nil
			}
		}
		band, err := pairedLimitPolygon(northern, southern)
		if err != nil {
			return geometry{}, "", err
		}
		endpointCaps := solarCentralBandEndpointCaps(northern, southern, centerLine)
		inputs := append([][]geodata.GeoPoint{band}, endpointCaps...)
		if len(footprints) > 0 {
			if sweep, sweepErr := solarCentralShadowSweepPolygons(footprints); sweepErr == nil {
				inputs = append(inputs, sweep...)
			}
		}
		if merged, mergeErr := geodata.UnionPolygons(inputs); mergeErr == nil {
			if value, geometryErr := multiPolygonGeometry(
				solarCentralBandWithCenterlineCorridor(merged, centerLine),
			); geometryErr == nil {
				return value, "paired-limits+central-shadow-sweep-union", nil
			}
		}
		value, geometryErr := multiPolygonGeometry(
			solarCentralBandWithCenterlineCorridor(inputs, centerLine),
		)
		if geometryErr != nil {
			return geometry{}, "", geometryErr
		}
		return value, "paired-limits-fallback", nil
	}
	bandPolygons, source := solarCentralPathBandPolygons(northern, southern)
	if len(bandPolygons) == 0 {
		band, err := pairedLimitPolygon(northern, southern)
		if err != nil {
			return geometry{}, "", err
		}
		bandPolygons = [][]geodata.GeoPoint{band}
		source = "paired-limits-fallback"
	}
	endpointCaps := solarCentralBandEndpointCaps(northern, southern, centerLine)
	if len(horizonClosures) == 2 {
		start, startOK := solarCentralBandHorizonTail(northern[0], southern[0], horizonClosures[0])
		end, endOK := solarCentralBandHorizonTail(northern[len(northern)-1], southern[len(southern)-1], horizonClosures[1])
		if startOK && endOK {
			endpointCaps = [][]geodata.GeoPoint{start, end}
			source += "+horizon-closures"
		}
	}
	// Keep the hybrid transition sweep, but close its ends at the solved
	// horizon limits rather than collapsing a finite-width shadow to the axis.
	polygons := append(append([][]geodata.GeoPoint(nil), bandPolygons...), endpointCaps...)
	if len(endpointCaps) > 0 {
		if merged, mergeErr := geodata.UnionPolygons(polygons); mergeErr == nil {
			polygons = merged
		}
	}
	paired, geometryErr := multiPolygonGeometry(polygons)
	if geometryErr != nil {
		return geometry{}, "", geometryErr
	}
	return paired, source, nil
}

func solarCentralTwoLimitBandPolygons(
	northern, southern, centerLine []eclipsecore.SolarEclipsePathPoint,
	footprints []eclipsecore.SolarEclipsePartialFootprint,
	horizonClosures [][]eclipsecore.SolarEclipsePathPoint,
) ([][]geodata.GeoPoint, string, bool) {
	// The map linework is deliberately trimmed to the center-line interval,
	// while this geometry must retain the complete U1/U4 contact interval.
	// MarshalSolarEclipse owns that presentation trim and always supplies the
	// complete paired limits here.
	north, south, ok := solarCentralTwoLimitPairedSamples(northern, southern)
	if !ok {
		return nil, "", false
	}
	north, south = solarCentralTwoLimitEnvelopeSamples(north, south)
	middle, err := pairedLimitPolygon(north, south)
	if err != nil {
		return nil, "", false
	}
	if len(horizonClosures) == 2 {
		startTail, startOK := solarCentralBandHorizonTail(north[0], south[0], horizonClosures[0])
		endTail, endOK := solarCentralBandHorizonTail(
			north[len(north)-1], south[len(south)-1], horizonClosures[1],
		)
		if startOK && endOK {
			inputs := [][]geodata.GeoPoint{middle, startTail, endTail}
			if merged, mergeErr := geodata.UnionPolygons(inputs); mergeErr == nil && len(merged) == 1 &&
				solarCentralBandRingsCover(merged, centerLine, footprints) {
				return merged, "paired-limits+horizon-closures", true
			}
		}
	}
	// Prefer the paired-limit ribbon that is explicitly validated against the
	// complete center line. End-sweep overlays can select a neighboring polar
	// face and create an artificial narrow neck even when the ribbon itself is
	// continuous.
	// The ribbon is traced over the center-line interval only. Accept it when
	// it really covers the complete umbral sweep; otherwise keep looking, since
	// grazing events lose hundreds of kilometres of genuine umbral area here.
	if merged, ok := solarCentralTwoLimitRibbonUnionPolygons(
		north, south, centerLine, nil,
	); ok && solarCentralBandRingsCover(merged, centerLine, footprints) {
		return merged, "paired-limits-ribbon-union", true
	}
	if len(footprints) > 0 {
		endSweeps, sweepErr := solarCentralMonotoneEndSweepPolygons(footprints)
		if sweepErr == nil {
			inputs := make([][]geodata.GeoPoint, 0, 1+len(endSweeps)+2)
			inputs = append(inputs, middle)
			inputs = append(inputs, endSweeps...)
			inputs = append(inputs, solarCentralBandInnerTransitionCaps(footprints)...)
			inputs = append(inputs, solarCentralBandContactCaps(
				footprints, northern[0], northern[len(northern)-1],
			)...)
			if merged, mergeErr := geodata.UnionPolygons(inputs); mergeErr == nil && len(merged) == 1 &&
				solarCentralBandRingsCover(merged, centerLine, footprints) {
				return merged, "paired-limits+central-shadow-end-sweeps", true
			}
			if merged, ok := solarCentralTwoLimitRibbonUnionPolygons(
				north, south, centerLine, inputs[1:],
			); ok && solarCentralBandRingsCover(merged, centerLine, footprints) {
				return merged, "paired-limits+central-shadow-ribbon-union", true
			}
		}
	}
	// The presentation limits above intentionally trim the two U1/U4 tails
	// for ordinary maps. Near a pole those trimmed ribbon pieces can fold into
	// hundreds of tiny triangles and lose the physical contact endpoints. Keep
	// the complete paired limits as one spherical ring before using the final
	// axis-cap fallback; multiPolygonGeometry performs the map split afterwards.
	if fullBand, fullErr := pairedLimitPolygon(north, south); fullErr == nil {
		if len(footprints) > 0 {
			if sweep, sweepErr := solarCentralShadowSweepPolygons(footprints); sweepErr == nil &&
				solarCentralBandRingsCover(sweep, centerLine, footprints) {
				contacts := []geodata.GeoPoint{
					{Longitude: northern[0].Longitude, Latitude: northern[0].Latitude},
					{Longitude: northern[len(northern)-1].Longitude, Latitude: northern[len(northern)-1].Latitude},
				}
				if snapSolarCentralSweepContacts(sweep, contacts) {
					return sweep, "central-shadow-complete-contact-fallback", true
				}
				return sweep, "central-shadow-complete-fallback", true
			}
		}
		return [][]geodata.GeoPoint{fullBand}, "paired-limits-complete-fallback", true
	}
	if fullBand, fullErr := pairedLimitPolygon(north, south); fullErr == nil {
		if pieces := solarCentralTwoLimitRibbonPieces(north, south, centerLine); len(pieces) > 0 {
			return pieces, "paired-limits-ribbon-pieces-fallback", true
		}
		return [][]geodata.GeoPoint{fullBand}, "paired-limits-full-limit-fallback", true
	}
	band, ok := solarCentralTwoLimitAxisCappedPolygon(north, south, centerLine)
	if !ok {
		return nil, "", false
	}
	return [][]geodata.GeoPoint{band}, "paired-limits-axis-cap-fallback", true
}

func solarCentralTwoLimitPairedSamples(
	northern, southern []eclipsecore.SolarEclipsePathPoint,
) ([]eclipsecore.SolarEclipsePathPoint, []eclipsecore.SolarEclipsePathPoint, bool) {
	if len(northern) < 2 || len(northern) != len(southern) {
		return nil, nil, false
	}
	for index := range northern {
		if northern[index].Time.IsZero() || southern[index].Time.IsZero() ||
			!northern[index].Time.Equal(southern[index].Time) {
			return nil, nil, false
		}
		if index > 0 && (!northern[index-1].Time.Before(northern[index].Time) ||
			!southern[index-1].Time.Before(southern[index].Time)) {
			return nil, nil, false
		}
	}
	return northern, southern, true
}

// snapSolarCentralSweepContacts moves the nearest sampled sweep vertices onto
// the exact U1/U4 contact points. The shadow footprints start at the contact
// times, but their finite angular/time sampling can leave the exported vertex
// a few kilometres away. Snapping the existing vertices preserves the sweep
// components and avoids adding overlapping endpoint triangles.
func snapSolarCentralSweepContacts(polygons [][]geodata.GeoPoint, contacts []geodata.GeoPoint) bool {
	if len(polygons) == 0 || len(contacts) == 0 {
		return false
	}
	const maximumSnapDistanceKM = 500.0
	used := make(map[[2]int]bool)
	for _, contact := range contacts {
		bestDistance := math.Inf(1)
		bestPolygon, bestPoint := -1, -1
		for polygonIndex, polygon := range polygons {
			for pointIndex, point := range polygon {
				if used[[2]int{polygonIndex, pointIndex}] {
					continue
				}
				distance := geodata.MeanDistanceKM(point, contact)
				if distance < bestDistance {
					bestDistance = distance
					bestPolygon, bestPoint = polygonIndex, pointIndex
				}
			}
		}
		if bestPolygon < 0 || bestDistance > maximumSnapDistanceKM {
			return false
		}
		polygons[bestPolygon][bestPoint] = contact
		used[[2]int{bestPolygon, bestPoint}] = true
	}
	return true
}

// solarCentralTwoLimitRibbonPieces keeps the fallback as a collection of
// adjacent time-slice faces instead of one polar ring. Each slice is split
// around the interpolated center-line segment, so a limit branch crossing a
// pole or the antimeridian cannot create a bow-tie polygon.
func solarCentralTwoLimitRibbonPieces(
	northern, southern, centerLine []eclipsecore.SolarEclipsePathPoint,
) [][]geodata.GeoPoint {
	if len(northern) < 2 || len(northern) != len(southern) {
		return nil
	}
	pieces := make([][]geodata.GeoPoint, 0, len(northern)+2)
	for index := 1; index < len(northern); index++ {
		northPrevious := geodata.GeoPoint{
			Longitude: northern[index-1].Longitude, Latitude: northern[index-1].Latitude,
		}
		northCurrent := geodata.GeoPoint{
			Longitude: northern[index].Longitude, Latitude: northern[index].Latitude,
		}
		southCurrent := geodata.GeoPoint{
			Longitude: southern[index].Longitude, Latitude: southern[index].Latitude,
		}
		southPrevious := geodata.GeoPoint{
			Longitude: southern[index-1].Longitude, Latitude: southern[index-1].Latitude,
		}
		centerPrevious := solarCentralBandCenterPointAt(centerLine, northern[index-1].Time)
		centerCurrent := solarCentralBandCenterPointAt(centerLine, northern[index].Time)
		for _, piece := range [][]geodata.GeoPoint{
			{northPrevious, northCurrent, centerCurrent},
			{northPrevious, centerCurrent, centerPrevious},
			{centerPrevious, centerCurrent, southCurrent},
			{centerPrevious, southCurrent, southPrevious},
		} {
			if geodata.MeanDistanceKM(piece[0], piece[1]) == 0 &&
				geodata.MeanDistanceKM(piece[1], piece[2]) == 0 {
				continue
			}
			pieces = append(pieces, piece)
		}
	}
	if len(centerLine) >= 2 {
		pieces = append(pieces,
			[]geodata.GeoPoint{
				{Longitude: centerLine[0].Longitude, Latitude: centerLine[0].Latitude},
				{Longitude: northern[0].Longitude, Latitude: northern[0].Latitude},
				{Longitude: southern[0].Longitude, Latitude: southern[0].Latitude},
			},
			[]geodata.GeoPoint{
				{Longitude: centerLine[len(centerLine)-1].Longitude, Latitude: centerLine[len(centerLine)-1].Latitude},
				{Longitude: northern[len(northern)-1].Longitude, Latitude: northern[len(northern)-1].Latitude},
				{Longitude: southern[len(southern)-1].Longitude, Latitude: southern[len(southern)-1].Latitude},
			},
		)
	}
	return pieces
}

func solarCentralBandCenterPointAt(
	centerLine []eclipsecore.SolarEclipsePathPoint,
	target time.Time,
) geodata.GeoPoint {
	if len(centerLine) == 0 {
		return geodata.GeoPoint{}
	}
	if !target.After(centerLine[0].Time) {
		return geodata.GeoPoint{Longitude: centerLine[0].Longitude, Latitude: centerLine[0].Latitude}
	}
	for index := 1; index < len(centerLine); index++ {
		if !target.After(centerLine[index].Time) {
			previous, current := centerLine[index-1], centerLine[index]
			span := current.Time.Sub(previous.Time)
			if span <= 0 {
				return geodata.GeoPoint{Longitude: current.Longitude, Latitude: current.Latitude}
			}
			fraction := float64(target.Sub(previous.Time)) / float64(span)
			return solarCentralBandSphericalInterpolate(
				geodata.GeoPoint{Longitude: previous.Longitude, Latitude: previous.Latitude},
				geodata.GeoPoint{Longitude: current.Longitude, Latitude: current.Latitude},
				fraction,
			)
		}
	}
	last := centerLine[len(centerLine)-1]
	return geodata.GeoPoint{Longitude: last.Longitude, Latitude: last.Latitude}
}

func solarCentralBandHorizonTail(
	north, south eclipsecore.SolarEclipsePathPoint,
	source []eclipsecore.SolarEclipsePathPoint,
) ([]geodata.GeoPoint, bool) {
	if len(source) < 2 {
		return nil, false
	}
	closure := append([]eclipsecore.SolarEclipsePathPoint(nil), source...)
	forwardDistance := solarCentralBandPathDistanceKM(north, closure[0]) +
		solarCentralBandPathDistanceKM(south, closure[len(closure)-1])
	reverseDistance := solarCentralBandPathDistanceKM(north, closure[len(closure)-1]) +
		solarCentralBandPathDistanceKM(south, closure[0])
	if reverseDistance < forwardDistance {
		for left, right := 0, len(closure)-1; left < right; left, right = left+1, right-1 {
			closure[left], closure[right] = closure[right], closure[left]
		}
	}
	tail := make([]geodata.GeoPoint, 0, len(closure)+3)
	tail = append(tail, geodata.GeoPoint{Longitude: north.Longitude, Latitude: north.Latitude})
	for _, point := range closure {
		tail = append(tail, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
	}
	tail = append(tail,
		geodata.GeoPoint{Longitude: south.Longitude, Latitude: south.Latitude},
		tail[0],
	)
	return tail, true
}

func solarCentralTwoLimitRibbonUnionPolygons(
	north, south, centerLine []eclipsecore.SolarEclipsePathPoint,
	overlays [][]geodata.GeoPoint,
) ([][]geodata.GeoPoint, bool) {
	if len(north) < 2 || len(north) != len(south) || len(centerLine) < 2 {
		return nil, false
	}
	ribbons := make([][]geodata.GeoPoint, 0, len(north)+2)
	for index := 1; index < len(north); index++ {
		ribbons = append(ribbons, []geodata.GeoPoint{
			{Longitude: north[index-1].Longitude, Latitude: north[index-1].Latitude},
			{Longitude: north[index].Longitude, Latitude: north[index].Latitude},
			{Longitude: south[index].Longitude, Latitude: south[index].Latitude},
			{Longitude: south[index-1].Longitude, Latitude: south[index-1].Latitude},
		})
	}
	ribbons = append(ribbons,
		[]geodata.GeoPoint{
			{Longitude: centerLine[0].Longitude, Latitude: centerLine[0].Latitude},
			{Longitude: north[0].Longitude, Latitude: north[0].Latitude},
			{Longitude: south[0].Longitude, Latitude: south[0].Latitude},
		},
		[]geodata.GeoPoint{
			{Longitude: centerLine[len(centerLine)-1].Longitude, Latitude: centerLine[len(centerLine)-1].Latitude},
			{Longitude: north[len(north)-1].Longitude, Latitude: north[len(north)-1].Latitude},
			{Longitude: south[len(south)-1].Longitude, Latitude: south[len(south)-1].Latitude},
		},
	)
	paths := [][]geodata.GeoPoint{
		solarCentralBandGeoPoints(north),
		solarCentralBandGeoPoints(south),
		solarCentralBandGeoPoints(centerLine),
	}
	for _, inputs := range [][][]geodata.GeoPoint{
		append(append([][]geodata.GeoPoint(nil), ribbons...), overlays...),
		ribbons,
	} {
		candidates := make([][][]geodata.GeoPoint, 0, 2)
		if polygons, err := geodata.UnionPolygons(inputs); err == nil {
			candidates = append(candidates, polygons)
		}
		// At a polar two-limit contact, longitude/latitude is a singular chart:
		// adjacent time-slice quads can be valid on the sphere but appear to
		// reverse around the pole in the global union. Retry the same faces in a
		// local gnomonic chart before selecting a disconnected fallback.
		if polygons, ok := solarCentralLocalChartUnion(inputs); ok {
			candidates = append(candidates, polygons)
		}
		// Split each temporal quad at the center-line interpolation. This is
		// topologically equivalent away from a pole, but prevents a quad's
		// diagonal from selecting the wrong side when the two limits wrap around
		// a polar chart branch.
		if len(overlays) == 0 {
			pieces := solarCentralTwoLimitRibbonPieces(north, south, centerLine)
			if len(pieces) > 0 {
				if polygons, ok := solarCentralLocalChartUnion(pieces); ok {
					candidates = append(candidates, polygons)
				}
			}
		}
		for _, polygons := range candidates {
			if len(polygons) == 0 || len(polygons) != 1 ||
				geodata.SphericalPolygonsPathMissDistanceKM(polygons, paths, false) > 1 {
				continue
			}
			refined := make([][]geodata.GeoPoint, 0, len(polygons))
			for _, polygon := range polygons {
				refined = append(refined, solarCentralBandRefineRingSpacing(polygon, 200))
			}
			return refined, true
		}
	}
	return nil, false
}

// solarCentralLocalChartUnion performs a boolean union in a local tangent
// chart. Coordinates are scaled before entering the planar union so the
// generic geodata union cannot mistake chart values for global latitudes.
func solarCentralLocalChartUnion(inputs [][]geodata.GeoPoint) ([][]geodata.GeoPoint, bool) {
	if len(inputs) == 0 {
		return nil, false
	}
	toVector := func(point geodata.GeoPoint) [3]float64 {
		latitude := point.Latitude * math.Pi / 180
		longitude := point.Longitude * math.Pi / 180
		cosLatitude := math.Cos(latitude)
		return [3]float64{
			cosLatitude * math.Cos(longitude),
			cosLatitude * math.Sin(longitude),
			math.Sin(latitude),
		}
	}
	dot := func(first, second [3]float64) float64 {
		return first[0]*second[0] + first[1]*second[1] + first[2]*second[2]
	}
	norm := func(value [3]float64) float64 {
		return math.Sqrt(dot(value, value))
	}
	center := [3]float64{}
	maximumAbsLatitude := 0.0
	polarSign := 1.0
	for _, polygon := range inputs {
		for _, point := range openRing(polygon) {
			if absLatitude := math.Abs(point.Latitude); absLatitude > maximumAbsLatitude {
				maximumAbsLatitude = absLatitude
				if point.Latitude < 0 {
					polarSign = -1
				} else {
					polarSign = 1
				}
			}
			vector := toVector(point)
			center[0] += vector[0]
			center[1] += vector[1]
			center[2] += vector[2]
		}
	}
	if maximumAbsLatitude >= 75 {
		// A polar event is better conditioned in a chart centred on the pole
		// than in the arithmetic mean of points whose longitudes wrap around it.
		center = [3]float64{0, 0, polarSign}
	} else {
		centerNorm := norm(center)
		if centerNorm <= 1e-12 {
			return nil, false
		}
		center[0] /= centerNorm
		center[1] /= centerNorm
		center[2] /= centerNorm
	}
	globalNorth := [3]float64{0, 0, 1}
	cross := func(first, second [3]float64) [3]float64 {
		return [3]float64{
			first[1]*second[2] - first[2]*second[1],
			first[2]*second[0] - first[0]*second[2],
			first[0]*second[1] - first[1]*second[0],
		}
	}
	east := cross(globalNorth, center)
	if norm(east) <= 1e-12 {
		east = cross([3]float64{1, 0, 0}, center)
	}
	eastNorm := norm(east)
	if eastNorm <= 1e-12 {
		return nil, false
	}
	east[0] /= eastNorm
	east[1] /= eastNorm
	east[2] /= eastNorm
	north := cross(center, east)
	northNorm := norm(north)
	if northNorm <= 1e-12 {
		return nil, false
	}
	north[0] /= northNorm
	north[1] /= northNorm
	north[2] /= northNorm
	const chartScale = 0.5
	project := func(point geodata.GeoPoint) (geodata.GeoPoint, bool) {
		vector := toVector(point)
		denominator := dot(vector, center)
		if denominator <= 0.02 {
			return geodata.GeoPoint{}, false
		}
		return geodata.GeoPoint{
			Longitude: chartScale * dot(vector, east) / denominator * 180 / math.Pi,
			Latitude:  chartScale * dot(vector, north) / denominator * 180 / math.Pi,
		}, true
	}
	unproject := func(point geodata.GeoPoint) geodata.GeoPoint {
		x := point.Longitude / chartScale * math.Pi / 180
		y := point.Latitude / chartScale * math.Pi / 180
		vector := [3]float64{
			center[0] + x*east[0] + y*north[0],
			center[1] + x*east[1] + y*north[1],
			center[2] + x*east[2] + y*north[2],
		}
		length := norm(vector)
		if length <= 1e-12 {
			return geodata.GeoPoint{}
		}
		return geodata.GeoPoint{
			Longitude: geodata.NormalizeLongitude(math.Atan2(vector[1]/length, vector[0]/length) * 180 / math.Pi),
			Latitude:  math.Asin(math.Max(-1, math.Min(1, vector[2]/length))) * 180 / math.Pi,
		}
	}
	projected := make([][]geodata.GeoPoint, len(inputs))
	for polygonIndex, polygon := range inputs {
		projected[polygonIndex] = make([]geodata.GeoPoint, len(polygon))
		for pointIndex, point := range polygon {
			value, ok := project(point)
			if !ok {
				return nil, false
			}
			projected[polygonIndex][pointIndex] = value
		}
	}
	merged, err := geodata.UnionPolygons(projected)
	if err != nil {
		return nil, false
	}
	result := make([][]geodata.GeoPoint, len(merged))
	for polygonIndex, polygon := range merged {
		result[polygonIndex] = make([]geodata.GeoPoint, len(polygon))
		for pointIndex, point := range polygon {
			result[polygonIndex][pointIndex] = unproject(point)
		}
	}
	return result, true
}

func solarCentralBandGeoPoints(points []eclipsecore.SolarEclipsePathPoint) []geodata.GeoPoint {
	result := make([]geodata.GeoPoint, len(points))
	for index, point := range points {
		result[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
	}
	return result
}

func solarCentralBandRefineRingSpacing(points []geodata.GeoPoint, targetSpacingKM float64) []geodata.GeoPoint {
	points = openRing(points)
	if len(points) < 2 || targetSpacingKM <= 0 {
		return append([]geodata.GeoPoint(nil), points...)
	}
	result := make([]geodata.GeoPoint, 0, len(points))
	for index, start := range points {
		end := points[(index+1)%len(points)]
		result = append(result, start)
		steps := int(math.Ceil(geodata.MeanDistanceKM(start, end) / targetSpacingKM))
		for step := 1; step < steps; step++ {
			result = append(result, solarCentralBandSphericalInterpolate(
				start, end, float64(step)/float64(steps),
			))
		}
	}
	return result
}

func solarCentralBandSphericalInterpolate(
	first, second geodata.GeoPoint,
	fraction float64,
) geodata.GeoPoint {
	toVector := func(point geodata.GeoPoint) [3]float64 {
		latitude := point.Latitude * math.Pi / 180
		longitude := point.Longitude * math.Pi / 180
		cosLatitude := math.Cos(latitude)
		return [3]float64{
			cosLatitude * math.Cos(longitude),
			cosLatitude * math.Sin(longitude),
			math.Sin(latitude),
		}
	}
	firstVector, secondVector := toVector(first), toVector(second)
	dot := math.Max(-1, math.Min(1,
		firstVector[0]*secondVector[0]+firstVector[1]*secondVector[1]+firstVector[2]*secondVector[2],
	))
	angle := math.Acos(dot)
	if angle <= 1e-12 {
		return first
	}
	firstWeight := math.Sin((1-fraction)*angle) / math.Sin(angle)
	secondWeight := math.Sin(fraction*angle) / math.Sin(angle)
	x := firstWeight*firstVector[0] + secondWeight*secondVector[0]
	y := firstWeight*firstVector[1] + secondWeight*secondVector[1]
	z := firstWeight*firstVector[2] + secondWeight*secondVector[2]
	return geodata.GeoPoint{
		Longitude: geodata.NormalizeLongitude(math.Atan2(y, x) * 180 / math.Pi),
		Latitude:  math.Atan2(z, math.Hypot(x, y)) * 180 / math.Pi,
	}
}

func solarCentralBandInnerTransitionCaps(
	footprints []eclipsecore.SolarEclipsePartialFootprint,
) [][]geodata.GeoPoint {
	samples, err := solarCentralShadowSweepSamples(footprints)
	if err != nil {
		return nil
	}
	samples = geodata.DecimateOpenBoundarySweepSamples(samples, len(samples), 40)
	return geodata.OpenBoundarySweepInnerCaps(samples, 500)
}

func solarCentralBandContactCaps(
	footprints []eclipsecore.SolarEclipsePartialFootprint,
	startContact, endContact eclipsecore.SolarEclipsePathPoint,
) [][]geodata.GeoPoint {
	if len(footprints) == 0 {
		return nil
	}
	caps := make([][]geodata.GeoPoint, 0, 2)
	appendCap := func(contact eclipsecore.SolarEclipsePathPoint, footprint eclipsecore.SolarEclipsePartialFootprint) {
		segments := make([][]geodata.GeoPoint, 0, len(footprint.Boundaries))
		for _, source := range footprint.Boundaries {
			segment := make([]geodata.GeoPoint, len(source))
			for index, point := range source {
				segment[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			segments = append(segments, segment)
		}
		boundary := openRing(geodata.JoinPolylineSegments(segments))
		if len(boundary) < 2 {
			return
		}
		contactPoint := geodata.GeoPoint{Longitude: contact.Longitude, Latitude: contact.Latitude}
		if geodata.MeanDistanceKM(contactPoint, boundary[0]) > 2000 ||
			geodata.MeanDistanceKM(contactPoint, boundary[len(boundary)-1]) > 2000 {
			return
		}
		caps = append(caps, []geodata.GeoPoint{contactPoint, boundary[0], boundary[len(boundary)-1]})
	}
	appendCap(startContact, footprints[0])
	appendCap(endContact, footprints[len(footprints)-1])
	return caps
}

func solarCentralTwoLimitAxisCappedPolygon(
	north, south, centerLine []eclipsecore.SolarEclipsePathPoint,
) ([]geodata.GeoPoint, bool) {
	if len(north) < 2 || len(north) != len(south) || len(centerLine) < 2 {
		return nil, false
	}
	polygon := make([]geodata.GeoPoint, 0, len(north)+len(south)+2)
	polygon = append(polygon, geodata.GeoPoint{
		Longitude: centerLine[0].Longitude,
		Latitude:  centerLine[0].Latitude,
	})
	for _, point := range north {
		polygon = append(polygon, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
	}
	lastCenter := centerLine[len(centerLine)-1]
	polygon = append(polygon, geodata.GeoPoint{
		Longitude: lastCenter.Longitude,
		Latitude:  lastCenter.Latitude,
	})
	for index := len(south) - 1; index >= 0; index-- {
		point := south[index]
		polygon = append(polygon, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
	}
	return polygon, true
}

// The public limit series keeps the earlier/later U1/U4 contacts on both
// sides. For map rendering, the axis contacts are the canonical band caps;
// retaining both pairs creates two overlapping triangles at each horizon.
func solarCentralTwoLimitPresentationLimits(
	northern, southern, centerLine []eclipsecore.SolarEclipsePathPoint,
) ([]eclipsecore.SolarEclipsePathPoint, []eclipsecore.SolarEclipsePathPoint, bool) {
	if len(northern) != len(southern) || len(northern) < 4 || len(centerLine) < 2 {
		return nil, nil, false
	}
	start := centerLine[0].Time
	end := centerLine[len(centerLine)-1].Time
	if start.IsZero() || end.IsZero() || !start.Before(end) ||
		!northern[0].Time.Before(start) || !northern[len(northern)-1].Time.After(end) ||
		solarCentralBandPointDistanceKM(northern[0], southern[0]) > 0.001 ||
		solarCentralBandPointDistanceKM(northern[len(northern)-1], southern[len(southern)-1]) > 0.001 {
		return nil, nil, false
	}
	first := 0
	for first < len(northern) && !northern[first].Time.After(start) {
		first++
	}
	last := first
	for last < len(northern) && northern[last].Time.Before(end) {
		last++
	}
	if first == 0 || last >= len(northern) || last-first < 2 {
		return nil, nil, false
	}
	for index := first; index < last; index++ {
		if northern[index].Time.IsZero() || !northern[index].Time.Equal(southern[index].Time) {
			return nil, nil, false
		}
	}
	return northern[first:last], southern[first:last], true
}

func solarCentralPathBandPolygons(
	northern, southern []eclipsecore.SolarEclipsePathPoint,
) ([][]geodata.GeoPoint, string) {
	if len(northern) < 2 || len(northern) != len(southern) {
		return nil, ""
	}
	samples := make([]geodata.OpenBoundarySweepSample, 0, len(northern))
	for index := range northern {
		if northern[index].Time.IsZero() || !northern[index].Time.Equal(southern[index].Time) {
			return nil, ""
		}
		samples = append(samples, geodata.OpenBoundarySweepSample{
			Boundaries: [][]geodata.GeoPoint{{
				{Longitude: northern[index].Longitude, Latitude: northern[index].Latitude},
				{Longitude: southern[index].Longitude, Latitude: southern[index].Latitude},
			}},
		})
	}
	polygons, err := geodata.OpenBoundarySweep(samples)
	if err != nil {
		return nil, ""
	}
	usable := make([][]geodata.GeoPoint, 0, len(polygons))
	for _, polygon := range polygons {
		if len(openRing(polygon)) >= 3 {
			usable = append(usable, polygon)
		}
	}
	if len(usable) == 0 {
		return nil, ""
	}
	return usable, "central-cross-section-sweep"
}

func solarCentralBandEndpointCaps(
	northern, southern, centerLine []eclipsecore.SolarEclipsePathPoint,
) [][]geodata.GeoPoint {
	if len(northern) == 0 || len(northern) != len(southern) || len(centerLine) == 0 {
		return nil
	}
	caps := make([][]geodata.GeoPoint, 0, 2)
	appendCap := func(center eclipsecore.SolarEclipsePathPoint, atStart bool) {
		limitIndex := -1
		if atStart {
			for index := range northern {
				if northern[index].Time.After(center.Time) {
					limitIndex = index
					break
				}
			}
		} else {
			for index := len(northern) - 1; index >= 0; index-- {
				if northern[index].Time.Before(center.Time) {
					limitIndex = index
					break
				}
			}
		}
		if limitIndex < 0 {
			return
		}
		north, south := northern[limitIndex], southern[limitIndex]
		if center.Time.IsZero() || north.Time.IsZero() || south.Time.IsZero() {
			return
		}
		if solarCentralBandPointDistanceKM(center, north) > 3000 || solarCentralBandPointDistanceKM(center, south) > 3000 {
			return
		}
		caps = append(caps, []geodata.GeoPoint{
			{Longitude: center.Longitude, Latitude: center.Latitude},
			{Longitude: north.Longitude, Latitude: north.Latitude},
			{Longitude: south.Longitude, Latitude: south.Latitude},
		})
	}
	appendCap(centerLine[0], true)
	appendCap(centerLine[len(centerLine)-1], false)
	return caps
}

func solarCentralBandPointDistanceKM(first, second eclipsecore.SolarEclipsePathPoint) float64 {
	return geodata.MeanDistanceKM(
		geodata.GeoPoint{Longitude: first.Longitude, Latitude: first.Latitude},
		geodata.GeoPoint{Longitude: second.Longitude, Latitude: second.Latitude},
	)
}

// Near a high-latitude apex a traced limit curve runs through a cusp: its time
// parameterisation folds back on itself, so the ribbon ring built from the two
// limits crosses itself and the exported band twists. The swept region is
// bounded by the envelope, so the samples inside the fold are dropped from both
// sides together — the pairs stay time-aligned, only the fold disappears.
const solarCentralTwoLimitFoldToleranceDegrees = 0.05

func solarCentralTwoLimitEnvelopeSamples(
	northern, southern []eclipsecore.SolarEclipsePathPoint,
) ([]eclipsecore.SolarEclipsePathPoint, []eclipsecore.SolarEclipsePathPoint) {
	drop := make(map[int]bool)
	solarCentralTwoLimitMarkFolds(northern, drop)
	solarCentralTwoLimitMarkFolds(southern, drop)
	if len(drop) == 0 || len(drop) >= len(northern)-2 {
		return northern, southern
	}
	keptNorth := make([]eclipsecore.SolarEclipsePathPoint, 0, len(northern)-len(drop))
	keptSouth := make([]eclipsecore.SolarEclipsePathPoint, 0, len(southern)-len(drop))
	for index := range northern {
		if drop[index] {
			continue
		}
		keptNorth = append(keptNorth, northern[index])
		keptSouth = append(keptSouth, southern[index])
	}
	return keptNorth, keptSouth
}

func solarCentralTwoLimitMarkFolds(limits []eclipsecore.SolarEclipsePathPoint, drop map[int]bool) {
	if len(limits) < 4 {
		return
	}
	// Project the curve onto its end-to-end tangent.  Longitude alone is
	// degenerate at polar apices and cannot distinguish a real turn from an
	// antimeridian wrap; the local east component is scaled by latitude and the
	// north component is retained, so both cusp types are detected.
	lat0 := limits[0].Latitude * math.Pi / 180
	dlon := math.Mod((limits[len(limits)-1].Longitude-limits[0].Longitude)+180, 360) - 180
	dx, dy := dlon*math.Cos(lat0), limits[len(limits)-1].Latitude-limits[0].Latitude
	length := math.Hypot(dx, dy)
	if length <= 1e-9 {
		return
	}
	dx, dy = dx/length, dy/length
	extreme := 0.0
	for index, point := range limits {
		deltaLon := math.Mod((point.Longitude-limits[0].Longitude)+180, 360) - 180
		delta := deltaLon*math.Cos(lat0)*dx + (point.Latitude-limits[0].Latitude)*dy
		if index == 0 {
			extreme = delta
			continue
		}
		if delta-extreme < -solarCentralTwoLimitFoldToleranceDegrees {
			drop[index] = true
			continue
		}
		if delta > extreme {
			extreme = delta
		}
	}
}

func pairedLimitPolygon(
	northern, southern []eclipsecore.SolarEclipsePathPoint,
) ([]geodata.GeoPoint, error) {
	if len(northern) != len(southern) {
		return nil, fmt.Errorf("paired limits must have the same sample count")
	}
	count := len(northern)
	if count < 2 {
		return nil, fmt.Errorf("paired limits require at least two points per side")
	}
	for index := range northern {
		if northern[index].Time.IsZero() || southern[index].Time.IsZero() {
			return nil, fmt.Errorf("paired limit sample %d time is required", index)
		}
		if !northern[index].Time.Equal(southern[index].Time) {
			return nil, fmt.Errorf("paired limit sample %d times must match", index)
		}
	}
	polygon := make([]geodata.GeoPoint, 0, 2*count)
	for _, point := range northern[:count] {
		polygon = append(polygon, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
	}
	for index := count - 1; index >= 0; index-- {
		point := southern[index]
		polygon = append(polygon, geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
	}
	return polygon, nil
}

func appendSolarPathLine(
	features []feature,
	role string,
	points []eclipsecore.SolarEclipsePathPoint,
	properties map[string]interface{},
) ([]feature, error) {
	samples := make([]pathSample, len(points))
	for index, point := range points {
		samples[index] = solarPathSample(point)
	}
	return appendTimedLineFeature(features, solarEclipseEvent, role, samples, properties)
}

func appendSolarSegmentedPathLine(
	features []feature,
	role string,
	segments [][]eclipsecore.SolarEclipsePathPoint,
	properties map[string]interface{},
	requireIncreasingTimes bool,
) ([]feature, error) {
	samples := make([][]pathSample, len(segments))
	for index, segment := range segments {
		samples[index] = sampleSphericalMapPath(solarPathSamples(segment))
	}
	value, times, err := timedMultiLineGeometryFromSegmentsWithTimeOrder(samples, requireIncreasingTimes)
	if err != nil {
		return nil, fmt.Errorf("geojson: %s: %w", role, err)
	}
	properties = cloneProperties(properties)
	properties["times"] = times
	return append(features, newFeature(solarEclipseEvent, role, value, properties)), nil
}

func appendSolarRiseSetCurveFeatures(
	features []feature,
	curves []eclipsecore.SolarEclipseRiseSetCurve,
	properties map[string]interface{},
) ([]feature, error) {
	for _, curve := range curves {
		curveProperties := cloneProperties(properties)
		curveProperties["phase"] = string(curve.Phase)
		curveProperties["horizon"] = string(curve.Direction)
		curveProperties["body"] = "sun"
		var err error
		features, err = appendSolarSegmentedPathLine(
			features, "visibility-boundary", curve.Segments, curveProperties, true,
		)
		if err != nil {
			return nil, err
		}
	}
	return features, nil
}

func solarPathSamples(points []eclipsecore.SolarEclipsePathPoint) []pathSample {
	samples := make([]pathSample, len(points))
	for index, point := range points {
		samples[index] = solarPathSample(point)
	}
	return samples
}

func solarPathSample(point eclipsecore.SolarEclipsePathPoint) pathSample {
	return pathSample{Time: point.Time, Longitude: point.Longitude, Latitude: point.Latitude}
}

func solarEclipseMetadata(info eclipsecore.SolarEclipseInfo) map[string]interface{} {
	properties := map[string]interface{}{
		"eclipse_type":           string(info.Type),
		"model":                  string(info.Model),
		"centrality":             string(info.Centrality),
		"magnitude":              info.Magnitude,
		"gamma":                  info.Gamma,
		"path_width_km":          info.PathWidthKM,
		"path_width_defined":     info.PathWidthDefined,
		"partial_begin_on_earth": formatTime(info.PartialBeginOnEarth),
		"partial_end_on_earth":   formatTime(info.PartialEndOnEarth),
		"central_begin_on_earth": formatTime(info.CentralBeginOnEarth),
		"central_end_on_earth":   formatTime(info.CentralEndOnEarth),
	}
	if info.CentralDuration > 0 {
		// The catalogued maximum duration of the central phase, measured at the
		// greatest eclipse point.
		properties["central_duration_seconds"] = info.CentralDuration.Seconds()
	}
	if info.HasCentral {
		properties["central_duration"] = info.CentralDuration.String()
	}
	return properties
}

// dropDegenerateMultiPolygonRings 删除已经没有面积的面（顶点少于三个互不相同的点，或平面
// 面积为零），保留其余面；全部退化时原样返回，避免把"没有有效环"变成导出失败。
// dropDegenerateMultiPolygonRings removes polygons with no area left (fewer than three
// distinct vertices, or zero planar area) while keeping the rest. When every polygon is
// degenerate the value is returned unchanged so an empty result never becomes an export error.
func dropDegenerateMultiPolygonRings(value geometry) geometry {
	polygons, ok := value.Coordinates.([][][][]float64)
	if !ok {
		return value
	}
	kept := make([][][][]float64, 0, len(polygons))
	for _, polygon := range polygons {
		hasArea := false
		for _, ring := range polygon {
			if !degenerateGeoJSONRing(ring) {
				hasArea = true
				break
			}
		}
		if hasArea {
			kept = append(kept, polygon)
		}
	}
	if len(kept) == 0 || len(kept) == len(polygons) {
		return value
	}
	return geometry{Type: value.Type, Coordinates: kept}
}

// degenerateGeoJSONRing 判断导出环是否已经没有面积。
// degenerateGeoJSONRing reports whether an exported ring has no area left.
func degenerateGeoJSONRing(ring [][]float64) bool {
	distinct := 0
	for index, point := range ring {
		if len(point) < 2 {
			continue
		}
		if index == 0 || !sameDegenerateRingPoint(ring[index-1], point) {
			distinct++
		}
	}
	if distinct > 1 && sameDegenerateRingPoint(ring[0], ring[len(ring)-1]) {
		distinct--
	}
	if distinct < 3 {
		return true
	}
	area := 0.0
	minimumLongitude, maximumLongitude := math.Inf(1), math.Inf(-1)
	for index := range ring {
		next := ring[(index+1)%len(ring)]
		if len(ring[index]) < 2 || len(next) < 2 {
			return false
		}
		area += ring[index][0]*next[1] - next[0]*ring[index][1]
		minimumLongitude = math.Min(minimumLongitude, ring[index][0])
		maximumLongitude = math.Max(maximumLongitude, ring[index][0])
	}
	// 顶点全落在同一条子午线上时面积只剩求和噪声，量级随顶点数增长，可能越过 1e-12 阈值；
	// 必须先按经度跨度判死，不能只靠面积。
	if maximumLongitude-minimumLongitude < 1e-9 {
		return true
	}
	return math.Abs(area/2) < 1e-12
}

// sameDegenerateRingPoint 比较同一环上的两个导出点。
// sameDegenerateRingPoint compares two exported points of one ring.
func sameDegenerateRingPoint(first, second []float64) bool {
	if len(first) < 2 || len(second) < 2 {
		return false
	}
	return math.Abs(math.Remainder(first[0]-second[0], 360)) < 1e-9 &&
		math.Abs(first[1]-second[1]) < 1e-9
}

func lunarEclipseMetadata(info eclipsecore.LunarEclipseInfo) map[string]interface{} {
	return map[string]interface{}{
		"eclipse_type":        string(info.Type),
		"penumbral_magnitude": info.PenumbralMagnitude,
		"umbral_magnitude":    info.UmbralMagnitude,
		"penumbral_start":     formatTime(info.PenumbralStart),
		"partial_start":       formatTime(info.PartialStart),
		"total_start":         formatTime(info.TotalStart),
		"total_end":           formatTime(info.TotalEnd),
		"partial_end":         formatTime(info.PartialEnd),
		"penumbral_end":       formatTime(info.PenumbralEnd),
	}
}

func solarSubsolarPoint(value time.Time, modules ...*Generator) geodata.GeoPoint {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	ttJDE := basic.UTC2TT(basic.Date2JD(value.UTC()))
	ra, dec := basic.HSunApparentRaDecWithInstance(ttJDE, module.inst)
	ut1JDE := basic.TT2UT1(ttJDE)
	longitude := geodata.NormalizeLongitude(ra - basic.ApparentSiderealTime(ut1JDE)*15)
	return geodata.GeoPoint{Longitude: longitude, Latitude: dec}
}

func lunarSubpoint(value time.Time, modules ...*Generator) geodata.GeoPoint {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	ttJDE := basic.UTC2TT(basic.Date2JD(value.UTC()))
	ra, dec := basic.HMoonTrueRaDecWithInstance(ttJDE, module.inst)
	ut1JDE := basic.TT2UT1(ttJDE)
	longitude := geodata.NormalizeLongitude(ra - basic.ApparentSiderealTime(ut1JDE)*15)
	return geodata.GeoPoint{Longitude: longitude, Latitude: dec}
}

func lunarEclipseTimeMarkerSamples(
	info eclipsecore.LunarEclipseInfo,
	scale basic.TimeScale,
	options TimeMarkerOptions, modules ...*Generator) ([]pathSample, error) {
	module := &defaultModule
	if len(modules) > 0 {
		module = modules[0]
	}

	step, err := normalizeTimeMarkerStep(options.Step)
	if err != nil {
		return nil, fmt.Errorf("geojson: lunar eclipse time markers: %w", err)
	}
	location := normalizeTimeMarkerLocation(options.Location)
	start, end := info.PenumbralStart, info.PenumbralEnd
	capacity, err := timeMarkerCapacity(start, end, step, location)
	if err != nil {
		return nil, fmt.Errorf("geojson: lunar eclipse time markers: %w", err)
	}
	current := firstTimeMarkerAfter(start, step, location)
	markers := make([]pathSample, 0, capacity)
	for current.Before(end) {
		point := lunarSubpoint(current, module)
		label := current
		if scale == basic.TimeScaleUT1 {
			label = basic.LabelIn(basic.TimeScaleUT1, current)
		}
		markers = append(markers, pathSample{
			Time:      label,
			Longitude: point.Longitude,
			Latitude:  point.Latitude,
		})
		current = current.Add(step)
	}
	return markers, nil
}

func normalizeLunarBoundaryPoints(value int) int {
	if value <= 0 {
		return defaultLunarBoundaryPoints
	}
	if value < minimumLunarBoundaryPoints {
		return minimumLunarBoundaryPoints
	}
	if value > maximumLunarBoundaryPoints {
		return maximumLunarBoundaryPoints
	}
	return value
}

// timeScaleForMarkers 校验导出时标选项：UT1 时刻没有时区语义，配非 UTC 时区时明确失败。
func timeScaleForMarkers(options *TimeMarkerOptions) (basic.TimeScale, error) {
	if options == nil || options.TimeScale != basic.TimeScaleUT1 {
		return basic.TimeScaleUTC, nil
	}
	if options.Location != nil && options.Location != time.UTC {
		return basic.TimeScaleUTC, fmt.Errorf("geojson: UT1 time properties do not take a non-UTC location")
	}
	return basic.TimeScaleUT1, nil
}
