// Package geojson 将日月食和月掩地理结果编码为 RFC 7946 GeoJSON FeatureCollections。
// 坐标是 WGS84 经度和纬度，单位为度；地图投影和样式由应用处理。
// Package geojson encodes eclipse and lunar-occultation geographic results as RFC 7946 GeoJSON FeatureCollections.
// Coordinates are WGS84 longitude and latitude in degrees; map projection and styling remain application concerns.
//
// 每个要素都包含 event 和 role 属性。带时间的 MultiLineString 要素还包含与坐标段对齐的嵌套 times 数组。
// Times 编码为 UTC RFC 3339 字符串；路径采样由上游日月食和月掩选项控制后再传入本包。
// WithTimeMarkers 变体还会追加 role 为 time-marker 的 Point 要素，标签按请求地点格式化。
// 月掩 fallback 掩带通常为 MultiPolygon；支路跳变留下的孤立零时长截面以 MultiLineString 表达，
// 与连续扫掠同时存在时使用 GeometryCollection，避免用虚假多边形重新连接跳变。
// Every feature has event and role properties. Timed MultiLineString features also contain a nested times array aligned with their coordinate segments.
// Times are encoded as UTC RFC 3339 strings. Path sampling is controlled by the source eclipse and occultation options before values reach this package.
// The WithTimeMarkers variants additionally append Point Features whose role is time-marker and whose label is formatted for the requested location.
// Fallback occultation bands normally use MultiPolygon. Isolated zero-duration sections left by branch changes use MultiLineString,
// combined with a continuous sweep in a GeometryCollection, so discontinuities are never reconnected by a false polygon.
package geojson

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

const (
	featureCollectionType  = "FeatureCollection"
	minimumTimeMarkerStep  = time.Minute
	maximumTimeMarkerCount = 1440
)

type featureCollection struct {
	Type     string    `json:"type"`
	Features []feature `json:"features"`
	// TimeScale 只在 UT1 导出时出现；UTC 导出与既有输出逐字节一致。
	TimeScale string `json:"time_scale,omitempty"`
}

type feature struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   geometry               `json:"geometry"`
}

type geometry struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates,omitempty"`
	Geometries  []geometry  `json:"geometries,omitempty"`
}

type pathSample struct {
	Time      time.Time
	Longitude float64
	Latitude  float64
}

// TimeMarkerOptions 控制供地图客户端绘制路径时间标签的可选 Point Feature。
// TimeMarkerOptions controls optional Point Features used by map clients to draw time labels along a moving event path.
// Step 控制标记间隔；零值使用 30 分钟，并对齐到下一个本地整点边界。
// Step controls the marker interval; zero uses 30 minutes and aligns markers to the next local clock boundary.
// Location 控制 HH:MM 标签，默认 UTC；底层 time 属性仍为 UTC RFC 3339。
// Location controls the HH:MM label and defaults to UTC. The underlying time property remains UTC RFC 3339.
// 正 Step 至少为一分钟，单次导出最多 1440 个标记。
// Positive Step values must be at least one minute, and one export is limited to 1440 markers.
type TimeMarkerOptions struct {
	// Step 是时间标记之间的间隔；零值使用 30 分钟。
	// Step is the interval between time markers; zero uses 30 minutes.
	Step time.Duration
	// Location 是格式化 HH:MM 标签时使用的时区；nil 使用 UTC。
	// Location is the timezone used to format HH:MM labels; nil uses UTC.
	Location *time.Location
	// TimeScale 选择时间属性的时标：零值 UTC；TimeScaleUT1 时所有 time 与 HH:MM 标注都写 UT1 时刻，
	// 并在集合上输出 time_scale 成员消歧（RFC 3339 的 Z 后缀严格说不等于 UT1）。此时 Location 必须是 nil 或 UTC。
	// TimeScale selects the scale of time properties: the zero value is UTC; TimeScaleUT1 writes UT1 labels
	// and adds a time_scale member (a RFC 3339 Z suffix is not exactly UT1). Location must then be nil or UTC.
	TimeScale basic.TimeScale
}

// dropFeaturesByRole 按 role 属性丢弃要素；空列表原样返回。
func dropFeaturesByRole(features []feature, skip []string) []feature {
	if len(skip) == 0 {
		return features
	}
	skipped := make(map[string]bool, len(skip))
	for _, role := range skip {
		skipped[role] = true
	}
	kept := make([]feature, 0, len(features))
	for _, item := range features {
		if role, _ := item.Properties["role"].(string); skipped[role] {
			continue
		}
		kept = append(kept, item)
	}
	return kept
}

func marshalFeatureCollection(features []feature) ([]byte, error) {
	if len(features) == 0 {
		return nil, fmt.Errorf("geojson: no geographic features")
	}
	return encodeFeatureCollection(features)
}

func marshalEmptyFeatureCollection() ([]byte, error) {
	return encodeFeatureCollection(make([]feature, 0))
}

func marshalFeatureCollectionWithTimeScale(features []feature, scale basic.TimeScale) ([]byte, error) {
	if len(features) == 0 {
		return nil, fmt.Errorf("geojson: no geographic features")
	}
	return encodeFeatureCollectionWithTimeScale(features, scale)
}

func encodeFeatureCollection(features []feature) ([]byte, error) {
	return encodeFeatureCollectionWithTimeScale(features, basic.TimeScaleUTC)
}

func encodeFeatureCollectionWithTimeScale(features []feature, scale basic.TimeScale) ([]byte, error) {
	member := ""
	if scale == basic.TimeScaleUT1 {
		member = "UT1"
	}
	value, err := json.Marshal(featureCollection{Type: featureCollectionType, Features: features, TimeScale: member})
	if err != nil {
		return nil, fmt.Errorf("geojson: encode feature collection: %w", err)
	}
	return value, nil
}

func newFeature(event, role string, value geometry, properties map[string]interface{}) feature {
	if properties == nil {
		properties = make(map[string]interface{})
	}
	properties["event"] = event
	properties["role"] = role
	return feature{Type: "Feature", Properties: properties, Geometry: value}
}

func appendTimedLineFeature(
	features []feature,
	event, role string,
	points []pathSample,
	properties map[string]interface{},
) ([]feature, error) {
	value, times, err := timedMultiLineGeometry(points)
	if err != nil {
		return nil, fmt.Errorf("geojson: %s: %w", role, err)
	}
	properties = cloneProperties(properties)
	properties["times"] = times
	return append(features, newFeature(event, role, value, properties)), nil
}

func appendPointFeature(
	features []feature,
	event, role string,
	point pathSample,
	properties map[string]interface{},
) ([]feature, error) {
	if point.Time.IsZero() {
		return nil, fmt.Errorf("geojson: %s: point time is required", role)
	}
	value, err := pointGeometry(point.Longitude, point.Latitude)
	if err != nil {
		return nil, fmt.Errorf("geojson: %s: %w", role, err)
	}
	properties = cloneProperties(properties)
	properties["time"] = formatTime(point.Time)
	return append(features, newFeature(event, role, value, properties)), nil
}

func appendTimeMarkerFeatures(
	features []feature,
	event, sourceRole string,
	points []pathSample,
	options TimeMarkerOptions,
) ([]feature, error) {
	markers, err := timeMarkerPoints(points, options)
	if err != nil {
		return nil, fmt.Errorf("geojson: %s time markers: %w", sourceRole, err)
	}
	return appendTimeMarkerPointFeatures(features, event, sourceRole, markers, options.Location)
}

func validateTimeMarkerOptions(options TimeMarkerOptions) error {
	if _, err := normalizeTimeMarkerStep(options.Step); err != nil {
		return fmt.Errorf("geojson: time markers: %w", err)
	}
	return nil
}

func appendTimeMarkerPointFeatures(
	features []feature,
	event, sourceRole string,
	markers []pathSample,
	location *time.Location,
) ([]feature, error) {
	location = normalizeTimeMarkerLocation(location)
	for _, marker := range markers {
		properties := map[string]interface{}{
			"source_role": sourceRole,
			"label":       marker.Time.In(location).Format("15:04"),
		}
		var err error
		features, err = appendPointFeature(features, event, "time-marker", marker, properties)
		if err != nil {
			return nil, err
		}
	}
	return features, nil
}

func timeMarkerPoints(points []pathSample, options TimeMarkerOptions) ([]pathSample, error) {
	step, err := normalizeTimeMarkerStep(options.Step)
	if err != nil {
		return nil, err
	}
	if len(points) < 2 {
		return nil, nil
	}
	location := normalizeTimeMarkerLocation(options.Location)
	for index, point := range points {
		if point.Time.IsZero() {
			return nil, fmt.Errorf("sample %d has a zero time", index)
		}
		if err := validateCoordinate(point.Longitude, point.Latitude); err != nil {
			return nil, fmt.Errorf("sample %d: %w", index, err)
		}
		if index > 0 && !point.Time.After(points[index-1].Time) {
			return nil, fmt.Errorf("sample times must be strictly increasing")
		}
	}
	start, end := points[0].Time, points[len(points)-1].Time
	capacity, err := timeMarkerCapacity(start, end, step, location)
	if err != nil {
		return nil, err
	}
	current := firstTimeMarkerAfter(start, step, location)
	markers := make([]pathSample, 0, capacity)
	segment := 1
	for current.Before(end) {
		for segment < len(points) && points[segment].Time.Before(current) {
			segment++
		}
		if segment >= len(points) {
			break
		}
		a, b := points[segment-1], points[segment]
		span := b.Time.Sub(a.Time)
		if span > 0 {
			fraction := float64(current.Sub(a.Time)) / float64(span)
			markers = append(markers, interpolateTimeMarker(a, b, fraction, current))
		}
		current = current.Add(step)
	}
	return markers, nil
}

func normalizeTimeMarkerStep(value time.Duration) (time.Duration, error) {
	if value < 0 {
		return 0, fmt.Errorf("step must be zero or positive")
	}
	if value == 0 {
		return 30 * time.Minute, nil
	}
	if value < minimumTimeMarkerStep {
		return 0, fmt.Errorf("step must be at least %s", minimumTimeMarkerStep)
	}
	return value, nil
}

func timeMarkerCapacity(start, end time.Time, step time.Duration, location *time.Location) (int, error) {
	if !end.After(start) {
		return 0, fmt.Errorf("marker time range must be strictly increasing")
	}
	first := firstTimeMarkerAfter(start, step, normalizeTimeMarkerLocation(location))
	if !first.Before(end) {
		return 0, nil
	}
	count := 1 + (end.Sub(first)-time.Nanosecond)/step
	if count > maximumTimeMarkerCount {
		return 0, fmt.Errorf("time marker count %d exceeds limit %d", count, maximumTimeMarkerCount)
	}
	return int(count), nil
}

func firstTimeMarkerAfter(value time.Time, step time.Duration, location *time.Location) time.Time {
	local := value.In(location)
	dayStart := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location)
	elapsed := local.Sub(dayStart)
	return dayStart.Add((elapsed/step + 1) * step)
}

func interpolateTimeMarker(a, b pathSample, fraction float64, value time.Time) pathSample {
	deltaLongitude := b.Longitude - a.Longitude
	if deltaLongitude > 180 {
		deltaLongitude -= 360
	} else if deltaLongitude < -180 {
		deltaLongitude += 360
	}
	return pathSample{
		Time:      value,
		Longitude: geodata.NormalizeLongitude(a.Longitude + fraction*deltaLongitude),
		Latitude:  a.Latitude + fraction*(b.Latitude-a.Latitude),
	}
}

func normalizeTimeMarkerLocation(location *time.Location) *time.Location {
	if location == nil {
		return time.UTC
	}
	return location
}

func cloneProperties(source map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(source)+2)
	for key, value := range source {
		result[key] = value
	}
	return result
}

func pointGeometry(longitude, latitude float64) (geometry, error) {
	if err := validateCoordinate(longitude, latitude); err != nil {
		return geometry{}, err
	}
	return geometry{Type: "Point", Coordinates: []float64{longitude, latitude}}, nil
}

func timedMultiLineGeometry(points []pathSample) (geometry, [][]string, error) {
	return timedMultiLineGeometryFromSegments([][]pathSample{points})
}

func timedMultiLineGeometryFromSegments(sourceSegments [][]pathSample) (geometry, [][]string, error) {
	return timedMultiLineGeometryFromSegmentsWithTimeOrder(sourceSegments, true)
}

func timedMultiLineGeometryFromSegmentsWithTimeOrder(
	sourceSegments [][]pathSample,
	requireIncreasingTimes bool,
) (geometry, [][]string, error) {
	coordinates := make([][][]float64, 0, len(sourceSegments))
	times := make([][]string, 0, len(sourceSegments))
	for _, source := range sourceSegments {
		if len(source) == 0 {
			continue
		}
		if len(source) == 1 {
			point := source[0]
			if point.Time.IsZero() {
				return geometry{}, nil, fmt.Errorf("geojson: timed line contains a zero time")
			}
			if err := validateCoordinate(point.Longitude, point.Latitude); err != nil {
				return geometry{}, nil, err
			}
			position := []float64{point.Longitude, point.Latitude}
			formattedTime := formatTime(point.Time)
			coordinates = append(coordinates, [][]float64{position, append([]float64(nil), position...)})
			times = append(times, []string{formattedTime, formattedTime})
			continue
		}
		segments, err := splitTimedLine(source, requireIncreasingTimes)
		if err != nil {
			return geometry{}, nil, err
		}
		for _, segment := range segments {
			line := make([][]float64, len(segment))
			lineTimes := make([]string, len(segment))
			for index, point := range segment {
				line[index] = []float64{point.Longitude, point.Latitude}
				lineTimes[index] = formatTime(point.Time)
			}
			coordinates = append(coordinates, line)
			times = append(times, lineTimes)
		}
	}
	if len(coordinates) == 0 {
		return geometry{}, nil, fmt.Errorf("geojson: line has no valid segments")
	}
	return geometry{Type: "MultiLineString", Coordinates: coordinates}, times, nil
}

func geoMultiLineGeometry(points []geodata.GeoPoint, closeLine bool) (geometry, error) {
	if len(points) < 2 {
		return geometry{}, fmt.Errorf("geojson: line requires at least two points")
	}
	for _, point := range points {
		if err := validateCoordinate(point.Longitude, point.Latitude); err != nil {
			return geometry{}, err
		}
	}
	geographic := append([]geodata.GeoPoint(nil), points...)
	if closeLine && !geodata.SameGeoPoint(geographic[0], geographic[len(geographic)-1]) {
		geographic = append(geographic, geographic[0])
	}
	segments := geodata.PolylineSegments(geographic, geodata.ClipView{Projection: geodata.ProjectionEquirectangular})
	coordinates := make([][][]float64, 0, len(segments))
	for _, segment := range segments {
		if len(segment) < 2 {
			continue
		}
		line := make([][]float64, len(segment))
		for index, point := range segment {
			line[index] = []float64{point.Longitude, point.Latitude}
		}
		coordinates = append(coordinates, line)
	}
	if len(coordinates) == 0 {
		return geometry{}, fmt.Errorf("geojson: line has no valid segments")
	}
	return geometry{Type: "MultiLineString", Coordinates: coordinates}, nil
}

// multiPolygonFillGeometry tags fill-only polygons, which tolerate a coarser
// map chord bound than the path strokes.
func multiPolygonFillGeometry(polygons [][]geodata.GeoPoint) (geometry, error) {
	return multiPolygonGeometryWithin(
		polygons, sphericalFillChordLimitKM, sphericalFillChordErrorDegrees,
	)
}

func multiPolygonGeometry(polygons [][]geodata.GeoPoint) (geometry, error) {
	return multiPolygonGeometryWithin(
		polygons, sphericalMapChordLimitKM, sphericalMapChordErrorDegrees,
	)
}

func multiPolygonGeometryWithin(
	polygons [][]geodata.GeoPoint,
	chordLimitKM, chordErrorDegrees float64,
) (geometry, error) {
	fragments := make([][]geodata.GeoPoint, 0, len(polygons))
	for index, polygon := range polygons {
		polygon = openRing(polygon)
		if len(polygon) < 3 {
			return geometry{}, fmt.Errorf("geojson: polygon %d requires at least three points", index)
		}
		for _, point := range polygon {
			if err := validateCoordinate(point.Longitude, point.Latitude); err != nil {
				return geometry{}, err
			}
		}
		fragments = append(fragments,
			geodata.PolygonFragments(
				sampleSphericalMapRingWithin(polygon, chordLimitKM, chordErrorDegrees),
				geodata.ClipView{Projection: geodata.ProjectionEquirectangular},
			)...)
	}
	return multiPolygonGeometryFromFragments(fragments)
}

func multiPolygonGeometryFromFragments(fragments [][]geodata.GeoPoint) (geometry, error) {
	coordinates := make([][][][]float64, 0, len(fragments))
	for _, fragment := range fragments {
		ring, err := geoJSONRing(fragment)
		if err != nil {
			return geometry{}, err
		}
		coordinates = append(coordinates, [][][]float64{ring})
	}
	if len(coordinates) == 0 {
		return geometry{}, fmt.Errorf("geojson: polygon has no valid rings")
	}
	return geometry{Type: "MultiPolygon", Coordinates: coordinates}, nil
}

func geoJSONRing(points []geodata.GeoPoint) ([][]float64, error) {
	points = openRing(points)
	if len(points) < 3 {
		return nil, fmt.Errorf("geojson: polygon ring requires at least three points")
	}
	for _, point := range points {
		if err := validateCoordinate(point.Longitude, point.Latitude); err != nil {
			return nil, err
		}
	}
	area := polygonArea(points)
	if math.Abs(area) < 1e-12 {
		return nil, fmt.Errorf("geojson: polygon ring has zero area")
	}
	if area < 0 {
		points = append([]geodata.GeoPoint(nil), points...)
		for left, right := 0, len(points)-1; left < right; left, right = left+1, right-1 {
			points[left], points[right] = points[right], points[left]
		}
	}
	ring := make([][]float64, 0, len(points)+1)
	for _, point := range points {
		ring = append(ring, []float64{point.Longitude, point.Latitude})
	}
	return append(ring, []float64{points[0].Longitude, points[0].Latitude}), nil
}

func openRing(points []geodata.GeoPoint) []geodata.GeoPoint {
	if len(points) > 1 && geodata.SameGeoPoint(points[0], points[len(points)-1]) {
		return points[:len(points)-1]
	}
	return points
}

func polygonArea(points []geodata.GeoPoint) float64 {
	area := 0.0
	for index, point := range points {
		next := points[(index+1)%len(points)]
		area += point.Longitude*next.Latitude - next.Longitude*point.Latitude
	}
	return area / 2
}

func splitTimedLine(points []pathSample, requireIncreasingTimes bool) ([][]pathSample, error) {
	if len(points) < 2 {
		return nil, fmt.Errorf("geojson: line requires at least two points")
	}
	for index, point := range points {
		if point.Time.IsZero() {
			return nil, fmt.Errorf("geojson: timed line contains a zero time")
		}
		if err := validateCoordinate(point.Longitude, point.Latitude); err != nil {
			return nil, err
		}
		if requireIncreasingTimes && index > 0 && !point.Time.After(points[index-1].Time) {
			return nil, fmt.Errorf("geojson: timed line times must be strictly increasing")
		}
	}

	segments := make([][]pathSample, 0, 2)
	current := []pathSample{points[0]}
	previousUnwrapped := points[0]
	worldShift := 0.0
	for index := 1; index < len(points); index++ {
		b := points[index]
		for b.Longitude-previousUnwrapped.Longitude > 180 {
			b.Longitude -= 360
		}
		for b.Longitude-previousUnwrapped.Longitude < -180 {
			b.Longitude += 360
		}
		localLongitude := b.Longitude - worldShift
		if localLongitude >= -180 && localLongitude <= 180 {
			b.Longitude = localLongitude
			current = append(current, b)
			previousUnwrapped = points[index]
			previousUnwrapped.Longitude = b.Longitude + worldShift
			continue
		}
		boundary := 180.0
		if localLongitude < -180 {
			boundary = -180
		}
		unwrappedBoundary := boundary + worldShift
		fraction := (unwrappedBoundary - previousUnwrapped.Longitude) /
			(b.Longitude - previousUnwrapped.Longitude)
		crossing := interpolatePathSample(previousUnwrapped, b, fraction, boundary)
		// 按位置而不是时刻判断是否重复：食甚等时线整条支路共用同一时刻，按时刻判断会把接缝点丢掉，
		// 于是反经线两侧各留一个端点，屏幕上就是一条缺口。
		last := current[len(current)-1]
		if crossing.Longitude != last.Longitude || crossing.Latitude != last.Latitude {
			current = append(current, crossing)
		}
		if len(current) >= 2 {
			segments = append(segments, current)
		}
		crossing.Longitude = -boundary
		if boundary > 0 {
			worldShift += 360
		} else {
			worldShift -= 360
		}
		b.Longitude -= worldShift
		current = []pathSample{crossing, b}
		previousUnwrapped = points[index]
		previousUnwrapped.Longitude = b.Longitude + worldShift
	}
	if len(current) >= 2 {
		segments = append(segments, current)
	}
	if len(segments) == 0 {
		return nil, fmt.Errorf("geojson: line has no valid segments")
	}
	return segments, nil
}

func interpolatePathSample(a, b pathSample, fraction, longitude float64) pathSample {
	duration := b.Time.Sub(a.Time)
	return pathSample{
		Time:      a.Time.Add(time.Duration(float64(duration) * fraction)),
		Longitude: longitude,
		Latitude:  a.Latitude + (b.Latitude-a.Latitude)*fraction,
	}
}

func validateCoordinate(longitude, latitude float64) error {
	if math.IsNaN(longitude) || math.IsInf(longitude, 0) || longitude < -180 || longitude > 180 {
		return fmt.Errorf("geojson: longitude must be finite and within [-180, 180]")
	}
	if math.IsNaN(latitude) || math.IsInf(latitude, 0) || latitude < -90 || latitude > 90 {
		return fmt.Errorf("geojson: latitude must be finite and within [-90, 90]")
	}
	return nil
}

func finiteGeoJSON(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}
