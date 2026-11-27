package basic

import "math"

func marsSolDate(jdTT float64) (float64, error) {
	if !isFiniteFloat(jdTT) {
		return 0, ErrInvalidObservationInput
	}
	value := (jdTT-2451549.5)/1.0274912517 + 44796.0 - 0.0009626
	if !isFiniteFloat(value) {
		return 0, ErrInvalidObservationInput
	}
	return value, nil
}

func marsCoordinatedTime(jdTT float64) (float64, error) {
	sol, err := MarsSolDate(jdTT)
	if err != nil {
		return 0, err
	}
	return surfaceHours(24 * (sol - math.Floor(sol))), nil
}

func marsLocalMeanSolarTime(jdTT, eastLongitude float64) (float64, error) {
	if !isFiniteFloat(eastLongitude) {
		return 0, ErrInvalidObservationInput
	}
	coordinated, err := MarsCoordinatedTime(jdTT)
	if err != nil {
		return 0, err
	}
	return surfaceHours(coordinated + surfaceHours(eastLongitude/15)), nil
}

// MarsLocalTrueSolarTimeWithInstance 火星地方真太阳时，东经为正，时刻为 TT 儒略日 / Mars local true solar time for east longitude at a TT Julian day.
func MarsLocalTrueSolarTimeWithInstance(jdTT, eastLongitude float64, inst *Instance) (float64, error) {
	local, err := SurfaceLocalTimeAtWithInstance(BodyObserver{Body: SkyMars, EastLongitude: eastLongitude}, jdTT, inst)
	if err != nil {
		return 0, err
	}
	return local.LocalTrueSolarTime, nil
}

// MarsSolarLongitudeWithInstance 火星太阳黄经 Ls，时刻为 TT 儒略日 / Mars solar longitude Ls at a TT Julian day.
func MarsSolarLongitudeWithInstance(jdTT float64, inst *Instance) (float64, error) {
	local, err := SurfaceLocalTimeAtWithInstance(BodyObserver{Body: SkyMars}, jdTT, inst)
	if err != nil {
		return 0, err
	}
	return local.SolarLongitude, nil
}
