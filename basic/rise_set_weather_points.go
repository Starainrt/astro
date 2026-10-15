package basic

import "math"

type riseSetHeightFunc func(localJD float64) float64

// pointRiseSetWithWeather 点光源（行星、恒星）带气象升落：沿用标准结果所在的事件窗口，残差为几何高度 + 地平俯角 + 指定气象折射。
// 行星与恒星的标准升落按最近中天取事件，可能落在输入日期的前一天，因此这里不做民用日校验。
func pointRiseSetWithWeather(seed float64, seedErr error, latitude, height, pressureHPa, temperatureC float64,
	isRise bool, heightAt riseSetHeightFunc) (float64, error) {
	if seedErr != nil {
		return 0, seedErr
	}
	if !isFiniteFloat(seed) || !isFiniteFloat(latitude) || !isFiniteFloat(height) || !validRiseSetWeather(pressureHPa, temperatureC) {
		return 0, ErrInvalidObservationInput
	}
	residual := func(jd float64) float64 {
		altitude := heightAt(jd)
		return altitude + HeightDegreeByLat(height, latitude) + RefractionFromTrueAltitude(altitude, pressureHPa, temperatureC)
	}
	const derivativeStep = 0.000005
	slope := math.NaN()
	refined, ok := eventNewtonRefine(seed, 1e-7, func(jd float64) float64 {
		slope = (residual(jd+derivativeStep) - residual(jd-derivativeStep)) / (2 * derivativeStep)
		return residual(jd) / slope
	})
	if !ok || math.Abs(refined-seed) > 0.02 || !isFiniteFloat(slope) || (isRise && slope <= 0) || (!isRise && slope >= 0) {
		return 0, ErrInvalidObservationInput
	}
	return refined, nil
}

// MercuryRiseTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算水星升起 / mercury rise with explicit pressure and temperature.
func MercuryRiseTimeWithAtmosphere(julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64, inst *Instance) (float64, error) {
	seed, err := MercuryRiseTimeWithInstance(julianDay, longitude, latitude, timeZone, 1, height, inst)
	return pointRiseSetWithWeather(seed, err, latitude, height, pressureHPa, temperatureC, true,
		func(jd float64) float64 { return MercuryHeightWithInstance(jd, longitude, latitude, timeZone, inst) })
}

// MercurySetTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算水星落下 / mercury set with explicit pressure and temperature.
func MercurySetTimeWithAtmosphere(julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64, inst *Instance) (float64, error) {
	seed, err := MercurySetTimeWithInstance(julianDay, longitude, latitude, timeZone, 1, height, inst)
	return pointRiseSetWithWeather(seed, err, latitude, height, pressureHPa, temperatureC, false,
		func(jd float64) float64 { return MercuryHeightWithInstance(jd, longitude, latitude, timeZone, inst) })
}

// VenusRiseTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算金星升起 / venus rise with explicit pressure and temperature.
func VenusRiseTimeWithAtmosphere(julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64, inst *Instance) (float64, error) {
	seed, err := VenusRiseTimeWithInstance(julianDay, longitude, latitude, timeZone, 1, height, inst)
	return pointRiseSetWithWeather(seed, err, latitude, height, pressureHPa, temperatureC, true,
		func(jd float64) float64 { return VenusHeightWithInstance(jd, longitude, latitude, timeZone, inst) })
}

// VenusSetTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算金星落下 / venus set with explicit pressure and temperature.
func VenusSetTimeWithAtmosphere(julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64, inst *Instance) (float64, error) {
	seed, err := VenusSetTimeWithInstance(julianDay, longitude, latitude, timeZone, 1, height, inst)
	return pointRiseSetWithWeather(seed, err, latitude, height, pressureHPa, temperatureC, false,
		func(jd float64) float64 { return VenusHeightWithInstance(jd, longitude, latitude, timeZone, inst) })
}

// MarsRiseTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算火星升起 / mars rise with explicit pressure and temperature.
func MarsRiseTimeWithAtmosphere(julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64, inst *Instance) (float64, error) {
	seed, err := MarsRiseTimeWithInstance(julianDay, longitude, latitude, timeZone, 1, height, inst)
	return pointRiseSetWithWeather(seed, err, latitude, height, pressureHPa, temperatureC, true,
		func(jd float64) float64 { return MarsHeightWithInstance(jd, longitude, latitude, timeZone, inst) })
}

// MarsSetTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算火星落下 / mars set with explicit pressure and temperature.
func MarsSetTimeWithAtmosphere(julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64, inst *Instance) (float64, error) {
	seed, err := MarsSetTimeWithInstance(julianDay, longitude, latitude, timeZone, 1, height, inst)
	return pointRiseSetWithWeather(seed, err, latitude, height, pressureHPa, temperatureC, false,
		func(jd float64) float64 { return MarsHeightWithInstance(jd, longitude, latitude, timeZone, inst) })
}

// JupiterRiseTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算木星升起 / jupiter rise with explicit pressure and temperature.
func JupiterRiseTimeWithAtmosphere(julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64, inst *Instance) (float64, error) {
	seed, err := JupiterRiseTimeWithInstance(julianDay, longitude, latitude, timeZone, 1, height, inst)
	return pointRiseSetWithWeather(seed, err, latitude, height, pressureHPa, temperatureC, true,
		func(jd float64) float64 { return JupiterHeightWithInstance(jd, longitude, latitude, timeZone, inst) })
}

// JupiterSetTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算木星落下 / jupiter set with explicit pressure and temperature.
func JupiterSetTimeWithAtmosphere(julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64, inst *Instance) (float64, error) {
	seed, err := JupiterSetTimeWithInstance(julianDay, longitude, latitude, timeZone, 1, height, inst)
	return pointRiseSetWithWeather(seed, err, latitude, height, pressureHPa, temperatureC, false,
		func(jd float64) float64 { return JupiterHeightWithInstance(jd, longitude, latitude, timeZone, inst) })
}

// SaturnRiseTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算土星升起 / saturn rise with explicit pressure and temperature.
func SaturnRiseTimeWithAtmosphere(julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64, inst *Instance) (float64, error) {
	seed, err := SaturnRiseTimeWithInstance(julianDay, longitude, latitude, timeZone, 1, height, inst)
	return pointRiseSetWithWeather(seed, err, latitude, height, pressureHPa, temperatureC, true,
		func(jd float64) float64 { return SaturnHeightWithInstance(jd, longitude, latitude, timeZone, inst) })
}

// SaturnSetTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算土星落下 / saturn set with explicit pressure and temperature.
func SaturnSetTimeWithAtmosphere(julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64, inst *Instance) (float64, error) {
	seed, err := SaturnSetTimeWithInstance(julianDay, longitude, latitude, timeZone, 1, height, inst)
	return pointRiseSetWithWeather(seed, err, latitude, height, pressureHPa, temperatureC, false,
		func(jd float64) float64 { return SaturnHeightWithInstance(jd, longitude, latitude, timeZone, inst) })
}

// UranusRiseTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算天王星升起 / uranus rise with explicit pressure and temperature.
func UranusRiseTimeWithAtmosphere(julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64, inst *Instance) (float64, error) {
	seed, err := UranusRiseTimeWithInstance(julianDay, longitude, latitude, timeZone, 1, height, inst)
	return pointRiseSetWithWeather(seed, err, latitude, height, pressureHPa, temperatureC, true,
		func(jd float64) float64 { return UranusHeightWithInstance(jd, longitude, latitude, timeZone, inst) })
}

// UranusSetTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算天王星落下 / uranus set with explicit pressure and temperature.
func UranusSetTimeWithAtmosphere(julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64, inst *Instance) (float64, error) {
	seed, err := UranusSetTimeWithInstance(julianDay, longitude, latitude, timeZone, 1, height, inst)
	return pointRiseSetWithWeather(seed, err, latitude, height, pressureHPa, temperatureC, false,
		func(jd float64) float64 { return UranusHeightWithInstance(jd, longitude, latitude, timeZone, inst) })
}

// NeptuneRiseTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算海王星升起 / neptune rise with explicit pressure and temperature.
func NeptuneRiseTimeWithAtmosphere(julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64, inst *Instance) (float64, error) {
	seed, err := NeptuneRiseTimeWithInstance(julianDay, longitude, latitude, timeZone, 1, height, inst)
	return pointRiseSetWithWeather(seed, err, latitude, height, pressureHPa, temperatureC, true,
		func(jd float64) float64 { return NeptuneHeightWithInstance(jd, longitude, latitude, timeZone, inst) })
}

// NeptuneSetTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算海王星落下 / neptune set with explicit pressure and temperature.
func NeptuneSetTimeWithAtmosphere(julianDay, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64, inst *Instance) (float64, error) {
	seed, err := NeptuneSetTimeWithInstance(julianDay, longitude, latitude, timeZone, 1, height, inst)
	return pointRiseSetWithWeather(seed, err, latitude, height, pressureHPa, temperatureC, false,
		func(jd float64) float64 { return NeptuneHeightWithInstance(jd, longitude, latitude, timeZone, inst) })
}

// StarRiseTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算恒星升起；ra/dec 为瞬时视赤经赤纬（度） / stellar rise with explicit pressure and temperature.
func StarRiseTimeWithAtmosphere(julianDay, ra, dec, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64) (float64, error) {
	seed, err := StarRiseTime(julianDay, ra, dec, longitude, latitude, height, timeZone, true)
	return pointRiseSetWithWeather(seed, err, latitude, height, pressureHPa, temperatureC, true,
		func(jd float64) float64 { return StarHeight(jd, ra, dec, longitude, latitude, timeZone) })
}

// StarSetTimeWithAtmosphere 按给定气压（hPa）与气温（°C）计算恒星落下；ra/dec 为瞬时视赤经赤纬（度） / stellar set with explicit pressure and temperature.
func StarSetTimeWithAtmosphere(julianDay, ra, dec, longitude, latitude, timeZone, height, pressureHPa, temperatureC float64) (float64, error) {
	seed, err := StarSetTime(julianDay, ra, dec, longitude, latitude, height, timeZone, true)
	return pointRiseSetWithWeather(seed, err, latitude, height, pressureHPa, temperatureC, false,
		func(jd float64) float64 { return StarHeight(jd, ra, dec, longitude, latitude, timeZone) })
}
