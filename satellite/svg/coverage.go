package svg

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/internal/svgmap"
	"github.com/starainrt/astro/satellite"
)

const coverageCurrentVisualStroke = "#793953"

func coverageRingValid(ring []satellite.GeoPoint) bool {
	if len(ring) < 3 {
		return false
	}
	for _, p := range ring {
		if !validPoint(p.Longitude, p.Latitude) {
			return false
		}
	}
	return true
}
func validateCoverageSVG(s satellite.CoverageSweep, o CoverageSVGOptions) (int, bool) {
	if s.Start.IsZero() || s.End.Before(s.Start) || s.Elements.Epoch.IsZero() || s.Elements.CatalogNumber <= 0 || len(s.Track) == 0 || len(s.Track) != len(s.Instants) || len(s.Track) != len(s.TrackIllumination) || !s.Track[0].Time.Equal(s.Start) || !s.Track[len(s.Track)-1].Time.Equal(s.End) {
		return -1, false
	}
	selected := -1
	for i, p := range s.Track {
		v := s.Instants[i]
		if p.Time.IsZero() || i > 0 && !p.Time.After(s.Track[i-1].Time) || !v.Time.Equal(p.Time) || !validPoint(p.Longitude, p.Latitude) || !validNumber(p.HeightKM) || !validNumber(p.SpeedKMS) || p.SpeedKMS < 0 || s.TrackIllumination[i] < satellite.IlluminationSunlit || s.TrackIllumination[i] > satellite.IlluminationUmbra || len(v.Footprint) != 0 && !coverageRingValid(v.Footprint) || !validPoint(v.SubPoint.Longitude, v.SubPoint.Latitude) || !validNumber(v.SubPoint.HeightKM) || !validNumber(v.SubPoint.SpeedKMS) {
			return -1, false
		}
		if len(v.Night) > 0 && (v.NightIsGlobal || !coverageRingValid(v.Night)) {
			return -1, false
		}
		for _, r := range v.Visible {
			if !coverageRingValid(r) {
				return -1, false
			}
		}
		if !coverageRegionsValid(v.VisibleRegions) {
			return -1, false
		}
		if !o.Highlight.IsZero() && o.Highlight.Equal(v.Time) {
			selected = i
		}
	}
	if !o.Highlight.IsZero() && selected < 0 {
		return -1, false
	}
	if !coverageRegionsValid(s.Swept) || !coverageRegionsValid(s.VisibleSwept) {
		return -1, false
	}
	if o.Observer != nil && (!validPoint(o.Observer.Longitude, o.Observer.Latitude) || !validNumber(o.Observer.Height)) {
		return -1, false
	}
	if o.Passes != nil && (o.Observer == nil || !coveragePassesValid(s.Start, s.End, o.Passes)) {
		return -1, false
	}
	switch o.Projection {
	case MapProjectionAuto, MapProjectionEquirectangular, MapProjectionNorthPolar, MapProjectionSouthPolar, MapProjectionOrthographic:
	default:
		return -1, false
	}
	return selected, true
}

func coverageRegionsValid(regions []satellite.Region) bool {
	for _, region := range regions {
		if !coverageRingValid(region.Outer) {
			return false
		}
		for _, hole := range region.Holes {
			if !coverageRingValid(hole) {
				return false
			}
		}
	}
	return true
}

// CoverageSVG 绘制已计算的卫星覆盖范围，不调用星历 / render computed satellite coverage without ephemeris calls.
func CoverageSVG(sweep satellite.CoverageSweep, options CoverageSVGOptions) (string, error) {
	selected, ok := validateCoverageSVG(sweep, options)
	if !ok {
		return "", satellite.ErrInvalidInput
	}
	w, h := options.Width, options.Height
	if w == 0 {
		w = 1200
	}
	if h == 0 {
		h = 800
	}
	if w < 640 || h < 480 {
		return "", satellite.ErrInvalidInput
	}
	loc := options.Location
	if loc == nil {
		loc = sweep.Start.Location()
	}
	clock := "15:04:05"
	_, startOffset := sweep.Start.In(loc).Zone()
	_, endOffset := sweep.End.In(loc).Zone()
	if startOffset != endOffset {
		clock += " -07:00"
	}
	c := chart{width: float64(w), height: float64(h), en: strings.EqualFold(options.Language, "en"), loc: loc, clock: clock}
	title := options.Title
	if title == "" {
		title = sweep.Start.In(loc).Format("2006-01-02") + " " + satelliteName(sweep.Elements) + c.words(" 可见范围", " Satellite Coverage")
	}
	endFormat := clock
	if sweep.Start.In(loc).Format("2006-01-02") != sweep.End.In(loc).Format("2006-01-02") {
		endFormat = "2006-01-02 " + clock
	}
	header := chart{width: c.width}
	headerBottom := header.heading(title, c.words("扫掠时段 ", "Sweep interval ")+sweep.Start.In(loc).Format("2006-01-02 "+clock)+" – "+sweep.End.In(loc).Format(endFormat)+" ("+loc.String()+")")
	highlighted := selected >= 0
	focus := len(sweep.Track) / 2
	if highlighted {
		focus = selected
	}
	f, panel, contentBottom := c.coverageLayout(sweep, options, focus, highlighted, headerBottom)
	station := chart{width: c.width, en: c.en, loc: loc}
	if options.Passes != nil {
		stationBottom := station.coveragePassPanel(options.Passes, math.Ceil(contentBottom)+20)
		c.height = math.Max(c.height, stationBottom+54)
	}
	c.openCanvas()
	c.b.WriteString(header.b.String())
	fmt.Fprintf(&c.b, `<defs>%s</defs><g data-role="map" data-x="%.6f" data-y="%.6f" data-width="%.6f" data-height="%.6f" data-projection="%s" data-longitude="%.9f" data-latitude="%.9f">`, f.ClipDefinition("map-clip"), f.X, f.Y, f.Width, f.Height, f.Projection, f.CenterLongitude, f.CenterLatitude)
	f.WriteOcean(&c.b)
	f.WriteLand(&c.b, "map-clip")
	f.WriteGraticule(&c.b, "map-clip")
	if selected >= 0 {
		c.coverageNight(sweep.Instants[selected], f)
	}
	c.coverageShape("coverage-sweep", sweep.Swept, f, "#e0ae43", "#a66f18", .28, 1.3)
	c.coverageShape("visible-sweep", sweep.VisibleSwept, f, "#e0ae43", "#a66f18", .45, 1.2)
	if selected >= 0 {
		v := sweep.Instants[selected]
		c.coverageShape("satellite-footprint", coverageRawRegions([][]satellite.GeoPoint{v.Footprint}), f, "none", "#a66f18", 1, 1.35)
		c.coverageShape("satellite-visible", coverageVisualRegions(v), f, "#e0ae43", coverageCurrentVisualStroke, .55, 1.6)
	}
	c.coverageTrack(sweep, f)
	c.coverageLabels(sweep, f, options, focus, selected >= 0)
	if o := options.Observer; o != nil {
		if x, y, ok := f.Project(o.Longitude, o.Latitude); ok {
			fmt.Fprintf(&c.b, `<circle data-role="observer" cx="%.6f" cy="%.6f" r="4" fill="#2d3e5a" stroke="white"/>`, x, y)
		}
	}
	f.WriteFrame(&c.b)
	c.b.WriteString(`</g>`)
	c.b.WriteString(panel)
	at := sweep.Start
	if selected >= 0 {
		at = sweep.Instants[selected].Time
	}
	c.b.WriteString(station.b.String())
	c.footer(sweep.Elements, at, options.FooterNote)
	return c.b.String(), nil
}

func (c *chart) coverageLayout(sweep satellite.CoverageSweep, options CoverageSVGOptions, focus int, highlighted bool, headerBottom float64) (svgmap.Frame, string, float64) {
	path := satellite.TransitPath{CenterLine: make([]satellite.TransitPathPoint, len(sweep.Track))}
	for index, point := range sweep.Track {
		path.CenterLine[index] = satellite.TransitPathPoint{Longitude: point.Longitude, Latitude: point.Latitude}
	}
	path.Event.GreatestPoint = path.CenterLine[focus]
	frame := mapFrame(path, TransitMapSVGOptions{Projection: options.Projection}, c.width, c.height)
	if frame.Projection == svgmap.ProjectionEquirectangular {
		frame.CenterLongitude = sweep.Track[focus].Longitude
	}
	panel := chart{width: c.width, en: c.en, loc: c.loc, clock: c.clock}
	hasFootprint := highlighted && len(sweep.Instants[focus].Footprint) > 0
	legendTop := coverageLegendTop(c.height, hasFootprint)
	currentVisual := highlighted && (len(sweep.Instants[focus].VisibleRegions) > 0 || len(sweep.Instants[focus].Visible) > 0)
	legendExtra := 0.0
	if currentVisual {
		legendExtra = 17
		legendTop -= legendExtra
	}
	legendWidth := 25 + svgchart.EstimatedTextWidth(c.coverageTrackLegend(), 11)
	compact := c.height < 600 || c.width*.30-14 < legendWidth
	var panelBottom float64
	if !compact {
		panelBottom = panel.coveragePanel(sweep, options, focus, highlighted, frame, headerBottom, false)
		compact = panelBottom > legendTop
	}
	if compact {
		top := math.Max(130, headerBottom+22)
		mapHeight := math.Max(1, math.Min(frame.Height, c.height-top-155-legendExtra))
		if frame.IsDisk() {
			frame.X += (frame.Width - mapHeight) / 2
			frame.Width = mapHeight
		}
		frame.Height = mapHeight
		frame.Y = top
		panel.b.Reset()
		panelBottom = panel.coveragePanel(sweep, options, focus, highlighted, frame, headerBottom, true)
		legendTop = c.height - 76 - legendExtra
	}
	mapBottom := frame.Y + frame.Height
	if options.Passes != nil {
		legendTop = panelBottom + 12
		if compact {
			legendTop = math.Max(legendTop, mapBottom+14)
		}
	}
	legendBottom := panel.coverageLegend(c.width*.70, legendTop, hasFootprint, currentVisual, compact)
	return frame, panel.b.String(), math.Max(mapBottom, math.Max(panelBottom, legendBottom))
}

func coverageVisualRegions(coverage satellite.Coverage) []satellite.Region {
	if len(coverage.VisibleRegions) != 0 {
		return coverage.VisibleRegions
	}
	return coverageRawRegions(coverage.Visible)
}

func coverageRawRegions(rings [][]satellite.GeoPoint) []satellite.Region {
	var result []satellite.Region
	for _, r := range rings {
		for _, piece := range svgmap.PolygonFragments(coverageGeo(r), svgmap.ClipView{Projection: svgmap.ProjectionEquirectangular}) {
			ring := make([]satellite.GeoPoint, len(piece))
			for i, p := range piece {
				ring[i] = satellite.GeoPoint{Longitude: p.Longitude, Latitude: p.Latitude}
			}
			result = append(result, satellite.Region{Outer: ring})
		}
	}
	return result
}
func coverageGeo(ring []satellite.GeoPoint) []svgmap.GeoPoint {
	out := make([]svgmap.GeoPoint, len(ring))
	for i, p := range ring {
		out[i] = svgmap.GeoPoint{Longitude: p.Longitude, Latitude: p.Latitude}
	}
	return out
}
func coverageRings(regions []satellite.Region) [][]satellite.GeoPoint {
	var rings [][]satellite.GeoPoint
	for _, r := range regions {
		rings = append(rings, r.Outer)
		rings = append(rings, r.Holes...)
	}
	return rings
}

func (c *chart) coverageNight(coverage satellite.Coverage, frame svgmap.Frame) {
	var path strings.Builder
	if coverage.NightIsGlobal {
		writeCoverageFrame(&path, frame)
		c.coverageFill("night-region", &path, "#87909a", .20)
		return
	}
	if len(coverage.Night) == 0 {
		return
	}
	points := coverageGeo(coverage.Night)
	view := frame.Clip()
	if frame.IsPolar() {
		view.Projection = svgmap.ProjectionOrthographic
		view.Center.Latitude = 90
		if frame.Projection == svgmap.ProjectionSouthPolar {
			view.Center.Latitude = -90
		}
	}
	for _, fragment := range svgmap.PolygonFragments(points, view) {
		writeCoverageFragment(&path, fragment, frame, nil)
	}
	if geodata.ClockwiseRegionIsComplement(points) {
		writeCoverageFrame(&path, frame)
	}
	c.coverageFill("night-region", &path, "#87909a", .20)
	if !geodata.SameGeoPoint(points[0], points[len(points)-1]) {
		points = append(points, points[0])
	}
	c.geoLine("night-region-boundary", points, frame, "#78838c", "", .9)
}

func (c *chart) coverageShape(role string, regions []satellite.Region, f svgmap.Frame, fill, stroke string, opacity, width float64) {
	if len(regions) == 0 {
		return
	}
	var path strings.Builder
	var probes []coverageFillProbe
	if f.Projection == svgmap.ProjectionOrthographic && fill != "none" {
		probes = coverageFillProbes(f)
	}
	for _, ring := range coverageRings(regions) {
		for _, piece := range svgmap.PolygonFragments(coverageGeo(ring), f.Clip()) {
			writeCoverageFragment(&path, piece, f, probes)
		}
	}
	if len(probes) > 0 && coverageFillNeedsDisk(probes, regions) {
		writeCoverageFrame(&path, f)
	}
	c.coverageFill(role, &path, fill, opacity)
	for _, ring := range coverageStrokePaths(regions) {
		dash := ""
		if role == "satellite-footprint" {
			dash = "5 4"
		}
		c.geoLine(role+"-boundary", coverageGeo(ring), f, stroke, dash, width)
	}
}

func (c *chart) coverageFill(role string, path *strings.Builder, fill string, opacity float64) {
	if path.Len() > 0 {
		fmt.Fprintf(&c.b, `<path data-role="%s" d="%s" fill="%s" fill-opacity="%.2f" fill-rule="evenodd" stroke="none" clip-path="url(#map-clip)"/>`, role, path.String(), fill, opacity)
	}
}

func writeCoverageFrame(path *strings.Builder, frame svgmap.Frame) {
	fmt.Fprintf(path, "M%.6f %.6f L%.6f %.6f L%.6f %.6f L%.6f %.6f Z ",
		frame.X, frame.Y, frame.X+frame.Width, frame.Y, frame.X+frame.Width, frame.Y+frame.Height, frame.X, frame.Y+frame.Height)
}

func writeCoverageFragment(path *strings.Builder, fragment []svgmap.GeoPoint, frame svgmap.Frame, probes []coverageFillProbe) {
	var firstX, firstY, previousX, previousY float64
	started := false
	for _, point := range fragment {
		x, y, ok := frame.Project(point.Longitude, point.Latitude)
		if !ok {
			continue
		}
		command := "L"
		if !started {
			command = "M"
			firstX, firstY = x, y
			started = true
		} else {
			coverageFillEdge(probes, previousX, previousY, x, y)
		}
		fmt.Fprintf(path, "%s%.6f %.6f ", command, x, y)
		previousX, previousY = x, y
	}
	if started {
		coverageFillEdge(probes, previousX, previousY, firstX, firstY)
		path.WriteString("Z ")
	}
}

type coverageFillProbe struct {
	point          satellite.GeoPoint
	x, y, distance float64
	inside         bool
}

func coverageFillProbes(frame svgmap.Frame) []coverageFillProbe {
	center := frame.Clip().Center
	points := append([]svgmap.GeoPoint{center}, svgmap.SphericalCircle(center, 45, 8)...)
	probes := make([]coverageFillProbe, 0, len(points))
	for _, point := range points {
		x, y, ok := frame.Project(point.Longitude, point.Latitude)
		if ok {
			probes = append(probes, coverageFillProbe{point: satellite.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}, x: x, y: y, distance: math.Inf(1)})
		}
	}
	return probes
}

func coverageFillEdge(probes []coverageFillProbe, x1, y1, x2, y2 float64) {
	dx, dy := x2-x1, y2-y1
	lengthSquared := dx*dx + dy*dy
	for index := range probes {
		probe := &probes[index]
		px, py := probe.x-x1, probe.y-y1
		if (y1 > probe.y) != (y2 > probe.y) && px < dx*py/dy {
			probe.inside = !probe.inside
		}
		if lengthSquared > 0 {
			fraction := math.Max(0, math.Min(1, (px*dx+py*dy)/lengthSquared))
			px, py = px-fraction*dx, py-fraction*dy
		}
		probe.distance = math.Min(probe.distance, px*px+py*py)
	}
}

func coverageFillNeedsDisk(probes []coverageFillProbe, regions []satellite.Region) bool {
	probe := probes[0]
	for _, candidate := range probes[1:] {
		if candidate.distance > probe.distance {
			probe = candidate
		}
	}
	return probe.inside != coverageInside(regions, probe.point)
}

func coverageColor(kind satellite.Illumination) string {
	switch kind {
	case satellite.IlluminationSunlit:
		return "#d9ac30"
	case satellite.IlluminationPenumbra:
		return "#da832b"
	default:
		return "#505050"
	}
}
func (c *chart) coverageTrack(s satellite.CoverageSweep, f svgmap.Frame) {
	c.b.WriteString(`<g stroke-linecap="round" stroke-linejoin="round">`)
	for i := 0; i < len(s.Track)-1; {
		j := i + 1
		for j < len(s.Track)-1 && s.TrackIllumination[j] == s.TrackIllumination[i] {
			j++
		}
		line := make([]svgmap.GeoPoint, j-i+1)
		for k, p := range s.Track[i : j+1] {
			line[k] = svgmap.GeoPoint{Longitude: p.Longitude, Latitude: p.Latitude}
		}
		c.geoLine("ground-track", line, f, coverageColor(s.TrackIllumination[i]), "", 1.7)
		i = j
	}
	c.b.WriteString(`</g>`)
}
func (c *chart) coverageLabels(s satellite.CoverageSweep, f svgmap.Frame, o CoverageSVGOptions, focus int, highlighted bool) {
	step := o.TimeLabelStep
	point := s.Track[focus]
	if highlighted {
		point = s.Instants[focus].SubPoint
		point.Time = s.Instants[focus].Time
	}
	if step < 0 {
		c.coverageTimeMark(f, nil, point, true)
		return
	}
	if step == 0 {
		span := s.End.Sub(s.Start)
		step = time.Minute * time.Duration(math.Ceil(span.Minutes()/6))
	}
	if step < time.Minute {
		step = time.Minute
	}
	var labels svgchart.LabelTable
	lines := append(coverageStrokePaths(s.Swept), coverageStrokePaths(s.VisibleSwept)...)
	track := make([]satellite.GeoPoint, len(s.Track))
	for i, p := range s.Track {
		track[i] = satellite.GeoPoint{Longitude: p.Longitude, Latitude: p.Latitude}
	}
	lines = append(lines, track)
	if highlighted {
		v := s.Instants[focus]
		lines = append(lines, v.Night)
		lines = append(lines, coverageStrokePaths(coverageRawRegions([][]satellite.GeoPoint{v.Footprint}))...)
		lines = append(lines, coverageStrokePaths(coverageVisualRegions(v))...)
	}
	if p := o.Observer; p != nil {
		if x, y, ok := f.Project(p.Longitude, p.Latitude); ok {
			labels.Reserve(x-6, y-6, 12, 12)
		}
	}
	coverageReserveLines(&labels, lines, f)
	c.coverageTimeMark(f, &labels, point, true)
	next := s.Start
	for i, tick := range s.Track {
		if tick.Time.Before(next) && i != len(s.Track)-1 {
			continue
		}
		next = tick.Time.Add(step)
		if i != focus {
			c.coverageTimeMark(f, &labels, tick, false)
		}
	}
}

func coverageReserveLines(labels *svgchart.LabelTable, lines [][]satellite.GeoPoint, f svgmap.Frame) {
	for _, points := range lines {
		for _, line := range svgmap.PolylineSegments(coverageGeo(points), f.Clip()) {
			for i := 1; i < len(line); i++ {
				x, y, ok := f.Project(line[i-1].Longitude, line[i-1].Latitude)
				qx, qy, valid := f.Project(line[i].Longitude, line[i].Latitude)
				if !ok || !valid {
					continue
				}
				steps := int(math.Ceil(math.Hypot(qx-x, qy-y)/8)) + 1
				for j := 0; j < steps; j++ {
					a, b := float64(j)/float64(steps), float64(j+1)/float64(steps)
					ax, ay, bx, by := x+a*(qx-x), y+a*(qy-y), x+b*(qx-x), y+b*(qy-y)
					labels.Reserve(math.Min(ax, bx)-3, math.Min(ay, by)-3, math.Abs(bx-ax)+6, math.Abs(by-ay)+6)
				}
			}
		}
	}
}
func (c *chart) coverageTimeMark(f svgmap.Frame, labels *svgchart.LabelTable, point satellite.GroundPoint, primary bool) {
	x, y, ok := f.Project(point.Longitude, point.Latitude)
	if !ok {
		return
	}
	role, radius := "time-marker", 2.4
	if primary {
		role, radius = "subsatellite-point", 4.5
	}
	at := point.Time.UTC().Format(time.RFC3339Nano)
	if labels != nil {
		labels.Reserve(x-radius-1, y-radius-1, 2*radius+2, 2*radius+2)
		c.coverageTimeLabel(labels, f, x, y, at, point.Time.In(c.loc).Format(c.clock))
	}
	fmt.Fprintf(&c.b, `<circle data-role="%s" data-time="%s" cx="%.6f" cy="%.6f" r="%.1f" fill="#2d3e5a" stroke="white"/>`, role, at, x, y, radius)
}

func (c *chart) coverageTimeLabel(labels *svgchart.LabelTable, f svgmap.Frame, x, y float64, at, text string) {
	var candidates []svgchart.LabelPlacement
	for _, radius := range []float64{16, 32, 48, 64} {
		for _, p := range svgchart.LabelCandidates(x+8, y-9, "start", radius) {
			box := svgchart.LabelTextBox(p.X, p.Y, 11, text, p.Anchor)
			if box.X < f.X || box.Y < f.Y || box.X+box.Width > f.X+f.Width || box.Y+box.Height > f.Y+f.Height {
				continue
			}
			if f.IsDisk() && math.Hypot(math.Abs(box.X+box.Width/2-f.X-f.Width/2)+box.Width/2, math.Abs(box.Y+box.Height/2-f.Y-f.Height/2)+box.Height/2) > f.Width/2-2 {
				continue
			}
			candidates = append(candidates, p)
		}
	}
	if p, ok := labels.Place(text, 11, candidates); ok {
		box := svgchart.LabelTextBox(p.X, p.Y, 11, text, p.Anchor)
		x2 := math.Max(box.X, math.Min(x, box.X+box.Width))
		y2 := math.Max(box.Y, math.Min(y, box.Y+box.Height))
		fmt.Fprintf(&c.b, `<line data-role="time-label-leader" data-time="%s" x1="%.6f" y1="%.6f" x2="%.6f" y2="%.6f" stroke="#667780" stroke-width="0.85"/>`, at, x, y, x2, y2)
		c.text("time-label", p.X, p.Y, 11, p.Anchor, text)
	}
}

func coverageInsideRing(ring []satellite.GeoPoint, p satellite.GeoPoint) bool {
	inside := false
	for i, a := range ring {
		b := ring[(i+1)%len(ring)]
		cross := (p.Longitude-a.Longitude)*(b.Latitude-a.Latitude) - (p.Latitude-a.Latitude)*(b.Longitude-a.Longitude)
		if math.Abs(cross) < 1e-9 && p.Longitude >= math.Min(a.Longitude, b.Longitude)-1e-9 && p.Longitude <= math.Max(a.Longitude, b.Longitude)+1e-9 && p.Latitude >= math.Min(a.Latitude, b.Latitude)-1e-9 && p.Latitude <= math.Max(a.Latitude, b.Latitude)+1e-9 {
			return true
		}
		if (a.Latitude > p.Latitude) != (b.Latitude > p.Latitude) && p.Longitude < a.Longitude+(p.Latitude-a.Latitude)*(b.Longitude-a.Longitude)/(b.Latitude-a.Latitude) {
			inside = !inside
		}
	}
	return inside
}
func coverageInside(regions []satellite.Region, p satellite.GeoPoint) bool {
	for _, r := range regions {
		if !coverageInsideRing(r.Outer, p) {
			continue
		}
		hole := false
		for _, h := range r.Holes {
			hole = hole || coverageInsideRing(h, p)
		}
		if !hole {
			return true
		}
	}
	return false
}
func (c *chart) coveragePanel(s satellite.CoverageSweep, o CoverageSVGOptions, focus int, highlighted bool, f svgmap.Frame, headerBottom float64, compact bool) float64 {
	x, width, y := c.width*.70, c.width*.30-35, math.Max(132, headerBottom+22)
	at := s.Start
	if highlighted {
		at = s.Instants[focus].Time
	}
	rows := []string{satelliteName(s.Elements), fmt.Sprintf("NORAD %d", s.Elements.CatalogNumber), c.words("根数历元 ", "Epoch ") + s.Elements.Epoch.UTC().Format("2006-01-02 15:04:05 UTC"), fmt.Sprintf(c.words("根数龄期 %.1f 小时", "Elements age %.1f hours"), at.Sub(s.Elements.Epoch).Hours())}
	for i, row := range rows {
		role := "panel-value"
		if i == 0 {
			role = "panel-label"
		}
		y = c.wrapped(role, x, y, width, 12, 16, "start", row) + 4
	}
	y = c.wrapped("panel-label", x, y+4, width, 12, 16, "start", c.words("时刻", "Times")+" ("+c.loc.String()+")") + 4
	format := "01-02 " + c.clock
	rows = []string{c.words("开始 ", "Start ") + s.Start.In(c.loc).Format(format), c.words("结束 ", "End ") + s.End.In(c.loc).Format(format)}
	if highlighted && len(s.Instants[focus].Footprint) > 0 {
		rows = append(rows, c.words("当前圈 ", "Current footprint ")+at.In(c.loc).Format(format))
	} else {
		rows = append(rows, c.words("星下点 ", "Subpoint ")+s.Track[focus].Time.In(c.loc).Format(format))
	}
	for _, row := range rows {
		y = c.wrapped("panel-value", x, y, width, 12, 16, "start", row) + 4
	}
	var details []string
	if highlighted {
		v := s.Instants[focus]
		p := v.SubPoint
		details = append(details, fmt.Sprintf(c.words("经度 %.4f°", "Longitude %.4f°"), p.Longitude), fmt.Sprintf(c.words("纬度 %.4f°", "Latitude %.4f°"), p.Latitude), fmt.Sprintf(c.words("高度 %.2f km", "Height %.2f km"), p.HeightKM), fmt.Sprintf(c.words("速度 %.3f km/s", "Speed %.3f km/s"), p.SpeedKMS), c.words("光照 ", "Illumination ")+c.coverageLightName(s.TrackIllumination[focus]))
	}
	if obs := o.Observer; obs != nil {
		regions := s.Swept
		if highlighted {
			regions = coverageRawRegions([][]satellite.GeoPoint{s.Instants[focus].Footprint})
		}
		inside := coverageInside(regions, satellite.GeoPoint{Longitude: obs.Longitude, Latitude: obs.Latitude})
		text := c.words("不在可见区内", "Outside coverage")
		if inside {
			text = c.words("在可见区内", "Inside coverage")
		}
		fmt.Fprintf(&c.b, `<g data-role="observer-membership" data-inside="%t">`, inside)
		details = append(details, c.words("观测站：", "Observer: ")+text)
	}
	bottom := y
	for i, row := range details {
		dx, dy, dw := x, y, width
		if compact {
			dx = 36 + float64(i%2)*((c.width*.67-45)/2+5)
			dy = f.Y + f.Height + 22 + float64(i/2)*22
			dw = (c.width*.67-45)/2 - 5
		}
		end := c.wrapped("panel-value", dx, dy, dw, 12, 16, "start", row)
		bottom = math.Max(bottom, end)
		if !compact {
			y = end + 6
		}
	}
	if o.Observer != nil {
		c.b.WriteString(`</g>`)
	}
	return bottom
}

func (c *chart) coverageTrackLegend() string {
	return c.words("向阳 / 半影 / 本影轨迹", "Sunlit / penumbra / umbra track")
}

func coverageLegendTop(height float64, highlighted bool) float64 {
	if highlighted {
		return height - 117
	}
	return height - 100
}

func (c *chart) coverageLegend(x, top float64, highlighted, currentVisual, compact bool) float64 {
	type legendItem struct {
		role, label, stroke string
		opacity             float64
	}
	items := []legendItem{
		{"coverage-sweep", c.words("扫掠可见区", "Swept coverage"), "#a66f18", .28},
		{"visible-sweep", c.words("目视扫掠区", "Swept visual coverage"), "#a66f18", .45},
	}
	if highlighted {
		items = append(items, legendItem{role: "satellite-footprint"})
	}
	if currentVisual {
		items = append(items, legendItem{"satellite-visible", c.words("当前目视区", "Current visual coverage"), coverageCurrentVisualStroke, .55})
	}
	items = append(items, legendItem{role: "ground-track", label: c.coverageTrackLegend()})
	bottom := top
	for i, item := range items {
		left := x
		baseline := top + float64(i)*17
		if compact {
			left = 36 + float64(i%2)*(c.width/2-18)
			baseline = top + float64(i/2)*17
		}
		bottom = baseline + 4
		if item.role == "satellite-footprint" {
			c.coverageFootprintLegend(left, baseline)
			continue
		}
		if item.stroke != "" {
			fmt.Fprintf(&c.b, `<rect data-role="%s-legend" x="%.4f" y="%.4f" width="18" height="9" fill="#e0ae43" fill-opacity="%.2f" stroke="%s" stroke-width="1.2"/>`, item.role, left, baseline-8, item.opacity, item.stroke)
		} else {
			for j, color := range []string{"#d9ac30", "#da832b", "#505050"} {
				fmt.Fprintf(&c.b, `<path d="M %.4f %.4f h 6" stroke="%s" stroke-width="2" stroke-linecap="round"/>`, left+float64(j)*6, baseline-4, color)
			}
		}
		c.text("legend", left+25, baseline, 11, "start", item.label)
	}
	return bottom
}
func (c *chart) coverageFootprintLegend(x, y float64) {
	fmt.Fprintf(&c.b, `<rect data-role="satellite-footprint-legend" x="%.4f" y="%.4f" width="18" height="9" fill="none" stroke="#a66f18" stroke-width="1.35" stroke-dasharray="5 4"/>`, x, y-8)
	c.text("legend", x+25, y, 11, "start", c.words("当前几何可见圈", "Current geometric footprint"))
}
func (c *chart) coverageLightName(kind satellite.Illumination) string {
	switch kind {
	case satellite.IlluminationSunlit:
		return c.words("向阳", "Sunlit")
	case satellite.IlluminationPenumbra:
		return c.words("半影", "Penumbra")
	default:
		return c.words("本影", "Umbra")
	}
}

type coverageInterval struct{ lo, hi float64 }

func coverageSeamSign(a, b satellite.GeoPoint) int {
	if math.Abs(a.Longitude-b.Longitude) < 1e-9 && math.Abs(math.Abs(a.Longitude)-180) < 1e-9 {
		if a.Longitude > 0 {
			return 1
		}
		return -1
	}
	return 0
}
func coverageSubtract(lo, hi float64, cuts []coverageInterval) []coverageInterval {
	pieces := []coverageInterval{{lo, hi}}
	for _, c := range cuts {
		var next []coverageInterval
		for _, p := range pieces {
			if c.hi <= p.lo || c.lo >= p.hi {
				next = append(next, p)
				continue
			}
			if p.lo < c.lo {
				next = append(next, coverageInterval{p.lo, c.lo})
			}
			if p.hi > c.hi {
				next = append(next, coverageInterval{c.hi, p.hi})
			}
		}
		pieces = next
	}
	return pieces
}
func coverageStrokePaths(regions []satellite.Region) [][]satellite.GeoPoint {
	rings := coverageRings(regions)
	seams := map[int][]coverageInterval{}
	for _, r := range rings {
		for i, a := range r {
			b := r[(i+1)%len(r)]
			if side := coverageSeamSign(a, b); side != 0 {
				seams[side] = append(seams[side], coverageInterval{math.Min(a.Latitude, b.Latitude), math.Max(a.Latitude, b.Latitude)})
			}
		}
	}
	var paths [][]satellite.GeoPoint
	for _, r := range rings {
		var current []satellite.GeoPoint
		flush := func() {
			if len(current) > 1 {
				paths = append(paths, current)
			}
			current = nil
		}
		add := func(a, b satellite.GeoPoint) {
			if len(current) == 0 {
				current = []satellite.GeoPoint{a, b}
			} else if current[len(current)-1] == a {
				current = append(current, b)
			} else {
				flush()
				current = []satellite.GeoPoint{a, b}
			}
		}
		for i, a := range r {
			b := r[(i+1)%len(r)]
			if math.Abs(a.Latitude) == 90 && a.Latitude == b.Latitude {
				flush()
				continue
			}
			if side := coverageSeamSign(a, b); side != 0 {
				pieces := coverageSubtract(math.Min(a.Latitude, b.Latitude), math.Max(a.Latitude, b.Latitude), seams[-side])
				if len(pieces) == 0 {
					flush()
				}
				for _, p := range pieces {
					x, y := p.lo, p.hi
					if b.Latitude < a.Latitude {
						x, y = y, x
					}
					add(satellite.GeoPoint{Longitude: a.Longitude, Latitude: x}, satellite.GeoPoint{Longitude: b.Longitude, Latitude: y})
				}
				continue
			}
			add(a, b)
		}
		flush()
	}
	return paths
}
