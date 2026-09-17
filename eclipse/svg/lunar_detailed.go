package svg

import (
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro/basic"
	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/svgasset"
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/internal/svgmap"
)

// LunarEclipseDetailedSVGOptions 控制 详细版式的月食组合图：摘要、穿影示意图与世界可见性底图合成一页。
// LunarEclipseDetailedSVGOptions controls the detailed combined lunar-eclipse chart: summary, shadow-path
// diagram, and world visibility map on a single page.
type LunarEclipseDetailedSVGOptions struct {
	// Width 与 Height 是 SVG 画布尺寸；<=0 时使用 1000x1414 的 NASA 开本。版式按 Height 推导：
	// 高度不足时先压缩穿影示意图，再压缩底图；两者都低于下限（示意图 200、底图 180）、
	// 数据块行距或底图图例行放不下时返回 false。
	// Width and Height are the SVG canvas size; values <= 0 use the 1000x1414 NASA page. The page
	// stack is derived from Height: a short canvas shrinks the shadow-path diagram first and the
	// map second, and the renderer returns false when the diagram and map minima (200 and 180), the
	// data-block row spacing, or the map legend rows cannot be kept apart.
	Width  int
	Height int
	// Language 为 "en" 时使用英文，其他值使用中文。
	// Language uses English for "en" and Chinese otherwise.
	Language string
	// Location 控制显示时刻的时区；nil 使用 UTC+8。
	// Location controls the display timezone; nil uses UTC+8.
	Location *time.Location
	// Step 是月心路径采样步长；<=0 时使用 5 分钟。
	// Step is the Moon-center path sampling step; values <= 0 use five minutes.
	Step time.Duration
	// Title 与 FooterNote 为空时自动生成。
	// Empty Title and FooterNote use automatic text.
	Title      string
	FooterNote string
}

// LunarEclipseDetailedSVG 生成 详细版式的月食组合图，默认模型与 LunarEclipseOnDate 一致：Danjon 为主，极浅半影回退 Chauvenet。
// LunarEclipseDetailedSVG renders a detailed combined lunar-eclipse chart with the default model of
// LunarEclipseOnDate: Danjon, falling back to Chauvenet for ultra-shallow penumbral cases.
func LunarEclipseDetailedSVG(date time.Time, options LunarEclipseDetailedSVGOptions) (string, bool) {
	model, ok := lunarEclipseDefaultShadowModel(date)
	if !ok {
		return "", false
	}
	if model == eclipsecore.LunarEclipseShadowModelChauvenet {
		return LunarEclipseDetailedSVGChauvenet(date, options)
	}
	return LunarEclipseDetailedSVGDanjon(date, options)
}

// LunarEclipseDetailedSVGDanjon 生成 详细版式月食组合图，使用 Danjon 影半径模型。
// LunarEclipseDetailedSVGDanjon renders a detailed chart with the Danjon shadow model.
func LunarEclipseDetailedSVGDanjon(date time.Time, options LunarEclipseDetailedSVGOptions) (string, bool) {
	return lunarEclipseDetailedSVG(date, options,
		basic.LunarEclipseDiagramDanjon, basic.LunarEclipseShadowGeometryAt, eclipsecore.ClosestLunarEclipseDanjon,
		eclipsecore.LunarEclipseShadowModelDanjon)
}

// LunarEclipseDetailedSVGChauvenet 生成 详细版式月食组合图，使用 Chauvenet 影半径模型。
// LunarEclipseDetailedSVGChauvenet renders a detailed chart with the Chauvenet shadow model.
func LunarEclipseDetailedSVGChauvenet(date time.Time, options LunarEclipseDetailedSVGOptions) (string, bool) {
	return lunarEclipseDetailedSVG(date, options,
		basic.LunarEclipseDiagramChauvenet, basic.LunarEclipseShadowGeometryChauvenetAt, eclipsecore.ClosestLunarEclipseChauvenet,
		eclipsecore.LunarEclipseShadowModelChauvenet)
}

func normalizeLunarEclipseDetailedSVGOptions(options LunarEclipseDetailedSVGOptions) LunarEclipseDetailedSVGOptions {
	if options.Width <= 0 {
		options.Width = 1000
	}
	if options.Height <= 0 {
		options.Height = 1414
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

func lunarEclipseDetailedSVG(
	date time.Time,
	options LunarEclipseDetailedSVGOptions,
	calculator lunarEclipseSVGCalculator,
	geometry func(float64) basic.LunarEclipseShadowGeometry,
	finder lunarEclipseSVGFinder,
	model eclipsecore.LunarEclipseShadowModel,
) (string, bool) {
	options = normalizeLunarEclipseDetailedSVGOptions(options)
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
	if !validLunarEclipseDetailedSize(options.Width, options.Height, info, options) {
		return "", false
	}
	return renderLunarEclipseDetailedSVG(
		info, diagram, geometry(diagram.Eclipse.Maximum), timeToTTJDE(date), options, model,
	), true
}

// lunarEclipseDetailedView 汇总 详细版式各区块的位置。
type lunarEclipseDetailedView struct {
	width, height          float64
	margin                 float64
	metaY                  float64
	blockY                 float64
	diagramTop             float64
	diagramBottom          float64
	rowY                   float64
	rowHeight              float64
	mapFrame               svgmap.Frame
	legendY                float64
	legendRows             int
	footerY                float64
	columnWidth, columnGap float64
}

// 详细版式的竖向固定量：标题/摘要/地心块占用顶部，示意图与底图分掉剩余高度。
const (
	lunarEclipseDetailedDiagramTop     = 268.0
	lunarEclipseDetailedDiagramHeight  = 368.0
	lunarEclipseDetailedPanelHeight    = 152.0
	lunarEclipseDetailedPanelGap       = 28.0
	lunarEclipseDetailedMapGap         = 26.0
	lunarEclipseDetailedLegendGap      = 30.0
	lunarEclipseDetailedFooterReserve  = 52.0
	lunarEclipseDetailedMinDiagram     = 200.0
	lunarEclipseDetailedMinMapHeight   = 180.0
	lunarEclipseDetailedMinColumnWidth = 150.0
	// 图例与页脚的字号、页脚基线到画布底边的固定内缩，以及两者之间的最小间隙。
	lunarEclipseDetailedLegendFontSize  = 10.0
	lunarEclipseDetailedFooterFontSize  = 11.0
	lunarEclipseDetailedFooterInset     = 42.0
	lunarEclipseDetailedLegendFooterGap = 4.0
	// 底图可见性图例：字号、图标加文本占位宽、列间隙、行距与页边内缩。
	lunarEclipseDetailedMapLegendFontSize  = 10.0
	lunarEclipseDetailedMapLegendIconWidth = 24.0
	lunarEclipseDetailedMapLegendGap       = 12.0
	lunarEclipseDetailedMapLegendLineStep  = 22.0
	lunarEclipseDetailedPageInset          = 20.0
)

// lunarEclipseDetailedVerticalBudget 是扣掉顶部固定带、数据面板、图例与页脚后可分给示意图与底图的高度。
func lunarEclipseDetailedVerticalBudget(height float64) float64 {
	return height - (lunarEclipseDetailedDiagramTop + lunarEclipseDetailedPanelGap +
		lunarEclipseDetailedPanelHeight + lunarEclipseDetailedMapGap +
		lunarEclipseDetailedLegendGap + lunarEclipseDetailedFooterReserve)
}

// lunarEclipseDetailedMapHeight 给出该画布下底图高度：先保示意图下限，再把剩余高度给底图，
// 底图宽度还要受画布宽度与 2:1 长宽比限制。
func lunarEclipseDetailedMapHeight(width, height, margin float64) float64 {
	budget := lunarEclipseDetailedVerticalBudget(height)
	return math.Min((width-2*margin)/2, budget-lunarEclipseDetailedMinDiagram)
}

// validLunarEclipseDetailedSize 报告详细版式在该画布尺寸下能否容下全部内容：
// 除竖向预算与栏宽外，数据块行距、图例带与页脚也必须互不压叠。
func validLunarEclipseDetailedSize(width, height int, info LunarEclipseInfo, options LunarEclipseDetailedSVGOptions) bool {
	w, h := float64(width), float64(height)
	margin := math.Max(30, math.Min(52, w*0.044))
	budget := lunarEclipseDetailedVerticalBudget(h)
	mapHeight := lunarEclipseDetailedMapHeight(w, h, margin)
	if mapHeight < lunarEclipseDetailedMinMapHeight || budget-mapHeight < lunarEclipseDetailedMinDiagram {
		return false
	}
	columnGap := math.Max(16, w*0.02)
	if (w-2*margin-2*columnGap)/3 < lunarEclipseDetailedMinColumnWidth {
		return false
	}
	view := lunarEclipseDetailedViewFor(options)
	if solarEclipsePanelsOverlap(lunarEclipseDetailedPanelCells(view, info, options)) {
		return false
	}
	_, legendBelow := svgchart.EstimatedTextExtents(lunarEclipseDetailedMapLegendFontSize)
	_, footerBelow := svgchart.EstimatedTextExtents(lunarEclipseDetailedFooterFontSize)
	if view.footerY < view.legendY+legendBelow {
		return false
	}
	// 页脚要留在白色页框内，越出画布就没有可读位置了。
	return view.footerY+footerBelow <= h-20
}

// lunarEclipseDetailedPanelContent 给出三栏数据块的内容：左栏历时、右栏接触时刻。
func lunarEclipseDetailedPanelContent(
	info LunarEclipseInfo,
	options LunarEclipseDetailedSVGOptions,
) (string, []svgchart.PanelRow, string, []svgchart.PanelRow) {
	english := options.Language == lunarEclipseSVGLanguageEnglish
	duration := func(start, end time.Time) string {
		if start.IsZero() || end.IsZero() || !end.After(start) {
			return "-"
		}
		return formatSolarEclipseMapDuration(end.Sub(start))
	}
	durationsTitle := "月食历时"
	contactsTitle := "接触时刻"
	if english {
		durationsTitle = "Eclipse durations"
		contactsTitle = "Eclipse contacts"
	}
	durations := []svgchart.PanelRow{
		{Label: "半影食", Value: duration(info.PenumbralStart, info.PenumbralEnd)},
		{Label: "偏食", Value: duration(info.PartialStart, info.PartialEnd)},
		{Label: "全食", Value: duration(info.TotalStart, info.TotalEnd)},
	}
	if english {
		durations[0].Label, durations[1].Label, durations[2].Label = "Penumbral", "Partial", "Total"
	}
	contacts := []svgchart.PanelRow{
		{Label: "P1 半影食始", Value: info.PenumbralStart.In(options.Location).Format("15:04:05")},
		{Label: "U1 初亏", Value: info.PartialStart.In(options.Location).Format("15:04:05")},
		{Label: "U2 食既", Value: info.TotalStart.In(options.Location).Format("15:04:05")},
		{Label: "食甚", Value: info.Maximum.In(options.Location).Format("15:04:05")},
		{Label: "U3 生光", Value: info.TotalEnd.In(options.Location).Format("15:04:05")},
		{Label: "U4 复圆", Value: info.PartialEnd.In(options.Location).Format("15:04:05")},
		{Label: "P4 半影食终", Value: info.PenumbralEnd.In(options.Location).Format("15:04:05")},
	}
	if english {
		contacts = []svgchart.PanelRow{
			{Label: "P1 penumbral begins", Value: info.PenumbralStart.In(options.Location).Format("15:04:05")},
			{Label: "U1 partial begins", Value: info.PartialStart.In(options.Location).Format("15:04:05")},
			{Label: "U2 total begins", Value: info.TotalStart.In(options.Location).Format("15:04:05")},
			{Label: "Greatest", Value: info.Maximum.In(options.Location).Format("15:04:05")},
			{Label: "U3 total ends", Value: info.TotalEnd.In(options.Location).Format("15:04:05")},
			{Label: "U4 partial ends", Value: info.PartialEnd.In(options.Location).Format("15:04:05")},
			{Label: "P4 penumbral ends", Value: info.PenumbralEnd.In(options.Location).Format("15:04:05")},
		}
	}
	if !info.HasTotal {
		contacts[2].Value, contacts[4].Value = "-", "-"
	}
	if !info.HasPartial {
		contacts[1].Value, contacts[5].Value = "-", "-"
	}
	return durationsTitle, durations, contactsTitle, contacts
}

// lunarEclipseDetailedPanelCells 给出三栏数据块的矩形与内容：占位与绘制共用同一份几何。
func lunarEclipseDetailedPanelCells(
	view lunarEclipseDetailedView,
	info LunarEclipseInfo,
	options LunarEclipseDetailedSVGOptions,
) []solarEclipsePanelCell {
	durationsTitle, durations, contactsTitle, contacts := lunarEclipseDetailedPanelContent(info, options)
	right := view.margin + 2*(view.columnWidth+view.columnGap)
	return []solarEclipsePanelCell{
		{title: durationsTitle, rows: durations, box: svgchart.LabelBox{
			X: view.margin, Y: view.rowY, Width: view.columnWidth, Height: view.rowHeight,
		}},
		{title: contactsTitle, rows: contacts, box: svgchart.LabelBox{
			X: right, Y: view.rowY, Width: view.columnWidth, Height: view.rowHeight,
		}},
	}
}

func lunarEclipseDetailedViewFor(options LunarEclipseDetailedSVGOptions) lunarEclipseDetailedView {
	width, height := float64(options.Width), float64(options.Height)
	margin := math.Max(30, math.Min(52, width*0.044))
	view := lunarEclipseDetailedView{
		width: width, height: height, margin: margin,
		metaY: 88, blockY: 196,
		diagramTop: lunarEclipseDetailedDiagramTop,
		rowHeight:  lunarEclipseDetailedPanelHeight,
	}
	view.columnGap = math.Max(16, width*0.02)
	view.columnWidth = (width - 2*margin - 2*view.columnGap) / 3
	// 竖向是一条带栈：示意图与底图分掉固定带之外的高度；高度不足时两者一起缩，
	// 到下限的画布由 validLunarEclipseDetailedSize 提前拒绝，这里不再兜底。
	budget := lunarEclipseDetailedVerticalBudget(height)
	mapHeight := lunarEclipseDetailedMapHeight(width, height, margin)
	diagramHeight := math.Min(lunarEclipseDetailedDiagramHeight, budget-mapHeight)
	view.diagramBottom = view.diagramTop + diagramHeight
	view.rowY = view.diagramBottom + lunarEclipseDetailedPanelGap
	mapTop := view.rowY + view.rowHeight + lunarEclipseDetailedMapGap
	mapWidth := mapHeight * 2
	if limit := width - 2*margin; mapWidth > limit {
		mapWidth = limit
		mapHeight = mapWidth / 2
	}
	view.mapFrame = svgmap.Frame{
		X: (width - mapWidth) / 2, Y: mapTop, Width: mapWidth, Height: mapHeight,
		Projection: svgmap.ProjectionEquirectangular,
	}
	view.legendY = mapTop + mapHeight + lunarEclipseDetailedLegendGap
	view.legendRows = len(lunarEclipseMapLegendRows(view.mapFrame, width, options.Language))
	if view.legendRows < 1 {
		view.legendRows = 1
	}
	// 页脚按图例实测高度往下排：图例带再往下压也不能与页脚同高。
	_, legendBelow := svgchart.EstimatedTextExtents(lunarEclipseDetailedMapLegendFontSize)
	footerAbove, _ := svgchart.EstimatedTextExtents(lunarEclipseDetailedFooterFontSize)
	footerFloor := view.legendY + float64(view.legendRows-1)*lunarEclipseDetailedMapLegendLineStep +
		legendBelow + lunarEclipseDetailedLegendFooterGap + footerAbove
	view.footerY = math.Max(height-lunarEclipseDetailedFooterInset, footerFloor)
	return view
}

func renderLunarEclipseDetailedSVG(
	info LunarEclipseInfo,
	diagram basic.LunarEclipseDiagramResult,
	shadow basic.LunarEclipseShadowGeometry,
	tt float64,
	options LunarEclipseDetailedSVGOptions,
	model eclipsecore.LunarEclipseShadowModel,
) string {
	view := lunarEclipseDetailedViewFor(options)
	english := options.Language == lunarEclipseSVGLanguageEnglish
	title := options.Title
	if title == "" {
		date := info.Maximum.In(options.Location).Format("2006-01-02")
		if english {
			title = fmt.Sprintf("Total Lunar Eclipse of  %s", date)
		} else {
			title = fmt.Sprintf("%s %s", date, lunarEclipseSVGTypeName(info.Type, "zh"))
		}
	}

	extent := diagram.PenumbraRadius + diagram.MoonRadius + 0.72
	scale := math.Min(
		(view.width-2*view.margin)/(2*extent),
		(view.diagramBottom-view.diagramTop)/(2*extent),
	)
	if scale <= 0 || math.IsNaN(scale) || math.IsInf(scale, 0) {
		scale = 1
	}
	layout := lunarEclipseSVGLayout{
		width: view.width, height: view.height,
		cx:         view.width / 2,
		cy:         (view.diagramTop + view.diagramBottom) / 2,
		scale:      scale,
		diagramTop: view.diagramTop, diagramEnd: view.diagramBottom,
		diagramLeft: view.margin, diagramRight: view.width - view.margin,
	}
	// 不动的元素先占位：数据块、底图、图例与页脚；示意图标注只在示意图带内选位置。
	labels := &svgchart.LabelTable{}
	for _, cell := range lunarEclipseDetailedPanelCells(view, info, options) {
		labels.Reserve(cell.box.X, cell.box.Y, cell.box.Width, cell.box.Height)
	}
	labels.Reserve(view.mapFrame.X, view.mapFrame.Y, view.mapFrame.Width, view.mapFrame.Height)
	for _, row := range lunarEclipseMapLegendRows(view.mapFrame, view.width, options.Language) {
		for _, item := range row {
			labels.ReserveText(item.x, item.y, lunarEclipseDetailedMapLegendFontSize, item.text, "start")
		}
	}
	labels.ReserveText(view.margin, view.footerY, lunarEclipseDetailedFooterFontSize,
		lunarEclipseDetailedFooterText(options, model), "start")

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s">`,
		options.Width, options.Height, options.Width, options.Height, html.EscapeString(title))
	b.WriteString(`<defs>`)
	b.WriteString(svgasset.MoonFaceSymbol())
	writeLunarEclipseMapDefinitions(&b, view.mapFrame, info)
	b.WriteString(`</defs>`)
	b.WriteString(`<rect width="100%" height="100%" fill="#efefed"/>`)
	fmt.Fprintf(&b, `<rect x="22" y="18" width="%.3f" height="%.3f" fill="#ffffff" stroke="#c9c9c6" stroke-width="1.2"/>`,
		view.width-44, view.height-36)
	titleText := title
	if options.Title != "" {
		titleText = svgchart.EllipsizeText(title, view.width-52, 25)
	}
	fmt.Fprintf(&b, `<text x="%.3f" y="52" fill="#111111" font-family="Georgia, 'Times New Roman', serif" font-size="25" font-weight="700" text-anchor="middle">%s</text>`,
		view.width/2, html.EscapeString(titleText))

	writeLunarEclipseDetailedSummary(&b, labels, info, shadow, view, options)
	writeLunarEclipseDetailedGeocentricBlocks(&b, labels, tt, view, options)
	writeLunarEclipseDiagram(&b, layout, diagram, options.Language, labels)
	writeLunarEclipseDetailedPanels(&b, info, shadow, scale, view, options)
	writeLunarEclipseMapRegions(&b, view.mapFrame, info)
	view.mapFrame.WriteFrame(&b)
	writeLunarEclipseMapLegend(&b, view.mapFrame, view.width, options.Language)
	writeLunarEclipseDetailedFooter(&b, view, options, model)
	b.WriteString(`</svg>`)
	return b.String()
}

// writeLunarEclipseDetailedSummary 写标题下居中的四行摘要。
func writeLunarEclipseDetailedSummary(
	b *strings.Builder,
	labels *svgchart.LabelTable,
	info LunarEclipseInfo,
	shadow basic.LunarEclipseShadowGeometry,
	view lunarEclipseDetailedView,
	options LunarEclipseDetailedSVGOptions,
) {
	english := options.Language == lunarEclipseSVGLanguageEnglish
	zone, _ := info.Maximum.In(options.Location).Zone()
	lines := make([]string, 0, 4)
	if english {
		lines = append(lines, fmt.Sprintf("Greatest Eclipse = %s (%s) | Umbral magnitude = %.4f",
			info.Maximum.In(options.Location).Format("15:04:05"), zone, info.UmbralMagnitude))
		lines = append(lines, fmt.Sprintf("Penumbral magnitude = %.4f | Gamma = %.4f",
			info.PenumbralMagnitude, shadow.Gamma))
		lines = append(lines, fmt.Sprintf("P. Radius = %.4f° | U. Radius = %.4f° | Axis = %.4f°",
			shadow.PenumbralRadiusDegrees, shadow.UmbralRadiusDegrees, shadow.AxisDegrees))
		lines = append(lines, fmt.Sprintf("Saros series = %d | member %d of %d", info.Saros.Series, info.Saros.Member, info.Saros.Count))
	} else {
		lines = append(lines, fmt.Sprintf("食甚 = %s（%s）| 本影食分 = %.4f",
			info.Maximum.In(options.Location).Format("15:04:05"), zone, info.UmbralMagnitude))
		lines = append(lines, fmt.Sprintf("半影食分 = %.4f | 伽马 = %.4f", info.PenumbralMagnitude, shadow.Gamma))
		lines = append(lines, fmt.Sprintf("半影半径 = %.4f° | 本影半径 = %.4f° | 影轴角距 = %.4f°",
			shadow.PenumbralRadiusDegrees, shadow.UmbralRadiusDegrees, shadow.AxisDegrees))
		lines = append(lines, fmt.Sprintf("沙罗序列 = %d | 第 %d / %d 个成员", info.Saros.Series, info.Saros.Member, info.Saros.Count))
	}
	if !info.HasSaros {
		lines[len(lines)-1] = ""
	}
	for index, line := range lines {
		if line == "" {
			continue
		}
		labels.ReserveText(view.width/2, view.metaY+float64(index)*22, 12.5, line, "middle")
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#293235" font-family="Arial, sans-serif" font-size="12.5" text-anchor="middle">%s</text>`,
			view.width/2, view.metaY+float64(index)*22, html.EscapeString(line))
	}
	zoneNote := "图中时刻为 UT"
	if offset := zoneOffsetSeconds(info.Maximum.In(options.Location)); offset != 0 {
		zoneNote = fmt.Sprintf("图中时刻为 %s（UT%s）", zone, formatZoneOffset(offset))
	}
	if english {
		zoneNote = "All times are UT"
		if offset := zoneOffsetSeconds(info.Maximum.In(options.Location)); offset != 0 {
			zoneNote = fmt.Sprintf("All times are %s (UT%s)", zone, formatZoneOffset(offset))
		}
	}
	labels.ReserveText(view.width/2, view.metaY+4*22, 11, zoneNote, "middle")
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#7b8794" font-family="Arial, sans-serif" font-size="11" text-anchor="middle">%s</text>`,
		view.width/2, view.metaY+4*22, html.EscapeString(zoneNote))
}

func zoneOffsetSeconds(value time.Time) int {
	_, offset := value.Zone()
	return offset
}

func formatZoneOffset(offsetSeconds int) string {
	sign := "+"
	if offsetSeconds < 0 {
		sign = "-"
		offsetSeconds = -offsetSeconds
	}
	return fmt.Sprintf("%s%02d:%02d", sign, offsetSeconds/3600, (offsetSeconds%3600)/60)
}

// writeLunarEclipseDetailedGeocentricBlocks 在球面两侧画日月地心坐标块。
func writeLunarEclipseDetailedGeocentricBlocks(
	b *strings.Builder,
	labels *svgchart.LabelTable,
	tt float64,
	view lunarEclipseDetailedView,
	options LunarEclipseDetailedSVGOptions,
) {
	english := options.Language == lunarEclipseSVGLanguageEnglish
	titles := [2]string{"食甚时的太阳（地心坐标）", "食甚时的月亮（地心坐标）"}
	if english {
		titles = [2]string{"Sun at greatest eclipse", "Moon at greatest eclipse"}
	}
	sunRa, sunDec := basic.SunApparentRaDec(tt)
	moonRa, moonDec := basic.HMoonTrueRaDec(tt)
	sunSd := basic.SunSemidiameter(tt)
	moonSd := basic.MoonSemidiameter(tt)
	blocks := [2][]svgchart.PanelRow{
		{
			{Label: "赤经 R.A.", Value: formatSolarEclipseRA(sunRa)},
			{Label: "赤纬 Dec.", Value: formatSolarEclipseDec(sunDec)},
			{Label: "视半径 S.D.", Value: formatSolarEclipseArcsec(sunSd)},
			{Label: "地平视差 H.P.", Value: formatSolarEclipseArcsec(horizontalParallaxArcsecRatio(sunSd, 6378.137/696000.0))},
		},
		{
			{Label: "赤经 R.A.", Value: formatSolarEclipseRA(moonRa)},
			{Label: "赤纬 Dec.", Value: formatSolarEclipseDec(moonDec)},
			{Label: "视半径 S.D.", Value: formatSolarEclipseArcsec(moonSd)},
			{Label: "地平视差 H.P.", Value: formatSolarEclipseArcsec(horizontalParallaxArcsecRatio(moonSd, 6378.137/1737.4))},
		},
	}
	if english {
		for index := range blocks {
			blocks[index][3].Label = "H.P."
		}
	}
	blockWidth := math.Min(212, (view.width-2*view.margin-40)/2)
	for index := 0; index < 2; index++ {
		x := view.margin
		if index == 1 {
			x = view.width - view.margin - blockWidth
		}
		fmt.Fprintf(b, `<g class="lunar-geocentric-block">`)
		labels.ReserveText(x, view.blockY+11, 12, titles[index], "start")
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#111827" font-family="Georgia, 'Times New Roman', serif" font-size="12" font-weight="700">%s</text>`,
			x, view.blockY+11, html.EscapeString(titles[index]))
		for rowIndex, row := range blocks[index] {
			rowY := view.blockY + 30 + float64(rowIndex)*17
			labels.ReserveText(x, rowY, 10.5, row.Label, "start")
			labels.ReserveText(x+blockWidth, rowY, 10.5, row.Value, "end")
			fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#475569" font-family="Arial, sans-serif" font-size="10.5">%s</text>`,
				x, rowY, html.EscapeString(row.Label))
			fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#0f172a" font-family="Arial, sans-serif" font-size="10.5" font-weight="700" text-anchor="end">%s</text>`,
				x+blockWidth, rowY, html.EscapeString(row.Value))
		}
		b.WriteString(`</g>`)
	}
}

// horizontalParallaxArcsecRatio 由视半径换算地平视差：sin(HP) = sin(SD) × R⊕ / R天体。
func horizontalParallaxArcsecRatio(semidiameterArcsec, radiusRatio float64) float64 {
	sine := math.Sin(semidiameterArcsec / 3600 * math.Pi / 180)
	return math.Asin(math.Max(-1, math.Min(1, sine*radiusRatio))) * 180 / math.Pi * 3600
}

// writeLunarEclipseDetailedPanels 写三栏：历时、弧分比例尺、接触时刻。
func writeLunarEclipseDetailedPanels(
	b *strings.Builder,
	info LunarEclipseInfo,
	shadow basic.LunarEclipseShadowGeometry,
	diagramScale float64,
	view lunarEclipseDetailedView,
	options LunarEclipseDetailedSVGOptions,
) {
	for _, cell := range lunarEclipseDetailedPanelCells(view, info, options) {
		writeLunarEclipseDetailedPanel(b, cell.box.X, cell.box.Y, cell.box.Width, cell.box.Height, cell.title, cell.rows)
	}
	writeLunarEclipseArcMinuteScaleBar(b, view, shadow, diagramScale)
}

func writeLunarEclipseDetailedPanel(
	b *strings.Builder,
	x, y, width, height float64,
	title string,
	rows []svgchart.PanelRow,
) {
	svgchart.WritePanelBox(b, "lunar-detailed-panel", x, y, width, height, title, rows)
}

// writeLunarEclipseArcMinuteScaleBar 画穿影示意图的弧分比例尺。
// 示意图以"月球半径"为单位，一个月球半径对应真实角径 asin(1/月距)，据此换算弧分。
func writeLunarEclipseArcMinuteScaleBar(
	b *strings.Builder,
	view lunarEclipseDetailedView,
	shadow basic.LunarEclipseShadowGeometry,
	diagramScale float64,
) {
	if shadow.MoonDistanceEarthRadii <= 0 || diagramScale <= 0 {
		return
	}
	// 示意图以"月球半径"为单位，一个月球半径对应的真实角径是 asin(R月 / 月距)，不是地球视差。
	moonSemidiameterArcmin := math.Asin((lunarMoonRadiusKM/lunarEarthRadiusKM)/shadow.MoonDistanceEarthRadii) * 180 / math.Pi * 60
	if moonSemidiameterArcmin <= 0 {
		return
	}
	pixelsPerArcmin := diagramScale / moonSemidiameterArcmin
	target := 150.0
	step := math.Pow(10, math.Floor(math.Log10(target/pixelsPerArcmin)))
	for _, multiplier := range []float64{1, 2, 5, 10} {
		if step*multiplier >= target/pixelsPerArcmin {
			step *= multiplier
			break
		}
	}
	barWidth := step * pixelsPerArcmin
	centreX := view.margin + view.columnWidth + view.columnGap + view.columnWidth/2
	y := view.rowY + view.rowHeight/2
	fmt.Fprintf(b, `<g class="lunar-scale-bar"><line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#334155" stroke-width="1"/>`,
		centreX-barWidth/2, y, centreX+barWidth/2, y)
	for _, tick := range []float64{0, 0.25, 0.5, 0.75, 1} {
		x := centreX - barWidth/2 + tick*barWidth
		length := 5.0
		if tick == 0 || tick == 1 {
			length = 9
		}
		fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#334155" stroke-width="1"/>`,
			x, y-length/2, x, y+length/2)
	}
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#334155" font-family="Arial, sans-serif" font-size="10" text-anchor="middle">0</text>`,
		centreX-barWidth/2, y-10)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#334155" font-family="Arial, sans-serif" font-size="10" text-anchor="middle">%.0f</text>`,
		centreX+barWidth/2, y-10, step)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#7b8794" font-family="Arial, sans-serif" font-size="10" text-anchor="middle">角分</text>`,
		centreX, y+22)
	b.WriteString(`</g>`)
}

// lunarEclipseDetailedFooterText 是页脚说明：默认模型会在极浅半影上回退 Chauvenet，必须写出实际用的那套。
func lunarEclipseDetailedFooterText(options LunarEclipseDetailedSVGOptions, model eclipsecore.LunarEclipseShadowModel) string {
	if options.FooterNote != "" {
		return options.FooterNote
	}
	if options.Language == lunarEclipseSVGLanguageEnglish {
		return "Shadow-path diagram uses the " + lunarEclipseShadowModelName(model) + " shadow model; the lower map marks where the whole eclipse is visible, where the Moon rises or sets eclipsed, and where it is not visible. Natural Earth 1:50m physical land, no administrative boundaries."
	}
	return "穿影示意图使用 " + lunarEclipseShadowModelName(model) + " 影半径模型；下方底图区分全程可见、带食月出、带食月落与不可见四类区域。Natural Earth 1:50m 物理陆地底图，不含行政边界。"
}

func writeLunarEclipseDetailedFooter(
	b *strings.Builder,
	view lunarEclipseDetailedView,
	options LunarEclipseDetailedSVGOptions,
	model eclipsecore.LunarEclipseShadowModel,
) {
	// 页脚按页宽折行、按画布底边截断，首行位置不变；默认说明放得下时折行结果仍是同一行。
	maxWidth := view.width - 2*view.margin
	lines := svgchart.TruncateTextLines(
		svgchart.WrapText(lunarEclipseDetailedFooterText(options, model), maxWidth, lunarEclipseDetailedFooterFontSize),
		maxWidth, lunarEclipseDetailedFooterFontSize,
		svgchart.BaselineLineLimit(lunarEclipseDetailedFooterFontSize, 15, view.footerY, view.height-4))
	for index, line := range lines {
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#596164" font-family="Georgia, 'Times New Roman', serif" font-size="%.0f">%s</text>`,
			view.margin, view.footerY+float64(index)*15, lunarEclipseDetailedFooterFontSize, html.EscapeString(line))
	}
}

const (
	// lunarMoonRadiusKM 与 lunarEarthRadiusKM 只用于把示意图的"月球半径"刻度换算成角分比例尺。
	lunarMoonRadiusKM  = 1737.4
	lunarEarthRadiusKM = 6378.137
)
