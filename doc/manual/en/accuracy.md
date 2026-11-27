# Accuracy & Performance

[中文](../accuracy.md) | [Back to README](../../../README.en.md)

This page summarizes the built-in ephemeris models, their measured accuracy against JPL DE440/DE441, and typical runtime order of magnitude for choosing between `ModelSemianalytic`, `ModelLegacy`, `Lite`, or an external `SPK` kernel.

## Contents

- [Choosing a Model](#choosing-a-model)
- [Ephemeris Composition](#ephemeris-composition)
- [Measurement Methodology](#measurement-methodology)
- [Geocentric Apparent Direction](#geocentric-apparent-direction)
- [Fixed ICRF Geometric Direction](#fixed-icrf-geometric-direction)
- [Geometric Earth-Moon Distance](#geometric-earth-moon-distance)
- [Legacy Truncation Reference](#legacy-truncation-reference)
- [Lite Approximation Chain](#lite-approximation-chain)
- [Performance](#performance)
- [Notes & Caveats](#notes--caveats)

## Choosing a Model

| Model / Package | Characteristics | Best For |
| --- | --- | --- |
| `ModelSemianalytic` (default) | Compact built-in tables, no external files; modern RMS `0.007″` Sun / `0.076″` Moon | Everyday positions, calendars, eclipses, occultations, phenomena & charts |
| `ModelLegacy` | Truncated VSOP87/ELP2000-82 + corrections (default under `-tags astro_legacy`) | Reproducing legacy values or classic-algorithm comparisons |
| `ModelCustom` (SPK / custom) | Reads JPL `.bsp` kernels or custom read-only engines | High-precision research, small-body kernels, long-span verification |
| `lite/sun`, `lite/moon` | Independent simplified formulas; unaffected by model switches | Constrained environments needing minute-level rise/set or lunar phase |

> See [Ephemeris Instances](ephemeris.md) for running multiple models in one process or loading SPK files.

## Ephemeris Composition

- **`ModelSemianalytic`** keeps the argument structure of VSOP2013 (inner planets), TOP2013 (outer planets), and ELP/MPP02 (Moon), refitted to DE441 over the long span and DE440 over the modern planetary span (J2000 ±1000 Julian years, blended over 100 years at both ends). Usable span: TT JD `-1931455.0 … 5373545.0` (~12,000 yr before to 8,000 yr after J2000). Embedded size: ~452 KiB (planets) + ~42 KiB (Moon).
- **`...N` functions**: `n < 0` evaluates the full series (matching non-`N` functions). Under `ModelSemianalytic`, any `n >= 0` uses a single fast table suitable only for coarse searches; only `ModelLegacy` truncates by `n`.
- **`ModelLegacy`** uses truncated VSOP87/ELP2000-82 series with long-term and per-body corrections.

## Measurement Methodology

All tables below compare against JPL DE440/DE441 on a TT grid (arcseconds), isolating ephemeris/frame accuracy from UTC/UT1/ΔT differences:

- **Modern span**: TT JD `2415020.5 … 2488434.25` (~1900–2100), 0.25-day step (`293,656` points).
- **Long span**: TT JD `990545.0 … 3912545.0` (J2000 ±4000 Julian years, `365,252` points).
- **Metrics**: `MAE` mean absolute error, `RMS` root-mean-square error, `P95` 95th percentile, `Max` sample maximum (not a continuous-time worst-case bound).

## Geocentric Apparent Direction

Sun (arcseconds):

| Span / Model | MAE | RMS | P95 | Max |
| --- | --- | --- | --- | --- |
| Modern DE440 / Legacy | `0.03895` | `0.04475` | `0.08013` | `0.1372` |
| Modern DE440 / Semianalytic full | `0.00556` | `0.00682` | `0.01329` | `0.02680` |
| Modern DE440 / Semianalytic fast | `0.1783` | `0.2219` | `0.4415` | `0.9675` |
| Long DE441 / Legacy | `2.659` | `4.171` | `9.751` | `16.68` |
| Long DE441 / Semianalytic full | `1.774` | `3.187` | `7.908` | `14.08` |
| Long DE441 / Semianalytic fast | `2.431` | `4.050` | `9.878` | `17.15` |

Moon (arcseconds):

| Span / Model | MAE | RMS | P95 | Max |
| --- | --- | --- | --- | --- |
| Modern DE440 / Legacy | `0.7431` | `0.7499` | `0.9171` | `1.3297` |
| Modern DE440 / Semianalytic full | `0.06302` | `0.07591` | `0.1437` | `0.4418` |
| Modern DE440 / Semianalytic fast | `3.455` | `3.991` | `7.092` | `18.18` |
| Long DE441 / Legacy | `42.16` | `67.10` | `156.35` | `233.58` |
| Long DE441 / Semianalytic full | `1.788` | `3.182` | `7.920` | `14.08` |
| Long DE441 / Semianalytic fast | `6.098` | `7.359` | `14.30` | `39.74` |

## Fixed ICRF Geometric Direction

Direct comparison of full semianalytic `Instance.State(target, BodyEarth, jd)` against DE in fixed ICRF (without of-date precession/nutation; arcseconds):

| Span / Body | MAE | RMS | P95 | Max |
| --- | --- | --- | --- | --- |
| Modern DE440 / Sun | `0.00551` | `0.00677` | `0.01320` | `0.02655` |
| Modern DE440 / Moon | `0.06301` | `0.07591` | `0.1437` | `0.4414` |
| Long DE441 / Sun | `0.04100` | `0.05754` | `0.1214` | `0.3613` |
| Long DE441 / Moon | `0.09163` | `0.1129` | `0.2170` | `0.8825` |

## Geometric Earth-Moon Distance

Geocentric distance error (km):

| Span / Model | MAE | P95 | Max |
| --- | --- | --- | --- |
| Modern DE440 / Legacy | `0.1434` | `0.3633` | `0.9143` |
| Modern DE440 / Semianalytic full | `0.04674` | `0.1172` | `0.3312` |
| Modern DE440 / Semianalytic fast | `1.730` | `4.356` | `9.818` |
| Long DE441 / Legacy | `2.204` | `9.834` | `26.90` |
| Long DE441 / Semianalytic full | `0.06496` | `0.1669` | `0.5576` |
| Long DE441 / Semianalytic fast | `2.005` | `5.160` | `16.10` |

## Legacy Truncation Reference

Reference truncation error of `ModelLegacy` relative to full VSOP87:

| Body | Lon / Lat | Distance |
| --- | --- | --- |
| Sun / Earth | `~0.1″` | `~0.1 × 10⁻⁶ AU` |
| Mercury, Venus | `~0.2″` | `~0.2 × 10⁻⁶ AU` |
| Mars | `~0.5″` | `~1 × 10⁻⁶ AU` |
| Jupiter | `~0.5″` | `~3 × 10⁻⁶ AU` |
| Saturn | `~0.5″` | `~5 × 10⁻⁶ AU` |
| Uranus | `~1″` | `~20 × 10⁻⁶ AU` |
| Neptune | `~1″` | `~40 × 10⁻⁶ AU` |

## Lite Approximation Chain

`lite/sun` and `lite/moon` use independent simplified formulas with fixed-step scan + bisection for rise/set (30 min Sun, 15 min Moon). Reference differences across 8 representative sites over a full year:

| Quantity | MAE | P95 | Max |
| --- | --- | --- | --- |
| Sun rise / set | `0.02 / 0.02 min` | `0.04 / 0.06 min` | `0.31 / 0.35 min` |
| Moon rise / set | `0.28 / 0.36 min` | `0.57 / 0.86 min` | `1.44 / 1.24 min` |
| `Phase()` | `0.00089` | `0.00185` | `0.00243` |
| `PhaseAge()` | `0.003 d` | `0.010 d` | `0.014 d` |

> High-latitude midnight-crossing dates may differ near polar boundaries. Use the main packages for eclipses, libration, and high-precision rise/set.

## Performance

Warm-cache ballpark timings on a typical desktop x86-64 machine (excluding one-time table load), showing relative cost and order of magnitude across models:

| Entry Point | Semianalytic full | Legacy full | Lite |
| --- | --- | --- | --- |
| Sun `ApparentRaDec` | `~0.6 µs` | `~9.0 µs` | `~0.2 µs` |
| Sun `Altitude` | `~0.8 µs` | `~9.1 µs` | `~0.6 µs` |
| Sun `RiseTime` | `~630 µs` | `~145 µs` | `~26 µs` |
| Moon `ApparentRaDec` | `~0.9 µs` | `~9.4 µs` | `~1.0 µs` |
| Moon `Phase` | `~1.2 µs` | `~15 µs` | `~0.9 µs` |
| Moon `Altitude` | `~1.0 µs` | `~9.5 µs` | `~1.1 µs` |
| Moon `RiseTime` | `~490 µs` | `~135 µs` | `~117 µs` |

- Single-epoch evaluations under `ModelSemianalytic` are sub-microsecond with zero heap allocations on hot paths.
- Rise/set searches evaluate many distinct epochs and run in the hundreds-of-microseconds range; actual latency varies with hardware and cache behavior.

## Notes & Caveats

- **Horizon definitions dominate rise/set differences**: Align refraction, semidiameter, observer ellipsoidal height, and local calendar-day rules before comparing rise/set outputs.
- **Compare catalogs on TT first**: Solar terms, phases, and eclipse contacts should be checked in TT; converting to civil UT/UTC adds whatever ΔT difference exists between sources (see [Time Scales](timescale.md)).
- **External SPK scope**: `Instance.State` takes TDB and returns ICRF km and km/day; high-level apparent APIs approximate TDB with TT and keep their own coordinate-frame, refraction, and terrain models.
- **Opt-in high-precision features**: IAU 2006/2000A, polar motion, CPO, and Kaguya lunar-limb terrain correction are never enabled automatically; see [Coordinates](coord.md), [Formulas](formula.md), and [Lunar Limb Correction](eclipse.md#lunar-limb-correction).
