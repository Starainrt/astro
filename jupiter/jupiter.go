package jupiter

import (
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/planet"
)

/*
  木星视黄经
  jde: 世界时UTC
*/
func SeeLo(jde float64) float64 {
	return basic.JupiterSeeLo(basic.TD2UT(jde, true))
}

/*
  木星视黄纬
  jde: 世界时UTC
*/
func SeeBo(jde float64) float64 {
	return basic.JupiterSeeBo(basic.TD2UT(jde, true))
}

/*
  木星视赤经
  jde: 世界时UTC
*/
func SeeRa(jde float64) float64 {
	return basic.JupiterSeeRa(basic.TD2UT(jde, true))
}

/*
  木星视赤纬
  jde: 世界时UTC
*/
func SeeDec(jde float64) float64 {
	return basic.JupiterSeeDec(basic.TD2UT(jde, true))
}

/*
  木星视赤经赤纬
  jde: 世界时UTC
*/
func SeeRaDec(jde float64) (float64, float64) {
	return basic.JupiterSeeRaDec(basic.TD2UT(jde, true))
}

/*
  木星视星等
  jde: 世界时UTC
*/
func SeeMag(jde float64) float64 {
	return basic.JupiterMag(basic.TD2UT(jde, true))
}

/*
  与地球距离（天文单位）
  jde: 世界时UTC
*/
func EarthAway(jde float64) float64 {
	return basic.EarthJupiterAway(basic.TD2UT(jde, true))
}

/*
  与太阳距离（天文单位）
  jde: 世界时UTC
*/
func SunAway(jde float64) float64 {
	return planet.WherePlanet(1, 2, basic.TD2UT(jde, true))
}
