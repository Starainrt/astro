package satellite

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/starainrt/astro/satellite/internal/sgp4"
)

// ErrMeanEccentricity 平均偏心率越界 / mean eccentricity outside its valid range.
var ErrMeanEccentricity = errors.New("mean eccentricity outside valid range")

// ErrMeanMotion 平均运动非正 / mean motion is not positive.
var ErrMeanMotion = errors.New("mean motion is not positive")

// ErrPerturbedEccentricity 摄动偏心率越界 / perturbed eccentricity outside its valid range.
var ErrPerturbedEccentricity = errors.New("perturbed eccentricity outside valid range")

// ErrSemiLatusRectum 半通径为负 / negative semi-latus rectum.
var ErrSemiLatusRectum = errors.New("negative semi-latus rectum")

// ErrDecayed 卫星轨道已衰减 / satellite orbit has decayed.
var ErrDecayed = errors.New("satellite orbit has decayed")

// State 卫星TEME位置与速度 / satellite TEME position and velocity.
type State struct {
	Time        time.Time
	PositionKM  [3]float64
	VelocityKMS [3]float64
}

// Propagate 按UTC时差传播根数，保留输入时区 / propagate using UTC elapsed time and preserve the input time zone.
func Propagate(e Elements, t time.Time) (State, error) {
	if !validElements(e) || t.IsZero() {
		return State{}, ErrInvalidElements
	}
	elapsed := t.Sub(e.Epoch)
	if !e.Epoch.Add(elapsed).Equal(t) {
		return State{}, fmt.Errorf("%w: propagation interval exceeds time.Duration range", ErrInvalidElements)
	}
	day, fraction := propagationEpochParts(e.Epoch)
	radians := math.Pi / 180
	motionFactor := 1440 / (2 * math.Pi)
	var s sgp4.Record
	s.Init(sgp4.Elements{
		EpochDays:                (day - 2433281.5) + fraction,
		BStar:                    e.BStar,
		Eccentricity:             e.Eccentricity,
		InclinationRad:           e.Inclination * radians,
		NodeRad:                  e.RightAscension * radians,
		ArgumentRad:              e.ArgumentOfPerigee * radians,
		MeanAnomalyRad:           e.MeanAnomaly * radians,
		MeanMotionRadPerMin:      e.MeanMotion / motionFactor,
		MeanMotionDotRadPerMin2:  e.MeanMotionDot / (motionFactor * 1440),
		MeanMotionDDotRadPerMin3: e.MeanMotionDDot / (motionFactor * 1440 * 1440),
	})
	minutes := float64(elapsed.Nanoseconds()) / 6e10
	r, v, code := s.Propagate(minutes)
	if err := propagationError(code); err != nil {
		return State{}, err
	}
	for i := range r {
		if math.IsNaN(r[i]) || math.IsInf(r[i], 0) || math.IsNaN(v[i]) || math.IsInf(v[i], 0) {
			return State{}, ErrInvalidElements
		}
	}
	return State{Time: t, PositionKM: r, VelocityKMS: v}, nil
}

func propagationEpochParts(epoch time.Time) (dayJD, fraction float64) {
	epoch = epoch.UTC()
	seconds := epoch.Unix()
	days := seconds / 86400
	if seconds%86400 < 0 {
		days--
	}
	h, m, s := epoch.Clock()
	nanos := (int64(h)*3600+int64(m)*60+int64(s))*1e9 + int64(epoch.Nanosecond())
	return float64(days) + 2440587.5, float64(nanos) / 86400e9
}

func propagationError(code int) error {
	switch code {
	case 0:
		return nil
	case 1:
		return ErrMeanEccentricity
	case 2:
		return ErrMeanMotion
	case 3:
		return ErrPerturbedEccentricity
	case 4:
		return ErrSemiLatusRectum
	case 6:
		return ErrDecayed
	default:
		return ErrInvalidElements
	}
}
