package basic

import "math"

type jupiterGalileanMutualEclipseGeometry struct {
	Active, Passive                                    int
	SeparationKM, AxialKM, SunDistanceKM               float64
	UmbraKM, PenumbraKM, PassiveRadiusKM               float64
	MinimumSeparationArcsec, ImpactParameter, Coverage float64
	Annular                                            bool
	PolarKM                                            float64
	PolarAllowed                                       bool
}

type jupiterGalileanEclipseFrame struct {
	jupiterHelio, pole Vector3
}

func newJupiterGalileanEclipseFrame(jde float64, inst *Instance) jupiterGalileanEclipseFrame {
	poleRA, poleDec, _ := jupiterPoleRotation(jde)
	return jupiterGalileanEclipseFrame{
		jupiterHelio: rotateEclipticToEquatorial(jupiterHeliocentricVectorJ2000(jde, inst), orbitJ2000Obliquity),
		pole:         raDecToVector(poleRA, poleDec),
	}
}

func (frame jupiterGalileanEclipseFrame) satellite(jde float64, satellite int) Vector3 {
	state := jupiterGalileanSatelliteCorrectedStateAtET(jde-jupiterGalileanReferenceJD, satellite-1, jupiterGalileanUseSolarLongPeriod(jde))
	return Vector3{frame.jupiterHelio[0] + state.X, frame.jupiterHelio[1] + state.Y, frame.jupiterHelio[2] + state.Z}
}

func jupiterGalileanMutualEclipseGeometryAt(jde float64, first, second int, inst *Instance) (jupiterGalileanMutualEclipseGeometry, bool) {
	first, second, ok := normalizeJupiterGalileanMutualPair(first, second)
	if !ok || !isFinite(jde) {
		return jupiterGalileanMutualEclipseGeometry{}, false
	}
	frame := newJupiterGalileanEclipseFrame(jde, inst)
	if vectorMagnitude(frame.jupiterHelio) <= 0 {
		return jupiterGalileanMutualEclipseGeometry{}, false
	}
	firstVector, secondVector := frame.satellite(jde, first), frame.satellite(jde, second)
	if vectorMagnitude(firstVector) > vectorMagnitude(secondVector) {
		first, second = second, first
		firstVector, secondVector = secondVector, firstVector
	}
	geometry, valid := jupiterGalileanEclipseProjection(firstVector, secondVector, first, second)
	if !valid {
		return geometry, false
	}
	axis := normalizeVector(firstVector)
	delta := Vector3{secondVector[0] - firstVector[0], secondVector[1] - firstVector[1], secondVector[2] - firstVector[2]}
	axial := vectorDot(delta, axis)
	perpendicular := Vector3{delta[0] - axial*axis[0], delta[1] - axial*axis[1], delta[2] - axial*axis[2]}
	geometry.PolarKM = math.Abs(vectorDot(perpendicular, frame.pole)) * astronomicalUnitKM
	geometry.PolarAllowed = jupiterGalileanMutualPolarAllows(vectorDot(normalizeVector(frame.jupiterHelio), frame.pole))
	return geometry, true
}

func jupiterGalileanEclipseProjection(activeVector, passiveVector Vector3, active, passive int) (jupiterGalileanMutualEclipseGeometry, bool) {
	axis := normalizeVector(activeVector)
	delta := Vector3{passiveVector[0] - activeVector[0], passiveVector[1] - activeVector[1], passiveVector[2] - activeVector[2]}
	axial := vectorDot(delta, axis)
	if axial <= 0 || !isFinite(axial) {
		return jupiterGalileanMutualEclipseGeometry{}, false
	}
	perpendicular := Vector3{delta[0] - axial*axis[0], delta[1] - axial*axis[1], delta[2] - axial*axis[2]}
	separation := vectorMagnitude(perpendicular) * astronomicalUnitKM
	distance := vectorMagnitude(activeVector) * astronomicalUnitKM
	activeRadius := jupiterGalileanSatelliteRadiusJupiterRadii(active) * jupiterGalileanEquatorialRadiusKM
	passiveRadius := jupiterGalileanSatelliteRadiusJupiterRadii(passive) * jupiterGalileanEquatorialRadiusKM
	umbra, penumbra := jupiterGalileanEclipseConeRadii(axial*astronomicalUnitKM, distance, activeRadius)
	coverage, annular := jupiterGalileanEclipseDiskMetrics(separation, umbra, passiveRadius)
	return jupiterGalileanMutualEclipseGeometry{
		Active: active, Passive: passive, SeparationKM: separation, AxialKM: axial * astronomicalUnitKM,
		SunDistanceKM: distance, UmbraKM: umbra, PenumbraKM: penumbra, PassiveRadiusKM: passiveRadius,
		MinimumSeparationArcsec: math.Atan2(separation/astronomicalUnitKM, vectorDot(passiveVector, axis)) * deg * 3600,
		ImpactParameter:         separation / (penumbra + passiveRadius), Coverage: coverage, Annular: annular,
	}, true
}

func jupiterGalileanEclipseConeRadii(axialKM, distanceKM, activeRadiusKM float64) (float64, float64) {
	solarRadiusKM := solarRadiusAU * astronomicalUnitKM
	return activeRadiusKM - axialKM*(solarRadiusKM-activeRadiusKM)/distanceKM, activeRadiusKM + axialKM*(solarRadiusKM+activeRadiusKM)/distanceKM
}

func jupiterGalileanEclipseDiskMetrics(separation, umbra, passiveRadius float64) (float64, bool) {
	if umbra <= 0 {
		return 0, false
	}
	_, coverage, annular := jupiterGalileanDiskMetrics(separation, umbra, passiveRadius)
	return coverage, annular
}

func jupiterGalileanEclipseReception(jde float64, passive int, inst *Instance) float64 {
	vector := newJupiterGalileanEclipseFrame(jde, inst).satellite(jde, passive)
	reception := jde
	for iteration := 0; iteration < 8; iteration++ {
		earth := rotateEclipticToEquatorial(earthHeliocentricVectorJ2000(reception, inst), orbitJ2000Obliquity)
		relative := Vector3{vector[0] - earth[0], vector[1] - earth[1], vector[2] - earth[2]}
		next := jde + lightTimeDaysPerAU*vectorMagnitude(relative)
		if math.Abs(next-reception) < 1e-12 {
			return next
		}
		reception = next
	}
	return reception
}

func jupiterGalileanEclipseEmission(jde float64, passive int, inst *Instance) float64 {
	context := newJupiterGalileanObservationContext(jde, inst)
	return jde - lightTimeDaysPerAU*context.observationForSatellite(passive-1).Distance
}

type jupiterGalileanMutualGeometry struct {
	Active  int
	Passive int

	ActiveX  float64
	ActiveY  float64
	ActiveZ  float64
	PassiveX float64
	PassiveY float64
	PassiveZ float64

	Separation      float64
	ImpactParameter float64
	Coverage        float64
	Annular         bool
	Polar           float64
	PolarAllowed    bool
}

func jupiterGalileanMutualGeometryAt(jde float64, first, second int, instances ...*Instance) (jupiterGalileanMutualGeometry, bool) {
	if !isFinite(jde) || first < 1 || first > 4 || second < 1 || second > 4 || first == second {
		return jupiterGalileanMutualGeometry{}, false
	}
	if first > second {
		first, second = second, first
	}
	context := newJupiterGalileanObservationContext(jde, instances...)
	if context.jupiterDistance == 0 {
		return jupiterGalileanMutualGeometry{}, false
	}
	firstObservation := context.observationForSatellite(first - 1)
	secondObservation := context.observationForSatellite(second - 1)
	if !isFinite(firstObservation.OffsetXJupiterRadii) || !isFinite(firstObservation.OffsetYJupiterRadii) || !isFinite(firstObservation.OffsetZJupiterRadii) || !isFinite(secondObservation.OffsetXJupiterRadii) || !isFinite(secondObservation.OffsetYJupiterRadii) || !isFinite(secondObservation.OffsetZJupiterRadii) {
		return jupiterGalileanMutualGeometry{}, false
	}

	active := firstObservation
	passive := secondObservation
	activeIndex := first
	passiveIndex := second
	if firstObservation.OffsetZJupiterRadii > secondObservation.OffsetZJupiterRadii {
		active = secondObservation
		passive = firstObservation
		activeIndex = second
		passiveIndex = first
	}
	separation := math.Hypot(active.OffsetXJupiterRadii-passive.OffsetXJupiterRadii, active.OffsetYJupiterRadii-passive.OffsetYJupiterRadii)
	activeRadius := jupiterGalileanSatelliteRadiusJupiterRadii(activeIndex)
	passiveRadius := jupiterGalileanSatelliteRadiusJupiterRadii(passiveIndex)
	impact, coverage, annular := jupiterGalileanDiskMetrics(separation, activeRadius, passiveRadius)
	return jupiterGalileanMutualGeometry{
		Active:          activeIndex,
		Passive:         passiveIndex,
		ActiveX:         active.OffsetXJupiterRadii,
		ActiveY:         active.OffsetYJupiterRadii,
		ActiveZ:         active.OffsetZJupiterRadii,
		PassiveX:        passive.OffsetXJupiterRadii,
		PassiveY:        passive.OffsetYJupiterRadii,
		PassiveZ:        passive.OffsetZJupiterRadii,
		Separation:      separation,
		ImpactParameter: impact,
		Coverage:        coverage,
		Annular:         annular,
		Polar:           math.Abs((active.OffsetXJupiterRadii-passive.OffsetXJupiterRadii)*vectorDot(context.bodyZ, context.east) + (active.OffsetYJupiterRadii-passive.OffsetYJupiterRadii)*vectorDot(context.bodyZ, context.north)),
		PolarAllowed:    jupiterGalileanMutualPolarAllows(vectorDot(context.earthDirection, context.bodyZ)),
	}, true
}

func jupiterGalileanMutualPolarAllows(sine float64) bool {
	return isFinite(sine) && math.Abs(sine) <= math.Sin(3.9*rad)
}

func jupiterGalileanMutualSeasonAllows(jde float64, first, second int, instances ...*Instance) bool {
	if !isFinite(jde) || first < 1 || first > 4 || second < 1 || second > 4 || first == second {
		return false
	}
	seasonLimit := jupiterGalileanMutualSeasonLimit(first, second)
	declination := math.Abs(jupiterGalileanMutualEarthDeclination(jde, instances...))
	return isFinite(declination) && declination <= seasonLimit
}

func jupiterGalileanMutualSeasonLimit(first, second int) float64 {
	if first < 1 || first > 4 || second < 1 || second > 4 || first == second {
		return 0
	}
	firstAxis := jupiterGalileanMutualSemimajorAxis(first)
	secondAxis := jupiterGalileanMutualSemimajorAxis(second)
	firstInclination := jupiterGalileanMutualMaximumInclination(first)
	secondInclination := jupiterGalileanMutualMaximumInclination(second)
	denominator := math.Abs(secondAxis - firstAxis)
	if denominator == 0 {
		return 0
	}
	sine := 1.2 * (jupiterGalileanSatelliteRadiusJupiterRadii(first) + jupiterGalileanSatelliteRadiusJupiterRadii(second) + firstAxis*math.Sin(firstInclination*rad) + secondAxis*math.Sin(secondInclination*rad)) / denominator
	if sine >= 1 {
		return 90
	}
	if sine <= 0 {
		return 0
	}
	return math.Asin(sine) * deg
}

func jupiterGalileanMutualEarthDeclination(jde float64, instances ...*Instance) float64 {
	context := newJupiterGalileanObservationContext(jde, instances...)
	if context.jupiterDistance == 0 {
		return math.NaN()
	}
	sine := vectorDot(context.earthDirection, context.bodyZ)
	if sine > 1 {
		sine = 1
	} else if sine < -1 {
		sine = -1
	}
	return math.Asin(sine) * deg
}

func jupiterGalileanMutualSemimajorAxis(satellite int) float64 {
	switch satellite {
	case 1:
		return 5.899599097645
	case 2:
		return 9.387190820450
	case 3:
		return 14.972774244034
	case 4:
		return 26.336544997512
	default:
		return math.NaN()
	}
}

func jupiterGalileanMutualMaximumInclination(satellite int) float64 {
	switch satellite {
	case 1:
		return 0.053544831525
	case 2:
		return 0.494132301812
	case 3:
		return 0.208027387931
	case 4:
		return 0.295550746319
	default:
		return math.NaN()
	}
}

func jupiterGalileanDiskMetrics(separation, activeRadius, passiveRadius float64) (float64, float64, bool) {
	if !isFinite(separation) || !isFinite(activeRadius) || !isFinite(passiveRadius) || separation < 0 || activeRadius <= 0 || passiveRadius <= 0 {
		return 0, 0, false
	}
	impact := separation / (activeRadius + passiveRadius)
	overlap := jupiterGalileanDiskOverlapArea(separation, activeRadius, passiveRadius)
	coverage := overlap / (math.Pi * passiveRadius * passiveRadius)
	if coverage < 0 {
		coverage = 0
	} else if coverage > 1 {
		coverage = 1
	}
	return impact, coverage, separation+activeRadius <= passiveRadius
}

func jupiterGalileanDiskOverlapArea(separation, firstRadius, secondRadius float64) float64 {
	if separation >= firstRadius+secondRadius {
		return 0
	}
	if separation <= math.Abs(firstRadius-secondRadius) {
		smallerRadius := math.Min(firstRadius, secondRadius)
		return math.Pi * smallerRadius * smallerRadius
	}
	firstCosine := (separation*separation + firstRadius*firstRadius - secondRadius*secondRadius) / (2 * separation * firstRadius)
	secondCosine := (separation*separation + secondRadius*secondRadius - firstRadius*firstRadius) / (2 * separation * secondRadius)
	firstCosine = math.Max(-1, math.Min(1, firstCosine))
	secondCosine = math.Max(-1, math.Min(1, secondCosine))
	radical := (-separation + firstRadius + secondRadius) * (separation + firstRadius - secondRadius) * (separation - firstRadius + secondRadius) * (separation + firstRadius + secondRadius)
	return firstRadius*firstRadius*math.Acos(firstCosine) + secondRadius*secondRadius*math.Acos(secondCosine) - 0.5*math.Sqrt(math.Max(0, radical))
}
