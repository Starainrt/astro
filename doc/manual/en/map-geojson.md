# Event Maps, GeoJSON and KML

[中文](../map-geojson.md) | [Back to README](../../../README.en.md)

> Full examples in this manual run from the repository root and write their figures to `doc/img/`.

## Contents

- [Exporting solar-eclipse GeoJSON](#exporting-solar-eclipse-geojson)
- [API Reference](#api-reference)
- [Usage examples](#usage-examples)
  - [Choosing a projection](#choosing-a-projection)
  - [Exporting GeoJSON with time markers](#exporting-geojson-with-time-markers)
  - [Declaring the scale and UT1](#declaring-the-scale-and-ut1)
  - [Single instants and the layer vocabulary](#single-instants-and-the-layer-vocabulary)
- [Map Projections](#map-projections)
- [Time Scale Declaration](#time-scale-declaration)
- [GeoJSON](#geojson)
  - [Filtering layers](#filtering-layers)
  - [Coordinates, times and the antimeridian](#coordinates-times-and-the-antimeridian)
  - [Solar shadows at a specified instant](#solar-shadows-at-a-specified-instant)
  - [Horizon closure and interpolation](#horizon-closure-and-interpolation)
  - [ΔT and ground position](#δt-and-ground-position)
  - [Instant occultation footprints](#instant-occultation-footprints)
  - [Finding events before calculating geometry](#finding-events-before-calculating-geometry)
- [KML](#kml)
  - [Converting a file](#converting-a-file)
  - [Options](#options)
  - [Layers and styles](#layers-and-styles)
  - [Time playback and static overlays](#time-playback-and-static-overlays)
  - [File size and camera view](#file-size-and-camera-view)
  - [Properties and input validation](#properties-and-input-validation)

## Exporting solar-eclipse GeoJSON

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2009, 7, 22, 0, 0, 0, 0, cst)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date,
		eclipse.SolarEclipsePartialFootprintOptions{Step: 10 * time.Minute, BoundaryPoints: 180})
	if !ok {
		log.Fatal("no solar eclipse")
	}
	path, ok := eclipse.SolarEclipseCentralPath(date,
		eclipse.SolarEclipsePathOptions{Step: time.Minute, TargetSpacingKM: 20})
	var centralPath *eclipse.SolarEclipsePath
	if ok {
		centralPath = &path
	}
	data, err := geojson.MarshalSolarEclipseWithTimeMarkers(partial, centralPath,
		geojson.TimeMarkerOptions{Step: 30 * time.Minute, Location: cst})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))
}
```

A partial eclipse may have no central path, so `centralPath` may be `nil`. This program writes GeoJSON to standard output for redirection to a file. SVG and KML examples follow.

## API Reference

SVG snippets use the import aliases `eclipsesvg "github.com/starainrt/astro/eclipse/svg"` and, for occultations, `moonsvg "github.com/starainrt/astro/moon/svg"`. Dates and time zones reuse the first example.

| Name | Purpose | Notes |
| --- | --- | --- |
| `eclipsesvg.EclipseMapProjectionAuto` / `...Equirectangular` / `...NorthPolar` / `...SouthPolar` / `...Orthographic` | Solar/lunar map projections | The zero value is automatic |
| `moonsvg.MapProjectionAuto` / `...Equirectangular` / `...NorthPolar` / `...SouthPolar` / `...Orthographic` | Occultation map projections | Same convention |
| `SolarEclipseMapSVG` / `LunarEclipseMapSVG` | Solar and lunar global maps | Return `(string, bool)` |
| `StarOccultationPathSVG` / `PlanetOccultationPathSVG` | Occultation global path maps | Return `(string, error)` |
| `MarshalSolarEclipse` / `MarshalSolarEclipseWithTimeMarkers` | Solar eclipse GeoJSON | Without and with time markers |
| `MarshalLunarEclipse` / `MarshalLunarEclipseWithTimeMarkers` / `MarshalLunarEclipseWithOptions` | Lunar eclipse GeoJSON | Same three |
| `MarshalStarOccultation` / `MarshalPlanetOccultation` (and `...WithTimeMarkers`) | Occultation GeoJSON | Same pair |
| `NewSolarEclipseShadowSolver` / `MarshalSolarEclipseShadowInstant` | Single-instant footprint and its GeoJSON | For time-axis scrubbing |
| `TimeMarkerOptions` | Marker `Step`, `Location` and `TimeScale` | `Step` defaults to 30 minutes, at most 1440 markers |
| `astro.TimeScaleUT1` / `astro.DUT1` | UT1 scale and the DUT1 offset | UT1 requires a UTC `Location` |
| `kml.FromGeoJSON` | Convert GeoJSON into KML 2.2 | Standard library only; groups layers by `role` and supplies a default palette |

## Usage examples

### Choosing a projection

```go
for _, spec := range []struct {
	name string
	p    eclipsesvg.EclipseMapProjection
}{
	{"equirectangular", eclipsesvg.EclipseMapProjectionEquirectangular},
	{"north-polar", eclipsesvg.EclipseMapProjectionNorthPolar},
	{"south-polar", eclipsesvg.EclipseMapProjectionSouthPolar},
	{"orthographic", eclipsesvg.EclipseMapProjectionOrthographic},
} {
	svg, ok := eclipsesvg.SolarEclipseMapSVG(date, eclipsesvg.SolarEclipseMapSVGOptions{
		Width: 1200, Height: 800, Location: cst, Projection: spec.p,
	})
	fmt.Println(spec.name, ok, len(svg))
}
```

```text
equirectangular true 265827
north-polar true 181602
south-polar true 91902
orthographic true 537001
```

The projection only affects SVG presentation and never the underlying WGS84 geography; `Auto` (the zero value) picks a polar map per event, so specify one explicitly only when the layout must be fixed.

The orthographic globe is viewed from the greatest-eclipse point, draws only the facing hemisphere and switches to the NASA layout - its cost and the base-map inverse projection are described under [Map Projections](#map-projections).

### Exporting GeoJSON with time markers

```go
data, err := geojson.MarshalSolarEclipseWithTimeMarkers(partial, centralPath,
	geojson.TimeMarkerOptions{Step: 30 * time.Minute, Location: cst})
fmt.Println(err, json.Valid(data), len(data))
```

```text
<nil> true 426253
```

- `WithTimeMarkers` appends Point features with `role=time-marker`: `label` is localised through `Location`, while `time` stays UTC RFC 3339.
- `Step` defaults to 30 minutes, positive values are at least one minute, and one export carries at most 1440 markers; use the unsuffixed `MarshalSolarEclipse` when no markers are wanted.
- Lunar eclipses and occultations have symmetric entry points (`MarshalLunarEclipse*`, `MarshalStarOccultation*`, `MarshalPlanetOccultation*`); GeoJSON carries no basemap, boundary, style or projection, leaving Web Mercator and tile choices to the application.

### Declaring the scale and UT1

```go
fmt.Println(astro.DUT1(date)) // UT1−UTC in seconds
```

```text
0.23
```

Charts are drawn in the UTC scale by default and state it inside the figure or its footer; `TimeScale: astro.TimeScaleUT1` switches to UT1 readings and prints `DUT1 = UT1−UTC = +0.23 s` in the figure, with `Location` required to be UTC.

On the GeoJSON side the counterpart is the `time_scale` member (omitted for UTC, `"UT1"` for UT1); the full convention is under [Time Scale Declaration](#time-scale-declaration).

### Single instants and the layer vocabulary

```go
solver := eclipse.NewSolarEclipseShadowSolver(eclipse.SolarEclipseShadowSolverOptions{})
instant, ok := solver.ShadowAt(date)
shadow, err := geojson.MarshalSolarEclipseShadowInstant(instant)
fmt.Println(ok, err, json.Valid(shadow), len(shadow))
```

```text
true <nil> true 4217
```

- The single-instant API computes only "this instant's umbral/penumbral footprint" and the station geometry; it produces no visibility bands, magnitude contours, rise/set boundaries or limits, and returns an empty value (not an error) when the umbra misses the Earth.
- Compare `interp_signature` before interpolating: only identical neighbouring instants are safe to interpolate vertex by vertex; a flipped `closed`, a changed segment or vertex count, or an empty/non-empty transition (near U1/U4) all require exact geometry instead.
- Degradable layers record their actual geometry source in `data-source`; the vocabulary is under [Map Projections](#map-projections).

## Map Projections

The solar, lunar and occultation SVGs share one set of projections: equirectangular, north-polar azimuthal equidistant, south-polar azimuthal equidistant, and orthographic globe.

The projection only affects SVG presentation and never the underlying WGS84 geography.

The automatic projection (the zero value `...Auto`) picks a polar map for the solar and occultation maps when that fits, while lunar maps default to equirectangular.

| Projection | Constant (solar/lunar - occultation) | When to use | Suggested canvas |
| --- | --- | --- | --- |
| Automatic | `EclipseMapProjectionAuto` - `MapProjectionAuto` | Default; picks a polar map per event | 1200x800 |
| Equirectangular | `...Equirectangular` | Long bands crossing the antimeridian | 1200x800, 1414x1000 |
| North/south polar equidistant | `...NorthPolar` / `...SouthPolar` | Event lies entirely at high latitude | 1200x800, 1000x1414 |
| Orthographic globe | `...Orthographic` | NASA-style hemisphere view centred on the event | 1000x1414 |

`EclipseMapProjectionOrthographic` / `MapProjectionOrthographic` produce the orthographic globe: the view point is the greatest-eclipse point (the event centre for an occultation), only the hemisphere facing it is drawn, and the projection boundary is the great circle of the visible hemisphere.

The orthographic projection uses the NASA-style centered-globe layout; a `1000x1414` canvas is recommended. It does not change the underlying geographic results.

Four projections of one event in a single pass:

```go
date := time.Date(2009, 7, 22, 12, 0, 0, 0, cst)
for _, spec := range []struct {
	name string
	p    eclipsesvg.EclipseMapProjection
}{
	{"equirectangular", eclipsesvg.EclipseMapProjectionEquirectangular},
	{"north-polar", eclipsesvg.EclipseMapProjectionNorthPolar},
	{"south-polar", eclipsesvg.EclipseMapProjectionSouthPolar},
	{"orthographic", eclipsesvg.EclipseMapProjectionOrthographic},
} {
	options := eclipsesvg.SolarEclipseMapSVGOptions{
		Width: 1200, Height: 800, Location: cst, Projection: spec.p,
	}
	svg, ok := eclipsesvg.SolarEclipseMapSVG(date, options)
	if !ok {
		continue
	}
	_ = os.WriteFile("solar-eclipse-"+spec.name+".svg", []byte(svg), 0o644)
}
```

Per-family options, layer switches and illustrated examples live in the [solar and lunar eclipse manual](eclipse.md#solar-and-lunar-eclipse-charts) and the [lunar occultation manual](occultation.md#lunar-occultation-charts); the GeoJSON side carries geographic results only and leaves the projection to the application.

Degradable layers record their actual geometry source in `data-source`; the vocabulary is listed in the `eclipse/svg` package comment: `partial-band-union`, `sampled-footprint-sweep`, `partial-band-contours`, `rise-set-phase-lines`, `magnitude-contours`, `greatest-time-isochrones`, `besselian-critical-envelope`, `paired-limit-chords`, `sampled-open-sweep`, `central-path-limits`, `penumbral-outlines`, `central-shadow-outlines`, `p1-p4-visibility-regions`, `p1-p4-horizon-boundaries`.

## Time Scale Declaration

The `TimeScale` option fixes the scale of every instant printed in the figures.

Solar eclipse maps, lunar eclipse maps, the lunar-eclipse shadow-path layout, and all three lunar-occultation layouts state it inside the figure or in the footer:

- Default (`astro.TimeScaleUTC`, the zero value): `All times are UTC`, or `All times are UTC (shown in CST, UTC+08:00)` when the display zone is not UTC, so the scale and the display zone are declared separately.
- `astro.TimeScaleUT1`: `Times are UT1 (Universal Time 1); DUT1 = UT1−UTC = +0.05 s.` The offset varies with the event date. `Location` must be UTC in this mode, otherwise rendering returns `false` (the occultation side returns an error).

The GeoJSON counterpart is the `time_scale` member: omitted for UTC, and `"UT1"` when every `time` string and `HH:MM` label is a UT1 reading (an RFC 3339 `Z` suffix is not strictly UT1).

Switching the scale never changes the band, footprint, horizon or contour geometry; only the emitted time text becomes a UT1 reading.

Time markers are placed at whole clock ticks of the output scale (a UT1 tick is a different physical instant), so their positions follow the instant they label.

## GeoJSON

`geojson` accepts an already computed solar-eclipse, lunar-eclipse, or lunar-occultation result and returns `[]byte`.

Those bytes are one complete UTF-8 RFC 7946 `FeatureCollection`, not an image or compressed payload: they can be written to `.geojson`, passed to `encoding/json`, or served directly to a map client.

```go
package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

func main() {
	date := time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(
		date,
		eclipse.SolarEclipsePartialFootprintOptions{
			Step: 10 * time.Minute, BoundaryPoints: 180,
		},
	)
	if !ok {
		return
	}
	central, hasCentral := eclipse.SolarEclipseCentralPath(
		date,
		eclipse.SolarEclipsePathOptions{Step: time.Minute, TargetSpacingKM: 20},
	)
	var centralPath *eclipse.SolarEclipsePath
	if hasCentral {
		centralPath = &central
	}

	data, err := geojson.MarshalSolarEclipseWithTimeMarkers(
		partial, centralPath,
		geojson.TimeMarkerOptions{
			Step:     30 * time.Minute,
			Location: time.FixedZone("CST", 8*3600),
		},
	)
	fmt.Println(err, json.Valid(data))
}
```

Each event type has plain and time-marker variants:

- `MarshalSolarEclipse` / `MarshalSolarEclipseWithTimeMarkers`
- `MarshalLunarEclipse` / `MarshalLunarEclipseWithTimeMarkers`
- `MarshalLunarEclipseWithOptions`: one `LunarEclipseOptions` value carrying the time markers, `SkipRoles` and the time-envelope sampling. `SkipRoles` can omit envelopes and penumbra-only bands; skipping both envelopes also skips their sampling. `EnvelopeSweepSamples` defaults to 48 and is clamped to `[2, 192]`; `EnvelopeLongitudePoints` defaults to `max(360, boundaryPoints)` and is clamped to `[12, 720]`.
- `MarshalStarOccultation` / `MarshalStarOccultationWithTimeMarkers`
- `MarshalPlanetOccultation` / `MarshalPlanetOccultationWithTimeMarkers`
- `MarshalSolarEclipseWithOptions`: one `SolarEclipseOptions` value carrying both the time markers and `SkipRoles`, equivalent to the solar entry points above plus layer trimming.

  The usual entry in `SkipRoles` is `partial-footprint` (the instantaneous penumbral outlines); `partial-band` is the edge of the eclipse band (the magnitude-0 limit) and is normally kept.

### Filtering layers

`SolarEclipseOptions.SkipRoles` excludes features by `role`; its zero value retains every layer. `MarshalSolarEclipseWithOptions` accepts both filtering and time-marker options.

Filtering occurs during encoding, after full geometry sampling. Omitting `partial-footprint` removes individual penumbral footprints without reducing the accuracy of `partial-band`, the central band, magnitude contours or rise/set edges.

Removing every feature returns an error.

### Coordinates, times and the antimeridian

Coordinates are WGS84 `[longitude, latitude]` in degrees. Lines and polygons crossing the antimeridian are split, with matching intersection points on both sides. Polar rings represent a pole using the two coordinates `[±180, ±90]`.

Closed rings include ±180° seam segments needed for filling. Omit those segments when drawing physical outlines. This also applies to closed line features such as `band-outline` and `total-band-outline`.

On timed paths, `times` matches the coordinates point by point within each segment. `WithTimeMarkers` adds Point features with `role=time-marker`. `label` is for display; `time` defaults to UTC RFC 3339.

Explicit UT1 output is identified by `time_scale`; see [Time scales](timescale.md).

| `TimeMarkerOptions` field | Behavior |
| --- | --- |
| `Step` | Zero means 30 minutes; positive values must be at least one minute; at most 1440 markers per export |
| `Location` | Display zone for labels; `nil` means UTC |
| `TimeScale` | UTC by default; UT1 requires `Location` to be `nil` or `time.UTC` |

GeoJSON contains no basemap, styling or projection. Applications can choose tiles, Web Mercator or a polar projection independently.

### Solar shadows at a specified instant

For an interactive time query, reuse the solver returned by `eclipse.NewSolarEclipseShadowSolver`:

| Method | Input and result |
| --- | --- |
| `ShadowAt(date)` | Civil `time.Time`; instantaneous global shadow footprint |
| `ShadowAtJDE(jdeTT)` | TT Julian day; the same footprint type |
| `StationStateAt` / `StationStateAtJDE` | Site-specific magnitude, obscuration, separation, semidiameters, solar altitude/azimuth and total/annular state |
| `ShadowBetween(start, end, step)` | Footprint sequence; instants without a shadow retain empty entries |
| `StationStatesBetween` | Time-sampled site states |

The default is umbra or antumbra. `Kind: SolarEclipseShadowPenumbra` selects the penumbra. Instant APIs do not calculate event-wide visibility bands, magnitude contours, limits or center lines.

A shadow missing Earth produces an empty result.

`geojson.MarshalSolarEclipseShadowInstant(instant)` encodes the result. Properties include `time`, `source_boundary_closed`, `geometry_role`, `closure`, `delta_t_seconds`, `model` and `interp_signature`.

No shadow yields an empty FeatureCollection.

| Shadow | Region role | Physical-boundary role |
| --- | --- | --- |
| Umbra/antumbra | `central-shadow-footprint` | `central-shadow-boundary` |
| Penumbra | `partial-footprint` | `partial-footprint-boundary` |

Instant penumbral output and event-wide partial sampling use the same default boundary settings: 96 points and 200 km refinement. Compare results at the same instant.

Prior single-machine measurements put a 96-point footprint at about 64 µs and an instantaneous site state at about 20 µs. These are cost estimates; first queries and batches depend on cache state.

### Horizon closure and interpolation

`central-shadow-footprint` is always a Polygon or MultiPolygon. If the horizon cuts it, the physical boundary extends to two horizon grazing points and closes along the horizon arc.

`source_boundary_closed=false` and `closure` describes the arc using `kind`, `time` and `subsolar`.

Draw `central-shadow-boundary` for the physical edge and fill the footprint for the region. At U1/U4, an empty footprint is omitted rather than degraded to a line. `source_boundary_closed=true` means the physical boundary closes by itself.

Sampled penumbral footprints carry the same properties. A footprint describes one instant; a static band describes the event-wide envelope. They are not interchangeable.

Compare adjacent `interp_signature` values, such as `umbra-closed-seg1-pt97`, before interpolating vertices. Signatures, segment counts and vertex counts must match. This condition is not a strict interpolation error bound.

Query exact geometry if closure changes, antimeridian splitting changes, or a frame switches between empty and non-empty. In prior two-minute samples, footprint centroids moved 78–232 km mid-event and up to about 520 km near contacts.

A fixed time step alone does not determine animation error.

### ΔT and ground position

The solver's `DeltaTSeconds > 0` sets ΔT for that handle; non-positive values use the process-wide model. Results report the value actually used.

At a fixed TT, changing ΔT changes Earth's rotation phase without changing the relative Sun–Moon geometry. Approximate east–west displacement is `0.4651 × |ΔΔT| × cos(latitude)` km; use `basic.DeltaTGroundShiftKM` for this conversion.

The library supplies no ΔT uncertainty model. Callers must provide the uncertainty being converted to a ground displacement.

### Instant occultation footprints

Pass results from `moon.StarOccultationFootprintAt` / `moon.PlanetOccultationFootprintsAt` to `geojson.MarshalStarOccultationFootprint` / `MarshalPlanetOccultationFootprints`.

Properties include `delta_t_seconds`, `source_boundary_closed`, `geometry_role` and `interp_signature`. Lunar-horizon closure uses `closure.kind=target-horizon`, `body=moon` and the `sublunar` point.

Occultations use process-wide ΔT and report its actual value.

### Finding events before calculating geometry

`eclipse.SolarEclipseCandidates(start, end, options)` returns greatest times, eclipse types, central types, magnitudes, gamma and optional Saros information. It does not calculate geographic geometry.

For central eclipses at a fixed site, `SearchLocalCentralSolarEclipse` accepts `Kind`, `MaxYears`, `Backward`, `Geometric` and `Model` options and returns `(info, status)`. `status.Exhausted` reports that the search range was exhausted.

`MaxYears<=0` uses the default budget of 6000 candidate steps, about 992 years. Calculate paths, footprints or SVG only after selecting an event when full maps are not needed for every candidate.

## KML

`kml.FromGeoJSON(data []byte, options kml.Options) ([]byte, error)` converts a GeoJSON FeatureCollection to KML 2.2. Input may come from this library or a third-party file. It converts existing geometry and time properties; it does not calculate ephemerides or generate animation frames.

### Converting a file

This program converts `eclipse.geojson` to `eclipse.kml` for Google Earth:

```go
package main

import (
	"log"
	"os"

	"github.com/starainrt/astro/kml"
)

func main() {
	data, err := os.ReadFile("eclipse.geojson")
	if err != nil {
		log.Fatal(err)
	}
	result, err := kml.FromGeoJSON(data, kml.Options{
		Name: "2009-07-22 Solar eclipse",
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("eclipse.kml", result, 0644); err != nil {
		log.Fatal(err)
	}
}
```

### Options

| Field | Zero value or default | Effect when set |
| --- | --- | --- |
| `Name` | Derived from event type and earliest time | Sets `Document/name` |
| `Language` | `"zh"` | `"en"` selects English layer names; unknown roles keep their identifiers |
| `Styles` | Built-in palette | Override by `role` or the more specific `event/role`, which takes precedence |
| `NoLookAt` | Automatic view | `true` omits `Document/LookAt` |
| `SkipRoles` | Keep every role | Remove entire roles, including geometry, styles and their contribution to the view |
| `FillContext` | Fill only highlighted central and occultation bands | `true` adds light-grey translucent fill to other polygons |
| `NoTimes` | Write available input times | `true` omits time primitives for a static overlay |

### Layers and styles

Features are grouped into Folders by `event` and `role`. Collections containing multiple event types add event prefixes to layer names. Third-party events and roles support the same style overrides.

| Layer | role | Default style |
| --- | --- | --- |
| Solar central band and umbra | `central-band`, `central-shadow`, `central-shadow-footprint`, `central-shadow-sweep`, `total-footprint` | Red; polygon fill about 35% opaque |
| Center line | `center-line` | Black, 3 px |
| Solar greatest-time isochrones | `greatest-time-line` | Green |
| Magnitude contours | `magnitude-line`, `magnitude-one-envelope` | Yellow |
| Total and partial occultation bands | `total-band`, `partial-band`, `total-band-outline`, `occultation-band` | Yellow with translucent band fill |
| Rise/set visibility edges, lunar time envelopes and solar partial-band edge | `visibility-boundary`, `p1-horizon`, `p4-horizon`, `visible-at-p1`, `visible-at-p4`, `visible-during-eclipse`, `visible-throughout-eclipse`, solar `partial-band` | Orange |
| Penumbra-only moonrise/moonset bands | `penumbra-moonrise`, `penumbra-moonset` | Blue-violet for moonrise, violet for moonset, with a translucent fill |
| Other limits, outlines, footprints and time markers | Other roles | Grey; polygons are outline-only by default |

Zero-valued `Style` fields retain palette defaults:

| Field | Meaning |
| --- | --- |
| `LineColor` | Line color in KML `aabbggrr` hexadecimal |
| `FillColor` | Polygon fill; ignored by points and lines |
| `LineWidth` | Pixels; values at or below zero use the default |
| `NoFill` | Forces no fill, overriding `FillColor` and the palette |

KML color order is alpha, blue, green, red. `ff0000ff` is opaque red; `590000ff` is about 35% opaque red. This differs from CSS `rrggbb`.

```go
options := kml.Options{
    Language: "en",
    Styles: map[string]kml.Style{
        "center-line": {LineColor: kml.ColorBlack, LineWidth: 4},
        "solar-eclipse/central-band": {FillColor: "590000ff"},
        "partial-footprint": {NoFill: true},
    },
}
result, err := kml.FromGeoJSON(data, options)
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(result))
```

Unfilled polygons become outline lines. Filled polygons and their outlines are represented separately; seams introduced by antimeridian splitting are omitted from the outlines.

Polygons remain closed without drawing an artificial boundary along ±180°.

### Time playback and static overlays

| GeoJSON time data | KML output |
| --- | --- |
| Scalar `time` property | One `TimeStamp` on the Placemark containing all component geometries |
| MultiLineString `times` | Separate Placemarks per segment, each timestamped with its first time |
| At least one valid timestamp | Document `TimeSpan` spanning the earliest to latest timestamp |
| `time_scale: "UT1"` | Converted back to UTC under the active model before writing `<when>` |
| `NoTimes: true` | No `TimeStamp` or `TimeSpan` |

Timestamps preserve fractional seconds. A `label` supplies the feature name when present; generated names display whole seconds. The original UT1 `time_scale` remains in properties.

Use the same ΔT model for GeoJSON generation and KML conversion.

Animation frames must already exist in the input. `times` does not turn every line vertex into a separate frame. To animate umbral or penumbral footprints, generate timestamped features through the GeoJSON sampling options first.

Google Earth hides timed features outside its selected time window. Use `NoTimes: true` to view a static full path. For playback, select the event date and adjust the visible time-window width.

### File size and camera view

Penumbral footprints can cover large areas. Dense sampling, many vertices and overlapping translucent polygons all increase rendering cost. Increase the sampling step or remove layers that the application does not need.

To keep the overall partial-visibility edge and central path while dropping individual penumbral footprints:

```go
result, err := kml.FromGeoJSON(data, kml.Options{
    SkipRoles: []string{"partial-footprint"},
    NoTimes: true,
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(result))
```

When generating GeoJSON, `SkipRoles` in `geojson.MarshalSolarEclipseWithOptions` can exclude layers earlier and reduce intermediate data. KML's `SkipRoles` also works on existing files.

Automatic framing prefers center lines, central bands, limits and umbral paths; it falls back to all features if those are absent. Bounds handle antimeridian crossings. `LookAt/range` is in metres, with a 200 km minimum.

Set `NoLookAt: true` to let the client choose the view.

### Properties and input validation

After `SkipRoles` filtering, only properties present with identical values on every retained feature move to `Document/ExtendedData`. Other properties remain on their Placemarks, including nested objects.

`times` is consumed as timestamps and is not repeated as a property.

Coordinates need at least longitude and latitude. Longitudes outside ±180° wrap, while existing +180° and −180° values are preserved. Latitude must lie within ±90°. Additional altitude components are ignored; output uses `clampToGround`.

Polygon rings are closed and oriented with counterclockwise outer rings and clockwise holes. Supported types are Point, MultiPoint, LineString, MultiLineString, Polygon, MultiPolygon and GeometryCollection.

Errors include invalid JSON or geometry types, non-finite coordinates, invalid latitudes, incompatible time-array/segment structure, filtering away every feature, or having no renderable geometry.

Two identical positions forming a single-point segment are skipped. An originally empty FeatureCollection produces a valid empty Document.
