package basic

import "math"

const (
	riseSetWeatherMaxPressureHPa  = 1200.0
	riseSetWeatherMinTemperatureC = -100.0
	riseSetWeatherMaxTemperatureC = 80.0
)

// validRiseSetWeather 气压与气温须落在地表可能出现的量程内。
func validRiseSetWeather(pressureHPa, temperatureC float64) bool {
	return pressureHPa > 0 && pressureHPa <= riseSetWeatherMaxPressureHPa &&
		temperatureC >= riseSetWeatherMinTemperatureC && temperatureC <= riseSetWeatherMaxTemperatureC
}

type riseSetWeatherResidual func(jd, longitude, latitude, timeZone, zenithShift, height, pressureHPa, temperatureC float64, n int, inst *Instance) float64

// riseSetWithWeather 以标准气象结果为种子，在给定气压气温下重新求解同一民用日的升落；种子失效时按分钟扫描。
func riseSetWithWeather(seed float64, seedErr error, julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64,
	isRise bool, n int, inst *Instance, residualAt riseSetWeatherResidual) (float64, error) {
	if !isFiniteFloat(julianDay) || !isFiniteFloat(longitude) || !isFiniteFloat(latitude) || !isFiniteFloat(timeZone) ||
		!isFiniteFloat(height) || !validRiseSetWeather(pressureHPa, temperatureC) {
		return 0, ErrInvalidObservationInput
	}
	civilDayStart := math.Floor(julianDay) + 0.5
	residual := func(jd float64) float64 {
		return residualAt(jd, longitude, latitude, timeZone, 1, height, pressureHPa, temperatureC, n, inst)
	}
	if seedErr == nil && isFiniteFloat(seed) {
		const derivativeStep = 0.000005
		slope := math.NaN()
		refined, ok := eventNewtonRefine(seed, 1e-7, func(jd float64) float64 {
			slope = (residual(jd+derivativeStep) - residual(jd-derivativeStep)) / (2 * derivativeStep)
			return residual(jd) / slope
		})
		if ok && math.Abs(refined-seed) < 0.02 && eventRiseSetCandidateValid(refined, civilDayStart, slope, isRise) {
			return refined, nil
		}
	}
	return eventDirectionalRiseSetSearch(civilDayStart, isRise, nil, residual)
}

// GetSunRiseTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算日出，口径同 GetSunRiseTimeWithInstance 开启大气修正 / sunrise with explicit pressure and temperature.
func GetSunRiseTimeWithAtmosphere(julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64, inst *Instance) (float64, error) {
	seed, err := GetSunRiseTimeWithInstance(julianDay, longitude, latitude, timeZone, 1, height, inst)
	return riseSetWithWeather(seed, err, julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC, true, -1, inst, sunRiseSetResidualWeather)
}

// GetSunSetTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算日落 / sunset with explicit pressure and temperature.
func GetSunSetTimeWithAtmosphere(julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64, inst *Instance) (float64, error) {
	seed, err := GetSunSetTimeWithInstance(julianDay, longitude, latitude, timeZone, 1, height, inst)
	return riseSetWithWeather(seed, err, julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC, false, -1, inst, sunRiseSetResidualWeather)
}

// GetMoonRiseTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算月出 / moonrise with explicit pressure and temperature.
func GetMoonRiseTimeWithAtmosphere(julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64, inst *Instance) (float64, error) {
	seed, err := GetMoonRiseTimeWithInstance(julianDay, longitude, latitude, timeZone, 1, height, inst)
	return riseSetWithWeather(seed, err, julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC, true, -1, inst, moonRiseSetResidualWeather)
}

// GetMoonSetTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算月落 / moonset with explicit pressure and temperature.
func GetMoonSetTimeWithAtmosphere(julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64, inst *Instance) (float64, error) {
	seed, err := GetMoonSetTimeWithInstance(julianDay, longitude, latitude, timeZone, 1, height, inst)
	return riseSetWithWeather(seed, err, julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC, false, -1, inst, moonRiseSetResidualWeather)
}
