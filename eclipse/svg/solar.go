package svg

import (
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro/basic"
	eclipsecore "github.com/starainrt/astro/eclipse"
)

const (
	localSolarEclipseSVGDefaultWidth  = 920
	localSolarEclipseSVGDefaultHeight = 720
	localSolarEclipseSVGDefaultZone   = 8 * 60 * 60

	localSolarEclipseSVGLanguageChinese = "zh"
	localSolarEclipseSVGLanguageEnglish = "en"
)

// LocalSolarEclipseSVGOptions 控制站心日食视圆 SVG 输出。
// LocalSolarEclipseSVGOptions controls local solar eclipse disk SVG output.
type LocalSolarEclipseSVGOptions struct {
	// Width / Height 是 SVG 画布尺寸；<=0 时使用默认尺寸。
	// Width/Height are SVG canvas size; values <= 0 use defaults.
	Width  int
	Height int
	// Step 是月心路径采样步长；<=0 时使用 5 分钟。
	// Step is the Moon-center path sampling step; values <= 0 use five minutes.
	Step time.Duration
	// Title 是图题；为空时自动生成。
	// Title is the chart title; empty values use an automatic title.
	Title string
	// SummaryText 是标题下第一行摘要；为空时自动生成。
	// SummaryText is the first summary line below the title; empty values use an automatic summary.
	SummaryText string
	// GreatestText 是标题下第二行食甚说明；为空时自动生成。
	// GreatestText is the second line below the title for greatest-eclipse details; empty values use automatic text.
	GreatestText string
	// MetaText 是标题下第三行补充说明；为空时自动生成沙罗/中心食历时等信息。
	// MetaText is the third line below the title; empty values use automatic Saros and central-duration text.
	MetaText string
	// OverviewTitle 是上方总览区标题；为空时自动生成。
	// OverviewTitle is the overview-panel title; empty values use an automatic title.
	OverviewTitle string
	// PhasePanelsTitle 是下方阶段视圆区标题；为空时自动生成。
	// PhasePanelsTitle is the phase-panels title; empty values use an automatic title.
	PhasePanelsTitle string
	// ContactsTitle 是右侧接触时刻区标题；为空时自动生成。
	// ContactsTitle is the contacts-panel title; empty values use an automatic title.
	ContactsTitle string
	// DirectionText 是底部方向说明；为空时自动生成。
	// DirectionText is the footer direction note; empty values use an automatic note.
	DirectionText string
	// FooterNote 是底部补充说明；为空时自动生成。
	// FooterNote is the footer explanatory note; empty values use an automatic note.
	FooterNote string
	// Language 是标签语言；"en" 使用英文，其他值或空值使用中文。
	// Language controls label language; "en" uses English, other values or empty values use Chinese.
	Language string
	// Location 是图中显示时刻的时区；nil 时使用 UTC+8。
	// Location is the display timezone for chart times; nil uses UTC+8.
	Location *time.Location
}

type localSolarEclipseSVGCalculator func(float64, float64, float64, float64, basic.LocalSolarEclipseDiagramOptions) basic.LocalSolarEclipseDiagramResult
type localSolarEclipseSVGFinder func(time.Time, float64, float64, float64) LocalSolarEclipseInfo

// LocalSolarEclipseSVG 生成给定地点的站心日食日月视圆 SVG，默认使用 NASA bulletin Split-K 模型。
// LocalSolarEclipseSVG generates an SVG local solar eclipse Sun-Moon disk chart, using NASA bulletin Split-K by default.
func LocalSolarEclipseSVG(
	date time.Time,
	lon, lat, height float64,
	options LocalSolarEclipseSVGOptions,
) (string, bool) {
	return LocalSolarEclipseSVGNASABulletinSplitK(date, lon, lat, height, options)
}

// LocalSolarEclipseSVGNASABulletinSplitK 生成站心日食 SVG，使用 NASA bulletin Split-K 模型。
// LocalSolarEclipseSVGNASABulletinSplitK generates a local solar eclipse SVG with the NASA bulletin Split-K model.
func LocalSolarEclipseSVGNASABulletinSplitK(
	date time.Time,
	lon, lat, height float64,
	options LocalSolarEclipseSVGOptions,
) (string, bool) {
	return localSolarEclipseSVG(date, lon, lat, height, options, basic.LocalSolarEclipseDiagramNASABulletinSplitK, eclipsecore.ClosestLocalSolarEclipseNASABulletinSplitK)
}

// LocalSolarEclipseSVGIAUSingleK 生成站心日食 SVG，使用 IAU Single-K 模型。
// LocalSolarEclipseSVGIAUSingleK generates a local solar eclipse SVG with the IAU Single-K model.
func LocalSolarEclipseSVGIAUSingleK(
	date time.Time,
	lon, lat, height float64,
	options LocalSolarEclipseSVGOptions,
) (string, bool) {
	return localSolarEclipseSVG(date, lon, lat, height, options, basic.LocalSolarEclipseDiagramIAUSingleK, eclipsecore.ClosestLocalSolarEclipseIAUSingleK)
}

func localSolarEclipseSVG(
	date time.Time,
	lon, lat, height float64,
	options LocalSolarEclipseSVGOptions,
	calculator localSolarEclipseSVGCalculator,
	finder localSolarEclipseSVGFinder,
) (string, bool) {
	options = normalizeLocalSolarEclipseSVGOptions(options)
	diagram := calculator(
		solarEclipseTimeToTTJDE(date),
		lon,
		lat,
		height,
		basic.LocalSolarEclipseDiagramOptions{StepDays: solarEclipseDurationToDays(options.Step)},
	)
	if diagram.Eclipse.Type == basic.SolarEclipseNone || len(diagram.Frames) == 0 {
		return "", false
	}
	info := localSolarEclipseInfoFromDiagram(diagram, lon, lat, height, options.Location)
	if finder != nil {
		coreInfo := finder(info.GreatestEclipse, lon, lat, height)
		info.HasSaros = coreInfo.HasSaros
		info.Saros = coreInfo.Saros
	}
	return renderLocalSolarEclipseSVG(info, diagram, options), true
}

func normalizeLocalSolarEclipseSVGOptions(options LocalSolarEclipseSVGOptions) LocalSolarEclipseSVGOptions {
	if options.Width <= 0 {
		options.Width = localSolarEclipseSVGDefaultWidth
	}
	if options.Height <= 0 {
		options.Height = localSolarEclipseSVGDefaultHeight
	}
	if options.Location == nil {
		options.Location = time.FixedZone("UTC+8", localSolarEclipseSVGDefaultZone)
	}
	if strings.EqualFold(options.Language, localSolarEclipseSVGLanguageEnglish) {
		options.Language = localSolarEclipseSVGLanguageEnglish
	} else {
		options.Language = localSolarEclipseSVGLanguageChinese
	}
	return options
}

func renderLocalSolarEclipseSVG(
	info LocalSolarEclipseInfo,
	diagram basic.LocalSolarEclipseDiagramResult,
	options LocalSolarEclipseSVGOptions,
) string {
	headerTexts := localSolarEclipseSVGHeaderTexts(info, options)
	headerBottom := localSolarEclipseSVGHeaderBottom(headerTexts)
	width := float64(options.Width)
	height := float64(options.Height)
	margin := math.Max(30, math.Min(46, width*0.05))
	panelWidth := math.Max(230, math.Min(280, width*0.28))
	diagramLeft := margin
	diagramRight := width - panelWidth - margin - 24
	if diagramRight-diagramLeft < width*0.48 {
		diagramRight = width - margin
	}
	footerHeight := 72.0
	stageHeight := math.Max(160, math.Min(210, height*0.27))
	stageTop := height - stageHeight - footerHeight
	if stageTop < headerBottom+160 {
		stageTop = headerBottom + 160
	}
	cx := (diagramLeft + diagramRight) / 2
	cy := headerBottom + (stageTop-headerBottom)/2 + 4
	extent := localSolarEclipseDiagramExtent(diagram)
	scale := math.Min((diagramRight-diagramLeft)/(2*extent), (stageTop-headerBottom-18)/(2*extent))
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		scale = 1
	}
	panelX := diagramRight + 24
	if panelX+panelWidth > width-margin/2 {
		panelX = width - panelWidth - margin/2
	}

	mapX := func(x float64) float64 { return cx - x*scale }
	mapY := func(y float64) float64 { return cy - y*scale }
	title := localSolarEclipseSVGTitleText(info, options)

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, options.Width, options.Height, options.Width, options.Height)
	b.WriteString(`<defs>`)
	b.WriteString(`<radialGradient id="se-sun" cx="40%" cy="34%" r="64%"><stop offset="0%" stop-color="#fff8bb"/><stop offset="58%" stop-color="#ffd55f"/><stop offset="100%" stop-color="#f5a623"/></radialGradient>`)
	b.WriteString(`</defs>`)
	b.WriteString(`<rect width="100%" height="100%" fill="#efefed"/>`)
	fmt.Fprintf(&b, `<rect x="22" y="18" width="%.3f" height="%.3f" fill="#ffffff" stroke="#c9c9c6" stroke-width="1.2"/>`,
		width-44, height-36)
	fmt.Fprintf(&b, `<text x="%.3f" y="44" fill="#111111" font-family="Georgia, 'Times New Roman', serif" font-size="26" font-weight="700" text-anchor="middle">%s</text>`,
		width/2, html.EscapeString(title))
	fmt.Fprintf(&b, `<line x1="%.3f" y1="57" x2="%.3f" y2="57" stroke="#555" stroke-width="1"/>`, width/2-78, width/2+78)
	for index, line := range headerTexts {
		fontSize := 13
		fill := "#333"
		if index == 0 {
			fontSize = 14
			fill = "#222"
		}
		fmt.Fprintf(&b, `<text x="%.3f" y="%.3f" fill="%s" font-family="Georgia, 'Times New Roman', serif" font-size="%d" text-anchor="middle">%s</text>`,
			width/2, localSolarEclipseSVGHeaderLineY(index), fill, fontSize, html.EscapeString(line))
	}

	eventFrames := localSolarEclipseSVGEventFrames(diagram.Frames)
	fmt.Fprintf(&b, `<text x="%.3f" y="%.3f" fill="#111" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700">%s</text>`,
		diagramLeft, headerBottom+10, html.EscapeString(localSolarEclipseSVGOverviewTitleText(options)))
	fmt.Fprintf(&b, `<circle cx="%.3f" cy="%.3f" r="%.3f" fill="url(#se-sun)" stroke="#c78211" stroke-width="1.4"/>`,
		cx, cy, scale)
	fmt.Fprintf(&b, `<circle cx="%.3f" cy="%.3f" r="%.3f" fill="none" stroke="#ffec92" stroke-opacity="0.58" stroke-width="7"/>`,
		cx, cy, scale+2)
	writeLocalSolarEclipseAxes(&b, cx, cy, scale, options.Language)
	writeLocalSolarEclipseEclipticLine(&b, diagram, mapX, mapY, extent, options.Language)

	if len(diagram.Frames) > 0 {
		b.WriteString(`<path d="`)
		for i, frame := range diagram.Frames {
			if i == 0 {
				fmt.Fprintf(&b, `M %.3f %.3f`, mapX(frame.MoonX), mapY(frame.MoonY))
				continue
			}
			fmt.Fprintf(&b, ` L %.3f %.3f`, mapX(frame.MoonX), mapY(frame.MoonY))
		}
		b.WriteString(`" fill="none" stroke="#555555" stroke-width="1.2" stroke-dasharray="5 4" stroke-linecap="round" stroke-linejoin="round"/>`)
	}

	for _, frame := range eventFrames {
		x := mapX(frame.MoonX)
		y := mapY(frame.MoonY)
		if localSolarEclipseSVGDrawOverviewMoon(frame.Label) {
			writeLocalSolarEclipseMoonOutline(&b, frame, x, y, scale)
			continue
		}
		writeLocalSolarEclipseEventPoint(&b, frame.Label, x, y, "#24518a")
	}
	writeLocalSolarEclipseContactMarkers(&b, info, cx, cy, scale, options.Language)
	for _, frame := range eventFrames {
		if frame.Label == "" {
			continue
		}
		x := mapX(frame.MoonX)
		y := mapY(frame.MoonY)
		writeLocalSolarEclipseEventLabel(&b, info.Type, frame.Label, x, y, cx, frame.MoonRadius*scale, options.Language)
	}
	writeLocalSolarEclipseStagePanels(&b, info, eventFrames, options, margin, stageTop, width-2*margin, stageHeight)

	fmt.Fprintf(&b, `<text x="%.3f" y="%.3f" fill="#333" font-family="Georgia, 'Times New Roman', serif" font-size="12">%s</text>`,
		40.0, height-54,
		html.EscapeString(localSolarEclipseSVGDirectionTextValue(options)))
	note := localSolarEclipseSVGFooterNoteText(options)
	fmt.Fprintf(&b, `<text x="%.3f" y="%.3f" fill="#555" font-family="Georgia, 'Times New Roman', serif" font-size="12">%s</text>`,
		40.0, height-34, html.EscapeString(note))
	writeLocalSolarEclipseContacts(&b, info, options, panelX, math.Max(154, cy-92))
	b.WriteString(`</svg>`)
	return b.String()
}

func localSolarEclipseSVGTitleText(info LocalSolarEclipseInfo, options LocalSolarEclipseSVGOptions) string {
	if options.Title != "" {
		return options.Title
	}
	return localSolarEclipseSVGTitle(info, options.Language)
}

func localSolarEclipseSVGHeaderTexts(info LocalSolarEclipseInfo, options LocalSolarEclipseSVGOptions) []string {
	lines := []string{
		localSolarEclipseSVGSummaryText(info, options),
		localSolarEclipseSVGGreatestTextValue(info, options),
		localSolarEclipseSVGMetaTextValue(info, options),
	}
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			filtered = append(filtered, line)
		}
	}
	return filtered
}

func localSolarEclipseSVGSummaryText(info LocalSolarEclipseInfo, options LocalSolarEclipseSVGOptions) string {
	if options.SummaryText != "" {
		return options.SummaryText
	}
	return localSolarEclipseSVGSummary(info, options.Language)
}

func localSolarEclipseSVGGreatestTextValue(info LocalSolarEclipseInfo, options LocalSolarEclipseSVGOptions) string {
	if options.GreatestText != "" {
		return options.GreatestText
	}
	return localSolarEclipseSVGGreatestText(info, options)
}

func localSolarEclipseSVGMetaTextValue(info LocalSolarEclipseInfo, options LocalSolarEclipseSVGOptions) string {
	if options.MetaText != "" {
		return options.MetaText
	}
	return localSolarEclipseSVGMetaText(info, options.Language)
}

func localSolarEclipseSVGOverviewTitleText(options LocalSolarEclipseSVGOptions) string {
	if options.OverviewTitle != "" {
		return options.OverviewTitle
	}
	return localSolarEclipseSVGOverviewTitle(options.Language)
}

func localSolarEclipseSVGPhasePanelsTitleText(options LocalSolarEclipseSVGOptions) string {
	if options.PhasePanelsTitle != "" {
		return options.PhasePanelsTitle
	}
	if options.Language == localSolarEclipseSVGLanguageEnglish {
		return "Phase disk panels"
	}
	return "阶段视圆图"
}

func localSolarEclipseSVGContactsTitleText(options LocalSolarEclipseSVGOptions) string {
	if options.ContactsTitle != "" {
		return options.ContactsTitle
	}
	if options.Language == localSolarEclipseSVGLanguageEnglish {
		return "Contacts"
	}
	return "接触时刻"
}

func localSolarEclipseSVGDirectionTextValue(options LocalSolarEclipseSVGOptions) string {
	if options.DirectionText != "" {
		return options.DirectionText
	}
	return localSolarEclipseSVGDirectionText(options.Language)
}

func localSolarEclipseSVGFooterNoteText(options LocalSolarEclipseSVGOptions) string {
	if options.FooterNote != "" {
		return options.FooterNote
	}
	if options.Language == localSolarEclipseSVGLanguageEnglish {
		return "Overview omits C2/C3 Moon outlines; lower panels show each phase separately. Contact PAs are measured from celestial north toward east."
	}
	return "上方为全局路径，C2/C3 只标点位；下方为各阶段独立视圆图。接触点位置角从天球北点起向东量。"
}

func localSolarEclipseSVGHeaderLineY(index int) float64 {
	return 86 + float64(index)*23
}

func localSolarEclipseSVGHeaderBottom(lines []string) float64 {
	if len(lines) == 0 {
		return 72
	}
	return localSolarEclipseSVGHeaderLineY(len(lines)-1) + 14
}

func localSolarEclipseSVGTitle(info LocalSolarEclipseInfo, language string) string {
	if language == localSolarEclipseSVGLanguageEnglish {
		return fmt.Sprintf("%s Local Solar Eclipse", info.GreatestEclipse.Format("2006-01-02"))
	}
	return fmt.Sprintf("%s 站心%s", info.GreatestEclipse.Format("2006-01-02"), localSolarEclipseSVGTypeName(info.Type, language))
}

func localSolarEclipseSVGSummary(info LocalSolarEclipseInfo, language string) string {
	if language == localSolarEclipseSVGLanguageEnglish {
		return fmt.Sprintf("lon=%.4f lat=%.4f type=%s magnitude=%.4f obscuration=%.4f",
			info.Longitude, info.Latitude, localSolarEclipseSVGTypeName(info.Type, language), info.Magnitude, info.Obscuration)
	}
	return fmt.Sprintf("经度=%.4f 纬度=%.4f 食型=%s 食分=%.4f 掩食比=%.4f",
		info.Longitude, info.Latitude, localSolarEclipseSVGTypeName(info.Type, language), info.Magnitude, info.Obscuration)
}

func localSolarEclipseSVGMetaText(info LocalSolarEclipseInfo, language string) string {
	parts := make([]string, 0, 2)
	if info.HasSaros {
		if language == localSolarEclipseSVGLanguageEnglish {
			parts = append(parts, fmt.Sprintf("Solar Saros %d  %d/%d", info.Saros.Series, info.Saros.Member, info.Saros.Count))
		} else {
			parts = append(parts, fmt.Sprintf("沙罗 %d  第 %d/%d 个成员", info.Saros.Series, info.Saros.Member, info.Saros.Count))
		}
	}
	if duration := localSolarEclipseSVGCentralDurationText(info, language); duration != "" {
		parts = append(parts, duration)
	}
	return strings.Join(parts, "   ")
}

func localSolarEclipseSVGDirectionText(language string) string {
	if language == localSolarEclipseSVGLanguageEnglish {
		return "Sun is fixed at center; Moon path uses the local tangent plane. East is left, north is up."
	}
	return "太阳固定在中心；月球路径使用站心切平面。图上左东右西，向上为北。"
}

func localSolarEclipseSVGGreatestText(info LocalSolarEclipseInfo, options LocalSolarEclipseSVGOptions) string {
	greatest := info.GreatestEclipse.In(options.Location).Format("2006-01-02 15:04:05 MST")
	constellation := localSolarEclipseSVGConstellationName(info, options.Language)
	if options.Language == localSolarEclipseSVGLanguageEnglish {
		return fmt.Sprintf("Greatest: %s  Sun altitude %.2f deg  Sun in %s", greatest, info.SunAltitude, constellation)
	}
	return fmt.Sprintf("食甚：%s  太阳高度 %.2f 度  太阳位于%s", greatest, info.SunAltitude, constellation)
}

func localSolarEclipseSVGCentralDurationText(info LocalSolarEclipseInfo, language string) string {
	if !info.HasCentral || info.CentralStart.IsZero() || info.CentralEnd.IsZero() {
		return ""
	}
	duration := lunarEclipseSVGFormatDuration(info.CentralEnd.Sub(info.CentralStart))
	if language == localSolarEclipseSVGLanguageEnglish {
		switch info.Type {
		case SolarEclipseTotal:
			return "Totality " + duration
		case SolarEclipseAnnular:
			return "Annularity " + duration
		default:
			return "Central phase " + duration
		}
	}
	switch info.Type {
	case SolarEclipseTotal:
		return "全食历时 " + duration
	case SolarEclipseAnnular:
		return "环食历时 " + duration
	default:
		return "中心食历时 " + duration
	}
}

func localSolarEclipseSVGOverviewTitle(language string) string {
	if language == localSolarEclipseSVGLanguageEnglish {
		return "Overview path"
	}
	return "全局路径"
}

func localSolarEclipseSVGOverviewEventName(label, language string) string {
	if label == "Greatest" {
		if language == localSolarEclipseSVGLanguageEnglish {
			return "GE"
		}
		return "食甚"
	}
	return label
}

func localSolarEclipseSVGTypeName(eclipseType SolarEclipseType, language string) string {
	if language == localSolarEclipseSVGLanguageEnglish {
		switch eclipseType {
		case SolarEclipsePartial:
			return "partial"
		case SolarEclipseAnnular:
			return "annular"
		case SolarEclipseTotal:
			return "total"
		case SolarEclipseHybrid:
			return "hybrid"
		default:
			return "none"
		}
	}
	switch eclipseType {
	case SolarEclipsePartial:
		return "日偏食"
	case SolarEclipseAnnular:
		return "日环食"
	case SolarEclipseTotal:
		return "日全食"
	case SolarEclipseHybrid:
		return "全环食"
	default:
		return "无日食"
	}
}

func localSolarEclipseSVGEventFrames(frames []basic.LocalSolarEclipseDiagramFrame) []basic.LocalSolarEclipseDiagramFrame {
	events := make([]basic.LocalSolarEclipseDiagramFrame, 0, 5)
	for _, frame := range frames {
		for _, label := range localSolarEclipseSVGFrameLabels(frame) {
			event := frame
			event.Label = label
			event.Labels = []string{label}
			events = append(events, event)
		}
	}
	return events
}

func localSolarEclipseSVGFrameLabels(frame basic.LocalSolarEclipseDiagramFrame) []string {
	if len(frame.Labels) > 0 {
		return frame.Labels
	}
	if frame.Label == "" {
		return nil
	}
	return []string{frame.Label}
}

func writeLocalSolarEclipseAxes(b *strings.Builder, cx, cy, radius float64, language string) {
	north, east, west, south := "北", "东", "西", "南"
	if language == localSolarEclipseSVGLanguageEnglish {
		north, east, west, south = "N", "E", "W", "S"
	}
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#111" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700" text-anchor="middle">%s</text>`,
		cx, cy-radius-17, html.EscapeString(north))
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#111" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700" text-anchor="middle">%s</text>`,
		cx-radius-20, cy+4, html.EscapeString(east))
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#111" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700" text-anchor="middle">%s</text>`,
		cx+radius+20, cy+4, html.EscapeString(west))
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#111" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700" text-anchor="middle">%s</text>`,
		cx, cy+radius+27, html.EscapeString(south))
}

func writeLocalSolarEclipseMoon(
	b *strings.Builder,
	frame basic.LocalSolarEclipseDiagramFrame,
	x, y, scale float64,
) {
	radius := frame.MoonRadius * scale
	fmt.Fprintf(b, `<circle class="stage-moon" cx="%.3f" cy="%.3f" r="%.3f" fill="#050505" stroke="#111111" stroke-width="1.1"/>`,
		x, y, radius)
}

func writeLocalSolarEclipseMoonOutline(
	b *strings.Builder,
	frame basic.LocalSolarEclipseDiagramFrame,
	x, y, scale float64,
) {
	radius := frame.MoonRadius * scale
	stroke := "#24518a"
	if frame.Label == "Greatest" {
		stroke = "#111111"
	}
	fmt.Fprintf(b, `<circle class="event-moon" data-label="%s" cx="%.3f" cy="%.3f" r="%.3f" fill="none" stroke="%s" stroke-width="1.3" stroke-dasharray="5 4"/>`,
		html.EscapeString(frame.Label), x, y, radius, stroke)
	writeLocalSolarEclipseEventPoint(b, frame.Label, x, y, stroke)
}

func localSolarEclipseSVGDrawOverviewMoon(label string) bool {
	return label != "C2" && label != "C3"
}

func writeLocalSolarEclipseEventPoint(b *strings.Builder, label string, x, y float64, stroke string) {
	fmt.Fprintf(b, `<circle class="event-center" data-label="%s" cx="%.3f" cy="%.3f" r="2.6" fill="#ffffff" stroke="%s" stroke-width="1.2"/>`,
		html.EscapeString(label), x, y, stroke)
}

func writeLocalSolarEclipseContactMarkers(
	b *strings.Builder,
	info LocalSolarEclipseInfo,
	cx, cy, radius float64,
	language string,
) {
	for _, point := range info.ContactPoints {
		angle := point.ContactPositionAngle * math.Pi / 180
		unitX := math.Sin(angle)
		unitY := math.Cos(angle)
		x := cx - radius*unitX
		y := cy - radius*unitY
		labelDistance := radius + localSolarEclipseSVGContactLabelDistance(radius)
		labelDistance += localSolarEclipseSVGContactLabelExtraDistance(unitX, unitY)
		labelX := cx - labelDistance*unitX
		labelY := cy - labelDistance*unitY + 4
		sideShift := localSolarEclipseSVGContactLabelSideShift(unitX, unitY)
		labelX += unitY * sideShift
		labelY += unitX * sideShift
		anchor := "middle"
		if unitX > 0.28 {
			anchor = "end"
		} else if unitX < -0.28 {
			anchor = "start"
		}
		fmt.Fprintf(b, `<line class="contact-leader" x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#c96b6b" stroke-width="0.8" opacity="0.58"/>`,
			x, y, labelX, labelY-4)
		fmt.Fprintf(b, `<circle class="contact-point" cx="%.3f" cy="%.3f" r="3.2" fill="#b51616" stroke="#ffffff" stroke-width="0.8"/>`,
			x, y)
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#b51616" font-family="Georgia, 'Times New Roman', serif" font-size="11" font-weight="700" text-anchor="%s">%s</text>`,
			labelX, labelY, anchor, html.EscapeString(localSolarEclipseSVGContactLabel(point, language)))
	}
}

func localSolarEclipseSVGContactLabel(point LocalSolarEclipseContactPoint, language string) string {
	if language == localSolarEclipseSVGLanguageEnglish {
		return fmt.Sprintf("%s %.0f°", point.Label, point.ContactPositionAngle)
	}
	return fmt.Sprintf("%s %.0f°", point.Label, point.ContactPositionAngle)
}

func writeLocalSolarEclipseEclipticLine(
	b *strings.Builder,
	diagram basic.LocalSolarEclipseDiagramResult,
	mapX, mapY func(float64) float64,
	extent float64,
	language string,
) {
	unitX, unitY, ok := localSolarEclipseSVGEclipticDirection(diagram.Eclipse.GreatestEclipse)
	if !ok {
		return
	}
	lineExtent := extent * 0.92
	startX := mapX(-unitX * lineExtent)
	startY := mapY(-unitY * lineExtent)
	endX := mapX(unitX * lineExtent)
	endY := mapY(unitY * lineExtent)
	if math.IsNaN(startX) || math.IsNaN(startY) || math.IsNaN(endX) || math.IsNaN(endY) {
		return
	}

	fmt.Fprintf(b, `<line class="ecliptic-line" x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#666" stroke-width="1" stroke-dasharray="4 3" opacity="0.82"/>`,
		startX, startY, endX, endY)

	labelX := startX
	labelY := startY
	anchor := "end"
	if endX < startX {
		labelX = endX
		labelY = endY
	}
	if labelX < endX {
		anchor = "start"
	}
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#444" font-family="Georgia, 'Times New Roman', serif" font-size="12" text-anchor="%s">%s</text>`,
		labelX, labelY-6, anchor, html.EscapeString(localSolarEclipseSVGLabelEcliptic(language)))
}

func localSolarEclipseSVGEclipticDirection(ttJDE float64) (float64, float64, bool) {
	originRA, originDec := basic.HSunApparentRaDec(ttJDE)
	centerLongitude := normalizeSolarEclipseDegree360(basic.HSunApparentLo(ttJDE))
	ra1, dec1 := basic.LoBoToRaDec(ttJDE, centerLongitude-1, 0)
	ra2, dec2 := basic.LoBoToRaDec(ttJDE, centerLongitude+1, 0)
	x1, y1 := localSolarEclipseSVGTangentOffset(originRA, originDec, ra1, dec1)
	x2, y2 := localSolarEclipseSVGTangentOffset(originRA, originDec, ra2, dec2)
	dx := x2 - x1
	dy := y2 - y1
	length := math.Hypot(dx, dy)
	if length == 0 || math.IsNaN(length) || math.IsInf(length, 0) {
		return 0, 0, false
	}
	return dx / length, dy / length, true
}

func localSolarEclipseSVGTangentOffset(originRA, originDec, targetRA, targetDec float64) (float64, float64) {
	separation := localSolarEclipseSVGAngularSeparation(originRA, originDec, targetRA, targetDec)
	positionAngleRad := localSolarEclipseSVGPositionAngle(originRA, originDec, targetRA, targetDec) * math.Pi / 180
	return separation * math.Sin(positionAngleRad), separation * math.Cos(positionAngleRad)
}

func localSolarEclipseSVGAngularSeparation(ra1, dec1, ra2, dec2 float64) float64 {
	ra1Rad := ra1 * math.Pi / 180
	dec1Rad := dec1 * math.Pi / 180
	ra2Rad := ra2 * math.Pi / 180
	dec2Rad := dec2 * math.Pi / 180
	cosDistance := math.Sin(dec1Rad)*math.Sin(dec2Rad) +
		math.Cos(dec1Rad)*math.Cos(dec2Rad)*math.Cos(ra2Rad-ra1Rad)
	if cosDistance > 1 {
		cosDistance = 1
	}
	if cosDistance < -1 {
		cosDistance = -1
	}
	return math.Acos(cosDistance) * 180 / math.Pi
}

func localSolarEclipseSVGPositionAngle(fromRA, fromDec, toRA, toDec float64) float64 {
	dRA := (toRA - fromRA) * math.Pi / 180
	fromDecRad := fromDec * math.Pi / 180
	toDecRad := toDec * math.Pi / 180
	angle := math.Atan2(
		math.Sin(dRA),
		math.Cos(fromDecRad)*math.Tan(toDecRad)-math.Sin(fromDecRad)*math.Cos(dRA),
	) * 180 / math.Pi
	return normalizeSolarEclipseDegree360(angle)
}

func localSolarEclipseSVGLabelEcliptic(language string) string {
	if language == localSolarEclipseSVGLanguageEnglish {
		return "Ecliptic"
	}
	return "黄道"
}

func writeLocalSolarEclipseEventLabel(
	b *strings.Builder,
	_ SolarEclipseType,
	label string,
	x, y, cx, moonRadius float64,
	language string,
) {
	text := localSolarEclipseSVGOverviewEventName(label, language)
	dx, dy, anchor := localSolarEclipseSVGEventLabelLayout(label, x, cx, moonRadius)
	lineX := x + dx*0.8
	lineY := y + dy*0.8 - 2
	fmt.Fprintf(b, `<line class="event-leader" x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#24518a" stroke-width="0.8" stroke-dasharray="3 3" opacity="0.72"/>`,
		x, y, lineX, lineY)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#174f8a" font-family="Georgia, 'Times New Roman', serif" font-size="12" font-weight="700" text-anchor="%s">%s</text>`,
		x+dx, y+dy, anchor, html.EscapeString(text))
}

func localSolarEclipseSVGEventLabelLayout(label string, x, cx, moonRadius float64) (float64, float64, string) {
	dx := moonRadius + 18
	anchor := "start"
	if x > cx {
		dx = -dx
		anchor = "end"
	}
	dy := -moonRadius*0.62 - 10
	switch label {
	case "C1":
		dx = -(moonRadius*1.12 + 26)
		dy = -(moonRadius*0.84 + 16)
		anchor = "end"
	case "C2":
		dx = -(moonRadius*0.96 + 18)
		dy = -(moonRadius*0.56 + 14)
		anchor = "end"
	case "C3":
		dx = moonRadius*1.04 + 24
		dy = moonRadius*0.42 + 16
		anchor = "start"
	case "C4":
		dx = moonRadius*0.94 + 20
		dy = moonRadius*0.58 + 18
		anchor = "start"
	case "Greatest":
		dx = -(moonRadius*0.42 + 16)
		dy = moonRadius + 24
		anchor = "end"
	}
	return dx, dy, anchor
}

func localSolarEclipseSVGConstellationName(info LocalSolarEclipseInfo, language string) string {
	jde := solarEclipseTimeToTTJDE(info.GreatestEclipse)
	ra, dec := basic.HSunApparentRaDec(jde)
	code := basic.ConstellationCode(ra, dec, jde)
	if language == localSolarEclipseSVGLanguageEnglish {
		return basic.ConstellationNameByCodeEN(code)
	}
	return basic.ConstellationNameByCodeZH(code)
}

func localSolarEclipseSVGContactLabelDistance(radius float64) float64 {
	return math.Max(40, math.Min(58, radius*0.86))
}

func localSolarEclipseSVGContactLabelExtraDistance(unitX, unitY float64) float64 {
	extra := 8.0
	if unitY > 0.72 {
		extra += 10
	}
	if math.Abs(unitX) < 0.18 {
		extra += 8
	}
	if math.Abs(unitX) > 0.82 {
		extra += 4
	}
	return extra
}

func localSolarEclipseSVGContactLabelSideShift(unitX, unitY float64) float64 {
	if unitY > 0.72 {
		return 14
	}
	if unitX > 0.82 {
		return 18
	}
	if unitX < -0.82 {
		return 12
	}
	if unitY < -0.72 {
		return 6
	}
	return 0
}

func localSolarEclipseSVGEventName(label, language string, eclipseType SolarEclipseType) string {
	if language == localSolarEclipseSVGLanguageEnglish {
		switch label {
		case "C2":
			switch eclipseType {
			case SolarEclipseTotal:
				return "C2 Total begins"
			case SolarEclipseAnnular:
				return "C2 Annularity begins"
			default:
				return "C2 Central begins"
			}
		case "C3":
			switch eclipseType {
			case SolarEclipseTotal:
				return "C3 Total ends"
			case SolarEclipseAnnular:
				return "C3 Annularity ends"
			default:
				return "C3 Central ends"
			}
		}
		return label
	}
	switch label {
	case "C1":
		return "C1 初亏"
	case "C2":
		switch eclipseType {
		case SolarEclipseTotal:
			return "C2 食既"
		case SolarEclipseAnnular:
			return "C2 环食始"
		default:
			return "C2 中心食始"
		}
	case "Greatest":
		return "食甚"
	case "C3":
		switch eclipseType {
		case SolarEclipseTotal:
			return "C3 生光"
		case SolarEclipseAnnular:
			return "C3 环食终"
		default:
			return "C3 中心食终"
		}
	case "C4":
		return "C4 复圆"
	default:
		return label
	}
}

func writeLocalSolarEclipseStagePanels(
	b *strings.Builder,
	info LocalSolarEclipseInfo,
	frames []basic.LocalSolarEclipseDiagramFrame,
	options LocalSolarEclipseSVGOptions,
	x, y, width, height float64,
) {
	if len(frames) == 0 {
		return
	}
	title := localSolarEclipseSVGPhasePanelsTitleText(options)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#111" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700">%s</text>`,
		x, y+14, html.EscapeString(title))

	gap := 10.0
	panelCount := float64(len(frames))
	panelWidth := (width - gap*(panelCount-1)) / panelCount
	if panelWidth < 74 {
		gap = 6
		panelWidth = (width - gap*(panelCount-1)) / panelCount
	}
	panelTop := y + 24
	panelHeight := height - 30
	panelScale := localSolarEclipseStageScaleForFrames(frames, panelWidth, panelHeight)
	contactPoints := localSolarEclipseContactPointMap(info.ContactPoints)
	for index, frame := range frames {
		panelX := x + float64(index)*(panelWidth+gap)
		writeLocalSolarEclipseStagePanel(b, info, frame, options, contactPoints, panelX, panelTop, panelWidth, panelHeight, panelScale)
	}
}

func writeLocalSolarEclipseStagePanel(
	b *strings.Builder,
	info LocalSolarEclipseInfo,
	frame basic.LocalSolarEclipseDiagramFrame,
	options LocalSolarEclipseSVGOptions,
	contactPoints map[string]LocalSolarEclipseContactPoint,
	x, y, width, height, scale float64,
) {
	fmt.Fprintf(b, `<rect x="%.3f" y="%.3f" width="%.3f" height="%.3f" fill="#fbfbf8" stroke="#d8d2c4" stroke-width="1"/>`,
		x, y, width, height)
	label := localSolarEclipseSVGEventName(frame.Label, options.Language, info.Type)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#111" font-family="Georgia, 'Times New Roman', serif" font-size="12" font-weight="700" text-anchor="middle">%s</text>`,
		x+width/2, y+18, html.EscapeString(label))

	centerX := x + width/2
	centerY := y + 30 + (height-62)/2
	moonX := centerX - frame.MoonX*scale
	moonY := centerY - frame.MoonY*scale

	fmt.Fprintf(b, `<circle cx="%.3f" cy="%.3f" r="%.3f" fill="url(#se-sun)" stroke="#c78211" stroke-width="1"/>`,
		centerX, centerY, scale)
	fmt.Fprintf(b, `<circle cx="%.3f" cy="%.3f" r="%.3f" fill="none" stroke="#ffec92" stroke-opacity="0.55" stroke-width="3"/>`,
		centerX, centerY, scale+1.2)
	writeLocalSolarEclipseMoon(b, frame, moonX, moonY, scale)
	if point, ok := contactPoints[frame.Label]; ok {
		writeLocalSolarEclipseStageContactMarker(b, point, centerX, centerY, scale)
	}

	if eventTime, ok := localSolarEclipseSVGEventTime(info, frame.Label); ok {
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#444" font-family="Georgia, 'Times New Roman', serif" font-size="11" text-anchor="middle">%s</text>`,
			x+width/2, y+height-12, html.EscapeString(eventTime.In(options.Location).Format("15:04:05")))
	}
}

func localSolarEclipseStageScaleForFrames(frames []basic.LocalSolarEclipseDiagramFrame, width, height float64) float64 {
	availableX := width/2 - 12
	availableY := (height - 62) / 2
	if availableX < 12 {
		availableX = 12
	}
	if availableY < 12 {
		availableY = 12
	}
	extentX := 1.15
	extentY := 1.15
	for _, frame := range frames {
		candidateX := math.Abs(frame.MoonX) + frame.MoonRadius + 0.14
		candidateY := math.Abs(frame.MoonY) + frame.MoonRadius + 0.14
		if candidateX > extentX {
			extentX = candidateX
		}
		if candidateY > extentY {
			extentY = candidateY
		}
	}
	scale := math.Min(48, math.Min(availableX/extentX, availableY/extentY))
	if scale < 13 {
		scale = 13
	}
	return scale
}

func writeLocalSolarEclipseStageContactMarker(
	b *strings.Builder,
	point LocalSolarEclipseContactPoint,
	cx, cy, radius float64,
) {
	angle := point.ContactPositionAngle * math.Pi / 180
	x := cx - radius*math.Sin(angle)
	y := cy - radius*math.Cos(angle)
	fmt.Fprintf(b, `<circle class="stage-contact-point" cx="%.3f" cy="%.3f" r="2.5" fill="#b51616" stroke="#ffffff" stroke-width="0.8"/>`,
		x, y)
}

func localSolarEclipseSVGEventTime(info LocalSolarEclipseInfo, label string) (time.Time, bool) {
	switch label {
	case "C1":
		return info.PartialStart, !info.PartialStart.IsZero()
	case "C2":
		return info.CentralStart, !info.CentralStart.IsZero()
	case "Greatest":
		return info.GreatestEclipse, !info.GreatestEclipse.IsZero()
	case "C3":
		return info.CentralEnd, !info.CentralEnd.IsZero()
	case "C4":
		return info.PartialEnd, !info.PartialEnd.IsZero()
	default:
		return time.Time{}, false
	}
}

type localSolarEclipseSVGContact struct {
	label    string
	name     string
	time     time.Time
	angle    float64
	hasAngle bool
}

func writeLocalSolarEclipseContacts(
	b *strings.Builder,
	info LocalSolarEclipseInfo,
	options LocalSolarEclipseSVGOptions,
	x, y float64,
) {
	contacts := localSolarEclipseSVGContacts(info, options.Language)
	if len(contacts) == 0 {
		return
	}
	title := localSolarEclipseSVGContactsTitleText(options)
	boxWidth := float64(options.Width) - x - 34
	if boxWidth < 210 {
		boxWidth = 210
	}
	if boxWidth > 260 {
		boxWidth = 260
	}
	boxHeight := 27 + float64(len(contacts))*18
	fmt.Fprintf(b, `<rect x="%.3f" y="%.3f" width="%.3f" height="%.3f" fill="#fbfbf8" stroke="#d8d2c4" stroke-width="1"/>`,
		x-12, y-20, boxWidth, boxHeight)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#111111" font-family="Georgia, 'Times New Roman', serif" font-size="13" font-weight="700">%s (%s)</text>`,
		x, y, html.EscapeString(title), html.EscapeString(options.Location.String()))
	for index, contact := range contacts {
		line := fmt.Sprintf("%s %s  %s", contact.label, contact.name, contact.time.In(options.Location).Format("15:04:05"))
		if contact.hasAngle {
			if options.Language == localSolarEclipseSVGLanguageEnglish {
				line = fmt.Sprintf("%s  PA %.1f°", line, contact.angle)
			} else {
				line = fmt.Sprintf("%s  方位 %.1f°", line, contact.angle)
			}
		}
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#333333" font-family="Georgia, 'Times New Roman', serif" font-size="11.5">%s</text>`,
			x, y+float64(index+1)*17.5, html.EscapeString(line))
	}
}

func localSolarEclipseSVGContacts(info LocalSolarEclipseInfo, language string) []localSolarEclipseSVGContact {
	angles := localSolarEclipseContactAngleMap(info.ContactPoints)
	contacts := []localSolarEclipseSVGContact{
		localSolarEclipseSVGContactFor("C1", localSolarEclipseSVGContactName("C1", language, info.Type), info.PartialStart, angles),
	}
	if info.HasCentral {
		contacts = append(contacts, localSolarEclipseSVGContactFor("C2", localSolarEclipseSVGContactName("C2", language, info.Type), info.CentralStart, angles))
	}
	contacts = append(contacts, localSolarEclipseSVGContact{label: "GE", name: localSolarEclipseSVGContactName("Greatest", language, info.Type), time: info.GreatestEclipse})
	if info.HasCentral {
		contacts = append(contacts, localSolarEclipseSVGContactFor("C3", localSolarEclipseSVGContactName("C3", language, info.Type), info.CentralEnd, angles))
	}
	contacts = append(contacts, localSolarEclipseSVGContactFor("C4", localSolarEclipseSVGContactName("C4", language, info.Type), info.PartialEnd, angles))
	return contacts
}

func localSolarEclipseSVGContactFor(
	label, name string,
	time time.Time,
	angles map[string]float64,
) localSolarEclipseSVGContact {
	angle, ok := angles[label]
	return localSolarEclipseSVGContact{
		label:    label,
		name:     name,
		time:     time,
		angle:    angle,
		hasAngle: ok,
	}
}

func localSolarEclipseContactAngleMap(points []LocalSolarEclipseContactPoint) map[string]float64 {
	angles := make(map[string]float64, len(points))
	for _, point := range points {
		angles[point.Label] = point.ContactPositionAngle
	}
	return angles
}

func localSolarEclipseContactPointMap(points []LocalSolarEclipseContactPoint) map[string]LocalSolarEclipseContactPoint {
	contacts := make(map[string]LocalSolarEclipseContactPoint, len(points))
	for _, point := range points {
		contacts[point.Label] = point
	}
	return contacts
}

func localSolarEclipseSVGContactName(label, language string, eclipseType SolarEclipseType) string {
	if language == localSolarEclipseSVGLanguageEnglish {
		switch label {
		case "C1":
			return "First contact"
		case "C2":
			switch eclipseType {
			case SolarEclipseTotal:
				return "Total begins"
			case SolarEclipseAnnular:
				return "Annularity begins"
			default:
				return "Central begins"
			}
		case "Greatest":
			return "Greatest"
		case "C3":
			switch eclipseType {
			case SolarEclipseTotal:
				return "Total ends"
			case SolarEclipseAnnular:
				return "Annularity ends"
			default:
				return "Central ends"
			}
		case "C4":
			return "Last contact"
		default:
			return label
		}
	}
	switch label {
	case "C1":
		return "初亏"
	case "C2":
		switch eclipseType {
		case SolarEclipseTotal:
			return "食既"
		case SolarEclipseAnnular:
			return "环食始"
		default:
			return "中心食始"
		}
	case "Greatest":
		return "食甚"
	case "C3":
		switch eclipseType {
		case SolarEclipseTotal:
			return "生光"
		case SolarEclipseAnnular:
			return "环食终"
		default:
			return "中心食终"
		}
	case "C4":
		return "复圆"
	default:
		return label
	}
}

func localSolarEclipseDiagramExtent(diagram basic.LocalSolarEclipseDiagramResult) float64 {
	extent := 1.45
	for _, frame := range diagram.Frames {
		candidate := math.Hypot(frame.MoonX, frame.MoonY) + frame.MoonRadius + 0.28
		if candidate > extent {
			extent = candidate
		}
	}
	return extent
}

func solarEclipseDurationToDays(duration time.Duration) float64 {
	if duration <= 0 {
		return 0
	}
	return duration.Hours() / 24
}
