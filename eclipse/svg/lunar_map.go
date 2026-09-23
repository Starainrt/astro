package svg

import (
	"github.com/starainrt/astro/internal/timenote"
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro"
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
	// TimeScale 选择图中时刻的时标：零值 UTC；TimeScaleUT1 改用 UT1 时刻，此时 Location 必须是
	// nil 或 UTC（否则返回 false），并在图注里声明尺度。
	// TimeScale selects the label scale: the zero value is UTC; TimeScaleUT1 uses UT1 labels, requires
	// Location to be nil or UTC (otherwise the renderer returns false) and declares the scale in the footer.
	TimeScale astro.TimeScale
	// Projection 选择地图投影；零值使用等经纬投影，不支持的值使渲染器返回 false。
	// Projection selects the map projection. The zero value selects the equirectangular projection; unsupported values make the renderer return false.
	Projection EclipseMapProjection
	// DisablePenumbralPhase 关闭半影阶段，回到只按 P1/P4 分区的四类可见区：零值画半影（补画 U1–U4
	// 地平边界、单独着色只看得见半影的月出与月落带「半影月出」「半影月落」，并在时刻行里标出本影接触）。
	// DisablePenumbralPhase falls back to the four-way P1/P4-only partition: the zero value draws the
	// penumbral phase, adding the U1-U4 horizons, the penumbra-only moonrise and moonset bands as
	// separate legend entries, and the umbral contact times.
	DisablePenumbralPhase bool
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
	if options.TimeScale == astro.TimeScaleUT1 {
		if options.Location != nil && options.Location != time.UTC {
			return "", false
		}
		options.Location = time.UTC
	}
	options = normalizeLunarEclipseMapSVGOptions(date, options)
	projection := internalEclipseMapProjection(options.Projection)
	if projection == "" {
		projection = svgmap.ProjectionEquirectangular
	}
	return renderLunarEclipseMapSVG(lunarEclipseDisplayInfo(info, options.TimeScale), info, options, projection), true
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
	geometry eclipsecore.LunarEclipseInfo,
	options LunarEclipseMapSVGOptions,
	projection svgmap.Projection,
) string {
	frame := eclipseMapFrame(options.Width, options.Height, projection, 142, 92)
	if projection == svgmap.ProjectionOrthographic {
		// 正射视点取食甚时刻的月下点：视点固定 (0°,0°) 会把半个可见半球让给与本次月食无关的区域。
		frame.CenterLongitude, frame.CenterLatitude = lunarEclipseOrthographicCenter(geometry.Maximum)
	}
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
	penumbralBands := lunarEclipseMapDrawsPenumbralBands(info, !options.DisablePenumbralPhase)
	builder.WriteString(`<defs>`)
	writeLunarEclipseMapDefinitions(&builder, frame, geometry, penumbralBands)
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

	writeLunarEclipseMapRegions(&builder, frame, geometry, penumbralBands)
	frame.WriteFrame(&builder)
	legendRows := lunarEclipseMapLegendRows(frame, float64(options.Width), options.Language, penumbralBands)
	writeLunarEclipseMapLegend(&builder, frame, float64(options.Width), options.Language, penumbralBands)
	writeLunarEclipseMapFooter(&builder, frame, options, projection, info.Maximum, len(legendRows))
	builder.WriteString(`</svg>`)
	return builder.String()
}

// lunarEclipseMapDrawsPenumbralBands 只在该场月食确有本影阶段且开关打开时才区分半影带。
func lunarEclipseMapDrawsPenumbralBands(info eclipsecore.LunarEclipseInfo, includePenumbral bool) bool {
	return includePenumbral && info.HasPartial
}

// 可见性由 P1、食甚、P4 三刻的月球地平状态共同界定：全程可见要求三刻都在地平上，
// 只在食甚前后露出地平的地点仍看得见本影食，算带食月落；三刻都不见才算不可见。
func writeLunarEclipseVisibilityMasks(builder *strings.Builder, frame svgmap.Frame) {
	writeLunarEclipseVisibilityMask(builder, "lunar-not-visible-mask", frame,
		"lunar-visible-p1-shape", "lunar-visible-p4-shape", "lunar-visible-maximum-shape")
	writeLunarEclipseVisibilityMask(builder, "lunar-not-p1-p4-mask", frame,
		"lunar-visible-p1-shape", "lunar-visible-p4-shape")
	writeLunarEclipseVisibilityMask(builder, "lunar-not-maximum-mask", frame, "lunar-visible-maximum-shape")
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
	points := basic.MoonHorizon(basic.Date2JD(value.UTC()), 360)
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
	closedBoundary := append(append([]svgmap.GeoPoint(nil), boundary...), boundary[0])
	return lunarEclipseGeoPolygonPath(boundary, frame), closedBoundary
}

// lunarEclipseGeoPolygonPath 把经纬度环投影成 SVG 面路径，反经线与画布边缘按共享片段切开。
func lunarEclipseGeoPolygonPath(boundary []svgmap.GeoPoint, frame svgmap.Frame) string {
	var builder strings.Builder
	for _, fragment := range svgmap.PolygonFragments(boundary, frame.Clip()) {
		if frame.IsPolar() {
			fragment = lunarEclipsePolarHorizonClosure(fragment)
		}
		appendEclipseMapPolygonPathConsistent(&builder, frame, fragment)
	}
	return builder.String()
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
	if line := lunarEclipseMapUmbralContactLine(info, options); line != "" {
		fmt.Fprintf(builder, `<text x="%.3f" y="100" fill="#293235" font-family="Arial, sans-serif" font-size="13" text-anchor="middle">%s</text>`,
			float64(options.Width)/2, html.EscapeString(line))
	}
}

// lunarEclipseMapUmbralContactLine 给出本影接触时刻行；开关关闭或没有本影阶段时返回空串。
func lunarEclipseMapUmbralContactLine(info eclipsecore.LunarEclipseInfo, options LunarEclipseMapSVGOptions) string {
	if options.DisablePenumbralPhase || !info.HasPartial {
		return ""
	}
	parts := []string{"U1 " + info.PartialStart.In(options.Location).Format("15:04:05")}
	if info.HasTotal {
		parts = append(parts,
			"U2 "+info.TotalStart.In(options.Location).Format("15:04:05"),
			"U3 "+info.TotalEnd.In(options.Location).Format("15:04:05"))
	}
	parts = append(parts, "U4 "+info.PartialEnd.In(options.Location).Format("15:04:05"))
	prefix := "本影阶段 "
	if options.Language == "en" {
		prefix = "Umbral phases "
	}
	return prefix + strings.Join(parts, " | ")
}

// lunarEclipseMapLegendRows 给出可见性图例的标注位次：每行按实测文本宽度排布并居中在地图框下方。
func lunarEclipseMapLegendRows(frame svgmap.Frame, canvasWidth float64, language string, penumbralBands bool) [][]solarEclipseCardinalLabel {
	labels := []string{"全程可见", "带食月出", "带食月落", "不可见"}
	if language == "en" {
		labels = []string{"Entire eclipse", "Moonrise during eclipse", "Moonset during eclipse", "Not visible"}
	}
	if penumbralBands {
		moonrise, moonset := "半影月出", "半影月落"
		if language == "en" {
			moonrise, moonset = "Penumbra moonrise", "Penumbra moonset"
		}
		labels = []string{labels[0], labels[1], moonrise, labels[2], moonset, labels[3]}
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

func writeLunarEclipseMapLegend(builder *strings.Builder, frame svgmap.Frame, canvasWidth float64, language string, penumbralBands bool) {
	colors := []string{"#5e846d", "#4e9da0", "#e2aa4b", "#747b7d"}
	if penumbralBands {
		colors = []string{"#5e846d", "#4e9da0", "#7488cc", "#e2aa4b", "#ab7fb5", "#747b7d"}
	}
	index := 0
	for _, row := range lunarEclipseMapLegendRows(frame, canvasWidth, language, penumbralBands) {
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
	instant time.Time,
	legendRows int,
) {
	text := options.FooterNote
	if text == "" {
		if options.Language == "en" {
			text = eclipseMapProjectionLabel(projection, "en") + "; P1/P4 Moon-visible hemispheres; Natural Earth 1:50m physical land, no administrative boundaries."
		} else {
			text = eclipseMapProjectionLabel(projection, "zh") + "；按 P1/P4 月球可见半球分区；Natural Earth 1:50m 物理陆地底图，不含行政边界。"
		}
	}
	// 页脚末行要落在白色图框内：图框在画布内缩 18px，11px 字的末行基线留 26px（框线 + 间隙 + 字下伸部）。
	const footerBottomPadding = 28.0
	// 首行基线再让出一个字下伸部，末行才刚好落在留白之上（BaselineLineLimit 按行下沿计行）。
	baseline := float64(options.Height) - footerBottomPadding - svgchart.FooterLineHeight(11) - 4
	// 图例换成多行时页脚要跟着下移，否则窄画布上两行文字会叠在一起；下行仍留在图框内。
	if legendRows > 1 {
		if shifted := frame.Y + frame.Height + 31 + float64(legendRows-1)*lunarEclipseDetailedMapLegendLineStep + 15; shifted > baseline {
			baseline = math.Min(shifted, float64(options.Height)-footerBottomPadding)
		}
	}
	// 默认说明是单行；调用方文本按图框宽度折行，行数按画布底边截断，首行位置不变。
	note := timenote.Scale(options.TimeScale, instant, options.Location, options.Language)
	lines := []string{text}
	if options.FooterNote != "" {
		lines = svgchart.WrapText(options.FooterNote, float64(options.Width)-frame.X-24, 11)
	}
	lines = svgchart.FooterLinesWithScale(lines, note, float64(options.Width)-frame.X-24, 11,
		svgchart.BaselineLineLimit(11, svgchart.FooterLineHeight(11), baseline,
			float64(options.Height)-footerBottomPadding))
	for index, line := range lines {
		fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#596164" font-family="Georgia, 'Times New Roman', serif" font-size="11">%s</text>`,
			frame.X, baseline+float64(index)*svgchart.FooterLineHeight(11), html.EscapeString(line))
	}
}

func normalizeDegree180(value float64) float64 {
	value = math.Mod(value+180, 360)
	if value < 0 {
		value += 360
	}
	return value - 180
}

// lunarEclipseOrthographicCenter 给出月食正射图的视点：食甚时刻的月下点（月球位于天顶的地点）。
func lunarEclipseOrthographicCenter(at time.Time) (float64, float64) {
	jd := basic.Date2JD(at.UTC())
	tt := basic.UTC2TT(jd)
	ra, dec := basic.HMoonTrueRaDec(tt)
	return normalizeDegree180(ra - basic.ApparentSiderealTime(basic.TT2UT1(tt))*15), dec
}

// writeLunarEclipseMapDefinitions 输出四类可见性区域所需的形状与掩膜定义。
// 独立全球图和 详细版式组合图共用，形状按传入的图框重建。
func writeLunarEclipseMapDefinitions(b *strings.Builder, frame svgmap.Frame, info eclipsecore.LunarEclipseInfo, penumbralBands bool) {
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
	if penumbralBands {
		writeLunarEclipseUmbralVisibilityShapes(b, frame, info)
	}
}

// writeLunarEclipseUmbralVisibilityShapes 输出 U1、U4 的月球可见半球形状，以及"本影阶段任意时刻
// 在地平上"的取样半球、裁剪路径与排除掩膜；底带取并集内部，半影带取外部。
func writeLunarEclipseUmbralVisibilityShapes(b *strings.Builder, frame svgmap.Frame, info eclipsecore.LunarEclipseInfo) {
	u1Path, _ := lunarEclipseVisibilityPath(info.PartialStart, frame)
	u4Path, _ := lunarEclipseVisibilityPath(info.PartialEnd, frame)
	fmt.Fprintf(b, `<path id="lunar-visible-u1-shape" d="%s"/>`, u1Path)
	fmt.Fprintf(b, `<path id="lunar-visible-u4-shape" d="%s"/>`, u4Path)
	b.WriteString(`<clipPath id="lunar-visible-u1"><use href="#lunar-visible-u1-shape"/></clipPath>`)
	b.WriteString(`<clipPath id="lunar-visible-u4"><use href="#lunar-visible-u4-shape"/></clipPath>`)
	writeLunarEclipseUmbralVisibleShapes(b, frame, info)
}

// writeLunarEclipseUmbralVisibleShapes 输出"本影阶段任意时刻在地平上"的取样半球：本影区间按固定档位
// 取样，每一刻的可见半球都是一个精确的圆盘（按经纬度插值会把近乎子午线方向的边界切进真实区域——
// 实测 1932-03-22 西经 75.5° 一带上边界 0.5° 经度内从 +19.9° 掉到 −27.2°，1° 列的弦会切掉十几度）。
// 这些形状同时用于底带的裁剪（取并集）与半影带的掩膜（取补集），两侧严丝合缝。
func writeLunarEclipseUmbralVisibleShapes(b *strings.Builder, frame svgmap.Frame, info eclipsecore.LunarEclipseInfo) {
	instants := lunarEclipseUmbralVisibilityInstants(info)
	if len(instants) == 0 {
		// 兜底：算不出本影可见区时至少挖掉两端接触（宁可少扣，也不让底带无裁剪地铺满）。
		b.WriteString(`<clipPath id="lunar-visible-umbral"><use href="#lunar-visible-u1-shape"/><use href="#lunar-visible-u4-shape"/></clipPath>`)
		writeLunarEclipseVisibilityMask(b, "lunar-not-umbral-visible-mask", frame,
			"lunar-visible-u1-shape", "lunar-visible-u4-shape")
		return
	}
	// 两端复用已经写好的精细半球形状，中间的取样用粗轮廓：单个圆盘在球面上的偏差约 0.035°，
	// 在 1200 像素宽的世界图上不到 0.1 像素，而每个形状能省下三分之二的字节。
	ids := []string{"lunar-visible-u1-shape"}
	for index := 1; index < len(instants)-1; index++ {
		id := fmt.Sprintf("lunar-visible-umbral-%d", index)
		fmt.Fprintf(b, `<path id="%s" d="%s"/>`, id, lunarEclipseMaskHorizonPath(instants[index], frame))
		ids = append(ids, id)
	}
	ids = append(ids, "lunar-visible-u4-shape")
	var clip strings.Builder
	clip.WriteString(`<clipPath id="lunar-visible-umbral">`)
	for _, id := range ids {
		fmt.Fprintf(&clip, `<use href="#%s"/>`, id)
	}
	clip.WriteString(`</clipPath>`)
	b.WriteString(clip.String())
	writeLunarEclipseVisibilityMask(b, "lunar-not-umbral-visible-mask", frame, ids...)
}

// lunarEclipseMaskHorizonPath 掩膜用的可见半球轮廓：90 点基础圆，等距圆柱投影下再按 0.2° 容差
// 细化，避免极区一段地平弧投影后横跨近 180° 经度、弦切出地平线以外的假可见帽。
func lunarEclipseMaskHorizonPath(value time.Time, frame svgmap.Frame) string {
	points := basic.MoonHorizon(basic.Date2JD(value.UTC()), 90)
	boundary := make([]svgmap.GeoPoint, len(points))
	for index, point := range points {
		boundary[index] = svgmap.GeoPoint{Longitude: point[0], Latitude: point[1]}
	}
	if frame.Projection == svgmap.ProjectionEquirectangular {
		boundary = lunarhorizon.RefineWithin(boundary, value, 0.2)
	}
	return lunarEclipseGeoPolygonPath(boundary, frame)
}

// lunarEclipseUmbralVisibilityInstants 本影区间按 15 分钟档位取样，至少两端、至多 12 刻：中间取样用
// 粗轮廓，圆盘分辨率约 0.035°，合起来残余的掠射窗口深度约 0.05°（约 0.15 像素）；要逐点精确判定
// 请用站点 API 或 GeoJSON。
func lunarEclipseUmbralVisibilityInstants(info eclipsecore.LunarEclipseInfo) []time.Time {
	const (
		step  = 15 * time.Minute
		limit = 12
	)
	duration := info.PartialEnd.Sub(info.PartialStart)
	if duration <= 0 {
		return nil
	}
	count := int(duration/step) + 1
	if count < 2 {
		count = 2
	}
	if count > limit {
		count = limit
	}
	instants := make([]time.Time, count)
	for index := range instants {
		instants[index] = info.PartialStart.Add(duration * time.Duration(index) / time.Duration(count-1))
	}
	return instants
}

func writeLunarEclipseMapRegions(b *strings.Builder, frame svgmap.Frame, info eclipsecore.LunarEclipseInfo, penumbralBands bool) {
	_, startBoundary := lunarEclipseVisibilityPath(info.PenumbralStart, frame)
	_, endBoundary := lunarEclipseVisibilityPath(info.PenumbralEnd, frame)
	frame.WriteOcean(b)
	frame.WriteGraticule(b, "lunar-map-clip")
	frame.WriteLand(b, "lunar-map-clip")
	fmt.Fprintf(b, `<g class="lunar-visibility-regions" data-source="%s" clip-path="url(#lunar-map-clip)">`, eclipseMapSourceLunarVisibilityRegions)
	writeLunarEclipseUnavailableRegion(b, frame)
	writeLunarEclipseBaseRegion(b, penumbralBands, "lunar-visible-umbral",
		`<use class="visible-at-p1 moonset-region" mask="url(#lunar-not-p4-mask)" href="#lunar-visible-p1-shape" fill="#e2aa4b" fill-opacity="0.34"/>`)
	writeLunarEclipseBaseRegion(b, penumbralBands, "lunar-visible-umbral",
		`<use class="visible-at-p4 moonrise-region" mask="url(#lunar-not-p1-mask)" href="#lunar-visible-p4-shape" fill="#4e9da0" fill-opacity="0.34"/>`)
	b.WriteString(`<g clip-path="url(#lunar-visible-p4)"><g clip-path="url(#lunar-visible-maximum)"><use class="entire-eclipse-region" href="#lunar-visible-p1-shape" fill="#5e846d" fill-opacity="0.34"/></g></g>`)
	// 两端可见、食甚却在地平下的高纬月落-月出带：两端可见推不出全程可见，归入带食月落。
	b.WriteString(`<g clip-path="url(#lunar-visible-p4)"><g mask="url(#lunar-not-maximum-mask)"><use class="visible-at-p1-below-maximum moonset-region" href="#lunar-visible-p1-shape" fill="#e2aa4b" fill-opacity="0.34"/></g></g>`)
	// 两端都在地平下、只有食甚前后在地平上的极区透镜：月落必然落在食甚之后，与带食月落同类着色。
	b.WriteString(`<g mask="url(#lunar-not-p1-p4-mask)"><use class="eclipse-at-maximum moonset-region" href="#lunar-visible-maximum-shape" fill="#e2aa4b" fill-opacity="0.34"/></g>`)
	if penumbralBands {
		writeLunarEclipsePenumbralRegions(b)
	}
	b.WriteString(`</g>`)
	writeEclipseMapGeoLine(b, frame, startBoundary, "p1-horizon", "#a56c16", 1.2, "4 3", "lunar-map-clip", eclipseMapSourceLunarHorizonBoundaries)
	writeEclipseMapGeoLine(b, frame, endBoundary, "p4-horizon", "#197a82", 1.2, "4 3", "lunar-map-clip", eclipseMapSourceLunarHorizonBoundaries)
	if penumbralBands {
		writeLunarEclipseUmbralHorizonLines(b, frame, info)
	}
}

// writeLunarEclipseBaseRegion 画一条月出/月落底带；半影带开启时只保留"本影阶段真的能看见"的部分，
// 半影段交给半影带：两条带用同一区域的正反面裁切，既不叠色也不留缝。
func writeLunarEclipseBaseRegion(b *strings.Builder, penumbralBands bool, clip, element string) {
	if penumbralBands {
		fmt.Fprintf(b, `<g clip-path="url(#%s)">%s</g>`, clip, element)
		return
	}
	b.WriteString(element)
}

// writeLunarEclipsePenumbralRegions 只着色只看得见半影的两条带（月落侧与月出侧）：主端在地平上，
// 本影阶段整段在地平下，再扣掉另一端的可见区（极区下中天会让两端同时可见）。
func writeLunarEclipsePenumbralRegions(b *strings.Builder) {
	b.WriteString(`<g mask="url(#lunar-not-umbral-visible-mask)"><use class="penumbra-only-region penumbra-moonset-region" mask="url(#lunar-not-p4-mask)" href="#lunar-visible-p1-shape" fill="#ab7fb5" fill-opacity="0.34"/></g>`)
	b.WriteString(`<g mask="url(#lunar-not-umbral-visible-mask)"><use class="penumbra-only-region penumbra-moonrise-region" mask="url(#lunar-not-p1-mask)" href="#lunar-visible-p4-shape" fill="#7488cc" fill-opacity="0.34"/></g>`)
}

// writeLunarEclipseUmbralHorizonLines 补画本影接触点的地平边界，边界之间就是各阶段被月出月落切掉的区间。
func writeLunarEclipseUmbralHorizonLines(b *strings.Builder, frame svgmap.Frame, info eclipsecore.LunarEclipseInfo) {
	type horizonContact struct {
		class string
		at    time.Time
	}
	contacts := []horizonContact{{"u1-horizon", info.PartialStart}}
	if info.HasTotal {
		contacts = append(contacts, horizonContact{"u2-horizon", info.TotalStart}, horizonContact{"u3-horizon", info.TotalEnd})
	}
	contacts = append(contacts, horizonContact{"u4-horizon", info.PartialEnd})
	for _, contact := range contacts {
		if contact.at.IsZero() {
			continue
		}
		_, boundary := lunarEclipseVisibilityPath(contact.at, frame)
		writeEclipseMapGeoLine(b, frame, boundary, contact.class, "#5f4b8b", 1.1, "3 3", "lunar-map-clip", eclipseMapSourceLunarHorizonBoundaries)
	}
}

// lunarEclipseInfoUT1Labels 把民用时刻换成同一物理时刻的 UT1 时刻。零值 time.Time 表示该阶段
