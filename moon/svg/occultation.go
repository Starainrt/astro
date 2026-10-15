// Package svg 生成自包含的月掩星图。
// Package svg renders self-contained lunar-occultation diagrams.
package svg

import (
	"github.com/starainrt/astro/internal/limbnote"
	"github.com/starainrt/astro/internal/timenote"
	"errors"
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/internal/occultationgeo"
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/internal/svgmap"
	"github.com/starainrt/astro/moon"
)

const (
	starOccultationSVGDefaultWidth  = 1200
	starOccultationSVGDefaultHeight = 800
	starOccultationSVGMinimumWidth  = 640
	starOccultationSVGMinimumHeight = 480
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
	// Width 和 Height 是 SVG 画布尺寸；非正值使用 1200x800 全球地图默认值，宽度小于 640 或高度小于 480 时返回 ErrInvalidStarOccultationSVGOptions。
	// Width and Height are the SVG canvas dimensions. Values <= 0 use the 1200x800 global-map default;
	// a width below 640 or a height below 480 returns ErrInvalidStarOccultationSVGOptions.
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
	// TimeScale 选择图中时刻的时标：零值 UTC；TimeScaleUT1 改用 UT1 时刻，此时 Location 必须是
	// nil 或 UTC（否则渲染器返回错误），并在图注里声明尺度。
	// TimeScale selects the label scale: the zero value is UTC; TimeScaleUT1 uses UT1 labels, requires
	// Location to be nil or UTC (otherwise the renderer returns an error) and declares the scale.
	TimeScale astro.TimeScale
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
	if options.TimeScale == astro.TimeScaleUT1 && options.Location != nil && options.Location != time.UTC {
		return "", fmt.Errorf("occultation SVG: UT1 labels do not take a non-UTC location")
	}
	if err := validateStarOccultationPath(path); err != nil {
		return "", err
	}
	if err := validateStarOccultationSVGOptions(options); err != nil {
		return "", err
	}
	if options.TimeScale == astro.TimeScaleUT1 {
		options.Location = time.UTC
		path = moon.StarOccultationPathInUT1(path)
	}

	options = normalizeStarOccultationSVGOptions(options)
	diagram, err := renderStarOccultationPathSVG(path, options)
	if err != nil {
		// 几何引擎失败不是数据格式错误，不能裹成 ErrInvalidStarOccultationPath 哨兵。
		return "", fmt.Errorf("stellar occultation SVG band geometry: %w", err)
	}
	return diagram, nil
}

func validateStarOccultationSVGOptions(options StarOccultationSVGOptions) error {
	if options.Width > 0 && options.Width < starOccultationSVGMinimumWidth {
		return fmt.Errorf("%w: width must be zero or at least %d", ErrInvalidStarOccultationSVGOptions, starOccultationSVGMinimumWidth)
	}
	if options.Height > 0 && options.Height < starOccultationSVGMinimumHeight {
		return fmt.Errorf("%w: height must be zero or at least %d", ErrInvalidStarOccultationSVGOptions, starOccultationSVGMinimumHeight)
	}
	switch options.Projection {
	case MapProjectionAuto, MapProjectionEquirectangular, MapProjectionNorthPolar, MapProjectionSouthPolar, MapProjectionOrthographic:
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

func renderStarOccultationPathSVG(path moon.StarOccultationPath, options StarOccultationSVGOptions) (string, error) {
	return renderOccultationPathSVG(path, nil, options, starOccultationSVGTextFlags(options))
}

// occultationSVGTextFlags 记录哪些文本字段由调用方显式给出：只有这类文本走保守折行与截断，
// 自动文案必须逐字节保持原样。
type occultationSVGTextFlags struct {
	title        bool
	summaryText  bool
	greatestText bool
	footerNote   bool
}

func starOccultationSVGTextFlags(options StarOccultationSVGOptions) occultationSVGTextFlags {
	return occultationSVGTextFlags{
		title:        options.Title != "",
		summaryText:  options.SummaryText != "",
		greatestText: options.GreatestText != "",
		footerNote:   options.FooterNote != "",
	}
}

func renderOccultationPathSVG(
	path moon.StarOccultationPath,
	planetPath *moon.PlanetOccultationPath,
	options StarOccultationSVGOptions,
	flags occultationSVGTextFlags,
) (string, error) {
	projection := resolveStarOccultationMapProjection(path, options.Projection)
	options.Projection = MapProjection(projection)
	title := starOccultationSVGTitle(path, options)
	headerLines := starOccultationSVGHeaderLines(path, options, flags)
	// 页眉字号 14/13，省略号按较大的 14 估算。
	headerLines = svgchart.TruncateTextLines(headerLines, float64(options.Width)-80, 14,
		starOccultationSVGHeaderLineLimit(options, projection))
	headerBottom := 72.0 + float64(len(headerLines))*19
	// 只有正射球面需要视点：不给就会退回 (0°,0°) 而完全不对准事件，与日食球面图不一致。
	// 等经纬与极区图不受影响，避免改动既有版面。
	center := svgmap.GeoPoint{}
	if projection == svgmap.ProjectionOrthographic {
		if planetPath != nil {
			center = planetOccultationOrthographicCentre(*planetPath)
		} else {
			center = svgmap.GeoPoint{Longitude: path.Greatest.Longitude, Latitude: path.Greatest.Latitude}
		}
	}
	layout := starOccultationSVGLayoutFor(options, headerBottom, projection, center)

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" preserveAspectRatio="xMidYMid meet" style="max-width:100%%;height:auto;display:block" role="img" aria-label="%s">`,
		options.Width, options.Height, options.Width, options.Height, html.EscapeString(title))
	b.WriteString(`<defs>`)
	b.WriteString(layout.mapFrame().ClipDefinition("occultation-map-clip"))
	b.WriteString(`</defs>`)
	b.WriteString(`<rect width="100%" height="100%" fill="#efefed"/>`)
	fmt.Fprintf(&b, `<rect x="22" y="18" width="%.3f" height="%.3f" fill="#ffffff" stroke="#c9c9c6" stroke-width="1.2"/>`,
		layout.width-44, layout.height-36)
	titleSize := starOccultationTitleFontSize(title, layout.width)
	titleText := title
	if flags.title {
		titleText = svgchart.EllipsizeText(title, layout.width-80, float64(titleSize))
	}
	fmt.Fprintf(&b, `<text x="%.3f" y="44" fill="#111111" font-family="Georgia, 'Times New Roman', serif" font-size="%d" font-weight="700" text-anchor="middle">%s</text>`,
		layout.width/2, titleSize, html.EscapeString(titleText))
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

	mapTitle := occultationTitleText(starOccultationSVGMapTitle(options), options.MapTitle != "",
		layout.width-layout.mapX-layout.margin, 14)
	fmt.Fprintf(&b, `<text x="%.3f" y="%.3f" fill="#161a1b" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700">%s</text>`,
		layout.mapX, layout.mapY-10, html.EscapeString(mapTitle))
	if planetPath == nil {
		if err := writeStarOccultationMap(&b, path, layout, options); err != nil {
			return "", err
		}
		writeStarOccultationEventsPanel(&b, path, layout, options)
	} else {
		if err := writePlanetOccultationMap(&b, *planetPath, path, layout, options); err != nil {
			return "", err
		}
		writePlanetOccultationEventsPanel(&b, *planetPath, layout, options)
	}
	writeStarOccultationFooter(&b, layout, options, flags, occultationScaleInstant(path, planetPath),
		occultationLimbApplied(path, planetPath))
	b.WriteString(`</svg>`)
	return b.String(), nil
}

func writeStarOccultationMap(
	b *strings.Builder,
	path moon.StarOccultationPath,
	layout starOccultationSVGLayout,
	options StarOccultationSVGOptions,
) error {
	layout.mapFrame().WriteOcean(b)
	writeStarOccultationGraticule(b, layout)
	writeStarOccultationLand(b, layout)
	band := occultationBandDraw{}
	footprints := path.BandFootprints
	compact := len(footprints) > 0
	if len(footprints) == 0 {
		footprints = path.Footprints
	}
	if len(footprints) > 0 {
		draw, err := (occultationFootprintSweep{
			footprints:         footprints,
			contours:           path.BandContours,
			visibilityContours: path.VisibilityContours,
			northern:           path.NorthernLimit,
			southern:           path.SouthernLimit,
			curves:             path.RiseSetCurves,
			layout:             layout,
			layerClass:         "occultation-band-layer",
			pathClass:          "occultation-band",
			color:              "#e0ae43",
			opacity:            0.28,
			compactBand:        compact,
			kind:               starOccultationFootprintSweep,
		}).write(b)
		if err != nil {
			return err
		}
		band = draw
	} else {
		writeStarOccultationBand(b, path, layout)
	}
	if !band.compact {
		writeStarOccultationGeoLine(b, path.NorthernLimit, layout,
			occultationLineStyle{className: "northern-limit", color: "#a66f18", strokeWidth: 1.35, boundary: true})
		writeStarOccultationGeoLine(b, path.SouthernLimit, layout,
			occultationLineStyle{className: "southern-limit", color: "#a66f18", strokeWidth: 1.35, boundary: true})
	}
	// Draw phase curves after the translucent band so the physical rise/set
	// lines remain visible across the filled area. These are open phase curves,
	// not an artificial outline of the static band.
	writeOccultationRiseSetCurves(b, path.RiseSetCurves, layout, "occultation")
	writeOccultationHorizonConnectors(
		b, footprints, path.NorthernLimit, path.SouthernLimit, path.RiseSetCurves, layout,
		true,
	)
	// 事件标记先占位、后绘制：三族标签共用一张占位表，后一族必须绕开先放置的标签。
	placer := &occultationLabelPlacer{}
	eventMarkers := starOccultationEventMarkerDraws(path, layout, options.Language, placer)
	writeOccultationGreatestTimeContours(b, path.GreatestTimeContours, layout, options, placer)
	writeStarOccultationGeoLine(b, path.CenterLine, layout,
		occultationLineStyle{className: "center-line", color: "#59676b", strokeWidth: 1.6, dash: "5 4"})
	writeStarOccultationVisibleCenterLine(b, path.CenterLine, layout)
	writeOccultationTimeMarkers(b, path.CenterLine, layout, options,
		[]time.Time{path.Start.Time, path.End.Time}, path.Greatest.Time, placer)
	for _, marker := range eventMarkers {
		marker.write(b)
	}
	layout.mapFrame().WriteFrame(b)
	writeStarOccultationLegend(b, layout, options.Language, band, len(path.RiseSetCurves) > 0,
		len(path.GreatestTimeContours) > 0)
	return nil
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
		text := fmt.Sprintf("%.0f°", latitude)
		// 图框左侧放不下刻度时改放框内左侧，避免负 x 越出画布。
		x, anchor := layout.mapX-5, "end"
		if x-occultationLabelTextWidth(text, 9) < 3 {
			x, anchor = layout.mapX+4, "start"
		}
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#687577" font-family="Arial, sans-serif" font-size="9" text-anchor="%s">%s</text>`,
			x, y+3, anchor, text)
	}
}

func writeStarOccultationLand(b *strings.Builder, layout starOccultationSVGLayout) {
	layout.mapFrame().WriteLand(b, "occultation-map-clip")
}

// writeOccultationRiseSetCurves 绘制一组升落阶段线；prefix 区分外接触与内接触曲线。
func writeOccultationRiseSetCurves(
	b *strings.Builder,
	curves []moon.OccultationRiseSetCurve,
	layout starOccultationSVGLayout,
	prefix string,
) {
	curves = occultationgeo.DensifyRiseSetCurves(curves, 35)
	for _, curve := range curves {
		className := fmt.Sprintf("%s-rise-set-boundary %s-%s-%s", prefix, prefix, curve.Phase, curve.Direction)
		for _, segment := range occultationgeo.StitchedRiseSetCurveSegments(curve) {
			writeStarOccultationGeoLine(
				b, segment, layout,
				occultationLineStyle{className: className, color: "#d97706", strokeWidth: 1.35},
			)
		}
	}
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
	segments := starOccultationBandFragments(northern, southern, layout.mapFrame().Clip())
	sections := starOccultationBandEndpointSections(northern, southern, layout.mapFrame().Clip())
	if len(segments) == 0 && len(sections) == 0 {
		return
	}
	fmt.Fprintf(b, `<g class="%s" clip-path="url(#occultation-map-clip)" fill="%s" fill-opacity="%.2f" stroke="none">`,
		layerClass, color, opacity)
	for _, segment := range segments {
		fmt.Fprintf(b, `<path class="%s" d="`, pathClass)
		writeOccultationPathCommands(b, layout.projectOccultationRing(segment))
		b.WriteString(`Z"/>`)
	}
	for _, section := range sections {
		fmt.Fprintf(b, `<path class="%s" d="`, pathClass)
		writeOccultationPathCommands(b, layout.projectOccultationPolyline(section))
		fmt.Fprintf(b, `" fill="none" stroke="%s" stroke-opacity="%.2f" stroke-width="1.25" stroke-linecap="round"/>`,
			color, math.Min(1, opacity+0.35))
	}
	b.WriteString(`</g>`)
}

// occultationLineStyle 描述掩带折线的画布样式。
type occultationLineStyle struct {
	className   string
	color       string
	strokeWidth float64
	dash        string
	// boundary 为 true 时按连续边界分段（掩带限线），否则按反经线分段。
	boundary bool
}

func writeStarOccultationGeoLine(
	b *strings.Builder,
	points []moon.OccultationPathPoint,
	layout starOccultationSVGLayout,
	style occultationLineStyle,
) {
	segments := starOccultationPathSegmentsForProjection(points, layout.mapFrame().Clip())
	if style.boundary {
		segments = starOccultationBoundarySegmentsForProjection(points, layout.mapFrame().Clip())
	}
	for _, segment := range segments {
		if len(segment) < 2 {
			continue
		}
		screen := layout.projectOccultationPathLine(segment)
		if len(screen) < 2 {
			continue
		}
		fmt.Fprintf(b, `<path class="%s" d="`, style.className)
		writeOccultationPathCommands(b, screen)
		fmt.Fprintf(b, `" clip-path="url(#occultation-map-clip)" fill="none" stroke="%s" stroke-width="%.2f"`,
			style.color, style.strokeWidth)
		if style.dash != "" {
			fmt.Fprintf(b, ` stroke-dasharray="%s"`, style.dash)
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
		writeStarOccultationGeoLine(b, visible, layout,
			occultationLineStyle{className: "visible-center-line", color: "#087f8c", strokeWidth: 2.8})
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

type starOccultationEventMarkerDraw struct {
	x, y        float64
	label, kind string
	placement   starOccultationEventMarkerPlacement
}

func (draw starOccultationEventMarkerDraw) write(b *strings.Builder) {
	writeStarOccultationProjectedEventMarker(b, draw.x, draw.y, draw.label, draw.kind, draw.placement)
}

// starOccultationEventMarkerDraws 计算起点、掩甚、终点三个事件标记及其标签占位。
func starOccultationEventMarkerDraws(
	path moon.StarOccultationPath,
	layout starOccultationSVGLayout,
	language string,
	placer *occultationLabelPlacer,
) []starOccultationEventMarkerDraw {
	markers := []struct {
		point       moon.OccultationPathPoint
		label, kind string
	}{
		{point: path.Start, label: starOccultationEventLabel("start", language), kind: "start"},
		{point: path.Greatest, label: starOccultationEventLabel("greatest", language), kind: "greatest"},
		{point: path.End, label: starOccultationEventLabel("end", language), kind: "end"},
	}
	draws := make([]starOccultationEventMarkerDraw, 0, len(markers))
	for _, marker := range markers {
		x, y, visible := layout.mapFrame().Project(marker.point.Longitude, marker.point.Latitude)
		if !visible {
			continue
		}
		draws = append(draws, starOccultationEventMarkerDraw{
			x: x, y: y, label: marker.label, kind: marker.kind,
			placement: starOccultationDefaultEventMarkerPlacement(x, y, marker.kind, layout),
		})
	}
	reserveOccultationEventMarkerLabels(draws, layout, placer)
	return draws
}

type starOccultationEventMarkerPlacement struct {
	labelX, labelY float64
	anchor         string
	leader         bool
}

// occultationEventMarkerShifts 是事件标签在默认位置之外的候选位移，按优先级排列。
var occultationEventMarkerShifts = [][2]float64{
	{0, 0}, {0, -14}, {0, 16}, {-12, 0}, {12, 0},
	{-12, -14}, {12, -14}, {-12, 16}, {12, 16},
	{0, -28}, {0, 30}, {-24, 0}, {24, 0}, {0, 44}, {0, -42},
}

// reserveOccultationEventMarkerLabels 把事件标签逐个登记到共用占位表。候选位置依次为默认位置、
// 上下左右位移与左右镜像；全部被占用时退回框内居中的兜底位置，避免标签压到右侧事件面板上。
func reserveOccultationEventMarkerLabels(draws []starOccultationEventMarkerDraw, layout starOccultationSVGLayout, placer *occultationLabelPlacer) {
	for index := range draws {
		draw := &draws[index]
		base := draw.placement
		found := false
		for _, mirror := range []bool{false, true} {
			for _, shift := range occultationEventMarkerShifts {
				if mirror && base.anchor == "middle" {
					continue
				}
				candidate := base
				if mirror {
					candidate.anchor = "start"
					candidate.labelX = draw.x + 7 + shift[0]
					if base.anchor == "start" {
						candidate.anchor, candidate.labelX = "end", draw.x-7+shift[0]
					}
				} else {
					candidate.labelX = base.labelX + shift[0]
				}
				candidate.labelY = math.Max(layout.mapY+12, math.Min(layout.mapY+layout.mapHeight-5, base.labelY+shift[1]))
				if !occultationLabelInsideFrame(layout, candidate.labelX, candidate.labelY, candidate.anchor, draw.label, 11) {
					continue
				}
				if placer.place(candidate.labelX, candidate.labelY, candidate.anchor, draw.label, 11, 2) {
					draw.placement = candidate
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if found {
			continue
		}
		fallback := base
		fallback.anchor = "middle"
		width := occultationLabelTextWidth(draw.label, 11)
		fallback.labelX = math.Max(layout.mapX+width/2+2, math.Min(layout.mapX+layout.mapWidth-width/2-2, draw.x))
		fallback.labelY = math.Max(layout.mapY+12, math.Min(layout.mapY+layout.mapHeight-5, base.labelY))
		draw.placement = fallback
		placer.reserve(occultationLabelRectFor(fallback.labelX, fallback.labelY, fallback.anchor, draw.label, 11).expanded(2))
	}
}

// occultationLabelInsideFrame 报告标签占位是否完全落在地图框内。
func occultationLabelInsideFrame(layout starOccultationSVGLayout, x, y float64, anchor, text string, fontSize float64) bool {
	rect := occultationLabelRectFor(x, y, anchor, text, fontSize)
	return rect.left >= layout.mapX+2 && rect.right <= layout.mapX+layout.mapWidth-2 &&
		rect.top >= layout.mapY+2 && rect.bottom <= layout.mapY+layout.mapHeight-2
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

func writeStarOccultationLegend(
	b *strings.Builder,
	layout starOccultationSVGLayout,
	language string,
	band occultationBandDraw,
	hasRiseSet bool,
	hasIsochrones bool,
) {
	y := layout.mapY + layout.mapHeight + 30
	labels := []string{"可见中心线", "几何中心线", "掩带边界"}
	if language == starOccultationSVGLanguageEnglish {
		labels = []string{"Visible center line", "Geometric center line", "Occultation limits"}
	}
	if band.drawn() {
		labels[2] = "掩带范围"
		if language == starOccultationSVGLanguageEnglish {
			labels[2] = "Band extent"
		}
		if band.claimBoundary() {
			labels[2] = "掩带范围与边界"
			if language == starOccultationSVGLanguageEnglish {
				labels[2] = "Band and limits"
			}
		}
	}
	if hasRiseSet {
		label := "初掩/掩甚/终掩月升月落线"
		if language == starOccultationSVGLanguageEnglish {
			label = "Rise/set phase lines"
		}
		labels = append(labels, label)
	}
	if hasIsochrones {
		label := "掩甚时刻等时线"
		if language == starOccultationSVGLanguageEnglish {
			label = "Greatest-time isochrones"
		}
		labels = append(labels, label)
	}
	colors := []string{"#087f8c", "#59676b", "#a66f18"}
	dashes := []string{"", "5 4", ""}
	if hasRiseSet {
		colors = append(colors, "#d97706")
		dashes = append(dashes, "")
	}
	if hasIsochrones {
		colors = append(colors, "#1f6fb2")
		dashes = append(dashes, "")
	}
	items := make([]occultationLegendItem, 0, len(labels))
	for index, label := range labels {
		item := occultationLegendItem{
			label: label, color: colors[index], dash: dashes[index], markerWidth: 22,
		}
		if band.drawn() && index == 2 {
			item.fill, item.fillOpacity = "#e0ae43", 0.45
		}
		items = append(items, item)
	}
	writeOccultationLegendItems(b, y, layout, items, 10)
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
	contactsTitle := occultationTitleText(starOccultationSVGContactsTitle(options), options.ContactsTitle != "",
		layout.width-layout.panelX-layout.margin, 14)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#161a1b" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700">%s</text>`,
		layout.panelX, layout.panelY+13, html.EscapeString(contactsTitle))
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

// occultationScaleInstant 取时标声明用的掩甚时刻：行星带优先用行星路径。
func occultationScaleInstant(path moon.StarOccultationPath, planetPath *moon.PlanetOccultationPath) time.Time {
	if planetPath != nil {
		return planetPath.Greatest.Time
	}
	return path.Greatest.Time
}

// occultationLimbApplied 报告图中界线是否按月缘剖面修正。
func occultationLimbApplied(path moon.StarOccultationPath, planetPath *moon.PlanetOccultationPath) bool {
	if planetPath != nil {
		return planetPath.LunarLimbApplied
	}
	return path.LunarLimbApplied
}

func writeStarOccultationFooter(
	b *strings.Builder,
	layout starOccultationSVGLayout,
	options StarOccultationSVGOptions,
	flags occultationSVGTextFlags,
	instant time.Time,
	limbApplied bool,
) {
	maxWidth := layout.width - 80
	lineHeight := svgchart.FooterLineHeight(11)
	scaleNote := limbnote.Join(timenote.Scale(options.TimeScale, instant, options.Location, options.Language),
		limbnote.Limits(limbApplied, options.Language), options.Language)
	// 页脚默认排在画布底边之内（与既有版面一致），说明行放得下时声明单独占最后一行；
	// 自定义说明太长时改用按字号留白的行数上限：截断说明也要保住时标声明和底边留白。
	lines := occultationWrappedTextLines(starOccultationSVGFooter(options), flags.footerNote, maxWidth, 11)
	if roomy := svgchart.BaselineLineLimit(11, lineHeight, layout.footerY, layout.height); len(lines) < roomy {
		lines = append(lines, scaleNote)
	} else {
		tight := svgchart.BaselineLineLimit(11, lineHeight, layout.footerY,
			layout.height-svgchart.FooterBottomPadding(11))
		lines = svgchart.FooterLinesWithScale(lines, scaleNote, maxWidth, 11, tight)
	}
	for index, line := range lines {
		fmt.Fprintf(b, `<text x="40" y="%.3f" fill="#596164" font-family="Georgia, 'Times New Roman', serif" font-size="11">%s</text>`,
			layout.footerY+float64(index)*svgchart.FooterLineHeight(11), html.EscapeString(line))
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
