package earth

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// EarthEccentricity 地球偏心率
// 返回date对应UTC时间的地球偏心率
func EarthEccentricity(date time.Time) float64 {
	jde := basic.Date2JDE(date.UTC())
	return basic.Earthe(basic.TD2UT(jde, true))
}
