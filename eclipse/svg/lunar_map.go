package svg

import (
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro/basic"
	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/lunarhorizon"
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/internal/svgmap"
)

const (
	lunarEclipseMapDefaultWidth  = 960
	lunarEclipseMapDefaultHeight = 640
)

// LunarEclipseMapSVGOptions 控制无国界全球可见性地图。
// LunarEclipseMapSVGOptions controls a border-free global visibility map.
type LunarEclipseMapSVGOptions struct {
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
	// Projection 选择地图投影；零值使用等经纬投影，不支持的值使渲染器返回 false。
	// Projection selects the map projection. The zero value selects the equirectangular projection; unsupported values make the renderer return false.
	Projection EclipseMapProjection
	// 空文本字段使用本地化的自动标签。
	// Empty text fields use localized automatic labels.
	Title      string
	FooterNote string
}

// LunarEclipseMapSVG 使用默认月食模型绘制全球 P1-P4 可见区域。
// LunarEclipseMapSVG renders the global P1-P4 visibility regions using the default lunar-eclipse model.
func LunarEclipseMapSVG(date time.Time, options LunarEclipseMapSVGOptions) (string, bool) {
	return lunarEclipseMapSVG(date, options, eclipsecore.LunarEclipseOnDate)
}

// LunarEclipseMapSVGDanjon 使用 Danjon 模型绘制全球可见区域。
// LunarEclipseMapSVGDanjon renders the global visibility regions with Danjon's model.
func LunarEclipseMapSVGDanjon(date time.Time, options LunarEclipseMapSVGOptions) (string, bool) {
	return lunarEclipseMapSVG(date, options, eclipsecore.LunarEclipseOnDateDanjon)
}

// LunarEclipseMapSVGChauvenet 使用 Chauvenet 模型绘制全球可见区域。
// LunarEclipseMapSVGChauvenet renders the global visibility regions with Chauvenet's model.
func LunarEclipseMapSVGChauvenet(date time.Time, options LunarEclipseMapSVGOptions) (string, bool) {
	return lunarEclipseMapSVG(date, options, eclipsecore.LunarEclipseOnDateChauvenet)
}

func lunarEclipseMapSVG(
	date time.Time,
	options LunarEclipseMapSVGOptions,
	calculator func(time.Time) (eclipsecore.LunarEclipseInfo, bool),
) (string, bool) {
	if !validEclipseMapProjection(options.Projection) {
		return "", false
	}
	info, ok := calculator(date)
	if !ok || info.PenumbralStart.IsZero() || info.PenumbralEnd.IsZero() {
		return "", false
	}
	options = normalizeLunarEclipseMapSVGOptions(date, options)
	projection := internalEclipseMapProjection(options.Projection)
	if projection == "" {
		projection = svgmap.ProjectionEquirectangular
	}
	return renderLunarEclipseMapSVG(info, options, projection), true
}

func normalizeLunarEclipseMapSVGOptions(date time.Time, options LunarEclipseMapSVGOptions) LunarEclipseMapSVGOptions {
	if options.Width < 640 {
		options.Width = lunarEclipseMapDefaultWidth
	}
	if options.Height < 420 {
		options.Height = lunarEclipseMapDefaultHeight
	}
	if strings.EqualFold(options.Language, "en") {
		options.Language = "en"
	} else {
		options.Language = "zh"
	}
	if options.Location == nil {
		options.Location = date.Location()
	}
	return options
}

func renderLunarEclipseMapSVG(
	info eclipsecore.LunarEclipseInfo,
	options LunarEclipseMapSVGOptions,
	projection svgmap.Projection,
) string {
	frame := eclipseMapFrame(options.Width, options.Height, projection, 142, 92)
	title := options.Title
	if title == "" {
		date := info.Maximum.In(options.Location).Format("2006-01-02")
		if options.Language == "en" {
			title = fmt.Sprintf("%s %s Global Visibility", date, lunarEclipseSVGTypeName(info.Type, "en"))
		} else {
			title = fmt.Sprintf("%s %s全球可见图", date, lunarEclipseSVGTypeName(info.Type, "zh"))
		}
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s">`,
		options.Width, options.Height, options.Width, options.Height, html.EscapeString(title))
	builder.WriteString(`<defs>`)
	writeLunarEclipseMapDefinitions(&builder, frame, info)
	builder.WriteString(`</defs>`)
	builder.WriteString(`<rect width="100%" height="100%" fill="#efefed"/>`)
	fmt.Fprintf(&builder, `<rect x="22" y="18" width="%d" height="%d" fill="#ffffff" stroke="#c9c9c6" stroke-width="1.2"/>`,
		options.Width-44, options.Height-36)
	titleText := title
	if options.Title != "" {
		titleText = svgchart.EllipsizeText(title, float64(options.Width)-52, 24)
	}
	fmt.Fprintf(&builder, `<text x="%.3f" y="47" fill="#111111" font-family="Georgia, 'Times New Roman', serif" font-size="24" font-weight="700" text-anchor="middle">%s</text>`,
		float64(options.Width)/2, html.EscapeString(titleText))
	writeLunarEclipseMapSummary(&builder, info, options)

	writeLunarEclipseMapRegions(&builder, frame, info)
	frame.WriteFrame(&builder)
	writeLunarEclipseMapLegend(&builder, frame, float64(options.Width), options.Language)
	writeLunarEclipseMapFooter(&builder, frame, options, projection)
	builder.WriteString(`</svg>`)
	return builder.String()
}

func writeLunarEclipseVisibilityMasks(builder *strings.Builder, frame svgmap.Frame) {
	writeLunarEclipseVisibilityMask(builder, "lunar-not-visible-mask", frame,
		"lunar-visible-p1-shape", "lunar-visible-p4-shape")
	writeLunarEclipseVisibilityMask(builder, "lunar-not-p4-mask", frame, "lunar-visible-p4-shape")
	writeLunarEclipseVisibilityMask(builder, "lunar-not-p1-mask", frame, "lunar-visible-p1-shape")
}

func writeLunarEclipseVisibilityMask(
	builder *strings.Builder,
	id string,
	frame svgmap.Frame,
	excludedShapeIDs ...string,
) {
	fmt.Fprintf(builder, `<mask id="%s" maskUnits="userSpaceOnUse" x="%.3f" y="%.3f" width="%.3f" height="%.3f">`,
		id, frame.X, frame.Y, frame.Width, frame.Height)
	fmt.Fprintf(builder, `<rect x="%.3f" y="%.3f" width="%.3f" height="%.3f" fill="#ffffff"/>`,
		frame.X, frame.Y, frame.Width, frame.Height)
	for _, shapeID := range excludedShapeIDs {
		fmt.Fprintf(builder, `<use href="#%s" fill="#000000"/>`, shapeID)
	}
	builder.WriteString(`</mask>`)
}

func eclipseMapFrame(width, height int, projection svgmap.Projection, top, bottom float64) svgmap.Frame {
	availableWidth := float64(width) - 90
	availableHeight := float64(height) - top - bottom
	mapWidth := math.Min(availableWidth, availableHeight*2)
	mapHeight := mapWidth / 2
	if projection != svgmap.ProjectionEquirectangular {
		mapWidth = math.Min(availableWidth, availableHeight)
		mapHeight = mapWidth
	}
	return svgmap.Frame{
		X:          (float64(width) - mapWidth) / 2,
		Y:          top,
		Width:      mapWidth,
		Height:     mapHeight,
		Projection: projection,
	}
}

func lunarEclipseVisibilityPath(value time.Time, frame svgmap.Frame) (string, []svgmap.GeoPoint) {
	points := basic.MoonHorizon(basic.Date2JDE(value.UTC()), 360)
	boundary := make([]svgmap.GeoPoint, len(points))
	for index, point := range points {
		boundary[index] = svgmap.GeoPoint{Longitude: point[0], Latitude: point[1]}
	}
	if frame.Projection == svgmap.ProjectionEquirectangular {
		// 等距圆柱投影在高纬把经度拉伸：1° 的球面地平弧在极点附近可以横跨近 180° 经度，
		// 只连端点会在极区切出地平线以外的假可见帽（1904-09-24 实测 −0.054°，约 6 km）。
		// 极射投影没有这个畸变，保持原输出。
		// Equirectangular projection stretches longitude near a pole: a 1-degree spherical
		// horizon arc can span almost 180 degrees there, so joining endpoints alone cuts a fake
		// visible cap outside the horizon (measured -0.054 deg, about 6 km, on 1904-09-24).
		// Polar projections have no such distortion and keep their existing output.
		// 门限取画布上约 0.03 像素：极区细分再多也画不出来，而 GeoJSON 用的 0.002°
		// 在 812 像素宽的世界图上相当于 0.005 像素，会让 SVG 无谓地大三分之一。
		// The tolerance is about 0.03 pixel on the canvas: finer subdivision cannot be drawn,
		// while the GeoJSON tolerance of 0.002 deg is 0.005 pixel on this 812-pixel-wide world
		// map and would inflate the SVG by a third for nothing.
		tolerance := 0.03 * 360 / frame.Width
		if tolerance < lunarhorizon.DefaultToleranceDegrees {
			tolerance = lunarhorizon.DefaultToleranceDegrees
		}
		boundary = lunarhorizon.RefineWithin(boundary, value, tolerance)
	}
	var builder strings.Builder
	for _, fragment := range svgmap.PolygonFragments(boundary, frame.Clip()) {
		if frame.IsPolar() {
			fragment = lunarEclipsePolarHorizonClosure(fragment)
		}
		appendEclipseMapPolygonPathConsistent(&builder, frame, fragment)
	}
	closedBoundary := append(append([]svgmap.GeoPoint(nil), boundary...), boundary[0])
	return builder.String(), closedBoundary
}

func lunarEclipsePolarHorizonClosure(points []svgmap.GeoPoint) []svgmap.GeoPoint {
	var result []svgmap.GeoPoint
	for index, point := range points {
		result = append(result, point)
		next := points[(index+1)%len(points)]
		if math.Abs(point.Latitude) > 1e-9 || math.Abs(next.Latitude) > 1e-9 {
			continue
		}
		// The equatorial clipping edge is an arc on a polar map, not a chord.
		// The Moon-visible part of the equator is always the minor arc.
		delta := math.Remainder(next.Longitude-point.Longitude, 360)
		steps := int(math.Ceil(math.Abs(delta)))
		for step := 1; step < steps; step++ {
			result = append(result, svgmap.GeoPoint{Longitude: point.Longitude + delta*float64(step)/float64(steps)})
		}
	}
	return result
}

func appendEclipseMapPolygonPath(builder *strings.Builder, frame svgmap.Frame, points []svgmap.GeoPoint) {
	if len(points) < 3 {
		return
	}
	for index, point := range points {
		x, y, ok := frame.Project(point.Longitude, point.Latitude)
		if !ok {
			continue
		}
		command := "L"
		if index == 0 {
			command = "M"
		}
		fmt.Fprintf(builder, `%s %.3f %.3f `, command, x, y)
	}
	builder.WriteString(`Z `)
}

func writeEclipseMapGeoLine(
	builder *strings.Builder,
	frame svgmap.Frame,
	points []svgmap.GeoPoint,
	className, color string,
	strokeWidth float64,
	dash, clipID, dataSource string,
) {
	for _, segment := range svgmap.PolylineSegments(points, frame.Clip()) {
		if len(segment) < 2 {
			continue
		}
		fmt.Fprintf(builder, `<path class="%s" d="`, className)
		for index, point := range segment {
			x, y, ok := frame.Project(point.Longitude, point.Latitude)
			if !ok {
				continue
			}
			command := "L"
			if index == 0 {
				command = "M"
			}
			fmt.Fprintf(builder, `%s %.3f %.3f `, command, x, y)
		}
		source := ""
		if dataSource != "" {
			source = ` data-source="` + html.EscapeString(dataSource) + `"`
		}
		fmt.Fprintf(builder, `" clip-path="url(#%s)"%s fill="none" stroke="%s" stroke-width="%.2f" stroke-dasharray="%s" stroke-linecap="round"/>`,
			clipID, source, color, strokeWidth, dash)
	}
}

func writeLunarEclipseUnavailableRegion(builder *strings.Builder, frame svgmap.Frame) {
	if frame.IsPolar() {
		fmt.Fprintf(builder, `<circle class="eclipse-unavailable-region" mask="url(#lunar-not-visible-mask)" cx="%.3f" cy="%.3f" r="%.3f" fill="#747b7d" fill-opacity="0.34"/>`,
			frame.X+frame.Width/2, frame.Y+frame.Height/2, frame.Width/2)
		return
	}
	fmt.Fprintf(builder, `<rect class="eclipse-unavailable-region" mask="url(#lunar-not-visible-mask)" x="%.3f" y="%.3f" width="%.3f" height="%.3f" fill="#747b7d" fill-opacity="0.34"/>`,
		frame.X, frame.Y, frame.Width, frame.Height)
}

func writeLunarEclipseMapSummary(builder *strings.Builder, info eclipsecore.LunarEclipseInfo, options LunarEclipseMapSVGOptions) {
	start := info.PenumbralStart.In(options.Location)
	maximum := info.Maximum.In(options.Location)
	end := info.PenumbralEnd.In(options.Location)
	zone, _ := maximum.Zone()
	if zone == "" {
		zone = "UTC"
	}
	text := fmt.Sprintf("P1 %s | 食甚 %s | P4 %s (%s) | 半影食分 %.3f | 本影食分 %.3f",
		start.Format("15:04:05"), maximum.Format("15:04:05"), end.Format("15:04:05"), zone,
		info.PenumbralMagnitude, info.UmbralMagnitude)
	if options.Language == "en" {
		text = fmt.Sprintf("P1 %s | Greatest %s | P4 %s (%s) | penumbral magnitude %.3f | umbral magnitude %.3f",
			start.Format("15:04:05"), maximum.Format("15:04:05"), end.Format("15:04:05"), zone,
			info.PenumbralMagnitude, info.UmbralMagnitude)
	}
	fmt.Fprintf(builder, `<text x="%.3f" y="82" fill="#293235" font-family="Arial, sans-serif" font-size="13" text-anchor="middle">%s</text>`,
		float64(options.Width)/2, html.EscapeString(text))
}

// lunarEclipseMapLegendRows 给出可见性图例的标注位次：每行按实测文本宽度排布并居中在地图框下方。
func lunarEclipseMapLegendRows(frame svgmap.Frame, canvasWidth float64, language string) [][]solarEclipseCardinalLabel {
	labels := []string{"全程可见", "带食月出", "带食月落", "不可见"}
	if language == "en" {
		labels = []string{"Entire eclipse", "Moonrise during eclipse", "Moonset during eclipse", "Not visible"}
	}
	width := func(text string) float64 {
		return lunarEclipseDetailedMapLegendIconWidth +
			svgchart.EstimatedTextWidth(text, lunarEclipseDetailedMapLegendFontSize)
	}
	available := canvasWidth - 2*lunarEclipseDetailedPageInset
	packed := make([][]string, 0, 2)
	current := make([]string, 0, len(labels))
	used := 0.0
	for _, label := range labels {
		item := width(label)
		next := item
		if len(current) > 0 {
			next = used + lunarEclipseDetailedMapLegendGap + item
			if next > available {
				packed = append(packed, current)
				current = make([]string, 0, len(labels))
				next = item
			}
		}
		current = append(current, label)
		used = next
	}
	if len(current) > 0 {
		packed = append(packed, current)
	}
	centreX := frame.X + frame.Width/2
	baseY := frame.Y + frame.Height + 31
	rows := make([][]solarEclipseCardinalLabel, 0, len(packed))
	for index, row := range packed {
		total := 0.0
		for position, label := range row {
			if position > 0 {
				total += lunarEclipseDetailedMapLegendGap
			}
			total += width(label)
		}
		x := centreX - total/2
		if x < lunarEclipseDetailedPageInset {
			x = lunarEclipseDetailedPageInset
		}
		if limit := canvasWidth - lunarEclipseDetailedPageInset - total; x > limit {
			x = limit
		}
		entries := make([]solarEclipseCardinalLabel, 0, len(row))
		for _, label := range row {
			entries = append(entries, solarEclipseCardinalLabel{
				text: label,
				x:    x + lunarEclipseDetailedMapLegendIconWidth,
				y:    baseY + float64(index)*lunarEclipseDetailedMapLegendLineStep,
			})
			x += width(label) + lunarEclipseDetailedMapLegendGap
		}
		rows = append(rows, entries)
	}
	return rows
}

func writeLunarEclipseMapLegend(builder *strings.Builder, frame svgmap.Frame, canvasWidth float64, language string) {
	colors := []string{"#5e846d", "#4e9da0", "#e2aa4b", "#747b7d"}
	index := 0
	for _, row := range lunarEclipseMapLegendRows(frame, canvasWidth, language) {
		for _, entry := range row {
			color := colors[0]
			if index < len(colors) {
				color = colors[index]
			}
			fmt.Fprintf(builder, `<rect x="%.3f" y="%.3f" width="18" height="9" fill="%s" fill-opacity="0.78"/>`,
				entry.x-lunarEclipseDetailedMapLegendIconWidth, entry.y-8, color)
			fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#465053" font-family="Arial, sans-serif" font-size="%.0f">%s</text>`,
				entry.x, entry.y, lunarEclipseDetailedMapLegendFontSize, html.EscapeString(entry.text))
			index++
		}
	}
}

func writeLunarEclipseMapFooter(
	builder *strings.Builder,
	frame svgmap.Frame,
	options LunarEclipseMapSVGOptions,
	projection svgmap.Projection,
) {
	text := options.FooterNote
	if text == "" {
		if options.Language == "en" {
			text = eclipseMapProjectionLabel(projection, "en") + "; P1/P4 Moon-visible hemispheres; Natural Earth 1:50m physical land, no administrative boundaries."
		} else {
			text = eclipseMapProjectionLabel(projection, "zh") + "；按 P1/P4 月球可见半球分区；Natural Earth 1:50m 物理陆地底图，不含行政边界。"
		}
	}
	baseline := float64(options.Height) - 38
	// 默认说明是单行；调用方文本按图框宽度折行，行数按画布底边截断，首行位置不变。
	lines := []string{text}
	if options.FooterNote != "" {
		maxWidth := float64(options.Width) - frame.X - 24
		lines = svgchart.TruncateTextLines(svgchart.WrapText(options.FooterNote, maxWidth, 11), maxWidth, 11,
			svgchart.BaselineLineLimit(11, 15, baseline, float64(options.Height)-4))
	}
	for index, line := range lines {
		fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#596164" font-family="Georgia, 'Times New Roman', serif" font-size="11">%s</text>`,
			frame.X, baseline+float64(index)*15, html.EscapeString(line))
	}
}

func normalizeDegree180(value float64) float64 {
	value = math.Mod(value+180, 360)
	if value < 0 {
		value += 360
	}
	return value - 180
}

// writeLunarEclipseMapDefinitions 输出四类可见性区域所需的形状与掩膜定义。
// 独立全球图和 详细版式组合图共用，形状按传入的图框重建。
func writeLunarEclipseMapDefinitions(b *strings.Builder, frame svgmap.Frame, info eclipsecore.LunarEclipseInfo) {
	b.WriteString(frame.ClipDefinition("lunar-map-clip"))
	startPath, _ := lunarEclipseVisibilityPath(info.PenumbralStart, frame)
	endPath, _ := lunarEclipseVisibilityPath(info.PenumbralEnd, frame)
	maximumPath, _ := lunarEclipseVisibilityPath(info.Maximum, frame)
	fmt.Fprintf(b, `<path id="lunar-visible-p1-shape" d="%s"/>`, startPath)
	fmt.Fprintf(b, `<path id="lunar-visible-p4-shape" d="%s"/>`, endPath)
	fmt.Fprintf(b, `<path id="lunar-visible-maximum-shape" d="%s"/>`, maximumPath)
	b.WriteString(`<clipPath id="lunar-visible-p4"><use href="#lunar-visible-p4-shape"/></clipPath>`)
	b.WriteString(`<clipPath id="lunar-visible-maximum"><use href="#lunar-visible-maximum-shape"/></clipPath>`)
	writeLunarEclipseVisibilityMasks(b, frame)
}

// writeLunarEclipseMapRegions 画底图与四类可见性区域，末尾补两条地平边界线。
func writeLunarEclipseMapRegions(b *strings.Builder, frame svgmap.Frame, info eclipsecore.LunarEclipseInfo) {
	_, startBoundary := lunarEclipseVisibilityPath(info.PenumbralStart, frame)
	_, endBoundary := lunarEclipseVisibilityPath(info.PenumbralEnd, frame)
	frame.WriteOcean(b)
	frame.WriteGraticule(b, "lunar-map-clip")
	frame.WriteLand(b, "lunar-map-clip")
	fmt.Fprintf(b, `<g class="lunar-visibility-regions" data-source="%s" clip-path="url(#lunar-map-clip)">`, eclipseMapSourceLunarVisibilityRegions)
	writeLunarEclipseUnavailableRegion(b, frame)
	b.WriteString(`<use class="visible-at-p1 moonset-region" mask="url(#lunar-not-p4-mask)" href="#lunar-visible-p1-shape" fill="#e2aa4b" fill-opacity="0.34"/>`)
	b.WriteString(`<use class="visible-at-p4 moonrise-region" mask="url(#lunar-not-p1-mask)" href="#lunar-visible-p4-shape" fill="#4e9da0" fill-opacity="0.34"/>`)
	b.WriteString(`<g clip-path="url(#lunar-visible-p4)"><g clip-path="url(#lunar-visible-maximum)"><use class="entire-eclipse-region" href="#lunar-visible-p1-shape" fill="#5e846d" fill-opacity="0.34"/></g></g>`)
	b.WriteString(`</g>`)
	writeEclipseMapGeoLine(b, frame, startBoundary, "p1-horizon", "#a56c16", 1.2, "4 3", "lunar-map-clip", eclipseMapSourceLunarHorizonBoundaries)
	writeEclipseMapGeoLine(b, frame, endBoundary, "p4-horizon", "#197a82", 1.2, "4 3", "lunar-map-clip", eclipseMapSourceLunarHorizonBoundaries)
}
