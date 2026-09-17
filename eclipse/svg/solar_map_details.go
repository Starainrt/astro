package svg

import (
	"fmt"
	"html"
	"math"
	"sort"
	"strings"
	"time"

	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/internal/svgmap"
)

type solarEclipseMapLayout struct {
	frame       svgmap.Frame
	panelX      float64
	panelY      float64
	panelWidth  float64
	panelHeight float64
	// nasa 为 true 时按 NASA 摆法排版：球面居中放大，阶段面板移到球面下方分三栏，图例再往下。
	// nasa switches to the NASA composition: a centred globe, phase panels in three columns
	// below it, and the legend under those.
	nasa         bool
	legendY      float64
	scaleY       float64
	margin       float64
	panelGap     float64
	blockY       float64
	blockHeight  float64
	secondPanelY float64
	// landscape 为真时用横版排布：地图在左，数据块两栏三行在右，地图下方那条放天平动与比例尺。
	landscape      bool
	gridX          float64
	gridCellWidth  float64
	gridCellHeight float64
	gridGapX       float64
	gridGapY       float64
	// 三行数据块的行高按各行行数分配：6 行块与 4 行块等分会把行距压到文字高度以下。
	gridRowY       [solarEclipseLandscapeRows]float64
	gridRowHeights [solarEclipseLandscapeRows]float64
	bottomY        float64
	// stripHeight 是地图下方那条里天平动盒的高度；scaleSlot 是比例尺槽，两者不重叠。
	stripHeight float64
	scaleSlot   svgchart.LabelBox
}

// solarEclipsePanelBlock 是一个带标题的数据块，竖版与横版共用同一批内容。
type solarEclipsePanelBlock struct {
	title string
	rows  []svgchart.PanelRow
}

// solarEclipsePanelBlocks 汇总日食详细版式的全部数据块，两种版式只是排布不同。
type solarEclipsePanelBlocks struct {
	sun, moon                solarEclipsePanelBlock
	penumbra, umbra          solarEclipsePanelBlock
	circumstances, ephemeris solarEclipsePanelBlock
	libration                solarEclipsePanelBlock
}

// solarEclipseMapLayoutFor 选详细版式：按画布朝向定排布，按投影定地图长宽比。
// 横版把数据块放到地图右侧，竖版放到地图下方；两者内容完全一致。
func solarEclipseMapLayoutFor(
	options SolarEclipseMapSVGOptions,
	projection svgmap.Projection,
	center svgmap.GeoPoint,
) solarEclipseMapLayout {
	// 只给几何关系的调用方用两行图例与四行、四行、六行的数据块行数。
	return solarEclipseMapLayoutForBlocks(options, projection, center, 2, solarEclipseLandscapeDefaultRowFields)
}

// solarEclipseMapLayoutForBlocks 按图例行数与数据块行数排版：两者决定横带上下限，必须先算出来。
func solarEclipseMapLayoutForBlocks(
	options SolarEclipseMapSVGOptions,
	projection svgmap.Projection,
	center svgmap.GeoPoint,
	legendRows int,
	rowFields [solarEclipseLandscapeRows]float64,
) solarEclipseMapLayout {
	width := float64(options.Width)
	height := float64(options.Height)
	if width >= height {
		return solarEclipseLandscapeMapLayout(width, height, projection, center, legendRows, rowFields)
	}
	return solarEclipsePortraitMapLayout(width, height, projection, center, legendRows)
}

func writeSolarEclipseMapSectionTitle(
	builder *strings.Builder,
	labels *svgchart.LabelTable,
	layout solarEclipseMapLayout,
	options SolarEclipseMapSVGOptions,
	hasCentral bool,
) {
	label := options.MapTitle
	if label == "" {
		if options.Language == "en" && hasCentral {
			label = "Global visibility and central path"
		} else if options.Language == "en" {
			label = "Global visibility"
		} else if hasCentral {
			label = "全球见食范围与中心食带"
		} else {
			label = "全球见食范围"
		}
	} else {
		label = svgchart.EllipsizeText(label, layout.frame.Width-8, 14)
	}
	anchor, x, y := "start", layout.frame.X, layout.frame.Y-10
	if layout.nasa {
		// 让开球面顶端的 N 标记。
		anchor, x, y = "middle", layout.frame.X+layout.frame.Width/2, layout.frame.Y-28
	}
	// 图框上沿是标题带：表头文字已经占位，这里只在这条带里选位置。
	placement, ok := labels.Place(label, 14, []svgchart.LabelPlacement{
		{X: x, Y: y, Anchor: anchor},
		{X: x, Y: layout.frame.Y - 10, Anchor: anchor},
		{X: layout.frame.X, Y: layout.frame.Y - 10, Anchor: "start"},
		{X: layout.frame.X, Y: layout.frame.Y - 28, Anchor: "start"},
	})
	if !ok {
		return
	}
	fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#161a1b" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700" text-anchor="%s">%s</text>`,
		placement.X, placement.Y, placement.Anchor, html.EscapeString(label))
}

func writeSolarEclipseRiseSetCurves(
	builder *strings.Builder,
	curves []eclipsecore.SolarEclipseRiseSetCurve,
	frame svgmap.Frame,
) {
	for _, curve := range curves {
		className := fmt.Sprintf("solar-rise-set-boundary solar-%s-%s",
			html.EscapeString(string(curve.Phase)), html.EscapeString(string(curve.Direction)))
		for _, segment := range curve.Segments {
			points := make([]svgmap.GeoPoint, len(segment))
			for index, point := range segment {
				points[index] = svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			writeEclipseMapGeoLine(
				builder, frame, points, className, "#d97706", 1.35, "6 4", "solar-map-clip",
				eclipseMapSourceRiseSetPhaseLines,
			)
		}
	}
}

func writeSolarEclipsePenumbralOutlines(
	builder *strings.Builder,
	info eclipsecore.SolarEclipsePartialFootprintsInfo,
	frame svgmap.Frame,
	options SolarEclipseMapSVGOptions,
	labels *svgchart.LabelTable,
) {
	if options.PenumbralOutlineStep <= 0 {
		return
	}
	selected := solarEclipseFootprintsAtStep(
		info.Footprints,
		options.PenumbralOutlineStep,
		options.Location,
		info.Eclipse.GreatestEclipse,
	)
	for _, footprint := range selected {
		writeSolarEclipseFootprintBoundary(
			builder, footprint, frame, "solar-penumbral-outline", "#b07a18", 0.75, "3 3",
			eclipseMapSourcePenumbralOutlines,
		)
		if mapTimeDistance(footprint.Time, info.Eclipse.GreatestEclipse) <= info.Step/2 {
			continue
		}
		x, y, ok := solarEclipseFootprintLabelPosition(footprint, frame)
		if !ok {
			continue
		}
		labelTime := solarEclipseMapAlignedTime(footprint.Time, options.PenumbralOutlineStep, options.Location)
		text := labelTime.Format("15:04")
		placed, ok := labels.Place(text, 8, labelCandidatesInsideBox(
			svgchart.LabelCandidates(x, y-4, "middle", 11), text, 8, solarEclipseFrameBox(frame)))
		if !ok {
			continue
		}
		fmt.Fprintf(builder, `<text class="solar-penumbral-time-label" x="%.3f" y="%.3f" fill="#8b5b08" stroke="#ffffff" stroke-width="2.4" paint-order="stroke" font-family="Arial, sans-serif" font-size="8" font-weight="700" text-anchor="%s">%s</text>`,
			placed.X, placed.Y, placed.Anchor, html.EscapeString(text))
	}
}

func writeSolarEclipseCentralShadowOutlines(
	builder *strings.Builder,
	footprints []eclipsecore.SolarEclipsePartialFootprint,
	frame svgmap.Frame,
) {
	for _, footprint := range footprints {
		writeSolarEclipseFootprintBoundary(
			builder, footprint, frame, "solar-central-shadow-outline", "#7b5a42", 0.8, "",
			eclipseMapSourceCentralShadowOutlines,
		)
	}
}

func writeSolarEclipseCentralShadowSweep(
	builder *strings.Builder,
	footprints []eclipsecore.SolarEclipsePartialFootprint,
	frame svgmap.Frame,
	eclipseType eclipsecore.SolarEclipseType,
) {
	if len(footprints) == 0 {
		return
	}
	var path strings.Builder
	hasOpen := false
	samples := solarEclipseCentralShadowSweepSamples(footprints)
	for _, footprint := range footprints {
		hasOpen = hasOpen || !footprint.Closed
	}
	if hasOpen {
		polygons, err := geodata.OpenBoundarySweep(samples)
		if err == nil {
			for _, polygon := range polygons {
				if len(polygon) < 3 {
					continue
				}
				for _, fragment := range svgmap.PolygonFragments(polygon, frame.Clip()) {
					if len(fragment) < 3 {
						continue
					}
					path.WriteString(`<path d="`)
					appendEclipseMapPolygonPathConsistent(&path, frame, fragment)
					path.WriteString(`"/>`)
				}
			}
		}
	}
	if !hasOpen {
		for _, footprint := range footprints {
			segments := make([][]svgmap.GeoPoint, 0, len(footprint.Boundaries))
			for _, source := range footprint.Boundaries {
				segment := make([]svgmap.GeoPoint, len(source))
				for index, point := range source {
					segment[index] = svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
				}
				segments = append(segments, segment)
			}
			boundary := svgmap.JoinPolylineSegments(segments)
			if len(boundary) > 1 && svgmap.SameGeoPoint(boundary[0], boundary[len(boundary)-1]) {
				boundary = boundary[:len(boundary)-1]
			}
			if len(boundary) < 3 {
				continue
			}
			for _, fragment := range svgmap.PolygonFragments(boundary, frame.Clip()) {
				if len(fragment) < 3 {
					continue
				}
				path.WriteString(`<path d="`)
				appendEclipseMapPolygonPathConsistent(&path, frame, fragment)
				path.WriteString(`"/>`)
			}
		}
	}
	if path.Len() == 0 {
		return
	}
	color := solarEclipseCentralPathColor(eclipseType)
	fmt.Fprintf(builder, `<g class="solar-central-shadow-sweep" data-source="%s" clip-path="url(#solar-map-clip)" fill="%s" fill-opacity="0.18" fill-rule="nonzero">%s</g>`, eclipseMapSourceSampledOpenSweep, color, path.String())
}

func solarEclipseMonotoneCentralShadowSweepPolygons(
	footprints []eclipsecore.SolarEclipsePartialFootprint,
) ([][]geodata.GeoPoint, error) {
	samples := solarEclipseCentralShadowSweepSamples(footprints)
	polygons, err := geodata.MonotoneOpenBoundarySweep(samples)
	if err != nil {
		return geodata.OpenBoundarySweep(
			geodata.DecimateOpenBoundarySweepSamples(samples, 24, 40),
		)
	}
	return polygons, nil
}

func solarEclipseCentralBandInnerTransitionCaps(
	footprints []eclipsecore.SolarEclipsePartialFootprint,
) [][]geodata.GeoPoint {
	samples := solarEclipseCentralShadowSweepSamples(footprints)
	samples = geodata.DecimateOpenBoundarySweepSamples(samples, len(samples), 40)
	return geodata.OpenBoundarySweepInnerCaps(samples, 500)
}

func solarEclipseCentralShadowSweepSamples(
	footprints []eclipsecore.SolarEclipsePartialFootprint,
) []geodata.OpenBoundarySweepSample {
	samples := make([]geodata.OpenBoundarySweepSample, 0, len(footprints))
	for _, footprint := range footprints {
		boundaries := make([][]geodata.GeoPoint, 0, len(footprint.Boundaries))
		for _, source := range footprint.Boundaries {
			segment := make([]geodata.GeoPoint, len(source))
			for index, point := range source {
				segment[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			boundaries = append(boundaries, segment)
		}
		samples = append(samples, geodata.OpenBoundarySweepSample{
			Boundaries: boundaries,
			Closed:     footprint.Closed,
		})
	}
	return samples
}

func writeSolarEclipseCentralBandEnvelope(
	builder *strings.Builder,
	segments [][]eclipsecore.SolarEclipsePathPoint,
	frame svgmap.Frame,
	eclipseType eclipsecore.SolarEclipseType,
	hasCentral bool,
) bool {
	if len(segments) == 0 {
		return false
	}
	polygons := make([][]geodata.GeoPoint, 0, len(segments))
	for _, segment := range segments {
		if len(segment) < 4 {
			return false
		}
		polygon := make([]geodata.GeoPoint, len(segment))
		for index, point := range segment {
			polygon[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		polygons = append(polygons, polygon)
	}
	merged, err := geodata.UnionPolygons(polygons)
	if err != nil {
		return false
	}
	var path strings.Builder
	for _, polygon := range merged {
		for _, fragment := range svgmap.PolygonFragments(polygon, frame.Clip()) {
			if len(fragment) < 3 {
				continue
			}
			path.WriteString(`<path d="`)
			appendEclipseMapPolygonPathConsistent(&path, frame, fragment)
			path.WriteString(`"/>`)
		}
	}
	if path.Len() == 0 {
		return false
	}
	color := solarEclipseCentralPathColor(eclipseType)
	// 非中心食的包络同样是从 CentralBandSegments 来的真实食带，只是没有中心线；
	// 用更淡的填充会让人以为没画，两者共用同一透明度，仅类名不同以便区分。
	class, opacity := "solar-central-shadow-sweep", 0.30
	if hasCentral {
		class = "central-eclipse-band"
	}
	fmt.Fprintf(builder, `<g class="%s" data-source="%s" clip-path="url(#solar-map-clip)" fill="%s" fill-opacity="%.2f" fill-rule="nonzero">%s</g>`, class, eclipseMapSourceBesselianEnvelope, color, opacity, path.String())
	return true
}

func solarEclipseFootprintLabelPosition(
	footprint eclipsecore.SolarEclipsePartialFootprint,
	frame svgmap.Frame,
) (float64, float64, bool) {
	bestX, bestY, bestScore := 0.0, 0.0, math.Inf(1)
	centerX := frame.X + frame.Width/2
	for _, boundary := range footprint.Boundaries {
		for _, point := range boundary {
			x, y, visible := frame.Project(point.Longitude, point.Latitude)
			if !visible || x < frame.X+24 || x > frame.X+frame.Width-24 ||
				y < frame.Y+14 || y > frame.Y+frame.Height-14 {
				continue
			}
			score := y + 0.05*math.Abs(x-centerX)
			if score < bestScore {
				bestX, bestY, bestScore = x, y, score
			}
		}
	}
	return bestX, bestY, !math.IsInf(bestScore, 1)
}

func solarEclipseMapAlignedTime(value time.Time, step time.Duration, location *time.Location) time.Time {
	local := value.In(location)
	dayStart := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location)
	elapsed := local.Sub(dayStart)
	return dayStart.Add(((elapsed + step/2) / step) * step)
}

func writeSolarEclipseFootprintBoundary(
	builder *strings.Builder,
	footprint eclipsecore.SolarEclipsePartialFootprint,
	frame svgmap.Frame,
	className, color string,
	strokeWidth float64,
	dash, dataSource string,
) {
	for _, boundary := range footprint.Boundaries {
		points := make([]svgmap.GeoPoint, len(boundary))
		for index, point := range boundary {
			points[index] = svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		writeEclipseMapGeoLine(builder, frame, points, className, color, strokeWidth, dash, "solar-map-clip", dataSource)
	}
}

func solarEclipseFootprintsAtStep(
	footprints []eclipsecore.SolarEclipsePartialFootprint,
	step time.Duration,
	location *time.Location,
	include time.Time,
) []eclipsecore.SolarEclipsePartialFootprint {
	if len(footprints) == 0 || step <= 0 {
		return nil
	}
	targets := make([]time.Time, 0)
	for value := firstMapTimeLabelAfter(footprints[0].Time, step, location); !value.After(footprints[len(footprints)-1].Time); value = value.Add(step) {
		targets = append(targets, value)
	}
	if !include.IsZero() {
		targets = append(targets, include)
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].Before(targets[j]) })

	selected := make([]eclipsecore.SolarEclipsePartialFootprint, 0, len(targets))
	index := 0
	for _, target := range targets {
		for index+1 < len(footprints) &&
			mapTimeDistance(footprints[index+1].Time, target) < mapTimeDistance(footprints[index].Time, target) {
			index++
		}
		candidate := footprints[index]
		if len(selected) == 0 || !selected[len(selected)-1].Time.Equal(candidate.Time) {
			selected = append(selected, candidate)
		}
	}
	return selected
}

func mapTimeDistance(a, b time.Time) time.Duration {
	value := a.Sub(b)
	if value < 0 {
		return -value
	}
	return value
}

func writeSolarEclipseContactMarkers(builder *strings.Builder, contacts []solarEclipseContactLabel) {
	for _, contact := range contacts {
		fmt.Fprintf(builder, `<g class="solar-shadow-contact solar-contact-%s"><circle cx="%.3f" cy="%.3f" r="2.7" fill="%s" stroke="#ffffff" stroke-width="0.9"/>`,
			strings.ToLower(contact.marker.name), contact.x, contact.y, contact.marker.color)
		if contact.ok {
			fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="%s" stroke="#ffffff" stroke-width="2.4" paint-order="stroke" font-family="Arial, sans-serif" font-size="8" font-weight="700" text-anchor="%s">%s</text>`,
				contact.placed.X, contact.placed.Y, contact.marker.color, contact.placed.Anchor, html.EscapeString(contact.text))
		}
		builder.WriteString(`</g>`)
	}
}

func writeSolarEclipseAxisMarkers(
	builder *strings.Builder,
	path eclipsecore.SolarEclipsePath,
	frame svgmap.Frame,
	options SolarEclipseMapSVGOptions,
) {
	if len(path.CenterLine) < 2 {
		return
	}
	points := []eclipsecore.SolarEclipsePathPoint{path.CenterLine[0], path.CenterLine[len(path.CenterLine)-1]}
	for index, point := range points {
		x, y, visible := frame.Project(point.Longitude, point.Latitude)
		if !visible {
			continue
		}
		label := "中心线始"
		if index == 1 {
			label = "中心线终"
		}
		if options.Language == "en" {
			label = "Axis enters"
			if index == 1 {
				label = "Axis exits"
			}
		}
		fmt.Fprintf(builder, `<g class="solar-axis-contact" aria-label="%s"><title>%s</title><rect x="%.3f" y="%.3f" width="5" height="5" fill="#263f58" stroke="#ffffff" stroke-width="0.8"/></g>`,
			html.EscapeString(label), html.EscapeString(label), x-2.5, y-2.5)
	}
}

func writeSolarEclipseSubsolarMarker(builder *strings.Builder, label solarEclipseMapPointLabel) {
	if label.text == "" {
		return
	}
	x, y := label.x, label.y
	fmt.Fprintf(builder, `<g class="solar-subsolar-marker"><circle cx="%.3f" cy="%.3f" r="4" fill="#f4c542" stroke="#714f00" stroke-width="1"/><path d="M %.3f %.3f h 8 M %.3f %.3f v 8" fill="none" stroke="#714f00" stroke-width="1"/>`,
		x, y, x-4, y, x, y-4)
	if label.ok {
		fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#714f00" stroke="#ffffff" stroke-width="2.5" paint-order="stroke" font-family="Arial, sans-serif" font-size="7" font-weight="700" text-anchor="%s" opacity="0.72">%s</text>`,
			label.placed.X, label.placed.Y, label.placed.Anchor, html.EscapeString(label.text))
	}
	builder.WriteString(`</g>`)
}

func solarEclipseFormatCoordinates(longitude, latitude float64) string {
	lonSuffix := "E"
	if longitude < 0 {
		lonSuffix = "W"
	}
	latSuffix := "N"
	if latitude < 0 {
		latSuffix = "S"
	}
	return fmt.Sprintf("%.4f°%s, %.4f°%s", math.Abs(longitude), lonSuffix, math.Abs(latitude), latSuffix)
}

func writeSolarEclipseTimeMarkers(builder *strings.Builder, labels []solarEclipseMapPointLabel) {
	for _, label := range labels {
		fmt.Fprintf(builder, `<g class="solar-time-marker"><circle cx="%.3f" cy="%.3f" r="2.3" fill="#263f58" stroke="#ffffff" stroke-width="1"/>`,
			label.x, label.y)
		if label.ok {
			fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#263f58" stroke="#ffffff" stroke-width="3" paint-order="stroke" font-family="Arial, sans-serif" font-size="9" font-weight="700" text-anchor="%s">%s</text>`,
				label.placed.X, label.placed.Y, label.placed.Anchor, html.EscapeString(label.text))
		}
		builder.WriteString(`</g>`)
	}
}

func solarEclipseGreatestTimeLabelWindow(step time.Duration) time.Duration {
	window := step / 3
	if window < 10*time.Minute {
		return 10 * time.Minute
	}
	return window
}

func solarEclipseTimeMarkerPoints(
	points []eclipsecore.SolarEclipsePathPoint,
	step time.Duration,
	location *time.Location,
	excluded []time.Time,
) []eclipsecore.SolarEclipsePathPoint {
	if len(points) < 2 || step <= 0 {
		return nil
	}
	start := points[0].Time
	end := points[len(points)-1].Time
	current := firstMapTimeLabelAfter(start, step, location)
	window := step / 4
	if window > 5*time.Minute {
		window = 5 * time.Minute
	}
	if window < 30*time.Second {
		window = 30 * time.Second
	}
	result := make([]eclipsecore.SolarEclipsePathPoint, 0)
	segment := 1
	for current.Before(end) {
		for segment < len(points) && points[segment].Time.Before(current) {
			segment++
		}
		if segment >= len(points) {
			break
		}
		if !mapTimeNearAny(current, excluded, window) {
			a, b := points[segment-1], points[segment]
			span := b.Time.Sub(a.Time)
			if span > 0 {
				fraction := float64(current.Sub(a.Time)) / float64(span)
				result = append(result, interpolateSolarEclipsePathPoint(a, b, fraction, current))
			}
		}
		current = current.Add(step)
	}
	return result
}

func interpolateSolarEclipsePathPoint(
	a, b eclipsecore.SolarEclipsePathPoint,
	fraction float64,
	value time.Time,
) eclipsecore.SolarEclipsePathPoint {
	deltaLongitude := b.Longitude - a.Longitude
	if deltaLongitude > 180 {
		deltaLongitude -= 360
	} else if deltaLongitude < -180 {
		deltaLongitude += 360
	}
	longitude := a.Longitude + fraction*deltaLongitude
	if longitude > 180 {
		longitude -= 360
	} else if longitude < -180 {
		longitude += 360
	}
	return eclipsecore.SolarEclipsePathPoint{
		Time:        value,
		Longitude:   longitude,
		Latitude:    a.Latitude + fraction*(b.Latitude-a.Latitude),
		SunAltitude: a.SunAltitude + fraction*(b.SunAltitude-a.SunAltitude),
		WidthKM:     a.WidthKM + fraction*(b.WidthKM-a.WidthKM),
	}
}

func firstMapTimeLabelAfter(value time.Time, step time.Duration, location *time.Location) time.Time {
	local := value.In(location)
	dayStart := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location)
	elapsed := local.Sub(dayStart)
	return dayStart.Add((elapsed/step + 1) * step)
}

func mapTimeNearAny(value time.Time, excluded []time.Time, window time.Duration) bool {
	for _, candidate := range excluded {
		if candidate.IsZero() {
			continue
		}
		delta := value.Sub(candidate)
		if delta < 0 {
			delta = -delta
		}
		if delta <= window {
			return true
		}
	}
	return false
}

func mapTimesNear(a, b time.Time, window time.Duration) bool {
	delta := a.Sub(b)
	if delta < 0 {
		delta = -delta
	}
	return delta <= window
}

// writeSolarEclipseGreatestTimeContours 绘制地方食甚时刻等时线及其 HH:MM 标注。
func writeSolarEclipseGreatestTimeContours(
	builder *strings.Builder,
	contours []eclipsecore.SolarEclipseGreatestTimeContour,
	frame svgmap.Frame,
	options SolarEclipseMapSVGOptions,
	labels *svgchart.LabelTable,
) {
	if len(contours) == 0 {
		return
	}
	for _, contour := range contours {
		for _, segment := range contour.Segments {
			points := make([]svgmap.GeoPoint, len(segment))
			for index, point := range segment {
				points[index] = svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			writeEclipseMapGeoLine(
				builder, frame, points, "solar-greatest-time-isoline", "#1f6fb2", 1.0, "", "solar-map-clip",
				eclipseMapSourceGreatestTimeIsochrones,
			)
		}
		text := contour.Time.In(options.Location).Format("15:04")
		candidates := solarEclipseContourLabelCandidates(contour.Segments, frame, 0, 0, false)
		placed, ok := labels.Place(text, 9, labelCandidatesInsideBox(candidates, text, 9, solarEclipseFrameBox(frame)))
		if !ok {
			continue
		}
		fmt.Fprintf(builder, `<text class="solar-greatest-time-label" x="%.3f" y="%.3f" fill="#1a5c96" stroke="#ffffff" stroke-width="2.6" paint-order="stroke" font-family="Arial, sans-serif" font-size="9" font-weight="700" text-anchor="%s">%s</text>`,
			placed.X, placed.Y, placed.Anchor, html.EscapeString(text))
	}
}

// writeSolarEclipseMagnitudeContours 绘制地方最大食分等值线及其数值标注。
func writeSolarEclipseMagnitudeContours(
	builder *strings.Builder,
	contours []eclipsecore.SolarEclipseMagnitudeContour,
	frame svgmap.Frame,
	labels *svgchart.LabelTable,
	axis [3]float64,
	normal [3]float64,
) {
	if len(contours) == 0 {
		return
	}
	for _, contour := range contours {
		for _, segment := range contour.Segments {
			points := make([]svgmap.GeoPoint, len(segment))
			for index, point := range segment {
				points[index] = svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			writeEclipseMapGeoLine(
				builder, frame, points, "solar-magnitude-contour", "#7c3aed",
				0.8+0.8*math.Max(0, math.Min(1, contour.Magnitude)), "4 3", "solar-map-clip",
				eclipseMapSourceMagnitudeContours,
			)
		}
		text := fmt.Sprintf("%.1f", contour.Magnitude)
		x, y, hasPreferred := solarEclipseMagnitudePreferredPosition(contour.Segments, frame, axis, normal)
		candidates := solarEclipseContourLabelCandidates(contour.Segments, frame, x, y, hasPreferred)
		placed, ok := labels.Place(text, 10, labelCandidatesInsideBox(candidates, text, 10, solarEclipseFrameBox(frame)))
		if !ok {
			continue
		}
		fmt.Fprintf(builder, `<text class="solar-magnitude-label" x="%.3f" y="%.3f" fill="#5b21b6" stroke="#ffffff" stroke-width="3" paint-order="stroke" font-family="Arial, sans-serif" font-size="10" font-weight="700" text-anchor="%s">%s</text>`,
			placed.X, placed.Y, placed.Anchor, html.EscapeString(text))
	}
}

// solarEclipseMagnitudeLabelAxis 返回过食甚点、垂直于中心线的那个大圆的法线。
// 食分等值线是绕中心线的一圈闭合曲线，让它与这个固定大圆相交，各条线的标注就落在同一条线上。
func solarEclipseMagnitudeLabelAxis(path eclipsecore.SolarEclipsePath) ([3]float64, [3]float64, bool) {
	points := path.CenterLine
	if len(points) < 3 {
		return [3]float64{}, [3]float64{}, false
	}
	nearest := 0
	best := math.Inf(1)
	for index, point := range points {
		distance := math.Hypot(point.Longitude-pathsGreatestLongitude(path), point.Latitude-pathsGreatestLatitude(path))
		if distance < best {
			best, nearest = distance, index
		}
	}
	if nearest == 0 {
		nearest = 1
	}
	if nearest >= len(points)-1 {
		nearest = len(points) - 2
	}
	before := solarEclipseUnitVector(points[nearest-1])
	after := solarEclipseUnitVector(points[nearest+1])
	normal := solarEclipseCross(before, after)
	at := solarEclipseUnitVector(points[nearest])
	tangent := solarEclipseCross(normal, at)
	length := math.Sqrt(tangent[0]*tangent[0] + tangent[1]*tangent[1] + tangent[2]*tangent[2])
	normalLength := math.Sqrt(normal[0]*normal[0] + normal[1]*normal[1] + normal[2]*normal[2])
	if length < 1e-12 || normalLength < 1e-12 {
		return [3]float64{}, [3]float64{}, false
	}
	return [3]float64{tangent[0] / length, tangent[1] / length, tangent[2] / length},
		[3]float64{normal[0] / normalLength, normal[1] / normalLength, normal[2] / normalLength}, true
}

func pathsGreatestLongitude(path eclipsecore.SolarEclipsePath) float64 {
	return path.Eclipse.GreatestLongitude
}

func pathsGreatestLatitude(path eclipsecore.SolarEclipsePath) float64 {
	return path.Eclipse.GreatestLatitude
}

func solarEclipseUnitVector(point eclipsecore.SolarEclipsePathPoint) [3]float64 {
	longitude := point.Longitude * math.Pi / 180
	latitude := point.Latitude * math.Pi / 180
	return [3]float64{
		math.Cos(latitude) * math.Cos(longitude),
		math.Cos(latitude) * math.Sin(longitude),
		math.Sin(latitude),
	}
}

func solarEclipseCross(first, second [3]float64) [3]float64 {
	return [3]float64{
		first[1]*second[2] - first[2]*second[1],
		first[2]*second[0] - first[0]*second[2],
		first[0]*second[1] - first[1]*second[0],
	}
}

// solarEclipseMagnitudePreferredPosition 取等值线与标注大圆的交点，让同族标注排在同一条线上。
// 标注大圆与每条闭合等值线有两个交点，一个在中心线北侧、一个在南侧；
// 按“相对中心线的哪一侧”固定取同一侧，否则各条线会各自跳到对面，看着就很散。
func solarEclipseMagnitudePreferredPosition(
	segments [][]eclipsecore.SolarEclipsePathPoint,
	frame svgmap.Frame,
	axis [3]float64,
	normal [3]float64,
) (float64, float64, bool) {
	best, bestDistance, found := eclipsecore.SolarEclipsePathPoint{}, math.Inf(1), false
	bestOther, bestOtherDistance, foundOther := eclipsecore.SolarEclipsePathPoint{}, math.Inf(1), false
	for _, segment := range segments {
		for _, point := range segment {
			vector := solarEclipseUnitVector(point)
			distance := math.Abs(vector[0]*axis[0] + vector[1]*axis[1] + vector[2]*axis[2])
			if vector[0]*normal[0]+vector[1]*normal[1]+vector[2]*normal[2] >= 0 {
				if distance < bestDistance {
					best, bestDistance, found = point, distance, true
				}
				continue
			}
			if distance < bestOtherDistance {
				bestOther, bestOtherDistance, foundOther = point, distance, true
			}
		}
	}
	if !found && foundOther {
		best, found = bestOther, true
	}
	if !found || math.IsInf(bestDistance, 1) {
		return 0, 0, false
	}
	x, y, visible := frame.Project(best.Longitude, best.Latitude)
	if !visible {
		return 0, 0, false
	}
	return x, y - 6, true
}

// writeSolarEclipsePartialBoundary 描出偏食可见域的真实边界。
// 直接描填充路径会把 ±180° 的切边画成假线，所以走按反经线分段的折线绘制。
func writeSolarEclipsePartialBoundary(
	builder *strings.Builder,
	contours [][]eclipsecore.SolarEclipsePathPoint,
	frame svgmap.Frame,
) {
	for _, contour := range contours {
		points := make([]svgmap.GeoPoint, len(contour))
		for index, point := range contour {
			points[index] = svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		writeEclipseMapGeoLine(builder, frame, points, "solar-partial-boundary", "#c8921a", 1.0, "", "solar-map-clip",
			eclipseMapSourcePartialBandContours)
	}
}

// solarEclipseDetailedBlocks 汇总日食详细版式的全部数据块。
// 竖版与横版只是排布不同，块内容必须来自这里，避免两套版式各写一遍造行逻辑。
func solarEclipseDetailedBlocks(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	local eclipsecore.LocalSolarEclipseInfo,
	hasLocal bool,
	geocentric eclipsecore.SolarEclipseGeocentricPanel,
	hasGeocentric bool,
	options SolarEclipseMapSVGOptions,
) solarEclipsePanelBlocks {
	penumbra, umbra, circumstances := solarEclipseContactRows(partial, local, hasLocal, options)
	penumbraTitle, circumstancesTitle, umbraTitle := solarEclipseContactTitles(options)
	ephemerisTitle, librationTitle := solarEclipseEphemerisTitles(options)
	ephemeris, libration := solarEclipseEphemerisRows(geocentric, options)
	sunTitle, moonTitle := solarEclipseGeocentricTitles(options)
	sunRows, moonRows := solarEclipseGeocentricRows(geocentric)
	if !hasGeocentric {
		ephemeris, libration, sunRows, moonRows = nil, nil, nil, nil
	}
	return solarEclipsePanelBlocks{
		sun:           solarEclipsePanelBlock{title: sunTitle, rows: sunRows},
		moon:          solarEclipsePanelBlock{title: moonTitle, rows: moonRows},
		penumbra:      solarEclipsePanelBlock{title: penumbraTitle, rows: penumbra},
		umbra:         solarEclipsePanelBlock{title: umbraTitle, rows: umbra},
		circumstances: solarEclipsePanelBlock{title: circumstancesTitle, rows: circumstances},
		ephemeris:     solarEclipsePanelBlock{title: ephemerisTitle, rows: ephemeris},
		libration:     solarEclipsePanelBlock{title: librationTitle, rows: libration},
	}
}
