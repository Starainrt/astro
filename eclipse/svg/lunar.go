package svg

import (
	"github.com/starainrt/astro/internal/timenote"
	"fmt"
	"html"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/svgasset"
	"github.com/starainrt/astro/internal/svgchart"
)

const (
	lunarEclipseSVGDefaultWidth  = 960
	lunarEclipseSVGDefaultHeight = 620
	lunarEclipseSVGDefaultZone   = 8 * 60 * 60

	lunarEclipseSVGLanguageChinese = "zh"
	lunarEclipseSVGLanguageEnglish = "en"
)

// LunarEclipseSVGOptions 控制月食穿影 SVG 输出。
// LunarEclipseSVGOptions controls lunar eclipse shadow-path SVG output.
type LunarEclipseSVGOptions struct {
	// Width 和 Height 是 SVG 画布尺寸；<=0 时使用默认尺寸。
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
	// MaximumText 是标题下第二行食甚说明；为空时自动生成。
	// MaximumText is the second line below the title for maximum-eclipse details; empty values use automatic text.
	MaximumText string
	// CoordinatesText 是标题下第三行坐标说明；为空时自动生成。
	// CoordinatesText is the third line below the title for coordinates; empty values use automatic text.
	CoordinatesText string
	// DurationText 是标题下第四行历时说明；为空时自动生成。
	// DurationText is the fourth line below the title for durations; empty values use automatic text.
	DurationText string
	// MetaText 是标题下第五行补充说明；为空时自动生成沙罗信息。
	// MetaText is the fifth line below the title; empty values use automatic Saros text.
	MetaText string
	// ContactsTitle 是接触时刻区标题；为空时自动生成。
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
	// TimeScale 选择图中时刻的时标：零值 UTC；TimeScaleUT1 改用 UT1 时刻，此时 Location 必须是
	// nil 或 UTC（否则返回 false）。nil Location 在 UTC 下默认 UTC+8，UT1 下强制为 UTC。
	// TimeScale selects the label scale: the zero value is UTC; TimeScaleUT1 uses UT1 labels and requires
	// Location to be nil or UTC (otherwise the renderer returns false).
	TimeScale astro.TimeScale
}

type lunarEclipseSVGCalculator func(float64, basic.LunarEclipseDiagramOptions) basic.LunarEclipseDiagramResult
type lunarEclipseSVGFinder func(time.Time) LunarEclipseInfo

type lunarEclipseSVGLayout struct {
	width        float64
	height       float64
	cx           float64
	cy           float64
	scale        float64
	diagramLeft  float64
	diagramRight float64
	diagramTop   float64
	diagramEnd   float64
	panelX       float64
	panelY       float64
}

type lunarEclipseSVGPoint struct {
	basic.LunarEclipseDiagramPoint
	X float64
	Y float64
}

type lunarEclipseSVGMaximumCoordinates struct {
	RA                float64
	Dec               float64
	EclipticLongitude float64
	EclipticLatitude  float64
	ConstellationCode string
	ConstellationName string
}

// LunarEclipseSVG 生成月食穿影图 SVG，默认模型与 LunarEclipseOnDate 一致：Danjon 为主，极浅半影回退 Chauvenet。
// LunarEclipseSVG generates an SVG lunar eclipse shadow-path chart with the default model of
// LunarEclipseOnDate: Danjon, falling back to Chauvenet for ultra-shallow penumbral cases.
func LunarEclipseSVG(date time.Time, options LunarEclipseSVGOptions) (string, bool) {
	model, ok := lunarEclipseDefaultShadowModel(date)
	if !ok {
		return "", false
	}
	if model == eclipsecore.LunarEclipseShadowModelChauvenet {
		return LunarEclipseSVGChauvenet(date, options)
	}
	return LunarEclipseSVGDanjon(date, options)
}

// lunarEclipseDefaultShadowModel 返回默认模型下该日期的月食影半径模型；当天没有月食时返回 false。
func lunarEclipseDefaultShadowModel(date time.Time) (eclipsecore.LunarEclipseShadowModel, bool) {
	info, ok := eclipsecore.LunarEclipseOnDate(date)
	if !ok {
		return eclipsecore.LunarEclipseShadowModelDanjon, false
	}
	return info.ShadowModel, true
}

// lunarEclipseShadowModelName 给出影半径模型的展示名。
func lunarEclipseShadowModelName(model eclipsecore.LunarEclipseShadowModel) string {
	if model == eclipsecore.LunarEclipseShadowModelChauvenet {
		return "Chauvenet"
	}
	return "Danjon"
}

// LunarEclipseSVGDanjon 生成月食穿影图 SVG，使用 Danjon 影半径模型。
// LunarEclipseSVGDanjon generates an SVG lunar eclipse shadow-path chart with the Danjon shadow model.
func LunarEclipseSVGDanjon(date time.Time, options LunarEclipseSVGOptions) (string, bool) {
	return lunarEclipseSVG(date, options, basic.LunarEclipseDiagramDanjon, eclipsecore.ClosestLunarEclipseDanjon)
}

// LunarEclipseSVGChauvenet 生成月食穿影图 SVG，使用 Chauvenet 影半径模型。
// LunarEclipseSVGChauvenet generates an SVG lunar eclipse shadow-path chart with the Chauvenet shadow model.
func LunarEclipseSVGChauvenet(date time.Time, options LunarEclipseSVGOptions) (string, bool) {
	return lunarEclipseSVG(date, options, basic.LunarEclipseDiagramChauvenet, eclipsecore.ClosestLunarEclipseChauvenet)
}

func lunarEclipseSVG(
	date time.Time,
	options LunarEclipseSVGOptions,
	calculator lunarEclipseSVGCalculator,
	finder lunarEclipseSVGFinder,
) (string, bool) {
	if options.TimeScale == astro.TimeScaleUT1 {
		if options.Location != nil && options.Location != time.UTC {
			return "", false
		}
		options.Location = time.UTC
	}
	options = normalizeLunarEclipseSVGOptions(options)
	diagram := calculator(timeToTTJDE(date), basic.LunarEclipseDiagramOptions{
		StepDays: durationToDays(options.Step),
	})
	if diagram.Eclipse.Type == basic.LunarEclipseNone || len(diagram.Points) == 0 {
		return "", false
	}
	info := lunarEclipseInfoFromBasic(diagram.Eclipse, options.Location)
	if finder != nil {
		coreInfo := finder(info.Maximum)
		info.HasSaros = coreInfo.HasSaros
		info.Saros = coreInfo.Saros
	}
	return renderLunarEclipseSVG(lunarEclipseDisplayInfo(info, options.TimeScale), info, diagram, options), true
}

func normalizeLunarEclipseSVGOptions(options LunarEclipseSVGOptions) LunarEclipseSVGOptions {
	if options.Width <= 0 {
		options.Width = lunarEclipseSVGDefaultWidth
	}
	if options.Height <= 0 {
		options.Height = lunarEclipseSVGDefaultHeight
	}
	if options.Location == nil {
		options.Location = time.FixedZone("UTC+8", lunarEclipseSVGDefaultZone)
	}
	if strings.EqualFold(options.Language, lunarEclipseSVGLanguageEnglish) {
		options.Language = lunarEclipseSVGLanguageEnglish
	} else {
		options.Language = lunarEclipseSVGLanguageChinese
	}
	return options
}

// writeLunarEclipseDiagram 在给定圆心与比例下画地影、月心轨迹与各阶段月面。
// 独立月食图和 详细版式组合图共用这一段，只有圆心的位置不同。
func writeLunarEclipseDiagram(
	b *strings.Builder,
	layout lunarEclipseSVGLayout,
	diagram basic.LunarEclipseDiagramResult,
	language string,
	labels *svgchart.LabelTable,
) {
	cx, cy, scale := layout.cx, layout.cy, layout.scale
	points := lunarEclipseSVGPoints(diagram.Points)
	mapX := func(x float64) float64 { return cx - x*scale }
	mapY := func(y float64) float64 { return cy - y*scale }
	area := svgchart.LabelBox{
		X: layout.diagramLeft, Y: layout.diagramTop,
		Width: layout.diagramRight - layout.diagramLeft, Height: layout.diagramEnd - layout.diagramTop,
	}
	fmt.Fprintf(b, `<circle cx="%.3f" cy="%.3f" r="%.3f" fill="#e9e9e9" stroke="#d6d6d6" stroke-width="1.2"/>`,
		cx, cy, diagram.PenumbraRadius*scale)
	fmt.Fprintf(b, `<circle cx="%.3f" cy="%.3f" r="%.3f" fill="#b21f16" stroke="#83170f" stroke-width="1.3"/>`,
		cx, cy, diagram.UmbraRadius*scale)
	writeLunarEclipseShadowLabels(b, layout, diagram, language, labels, area)
	writeLunarEclipseAxes(b, cx, cy, diagram.PenumbraRadius*scale, language, labels, area)
	writeLunarEclipseEclipticLine(b, layout, diagram, mapX, mapY, language, labels, area)

	if len(points) > 0 {
		b.WriteString(`<path d="`)
		for i, point := range points {
			if i == 0 {
				fmt.Fprintf(b, `M %.3f %.3f`, mapX(point.X), mapY(point.Y))
				continue
			}
			fmt.Fprintf(b, ` L %.3f %.3f`, mapX(point.X), mapY(point.Y))
		}
		b.WriteString(`" fill="none" stroke="#333333" stroke-width="1.1" stroke-dasharray="5 4" stroke-linecap="round" stroke-linejoin="round"/>`)
	}

	eventPoints := lunarEclipseSVGEventPoints(points)
	for _, point := range eventPoints {
		if point.Label == "Greatest" {
			continue
		}
		writeLunarEclipseMoon(b, point, diagram, mapX(point.X), mapY(point.Y), scale, false)
	}
	for _, point := range eventPoints {
		if point.Label == "Greatest" {
			writeLunarEclipseMoon(b, point, diagram, mapX(point.X), mapY(point.Y), scale, true)
			break
		}
	}
	writeLunarEclipseEventLabels(b, eventPoints, mapX, mapY, cx, diagram.MoonRadius*scale, language, labels, area)
}

// lunarEclipseEventLabelRank 是事件标注的占位优先级：数字小的先占位，放不下的后放的让位。
func lunarEclipseEventLabelRank(label string) int {
	switch label {
	case "Greatest":
		return 0
	case "U2", "U3":
		return 1
	case "U1", "U4":
		return 2
	case "P1", "P4":
		return 3
	}
	return 4
}

// writeLunarEclipseEventLabels 先按重要性占位再按原顺序绘制：食甚最优先，其次本影接触，最后半影接触。
func writeLunarEclipseEventLabels(
	b *strings.Builder,
	eventPoints []lunarEclipseSVGPoint,
	mapX, mapY func(float64) float64,
	cx, moonRadius float64,
	language string,
	labels *svgchart.LabelTable,
	area svgchart.LabelBox,
) {
	type eventLabel struct {
		rank       int
		text       string
		candidates []svgchart.LabelPlacement
		placed     svgchart.LabelPlacement
		ok         bool
	}
	entries := make([]eventLabel, 0, len(eventPoints))
	for _, point := range eventPoints {
		if point.Label == "" {
			continue
		}
		x := mapX(point.X)
		y := mapY(point.Y)
		dx := moonRadius*0.72 + 6
		anchor := "start"
		if x < cx {
			dx = -dx
			anchor = "end"
		}
		dy := -moonRadius*0.35 - 4
		if point.Label == "Greatest" {
			dy = moonRadius + 15
			anchor = "middle"
			dx = 0
		}
		candidates := []svgchart.LabelPlacement{{X: x + dx, Y: y + dy, Anchor: anchor}}
		// 浅半影食里 P1 与 P4 挨得很近，近环放不下时还要一圈更远的候选。
		candidates = append(candidates, svgchart.LabelCandidates(x+dx, y+dy, anchor, math.Max(14, moonRadius*0.8))...)
		candidates = append(candidates, svgchart.LabelCandidates(x+dx, y+dy, anchor, math.Max(30, moonRadius*1.6))...)
		candidates = append(candidates, svgchart.LabelCandidates(x+dx, y+dy, anchor, math.Max(48, moonRadius*2.4))...)
		entries = append(entries, eventLabel{
			rank:       lunarEclipseEventLabelRank(point.Label),
			text:       lunarEclipseSVGEventName(point.Label, language),
			candidates: candidates,
		})
	}
	order := make([]int, len(entries))
	for index := range entries {
		order[index] = index
	}
	sort.SliceStable(order, func(i, j int) bool { return entries[order[i]].rank < entries[order[j]].rank })
	for _, index := range order {
		entry := &entries[index]
		entry.placed, entry.ok = labels.Place(entry.text, 12,
			labelCandidatesInsideBox(entry.candidates, entry.text, 12, area))
	}
	for _, entry := range entries {
		if !entry.ok {
			continue
		}
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#2554c7" font-family="Georgia, 'Times New Roman', serif" font-size="12" text-anchor="%s">%s</text>`,
			entry.placed.X, entry.placed.Y, entry.placed.Anchor, html.EscapeString(entry.text))
	}
}

func renderLunarEclipseSVG(
	info LunarEclipseInfo,
	geometry LunarEclipseInfo,
	diagram basic.LunarEclipseDiagramResult,
	options LunarEclipseSVGOptions,
) string {
	headerTexts := lunarEclipseSVGHeaderTexts(info, geometry, options)
	headerBottom := lunarEclipseSVGHeaderBottom(headerTexts)
	direction := lunarEclipseSVGDirectionTextValue(options)
	if options.DirectionText != "" {
		direction = svgchart.EllipsizeText(direction, float64(options.Width)-80, 12)
	}
	footerLines := lunarEclipseSVGFooterLines(direction, info, options, headerBottom)
	footerBaselines, footerOccupied := svgchart.FooterBlock(len(footerLines), float64(options.Height), 12,
		svgchart.FooterLineHeight(12), localDiagramFooterBottomPadding)
	// 页脚高度按实际行数算，再留一点与图区的空，说明文字不会顶到画布底边。
	layout := lunarEclipseSVGLayoutFor(diagram, options, headerBottom, footerOccupied+10)
	title := lunarEclipseSVGTitleText(info, options)

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, options.Width, options.Height, options.Width, options.Height)
	b.WriteString(`<defs>`)
	b.WriteString(svgasset.MoonFaceSymbol())
	b.WriteString(`</defs>`)
	b.WriteString(`<rect width="100%" height="100%" fill="#efefed"/>`)
	fmt.Fprintf(&b, `<rect x="22" y="18" width="%.3f" height="%.3f" fill="#ffffff" stroke="#c9c9c6" stroke-width="1.2"/>`,
		layout.width-44, layout.height-36)
	titleText := title
	if options.Title != "" {
		titleText = svgchart.EllipsizeText(title, layout.width-52, 26)
	}
	fmt.Fprintf(&b, `<text x="%.3f" y="44" fill="#111111" font-family="Georgia, 'Times New Roman', serif" font-size="26" font-weight="700" text-anchor="middle">%s</text>`,
		layout.width/2, html.EscapeString(titleText))
	fmt.Fprintf(&b, `<line x1="%.3f" y1="57" x2="%.3f" y2="57" stroke="#555" stroke-width="1"/>`, layout.width/2-78, layout.width/2+78)
	writeLunarEclipseSummary(&b, headerTexts, options)
	// 穿影示意图的标注要避开已经写好的表头与右侧接触时刻面板。
	labels := &svgchart.LabelTable{}
	labels.Reserve(0, 0, layout.width, layout.diagramTop)
	labels.Reserve(layout.diagramRight, layout.diagramTop,
		layout.width-layout.diagramRight, layout.diagramEnd-layout.diagramTop)
	labels.Reserve(0, layout.diagramEnd, layout.width, layout.height-layout.diagramEnd)
	writeLunarEclipseDiagram(&b, layout, diagram, options.Language, labels)

	writeLunarEclipseContacts(&b, info, options, layout.panelX, layout.panelY)
	writeLunarEclipseFooter(&b, footerLines, footerBaselines)

	b.WriteString(`</svg>`)
	return b.String()
}

func lunarEclipseSVGLayoutFor(
	diagram basic.LunarEclipseDiagramResult,
	options LunarEclipseSVGOptions,
	headerBottom, footerHeight float64,
) lunarEclipseSVGLayout {
	width := float64(options.Width)
	height := float64(options.Height)
	margin := math.Max(32, math.Min(48, width*0.05))
	panelWidth := math.Max(210, math.Min(260, width*0.24))
	diagramLeft := margin
	diagramRight := width - panelWidth - 34
	if diagramRight-diagramLeft < width*0.48 {
		diagramRight = width - margin
	}
	topReserved := math.Max(166, headerBottom+24)
	bottomReserved := footerHeight
	extent := diagram.PenumbraRadius + diagram.MoonRadius + 0.72
	scale := math.Min((diagramRight-diagramLeft)/(2*extent), (height-topReserved-bottomReserved)/(2*extent))
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		scale = 1
	}
	cx := (diagramLeft + diagramRight) / 2
	cy := topReserved + (height-topReserved-bottomReserved)/2 + 12
	panelX := diagramRight + 22
	if panelX+panelWidth > width-margin/2 {
		panelX = width - panelWidth - margin/2
	}
	if panelX < diagramRight {
		panelX = diagramRight
	}
	return lunarEclipseSVGLayout{
		width:        width,
		height:       height,
		cx:           cx,
		cy:           cy,
		scale:        scale,
		diagramLeft:  diagramLeft,
		diagramRight: diagramRight,
		diagramTop:   topReserved,
		diagramEnd:   height - bottomReserved,
		panelX:       panelX,
		panelY:       math.Max(148, cy-88),
	}
}

func lunarEclipseSVGTitle(info LunarEclipseInfo, language string) string {
	return fmt.Sprintf("%s %s", info.Maximum.Format("2006-01-02"), lunarEclipseSVGTypeName(info.Type, language))
}

func lunarEclipseSVGTitleText(info LunarEclipseInfo, options LunarEclipseSVGOptions) string {
	if options.Title != "" {
		return options.Title
	}
	return lunarEclipseSVGTitle(info, options.Language)
}

func lunarEclipseSVGHeaderTexts(
	info LunarEclipseInfo,
	geometry LunarEclipseInfo,
	options LunarEclipseSVGOptions,
) []string {
	lines := []string{
		lunarEclipseSVGSummaryText(info, options),
		lunarEclipseSVGMaximumTextValue(info, options),
		lunarEclipseSVGCoordinatesTextValue(geometry, options),
		lunarEclipseSVGDurationTextValue(info, options),
		lunarEclipseSVGMetaTextValue(info, options),
	}
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			filtered = append(filtered, line)
		}
	}
	return filtered
}

func lunarEclipseSVGSummaryText(info LunarEclipseInfo, options LunarEclipseSVGOptions) string {
	if options.SummaryText != "" {
		return options.SummaryText
	}
	return lunarEclipseSVGSummary(info, options.Language)
}

func lunarEclipseSVGMaximumTextValue(info LunarEclipseInfo, options LunarEclipseSVGOptions) string {
	if options.MaximumText != "" {
		return options.MaximumText
	}
	return lunarEclipseSVGMaximumText(info, options)
}

func lunarEclipseSVGCoordinatesTextValue(info LunarEclipseInfo, options LunarEclipseSVGOptions) string {
	if options.CoordinatesText != "" {
		return options.CoordinatesText
	}
	coordinates := lunarEclipseSVGMaximumCoordinatesFor(info, options.Language)
	return lunarEclipseSVGMaximumCoordinatesText(coordinates, options.Language)
}

func lunarEclipseSVGDurationTextValue(info LunarEclipseInfo, options LunarEclipseSVGOptions) string {
	if options.DurationText != "" {
		return options.DurationText
	}
	return lunarEclipseSVGDurationSummary(info, options.Language)
}

func lunarEclipseSVGMetaTextValue(info LunarEclipseInfo, options LunarEclipseSVGOptions) string {
	if options.MetaText != "" {
		return options.MetaText
	}
	return lunarEclipseSVGMetaText(info, options.Language)
}

func lunarEclipseSVGContactsTitleText(options LunarEclipseSVGOptions) string {
	if options.ContactsTitle != "" {
		return options.ContactsTitle
	}
	if options.Language == lunarEclipseSVGLanguageEnglish {
		return "Contacts"
	}
	return "接触时刻"
}

func lunarEclipseSVGDirectionTextValue(options LunarEclipseSVGOptions) string {
	if options.DirectionText != "" {
		return options.DirectionText
	}
	return lunarEclipseSVGDirectionText(options.Language)
}

func lunarEclipseSVGFooterNoteText(options LunarEclipseSVGOptions) string {
	return lunarEclipseSVGFooterNoteTextBase(options)
}

func lunarEclipseSVGFooterNoteTextBase(options LunarEclipseSVGOptions) string {
	if options.FooterNote != "" {
		return options.FooterNote
	}
	note := "图中月面大小和影半径均按实际相对角半径缩放。"
	if options.Language == lunarEclipseSVGLanguageEnglish {
		note = "Moon disks and shadow radii are drawn to the same relative angular-radius scale."
	}
	return note
}

func lunarEclipseSVGHeaderLineY(index int) float64 {
	return 84 + float64(index)*22
}

func lunarEclipseSVGHeaderBottom(lines []string) float64 {
	if len(lines) == 0 {
		return 72
	}
	return lunarEclipseSVGHeaderLineY(len(lines)-1) + 14
}

func lunarEclipseSVGSummary(info LunarEclipseInfo, language string) string {
	if language == lunarEclipseSVGLanguageEnglish {
		return fmt.Sprintf("type=%s  penumbral=%.4f  umbral=%.4f",
			lunarEclipseSVGTypeName(info.Type, language), info.PenumbralMagnitude, info.UmbralMagnitude)
	}
	return fmt.Sprintf("食型=%s  半影食分=%.4f  本影食分=%.4f",
		lunarEclipseSVGTypeName(info.Type, language), info.PenumbralMagnitude, info.UmbralMagnitude)
}

func writeLunarEclipseSummary(b *strings.Builder, lines []string, options LunarEclipseSVGOptions) {
	for index, line := range lines {
		fontSize := 13
		fill := "#333"
		if index == 0 {
			fontSize = 14
			fill = "#222"
		}
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="%s" font-family="Georgia, 'Times New Roman', serif" font-size="%d" text-anchor="middle">%s</text>`,
			float64(options.Width)/2, lunarEclipseSVGHeaderLineY(index), fill, fontSize, html.EscapeString(line))
	}
}

func lunarEclipseSVGDirectionText(language string) string {
	if language == lunarEclipseSVGLanguageEnglish {
		return "North is up and east is left; the ecliptic is projected near greatest eclipse."
	}
	return "上北下南，左东右西；黄道按食甚附近天球投影绘制。"
}

func lunarEclipseSVGMaximumText(info LunarEclipseInfo, options LunarEclipseSVGOptions) string {
	maximum := info.Maximum.In(options.Location).Format("2006-01-02 15:04:05 MST")
	if options.Language == lunarEclipseSVGLanguageEnglish {
		return "Maximum: " + maximum
	}
	return "食甚：" + maximum
}

func lunarEclipseSVGMaximumCoordinatesFor(
	info LunarEclipseInfo,
	language string,
) lunarEclipseSVGMaximumCoordinates {
	jde := timeToTTJDE(info.Maximum)
	ra, dec := basic.HMoonTrueRaDec(jde)
	lon := basic.HMoonApparentLo(jde)
	lat := basic.HMoonTrueBo(jde)
	code := basic.ConstellationCode(ra, dec, jde)
	name := basic.ConstellationNameByCodeZH(code)
	if language == lunarEclipseSVGLanguageEnglish {
		name = basic.ConstellationNameByCodeEN(code)
	}
	return lunarEclipseSVGMaximumCoordinates{
		RA:                ra,
		Dec:               dec,
		EclipticLongitude: lon,
		EclipticLatitude:  lat,
		ConstellationCode: code,
		ConstellationName: name,
	}
}

func lunarEclipseSVGMaximumCoordinatesText(
	coordinates lunarEclipseSVGMaximumCoordinates,
	language string,
) string {
	if language == lunarEclipseSVGLanguageEnglish {
		return fmt.Sprintf("Moon: RA %s  Dec %s  ecl.lon %.4f deg  ecl.lat %.4f deg  %s",
			lunarEclipseSVGFormatRA(coordinates.RA),
			lunarEclipseSVGFormatSignedAngle(coordinates.Dec),
			coordinates.EclipticLongitude,
			coordinates.EclipticLatitude,
			coordinates.ConstellationName,
		)
	}
	return fmt.Sprintf("月球：赤经 %s  赤纬 %s  黄经 %.4f°  黄纬 %.4f°  %s",
		lunarEclipseSVGFormatRA(coordinates.RA),
		lunarEclipseSVGFormatSignedAngle(coordinates.Dec),
		coordinates.EclipticLongitude,
		coordinates.EclipticLatitude,
		coordinates.ConstellationName,
	)
}

func lunarEclipseSVGFormatRA(degree float64) string {
	totalSeconds := int(math.Round(normalizeDegree360(degree) / 15 * 3600))
	totalSeconds %= 24 * 3600
	hours := totalSeconds / 3600
	minutes := totalSeconds % 3600 / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02dh%02dm%02ds", hours, minutes, seconds)
}

func lunarEclipseSVGFormatSignedAngle(degree float64) string {
	sign := "+"
	if degree < 0 {
		sign = "-"
		degree = -degree
	}
	totalSeconds := int(math.Round(degree * 3600))
	degrees := totalSeconds / 3600
	minutes := totalSeconds % 3600 / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%s%02d°%02d′%02d″", sign, degrees, minutes, seconds)
}

func lunarEclipseSVGDurationSummary(info LunarEclipseInfo, language string) string {
	penumbral := lunarEclipseSVGFormatDuration(info.PenumbralEnd.Sub(info.PenumbralStart))
	if language == lunarEclipseSVGLanguageEnglish {
		parts := []string{"Penumbral duration " + penumbral}
		if info.HasPartial {
			parts = append(parts, "Umbral duration "+lunarEclipseSVGFormatDuration(info.PartialEnd.Sub(info.PartialStart)))
		}
		if info.HasTotal {
			parts = append(parts, "Total duration "+lunarEclipseSVGFormatDuration(info.TotalEnd.Sub(info.TotalStart)))
		}
		return strings.Join(parts, "   ")
	}
	parts := []string{"半影历时 " + penumbral}
	if info.HasPartial {
		parts = append(parts, "本影历时 "+lunarEclipseSVGFormatDuration(info.PartialEnd.Sub(info.PartialStart)))
	}
	if info.HasTotal {
		parts = append(parts, "全食历时 "+lunarEclipseSVGFormatDuration(info.TotalEnd.Sub(info.TotalStart)))
	}
	return strings.Join(parts, "   ")
}

func lunarEclipseSVGMetaText(info LunarEclipseInfo, language string) string {
	if !info.HasSaros {
		return ""
	}
	if language == lunarEclipseSVGLanguageEnglish {
		label := fmt.Sprintf("Lunar Saros %d  %d/%d", info.Saros.Series, info.Saros.Member, info.Saros.Count)
		if !info.Saros.Verified {
			label += " (provisional)"
		}
		return label
	}
	label := fmt.Sprintf("沙罗 %d  第 %d/%d 个成员", info.Saros.Series, info.Saros.Member, info.Saros.Count)
	if !info.Saros.Verified {
		label += "（推算）"
	}
	return label
}

func lunarEclipseSVGFormatDuration(duration time.Duration) string {
	if duration < 0 {
		duration = -duration
	}
	totalSeconds := int(duration.Round(time.Second).Seconds())
	hours := totalSeconds / 3600
	minutes := totalSeconds % 3600 / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func lunarEclipseSVGTypeName(eclipseType LunarEclipseType, language string) string {
	if language == lunarEclipseSVGLanguageEnglish {
		switch eclipseType {
		case LunarEclipsePenumbral:
			return "Penumbral Lunar Eclipse"
		case LunarEclipsePartial:
			return "Partial Lunar Eclipse"
		case LunarEclipseTotal:
			return "Total Lunar Eclipse"
		default:
			return "No Lunar Eclipse"
		}
	}
	switch eclipseType {
	case LunarEclipsePenumbral:
		return "半影月食"
	case LunarEclipsePartial:
		return "月偏食"
	case LunarEclipseTotal:
		return "月全食"
	default:
		return "无月食"
	}
}

func writeLunarEclipseAxes(
	b *strings.Builder,
	cx, cy, radius float64,
	language string,
	labels *svgchart.LabelTable,
	area svgchart.LabelBox,
) {
	north, east, west, south := "北", "东", "西", "南"
	if language == lunarEclipseSVGLanguageEnglish {
		north, east, west, south = "N", "E", "W", "S"
	}
	entries := [4]struct {
		text string
		x, y float64
	}{
		{text: north, x: cx, y: cy - radius - 18},
		{text: east, x: cx - radius - 22, y: cy + 4},
		{text: west, x: cx + radius + 22, y: cy + 4},
		{text: south, x: cx, y: cy + radius + 28},
	}
	for _, entry := range entries {
		candidates := svgchart.LabelCandidates(entry.x, entry.y, "middle", 15)
		placed, ok := labels.Place(entry.text, 14, labelCandidatesInsideBox(candidates, entry.text, 14, area))
		if !ok {
			continue
		}
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#111" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700" text-anchor="%s">%s</text>`,
			placed.X, placed.Y, placed.Anchor, html.EscapeString(entry.text))
	}
}

func lunarEclipseSVGPoints(points []basic.LunarEclipseDiagramPoint) []lunarEclipseSVGPoint {
	result := make([]lunarEclipseSVGPoint, 0, len(points))
	for _, point := range points {
		distance := math.Hypot(point.X, point.Y)
		positionAngleRad := lunarEclipseMoonCenterPositionAngle(point.JDE) * math.Pi / 180
		result = append(result, lunarEclipseSVGPoint{
			LunarEclipseDiagramPoint: point,
			X:                        distance * math.Sin(positionAngleRad),
			Y:                        distance * math.Cos(positionAngleRad),
		})
	}
	return result
}

func writeLunarEclipseEclipticLine(
	b *strings.Builder,
	layout lunarEclipseSVGLayout,
	diagram basic.LunarEclipseDiagramResult,
	mapX, mapY func(float64) float64,
	language string,
	labels *svgchart.LabelTable,
	area svgchart.LabelBox,
) {
	unitX, unitY, ok := lunarEclipseSVGEclipticDirection(diagram.Eclipse.Maximum)
	if !ok {
		return
	}

	extent := (diagram.PenumbraRadius + diagram.MoonRadius) * 1.42
	screenStartX := mapX(-unitX * extent)
	screenStartY := mapY(-unitY * extent)
	screenEndX := mapX(unitX * extent)
	screenEndY := mapY(unitY * extent)
	if math.IsNaN(screenStartX) || math.IsNaN(screenStartY) || math.IsNaN(screenEndX) || math.IsNaN(screenEndY) {
		return
	}

	fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#555" stroke-width="1" stroke-dasharray="4 3" opacity="0.82"/>`,
		screenStartX, screenStartY, screenEndX, screenEndY)

	labelX := screenStartX
	labelY := screenStartY
	anchor := "end"
	if screenEndX < screenStartX {
		labelX = screenEndX
		labelY = screenEndY
	}
	labelX -= 8
	if labelX < layout.diagramLeft+18 {
		labelX += 16
		anchor = "start"
	}
	text := lunarEclipseSVGLabelEcliptic(language)
	candidates := append([]svgchart.LabelPlacement{{X: labelX, Y: labelY + 4, Anchor: anchor}},
		svgchart.LabelCandidates(labelX, labelY+4, anchor, 14)...)
	placed, ok := labels.Place(text, 12, labelCandidatesInsideBox(candidates, text, 12, area))
	if !ok {
		return
	}
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#444" font-family="Georgia, 'Times New Roman', serif" font-size="12" text-anchor="%s">%s</text>`,
		placed.X, placed.Y, placed.Anchor, html.EscapeString(text))
}

func lunarEclipseSVGEclipticDirection(ttJDE float64) (float64, float64, bool) {
	originRA, originDec := lunarEclipseShadowCenterRaDec(ttJDE)
	centerLongitude := normalizeDegree360(basic.HSunApparentLo(ttJDE) + 180)
	ra1, dec1 := basic.LoBoToRaDec(ttJDE, centerLongitude-1, 0)
	ra2, dec2 := basic.LoBoToRaDec(ttJDE, centerLongitude+1, 0)
	x1, y1 := lunarEclipseSVGTangentOffset(originRA, originDec, ra1, dec1)
	x2, y2 := lunarEclipseSVGTangentOffset(originRA, originDec, ra2, dec2)
	dx := x2 - x1
	dy := y2 - y1
	length := math.Hypot(dx, dy)
	if length == 0 || math.IsNaN(length) || math.IsInf(length, 0) {
		return 0, 0, false
	}
	return dx / length, dy / length, true
}

func lunarEclipseSVGTangentOffset(originRA, originDec, targetRA, targetDec float64) (float64, float64) {
	separation := lunarEclipseSVGAngularSeparation(originRA, originDec, targetRA, targetDec)
	positionAngleRad := lunarEclipsePositionAngle(originRA, originDec, targetRA, targetDec) * math.Pi / 180
	return separation * math.Sin(positionAngleRad), separation * math.Cos(positionAngleRad)
}

func lunarEclipseSVGAngularSeparation(ra1, dec1, ra2, dec2 float64) float64 {
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

func lunarEclipseSVGLabelEcliptic(language string) string {
	if language == lunarEclipseSVGLanguageEnglish {
		return "Ecliptic"
	}
	return "黄道"
}

func writeLunarEclipseShadowLabels(
	b *strings.Builder,
	layout lunarEclipseSVGLayout,
	diagram basic.LunarEclipseDiagramResult,
	language string,
	labels *svgchart.LabelTable,
	area svgchart.LabelBox,
) {
	penumbraLabel, umbraLabel := "地球半影", "地球本影"
	if language == lunarEclipseSVGLanguageEnglish {
		penumbraLabel, umbraLabel = "Earth's Penumbra", "Earth's Umbra"
	}
	entries := [2]struct {
		text   string
		x, y   float64
		fill   string
		weight string
	}{
		{text: penumbraLabel, x: layout.cx, y: layout.cy - diagram.PenumbraRadius*layout.scale + 28, fill: "#333", weight: "700"},
		{text: umbraLabel, x: layout.cx, y: layout.cy - diagram.UmbraRadius*layout.scale + 22, fill: "#111", weight: "700"},
	}
	for _, entry := range entries {
		candidates := svgchart.LabelCandidates(entry.x, entry.y, "middle", 14)
		placed, ok := labels.Place(entry.text, 13, labelCandidatesInsideBox(candidates, entry.text, 13, area))
		if !ok {
			continue
		}
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="%s" font-family="Georgia, 'Times New Roman', serif" font-size="13" font-weight="%s" text-anchor="%s">%s</text>`,
			placed.X, placed.Y, entry.fill, entry.weight, placed.Anchor, html.EscapeString(entry.text))
	}
}

func lunarEclipseSVGEventPoints(points []lunarEclipseSVGPoint) []lunarEclipseSVGPoint {
	events := make([]lunarEclipseSVGPoint, 0, 7)
	for _, point := range points {
		for _, label := range lunarEclipseSVGPointLabels(point) {
			event := point
			event.Label = label
			event.Labels = []string{label}
			events = append(events, event)
		}
	}
	return events
}

func lunarEclipseSVGPointLabels(point lunarEclipseSVGPoint) []string {
	if len(point.Labels) > 0 {
		return point.Labels
	}
	if point.Label == "" {
		return nil
	}
	return []string{point.Label}
}

func writeLunarEclipseMoon(
	b *strings.Builder,
	point lunarEclipseSVGPoint,
	diagram basic.LunarEclipseDiagramResult,
	x, y, scale float64,
	greatest bool,
) {
	radius := diagram.MoonRadius * scale
	opacity := 0.64
	if greatest {
		opacity = 0.9
	}
	fmt.Fprintf(b, `<g class="event-moon">`)
	fmt.Fprintf(b, `<use href="#le-moon" x="%.3f" y="%.3f" width="%.3f" height="%.3f" opacity="%.2f"/>`,
		x-radius, y-radius, radius*2, radius*2, opacity)
	if lunarEclipseSVGDeepUmbra(point.LunarEclipseDiagramPoint, diagram) {
		tintOpacity := 0.46
		if greatest {
			tintOpacity = 0.58
		}
		fmt.Fprintf(b, `<circle cx="%.3f" cy="%.3f" r="%.3f" fill="#d66f1f" opacity="%.2f"/>`,
			x, y, radius, tintOpacity)
	}
	fmt.Fprintf(b, `<circle cx="%.3f" cy="%.3f" r="%.3f" fill="none" stroke="#b9b9b9" stroke-width="0.8" opacity="0.9"/>`, x, y, radius)
	b.WriteString(`</g>`)
}

func lunarEclipseSVGDeepUmbra(point basic.LunarEclipseDiagramPoint, diagram basic.LunarEclipseDiagramResult) bool {
	return math.Hypot(point.X, point.Y) < diagram.UmbraRadius+diagram.MoonRadius*0.75
}

func writeLunarEclipseEventLabel(
	b *strings.Builder,
	label string,
	x, y, cx, moonRadius float64,
	language string,
) {
	text := lunarEclipseSVGEventName(label, language)
	dx := moonRadius*0.72 + 6
	anchor := "start"
	if x < cx {
		dx = -dx
		anchor = "end"
	}
	dy := -moonRadius*0.35 - 4
	if label == "Greatest" {
		dy = moonRadius + 15
		anchor = "middle"
		dx = 0
	}
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#2554c7" font-family="Georgia, 'Times New Roman', serif" font-size="12" text-anchor="%s">%s</text>`,
		x+dx, y+dy, anchor, html.EscapeString(text))
}

func lunarEclipseSVGEventName(label, language string) string {
	if language == lunarEclipseSVGLanguageEnglish {
		switch label {
		case "Greatest":
			return "Greatest"
		default:
			return label
		}
	}
	switch label {
	case "P1":
		return "P1 半影始"
	case "U1":
		return "U1 初亏"
	case "U2":
		return "U2 食既"
	case "Greatest":
		return "食甚"
	case "U3":
		return "U3 生光"
	case "U4":
		return "U4 复圆"
	case "P4":
		return "P4 半影终"
	default:
		return label
	}
}

type lunarEclipseSVGContact struct {
	label    string
	name     string
	time     time.Time
	angle    float64
	hasAngle bool
}

func writeLunarEclipseContacts(
	b *strings.Builder,
	info LunarEclipseInfo,
	options LunarEclipseSVGOptions,
	x, y float64,
) {
	contacts := lunarEclipseSVGContacts(info, options.Language)
	if len(contacts) == 0 {
		return
	}
	title := lunarEclipseSVGContactsTitleText(options)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#111" font-family="Georgia, 'Times New Roman', serif" font-size="13" font-weight="700">%s (%s)</text>`,
		x, y, html.EscapeString(title), html.EscapeString(options.Location.String()))
	for index, contact := range contacts {
		line := fmt.Sprintf("%s %s  %s", contact.label, contact.name, contact.time.In(options.Location).Format("15:04:05"))
		if contact.hasAngle {
			if options.Language == lunarEclipseSVGLanguageEnglish {
				line = fmt.Sprintf("%s  PA %.1f°", line, contact.angle)
			} else {
				line = fmt.Sprintf("%s  方位 %.1f°", line, contact.angle)
			}
		}
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#222" font-family="Georgia, 'Times New Roman', serif" font-size="12">%s</text>`,
			x, y+float64(index+1)*18, html.EscapeString(line))
	}
}

// lunarEclipseSVGFooterLines 返回方向行在前的页脚各行：行数上限由图区下限与画布高度决定。
func lunarEclipseSVGFooterLines(direction string, info LunarEclipseInfo, options LunarEclipseSVGOptions, headerBottom float64) []string {
	const (
		fontSize     = 12.0
		diagramFloor = 160.0
		stageGap     = 10.0
	)
	height := float64(options.Height)
	lineHeight := svgchart.FooterLineHeight(fontSize)
	padding := localDiagramFooterBottomPadding
	maxLines := svgchart.TextLineLimit(fontSize, lineHeight, height-headerBottom-diagramFloor-stageGap-padding-lineHeight)
	if cap := int(height * 0.20 / lineHeight); maxLines > cap {
		maxLines = cap
	}
	if maxLines < 1 {
		maxLines = 1
	}
	note := timenote.Scale(options.TimeScale, info.Maximum, options.Location, options.Language)
	lines := []string{lunarEclipseSVGFooterNoteText(options)}
	if options.FooterNote != "" {
		lines = svgchart.WrapText(options.FooterNote, float64(options.Width)-80, fontSize)
	}
	lines = svgchart.FooterLinesWithScale(lines, note, float64(options.Width)-80, fontSize, maxLines)
	return append([]string{direction}, lines...)
}

func writeLunarEclipseFooter(b *strings.Builder, lines []string, baselines []float64) {
	for index, line := range lines {
		fill := "#555"
		if index == 0 {
			fill = "#333"
		}
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="%s" font-family="Georgia, 'Times New Roman', serif" font-size="12">%s</text>`,
			40.0, baselines[index], fill, html.EscapeString(line))
	}
}

func lunarEclipseSVGContacts(info LunarEclipseInfo, language string) []lunarEclipseSVGContact {
	angles := lunarEclipseContactAngleMap(info.ContactPoints)
	contacts := []lunarEclipseSVGContact{
		lunarEclipseSVGContactFor("P1", lunarEclipseSVGContactName("P1", language), info.PenumbralStart, angles),
	}
	if info.HasPartial {
		contacts = append(contacts, lunarEclipseSVGContactFor("U1", lunarEclipseSVGContactName("U1", language), info.PartialStart, angles))
	}
	if info.HasTotal {
		contacts = append(contacts, lunarEclipseSVGContactFor("U2", lunarEclipseSVGContactName("U2", language), info.TotalStart, angles))
	}
	contacts = append(contacts, lunarEclipseSVGContact{label: "GE", name: lunarEclipseSVGContactName("Greatest", language), time: info.Maximum})
	if info.HasTotal {
		contacts = append(contacts, lunarEclipseSVGContactFor("U3", lunarEclipseSVGContactName("U3", language), info.TotalEnd, angles))
	}
	if info.HasPartial {
		contacts = append(contacts, lunarEclipseSVGContactFor("U4", lunarEclipseSVGContactName("U4", language), info.PartialEnd, angles))
	}
	contacts = append(contacts, lunarEclipseSVGContactFor("P4", lunarEclipseSVGContactName("P4", language), info.PenumbralEnd, angles))
	return contacts
}

func lunarEclipseSVGContactFor(
	label, name string,
	time time.Time,
	angles map[string]float64,
) lunarEclipseSVGContact {
	angle, ok := angles[label]
	return lunarEclipseSVGContact{
		label:    label,
		name:     name,
		time:     time,
		angle:    angle,
		hasAngle: ok,
	}
}

func lunarEclipseContactAngleMap(points []LunarEclipseContactPoint) map[string]float64 {
	angles := make(map[string]float64, len(points))
	for _, point := range points {
		angles[point.Label] = point.ContactPositionAngle
	}
	return angles
}

func lunarEclipseSVGContactName(label, language string) string {
	if language == lunarEclipseSVGLanguageEnglish {
		switch label {
		case "P1":
			return "Penumbral begins"
		case "U1":
			return "Partial begins"
		case "U2":
			return "Total begins"
		case "Greatest":
			return "Greatest"
		case "U3":
			return "Total ends"
		case "U4":
			return "Partial ends"
		case "P4":
			return "Penumbral ends"
		default:
			return label
		}
	}
	switch label {
	case "P1":
		return "半影始"
	case "U1":
		return "初亏"
	case "U2":
		return "食既"
	case "Greatest":
		return "食甚"
	case "U3":
		return "生光"
	case "U4":
		return "复圆"
	case "P4":
		return "半影终"
	default:
		return label
	}
}

func durationToDays(duration time.Duration) float64 {
	if duration <= 0 {
		return 0
	}
	return duration.Hours() / 24
}
