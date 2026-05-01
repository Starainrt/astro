package earth

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// EarthEccentricity 地球轨道偏心率 / orbital eccentricity of Earth.
//
// 返回 date 对应绝对时刻的地球轨道偏心率，无量纲。
// Returns Earth's orbital eccentricity at the instant represented by date; the value is dimensionless.
func EarthEccentricity(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.Earthe(basic.TD2UT(jde, true))
}
