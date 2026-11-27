# Sky Phenomena

[中文](../phenomena.md) | [Back to README](../../../README.en.md)

The `phenomena` package provides generic sky-event searches across bodies: ecliptic-longitude crossings, oppositions/conjunctions/quadratures, stationary points, closest approaches (including stars), apsides, orbital nodes, rise/set at any altitude, heliacal visibility, and airless observations from the surface of Mercury through Neptune.

## Contents

- [Basic Example](#basic-example)
- [Usage](#usage)
  - [Longitude Crossings, Aspects & Stations](#longitude-crossings-aspects--stations)
  - [Perihelion / Aphelion & Orbital Nodes](#perihelion--aphelion--orbital-nodes)
  - [Crossing Any Altitude (Twilight / Golden Hour)](#crossing-any-altitude-twilight--golden-hour)
  - [Heliacal Rising & Setting](#heliacal-rising--setting)
  - [Planetary Surface Observations](#planetary-surface-observations)
- [API Reference](#api-reference)
- [Conventions and Limitations](#conventions-and-limitations)

## Basic Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/phenomena"
)

func main() {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)

	// 1. Solar longitude crossing 0° (March equinox)
	vernal := phenomena.LongitudeCrossings(phenomena.Sun, 0, start, end)
	fmt.Printf("2026 March equinox: %s UTC\n", vernal[0].Time.Format("2006-01-02 15:04:05"))

	// 2. Jupiter–Moon approaches within 5° (Jan–Feb 2026)
	approaches := phenomena.ClosestApproaches(phenomena.Jupiter, phenomena.Moon, 5,
		start, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))
	for _, e := range approaches {
		fmt.Printf("Jupiter-Moon approach: %s UTC  sep %.2f°\n", e.Time.Format("2006-01-02 15:04:05"), e.Value)
	}

	// 3. Mars oppositions (2025–2028)
	opps := phenomena.AspectCrossings(phenomena.Mars, phenomena.Sun, 180,
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2028, 1, 1, 0, 0, 0, 0, time.UTC))
	for _, e := range opps {
		fmt.Printf("Mars opposition: %s UTC\n", e.Time.Format("2006-01-02 15:04:05"))
	}
}
```

Output:

```text
2026 March equinox: 2026-03-20 14:45:57 UTC
Jupiter-Moon approach: 2026-01-03 23:14:28 UTC  sep 3.61°
Jupiter-Moon approach: 2026-01-31 03:39:08 UTC  sep 3.78°
Jupiter-Moon approach: 2026-02-27 07:32:19 UTC  sep 3.91°
Mars opposition: 2025-01-16 02:38:35 UTC
Mars opposition: 2027-02-19 15:50:55 UTC
```

> Results preserve the time zone of `start`. Empty slices indicate no match in the window. Use `ctx.Phenomena()` to bind an explicit ephemeris instance.

## Usage

### Longitude Crossings, Aspects & Stations

- `LongitudeCrossings(body, deg, start, end)`: geocentric apparent ecliptic longitude crossings (every 15° of the Sun gives the solar terms).
- `AspectCrossings(a, b, angle, start, end)`: ecliptic longitude difference `λ(a) - λ(b)` crossing `angle` (`0°` conjunction, `180°` opposition, `90°`/`270°` quadrature).
- `Stations(body, start, end)`: planetary stationary points (`Event.Value` is apparent longitude in degrees).
- `StarApproaches(body, star, maxSep, start, end)`: closest approaches to a fixed star (`StarCoordinateByName` / `StarCoordinateByHR` load from the built-in catalog).

```go
for _, s := range phenomena.Stations(phenomena.Mercury,
	time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)) {
	fmt.Printf("Mercury station: %s UTC (lon %.2f°)\n", s.Time.Format("2006-01-02 15:04:05"), s.Value)
}
```

Output:

```text
Mercury station: 2026-02-26 06:48:10 UTC (lon 352.57°)
Mercury station: 2026-03-20 19:32:51 UTC (lon 338.49°)
```

### Perihelion / Aphelion & Orbital Nodes

- `Apsides(body, start, end)`: heliocentric distance extrema (`Perihelion` / `Aphelion`) for `Earth` and the seven planets.
- `Nodes(body, start, end)`: heliocentric ecliptic-latitude crossings (`AscendingNode` / `DescendingNode`).

```go
for _, a := range phenomena.Apsides(phenomena.Earth,
	time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)) {
	label := "Aphelion"
	if a.Kind == phenomena.Perihelion {
		label = "Perihelion"
	}
	fmt.Printf("Earth %-10s: %s UTC  %.6f AU\n", label, a.Time.Format("2006-01-02 15:04:05"), a.DistanceAU)
}
```

Output:

```text
Earth Perihelion: 2026-01-03 17:15:43 UTC  0.983302 AU
Earth Aphelion  : 2026-07-06 17:30:19 UTC  1.016644 AU
```

### Crossing Any Altitude (Twilight / Golden Hour)

`AltitudeCrossing` finds when the Sun, Moon, or a planet crosses any target altitude on a local date—useful for twilight (`-6°`/`-12°`/`-18°`), golden hour (`+6°`), or obstructed horizons (use `StarAltitudeCrossing` for stars):

```go
cst := time.FixedZone("CST", 8*3600)
day := time.Date(2026, 6, 21, 0, 0, 0, 0, cst)

res, err := phenomena.AltitudeCrossing(phenomena.Sun, day, 116.4, 39.9, 0, -6, phenomena.AltitudeOptions{})
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Beijing solstice civil twilight (-6°): start %s  end %s\n",
	res.Rise.Format("15:04:05"), res.Set.Format("15:04:05"))
```

Output:

```text
Beijing solstice civil twilight (-6°): start 04:11:26  end 20:20:51
```

- Check `res.State` (`AltitudeCrosses`, `AltitudeAlwaysAbove`, `AltitudeAlwaysBelow`) and `HasRise` / `HasSet`.
- Default options use body center with no refraction; enable `LimbUpper`/`LimbLower` and `Refraction: true` as needed.

### Heliacal Rising & Setting

`PlanetHeliacalEvents` and `StarHeliacalEvents` estimate first/last twilight visibility windows:

```go
name := map[phenomena.HeliacalEventKind]string{
	phenomena.HeliacalMorningFirst: "Morning first (heliacal rising)",
	phenomena.HeliacalMorningLast:  "Morning last",
	phenomena.HeliacalEveningFirst: "Evening first",
	phenomena.HeliacalEveningLast:  "Evening last (heliacal setting)",
}
for _, e := range phenomena.PlanetHeliacalEvents(phenomena.Mercury,
	time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), 116.4, 39.9, phenomena.HeliacalOptions{}) {
	fmt.Printf("Mercury %-31s: %s UTC  alt %.1f°  sun %.1f°  margin %.2f mag\n",
		name[e.Kind], e.Time.Format("2006-01-02 15:04"), e.TargetAltitude, e.SunAltitude, e.Margin)
}
```

Output:

```text
Mercury Evening first                  : 2026-02-08 10:20 UTC  alt 3.7°  sun -8.0°  margin 0.10 mag
Mercury Evening last (heliacal setting): 2026-02-25 10:50 UTC  alt 5.6°  sun -10.2°  margin 0.44 mag
```

### Planetary Surface Observations

Compute local solar/sidereal time, target altitude/azimuth, and airless rise/set/culmination from the surface of Mercury through Neptune (1 bar reference layer for gas giants). Example at Gale Crater on Mars (`137.4°E, -4.6°N`):

```go
obs := phenomena.BodyObserver{Body: phenomena.Mars, EastLongitude: 137.4, Latitude: -4.6}
t := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)

lt, err := phenomena.SurfaceLocalTimeAt(obs, t)
if err != nil {
	log.Fatal(err)
}
sunRise, err := phenomena.NextSurfaceRise(phenomena.Sun, obs, t, phenomena.SurfaceOptions{})
if err != nil {
	log.Fatal(err)
}
earthPos, err := phenomena.SurfacePositionAt(phenomena.Earth, obs, sunRise.Time)
if err != nil {
	log.Fatal(err)
}

fmt.Printf("Gale Crater LTST: %.2f h  LMST: %.2f h  solar day: %.5f Earth days\n",
	lt.LocalTrueSolarTime, lt.LocalMeanSolarTime, lt.SolarDayLength)
fmt.Printf("Next Mars sunrise: %s UTC (LTST %.2f h, az %.1f°)\n",
	sunRise.Time.Format("2006-01-02 15:04:05"), sunRise.Local.LocalTrueSolarTime, sunRise.Position.Azimuth)
fmt.Printf("Earth from Mars:   alt %.1f°  az %.1f°  dist %.3f AU\n",
	earthPos.Altitude, earthPos.Azimuth, earthPos.DistanceAU)
```

Output:

```text
Gale Crater LTST: 21.99 h  LMST: 22.43 h  solar day: 1.02749 Earth days
Next Mars sunrise: 2026-11-14 08:15:55 UTC (LTST 6.04 h, az 80.9°)
Earth from Mars:   alt -35.7°  az 66.7°  dist 1.322 AU
```

- Use `SurfaceEvents` for all rise/set/culmination events in an interval (and `StarSurface*` for stars). For Mars MSD/MTC shortcuts, see [Planetary Local Time & Jupiter Rotation Angle](planets.md#planetary-local-time--jupiter-rotation-angle).

## API Reference

| Group | Entry Points | Notes |
| --- | --- | --- |
| Bodies | `Sun`, `Moon`, `Mercury` … `Neptune`, `Earth` | `Earth` is valid for `Apsides` and planetary surface targets |
| Ecliptic & stations | `LongitudeCrossings`, `AspectCrossings`, `Stations` | Returns `[]Event` (`Time`, angle `Value` in deg) |
| Approaches | `ClosestApproaches`, `StarApproaches` | Local separation minima `≤ maxSeparation` (deg) |
| Orbital events | `Apsides`, `Nodes` | Returns `[]OrbitalEvent` with `DistanceAU` |
| Altitude crossings | `AltitudeCrossing`, `StarAltitudeCrossing` | Rise/set crossing any altitude on a local calendar day |
| Heliacal | `PlanetHeliacalEvents`, `StarHeliacalEvents`, `LimitingMagnitude` | Twilight visibility dates and sky limiting magnitude |
| Planetary surfaces | `SurfacePositionAt`, `SurfaceLocalTimeAt`, `SurfaceEvents`, `NextSurfaceRise/Set/Culmination` | Add `Star` prefix for fixed stars |

## Conventions and Limitations

- **Parallax & horizon dip**: Ecliptic searches use geocentric apparent positions. `AltitudeCrossing` applies topocentric parallax only to the Moon; subtract horizon dip from `altitude` yourself if needed.
- **Heliacal visibility**: Empirical sky-brightness model (default extinction `0.25 mag/airmass`, tunable with `formula.ExtinctionSchaefer`); real visibility strongly depends on local light pollution and haze.
- **Planetary surfaces**: Airless geometric model. `SurfaceOptions.UpperLimb` (default `true`) applies only to the Sun. `LocalTrueSolarTime`/`LocalMeanSolarTime` hours are 1/24 of the local solar day, while `SolarDayLength` is in Earth days; `LocalSiderealTime` and `SiderealRotationAngle` are in **degrees**. Linear IAU rotation models are most accurate within ~±500 years (`Local.ModelValid`). Polar-day/night searches exceeding the scan budget return `ErrSurfaceEventNotFound`.
