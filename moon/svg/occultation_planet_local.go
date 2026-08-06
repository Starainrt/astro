package svg

import (
	"errors"
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro/internal/svgasset"
	"github.com/starainrt/astro/moon"
)

const (
	localPlanetOccultationSVGDefaultWidth  = 1040
	localPlanetOccultationSVGDefaultHeight = 760
	localPlanetOccultationSVGMinimumWidth  = 760
	localPlanetOccultationSVGMinimumHeight = 600
	localPlanetOccultationSVGDefaultStep   = 2 * time.Minute
)

// ErrInvalidLocalPlanetOccultationSVGOptions 表示本地行星图表选项无效。
// ErrInvalidLocalPlanetOccultationSVGOptions reports invalid local planetary chart options.
var ErrInvalidLocalPlanetOccultationSVGOptions = errors.New("invalid local planetary occultation SVG options")

// ErrInvalidLocalPlanetOccultationInfo 表示固定地点行星事件数据格式错误。
// ErrInvalidLocalPlanetOccultationInfo reports malformed fixed-site planetary event data.
var ErrInvalidLocalPlanetOccultationInfo = errors.New("invalid local planetary occultation info")

// LocalPlanetOccultationSVGOptions 控制固定地点月球与行星盘面图。
// LocalPlanetOccultationSVGOptions controls a fixed-site Moon-planet disk chart.
type LocalPlanetOccultationSVGOptions struct {
	// Width 和 Height 是 SVG 画布尺寸；非正值使用默认值。
	// Width and Height are SVG canvas dimensions. Values <= 0 use defaults.
	Width  int
	Height int
	// Step 是月心行星轨迹请求采样步长；零值使用两分钟，负值无效，正值小于一秒时提升到一秒。长事件可能增大实际步长，以保持基础轨迹在有界采样预算内。
	// Step is the requested Moon-centered planetary-track sampling step. Zero uses two minutes; negative values are invalid. Positive values below one second are raised to one second, and long events may use a larger effective step to keep the base track within its bounded sample budget.
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

type localPlanetOccultationEventFrame struct {
	label  string
	name   string
	time   time.Time
	hidden bool
	frame  moon.PlanetOccultationDiagramFrame
}

// FindLocalPlanetOccultationSVGs 搜索固定观测地点，并渲染请求时间窗口内的每次有限盘面行星月掩。空切片表示没有本地事件。
// FindLocalPlanetOccultationSVGs searches one fixed observing site and renders every finite-disk planetary occultation in the requested time window. An empty slice means no local event.
func FindLocalPlanetOccultationSVGs(
	start, end time.Time,
	planet moon.OccultationPlanet,
	longitude, latitude, height float64,
	searchOptions moon.OccultationSearchOptions,
	options LocalPlanetOccultationSVGOptions,
) ([]string, error) {
	if err := validateLocalPlanetOccultationSVGOptions(options); err != nil {
		return nil, err
	}
	events, err := moon.FindPlanetOccultations(
		start, end, planet, longitude, latitude, height, searchOptions,
	)
	if err != nil {
		return nil, err
	}
	diagrams := make([]string, 0, len(events))
	for _, event := range events {
		diagram, renderErr := LocalPlanetOccultationSVG(event, options)
		if renderErr != nil {
			return nil, renderErr
		}
		diagrams = append(diagrams, diagram)
	}
	return diagrams, nil
}

// LocalPlanetOccultationSVG 渲染已求解的固定地点行星月掩的月心轨迹及适用的 C1-C4 阶段面板；不会运行事件搜索。
// LocalPlanetOccultationSVG renders an already solved fixed-site planetary occultation as a Moon-centered track with all applicable C1-C4 stage panels. It does not run the event search.
func LocalPlanetOccultationSVG(
	info moon.PlanetOccultationInfo,
	options LocalPlanetOccultationSVGOptions,
) (string, error) {
	if err := validateLocalPlanetOccultationInfo(info); err != nil {
		return "", err
	}
	if err := validateLocalPlanetOccultationSVGOptions(options); err != nil {
		return "", err
	}
	options = normalizeLocalPlanetOccultationSVGOptions(options)
	diagram := moon.PlanetOccultationDiagram(info, moon.PlanetOccultationDiagramOptions{
		StepDays: options.Step.Hours() / 24,
	})
	if len(diagram.Frames) == 0 {
		return "", fmt.Errorf("%w: diagram geometry is unavailable", ErrInvalidLocalPlanetOccultationInfo)
	}
	return renderLocalPlanetOccultationSVG(info, diagram, options), nil
}

func validateLocalPlanetOccultationSVGOptions(options LocalPlanetOccultationSVGOptions) error {
	if options.Width > 0 && options.Width < localPlanetOccultationSVGMinimumWidth {
		return fmt.Errorf("%w: width must be zero or at least %d", ErrInvalidLocalPlanetOccultationSVGOptions, localPlanetOccultationSVGMinimumWidth)
	}
	if options.Height > 0 && options.Height < localPlanetOccultationSVGMinimumHeight {
		return fmt.Errorf("%w: height must be zero or at least %d", ErrInvalidLocalPlanetOccultationSVGOptions, localPlanetOccultationSVGMinimumHeight)
	}
	if options.Step < 0 {
		return fmt.Errorf("%w: step cannot be negative", ErrInvalidLocalPlanetOccultationSVGOptions)
	}
	return nil
}

func validateLocalPlanetOccultationInfo(info moon.PlanetOccultationInfo) error {
	if err := info.Planet.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidLocalPlanetOccultationInfo, err)
	}
	if err := info.Observer.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidLocalPlanetOccultationInfo, err)
	}
	if !info.ContactsComplete || info.ExternalImmersion.IsZero() || info.Greatest.IsZero() || info.ExternalEmersion.IsZero() {
		return fmt.Errorf("%w: complete external contacts and greatest time are required", ErrInvalidLocalPlanetOccultationInfo)
	}
	if info.Greatest.Before(info.ExternalImmersion) || info.ExternalEmersion.Before(info.Greatest) {
		return fmt.Errorf("%w: event times must be ordered", ErrInvalidLocalPlanetOccultationInfo)
	}
	if info.Type == moon.OccultationTotal {
		if !info.HasInternalContacts || info.InternalImmersion.IsZero() || info.InternalEmersion.IsZero() ||
			!info.InternalImmersion.After(info.ExternalImmersion) || !info.InternalImmersion.Before(info.Greatest) ||
			!info.InternalEmersion.After(info.Greatest) || !info.InternalEmersion.Before(info.ExternalEmersion) {
			return fmt.Errorf("%w: a total event requires ordered C2 and C3 contacts", ErrInvalidLocalPlanetOccultationInfo)
		}
	} else if info.Type != moon.OccultationPartial && info.Type != moon.OccultationGrazing {
		return fmt.Errorf("%w: unsupported finite-disk geometry %q", ErrInvalidLocalPlanetOccultationInfo, info.Type)
	} else if info.HasInternalContacts || !info.InternalImmersion.IsZero() || !info.InternalEmersion.IsZero() {
		return fmt.Errorf("%w: partial and grazing events cannot contain internal contacts", ErrInvalidLocalPlanetOccultationInfo)
	}
	if !starOccultationFinite(info.MinimumSeparationArcsec) || info.MinimumSeparationArcsec < 0 ||
		!starOccultationFinite(info.MoonSemidiameterArcsec) || info.MoonSemidiameterArcsec <= 0 ||
		!starOccultationFinite(info.PlanetSemidiameterArcsec) || info.PlanetSemidiameterArcsec <= 0 ||
		info.PlanetSemidiameterArcsec >= info.MoonSemidiameterArcsec ||
		!starOccultationFinite(info.PositionAngleDeg) ||
		!starOccultationFinite(info.MoonAltitudeAtGreatest) ||
		!starOccultationFinite(info.MoonAzimuthAtGreatest) {
		return fmt.Errorf("%w: event geometry must be finite", ErrInvalidLocalPlanetOccultationInfo)
	}
	return nil
}

func normalizeLocalPlanetOccultationSVGOptions(options LocalPlanetOccultationSVGOptions) LocalPlanetOccultationSVGOptions {
	if options.Width <= 0 {
		options.Width = localPlanetOccultationSVGDefaultWidth
	}
	if options.Height <= 0 {
		options.Height = localPlanetOccultationSVGDefaultHeight
	}
	if options.Step <= 0 {
		options.Step = localPlanetOccultationSVGDefaultStep
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

func renderLocalPlanetOccultationSVG(
	info moon.PlanetOccultationInfo,
	diagram moon.PlanetOccultationDiagramResult,
	options LocalPlanetOccultationSVGOptions,
) string {
	title := localPlanetOccultationSVGTitle(info, options)
	headerLines := localPlanetOccultationSVGHeaderLines(info, options)
	headerBottom := 72.0 + float64(len(headerLines))*19
	layout := localPlanetOccultationSVGLayoutFor(options, headerBottom)
	events := localPlanetOccultationSVGEventFrames(info, diagram.Frames, options.Language)

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s">`,
		options.Width, options.Height, options.Width, options.Height, html.EscapeString(title))
	b.WriteString(`<defs>`)
	b.WriteString(svgasset.MoonFaceSymbol())
	b.WriteString(`<marker id="local-planet-occultation-arrow" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="5" markerHeight="5" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="#9b3f36"/></marker>`)
	b.WriteString(`</defs>`)
	b.WriteString(`<rect width="100%" height="100%" fill="#efefed"/>`)
	fmt.Fprintf(&b, `<rect x="22" y="18" width="%.3f" height="%.3f" fill="#ffffff" stroke="#c9c9c6" stroke-width="1.2"/>`, layout.width-44, layout.height-36)
	fmt.Fprintf(&b, `<text x="%.3f" y="44" fill="#111111" font-family="Georgia, 'Times New Roman', serif" font-size="%d" font-weight="700" text-anchor="middle">%s</text>`,
		layout.width/2, starOccultationTitleFontSize(title, layout.width), html.EscapeString(title))
	fmt.Fprintf(&b, `<line x1="%.3f" y1="57" x2="%.3f" y2="57" stroke="#555" stroke-width="1"/>`, layout.width/2-78, layout.width/2+78)
	for index, line := range headerLines {
		fontSize, fill := 13, "#3b4143"
		if index == 0 {
			fontSize, fill = 14, "#222222"
		}
		fmt.Fprintf(&b, `<text x="%.3f" y="%.3f" fill="%s" font-family="Georgia, 'Times New Roman', serif" font-size="%d" text-anchor="middle">%s</text>`,
			layout.width/2, 82+float64(index)*19, fill, fontSize, html.EscapeString(line))
	}

	writeLocalPlanetOccultationOverview(&b, diagram, events, layout, options)
	writeLocalPlanetOccultationContacts(&b, events, layout, options)
	writeLocalPlanetOccultationStages(&b, events, layout, options)
	writeLocalPlanetOccultationFooter(&b, info, layout, options)
	b.WriteString(`</svg>`)
	return b.String()
}

func localPlanetOccultationSVGLayoutFor(
	options LocalPlanetOccultationSVGOptions,
	headerBottom float64,
) localStarOccultationSVGLayout {
	width := float64(options.Width)
	height := float64(options.Height)
	margin := math.Max(34, math.Min(44, width*0.045))
	gap := math.Max(18, math.Min(24, width*0.025))
	panelWidth := math.Max(210, math.Min(258, width*0.25))
	overviewLeft := margin
	overviewRight := width - margin - panelWidth - gap
	footerSpace := 88.0
	stageHeight := math.Max(156, math.Min(190, height*0.25))
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
		footerY:        height - 62,
	}
}

func writeLocalPlanetOccultationOverview(
	b *strings.Builder,
	diagram moon.PlanetOccultationDiagramResult,
	events []localPlanetOccultationEventFrame,
	layout localStarOccultationSVGLayout,
	options LocalPlanetOccultationSVGOptions,
) {
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#161a1b" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700">%s</text>`,
		layout.overviewLeft, layout.overviewTop-10, html.EscapeString(localPlanetOccultationSVGOverviewTitle(options)))
	cx := (layout.overviewLeft + layout.overviewRight) / 2
	cy := (layout.overviewTop+layout.overviewBottom)/2 + 4
	extent := localPlanetOccultationSVGExtent(diagram.Frames)
	availableWidth := layout.overviewRight - layout.overviewLeft - 74
	availableHeight := layout.overviewBottom - layout.overviewTop - 50
	scale := math.Min(availableWidth/(2*extent), availableHeight/(2*extent))
	if !starOccultationFinite(scale) || scale <= 0 {
		scale = 1
	}
	mapX := func(value float64) float64 { return cx - value*scale }
	mapY := func(value float64) float64 { return cy - value*scale }
	moonRadius := localPlanetOccultationSVGMaximumMoonRadius(diagram.Frames) * scale

	writeLocalStarOccultationAxes(b, cx, cy, moonRadius, options.Language)
	writeLocalPlanetOccultationLunarPath(b, diagram.Frames, mapX, mapY, extent, options.Language)
	writeLocalPlanetOccultationTrack(b, diagram.Frames, mapX, mapY)
	for _, event := range events {
		writeLocalPlanetOccultationDisk(b,
			mapX(event.frame.PlanetXArcsec), mapY(event.frame.PlanetYArcsec),
			event.frame.PlanetRadiusArcsec*scale, false, "overview-planet-disk", event.label,
		)
	}
	writeLocalStarOccultationMoon(b, cx, cy, moonRadius, "overview-moon")
	for _, event := range events {
		x := mapX(event.frame.PlanetXArcsec)
		y := mapY(event.frame.PlanetYArcsec)
		if event.hidden {
			writeLocalPlanetOccultationDisk(b, x, y, event.frame.PlanetRadiusArcsec*scale, true, "overview-planet-outline", event.label)
		}
		if event.label != "C2" && event.label != "C3" {
			writeLocalPlanetOccultationOverviewLabel(b, event, x, y, cx, cy)
		}
	}
}

func localPlanetOccultationSVGExtent(frames []moon.PlanetOccultationDiagramFrame) float64 {
	extent := 1.0
	for _, frame := range frames {
		extent = math.Max(extent, frame.MoonRadiusArcsec)
		extent = math.Max(extent, math.Abs(frame.PlanetXArcsec)+frame.PlanetRadiusArcsec)
		extent = math.Max(extent, math.Abs(frame.PlanetYArcsec)+frame.PlanetRadiusArcsec)
	}
	return extent * 1.28
}

func localPlanetOccultationSVGMaximumMoonRadius(frames []moon.PlanetOccultationDiagramFrame) float64 {
	radius := 1.0
	for _, frame := range frames {
		radius = math.Max(radius, frame.MoonRadiusArcsec)
	}
	return radius
}

func writeLocalPlanetOccultationTrack(
	b *strings.Builder,
	frames []moon.PlanetOccultationDiagramFrame,
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
		fmt.Fprintf(&path, "%s %.3f %.3f ", command, mapX(frame.PlanetXArcsec), mapY(frame.PlanetYArcsec))
	}
	fmt.Fprintf(b, `<path class="local-planet-track" d="%s" fill="none" stroke="#9b3f36" stroke-width="1.7" stroke-dasharray="5 4" opacity="0.82" marker-end="url(#local-planet-occultation-arrow)"/>`, strings.TrimSpace(path.String()))
}

func writeLocalPlanetOccultationLunarPath(
	b *strings.Builder,
	frames []moon.PlanetOccultationDiagramFrame,
	mapX, mapY func(float64) float64,
	extent float64,
	language string,
) {
	unitX, unitY, ok := localPlanetOccultationSVGLunarPathDirection(frames)
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
	labelX, labelY := startX, startY
	if endX < startX {
		labelX, labelY = endX, endY
	}
	fmt.Fprintf(b, `<text class="lunar-path-label" x="%.3f" y="%.3f" fill="#405966" stroke="#ffffff" stroke-width="3" paint-order="stroke" font-family="Georgia, 'Times New Roman', serif" font-size="12" text-anchor="end">%s</text>`,
		labelX-6, labelY-6, html.EscapeString(localStarOccultationSVGLunarPathLabel(language)))
}

func localPlanetOccultationSVGLunarPathDirection(frames []moon.PlanetOccultationDiagramFrame) (float64, float64, bool) {
	if len(frames) < 2 {
		return 0, 0, false
	}
	greatest := -1
	for index, frame := range frames {
		if localPlanetOccultationSVGFrameHasLabel(frame, "Greatest") {
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
	dx := frames[right].PlanetXArcsec - frames[left].PlanetXArcsec
	dy := frames[right].PlanetYArcsec - frames[left].PlanetYArcsec
	length := math.Hypot(dx, dy)
	if !starOccultationFinite(length) || length == 0 {
		return 0, 0, false
	}
	return dx / length, dy / length, true
}

func writeLocalPlanetOccultationDisk(
	b *strings.Builder,
	x, y, radius float64,
	hidden bool,
	class, label string,
) {
	fill, stroke, dash := "#d6ad69", "#704b2d", ""
	if hidden {
		fill, stroke, dash = "none", "#7d3430", ` stroke-dasharray="2 2"`
	}
	fmt.Fprintf(b, `<circle class="%s" data-label="%s" cx="%.3f" cy="%.3f" r="%.3f" fill="%s" stroke="%s" stroke-width="1.2"%s/>`,
		html.EscapeString(class), html.EscapeString(label), x, y, math.Max(radius, 0.15), fill, stroke, dash)
}

func writeLocalPlanetOccultationOverviewLabel(
	b *strings.Builder,
	event localPlanetOccultationEventFrame,
	x, y, cx, cy float64,
) {
	dx, dy, anchor := 8.0, -9.0, "start"
	if x > cx {
		dx, anchor = -8, "end"
	}
	if y < cy-20 {
		dy = 15
	}
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#792d28" stroke="#ffffff" stroke-width="3" paint-order="stroke" font-family="Arial, sans-serif" font-size="11" font-weight="700" text-anchor="%s">%s</text>`,
		x+dx, y+dy, anchor, html.EscapeString(event.name))
}

func writeLocalPlanetOccultationContacts(
	b *strings.Builder,
	events []localPlanetOccultationEventFrame,
	layout localStarOccultationSVGLayout,
	options LocalPlanetOccultationSVGOptions,
) {
	fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#d0d3d1" stroke-width="1"/>`,
		layout.panelX-12, layout.overviewTop-12, layout.panelX-12, layout.overviewBottom)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#161a1b" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700">%s</text>`,
		layout.panelX, layout.overviewTop-10, html.EscapeString(localPlanetOccultationSVGContactsTitle(options)))
	available := layout.overviewBottom - layout.overviewTop - 10
	rowHeight := math.Min(66, available/math.Max(1, float64(len(events))))
	for index, event := range events {
		y := layout.overviewTop + 16 + float64(index)*rowHeight
		if index > 0 {
			fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#e0e1df" stroke-width="0.8"/>`,
				layout.panelX, y-12, layout.panelX+layout.panelWidth, y-12)
		}
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#1c2528" font-family="Arial, sans-serif" font-size="11" font-weight="700">%s</text>`,
			layout.panelX, y, html.EscapeString(event.name))
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#1c2528" font-family="Arial, sans-serif" font-size="10" text-anchor="end">%s</text>`,
			layout.panelX+layout.panelWidth, y, html.EscapeString(event.time.In(options.Location).Format("15:04:05.0")))
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#536063" font-family="Arial, sans-serif" font-size="9.5">PA %.1f° | %s %.1f° | %s %.1f°</text>`,
			layout.panelX, y+16, event.frame.PositionAngleDeg,
			localStarOccultationSVGAltitudeLabel(options.Language), event.frame.MoonAltitudeDeg,
			localStarOccultationSVGAzimuthLabel(options.Language), event.frame.MoonAzimuthDeg)
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#6b7577" font-family="Arial, sans-serif" font-size="9.5">r %.2f″ | %s</text>`,
			layout.panelX, y+31, event.frame.PlanetRadiusArcsec,
			html.EscapeString(localStarOccultationSVGVisibilityText(event.frame.MoonAltitudeDeg >= 0, options.Language)))
	}
}

func writeLocalPlanetOccultationStages(
	b *strings.Builder,
	events []localPlanetOccultationEventFrame,
	layout localStarOccultationSVGLayout,
	options LocalPlanetOccultationSVGOptions,
) {
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#161a1b" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700">%s</text>`,
		layout.margin, layout.stageTop+13, html.EscapeString(localPlanetOccultationSVGPhasePanelsTitle(options)))
	if len(events) == 0 {
		return
	}
	usableWidth := layout.width - 2*layout.margin
	columnWidth := usableWidth / float64(len(events))
	maxMoonRadius := localPlanetOccultationSVGMaximumMoonRadiusFromEvents(events)
	radius := math.Min(43, math.Max(31, (layout.stageBottom-layout.stageTop-59)/2))
	scale := radius / maxMoonRadius
	cy := layout.stageTop + 35 + radius
	for index, event := range events {
		cx := layout.margin + columnWidth*(float64(index)+0.5)
		trueRadius := event.frame.PlanetRadiusArcsec * scale
		displayRadius := math.Max(trueRadius, 5.5)
		magnification := displayRadius / math.Max(trueRadius, 1e-9)
		x, y := localPlanetOccultationSVGStageCenter(event, cx, cy, radius, displayRadius, scale)
		fmt.Fprintf(b, `<circle class="stage-planet-disk" data-label="%s" data-true-radius="%.3f" data-magnification="%.3f" cx="%.3f" cy="%.3f" r="%.3f" fill="#d6ad69" stroke="#704b2d" stroke-width="1.1"/>`,
			html.EscapeString(event.label), trueRadius, magnification, x, y, displayRadius)
		writeLocalStarOccultationMoon(b, cx, cy, event.frame.MoonRadiusArcsec*scale, "stage-moon")
		if event.hidden {
			fmt.Fprintf(b, `<circle class="stage-planet-hidden-outline" cx="%.3f" cy="%.3f" r="%.3f" fill="none" stroke="#7d3430" stroke-width="1.1" stroke-dasharray="2 2"/>`,
				x, y, displayRadius)
		}
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#1c2528" font-family="Arial, sans-serif" font-size="10.5" font-weight="700" text-anchor="middle">%s</text>`,
			cx, cy+radius+18, html.EscapeString(event.name))
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#667174" font-family="Arial, sans-serif" font-size="9.5" text-anchor="middle">%s</text>`,
			cx, cy+radius+32, html.EscapeString(event.time.In(options.Location).Format("15:04:05.0")))
	}
}

// 阶段面板会放大较小的行星盘面以便观察，并在接触时同步移动显示中心，使 C1/C4 仍外切、C2/C3 仍内切。这只影响展示；事件时刻和总览使用未修改的站心几何。
// The stage panels enlarge small planet disks for visibility and move the display center by the same amount at contacts so C1/C4 remain externally tangent and C2/C3 remain internally tangent. This only affects presentation; event times and the overview use the unmodified topocentric geometry.
func localPlanetOccultationSVGStageCenter(
	event localPlanetOccultationEventFrame,
	cx, cy, moonRadius, displayRadius, scale float64,
) (float64, float64) {
	distance := event.frame.SeparationArcsec * scale
	switch event.label {
	case "C1", "C4":
		distance = moonRadius + displayRadius
	case "C2", "C3":
		distance = math.Max(0, moonRadius-displayRadius)
	}
	angle := event.frame.PositionAngleDeg * math.Pi / 180
	return cx - distance*math.Sin(angle), cy - distance*math.Cos(angle)
}

func localPlanetOccultationSVGMaximumMoonRadiusFromEvents(events []localPlanetOccultationEventFrame) float64 {
	radius := 1.0
	for _, event := range events {
		radius = math.Max(radius, event.frame.MoonRadiusArcsec)
	}
	return radius
}

func writeLocalPlanetOccultationFooter(
	b *strings.Builder,
	info moon.PlanetOccultationInfo,
	layout localStarOccultationSVGLayout,
	options LocalPlanetOccultationSVGOptions,
) {
	directionLines := starOccultationWrapText(localPlanetOccultationSVGDirectionText(options), layout.width-80, 10)
	for index, line := range directionLines {
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#4f595b" font-family="Georgia, 'Times New Roman', serif" font-size="10">%s</text>`,
			layout.margin, layout.footerY+float64(index)*14-8, html.EscapeString(line))
	}
	footerLines := starOccultationWrapText(localPlanetOccultationSVGFooterNote(info, options), layout.width-80, 9)
	for index, line := range footerLines {
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#747c7d" font-family="Georgia, 'Times New Roman', serif" font-size="9">%s</text>`,
			layout.margin, layout.footerY+float64(len(directionLines))*14+float64(index)*12-6, html.EscapeString(line))
	}
}

func localPlanetOccultationSVGEventFrames(
	info moon.PlanetOccultationInfo,
	frames []moon.PlanetOccultationDiagramFrame,
	language string,
) []localPlanetOccultationEventFrame {
	times := map[string]time.Time{
		"C1":       info.ExternalImmersion,
		"C2":       info.InternalImmersion,
		"Greatest": info.Greatest,
		"C3":       info.InternalEmersion,
		"C4":       info.ExternalEmersion,
	}
	labels := []string{"C1", "Greatest", "C4"}
	if info.HasInternalContacts {
		labels = []string{"C1", "C2", "Greatest", "C3", "C4"}
	}
	result := make([]localPlanetOccultationEventFrame, 0, len(labels))
	for _, label := range labels {
		for _, frame := range frames {
			if localPlanetOccultationSVGFrameHasLabel(frame, label) {
				hidden := frame.FullyOcculted || info.Type == moon.OccultationTotal &&
					(label == "C2" || label == "Greatest" || label == "C3")
				result = append(result, localPlanetOccultationEventFrame{
					label:  label,
					name:   localPlanetOccultationSVGEventName(label, language),
					time:   times[label],
					hidden: hidden,
					frame:  frame,
				})
				break
			}
		}
	}
	return result
}

func localPlanetOccultationSVGFrameHasLabel(frame moon.PlanetOccultationDiagramFrame, label string) bool {
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

func localPlanetOccultationSVGHeaderLines(
	info moon.PlanetOccultationInfo,
	options LocalPlanetOccultationSVGOptions,
) []string {
	lines := make([]string, 0, 4)
	for _, value := range []string{
		localPlanetOccultationSVGSummaryText(info, options),
		localPlanetOccultationSVGGreatestText(info, options),
	} {
		lines = append(lines, starOccultationWrapText(value, float64(options.Width)-80, 13)...)
	}
	return lines
}

func localPlanetOccultationSVGTargetName(
	info moon.PlanetOccultationInfo,
	language string,
) string {
	target := strings.TrimSpace(info.TargetID)
	if language == starOccultationSVGLanguageChinese && (target == "" || target == info.Planet.String()) {
		return planetOccultationChineseName(info.Planet)
	}
	if target != "" {
		return target
	}
	if language == starOccultationSVGLanguageEnglish {
		return info.Planet.String()
	}
	return "行星"
}

func localPlanetOccultationSVGTitle(
	info moon.PlanetOccultationInfo,
	options LocalPlanetOccultationSVGOptions,
) string {
	if options.Title != "" {
		return options.Title
	}
	date := info.Greatest.In(options.Location).Format("2006-01-02")
	target := localPlanetOccultationSVGTargetName(info, options.Language)
	if options.Language == starOccultationSVGLanguageEnglish {
		return fmt.Sprintf("%s Local Lunar Occultation of %s", date, target)
	}
	return fmt.Sprintf("%s 指定地点月掩%s", date, target)
}

func localPlanetOccultationSVGSummaryText(
	info moon.PlanetOccultationInfo,
	options LocalPlanetOccultationSVGOptions,
) string {
	if options.SummaryText != "" {
		return options.SummaryText
	}
	coordinates := starOccultationFormatCoordinates(info.Observer.Longitude, info.Observer.Latitude)
	duration := localStarOccultationSVGFormatDuration(info.ExternalEmersion.Sub(info.ExternalImmersion), options.Language)
	if options.Language == starOccultationSVGLanguageEnglish {
		return fmt.Sprintf("Site %s | elevation %.0f m | %s | C1-C4 duration %s", coordinates, info.Observer.Height,
			localPlanetOccultationSVGTypeName(info.Type, options.Language), duration)
	}
	return fmt.Sprintf("观测点 %s | 海拔 %.0f 米 | %s | C1-C4 历时 %s", coordinates, info.Observer.Height,
		localPlanetOccultationSVGTypeName(info.Type, options.Language), duration)
}

func localPlanetOccultationSVGGreatestText(
	info moon.PlanetOccultationInfo,
	options LocalPlanetOccultationSVGOptions,
) string {
	if options.GreatestText != "" {
		return options.GreatestText
	}
	value := info.Greatest.In(options.Location)
	zone := starOccultationLocationLabel(value, options.Location)
	if options.Language == starOccultationSVGLanguageEnglish {
		return fmt.Sprintf("Greatest %s %s | separation %.2f arcsec | radii Moon %.2f / planet %.2f arcsec | Moon altitude %+.1f°",
			value.Format("2006-01-02 15:04:05.0"), zone, info.MinimumSeparationArcsec,
			info.MoonSemidiameterArcsec, info.PlanetSemidiameterArcsec, info.MoonAltitudeAtGreatest)
	}
	return fmt.Sprintf("掩甚 %s %s | 中心角距 %.2f 角秒 | 视半径 月 %.2f / 行星 %.2f 角秒 | 月球高度 %+.1f°",
		value.Format("2006-01-02 15:04:05.0"), zone, info.MinimumSeparationArcsec,
		info.MoonSemidiameterArcsec, info.PlanetSemidiameterArcsec, info.MoonAltitudeAtGreatest)
}

func localPlanetOccultationSVGOverviewTitle(options LocalPlanetOccultationSVGOptions) string {
	if options.OverviewTitle != "" {
		return options.OverviewTitle
	}
	if options.Language == starOccultationSVGLanguageEnglish {
		return "Topocentric planet track (true disk scale)"
	}
	return "站心行星轨迹（圆盘真实比例）"
}

func localPlanetOccultationSVGPhasePanelsTitle(options LocalPlanetOccultationSVGOptions) string {
	if options.PhasePanelsTitle != "" {
		return options.PhasePanelsTitle
	}
	if options.Language == starOccultationSVGLanguageEnglish {
		return "C1-C4 contact stages (planet disk enlarged)"
	}
	return "C1-C4 接触阶段（行星圆盘放大示意）"
}

func localPlanetOccultationSVGContactsTitle(options LocalPlanetOccultationSVGOptions) string {
	if options.ContactsTitle != "" {
		return options.ContactsTitle
	}
	if options.Language == starOccultationSVGLanguageEnglish {
		return "Local contacts"
	}
	return "本地接触时刻"
}

func localPlanetOccultationSVGDirectionText(options LocalPlanetOccultationSVGOptions) string {
	if options.DirectionText != "" {
		return options.DirectionText
	}
	if options.Language == starOccultationSVGLanguageEnglish {
		return "Moon fixed at center; east is left and north is up. Blue-gray dashed: local lunar path. Red dashed: planet-center track."
	}
	return "月球固定在中心；图上左东右西，向上为北。灰蓝虚线为掩甚附近的站心白道，红色虚线为行星中心相对月心轨迹。"
}

func localPlanetOccultationSVGFooterNote(
	info moon.PlanetOccultationInfo,
	options LocalPlanetOccultationSVGOptions,
) string {
	if options.FooterNote != "" {
		return options.FooterNote
	}
	ringNote := ""
	if info.Planet == moon.OccultationSaturn {
		if options.Language == starOccultationSVGLanguageEnglish {
			ringNote = " Saturn's rings are excluded from contact calculation and drawing."
		} else {
			ringNote = "土星环不参与接触计算，也不作为圆盘边界绘制。"
		}
	}
	if options.Language == starOccultationSVGLanguageEnglish {
		return "C1/C4 are external contacts; C2/C3 are internal contacts. Stage disks are enlarged for visibility; contact times use true topocentric radii." + ringNote
	}
	return "C1/C4 为外切，C2/C3 为内切。阶段图中的行星圆盘为可见性放大，接触时刻按真实站心视半径求解。" + ringNote
}

func localPlanetOccultationSVGEventName(label, language string) string {
	if language == starOccultationSVGLanguageEnglish {
		switch label {
		case "C1":
			return "C1 external ingress"
		case "C2":
			return "C2 internal ingress"
		case "Greatest":
			return "Greatest"
		case "C3":
			return "C3 internal egress"
		case "C4":
			return "C4 external egress"
		}
	}
	switch label {
	case "C1":
		return "C1 外切始"
	case "C2":
		return "C2 内切始"
	case "Greatest":
		return "掩甚"
	case "C3":
		return "C3 内切终"
	case "C4":
		return "C4 外切终"
	default:
		return label
	}
}

func localPlanetOccultationSVGTypeName(eventType moon.OccultationType, language string) string {
	if language == starOccultationSVGLanguageEnglish {
		switch eventType {
		case moon.OccultationTotal:
			return "total occultation"
		case moon.OccultationPartial:
			return "partial occultation"
		default:
			return "grazing occultation"
		}
	}
	switch eventType {
	case moon.OccultationTotal:
		return "行星圆面全掩"
	case moon.OccultationPartial:
		return "行星圆面部分掩"
	default:
		return "行星圆面擦边"
	}
}
