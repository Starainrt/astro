# Stars

[中文](../star.md) | [Back to README](../../../README.en.md)

The `star` package ships a built-in catalog of 9100 bright stars (BSC / HR `1–9110`, mag `-1.46` to `7.96`) with Chinese & English common names, proper motion, constellation lookup, and rise/set calculations. The `star/catalog` subpackage loads external CSV, Hipparcos, BSC5, and Stellarium catalogs.

## Contents

- [Basic Example](#basic-example)
- [Usage](#usage)
  - [Lookup by Name or HR Number (and Handling Ambiguous Names)](#lookup-by-name-or-hr-number-and-handling-ambiguous-names)
  - [Bright-Star List & Batch Queries](#bright-star-list--batch-queries)
  - [External Catalogs](#external-catalogs)
- [API Reference](#api-reference)
- [Conventions and Limitations](#conventions-and-limitations)

## Basic Example

> **Important**: Catalog coordinates are stored at **J2000**. Always call `s.RaDecByDate(date)` (which applies proper motion, precession, and nutation) before passing RA/Dec to rise/set, altitude, or constellation APIs.

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro/star"
	"github.com/starainrt/astro/tools"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2026, 1, 1, 20, 0, 0, 0, cst)
	lon, lat := 116.4, 39.9 // Beijing

	sirius, err := star.StarDataByName("Sirius") // also accepts "天狼" or StarDataByHR(2491)
	if err != nil {
		log.Fatal(err)
	}
	ra, dec := sirius.RaDecByDate(date)

	rise, err := star.RiseTime(date, ra, dec, lon, lat, 0, true)
	if err != nil {
		log.Fatal(err)
	}
	culm := star.CulminationTime(date, ra, lon)
	set, err := star.SetTime(date, ra, dec, lon, lat, 0, true)
	if err != nil {
		log.Fatal(err)
	}
	alt := star.ApparentAltitude(date, ra, dec, lon, lat, 1010, 10)
	az := star.Azimuth(date, ra, dec, lon, lat)

	fmt.Printf("Sirius (%s / %s, mag %.2f): RA=%s Dec=%s\n",
		star.ConstellationEN(ra, dec, date), star.ConstellationCode(ra, dec, date),
		sirius.Mag, tools.Format(ra/15, 1), tools.Format(dec, 0))
	fmt.Printf("2026-01-01 20:00 Beijing: apparent alt %.2f°  az %.2f°\n", alt, az)
	fmt.Printf("Rise %s  Culmination %s  Set %s\n",
		rise.Format("01-02 15:04:05"), culm.Format("01-02 15:04:05"), set.Format("01-02 15:04:05"))
}
```

Output:

```text
Sirius (Canis Major / CMA, mag -1.46): RA=6h46m17.61s Dec=-16°45′12.59″
2026-01-01 20:00 Beijing: apparent alt 7.98°  az 119.66°
Rise 12-31 19:15:16  Culmination 01-01 00:19:18  Set 01-01 05:23:20
```

## Usage

### Lookup by Name or HR Number (and Handling Ambiguous Names)

- `StarDataByName` matches Chinese names/aliases and ~218 English common names (case-insensitive, e.g. `"sirius"`, `"Vega"`, `"Polaris"`).
- `StarDataByHR(2491)` always identifies a unique star.
- Some names match more than one star (e.g. `"Gienah"`, `"杵增一"`). `StarDataByName` returns `star.ErrStarNameAmbiguous` (and `star.ErrStarNotFound` when absent; check with `errors.Is`). Use `StarDataByNameAll` to list all matches or `StarDataByNameIn(name, region)` to disambiguate Chinese asterisms.

### Bright-Star List & Batch Queries

The built-in catalog lazy-loads once and caches read-only (call `star.InitStarDatabase()` only if you want to preload at startup). `star.TopBrightStars()` returns 169 bright stars (`mag ≤ ~3`):

```go
bright, err := star.TopBrightStars()
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Bright stars: %d; top 3: %s(%.2f), %s(%.2f), %s(%.2f)\n",
	len(bright), bright[0].CommonName, bright[0].Mag,
	bright[1].CommonName, bright[1].Mag, bright[2].CommonName, bright[2].Mag)
```

Output:

```text
Bright stars: 169; top 3: Sirius(-1.46), Canopus(-0.72), Arcturus(-0.04)
```

### External Catalogs

`star/catalog` loads external catalogs without auto-downloading large files. Each `entry.Coordinate` is a `basic.StarCoordinate` ready for `moon.FindStarOccultations` or `phenomena.StarApproaches`:

```go
input := strings.NewReader("id,ra,dec,pmra,pmdec,V\nSirius,101.28715533,-16.71611586,-546.01,-1223.07,-1.46\n")
stars, err := catalog.ReadCSV(input, catalog.ColumnMap{
	ID: "id", RA: "ra", Dec: "dec", PMRA: "pmra", PMDec: "pmdec", Magnitude: "V",
	PMRAIncludesCosDec: true, Band: catalog.BandV,
	Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: basic.CoordinateFrameJ2000,
}, catalog.ReadOptions{Strict: true})
if err != nil {
	log.Fatal(err)
}
entry, ok := stars.Lookup("Sirius")
```

| Entry Point | Purpose |
| --- | --- |
| `ReadCSV` | Headered CSV/TSV mapped via `ColumnMap` (requires `Epoch`/`EpochColumn` & `Frame`) |
| `ReadHipparcos2007` / `ReadBSC5` | Fixed-width text catalogs |
| `ReadStellariumStars` / `ReadStellariumDir` | Binary shards or `fs.FS` directory (supports `MaxMagnitude` filter) |
| `Lookup` / `Entries` / `Len` / `Near` / `Band` / `Merge` / `WithNames` | Query, cone/band search (at catalog epoch), merge & attach names |

## API Reference

| Group | Entry Points | Notes |
| --- | --- | --- |
| Catalog | `InitStarDatabase`, `StarDataByHR`, `StarDataByName`, `StarDataByNameIn`, `StarDataByNameAll`, `TopBrightStars` | Sentinels: `ErrStarNotFound`, `ErrStarNameAmbiguous` |
| Position | `StarData.RaDecByDate(date)` | J2000 → of-date (proper motion + precession + nutation; 3D when `Pc > 0`) |
| Constellation | `Constellation`, `ConstellationEN`, `ConstellationCode` | Chinese, English, or 3-letter IAU code (pass of-date RA/Dec) |
| Rise / set | `RiseTime`, `SetTime`, `CulminationTime`, `RiseTimeWithAtmosphere`, `SetTimeWithAtmosphere` | Sentinels: `ERR_STAR_NEVER_RISE`, `ERR_STAR_NEVER_SET` |
| Topocentric | `Altitude`, `ApparentAltitude`, `Azimuth`, `Zenith`, `ApparentZenith`, `HourAngle`, `ParallacticAngle` | Degrees; `Apparent*` takes pressure (hPa) & temp (°C) |
| Sidereal time | `MeanSiderealTime`, `ApparentSiderealTime` | Greenwich mean/apparent sidereal time (hours) |

## Conventions and Limitations

- **Circumpolar & never-rising stars**: `SetTime` returns `ERR_STAR_NEVER_SET` and `RiseTime` returns `ERR_STAR_NEVER_RISE` (`DownTime` and `ERR_STAR_NEVER_DOWN` are deprecated aliases).
- **3D space motion**: When distance `Pc > 0`, `RaDecByDate` advances a 3D position vector using proper motion and `RadVel`; without distance it falls back to 2D angular motion.
- **Refraction bounds**: `ApparentAltitude` applies Saemundsson refraction only for true altitudes in `(-5°, 90°)`; invalid pressure/temperature yields `NaN`.
- **Height & time scale**: `height` is WGS84 ellipsoidal height in metres ([Observer Height](coord.md#observer-height)); times are civil ([Time Scales](timescale.md)).
