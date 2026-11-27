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
	planeCenterX = 300.0
	panelX       = 594.0
)

// planeLayout 视平面让出下方给地球视点相位条。
type planeLayout struct {
	centerY, radius float64
}

func layoutFor(event jupiter.GalileanMutualEvent) planeLayout {
	return planeLayout{centerY: 286, radius: 118}
}

type mutualSample struct {
	date     time.Time
	geometry basic.JupiterGalileanMutualDiagram
}

type mutualTrack struct {
	contacts [3]mutualSample
	path     []basic.JupiterGalileanMutualDiagram
	scale    float64
	active   float64
	eclipse  bool
	layout   planeLayout
}

func mutualEventValid(event jupiter.GalileanMutualEvent) bool {
	if !event.Valid || event.Active < 1 || event.Active > 4 || event.Passive < 1 || event.Passive > 4 || event.Active == event.Passive {
		return false
	}
	if event.Type != jupiter.GalileanMutualOccultation && event.Type != jupiter.GalileanMutualEclipse {
		return false
	}
	if event.Start.IsZero() || event.End.IsZero() || event.Greatest.Before(event.Start) || event.Greatest.After(event.End) || !event.End.After(event.Start) {
		return false
	}
	for _, value := range []float64{event.MinimumSeparationArcsec, event.ImpactParameter, event.Coverage} {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			return false
		}
	}
	return event.Coverage <= 1 && event.Duration > 0
}

func mutualDiagramAt(date time.Time, event jupiter.GalileanMutualEvent) (basic.JupiterGalileanMutualDiagram, bool) {
	geometry, ok := basic.JupiterGalileanMutualDiagramWithInstance(basic.UTC2TT(basic.Date2JD(date.UTC())), event.Active, event.Passive, basic.JupiterGalileanMutualPhenomenonType(event.Type), nil)
	return geometry, ok && geometry.Active == event.Active && geometry.Passive == event.Passive
}

func activeRadius(event jupiter.GalileanMutualEvent, geometry basic.JupiterGalileanMutualDiagram) float64 {
	if event.Type == jupiter.GalileanMutualEclipse {
		return geometry.PenumbraRadiusJupiterR
	}
	return geometry.ActiveRadiusJupiterR
}

// buildMutualTrack 互掩以被遮星为图心，互食以影锥轴为图心；快照给的是被遮星相对遮挡星的东、北偏移。
func buildMutualTrack(event jupiter.GalileanMutualEvent) (mutualTrack, bool) {
	var track mutualTrack
	extent := 0.0
	for index, date := range []time.Time{event.Start, event.Greatest, event.End} {
		geometry, ok := mutualDiagramAt(date, event)
		if !ok {
			return mutualTrack{}, false
		}
		track.contacts[index] = mutualSample{date, geometry}
	}
	span := event.End.Sub(event.Start)
	for index := -6; index <= 30; index++ {
		date := event.Start.Add(time.Duration(float64(span) * float64(index) / 24))
		geometry, ok := mutualDiagramAt(date, event)
		if !ok {
			return mutualTrack{}, false
		}
		track.path = append(track.path, geometry)
		if index >= 0 && index <= 24 {
			moving := activeRadius(event, geometry)
			if event.Type == jupiter.GalileanMutualEclipse {
				moving = geometry.PassiveRadiusJupiterR
			}
			extent = math.Max(extent, math.Hypot(geometry.XJupiterR, geometry.YJupiterR)+moving)
		}
	}
	greatest := track.contacts[1].geometry
	extent = math.Max(extent, greatest.PassiveRadiusJupiterR)
	if event.Type == jupiter.GalileanMutualEclipse {
		extent = math.Max(extent, greatest.PenumbraRadiusJupiterR)
		track.eclipse = true
	}
	track.layout = layoutFor(event)
	track.scale = track.layout.radius / extent
	track.active = activeRadius(event, greatest)
	return track, true
}

// point 互掩返回遮挡星（被遮星居中），互食返回被食星（影锥轴居中）；均为东左北上。
func (track mutualTrack) point(geometry basic.JupiterGalileanMutualDiagram) (float64, float64) {
	if track.eclipse {
		return planeCenterX - geometry.XJupiterR*track.scale, track.layout.centerY - geometry.YJupiterR*track.scale
	}
	return planeCenterX + geometry.XJupiterR*track.scale, track.layout.centerY + geometry.YJupiterR*track.scale
}

// GalileanMutualEventSVG 互现象预报图：互掩为地球视平面，互食为太阳视平面，被遮星居中 / mutual-event forecast chart centred on the passive satellite, Earth plane for occultations and Sun plane for eclipses.
func GalileanMutualEventSVG(event jupiter.GalileanMutualEvent, options GalileanMutualEventSVGOptions) (string, bool) {
	if !mutualEventValid(event) {
		return "", false
	}
	settings := normalizeOptions(options.Width, options.Height, options.Title, options.Language, options.Location, options.TrueScale)
	track, ok := buildMutualTrack(event)
	if !ok {
		return "", false
	}
	probe := &chart{options: settings}
	title := settings.title
	if title == "" {
		title = event.Greatest.In(settings.location).Format("2006-01-02") + " " + mutualTitle(probe, event)
	}
	drawing := newChart(settings, title, "")
	headerBottom := drawing.header(title, mutualHeaderLines(drawing, event))
	contacts := mutualContacts(drawing, event, track)
	visibility := mutualVisibility(event, contacts)
	warning := mutualWarning(drawing, event, contacts, visibility)
	if warning != "" {
		drawing.warning(headerBottom+21, warning)
		drawing.labels.ReserveText(450, headerBottom+21, 13, warning, "middle")
		oldRadius := track.layout.radius
		track.layout.radius -= 8
		track.layout.centerY += 10
		track.scale *= track.layout.radius / oldRadius
	}
	drawMutualPlane(drawing, event, track, contacts, visibility)
	drawMutualContacts(drawing, event, contacts, visibility, track.scale, warning != "")
	if finder, ok := mutualFinder(event); ok {
		drawMutualFinder(drawing, event, finder)
	}
	if event.Type == jupiter.GalileanMutualEclipse {
		samples, ok := earthShadowSamples(event, contacts)
		if !ok {
			return "", false
		}
		drawEarthPhaseStrip(drawing, event, samples, visibility)
	} else {
		drawOccultationStrip(drawing, event, contacts, visibility)
	}
	drawing.footerWithin(mutualFooter(drawing, event), finderLeft-50)
	return drawing.finish()
}

func mutualTitle(drawing *chart, event jupiter.GalileanMutualEvent) string {
	active := satelliteName(event.Active-1, drawing.options.english)
	passive := satelliteName(event.Passive-1, drawing.options.english)
	kind := mutualClass(drawing, event)
	if drawing.options.english {
		verb := "occults"
		if event.Type == jupiter.GalileanMutualEclipse {
			verb = "eclipses"
		}
		return fmt.Sprintf("%s %s %s (%s)", active, verb, passive, kind)
	}
	verb := "掩"
	if event.Type == jupiter.GalileanMutualEclipse {
		verb = "食"
	}
	return fmt.Sprintf("%s%s%s（%s）", active, verb, passive, kind)
}

// mutualClass 按覆盖与内含给出全、环、偏、半影四类；互食无本影覆盖时为半影食。
func mutualClass(drawing *chart, event jupiter.GalileanMutualEvent) string {
	eclipse := event.Type == jupiter.GalileanMutualEclipse
	switch {
	case event.Annular:
		return drawing.words(map[bool]string{false: "环掩", true: "环食"}[eclipse], "annular")
	case event.Coverage >= 1:
		return drawing.words(map[bool]string{false: "全掩", true: "全食"}[eclipse], "total")
	case eclipse && event.Coverage <= 0:
		return drawing.words("半影食", "penumbral")
	default:
		return drawing.words(map[bool]string{false: "偏掩", true: "偏食"}[eclipse], "partial")
	}
}

func mutualVisible(event jupiter.GalileanMutualEvent) bool {
	passive := event.GreatestState[event.Passive-1]
	return !passive.Eclipse && !passive.Occultation
}

func mutualHeaderLines(drawing *chart, event jupiter.GalileanMutualEvent) []string {
	greatestName := drawing.words("掩甚", "Greatest")
	if event.Type == jupiter.GalileanMutualEclipse {
		greatestName = drawing.words("食甚", "Greatest")
	}
	return []string{
		fmt.Sprintf(drawing.words("几何覆盖率 %.1f%%  碰撞参数 b=%.4f  最小角距 %.4f″", "Coverage %.1f%%  impact b=%.4f  minimum separation %.4f″"), 100*event.Coverage, event.ImpactParameter, event.MinimumSeparationArcsec),
		greatestName + drawing.words("：", ": ") + drawing.stamp(event.Greatest),
		fmt.Sprintf(drawing.words("历时 %s  木星相位角 %.1f°", "Duration %s  Jupiter phase angle %.1f°"), formatDuration(event.Duration), jupiter.PhaseAngle(event.Greatest)),
	}
}

func drawMutualPlane(drawing *chart, event jupiter.GalileanMutualEvent, track mutualTrack, contacts []mutualContact, visibility []contactVisibility) {
	planeCenterY, planeRadius := track.layout.centerY, track.layout.radius
	eclipse := event.Type == jupiter.GalileanMutualEclipse
	if eclipse {
		drawEclipsePlane(drawing, event, track, contacts, visibility)
		return
	}
	drawing.labels.Reserve(panelX-12, 140, 900-panelX, 400)
	drawing.labels.Reserve(0, 540, 900, 60)
	fmt.Fprintf(drawing, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#8b8b86" stroke-width="0.8" stroke-dasharray="3 4"/>`, planeCenterX, planeCenterY-planeRadius-6, planeCenterX, planeCenterY+planeRadius+6)
	fmt.Fprintf(drawing, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#8b8b86" stroke-width="0.8" stroke-dasharray="3 4"/>`, planeCenterX-planeRadius-6, planeCenterY, planeCenterX+planeRadius+6, planeCenterY)
	drawing.compass(planeCenterX, planeCenterY, planeCenterY-planeRadius-14, planeCenterY+planeRadius+26, planeCenterX-planeRadius+2, planeCenterX+planeRadius-2)

	greatest := track.contacts[1].geometry
	passiveRadius := greatest.PassiveRadiusJupiterR * track.scale
	activeX, activeY := track.point(greatest)
	clipID := "mutual-passive-clip"
	fmt.Fprintf(drawing, `<clipPath id="%s"><circle cx="%.3f" cy="%.3f" r="%.3f"/></clipPath>`, clipID, planeCenterX, planeCenterY, passiveRadius)

	if eclipse {
		fmt.Fprintf(drawing, `<circle id="mutual-penumbra" cx="%.3f" cy="%.3f" r="%.3f" fill="#e9e9e9" stroke="#d6d6d6" stroke-width="1.2" opacity="0.85"/>`, activeX, activeY, greatest.PenumbraRadiusJupiterR*track.scale)
	}
	fmt.Fprintf(drawing, `<circle id="mutual-passive" cx="%.3f" cy="%.3f" r="%.3f" fill="%s" stroke="%s" stroke-width="1.6"/>`, planeCenterX, planeCenterY, passiveRadius, dayFill, passiveStroke)
	if eclipse {
		fmt.Fprintf(drawing, `<circle cx="%.3f" cy="%.3f" r="%.3f" fill="#9a9a9a" opacity="0.35" clip-path="url(#%s)"/>`, activeX, activeY, greatest.PenumbraRadiusJupiterR*track.scale, clipID)
		umbra := math.Max(0, greatest.UmbraRadiusJupiterR) * track.scale
		fmt.Fprintf(drawing, `<circle id="mutual-covered" cx="%.3f" cy="%.3f" r="%.3f" fill="#b21f16" opacity="0.78" clip-path="url(#%s)"/>`, activeX, activeY, umbra, clipID)
		fmt.Fprintf(drawing, `<circle cx="%.3f" cy="%.3f" r="%.3f" fill="none" stroke="#7a705f" stroke-width="1.2"/>`, planeCenterX, planeCenterY, passiveRadius)
		fmt.Fprintf(drawing, `<circle id="mutual-umbra" cx="%.3f" cy="%.3f" r="%.3f" fill="none" stroke="#83170f" stroke-width="1.2" stroke-dasharray="4 3"/>`, activeX, activeY, umbra)
	}

	if !eclipse {
		activeRadius := greatest.ActiveRadiusJupiterR * track.scale
		fmt.Fprintf(drawing, `<clipPath id="mutual-active-clip"><circle cx="%.3f" cy="%.3f" r="%.3f"/></clipPath>`, activeX, activeY, activeRadius)
		fmt.Fprintf(drawing, `<circle id="mutual-active" cx="%.3f" cy="%.3f" r="%.3f" fill="%s" fill-opacity="0.55" stroke="none"/>`, activeX, activeY, activeRadius, activeFill)
		fmt.Fprintf(drawing, `<circle id="mutual-covered" cx="%.3f" cy="%.3f" r="%.3f" fill="#b21f16" opacity="0.55" clip-path="url(#mutual-active-clip)"/>`, planeCenterX, planeCenterY, passiveRadius)
		fmt.Fprintf(drawing, `<circle cx="%.3f" cy="%.3f" r="%.3f" fill="none" stroke="%s" stroke-width="1" stroke-dasharray="3 3" clip-path="url(#mutual-active-clip)"/>`, planeCenterX, planeCenterY, passiveRadius, passiveStroke)
		fmt.Fprintf(drawing, `<circle class="mutual-active-outline" cx="%.3f" cy="%.3f" r="%.3f" fill="none" stroke="%s" stroke-width="1.8"/>`, activeX, activeY, activeRadius, activeStroke)
	}
	drawing.WriteString(`<path id="mutual-path" d="`)
	for index, geometry := range track.path {
		x, y := track.point(geometry)
		command := "L"
		if index == 0 {
			command = "M"
		}
		fmt.Fprintf(drawing, "%s %.3f %.3f ", command, x, y)
	}
	drawing.WriteString(`" fill="none" stroke="#333333" stroke-width="1.1" stroke-dasharray="5 4" stroke-linecap="round"/>`)

	for index, contact := range contacts {
		if contact.greatest {
			continue
		}
		x, y := track.point(contact.geometry)
		fmt.Fprintf(drawing, `<circle id="mutual-contact-%s" cx="%.3f" cy="%.3f" r="%.3f" fill="none" stroke="%s" stroke-width="0.9" stroke-dasharray="2 3" opacity="0.8"/>`, contactID(contact, index), x, y, contact.geometry.ActiveRadiusJupiterR*track.scale, activeStroke)
	}
	drawMutualLabels(drawing, event, track, contacts, visibility)
}

func drawPlaneAxes(drawing *chart, layout planeLayout) {
	planeCenterY, planeRadius := layout.centerY, layout.radius
	drawing.labels.Reserve(panelX-12, 140, 900-panelX, 400)
	drawing.labels.Reserve(0, 540, 900, 60)
	fmt.Fprintf(drawing, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#8b8b86" stroke-width="0.8" stroke-dasharray="3 4"/>`, planeCenterX, planeCenterY-planeRadius-6, planeCenterX, planeCenterY+planeRadius+6)
	fmt.Fprintf(drawing, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#8b8b86" stroke-width="0.8" stroke-dasharray="3 4"/>`, planeCenterX-planeRadius-6, planeCenterY, planeCenterX+planeRadius+6, planeCenterY)
	drawing.compass(planeCenterX, planeCenterY, planeCenterY-planeRadius-14, planeCenterY+planeRadius+26, planeCenterX-planeRadius+2, planeCenterX+planeRadius-2)
}

// drawEclipsePlane 仿月食穿影图：半影、本影截面固定在图心，被食星沿相对路径穿过。
func drawEclipsePlane(drawing *chart, event jupiter.GalileanMutualEvent, track mutualTrack, contacts []mutualContact, visibility []contactVisibility) {
	planeCenterY := track.layout.centerY
	drawPlaneAxes(drawing, track.layout)
	greatest := track.contacts[1].geometry
	penumbra := greatest.PenumbraRadiusJupiterR * track.scale
	passiveRadius := greatest.PassiveRadiusJupiterR * track.scale
	fmt.Fprintf(drawing, `<circle id="mutual-penumbra" cx="%.3f" cy="%.3f" r="%.3f" fill="%s" stroke="#bdbdb8" stroke-width="1.2"/>`, planeCenterX, planeCenterY, penumbra, penumbraFill)
	core := coreShadow(greatest)
	coreRadius := core.radius * track.scale
	drawing.WriteString(`<path id="mutual-path" d="`)
	for index, geometry := range track.path {
		x, y := track.point(geometry)
		command := "L"
		if index == 0 {
			command = "M"
		}
		fmt.Fprintf(drawing, "%s %.3f %.3f ", command, x, y)
	}
	drawing.WriteString(`" fill="none" stroke="#333333" stroke-width="1.1" stroke-dasharray="5 4" stroke-linecap="round"/>`)
	order := make([]int, 0, len(contacts))
	for index, contact := range contacts {
		if !contact.greatest {
			order = append(order, index)
		}
	}
	for index, contact := range contacts {
		if contact.greatest {
			order = append(order, index)
		}
	}
	for _, index := range order {
		contact := contacts[index]
		x, y := track.point(contact.geometry)
		radius := contact.geometry.PassiveRadiusJupiterR * track.scale
		id := "mutual-contact-" + contactID(contact, index)
		if !contact.greatest {
			fmt.Fprintf(drawing, `<circle id="%s" cx="%.3f" cy="%.3f" r="%.3f" fill="none" stroke="#7a705f" stroke-width="0.8" stroke-dasharray="3 2" opacity="0.8"/>`, id, x, y, radius)
			continue
		}
		id = "mutual-passive"
		clip := "eclipse-clip-" + contactID(contact, index)
		fmt.Fprintf(drawing, `<clipPath id="%s"><circle cx="%.3f" cy="%.3f" r="%.3f"/></clipPath>`, clip, x, y, radius)
		fmt.Fprintf(drawing, `<circle id="%s" cx="%.3f" cy="%.3f" r="%.3f" fill="%s" stroke="#7a705f" stroke-width="0.9"/>`, id, x, y, radius, dayFill)
		fmt.Fprintf(drawing, `<circle class="passive-penumbra" cx="%.3f" cy="%.3f" r="%.3f" fill="%s" opacity="%.2f" clip-path="url(#%s)"/>`, planeCenterX, planeCenterY, penumbra, penumbraOverlay, penumbraOpacity, clip)
	}
	if core.kind != coreNone {
		visible := math.Max(coreRadius, minimumCoreRadius)
		fmt.Fprintf(drawing, `<circle id="mutual-%s" cx="%.3f" cy="%.3f" r="%.3f" fill="%s" stroke="%s" stroke-width="1"/>`, core.id(), planeCenterX, planeCenterY, visible, core.fill(), core.stroke())
	}
	for _, contact := range contacts {
		x, y := track.point(contact.geometry)
		r := contact.geometry.PassiveRadiusJupiterR * track.scale
		drawing.labels.Reserve(x-r, y-r, 2*r, 2*r)
	}
	drawing.labels.Reserve(planeCenterX-penumbra, planeCenterY-penumbra, 2*penumbra, 2*penumbra)
	drawContactLabels(drawing, event, track, contacts, visibility, passiveRadius, penumbra)
}

func drawEclipseLegend(drawing *chart, event jupiter.GalileanMutualEvent, y float64, core coreShadowInfo, scale float64) {
	activeName := satelliteName(event.Active-1, drawing.options.english)
	rows := []struct{ fill, stroke, label string }{
		{penumbraFill, "#bdbdb8", fmt.Sprintf(drawing.words("%s半影", "%s penumbra"), activeName)},
		{dayFill, "#7a705f", drawing.words("被食星亮面", "lit disk")},
	}
	if core.kind != coreNone {
		label := fmt.Sprintf("%s r=%.0f km", core.name(drawing), core.radius*71492)
		if core.radius*scale < minimumCoreRadius {
			label += drawing.words("（放大）", " (enlarged)")
		}
		rows = append(rows, struct{ fill, stroke, label string }{core.fill(), core.stroke(), label})
	}
	for index, row := range rows {
		fontSize := 11.0
		labelX := math.Min(panelX+150, 870-svgchart.EstimatedTextWidth(row.label, fontSize))
		if labelX < panelX+146 {
			labelX = panelX + 146
			for fontSize > 7.0 && svgchart.EstimatedTextWidth(row.label, fontSize) > 870-labelX {
				fontSize -= 0.5
			}
		}
		fmt.Fprintf(drawing, `<rect x="%.3f" y="%.3f" width="12" height="9" fill="%s" stroke="%s" stroke-width="0.8"/>`, labelX-18, y+float64(index)*18-8, row.fill, row.stroke)
		drawing.text(labelX, y+float64(index)*18, fontSize, "start", "#333", "", row.label)
	}
}

func drawMutualLabels(drawing *chart, event jupiter.GalileanMutualEvent, track mutualTrack, contacts []mutualContact, visibility []contactVisibility) {
	planeCenterY := track.layout.centerY
	greatest := track.contacts[1].geometry
	passiveRadius := greatest.PassiveRadiusJupiterR * track.scale
	activeX, activeY := track.point(greatest)
	activeRadius := greatest.ActiveRadiusJupiterR * track.scale
	drawing.labels.Reserve(planeCenterX-passiveRadius, planeCenterY-passiveRadius, 2*passiveRadius, 2*passiveRadius)
	drawing.labels.Reserve(activeX-activeRadius, activeY-activeRadius, 2*activeRadius, 2*activeRadius)
	reach := math.Max(passiveRadius, math.Hypot(activeX-planeCenterX, activeY-planeCenterY)+activeRadius)
	drawContactLabels(drawing, event, track, contacts, visibility, activeRadius, reach)
}

// drawContactLabels 各接触点画蓝点、标“代号 时刻”，标签沿路径法向外推，先放在离盘面较远的一侧。
func drawContactLabels(drawing *chart, event jupiter.GalileanMutualEvent, track mutualTrack, contacts []mutualContact, visibility []contactVisibility, movingRadius, reach float64) {
	planeCenterY := track.layout.centerY
	firstX, firstY := track.point(contacts[0].geometry)
	lastX, lastY := track.point(contacts[len(contacts)-1].geometry)
	length := math.Hypot(lastX-firstX, lastY-firstY)
	normalX, normalY := 0.0, -1.0
	if length > 0 {
		normalX, normalY = -(lastY-firstY)/length, (lastX-firstX)/length
	}
	if normalY > 0 {
		normalX, normalY = -normalX, -normalY
	}
	for index, contact := range contacts {
		cx, cy := track.point(contact.geometry)
		color := "#2554c7"
		if visibility[index].hidden() {
			color = "#9a9a9a"
		}
		fmt.Fprintf(drawing, `<circle id="mutual-position-%s" cx="%.3f" cy="%.3f" r="2.4" fill="%s"/>`, contactID(contact, index), cx, cy, color)
		drawing.labels.Reserve(cx-3, cy-3, 6, 6)
	}
	drawContactCallouts(drawing, track, contacts, visibility)
	passive := track.contacts[1].geometry
	name := satelliteName(event.Passive-1, drawing.options.english)
	if event.Type == jupiter.GalileanMutualEclipse {
		gx, gy := track.point(passive)
		drawing.placeText(13, "#111", name, svgchart.LabelCandidates(gx, gy+4, "middle", passive.PassiveRadiusJupiterR*track.scale+20))
		return
	}
	ax, ay := track.point(passive)
	passiveRadius := passive.PassiveRadiusJupiterR * track.scale
	activeRadius := passive.ActiveRadiusJupiterR * track.scale
	dx, dy := ax-planeCenterX, ay-planeCenterY
	if length := math.Hypot(dx, dy); length > 0 {
		dx, dy = dx/length, dy/length
	} else {
		dx, dy = 0, -1
	}
	labelDisk(drawing, planeCenterX, planeCenterY, passiveRadius, -dx, -dy, passiveStroke, name+drawing.words("（后，被掩）", " (behind)"))
	labelDisk(drawing, ax, ay, activeRadius, dx, dy, activeStroke, satelliteName(event.Active-1, drawing.options.english)+drawing.words("（前，遮挡）", " (front)"))
}

type callout struct {
	index        int
	x, y         float64
	side         float64
	label, color string
}

// drawContactCallouts 接触点标签排成左右两列（西侧点归右列），按点的高度排序后自上而下等距排开，细线连回各点，慢事件点挤在一起时也不重叠。
func drawContactCallouts(drawing *chart, track mutualTrack, contacts []mutualContact, visibility []contactVisibility) {
	var columns [2][]callout
	for index, contact := range contacts {
		x, y := track.point(contact.geometry)
		label := contactShort(drawing, contact) + " " + drawing.clock(contact.date)
		color := "#2554c7"
		if visibility[index].hidden() {
			color = "#9a9a9a"
			label += drawing.words("（不可见）", " (hidden)")
		}
		side := 0
		if x > planeCenterX || (x == planeCenterX && index >= len(contacts)/2) {
			side = 1
		}
		columns[side] = append(columns[side], callout{index, x, y, float64(2*side - 1), label, color})
	}
	top, bottom := track.layout.centerY-track.layout.radius+8, track.layout.centerY+track.layout.radius-8
	for side, column := range columns {
		if len(column) == 0 {
			continue
		}
		// 同侧点按时间先后自上而下排列，避免引线交叉时也保持读序。
		sort.SliceStable(column, func(i, j int) bool { return column[i].index < column[j].index })
		if len(column) > 1 && column[0].y > column[len(column)-1].y {
			for left, right := 0, len(column)-1; left < right; left, right = left+1, right-1 {
				column[left], column[right] = column[right], column[left]
			}
		}
		gap := math.Min(20, (bottom-top)/math.Max(1, float64(len(column)-1)))
		start := math.Max(top, math.Min(bottom-gap*float64(len(column)-1), averageY(column)-gap*float64(len(column)-1)/2))
		labelX, anchor := planeCenterX-track.layout.radius-24, "end"
		if side == 1 {
			labelX, anchor = planeCenterX+track.layout.radius+24, "start"
		}
		for order, item := range column {
			baseline := start + gap*float64(order)
			elbow := labelX - 10*item.side
			fmt.Fprintf(drawing, `<path class="callout" d="M %.3f %.3f L %.3f %.3f L %.3f %.3f" fill="none" stroke="%s" stroke-width="0.6" opacity="0.75"/>`, item.x, item.y, elbow, baseline-4, labelX, baseline-4, item.color)
			drawing.text(labelX+3*item.side, baseline, 11, anchor, item.color, "", item.label)
			drawing.labels.ReserveText(labelX+3*item.side, baseline, 11, item.label, anchor)
		}
	}
}

func averageY(items []callout) float64 {
	total := 0.0
	for _, item := range items {
		total += item.y
	}
	return total / float64(len(items))
}

func contactID(contact mutualContact, index int) string {
	if contact.code != "" {
		return contact.code
	}
	return fmt.Sprintf("g%d", index)
}

// contactShort 图上只标代号，甚无代号时用“甚 / G”。
func contactShort(drawing *chart, contact mutualContact) string {
	if contact.code != "" {
		return contact.code
	}
	return drawing.words("甚", "G")
}

func contactLabel(contact mutualContact) string {
	if contact.code == "" {
		return contact.name
	}
	return contact.code + " " + contact.name
}

func drawMutualContacts(drawing *chart, event jupiter.GalileanMutualEvent, contacts []mutualContact, visibility []contactVisibility, scale float64, hasWarning bool) {
	var earthLines []string
	if event.Type == jupiter.GalileanMutualEclipse {
		jde := basic.UTC2TT(basic.Date2JD(event.Greatest.UTC()))
		if shadow, ok := basic.JupiterGalileanEarthShadowWithInstance(jde, event.Active, event.Passive, nil); ok {
			earthLines = earthCoverageLines(drawing, shadow)
		}
	}
	y := 160.0
	rowStep, detailGap, detailStep := 18.0, 10.0, 18.0
	if hasWarning {
		y = 170.0
		if len(contacts) > 5 {
			rowStep, detailGap, detailStep = 16.0, 8.0, 16.0
		} else {
			rowStep, detailGap, detailStep = 17.0, 8.0, 17.0
		}
	}
	headerY := y
	drawing.text(panelX, headerY, 13, "start", "#111", "700", fmt.Sprintf(drawing.words("接触时刻 (%s)", "Contacts (%s)"), zoneLabel(event.Greatest.In(drawing.options.location))))
	for _, item := range visibility {
		if item.hidden() {
			drawing.text(870, headerY, 11, "end", "#b21f16", "", drawing.words("✕ 地面看不到", "✕ not observable"))
			break
		}
	}
	for index, contact := range contacts {
		baseline := y + 20 + float64(index)*rowStep
		color := "#222"
		if visibility[index].hidden() {
			color = "#9a9a9a"
			drawing.text(panelX+200, baseline, 12, "start", "#b21f16", "", "✕")
		}
		drawing.text(panelX, baseline, 12, "start", color, "", svgchart.EllipsizeText(contactLabel(contact), 132, 12))
		drawing.text(panelX+194, baseline, 12, "end", color, "", drawing.clock(contact.date))
	}
	y += 20 + float64(len(contacts))*rowStep + detailGap
	details := []string{
		fmt.Sprintf(drawing.words("历时  %s", "Duration  %s"), formatDuration(event.Duration)),
		fmt.Sprintf(drawing.words("遮挡  %s", "Active  %s"), satelliteName(event.Active-1, drawing.options.english)),
		fmt.Sprintf(drawing.words("被遮  %s", "Passive  %s"), satelliteName(event.Passive-1, drawing.options.english)),
	}
	for index, line := range details {
		drawing.text(panelX, y+float64(index)*detailStep, 12, "start", "#333", "", line)
	}
	if event.Type == jupiter.GalileanMutualEclipse {
		greatest := greatestContact(contacts)
		drawEclipseLegend(drawing, event, y, coreShadow(greatest), scale)
	} else {
		drawOccultationLegend(drawing, event, y)
	}
	y += float64(len(details)) * detailStep
	for index, line := range earthLines {
		drawing.text(panelX, y+float64(index)*16, 11, "start", "#333", "", line)
	}
}

func greatestContact(contacts []mutualContact) basic.JupiterGalileanMutualDiagram {
	for _, contact := range contacts {
		if contact.greatest {
			return contact.geometry
		}
	}
	return basic.JupiterGalileanMutualDiagram{}
}

func mutualFooter(drawing *chart, event jupiter.GalileanMutualEvent) []string {
	plane := drawing.words("地球视平面，上北下南、左东右西；被遮星居中，虚线为遮挡星相对路径。", "Earth view, N up, E left; dashed relative path.")
	if event.Type == jupiter.GalileanMutualEclipse {
		plane = drawing.words("被食星处太阳视平面，上北下南、左东右西；与地球视角相差约一个相位角。", "Sun view, N up, E left; Earth view differs by the phase angle.")
	}
	scale := drawing.words("盘面按相对尺寸绘制，覆盖率为几何面积比。", "Disks to scale; geometric coverage. ")
	return []string{plane, scale + drawing.timeNote(event.Greatest)}
}

// mutualWarning 极大时不可见优先；否则列出地面看不到的接触点。
func mutualWarning(drawing *chart, event jupiter.GalileanMutualEvent, contacts []mutualContact, visibility []contactVisibility) string {
	if !mutualVisible(event) {
		return drawing.words("极大时被遮星位于木星后方或木影中，地面不可见", "At greatest the passive satellite is behind Jupiter or in its shadow; not observable")
	}
	var hidden []string
	note := ""
	for index, contact := range contacts {
		if visibility[index].hidden() {
			hidden = append(hidden, contactShort(drawing, contact))
			note = drawing.visibilityNote(visibility[index])
		}
	}
	if len(hidden) == 0 {
		return ""
	}
	targetWord := drawing.words("被掩星", "occulted satellite")
	if event.Type == jupiter.GalileanMutualEclipse {
		targetWord = drawing.words("被食星", "eclipsed satellite")
	}
	return fmt.Sprintf(drawing.words("%s 时%s%s，地面看不到", "At %s the %s is %s; not observable from Earth"), strings.Join(hidden, drawing.words("、", ", ")), targetWord, note)
}

const (
	dayFill           = "#eadfc8"
	penumbraFill      = "#e4e4e1"
	penumbraOverlay   = "#8a8478"
	penumbraOpacity   = 0.22
	minimumCoreRadius = 2.5
)

type coreKind int

const (
	coreNone coreKind = iota
	coreUmbra
	coreAntumbra
)

type coreShadowInfo struct {
	kind   coreKind
	radius float64
}

// coreShadow r_u > 0 为本影，r_u < 0 时锥顶已在被食星之前、|r_u| 为伪本影半径。
func coreShadow(geometry basic.JupiterGalileanMutualDiagram) coreShadowInfo {
	switch {
	case geometry.UmbraRadiusJupiterR > 0:
		return coreShadowInfo{coreUmbra, geometry.UmbraRadiusJupiterR}
	case geometry.UmbraRadiusJupiterR < 0:
		return coreShadowInfo{coreAntumbra, -geometry.UmbraRadiusJupiterR}
	}
	return coreShadowInfo{}
}

func (core coreShadowInfo) id() string {
	if core.kind == coreAntumbra {
		return "antumbra"
	}
	return "umbra"
}

func (core coreShadowInfo) fill() string {
	if core.kind == coreAntumbra {
		return "#e08a2c"
	}
	return "#b21f16"
}

func (core coreShadowInfo) stroke() string {
	if core.kind == coreAntumbra {
		return "#a85f12"
	}
	return "#83170f"
}

func (core coreShadowInfo) name(drawing *chart) string {
	if core.kind == coreAntumbra {
		return drawing.words("伪本影", "antumbra")
	}
	return drawing.words("本影", "umbra")
}

const (
	passiveStroke = "#2b5a8a"
	activeFill    = "#c98b5e"
	activeStroke  = "#8a3d14"
)

// drawOccultationLegend 互掩图例：前景遮挡星、后景被掩星与被遮部分，颜色与盘面一致。
func drawOccultationLegend(drawing *chart, event jupiter.GalileanMutualEvent, y float64) {
	rows := []struct{ fill, stroke, label string }{
		{activeFill, activeStroke, satelliteName(event.Active-1, drawing.options.english) + drawing.words("（前，遮挡）", " (front)")},
		{dayFill, passiveStroke, satelliteName(event.Passive-1, drawing.options.english) + drawing.words("（后，被掩）", " (behind)")},
		{"#d98a80", "#b21f16", drawing.words("被遮部分", "covered part")},
	}
	for index, row := range rows {
		fmt.Fprintf(drawing, `<rect x="%.3f" y="%.3f" width="12" height="9" fill="%s" stroke="%s" stroke-width="1"/>`, panelX+132, y+float64(index)*18-8, row.fill, row.stroke)
		drawing.text(panelX+150, y+float64(index)*18, 11, "start", "#333", "", row.label)
	}
}

// labelDisk 沿 (dx,dy) 方向在盘缘外放名称，短引线从盘缘连到文字；被占时沿圆周旋转找空位。
func labelDisk(drawing *chart, cx, cy, radius, dx, dy float64, color, text string) {
	base := math.Atan2(dy, dx)
	for _, turn := range []float64{0, 0.5, -0.5, 1, -1, 1.5, -1.5, 2, -2, math.Pi} {
		angle := base + turn
		ux, uy := math.Cos(angle), math.Sin(angle)
		for _, gap := range []float64{14, 26} {
			lx, ly := cx+ux*(radius+gap), cy+uy*(radius+gap)
			anchor := "middle"
			if ux > 0.35 {
				anchor = "start"
			} else if ux < -0.35 {
				anchor = "end"
			}
			baseline := ly + 4
			if uy < -0.35 {
				baseline = ly
			} else if uy > 0.35 {
				baseline = ly + 10
			}
			if drawing.labels.OverlapsText(lx, baseline, 13, text, anchor) {
				continue
			}
			fmt.Fprintf(drawing, `<line class="disk-label-leader" x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="%s" stroke-width="0.8"/>`, cx+ux*(radius+1), cy+uy*(radius+1), cx+ux*(radius+gap-4), cy+uy*(radius+gap-4), color)
			drawing.text(lx, baseline, 13, anchor, color, "700", text)
			drawing.labels.ReserveText(lx, baseline, 13, text, anchor)
			return
		}
	}
	drawing.placeText(13, color, text, svgchart.LabelCandidates(cx, cy+4, "middle", radius+18))
}
