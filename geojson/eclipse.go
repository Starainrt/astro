package geojson

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/basic"
	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/geodata"
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
	if info.PartialBeginOnEarth.IsZero() || info.PartialEndOnEarth.IsZero() {
		return fmt.Errorf("geojson: solar eclipse partial contact times are required")
	}
	if !info.PartialBeginOnEarth.Before(info.GreatestEclipse) ||
		!info.GreatestEclipse.Before(info.PartialEndOnEarth) {
		return fmt.Errorf("geojson: solar eclipse times must be ordered partial begin, greatest, partial end")
	}
	if err := validateSolarPathPoint("solar greatest", eclipsecore.SolarEclipsePathPoint{
		Time: info.GreatestEclipse, Longitude: info.GreatestLongitude, Latitude: info.GreatestLatitude,
	}); err != nil {
		return err
	}
	previous := time.Time{}
	for index, footprint := range partial.Footprints {
		if footprint.Time.IsZero() {
			return fmt.Errorf("geojson: solar partial footprint %d time is required", index)
		}
		if !previous.IsZero() && !footprint.Time.After(previous) {
			return fmt.Errorf("geojson: solar partial footprint times must be strictly increasing")
		}
		if footprint.Time.Before(info.PartialBeginOnEarth) || footprint.Time.After(info.PartialEndOnEarth) {
			return fmt.Errorf("geojson: solar partial footprint %d time is outside the partial interval", index)
		}
		previous = footprint.Time
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
	solarEclipseEvent = "solar-eclipse"
	lunarEclipseEvent = "lunar-eclipse"

	defaultLunarBoundaryPoints = 360
	minimumLunarBoundaryPoints = 12
	maximumLunarBoundaryPoints = 1440
)

// MarshalSolarEclipse 将日食半影足迹和可选中心食带编码为 GeoJSON。
// MarshalSolarEclipse encodes penumbral footprints and an optional central path as GeoJSON.
func MarshalSolarEclipse(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	central *eclipsecore.SolarEclipsePath,
) ([]byte, error) {
	return marshalSolarEclipse(partial, central, nil)
}

// MarshalSolarEclipseWithTimeMarkers 编码日食，并沿中心线按固定间隔追加 Point 要素；已有要素不变，标记标签使用 options.Location，时间值保持 UTC。
// MarshalSolarEclipseWithTimeMarkers encodes a solar eclipse and adds Point Features at regular intervals along the central line. Existing features are unchanged; marker labels use options.Location while time values stay UTC.
func MarshalSolarEclipseWithTimeMarkers(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	central *eclipsecore.SolarEclipsePath,
	options TimeMarkerOptions,
) ([]byte, error) {
	return marshalSolarEclipse(partial, central, &options)
}

func marshalSolarEclipse(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	central *eclipsecore.SolarEclipsePath,
	markerOptions *TimeMarkerOptions,
) ([]byte, error) {
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

	properties := map[string]interface{}{
		"eclipse_type": string(partial.Eclipse.Type),
		"model":        string(partial.Eclipse.Model),
	}
	features := make([]feature, 0, len(partial.Footprints)+8)
	for _, footprint := range partial.Footprints {
		polygon, err := solarPartialFootprintPolygon(footprint)
		if err != nil {
			return nil, err
		}
		footprintProperties := cloneProperties(properties)
		footprintProperties["time"] = formatTime(footprint.Time)
		footprintProperties["source_boundary_closed"] = footprint.Closed
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
		value, err := multiPolygonGeometry([][]geodata.GeoPoint{polygon})
		if err != nil {
			return nil, fmt.Errorf("geojson: solar partial footprint at %s: %w", formatTime(footprint.Time), err)
		}
		features = append(features, newFeature(
			solarEclipseEvent, "partial-footprint", value, footprintProperties,
		))
	}

	if central != nil {
		if len(central.NorthernLimit) > 0 {
			band, err := pairedLimitPolygon(central.NorthernLimit, central.SouthernLimit)
			if err != nil {
				return nil, fmt.Errorf("geojson: solar central band: %w", err)
			}
			value, err := multiPolygonGeometry([][]geodata.GeoPoint{band})
			if err != nil {
				return nil, fmt.Errorf("geojson: solar central band: %w", err)
			}
			features = append(features, newFeature(
				solarEclipseEvent, "central-band", value, cloneProperties(properties),
			))
		}

		var err error
		features, err = appendSolarPathLine(features, "center-line", central.CenterLine, properties)
		if err != nil {
			return nil, err
		}
		if len(central.NorthernLimit) > 0 {
			features, err = appendSolarPathLine(features, "north-limit", central.NorthernLimit, properties)
			if err != nil {
				return nil, err
			}
			features, err = appendSolarPathLine(features, "south-limit", central.SouthernLimit, properties)
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
	}
	var err error
	features, err = appendPointFeature(
		features, solarEclipseEvent, "greatest", greatest, greatestProperties,
	)
	if err != nil {
		return nil, err
	}
	return marshalFeatureCollection(features)
}

// MarshalLunarEclipse 将月食 P1/P4 可见半球和地平线边界编码为 GeoJSON。
// MarshalLunarEclipse encodes the P1/P4 visible hemispheres and horizon boundaries as GeoJSON.
// boundaryPoints 小于等于零时使用 360；其他值限制在 [12, 1440]。
// boundaryPoints values <= 0 use 360; other values are clamped to [12, 1440].
func MarshalLunarEclipse(info eclipsecore.LunarEclipseInfo, boundaryPoints int) ([]byte, error) {
	return marshalLunarEclipse(info, boundaryPoints, nil)
}

// MarshalLunarEclipseWithTimeMarkers 编码月食，并沿半影开始到结束的月下点轨迹追加 Point 要素。
// MarshalLunarEclipseWithTimeMarkers encodes a lunar eclipse and adds Point Features along the sublunar track from penumbral start through end.
// 已有要素保持不变；标记标签使用 options.Location，时间值保持 UTC。
// Existing features are unchanged; marker labels use options.Location while time values stay UTC.
func MarshalLunarEclipseWithTimeMarkers(
	info eclipsecore.LunarEclipseInfo,
	boundaryPoints int,
	options TimeMarkerOptions,
) ([]byte, error) {
	return marshalLunarEclipse(info, boundaryPoints, &options)
}

func marshalLunarEclipse(
	info eclipsecore.LunarEclipseInfo,
	boundaryPoints int,
	markerOptions *TimeMarkerOptions,
) ([]byte, error) {
	if markerOptions != nil {
		if err := validateTimeMarkerOptions(*markerOptions); err != nil {
			return nil, err
		}
	}
	if err := validateLunarEclipseInfo(info); err != nil {
		return nil, err
	}
	boundaryPoints = normalizeLunarBoundaryPoints(boundaryPoints)
	properties := map[string]interface{}{
		"eclipse_type":    string(info.Type),
		"boundary_points": boundaryPoints,
	}
	features := make([]feature, 0, 5)

	contacts := []struct {
		role        string
		horizonRole string
		time        time.Time
	}{
		{role: "visible-at-p1", horizonRole: "p1-horizon", time: info.PenumbralStart},
		{role: "visible-at-p4", horizonRole: "p4-horizon", time: info.PenumbralEnd},
	}
	for _, contact := range contacts {
		center := lunarSubpoint(contact.time)
		polygons := geodata.VisibleHemispherePolygons(
			center, geodata.ProjectionEquirectangular, boundaryPoints,
		)
		value, err := multiPolygonGeometryFromFragments(polygons)
		if err != nil {
			return nil, fmt.Errorf("geojson: %s: %w", contact.role, err)
		}
		contactProperties := cloneProperties(properties)
		contactProperties["time"] = formatTime(contact.time)
		features = append(features, newFeature(
			lunarEclipseEvent, contact.role, value, contactProperties,
		))

		horizon := geodata.SphericalCircle(center, 90, boundaryPoints)
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
				"time":         formatTime(contact.time),
			},
		))
	}

	maximum := lunarSubpoint(info.Maximum)
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
		markers, markerErr := lunarEclipseTimeMarkerSamples(info, *markerOptions)
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
	return marshalFeatureCollection(features)
}

func solarPartialFootprintPolygon(
	footprint eclipsecore.SolarEclipsePartialFootprint,
) ([]geodata.GeoPoint, error) {
	if footprint.Time.IsZero() {
		return nil, fmt.Errorf("geojson: solar partial footprint time is required")
	}
	segments := make([][]geodata.GeoPoint, 0, len(footprint.Boundaries))
	for _, source := range footprint.Boundaries {
		segment := make([]geodata.GeoPoint, len(source))
		for index, point := range source {
			if err := validateCoordinate(point.Longitude, point.Latitude); err != nil {
				return nil, err
			}
			segment[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		segments = append(segments, segment)
	}
	boundary := geodata.JoinPolylineSegments(segments)
	boundary = openRing(boundary)
	if len(boundary) == 1 && !footprint.Closed {
		return boundary, nil
	}
	minimumPoints := 3
	if !footprint.Closed {
		minimumPoints = 2
	}
	if len(boundary) < minimumPoints {
		return nil, fmt.Errorf("geojson: solar partial footprint boundary is incomplete")
	}
	polygon := append([]geodata.GeoPoint(nil), boundary...)
	if !footprint.Closed {
		terminator := geodata.SphericalCircle(solarSubsolarPoint(footprint.Time), 90, 360)
		arc := geodata.ShortestCircleArc(terminator, boundary[len(boundary)-1], boundary[0])
		if len(arc) > 1 {
			polygon = append(polygon, arc[1:]...)
		}
	}
	if len(openRing(polygon)) < 3 {
		return nil, fmt.Errorf("geojson: solar partial footprint polygon is incomplete")
	}
	return polygon, nil
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
	return map[string]interface{}{
		"eclipse_type":           string(info.Type),
		"model":                  string(info.Model),
		"centrality":             string(info.Centrality),
		"magnitude":              info.Magnitude,
		"gamma":                  info.Gamma,
		"path_width_km":          info.PathWidthKM,
		"partial_begin_on_earth": formatTime(info.PartialBeginOnEarth),
		"partial_end_on_earth":   formatTime(info.PartialEndOnEarth),
		"central_begin_on_earth": formatTime(info.CentralBeginOnEarth),
		"central_end_on_earth":   formatTime(info.CentralEndOnEarth),
	}
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

func solarSubsolarPoint(value time.Time) geodata.GeoPoint {
	ttJDE := basic.TD2UT(basic.Date2JDE(value.UTC()), true)
	ra, dec := basic.HSunApparentRaDec(ttJDE)
	utJDE := basic.TD2UT(ttJDE, false)
	longitude := normalizeLongitude(ra - basic.ApparentSiderealTime(utJDE)*15)
	return geodata.GeoPoint{Longitude: longitude, Latitude: dec}
}

func lunarSubpoint(value time.Time) geodata.GeoPoint {
	ttJDE := basic.TD2UT(basic.Date2JDE(value.UTC()), true)
	ra, dec := basic.HMoonTrueRaDec(ttJDE)
	utJDE := basic.TD2UT(ttJDE, false)
	longitude := normalizeLongitude(ra - basic.ApparentSiderealTime(utJDE)*15)
	return geodata.GeoPoint{Longitude: longitude, Latitude: dec}
}

func lunarEclipseTimeMarkerSamples(
	info eclipsecore.LunarEclipseInfo,
	options TimeMarkerOptions,
) ([]pathSample, error) {
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
		point := lunarSubpoint(current)
		markers = append(markers, pathSample{
			Time:      current,
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
