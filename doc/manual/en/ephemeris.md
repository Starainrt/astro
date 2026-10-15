# Explicit ephemeris instances

[中文](../ephemeris.md) | [Back to README](../../../README.en.md)

Package-level functions use the current ephemeris model of the process, which is the semianalytic model by default. To use several models side by side in one process, or to plug in an external ephemeris (SPK, a home-grown engine), pin the ephemeris down with `astro.Context`; when geometric state (position and velocity) has to be read directly, drop down to `astro.Instance`.

## Contents

- [Switching to another built-in model](#switching-to-another-built-in-model)
- [Coordinates, altitude and rise/set with a Context](#coordinates-altitude-and-riseset-with-a-context)
- [Querying events by date](#querying-events-by-date)
- [Plugging in an external ephemeris](#plugging-in-an-external-ephemeris)
- [Stacking multiple sources](#stacking-multiple-sources)
- [Engine contract](#engine-contract)
- [Limits](#limits)

## Switching to another built-in model

`astro.NewContext` creates an independent context, and the model is fixed when the context is created:

```go
import (
	"fmt"
	"time"

	"github.com/starainrt/astro"
)

date := time.Date(2026, 2, 17, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))

semi, err := astro.NewContext(astro.ModelSemianalytic) // the model used by default builds
if err != nil {
	panic(err)
}
legacy, err := astro.NewContext(astro.ModelLegacy)
if err != nil {
	panic(err)
}

ra1, dec1 := semi.Sun().ApparentRaDec(date)
ra2, dec2 := legacy.Sun().ApparentRaDec(date)
fmt.Printf("%.8f %.8f\n", ra1, dec1)
fmt.Printf("%.8f %.8f\n", ra2, dec2)
```

```text
330.64554489 -11.99819781
330.64554785 -11.99820252
```

Once a context is created its model is fixed, and a later `astro.SetModel` does not affect it. `SetModel` changes the default model of the package-level functions and accepts only the two built-in models; any other model returns `astro.ErrUnknownModel` and leaves the global state unchanged. A custom engine must be passed explicitly to a `Context` or an `Instance`. `NewContext` returns `ErrUnknownModel` for an unknown model and `ErrInvalidEngine` when the engines do not fit the model (an engine passed with a built-in model, or `ModelCustom` without exactly one non-nil engine); `NewModelInstance` accepts only built-in models.

## Coordinates, altitude and rise/set with a Context

The methods returned by the accessors take the same arguments and return the same values as the package-level functions of the same name, so `sun.Xxx` becomes `ctx.Sun().Xxx`:

```go
ctx, err := astro.NewContext(astro.ModelSemianalytic)
if err != nil {
	panic(err)
}
date := time.Date(2026, 2, 17, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
lon, lat := 116.4, 39.9

// The Sun: geocentric apparent right ascension/declination, topocentric altitude/azimuth
ra, dec := ctx.Sun().ApparentRaDec(date)
alt := ctx.Sun().Altitude(date, lon, lat)
az := ctx.Sun().Azimuth(date, lon, lat)
fmt.Printf("Sun ra=%.6f dec=%.6f alt=%.6f az=%.6f\n", ra, dec, alt, az)

// The Moon: use GeocentricApparentRaDec for geocentric coordinates, topocentric quantities take the longitude and latitude directly
moonRa, moonDec := ctx.Moon().GeocentricApparentRaDec(date)
moonAlt := ctx.Moon().Altitude(date, lon, lat)
moonRise, err := ctx.Moon().RiseTime(date, lon, lat, 0, true)
fmt.Printf("Moon ra=%.6f dec=%.6f alt=%.6f rise=%s err=%v\n",
	moonRa, moonDec, moonAlt, moonRise.Format("15:04:05"), err)

// Planets are reached through accessors named after their packages
marsRa, marsDec := ctx.Mars().ApparentRaDec(date)
marsAlt := ctx.Mars().Altitude(date, lon, lat)
fmt.Printf("Mars ra=%.6f dec=%.6f alt=%.6f\n", marsRa, marsDec, marsAlt)
```

```text
Sun ra=330.645545 dec=-11.998198 alt=37.682918 az=171.221286
Moon ra=327.197499 dec=-14.621205 alt=34.596340 rise=06:59:57 err=<nil>
Mars ra=322.193276 dec=-16.023967 alt=34.060823
```

The instant is 2026-02-17 12:00 CST, and the site is 116.4° east and 39.9° north. `RiseTime` returns an error instead of a zero time when it finds no rise or set within the local civil day.

| Entry point | Package | Purpose |
| --- | --- | --- |
| `Sun()` | `sun` | Geocentric coordinates, topocentric altitude/azimuth/zenith distance, rise and set and twilight, apparent diameter, physical ephemeris |
| `Moon()` | `moon` | Geocentric and topocentric coordinates, altitude/azimuth, rise and set, lunar phase, geocentric distance, libration |
| `Mercury()` … `Neptune()` | The planet packages | Apparent coordinates, altitude/azimuth, rise and set, phase and apparent magnitude, events such as opposition, conjunction and greatest elongation |
| `Eclipse()` | `eclipse` | Solar and lunar eclipse queries, local contact times, eclipse paths |
| `Phenomena()` | `phenomena` | Ecliptic longitude crossings, stations, closest approach, rise and set at any altitude, heliacal visibility |
| `GeoJSON()` | `geojson` | Map data for celestial events |

The `...N` methods use the fast table, for example `ctx.Sun().AltitudeN(date, lon, lat, n)` and `ctx.Moon().TrueLoN(date, n)`; see [Accuracy and performance](accuracy.md) for the meaning of `n`.

## Querying events by date

Event queries also hang off the modules:

```go
date := time.Date(2026, 3, 3, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))

info, ok := ctx.Eclipse().LunarEclipseOnDate(date)
next := ctx.Eclipse().NextLunarEclipse(date)
fmt.Println(ok, info.Maximum.Format("2006-01-02 15:04 MST"), info.Type)
fmt.Println(next.Maximum.Format("2006-01-02 15:04 MST"), next.Type)
```

```text
true 2026-03-03 19:33 CST total
2026-03-03 19:33 CST total
```

`ok` is `true` only when a lunar eclipse occurs on that date; the example date 2026-03-03 has one. Filtering and searching both use the instance of that context. The Saros tables of the built-in models are selected according to the instance model; a custom engine can describe an arbitrary trajectory, so no Saros number is inferred and `HasSaros` is `false` in the result.

## Plugging in an external ephemeris

`spk.Open` reads a JPL SPK kernel and passes it to `NewContext` as an `astro.Engine`; afterwards the calls are exactly the same as with a built-in model:

```go
import "github.com/starainrt/astro/planet/spk"

kernel, err := spk.Open("de440s.bsp")
if err != nil {
	panic(err)
}
defer kernel.Close()

custom, err := astro.NewContext(astro.ModelCustom, kernel)
if err != nil {
	panic(err)
}

// Reuse the date above (2026-02-17 12:00 CST) and the longitude and latitude
ra, dec := custom.Sun().ApparentRaDec(date)
alt := custom.Sun().Altitude(date, lon, lat)
marsRa, marsDec := custom.Mars().ApparentRaDec(date)
marsAlt := custom.Mars().Altitude(date, lon, lat)
moonAlt := custom.Moon().Altitude(date, lon, lat)
fmt.Printf("Sun ra=%.6f dec=%.6f alt=%.6f\n", ra, dec, alt)
fmt.Printf("Mars ra=%.6f dec=%.6f alt=%.6f\n", marsRa, marsDec, marsAlt)
fmt.Printf("Moon alt=%.6f\n", moonAlt)
```

```text
Sun ra=330.645545 dec=-11.998198 alt=37.682918
Mars ra=322.193276 dec=-16.023967 alt=34.060823
Moon alt=34.596325
```

The results agree closely with the built-in semianalytic model at the same instant; here the lunar altitude differs by about `0.05″`.

Use `Instance` instead when the raw geometric state (position and velocity) is needed:

```go
inst := astro.NewInstance(kernel)
jdTDB := calendar.Date2JD(astro.TDBFromTT(date))
pos, vel, err := inst.State(astro.BodyMoon, astro.BodyEarth, jdTDB)
start, end, ok := inst.Coverage(astro.BodyMoon)
fmt.Printf("pos=%.1f,%.1f,%.1f km\nvel=%.1f,%.1f,%.1f km/day\nerr=%v coverage=%.1f..%.1f ok=%v\n",
	pos[0], pos[1], pos[2], vel[0], vel[1], vel[2], err, start, end, ok)
```

```text
pos=312305.1,-204135.4,-98208.0 km
vel=48600.2,64310.3,36543.8 km/day
err=<nil> coverage=2396752.5..2506352.5 ok=true
```

`State` takes TDB epochs and returns the ICRF geometric position (km) and velocity (km/day). `Coverage` gives the usable epoch interval of that body, and going outside it returns `ErrEphemerisOutOfRange`.

## Stacking multiple sources

`astro.NewStackInstance` composes several engines in argument order, and the source passed first has priority. Small-body kernels often provide only the state of the target relative to the Sun, and Stack composes a state of the target relative to the solar-system barycenter across sources:

```go
de440, _ := spk.Open("de440s.bsp")
ceres, _ := spk.Open("ceres.bsp")
defer de440.Close()
defer ceres.Close()

stack := astro.NewStackInstance(de440, ceres)

// A small-body kernel usually covers a narrow range, so the query epoch must fall inside its coverage
smallDate := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
jdTT := calendar.Date2JD(astro.TTFromUTC(smallDate))
dir, dist, tau, err := stack.Apparent(astro.Body(20000001), jdTT)
fmt.Printf("%.8f,%.8f,%.8f\ndist=%.0f km tau=%.5f day err=%v\n", dir[0], dir[1], dir[2], dist, tau, err)
```

```text
0.98097637,0.17021858,-0.09333272
dist=417020638 km tau=0.01610 day err=<nil>
```

`astro.Body` is an `int`, so any NAIF identifier can be written directly; Ceres is `20000001`. The Ceres kernel in this example covers only 2026-01-01 to 01-03 (`stack.Coverage` returns `2461041.5..2461043.5`); going outside the coverage returns `ErrEphemerisOutOfRange` and never silently switches to another source.

## Engine contract

- `Engine.State(target, center, jdTDB)`: ICRF geometric position (km) and velocity (km/day), with TDB epochs. A custom engine must implement it.
- The optional `ApparentEngine.Apparent(target, jdTT)`: geocentric ICRF unit apparent direction, distance (km) and light time (days) at a TT epoch. An engine that implements only `State` (such as SPK) goes through the generic vector apparent-place chain: iterate the light time, then add solar gravitational deflection and relativistic aberration; this requires the engine to provide states of the Earth, the Sun and the target relative to the solar-system barycenter (`BodySSB`).
- Engines are called concurrently and must be read-only and reentrant; the library does not start goroutines.
- A failure of a custom engine does not fall back to the built-in tables: scalar entry points return `NaN`, the `Maximum` of a lunar eclipse core is `NaN` and each phase flag is `false`, and date queries report a miss.
- Scalar coordinates and the `basic` lunar eclipse entry points take TT epochs; the geometric engine currently approximates TDB with TT, so call `Instance.State` directly when strict TDB is required.

## Limits

- `EphemerisLegacy` instances have scalar entry points only and no geometric engine: `Instance.State` returns `astro.ErrEphemerisNoEngine`, and `Coverage` and `Apparent` report unavailable. Use the semianalytic model or a custom engine when positions and velocities are needed.
- An explicit instance only decides which ephemeris is used, and does not guarantee a smaller binary: both built-in tables are linked into the program.
- The usable range of the semianalytic model is TT JD `-1931455.0 … 5373545.0`, and evaluation still extrapolates outside it; the modern planetary segment is JD `2086295.0 … 2816795.0`. See [Accuracy and performance](accuracy.md) for the errors.
- `Stack.Coverage` is the union of the source intervals, may contain gaps in between, and does not guarantee that every body needed by a center chain (the path of centers Stack searches for between its sources to compose a state) is available; rely on what `State` returns at a specific epoch.
- Most `basic` entry points that support instances carry the `WithInstance` suffix, and the `n`-parameterised ones carry `NWithInstance`; passing `nil` means the current global model of the process. The exceptions are the shadow-radius entry `LunarEclipseWithLunarLimb` and the 18 atmospheric rise/set entries, which take `*Instance` as their last parameter. Direct calls to these entry points are normally unnecessary.
