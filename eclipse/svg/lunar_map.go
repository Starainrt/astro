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
	startPath, startBoundary := lunarEclipseVisibilityPath(info.PenumbralStart, frame)
	endPath, endBoundary := lunarEclipseVisibilityPath(info.PenumbralEnd, frame)
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
	builder.WriteString(frame.ClipDefinition("lunar-map-clip"))
	fmt.Fprintf(&builder, `<path id="lunar-visible-p1-shape" d="%s"/>`, startPath)
	fmt.Fprintf(&builder, `<path id="lunar-visible-p4-shape" d="%s"/>`, endPath)
	builder.WriteString(`<clipPath id="lunar-visible-p4"><use href="#lunar-visible-p4-shape"/></clipPath>`)
	writeLunarEclipseVisibilityMasks(&builder, frame)
	builder.WriteString(`</defs>`)
	builder.WriteString(`<rect width="100%" height="100%" fill="#efefed"/>`)
	fmt.Fprintf(&builder, `<rect x="22" y="18" width="%d" height="%d" fill="#ffffff" stroke="#c9c9c6" stroke-width="1.2"/>`,
		options.Width-44, options.Height-36)
	fmt.Fprintf(&builder, `<text x="%.3f" y="47" fill="#111111" font-family="Georgia, 'Times New Roman', serif" font-size="24" font-weight="700" text-anchor="middle">%s</text>`,
		float64(options.Width)/2, html.EscapeString(title))
	writeLunarEclipseMapSummary(&builder, info, options)

	frame.WriteOcean(&builder)
	frame.WriteGraticule(&builder, "lunar-map-clip")
	frame.WriteLand(&builder, "lunar-map-clip")
	fmt.Fprintf(&builder, `<g class="lunar-visibility-regions" clip-path="url(#lunar-map-clip)">`)
	writeLunarEclipseUnavailableRegion(&builder, frame)
	builder.WriteString(`<use class="visible-at-p1 moonset-region" mask="url(#lunar-not-p4-mask)" href="#lunar-visible-p1-shape" fill="#e2aa4b" fill-opacity="0.34"/>`)
	builder.WriteString(`<use class="visible-at-p4 moonrise-region" mask="url(#lunar-not-p1-mask)" href="#lunar-visible-p4-shape" fill="#4e9da0" fill-opacity="0.34"/>`)
	builder.WriteString(`<g clip-path="url(#lunar-visible-p4)"><use class="entire-eclipse-region" href="#lunar-visible-p1-shape" fill="#5e846d" fill-opacity="0.34"/></g>`)
	builder.WriteString(`</g>`)
	writeEclipseMapGeoLine(&builder, frame, startBoundary, "p1-horizon", "#a56c16", 1.2, "4 3", "lunar-map-clip")
	writeEclipseMapGeoLine(&builder, frame, endBoundary, "p4-horizon", "#197a82", 1.2, "4 3", "lunar-map-clip")
	frame.WriteFrame(&builder)
	writeLunarEclipseMapLegend(&builder, frame, options.Language)
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
	center := lunarEclipseSubpoint(value)
	boundary := svgmap.SphericalCircle(center, 90, 360)
	path := lunarEclipseVisibilityPathForCenter(center, frame)
	closedBoundary := append(append([]svgmap.GeoPoint(nil), boundary...), boundary[0])
	return path, closedBoundary
}

func lunarEclipseVisibilityPathForCenter(center svgmap.GeoPoint, frame svgmap.Frame) string {
	var builder strings.Builder
	for _, polygon := range svgmap.VisibleHemispherePolygons(center, frame.Projection, 360) {
		appendEclipseMapPolygonPath(&builder, frame, polygon)
	}
	return builder.String()
}

func lunarEclipseVisibilityPolygons(
	center svgmap.GeoPoint,
	projection svgmap.Projection,
	samples int,
) [][]svgmap.GeoPoint {
	return svgmap.VisibleHemispherePolygons(center, projection, samples)
}

func lunarEclipseSubpoint(value time.Time) svgmap.GeoPoint {
	ttJDE := timeToTTJDE(value)
	ra, dec := basic.HMoonTrueRaDec(ttJDE)
	utJDE := basic.TD2UT(ttJDE, false)
	longitude := normalizeDegree180(ra - basic.ApparentSiderealTime(utJDE)*15)
	return svgmap.GeoPoint{Longitude: longitude, Latitude: dec}
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
	dash, clipID string,
) {
	for _, segment := range svgmap.PolylineSegments(points, frame.Projection) {
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
		fmt.Fprintf(builder, `" clip-path="url(#%s)" fill="none" stroke="%s" stroke-width="%.2f" stroke-dasharray="%s" stroke-linecap="round"/>`,
			clipID, color, strokeWidth, dash)
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

func writeLunarEclipseMapLegend(builder *strings.Builder, frame svgmap.Frame, language string) {
	labels := []string{"全程可见", "带食月出", "带食月落", "不可见"}
	colors := []string{"#5e846d", "#4e9da0", "#e2aa4b", "#747b7d"}
	if language == "en" {
		labels = []string{"Entire eclipse", "Moonrise during eclipse", "Moonset during eclipse", "Not visible"}
	}
	y := frame.Y + frame.Height + 31
	itemWidth := frame.Width / 4
	for index, label := range labels {
		x := frame.X + float64(index)*itemWidth
		fmt.Fprintf(builder, `<rect x="%.3f" y="%.3f" width="18" height="9" fill="%s" fill-opacity="0.78"/>`, x, y-8, colors[index])
		fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#465053" font-family="Arial, sans-serif" font-size="10">%s</text>`,
			x+24, y, html.EscapeString(label))
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
	fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#596164" font-family="Georgia, 'Times New Roman', serif" font-size="11">%s</text>`,
		frame.X, float64(options.Height)-38, html.EscapeString(text))
}

func normalizeDegree180(value float64) float64 {
	value = math.Mod(value+180, 360)
	if value < 0 {
		value += 360
	}
	return value - 180
}
