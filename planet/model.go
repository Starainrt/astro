package planet

import (
	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/planet/internal/semianalytic"
)

func init() {
	ephem.RegisterModelEngine(ephem.Semianalytic, func() ephem.Engine { return semianalytic.Ephemeris{} })
}

func semianalyticActive() bool { return ephem.Active() == ephem.Semianalytic }

// semiPlanet 与 WherePlanetN 同语义，调用方已做越界检查。
func semiPlanet(xt, zn int, jde float64, n int) float64 { return semianalytic.Planet(xt, zn, jde, n) }

func semiMoon(coord int, jde float64, n int) float64 { return semianalytic.Moon(coord, jde, n) }
