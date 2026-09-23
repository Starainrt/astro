# Calendar and Solar Terms

[中文](../calendar.md) | [Back to README](../../../README.en.md)

The `calendar` package converts between Gregorian dates and the traditional Chinese lunisolar calendar, and exposes solar terms. The supported range is from 721 BCE through 3000 CE; 104 BCE is the calendar-table switch point, and beyond 3000 CE the result is pure theoretical extrapolation computed on the fly.

The calendar is lunisolar in the strict sense, but public function names use `Lunar` rather than the more academic `Lunisolar`, for readability and convention.

For historical input, Chinese era names stay in Chinese. This is part of the API surface, because historical Chinese dates are normally written that way.

## Contents

- [Converting a date and calculating a solar term](#converting-a-date-and-calculating-a-solar-term)
- [API Reference](#api-reference)
  - [Gregorian and Lunisolar Conversion](#gregorian-and-lunisolar-conversion)
  - [Julian-only Leap Days and Multiple Candidates](#julian-only-leap-days-and-multiple-candidates)
  - [Solar Terms and Pentads](#solar-terms-and-pentads)
  - [Ancient Calendar Systems](#ancient-calendar-systems)
  - [Ganzhi and Era Names](#ganzhi-and-era-names)
  - [Structure and JSON](#structure-and-json)
- [Usage examples](#usage-examples)
  - [Gregorian to lunisolar and the sexagenary day](#gregorian-to-lunisolar-and-the-sexagenary-day)
  - [Solar-term instants and calendrical dates](#solar-term-instants-and-calendrical-dates)
  - [Custom time zones and Julian-only leap days](#custom-time-zones-and-julian-only-leap-days)
  - [Several candidates for a dual-numbering reform](#several-candidates-for-a-dual-numbering-reform)
  - [Ancient calendars and era-name conventions](#ancient-calendars-and-era-name-conventions)
- [Calendar notes](#calendar-notes)
- [Usage notes](#usage-notes)
  - [1. One Gregorian date may map to several lunisolar dates](#1-one-gregorian-date-may-map-to-several-lunisolar-dates)
  - [2. One lunisolar date may map to several Gregorian dates](#2-one-lunisolar-date-may-map-to-several-gregorian-dates)
  - [3. Gregorian calendar rules](#3-gregorian-calendar-rules)
  - [4. Time zone](#4-time-zone)
  - [5. Go-specific note](#5-go-specific-note)
  - [6. Julian-only leap days (for example 700-02-29)](#6-julian-only-leap-days-for-example-700-02-29)
  - [7. Several Gregorian candidates for one lunar date](#7-several-gregorian-candidates-for-one-lunar-date)
- [Calendar conversion](#calendar-conversion)
  - [Gregorian to lunar](#gregorian-to-lunar)
  - [Lunar to Gregorian](#lunar-to-gregorian)
  - [Code example](#code-example)
- [Solar terms](#solar-terms)
- [Parameter and result conventions](#parameter-and-result-conventions)
  - [Units and Parameter Conventions](#units-and-parameter-conventions)
  - [Time Scale](#time-scale)
  - [Zero Values and Out-of-Range Input](#zero-values-and-out-of-range-input)
  - [Accuracy and Applicability](#accuracy-and-applicability)
  - [Related Manuals](#related-manuals)

## Converting a date and calculating a solar term

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro/calendar"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2026, 2, 17, 0, 0, 0, 0, cst)
	solar, err := calendar.SolarToLunar(date) // Gregorian to lunisolar
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(solar.Lunar().LunarYear(), solar.Lunar().MonthDay())
	fmt.Println(calendar.JieQi(2026, calendar.JQ_立春)) // solar-term instant in UTC+08:00
}
```

Handle conversion errors before reading the result. `Time.Lunar()` returns the first lunar-calendar candidate; `Time.Lunars()` returns all candidates for periods with concurrent calendars. Solar terms use Beijing time.

## API Reference

### Gregorian and Lunisolar Conversion

| Name | Purpose | Units and convention |
| --- | --- | --- |
| `SolarToLunar` | Gregorian `time.Time` to lunisolar | the input is moved to UTC+8 before the civil day is taken; returns a `Time` that may hold several parallel-regime candidates |
| `SolarToLunarByYMD` | integer Gregorian year/month/day to lunisolar | never goes through `time.Time`, so a Julian-only leap day survives; negative years are BCE |
| `LunarToSolar` | lunisolar description string to Gregorian candidates | accepts `year+month+day`, `era name+month+day`, and the `... + ganzhi day` forms; returns `[]Time` |
| `LunarToSolarByYMD` | lunisolar year/month/day plus leap flag to Gregorian | `month` counts the distance from the first month, which is 1; `leap` is the leap-month flag |
| `LunarToSolarSingle` | the same, returning one `Time` | deprecated, identical to `LunarToSolarByYMD` |
| `Solar` | lunisolar year/month/day to Gregorian under a custom zone | `timezone` is in hours; the returned zone is always named `CST` with the offset of the supplied `timezone` |
| `Lunar` | Gregorian year/month/day to lunisolar under a custom zone | returns `(lunisolar year, month, day, leap flag, text description)`; `timezone` is in hours |
| `Time` / `LunarTime` | Gregorian-plus-lunisolar carriers | fields and JSON keys are listed under "Structure and JSON" |

```go
// Prelude: date is a time.Time; the string entry point may return several parallel-regime results.
solar, _ := calendar.SolarToLunar(date)
fmt.Println(solar.LunarDesc(), solar.Lunar().MonthDay())

lunar, _ := calendar.LunarToSolar("元丰六年十月十二日")
for _, v := range lunar {
	fmt.Println(v.Time().Format("2006-01-02"), v.Eras())
}

old, _ := calendar.LunarToSolarByYMD(-202, 1, 1, false) // negative years are BCE
fmt.Println(old.Solar().Format("2006-01-02"))
```

### Julian-only Leap Days and Multiple Candidates

| Name | Purpose | Units and convention |
| --- | --- | --- |
| `Time.JulianOnly` | whether the primary lunisolar day exists only in the Julian calendar | boolean; before 1582 this covers 29 February of years divisible by 100 but not 400 |
| `Time.JD` | exact Julian day of the primary lunisolar day | for a Julian-only leap day it is one day earlier than `Solar()`, otherwise equal to `Date2JD(Solar())` |
| `Time.SolarCandidates` | every legal Gregorian candidate of that lunisolar day | the first element always equals `Solar()`; with a single candidate the slice holds only `Solar()` |
| `LunarTime.JulianOnly` / `LunarTime.JD` | the same information on one candidate | same convention as the methods on `Time` |
| `Date2JD` | `time.Time` to Julian day | reads the year/month/day/hour/minute/second fields directly, with no time-zone conversion |
| `JD2Date` | Julian day to `time.Time` | the result lands in the local zone of the host (the underlying helper uses `Local`) |
| `NowJD` | current Julian day | same convention as `Date2JD` |
| `Time.Add` | time offset | crossing the 1582 reform gap moves along the Julian-day axis and skips the ten missing days |

```go
// The integer entry point keeps the leap day; 700-02-29 does not exist in Go's time.Time.
julian, _ := calendar.SolarToLunarByYMD(700, 2, 29)
fmt.Println(julian.JulianOnly(), julian.JD(), julian.Solar().Format("2006-01-02"))

// A dual-numbering reform yields several legal Gregorian candidates, Solar() still first.
res, _ := calendar.LunarToSolarByYMD(700, 11, 1, false)
fmt.Println(res.Solar().Format("2006-01-02"), len(res.SolarCandidates()))

fmt.Println(calendar.Date2JD(time.Date(2026, 2, 17, 0, 0, 0, 0, time.UTC)), calendar.JD2Date(2461088.5))
```

### Solar Terms and Pentads

| Name | Purpose | Units and convention |
| --- | --- | --- |
| `JieQi` | solar-term instant | Beijing time; `term` is the apparent solar longitude in degrees |
| `CalendricalJieQi` | calendar-compatible solar-term date | the date the term falls on under the default calendar, fixed at 00:00 Beijing time |
| `CalendricalJieQiWithCalendar` | calendar-compatible date with an explicit ancient calendar | the Chunqiu calendar and years without calendrical term data return an error |
| `WuHou` | pentad instant | Beijing time; the 72 pentads step by 5°, and the argument matches `JieQi` |
| `JQ_春分` … `JQ_惊蛰` | solar-term constants | apparent solar longitude on a 15° grid, usable directly as the `term` of `JieQi`/`WuHou` |

There are 24 constants, ordered by increasing longitude:

| Constant | Longitude | Constant | Longitude | Constant | Longitude |
| --- | --- | --- | --- | --- | --- |
| `JQ_春分` | 0° | `JQ_清明` | 15° | `JQ_谷雨` | 30° |
| `JQ_立夏` | 45° | `JQ_小满` | 60° | `JQ_芒种` | 75° |
| `JQ_夏至` | 90° | `JQ_小暑` | 105° | `JQ_大暑` | 120° |
| `JQ_立秋` | 135° | `JQ_处暑` | 150° | `JQ_白露` | 165° |
| `JQ_秋分` | 180° | `JQ_寒露` | 195° | `JQ_霜降` | 210° |
| `JQ_立冬` | 225° | `JQ_小雪` | 240° | `JQ_大雪` | 255° |
| `JQ_冬至` | 270° | `JQ_小寒` | 285° | `JQ_大寒` | 300° |
| `JQ_立春` | 315° | `JQ_雨水` | 330° | `JQ_惊蛰` | 345° |

```go
// term is the apparent solar longitude in degrees; a raw number works too, 0 being the March Equinox.
fmt.Println(calendar.JieQi(2020, calendar.JQ_立春))
fmt.Println(calendar.JieQi(2020, calendar.JQ_冬至))
fmt.Println(calendar.WuHou(2020, calendar.JQ_立春+5))

termDate, err := calendar.CalendricalJieQi(1582, calendar.JQ_冬至)
fmt.Println(termDate, err)
```

### Ancient Calendar Systems

| Name | Purpose | Units and convention |
| --- | --- | --- |
| `AncientCalendarSystem` | ancient-calendar enumerator | a string type |
| `AncientCalendarDefault` | default routing (zero value) | selects by year; the explicit-calendar entry points fall back to the default implementation |
| `AncientCalendarChunqiu` / `AncientCalendarZhou` / `AncientCalendarLu` | pre-Qin Chunqiu, Zhou, and Lu calendars | values `chunqiu`/`zhou`/`lu` |
| `AncientCalendarHuangdi` / `AncientCalendarYin` | pre-Qin Huangdi and Yin calendars | values `huangdi`/`yin` |
| `AncientCalendarXia1` / `AncientCalendarXia2` | Xia calendar, winter-solstice and rain-water editions | values `xia1`/`xia2` |
| `AncientCalendarZhuanxu` | pre-Qin Zhuanxu calendar | value `zhuanxu` |
| `AncientCalendarQinHan` | Qin/Han Zhuanxu calendar | value `qin_han` |
| `SolarToLunarWithCalendar` | Gregorian `time.Time` to lunisolar with an explicit calendar | the input is moved to UTC+8 before the civil day is taken |
| `SolarToLunarByYMDWithCalendar` | integer Gregorian year/month/day to lunisolar with an explicit calendar | never goes through `time.Time` |
| `LunarToSolarWithCalendar` | lunisolar description to Gregorian with an explicit calendar | an era-name description works only when the ancient era table matches, otherwise it errors |
| `LunarToSolarByYMDWithCalendar` | lunisolar year/month/day plus leap flag with an explicit calendar | years without ancient-calendar data return an error |

```go
// The zero value AncientCalendarDefault routes by year; an explicit calendar needs years that have data.
zhuanxu, err := calendar.SolarToLunarByYMDWithCalendar(-300, 1, 5, calendar.AncientCalendarZhuanxu)
fmt.Println(zhuanxu.Lunar().CalendarSystem(), zhuanxu.Lunar().CalendarName(), err)

back, err := calendar.LunarToSolarByYMDWithCalendar(
	zhuanxu.Lunar().LunarYear(), zhuanxu.Lunar().LunarMonth(), zhuanxu.Lunar().LunarDay(),
	zhuanxu.Lunar().IsLeap(), calendar.AncientCalendarZhuanxu)
fmt.Println(back.Solar().Format("2006-01-02"), err)
```

### Ganzhi and Era Names

| Name | Purpose | Units and convention |
| --- | --- | --- |
| `GanZhiOfYear` | sexagenary year name | the argument is a Gregorian year; returns a two-character name such as `甲子` |
| `GanZhiOfDay` | sexagenary day name | reads the year/month/day fields of `t` and evaluates the day at 00:00 Beijing time |
| `LunarTime.GanZhiYear` / `GanZhiMonth` / `GanZhiDay` | sexagenary year, month, and day of one candidate | a leap month inherits the month name of the preceding month |
| `Era` | era-table record | fields `Year`/`Emperor`/`Nianhao`/`OtherNianHaoStart`/`Dynasty`/`Offset` |
| `EraDesc` | era description | fields `YearOfNianHao`/`Emperor`/`Nianhao`/`Dynasty` |
| `EraDesc.String` | era-name string | the first year reads `元年` and later years use Chinese numerals plus `年` |
| `Time.Eras` | era information of every candidate | returns `[]EraDesc` |
| `LunarTime.Eras` | era information of one candidate | returns `[]EraDesc` |
| `ERR_NIANHAO_NOT_FOUND` | sentinel for an unknown era name | returned when the text entry point cannot resolve the era name |

```go
fmt.Println(calendar.GanZhiOfYear(2026))
fmt.Println(calendar.GanZhiOfDay(time.Date(2026, 2, 17, 0, 0, 0, 0, time.UTC)))

res, _ := calendar.LunarToSolarByYMD(1083, 10, 12, false)
fmt.Println(res.Lunar().GanZhiYear(), res.Lunar().GanZhiMonth(), res.Lunar().GanZhiDay())
for _, era := range res.Eras() {
	fmt.Println(era.Dynasty, era.Emperor, era.String())
}
```

### Structure and JSON

| Name | Purpose | Units and convention |
| --- | --- | --- |
| `Time` | Gregorian-plus-lunisolar carrier | `Solar()`/`Time()` give the Gregorian side, `Lunar()`/`Lunars()` the lunisolar candidates, `LunarInfo()` the structured records |
| `LunarTime` | one lunisolar candidate | `LunarYear`/`LunarMonth`/`LunarDay`/`IsLeap` plus `MonthDay`, ganzhi, zodiac, and calendar name |
| `LunarInfo` | structured lunisolar record | fields and JSON keys are listed below |
| `Time.Solar` / `Time.Time` | the stored Gregorian `time.Time` | no further zone or calendar conversion; the two are synonyms |
| `Time.Lunar` / `Time.Lunars` | first / all lunisolar candidates | with no result `Lunar` returns a zero-value `LunarTime` |
| `Time.LunarInfo` / `LunarTime.LunarInfo` | structured record slice | several records when parallel era names exist |
| `Time.LunarDesc` / `Time.LunarDescWithEmperor` / `Time.LunarDescWithDynasty` / `Time.LunarDescWithDynastyAndEmperor` | description family over all candidates | all return `[]string` |
| `LunarTime.LunarDesc` / `LunarDescWithEmperor` / `LunarDescWithDynasty` / `LunarDescWithDynastyAndEmperor` | description family over one candidate | all return `[]string` |
| `LunarTime.LunarYear` / `LunarMonth` / `LunarDay` / `IsLeap` | lunisolar year, month, day, and leap flag | `LunarMonth` counts the distance from the first month |
| `LunarTime.MonthDay` | month-day description | month 11 is written `冬月` and month 12 `腊月` |
| `LunarTime.ShengXiao` / `LunarTime.Zodiac` | Chinese zodiac | synonyms, computed from the lunisolar year modulo 12 |
| `LunarTime.CalendarSystem` / `LunarTime.CalendarName` | the ancient calendar and its Chinese name | empty strings for default routing and the modern range |

`LunarInfo` JSON keys and fields:

| JSON key | Field | Convention |
| --- | --- | --- |
| `solarDate` | `SolarDate` | Gregorian instant |
| `lunarYear` / `lunarYearChn` | `LunarYear` / `LunarYearChn` | the civil-year mapping of the lunisolar year and its Chinese numeral form |
| `lunarMonth` / `lunarDay` | `LunarMonth` / `LunarDay` | the month counts the distance from the first month; the day is within `[1,30]` |
| `isLeap` | `IsLeap` | leap-month flag |
| `lunarMonthDayDesc` | `LunarMonthDayDesc` | month-day description, month 11 as `冬月` and month 12 as `腊月` |
| `ganzhiYear` / `ganzhiMonth` / `ganzhiDay` | `GanzhiYear` / `GanzhiMonth` / `GanzhiDay` | sexagenary year, month, and day |
| `calendarSystem` / `calendarName` | `CalendarSystem` / `CalendarName` | ancient calendar and its name; empty strings for default routing |
| `jd` | `JD` | exact Julian day of the lunisolar day |
| `julianOnly` | `JulianOnly` | carries `omitempty`; only a Julian-only leap day emits `true` |
| `dynasty` / `emperor` / `nianhao` / `yearOfNianhao` / `eraDesc` | `Dynasty` / `Emperor` / `Nianhao` / `YearOfNianhao` / `EraDesc` | dynasty, emperor, era name, and era regnal year |
| `lunarWithNianhaoDesc` | `LunarWithEraDesc` | note that the field name differs from the JSON key |
| `chineseZodiac` | `ChineseZodiac` | Chinese zodiac |

```go
// Prelude: import "encoding/json".
res, _ := calendar.SolarToLunarByYMD(2026, 2, 17)
for _, info := range res.LunarInfo() {
	fmt.Println(info.LunarYear, info.LunarMonthDayDesc, info.GanzhiDay, info.ChineseZodiac, info.JD)
}
data, _ := json.Marshal(res.LunarInfo()[0])
fmt.Println(string(data))
```

## Usage examples

### Gregorian to lunisolar and the sexagenary day

```go
solar, _ := calendar.SolarToLunar(date) // date = 2026-02-17 00:00:00 CST
lunar := solar.Lunar()
fmt.Println(lunar.MonthDay(), lunar.LunarYear(), lunar.GanZhiDay())
fmt.Println(calendar.GanZhiOfDay(date))
```

```text
正月初一 2026 壬戌
壬戌
```

`SolarToLunar` and `GanZhiOfDay` share one day-counting convention, so the last line equals `GanZhiDay()`.

Parallel historical regimes can give one Gregorian day several lunisolar candidates (`Lunar()` returns the first and `Lunars()` all of them), see [Usage notes](#usage-notes); a leap month is marked separately by `IsLeap()`, see [Parameter and result conventions](#parameter-and-result-conventions).

### Solar-term instants and calendrical dates

```go
fmt.Println(calendar.JieQi(2020, calendar.JQ_立春))                // modern astronomical instant
termDate, err := calendar.CalendricalJieQi(1582, calendar.JQ_冬至) // calendrical date
fmt.Println(termDate, err)
```

```text
2020-02-04 17:03:20.471614301 +0800 CST
1582-12-22 00:00:00 +0800 CST <nil>
```

`JieQi` returns the exact instant from modern astronomy while `CalendricalJieQi` returns the date the calendar assigns (always 00:00 Beijing time); the two are not interchangeable, and calendrical terms exist only for years that have table data (2026 returns an error). See [Solar terms](#solar-terms) and [Calendar notes](#calendar-notes).

### Custom time zones and Julian-only leap days

```go
fmt.Println(calendar.Solar(1985, 1, 1, false, 8.0), calendar.Solar(1985, 1, 1, false, 7.0))
julian, _ := calendar.SolarToLunarByYMD(700, 2, 29) // a Julian-only leap day
fmt.Println(julian.Solar().Format("2006-01-02"), julian.JulianOnly(), julian.JD())
```

```text
1985-02-20 00:00:00 +0800 CST 1985-01-21 00:00:00 +0700 CST
0700-03-01 true 1.9767915e+06
```

New moons and solar terms are computed in Beijing time by default, so under UTC+7 one lunisolar day maps to a different Gregorian date (here the first day of month 1 moves from 20 February to 21 January, because a shifted new moon or solstice reorders the months); the `timezone` of `Solar`/`Lunar` is in hours. 700-02-29 exists only in the Julian calendar: `Solar()` always reports the following day and `JD()` holds the exact date, see [Usage notes](#usage-notes).

### Several candidates for a dual-numbering reform

```go
res, _ := calendar.LunarToSolarByYMD(700, 11, 1, false)
for _, c := range res.SolarCandidates() {
	fmt.Println(c.Format("2006-01-02"))
}
```

```text
0700-12-15
0700-10-17
```

The dual-numbering reforms (Wang Mang, Emperor Ming of Wei, Wu Zetian, Emperor Suzong of Tang) and the Taichu handoff give one lunisolar day two legal Gregorian days; the first line is the default candidate `Solar()` and `SolarCandidates()` returns them all, while a single-element slice is returned outside those windows, see [Usage notes](#usage-notes).

### Ancient calendars and era-name conventions

```go
zhuanxu, err := calendar.SolarToLunarByYMDWithCalendar(-300, 1, 5, calendar.AncientCalendarZhuanxu)
fmt.Println(zhuanxu.Lunar().CalendarSystem(), zhuanxu.Lunar().CalendarName(), err)
era, _ := calendar.LunarToSolarByYMD(1083, 10, 12, false)
fmt.Println(era.LunarDescWithEmperor())
```

```text
zhuanxu 颛顼历 <nil>
[宋神宗 元丰六年十月十二 辽道宗 大康九年十月十二]
```

The explicit-calendar entry points work only for years that have data, and `AncientCalendarDefault` is the zero value that falls back to routing by year; `CalendarSystem()`/`CalendarName()` are empty under default routing and in the modern range, and era names come from `Eras()`/`LunarDescWithEmperor()`, see [Calendar notes](#calendar-notes) and [Ancient Calendar Systems](#ancient-calendar-systems).

## Calendar notes

- **Default routing**: the package selects by year automatically.

  The pre-Qin range uses reconstructed Chunqiu and ancient-six-calendar systems, `-220..-104` uses the Qin/Han Zhuanxu calendar, `-103..1912` uses calendar tables, and `1913` onward uses the modern algorithm.
- **Explicit ancient calendars**: use APIs such as `SolarToLunarWithCalendar` / `LunarToSolarWithCalendar` when a specific ancient calendar system is required.
- **Data sources**: ancient-calendar support mainly references 《寿星天文历》; [Professor ytliu0's ChineseCalendar data](https://ytliu0.github.io/ChineseCalendar/index_simp.html) is used for validation.

  The modern range follows GB/T 33661-2017 and uses VSOP87/ELP computations for solar terms and new moons.
- **Solar terms**: `JieQi` returns modern astronomical solar-term instants; `CalendricalJieQi` returns calendar-compatible solar-term dates.

## Usage notes

### 1. One Gregorian date may map to several lunisolar dates
In periods when several regimes coexisted with different calendars (the Three Kingdoms, for example), one Gregorian date can map to several lunisolar dates. The package returns every conversion it can determine.

### 2. One lunisolar date may map to several Gregorian dates
Rival calendars are not the only cause: one regime can produce the same effect during a calendar reform. After Wu Zetian's reform, for example, the third year of Shengli had two twelfth months.

### 3. Gregorian calendar rules
All computation goes through Julian Days. The Gregorian side follows these rules:
- after `1582-10-15`: Gregorian calendar
- before `1582-10-04`: Julian calendar
- before year 8 CE: proleptic Julian calendar
- the day after `1582-10-04` is `1582-10-15`
- `1582-10-05` through `1582-10-14` do not exist and the corresponding entry points reject them
- year numbering: year `0` is 1 BCE, year `-1` is 2 BCE, and so on

### 4. Time zone
The package targets the Chinese calendar, so solar terms and new moons are computed in Beijing time (UTC+8) by default. Applying Chinese lunisolar rules in another time zone can shift dates.

For exploration and research, the lower-level `Solar` and `Lunar` methods convert between the Gregorian and lunisolar calendars under a **custom time zone**, following the **current Chinese calendar algorithm (GB/T 33661-2017)**.
For standard conversions in Beijing time, use the wrapped `SolarToLunar` and `LunarToSolar` methods.

**Example**: the calendar rules require the winter solstice to fall in the eleventh lunisolar month. For the 1984 winter solstice:
```go
ws := calendar.JieQi(1984, 270)
fmt.Println(ws)
fmt.Println(moon.ClosestShuoYue(ws))
```

| Event | UTC+8 | UTC+7 |
| --- | --- | --- |
| Winter solstice | 1984-12-22 | 1984-12-21 |
| New moon | 1984-12-22 | 1984-12-22 |

In UTC+8 (China), 1984-12-22 is both the winter solstice and the new moon, so it is the first day of the eleventh lunisolar month; in UTC+7 the solstice moves up to 12-21, which makes 12-22 the first day of the twelfth month.

Chinese New Year can differ by a month as well. Under this calendar convention, the first day of the first lunisolar month of 1985 maps to February 20 in UTC+8 and January 21 in UTC+7:
```go
fmt.Println(calendar.Solar(1985, 1, 1, false, 8.0))
fmt.Println(calendar.Solar(1985, 1, 1, false, 7.0))
```

### 5. Go-specific note

⚠️ Go's standard-library `time.Time` differs from this package in calendar handling:

- Before 1582-10-15 Go uses the proleptic Gregorian calendar rather than the Julian calendar. Without `Add`, it generally works as expected.
- So **before 1582-10-15, `time.Time.Weekday()` does not agree with this package**:
  for 1582-10-04 this package reports Thursday, Go reports Monday.

#### Weekdays and date arithmetic
To obtain the weekday this package uses:

```go
// date should be the local midnight of the target day.
weekday := int(calendar.Date2JD(date)+1.5) % 7
// 0 means Sunday, 1 means Monday, ..., 6 means Saturday.
```
Using `Add` or `AddDate` on a `time.Time` before 1582 runs through the proleptic Gregorian calendar and lands one day away from the Julian calendar whenever it crosses a leap day that only the Julian calendar has.
For example, 700 is a leap year in the Julian calendar but not in Go's proleptic Gregorian calendar.

### 6. Julian-only leap days (for example 700-02-29)
Before 1582, years divisible by 100 but not 400 (such as 100, 700 or 1500) have a 29 February in the Julian calendar,
while Go's `time.Time` uses the proleptic Gregorian calendar and has no such day (`time.Date(700, 2, 29, ...)` normalises to 700-03-01). The library accepts 700-02-29 as an existing day, with these constraints:

- `Time.Solar()` / `LunarTime.SolarDate`: the library's canonical output, always the **following day**
  (700-03-01 for 700-02-29), matching `basic.JD2DateByZone`;
- `Time.JulianOnly()`: whether the lunar date exists only in the Julian calendar (JSON field `julianOnly`);
- `Time.JD()`: the exact Julian day; for a Julian-only leap day it is one day earlier than `Solar()`,
  otherwise the two agree (JSON field `jd`).

```go
julian, _ := calendar.SolarToLunarByYMD(700, 2, 29)
fmt.Println(julian.Solar().Format("2006-01-02"), julian.JulianOnly(), julian.JD(), julian.Lunar().MonthDay())
// 0700-03-01 true 1.9767915e+06 二月初五
```

> Two kinds of entry point keep this leap day: the integer year/month/day entry points `SolarToLunarByYMD` / `LunarToSolarByYMD`, and a direct
> `basic.JDCalc(700, 2, 29)` call returning the exact Julian day `1976791.5`; building a `time.Time` first is the step that loses it.

### 7. Several Gregorian candidates for one lunar date

The dual-numbering reforms (Wang Mang 9-23 CE, Emperor Ming of Wei 237-240 CE, Wu Zetian 689-700 CE, Emperor Suzong of Tang 761-762 CE)
and the Taichu calendar handoff (104 BCE) give one lunar date two legal Gregorian days. `Solar()` remains the library default, and
`SolarCandidates()` returns every candidate with the first element always equal to `Solar()`:

```go
res, _ := calendar.LunarToSolarByYMD(700, 11, 1, false)
fmt.Println(res.Solar().Format("2006-01-02"))
fmt.Println(len(res.SolarCandidates()))
// 0700-12-15
// 2
```

Outside the reform windows there is a single candidate and `SolarCandidates()` returns a one-element
slice; a Julian-only leap day also reports only its canonical output, because its single legal day
cannot be expressed as a `time.Time` - use `JD()` for the exact date.

## Calendar conversion

### Gregorian to lunar

- **Input**: a Gregorian date, as `time.Time`
- **Output**: a `calendar.Time` object, holding one or more matching lunisolar dates
- **The result carries**:
  - a full lunisolar date description
  - sexagenary (ganzhi) year, month, and day
  - dynasty, emperor, and era name
  - the complete structured lunisolar record

### Lunar to Gregorian

Two calling styles:

#### Style 1: pass a lunisolar string
Formats:
1. `era name + year + month + day`, for example **`"元丰六年十月十二"`** (prefix a leap month with `闰`; day names read `初一`, `二十`, and so on)
2. `era name + year + month + ganzhi day`, for example **`"元嘉二十七年七月庚午"`**
3. `year + month + day`, for example **`"二零二五年正月初一"`** (prefix a leap month with `闰`; suits modern dates)
4. `year + month + ganzhi day`, for example **`"二零二五年正月戊戌日"`**
5. `Arabic digits + month + day`: Chinese numerals may be written as Arabic digits, for example **`"2025年1月1日"`**, which stands for `二零二五年正月初一`
6. Historical cases: month names could differ from modern usage (under Wu Zetian, `正月` and `一月` denoted different months); such month names are read as Chinese numerals

> ⚠️ A lunisolar year and a Gregorian year do not line up exactly. 2025-01-28, Chinese New Year's Eve, is the 29th day of the 12th lunisolar month of 2024, so its string is `"二零二四年腊月廿九"`.

#### Style 2: pass numeric fields
- **Parameters**: year (`int`), month (`int`), day (`int`), leap-month flag (`bool`)
- **Semantics**: locate the date by lunisolar year, month, day, and the leap-month flag; for modern conversions

One lunisolar day can have several legal Gregorian candidates: `Solar()` takes the first, `SolarCandidates()` returns all of them (see note 7 above).

### Code example

```go
package main

import (
	"github.com/starainrt/astro/calendar"
	"encoding/json"
	"fmt"
	"time"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)

	// Example 1: Gregorian to lunisolar. This date is in the Three Kingdoms period, so multiple parallel-calendar results are returned.
	date := time.Date(240, 1, 1, 8, 8, 8, 8, cst)
	lunar, _ := calendar.SolarToLunar(date)
	fmt.Println(lunar.LunarDescWithEmperor())

	// Structured lunisolar information for each matching historical result.
	info := lunar.LunarInfo()
	data, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println(string(data))

	// Example 2: lunisolar to Gregorian by string. This is the date from Su Shi's 《记承天寺夜游》.
	solar, _ := calendar.LunarToSolar("元丰六年十月十二日")
	for _, v := range solar {
		fmt.Println(v.Time())
		fmt.Println(v.LunarDescWithEmperor())
	}

	// Example 3: lunisolar to Gregorian by numeric fields. 2026 month 1 day 1 is Chinese New Year.
	modernDate, _ := calendar.LunarToSolarByYMD(2026, 1, 1, false)
	fmt.Println(modernDate.Time())
}
```

Output:

```text
[魏明帝 景初三年腊月二十 蜀后主 延熙二年冬月十九 吴大帝 赤乌二年冬月二十] // one Gregorian instant maps to parallel Three Kingdoms lunisolar results
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
    "calendarSystem": "",
    "calendarName": "",
    "jd": 1808717.8389814815,
    "dynasty": "魏",
    "emperor": "魏明帝",
    "nianhao": "景初",
    "yearOfNianhao": 3,
    "eraDesc": "景初三年",
    "lunarWithNianhaoDesc": "景初三年腊月二十",
    "chineseZodiac": "羊"
  },
  {
    "solarDate": "0240-01-01T08:08:08.000000008+08:00",
    "lunarYear": 239,
    "lunarYearChn": "二三九",
    "lunarMonth": 11,
    "lunarDay": 19,
    "isLeap": false,
    "lunarMonthDayDesc": "冬月十九",
    "ganzhiYear": "己未",
    "ganzhiMonth": "丙子",
    "ganzhiDay": "辛未",
    "calendarSystem": "",
    "calendarName": "",
    "jd": 1808717.8389814815,
    "dynasty": "蜀",
    "emperor": "蜀后主",
    "nianhao": "延熙",
    "yearOfNianhao": 2,
    "eraDesc": "延熙二年",
    "lunarWithNianhaoDesc": "延熙二年冬月十九",
    "chineseZodiac": "羊"
  },
  {
    "solarDate": "0240-01-01T08:08:08.000000008+08:00",
    "lunarYear": 239,
    "lunarYearChn": "二三九",
    "lunarMonth": 11,
    "lunarDay": 20,
    "isLeap": false,
    "lunarMonthDayDesc": "冬月二十",
    "ganzhiYear": "己未",
    "ganzhiMonth": "丙子",
    "ganzhiDay": "辛未",
    "calendarSystem": "",
    "calendarName": "",
    "jd": 1808717.8389814815,
    "dynasty": "吴",
    "emperor": "吴大帝",
    "nianhao": "赤乌",
    "yearOfNianhao": 2,
    "eraDesc": "赤乌二年",
    "lunarWithNianhaoDesc": "赤乌二年冬月二十",
    "chineseZodiac": "羊"
  }
] // structured lunisolar records, one object per matching historical result
1083-11-24 00:00:00 +0800 CST // Gregorian date corresponding to 元丰六年十月十二日
[宋神宗 元丰六年十月十二 辽道宗 大康九年十月十二] // the same day also matches a Liao calendar result
2026-02-17 00:00:00 +0800 CST // Chinese New Year in 2026
```

## Solar terms

`JieQi(year, term)` returns the exact solar-term instant computed by the modern astronomical algorithm.

`CalendricalJieQi(year, term)` returns the date on which the solar term falls under the default calendar, fixed at 00:00 Beijing time for that day.

Use `CalendricalJieQiWithCalendar(year, term, system)` when a specific ancient calendar system is required.

```go
package main

import (
	"fmt"

	"github.com/starainrt/astro/calendar"
)

func main() {
	// Beginning of Spring in 2020. Solar-term constants correspond to apparent solar longitude.
	fmt.Println(calendar.JieQi(2020, calendar.JQ_立春))
	// Winter Solstice in 2020.
	fmt.Println(calendar.JieQi(2020, calendar.JQ_冬至))
	// March Equinox in 2020.
	fmt.Println(calendar.JieQi(2020, calendar.JQ_春分))
	// Direct longitude input is also supported; 0 degrees is the March Equinox.
	fmt.Println(calendar.JieQi(2020, 0))
}
```

Output:

```text
2020-02-04 17:03:20.471614301 +0800 CST // Beginning of Spring
2020-12-21 18:02:20.648710727 +0800 CST // Winter Solstice
2020-03-20 11:49:37.149532735 +0800 CST // March Equinox
2020-03-20 11:49:37.149532735 +0800 CST // same result from direct longitude input
```

Calendrical solar-term example:

```go
date, err := calendar.CalendricalJieQi(1582, calendar.JQ_冬至)
fmt.Println(date, err)

date, err = calendar.CalendricalJieQiWithCalendar(-202, calendar.JQ_冬至, calendar.AncientCalendarQinHan)
fmt.Printf("%d-%02d-%02d %v\n", date.Year(), int(date.Month()), date.Day(), err)
```

Output:

```text
1582-12-22 00:00:00 +0800 CST <nil>
-202-12-25 <nil>
```

## Parameter and result conventions

### Units and Parameter Conventions

- Year numbering is astronomical: `0` is 1 BCE, `-1` is 2 BCE; Gregorian results are supported from 721 BCE through 3000 CE.
- The `timezone` argument of `Solar` and `Lunar` is in hours: `8.0` is UTC+8 and `7.0` is UTC+7. `JieQi`, `WuHou`, and `CalendricalJieQi*` take no time zone and always output Beijing time (UTC+8).
- The lunisolar month counts the distance from the first month: month 1 is `正月`, month 2 is `二月`, and so on. A leap month is marked by the separate `leap`/`IsLeap` flag and is never inferred from the month number.
- The lunisolar year argument uses the civil year as a proxy but starts at the lunisolar new year: the 30th day of the 12th month of the Ji-Hai year must be passed as `Solar(2019, 12, 30, false, 8)`, not `Solar(2020, 12, 30, false, 8)`.
- `Date2JD` reads only the year/month/day/hour/minute/second fields of a `time.Time` and performs no time-zone conversion, so Beijing midnight and UTC midnight of the same civil date give the same Julian day; `JD2Date` returns a time in the local zone of the host.

### Time Scale

- Every public instant is treated as a civil value, meaning the fields of the `time.Time` are read directly as a UTC label.

  When the library converts civil time to TT, values before 1972-01-01 are treated as UT1 and the exact window uses the built-in leap-second table; for explicit conversion use the root package's `astro.UT1FromUTC` / `astro.TTFromUTC` / `astro.DUT1`.
- `JieQi` and `WuHou` return civil instants in Beijing time with full fractional-second precision; `CalendricalJieQi*` always returns 00:00 Beijing time on the matching day.
- `JD`, `Date2JD`, and `NowJD` are all Julian-day floats on that same civil convention: the Beijing civil day `2026-02-17` corresponds to `2461088.5`.

### Zero Values and Out-of-Range Input

- The ten Gregorian days from 1582-10-05 through 1582-10-14 do not exist; the affected entry points report an error instead of silently normalising to a neighbouring date.
- `SolarToLunar` / `SolarToLunarByYMD` return an error outside `[-721,3000]`. `LunarToSolar*` is limited to the same Gregorian range, and the boundary lunisolar year `-722` is accepted while its result still falls inside that range.
- `CalendricalJieQiWithCalendar` returns an error for the Chunqiu calendar and for years without calendrical solar-term data; `LunarToSolarWithCalendar` accepts an era-name description only when the ancient era table matches and errors otherwise.
- `SolarCandidates()` returns a one-element slice holding `Solar()` when there is no second candidate, never an empty slice; `Lunar()` returns a zero-value `LunarTime` when there is no candidate, and its `JulianOnly()` then reports `false`.
- `LunarToSolarSingle` is deprecated, use `LunarToSolarByYMD`. `AncientCalendarDefault` is the zero value, and the explicit-calendar entry points fall back to the default implementation when they receive it.
- `LunarTime.CalendarSystem()` and `CalendarName()` return empty strings under default routing and in the modern range, which means "no explicit ancient calendar" rather than an error.

### Accuracy and Applicability

- The modern range follows the current Chinese calendar standard GB/T 33661-2017 and is recommended for `[1929,3000]`.

  `Solar`/`Lunar` use the same new-moon and solar-term rules and merely allow a custom time zone, so outside the recommended years the dates can differ from historical records.
- `[-103,1912]` uses the built-in calendar tables, `[-220,-104]` uses the reconstructed Qin/Han Zhuanxu calendar, and `[-721,-221]` uses the reconstructed default pre-Qin calendars.

  All three ranges are reconstructions and are not guaranteed to match every day actually promulgated at the time.
- New moons and solar terms are computed with modern astronomy, so for ancient dates their instants differ from contemporary observation and the resulting lunisolar dates can differ from the historical record.
- The ancient-calendar reconstruction stops at the level of whole days: `CalendricalJieQi*` returns the calendar-compatible date while `JieQi` returns the modern astronomical instant, and the two are not interchangeable.

### Related Manuals

The astronomical definition of a solar term (apparent solar longitude) and the determination of new moons are covered by [Sun and Moon](sun-moon.md#sun-and-moon-position) and [Sun and Moon](sun-moon.md#lunar-phases); those pages give modern astronomical quantities, while this manual gives the calendrical arrangement.
