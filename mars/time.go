package mars

import (
	"math"
	"time"

	"github.com/starainrt/astro/basic"
)

func marsTimeValue(value float64, err error) float64 {
	if err != nil {
		return math.NaN()
	}
	return value
}

// SolDate 火星 Sol 日期，输入无效时为 NaN / Mars Sol Date, NaN for invalid input.
func (module *Mars) SolDate(date time.Time) float64 {
	return marsTimeValue(basic.MarsSolDate(phaseJD(date)))
}

// CoordinatedTime 火星协调时，小时 [0,24)，输入无效时为 NaN / Mars coordinated time in hours [0,24), NaN for invalid input.
func (module *Mars) CoordinatedTime(date time.Time) float64 {
	return marsTimeValue(basic.MarsCoordinatedTime(phaseJD(date)))
}

// LocalMeanSolarTime 火星约定地方平太阳时，东经为正，输入无效时为 NaN / conventional Mars local mean solar time for east longitude, NaN for invalid input.
func (module *Mars) LocalMeanSolarTime(date time.Time, eastLongitude float64) float64 {
	return marsTimeValue(basic.MarsLocalMeanSolarTime(phaseJD(date), eastLongitude))
}

// LocalTrueSolarTime 火星地方真太阳时，采用通用站心视太阳方向，输入无效时为 NaN / Mars local true solar time from the generic apparent solar direction, NaN for invalid input.
func (module *Mars) LocalTrueSolarTime(date time.Time, eastLongitude float64) float64 {
	return marsTimeValue(basic.MarsLocalTrueSolarTimeWithInstance(phaseJD(date), eastLongitude, module.inst))
}

// SolarLongitude 火星太阳黄经 Ls，输入无效时为 NaN / Mars solar longitude Ls, NaN for invalid input.
func (module *Mars) SolarLongitude(date time.Time) float64 {
	return marsTimeValue(basic.MarsSolarLongitudeWithInstance(phaseJD(date), module.inst))
}
