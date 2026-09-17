package svg

import (
	"fmt"
	"html"
	"strings"

	"github.com/starainrt/astro/moon"
)

// writeOccultationGreatestTimeContours 绘制地方掩甚时刻等值线及其 HH:MM 标注。
func writeOccultationGreatestTimeContours(
	b *strings.Builder,
	contours []moon.OccultationGreatestTimeContour,
	layout starOccultationSVGLayout,
	options StarOccultationSVGOptions,
	placer *occultationLabelPlacer,
) {
	if len(contours) == 0 {
		return
	}
	type isochroneLabel struct {
		x, y float64
		text string
	}
	labels := make([]isochroneLabel, 0, len(contours))
	for _, contour := range contours {
		for _, segment := range contour.Segments {
			writeStarOccultationGeoLine(b, segment, layout,
				occultationLineStyle{className: "greatest-time-isoline", color: "#1f6fb2", strokeWidth: 1.1})
		}
		text := contour.Time.In(options.Location).Format("15:04")
		if x, y, ok := occultationGreatestTimeIsolineLabelPlacement(contour, text, layout, placer); ok {
			labels = append(labels, isochroneLabel{x: x, y: y, text: text})
		}
	}
	for _, label := range labels {
		fmt.Fprintf(b, `<text class="greatest-time-isoline-label" x="%.3f" y="%.3f" fill="#1a5c96" stroke="#ffffff" stroke-width="2.6" paint-order="stroke" font-family="Arial, sans-serif" font-size="9" font-weight="700" text-anchor="middle">%s</text>`,
			label.x, label.y, html.EscapeString(label.text))
	}
}

// occultationGreatestTimeIsolineLabelPlacement 在最长的支路上取首个未被其它标签占用的标注点。
func occultationGreatestTimeIsolineLabelPlacement(
	contour moon.OccultationGreatestTimeContour,
	text string,
	layout starOccultationSVGLayout,
	placer *occultationLabelPlacer,
) (float64, float64, bool) {
	longest := -1
	for index, segment := range contour.Segments {
		if longest < 0 || len(segment) > len(contour.Segments[longest]) {
			longest = index
		}
	}
	if longest < 0 {
		return 0, 0, false
	}
	segment := contour.Segments[longest]
	frame := layout.mapFrame()
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
			// 优先压在等时线上方，其次下方；两者都被占用就换下一个候选点。
			for _, baseline := range []float64{y - 6, y + 13} {
				if placer.place(x, baseline, "middle", text, 9, 2) {
					return x, baseline, true
				}
			}
		}
	}
	return 0, 0, false
}
