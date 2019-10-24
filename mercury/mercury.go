package mercury

import (
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/planet"
)

/*
  水星视黄经
  jde: 世界时UTC
*/
func SeeLo(jde float64) float64 {
	return basic.MercurySeeLo(basic.TD2UT(jde, true))
}

/*
  水星视黄纬
  jde: 世界时UTC
*/
func SeeBo(jde float64) float64 {
	return basic.MercurySeeBo(basic.TD2UT(jde, true))
}

/*
  水星视赤经
  jde: 世界时UTC
*/
func SeeRa(jde float64) float64 {
	return basic.MercurySeeRa(basic.TD2UT(jde, true))
}

/*
  水星视赤纬
  jde: 世界时UTC
*/
func SeeDec(jde float64) float64 {
	return basic.MercurySeeDec(basic.TD2UT(jde, true))
}

/*
  水星视赤经赤纬
  jde: 世界时UTC
*/
func SeeRaDec(jde float64) (float64, float64) {
	return basic.MercurySeeRaDec(basic.TD2UT(jde, true))
}

/*
  水星视星等
  jde: 世界时UTC
*/
func SeeMag(jde float64) float64 {
	return basic.MercuryMag(basic.TD2UT(jde, true))
}

/*
  与地球距离（天文单位）
  jde: 世界时UTC
*/
func EarthAway(jde float64) float64 {
	return basic.EarthMercuryAway(basic.TD2UT(jde, true))
}

/*
  与太阳距离（天文单位）
  jde: 世界时UTC
*/
func SunAway(jde float64) float64 {
	return planet.WherePlanet(1, 2, basic.TD2UT(jde, true))
}
