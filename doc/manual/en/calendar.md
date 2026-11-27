# Calendar & Solar Terms

[中文](../calendar.md) | [Back to README](../../../README.en.md)

The `calendar` package converts between civil dates and the Chinese lunisolar calendar, computes the 24 solar terms and 72 pentads, looks up sexagenary (GanZhi) cycles, zodiac animals, dynasty/era names, and historical ancient calendars from 721 BCE to 3000 CE (astronomical years `-721` to `3000`, where year `0` is 1 BCE).

## Contents

- [Basic Example](#basic-example)
- [Usage](#usage)
  - [Chinese Calendar → Civil Date (Text, Era Name, or YMD)](#chinese-calendar--civil-date-text-era-name-or-ymd)
  - [Parallel Dynasties & Calendar-Reform Duplicates](#parallel-dynasties--calendar-reform-duplicates)
  - [Ancient Calendars & Calendrical Solar Terms](#ancient-calendars--calendrical-solar-terms)
  - [Custom Time Zones & Julian-Only Leap Days (e.g. 700-02-29)](#custom-time-zones--julian-only-leap-days-eg-700-02-29)
- [API Reference](#api-reference)
- [Conventions and Limitations](#conventions-and-limitations)

## Basic Example

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

	day, err := calendar.SolarToLunar(date)
	if err != nil {
		log.Fatal(err)
	}
	l := day.Lunar()

	fmt.Printf("2026-02-17 -> Lunar %d %s (%s year, %s month, %s day, zodiac %s)\n",
		l.LunarYear(), l.MonthDay(), l.GanZhiYear(), l.GanZhiMonth(), l.GanZhiDay(), l.ShengXiao())
	fmt.Printf("2026 Start of Spring: %s CST  Winter Solstice: %s CST\n",
		calendar.JieQi(2026, calendar.JQ_立春).Format("2006-01-02 15:04:05"),
		calendar.JieQi(2026, calendar.JQ_冬至).Format("2006-01-02 15:04:05"))
}
```

Output:

```text
2026-02-17 -> Lunar 2026 正月初一 (丙午 year, 庚寅 month, 壬戌 day, zodiac 马)
2026 Start of Spring: 2026-02-04 04:02:07 CST  Winter Solstice: 2026-12-22 04:50:13 CST
```

> `day.Lunar()` returns the primary result. During periods with coexisting dynasties (e.g. Three Kingdoms, Song/Liao/Jin), call `day.Lunars()` or `day.LunarInfo()` (JSON-serializable) to get every regime's calendar date.

## Usage

### Chinese Calendar → Civil Date (Text, Era Name, or YMD)

- **Text parser `LunarToSolar`**: accepts Chinese era strings (`"元丰六年十月十二日"`), GanZhi days (`"元嘉二十七年七月庚午"`), Chinese or Arabic numerals (`"二零二五年正月初一"`, `"2025年1月1日"`, prefix `"闰"` for leap months), returning `[]calendar.Time`.
- **Numeric `LunarToSolarByYMD(year, month, day, isLeap)`**: `month` counts from the first month (`1`), with `isLeap` flagging leap months.

```go
res, err := calendar.LunarToSolar("元丰六年十月十二日")
if err != nil {
	log.Fatal(err)
}
for _, r := range res {
	fmt.Printf("Yuanfeng 6-10-12 -> %s  eras: %v\n", r.Solar().Format("2006-01-02"), r.LunarDescWithEmperor())
}

modern, err := calendar.LunarToSolarByYMD(2026, 1, 1, false)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Lunar 2026-01-01 -> %s\n", modern.Solar().Format("2006-01-02"))
```

Output:

```text
Yuanfeng 6-10-12 -> 1083-11-24  eras: [宋神宗 元丰六年十月十二 辽道宗 大康九年十月十二]
Lunar 2026-01-01 -> 2026-02-17
```

> Before Chinese New Year, the lunar year is one less than the civil year (e.g. civil `2025-01-28` is lunar `2024-12-29`).

### Parallel Dynasties & Calendar-Reform Duplicates

- **Multiple dynasties on one civil day**: read `Lunars()` or `LunarDescWithEmperor()`.
- **Multiple civil days for one lunar date**: during calendar reforms (Wang Mang, Emperor Ming of Wei, Wu Zetian, Tang Suzong, Taichu reform), a lunar date can occur twice. `Solar()` returns the primary date and `SolarCandidates()` returns every legal civil date (with `Solar()` always first):

```go
tri, _ := calendar.SolarToLunarByYMD(240, 1, 1)
fmt.Printf("240-01-01 Three Kingdoms: %v\n", tri.LunarDescWithEmperor())

dbl, _ := calendar.LunarToSolarByYMD(700, 11, 1, false)
cands := make([]string, 0, len(dbl.SolarCandidates()))
for _, c := range dbl.SolarCandidates() {
	cands = append(cands, c.Format("2006-01-02"))
}
fmt.Printf("700-11-01 Wu Zetian reform candidates: %v\n", cands)
```

Output:

```text
240-01-01 Three Kingdoms: [魏明帝 景初三年腊月二十 蜀后主 延熙二年冬月十九 吴大帝 赤乌二年冬月二十]
700-11-01 Wu Zetian reform candidates: [0700-12-15 0700-10-17]
```

### Ancient Calendars & Calendrical Solar Terms

- Default routing picks the historical system automatically (`-721..-221` pre-Qin, `-220..-104` Qin-Han Zhuanxu, `-103..1912` historical tables, `1913+` modern GB/T 33661-2017). To force a specific ancient system, use `...WithCalendar`.
- **`JieQi` vs `CalendricalJieQi`**: `JieQi(year, deg)` returns the exact modern astronomical solar-term instant (UTC+8); `CalendricalJieQi(year, deg)` returns the historical tabulated date (midnight UTC+8, available only for tabulated years).

```go
zx, _ := calendar.SolarToLunarByYMDWithCalendar(-300, 1, 5, calendar.AncientCalendarZhuanxu)
fmt.Printf("-300-01-05 (%s): %d %s\n", zx.Lunar().CalendarName(), zx.Lunar().LunarYear(), zx.Lunar().MonthDay())

cjq, _ := calendar.CalendricalJieQi(1582, calendar.JQ_冬至)
fmt.Printf("1582 calendrical winter solstice: %s\n", cjq.Format("2006-01-02"))
```

Output:

```text
-300-01-05 (颛顼历): -300 十一月廿四
1582 calendrical winter solstice: 1582-12-22
```

### Custom Time Zones & Julian-Only Leap Days (e.g. 700-02-29)

- **Custom time zones**: Modern Chinese calendar rules default to UTC+8. To apply the same lunisolar rules in another time zone (e.g. UTC+7), use `calendar.Solar(y, m, d, leap, tzHours)` / `calendar.Lunar(y, m, d, tzHours)`.
- **Julian-only leap days**: Years before 1582 divisible by 100 but not 400 (e.g. 700 CE) had a Feb 29 in the Julian calendar, which Go's proleptic-Gregorian `time.Time` cannot represent. Use `SolarToLunarByYMD(700, 2, 29)` and check `JulianOnly()` and `JD()` (`Solar()` shifts to `0700-03-01`):

```go
fmt.Printf("1985 New Year: UTC+8=%s  UTC+7=%s\n",
	calendar.Solar(1985, 1, 1, false, 8).Format("2006-01-02"),
	calendar.Solar(1985, 1, 1, false, 7).Format("2006-01-02"))

jul, _ := calendar.SolarToLunarByYMD(700, 2, 29)
fmt.Printf("700-02-29: JulianOnly=%v  JD=%.1f  lunar=%s  Solar()=%s\n",
	jul.JulianOnly(), jul.JD(), jul.Lunar().MonthDay(), jul.Solar().Format("2006-01-02"))
```

Output:

```text
1985 New Year: UTC+8=1985-02-20  UTC+7=1985-01-21
700-02-29: JulianOnly=true  JD=1976791.5  lunar=二月初五  Solar()=0700-03-01
```

## API Reference

| Group | Entry Points | Notes |
| --- | --- | --- |
| Civil → Lunar | `SolarToLunar(t)` / `SolarToLunarByYMD(y, m, d)` | Integer entry avoids losing Julian-only leap days |
| Lunar → Civil | `LunarToSolar(text)` / `LunarToSolarByYMD(y, m, d, leap)` | Text returns `[]Time`; numeric returns `Time` |
| Custom zone | `Solar(y, m, d, leap, tz)` / `Lunar(y, m, d, tz)` | `tz` in hours |
| Ancient calendars | `SolarToLunar[ByYMD]WithCalendar`, `LunarToSolar[ByYMD]WithCalendar` | Systems: `Chunqiu`, `Zhou`, `Lu`, `Huangdi`, `Yin`, `Xia1`, `Xia2`, `Zhuanxu`, `QinHan` |
| Solar terms | `JieQi(y, deg)`, `WuHou(y, deg)`, `CalendricalJieQi[WithCalendar]` | `deg` is solar longitude (`0°`=vernal equinox, `270°`=winter solstice, `315°`=start of spring, or `JQ_*` constants) |
| GanZhi & eras | `GanZhiOfYear`, `GanZhiOfDay`, `Time.Eras`, `LunarTime.GanZhiYear/Month/Day` | Unmatched era returns `ERR_NIANHAO_NOT_FOUND` |
| Julian days | `Date2JD`, `JD2Date`, `NowJD`, `Time.Add`, `Time.JD`, `Time.JulianOnly`, `Time.SolarCandidates` | `Time.Add` skips the 10-day 1582 Gregorian gap |

## Conventions and Limitations

- **Gregorian reform**: Dates from 1582-10-15 onward use Gregorian; 1582-10-04 and earlier use Julian. Dates 1582-10-05 through 1582-10-14 do not exist and return an error.
- **Go `time.Time` before 1582**: Because Go uses a proleptic Gregorian calendar, `time.Time.Weekday()` and `AddDate()` disagree with the Julian calendar before 1582-10-15. Compute Julian weekday via `int(calendar.Date2JD(date)+1.5) % 7` (0=Sunday) and shift dates via `calendar.Time.Add`.
- **`Date2JD` / `JD2Date`**: `Date2JD` reads wall-clock fields directly without converting zones (call `.UTC()` first for UTC JD); `JD2Date` returns local time.
- **Historical accuracy**: Modern conversions follow GB/T 33661-2017 (`1929–3000`); pre-1913 dates are reconstructed from historical tables and rules and may differ from occasional historical records.
