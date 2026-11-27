package svg

import (
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/internal/svgmap"
	"github.com/starainrt/astro/satellite"
	"fmt"
	"html"
	"math"
	"reflect"
	"strings"
	"time"
)

type chart struct {
	b             strings.Builder
	width, height float64
	en            bool
	loc           *time.Location
	clock         string
}

func (c *chart) words(zh, en string) string {
	if c.en {
		return en
	}
	return zh
}
func validNumber(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func validPoint(lon, lat float64) bool {
	return validNumber(lon) && validNumber(lat) && math.Abs(lon) <= 180 && math.Abs(lat) <= 90
}
func (c *chart) text(role string, x, y, size float64, anchor, text string) {
	family, weight, fill := "Arial, sans-serif", "400", "#34464b"
	if role == "title" || role == "header" || role == "footer" || strings.HasSuffix(role, "title") {
		family = "Georgia, 'Times New Roman', serif"
	}
	if role == "title" || strings.HasSuffix(role, "title") || role == "panel-label" || role == "stage-label" {
		weight = "700"
		fill = "#161a1b"
	}
	fmt.Fprintf(&c.b, `<text data-role="%s" x="%.4f" y="%.4f" font-family="%s" font-size="%.2f" font-weight="%s" text-anchor="%s" fill="%s">%s</text>`, role, x, y, family, size, weight, anchor, fill, html.EscapeString(text))
}
func (c *chart) wrapped(role string, x, y, width, size, leading float64, anchor, value string) float64 {
	for _, s := range svgchart.WrapText(value, width, size) {
		c.text(role, x, y, size, anchor, s)
		y += leading
	}
	return y
}
func (c *chart) open(title, subtitle string) float64 {
	c.openCanvas()
	return c.heading(title, subtitle)
}
func (c *chart) openCanvas() {
	fmt.Fprintf(&c.b, `<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f"><rect width="100%%" height="100%%" fill="#efefed"/><rect x="14" y="14" width="%.0f" height="%.0f" rx="5" fill="white"/>`, c.width, c.height, c.width, c.height, c.width-28, c.height-28)
}
func (c *chart) heading(title, subtitle string) float64 {
	size := 26.
	for size > 16 && svgchart.EstimatedTextWidth(title, size) > c.width-80 {
		size--
	}
	y := c.wrapped("title", c.width/2, 44, c.width-80, size, size+5, "middle", title)
	fmt.Fprintf(&c.b, `<path d="M %.2f %.2f H %.2f" stroke="#555"/>`, c.width/2-78, y-16, c.width/2+78)
	return c.wrapped("header", c.width/2, y+12, c.width-70, 14, 19, "middle", subtitle) - 19
}
func (c *chart) footer(e satellite.Elements, at time.Time, note string) {
	a := fmt.Sprintf(c.words("根数历元 %s UTC · 龄期 %.1f 小时", "Elements epoch %s UTC · age %.1f hours"), e.Epoch.UTC().Format("2006-01-02 15:04:05"), at.Sub(e.Epoch).Hours())
	c.wrapped("footer", c.width/2, c.height-40, c.width-70, 11, 14, "middle", a)
	if note != "" {
		c.wrapped("footer", c.width/2, c.height-22, c.width-70, 10, 12, "middle", note)
	}
	c.b.WriteString(`</svg>`)
}
func satelliteName(e satellite.Elements) string {
	if e.Name != "" {
		return e.Name
	}
	return fmt.Sprintf("NORAD %d", e.CatalogNumber)
}
func (c *chart) targetName(t satellite.TransitTarget) string {
	if t == satellite.TransitMoon {
		return c.words("凌月", "Lunar Transit")
	}
	return c.words("凌日", "Solar Transit")
}
func validPathPoint(q satellite.TransitPathPoint) bool {
	return validPoint(q.Longitude, q.Latitude) && !q.Time.IsZero() && validNumber(q.TargetAltitude) && validNumber(q.HalfWidthKM) && q.HalfWidthKM >= 0
}
func validateMap(p satellite.TransitPath, o TransitMapSVGOptions) bool {
	if len(p.CenterLine) < 2 || len(p.LeftLimit) < 2 || len(p.RightLimit) < 2 || p.Event.Start.IsZero() || p.Event.Greatest.Before(p.Event.Start) || p.Event.End.Before(p.Event.Greatest) || (p.Event.Target != satellite.TransitSun && p.Event.Target != satellite.TransitMoon) {
		return false
	}
	if !validPathPoint(p.Event.StartPoint) || !validPathPoint(p.Event.GreatestPoint) || !validPathPoint(p.Event.EndPoint) {
		return false
	}
	if o.TimeLabelStep < 0 || (o.TimeLabelStep > 0 && p.Event.End.Sub(p.Event.Start)/o.TimeLabelStep > 2000) {
		return false
	}
	for _, line := range [][]satellite.TransitPathPoint{p.CenterLine, p.LeftLimit, p.RightLimit} {
		for i, q := range line {
			if !validPoint(q.Longitude, q.Latitude) || q.Time.IsZero() || (i > 0 && !q.Time.After(line[i-1].Time)) {
				return false
			}
		}
	}
	for _, ring := range [][]satellite.GeoPoint{p.StartHorizon, p.EndHorizon} {
		for _, q := range ring {
			if !validPoint(q.Longitude, q.Latitude) {
				return false
			}
		}
	}
	if o.Observer != nil && (!validPoint(o.Observer.Longitude, o.Observer.Latitude) || !validNumber(o.Observer.Height)) {
		return false
	}
	return o.Nearby == nil || (o.Observer != nil && reflect.DeepEqual(o.Nearby.Event, p.Event) && validNumber(o.Nearby.NearestDistanceKM) && o.Nearby.NearestDistanceKM >= 0 && validNumber(o.Nearby.DistanceKM) && validPoint(o.Nearby.Nearest.Longitude, o.Nearby.Nearest.Latitude) && validNumber(o.Nearby.Nearest.HalfWidthKM) && o.Nearby.Nearest.HalfWidthKM >= 0)
}

// TransitMapSVG 绘制已计算的凌现象带，不调用星历 / render a computed transit band without ephemeris calls.
func TransitMapSVG(path satellite.TransitPath, options TransitMapSVGOptions) (string, error) {
	if !validateMap(path, options) {
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
	switch options.Projection {
	case MapProjectionAuto, MapProjectionEquirectangular, MapProjectionNorthPolar, MapProjectionSouthPolar, MapProjectionOrthographic:
	default:
		return "", satellite.ErrInvalidInput
	}
	c := chart{width: float64(w), height: float64(h), en: strings.EqualFold(options.Language, "en"), loc: options.Location}
	if c.loc == nil {
		c.loc = path.Event.Start.Location()
	}
	title := options.Title
	if title == "" {
		title = path.Event.Greatest.In(c.loc).Format("2006-01-02") + " " + satelliteName(path.Elements) + " " + c.targetName(path.Event.Target) + c.words("带", " Path")
	}
	headerBottom := c.open(title, c.words("地面中心线与目标圆盘边界 · WGS84", "Ground centre line and target-disk limits · WGS84"))
	extraY := 0.0
	if headerBottom > 108 {
		extraY = headerBottom - 108
	}
	frame := mapFrame(path, options, c.width, c.height)
	if extraY > 0 {
		frame.Y += extraY
		frame.Height = math.Max(120, frame.Height-extraY)
		if frame.Projection != svgmap.ProjectionEquirectangular {
			s := math.Min(frame.Width, frame.Height)
			frame.X += (frame.Width - s) / 2
			frame.Y += (frame.Height - s) / 2
			frame.Width = s
			frame.Height = s
		}
	}
	fmt.Fprintf(&c.b, `<defs>%s</defs><g data-role="map" data-x="%.6f" data-y="%.6f" data-width="%.6f" data-height="%.6f" data-projection="%s" data-longitude="%.9f" data-latitude="%.9f">`, frame.ClipDefinition("map-clip"), frame.X, frame.Y, frame.Width, frame.Height, frame.Projection, frame.CenterLongitude, frame.CenterLatitude)
	frame.WriteOcean(&c.b)
	frame.WriteLand(&c.b, "map-clip")
	frame.WriteGraticule(&c.b, "map-clip")
	c.mapLayers(path, frame, options)
	frame.WriteFrame(&c.b)
	c.b.WriteString(`</g>`)
	c.mapPanel(path, frame, options, extraY)
	c.footer(path.Elements, path.Event.Greatest, options.FooterNote)
	return c.b.String(), nil
}
func mapFrame(p satellite.TransitPath, o TransitMapSVGOptions, w, h float64) svgmap.Frame {
	min, max := 90., -90.
	for _, q := range p.CenterLine {
		min = math.Min(min, q.Latitude)
		max = math.Max(max, q.Latitude)
	}
	proj := svgmap.ResolveProjection(svgmap.Projection(o.Projection), p.Event.GreatestPoint.Latitude, min, max)
	f := svgmap.Frame{X: 36, Y: 145, Width: w*.67 - 45, Height: h - 245, Projection: proj, CenterLatitude: p.Event.GreatestPoint.Latitude}
	if proj != svgmap.ProjectionEquirectangular {
		f.CenterLongitude = p.Event.GreatestPoint.Longitude
		s := math.Min(f.Width, f.Height)
		f.X += (f.Width - s) / 2
		f.Y += (f.Height - s) / 2
		f.Width = s
		f.Height = s
	} else {
		f.Height = math.Min(f.Height, f.Width/2)
		f.Y = 145 + (h-245-f.Height)/2
	}
	return f
}
func pathGeo(p []satellite.TransitPathPoint) []svgmap.GeoPoint {
	out := make([]svgmap.GeoPoint, len(p))
	for i, q := range p {
		out[i] = svgmap.GeoPoint{Longitude: q.Longitude, Latitude: q.Latitude}
	}
	return out
}
func (c *chart) geoLine(role string, points []svgmap.GeoPoint, f svgmap.Frame, color, dash string, width float64) {
	for _, line := range svgmap.PolylineSegments(points, f.Clip()) {
		var s strings.Builder
		for _, q := range line {
			x, y, ok := f.Project(q.Longitude, q.Latitude)
			if ok {
				fmt.Fprintf(&s, "%.6f,%.6f ", x, y)
			}
		}
		if s.Len() > 0 {
			fmt.Fprintf(&c.b, `<polyline data-role="%s" points="%s" fill="none" stroke="%s" stroke-width="%.2f" stroke-dasharray="%s" clip-path="url(#map-clip)"/>`, role, s.String(), color, width, dash)
		}
	}
}
func (c *chart) mapLayers(p satellite.TransitPath, f svgmap.Frame, o TransitMapSVGOptions) {
	ring := pathGeo(p.LeftLimit)
	right := pathGeo(p.RightLimit)
	for i := len(right) - 1; i >= 0; i-- {
		ring = append(ring, right[i])
	}
	for _, poly := range svgmap.PolygonFragments(ring, f.Clip()) {
		var s strings.Builder
		for _, q := range poly {
			x, y, ok := f.Project(q.Longitude, q.Latitude)
			if ok {
				fmt.Fprintf(&s, "%.6f,%.6f ", x, y)
			}
		}
		fmt.Fprintf(&c.b, `<polygon data-role="transit-band" points="%s" fill="#e0ae43" fill-opacity=".30" clip-path="url(#map-clip)"/>`, s.String())
	}
	c.geoLine("transit-band", pathGeo(p.CenterLine), f, "#e0ae43", "", 2)
	c.geoLine("left-limit", pathGeo(p.LeftLimit), f, "#b77828", "", .8)
	c.geoLine("right-limit", pathGeo(p.RightLimit), f, "#b77828", "", .8)
	for _, r := range [][]satellite.GeoPoint{p.StartHorizon, p.EndHorizon} {
		v := make([]svgmap.GeoPoint, len(r))
		for i, q := range r {
			v[i] = svgmap.GeoPoint{Longitude: q.Longitude, Latitude: q.Latitude}
		}
		c.geoLine("target-horizon", v, f, "#78888b", "4 4", 1)
	}
	c.geoLine("center-line", pathGeo(p.CenterLine), f, "#ae3f35", "", 1.2)
	var labels svgchart.LabelTable
	for i, q := range []satellite.TransitPathPoint{p.Event.StartPoint, p.Event.GreatestPoint, p.Event.EndPoint} {
		x, y, ok := f.Project(q.Longitude, q.Latitude)
		if !ok {
			continue
		}
		role := []string{"event-start", "event-greatest", "event-end"}[i]
		fmt.Fprintf(&c.b, `<circle data-role="%s" cx="%.6f" cy="%.6f" r="3" fill="#8c302b"/>`, role, x, y)
		label := []string{c.words("始", "Start"), c.words("甚", "Greatest"), c.words("终", "End")}[i]
		c.mapLabel(&labels, f, x, y, label, "event-label")
	}
	step := o.TimeLabelStep
	if step == 0 {
		for _, s := range []int{1, 2, 5, 10, 15, 30, 60, 120, 300} {
			step = time.Duration(s) * time.Second
			if p.Event.End.Sub(p.Event.Start)/step <= 12 {
				break
			}
		}
	}
	if step > 0 {
		index := 1
		for at := p.Event.Start.Add(step); at.Before(p.Event.End); at = at.Add(step) {
			for index < len(p.CenterLine)-1 && p.CenterLine[index].Time.Before(at) {
				index++
			}
			a, b := p.CenterLine[index-1], p.CenterLine[index]
			v := float64(at.Sub(a.Time)) / float64(b.Time.Sub(a.Time))
			lon := math.Remainder(a.Longitude+v*math.Remainder(b.Longitude-a.Longitude, 360), 360)
			lat := a.Latitude + v*(b.Latitude-a.Latitude)
			x, y, ok := f.Project(lon, lat)
			if ok {
				c.mapLabel(&labels, f, x, y, at.In(c.loc).Format("15:04:05"), "time-label")
			}
		}
	}
	if o.Observer != nil {
		x, y, ok := f.Project(o.Observer.Longitude, o.Observer.Latitude)
		if ok {
			c.star("observer", x, y, 5)
			c.mapLabel(&labels, f, x, y, c.words("观测点", "Observer"), "observer-label")
		}
	}
}
func (c *chart) mapLabel(table *svgchart.LabelTable, f svgmap.Frame, x, y float64, label, role string) {
	candidates := svgchart.LabelCandidates(x+8, y-9, "start", 16)
	filtered := candidates[:0]
	for _, p := range candidates {
		box := svgchart.LabelTextBox(p.X, p.Y, 11, label, p.Anchor)
		if box.X >= f.X && box.Y >= f.Y && box.X+box.Width <= f.X+f.Width && box.Y+box.Height <= f.Y+f.Height {
			filtered = append(filtered, p)
		}
	}
	if p, ok := table.Place(label, 11, filtered); ok {
		c.text(role, p.X, p.Y, 11, p.Anchor, label)
	}
}
func (c *chart) star(role string, x, y, r float64) {
	var pts strings.Builder
	for i := 0; i < 10; i++ {
		rr := r
		if i%2 == 1 {
			rr *= .4
		}
		a := float64(i)*math.Pi/5 - math.Pi/2
		fmt.Fprintf(&pts, "%.4f,%.4f ", x+rr*math.Cos(a), y+rr*math.Sin(a))
	}
	fmt.Fprintf(&c.b, `<polygon data-role="%s" points="%s" fill="#343e71"/>`, role, pts.String())
}
func (c *chart) mapPanel(p satellite.TransitPath, f svgmap.Frame, o TransitMapSVGOptions, extraY float64) {
	x, w := c.width*.70, c.width*.30-35
	y := 130. + extraY
	compact := math.Min(1, math.Max(.65, (c.height-200)/600))
	rows := []struct {
		label string
		q     satellite.TransitPathPoint
	}{{c.words("凌始", "Start"), p.Event.StartPoint}, {c.words("凌甚", "Greatest"), p.Event.GreatestPoint}, {c.words("凌终", "End"), p.Event.EndPoint}}
	for _, row := range rows {
		y = c.wrapped("panel-label", x, y, w, 12, 16, "start", row.label+" · "+row.q.Time.In(c.loc).Format("15:04:05.000 MST"))
		c.text("panel-value", x, y, 12, "start", fmt.Sprintf("%.4f°  %.4f°", row.q.Longitude, row.q.Latitude))
		y += math.Max(18, 25*compact)
	}
	var length float64
	for i := 1; i < len(p.CenterLine); i++ {
		d, _ := mapInverse(p.CenterLine[i-1].Longitude, p.CenterLine[i-1].Latitude, p.CenterLine[i].Longitude, p.CenterLine[i].Latitude)
		length += d
	}
	statsX, statsY, statsWidth := x, y, w
	if c.height < 600 && o.Nearby != nil {
		statsX, statsY, statsWidth = f.X, f.Y+f.Height+20, f.Width
	}
	for _, text := range []string{fmt.Sprintf(c.words("甚时目标高度 %.2f°", "Target altitude %.2f°"), p.Event.GreatestPoint.TargetAltitude), fmt.Sprintf(c.words("甚时带宽 %.2f km · 带长 %.0f km", "Width at greatest %.2f km · length %.0f km"), 2*p.Event.GreatestPoint.HalfWidthKM, length)} {
		statsY = c.wrapped("panel-value", statsX, statsY, statsWidth, 12, 16, "start", text) + 4
	}
	if statsX == x {
		y = statsY
	}
	if !p.Event.Complete {
		y = c.wrapped("panel-value", x, y, w, 11, 15, "start", c.words("被查询窗口截断", "Clipped by query window")) + 5
	}
	if o.Nearby != nil {
		size := math.Min(math.Min(w, 210), (c.height-220)*.35)
		top := math.Max(y+30, c.height-135-size)
		size = math.Min(size, c.height-115-top)
		if size < 40 {
			top = math.Max(y+12, c.height-115-math.Min(w, 90))
			size = math.Min(math.Min(w, 210), c.height-115-top)
		}
		if size >= 24 {
			c.inset(p, *o.Observer, *o.Nearby, x, top, size)
		}
	}
}
func mapInverse(lon1, lat1, lon2, lat2 float64) (float64, float64) {
	const rad = math.Pi / 180
	const flatten = 1 / 298.257223563
	a, b := math.Atan((1-flatten)*math.Tan(lat1*rad)), math.Atan((1-flatten)*math.Tan(lat2*rad))
	sa, ca := math.Sincos(a)
	sb, cb := math.Sincos(b)
	l := math.Remainder(lon2-lon1, 360) * rad
	lambda := l
	var ss, cs, sigma, alpha, c2, c2m float64
	for i := 0; i < 100; i++ {
		sl, cl := math.Sincos(lambda)
		ss = math.Hypot(cb*sl, ca*sb-sa*cb*cl)
		if ss == 0 {
			return 0, 0
		}
		cs = sa*sb + ca*cb*cl
		sigma = math.Atan2(ss, cs)
		alpha = ca * cb * sl / ss
		c2 = 1 - alpha*alpha
		c2m = 0
		if c2 > 1e-15 {
			c2m = cs - 2*sa*sb/c2
		}
		c := flatten / 16 * c2 * (4 + flatten*(4-3*c2))
		next := l + (1-c)*flatten*alpha*(sigma+c*ss*(c2m+c*cs*(-1+2*c2m*c2m)))
		if math.Abs(next-lambda) < 1e-12 {
			break
		}
		lambda = next
	}
	u := c2 * (1/(1-flatten)/(1-flatten) - 1)
	aa := 1 + u/16384*(4096+u*(-768+u*(320-175*u)))
	bb := u / 1024 * (256 + u*(-128+u*(74-47*u)))
	delta := bb * ss * (c2m + bb/4*(cs*(-1+2*c2m*c2m)-bb/6*c2m*(-3+4*ss*ss)*(-3+4*c2m*c2m)))
	return 6378.137 * (1 - flatten) * aa * (sigma - delta), math.Atan2(cb*math.Sin(lambda), ca*sb-sa*cb*math.Cos(lambda)) / rad
}
func (c *chart) inset(p satellite.TransitPath, obs satellite.Observer, n satellite.NearbyTransit, x, y, size float64) {
	// 方位等距坐标直接用WGS84反算距离与方位，保留端点连线的真实方向。
	side := math.Max(20, math.Max(4*n.NearestDistanceKM, 4*n.Nearest.HalfWidthKM))
	cx, cy := x+size/2, y+size/2
	scale := size / side
	fmt.Fprintf(&c.b, `<defs><clipPath id="inset-clip"><rect x="%.4f" y="%.4f" width="%.4f" height="%.4f"/></clipPath></defs><g data-role="inset" data-cx="%.6f" data-cy="%.6f" data-scale="%.9f" data-side-km="%.9f"><rect x="%.4f" y="%.4f" width="%.4f" height="%.4f" fill="#f5f7f6" stroke="#c5cccc"/>`, x, y, size, size, cx, cy, scale, side, x, y, size, size)
	project := func(lon, lat float64) (float64, float64) {
		d, a := mapInverse(obs.Longitude, obs.Latitude, lon, lat)
		return cx + d*math.Sin(a*math.Pi/180)*scale, cy - d*math.Cos(a*math.Pi/180)*scale
	}
	fmt.Fprintf(&c.b, `<g clip-path="url(#inset-clip)">`)
	for i := -2; i <= 2; i++ {
		v := float64(i) * size / 4
		fmt.Fprintf(&c.b, `<path data-role="km-grid" d="M %.4f %.4f V %.4f M %.4f %.4f H %.4f" fill="none" stroke="#d4dedd"/>`, cx+v, y, y+size, x, cy+v, x+size)
	}
	var band strings.Builder
	for _, q := range p.LeftLimit {
		a, b := project(q.Longitude, q.Latitude)
		fmt.Fprintf(&band, "%.6f,%.6f ", a, b)
	}
	for i := len(p.RightLimit) - 1; i >= 0; i-- {
		q := p.RightLimit[i]
		a, b := project(q.Longitude, q.Latitude)
		fmt.Fprintf(&band, "%.6f,%.6f ", a, b)
	}
	fmt.Fprintf(&c.b, `<polygon data-role="inset-band" points="%s" fill="#e0ae43" fill-opacity=".3"/>`, band.String())
	for j, line := range [][]satellite.TransitPathPoint{p.CenterLine, p.LeftLimit, p.RightLimit} {
		var pts strings.Builder
		for _, q := range line {
			a, b := project(q.Longitude, q.Latitude)
			fmt.Fprintf(&pts, "%.6f,%.6f ", a, b)
		}
		fmt.Fprintf(&c.b, `<polyline data-role="inset-%s" points="%s" fill="none" stroke="%s" stroke-width="1.2"/>`, []string{"center-line", "left-limit", "right-limit"}[j], pts.String(), []string{"#ae3f35", "#b77828", "#b77828"}[j])
	}
	a, b := project(n.Nearest.Longitude, n.Nearest.Latitude)
	fmt.Fprintf(&c.b, `<path data-role="inset-connector" d="M %.6f %.6f L %.6f %.6f" stroke="#444875" stroke-dasharray="3 3" data-distance-km="%.12g"/>`, cx, cy, a, b, n.NearestDistanceKM)
	c.star("inset-observer", cx, cy, 5)
	fmt.Fprintf(&c.b, `<circle data-role="inset-nearest" cx="%.6f" cy="%.6f" r="3" fill="#ae3f35"/>`, a, b)
	i := 1
	for i < len(p.CenterLine)-1 && p.CenterLine[i].Time.Before(n.Nearest.Time) {
		i++
	}
	a0, b0 := project(p.CenterLine[i-1].Longitude, p.CenterLine[i-1].Latitude)
	a1, b1 := project(p.CenterLine[i].Longitude, p.CenterLine[i].Latitude)
	length := math.Hypot(a1-a0, b1-b0)
	if length > 0 {
		dx, dy := (a1-a0)/length, (b1-b0)/length
		fmt.Fprintf(&c.b, `<path data-role="inset-motion" d="M %.5f %.5f L %.5f %.5f M %.5f %.5f L %.5f %.5f L %.5f %.5f" fill="none" stroke="#ae3f35"/>`, a-10*dx, b-10*dy, a+10*dx, b+10*dy, a+5*dx-3*dy, b+5*dy+3*dx, a+10*dx, b+10*dy, a+5*dx+3*dy, b+5*dy-3*dx)
	}
	c.b.WriteString(`</g></g>`)
	text := fmt.Sprintf(c.words("距最近点 %.2f km", "Nearest point %.2f km"), n.NearestDistanceKM)
	c.wrapped("inset-label", x, y-25, c.width*.30-35, 11, 14, "start", text)
	c.text("inset-scale", x+size-4, y+15, 10, "end", fmt.Sprintf("%.0f km", side))
	label := c.words("横向偏移 0 km", "Cross-track offset 0 km")
	if n.DistanceKM != 0 {
		sideName := c.words("右侧", "right")
		if n.DistanceKM > 0 {
			sideName = c.words("左侧", "left")
		}
		label = fmt.Sprintf(c.words("横向%s %.2f km", "Cross-track %s %.2f km"), sideName, math.Abs(n.DistanceKM))
	}
	next := c.wrapped("inset-label", x, y+size+15, c.width*.30-35, 11, 14, "start", label)
	c.text("inset-label", x, next, 11, "start", n.Nearest.Time.In(c.loc).Format("15:04:05.000 MST"))
}
