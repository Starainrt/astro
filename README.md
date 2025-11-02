# Astro

**English | [中文](README.zh.md)**

[![Go Reference](https://pkg.go.dev/badge/github.com/starainrt/astro.svg)](https://pkg.go.dev/github.com/starainrt/astro)

A personal astronomy calculation library developed over many years for hobbyist astronomical calendar applications.

>  📚 This project is primarily for astronomical algorithm learning and verification. Calculation results meet amateur-level requirements.

Implementation based on "Astronomical Algorithms" by Jean Meeus. Provides calendar conversion, planetary positions, moon phases, sunrise/sunset, moonrise/moonset, and other astronomical calculations. Includes VSOP87 planetary algorithms and ELP2000/82 lunar algorithms.

Unless otherwise specified, coordinates provided by this program are instantaneous celestial coordinates.


## Contents

- [Installation](#installation)
- [Feature Overview](#feature-overview)
- [Quick Start](#quick-start)
   - [Calendar Conversion & Solar Terms](#calendar-conversion--solar-terms)
   - [Sun & Moon](#sun--moon)
   - [Planets](#planets)
   - [Stars](#stars)
- [Implemented Features](#implemented-features)
- [TODO](#todo)

## Installation

```bash
go get github.com/starainrt/astro
```

## Feature Overview

- 📅 **Calendar Conversion**: Gregorian ↔︎ Chinese lunisolar calendar conversion (104 BCE - 3000 CE or beyond), solar terms
- 🌞 **Sun Calculations**: Celestial position, sunrise/sunset, Earth-Sun distance, true solar time
-  🌙 **Moon Calculations**: Celestial position, moonrise/moonset, Earth-Moon distance, moon phases, new/full moon times
-  🪐 **Planet Calculations**: Celestial positions of seven planets, rise/set times, special phenomena (conjunction/opposition/station)
-  ⭐ **Star Calculations**: Constellation identification from coordinates; includes 9,100-star database with rise/set times and coordinate data

## Quick Start

### Calendar Conversion & Solar Terms

Supports conversion between Gregorian and Chinese lunar dates for years 104 BCE to 3000 CE (i.e., [-103, 3000]).

#### Data Sources & Verification

- **[-103, 1912]**: Based on "寿星天文历" data, corrected using [Prof. ytliu0's historical tables](https://ytliu0.github.io/ChineseCalendar/index_simp.html) 
- **[1913, 3000]**: Calculated per GB/T 33661-2017 standard using VSOP87 for solar terms and ELP2000 for new moons

---

#### Important Notes

##### 1. Multiple Chinese Lunar Dates per Gregorian Date
During periods with concurrent regimes (e.g., Three Kingdoms), different calendars may assign multiple lunar dates to one Gregorian date. This program provides all possible conversions.

##### 2. Multiple Gregorian Dates per Chinese Lunar Date
Occurs due to calendar reforms or concurrent regimes (e.g., two 腊月 months in Empress Wu's calendar reform).

##### 3. Gregorian Calendar Handling Rules
Based on Julian Day calculations:
- After Oct 15, 1582: Gregorian calendar
- Before Oct 4, 1582: Julian calendar
- Before 8 CE: Proleptic Julian calendar
- Day after Oct 4, 1582: Oct 15, 1582
- Year notation: 0 = 1 BCE, -1 = 2 BCE, etc.

##### 4. Go Standard Library Compatibility

⚠️ **`time.Time` handling differs before 1582:**

- Go uses proleptic Gregorian calendar before Oct 15, 1582 (not Julian). This doesn't affect basic usage without `Add()` methods.
- **Before 1582, `time.Time.Weekday()` may differ from this library**.  
  Example: Oct 4, 1582 is Thursday here, but Monday in Go.

#### Recommended Solution:
Use this method for weekday consistency:
```go
// date should be at 00:00
weekday := int(calendar.Date2JDE(date)+1.5) % 7
// 0=Sunday, 1=Monday, ..., 6=Saturday
```
Caution: Go's `Add`/`AddDate` may be inaccurate before 1582 (e.g., 700 CE is leap in Julian but not in Go's calendar).

#### Calendar Conversion

##### Gregorian to Lunar
- **Input**: Gregorian date (`time.Time`)
- **Output**: `calendar.Time` object (may contain multiple lunar dates)
- **Access**: Lunar details, heavenly stems/earthly branches, dynasty/emperor/era info, structured lunar data

##### Lunar to Gregorian
Two methods:

###### Method 1: Lunar string
Supported formats:
1. `Era + Year + Month + Day`: e.g., **`"元丰六年十月十二"`** (add "闰" for leap months, use "初一", "二十" for days)
2. `Era + Year + Month + Stem-Branch Day`: e.g., **`"元嘉二十七年七月庚午"`**
3. `Year + Month + Day`: e.g., **`"二零二五年正月初一"`** (add "闰" for leap months, modern dates)
4. `Year + Month + Stem-Branch Day`: e.g., **`"二零二五年正月戊戌日"`**
5. `Arabic numerals`: e.g., **`"2025年1月1日"`** = `二零二五年正月初一`
6. **Historical note**: Month names vary by era (e.g., "正月" vs "一月" in Wu Zetian's reign). Use Chinese numerals for accuracy.

> ️ **Note**: Lunar years don't perfectly align with Gregorian years.  
> Example: Gregorian Jan 28, 2025 (New Year's Eve) is lunar `"二零二四年腊月廿九"`.

###### Method 2: Numeric parameters
- **Params**: Year (`int`), Month (`int`), Day (`int`), IsLeap (`bool`)
- Best for modern lunar dates

##### Code Examples

```go
package main

import (
   "encoding/json"
   "fmt"
   "github.com/starainrt/astro/calendar"
   "time"
)

func main() {
   cst := time.FixedZone("CST", 8*3600)

   // Example 1: Gregorian to Lunar
   date := time.Date(240, 1, 1, 8, 8, 8, 8, cst)
   lunar, _ := calendar.SolarToLunar(date)
   fmt.Println(lunar.LunarDescWithEmperor()) // Lunar description

   info := lunar.LunarInfo()
   data, _ := json.MarshalIndent(info, "", "  ")
   fmt.Println(string(data)) // Structured lunar info

   // Example 2: Lunar to Gregorian (string)
   solar, _ := calendar.LunarToSolar("元丰六年十月十二日")
   for _, v := range solar {
      fmt.Println(v.Time())
      fmt.Println(v.LunarDescWithEmperor())
   }

   // Example 3: Lunar to Gregorian (numeric)
   modernDate, _ := calendar.LunarToSolarSingle(2025, 1, 1, false)
   fmt.Println(modernDate.Time())
}
```

Sample output:
```aiignore
// Three Kingdoms: One Gregorian date, three lunar dates
[Wei Mingdi 景初三年腊月二十 Shu Houzhu 延熙二年冬月十九 Wu Dadi 赤乌二年冬月二十]

// Structured lunar info (abbreviated)
[
  {
    "solarDate": "0240-01-01T08:08:08.000000008+08:00",
    "lunarYear": 239,
    "lunarYearChn": "二三九",
    "lunarMonth": 12,
    "lunarDay": 20,
    "isLeap": false,
    "lunarMonthDayDesc": "腊月二十",
    "ganzhiYear": "己未",
    "ganzhiMonth": "丙子",
    "ganzhiDay": "辛未",
    "dynasty": "魏",
    "emperor": "魏明帝",
    "nianhao": "景初",
    "yearOfNianhao": 3,
    "eraDesc": "景初三年",
    "lunarWithNianhaoDesc": "景初三年腊月二十",
    "chineseZodiac": "羊"
  },
    // ... additional fields ...
]

// Su Shi's poem date conversion
1083-11-24 00:00:00 +0800 CST
[Song Shenzong 元丰六年十月十二 Liao Daozong 大康九年十月十二]

// Modern Chinese lunar conversion
2025-01-29 00:00:00 +0800 CST
```

#### Solar Terms

```go
package main

import (
	"fmt"
	"github.com/starainrt/astro/calendar"
)

func main() {
  // 2020 Start of Spring
	fmt.Println(calendar.JieQi(2020, calendar.JQ_立春))
  // 2020 Winter Solstice
	fmt.Println(calendar.JieQi(2020, calendar.JQ_冬至))
  // 2020 Vernal Equinox
	fmt.Println(calendar.JieQi(2020, calendar.JQ_春分))
  // Using ecliptic longitude (0° = Vernal Equinox)
	fmt.Println(calendar.JieQi(2020, 0))
}
```

Output:
```
2020-02-04 17:03:17.820854187 +0800 CST // Start of Spring
2020-12-21 18:02:17.568823993 +0800 CST // Winter Solstice
2020-03-20 11:49:34.502393603 +0800 CST // Vernal Equinox
2020-03-20 11:49:34.502393603 +0800 CST // Vernal Equinox (alt method)
```


### Sun & Moon

#### Rise/Set Times

> ️ **Important**:  
> Lunar rise/set times are calculated per day and may not be continuous.
>
> **Possible scenarios**:
> - Moon may set at 1 AM and rise at 12 PM (rise after set). For evening moonset, calculate using next day's date.
> 
> **For full cycles**: Check if rise time is after set time to determine correct subsequent events.

```go
package main

import (
	"fmt"
	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/sun"
	"time"
)

func main() {
	// Xi'an, Shaanxi parameters
	lon, lat, height := 108.93, 34.27, 0.0
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	
	// Civil dawn (-6°)
	fmt.Println(sun.MorningTwilight(date, lon, lat, -6))
	// Sunrise (with atmospheric refraction)
	fmt.Println(sun.RiseTime(date, lon, lat, height, true))
	// Solar noon
	fmt.Println(sun.CulminationTime(date, lon))
	// Sunset
	fmt.Println(sun.SetTime(date, lon, lat, height, true))
	// Civil dusk
	fmt.Println(sun.EveningTwilight(date, lon, lat, -6))

	// Moonrise
	fmt.Println(moon.RiseTime(date, lon, lat, height, true))
	// Lunar transit
	fmt.Println(moon.CulminationTime(date, lon, lat))
	// Moonset
	fmt.Println(moon.SetTime(date, lon, lat, height, true))
}
```

Output:
```
2020-01-01 07:22:27.964431345 +0800 CST <nil> // Civil dawn
2020-01-01 07:50:14.534510672 +0800 CST <nil> // Sunrise
2020-01-01 12:47:35.933117866 +0800 CST       // Solar noon
2020-01-01 17:44:47.076647579 +0800 CST <nil> // Sunset
2020-01-01 18:12:33.629668056 +0800 CST <nil> // Civil dusk
2020-01-01 11:52:44.643359184 +0800 CST <nil> // Moonrise
2020-01-01 17:38:03.879639208 +0800 CST       // Lunar transit
2020-01-01 23:26:52.202896177 +0800 CST <nil> // Moonset
```

#### Positions

```go
package main

import (
	"fmt"
	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/star"
	"github.com/starainrt/astro/sun"
	"github.com/starainrt/astro/tools"
	"time"
)

func main() {
	lon, lat := 108.93, 34.27
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	
	// Sun ecliptic longitude
	fmt.Println(sun.ApparentLo(date))
	// Obliquity of ecliptic
	fmt.Println(sun.EclipticObliquity(date, true))
	// Sun apparent RA/Dec
	ra, dec := sun.ApparentRaDec(date)
	fmt.Println("RA:", tools.Format(ra/15, 1), "Dec:", tools.Format(dec, 0))
	// Sun's constellation
	fmt.Println(star.Constellation(ra, dec, date))
	// Sun az/el in Xi'an
	fmt.Println("Azimuth:", sun.Azimuth(date, lon, lat), "Elevation:", sun.Zenith(date, lon, lat))
    // Earth-Sun distance (AU)
	fmt.Println(sun.EarthDistance(date))

	// Moon apparent RA/Dec (topocentric)
	ra, dec = moon.ApparentRaDec(date, lon, lat)
	fmt.Println("RA:", tools.Format(ra/15, 1), "Dec:", tools.Format(dec, 0))
	// Moon's constellation
	fmt.Println(star.Constellation(ra, dec, date))
	// Moon az/el in Xi'an
	fmt.Println("Azimuth:", moon.Azimuth(date, lon, lat), "Elevation:", moon.Zenith(date, lon, lat))
    // Earth-Moon distance (km)
	fmt.Println(moon.EarthDistance(date))
}
```

Output:
```
280.0152925179703    // Ecliptic longitude
23.436215552851408   // Obliquity
RA: 18h43m34.83s Dec: -23°3′30.25″
人马座         // Constellation
Azimuth: 120.19483856399326 Elevation: 2.4014324584398516
0.9832929365443133   // Distance (AU)

RA: 23h17m51.93s Dec: -10°19′17.02″
宝瓶座             // Constellation
Azimuth: 67.84449893794012 Elevation: -45.13018696439911
404238.6354387698    // Distance (km)
```

#### Moon Phases

```go
package main

import (
	"fmt"
	"github.com/starainrt/astro/moon"
	"time"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	
	// Illuminated fraction
	fmt.Println(moon.Phase(date))
	// Phase description
	fmt.Println(moon.PhaseDesc(date))
	// Next new moon
	fmt.Println(moon.NextShuoYue(date))
	// Next first quarter
	fmt.Println(moon.NextShangXianYue(date))
	// Next full moon
	fmt.Println(moon.NextWangYue(date))
	// Next last quarter
	fmt.Println(moon.NextXiaXianYue(date))
}
```

Output:
```
0.3000437415436273 // 30% illuminated
上峨眉月    // Phase description
2020-01-25 05:41:55.820311009 +0800 CST // New moon
2020-01-03 12:45:20.809730887 +0800 CST // First quarter
2020-01-11 03:21:14.729664623 +0800 CST // Full moon
2020-01-17 20:58:20.955985486 +0800 CST // Last quarter
```

### Planets

#### Inner Planets

```go
package main

import (
	"fmt"
	"github.com/starainrt/astro/mercury"
	"github.com/starainrt/astro/venus"
	"time"
)

func main() {
	lon, lat, height := 108.93, 34.27, 0.0
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	
	// Mercury's last inferior conjunction
	fmt.Println(mercury.LastInferiorConjunction(date))
	// Venus' next superior conjunction
	fmt.Println(venus.NextSuperiorConjunction(date))
	// Mercury's last station (direct→retrograde)
	fmt.Println(mercury.LastProgradeToRetrograde(date))
	// Venus' next station (retrograde→direct)
	fmt.Println(venus.NextRetrogradeToPrograde(date))
	// Mercury's last greatest eastern elongation
	fmt.Println(mercury.LastGreatestElongationEast(date))
	// Venus' next greatest western elongation
	fmt.Println(venus.NextGreatestElongationWest(date))
	// Venus rise/set in Xi'an
	fmt.Println(venus.RiseTime(date, lon, lat, height, true))
	fmt.Println(venus.SetTime(date, lon, lat, height, true))
	// Venus apparent magnitude
	fmt.Println(venus.ApparentMagnitude(date))
	// Venus-Earth distance (AU)
	fmt.Println(venus.EarthDistance(date))
	// Venus-Sun distance (AU)
	fmt.Println(venus.SunDistance(date))
}
```

Output:
```
2019-11-11 23:21:39.702344834 +0800 CST // Inf. conjunction
2021-03-26 14:57:38.289429545 +0800 CST // Sup. conjunction
2019-11-01 04:31:47.807287573 +0800 CST // Station (D→R)
2021-12-18 18:59:12.762369811 +0800 CST // Venus next station (retrograde to prograde)
2019-10-20 11:59:33.893027007 +0800 CST // Mercury last greatest eastern elongation
2020-08-13 07:56:02.326616048 +0800 CST // Venus next greatest western elongation
2020-01-01 10:01:10.821288228 +0800 CST <nil> // Venus rise time in Xi'an today
2020-01-01 20:27:00.741534233 +0800 CST <nil> // Venus set time in Xi'an today
-4 // Venus apparent magnitude
1.2760033106813273 // Venus-Earth distance (AU)
0.7262288470390035 // Venus-Sun distance (AU)
```


#### Outer Planets

```golang
package main

import (
	"fmt"
	"github.com/starainrt/astro/jupiter"
	"github.com/starainrt/astro/mars"
	"github.com/starainrt/astro/neptune"
	"github.com/starainrt/astro/saturn"
	"github.com/starainrt/astro/uranus"
	"time"
)

func main() {
	// Xi'an, Shaanxi parameters
	lon, lat, height := 108.93, 34.27, 0.0
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	
	// Mars next opposition
	fmt.Println(mars.NextOpposition(date))
	// Jupiter next conjunction
	fmt.Println(jupiter.NextConjunction(date))
	// Saturn's last station (direct→retrograde)
	fmt.Println(saturn.LastProgradeToRetrograde(date))
	// Uranus next station (retrograde→direct)
	fmt.Println(uranus.NextRetrogradeToPrograde(date))
	// Neptune's last eastern quadrature
	fmt.Println(neptune.LastEasternQuadrature(date))
	// Mars next western quadrature
	fmt.Println(mars.NextWesternQuadrature(date))
	// Mars rise/set in Xi'an
	fmt.Println(mars.RiseTime(date, lon, lat, height, true))
	fmt.Println(mars.SetTime(date, lon, lat, height, true))
	// Mars apparent magnitude
	fmt.Println(mars.ApparentMagnitude(date))
	// Earth-Mars distance (AU)
	fmt.Println(mars.EarthDistance(date))
	// Sun-Mars distance (AU)
	fmt.Println(mars.SunDistance(date))
}
```

Output:
```
2020-10-14 07:25:47.740884125 +0800 CST // Mars opposition
2021-01-29 09:39:30.916356146 +0800 CST // Jupiter conjunction
2019-04-30 10:28:27.453395426 +0800 CST // Saturn station (D→R)
2021-01-14 21:35:01.269377768 +0800 CST // Uranus station (R→D)
2019-12-08 17:00:13.772284984 +0800 CST // Neptune eastern quadrature
2020-06-07 03:10:57.179121673 +0800 CST // Mars western quadrature
2020-01-01 04:40:05.409269034 +0800 CST <nil> // Mars rise time
2020-01-01 14:56:57.175483703 +0800 CST <nil> // Mars set time
1.57 // Apparent magnitude
2.1820316323604088 // Earth-Mars distance (AU)
1.5894169865107062 // Sun-Mars distance (AU)
```

### Stars

1. Includes database of 9,100 stars with proper motion calculations

```go
package main

import (
	"fmt"
	"github.com/starainrt/astro/star"
	"time"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)

	// Initialize star database
	star.InitStarDatabase()
	sirius, _ := star.StarDataByName("天狼") // Sirius
	// Sirius rise time
	riseDate, _ := star.RiseTime(date, sirius.Ra, sirius.Dec, 115, 40, 0, true)
	fmt.Println(riseDate)
	// Sirius set time
	setDate, _ := star.SetTime(date, sirius.Ra, sirius.Dec, 115, 40, 0, true)
	fmt.Println(setDate)
}
```

Output:
```
2019-12-31 19:21:56.993647813 +0800 CST // Sirius rise time
2020-01-01 05:29:53.535125255 +0800 CST  // Sirius set time
```

## Implemented Features

- ✅ Sun position, elevation, azimuth, transit, twilight, rise/set, solar terms
- ✅ Moon position, elevation, azimuth, transit, rise/set, phases
- ✅ Earth eccentricity, Earth-Sun distance
- ✅ True/apparent sidereal time, constellation identification
- ✅ Positions of seven planets, Sun/Earth distances, special phenomena
- ✅ Gregorian/Lunar conversion (104 BCE - 3000 CE)
- ✅ 9,100+ star database with proper motion

## TODO

-  🔄 Code standardization and performance optimization
- 🔄 Enhanced star calculation features
- 🔄 Solar/lunar eclipse calculations
-  🔄 More astronomical phenomena calculations
```
