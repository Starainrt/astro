# Formula Helpers

[中文](../formula.md) | [Back to README](../../../README.en.md)

`formula` holds pure formulas that read no ephemeris tables, covering radiation, photometry, telescopes, the atmosphere and Earth orientation for estimates and observing aids.

It performs no time-scale conversion and reads no global ephemeris. Most inputs are physical quantities such as temperature, wavelength, distance, aperture and altitude; `TIOLocatorArcsec` is the exception that accepts a TT Julian date.

Altitude and zenith-distance conventions follow [Observing-angle semantics](sun-moon.md#observing-angle-semantics); overall accuracy and applicability are in [Accuracy and performance](accuracy.md).

## Contents

- [Magnitude, synodic period and blackbody radiation](#magnitude-synodic-period-and-blackbody-radiation)
- [Refraction, extinction and Earth orientation](#refraction-extinction-and-earth-orientation)
- [API Reference](#api-reference)
  - [Blackbody and Radiation](#blackbody-and-radiation)
  - [Synodic Period](#synodic-period)
  - [Magnitude and Distance](#magnitude-and-distance)
  - [Distance Units](#distance-units)
  - [Telescope Metrics](#telescope-metrics)
  - [Stellar Parameter Conversions](#stellar-parameter-conversions)
  - [Airmass Models](#airmass-models)
- [Usage examples](#usage-examples)
  - [Blackbody peak, total flux and stellar parameters](#blackbody-peak-total-flux-and-stellar-parameters)
  - [Synodic period and magnitude/distance](#synodic-period-and-magnitudedistance)
  - [Telescope limiting magnitude and resolution](#telescope-limiting-magnitude-and-resolution)
  - [Comparing airmass models](#comparing-airmass-models)
- [Parameter and result conventions](#parameter-and-result-conventions)
  - [Units](#units)
  - [Time Scales](#time-scales)
  - [Angle Quadrants and Degree/Radian Boundaries](#angle-quadrants-and-degreeradian-boundaries)
  - [Zero Values and Invalid Input](#zero-values-and-invalid-input)
  - [Accuracy and Applicability](#accuracy-and-applicability)

## Refraction, extinction and Earth orientation

| Entry point | Input and output convention |
| --- | --- |
| `ApparentAltitude` / `TrueAltitude` | True/apparent altitude conversion in degrees; pressure hPa and temperature °C |
| `RefractionFromTrueAltitude` / `RefractionFromApparentAltitude` | Refraction in degrees, added to true altitude or subtracted from apparent altitude |
| `ExtinctionSchaefer` | V-band zenith extinction components and `Total`, in magnitudes per airmass |
| `TIOLocatorArcsec` | TT Julian date to the TIO locator s′ in arcseconds |
| `PolarMotionMatrix` | `xp`, `yp` and s′ in arcseconds; a matrix acting on column vectors from TIRS to ITRS |
| `CelestialPoleOffsetMatrix` | dX/dY in mas; a first-order small-angle correction matrix multiplying an existing celestial rotation on the left |

Abbreviations in the table: `TIO` is the terrestrial intermediate origin, `TIRS` and `ITRS` are the terrestrial intermediate and terrestrial reference systems, `s′` is the TIO locator, and `CPO` is the celestial pole offset (`dX`, `dY`) in `mas` (milliarcseconds).

```go
apparent := formula.ApparentAltitude(5, 1010, 10)
trueAltitude := formula.TrueAltitude(apparent, 1010, 10)
extinction := formula.ExtinctionSchaefer(100, 20, 50, 39.9, 120)
polarMotion := formula.PolarMotionMatrix(0.1, 0.2, formula.TIOLocatorArcsec(2460000.5))
_ = trueAltitude
_ = extinction.Total
_ = polarMotion
```

The standard refraction constants are 1010 hPa and 10 °C. Pressure must be positive, temperature above absolute zero and inputs finite. Both directions use the same model; a missing inverse returns `NaN`. Refraction is zero at altitudes ≤−5° or ≥90°. These are empirical-model boundaries, not a guarantee for precision near-horizon observing.

`ExtinctionSchaefer` uses site elevation in metres, unlike the ellipsoidal height of most observing APIs. Humidity is a percentage in `(0,100)`; latitude and solar right ascension are degrees. Assign `Total` to `phenomena.HeliacalOptions.ExtinctionCoefficient` to use it; when it is not assigned, the visibility calculation uses that option's own default extinction. Do not reuse a V-band coefficient for a different band without conversion.

Callers supply polar-motion and CPO parameters. The package downloads no IERS (International Earth Rotation and Reference Systems Service) data and does not automatically apply these corrections to default rise/set, eclipse or occultation calculations. The arcseconds of s′ and xp/yp differ from the mas of dX/dY; matrix elements are dimensionless.

## Magnitude, synodic period and blackbody radiation

```go
package main

import (
	"fmt"

	"github.com/starainrt/astro/formula"
)

func main() {
	// Empirical limiting magnitude for a 70 mm refractor at a site with naked-eye limit 6.
	fmt.Printf("limiting=%.6f\n", formula.LimitingMagnitudeEmpirical(70, 6))

	// Synodic period of Earth and Venus. Inputs and output are days.
	fmt.Printf("synodic=%.6f\n", formula.SynodicPeriod(365.25636, 224.70069))

	// Apparent magnitude of a Sun-like absolute-magnitude object at 100 pc.
	fmt.Printf("apparent=%.6f\n", formula.ApparentMagnitudeFromAbsolute(4.83, 100))

	// Treat the Sun as a 5772 K blackbody; compute peak wavelength and total radiant exitance.
	fmt.Printf("peak=%.9em flux=%.6e\n",
		formula.WienPeakWavelength(5772),
		formula.StefanBoltzmannFlux(5772),
	)
}
```

Output:

```text
limiting=11.000000
synodic=583.920635
apparent=9.830000
peak=5.020394932e-07m flux=6.293859e+07
```

## API Reference

The interfaces, units, and return values are listed below by what is computed.

The group snippets omit shared preamble: `fmt` is imported at the top of the file, and `formula` means `github.com/starainrt/astro/formula`.

### Blackbody and Radiation

| Name | Purpose | Units and conventions |
| --- | --- | --- |
| `PlanckRadianceByWavelength` | Planck spectral radiance by wavelength | wavelength meters, temperature K; returns W·sr⁻¹·m⁻³; non-positive temperature or wavelength returns NaN |
| `WienPeakWavelength` | Wien displacement peak wavelength | temperature K; returns meters; temperature `≤ 0` or non-finite returns NaN |
| `StefanBoltzmannFlux` | total radiant exitance per unit area | temperature K; returns W/m²; `0 K` is valid and returns `0` |
| `SolarEffectiveTemperature` | built-in solar effective temperature constant | no arguments; returns K, currently `5772` |

```go
// Treat the Sun as a 5772 K blackbody: peak wavelength, total exitance, and spectral radiance at 500 nm.
tSun := formula.SolarEffectiveTemperature()
fmt.Printf("peak=%.9e m flux=%.6e W/m^2\n",
	formula.WienPeakWavelength(tSun), formula.StefanBoltzmannFlux(tSun))
fmt.Printf("radiance@500nm=%.6e W·sr^-1·m^-3\n",
	formula.PlanckRadianceByWavelength(500e-9, tSun))
```

### Synodic Period

| Name | Purpose | Units and conventions |
| --- | --- | --- |
| `SynodicPeriod` | synodic period of two orbiting bodies | both inputs must share one unit and the output uses it; a period `≤ 0` or non-finite returns NaN, equal periods return `+Inf` |

```go
// Synodic periods of Earth with other planets; inputs and outputs are days.
earth := 365.25636
fmt.Printf("venus=%.6f mars=%.6f jupiter=%.6f\n",
	formula.SynodicPeriod(earth, 224.70069),
	formula.SynodicPeriod(earth, 686.980),
	formula.SynodicPeriod(earth, 4332.589))
```

### Magnitude and Distance

| Name | Purpose | Units and conventions |
| --- | --- | --- |
| `DistanceModulus` | distance modulus | distance pc; returns `m − M`; it is `0` at `10 pc`, and a distance `≤ 0` returns NaN |
| `ApparentMagnitudeFromAbsolute` | absolute magnitude plus distance to apparent magnitude | magnitude mag, distance pc; equals `M + distance modulus` |
| `AbsoluteMagnitudeFromApparent` | apparent magnitude plus distance to absolute magnitude | magnitude mag, distance pc; equals `m − distance modulus` |

```go
// Sun-like absolute magnitude 4.83 placed at 10 pc / 100 pc / 1 kpc, and the inverse.
for _, d := range []float64{10, 100, 1000} {
	m := formula.ApparentMagnitudeFromAbsolute(4.83, d)
	fmt.Printf("d=%.0f pc m=%.6f M=%.6f mod=%.6f\n",
		d, m, formula.AbsoluteMagnitudeFromApparent(m, d), formula.DistanceModulus(d))
}
```

### Distance Units

| Name | Purpose | Units and convention |
| --- | --- | --- |
| `Distance` | Converts parsecs, light-years or astronomical units to parsecs | A positive value plus a `DistanceUnit`; non-positive values, NaN and unknown units return NaN |

| Constant | Meaning |
| --- | --- |
| `DistanceParsec` | Parsecs, an identity conversion |
| `DistanceLightYear` | Light-years |
| `DistanceAU` | Astronomical units |

```go
// Sirius has a parallax of 0.375 arcseconds, i.e. 2.667 pc or 8.70 light-years.
pc := formula.Distance(1/0.375, formula.DistanceParsec)
fmt.Printf("%.3f pc = %.2f ly\n", pc, formula.Distance(pc, formula.DistanceParsec)/formula.Distance(1, formula.DistanceLightYear))
```

Convention: the astronomical unit is `149597870.7` km, the light-year is the IAU defined `9460730472580.8` km, and the parsec follows from the exact relation `648000/π` astronomical units, giving `1 pc = 3.261563777 ly`.

### Telescope Metrics

| Name | Purpose | Units and conventions |
| --- | --- | --- |
| `DawesLimitArcsec` | Dawes resolution limit | aperture mm; returns arcseconds from the empirical `116 / D` |
| `RayleighLimitArcsec` | Rayleigh resolution limit | aperture mm; returns arcseconds from the empirical `138.4 / D` |
| `LightGatheringPowerRatio` | light-gathering power ratio | two apertures in mm; returns `(D1 / D2)²`, dimensionless |
| `LimitingMagnitudeEmpirical` | empirical limiting magnitude | aperture mm, naked-eye limit mag; estimated as `naked-eye limit + 5·log10(D / 7)`, with 7 mm as the built-in dark-adapted pupil |

```go
// 70 mm refractor: resolution limits, light grasp relative to a 7 mm dark-adapted pupil, and empirical limit at naked-eye limit 6.
fmt.Printf("dawes=%.6f rayleigh=%.6f\n",
	formula.DawesLimitArcsec(70), formula.RayleighLimitArcsec(70))
fmt.Printf("power=%.6f limiting=%.6f\n",
	formula.LightGatheringPowerRatio(70, 7),
	formula.LimitingMagnitudeEmpirical(70, 6))
```

### Stellar Parameter Conversions

| Name | Purpose | Units and conventions |
| --- | --- | --- |
| `LuminosityFromRadiusTemperature` | radius plus temperature to luminosity | radius meters, temperature K; returns W via `4πR²σT⁴` |
| `LuminositySolarFromRadiusTemperature` | same, solar units | radius R☉, temperature K; returns L☉ |
| `RadiusFromLuminosityTemperature` | luminosity plus temperature to radius | luminosity W, temperature K; returns meters |
| `RadiusSolarFromLuminosityTemperature` | same, solar units | luminosity L☉, temperature K; returns R☉ |
| `EffectiveTemperatureFromLuminosityRadius` | luminosity plus radius to effective temperature | luminosity W, radius meters; returns K |
| `EffectiveTemperatureFromLuminositySolarRadius` | same, solar units | luminosity L☉, radius R☉; returns K |
| `SolarEffectiveTemperature` | built-in solar effective temperature | no arguments; returns K |

```go
// A 2.5 R☉, 20 L☉ main-sequence star: solve for temperature, then recompute luminosity and radius.
t := formula.EffectiveTemperatureFromLuminositySolarRadius(20, 2.5)
fmt.Printf("Teff=%.6f K\n", t)
fmt.Printf("L=%.6f Lsun R=%.6f Rsun\n",
	formula.LuminositySolarFromRadiusTemperature(2.5, t),
	formula.RadiusSolarFromLuminosityTemperature(20, t))
// The MKS version of the same quantities; the solar radius is the built-in constant 6.957e8 m.
rM := 2.5 * 6.957e8
lW := formula.LuminosityFromRadiusTemperature(rM, t)
fmt.Printf("L=%.6e W R=%.6e m Teff=%.6f K\n",
	lW, formula.RadiusFromLuminosityTemperature(lW, t),
	formula.EffectiveTemperatureFromLuminosityRadius(lW, rM))
```

### Airmass Models

| Name | Purpose | Units and conventions |
| --- | --- | --- |
| `AirmassPlaneParallel` | plane-parallel model | true altitude in degrees; equivalent to `sec(z)`, returns `+Inf` at `0°` |
| `AirmassPlaneParallelByZenithDistance` | plane-parallel model by zenith distance | zenith distance in degrees; returns `+Inf` at `90°` |
| `AirmassKastenYoung` | Kasten-Young 1989 | apparent altitude in degrees; more robust than `sec(z)` at low altitude |
| `AirmassPickering` | Pickering 2002 | apparent altitude in degrees; intended for low-altitude correction |

All four limit the input to `[0,90]` and return NaN outside that range or for non-finite input; altitude is measured from the horizon at `0°` to the zenith at `+90°`, zenith distance is its complement, and the convention is described under [Observing-angle semantics](sun-moon.md#observing-angle-semantics).

```go
fmt.Println(formula.AirmassPlaneParallel(30))
fmt.Println(formula.AirmassKastenYoung(5))
fmt.Println(formula.AirmassPickering(5))
fmt.Println(formula.AirmassPlaneParallelByZenithDistance(60))
```

If coordinate-layer refraction correction is not needed, `formula` also provides the three airmass models directly, with more direct input semantics:

- `AirmassPlaneParallel`: true altitude input, equivalent to the geometric `sec(z)` approximation
- `AirmassPlaneParallelByZenithDistance`: zenith-distance input
- `AirmassKastenYoung` / `AirmassPickering`: apparent-altitude input, no automatic refraction correction

## Usage examples

### Blackbody peak, total flux and stellar parameters

```go
fmt.Println(formula.WienPeakWavelength(5772))                            // peak wavelength (metres)
fmt.Println(formula.StefanBoltzmannFlux(5772))                           // flux per unit area (W/m^2)
fmt.Println(formula.RadiusSolarFromLuminosityTemperature(1, 5772))       // radius from luminosity and temperature
fmt.Println(formula.EffectiveTemperatureFromLuminositySolarRadius(1, 1)) // temperature from luminosity and radius
```

```text
5.020394932432432e-07
6.293859246828887e+07
1.0000011882005775
5772.003429145848
```

The three conversions invert each other and return the input under one convention (the `1 -> 1.0000012` and `5772 -> 5772.0034` residuals come from both sides rounding to 5772 K); the `...Solar` variants use solar units and the plain ones use SI.

### Synodic period and magnitude/distance

```go
fmt.Println(formula.SynodicPeriod(365.25636, 224.70069))      // Earth and Venus (days)
fmt.Println(formula.DistanceModulus(10))                      // distance modulus at 10 pc
fmt.Println(formula.ApparentMagnitudeFromAbsolute(4.83, 100)) // absolute magnitude 4.83 seen from 100 pc
fmt.Println(formula.AbsoluteMagnitudeFromApparent(4.83, 100)) // the inverse conversion
```

```text
583.9206352820089
0
9.83
-0.16999999999999993
```

`DistanceModulus(10)` is 0 because 10 pc is the defining distance of absolute magnitude; synodic periods take and return days, and the argument order does not matter.

### Telescope limiting magnitude and resolution

```go
fmt.Println(formula.DawesLimitArcsec(70), formula.RayleighLimitArcsec(70)) // both resolution limits at 70 mm
fmt.Println(formula.LightGatheringPowerRatio(200, 70))                     // 200 mm against 70 mm
fmt.Println(formula.LimitingMagnitudeEmpirical(70, 6))                     // limiting magnitude at 70 mm with a naked-eye limit of 6
```

```text
1.6571428571428573 1.9771428571428573
8.16326530612245
11
```

Dawes and Rayleigh differ by a coefficient (1.66" versus 1.98" at 70 mm), so state which one a report uses; the second argument of `LimitingMagnitudeEmpirical` is the naked-eye limit of the site and changes with it.

### Comparing airmass models

```go
for _, alt := range []float64{5, 30, 60, 90} {
	fmt.Printf("alt=%.0f KY=%.6f Pickering=%.6f plane=%.6f\n", alt,
		formula.AirmassKastenYoung(alt), formula.AirmassPickering(alt), formula.AirmassPlaneParallel(alt))
}
```

```text
alt=5 KY=10.305791 Pickering=10.333706 plane=11.473713
alt=30 KY=1.994293 Pickering=1.993154 plane=2.000000
alt=60 KY=1.153992 Pickering=1.154058 plane=1.154701
alt=90 KY=0.999712 Pickering=1.000000 plane=1.000000
```

The three models nearly coincide at moderate and high altitude and differ most at 5 degrees (Kasten-Young against the plane-parallel model is about 1.2 airmasses); the plane-parallel model diverges at 0 degrees, so use Kasten-Young or Pickering for careful low-altitude work.

Pressure- and temperature-corrected versions live in [Coordinate tools](coord.md#airmass) and this package keeps the raw formulas.

## Parameter and result conventions

### Units

- Blackbody family: wavelength meters, temperature kelvin; `PlanckRadianceByWavelength` returns spectral radiance `W·sr⁻¹·m⁻³`, `StefanBoltzmannFlux` returns `W/m²`, and `WienPeakWavelength` returns meters.
- Stellar family: the MKS variants use radius meters, luminosity watts, and temperature kelvin; the Solar variants use solar radius R☉, solar luminosity L☉, and temperature kelvin, and both inputs and outputs are dimensionless solar multiples.
- Magnitude family: distance parsecs, magnitude mag; `DistanceModulus` returns `m − M`, also in mag.
- Distance units: `Distance` performs unit conversion only; the input unit comes from `DistanceUnit` and the return value is always parsecs.
- Telescope family: aperture millimeters; `DawesLimitArcsec` and `RayleighLimitArcsec` return **arcseconds**, not degrees; `LightGatheringPowerRatio` and `LimitingMagnitudeEmpirical` return a dimensionless ratio and a magnitude respectively.
- Airmass family: altitude (or zenith distance) in degrees; the return value is the dimensionless relative airmass with 1 at the zenith.
- Synodic period: the unit is chosen by the caller, both inputs must match, and the output matches them; the package does not assume days.
- This package produces no apparent diameter or radius. Ordinary angles and refraction use degrees; Dawes/Rayleigh and the TIO locator use arcseconds, polar-motion inputs arcseconds and CPO inputs mas.

### Time Scales

- The package performs no time-scale conversion. `TIOLocatorArcsec` accepts a TT Julian date; the other formulas use caller-supplied physical quantities or periods without reading global time-scale settings.
- The synodic period is a length of time, not the instant of the next conjunction.

  Landing on a date requires the conjunction APIs in the planet packages plus civil time, whose convention is in the manual [Time Scale Conventions](timescale.md).
- The package has no ΔT, leap-second, or UT1 entry point; those only appear in instant-dependent chains such as `coord`, `eclipse`, and `moon`.

### Angle Quadrants and Degree/Radian Boundaries

- Ordinary angles are degrees, polar-motion angles arcseconds and CPO mas. Follow each function's declared units; internal radians never leak to the caller.
- Airmass models limit altitude and zenith distance to `[0,90]`, with no quadrant folding. Refraction functions have their own low-altitude boundary; do not apply the airmass range to them.

  Negative altitude is invalid for the airmass models rather than folded to the zenith or replaced by an absolute value.
- Zenith distance `z` and altitude `h` satisfy `z = 90° − h`; `AirmassPlaneParallel` takes `h` and `AirmassPlaneParallelByZenithDistance` takes `z`, and both must agree for the same geometry.
- Dawes/Rayleigh return arcseconds; divide by 3600 to compare with the degree-based angles in the other manuals.

### Zero Values and Invalid Input

- The blackbody family deliberately treats invalid input differently: `WienPeakWavelength` and `PlanckRadianceByWavelength` return NaN for temperature `≤ 0` or non-finite, while `StefanBoltzmannFlux` only returns NaN for temperature `< 0` or non-finite, so `0 K` is valid and returns `0` (a 0 K body has zero flux, while its peak wavelength is undefined).
- Synodic period: either period `≤ 0` or non-finite returns NaN; equal periods make the frequency difference zero and return `+Inf`.
- Magnitude family: `distanceParsec ≤ 0` or non-finite makes `DistanceModulus` return NaN, and the two conversion functions then return NaN as well; `DistanceModulus(10)` is always `0`.
- Telescope family: an aperture (or the second aperture) `≤ 0` or non-finite returns NaN; a zero second aperture in `LightGatheringPowerRatio` is rejected as invalid rather than dividing by zero; `LimitingMagnitudeEmpirical` takes no pupil argument, and 7 mm is a built-in constant.
- Stellar family: every input must be `> 0` or the function returns NaN; `EffectiveTemperatureFromLuminositySolarRadius`, `LuminositySolarFromRadiusTemperature`, and `RadiusSolarFromLuminosityTemperature` convert solar units to SI before computing.
- Airmass family: an altitude or zenith distance outside `[0,90]` (including negative values) or non-finite returns NaN; `AirmassPlaneParallel(0)` and `AirmassPlaneParallelByZenithDistance(90)` return `+Inf`, while `AirmassKastenYoung(0)` and `AirmassPickering(0)` stay finite.
- The package has no `(value, error)` or `(value, ok)` returns: invalid input is always expressed as NaN, which the caller checks with `math.IsNaN`.

### Accuracy and Applicability

- The blackbody family is an ideal-blackbody model with no absorption lines, interstellar extinction, or atmospheric extinction. `WienPeakWavelength` uses the Wien displacement constant `b = 2.897771955e-3 m·K`; because the per-frequency and per-wavelength peak conventions differ, `b/T` is strictly the peak in the wavelength convention.
- Constant conventions: `h = 6.62607015e-34 J·s`, `c = 299792458 m/s`, `k = 1.380649e-23 J/K`, `σ = 5.670374419e-8 W·m⁻²·K⁻⁴`; solar parameters `L☉ = 3.828e26 W`, `R☉ = 6.957e8 m`, `Teff = 5772 K`, the last exposed through `SolarEffectiveTemperature`.
- The magnitude family assumes no extinction, no K-correction, and no cosmological term; the two conversion functions merely add or subtract `DistanceModulus`, so their accuracy is entirely that of the externally supplied absolute magnitude and distance.
- The telescope family contains visible-light empirical values: Dawes and Rayleigh use fixed coefficients (`116` and `138.4`, aperture in mm) and do not vary with wavelength; `LightGatheringPowerRatio` compares aperture squares only and ignores central obstruction, transmission, and secondary-mirror losses; `LimitingMagnitudeEmpirical` ignores sky background, magnification, transmission, and observer skill.
- The stellar family solves `L = 4πR²σT⁴` in both directions, assuming spherical symmetry with no limb-darkening correction, rotation, or magnetic effects; the Solar and MKS variants share the same solar constants, so any difference between them comes only from that constant convention.
- Airmass family: the plane-parallel model is purely geometric `sec(z)`, usable only at moderate and high altitude, and it diverges near the horizon; Kasten-Young (1989) and Pickering (2002) are empirical fits that agree at moderate altitude and differ most near the horizon. With an apparent altitude already in hand, call `AirmassKastenYoung` / `AirmassPickering` directly;

  with only a true altitude plus refraction, use `coord`'s [airmass](coord.md#airmass).
- Every function in this package is a stateless pure function: it caches nothing and reads no global time-scale state, so it can be called in any order and, with caller-side synchronization, concurrently.
