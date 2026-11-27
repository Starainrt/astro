# Maps, GeoJSON & KML

[中文](../map-geojson.md) | [Back to README](../../../README.en.md)

The `geojson` and `kml` packages export solar/lunar eclipses, Mercury/Venus transits, lunar occultations, and Earth-satellite geometry into Leaflet, Mapbox, Cesium, QGIS, or Google Earth. This manual also documents the shared SVG map projections and UTC/UT1 time-scale rules.

## Contents

- [Basic Example](#basic-example)
- [Usage](#usage)
  - [Lunar Eclipse, Transit & Occultation GeoJSON](#lunar-eclipse-transit--occultation-geojson)
  - [Satellite Layers](#satellite-layers)
  - [Interactive Time Slider: Instant Shadow Footprints](#interactive-time-slider-instant-shadow-footprints)
  - [KML 2.2 Options & File Size Control](#kml-22-options--file-size-control)
- [Map Projections](#map-projections)
- [Time-Scale Declarations](#time-scale-declarations)
- [API Reference](#api-reference)
- [Conventions and Limitations](#conventions-and-limitations)

## Basic Example

```go
package main

import (
	"log"
	"os"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/kml"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2009, 7, 22, 0, 0, 0, 0, cst)

	partial, ok := eclipse.SolarEclipsePartialFootprints(date,
		eclipse.SolarEclipsePartialFootprintOptions{Step: 10 * time.Minute, BoundaryPoints: 180})
	if !ok {
		log.Fatal("no solar eclipse on date")
	}

	var centralPtr *eclipse.SolarEclipsePath
	if path, hasCentral := eclipse.SolarEclipseCentralPath(date,
		eclipse.SolarEclipsePathOptions{Step: time.Minute, TargetSpacingKM: 20}); hasCentral {
		centralPtr = &path
	}

	// 1. Marshal RFC 7946 GeoJSON (centralPtr may be nil for partial eclipses)
	gj, err := geojson.MarshalSolarEclipseWithTimeMarkers(partial, centralPtr,
		geojson.TimeMarkerOptions{Step: 30 * time.Minute, Location: cst})
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("eclipse-2009.geojson", gj, 0o644); err != nil {
		log.Fatal(err)
	}

	// 2. Convert to styled KML 2.2 with time playback (standard library only)
	kmlData, err := kml.FromGeoJSON(gj, kml.Options{Name: "2009-07-22 Total Solar Eclipse", Language: "en"})
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("eclipse-2009.kml", kmlData, 0o644); err != nil {
		log.Fatal(err)
	}
}
```

> To shrink output size (e.g. omit per-step `partial-footprint` polygons while keeping `partial-band` and the central path), call `geojson.MarshalSolarEclipseWithOptions` with `SkipRoles: []string{"partial-footprint"}`.

## Usage

### Lunar Eclipse, Transit & Occultation GeoJSON

All `Marshal*` functions return UTF-8 RFC 7946 `FeatureCollection` bytes (`[lon, lat]` in WGS84 degrees):

- **Lunar eclipses**: `MarshalLunarEclipse`, `MarshalLunarEclipseWithTimeMarkers`, `MarshalLunarEclipseWithOptions`.
- **Mercury / Venus transits**: `MarshalMercuryTransit`, `MarshalVenusTransit` (with optional C1–C4 topocentric isochrones):
  ```go
  m := mercury.NextTransit(time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC))
  data, err := geojson.MarshalMercuryTransit(m, geojson.PlanetTransitOptions{
    BoundaryPoints: 360,
    IsochroneStep:  5 * time.Minute,
    TimeMarkers:    &geojson.TimeMarkerOptions{Step: time.Hour, Location: time.UTC},
  })
  ```
- **Lunar occultations**: `MarshalStarOccultation`, `MarshalPlanetOccultation` (and `WithTimeMarkers` variants).

### Satellite Layers

Compute geometry with [`satellite`](satellite.md), marshal with `geojson`, and optionally convert with `kml.FromGeoJSON`:

| GeoJSON Function | Input | Main `role` Values |
| --- | --- | --- |
| `MarshalSatelliteTransitPath` | `satellite.TransitEventPath` | `transit-band`, `center-line`, `left-limit`, `right-limit`, `target-horizon` |
| `MarshalSatelliteTransitShadow` | `satellite.TransitShadowAt` | `current-transit-footprint`, `current-transit-point`, `subsatellite-point` |
| `MarshalSatelliteGroundTrack` | `satellite.GroundTrack` | `ground-track` (split by `sunlit` / `penumbra` / `umbra`) |
| `MarshalSatelliteCoverage` | `satellite.CoverageAt` | `subsatellite-point`, `satellite-footprint`, `night-region`, `satellite-visible` |
| `MarshalSatelliteCoverageSweep` | `satellite.CoverageBetween` | `coverage-sweep`, `visible-sweep`, `ground-track`, `current-satellite-footprint` |

- Antimeridian crossings and enclosed holes are encoded as standard GeoJSON polygons with interior rings (`innerBoundaryIs` in KML).
- Default KML styling colors geometric coverage blue, visual coverage teal, night gray, transit bands yellow, and centerlines red.

### Interactive Time Slider: Instant Shadow Footprints

For scrubbing a live map without computing a whole eclipse:

```go
solver := eclipse.NewSolarEclipseShadowSolver(eclipse.SolarEclipseShadowSolverOptions{})
instant, ok := solver.ShadowAt(time.Date(2009, 7, 22, 1, 40, 0, 0, time.UTC))
if ok {
	data, err := geojson.MarshalSolarEclipseShadowInstant(instant)
	_ = data
	_ = err
}
```

- `solver` also offers `StationStateAt` (local magnitude, obscuration, Sun altitude) and `ShadowBetween`.
- Lunar occultations have matching `moon.StarOccultationFootprintAt` / `PlanetOccultationFootprintsAt` + `geojson.Marshal*Footprint`.
- Before interpolating vertices between two frames, compare `interp_signature`; query an exact frame whenever signatures differ.

### KML 2.2 Options & File Size Control

`kml.FromGeoJSON(data, options)` works on both library-generated and third-party GeoJSON. `kml.Options` fields include `Name`, `Language` (`"zh"` / `"en"`), `Styles` (override palette by `role` or `"event/role"`), `SkipRoles` (drop heavy layers), `FillContext` (fill neutral context polygons), `NoTimes` (omit time primitives for a static overlay), and `NoLookAt` (omit default `Document/LookAt` camera).

## Map Projections

Eclipse, occultation, and satellite SVG maps share four projections (SVG presentation only; underlying WGS84 geometry is unchanged):

| Projection | Eclipse · Occultation/Satellite Constants | Best For | Suggested Size |
| --- | --- | --- | --- |
| Auto (zero value) | `EclipseMapProjectionAuto` · `MapProjectionAuto` | Default; switches to polar maps for high-latitude events | 1200×800 |
| Equirectangular | `...Equirectangular` | Long paths crossing the antimeridian | 1200×800 or 1414×1000 |
| North / South Polar | `...NorthPolar` / `...SouthPolar` | High-latitude / polar events | 1200×800 |
| Orthographic | `...Orthographic` | NASA-style globe centered on the event | 1000×1414 |

## Time-Scale Declarations

SVG and GeoJSON default to **UTC**, with an explicit **UT1** option (`TimeScale: astro.TimeScaleUT1`):

- **UTC (default)**: SVGs state UTC (plus display time zone if non-UTC); GeoJSON omits `time_scale` and writes UTC RFC 3339 `time` strings.
- **UT1**: Requires `Location` to be `nil` or `time.UTC`. SVGs print `DUT1 = UT1−UTC`; GeoJSON adds `"time_scale": "UT1"`.
- **KML**: Because KML `<when>` must be UTC, `kml.FromGeoJSON` converts UT1 timestamps back to UTC using the active ΔT model while preserving the original `time_scale` attribute.

## API Reference

| Package | Entry Points | Purpose |
| --- | --- | --- |
| `geojson` | `MarshalSolarEclipse[WithTimeMarkers/WithOptions]`, `MarshalSolarEclipseShadowInstant` | Global & instant solar eclipse GeoJSON |
| `geojson` | `MarshalLunarEclipse[WithTimeMarkers/WithOptions]` | Lunar visibility & time-envelope GeoJSON |
| `geojson` | `MarshalMercuryTransit`, `MarshalVenusTransit` | Planetary transit GeoJSON |
| `geojson` | `MarshalStarOccultation*`, `MarshalPlanetOccultation*` | Occultation bands & instant footprints |
| `geojson` | `MarshalSatelliteTransitPath/Shadow`, `GroundTrack`, `Coverage`, `CoverageSweep` | Satellite GeoJSON layers |
| `kml` | `FromGeoJSON(data, options)` | GeoJSON → KML 2.2 |

## Conventions and Limitations

- **Antimeridian seams**: Polygons split at ±180° include closing seam segments for area fill; when stroking physical boundaries only, use the corresponding `*-boundary` line role (e.g. `central-shadow-boundary`) or skip seam edges.
- **Time markers**: `TimeMarkerOptions.Step` defaults to 30 min (minimum 1 min), capped at 1440 markers per export.
- **ΔT ground shift**: Changing ΔT shifts terrestrial longitudes by approximately `0.4651 × |ΔΔT| × cos(lat)` km (`basic.DeltaTGroundShiftKM`); set `SolarEclipseShadowSolverOptions.DeltaTSeconds > 0` to override ΔT per solver.
- **SVG `data-source`**: Degradable SVG layers carry a `data-source` attribute recording the geometry builder used (e.g. `partial-band-union`, `magnitude-contours`, `greatest-time-isochrones`).
