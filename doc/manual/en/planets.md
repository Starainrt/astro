# Planets

[中文](../planets.md) | [Back to README](../../../README.en.md)

The seven major planets each have a package of their own: `mercury`, `venus`, `mars`, `jupiter`, `saturn`, `uranus`, and `neptune`.

They share the same API shape: a civil instant is passed as `time.Time`, angles and distances come back as `float64`, and event searches return `time.Time` or a struct.

The inner planets (Mercury and Venus) add superior/inferior conjunction, greatest elongation, and geocentric and topocentric transits; the outer planets (Mars through Neptune) add opposition and quadrature; Jupiter alone has the Galilean satellites and Saturn alone has ring parameters. The `planet` package holds the built-in semianalytic model and the Legacy series, shared by these packages and by `sun` / `moon`. Default builds use to semianalytic; use an explicit Context to fix the model.

- The function names of the common capabilities are identical across the seven packages (for example all of them provide `ApparentRa`, `ApparentDec`, and `ApparentRaDec`); only the extra families differ.

  Always qualify a call with its package name instead of mixing packages.
- Position functions return a **geocentric apparent place**.

  Horizontal quantities are separate: `Altitude` / `Azimuth` / `Zenith` are topocentric, while `HourAngle` / `ParallacticAngle` are geocentric (no topocentric parallax), with formulas under [Coordinate Tools](coord.md).
- Event-search families come as `Last...` / `Next...` pairs (some also `Closest...`) and always return the nearest event at or before/after the input instant, endpoints included.
- The planets have no dedicated SVG chart entry point; occultation-related charts (including Saturn-ring occultations) are documented in [Lunar Occultations](occultation.md#lunar-occultation-charts).

## Contents

- [The position and rise time of Mars](#the-position-and-rise-time-of-mars)
- [API Reference](#api-reference)
  - [Cross-planet comparison](#cross-planet-comparison)
  - [Common capabilities and units](#common-capabilities-and-units)
  - [The `...N` family](#the-n-family)
- [Usage examples](#usage-examples)
  - [Which planet can I see tonight?](#which-planet-can-i-see-tonight)
  - [Oppositions, conjunctions, elongations, stations and retrogrades](#oppositions-conjunctions-elongations-stations-and-retrogrades)
  - [Mercury and Venus transits](#mercury-and-venus-transits)
  - [Phase, diameter, magnitude and nodes](#phase-diameter-magnitude-and-nodes)
  - [Physical ephemerides and the Galilean satellites](#physical-ephemerides-and-the-galilean-satellites)
  - [Comparison conventions](#comparison-conventions)
- [Basic examples](#basic-examples)
  - [Inner planets](#inner-planets)
  - [Outer planets](#outer-planets)
- [Topic examples](#topic-examples)
  - [Position and coordinates](#position-and-coordinates)
  - [Rise, set and culmination](#rise-set-and-culmination)
  - [Conjunctions, oppositions, stations and quadratures](#conjunctions-oppositions-stations-and-quadratures)
  - [Greatest elongation and transits](#greatest-elongation-and-transits)
  - [Nodes, phase, magnitude, diameter and parallactic angle](#nodes-phase-magnitude-diameter-and-parallactic-angle)
  - [Division of labour with other manuals](#division-of-labour-with-other-manuals)
  - [Physical ephemerides](#physical-ephemerides)
  - [Galilean satellites of Jupiter](#galilean-satellites-of-jupiter)
  - [Shared low-level entry points in the `planet` package](#shared-low-level-entry-points-in-the-planet-package)
- [Parameter and result conventions](#parameter-and-result-conventions)
  - [Time scale and civil time](#time-scale-and-civil-time)
  - [Units and conventions](#units-and-conventions)
  - [Zero values, out-of-range input and sentinel errors](#zero-values-out-of-range-input-and-sentinel-errors)
  - [Topocentric versus geocentric](#topocentric-versus-geocentric)
  - [Accuracy and scope](#accuracy-and-scope)

## The position and rise time of Mars

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro/mars"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2020, 1, 1, 8, 8, 8, 0, cst)
	lon, lat, height := 108.93, 34.27, 0.0
	ra, dec := mars.ApparentRaDec(date)
	fmt.Printf("RA=%.6f Dec=%.6f deg\n", ra, dec)
	rise, err := mars.RiseTime(date, lon, lat, height, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(rise.Format(time.RFC3339))
}
```

`ApparentRaDec` returns geocentric apparent right ascension and declination in degrees. Rise/set also needs a site and ellipsoidal height, and returns an error when no rise event exists.

## API Reference

### Cross-planet comparison

Rows list capabilities and columns list planetary packages. Qualify each function with its package, for example `mars.NextOpposition`. A slash separates function names; `—` means the package has no corresponding API.

| Quantity | `mercury` | `venus` | `mars` | `jupiter` | `saturn` | `uranus` | `neptune` |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Apparent ecliptic longitude / latitude | `ApparentLo` / `ApparentBo` | `ApparentLo` / `ApparentBo` | `ApparentLo` / `ApparentBo` | `ApparentLo` / `ApparentBo` | `ApparentLo` / `ApparentBo` | `ApparentLo` / `ApparentBo` | `ApparentLo` / `ApparentBo` |
| Apparent right ascension / declination | `ApparentRa` / `ApparentDec` / `ApparentRaDec` | `ApparentRa` / `ApparentDec` / `ApparentRaDec` | `ApparentRa` / `ApparentDec` / `ApparentRaDec` | `ApparentRa` / `ApparentDec` / `ApparentRaDec` | `ApparentRa` / `ApparentDec` / `ApparentRaDec` | `ApparentRa` / `ApparentDec` / `ApparentRaDec` | `ApparentRa` / `ApparentDec` / `ApparentRaDec` |
| Apparent magnitude | `ApparentMagnitude` | `ApparentMagnitude` | `ApparentMagnitude` | `ApparentMagnitude` | `ApparentMagnitude` | `ApparentMagnitude` | `ApparentMagnitude` |
| Geocentric / heliocentric distance | `EarthDistance` / `SunDistance` | `EarthDistance` / `SunDistance` | `EarthDistance` / `SunDistance` | `EarthDistance` / `SunDistance` | `EarthDistance` / `SunDistance` | `EarthDistance` / `SunDistance` | `EarthDistance` / `SunDistance` |
| Orbital ascending / descending node | `AscendingNode` / `DescendingNode` | `AscendingNode` / `DescendingNode` | `AscendingNode` / `DescendingNode` | `AscendingNode` / `DescendingNode` | `AscendingNode` / `DescendingNode` | `AscendingNode` / `DescendingNode` | `AscendingNode` / `DescendingNode` |
| Apparent diameter / semidiameter | `Diameter` / `Semidiameter` | `Diameter` / `Semidiameter` | `Diameter` / `Semidiameter` | `Diameter` / `Semidiameter` | `Diameter` / `Semidiameter` | `Diameter` / `Semidiameter` | `Diameter` / `Semidiameter` |
| Phase angle / illuminated fraction / bright-limb position angle | `PhaseAngle` / `Phase` / `IlluminatedFraction` / `BrightLimbPositionAngle` | `PhaseAngle` / `Phase` / `IlluminatedFraction` / `BrightLimbPositionAngle` | `PhaseAngle` / `Phase` / `IlluminatedFraction` / `BrightLimbPositionAngle` | `PhaseAngle` / `Phase` / `IlluminatedFraction` / `BrightLimbPositionAngle` | `PhaseAngle` / `Phase` / `IlluminatedFraction` / `BrightLimbPositionAngle` | `PhaseAngle` / `Phase` / `IlluminatedFraction` / `BrightLimbPositionAngle` | `PhaseAngle` / `Phase` / `IlluminatedFraction` / `BrightLimbPositionAngle` |
| Horizontal quantities | `Altitude` / `Azimuth` / `Zenith` / `HourAngle` / `ParallacticAngle` | `Altitude` / `Azimuth` / `Zenith` / `HourAngle` / `ParallacticAngle` | `Altitude` / `Azimuth` / `Zenith` / `HourAngle` / `ParallacticAngle` | `Altitude` / `Azimuth` / `Zenith` / `HourAngle` / `ParallacticAngle` | `Altitude` / `Azimuth` / `Zenith` / `HourAngle` / `ParallacticAngle` | `Altitude` / `Azimuth` / `Zenith` / `HourAngle` / `ParallacticAngle` | `Altitude` / `Azimuth` / `Zenith` / `HourAngle` / `ParallacticAngle` |
| Rise / set / culmination | `RiseTime` / `SetTime` / `DownTime` / `CulminationTime` | `RiseTime` / `SetTime` / `DownTime` / `CulminationTime` | `RiseTime` / `SetTime` / `DownTime` / `CulminationTime` | `RiseTime` / `SetTime` / `DownTime` / `CulminationTime` | `RiseTime` / `SetTime` / `DownTime` / `CulminationTime` | `RiseTime` / `SetTime` / `DownTime` / `CulminationTime` | `RiseTime` / `SetTime` / `DownTime` / `CulminationTime` |
| Physical ephemeris | `Physical` / `PhysicalN` | `Physical` / `PhysicalN` | `Physical` / `PhysicalN` | `Physical` / `PhysicalN` / `CentralMeridians` / `CentralMeridiansN` | `Physical` / `PhysicalN` / `PhysicalSystemIII` / `Ring` | `Physical` / `PhysicalN` / `PhysicalSystemIII` | `Physical` / `PhysicalN` |
| Conjunction | `LastConjunction` / `NextConjunction` | `LastConjunction` / `NextConjunction` | `LastConjunction` / `NextConjunction` | `LastConjunction` / `NextConjunction` | `LastConjunction` / `NextConjunction` | `LastConjunction` / `NextConjunction` | `LastConjunction` / `NextConjunction` |
| Superior / inferior conjunction | `LastSuperiorConjunction` / `NextSuperiorConjunction` / `LastInferiorConjunction` / `NextInferiorConjunction` | `LastSuperiorConjunction` / `NextSuperiorConjunction` / `LastInferiorConjunction` / `NextInferiorConjunction` | — | — | — | — | — |
| Station | `LastProgradeToRetrograde` / `NextProgradeToRetrograde` / `LastRetrogradeToPrograde` / `NextRetrogradeToPrograde`, plus `LastRetrograde` / `NextRetrograde` | `LastProgradeToRetrograde` / `NextProgradeToRetrograde` / `LastRetrogradeToPrograde` / `NextRetrogradeToPrograde`, plus `LastRetrograde` / `NextRetrograde` | `LastProgradeToRetrograde` / `NextProgradeToRetrograde` / `LastRetrogradeToPrograde` / `NextRetrogradeToPrograde` | `LastProgradeToRetrograde` / `NextProgradeToRetrograde` / `LastRetrogradeToPrograde` / `NextRetrogradeToPrograde` | `LastProgradeToRetrograde` / `NextProgradeToRetrograde` / `LastRetrogradeToPrograde` / `NextRetrogradeToPrograde` | `LastProgradeToRetrograde` / `NextProgradeToRetrograde` / `LastRetrogradeToPrograde` / `NextRetrogradeToPrograde` | `LastProgradeToRetrograde` / `NextProgradeToRetrograde` / `LastRetrogradeToPrograde` / `NextRetrogradeToPrograde` |
| Opposition | — | — | `LastOpposition` / `NextOpposition` | `LastOpposition` / `NextOpposition` | `LastOpposition` / `NextOpposition` | `LastOpposition` / `NextOpposition` | `LastOpposition` / `NextOpposition` |
| Quadrature | — | — | `LastEasternQuadrature` / `NextEasternQuadrature` / `LastWesternQuadrature` / `NextWesternQuadrature` | `LastEasternQuadrature` / `NextEasternQuadrature` / `LastWesternQuadrature` / `NextWesternQuadrature` | `LastEasternQuadrature` / `NextEasternQuadrature` / `LastWesternQuadrature` / `NextWesternQuadrature` | `LastEasternQuadrature` / `NextEasternQuadrature` / `LastWesternQuadrature` / `NextWesternQuadrature` | `LastEasternQuadrature` / `NextEasternQuadrature` / `LastWesternQuadrature` / `NextWesternQuadrature` |
| Greatest elongation | `LastGreatestElongation` / `NextGreatestElongation` / `LastGreatestElongationEast` / `NextGreatestElongationEast` / `LastGreatestElongationWest` / `NextGreatestElongationWest` | `LastGreatestElongation` / `NextGreatestElongation` / `LastGreatestElongationEast` / `NextGreatestElongationEast` / `LastGreatestElongationWest` / `NextGreatestElongationWest` | — | — | — | — | — |
| Transit | `LastTransit` / `NextTransit` / `ClosestTransit` / `LocalTransit` | `LastTransit` / `NextTransit` / `ClosestTransit` / `LocalTransit` | — | — | — | — | — |
| Galilean satellites | — | — | — | `Satellites` / `SatellitePhenomena` / `LastGalileanPhenomenonEvent` / `NextGalileanPhenomenonEvent` / `ClosestGalileanPhenomenonEvent` / `LastGalileanPhenomenonContactEvent` / `NextGalileanPhenomenonContactEvent` / `ClosestGalileanPhenomenonContactEvent` | — | — | — |
| Result types | `PhysicalInfo` / `TransitInfo` / `LocalTransitInfo` | `PhysicalInfo` / `TransitInfo` / `LocalTransitInfo` | `PhysicalInfo` | `PhysicalInfo` / `CentralMeridianInfo` / `GalileanSatellitesInfo` / `GalileanPhenomenaInfo` / `GalileanSatellitePosition` / `GalileanSatellitePhenomenon` / `GalileanPhenomenonEvent` / `GalileanPhenomenonContactEvent` | `PhysicalInfo` / `RingInfo` | `PhysicalInfo` | `PhysicalInfo` |

All seven packages spell the **`...N` family** the same way: append `N` to an instantaneous evaluation function. `n < 0` selects the full series; all nonnegative `n` select one fixed semianalytic fast table, while Legacy truncates according to `n`.

See [the `...N` family](#the-n-family) for details. Event-search families (`Last...` / `Next...` / `Closest...`) have no `N` variants.

### Common capabilities and units

| Quantity | Purpose | Unit and convention |
| --- | --- | --- |
| `ApparentLo` / `ApparentBo` | Geocentric apparent ecliptic longitude and latitude | degrees; true equinox of date, including light time, aberration, and nutation |
| `ApparentRa` / `ApparentDec` / `ApparentRaDec` | Geocentric apparent right ascension and declination | degrees; true equator of date; `ApparentRaDec` returns both at once |
| `ApparentMagnitude` | Apparent visual magnitude | magnitudes (dimensionless) |
| `Altitude` / `Azimuth` / `Zenith` | Topocentric horizontal coordinates | degrees; azimuth runs from north toward east, and `Zenith` equals `90 - Altitude` |
| `HourAngle` | Geocentric apparent hour angle, without topocentric parallax | degrees |
| `RiseTime` / `SetTime` / `DownTime` | Rise and set within the local civil day | `time.Time` keeping the input time zone; `DownTime` is a compatibility alias for `SetTime` |
| `RiseTimeWithAtmosphere` / `SetTimeWithAtmosphere` | Rise and set with measured pressure (hPa) and temperature (°C) | `time.Time`; refraction always applied |
| `CulminationTime` | Upper culmination instant | `time.Time` keeping the input time zone |
| `ParallacticAngle` | Geocentric parallactic angle (zenith direction angle), from the geocentric hour angle and declination | degrees |
| `Diameter` / `Semidiameter` | Geocentric apparent diameter and semidiameter | arcseconds |
| `PhaseAngle` | Sun-planet-Earth angle | degrees |
| `IlluminatedFraction` / `Phase` | Illuminated fraction | `0-1`; `Phase` is an alias for `IlluminatedFraction` |
| `BrightLimbPositionAngle` | Position angle of the bright-limb center | degrees |
| `EarthDistance` / `SunDistance` | Earth distance and Sun distance | AU |
| `AscendingNode` / `DescendingNode` | Ecliptic longitude of the orbital plane's intersections with the ecliptic | degrees; the two differ by about `180°` at the same instant |
| `Physical` | Disk orientation, sub-Earth/sub-Solar coordinates, north-pole position angle | degrees; the positive longitude direction follows each body's IAU convention |
| `planet.WherePlanet` / `planet.WherePlanetN` | Selected built-in model's longitude, latitude, and heliocentric distance | degrees / AU; out-of-range input returns `NaN` instead of panicking |

### The `...N` family

Every instantaneous evaluation function has an `...N` variant that trades accuracy for cost. `n < 0` selects the full series and matches the plain function. All nonnegative `n` select one fixed semianalytic fast table, so increasing `n` does not improve accuracy; Legacy scales its principal and higher-order terms according to `n`.

Event-search families (`Last...` / `Next...` / `Closest...`) have no `N` variant.

Functions with an `N` variant, grouped by purpose:

- Position and distance: `ApparentLoN`, `ApparentBoN`, `ApparentRaN`, `ApparentDecN`, `ApparentRaDecN`, `EarthDistanceN`, `SunDistanceN`
- Horizon and instants: `AltitudeN`, `AzimuthN`, `ZenithN`, `HourAngleN`, `CulminationTimeN`, `RiseTimeN`, `SetTimeN`, `DownTimeN`, `ParallacticAngleN`
- Disc and phase: `ApparentMagnitudeN`, `DiameterN`, `SemidiameterN`, `PhaseAngleN`, `PhaseN`, `IlluminatedFractionN`, `BrightLimbPositionAngleN`
- Nodes and rotation: `AscendingNodeN`, `DescendingNodeN`, `PhysicalN`, plus `CentralMeridiansN` for Jupiter, `RingN` and `PhysicalSystemIIIN` for Saturn, and `PhysicalSystemIIIN` for Uranus

```go
fmt.Println(mars.ApparentLo(date), mars.ApparentLoN(date, 8))   // full series / fast table; Legacy truncates according to n
fmt.Println(mars.SunDistance(date), mars.SunDistanceN(date, 8)) // heliocentric distance, AU
```

## Usage examples

### Which planet can I see tonight?

```go
fmt.Println(venus.RiseTime(date, lon, lat, height, true))                // Venus rises today
fmt.Println(jupiter.CulminationTime(date, lon))                          // Jupiter's upper culmination
fmt.Println(mars.Altitude(date, lon, lat), mars.Azimuth(date, lon, lat)) // Mars's altitude and azimuth now
```

```text
2020-01-01 10:02:34.350145161 +0800 CST <nil>
2020-01-01 12:32:17.585815787 +0800 CST
31.194578177219057 152.07031660415714
```

Azimuth increases from due north towards the east; `aero = true` lowers the geometric horizon to about `-0.5667°` for standard refraction.

Per-parameter conventions and the polar sentinel errors are in [Rise, set and culmination](#rise-set-and-culmination).

### Oppositions, conjunctions, elongations, stations and retrogrades

```go
fmt.Println(mars.NextOpposition(date))             // Mars's next opposition
fmt.Println(jupiter.NextConjunction(date))         // Jupiter's next conjunction
fmt.Println(saturn.NextProgradeToRetrograde(date)) // Saturn's next prograde-to-retrograde station
```

```text
2020-10-14 07:25:50.441412627 +0800 CST
2021-01-29 09:39:33.697994649 +0800 CST
2020-05-11 17:26:53.961271941 +0800 CST
```

Event searches return the nearest event at or after the input instant and keep its time zone; the paired `Last...` functions and the direction-agnostic `NextRetrograde` are in [Conjunctions, oppositions, stations and quadratures](#conjunctions-oppositions-stations-and-quadratures);

Mercury's and Venus's `NextGreatestElongationEast` / `...West` are in [Greatest elongation and transits](#greatest-elongation-and-transits).

### Mercury and Venus transits

```go
transit := mercury.NextTransit(date)                           // next Mercury transit after 2020
fmt.Println(transit.Valid, transit.Start, transit.Greatest)     // whether one was found, first contact, greatest transit
fmt.Println(transit.Duration, transit.MinimumSeparationArcsec)  // duration and minimum separation at greatest transit
```

```text
true 2032-11-13 14:41:13.691189289 +0800 CST 2032-11-13 16:54:13.351347148 +0800 CST
4h26m2.484410405s 572.0831814505963
```

`TransitInfo.Valid` false means no transit inside the search window and every other field is the zero value; otherwise `TransitInfo` describes a global event whose contact fields are geocentric reference times. Without geocentric contacts `HasExternal` is false, `Start` / `End` are zero, and only the `WindowStart` / `WindowEnd` search window is set. For the contact times seen from one site, `LocalTransit` returns the topocentric result:

```go
transit := mercury.NextTransit(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC))
local := mercury.LocalTransit(transit, -77.0, 38.9, 0)        // Washington: longitude, latitude, height in metres
fmt.Println(local.Valid, local.HasInternal, local.Visible)     // topocentric transit, internal contacts, Sun above the apparent horizon
fmt.Println(local.Start.Time, local.Start.SunAltitude)         // topocentric C1 and the Sun's altitude then
```

```text
true true true
2019-11-11 12:36:08.979182839 +0000 UTC 7.933092450591312
```

Topocentric contacts are re-solved after applying parallax to the Sun and the planet separately; Venus contacts can shift by about ±7 minutes from the geocentric ones across the globe, Mercury by about ±2 minutes. `SunAltitude` is the geometric altitude of the Sun's center; `Visible`, `Sunrise` and `Sunset` test whether the Sun's center is above the apparent horizon with standard refraction (1010 hPa, 10 °C), a geometric threshold of about −34′. A site may see only a grazing transit even when the geocentric one has internal contacts (the 1999-11-15 Mercury transit at high southern latitudes); then `HasInternal` is false and `InternalStart` / `InternalEnd` are zero. Global maps and topocentric disk charts are in [Eclipses and transits](eclipse.md#transit-charts).

The four contacts, partial transits and the internal-contact fields are in [Greatest elongation and transits](#greatest-elongation-and-transits).

`Gamma` is the signed distance from Earth's centre to the shadow axis at greatest (positive to the north) and `PenumbralRadius` is the penumbral radius on the fundamental plane, both in Earth equatorial radii. The transit penumbra is far larger than Earth, so the solar-eclipse test `|Gamma| < 1` does not apply. `Series` is the transit family: Venus families average 243 years and are numbered for astronomical years −2000 to +4000, Mercury families average 46 years and are numbered for +1601 to +2300; outside those spans, or for an unnumbered family, `Number` is zero while `PeriodYears` still gives the mean period.

### Phase, diameter, magnitude and nodes

```go
fmt.Println(venus.PhaseAngle(date), venus.Phase(date))           // phase angle (degrees) and illuminated fraction
fmt.Println(venus.Diameter(date), venus.ApparentMagnitude(date)) // apparent diameter (arcseconds) and magnitude
fmt.Println(mars.AscendingNode(date), mars.DescendingNode(date)) // ascending/descending node longitudes (degrees)
```

```text
49.98145049145023 0.8215177914415865
13.059409604614839 -4
49.71479005849112 229.71479005849113
```

`Phase` is an alias of `IlluminatedFraction` and runs from `0` to `1`; the two nodes are about `180°` apart at any instant.

Definitions, aliases and the `n` variants are in [Nodes, phase, magnitude, diameter and parallactic angle](#nodes-phase-magnitude-diameter-and-parallactic-angle).

### Physical ephemerides and the Galilean satellites

```go
j := jupiter.Physical(date)                                    // Jupiter's physical ephemeris
fmt.Println(j.DS, j.DE, j.CentralMeridianSystemIII)            // sub-solar/sub-Earth latitude and System III central meridian
sats := jupiter.Satellites(date)                               // instantaneous positions of the four Galilean satellites
fmt.Println(sats.Io.OffsetXJupiterR, sats.Io.InFrontOfJupiter) // Io's offset and whether it is in front of the disk
```

```text
-56.55778470155335 -2.039966127259664 311.37430665615585
3.8223302102343975 false
```

Saturn's rings have their own `saturn.Ring` (`EarthLatitude`, `MinorAxis`, and so on; angles in degrees and axes in arcseconds). Disk orientation and the central meridians are covered in [Physical ephemerides](#physical-ephemerides); the Galilean-satellite public entry points are in the `jupiter` package, while `JupiterGalilean*` in `basic` is the low-level entry that takes a Julian day, see [Galilean satellites of Jupiter](#galilean-satellites-of-jupiter).

### Comparison conventions

```go
fmt.Println(mars.ApparentLo(date), mars.ApparentLoN(date, 8)) // full series / fast table; Legacy truncates according to n
_, err := mars.RiseTime(date, 0, 89, 0, true)                 // an observing site in the polar region
fmt.Println(errors.Is(err, mars.ERR_MARS_NEVER_RISE), errors.Is(err, mars.ERR_MARS_NEVER_SET))
```

```text
238.38840227925655 238.39637464888327
true false
```

`ApparentLo` and friends are geocentric apparent places referred to the **true equinox of date**. Comparing them with external J2000 or mean places requires matching the frame, nutation and apparent-place corrections; precession alone does not remove every difference.

For model-dependent `n` semantics, see [the `...N` family](#the-n-family).

The convention gap between the Galilean contact events and JPL Horizons / IMCCE tables is documented under [External baselines](#external-baselines), and the overall boundaries under [Parameter and result conventions](#parameter-and-result-conventions).

## Basic examples

The two examples below are the smallest runnable programs and use `date = 2020-01-01 08:08:08 CST` with the Xi'an coordinates.

### Inner planets

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/mercury"
	"github.com/starainrt/astro/venus"
)

func main() {
	// Xi'an, China. Longitude east and latitude north are positive; elevation is 0 m.
	var lon, lat, height float64 = 108.93, 34.27, 0
	cst := time.FixedZone("CST", 8*3600)
	// Instant of observation.
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)

	// Previous inferior conjunction of Mercury.
	fmt.Println(mercury.LastInferiorConjunction(date))
	// Next superior conjunction of Venus.
	fmt.Println(venus.NextSuperiorConjunction(date))
	// Previous Mercury station from prograde to retrograde.
	fmt.Println(mercury.LastProgradeToRetrograde(date))
	// Next Venus station from retrograde to prograde.
	fmt.Println(venus.NextRetrogradeToPrograde(date))
	// Previous greatest eastern elongation of Mercury.
	fmt.Println(mercury.LastGreatestElongationEast(date))
	// Next greatest western elongation of Venus.
	fmt.Println(venus.NextGreatestElongationWest(date))
	// Venus rise and set times in Xi'an.
	fmt.Println(venus.RiseTime(date, lon, lat, height, true))
	fmt.Println(venus.SetTime(date, lon, lat, height, true))
	// Current apparent magnitude of Venus.
	fmt.Println(venus.ApparentMagnitude(date))
	// Venus phase angle, illuminated fraction, and bright-limb position angle.
	fmt.Println(venus.PhaseAngle(date))
	fmt.Println(venus.Phase(date))
	fmt.Println(venus.BrightLimbPositionAngle(date))
	// Earth-Venus distance.
	fmt.Println(venus.EarthDistance(date))
	// Sun-Venus distance.
	fmt.Println(venus.SunDistance(date))
}
```

Output:

```text
2019-11-11 23:21:41.729008555 +0800 CST // previous inferior conjunction of Mercury
2021-03-26 14:57:43.352850079 +0800 CST // next superior conjunction of Venus
2019-11-01 04:31:49.712648391 +0800 CST // previous Mercury station from prograde to retrograde
2020-06-25 02:07:26.280500292 +0800 CST // next Venus station from retrograde to prograde
2019-10-20 12:01:37.925546765 +0800 CST // previous greatest eastern elongation of Mercury
2020-08-13 08:14:45.946914553 +0800 CST // next greatest western elongation of Venus
2020-01-01 10:02:34.94517982 +0800 CST <nil> // Venus rise time in Xi'an; no error
2020-01-01 20:25:36.944952607 +0800 CST <nil> // Venus set time in Xi'an; no error
-4 // Venus apparent magnitude
49.98142313573699 // Venus phase angle, degrees
0.8215179742647736 // illuminated fraction of Venus
255.63803631896667 // bright-limb position angle of Venus, degrees
1.2778823747747166 // Earth-Venus distance, AU
0.7262651318993368 // Sun-Venus distance, AU
```

Inner and outer planets also expose `Diameter` / `Semidiameter` and `N` variants, returning geocentric apparent diameter/semidiameter in arcseconds.

Planet apparent diameter and orbital nodes can be queried directly:

```go
fmt.Println(mars.Diameter(date), mars.Semidiameter(date))                 // Martian apparent diameter and semidiameter, arcseconds
fmt.Println(venus.AscendingNode(date), venus.DescendingNode(date))         // Venus ascending-node and descending-node ecliptic longitudes, degrees
```

Ascending node / descending node here means the two intersections of the body's orbital plane with the ecliptic:

- `AscendingNode`: ecliptic longitude where the body crosses from south of the ecliptic to north of it
- `DescendingNode`: ecliptic longitude where the body crosses from north of the ecliptic to south of it
- return values are degrees; for the same instant, descending node is usually about `180°` from ascending node

For `date := 2020-01-01 08:08:08 CST`, the output is:

```text
4.287299886569956 2.143649943284978 // Mars apparent diameter and semidiameter, arcseconds
76.86008484515058 256.8600848451506 // Venus ascending-node and descending-node longitudes, degrees
```

Mercury and Venus also expose `NextTransit` / `LastTransit` / `ClosestTransit` for geocentric planetary transits. "Geocentric" means the planet disk crosses the solar disk as seen from Earth's center; it does not test whether the Sun is above the horizon at a particular observing site. For observing plans, use `LocalTransit` for topocentric contacts and solar altitudes (see above).

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/mercury"
	"github.com/starainrt/astro/venus"
)

func main() {
	// Next geocentric Mercury transit after the beginning of 2019.
	mercuryTransit := mercury.NextTransit(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC))
	fmt.Println(mercuryTransit.Valid)
	fmt.Println(mercuryTransit.Start)
	fmt.Println(mercuryTransit.InternalStart)
	fmt.Println(mercuryTransit.Greatest)
	fmt.Println(mercuryTransit.InternalEnd)
	fmt.Println(mercuryTransit.End)
	fmt.Println(mercuryTransit.Duration)
	fmt.Println(mercuryTransit.MinimumSeparationArcsec)
	fmt.Println(mercuryTransit.SunSemidiameterArcsec)
	fmt.Println(mercuryTransit.PlanetSemidiameterArcsec)

	// Next geocentric Venus transit after the beginning of 2012.
	venusTransit := venus.NextTransit(time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC))
	fmt.Println(venusTransit.Valid)
	fmt.Println(venusTransit.Start)
	fmt.Println(venusTransit.InternalStart)
	fmt.Println(venusTransit.Greatest)
	fmt.Println(venusTransit.InternalEnd)
	fmt.Println(venusTransit.End)
	fmt.Println(venusTransit.Duration)
}
```

Output:

```text
true // a valid geocentric Mercury transit was found
2019-11-11 12:35:31.142292916 +0000 UTC // first contact: Mercury externally enters the solar disk
2019-11-11 12:37:12.603219151 +0000 UTC // second contact: Mercury is fully inside the solar disk
2019-11-11 15:19:48.146160542 +0000 UTC // greatest transit: Mercury center is closest to the Sun center
2019-11-11 18:02:28.751637339 +0000 UTC // third contact: Mercury starts leaving the solar disk
2019-11-11 18:04:10.212523341 +0000 UTC // fourth contact: Mercury externally leaves the solar disk
5h28m39.070230425s // geocentric transit duration from first to fourth contact
75.93592092696468 // minimum Mercury-Sun center separation at greatest transit, arcseconds
968.8881399120452 // solar semidiameter at greatest transit, arcseconds
4.978442587642365 // Mercury semidiameter at greatest transit, arcseconds
true // a valid geocentric Venus transit was found
2012-06-05 22:09:48.690456748 +0000 UTC // first contact: Venus externally enters the solar disk
2012-06-05 22:27:37.088886201 +0000 UTC // second contact: Venus is fully inside the solar disk
2012-06-06 01:29:36.58495903 +0000 UTC // greatest transit: Venus center is closest to the Sun center
2012-06-06 04:31:35.87008953 +0000 UTC // third contact: Venus starts leaving the solar disk
2012-06-06 04:49:24.268559217 +0000 UTC // fourth contact: Venus externally leaves the solar disk
6h39m35.578102469s // geocentric transit duration from first to fourth contact
```

### Outer planets

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/jupiter"
	"github.com/starainrt/astro/mars"
	"github.com/starainrt/astro/neptune"
	"github.com/starainrt/astro/saturn"
	"github.com/starainrt/astro/uranus"
)

func main() {
	// Xi'an, China. Longitude east and latitude north are positive; elevation is 0 m.
	var lon, lat, height float64 = 108.93, 34.27, 0
	cst := time.FixedZone("CST", 8*3600)
	// Instant of observation.
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)

	// Next opposition of Mars.
	fmt.Println(mars.NextOpposition(date))
	// Next conjunction of Jupiter.
	fmt.Println(jupiter.NextConjunction(date))
	// Previous Saturn station from prograde to retrograde.
	fmt.Println(saturn.LastProgradeToRetrograde(date))
	// Saturn ring observing parameters.
	ring := saturn.Ring(date)
	fmt.Printf("saturn B=%.6f Bp=%.6f P=%.6f dU=%.6f major=%.6f minor=%.6f\n",
		ring.EarthLatitude,
		ring.SunLatitude,
		ring.PositionAngle,
		ring.DeltaU,
		ring.MajorAxis,
		ring.MinorAxis,
	)
	// Next Uranus station from retrograde to prograde.
	fmt.Println(uranus.NextRetrogradeToPrograde(date))
	// Previous eastern quadrature of Neptune.
	fmt.Println(neptune.LastEasternQuadrature(date))
	// Next western quadrature of Mars.
	fmt.Println(mars.NextWesternQuadrature(date))
	// Mars rise and set times in Xi'an.
	fmt.Println(mars.RiseTime(date, lon, lat, height, true))
	fmt.Println(mars.SetTime(date, lon, lat, height, true))
	// Current apparent magnitude of Mars.
	fmt.Println(mars.ApparentMagnitude(date))
	// Earth-Mars distance.
	fmt.Println(mars.EarthDistance(date))
	// Sun-Mars distance.
	fmt.Println(mars.SunDistance(date))
}
```

Output:

```text
2020-10-14 07:25:53.541214764 +0800 CST // next opposition of Mars
2021-01-29 09:39:32.556620836 +0800 CST // next conjunction of Jupiter
2019-04-30 10:28:17.254818677 +0800 CST // previous Saturn station from prograde to retrograde
saturn B=23.577010 Bp=23.266930 P=6.629810 dU=1.171003 major=34.133617 minor=13.652809 // Saturn ring B, B', P, dU, major axis, minor axis
2020-01-11 15:22:18.532456755 +0800 CST // next Uranus station from retrograde to prograde
2019-12-08 17:00:04.494322836 +0800 CST // previous eastern quadrature of Neptune
2020-06-07 03:10:55.059157311 +0800 CST // next western quadrature of Mars
2020-01-01 04:41:30.157055854 +0800 CST <nil> // Mars rise time in Xi'an; no error
2020-01-01 14:55:32.792302966 +0800 CST <nil> // Mars set time in Xi'an; no error
1.57 // Mars apparent magnitude
2.1844226019239925 // Earth-Mars distance, AU
1.5897858533008575 // Sun-Mars distance, AU
```

`saturn.Ring` returns `RingInfo`: `EarthLatitude` is ring opening angle B, `SunLatitude` is B', `PositionAngle` is the position angle of the northern semiminor axis, `DeltaU` is the Saturnicentric longitude difference between the Sun and Earth in the ring plane, and `MajorAxis` / `MinorAxis` are the apparent outer major/minor axes in arcseconds.

## Topic examples

The snippets below are grouped by topic and all build on the shared variables `date`, `lon`, `lat`, and `height`; complete runnable versions are in the basic examples above.

### Position and coordinates

`ApparentLo` / `ApparentBo` / `ApparentRa` / `ApparentDec` / `ApparentRaDec` return a **geocentric apparent place**: the geometric geocentric position corrected for light time, aberration, and nutation, then converted to the true equator and equinox of date.

These packages do not provide J2000 or mean-place output; for a J2000 frame, use `coord.Precess` for precession and `coord.Nutation2000B` for nutation, and use `coord.TopocentricEquatorial` and `coord.EquatorialToHorizontal` for topocentric quantities. See [Coordinate Tools](coord.md).

```go
// Apparent place of date: ecliptic and equatorial coordinates, degrees.
lo, bo := venus.ApparentLo(date), venus.ApparentBo(date)
ra, dec := venus.ApparentRaDec(date)
fmt.Println(lo, bo, venus.ApparentRa(date), venus.ApparentDec(date), ra, dec)

// Earth distance and Sun distance, AU.
fmt.Println(venus.EarthDistance(date), venus.SunDistance(date))
```

### Rise, set and culmination

`RiseTime` / `SetTime` / `DownTime` work on the **local civil day** only: `date` selects the local date and the output time zone, and the returned events belong to that local civil day; `height` is the ellipsoidal height in meters, and `aero == true` adds standard atmospheric refraction (lowering the geometric horizon to about `-0.5667°`).

Polar day, polar night, or a day with no crossing returns a [sentinel error](#parameter-and-result-conventions) instead of an instant. `CulminationTime` gives upper culmination; `Altitude` / `Azimuth` / `Zenith` are topocentric instantaneous quantities and `HourAngle` / `ParallacticAngle` geocentric ones, with geometry under [Coordinate Tools](coord.md).

```go
// Rise, set, and culmination on the local civil day in Xi'an; aero=true adds refraction.
rise, err := mars.RiseTime(date, lon, lat, height, true)
set, err := mars.SetTime(date, lon, lat, height, true)
fmt.Println(rise, set, err)
fmt.Println(mars.CulminationTime(date, lon))
```

```go
// Topocentric horizontal quantities: east-positive longitude, north-positive latitude, degrees.
fmt.Println(mars.Altitude(date, lon, lat), mars.Azimuth(date, lon, lat))
fmt.Println(mars.Zenith(date, lon, lat), mars.HourAngle(date, lon))
fmt.Println(mars.ParallacticAngle(date, lon, lat))
```

```go
// DownTime is a compatibility alias for SetTime; the N variant evaluates the `n` variant.
set, err := mars.DownTime(date, lon, lat, height, true)
fmt.Println(set, err)
fmt.Println(mars.CulminationTimeN(date, lon, 12))
```

### Conjunctions, oppositions, stations and quadratures

Event searches always return the nearest event at or before/after the input instant (endpoints included), in the input time zone. The inner planets use `LastConjunction` / `NextConjunction` plus the superior/inferior conjunction family; the outer planets use `LastConjunction` / `NextConjunction`, `LastOpposition` / `NextOpposition`, and the eastern/western quadrature family.

Stations come in two directions: `LastProgradeToRetrograde` / `NextProgradeToRetrograde` (prograde to retrograde) and `LastRetrogradeToPrograde` / `NextRetrogradeToPrograde` (retrograde to prograde); Mercury and Venus additionally have `LastRetrograde` / `NextRetrograde`, which ignore the direction, and the outer planets do not have those two names.

For the inner planets a conjunction is simply "on the same side as the Sun"; the outer planets additionally split opposition and quadrature: only the five planets outside Earth's orbit can reach opposition (Sun-Earth-planet in a line) or eastern/western quadrature (about 90° of ecliptic longitude from the Sun).

Mercury and Venus stay inside Earth's orbit, so those two families exist only in the outer-planet packages.

```go
// Inner planets: superior/inferior conjunction, and direction-agnostic stations.
fmt.Println(mercury.LastSuperiorConjunction(date), mercury.NextInferiorConjunction(date))
fmt.Println(mercury.LastRetrograde(date), mercury.NextRetrograde(date))
```

```go
// Outer planets: conjunction, opposition, and eastern/western quadrature.
fmt.Println(jupiter.NextConjunction(date), mars.NextOpposition(date))
fmt.Println(mars.NextEasternQuadrature(date), neptune.LastWesternQuadrature(date))
```

```go
// Both station directions have Last/Next forms.
fmt.Println(saturn.LastProgradeToRetrograde(date), saturn.NextProgradeToRetrograde(date))
fmt.Println(saturn.LastRetrogradeToPrograde(date), saturn.NextRetrogradeToPrograde(date))
```

### Greatest elongation and transits

Greatest elongation only makes sense for Mercury and Venus: `LastGreatestElongation` / `NextGreatestElongation` ignore the side, while `LastGreatestElongationEast` / `NextGreatestElongationEast` and the `...West` forms distinguish it.

The public transit entry points are likewise only in `mercury` / `venus`: `LastTransit` / `NextTransit` / `ClosestTransit` return a `TransitInfo`.

`Valid == false` means no transit inside the search window and every other field is a zero value. Without geocentric contacts `HasExternal` is false and `Start` / `End` are zero; the event is then accepted or rejected by topocentric surface visibility. Without a topocentric internal contact `HasInternal` is false and `InternalStart` / `InternalEnd` are zero.

Contact fields come from the geocentric geometry, but whether a site can see the transit, and which contacts it sees, is decided by `LocalTransit`.

```go
// Greatest elongation: use ...East / ...West to pick a side, or the plain form for either.
fmt.Println(mercury.NextGreatestElongationEast(date), venus.LastGreatestElongationWest(date))
fmt.Println(venus.NextGreatestElongation(date))
```

```go
// Geocentric transit: when Valid is false every other field is a zero value.
transit := mercury.NextTransit(date)
if transit.Valid {
	fmt.Println(transit.Start, transit.InternalStart, transit.Greatest, transit.InternalEnd, transit.End)
	fmt.Println(transit.Duration, transit.MinimumSeparationArcsec, transit.SunSemidiameterArcsec)
}
```

```go
// Remaining transit fields; InternalDuration is 0 when there are no internal contacts.
transit := venus.ClosestTransit(date)
fmt.Println(transit.HasInternal, transit.InternalDuration, transit.MinimumSeparationArcsec)
fmt.Println(transit.SunSemidiameterArcsec, transit.PlanetSemidiameterArcsec)
```

### Nodes, phase, magnitude, diameter and parallactic angle

`AscendingNode` / `DescendingNode` are the ecliptic longitudes of the two intersections of the planet's orbital plane with the ecliptic, in degrees, about `180°` apart at the same instant.

`PhaseAngle` is the Sun-planet-Earth angle in degrees; `IlluminatedFraction` (alias `Phase`) is the illuminated fraction from `0` to `1`; `BrightLimbPositionAngle` is the position angle of the bright-limb center in degrees.

`Diameter` / `Semidiameter` return the geocentric apparent diameter and semidiameter in arcseconds. `ParallacticAngle` returns the geocentric parallactic angle (zenith direction angle) in degrees, computed from the geocentric hour angle and declination without topocentric parallax; topocentric coordinates are under [Coordinate Tools](coord.md).

```go
fmt.Println(mars.AscendingNode(date), mars.DescendingNode(date))                     // node longitudes, degrees
fmt.Println(mars.PhaseAngle(date), mars.IlluminatedFraction(date), mars.Phase(date)) // phase angle in degrees; illuminated fraction
fmt.Println(mars.ApparentMagnitude(date), mars.BrightLimbPositionAngle(date))        // apparent magnitude; bright-limb position angle
```

```go
fmt.Println(mars.Diameter(date), mars.Semidiameter(date)) // apparent diameter and semidiameter, arcseconds
fmt.Println(mars.ParallacticAngle(date, lon, lat))        // parallactic angle, degrees
```

```go
// Nodes and apparent diameter also have `n` variants.
fmt.Println(mars.AscendingNodeN(date, 12), mars.DescendingNodeN(date, 12))
fmt.Println(mars.DiameterN(date, 12), mars.SemidiameterN(date, 12))
```

### Division of labour with other manuals

- Sidereal time, precession and nutation, and topocentric/horizontal conversion are in [Coordinate Tools](coord.md).
- Sun and Moon positions, phases, rise/set, and syzygies are in the [Sun and Moon manual](sun-moon.md).
- Planetary occultations, occultation bands, and lunar-disk charts are in [Lunar Occultations](occultation.md#lunar-occultation-charts).
- Asteroids, comets, and other bodies computed from orbital elements are in [Small-body orbits](orbit.md).
- Time-scale declarations, UT1 conventions, and GeoJSON output are in [Event Maps and GeoJSON](map-geojson.md); the scale convention itself is in [Time Scale Declaration](map-geojson.md#time-scale-declaration).
- The authoritative capability list, dependency matrix, and accuracy summary are in the root [README](../../../README.en.md).

### Physical ephemerides

All seven major planets provide `Physical` / `PhysicalN` for disk orientation, sub-Earth/sub-Sun coordinates, and north-pole position angle. Jupiter additionally exposes System I/II/III central meridians, and Saturn exposes ring parameters.

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/jupiter"
	"github.com/starainrt/astro/saturn"
)

func main() {
	date := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)

	// Jupiter: DS and DE are planetocentric declinations of the Sun and Earth relative to Jupiter's equator.
	// CMI/CMII/CMIII are Jupiter System I/II/III central meridians, degrees.
	j := jupiter.Physical(date)
	fmt.Printf("jupiter DS=%.6f DE=%.6f CMI=%.6f CMII=%.6f CMIII=%.6f\n",
		j.DS,
		j.DE,
		j.CentralMeridianSystemI,
		j.CentralMeridianSystemII,
		j.CentralMeridianSystemIII,
	)

	// Saturn ring: B/B' are ring-plane latitudes seen from Earth and Sun; P is the position angle of the ring minor axis.
	ring := saturn.Ring(date)
	fmt.Printf("saturn B=%.6f Bp=%.6f P=%.6f major=%.6f minor=%.6f\n",
		ring.EarthLatitude,
		ring.SunLatitude,
		ring.PositionAngle,
		ring.MajorAxis,
		ring.MinorAxis,
	)
}
```

Output:

```text
jupiter DS=54.342201 DE=1.436449 CMI=292.712946 CMII=276.309085 CMIII=147.241848 // Jupiter DS/DE and System I/II/III central meridians, degrees
saturn B=-0.608032 Bp=-2.675667 P=4.480277 major=42.709152 minor=0.453228 // Saturn ring B, B', minor-axis position angle, outer major/minor axes
```

If only Jupiter central meridians are needed:

```go
cm := jupiter.CentralMeridians(date)
fmt.Printf("CMI=%.6f CMII=%.6f CMIII=%.6f\n", cm.SystemI, cm.SystemII, cm.SystemIII) // Jupiter System I/II/III central meridians
```

Saturn and Uranus also retain explicit `System III` semantic aliases:

```go
sat3 := saturn.PhysicalSystemIII(date)
ura3 := uranus.PhysicalSystemIII(date)
fmt.Printf("saturn systemIII lon=%.6f lat=%.6f P=%.6f\n", sat3.SubEarthLongitude, sat3.SubEarthLatitude, sat3.NorthPolePositionAngle) // Saturn sub-Earth longitude/latitude and north-pole position angle
fmt.Printf("uranus systemIII lon=%.6f lat=%.6f P=%.6f\n", ura3.SubEarthLongitude, ura3.SubEarthLatitude, ura3.NorthPolePositionAngle) // Uranus sub-Earth longitude/latitude and north-pole position angle
```

All seven packages return their own `PhysicalInfo`, whose fields mirror `basic.PlanetPhysicalInfo` one for one.

The positive direction of `SubEarthLongitude` / `SubSolarLongitude` follows each body's current IAU/Horizons cartographic convention: Mercury, Mars, Jupiter, Saturn, and Neptune use west-positive longitudes, while Venus and Uranus use east-positive longitudes.

Saturn-ring parameters describe only the disk and band; they take no part in occultation contact geometry, and charts are in [Lunar Occultations](occultation.md#lunar-occultation-charts).

```go
p := uranus.Physical(date) // Uranus sub-Earth/sub-Solar coordinates and north-pole position angle, degrees
fmt.Println(p.SubEarthLongitude, p.SubEarthLatitude, p.SubSolarLongitude, p.SubSolarLatitude, p.NorthPolePositionAngle)
```

```go
// Truncated variants of the Saturn ring and the Jupiter central meridians.
fmt.Println(saturn.RingN(date, 12).MinorAxis, jupiter.CentralMeridiansN(date, 12).SystemIII)
```

### Galilean satellites of Jupiter

The public entry points are in the **`jupiter` package** (`jupiter.Satellites`, `jupiter.SatellitePhenomena`, `jupiter.NextGalileanPhenomenonEvent`, and so on); they take a `time.Time` and return Jupiter's own types.

The `JupiterGalilean*` functions in `basic` (such as `basic.JupiterGalileanSatelliteObservations` and `basic.NextJupiterGalileanPhenomenonEvent`) are the low-level entries of the same implementation: they take a Julian day and return `basic` types.

Applications should use the `jupiter` layer.

```go
// Entry point is in the jupiter package; satellite numbers use jupiter.GalileanSatelliteIo and friends.
sats := jupiter.Satellites(date)
fmt.Println(sats.Io.OffsetXJupiterR, sats.Io.OffsetYJupiterR, sats.Io.InFrontOfJupiter)
```

Satellite numbers and phenomenon types have named constants, so there is no need to write raw numbers or strings:

- Satellite numbers: `GalileanSatelliteIo`, `GalileanSatelliteEuropa`, `GalileanSatelliteGanymede`, `GalileanSatelliteCallisto`
- Phenomenon types (`GalileanPhenomenonType`): `GalileanPhenomenonTransit`, `GalileanPhenomenonOccultation`, `GalileanPhenomenonEclipse`, `GalileanPhenomenonShadowTransit`
- Contact phases (`GalileanPhenomenonContactPhase`): `GalileanPhenomenonContactDisappearance`, `GalileanPhenomenonContactReappearance`

```go
// Satellite numbers and phenomenon types are constants, so no raw numbers or strings are needed.
event := jupiter.NextGalileanPhenomenonEvent(date, jupiter.GalileanSatelliteCallisto, jupiter.GalileanPhenomenonShadowTransit)
fmt.Println(event.Type == jupiter.GalileanPhenomenonShadowTransit)
```

The `jupiter` package provides apparent positions, instantaneous phenomena, and event searches for the four Galilean satellites.

Common entry points:

- `Satellites`: instantaneous apparent positions relative to Jupiter's disk
- `SatellitePhenomena`: instantaneous transit, occultation, eclipse, and shadow-transit flags
- `LastGalileanPhenomenonEvent` / `NextGalileanPhenomenonEvent` / `ClosestGalileanPhenomenonEvent`: search whole phenomenon intervals
- `LastGalileanPhenomenonContactEvent` / `NextGalileanPhenomenonContactEvent` / `ClosestGalileanPhenomenonContactEvent`: search IMCCE-style D/F contact events

Two conventions matter:

- `GalileanPhenomenonEvent` treats the satellite as a point and checks when its center enters or leaves Jupiter's disk. It is suitable for fast phenomenon search and internal state checks.
- `GalileanPhenomenonContactEvent` includes the finite disk of the satellite and splits disappearance and reappearance contact windows. It is the better match for IMCCE tables such as `TR.D/TR.F/OC.D/OC.F/EC.D/EC.F/SH.D/SH.F`.

The two conventions may differ by up to about 7 minutes in duration. This is a definition difference, not a timing-accuracy failure. Use `GalileanPhenomenonContactEvent` for observing predictions and direct comparison with public almanacs.

#### Code example

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/jupiter"
)

func main() {
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// Instantaneous positions of the four satellites relative to Jupiter's center.
	sats := jupiter.Satellites(date)
	fmt.Printf("io x=%.6f y=%.6f front=%v\n", sats.Io.OffsetXJupiterR, sats.Io.OffsetYJupiterR, sats.Io.InFrontOfJupiter)
	fmt.Printf("europa ra=%.6f dec=%.6f\n", sats.Europa.ApparentRA, sats.Europa.ApparentDec)

	// Instantaneous phenomenon flags.
	ph := jupiter.SatellitePhenomena(date)
	fmt.Printf("io transit=%v occultation=%v eclipse=%v shadow=%v\n", ph.Io.Transit, ph.Io.Occultation, ph.Io.Eclipse, ph.Io.ShadowTransit)
	fmt.Printf("europa transit=%v occultation=%v eclipse=%v shadow=%v\n", ph.Europa.Transit, ph.Europa.Occultation, ph.Europa.Eclipse, ph.Europa.ShadowTransit)

	// Next full Io transit event.
	event := jupiter.NextGalileanPhenomenonEvent(date, jupiter.GalileanSatelliteIo, jupiter.GalileanPhenomenonTransit)
	fmt.Printf("event valid=%v sat=%d type=%s\n", event.Valid, event.Satellite, event.Type)
	fmt.Println(event.Start)
	fmt.Println(event.Greatest)
	fmt.Println(event.End)
	fmt.Println(event.Duration)

	// Next IMCCE-style contact window for a Europa occultation.
	contact := jupiter.NextGalileanPhenomenonContactEvent(date, jupiter.GalileanSatelliteEuropa, jupiter.GalileanPhenomenonOccultation)
	fmt.Printf("contact valid=%v sat=%d type=%s\n", contact.Valid, contact.Satellite, contact.Type)
	fmt.Println(contact.Disappearance.Start)
	fmt.Println(contact.Disappearance.ModelCrossing)
	fmt.Println(contact.Disappearance.End)
	fmt.Println(contact.Greatest)
	fmt.Println(contact.Reappearance.Start)
	fmt.Println(contact.Reappearance.ModelCrossing)
	fmt.Println(contact.Reappearance.End)
}
```

Output:

```text
io x=-0.675023 y=-0.032794 front=true // Io X/Y offset from Jupiter center, in Jupiter radii; in front of Jupiter
europa ra=110.769095 dec=22.335877 // Europa apparent RA and Dec, degrees
io transit=true occultation=false eclipse=false shadow=true // Io is transiting, and its shadow is also transiting
europa transit=false occultation=false eclipse=false shadow=false // Europa has no transit, occultation, eclipse, or shadow transit at this instant
event valid=true sat=1 type=transit // next valid event is an Io transit
2026-01-16 16:32:47.552742362 +0000 UTC // Io transit begins
2026-01-16 17:40:44.189371168 +0000 UTC // midpoint of the Io transit
2026-01-16 18:48:40.287077128 +0000 UTC // Io transit ends
2h15m52.734334766s // Io transit duration
contact valid=true sat=2 type=occultation // next valid contact event is a Europa occultation
2026-01-17 01:00:34.99533087 +0000 UTC // Europa occultation disappearance starts
2026-01-17 01:02:31.714070141 +0000 UTC // model center crossing during disappearance
2026-01-17 01:04:28.432809412 +0000 UTC // disappearance ends
2026-01-17 02:27:37.807798683 +0000 UTC // deepest occultation
2026-01-17 03:50:48.120300471 +0000 UTC // reappearance starts
2026-01-17 03:52:43.901527225 +0000 UTC // model center crossing during reappearance
2026-01-17 03:54:39.68275398 +0000 UTC // reappearance ends
```

#### External baselines

The Galilean-satellite implementation was checked against two external baselines:

- **JPL Horizons**: apparent positions of the four satellites relative to Jupiter's center, and shadow-center offsets from Jupiter's disk during shadow transits.
- **IMCCE 2026 tables**: transits, occultations, Jupiter eclipses, shadow transits, and D/F contact windows.

Comparison results:

- `Satellites` positions relative to Jupiter center: maximum sample difference against JPL Horizons about `X=0.054"`, `Y=0.048"`.
- `SatellitePhenomena` shadow-transit shadow-center offsets: maximum sample difference against JPL Horizons about `X=0.051"`, `Y=0.016"`; boolean phenomenon flags match in the samples.
- `GalileanPhenomenonContactEvent` differs from IMCCE 2026 D/F contact times by at most about `79 s`, and contact durations by about `17 s` in the compared cases.
- `GalileanPhenomenonEvent` uses a different definition from IMCCE D/F contacts, so start and end times can differ by about `7 min`.

### Shared low-level entry points in the `planet` package

The `planet` package provides low-level entry points to the built-in models and series, shared by the seven planet packages and by `sun` / `moon`.

It exports functions only, not types. The planet packages share numerical conventions and model choices rather than result types: each package's `PhysicalInfo` mirrors `basic.PlanetPhysicalInfo` and `TransitInfo` mirrors `basic.PlanetTransitResult`, field for field, while the types themselves still belong to their own packages. External kernel-engine types live in the separate `planet/spk` subpackage.

| Entry point | Purpose | Unit |
| --- | --- | --- |
| `planet.WherePlanet` / `planet.WherePlanetN` | Selected built-in model: `xt` is `1..7` for Mercury through Neptune and `-1` or `0` for Earth, `zn` is `0` longitude, `1` latitude, `2` heliocentric distance; an out-of-range `xt` / `zn` returns `NaN` instead of panicking | degrees / AU |
| `planet.Distance` | Sun-Earth distance | AU |
| `planet.SunLo` / `planet.SunM` / `planet.SunMidFun` / `planet.SunTrueLo` / `planet.SunApparentLo` | Solar geometric longitude, mean anomaly, equation of center, true longitude, apparent longitude | degrees |
| `planet.Earthe` / `planet.EarthPI` | Earth's orbital eccentricity and longitude of perihelion | dimensionless / degrees |
| `planet.MoonLo` / `planet.MoonM` / `planet.MoonLonX` / `planet.SunMoonAngle` | Lunar mean longitude, mean anomaly, mean argument of latitude, and mean elongation | degrees |
| `planet.MoonI` / `planet.MoonB` / `planet.MoonR` | Periodic longitude, latitude, and distance terms (truncated ELP2000/82-style series) | `10⁻⁶` degrees / `10⁻⁶` degrees / `10⁻³ km` |
| `planet.MoonTrueLo` / `planet.MoonTrueBo` / `planet.MoonAway` | Lunar true longitude, true latitude, and Earth distance | degrees / degrees / km |

```go
// xt=1..7 is Mercury..Neptune; zn=0 longitude, 1 latitude, 2 heliocentric distance (AU).
fmt.Println(planet.WherePlanet(4, 2, 2460000.5))      // Jupiter heliocentric distance, AU
fmt.Println(planet.WherePlanetN(4, 2, 2460000.5, 12)) // semianalytic fast table; Legacy truncates according to n
fmt.Println(planet.WherePlanet(8, 0, 2460000.5))      // xt out of range: NaN
```

```go
// Earth's heliocentric longitude (xt=-1) and a planet's can be compared on the same convention.
fmt.Println(planet.WherePlanet(-1, 0, 2460000.5), planet.WherePlanet(4, 0, 2460000.5))
```

## Parameter and result conventions

The conventions below are shared by all seven packages; per-capability return units are summarised under [Common capabilities and units](#common-capabilities-and-units).

### Time scale and civil time

Every public API treats its `time.Time` argument as a **civil instant** (a UTC label): position and physical functions take `date.UTC()` and convert to TT internally before evaluating the ephemeris, while rise/set, culmination, and the topocentric horizontal family additionally read `date.Zone()` for local-time computation.

UTC had no leap seconds before 1972-01-01; the library treats that span as UT1, so a civil-time label there is equal to UT1.

From 1972 on the built-in leap-second table is used, and beyond the exact window the active UTC tracking policy applies. For explicit conversion use `astro.UT1FromUTC`, `astro.TTFromUTC`, and `astro.DUT1` from the root package; time-scale declarations inside figures and captions and the UT1 convention are in [Time Scale Declaration](map-geojson.md#time-scale-declaration).

### Units and conventions

Angles are always in degrees; apparent diameter and semidiameter are in arcseconds; distances follow the function name -- `EarthDistance` / `SunDistance` and `planet.WherePlanet` with `zn=2` are in AU, while `planet.MoonAway` is in km.

`PhaseAngle` is in degrees, `IlluminatedFraction` and its alias `Phase` are `0-1`, and `ApparentMagnitude` is a magnitude (dimensionless, classic phase polynomials); `ApparentMagnitude2018` uses Mallama & Hilton 2018, better for Mercury/Venus at large phase and Saturn's rings, without the Mars rotation and season terms; rise/set, culmination, and every event search return `time.Time` in the input time zone.

Duration fields inside event structs (`TransitInfo.Duration`, `TransitInfo.InternalDuration`, `GalileanPhenomenonEvent.Duration`, `GalileanPhenomenonContact.Duration`) are Go `time.Duration` values, not Julian days or day counts; the `Start` / `Greatest` / `End` / `InternalStart` / `InternalEnd` fields of `TransitInfo` keep the caller's time zone.

Geocentric, topocentric, and distance are three different conventions: `ApparentLo` / `ApparentBo` / `ApparentRa` / `ApparentDec` / `ApparentRaDec` are geocentric apparent places, `Altitude` / `Azimuth` / `Zenith` are topocentric quantities, `HourAngle` / `ParallacticAngle` are geocentric quantities without topocentric parallax, and `EarthDistance` / `SunDistance` are geometric geocentric distances.

### Zero values, out-of-range input and sentinel errors

When an event search finds nothing inside its window it returns zero values or a struct with `Valid == false` (`TransitInfo.Valid`, `GalileanPhenomenonEvent.Valid`, `GalileanPhenomenonContactEvent.Valid`) instead of an error; when `TransitInfo.HasExternal` is false, `Start` / `End` are zero values, and when `TransitInfo.HasInternal` is false, `InternalStart` / `InternalEnd` are zero values.

`planet.WherePlanet` / `planet.WherePlanetN` return `NaN` instead of panicking when `xt` / `zn` is out of range. `n < 0` selects the full series; all nonnegative `n` select one semianalytic fast table, while Legacy truncates according to `n`.

`RiseTime` / `SetTime` / `DownTime` (including their `N` forms) are the only family that returns an `error`: when the body has no geometric rise or set on that day they return the sentinel error below with a zero instant.

Each planet package has its own polar-day/polar-night errors.

`NEVER_RISE` in the name means "never rises that day" (polar night), `NEVER_SET` means "never sets that day" (polar day), and `NEVER_DOWN` is a compatibility alias for `NEVER_SET`.

| Package | Never rises | Never sets | Set alias |
| --- | --- | --- | --- |
| `mercury` | `ERR_MERCURY_NEVER_RISE` | `ERR_MERCURY_NEVER_SET` | `ERR_MERCURY_NEVER_DOWN` |
| `venus` | `ERR_VENUS_NEVER_RISE` | `ERR_VENUS_NEVER_SET` | `ERR_VENUS_NEVER_DOWN` |
| `mars` | `ERR_MARS_NEVER_RISE` | `ERR_MARS_NEVER_SET` | `ERR_MARS_NEVER_DOWN` |
| `jupiter` | `ERR_JUPITER_NEVER_RISE` | `ERR_JUPITER_NEVER_SET` | `ERR_JUPITER_NEVER_DOWN` |
| `saturn` | `ERR_SATURN_NEVER_RISE` | `ERR_SATURN_NEVER_SET` | `ERR_SATURN_NEVER_DOWN` |
| `uranus` | `ERR_URANUS_NEVER_RISE` | `ERR_URANUS_NEVER_SET` | `ERR_URANUS_NEVER_DOWN` |
| `neptune` | `ERR_NEPTUNE_NEVER_RISE` | `ERR_NEPTUNE_NEVER_SET` | `ERR_NEPTUNE_NEVER_DOWN` |

```go
// Polar day/night: the rise/set functions return a sentinel error and a zero time.Time.
rise, err := mercury.RiseTime(date, lon, lat, height, true)
switch {
case errors.Is(err, mercury.ERR_MERCURY_NEVER_RISE):
	fmt.Println("polar night: never rises that day", rise.IsZero())
case errors.Is(err, mercury.ERR_MERCURY_NEVER_SET):
	fmt.Println("polar day: never sets that day", rise.IsZero())
}
```

### Topocentric versus geocentric

`Altitude` / `Azimuth` / `Zenith` / `HourAngle` / `ParallacticAngle` use `date`'s time zone for local-time computation, with east-positive longitude, north-positive latitude, and `height` as the ellipsoidal height in meters (not orthometric elevation).

`Altitude` / `Azimuth` / `Zenith` include topocentric parallax and are a different convention from the geocentric apparent place of `ApparentRa` / `ApparentDec`; `HourAngle` / `ParallacticAngle` come straight from the geocentric apparent place without topocentric parallax. Convert through [Coordinate Tools](coord.md) before mixing them.

### Accuracy and scope

The planet packages default to the semianalytic model, with Legacy or custom engines selectable explicitly. An available range does not imply uniform accuracy across it; Legacy truncation error against complete VSOP87 is not total error against DE ephemerides. See [Legacy truncation error](accuracy.md#legacy-truncation-error).

The fast table suits coarse scans. Use the full series and matching time/frame conventions for event refinement or comparisons with external tables.

`saturn.Ring` and the physical ephemerides affect only the disk and band rendering: the Saturn ring takes no part in occultation contact geometry and is not drawn as a disk boundary. The planets have no dedicated chart entry point; occultation and Saturn-ring charts are in [Lunar Occultations](occultation.md#lunar-occultation-charts).
