# Astro

**English | [中文](README.md)**

[![Go Reference](https://pkg.go.dev/badge/github.com/starainrt/astro.svg)](https://pkg.go.dev/github.com/starainrt/astro)

An astronomy library I have used for years for my own interest in astronomy and calendars.

## Install

```sh
go get github.com/starainrt/astro
```

## Highlights

- Civil and Chinese calendar conversion from 721 BCE to 3000 CE, solar terms, sexagenary cycles, era names, and ancient calendars.
- Positions, rise/set, transit, distance, apparent diameter, and physical ephemerides for the Sun, Moon, and seven planets; lightweight Sun and Moon algorithms are also included.
- Solar and lunar eclipses, Mercury and Venus transits, stellar and planetary lunar occultations, local contacts, global paths, and optional lunar-limb terrain correction.
- Generic sky-event searches: ecliptic-longitude crossings, stationary points, phases, closest approaches, rise/set at any altitude, perihelion/aphelion, heliacal visibility, and planetary surface observations.
- Earth satellite TLE/OMM parsing, SGP4/SDP4 propagation, passes, solar/lunar transits, and ground coverage.
- Switchable built-in ephemerides, multiple isolated instances in one process, JPL SPK multi-source stacking, and optional perturbed orbit propagation.
- SVG sky-event charts (eclipses, occultations, Jupiter's satellites, Earth satellites), GeoJSON geographic data, and KML for viewing and time playback in Google Earth.
- Built-in catalog of 9100 stars with constellation lookup, coordinate corrections, and rise/set calculations; external CSV, Hipparcos, BSC5, and Stellarium catalogs are also supported.
- Celestial coordinate transformations, sidereal time, precession and nutation, refraction, astronomical formulas, and sundial geometry.

## Example

Calculate sunrise, sunset, lunar phase, and the Chinese calendar date for Xi'an on 2026-02-17:

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
	lon, lat := 108.93, 34.27

	rise, err := sun.RiseTime(date, lon, lat, 0, true)
	if err != nil {
		log.Fatal(err)
	}
	set, err := sun.SetTime(date, lon, lat, 0, true)
	if err != nil {
		log.Fatal(err)
	}
	day, err := calendar.SolarToLunar(date)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Xi'an 2026-02-17")
	fmt.Printf("Sunrise / Sunset: %s / %s\n", rise.Format("15:04:05"), set.Format("15:04:05"))
	fmt.Printf("Lunar phase: %s (%.1f%% illuminated)\n", moon.PhaseDesc(date), moon.Phase(date)*100)
	fmt.Printf("Chinese calendar: %d %s (%s year, %s day)\n",
		day.Lunar().LunarYear(), day.Lunar().MonthDay(),
		day.Lunar().GanZhiYear(), day.Lunar().GanZhiDay())
}
```

Output:

```text
Xi'an 2026-02-17
Sunrise / Sunset: 07:27:16 / 18:29:40
Lunar phase: 残月 (0.1% illuminated)
Chinese calendar: 2026 正月初一 (丙午 year, 壬戌 day)
```

Longitude and latitude are in degrees, positive east and north; observer height is ellipsoidal height in metres, and `true` includes atmospheric refraction and apparent semidiameter. Rise/set functions return an error during polar day, polar night, or when no event occurs on that date.

Package-level functions use the process-wide ephemeris model (semianalytic in default builds). To pin a model or run multiple ephemerides in parallel in one process, use `astro.NewContext`:

```go
legacy, _ := astro.NewContext(astro.ModelLegacy)
modern, _ := astro.NewContext(astro.ModelSemianalytic)

riseL, _ := legacy.Sun().RiseTime(date, 108.93, 34.27, 0, true)
riseM, _ := modern.Sun().RiseTime(date, 108.93, 34.27, 0, true)
fmt.Printf("Sunrise: semianalytic=%s  legacy=%s\n", riseM.Format("15:04:05"), riseL.Format("15:04:05"))
// Sunrise: semianalytic=07:27:16  legacy=07:27:16
```

A context exposes per-module accessors (`Sun()`, `Moon()`, `Mars()`, `Eclipse()`, `Phenomena()`, `Satellite()`, etc.) with the same signatures as their package-level counterparts. See [Ephemeris Instances](doc/manual/en/ephemeris.md) for custom engines, SPK kernels, and multi-source stacking.

## Package Overview

| Package | Purpose and Documentation |
| --- | --- |
| `calendar` | [Calendar conversion, solar terms, sexagenary cycles, and ancient calendars](doc/manual/en/calendar.md) |
| `sun` / `moon` / `lite/sun` / `lite/moon` | [Sun and Moon positions, rise/set, lunar phases, and physical quantities](doc/manual/en/sun-moon.md) |
| `mercury` / `venus` / `mars` / `jupiter` / `jupiter/svg` / `saturn` / `uranus` / `neptune` / `earth` | [Planetary positions, events, Jupiter's satellites, and Saturn's rings](doc/manual/en/planets.md) / [Jupiter Satellite SVG](doc/manual/en/jupiter-svg.md) |
| `eclipse` / `eclipse/svg` | [Solar and lunar eclipses, planetary transits, local visibility, paths, and SVG charts](doc/manual/en/eclipse.md) |
| `moon` / `moon/svg` / `moon/limb` | [Stellar and planetary lunar occultations, contacts, paths, and optional lunar-limb correction](doc/manual/en/occultation.md) |
| `phenomena` | [Longitude crossings, stations, phases, approaches, heliacal visibility, altitude crossings, and planetary surface observations](doc/manual/en/phenomena.md) |
| `satellite` | [TLE / OMM, SGP4 propagation, passes, transits, and ground coverage](doc/manual/en/satellite.md) |
| `satellite/svg` | [Satellite transit maps, local disc charts, and coverage SVG](doc/manual/en/satellite-svg.md) |
| `geojson` / `kml` | [Event maps, GeoJSON, and KML](doc/manual/en/map-geojson.md) |
| `star` / `star/catalog` | [Star catalog, constellations, observing quantities, and external catalogs](doc/manual/en/star.md#external-catalogs) |
| `coord` | [Coordinate transforms, sidereal time, precession, nutation, and refraction](doc/manual/en/coord.md) |
| `orbit` | [Asteroid and comet two-body or perturbed propagation, visual binaries](doc/manual/en/orbit.md) |
| `sundial` | [Apparent solar time and planar sundial geometry](doc/manual/en/sundial.md) |
| `formula` | [Radiation, magnitudes, telescopes, atmospheric refraction, extinction, and Earth-orientation tools](doc/manual/en/formula.md) |
| `astro` / `basic` / `planet/spk` | [Ephemeris models, isolated instances, multi-source stacking, and custom engines](doc/manual/en/ephemeris.md) |
| `astro` (root) | [UTC, UT1, TT conversion and ΔT models](doc/manual/en/timescale.md) |

`basic` contains the underlying algorithms, `planet` holds the built-in models and series, `planet/spk` reads external kernels, and `tools` provides numerical helpers. See [Accuracy and Performance](doc/manual/en/accuracy.md) for model composition, measured errors, and performance; full signatures are available on [pkg.go.dev](https://pkg.go.dev/github.com/starainrt/astro).

## Ephemeris and Time Scale Conventions

> **Time-scale change: UTC switches to leap hours on 2027-05-20**
>
> Under CGPM [Resolution 5](https://www.bipm.org/en/cgpm-2026/resolution-5) (`Resolution-CGPM-28-5`), continuous UTC takes effect on 2027-05-20 and the `|UT1−UTC|` limit is widened from 0.9 s to 3600 s (leap hours replacing leap seconds). [IERS Bulletin C 72](https://datacenter.iers.org/data/latestVersion/bulletinC.txt) also confirms no leap second at the end of 2026. Since v0.4.2, the library keeps the 1972–2026 leap-second table unchanged and defaults future extrapolation beyond the built-in window to `TimeScaleLeapHour` (zero value). Callers passing raw integers to `SetTimeScaleFuturePolicy` must update to named constants. See [Time Scales](doc/manual/en/timescale.md).

- **Built-in ephemerides**: Default is **`ModelSemianalytic`** (VSOP2013/TOP2013/ELP-MPP02 refitted to DE440/DE441, requiring no external files; modern geocentric RMS `~0.007″` Sun, `~0.076″` Moon). Switch to **`ModelLegacy`** (VSOP87/ELP2000-82) to reproduce older outputs. `lite/sun` and `lite/moon` use independent approximation chains unaffected by model switches.
- **Civil time & time scales**: Most observing APIs take civil `time.Time` (treated as UT1 before 1972-01-01 and UTC thereafter). SVG and GeoJSON default to UTC labels with an explicit UT1 option.
- **Observer height**: Ellipsoidal height in metres. Convert orthometric elevation `H` with local geoid undulation `N` via `height = H + N`. See [Observer Height](doc/manual/en/coord.md#observer-height).
- **Scope**: Built-in ephemerides target calendars, observing predictions, education, and charts. Precision occultation timing or spacecraft navigation requires matching SPK kernels, measured ΔT/EOP, and terrain models. See [Accuracy and Performance](doc/manual/en/accuracy.md).

## Low-Memory Build (Experimental)

`-tags astro_lowmem` targets RAM-constrained platforms (such as TinyGo on ESP32-C3): ephemeris tables remain in read-only storage and are evaluated on demand with bit-identical results and unchanged APIs, at the cost of slower evaluation and a binary ~0.3 MB larger. Desktop, server, and WASM builds do not need this tag.

Querying the built-in 9100-star catalog (`StarDataBy*`), enabling lunar-limb terrain correction (`moon/limb`), or switching to `ModelLegacy` (~240 KB heap decode on first use) still requires substantial heap memory and is unsuitable for very low-RAM targets.

## License

The code and documentation are licensed under [Apache-2.0](LICENSE). Embedded third-party data are **not** covered by Apache-2.0 and remain subject to their original terms; see [NOTICE](NOTICE) for details:

- **SELENE (KAGUYA) LALT lunar limb profile** (`moon/limb/kaguya_lalt.bin`): ISAS/JAXA, under the ISAS Data Policy (compatible with CC BY 4.0).
- **Refits of VSOP2013/TOP2013 and ELP/MPP02** (`planet/internal/semianalytic/`): underlying theories are IMCCE / SYRTE; Apache-2.0 applies only to this project's code and compilation.
- **JPL DE440/DE441 and Horizons test data** (`*/testdata/`, `planet/spk/testdata/`): NASA/JPL.
- **Built-in star catalog astrometry** (`basic/star_catalog.dat`): derived from Yale Bright Star Catalogue (VizieR [V/50](https://cdsarc.cds.unistra.fr/viz-bin/cat/V/50)).
- **Natural Earth 1:50m physical land** (`internal/svgmap/land_*.path`): public domain.
