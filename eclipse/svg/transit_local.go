package svg

import (
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/internal/timenote"
)

const (
	transitLocalDefaultWidth  = 960
	transitLocalDefaultHeight = 680
)

// TransitLocalSVGOptions 控制站心凌日日面图。
// TransitLocalSVGOptions controls a topocentric planetary-transit solar-disk chart.
type TransitLocalSVGOptions struct {
	// Width 和 Height 是 SVG 画布尺寸；宽度小于 640 或高度小于 480 时使用 960x680 默认值。
	// Width and Height are SVG canvas dimensions. Width values below 640 and height values below 480 use the 960x680 defaults.
	Width  int
	Height int
	// Step 是行星轨迹采样步长；非正值使用两分钟。
	// Step is the planet-track sampling step. Non-positive values use two minutes.
	Step time.Duration
	// Language 为 "en"（不区分大小写）时使用英文，否则使用中文。
	// Language uses English for "en" (case-insensitive) and Chinese otherwise.
	Language string
	// Location 控制显示的事件时刻；nil 使用 date.Location()。
	// Location controls displayed event times. Nil uses date.Location().
	Location *time.Location
	// TimeScale 选择图中时刻的时标：零值 UTC；TimeScaleUT1 要求 Location 为 nil 或 UTC，否则返回 false。
	// TimeScale selects the label scale: the zero value is UTC; TimeScaleUT1 requires Location to be nil or UTC, otherwise the renderer returns false.
	TimeScale astro.TimeScale
	// 空文本字段使用本地化的自动标签。
	// Empty text fields use localized automatic labels.
	Title      string
	FooterNote string
}

// TransitLocalSVG 绘制距 date 最近一次水星或金星凌日在观测地（经度 lon、纬度 lat，东经北纬为正，海拔 height 米）
// 的站心日面图：上北左东的天空视向，画出行星中心轨迹、各接触与凌甚时的行星圆面，以及接触时刻表。
// 轨迹上太阳中心在视地平下（计入标准大气折射）的部分画成灰色虚线。该地几何上看不到凌日、找不到凌日或选项无效时返回 false。
// TransitLocalSVG renders the topocentric solar-disk view of the Mercury or Venus transit closest to date, seen from
// longitude lon and latitude lat (east and north positive) at height metres: north up and east left as on the sky, with
// the planet-centre track, the planet disk at each contact and at greatest, and a contact table. Track parts where the
// Sun's centre is below the apparent horizon (standard refraction) are drawn grey and dashed. It returns false when the
// site sees no geometric transit, no transit is found or the options are invalid.
func TransitLocalSVG(planet TransitPlanet, date time.Time, lon, lat, height float64, options TransitLocalSVGOptions) (string, bool) {
	transit, ok := closestPlanetTransit(planet, date)
	if !ok {
		return "", false
	}
	if options.TimeScale == astro.TimeScaleUT1 {
		if options.Location != nil && options.Location != time.UTC {
			return "", false
		}
		options.Location = time.UTC
	}
	local := basic.PlanetTransitLocal(transit, basic.Observer{Longitude: lon, Latitude: lat, Height: height})
	if !local.Valid {
		return "", false
	}
	if options.Step <= 0 {
		options.Step = 2 * time.Minute
	}
	frames := basic.PlanetTransitLocalFrames(transit, local, options.Step.Hours()/24)
	if len(frames) < 2 {
		return "", false
	}
	options = normalizeTransitLocalSVGOptions(date, options)
	return renderTransitLocalSVG(planet, local, frames, options), true
}

func normalizeTransitLocalSVGOptions(date time.Time, options TransitLocalSVGOptions) TransitLocalSVGOptions {
	if options.Width < 640 {
		options.Width = transitLocalDefaultWidth
	}
	if options.Height < 480 {
		options.Height = transitLocalDefaultHeight
	}
	if strings.EqualFold(options.Language, "en") {
		options.Language = "en"
	} else {
		options.Language = "zh"
	}
	if options.Location == nil {
		options.Location = date.Location()
	}
	return options
}

// transitLocalEvent 是日面上标注的一个接触或凌甚。
type transitLocalEvent struct {
	label, name string
	contact     basic.PlanetTransitLocalContact
	separation  float64
}

func transitLocalEvents(local basic.PlanetTransitLocalResult, language string) []transitLocalEvent {
	names := []string{"一触", "二触", "凌甚", "三触", "四触"}
	if language == "en" {
		names = []string{"C1", "C2", "Greatest", "C3", "C4"}
	}
	external := local.SunSemidiameterArcsec + local.PlanetSemidiameterArcsec
	internal := local.SunSemidiameterArcsec - local.PlanetSemidiameterArcsec
	candidates := []transitLocalEvent{
		{"C1", names[0], local.ExternalIngress, external},
		{"C2", names[1], local.InternalIngress, internal},
		{"G", names[2], local.Greatest, local.MinimumSeparationArcsec},
		{"C3", names[3], local.InternalEgress, internal},
		{"C4", names[4], local.ExternalEgress, external},
	}
	events := candidates[:0]
	for _, event := range candidates {
		if event.contact.Valid {
			events = append(events, event)
		}
	}
	return events
}

// transitLocalSunUp 判断太阳中心几何高度 altitude 是否在视地平上（标准大气折射）。
func transitLocalSunUp(altitude float64) bool {
	return altitude > -basic.RefractionFromApparentAltitude(0, transitMapRefractionPressureHPa, transitMapRefractionTemperatureC)
}

func transitDisplayTime(value time.Time, scale astro.TimeScale, location *time.Location) time.Time {
	if value.IsZero() {
		return value
	}
	if scale == astro.TimeScaleUT1 {
		return basic.UT1FromUTC(value)
	}
	return value.In(location)
}

func renderTransitLocalSVG(
	planet TransitPlanet,
	local basic.PlanetTransitLocalResult,
	frames []basic.PlanetTransitLocalFrame,
	options TransitLocalSVGOptions,
) string {
	width, height := float64(options.Width), float64(options.Height)
	at := func(jd float64) time.Time {
		return transitDisplayTime(basic.JD2DateByZone(jd, time.UTC, false), options.TimeScale, options.Location)
	}
	title := options.Title
	if title == "" {
		date := at(local.Greatest.JD).Format("2006-01-02")
		if options.Language == "en" {
			title = fmt.Sprintf("%s Transit of %s, Local Circumstances", date, transitMapPlanetName(planet, "en"))
		} else {
			title = fmt.Sprintf("%s %s凌日站心图", date, transitMapPlanetName(planet, "zh"))
		}
	}
	site := transitLocalSiteText(local.Observer, options.Language)
	summary := transitLocalSummaryText(local, options.Language)

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s">`,
		options.Width, options.Height, options.Width, options.Height, html.EscapeString(title))
	b.WriteString(`<defs><radialGradient id="transit-sun" cx="40%" cy="34%" r="64%"><stop offset="0%" stop-color="#fff8bb"/><stop offset="58%" stop-color="#ffd55f"/><stop offset="100%" stop-color="#f5a623"/></radialGradient></defs>`)
	b.WriteString(`<rect width="100%" height="100%" fill="#efefed"/>`)
	fmt.Fprintf(&b, `<rect x="22" y="18" width="%.3f" height="%.3f" fill="#ffffff" stroke="#c9c9c6" stroke-width="1.2"/>`, width-44, height-36)
	titleText := title
	if options.Title != "" {
		titleText = svgchart.EllipsizeText(title, width-52, 24)
	}
	fmt.Fprintf(&b, `<text x="%.3f" y="47" fill="#111111" font-family="Georgia, 'Times New Roman', serif" font-size="24" font-weight="700" text-anchor="middle">%s</text>`,
		width/2, html.EscapeString(titleText))
	for index, line := range []string{site, summary} {
		fmt.Fprintf(&b, `<text x="%.3f" y="%.3f" fill="#333" font-family="Georgia, 'Times New Roman', serif" font-size="13" text-anchor="middle">%s</text>`,
			width/2, 72+float64(index)*19, html.EscapeString(line))
	}

	// 日面区占左侧，接触表占右侧。
	panelWidth := 300.0
	top, bottom := 112.0, height-86
	diskLeft, diskRight := 46.0, width-panelWidth-60
	cx, cy := (diskLeft+diskRight)/2, (top+bottom)/2
	extent := local.SunSemidiameterArcsec + 2.2*local.PlanetSemidiameterArcsec + 60
	scale := math.Min((diskRight-diskLeft)/2, (bottom-top)/2) / extent
	mapX := func(x float64) float64 { return cx - x*scale } // 东在左
	mapY := func(y float64) float64 { return cy - y*scale }
	sunRadius := local.SunSemidiameterArcsec * scale
	fmt.Fprintf(&b, `<circle class="transit-sun" cx="%.3f" cy="%.3f" r="%.3f" fill="url(#transit-sun)" stroke="#c78211" stroke-width="1.4"/>`, cx, cy, sunRadius)
	writeTransitLocalAxes(&b, cx, cy, sunRadius, options.Language)
	writeTransitLocalTrack(&b, frames, mapX, mapY)

	events := transitLocalEvents(local, options.Language)
	planetRadius := math.Max(local.PlanetSemidiameterArcsec*scale, 2)
	table := &svgchart.LabelTable{}
	for _, event := range events {
		x, y := transitLocalEventPosition(frames, event.contact.JD, mapX, mapY)
		fill, stroke := "#1d1d1d", "#000000"
		if !transitLocalSunUp(event.contact.SunAltitudeDeg) {
			fill, stroke = "#9a9a9a", "#6d6d6d"
		}
		fmt.Fprintf(&b, `<circle class="transit-planet-%s" cx="%.3f" cy="%.3f" r="%.3f" fill="%s" fill-opacity="0.92" stroke="%s" stroke-width="0.8"/>`,
			strings.ToLower(event.label), x, y, planetRadius, fill, stroke)
		table.Reserve(x-planetRadius, y-planetRadius, 2*planetRadius, 2*planetRadius)
	}
	for _, event := range events {
		x, y := transitLocalEventPosition(frames, event.contact.JD, mapX, mapY)
		// 标签放在远离日心的一侧，避开日面内的轨迹与圆面。
		dx, dy := x-cx, y-cy
		norm := math.Hypot(dx, dy)
		if norm < 1e-6 {
			dx, dy, norm = 0, -1, 1
		}
		text := event.name + " " + at(event.contact.JD).Format("15:04:05")
		var placement svgchart.LabelPlacement
		placed := false
		// 相邻接触相隔仅数分钟时近处位次全被占，逐圈外推并画引线，不得丢标签。
		for ring := 0; ring < 6 && !placed; ring++ {
			offset := planetRadius + 12 + float64(ring)*16
			candidates := []svgchart.LabelPlacement{}
			for _, turn := range []float64{0, 0.5, -0.5, 1, -1} {
				ux := dx/norm*math.Cos(turn) - dy/norm*math.Sin(turn)
				uy := dx/norm*math.Sin(turn) + dy/norm*math.Cos(turn)
				anchor := "start"
				if ux < -0.3 {
					anchor = "end"
				} else if ux < 0.3 {
					anchor = "middle"
				}
				candidate := svgchart.LabelPlacement{X: x + ux*offset, Y: y + uy*offset + 4, Anchor: anchor}
				// 标签平移进白色图框内、接触表左侧，不能伸出画面。
				box := svgchart.LabelTextBox(candidate.X, candidate.Y, 11, text, anchor)
				candidate.X += math.Max(0, 30-box.X) - math.Max(0, box.X+box.Width-(width-panelWidth-44))
				candidate.Y += math.Max(0, top-8-box.Y) - math.Max(0, box.Y+box.Height-(height-30))
				candidates = append(candidates, candidate)
			}
			placement, placed = table.Place(text, 11, candidates)
			if placed && transitLocalLabelGap(placement, text, x, y) > planetRadius+16 {
				// 引线连到标签框上离接触点最近的位置。
				box := svgchart.LabelTextBox(placement.X, placement.Y, 11, text, placement.Anchor)
				endX := math.Max(box.X, math.Min(x, box.X+box.Width))
				endY := math.Max(box.Y, math.Min(y, box.Y+box.Height))
				fmt.Fprintf(&b, `<line class="transit-event-leader" x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#555" stroke-width="0.7" opacity="0.7"/>`,
					x, y, endX, endY)
			}
		}
		if !placed {
			continue
		}
		fmt.Fprintf(&b, `<text class="transit-event-label" x="%.3f" y="%.3f" fill="#222" stroke="#ffffff" stroke-width="2.6" paint-order="stroke" font-family="Arial, sans-serif" font-size="11" text-anchor="%s">%s</text>`,
			placement.X, placement.Y, placement.Anchor, html.EscapeString(text))
	}

	writeTransitLocalContactTable(&b, local, events, at, width-panelWidth-36, top+10, panelWidth, options.Language)
	writeTransitLocalFooter(&b, local, options, width, height)
	b.WriteString(`</svg>`)
	return b.String()
}

// transitLocalLabelGap 是接触点到标签框的最近距离。
func transitLocalLabelGap(placement svgchart.LabelPlacement, text string, x, y float64) float64 {
	box := svgchart.LabelTextBox(placement.X, placement.Y, 11, text, placement.Anchor)
	return math.Hypot(x-math.Max(box.X, math.Min(x, box.X+box.Width)), y-math.Max(box.Y, math.Min(y, box.Y+box.Height)))
}

func transitLocalSiteText(observer basic.Observer, language string) string {
	ew, ns := "E", "N"
	if language != "en" {
		ew, ns = "东经", "北纬"
	}
	lon, lat := observer.Longitude, observer.Latitude
	if lon < 0 {
		ew = map[bool]string{true: "W", false: "西经"}[language == "en"]
		lon = -lon
	}
	if lat < 0 {
		ns = map[bool]string{true: "S", false: "南纬"}[language == "en"]
		lat = -lat
	}
	if language == "en" {
		return fmt.Sprintf("Site %.4f°%s %.4f°%s, height %.0f m", lon, ew, lat, ns, observer.Height)
	}
	return fmt.Sprintf("观测地 %s %.4f° %s %.4f°，海拔 %.0f 米", ew, lon, ns, lat, observer.Height)
}

func transitLocalSummaryText(local basic.PlanetTransitLocalResult, language string) string {
	kind := map[bool]string{true: "完整凌日", false: "掠凌（无内切接触）"}[local.HasInternal]
	visible := map[bool]string{true: "可见", false: "全程太阳在地平下"}[local.Visible]
	if language == "en" {
		kind = map[bool]string{true: "full transit", false: "grazing transit (no internal contacts)"}[local.HasInternal]
		visible = map[bool]string{true: "visible", false: "Sun below the horizon throughout"}[local.Visible]
		return fmt.Sprintf("Topocentric %s, %s | minimum separation %.1f″ | Sun %.1f″ | planet %.2f″",
			kind, visible, local.MinimumSeparationArcsec, local.SunSemidiameterArcsec, local.PlanetSemidiameterArcsec)
	}
	return fmt.Sprintf("站心%s，%s | 最小角距 %.1f″ | 太阳视半径 %.1f″ | 行星视半径 %.2f″",
		kind, visible, local.MinimumSeparationArcsec, local.SunSemidiameterArcsec, local.PlanetSemidiameterArcsec)
}

// writeTransitLocalAxes 画过日心的南北、东西虚线与方向字。
func writeTransitLocalAxes(b *strings.Builder, cx, cy, radius float64, language string) {
	reach := radius + 18
	fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#8b6a2b" stroke-width="0.8" stroke-dasharray="3 4" stroke-opacity="0.7"/>`, cx, cy-reach, cx, cy+reach)
	fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#8b6a2b" stroke-width="0.8" stroke-dasharray="3 4" stroke-opacity="0.7"/>`, cx-reach, cy, cx+reach, cy)
	labels := [4]string{"北", "南", "东", "西"}
	if language == "en" {
		labels = [4]string{"N", "S", "E", "W"}
	}
	for index, item := range []struct{ x, y float64 }{
		{cx, cy - reach - 6}, {cx, cy + reach + 16}, {cx - reach - 12, cy + 5}, {cx + reach + 12, cy + 5},
	} {
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#6b5222" font-family="Georgia, 'Times New Roman', serif" font-size="13" text-anchor="middle">%s</text>`,
			item.x, item.y, labels[index])
	}
}

// writeTransitLocalTrack 画行星中心轨迹：太阳在视地平上的段为实线，地平下为灰色虚线。
func writeTransitLocalTrack(b *strings.Builder, frames []basic.PlanetTransitLocalFrame, mapX, mapY func(float64) float64) {
	start := 0
	for index := 1; index <= len(frames); index++ {
		if index < len(frames) && transitLocalSunUp(frames[index].SunAltitudeDeg) == transitLocalSunUp(frames[start].SunAltitudeDeg) {
			continue
		}
		// 段尾取下一段首点，相邻两段首尾相接。
		end := index
		if end == len(frames) {
			end--
		}
		class, color, dash := "transit-track", "#24518a", "none"
		if !transitLocalSunUp(frames[start].SunAltitudeDeg) {
			class, color, dash = "transit-track-below-horizon", "#8a8a8a", "5 4"
		}
		fmt.Fprintf(b, `<path class="%s" d="`, class)
		for position := start; position <= end; position++ {
			command := "L"
			if position == start {
				command = "M"
			}
			fmt.Fprintf(b, `%s %.3f %.3f `, command, mapX(frames[position].PlanetXArcsec), mapY(frames[position].PlanetYArcsec))
		}
		fmt.Fprintf(b, `" fill="none" stroke="%s" stroke-width="1.4" stroke-dasharray="%s" stroke-linecap="round"/>`, color, dash)
		start = index
	}
}

// transitLocalEventPosition 取时刻最接近 jd 的采样帧位置；接触与凌甚时刻本身在采样里。
func transitLocalEventPosition(frames []basic.PlanetTransitLocalFrame, jd float64, mapX, mapY func(float64) float64) (float64, float64) {
	best := 0
	for index := range frames {
		if math.Abs(frames[index].JD-jd) < math.Abs(frames[best].JD-jd) {
			best = index
		}
	}
	return mapX(frames[best].PlanetXArcsec), mapY(frames[best].PlanetYArcsec)
}

func writeTransitLocalContactTable(
	b *strings.Builder,
	local basic.PlanetTransitLocalResult,
	events []transitLocalEvent,
	at func(float64) time.Time,
	x, y, width float64,
	language string,
) {
	heading := [4]string{"接触时刻", "时刻", "太阳高度", "方位"}
	if language == "en" {
		heading = [4]string{"Contacts", "Time", "Sun alt.", "Az."}
	}
	fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#111" font-family="Georgia, 'Times New Roman', serif" font-size="15" font-weight="700">%s</text>`, x, y, heading[0])
	columns := []float64{x, x + 78, x + 176, x + 252}
	rowY := y + 26
	for index, text := range []string{"", heading[1], heading[2], heading[3]} {
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#555" font-family="Arial, sans-serif" font-size="11">%s</text>`, columns[index], rowY, text)
	}
	fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#bbb" stroke-width="0.8"/>`, x, rowY+6, x+width-8, rowY+6)
	for _, event := range events {
		rowY += 22
		fill := "#222"
		if !transitLocalSunUp(event.contact.SunAltitudeDeg) {
			fill = "#8a8a8a"
		}
		cells := []string{
			event.name,
			at(event.contact.JD).Format("15:04:05"),
			fmt.Sprintf("%+.1f°", event.contact.SunAltitudeDeg),
			fmt.Sprintf("%.1f°", event.contact.SunAzimuthDeg),
		}
		for index, text := range cells {
			fmt.Fprintf(b, `<text class="transit-contact-row" x="%.3f" y="%.3f" fill="%s" font-family="Arial, sans-serif" font-size="12">%s</text>`,
				columns[index], rowY, fill, html.EscapeString(text))
		}
	}
	rowY += 30
	lines := []string{}
	duration := time.Duration((local.ExternalEgress.JD - local.ExternalIngress.JD) * 86400 * float64(time.Second)).Round(time.Second)
	if language == "en" {
		lines = append(lines, "Duration C1–C4 "+duration.String())
	} else {
		lines = append(lines, "一触至四触历时 "+duration.String())
	}
	for _, item := range []struct {
		jd     float64
		zh, en string
	}{{local.Sunrise, "凌日中日出 ", "Sunrise during transit "}, {local.Sunset, "凌日中日落 ", "Sunset during transit "}} {
		if item.jd == 0 {
			continue
		}
		label := item.zh
		if language == "en" {
			label = item.en
		}
		lines = append(lines, label+at(item.jd).Format("15:04:05"))
	}
	for _, line := range lines {
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#333" font-family="Arial, sans-serif" font-size="12">%s</text>`, x, rowY, html.EscapeString(line))
		rowY += 19
	}
}

func writeTransitLocalFooter(b *strings.Builder, local basic.PlanetTransitLocalResult, options TransitLocalSVGOptions, width, height float64) {
	note := options.FooterNote
	if note == "" {
		if options.Language == "en" {
			note = "Sky view, north up and east left. Topocentric contacts; Sun altitudes are geometric centre altitudes, grey parts have the Sun's centre below the apparent horizon (standard refraction)."
		} else {
			note = "天空视向，上北左东；站心接触时刻，太阳高度为日心几何高度，灰色部分太阳中心在视地平下（计入标准大气折射）。"
		}
	}
	scaleNote := timenote.Scale(options.TimeScale, basic.JD2DateByZone(local.Greatest.JD, time.UTC, false), options.Location, options.Language)
	lines := svgchart.FooterLinesWithScale(svgchart.WrapText(note, width-80, 12), scaleNote, width-80, 12, 3)
	baselines, _ := svgchart.FooterBlock(len(lines), height, 12, svgchart.FooterLineHeight(12), svgchart.FooterBottomPadding(12)+18)
	for index, line := range lines {
		fmt.Fprintf(b, `<text x="40" y="%.3f" fill="#555" font-family="Georgia, 'Times New Roman', serif" font-size="12">%s</text>`,
			baselines[index], html.EscapeString(line))
	}
}
