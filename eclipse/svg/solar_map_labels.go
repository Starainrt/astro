package svg

import (
	"math"
	"time"

	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/internal/svgmap"
)

// 数据块的表头高度与行字号：取值与 internal/svgchart 的排版常量相同，这里只用它们反推行距下限。
const (
	solarEclipsePanelHeaderHeight = 30.0
	solarEclipsePanelRowFontSize  = 10.5
)

// solarEclipsePanelRowSpacingMin 是数据块行距下限：低于它相邻两行文字必然压叠。
func solarEclipsePanelRowSpacingMin() float64 {
	above, below := svgchart.EstimatedTextExtents(solarEclipsePanelRowFontSize)
	return above + below
}

// solarEclipseMinPanelHeight 是容下 count 行文字的数据块高度下限；0.5 px 是等分行高时浮点取整的余量。
func solarEclipseMinPanelHeight(count int) float64 {
	return solarEclipsePanelHeaderHeight + float64(count)*solarEclipsePanelRowSpacingMin() + 0.5
}

// solarEclipsePanelCell 是一个数据块的矩形与内容：占位与绘制共用同一份几何。
type solarEclipsePanelCell struct {
	title string
	rows  []svgchart.PanelRow
	box   svgchart.LabelBox
}

// solarEclipsePanelCells 给出该版式下全部数据块矩形。
func solarEclipsePanelCells(layout solarEclipseMapLayout, blocks solarEclipsePanelBlocks) []solarEclipsePanelCell {
	if layout.landscape {
		order := solarEclipseLandscapeBlockOrder(blocks)
		cells := make([]solarEclipsePanelCell, 0, len(order)+1)
		for index, block := range order {
			if len(block.rows) == 0 {
				continue
			}
			box, ok := solarEclipseLandscapeCellBox(layout, index)
			if !ok {
				continue
			}
			cells = append(cells, solarEclipsePanelCell{title: block.title, rows: block.rows, box: box})
		}
		if len(blocks.libration.rows) > 0 && layout.stripHeight > 0 {
			cells = append(cells, solarEclipsePanelCell{
				title: blocks.libration.title, rows: blocks.libration.rows,
				box: svgchart.LabelBox{
					X: layout.margin, Y: layout.bottomY,
					Width: solarEclipseLandscapeCellWidth, Height: layout.stripHeight,
				},
			})
		}
		return cells
	}
	if !layout.nasa {
		return nil
	}
	columnWidth := solarEclipsePortraitColumnWidth(layout)
	cells := make([]solarEclipsePanelCell, 0, 5)
	for column, block := range []solarEclipsePanelBlock{blocks.penumbra, blocks.circumstances, blocks.umbra} {
		if len(block.rows) == 0 {
			continue
		}
		cells = append(cells, solarEclipsePanelCell{title: block.title, rows: block.rows, box: svgchart.LabelBox{
			X: solarEclipsePortraitColumnX(layout, column), Y: layout.panelY,
			Width: columnWidth, Height: layout.panelHeight,
		}})
	}
	second := [2]struct {
		column int
		block  solarEclipsePanelBlock
	}{{0, blocks.ephemeris}, {2, blocks.libration}}
	for _, entry := range second {
		if len(entry.block.rows) == 0 {
			continue
		}
		cells = append(cells, solarEclipsePanelCell{title: entry.block.title, rows: entry.block.rows, box: svgchart.LabelBox{
			X: solarEclipsePortraitColumnX(layout, entry.column), Y: layout.secondPanelY,
			Width: columnWidth, Height: layout.panelHeight,
		}})
	}
	return cells
}

// solarEclipsePanelsOverlap 报告任一数据块内文字是否压叠；为真时调用方必须拒绝该画布。
func solarEclipsePanelsOverlap(cells []solarEclipsePanelCell) bool {
	for _, cell := range cells {
		if len(cell.rows) == 0 || cell.box.Width <= 0 || cell.box.Height <= 0 {
			continue
		}
		if svgchart.PanelBoxOverlaps(cell.box.X, cell.box.Y, cell.box.Width, cell.box.Height, cell.title, cell.rows) {
			return true
		}
	}
	return false
}

func solarEclipsePortraitColumnWidth(layout solarEclipseMapLayout) float64 {
	return (layout.panelWidth - 2*layout.panelGap) / 3
}

func solarEclipsePortraitColumnX(layout solarEclipseMapLayout, column int) float64 {
	return layout.panelX + float64(column)*(solarEclipsePortraitColumnWidth(layout)+layout.panelGap)
}

func solarEclipseLandscapeCellBox(layout solarEclipseMapLayout, index int) (svgchart.LabelBox, bool) {
	row := index / solarEclipseLandscapeColumns
	column := index % solarEclipseLandscapeColumns
	if row < 0 || row >= len(layout.gridRowHeights) || layout.gridRowHeights[row] <= 0 {
		return svgchart.LabelBox{}, false
	}
	return svgchart.LabelBox{
		X: layout.gridX + float64(column)*(layout.gridCellWidth+layout.gridGapX),
		Y: layout.gridRowY[row], Width: layout.gridCellWidth, Height: layout.gridRowHeights[row],
	}, true
}

// solarEclipseMapLegendItem 是图例的一条：标签、图形种类、颜色与虚线。
type solarEclipseMapLegendItem struct {
	label string
	kind  string
	color string
	dash  string
}

const (
	solarEclipseLegendFontSize   = 10.0
	solarEclipseLegendLineStep   = 22.0
	solarEclipseLegendIconWidth  = 25.0
	solarEclipseLegendColumnGap  = 12.0
	solarEclipseLegendMaxColumns = 4
)

// solarEclipseMapLegendItems 汇总该事件要画的图例条目，顺序即优先级。
func solarEclipseMapLegendItems(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	hasCentral bool,
	hasCentralBandFootprints bool,
	options SolarEclipseMapSVGOptions,
) []solarEclipseMapLegendItem {
	info := partial.Eclipse
	partialLabel := "偏食可见区"
	centerLineLabel := "中心线"
	if options.Language == "en" {
		partialLabel = "Partial-eclipse visibility"
		centerLineLabel = "Center line"
	}
	items := []solarEclipseMapLegendItem{{label: partialLabel, kind: "fill", color: "#f7dd93"}}
	if hasCentral || hasCentralBandFootprints {
		items = append(items, solarEclipseMapLegendItem{
			label: solarEclipseCentralPathLabel(info.Type, options.Language), kind: "fill",
			color: solarEclipseCentralPathColor(info.Type),
		})
	}
	if hasCentral {
		items = append(items, solarEclipseMapLegendItem{label: centerLineLabel, kind: "line", color: "#263f58", dash: "5 3"})
	}
	if options.PenumbralOutlineStep > 0 {
		items = append(items, solarEclipseMapLegendItem{
			label: solarEclipseOutlineLegendLabel("penumbra", info.Type, options.PenumbralOutlineStep, options.Language),
			kind:  "line", color: "#b07a18", dash: "3 3",
		})
	}
	riseSetLabel := "初亏/食甚/复圆日升日落线"
	if options.Language == "en" {
		riseSetLabel = "Rise/set phase lines"
	}
	if len(partial.RiseSetCurves) > 0 {
		items = append(items, solarEclipseMapLegendItem{label: riseSetLabel, kind: "line", color: "#d97706", dash: "6 4"})
	}
	if len(partial.MagnitudeContours) > 0 {
		magnitudeLabel := "地方食分 0.2–0.8"
		if options.Language == "en" {
			magnitudeLabel = "Local magnitude 0.2-0.8"
		}
		items = append(items, solarEclipseMapLegendItem{label: magnitudeLabel, kind: "line", color: "#7c3aed", dash: "4 3"})
	}
	if len(partial.GreatestTimeContours) > 0 {
		isochroneLabel := "食甚时刻等时线"
		if options.Language == "en" {
			isochroneLabel = "Greatest-eclipse isochrones"
		}
		items = append(items, solarEclipseMapLegendItem{label: isochroneLabel, kind: "line", color: "#1f6fb2"})
	}
	if info.Centrality != eclipsecore.SolarEclipseNonCentral &&
		len(partial.CentralShadowFootprints) > 0 && options.CentralShadowStep > 0 {
		items = append(items, solarEclipseMapLegendItem{
			label: solarEclipseOutlineLegendLabel("central", info.Type, options.CentralShadowStep, options.Language),
			kind:  "line", color: "#7b5a42",
		})
	}
	contactLabel := "P/U 影锥接触"
	hasUContacts := !partial.U1.Time.IsZero() || !partial.U2.Time.IsZero() ||
		!partial.U3.Time.IsZero() || !partial.U4.Time.IsZero()
	if !hasUContacts {
		contactLabel = "P 半影接触"
	}
	if options.Language == "en" {
		contactLabel = "P/U shadow contacts"
		if !hasUContacts {
			contactLabel = "P penumbral contacts"
		}
	}
	return append(items, solarEclipseMapLegendItem{label: contactLabel, kind: "contact", color: "#a52d70"})
}

// solarEclipseMapLegendRows 按可用宽度把条目分行：每行不超过四栏，栏位由实测文本宽度累加，长标签不会挤进下一栏。
func solarEclipseMapLegendRows(items []solarEclipseMapLegendItem, available float64) [][]solarEclipseMapLegendItem {
	rows := make([][]solarEclipseMapLegendItem, 0, 2)
	current := make([]solarEclipseMapLegendItem, 0, solarEclipseLegendMaxColumns)
	used := 0.0
	flush := func() {
		if len(current) == 0 {
			return
		}
		rows = append(rows, current)
		current = make([]solarEclipseMapLegendItem, 0, solarEclipseLegendMaxColumns)
		used = 0
	}
	for _, item := range items {
		width := solarEclipseLegendIconWidth + svgchart.EstimatedTextWidth(item.label, solarEclipseLegendFontSize)
		next := width
		if len(current) > 0 {
			next = used + solarEclipseLegendColumnGap + width
			if len(current) >= solarEclipseLegendMaxColumns || next > available {
				flush()
				next = width
			}
		}
		current = append(current, item)
		used = next
	}
	flush()
	return rows
}

// solarEclipseMapLegendBlockHeight 由图例行数给出图例带高度：首行基线上方还要留出字高。
func solarEclipseMapLegendBlockHeight(rows int) float64 {
	if rows < 1 {
		rows = 1
	}
	above, below := svgchart.EstimatedTextExtents(solarEclipseLegendFontSize)
	return float64(rows-1)*solarEclipseLegendLineStep + above + below + 1
}

// solarEclipseMapLegendAvailableWidth 只由画布宽度与排布给出图例可用宽度，与竖向堆叠无关。
func solarEclipseMapLegendAvailableWidth(options SolarEclipseMapSVGOptions) float64 {
	width, height := float64(options.Width), float64(options.Height)
	if width >= height {
		panelGap := math.Max(12, width*0.011)
		return solarEclipseLandscapeColumns*solarEclipseLandscapeCellWidth +
			float64(solarEclipseLandscapeColumns-1)*panelGap - 26
	}
	margin := math.Max(30, math.Min(52, width*0.044))
	return width - 2*margin - 26
}

// solarEclipseMapPointLabel 是一个点标注：图形锚点、文本、字号与落位结果。
type solarEclipseMapPointLabel struct {
	x, y     float64
	text     string
	fontSize float64
	placed   svgchart.LabelPlacement
	ok       bool
}

func solarEclipseMapPlaceLabel(
	labels *svgchart.LabelTable,
	label solarEclipseMapPointLabel,
	candidates []svgchart.LabelPlacement,
	frame svgmap.Frame,
) solarEclipseMapPointLabel {
	label.placed, label.ok = labels.Place(label.text, label.fontSize,
		labelCandidatesInsideBox(candidates, label.text, label.fontSize, solarEclipseFrameBox(frame)))
	return label
}

// labelCandidatesInsideBox 丢掉会让文本跑出给定区域的候选位次。
func labelCandidatesInsideBox(
	candidates []svgchart.LabelPlacement,
	text string,
	fontSize float64,
	area svgchart.LabelBox,
) []svgchart.LabelPlacement {
	kept := make([]svgchart.LabelPlacement, 0, len(candidates))
	for _, candidate := range candidates {
		box := svgchart.LabelTextBox(candidate.X, candidate.Y, fontSize, text, candidate.Anchor)
		if box.X < area.X || box.X+box.Width > area.X+area.Width {
			continue
		}
		if box.Y < area.Y || box.Y+box.Height > area.Y+area.Height {
			continue
		}
		kept = append(kept, candidate)
	}
	return kept
}

// solarEclipseFrameBox 把图框写成占位矩形。
func solarEclipseFrameBox(frame svgmap.Frame) svgchart.LabelBox {
	return svgchart.LabelBox{X: frame.X, Y: frame.Y, Width: frame.Width, Height: frame.Height}
}

// solarEclipseMapPointLabels 汇总优先放置的点标注：食甚、影锥接触、太阳直射点与中心线时刻。
type solarEclipseMapPointLabels struct {
	greatest solarEclipseMapPointLabel
	contacts []solarEclipseContactLabel
	subsolar solarEclipseMapPointLabel
	times    []solarEclipseMapPointLabel
}

// solarEclipseContactMarker 是一个影锥接触标记：点名、位置、颜色与原位标注偏移。
type solarEclipseContactMarker struct {
	name       string
	point      eclipsecore.SolarEclipsePathPoint
	color      string
	labelDX    float64
	labelDY    float64
	textAnchor string
}

// solarEclipseContactLabel 是接触标记的落位结果：图形锚点与标注文本。
type solarEclipseContactLabel struct {
	marker solarEclipseContactMarker
	x, y   float64
	text   string
	placed svgchart.LabelPlacement
	ok     bool
}

func solarEclipseContactMarkerSpecs(info eclipsecore.SolarEclipsePartialFootprintsInfo) []solarEclipseContactMarker {
	return []solarEclipseContactMarker{
		{name: "P1", point: info.P1, color: "#a52d70", labelDX: -7, labelDY: 14, textAnchor: "end"},
		{name: "P2", point: info.P2, color: "#a52d70", labelDX: 7, labelDY: -7, textAnchor: "start"},
		{name: "P3", point: info.P3, color: "#a52d70", labelDX: -7, labelDY: -7, textAnchor: "end"},
		{name: "P4", point: info.P4, color: "#a52d70", labelDX: 7, labelDY: 14, textAnchor: "start"},
		{name: "U1", point: info.U1, color: "#a12c25", labelDX: -7, labelDY: -7, textAnchor: "end"},
		{name: "U2", point: info.U2, color: "#a12c25", labelDX: 7, labelDY: 14, textAnchor: "start"},
		{name: "U3", point: info.U3, color: "#a12c25", labelDX: -7, labelDY: 14, textAnchor: "end"},
		{name: "U4", point: info.U4, color: "#a12c25", labelDX: 7, labelDY: -7, textAnchor: "start"},
	}
}

// solarEclipseMapPlacePointLabels 按重要性放置点标注：食甚最先占位，其余标注再让开。
func solarEclipseMapPlacePointLabels(
	civilPartial eclipsecore.SolarEclipsePartialFootprintsInfo,
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	central eclipsecore.SolarEclipsePath,
	hasCentral bool,
	frame svgmap.Frame,
	options SolarEclipseMapSVGOptions,
	labels *svgchart.LabelTable,
) solarEclipseMapPointLabels {
	result := solarEclipseMapPointLabels{}
	if x, y, ok := frame.Project(partial.Eclipse.GreatestLongitude, partial.Eclipse.GreatestLatitude); ok {
		text := "食甚"
		if options.Language == "en" {
			text = "Greatest"
		}
		result.greatest = solarEclipseMapPlaceLabel(labels, solarEclipseMapPointLabel{x: x, y: y, text: text, fontSize: 11},
			[]svgchart.LabelPlacement{
				{X: x, Y: y - 10, Anchor: "middle"},
				{X: x, Y: y + 19, Anchor: "middle"},
				{X: x + 9, Y: y + 4, Anchor: "start"},
				{X: x - 9, Y: y + 4, Anchor: "end"},
				{X: x + 9, Y: y - 8, Anchor: "start"},
				{X: x - 9, Y: y - 8, Anchor: "end"},
				{X: x, Y: y - 24, Anchor: "middle"},
				{X: x, Y: y + 33, Anchor: "middle"},
			}, frame)
	}
	result.contacts = solarEclipseMapPlaceContactLabels(partial, frame, labels)
	result.subsolar = solarEclipseMapPlaceSubsolarLabel(civilPartial.Eclipse, frame, options.Language, labels)
	if hasCentral {
		result.times = solarEclipseMapPlaceTimeLabels(central, frame, options, labels)
	}
	return result
}

func solarEclipseMapPlaceContactLabels(
	info eclipsecore.SolarEclipsePartialFootprintsInfo,
	frame svgmap.Frame,
	labels *svgchart.LabelTable,
) []solarEclipseContactLabel {
	specs := solarEclipseContactMarkerSpecs(info)
	result := make([]solarEclipseContactLabel, 0, len(specs))
	for _, marker := range specs {
		if marker.point.Time.IsZero() {
			continue
		}
		x, y, visible := frame.Project(marker.point.Longitude, marker.point.Latitude)
		if !visible {
			continue
		}
		labelX, anchor := x+marker.labelDX, marker.textAnchor
		// 贴到图框边上的点名改到另一侧，避免被裁掉。
		if labelX < frame.X+18 {
			labelX, anchor = x+7, "start"
		} else if labelX > frame.X+frame.Width-18 {
			labelX, anchor = x-7, "end"
		}
		candidates := make([]svgchart.LabelPlacement, 0, 22)
		candidates = append(candidates, svgchart.LabelPlacement{X: labelX, Y: y + marker.labelDY, Anchor: anchor})
		candidates = append(candidates, svgchart.LabelCandidates(x+7, y+14, "start", 9)...)
		candidates = append(candidates, svgchart.LabelCandidates(x-7, y-7, "end", 9)...)
		candidates = append(candidates, svgchart.LabelCandidates(x+7, y-7, "start", 16)...)
		candidates = append(candidates, svgchart.LabelCandidates(x-7, y+14, "end", 16)...)
		label := solarEclipseMapPlaceLabel(labels, solarEclipseMapPointLabel{x: x, y: y, text: marker.name, fontSize: 8}, candidates, frame)
		result = append(result, solarEclipseContactLabel{marker: marker, x: x, y: y, text: marker.name, placed: label.placed, ok: label.ok})
	}
	return result
}

func solarEclipseMapPlaceSubsolarLabel(
	info eclipsecore.SolarEclipseInfo,
	frame svgmap.Frame,
	language string,
	labels *svgchart.LabelTable,
) solarEclipseMapPointLabel {
	point := solarEclipseSubsolarPoint(info.GreatestEclipse)
	x, y, visible := frame.Project(point.Longitude, point.Latitude)
	if !visible {
		return solarEclipseMapPointLabel{}
	}
	text := "太阳直射点"
	if language == "en" {
		text = "Subsolar"
	}
	return solarEclipseMapPlaceLabel(labels, solarEclipseMapPointLabel{x: x, y: y, text: text, fontSize: 7},
		[]svgchart.LabelPlacement{
			{X: x, Y: y - 9, Anchor: "middle"},
			{X: x, Y: y + 16, Anchor: "middle"},
			{X: x + 7, Y: y + 4, Anchor: "start"},
			{X: x - 7, Y: y + 4, Anchor: "end"},
			{X: x, Y: y - 22, Anchor: "middle"},
			{X: x, Y: y + 29, Anchor: "middle"},
			{X: x + 16, Y: y + 4, Anchor: "start"},
			{X: x - 16, Y: y + 4, Anchor: "end"},
		}, frame)
}

func solarEclipseMapPlaceTimeLabels(
	path eclipsecore.SolarEclipsePath,
	frame svgmap.Frame,
	options SolarEclipseMapSVGOptions,
	labels *svgchart.LabelTable,
) []solarEclipseMapPointLabel {
	if options.TimeLabelStep <= 0 || len(path.CenterLine) < 2 {
		return nil
	}
	excluded := []time.Time{path.Eclipse.CentralBeginOnEarth, path.Eclipse.CentralEndOnEarth}
	markers := solarEclipseTimeMarkerPoints(path.CenterLine, options.TimeLabelStep, options.Location, excluded)
	result := make([]solarEclipseMapPointLabel, 0, len(markers))
	projected := make([][2]float64, 0, len(markers))
	for _, marker := range markers {
		x, y, visible := frame.Project(marker.Longitude, marker.Latitude)
		if !visible || x < frame.X+22 || x > frame.X+frame.Width-22 || y < frame.Y+12 || y > frame.Y+frame.Height-12 {
			continue
		}
		tooClose := false
		for _, previous := range projected {
			if math.Hypot(x-previous[0], y-previous[1]) < 44 {
				tooClose = true
				break
			}
		}
		if tooClose {
			continue
		}
		projected = append(projected, [2]float64{x, y})
		preferredY := y - 8
		if mapTimesNear(marker.Time, path.Eclipse.GreatestEclipse, solarEclipseGreatestTimeLabelWindow(options.TimeLabelStep)) {
			preferredY = y + 15
		} else if preferredY < frame.Y+10 {
			preferredY = y + 15
		}
		text := marker.Time.In(options.Location).Format("15:04")
		label := solarEclipseMapPlaceLabel(labels, solarEclipseMapPointLabel{x: x, y: y, text: text, fontSize: 9},
			[]svgchart.LabelPlacement{
				{X: x, Y: preferredY, Anchor: "middle"},
				{X: x, Y: y + 15, Anchor: "middle"},
				{X: x, Y: y - 8, Anchor: "middle"},
				{X: x + 11, Y: y + 4, Anchor: "start"},
				{X: x - 11, Y: y + 4, Anchor: "end"},
				{X: x, Y: y + 27, Anchor: "middle"},
				{X: x, Y: y - 21, Anchor: "middle"},
			}, frame)
		result = append(result, label)
	}
	return result
}

// solarEclipseContourLabelCandidates 先给首选点，再沿最长支路从中间向两端取候选位次，标注始终贴着等值线。
func solarEclipseContourLabelCandidates(
	segments [][]eclipsecore.SolarEclipsePathPoint,
	frame svgmap.Frame,
	preferredX, preferredY float64,
	hasPreferred bool,
) []svgchart.LabelPlacement {
	candidates := make([]svgchart.LabelPlacement, 0, 48)
	if hasPreferred {
		candidates = append(candidates, svgchart.LabelPlacement{X: preferredX, Y: preferredY, Anchor: "middle"})
	}
	longest := -1
	for index, segment := range segments {
		if longest < 0 || len(segment) > len(segments[longest]) {
			longest = index
		}
	}
	if longest < 0 {
		return candidates
	}
	segment := segments[longest]
	middle := len(segment) / 2
	for offset := 0; offset <= middle; offset++ {
		for _, index := range []int{middle + offset, middle - offset} {
			if index < 0 || index >= len(segment) {
				continue
			}
			point := segment[index]
			x, y, visible := frame.Project(point.Longitude, point.Latitude)
			if !visible ||
				x < frame.X+30 || x > frame.X+frame.Width-30 ||
				y < frame.Y+18 || y > frame.Y+frame.Height-18 {
				continue
			}
			candidates = append(candidates, svgchart.LabelPlacement{X: x, Y: y - 6, Anchor: "middle"})
			if len(candidates) >= 48 {
				return candidates
			}
		}
	}
	return candidates
}

// solarEclipseScaleStep 在槽位宽度内取最大的整刻度，required 给出该刻度下刻度条加标注的总横向占用。
func solarEclipseScaleStep(target, available float64, required func(step float64) float64) float64 {
	if target <= 0 || available <= 0 {
		return 0
	}
	start := int(math.Floor(math.Log10(target)))
	for exponent := start + 1; exponent >= start-4; exponent-- {
		decade := math.Pow(10, float64(exponent))
		for _, multiplier := range []float64{5, 2, 1} {
			step := decade * multiplier
			if required(step) <= available {
				return step
			}
		}
	}
	return 0
}
