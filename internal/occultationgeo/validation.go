package occultationgeo

import (
	"fmt"
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

// ValidateRiseSetCurves 检查 SVG 和 GeoJSON 渲染器消费的公开地平曲线数据。
// ValidateRiseSetCurves checks the public horizon-curve data consumed by SVG and GeoJSON renderers.
func ValidateRiseSetCurves(curves []basic.OccultationRiseSetCurve, start, end time.Time) error {
	seen := make(map[[2]string]bool, len(curves))
	for curveIndex, curve := range curves {
		if !validRiseSetPhase(curve.Phase) {
			return fmt.Errorf("curve[%d] has unsupported phase %q", curveIndex, curve.Phase)
		}
		if !validRiseSetDirection(curve.Direction) {
			return fmt.Errorf("curve[%d] has unsupported direction %q", curveIndex, curve.Direction)
		}
		key := [2]string{string(curve.Phase), string(curve.Direction)}
		if seen[key] {
			return fmt.Errorf("curve[%d] duplicates phase %q and direction %q", curveIndex, curve.Phase, curve.Direction)
		}
		seen[key] = true
		if len(curve.Segments) == 0 {
			return fmt.Errorf("curve[%d] must contain at least one segment", curveIndex)
		}
		for segmentIndex, segment := range curve.Segments {
			if len(segment) < 2 {
				return fmt.Errorf("curve[%d].segment[%d] must contain at least two points", curveIndex, segmentIndex)
			}
			for pointIndex, point := range segment {
				if err := validatePathPoint(point, start, end); err != nil {
					return fmt.Errorf("curve[%d].segment[%d].point[%d]: %w", curveIndex, segmentIndex, pointIndex, err)
				}
				if pointIndex > 0 && !point.Time.After(segment[pointIndex-1].Time) {
					return fmt.Errorf("curve[%d].segment[%d] times must be strictly increasing", curveIndex, segmentIndex)
				}
				if pointIndex > 0 && riseSetCurveBranchChanged(segment[pointIndex-1], point) {
					return fmt.Errorf("curve[%d].segment[%d] contains a discontinuous branch jump between points %d and %d", curveIndex, segmentIndex, pointIndex-1, pointIndex)
				}
			}
		}
	}
	return nil
}

func riseSetCurveBranchChanged(first, second basic.OccultationPathPoint) bool {
	distance := DistanceKM(first, second)
	if !finite(distance) {
		return true
	}
	if distance <= 500 {
		return false
	}
	duration := math.Abs(second.Time.Sub(first.Time).Seconds())
	return duration <= 0 || duration < 1 && distance/duration > 10
}

func validRiseSetPhase(phase basic.RiseSetPhase) bool {
	switch phase {
	case basic.RiseSetPhaseStart, basic.RiseSetPhaseGreatest, basic.RiseSetPhaseEnd:
		return true
	default:
		return false
	}
}

func validRiseSetDirection(direction basic.RiseSetDirection) bool {
	switch direction {
	case basic.RiseSetDirectionRise, basic.RiseSetDirectionSet:
		return true
	default:
		return false
	}
}

// ValidateFootprints 检查 SVG 和 GeoJSON 渲染器消费的瞬时月掩区域。
// ValidateFootprints checks instantaneous occultation regions consumed by SVG and GeoJSON renderers.
func ValidateFootprints(footprints []basic.OccultationFootprint, start, end time.Time) error {
	previous := time.Time{}
	for footprintIndex, footprint := range footprints {
		if footprint.Time.IsZero() {
			return fmt.Errorf("footprint[%d] time is required", footprintIndex)
		}
		if footprint.Time.Before(start) || footprint.Time.After(end) {
			return fmt.Errorf("footprint[%d] time must be inside its contact interval", footprintIndex)
		}
		if !previous.IsZero() && !footprint.Time.After(previous) {
			return fmt.Errorf("footprint times must be strictly increasing")
		}
		if len(footprint.Polygons) == 0 {
			return fmt.Errorf("footprint[%d] must contain a polygon", footprintIndex)
		}
		for polygonIndex, polygon := range footprint.Polygons {
			if len(polygon) < 3 {
				return fmt.Errorf("footprint[%d].polygon[%d] must contain at least three points", footprintIndex, polygonIndex)
			}
			for pointIndex, point := range polygon {
				if err := validatePathPoint(point, start, end); err != nil {
					return fmt.Errorf("footprint[%d].polygon[%d].point[%d]: %w", footprintIndex, polygonIndex, pointIndex, err)
				}
				if !point.Time.Equal(footprint.Time) {
					return fmt.Errorf("footprint[%d].polygon[%d].point[%d] time must match its footprint", footprintIndex, polygonIndex, pointIndex)
				}
			}
		}
		if footprint.Closed && len(footprint.Boundaries) != 1 {
			return fmt.Errorf("footprint[%d] closed boundary must contain exactly one segment", footprintIndex)
		}
		for boundaryIndex, boundary := range footprint.Boundaries {
			if len(boundary) < 2 {
				return fmt.Errorf("footprint[%d].boundary[%d] must contain at least two points", footprintIndex, boundaryIndex)
			}
			for pointIndex, point := range boundary {
				if err := validatePathPoint(point, start, end); err != nil {
					return fmt.Errorf("footprint[%d].boundary[%d].point[%d]: %w", footprintIndex, boundaryIndex, pointIndex, err)
				}
				if !point.Time.Equal(footprint.Time) {
					return fmt.Errorf("footprint[%d].boundary[%d].point[%d] time must match its footprint", footprintIndex, boundaryIndex, pointIndex)
				}
			}
		}
		previous = footprint.Time
	}
	return nil
}

func validatePathPoint(point basic.OccultationPathPoint, start, end time.Time) error {
	if point.Time.IsZero() {
		return fmt.Errorf("time is required")
	}
	if point.Time.Before(start) || point.Time.After(end) {
		return fmt.Errorf("time must be inside the occultation interval")
	}
	if !finite(point.Longitude) || point.Longitude < -180 || point.Longitude > 180 {
		return fmt.Errorf("longitude must be finite and within [-180, 180]")
	}
	if !finite(point.Latitude) || point.Latitude < -90 || point.Latitude > 90 {
		return fmt.Errorf("latitude must be finite and within [-90, 90]")
	}
	if !finite(point.MoonAltitude) || point.MoonAltitude < -90 || point.MoonAltitude > 90 {
		return fmt.Errorf("Moon altitude must be finite and within [-90, 90]")
	}
	if !finite(point.WidthKM) || point.WidthKM < 0 {
		return fmt.Errorf("width must be finite and non-negative")
	}
	return nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
