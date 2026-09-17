package svg

import (
	"fmt"
	"html"
	"strings"

	"github.com/starainrt/astro/internal/svgchart"
)

// occultationLabelRect 是画布坐标下的标签占位矩形。
type occultationLabelRect struct {
	left, top, right, bottom float64
}

func (rect occultationLabelRect) overlaps(other occultationLabelRect) bool {
	return rect.left < other.right && other.left < rect.right && rect.top < other.bottom && other.top < rect.bottom
}

func (rect occultationLabelRect) expanded(padding float64) occultationLabelRect {
	return occultationLabelRect{
		left: rect.left - padding, top: rect.top - padding,
		right: rect.right + padding, bottom: rect.bottom + padding,
	}
}

// occultationLabelPlacer 是三族地图标签（等时线、中心线时刻、事件标记）共用的已放置表；
// nil 表示不做避让。
type occultationLabelPlacer struct {
	placed []occultationLabelRect
}

func (placer *occultationLabelPlacer) reserved(rect occultationLabelRect) bool {
	if placer == nil {
		return false
	}
	for _, current := range placer.placed {
		if rect.overlaps(current) {
			return true
		}
	}
	return false
}

// place 在占位可用时登记并返回 true；padding 是标签之间要求的最小空隙。
func (placer *occultationLabelPlacer) place(x, y float64, anchor, text string, fontSize, padding float64) bool {
	rect := occultationLabelRectFor(x, y, anchor, text, fontSize).expanded(padding)
	if placer.reserved(rect) {
		return false
	}
	placer.reserve(rect)
	return true
}

func (placer *occultationLabelPlacer) reserve(rect occultationLabelRect) {
	if placer == nil {
		return
	}
	placer.placed = append(placer.placed, rect)
}

// occultationLabelTextWidth 估算标签宽度，口径与数据块的压叠判定一致。
func occultationLabelTextWidth(text string, fontSize float64) float64 {
	return svgchart.EstimatedTextWidth(text, fontSize)
}

// occultationLabelRectFor 返回锚点为 (x, y) 的文本占位，上下留白与数据块同口径。
func occultationLabelRectFor(x, y float64, anchor, text string, fontSize float64) occultationLabelRect {
	width := occultationLabelTextWidth(text, fontSize)
	above, below := svgchart.EstimatedTextExtents(fontSize)
	left := x
	switch anchor {
	case "middle":
		left = x - width/2
	case "end":
		left = x - width
	}
	return occultationLabelRect{
		left: left, top: y - above, right: left + width, bottom: y + below,
	}
}

// occultationLegendItem 是图例中的一项：色块（fill）与线段（color）至少有一个。
type occultationLegendItem struct {
	label       string
	color       string
	dash        string
	markerWidth float64
	fill        string
	fillOpacity float64
}

const (
	occultationLegendRowHeight       = 18.0
	occultationLegendMinimumFontSize = 7.0
)

func occultationLegendItemWidth(item occultationLegendItem, fontSize float64) float64 {
	return item.markerWidth + 6 + occultationLabelTextWidth(item.label, fontSize)
}

// occultationLegendItemRows 按画布宽度把图例项贪心排成多行；行数不得超过布局预留的两行。
func occultationLegendItemRows(
	layout starOccultationSVGLayout,
	items []occultationLegendItem,
	fontSize float64,
) [][]occultationLegendItem {
	rows := make([][]occultationLegendItem, 0, 2)
	current := make([]occultationLegendItem, 0, len(items))
	width := 0.0
	for _, item := range items {
		itemWidth := occultationLegendItemWidth(item, fontSize)
		// 布局只预留两行，第三行会压到页脚；超出的项并回第二行而不是换行。
		if len(current) > 0 && len(rows) < 1 && width+itemWidth > layout.mapWidth {
			rows = append(rows, current)
			current = make([]occultationLegendItem, 0, len(items))
			width = 0
		}
		current = append(current, item)
		width += itemWidth
	}
	if len(current) > 0 {
		rows = append(rows, current)
	}
	return rows
}

// occultationLegendFontSize 在放得下的前提下取最大字号：窄地图（详细版式）必须缩字号，
// 否则图例项会溢出地图横向范围压到右侧数据块上。
func occultationLegendFontSize(layout starOccultationSVGLayout, items []occultationLegendItem, fontSize float64) float64 {
	for candidate := fontSize; candidate >= occultationLegendMinimumFontSize; candidate-- {
		rows := occultationLegendItemRows(layout, items, candidate)
		fits := true
		for _, row := range rows {
			total := 0.0
			for _, item := range row {
				total += occultationLegendItemWidth(item, candidate)
			}
			if total > layout.mapWidth {
				fits = false
				break
			}
		}
		if fits {
			return candidate
		}
	}
	return occultationLegendMinimumFontSize
}

// writeOccultationLegendItems 逐行写图例；每行内的项按剩余空间均分间隔，避免文字互相压叠。
func writeOccultationLegendItems(
	b *strings.Builder,
	top float64,
	layout starOccultationSVGLayout,
	items []occultationLegendItem,
	fontSize float64,
) {
	fontSize = occultationLegendFontSize(layout, items, fontSize)
	rows := occultationLegendItemRows(layout, items, fontSize)
	y := top
	for _, row := range rows {
		total := 0.0
		for _, item := range row {
			total += occultationLegendItemWidth(item, fontSize)
		}
		gap := 0.0
		if len(row) > 0 && total < layout.mapWidth {
			gap = (layout.mapWidth - total) / float64(len(row))
		}
		x := layout.mapX
		for _, item := range row {
			if item.fill != "" {
				fmt.Fprintf(b, `<rect x="%.3f" y="%.3f" width="%.3f" height="9" fill="%s" fill-opacity="%.2f"/>`,
					x, y-10, item.markerWidth, item.fill, item.fillOpacity)
			}
			if item.color != "" {
				fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="%s" stroke-width="2"`,
					x, y-5, x+item.markerWidth, y-5, item.color)
				if item.dash != "" {
					fmt.Fprintf(b, ` stroke-dasharray="%s"`, item.dash)
				}
				b.WriteString(`/>`)
			}
			fmt.Fprintf(b, `<text class="occultation-legend-text" x="%.3f" y="%.3f" fill="#465053" font-family="Arial, sans-serif" font-size="%.0f">%s</text>`,
				x+item.markerWidth+6, y, fontSize, html.EscapeString(item.label))
			x += occultationLegendItemWidth(item, fontSize) + gap
		}
		y += occultationLegendRowHeight
	}
}
