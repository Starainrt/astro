package svg

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/svgchart"
)

const (
	systemCenterX     = 450.0
	systemCenterY     = 262.0
	systemHalfY       = 118.0
	systemHalf        = 380.0
	systemDiskStroke  = 0.9
	jupiterDiskStroke = 1.0
)

const jovianDefs = `<radialGradient id="jovian-disk" cx="42%" cy="38%" r="70%"><stop offset="0%" stop-color="#f6ead2"/><stop offset="70%" stop-color="#dcc19a"/><stop offset="100%" stop-color="#b8956a"/></radialGradient>`

type systemMutual struct {
	active, passive int
	eclipse         bool
	coverage        float64
}

// GalileanSystemSVG 伽利略卫星地球视点位置图与状态表 / Earth-view Galilean satellite configuration chart with a state table.
func GalileanSystemSVG(date time.Time, options GalileanSystemSVGOptions) (string, bool) {
	settings := normalizeOptions(options.Width, options.Height, options.Title, options.Language, options.Location, options.TrueScale)
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	snapshot, ok := basic.JupiterGalileanDiagramWithInstance(jde, nil)
	if !ok {
		return "", false
	}
	probe := &chart{options: settings}
	title := settings.title
	if title == "" {
		title = probe.words("木星伽利略卫星位置图", "Galilean Satellites of Jupiter")
	}
	drawing := newChart(settings, title, jovianDefs)
	drawing.header(title, []string{
		drawing.stamp(date),
		fmt.Sprintf(drawing.words("木星北极位置角 %.1f°  投影扁率 %.4f", "Jupiter north pole PA %.1f°  projected axis ratio %.4f"), normalizeDegrees(snapshot.PolePositionAngle), snapshot.MinorRadiusRatio),
	})
	mutuals := systemMutuals(jde)
	scale := systemScale(snapshot)
	closeDisks := drawSystemPlane(drawing, snapshot, scale, jde)
	drawSystemTable(drawing, snapshot, mutuals)
	drawJupiterInset(drawing, snapshot)
	drawing.footer(systemFooter(drawing, date, closeDisks))
	return drawing.finish()
}

func normalizeDegrees(value float64) float64 {
	value = math.Mod(value, 360)
	if value < 0 {
		value += 360
	}
	return value
}

func systemScale(snapshot basic.JupiterGalileanDiagram) float64 {
	extentX, extentY := 3.0, 1.5
	for _, observation := range snapshot.Observations {
		extentX = math.Max(extentX, math.Abs(observation.OffsetXJupiterRadii)+1)
		extentY = math.Max(extentY, math.Abs(observation.OffsetYJupiterRadii)+0.6)
	}
	return math.Min(systemHalf/extentX, systemHalfY/extentY)
}

func systemMutuals(jde float64) []systemMutual {
	var result []systemMutual
	for first := 1; first <= 4; first++ {
		for second := first + 1; second <= 4; second++ {
			for _, kind := range []basic.JupiterGalileanMutualPhenomenonType{basic.JupiterGalileanMutualOccultation, basic.JupiterGalileanMutualEclipse} {
				geometry, ok := basic.JupiterGalileanMutualDiagramWithInstance(jde, first, second, kind, nil)
				if !ok {
					continue
				}
				radius := geometry.ActiveRadiusJupiterR
				if kind == basic.JupiterGalileanMutualEclipse {
					radius = geometry.PenumbraRadiusJupiterR
				}
				if math.Hypot(geometry.XJupiterR, geometry.YJupiterR) >= radius+geometry.PassiveRadiusJupiterR {
					continue
				}
				result = append(result, systemMutual{geometry.Active, geometry.Passive, kind == basic.JupiterGalileanMutualEclipse, geometry.Coverage})
			}
		}
	}
	return result
}

func (drawing *chart) satelliteRadius(snapshot basic.JupiterGalileanDiagram, index int, scale float64) float64 {
	radius := snapshot.AngularRadiiJupiterR[index] * scale
	if drawing.options.trueScale {
		return radius
	}
	return schematicSatelliteRadius(radius)
}

func schematicSatelliteRadius(radius float64) float64 {
	return math.Max(4.5, 3*radius)
}

type systemDiskStyle struct {
	radius, stroke float64
}

func (drawing *chart) systemDiskStyles(snapshot basic.JupiterGalileanDiagram, scale float64) (styles [4]systemDiskStyle, closeDisks, clipLimb bool) {
	const svgRoundingMargin = 0.003
	var trueSize [4]bool
	sine, cosine := math.Sincos(snapshot.PolePositionAngle * math.Pi / 180)
	minorRatio := snapshot.MinorRadiusRatio
	for index, observation := range snapshot.Observations {
		physicalRadius := snapshot.AngularRadiiJupiterR[index] * scale
		enlargedRadius := schematicSatelliteRadius(physicalRadius)
		styles[index] = systemDiskStyle{enlargedRadius, systemDiskStroke}
		if drawing.options.trueScale {
			styles[index].radius = physicalRadius
		}
		east, north := observation.OffsetXJupiterRadii, observation.OffsetYJupiterRadii
		along, across := east*cosine-north*sine, east*sine+north*cosine
		limbGapLowerBound := (math.Hypot(along, across/minorRatio) - 1) * minorRatio * scale
		if limbGapLowerBound <= enlargedRadius+(systemDiskStroke+jupiterDiskStroke)/2+svgRoundingMargin {
			trueSize[index], clipLimb = true, true
		}
	}
	if !drawing.options.trueScale {
		for first, left := range snapshot.Observations {
			for second := first + 1; second < len(styles); second++ {
				right := snapshot.Observations[second]
				separation := math.Hypot(left.OffsetXJupiterRadii-right.OffsetXJupiterRadii, left.OffsetYJupiterRadii-right.OffsetYJupiterRadii) * scale
				if separation <= styles[first].radius+styles[second].radius+systemDiskStroke+svgRoundingMargin {
					trueSize[first], trueSize[second] = true, true
				}
			}
		}
	}
	for index, physical := range trueSize {
		if physical {
			styles[index] = systemDiskStyle{radius: snapshot.AngularRadiiJupiterR[index] * scale}
			closeDisks = true
		}
	}
	return styles, closeDisks, clipLimb
}

func galileanPaintOrder(snapshot basic.JupiterGalileanDiagram) [4]int {
	order := [4]int{0, 1, 2, 3}
	sort.SliceStable(order[:], func(left, right int) bool {
		return snapshot.Observations[order[left]].OffsetZJupiterRadii > snapshot.Observations[order[right]].OffsetZJupiterRadii
	})
	return order
}

func drawSystemPlane(drawing *chart, snapshot basic.JupiterGalileanDiagram, scale, jde float64) bool {
	styles, closeDisks, clipLimb := drawing.systemDiskStyles(snapshot, scale)
	left, right := systemCenterX-systemHalf-20, systemCenterX+systemHalf+20
	fmt.Fprintf(drawing, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#8b8b86" stroke-width="0.8" stroke-dasharray="3 4"/>`, left, systemCenterY, right, systemCenterY)
	drawing.text(left-6, systemCenterY+5, 14, "end", "#111", "700", drawing.words("东", "E"))
	drawing.text(right+6, systemCenterY+5, 14, "start", "#111", "700", drawing.words("西", "W"))
	drawing.labels.Reserve(0, 0, 900, 140)
	drawing.labels.Reserve(0, 410, 900, 190)
	drawing.text(systemCenterX, systemCenterY-systemHalfY-4, 14, "middle", "#111", "700", drawing.words("北", "N"))
	drawing.labels.ReserveText(systemCenterX, systemCenterY-systemHalfY-4, 14, "北", "middle")
	paintOrder := galileanPaintOrder(snapshot)
	for _, foreground := range []bool{false, true} {
		if foreground {
			drawJupiterDisk(drawing, snapshot, scale, clipLimb)
		}
		for _, index := range paintOrder {
			observation := snapshot.Observations[index]
			if observation.InFrontOfJupiter != foreground {
				continue
			}
			x, y := systemCenterX-observation.OffsetXJupiterRadii*scale, systemCenterY-observation.OffsetYJupiterRadii*scale
			style := styles[index]
			fill := satelliteColor(index)
			if snapshot.Phenomena[index].Eclipse {
				fill = "#555"
			}
			fmt.Fprintf(drawing, `<circle id="satellite-%d" cx="%.3f" cy="%.3f" r="%.3f" fill="%s" stroke="#4a4237" stroke-width="%.1f"/>`, index+1, x, y, style.radius, fill, style.stroke)
		}
	}
	for index, observation := range snapshot.Observations {
		x, y := systemCenterX-observation.OffsetXJupiterRadii*scale, systemCenterY-observation.OffsetYJupiterRadii*scale
		radius := styles[index].radius
		drawing.labels.Reserve(x-radius, y-radius, 2*radius, 2*radius)
	}
	drawMotionArrows(drawing, snapshot, styles, scale, jde)
	for index, observation := range snapshot.Observations {
		x, y := systemCenterX-observation.OffsetXJupiterRadii*scale, systemCenterY-observation.OffsetYJupiterRadii*scale
		radius := styles[index].radius
		name := satelliteName(index, drawing.options.english)
		candidates := []svgchart.LabelPlacement{
			{X: x, Y: y - radius - 8, Anchor: "middle"}, {X: x, Y: y + radius + 18, Anchor: "middle"},
			{X: x, Y: y - radius - 26, Anchor: "middle"}, {X: x, Y: y + radius + 36, Anchor: "middle"},
			{X: x, Y: y - radius - 44, Anchor: "middle"}, {X: x, Y: y + radius + 54, Anchor: "middle"},
		}
		drawing.placeText(13, "#2554c7", name, candidates)
	}
	return closeDisks
}

// jovianBands 自北向南的带纹，纬度为行星图心纬度（度），只作示意。
var jovianBands = [...]struct {
	north, south float64
	fill         string
	opacity      float64
}{
	{90, 48, "#b9a58a", 0.55},
	{36, 31, "#a8805a", 0.45},
	{24, 20, "#8f5f3c", 0.55},
	{18, 7, "#9a6a45", 0.75},
	{1, -2, "#c9a77d", 0.35},
	{-7, -19, "#9a6a45", 0.8},
	{-26, -30, "#a8805a", 0.45},
	{-46, -90, "#b9a58a", 0.55},
}

func drawJupiterDisk(drawing *chart, snapshot basic.JupiterGalileanDiagram, scale float64, clipLimb bool) {
	drawJupiterBody(drawing, snapshot, systemCenterX, systemCenterY, scale, "jupiter-disk", "jovian-limb", clipLimb)
	shadowClip := ""
	for index, phenomenon := range snapshot.Phenomena {
		if phenomenon.ShadowTransit {
			if shadowClip == "" {
				shadowClip = "jovian-shadow-limb"
				drawShadowClip(drawing, snapshot, systemCenterX, systemCenterY, scale, shadowClip)
			}
			radius := snapshot.ShadowUmbraRadii[index] * scale
			if !drawing.options.trueScale {
				radius = math.Max(1.5, radius)
			}
			fmt.Fprintf(drawing, `<circle id="shadow-%d" cx="%.3f" cy="%.3f" r="%.3f" fill="#222" clip-path="url(#%s)"/>`, index+1, systemCenterX-phenomenon.ShadowOffsetXJupiterRadii*scale, systemCenterY-phenomenon.ShadowOffsetYJupiterRadii*scale, radius, shadowClip)
		}
	}
	minor := scale * snapshot.MinorRadiusRatio
	angleRad := snapshot.PolePositionAngle * math.Pi / 180
	tipX, tipY := systemCenterX-(minor+22)*math.Sin(angleRad), systemCenterY-(minor+22)*math.Cos(angleRad)
	fmt.Fprintf(drawing, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#555" stroke-width="1"/>`, systemCenterX-minor*math.Sin(angleRad), systemCenterY-minor*math.Cos(angleRad), tipX, tipY)
	label := drawing.words("木星北极", "Jupiter N")
	labelX, labelY := tipX-12*math.Sin(angleRad), tipY-12*math.Cos(angleRad)+4
	drawing.text(labelX, labelY, 11, "middle", "#555", "", label)
	drawing.labels.ReserveText(labelX, labelY, 11, label, "middle")
	drawing.labels.Reserve(systemCenterX-scale, systemCenterY-scale, 2*scale, 2*scale)
}

func drawShadowClip(drawing *chart, snapshot basic.JupiterGalileanDiagram, cx, cy, scale float64, id string) {
	fmt.Fprintf(drawing, `<clipPath id="%s"><ellipse cx="%.3f" cy="%.3f" rx="%.3f" ry="%.3f" transform="rotate(%.3f %.3f %.3f)"/></clipPath>`, id, cx, cy, scale, scale*snapshot.MinorRadiusRatio, -snapshot.PolePositionAngle, cx, cy)
}

// jupiterBody 画按投影扁率与北极位置角旋转的木星盘及带纹。
func jupiterBody(drawing *chart, snapshot basic.JupiterGalileanDiagram, cx, cy, scale float64, id, clip string) {
	drawJupiterBody(drawing, snapshot, cx, cy, scale, id, clip, false)
}

func drawJupiterBody(drawing *chart, snapshot basic.JupiterGalileanDiagram, cx, cy, scale float64, id, clip string, clipOutline bool) {
	minor := scale * snapshot.MinorRadiusRatio
	fmt.Fprintf(drawing, `<g transform="rotate(%.3f %.3f %.3f)"><clipPath id="%s"><ellipse cx="%.3f" cy="%.3f" rx="%.3f" ry="%.3f"/></clipPath>`, -snapshot.PolePositionAngle, cx, cy, clip, cx, cy, scale, minor)
	outlineClip := ""
	if clipOutline {
		outlineClip = fmt.Sprintf(` clip-path="url(#%s)"`, clip)
	}
	fmt.Fprintf(drawing, `<ellipse id="%s" cx="%.3f" cy="%.3f" rx="%.3f" ry="%.3f" fill="url(#jovian-disk)" stroke="#9c7a50" stroke-width="%g"%s/>`, id, cx, cy, scale, minor, jupiterDiskStroke, outlineClip)
	fmt.Fprintf(drawing, `<g clip-path="url(#%s)">`, clip)
	for _, band := range jovianBands {
		top := cy - minor*math.Sin(band.north*math.Pi/180)
		bottom := cy - minor*math.Sin(band.south*math.Pi/180)
		fmt.Fprintf(drawing, `<rect x="%.3f" y="%.3f" width="%.3f" height="%.3f" fill="%s" opacity="%.2f"/>`, cx-scale, top, 2*scale, bottom-top, band.fill, band.opacity)
	}
	fmt.Fprintf(drawing, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#7a5535" stroke-width="0.6" stroke-dasharray="2 2" opacity="0.6"/>`, cx-scale, cy, cx+scale, cy)
	fmt.Fprintf(drawing, `<ellipse cx="%.3f" cy="%.3f" rx="%.3f" ry="%.3f" fill="none" stroke="#6b4a2c" stroke-width="%.3f" opacity="0.18"/>`, cx, cy, scale*0.985, minor*0.985, scale*0.08)
	drawing.WriteString(`</g></g>`)
}

// drawJupiterInset 右下角放大的木星盘：凌木卫星、影凌黑斑按实际角尺寸画在盘面上。
func drawJupiterInset(drawing *chart, snapshot basic.JupiterGalileanDiagram) {
	const cx, cy, scale = 790.0, 470.0, 52.0
	drawing.text(cx, 408, 12, "middle", "#555", "700", drawing.words("木星盘面", "Jupiter disk"))
	jupiterBody(drawing, snapshot, cx, cy, scale, "jupiter-inset", "jovian-inset-limb")
	shadowClip := ""
	for index, phenomenon := range snapshot.Phenomena {
		if phenomenon.ShadowTransit {
			if shadowClip == "" {
				shadowClip = "jovian-inset-shadow-limb"
				drawShadowClip(drawing, snapshot, cx, cy, scale, shadowClip)
			}
			fmt.Fprintf(drawing, `<circle id="inset-shadow-%d" cx="%.3f" cy="%.3f" r="%.3f" fill="#1d1d1d" clip-path="url(#%s)"/>`, index+1, cx-phenomenon.ShadowOffsetXJupiterRadii*scale, cy-phenomenon.ShadowOffsetYJupiterRadii*scale, snapshot.ShadowUmbraRadii[index]*scale, shadowClip)
		}
	}
	for _, index := range galileanPaintOrder(snapshot) {
		if !snapshot.Phenomena[index].Transit {
			continue
		}
		observation := snapshot.Observations[index]
		x, y := cx-observation.OffsetXJupiterRadii*scale, cy-observation.OffsetYJupiterRadii*scale
		fmt.Fprintf(drawing, `<circle id="inset-transit-%d" cx="%.3f" cy="%.3f" r="%.3f" fill="%s" stroke="#4a4237" stroke-width="0.6"/>`, index+1, x, y, snapshot.AngularRadiiJupiterR[index]*scale, satelliteColor(index))
	}
}

func drawSystemTable(drawing *chart, snapshot basic.JupiterGalileanDiagram, mutuals []systemMutual) {
	x0, y0 := 64.0, 432.0
	columns := []float64{x0, x0 + 110, x0 + 250, x0 + 360}
	headers := []string{drawing.words("卫星", "Satellite"), drawing.words("东西距（木星半径）", "E–W offset (RJ)"), drawing.words("前后", "Depth"), drawing.words("状态", "State")}
	for index, header := range headers {
		drawing.text(columns[index], y0, 12, "start", "#555", "700", header)
	}
	fmt.Fprintf(drawing, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#c9c9c6" stroke-width="1"/>`, x0, y0+7, 700.0, y0+7)
	for index, observation := range snapshot.Observations {
		y := y0 + 24 + float64(index)*21
		fmt.Fprintf(drawing, `<circle cx="%.3f" cy="%.3f" r="4.5" fill="%s" stroke="#4a4237" stroke-width="0.8"/>`, columns[0]-12, y-4, satelliteColor(index))
		drawing.text(columns[0], y, 13, "start", "#111", "", satelliteName(index, drawing.options.english))
		side := drawing.words("东", "E")
		if observation.OffsetXJupiterRadii < 0 {
			side = drawing.words("西", "W")
		}
		drawing.text(columns[1], y, 13, "start", "#222", "", fmt.Sprintf("%s %.2f", side, math.Abs(observation.OffsetXJupiterRadii)))
		depth := drawing.words("木星后方", "behind")
		if observation.InFrontOfJupiter {
			depth = drawing.words("木星前方", "in front")
		}
		drawing.text(columns[2], y, 13, "start", "#222", "", depth)
		drawing.text(columns[3], y, 13, "start", "#222", "", svgchart.EllipsizeText(systemState(drawing, index, snapshot.Phenomena[index], mutuals), 276, 13))
	}
}

func systemState(drawing *chart, index int, phenomenon basic.JupiterGalileanPhenomenon, mutuals []systemMutual) string {
	var labels []string
	for flag, active := range []bool{phenomenon.Transit, phenomenon.Occultation, phenomenon.Eclipse, phenomenon.ShadowTransit} {
		if active {
			labels = append(labels, drawing.words([]string{"凌木", "被木星掩", "木影中", "影凌"}[flag], []string{"transit", "occulted", "eclipsed", "shadow transit"}[flag]))
		}
	}
	for _, mutual := range mutuals {
		if mutual.passive != index+1 {
			continue
		}
		other := satelliteName(mutual.active-1, drawing.options.english)
		if mutual.eclipse {
			labels = append(labels, fmt.Sprintf(drawing.words("被%s食 %.0f%%", "eclipsed by %s %.0f%%"), other, 100*mutual.coverage))
		} else {
			labels = append(labels, fmt.Sprintf(drawing.words("被%s掩 %.0f%%", "occulted by %s %.0f%%"), other, 100*mutual.coverage))
		}
	}
	if len(labels) == 0 {
		return drawing.words("可见", "visible")
	}
	return strings.Join(labels, drawing.words("，", ", "))
}

func systemFooter(drawing *chart, date time.Time, closeDisks bool) []string {
	scale := drawing.words("主图卫星盘放大示意，半径至少 4.5 图面单位。", "Main-view disks enlarged for visibility (minimum radius 4.5 chart units). ")
	if drawing.options.trueScale {
		scale = drawing.words("卫星与木星盘按实际角尺寸绘制；", "Satellites and Jupiter at true angular size; ")
	} else if closeDisks {
		scale = drawing.words("接近其他卫星或木星盘的卫星按实际角尺寸绘制，其余放大示意。", "Satellites near each other or Jupiter use true angular size. Others are enlarged. ")
	}
	return []string{
		drawing.words("地心视向，左东右西、上北下南；木星盘按投影扁率与北极位置角绘制，后方卫星被木星盘遮住。", "Geocentric view, N up, E left; Jupiter drawn with projected flattening and pole angle."),
		scale + drawing.timeNote(date),
	}
}

// drawMotionArrows 用 10 分钟后的位置给出各卫星视运动方向，箭头长度固定、只表方向。
func drawMotionArrows(drawing *chart, snapshot basic.JupiterGalileanDiagram, styles [4]systemDiskStyle, scale, jde float64) {
	later, ok := basic.JupiterGalileanDiagramWithInstance(jde+10.0/1440, nil)
	if !ok {
		return
	}
	for index, observation := range snapshot.Observations {
		dx := -(later.Observations[index].OffsetXJupiterRadii - observation.OffsetXJupiterRadii)
		dy := -(later.Observations[index].OffsetYJupiterRadii - observation.OffsetYJupiterRadii)
		length := math.Hypot(dx, dy)
		if length == 0 {
			continue
		}
		dx, dy = dx/length, dy/length
		x, y := systemCenterX-observation.OffsetXJupiterRadii*scale, systemCenterY-observation.OffsetYJupiterRadii*scale
		radius := styles[index].radius
		x0, y0 := x+dx*(radius+3), y+dy*(radius+3)
		x1, y1 := x+dx*(radius+20), y+dy*(radius+20)
		headX, headY := -dy*3.5, dx*3.5
		fmt.Fprintf(drawing, `<g id="motion-%d" stroke="#2554c7" fill="#2554c7"><line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke-width="1.2"/><path d="M %.3f %.3f L %.3f %.3f L %.3f %.3f Z" stroke="none"/></g>`,
			index+1, x0, y0, x1, y1, x1+dx*6, y1+dy*6, x1+headX, y1+headY, x1-headX, y1-headY)
		drawing.labels.Reserve(math.Min(x0, x1+dx*6)-4, math.Min(y0, y1+dy*6)-4, math.Abs(x1+dx*6-x0)+8, math.Abs(y1+dy*6-y0)+8)
	}
}
