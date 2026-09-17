package occultationgeo

import (
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

const (
	horizonConnectorMatchDistanceKM = 180.0
	horizonConnectorMinimumLengthKM = 35.0
	// Footprint timelines are sampled at five-minute cadence while rise/set
	// endpoints are solved independently. Allow the nearest sparse sample to
	// be about one minute away without accepting a different event branch.
	horizonConnectorMatchWindow   = 90 * time.Second
	horizonConnectorDensifyEdgeKM = 25.0
	// 同相位端点在 5 km 以内属于同一极区折叠：显示曲线已由 StitchedRiseSetCurveSegments
	// 缝合，再输出连接段会在折点处多出一条悬空线，因此这类配对只用于否决该边界的连接段。
	horizonConnectorSamePhaseEndpointDistanceKM = 5.0
	// A same-phase opening/closing fold can expose two branches of one sampled
	// phase curve. It is eligible for a connector only while the footprint
	// endpoints remain within this bound.
	horizonConnectorSamePhaseBranchDistanceKM = 1200.0
	horizonConnectorBoundaryShiftToleranceKM  = 25.0
	// A contact-phase pinch can give two different phase curves the same
	// geographic horizon endpoint. There is no missing horizon interval to
	// draw in that case; retaining the footprint's closure produces a small
	// out-and-back loop instead of a boundary segment.
	horizonConnectorDegenerateEndpointDistanceKM = 0.01
	// A start and end phase can meet the same rise/set horizon branch at
	// slightly different instants. The missing boundary is then the short H=0
	// trajectory between those exact endpoints, not an instantaneous footprint
	// closure.
	horizonConnectorTemporalWindow     = 2 * time.Minute
	horizonConnectorTemporalDistanceKM = 250.0
)

// HorizonConnector 是在可见接触弧两端之间连接的同时刻月升或月落边界段。
// HorizonConnector is a same-instant moonrise/moonset boundary segment that
// connects adjacent local-phase curves at the visible footprint opening or
// closing. It is not itself one of the start/greatest/end phase curves.
type HorizonConnector struct {
	Direction basic.RiseSetDirection
	Points    []basic.OccultationPathPoint
}

type horizonConnectorEndpoint struct {
	point        basic.OccultationPathPoint
	phase        basic.RiseSetPhase
	direction    basic.RiseSetDirection
	curveIndex   int
	segmentIndex int
	atStart      bool
	index        int
}

type horizonConnectorMatch struct {
	endpoint horizonConnectorEndpoint
	metric   float64
}

type horizonConnectorCandidate struct {
	connector HorizonConnector
	metric    float64
}

// HorizonConnectorSegments 仅提取端点同时与公开升落阶段端点重合的地平闭合段。
// HorizonConnectorSegments extracts only the horizon closures whose open
// contact-limb endpoints tie two sampled rise/set phase endpoints together.
// Ordinary per-instant horizon closures are intentionally ignored so compact
// bands do not render as many parallel horizon stripes. Supplied north/south
// limits filter out closures that belong to a branch outside the exported
// limit envelope.
func HorizonConnectorSegments(
	footprints []basic.OccultationFootprint,
	curves []basic.OccultationRiseSetCurve,
	limits ...[]basic.OccultationPathPoint,
) []HorizonConnector {
	return horizonConnectorSegments(footprints, curves, false, limits...)
}

// StarHorizonConnectorSegments 额外闭合点光源起止阶段在同一月球地平支路上相遇时的短时间缺口。
// StarHorizonConnectorSegments additionally closes a short temporal gap where
// a point-source start and end phase meet the same lunar-horizon branch at
// nearby instants. Finite-disk planet contacts retain same-instant connectors.
func StarHorizonConnectorSegments(
	footprints []basic.OccultationFootprint,
	curves []basic.OccultationRiseSetCurve,
	limits ...[]basic.OccultationPathPoint,
) []HorizonConnector {
	return horizonConnectorSegments(footprints, curves, true, limits...)
}
func horizonConnectorSegments(
	footprints []basic.OccultationFootprint,
	curves []basic.OccultationRiseSetCurve,
	includeTemporal bool,
	limits ...[]basic.OccultationPathPoint,
) []HorizonConnector {
	endpoints := horizonConnectorEndpoints(curves)
	if len(endpoints) < 2 || len(footprints) == 0 {
		return nil
	}
	best := make(map[string]horizonConnectorCandidate)
	for _, footprint := range footprints {
		if footprint.Closed {
			continue
		}
		for boundaryIndex, boundary := range footprint.Boundaries {
			if len(boundary) < 2 {
				continue
			}
			firstPoint, lastPoint := boundary[0], boundary[len(boundary)-1]
			if DistanceKM(firstPoint, lastPoint) < horizonConnectorMinimumLengthKM {
				continue
			}
			firstMatches := horizonConnectorEndpointMatches(firstPoint, endpoints)
			lastMatches := horizonConnectorEndpointMatches(lastPoint, endpoints)
			first, last, matched := selectHorizonConnectorMatch(firstMatches, lastMatches)
			if !matched {
				continue
			}
			points := horizonClosurePoints(footprint, boundaryIndex)
			if len(points) < 2 {
				continue
			}
			samePhaseFold := first.endpoint.phase == last.endpoint.phase &&
				DistanceKM(first.endpoint.point, last.endpoint.point) <= horizonConnectorSamePhaseEndpointDistanceKM
			if samePhaseFold {
				// Same-phase fold branches are stitched into one display curve by
				// StitchedRiseSetCurveSegments. Emitting the sampled horizon suffix
				// here would create a second dangling purple stroke at the fold.
				continue
			}
			// The sparse footprint time can sit a fraction of a second between
			// the independently refined phase-curve endpoints. Use those
			// authoritative endpoints so the rendered lines meet exactly while
			// retaining the already-computed horizon arc between them.
			points[0] = last.endpoint.point
			points[len(points)-1] = first.endpoint.point
			points = trimHorizonConnectorSamples(points)
			points = smoothHorizonConnectorJunctions(points)
			if horizonConnectorIsDegenerate(points) {
				continue
			}
			if len(points) <= 8 {
				points = densifyHorizonConnectorPoints(points, horizonConnectorDensifyEdgeKM)
			}
			targetTime := first.endpoint.point.Time
			if last.endpoint.point.Time.Before(targetTime) {
				targetTime = last.endpoint.point.Time
			}
			firstIndex, lastIndex := first.endpoint.index, last.endpoint.index
			if firstIndex > lastIndex {
				firstIndex, lastIndex = lastIndex, firstIndex
			}
			key := string(first.endpoint.direction) + "/" +
				targetTime.UTC().Round(time.Second).Format(time.RFC3339) + "/" +
				strconv.Itoa(firstIndex) + "/" + strconv.Itoa(lastIndex)
			metric := first.metric + last.metric + absDuration(first.endpoint.point.Time.Sub(last.endpoint.point.Time)).Seconds()
			if current, ok := best[key]; !ok || metric < current.metric {
				best[key] = horizonConnectorCandidate{
					connector: HorizonConnector{Direction: first.endpoint.direction, Points: points},
					metric:    metric,
				}
			}
		}
	}
	if includeTemporal {
		for _, candidate := range temporalHorizonConnectorCandidates(footprints, endpoints) {
			if current, ok := best[candidate.key]; !ok || candidate.value.metric < current.metric {
				best[candidate.key] = candidate.value
			}
		}
	}
	if len(best) == 0 {
		return nil
	}
	keys := make([]string, 0, len(best))
	for key := range best {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]HorizonConnector, 0, len(keys))
	for _, key := range keys {
		result = append(result, best[key].connector)
	}
	return horizonConnectorsWithinLimits(result, limits)
}

// horizonConnectorsWithinLimits 丢弃整段都远离导出南北限的连接段（换支伪影）。
func horizonConnectorsWithinLimits(
	connectors []HorizonConnector,
	limits [][]basic.OccultationPathPoint,
) []HorizonConnector {
	if len(connectors) == 0 || len(limits) == 0 {
		return connectors
	}
	filtered := make([]HorizonConnector, 0, len(connectors))
	for _, connector := range connectors {
		if horizonConnectorTouchesLimits(connector.Points, limits) {
			filtered = append(filtered, connector)
		}
	}
	return filtered
}

func horizonConnectorTouchesLimits(
	points []basic.OccultationPathPoint,
	limits [][]basic.OccultationPathPoint,
) bool {
	hasLimit := false
	for _, limit := range limits {
		if len(limit) < 2 {
			continue
		}
		hasLimit = true
		for index := 1; index < len(limit); index++ {
			start := geodata.GeoPoint{Longitude: limit[index-1].Longitude, Latitude: limit[index-1].Latitude}
			end := geodata.GeoPoint{Longitude: limit[index].Longitude, Latitude: limit[index].Latitude}
			for _, point := range points {
				if geoPointSegmentDistanceKM(
					geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude},
					start, end,
				) <= horizonConnectorMatchDistanceKM {
					return true
				}
			}
		}
	}
	// 限线为空时没有可用的过滤依据，保留原结果。
	return !hasLimit
}

// smoothHorizonConnectorJunctions removes a short numerical corner where an
// exact phase endpoint replaces the nearest sparse horizon sample. The source
// arc is still the lunar-altitude-zero boundary; deleting only the offending
// interior sample preserves both exact endpoints and time order, while the
// caller's densification restores the display spacing. Real connector bends
// span a longer arc or have a larger deviation and are left unchanged.
func smoothHorizonConnectorJunctions(points []basic.OccultationPathPoint) []basic.OccultationPathPoint {
	if len(points) < 4 {
		return points
	}
	result := append([]basic.OccultationPathPoint(nil), points...)
	const (
		maximumAdjacentEdgeKM = 140.0
		maximumChordKM        = 140.0
		minimumTurnDegrees    = 55.0
		minimumDeviationKM    = 5.0
		maximumDeviationKM    = 25.0
	)
	for pass := 0; pass < 2; pass++ {
		changed := false
		for index := 1; index+1 < len(result); index++ {
			// Only the first/last interior sample can be the sparse sample
			// replaced by an independently refined phase endpoint. Interior
			// bends belong to the physical H=0 trajectory and must remain.
			if index != 1 && index+2 != len(result) {
				continue
			}
			first, middle, last := result[index-1], result[index], result[index+1]
			if DistanceKM(first, middle) > maximumAdjacentEdgeKM ||
				DistanceKM(middle, last) > maximumAdjacentEdgeKM ||
				DistanceKM(first, last) > maximumChordKM {
				continue
			}
			turn := 180 - occultationTurnAngleDegrees(
				geodata.GeoPoint{Longitude: first.Longitude, Latitude: first.Latitude},
				geodata.GeoPoint{Longitude: middle.Longitude, Latitude: middle.Latitude},
				geodata.GeoPoint{Longitude: last.Longitude, Latitude: last.Latitude},
			)
			if turn < minimumTurnDegrees {
				continue
			}
			deviation := occultationProjectedPointLineDistanceKM(
				geodata.GeoPoint{Longitude: middle.Longitude, Latitude: middle.Latitude},
				geodata.GeoPoint{Longitude: first.Longitude, Latitude: first.Latitude},
				geodata.GeoPoint{Longitude: last.Longitude, Latitude: last.Latitude},
			)
			if deviation < minimumDeviationKM || deviation > maximumDeviationKM {
				continue
			}
			result = append(result[:index], result[index+1:]...)
			changed = true
			break
		}
		if !changed {
			break
		}
	}
	return result
}

func horizonConnectorIsDegenerate(points []basic.OccultationPathPoint) bool {
	return len(points) >= 2 &&
		DistanceKM(points[0], points[len(points)-1]) <= horizonConnectorDegenerateEndpointDistanceKM
}

type keyedHorizonConnectorCandidate struct {
	key   string
	value horizonConnectorCandidate
}

func temporalHorizonConnectorCandidates(
	footprints []basic.OccultationFootprint,
	endpoints []horizonConnectorEndpoint,
) []keyedHorizonConnectorCandidate {
	result := make([]keyedHorizonConnectorCandidate, 0, 2)
	for firstIndex, first := range endpoints {
		if first.phase != basic.RiseSetPhaseStart && first.phase != basic.RiseSetPhaseEnd {
			continue
		}
		for secondIndex := firstIndex + 1; secondIndex < len(endpoints); secondIndex++ {
			second := endpoints[secondIndex]
			if first.phase == second.phase || first.direction != second.direction ||
				(second.phase != basic.RiseSetPhaseStart && second.phase != basic.RiseSetPhaseEnd) {
				continue
			}
			start, end := first, second
			if end.point.Time.Before(start.point.Time) {
				start, end = end, start
			}
			duration := end.point.Time.Sub(start.point.Time)
			distance := DistanceKM(start.point, end.point)
			if duration <= time.Second || duration > horizonConnectorTemporalWindow ||
				distance <= 1 || distance > horizonConnectorTemporalDistanceKM {
				continue
			}
			points, metric, ok := temporalHorizonConnectorPoints(footprints, start.point, end.point)
			if !ok {
				continue
			}
			points = densifyHorizonConnectorPoints(points, horizonConnectorDensifyEdgeKM)
			key := "temporal/" + string(start.direction) + "/" +
				start.point.Time.UTC().Round(time.Second).Format(time.RFC3339) + "/" +
				end.point.Time.UTC().Round(time.Second).Format(time.RFC3339)
			result = append(result, keyedHorizonConnectorCandidate{
				key: key,
				value: horizonConnectorCandidate{
					connector: HorizonConnector{Direction: start.direction, Points: points},
					metric:    metric,
				},
			})
		}
	}
	return result
}

func temporalHorizonConnectorPoints(
	footprints []basic.OccultationFootprint,
	start, end basic.OccultationPathPoint,
) ([]basic.OccultationPathPoint, float64, bool) {
	if !start.Time.Before(end.Time) {
		return nil, 0, false
	}
	startAnchor := temporalHorizonEndpointAnchorDistance(footprints, start)
	endAnchor := temporalHorizonEndpointAnchorDistance(footprints, end)
	if startAnchor > horizonConnectorMatchDistanceKM || endAnchor > horizonConnectorMatchDistanceKM {
		return nil, 0, false
	}
	points := []basic.OccultationPathPoint{start}
	for _, footprint := range footprints {
		if footprint.Closed || !footprint.Time.After(start.Time) || !footprint.Time.Before(end.Time) {
			continue
		}
		fraction := float64(footprint.Time.Sub(start.Time)) / float64(end.Time.Sub(start.Time))
		target := interpolateOccultationGeoPoint(
			geodata.GeoPoint{Longitude: start.Longitude, Latitude: start.Latitude},
			geodata.GeoPoint{Longitude: end.Longitude, Latitude: end.Latitude}, fraction,
		)
		candidate, distance, ok := temporalHorizonBoundaryEndpointNear(
			footprint, basic.OccultationPathPoint{Longitude: target.Longitude, Latitude: target.Latitude},
		)
		if !ok || distance > horizonConnectorMatchDistanceKM {
			continue
		}
		points = append(points, candidate)
	}
	points = append(points, end)
	sort.SliceStable(points[1:len(points)-1], func(first, second int) bool {
		return points[first+1].Time.Before(points[second+1].Time)
	})
	points = deduplicateTemporalHorizonConnectorPoints(points)
	return points, startAnchor + endAnchor, len(points) >= 2
}

func temporalHorizonEndpointAnchorDistance(
	footprints []basic.OccultationFootprint,
	endpoint basic.OccultationPathPoint,
) float64 {
	minimum := math.Inf(1)
	for _, footprint := range footprints {
		if footprint.Closed || absDuration(footprint.Time.Sub(endpoint.Time)) > horizonConnectorMatchWindow {
			continue
		}
		_, distance, ok := temporalHorizonBoundaryEndpointNear(footprint, endpoint)
		if ok {
			minimum = math.Min(minimum, distance)
		}
	}
	return minimum
}

func temporalHorizonBoundaryEndpointNear(
	footprint basic.OccultationFootprint,
	target basic.OccultationPathPoint,
) (basic.OccultationPathPoint, float64, bool) {
	best := basic.OccultationPathPoint{}
	minimum := math.Inf(1)
	for _, boundary := range footprint.Boundaries {
		if len(boundary) < 2 {
			continue
		}
		for _, candidate := range []basic.OccultationPathPoint{boundary[0], boundary[len(boundary)-1]} {
			if distance := DistanceKM(candidate, target); distance < minimum {
				best, minimum = candidate, distance
			}
		}
	}
	return best, minimum, minimum < math.Inf(1)
}

func deduplicateTemporalHorizonConnectorPoints(
	points []basic.OccultationPathPoint,
) []basic.OccultationPathPoint {
	if len(points) < 2 {
		return points
	}
	result := make([]basic.OccultationPathPoint, 0, len(points))
	for _, point := range points {
		if len(result) > 0 && (!point.Time.After(result[len(result)-1].Time) ||
			DistanceKM(point, result[len(result)-1]) <= 0.001) {
			continue
		}
		result = append(result, point)
	}
	return result
}

func densifyHorizonConnectorPoints(
	points []basic.OccultationPathPoint,
	maximumEdgeKM float64,
) []basic.OccultationPathPoint {
	if len(points) < 2 || maximumEdgeKM <= 0 {
		return points
	}
	result := make([]basic.OccultationPathPoint, 0, len(points)*2)
	for index, point := range points {
		result = append(result, point)
		if index+1 >= len(points) {
			continue
		}
		next := points[index+1]
		distance := DistanceKM(point, next)
		steps := int(math.Ceil(distance / maximumEdgeKM))
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

func selectHorizonConnectorMatch(
	firstMatches, lastMatches []horizonConnectorMatch,
) (horizonConnectorMatch, horizonConnectorMatch, bool) {
	bestMetric := 0.0
	var bestFirst, bestLast horizonConnectorMatch
	found := false
	for _, first := range firstMatches {
		for _, last := range lastMatches {
			if first.endpoint.index == last.endpoint.index ||
				first.endpoint.direction != last.endpoint.direction ||
				absDuration(first.endpoint.point.Time.Sub(last.endpoint.point.Time)) > horizonConnectorMatchWindow {
				continue
			}
			if first.endpoint.phase == last.endpoint.phase {
				if !samePhaseHorizonConnectorAllowed(first.endpoint, last.endpoint) {
					continue
				}
			}
			metric := first.metric + last.metric +
				absDuration(first.endpoint.point.Time.Sub(last.endpoint.point.Time)).Seconds()
			if !found || metric < bestMetric {
				bestFirst, bestLast, bestMetric, found = first, last, metric, true
			}
		}
	}
	return bestFirst, bestLast, found
}

// horizonClosurePoints returns the closing lunar-horizon arc stored after an
// open contact-limb boundary in the corresponding visible polygon. The two
// endpoint matches are deliberately performed on Boundaries by the caller:
// those are the contact-phase points, while this suffix is the physical arc
// that must be drawn between them.
func horizonClosurePoints(
	footprint basic.OccultationFootprint,
	boundaryIndex int,
) []basic.OccultationPathPoint {
	if boundaryIndex < 0 || boundaryIndex >= len(footprint.Boundaries) ||
		boundaryIndex >= len(footprint.Polygons) {
		return nil
	}
	boundary := footprint.Boundaries[boundaryIndex]
	polygon := footprint.Polygons[boundaryIndex]
	if len(boundary) < 2 || len(polygon) <= len(boundary) {
		return nil
	}
	// Footprint construction copies the contact-limb boundary verbatim before
	// appending its horizon closure. Reject a mismatched polygon rather than
	// slicing an unrelated ring by position.
	for index := range boundary {
		if DistanceKM(boundary[index], polygon[index]) > horizonConnectorBoundaryShiftToleranceKM {
			return nil
		}
	}
	points := make([]basic.OccultationPathPoint, 0, 1+len(polygon)-len(boundary))
	points = append(points, boundary[len(boundary)-1])
	for _, point := range polygon[len(boundary):] {
		if DistanceKM(points[len(points)-1], point) <= 1e-6 {
			continue
		}
		points = append(points, point)
	}
	if len(points) < 2 || DistanceKM(points[len(points)-1], boundary[0]) > horizonConnectorBoundaryShiftToleranceKM {
		return nil
	}
	points[len(points)-1] = boundary[0]
	return points
}

func horizonConnectorEndpoints(curves []basic.OccultationRiseSetCurve) []horizonConnectorEndpoint {
	endpoints := make([]horizonConnectorEndpoint, 0, len(curves)*4)
	for curveIndex, curve := range curves {
		for segmentIndex, segment := range curve.Segments {
			if len(segment) == 0 {
				continue
			}
			endpoints = append(endpoints, horizonConnectorEndpoint{
				point: segment[0], phase: curve.Phase, direction: curve.Direction,
				curveIndex: curveIndex, segmentIndex: segmentIndex, atStart: true, index: len(endpoints),
			})
			if len(segment) > 1 {
				endpoints = append(endpoints, horizonConnectorEndpoint{
					point: segment[len(segment)-1], phase: curve.Phase, direction: curve.Direction,
					curveIndex: curveIndex, segmentIndex: segmentIndex, atStart: false, index: len(endpoints),
				})
			}
		}
	}
	return endpoints
}

// samePhaseHorizonConnectorAllowed 允许同相位的独立支路缺口配对；≤5 km 的折叠端点也
// 允许进入选择，以便否决该边界上的其他配对（真正的丢弃发生在 samePhaseFold）。
func samePhaseHorizonConnectorAllowed(
	first, last horizonConnectorEndpoint,
) bool {
	if first.curveIndex != last.curveIndex || first.segmentIndex == last.segmentIndex ||
		first.atStart != last.atStart {
		return false
	}
	return DistanceKM(first.point, last.point) <= horizonConnectorSamePhaseBranchDistanceKM
}

// Exact phase endpoints can lie inside the sparse footprint horizon arc.
// Keep only the source samples between them, in their original order.
func trimHorizonConnectorSamples(points []basic.OccultationPathPoint) []basic.OccultationPathPoint {
	if len(points) < 3 || DistanceKM(points[0], points[len(points)-1]) > horizonConnectorSamePhaseBranchDistanceKM {
		return points
	}
	first, last := points[0], points[len(points)-1]
	const radians = math.Pi / 180
	sinLat, cosLat := math.Sincos(first.Latitude * radians)
	components := func(point basic.OccultationPathPoint) (float64, float64, float64) {
		latitude, longitude := point.Latitude*radians, (point.Longitude-first.Longitude)*radians
		east := math.Cos(latitude) * math.Sin(longitude)
		north := cosLat*math.Sin(latitude) - sinLat*math.Cos(latitude)*math.Cos(longitude)
		radial := sinLat*math.Sin(latitude) + cosLat*math.Cos(latitude)*math.Cos(longitude)
		return east, north, radial
	}
	east, north, radial := components(last)
	length := math.Hypot(east, north)
	if length <= 1e-12 {
		return points
	}
	east, north = east/length, north/length
	total := math.Atan2(length, radial)
	result := make([]basic.OccultationPathPoint, 1, len(points))
	result[0] = first
	previous := 0.0
	for _, point := range points[1 : len(points)-1] {
		x, y, z := components(point)
		progress := math.Atan2(x*east+y*north, z)
		if progress > previous && progress < total {
			result = append(result, point)
			previous = progress
		}
	}
	return append(result, last)
}

func horizonConnectorEndpointMatches(
	point basic.OccultationPathPoint,
	endpoints []horizonConnectorEndpoint,
) []horizonConnectorMatch {
	matches := make([]horizonConnectorMatch, 0, len(endpoints))
	for _, endpoint := range endpoints {
		deltaTime := absDuration(point.Time.Sub(endpoint.point.Time))
		if deltaTime > horizonConnectorMatchWindow {
			continue
		}
		distance := DistanceKM(point, endpoint.point)
		if distance > horizonConnectorMatchDistanceKM {
			continue
		}
		metric := distance + deltaTime.Seconds()*5
		matches = append(matches, horizonConnectorMatch{endpoint: endpoint, metric: metric})
	}
	sort.SliceStable(matches, func(first, second int) bool {
		return matches[first].metric < matches[second].metric
	})
	if len(matches) == 0 {
		return nil
	}
	return matches
}
