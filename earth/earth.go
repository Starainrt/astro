package earth

import (
	"github.com/starainrt/astro/basic"
)

/*
地球偏心率
jde, utc世界时
*/

func EarthEccentricity(jde float64) float64 {
	return basic.Earthe(basic.TD2UT(jde, true))
}
