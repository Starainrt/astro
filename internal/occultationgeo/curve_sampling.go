package occultationgeo

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

const defaultRiseSetProjectedSpacingKM = 35.0

// DensifyRiseSetCurves 按投影距离插入大地线样本，且不改变端点和分支拓扑。
// DensifyRiseSetCurves inserts geodesic samples using the projected distance
// consumed by map renderers. Rise/set solvers intentionally use a coarser time
// step for performance; using those raw vertices as SVG/GeoJSON linework makes
// high-latitude curves visibly angular even when the physical path is smooth.
// Endpoints, phase, direction, and segment topology are preserved exactly.
func DensifyRiseSetCurves(
	curves []basic.OccultationRiseSetCurve,
	maximumProjectedSpacingKM float64,
) []basic.OccultationRiseSetCurve {
	if len(curves) == 0 || maximumProjectedSpacingKM <= 0 {
		return curves
	}
	result := make([]basic.OccultationRiseSetCurve, len(curves))
	for curveIndex, curve := range curves {
		result[curveIndex] = basic.OccultationRiseSetCurve{
			Phase: curve.Phase, Direction: curve.Direction,
			Segments: make([][]basic.OccultationPathPoint, len(curve.Segments)),
		}
		for segmentIndex, segment := range curve.Segments {
			result[curveIndex].Segments[segmentIndex] = densifyRiseSetSegment(segment, maximumProjectedSpacingKM)
		}
	}
	return result
}

// DensifyOccultationPathPoints 将相同的投影感知间距应用于单条月掩路径。
// DensifyOccultationPathPoints applies the same projection-aware spacing to a
// standalone geographic line such as a horizon connector. It keeps both
// endpoints and does not attach phase metadata to the result.
func DensifyOccultationPathPoints(
	points []basic.OccultationPathPoint,
	maximumProjectedSpacingKM float64,
) []basic.OccultationPathPoint {
	return densifyRiseSetSegment(points, maximumProjectedSpacingKM)
}

func densifyRiseSetSegment(
	segment []basic.OccultationPathPoint,
	maximumProjectedSpacingKM float64,
) []basic.OccultationPathPoint {
	if len(segment) < 2 {
		return append([]basic.OccultationPathPoint(nil), segment...)
	}
	result := make([]basic.OccultationPathPoint, 0, len(segment)*2)
	// The interpolation is geodesic on the sphere, while the spacing check is
	// made in Web Mercator.  A chord that is exactly at the projected limit can
	// grow slightly after interpolation because Mercator is nonlinear in
	// latitude.  Keep a small margin so the emitted linework stays below the
	// requested limit after projection.
	effectiveSpacingKM := maximumProjectedSpacingKM * 0.75
	for index, point := range segment {
		result = append(result, point)
		if index+1 >= len(segment) {
			continue
		}
		next := segment[index+1]
		steps := int(math.Ceil(projectedGeoPointDistanceKM(point, next) / effectiveSpacingKM))
		if steps < 2 {
			continue
		}
		for step := 1; step < steps; step++ {
			fraction := float64(step) / float64(steps)
			middle := interpolateOccultationGeoPoint(
				geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude},
				geodata.GeoPoint{Longitude: next.Longitude, Latitude: next.Latitude},
				fraction,
			)
			result = append(result, basic.OccultationPathPoint{
				Time:         point.Time.Add(time.Duration(float64(next.Time.Sub(point.Time)) * fraction)),
				Longitude:    middle.Longitude,
				Latitude:     middle.Latitude,
				MoonAltitude: point.MoonAltitude + (next.MoonAltitude-point.MoonAltitude)*fraction,
				WidthKM:      point.WidthKM + (next.WidthKM-point.WidthKM)*fraction,
			})
		}
	}
	return result
}

func projectedGeoPointDistanceKM(first, second basic.OccultationPathPoint) float64 {
	const maxLatitude = 85.05112878
	const radiusKM = 6378.1366
	clampLatitude := func(value float64) float64 {
		return math.Max(-maxLatitude, math.Min(maxLatitude, value))
	}
	longitude := math.Remainder(second.Longitude-first.Longitude, 360) * math.Pi / 180
	firstLatitude := clampLatitude(first.Latitude) * math.Pi / 180
	secondLatitude := clampLatitude(second.Latitude) * math.Pi / 180
	firstY := math.Log(math.Tan(math.Pi/4 + firstLatitude/2))
	secondY := math.Log(math.Tan(math.Pi/4 + secondLatitude/2))
	return radiusKM * math.Hypot(longitude, secondY-firstY)
}
