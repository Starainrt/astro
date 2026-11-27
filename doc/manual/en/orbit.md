# Orbits

[中文](../orbit.md) | [Back to README](../../../README.en.md)

The `orbit` package propagates asteroids, comets, dwarf planets, and custom objects from heliocentric orbital elements—supporting elliptical and perihelion forms, optional numerical planetary perturbations, topocentric rise/set, H-G magnitudes, and classic visual binary orbits.

## Contents

- [Basic Example](#basic-example)
- [Usage](#usage)
  - [Comets, Parabolic & Hyperbolic Orbits (Perihelion Form)](#comets-parabolic--hyperbolic-orbits-perihelion-form)
  - [Perturbed Propagation & Ephemeris Stacking](#perturbed-propagation--ephemeris-stacking)
  - [Visual Binary Position Angle & Separation](#visual-binary-position-angle--separation)
- [API Reference](#api-reference)
- [Conventions and Limitations](#conventions-and-limitations)

## Basic Example

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
	lon, lat, height := 121.4737, 31.2304, 20.0 // Shanghai, 20 m ellipsoidal height

	ceres := orbit.Elements{
		EpochJD: 2461000.5, // TT/TDB Julian date
		A:       2.765615651508659,
		E:       0.07957631994408416,
		I:       10.58788658206854,
		Omega:   80.24963090816965,
		W:       73.29975464616518,
		M0:      231.5397330043706,
	}

	pos := orbit.ApparentGeocentricEquatorial(when, ceres)
	mag := orbit.AsteroidMagnitudeHG(when, ceres, 3.34, 0.12) // H=3.34, G=0.12
	alt := orbit.Altitude(when, ceres, lon, lat, height)
	rise, err := orbit.RiseTime(when, ceres, lon, lat, height, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("1 Ceres (2025-11-21 20:00 CST): RA=%.3f° Dec=%.3f° dist=%.4f AU mag=%.2f\n",
		pos.RA, pos.Dec, pos.Distance, mag)
	fmt.Printf("Shanghai alt %.2f°  rise %s\n", alt, rise.Format("15:04:05"))
}
```

Output:

```text
1 Ceres (2025-11-21 20:00 CST): RA=7.128° Dec=-10.028° dist=2.2625 AU mag=8.36
Shanghai alt 48.47°  rise 14:41:49
```

- Elements are always referenced to **J2000 mean ecliptic / equinox**, with `EpochJD`/`TpJD` in TT/TDB, angles in degrees, and distances in AU.
- Use `AstrometricGeocentricEquatorialJ2000` when comparing with J2000 star catalogs, or `ApparentTopocentricEquatorial` for close approaches.

## Usage

### Comets, Parabolic & Hyperbolic Orbits (Perihelion Form)

Use perihelion distance `Q` (AU) and time of perihelion passage `TpJD` (TT/TDB) for comets or `E >= 1` orbits:

```go
halley := orbit.Elements{
	Q: 0.5870992, E: 0.9671429, I: 162.26269,
	Omega: 58.42008, W: 111.33249, TpJD: 2446467.395,
}
hp := orbit.ApparentGeocentricEquatorial(time.Date(1986, 2, 9, 0, 0, 0, 0, time.UTC), halley)
fmt.Printf("Halley (1986-02-09 UTC): RA=%.2f° Dec=%.2f° dist=%.4f AU\n", hp.RA, hp.Dec, hp.Distance)
```

Output:

```text
Halley (1986-02-09 UTC): RA=312.11° Dec=-11.83° dist=1.5339 AU
```

> When `Q > 0` and `TpJD` is finite, perihelion form takes precedence. For `E >= 1`, `MeanMotion` and `MeanAnomaly` return `NaN`, while `TrueAnomaly` and all position APIs work normally.

### Perturbed Propagation & Ephemeris Stacking

Use `StateWithOptions` to include numerical planetary perturbations, or wrap it with `NewEphemerisEngine` to plug into `astro.NewStackInstance`:

| `PropagationOptions` Field | Description |
| --- | --- |
| `Perturbers` | Perturbing bodies (Mercury–Neptune, Earth, Moon, EMB); requires an ephemeris source when non-empty |
| `Engine` / `Context` | Ephemeris source (`*spk.File`, `astro.Instance`, etc.); `Engine` takes precedence |
| `Step` / `Tolerance` | Maximum integration step (days) and adaptive relative tolerance |

> Zero-value options run the exact two-body path. For long high-precision spans, prefer JPL SPK files (see [Ephemeris Instances](ephemeris.md)).

### Visual Binary Position Angle & Separation

```go
gammaVir := orbit.VisualBinaryElements{
	PeriodYears: 171.37, PeriastronYear: 1836.433, Eccentricity: 0.8808,
	SemiMajorAxis: 3.746, Inclination: 146.05, AscendingNode: 31.78, PeriastronArgument: 252.88,
}
vb := orbit.VisualBinaryByYear(2026.0, gammaVir)
fmt.Printf("Gamma Vir (2026.0): PA %.2f°  sep %.3f\"\n", vb.PositionAngle, vb.Separation)
```

Output:

```text
Gamma Vir (2026.0): PA 353.10°  sep 3.383"
```

## API Reference

| Group | Entry Points | Notes |
| --- | --- | --- |
| Elements & anomalies | `Elements`, `MeanMotion`, `MeanAnomaly`, `TrueAnomaly` | Classic (`A/M0/EpochJD`, optional rates `ADot…MDot`) or perihelion (`Q/TpJD`) |
| Positions | `HeliocentricEcliptic[J2000]`, `GeocentricEcliptic[J2000]`, `GeocentricEquatorial[J2000]`, `AstrometricGeocentricEquatorialJ2000`, `ApparentGeocentricEcliptic`, `ApparentGeocentricEquatorial`, `ApparentTopocentricEquatorial` | Progressing from geometric -> light-time -> nutation -> topocentric parallax |
| Perturbations | `State`, `StateWithOptions`, `NewEphemerisEngine` | ICRF state & `astro.Engine` adapter |
| Geometry & photometry | `SunDistance`, `EarthDistance`, `Elongation`, `PhaseAngle`, `IlluminatedFraction` (`Phase`), `AsteroidMagnitudeHG` | AU, deg, fraction, H-G mag |
| Rise / set | `Altitude`, `Azimuth`, `Zenith`, `HourAngle`, `ParallacticAngle`, `CulminationTime`, `RiseTime`, `SetTime` | Sentinels: `ERR_ORBIT_NEVER_RISE` / `ERR_ORBIT_NEVER_SET` |
| Visual binaries | `VisualBinary`, `VisualBinaryByYear` | Position angle (deg, N→E) & separation (arcsec) |

## Conventions and Limitations

- **Refraction**: `Apparent*` and `Altitude` do not include atmospheric refraction; only `RiseTime`/`SetTime` with `aero=true` apply standard horizon refraction and dip.
- **`Phase` vs `PhaseAngle`**: `Phase` is an alias for `IlluminatedFraction` (`0..1`), not the phase angle.
- **Invalid elements**: Zero-value or invalid `Elements` silently yield `NaN` in position/anomaly calls; unsupported perturbers or combining EMB with Earth/Moon return `ErrUnsupportedPerturber`.
- **Linear rates**: `ADot`…`WDot` apply only to classic elliptical form; under perihelion form, only non-zero `MDot` overrides mean motion, and setting `EpochJD` chooses the osculating epoch for perturbed integration.


Related manuals: [Stars](star.md) | [Coordinates](coord.md) | [Ephemeris Instances](ephemeris.md)
