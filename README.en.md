# Astro

**English | [中文](README.md)**

[![Go Reference](https://pkg.go.dev/badge/github.com/starainrt/astro.svg)](https://pkg.go.dev/github.com/starainrt/astro)

An astronomy library I have used for years for my own interest in astronomy and calendars.

Based on Jean Meeus's *Astronomical Algorithms*. Solar and planetary calculations use built-in VSOP87 terms; lunar calculations use ELP2000/82-style truncated series. No external ephemeris files are required.

Intended for calendar calculations, amateur observation and studying astronomical algorithms.

## Install

```sh
go get github.com/starainrt/astro
```

## Highlights

- Civil and Chinese calendar conversion from 721 BCE to 3000 CE, solar terms, sexagenary dates, era names and historical calendars.
- Positions, rise/set, transit, distances, angular diameters and physical ephemerides for the Sun, Moon and seven planets; lightweight Sun and Moon algorithms are also available.
- Solar and lunar eclipses, stellar and planetary lunar occultations, local contacts and global paths.
- SVG charts, GeoJSON data and KML for viewing and time playback in Google Earth.
- A built-in catalog of 9,100 stars, with constellation lookup, coordinate corrections and rise/set calculations.
- Coordinate transformations, sidereal time, precession, nutation, refraction, small-body orbits, sundials and astronomical formulas.

## Example

Calculate sunrise in Xi'an, the lunar phase and the Chinese calendar date:

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro/calendar"
	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/sun"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2026, 2, 17, 12, 0, 0, 0, cst)

	rise, err := sun.RiseTime(date, 108.93, 34.27, 0, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Sunrise:", rise.Format("15:04:05"))
	fmt.Println("Lunar phase:", moon.PhaseDesc(date))

	day, err := calendar.SolarToLunar(date)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Chinese calendar:", day.Lunar().MonthDay())
}
```

Longitude and latitude are in degrees, positive east and north. The example uses a height of 0 metres and enables refraction with `true`. Rise/set functions return errors for polar conditions or a missing event on that date.

Calendar and phase descriptions in this example are returned in Chinese.

## Package Overview

| Package | Purpose and documentation |
| --- | --- |
| `calendar` | [Calendar conversion, solar terms, sexagenary dates and historical calendars](doc/manual/en/calendar.md) |
| `sun` / `moon` / `lite/sun` / `lite/moon` | [Sun and Moon positions, rise/set, phases and physical quantities](doc/manual/en/sun-moon.md) |
| `mercury` / `venus` / `mars` / `jupiter` / `saturn` / `uranus` / `neptune` / `earth` | [Planetary positions, events, Jupiter's satellites and Saturn's rings](doc/manual/en/planets.md) |
| `eclipse` / `eclipse/svg` | [Eclipse searches, local visibility, paths and charts](doc/manual/en/eclipse.md) |
| `moon` / `moon/svg` | [Stellar and planetary lunar occultations, contacts and bands](doc/manual/en/occultation.md) |
| `geojson` / `kml` | [Event maps, GeoJSON and KML](doc/manual/en/map-geojson.md) |
| `star` | [Star catalog, constellations and observing quantities](doc/manual/en/star.md) |
| `coord` | [Coordinates, sidereal time, precession, nutation and refraction](doc/manual/en/coord.md) |
| `orbit` | [Asteroid and comet two-body orbits, visual binaries](doc/manual/en/orbit.md) |
| `sundial` | [Apparent solar time and sundial geometry](doc/manual/en/sundial.md) |
| `formula` | [Radiation, magnitudes, synodic periods and telescope formulas](doc/manual/en/formula.md) |
| `astro` (root) | [UTC, UT1, TT and ΔT models](doc/manual/en/timescale.md) |

`basic` contains the underlying algorithms, `planet` holds analytical series, and `tools` provides numerical helpers. Most applications can use the packages above.

Full signatures are also available in the [Go API documentation](https://pkg.go.dev/github.com/starainrt/astro).

## Time Scale Conventions

Most observing APIs accept a civil instant as `time.Time` and handle UTC, UT1 and TT conversions internally. Before 1972 the library treats civil time as UT1, and as UTC thereafter. Orbital epochs and some low-level APIs have separate scale requirements.

Angles default to degrees, angular diameters to arcseconds, sidereal time to hours, and distances to AU or km as specified by the API. SVG and GeoJSON time labels default to UTC, with explicit UT1 options.

See [Time scales](doc/manual/en/timescale.md).

DUT1 and TT−UTC default to extrapolating the current leap-second rule beyond the observed window (through September 2026); `SetTimeScaleFuturePolicy` switches the TT−UTC convention to one of four alternatives: a frozen offset, continued UT1 tracking, the leap-hour rule, or a civil scale permanently identical to UT1.

## Observer Height Convention

Observer height is ellipsoidal height in metres. Convert orthometric height `H` using the local geoid undulation `N`: `height = H + N`. See [Observer height](doc/manual/en/coord.md#observer-height).

## Scope And Accuracy

The library uses analytical models and truncated series. Accuracy varies with date, body and calculation. Professional occultation predictions and spacecraft navigation require more precise ephemerides and physical models.

`lite/sun` and `lite/moon` suit resource-constrained applications. Some main APIs also provide truncated variants with an `N` suffix.

Model ranges, measured differences and benchmarks are documented under [Accuracy and performance](doc/manual/en/accuracy.md).
