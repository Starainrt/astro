package svg

import (
	"github.com/starainrt/astro/internal/svgasset"
	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/satellite"
	"fmt"
	"math"
	"strings"
	"time"
)

type localView struct{ cx, cy, radius, scale, left, top, right, bottom float64 }

func validateTrack(t satellite.TransitTrack) bool {
	tr := t.Transit
	if tr.Center.IsZero() || (tr.Target != satellite.TransitSun && tr.Target != satellite.TransitMoon) || !validNumber(tr.TargetRadiusArcsec) || tr.TargetRadiusArcsec <= 0 || len(t.Frames) < 2 {
		return false
	}
	if !validPoint(t.Observer.Longitude, t.Observer.Latitude) || !validNumber(t.Observer.Height) {
		return false
	}
	if tr.Target == satellite.TransitMoon && (!validNumber(t.MoonIlluminatedFraction) || t.MoonIlluminatedFraction < 0 || t.MoonIlluminatedFraction > 1 || !validNumber(t.MoonBrightLimbAngle)) {
		return false
	}
	for _, value := range []float64{tr.MinSeparationArcsec, tr.Satellite.Azimuth, tr.Satellite.Elevation, tr.Satellite.Dec, tr.Satellite.RangeKM, tr.TargetAltitude} {
		if !validNumber(value) {
			return false
		}
	}
	if tr.Duration < 0 || (!tr.NearMiss && (tr.Start.After(tr.Center) || tr.End.Before(tr.Center))) {
		return false
	}
	found := map[time.Time]bool{}
	for i, f := range t.Frames {
		if f.Time.IsZero() || !validNumber(f.XArcsec) || !validNumber(f.YArcsec) || !validNumber(f.TargetRadiusArcsec) || f.TargetRadiusArcsec <= 0 || !validNumber(f.SatelliteElevation) || !validNumber(f.TargetAltitude) || (i > 0 && !f.Time.After(t.Frames[i-1].Time)) {
			return false
		}
		found[f.Time] = true
	}
	return found[tr.Center] && (tr.NearMiss || (!tr.Start.IsZero() && !tr.End.IsZero() && found[tr.Start] && found[tr.End]))
}

// TransitLocalSVG 绘制已采样的站心凌现象，不调用星历 / render a sampled local transit without ephemeris calls.
func TransitLocalSVG(track satellite.TransitTrack, options TransitLocalSVGOptions) (string, error) {
	if !validateTrack(track) {
		return "", satellite.ErrInvalidInput
	}
	w, h := options.Width, options.Height
	if w == 0 {
		w = 920
	}
	if h == 0 {
		h = 720
	}
	if w < 640 || h < 520 {
		return "", satellite.ErrInvalidInput
	}
	c := chart{width: float64(w), height: float64(h), en: strings.EqualFold(options.Language, "en"), loc: options.Location}
	if c.loc == nil {
		c.loc = track.Transit.Center.Location()
	}
	title := options.Title
	if title == "" {
		title = track.Transit.Center.In(c.loc).Format("2006-01-02") + " " + satelliteName(track.Elements) + " " + c.targetName(track.Transit.Target) + c.words("（本站）", ", Local Circumstances")
		if track.Transit.NearMiss {
			title += c.words("（掠过）", " (near miss)")
		}
	}
	longitude, latitude := "E", "N"
	if track.Observer.Longitude < 0 {
		longitude = "W"
	}
	if track.Observer.Latitude < 0 {
		latitude = "S"
	}
	header := fmt.Sprintf(c.words("观测点 %.4f°%s  %.4f°%s · 椭球高 %.0f m", "Site %.4f°%s  %.4f°%s · elevation %.0f m"), math.Abs(track.Observer.Longitude), longitude, math.Abs(track.Observer.Latitude), latitude, track.Observer.Height)
	heading := chart{width: c.width}
	headerBottom := heading.heading(title, header)
	viewTop := math.Max(130, headerBottom+44)
	contacts := chart{width: c.width, en: c.en, loc: c.loc}
	contactsBottom := contacts.localContacts(track, math.Max(137, headerBottom+42))
	c.height = math.Max(c.height, math.Max(viewTop+390, contactsBottom+50))
	c.openCanvas()
	c.b.WriteString(heading.b.String())
	c.b.WriteString(`<style>[data-role="track-label"]{paint-order:stroke;stroke:white;stroke-width:2px;stroke-linejoin:round}</style><defs>`)
	if track.Transit.Target == satellite.TransitMoon {
		c.b.WriteString(svgasset.MoonFaceSymbol())
	}
	c.b.WriteString(`<radialGradient id="transit-sun"><stop offset="0" stop-color="#fff9cf"/><stop offset=".72" stop-color="#f9dc69"/><stop offset="1" stop-color="#edae36"/></radialGradient><marker id="track-arrow" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="5" markerHeight="5" orient="auto"><path d="M 0 0 L 10 5 L 0 10 Z" fill="#9b3f36"/></marker></defs>`)
	view := localView{left: 35, top: viewTop, right: c.width*.63 - 20, bottom: c.height - 230}
	view.cx = (view.left + view.right) / 2
	view.cy = (view.top + view.bottom) / 2
	view.radius = math.Min((view.right-view.left)/3, (view.bottom-view.top)/3)
	view.scale = view.radius / track.Transit.TargetRadiusArcsec
	c.localOverview(track, view)
	c.b.WriteString(contacts.b.String())
	stagesRight := c.width - 35
	if contactsBottom > c.height-185 {
		stagesRight = view.right
	}
	c.localStages(track, stagesRight)
	note := c.words("卫星按点源绘制；时刻精度取决于根数新旧", "Satellite drawn as a point source; timing accuracy depends on element age")
	if options.FooterNote != "" {
		note += " · " + options.FooterNote
	}
	c.footer(track.Elements, track.Transit.Center, note)
	return c.b.String(), nil
}
func localPoint(v localView, f satellite.TransitFrame) (float64, float64) {
	return v.cx - f.XArcsec*v.scale, v.cy - f.YArcsec*v.scale
}
func (c *chart) disk(track satellite.TransitTrack, cx, cy, radius float64, id string) {
	fmt.Fprintf(&c.b, `<g data-role="target-disk" data-cx="%.6f" data-cy="%.6f" data-radius="%.6f">`, cx, cy, radius)
	if track.Transit.Target == satellite.TransitSun {
		fmt.Fprintf(&c.b, `<circle cx="%.6f" cy="%.6f" r="%.6f" fill="url(#transit-sun)" stroke="#b58c3c"/>`, cx, cy, radius)
	} else {
		fmt.Fprintf(&c.b, `<circle cx="%.6f" cy="%.6f" r="%.6f" fill="#24272a"/>`, cx, cy, radius)
		fmt.Fprintf(&c.b, `<defs><clipPath id="%s"><path d="%s" transform="translate(%.6f %.6f) rotate(%.9f)"/></clipPath></defs><g data-role="moon-phase" data-illumination="%.15g" data-angle="%.15g" clip-path="url(#%s)"><use href="#le-moon" x="%.6f" y="%.6f" width="%.6f" height="%.6f"/></g><circle cx="%.6f" cy="%.6f" r="%.6f" fill="none" stroke="#6a7477"/>`, id, moonPhasePath(radius, track.MoonIlluminatedFraction), cx, cy, -track.MoonBrightLimbAngle-90, track.MoonIlluminatedFraction, track.MoonBrightLimbAngle, id, cx-radius, cy-radius, 2*radius, 2*radius, cx, cy, radius)
	}
	c.b.WriteString(`</g>`)
}
func moonPhasePath(radius, illumination float64) string {
	cosPhase := 2*illumination - 1
	terminatorRadius := radius * math.Abs(cosPhase)
	sweep := 1
	if cosPhase < 0 {
		sweep = 0
	}
	return fmt.Sprintf("M 0 %.3f A %.3f %.3f 0 0 1 0 %.3f A %.3f %.3f 0 0 %d 0 %.3f Z", -radius, radius, radius, radius, terminatorRadius, radius, sweep, -radius)
}
func localFrame(track satellite.TransitTrack, t time.Time) satellite.TransitFrame {
	for _, f := range track.Frames {
		if f.Time.Equal(t) {
			return f
		}
	}
	return satellite.TransitFrame{}
}
func (c *chart) localOverview(track satellite.TransitTrack, v localView) {
	fmt.Fprintf(&c.b, `<g data-role="overview" data-cx="%.9f" data-cy="%.9f" data-scale="%.12g"><defs><clipPath id="track-clip"><rect x="%.6f" y="%.6f" width="%.6f" height="%.6f"/></clipPath></defs>`, v.cx, v.cy, v.scale, v.left, v.top, v.right-v.left, v.bottom-v.top)
	c.text("overview-title", v.left, v.top-15, 14, "start", c.words("目标中心盘面轨迹", "Track relative to target centre"))
	c.disk(track, v.cx, v.cy, v.radius, "overview-phase")
	var points strings.Builder
	for _, f := range track.Frames {
		x, y := localPoint(v, f)
		fmt.Fprintf(&points, "%.6f,%.6f ", x, y)
	}
	dash := ""
	if track.Transit.NearMiss {
		dash = "5 4"
	}
	fmt.Fprintf(&c.b, `<polyline data-role="track" points="%s" fill="none" stroke="#9b3f36" stroke-width="1.3" stroke-dasharray="%s" clip-path="url(#track-clip)"/>`, points.String(), dash)
	middle := (len(track.Frames) - 1) / 2
	f1, f2 := track.Frames[middle], track.Frames[middle+1]
	x1, y1 := localPoint(v, f1)
	x2, y2 := localPoint(v, f2)
	l := math.Hypot(x2-x1, y2-y1)
	if l > 0 {
		fmt.Fprintf(&c.b, `<path data-role="motion" d="M %.6f %.6f L %.6f %.6f" stroke="#9b3f36" marker-end="url(#track-arrow)" clip-path="url(#track-clip)"/>`, x1, y1, x1+(x2-x1)/l*20, y1+(y2-y1)/l*20)
	}
	c.localTrackMarks(track, v)
	c.text("direction", v.left, c.height-209, 11, "start", c.words("北 ↑ · 东 ← ；卫星按点源绘制", "North ↑ · East ← ; satellite shown as a point"))
	c.b.WriteString(`</g>`)
}
func (c *chart) localTrackMarks(track satellite.TransitTrack, v localView) {
	x0, y0 := localPoint(v, track.Frames[0])
	x1, y1 := localPoint(v, track.Frames[len(track.Frames)-1])
	length := math.Hypot(x1-x0, y1-y0)
	nx, ny := 0., -1.
	if length > 0 {
		nx, ny = -(y1-y0)/length, (x1-x0)/length
	}
	if ny > 0 || (ny == 0 && nx < 0) {
		nx, ny = -nx, -ny
	}
	xc, yc := localPoint(v, localFrame(track, track.Transit.Center))
	if nx*(v.cx-xc)+ny*(v.cy-yc) < -1 {
		nx, ny = -nx, -ny
	}
	table := svgchart.LabelTable{}
	events := []struct {
		at          time.Time
		role, label string
	}{{track.Transit.Start, "event-start", c.words("始", "Start")}, {track.Transit.Center, "event-center", c.words("甚", "Centre")}, {track.Transit.End, "event-end", c.words("终", "End")}}
	for _, e := range events {
		if e.at.IsZero() {
			continue
		}
		f := localFrame(track, e.at)
		x, y := localPoint(v, f)
		// 白色外描边保证卫星点落在月球暗面（#24272a）时依然清晰可辨。
		fmt.Fprintf(&c.b, `<circle data-role="%s" cx="%.6f" cy="%.6f" r="4" fill="#24262a" stroke="#ffffff" stroke-width="1.5" clip-path="url(#track-clip)"/>`, e.role, x, y)
		c.localLabel(&table, v, x, y, -nx, -ny, e.label, true)
	}
	c.localTimeMarks(track, v, nx, ny, &table)
}
func (c *chart) localTimeMarks(track satellite.TransitTrack, v localView, nx, ny float64, table *svgchart.LabelTable) {
	first := track.Frames[0].Time.Sub(track.Transit.Center).Seconds()
	last := track.Frames[len(track.Frames)-1].Time.Sub(track.Transit.Center).Seconds()
	speed := localOmega(track) * v.scale
	if speed <= 0 {
		return
	}
	label := svgchart.LabelTextBox(0, 0, 11, "-0.00 s", "middle")
	spacing := math.Max((label.Width*math.Abs(ny)+label.Height*math.Abs(nx)+8)/speed, (last-first)/6)
	base := math.Pow(10, math.Floor(math.Log10(spacing)))
	delta := base * 10
	for _, factor := range []float64{1, 2, 5, 10} {
		if base*factor >= spacing {
			delta = base * factor
			break
		}
	}
	for extent := math.Max(-first, last); delta > extent; delta /= 2 {
	}
	j := 1
	for k := int(math.Ceil(first / delta)); float64(k)*delta <= last; k++ {
		if k == 0 {
			continue
		}
		at := track.Transit.Center.Add(time.Duration(float64(k) * delta * 1e9))
		for j < len(track.Frames)-1 && track.Frames[j].Time.Before(at) {
			j++
		}
		a, b := track.Frames[j-1], track.Frames[j]
		u := float64(at.Sub(a.Time)) / float64(b.Time.Sub(a.Time))
		f := satellite.TransitFrame{XArcsec: a.XArcsec + u*(b.XArcsec-a.XArcsec), YArcsec: a.YArcsec + u*(b.YArcsec-a.YArcsec)}
		x, y := localPoint(v, f)
		if c.localLabel(table, v, x, y, nx, ny, fmt.Sprintf("%+.2f s", float64(k)*delta), false) {
			fmt.Fprintf(&c.b, `<path data-role="tick" d="M %.6f %.6f L %.6f %.6f" stroke="#973a32" stroke-width="1"/>`, x+nx*2, y+ny*2, x+nx*7, y+ny*7)
		}
	}
}
func (c *chart) localLabel(table *svgchart.LabelTable, v localView, x, y, nx, ny float64, value string, required bool) bool {
	box := svgchart.LabelTextBox(0, 0, 11, value, "middle")
	distance := (box.Width*math.Abs(nx)+box.Height*math.Abs(ny))/2 + 10
	cx, cy := x+nx*distance, y+ny*distance
	shifts := []float64{0}
	if required {
		low, high := math.Inf(-1), math.Inf(1)
		for _, axis := range [][4]float64{{cx, -ny, v.left + box.Width/2, v.right - box.Width/2}, {cy, nx, v.top + box.Height/2, v.bottom - box.Height/2}} {
			if axis[1] == 0 {
				continue
			}
			a, b := (axis[2]-axis[0])/axis[1], (axis[3]-axis[0])/axis[1]
			low, high = math.Max(low, math.Min(a, b)), math.Min(high, math.Max(a, b))
		}
		shift := math.Max(low, math.Min(high, 0))
		shifts = []float64{shift, shift + 18, shift - 18, shift + 36, shift - 36, shift + 54, shift - 54}
	}
	for _, shift := range shifts {
		p := svgchart.LabelPlacement{X: cx - ny*shift, Y: cy + nx*shift - box.Y - box.Height/2, Anchor: "middle"}
		b := svgchart.LabelTextBox(p.X, p.Y, 11, value, p.Anchor)
		if b.X < v.left || b.Y < v.top || b.X+b.Width > v.right || b.Y+b.Height > v.bottom {
			continue
		}
		if _, ok := table.Place(value, 11, []svgchart.LabelPlacement{p}); !ok {
			continue
		}
		if required {
			fmt.Fprintf(&c.b, `<path data-role="event-leader" d="M %.6f %.6f L %.6f %.6f" stroke="#7c8588" stroke-width=".6"/>`, x+nx*6, y+ny*6, p.X-nx*box.Height/2, p.Y+box.Y+box.Height/2-ny*box.Height/2)
		}
		c.text("track-label", p.X, p.Y, 11, p.Anchor, value)
		return true
	}
	return false
}
func localOmega(track satellite.TransitTrack) float64 {
	index := 1
	for index < len(track.Frames)-1 && !track.Frames[index].Time.After(track.Transit.Center) {
		index++
	}
	a, b := track.Frames[index-1], track.Frames[index]
	return math.Hypot(b.XArcsec-a.XArcsec, b.YArcsec-a.YArcsec) / b.Time.Sub(a.Time).Seconds()
}
func (c *chart) localContacts(track satellite.TransitTrack, y float64) float64 {
	tr := track.Transit
	x, w := c.width*.66, c.width*.34-34
	line := 17.
	size := 12.
	c.text("contacts-title", x, y-20, 14, "start", c.words("接触时刻", "Local circumstances"))
	for i, at := range []time.Time{tr.Start, tr.Center, tr.End} {
		label := []string{c.words("始", "Start"), c.words("甚", "Centre"), c.words("终", "End")}[i]
		if at.IsZero() {
			c.text("contact", x, y, size, "start", label+" —")
			y += 3 * line
			continue
		}
		f := localFrame(track, at)
		c.text("contact", x, y, size, "start", label+" · "+at.In(c.loc).Format("15:04:05.000 MST"))
		y += line
		c.text("contact", x, y, size, "start", fmt.Sprintf(c.words("方位 %.1f° · 仰角 %.1f°", "Az %.1f° · Alt %.1f°"), localAzimuth(track, f), f.SatelliteElevation))
		y += line
		c.text("contact", x, y, size, "start", fmt.Sprintf(c.words("目标 %.1f° · 位置角 %.1f°", "Target %.1f° · PA %.1f°"), f.TargetAltitude, math.Mod(math.Atan2(f.XArcsec, f.YArcsec)*180/math.Pi+360, 360)))
		y += line + 4
	}
	offset := c.words("中心射线未落地", "Centre ray misses Earth")
	if validNumber(tr.CenterlineOffsetKM) {
		side := c.words("右侧", "right")
		if tr.CenterlineOffsetKM > 0 {
			side = c.words("左侧", "left")
		}
		offset = fmt.Sprintf(c.words("中心线%s %.2f km", "Centre-line %s %.2f km"), side, math.Abs(tr.CenterlineOffsetKM))
	}
	texts := []string{fmt.Sprintf(c.words("时长 %.1f ms", "Duration %.1f ms"), float64(tr.Duration)/1e6), fmt.Sprintf(c.words("最小角距 %.2f″", "Minimum separation %.2f″"), tr.MinSeparationArcsec), fmt.Sprintf(c.words("目标视半径 %.2f″", "Target radius %.2f″"), tr.TargetRadiusArcsec), fmt.Sprintf(c.words("斜距 %.2f km", "Range %.2f km"), tr.Satellite.RangeKM), fmt.Sprintf(c.words("相对角速度 %.2f″/s", "Relative speed %.2f″/s"), localOmega(track)), offset}
	for _, text := range texts {
		y = c.wrapped("statistic", x, y, w, size, line, "start", text)
	}
	return y
}
func (c *chart) localStages(track satellite.TransitTrack, right float64) {
	tr := track.Transit
	at := []time.Time{tr.Start, tr.Center, tr.End}
	labels := []string{c.words("始", "Start"), c.words("甚", "Centre"), c.words("终", "End")}
	if tr.NearMiss {
		at = []time.Time{tr.Center}
		labels = []string{c.words("甚（掠过）", "Centre (near miss)")}
	}
	c.text("stage-title", 35, c.height-189, 14, "start", c.words("始 · 甚 · 终", "Stages of the transit"))
	radius := math.Min(42, (c.height-300)/7)
	for i, t := range at {
		cx := 35 + (right-35)*(float64(i)+.5)/float64(len(at))
		cy := c.height - 140
		f := localFrame(track, t)
		scale := radius / tr.TargetRadiusArcsec
		fmt.Fprintf(&c.b, `<g data-role="stage-panel" data-time="%s" data-cx="%.6f" data-cy="%.6f" data-scale="%.12g">`, t.Format(time.RFC3339Nano), cx, cy, scale)
		c.disk(track, cx, cy, f.TargetRadiusArcsec*scale, fmt.Sprintf("stage-phase-%d", i))
		fmt.Fprintf(&c.b, `<circle data-role="stage-satellite" cx="%.6f" cy="%.6f" r="4" fill="#24262a" stroke="#ffffff" stroke-width="1.5"/>`, cx-f.XArcsec*scale, cy-f.YArcsec*scale)
		c.text("stage-label", cx, c.height-72, 12, "middle", labels[i])
		c.text("stage-time", cx, c.height-57, 11, "middle", t.In(c.loc).Format("15:04:05.000 MST"))
		c.b.WriteString(`</g>`)
	}
}
func localAzimuth(track satellite.TransitTrack, f satellite.TransitFrame) float64 {
	if f.Time.Equal(track.Transit.Center) {
		return track.Transit.Satellite.Azimuth
	}
	const rad = math.Pi / 180
	arc := rad / 3600
	center := localFrame(track, track.Transit.Center)
	x, y := center.XArcsec*arc, center.YArcsec*arc
	sd := track.Transit.Satellite.Dec * rad
	dt := math.Asin(math.Max(-1, math.Min(1, math.Sin(sd)*math.Sqrt(1+x*x+y*y)/math.Sqrt(1+y*y)))) - math.Atan(y)
	lat := track.Observer.Latitude * rad
	az := track.Transit.Satellite.Azimuth * rad
	el := track.Transit.Satellite.Elevation * rad
	hs := math.Atan2(-math.Cos(el)*math.Sin(az), (math.Sin(el)-math.Sin(lat)*math.Sin(sd))/math.Cos(lat))
	ht := hs + math.Atan2(x, math.Cos(dt)-y*math.Sin(dt))
	b0 := math.Cos(lat) * math.Sin(ht)
	c0 := math.Sin(lat)*math.Cos(dt) - math.Cos(lat)*math.Sin(dt)*math.Cos(ht)
	x, y = f.XArcsec*arc, f.YArcsec*arc
	rr := x*x + y*y
	if rr < 1e-20 {
		return track.Transit.Satellite.Azimuth
	}
	a := math.Sin(f.TargetAltitude * rad)
	z := math.Sin(f.SatelliteElevation*rad)*math.Sqrt(1+rr) - a
	k := math.Sqrt(math.Max(0, (1-a*a)-z*z/rr) / rr)
	bestB, bestC := z*x/rr-k*y, z*y/rr+k*x
	b2, c2 := z*x/rr+k*y, z*y/rr-k*x
	if math.Hypot(b2-b0, c2-c0) < math.Hypot(bestB-b0, bestC-c0) {
		bestB, bestC = b2, c2
	}
	denom := math.Hypot(a, bestC)
	if denom == 0 {
		return track.Transit.Satellite.Azimuth
	}
	phase := math.Atan2(bestC, a)
	v := math.Asin(math.Max(-1, math.Min(1, math.Sin(lat)/denom)))
	decl := math.Remainder(v-phase, 2*math.Pi)
	other := math.Remainder(math.Pi-v-phase, 2*math.Pi)
	if math.Abs(other-dt) < math.Abs(decl-dt) {
		decl = other
	}
	hour := math.Atan2(bestB, (a-math.Sin(lat)*math.Sin(decl))/math.Cos(decl))
	raShift := math.Atan2(x, math.Cos(decl)-y*math.Sin(decl))
	satDec := math.Asin(math.Max(-1, math.Min(1, (math.Sin(decl)+y*math.Cos(decl))/math.Sqrt(1+rr))))
	hour -= raShift
	return math.Mod(math.Atan2(-math.Cos(satDec)*math.Sin(hour), math.Sin(satDec)*math.Cos(lat)-math.Cos(satDec)*math.Cos(hour)*math.Sin(lat))/rad+360, 360)
}
