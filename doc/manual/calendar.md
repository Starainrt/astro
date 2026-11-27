# 历法转换与节气

[English](en/calendar.md) | [返回 README](../../README.md)

`calendar` 包提供公历与中国传统农历双向转换、二十四节气与七十二候、干支生肖、历史朝代年号以及先秦至秦汉古历系统，支持范围为公元前 721 年至公元 3000 年（天文纪年 `-721` 至 `3000`，其中 `0` 年即公元前 1 年）。

## 目录

- [基本示例](#基本示例)
- [主要功能](#主要功能)
  - [农历转公历](#农历转公历)
  - [并行政权与改历双候选](#并行政权与改历双候选)
  - [古历系统与历法节气](#古历系统与历法节气)
  - [自定义时区与儒略历独有闰日](#自定义时区与儒略历独有闰日)
- [API 概览](#api-概览)
  - [二十四节气常量](#二十四节气常量)
  - [LunarInfo 结构与 JSON 字段](#lunarinfo-结构与-json-字段)
- [约定与限制](#约定与限制)

## 基本示例

计算公历 2026-02-17 对应的农历日期、干支生肖以及当年立春与冬至时刻：

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

	fmt.Printf("公历 2026-02-17 → 农历 %d年%s（%s年 %s月 %s日，生肖%s）\n",
		l.LunarYear(), l.MonthDay(), l.GanZhiYear(), l.GanZhiMonth(), l.GanZhiDay(), l.ShengXiao())
	fmt.Printf("2026 立春: %s CST  冬至: %s CST\n",
		calendar.JieQi(2026, calendar.JQ_立春).Format("2006-01-02 15:04:05"),
		calendar.JieQi(2026, calendar.JQ_冬至).Format("2006-01-02 15:04:05"))
}
```

输出：

```text
公历 2026-02-17 → 农历 2026年正月初一（丙午年 庚寅月 壬戌日，生肖马）
2026 立春: 2026-02-04 04:02:07 CST  冬至: 2026-12-22 04:50:13 CST
```

`day.Lunar()` 返回主历法候选；多政权并存的历史时期，可通过 `day.Lunars()` 获取全部候选，或通过 `day.LunarInfo()` 获取结构化切片（支持 `json.Marshal`）。

## 主要功能

### 农历转公历

提供文本解析与数字参数两类入口：
- **`LunarToSolar(text)`**：支持年号（`"元丰六年十月十二日"`）、干支日（`"元嘉二十七年七月庚午"`）、中文或阿拉伯数字（`"二零二五年正月初一"`、`"2025年1月1日"`，闰月前加`"闰"`），返回 `[]calendar.Time`。
- **`LunarToSolarByYMD(year, month, day, isLeap)`**：`month` 为与正月的月序距离（正月为 `1`），闰月由 `isLeap` 单独标识。

```go
res, err := calendar.LunarToSolar("元丰六年十月十二日")
if err != nil {
	log.Fatal(err)
}
for _, r := range res {
	fmt.Printf("元丰六年十月十二 → 公历 %s  并行纪年: %v\n",
		r.Solar().Format("2006-01-02"), r.LunarDescWithEmperor())
}

modern, err := calendar.LunarToSolarByYMD(2026, 1, 1, false)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("农历 2026 正月初一 → 公历 %s\n", modern.Solar().Format("2006-01-02"))
```

输出：

```text
元丰六年十月十二 → 公历 1083-11-24  并行纪年: [宋神宗 元丰六年十月十二 辽道宗 大康九年十月十二]
农历 2026 正月初一 → 公历 2026-02-17
```

农历岁首与公历元旦不重合：例如公历 2025-01-28（除夕）属农历 **2024** 年腊月廿九，数字参数应传 `(2024, 12, 29, false)`。

### 并行政权与改历双候选

- **多政权并存**：同一公历日在三国、南北朝、宋辽金等时期对应多个政权历法，结果可通过 `Lunars()` 或 `LunarDescWithEmperor()` 读取。
- **改历双纪年**：王莽、魏明帝、武则天、唐肃宗与太初改历交接期间，同一农历日可对应两个合法公历日。`Solar()` 返回默认候选，`SolarCandidates()` 返回全部合法候选（首项恒等于 `Solar()`）。

```go
tri, _ := calendar.SolarToLunarByYMD(240, 1, 1)
fmt.Printf("公元 240-01-01 三国并行政权: %v\n", tri.LunarDescWithEmperor())

dbl, _ := calendar.LunarToSolarByYMD(700, 11, 1, false)
cands := make([]string, 0, len(dbl.SolarCandidates()))
for _, c := range dbl.SolarCandidates() {
	cands = append(cands, c.Format("2006-01-02"))
}
fmt.Printf("武则天改历双候选 (700年十一月初一): %v\n", cands)
```

输出：

```text
公元 240-01-01 三国并行政权: [魏明帝 景初三年腊月二十 蜀后主 延熙二年冬月十九 吴大帝 赤乌二年冬月二十]
武则天改历双候选 (700年十一月初一): [0700-12-15 0700-10-17]
```

### 古历系统与历法节气

- 默认按年份自动路由：`-721..-221` 先秦古历、`-220..-104` 秦汉颛顼历、`-103..1912` 历史历表、`1913+` 现代 GB/T 33661-2017。如需显式指定古历（春秋、周、鲁、黄帝、殷、夏一/二、颛顼、秦汉颛顼），使用 `...WithCalendar` 后缀入口。
- `JieQi(year, deg)` 返回现代天文算法计算的精确时刻（北京时间）；`CalendricalJieQi(year, deg)` 返回历史历法编排中节气所在的日期（北京时间当天 0 点，仅在有历表资料的年份可用）。

```go
zx, _ := calendar.SolarToLunarByYMDWithCalendar(-300, 1, 5, calendar.AncientCalendarZhuanxu)
fmt.Printf("公元前301-01-05 (%s): %d年%s\n", zx.Lunar().CalendarName(), zx.Lunar().LunarYear(), zx.Lunar().MonthDay())

cjq, _ := calendar.CalendricalJieQi(1582, calendar.JQ_冬至)
fmt.Printf("1582 历法冬至日: %s\n", cjq.Format("2006-01-02"))
```

输出：

```text
公元前301-01-05 (颛顼历): -300年十一月廿四
1582 历法冬至日: 1582-12-22
```

### 自定义时区与儒略历独有闰日

- **自定义时区**：中国农历定朔定气默认采用北京时间（UTC+8）。如需在其他时区按现行农历规则互转，调用 `calendar.Solar(y, m, d, leap, tzHours)` / `calendar.Lunar(y, m, d, tzHours)`。
- **儒略历独有闰日**：1582 年前“能被 100 整除但不能被 400 整除”的年份（如 700 年）在儒略历中存在 2 月 29 日，而 Go 标准库 `time.Time` 采用前推格里高利历无法表示该日。处理此类日期须使用整型入口 `SolarToLunarByYMD(700, 2, 29)`，并通过 `JulianOnly()` 与 `JD()` 获取精确日期（`Solar()` 返回次日 `0700-03-01`）。

```go
fmt.Printf("1985 春节对比: 东八区=%s  东七区=%s\n",
	calendar.Solar(1985, 1, 1, false, 8).Format("2006-01-02"),
	calendar.Solar(1985, 1, 1, false, 7).Format("2006-01-02"))

jul, _ := calendar.SolarToLunarByYMD(700, 2, 29)
fmt.Printf("700-02-29: JulianOnly=%v  JD=%.1f  农历=%s  Solar()回退=%s\n",
	jul.JulianOnly(), jul.JD(), jul.Lunar().MonthDay(), jul.Solar().Format("2006-01-02"))
```

输出：

```text
1985 春节对比: 东八区=1985-02-20  东七区=1985-01-21
700-02-29: JulianOnly=true  JD=1976791.5  农历=二月初五  Solar()回退=0700-03-01
```

## API 概览

| 分类 | 入口 | 说明 |
| --- | --- | --- |
| 公历 → 农历 | `SolarToLunar(t)` / `SolarToLunarByYMD(y, m, d)` | 后者不经 `time.Time`，保留儒略历独有闰日 |
| 农历 → 公历 | `LunarToSolar(text)` / `LunarToSolarByYMD(y, m, d, leap)` | 文本入口返回 `[]Time`；数字入口返回 `Time` |
| 自定义时区 | `Solar(y, m, d, leap, tz)` / `Lunar(y, m, d, tz)` | 按现行农历规则在指定时区偏移（小时）下互转 |
| 古历系统 | `SolarToLunar[ByYMD]WithCalendar` / `LunarToSolar[ByYMD]WithCalendar` | 常量：`AncientCalendarDefault`、`Chunqiu`、`Zhou`、`Lu`、`Huangdi`、`Yin`、`Xia1`、`Xia2`、`Zhuanxu`、`QinHan` |
| 节气与物候 | `JieQi(y, deg)` / `WuHou(y, deg)` / `CalendricalJieQi[WithCalendar]` | 现代节气/七十二候时刻（北京时间）与历法节气日 |
| 干支与年号 | `GanZhiOfYear` / `GanZhiOfDay` / `Time.Eras` / `LunarTime.GanZhiYear/Month/Day` | 解析年号失败返回 `ERR_NIANHAO_NOT_FOUND` |
| 儒略日与日期 | `Date2JD` / `JD2Date` / `NowJD` / `Time.Add` / `Time.JD` / `Time.JulianOnly` / `Time.SolarCandidates` | `Time.Add` 跨 1582 年改历自动跳过不存在的 10 天 |

### 二十四节气常量

`JieQi` 与 `WuHou` 的第二个参数为太阳视黄经（度），可直接传入数值或下列常量：

| 黄经 | 常量 | 黄经 | 常量 | 黄经 | 常量 | 黄经 | 常量 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 0° | `JQ_春分` | 90° | `JQ_夏至` | 180° | `JQ_秋分` | 270° | `JQ_冬至` |
| 15° | `JQ_清明` | 105° | `JQ_小暑` | 195° | `JQ_寒露` | 285° | `JQ_小寒` |
| 30° | `JQ_谷雨` | 120° | `JQ_大暑` | 210° | `JQ_霜降` | 300° | `JQ_大寒` |
| 45° | `JQ_立夏` | 135° | `JQ_立秋` | 225° | `JQ_立冬` | 315° | `JQ_立春` |
| 60° | `JQ_小满` | 150° | `JQ_处暑` | 240° | `JQ_小雪` | 330° | `JQ_雨水` |
| 75° | `JQ_芒种` | 165° | `JQ_白露` | 255° | `JQ_大雪` | 345° | `JQ_惊蛰` |

### `LunarInfo` 结构与 JSON 字段

`day.LunarInfo()` 返回 `[]LunarInfo`，JSON 键包括：`solarDate`、`lunarYear` / `lunarYearChn`、`lunarMonth` / `lunarDay` / `isLeap`、`lunarMonthDayDesc`、`ganzhiYear` / `ganzhiMonth` / `ganzhiDay`、`chineseZodiac`、`jd`、`julianOnly`（仅儒略历独有闰日输出）、`dynasty` / `emperor` / `nianhao` / `yearOfNianhao` / `eraDesc`、`lunarWithNianhaoDesc`（对应字段 `LunarWithEraDesc`）、`calendarSystem` / `calendarName`。

## 约定与限制

- **1582 格里高利改历**：1582-10-15 起采用格里高利历，1582-10-04 及以前采用儒略历；1582-10-05 至 10-14 这 10 个日期不存在，相关接口直接报错。
- **Go `time.Time` 差异**：Go 标准库在 1582 年前采用前推格里高利历，因此 `1582-10-15` 之前 `time.Time.Weekday()` 与 `AddDate()` 结果与儒略历不一致。求儒略历星期可用 `int(calendar.Date2JD(date)+1.5) % 7`（0 表示星期日）；跨 1582 年日期偏移应使用 `calendar.Time.Add`。
- **`Date2JD` 与 `JD2Date`**：`Date2JD` 直接读取年月日时分秒字段而不做时区换算（求 UTC JD 须先调用 `.UTC()`）；`JD2Date` 返回运行环境本地时区。
- **适用范围**：现代段按 GB/T 33661-2017 编排（推荐 `1929–3000`）；`-721..1912` 为依据历史历表与古历规则的复原结果，不保证与全部史籍记载逐日一致。
