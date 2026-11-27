# Time Scales

[中文](../timescale.md) | [Back to README](../../../README.en.md)

Standard observing APIs accept civil `time.Time` values—the library handles UTC, UT1, and TT internally. This manual covers converting time scales explicitly, switching ΔT models, or configuring post-2027 leap-hour policy.

## Contents

- [Basic Example](#basic-example)
- [Usage](#usage)
  - [Compare or Switch ΔT Models](#compare-or-switch-δt-models)
  - [Future UTC Leap-Hour Policy (Effective 2027-05-20)](#future-utc-leap-hour-policy-effective-2027-05-20)
  - [UTC Calendar Date from Any Time Zone (EarthUTCDate)](#utc-calendar-date-from-any-time-zone-earthutcdate)
  - [UT1 Labels in SVG, GeoJSON & KML](#ut1-labels-in-svg-geojson--kml)
- [API Reference](#api-reference)
- [Conventions and Limitations](#conventions-and-limitations)

## Basic Example

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

	fmt.Printf("UTC:    %s\n", date.Format("2006-01-02 15:04:05.000"))
	fmt.Printf("TT:     %s  (TT-UTC = %.3f s)\n", tt.Format("2006-01-02 15:04:05.000"), tt.Sub(date).Seconds())
	fmt.Printf("UT1:    %s  (DUT1   = %+.3f s)\n", ut1.Format("2006-01-02 15:04:05.000"), astro.DUT1(date))
	fmt.Printf("TDB-TT: %+.3f ms\n", astro.TDBMinusTT(tt)*1000)
}
```

Output:

```text
UTC:    2026-04-01 00:00:00.000
TT:     2026-04-01 00:01:09.184  (TT-UTC = 69.184 s)
UT1:    2026-04-01 00:00:00.051  (DUT1   = +0.051 s)
TDB-TT: +1.630 ms
```

> **Caution**: `TTFromUTC`, `UT1FromUTC`, and `TDBFromTT` return `time.Time` values whose fields hold **TT / UT1 / TDB readings** (with UTC location). Do not pass them back into civil observing APIs like `sun` or `moon`, or the offset will be applied twice; convert back with `UTCFromTT` / `UTCFromUT1` first.

## Usage

### Compare or Switch ΔT Models

`ΔT = TT − UT1` tracks Earth's irregular rotation and directly shifts distant eclipse timings and ground paths:
- Compare models without changing global state: `astro.DeltaTModelSeconds(model, jd, keepObserved)`.
- Switch the process-wide model (e.g. to match NASA's 2006 eclipse canon): `astro.SetDeltaTModel(astro.DeltaTModelNASACanon2006, true)`.
- Inject custom callbacks: `astro.SetDeltaT(fn)` / `astro.SetTTMinusUTC(fn)` (pass `nil` to restore defaults).

```go
jd2100 := basic.Date2JD(time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC))
fmt.Printf("2100 ΔT: SMH2016=%.1f s  Espenak-Meeus=%.1f s\n",
	astro.DeltaTModelSeconds(astro.DeltaTModelSMH2016, jd2100, false),
	astro.DeltaTModelSeconds(astro.DeltaTModelEspenakMeeus2006, jd2100, false))
```

Output:

```text
2100 ΔT: SMH2016=87.0 s  Espenak-Meeus=202.7 s
```

Built-in models: `DeltaTModelDefault` / `DeltaTModelSMH2016` (monthly observed table through 2026-10-01 + SMH2016/Morrison 2021 spline), `DeltaTModelEspenakMeeus2006`, `DeltaTModelNASACanon2006`, and `DeltaTModelMS2004`.

> ΔT, TT−UTC, and future policies are process-wide settings; configure them during startup rather than per request.

### Future UTC Leap-Hour Policy (Effective 2027-05-20)

Under CGPM Resolution 5 ([Resolution-CGPM-28-5](https://www.bipm.org/en/cgpm-2026/resolution-5)), continuous UTC takes effect on 2027-05-20 with `|UT1−UTC|` widened to 3600 seconds (leap hours instead of leap seconds), and [IERS Bulletin C 72](https://datacenter.iers.org/data/latestVersion/bulletinC.txt) confirms no leap second at the end of 2026.

The 1972–2026 leap-second table is unchanged; beyond the built-in window, the default policy is `TimeScaleLeapHour` (keeping `TT−UTC = 69.184 s` for centuries). Switch via `astro.SetTimeScaleFuturePolicy`:

| Policy | Assumption Beyond Observed Window |
| --- | --- |
| `TimeScaleLeapHour` (default, zero value) | Leap-hour mode (`\|DUT1\| ≤ 3600 s`), matching continuous UTC from 2027-05-20 |
| `TimeScaleLeapSecond` | Legacy leap-second simulation (`\|DUT1\| ≤ 0.9 s`) for historical comparison |
| `TimeScaleAssumeUT1Tracking` | Hold terminal DUT1; TT−UTC tracks ΔT smoothly |
| `TimeScaleFreezeUTCOffset` | Freeze TT−UTC at the terminal table value (`69.184 s`) |
| `TimeScaleUT1Civil` | Treat civil time as UT1 across all epochs (`DUT1 = 0`) |

### UTC Calendar Date from Any Time Zone (`EarthUTCDate`)

`astro.EarthUTCDate(instant)` returns midnight UTC of the UTC calendar date containing `instant` (useful for zones like UTC+14):

```go
z14 := time.FixedZone("UTC+14", 14*3600)
inst := time.Date(2026, 6, 4, 1, 30, 0, 0, z14)
fmt.Printf("EarthUTCDate(2026-06-04 01:30 UTC+14) = %s\n", astro.EarthUTCDate(inst).Format("2006-01-02"))
```

Output:

```text
EarthUTCDate(2026-06-04 01:30 UTC+14) = 2026-06-03
```

### UT1 Labels in SVG, GeoJSON & KML

Set `TimeScale: astro.TimeScaleUT1` (requiring `Location` to be `nil` or `time.UTC`), or call `...InUT1` helpers on eclipse/occultation structs. KML `<when>` must be UTC, so `kml.FromGeoJSON` converts UT1 back to UTC while preserving `time_scale`. See [Time-Scale Declarations](map-geojson.md#time-scale-declarations).

## API Reference

| Package | Entry Points | Purpose |
| --- | --- | --- |
| `astro` (`time.Time`) | `TTFromUTC`, `UTCFromTT`, `UT1FromUTC`, `UTCFromUT1`, `DUT1` | Civil UTC ↔ TT / UT1 and DUT1 (sec) |
| `astro` (coordinate times) | `TCGFromTT`, `TCBFromTT`, `TDBFromTT`, `TCBFromTDB` (and inverses) | TT ↔ TCG / TCB / TDB (inputs are TT readings) |
| `astro` (microsecond diffs) | `TCGMinusTT`, `TCBMinusTT`, `TDBMinusTT` | Direct differences in seconds (avoids ~40 µs JD rounding) |
| `astro` (configuration) | `SetDeltaTModel`, `GetDeltaTModel`, `DeltaTModelSeconds`, `SetDeltaT`, `SetTTMinusUTC`, `SetTimeScaleFuturePolicy` | Process-wide ΔT and UTC policy |
| `astro` (helpers) | `LabelIn`, `EarthUTCDate` | Time-scale labeling & UTC date truncation |
| `basic` (float64 JD) | `UTC2TT`, `TT2UTC`, `UT12TT`, `TT2UT1`, `UTC2UT1`, `UT12UTC`, `DeltaT`, `DUT1Seconds`, `TTMinusUTCSeconds`, `TT2TDB`, etc. | Low-level Julian-day conversions |

## Conventions and Limitations

- **Pre-1972 convention**: Civil times before `1972-01-01` are treated as UT1; later dates use the leap-second table and future policy.
- **Sub-millisecond precision**: A single `float64` JD has ~40 µs resolution near modern dates—always use `*MinusTT` / `basic.*MinusTTSeconds` instead of subtracting two JDs. `TDB−TT` uses a 40-term geocentric series (~11 µs truncation bound over `-3000..+6000`).
- **Leap seconds**: Go's `time.Time` cannot represent `23:59:60`; for strict TT timestamps across historical positive leap seconds (e.g. planetary surface events), use `basic.SurfaceEvents` / `basic.NextSurfaceEvent`.
