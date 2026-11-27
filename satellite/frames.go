package satellite

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/satellite/internal/sgp4"
)

const (
	earthRadiusKM     = 6378.137
	earthFlattening   = 1 / 298.257223563
	earthRotationRate = 7.29211514670698e-5
	degreesToRadians  = math.Pi / 180
)

type satelliteOrbit struct {
	epoch  time.Time
	record sgp4.Record
}

func newSatelliteOrbit(e Elements) (*satelliteOrbit, error) {
	if !validElements(e) {
		return nil, ErrInvalidElements
	}
	day, fraction := propagationEpochParts(e.Epoch)
	motionFactor := 1440 / (2 * math.Pi)
	orbit := &satelliteOrbit{epoch: e.Epoch}
	orbit.record.Init(sgp4.Elements{
		EpochDays:                (day - 2433281.5) + fraction,
		BStar:                    e.BStar,
		Eccentricity:             e.Eccentricity,
		InclinationRad:           e.Inclination * degreesToRadians,
		NodeRad:                  e.RightAscension * degreesToRadians,
		ArgumentRad:              e.ArgumentOfPerigee * degreesToRadians,
		MeanAnomalyRad:           e.MeanAnomaly * degreesToRadians,
		MeanMotionRadPerMin:      e.MeanMotion / motionFactor,
		MeanMotionDotRadPerMin2:  e.MeanMotionDot / (motionFactor * 1440),
		MeanMotionDDotRadPerMin3: e.MeanMotionDDot / (motionFactor * 1440 * 1440),
	})
	return orbit, nil
}

func (o *satelliteOrbit) stateAt(t time.Time) (State, error) {
	elapsed := t.Sub(o.epoch)
	if t.IsZero() || !o.epoch.Add(elapsed).Equal(t) {
		return State{}, ErrInvalidInput
	}
	r, v, code := o.record.Propagate(float64(elapsed.Nanoseconds()) / 6e10)
	if err := propagationError(code); err != nil {
		return State{}, err
	}
	for i := range r {
		if !finite(r[i]) || !finite(v[i]) {
			return State{}, ErrInvalidElements
		}
	}
	return State{Time: t, PositionKM: r, VelocityKMS: v}, nil
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func rotateZ(theta float64, v [3]float64) [3]float64 {
	s, c := math.Sincos(theta)
	return [3]float64{c*v[0] + s*v[1], -s*v[0] + c*v[1], v[2]}
}

func earthFixed(s State) (r, v [3]float64) {
	jdUT1 := basic.UTC2UT1(basic.Date2JD(s.Time.UTC()))
	theta := sgp4.GSTime(jdUT1)
	r = rotateZ(theta, s.PositionKM)
	v = rotateZ(theta, s.VelocityKMS)
	v[0] += earthRotationRate * r[1]
	v[1] -= earthRotationRate * r[0]
	return
}

func vectorNorm(v [3]float64) float64 { return math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2]) }
