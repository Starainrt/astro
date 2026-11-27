# Coordinates

[中文](../coord.md) | [Back to README](../../../README.en.md)

The `coord` package handles celestial coordinate conversions and observing helpers: ecliptic, equatorial, horizontal, galactic, and planetary-equatorial transforms, topocentric parallax, sidereal time, precession & nutation, angular separation, atmospheric refraction, airmass, and parallactic angle.

## Contents

- [Basic Example](#basic-example)
- [Usage](#usage)
  - [Refraction & Airmass](#refraction--airmass)
  - [Topocentric Parallax, Precession & Galactic Coordinates](#topocentric-parallax-precession--galactic-coordinates)
  - [Planetary Equatorial Coordinates](#planetary-equatorial-coordinates)
  - [Observer Height](#observer-height)
- [API Reference](#api-reference)
  - [Airmass](#airmass)
- [Conventions and Limitations](#conventions-and-limitations)

## Basic Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/coord"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2026, 4, 27, 10, 30, 45, 0, cst)

	eq := coord.EclipticToEquatorial(date, 139.686111, 4.875278)
	hz := coord.EquatorialToHorizontal(date, eq.RA, eq.Dec, 115, 40)

	fmt.Printf("Equatorial: RA=%.4f° Dec=%.4f°  Horizontal: Az=%.2f° Alt=%.2f°  GAST=%.4f h\n",
		eq.RA, eq.Dec, hz.Azimuth, hz.Altitude, coord.ApparentSiderealTime(date))
}
```

Output:

```text
Equatorial: RA=143.7222° Dec=19.5351°  Horizontal: Az=43.47° Alt=-17.69°  GAST=16.8526 h
```

> Angles are in **degrees**, sidereal time in **hours**, and height in **metres**. Longitude is east-positive; azimuth is clockwise from north (`0°`).

## Usage

### Refraction & Airmass

- Convert true ↔ apparent altitude with `ApparentAltitude(alt, pressureHPa, tempC)` and `TrueAltitude`, or jump directly from equatorial coordinates with `EquatorialToApparentHorizontal`.
- Compute airmass directly from true altitude (applying refraction first) via `AirmassKastenYoungFromTrueAltitude`:

```go
appAlt := coord.ApparentAltitude(10, 1010, 10)
ky := coord.AirmassKastenYoungFromTrueAltitude(10, 1010, 10)
fmt.Printf("True 10° -> Apparent %.4f° (refraction +%.2f'), airmass %.3f\n",
	appAlt, (appAlt-10)*60, ky)
```

Output:

```text
True 10° -> Apparent 10.0901° (refraction +5.41'), airmass 5.540
```

### Topocentric Parallax, Precession & Galactic Coordinates

- **Topocentric parallax**: `TopocentricEquatorial(date, ra, dec, lon, lat, distanceAU, heightM)` and `TopocentricEcliptic` (essential for the Moon at `~0.00257 AU`).
- **Precession & nutation**: `Precess(fromDate, toDate, ra, dec)` (precession only, no proper motion), `EclipticObliquity(date, withNutation)`, `Nutation2000B`, `Nutation1980`, and opt-in high-precision `Nutation2000A` / `MeanObliquity2006`.
- **Galactic coordinates**: `EquatorialToGalactic` / `GalacticToEquatorial` apply the fixed ICRS ↔ Galactic rotation matrix.
- **Separation & parallactic angle**: `AngularSeparation(ra1, dec1, ra2, dec2)` and `ParallacticAngle`.

### Planetary Equatorial Coordinates

`ToBodyEquatorial(body, date, ra, dec, origin)` and `FromBodyEquatorial` rotate ICRF/J2000 directions into a planet's equatorial frame (Mercury through Neptune):

```go
t := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
marsEq, err := coord.ToBodyEquatorial(basic.SkyMars, t, 101.287155, -16.716116, coord.BodyEquatorOriginEquinox)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Sirius in Mars equatorial (equinox origin): RA=%.2f°  Dec=%.2f°\n", marsEq.RA, marsEq.Dec)
```

Output:

```text
Sirius in Mars equatorial (equinox origin): RA=174.63°  Dec=-44.02°
```

- `origin`: `BodyEquatorOriginIAUNode` (zero value, ascending node Q on ICRF equator) or `BodyEquatorOriginEquinox` (planet's vernal equinox; Mercury's equinox is degenerate and returns an error—use Q instead).
- This is a pure directional rotation; for horizon altitude/azimuth from a planetary surface, see [Planetary Surface Observations](phenomena.md#planetary-surface-observations).

### Observer Height

Throughout this library (`sun`, `moon`, planets, `star`, `eclipse`, `orbit`, `satellite`, `coord`), observer `height` is WGS84 **ellipsoidal height in metres**, not orthometric elevation above mean sea level (`H`).

If you only have elevation above sea level `H`, add the local geoid undulation `N` (e.g. from EGM2008):

```text
height = H + N
```

A few tens of metres barely affects casual rise/set times, but converting to ellipsoidal height is recommended for precise occultation timing, eclipse edges, and high-altitude sites.

## API Reference

| Group | Entry Points | Notes |
| --- | --- | --- |
| Ecliptic ↔ Equatorial | `EclipticToEquatorial` / `EquatorialToEcliptic` (plus `...ByObliquity`) | Instantaneous true obliquity or custom obliquity |
| Equatorial ↔ Horizontal | `EquatorialToHorizontal`, `EquatorialToApparentHorizontal`, `...ByLocalSiderealTime`, `HourAngleDeclinationToHorizontal`, `HorizontalToHourAngleDeclination` | Azimuth `[0,360)`, altitude `[-90,90]` |
| Topocentric | `TopocentricEquatorial`, `TopocentricEcliptic` | `distanceAU` in AU, `height` in metres |
| Sidereal time | `MeanSiderealTime`, `ApparentSiderealTime`, `HourAngle` | GMST / GAST in hours (IAU 2006) |
| Precession & nutation | `Precess`, `EclipticObliquity`, `Nutation2000B`, `Nutation1980`, `Nutation2000A`, `MeanObliquity2006` | Degrees |
| Galactic & planetary | `EquatorialToGalactic`, `GalacticToEquatorial`, `ToBodyEquatorial`, `FromBodyEquatorial` | Fixed ICRS rotation / planet equatorial frames |
| Separation & PA | `AngularSeparation`, `ParallacticAngle`, `ParallacticAngleByHourAngle` | Great-circle deg; parallactic angle `(-180, 180]` |
| Refraction | `ApparentAltitude`, `TrueAltitude`, `AtmosphericRefractionFromTrueAltitude`, `AtmosphericRefractionFromApparentAltitude` | Saemundsson model, pressure hPa & temp °C |

### Airmass

| Entry Point | Input | Notes |
| --- | --- | --- |
| `AirmassPlaneParallelFromTrueAltitude` | True altitude (deg) | `sec(z)` approximation |
| `AirmassKastenYoungFromTrueAltitude` / `AirmassPickeringFromTrueAltitude` | True altitude + pressure + temp | Applies refraction first |
| `AirmassKastenYoungFromApparentAltitude` / `AirmassPickeringFromApparentAltitude` | Apparent altitude (deg) | No extra refraction (also in [Formulas](formula.md#airmass-models)) |

## Conventions and Limitations

- **Normalization**: `RA`, `Lon`, `Azimuth`, and `HourAngle` normalize to `[0°, 360°)` (except `TopocentricEquatorial`, which adds parallax directly without wrapping `RA`).
- **Time scales**: Sidereal time uses UT1 (IAU 2006 ERA); precession and nutation use TT. See [Time Scales](timescale.md).
- **Refraction limits**: Active for true altitudes in `(-5°, 90°)` and `0` outside; invalid pressure/temperature or unsolvable inverses return `NaN`.
- **Planetary mapping helpers**: `basic.PlanetographicToEastLongitude` / `EastToPlanetographicLongitude` and `PlanetocentricToPlanetodeticLatitude` / `PlanetodeticToPlanetocentricLatitude` convert planetary coordinate conventions.
