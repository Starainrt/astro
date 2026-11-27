package basic

import (
	"math"

	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/tools"
)

// BodyEquatorOrigin 行星赤道坐标的赤经零点 / right ascension origin of planetary equatorial coordinates.
type BodyEquatorOrigin int

const (
	// BodyEquatorOriginIAUNode 行星赤道与 ICRF 赤道的升交点 / ascending node of the planetary equator on the ICRF equator.
	BodyEquatorOriginIAUNode BodyEquatorOrigin = iota
	// BodyEquatorOriginEquinox 行星赤道与瞬时轨道面的春分点 / equinox of the planetary equator and instantaneous orbital plane.
	BodyEquatorOriginEquinox
)

type bodyEquatorialFrame struct {
	pole, node, east Vector3
	normal, equinox  Vector3
	equinoxAngle     float64
	equinoxDefined   bool
}

func bodyEquatorialNodeFrame(orientation BodyOrientation) bodyEquatorialFrame {
	pole, node := bodyPoleNode(orientation)
	return bodyEquatorialFrame{pole: pole, node: node, east: pxp(pole, node)}
}

func bodyEquatorialFrameFromState(orientation BodyOrientation, positionKM, velocityKMPerDay Vector3) (bodyEquatorialFrame, error) {
	frame := bodyEquatorialNodeFrame(orientation)
	angularMomentum := pxp(positionKM, velocityKMPerDay)
	magnitude := vectorMagnitude(angularMomentum)
	if !isFiniteFloat(magnitude) || magnitude <= 0 {
		return bodyEquatorialFrame{}, ErrInvalidObservationInput
	}
	frame.normal = normalizeVector(angularMomentum)
	equinox := pxp(frame.pole, frame.normal)
	magnitude = vectorMagnitude(equinox)
	frame.equinoxDefined = magnitude >= math.Sin(math.Pi/180)
	if frame.equinoxDefined {
		frame.equinox = normalizeVector(equinox)
		frame.equinoxAngle = tools.Limit360(math.Atan2(vectorDot(pxp(frame.node, frame.equinox), frame.pole), vectorDot(frame.node, frame.equinox)) * 180 / math.Pi)
	}
	return frame, nil
}

func bodyEquatorialFrameAt(body SkyBody, jdTDB float64, inst *Instance) (bodyEquatorialFrame, error) {
	orientation, err := BodyOrientationAt(body, jdTDB)
	if err != nil {
		return bodyEquatorialFrame{}, err
	}
	positionKM, velocityKMPerDay, err := bodyHeliocentricState(body, jdTDB, inst)
	if err != nil {
		return bodyEquatorialFrame{}, err
	}
	return bodyEquatorialFrameFromState(orientation, positionKM, velocityKMPerDay)
}

func bodyHeliocentricState(body SkyBody, jdTDB float64, inst *Instance) (positionKM, velocityKMPerDay Vector3, err error) {
	model, _, _, _, ok := bodyOrientationModel(body)
	if !ok || !isFiniteFloat(jdTDB) {
		return positionKM, velocityKMPerDay, ErrInvalidObservationInput
	}
	if inst == nil {
		inst = ephem.Global()
	}
	if inst.Engine() != nil {
		target := ephem.Body(model.planetIndex)
		if model.planetIndex >= 3 {
			target++
		}
		var position, velocity [3]float64
		position, velocity, err = inst.State(target, ephem.Sun, jdTDB)
		positionKM, velocityKMPerDay = Vector3(position), Vector3(velocity)
	} else if inst.Model() == ephem.Legacy {
		const step = 0.01
		positionKM = bodyHeliocentricPositionJ2000(model.planetIndex, jdTDB, inst)
		before := bodyHeliocentricPositionJ2000(model.planetIndex, jdTDB-step, inst)
		after := bodyHeliocentricPositionJ2000(model.planetIndex, jdTDB+step, inst)
		for i := range velocityKMPerDay {
			velocityKMPerDay[i] = (after[i] - before[i]) / (2 * step)
		}
	} else {
		err = ephem.ErrNoEngine
	}
	if err != nil {
		return Vector3{}, Vector3{}, err
	}
	for i := range positionKM {
		if !isFiniteFloat(positionKM[i]) || !isFiniteFloat(velocityKMPerDay[i]) {
			return Vector3{}, Vector3{}, ErrInvalidObservationInput
		}
	}
	return
}

func bodyHeliocentricPositionJ2000(index int, jdTDB float64, inst *Instance) Vector3 {
	longitude := PlanetCoordinateNWithInstance(index, 0, jdTDB, -1, inst)
	latitude := PlanetCoordinateNWithInstance(index, 1, jdTDB, -1, inst)
	radiusKM := PlanetCoordinateNWithInstance(index, 2, jdTDB, -1, inst) * astronomicalUnitKM
	ofDate := eclipticCartesian(longitude, latitude, radiusKM)
	j2000 := eclipticVectorAtReferenceEpoch(ofDate, jdTDB, 2451545)
	return rotateEclipticToEquatorial(j2000, orbitJ2000Obliquity)
}

func bodyEquatorialFrameForOrigin(body SkyBody, jdTDB float64, origin BodyEquatorOrigin, inst *Instance) (bodyEquatorialFrame, error) {
	if origin != BodyEquatorOriginIAUNode && origin != BodyEquatorOriginEquinox {
		return bodyEquatorialFrame{}, ErrInvalidObservationInput
	}
	if origin == BodyEquatorOriginIAUNode {
		orientation, err := BodyOrientationAt(body, jdTDB)
		if err != nil {
			return bodyEquatorialFrame{}, err
		}
		return bodyEquatorialNodeFrame(orientation), nil
	}
	frame, err := bodyEquatorialFrameAt(body, jdTDB, inst)
	if err != nil {
		return bodyEquatorialFrame{}, err
	}
	if !frame.equinoxDefined {
		return bodyEquatorialFrame{}, ErrInvalidObservationInput
	}
	return frame, nil
}

// ToBodyEquatorialWithInstance 指定实例将 ICRF 赤经赤纬转到行星赤道系，时刻为 TT/TDB 儒略日 / transform ICRF RA/Dec to body equatorial coordinates with an explicit instance at a TT/TDB Julian day.
func ToBodyEquatorialWithInstance(body SkyBody, jdTDB, ra, dec float64, origin BodyEquatorOrigin, inst *Instance) (bodyRA, bodyDec float64, err error) {
	return toBodyEquatorial(body, jdTDB, ra, dec, origin, inst)
}

// FromBodyEquatorialWithInstance 指定实例将行星赤道坐标转回 ICRF 赤经赤纬，时刻为 TT/TDB 儒略日 / transform body equatorial coordinates back to ICRF RA/Dec with an explicit instance at a TT/TDB Julian day.
func FromBodyEquatorialWithInstance(body SkyBody, jdTDB, bodyRA, bodyDec float64, origin BodyEquatorOrigin, inst *Instance) (ra, dec float64, err error) {
	return fromBodyEquatorial(body, jdTDB, bodyRA, bodyDec, origin, inst)
}

func toBodyEquatorial(body SkyBody, jdTDB, ra, dec float64, origin BodyEquatorOrigin, inst *Instance) (float64, float64, error) {
	if !isFiniteFloat(ra) || !isFiniteFloat(dec) || dec < -90 || dec > 90 {
		return 0, 0, ErrInvalidObservationInput
	}
	frame, err := bodyEquatorialFrameForOrigin(body, jdTDB, origin, inst)
	if err != nil {
		return 0, 0, err
	}
	direction := raDecToVector(tools.Limit360(ra), dec)
	x, y, z := vectorDot(direction, frame.node), vectorDot(direction, frame.east), vectorDot(direction, frame.pole)
	bodyRA := math.Atan2(y, x) * 180 / math.Pi
	if origin == BodyEquatorOriginEquinox {
		bodyRA -= frame.equinoxAngle
	}
	// atan2 与单位矢量的 asin 投影等价，并避免极点附近的舍入放大。
	bodyDec := math.Atan2(z, math.Hypot(x, y)) * 180 / math.Pi
	return tools.Limit360(bodyRA), bodyDec, nil
}

func fromBodyEquatorial(body SkyBody, jdTDB, bodyRA, bodyDec float64, origin BodyEquatorOrigin, inst *Instance) (float64, float64, error) {
	if !isFiniteFloat(bodyRA) || !isFiniteFloat(bodyDec) || bodyDec < -90 || bodyDec > 90 {
		return 0, 0, ErrInvalidObservationInput
	}
	frame, err := bodyEquatorialFrameForOrigin(body, jdTDB, origin, inst)
	if err != nil {
		return 0, 0, err
	}
	ra := tools.Limit360(bodyRA)
	if origin == BodyEquatorOriginEquinox {
		ra += frame.equinoxAngle
	}
	local := raDecToVector(ra, bodyDec)
	direction := Vector3{}
	for i := range direction {
		direction[i] = frame.node[i]*local[0] + frame.east[i]*local[1] + frame.pole[i]*local[2]
	}
	return tools.Limit360(math.Atan2(direction[1], direction[0]) * 180 / math.Pi), math.Atan2(direction[2], math.Hypot(direction[0], direction[1])) * 180 / math.Pi, nil
}
