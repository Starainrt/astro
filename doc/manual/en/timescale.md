# Time scales

[中文](../timescale.md) | [README](../../../README.en.md)

UTC, UT1 and TT express the same physical instant using different scales. UTC is civil time, UT1 follows Earth's rotation, and TT is used for ephemerides. The relevant differences are `ΔT = TT − UT1` and `DUT1 = UT1 − UTC`, both in seconds.

## Contents

- [Time arguments](#time-arguments)
- [time.Time conversion API](#timetime-conversion-api)
- [Julian-day conversion API](#julian-day-conversion-api)
- [TCG, TCB and TDB](#tcg-tcb-and-tdb)
- [ΔT models](#δt-models)
  - [Comparing models](#comparing-models)
  - [Supplying an external model](#supplying-an-external-model)
- [UTC assumptions beyond the observed interval](#utc-assumptions-beyond-the-observed-interval)
- [SVG, GeoJSON and KML labels](#svg-geojson-and-kml-labels)

## Time arguments

Most observing APIs accept a civil instant as `time.Time`, obtain its UTC value, and convert to UT1 or TT as needed. Changing `Location` changes the displayed clock reading.

Date-based searches such as rise/set also use it to identify the local civil date.

Before 1972-01-01, this library treats civil time as UT1. This is a library convention, not a claim that UTC did not exist. From 1972 onward, the leap-second table, selected policy or explicit override supplies TT−UTC, while the ΔT model supplies TT−UT1.

Some inputs have separate conventions:

| Input | Interpretation |
| --- | --- |
| Argument to `astro.TTFromUTC` / `UT1FromUTC` | Civil instant in any time zone |
| Argument to `astro.UTCFromTT` / `UTCFromUT1` | TT / UT1 reading carried in a UTC Location |
| Argument to `astro.TCGFromTT` / `TCBFromTT` / `TDBFromTT` | TT reading, not a civil UTC instant |
| `orbit.Elements.EpochJD` / `TpJD` | TT/TDB Julian day; see [Orbits](orbit.md) |
| Numerical JD arguments in `basic` | Scale specified by the function contract; the number carries no scale metadata |
| `calendar.Date2JD` / `basic.Date2JD` | Reads calendar and clock fields without converting zones; pass `date.UTC()` for a UTC JD |
| Chinese calendar conversion | Date semantics described in [Calendars](calendar.md), using Beijing time by default |

`time.Time` cannot represent the leap-second reading `23:59:60`. Historical dates also require care: the library's Julian/Gregorian transition differs from Go's proleptic Gregorian calendar.

## time.Time conversion API

These functions belong to the root package `github.com/starainrt/astro`.

| Function | Result |
| --- | --- |
| `TTFromUTC(date)` | TT reading for the same instant |
| `UTCFromTT(tt)` | Civil instant corresponding to a TT reading |
| `UT1FromUTC(date)` | UT1 reading for the same instant |
| `UTCFromUT1(ut1)` | Civil instant corresponding to a UT1 reading |
| `DUT1(date)` | UT1−UTC in seconds |
| `LabelIn(scale, date)` | Original input for `TimeScaleUTC`; UT1 reading for `TimeScaleUT1` |
| `TCGFromTT(tt)` / `TTFromTCG(tcg)` | Convert between TT and Geocentric Coordinate Time |
| `TCBFromTT(tt)` / `TTFromTCB(tcb)` | Convert between TT and Barycentric Coordinate Time using a geocentric approximation |
| `TDBFromTT(tt)` / `TTFromTDB(tdb)` | Convert between TT and Barycentric Dynamical Time using a geocentric approximation |
| `TCBFromTDB(tdb)` / `TDBFromTCB(tcb)` | Linear conversions between TDB and TCB |
| `TCGMinusTT(tt)` / `TCBMinusTT(tt)` / `TDBMinusTT(tt)` | Offset from TT in seconds; input is a TT reading |

Converted TT, UT1, TCG, TCB and TDB values use `time.Time` and a UTC Location, but their fields are readings on the named scale. Passing them to an API expecting civil time, such as the Sun and Moon functions, would shift the calculation instant again.

Use the inverse conversion to recover civil time.

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro"
)

func main() {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	tt := astro.TTFromUTC(date)
	ut1 := astro.UT1FromUTC(date)
	fmt.Println("UTC:", date.Format(time.RFC3339Nano))
	fmt.Println("TT:", tt.Format("2006-01-02 15:04:05.000000"))
	fmt.Println("UT1:", ut1.Format("2006-01-02 15:04:05.000000"))
	fmt.Printf("DUT1: %.6f s\n", astro.DUT1(date))
	fmt.Println("UTC from TT:", astro.UTCFromTT(tt).Format(time.RFC3339Nano))
}
```

```text
UTC: 2026-04-01T00:00:00Z
TT: 2026-04-01 00:01:09.184014
UT1: 2026-04-01 00:00:00.051015
DUT1: 0.051000 s
UTC from TT: 2026-04-01T00:00:00Z
```

Conversions pass through floating-point Julian days, so round trips can have small rounding differences. The TT/UT1 display formats omit `Z` to avoid presenting them as UTC timestamps.

## Julian-day conversion API

The following `basic` functions use `float64`. Conversion functions return Julian days; offset functions return seconds.

| Function | Input → output | Notes |
| --- | --- | --- |
| `UTC2TT` | UTC JD → TT JD | Uses UT1 before 1972, then the TT−UTC model |
| `TT2UTC` | TT JD → UTC JD | Inverse conversion |
| `UT12TT` | UT1 JD → TT JD | Uses the active ΔT model |
| `TT2UT1` | TT JD → UT1 JD | Inverse conversion |
| `UTC2UT1` | UTC JD → UT1 JD | Equivalent to `TT2UT1(UTC2TT(jd))` |
| `UT12UTC` | UT1 JD → UTC JD | Inverse conversion |
| `TTMinusUTCSeconds` | UTC JD → seconds | `32.184 + (TAI−UTC)` in the built-in leap-second interval, unless overridden |
| `DUT1Seconds` | UTC JD → seconds | `(TT−UTC) − ΔT` |
| `DeltaT` | JD or decimal year → seconds | Second argument `true`: UT Julian day; `false`: decimal year |
| `TT2TCG` / `TCG2TT` | TT JD ↔ TCG JD | Linear conversion |
| `TT2TCB` / `TCB2TT` | TT JD ↔ TCB JD | Geocentric approximation |
| `TT2TDB` / `TDB2TT` | TT JD ↔ TDB JD | Geocentric approximation |
| `TCB2TDB` / `TDB2TCB` | TCB JD ↔ TDB JD | Linear conversion including TDB0 |
| `TCGMinusTTSeconds` / `TCBMinusTTSeconds` / `TDBMinusTTSeconds` | TT JD → seconds | Offset from TT |

A leap second or simulated leap hour introduces a step. The inverse TT-to-civil conversion has an unrepresentable interval of that width, so pointwise round-trip identity cannot hold across it.

`TTMinusUTCSeconds` and `DefaultTTMinusUTC()` expose the leap-table value (or the override for the former); they do not apply the future-policy extrapolation used by `UTC2TT`. For future civil-to-TT conversion, call `UTC2TT` or `TTFromUTC` directly.

## TCG, TCB and TDB

TCG is Geocentric Coordinate Time. TCB and TDB are Barycentric Coordinate Time and Barycentric Dynamical Time, referring to the Solar System barycenter rather than the center of the Sun. TT/TCG and TCB/TDB have defining linear relations. The TT/TDB relation implemented here is a geocentric approximation, without observer-dependent diurnal terms.

With `T0 = 2443144.5003725`, `LG = 6.969290134e-10`, `LB = 1.550519768e-8` and `TDB0 = −65.5e-6` seconds, the relations for JD readings are:

```text
TCG − TT = LG / (1 − LG) × (TT − T0)
TDB = TCB − LB × (TCB − T0) + TDB0 / 86400
TCB − TT = [LB × (TT − T0) + (TDB − TT) − TDB0 / 86400] / (1 − LB)
```

The reference epoch T0 does not imply that all four scales have equal readings there. Near modern dates, the main annual TDB−TT term has an amplitude of about 1.7 ms. TCG−TT increases by about 0.022 seconds per year and TCB−TT by about 0.489 seconds per year.

TDB−TT uses a 40-term truncated series with mass adjustments. Over years −3000 to +6000, the sum of omitted absolute amplitudes gives a conservative truncation-error bound of 11 µs relative to the full 787-term geocentric series. Sampling every 31 days gives a maximum difference of about 2.0 µs. A sampled maximum is not an all-time guarantee, and truncation error excludes the error of the full model itself. No accuracy is promised outside that interval.

A single `float64` JD has a resolution of about 40 µs near modern dates; the `time.Time` wrappers pass through the same rounding. For microsecond-level offsets, use `*MinusTT` or `*MinusTTSeconds` rather than subtracting two full Julian days and converting to seconds.

## ΔT models

The default combines SMH2016 and Morrison 2021 splines and long-term extrapolation, using a monthly observed ΔT table wherever available. The table currently extends to October 1, 2026; the September 1 point is final, while the last point is a preliminary extension. Constant endpoint adjustments keep extrapolation continuous.

`DeltaT` returns `NaN` outside ±40000 years.

| Constant | Model |
| --- | --- |
| `DeltaTModelDefault` / `DeltaTModelSMH2016` | Built-in default |
| `DeltaTModelMS2004` | Morrison & Stephenson 2004: `ΔT = −20 + 32u²`, `u = (year−1820)/100` |
| `DeltaTModelEspenakMeeus2006` | Espenak & Meeus 2006 piecewise polynomials |
| `DeltaTModelNASACanon2006` | Those polynomials plus the NASA canon pairing term `−0.000012932(year−1955)²` before 1955 |
| `DeltaTModelManual` | Status indicating an injected function; not an installable named model |

`SetDeltaTModel(model, keepObserved)` changes the process-wide model and reports success. Unknown models leave the current state unchanged. With `keepObserved=true`, observations take precedence; `false` uses the selected model everywhere.

`GetDeltaTModel()` returns both settings.

### Comparing models

`DeltaTModelSeconds` evaluates a model without changing process state:

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
)

func main() {
	date := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	jd := basic.Date2JD(date)
	for _, model := range []astro.DeltaTModel{
		astro.DeltaTModelSMH2016,
		astro.DeltaTModelEspenakMeeus2006,
		astro.DeltaTModelNASACanon2006,
	} {
		fmt.Println(model, astro.DeltaTModelSeconds(model, jd, false))
	}
}
```

```text
smh2016 86.98252070239269
espenak_meeus_2006 202.7448354310582
nasa_canon_2006 202.7448354310582
```

Future ΔT is uncertain. The current Espenak–Meeus 2006 and default models differ by about 11, 21, 116 and 275 seconds in 2035, 2050, 2100 and 2200. This affects future eclipse times and ground paths.

Match ΔT models before comparing eclipse catalogs.

### Supplying an external model

| Root-package function | Purpose |
| --- | --- |
| `DeltaT()` / `SetDeltaT(fn)` | Get/set a ΔT function of type `func(float64, bool) float64` |
| `DefaultDeltaT()` | Obtain the built-in default ΔT function |
| `TTMinusUTC()` / `SetTTMinusUTC(fn)` | Get/set a TT−UTC override of type `func(float64) float64`; argument is a civil JD |
| `DefaultTTMinusUTC()` | Built-in leap-second-table function, ignoring overrides and future policy |

Both callbacks return seconds. `SetDeltaT(nil)` restores default ΔT; `SetTTMinusUTC(nil)` restores the built-in leap-second table and future policy. `TTMinusUTC()` returns `nil` when no override is installed.

The corresponding `basic` functions are `GetDeltaTFn` / `SetDeltaTFn` and `GetTTMinusUTCFn` / `SetTTMinusUTCFn`. They share root-package state. A TT−UTC override takes precedence over future policy for queries from 1972 onward; dates before 1972 retain the UT1 convention.

These settings affect subsequent calculations throughout the process. Configure them before calculating. Restore the original function or named model after temporary comparisons; a model change is not a per-call option.

## UTC assumptions beyond the observed interval

> **UTC stops inserting leap seconds on 20 May 2027 (leap hour replaces leap second)**
>
> On 15 October 2026 the 28th General Conference on Weights and Measures (CGPM) adopted [Resolution 5, "On the technical actions needed to ensure the continuity of UTC"](https://www.bipm.org/en/cgpm-2026/resolution-5) (`Resolution-CGPM-28-5`, DOI `10.59161/CGPM2026RES5E`): **continuous UTC becomes effective on 20 May 2027** and the maximum value of `|UT1−UTC|` is widened from 0.9 seconds to **3600 seconds (1 hour)**.
>
> [IERS Bulletin C 72 (2026-07-06)](https://datacenter.iers.org/data/latestVersion/bulletinC.txt) confirms that **no leap second will be introduced at the end of December 2026**; the leap-second mechanism ends on 2027-05-20, so the possible insertion point at the end of June 2027 will not carry one either.
>
> This library therefore changes the default policy beyond the built-in window (which ends 2026-10-01) to `TimeScaleLeapHour`: 1972–2026 still uses the definitive leap-second table and those historical values are unchanged; past the window TT−UTC stays at 69.184 seconds until `|UT1−UTC|` approaches one hour, centuries from now. The 1900–2026 span of the modern lunisolar fast tables is consequently bit-identical as well.

`SetTimeScaleFuturePolicy` and `GetTimeScaleFuturePolicy` select the civil-time conversion policy, defaulting to `TimeScaleLeapHour`. All policies apply only after the observed interval except `TimeScaleUT1Civil`, which replaces the scale throughout the timeline. An explicit TT−UTC override takes precedence from 1972 onward. This setting is separate from the output-scale options `TimeScaleUTC` and `TimeScaleUT1`.

| Policy | Assumption |
| --- | --- |
| `TimeScaleLeapHour` (zero value, default) | Applies the fewest whole-hour corrections to bring DUT1 within ±3600 seconds; matches the continuous UTC effective 2027-05-20 |
| `TimeScaleLeapSecond` | Applies the fewest integer-second corrections to the final TT−UTC offset to bring extrapolated DUT1 within ±0.9 seconds; the definitive 1972–2026 convention, kept for reproduction and comparison |
| `TimeScaleAssumeUT1Tracking` | Holds the final observed DUT1 constant, allowing TT−UTC to follow extrapolated ΔT smoothly |
| `TimeScaleFreezeUTCOffset` | Holds TT−UTC at its final built-in value, currently 69.184 seconds |
| `TimeScaleUT1Civil` | Treats civil time as UT1 throughout the timeline, including the historical leap-table interval; DUT1 is zero unless explicitly overridden |

The default policy `TimeScaleLeapHour` follows the continuous UTC that takes effect on 2027-05-20, so it is no longer a scenario assumption; `TimeScaleLeapSecond` is kept to reproduce the definitive 1972–2026 convention or for comparison. Whole-hour corrections depend on ΔT at the queried instant; they do not track prior corrections or restrict steps to announcement calendar boundaries. Pointwise round trips need not be identical near a step. A stepping policy returns `NaN` when its ΔT is invalid or outside the model range, rather than falling back to a normal offset.

Under `TimeScaleLeapHour` the `|UT1−UTC|` tolerance is 3600 seconds, so TT−UTC past the window stays at 69.184 seconds for a very long time; `TimeScaleLeapSecond` still extrapolates against ±0.9 seconds and reproduces the old one-to-two-year stepping, which no longer reflects reality. Precise civil timing requires published data for the relevant date.

## SVG, GeoJSON and KML labels

SVG and GeoJSON default to civil labels. For UT1, use a `...InUT1` result-conversion function or set `TimeScale: astro.TimeScaleUT1` in export options. UT1 output requires `Location` to be `nil` or `time.UTC`.

Changing a label scale leaves existing geometry at the same physical instant. Time markers aligned to whole ticks of the selected scale may sample different instants, so their positions can move. GeoJSON records the scale in `time_scale`.

KML requires UTC `<when>` values, so the converter changes UT1 labels back to UTC using the active model and preserves the original `time_scale` property.

Use the same time-scale model when generating GeoJSON and converting it to KML. Export options, marker steps and KML playback are described in [Maps and data export](map-geojson.md).
