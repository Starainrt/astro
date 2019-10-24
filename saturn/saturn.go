package saturn

import (
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/planet"
)

/*
  土星视黄经
  jde: 世界时UTC
*/
func SeeLo(jde float64) float64 {
	return basic.SaturnSeeLo(basic.TD2UT(jde, true))
}

/*
  土星视黄纬
  jde: 世界时UTC
*/
func SeeBo(jde float64) float64 {
	return basic.SaturnSeeBo(basic.TD2UT(jde, true))
}

/*
  土星视赤经
  jde: 世界时UTC
*/
func SeeRa(jde float64) float64 {
	return basic.SaturnSeeRa(basic.TD2UT(jde, true))
}

/*
  土星视赤纬
  jde: 世界时UTC
*/
func SeeDec(jde float64) float64 {
	return basic.SaturnSeeDec(basic.TD2UT(jde, true))
}

/*
  土星视赤经赤纬
  jde: 世界时UTC
*/
func SeeRaDec(jde float64) (float64, float64) {
	return basic.SaturnSeeRaDec(basic.TD2UT(jde, true))
}

/*
  土星视星等
  jde: 世界时UTC
*/
func SeeMag(jde float64) float64 {
	return basic.SaturnMag(basic.TD2UT(jde, true))
}

/*
  与地球距离（天文单位）
  jde: 世界时UTC
*/
func EarthAway(jde float64) float64 {
	return basic.EarthSaturnAway(basic.TD2UT(jde, true))
}

/*
  与太阳距离（天文单位）
  jde: 世界时UTC
*/
func SunAway(jde float64) float64 {
	return planet.WherePlanet(1, 2, basic.TD2UT(jde, true))
}
