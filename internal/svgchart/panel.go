// Package svgchart 提供日月食与月掩星详细版式共用的数据块与版面计算。
// Package svgchart holds the data blocks and page geometry shared by the detailed eclipse and occultation charts.
package svgchart

import (
	"fmt"
	"html"
	"math"
	"strings"
)

// PanelRow 是数据块里的一行；Detail 非空时在数值下方补一行灰色小字。
type PanelRow struct {
	Label  string
	Value  string
	Detail string
}

// PanelBlock 是一个带标题的数据块。
type PanelBlock struct {
	Title string
	Rows  []PanelRow
}

// WritePanelBox 画一个数据块：标题加若干行“标签 + 数值”。
// 行高按行数等分，因此块高变化时不会溢出；Class 用于区分各族图的样式钩子。
func WritePanelBox(
	b *strings.Builder,
	class string,
	x, y, width, height float64,
	title string,
	rows []PanelRow,
) {
	if len(rows) == 0 || height <= 0 {
		return
	}
	fmt.Fprintf(b, `<g class="%s">`, class)
	fmt.Fprintf(b, `<rect x="%.3f" y="%.3f" width="%.3f" height="%.3f" rx="3" fill="#fbfcfd" stroke="#dde3e9"/>`,
		x, y, width, height)
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#111827" font-family="Georgia, 'Times New Roman', serif" font-size="12" font-weight="700">%s</text>`,
		x+12, y+20, html.EscapeString(title))
	rowHeight := (height - panelBoxHeaderHeight) / float64(len(rows))
	for index, row := range rows {
		if row.Label == "" {
			continue
		}
		rowY := y + panelBoxHeaderHeight + float64(index)*rowHeight
		size := PanelRowFontSize(width, row)
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#475569" font-family="Arial, sans-serif" font-size="%.1f">%s</text>`,
			x+panelBoxPadding, rowY+11, size, html.EscapeString(row.Label))
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#0f172a" font-family="Arial, sans-serif" font-size="%.1f" font-weight="700" text-anchor="end">%s</text>`,
			x+width-panelBoxPadding, rowY+11, size, html.EscapeString(row.Value))
		if row.Detail != "" {
			fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#7b8794" font-family="Arial, sans-serif" font-size="%d">%s</text>`,
				x+panelBoxPadding, rowY+22, panelRowDetailSize, html.EscapeString(row.Detail))
		}
	}
	b.WriteString(`</g>`)
}

const (
	panelBoxHeaderHeight = 30
	panelRowLabelSize    = 10.5
	panelRowDetailSize   = 9
	panelBoxPadding      = 12
	panelLabelValueGap   = 8
	panelMinimumFontSize = 8
)

// EstimatedTextWidth 按字号估算文本宽度：ASCII 取 0.62em，非 ASCII（含全角中文）取 1.05em。
// 没有字体度量可用，这里只服务于压叠判定，估值偏大以保证判定保守。
// EstimatedTextWidth estimates text width from the font size: 0.62em per ASCII rune and 1.05em
// otherwise. It exists only for overlap checks, so it deliberately over-estimates.
func EstimatedTextWidth(text string, fontSize float64) float64 {
	units := 0.0
	for _, symbol := range text {
		if symbol < 128 {
			units += 0.62
			continue
		}
		units += 1.05
	}
	return units * fontSize
}

// EstimatedTextExtents 返回文本相对基线的上下留白，用于构造占位矩形。
// EstimatedTextExtents returns the space a text keeps above and below its baseline.
func EstimatedTextExtents(fontSize float64) (above, below float64) {
	return 0.85 * fontSize, 0.30 * fontSize
}

// PanelRowFontSize 返回该行“标签 + 数值”放不下时等比缩小后的字号；缩小到下限仍放不下说明块太窄。
// PanelRowFontSize shrinks a row's font until its label and value fit side by side.
func PanelRowFontSize(width float64, row PanelRow) float64 {
	content := width - 2*panelBoxPadding
	if content <= 0 {
		return panelMinimumFontSize
	}
	pair := EstimatedTextWidth(row.Label, panelRowLabelSize) + EstimatedTextWidth(row.Value, panelRowLabelSize)
	if pair+panelLabelValueGap <= content {
		return panelRowLabelSize
	}
	size := panelRowLabelSize * (content - panelLabelValueGap) / pair
	if size < panelMinimumFontSize {
		return panelMinimumFontSize
	}
	return size
}

func panelTextRect(x, baseline, fontSize float64, text string, anchorEnd bool) panelTextBox {
	width := EstimatedTextWidth(text, fontSize)
	left := x
	if anchorEnd {
		left = x - width
	}
	above, below := EstimatedTextExtents(fontSize)
	return panelTextBox{x: left, y: baseline - above, width: width, height: above + below}
}

// panelTextBox 是文本的保守包围盒。
type panelTextBox struct {
	x, y, width, height float64
}

func panelTextBoxesOverlap(first, second panelTextBox) bool {
	left, right := math.Max(first.x, second.x), math.Min(first.x+first.width, second.x+second.width)
	top, bottom := math.Max(first.y, second.y), math.Min(first.y+first.height, second.y+second.height)
	return right-left > 0.5 && bottom-top > 0
}

// PanelBoxOverlaps 报告该块在给定位次与尺寸下，块内文本是否会互相压叠。
// 行高按行数等分，所以块太矮时行距会小于文字高度，此时必须由调用方拒绝该画布而不是画出压叠的文字。
// PanelBoxOverlaps reports whether the block's own texts would collide at the given box; callers must
// reject a canvas whose panels cannot keep the rows apart.
func PanelBoxOverlaps(x, y, width, height float64, title string, rows []PanelRow) bool {
	if len(rows) == 0 || height <= 0 {
		return false
	}
	boxes := make([]panelTextBox, 0, 2*len(rows)+1)
	boxes = append(boxes, panelTextRect(x+12, y+20, 12, title, false))
	rowHeight := (height - panelBoxHeaderHeight) / float64(len(rows))
	for index, row := range rows {
		if row.Label == "" {
			continue
		}
		rowY := y + panelBoxHeaderHeight + float64(index)*rowHeight
		size := PanelRowFontSize(width, row)
		boxes = append(boxes, panelTextRect(x+panelBoxPadding, rowY+11, size, row.Label, false))
		boxes = append(boxes, panelTextRect(x+width-panelBoxPadding, rowY+11, size, row.Value, true))
		if row.Detail != "" {
			boxes = append(boxes, panelTextRect(x+panelBoxPadding, rowY+22, panelRowDetailSize, row.Detail, false))
		}
	}
	for first := 0; first < len(boxes); first++ {
		for second := first + 1; second < len(boxes); second++ {
			if panelTextBoxesOverlap(boxes[first], boxes[second]) {
				return true
			}
		}
	}
	return false
}
