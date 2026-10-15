package svg

import (
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/internal/svgmap"
)

// solarEclipseDetailedMapLayout 给出 NASA 摆法：球面居中放大，阶段面板在球面下方分三栏。
// solarEclipseMapAspect 返回各投影下地图框的长宽比：圆盘类为 1，等经纬为 2。
func solarEclipseMapAspect(projection svgmap.Projection) float64 {
	if projection == svgmap.ProjectionEquirectangular {
		return 2
	}
	return 1
}

// solarEclipsePortraitMapLayout 是竖版详细版式：地图居中在上、数据块两行在下。
func solarEclipsePortraitMapLayout(
	width, height float64,
	projection svgmap.Projection,
	center svgmap.GeoPoint,
	legendRows int,
) solarEclipseMapLayout {
	aspect := solarEclipseMapAspect(projection)
	margin := math.Max(30, math.Min(52, width*0.044))
	headerHeight := math.Max(190, height*0.175)
	blockHeight := math.Max(76, height*0.058)
	footerHeight := math.Max(64, height*0.05)
	legendHeight := math.Max(46, solarEclipseMapLegendBlockHeight(legendRows))
	// 数据块最多六行，行高必须容下六行文字。
	panelHeight := math.Max(solarEclipseMinPanelHeight(6), height*0.10)
	gap := math.Max(16, height*0.014)
	panelGap := math.Max(18, width*0.022)
	reserved := headerHeight + blockHeight + footerHeight + legendHeight + 2*panelHeight + 3*gap
	mapWidth := math.Min(width-2*margin, math.Max(140, height-reserved)*aspect)
	if mapWidth < 140 {
		mapWidth = math.Max(140, width-2*margin)
	}
	mapHeight := mapWidth / aspect
	frame := svgmap.Frame{
		X:     (width - mapWidth) / 2,
		Y:     headerHeight + blockHeight,
		Width: mapWidth, Height: mapHeight, Projection: projection,
		CenterLongitude: center.Longitude, CenterLatitude: center.Latitude,
	}
	// 第一行三栏放阶段信息，第二行中间留给比例尺，两侧放历表常数与天平动。
	panelY := frame.Y + mapHeight + gap + 14
	panel2Y := panelY + panelHeight + gap
	columnWidth := (width - 2*margin - 2*panelGap) / 3
	scaleSlot := svgchart.LabelBox{
		X: margin + columnWidth + panelGap, Y: panel2Y, Width: columnWidth, Height: panelHeight,
	}
	return solarEclipseMapLayout{
		frame:        frame,
		panelX:       margin,
		panelY:       panelY,
		panelWidth:   width - 2*margin,
		panelHeight:  panelHeight,
		nasa:         true,
		legendY:      panel2Y + panelHeight + gap,
		scaleY:       scaleSlot.Y + scaleSlot.Height*0.45,
		margin:       margin,
		panelGap:     panelGap,
		blockY:       headerHeight,
		blockHeight:  blockHeight,
		secondPanelY: panel2Y,
		scaleSlot:    scaleSlot,
	}
}

// writeSolarEclipseScaleBar 在球面图下方画比例尺；球面半径对应 90° 大圆弧长。
// 刻度长度由槽位宽度反推，标注不会越过槽位压到旁边的数据块上。
func writeSolarEclipseScaleBar(
	builder *strings.Builder,
	layout solarEclipseMapLayout,
	options SolarEclipseMapSVGOptions,
) {
	if !layout.nasa {
		return
	}
	slot := layout.scaleSlot
	radius := layout.frame.Width / 2
	if radius <= 0 || slot.Width <= 0 || slot.Height <= 0 {
		return
	}
	kilometresPerPixel := (math.Pi / 2 * geodata.EarthMeanRadiusKM) / radius
	target := 130.0 * kilometresPerPixel
	step := solarEclipseScaleStep(target, slot.Width, func(candidate float64) float64 {
		barWidth := candidate / kilometresPerPixel
		return barWidth + svgchart.EstimatedTextWidth("0", 10) +
			svgchart.EstimatedTextWidth(formatSolarEclipseScale(candidate)+" km", 10) + 24
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
	label := "比例尺"
	if options.Language == "en" {
		label = "Scale"
	}
	fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#334155" font-family="Arial, sans-serif" font-size="10" text-anchor="end">0</text>`,
		left-8, y+4)
	fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#334155" font-family="Arial, sans-serif" font-size="10">%s km</text>`,
		left+barWidth+8, y+4, html.EscapeString(formatSolarEclipseScale(step)))
	fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#7b8794" font-family="Arial, sans-serif" font-size="10" text-anchor="middle">%s</text>`,
		centreX, y+20, html.EscapeString(label))
	builder.WriteString(`</g>`)
}

func formatSolarEclipseScale(kilometres float64) string {
	if kilometres >= 1000 {
		return fmt.Sprintf("%.0f", kilometres)
	}
	return fmt.Sprintf("%.0f", kilometres)
}

func writeSolarEclipseDetailedPanels(builder *strings.Builder, cells []solarEclipsePanelCell, layout solarEclipseMapLayout) {
	for _, cell := range cells {
		if cell.box.Y < layout.secondPanelY {
			svgchart.WritePanelBox(builder, "solar-detailed-panel", cell.box.X, cell.box.Y,
				cell.box.Width, cell.box.Height, cell.title, cell.rows)
		}
	}
}

var _ = time.Second

// formatSolarEclipseRA 把赤经度数写成 HHhMMmSS.Ss。
func formatSolarEclipseRA(degrees float64) string {
	total := math.Mod(degrees, 360) / 15
	hours := math.Floor(total)
	minutes := math.Floor((total - hours) * 60)
	seconds := ((total-hours)*60 - minutes) * 60
	return fmt.Sprintf("%02.0fh%02.0fm%04.1fs", hours, minutes, seconds)
}

// formatSolarEclipseDec 把赤纬度数写成 ±DD°MM'SS.S"。
func formatSolarEclipseDec(degrees float64) string {
	sign := "+"
	if degrees < 0 {
		sign = "-"
		degrees = -degrees
	}
	whole := math.Floor(degrees)
	minutes := math.Floor((degrees - whole) * 60)
	seconds := ((degrees-whole)*60 - minutes) * 60
	return fmt.Sprintf("%s%02.0f°%02.0f'%04.1f\"", sign, whole, minutes, seconds)
}

// formatSolarEclipseArcsec 把角秒写成 DD°MM'SS.S"。
func formatSolarEclipseArcsec(arcsec float64) string {
	whole := math.Floor(arcsec / 3600)
	minutes := math.Floor((arcsec/3600 - whole) * 60)
	seconds := ((arcsec/3600-whole)*60 - minutes) * 60
	return fmt.Sprintf("%02.0f°%02.0f'%04.1f\"", whole, minutes, seconds)
}

func writeSolarEclipseGeocentricBlocks(
	builder *strings.Builder,
	panel eclipsecore.SolarEclipseGeocentricPanel,
	layout solarEclipseMapLayout,
	options SolarEclipseMapSVGOptions,
) {
	if !layout.nasa {
		return
	}
	sunTitle, moonTitle := solarEclipseGeocentricTitles(options)
	sunRows, moonRows := solarEclipseGeocentricRows(panel)
	if options.Language == "en" {
		sunRows[3].Label, moonRows[3].Label = "H.P.", "H.P."
	}
	blocks := [2]solarEclipsePanelBlock{
		{title: sunTitle, rows: sunRows},
		{title: moonTitle, rows: moonRows},
	}
	// 两块都按固定宽度排版，数值一律右对齐到块右缘，右侧块才不会顶出图框。
	blockWidth := math.Min(212, (layout.panelWidth-40)/2)
	for index, block := range blocks {
		x := layout.margin
		if index == 1 {
			x = layout.margin + layout.panelWidth - blockWidth
		}
		y := layout.blockY
		fmt.Fprintf(builder, `<g class="solar-geocentric-block">`)
		fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#111827" font-family="Georgia, 'Times New Roman', serif" font-size="12" font-weight="700">%s</text>`,
			x, y+11, html.EscapeString(block.title))
		for rowIndex, row := range block.rows {
			rowY := y + 30 + float64(rowIndex)*17
			fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#475569" font-family="Arial, sans-serif" font-size="10.5">%s</text>`,
				x, rowY, html.EscapeString(row.Label))
			fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#0f172a" font-family="Arial, sans-serif" font-size="10.5" font-weight="700" text-anchor="end">%s</text>`,
				x+blockWidth, rowY, html.EscapeString(row.Value))
		}
		builder.WriteString(`</g>`)
	}
}

func writeSolarEclipseEphemerisPanels(builder *strings.Builder, cells []solarEclipsePanelCell, layout solarEclipseMapLayout) {
	if !layout.nasa {
		return
	}
	for _, cell := range cells {
		if cell.box.Y < layout.secondPanelY {
			continue
		}
		svgchart.WritePanelBox(builder, "solar-detailed-panel", cell.box.X, cell.box.Y,
			cell.box.Width, cell.box.Height, cell.title, cell.rows)
	}
}

// solarEclipseCardinalLabel 是圆盘四边的方向标记。
type solarEclipseCardinalLabel struct {
	text string
	x, y float64
}

// solarEclipseCardinalLabels 给出圆盘四边的方向标记位次；平面图没有圆盘边界，返回空。
func solarEclipseCardinalLabels(layout solarEclipseMapLayout) []solarEclipseCardinalLabel {
	if !layout.nasa || !layout.frame.IsDisk() {
		return nil
	}
	centreX := layout.frame.X + layout.frame.Width/2
	centreY := layout.frame.Y + layout.frame.Height/2
	radius := math.Min(layout.frame.Width, layout.frame.Height)/2 + 13
	result := make([]solarEclipseCardinalLabel, 0, 4)
	for index, text := range []string{"N", "E", "S", "W"} {
		angle := float64(index) * math.Pi / 2
		result = append(result, solarEclipseCardinalLabel{
			text: text,
			x:    centreX + radius*math.Sin(angle),
			y:    centreY - radius*math.Cos(angle) + 4,
		})
	}
	return result
}

// writeSolarEclipseCardinalMarkers 在圆盘四边标出 N/E/S/W。
func writeSolarEclipseCardinalMarkers(builder *strings.Builder, layout solarEclipseMapLayout) {
	for _, cardinal := range solarEclipseCardinalLabels(layout) {
		fmt.Fprintf(builder, `<text class="solar-cardinal" x="%.3f" y="%.3f" fill="#334155" font-family="Georgia, 'Times New Roman', serif" font-size="13" font-weight="700" text-anchor="middle">%s</text>`,
			cardinal.x, cardinal.y, cardinal.text)
	}
}
