package svg

import (
	"fmt"
	"html"
	"math"
	"strings"

	"github.com/starainrt/astro/internal/svgchart"
)

// localOccultationOverviewLabel 是总览区一个接触标签：文本与它在图上的接触点位置。
type localOccultationOverviewLabel struct {
	text string
	x, y float64
}

// writeLocalOccultationOverviewLabels 排布总览区的接触标签：先按原偏移试位，压叠时沿接触点径向
// 外推、再上下错开，仍放不下就贴着月球外缘排成竖列；离接触点较远的标签用细引线连回去。
func writeLocalOccultationOverviewLabels(
	b *strings.Builder,
	labels []localOccultationOverviewLabel,
	table *svgchart.LabelTable,
	cx, cy, moonRadius float64,
	color string,
) {
	if len(labels) == 0 {
		return
	}
	const fontSize = 11.0
	spillTop := cy - moonRadius - 24
	spill := 0
	for _, item := range labels {
		anchor, side := "start", 1.0
		if item.x > cx {
			anchor, side = "end", -1
		}
		baseY := item.y - 9
		if item.y < cy-20 {
			baseY = item.y + 15
		}
		candidates := []svgchart.LabelPlacement{{X: item.x + side*8, Y: baseY, Anchor: anchor}}
		ux, uy := item.x-cx, item.y-cy
		if norm := math.Hypot(ux, uy); norm > 1e-9 {
			ux, uy = ux/norm, uy/norm
		} else {
			ux, uy = 0, -1
		}
		for step := 1; step <= 5; step++ {
			reach := 12 + float64(step)*13
			candidates = append(candidates,
				svgchart.LabelPlacement{X: item.x + ux*reach + side*6, Y: item.y + uy*reach, Anchor: anchor},
				svgchart.LabelPlacement{X: item.x + ux*reach - side*6, Y: item.y + uy*reach, Anchor: svgchart.MirrorLabelAnchor(anchor)},
			)
		}
		for step := 1; step <= 4; step++ {
			candidates = append(candidates,
				svgchart.LabelPlacement{X: item.x + side*8, Y: baseY - float64(step)*15, Anchor: anchor},
				svgchart.LabelPlacement{X: item.x + side*8, Y: baseY + float64(step)*15, Anchor: anchor},
			)
		}
		placement, placed := table.Place(item.text, fontSize, candidates)
		if !placed {
			placement = svgchart.LabelPlacement{X: cx + side*(moonRadius+30), Y: spillTop + float64(spill)*15, Anchor: anchor}
			for spill < 40 && table.OverlapsText(placement.X, placement.Y, fontSize, item.text, placement.Anchor) {
				spill++
				placement.Y = spillTop + float64(spill)*15
			}
			table.ReserveText(placement.X, placement.Y, fontSize, item.text, placement.Anchor)
			spill++
		}
		if math.Hypot(placement.X-item.x, placement.Y-item.y) > 16 {
			fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="%s" stroke-width="0.7" opacity="0.7"/>`,
				item.x, item.y, placement.X, placement.Y, color)
		}
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="%s" stroke="#ffffff" stroke-width="3" paint-order="stroke" font-family="Arial, sans-serif" font-size="%.0f" font-weight="700" text-anchor="%s">%s</text>`,
			placement.X, placement.Y, color, fontSize, placement.Anchor, html.EscapeString(item.text))
	}
}
