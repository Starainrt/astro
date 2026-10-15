# Accuracy and Performance

[中文](../accuracy.md) | [Back to README](../../../README.en.md)

Two ephemerides are built in, and their errors vary with model, epoch and calculated quantity. Values do not apply outside the model and range stated in each table header.

## Contents

- [Ephemeris models](#ephemeris-models)
- [Model selection](#model-selection)
- [Measurement conventions](#measurement-conventions)
- [Geocentric apparent direction](#geocentric-apparent-direction)
- [Fixed-ICRF geometric directions](#fixed-icrf-geometric-directions)
- [Geometric lunar distance](#geometric-lunar-distance)
- [Legacy truncation error](#legacy-truncation-error)
- [Lite lightweight chains](#lite-lightweight-chains)
- [Performance](#performance)
- [Conversions and caveats](#conversions-and-caveats)

## Ephemeris models

The library ships two ephemerides; the default is a **semianalytic model** (`ModelSemianalytic`): it keeps the argument structure (integer multipliers) and initial amplitudes of the planetary theories, then refits the amplitudes, fundamental-argument polynomials and a small number of correction terms to DE441, folding everything into a single series table. No residual layer is attached at runtime and no external ephemeris file is needed.

- **Base theories**: VSOP2013 for Mercury, Venus, Earth and Mars; TOP2013 for Jupiter, Saturn, Uranus and Neptune; ELP/MPP02 for the Moon.
- **Fitting truth**: DE441 for the long range; DE440 for the planetary modern segment.
- **Coverage**: TT JD `-1931455.0 … 5373545.0` (12000 Julian years before and 8000 after J2000); values outside it are extrapolated with no accuracy guarantee.
- **Modern segment** (planets only): a separate table fitted over J2000 ±1000 Julian years (TT JD `2086295.0 … 2816795.0`) against DE440, blended with the long-range table over 100 years at each end; use it for refinement at recent epochs.
- **Evaluation scale**: `n < 0` uses the full series, matching the APIs without an `N`; every `n >= 0` shares one fast table intended for coarse searches. Only Legacy truncates its series according to `n`.
- **Data size**: the planetary table is about 452 KiB and the lunar table about 42 KiB, compiled into the program; both are linked even when only some bodies are used, and explicit instances do not reduce the binary size.
- **Frame**: the tables hold geometric positions in the J2000 mean ecliptic; at evaluation they are rotated to the mean ecliptic of date and then follow the same apparent-place chain as Legacy (light time, aberration, precession and nutation).

`ModelLegacy` uses truncated VSOP87/ELP2000-82 series with correction terms added on top: a cubic polynomial for Earth and constant offsets for the other planets, and for the Moon an extra polynomial, a precession term and a fit for 3000–5000 CE. `lite/sun` and `lite/moon` use a different set of simplified formulas; do not mix the two.

## Model selection

| Model | Content | Use |
| --- | --- | --- |
| `ModelSemianalytic` | Built-in semianalytic tables, the model in default builds (without the `astro_legacy` tag) | Everyday positions, calendars, observing and event calculations |
| `ModelLegacy` | Truncated VSOP87/ELP2000-82 series with long-term and per-body corrections, `n` truncating by term count | Reproducing old values and comparing models; the default under an `astro_legacy` build |
| `ModelCustom` | SPK or a caller-supplied read-only engine | Accuracy determined jointly by the external kernel, its usable interval and the upper-level transforms |
| `lite/sun`, `lite/moon` | Independent approximation chains | Resource-constrained environments; do not follow model switches |

An explicit `Context`/`Instance` fixes the ephemeris, but ΔT, TT−UTC and civil time-scale policy remain process-wide settings. Signatures are unchanged while default values may change, which differs from older versions; select Legacy explicitly when old model values are needed. See [Explicit ephemeris instances](ephemeris.md).

## Measurement conventions

Unless stated otherwise, the data below are comparisons against DE440/DE441 on TT grid points, in arcseconds:

- The reference side uses Skyfield 1.55 and converts only TT to TDB, introducing no UTC, UT1 or ΔT differences. "Semianalytic full" in the tables means `n < 0`, and "semianalytic fast" means `n >= 0`.
- Modern range: TT JD `2415020.5 … 2488434.25`, step 0.25 day, `293656` points, approximately 1900–2100.
- Long range: TT JD `990545.0 … 3912545.0`, J2000 ±4000 Julian years, `365252` points; the two DE441 kernel segments switch at JD `2440416.5`.
- Geometric quantities use the mean ecliptic/equinox of date; apparent RA/Dec use the true equator/equinox of date.
- In the tables MAE is the mean absolute error, RMS the root-mean-square error, P95 the 95th percentile of the absolute error and Max the sampled maximum; angles are geocentric apparent direction angles, not errors in a single coordinate component. The maxima are sampled maxima, not continuous-time error bounds.

## Geocentric apparent direction

Sun:

| Range / model | MAE | RMS | P95 | Max |
| --- | --- | --- | --- | --- |
| Modern DE440 / Legacy | `0.03895` | `0.04475` | `0.08013` | `0.1372` |
| Modern DE440 / semianalytic full | `0.00556` | `0.00682` | `0.01329` | `0.02680` |
| Modern DE440 / semianalytic fast | `0.1783` | `0.2219` | `0.4415` | `0.9675` |
| Long DE441 / Legacy | `2.659` | `4.171` | `9.751` | `16.68` |
| Long DE441 / semianalytic full | `1.774` | `3.187` | `7.908` | `14.08` |
| Long DE441 / semianalytic fast | `2.431` | `4.050` | `9.878` | `17.15` |

Moon:

| Range / model | MAE | RMS | P95 | Max |
| --- | --- | --- | --- | --- |
| Modern DE440 / Legacy | `0.7431` | `0.7499` | `0.9171` | `1.3297` |
| Modern DE440 / semianalytic full | `0.06302` | `0.07591` | `0.1437` | `0.4418` |
| Modern DE440 / semianalytic fast | `3.455` | `3.991` | `7.092` | `18.18` |
| Long DE441 / Legacy | `42.16` | `67.10` | `156.35` | `233.58` |
| Long DE441 / semianalytic full | `1.788` | `3.182` | `7.920` | `14.08` |
| Long DE441 / semianalytic fast | `6.098` | `7.359` | `14.30` | `39.74` |

In the modern samples the semianalytic full series has an apparent-direction RMS of `0.00682″` (Sun) and `0.07591″` (Moon), below the Legacy `0.04475″` and `0.7499″`; the fast table gives `0.2219″` and `3.991″`, worse than Legacy, especially for the Moon. Long-range results include the whole upper-level transformation chain and reference-frame differences, so they cannot be attributed entirely to the series themselves.

## Fixed-ICRF geometric directions

The following table compares the semianalytic full `Instance.State(target, BodyEarth, jd)` against DE geocentric ICRF geometric vectors directly, without date precession/nutation transforms. The library approximates TDB numerically with TT, and the reference also converts the same TT to TDB, so this set is a comparison under approximate TDB. Values are arcseconds:

| Range / body | MAE | RMS | P95 | Sample maximum |
| --- | --- | --- | --- | --- |
| Modern DE440 / Sun | `0.00551` | `0.00677` | `0.01320` | `0.02655` |
| Modern DE440 / Moon | `0.06301` | `0.07591` | `0.1437` | `0.4414` |
| Long DE441 / Sun | `0.04100` | `0.05754` | `0.1214` | `0.3613` |
| Long DE441 / Moon | `0.09163` | `0.1129` | `0.2170` | `0.8825` |

In the long range the fixed-frame errors are much smaller than the upper-level apparent date-frame differences, showing that the error of the whole calculation chain cannot be summarized by a single "model accuracy" number. Legacy has no geometric engine and is not listed in this table.

## Geometric lunar distance

Values are km, as geocentric distances:

| Range / model | MAE | P95 | Max |
| --- | --- | --- | --- |
| Modern DE440 / Legacy | `0.1434` | `0.3633` | `0.9143` |
| Modern DE440 / semianalytic full | `0.04674` | `0.1172` | `0.3312` |
| Modern DE440 / semianalytic fast | `1.730` | `4.356` | `9.818` |
| Long DE441 / Legacy | `2.204` | `9.834` | `26.90` |
| Long DE441 / semianalytic full | `0.06496` | `0.1669` | `0.5576` |
| Long DE441 / semianalytic fast | `2.005` | `5.160` | `16.10` |

## Legacy truncation error

The table below is a truncation-error reference for Legacy against complete VSOP87; it is neither semianalytic accuracy nor total error against DE440/DE441:

| Target | Longitude / latitude | Distance |
| --- | --- | --- |
| Sun / Earth | about `0.1″` | about `0.1 × 10^-6 AU` |
| Mercury, Venus | about `0.2″` | about `0.2 × 10^-6 AU` |
| Mars | about `0.5″` | about `1 × 10^-6 AU` |
| Jupiter | about `0.5″` | about `3 × 10^-6 AU` |
| Saturn | about `0.5″` | about `5 × 10^-6 AU` |
| Uranus | about `1″` | about `20 × 10^-6 AU` |
| Neptune | about `1″` | about `40 × 10^-6 AU` |

## Lite lightweight chains

`lite/sun` uses simplified solar formulas and `lite/moon` an approximate lunar chain with a small number of perturbation terms; their rise/set searches use a fixed-step scan plus bisection, 30 minutes for the Sun and 15 minutes for the Moon. Neither uses the main-chain tables and neither follows model switches. The following are errors over all of 2026 at 8 sites relative to the Legacy main chain (statistics as in [Measurement conventions](#measurement-conventions)):

| Quantity | Mean absolute error | P95 | Max |
| --- | --- | --- | --- |
| Sunrise / sunset | `0.02 / 0.02 min` | `0.04 / 0.06 min` | `0.31 / 0.35 min` |
| Moonrise / moonset | `0.28 / 0.36 min` | `0.57 / 0.86 min` | `1.44 / 1.24 min` |
| `Phase()` | `0.00089` | `0.00185` | `0.00243` |
| `PhaseAge()` | `0.003 d` | `0.010 d` | `0.014 d` |

These samples are not all semantically identical in their events: sunsets have 2 high-latitude date-attribution differences across midnight and moonset has 1 high-latitude disagreement about whether the event occurs that day. Rise/set are sampled every 7 or 15 days and phase/age every 6 hours. The time differences in the table only indicate agreement for events that do occur; they cannot be used to decide whether an event exists. Eclipses, physical libration and high-latitude event boundaries are computed by the main chain, not by the lite chains.

## Performance

Linux/WSL, Intel i7-12700, Go 1.25.9, benchmark concurrency 20, 300 ms per benchmark. Calls repeat the same instant and site, so these are warm-cache measurements and exclude first-load cost:

| Entry point | Semianalytic full | Legacy full | Lite |
| --- | --- | --- | --- |
| Sun `ApparentRaDec` | `0.606 µs` | `8.988 µs` | `0.213 µs` |
| Sun `Altitude` | `0.770 µs` | `9.134 µs` | `0.627 µs` |
| Sun `RiseTime` | `628.6 µs` | `144.8 µs` | `25.88 µs` |
| Moon `ApparentRaDec` | `0.855 µs` | `9.373 µs` | `1.003 µs` |
| Moon `Phase` | `1.171 µs` | `14.81 µs` | `0.914 µs` |
| Moon `Altitude` | `1.006 µs` | `9.520 µs` | `1.136 µs` |
| Moon `RiseTime` | `487.4 µs` | `135.0 µs` | `117.4 µs` |

The warm-cache semianalytic position APIs are faster than Lite in some cases and slower in others; rise/set searches evaluate at a new epoch each time and therefore run slower, while Legacy is in fact faster for some rise/set entry points. These are absolute single-run values and vary with machine and call shape.

Semianalytic Sun/Moon `RiseTime` allocate 60/36 times respectively, and the remaining evaluation APIs in the table allocate 0 times. Zero allocations cover only these paths and are not guaranteed for other APIs.

## Conversions and caveats

- Before comparing rise/set differences, align the refraction model, apparent-radius convention, observer height and local day attribution; the horizon definition often matters more than the ephemeris difference.
- Compare solar terms, syzygies and eclipse contacts directly as TT event instants; before turning positional arcseconds into a time error, divide by the local angular rate at that point.
- Civil dates also depend on the time zone and time-scale policy, and events near midnight are especially sensitive.
- Custom SPK geometric states use ICRF, km, km/day and TDB; high-level scalar and apparent-place entry points accept TT and currently approximate TDB with TT. Call `Instance.State` directly when strict TDB is required; switching to a more accurate kernel does not automatically improve the upper-level coordinate frame, atmosphere or terrain models.
- IAU 2006/2000A, polar motion, CPO and the lunar limb profile all require explicit calls or explicit settings and are not enabled automatically in the default chain; see [Coordinate tools](coord.md), [Formula helpers](formula.md) and [Eclipses and transits](eclipse.md#lunar-limb-correction).
