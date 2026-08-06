package svg

import (
	"fmt"
	"html"
	"math"
	"sort"
	"strings"
	"time"

	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/svgmap"
)

type solarEclipseMapLayout struct {
	frame       svgmap.Frame
	panelX      float64
	panelY      float64
	panelWidth  float64
	panelHeight float64
}

type solarEclipseGlobalEventRow struct {
	kind     string
	name     string
	time     time.Time
	point    eclipsecore.SolarEclipsePathPoint
	hasPoint bool
}

func solarEclipseMapLayoutFor(options SolarEclipseMapSVGOptions, projection svgmap.Projection) solarEclipseMapLayout {
	width := float64(options.Width)
	height := float64(options.Height)
	margin := math.Max(26, math.Min(45, width*0.04))
	gap := math.Max(16, math.Min(24, width*0.025))
	panelWidth := math.Max(148, math.Min(238, width*0.22))
	availableWidth := width - 2*margin - gap - panelWidth
	bottomReserve := 92.0
	if projection != svgmap.ProjectionEquirectangular {
		bottomReserve = 118
	}
	availableHeight := math.Max(110, height-142-bottomReserve)
	mapWidth := math.Min(availableWidth, availableHeight*2)
	mapHeight := mapWidth / 2
	if projection != svgmap.ProjectionEquirectangular {
		mapWidth = math.Min(availableWidth, availableHeight)
		mapHeight = mapWidth
	}
	mapY := 142 + math.Max(0, (availableHeight-mapHeight)/2)
	groupWidth := mapWidth + gap + panelWidth
	mapX := math.Max(margin, (width-groupWidth)/2)
	return solarEclipseMapLayout{
		frame: svgmap.Frame{
			X: mapX, Y: mapY, Width: mapWidth, Height: mapHeight, Projection: projection,
		},
		panelX:      mapX + mapWidth + gap,
		panelY:      mapY,
		panelWidth:  panelWidth,
		panelHeight: mapHeight,
	}
}

func writeSolarEclipseMapSectionTitle(
	builder *strings.Builder,
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
	}
	fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#161a1b" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700">%s</text>`,
		layout.frame.X, layout.frame.Y-10, html.EscapeString(label))
}

func writeSolarEclipseGlobalEventsPanel(
	builder *strings.Builder,
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	path eclipsecore.SolarEclipsePath,
	hasCentral bool,
	layout solarEclipseMapLayout,
	options SolarEclipseMapSVGOptions,
) {
	rows := solarEclipseGlobalEventRows(partial, path, hasCentral, options.Language)
	title := options.EventsTitle
	if title == "" {
		if options.Language == "en" {
			title = "Global phases"
		} else {
			title = "全球阶段"
		}
	}
	fmt.Fprintf(builder, `<g class="solar-global-events">`)
	fmt.Fprintf(builder, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#d0d3d1" stroke-width="1"/>`,
		layout.panelX-10, layout.panelY, layout.panelX-10, layout.panelY+layout.panelHeight)
	fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#161a1b" font-family="Georgia, 'Times New Roman', serif" font-size="14" font-weight="700">%s</text>`,
		layout.panelX, layout.panelY+13, html.EscapeString(title))
	rowTop := layout.panelY + 27
	rowHeight := math.Max(21, (layout.panelHeight-27)/float64(len(rows)))
	for index, row := range rows {
		y := rowTop + float64(index)*rowHeight
		if index > 0 {
			fmt.Fprintf(builder, `<line x1="%.3f" y1="%.3f" x2="%.3f" y2="%.3f" stroke="#e0e1df" stroke-width="0.8"/>`,
				layout.panelX, y-5, layout.panelX+layout.panelWidth, y-5)
		}
		fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#1c2528" font-family="Arial, sans-serif" font-size="11" font-weight="700">%s</text>`,
			layout.panelX, y+10, html.EscapeString(row.name))
		fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#1c2528" font-family="Arial, sans-serif" font-size="10" text-anchor="end">%s</text>`,
			layout.panelX+layout.panelWidth, y+10, html.EscapeString(row.time.In(options.Location).Format("15:04:05")))
		if row.hasPoint && rowHeight >= 45 {
			fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#4c585a" font-family="Arial, sans-serif" font-size="10">%s</text>`,
				layout.panelX, y+27, html.EscapeString(solarEclipseFormatCoordinates(row.point.Longitude, row.point.Latitude)))
		}
		if row.hasPoint && rowHeight >= 52 {
			detail := solarEclipseGlobalEventDetail(row, partial.Eclipse, options.Language)
			fmt.Fprintf(builder, `<text x="%.3f" y="%.3f" fill="#697476" font-family="Arial, sans-serif" font-size="9">%s</text>`,
				layout.panelX, y+43, html.EscapeString(detail))
		}
	}
	builder.WriteString(`</g>`)
}

func solarEclipseGlobalEventRows(
	partial eclipsecore.SolarEclipsePartialFootprintsInfo,
	path eclipsecore.SolarEclipsePath,
	hasCentral bool,
	language string,
) []solarEclipseGlobalEventRow {
	info := partial.Eclipse
	names := []string{"偏食始", "中心食始", "食甚", "中心食终", "偏食终"}
	if language == "en" {
		names = []string{"Partial begins", "Central begins", "Greatest", "Central ends", "Partial ends"}
	}
	rows := make([]solarEclipseGlobalEventRow, 0, 11)
	appendContact := func(name string, point eclipsecore.SolarEclipsePathPoint) {
		if point.Time.IsZero() {
			return
		}
		rows = append(rows, solarEclipseGlobalEventRow{
			kind: "shadow-contact", name: name, time: point.Time, point: point, hasPoint: true,
		})
	}
	appendContact("P1 "+names[0], partial.P1)
	appendContact("P2", partial.P2)
	appendContact("U1", partial.U1)
	appendContact("U2", partial.U2)
	if hasCentral && len(path.CenterLine) > 0 {
		rows = append(rows, solarEclipseGlobalEventRow{
			kind: "central-start", name: names[1], time: info.CentralBeginOnEarth,
			point: path.CenterLine[0], hasPoint: true,
		})
	}
	greatest := eclipsecore.SolarEclipsePathPoint{
		Time: info.GreatestEclipse, Longitude: info.GreatestLongitude,
		Latitude: info.GreatestLatitude, WidthKM: info.PathWidthKM,
	}
	if hasCentral {
		greatest = path.Greatest
	}
	rows = append(rows, solarEclipseGlobalEventRow{
		kind: "greatest", name: names[2], time: info.GreatestEclipse,
		point: greatest, hasPoint: true,
	})
	appendContact("U3", partial.U3)
	appendContact("U4", partial.U4)
	if hasCentral && len(path.CenterLine) > 0 {
		rows = append(rows, solarEclipseGlobalEventRow{
			kind: "central-end", name: names[3], time: info.CentralEndOnEarth,
			point: path.CenterLine[len(path.CenterLine)-1], hasPoint: true,
		})
	}
	appendContact("P3", partial.P3)
	appendContact("P4 "+names[4], partial.P4)
	if len(rows) == 0 || partial.P1.Time.IsZero() {
		rows = append(rows, solarEclipseGlobalEventRow{kind: "partial-start", name: names[0], time: info.PartialBeginOnEarth})
	}
	if partial.P4.Time.IsZero() {
		rows = append(rows, solarEclipseGlobalEventRow{kind: "partial-end", name: names[4], time: info.PartialEndOnEarth})
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].time.Before(rows[j].time) })
	return rows
}

func solarEclipseGlobalEventDetail(
	row solarEclipseGlobalEventRow,
	info eclipsecore.SolarEclipseInfo,
	language string,
) string {
	if row.kind == "greatest" {
		if info.HasCentral {
			if language == "en" {
				return fmt.Sprintf("Path width %.1f km", info.PathWidthKM)
			}
			return fmt.Sprintf("食带宽 %.1f km", info.PathWidthKM)
		}
		if language == "en" {
			return fmt.Sprintf("Magnitude %.3f", info.Magnitude)
		}
		return fmt.Sprintf("食分 %.3f", info.Magnitude)
	}
	if language == "en" {
		return fmt.Sprintf("Sun altitude %+.1f°", row.point.SunAltitude)
	}
	return fmt.Sprintf("太阳高度 %+.1f°", row.point.SunAltitude)
}

func writeSolarEclipseTerminator(
	builder *strings.Builder,
	info eclipsecore.SolarEclipseInfo,
	frame svgmap.Frame,
) {
	terminator := svgmap.SphericalCircle(solarEclipseSubsolarPoint(info.GreatestEclipse), 90, 360)
	if len(terminator) > 0 {
		terminator = append(terminator, terminator[0])
	}
	writeEclipseMapGeoLine(
		builder, frame, terminator, "solar-greatest-terminator", "#6f7778", 1.05, "4 3", "solar-map-clip",
	)
}

func writeSolarEclipsePenumbralOutlines(
	builder *strings.Builder,
	info eclipsecore.SolarEclipsePartialFootprintsInfo,
	frame svgmap.Frame,
	options SolarEclipseMapSVGOptions,
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
	labelPositions := make([][2]float64, 0, len(selected))
	for _, footprint := range selected {
		writeSolarEclipseFootprintBoundary(
			builder, footprint, frame, "solar-penumbral-outline", "#b07a18", 0.75, "3 3",
		)
		if mapTimeDistance(footprint.Time, info.Eclipse.GreatestEclipse) <= info.Step/2 {
			continue
		}
		x, y, ok := solarEclipseFootprintLabelPosition(footprint, frame)
		if !ok || solarEclipseMapLabelOverlaps(x, y, labelPositions) {
			continue
		}
		labelPositions = append(labelPositions, [2]float64{x, y})
		labelTime := solarEclipseMapAlignedTime(footprint.Time, options.PenumbralOutlineStep, options.Location)
		fmt.Fprintf(builder, `<text class="solar-penumbral-time-label" x="%.3f" y="%.3f" fill="#8b5b08" stroke="#ffffff" stroke-width="2.4" paint-order="stroke" font-family="Arial, sans-serif" font-size="8" font-weight="700" text-anchor="middle">%s</text>`,
			x, y-4, html.EscapeString(labelTime.Format("15:04")))
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
		)
	}
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

func solarEclipseMapLabelOverlaps(x, y float64, positions [][2]float64) bool {
	for _, position := range positions {
		if math.Abs(x-position[0]) < 52 && math.Abs(y-position[1]) < 17 {
			return true
		}
	}
	return false
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
	dash string,
) {
	for _, boundary := range footprint.Boundaries {
		points := make([]svgmap.GeoPoint, len(boundary))
		for index, point := range boundary {
			points[index] = svgmap.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		writeEclipseMapGeoLine(builder, frame, points, className, color, strokeWidth, dash, "solar-map-clip")
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

func writeSolarEclipseContactMarkers(
	builder *strings.Builder,
	info eclipsecore.SolarEclipsePartialFootprintsInfo,
	frame svgmap.Frame,
) {
	type marker struct {
		name       string
		point      eclipsecore.SolarEclipsePathPoint
		color      string
		labelDX    float64
		labelDY    float64
		textAnchor string
	}
	markers := []marker{
		{name: "P1", point: info.P1, color: "#a52d70", labelDX: -7, labelDY: 14, textAnchor: "end"},
		{name: "P2", point: info.P2, color: "#a52d70", labelDX: 7, labelDY: -7, textAnchor: "start"},
		{name: "P3", point: info.P3, color: "#a52d70", labelDX: -7, labelDY: -7, textAnchor: "end"},
		{name: "P4", point: info.P4, color: "#a52d70", labelDX: 7, labelDY: 14, textAnchor: "start"},
		{name: "U1", point: info.U1, color: "#a12c25", labelDX: -7, labelDY: -7, textAnchor: "end"},
		{name: "U2", point: info.U2, color: "#a12c25", labelDX: 7, labelDY: 14, textAnchor: "start"},
		{name: "U3", point: info.U3, color: "#a12c25", labelDX: -7, labelDY: 14, textAnchor: "end"},
		{name: "U4", point: info.U4, color: "#a12c25", labelDX: 7, labelDY: -7, textAnchor: "start"},
	}
	for _, marker := range markers {
		if marker.point.Time.IsZero() {
			continue
		}
		x, y, visible := frame.Project(marker.point.Longitude, marker.point.Latitude)
		if !visible {
			continue
		}
		labelX := x + marker.labelDX
		textAnchor := marker.textAnchor
		if labelX < frame.X+18 {
			labelX = x + 7
			textAnchor = "start"
		} else if labelX > frame.X+frame.Width-18 {
			labelX = x - 7
			textAnchor = "end"
		}
		fmt.Fprintf(builder, `<g class="solar-shadow-contact solar-contact-%s"><circle cx="%.3f" cy="%.3f" r="2.7" fill="%s" stroke="#ffffff" stroke-width="0.9"/><text x="%.3f" y="%.3f" fill="%s" stroke="#ffffff" stroke-width="2.4" paint-order="stroke" font-family="Arial, sans-serif" font-size="8" font-weight="700" text-anchor="%s">%s</text></g>`,
			strings.ToLower(marker.name), x, y, marker.color, labelX, y+marker.labelDY, marker.color,
			textAnchor, marker.name)
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

func writeSolarEclipseSubsolarMarker(
	builder *strings.Builder,
	info eclipsecore.SolarEclipseInfo,
	frame svgmap.Frame,
	language string,
) {
	point := solarEclipseSubsolarPoint(info.GreatestEclipse)
	x, y, visible := frame.Project(point.Longitude, point.Latitude)
	if !visible {
		return
	}
	label := "日下点"
	if language == "en" {
		label = "Subsolar"
	}
	fmt.Fprintf(builder, `<g class="solar-subsolar-marker"><circle cx="%.3f" cy="%.3f" r="4" fill="#f4c542" stroke="#714f00" stroke-width="1"/><path d="M %.3f %.3f h 8 M %.3f %.3f v 8" fill="none" stroke="#714f00" stroke-width="1"/><text x="%.3f" y="%.3f" fill="#714f00" stroke="#ffffff" stroke-width="2.5" paint-order="stroke" font-family="Arial, sans-serif" font-size="9" font-weight="700" text-anchor="middle">%s</text></g>`,
		x, y, x-4, y, x, y-4, x, y-9, html.EscapeString(label))
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

func writeSolarEclipseTimeMarkers(
	builder *strings.Builder,
	path eclipsecore.SolarEclipsePath,
	frame svgmap.Frame,
	options SolarEclipseMapSVGOptions,
) {
	if options.TimeLabelStep <= 0 || len(path.CenterLine) < 2 {
		return
	}
	excluded := []time.Time{
		path.Eclipse.CentralBeginOnEarth,
		path.Eclipse.CentralEndOnEarth,
	}
	markers := solarEclipseTimeMarkerPoints(path.CenterLine, options.TimeLabelStep, options.Location, excluded)
	projected := make([][2]float64, 0, len(markers))
	for _, marker := range markers {
		x, y, visible := frame.Project(marker.Longitude, marker.Latitude)
		if !visible || x < frame.X+22 || x > frame.X+frame.Width-22 || y < frame.Y+12 || y > frame.Y+frame.Height-12 {
			continue
		}
		tooClose := false
		for _, previous := range projected {
			if math.Hypot(x-previous[0], y-previous[1]) < 44 {
				tooClose = true
				break
			}
		}
		if tooClose {
			continue
		}
		projected = append(projected, [2]float64{x, y})
		labelY := y - 8
		if mapTimesNear(marker.Time, path.Eclipse.GreatestEclipse, solarEclipseGreatestTimeLabelWindow(options.TimeLabelStep)) {
			labelY = y + 15
		} else if labelY < frame.Y+10 {
			labelY = y + 15
		}
		fmt.Fprintf(builder, `<g class="solar-time-marker"><circle cx="%.3f" cy="%.3f" r="2.3" fill="#263f58" stroke="#ffffff" stroke-width="1"/><text x="%.3f" y="%.3f" fill="#263f58" stroke="#ffffff" stroke-width="3" paint-order="stroke" font-family="Arial, sans-serif" font-size="9" font-weight="700" text-anchor="middle">%s</text></g>`,
			x, y, x, labelY, html.EscapeString(marker.Time.In(options.Location).Format("15:04")))
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
