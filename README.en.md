# Astro

**English | [中文](README.md)**

[![Go Reference](https://pkg.go.dev/badge/github.com/starainrt/astro.svg)](https://pkg.go.dev/github.com/starainrt/astro)

An astronomy library I have used for years for my own interest in astronomy and calendars.

> ### ⚠️ Time-scale change: UTC switches to leap hours on 2027-05-20
>
> On 15 October 2026 the 28th General Conference on Weights and Measures (CGPM) adopted [Resolution 5, "On the technical actions needed to ensure the continuity of UTC"](https://www.bipm.org/en/cgpm-2026/resolution-5) (`Resolution-CGPM-28-5`, DOI `10.59161/CGPM2026RES5E`): **continuous UTC becomes effective on 20 May 2027** and the maximum value of `|UT1−UTC|` is widened from 0.9 seconds to **3600 seconds (1 hour)** — a leap hour in place of a leap second.
>
> [IERS Bulletin C 72 (2026-07-06)](https://datacenter.iers.org/data/latestVersion/bulletinC.txt) confirms that **no leap second will be introduced at the end of December 2026**; the leap-second mechanism ends on 2027-05-20, so the possible insertion point at the end of June 2027 will not carry one either.
>
> From this release the library keeps **1972–2026 on the definitive leap-second table** (those historical values are unchanged) and changes the default policy beyond the built-in window (which ends 2026-10-01) from `TimeScaleLeapSecond` to **`TimeScaleLeapHour`**; past the window TT−UTC stays at 69.184 seconds. The numeric values of the `SetTimeScaleFuturePolicy` constants changed accordingly (`TimeScaleLeapHour` is now the zero value), so callers passing raw integers must update. See [Time scales](doc/manual/en/timescale.md).

The library ships two ephemerides, selectable in one process or used side by side:

1. **Semianalytic model** (`ModelSemianalytic`, default): the argument structure and initial amplitudes of VSOP2013/TOP2013 (planets) and ELP/MPP02 (Moon), refitted to the DE ephemerides — DE441 for the planetary long range, DE440 for the modern segment and DE441 for the Moon — and folded into a single compact series table. The planetary table is about 452 KiB and the lunar table about 42 KiB, with no external ephemeris files. Its geocentric apparent direction error (RMS) in the modern range is about 0.007″ (Sun) and 0.076″ (Moon).
2. **Legacy model** (`ModelLegacy`): the VSOP87/ELP2000-82 series with added long-term and per-body corrections. Switch to it to reproduce old values; its geocentric apparent direction error (RMS) in the modern range is about 0.04″ (Sun) and 0.75″ (Moon). See [Accuracy and Performance](doc/manual/en/accuracy.md).

External ephemerides can also be supplied through SPK or a custom engine.

## Install

```sh
go get github.com/starainrt/astro
```

## Highlights

- Civil and Chinese calendar conversion from 721 BCE to 3000 CE, solar terms, sexagenary dates, era names and historical calendars.
- Positions, rise/set, transit, distance, apparent diameter and physical ephemerides for the Sun, Moon and seven planets; lightweight Sun and Moon algorithms are also included.
- Solar and lunar eclipses, stellar and planetary lunar occultations, event searches, local contact times and global paths.
- Generic sky-event searches: ecliptic-longitude crossings, stationary points, phases, closest approach, rise/set at any altitude, perihelion/aphelion and heliacal visibility.
- Switchable ephemeris models, multiple instances used in parallel in one process, SPK multi-source stacking and optional perturbed propagation.
- SVG sky-event charts, GeoJSON geographic data and KML for viewing and time playback in Google Earth.
- A built-in catalog of 9100 stars with constellation lookup, coordinate corrections and rise/set calculations; CSV, Hipparcos, BSC5 and Stellarium external catalogs are also available.
- Celestial coordinate transformations, sidereal time, precession and nutation, refraction, astronomical formulas, sundial geometry and an optional lunar-limb terrain correction.

## Example

Calculate the sunrise time, lunar phase and Chinese calendar date for a day in Xi'an:

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

Longitude and latitude are in degrees, positive east and north; the example uses a height of 0 metres and `true` makes the rise/set calculation account for atmospheric refraction. Rise/set functions return an error during polar day, polar night or when no event occurs on that date.

Package-level functions use the process-wide ephemeris model (semianalytic in default builds). To pin a model for a calculation, use `astro.NewContext`; multiple contexts can be used in parallel in one process:

```go
func compareModels(date time.Time) {
	legacy, err := astro.NewContext(astro.ModelLegacy)
	if err != nil {
		panic(err)
	}
	modern, err := astro.NewContext(astro.ModelSemianalytic)
	if err != nil {
		panic(err)
	}

	fmt.Println(legacy.Moon().PhaseDesc(date), modern.Moon().PhaseDesc(date))

	riseLegacy, _ := legacy.Sun().RiseTime(date, 108.93, 34.27, 0, true)
	riseModern, _ := modern.Sun().RiseTime(date, 108.93, 34.27, 0, true)
	fmt.Println(riseLegacy.Format("15:04:05"), riseModern.Format("15:04:05"))
}
```

A context exposes per-module entry points such as `Sun()`, `Moon()`, `Mars()`, `Eclipse()` and `Phenomena()`, with the same signatures as their package-level counterparts. Models isolate the ephemeris only; ΔT, TT−UTC and civil time-scale policy remain process-wide settings. See [Explicit ephemeris instances](doc/manual/en/ephemeris.md) for custom engines, SPK and multi-source stacking.

## Package Overview

| Package | Purpose and documentation |
| --- | --- |
| `calendar` | [Calendar conversion, solar terms, sexagenary dates and historical calendars](doc/manual/en/calendar.md) |
| `sun` / `moon` / `lite/sun` / `lite/moon` | [Sun and Moon positions, rise/set, lunar phases and physical quantities](doc/manual/en/sun-moon.md) |
| `mercury` / `venus` / `mars` / `jupiter` / `saturn` / `uranus` / `neptune` / `earth` | [Planetary positions, special events, Jupiter's satellites and Saturn's rings](doc/manual/en/planets.md) |
| `eclipse` / `eclipse/svg` | [Solar and lunar eclipse and transit searches, local visibility, paths and charts](doc/manual/en/eclipse.md) |
| `moon` / `moon/svg` / `moon/limb` | [Stellar and planetary lunar occultations, contact times, bands and the optional lunar-limb terrain correction](doc/manual/en/occultation.md) |
| `phenomena` | [Ecliptic-longitude crossings, stationary points, phases, closest approach, heliacal visibility and rise/set at any altitude for any two bodies](doc/manual/en/phenomena.md) |
| `geojson` / `kml` | [Event maps, GeoJSON and KML](doc/manual/en/map-geojson.md) |
| `star` / `star/catalog` | [Star database, constellations, observing quantities and external catalogs](doc/manual/en/star.md#external-catalogs) |
| `coord` | [Coordinate conversion, sidereal time, precession, nutation and refraction](doc/manual/en/coord.md) |
| `orbit` | [Asteroid and comet two-body or optional perturbed propagation, visual binaries](doc/manual/en/orbit.md) |
| `sundial` | [Apparent solar time and sundial geometry](doc/manual/en/sundial.md) |
| `formula` | [Radiation, magnitudes, telescopes, atmospheric refraction, extinction and Earth-orientation tools](doc/manual/en/formula.md) |
| `astro` / `basic` / `planet/spk` | [Ephemeris models and instances, multi-source ephemerides and custom engines](doc/manual/en/ephemeris.md) |
| `astro` (root) | [UTC, UT1, TT conversion and ΔT models](doc/manual/en/timescale.md) |

`basic` contains the underlying algorithms, `planet` holds the built-in models and series, `planet/spk` reads external kernels, and `tools` provides numerical helpers. See [Accuracy and performance](doc/manual/en/accuracy.md) for model composition, measured errors and performance; full signatures are also available in the [Go API documentation](https://pkg.go.dev/github.com/starainrt/astro).

## Time Scale Conventions

Most observing APIs accept a civil instant as `time.Time` and handle UTC, UT1 and TT conversions internally. Before 1972 the library treats civil time as UT1, and as UTC thereafter; orbital epochs and some low-level APIs have separate time-scale requirements.

Angles default to degrees, apparent diameters to arcseconds, sidereal time to hours, and distances to AU or km as specified by the API. SVG and GeoJSON time labels default to UTC, with an explicit UT1 option. See [Time scales](doc/manual/en/timescale.md) for the exact conventions.

Beyond the monthly ΔT table, DUT1 and TT−UTC default to **leap-hour** extrapolation (`TimeScaleLeapHour`, matching the continuous UTC effective 2027-05-20); `SetTimeScaleFuturePolicy` switches between `TimeScaleLeapHour` (default), `TimeScaleLeapSecond`, `TimeScaleAssumeUT1Tracking`, `TimeScaleFreezeUTCOffset` and `TimeScaleUT1Civil`.

## Observer Height Convention

Observer height is ellipsoidal height in metres. If only orthometric height `H` is available, convert it with the local geoid undulation `N`: `height = H + N`. See [Observer height](doc/manual/en/coord.md#observer-height).

## Scope And Accuracy

Accuracy varies with model, epoch, body and calculated quantity. The default model has changed from VSOP87/ELP2000-82 to the semianalytic tables: signatures are unchanged but values may differ. `lite/sun` and `lite/moon` use independent approximation chains and do not follow model switches.

See [Accuracy and performance](doc/manual/en/accuracy.md) for model composition, error conventions and measured results. Professional occultation prediction, spacecraft navigation and similar uses also need ephemerides, time scales and physical models of matching accuracy; loading an SPK alone does not make every high-level result as accurate as the kernel.

## License

The code and documentation are licensed under [Apache-2.0](LICENSE). The third-party data embedded in this repository is **not** covered by Apache-2.0 and follows the terms of its own source; see [NOTICE](NOTICE) for the item-by-item statement:

- **SELENE (KAGUYA) LALT lunar limb profile** (`moon/limb/kaguya_lalt.bin`): ISAS/JAXA, distributed under the ISAS Data Policy, which states that its rules are compatible with CC BY 4.0.
- **This project's own refits of VSOP2013/TOP2013 and ELP/MPP02** (`planet/internal/semianalytic/`): the underlying theories are IMCCE / SYRTE. This project can neither grant nor restrict rights in those theories; Apache-2.0 applies only to this project's own code and compilation.
- **JPL DE440/DE441 and Horizons data** (`*/testdata/`, `planet/spk/testdata/`): NASA/JPL.
- **Astrometry in the built-in star catalogue** (`basic/star_catalog.dat`): derived from the Yale Bright Star Catalogue (VizieR [V/50](https://cdsarc.cds.unistra.fr/viz-bin/cat/V/50)).
- **Natural Earth 1:50m physical land** (`internal/svgmap/land_*.path`): public domain.
