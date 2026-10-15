# Astro

**[English](README.en.md) | 中文**

[![Go Reference](https://pkg.go.dev/badge/github.com/starainrt/astro.svg)](https://pkg.go.dev/github.com/starainrt/astro)

自用多年的天文算法库，用于个人天文历法爱好。

> ### ⚠️ 时标变更：2027-05-20 起 UTC 改用闰时
>
> 2026 年 10 月 15 日，第 28 届国际计量大会（CGPM）通过[第 5 号决议《关于确保 UTC 连续性所需的技术行动》](https://www.bipm.org/en/cgpm-2026/resolution-5)（`Resolution-CGPM-28-5`，DOI `10.59161/CGPM2026RES5E`）：**连续 UTC 自 2027 年 5 月 20 日起生效**，`|UT1−UTC|` 的最大值由 0.9 秒放宽到 **3600 秒（1 小时）**，即以闰时取代闰秒。
>
> [IERS Bulletin C 72（2026-07-06）](https://datacenter.iers.org/data/latestVersion/bulletinC.txt) 已确认 **2026 年 12 月底不插入闰秒**；闰秒机制在 2027-05-20 终止，2027 年 6 月底这一可能的插入点同样不会有闰秒。
>
> 本库从本版本起：**1972–2026 仍按确定的闰秒表计算**，历史值不变；内置窗口（截至 2026-10-01）之后的默认政策由 `TimeScaleLeapSecond` 改为 **`TimeScaleLeapHour`**，窗口之后 TT−UTC 将长期保持 69.184 秒。`SetTimeScaleFuturePolicy` 的常量数值随之重排（`TimeScaleLeapHour` 现为零值），传裸整数的调用方需要更新。详见[时标](doc/manual/timescale.md)。

本库内置两套星历，可在同一进程内切换或并行使用：

1. **半解析模型**（`ModelSemianalytic`，默认）：以 VSOP2013/TOP2013（行星）和 ELP/MPP02（月球）为基准，按 DE 星历重新拟合——行星长时段用 DE441、现代段用 DE440，月球用 DE441——折叠成一张紧凑级数表。行星表约 452 KiB、月球表约 42 KiB，无需外部星历文件。现代段地心视方向均方根误差约 0.007″（太阳）、0.076″（月球）。
2. **旧模型**（`ModelLegacy`）：VSOP87/ELP2000-82 级数，另加长期项与逐天体订正。需要复现旧版数值时切换；现代段地心视方向均方根误差约 0.04″（太阳）、0.75″（月球），详见[精度与性能](doc/manual/accuracy.md)。

除内置模型外，也可以通过 SPK 或自定义引擎接入外部星历。

## 安装

```sh
go get github.com/starainrt/astro
```

## 功能概览

- 公历与农历互转，支持公元前 721 年至公元 3000 年；节气、干支、年号与古历。
- 太阳、月亮、七大行星的位置、升落、中天、距离、视直径与物理星历；另有轻量日月算法。
- 日月食、恒星与行星月掩的事件搜索、地方接触时刻和全球路径。
- 通用天象搜索：黄经穿越、留、相位、最近角距、任意高度升落、近远日点与偕日可见。
- 可切换星历模型、同一进程并行使用多个实例，支持 SPK 多源叠加与可选摄动传播。
- SVG 天象图、GeoJSON 地理数据和可供 Google Earth 查看及按时间播放的 KML。
- 内置 9100 颗恒星，支持星座判定、坐标修正与升落计算；另有 CSV、Hipparcos、BSC5 与 Stellarium 外挂星表。
- 天球坐标转换、恒星时、岁差章动、折射、天文公式、日晷几何与可选月缘地形修正。

## 使用示例

计算西安某日的日出时刻、月相，以及当天的农历日期：

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro/calendar"
	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/sun"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2026, 2, 17, 12, 0, 0, 0, cst)

	rise, err := sun.RiseTime(date, 108.93, 34.27, 0, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("日出：", rise.Format("15:04:05"))
	fmt.Println("月相：", moon.PhaseDesc(date))

	day, err := calendar.SolarToLunar(date)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("农历：", day.Lunar().MonthDay())
}
```

经纬度以东经、北纬为正，单位为度；示例高度为 0 米，`true` 表示升落计算考虑大气折射。极昼、极夜或当天没有升落事件时，升落接口会返回错误。

包级函数使用进程当前的星历模型（默认构建为半解析）。要给某次计算固定模型，用 `astro.NewContext`，同一进程里可以有多个上下文并行：

```go
func compareModels(date time.Time) {
	legacy, err := astro.NewContext(astro.ModelLegacy)
	if err != nil {
		panic(err)
	}
	modern, err := astro.NewContext(astro.ModelSemianalytic)
	if err != nil {
		panic(err)
	}

	fmt.Println(legacy.Moon().PhaseDesc(date), modern.Moon().PhaseDesc(date))

	riseLegacy, _ := legacy.Sun().RiseTime(date, 108.93, 34.27, 0, true)
	riseModern, _ := modern.Sun().RiseTime(date, 108.93, 34.27, 0, true)
	fmt.Println(riseLegacy.Format("15:04:05"), riseModern.Format("15:04:05"))
}
```

上下文按模块提供 `Sun()`、`Moon()`、`Mars()`、`Eclipse()`、`Phenomena()` 等入口，方法与同名包级函数签名一致。模型只隔离星历，ΔT、TT−UTC 和民用时标政策仍是进程级设置。自定义引擎、SPK 与多源叠加见[显式星历实例](doc/manual/ephemeris.md)。

## 包概览

| 包 | 用途与文档 |
| --- | --- |
| `calendar` | [历法转换、节气、干支与古历](doc/manual/calendar.md) |
| `sun` / `moon` / `lite/sun` / `lite/moon` | [太阳与月亮的位置、升落、月相和物理量](doc/manual/sun-moon.md) |
| `mercury` / `venus` / `mars` / `jupiter` / `saturn` / `uranus` / `neptune` / `earth` | [行星位置、特殊天象、木星卫星与土星环](doc/manual/planets.md) |
| `eclipse` / `eclipse/svg` | [日月食与凌日的查询、地方可见性、路径与出图](doc/manual/eclipse.md) |
| `moon` / `moon/svg` / `moon/limb` | [恒星及行星月掩、接触时刻、掩带与可选月缘地形修正](doc/manual/occultation.md) |
| `phenomena` | [任意两天体黄经穿越、留、相位、最近角距、偕日可见与任意高度升落](doc/manual/phenomena.md) |
| `geojson` / `kml` | [天象地图、GeoJSON 与 KML](doc/manual/map-geojson.md) |
| `star` / `star/catalog` | [恒星数据库、星座、观测量与外挂星表](doc/manual/star.md#外挂星表) |
| `coord` | [坐标转换、恒星时、岁差、章动与折射](doc/manual/coord.md) |
| `orbit` | [小行星、彗星的二体或可选摄动传播与视双星](doc/manual/orbit.md) |
| `sundial` | [真太阳时与日晷几何](doc/manual/sundial.md) |
| `formula` | [辐射、星等、望远镜、大气折射、消光与地球姿态工具](doc/manual/formula.md) |
| `astro` / `basic` / `planet/spk` | [星历模型与实例、多源星历与自定义引擎](doc/manual/ephemeris.md) |
| `astro`（根包） | [UTC、UT1、TT 换算与 ΔT 模型](doc/manual/timescale.md) |

`basic` 是底层算法包，`planet` 保存内置模型与级数，`planet/spk` 读取外部核，`tools` 提供数值辅助函数。模型构成、实测误差和性能见[精度与性能](doc/manual/accuracy.md)；完整函数签名也可查 [Go API 文档](https://pkg.go.dev/github.com/starainrt/astro)。

## 时标约定

一般观测接口接收表示民用时刻的 `time.Time`，内部处理 UTC、UT1 与 TT 换算。1972 年以前，库将民用时间按 UT1 处理，之后按 UTC 处理；轨道历元和部分底层接口另有时标要求。

角度默认用度，视直径用角秒，恒星时用小时，距离按接口使用 AU 或 km。SVG 和 GeoJSON 默认输出 UTC 时间标签，也可显式选择 UT1。具体约定见[时标手册](doc/manual/timescale.md)。

DUT1 与 TT−UTC 在月度 ΔT 表覆盖范围之后默认按**闰时**外推（`TimeScaleLeapHour`，对应 2027-05-20 起的连续 UTC），可用 `SetTimeScaleFuturePolicy` 在 `TimeScaleLeapHour`（默认）、`TimeScaleLeapSecond`、`TimeScaleAssumeUT1Tracking`、`TimeScaleFreezeUTCOffset`、`TimeScaleUT1Civil` 之间切换。

## 观测点高度约定

观测点高度为椭球高，单位米。若只有海拔（正高）`H`，应结合当地大地水准面差距 `N` 换算：`height = H + N`。详见[观测点高度](doc/manual/coord.md#观测点高度)。

## 适用范围与精度

精度随模型、年代、天体和计算项目变化。默认模型已从 VSOP87/ELP2000-82 换成半解析表，签名不变但数值可能不同；`lite/sun`、`lite/moon` 使用独立近似链路，不随模型切换。

模型构成、误差口径和实测结果见[精度与性能](doc/manual/accuracy.md)。专业掩星、航天导航等用途还需要相应精度的星历、时标与物理模型，接入 SPK 本身不保证全部高层结果达到核文件精度。

## 许可

代码与文档采用 [Apache-2.0](LICENSE)。仓库内嵌的第三方数据**不适用** Apache-2.0，各自遵循其来源条款，逐项声明见 [NOTICE](NOTICE)：

- **SELENE（KAGUYA）LALT 月缘剖面**（`moon/limb/kaguya_lalt.bin`）：ISAS/JAXA，按 ISAS 数据政策分发（该政策声明其规则与 CC BY 4.0 兼容）。
- **VSOP2013/TOP2013 与 ELP/MPP02 的自行再拟合表**（`planet/internal/semianalytic/`）：上游理论属 IMCCE / SYRTE。 Apache-2.0 只适用本库自己的代码与汇编。
- **JPL DE440/DE441 与 Horizons 数据**（`*/testdata/`、`planet/spk/testdata/`）：NASA/JPL。
- **内置星表的天测数据**（`basic/star_catalog.dat`）：源自 Yale Bright Star Catalogue（VizieR [V/50](https://cdsarc.cds.unistra.fr/viz-bin/cat/V/50)）。
- **Natural Earth 1:50m 物理陆地**（`internal/svgmap/land_*.path`）：公有领域。
