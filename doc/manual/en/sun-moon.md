# Sun & Moon

[中文](../sun-moon.md) | [Back to README](../../../README.en.md)

This manual covers everyday Sun and Moon calculations: sunrise/sunset, twilight, moonrise/moonset, coordinates, lunar phases, apogees/perigees, nodes, maximum declinations, libration, Earth's perihelion/aphelion, and the lightweight `lite/sun` & `lite/moon` packages.

> See [Eclipses & Transits](eclipse.md) for eclipses, [Lunar Occultations](occultation.md) for occultations, and [Coordinates](coord.md) for general transforms.

## Contents

- [Basic Example](#basic-example)
- [Usage](#usage)
  - [Rise, Set & Twilight Notes](#rise-set--twilight-notes)
  - [Sun & Moon Positions](#sun--moon-positions)
  - [Lunar Phases & Conjunctions](#lunar-phases--conjunctions)
  - [Perigee, Apogee, Nodes & Earth Perihelion](#perigee-apogee-nodes--earth-perihelion)
  - [Libration & Bright-Limb Position Angle](#libration--bright-limb-position-angle)
  - [Lightweight Chain (lite/sun & lite/moon)](#lightweight-chain-litesun--litemoon)
- [Observing Angle Conventions](#observing-angle-conventions)
- [API Reference](#api-reference)
- [Conventions and Limitations](#conventions-and-limitations)

## Basic Example

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/sun"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2026, 1, 1, 12, 0, 0, 0, cst)
	lon, lat, height := 108.93, 34.27, 0.0 // Xi'an (deg E/N, ellipsoidal height m)

	twM, err := sun.MorningTwilight(date, lon, lat, -6) // -6° civil, -12° nautical, -18° astronomical
	if err != nil {
		log.Fatal(err)
	}
	sRise, err := sun.RiseTime(date, lon, lat, height, true)
	if err != nil {
		log.Fatal(err)
	}
	sCulm := sun.CulminationTime(date, lon)
	sSet, err := sun.SetTime(date, lon, lat, height, true)
	if err != nil {
		log.Fatal(err)
	}
	twE, err := sun.EveningTwilight(date, lon, lat, -6)
	if err != nil {
		log.Fatal(err)
	}

	mRise, err := moon.RiseTime(date, lon, lat, height, true)
	if err != nil {
		log.Fatal(err)
	}
	mCulm := moon.CulminationTime(date, lon, lat)
	mSet, err := moon.SetTime(date, lon, lat, height, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Xi'an 2026-01-01")
	fmt.Printf("Civil dawn: %s  Sunrise: %s  Noon: %s  Sunset: %s  Civil dusk: %s\n",
		twM.Format("15:04:05"), sRise.Format("15:04:05"), sCulm.Format("15:04:05"), sSet.Format("15:04:05"), twE.Format("15:04:05"))
	fmt.Printf("Moonrise: %s  Transit: %s  Moonset: %s\n",
		mRise.Format("15:04:05"), mCulm.Format("15:04:05"), mSet.Format("15:04:05"))
}
```

Output:

```text
Xi'an 2026-01-01
Civil dawn: 07:22:35  Sunrise: 07:49:59  Noon: 12:47:50  Sunset: 17:45:31  Civil dusk: 18:12:55
Moonrise: 15:14:55  Transit: 22:56:17  Moonset: 05:28:16
```

## Usage

The snippets below reuse Xi'an (`108.93°E, 34.27°N`) and `date := 2026-01-01 12:00:00 CST`.

### Rise, Set & Twilight Notes

- **The `aero` boolean**: `true` computes upper-limb crossings with standard refraction (`1010 hPa / 10 °C`) and apparent semidiameter; `false` computes geometric center crossing `0°`.
- **Custom weather**: Use `RiseTimeWithAtmosphere` / `SetTimeWithAtmosphere` to pass measured pressure (hPa) and temperature (°C).
- **Moonrise/moonset can skip a calendar day**: Because the Moon rises ~50 min later each day, moonset can precede moonrise (as above: `05:28` set, `15:14` rise), or a day may have no moonrise/moonset (`moon.ERR_NOT_TODAY`—query the next day).
- **Polar day/night**: Returns `ERR_SUN_NEVER_RISE` / `ERR_SUN_NEVER_SET` (and `ERR_MOON_NEVER_*`); check with `errors.Is`.

### Sun & Moon Positions

```go
sRA, sDec := sun.ApparentRaDec(date)
mRA, mDec := moon.ApparentRaDec(date, lon, lat) // topocentric for observing
mgRA, mgDec := moon.GeocentricApparentRaDec(date)

fmt.Printf("Sun:  RA %s  Dec %s  alt %.2f°  az %.2f°  dist %.6f AU  diam %.1f\"\n",
	tools.Format(sRA/15, 1), tools.Format(sDec, 0),
	sun.Altitude(date, lon, lat), sun.Azimuth(date, lon, lat),
	sun.EarthDistance(date), sun.Diameter(date))
fmt.Printf("True solar time: %s  EoT: %.2f min\n",
	sun.ApparentSolarTime(date, lon).Format("15:04:05"), sun.EquationTime(date)*60)
fmt.Printf("Moon: RA %s  Dec %s  alt %.2f°  az %.2f°  dist %.1f km  diam %.1f\"\n",
	tools.Format(mRA/15, 1), tools.Format(mDec, 0),
	moon.Altitude(date, lon, lat), moon.Azimuth(date, lon, lat),
	moon.EarthDistance(date), moon.Diameter(date))
fmt.Printf("Lunar parallax (topo-geo): dRA=%+.3f°  dDec=%+.3f°\n", mRA-mgRA, mDec-mgDec)
```

Output:

```text
Sun:  RA 18h46m42.87s  Dec -23°0′13.25″  alt 31.61°  az 167.10°  dist 0.983324 AU  diam 1951.0"
True solar time: 11:12:18  EoT: -3.40 min
Moon: RA 4h28m6.57s  Dec 25°58′57.50″  alt -26.01°  az 22.73°  dist 360486.1 km  diam 1986.5"
Lunar parallax (topo-geo): dRA=+0.363°  dDec=-0.848°
```

- **Topocentric vs geocentric**: Lunar parallax reaches nearly `1°`, so always use `moon.ApparentRaDec(date, lon, lat)` for pointing and `moon.GeocentricApparentRaDec(date)` for geocentric almanacs.
- **Distance units**: `sun.EarthDistance` returns **AU**; `moon.EarthDistance` returns **km**.

### Lunar Phases & Conjunctions

```go
fmt.Printf("Phase: %s (%.1f%% illuminated)\n", moon.PhaseDesc(date), moon.Phase(date)*100)
fmt.Printf("Next Full: %s  LastQ: %s  New: %s  FirstQ: %s\n",
	moon.NextFullMoon(date).Format("01-02 15:04"),
	moon.NextLastQuarter(date).Format("01-02 15:04"),
	moon.NextNewMoon(date).Format("01-02 15:04"),
	moon.NextFirstQuarter(date).Format("01-02 15:04"))
```

Output:

```text
Phase: 盈凸月 (92.5% illuminated)
Next Full: 01-03 18:02  LastQ: 01-10 23:48  New: 01-19 03:51  FirstQ: 01-26 12:47
```

- `Last*` and `Closest*` variants exist alongside `Next*` (as well as Pinyin aliases `ShuoYue`/`WangYue`/`ShangXianYue`/`XiaXianYue`).
- Find the closest planetary conjunction in RA with `moon.ClosestConjunctionWithPlanet(date, moon.ConjunctionJupiter)`.

### Perigee, Apogee, Nodes & Earth Perihelion

```go
peris := moon.PerigeesInMonth(2026, time.January)
apos := moon.ApogeesInMonth(2026, time.January)
fmt.Printf("2026-01 Moon perigee: %s UTC (%.0f km), apogee: %s UTC (%.0f km)\n",
	peris[0].Time.Format("01-02 15:04"), peris[0].Distance,
	apos[0].Time.Format("01-02 15:04"), apos[0].Distance)

ep := earth.Perihelion(2026)
ea := earth.Aphelion(2026)
fmt.Printf("2026 Earth perihelion: %s UTC (%.6f AU), aphelion: %s UTC (%.6f AU)\n",
	ep.Time.Format("01-02 15:04"), ep.Distance, ea.Time.Format("01-02 15:04"), ea.Distance)
```

Output:

```text
2026-01 Moon perigee: 01-01 21:44 UTC (360348 km), apogee: 01-13 20:47 UTC (405438 km)
2026 Earth perihelion: 01-03 17:15 UTC (0.983302 AU), aphelion: 07-06 17:30 UTC (1.016644 AU)
```

- Lunar orbital nodes: `moon.AscendingNode(date)` / `DescendingNode(date)`.
- Extreme lunar declinations: `moon.MaximumNorthDeclinationsInMonth`, `NextMaximumNorthDeclination` (and `South` counterparts).

### Libration & Bright-Limb Position Angle

```go
p := moon.TopocentricPhysical(date, lon, lat, 0)
fmt.Printf("Topocentric libration: lon %+.2f°  lat %+.2f°  axis PA %+.2f°  bright limb PA %.2f°\n",
	p.LibrationLongitude, p.LibrationLatitude, p.PositionAngle,
	moon.TopocentricBrightLimbPositionAngle(date, lon, lat, 0))
```

Output:

```text
Topocentric libration: lon -0.78°  lat -5.66°  axis PA -8.88°  bright limb PA 267.86°
```

Geocentric equivalents: `moon.Physical`, `moon.BrightLimbPositionAngle`, and `sun.Physical` (P / B0 / L0).

### Lightweight Chain (`lite/sun` & `lite/moon`)

Independent simplified formulas that do not load main ephemeris tables and ignore model switches:

```go
lRise, _ := litesun.RiseTime(date, lon, lat, 0, true)
fmt.Printf("Lite: sunrise %s (main %s)  phase %.2f%%  age %.2f days\n",
	lRise.Format("15:04:05"), sRise.Format("15:04:05"),
	litemoon.Phase(date)*100, litemoon.PhaseAge(date))
```

Output:

```text
Lite: sunrise 07:49:59 (main 07:49:59)  phase 92.63%  age 12.18 days
```

See [Accuracy & Performance](accuracy.md#lite-approximation-chain) for measured differences.

## Observing Angle Conventions

- **`Altitude` / `Zenith`**: Geometric altitude (`0°` horizon, `+90°` zenith) and complementary zenith distance (`Altitude + Zenith = 90°`), without refraction; use `ApparentAltitude` / `ApparentZenith` for refracted angles.
- **`Azimuth`**: Clockwise from north (`0°` N, `90°` E), `[0°, 360°)`.
- **`HourAngle`**: `0°` at upper culmination, increasing westward, `[0°, 360°)`.
- **`ParallacticAngle`**: Signed zenith direction angle in degrees.

## API Reference

| Group | `sun` | `moon` | Units / Notes |
| --- | --- | --- | --- |
| Coordinates | `TrueLo`, `ApparentLo`, `TrueBo`, `ApparentRaDec`, `EclipticObliquity` | `TrueLo`, `ApparentLo`, `TrueBo`, `TrueRaDec`, `GeocentricApparentRaDec`, `ApparentRaDec` | Degrees; `moon.ApparentRaDec` is topocentric |
| Rise / set | `RiseTime`, `SetTime`, `CulminationTime`, `MorningTwilight`, `EveningTwilight` | `RiseTime`, `SetTime`, `CulminationTime` | Plus `...WithAtmosphere` variants |
| Horizontal | `Altitude`, `ApparentAltitude`, `Azimuth`, `Zenith`, `HourAngle`, `ParallacticAngle` | Same | Degrees |
| Distance & size | `EarthDistance` (AU), `Diameter`, `Semidiameter` (arcsec) | `EarthDistance` (km), `Diameter`, `Semidiameter` (arcsec) | Sun in AU, Moon in km |
| Solar time / phases | `ApparentSolarTime`, `EquationTime` (hours) | `Phase`, `PhaseDesc`, `SunMoonLoDiff`, `Next/Last/ClosestNewMoon` (etc.) | Both English & Pinyin phase names |
| Physical & orbital | `Physical`; `earth.Perihelion/Aphelion/EarthEccentricity` | `Physical`, `TopocentricPhysical`, `BrightLimbPositionAngle`, `PerigeesInMonth`, `ApogeesInMonth`, `AscendingNode`, `MaximumNorthDeclinationsInMonth` | Deg / km |

## Conventions and Limitations

- **`...N` functions**: `n < 0` evaluates the full series; under `ModelSemianalytic`, any `n >= 0` uses a single fast table suitable only for coarse searches (lunar errors reach arcseconds).
- **Time scale & height**: Inputs are civil times ([Time Scales](timescale.md)) and heights are WGS84 ellipsoidal metres ([Observer Height](coord.md#observer-height)). Never pass `sun.ApparentSolarTime` back into ephemeris APIs.
- **Sentinels & aliases**: `ERR_SUN_NEVER_RISE/SET`, `ERR_MOON_NEVER_RISE/SET`, `ERR_TWILIGHT_NOT_EXISTS`, `ERR_NOT_TODAY`. `DownTime` and `ERR_*_NEVER_DOWN` are deprecated aliases for `SetTime` and `ERR_*_NEVER_SET`.
