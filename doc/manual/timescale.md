# 时标

[English](en/timescale.md) | [返回 README](../../README.md)

一般观测接口接收民用时刻 `time.Time`，库内部自动处理 UTC、UT1 与 TT 换算。本手册说明显式时标换算接口、ΔT 模型切换、2027 年后的 UTC 闰时政策，以及导出文件的时间标签约定。

## 目录

- [基本示例](#基本示例)
- [主要功能](#主要功能)
  - [比较与切换 ΔT 模型](#比较与切换-δt-模型)
  - [未来 UTC 闰时政策（2027-05-20 起）](#未来-utc-闰时政策2027-05-20-起)
  - [EarthUTCDate](#earthutcdate)
  - [SVG、GeoJSON 与 KML 的时间标签](#svggeojson-与-kml-的时间标签)
- [API 概览](#api-概览)
- [约定与限制](#约定与限制)

## 基本示例

将同一物理时刻在 UTC、TT、UT1 与 TDB 之间换算并读取 `DUT1` 与 `TDB − TT`：

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro"
)

func main() {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	tt := astro.TTFromUTC(date)
	ut1 := astro.UT1FromUTC(date)

	fmt.Printf("UTC:  %s\n", date.Format("2006-01-02 15:04:05.000"))
	fmt.Printf("TT:   %s  (TT-UTC = %.3f s)\n", tt.Format("2006-01-02 15:04:05.000"), tt.Sub(date).Seconds())
	fmt.Printf("UT1:  %s  (DUT1   = %+.3f s)\n", ut1.Format("2006-01-02 15:04:05.000"), astro.DUT1(date))
	fmt.Printf("TDB-TT: %+.3f ms\n", astro.TDBMinusTT(tt)*1000)
}
```

输出：

```text
UTC:  2026-04-01 00:00:00.000
TT:   2026-04-01 00:01:09.184  (TT-UTC = 69.184 s)
UT1:  2026-04-01 00:00:00.051  (DUT1   = +0.051 s)
TDB-TT: +1.630 ms
```

`TTFromUTC`、`UT1FromUTC` 与 `TDBFromTT` 返回的 `time.Time` 以 UTC Location 承载对应时标的读数，不可再将其当作民用时刻传回 `sun`、`moon` 等观测接口，否则会重复叠加时标差；还原回 UTC 应调用 `UTCFromTT` / `UTCFromUT1`。

## 主要功能

### 比较与切换 ΔT 模型

`ΔT = TT − UT1` 反映地球自转速率变化，直接影响远期日月食时刻与地面路径经度：

- 不改动全局状态比较模型取值：`astro.DeltaTModelSeconds(model, jd, keepObserved)`。
- 切换进程级 ΔT 模型：`astro.SetDeltaTModel(astro.DeltaTModelNASACanon2006, true)`（`keepObserved=true` 保留实测段，模型仅用于表外）。
- 注入自定义 ΔT 或 `TT − UTC` 回调：`astro.SetDeltaT(fn)` / `astro.SetTTMinusUTC(fn)`（传 `nil` 恢复内置默认）。

```go
jd2100 := basic.Date2JD(time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC))
fmt.Printf("2100年 ΔT 对比: SMH2016=%.1f s  Espenak-Meeus=%.1f s\n",
	astro.DeltaTModelSeconds(astro.DeltaTModelSMH2016, jd2100, false),
	astro.DeltaTModelSeconds(astro.DeltaTModelEspenakMeeus2006, jd2100, false))
```

输出：

```text
2100年 ΔT 对比: SMH2016=87.0 s  Espenak-Meeus=202.7 s
```

内置 ΔT 模型常量：

| 常量 | 说明 |
| --- | --- |
| `DeltaTModelDefault` / `DeltaTModelSMH2016` | 默认模型：实测覆盖期使用逐月表（截至 2026-10-01），表外采用 SMH2016 与 Morrison 2021 样条外推 |
| `DeltaTModelEspenakMeeus2006` | Espenak & Meeus 2006 分段多项式 |
| `DeltaTModelNASACanon2006` | Espenak–Meeus 多项式叠加 NASA 目录 1955 年前配对修正 |
| `DeltaTModelMS2004` | Morrison & Stephenson 2004 二次模型 |

ΔT、TT−UTC 与未来政策均为进程级设置，应在初始化阶段配置，不宜作为并发请求间的临时局部选项。

### 未来 UTC 闰时政策（2027-05-20 起）

依据第 28 届国际计量大会第 5 号决议（[Resolution-CGPM-28-5](https://www.bipm.org/en/cgpm-2026/resolution-5)），连续 UTC 自 2027-05-20 起生效，`|UT1−UTC|` 容限由 0.9 秒放宽至 3600 秒（闰时取代闰秒）；[IERS Bulletin C 72](https://datacenter.iers.org/data/latestVersion/bulletinC.txt) 亦确认 2026 年底无闰秒。

本库保留 1972–2026 实测闰秒表不变，窗口之后默认采用 `TimeScaleLeapHour`（TT−UTC 在数百年内维持 `69.184 s`）。通过 `astro.SetTimeScaleFuturePolicy` 可切换以下政策：

| 政策常量 | 实测窗口之后的计算假设 |
| --- | --- |
| `TimeScaleLeapHour`（默认、零值） | 闰时模式（`|DUT1| ≤ 3600 s`），对应 2027-05-20 起生效的连续 UTC |
| `TimeScaleLeapSecond` | 旧规则闰秒外推（`|DUT1| ≤ 0.9 s`），仅供复现与对照 |
| `TimeScaleAssumeUT1Tracking` | 保持末端 DUT1，TT−UTC 随外推 ΔT 平滑变化 |
| `TimeScaleFreezeUTCOffset` | 将 TT−UTC 冻结在窗口末端值（当前 `69.184 s`） |
| `TimeScaleUT1Civil` | 全时轴将民用时间视为 UT1（`DUT1 = 0`） |

### `EarthUTCDate`

`astro.EarthUTCDate(instant)` 返回该绝对时刻在 UTC 下的日期零点（`Location == time.UTC`，时分秒及纳秒为 0），适用于跨日界线时区（如 UTC+14）的日期归一化：

```go
z14 := time.FixedZone("UTC+14", 14*3600)
inst := time.Date(2026, 6, 4, 1, 30, 0, 0, z14)
fmt.Printf("EarthUTCDate(2026-06-04 01:30 UTC+14) = %s\n", astro.EarthUTCDate(inst).Format("2006-01-02"))
```

输出：

```text
EarthUTCDate(2026-06-04 01:30 UTC+14) = 2026-06-03
```

### SVG、GeoJSON 与 KML 的时间标签

SVG 与 GeoJSON 默认输出 UTC 时间标签。需要 UT1 时，可在选项中设置 `TimeScale: astro.TimeScaleUT1`（此时 `Location` 须为 `nil` 或 `time.UTC`），或调用 `eclipse` / `moon` 的 `...InUT1` 转换函数。KML `<when>` 规范要求 UTC，转换器会自动按当前 ΔT 将 UT1 标签换回 UTC 并保留原 `time_scale` 属性。详见[时标声明](map-geojson.md#时标声明)。

## API 概览

| 所属包 | 入口 | 用途 |
| --- | --- | --- |
| `astro`（`time.Time`） | `TTFromUTC` / `UTCFromTT` / `UT1FromUTC` / `UTCFromUT1` / `DUT1` | 民用 UTC 与 TT / UT1 互转及 DUT1 秒差 |
| `astro`（坐标时） | `TCGFromTT` / `TCBFromTT` / `TDBFromTT` / `TCBFromTDB`（及逆函数） | TT ↔ TCG / TCB / TDB 互转（参数为 TT 读数） |
| `astro`（微秒差） | `TCGMinusTT` / `TCBMinusTT` / `TDBMinusTT` | 直接返回秒差（避免单个 `float64` JD ~40 µs 舍入） |
| `astro`（全局策略） | `SetDeltaTModel` / `GetDeltaTModel` / `DeltaTModelSeconds` / `SetDeltaT` / `SetTTMinusUTC` / `SetTimeScaleFuturePolicy` | ΔT 模型与未来 UTC 政策配置 |
| `astro`（工具） | `LabelIn(scale, t)` / `EarthUTCDate(t)` | 时标标签转换与 UTC 零点截断 |
| `basic`（JD 数值） | `UTC2TT` / `TT2UTC` / `UT12TT` / `TT2UT1` / `UTC2UT1` / `UT12UTC` / `DeltaT` / `DUT1Seconds` / `TTMinusUTCSeconds` / `TT2TDB` 等 | 面向 `float64` 儒略日的底层换算 |

## 约定与限制

- **1972 年分界**：本库将 `1972-01-01` 之前的民用时刻按 UT1 处理，之后采用内置闰秒表与未来政策。
- **浮点 JD 与微秒差**：现代日期附近单个 `float64` JD 分辨率约为 40 µs，比较微秒级时标差值应直接调用 `*MinusTT` 或 `basic.*MinusTTSeconds`，不应将两个完整 JD 相减。`TDB − TT` 采用 40 项地心近似级数（`-3000..+6000` 年截断上界约 11 µs，不含测站日周项）。
- **闰秒表示限制**：Go 标准库 `time.Time` 不支持 `23:59:60`，跨闰秒阶跃区间不保证逐值往返恒等。若行星表面事件恰好落入历史正闰秒内且需严格 TT 时刻，应使用 `basic.SurfaceEvents` / `basic.NextSurfaceEvent`。
- **外部输入口径**：`orbit.Elements.EpochJD` / `TpJD` 为 TT/TDB 儒略日；`calendar.Date2JD` 直接读取钟面字段而不作时区换算。
