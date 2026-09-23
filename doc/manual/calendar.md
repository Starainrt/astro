# 历法转换与节气

[English](en/calendar.md) | [返回 README](../../README.md)

本package支持公历与中国传统农历日期之间的相互转换，并提供节气信息。支持年份范围为公元前721年至公元3000年（公元前104年为历法表切换点，公元3000年后为纯理论实时计算外推）。
农历本质上是阴阳合历（Lunisolar Calendar），但为兼顾大众习惯与代码简洁性，相关函数命名采用 `Lunar` 而非更学术的 `Lunisolar`。

## 目录

- [公历转农历与节气](#公历转农历与节气)
- [API 参考](#api-参考)
  - [公历与农历互转](#公历与农历互转)
  - [儒略历独有闰日与多候选](#儒略历独有闰日与多候选)
  - [节气与物候](#节气与物候)
  - [古历系统](#古历系统)
  - [干支与年号](#干支与年号)
  - [结构与 JSON](#结构与-json)
- [常用场景](#常用场景)
  - [公历转农历与干支纪日](#公历转农历与干支纪日)
  - [节气时刻与历法节气](#节气时刻与历法节气)
  - [自定义时区与儒略历独有闰日](#自定义时区与儒略历独有闰日)
  - [改历双纪年的多个候选](#改历双纪年的多个候选)
  - [古历系统与年号口径](#古历系统与年号口径)
- [历法说明](#历法说明)
- [使用须知](#使用须知)
  - [1. 同一公历日期可能对应多个农历日期](#1-同一公历日期可能对应多个农历日期)
  - [2. 同一农历日期可能对应多个公历日期](#2-同一农历日期可能对应多个公历日期)
  - [3. 公历历法处理规则](#3-公历历法处理规则)
  - [4. 时区说明](#4-时区说明)
  - [5. Go 语言特别注意](#5-go-语言特别注意)
  - [6. 儒略历独有的闰日（如 700-02-29）](#6-儒略历独有的闰日如-700-02-29)
  - [7. 同一农历日的多个公历候选](#7-同一农历日的多个公历候选)
- [历法转换](#历法转换)
  - [公历转农历](#公历转农历)
  - [农历转公历](#农历转公历)
  - [代码示例](#代码示例)
- [节气](#节气)
- [参数与返回值约定](#参数与返回值约定)
  - [单位与参数口径](#单位与参数口径)
  - [时标](#时标)
  - [零值与越界](#零值与越界)
  - [精度与适用范围](#精度与适用范围)
  - [相关手册](#相关手册)

## 公历转农历与节气

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
	solar, err := calendar.SolarToLunar(date) //公历转农历
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(solar.Lunar().LunarYear(), solar.Lunar().MonthDay())
	fmt.Println(calendar.JieQi(2026, calendar.JQ_立春)) //节气UTC+08:00时间
}
```

转换失败时先处理 `error`。`Time.Lunar()` 返回首个农历候选；历史并行政权的全部结果由 `Time.Lunars()` 返回。节气计算采用北京时间。

## API 参考

### 公历与农历互转

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `SolarToLunar` | 公历 `time.Time` → 农历结果 | 输入的时区会被换到东八区再取民用日；返回 `Time`，可能含多个并行政权候选 |
| `SolarToLunarByYMD` | 整型公历年月日 → 农历结果 | 不经过 `time.Time`，因此不丢儒略历独有闰日；公元前用负数年份 |
| `LunarToSolar` | 农历描述字符串 → 公历候选 | 支持“年份+月+日”“年号+月+日”“…+干支日”等写法，返回 `[]Time` |
| `LunarToSolarByYMD` | 农历年月日 + 闰月标志 → 公历 | `month` 是“与正月的距离”，正月为 1；`leap` 为闰月标志 |
| `LunarToSolarSingle` | 同上，返回单个 `Time` | 已弃用，等价于 `LunarToSolarByYMD` |
| `Solar` | 农历年月日 → 公历（自定义时区） | `timezone` 单位小时；返回时刻的时区名固定为 `CST`，偏移等于传入的 `timezone` |
| `Lunar` | 公历年月日 → 农历（自定义时区） | 返回 `(农历年, 农历月, 农历日, 是否闰月, 文字描述)`；`timezone` 单位小时 |
| `Time` / `LunarTime` | 公农历双信息容器 | 字段与 JSON 键见“结构与 JSON” |

```go
// 前置：date 是一个 time.Time；字符串入口可以返回多个并行政权的结果。
solar, _ := calendar.SolarToLunar(date)
fmt.Println(solar.LunarDesc(), solar.Lunar().MonthDay())

lunar, _ := calendar.LunarToSolar("元丰六年十月十二日")
for _, v := range lunar {
	fmt.Println(v.Time().Format("2006-01-02"), v.Eras())
}

old, _ := calendar.LunarToSolarByYMD(-202, 1, 1, false) // 公元前用负数年份
fmt.Println(old.Solar().Format("2006-01-02"))
```

### 儒略历独有闰日与多候选

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `Time.JulianOnly` | 主历法农历日是否只存在于儒略历 | 布尔；1582 年前“能被 100 整除但不能被 400 整除”的年份的 2 月 29 日属于这一类 |
| `Time.JD` | 主历法农历日精确的儒略日 | 儒略历独有闰日比 `Solar()` 早一天，其余与 `Date2JD(Solar())` 一致 |
| `Time.SolarCandidates` | 该农历日的全部合法公历候选 | 首个恒等于 `Solar()`；单候选时返回只含 `Solar()` 的切片 |
| `LunarTime.JulianOnly` / `LunarTime.JD` | 单个候选上的同一信息 | 口径与 `Time` 上的同名方法一致 |
| `Date2JD` | `time.Time` → 儒略日 | 直接取 `date` 的年月日时分秒字段，不做时区换算 |
| `JD2Date` | 儒略日 → `time.Time` | 返回值落在运行环境的本地时区（底层使用 `Local`） |
| `NowJD` | 当前时刻的儒略日 | 与 `Date2JD` 同一口径 |
| `Time.Add` | 时间偏移 | 跨过 1582 改历空窗时改走儒略日轴，跳过不存在的 10 天 |

```go
// 整型年月日入口不丢闰日；700-02-29 在 Go 的 time.Time 里不存在。
julian, _ := calendar.SolarToLunarByYMD(700, 2, 29)
fmt.Println(julian.JulianOnly(), julian.JD(), julian.Solar().Format("2006-01-02"))

// 改历双纪年会给出多个合法公历候选，首个仍是 Solar()。
res, _ := calendar.LunarToSolarByYMD(700, 11, 1, false)
fmt.Println(res.Solar().Format("2006-01-02"), len(res.SolarCandidates()))

fmt.Println(calendar.Date2JD(time.Date(2026, 2, 17, 0, 0, 0, 0, time.UTC)), calendar.JD2Date(2461088.5))
```

### 节气与物候

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `JieQi` | 节气时刻 | 北京时间；`term` 是太阳视黄经，单位度 |
| `CalendricalJieQi` | 历法相符节气日期 | 默认历法下节气落在的日期，时间固定为北京时间当天 0 点 |
| `CalendricalJieQiWithCalendar` | 历法相符节气日期（显式古历） | 春秋历及缺少历法节气资料的年份返回错误 |
| `WuHou` | 物候时刻 | 北京时间；七十二候按 5° 一候，入参与 `JieQi` 同口径 |
| `JQ_春分` … `JQ_惊蛰` | 节气常量族 | 15° 档位的太阳视黄经，可直接当 `JieQi`/`WuHou` 的 `term` |

节气常量一共 24 个，按黄经递增排列：

| 常量 | 黄经 | 常量 | 黄经 | 常量 | 黄经 |
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
// term 是太阳视黄经（度），也可以直接传数值，例如春分为 0。
fmt.Println(calendar.JieQi(2020, calendar.JQ_立春))
fmt.Println(calendar.JieQi(2020, calendar.JQ_冬至))
fmt.Println(calendar.WuHou(2020, calendar.JQ_立春+5))

termDate, err := calendar.CalendricalJieQi(1582, calendar.JQ_冬至)
fmt.Println(termDate, err)
```

### 古历系统

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `AncientCalendarSystem` | 古历系统枚举 | 字符串类型 |
| `AncientCalendarDefault` | 默认路由（零值） | 按年份自动选择历法，显式古历入口遇到它时退回默认实现 |
| `AncientCalendarChunqiu` / `AncientCalendarZhou` / `AncientCalendarLu` | 先秦春秋、周、鲁历 | 值分别为 `chunqiu`/`zhou`/`lu` |
| `AncientCalendarHuangdi` / `AncientCalendarYin` | 先秦黄帝历、殷历 | 值分别为 `huangdi`/`yin` |
| `AncientCalendarXia1` / `AncientCalendarXia2` | 夏历（冬至版）与夏历（雨水版） | 值分别为 `xia1`/`xia2` |
| `AncientCalendarZhuanxu` | 先秦颛顼历 | 值 `zhuanxu` |
| `AncientCalendarQinHan` | 秦汉颛顼历 | 值 `qin_han` |
| `SolarToLunarWithCalendar` | 公历 `time.Time` → 农历（显式古历） | 输入先换到东八区再取民用日 |
| `SolarToLunarByYMDWithCalendar` | 整型公历年月日 → 农历（显式古历） | 不经过 `time.Time` |
| `LunarToSolarWithCalendar` | 农历描述 → 公历（显式古历） | 年号描述只在古历年号表命中时可用，否则报错 |
| `LunarToSolarByYMDWithCalendar` | 农历年月日 + 闰月标志 → 公历（显式古历） | 古历无数据的年份返回错误 |

```go
// 零值 AncientCalendarDefault 走默认路由；显式古历只在有数据的年份可用。
zhuanxu, err := calendar.SolarToLunarByYMDWithCalendar(-300, 1, 5, calendar.AncientCalendarZhuanxu)
fmt.Println(zhuanxu.Lunar().CalendarSystem(), zhuanxu.Lunar().CalendarName(), err)

back, err := calendar.LunarToSolarByYMDWithCalendar(
	zhuanxu.Lunar().LunarYear(), zhuanxu.Lunar().LunarMonth(), zhuanxu.Lunar().LunarDay(),
	zhuanxu.Lunar().IsLeap(), calendar.AncientCalendarZhuanxu)
fmt.Println(back.Solar().Format("2006-01-02"), err)
```

### 干支与年号

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `GanZhiOfYear` | 年干支 | 入参是公历年，返回“甲子”式两字字符串 |
| `GanZhiOfDay` | 日干支 | 取 `t` 的年月日字段，按北京时间当天 0 点计算 |
| `LunarTime.GanZhiYear` / `GanZhiMonth` / `GanZhiDay` | 单个农历候选的年、月、日干支 | 闰月的月干支从上一个月 |
| `Era` | 年号表记录 | 字段 `Year`/`Emperor`/`Nianhao`/`OtherNianHaoStart`/`Dynasty`/`Offset` |
| `EraDesc` | 年号描述 | 字段 `YearOfNianHao`/`Emperor`/`Nianhao`/`Dynasty` |
| `EraDesc.String` | 年号字符串 | 第一年写作“元年”，其余写作中文数字加“年” |
| `Time.Eras` | 全部候选的年号信息 | 返回 `[]EraDesc` |
| `LunarTime.Eras` | 单个候选的年号信息 | 返回 `[]EraDesc` |
| `ERR_NIANHAO_NOT_FOUND` | 未找到年号的哨兵错误 | 文本入口解析年号失败时返回 |

```go
fmt.Println(calendar.GanZhiOfYear(2026))
fmt.Println(calendar.GanZhiOfDay(time.Date(2026, 2, 17, 0, 0, 0, 0, time.UTC)))

res, _ := calendar.LunarToSolarByYMD(1083, 10, 12, false)
fmt.Println(res.Lunar().GanZhiYear(), res.Lunar().GanZhiMonth(), res.Lunar().GanZhiDay())
for _, era := range res.Eras() {
	fmt.Println(era.Dynasty, era.Emperor, era.String())
}
```

### 结构与 JSON

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `Time` | 公农历双信息容器 | `Solar()`/`Time()` 取公历，`Lunar()`/`Lunars()` 取农历候选，`LunarInfo()` 取结构化信息 |
| `LunarTime` | 单个农历候选 | `LunarYear`/`LunarMonth`/`LunarDay`/`IsLeap` 与 `MonthDay`、干支、生肖、历法名 |
| `LunarInfo` | 结构化农历信息 | 字段与 JSON 键见下表 |
| `Time.Solar` / `Time.Time` | 内部保存的公历 `time.Time` | 不做时区或历法再计算；两者同义 |
| `Time.Lunar` / `Time.Lunars` | 首个 / 全部农历候选 | 无结果时 `Lunar` 返回零值 `LunarTime` |
| `Time.LunarInfo` / `LunarTime.LunarInfo` | 结构化信息切片 | 存在并行年号时返回多条记录 |
| `Time.LunarDesc` / `Time.LunarDescWithEmperor` / `Time.LunarDescWithDynasty` / `Time.LunarDescWithDynastyAndEmperor` | 全部候选的描述族 | 均返回 `[]string` |
| `LunarTime.LunarDesc` / `LunarDescWithEmperor` / `LunarDescWithDynasty` / `LunarDescWithDynastyAndEmperor` | 单个候选的描述族 | 均返回 `[]string` |
| `LunarTime.LunarYear` / `LunarMonth` / `LunarDay` / `IsLeap` | 农历年、月、日与闰月标志 | `LunarMonth` 是“与正月的距离” |
| `LunarTime.MonthDay` | 月日描述 | 十一月写作“冬月”、十二月写作“腊月” |
| `LunarTime.ShengXiao` / `LunarTime.Zodiac` | 生肖 | 两者同义，按农历年对 12 取模 |
| `LunarTime.CalendarSystem` / `LunarTime.CalendarName` | 所属古历系统与中文名 | 默认路由与现代段为空字符串 |

`LunarInfo` 的 JSON 键与字段：

| JSON 键 | 字段 | 口径 |
| --- | --- | --- |
| `solarDate` | `SolarDate` | 公历时刻 |
| `lunarYear` / `lunarYearChn` | `LunarYear` / `LunarYearChn` | 农历年的公历映射及其中文数字写法 |
| `lunarMonth` / `lunarDay` | `LunarMonth` / `LunarDay` | 月是“与正月的距离”，日范围 `[1,30]` |
| `isLeap` | `IsLeap` | 是否闰月 |
| `lunarMonthDayDesc` | `LunarMonthDayDesc` | 月日描述，十一月作“冬月”、十二月作“腊月” |
| `ganzhiYear` / `ganzhiMonth` / `ganzhiDay` | `GanzhiYear` / `GanzhiMonth` / `GanzhiDay` | 年、月、日干支 |
| `calendarSystem` / `calendarName` | `CalendarSystem` / `CalendarName` | 古历系统与名称，默认路由为空串 |
| `jd` | `JD` | 该农历日精确的儒略日 |
| `julianOnly` | `JulianOnly` | 带 `omitempty`，只有儒略历独有闰日才出现 `true` |
| `dynasty` / `emperor` / `nianhao` / `yearOfNianhao` / `eraDesc` | `Dynasty` / `Emperor` / `Nianhao` / `YearOfNianhao` / `EraDesc` | 朝代、皇帝、年号与年号纪年 |
| `lunarWithNianhaoDesc` | `LunarWithEraDesc` | 注意字段名与 JSON 键不同 |
| `chineseZodiac` | `ChineseZodiac` | 生肖 |

```go
// 前置：需要 import "encoding/json"。
res, _ := calendar.SolarToLunarByYMD(2026, 2, 17)
for _, info := range res.LunarInfo() {
	fmt.Println(info.LunarYear, info.LunarMonthDayDesc, info.GanzhiDay, info.ChineseZodiac, info.JD)
}
data, _ := json.Marshal(res.LunarInfo()[0])
fmt.Println(string(data))
```

## 常用场景

### 公历转农历与干支纪日

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

`SolarToLunar` 与 `GanZhiOfDay` 共用同一套纪日口径，所以末行与 `GanZhiDay()` 一致。历史并行政权会让同一个公历日给出多个农历候选（`Lunar()` 取首个、`Lunars()` 取全部），口径见[使用须知](#使用须知)；闰月由 `IsLeap()` 单独标记，见[参数与返回值约定](#参数与返回值约定)。

### 节气时刻与历法节气

```go
fmt.Println(calendar.JieQi(2020, calendar.JQ_立春))                // 现代天文时刻
termDate, err := calendar.CalendricalJieQi(1582, calendar.JQ_冬至) // 历法日期
fmt.Println(termDate, err)
```

```text
2020-02-04 17:03:20.471614301 +0800 CST
1582-12-22 00:00:00 +0800 CST <nil>
```

`JieQi` 给的是现代天文算出的精确时刻，`CalendricalJieQi` 给的是历法编排的日期（固定北京时间 0 点），两者不可互换；历法节气只在有资料的年份可用（2026 年会返回错误），细节见[节气](#节气)与[历法说明](#历法说明)。

### 自定义时区与儒略历独有闰日

```go
fmt.Println(calendar.Solar(1985, 1, 1, false, 8.0), calendar.Solar(1985, 1, 1, false, 7.0))
julian, _ := calendar.SolarToLunarByYMD(700, 2, 29) // 儒略历独有闰日
fmt.Println(julian.Solar().Format("2006-01-02"), julian.JulianOnly(), julian.JD())
```

```text
1985-02-20 00:00:00 +0800 CST 1985-01-21 00:00:00 +0700 CST
0700-03-01 true 1.9767915e+06
```

定朔定气默认按北京时间，换到东七区后同一个农历日的公历日期会改变（示例中正月初一由 2 月 20 日变成 1 月 21 日，朔与冬至落点改变会牵动整个月序）；`Solar`/`Lunar` 的 `timezone` 参数单位是小时。700-02-29 只存在于儒略历：`Solar()` 固定在次日、精确日期看 `JD()`，入口与口径见[使用须知](#使用须知)。

### 改历双纪年的多个候选

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

改历双纪年（王莽、魏明帝、武则天、唐肃宗）与太初改历交接，会让同一个农历日对应两个合法公历日；首行是默认候选 `Solar()`，`SolarCandidates()` 返回全部候选，非改历年份只返回单元素切片，见[使用须知](#使用须知)。

### 古历系统与年号口径

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

显式古历入口只在有数据的年份可用，`AncientCalendarDefault` 是零值，会退回按年份自动路由；`CalendarSystem()`/`CalendarName()` 在默认路由与现代段返回空字符串，年号以 `Eras()`/`LunarDescWithEmperor()` 为准，口径见[历法说明](#历法说明)与[古历系统](#古历系统)。

## 历法说明

- **默认路由**：按年份自动选择，先秦段使用春秋/古六历重建，`-220..-104` 使用秦汉颛顼历，`-103..1912` 使用历表，`1913` 年后使用现代算法。
- **显式古历**：如果需要指定某一古历系统，请使用 `SolarToLunarWithCalendar` / `LunarToSolarWithCalendar` 这类 API。
- **数据来源**：古历部分主要参考《寿星天文历》；使用 [ytliu0教授的网站数据](https://ytliu0.github.io/ChineseCalendar/index_simp.html)做验证校验；现代段依据GB/T 33661-2017编排，通过 VSOP87、ELP定气定朔 。
- **节气**：`JieQi` 返回现代天文计算的节气时刻；`CalendricalJieQi` 返回历法相符节气日期。

---

## 使用须知

### 1. 同一公历日期可能对应多个农历日期
在多个政权并存的历史时期（如三国时期），不同政权可能使用不同历法，造成同一公历日期对应多个农历日期。本程序尽可能提供所有可能的转换结果。

### 2. 同一农历日期可能对应多个公历日期
不仅因多个政权历法不同，同一政权在历法改革中也可能出现此类情况。例如，武则天改历后，圣历三年出现了两个腊月。

### 3. 公历历法处理规则
本程序基于儒略日进行计算，公历部分处理规则如下：
- 1582年10月15日之后：使用格里高利历
- 1582年10月4日之前：使用儒略历
- 公元8年之前：使用逆推儒略历
- 1582年10月4日的下一天为1582年10月15日
- 1582年10月5日到10月14日这10个公历日期不存在，相关接口会直接拒绝
- 年份表示：0年表示公元前1年，-1年表示公元前2年，以此类推

### 4. 时区说明

本package主要面向中国历法，因此定气和定朔的计算默认采用北京时间（UTC+8）。对于使用其他时区的地区，若直接套用中国农历的编排规则，可能会产生日期偏差。

为方便探索与研究，本包 提供了底层方法 `Solar` 和 `Lunar`，它们支持在**自定义时区**下，按照**现行中国农历算法（GB/T 33661-2017）** 进行公历与农历的相互转换。
如果只需北京时间下的标准转换，请直接使用封装好的 `SolarToLunar` 和 `LunarToSolar` 方法。

**示例**：农历规则要求冬至必须落在农历十一月。以1984年冬至为例，计算可得：
```go
ws := calendar.JieQi(1984, 270)
fmt.Println(ws)
fmt.Println(moon.ClosestShuoYue(ws))
```

| 节令 | 东八区 (UTC+8) | 东七区 (UTC+7) |
|------|----------------|----------------|
| 冬至 | 1984-12-22     | 1984-12-21     |
| 朔日 | 1984-12-22     | 1984-12-22     |

可见，对于东八区（中国），1984年12月22日既是冬至又是朔日，因此该日为农历十一月初一；而在东七区，冬至提前至12月21日，导致12月22日已成为腊月初一。

类似地，春节的公历日期也可能相差一个月。按本例的历法口径，1985 年正月初一在东八区对应 2 月 20 日，在东七区对应 1 月 21 日：
```go
fmt.Println(calendar.Solar(1985, 1, 1, false, 8.0))
fmt.Println(calendar.Solar(1985, 1, 1, false, 7.0))
```

### 5. Go 语言特别注意

⚠️ Go 标准库 `time.Time` 在历法处理上与本程序存在差异：

- Go 语言在1582年10月15日之前使用逆推格里高利历，而非儒略历。若不使用 `Add` 方法，一般可正常使用。
- 因此，**在1582年10月15日之前，`time.Time.Weekday()` 返回结果与本程序计算结果不一致**。
  例如：1582年10月4日，本程序为星期四，Go 语言判断为星期一。

#### 计算星期与跨日运算
如需获得与本程序一致的星期数，可使用如下方法：

```go
// date 应为当日0时的 time.Time
weekday := int(calendar.Date2JD(date)+1.5) % 7
// 0表示星期日，1表示星期一，……，6表示星期六
```
在 1582 年之前使用 `time.Time` 的 `Add` 或 `AddDate` 会经过逆推格里高利历，跨过儒略历独有的闰日时与儒略历相差一天。
例如：700年儒略历为闰年，而 Go 使用的逆推格里高利历中700年不是闰年。

### 6. 儒略历独有的闰日（如 700-02-29）

1582 年以前"能被 100 整除但不能被 400 整除"的年份（如 100、700、1500 年）在儒略历中有 2 月 29 日，
而 Go 的`time.Time`使用逆推格里高利历，没有这一天（`time.Date(700, 2, 29, ...)` 会被规范化成 700-03-01）。本库承认 700-02-29 这一天存在，对应的约束如下：

- `Time.JulianOnly()`：该农历日是否只存在于儒略历（对应 JSON 字段 `julianOnly`）；
- `Time.JD()`：该日精确的儒略日；儒略历闰日比 `Solar()` 早一天，其余情况两者一致（对应 JSON 字段 `jd`）。
- `Time.Solar()` / `LunarTime.SolarDate`：库内标准输出，对于700-02-29Go标准库表示不出来的日期，固定返回为**后一天**（700-02-29 的后一天是 700-03-01），与 `basic.JD2DateByZone` 的约定一致；

```go
julian, _ := calendar.SolarToLunarByYMD(700, 2, 29)
fmt.Println(julian.Solar().Format("2006-01-02"), julian.JulianOnly(), julian.JD(), julian.Lunar().MonthDay())
// 0700-03-01 true 1.9767915e+06 二月初五
```

> 涉及这类日期时，不丢闰日的入口有两类：整型年月日入口 `SolarToLunarByYMD` / `LunarToSolarByYMD`，以及直接调用
> `basic.JDCalc(700, 2, 29)` 得到精确儒略日 `1976791.5`；先构造 `time.Time` 的那一步就会丢掉闰日。

### 7. 同一农历日的多个公历候选

改历双纪年（王莽 9–23 年、魏明帝 237–240 年、武则天 689–700 年、唐肃宗 761–762 年）与太初改历交接
（公元前 104 年）会让同一个农历日对应两个合法公历日。`Solar()` 仍是库内默认选择，`SolarCandidates()`
返回全部候选、首个恒等于 `Solar()`：

```go
res, _ := calendar.LunarToSolarByYMD(700, 11, 1, false)
fmt.Println(res.Solar().Format("2006-01-02"))
fmt.Println(len(res.SolarCandidates()))
// 0700-12-15
// 2
```

非改历年份的农历日只有一个候选，`SolarCandidates()` 返回仅含 `Solar()` 的slice；儒略历独有的闰日也只返回
标准输出，因为它唯一合法的那一天无法表示成 `time.Time`，精确日期见 `JD()`。

## 历法转换

### 公历转农历

- **输入**：公历日期 (`time.Time`)
- **输出**：`calendar.Time` 对象，可能包含多个对应的农历日期
- **功能**：可从返回对象中获取：
   - 农历日期的详细描述
   - 年、月、日的天干地支
   - 所属朝代、皇帝、年号等信息
   - 完整的结构化农历信息

### 农历转公历

支持两种调用方式：

#### 方式一：传入农历字符串
支持以下格式（示例）：
1. `年号+年+月+日`：如 **`"元丰六年十月十二"`**（闰月前加"闰"，日期格式为"初一"、"二十"等）
2. `年号+年+月+干支日`：如 **`"元嘉二十七年七月庚午"`**
3. `年份+月+日`：如 **`"二零二五年正月初一"`**（闰月前加"闰"，适用于现代日期）
4. `年份+月+干支日`：如 **`"二零二五年正月戊戌日"`**
5. `阿拉伯数字+月+日`：可以将中文数字替换为阿拉伯数字，如 **`"2025年1月1日"`**，代表`二零二五年正月初一`
6. 历史场景：历史上月份名称可能与现代不同（如武则天时期“正月”与“一月”代表不同月份），这类场景下月份名称按汉字数字解释

> ⚠️ 农历年份与公历年份并非完全重合。例如：公历2025年1月28日（除夕）对应农历2024年腊月二十九，对应的字符串是 `"二零二四年腊月廿九"`。

#### 方式二：传入数字参数
- **参数**：年份 (`int`)、月份 (`int`)、日期 (`int`)、是否闰月 (`bool`)
- **语义**：按农历年、月、日与闰月标志定位日期，适用于现代农历日期转换

### 代码示例

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

	// 示例1：公历转农历；这里故意选三国时期，会返回多个政权并行历法结果。
	date := time.Date(240, 1, 1, 8, 8, 8, 8, cst)
	lunar, _ := calendar.SolarToLunar(date)
	fmt.Println(lunar.LunarDescWithEmperor())

	info := lunar.LunarInfo()
	data, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println(string(data))

	// 示例2：农历转公历（字符串格式）；这里用苏轼《记承天寺夜游》的日期。
	solar, _ := calendar.LunarToSolar("元丰六年十月十二日")
	for _, v := range solar {
		fmt.Println(v.Time())
		fmt.Println(v.LunarDescWithEmperor())
	}

	// 示例3：农历转公历（数字参数格式）；2026 年正月初一，也就是春节。
	modernDate, _ := calendar.LunarToSolarByYMD(2026, 1, 1, false)
	fmt.Println(modernDate.Time())
}
```

输出结果：

```text
// 同一公历时刻在三国并立时期会映射到多个政权各自的农历结果
[魏明帝 景初三年腊月二十 蜀后主 延熙二年冬月十九 吴大帝 赤乌二年冬月二十]
// 结构化农历信息输出；每个对象对应一个政权口径下的结果
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
]
// “元丰六年十月十二日”对应的公历日期
1083-11-24 00:00:00 +0800 CST
// 同一天在并行政权下还会命中辽道宗大康九年十月十二
[宋神宗 元丰六年十月十二 辽道宗 大康九年十月十二]
// 现代农历日期转换结果；2026 年正月初一对应 2026-02-17
2026-02-17 00:00:00 +0800 CST
```

## 节气

`JieQi(year, term)` 返回现代天文算法计算出的节气精确时刻；`CalendricalJieQi(year, term)` 返回默认历法下节气落在的日期，时间固定为北京时间当天 0 点。需要指定古历系统时，使用 `CalendricalJieQiWithCalendar(year, term, system)`。

```go
package main

import (
	"fmt"

	"github.com/starainrt/astro/calendar"
)

func main() {
	// 计算 2020 年立春时刻；节气常量本质上对应太阳视黄经。
	fmt.Println(calendar.JieQi(2020, calendar.JQ_立春))
	// 计算 2020 年冬至时刻。
	fmt.Println(calendar.JieQi(2020, calendar.JQ_冬至))
	// 计算 2020 年春分时刻。
	fmt.Println(calendar.JieQi(2020, calendar.JQ_春分))
	// 也可直接传入黄经数值；春分对应太阳视黄经 0°。
	fmt.Println(calendar.JieQi(2020, 0))
}
```

输出结果

```
2020-02-04 17:03:20.471614301 +0800 CST
2020-12-21 18:02:20.648710727 +0800 CST
2020-03-20 11:49:37.149532735 +0800 CST
2020-03-20 11:49:37.149532735 +0800 CST

```

历法相符节气示例：

```go
date, err := calendar.CalendricalJieQi(1582, calendar.JQ_冬至)
fmt.Println(date, err)

date, err = calendar.CalendricalJieQiWithCalendar(-202, calendar.JQ_冬至, calendar.AncientCalendarQinHan)
fmt.Printf("%d-%02d-%02d %v\n", date.Year(), int(date.Month()), date.Day(), err)
```

输出结果

```
1582-12-22 00:00:00 +0800 CST <nil>
-202-12-25 <nil>
```

## 参数与返回值约定

### 单位与参数口径

- 年份沿用天文记年：`0` 表示公元前 1 年，`-1` 表示公元前 2 年；公历结果的支持范围是公元前 721 年到公元 3000 年。
- `Solar` 与 `Lunar` 的 `timezone` 参数单位是小时：`8.0` 是东八区、`7.0` 是东七区。`JieQi`、`WuHou`、`CalendricalJieQi*` 不接受时区参数，固定按北京时间（UTC+8）输出。
- 农历月是“与正月的距离”，正月为 1、二月为 2，依此类推；闰月由独立的 `leap`/`IsLeap` 标记，不从月号里推断。
- 农历年参数用公历年份代替，但岁首取农历岁首：己亥年腊月三十要传 `Solar(2019, 12, 30, false, 8)`，而不是 `Solar(2020, 12, 30, false, 8)`。
- `Date2JD` 只看 `time.Time` 的年月日时分秒字段，不做时区换算，所以同一民用日期的东八区 0 点与 UTC 0 点得到同一个儒略日；`JD2Date` 则把结果落在运行环境的本地时区。

### 时标

- 公开 API 的时刻一律按民用时刻处理，也就是把 `time.Time` 的字段直接当作 UTC 标签读取。库内把民用时刻换算到 TT 时，1972-01-01 之前按 UT1 处理，精确窗口内使用内置闰秒表；需要显式换算时用根包的 `astro.UT1FromUTC` / `astro.TTFromUTC` / `astro.DUT1`。
- `JieQi` 与 `WuHou` 的返回值是北京时间下的民用时刻，小数秒保留完整精度；`CalendricalJieQi*` 固定返回北京时间当天 0 点。
- `JD`、`Date2JD`、`NowJD` 都是同一个民用口径的儒略日浮点数：`2026-02-17` 的北京民用日对应 `2461088.5`。

### 零值与越界

- 公历 1582-10-05 到 1582-10-14 这 10 天不存在，相关接口返回错误，不会静默规范化到邻近日期。
- `SolarToLunar` / `SolarToLunarByYMD` 超出 `[-721,3000]` 时返回错误；`LunarToSolar*` 的公历结果同样限制在 `[-721,3000]`，边界农历年 `-722` 在结果仍落在范围内时可被接受。
- `CalendricalJieQiWithCalendar` 对春秋历和缺少历法节气资料的年份返回错误；`LunarToSolarWithCalendar` 对年号描述只在古历年号表命中时可用，否则报错。
- `SolarCandidates()` 在没有多候选时返回只含 `Solar()` 的单元素切片，不返回空切片；`Lunar()` 在没有候选时返回零值 `LunarTime`，对应的 `JulianOnly()` 返回 `false`。
- `LunarToSolarSingle` 已弃用，新代码用 `LunarToSolarByYMD`；`AncientCalendarDefault` 是零值，显式古历入口遇到它会退回默认实现。
- `LunarTime.CalendarSystem()` 与 `CalendarName()` 在默认路由和现代段返回空字符串，表示“没有显式古历”，不是错误。

### 精度与适用范围

- 现代段按现行农历 GB/T 33661-2017 编排，推荐年限是 `[1929,3000]`；`Solar`/`Lunar` 使用同一套定朔定气规则，只是允许自定义时区，所以推荐年限之外可能出现与史籍不同的日期。
- `[-103,1912]` 使用内置历表，`[-220,-104]` 使用秦汉颛顼历复原算法，`[-721,-221]` 按默认先秦古历重建。这三段都是复原结果，不保证与当时颁行的历日逐日一致。
- 定朔定气用现代天文算法计算，古代日期的朔与节气时刻与当时实测存在偏差，因此古代农历结果可能与史籍记载不同。
- 古历重建只到“历日”一级：`CalendricalJieQi*` 返回的是历法相符的日期，`JieQi` 返回的是现代天文算出的节气时刻，两者不可互换。

### 相关手册

节气的天文定义（太阳视黄经）与朔望判定见[太阳与月亮](sun-moon.md#日月位置)与[太阳与月亮](sun-moon.md#月相)；那里给出的是现代天文量，本手册给出的是历法编排结果。
