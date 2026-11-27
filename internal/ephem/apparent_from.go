package ephem

import (
	"errors"
	"math"
)

// ApparentFrom 站心 ICRF 视方向、观测时刻几何距 km 与光行时日，偏移为 km 和 km/日 / topocentric apparent ICRF direction, reception-epoch geometric km and light-time days, with km and km/day offsets.
func ApparentFrom(engine Engine, center Body, observer Body, offsetKM, offsetVelKMPerDay [3]float64, target Body, jdTT float64) (dir [3]float64, distKM, tauDays float64, err error) {
	if center != SSB && center != Sun || IsNilEngine(engine) || observer == target || math.IsNaN(jdTT) || math.IsInf(jdTT, 0) || !apparentFromFinite(offsetKM) || !apparentFromFinite(offsetVelKMPerDay) {
		return dir, 0, 0, ErrUnsupportedApparent
	}
	position, velocity, err := engine.State(observer, center, jdTT)
	if err != nil || !apparentFromFinite(position) || !apparentFromFinite(velocity) {
		return dir, 0, 0, apparentFromStateError(err)
	}
	for i := range position {
		position[i] += offsetKM[i]
		velocity[i] += offsetVelKMPerDay[i]
	}
	if !apparentFromFinite(position) || !(genericLen(velocity) < genericCKMPerDay) {
		return dir, 0, 0, ErrUnsupportedApparent
	}
	now, _, err := engine.State(target, center, jdTT)
	if err != nil || !apparentFromFinite(now) {
		return dir, 0, 0, apparentFromStateError(err)
	}
	g := genericSub(now, position)
	distKM = genericLen(g)
	if !(distKM > 0) || math.IsInf(distKM, 0) {
		return dir, 0, 0, ErrUnsupportedApparent
	}
	tauDays = distKM / genericCKMPerDay
	var retarded [3]float64
	for k := 0; k < 3; k++ {
		retarded, _, err = engine.State(target, center, jdTT-tauDays)
		if err != nil || !apparentFromFinite(retarded) {
			return dir, 0, 0, apparentFromStateError(err)
		}
		g = genericSub(retarded, position)
		length := genericLen(g)
		if !(length > 0) || math.IsInf(length, 0) {
			return dir, 0, 0, ErrUnsupportedApparent
		}
		tauDays = length / genericCKMPerDay
	}
	q := genericScale(g, 1/genericLen(g))
	if target != Sun && target != Moon && center == Sun {
		q = genericDeflect(q, retarded, position)
	} else if target != Sun && target != Moon {
		sunThen, _, sunErr := engine.State(Sun, SSB, jdTT-tauDays)
		sunNow, _, sunNowErr := engine.State(Sun, SSB, jdTT)
		if sunErr == nil && sunNowErr == nil {
			q = genericDeflect(q, genericSub(retarded, sunThen), genericSub(position, sunNow))
		}
	}
	dir = genericAberrate(q, velocity)
	if !apparentFromFinite(dir) {
		return [3]float64{}, 0, 0, ErrUnsupportedApparent
	}
	return dir, distKM, tauDays, nil
}

func apparentFromStateError(err error) error {
	if errors.Is(err, ErrOutOfRange) {
		return err
	}
	return ErrUnsupportedApparent
}

func apparentFromFinite(vector [3]float64) bool {
	for _, value := range vector {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}
