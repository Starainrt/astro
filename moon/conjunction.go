package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// ConjunctionPlanet 月球合月目标行星 / target planet for Moon-planet conjunction.
type ConjunctionPlanet string

const (
	ConjunctionMercury ConjunctionPlanet = "mercury"
	ConjunctionVenus   ConjunctionPlanet = "venus"
	ConjunctionMars    ConjunctionPlanet = "mars"
	ConjunctionJupiter ConjunctionPlanet = "jupiter"
	ConjunctionSaturn  ConjunctionPlanet = "saturn"
	ConjunctionUranus  ConjunctionPlanet = "uranus"
	ConjunctionNeptune ConjunctionPlanet = "neptune"
)

func conjunctionPlanetToBasic(planet ConjunctionPlanet) basic.MoonPlanetConjunctionPlanet {
	switch planet {
	case ConjunctionMercury:
		return basic.MoonPlanetConjunctionMercury
	case ConjunctionVenus:
		return basic.MoonPlanetConjunctionVenus
	case ConjunctionMars:
		return basic.MoonPlanetConjunctionMars
	case ConjunctionJupiter:
		return basic.MoonPlanetConjunctionJupiter
	case ConjunctionSaturn:
		return basic.MoonPlanetConjunctionSaturn
	case ConjunctionUranus:
		return basic.MoonPlanetConjunctionUranus
	case ConjunctionNeptune:
		return basic.MoonPlanetConjunctionNeptune
	default:
		return 0
	}
}

func validConjunctionPlanet(planet ConjunctionPlanet) bool {
	return conjunctionPlanetToBasic(planet) != 0
}

// LastConjunctionWithPlanet 上一次行星合月（赤经合） / previous Moon-planet conjunction.
func (module *Moon) LastConjunctionWithPlanet(date time.Time, planet ConjunctionPlanet) time.Time {
	if !validConjunctionPlanet(planet) {
		return time.Time{}
	}
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.LastMoonPlanetConjunctionWithInstance(jde, conjunctionPlanetToBasic(planet), module.inst), date.Location(), false)
}

// NextConjunctionWithPlanet 下一次行星合月（赤经合） / next Moon-planet conjunction.
func (module *Moon) NextConjunctionWithPlanet(date time.Time, planet ConjunctionPlanet) time.Time {
	if !validConjunctionPlanet(planet) {
		return time.Time{}
	}
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.NextMoonPlanetConjunctionWithInstance(jde, conjunctionPlanetToBasic(planet), module.inst), date.Location(), false)
}

// ClosestConjunctionWithPlanet 最近一次行星合月（赤经合） / closest Moon-planet conjunction.
func (module *Moon) ClosestConjunctionWithPlanet(date time.Time, planet ConjunctionPlanet) time.Time {
	if !validConjunctionPlanet(planet) {
		return time.Time{}
	}
	jde := basic.UTC2TT(basic.Date2JD(date.UTC()))
	return basic.JD2DateByZone(basic.ClosestMoonPlanetConjunctionWithInstance(jde, conjunctionPlanetToBasic(planet), module.inst), date.Location(), false)
}
