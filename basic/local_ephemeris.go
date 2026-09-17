package basic

import "math"

// localEphemerisVectorNode is a pair of Cartesian ephemeris samples at one
// TT. Cartesian interpolation avoids right-ascension wraparound at 0/360.
type localEphemerisVectorNode struct {
	tt          float64
	first, next [3]float64
}

const (
	solarLocalEphemerisNodeCount       = 13
	solarLocalEphemerisStepDays        = 1.0 / 24.0
	occultationLocalEphemerisNodeCount = 49
	occultationLocalEphemerisStepDays  = 2.0 / 24.0
)

func interpolateLocalEphemerisVectors(
	nodes []localEphemerisVectorNode,
	tt float64,
) ([3]float64, [3]float64, bool) {
	const interpolationPoints = 6
	if len(nodes) < interpolationPoints || !finite(tt) {
		return [3]float64{}, [3]float64{}, false
	}
	step := nodes[1].tt - nodes[0].tt
	if !finite(step) || step <= 0 {
		return [3]float64{}, [3]float64{}, false
	}
	u := (tt - nodes[0].tt) / step
	if u < 0 || u > float64(len(nodes)-1) {
		return [3]float64{}, [3]float64{}, false
	}
	start := int(math.Floor(u)) - 2
	if start < 0 {
		start = 0
	}
	if start > len(nodes)-interpolationPoints {
		start = len(nodes) - interpolationPoints
	}
	var weights [interpolationPoints]float64
	for point := 0; point < interpolationPoints; point++ {
		x := float64(start + point)
		weight := 1.0
		for other := 0; other < interpolationPoints; other++ {
			if other == point {
				continue
			}
			xOther := float64(start + other)
			weight *= (u - xOther) / (x - xOther)
		}
		weights[point] = weight
	}
	var first, next [3]float64
	for coordinate := 0; coordinate < 3; coordinate++ {
		for point := 0; point < interpolationPoints; point++ {
			first[coordinate] += weights[point] * nodes[start+point].first[coordinate]
			next[coordinate] += weights[point] * nodes[start+point].next[coordinate]
		}
	}
	return first, next, finiteVector3(first) && finiteVector3(next)
}

func finiteVector3(value [3]float64) bool {
	return finite(value[0]) && finite(value[1]) && finite(value[2])
}

func occultationPathVectorRaDec(vector occultationPathVector) (float64, float64, float64, bool) {
	distance := occultationPathNorm(vector)
	if !finite(distance) || distance <= 0 {
		return 0, 0, 0, false
	}
	ra := math.Atan2(vector.y, vector.x) / rad
	dec := math.Asin(math.Max(-1, math.Min(1, vector.z/distance))) / rad
	return ra, dec, distance, finite(ra) && finite(dec)
}

type solarEclipseLocalEphemeris struct {
	nodes []localEphemerisVectorNode
}

func newSolarEclipseLocalEphemeris(center float64) *solarEclipseLocalEphemeris {
	half := solarLocalEphemerisNodeCount / 2
	nodes := make([]localEphemerisVectorNode, solarLocalEphemerisNodeCount)
	for index := range nodes {
		tt := center + float64(index-half)*solarLocalEphemerisStepDays
		sun, moon := solarEclipseSunMoonEquatorial(tt)
		nodes[index] = localEphemerisVectorNode{
			tt:    tt,
			first: solarEclipseLLRToXYZ(sun[0], sun[1], sun[2]),
			next:  solarEclipseLLRToXYZ(moon[0], moon[1], moon[2]),
		}
	}
	return &solarEclipseLocalEphemeris{nodes: nodes}
}

func (ephemeris *solarEclipseLocalEphemeris) equatorialAt(tt float64) ([3]float64, [3]float64, bool) {
	var empty [3]float64
	if ephemeris == nil {
		return empty, empty, false
	}
	sunXYZ, moonXYZ, ok := interpolateLocalEphemerisVectors(ephemeris.nodes, tt)
	if !ok {
		return empty, empty, false
	}
	return solarEclipseXYZToLLR(sunXYZ[0], sunXYZ[1], sunXYZ[2]),
		solarEclipseXYZToLLR(moonXYZ[0], moonXYZ[1], moonXYZ[2]), true
}

type starOccultationLocalEphemeris struct {
	star  StarCoordinate
	nodes []localEphemerisVectorNode
	dense bool
}

func newStarOccultationLocalEphemeris(center float64, star StarCoordinate) *starOccultationLocalEphemeris {
	half := occultationLocalEphemerisNodeCount / 2
	nodes := make([]localEphemerisVectorNode, occultationLocalEphemerisNodeCount)
	for index := range nodes {
		tt := center + float64(index-half)*occultationLocalEphemerisStepDays
		state := starOccultationEphemerisStateAt(tt, star)
		moon := occultationPathRaDecVector(state.moonRA, state.moonDec, state.moonDistanceKM)
		targetDistance := state.starDistanceKM
		if targetDistance <= 0 {
			// 恒星距离未知时只需方向：按单位球方向装配，几何只用归一化后的矢量。
			targetDistance = 1
		}
		target := occultationPathRaDecVector(state.starRA, state.starDec, targetDistance)
		nodes[index] = localEphemerisVectorNode{
			tt:    tt,
			first: [3]float64{moon.x, moon.y, moon.z},
			next:  [3]float64{target.x, target.y, target.z},
		}
	}
	return &starOccultationLocalEphemeris{star: star, nodes: nodes}
}

func (ephemeris *starOccultationLocalEphemeris) stateAt(tt float64) (starOccultationEphemerisState, bool) {
	moonXYZ, targetXYZ, ok := ephemeris.vectorsAt(tt)
	if !ok {
		return starOccultationEphemerisState{}, false
	}
	moonRA, moonDec, moonDistance, moonOK := occultationPathVectorRaDec(occultationPathVector{
		x: moonXYZ[0], y: moonXYZ[1], z: moonXYZ[2],
	})
	targetRA, targetDec, _, targetOK := occultationPathVectorRaDec(occultationPathVector{
		x: targetXYZ[0], y: targetXYZ[1], z: targetXYZ[2],
	})
	if !moonOK || !targetOK {
		return starOccultationEphemerisState{}, false
	}
	return starOccultationEphemerisState{
		moonRA: moonRA, moonDec: moonDec, moonDistanceKM: moonDistance,
		starRA: targetRA, starDec: targetDec, starDistanceKM: ephemeris.starDistanceKM(),
		valid: true,
	}, true
}

func (ephemeris *starOccultationLocalEphemeris) vectorsAt(tt float64) ([3]float64, [3]float64, bool) {
	if ephemeris == nil {
		return [3]float64{}, [3]float64{}, false
	}
	if ephemeris.dense {
		return interpolateDenseOccultationVectors(ephemeris.nodes, tt)
	}
	return interpolateLocalEphemerisVectors(ephemeris.nodes, tt)
}

func (ephemeris *starOccultationLocalEphemeris) starDistanceKM() float64 {
	if ephemeris.star.ParallaxMas <= 0 {
		return 0
	}
	return 206264806.247 / ephemeris.star.ParallaxMas * occultationPathAstronomicalUnitKM
}

type planetOccultationLocalEphemeris struct {
	nodes []localEphemerisVectorNode
	dense bool
}

func newPlanetOccultationLocalEphemeris(center float64, config planetOccultationConfig) *planetOccultationLocalEphemeris {
	half := occultationLocalEphemerisNodeCount / 2
	nodes := make([]localEphemerisVectorNode, occultationLocalEphemerisNodeCount)
	for index := range nodes {
		tt := center + float64(index-half)*occultationLocalEphemerisStepDays
		state := planetOccultationEphemerisStateAt(tt, config)
		moon := occultationPathRaDecVector(state.moonRA, state.moonDec, state.moonDistanceKM)
		target := occultationPathRaDecVector(state.planetRA, state.planetDec, state.planetDistanceKM)
		nodes[index] = localEphemerisVectorNode{
			tt:    tt,
			first: [3]float64{moon.x, moon.y, moon.z},
			next:  [3]float64{target.x, target.y, target.z},
		}
	}
	return &planetOccultationLocalEphemeris{nodes: nodes}
}

func (ephemeris *planetOccultationLocalEphemeris) stateAt(tt float64) (planetOccultationEphemerisState, bool) {
	moonXYZ, targetXYZ, ok := ephemeris.vectorsAt(tt)
	if !ok {
		return planetOccultationEphemerisState{}, false
	}
	moonRA, moonDec, moonDistance, moonOK := occultationPathVectorRaDec(occultationPathVector{
		x: moonXYZ[0], y: moonXYZ[1], z: moonXYZ[2],
	})
	targetRA, targetDec, planetDistance, targetOK := occultationPathVectorRaDec(occultationPathVector{
		x: targetXYZ[0], y: targetXYZ[1], z: targetXYZ[2],
	})
	if !moonOK || !targetOK {
		return planetOccultationEphemerisState{}, false
	}
	return planetOccultationEphemerisState{
		moonRA: moonRA, moonDec: moonDec, moonDistanceKM: moonDistance,
		planetRA: targetRA, planetDec: targetDec, planetDistanceKM: planetDistance,
		valid: true,
	}, true
}

func (ephemeris *planetOccultationLocalEphemeris) vectorsAt(tt float64) ([3]float64, [3]float64, bool) {
	if ephemeris == nil {
		return [3]float64{}, [3]float64{}, false
	}
	if ephemeris.dense {
		return interpolateDenseOccultationVectors(ephemeris.nodes, tt)
	}
	return interpolateLocalEphemerisVectors(ephemeris.nodes, tt)
}
