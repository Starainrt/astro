# Ephemeris Instances

[中文](../ephemeris.md) | [Back to README](../../../README.en.md)

Package-level functions (`sun.RiseTime`, `mars.ApparentRaDec`, etc.) use the process default ephemeris (`ModelSemianalytic`). Use `astro.Context` and `astro.Instance` to run multiple ephemeris models side by side in one process, load external JPL SPK kernels (`.bsp`), or stack multiple ephemeris sources.

## Contents

- [Basic Example](#basic-example)
- [Usage](#usage)
  - [Load an External JPL SPK Kernel](#load-an-external-jpl-spk-kernel)
  - [Query Raw Geometric Position & Velocity (Instance.State)](#query-raw-geometric-position--velocity-instancestate)
  - [Multi-Source Stacking (Planetary + Small-Body Kernels)](#multi-source-stacking-planetary--small-body-kernels)
- [API Reference](#api-reference)
- [Conventions and Limitations](#conventions-and-limitations)

## Basic Example

Each `astro.Context` pins its ephemeris at creation time and exposes per-module accessors (`Sun()`, `Moon()`, `Eclipse()`, etc.) with the same signatures as the package-level functions:

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2026, 2, 17, 12, 0, 0, 0, cst)

	semi, err := astro.NewContext(astro.ModelSemianalytic) // default built-in model
	if err != nil {
		log.Fatal(err)
	}
	legacy, err := astro.NewContext(astro.ModelLegacy) // VSOP87/ELP2000-82
	if err != nil {
		log.Fatal(err)
	}

	ra1, dec1 := semi.Sun().ApparentRaDec(date)
	ra2, dec2 := legacy.Sun().ApparentRaDec(date)
	fmt.Printf("Sun RA/Dec: semianalytic (%.6f°, %.6f°)  legacy (%.6f°, %.6f°)\n", ra1, dec1, ra2, dec2)

	eDate := time.Date(2026, 3, 3, 12, 0, 0, 0, cst)
	le, ok := semi.Eclipse().LunarEclipseOnDate(eDate)
	fmt.Printf("2026-03-03 lunar eclipse: found=%v  type=%s  max=%s CST\n", ok, le.Type, le.Maximum.Format("15:04:05"))
}
```

Output:

```text
Sun RA/Dec: semianalytic (330.645545°, -11.998198°)  legacy (330.645548°, -11.998203°)
2026-03-03 lunar eclipse: found=true  type=total  max=19:33:42 CST
```

> To switch the process-wide default used by package-level functions, call `astro.SetModel(astro.ModelLegacy)`; existing contexts are unaffected.

## Usage

### Load an External JPL SPK Kernel

Open a `.bsp` file with `spk.Open` and pass it with `astro.ModelCustom` to `astro.NewContext`:

```go
kernel, err := spk.Open("de440s.bsp")
if err != nil {
	log.Fatal(err)
}
defer kernel.Close()

ctx, err := astro.NewContext(astro.ModelCustom, kernel)
if err != nil {
	log.Fatal(err)
}
ra, dec := ctx.Sun().ApparentRaDec(date)
alt := ctx.Moon().Altitude(date, 116.4, 39.9)
fmt.Printf("DE440s: Sun (%.6f°, %.6f°)  Moon alt %.4f°\n", ra, dec, alt)
```

> Custom engines never fall back silently to built-in tables on out-of-range epochs or errors (scalar APIs return `NaN`; event lookups report not found).

### Query Raw Geometric Position & Velocity (`Instance.State`)

For raw ICRF position (km) and velocity (km/day) at a TDB Julian date, use `astro.NewInstance(engine)` or `astro.NewModelInstance(astro.ModelSemianalytic)`:

```go
inst := astro.NewInstance(kernel)
jdTDB := calendar.Date2JD(astro.TDBFromTT(astro.TTFromUTC(date)))

pos, vel, err := inst.State(astro.BodyMoon, astro.BodyEarth, jdTDB)
startJD, endJD, ok := inst.Coverage(astro.BodyMoon)
```

- Out-of-range epochs return `astro.ErrEphemerisOutOfRange`.
- `ModelLegacy` has no 3D state engine and returns `astro.ErrEphemerisNoEngine`.

### Multi-Source Stacking (Planetary + Small-Body Kernels)

Small-body SPK files typically store only target-relative-to-Sun states. `astro.NewStackInstance(de440, ceres)` combines multiple sources in priority order and resolves center chains automatically:

```go
de440, _ := spk.Open("de440s.bsp")
ceres, _ := spk.Open("ceres.bsp")
defer de440.Close()
defer ceres.Close()

stack := astro.NewStackInstance(de440, ceres) // earlier sources take priority
jdTT := calendar.Date2JD(astro.TTFromUTC(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)))
dir, distKM, lightTimeDays, err := stack.Apparent(astro.Body(20000001), jdTT) // 20000001 = 1 Ceres
```

> You can also wrap an `orbit` propagator with `orbit.NewEphemerisEngine` and include it in a stack (see [Orbits](orbit.md)).

## API Reference

| Entry Point | Purpose |
| --- | --- |
| `astro.NewContext(model, engines...)` | Isolated context exposing `Sun()`, `Moon()`, `Mercury()`…`Neptune()`, `Eclipse()`, `Phenomena()`, `Satellite()`, `GeoJSON()`, etc. |
| `ctx.WithLunarLimb(limb)` | Derive a new context with Kaguya/custom lunar-limb correction enabled |
| `astro.SetModel(model)` | Switch process-wide default (`ModelSemianalytic` or `ModelLegacy`) |
| `astro.NewModelInstance` / `astro.NewInstance` | Low-level `*astro.Instance` for `State`, `Apparent`, and `Coverage` |
| `astro.NewStackInstance(engines...)` | Priority-ordered multi-source engine stack |
| `spk.Open(path)` | Read a JPL SPK `.bsp` file as an `astro.Engine` |

## Conventions and Limitations

- **Engine contract**: Implement `Engine.State(target, center, jdTDB)` returning ICRF km and km/day. Engines must be read-only and safe for concurrent calls (the library never spawns goroutines). Engines without optional `ApparentEngine.Apparent` automatically use the generic light-time, deflection, and aberration chain (requiring Earth, Sun, and target relative to `BodySSB`).
- **Scope of isolation**: Contexts isolate only the ephemeris; ΔT, TT−UTC, and future leap-hour policy remain process-wide settings.
- **Saros numbers**: Custom engines may model arbitrary trajectories, so eclipse searches leave `HasSaros == false`.
- **Binary size**: Selecting one model explicitly does not strip the other built-in table from the binary.
