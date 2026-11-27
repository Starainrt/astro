# Satellites

[中文](../satellite.md) | [Back to README](../../../README.en.md) | [Satellite SVG](satellite-svg.md) | [GeoJSON / KML](map-geojson.md#satellite-layers)

The `satellite` package parses TLE and OMM orbital elements, runs SGP4/SDP4 propagation, and computes observer look angles, passes, solar/lunar transits, and ground coverage footprints.

## Contents

- [Basic Example](#basic-example)
- [Usage](#usage)
  - [Parse TLE & OMM Elements](#parse-tle--omm-elements)
  - [Sub-Satellite Point & Topocentric Look Angles](#sub-satellite-point--topocentric-look-angles)
  - [Solar / Lunar Transit Bands & Local Transits](#solar--lunar-transit-bands--local-transits)
  - [Instantaneous Footprint & Coverage Sweep](#instantaneous-footprint--coverage-sweep)
- [API Reference](#api-reference)
- [Conventions and Limitations](#conventions-and-limitations)

## Basic Example

A complete runnable example that parses an ISS TLE and finds the next pass above 10° elevation from 40°N, 0°E, including naked-eye visibility and brightest magnitude:

```go
package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/starainrt/astro/satellite"
)

func main() {
	tle := []byte(`ISS (ZARYA)
1 25544U 98067A   26317.91612726  .00008668  00000+0  16320-3 0  9995
2 25544  51.6324 279.1461 0007381  22.6178 337.5134 15.49359724594966`)

	elements, err := satellite.ParseTLEText(tle)
	if err != nil {
		log.Fatal(err)
	}
	iss := elements[0]

	obs := satellite.Observer{Longitude: 0, Latitude: 40, Height: 0} // deg E/N, ellipsoidal height (m)
	start := time.Date(2026, 11, 14, 18, 0, 0, 0, time.UTC)
	stdMag := 0.0 // provide standard magnitude to estimate visual brightness; nil leaves Magnitude as NaN

	pass, err := satellite.NextPass(iss, obs, start, satellite.PassOptions{
		MinElevation:   10,
		ObserveOptions: satellite.ObserveOptions{StandardMagnitude: &stdMag},
	})
	if err != nil {
		log.Fatal(err)
	}

	kindName := map[satellite.PassKind]string{
		satellite.PassVisible:  "visible",
		satellite.PassDaylight: "daylight",
		satellite.PassEclipsed: "eclipsed",
	}

	fmt.Printf("Satellite: %s (NORAD %d)\n", iss.Name, iss.CatalogNumber)
	fmt.Printf("Rise: %s UTC  Az %.1f°\n", pass.Rise.Time.Format("2006-01-02 15:04:05"), pass.Rise.Azimuth)
	fmt.Printf("Max:  %s UTC  El %.1f°  Az %.1f°  Range %.0f km\n",
		pass.Culmination.Time.Format("15:04:05"), pass.Culmination.Elevation, pass.Culmination.Azimuth, pass.Culmination.RangeKM)
	fmt.Printf("Set:  %s UTC  Az %.1f°  Duration %s\n",
		pass.Set.Time.Format("15:04:05"), pass.Set.Azimuth, pass.Duration.Round(time.Second))
	fmt.Printf("Visibility: %s", kindName[pass.Kind])
	if pass.Kind == satellite.PassVisible {
		fmt.Printf(" (%s – %s UTC", pass.VisibleStart.Format("15:04:05"), pass.VisibleEnd.Format("15:04:05"))
		if !math.IsNaN(pass.BrightestMagnitude) {
			fmt.Printf(", brightest %.1f mag @ %s", pass.BrightestMagnitude, pass.BrightestTime.Format("15:04:05"))
		}
		fmt.Print(")")
	}
	fmt.Println()
}
```

Output:

```text
Satellite: ISS (ZARYA) (NORAD 25544)
Rise: 2026-11-14 18:19:55 UTC  Az 266.1°
Max:  18:22:58 UTC  El 31.6°  Az 331.7°  Range 745 km
Set:  18:26:02 UTC  Az 37.2°  Duration 6m7s
Visibility: visible (18:19:55 – 18:23:47 UTC, brightest -0.9 mag @ 18:23:43)
```

> To list all passes over a time range (e.g. 24 hours), use `Passes(iss, obs, start, end, opt)`. Set `VisibleOnly: true` to return only naked-eye visible passes.

## Usage

The snippets below reuse `iss`, `obs`, and `start` from the Quick Start.

### Parse TLE & OMM Elements

- Use `ParseTLE(line1, line2)` for a single two-line element set, or `ParseTLEText(data)` for multi-satellite or 3-line text.
- Use `ParseOMM(data)` for modern CelesTrak / Space-Track OMM feeds—it auto-detects JSON, XML, and KVN (explicit `ParseOMMJSON` / `ParseOMMXML` / `ParseOMMKVN` are also available).

```go
ommJSON := []byte(`[{
	"OBJECT_NAME": "ISS (ZARYA)",
	"OBJECT_ID": "1998-067A",
	"EPOCH": "2026-11-13T21:59:13.395264",
	"MEAN_MOTION": 15.49359724,
	"ECCENTRICITY": 0.0007381,
	"INCLINATION": 51.6324,
	"RA_OF_ASC_NODE": 279.1461,
	"ARG_OF_PERICENTER": 22.6178,
	"MEAN_ANOMALY": 337.5134,
	"EPHEMERIS_TYPE": 0,
	"CLASSIFICATION_TYPE": "U",
	"NORAD_CAT_ID": 25544,
	"ELEMENT_SET_NO": 999,
	"REV_AT_EPOCH": 59496,
	"BSTAR": 0.0001632,
	"MEAN_MOTION_DOT": 0.00008668,
	"MEAN_MOTION_DDOT": 0
}]`)

list, err := satellite.ParseOMM(ommJSON)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("OMM: %s (NORAD %d, epoch %s UTC)\n",
	list[0].Name, list[0].CatalogNumber, list[0].Epoch.Format("2006-01-02 15:04:05"))
```

Output:

```text
OMM: ISS (ZARYA) (NORAD 25544, epoch 2026-11-13 21:59:13 UTC)
```

> Low-Earth-orbit elements degrade with atmospheric drag; use fresh elements within 1–3 days of epoch for accurate predictions.

### Sub-Satellite Point & Topocentric Look Angles

- `GroundPointAt(e, t)` returns WGS84 latitude, longitude, and altitude (km); `GroundTrack(e, start, end, step)` samples a ground track.
- `LookAt(e, obs, t, opt)` returns azimuth, elevation, slant range, range rate (for Doppler), apparent RA/Dec, illumination state, and estimated magnitude.

```go
at := start.Add(23 * time.Minute)
gp, err := satellite.GroundPointAt(iss, at)
if err != nil {
	log.Fatal(err)
}
look, err := satellite.LookAt(iss, obs, at, satellite.ObserveOptions{StandardMagnitude: &stdMag})
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Sub-point: lon %.2f°  lat %.2f°  alt %.1f km\n", gp.Longitude, gp.Latitude, gp.HeightKM)
fmt.Printf("Look:      az %.1f°  el %.1f°  range %.0f km  mag %.1f\n",
	look.Azimuth, look.Elevation, look.RangeKM, look.Magnitude)
```

Output:

```text
Sub-point: lon -3.47°  lat 44.71°  alt 419.0 km
Look:      az 332.5°  el 31.5°  range 745 km  mag -0.5
```

- `LookAt` includes atmospheric refraction by default; set `NoRefraction: true` when matching geometric coverage rings.

### Solar / Lunar Transit Bands & Local Transits

1. Call `TransitEvents` to find global solar/lunar transit bands (`TransitSun`, `TransitMoon`, or empty for both).
2. Call `Transits` at a specific site for exact entry, center, exit times, and duration—or `NearbyTransits` to filter multiple satellites passing within a ground distance of your site.

```go
tStart := time.Date(2026, 11, 14, 0, 40, 0, 0, time.UTC)
events, err := satellite.TransitEvents(iss, tStart, tStart.Add(40*time.Minute), satellite.TransitSearchOptions{
	Targets: []satellite.TransitTarget{satellite.TransitMoon},
})
if err != nil {
	log.Fatal(err)
}
ev := events[0]
fmt.Printf("Lunar transit band: %s – %s UTC  max half-width %.1f km  center (%.2f°E, %.2f°N)\n",
	ev.Start.Format("15:04:05"), ev.End.Format("15:04:05"),
	ev.MaxHalfWidthKM, ev.GreatestPoint.Longitude, ev.GreatestPoint.Latitude)

site := satellite.Observer{Longitude: ev.GreatestPoint.Longitude, Latitude: ev.GreatestPoint.Latitude}
transits, err := satellite.Transits(iss, site, ev.Start.Add(-time.Minute), ev.End.Add(time.Minute), satellite.TransitOptions{
	Targets: []satellite.TransitTarget{satellite.TransitMoon},
})
if err != nil {
	log.Fatal(err)
}
tr := transits[0]
fmt.Printf("Site transit: center %s UTC  duration %.2f s  min sep %.1f\"  Moon radius %.1f\"\n",
	tr.Center.Format("15:04:05.00"), tr.Duration.Seconds(), tr.MinSeparationArcsec, tr.TargetRadiusArcsec)
```

Output:

```text
Lunar transit band: 00:43:16 – 01:17:54 UTC  max half-width 23.5 km  center (-161.77°E, -12.48°N)
Site transit: center 01:00:34.27 UTC  duration 0.53 s  min sep 0.3"  Moon radius 896.6"
```

- Pass `TransitEventPath(iss, ev, time.Second)` to [TransitMapSVG](satellite-svg.md#1-lunar-transit-band-map-transitmapsvg) or [GeoJSON / KML](map-geojson.md#satellite-layers); pass `TransitFrames(iss, site, tr, 50*time.Millisecond)` to [TransitLocalSVG](satellite-svg.md#2-local-disc-transit-chart-transitlocalsvg).

### Instantaneous Footprint & Coverage Sweep

- `CoverageAt` computes the geometric visibility ring (`Footprint`), Earth's night region (`Night`), and visual coverage (prefer `VisibleRegions`, falling back to `Visible` when empty).
- `CoverageBetween` computes the ground track plus swept geometric (`Swept`) and visual (`VisibleSwept`) unions over an interval.

```go
cov, err := satellite.CoverageAt(iss, start.Add(23*time.Minute), satellite.CoverageOptions{MinElevation: 10})
if err != nil {
	log.Fatal(err)
}
sweep, err := satellite.CoverageBetween(iss, start, start.Add(time.Hour), 30*time.Second, satellite.CoverageOptions{MinElevation: 10})
if err != nil {
	log.Fatal(err)
}
rKM, _ := satellite.SphericalFootprintRadiusKM(cov.SubPoint.HeightKM, 10)
visCount := len(cov.VisibleRegions)
if visCount == 0 {
	visCount = len(cov.Visible)
}
fmt.Printf("Footprint: alt %.1f km  10° radius ~%.0f km  visual regions=%d\n",
	cov.SubPoint.HeightKM, rKM, visCount)
fmt.Printf("1h sweep:  track pts=%d  swept regions=%d  visual regions=%d\n",
	len(sweep.Track), len(sweep.Swept), len(sweep.VisibleSwept))
```

Output:

```text
Footprint: alt 419.0 km  10° radius ~1387 km  visual regions=1
1h sweep:  track pts=121  swept regions=1  visual regions=2
```

## API Reference

| Group | Entry Points | Notes |
| --- | --- | --- |
| Parsing | `ParseTLE` / `ParseTLEText` / `ParseOMM` / `ParseOMMJSON` / `ParseOMMXML` / `ParseOMMKVN` | TLE & CCSDS OMM element loaders |
| Propagation | `Propagate` / `GroundPointAt` / `GroundTrack` | TEME state (km, km/s) and WGS84 sub-satellite points |
| Observing | `LookAt` / `NextPass` / `Passes` | Az/El, range, RA/Dec, illumination, magnitude & pass visibility |
| Transits | `Transits` / `TransitFrames` / `TransitEvents` / `ClosestTransitEvent` / `TransitEventPath` / `TransitShadowAt` / `NearbyTransits` | Site transits, global bands, center/limit paths & instant shadows |
| Coverage | `CoverageAt` / `CoverageBetween` / `SphericalFootprintRadiusKM` | Instant footprints, night region, visual overlap & interval sweeps |
| Ephemeris | `NewWithInstance` | Bind an explicit `*astro.Instance` (or use `ctx.Satellite()`) |

Sentinel errors (`errors.Is`): `ErrInvalidElements`, `ErrInvalidInput`, `ErrDecayed`, `ErrMeanMotion`, `ErrMeanEccentricity`, `ErrPerturbedEccentricity`, `ErrSemiLatusRectum`, `ErrPassNotFound`, `ErrTransitNotFound`.

## Conventions and Limitations

- **SGP4/SDP4 & time**: Near-Earth satellites use SGP4 and deep-space satellites (period ≥ 225 min) use SDP4 in TEME, rotated to WGS84. `Propagate` rejects spans exceeding 292 years (`time.Duration` range). Output timestamps keep the query's time zone.
- **Transits**: Treat the satellite as a point and the Sun/Moon as apparent discs on an airless WGS84 ellipsoid. `TransitEvent.Complete == false` indicates the band was clipped by the query window. `MaxSeparationFactor > 1` in `TransitOptions` also captures near misses (`NearMiss: true`).
- **Antimeridian & holes**: Polar or high-altitude footprints can wrap the poles or contain enclosed daylight holes. Prefer `Coverage.VisibleRegions` (seam-split and hole-aware) before falling back to `Coverage.Visible`.
- **Visual magnitude**: `Look.Magnitude` and `Pass.BrightestMagnitude` are finite only when `StandardMagnitude` is supplied and the satellite is outside Earth's umbra; otherwise they are `NaN`.
