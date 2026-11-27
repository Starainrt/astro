# Formulas

[中文](../formula.md) | [Back to README](../../../README.en.md)

The `formula` package contains stateless astronomical formulas that do not read ephemeris tables: telescope metrics, magnitudes & distance conversions, synodic periods, blackbody radiation, stellar parameters, atmospheric refraction/extinction/airmass, and Earth-orientation matrices.

## Contents

- [Basic Example](#basic-example)
- [Usage](#usage)
  - [Refraction, Extinction & Airmass](#refraction-extinction--airmass)
  - [Magnitudes, Distances & Stellar Parameters](#magnitudes-distances--stellar-parameters)
  - [Airmass Models](#airmass-models)
  - [Earth Orientation: Polar Motion & CPO](#earth-orientation-polar-motion--cpo)
- [API Reference](#api-reference)
- [Conventions and Limitations](#conventions-and-limitations)

## Basic Example

```go
package main

import (
	"fmt"

	"github.com/starainrt/astro/formula"
)

func main() {
	// 1. 70 mm refractor under mag-6 naked-eye skies
	fmt.Printf("70mm refractor: Dawes %.2f\"  Rayleigh %.2f\"  limiting mag %.2f\n",
		formula.DawesLimitArcsec(70),
		formula.RayleighLimitArcsec(70),
		formula.LimitingMagnitudeEmpirical(70, 6))

	// 2. Earth–Venus synodic period (same units in and out; days here)
	fmt.Printf("Earth-Venus synodic period: %.2f days\n",
		formula.SynodicPeriod(365.25636, 224.70069))

	// 3. Apparent magnitude of a Sun-like star (M=4.83) at 100 pc
	fmt.Printf("Sun at 100 pc: mag %.2f\n",
		formula.ApparentMagnitudeFromAbsolute(4.83, 100))

	// 4. 5772 K blackbody peak wavelength and flux
	fmt.Printf("5772K blackbody: peak %.1f nm  flux %.3e W/m²\n",
		formula.WienPeakWavelength(5772)*1e9,
		formula.StefanBoltzmannFlux(5772))
}
```

Output:

```text
70mm refractor: Dawes 1.66"  Rayleigh 1.98"  limiting mag 11.00
Earth-Venus synodic period: 583.92 days
Sun at 100 pc: mag 9.83
5772K blackbody: peak 502.0 nm  flux 6.294e+07 W/m²
```

## Usage

### Refraction, Extinction & Airmass

- True ↔ apparent altitude: `ApparentAltitude(trueAlt, pressureHPa, tempC)` / `TrueAltitude(appAlt, pressureHPa, tempC)` (typical defaults: `1010 hPa`, `10 °C`).
- Schaefer V-band zenith extinction: `ExtinctionSchaefer(elevM, tempC, humidityPct, latDeg, sunRADeg)` returns `Total` (mag/airmass), which can be assigned directly to `phenomena.HeliacalOptions.ExtinctionCoefficient`.

```go
ext := formula.ExtinctionSchaefer(100, 20, 50, 39.9, 120)
fmt.Printf("V-band zenith extinction: %.3f mag/airmass  airmass @ 5° (KY): %.2f\n",
	ext.Total, formula.AirmassKastenYoung(5))
```

Output:

```text
V-band zenith extinction: 0.391 mag/airmass  airmass @ 5° (KY): 10.31
```

> `ExtinctionSchaefer` takes site **altitude above sea level (m)** and humidity in `(0, 100)`.

### Magnitudes, Distances & Stellar Parameters

- Distance conversion to parsecs: `formula.Distance(val, unit)` with `DistanceParsec`, `DistanceLightYear`, or `DistanceAU`.
- Magnitudes: `DistanceModulus(pc)`, `ApparentMagnitudeFromAbsolute(M, pc)`, `AbsoluteMagnitudeFromApparent(m, pc)`.
- Stefan–Boltzmann stellar scaling (`L = 4πR²σT⁴`): available in both SI (m, W, K) and solar units (`...Solar...`, R☉, L☉, K).

### Airmass Models

Pure-formula airmass models without automatic refraction (input altitude/zenith distance in `[0°, 90°]`):

| Function | Input | Notes |
| --- | --- | --- |
| `AirmassPlaneParallel` / `AirmassPlaneParallelByZenithDistance` | True altitude / zenith distance | Geometric `sec(z)`; diverges (`+Inf`) at `0°` |
| `AirmassKastenYoung` | Apparent altitude | Kasten-Young (1989), reliable at low altitude |
| `AirmassPickering` | Apparent altitude | Pickering (2002), tailored for low-altitude observing |

> If you only have true altitude and want refraction applied automatically first, use [Coordinates: Airmass](coord.md#airmass).

### Earth Orientation: Polar Motion & CPO

For external IERS EOP parameters:
- `TIOLocatorArcsec(jdTT)`: TIO locator `s′` in arcseconds.
- `PolarMotionMatrix(xpArcsec, ypArcsec, sPrimeArcsec)`: TIRS → ITRS column-vector rotation.
- `CelestialPoleOffsetMatrix(dXMas, dYMas)`: first-order celestial pole offset matrix (`dX`/`dY` in **mas**).

> The library does not download IERS data and does not apply polar motion/CPO automatically to default pipelines.

## API Reference

| Group | Entry Points | Units |
| --- | --- | --- |
| Blackbody | `PlanckRadianceByWavelength`, `WienPeakWavelength`, `StefanBoltzmannFlux`, `SolarEffectiveTemperature` | Wavelength m, temp K → `W·sr⁻¹·m⁻³` / m / `W/m²` / K |
| Synodic | `SynodicPeriod(p1, p2)` | Same units in and out |
| Magnitudes & distance | `DistanceModulus`, `ApparentMagnitudeFromAbsolute`, `AbsoluteMagnitudeFromApparent`, `Distance` | Parsecs, mag |
| Telescopes | `DawesLimitArcsec`, `RayleighLimitArcsec`, `LightGatheringPowerRatio`, `LimitingMagnitudeEmpirical` | Aperture mm → resolution in **arcseconds**, ratio, mag |
| Stellar scaling | `Luminosity[Solar]FromRadiusTemperature`, `Radius[Solar]FromLuminosityTemperature`, `EffectiveTemperatureFromLuminosity[Solar]Radius` | SI (m / W / K) or solar units (R☉ / L☉ / K) |
| Refraction & extinction | `ApparentAltitude`, `TrueAltitude`, `RefractionFromTrueAltitude`, `RefractionFromApparentAltitude`, `ExtinctionSchaefer` | Angles in deg; pressure hPa, temp °C |
| Airmass | `AirmassPlaneParallel*`, `AirmassKastenYoung`, `AirmassPickering` | Deg, dimensionless (`1` at zenith) |
| Earth orientation | `TIOLocatorArcsec`, `PolarMotionMatrix`, `CelestialPoleOffsetMatrix` | `xp`/`yp`/`s′` arcsec; `dX`/`dY` mas |

## Conventions and Limitations

- **Invalid inputs**: Functions return `NaN` rather than `error` (check with `math.IsNaN`), except defined limits `StefanBoltzmannFlux(0) == 0`, `SynodicPeriod(p, p) == +Inf`, and `AirmassPlaneParallel(0) == +Inf`.
- **Refraction bounds**: Saemundsson refraction is active only for true altitudes in `(-5°, 90°)` and returns `0` outside; `TrueAltitude` returns `NaN` if no inverse exists.
- **Constants**: Exact 2019 SI values (`h`, `c`, `k`, `σ`), IAU nominal solar constants (`L☉ = 3.828e26 W`, `R☉ = 6.957e8 m`, `Teff = 5772 K`, `1 AU = 149597870.7 km`, `1 ly = 9460730472580.8 km`).
