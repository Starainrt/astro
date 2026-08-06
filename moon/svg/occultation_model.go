package svg

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/starainrt/astro/internal/svgmap"
	"github.com/starainrt/astro/moon"
)

// ErrInvalidStarOccultationPath 表示传入 SVG 渲染器的掩带数据无效。
// ErrInvalidStarOccultationPath reports malformed path data passed to the SVG renderer.
var ErrInvalidStarOccultationPath = errors.New("invalid stellar occultation path")

type starOccultationSVGLayout struct {
	width      float64
	height     float64
	margin     float64
	mapX       float64
	mapY       float64
	mapWidth   float64
	mapHeight  float64
	panelX     float64
	panelY     float64
	panelWidth float64
	footerY    float64
	projection svgmap.Projection
}

type starOccultationGeoPoint struct {
	longitude float64
	latitude  float64
}

type starOccultationEventRow struct {
	name  string
	point moon.OccultationPathPoint
}

func validateStarOccultationPath(path moon.StarOccultationPath) error {
	if !path.Complete {
		return fmt.Errorf("%w: global path is incomplete", ErrInvalidStarOccultationPath)
	}
	if err := validateStarOccultationPathPoint("start", path.Start); err != nil {
		return err
	}
	if err := validateStarOccultationPathPoint("greatest", path.Greatest); err != nil {
		return err
	}
	if err := validateStarOccultationPathPoint("end", path.End); err != nil {
		return err
	}
	if path.Greatest.Time.Before(path.Start.Time) || path.End.Time.Before(path.Greatest.Time) {
		return fmt.Errorf("%w: event times must be ordered start, greatest, end", ErrInvalidStarOccultationPath)
	}
	if err := (moon.OccultationPathOptions{Step: path.Step, TargetSpacingKM: path.TargetSpacingKM}).Validate(); err != nil {
		return fmt.Errorf("%w: invalid path sampling metadata: %v", ErrInvalidStarOccultationPath, err)
	}
	series := []struct {
		name   string
		points []moon.OccultationPathPoint
	}{
		{"center line", path.CenterLine},
		{"northern limit", path.NorthernLimit},
		{"southern limit", path.SouthernLimit},
	}
	for _, current := range series {
		name, points := current.name, current.points
		for index, point := range points {
			if err := validateStarOccultationPathPoint(fmt.Sprintf("%s[%d]", name, index), point); err != nil {
				return err
			}
			if index > 0 && !point.Time.After(points[index-1].Time) {
				return fmt.Errorf("%w: %s times must be strictly increasing", ErrInvalidStarOccultationPath, name)
			}
		}
	}
	if len(path.NorthernLimit) != len(path.SouthernLimit) {
		return fmt.Errorf("%w: northern and southern limits must have the same sample count", ErrInvalidStarOccultationPath)
	}
	if len(path.NorthernLimit) < 2 {
		return fmt.Errorf("%w: global limits must contain start and end", ErrInvalidStarOccultationPath)
	}
	for index := range path.NorthernLimit {
		if !path.NorthernLimit[index].Time.Equal(path.SouthernLimit[index].Time) {
			return fmt.Errorf("%w: northern and southern limit sample %d times must match", ErrInvalidStarOccultationPath, index)
		}
	}
	last := len(path.NorthernLimit) - 1
	if !path.NorthernLimit[0].Time.Equal(path.Start.Time) || !path.SouthernLimit[0].Time.Equal(path.Start.Time) ||
		!path.NorthernLimit[last].Time.Equal(path.End.Time) || !path.SouthernLimit[last].Time.Equal(path.End.Time) {
		return fmt.Errorf("%w: global limits must span start through end", ErrInvalidStarOccultationPath)
	}
	return nil
}

func validateStarOccultationPathPoint(name string, point moon.OccultationPathPoint) error {
	if point.Time.IsZero() {
		return fmt.Errorf("%w: %s time is required", ErrInvalidStarOccultationPath, name)
	}
	if !starOccultationFinite(point.Longitude) || point.Longitude < -180 || point.Longitude > 180 {
		return fmt.Errorf("%w: %s longitude must be in [-180, 180]", ErrInvalidStarOccultationPath, name)
	}
	if !starOccultationFinite(point.Latitude) || point.Latitude < -90 || point.Latitude > 90 {
		return fmt.Errorf("%w: %s latitude must be in [-90, 90]", ErrInvalidStarOccultationPath, name)
	}
	if !starOccultationFinite(point.MoonAltitude) || point.MoonAltitude < -90 || point.MoonAltitude > 90 {
		return fmt.Errorf("%w: %s Moon altitude must be in [-90, 90]", ErrInvalidStarOccultationPath, name)
	}
	if !starOccultationFinite(point.WidthKM) || point.WidthKM < 0 {
		return fmt.Errorf("%w: %s width must be finite and non-negative", ErrInvalidStarOccultationPath, name)
	}
	return nil
}

func starOccultationSVGLayoutFor(
	options StarOccultationSVGOptions,
	headerBottom float64,
	projection svgmap.Projection,
) starOccultationSVGLayout {
	width := float64(options.Width)
	height := float64(options.Height)
	margin := math.Max(22, math.Min(42, width*0.04))
	gap := math.Max(16, math.Min(24, width*0.025))
	panelWidth := math.Max(148, math.Min(238, width*0.23))
	mapWidth := width - 2*margin - gap - panelWidth
	if mapWidth < 220 {
		panelWidth = math.Max(126, width*0.21)
		mapWidth = width - 2*margin - gap - panelWidth
	}
	contentTop := headerBottom + 28
	footerSpace := 82.0
	if projection != svgmap.ProjectionEquirectangular {
		footerSpace = 116
	}
	availableHeight := math.Max(110, height-contentTop-footerSpace)
	mapHeight := math.Min(mapWidth/2, availableHeight)
	if projection != svgmap.ProjectionEquirectangular {
		mapHeight = math.Min(mapWidth, availableHeight)
		mapWidth = mapHeight
	}
	if mapHeight < 110 {
		mapHeight = 110
		mapWidth = math.Min(mapWidth, 2*mapHeight)
	}
	mapY := contentTop + math.Max(0, (availableHeight-mapHeight)/2)
	return starOccultationSVGLayout{
		width:      width,
		height:     height,
		margin:     margin,
		mapX:       margin,
		mapY:       mapY,
		mapWidth:   mapWidth,
		mapHeight:  mapHeight,
		panelX:     margin + mapWidth + gap,
		panelY:     mapY,
		panelWidth: panelWidth,
		footerY:    height - 54,
		projection: projection,
	}
}

func (layout starOccultationSVGLayout) project(longitude, latitude float64) (float64, float64) {
	x, y, _ := layout.mapFrame().Project(longitude, latitude)
	return x, y
}

func (layout starOccultationSVGLayout) mapFrame() svgmap.Frame {
	return svgmap.Frame{
		X:          layout.mapX,
		Y:          layout.mapY,
		Width:      layout.mapWidth,
		Height:     layout.mapHeight,
		Projection: layout.projection,
	}
}

func resolveStarOccultationMapProjection(path moon.StarOccultationPath, requested MapProjection) svgmap.Projection {
	minimumLatitude := path.Greatest.Latitude
	maximumLatitude := path.Greatest.Latitude
	for _, series := range [][]moon.OccultationPathPoint{path.CenterLine, path.NorthernLimit, path.SouthernLimit} {
		for _, point := range series {
			minimumLatitude = math.Min(minimumLatitude, point.Latitude)
			maximumLatitude = math.Max(maximumLatitude, point.Latitude)
		}
	}
	return svgmap.ResolveProjection(internalMapProjection(requested), path.Greatest.Latitude, minimumLatitude, maximumLatitude)
}

func starOccultationPathSegments(points []moon.OccultationPathPoint) [][]moon.OccultationPathPoint {
	if len(points) == 0 {
		return nil
	}
	segments := make([][]moon.OccultationPathPoint, 0, 2)
	current := []moon.OccultationPathPoint{points[0]}
	for index := 1; index < len(points); index++ {
		if math.Abs(points[index].Longitude-points[index-1].Longitude) <= 180 {
			current = append(current, points[index])
			continue
		}
		boundary, fraction, ok := starOccultationAntimeridianCrossing(points[index-1], points[index])
		if !ok {
			current = append(current, points[index])
			continue
		}
		crossing := starOccultationInterpolatePathPoint(points[index-1], points[index], fraction, boundary)
		current = append(current, crossing)
		if len(current) >= 2 {
			segments = append(segments, current)
		}
		wrapped := crossing
		wrapped.Longitude = -boundary
		current = []moon.OccultationPathPoint{wrapped, points[index]}
	}
	if len(current) >= 2 {
		segments = append(segments, current)
	}
	return segments
}

func starOccultationPathSegmentsForProjection(
	points []moon.OccultationPathPoint,
	projection svgmap.Projection,
) [][]moon.OccultationPathPoint {
	if projection == svgmap.ProjectionEquirectangular {
		return starOccultationPathSegments(points)
	}
	geographic := make([]svgmap.GeoPoint, len(points))
	for index, point := range points {
		geographic[index] = svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
	}
	clipped := svgmap.PolylineSegments(geographic, projection)
	result := make([][]moon.OccultationPathPoint, 0, len(clipped))
	for _, segment := range clipped {
		converted := make([]moon.OccultationPathPoint, len(segment))
		for index, point := range segment {
			converted[index] = moon.OccultationPathPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		result = append(result, converted)
	}
	return result
}

func starOccultationAntimeridianCrossing(a, b moon.OccultationPathPoint) (float64, float64, bool) {
	if math.Abs(b.Longitude-a.Longitude) <= 180 {
		return 0, 0, false
	}
	boundary := 180.0
	adjustedB := b.Longitude
	if a.Longitude < 0 {
		boundary = -180
		adjustedB -= 360
	} else {
		adjustedB += 360
	}
	denominator := adjustedB - a.Longitude
	if math.Abs(denominator) < 1e-12 {
		return 0, 0, false
	}
	fraction := (boundary - a.Longitude) / denominator
	if fraction <= 0 || fraction >= 1 {
		return 0, 0, false
	}
	return boundary, fraction, true
}

func starOccultationInterpolatePathPoint(a, b moon.OccultationPathPoint, fraction, longitude float64) moon.OccultationPathPoint {
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}
	duration := b.Time.Sub(a.Time)
	return moon.OccultationPathPoint{
		Time:         a.Time.Add(time.Duration(float64(duration) * fraction)),
		Longitude:    longitude,
		Latitude:     a.Latitude + (b.Latitude-a.Latitude)*fraction,
		MoonAltitude: a.MoonAltitude + (b.MoonAltitude-a.MoonAltitude)*fraction,
		WidthKM:      a.WidthKM + (b.WidthKM-a.WidthKM)*fraction,
	}
}

func starOccultationBandSegments(
	northern, southern []moon.OccultationPathPoint,
) [][]starOccultationGeoPoint {
	return starOccultationBandFragments(northern, southern, svgmap.ProjectionEquirectangular)
}

func starOccultationBandFragments(
	northern, southern []moon.OccultationPathPoint,
	projection svgmap.Projection,
) [][]starOccultationGeoPoint {
	count := len(northern)
	if len(southern) < count {
		count = len(southern)
	}
	if count < 2 {
		return nil
	}
	polygon := make([]starOccultationGeoPoint, 0, 2*count)
	for _, point := range northern[:count] {
		polygon = append(polygon, starOccultationGeoPoint{point.Longitude, point.Latitude})
	}
	for index := count - 1; index >= 0; index-- {
		point := southern[index]
		polygon = append(polygon, starOccultationGeoPoint{point.Longitude, point.Latitude})
	}
	geographic := make([]svgmap.GeoPoint, len(polygon))
	for index, point := range polygon {
		geographic[index] = svgmap.GeoPoint{Longitude: point.longitude, Latitude: point.latitude}
	}
	fragments := svgmap.PolygonFragments(geographic, projection)
	segments := make([][]starOccultationGeoPoint, 0, len(fragments))
	for _, fragment := range fragments {
		converted := make([]starOccultationGeoPoint, len(fragment))
		for index, point := range fragment {
			converted[index] = starOccultationGeoPoint{longitude: point.Longitude, latitude: point.Latitude}
		}
		if math.Abs(starOccultationPolygonArea(converted)) > 1e-9 {
			segments = append(segments, converted)
		}
	}
	return segments
}

func legacyStarOccultationBandSegments(polygon []starOccultationGeoPoint) [][]starOccultationGeoPoint {
	polygon = starOccultationUnwrapPolygon(polygon)
	minimumLongitude, maximumLongitude := polygon[0].longitude, polygon[0].longitude
	for _, point := range polygon[1:] {
		minimumLongitude = math.Min(minimumLongitude, point.longitude)
		maximumLongitude = math.Max(maximumLongitude, point.longitude)
	}
	firstWorld := int(math.Floor((minimumLongitude + 180) / 360))
	lastWorld := int(math.Floor((maximumLongitude + 180) / 360))
	segments := make([][]starOccultationGeoPoint, 0, lastWorld-firstWorld+1)
	for world := firstWorld; world <= lastWorld; world++ {
		left := -180.0 + 360*float64(world)
		right := 180.0 + 360*float64(world)
		clipped := starOccultationClipPolygonLongitude(polygon, left, true)
		clipped = starOccultationClipPolygonLongitude(clipped, right, false)
		if len(clipped) < 3 {
			continue
		}
		for index := range clipped {
			clipped[index].longitude -= 360 * float64(world)
		}
		if math.Abs(starOccultationPolygonArea(clipped)) > 1e-9 {
			segments = append(segments, clipped)
		}
	}
	return segments
}

func starOccultationUnwrapPolygon(points []starOccultationGeoPoint) []starOccultationGeoPoint {
	if len(points) < 2 {
		return points
	}
	unwrapped := make([]starOccultationGeoPoint, len(points))
	unwrapped[0] = points[0]
	for index := 1; index < len(points); index++ {
		point := points[index]
		previous := unwrapped[index-1].longitude
		for point.longitude-previous > 180 {
			point.longitude -= 360
		}
		for point.longitude-previous < -180 {
			point.longitude += 360
		}
		unwrapped[index] = point
	}
	return unwrapped
}

func starOccultationClipPolygonLongitude(
	points []starOccultationGeoPoint,
	boundary float64,
	keepGreater bool,
) []starOccultationGeoPoint {
	if len(points) == 0 {
		return nil
	}
	inside := func(point starOccultationGeoPoint) bool {
		if keepGreater {
			return point.longitude >= boundary
		}
		return point.longitude <= boundary
	}
	intersect := func(a, b starOccultationGeoPoint) starOccultationGeoPoint {
		fraction := (boundary - a.longitude) / (b.longitude - a.longitude)
		return starOccultationGeoPoint{
			longitude: boundary,
			latitude:  a.latitude + (b.latitude-a.latitude)*fraction,
		}
	}
	clipped := make([]starOccultationGeoPoint, 0, len(points)+2)
	previous := points[len(points)-1]
	previousInside := inside(previous)
	for _, current := range points {
		currentInside := inside(current)
		if currentInside != previousInside {
			clipped = append(clipped, intersect(previous, current))
		}
		if currentInside {
			clipped = append(clipped, current)
		}
		previous = current
		previousInside = currentInside
	}
	return clipped
}

func starOccultationPolygonArea(points []starOccultationGeoPoint) float64 {
	area := 0.0
	for index, point := range points {
		next := points[(index+1)%len(points)]
		area += point.longitude*next.latitude - next.longitude*point.latitude
	}
	return area / 2
}

func starOccultationSVGHeaderLines(path moon.StarOccultationPath, options StarOccultationSVGOptions) []string {
	lines := make([]string, 0, 4)
	for _, item := range []struct {
		text     string
		fontSize float64
	}{
		{starOccultationSVGSummaryText(path, options), 14},
		{starOccultationSVGGreatestText(path, options), 13},
	} {
		lines = append(lines, starOccultationWrapText(item.text, float64(options.Width)-80, item.fontSize)...)
	}
	return lines
}

func starOccultationSVGTitle(path moon.StarOccultationPath, options StarOccultationSVGOptions) string {
	if options.Title != "" {
		return options.Title
	}
	target := path.TargetID
	if target == "" {
		if options.Language == starOccultationSVGLanguageEnglish {
			target = "star"
		} else {
			target = "恒星"
		}
	}
	date := path.Greatest.Time.In(options.Location).Format("2006-01-02")
	if options.Language == starOccultationSVGLanguageEnglish {
		return fmt.Sprintf("%s Lunar Occultation of %s", date, target)
	}
	return fmt.Sprintf("%s 月掩%s全球掩带", date, target)
}

func starOccultationSVGSummaryText(path moon.StarOccultationPath, options StarOccultationSVGOptions) string {
	if options.SummaryText != "" {
		return options.SummaryText
	}
	start := path.Start.Time.In(options.Location)
	greatest := path.Greatest.Time.In(options.Location)
	end := path.End.Time.In(options.Location)
	zone := starOccultationLocationLabel(greatest, options.Location)
	if options.Language == starOccultationSVGLanguageEnglish {
		return fmt.Sprintf("Start %s | Greatest %s | End %s (%s)",
			starOccultationFormatEventTime(start, true),
			starOccultationFormatEventTime(greatest, !starOccultationSameDate(start, greatest)),
			starOccultationFormatEventTime(end, !starOccultationSameDate(start, end)), zone)
	}
	return fmt.Sprintf("掩始 %s | 掩甚 %s | 掩终 %s (%s)",
		starOccultationFormatEventTime(start, true),
		starOccultationFormatEventTime(greatest, !starOccultationSameDate(start, greatest)),
		starOccultationFormatEventTime(end, !starOccultationSameDate(start, end)), zone)
}

func starOccultationSVGGreatestText(path moon.StarOccultationPath, options StarOccultationSVGOptions) string {
	if options.GreatestText != "" {
		return options.GreatestText
	}
	coordinates := starOccultationFormatCoordinates(path.Greatest.Longitude, path.Greatest.Latitude)
	altitude := starOccultationFormatSignedDegree(path.Greatest.MoonAltitude)
	if options.Language == starOccultationSVGLanguageEnglish {
		return fmt.Sprintf("Greatest point %s | path width %.1f km | Moon altitude %s", coordinates, path.Greatest.WidthKM, altitude)
	}
	return fmt.Sprintf("掩甚点 %s | 掩带宽 %.1f km | 月球高度 %s", coordinates, path.Greatest.WidthKM, altitude)
}

func starOccultationSVGMapTitle(options StarOccultationSVGOptions) string {
	if options.MapTitle != "" {
		return options.MapTitle
	}
	if options.Language == starOccultationSVGLanguageEnglish {
		return "Global center line and occultation limits"
	}
	return "全球中心线与掩带边界"
}

func starOccultationSVGContactsTitle(options StarOccultationSVGOptions) string {
	if options.ContactsTitle != "" {
		return options.ContactsTitle
	}
	if options.Language == starOccultationSVGLanguageEnglish {
		return "Global events"
	}
	return "全球事件"
}

func starOccultationSVGFooter(options StarOccultationSVGOptions) string {
	if options.FooterNote != "" {
		return options.FooterNote
	}
	projection := starOccultationProjectionLabel(options.Projection, options.Language)
	if options.Language == starOccultationSVGLanguageEnglish {
		return fmt.Sprintf("%s with Natural Earth 1:50m physical land and no administrative boundaries. Limits use the outer lunar limb on the Earth ellipsoid.", projection)
	}
	return fmt.Sprintf("%s；Natural Earth 1:50m 物理陆地底图，不含行政边界；掩带边界为地球椭球上的月球外缘投影。", projection)
}

func starOccultationProjectionLabel(projection MapProjection, language string) string {
	if language == starOccultationSVGLanguageEnglish {
		switch projection {
		case MapProjectionNorthPolar:
			return "North-polar azimuthal equidistant projection"
		case MapProjectionSouthPolar:
			return "South-polar azimuthal equidistant projection"
		default:
			return "Equirectangular projection"
		}
	}
	switch projection {
	case MapProjectionNorthPolar:
		return "北极方位等距投影"
	case MapProjectionSouthPolar:
		return "南极方位等距投影"
	default:
		return "等经纬投影"
	}
}

func starOccultationSVGEventRows(path moon.StarOccultationPath, language string) []starOccultationEventRow {
	names := []string{"掩始", "掩甚", "掩终"}
	if language == starOccultationSVGLanguageEnglish {
		names = []string{"Start", "Greatest", "End"}
	}
	return []starOccultationEventRow{
		{name: names[0], point: path.Start},
		{name: names[1], point: path.Greatest},
		{name: names[2], point: path.End},
	}
}

func starOccultationFormatEventTime(value time.Time, withDate bool) string {
	layout := "15:04:05.0"
	if withDate {
		layout = "2006-01-02 15:04:05.0"
	}
	return value.Format(layout)
}

func starOccultationFormatCoordinates(longitude, latitude float64) string {
	longitudeSuffix := "E"
	if longitude < 0 {
		longitudeSuffix = "W"
	}
	latitudeSuffix := "N"
	if latitude < 0 {
		latitudeSuffix = "S"
	}
	return fmt.Sprintf("%.4f°%s, %.4f°%s", math.Abs(longitude), longitudeSuffix, math.Abs(latitude), latitudeSuffix)
}

func starOccultationFormatSignedDegree(value float64) string {
	return fmt.Sprintf("%+.1f°", value)
}

func starOccultationLocationLabel(value time.Time, location *time.Location) string {
	if location == time.UTC {
		return "UTC"
	}
	name, offset := value.Zone()
	if name != "" && name != "Local" {
		return name
	}
	hours := float64(offset) / 3600
	return fmt.Sprintf("UTC%+.1f", hours)
}

func starOccultationSameDate(first, second time.Time) bool {
	y1, m1, d1 := first.Date()
	y2, m2, d2 := second.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func starOccultationWrapText(value string, maxWidth, fontSize float64) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if starOccultationTextWidth(value, fontSize) <= maxWidth {
		return []string{value}
	}
	runes := []rune(value)
	lines := make([]string, 0, 2)
	for len(runes) > 0 {
		width := 0.0
		end := 0
		lastSpace := -1
		for end < len(runes) {
			nextWidth := width + starOccultationRuneWidth(runes[end], fontSize)
			if nextWidth > maxWidth && end > 0 {
				break
			}
			width = nextWidth
			if unicode.IsSpace(runes[end]) {
				lastSpace = end
			}
			end++
		}
		if end < len(runes) && lastSpace > 0 {
			end = lastSpace
		}
		if end == 0 {
			end = 1
		}
		line := strings.TrimSpace(string(runes[:end]))
		if line != "" {
			lines = append(lines, line)
		}
		runes = runes[end:]
		for len(runes) > 0 && unicode.IsSpace(runes[0]) {
			runes = runes[1:]
		}
	}
	return lines
}

func starOccultationTitleFontSize(title string, width float64) int {
	for size := 26; size >= 16; size-- {
		if starOccultationTextWidth(title, float64(size)) <= width-80 {
			return size
		}
	}
	return 16
}

func starOccultationTextWidth(value string, fontSize float64) float64 {
	width := 0.0
	for _, current := range value {
		width += starOccultationRuneWidth(current, fontSize)
	}
	return width
}

func starOccultationRuneWidth(value rune, fontSize float64) float64 {
	if unicode.Is(unicode.Han, value) || value > utf8.RuneSelf {
		return fontSize
	}
	if unicode.IsSpace(value) {
		return fontSize * 0.34
	}
	return fontSize * 0.58
}

func starOccultationFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
