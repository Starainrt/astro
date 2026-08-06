package svg

import (
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro/basic"
	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/svgmap"
)

const (
	solarEclipseMapDefaultWidth  = 960
	solarEclipseMapDefaultHeight = 640
)

// SolarEclipseMapSVGOptions 控制无国界全球日食地图。
// SolarEclipseMapSVGOptions controls a border-free global solar-eclipse map.
type SolarEclipseMapSVGOptions struct {
	// Width 和 Height 是 SVG 画布尺寸；宽度小于 640 或高度小于 420 时使用 960x640 默认值。
	// Width and Height are SVG canvas dimensions in user units. Width values below 640 and height values below 420 use the 960x640 defaults.
	Width  int
	Height int
	// Language 为 "en"（不区分大小写）时使用英文，否则使用中文。
	// Language uses English for "en" (case-insensitive) and Chinese otherwise.
	Language string
	// Location 控制显示的事件时刻；nil 使用 date.Location()。
	// Location controls displayed event times. Nil uses date.Location().
	Location *time.Location
	// Projection 选择地图投影；零值从事件几何中自动选择，不支持的值使渲染器返回 false。
	// Projection selects the map projection. The zero value selects one from the event geometry; unsupported values make the renderer return false.
	Projection EclipseMapProjection
	// 空文本字段使用本地化的自动标签。
	// Empty text fields use localized automatic labels.
	Title       string
	MapTitle    string
	EventsTitle string
	FooterNote  string

	// PartialStep 是半影足迹请求的时间步长；非正值使用两分钟，正值小于一秒时使用一秒。长事件可能增大实际步长，以保持时间序列不超过 30000 个采样点。
	// PartialStep is the requested partial-footprint time step. Values <= 0 use two minutes; positive values below one second use one second. Long events may use a larger effective step to keep the time series within 30000 samples.
	PartialStep time.Duration
	// BoundaryPoints 是每个瞬时偏食足迹的角向采样数；非正值使用 180，正值限制在 12..1440。
	// BoundaryPoints is the angular sample count for each instantaneous partial footprint. Values <= 0 use 180; positive values are clamped to 12..1440.
	BoundaryPoints int
	// PenumbralOutlineStep 控制半影边界轮廓采样；零值使用 60 分钟，负值禁用，正值小于一分钟时使用一分钟。
	// PenumbralOutlineStep controls sampled penumbral boundary outlines. Zero uses 60 minutes, negative values disable them, and positive values below one minute use one minute.
	PenumbralOutlineStep time.Duration
	// CentralShadowStep 控制本影/反本影轮廓采样；零值使用 10 分钟，负值禁用，正值小于一分钟时使用一分钟。
	// CentralShadowStep controls sampled umbral/antumbral outlines. Zero uses 10 minutes, negative values disable them, and positive values below one minute use one minute.
	CentralShadowStep time.Duration
	// CentralStep 是中心路径请求的时间步长；非正值使用两分钟，正值小于一秒时使用一秒。长事件可能增大实际步长，以保持基础路径不超过 30000 个采样点。
	// CentralStep is the requested central-path time step. Values <= 0 use two minutes; positive values below one second use one second. Long events may use a larger effective step to keep the base path within 30000 samples.
	CentralStep time.Duration
	// TargetSpacingKM 是中心线地面间距上限，单位为千米；非正值使用 150 km，非有限值禁用加密。
	// TargetSpacingKM is the requested maximum center-line ground spacing in kilometers. Values <= 0 use 150 km; non-finite values disable refinement.
	TargetSpacingKM float64
	// TimeLabelStep 控制中心线上的 HH:MM 标签；零值使用 30 分钟，负值禁用标签，正值小于一分钟时使用一分钟。
	// TimeLabelStep controls HH:MM labels along the central line. Zero uses 30 minutes, negative values disable labels, and positive values below one minute use one minute.
	TimeLabelStep time.Duration
}

type solarEclipseMapCalculators struct {
	partial func(time.Time, eclipsecore.SolarEclipsePartialFootprintOptions) (eclipsecore.SolarEclipsePartialFootprintsInfo, bool)
	central func(time.Time, eclipsecore.SolarEclipsePathOptions) (eclipsecore.SolarEclipsePath, bool)
	local   func(time.Time, float64, float64, float64) (eclipsecore.LocalSolarEclipseInfo, bool)
}

// SolarEclipseMapSVG 使用 NASA bulletin Split-K 绘制完整偏食可见范围，并在存在时绘制全食或环食中心线。
// SolarEclipseMapSVG renders the full partial-visibility sweep and, when present, the total or annular central path using NASA bulletin Split-K.
func SolarEclipseMapSVG(date time.Time, options SolarEclipseMapSVGOptions) (string, bool) {
	return SolarEclipseMapSVGNASABulletinSplitK(date, options)
}

// SolarEclipseMapSVGNASABulletinSplitK 使用 NASA bulletin Split-K 绘制地图。
// SolarEclipseMapSVGNASABulletinSplitK renders a NASA bulletin Split-K map.
func SolarEclipseMapSVGNASABulletinSplitK(date time.Time, options SolarEclipseMapSVGOptions) (string, bool) {
	return solarEclipseMapSVG(date, options, solarEclipseMapCalculators{
		partial: eclipsecore.SolarEclipsePartialFootprintsNASABulletinSplitK,
		central: eclipsecore.SolarEclipseCentralPathNASABulletinSplitK,
		local:   eclipsecore.GeometricLocalSolarEclipseOnDateNASABulletinSplitK,
	})
}

// SolarEclipseMapSVGIAUSingleK 使用 IAU Single-K 模型绘制地图。
// SolarEclipseMapSVGIAUSingleK renders an IAU Single-K map.
func SolarEclipseMapSVGIAUSingleK(date time.Time, options SolarEclipseMapSVGOptions) (string, bool) {
	return solarEclipseMapSVG(date, options, solarEclipseMapCalculators{
		partial: eclipsecore.SolarEclipsePartialFootprintsIAUSingleK,
		central: eclipsecore.SolarEclipseCentralPathIAUSingleK,
		local:   eclipsecore.GeometricLocalSolarEclipseOnDateIAUSingleK,
	})
}

func solarEclipseMapSVG(
	date time.Time,
	options SolarEclipseMapSVGOptions,
	calculators solarEclipseMapCalculators,
) (string, bool) {
	if !validEclipseMapProjection(options.Projection) {
		return "", false
	}
	options = normalizeSolarEclipseMapSVGOptions(date, options)
	partial, ok := calculators.partial(date, eclipsecore.SolarEclipsePartialFootprintOptions{
		Step:              options.PartialStep,
		BoundaryPoints:    options.BoundaryPoints,
		CentralShadowStep: options.CentralShadowStep,
	})
	if !ok {
		return "", false
	}
	central, hasCentral := calculators.central(date, eclipsecore.SolarEclipsePathOptions{
		Step:            options.CentralStep,
		TargetSpacingKM: options.TargetSpacingKM,
	})
	local, hasLocal := calculators.local(
		partial.Eclipse.GreatestEclipse,
		partial.Eclipse.GreatestLongitude,
		partial.Eclipse.GreatestLatitude,
		0,
	)
	projection := resolveSolarEclipseMapProjection(partial, central, hasCentral, options.Projection)
	return renderSolarEclipseMapSVG(partial, central, hasCentral, local, hasLocal, options, projection), true
}

func normalizeSolarEclipseMapSVGOptions(date time.Time, options SolarEclipseMapSVGOptions) SolarEclipseMapSVGOptions {
	if options.Width < 640 {
		options.Width = solarEclipseMapDefaultWidth
	}
	if options.Height < 420 {
		options.Height = solarEclipseMapDefaultHeight
	}
	if strings.EqualFold(options.Language, "en") {
		options.Language = "en"
	} else {
		options.Language = "zh"
	}
	if options.Location == nil {
		options.Location = date.Location()
	}
	if options.PartialStep <= 0 {
		options.PartialStep = 2 * time.Minute
	}
	if options.BoundaryPoints <= 0 {
		options.BoundaryPoints = 180
	}
	if options.PenumbralOutlineStep < 0 {
		options.PenumbralOutlineStep = 0
	} else if options.PenumbralOutlineStep == 0 {
		options.PenumbralOutlineStep = time.Hour
	} else if options.PenumbralOutlineStep < time.Minute {
		options.PenumbralOutlineStep = time.Minute
	}
	if options.CentralShadowStep < 0 {
		options.CentralShadowStep = 0
	} else if options.CentralShadowStep == 0 {
		options.CentralShadowStep = 10 * time.Minute
	} else if options.CentralShadowStep < time.Minute {
		options.CentralShadowStep = time.Minute
	}
	if options.CentralStep <= 0 {
		options.CentralStep = 2 * time.Minute
	}
	if options.TargetSpacingKM <= 0 {
		options.TargetSpacingKM = 150
	}
	if options.TimeLabelStep < 0 {
		options.TimeLabelStep = 0
	} else if options.TimeLabelStep == 0 {
		options.TimeLabelStep = 30 * time.Minute
	} else if options.TimeLabelStep < time.Minute {
		options.TimeLabelStep = time.Minute
	}
	return options
}

func resolveSolarEclipseMapProjection(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	central eclipsecore.SolarEclipsePath,
	hasCentral bool,
	requested EclipseMapProjection,
) svgmap.Projection {
	if requested != EclipseMapProjectionAuto {
		return internalEclipseMapProjection(requested)
	}
	focus := partial.Eclipse.GreatestLatitude
	minimum, maximum := focus, focus
	for _, footprint := range partial.Footprints {
		for _, boundary := range footprint.Boundaries {
			for _, point := range boundary {
				minimum = math.Min(minimum, point.Latitude)
				maximum = math.Max(maximum, point.Latitude)
			}
		}
	}
	if hasCentral {
		for _, series := range [][]eclipsecore.SolarEclipsePathPoint{central.CenterLine, central.NorthernLimit, central.SouthernLimit} {
			for _, point := range series {
				minimum = math.Min(minimum, point.Latitude)
				maximum = math.Max(maximum, point.Latitude)
			}
		}
	}
	resolved := svgmap.ResolveProjection("", focus, minimum, maximum)
	if resolved == svgmap.ProjectionEquirectangular && focus >= 65 {
		return svgmap.ProjectionNorthPolar
	}
	if resolved == svgmap.ProjectionEquirectangular && focus <= -65 {
		return svgmap.ProjectionSouthPolar
	}
	return resolved
}

func renderSolarEclipseMapSVG(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	central eclipsecore.SolarEclipsePath,
	hasCentral bool,
	local eclipsecore.LocalSolarEclipseInfo,
	hasLocal bool,
	options SolarEclipseMapSVGOptions,
	projection svgmap.Projection,
) string {
	layout := solarEclipseMapLayoutFor(options, projection)
	frame := layout.frame
	title := solarEclipseMapTitle(partial.Eclipse, options)
	partialPath := solarEclipsePartialSweepPath(partial, frame)

	var builder strings.Builder
	fmt.Fprintf(&builder, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s">`,
		options.Width, options.Height, options.Width, options.Height, html.EscapeString(title))
	builder.WriteString(`<defs>`)
	builder.WriteString(frame.ClipDefinition("solar-map-clip"))
	builder.WriteString(`</defs>`)
	builder.WriteString(`<rect width="100%" height="100%" fill="#efefed"/>`)
	fmt.Fprintf(&builder, `<rect x="22" y="18" width="%d" height="%d" fill="#ffffff" stroke="#c9c9c6" stroke-width="1.2"/>`,
		options.Width-44, options.Height-36)
	fmt.Fprintf(&builder, `<text x="%.3f" y="47" fill="#111111" font-family="Georgia, 'Times New Roman', serif" font-size="24" font-weight="700" text-anchor="middle">%s</text>`,
		float64(options.Width)/2, html.EscapeString(title))
	writeSolarEclipseMapSummary(&builder, partial.Eclipse, local, hasLocal, options)
	writeSolarEclipseMapSectionTitle(&builder, layout, options, hasCentral)

	frame.WriteOcean(&builder)
	frame.WriteGraticule(&builder, "solar-map-clip")
	frame.WriteLand(&builder, "solar-map-clip")
	fmt.Fprintf(&builder, `<path class="partial-eclipse-region" d="%s" clip-path="url(#solar-map-clip)" fill="#dfb84d" fill-opacity="0.46" fill-rule="nonzero"/>`, partialPath)
	writeSolarEclipseTerminator(&builder, partial.Eclipse, frame)
	writeSolarEclipsePenumbralOutlines(&builder, partial, frame, options)
	if hasCentral {
		writeSolarEclipseCentralPath(&builder, central, frame, partial.Eclipse.Type)
	}
	writeSolarEclipseCentralShadowOutlines(&builder, partial.CentralShadowFootprints, frame)
	if hasCentral {
		writeSolarEclipseTimeMarkers(&builder, central, frame, options)
		writeSolarEclipseAxisMarkers(&builder, central, frame, options)
	}
	writeSolarEclipseContactMarkers(&builder, partial, frame)
	writeSolarEclipseGreatestMarker(&builder, partial.Eclipse, frame, options.Language)
	writeSolarEclipseSubsolarMarker(&builder, partial.Eclipse, frame, options.Language)
	frame.WriteFrame(&builder)
	writeSolarEclipseMapLegend(&builder, layout, partial.Eclipse, hasCentral, options)
	writeSolarEclipseGlobalEventsPanel(&builder, partial, central, hasCentral, layout, options)
	writeSolarEclipseMapFooter(&builder, frame, options, projection)
	builder.WriteString(`</svg>`)
	return builder.String()
}

func solarEclipsePartialSweepPath(info eclipsecore.SolarEclipsePartialFootprintsInfo, frame svgmap.Frame) string {
	var builder strings.Builder
	for _, footprint := range info.Footprints {
		segments := make([][]svgmap.GeoPoint, 0, len(footprint.Boundaries))
		for _, source := range footprint.Boundaries {
			segment := make([]svgmap.GeoPoint, len(source))
			for index, point := range source {
				segment[index] = svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			segments = append(segments, segment)
		}
		boundary := svgmap.JoinPolylineSegments(segments)
		if len(boundary) < 3 {
			continue
		}
		if len(boundary) > 1 && svgmap.SameGeoPoint(boundary[0], boundary[len(boundary)-1]) {
			boundary = boundary[:len(boundary)-1]
		}
		polygon := solarEclipsePartialFootprintPolygon(boundary, footprint.Time, footprint.Closed, frame.Projection)
		for _, fragment := range svgmap.PolygonFragments(polygon, frame.Projection) {
			appendEclipseMapPolygonPathConsistent(&builder, frame, fragment)
		}
	}
	return builder.String()
}

func solarEclipsePartialFootprintPolygon(
	boundary []svgmap.GeoPoint,
	value time.Time,
	closed bool,
	projection svgmap.Projection,
) []svgmap.GeoPoint {
	polygon := append([]svgmap.GeoPoint(nil), boundary...)
	if len(boundary) < 2 {
		return polygon
	}

	if !closed {
		terminator := svgmap.SphericalCircle(solarEclipseSubsolarPoint(value), 90, 360)
		polygon = append(polygon, svgmap.ShortestCircleArc(terminator, boundary[len(boundary)-1], boundary[0])...)
	}
	if interior, ok := solarEclipseSphericalBoundaryCentroid(polygon); ok && projection == svgmap.ProjectionEquirectangular {
		polygon = solarEclipseAppendEquirectangularPoleRim(polygon, interior)
	}
	return polygon
}

func solarEclipseSphericalBoundaryCentroid(points []svgmap.GeoPoint) (svgmap.GeoPoint, bool) {
	var sum solarEclipseMapVector
	for _, point := range points {
		value := solarEclipseMapUnitVector(point)
		sum[0] += value[0]
		sum[1] += value[1]
		sum[2] += value[2]
	}
	length := math.Sqrt(solarEclipseMapDot(sum, sum))
	if length < 1e-12 {
		return svgmap.GeoPoint{}, false
	}
	for index := range sum {
		sum[index] /= length
	}
	return svgmap.GeoPoint{
		Longitude: math.Atan2(sum[1], sum[0]) * 180 / math.Pi,
		Latitude:  math.Asin(math.Max(-1, math.Min(1, sum[2]))) * 180 / math.Pi,
	}, true
}

func solarEclipseAppendEquirectangularPoleRim(
	polygon []svgmap.GeoPoint,
	interior svgmap.GeoPoint,
) []svgmap.GeoPoint {
	interiorAngle := solarEclipseSphericalPolygonAngle(polygon, interior)
	if math.Abs(interiorAngle) < math.Pi {
		return polygon
	}
	poleLatitude := 0.0
	for _, candidate := range []float64{90, -90} {
		angle := solarEclipseSphericalPolygonAngle(polygon, svgmap.GeoPoint{Latitude: candidate})
		if math.Abs(angle) >= math.Pi && math.Signbit(angle) == math.Signbit(interiorAngle) {
			poleLatitude = candidate
			break
		}
	}
	if poleLatitude == 0 {
		return polygon
	}

	firstLongitude, lastLongitude := solarEclipseUnwrappedLongitudeEndpoints(polygon)
	delta := firstLongitude - lastLongitude
	if math.Abs(delta) < 180 {
		return polygon
	}
	steps := int(math.Ceil(math.Abs(delta) / 90))
	result := append([]svgmap.GeoPoint(nil), polygon...)
	for step := 1; step <= steps; step++ {
		longitude := lastLongitude + delta*float64(step)/float64(steps)
		result = append(result, svgmap.GeoPoint{
			Longitude: normalizeDegree180(longitude),
			Latitude:  poleLatitude,
		})
	}
	return result
}

func solarEclipseUnwrappedLongitudeEndpoints(points []svgmap.GeoPoint) (float64, float64) {
	first := points[0].Longitude
	previous := first
	for _, point := range points[1:] {
		longitude := point.Longitude
		for longitude-previous > 180 {
			longitude -= 360
		}
		for longitude-previous < -180 {
			longitude += 360
		}
		previous = longitude
	}
	return first, previous
}

type solarEclipseMapVector [3]float64

func solarEclipseSphericalPolygonAngle(points []svgmap.GeoPoint, target svgmap.GeoPoint) float64 {
	if len(points) < 3 {
		return 0
	}
	reference := solarEclipseMapUnitVector(target)
	angle := 0.0
	for index, point := range points {
		current, currentOK := solarEclipseMapTangentDirection(reference, solarEclipseMapUnitVector(point))
		next, nextOK := solarEclipseMapTangentDirection(reference, solarEclipseMapUnitVector(points[(index+1)%len(points)]))
		if !currentOK || !nextOK {
			return math.Copysign(2*math.Pi, angle)
		}
		angle += math.Atan2(
			solarEclipseMapDot(reference, solarEclipseMapCross(current, next)),
			solarEclipseMapDot(current, next),
		)
	}
	return angle
}

func solarEclipseMapUnitVector(point svgmap.GeoPoint) solarEclipseMapVector {
	longitude := point.Longitude * math.Pi / 180
	latitude := point.Latitude * math.Pi / 180
	return solarEclipseMapVector{
		math.Cos(latitude) * math.Cos(longitude),
		math.Cos(latitude) * math.Sin(longitude),
		math.Sin(latitude),
	}
}

func solarEclipseMapTangentDirection(
	reference, value solarEclipseMapVector,
) (solarEclipseMapVector, bool) {
	projection := solarEclipseMapDot(reference, value)
	result := solarEclipseMapVector{
		value[0] - projection*reference[0],
		value[1] - projection*reference[1],
		value[2] - projection*reference[2],
	}
	length := math.Sqrt(solarEclipseMapDot(result, result))
	if length < 1e-12 {
		return solarEclipseMapVector{}, false
	}
	return solarEclipseMapVector{result[0] / length, result[1] / length, result[2] / length}, true
}

func solarEclipseMapDot(a, b solarEclipseMapVector) float64 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

func solarEclipseMapCross(a, b solarEclipseMapVector) solarEclipseMapVector {
	return solarEclipseMapVector{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}

func solarEclipseSubsolarPoint(value time.Time) svgmap.GeoPoint {
	ttJDE := solarEclipseTimeToTTJDE(value)
	ra, dec := basic.HSunApparentRaDec(ttJDE)
	utJDE := basic.TD2UT(ttJDE, false)
	longitude := normalizeDegree180(ra - basic.ApparentSiderealTime(utJDE)*15)
	return svgmap.GeoPoint{Longitude: longitude, Latitude: dec}
}

func appendEclipseMapPolygonPathConsistent(builder *strings.Builder, frame svgmap.Frame, points []svgmap.GeoPoint) {
	if projectedPolygonArea(frame, points) < 0 {
		points = append([]svgmap.GeoPoint(nil), points...)
		reverseGeoPoints(points)
	}
	appendEclipseMapPolygonPath(builder, frame, points)
}

func projectedPolygonArea(frame svgmap.Frame, points []svgmap.GeoPoint) float64 {
	area := 0.0
	for index, point := range points {
		next := points[(index+1)%len(points)]
		x1, y1, ok1 := frame.Project(point.Longitude, point.Latitude)
		x2, y2, ok2 := frame.Project(next.Longitude, next.Latitude)
		if ok1 && ok2 {
			area += x1*y2 - x2*y1
		}
	}
	return area / 2
}

func writeSolarEclipseCentralPath(
	builder *strings.Builder,
	path eclipsecore.SolarEclipsePath,
	frame svgmap.Frame,
	eclipseType eclipsecore.SolarEclipseType,
) {
	count := len(path.NorthernLimit)
	paired := count == len(path.SouthernLimit) && count >= 2
	for index := 0; paired && index < count; index++ {
		paired = !path.NorthernLimit[index].Time.IsZero() &&
			path.NorthernLimit[index].Time.Equal(path.SouthernLimit[index].Time)
	}
	if paired {
		polygon := make([]svgmap.GeoPoint, 0, 2*count)
		for _, point := range path.NorthernLimit[:count] {
			polygon = append(polygon, svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
		}
		for index := count - 1; index >= 0; index-- {
			point := path.SouthernLimit[index]
			polygon = append(polygon, svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude})
		}
		color := solarEclipseCentralPathColor(eclipseType)
		fmt.Fprintf(builder, `<g class="central-eclipse-band" clip-path="url(#solar-map-clip)" fill="%s" fill-opacity="0.82">`, color)
		for _, fragment := range svgmap.PolygonFragments(polygon, frame.Projection) {
			builder.WriteString(`<path d="`)
			appendEclipseMapPolygonPathConsistent(builder, frame, fragment)
			builder.WriteString(`"/>`)
		}
		builder.WriteString(`</g>`)
	}
	writeSolarPathLine(builder, frame, path.NorthernLimit, "northern-central-limit", "#7c2f28", 1.2, "")
	writeSolarPathLine(builder, frame, path.SouthernLimit, "southern-central-limit", "#7c2f28", 1.2, "")
	writeSolarPathLine(builder, frame, path.CenterLine, "solar-center-line", "#263f58", 1.8, "5 3")
}

func writeSolarPathLine(
	builder *strings.Builder,
	frame svgmap.Frame,
	points []eclipsecore.SolarEclipsePathPoint,
	className, color string,
	width float64,
	dash string,
) {
	geographic := make([]svgmap.GeoPoint, len(points))
	for index, point := range points {
		geographic[index] = svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
	}
	writeEclipseMapGeoLine(builder, frame, geographic, className, color, width, dash, "solar-map-clip")
}

func writeSolarEclipseGreatestMarker(
	builder *strings.Builder,
	info eclipsecore.SolarEclipseInfo,
	frame svgmap.Frame,
	language string,
) {
	x, y, ok := frame.Project(info.GreatestLongitude, info.GreatestLatitude)
	if !ok {
		return
	}
	label := "食甚"
	if language == "en" {
		label = "Greatest"
	}
	fmt.Fprintf(builder, `<g class="solar-greatest-marker"><circle cx="%.3f" cy="%.3f" r="5" fill="#c44336" stroke="#fff" stroke-width="1.3"/><text x="%.3f" y="%.3f" fill="#182124" stroke="#fff" stroke-width="3" paint-order="stroke" font-family="Arial, sans-serif" font-size="11" font-weight="700" text-anchor="middle">%s</text></g>`,
		x, y, x, y-10, html.EscapeString(label))
}

func solarEclipseMapTitle(info eclipsecore.SolarEclipseInfo, options SolarEclipseMapSVGOptions) string {
	if options.Title != "" {
		return options.Title
	}
	date := info.GreatestEclipse.In(options.Location).Format("2006-01-02")
	if options.Language == "en" {
		return fmt.Sprintf("%s %s Global Visibility", date, solarEclipseMapTypeName(info.Type, "en"))
	}
	return fmt.Sprintf("%s %s全球见食图", date, solarEclipseMapTypeName(info.Type, "zh"))
}

func writeSolarEclipseMapSummary(
	builder *strings.Builder,
	info eclipsecore.SolarEclipseInfo,
	local eclipsecore.LocalSolarEclipseInfo,
	hasLocal bool,
	options SolarEclipseMapSVGOptions,
) {
	start := info.PartialBeginOnEarth.In(options.Location)
	maximum := info.GreatestEclipse.In(options.Location)
	end := info.PartialEndOnEarth.In(options.Location)
	zone, _ := maximum.Zone()
	if zone == "" {
		zone = "UTC"
	}
	text := fmt.Sprintf("偏食始 %s | 食甚 %s | 偏食终 %s (%s) | 食分 %.3f | Gamma %.4f",
		start.Format("15:04:05"), maximum.Format("15:04:05"), end.Format("15:04:05"), zone, info.Magnitude, info.Gamma)
	if options.Language == "en" {
		text = fmt.Sprintf("Partial begins %s | Greatest %s | Partial ends %s (%s) | magnitude %.3f | Gamma %.4f",
			start.Format("15:04:05"), maximum.Format("15:04:05"), end.Format("15:04:05"), zone, info.Magnitude, info.Gamma)
	}
	fmt.Fprintf(builder, `<text x="%.3f" y="82" fill="#293235" font-family="Arial, sans-serif" font-size="13" text-anchor="middle">%s</text>`,
		float64(options.Width)/2, html.EscapeString(text))

	details := make([]string, 0, 4)
	if info.HasCentral {
		if options.Language == "en" {
			details = append(details, fmt.Sprintf("path width %.1f km", info.PathWidthKM))
		} else {
			details = append(details, fmt.Sprintf("食带宽 %.1f km", info.PathWidthKM))
		}
	}
	if info.HasSaros {
		if options.Language == "en" {
			details = append(details, fmt.Sprintf("Solar Saros %d, member %d/%d", info.Saros.Series, info.Saros.Member, info.Saros.Count))
		} else {
			details = append(details, fmt.Sprintf("太阳沙罗 %d，第 %d/%d 个成员", info.Saros.Series, info.Saros.Member, info.Saros.Count))
		}
	}
	if hasLocal {
		if options.Language == "en" {
			details = append(details, fmt.Sprintf("Sun alt %.1f° az %.1f°", local.SunAltitude, local.SunAzimuth))
		} else {
			details = append(details, fmt.Sprintf("食甚点太阳高度 %.1f° 方位 %.1f°", local.SunAltitude, local.SunAzimuth))
		}
		if local.HasCentral && !local.CentralStart.IsZero() && !local.CentralEnd.IsZero() {
			duration := local.CentralEnd.Sub(local.CentralStart)
			if duration > 0 {
				if options.Language == "en" {
					details = append(details, "central duration "+formatSolarEclipseMapDuration(duration))
				} else {
					details = append(details, "中心食持续 "+formatSolarEclipseMapDuration(duration))
				}
			}
		}
	}
	if len(details) > 0 {
		fmt.Fprintf(builder, `<text x="%.3f" y="106" fill="#596164" font-family="Arial, sans-serif" font-size="11" text-anchor="middle">%s</text>`,
			float64(options.Width)/2, html.EscapeString(strings.Join(details, " | ")))
	}
}

func formatSolarEclipseMapDuration(value time.Duration) string {
	seconds := int(math.Round(value.Seconds()))
	if seconds < 0 {
		seconds = -seconds
	}
	return fmt.Sprintf("%02d:%02d", seconds/60, seconds%60)
}

func writeSolarEclipseMapLegend(
	builder *strings.Builder,
	layout solarEclipseMapLayout,
	info eclipsecore.SolarEclipseInfo,
	hasCentral bool,
	options SolarEclipseMapSVGOptions,
) {
	type legendItem struct {
		label string
		kind  string
		color string
		dash  string
	}
	items := []legendItem{{label: "偏食可见区", kind: "fill", color: "#dfb84d"}}
	if hasCentral {
		items = append(items,
			legendItem{label: solarEclipseCentralPathLabel(info.Type, options.Language), kind: "fill", color: solarEclipseCentralPathColor(info.Type)},
			legendItem{label: "中心线", kind: "line", color: "#263f58", dash: "5 3"},
		)
	}
	if options.PenumbralOutlineStep > 0 {
		items = append(items, legendItem{
			label: solarEclipseOutlineLegendLabel("penumbra", info.Type, options.PenumbralOutlineStep, options.Language),
			kind:  "line", color: "#b07a18", dash: "3 3",
		})
	}
	items = append(items, legendItem{label: "食甚晨昏圈", kind: "line", color: "#6f7778", dash: "4 3"})
	if hasCentral && options.CentralShadowStep > 0 {
		items = append(items, legendItem{
			label: solarEclipseOutlineLegendLabel("central", info.Type, options.CentralShadowStep, options.Language),
			kind:  "line", color: "#7b5a42",
		})
	}
	contactLabel := "P/U 影锥接触"
	if !hasCentral {
		contactLabel = "P 半影接触"
	}
	if options.Language == "en" {
		items[0].label = "Partial-eclipse visibility"
		if hasCentral {
			items[2].label = "Center line"
		}
		for index := range items {
			if items[index].label == "食甚晨昏圈" {
				items[index].label = "Terminator at greatest"
			}
		}
		contactLabel = "P/U shadow contacts"
		if !hasCentral {
			contactLabel = "P penumbral contacts"
		}
	}
	items = append(items, legendItem{label: contactLabel, kind: "contact", color: "#a52d70"})

	columns := len(items)
	if columns > 4 {
		columns = 4
	}
	legendWidth := layout.panelX + layout.panelWidth - layout.frame.X
	itemWidth := legendWidth / float64(columns)
	baseY := layout.frame.Y + layout.frame.Height + 29
	builder.WriteString(`<g class="solar-map-legend">`)
	for index, item := range items {
		row, column := index/columns, index%columns
		x := layout.frame.X + float64(column)*itemWidth
		y := baseY + float64(row)*22
		switch item.kind {
		case "line":
			fmt.Fprintf(builder, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="%s" stroke-width="1.6" stroke-dasharray="%s"/>`,
				x, y-4, x+20, y-4, item.color, item.dash)
		case "contact":
			fmt.Fprintf(builder, `<circle cx="%.3f" cy="%.3f" r="3" fill="%s" stroke="#ffffff" stroke-width="0.8"/>`,
				x+8, y-4, item.color)
		default:
			fmt.Fprintf(builder, `<rect x="%.3f" y="%.3f" width="18" height="9" fill="%s" fill-opacity="0.78"/>`,
				x, y-8, item.color)
		}
		fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#465053" font-family="Arial, sans-serif" font-size="10">%s</text>`,
			x+25, y, html.EscapeString(item.label))
	}
	builder.WriteString(`</g>`)
}

func solarEclipseOutlineLegendLabel(kind string, eclipseType eclipsecore.SolarEclipseType, step time.Duration, language string) string {
	interval := solarEclipseMapStepLabel(step, language)
	if kind == "penumbra" {
		if language == "en" {
			return "Penumbral outlines (" + interval + ")"
		}
		return "半影时刻线（" + interval + "）"
	}
	name := "本影轮廓"
	if eclipseType == eclipsecore.SolarEclipseAnnular {
		name = "反本影轮廓"
	} else if eclipseType == eclipsecore.SolarEclipseHybrid {
		name = "本影/反本影轮廓"
	}
	if language == "en" {
		name = "Umbral outlines"
		if eclipseType == eclipsecore.SolarEclipseAnnular {
			name = "Antumbral outlines"
		} else if eclipseType == eclipsecore.SolarEclipseHybrid {
			name = "Umbral/antumbral outlines"
		}
		return name + " (" + interval + ")"
	}
	return name + "（" + interval + "）"
}

func solarEclipseMapStepLabel(step time.Duration, language string) string {
	if step%time.Minute != 0 {
		return step.String()
	}
	minutes := int(step / time.Minute)
	if language == "en" {
		return fmt.Sprintf("%d min", minutes)
	}
	return fmt.Sprintf("%d 分钟", minutes)
}

func writeSolarEclipseMapFooter(
	builder *strings.Builder,
	frame svgmap.Frame,
	options SolarEclipseMapSVGOptions,
	projection svgmap.Projection,
) {
	text := options.FooterNote
	if text == "" {
		if options.Language == "en" {
			text = eclipseMapProjectionLabel(projection, "en") + "; sampled penumbral sweep and central path; Natural Earth 1:50m physical land, no administrative boundaries."
		} else {
			text = eclipseMapProjectionLabel(projection, "zh") + "；偏食区为半影足迹时间扫掠，叠加中心食带；Natural Earth 1:50m 物理陆地底图，不含行政边界。"
		}
	}
	fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#596164" font-family="Georgia, 'Times New Roman', serif" font-size="11">%s</text>`,
		frame.X, float64(options.Height)-38, html.EscapeString(text))
}

func solarEclipseMapTypeName(value eclipsecore.SolarEclipseType, language string) string {
	if language == "en" {
		switch value {
		case eclipsecore.SolarEclipseTotal:
			return "Total Solar Eclipse"
		case eclipsecore.SolarEclipseAnnular:
			return "Annular Solar Eclipse"
		case eclipsecore.SolarEclipseHybrid:
			return "Hybrid Solar Eclipse"
		default:
			return "Partial Solar Eclipse"
		}
	}
	switch value {
	case eclipsecore.SolarEclipseTotal:
		return "日全食"
	case eclipsecore.SolarEclipseAnnular:
		return "日环食"
	case eclipsecore.SolarEclipseHybrid:
		return "全环食"
	default:
		return "日偏食"
	}
}

func solarEclipseCentralPathLabel(value eclipsecore.SolarEclipseType, language string) string {
	if language == "en" {
		switch value {
		case eclipsecore.SolarEclipseAnnular:
			return "Annular path"
		case eclipsecore.SolarEclipseHybrid:
			return "Hybrid central path"
		default:
			return "Path of totality"
		}
	}
	switch value {
	case eclipsecore.SolarEclipseAnnular:
		return "环食带"
	case eclipsecore.SolarEclipseHybrid:
		return "全环食中心带"
	default:
		return "全食带"
	}
}

func solarEclipseCentralPathColor(value eclipsecore.SolarEclipseType) string {
	if value == eclipsecore.SolarEclipseAnnular {
		return "#a94f3f"
	}
	if value == eclipsecore.SolarEclipseHybrid {
		return "#76506f"
	}
	return "#38516d"
}

func reverseGeoPoints(points []svgmap.GeoPoint) {
	for left, right := 0, len(points)-1; left < right; left, right = left+1, right-1 {
		points[left], points[right] = points[right], points[left]
	}
}
