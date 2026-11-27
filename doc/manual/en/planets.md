# Planets

[中文](../planets.md) | [Back to README](../../../README.en.md)

Each of the seven planets has its own package (`mercury`, `venus`, `mars`, `jupiter`, `saturn`, `uranus`, `neptune`) sharing one consistent API shape, plus planet-specific features:
- **Inner planets (`mercury`, `venus`)**: superior/inferior conjunctions, greatest elongations, geocentric & local solar transits.
- **Outer planets (`mars`–`neptune`)**: oppositions and eastern/western quadratures.
- **Special capabilities**: Jupiter's Galilean moons & mutual events (`jupiter`), Saturn's rings (`saturn`), Mars Sol Date / MTC / Mars sunrise (`mars`, `sun`).

> See [Jupiter Satellite SVG](jupiter-svg.md) for satellite charts, [Transit Charts](eclipse.md#transit-charts) for Mercury/Venus transit SVGs, and [Lunar Occultations](occultation.md) for planetary occultations.

## Contents

- [Basic Example](#basic-example)
- [Usage](#usage)
  - [Oppositions, Conjunctions, Elongations & Stations](#oppositions-conjunctions-elongations--stations)
  - [Mercury & Venus Transits](#mercury--venus-transits)
  - [Physical Ephemerides & Saturn's Rings](#physical-ephemerides--saturns-rings)
  - [Galilean Satellites: Phenomena, Contacts & Mutual Events](#galilean-satellites-phenomena-contacts--mutual-events)
  - [Planetary Local Time & Jupiter Rotation Angle](#planetary-local-time--jupiter-rotation-angle)
- [API Reference](#api-reference)
- [Conventions and Limitations](#conventions-and-limitations)

## Basic Example

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro/mars"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2026, 1, 15, 20, 0, 0, 0, cst)
	lon, lat := 108.93, 34.27 // Xi'an (deg E/N)

	ra, dec := mars.ApparentRaDec(date)
	rise, err := mars.RiseTime(date, lon, lat, 0, true)
	if err != nil {
		log.Fatal(err)
	}
	set, err := mars.SetTime(date, lon, lat, 0, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Mars (2026-01-15 20:00 CST): RA=%.2f° Dec=%.2f° mag=%.2f dist=%.3f AU\n",
		ra, dec, mars.ApparentMagnitude(date), mars.EarthDistance(date))
	fmt.Printf("Xi'an rise/set: %s / %s  next opposition: %s\n",
		rise.Format("15:04:05"), set.Format("15:04:05"),
		mars.NextOpposition(date).Format("2006-01-02 15:04"))
}
```

Output:

```text
Mars (2026-01-15 20:00 CST): RA=295.93° Dec=-22.29° mag=1.15 dist=2.398 AU
Xi'an rise/set: 07:50:30 / 17:46:23  next opposition: 2027-02-19 23:50
```

> Replace `mars` with any other planet package for the same core API, or call `ctx.Mars()` to use an explicit ephemeris context.

## Usage

### Oppositions, Conjunctions, Elongations & Stations

Event searches come in `Last*` / `Next*` pairs preserving the input time zone:

```go
fmt.Printf("Next Venus greatest eastern elongation: %s\n",
	venus.NextGreatestElongationEast(date).Format("2006-01-02 15:04"))
fmt.Printf("Next Mercury inferior conjunction:      %s\n",
	mercury.NextInferiorConjunction(date).Format("2006-01-02 15:04"))
```

Output:

```text
Next Venus greatest eastern elongation: 2026-08-15 14:31
Next Mercury inferior conjunction:      2026-03-07 19:01
```

- **Stations**: All seven packages provide `Last/NextProgradeToRetrograde` and `Last/NextRetrogradeToPrograde` (`mercury` & `venus` also provide directionless `Last/NextRetrograde`).
- **Inner planets**: `SuperiorConjunction`, `InferiorConjunction`, `GreatestElongation[East/West]`.
- **Outer planets**: `Opposition`, `EasternQuadrature`, `WesternQuadrature`.

### Mercury & Venus Transits

Search the global event with `NextTransit` / `LastTransit` / `ClosestTransit` (check `tr.Valid`), then evaluate local visibility and C1–C4 contacts with `LocalTransit`:

```go
tr := mercury.NextTransit(date)
if tr.Valid {
	ltr := mercury.LocalTransit(tr, 116.4, 39.9, 50) // Beijing
	fmt.Printf("Next Mercury transit: %s UTC (C1=%s, greatest=%s, C4=%s, Beijing visible=%v)\n",
		tr.Greatest.UTC().Format("2006-01-02"),
		tr.Start.UTC().Format("15:04:05"),
		tr.Greatest.UTC().Format("15:04:05"),
		tr.End.UTC().Format("15:04:05"),
		ltr.Visible)
}
```

Output:

```text
Next Mercury transit: 2032-11-13 UTC (C1=06:41:13, greatest=08:54:13, C4=11:07:16, Beijing visible=true)
```

> Grazing or edge-only transits may have `HasInternal == false` or even `HasExternal == false` geocentrically; always check `LocalTransit` for a specific observing site.

### Physical Ephemerides & Saturn's Rings

- Every planet package offers `Physical(date)` (sub-Earth/sub-Solar coordinates, north-pole PA).
- Jupiter adds `CentralMeridians(date)` (Systems I/II/III); Saturn & Uranus add `PhysicalSystemIII(date)`; Saturn adds `Ring(date)`:

```go
ring := saturn.Ring(date)
fmt.Printf("Saturn's rings: tilt B=%.2f°  major axis=%.1f\"  minor axis=%.1f\"\n",
	ring.EarthLatitude, ring.MajorAxis, ring.MinorAxis)
```

Output:

```text
Saturn's rings: tilt B=-1.47°  major axis=37.8"  minor axis=1.0"
```

### Galilean Satellites: Phenomena, Contacts & Mutual Events

The `jupiter` package models Io, Europa, Ganymede, and Callisto (`GalileanSatelliteIo` … `Callisto`) at four levels:
1. **Instantaneous positions & flags**: `Satellites(date)` and `SatellitePhenomena(date)`.
2. **Regular phenomena**: point-center `NextGalileanPhenomenonEvent` (fast interval search) vs finite-disc IMCCE-style D/F windows `NextGalileanPhenomenonContactEvent` (preferred for observing predictions).
3. **Mutual occultations & eclipses**: `NextGalileanMutualEvent(date, satA, satB, kind)` returns active/passive moons, contacts, geometric `Coverage` (`0..1`), and `Annular`.
4. **Batch schedule**: `GalileanEventsBetween(start, end, options)` collects all phenomena, contacts, and mutual events in chronological order (up to 366 days).

```go
satName := map[int]string{
	jupiter.GalileanSatelliteIo:       "Io",
	jupiter.GalileanSatelliteEuropa:   "Europa",
	jupiter.GalileanSatelliteGanymede: "Ganymede",
	jupiter.GalileanSatelliteCallisto: "Callisto",
}
t := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

sats := jupiter.Satellites(t)
fmt.Printf("Io (2026-06-01 00:00 UTC): X=%+.2f Rj  Y=%+.2f Rj  inFront=%v\n",
	sats.Io.OffsetXJupiterR, sats.Io.OffsetYJupiterR, sats.Io.InFrontOfJupiter)

occ := jupiter.NextGalileanMutualEvent(t, jupiter.GalileanSatelliteIo, jupiter.GalileanSatelliteEuropa, jupiter.GalileanMutualOccultation)
if occ.Valid {
	fmt.Printf("Mutual occultation: %s occults %s  %s – %s UTC  max %s  cov %.1f%%  dur %s\n",
		satName[occ.Active], satName[occ.Passive],
		occ.Start.Format("2006-01-02 15:04:05"), occ.End.Format("15:04:05"),
		occ.Greatest.Format("15:04:05"), occ.Coverage*100, occ.Duration.Round(time.Second))
}

ecl := jupiter.NextGalileanMutualEvent(t, jupiter.GalileanSatelliteIo, jupiter.GalileanSatelliteEuropa, jupiter.GalileanMutualEclipse)
if ecl.Valid {
	fmt.Printf("Mutual eclipse:     %s eclipses %s  %s – %s UTC  max %s  umbra %.1f%%  dur %s\n",
		satName[ecl.Active], satName[ecl.Passive],
		ecl.Start.Format("2006-01-02 15:04:05"), ecl.End.Format("15:04:05"),
		ecl.Greatest.Format("15:04:05"), ecl.Coverage*100, ecl.Duration.Round(time.Second))
}

all, err := jupiter.GalileanEventsBetween(t, t.Add(7*24*time.Hour), jupiter.GalileanEventsOptions{
	IncludeMutual: true,
})
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Total Galilean events in 7 days: %d (first: %s %s)\n",
	len(all), all[0].Greatest.Format("01-02 15:04"), all[0].Family)
```

Output:

```text
Io (2026-06-01 00:00 UTC): X=-2.21 Rj  Y=+0.56 Rj  inFront=false
Mutual occultation: Io occults Europa  2026-06-01 01:56:24 – 02:26:20 UTC  max 02:11:26  cov 100.0%  dur 29m56s
Mutual eclipse:     Io eclipses Europa  2026-06-01 13:52:46 – 14:02:19 UTC  max 13:57:34  umbra 4.9%  dur 9m33s
Total Galilean events in 7 days: 35 (first: 06-01 02:48 phenomenon)
```

- Mutual events include all geometric events even when the passive satellite is occulted or eclipsed by Jupiter; check `event.GreatestState[event.Passive-1]` for observability.
- If a regular Jovian eclipse lacks full D/F windows, both segments set `PenumbralFallback: true`. See [Jupiter Satellite SVG](jupiter-svg.md) for charts.

### Planetary Local Time & Jupiter Rotation Angle

```go
tM := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
mRise, err := sun.NextRiseOnMars(tM, 137.4, -4.6) // Gale Crater (137.4°E, -4.6°N)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Mars (2026-11-14 00:00 UTC): MSD=%.2f  MTC=%.2f h  Ls=%.2f°  Gale LMST=%.2f h  next sunrise=%s UTC\n",
	mars.SolDate(tM), mars.CoordinatedTime(tM), mars.SolarLongitude(tM),
	mars.LocalMeanSolarTime(tM, 137.4), mRise.Format("15:04:05"))
fmt.Printf("Jupiter System III W: %.2f°\n", jupiter.RotationAngle(tM, jupiter.SystemIII))
```

Output:

```text
Mars (2026-11-14 00:00 UTC): MSD=54342.55  MTC=13.27 h  Ls=21.58°  Gale LMST=22.43 h  next sunrise=08:15:55 UTC
Jupiter System III W: 130.68°
```

- `mars.LocalMeanSolarTime` follows the Mars24 linear convention (`MTC + eastLon/15`); `jupiter.RotationAngle` supports `SystemI`, `SystemII`, and `SystemIII` (ICRF prime meridian angle W at the requested instant, without light-time subtraction). For all planets and stars on planetary surfaces, see [Planetary Surface Observations](phenomena.md#planetary-surface-observations).

## API Reference

| Capability | Packages | Representative Entry Points | Units |
| --- | --- | --- | --- |
| Geocentric positions | All 7 | `ApparentLo/Bo`, `ApparentRa/Dec/RaDec`, `EarthDistance`, `SunDistance` | deg / AU |
| Horizontal & rise/set | All 7 | `Altitude`, `Azimuth`, `Zenith`, `RiseTime`, `SetTime`, `CulminationTime` | deg / `time.Time` |
| Photometry & size | All 7 | `ApparentMagnitude`, `Diameter`, `Semidiameter`, `PhaseAngle`, `IlluminatedFraction` (`Phase`), `BrightLimbPositionAngle` | mag / arcsec / deg |
| Nodes, conjunctions & stations | All 7 | `AscendingNode`, `DescendingNode`, `Last/NextConjunction`, `Last/NextProgradeToRetrograde`, `Last/NextRetrogradeToPrograde` | deg / `time.Time` |
| Inner planet events | `mercury`, `venus` | `SuperiorConjunction`, `InferiorConjunction`, `GreatestElongation[East/West]`, `Last/Next/ClosestTransit`, `LocalTransit` | `time.Time` / `TransitInfo` |
| Outer planet events | `mars`–`neptune` | `Opposition`, `EasternQuadrature`, `WesternQuadrature` | `time.Time` |
| Physical & rings | All 7 | `Physical`, `jupiter.CentralMeridians`, `saturn.Ring`, `saturn/uranus.PhysicalSystemIII` | deg / arcsec |
| Galilean moons | `jupiter` | `Satellites`, `SatellitePhenomena`, `*GalileanPhenomenonEvent`, `*GalileanPhenomenonContactEvent`, `*GalileanMutualEvent`, `GalileanEventsBetween` | Rj / arcsec / `time.Time` |
| Mars clock & Jupiter W | `mars`, `sun`, `jupiter` | `mars.SolDate`, `CoordinatedTime`, `LocalMeanSolarTime`, `LocalTrueSolarTime`, `SolarLongitude`, `sun.NextRiseOnMars/NextSetOnMars`, `jupiter.RotationAngle` | Sol / hours / deg |

## Conventions and Limitations

- **Geocentric vs topocentric**: `ApparentRaDec`, `HourAngle`, and `ParallacticAngle` are geocentric; `Altitude`/`Azimuth`/`Zenith` and `RiseTime`/`SetTime` are topocentric (`height` in ellipsoidal metres; polar/missing events return `ERR_<PLANET>_NEVER_RISE/SET`; `DownTime` is a deprecated alias for `SetTime`).
- **`...N` variants**: Instantaneous functions offer `...N(..., n)` (`n < 0` full series; under `ModelSemianalytic`, `n >= 0` uses the fast search table). Event searches have no `N` variant.
- **Physical longitude sign**: Follows IAU cartographic convention—west-positive for Mercury, Mars, Jupiter, Saturn, and Neptune; east-positive for Venus and Uranus.
- **Galilean accuracy**: Mutual-event timings typically agree with JPL Horizons within seconds to tens of seconds, suitable for observing predictions and charts.
