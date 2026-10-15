# Sun and Moon

[中文](../sun-moon.md) | [Back to README](../../../README.en.md)

`sun` and `moon` are the main chains; `lite/sun` and `lite/moon` are independent approximation implementations.

The numerical output blocks in this manual illustrate call shapes and come from runs with the default semianalytic model. The differences from Legacy are small: lunar event times by about `1-10 s`, lunar distance by about `5 km`, phase by about `1e-6`, solar and lunar angles by about `1e-5°` and rise/set instants by milliseconds, with `lite/*` unaffected. Use an explicit `Context` / `Instance` to reproduce a fixed model; see [Accuracy and performance](accuracy.md) for the quantified errors of both models.

Unless stated otherwise, angles are in degrees, apparent diameters and semidiameters are in arcseconds, `sun.EarthDistance` is in AU and `moon.EarthDistance` is in kilometers. Observing APIs generally use civil instants.

Apparent-solar-time results instead represent local solar readings; see [Time scales](timescale.md).

This manual covers the Sun and the Moon themselves: position, rise/set and culmination, topocentric quantities, phases and syzygies, perigee/apogee and nodes, maximum declinations, libration, apparent size, and physical ephemeris.

Eclipse geometry lives in [Eclipses and Transits](eclipse.md#solar-and-lunar-eclipse-charts), occultations in [Lunar Occultation Charts](occultation.md#lunar-occultation-charts), and general topocentric and refraction conversions in [Coordinate Tools](coord.md).

## Contents

- [Sunrise, sunset and lunar phase](#sunrise-sunset-and-lunar-phase)
- [API Reference](#api-reference)
  - [sun](#sun)
  - [moon](#moon)
  - [lite/sun](#litesun)
  - [lite/moon](#litemoon)
  - [The `...N` family](#the-n-family)
- [Usage examples](#usage-examples)
  - [Today's sunrise, sunset and twilight](#todays-sunrise-sunset-and-twilight)
  - [Moonrise, moonset and the Moon's altitude now](#moonrise-moonset-and-the-moons-altitude-now)
  - [Lunar phase and the next new / full moon](#lunar-phase-and-the-next-new--full-moon)
  - [Apparent size, Earth-Moon distance and libration](#apparent-size-earth-moon-distance-and-libration)
  - [Geocentric vs topocentric (getting the Moon position right)](#geocentric-vs-topocentric-getting-the-moon-position-right)
  - [Main chain vs lite](#main-chain-vs-lite)
- [Observing-angle semantics](#observing-angle-semantics)
- [Combined example: rise, set, and position](#combined-example-rise-set-and-position)
  - [Sunrise/sunset and moonrise/moonset](#sunrisesunset-and-moonrisemoonset)
  - [Sun and Moon position](#sun-and-moon-position)
- [The Sun](#the-sun)
  - [Position](#position)
  - [Rise, set and culmination](#rise-set-and-culmination)
  - [Topocentric quantities and parallactic angle](#topocentric-quantities-and-parallactic-angle)
  - [Apparent solar time and equation of time](#apparent-solar-time-and-equation-of-time)
  - [Physical ephemeris and apparent size](#physical-ephemeris-and-apparent-size)
  - [Earth orbit extrema](#earth-orbit-extrema)
- [The Moon](#the-moon)
  - [Position](#position-1)
  - [Rise, set and culmination](#rise-set-and-culmination-1)
  - [Lunar phases](#lunar-phases)
  - [Perigee and apogee](#perigee-and-apogee)
  - [Nodes](#nodes)
  - [Maximum declinations](#maximum-declinations)
  - [Libration and bright-limb position angle](#libration-and-bright-limb-position-angle)
  - [Apparent size and Earth-Moon distance](#apparent-size-and-earth-moon-distance)
- [Lite chains](#lite-chains)
  - [lite/sun](#litesun-1)
  - [lite/moon](#litemoon-1)
  - [Differences from the main chain and error levels](#differences-from-the-main-chain-and-error-levels)
- [Parameter and result conventions](#parameter-and-result-conventions)
  - [Units and angle conventions](#units-and-angle-conventions)
  - [Time scale](#time-scale)
  - [Height and aero](#height-and-aero)
  - [Zero values and out-of-range](#zero-values-and-out-of-range)
  - [Accuracy and scope](#accuracy-and-scope)

## Sunrise, sunset and lunar phase

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/sun"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2020, 1, 1, 8, 8, 8, 0, cst)
	lon, lat, height := 108.93, 34.27, 0.0
	rise, err := sun.RiseTime(date, lon, lat, height, true) // sunrise time
	if err != nil {
		log.Fatal(err)
	}
	set, err := sun.SetTime(date, lon, lat, height, true) // sunset time
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(rise.Format(time.RFC3339), set.Format(time.RFC3339))
	fmt.Println(moon.Phase(date), moon.PhaseDesc(date)) // lunar phase (illuminated fraction) and phase description
}
```

`aero=true` uses a rise/set criterion with refraction. Rise/set functions return an error for polar conditions or a missing event on that date. `moon.Phase` is the illuminated fraction, not the lunar age.

## API Reference

The tables below group the exported entry points of all four packages. Most evaluation entry points also have a `...N` variant, described together under [the `...N` family](#the-n-family).

Every `time.Time` parameter is an absolute instant, and `lon`/`lat` are east-positive and north-positive.

### sun

| Name | Purpose | Unit and convention |
| --- | --- | --- |
| `TrueLo` / `ApparentLo` | True / apparent solar longitude | degrees, geocentric |
| `TrueBo` | True solar latitude | degrees, geocentric; no separate apparent latitude |
| `GeometricLo` / `MidFunc` | Geometric solar longitude / equation of center | degrees |
| `ApparentRa` / `ApparentDec` / `ApparentRaDec` | Apparent right ascension / declination | degrees, geocentric |
| `EclipticObliquity` | Obliquity of the ecliptic | degrees; the second argument adds nutation in obliquity when `true` |
| `EclipticNutation` / `EclipticNutation1980` | Nutation in longitude | degrees, IAU 2000B / IAU 1980 |
| `AxialtiltNutation` / `AxialtiltNutation1980` | Nutation in obliquity | degrees, IAU 2000B / IAU 1980 |
| `RiseTime` / `SetTime` | Sunrise / sunset | `(time.Time, error)`; `aero` and `height` are described under Parameter and result conventions |
| `DownTime` | Sunset alias | deprecated; calls `SetTime` internally |
| `CulminationTime` | Upper culmination | `time.Time` |
| `MorningTwilight` / `EveningTwilight` | Morning / evening twilight | `(time.Time, error)`; the angle is usually -6 / -12 / -18 degrees |
| `Altitude` / `Zenith` / `Azimuth` / `HourAngle` | Geometric altitude / zenith distance / azimuth / hour angle | degrees; `Altitude` / `Zenith` / `Azimuth` are topocentric, while `HourAngle` is geocentric apparent (without topocentric parallax) |
| `ApparentAltitude` / `ApparentZenith` | Apparent altitude / apparent zenith distance | degrees; requires pressure in hPa and temperature in degrees Celsius |
| `ParallacticAngle` | Parallactic angle (zenith direction angle) | degrees, signed |
| `ApparentSolarTime` | Apparent solar time | `time.Time`, with the zone derived from longitude |
| `EquationTime` | Equation of time | hours |
| `Diameter` / `Semidiameter` | Apparent diameter / semidiameter | arcseconds |
| `EarthDistance` | Earth-Sun distance | AU |
| `Physical` | Solar disk physical quantities | returns `PhysicalInfo`, fields in degrees |

### moon

| Name | Purpose | Unit and convention |
| --- | --- | --- |
| `TrueLo` / `TrueBo` / `ApparentLo` | Geocentric true longitude / true latitude / apparent longitude | degrees |
| `TrueRa` / `TrueDec` / `TrueRaDec` | Geocentric true equatorial coordinates | degrees |
| `GeocentricApparentRa` / `GeocentricApparentDec` / `GeocentricApparentRaDec` | Geocentric apparent equatorial coordinates | degrees |
| `ApparentRa` / `ApparentDec` / `ApparentRaDec` | Topocentric apparent equatorial coordinates | degrees; requires observer longitude and latitude |
| `Altitude` / `Zenith` / `Azimuth` / `HourAngle` | Topocentric altitude / zenith distance / azimuth / hour angle | degrees |
| `ApparentAltitude` / `ApparentZenith` | Apparent altitude / apparent zenith distance | degrees; requires pressure in hPa and temperature in degrees Celsius |
| `ParallacticAngle` | Parallactic angle | degrees, signed, explicitly depends on observer longitude and latitude |
| `RiseTime` / `SetTime` | Moonrise / moonset | `(time.Time, error)` |
| `DownTime` | Moonset alias | deprecated; calls `SetTime` internally |
| `CulminationTime` | Upper culmination | `time.Time`; requires longitude and latitude |
| `Phase` / `PhaseDesc` | Illuminated fraction / Chinese textual phase | fraction `[0,1]` / string |
| `SunMoonLoDiff` | Apparent Moon-Sun longitude difference | degrees, `[0,360)` |
| `ShuoYue` / `ShangXianYue` / `WangYue` / `XiaXianYue` | New / first-quarter / full / last-quarter moon solved near a decimal-year anchor | `time.Time`, UTC |
| `NewMoon` / `FullMoon` / `FirstQuarter` / `LastQuarter` | English aliases for the four phases above | `time.Time`, UTC |
| `Next*` / `Last*` / `Closest*` | Next / previous / closest phase and maximum-declination events | `time.Time`; results keep the input time zone |
| `NextConjunctionWithPlanet` / `LastConjunctionWithPlanet` / `ClosestConjunctionWithPlanet` | Moon-planet conjunction (in right ascension) | `time.Time`; the target is a `ConjunctionPlanet` constant |
| `PerigeesInMonth` / `ApogeesInMonth` | All perigees / apogees in a Gregorian month | `[]ApsisInfo`, distance in km |
| `MaximumNorthDeclinationsInMonth` / `MaximumSouthDeclinationsInMonth` | All maximum northern / southern declination events in a month | `[]MaximumDeclinationInfo`, declination in degrees |
| `AscendingNode` / `DescendingNode` | Ascending / descending node longitude | degrees |
| `Physical` / `TopocentricPhysical` | Geocentric / topocentric libration and rotation-axis position angle | returns `PhysicalInfo`, fields in degrees |
| `BrightLimbPositionAngle` / `TopocentricBrightLimbPositionAngle` | Geocentric / topocentric bright-limb position angle | degrees |
| `Diameter` / `Semidiameter` | Apparent diameter / semidiameter | arcseconds |
| `EarthDistance` | Earth-Moon distance | kilometers |

The `moon` package also exposes occultation APIs. Their parameters, results and examples are documented under [Lunar occultations](occultation.md).

| Group | Exports |
| --- | --- |
| Events and paths | `FindStarOccultations`, `FindPlanetOccultations`, `FindBestStarOccultations`, `FindBestPlanetOccultations`, `FindStarOccultationPaths`, `FindPlanetOccultationPaths` |
| Instant footprints and disk geometry | `StarOccultationFootprintAt`, `PlanetOccultationFootprintsAt`, `StarOccultationDiagram`, `PlanetOccultationDiagram` |
| UT1 label conversion | `StarOccultationInfoInUT1`, `StarOccultationPathInUT1`, `PlanetOccultationInfoInUT1`, `PlanetOccultationPathInUT1` |
| Stellar results | `StarOccultationInfo`, `StarOccultationPath`, `StarOccultationInstant` |
| Planetary results | `PlanetOccultationInfo`, `PlanetOccultationPath`, `PlanetOccultationInstant`, `PlanetOccultationFootprint` |
| Path data | `OccultationFootprint`, `OccultationPathPoint`, `OccultationGreatestTimeContour`, `OccultationRiseSetCurve` |
| Search and path options | `OccultationSearchOptions`, `OccultationPathOptions`, `OccultationPathAlgorithm` |
| Disk diagram data | `StarOccultationDiagramFrame`, `StarOccultationDiagramOptions`, `StarOccultationDiagramResult`, `PlanetOccultationDiagramFrame`, `PlanetOccultationDiagramOptions`, `PlanetOccultationDiagramResult` |
| Targets and coordinates | `StarData`, `StarCoordinate`, `StarCoordinateFromStarData`, `Observer`, `CoordinateFrame`, `CoordinateFrameICRS`, `CoordinateFrameJ2000`, `CoordinateFrameApparentOfDate` |
| Event types | `OccultationPlanet`, `OccultationType`, `OccultationTotal`, `OccultationPartial`, `OccultationGrazing` |
| Path algorithms | `OccultationPathAlgorithmOptimized`, `OccultationPathAlgorithmExact` |
| Rise/set phases | `RiseSetPhase`, `RiseSetDirection`, `RiseSetPhaseStart`, `RiseSetPhaseGreatest`, `RiseSetPhaseEnd`, `RiseSetDirectionRise`, `RiseSetDirectionSet` |
| Errors | `ErrInvalidOccultationInput`, `ErrOccultationPathSamplingLimit` |

Planet targets use constants `OccultationMercury` through `OccultationNeptune`.

### lite/sun

| Name | Purpose | Unit and convention |
| --- | --- | --- |
| `TrueLo` / `ApparentLo` | Lightweight true / apparent longitude | degrees, geocentric |
| `TrueRa` / `TrueDec` / `TrueRaDec` | Lightweight true equatorial coordinates | degrees, geocentric |
| `ApparentRa` / `ApparentDec` / `ApparentRaDec` | Lightweight apparent equatorial coordinates | degrees, geocentric |
| `Distance` | Lightweight Earth-Sun distance | AU |
| `HourAngle` / `Azimuth` / `Altitude` / `Zenith` | Lightweight hour angle / azimuth / altitude / zenith distance | degrees; all use the geocentric apparent position, ignoring solar parallax of about 8.8 arcseconds |
| `RiseTime` / `SetTime` | Lightweight sunrise / sunset | `(time.Time, error)` |
| `ERR_SUN_NEVER_RISE` / `ERR_SUN_NEVER_SET` | Polar night / polar day | error values |

### lite/moon

| Name | Purpose | Unit and convention |
| --- | --- | --- |
| `TrueLo` / `TrueBo` | Lightweight geocentric true longitude / latitude | degrees |
| `TrueRa` / `TrueDec` / `TrueRaDec` | Lightweight geocentric true equatorial coordinates | degrees |
| `ApparentRa` / `ApparentDec` / `ApparentRaDec` | Lightweight topocentric apparent equatorial coordinates | degrees; requires observer longitude and latitude |
| `HourAngle` / `Azimuth` / `Altitude` / `Zenith` | Lightweight hour angle / azimuth / altitude / zenith distance | degrees, topocentric |
| `SunMoonLoDiff` / `Phase` / `PhaseAge` | Lightweight Moon-Sun longitude difference / illuminated fraction / lunar age | degrees / `[0,1]` / days |
| `RiseTime` / `SetTime` | Lightweight moonrise / moonset | `(time.Time, error)` |
| `ERR_MOON_NEVER_RISE` / `ERR_MOON_NEVER_SET` / `ERR_NOT_TODAY` | Polar night / polar day / event not on the queried date | error values |

### The `...N` family

Most evaluation entry points of `sun` and `moon` have a `...N` variant: the name is the base name plus `N`, one extra `n int` parameter is appended, and the return shape is unchanged.

`n < 0` selects the full series and matches the non-`N` version. Default builds use to the semianalytic model, where all `n >= 0` select one fixed fast table: increasing `n` does not add terms. Legacy shortens its series according to `n`. Neither changes units or time zones; measured errors are in [Accuracy and performance](accuracy.md).

- sun: `TrueLoN`, `TrueBoN`, `AltitudeN`, `ZenithN`, `AzimuthN`, `HourAngleN`, `ParallacticAngleN`, `ApparentAltitudeN`, `ApparentZenithN`, `DiameterN`, `SemidiameterN`, `PhysicalN`, `RiseTimeN`, `SetTimeN`, `DownTimeN`, `CulminationTimeN`, `MorningTwilightN`, `EveningTwilightN`, `ApparentSolarTimeN`
- moon: `TrueLoN`, `TrueBoN`, `AscendingNodeN`, `DescendingNodeN`, `DiameterN`, `SemidiameterN`, `PhysicalN`, `TopocentricPhysicalN`, `BrightLimbPositionAngleN`, `TopocentricBrightLimbPositionAngleN`

The topocentric equatorial coordinates (`ApparentRa` / `ApparentDec` / `ApparentRaDec`), the lunar rise/set entry points (`RiseTime` / `SetTime`), and the phase family have no `N` variant and are not controlled by the truncation switch.

## Usage examples

### Today's sunrise, sunset and twilight

```go
fmt.Println(sun.MorningTwilight(date, lon, lat, -6)) // civil dawn
fmt.Println(sun.RiseTime(date, lon, lat, height, true))
fmt.Println(sun.SetTime(date, lon, lat, height, true))
fmt.Println(sun.EveningTwilight(date, lon, lat, -6)) // civil dusk
```

```text
2020-01-01 07:22:28.138198256 +0800 CST <nil>
2020-01-01 07:49:52.591398954 +0800 CST <nil>
2020-01-01 17:45:09.366609156 +0800 CST <nil>
2020-01-01 18:12:33.801986575 +0800 CST <nil>
```

Passing `-12` / `-18` instead selects nautical and astronomical twilight. With `aero = true` the rise/set solution uses the horizon corrected for refraction and apparent radius; the difference from the geometric horizon is described under [Rise, set and culmination](#rise-set-and-culmination).

### Moonrise, moonset and the Moon's altitude now

```go
rise, _ := moon.RiseTime(date, lon, lat, height, true)
set, _ := moon.SetTime(date, lon, lat, height, true)
fmt.Println(rise)
fmt.Println(set)
fmt.Println(moon.Altitude(date, lon, lat), moon.Azimuth(date, lon, lat))
```

```text
2020-01-01 11:52:50.042243599 +0800 CST
2020-01-01 23:26:49.498263895 +0800 CST
-45.349728852972675 67.63824603392399
```

Lunar rise and set are computed for the local civil day and the pair is not necessarily continuous, so the full cycle after `date` follows from the order of the rise and set instants; see [Sunrise/sunset and moonrise/moonset](#sunrisesunset-and-moonrisemoonset) for the details. A negative altitude means the Moon is below the horizon, so `-45.35` degrees here means it is not visible.

### Lunar phase and the next new / full moon

```go
fmt.Println(moon.Phase(date), moon.PhaseDesc(date)) // illuminated fraction and phase name
fmt.Println(moon.NextShuoYue(date))                 // next new moon
fmt.Println(moon.NextWangYue(date))                 // next full moon
```

```text
0.30004130960877884 上峨眉月
2020-01-25 05:41:58.271192908 +0800 CST
2020-01-11 03:21:17.159625291 +0800 CST
```

`Next*` / `Last*` / `Closest*` are the next / previous / closest search conventions; first and last quarter are `moon.NextShangXianYue` / `moon.NextXiaXianYue`. Results keep the input time zone; the full conventions for the four phases and the synodic month are under [Lunar phases](#lunar-phases).

### Apparent size, Earth-Moon distance and libration

```go
fmt.Println(moon.Diameter(date), moon.EarthDistance(date))
p := moon.Physical(date)
fmt.Println(p.LibrationLongitude, p.LibrationLatitude, p.PositionAngle)
```

```text
1774.6658461637385 404238.6096080479
0.7655535663486027 6.382898400777244 -23.672356410246774
```

`Diameter` is in arcseconds and `EarthDistance` in kilometers; `1774.67` arcseconds (about 29.6 arcminutes) corresponds to roughly 404,000 km near apogee, about 5% smaller than the mean apparent diameter.

`Physical` gives the geocentric libration; the topocentric counterpart `TopocentricPhysical` and the field conventions are under [Libration and bright-limb position angle](#libration-and-bright-limb-position-angle).

### Geocentric vs topocentric (getting the Moon position right)

```go
geoRa, geoDec := moon.GeocentricApparentRaDec(date) // geocentric apparent position
topRa, topDec := moon.ApparentRaDec(date, lon, lat) // topocentric apparent position
fmt.Printf("geocentric %.4f %.4f\n", geoRa, geoDec)
fmt.Printf("topocentric %.4f %.4f\n", topRa, topDec)
fmt.Printf("delta dRA=%.4f dDec=%.4f\n", topRa-geoRa, topDec-geoDec)
fmt.Println(sun.ApparentRaDec(date)) // the Sun only exposes geocentric coordinates
```

```text
geocentric 349.2322 -9.9506
topocentric 349.7343 -10.3485
delta dRA=0.5021 dDec=-0.3978
280.8950939694744 -23.05840775453492
```

The Moon is close, so geocentric and topocentric positions differ by half a degree (here `0.50` degrees in right ascension and `0.40` in declination), together more than one lunar apparent diameter; anything an observer sees must go through `moon.ApparentRaDec`. The Sun, 1 AU away, has negligible parallax and is exposed geocentrically only; convert with `coord.TopocentricEquatorial` when a topocentric position is needed. The conventions are in [Observing-angle semantics](#observing-angle-semantics).

### Main chain vs lite

```go
fmt.Println(sun.Altitude(date, lon, lat), litesun.Altitude(date, lon, lat)) // main / lightweight geometric altitude
fmt.Println(moon.Phase(date), litemoon.Phase(date))                         // illuminated fraction
fmt.Println(litemoon.PhaseAge(date))                                        // lightweight lunar age (days)
fmt.Println(litesun.RiseTime(date, lon, lat, height, true))                 // lightweight sunrise
```

```text
2.40091496867759 2.403576774819768
0.30004130960877884 0.2978124633132848
5.42608394367707
2020-01-01 07:49:51.69717729 +0800 CST <nil>
```

The lightweight chains mirror the main-chain shapes but not its accuracy: in the output above the main chain and `lite/*` differ by about `0.0027` degrees in altitude and `0.0022` in phase.

Historical comparisons against Legacy over 2026 at 8 sites give a mean absolute difference of `0.02 min` for `lite/sun` sunrise (P95 `0.04 min`, max `0.31 min`) and `0.28 min` for `lite/moon` moonrise (P95 `0.57 min`, max `1.44 min`); the maximum `lite/moon` `Phase()` difference is `0.00243`. These are model-to-model differences, not differences against an independent ephemeris.

The full comparison and the truncation errors are under [Differences from the main chain and error levels](#differences-from-the-main-chain-and-error-levels) and in [Lite chains](accuracy.md#lite-lightweight-chains).

## Observing-angle semantics

- `Altitude`: altitude angle; horizon is `0°`, zenith is `+90°`
- `Zenith`: zenith distance; zenith is `0°`, horizon is `90°`
- `Zenith` and `Altitude` are complements; the two add up to `90°`
- `Azimuth`: azimuth, measured from north (`0°`) toward east, in `[0°, 360°)`
- `HourAngle`: hour angle, `0°` at upper culmination and increasing westward (afternoon), normalized to `[0°, 360°)`
- `ParallacticAngle`: parallactic angle (zenith direction angle), signed, in degrees; the shared sign convention and topocentric geometry live in [Coordinate Tools](coord.md)
- `Altitude` / `Azimuth` are geometric center altitude and azimuth without refraction or semidiameter correction; `ApparentAltitude` / `ApparentZenith` add atmospheric refraction and need pressure (hPa) and temperature (degrees Celsius)
- Solar equatorial coordinates are geocentric; for the Moon, `TrueRaDec` and `GeocentricApparentRaDec` are geocentric while `ApparentRa` / `ApparentDec` / `ApparentRaDec` are topocentric and require observer longitude and latitude

## Combined example: rise, set, and position

The two complete examples below cover the most common entry points. Their shared setup is Xi'an (`108.93°E, 34.27°N`) and `2020-01-01 08:08:08 CST`.

The later snippets omit shared variables and keep only the statements relevant to their capability.

### Sunrise/sunset and moonrise/moonset

> ⚠️ Moon rise/set times are computed for the queried civil date, so the rise and set instants need not be continuous.
>
> For example, the Moon may set at 01:00 and rise again at noon, in which case the rise time is later than the set time; the evening moonset in that scenario corresponds to the next day's date.
>
> The full rise/set cycle follows from the order of the two instants: check whether the rise time falls after the set time to pick the correct subsequent instants.

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/sun"
)

func main() {
	// Xi'an, China. Longitude east and latitude north are positive; elevation is 0 m.
	var lon, lat, height float64 = 108.93, 34.27, 0
	cst := time.FixedZone("CST", 8*3600)
	// All "today" semantics are based on this local civil date.
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	// Civil morning twilight begins when the Sun is 6 degrees below the horizon.
	// Civil twilight is 6 degrees below the horizon, nautical 12, astronomical 18.
	fmt.Println(sun.MorningTwilight(date, lon, lat, -6))
	// Sunrise: dynamic standard refraction and instantaneous solar semidiameter, upper limb.
	fmt.Println(sun.RiseTime(date, lon, lat, height, true))
	// Upper culmination of the Sun in Xi'an.
	fmt.Println(sun.CulminationTime(date, lon))
	// Sunset: dynamic standard refraction and instantaneous solar semidiameter, upper limb.
	fmt.Println(sun.SetTime(date, lon, lat, height, true))
	// Civil evening twilight ends when the Sun is 6 degrees below the horizon.
	fmt.Println(sun.EveningTwilight(date, lon, lat, -6))

	// Moonrise: dynamic standard refraction and instantaneous lunar semidiameter, upper limb.
	fmt.Println(moon.RiseTime(date, lon, lat, height, true))
	// Upper culmination of the Moon in Xi'an.
	fmt.Println(moon.CulminationTime(date, lon, lat))
	// Moonset: dynamic standard refraction and instantaneous lunar semidiameter, upper limb.
	fmt.Println(moon.SetTime(date, lon, lat, height, true))
}
```

Output:

```text
2020-01-01 07:22:28.91963303 +0800 CST <nil>
2020-01-01 07:49:53.401680886 +0800 CST <nil>
2020-01-01 12:47:35.933721363 +0800 CST
2020-01-01 17:45:08.557346463 +0800 CST <nil>
2020-01-01 18:12:33.02157104 +0800 CST <nil>
2020-01-01 11:52:50.013020932 +0800 CST <nil>
2020-01-01 17:36:48.954946696 +0800 CST
2020-01-01 23:26:49.448562562 +0800 CST <nil>
```

### Sun and Moon position

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/star"
	"github.com/starainrt/astro/sun"
	"github.com/starainrt/astro/tools"
)

func main() {
	// Xi'an, China.
	var lon, lat float64 = 108.93, 34.27
	cst := time.FixedZone("CST", 8*3600)
	// Instant of observation.
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	// Apparent ecliptic longitude of the Sun, in degrees.
	fmt.Println(sun.ApparentLo(date))
	// True obliquity of the ecliptic at this instant.
	fmt.Println(sun.EclipticObliquity(date, true))
	// Apparent right ascension and declination of the Sun.
	ra, dec := sun.ApparentRaDec(date)
	fmt.Println("RA:", tools.Format(ra/15, 1), "Dec:", tools.Format(dec, 0))
	// English constellation containing the Sun.
	fmt.Println(star.ConstellationEN(ra, dec, date))
	// Solar azimuth, altitude, and zenith distance at Xi'an.
	fmt.Println("Azimuth:", sun.Azimuth(date, lon, lat), "Altitude:", sun.Altitude(date, lon, lat), "Zenith:", sun.Zenith(date, lon, lat))
	// Sun-Earth distance, in AU.
	fmt.Println(sun.EarthDistance(date))

	// Topocentric apparent right ascension and declination of the Moon.
	ra, dec = moon.ApparentRaDec(date, lon, lat)
	fmt.Println("RA:", tools.Format(ra/15, 1), "Dec:", tools.Format(dec, 0))
	// English constellation containing the Moon.
	fmt.Println(star.ConstellationEN(ra, dec, date))
	// Lunar azimuth, altitude, and zenith distance at Xi'an.
	fmt.Println("Azimuth:", moon.Azimuth(date, lon, lat), "Altitude:", moon.Altitude(date, lon, lat), "Zenith:", moon.Zenith(date, lon, lat))
	// Earth-Moon distance, in km.
	fmt.Println(moon.EarthDistance(date))
}
```

Output:

```text
280.0152652754265
23.4362178391013
RA: 18h43m34.82s Dec: -23°3′30.20″
Sagittarius
Azimuth: 120.19442333006145 Altitude: 2.398445319378493 Zenith: 87.6015546806215
0.9832929289388871
RA: 23h18m56.20s Dec: -10°20′54.64″
Aquarius
Azimuth: 67.63846015645981 Altitude: -45.349642390635466 Zenith: 135.34964239063547
404243.3922335782
```

## The Sun

These snippets omit the shared setup: `cst := time.FixedZone("CST", 8*3600)`, `date := time.Date(2026, 1, 1, 12, 0, 0, 0, cst)`, and `var lon, lat float64 = 108.93, 34.27` (Xi'an); `fmt`, `time`, `sun`, and `moon` are assumed to be imported.

### Position

Solar position entry points depend only on the absolute instant, not on the observer. True ecliptic latitude comes from `TrueBo`; there is no separate apparent latitude.

The second argument of `EclipticObliquity` decides whether nutation in obliquity is added.

```go
// True and apparent solar longitude, and true latitude, in degrees.
fmt.Println(sun.TrueLo(date), sun.ApparentLo(date), sun.TrueBo(date))
// Apparent right ascension and declination of the Sun.
ra, dec := sun.ApparentRaDec(date)
fmt.Println("RA:", ra, "Dec:", dec)
fmt.Println(sun.ApparentRa(date), sun.ApparentDec(date))
// Obliquity of the ecliptic, geometric longitude, and equation of center.
fmt.Println(sun.EclipticObliquity(date, true), sun.GeometricLo(date), sun.MidFunc(date))
```

Output:

```text
280.742671383543 280.7383965677222 0.00018280886212040676
RA: 281.6786097810291 Dec: -23.00369182413533
281.6786097810291 -23.00387403948847
23.438148552330773 280.83169623098 -0.08703499790144194
```

Every entry point in this family also has a `...N` variant. Here `n = 8` selects the semianalytic fast table (series truncation in Legacy); `n < 0` matches the non-`N` version exactly. The output below comes from the default semianalytic model:

```go
// n<0 selects the full series; n>=0 selects fast, or Legacy truncation.
fmt.Println(sun.TrueLo(date), sun.TrueLoN(date, 8))
fmt.Println(sun.Altitude(date, lon, lat), sun.AltitudeN(date, lon, lat, 8))
fmt.Println(sun.Diameter(date), sun.DiameterN(date, 8))
```

Output:

```text
280.742671383543 280.7439900413756
31.61569462953789 31.615524473491835
1950.9979407481142 1950.9994358395434
```

### Rise, set and culmination

`RiseTime` / `SetTime` anchor on the local civil day of the time zone carried by `date` and keep the same time zone in the result. `height` is the observer elevation interpreted as ellipsoidal (geodetic) height in meters.

With `aero = true`, the **upper limb** crossing is computed with dynamic standard refraction and the instantaneous semidiameter; with `aero = false`, only the geometric center crossing is tested.

Twilight entry points take the target altitude as a parameter: civil twilight `-6°`, nautical `-12°`, astronomical `-18°`, with `MorningTwilight` and `EveningTwilight` for the two sides.

```go
// Target altitudes for civil, nautical, and astronomical morning twilight.
for _, angle := range []float64{-6, -12, -18} {
	t, err := sun.MorningTwilight(date, lon, lat, angle)
	fmt.Println(angle, t.Format("15:04:05"), err)
}
// Upper culmination and sunrise; err is non-nil when no event exists.
fmt.Println(sun.CulminationTime(date, lon).Format("15:04:05"))
t, err := sun.RiseTime(date, lon, lat, 0, true)
fmt.Println(t.Format("15:04:05"), err)
```

Output:

```text
-6 07:22:47 <nil>
-12 06:51:31 <nil>
-18 06:21:00 <nil>
12:47:50
07:50:10 <nil>
```

### Topocentric quantities and parallactic angle

`Altitude` / `Zenith` / `Azimuth` / `HourAngle` follow the geometric chain without refraction; `ApparentAltitude` / `ApparentZenith` add atmospheric refraction and need pressure and temperature; `ParallacticAngle` is the signed parallactic angle. General topocentric and refraction conversions (including apparent altitude and topocentric equatorial coordinates) live in [Coordinate Tools](coord.md).

```go
fmt.Println(sun.Azimuth(date, lon, lat), sun.Altitude(date, lon, lat), sun.Zenith(date, lon, lat))
fmt.Println(sun.ApparentAltitude(date, lon, lat, 1010, 10), sun.ApparentZenith(date, lon, lat, 1010, 10))
// Hour angle and signed parallactic angle.
fmt.Println(sun.HourAngle(date, lon, lat), sun.ParallacticAngle(date, lon, lat))
```

Output:

```text
167.09774780715728 31.61569462953789 58.38430537046211
31.643010360459822 58.356989639540174
348.07820699607544 -11.564174033740159
```

### Apparent solar time and equation of time

`ApparentSolarTime` returns the apparent solar time at a longitude; the result uses a fixed-offset time zone derived from that longitude, not the zone passed by the caller.

`EquationTime` returns the equation of time at the same instant, in hours. `sundial.TrueSolarTime` uses the same convention, and `sundial` additionally provides local mean solar time and sundial geometry; see [Sundial and Apparent Solar Time](sundial.md).

```go
// Apparent solar time; the result time zone is derived from longitude.
fmt.Println(sun.ApparentSolarTime(date, lon).Format("2006-01-02 15:04:05 -0700"))
// Equation of time, in hours.
fmt.Println(sun.EquationTime(date))
// sundial.TrueSolarTime uses the same convention.
fmt.Println(sundial.TrueSolarTime(date, lon).Format("2006-01-02 15:04:05 -0700"))
```

Output:

```text
2026-01-01 11:12:18 +0715
-0.05674946079069686
2026-01-01 11:12:18 +0715
```

### Physical ephemeris and apparent size

`sun.Physical` returns `PhysicalInfo`: `P` is the position angle of the solar north pole, `B0` is the heliographic latitude of the disk center, and `L0` is the Carrington heliographic longitude of the disk center, all in degrees.

`Diameter` / `Semidiameter` give the apparent diameter and semidiameter in arcseconds; `EarthDistance` gives the Earth-Sun distance in AU.

```go
// Apparent diameter and semidiameter (arcseconds), and Earth-Sun distance (AU).
fmt.Println(sun.Diameter(date), sun.Semidiameter(date), sun.EarthDistance(date))
// Solar physical quantities P/B0/L0, in degrees.
p := sun.Physical(date)
fmt.Println(p.P, p.B0, p.L0)
```

Output:

```text
1950.9979407481142 975.4989703740571 0.9833237486528845
1.979086377118846 -3.0131029209723916 296.9595333604375
```

The Sun, Moon, and seven major planets share the same interface shapes, so they can be compared side by side:

```go
fmt.Println(sun.Diameter(date), sun.Semidiameter(date))
fmt.Println(sun.Physical(date))
fmt.Println(moon.Diameter(date), moon.Semidiameter(date))
fmt.Println(mars.Diameter(date), mars.Semidiameter(date))
```

### Earth orbit extrema

The extrema of the Earth-Sun distance come from the `earth` package, with UTC times and AU distances; for the orbital eccentricity at one instant, call `EarthEccentricity` directly:

```go
// Earth perihelion and aphelion in 2026; time is UTC, distance is AU.
peri := earth.Perihelion(2026)
aphe := earth.Aphelion(2026)
fmt.Printf("earth perihelion=%s distance=%.9fAU\n", peri.Time.Format(time.RFC3339), peri.Distance)
fmt.Printf("earth aphelion=%s distance=%.9fAU\n", aphe.Time.Format(time.RFC3339), aphe.Distance)
```

Output:

```text
earth perihelion=2026-01-03T17:15:35Z distance=0.983302050AU
earth aphelion=2026-07-06T17:31:24Z distance=1.016643936AU
```

```go
fmt.Printf("earth e=%.9f\n", earth.EarthEccentricity(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
```

## The Moon

These snippets reuse the shared setup of the solar section: `date` is still `2026-01-01 12:00:00 CST` and the observer is still in Xi'an.

### Position

Lunar equatorial coordinates come in three layers: `TrueRaDec` is the geocentric true place, `GeocentricApparentRaDec` is the geocentric apparent place, and `ApparentRaDec` is the topocentric apparent place.

On the ecliptic side only geocentric quantities exist: `TrueLo` true longitude, `TrueBo` true latitude, and `ApparentLo` apparent longitude.

```go
// Geocentric true equatorial coordinates.
fmt.Println(moon.TrueRaDec(date))
// Geocentric apparent equatorial coordinates.
fmt.Println(moon.GeocentricApparentRaDec(date))
// Topocentric apparent equatorial coordinates.
fmt.Println(moon.ApparentRaDec(date, lon, lat))
// True longitude, true latitude, and apparent longitude.
fmt.Println(moon.TrueLo(date), moon.TrueBo(date), moon.ApparentLo(date))
```

Output:

```text
66.66309709020791 26.830370234236764
66.66476688311091 26.830608930005567
67.0275694157259 25.982665403390747
69.21422925147913 5.060516750865828 69.21574418700706
```

### Rise, set and culmination

Lunar rise/set, like the solar one, anchors on the local civil day: `RiseTime` / `SetTime` return that day's event instants and `CulminationTime` returns that day's upper culmination.

Successive lunar rises and sets need not be continuous, and the full cycle after `date` follows from the order of the two instants. When an event falls outside the queried date, the rise/set entry points return `ERR_NOT_TODAY`.

```go
// Moonrise, upper culmination, and moonset on the local civil day.
rise, err := moon.RiseTime(date, lon, lat, 0, true)
fmt.Println(rise.Format("15:04:05"), err)
fmt.Println(moon.CulminationTime(date, lon, lat).Format("15:04:05"))
set, err := moon.SetTime(date, lon, lat, 0, true)
fmt.Println(set.Format("15:04:05"), err)
```

Output:

```text
16:17:11 <nil>
00:03:16
06:41:37 <nil>
```

### Lunar phases

`Phase` returns the illuminated fraction in `[0,1]`, `PhaseDesc` returns a Chinese phase name, and `SunMoonLoDiff` returns the apparent Moon-Sun longitude difference normalized to `[0,360)` (near `0°` at new moon and near `180°` at full moon).

`Next*` / `Last*` / `Closest*` are the three search conventions, and results keep the input time zone.

Each of the four phases has both a pinyin name and an English alias, for example `ShuoYue` / `NewMoon`, `WangYue` / `FullMoon`, `ShangXianYue` / `FirstQuarter`, and `XiaXianYue` / `LastQuarter`.

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/moon"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	// Instant of observation.
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	// Illuminated fraction of the lunar disk.
	fmt.Println(moon.Phase(date))
	// Chinese textual phase description.
	fmt.Println(moon.PhaseDesc(date))
	// Next new moon; moon.NextNewMoon(date) is the English alias.
	fmt.Println(moon.NextShuoYue(date))
	// Next first quarter; moon.NextFirstQuarter(date) is the English alias.
	fmt.Println(moon.NextShangXianYue(date))
	// Next full moon; moon.NextFullMoon(date) is the English alias.
	fmt.Println(moon.NextWangYue(date))
	// Next last quarter; moon.NextLastQuarter(date) is the English alias.
	fmt.Println(moon.NextXiaXianYue(date))
}
```

Output:

```text
0.3000398794307665 // about 30% of the lunar disk is illuminated
上峨眉月 // Chinese phase description
2020-01-25 05:41:59.860723614 +0800 CST // next new moon
2020-01-03 12:45:24.51480031 +0800 CST // next first quarter
2020-01-11 03:21:18.524373471 +0800 CST // next full moon
2020-01-17 20:58:24.686442017 +0800 CST // next last quarter
```

`Last*`, `Closest*`, and the decimal-year anchors `ShuoYue` / `FullMoon` return UTC. `ClosestConjunctionWithPlanet` finds the closest Moon-planet conjunction, with the target given by a `ConjunctionPlanet` constant:

```go
// Previous and closest new moon and full moon.
fmt.Println(moon.LastShuoYue(date), moon.ClosestShuoYue(date))
fmt.Println(moon.LastWangYue(date), moon.ClosestWangYue(date))
// First and last quarter.
fmt.Println(moon.LastFirstQuarter(date), moon.ClosestLastQuarter(date))
// Decimal-year anchors; the results are UTC.
fmt.Println(moon.ShuoYue(2025.5).Format(time.RFC3339), moon.FullMoon(2025.5).Format(time.RFC3339))
// Closest Moon-planet conjunction in right ascension.
fmt.Println(moon.ClosestConjunctionWithPlanet(date, moon.ConjunctionJupiter))
```

Output:

```text
2025-12-20 09:43:19.074603617 +0800 CST 2025-12-20 09:43:19.074603617 +0800 CST
2025-12-05 07:14:03.670351803 +0800 CST 2026-01-03 18:02:53.55531156 +0800 CST
2025-12-28 03:09:50.141303837 +0800 CST 2026-01-10 23:48:22.03346461 +0800 CST
2025-06-25T10:31:35Z 2025-07-10T20:36:46Z
2026-01-04 05:59:19.9425897 +0800 CST
```

### Perigee and apogee

`PerigeesInMonth` / `ApogeesInMonth` return every perigee and apogee event in a Gregorian month. Each element is an `ApsisInfo` with `Time` (UTC) and `Distance` (km); a month may contain zero, one, or several events.

```go
// Lunar perigee and apogee in January 2026; distance is in km.
perigees := moon.PerigeesInMonth(2026, time.January)
apogees := moon.ApogeesInMonth(2026, time.January)
fmt.Printf("moon perigee=%s distance=%.1fkm count=%d\n", perigees[0].Time.Format(time.RFC3339), perigees[0].Distance, len(perigees))
fmt.Printf("moon apogee=%s distance=%.1fkm count=%d\n", apogees[0].Time.Format(time.RFC3339), apogees[0].Distance, len(apogees))
```

Output:

```text
moon perigee=2026-01-01T21:44:24Z distance=360348.1km count=2
moon apogee=2026-01-13T20:47:13Z distance=405437.9km count=1
```

### Nodes

The Moon also exposes ascending-node and descending-node longitudes, which are useful for eclipse seasons, orbital geometry, and lunar-orbit studies:

```go
nodeDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
fmt.Println(moon.AscendingNode(nodeDate), moon.DescendingNode(nodeDate))
```

The ascending / descending node definitions here are the same as in the planets chapter:

- `AscendingNode`: ecliptic longitude where the Moon crosses from south of the ecliptic to north of it
- `DescendingNode`: ecliptic longitude where the Moon crosses from north of the ecliptic to south of it
- both values are degrees, and are usually about `180°` apart at the same instant

For the `nodeDate := 2026-01-01 00:00:00 UTC` example above, the output is:

```text
340.95708624505863 160.9570862450587
```

### Maximum declinations

Lunar declination reaches northern and southern extrema within one nodal month. `MaximumDeclinationInfo` carries `Time` (the event instant) and `Declination` (the geocentric declination at that instant, in degrees).

The monthly entry points return every event in the month, while `Next*` / `Last*` / `Closest*` search by instant:

```go
// Closest maximum northern and previous maximum southern declination.
north := moon.ClosestMaximumNorthDeclination(date)
south := moon.LastMaximumSouthDeclination(date)
fmt.Println(north.Time.Format(time.RFC3339), north.Declination)
fmt.Println(south.Time.Format(time.RFC3339), south.Declination)
// Next maximum northern declination and all events in the month.
fmt.Println(moon.NextMaximumNorthDeclination(date).Time.Format(time.RFC3339))
events := moon.MaximumNorthDeclinationsInMonth(2026, time.January)
fmt.Println(len(events))
for _, event := range events {
	fmt.Println(event.Time.Format(time.RFC3339), event.Declination)
}
```

Output:

```text
2026-01-02T16:10:49+08:00 28.266373428242343
2025-12-20T07:06:57+08:00 -28.23514705130737
2026-01-02T16:10:49+08:00
2 2026-01-02T08:10:49Z 28.266373428242343
```

The monthly-list convention for the same events looks like this (input `2026-01-01 00:00:00 UTC`):

```go
// Maximum northern and southern lunar declinations in January 2026.
north := moon.MaximumNorthDeclinationsInMonth(2026, time.January)
south := moon.MaximumSouthDeclinationsInMonth(2026, time.January)
fmt.Printf("north=%s dec=%.6f\n", north[0].Time.Format(time.RFC3339), north[0].Declination)
fmt.Printf("south=%s dec=%.6f\n", south[0].Time.Format(time.RFC3339), south[0].Declination)
```

Output:

```text
north=2026-01-02T08:10:49Z dec=28.266373
south=2026-01-16T05:15:14Z dec=-28.304184
```

### Libration and bright-limb position angle

`Physical` returns the geocentric libration and `TopocentricPhysical` the topocentric libration.

Both are `PhysicalInfo`, carrying the optical, physical, and total libration components plus the rotation-axis position angle `PositionAngle`, in degrees.

The bright-limb position angle starts at `0°` at the lunar north point and increases eastward.

```go
// Geocentric libration and rotation-axis position angle.
p := moon.Physical(date)
fmt.Println(p.LibrationLongitude, p.LibrationLatitude, p.PositionAngle)
// Topocentric libration and rotation-axis position angle.
topo := moon.TopocentricPhysical(date, lon, lat, 0)
fmt.Println(topo.LibrationLongitude, topo.LibrationLatitude, topo.PositionAngle)
// Geocentric and topocentric bright-limb position angles.
fmt.Println(moon.BrightLimbPositionAngle(date), moon.TopocentricBrightLimbPositionAngle(date, lon, lat, 0))
```

Output:

```text
-0.9680924808747591 -6.547834757841939 -9.025022841390472
-0.7780085060449551 -5.659649431558411 -8.883877708625544
269.08333384819935 267.8559531949645
```

A complete topocentric example (Shanghai, `height = 4 m`):

```go
// Lunar libration and rotation-axis position angle.
physical := moon.Physical(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
fmt.Printf("libration lon=%.6f lat=%.6f pa=%.6f\n", physical.LibrationLongitude, physical.LibrationLatitude, physical.PositionAngle)
// Bright-limb position angle; 0 degrees starts at the lunar north point and increases eastward.
fmt.Printf("bright limb=%.6f\n", moon.BrightLimbPositionAngle(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
// Topocentric libration, rotation-axis position angle, and bright-limb angle from Shanghai.
topo := moon.TopocentricPhysical(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 121.4737, 31.2304, 4)
fmt.Printf("topo libration lon=%.6f lat=%.6f pa=%.6f\n", topo.LibrationLongitude, topo.LibrationLatitude, topo.PositionAngle)
fmt.Printf("topo bright limb=%.6f\n", moon.TopocentricBrightLimbPositionAngle(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 121.4737, 31.2304, 4))
```

Output:

```text
libration lon=-1.278902 lat=-6.531444 pa=-9.967050
bright limb=267.364849
topo libration lon=-1.736754 lat=-5.780730 pa=-10.072846
topo bright limb=266.038258
```

### Apparent size and Earth-Moon distance

`Diameter` / `Semidiameter` give the apparent diameter and semidiameter in arcseconds, and `EarthDistance` gives the Earth-Moon distance in kilometers. All three depend only on the absolute instant:

```go
// Apparent diameter and semidiameter (arcseconds), and Earth-Moon distance (km).
fmt.Println(moon.Diameter(date), moon.Semidiameter(date), moon.EarthDistance(date))
```

Output:

```text
1986.4975069969655 993.2487534984828 360488.4234539985
```

## Lite chains

`lite/sun` and `lite/moon` are independent approximations: they use neither the semianalytic nor Legacy tables and do not follow global model switches. They target CPU- and memory-constrained environments, and their calling style matches the main chain.

Rise/set search uses fixed-step scanning plus bisection, without the main chain's high-precision nutation iteration.

```go
package main

import (
	"fmt"
	"time"

	litemoon "github.com/starainrt/astro/lite/moon"
	litesun "github.com/starainrt/astro/lite/sun"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2026, 1, 1, 20, 0, 0, 0, cst)

	fmt.Println(litesun.Altitude(date, 121.4737, 31.2304))
	fmt.Println(litesun.RiseTime(date, 121.4737, 31.2304, 0, true))

	fmt.Println(litemoon.Phase(date))
	fmt.Println(litemoon.PhaseAge(date))
	fmt.Println(litemoon.RiseTime(date, 121.4737, 31.2304, 0, true))
}
```

The snippets below reuse this section's shared setup: Shanghai (`121.4737°E, 31.2304°N`) and `2026-01-01 20:00:00 CST`.

### lite/sun

The lightweight solar chain provides longitude, equatorial coordinates, distance, horizontal coordinates, and rise/set, but no apparent size, solar disk physical quantities, or twilight entry points.

`Distance` is the Earth-Sun distance in AU; note that the main chain names this capability `EarthDistance` while the lite chain names it `Distance`.

```go
// Lightweight true and apparent solar longitude, and Earth-Sun distance (AU).
fmt.Println(litesun.TrueLo(date), litesun.ApparentLo(date), litesun.Distance(date))
// Lightweight apparent right ascension and declination.
ra, dec := litesun.ApparentRaDec(date)
fmt.Println(ra, dec)
// Lightweight hour angle, azimuth, altitude, and zenith distance.
fmt.Println(litesun.HourAngle(date, 121.4737, 31.2304), litesun.Azimuth(date, 121.4737, 31.2304), litesun.Altitude(date, 121.4737, 31.2304), litesun.Zenith(date, 121.4737, 31.2304))
// Lightweight rise and set.
fmt.Println(litesun.RiseTime(date, 121.4737, 31.2304, 0, true))
fmt.Println(litesun.SetTime(date, 121.4737, 31.2304, 0, true))
```

Output:

```text
281.0835889076667 281.0793650422643 0.9833163427233701
282.0475744673639 -22.973803828458102
120.58011725187828 263.4579582386061 -37.07676039520773 127.07676039520773
2026-01-01 06:52:24.6475178 +0800 CST <nil>
2026-01-01 17:02:44.014452695 +0800 CST <nil>
```

### lite/moon

The lightweight lunar chain provides a few-perturbation-term lunar position, light topocentric correction, phase and age, and rise/set, but no libration, apparent size, Earth-Moon distance, or nodes.

`PhaseAge` returns the lunar age in days and exists only in the lite chain:

```go
// Lightweight true longitude and latitude.
fmt.Println(litemoon.TrueLo(date), litemoon.TrueBo(date))
// Lightweight geocentric true equatorial coordinates.
ra, dec := litemoon.TrueRaDec(date)
fmt.Println(ra, dec)
// Lightweight topocentric apparent equatorial coordinates.
fmt.Println(litemoon.ApparentRaDec(date, 121.4737, 31.2304))
// Moon-Sun longitude difference, illuminated fraction, and lunar age.
fmt.Println(litemoon.SunMoonLoDiff(date), litemoon.Phase(date), litemoon.PhaseAge(date))
// Lightweight horizontal coordinates.
fmt.Println(litemoon.Altitude(date, 121.4737, 31.2304), litemoon.Azimuth(date, 121.4737, 31.2304), litemoon.Zenith(date, 121.4737, 31.2304))
```

Output:

```text
74.25630740893 5.078407838127742
72.2518040645064 27.549605139478064
72.74256784654426 27.432483413326057
153.17694236666568 0.9462021494002484 12.565014741082448
63.55513206820331 90.53047230027812 26.444867931796693
```

### Differences from the main chain and error levels

The lite chains differ from the main chain in implementation and accuracy while keeping nearly the same interface shapes.

Performance depends on the main-chain model, epoch cache and whether a search is required; there is no universal speedup factor. Lite's listed benchmarks allocate zero times, and rise/set scan every 30/15 minutes. Current model-specific measurements are in [Accuracy and Performance](accuracy.md#performance).

The following are historical comparisons against the old Legacy `sun` / `moon` chain (2026, 8 sites), not a new semianalytic validation; see [Lite lightweight chains](accuracy.md#lite-lightweight-chains):

| Quantity | Mean absolute error | P95 | Max absolute error |
| --- | --- | --- | --- |
| `lite/sun` sunrise | `0.02 min` | `0.04 min` | `0.31 min` |
| `lite/sun` sunset | `0.02 min` | `0.06 min` | `0.35 min` |
| `lite/moon` moonrise | `0.28 min` | `0.57 min` | `1.44 min` |
| `lite/moon` moonset | `0.36 min` | `0.86 min` | `1.24 min` |
| `lite/moon` `Phase()` | `0.00089` | `0.00185` | `0.00243` |
| `lite/moon` `PhaseAge()` | `0.003 d` | `0.010 d` | `0.014 d` |
| `lite/moon` geocentric longitude | `2.41'` | `6.82'` | `9.91'` |
| `lite/moon` geocentric latitude | `0.87'` | `1.83'` | `2.92'` |

Neither package provides a `...N` family. On the solar side the polar-night / polar-day errors are `ERR_SUN_NEVER_RISE` / `ERR_SUN_NEVER_SET`; on the lunar side, besides `ERR_MOON_NEVER_RISE` / `ERR_MOON_NEVER_SET`, there is `ERR_NOT_TODAY`, with the same semantics as the main chain.

## Parameter and result conventions

### Units and angle conventions

- Angles are always in degrees; `RA`, `Lon`, and `Azimuth` are normalized to `[0°, 360°)`, while declination and ecliptic latitude lie in `[−90°, 90°]`
- Apparent diameters and semidiameters are in arcseconds; `sun.EarthDistance` is AU and `moon.EarthDistance` is kilometers
- `Phase` is the illuminated fraction in `[0,1]`, `PhaseAge` (lite chain) is in days, and `EquationTime` is in hours
- Distance extrema: `earth.Perihelion` / `earth.Aphelion` are AU, and `ApsisInfo.Distance` is kilometers

### Time scale

Observing inputs and event times use the civil-time convention. Before `1972-01-01`, the library treats those readings as UT1. `ApparentSolarTime` returns a local solar clock reading rather than another civil event time.

The time zone carried by a `time.Time` only affects the division of local civil days and the zone of the returned value, never the absolute instant.

The full UT1 and chart-label declaration is in the [Time Scale Declaration](map-geojson.md#time-scale-declaration).

### Height and aero

- The `height` of the rise/set entry points is the observer elevation interpreted as ellipsoidal (geodetic) height in meters, not orthometric height; see [Observer Height Convention](coord.md#observer-height)
- `aero = true`: upper-limb crossing with dynamic standard refraction and the instantaneous semidiameter; `aero = false`: only the geometric center crossing
- Rise/set refraction uses the standard atmosphere (1010 hPa, 10 °C) by default. For measured weather use `sun.RiseTimeWithAtmosphere` / `SetTimeWithAtmosphere` and `moon.RiseTimeWithAtmosphere` / `SetTimeWithAtmosphere`, passing pressure (hPa) and temperature (°C) on each call; refraction is always applied. Existing APIs are unchanged
- Body-independent true/apparent altitude conversion lives in `formula`: `formula.ApparentAltitude`, `formula.TrueAltitude`, `formula.RefractionFromTrueAltitude`, `formula.RefractionFromApparentAltitude`
- `pressureHPa` and `temperatureC` of `ApparentAltitude` / `ApparentZenith` are the observed pressure (hPa) and temperature (degrees Celsius)

### Zero values and out-of-range

- During polar night the rise/set entry points return `ERR_SUN_NEVER_RISE` / `ERR_MOON_NEVER_RISE`; during polar day they return `ERR_SUN_NEVER_SET` / `ERR_MOON_NEVER_SET`, with a zero `time.Time` as the first return value
- When no twilight exists, they return `ERR_TWILIGHT_NOT_EXISTS`
- When a lunar rise/set event falls outside the queried date, they return `ERR_NOT_TODAY`, matching the "moonrise/moonset are computed for the queried civil date" semantics above
- `DownTime` / `DownTimeN` are deprecated aliases of `SetTime`, and `ERR_SUN_NEVER_DOWN` / `ERR_MOON_NEVER_DOWN` are deprecated aliases of the polar-day errors; new code should not use them
- In `...N`, `n < 0` selects the full series; all `n >= 0` select the same semianalytic fast table, while Legacy truncates according to `n`. Units and time zones never change

### Accuracy and scope

Default builds use the embedded semianalytic model. The VSOP87 / ELP2000/82-style series remain as Legacy; neither requires external files. An available range is not the same as a measured range, and a measured error only holds for the measured samples.

Model-specific DE440 / DE441 Sun and Moon comparisons, fast-table errors and Lite limitations are documented in [Accuracy and performance](accuracy.md).

The solar and lunar entry points in this manual suit calendars, observing aids, outreach, and amateur prediction; spacecraft navigation, precise occultation prediction, and rigorous dynamical integration require a professional ephemeris such as JPL DE.
