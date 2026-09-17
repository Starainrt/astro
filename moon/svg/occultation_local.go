package svg

import (
	"errors"
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro/internal/svgasset"
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/moon"
)

const (
	localStarOccultationSVGDefaultWidth  = 920
	localStarOccultationSVGDefaultHeight = 720
	localStarOccultationSVGMinimumWidth  = 640
	localStarOccultationSVGMinimumHeight = 520
	localStarOccultationSVGDefaultStep   = 2 * time.Minute
)

// ErrInvalidLocalStarOccultationSVGOptions 表示本地图表选项无效。
// ErrInvalidLocalStarOccultationSVGOptions reports invalid local chart options.
var ErrInvalidLocalStarOccultationSVGOptions = errors.New("invalid local stellar occultation SVG options")

// ErrInvalidLocalStarOccultationInfo 表示固定地点事件数据格式错误。
// ErrInvalidLocalStarOccultationInfo reports malformed fixed-site event data.
var ErrInvalidLocalStarOccultationInfo = errors.New("invalid local stellar occultation info")

// LocalStarOccultationSVGOptions 控制固定地点月球与恒星盘面图。
// LocalStarOccultationSVGOptions controls a fixed-site Moon-star disk chart.
type LocalStarOccultationSVGOptions struct {
	// Width 和 Height 是 SVG 画布尺寸；非正值使用默认值。
	// Width and Height are SVG canvas dimensions. Values <= 0 use defaults.
	Width  int
	Height int
	// Step 是月心恒星轨迹请求采样步长；零值使用两分钟，负值无效，正值小于一秒时提升到一秒。长事件可能增大实际步长，以保持基础轨迹在有界采样预算内。
	// Step is the requested Moon-centered stellar-track sampling step. Zero uses two minutes; negative values are invalid. Positive values below one second are raised to one second, and long events may use a larger effective step to keep the base track within its bounded sample budget.
	Step time.Duration
	// 空文本字段使用本地化的自动标签。
	// Empty text fields use localized automatic labels.
	Title            string
	SummaryText      string
	GreatestText     string
	OverviewTitle    string
	PhasePanelsTitle string
	ContactsTitle    string
	DirectionText    string
	FooterNote       string
	// Language 为 "en" 时使用英文，其他值使用中文。
	// Language uses English for "en" and Chinese for every other value.
	Language string
	// Location 控制显示的事件时刻；nil 使用 UTC+8。
	// Location controls displayed event times. Nil uses UTC+8.
	Location *time.Location
}

type localStarOccultationSVGLayout struct {
	width          float64
	height         float64
	margin         float64
	overviewLeft   float64
	overviewRight  float64
	overviewTop    float64
	overviewBottom float64
	panelX         float64
	panelWidth     float64
	stageTop       float64
	stageBottom    float64
	footerY        float64
	headerMaxLines int
}

type localStarOccultationEventFrame struct {
	label      string
	name       string
	time       time.Time
	frame      moon.StarOccultationDiagramFrame
	appearance localOccultationMoonAppearance
}

// FindLocalStarOccultationSVGs 搜索固定观测地点，并渲染请求时间窗口内的每次点光源恒星月掩。不会加载内嵌星表；空切片表示没有本地事件。
// FindLocalStarOccultationSVGs searches one fixed observing site and renders every point-source stellar occultation in the requested time window. It does not load the embedded star catalog; an empty slice means no local event.
func FindLocalStarOccultationSVGs(
	start, end time.Time,
	star moon.StarCoordinate,
	longitude, latitude, height float64,
	searchOptions moon.OccultationSearchOptions,
	options LocalStarOccultationSVGOptions,
) ([]string, error) {
	if err := validateLocalStarOccultationSVGOptions(options); err != nil {
		return nil, err
	}
	events, err := moon.FindStarOccultations(start, end, star, longitude, latitude, height, searchOptions)
	if err != nil {
		return nil, err
	}
	diagrams := make([]string, 0, len(events))
	for _, event := range events {
		diagram, renderErr := LocalStarOccultationSVG(event, star, options)
		if renderErr != nil {
			return nil, renderErr
		}
		diagrams = append(diagrams, diagram)
	}
	return diagrams, nil
}

// LocalStarOccultationSVG 渲染已求解的固定地点恒星月掩的月心轨迹及掩始、掩甚、掩终阶段面板；不会运行事件搜索或加载星表。
// LocalStarOccultationSVG renders an already solved fixed-site stellar occultation as a Moon-centered track with immersion, greatest, and emersion stage panels. It does not run the event search or load the star catalog.
func LocalStarOccultationSVG(
	info moon.StarOccultationInfo,
	star moon.StarCoordinate,
	options LocalStarOccultationSVGOptions,
) (string, error) {
	if err := validateLocalStarOccultationInfo(info, star); err != nil {
		return "", err
	}
	if err := validateLocalStarOccultationSVGOptions(options); err != nil {
		return "", err
	}
	options = normalizeLocalStarOccultationSVGOptions(options)
	diagram := moon.StarOccultationDiagram(info, star, moon.StarOccultationDiagramOptions{
		StepDays: options.Step.Hours() / 24,
	})
	if len(diagram.Frames) == 0 {
		return "", fmt.Errorf("%w: diagram geometry is unavailable", ErrInvalidLocalStarOccultationInfo)
	}
	return renderLocalStarOccultationSVG(info, diagram, options), nil
}

func validateLocalStarOccultationSVGOptions(options LocalStarOccultationSVGOptions) error {
	if options.Width > 0 && options.Width < localStarOccultationSVGMinimumWidth {
		return fmt.Errorf("%w: width must be zero or at least %d", ErrInvalidLocalStarOccultationSVGOptions, localStarOccultationSVGMinimumWidth)
	}
	if options.Height > 0 && options.Height < localStarOccultationSVGMinimumHeight {
		return fmt.Errorf("%w: height must be zero or at least %d", ErrInvalidLocalStarOccultationSVGOptions, localStarOccultationSVGMinimumHeight)
	}
	if options.Step < 0 {
		return fmt.Errorf("%w: step cannot be negative", ErrInvalidLocalStarOccultationSVGOptions)
	}
	return nil
}

func validateLocalStarOccultationInfo(info moon.StarOccultationInfo, star moon.StarCoordinate) error {
	if err := star.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidLocalStarOccultationInfo, err)
	}
	if info.TargetID != "" && star.ID != "" && info.TargetID != star.ID {
		return fmt.Errorf("%w: event target %q does not match star %q", ErrInvalidLocalStarOccultationInfo, info.TargetID, star.ID)
	}
	if err := info.Observer.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidLocalStarOccultationInfo, err)
	}
	if !info.ContactsComplete || info.Immersion.IsZero() || info.Greatest.IsZero() || info.Emersion.IsZero() {
		return fmt.Errorf("%w: complete immersion, greatest, and emersion times are required", ErrInvalidLocalStarOccultationInfo)
	}
	if info.Greatest.Before(info.Immersion) || info.Emersion.Before(info.Greatest) {
		return fmt.Errorf("%w: event times must be ordered", ErrInvalidLocalStarOccultationInfo)
	}
	if info.Type != moon.OccultationTotal && info.Type != moon.OccultationGrazing {
		return fmt.Errorf("%w: unsupported point-source geometry %q", ErrInvalidLocalStarOccultationInfo, info.Type)
	}
	if !starOccultationFinite(info.MinimumSeparationArcsec) || info.MinimumSeparationArcsec < 0 ||
		!starOccultationFinite(info.MoonSemidiameterArcsec) || info.MoonSemidiameterArcsec <= 0 ||
		!starOccultationFinite(info.PositionAngleDeg) ||
		!starOccultationFinite(info.MoonAltitudeAtGreatest) ||
		!starOccultationFinite(info.MoonAzimuthAtGreatest) {
		return fmt.Errorf("%w: event geometry must be finite", ErrInvalidLocalStarOccultationInfo)
	}
	return nil
}

func normalizeLocalStarOccultationSVGOptions(options LocalStarOccultationSVGOptions) LocalStarOccultationSVGOptions {
	if options.Width <= 0 {
		options.Width = localStarOccultationSVGDefaultWidth
	}
	if options.Height <= 0 {
		options.Height = localStarOccultationSVGDefaultHeight
	}
	if options.Step <= 0 {
		options.Step = localStarOccultationSVGDefaultStep
	}
	if options.Location == nil {
		options.Location = time.FixedZone("UTC+8", starOccultationSVGDefaultZone)
	}
	if strings.EqualFold(options.Language, starOccultationSVGLanguageEnglish) {
		options.Language = starOccultationSVGLanguageEnglish
	} else {
		options.Language = starOccultationSVGLanguageChinese
	}
	return options
}

func renderLocalStarOccultationSVG(
	info moon.StarOccultationInfo,
	diagram moon.StarOccultationDiagramResult,
	options LocalStarOccultationSVGOptions,
) string {
	title := localStarOccultationSVGTitle(info, options)
	headerLines := localStarOccultationSVGHeaderLines(info, options)
	// 页眉字号 14/13，省略号按较大的 14 估算；行数上限只与画布高度有关，先按空页眉取上限。
	headerLimit := localStarOccultationSVGLayoutFor(options, 72).headerMaxLines
	headerLines = svgchart.TruncateTextLines(headerLines, float64(options.Width)-80, 14, headerLimit)
	headerBottom := 72.0 + float64(len(headerLines))*19
	layout := localStarOccultationSVGLayoutFor(options, headerBottom)
	events := localStarOccultationSVGEventFrames(info, diagram.Frames, options.Language)

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s">`,
		options.Width, options.Height, options.Width, options.Height, html.EscapeString(title))
	b.WriteString(`<defs>`)
	b.WriteString(svgasset.MoonFaceSymbol())
	b.WriteString(`<marker id="local-occultation-arrow" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="5" markerHeight="5" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="#9b3f36"/></marker>`)
	b.WriteString(`</defs>`)
	b.WriteString(`<rect width="100%" height="100%" fill="#efefed"/>`)
	fmt.Fprintf(&b, `<rect x="22" y="18" width="%.3f" height="%.3f" fill="#ffffff" stroke="#c9c9c6" stroke-width="1.2"/>`, layout.width-44, layout.height-36)
	titleSize := starOccultationTitleFontSize(title, layout.width)
	titleText := title
	if options.Title != "" {
		titleText = svgchart.EllipsizeText(title, layout.width-80, float64(titleSize))
	}
	fmt.Fprintf(&b, `<text x="%.3f" y="44" fill="#111111" font-family="Georgia, 'Times New Roman', serif" font-size="%d" font-weight="700" text-anchor="middle">%s</text>`,
		layout.width/2, titleSize, html.EscapeString(titleText))
	fmt.Fprintf(&b, `<line x1="%.3f" y1="57" x2="%.3f" y2="57" stroke="#555" stroke-width="1"/>`, layout.width/2-78, layout.width/2+78)
	for index, line := range headerLines {
		fontSize, fill := 13, "#3b4143"
		if index == 0 {
			fontSize, fill = 14, "#222222"
		}
		fmt.Fprintf(&b, `<text x="%.3f" y="%.3f" fill="%s" font-family="Georgia, 'Times New Roman', serif" font-size="%d" text-anchor="middle">%s</text>`,
			layout.width/2, 82+float64(index)*19, fill, fontSize, html.EscapeString(line))
	}

	writeLocalStarOccultationOverview(&b, diagram, events, layout, options)
	writeLocalStarOccultationContacts(&b, events, layout, options)
	writeLocalStarOccultationStages(&b, events, layout, options)
	writeLocalStarOccultationFooter(&b, layout, options)
	b.WriteString(`</svg>`)
	return b.String()
}

func localStarOccultationSVGLayoutFor(options LocalStarOccultationSVGOptions, headerBottom float64) localStarOccultationSVGLayout {
	width := float64(options.Width)
	height := float64(options.Height)
	margin := math.Max(30, math.Min(44, width*0.05))
	gap := math.Max(18, math.Min(24, width*0.025))
	panelWidth := math.Max(188, math.Min(236, width*0.25))
	overviewLeft := margin
	overviewRight := width - margin - panelWidth - gap
	footerSpace := 64.0
	stageHeight := math.Max(132, math.Min(170, height*0.23))
	stageBottom := height - footerSpace
	stageTop := stageBottom - stageHeight
	overviewTop := headerBottom + 26
	overviewBottom := stageTop - 20
	return localStarOccultationSVGLayout{
		width:          width,
		height:         height,
		margin:         margin,
		overviewLeft:   overviewLeft,
		overviewRight:  overviewRight,
		overviewTop:    overviewTop,
		overviewBottom: overviewBottom,
		panelX:         overviewRight + gap,
		panelWidth:     panelWidth,
		stageTop:       stageTop,
		stageBottom:    stageBottom,
		footerY:        height - 43,
		headerMaxLines: localOccultationHeaderLineLimit(height, footerSpace, stageHeight),
	}
}

func writeLocalStarOccultationOverview(
	b *strings.Builder,
	diagram moon.StarOccultationDiagramResult,
	events []localStarOccultationEventFrame,
	layout localStarOccultationSVGLayout,
	options LocalStarOccultationSVGOptions,
) {
	overviewTitle := occultationTitleText(localStarOccultationSVGOverviewTitle(options), options.OverviewTitle != "",
		layout.panelX-layout.overviewLeft-12, 14)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#161a1b" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700">%s</text>`,
		layout.overviewLeft, layout.overviewTop-10, html.EscapeString(overviewTitle))
	cx := (layout.overviewLeft + layout.overviewRight) / 2
	cy := (layout.overviewTop+layout.overviewBottom)/2 + 4
	extent := localStarOccultationSVGExtent(diagram.Frames)
	availableWidth := layout.overviewRight - layout.overviewLeft - 74
	availableHeight := layout.overviewBottom - layout.overviewTop - 50
	scale := math.Min(availableWidth/(2*extent), availableHeight/(2*extent))
	if !starOccultationFinite(scale) || scale <= 0 {
		scale = 1
	}
	mapX := func(value float64) float64 { return cx - value*scale }
	mapY := func(value float64) float64 { return cy - value*scale }
	moonRadius := localStarOccultationSVGMaximumMoonRadius(diagram.Frames) * scale

	writeLocalStarOccultationAxes(b, cx, cy, moonRadius, options.Language)
	writeLocalOccultationMoon(b, cx, cy, moonRadius, "overview-moon",
		localOccultationMoonAppearanceAt(diagram.Occultation.Greatest, diagram.Occultation.Observer),
		"local-star-occultation-moon-overview")
	writeLocalStarOccultationLunarPath(b, diagram.Frames, mapX, mapY, extent, options.Language)
	writeLocalStarOccultationTrack(b, diagram.Frames, mapX, mapY)
	for _, event := range events {
		x := mapX(event.frame.StarXArcsec)
		y := mapY(event.frame.StarYArcsec)
		writeLocalStarOccultationStar(b, x, y, event.frame.BehindMoon, "overview-star", event.label)
		writeLocalStarOccultationOverviewLabel(b, event, x, y, cx, cy, options.Language)
	}
}

func localStarOccultationSVGExtent(frames []moon.StarOccultationDiagramFrame) float64 {
	extent := 1.0
	for _, frame := range frames {
		extent = math.Max(extent, frame.MoonRadiusArcsec)
		extent = math.Max(extent, math.Abs(frame.StarXArcsec))
		extent = math.Max(extent, math.Abs(frame.StarYArcsec))
	}
	return extent * 1.28
}

func localStarOccultationSVGMaximumMoonRadius(frames []moon.StarOccultationDiagramFrame) float64 {
	radius := 1.0
	for _, frame := range frames {
		radius = math.Max(radius, frame.MoonRadiusArcsec)
	}
	return radius
}

func writeLocalStarOccultationTrack(
	b *strings.Builder,
	frames []moon.StarOccultationDiagramFrame,
	mapX, mapY func(float64) float64,
) {
	if len(frames) < 2 {
		return
	}
	var path strings.Builder
	for index, frame := range frames {
		command := "L"
		if index == 0 {
			command = "M"
		}
		fmt.Fprintf(&path, "%s %.3f %.3f ", command, mapX(frame.StarXArcsec), mapY(frame.StarYArcsec))
	}
	fmt.Fprintf(b, `<path class="local-star-track" d="%s" fill="none" stroke="#9b3f36" stroke-width="1.7" stroke-dasharray="5 4" opacity="0.82" marker-end="url(#local-occultation-arrow)"/>`, strings.TrimSpace(path.String()))
}

func writeLocalStarOccultationLunarPath(
	b *strings.Builder,
	frames []moon.StarOccultationDiagramFrame,
	mapX, mapY func(float64) float64,
	extent float64,
	language string,
) {
	unitX, unitY, ok := localStarOccultationSVGLunarPathDirection(frames)
	if !ok {
		return
	}
	lineExtent := extent * 0.92
	startX := mapX(-unitX * lineExtent)
	startY := mapY(-unitY * lineExtent)
	endX := mapX(unitX * lineExtent)
	endY := mapY(unitY * lineExtent)
	if !starOccultationFinite(startX) || !starOccultationFinite(startY) ||
		!starOccultationFinite(endX) || !starOccultationFinite(endY) {
		return
	}

	fmt.Fprintf(b, `<line class="lunar-path-line" x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#506975" stroke-width="1.1" stroke-dasharray="7 4" opacity="0.86"/>`,
		startX, startY, endX, endY)
	labelX, labelY, anchor := startX, startY, "end"
	if endX < startX {
		labelX, labelY = endX, endY
	}
	fmt.Fprintf(b, `<text class="lunar-path-label" x="%.3f" y="%.3f" fill="#405966" stroke="#ffffff" stroke-width="3" paint-order="stroke" font-family="Georgia, 'Times New Roman', serif" font-size="12" text-anchor="%s">%s</text>`,
		labelX-6, labelY-6, anchor, html.EscapeString(localStarOccultationSVGLunarPathLabel(language)))
}

// 恒星轨迹相对固定月心测量，因此其切线与月球站心视运动反向。参考线没有箭头，因此归一化轨迹切线就是图表使用的月掩路径轴。
// The star track is measured relative to a fixed lunar center, so its tangent is antiparallel to the Moon's topocentric apparent motion. A reference line has no arrow, therefore the normalized track tangent is the same lunar-path axis used by the chart.
func localStarOccultationSVGLunarPathDirection(frames []moon.StarOccultationDiagramFrame) (float64, float64, bool) {
	if len(frames) < 2 {
		return 0, 0, false
	}
	greatest := -1
	for index, frame := range frames {
		if localStarOccultationSVGFrameHasLabel(frame, "Greatest") {
			greatest = index
			break
		}
	}
	left, right := 0, len(frames)-1
	if greatest > 0 && greatest+1 < len(frames) {
		left, right = greatest-1, greatest+1
	} else if greatest == 0 {
		left, right = 0, 1
	} else if greatest == len(frames)-1 {
		left, right = len(frames)-2, len(frames)-1
	}
	dx := frames[right].StarXArcsec - frames[left].StarXArcsec
	dy := frames[right].StarYArcsec - frames[left].StarYArcsec
	length := math.Hypot(dx, dy)
	if !starOccultationFinite(length) || length == 0 {
		return 0, 0, false
	}
	return dx / length, dy / length, true
}

func localStarOccultationSVGLunarPathLabel(language string) string {
	if language == starOccultationSVGLanguageEnglish {
		return "Lunar path"
	}
	return "白道"
}

func writeLocalStarOccultationAxes(b *strings.Builder, cx, cy, radius float64, language string) {
	north, east, west, south := "北", "东", "西", "南"
	if language == starOccultationSVGLanguageEnglish {
		north, east, west, south = "N", "E", "W", "S"
	}
	labels := []struct {
		x, y   float64
		anchor string
		text   string
	}{
		{cx, cy - radius - 15, "middle", north},
		{cx - radius - 18, cy + 4, "middle", east},
		{cx + radius + 18, cy + 4, "middle", west},
		{cx, cy + radius + 24, "middle", south},
	}
	for _, label := range labels {
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#293235" font-family="Georgia, 'Times New Roman', serif" font-size="12" font-weight="700" text-anchor="%s">%s</text>`,
			label.x, label.y, label.anchor, html.EscapeString(label.text))
	}
}

func writeLocalStarOccultationStar(b *strings.Builder, x, y float64, hidden bool, class, label string) {
	fill, stroke, dash := "#fff7d6", "#9b3028", ""
	if hidden {
		fill, stroke, dash = "none", "#7d3430", ` stroke-dasharray="2 2"`
	}
	fmt.Fprintf(b, `<g class="%s" data-label="%s">`, html.EscapeString(class), html.EscapeString(label))
	fmt.Fprintf(b, `<circle cx="%.3f" cy="%.3f" r="4.2" fill="%s" stroke="%s" stroke-width="1.4"%s/>`, x, y, fill, stroke, dash)
	fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="%s" stroke-width="1"/>`, x-6, y, x+6, y, stroke)
	fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="%s" stroke-width="1"/>`, x, y-6, x, y+6, stroke)
	b.WriteString(`</g>`)
}

func writeLocalStarOccultationOverviewLabel(
	b *strings.Builder,
	event localStarOccultationEventFrame,
	x, y, cx, cy float64,
	language string,
) {
	dx, dy, anchor := 8.0, -9.0, "start"
	if x > cx {
		dx, anchor = -8, "end"
	}
	if y < cy-20 {
		dy = 15
	}
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#792d28" stroke="#ffffff" stroke-width="3" paint-order="stroke" font-family="Arial, sans-serif" font-size="11" font-weight="700" text-anchor="%s">%s</text>`,
		x+dx, y+dy, anchor, html.EscapeString(localStarOccultationSVGEventName(event.label, language)))
}

func writeLocalStarOccultationContacts(
	b *strings.Builder,
	events []localStarOccultationEventFrame,
	layout localStarOccultationSVGLayout,
	options LocalStarOccultationSVGOptions,
) {
	fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#d0d3d1" stroke-width="1"/>`,
		layout.panelX-12, layout.overviewTop-12, layout.panelX-12, layout.overviewBottom)
	contactsTitle := occultationTitleText(localStarOccultationSVGContactsTitle(options), options.ContactsTitle != "",
		layout.width-layout.panelX-layout.margin, 14)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#161a1b" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700">%s</text>`,
		layout.panelX, layout.overviewTop-10, html.EscapeString(contactsTitle))
	available := layout.overviewBottom - layout.overviewTop - 10
	rowHeight := math.Min(86, available/math.Max(1, float64(len(events))))
	for index, event := range events {
		y := layout.overviewTop + 17 + float64(index)*rowHeight
		if index > 0 {
			fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#e0e1df" stroke-width="0.8"/>`,
				layout.panelX, y-13, layout.panelX+layout.panelWidth, y-13)
		}
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#1c2528" font-family="Arial, sans-serif" font-size="12" font-weight="700">%s</text>`,
			layout.panelX, y, html.EscapeString(event.name))
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#1c2528" font-family="Arial, sans-serif" font-size="11" text-anchor="end">%s</text>`,
			layout.panelX+layout.panelWidth, y, html.EscapeString(event.time.In(options.Location).Format("15:04:05.0")))
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#536063" font-family="Arial, sans-serif" font-size="10">PA %.1f° | %s %.1f°</text>`,
			layout.panelX, y+18, event.frame.PositionAngleDeg, localStarOccultationSVGAltitudeLabel(options.Language), event.frame.MoonAltitudeDeg)
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#6b7577" font-family="Arial, sans-serif" font-size="10">%s %.1f° | %s</text>`,
			layout.panelX, y+34, localStarOccultationSVGAzimuthLabel(options.Language), event.frame.MoonAzimuthDeg,
			html.EscapeString(localStarOccultationSVGVisibilityText(event.frame.MoonAltitudeDeg >= 0, options.Language)))
	}
}

func writeLocalStarOccultationStages(
	b *strings.Builder,
	events []localStarOccultationEventFrame,
	layout localStarOccultationSVGLayout,
	options LocalStarOccultationSVGOptions,
) {
	phaseTitle := occultationTitleText(localStarOccultationSVGPhasePanelsTitle(options), options.PhasePanelsTitle != "",
		layout.width-2*layout.margin, 14)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#161a1b" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700">%s</text>`,
		layout.margin, layout.stageTop+13, html.EscapeString(phaseTitle))
	if len(events) == 0 {
		return
	}
	usableWidth := layout.width - 2*layout.margin
	columnWidth := usableWidth / float64(len(events))
	maxMoonRadius := localStarOccultationSVGMaximumMoonRadiusFromEvents(events)
	radius := math.Min(46, math.Max(31, (layout.stageBottom-layout.stageTop-53)/2))
	scale := radius / maxMoonRadius
	cy := layout.stageTop + 34 + radius
	for index, event := range events {
		cx := layout.margin + columnWidth*(float64(index)+0.5)
		writeLocalOccultationMoon(b, cx, cy, event.frame.MoonRadiusArcsec*scale, "stage-moon",
			event.appearance, fmt.Sprintf("local-star-occultation-moon-stage-%d", index))
		x := cx - event.frame.StarXArcsec*scale
		y := cy - event.frame.StarYArcsec*scale
		writeLocalStarOccultationStar(b, x, y, event.frame.BehindMoon, "stage-star", event.label)
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#1c2528" font-family="Arial, sans-serif" font-size="11" font-weight="700" text-anchor="middle">%s</text>`,
			cx, cy+radius+19, html.EscapeString(event.name))
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#667174" font-family="Arial, sans-serif" font-size="10" text-anchor="middle">%s</text>`,
			cx, cy+radius+34, html.EscapeString(event.time.In(options.Location).Format("15:04:05.0")))
	}
}

func localStarOccultationSVGMaximumMoonRadiusFromEvents(events []localStarOccultationEventFrame) float64 {
	radius := 1.0
	for _, event := range events {
		radius = math.Max(radius, event.frame.MoonRadiusArcsec)
	}
	return radius
}

func writeLocalStarOccultationFooter(
	b *strings.Builder,
	layout localStarOccultationSVGLayout,
	options LocalStarOccultationSVGOptions,
) {
	directionLines, footerLines := localOccultationFooterLines(
		localStarOccultationSVGDirectionText(options), options.DirectionText != "",
		localStarOccultationSVGFooterNote(options), options.FooterNote != "",
		layout.footerY, layout.width, layout.height)
	for index, line := range directionLines {
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#4f595b" font-family="Georgia, 'Times New Roman', serif" font-size="10">%s</text>`,
			layout.margin, layout.footerY+float64(index)*14-8, html.EscapeString(line))
	}
	for index, line := range footerLines {
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#747c7d" font-family="Georgia, 'Times New Roman', serif" font-size="9">%s</text>`,
			layout.margin, layout.footerY+float64(len(directionLines))*14+float64(index)*12-6, html.EscapeString(line))
	}
}

func localStarOccultationSVGEventFrames(
	info moon.StarOccultationInfo,
	frames []moon.StarOccultationDiagramFrame,
	language string,
) []localStarOccultationEventFrame {
	times := map[string]time.Time{
		"Immersion": info.Immersion,
		"Greatest":  info.Greatest,
		"Emersion":  info.Emersion,
	}
	labels := []string{"Immersion", "Greatest", "Emersion"}
	result := make([]localStarOccultationEventFrame, 0, len(labels))
	for _, label := range labels {
		for _, frame := range frames {
			if localStarOccultationSVGFrameHasLabel(frame, label) {
				result = append(result, localStarOccultationEventFrame{
					label:      label,
					name:       localStarOccultationSVGEventName(label, language),
					time:       times[label],
					frame:      frame,
					appearance: localOccultationMoonAppearanceAt(times[label], info.Observer),
				})
				break
			}
		}
	}
	return result
}

func localStarOccultationSVGFrameHasLabel(frame moon.StarOccultationDiagramFrame, label string) bool {
	if frame.Label == label {
		return true
	}
	for _, current := range frame.Labels {
		if current == label {
			return true
		}
	}
	return false
}

func localStarOccultationSVGHeaderLines(info moon.StarOccultationInfo, options LocalStarOccultationSVGOptions) []string {
	lines := make([]string, 0, 4)
	for _, item := range []struct {
		value  string
		custom bool
	}{
		{localStarOccultationSVGSummaryText(info, options), options.SummaryText != ""},
		{localStarOccultationSVGGreatestText(info, options), options.GreatestText != ""},
	} {
		lines = append(lines, occultationWrappedTextLines(item.value, item.custom, float64(options.Width)-80, 13)...)
	}
	return lines
}

// localOccultationFooterLines 排布本地图页脚：补充说明贴画布底边，方向说明占用上方余高，两块都按行数上限截断。
func localOccultationFooterLines(
	directionText string, directionCustom bool,
	noteText string, noteCustom bool,
	footerY, width, height float64,
) ([]string, []string) {
	maxWidth := width - 80
	direction := occultationWrappedTextLines(directionText, directionCustom, maxWidth, 10)
	note := occultationWrappedTextLines(noteText, noteCustom, maxWidth, 9)
	noteAbove, noteBelow := svgchart.EstimatedTextExtents(9)
	noteHeight := noteAbove + float64(len(note)-1)*12 + noteBelow
	direction = svgchart.TruncateTextLines(direction, maxWidth, 10,
		svgchart.BaselineLineLimit(10, 14, footerY-8, height-2-noteHeight))
	noteFirst := footerY + float64(len(direction))*14 - 6
	note = svgchart.TruncateTextLines(note, maxWidth, 9,
		svgchart.BaselineLineLimit(9, 12, noteFirst, height-2))
	return direction, note
}

// localOccultationHeaderLineLimit 由版式的页脚预留与阶段区高度推出页眉行数上限，总览区被压到 120 px 以下时圆盘与方向标记会跑出画布。
func localOccultationHeaderLineLimit(height, footerSpace, stageHeight float64) int {
	const (
		stageGap       = 20.0
		overviewTopGap = 26.0
		overviewMin    = 120.0
		headerStart    = 72.0
	)
	limit := int(math.Floor((height - footerSpace - stageHeight - stageGap - overviewMin - overviewTopGap - headerStart) / 19))
	if limit < 1 {
		return 1
	}
	return limit
}

func localStarOccultationSVGTitle(info moon.StarOccultationInfo, options LocalStarOccultationSVGOptions) string {
	if options.Title != "" {
		return options.Title
	}
	target := info.TargetID
	if target == "" {
		if options.Language == starOccultationSVGLanguageEnglish {
			target = "star"
		} else {
			target = "恒星"
		}
	}
	date := info.Greatest.In(options.Location).Format("2006-01-02")
	if options.Language == starOccultationSVGLanguageEnglish {
		return fmt.Sprintf("%s Local Lunar Occultation of %s", date, target)
	}
	return fmt.Sprintf("%s 指定地点月掩%s", date, target)
}

func localStarOccultationSVGSummaryText(info moon.StarOccultationInfo, options LocalStarOccultationSVGOptions) string {
	if options.SummaryText != "" {
		return options.SummaryText
	}
	coordinates := starOccultationFormatCoordinates(info.Observer.Longitude, info.Observer.Latitude)
	duration := localStarOccultationSVGFormatDuration(info.Emersion.Sub(info.Immersion), options.Language)
	if options.Language == starOccultationSVGLanguageEnglish {
		return fmt.Sprintf("Site %s | elevation %.0f m | %s | duration %s", coordinates, info.Observer.Height,
			localStarOccultationSVGTypeName(info.Type, options.Language), duration)
	}
	return fmt.Sprintf("观测点 %s | 海拔 %.0f 米 | %s | 掩星历时 %s", coordinates, info.Observer.Height,
		localStarOccultationSVGTypeName(info.Type, options.Language), duration)
}

func localStarOccultationSVGGreatestText(info moon.StarOccultationInfo, options LocalStarOccultationSVGOptions) string {
	if options.GreatestText != "" {
		return options.GreatestText
	}
	value := info.Greatest.In(options.Location)
	zone := starOccultationLocationLabel(value, options.Location)
	if options.Language == starOccultationSVGLanguageEnglish {
		return fmt.Sprintf("Greatest %s %s | minimum separation %.2f arcsec | Moon altitude %+.1f° azimuth %.1f°",
			value.Format("2006-01-02 15:04:05.0"), zone, info.MinimumSeparationArcsec, info.MoonAltitudeAtGreatest, info.MoonAzimuthAtGreatest)
	}
	return fmt.Sprintf("掩甚 %s %s | 最小角距 %.2f 角秒 | 月球高度 %+.1f° 方位 %.1f°",
		value.Format("2006-01-02 15:04:05.0"), zone, info.MinimumSeparationArcsec, info.MoonAltitudeAtGreatest, info.MoonAzimuthAtGreatest)
}

func localStarOccultationSVGOverviewTitle(options LocalStarOccultationSVGOptions) string {
	if options.OverviewTitle != "" {
		return options.OverviewTitle
	}
	if options.Language == starOccultationSVGLanguageEnglish {
		return "Topocentric star track"
	}
	return "站心恒星轨迹"
}

func localStarOccultationSVGPhasePanelsTitle(options LocalStarOccultationSVGOptions) string {
	if options.PhasePanelsTitle != "" {
		return options.PhasePanelsTitle
	}
	if options.Language == starOccultationSVGLanguageEnglish {
		return "Contact stages"
	}
	return "掩始、掩甚与掩终视圆"
}

func localStarOccultationSVGContactsTitle(options LocalStarOccultationSVGOptions) string {
	if options.ContactsTitle != "" {
		return options.ContactsTitle
	}
	if options.Language == starOccultationSVGLanguageEnglish {
		return "Local contacts"
	}
	return "本地接触时刻"
}

func localStarOccultationSVGDirectionText(options LocalStarOccultationSVGOptions) string {
	if options.DirectionText != "" {
		return options.DirectionText
	}
	if options.Language == starOccultationSVGLanguageEnglish {
		return "Moon fixed at center; east is left and north is up. The blue-gray dashed line is the local lunar path; the red dashed line is the star track."
	}
	return "月球固定在中心；图上左东右西，向上为北。灰蓝虚线为掩甚附近的站心白道，红色虚线为恒星相对月心轨迹。"
}

func localStarOccultationSVGFooterNote(options LocalStarOccultationSVGOptions) string {
	if options.FooterNote != "" {
		return options.FooterNote
	}
	if options.Language == starOccultationSVGLanguageEnglish {
		return "The star is a point source; immersion and emersion occur where it crosses the topocentric apparent lunar limb. Lunar texture is schematic."
	}
	return "恒星按点光源绘制；掩始和掩终是恒星穿越站心月球视圆外缘的时刻，月面纹理仅作方向辅助。"
}

func localStarOccultationSVGEventName(label, language string) string {
	if language == starOccultationSVGLanguageEnglish {
		switch label {
		case "Immersion":
			return "Immersion"
		case "Greatest":
			return "Greatest"
		case "Emersion":
			return "Emersion"
		}
	}
	switch label {
	case "Immersion":
		return "掩始"
	case "Greatest":
		return "掩甚"
	case "Emersion":
		return "掩终"
	default:
		return label
	}
}

func localStarOccultationSVGTypeName(eventType moon.OccultationType, language string) string {
	if language == starOccultationSVGLanguageEnglish {
		if eventType == moon.OccultationGrazing {
			return "grazing"
		}
		return "total"
	}
	if eventType == moon.OccultationGrazing {
		return "擦边掩星"
	}
	return "全掩"
}

func localStarOccultationSVGFormatDuration(duration time.Duration, language string) string {
	if duration < 0 {
		duration = -duration
	}
	duration = duration.Round(100 * time.Millisecond)
	hours := int(duration / time.Hour)
	duration -= time.Duration(hours) * time.Hour
	minutes := int(duration / time.Minute)
	duration -= time.Duration(minutes) * time.Minute
	seconds := float64(duration) / float64(time.Second)
	if language == starOccultationSVGLanguageEnglish {
		if hours > 0 {
			return fmt.Sprintf("%dh %02dm %04.1fs", hours, minutes, seconds)
		}
		return fmt.Sprintf("%dm %04.1fs", minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%d时%02d分%04.1f秒", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d分%04.1f秒", minutes, seconds)
}

func localStarOccultationSVGAltitudeLabel(language string) string {
	if language == starOccultationSVGLanguageEnglish {
		return "alt"
	}
	return "高度"
}

func localStarOccultationSVGAzimuthLabel(language string) string {
	if language == starOccultationSVGLanguageEnglish {
		return "az"
	}
	return "方位"
}

func localStarOccultationSVGVisibilityText(visible bool, language string) string {
	if language == starOccultationSVGLanguageEnglish {
		if visible {
			return "above horizon"
		}
		return "below horizon"
	}
	if visible {
		return "地平线上"
	}
	return "地平线下"
}
