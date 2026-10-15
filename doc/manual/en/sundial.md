# Sundials and Apparent Solar Time

[中文](../sundial.md) | [Back to README](../../../README.en.md)

> Full examples in this manual run from the repository root.

`sundial` gathers the apparent solar time, solar hour angle and dial geometry of the `sun` package in one place and does not introduce a second algorithm.

The dial side follows the classical planar-dial model: a polar-axis stylus whose shadow falls on an arbitrary plane, with the coordinate convention given by the constructor - a horizontal dial has **x pointing east and y pointing north**.

## Contents

- [Calculating a shadow on a horizontal sundial](#calculating-a-shadow-on-a-horizontal-sundial)
- [API Reference](#api-reference)
  - [True, mean solar time and the equation of time](#true-mean-solar-time-and-the-equation-of-time)
  - [Hour angle](#hour-angle)
  - [Horizontal hour-line angle](#horizontal-hour-line-angle)
  - [Planar dial core](#planar-dial-core)
  - [Plate illumination intervals](#plate-illumination-intervals)
  - [Time lines and declination curves](#time-lines-and-declination-curves)
  - [Equatorial, horizontal and vertical dials](#equatorial-horizontal-and-vertical-dials)
  - [Returned structures](#returned-structures)
  - [Complete example](#complete-example)
  - [Combined example: usable hour angles and a time line](#combined-example-usable-hour-angles-and-a-time-line)
  - [Common pitfalls](#common-pitfalls)
- [Usage examples](#usage-examples)
  - [Apparent solar time versus clock time](#apparent-solar-time-versus-clock-time)
  - [Horizontal hour-line angles and the shadow point](#horizontal-hour-line-angles-and-the-shadow-point)
  - [Plate illumination intervals and time lines](#plate-illumination-intervals-and-time-lines)
  - [Degenerate geometry and per-face dial conventions](#degenerate-geometry-and-per-face-dial-conventions)
- [Parameter and result conventions](#parameter-and-result-conventions)
- [Related manuals](#related-manuals)

## Calculating a shadow on a horizontal sundial

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/sundial"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2026, 6, 21, 9, 30, 0, 0, cst)
	lon, lat := 121.4737, 31.2304
	fmt.Println(sundial.TrueSolarTime(date, lon))
	fmt.Println(sundial.HourAngle(date, lon))
	dial := sundial.HorizontalDial(lat, 10)
	shadow := dial.ShadowPointAt(date, lon)
	if !shadow.Illuminated {
		fmt.Println("no illuminated shadow")
		return
	}
	fmt.Printf("x=%.6f y=%.6f\n", shadow.X, shadow.Y)
}
```

The dial length and returned coordinates use the same unit: a stylus length in centimetres gives shadow coordinates in centimetres. x points east and y north. Use the shadow only when `Illuminated` is true.

## API Reference

| Group | Entry points | Purpose | Units and convention |
| --- | --- | --- | --- |
| True/mean solar time | `TrueSolarTime` / `MeanSolarTime` | Local apparent/mean solar time at a longitude for an absolute instant | Local solar clock readings; longitude east positive (degrees) |
| Solar hour angle | `HourAngle` | Apparent solar hour angle | Degrees, negative in the morning and positive in the afternoon |
| Hour-angle helpers | `MeanSolarHourAngle` / `ZoneTimeHourAngle` | Turn a local mean-solar or zone clock reading into an apparent solar hour angle | Hours and longitude in degrees |
| Horizontal hour line | `HorizontalHourLineAngle` / `HorizontalHourLineAngleAt` | Hour-line angle of a horizontal dial relative to the noon line | Degrees |
| Planar dial core | `PlanarDial` (fields) + `Geometry` / `ShadowPointByHourAngleDeclination` / `ShadowPointAt` | Geometry and shadow point of an arbitrary plane | Coordinates share the stylus-length unit |
| Plate illumination | `PlaneIlluminatedHourAngleIntervals` / `IlluminatedHourAngleIntervals` | Plate-lit hour-angle intervals and the final usable intervals | Degrees, `[-180, 180]` |
| Time lines | `MeanSolarTimePoint` / `ZoneTimePoint` / `MeanSolarTimeLine` / `ZoneTimeLine` | Attach mean-solar or zone time lines directly to the dial geometry | Returns `PlanarShadowPoint` / `TimeLineSample` |
| Declination curves | `DeclinationCurve` / `DeclinationCurveAt` | Segmented sample chains by declination or by date | Hour-angle step in degrees |
| Special dials | `EquatorialNorthDial` / `EquatorialSouthDial` / `HorizontalDial` / `VerticalDial` | Equatorial (north/south face), horizontal and vertical dials | Latitude and normal azimuth in degrees |

The snippets below omit shared preamble variables: `date` (instant, civil time scale), `lon` (longitude, east positive, degrees), `lat` (latitude, degrees).

### True, mean solar time and the equation of time

```go
fmt.Println(sundial.TrueSolarTime(date, lon))
fmt.Println(sundial.MeanSolarTime(date, lon))
fmt.Println(sundial.TrueSolarTime(date, lon).Sub(sundial.MeanSolarTime(date, lon))) // equation of time
```

All three share the `sun` conventions: `TrueSolarTime` is apparent solar time, `MeanSolarTime` is local mean solar time, and their difference is the equation of time (apparent minus mean).

### Hour angle

```go
fmt.Println(sundial.HourAngle(date, lon))          // apparent solar hour angle, negative in the morning
fmt.Println(sundial.MeanSolarHourAngle(date, 9.5)) // hour angle for local mean solar time 9:30
fmt.Println(sundial.ZoneTimeHourAngle(date, lon, 9.5)) // hour angle for zone clock time 9:30
```

`HourAngle` solves the hour angle for an absolute instant; the other two answer "given a clock reading, where does the shadow point", one in local mean solar time and one in zone time.

### Horizontal hour-line angle

```go
fmt.Println(sundial.HorizontalHourLineAngle(31.2304, -45))
fmt.Println(sundial.HorizontalHourLineAngleAt(date, lon, 31.2304))
```

The first takes latitude and a signed hour angle, the second takes the instant and coordinates directly; both return the angle of the hour line relative to the noon line.

### Planar dial core

```go
dial := sundial.PlanarDial{
	Latitude: 31.2304, PlaneNormalAzimuth: 180, PlaneNormalZenithDistance: 90, StylusLength: 10,
}
g := dial.Geometry()
fmt.Println(g.HasFiniteCenter, g.PolarStylusLength, g.PolarStylusPlaneAngle)
p := dial.ShadowPointByHourAngleDeclination(-45, 23.44)
fmt.Println(p.X, p.Y, p.Illuminated)
```

The four `PlanarDial` fields are latitude, plate-normal azimuth, normal zenith distance and stylus length. `Geometry` returns the dial center (where the polar stylus is fixed), the polar-stylus length and its angle to the plate; `ShadowPointByHourAngleDeclination` takes a signed hour angle and the solar declination, while `ShadowPointAt` takes the instant and longitude instead.

### Plate illumination intervals

```go
for _, iv := range dial.PlaneIlluminatedHourAngleIntervals(23.44) {
	fmt.Println(iv.Start, iv.End)
}
for _, iv := range dial.IlluminatedHourAngleIntervals(23.44) {
	fmt.Println(iv.Start, iv.End)
}
```

The first only answers "does the plate face the Sun" (geometric illumination); the second also requires the Sun to be above the horizon and therefore gives the **final usable** hour-angle intervals.

Intervals live in `[-180, 180]` with `Start <= End`.

### Time lines and declination curves

```go
dates := []time.Time{date, date.Add(30 * time.Minute), date.Add(time.Hour)}
fmt.Println(len(dial.MeanSolarTimeLine(dates, 9.5)))
segs := dial.DeclinationCurve(23.44, 1.0)
segsAt := dial.DeclinationCurveAt(date, 1.0)
fmt.Println(len(segs), len(segsAt))
```

A time line projects the equal-instant points of a mean-solar-time line straight onto the dial as `TimeLineSample` values; a declination curve samples the plate for a fixed declination or for the declination of the day, and each segment's `Interval` is the usable hour-angle interval for that declination.

### Equatorial, horizontal and vertical dials

```go
h := sundial.HorizontalDial(31.2304, 10)
n := sundial.EquatorialNorthDial(31.2304, 10)
s := sundial.EquatorialSouthDial(31.2304, 10)
v := sundial.VerticalDial(31.2304, 180, 10)
fmt.Println(h.PlaneNormalZenithDistance, n.PlaneNormalAzimuth, s.PlaneNormalAzimuth)
fmt.Println(v.PlaneNormalAzimuth, v.PlaneNormalZenithDistance)
```

Each constructor sets the plate normal differently; when drawing, read it from the `PlanarDial` fields:

| Constructor | `PlaneNormalAzimuth` | `PlaneNormalZenithDistance` |
| --- | --- | --- |
| `HorizontalDial` | `180°` | `0°` (normal points at the zenith; x east, y north) |
| `EquatorialNorthDial` | `0°` | `90° - latitude` |
| `EquatorialSouthDial` | `180°` | `90° + latitude` |
| `VerticalDial` | argument normalized to `[0°, 360°)` | `90°` |

In the northern hemisphere the north-face equatorial dial serves the spring/summer half-year (positive solar declination) and the south face the autumn/winter half-year.

For `VerticalDial` the normal azimuth runs from north toward east, so a south-facing wall is `180` and an east-facing wall is `90`.

### Returned structures

| Type | Field | Meaning |
| --- | --- | --- |
| `PlanarShadowPoint` | `X` / `Y` | Shadow-point coordinates in the stylus-length unit |
| | `DenominatorQ` | Projection denominator; approaching zero means the shadow runs to infinity |
| | `SunAboveHorizon` / `PlaneIlluminated` / `Illuminated` | Sun above the horizon, plate lit, and the final combined test |
| `PlanarGeometry` | `CenterX` / `CenterY` | Dial center (where the polar stylus is fixed) |
| | `PolarStylusLength` / `PolarStylusPlaneAngle` | Polar-stylus length and its angle to the plate |
| | `HasFiniteCenter` | False when the center degenerates to infinity; the related quantities are `NaN` |
| `HourAngleInterval` | `Start` / `End` | Signed hour-angle interval in degrees, `Start <= End` |
| `TimeLineSample` | `Date` / `Declination` / `HourAngle` / `Point` | Instant, solar declination, apparent hour angle and the shadow point |
| `DeclinationCurveSegment` | `Declination` / `Interval` / `Samples` | Segment declination, usable hour-angle interval and sample chain |

### Complete example

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/sundial"
)

func main() {
	date := time.Date(2026, 6, 21, 9, 30, 0, 0, time.FixedZone("CST", 8*3600))
	lon, lat := 121.4737, 31.2304

	trueSolar := sundial.TrueSolarTime(date, lon)
	hourAngle := sundial.HourAngle(date, lon)
	lineAngle := sundial.HorizontalHourLineAngle(lat, -45)
	lineAngleNow := sundial.HorizontalHourLineAngleAt(date, lon, lat)

	fmt.Println(trueSolar)
	fmt.Printf("hour angle=%.6f line@9am=%.6f line@now=%.6f\n", hourAngle, lineAngle, lineAngleNow)
}
```

Output:

```text
2026-06-21 09:34:10.404156224 +0805 LTZ
hour angle=-36.456649 line@9am=-27.405871 line@now=-20.959281
```

The zone in the first line is a synthetic local apparent solar time zone (`+08:05`, `LTZ`, for longitude `121.4737`), so the printed value already shows how far local apparent solar time is from clock time.

### Combined example: usable hour angles and a time line

```go
dial := sundial.HorizontalDial(31.2304, 10)
for _, seg := range dial.DeclinationCurveAt(date, 1.0) {
	fmt.Printf("decl=%.2f usable=%.2f..%.2f samples=%d\n", seg.Declination, seg.Interval.Start, seg.Interval.End, len(seg.Samples))
}
mean := sundial.MeanSolarTime(date, 121.4737)
samples := dial.MeanSolarTimeLine([]time.Time{mean, mean.Add(30 * time.Minute)}, 9.5)
fmt.Println(len(samples), samples[0].HourAngle)
```

Measured with `date = 2026-06-21 09:30 CST`, `121.4737 E, 31.2304 N` and a stylus of length 10:

```text
decl=23.44 usable=-105.24..105.24 samples=211
2 -37.92998436772365
```

The first line says that with the solar declination at `23.44` degrees, a horizontal dial at latitude `31.2304` is usable over hour angles `[-105.24, +105.24]` (about 14 hours) sampled at 211 points; the second gives the two samples of the "local mean solar time 9:30" time line and the first hour angle. The usable interval can be used directly as the drawing clip range for the time lines.

### Common pitfalls

- Treating the `date` of `ZoneTimePoint` as the site's local apparent solar time - it uses the date and zone only, and takes the clock reading from `zoneTimeHours`.
- Using `PlaneIlluminated` to decide whether a dial is usable during the day - use `Illuminated`, which also requires the Sun above the horizon.
- Passing the wall orientation to `VerticalDial` - `planeNormalAzimuth` is the **normal** azimuth, so a south-facing wall is `180`.
- Flipping the hour-angle sign - `HourAngle` is negative in the morning and positive in the afternoon, and `HorizontalHourLineAngle` follows the same convention.
- Assuming one interval per day - a day crossing midnight splits into several, so iterate over the returned slice.

## Usage examples

### Apparent solar time versus clock time

```go
trueSolar := sundial.TrueSolarTime(date, lon)
mean := sundial.MeanSolarTime(date, lon)
fmt.Println(trueSolar, mean)
fmt.Println(trueSolar.Sub(mean)) // equation of time
```

```text
2026-06-21 09:34:10.438158222 +0805 LTZ 2026-06-21 09:35:53.532790863 +0805 LTZ
-1m43.094632641s
```

- `TrueSolarTime` returns an instant in a synthetic zone, so subtracting your own clock time gives "how far apparent solar time runs ahead or behind".
- To turn a **clock reading** into an hour angle (for example to lay out shadow marks in zone time), use `MeanSolarHourAngle` / `ZoneTimeHourAngle` instead of adding the equation of time by hand.

### Horizontal hour-line angles and the shadow point

```go
dial := sundial.HorizontalDial(lat, 10)
fmt.Println(sundial.HorizontalHourLineAngle(lat, -45))        // hour-line angle at hour angle -45
fmt.Println(sundial.HorizontalHourLineAngleAt(date, lon, lat)) // hour-line angle right now
p := dial.ShadowPointAt(date, lon)
fmt.Println(p.X, p.Y, p.Illuminated)
```

```text
-27.40587112370779
-20.95918157094186
-6.511723246549 0.507600045956 true
```

- A horizontal dial uses `x` pointing east and `y` pointing north; `Illuminated` is the final test "Sun above the horizon and plate facing the Sun", so draw a shadow only when it is true.
- With an hour angle but no instant, use `ShadowPointByHourAngleDeclination`, which gives `-6.511402572556 0.506762879485 true` for the same geometry.

### Plate illumination intervals and time lines

```go
dial := sundial.HorizontalDial(31.2304, 10)
for _, seg := range dial.DeclinationCurveAt(date, 1.0) {
	fmt.Printf("decl=%.2f usable=%.2f..%.2f samples=%d\n", seg.Declination, seg.Interval.Start, seg.Interval.End, len(seg.Samples))
}
mean := sundial.MeanSolarTime(date, 121.4737)
samples := dial.MeanSolarTimeLine([]time.Time{mean, mean.Add(30 * time.Minute)}, 9.5)
fmt.Println(len(samples), samples[0].HourAngle)
```

```text
decl=23.44 usable=-105.24..105.24 samples=211
2 -37.92998436772365
```

- `PlaneIlluminatedHourAngleIntervals` only asks whether the plate faces the Sun, while `IlluminatedHourAngleIntervals` also requires the Sun above the horizon; use the latter to answer "how long is this dial usable in a day".
- Intervals live in `[-180, 180]` with `Start <= End`, and a day crossing midnight splits into several; the full declination-curve and time-line APIs are under [Time lines and declination curves](#time-lines-and-declination-curves).

### Degenerate geometry and per-face dial conventions

```go
h := sundial.HorizontalDial(31.2304, 10)
n := sundial.EquatorialNorthDial(31.2304, 10)
s := sundial.EquatorialSouthDial(31.2304, 10)
v := sundial.VerticalDial(31.2304, 180, 10)
fmt.Println(h.PlaneNormalZenithDistance, n.PlaneNormalAzimuth, s.PlaneNormalAzimuth)
fmt.Println(v.PlaneNormalAzimuth, v.PlaneNormalZenithDistance)
```

```text
0 0 180
180 90
```

- **Degenerate case**: when the plate normal is perpendicular to the polar axis (equivalently the polar stylus is parallel to the plate), `Geometry().HasFiniteCenter` is `false`, `CenterX`/`CenterY`/`PolarStylusLength` are `NaN` and `PolarStylusPlaneAngle` is 0 - latitude `45`, normal azimuth `180` and normal zenith distance `45` is exactly such a point, and nothing drawn relative to the center is usable.
- **Per-face conventions**: a horizontal plate's normal points at the zenith (zenith distance `0`); a vertical plate has zenith distance `90` and its `planeNormalAzimuth` is the **normal** direction (south-facing wall `180`, east-facing wall `90`); the equatorial north and south faces use `90 - latitude` and `90 + latitude`.

  The full table of the four constructors is under [Equatorial, horizontal and vertical dials](#equatorial-horizontal-and-vertical-dials).

## Parameter and result conventions

- **Units**: hour angles, hour-line angles, declination, latitude, normal azimuth and normal zenith distance are all **degrees**.

  `PlanarDial.StylusLength` and the returned `X`/`Y` share one length unit, which may be anything self-consistent (millimetres, metres or canvas coordinates).
- **Time scale**: observing inputs are civil instants. `TrueSolarTime` / `MeanSolarTime` return local solar clock readings; do not pass those readings back as new observing instants. See [Time scales](timescale.md).
- **Hour-angle sign**: `HourAngle` is negative in the morning and positive in the afternoon; `HourAngleInterval` uses `[-180, 180]` and guarantees `Start <= End`, so a day crossing midnight splits into several intervals.
- **The time zone of `date` (easy to get wrong)**: for `MeanSolarTimePoint` / `MeanSolarTimeLine` the `date` is the **local mean solar time of the target site** (usually the value returned by `MeanSolarTime(...)`).

  `ZoneTimePoint` / `ZoneTimeLine` ignore the hour, minute and second of `date`, keep only its date and zone, and substitute the `zoneTimeHours` argument for the clock reading. Passing the wrong convention shifts the whole time line.
- **Three booleans gate a valid shadow**: `SunAboveHorizon` says the Sun is up, `PlaneIlluminated` says the plate faces the Sun, and `Illuminated` is the final test requiring both; draw a shadow only when `Illuminated` is true.
- **Degenerate case**: when the plate normal is perpendicular to the polar axis, `HasFiniteCenter` is `false`, `CenterX`/`CenterY`/`PolarStylusLength` are `NaN`, the center is at infinity and nothing drawn relative to the center is usable; see [Degenerate geometry and per-face dial conventions](#degenerate-geometry-and-per-face-dial-conventions).
- **`PlaneIlluminated` versus `Illuminated`**: the former is geometric only, the latter also requires the Sun above the horizon; use the latter when answering "how long is this dial usable in a day".

## Related manuals

- Apparent solar time, equation of time and solar position: [Sun and Moon](sun-moon.md)
- Hour angle, sidereal time and horizontal transforms: [Coordinate tools](coord.md)
- Time scale in figures: [Time Scale Declaration](map-geojson.md#time-scale-declaration)
