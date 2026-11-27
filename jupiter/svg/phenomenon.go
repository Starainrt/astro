package svg

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/jupiter"
)

const (
	phenomenonCenterX = 300.0
	phenomenonCenterY = 300.0
	phenomenonRadius  = 150.0
)

type phenomenonSample struct {
	code, name string
	date       time.Time
	diagram    basic.JupiterGalileanPhenomenonDiagram
	greatest   bool
	visibility contactVisibility
}

func phenomenonDiagramAt(date time.Time, satellite int) (basic.JupiterGalileanPhenomenonDiagram, bool) {
	return basic.JupiterGalileanPhenomenonDiagramWithInstance(basic.UTC2TT(basic.Date2JD(date.UTC())), satellite, nil)
}

func phenomenonEventValid(event jupiter.GalileanPhenomenonContactEvent) bool {
	if !event.Valid || event.Satellite < 1 || event.Satellite > 4 || event.Greatest.IsZero() {
		return false
	}
	switch event.Type {
	case jupiter.GalileanPhenomenonTransit, jupiter.GalileanPhenomenonShadowTransit, jupiter.GalileanPhenomenonEclipse, jupiter.GalileanPhenomenonOccultation:
	default:
		return false
	}
	if event.Type == jupiter.GalileanPhenomenonEclipse {
		return true
	}
	for _, contact := range []jupiter.GalileanPhenomenonContact{event.Disappearance, event.Reappearance} {
		if !contact.Valid || contact.Start.IsZero() || contact.End.IsZero() || contact.End.Before(contact.Start) {
			return false
		}
	}
	return !event.Disappearance.End.After(event.Greatest) && !event.Greatest.After(event.Reappearance.Start)
}

func phenomenonContacts(drawing *chart, event jupiter.GalileanPhenomenonContactEvent) ([]phenomenonSample, bool) {
	if event.Type == jupiter.GalileanPhenomenonEclipse {
		return jovianEclipseContacts(drawing, event)
	}
	names := phenomenonContactNames(drawing, event.Type)
	dates := []time.Time{event.Disappearance.Start, event.Disappearance.End, event.Greatest, event.Reappearance.Start, event.Reappearance.End}
	codes := []string{"1", "2", "", "3", "4"}
	if event.Type == jupiter.GalileanPhenomenonShadowTransit && !(dates[1].Equal(event.Greatest) && dates[3].Equal(event.Greatest)) {
		roots, ok := basic.JupiterGalileanShadowContactsWithInstance(basic.Date2JD(event.Disappearance.ModelCrossing.UTC()), basic.Date2JD(event.Reappearance.ModelCrossing.UTC()), event.Satellite, nil)
		if !ok {
			return nil, false
		}
		for index, slot := range [...]int{0, 1, 3, 4} {
			dates[slot] = basic.JD2DateByZone(roots[index], event.Greatest.Location(), false)
		}
	}
	samples := make([]phenomenonSample, 0, len(dates))
	for index, date := range dates {
		if (index == 1 || index == 3) && date.Equal(event.Greatest) {
			continue
		}
		diagram, ok := phenomenonDiagramAt(date, event.Satellite)
		if !ok {
			return nil, false
		}
		samples = append(samples, phenomenonSample{code: codes[index], name: names[index], date: date, diagram: diagram, greatest: index == 2})
	}
	return samples, true
}

func jovianEclipseContacts(drawing *chart, event jupiter.GalileanPhenomenonContactEvent) ([]phenomenonSample, bool) {
	minimum, ok := jovianEclipseMinimum(event.Greatest, event.Satellite)
	if !ok {
		return nil, false
	}
	result := []phenomenonSample{{name: drawing.words("食甚", "greatest"), date: minimum, greatest: true}}
	for _, level := range []struct {
		ingress, egress, beginName, endName string
		penumbra                            bool
		sign                                float64
	}{
		{"P1", "P4", drawing.words("半影始", "penumbra in"), drawing.words("半影终", "penumbra out"), true, 1},
		{"U1", "U4", drawing.words("初亏", "umbra in"), drawing.words("复圆", "umbra out"), false, 1},
		{"U2", "U3", drawing.words("食既", "total in"), drawing.words("生光", "total out"), false, -1},
	} {
		gap := func(date time.Time) (float64, bool) {
			diagram, valid := phenomenonDiagramAt(date, event.Satellite)
			edge := diagram.UmbraEdgeDistance
			if level.penumbra {
				edge = diagram.PenumbraEdgeDistance
			}
			return edge - level.sign*diagram.SatelliteRadiusSunPlane, valid
		}
		value, valid := gap(minimum)
		if !valid {
			return nil, false
		}
		if value >= 0 {
			if level.penumbra {
				return nil, false
			}
			continue
		}
		in, inOK := jovianEclipseRoot(minimum, -1, gap)
		out, outOK := jovianEclipseRoot(minimum, 1, gap)
		if !inOK || !outOK {
			return nil, false
		}
		result = append([]phenomenonSample{{code: level.ingress, name: level.beginName, date: in}}, result...)
		result = append(result, phenomenonSample{code: level.egress, name: level.endName, date: out})
	}
	sort.Slice(result, func(first, second int) bool { return result[first].date.Before(result[second].date) })
	for index := range result {
		diagram, valid := phenomenonDiagramAt(result[index].date, event.Satellite)
		if !valid {
			return nil, false
		}
		result[index].diagram = diagram
		visibility := contactVisibilityAt(result[index].date, event.Satellite)
		visibility.shadow = false
		result[index].visibility = visibility
	}
	return result, true
}

func jovianEclipseMinimum(seed time.Time, satellite int) (time.Time, bool) {
	best, value := seed, math.Inf(1)
	for offset := -360; offset <= 360; offset++ {
		date := seed.Add(time.Duration(offset) * time.Minute)
		diagram, ok := phenomenonDiagramAt(date, satellite)
		if !ok {
			return time.Time{}, false
		}
		if diagram.UmbraEdgeDistance < value {
			best, value = date, diagram.UmbraEdgeDistance
		}
	}
	left, right := best.Add(-time.Minute), best.Add(time.Minute)
	for iteration := 0; iteration < 42; iteration++ {
		first := left.Add(right.Sub(left) / 3)
		second := right.Add(-right.Sub(left) / 3)
		firstDiagram, firstOK := phenomenonDiagramAt(first, satellite)
		secondDiagram, secondOK := phenomenonDiagramAt(second, satellite)
		if !firstOK || !secondOK {
			return time.Time{}, false
		}
		if firstDiagram.UmbraEdgeDistance < secondDiagram.UmbraEdgeDistance {
			right = second
		} else {
			left = first
		}
	}
	return left.Add(right.Sub(left) / 2), true
}

func jovianEclipseRoot(minimum time.Time, direction int, gap func(time.Time) (float64, bool)) (time.Time, bool) {
	inside := minimum
	for step := 1; step <= 600; step++ {
		outside := minimum.Add(time.Duration(direction*step) * time.Minute)
		value, ok := gap(outside)
		if !ok {
			return time.Time{}, false
		}
		if value >= 0 {
			for iteration := 0; iteration < 42; iteration++ {
				middle := outside.Add(inside.Sub(outside) / 2)
				value, valid := gap(middle)
				if !valid {
					return time.Time{}, false
				}
				if value >= 0 {
					outside = middle
				} else {
					inside = middle
				}
			}
			return inside, true
		}
		inside = outside
	}
	return time.Time{}, false
}

func phenomenonGreatest(samples []phenomenonSample) phenomenonSample {
	for _, sample := range samples {
		if sample.greatest {
			return sample
		}
	}
	return phenomenonSample{}
}

func phenomenonContactNames(drawing *chart, kind jupiter.GalileanPhenomenonType) []string {
	switch kind {
	case jupiter.GalileanPhenomenonTransit:
		return []string{drawing.words("外切入凌", "ingress begins"), drawing.words("内切入凌", "ingress ends"), drawing.words("凌甚", "mid-transit"), drawing.words("内切出凌", "egress begins"), drawing.words("外切出凌", "egress ends")}
	case jupiter.GalileanPhenomenonShadowTransit:
		return []string{drawing.words("影始入", "shadow ingress begins"), drawing.words("影全入", "shadow ingress ends"), drawing.words("影凌甚", "mid shadow transit"), drawing.words("影始出", "shadow egress begins"), drawing.words("影全出", "shadow egress ends")}
	default:
		return []string{drawing.words("开始被掩", "begins to vanish"), drawing.words("全部被掩", "disappeared"), drawing.words("掩甚", "mid-occultation"), drawing.words("开始重现", "reappearance begins"), drawing.words("完全重现", "fully reappeared")}
	}
}

func phenomenonTitle(drawing *chart, event jupiter.GalileanPhenomenonContactEvent) string {
	name := satelliteName(event.Satellite-1, drawing.options.english)
	switch event.Type {
	case jupiter.GalileanPhenomenonTransit:
		return fmt.Sprintf(drawing.words("%s凌木星", "%s transits Jupiter"), name)
	case jupiter.GalileanPhenomenonShadowTransit:
		return fmt.Sprintf(drawing.words("%s影凌木星", "Shadow of %s transits Jupiter"), name)
	case jupiter.GalileanPhenomenonEclipse:
		return fmt.Sprintf(drawing.words("%s被木影食", "%s eclipsed by Jupiter"), name)
	default:
		return fmt.Sprintf(drawing.words("%s被木星掩", "%s occulted by Jupiter"), name)
	}
}

// GalileanPhenomenonSVG 单星现象预报图：凌木、影凌、木影食、木星掩，按事件类型绘制 / single-satellite phenomenon forecast chart for transit, shadow transit, eclipse and occultation.
func GalileanPhenomenonSVG(event jupiter.GalileanPhenomenonContactEvent, options GalileanPhenomenonSVGOptions) (string, bool) {
	if !phenomenonEventValid(event) {
		return "", false
	}
	settings := normalizeOptions(options.Width, options.Height, options.Title, options.Language, options.Location, options.TrueScale)
	probe := &chart{options: settings}
	samples, ok := phenomenonContacts(probe, event)
	if !ok {
		return "", false
	}
	greatest := phenomenonGreatest(samples).date
	title := settings.title
	if title == "" {
		title = greatest.In(settings.location).Format("2006-01-02") + " " + phenomenonTitle(probe, event)
	}
	drawing := newChart(settings, title, jovianDefs)
	duration := samples[len(samples)-1].date.Sub(samples[0].date)
	headerBottom := drawing.header(title, []string{
		phenomenonSummary(drawing, event, samples),
		drawing.words("甚：", "Mid: ") + drawing.stamp(greatest),
		fmt.Sprintf(drawing.words("历时 %s  木星相位角 %.1f°", "Duration %s  Jupiter phase angle %.1f°"), formatDuration(duration), jupiter.PhaseAngle(greatest)),
	})
	if event.Type == jupiter.GalileanPhenomenonEclipse {
		var hidden []string
		for _, sample := range samples {
			if sample.visibility.hidden() {
				hidden = append(hidden, phenomenonShort(drawing, sample))
			}
		}
		if len(hidden) > 0 {
			drawing.warning(headerBottom+21, fmt.Sprintf(drawing.words("%s 时卫星被木星或前方卫星遮挡，地面不可见", "At %s the satellite is occulted; not observable from Earth"), strings.Join(hidden, drawing.words("、", ", "))))
		}
		drawEclipseByJupiter(drawing, event, samples)
	} else {
		drawEarthPhenomenon(drawing, event, samples)
	}
	drawPhenomenonTable(drawing, event, samples)
	drawing.footer(phenomenonFooter(drawing, event))
	return drawing.finish()
}

func phenomenonSummary(drawing *chart, event jupiter.GalileanPhenomenonContactEvent, samples []phenomenonSample) string {
	greatest := phenomenonGreatest(samples).diagram
	switch event.Type {
	case jupiter.GalileanPhenomenonEclipse:
		return fmt.Sprintf(drawing.words("卫星距木星本影边缘 %.3f 木星半径（负值在影内）", "Satellite to umbra edge %.3f RJ (negative inside)"), greatest.UmbraEdgeDistance)
	case jupiter.GalileanPhenomenonShadowTransit:
		return fmt.Sprintf(drawing.words("影斑距木星中心 %.3f 木星半径", "Shadow %.3f RJ from Jupiter's centre"), math.Hypot(greatest.ShadowX, greatest.ShadowY))
	default:
		return fmt.Sprintf(drawing.words("卫星距木星中心 %.3f 木星半径", "Satellite %.3f RJ from Jupiter's centre"), math.Hypot(greatest.SatelliteX, greatest.SatelliteY))
	}
}

// phenomenonScale 视窗取木星盘与全部接触位置；凌、影凌、掩在地球视平面，食在太阳视平面。
func phenomenonScale(event jupiter.GalileanPhenomenonContactEvent, samples []phenomenonSample) (float64, float64, float64) {
	minX, maxX, minY, maxY := -1.05, 1.05, -1.05, 1.05
	for _, sample := range samples {
		x, y, r := phenomenonPoint(event, sample.diagram)
		minX, maxX = math.Min(minX, x-r), math.Max(maxX, x+r)
		minY, maxY = math.Min(minY, y-r), math.Max(maxY, y+r)
	}
	if event.Type == jupiter.GalileanPhenomenonEclipse {
		greatest := phenomenonGreatest(samples).diagram
		minX, maxX = math.Min(minX, -greatest.PenumbraScale), math.Max(maxX, greatest.PenumbraScale)
	}
	span := math.Max(maxX-minX, maxY-minY) / 2
	return phenomenonRadius / span, (minX + maxX) / 2, (minY + maxY) / 2
}

func phenomenonPoint(event jupiter.GalileanPhenomenonContactEvent, diagram basic.JupiterGalileanPhenomenonDiagram) (float64, float64, float64) {
	switch event.Type {
	case jupiter.GalileanPhenomenonEclipse:
		return diagram.SunX, diagram.SunY, diagram.SatelliteRadiusSunPlane
	case jupiter.GalileanPhenomenonShadowTransit:
		return diagram.ShadowX, diagram.ShadowY, diagram.ShadowUmbraRadius
	default:
		return diagram.SatelliteX, diagram.SatelliteY, diagram.SatelliteRadius
	}
}

type phenomenonView struct {
	scale, offsetX, offsetY float64
	shiftY                  float64
}

// screen 东为屏幕左、北为屏幕上。
func (view phenomenonView) screen(x, y float64) (float64, float64) {
	return phenomenonCenterX - (x-view.offsetX)*view.scale, phenomenonCenterY + view.shiftY - (y-view.offsetY)*view.scale
}

func drawPhenomenonAxes(drawing *chart, view phenomenonView) {
	drawing.labels.Reserve(panelX-12, 140, 900-panelX, 400)
	drawing.labels.Reserve(0, 520, 900, 80)
	top, bottom := phenomenonCenterY+view.shiftY-phenomenonRadius-14, phenomenonCenterY+view.shiftY+phenomenonRadius+26
	drawing.text(phenomenonCenterX, top, 14, "middle", "#111", "700", drawing.words("北", "N"))
	drawing.text(phenomenonCenterX, bottom, 14, "middle", "#111", "700", drawing.words("南", "S"))
	drawing.text(phenomenonCenterX-phenomenonRadius-60, bottom, 13, "start", "#111", "700", drawing.words("← 东", "← E"))
	drawing.text(phenomenonCenterX+phenomenonRadius+60, bottom, 13, "end", "#111", "700", drawing.words("西 →", "W →"))
	if view.shiftY != 0 {
		drawing.labels.ReserveText(phenomenonCenterX, top, 14, drawing.words("北", "N"), "middle")
		drawing.labels.ReserveText(phenomenonCenterX, bottom, 14, drawing.words("南", "S"), "middle")
	}
}

func drawEarthPhenomenon(drawing *chart, event jupiter.GalileanPhenomenonContactEvent, samples []phenomenonSample) {
	scale, offsetX, offsetY := phenomenonScale(event, samples)
	view := phenomenonView{scale: scale, offsetX: offsetX, offsetY: offsetY}
	drawPhenomenonAxes(drawing, view)
	greatest := phenomenonGreatest(samples).diagram
	jx, jy := view.screen(0, 0)
	snapshot := basic.JupiterGalileanDiagram{MinorRadiusRatio: greatest.MinorRadiusRatio, PolePositionAngle: greatest.PolePositionAngle}
	front := event.Type != jupiter.GalileanPhenomenonOccultation
	if front {
		jupiterBody(drawing, snapshot, jx, jy, scale, "phenomenon-jupiter", "phenomenon-limb")
	}
	drawPhenomenonPath(drawing, event, samples, view)
	fill, stroke := satelliteColor(event.Satellite-1), passiveStroke
	shadowClip := ""
	if event.Type == jupiter.GalileanPhenomenonShadowTransit {
		fill, stroke = "#1d1d1d", "#000"
		drawShadowClip(drawing, snapshot, jx, jy, scale, "phenomenon-shadow-limb")
		shadowClip = ` clip-path="url(#phenomenon-shadow-limb)"`
	}
	for _, sample := range samples {
		x, y, r := phenomenonPoint(event, sample.diagram)
		if event.Type == jupiter.GalileanPhenomenonShadowTransit && (sample.code == "1" || sample.code == "4") {
			r = sample.diagram.ShadowPenumbraRadius
		}
		sx, sy := view.screen(x, y)
		id := "phenomenon-contact-" + sample.code
		if sample.greatest {
			fmt.Fprintf(drawing, `<circle id="phenomenon-body" cx="%.3f" cy="%.3f" r="%.3f" fill="%s" stroke="%s" stroke-width="1.2"%s/>`, sx, sy, r*scale, fill, stroke, shadowClip)
			continue
		}
		fmt.Fprintf(drawing, `<circle id="%s" cx="%.3f" cy="%.3f" r="%.3f" fill="none" stroke="%s" stroke-width="0.9" stroke-dasharray="3 2"/>`, id, sx, sy, r*scale, stroke)
	}
	if !front {
		jupiterBody(drawing, snapshot, jx, jy, scale, "phenomenon-jupiter", "phenomenon-limb")
	}
	labelPhenomenonSamples(drawing, event, samples, view)
}

func drawEclipseByJupiter(drawing *chart, event jupiter.GalileanPhenomenonContactEvent, samples []phenomenonSample) {
	scale, offsetX, offsetY := phenomenonScale(event, samples)
	view := phenomenonView{scale: scale, offsetX: offsetX, offsetY: offsetY, shiftY: 30}
	drawing.labels.Reserve(0, 0, canvasWidth, 160)
	drawPhenomenonAxes(drawing, view)
	greatest := phenomenonGreatest(samples).diagram
	cx, cy := view.screen(0, 0)
	minor := greatest.SunMinorRadiusRatio
	fmt.Fprintf(drawing, `<ellipse id="jupiter-penumbra" cx="%.3f" cy="%.3f" rx="%.3f" ry="%.3f" fill="%s" stroke="#bdbdb8" stroke-width="1"/>`, cx, cy, greatest.PenumbraScale*scale, greatest.PenumbraScale*minor*scale, penumbraFill)
	fmt.Fprintf(drawing, `<ellipse id="jupiter-umbra" cx="%.3f" cy="%.3f" rx="%.3f" ry="%.3f" fill="#3a3632" stroke="#111" stroke-width="1"/>`, cx, cy, greatest.UmbraScale*scale, greatest.UmbraScale*minor*scale)
	drawPhenomenonPath(drawing, event, samples, view)
	for _, sample := range samples {
		x, y, r := phenomenonPoint(event, sample.diagram)
		sx, sy := view.screen(x, y)
		stroke := passiveStroke
		fill := satelliteColor(event.Satellite - 1)
		if sample.visibility.hidden() {
			stroke, fill = "#9a9a9a", "#c5c5c5"
		}
		if sample.greatest {
			fmt.Fprintf(drawing, `<circle id="phenomenon-body" cx="%.3f" cy="%.3f" r="%.3f" fill="%s" stroke="%s" stroke-width="1.2" opacity="0.9"/>`, sx, sy, r*scale, fill, stroke)
			continue
		}
		fmt.Fprintf(drawing, `<circle id="phenomenon-contact-%s" cx="%.3f" cy="%.3f" r="%.3f" fill="%s" fill-opacity="0.6" stroke="%s" stroke-width="0.9" stroke-dasharray="3 2"/>`, sample.code, sx, sy, r*scale, dayFill, stroke)
	}
	labelPhenomenonSamples(drawing, event, samples, view)
	rows := []struct{ fill, stroke, label string }{
		{penumbraFill, "#bdbdb8", drawing.words("木星半影", "Jupiter penumbra")},
		{"#3a3632", "#111", drawing.words("木星本影", "Jupiter umbra")},
		{satelliteColor(event.Satellite - 1), passiveStroke, satelliteName(event.Satellite-1, drawing.options.english)},
	}
	for index, row := range rows {
		fmt.Fprintf(drawing, `<rect x="%.3f" y="%.3f" width="12" height="9" fill="%s" stroke="%s" stroke-width="0.8"/>`, panelX, 386+float64(index)*18-8, row.fill, row.stroke)
		drawing.text(panelX+18, 386+float64(index)*18, 11, "start", "#333", "", row.label)
	}
}

func labelPhenomenonSamples(drawing *chart, event jupiter.GalileanPhenomenonContactEvent, samples []phenomenonSample, view phenomenonView) {
	var columns [2][]callout
	for index, sample := range samples {
		x, y, r := phenomenonPoint(event, sample.diagram)
		sx, sy := view.screen(x, y)
		color := "#2554c7"
		if sample.visibility.hidden() {
			color = "#9a9a9a"
		}
		fmt.Fprintf(drawing, `<circle id="phenomenon-position-%d" cx="%.3f" cy="%.3f" r="2.2" fill="%s"/>`, index, sx, sy, color)
		drawing.labels.Reserve(sx-r*view.scale, sy-r*view.scale, 2*r*view.scale, 2*r*view.scale)
		label := phenomenonShort(drawing, sample) + " " + drawing.clock(sample.date)
		if sample.visibility.hidden() {
			label += drawing.words("（不可见）", " (hidden)")
		}
		side := 0
		if sx > phenomenonCenterX || (sx == phenomenonCenterX && index >= 2) {
			side = 1
		}
		columns[side] = append(columns[side], callout{index, sx, sy, float64(2*side - 1), label, color})
	}
	for side, column := range columns {
		if len(column) == 0 {
			continue
		}
		if len(column) > 1 && column[0].y > column[len(column)-1].y {
			for left, right := 0, len(column)-1; left < right; left, right = left+1, right-1 {
				column[left], column[right] = column[right], column[left]
			}
		}
		top, bottom := phenomenonCenterY+view.shiftY-phenomenonRadius+8, phenomenonCenterY+view.shiftY+phenomenonRadius-8
		gap := math.Min(22, (bottom-top)/math.Max(1, float64(len(column)-1)))
		start := math.Max(top, math.Min(bottom-gap*float64(len(column)-1), averageY(column)-gap*float64(len(column)-1)/2))
		labelX, anchor := phenomenonCenterX-phenomenonRadius-28, "end"
		if side == 1 {
			labelX, anchor = phenomenonCenterX+phenomenonRadius+28, "start"
		}
		for order, item := range column {
			baseline := start + gap*float64(order)
			placedX := labelX
			if event.Type == jupiter.GalileanPhenomenonEclipse {
				box := svgchart.LabelTextBox(placedX+3*item.side, baseline, 11, item.label, anchor)
				if box.X < 38 {
					placedX += 38 - box.X
				}
				if box.X+box.Width > panelX-12 {
					placedX -= box.X + box.Width - panelX + 12
				}
			}
			fmt.Fprintf(drawing, `<path class="callout" d="M %.3f %.3f L %.3f %.3f L %.3f %.3f" fill="none" stroke="%s" stroke-width="0.6" opacity="0.75"/>`, item.x, item.y, placedX-10*item.side, baseline-4, placedX, baseline-4, item.color)
			drawing.text(placedX+3*item.side, baseline, 11, anchor, item.color, "", item.label)
			if event.Type == jupiter.GalileanPhenomenonEclipse {
				drawing.labels.ReserveText(placedX+3*item.side, baseline, 11, item.label, anchor)
			}
		}
	}
	name := satelliteName(event.Satellite-1, drawing.options.english)
	if event.Type == jupiter.GalileanPhenomenonShadowTransit {
		name = fmt.Sprintf(drawing.words("%s的影子", "shadow of %s"), name)
	}
	x, y, r := phenomenonPoint(event, phenomenonGreatest(samples).diagram)
	sx, sy := view.screen(x, y)
	labelDisk(drawing, sx, sy, r*view.scale+2, 0, -1, passiveStroke, name)
}

func phenomenonShort(drawing *chart, sample phenomenonSample) string {
	if sample.code == "" {
		return drawing.words("甚", "Mid")
	}
	if strings.HasPrefix(sample.code, "P") || strings.HasPrefix(sample.code, "U") {
		return sample.code
	}
	return drawing.words([]string{"一触", "二触", "", "三触", "四触"}[map[string]int{"1": 0, "2": 1, "3": 3, "4": 4}[sample.code]], "C"+sample.code)
}

func drawPhenomenonTable(drawing *chart, event jupiter.GalileanPhenomenonContactEvent, samples []phenomenonSample) {
	y := 170.0
	drawing.text(panelX, y, 13, "start", "#111", "700", fmt.Sprintf(drawing.words("接触时刻 (%s)", "Contacts (%s)"), zoneLabel(event.Greatest.In(drawing.options.location))))
	for index, sample := range samples {
		baseline := y + 20 + float64(index)*18
		label := phenomenonShort(drawing, sample) + " " + sample.name
		color := "#222"
		if sample.visibility.hidden() {
			color = "#9a9a9a"
			drawing.text(panelX+256, baseline, 12, "start", "#b21f16", "", "✕")
		}
		drawing.text(panelX, baseline, 12, "start", color, "", svgchart.EllipsizeText(label, 170, 12))
		drawing.text(panelX+250, baseline, 12, "end", color, "", drawing.clock(sample.date))
	}
	y += 20 + float64(len(samples))*18 + 12
	state := event.GreatestState
	var notes []string
	for _, sample := range samples {
		if sample.visibility.hidden() {
			notes = append(notes, drawing.words("✕ 该时刻被遮挡，地面看不到", "✕ occulted, not observable"))
			break
		}
	}
	if event.Type != jupiter.GalileanPhenomenonEclipse && state.Eclipse {
		notes = append(notes, drawing.words("甚时卫星在木影中", "satellite in Jupiter's shadow at mid"))
	}
	if event.Type == jupiter.GalileanPhenomenonTransit && state.ShadowTransit {
		notes = append(notes, drawing.words("同时有影凌", "its shadow is also on the disk"))
	}
	if event.Type == jupiter.GalileanPhenomenonShadowTransit && state.Transit {
		notes = append(notes, drawing.words("卫星本身也在凌木", "the satellite is also transiting"))
	}
	for index, note := range notes {
		drawing.text(panelX, y+float64(index)*18, 12, "start", "#b21f16", "", note)
	}
}

func phenomenonFooter(drawing *chart, event jupiter.GalileanPhenomenonContactEvent) []string {
	plane := drawing.words("地心视向，上北下南、左东右西；虚线为相对路径，虚线圆为各接触时刻位置。", "Geocentric view, N up, E left; dashed line is the path, dashed circles mark contacts.")
	if event.Type == jupiter.GalileanPhenomenonEclipse {
		plane = drawing.words("太阳视平面（从太阳看向木星），上北下南、左东右西；卫星穿过木星影锥截面。", "Sun view toward Jupiter, N up, E left; the satellite crosses Jupiter's shadow.")
	}
	contacts := drawing.words("一触/四触为有限圆盘外切，二触/三触为内切。", "C1/C4 external, C2/C3 internal contacts.")
	if event.Type == jupiter.GalileanPhenomenonEclipse {
		contacts = drawing.words("P1/P4：半影外切；U1/U4：本影外切；U2/U3：本影内切，掠食无此二项。", "P1/P4: penumbral outer contacts; U1/U4: umbral outer; U2/U3: inner, absent for grazing events.")
	}
	return []string{plane, contacts + " " + drawing.timeNote(event.Greatest)}
}

func drawPhenomenonPath(drawing *chart, event jupiter.GalileanPhenomenonContactEvent, samples []phenomenonSample, view phenomenonView) {
	path := make([]basic.JupiterGalileanPhenomenonDiagram, 0, 41)
	span := samples[len(samples)-1].date.Sub(samples[0].date)
	for index := -4; index <= 44; index++ {
		date := samples[0].date.Add(time.Duration(float64(span) * float64(index) / 40))
		if diagram, ok := phenomenonDiagramAt(date, event.Satellite); ok {
			path = append(path, diagram)
		}
	}
	drawing.WriteString(`<path id="phenomenon-path" d="`)
	for index, diagram := range path {
		x, y, _ := phenomenonPoint(event, diagram)
		sx, sy := view.screen(x, y)
		command := "L"
		if index == 0 {
			command = "M"
		}
		fmt.Fprintf(drawing, "%s %.3f %.3f ", command, sx, sy)
	}
	drawing.WriteString(`" fill="none" stroke="#333" stroke-width="1" stroke-dasharray="5 4"/>`)
}
