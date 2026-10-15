# Stars

[中文](../star.md) | [Back to README](../../../README.en.md)

> Full examples in this manual run from the repository root.

The library ships a catalog of 9100 stars (BSC / HR numbers `1-9110`, apparent magnitudes `-1.46` to `7.96`) and propagates proper motion automatically. The catalog stores **J2000** right ascension and declination (`InnerStarData.Ra`/`Dec`); to get the position at a given instant you must apply proper motion, precession and nutation through `StarData.RaDecByDate(date)`. Rise/set, topocentric quantities and constellation lookup all expect the corrected coordinates.

## Contents

- [External catalogs](#external-catalogs)
- [Finding Sirius and its rise time](#finding-sirius-and-its-rise-time)
- [API Reference](#api-reference)
  - [Constellation lookup](#constellation-lookup)
  - [Star catalog](#star-catalog)
  - [Rise, set and culmination](#rise-set-and-culmination)
  - [Topocentric quantities](#topocentric-quantities)
  - [Sidereal time](#sidereal-time)
  - [Calculating observing quantities for bright stars](#calculating-observing-quantities-for-bright-stars)
  - [Complete example](#complete-example)
  - [Bulk lookups and caching](#bulk-lookups-and-caching)
- [Usage examples](#usage-examples)
  - [Can I see this star tonight?](#can-i-see-this-star-tonight)
  - [Using a J2000 position at a given instant](#using-a-j2000-position-at-a-given-instant)
  - [Constellation lookup and the bright-star catalog](#constellation-lookup-and-the-bright-star-catalog)
  - [Polar edges and catalog conventions](#polar-edges-and-catalog-conventions)
- [Parameter and result conventions](#parameter-and-result-conventions)
- [Related manuals](#related-manuals)

## External catalogs

`star/catalog` is independent of the embedded 9100-star database and does not automatically download or load large catalogs. `Entry.Coordinate` is a `basic.StarCoordinate` with epoch, frame and proper motion; pass it directly to `moon.FindStarOccultations` or `phenomena.StarApproaches`.

| Entry point | Input and purpose |
| --- | --- |
| `ReadCSV` | Header-based CSV / TSV; `ColumnMap` sets columns, epoch and frame; zero `Delimiter` enables detection |
| `ReadHipparcos2007` / `ReadBSC5` | Corresponding fixed-width catalog text |
| `ReadStellariumStars` / `ReadStellariumDir` | One binary shard or an `fs.FS` directory; directory options load names/cross-IDs and filter by `MaxMagnitude` |
| `Lookup` / `Entries` / `Len` | ID or alias lookup, copied records and count |
| `Near` / `Band` | Cone / declination-band searches at catalog epochs, without propagation to the observing date |
| `Merge` / `WithNames` | Return new catalogs; right-hand records win for the same primary ID and component; names use `id\|name` lines |

```go
import (
	"strings"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/star/catalog"
)

input := strings.NewReader("id,ra,dec,pmra,pmdec,V\nSirius,101.28715533,-16.71611586,-546.01,-1223.07,-1.46\n")
stars, err := catalog.ReadCSV(input, catalog.ColumnMap{
	ID: "id", RA: "ra", Dec: "dec", PMRA: "pmra", PMDec: "pmdec", Magnitude: "V",
	PMRAIncludesCosDec: true, Band: catalog.BandV,
	Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: basic.CoordinateFrameJ2000,
}, catalog.ReadOptions{Strict: true})
if err != nil {
	panic(err)
}
entry, found := stars.Lookup("Sirius")
_ = entry.Coordinate
_ = found
```

Units and required fields:

- `RA` and `Dec` default to degrees; `AnglesInRadians` selects radians. Proper motions are mas/year, parallax mas and radial velocity km/s.
- Set `PMRAIncludesCosDec: true` if PMRA already represents `dRA*cos(Dec)`; otherwise the reader multiplies by the cosine.
- Supply `Epoch` or `EpochColumn` (Julian year) and the correct `Frame`; a zero epoch is not an unknown-epoch placeholder.

Reading behavior:

- With `ReadOptions.Strict`, invalid records return an error; otherwise they are skipped and counted by `Skipped()`. Structural errors such as missing required headers always fail.
- `Filter` controls admission without counting rejected records as bad.
- Missing magnitudes are `NaN`; do not use `BandV`, `BandG` and `BandHp` interchangeably for a V-band visibility threshold.
- `Near` searches at each record's own epoch: normalize mixed epochs/frames first and apply proper-motion propagation for exact observing-date selection.

## Finding Sirius and its rise time

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro/star"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2020, 1, 1, 8, 8, 8, 0, cst)
	lon, lat, height := 115.0, 40.0, 0.0
	sirius, err := star.StarDataByName("天狼")
	if err != nil {
		log.Fatal(err)
	}
	ra, dec := sirius.RaDecByDate(date)
	rise, err := star.RiseTime(date, ra, dec, lon, lat, height, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(star.Constellation(ra, dec, date))
	fmt.Printf("RA=%.6f Dec=%.6f deg\n", ra, dec)
	fmt.Println(rise.Format(time.RFC3339))
}
```

Catalog coordinates have epoch J2000. `RaDecByDate` applies proper motion, precession and nutation before date-specific rise/set, altitude or constellation calculations.

Proper motion is propagated in **Julian years** (365.25 days), and `RaDecByDate` converts the civil instant to TT before taking the epoch difference. Whenever a record carries a distance (`Pc > 0`), `RaDecByJde` advances a full **three-dimensional space motion**: proper motion supplies the tangential velocity and `RadVel` the line-of-sight component, the position vector is extrapolated linearly and its direction is taken. Records without a distance fall back to two dimensions, advancing only the two angular components, which is equivalent to treating the star as infinitely distant. The two differ only in second-order terms: three-dimensional propagation also changes the star's distance, so the apparent place is slightly smaller than in the two-dimensional case; even the fastest bright star moves less than about 0.09 arcseconds from it within 26 years.

## API Reference

| Group | Entry points | Purpose | Units and convention |
| --- | --- | --- | --- |
| Constellation | `Constellation` / `ConstellationEN` / `ConstellationCode` | Chinese name / English name / IAU three-letter code | Corrected RA/Dec in degrees plus the instant |
| Star catalog | `InitStarDatabase` / `StarDataByHR` / `StarDataByName` / `TopBrightStars` | Initialise the embedded catalog, look up by HR number or Chinese name, fetch the bright-star sample | HR `1-9110`; `Mag` is apparent magnitude |
| Coordinate correction | `StarData.RaDecByDate` (`RaDecByJde` on the `basic` side) | Correct a J2000 position to the given instant | Returns RA/Dec in degrees |
| Rise, set, culmination | `RiseTime` / `SetTime` / `CulminationTime` (`DownTime` is a deprecated alias of `SetTime`) | Rise, set and culmination instants | Civil instants; polar cases return sentinel errors |
| Rise/set with weather | `RiseTimeWithAtmosphere` / `SetTimeWithAtmosphere` | Rise and set with measured pressure (hPa) and temperature (°C) | Refraction always applied |
| Topocentric quantities | `Altitude` / `ApparentAltitude` / `Azimuth` / `Zenith` / `ApparentZenith` | (Apparent) altitude, azimuth, (apparent) zenith distance | Degrees |
| Hour angle and parallactic angle | `HourAngle` / `ParallacticAngle` | Stellar hour angle, zenith-direction parallactic angle | Degrees |
| Sidereal time | `MeanSiderealTime` / `ApparentSiderealTime` | Mean and apparent sidereal time | Hours |

The `star` package has no `...N` entry points; the `n`-parameterised evaluation functions live on the matching functions of `sun`, `moon`, the planets and `coord`, with the semantics described in those manuals (`n < 0` uses the full series; under the semianalytic model every `n >= 0` uses the fast table, and Legacy truncates by `n`).

The snippets below omit shared preamble variables: `date` (observation instant, civil time scale), `lon`/`lat` (observer longitude/latitude, east/north positive, degrees), `height` (observer height, **ellipsoidal**, metres), `aero` (whether refraction and apparent-radius corrections are applied).

### Constellation lookup

```go
sirius, _ := star.StarDataByName("天狼")
ra, dec := sirius.RaDecByDate(date)
fmt.Println(star.Constellation(ra, dec, date))     // 大犬座
fmt.Println(star.ConstellationEN(ra, dec, date))   // Canis Major
fmt.Println(star.ConstellationCode(ra, dec, date)) // CMA
```

All three share one boundary table and differ only in output convention. The lookup uses the **apparent position of the day**, so pass the result of `RaDecByDate`, never the J2000 coordinates from the catalog.

### Star catalog

```go
_ = star.InitStarDatabase()
s, _ := star.StarDataByHR(2491)
fmt.Println(s.HR, s.ChineseName, s.CommonName, s.Mag) // 2491 天狼 Sirius -1.46
bright, _ := star.TopBrightStars()
fmt.Println(len(bright), bright[0].HR) // size of the bright-star sample and the leading HR number
```

The catalog loads on first lookup; `InitStarDatabase()` only warms it up. It is idempotent and surfaces the load error explicitly, and skipping it still works because the first lookup loads the catalog automatically.

`TopBrightStars()` returns 169 built-in entries around magnitude 3 or brighter, roughly ordered from brighter to dimmer.

`basic.StarData` adds naming fields on top of `InnerStarData`, whose conventions are:

| Field | Meaning |
| --- | --- |
| `HR` / `HD` / `HIP` | Bright Star number (`1-9110`) / Henry Draper number / Hipparcos number |
| `Ra` / `Dec` | **J2000** right ascension and declination in degrees (run `RaDecByDate` for the apparent position) |
| `Mag` | Apparent magnitude |
| `PmRA` / `PmDec` | Annual proper motion `cos(dec)*dRA/dt` and in declination, arcseconds per year |
| `RadVel` / `RotVel` | Radial and proper-motion velocity, km/s |
| `Pc` | Distance in parsecs; `> 0` propagates proper motion as a 3D space motion |
| `ChineseName` / `ChineseAlias` / `ChineseBayerName` | Chinese name, alias and Bayer designation |
| `CommonName` / `CommonAliasName` | Common English name and alias |
| `Cst` / `CstChinese` | Constellation name in English and Chinese |

Lookup by name matches the catalog's Chinese names only (`天狼` for Sirius, `织女一` for Vega); English names are not matched. Use `StarDataByHR` to look up by number.

### Rise, set and culmination

```go
ra, dec := 101.28715533, -16.71611586 // Sirius-like sample coordinates near J2000
rise, _ := star.RiseTime(date, ra, dec, lon, lat, height, true)
set, _ := star.SetTime(date, ra, dec, lon, lat, height, true)
fmt.Println(rise, set)
fmt.Println(star.CulminationTime(date, ra, lon))
```

With `aero` true the rise/set solution uses the horizon corrected by refraction and apparent radius; false uses the geometric horizon. `CulminationTime` needs only right ascension and longitude.

Under a polar night or midnight sun, `RiseTime`/`SetTime` return sentinel errors instead of an instant, see below.

### Topocentric quantities

```go
fmt.Println(star.Altitude(date, ra, dec, lon, lat))
fmt.Println(star.ApparentAltitude(date, ra, dec, lon, lat, 1010, 10))
fmt.Println(star.Azimuth(date, ra, dec, lon, lat), star.Zenith(date, ra, dec, lon, lat))
fmt.Println(star.HourAngle(date, ra, lon), star.ParallacticAngle(date, ra, dec, lon, lat))
```

`ApparentAltitude`/`ApparentZenith` take two extra arguments, pressure (hPa) and temperature (C), for the refraction correction. `ParallacticAngle` is the usual input for rotating a camera or a spectrograph slit.

### Sidereal time

```go
fmt.Println(star.MeanSiderealTime(date), star.ApparentSiderealTime(date))
```

Both return hours; the relation between sidereal time, hour angle and horizontal transforms is documented in the `coord` manual.

### Calculating observing quantities for bright stars

```go
bright, _ := star.TopBrightStars()
for _, s := range bright[:3] {
	ra, dec := s.RaDecByDate(date)
	alt := star.ApparentAltitude(date, ra, dec, lon, lat, 1010, 10)
	fmt.Printf("%-6s %-10s mag=%.2f alt=%.3f\n", s.ChineseName, star.ConstellationEN(ra, dec, date), s.Mag, alt)
}
```

Run at `date = 2020-01-01 08:08:08 CST` from `115°E, 40°N` with pressure `1010 hPa` and temperature `10 °C`, this prints:

```text
天狼     Canis Major    mag=-1.46 alt=-30.180
老人     Carina         mag=-0.72 alt=-48.661
大角     Bootes         mag=-0.04 alt=68.926
```

### Complete example

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/star"
	"github.com/starainrt/astro/tools"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	// Observation instant.
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)

	// Initialise the star catalog.
	_ = star.InitStarDatabase()
	sirius, _ := star.StarDataByName("天狼")
	ra, dec := sirius.RaDecByDate(date)
	// Sirius rising.
	riseDate, _ := star.RiseTime(date, ra, dec, 115, 40, 0, true)
	fmt.Println(riseDate)
	// Sirius setting.
	setDate, _ := star.SetTime(date, ra, dec, 115, 40, 0, true)
	fmt.Println(setDate)
	fmt.Println(star.Constellation(ra, dec, date))

	// Vega.
	vega, _ := star.StarDataByName("织女一")
	ra, dec = vega.RaDecByDate(time.Date(13600, 1, 1, 0, 0, 0, 0, time.Local))
	// Vega right ascension in 13600 CE.
	fmt.Println(tools.Format(ra/15, 1))
	// Vega declination in 13600 CE.
	fmt.Println(tools.Format(dec, 0))

	bright, _ := star.TopBrightStars()
	fmt.Println(bright[0].ChineseName, bright[0].CommonName, bright[0].Mag)
}
```

Output:

```text
2019-12-31 19:22:56.32125467 +0800 CST // Sirius rising
2020-01-01 05:30:39.979518949 +0800 CST // Sirius setting
大犬座 // constellation of Sirius
6h4m44.89s // Vega right ascension in 13600 CE
84°19′13.71″ // Vega declination in 13600 CE
天狼 Sirius -1.46 // first bright-star entry: Chinese name, common name, magnitude
```

The printed constellation name is the Chinese one because `Constellation` is the Chinese-name entry point; use `ConstellationEN` for `Canis Major` and `ConstellationCode` for `CMA`.

### Bulk lookups and caching

```go
for _, hr := range []int{2491, 2326, 5340} {
	s, err := star.StarDataByHR(hr)
	if err != nil {
		continue
	}
	ra, dec := s.RaDecByDate(date)
	fmt.Println(s.ChineseName, star.ConstellationCode(ra, dec, date), s.Mag)
}
```

The catalog loads once on first access and is then a read-only cache, so share it across calls and goroutine-free request paths instead of caching it yourself. Unknown numbers or names return an error; in bulk code it is enough to skip on `err != nil` without classifying the error.

## Usage examples

### Can I see this star tonight?

```go
ra, dec := sirius.RaDecByDate(date)
fmt.Println(star.RiseTime(date, ra, dec, lon, lat, height, true)) // rise
fmt.Println(star.SetTime(date, ra, dec, lon, lat, height, true))  // set
fmt.Println(star.CulminationTime(date, ra, lon))                  // culmination
fmt.Println(star.ApparentAltitude(date, ra, dec, lon, lat, 1010, 10), star.Azimuth(date, ra, dec, lon, lat))
```

`aero = true` solves rise/set against the horizon corrected for refraction and apparent radius, which sits closer to the visual "just above the horizon" moment; `false` uses the geometric horizon, and the two differ on the order of 2-3 minutes.

Use the apparent altitude to decide whether the star is visible now, with `1010 hPa / 10 C` as the usual default weather values. `height` is **ellipsoidal height** in metres; when you only have orthometric height, add the geoid undulation first, see [Observer height conventions](coord.md#observer-height) in the README.

### Using a J2000 position at a given instant

```go
ra, dec := sirius.RaDecByDate(date)                        // J2000 -> instant (proper motion + precession + nutation)
fmt.Println(star.ApparentSiderealTime(date))               // apparent sidereal time, hours
fmt.Println(star.HourAngle(date, ra, lon))                 // hour angle, degrees
```

The catalog stores J2000 positions, so skipping `RaDecByDate` puts rise/set and constellation lookups off by degrees. Sidereal time, horizontal transforms, precession and nutation are covered in [Coordinate tools](coord.md); every time argument is a civil instant (a UTC label, equal to UT1 before 1972-01-01), see [Time Scale Declaration](map-geojson.md#time-scale-declaration).

### Constellation lookup and the bright-star catalog

```go
bright, _ := star.TopBrightStars()
for _, s := range bright[:3] {
	ra, dec := s.RaDecByDate(date)
	fmt.Println(s.ChineseName, star.ConstellationCode(ra, dec, date), s.Mag, s.CommonName)
}
```

```text
天狼 CMA -1.46 Sirius
老人 CAR -0.72 Canopus
大角 BOO -0.04 Arcturus
```

All three constellation entry points share one boundary table and differ only in output: `Constellation` gives the Chinese name, `ConstellationEN` the English name and `ConstellationCode` the IAU three-letter code; the lookup uses the apparent position of the day. Catalog fields (`HR`/`HD`/`HIP`, J2000 RA/Dec, magnitude, proper motion, parallax distance and the naming fields) are listed as a table under [API Reference](#star-catalog).

### Polar edges and catalog conventions

```go
_, err := star.RiseTime(date, ra, dec, 0, 89, 0, true)
if errors.Is(err, star.ERR_STAR_NEVER_RISE) || errors.Is(err, star.ERR_STAR_NEVER_SET) {
	fmt.Println("该日无升落：", err)
}
```

- **Polar sentinel errors**: `RiseTime` returns `star.ERR_STAR_NEVER_RISE` during a polar night (below the horizon all day) and `SetTime` returns `star.ERR_STAR_NEVER_SET` during a midnight sun (above it all day); branch with `errors.Is`.

  `ERR_STAR_NEVER_DOWN` is a deprecated alias of `ERR_STAR_NEVER_SET`, and `DownTime` is a deprecated alias of `SetTime`.
- **Refraction range**: the Saemundsson approximation only applies for true altitudes inside `(-5, 90)` degrees and contributes zero outside, so well below the horizon the apparent altitude equals the geometric one.
- **Catalog conventions**: BSC / HR `1-9110`, apparent magnitudes `-1.46` to `7.96`; name lookup matches the built-in Chinese names only (`天狼`, `织女一`), never English names.
- **Cross-checking against external catalogs**: positions here are J2000 mean places with proper motion plus precession and nutation; when comparing term by term, bring both sides to the same instant with `RaDecByDate` first.

## Parameter and result conventions

- **Units**: right ascension, declination, altitude, zenith distance, azimuth, hour angle, parallactic angle are in **degrees**; sidereal time is in **hours**; `tools.Format` renders them as degrees-minutes-seconds or hours-minutes-seconds.

  In the catalog `PmRA`/`PmDec` are arcseconds per year, `RadVel`/`RotVel` are km/s and `Pc` is parsecs.
- **Time scale**: every public time argument and return value is a civil instant (a UTC label, equal to UT1 before 1972-01-01); see [Time Scale Declaration](map-geojson.md#time-scale-declaration).
- **Coordinate convention**: the catalog is J2000; `RaDecByDate` adds proper motion, precession and nutation. Skipping it puts rise/set and constellation lookups off by a large margin (a century of precession is a degree-level shift).
- **Observer height**: `height` is ellipsoidal height in metres; when you only have orthometric height, add the geoid undulation first, as described under "Observer Height Convention" in the README.
- **Refraction range**: the Saemundsson approximation behind `ApparentAltitude`/`ApparentZenith` only applies for true altitudes inside `(-5, 90)` degrees and contributes zero outside that range, so well below the horizon the apparent altitude equals the geometric one (the `alt=-30.180` above is the same value `Altitude` returns).

  Pressure must be positive and temperature above absolute zero, otherwise the result is `NaN`.
- **`aero` semantics**: a true value solves rise/set against the horizon corrected by standard refraction, a false value uses the geometric horizon; the difference is about 2-3 minutes at low latitudes and grows noticeably at high latitudes.
- **Catalog range**: HR `1-9110`, apparent magnitudes `-1.46` to `7.96`; name lookup matches the built-in Chinese names only.

## Related manuals

- Sidereal time, horizontal transforms, precession/nutation and parallactic angle: [Coordinate tools](coord.md)
- Rise/set semantics next to the Sun and Moon: [Sun and Moon](sun-moon.md)
- Time scale in figures: [Time Scale Declaration](map-geojson.md#time-scale-declaration)
