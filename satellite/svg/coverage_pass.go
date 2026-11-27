package svg

import (
	"fmt"
	"html"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/starainrt/astro/internal/svgchart"
	"github.com/starainrt/astro/satellite"
)

func coveragePassesValid(start, end time.Time, passes []CoveragePass) bool {
	for _, pass := range passes {
		if pass.Rise.Time.Before(start) || pass.Set.Time.After(end) || pass.Culmination.Time.Before(pass.Rise.Time) || pass.Set.Time.Before(pass.Culmination.Time) {
			return false
		}
		start = pass.Set.Time
		for _, look := range [...]satellite.Look{pass.Rise, pass.Culmination, pass.Set} {
			if !validNumber(look.Azimuth) || !validNumber(look.Elevation) || !validNumber(look.RangeKM) || math.IsInf(look.Magnitude, 0) {
				return false
			}
		}
		inside := func(at time.Time) bool { return !at.Before(pass.Rise.Time) && !at.After(pass.Set.Time) }
		if (!pass.VisibleStart.IsZero() || !pass.VisibleEnd.IsZero()) && (!inside(pass.VisibleStart) || !inside(pass.VisibleEnd) || pass.VisibleEnd.Before(pass.VisibleStart)) {
			return false
		}
		if !pass.BrightestTime.IsZero() && (!inside(pass.BrightestTime) || !validNumber(pass.BrightestMagnitude)) {
			return false
		}
		for index, sample := range pass.Samples {
			if !inside(sample.Time) || math.IsInf(sample.Magnitude, 0) || index > 0 && !sample.Time.After(pass.Samples[index-1].Time) {
				return false
			}
		}
	}
	return true
}

func (c *chart) coveragePassPanel(passes []CoveragePass, y float64) float64 {
	c.b.WriteString(`<g data-role="coverage-station-passes">`)
	defer c.b.WriteString(`</g>`)
	fmt.Fprintf(&c.b, `<path d="M 36 %.1f H %.1f" stroke="#d6dbda"/>`, y, c.width-36)
	y = c.wrapped("coverage-station-title", 36, y+25, c.width-72, 17, 21, "start", c.words("站心过境", "Station passes")+" ("+c.loc.String()+")") + 12
	if len(passes) == 0 {
		c.text("coverage-pass-empty", 36, y, 13, "start", c.words("本时段无过境", "No passes in this interval"))
		return y
	}
	for index, pass := range passes {
		y = c.coveragePassDetails(pass, index+1, y) + 20
	}
	return y - 20
}

func (c *chart) coveragePassDetails(pass CoveragePass, number int, y float64) float64 {
	fmt.Fprintf(&c.b, `<g data-role="coverage-pass" data-rise="%s" data-set="%s">`, pass.Rise.Time.UTC().Format(time.RFC3339Nano), pass.Set.Time.UTC().Format(time.RFC3339Nano))
	defer c.b.WriteString(`</g>`)
	status := c.words("无目视时段", "Not visually observable")
	switch pass.Kind {
	case satellite.PassDaylight:
		status = c.words("白昼", "Daylight")
	case satellite.PassVisible:
		status = c.words("有目视时段", "Visually observable")
	}
	c.text("panel-label", 36, y, 13, "start", fmt.Sprintf(c.words("第 %d 场 · %s · %s", "Pass %d · %s · %s"), number, pass.Rise.Time.In(c.loc).Format("2006-01-02"), status))
	fraction, _ := coveragePassTimePrecision(pass.Set.Time.Sub(pass.Rise.Time))
	timeFormat := "15:04:05" + fraction
	localRise, localSet := pass.Rise.Time.In(c.loc), pass.Set.Time.In(c.loc)
	if localRise.Format("2006-01-02") != localSet.Format("2006-01-02") {
		timeFormat = "01-02 " + timeFormat
	}
	_, riseOffset := localRise.Zone()
	_, setOffset := localSet.Zone()
	if riseOffset != setOffset {
		timeFormat += " -07:00"
	}
	y += 23
	width := c.width - 72
	columns := [...]float64{0, .18, .43, .57, .71, .88}
	row := func(role string, values [6]string) {
		for index, value := range values {
			c.text(role, 36+columns[index]*width, y, 12, "start", value)
		}
		y += 18
	}
	row("panel-label", [6]string{c.words("阶段", "Stage"), c.words("时间", "Time"), c.words("方位", "Azimuth"), c.words("仰角", "Elevation"), c.words("距离", "Range"), c.words("星等", "Mag.")})
	rise, set := c.words("升起", "Rise"), c.words("落下", "Set")
	if pass.StartsAbove {
		rise = c.words("窗口起点", "Window start")
	}
	if pass.EndsAbove {
		set = c.words("窗口终点", "Window end")
	}
	stages := [...]string{rise, c.words("最高", "Highest"), set}
	for index, look := range [...]satellite.Look{pass.Rise, pass.Culmination, pass.Set} {
		stage := stages[index]
		magnitude := "—"
		if validNumber(look.Magnitude) {
			magnitude = fmt.Sprintf("%.1f", look.Magnitude)
		}
		row("coverage-pass-stage", [6]string{stage, look.Time.In(c.loc).Format(timeFormat), fmt.Sprintf("%.1f°", look.Azimuth), coverageElevationLabel(look.Elevation), fmt.Sprintf("%.0f km", look.RangeKM), magnitude})
	}
	if !pass.VisibleStart.IsZero() {
		y = c.wrapped("coverage-pass-visibility", 36, y+4, width, 12, 17, "start", c.words("目视首末 ", "First / last visible ")+pass.VisibleStart.In(c.loc).Format(timeFormat)+" / "+pass.VisibleEnd.In(c.loc).Format(timeFormat))
	}
	if !pass.BrightestTime.IsZero() {
		y = c.wrapped("coverage-pass-brightest", 36, y, width, 12, 17, "start", fmt.Sprintf(c.words("目视最亮 %s · %.1f 等", "Brightest %s · %.1f mag"), pass.BrightestTime.In(c.loc).Format(timeFormat), pass.BrightestMagnitude))
	}
	y = c.coveragePassChart(pass, timeFormat, y)
	return c.coveragePassPolarChart(pass, stages, timeFormat, y)
}

func (c *chart) coveragePassChart(pass CoveragePass, timeFormat string, y float64) float64 {
	if len(pass.Samples) == 0 {
		return y
	}
	elevationPass := pass
	elevationPass.Samples = coveragePassLooks(pass)
	magnitudeLow, magnitudeHigh := math.Inf(1), math.Inf(-1)
	elevationLow, elevationHigh := math.Inf(1), math.Inf(-1)
	for _, sample := range pass.Samples {
		if validNumber(sample.Magnitude) {
			magnitudeLow, magnitudeHigh = math.Min(magnitudeLow, sample.Magnitude), math.Max(magnitudeHigh, sample.Magnitude)
		}
	}
	for _, sample := range elevationPass.Samples {
		if validNumber(sample.Elevation) {
			elevationLow, elevationHigh = math.Min(elevationLow, sample.Elevation), math.Max(elevationHigh, sample.Elevation)
		}
	}
	if magnitudeLow > magnitudeHigh && elevationLow > elevationHigh {
		return y
	}
	if magnitudeLow == magnitudeHigh {
		magnitudeLow, magnitudeHigh = magnitudeLow-.5, magnitudeHigh+.5
	}
	if elevationLow == elevationHigh {
		elevationLow, elevationHigh = elevationLow-.5, elevationHigh+.5
	}
	x, top, right, bottom := 94., y+30, c.width-94, y+126
	const magnitudeColor = "#a66f18"
	const elevationColor = "#28718a"
	if magnitudeLow <= magnitudeHigh {
		coveragePassText(c, "coverage-magnitude-title", 36, y+7, "start", magnitudeColor, c.words("估算星等", "Estimated magnitude"))
		coveragePassText(c, "coverage-magnitude-tick", x-8, top+4, "end", magnitudeColor, fmt.Sprintf("%.3g", magnitudeLow))
		coveragePassText(c, "coverage-magnitude-tick", x-8, bottom, "end", magnitudeColor, fmt.Sprintf("%.3g", magnitudeHigh))
	}
	if elevationLow <= elevationHigh {
		coveragePassText(c, "coverage-elevation-title", c.width-36, y+7, "end", elevationColor, c.words("仰角 (°)", "Elevation (°)"))
		coveragePassText(c, "coverage-elevation-tick", right+8, top+4, "start", elevationColor, coverageElevationLabel(elevationHigh))
		coveragePassText(c, "coverage-elevation-tick", right+8, bottom, "start", elevationColor, coverageElevationLabel(elevationLow))
	}
	fmt.Fprintf(&c.b, `<path data-role="coverage-pass-axes" d="M %.1f %.1f V %.1f H %.1f V %.1f M %.1f %.1f H %.1f" fill="none" stroke="#aeb8b6"/>`, x, top, bottom, right, top, x, top, right)
	duration := pass.Set.Time.Sub(pass.Rise.Time)
	if duration <= 0 {
		duration = time.Second
	}
	tickStep := coveragePassTimeStep(duration, right-x, timeFormat)
	ticks := []time.Time{pass.Rise.Time}
	for at := pass.Rise.Time.Add(tickStep); at.Before(pass.Set.Time); at = at.Add(tickStep) {
		ticks = append(ticks, at)
	}
	if !pass.Set.Time.Equal(pass.Rise.Time) {
		endLabelWidth := svgchart.EstimatedTextWidth(pass.Set.Time.In(c.loc).Format(timeFormat), 11)
		for len(ticks) > 1 {
			last := ticks[len(ticks)-1]
			lastX := x + (right-x)*float64(last.Sub(pass.Rise.Time))/float64(duration)
			lastLabelWidth := svgchart.EstimatedTextWidth(last.In(c.loc).Format(timeFormat), 11)
			if right-lastX >= (lastLabelWidth+endLabelWidth)/2+8 {
				break
			}
			ticks = ticks[:len(ticks)-1]
		}
		ticks = append(ticks, pass.Set.Time)
	}
	for _, at := range ticks {
		px := x + (right-x)*float64(at.Sub(pass.Rise.Time))/float64(duration)
		fmt.Fprintf(&c.b, `<path d="M %.2f %.1f v 5" stroke="#aeb8b6"/>`, px, bottom)
		c.text("coverage-time-tick", px, bottom+20, 11, "middle", at.In(c.loc).Format(timeFormat))
	}
	if magnitudeLow <= magnitudeHigh {
		path := c.coveragePassCurvePath(pass, x, right, top, bottom, magnitudeLow, magnitudeHigh, false, func(look satellite.Look) float64 { return look.Magnitude })
		if path != "" {
			fmt.Fprintf(&c.b, `<path data-role="coverage-magnitude-curve" d="%s" fill="none" stroke="%s" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>`, path, magnitudeColor)
		}
	}
	if elevationLow <= elevationHigh {
		path := c.coveragePassCurvePath(elevationPass, x, right, top, bottom, elevationLow, elevationHigh, true, func(look satellite.Look) float64 { return look.Elevation })
		if path != "" {
			fmt.Fprintf(&c.b, `<path data-role="coverage-elevation-curve" d="%s" fill="none" stroke="%s" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>`, path, elevationColor)
		}
	}
	return bottom + 38
}

func (c *chart) coveragePassPolarChart(pass CoveragePass, stages [3]string, timeFormat string, y float64) float64 {
	const trackColor = "#bd4c3d"
	const gridColor = "#aeb8b6"
	const radius = 125.0
	centerX, centerY := c.width/2, y+radius+48
	c.b.WriteString(`<g data-role="coverage-pass-polar">`)
	c.text("coverage-pass-polar-title", centerX, y+17, 13, "middle", c.words("方位 / 仰角路径", "Azimuth / elevation path"))
	fmt.Fprintf(&c.b, `<path data-role="coverage-pass-polar-axes" d="M %.3f %.3f H %.3f M %.3f %.3f V %.3f" fill="none" stroke="#d6dbda"/>`, centerX-radius, centerY, centerX+radius, centerX, centerY-radius, centerY+radius)
	for _, elevation := range [...]float64{0, 30, 60} {
		ringRadius := radius * (90 - elevation) / 90
		fmt.Fprintf(&c.b, `<circle data-role="coverage-pass-polar-ring" data-elevation="%.1f" cx="%.3f" cy="%.3f" r="%.3f" fill="none" stroke="%s"/>`, elevation, centerX, centerY, ringRadius, gridColor)
		labelX := centerX + ringRadius*.72 + 5
		labelY := centerY - ringRadius*.72 - 3
		c.text("coverage-pass-polar-elevation", labelX, labelY, 10, "start", fmt.Sprintf("%.0f°", elevation))
	}
	fmt.Fprintf(&c.b, `<circle data-role="coverage-pass-polar-zenith" data-elevation="90.0" cx="%.3f" cy="%.3f" r="2.5" fill="%s"/>`, centerX, centerY, gridColor)
	for _, direction := range [...]struct {
		name                 string
		positionX, positionY float64
	}{
		{name: c.words("北", "N"), positionX: centerX, positionY: centerY - radius - 9},
		{name: c.words("东", "E"), positionX: centerX + radius + 17, positionY: centerY + 4},
		{name: c.words("南", "S"), positionX: centerX, positionY: centerY + radius + 19},
		{name: c.words("西", "W"), positionX: centerX - radius - 17, positionY: centerY + 4},
	} {
		c.text("coverage-pass-polar-direction", direction.positionX, direction.positionY, 11, "middle", direction.name)
	}

	path, arrow := coveragePassPolarPath(pass, centerX, centerY, radius)
	if path != "" {
		fmt.Fprintf(&c.b, `<path data-role="coverage-pass-polar-track" d="%s" fill="none" stroke="%s" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>`, path, trackColor)
	}
	if arrow != "" {
		fmt.Fprintf(&c.b, `<path data-role="coverage-pass-polar-arrow" d="%s" fill="%s" stroke="white" stroke-width="1" stroke-linejoin="round"/>`, arrow, trackColor)
	}
	for index, marker := range [...]struct {
		stage string
		look  satellite.Look
		color string
	}{
		{stage: "rise", look: pass.Rise, color: "#007653"},
		{stage: "culmination", look: pass.Culmination, color: trackColor},
		{stage: "set", look: pass.Set, color: "#692984"},
	} {
		pointX, pointY, ok := coveragePassPolarPosition(centerX, centerY, radius, marker.look.Azimuth, marker.look.Elevation)
		if !ok {
			continue
		}
		label := stages[index]
		title := label + " · " + marker.look.Time.In(c.loc).Format(timeFormat)
		fill, stroke := marker.color, "white"
		if index == 0 {
			fill, stroke = "white", marker.color
		}
		fmt.Fprintf(&c.b, `<circle data-role="coverage-pass-polar-marker" data-stage="%s" data-azimuth="%.6f" data-elevation="%.6f" cx="%.3f" cy="%.3f" r="4.5" fill="%s" stroke="%s" stroke-width="1.5"><title>%s</title></circle>`, marker.stage, marker.look.Azimuth, marker.look.Elevation, pointX, pointY, fill, stroke, html.EscapeString(title))
		legendX := centerX + float64(index-1)*160
		fmt.Fprintf(&c.b, `<circle data-role="coverage-pass-polar-legend-marker" cx="%.3f" cy="%.3f" r="3.5" fill="%s" stroke="%s" stroke-width="1.5"/>`, legendX-43, centerY+radius+43, fill, stroke)
		c.text("coverage-pass-polar-legend", legendX-34, centerY+radius+47, 11, "start", label)
	}
	c.b.WriteString(`</g>`)
	return centerY + radius + 62
}

func coveragePassLooks(pass CoveragePass) []satellite.Look {
	looks := make([]satellite.Look, 0, len(pass.Samples)+3)
	looks = append(looks, pass.Rise, pass.Culmination, pass.Set)
	looks = append(looks, pass.Samples...)
	sort.SliceStable(looks, func(leftIndex, rightIndex int) bool { return looks[leftIndex].Time.Before(looks[rightIndex].Time) })
	unique := looks[:0]
	for _, look := range looks {
		if len(unique) == 0 || !look.Time.Equal(unique[len(unique)-1].Time) {
			unique = append(unique, look)
		}
	}
	return unique
}

func coveragePassPolarPath(pass CoveragePass, centerX, centerY, radius float64) (string, string) {
	var path strings.Builder
	var run [][2]float64
	var length float64
	longest, arrow := 0.0, ""
	finishRun := func() {
		if length > 0 && length >= longest {
			offset := 0.0
			if length < 24 {
				offset = 13
			}
			longest, arrow = length, coveragePassPolarArrow(run, length*.55, offset)
		}
		run, length = run[:0], 0
	}
	command := "M"
	for _, look := range coveragePassLooks(pass) {
		pointX, pointY, ok := coveragePassPolarPosition(centerX, centerY, radius, look.Azimuth, look.Elevation)
		if !ok {
			finishRun()
			command = "M"
			continue
		}
		fmt.Fprintf(&path, " %s %.3f %.3f", command, pointX, pointY)
		command = "L"
		if len(run) > 0 {
			previous := run[len(run)-1]
			length += math.Hypot(pointX-previous[0], pointY-previous[1])
		}
		point := [2]float64{pointX, pointY}
		run = append(run, point)
		if look.Time.Equal(pass.Culmination.Time) {
			finishRun()
			run = append(run, point)
		}
	}
	finishRun()
	return path.String(), arrow
}

func coveragePassPolarArrow(run [][2]float64, distance, offset float64) string {
	for index := 1; index < len(run); index++ {
		start, end := run[index-1], run[index]
		deltaX, deltaY := end[0]-start[0], end[1]-start[1]
		segment := math.Hypot(deltaX, deltaY)
		if segment == 0 || distance > segment {
			distance -= segment
			continue
		}
		unitX, unitY := deltaX/segment, deltaY/segment
		tipX := start[0] + distance*unitX - offset*unitY
		tipY := start[1] + distance*unitY + offset*unitX
		baseX, baseY := tipX-10*unitX, tipY-10*unitY
		return fmt.Sprintf("M%.3f %.3f L%.3f %.3f L%.3f %.3f Z", tipX, tipY, baseX-4.5*unitY, baseY+4.5*unitX, baseX+4.5*unitY, baseY-4.5*unitX)
	}
	return ""
}

func coveragePassPolarPosition(centerX, centerY, radius, azimuth, elevation float64) (float64, float64, bool) {
	if !validNumber(azimuth) || !validNumber(elevation) || elevation < -0.05 || elevation > 90 {
		return 0, 0, false
	}
	angle := math.Mod(azimuth, 360) * math.Pi / 180
	trackRadius := radius * (90 - math.Max(0, elevation)) / 90
	return centerX + trackRadius*math.Sin(angle), centerY - trackRadius*math.Cos(angle), true
}

func coverageElevationLabel(value float64) string {
	if math.Abs(value) < .05 {
		return "0°"
	}
	return fmt.Sprintf("%.1f°", value)
}

func coveragePassTimePrecision(duration time.Duration) (string, time.Duration) {
	switch {
	case duration <= 0 || duration >= time.Second:
		return "", time.Second
	case duration >= time.Millisecond:
		return ".000", time.Millisecond
	case duration >= time.Microsecond:
		return ".000000", time.Microsecond
	default:
		return ".000000000", time.Nanosecond
	}
}

func coveragePassTimeStep(duration time.Duration, width float64, timeFormat string) time.Duration {
	labelWidth := svgchart.EstimatedTextWidth(time.Time{}.UTC().Format(timeFormat), 11)
	maxIntervals := int(width / (labelWidth + 12))
	if maxIntervals < 2 {
		maxIntervals = 2
	}
	_, unit := coveragePassTimePrecision(duration)
	for _, multiple := range [...]time.Duration{1, 2, 5, 10, 15, 30, 60, 120, 300, 600, 900, 1800, 3600, 7200, 10800, 21600, 43200} {
		step := multiple * unit
		if duration < step*time.Duration(maxIntervals+1) {
			return step
		}
	}
	return duration / time.Duration(maxIntervals)
}

func (c *chart) coveragePassCurvePath(pass CoveragePass, x, right, top, bottom, low, high float64, increasingUpward bool, value func(satellite.Look) float64) string {
	var curve strings.Builder
	penDown := false
	duration := math.Max(1, float64(pass.Set.Time.Sub(pass.Rise.Time)))
	for _, sample := range pass.Samples {
		value := value(sample)
		if !validNumber(value) {
			penDown = false
			continue
		}
		px := x + (right-x)*float64(sample.Time.Sub(pass.Rise.Time))/duration
		fraction := (value - low) / math.Max(1e-12, high-low)
		py := top + (bottom-top)*fraction
		if increasingUpward {
			py = bottom - (bottom-top)*fraction
		}
		if penDown {
			fmt.Fprintf(&curve, " L %.2f %.2f", px, py)
		} else {
			fmt.Fprintf(&curve, " M %.2f %.2f l 0 0", px, py)
			penDown = true
		}
	}
	return curve.String()
}

func coveragePassText(c *chart, role string, x, y float64, anchor, color, value string) {
	fmt.Fprintf(&c.b, `<text data-role="%s" x="%.4f" y="%.4f" font-family="Arial, sans-serif" font-size="12.00" font-weight="700" text-anchor="%s" fill="%s">%s</text>`, role, x, y, anchor, color, html.EscapeString(value))
}
