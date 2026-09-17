package svg

import (
	"fmt"
	"html"
	"math"
	"strings"

	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/internal/svgmap"
)

// 横版详细版式的固定量：数据块栏宽、栏数行数与地图长宽比。
const (
	solarEclipseLandscapeCellWidth = 230.0
	solarEclipseLandscapeColumns   = 2
	solarEclipseLandscapeRows      = 3
	// 等经纬地图经度跨 360°，比例尺按赤道弧长换算。
	solarEclipseEquirectangularAspect = 2.0
	solarEclipseEarthCircumferenceKM  = 40075.017
	// 横带上下限：低于这些值时数据块行距会压到文字高度以下，画布必须拒绝。
	// 表头下限让开摘要最后一行；横带与图例之间的空隙还要容下末行文字的伸出部分。
	solarEclipseLandscapeMinHeader = 156.0
	solarEclipseLandscapeMinFooter = 44.0
	solarEclipseLandscapeMinGap    = 12.0
	solarEclipseLandscapeMinRowGap = 8.0
	solarEclipseLandscapeMinLegend = 42.0
	// 比例尺槽最小高度：刻度上下各要一行标注。
	solarEclipseScaleSlotMinHeight = 46.0
)

// 横版两栏三行的行序固定为日月地心块、接触时刻块、地方情况与历表块，最大行数依次是 4、4、6。
var solarEclipseLandscapeDefaultRowFields = [solarEclipseLandscapeRows]float64{4, 4, 6}

// solarEclipseLandscapeRowFields 数出该事件下每行最多几行文字：行高必须按它分配。
func solarEclipseLandscapeRowFields(blocks solarEclipsePanelBlocks) [solarEclipseLandscapeRows]float64 {
	fields := [solarEclipseLandscapeRows]float64{}
	for index, block := range solarEclipseLandscapeBlockOrder(blocks) {
		row := index / solarEclipseLandscapeColumns
		if row >= solarEclipseLandscapeRows {
			continue
		}
		if count := float64(len(block.rows)); count > fields[row] {
			fields[row] = count
		}
	}
	return fields
}

// solarEclipseLandscapeRowHeights 把横带高度分给三行：行距一致，行高为表头加上行数乘行距；空行高度为 0 也不占行距。
func solarEclipseLandscapeRowHeights(
	bandHeight, rowGap float64,
	fields [solarEclipseLandscapeRows]float64,
) ([solarEclipseLandscapeRows]float64, float64) {
	heights := [solarEclipseLandscapeRows]float64{}
	totalFields, filled := 0.0, 0
	for _, count := range fields {
		totalFields += count
		if count > 0 {
			filled++
		}
	}
	if totalFields <= 0 || filled == 0 {
		return heights, 0
	}
	spacing := (bandHeight - float64(filled-1)*rowGap -
		float64(filled)*solarEclipsePanelHeaderHeight) / totalFields
	for index, count := range fields {
		if count <= 0 {
			continue
		}
		heights[index] = solarEclipsePanelHeaderHeight + count*spacing
	}
	return heights, spacing
}

// solarEclipseStripDivision 把地图下方那条分成天平动盒与比例尺槽，留不出比例尺高度时天平动盒吃掉整条。
func solarEclipseStripDivision(
	bottomY, bottomHeight, width, margin, gap, rowSpacing float64,
) (float64, svgchart.LabelBox) {
	natural := solarEclipsePanelHeaderHeight + 4*rowSpacing + 0.5
	if slotHeight := bottomHeight - natural - gap; slotHeight >= solarEclipseScaleSlotMinHeight {
		return natural, svgchart.LabelBox{X: margin, Y: bottomY + natural + gap, Width: width, Height: slotHeight}
	}
	shorter := bottomHeight - solarEclipseScaleSlotMinHeight - gap
	if shorter >= solarEclipseMinPanelHeight(4) {
		return shorter, svgchart.LabelBox{
			X: margin, Y: bottomY + shorter + gap,
			Width: width, Height: solarEclipseScaleSlotMinHeight,
		}
	}
	return bottomHeight, svgchart.LabelBox{}
}

// solarEclipseLandscapeMapLayout 是横版详细版式：地图在左，数据块两栏三行在右，
// 地图下方那条放天平动与比例尺。竖版是“地图在上、面板在下”，横版把排布轴换到水平方向。
func solarEclipseLandscapeMapLayout(
	width, height float64,
	projection svgmap.Projection,
	center svgmap.GeoPoint,
	legendRows int,
	rowFields [solarEclipseLandscapeRows]float64,
) solarEclipseMapLayout {
	aspect := solarEclipseMapAspect(projection)
	margin := math.Max(30, math.Min(52, width*0.044))
	headerHeight := math.Max(solarEclipseLandscapeMinHeader, height*0.17)
	footerHeight := math.Max(solarEclipseLandscapeMinFooter, height*0.04)
	legendHeight := math.Max(solarEclipseLandscapeMinLegend, solarEclipseMapLegendBlockHeight(legendRows))
	gap := math.Max(solarEclipseLandscapeMinGap, height*0.01)
	panelGap := math.Max(12, width*0.011)
	rowGap := math.Max(solarEclipseLandscapeMinRowGap, height*0.008)

	legendY := height - footerHeight - legendHeight
	bandTop := headerHeight + gap
	bandBottom := math.Max(bandTop, legendY-gap)
	bandHeight := bandBottom - bandTop

	gridWidth := solarEclipseLandscapeColumns*solarEclipseLandscapeCellWidth +
		float64(solarEclipseLandscapeColumns-1)*panelGap
	gridX := width - margin - gridWidth
	rowHeights, rowSpacing := solarEclipseLandscapeRowHeights(bandHeight, rowGap, rowFields)
	rowY := [solarEclipseLandscapeRows]float64{}
	gridCellHeight := 0.0
	cursor := bandTop
	for index := range rowHeights {
		rowY[index] = cursor
		if rowHeights[index] <= 0 {
			continue
		}
		cursor += rowHeights[index] + rowGap
		if rowHeights[index] > gridCellHeight {
			gridCellHeight = rowHeights[index]
		}
	}

	// 地图下方那条要放天平动与比例尺，先扣掉它的高度，圆盘图才不会把整条横带吃光。
	stripHeight := math.Max(150, bandHeight*0.24)
	mapMaxHeight := math.Max(120, bandHeight-gap-stripHeight)
	leftWidth := math.Max(240, gridX-gap-margin)
	mapWidth := leftWidth
	mapHeight := mapWidth / aspect
	if mapHeight > mapMaxHeight {
		mapHeight = mapMaxHeight
		mapWidth = mapHeight * aspect
	}
	mapY := bandTop
	bottomY := mapY + mapHeight + gap
	bottomHeight := math.Max(0, bandBottom-bottomY)
	stripPanelHeight, scaleSlot := solarEclipseStripDivision(bottomY, bottomHeight, leftWidth, margin, gap, rowSpacing)

	return solarEclipseMapLayout{
		frame: svgmap.Frame{
			X: margin + (leftWidth-mapWidth)/2, Y: mapY, Width: mapWidth, Height: mapHeight,
			Projection:      projection,
			CenterLongitude: center.Longitude, CenterLatitude: center.Latitude,
		},
		nasa:           true,
		landscape:      true,
		panelX:         gridX,
		panelY:         bandTop,
		panelWidth:     gridWidth,
		panelHeight:    gridCellHeight,
		panelGap:       panelGap,
		gridX:          gridX,
		gridCellWidth:  solarEclipseLandscapeCellWidth,
		gridCellHeight: gridCellHeight,
		gridGapX:       panelGap,
		gridGapY:       rowGap,
		gridRowY:       rowY,
		gridRowHeights: rowHeights,
		bottomY:        bottomY,
		stripHeight:    stripPanelHeight,
		scaleSlot:      scaleSlot,
		margin:         margin,
		legendY:        legendY,
		scaleY:         scaleSlot.Y + scaleSlot.Height*0.45,
	}
}

// solarEclipseLandscapeBlockOrder 是横版两栏三行里的数据块顺序。
func solarEclipseLandscapeBlockOrder(blocks solarEclipsePanelBlocks) []solarEclipsePanelBlock {
	return []solarEclipsePanelBlock{
		blocks.sun, blocks.moon,
		blocks.penumbra, blocks.umbra,
		blocks.circumstances, blocks.ephemeris,
	}
}

// writeSolarEclipseLandscapePanels 画横版的数据块网格、地图下方那条与比例尺。
func writeSolarEclipseLandscapePanels(
	builder *strings.Builder,
	cells []solarEclipsePanelCell,
	layout solarEclipseMapLayout,
	options SolarEclipseMapSVGOptions,
) {
	for _, cell := range cells {
		svgchart.WritePanelBox(builder, "solar-detailed-panel", cell.box.X, cell.box.Y,
			cell.box.Width, cell.box.Height, cell.title, cell.rows)
	}
	if layout.scaleSlot.Height <= 0 {
		return
	}
	if layout.frame.Projection == svgmap.ProjectionEquirectangular {
		writeSolarEclipseEquirectangularScaleBar(builder, layout)
	} else {
		writeSolarEclipseScaleBar(builder, layout, options)
	}
}

// writeSolarEclipseEquirectangularScaleBar 画等经纬图的比例尺；圆盘版式的换算不适用于平面图。
// 刻度长度由槽位宽度反推，绝不越过槽位压到数据块上。
func writeSolarEclipseEquirectangularScaleBar(builder *strings.Builder, layout solarEclipseMapLayout) {
	slot := layout.scaleSlot
	if slot.Width <= 0 || slot.Height <= 0 || layout.frame.Width <= 0 {
		return
	}
	kilometresPerPixel := solarEclipseEarthCircumferenceKM / layout.frame.Width
	target := 150.0 * kilometresPerPixel
	step := solarEclipseScaleStep(target, slot.Width, func(candidate float64) float64 {
		barWidth := candidate / kilometresPerPixel
		return barWidth + 0.5*svgchart.EstimatedTextWidth("0", 10) +
			0.5*svgchart.EstimatedTextWidth(formatSolarEclipseScale(candidate)+" km", 10) + 6
	})
	if step <= 0 {
		return
	}
	barWidth := step / kilometresPerPixel
	centreX := slot.X + slot.Width/2
	y := layout.scaleY
	left := centreX - barWidth/2
	fmt.Fprintf(builder, `<g class="solar-scale-bar"><line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#334155" stroke-width="1"/>`,
		left, y, left+barWidth, y)
	for _, tick := range []float64{0, 0.25, 0.5, 0.75, 1} {
		x := left + tick*barWidth
		length := 5.0
		if tick == 0 || tick == 1 {
			length = 9
		}
		fmt.Fprintf(builder, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#334155" stroke-width="1"/>`,
			x, y-length/2, x, y+length/2)
	}
	fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#334155" font-family="Arial, sans-serif" font-size="10" text-anchor="middle">0</text>`, left, y-10)
	fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#334155" font-family="Arial, sans-serif" font-size="10" text-anchor="middle">%s km</text>`,
		left+barWidth, y-10, html.EscapeString(formatSolarEclipseScale(step)))
	fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#7b8794" font-family="Arial, sans-serif" font-size="10" text-anchor="middle">比例尺</text>`,
		centreX, y+22)
	builder.WriteString(`</g>`)
}
