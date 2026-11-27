# Eclipses & Transits

[中文](../eclipse.md) | [Back to README](../../../README.en.md)

The `eclipse` package searches and computes solar and lunar eclipses, local contact times, central paths, partial footprints, Besselian elements, and Saros numbers. `eclipse/svg` renders local and global eclipse charts plus Mercury/Venus transit maps, and `moon/limb` adds optional Kaguya lunar-limb terrain correction.

> For Mercury/Venus transit contact times see [Planets: Mercury & Venus Transits](planets.md#mercury--venus-transits); for GeoJSON / KML export see [Maps, GeoJSON & KML](map-geojson.md).

## Contents

- [Basic Example](#basic-example)
- [Usage](#usage)
  - [Solar Eclipses: Local Search, Multi-Year Scans & Central Paths](#solar-eclipses-local-search-multi-year-scans--central-paths)
  - [Lunar Eclipses: Danjon vs Chauvenet](#lunar-eclipses-danjon-vs-chauvenet)
  - [Lunar Limb Correction](#lunar-limb-correction)
  - [Global & Local Eclipse Charts (SVG)](#global--local-eclipse-charts-svg)
  - [Transit Charts](#transit-charts)
- [API Reference](#api-reference)
- [Conventions and Limitations](#conventions-and-limitations)

## Basic Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/eclipse"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)

	// 1. Local solar eclipse: Yangshan, Shanghai (2009-07-22)
	d2009 := time.Date(2009, 7, 22, 12, 0, 0, 0, cst)
	if s, ok := eclipse.LocalSolarEclipseOnDate(d2009, 121.9850, 30.6167, 0); ok {
		fmt.Printf("2009-07-22 Yangshan: type=%s  Saros %d (%d/%d)\n",
			s.Type, s.Saros.Series, s.Saros.Member, s.Saros.Count)
		fmt.Printf("  C1 %s  C2 %s  Max %s  C3 %s  C4 %s\n",
			s.PartialStart.Format("15:04:05"), s.CentralStart.Format("15:04:05"),
			s.GreatestEclipse.Format("15:04:05"), s.CentralEnd.Format("15:04:05"),
			s.PartialEnd.Format("15:04:05"))
		fmt.Printf("  Totality %s  mag %.4f  Sun alt %.1f°\n",
			s.CentralEnd.Sub(s.CentralStart).Round(time.Second), s.Magnitude, s.SunAltitude)
	}

	// 2. Lunar eclipse: 2029-01-01 total lunar eclipse
	d2029 := time.Date(2029, 1, 1, 12, 0, 0, 0, cst)
	if l, ok := eclipse.LunarEclipseOnDate(d2029); ok {
		fmt.Printf("2029-01-01 Lunar:    type=%s  Saros %d (%d/%d)\n",
			l.Type, l.Saros.Series, l.Saros.Member, l.Saros.Count)
		fmt.Printf("  P1 %s  U1 %s  U2 %s  Max %s  U3 %s  U4 %s  P4 %s\n",
			l.PenumbralStart.Format("01-02 15:04"), l.PartialStart.Format("15:04"),
			l.TotalStart.Format("15:04"), l.Maximum.Format("01-02 15:04:05"),
			l.TotalEnd.Format("15:04"), l.PartialEnd.Format("15:04"), l.PenumbralEnd.Format("15:04"))
		fmt.Printf("  Umbral mag %.3f  Penumbral mag %.3f\n", l.UmbralMagnitude, l.PenumbralMagnitude)
	}
}
```

Output:

```text
2009-07-22 Yangshan: type=total  Saros 136 (37/71)
  C1 08:23:55  C2 09:37:23  Max 09:40:21  C3 09:43:20  C4 11:03:14
  Totality 5m57s  mag 1.0770  Sun alt 57.3°
2029-01-01 Lunar:    type=total  Saros 125 (49/72)
  P1 12-31 22:03  U1 23:07  U2 00:16  Max 01-01 00:52:05  U3 01:27  U4 02:36  P4 03:40
  Umbral mag 1.246  Penumbral mag 2.274
```

> `ok == false` simply means no eclipse occurred on that date/location.

## Usage

### Solar Eclipses: Local Search, Multi-Year Scans & Central Paths

- **Global**: `SolarEclipseOnDate`, `Last/Next/ClosestSolarEclipse`; fast multi-year candidate lists without map geometry via `SolarEclipseCandidates(start, end, opt)`.
- **Local**: `LocalSolarEclipseOnDate`, `NextLocalSolarEclipse`, `NextLocalTotalSolarEclipse`, `NextLocalAnnularSolarEclipse`, and bounded century-scale searches via `SearchLocalCentralSolarEclipse`.
- **Global paths & footprints**: `SolarEclipseCentralPath` (centerline, north/south limits, width) and `SolarEclipsePartialFootprints` (partial visibility domain, P1–P4/U1–U4, optional greatest-time isochrones `GreatestTimeStep`, instant umbra `CentralShadowStep`).
- **Besselian elements**: `SolarEclipseBesselianElements(jdeTT, opt)` outputs cubic polynomials `X`/`Y`/`D`/`L1`/`L2`/`Mu` + `TanF1`/`TanF2` (convert `Mu` with `SolarEclipseBesselianMuForPublishedTable` when comparing against published tables).

```go
d2035 := time.Date(2035, 9, 2, 12, 0, 0, 0, cst)
if bj, ok := eclipse.LocalSolarEclipseOnDate(d2035, 116.4074, 39.9042, 0); ok {
	fmt.Printf("2035-09-02 Beijing: %s  max %s  totality %s  Saros %d\n",
		bj.Type, bj.GreatestEclipse.Format("15:04:05"),
		bj.CentralEnd.Sub(bj.CentralStart).Round(time.Second), bj.Saros.Series)
}
```

Output:

```text
2035-09-02 Beijing: total  max 08:33:38  totality 1m33s  Saros 145
```

### Lunar Eclipses: Danjon vs Chauvenet

- Global: `LunarEclipseOnDate`, `Last/Next/ClosestLunarEclipse`.
- Local: `LocalLunarEclipseOnDate`, `NextLocalLunarEclipse`, `NextLocalTotalLunarEclipse` (returning `Visibility`: `full`, `moonrise`, `moonset`, `rise-and-set`, `interrupted`, `penumbra-moonrise`, `penumbra-moonset`, `invisible`).
- **Shadow enlargement model**: defaults to **Danjon** (matching modern NASA catalogs). Append `Chauvenet` (e.g. `ClosestLunarEclipseChauvenet`) to reproduce classic tables (~`+0.025` penumbral mag, ~`+0.005` umbral mag).

### Lunar Limb Correction

By default the Moon is modeled as a smooth sphere. Import `github.com/starainrt/astro/moon/limb` (~1.5 MB Kaguya LALT profile, linked only when imported) to refine contact times, totality/annularity durations (typically seconds), and path limits (typically kilometres):

```go
import "github.com/starainrt/astro/moon/limb"

// Single call
info, ok := eclipse.LocalSolarEclipseOnDateWithOptions(date, lon, lat, 0,
	eclipse.SolarEclipseOptions{LunarLimb: limb.Default()})

// Or attach to an astro.Context for eclipses & occultations
base, _ := astro.NewContext(astro.ModelSemianalytic)
ctx := base.WithLunarLimb(limb.Default())
info, ok = ctx.Eclipse().LocalSolarEclipseOnDate(date, lon, lat, 0)
```

- All five `eclipse/svg` renderers accept `LunarLimb` and automatically note in the footer whether the limb profile was applied (`LunarLimbApplied`).

### Global & Local Eclipse Charts (SVG)

All five `eclipse/svg` functions return `(string, bool)` (`false` when no eclipse exists on that date or options are invalid). Set `Language: "en"` for English labels.

| Function | Chart Type | Recommended Size |
| --- | --- | --- |
| `LocalSolarEclipseSVG` | Local solar-disc phase diagram & contact table | 920×720+ |
| `SolarEclipseMapSVG` | Global visibility map (4 projections) | 1200×800 (1000×1414 orthographic) |
| `LunarEclipseSVG` | Moon path through Earth's shadow | 960×620+ |
| `LunarEclipseMapSVG` | Global lunar visibility map (includes penumbral bands by default) | 1200×800 |
| `LunarEclipseDetailedSVG` | Combined shadow diagram, coordinates & visibility map | 1000×1414 or 1414×1000 |

#### Local Solar Eclipse Charts

```go
svg, ok := eclipsesvg.LocalSolarEclipseSVG(
	time.Date(2009, 7, 22, 12, 0, 0, 0, cst),
	121.9850, 30.6167, 0,
	eclipsesvg.LocalSolarEclipseSVGOptions{Width: 920, Height: 720, Step: 5 * time.Minute, Location: cst, Language: "en"},
)
```

![2009 Yangshan total solar eclipse](../../img/solar-eclipse-yangshan-2009-en.svg)
![2012 Xiamen annular solar eclipse](../../img/solar-eclipse-xiamen-2012-en.svg)
![2035 Beijing total solar eclipse](../../img/solar-eclipse-beijing-2035-en.svg)

#### Global Solar Eclipse Maps & Projections

Switch `Projection` between `Equirectangular`, `Orthographic` (NASA globe layout), `NorthPolar`, and `SouthPolar`; set `TimeScale: astro.TimeScaleUT1` (with `Location: time.UTC`) for UT1 labels:

```go
globeSVG, ok := eclipsesvg.SolarEclipseMapSVG(
	time.Date(2009, 7, 22, 12, 0, 0, 0, cst),
	eclipsesvg.SolarEclipseMapSVGOptions{
		Width: 1000, Height: 1414, Location: cst, Language: "en",
		Projection:    eclipsesvg.EclipseMapProjectionOrthographic,
		TimeLabelStep: 30 * time.Minute, GreatestTimeStep: 30 * time.Minute,
	},
)
```

![2009 total solar eclipse global map](../../img/solar-eclipse-yangshan-2009-global-en.svg)
![2012 annular solar eclipse global map](../../img/solar-eclipse-xiamen-2012-global-en.svg)
![2035 Beijing total solar eclipse global map](../../img/solar-eclipse-beijing-2035-global-en.svg)
![2009 total solar eclipse orthographic globe](../../img/solar-eclipse-yangshan-2009-globe-en.svg)
![2012 annular eclipse north polar map](../../img/solar-eclipse-arctic-2012-global-en.svg)
![2021 Antarctic total solar eclipse south polar map](../../img/solar-eclipse-southpolar-2021-12-04-en.svg)
![2009 total solar eclipse global map (UT1)](../../img/solar-eclipse-yangshan-2009-global-ut1-en.svg)

#### Lunar Eclipse Charts

```go
d2029 := time.Date(2029, 1, 1, 0, 0, 0, 0, cst)
diagram, _ := eclipsesvg.LunarEclipseSVG(d2029, eclipsesvg.LunarEclipseSVGOptions{Width: 960, Height: 620, Step: 10 * time.Minute, Location: cst, Language: "en"})
world, _ := eclipsesvg.LunarEclipseMapSVG(d2029, eclipsesvg.LunarEclipseMapSVGOptions{Width: 1200, Height: 800, Location: cst, Language: "en"})
detailed, _ := eclipsesvg.LunarEclipseDetailedSVG(d2029, eclipsesvg.LunarEclipseDetailedSVGOptions{Location: cst, Language: "en"})
```

![2029 lunar eclipse diagram](../../img/lunar-eclipse-2029-01-01-en.svg)
![2029 lunar eclipse global map](../../img/lunar-eclipse-2029-01-01-global-en.svg)
![2029 lunar eclipse detailed sheet](../../img/lunar-eclipse-2029-01-01-detailed-en.svg)
![2026 lunar eclipse diagram](../../img/lunar-eclipse-2026-03-03-en.svg)
![2026 lunar eclipse detailed sheet](../../img/lunar-eclipse-2026-03-03-detailed-en.svg)

### Transit Charts

`eclipsesvg.TransitMapSVG` and `eclipsesvg.TransitLocalSVG` plot Mercury (`TransitMercury`) and Venus (`TransitVenus`) transits:

```go
venusDate := time.Date(2012, 6, 6, 0, 0, 0, 0, time.UTC)
world, ok := eclipsesvg.TransitMapSVG(eclipsesvg.TransitVenus, venusDate, eclipsesvg.TransitMapSVGOptions{
	Language:      "en",
	Isochrones:    eclipsesvg.TransitIsochroneExternal,
	IsochroneStep: 10 * time.Minute,
})
local, ok := eclipsesvg.TransitLocalSVG(eclipsesvg.TransitVenus, venusDate, 116.4, 39.9, 50,
	eclipsesvg.TransitLocalSVGOptions{Location: cst, Language: "en"})
```

![2032 Mercury transit Beijing local](../../img/transit-mercury-beijing-2032-en.svg)
![2032 Mercury transit global map](../../img/transit-mercury-beijing-2032-global-en.svg)
![2012 Venus transit Beijing local](../../img/transit-venus-beijing-2012-en.svg)
![2012 Venus transit global map](../../img/transit-venus-beijing-2012-global-en.svg)
![1937 partial Mercury transit (no geocentric contact)](../../img/transit-mercury-1937-partial-en.svg)

## API Reference

| Group | Entry Points | Notes |
| --- | --- | --- |
| Solar search | `SolarEclipseOnDate`, `LocalSolarEclipseOnDate`, `Last/Next/Closest*SolarEclipse`, `SearchLocalCentralSolarEclipse`, `SolarEclipseCandidates` | Plus `...WithOptions` and `...IAUSingleK` |
| Solar geometry | `SolarEclipseCentralPath`, `SolarEclipsePartialFootprints`, `SolarEclipseBesselianElements`, `NewSolarEclipseShadowSolver` | Paths, footprints, Besselian polynomials, interactive shadow solver |
| Lunar search | `LunarEclipseOnDate`, `LocalLunarEclipseOnDate`, `GeometricLocalLunarEclipseOnDate`, `Last/Next/Closest*LunarEclipse` | Danjon default; `Chauvenet` suffix available |
| UT1 helpers | `SolarEclipseInfoInUT1`, `LocalSolarEclipseInfoInUT1`, `LunarEclipseInfoInUT1`, etc. | Converts timestamps to UT1 |
| SVG charts | `LocalSolarEclipseSVG`, `SolarEclipseMapSVG`, `LunarEclipseSVG`, `LunarEclipseMapSVG`, `LunarEclipseDetailedSVG`, `TransitMapSVG`, `TransitLocalSVG` | In `github.com/starainrt/astro/eclipse/svg` |

## Conventions and Limitations

- **Radii**: Solar eclipses default to NASA Split-K lunar radii (`k=0.2724880` penumbra / `0.2722810` umbra) and standard solar radius (`959.639″` at 1 AU); edge-flash solar radius (`959.95″`) and IAU single-K (`0.2725076`) are selectable.
- **Comparing with NASA catalogs**: In TT/TD (independent of ΔT), modern solar/lunar eclipse maxima typically agree within ~0.1–0.4 s. Published NASA UT pages used projected ΔT values (`74–75 s` for 2024–2026) vs measured `~69.1 s`, causing a `5–6 s` UT offset unless aligned via `astro.SetDeltaT`.
- **Saros**: Verified NASA anchors (`Verified == true`) are used first, then precomputed tables (`-3000..+6000`), then live extrapolation (`Verified == false`).
- **Canvas minimums & UT1**: `SolarEclipseMapSVG` minimum `800×560` (falls back to `960×640`); `PartialStep` clamps to ≥ 2 min; `TimeScaleUT1` requires `Location` to be `nil` or `time.UTC` ([Time-Scale Declarations](map-geojson.md#time-scale-declarations)).
