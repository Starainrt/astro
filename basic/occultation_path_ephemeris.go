package basic

import (
	"math"
)

const (
	occultationDenseEphemerisNodeCount = 193
	occultationDenseEphemerisStepDays  = 30.0 / 1440.0
	occultationDenseEphemerisTolerance = 1e-8 // relaxed for vector-based apparent positions
)

// Width extrema near a limb can change branches under tiny frame differences.
// Keep this metadata on the full-term model without moving the traced points.
func correctOccultationCenterWidths(points []OccultationPathPoint, frameAt occultationPathFrameFunc) {
	for index := range points {
		_, _, width, ok := occultationPathLimitsAndWidthForFrame(centerTimeTT(points[index].Time), frameAt)
		points[index].WidthKM = 0
		if ok {
			points[index].WidthKM = width
		}
	}
}

func (cache *starOccultationEventCache) preparePathEphemeris(center float64, algorithm OccultationPathAlgorithm) {
	if algorithm != OccultationPathAlgorithmExact {
		nodes := newDenseOccultationEphemerisNodes(center, func(tt float64) ([3]float64, [3]float64) {
			state := starOccultationEphemerisStateAt(tt, cache.star, cache.inst)
			moon := occultationPathRaDecVector(state.moonRA, state.moonDec, state.moonDistanceKM)
			distance := state.starDistanceKM
			if distance <= 0 {
				distance = 1
			}
			target := occultationPathRaDecVector(state.starRA, state.starDec, distance)
			return [3]float64{moon.x, moon.y, moon.z}, [3]float64{target.x, target.y, target.z}
		})
		if len(nodes) > 0 {
			cache.local = newStarOccultationLocalEphemerisFromNodes(center, cache.star, nodes, true, cache.inst)
		}
	}
	cache.prepareLocalEphemeris(center)
}

func (cache *planetOccultationEventCache) preparePathEphemeris(center float64, algorithm OccultationPathAlgorithm) {
	if algorithm != OccultationPathAlgorithmExact {
		nodes := newDenseOccultationEphemerisNodes(center, func(tt float64) ([3]float64, [3]float64) {
			state := planetOccultationEphemerisStateAt(tt, cache.config, cache.inst)
			moon := occultationPathRaDecVector(state.moonRA, state.moonDec, state.moonDistanceKM)
			target := occultationPathRaDecVector(state.planetRA, state.planetDec, state.planetDistanceKM)
			return [3]float64{moon.x, moon.y, moon.z}, [3]float64{target.x, target.y, target.z}
		})
		if len(nodes) > 0 {
			cache.local = &planetOccultationLocalEphemeris{nodes: nodes, dense: true}
		}
	}
	cache.prepareLocalEphemeris(center)
}

// Midpoint checks reject inaccurate or invalid tables before any path state
// is cached. They are sampled safeguards, not a rigorous global error bound.
func newDenseOccultationEphemerisNodes(
	center float64,
	sample func(float64) ([3]float64, [3]float64),
) []localEphemerisVectorNode {
	if !finite(center) {
		return nil
	}
	nodes := make([]localEphemerisVectorNode, occultationDenseEphemerisNodeCount)
	for index := range nodes {
		tt := center + float64(index-len(nodes)/2)*occultationDenseEphemerisStepDays
		first, next := sample(tt)
		if !finiteVector3(first) || !finiteVector3(next) {
			return nil
		}
		nodes[index] = localEphemerisVectorNode{tt: tt, first: first, next: next}
	}
	for index := 1; index < len(nodes); index++ {
		tt := (nodes[index-1].tt + nodes[index].tt) / 2
		first, next, ok := interpolateDenseOccultationVectors(nodes, tt)
		exactFirst, exactNext := sample(tt)
		if !ok || !denseOccultationVectorAccurate(first, exactFirst) || !denseOccultationVectorAccurate(next, exactNext) {
			return nil
		}
	}
	return nodes
}

func denseOccultationVectorAccurate(approximate, exact [3]float64) bool {
	if !finiteVector3(approximate) || !finiteVector3(exact) {
		return false
	}
	norm, difference := 0.0, 0.0
	for index := range exact {
		norm += exact[index] * exact[index]
		delta := approximate[index] - exact[index]
		difference += delta * delta
	}
	return norm > 0 && finite(norm) && difference <= norm*occultationDenseEphemerisTolerance*occultationDenseEphemerisTolerance
}

func interpolateDenseOccultationVectors(
	nodes []localEphemerisVectorNode,
	tt float64,
) ([3]float64, [3]float64, bool) {
	const points = 6
	if len(nodes) < points || !finite(tt) || tt < nodes[0].tt || tt > nodes[len(nodes)-1].tt {
		return [3]float64{}, [3]float64{}, false
	}
	step := nodes[1].tt - nodes[0].tt
	if !finite(step) || step <= 0 {
		return [3]float64{}, [3]float64{}, false
	}
	start := int(math.Floor((tt-nodes[0].tt)/step)) - 2
	if start < 0 {
		start = 0
	}
	if start > len(nodes)-points {
		start = len(nodes) - points
	}
	var first, next [3]float64
	for point := 0; point < points; point++ {
		weight := 1.0
		// Rounded Julian days are not exactly uniformly spaced. Preserve the
		// actual node times, leaving the legacy/solar interpolator unchanged.
		for other := 0; other < points; other++ {
			if other != point {
				weight *= (tt - nodes[start+other].tt) / (nodes[start+point].tt - nodes[start+other].tt)
			}
		}
		for coordinate := 0; coordinate < 3; coordinate++ {
			first[coordinate] += weight * nodes[start+point].first[coordinate]
			next[coordinate] += weight * nodes[start+point].next[coordinate]
		}
	}
	return first, next, finiteVector3(first) && finiteVector3(next)
}
