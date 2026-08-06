package basic

import (
	"math"
	"time"
)

const (
	planetOccultationFootprintBoundaryPoints = 180
	planetOccultationHorizonPoints           = 360
	planetOccultationFootprintMaxSamples     = 360
)

type planetOccultationFootprintSample struct {
	point OccultationPathPoint
	ok    bool
}

func planetOccultationFootprints(
	startTT, endTT, greatestTT float64,
	frameAt occultationPathFrameFunc,
	options OccultationPathOptions,
	location *time.Location,
) []PlanetOccultationFootprint {
	stepDays := float64(options.Step) / float64(24*time.Hour)
	times := occultationPathSampleTimesWithLimit(
		startTT, endTT, greatestTT, stepDays, planetOccultationFootprintMaxSamples,
	)
	footprints := make([]PlanetOccultationFootprint, 0, len(times))
	for _, tt := range times {
		footprint, ok := planetOccultationFootprintAt(tt, frameAt, location)
		if ok {
			footprints = append(footprints, footprint)
		}
	}
	return footprints
}

func planetOccultationFootprintAt(
	tt float64,
	frameAt occultationPathFrameFunc,
	location *time.Location,
) (PlanetOccultationFootprint, bool) {
	frame, ok := frameAt(tt)
	if !ok {
		return PlanetOccultationFootprint{}, false
	}
	samples := make([]planetOccultationFootprintSample, planetOccultationFootprintBoundaryPoints)
	for index := range samples {
		theta := 2 * math.Pi * float64(index) / float64(len(samples))
		vector, _, valid := occultationPathBoundaryVector(frame, theta)
		if valid {
			samples[index] = planetOccultationFootprintSample{
				point: occultationPathPointFromVector(tt, vector, 0, location),
				ok:    true,
			}
		}
	}

	segments, closed := planetOccultationFootprintSegments(samples)
	polygons := make([][]OccultationPathPoint, 0, len(segments))
	for _, segment := range segments {
		if len(segment) < 2 {
			continue
		}
		polygon := append([]OccultationPathPoint(nil), segment...)
		if closed {
			polygon = append(polygon, polygon[0])
		} else {
			polygon = append(polygon, planetOccultationHorizonArc(tt, frame, segment[len(segment)-1], segment[0], location)...)
		}
		if len(polygon) >= 4 {
			polygons = append(polygons, polygon)
		}
	}
	if len(polygons) == 0 {
		return PlanetOccultationFootprint{}, false
	}
	return PlanetOccultationFootprint{
		Time:     occultationTTToLocation(tt, location),
		Polygons: polygons,
	}, true
}

func planetOccultationFootprintSegments(
	samples []planetOccultationFootprintSample,
) ([][]OccultationPathPoint, bool) {
	segments := make([][]OccultationPathPoint, 0, 2)
	current := make([]OccultationPathPoint, 0, len(samples))
	allValid := len(samples) > 0
	for _, sample := range samples {
		if !sample.ok {
			allValid = false
			if len(current) > 0 {
				segments = append(segments, current)
				current = nil
			}
			continue
		}
		current = append(current, sample.point)
	}
	if len(current) > 0 {
		segments = append(segments, current)
	}
	if len(segments) > 1 && samples[0].ok && samples[len(samples)-1].ok {
		first := segments[0]
		last := segments[len(segments)-1]
		merged := make([]OccultationPathPoint, 0, len(last)+len(first))
		merged = append(merged, last...)
		merged = append(merged, first...)
		segments[0] = merged
		segments = segments[:len(segments)-1]
	}
	return segments, allValid && len(segments) == 1
}

func planetOccultationHorizonArc(
	tt float64,
	frame occultationPathFrame,
	from, to OccultationPathPoint,
	location *time.Location,
) []OccultationPathPoint {
	sublunarLongitude, sublunarLatitude := occultationPathGeodetic(tt, frame.moon)
	circle := planetOccultationSphericalCircle(
		occultationTTToLocation(tt, location), sublunarLongitude, sublunarLatitude,
		90, planetOccultationHorizonPoints,
	)
	fromIndex := planetOccultationNearestPointIndex(circle, from)
	toIndex := planetOccultationNearestPointIndex(circle, to)
	forwardSteps := (toIndex - fromIndex + len(circle)) % len(circle)
	backwardSteps := (fromIndex - toIndex + len(circle)) % len(circle)
	direction := 1
	steps := forwardSteps
	if backwardSteps < forwardSteps {
		direction = -1
		steps = backwardSteps
	}
	arc := make([]OccultationPathPoint, 0, steps+1)
	for step := 1; step < steps; step++ {
		index := (fromIndex + direction*step) % len(circle)
		if index < 0 {
			index += len(circle)
		}
		arc = append(arc, circle[index])
	}
	return append(arc, to)
}

func planetOccultationSphericalCircle(
	value time.Time,
	centerLongitude, centerLatitude, radius float64,
	count int,
) []OccultationPathPoint {
	centerLongitude *= math.Pi / 180
	centerLatitude *= math.Pi / 180
	radius *= math.Pi / 180
	points := make([]OccultationPathPoint, count)
	for index := range points {
		bearing := 2 * math.Pi * float64(index) / float64(count)
		latitude := math.Asin(
			math.Sin(centerLatitude)*math.Cos(radius) +
				math.Cos(centerLatitude)*math.Sin(radius)*math.Cos(bearing),
		)
		longitude := centerLongitude + math.Atan2(
			math.Sin(bearing)*math.Sin(radius)*math.Cos(centerLatitude),
			math.Cos(radius)-math.Sin(centerLatitude)*math.Sin(latitude),
		)
		points[index] = OccultationPathPoint{
			Time:      value,
			Longitude: normalizeLongitude(longitude * 180 / math.Pi),
			Latitude:  latitude * 180 / math.Pi,
		}
	}
	return points
}

func planetOccultationNearestPointIndex(points []OccultationPathPoint, target OccultationPathPoint) int {
	nearest := 0
	distance := math.Inf(1)
	for index, point := range points {
		candidate := occultationPathDistanceKM(point, target)
		if candidate < distance {
			nearest = index
			distance = candidate
		}
	}
	return nearest
}
