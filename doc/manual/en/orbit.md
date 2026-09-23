# Generic Small-Body Orbits

[中文](../orbit.md) | [Back to README](../../../README.en.md)

`orbit` propagates heliocentric two-body positions from orbital elements. It supports asteroids, comets, dwarf planets, and custom hypothetical orbits.

The seven major planets are still computed by their own packages using built-in VSOP87 analytical terms.

`orbit.Elements` supports two common forms:

- classical elliptical elements: `A/E/I/Omega/W/M0`
- perihelion form: `Q/E/I/Omega/W/TpJD`, useful for comets and high-eccentricity orbits

The reference frame of `orbit.Elements` is always the J2000 mean ecliptic and mean equinox. The examples below use one set of Ceres elements.

Topocentric and rise/set helpers take east-positive longitude, north-positive latitude, and height in meters.

The positional and topocentric quantities here also work together with the interfaces documented in the [star](star.md#stars) and [coordinate tools](coord.md#coordinate-tools) manuals.

## Contents

- [Calculating the position of Ceres from orbital elements](#calculating-the-position-of-ceres-from-orbital-elements)
- [API Reference](#api-reference)
  - [Orbital Elements](#orbital-elements)
  - [Positions](#positions)
  - [Geometry](#geometry)
  - [Rise, Set, and Culmination](#rise-set-and-culmination)
  - [Photometry](#photometry)
  - [Visual Binaries](#visual-binaries)
- [Usage examples](#usage-examples)
  - [Picking a position layer](#picking-a-position-layer)
  - [Will it rise tonight, and how high is it now?](#will-it-rise-tonight-and-how-high-is-it-now)
  - [Distances, phase angle and H-G magnitude](#distances-phase-angle-and-h-g-magnitude)
  - [Orbit types and position layers](#orbit-types-and-position-layers)
- [Parameter and result conventions](#parameter-and-result-conventions)
  - [Position Layers](#position-layers)
  - [Units and Frame Conventions](#units-and-frame-conventions)
  - [Time Scale and Input Reading](#time-scale-and-input-reading)
  - [Zero Values and Out-of-Range Input](#zero-values-and-out-of-range-input)
  - [Accuracy and Applicability](#accuracy-and-applicability)
  - [Common Pitfalls](#common-pitfalls)
- [Related Manuals](#related-manuals)

## Calculating the position of Ceres from orbital elements

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro/orbit"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	when := time.Date(2025, 11, 21, 20, 0, 0, 0, cst)
	ceres := orbit.Elements{
		EpochJD: 2461000.5, A: 2.765615651508659, E: 0.07957631994408416,
		I: 10.58788658206854, Omega: 80.24963090816965,
		W: 73.29975464616518, M0: 231.5397330043706,
	}
	pos := orbit.ApparentGeocentricEquatorial(when, ceres)
	fmt.Printf("RA=%.6f Dec=%.6f deg distance=%.6f AU\n", pos.RA, pos.Dec, pos.Distance)
	rise, err := orbit.RiseTime(when, ceres, 121.4737, 31.2304, 20, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(rise.Format(time.RFC3339))
}
```

Elements use the J2000 mean ecliptic frame and a TT/TDB Julian-day epoch. This is heliocentric two-body propagation; perturbation errors need separate evaluation over long spans or near planetary encounters.

## API Reference

### Orbital Elements

| Name | Purpose | Units and convention |
| --- | --- | --- |
| `Elements` | Heliocentric two-body conic elements | `EpochJD`/`TpJD` are TT/TDB Julian days; `A`/`Q` in AU, `I`/`Omega`/`W`/`M0` in degrees, `E` dimensionless; `ADot`…`MDot` are per-day rates that apply to the classical elliptical form only |
| `MeanMotion` | Mean angular rate | degrees/day; `NaN` for parabolic and hyperbolic cases; a non-zero `MDot` is used directly |
| `MeanAnomaly` | Mean anomaly | degrees (`[0,360)`); `NaN` for parabolic and hyperbolic cases |
| `TrueAnomaly` | True anomaly | degrees (`[0,360)`); defined for elliptical, parabolic, and hyperbolic orbits, `NaN` for invalid elements |

Mean anomaly and true anomaly are solved from the same element set, and `MDot` can replace the default mean motion:

```go
fmt.Println(orbit.MeanMotion(ceres), orbit.MeanAnomaly(when, ceres), orbit.TrueAnomaly(when, ceres))

// Parabolic and hyperbolic orbits only accept the perihelion form; mean motion and mean anomaly are undefined.
parabolic := orbit.Elements{Q: 0.9, E: 1, I: 30, Omega: 40, W: 50, TpJD: 2461000.5}
hyperbolic := orbit.Elements{Q: 1.2, E: 1.05, I: 30, Omega: 40, W: 50, TpJD: 2461000.5}
fmt.Println(orbit.MeanMotion(parabolic), orbit.MeanAnomaly(when, parabolic))
fmt.Println(orbit.TrueAnomaly(when, parabolic), orbit.TrueAnomaly(when, hyperbolic))
```

The complete example exercises both the elliptical and the perihelion element form:

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/orbit"
)

func main() {
	// Classical elliptical elements for 1 Ceres, referenced to J2000 mean ecliptic/equinox.
	ceres := orbit.Elements{
		EpochJD: 2461000.5,
		A:       2.765615651508659,
		E:       0.07957631994408416,
		I:       10.58788658206854,
		Omega:   80.24963090816965,
		W:       73.29975464616518,
		M0:      231.5397330043706,
	}
	ceresPos := orbit.ApparentGeocentricEquatorial(
		time.Date(2025, 11, 12, 0, 0, 0, 0, time.UTC),
		ceres,
	)
	fmt.Printf("ceres ra=%.6f dec=%.6f distance=%.6f\n", ceresPos.RA, ceresPos.Dec, ceresPos.Distance)

	// Halley's Comet example using perihelion distance Q and perihelion passage time TpJD.
	halley := orbit.Elements{
		Q:     0.5870992,
		E:     0.9671429,
		I:     162.26269,
		Omega: 58.42008,
		W:     111.33249,
		TpJD:  2446467.395,
	}
	halleyPos := orbit.ApparentGeocentricEquatorial(
		time.Date(1986, 2, 9, 0, 0, 0, 0, time.UTC),
		halley,
	)
	fmt.Printf("halley ra=%.6f dec=%.6f distance=%.6f\n", halleyPos.RA, halleyPos.Dec, halleyPos.Distance)
}
```

Output:

```text
ceres ra=7.739532 dec=-10.625981 distance=2.164391
halley ra=312.112360 dec=-11.826451 distance=1.533936
```

Orbital elements have epochs. The farther the target date is from the epoch, the more static-element error can grow.

If the source provides long-term linear rates such as `ADot/EDot/IDot/OmegaDot/WDot/MDot`, they can be filled into `Elements` to reduce medium- and long-term drift.

### Positions

| Name | Purpose | Units and convention |
| --- | --- | --- |
| `EclipticPosition` | Ecliptic spherical return value | `Lon`/`Lat` in degrees, `Distance` in AU |
| `EquatorialPosition` | Equatorial spherical return value | `RA`/`Dec` in degrees, `Distance` in AU |
| `HeliocentricEclipticJ2000` | Heliocentric J2000 mean ecliptic | geometric, no light-time |
| `HeliocentricEcliptic` | Heliocentric ecliptic of date | geometric, referred to the mean equinox of date |
| `GeocentricEclipticJ2000` | Geocentric J2000 mean ecliptic | geometric, Earth and target at the same instant |
| `GeocentricEcliptic` | Geocentric ecliptic of date | geometric, referred to the mean equinox of date |
| `GeocentricEquatorialJ2000` | Geocentric J2000 mean equatorial | geometric, J2000 obliquity |
| `GeocentricEquatorial` | Geocentric mean equatorial of date | geometric, obliquity of date |
| `AstrometricGeocentricEquatorialJ2000` | Astrometric geocentric J2000 equatorial | geometric position plus light-time, directly comparable with J2000 catalogues |
| `ApparentGeocentricEcliptic` | Apparent geocentric ecliptic | light-time plus nutation, without a full aberration model |
| `ApparentGeocentricEquatorial` | Apparent geocentric equatorial | light-time plus nutation, without a full aberration model |
| `ApparentTopocentricEquatorial` | Apparent topocentric equatorial | apparent geocentric plus topocentric parallax |

Each step down the chain adds one correction to the same instant — geometric, then light-time, then nutation, then topocentric:

```go
h := orbit.HeliocentricEcliptic(when, ceres)                  // heliocentric ecliptic of date, geometric
g := orbit.GeocentricEquatorialJ2000(when, ceres)             // geocentric J2000 mean equatorial
a := orbit.ApparentGeocentricEquatorial(when, ceres)          // apparent geocentric equatorial
t := orbit.ApparentTopocentricEquatorial(when, ceres, 121.4737, 31.2304, 20)
e := orbit.ApparentGeocentricEcliptic(when, ceres)
fmt.Printf("h=%.6f %.6f %.6f\n", h.Lon, h.Lat, h.Distance)
fmt.Printf("g=%.6f %.6f %.6f\n", g.RA, g.Dec, g.Distance)
fmt.Printf("a=%.6f %.6f t=%.6f %.6f e=%.6f\n", a.RA, a.Dec, t.RA, t.Dec, e.Lon)
```

The `...J2000` variants and the of-date variants are two different frames: the former stay in the J2000 mean ecliptic/equinox, which suits catalogue comparison and long-term archiving; the latter use the mean ecliptic/equinox of date, which suits "today's sky" expressions. `Distance` is the instantaneous distance for the geometric interfaces and the light-time-converged distance for `Astrometric...`; the two differ by roughly the displacement during the light-time.

### Geometry

| Name | Purpose | Units and convention |
| --- | --- | --- |
| `SunDistance` | Heliocentric distance | AU, geometric |
| `EarthDistance` | Geocentric distance | AU, geometric |
| `Elongation` | Solar elongation | degrees, apparent geocentric angular separation |
| `PhaseAngle` | Phase angle | degrees, 0° means the fully lit face points at the observer |
| `IlluminatedFraction` | Illuminated fraction | dimensionless, typically within `[0,1]` |
| `Phase` | Alias of the illuminated fraction | identical to `IlluminatedFraction` |
| `ParallacticAngle` | Parallactic (zenith-direction) angle | degrees, hour angle and declination from one and the same topocentric solve |

`orbit` also provides common observing geometry and lightweight photometry helpers:

```go
r := orbit.SunDistance(when, ceres)                         // heliocentric distance
delta := orbit.EarthDistance(when, ceres)                   // geocentric distance
elong := orbit.Elongation(when, ceres)                      // solar elongation
phase := orbit.PhaseAngle(when, ceres)                      // phase angle
k := orbit.IlluminatedFraction(when, ceres)                 // illuminated fraction
mag := orbit.AsteroidMagnitudeHG(when, ceres, 3.34, 0.12)   // H-G asteroid magnitude
q := orbit.ParallacticAngle(when, ceres, 121.4737, 31.2304, 20) // parallactic angle from a site

fmt.Printf("r=%.6f delta=%.6f elong=%.6f phase=%.6f k=%.6f mag=%.3f q=%.6f\n",
	r, delta, elong, phase, k, mag, q)
```

Near 180° elongation the phase angle approaches 0° and `IlluminatedFraction` approaches 1; all three quantities come from the same geocentric geometry.

`ParallacticAngle` does not solve for declination separately: it reuses the same topocentric solve as `HourAngle`, so that hour angle and declination never come from two Julian-day paths about 1 ULP apart and introduce sub-nanodegree drift.

All geometry helpers depend only on the absolute instant of `date`; `ParallacticAngle` is the only one that also takes observer parameters.

### Rise, Set, and Culmination

| Name | Purpose | Units and convention |
| --- | --- | --- |
| `Altitude` | Apparent altitude | degrees, apparent topocentric position on the observer's local civil day |
| `Zenith` | Zenith distance | degrees, equal to `90 - Altitude` |
| `Azimuth` | Apparent azimuth | degrees, north 0°, increasing toward east |
| `HourAngle` | Topocentric hour angle | degrees |
| `CulminationTime` | Culmination time | `time.Time`, keeps the location of the input `date` |
| `RiseTime` | Rise time | `(time.Time, error)`, second value is a sentinel error |
| `SetTime` | Set time | `(time.Time, error)`, second value is a sentinel error |
| `ERR_ORBIT_NEVER_RISE` | Sentinel: target never rises that day | returned by `RiseTime` |
| `ERR_ORBIT_NEVER_SET` | Sentinel: target never sets that day | returned by `SetTime` |

```go
fmt.Println(orbit.Zenith(when, ceres, 121.4737, 31.2304, 20))
fmt.Println(orbit.HourAngle(when, ceres, 121.4737, 31.2304, 20))
fmt.Println(orbit.CulminationTime(when, ceres, 121.4737, 31.2304, 20).Format(time.RFC3339))

day := time.Date(2025, 11, 21, 0, 0, 0, 0, site)
set, err := orbit.SetTime(day, ceres, 121.4737, 31.2304, 20, true)
if errors.Is(err, orbit.ERR_ORBIT_NEVER_SET) {
	fmt.Println("never sets today", set)
}
```

Treat an orbit as an observable target for topocentric pointing:

```go
site := time.FixedZone("CST", 8*3600)
when := time.Date(2025, 11, 21, 20, 0, 0, 0, site)

alt := orbit.Altitude(when, ceres, 121.4737, 31.2304, 20) // topocentric altitude
az := orbit.Azimuth(when, ceres, 121.4737, 31.2304, 20)   // topocentric azimuth
rise, _ := orbit.RiseTime(time.Date(2025, 11, 21, 0, 0, 0, 0, site), ceres, 121.4737, 31.2304, 20, true) // rise time

fmt.Printf("alt=%.6f az=%.6f rise=%s\n", alt, az, rise.Format(time.RFC3339))
```

These observing helpers work on topocentric apparent coordinates and suit rise/set and pointing support for asteroids, comets, or custom two-body targets.

With `aero` false the criterion is the geometric horizon; with `aero` true the target altitude is taken as `-0.5667°` plus the horizon dip derived from ellipsoidal height and latitude.

The second return value of `RiseTime`/`SetTime` is a real error: only a day without a rise/set is mapped to `ERR_ORBIT_NEVER_RISE` / `ERR_ORBIT_NEVER_SET`, and any other failure passes through unchanged.

### Photometry

| Name | Purpose | Units and convention |
| --- | --- | --- |
| `AsteroidMagnitudeHG` | Asteroid apparent magnitude, H-G model | `absoluteMagnitude` is H and `slopeParameter` is G; both dimensionless |

```go
fmt.Printf("H-G magnitude=%.3f\n", orbit.AsteroidMagnitudeHG(when, ceres, 3.34, 0.12))
fmt.Printf("r=%.6f delta=%.6f elong=%.6f\n",
	orbit.SunDistance(when, ceres), orbit.EarthDistance(when, ceres), orbit.Elongation(when, ceres))
fmt.Printf("phase=%.6f k=%.6f k2=%.6f\n",
	orbit.PhaseAngle(when, ceres), orbit.IlluminatedFraction(when, ceres), orbit.Phase(when, ceres))
```

The H-G model uses only heliocentric distance, geocentric distance, and phase angle; it does not introduce the target radius, albedo, or rotation, and this package never fills in a default value for `G` — that comes from the external catalogue.

### Visual Binaries

| Name | Purpose | Units and convention |
| --- | --- | --- |
| `VisualBinaryElements` | Visual-binary orbital elements | `PeriodYears` in mean solar years, `PeriastronYear` as a decimal year, `SemiMajorAxis` in arcseconds, `Inclination`/`AscendingNode`/`PeriastronArgument` in degrees, `Eccentricity` dimensionless |
| `VisualBinaryPosition` | Computed visual-binary position | `MeanAnomaly`/`EccentricAnomaly`/`TrueAnomaly`/`PositionAngle` in degrees, `Radius`/`Separation` in arcseconds |
| `VisualBinary` | Visual-binary position at an instant | converts the instant to a UTC decimal year, then applies the classical apparent-orbit formula |
| `VisualBinaryByYear` | Visual-binary position by decimal year | takes the decimal year directly, skipping the instant conversion |

```go
gammaVir := orbit.VisualBinaryElements{
	PeriodYears: 171.37, PeriastronYear: 1836.433, Eccentricity: 0.8808,
	SemiMajorAxis: 3.746, Inclination: 146.05, AscendingNode: 31.78, PeriastronArgument: 252.88,
}
vb := orbit.VisualBinaryByYear(2026.0, gammaVir)
fmt.Printf("theta=%.6f rho=%.6f M=%.6f\n", vb.PositionAngle, vb.Separation, vb.MeanAnomaly)
```

`orbit` also includes a lightweight visual-binary solver using the classical apparent-orbit formula from chapter 55 of *Astronomical Algorithms*:

```go
gammaVir := orbit.VisualBinaryElements{
	PeriodYears:        171.37,
	PeriastronYear:     1836.433,
	Eccentricity:       0.8808,
	SemiMajorAxis:      3.746,
	Inclination:        146.05,
	AscendingNode:      31.78,
	PeriastronArgument: 252.88,
}
vb := orbit.VisualBinary(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), gammaVir)
fmt.Printf("theta=%.6f rho=%.6f\n", vb.PositionAngle, vb.Separation) // position angle and separation
```

Position angle is measured with north at 0° and east at 90°, and both the separation and the radius vector `Radius` are in arcseconds.

## Usage examples

### Picking a position layer

```go
helio := orbit.HeliocentricEcliptic(when, ceres)
geo := orbit.GeocentricEclipticJ2000(when, ceres)
ast := orbit.AstrometricGeocentricEquatorialJ2000(when, ceres)
app := orbit.ApparentGeocentricEquatorial(when, ceres)
fmt.Println(helio.Lon, helio.Lat, helio.Distance)
fmt.Println(geo.Lon, geo.Lat, geo.Distance)
fmt.Println(ast.RA, ast.Dec)
fmt.Println(app.RA, app.Dec, app.Distance)
```

```text
19.251489 -9.315340 2.912174
2.147652 -12.026350 2.262445
6.795211 -10.172420
7.125008 -10.028726 2.262489
```

The four layers mean different things: `Heliocentric*` is relative to the Sun (the first number is the heliocentric distance), `Geocentric*J2000` is the J2000 geocentric position, `Astrometric*J2000` removes light-time and suits catalog comparison, and `Apparent*` is the apparent position of the day used for observing and charts.

For a topocentric apparent position use `ApparentTopocentricEquatorial(when, ceres, lon, lat, height)`.

### Will it rise tonight, and how high is it now?

```go
fmt.Println(orbit.RiseTime(when, ceres, 121.4737, 31.2304, 20, true))
fmt.Println(orbit.CulminationTime(when, ceres, 121.4737, 31.2304, 20))
fmt.Println(orbit.Altitude(when, ceres, 121.4737, 31.2304, 20),
	orbit.Azimuth(when, ceres, 121.4737, 31.2304, 20))
```

```text
2025-11-21 14:41:48.913 CST <nil>
2025-11-21 20:19:34 CST
48.472628 172.699168
```

- `aero = true` solves against the horizon corrected for refraction and apparent radius; `height` is ellipsoidal height in metres and longitude is east positive.
- Circumpolar or polar targets have no rise or set: `RiseTime`/`SetTime` return the `orbit.ERR_ORBIT_NEVER_RISE` / `ERR_ORBIT_NEVER_SET` sentinels, so branch with `errors.Is`.

### Distances, phase angle and H-G magnitude

```go
fmt.Println(orbit.SunDistance(when, ceres), orbit.EarthDistance(when, ceres))
fmt.Println(orbit.Elongation(when, ceres), orbit.PhaseAngle(when, ceres))
fmt.Println(orbit.IlluminatedFraction(when, ceres), orbit.AsteroidMagnitudeHG(when, ceres, 3.34, 0.12))
```

```text
2.912174 2.262445
122.264630 16.670089
0.978986 8.360
```

- `PhaseAngle` is the Sun-target-Earth angle in degrees and `IlluminatedFraction` is the illuminated fraction; `Phase` is only an alias of `IlluminatedFraction`, so do not read it as a phase angle.
- H-G magnitudes take the absolute magnitude `H` and the slope parameter `G` (Ceres uses `3.34` and `0.12` in the example).

### Orbit types and position layers

```go
parabolic := orbit.Elements{Q: 0.9, E: 1, I: 30, Omega: 40, W: 50, TpJD: 2461000.5}
hyperbolic := orbit.Elements{Q: 1.2, E: 1.05, I: 30, Omega: 40, W: 50, TpJD: 2461000.5}
fmt.Println(orbit.MeanMotion(parabolic), orbit.MeanAnomaly(when, parabolic))
fmt.Println(orbit.TrueAnomaly(when, parabolic), orbit.TrueAnomaly(when, hyperbolic))
fmt.Println(orbit.MeanMotion(ceres))
```

```text
NaN NaN
0.817533 0.537610
0.21429712142765137
```

- Parabolic and hyperbolic orbits only work in perihelion form (`Q` plus `TpJD`): mean motion and mean anomaly are undefined and return `NaN`, while the true anomaly solves for all three orbit types.
- When cross-checking against an external ephemeris, align the position layer and frame first (`Elements` is always J2000 mean ecliptic/equinox); `ADot…WDot` only apply to the classical ellipse form, and a non-zero `MDot` replaces the default mean motion.

## Parameter and result conventions

### Position Layers

| Layer | Representative interfaces | Corrections included |
| --- | --- | --- |
| Heliocentric geometric | `HeliocentricEcliptic` / `HeliocentricEclipticJ2000` | none |
| Geocentric geometric | `GeocentricEcliptic` / `GeocentricEclipticJ2000` / `GeocentricEquatorial` / `GeocentricEquatorialJ2000` | minus the heliocentric Earth position |
| Geocentric astrometric | `AstrometricGeocentricEquatorialJ2000` | light-time |
| Geocentric apparent | `ApparentGeocentricEcliptic` / `ApparentGeocentricEquatorial` | light-time + nutation |
| Topocentric apparent | `ApparentTopocentricEquatorial` and all rise/set helpers | light-time + nutation + topocentric parallax |

### Units and Frame Conventions

- Angles are always in degrees, distances in AU, ellipsoidal heights for topocentric and rise/set helpers in meters, and time as `time.Time`.
- The frame of `Elements` is the J2000 mean ecliptic and mean equinox. Interfaces ending in `...J2000` keep that frame; the other `HeliocentricEcliptic`/`GeocentricEcliptic`/`GeocentricEquatorial` variants are of-date quantities, so do not mix the two frames.
- Positions come in three layers: `Heliocentric*`/`Geocentric*` are geometric, `AstrometricGeocentricEquatorialJ2000` adds light-time to the geometric position (solved iteratively in distance, at most 8 passes, converged at `1e-12` days), and `Apparent*` adds nutation on top of light-time.

  The planetary convention of this repository is exactly "light-time + nutation, without a full external aberration model", so `Apparent*` is at the same level as the planet packages and must not be treated as a full apparent place.
- `MeanMotion`/`MeanAnomaly`/`TrueAnomaly` return degrees; mean motion and mean anomaly are undefined for parabolic and hyperbolic orbits.

### Time Scale and Input Reading

- `EpochJD` and `TpJD` are TT/TDB Julian days. The coordinate, geometry, and photometry interfaces treat `date` as an absolute instant (`date.UTC()`, then `UTC2TT` to TT/TDB).

  `UTC2TT` treats civil values as UT1 before 1972-01-01 and uses the built-in leap-second table inside the exact window; that table can be overridden with `astro.SetTTMinusUTC`.
- Instantaneous topocentric quantities combine the local clock fields with `date.Zone()` to recover the absolute instant. UTC and local-zone representations of the same instant give the same position. Rise/set searches also use the local date, so choose the zone for the observing calendar.
- `CulminationTime`, `RiseTime`, and `SetTime` keep the zone of the input `date` and return civil instants. Internally they step back 12 hours when `date.Hour() > 12`, so that the search anchor stays within the selected date.
- Every public `time.Time` output is a civil value (UTC label); UT1 and TT only appear in internal conversions. For explicit conversion use the root package's `astro.UT1FromUTC` / `astro.TTFromUTC`.

### Zero Values and Out-of-Range Input

- The zero value of `Elements` is not a valid orbit. The classical elliptical form requires a finite positive `A`, `E` within `[0,1)`, and finite `EpochJD` and `M0`.

  Parabolic and hyperbolic orbits with `E >= 1` can only use the perihelion form, which requires a finite positive `Q`, a finite `TpJD`, `E >= 0`, and finite `I`/`Omega`/`W`.
- When `Q > 0` and `TpJD` is finite the perihelion form wins: `A`, `M0`, and `EpochJD` are ignored, `ADot`/`EDot`/`IDot`/`OmegaDot`/`WDot` have no effect, and only a non-zero `MDot` is used as the mean angular rate.
- With invalid elements `MeanMotion` and `MeanAnomaly` return `NaN` and the positional interfaces return three `NaN`s.

  `AsteroidMagnitudeHG` returns `NaN` when its inputs are non-finite or when heliocentric distance, geocentric distance, or phase angle is non-positive, and `+Inf` when the H-G phase blend is zero.
- `RiseTime` and `SetTime` return the sentinel errors `ERR_ORBIT_NEVER_RISE` / `ERR_ORBIT_NEVER_SET` when no rise/set exists on the supplied local day, and pass any other failure through; on success the second return value is `nil`.
- `VisualBinary` and `VisualBinaryByYear` fill every numeric field of `VisualBinaryPosition` with `NaN` when `PeriodYears <= 0`, `SemiMajorAxis <= 0`, `Eccentricity` outside `[0,1)`, or any element is non-finite.

### Accuracy and Applicability

- `orbit` is two-body conic propagation: it contains only the supplied elements, with no planetary perturbations, non-gravitational terms, or relativistic corrections. The farther the date from the epoch, the larger the static-element error; `ADot`…`WDot` only mitigate linear drift and cannot replace a re-fit or a numerical integration.
- The difference between the `Apparent*` and topocentric quantities is purely geometric plus nutation, and neither includes atmospheric refraction; only `aero=true` in the rise/set helpers folds refraction into the criterion (target altitude `-0.5667°` plus the horizon dip derived from ellipsoidal height and latitude).
- `RiseTime`/`SetTime` iterate the rise/set geometry in the nominal zone `round(observerLon/15)`, so the site should sit near the center of its zone; only one root is solved, and boundary cases such as polar or circumpolar targets are reported through the sentinel errors.
- The visual-binary solver uses the classical apparent-orbit formula from chapter 55 of *Astronomical Algorithms* and does not apply when `Eccentricity >= 1`; it is a geometric projection only, with no mass, photometry, or perturbation information.

### Common Pitfalls

- Treating a `HeliocentricEcliptic` result as geocentric: heliocentric quantities are centered on the Sun, while geocentric ones have already subtracted the heliocentric Earth position.
- Treating `Apparent*` as a refraction-corrected apparent place: refraction appears only in the rise/set and observing helpers, not in the coordinate interfaces.
- Probing with a zero-value `Elements`: mean motion and the positional interfaces quietly return `NaN` instead of reporting an error.
- Expecting `ADot`…`WDot` to take effect in the perihelion form: those rates only propagate in the classical elliptical form.

## Related Manuals

- Stars and catalogues: [star](star.md#stars)
- Frame conversion and topocentric quantities: [coordinate tools](coord.md#coordinate-tools)
- The analogous rise/set interfaces: [Sun and Moon](sun-moon.md#sunrisesunset-and-moonrisemoonset)

> To feed the ecliptic/equatorial coordinates here into a star catalogue or topocentric quantities, see [star](star.md#stars) and [coordinate tools](coord.md#coordinate-tools).
