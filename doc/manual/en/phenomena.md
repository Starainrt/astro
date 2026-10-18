# Generic sky events and heliacal visibility

[中文](../phenomena.md) | [Back to README](../../../README.en.md)

The `phenomena` package searches events between any two bodies by geocentric apparent ecliptic longitude and latitude, computes morning and evening first and last visibility for planets and stars, and additionally provides rise and set at any altitude plus perihelion, aphelion and node passages. Package-level entry points use the process's current ephemeris model, and `ctx.Phenomena()` can fix an ephemeris instance; only a star's rise and set at any altitude is computed from the apparent coordinates passed in and does not read the ephemeris.

## Contents

- [Quick example](#quick-example)
- [Pairwise events](#pairwise-events)
- [Star coordinates](#star-coordinates)
- [Heliacal visibility](#heliacal-visibility)
- [Rise and set at any altitude](#rise-and-set-at-any-altitude)
- [Apsides and nodes](#apsides-and-nodes)
- [API reference](#api-reference)
- [Parameter and return conventions](#parameter-and-return-conventions)

## Quick example

```go
start := time.Date(2020, 10, 1, 0, 0, 0, 0, time.UTC)
end := time.Date(2021, 3, 1, 0, 0, 0, 0, time.UTC)

events := phenomena.ClosestApproaches(phenomena.Jupiter, phenomena.Saturn, 1, start, end)
fmt.Println(events[0].Time.UTC().Format("2006-01-02 15:04"), fmt.Sprintf("%.4f", events[0].Value))
```

```text
2020-12-21 18:21 0.1018
```

`ClosestApproaches` returns the **local minima** of the angular separation of the two bodies within the window, where `Value` is the separation at that instant in degrees. It returns an empty slice when the window contains no local minimum below `maxSeparation`.

To fix the ephemeris instance, use the module method instead:

```go
ctx, err := astro.NewContext(astro.ModelLegacy)
if err != nil {
	panic(err)
}
events = ctx.Phenomena().ClosestApproaches(phenomena.Jupiter, phenomena.Saturn, 1, start, end)
```

## Pairwise events

| Entry | Event | `Event.Value` |
| --- | --- | --- |
| `LongitudeCrossings(body, lon, start, end)` | apparent longitude passes through `lon` | apparent longitude at the event instant (degrees) |
| `Stations(body, start, end)` | stations of apparent longitude (planets) | apparent longitude at the station (degrees) |
| `AspectCrossings(a, b, angle, start, end)` | `lon(a) - lon(b) = angle`; `0` is conjunction in longitude, `180` opposition | the actual longitude difference (degrees) |
| `ClosestApproaches(a, b, max, start, end)` | local minima of the angular separation of two bodies | angular separation (degrees) |
| `StarApproaches(body, star, max, start, end)` | local minima of the angular separation from a star | angular separation (degrees) |

Each entry accepts only the parameters listed in the table above, and results are returned in ascending time order. Stations defined by right ascension in the per-planet packages use a different convention from the longitude stations here, so the times can differ by several hours.

```go
// first two Mercury stations of 2026
stations := phenomena.Stations(phenomena.Mercury,
	time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC))
for _, e := range stations {
	fmt.Println(e.Time.UTC().Format("01-02 15:04"), fmt.Sprintf("%.3f", e.Value))
}
```

```text
02-26 06:48 352.565
03-20 19:32 338.491
```

```go
// apparent solar longitude passing through 0° (vernal equinox)
crossings := phenomena.LongitudeCrossings(phenomena.Sun, 0,
	time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC), time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC))
fmt.Println(crossings[0].Time.UTC().Format("2006-01-02 15:04"))
```

```text
2026-03-20 14:45
```

## Star coordinates

Both the angular separation of a planet and a star and star heliacal visibility require a `StarCoordinate`; the embedded 9100-star catalog can be queried directly:

| Entry | Input |
| --- | --- |
| `StarCoordinateByName(name)` | catalog Chinese name/alias or English common name, for example `"天狼星"` or `"sirius"`; duplicate names return `ErrStarNameAmbiguous` |
| `StarCoordinateByHR(hr)` | HR number |
| `StarCoordinateFromStarData(star)` | a catalog record such as one obtained from `star.StarDataByHR` |

```go
star, err := phenomena.StarCoordinateByName("天狼星")
fmt.Println(err, star.ID, fmt.Sprintf("%.6f %.6f %.2f %.2f", star.RA, star.Dec,
	star.ProperMotionRACosDecMasPerYear, star.ProperMotionDecMasPerYear))
```

```text
<nil> 天狼 101.287083 -16.716111 -553.00 -1205.00
```

`StarCoordinateByName` returns `phenomena.ErrStarNameAmbiguous` for duplicate names and `phenomena.ErrStarNotFound` when nothing matches; both are usable with `errors.Is`. For duplicates, resolve the catalog record with `star.StarDataByNameIn(name, region)` first, then pass it to `StarCoordinateFromStarData`:

```go
data, err := star.StarDataByNameIn("杵增一", "危宿")
if err != nil {
	panic(err)
}
coord, err := phenomena.StarCoordinateFromStarData(data)
if err != nil {
	panic(err)
}
```

`RA` and `Dec` are in degrees, and `Epoch` and `Frame` are required; the proper-motion fields (`ProperMotionRACosDecMasPerYear` and friends) follow the catalog convention `dRA·cos(Dec)`, in milliarcseconds per year. `ParallaxMas` (or the equivalent `DistanceLightYear`) gives the distance, and `RadialVelocityKmPerSecond` gives the radial velocity; when the distance is missing the propagation is two-dimensional, and the radial velocity is included only when a distance is present.

## Heliacal visibility

`PlanetHeliacalEvents(body, start, end, lon, lat, options)` and `StarHeliacalEvents(star, magnitude, start, end, lon, lat, options)` scan the morning and evening intervals day by day while the solar altitude runs from -18° to 0°, derive the naked-eye limiting magnitude from a sky-brightness model, and compare it with the target apparent magnitude (Mallama & Hilton 2018 for planets). They return four kinds of event:

| Constant | Meaning |
| --- | --- |
| `HeliacalMorningFirst` | morning first visibility (heliacal rising) |
| `HeliacalMorningLast` | morning last visibility |
| `HeliacalEveningFirst` | evening first visibility |
| `HeliacalEveningLast` | evening last visibility (heliacal setting) |

In `HeliacalEvent`, `Time` is the best visible instant of that twilight, `TargetAltitude` and `SunAltitude` are the geometric altitudes of the target and the Sun at that instant (degrees), and `Margin` is the limiting magnitude minus the target magnitude.

| `HeliacalOptions` | Default | Description |
| --- | --- | --- |
| `ExtinctionCoefficient` | `0.25` | atmospheric extinction, magnitudes per airmass |
| `MinTargetAltitude` | `0.5` | minimum geometric altitude of the target (degrees) |
| `Moonlight` | `false` | include scattered moonlight (Krisciunas-Schaefer) |

```go
events := phenomena.PlanetHeliacalEvents(phenomena.Venus,
	time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
	31.23, 30.04, phenomena.HeliacalOptions{})
for _, e := range events {
	fmt.Println(e.Kind, e.Time.UTC().Format("2006-01-02"), fmt.Sprintf("margin=%.2f", e.Margin))
}
```

```text
3 2025-03-18 margin=0.10
0 2025-03-25 margin=0.58
1 2025-12-01 margin=0.10
```

`Kind` prints as the constant's ordinal, so compare against the constants themselves when testing. `LimitingMagnitude(targetAltitude, sunAltitude, separation, extinction)` gives the limiting magnitude at a single instant on a moonless night, from the same source as the event test:

```go
fmt.Println(fmt.Sprintf("%.3f", phenomena.LimitingMagnitude(20, -12, 60, 0.25)))
```

```text
3.408
```

Results are sensitive to the extinction coefficient, and a spread of one or two days in the visible date is normal.

## Rise and set at any altitude

`AltitudeCrossing(body, date, lon, lat, height, altitude, options)` finds the rise and set times at which the Sun, Moon or a planet first crosses the apparent altitude `altitude` within the **local civil day** containing `date`. `StarAltitudeCrossing(date, ra, dec, lon, lat, altitude, options)` performs the same computation for a star's apparent coordinates; it does not read the ephemeris and has no module method.

| `AltitudeOptions` | Default | Description |
| --- | --- | --- |
| `Limb` | `LimbUpper` | upper limb for the Sun and Moon, or `LimbCenter` and `LimbLower`; ignored for planets and stars |
| `PressureHPa` / `TemperatureC` | standard atmosphere (the `0` value) | specify measured weather |
| `NoRefraction` | `false` | when `true`, compare geometric altitudes |

| `AltitudeCrossingResult` | Description |
| --- | --- |
| `Rise`, `Set` | rise and set times, zero when there is no event |
| `HasRise`, `HasSet` | whether a rise or set occurs that day |
| `State` | `AltitudeCrosses`, `AltitudeAlwaysAbove`, `AltitudeAlwaysBelow` or `AltitudeTangent` (the extremum differs from the target altitude by less than `1″`) |
| `MinResidual`, `MaxResidual` | minimum and maximum of the altitude minus the target altitude that day (degrees) |

```go
cst := time.FixedZone("CST", 8*3600)
r, err := phenomena.AltitudeCrossing(phenomena.Sun, time.Date(2026, 3, 21, 12, 0, 0, 0, cst),
	116.4, 39.9, 0, -18, phenomena.AltitudeOptions{})
fmt.Println(err, r.State, r.Rise.In(cst).Format("15:04:05"), r.Set.In(cst).Format("15:04:05"),
	fmt.Sprintf("%.3f %.3f", r.MinResidual, r.MaxResidual))
```

```text
<nil> 0 04:44:54 19:59:16 -31.806 68.602
```

```go
// 89°N polar day: the Sun stays above -18° all day
polar, err := phenomena.AltitudeCrossing(phenomena.Sun, time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC),
	0, 89, 0, -18, phenomena.AltitudeOptions{})
fmt.Println(err, polar.State, polar.HasRise, polar.HasSet, fmt.Sprintf("%.3f", polar.MinResidual))
```

```text
<nil> 1 false false 40.737
```

```go
// rise and set of Sirius through -0.833°
sirius, err := phenomena.StarAltitudeCrossing(time.Date(2026, 1, 1, 12, 0, 0, 0, cst),
	star.RA, star.Dec, 116.4, 39.9, -0.833, phenomena.AltitudeOptions{Limb: phenomena.LimbCenter})
fmt.Println(err, sirius.State, sirius.Rise.In(cst).Format("15:04:05"), sirius.Set.In(cst).Format("15:04:05"))
```

```text
<nil> 0 19:04:31 05:27:52
```

`height` only affects lunar parallax. When a depressed horizon is required, the caller subtracts the horizon dip from `altitude`. When the body is not the Sun, Moon or a planet (for example `phenomena.Earth`), or a star's apparent coordinates are not finite, the function returns `basic.ErrInvalidObservationInput`.

## Apsides and nodes

| Constant | Event |
| --- | --- |
| `Perihelion`, `Aphelion` | perihelion, aphelion |
| `AscendingNode`, `DescendingNode` | ascending and descending node passages through the mean ecliptic of epoch |

`Apsides(body, start, end)` accepts a planet or the Earth (`phenomena.Earth`), while `Nodes(body, start, end)` accepts only planets and returns an empty slice for the Earth. In `OrbitalEvent`, `Time` is the event instant, `Kind` is a constant from the table above, and `DistanceAU` is the heliocentric distance at that instant (AU).

```go
apsides := phenomena.Apsides(phenomena.Earth,
	time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC))
for _, e := range apsides {
	fmt.Println(e.Kind, e.Time.UTC().Format("2006-01-02 15:04"), fmt.Sprintf("%.9f AU", e.DistanceAU))
}
nodes := phenomena.Nodes(phenomena.Jupiter,
	time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC))
for _, e := range nodes {
	fmt.Println(e.Kind, e.Time.UTC().Format("2006-01-02 15:04"), fmt.Sprintf("%.6f AU", e.DistanceAU))
}
```

```text
0 2026-01-03 09:15 0.983302049 AU
1 2026-07-06 09:30 1.016643962 AU
3 2020-02-25 18:09 5.207056 AU
2 2025-09-19 11:58 5.174312 AU
```

The extremum is taken on the heliocentric distance itself: the Sun's reflex motion about the barycenter makes the heliocentric distance of Uranus and Neptune show several small extrema near perihelion, so when several extrema appear in one span only the one with the smallest heliocentric distance within `0.3` orbital periods before and after is returned, which can differ from the osculating perihelion by several years.

## API reference

The methods of `ctx.Phenomena()` match the package-level functions of the same name (except `StarAltitudeCrossing`); `NewWithInstance` creates a module bound to a given ephemeris instance. Parameters and units are in the sections above.

## Parameter and return conventions

- Angles are always in degrees, `OrbitalEvent.DistanceAU` is in AU, and `magnitude` in `StarHeliacalEvents` is V magnitude.
- Time parameters are interpreted as absolute instants (their UTC fields are read), and returned instants are converted to the time zone of `start` (`date` for the rise and set entries).
- Event searches return an empty slice rather than an error when there is no result; only rise and set at any altitude returns `basic.ErrInvalidObservationInput` for invalid input.
- Event searches use the ephemeris, which a module method can fix; `StarAltitudeCrossing`, `StarCoordinate*` and `LimitingMagnitude` are independent of the ephemeris.
- For differences in convention such as planetary apparent magnitude and right-ascension stations see the [planet manual](planets.md); for Sun and Moon rise and set and observing-angle semantics see [Sun and Moon](sun-moon.md).
