// Package svg 生成自包含的月掩星图。
// Package svg renders self-contained lunar-occultation diagrams.
package svg

import (
	"errors"
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro/internal/svgmap"
	"github.com/starainrt/astro/moon"
)

const (
	starOccultationSVGDefaultWidth  = 1200
	starOccultationSVGDefaultHeight = 800
	starOccultationSVGMinimumWidth  = 480
	starOccultationSVGMinimumHeight = 360
	starOccultationSVGDefaultZone   = 8 * 60 * 60

	starOccultationSVGLanguageChinese = "zh"
	starOccultationSVGLanguageEnglish = "en"
)

// ErrInvalidStarOccultationSVGOptions 表示画布选项无法容纳地图、事件面板和页脚而不发生重叠。
// ErrInvalidStarOccultationSVGOptions reports canvas options that cannot hold the map, event panel, and footer without overlap.
var ErrInvalidStarOccultationSVGOptions = errors.New("invalid stellar occultation SVG options")

// StarOccultationSVGOptions 控制恒星月掩全球掩带 SVG 输出。
// StarOccultationSVGOptions controls a global stellar-occultation path SVG.
type StarOccultationSVGOptions struct {
	// Width 和 Height 是 SVG 画布尺寸；非正值使用 1200x800 全球地图默认值。
	// Width and Height are the SVG canvas dimensions. Values <= 0 use the 1200x800 global-map default.
	Width  int
	Height int
	// Title 及后续文本字段覆盖自动生成的标签。
	// Title and the following text fields override automatically generated labels.
	Title         string
	SummaryText   string
	GreatestText  string
	MapTitle      string
	ContactsTitle string
	FooterNote    string
	// Language 为 "en" 时使用英文，其他值使用中文。
	// Language uses English for "en" and Chinese for every other value.
	Language string
	// Location 控制显示的事件时刻；nil 使用 UTC+8。
	// Location controls displayed event times. Nil uses UTC+8.
	Location *time.Location
	// Projection 选择全球地图投影；零值会为限于单半球的高纬事件自动使用极区地图。
	// Projection selects the global map projection. The zero value automatically uses a polar map for a high-latitude event confined to one hemisphere.
	Projection MapProjection
	// TimeLabelStep 控制可见中心线上的 HH:MM 标签；零值使用 30 分钟，负值禁用标签。
	// TimeLabelStep controls HH:MM labels along the visible center line. Zero uses 30 minutes; a negative value disables the labels.
	TimeLabelStep time.Duration
}

// FindStarOccultationSVGs 搜索时间窗并渲染其中所有恒星月掩全球掩带。
// 使用调用者提供的坐标，不加载内嵌星表；无错误的空切片表示没有事件。
// FindStarOccultationSVGs searches a time window and renders every global path it finds. It uses the coordinate supplied by the caller and does not load the embedded star catalog. An empty slice without an error means no event.
func FindStarOccultationSVGs(
	start, end time.Time,
	star moon.StarCoordinate,
	pathOptions moon.OccultationPathOptions,
	options StarOccultationSVGOptions,
) ([]string, error) {
	paths, err := moon.FindStarOccultationPaths(start, end, star, pathOptions)
	if err != nil {
		return nil, err
	}
	diagrams := make([]string, 0, len(paths))
	for _, path := range paths {
		diagram, renderErr := StarOccultationPathSVG(path, options)
		if renderErr != nil {
			return nil, renderErr
		}
		diagrams = append(diagrams, diagram)
	}
	return diagrams, nil
}

// StarOccultationPathSVG 将已计算的恒星月掩全球路径渲染为 SVG。
// 不会运行事件搜索，也不会加载内嵌星表。
// StarOccultationPathSVG renders an already computed global stellar-occultation path. It does not run the event search or load the embedded star catalog.
func StarOccultationPathSVG(
	path moon.StarOccultationPath,
	options StarOccultationSVGOptions,
) (string, error) {
	if err := validateStarOccultationPath(path); err != nil {
		return "", err
	}
	if err := validateStarOccultationSVGOptions(options); err != nil {
		return "", err
	}
	options = normalizeStarOccultationSVGOptions(options)
	return renderStarOccultationPathSVG(path, options), nil
}

func validateStarOccultationSVGOptions(options StarOccultationSVGOptions) error {
	if options.Width > 0 && options.Width < starOccultationSVGMinimumWidth {
		return fmt.Errorf("%w: width must be zero or at least %d", ErrInvalidStarOccultationSVGOptions, starOccultationSVGMinimumWidth)
	}
	if options.Height > 0 && options.Height < starOccultationSVGMinimumHeight {
		return fmt.Errorf("%w: height must be zero or at least %d", ErrInvalidStarOccultationSVGOptions, starOccultationSVGMinimumHeight)
	}
	switch options.Projection {
	case MapProjectionAuto, MapProjectionEquirectangular, MapProjectionNorthPolar, MapProjectionSouthPolar:
	default:
		return fmt.Errorf("%w: unsupported map projection %q", ErrInvalidStarOccultationSVGOptions, options.Projection)
	}
	return nil
}

func normalizeStarOccultationSVGOptions(options StarOccultationSVGOptions) StarOccultationSVGOptions {
	if options.Width <= 0 {
		options.Width = starOccultationSVGDefaultWidth
	}
	if options.Height <= 0 {
		options.Height = starOccultationSVGDefaultHeight
	}
	if options.Location == nil {
		options.Location = time.FixedZone("UTC+8", starOccultationSVGDefaultZone)
	}
	if strings.EqualFold(options.Language, starOccultationSVGLanguageEnglish) {
		options.Language = starOccultationSVGLanguageEnglish
	} else {
		options.Language = starOccultationSVGLanguageChinese
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

func renderStarOccultationPathSVG(path moon.StarOccultationPath, options StarOccultationSVGOptions) string {
	return renderOccultationPathSVG(path, nil, options)
}

func renderOccultationPathSVG(
	path moon.StarOccultationPath,
	planetPath *moon.PlanetOccultationPath,
	options StarOccultationSVGOptions,
) string {
	projection := resolveStarOccultationMapProjection(path, options.Projection)
	options.Projection = MapProjection(projection)
	title := starOccultationSVGTitle(path, options)
	headerLines := starOccultationSVGHeaderLines(path, options)
	headerBottom := 72.0 + float64(len(headerLines))*19
	layout := starOccultationSVGLayoutFor(options, headerBottom, projection)

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" preserveAspectRatio="xMidYMid meet" style="max-width:100%%;height:auto;display:block" role="img" aria-label="%s">`,
		options.Width, options.Height, options.Width, options.Height, html.EscapeString(title))
	b.WriteString(`<defs>`)
	b.WriteString(layout.mapFrame().ClipDefinition("occultation-map-clip"))
	b.WriteString(`</defs>`)
	b.WriteString(`<rect width="100%" height="100%" fill="#efefed"/>`)
	fmt.Fprintf(&b, `<rect x="22" y="18" width="%.3f" height="%.3f" fill="#ffffff" stroke="#c9c9c6" stroke-width="1.2"/>`,
		layout.width-44, layout.height-36)
	fmt.Fprintf(&b, `<text x="%.3f" y="44" fill="#111111" font-family="Georgia, 'Times New Roman', serif" font-size="%d" font-weight="700" text-anchor="middle">%s</text>`,
		layout.width/2, starOccultationTitleFontSize(title, layout.width), html.EscapeString(title))
	fmt.Fprintf(&b, `<line x1="%.3f" y1="57" x2="%.3f" y2="57" stroke="#555" stroke-width="1"/>`, layout.width/2-78, layout.width/2+78)
	for index, line := range headerLines {
		fontSize := 13
		fill := "#3b4143"
		if index == 0 {
			fontSize = 14
			fill = "#222222"
		}
		fmt.Fprintf(&b, `<text x="%.3f" y="%.3f" fill="%s" font-family="Georgia, 'Times New Roman', serif" font-size="%d" text-anchor="middle">%s</text>`,
			layout.width/2, 82+float64(index)*19, fill, fontSize, html.EscapeString(line))
	}

	fmt.Fprintf(&b, `<text x="%.3f" y="%.3f" fill="#161a1b" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700">%s</text>`,
		layout.mapX, layout.mapY-10, html.EscapeString(starOccultationSVGMapTitle(options)))
	if planetPath == nil {
		writeStarOccultationMap(&b, path, layout, options)
		writeStarOccultationEventsPanel(&b, path, layout, options)
	} else {
		writePlanetOccultationMap(&b, *planetPath, path, layout, options)
		writePlanetOccultationEventsPanel(&b, *planetPath, layout, options)
	}
	writeStarOccultationFooter(&b, layout, options)
	b.WriteString(`</svg>`)
	return b.String()
}

func writeStarOccultationMap(
	b *strings.Builder,
	path moon.StarOccultationPath,
	layout starOccultationSVGLayout,
	options StarOccultationSVGOptions,
) {
	layout.mapFrame().WriteOcean(b)
	writeStarOccultationGraticule(b, layout)
	writeStarOccultationLand(b, layout)
	writeStarOccultationBand(b, path, layout)
	writeStarOccultationGeoLine(b, path.NorthernLimit, layout, "northern-limit", "#a66f18", 1.35, "")
	writeStarOccultationGeoLine(b, path.SouthernLimit, layout, "southern-limit", "#a66f18", 1.35, "")
	writeStarOccultationGeoLine(b, path.CenterLine, layout, "center-line", "#59676b", 1.6, "5 4")
	writeStarOccultationVisibleCenterLine(b, path.CenterLine, layout)
	writeOccultationTimeMarkers(b, path.CenterLine, layout, options,
		[]time.Time{path.Start.Time, path.End.Time}, path.Greatest.Time)
	writeStarOccultationEventMarker(b, path.Start, layout, starOccultationEventLabel("start", options.Language), "start")
	writeStarOccultationEventMarker(b, path.Greatest, layout, starOccultationEventLabel("greatest", options.Language), "greatest")
	writeStarOccultationEventMarker(b, path.End, layout, starOccultationEventLabel("end", options.Language), "end")
	layout.mapFrame().WriteFrame(b)
	writeStarOccultationLegend(b, layout, options.Language)
}

func writeStarOccultationGraticule(b *strings.Builder, layout starOccultationSVGLayout) {
	layout.mapFrame().WriteGraticule(b, "occultation-map-clip")
	if layout.projection != svgmap.ProjectionEquirectangular {
		return
	}
	for _, longitude := range []float64{-120, -60, 0, 60, 120} {
		x, _ := layout.project(longitude, 0)
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#687577" font-family="Arial, sans-serif" font-size="9" text-anchor="middle">%.0f°</text>`,
			x, layout.mapY+layout.mapHeight+13, longitude)
	}
	for _, latitude := range []float64{-60, -30, 0, 30, 60} {
		_, y := layout.project(0, latitude)
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#687577" font-family="Arial, sans-serif" font-size="9" text-anchor="end">%.0f°</text>`,
			layout.mapX-5, y+3, latitude)
	}
}

func writeStarOccultationLand(b *strings.Builder, layout starOccultationSVGLayout) {
	layout.mapFrame().WriteLand(b, "occultation-map-clip")
}

func writeStarOccultationBand(b *strings.Builder, path moon.StarOccultationPath, layout starOccultationSVGLayout) {
	writeOccultationBand(b, path.NorthernLimit, path.SouthernLimit, layout,
		"occultation-band-layer", "occultation-band", "#e0ae43", 0.28)
}

func writeOccultationBand(
	b *strings.Builder,
	northern, southern []moon.OccultationPathPoint,
	layout starOccultationSVGLayout,
	layerClass, pathClass, color string,
	opacity float64,
) {
	segments := starOccultationBandFragments(northern, southern, layout.projection)
	if len(segments) == 0 {
		return
	}
	fmt.Fprintf(b, `<g class="%s" clip-path="url(#occultation-map-clip)" fill="%s" fill-opacity="%.2f" stroke="none">`,
		layerClass, color, opacity)
	for _, segment := range segments {
		fmt.Fprintf(b, `<path class="%s" d="`, pathClass)
		for index, point := range segment {
			x, y := layout.project(point.longitude, point.latitude)
			command := "L"
			if index == 0 {
				command = "M"
			}
			fmt.Fprintf(b, `%s %.3f %.3f `, command, x, y)
		}
		b.WriteString(`Z"/>`)
	}
	b.WriteString(`</g>`)
}

func writeStarOccultationGeoLine(
	b *strings.Builder,
	points []moon.OccultationPathPoint,
	layout starOccultationSVGLayout,
	className, color string,
	strokeWidth float64,
	dash string,
) {
	for _, segment := range starOccultationPathSegmentsForProjection(points, layout.projection) {
		if len(segment) < 2 {
			continue
		}
		fmt.Fprintf(b, `<path class="%s" d="`, className)
		for index, point := range segment {
			x, y := layout.project(point.Longitude, point.Latitude)
			command := "L"
			if index == 0 {
				command = "M"
			}
			fmt.Fprintf(b, `%s %.3f %.3f `, command, x, y)
		}
		fmt.Fprintf(b, `" clip-path="url(#occultation-map-clip)" fill="none" stroke="%s" stroke-width="%.2f"`, color, strokeWidth)
		if dash != "" {
			fmt.Fprintf(b, ` stroke-dasharray="%s"`, dash)
		}
		b.WriteString(` stroke-linecap="round" stroke-linejoin="round"/>`)
	}
}

func writeStarOccultationVisibleCenterLine(
	b *strings.Builder,
	points []moon.OccultationPathPoint,
	layout starOccultationSVGLayout,
) {
	visible := make([]moon.OccultationPathPoint, 0, len(points))
	flush := func() {
		writeStarOccultationGeoLine(b, visible, layout, "visible-center-line", "#087f8c", 2.8, "")
		visible = visible[:0]
	}
	for _, point := range points {
		if point.MoonAltitude <= 0 {
			flush()
			continue
		}
		if len(visible) > 0 && math.Abs(point.Longitude-visible[len(visible)-1].Longitude) > 180 {
			flush()
		}
		visible = append(visible, point)
	}
	flush()
}

func writeStarOccultationEventMarker(
	b *strings.Builder,
	point moon.OccultationPathPoint,
	layout starOccultationSVGLayout,
	label, kind string,
) {
	x, y, visible := layout.mapFrame().Project(point.Longitude, point.Latitude)
	if !visible {
		return
	}
	placement := starOccultationDefaultEventMarkerPlacement(x, y, kind, layout)
	writeStarOccultationProjectedEventMarker(b, x, y, label, kind, placement)
}

type starOccultationEventMarkerPlacement struct {
	labelX, labelY float64
	anchor         string
	leader         bool
}

func starOccultationDefaultEventMarkerPlacement(
	x, y float64,
	kind string,
	layout starOccultationSVGLayout,
) starOccultationEventMarkerPlacement {
	placement := starOccultationEventMarkerPlacement{
		labelX: x - 7,
		labelY: y - 8,
		anchor: "end",
	}
	if kind == "greatest" {
		placement.anchor = "middle"
		placement.labelX = x
		placement.labelY = y - 10
	} else if kind == "end" || kind == "total-end" {
		placement.anchor = "start"
		placement.labelX = x + 7
	}
	placement.labelY = math.Max(layout.mapY+12, math.Min(layout.mapY+layout.mapHeight-5, placement.labelY))
	return placement
}

func writeStarOccultationProjectedEventMarker(
	b *strings.Builder,
	x, y float64,
	label, kind string,
	placement starOccultationEventMarkerPlacement,
) {
	color := "#273d49"
	radius := 4.0
	strokeWidth := 1.4
	if kind == "greatest" {
		color = "#c44336"
		radius = 5.2
	} else if kind == "total-start" || kind == "total-end" {
		color = "#087f8c"
		radius = 3.2
		strokeWidth = 1.0
	}
	fmt.Fprintf(b, `<g class="event-marker event-%s">`, kind)
	if placement.leader {
		leaderX := placement.labelX
		switch placement.anchor {
		case "end":
			leaderX += 3
		case "start":
			leaderX -= 3
		}
		leaderY := placement.labelY + 3
		if placement.labelY > y {
			leaderY = placement.labelY - 12
		}
		fmt.Fprintf(b, `<line class="event-marker-leader" x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#52646b" stroke-width="0.9" vector-effect="non-scaling-stroke"/>`,
			x, y, leaderX, leaderY)
	}
	fmt.Fprintf(b, `<circle cx="%.3f" cy="%.3f" r="%.3f" fill="%s" stroke="#ffffff" stroke-width="%.1f"/>`,
		x, y, radius, color, strokeWidth)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#182124" stroke="#ffffff" stroke-width="3" paint-order="stroke" font-family="Arial, sans-serif" font-size="11" font-weight="700" text-anchor="%s">%s</text></g>`,
		placement.labelX, placement.labelY, placement.anchor, html.EscapeString(label))
}

func writeStarOccultationLegend(b *strings.Builder, layout starOccultationSVGLayout, language string) {
	y := layout.mapY + layout.mapHeight + 30
	labels := []string{"可见中心线", "几何中心线", "掩带边界"}
	if language == starOccultationSVGLanguageEnglish {
		labels = []string{"Visible center line", "Geometric center line", "Occultation limits"}
	}
	available := layout.mapWidth / 3
	styles := []struct {
		color string
		dash  string
	}{
		{"#087f8c", ""},
		{"#59676b", "5 4"},
		{"#a66f18", ""},
	}
	for index, label := range labels {
		x := layout.mapX + float64(index)*available
		fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="%s" stroke-width="2"`, x, y-4, x+22, y-4, styles[index].color)
		if styles[index].dash != "" {
			fmt.Fprintf(b, ` stroke-dasharray="%s"`, styles[index].dash)
		}
		b.WriteString(`/>`)
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#465053" font-family="Arial, sans-serif" font-size="10">%s</text>`,
			x+28, y, html.EscapeString(label))
	}
}

func writeStarOccultationEventsPanel(
	b *strings.Builder,
	path moon.StarOccultationPath,
	layout starOccultationSVGLayout,
	options StarOccultationSVGOptions,
) {
	writeOccultationEventsPanel(b, starOccultationSVGEventRows(path, options.Language), layout, options)
}

func writeOccultationEventsPanel(
	b *strings.Builder,
	rows []starOccultationEventRow,
	layout starOccultationSVGLayout,
	options StarOccultationSVGOptions,
) {
	fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#d0d3d1" stroke-width="1"/>`,
		layout.panelX-10, layout.panelY, layout.panelX-10, layout.panelY+layout.mapHeight)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#161a1b" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700">%s</text>`,
		layout.panelX, layout.panelY+13, html.EscapeString(starOccultationSVGContactsTitle(options)))
	rowTop := layout.panelY + 27
	rowHeight := math.Max(34, (layout.mapHeight-27)/float64(len(rows)))
	for index, row := range rows {
		y := rowTop + float64(index)*rowHeight
		if index > 0 {
			fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#e0e1df" stroke-width="0.8"/>`,
				layout.panelX, y-5, layout.panelX+layout.panelWidth, y-5)
		}
		pointTime := row.point.Time.In(options.Location)
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#1c2528" font-family="Arial, sans-serif" font-size="12" font-weight="700">%s</text>`,
			layout.panelX, y+11, html.EscapeString(row.name))
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#1c2528" font-family="Arial, sans-serif" font-size="11" text-anchor="end">%s</text>`,
			layout.panelX+layout.panelWidth, y+11, html.EscapeString(pointTime.Format("15:04:05.0")))
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#4c585a" font-family="Arial, sans-serif" font-size="10">%s</text>`,
			layout.panelX, y+29, html.EscapeString(starOccultationFormatCoordinates(row.point.Longitude, row.point.Latitude)))
		if rowHeight >= 59 {
			altitudeLabel := "Moon alt."
			if options.Language != starOccultationSVGLanguageEnglish {
				altitudeLabel = "月球高度"
			}
			fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#697476" font-family="Arial, sans-serif" font-size="10">%s %s</text>`,
				layout.panelX, y+45, html.EscapeString(altitudeLabel), html.EscapeString(starOccultationFormatSignedDegree(row.point.MoonAltitude)))
		}
	}
}

func writeStarOccultationFooter(
	b *strings.Builder,
	layout starOccultationSVGLayout,
	options StarOccultationSVGOptions,
) {
	lines := starOccultationWrapText(starOccultationSVGFooter(options), layout.width-80, 11)
	for index, line := range lines {
		fmt.Fprintf(b, `<text x="40" y="%.3f" fill="#596164" font-family="Georgia, 'Times New Roman', serif" font-size="11">%s</text>`,
			layout.footerY+float64(index)*15, html.EscapeString(line))
	}
}

func starOccultationEventLabel(kind, language string) string {
	if language == starOccultationSVGLanguageEnglish {
		switch kind {
		case "start":
			return "Start"
		case "greatest":
			return "Greatest"
		default:
			return "End"
		}
	}
	switch kind {
	case "start":
		return "掩始"
	case "greatest":
		return "掩甚"
	default:
		return "掩终"
	}
}
