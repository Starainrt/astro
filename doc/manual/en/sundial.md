# Sundials & Solar Time

[中文](../sundial.md) | [Back to README](../../../README.en.md)

The `sundial` package computes local apparent (true) solar time, local mean solar time, the equation of time, solar hour angles, and planar sundial geometry—including shadow points, usable illumination intervals, time lines, and declination curves for horizontal, vertical, equatorial, or arbitrarily tilted dials.

## Contents

- [Basic Example](#basic-example)
- [Usage](#usage)
  - [Usable Illumination Intervals & Declination Curves](#usable-illumination-intervals--declination-curves)
  - [Mean Solar & Zone Time Lines (Analemma / Clock Markings)](#mean-solar--zone-time-lines-analemma--clock-markings)
  - [Horizontal, Equatorial, Vertical & Arbitrary Dials](#horizontal-equatorial-vertical--arbitrary-dials)
- [API Reference](#api-reference)
- [Conventions and Limitations](#conventions-and-limitations)

## Basic Example

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
	lon, lat := 121.4737, 31.2304 // Shanghai (deg E/N)

	ts := sundial.TrueSolarTime(date, lon)
	ms := sundial.MeanSolarTime(date, lon)
	ha := sundial.HourAngle(date, lon)

	dial := sundial.HorizontalDial(lat, 10) // 10 cm stylus -> shadow X/Y in cm
	sh := dial.ShadowPointAt(date, lon)

	fmt.Printf("Shanghai 09:30 CST: true solar %s  mean solar %s  EoT %s\n",
		ts.Format("15:04:05"), ms.Format("15:04:05"), ts.Sub(ms).Round(time.Second))
	fmt.Printf("Hour angle %.2f°  shadow (10cm stylus): x=%.2f cm (E), y=%.2f cm (N)  lit=%v\n",
		ha, sh.X, sh.Y, sh.Illuminated)
}
```

Output:

```text
Shanghai 09:30 CST: true solar 09:34:10  mean solar 09:35:53  EoT -1m43s
Hour angle -36.46°  shadow (10cm stylus): x=-6.51 cm (E), y=0.51 cm (N)  lit=true
```

- Horizontal dials use **+x East, +y North** in the same length unit as `StylusLength`.
- Always check `sh.Illuminated` (both Sun above horizon and dial face lit) before drawing a shadow.

## Usage

### Usable Illumination Intervals & Declination Curves

Find how long a dial is illuminated on a given day and sample the hyperbolic shadow tip path with `DeclinationCurveAt(date, stepDeg)` (or `DeclinationCurve(declDeg, stepDeg)`):

```go
dial := sundial.HorizontalDial(31.2304, 10)
for _, seg := range dial.DeclinationCurveAt(date, 1.0) {
	fmt.Printf("Solstice usable hour angle: %.1f° .. %.1f° (decl %.2f°, %d samples)\n",
		seg.Interval.Start, seg.Interval.End, seg.Declination, len(seg.Samples))
}
```

Output:

```text
Solstice usable hour angle: -105.2° .. 105.2° (decl 23.44°, 211 samples)
```

- Use `IlluminatedHourAngleIntervals(decl)` for intervals only (or `PlaneIlluminatedHourAngleIntervals` for face-only geometry). Always iterate over the returned slice—intervals wrapping midnight split into multiple segments.

### Mean Solar & Zone Time Lines (Analemma / Clock Markings)

- **Standard clock (zone) time**: `dial.ZoneTimePoint(date, lon, 9.5)` / `dial.ZoneTimeLine(dates, lon, 9.5)` uses the date & time zone of `date` with clock reading `9.5` (09:30).
- **Local mean solar time**: `dial.MeanSolarTimePoint(meanDate, 9.5)` / `dial.MeanSolarTimeLine(meanDates, 9.5)` expects local mean solar time (from `sundial.MeanSolarTime`).
- **Horizontal hour-line angles**: `sundial.HorizontalHourLineAngle(lat, hourAngleDeg)` or `HorizontalHourLineAngleAt(date, lon, lat)`.

### Horizontal, Equatorial, Vertical & Arbitrary Dials

| Constructor | Dial Type | `PlaneNormalAzimuth` | `PlaneNormalZenithDistance` |
| --- | --- | --- | --- |
| `HorizontalDial(lat, len)` | Horizontal | `180°` | `0°` (zenith) |
| `EquatorialNorthDial(lat, len)` | Equatorial north face | `0°` | `90° − lat` |
| `EquatorialSouthDial(lat, len)` | Equatorial south face | `180°` | `90° + lat` |
| `VerticalDial(lat, normalAz, len)` | Vertical wall | Wall **normal** azimuth (`180°` south-facing, `90°` east-facing) | `90°` |

For arbitrary orientations, fill `PlanarDial` directly and call `dial.Geometry()` for the dial center (`CenterX`, `CenterY`), polar stylus length, and stylus-to-plane angle.

## API Reference

| Group | Entry Points | Notes |
| --- | --- | --- |
| Solar time | `TrueSolarTime`, `MeanSolarTime` | Local solar-time readings (synthetic `LTZ` zone) |
| Hour angles | `HourAngle`, `MeanSolarHourAngle`, `ZoneTimeHourAngle` | Degrees, negative morning / positive afternoon |
| Horizontal lines | `HorizontalHourLineAngle`, `HorizontalHourLineAngleAt` | Hour-line angle relative to meridian (deg) |
| Constructors | `HorizontalDial`, `VerticalDial`, `EquatorialNorthDial`, `EquatorialSouthDial`, `PlanarDial` | Planar dial definitions |
| Geometry & shadow | `Geometry`, `ShadowPointAt`, `ShadowPointByHourAngleDeclination` | Dial center & shadow coordinates |
| Illumination & curves | `IlluminatedHourAngleIntervals`, `PlaneIlluminatedHourAngleIntervals`, `DeclinationCurve`, `DeclinationCurveAt` | Usable intervals & sampled declination paths |
| Time lines | `MeanSolarTimePoint/Line`, `ZoneTimePoint/Line` | Analemma and zone-time curve samples |

## Conventions and Limitations

- **Hour-angle sign**: Negative in the morning, positive in the afternoon (`[-180°, 180°]`).
- **Do not feed solar-time readings back into ephemeris APIs**: `TrueSolarTime` / `MeanSolarTime` return synthetic local-solar-time wall readings.
- **Degenerate dials**: When the plane normal is perpendicular to Earth's polar axis (e.g. `lat=45°`, `normalAz=180°`, `normalZD=45°`), the dial center lies at infinity (`Geometry().HasFiniteCenter == false`, center fields `NaN`), though `ShadowPointAt` still computes shadows normally.
- **Pitfalls**: `VerticalDial` takes the wall **normal** azimuth (`180°` for a south-facing wall); `ZoneTimePoint` ignores `date`'s clock time and uses `zoneTimeHours`; always test `Illuminated` (not `PlaneIlluminated`) for real shadows.


Related manuals: [Sun & Moon](sun-moon.md) | [Coordinates](coord.md) | [Time Scales](timescale.md)
