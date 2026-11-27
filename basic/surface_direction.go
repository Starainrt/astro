package basic

import (
	"math"

	"github.com/starainrt/astro/internal/ephem"
)

const surfaceSpeedOfLightKMPerDay = 299792.458 * 86400

type surfaceDirection struct {
	direction           Vector3
	distanceKM, tauDays float64
	frame               bodyObserverFrame
}

func surfaceBodyID(body SkyBody) (ephem.Body, bool) {
	switch body {
	case SkySun:
		return ephem.Sun, true
	case SkyEarth:
		return ephem.Earth, true
	case SkyMercury:
		return ephem.Mercury, true
	case SkyVenus:
		return ephem.Venus, true
	case SkyMars:
		return ephem.Mars, true
	case SkyJupiter:
		return ephem.Jupiter, true
	case SkySaturn:
		return ephem.Saturn, true
	case SkyUranus:
		return ephem.Uranus, true
	case SkyNeptune:
		return ephem.Neptune, true
	default:
		return 0, false
	}
}

func surfaceSpinVelocity(frame bodyObserverFrame) Vector3 {
	pole := raDecToVector(frame.orientation.PoleRA, frame.orientation.PoleDec)
	velocity := pxp(pole, frame.positionKM)
	for i := range velocity {
		velocity[i] *= frame.orientation.RotationRate * math.Pi / 180
	}
	return velocity
}

func surfaceObserverState(body SkyBody, frame bodyObserverFrame, jdTT float64, inst *Instance) (position, velocity Vector3, err error) {
	if inst.Engine() != nil {
		id, _ := surfaceBodyID(body)
		var p, v [3]float64
		p, v, err = inst.State(id, ephem.Sun, jdTT)
		position, velocity = Vector3(p), Vector3(v)
	} else if inst.Model() == ephem.Legacy {
		position, velocity, err = bodyHeliocentricState(body, jdTT, inst)
	} else {
		err = ephem.ErrNoEngine
	}
	if err != nil {
		return Vector3{}, Vector3{}, err
	}
	spin := surfaceSpinVelocity(frame)
	for i := range position {
		position[i] += frame.positionKM[i]
		velocity[i] += spin[i]
		if !isFiniteFloat(position[i]) || !isFiniteFloat(velocity[i]) {
			return Vector3{}, Vector3{}, ErrInvalidObservationInput
		}
	}
	return position, velocity, nil
}

func surfaceDirectionAt(target SkyBody, observer BodyObserver, jdTT float64, inst *Instance) (surfaceDirection, error) {
	targetID, ok := surfaceBodyID(target)
	if !ok || target == observer.Body {
		return surfaceDirection{}, ErrInvalidObservationInput
	}
	frame, err := bodyObserverFrameAt(observer, jdTT)
	if err != nil {
		return surfaceDirection{}, err
	}
	if inst == nil {
		inst = ephem.Global()
	}
	if inst.Engine() != nil {
		return surfaceEngineDirection(inst.Engine(), observer.Body, targetID, frame, jdTT)
	}
	position, velocity, err := surfaceObserverState(observer.Body, frame, jdTT, inst)
	if err != nil {
		return surfaceDirection{}, err
	}
	targetNow := surfaceLegacyTarget(target, jdTT, inst)
	var geometric Vector3
	for i := range geometric {
		geometric[i] = targetNow[i] - position[i]
	}
	distance := vectorMagnitude(geometric)
	if !isFiniteFloat(distance) || distance <= 0 {
		return surfaceDirection{}, ErrInvalidObservationInput
	}
	tau := distance / surfaceSpeedOfLightKMPerDay
	for k := 0; k < 3; k++ {
		retarded := surfaceLegacyTarget(target, jdTT-tau, inst)
		for i := range geometric {
			geometric[i] = retarded[i] - position[i]
		}
		length := vectorMagnitude(geometric)
		if !isFiniteFloat(length) || length <= 0 {
			return surfaceDirection{}, ErrInvalidObservationInput
		}
		tau = length / surfaceSpeedOfLightKMPerDay
	}
	direction, err := surfaceAberrate(normalizeVector(geometric), velocity)
	if err != nil {
		return surfaceDirection{}, err
	}
	return surfaceDirection{direction, distance, tau, frame}, nil
}

func surfaceEngineDirection(engine ephem.Engine, observer SkyBody, target ephem.Body, frame bodyObserverFrame, jdTT float64) (surfaceDirection, error) {
	observerID, _ := surfaceBodyID(observer)
	direction, distance, tau, err := ephem.ApparentFrom(engine, ephem.Sun, observerID, [3]float64(frame.positionKM), [3]float64(surfaceSpinVelocity(frame)), target, jdTT)
	if err != nil {
		return surfaceDirection{}, err
	}
	return surfaceDirection{Vector3(direction), distance, tau, frame}, nil
}

func surfaceLegacyTarget(target SkyBody, jdTT float64, inst *Instance) Vector3 {
	if target == SkySun {
		return Vector3{}
	}
	index := -1
	if target != SkyEarth {
		model, _, _, _, _ := bodyOrientationModel(target)
		index = model.planetIndex
	}
	return bodyHeliocentricPositionJ2000(index, jdTT, inst)
}

func surfaceAberrate(direction, velocity Vector3) (Vector3, error) {
	beta := Vector3{}
	for i := range beta {
		beta[i] = velocity[i] / surfaceSpeedOfLightKMPerDay
	}
	speedSquared := vectorDot(beta, beta)
	if !isFiniteFloat(speedSquared) || speedSquared >= 1 {
		return Vector3{}, ErrInvalidObservationInput
	}
	inverseGamma := math.Sqrt(1 - speedSquared)
	projection := vectorDot(direction, beta)
	weight := 1 + projection/(1+inverseGamma)
	var result Vector3
	for i := range result {
		result[i] = (inverseGamma*direction[i] + weight*beta[i]) / (1 + projection)
	}
	length := vectorMagnitude(result)
	if !isFiniteFloat(length) || length <= 0 {
		return Vector3{}, ErrInvalidObservationInput
	}
	for i := range result {
		result[i] /= length
	}
	return result, nil
}

func surfaceStarDirectionAt(star StarCoordinate, observer BodyObserver, jdTT float64, inst *Instance) (surfaceDirection, error) {
	if star.Validate() != nil || star.Frame != CoordinateFrameICRS && star.Frame != CoordinateFrameJ2000 {
		return surfaceDirection{}, ErrInvalidObservationInput
	}
	frame, err := bodyObserverFrameAt(observer, jdTT)
	if err != nil {
		return surfaceDirection{}, err
	}
	if inst == nil {
		inst = ephem.Global()
	}
	position, velocity, err := surfaceObserverState(observer.Body, frame, jdTT, inst)
	if err != nil {
		return surfaceDirection{}, err
	}
	ra, dec := star.RA, star.Dec
	if star.Frame == CoordinateFrameICRS {
		ra, dec = starICRSToMeanJ2000RaDec(ra, dec)
	}
	epochDistance := starEpochDistanceAU(star.parallaxMas())
	if !isFiniteFloat(epochDistance) || (star.ParallaxMas > 0 || star.DistanceLightYear > 0) && epochDistance <= 0 {
		return surfaceDirection{}, ErrInvalidObservationInput
	}
	ra, dec, distanceAU := starProperMotionRaDec(jdTT, star, ra, dec, epochDistance)
	if !isFiniteFloat(ra) || !isFiniteFloat(dec) || !isFiniteFloat(distanceAU) || epochDistance > 0 && distanceAU <= 0 {
		return surfaceDirection{}, ErrInvalidObservationInput
	}
	mean := raDecToVector(ra, dec)
	bias := FrameBiasMatrix()
	var direction Vector3
	for row := range direction {
		for col := range mean {
			direction[row] += bias[col][row] * mean[col]
		}
	}
	distanceKM := 0.0
	if distanceAU > 0 {
		for i := range direction {
			direction[i] = distanceAU*astronomicalUnitKM*direction[i] - position[i]
		}
		distanceKM = vectorMagnitude(direction)
		if !isFiniteFloat(distanceKM) || distanceKM <= 0 {
			return surfaceDirection{}, ErrInvalidObservationInput
		}
	}
	direction, err = surfaceAberrate(normalizeVector(direction), velocity)
	if err != nil {
		return surfaceDirection{}, err
	}
	return surfaceDirection{direction, distanceKM, 0, frame}, nil
}
