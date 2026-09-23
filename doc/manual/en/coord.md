# Coordinate Tools

[中文](../coord.md) | [Back to README](../../../README.en.md)

`coord` provides celestial coordinate conversion and observing helper calculations. Unless noted otherwise, angles are degrees, sidereal time is in hours, and `time.Time` is treated as an absolute instant and internally converted to UTC.

Altitude, zenith distance, and horizon-visibility semantics follow [Observing-angle semantics](sun-moon.md#observing-angle-semantics); full topocentric usage appears in [Occultation · Local event chart](occultation.md#fixed-site-charts) and [Sun and Moon positions](sun-moon.md#sun-and-moon-position);

topocentric solar-eclipse contacts appear in [Solar eclipses](eclipse.md#solar-eclipse).

Time scales, observer height, and overall accuracy are documented in [Time scale conventions](timescale.md), [Observer height conventions](coord.md#observer-height), and [Applicability and accuracy](accuracy.md).

## Contents

- [Converting ecliptic coordinates to equatorial and horizontal](#converting-ecliptic-coordinates-to-equatorial-and-horizontal)
- [API Reference](#api-reference)
  - [Ecliptic ↔ Equatorial](#ecliptic--equatorial)
  - [Equatorial ↔ Horizontal](#equatorial--horizontal)
  - [Topocentric Coordinates](#topocentric-coordinates)
  - [Sidereal Time](#sidereal-time)
  - [Precession and Nutation](#precession-and-nutation)
  - [Galactic Coordinates](#galactic-coordinates)
  - [Angular Separation](#angular-separation)
  - [Atmospheric Refraction](#atmospheric-refraction)
  - [Airmass](#airmass)
  - [Parallactic Angle](#parallactic-angle)
- [Usage examples](#usage-examples)
  - [Round-trip conversion between ecliptic and horizontal coordinates](#round-trip-conversion-between-ecliptic-and-horizontal-coordinates)
  - [Topocentric correction and sidereal time](#topocentric-correction-and-sidereal-time)
  - [Refraction, airmass and parallactic angle](#refraction-airmass-and-parallactic-angle)
  - [Precession, nutation and obliquity conventions](#precession-nutation-and-obliquity-conventions)
- [Research APIs and Observing Helpers](#research-apis-and-observing-helpers)
- [Parameter and result conventions](#parameter-and-result-conventions)
  - [Units](#units)
  - [Time Scales](#time-scales)
  - [Angle Quadrants and Normalization](#angle-quadrants-and-normalization)
  - [Zero Values and Invalid Input](#zero-values-and-invalid-input)
  - [Accuracy and Applicability](#accuracy-and-applicability)
- [Observer height](#observer-height)

## Converting ecliptic coordinates to equatorial and horizontal

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
	fmt.Printf("RA=%.6f Dec=%.6f deg\n", eq.RA, eq.Dec)
	fmt.Printf("azimuth=%.6f altitude=%.6f deg\n", hz.Azimuth, hz.Altitude)
	fmt.Printf("GAST=%.6f h\n", coord.ApparentSiderealTime(date))
}
```

Right ascension, declination and horizontal angles are in degrees; sidereal time is in hours. These equatorial coordinates belong to the observing date. Galactic conversion requires ICRS coordinates, so the two must not be mixed directly.

## API Reference

The tables group functions by the quantity being calculated.

The snippets reuse `date`, `cst` and `eq` from the first example. The sidereal-time example also needs the standard `math` package.

### Ecliptic ↔ Equatorial

| Name | Purpose | Units and conventions |
| --- | --- | --- |
| `EclipticToEquatorial` | ecliptic to equatorial | ecliptic longitude/latitude and right ascension/declination in degrees; uses the true obliquity at that instant, `RA ∈ [0,360)` |
| `EquatorialToEcliptic` | equatorial to ecliptic | inverse of the same obliquity; `Lon ∈ [0,360)`, `Lat ∈ [−90,90]` |
| `EclipticToEquatorialByObliquity` | ecliptic to equatorial with manual obliquity | all three parameters in degrees; for experiments with custom axial tilts, no date conversion |
| `EquatorialToEclipticByObliquity` | equatorial to ecliptic with manual obliquity | same; `Lon ∈ [0,360)`, `Lat ∈ [−90,90]` |
| `Ecliptic` | ecliptic result value | fields `Lon` and `Lat`, in degrees |
| `Equatorial` | equatorial result value | fields `RA` and `Dec`, in degrees |

```go
// date is the shared preamble; take the true obliquity at that instant, then compare with the date-based path.
obliquity := coord.EclipticObliquity(date, true)
auto := coord.EclipticToEquatorial(date, 139.686111, 4.875278)
manual := coord.EclipticToEquatorialByObliquity(139.686111, 4.875278, obliquity)
back := coord.EquatorialToEclipticByObliquity(manual.RA, manual.Dec, obliquity)
fmt.Printf("auto=(%.9f, %.9f) manual=(%.9f, %.9f)\n",
	auto.RA, auto.Dec, manual.RA, manual.Dec)
fmt.Printf("back=(%.9f, %.9f)\n", back.Lon, back.Lat)
```

### Equatorial ↔ Horizontal

| Name | Purpose | Units and conventions |
| --- | --- | --- |
| `EquatorialToHorizontal` | apparent equatorial to horizontal | longitude east-positive, latitude north-positive, degrees; uses apparent sidereal time (with nutation); `Azimuth ∈ [0,360)`, `Altitude ∈ [−90,90]`, `Zenith = 90 − Altitude` |
| `EquatorialToHorizontalByLocalSiderealTime` | equatorial to horizontal with manual sidereal time | sidereal time in hours, multiplied by 15 internally; no date lookup, suitable for manual experiments |
| `HourAngleDeclinationToHorizontal` | hour angle plus declination to horizontal | input hour angle normalized into `[0,360)`; `Azimuth ∈ [0,360)` |
| `HorizontalToHourAngleDeclination` | horizontal to hour angle plus declination | returned hour angle is normalized into `[0,360)` (not `[−180,180]`), declination `[−90,90]` |
| `HorizontalToEquatorialByLocalSiderealTime` | horizontal to equatorial with manual sidereal time | sidereal time in hours; `RA ∈ [0,360)` |
| `Horizontal` | horizontal result value | fields `Azimuth`, `Altitude`, `Zenith`, and `HourAngle`, all in degrees |

```go
// Observer at 115°E, 40°N; eq comes from the shared preamble.
hz := coord.EquatorialToHorizontal(date, eq.RA, eq.Dec, 115, 40)
manual := coord.EquatorialToHorizontalByLocalSiderealTime(10.5, eq.RA, eq.Dec, 40)
hz2 := coord.HourAngleDeclinationToHorizontal(hz.HourAngle, eq.Dec, 40)
ha, dec := coord.HorizontalToHourAngleDeclination(hz2.Azimuth, hz2.Altitude, 40)
eq2 := coord.HorizontalToEquatorialByLocalSiderealTime(10.5, hz2.Azimuth, hz2.Altitude, 40)
fmt.Printf("auto=(%.6f, %.6f, %.6f) manual=(%.6f, %.6f)\n",
	hz.Azimuth, hz.Altitude, hz.Zenith, manual.Azimuth, manual.Altitude)
fmt.Printf("roundtrip ha=%.6f dec=%.6f ra=%.6f\n", ha, dec, eq2.RA)
```

### Topocentric Coordinates

| Name | Purpose | Units and conventions |
| --- | --- | --- |
| `TopocentricEquatorial` | geocentric to topocentric equatorial | `distanceAU` is the geocentric distance in AU; `height` is the ellipsoidal height in meters; `RA` only receives a small parallax correction and is not normalized into 360° |
| `TopocentricEcliptic` | geocentric to topocentric ecliptic | same; it solves the topocentric equatorial position first and converts to the ecliptic, so `Lon` is normalized into `[0,360)` and `Lat` lands in `[−90,90]` |
| `Ecliptic` / `Equatorial` | return value types | same field conventions as the ecliptic ↔ equatorial group |

```go
// Target 0.00257 AU from the geocenter (roughly the Moon), observer at 115°E, 40°N, ellipsoidal height 53 m.
top := coord.TopocentricEquatorial(date, eq.RA, eq.Dec, 115, 40, 0.00257, 53)
topEcl := coord.TopocentricEcliptic(date, 139.686111, 4.875278, 115, 40, 0.00257, 53)
fmt.Printf("top=(%.9f, %.9f) dRA=%.9f dDec=%.9f\n",
	top.RA, top.Dec, top.RA-eq.RA, top.Dec-eq.Dec)
fmt.Printf("topEcl=(%.9f, %.9f)\n", topEcl.Lon, topEcl.Lat)
```

Topocentric geometry drives per-site occultation contact times, see [Local event chart](occultation.md#fixed-site-charts); topocentric quantities in `sun` and `moon` appear under [Sun and Moon positions](sun-moon.md#sun-and-moon-position).

### Sidereal Time

| Name | Purpose | Units and conventions |
| --- | --- | --- |
| `MeanSiderealTime` | Greenwich mean sidereal time | hours; converts civil time to UT1 first, model is IAU 2006 (ERA route) |
| `ApparentSiderealTime` | Greenwich apparent sidereal time | hours; mean sidereal time plus the projection of the IAU 2000B longitude nutation |
| `HourAngle` | hour angle from apparent RA and site longitude | longitude east-positive, degrees; hour angle `[0,360)`; based on apparent sidereal time |

```go
// Local sidereal time = Greenwich sidereal time + east longitude/15, folded back into [0,24) hours; math is only for this step.
gmst := coord.MeanSiderealTime(date)
gast := coord.ApparentSiderealTime(date)
lst := math.Mod(gmst+115.0/15, 24)
fmt.Printf("GMST=%.9f h GAST=%.9f h LST=%.9f h\n", gmst, gast, lst)
fmt.Printf("HA=%.6f deg\n", coord.HourAngle(date, eq.RA, 115))
```

### Precession and Nutation

| Name | Purpose | Units and conventions |
| --- | --- | --- |
| `Precess` | precess equatorial coordinates from one date to another | RA/Dec in degrees; both dates are civil instants; pure precession rotation, no proper motion |
| `EclipticObliquity` | ecliptic obliquity | degrees; `nutation=false` gives the IAU 1980 mean obliquity, `true` adds the IAU 2000B obliquity nutation |
| `Nutation2000B` | IAU 2000B nutation | returns `(longitude nutation, obliquity nutation)` in degrees |
| `Nutation1980` | IAU 1980 nutation | returns `(longitude nutation, obliquity nutation)` in degrees |

```go
// Precess J2000 equatorial coordinates to date and list both nutation series for comparison.
j2000 := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
p := coord.Precess(j2000, date, 83.6331, 22.0145)
dLon2000B, dObl2000B := coord.Nutation2000B(date)
dLon1980, dObl1980 := coord.Nutation1980(date)
fmt.Printf("precessed=(%.9f, %.9f)\n", p.RA, p.Dec)
fmt.Printf("eps mean=%.9f true=%.9f\n",
	coord.EclipticObliquity(date, false), coord.EclipticObliquity(date, true))
fmt.Printf("nut 2000B=(%.9f, %.9f) 1980=(%.9f, %.9f)\n",
	dLon2000B, dObl2000B, dLon1980, dObl1980)
```

### Galactic Coordinates

| Name | Purpose | Units and conventions |
| --- | --- | --- |
| `EquatorialToGalactic` | ICRS equatorial to Galactic | ICRS input in degrees; `Lon ∈ [0,360)`, `Lat ∈ [−90,90]`; fixed rotation matrix, no precession or proper motion |
| `GalacticToEquatorial` | Galactic to ICRS equatorial | transpose of the same matrix; `RA ∈ [0,360)` |
| `Galactic` | Galactic result value | fields `Lon` and `Lat`, in degrees |

```go
// The Galactic-center direction in ICRS (266.4051, -28.936175) should come back near l=0, b=0.
gal := coord.EquatorialToGalactic(266.4051, -28.936175)
back := coord.GalacticToEquatorial(gal.Lon, gal.Lat)
fmt.Printf("gal=(%.6f, %.6f) back=(%.9f, %.9f)\n", gal.Lon, gal.Lat, back.RA, back.Dec)
```

### Angular Separation

| Name | Purpose | Units and conventions |
| --- | --- | --- |
| `AngularSeparation` | angular distance between two equatorial coordinates | inputs and output in degrees; great-circle angle, independent of input order |

```go
// Angular distance between the Crab pulsar direction and the Galactic center.
sep := coord.AngularSeparation(83.6331, 22.0145, 266.4051, -28.936175)
fmt.Printf("sep=%.9f deg = %.6f arcsec\n", sep, sep*3600)
```

### Atmospheric Refraction

| Name | Purpose | Units and conventions |
| --- | --- | --- |
| `ApparentAltitude` | true altitude to apparent altitude | altitude and return value in degrees; pressure hPa, temperature °C |
| `TrueAltitude` | apparent altitude to true altitude | numerically inverts the Saemundsson model; returns NaN when there is no solution |
| `AtmosphericRefractionFromTrueAltitude` | refraction at a true altitude | degrees; add it to the true altitude to get the apparent altitude |
| `AtmosphericRefractionFromApparentAltitude` | refraction at an apparent altitude | degrees; subtract it from the apparent altitude to get the true altitude |
| `EquatorialToApparentHorizontal` | equatorial to apparent horizontal | applies refraction on top of `EquatorialToHorizontal`, changing only `Altitude`/`Zenith`, leaving `Azimuth`/`HourAngle` untouched |

```go
// True altitude 10 deg, standard pressure 1010 hPa, temperature 0 C.
apparent := coord.ApparentAltitude(10, 1010, 0)
ref := coord.AtmosphericRefractionFromTrueAltitude(10, 1010, 0)
appHz := coord.EquatorialToApparentHorizontal(date, eq.RA, eq.Dec, 115, 40, 1010, 0)
fmt.Printf("true=10 apparent=%.9f ref=%.9f\n", apparent, ref)
fmt.Printf("back=%.9f appHz alt=%.6f zen=%.6f\n",
	coord.TrueAltitude(apparent, 1010, 0), appHz.Altitude, appHz.Zenith)
```

### Airmass

| Name | Purpose | Units and conventions |
| --- | --- | --- |
| `AirmassPlaneParallelFromTrueAltitude` | plane-parallel model | true altitude in degrees; equivalent to `sec(z)`, good at moderate altitude, diverges near the horizon |
| `AirmassKastenYoungFromApparentAltitude` | Kasten-Young 1989 | apparent altitude in degrees; no automatic refraction correction |
| `AirmassPickeringFromApparentAltitude` | Pickering 2002 | apparent altitude in degrees; aimed at low-altitude observations |
| `AirmassKastenYoungFromTrueAltitude` | refraction first, then Kasten-Young | true altitude in degrees; pressure hPa, temperature °C |
| `AirmassPickeringFromTrueAltitude` | refraction first, then Pickering | same |

```go
// True altitude 10 deg, standard pressure 1010 hPa, temperature 0 C; FromTrueAltitude applies refraction first.
apparent := coord.ApparentAltitude(10, 1010, 0)
fmt.Printf("plane=%.9f\n", coord.AirmassPlaneParallelFromTrueAltitude(10))
fmt.Printf("ky true=%.9f apparent=%.9f\n",
	coord.AirmassKastenYoungFromTrueAltitude(10, 1010, 0),
	coord.AirmassKastenYoungFromApparentAltitude(apparent))
fmt.Printf("pickering true=%.9f apparent=%.9f\n",
	coord.AirmassPickeringFromTrueAltitude(10, 1010, 0),
	coord.AirmassPickeringFromApparentAltitude(apparent))
```

When only the raw formula is needed without the coordinate-layer refraction, use the three [airmass models](formula.md#airmass-models) in `formula`.

### Parallactic Angle

| Name | Purpose | Units and conventions |
| --- | --- | --- |
| `ParallacticAngle` | parallactic angle from apparent RA/Dec | longitude/latitude in degrees; internally hour angle plus declination, returns `(−180,180]` |
| `ParallacticAngleByHourAngle` | parallactic angle from hour angle | hour angle, declination, latitude in degrees; returns `(−180,180]` |

```go
// The parallactic angle drives camera rotation, spectrograph slit direction, and field orientation.
q := coord.ParallacticAngle(date, eq.RA, eq.Dec, 115, 40)
q2 := coord.ParallacticAngleByHourAngle(coord.HourAngle(date, eq.RA, 115), eq.Dec, 40)
fmt.Printf("q=%.9f qByHA=%.9f\n", q, q2)
```

The direction convention for the parallactic angle follows [Observing-angle semantics](sun-moon.md#observing-angle-semantics).

## Usage examples

### Round-trip conversion between ecliptic and horizontal coordinates

```go
eq := coord.EclipticToEquatorial(date, 139.686111, 4.875278)
hz := coord.EquatorialToHorizontal(date, eq.RA, eq.Dec, 115, 40)
back := coord.EquatorialToEcliptic(date, eq.RA, eq.Dec)
fmt.Println(eq.RA, eq.Dec, hz.Altitude, back.Lon)
```

```text
143.72223158223719 19.53512536790277 -17.686511328302952 139.68611100000000
```

The instant and the coordinates are enough; the library converts with the apparent obliquity of that day.

Use the `EclipticToEquatorialByObliquity` / `EquatorialToEclipticByObliquity` research entry points to prescribe the obliquity yourself.

Round trips are self-consistent: `EquatorialToEcliptic(EclipticToEquatorial(...))` returns the original values under one convention.

### Topocentric correction and sidereal time

```go
top := coord.TopocentricEquatorial(date, eq.RA, eq.Dec, 115, 40, 0.00257, 53)
fmt.Println(top.RA, top.Dec)
fmt.Println(coord.MeanSiderealTime(date), coord.ApparentSiderealTime(date))
// With a local sidereal time in hand you can skip the library's own computation:
manual := coord.EquatorialToHorizontalByLocalSiderealTime(10.5, 83.6331, 22.0145, 31.2)
fmt.Println(manual.Azimuth, manual.Altitude, manual.HourAngle)
```

```text
144.25512626944376 18.79025523004319
16.852455700085 16.852556781127
281.869347 24.489608 73.8669
```

`distanceAU` must be a real geocentric distance (about `0.00257 AU` for the Moon); whether parallax can be neglected for a more distant target depends on the required accuracy.

`height` is ellipsoidal height in metres and sidereal time is in hours.

### Refraction, airmass and parallactic angle

```go
fmt.Println(coord.ApparentAltitude(10, 1010, 0))                   // true 10 -> apparent
fmt.Println(coord.TrueAltitude(10.093428, 1010, 0))                // apparent -> true (inverse)
fmt.Println(coord.AirmassKastenYoungFromTrueAltitude(10, 1010, 0)) // airmass at true altitude 10
fmt.Println(coord.ParallacticAngle(date, eq.RA, eq.Dec, 115, 40))  // camera/slit rotation angle
```

```text
10.093427592862
10.000000410649
5.537933369472
-34.000957202636
```

Refraction only applies for true altitudes inside `(-5, 90)` degrees and contributes nothing outside; the airmass family also offers simplified entry points that take only an apparent altitude (`...FromApparentAltitude`).

### Precession, nutation and obliquity conventions

```go
fmt.Println(coord.EclipticObliquity(date, true)) // true obliquity; false gives the mean one
lon, obl := coord.Nutation2000B(date)            // IAU 2000B longitude/obliquity nutation
pre := coord.Precess(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), date, eq.RA, eq.Dec)
fmt.Println(lon, obl, pre.RA, pre.Dec)
```

```text
23.438261476424
0.001652570533 0.002392788113 144.089973153762 19.416731576422
```

Both `Nutation1980` and `Nutation2000B` can be called explicitly for term-by-term comparison; `Precess` applies precession only, without proper motion, nutation or aberration.

Research entry points such as a manual sidereal time or a manual obliquity live under [Research APIs and observing helpers](#research-apis-and-observing-helpers).

## Research APIs and Observing Helpers

Research-style `coord` helpers do not automatically substitute the current obliquity or sidereal time. They are useful for experiments with custom axial tilts or manually specified hour angles.

For ordinary observing calculations, use the `time.Time` based APIs such as `EclipticToEquatorial` and `EquatorialToHorizontal`.

Observing helpers:

- `ParallacticAngle` / `ParallacticAngleByHourAngle`: parallactic angle, or the direction angle of the zenith at the target
- `Airmass...FromApparentAltitude`: apply empirical airmass formula directly when apparent altitude is already known
- `Airmass...FromTrueAltitude`: estimate refraction from pressure/temperature, convert true altitude to apparent altitude, then compute airmass

```go
// Parallactic angle of the target, useful for camera rotation, spectrograph slit direction, and field orientation.
q := coord.ParallacticAngle(date, eq.RA, eq.Dec, 115, 40)

// With true altitude as input, estimate refraction first and then compute empirical airmass.
x := coord.AirmassKastenYoungFromTrueAltitude(10, 1010, 0)
fmt.Printf("q=%.6f airmass=%.6f\n", q, x)
```

The same observing helpers are also exposed in `sun`, `moon`, `star`, and the seven major-planet packages. If apparent altitude is already available and only the raw formula is needed, use `formula.Airmass...`.

## Parameter and result conventions

### Units

- Angles are always degrees: ecliptic longitude/latitude, right ascension/declination, azimuth/altitude/zenith distance/hour angle, parallactic angle, and angular separation.

  Multiply by 3600 yourself when arcseconds are needed (for example the return value of `AngularSeparation`).
- Sidereal time is always hours: the return values of `MeanSiderealTime` and `ApparentSiderealTime`, and the `localSiderealTimeHours` argument of every `*ByLocalSiderealTime` entry point, are all hours; internally they are multiplied by 15 to become degrees, so do not confuse them with the hour notation of right ascension.
- `Equatorial.RA`, `Galactic.Lon`, `Horizontal.Azimuth`, and `Horizontal.HourAngle` are all **degrees**, not hours; the `Equatorial` type does not distinguish J2000, mean-of-date, or apparent coordinates, so the caller must keep the input convention consistent.
- Distances are AU (the `distanceAU` argument of `TopocentricEquatorial` and `TopocentricEcliptic`); observer height is the **ellipsoidal height (geodetic height)** in meters, with no geoid modeling, see the manual [Observer height conventions](coord.md#observer-height).
- Pressure is hPa and temperature is °C; refraction is calibrated at the standard state of `1010 hPa` and `10 °C`.
- This package produces no apparent diameter or apparent radius. The apparent-radius fields in the Sun/Moon and eclipse manuals are in arcseconds; do not interchange them with the degrees used here.

### Time Scales

- Every public entry point treats `time.Time` as an absolute (civil) instant and converts it with `date.UTC()` before computing the Julian date; the value itself is always interpreted as a UTC label, and equals UT1 before `1972-01-01`.

  The full convention is in the manual [Time scale conventions](timescale.md).
- The sidereal-time chain converts to UT1 explicitly: `MeanSiderealTime` / `ApparentSiderealTime` run the IAU 2006 model after `UTC2UT1`, and the apparent sidereal time inside `HourAngle`, `EquatorialToHorizontal`, and `TopocentricEquatorial` is UT1-based as well.
- Both `from` and `to` of `Precess(from, to)` are civil instants, each supplying its own precession epoch.
- Refraction and airmass do not involve time at all; they depend only on altitude and meteorological parameters.

### Angle Quadrants and Normalization

- The internal `normalize360` normalizes every angle that is meant to be displayed into `[0,360)`: right ascension, ecliptic longitude, Galactic longitude, azimuth, and hour angle all land in that interval.
- Declination, ecliptic latitude, Galactic latitude, and altitude pass through `Asin` (with an extra `[−1,1]` clamp on some paths) and land in `[−90,90]`.
- `ParallacticAngle` / `ParallacticAngleByHourAngle` use `Atan2` and return `(−180,180]`; this is the only signed angle in the package.
- The topocentric family is the exception: `TopocentricEquatorial.RA` is the input RA plus a small correction and is not normalized into 360°, so it may cross the boundary slightly when the target is near `0°/360°`; `TopocentricEcliptic` normalizes `Lon` into `[0,360)` and returns `Lat` through `Asin` inside `[−90,90]`, because it solves the topocentric equatorial position first and converts to the ecliptic afterwards.
- The hour angle returned by `HourAngleDeclinationToHorizontal` and `HorizontalToHourAngleDeclination` is normalized into `[0,360)`, so a negative hour angle west of the meridian appears here as `360−|HA|`; subtract 360 yourself when a signed hour angle is needed.
- `EquatorialToHorizontal` and `HourAngleDeclinationToHorizontal` both return a **geometric** altitude without refraction; use `EquatorialToApparentHorizontal` or `ApparentAltitude` for an apparent altitude.

### Zero Values and Invalid Input

- The refraction family requires `pressureHPa > 0` and `temperatureC > −273.15`; any NaN/Inf or out-of-range parameter returns NaN.
- At a true altitude `≤ −5°` or `≥ 90°` the refraction is treated as 0: `ApparentAltitude` returns the true altitude unchanged, and `TrueAltitude` and `AtmosphericRefractionFromApparentAltitude` return the input apparent altitude unchanged; `TrueAltitude` only solves numerically inside `(−5°, 90°)` and returns NaN when the inversion fails.
- The airmass family limits altitude (or zenith distance) to `[0,90]` and returns NaN outside or for non-finite input; the plane-parallel model returns `+Inf` exactly at altitude `0°` (zenith distance `90°`), while Kasten-Young and Pickering stay finite at `0°`.
- `EquatorialToApparentHorizontal` passes pressure and temperature straight to the refraction function: invalid meteorological parameters turn `Altitude` and therefore `Zenith` into NaN, while `Azimuth` and `HourAngle` remain geometric values.
- The topocentric family does not validate `distanceAU` or `height`; passing `0` or a negative `distanceAU` yields meaningless results, so pass a real geocentric distance (about `0.00257 AU` for the Moon).
- Purely geometric entry points such as the Galactic conversions and the angular separation do not validate parameters: non-finite input propagates to NaN, and an ecliptic latitude outside `[−90,90]` is not clamped but reinterpreted as a spherical direction.
- The four coordinate types have no `Valid` field: `Ecliptic{}`, `Equatorial{}`, `Horizontal{}`, and `Galactic{}` mean "all angles zero", not "unset", so zero values are not usable as an unset sentinel.

### Accuracy and Applicability

- Ecliptic obliquity: the mean term uses the IAU 1980 model, and `EclipticObliquity(date, true)` adds the IAU 2000B obliquity nutation on top; both `Nutation1980` and `Nutation2000B` are available and can be called explicitly for term-by-term comparison.
- Sidereal time: the IAU 2006 ERA route, with apparent sidereal time adding the IAU 2000B longitude nutation; apparent sidereal time for the same Julian date is memoized in a bounded table, and an injected ΔT override invalidates the memo by generation.
- Precession: `Precess` applies a long-term precession rotation to equatorial coordinates (its equatorial/ecliptic pole vectors include periodic terms), without proper motion, nutation, or aberration; it moves mean coordinates from one epoch to another and cannot replace a full "J2000 to apparent-of-date" chain.
- Galactic: a fixed ICRS ↔ Galactic rotation matrix, equivalent to SOFA `iauIcrs2g`/`iauG2icrs`, valid only for ICRS-convention input.
- Topocentric correction: the diurnal parallax scales inversely with `distanceAU` and is largest for the Moon (about 1°), and whether it can be neglected for a more distant body depends on the required accuracy; `height` only scales the parallax term, with meter-level elevation producing arcsecond-level differences.
- Topocentric ecliptic coordinates come from the topocentric equatorial position (`basic.TopocentricLoBo` returns both from a single evaluation): `TopocentricEcliptic.Lat` is guaranteed to land in `[−90,90]` and agrees with the independent `EquatorialToEcliptic(TopocentricEquatorial(...))` route.
- Refraction: the Saemundsson formula, valid for true altitudes in `(−5°, 90°)`, with no correction outside that window; refraction changes rapidly at low altitude (< 5°), where ±1 hPa or ±1 °C already produces a visible error.
- Airmass: the three empirical models agree at moderate altitude and differ most near the horizon; the plane-parallel model is a purely geometric approximation that diverges near the horizon, so use Kasten-Young or Pickering for careful low-altitude estimates, with the raw formulas in `formula` under [airmass models](formula.md#airmass-models).
- Parallactic angle: `ParallacticAngle` uses apparent RA/Dec and the site longitude/latitude, while `ParallacticAngleByHourAngle` uses the hour-angle form of the same geometry; both must agree for the same geometry.

## Observer height

Topocentric coordinates, rise/set, local eclipses and occultations interpret observer height as ellipsoidal height in metres. Convert orthometric height `H` using geoid undulation `N`: `height = H + N`.

The library does not include a geoid model. `height=0` therefore denotes the reference ellipsoid, rather than local mean sea level.

Obtain `N` from an external model when that difference matters, and match the observer-height convention when comparing external predictions.

For scale, a height difference of 30 m projects to roughly 52 m on the ground along a ray at 30° altitude, using `|Δh|·cot(altitude)`. Its effect on a contact time depends on the local shadow velocity and contact geometry; it is not a fixed time correction.
