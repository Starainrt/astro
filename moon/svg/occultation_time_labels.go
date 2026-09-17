package svg

import (
	"fmt"
	"html"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro/moon"
)

func writeOccultationTimeMarkers(
	b *strings.Builder,
	points []moon.OccultationPathPoint,
	layout starOccultationSVGLayout,
	options StarOccultationSVGOptions,
	excluded []time.Time,
	greatest time.Time,
	placer *occultationLabelPlacer,
) {
	if options.TimeLabelStep <= 0 || len(points) < 2 {
		return
	}
	markers := occultationTimeMarkerPoints(points, options.TimeLabelStep, options.Location, excluded)
	frame := layout.mapFrame()
	projected := make([][2]float64, 0, len(markers))
	for _, marker := range markers {
		if marker.MoonAltitude < 0 {
			continue
		}
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
		text := marker.Time.In(options.Location).Format("15:04")
		// 掩甚附近的标签默认压到圆点下方，给掩甚事件标记让位；两个方向都被占用就丢弃这个时刻。
		baselines := []float64{y - 8, y + 15}
		if occultationTimesNear(marker.Time, greatest, occultationGreatestTimeLabelWindow(options.TimeLabelStep)) {
			baselines[0], baselines[1] = baselines[1], baselines[0]
		}
		labelY, placed := 0.0, false
		for _, baseline := range baselines {
			if baseline < frame.Y+10 || baseline > frame.Y+frame.Height-8 {
				continue
			}
			if placer.place(x, baseline, "middle", text, 9, 2) {
				labelY, placed = baseline, true
				break
			}
		}
		if !placed {
			continue
		}
		projected = append(projected, [2]float64{x, y})
		fmt.Fprintf(b, `<g class="occultation-time-marker"><circle cx="%.3f" cy="%.3f" r="2.3" fill="#087f8c" stroke="#ffffff" stroke-width="1"/><text x="%.3f" y="%.3f" fill="#075f69" stroke="#ffffff" stroke-width="3" paint-order="stroke" font-family="Arial, sans-serif" font-size="9" font-weight="700" text-anchor="middle">%s</text></g>`,
			x, y, x, labelY, html.EscapeString(text))
	}
}

func occultationGreatestTimeLabelWindow(step time.Duration) time.Duration {
	window := step / 3
	if window < 10*time.Minute {
		return 10 * time.Minute
	}
	return window
}

func occultationTimeMarkerPoints(
	points []moon.OccultationPathPoint,
	step time.Duration,
	location *time.Location,
	excluded []time.Time,
) []moon.OccultationPathPoint {
	if len(points) < 2 || step <= 0 {
		return nil
	}
	start := points[0].Time
	end := points[len(points)-1].Time
	current := firstOccultationTimeLabelAfter(start, step, location)
	window := step / 4
	if window > 5*time.Minute {
		window = 5 * time.Minute
	}
	if window < 30*time.Second {
		window = 30 * time.Second
	}
	result := make([]moon.OccultationPathPoint, 0)
	segment := 1
	for current.Before(end) {
		for segment < len(points) && points[segment].Time.Before(current) {
			segment++
		}
		if segment >= len(points) {
			break
		}
		if !occultationTimeNearAny(current, excluded, window) {
			a, next := points[segment-1], points[segment]
			span := next.Time.Sub(a.Time)
			if span > 0 {
				fraction := float64(current.Sub(a.Time)) / float64(span)
				result = append(result, interpolateOccultationTimeMarker(a, next, fraction, current))
			}
		}
		current = current.Add(step)
	}
	return result
}

func interpolateOccultationTimeMarker(
	a, b moon.OccultationPathPoint,
	fraction float64,
	value time.Time,
) moon.OccultationPathPoint {
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
	point := starOccultationInterpolatePathPoint(a, b, fraction, longitude)
	point.Time = value
	return point
}

func firstOccultationTimeLabelAfter(value time.Time, step time.Duration, location *time.Location) time.Time {
	local := value.In(location)
	dayStart := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location)
	elapsed := local.Sub(dayStart)
	return dayStart.Add((elapsed/step + 1) * step)
}

func occultationTimeNearAny(value time.Time, excluded []time.Time, window time.Duration) bool {
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

func occultationTimesNear(a, b time.Time, window time.Duration) bool {
	delta := a.Sub(b)
	if delta < 0 {
		delta = -delta
	}
	return delta <= window
}
