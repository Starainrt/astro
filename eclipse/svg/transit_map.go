package svg

import (
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/internal/svgmap"
	"github.com/starainrt/astro/internal/timenote"
)

const (
	transitMapDefaultWidth  = 960
	transitMapDefaultHeight = 640
)

// TransitPlanet 选择凌日行星 / selects the transiting planet.
type TransitPlanet int

const (
	// TransitMercury 是水星凌日 / Mercury transit.
	TransitMercury TransitPlanet = 1
	// TransitVenus 是金星凌日 / Venus transit.
	TransitVenus TransitPlanet = 2
)

// TransitIsochrones 是等时线开关的位组合 / bit set of contact isochrones.
type TransitIsochrones uint8

const (
	// TransitIsochroneC1 画一触等时线 / draw C1 isochrones.
	TransitIsochroneC1 TransitIsochrones = 1 << iota
	// TransitIsochroneC2 画二触等时线；掠凌没有内切接触时忽略 / draw C2 isochrones, ignored for a grazing transit.
	TransitIsochroneC2
	// TransitIsochroneC3 画三触等时线；掠凌没有内切接触时忽略 / draw C3 isochrones, ignored for a grazing transit.
	TransitIsochroneC3
	// TransitIsochroneC4 画四触等时线 / draw C4 isochrones.
	TransitIsochroneC4
	// TransitIsochroneExternal 是一触与四触 / C1 and C4.
	TransitIsochroneExternal = TransitIsochroneC1 | TransitIsochroneC4
	// TransitIsochroneAll 是全部四个接触 / all four contacts.
	TransitIsochroneAll = TransitIsochroneC1 | TransitIsochroneC2 | TransitIsochroneC3 | TransitIsochroneC4
)

// TransitMapSVGOptions 控制全球凌日可见图。
// TransitMapSVGOptions controls a global planetary-transit visibility map.
type TransitMapSVGOptions struct {
	// Width 和 Height 是 SVG 画布尺寸；宽度小于 640 或高度小于 420 时使用 960x640 默认值。
	// Width and Height are SVG canvas dimensions. Width values below 640 and height values below 420 use the 960x640 defaults.
	Width  int
	Height int
	// Language 为 "en"（不区分大小写）时使用英文，否则使用中文。
	// Language uses English for "en" (case-insensitive) and Chinese otherwise.
	Language string
	// Location 控制显示的事件时刻；nil 使用 date.Location()。
	// Location controls displayed event times. Nil uses date.Location().
	Location *time.Location
	// TimeScale 选择图中时刻的时标：零值 UTC；TimeScaleUT1 要求 Location 为 nil 或 UTC，否则返回 false。
	// TimeScale selects the label scale: the zero value is UTC; TimeScaleUT1 requires Location to be nil or UTC, otherwise the renderer returns false.
	TimeScale astro.TimeScale
	// Projection 选择地图投影；零值使用等经纬投影，正射投影视点取凌甚时的太阳直射点。
	// Projection selects the map projection. The zero value uses equirectangular; orthographic centres on the subsolar point at greatest.
	Projection EclipseMapProjection
	// Isochrones 选择要画的站心接触等时线，可按位组合；零值不画（默认关闭，需显式开启）。
	// Isochrones selects the topocentric contact isochrones to draw and may combine flags. The zero value draws none (off by default).
	Isochrones TransitIsochrones
	// IsochroneStep 是等时线的时间间隔；非正值按行星自动取（水星 20 秒、金星 1 分钟）。
	// IsochroneStep is the isochrone spacing. Non-positive values select 20 s for Mercury and 1 min for Venus.
	IsochroneStep time.Duration
	// 空文本字段使用本地化的自动标签。
	// Empty text fields use localized automatic labels.
	Title      string
	FooterNote string
}

// TransitMapSVG 绘制距 date 最近一次水星或金星凌日的全球可见图：按一触、凌甚、四触太阳中心在视地平上
// （计入标准大气折射）的区域分区；Options.Isochrones 显式开启时叠加所选接触的站心等时线。找不到凌日或选项无效时返回 false。
// TransitMapSVG renders the global visibility of the Mercury or Venus transit closest to date: regions follow where the
// Sun's centre is above the apparent horizon (standard refraction) at C1, greatest and C4. Topocentric isochrones of the
// selected contacts are overlaid only when Options.Isochrones enables them. It returns false when no transit is found
// or the options are invalid.
func TransitMapSVG(planet TransitPlanet, date time.Time, options TransitMapSVGOptions) (string, bool) {
	transit, ok := closestPlanetTransit(planet, date)
	if !ok || !validEclipseMapProjection(options.Projection) {
		return "", false
	}
	if options.TimeScale == astro.TimeScaleUT1 {
		if options.Location != nil && options.Location != time.UTC {
			return "", false
		}
		options.Location = time.UTC
	}
	options = normalizeTransitMapSVGOptions(date, options)
	if !transit.HasInternal {
		options.Isochrones &^= TransitIsochroneC2 | TransitIsochroneC3
	}
	options.Isochrones &= TransitIsochroneAll
	if options.IsochroneStep <= 0 {
		options.IsochroneStep = time.Minute
		if planet == TransitMercury {
			options.IsochroneStep = 20 * time.Second
		}
	}
	projection := internalEclipseMapProjection(options.Projection)
	if projection == "" {
		projection = svgmap.ProjectionEquirectangular
	}
	return renderTransitMapSVG(planet, transit, options, projection), true
}

// closestPlanetTransit 返回距 date 最近的一次地心凌日；行星无效或找不到时第二个返回值为 false。
func closestPlanetTransit(planet TransitPlanet, date time.Time) (basic.PlanetTransitResult, bool) {
	jd := basic.Date2JD(date.UTC())
	var transit basic.PlanetTransitResult
	switch planet {
	case TransitMercury:
		transit = basic.ClosestMercuryTransit(jd)
	case TransitVenus:
		transit = basic.ClosestVenusTransit(jd)
	default:
		return transit, false
	}
	return transit, transit.Valid
}

func normalizeTransitMapSVGOptions(date time.Time, options TransitMapSVGOptions) TransitMapSVGOptions {
	if options.Width < 640 {
		options.Width = transitMapDefaultWidth
	}
	if options.Height < 420 {
		options.Height = transitMapDefaultHeight
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

// transitMapContacts 是地心一触、凌甚、四触的 UTC 时刻。
type transitMapContacts struct {
	start, greatest, end       time.Time
	internalStart, internalEnd time.Time
	hasInternal                bool
}

func transitMapContactsOf(transit basic.PlanetTransitResult) transitMapContacts {
	at := func(jd float64) time.Time { return basic.JD2DateByZone(jd, time.UTC, false) }
	contacts := transitMapContacts{
		start:       at(transit.ExternalIngress),
		greatest:    at(transit.Greatest),
		end:         at(transit.ExternalEgress),
		hasInternal: transit.HasInternal,
	}
	if transit.HasInternal {
		contacts.internalStart, contacts.internalEnd = at(transit.InternalIngress), at(transit.InternalEgress)
	}
	return contacts
}

func renderTransitMapSVG(
	planet TransitPlanet,
	transit basic.PlanetTransitResult,
	options TransitMapSVGOptions,
	projection svgmap.Projection,
) string {
	contacts := transitMapContactsOf(transit)
	// 底部比月食图多留两行：图例下要放可折行的说明与时标声明。
	frame := eclipseMapFrame(options.Width, options.Height, projection, 142, 122)
	if projection == svgmap.ProjectionOrthographic {
		center := solarEclipseSubsolarPoint(contacts.greatest)
		frame.CenterLongitude, frame.CenterLatitude = center.Longitude, center.Latitude
	}
	title := options.Title
	if title == "" {
		date := transitMapDisplayTime(contacts.greatest, options).Format("2006-01-02")
		if options.Language == "en" {
			title = fmt.Sprintf("%s Transit of %s Global Visibility", date, transitMapPlanetName(planet, "en"))
		} else {
			title = fmt.Sprintf("%s %s凌日全球可见图", date, transitMapPlanetName(planet, "zh"))
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s">`,
		options.Width, options.Height, options.Width, options.Height, html.EscapeString(title))
	b.WriteString(`<defs>`)
	writeTransitMapDefinitions(&b, frame, contacts)
	b.WriteString(`</defs>`)
	b.WriteString(`<rect width="100%" height="100%" fill="#efefed"/>`)
	fmt.Fprintf(&b, `<rect x="22" y="18" width="%d" height="%d" fill="#ffffff" stroke="#c9c9c6" stroke-width="1.2"/>`,
		options.Width-44, options.Height-36)
	titleText := title
	if options.Title != "" {
		titleText = svgchart.EllipsizeText(title, float64(options.Width)-52, 24)
	}
	fmt.Fprintf(&b, `<text x="%.3f" y="47" fill="#111111" font-family="Georgia, 'Times New Roman', serif" font-size="24" font-weight="700" text-anchor="middle">%s</text>`,
		float64(options.Width)/2, html.EscapeString(titleText))
	writeTransitMapSummary(&b, transit, contacts, options)

	frame.WriteOcean(&b)
	frame.WriteGraticule(&b, "transit-map-clip")
	frame.WriteLand(&b, "transit-map-clip")
	writeTransitMapRegions(&b, frame, contacts)
	if options.Isochrones != 0 {
		writeTransitMapIsochrones(&b, frame, transit, options)
	}
	frame.WriteFrame(&b)
	legendRows := writeTransitMapLegend(&b, frame, float64(options.Width), options)
	writeTransitMapFooter(&b, frame, options, projection, contacts.greatest, legendRows)
	b.WriteString(`</svg>`)
	return b.String()
}

func transitMapPlanetName(planet TransitPlanet, language string) string {
	if language == "en" {
		if planet == TransitMercury {
			return "Mercury"
		}
		return "Venus"
	}
	if planet == TransitMercury {
		return "水星"
	}
	return "金星"
}

// transitMapDisplayTime 把 UTC 时刻换成图上显示的时刻：UT1 时标给 UT1 读数，否则换到 Location。
func transitMapDisplayTime(value time.Time, options TransitMapSVGOptions) time.Time {
	return transitDisplayTime(value, options.TimeScale, options.Location)
}

// transitMapSunUpPath 返回 value 时刻太阳中心在视地平上的区域面路径（evenodd）与地平边界。
// 视地平计入标准大气折射：太阳中心距直射点小于 90°+34′ 即可见，太阳视差 8.8″ 忽略。
// 半径超过 90° 的球冠环在拆分日界线时分不清内外，所以改画对日点为心、半径不足 90° 的不可见冠，
// 再用图框减去它得到可见区。
func transitMapSunUpPath(value time.Time, frame svgmap.Frame) (string, []svgmap.GeoPoint) {
	subsolar := solarEclipseSubsolarPoint(value)
	antisolar := svgmap.GeoPoint{Longitude: geodata.NormalizeLongitude(subsolar.Longitude + 180), Latitude: -subsolar.Latitude}
	radius := 90 - basic.RefractionFromApparentAltitude(0, transitMapRefractionPressureHPa, transitMapRefractionTemperatureC)
	boundary := geodata.SphericalCircle(antisolar, radius, 360)
	closed := append(append([]svgmap.GeoPoint(nil), boundary...), boundary[0])
	path := fmt.Sprintf("M %.3f %.3f H %.3f V %.3f H %.3f Z %s",
		frame.X, frame.Y, frame.X+frame.Width, frame.Y+frame.Height, frame.X, lunarEclipseGeoPolygonPath(boundary, frame))
	return path, closed
}

// 标准大气，与 basic 站心凌日的视地平判定一致。
const (
	transitMapRefractionPressureHPa  = 1010.0
	transitMapRefractionTemperatureC = 10.0
)

// writeTransitMapDefinitions 输出一触、凌甚、四触三个太阳在地平上半球的形状与排除掩膜，分区口径同月食可见图。
func writeTransitMapDefinitions(b *strings.Builder, frame svgmap.Frame, contacts transitMapContacts) {
	b.WriteString(frame.ClipDefinition("transit-map-clip"))
	startPath, _ := transitMapSunUpPath(contacts.start, frame)
	endPath, _ := transitMapSunUpPath(contacts.end, frame)
	greatestPath, _ := transitMapSunUpPath(contacts.greatest, frame)
	fmt.Fprintf(b, `<path id="transit-sun-c1-shape" fill-rule="evenodd" clip-rule="evenodd" d="%s"/>`, startPath)
	fmt.Fprintf(b, `<path id="transit-sun-c4-shape" fill-rule="evenodd" clip-rule="evenodd" d="%s"/>`, endPath)
	fmt.Fprintf(b, `<path id="transit-sun-greatest-shape" fill-rule="evenodd" clip-rule="evenodd" d="%s"/>`, greatestPath)
	b.WriteString(`<clipPath id="transit-sun-c4"><use href="#transit-sun-c4-shape"/></clipPath>`)
	b.WriteString(`<clipPath id="transit-sun-greatest"><use href="#transit-sun-greatest-shape"/></clipPath>`)
	writeLunarEclipseVisibilityMask(b, "transit-not-visible-mask", frame,
		"transit-sun-c1-shape", "transit-sun-c4-shape", "transit-sun-greatest-shape")
	writeLunarEclipseVisibilityMask(b, "transit-not-c1-c4-mask", frame, "transit-sun-c1-shape", "transit-sun-c4-shape")
	writeLunarEclipseVisibilityMask(b, "transit-not-greatest-mask", frame, "transit-sun-greatest-shape")
	writeLunarEclipseVisibilityMask(b, "transit-not-c4-mask", frame, "transit-sun-c4-shape")
	writeLunarEclipseVisibilityMask(b, "transit-not-c1-mask", frame, "transit-sun-c1-shape")
}

const (
	transitMapColorEntire  = "#5e846d"
	transitMapColorSunrise = "#4e9da0"
	transitMapColorSunset  = "#e2aa4b"
	transitMapColorNone    = "#747b7d"
	transitMapColorC1      = "#a3392f"
	transitMapColorC2      = "#c0702a"
	transitMapColorC3      = "#7a4ca0"
	transitMapColorC4      = "#2f5ea3"
)

// writeTransitMapRegions 着色四类可见区：全程可见、日出时凌日已开始、日落时凌日未结束、不可见。
// 只有凌甚前后太阳在地平上的极区透镜与两端可见、凌甚不可见的高纬带都归入日落带，同月食可见图。
func writeTransitMapRegions(b *strings.Builder, frame svgmap.Frame, contacts transitMapContacts) {
	_, startBoundary := transitMapSunUpPath(contacts.start, frame)
	_, endBoundary := transitMapSunUpPath(contacts.end, frame)
	fmt.Fprintf(b, `<g class="transit-visibility-regions" data-source="%s" clip-path="url(#transit-map-clip)">`, eclipseMapSourceTransitVisibilityRegions)
	if frame.IsPolar() {
		fmt.Fprintf(b, `<circle class="transit-not-visible-region" mask="url(#transit-not-visible-mask)" cx="%.3f" cy="%.3f" r="%.3f" fill="%s" fill-opacity="0.34"/>`,
			frame.X+frame.Width/2, frame.Y+frame.Height/2, frame.Width/2, transitMapColorNone)
	} else {
		fmt.Fprintf(b, `<rect class="transit-not-visible-region" mask="url(#transit-not-visible-mask)" x="%.3f" y="%.3f" width="%.3f" height="%.3f" fill="%s" fill-opacity="0.34"/>`,
			frame.X, frame.Y, frame.Width, frame.Height, transitMapColorNone)
	}
	fmt.Fprintf(b, `<use class="transit-sunset-region" mask="url(#transit-not-c4-mask)" href="#transit-sun-c1-shape" fill="%s" fill-opacity="0.34"/>`, transitMapColorSunset)
	fmt.Fprintf(b, `<use class="transit-sunrise-region" mask="url(#transit-not-c1-mask)" href="#transit-sun-c4-shape" fill="%s" fill-opacity="0.34"/>`, transitMapColorSunrise)
	fmt.Fprintf(b, `<g clip-path="url(#transit-sun-c4)"><g clip-path="url(#transit-sun-greatest)"><use class="transit-entire-region" href="#transit-sun-c1-shape" fill="%s" fill-opacity="0.34"/></g></g>`, transitMapColorEntire)
	fmt.Fprintf(b, `<g clip-path="url(#transit-sun-c4)"><g mask="url(#transit-not-greatest-mask)"><use class="transit-sunset-region" href="#transit-sun-c1-shape" fill="%s" fill-opacity="0.34"/></g></g>`, transitMapColorSunset)
	fmt.Fprintf(b, `<g mask="url(#transit-not-c1-c4-mask)"><use class="transit-sunset-region" href="#transit-sun-greatest-shape" fill="%s" fill-opacity="0.34"/></g>`, transitMapColorSunset)
	b.WriteString(`</g>`)
	writeEclipseMapGeoLine(b, frame, startBoundary, "c1-horizon", "#a56c16", 1.2, "4 3", "transit-map-clip", eclipseMapSourceTransitHorizonBoundaries)
	writeEclipseMapGeoLine(b, frame, endBoundary, "c4-horizon", "#197a82", 1.2, "4 3", "transit-map-clip", eclipseMapSourceTransitHorizonBoundaries)
}

// writeTransitMapIsochrones 画选中接触的站心等时线：每条线是同一时刻看到该接触的地点，只保留该时刻太阳在视地平上的部分。
// 每隔若干条标一次时刻，标签按 LabelTable 避让。
func writeTransitMapIsochrones(b *strings.Builder, frame svgmap.Frame, transit basic.PlanetTransitResult, options TransitMapSVGOptions) {
	table := &svgchart.LabelTable{}
	format := "15:04"
	if options.IsochroneStep%time.Minute != 0 {
		format = "15:04:05"
	}
	for _, contact := range []struct {
		kind  basic.PlanetTransitContactKind
		class string
		color string
		flag  TransitIsochrones
	}{
		{basic.PlanetTransitC1, "c1-isochrone", transitMapColorC1, TransitIsochroneC1},
		{basic.PlanetTransitC2, "c2-isochrone", transitMapColorC2, TransitIsochroneC2},
		{basic.PlanetTransitC3, "c3-isochrone", transitMapColorC3, TransitIsochroneC3},
		{basic.PlanetTransitC4, "c4-isochrone", transitMapColorC4, TransitIsochroneC4},
	} {
		if options.Isochrones&contact.flag == 0 {
			continue
		}
		lines := basic.PlanetTransitContactIsochrones(transit, contact.kind, options.IsochroneStep)
		labelEvery := transitMapIsochroneLabelEvery(len(lines))
		for index, line := range lines {
			segments := make([][]svgmap.GeoPoint, 0, len(line.Segments))
			for _, segment := range line.Segments {
				points := make([]svgmap.GeoPoint, len(segment))
				for position, point := range segment {
					points[position] = svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
				}
				segments = append(segments, points)
				writeEclipseMapGeoLine(b, frame, points, contact.class, contact.color, 0.9, "none", "transit-map-clip", eclipseMapSourceTransitContactIsochrones)
			}
			if index%labelEvery != 0 {
				continue
			}
			text := transitMapDisplayTime(basic.JD2DateByZone(line.JD, time.UTC, false), options).Format(format)
			placement, ok := table.Place(text, 10, transitMapIsochroneLabelCandidates(segments, frame))
			if !ok {
				continue
			}
			fmt.Fprintf(b, `<text class="%s-label" x="%.3f" y="%.3f" fill="%s" stroke="#ffffff" stroke-width="2.6" paint-order="stroke" font-family="Arial, sans-serif" font-size="10" text-anchor="%s">%s</text>`,
				contact.class, placement.X, placement.Y, contact.color, placement.Anchor, html.EscapeString(text))
		}
	}
}

// transitMapIsochroneLabelEvery 让每个接触最多标八条线。
func transitMapIsochroneLabelEvery(count int) int {
	every := int(math.Ceil(float64(count) / 8))
	if every < 1 {
		every = 1
	}
	return every
}

// transitMapIsochroneLabelCandidates 从最长支路的中点向两端取画布内的候选位置。
func transitMapIsochroneLabelCandidates(segments [][]svgmap.GeoPoint, frame svgmap.Frame) []svgchart.LabelPlacement {
	longest := -1
	for index, segment := range segments {
		if longest < 0 || len(segment) > len(segments[longest]) {
			longest = index
		}
	}
	if longest < 0 {
		return nil
	}
	segment := segments[longest]
	middle := len(segment) / 2
	candidates := make([]svgchart.LabelPlacement, 0, 32)
	for offset := 0; offset <= middle && len(candidates) < 32; offset++ {
		for _, index := range []int{middle + offset, middle - offset} {
			if index < 0 || index >= len(segment) || (offset == 0 && index != middle) {
				continue
			}
			x, y, visible := frame.Project(segment[index].Longitude, segment[index].Latitude)
			if !visible || x < frame.X+24 || x > frame.X+frame.Width-24 || y < frame.Y+14 || y > frame.Y+frame.Height-6 {
				continue
			}
			candidates = append(candidates, svgchart.LabelPlacement{X: x, Y: y - 4, Anchor: "middle"})
		}
	}
	return candidates
}

func writeTransitMapSummary(b *strings.Builder, transit basic.PlanetTransitResult, contacts transitMapContacts, options TransitMapSVGOptions) {
	clock := func(value time.Time) string { return transitMapDisplayTime(value, options).Format("15:04:05") }
	zone, _ := transitMapDisplayTime(contacts.greatest, options).Zone()
	if zone == "" {
		zone = "UTC"
	}
	if options.TimeScale == astro.TimeScaleUT1 {
		zone = "UT1"
	}
	greatest, separation := "凌甚", "最小角距"
	if options.Language == "en" {
		greatest, separation = "Greatest", "least separation"
	}
	parts := []string{"C1 " + clock(contacts.start)}
	if contacts.hasInternal {
		parts = append(parts, "C2 "+clock(contacts.internalStart))
	}
	parts = append(parts, greatest+" "+clock(contacts.greatest))
	if contacts.hasInternal {
		parts = append(parts, "C3 "+clock(contacts.internalEnd))
	}
	parts = append(parts, "C4 "+clock(contacts.end))
	line := fmt.Sprintf("%s (%s)", strings.Join(parts, " | "), zone)
	fmt.Fprintf(b, `<text x="%.3f" y="82" fill="#293235" font-family="Arial, sans-serif" font-size="13" text-anchor="middle">%s</text>`,
		float64(options.Width)/2, html.EscapeString(svgchart.EllipsizeText(line, float64(options.Width)-60, 13)))
	note := fmt.Sprintf("%s %.1f″ | 太阳视半径 %.1f″ | 行星视半径 %.2f″（地心）", separation,
		transit.MinimumSeparationArcsec, transit.SunSemidiameterArcsec, transit.PlanetSemidiameterArcsec)
	if options.Language == "en" {
		note = fmt.Sprintf("%s %.1f″ | solar semidiameter %.1f″ | planet semidiameter %.2f″ (geocentric)", separation,
			transit.MinimumSeparationArcsec, transit.SunSemidiameterArcsec, transit.PlanetSemidiameterArcsec)
	}
	fmt.Fprintf(b, `<text x="%.3f" y="100" fill="#293235" font-family="Arial, sans-serif" font-size="13" text-anchor="middle">%s</text>`,
		float64(options.Width)/2, html.EscapeString(note))
}

// writeTransitMapLegend 画区域色块与等时线图例，按画布宽度折行，返回行数。
func writeTransitMapLegend(b *strings.Builder, frame svgmap.Frame, canvasWidth float64, options TransitMapSVGOptions) int {
	type entry struct {
		text, color string
		line        bool
	}
	labels := []string{"全程可见", "日出时已在凌日", "日落时仍在凌日", "不可见", "一触等时线", "二触等时线", "三触等时线", "四触等时线"}
	if options.Language == "en" {
		labels = []string{"Entire transit", "Transit at sunrise", "Transit at sunset", "Not visible",
			"C1 isochrones", "C2 isochrones", "C3 isochrones", "C4 isochrones"}
	}
	entries := []entry{
		{labels[0], transitMapColorEntire, false}, {labels[1], transitMapColorSunrise, false},
		{labels[2], transitMapColorSunset, false}, {labels[3], transitMapColorNone, false},
	}
	for index, item := range []struct {
		flag  TransitIsochrones
		color string
	}{
		{TransitIsochroneC1, transitMapColorC1}, {TransitIsochroneC2, transitMapColorC2},
		{TransitIsochroneC3, transitMapColorC3}, {TransitIsochroneC4, transitMapColorC4},
	} {
		if options.Isochrones&item.flag != 0 {
			entries = append(entries, entry{labels[4+index], item.color, true})
		}
	}
	width := func(item entry) float64 {
		return lunarEclipseDetailedMapLegendIconWidth + svgchart.EstimatedTextWidth(item.text, lunarEclipseDetailedMapLegendFontSize)
	}
	available := canvasWidth - 2*lunarEclipseDetailedPageInset
	var rows [][]entry
	used := 0.0
	for _, item := range entries {
		if len(rows) == 0 || used+lunarEclipseDetailedMapLegendGap+width(item) > available {
			rows = append(rows, nil)
			used = -lunarEclipseDetailedMapLegendGap
		}
		rows[len(rows)-1] = append(rows[len(rows)-1], item)
		used += lunarEclipseDetailedMapLegendGap + width(item)
	}
	baseY := frame.Y + frame.Height + 31
	for rowIndex, row := range rows {
		total := -lunarEclipseDetailedMapLegendGap
		for _, item := range row {
			total += lunarEclipseDetailedMapLegendGap + width(item)
		}
		x := math.Max(lunarEclipseDetailedPageInset, frame.X+frame.Width/2-total/2)
		y := baseY + float64(rowIndex)*lunarEclipseDetailedMapLegendLineStep
		for _, item := range row {
			if item.line {
				fmt.Fprintf(b, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="%s" stroke-width="1.4"/>`, x, y-4, x+18, y-4, item.color)
			} else {
				fmt.Fprintf(b, `<rect x="%.3f" y="%.3f" width="18" height="9" fill="%s" fill-opacity="0.78"/>`, x, y-8, item.color)
			}
			fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#465053" font-family="Arial, sans-serif" font-size="%.0f">%s</text>`,
				x+lunarEclipseDetailedMapLegendIconWidth, y, lunarEclipseDetailedMapLegendFontSize, html.EscapeString(item.text))
			x += width(item) + lunarEclipseDetailedMapLegendGap
		}
	}
	return len(rows)
}

func writeTransitMapFooter(
	b *strings.Builder,
	frame svgmap.Frame,
	options TransitMapSVGOptions,
	projection svgmap.Projection,
	instant time.Time,
	legendRows int,
) {
	text := options.FooterNote
	if text == "" {
		if options.Language == "en" {
			text = eclipseMapProjectionLabel(projection, "en") + "; regions follow whether the Sun's centre is above the apparent horizon (standard refraction) at C1, greatest and C4"
			if options.Isochrones != 0 {
				text += "; isochrones are topocentric contact times"
			}
			text += "."
		} else {
			text = eclipseMapProjectionLabel(projection, "zh") + "；按一触、凌甚、四触太阳中心是否在视地平上分区（计入标准大气折射）"
			if options.Isochrones != 0 {
				text += "；等时线为站心接触时刻"
			}
			text += "。"
		}
	}
	const footerBottomPadding = 26.0
	// 页脚紧跟图例末行，不与图例重叠；行数按画布底边截断，时标声明独占最后一行。

	baseline := frame.Y + frame.Height + 31 + float64(legendRows-1)*lunarEclipseDetailedMapLegendLineStep + 18
	maxWidth := float64(options.Width) - frame.X - 24
	lines := svgchart.WrapText(text, maxWidth, 11)
	lines = svgchart.FooterLinesWithScale(lines, timenote.Scale(options.TimeScale, instant, options.Location, options.Language), maxWidth, 11,
		svgchart.BaselineLineLimit(11, svgchart.FooterLineHeight(11), baseline, float64(options.Height)-footerBottomPadding))
	for index, line := range lines {
		fmt.Fprintf(b, `<text x="%.3f" y="%.3f" fill="#596164" font-family="Georgia, 'Times New Roman', serif" font-size="11">%s</text>`,
			frame.X, baseline+float64(index)*svgchart.FooterLineHeight(11), html.EscapeString(line))
	}
}
