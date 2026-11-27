# Astro

**[English](README.en.md) | 中文**

[![Go Reference](https://pkg.go.dev/badge/github.com/starainrt/astro.svg)](https://pkg.go.dev/github.com/starainrt/astro)

自用多年的天文算法库，用于个人天文历法爱好。

## 安装

```sh
go get github.com/starainrt/astro
```

## 功能概览

- 公历与农历互转，支持公元前 721 年至公元 3000 年；节气、干支、年号与古历。
- 太阳、月亮、七大行星的位置、升落、中天、距离、视直径与物理星历；另有轻量日月算法。
- 日月食、水星/金星凌日、恒星与行星月掩的事件搜索、地方接触时刻、全球路径与可选月缘地形修正。
- 通用天象搜索：黄经穿越、留、相位、最近角距、任意高度升落、近远日点、偕日可见与跨行星表面观测。
- 人造卫星 TLE / OMM 解析、SGP4/SDP4 传播、过境预报、凌日凌月与地面可见覆盖。
- 可切换内置星历模型、同一进程并行使用多个实例，支持 JPL SPK 多源叠加与可选摄动传播。
- SVG 天象图（日月食、月掩、木星卫星、人造卫星）、GeoJSON 地理数据与可供 Google Earth 按时间播放的 KML。
- 内置 9100 颗恒星，支持星座判定、坐标修正与升落计算；另有 CSV、Hipparcos、BSC5 与 Stellarium 外挂星表。
- 天球坐标转换、恒星时、岁差章动、折射、天文公式与日晷几何。

## 使用示例

计算 2026 年 2 月 17 日西安的日出日落、月相与农历日期：

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
	lon, lat := 108.93, 34.27

	rise, err := sun.RiseTime(date, lon, lat, 0, true)
	if err != nil {
		log.Fatal(err)
	}
	set, err := sun.SetTime(date, lon, lat, 0, true)
	if err != nil {
		log.Fatal(err)
	}
	day, err := calendar.SolarToLunar(date)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("西安 2026-02-17")
	fmt.Printf("日出 / 日落: %s / %s\n", rise.Format("15:04:05"), set.Format("15:04:05"))
	fmt.Printf("月相: %s（照亮 %.1f%%）\n", moon.PhaseDesc(date), moon.Phase(date)*100)
	fmt.Printf("农历: %d年%s（%s年 %s日）\n",
		day.Lunar().LunarYear(), day.Lunar().MonthDay(),
		day.Lunar().GanZhiYear(), day.Lunar().GanZhiDay())
}
```

输出：

```text
西安 2026-02-17
日出 / 日落: 07:27:16 / 18:29:40
月相: 残月（照亮 0.1%）
农历: 2026年正月初一（丙午年 壬戌日）
```

经纬度以东经、北纬为正，单位为度；高度参数为椭球高（米），`true` 表示升落计算计入大气折射与视半径。极昼、极夜或当天无升落事件时，升落接口返回错误。

包级函数使用进程当前的星历模型（默认构建为半解析模型）。如需固定模型或在同一进程内并行使用多套星历，使用 `astro.NewContext`：

```go
legacy, _ := astro.NewContext(astro.ModelLegacy)
modern, _ := astro.NewContext(astro.ModelSemianalytic)

riseL, _ := legacy.Sun().RiseTime(date, 108.93, 34.27, 0, true)
riseM, _ := modern.Sun().RiseTime(date, 108.93, 34.27, 0, true)
fmt.Printf("日出对比: 半解析=%s  Legacy=%s\n", riseM.Format("15:04:05"), riseL.Format("15:04:05"))
// 日出对比: 半解析=07:27:16  Legacy=07:27:16
```

上下文按模块提供 `Sun()`、`Moon()`、`Mars()`、`Eclipse()`、`Phenomena()`、`Satellite()` 等入口，方法与同名包级函数签名一致。自定义引擎、SPK 与多源叠加见[显式星历实例](doc/manual/ephemeris.md)。

## 包概览

| 包 | 用途与文档 |
| --- | --- |
| `calendar` | [历法转换、节气、干支与古历](doc/manual/calendar.md) |
| `sun` / `moon` / `lite/sun` / `lite/moon` | [太阳与月亮的位置、升落、月相和物理量](doc/manual/sun-moon.md) |
| `mercury` / `venus` / `mars` / `jupiter` / `jupiter/svg` / `saturn` / `uranus` / `neptune` / `earth` | [行星位置、特殊天象、木星卫星与土星环](doc/manual/planets.md) / [木星卫星 SVG](doc/manual/jupiter-svg.md) |
| `eclipse` / `eclipse/svg` | [日月食与凌日的查询、地方可见性、路径与出图](doc/manual/eclipse.md) |
| `moon` / `moon/svg` / `moon/limb` | [恒星及行星月掩、接触时刻、掩带与可选月缘地形修正](doc/manual/occultation.md) |
| `phenomena` | [黄经穿越、留、相位、最近角距、偕日可见、任意高度升落与跨行星表面观测](doc/manual/phenomena.md) |
| `satellite` | [TLE / OMM、SGP4 传播、过境、凌日凌月与可见范围](doc/manual/satellite.md) |
| `satellite/svg` | [卫星凌现象地面图、站心图与可见范围 SVG](doc/manual/satellite-svg.md) |
| `geojson` / `kml` | [天象地图、GeoJSON 与 KML](doc/manual/map-geojson.md) |
| `star` / `star/catalog` | [恒星数据库、星座、观测量与外挂星表](doc/manual/star.md#外挂星表) |
| `coord` | [坐标转换、恒星时、岁差、章动与折射](doc/manual/coord.md) |
| `orbit` | [小行星、彗星的二体或可选摄动传播与视双星](doc/manual/orbit.md) |
| `sundial` | [真太阳时与日晷几何](doc/manual/sundial.md) |
| `formula` | [辐射、星等、望远镜、大气折射、消光与地球姿态工具](doc/manual/formula.md) |
| `astro` / `basic` / `planet/spk` | [星历模型与实例、多源星历与自定义引擎](doc/manual/ephemeris.md) |
| `astro`（根包） | [UTC、UT1、TT 换算与 ΔT 模型](doc/manual/timescale.md) |

`basic` 为底层算法包，`planet` 保存内置模型与级数，`planet/spk` 读取外部核，`tools` 提供数值辅助函数。模型构成、实测误差与性能见[精度与性能](doc/manual/accuracy.md)；完整函数签名可查阅 [Go API 文档](https://pkg.go.dev/github.com/starainrt/astro)。

## 星历模型与时标约定

> **时标变更：2027-05-20 起 UTC 改用闰时**
>
> 依据第 28 届国际计量大会（CGPM）[第 5 号决议](https://www.bipm.org/en/cgpm-2026/resolution-5)（`Resolution-CGPM-28-5`），连续 UTC 自 2027-05-20 起生效，`|UT1−UTC|` 容限由 0.9 秒放宽至 3600 秒（闰时取代闰秒）；[IERS Bulletin C 72](https://datacenter.iers.org/data/latestVersion/bulletinC.txt) 亦确认 2026 年底无闰秒。本库自 v0.4.2 起保留 1972–2026 实测闰秒表不变，窗口之后的默认政策改为 `TimeScaleLeapHour`（零值）。`SetTimeScaleFuturePolicy` 的常量值已随之重排，直接传递裸整数的调用方须改用具名常量。详见[时标](doc/manual/timescale.md)。

- **内置星历**：默认使用**半解析模型**（`ModelSemianalytic`，基于 VSOP2013/TOP2013/ELP-MPP02 拟合至 DE440/DE441，无需外部文件；现代段地心视方向 RMS 约 `0.007″` 太阳、`0.076″` 月球）；需要复现早期数值时可切换至 **`ModelLegacy`**（VSOP87/ELP2000-82 级数）。`lite/sun` 与 `lite/moon` 为独立近似链路，不随模型切换。
- **民用时间与时标**：一般观测接口接收民用 `time.Time`，1972-01-01 之前按 UT1 处理，之后按 UTC 处理；SVG 与 GeoJSON 默认输出 UTC 标签，支持显式指定 UT1。
- **观测点高度**：观测点高程参数为 WGS84 **椭球高（米）**；若仅有海拔正高 `H`，须结合当地大地水准面差距 `N` 换算：`height = H + N`。详见[观测点高度](doc/manual/coord.md#观测点高度)。
- **适用范围**：内置星历面向历法、观测预报、科普与可视化；航天导航、精密掩星计时等用途需配套相应精度的外部 SPK、实测 ΔT/EOP 与地形模型。详见[精度与性能](doc/manual/accuracy.md)。

## 低内存构建（实验性）

`-tags astro_lowmem` 面向 RAM 受限平台（如 TinyGo 与 ESP32-C3）：星历表保留在只读存储中按需读取，计算结果与默认构建逐位一致，API 不变；代价是求值速度降低且二进制体积增大约 0.3 MB。桌面、服务端与 WASM 构建无需此标签。

查询内置 9100 星表（`StarDataBy*` 等）、启用月缘地形修正（`moon/limb`）以及切换至 `ModelLegacy`（首次解码需约 240 KB 堆内存）仍需较大内存，不宜在极低 RAM 平台上使用。

## 许可

代码与文档采用 [Apache-2.0](LICENSE)。仓库内嵌的第三方数据**不适用** Apache-2.0，各自遵循其来源条款，逐项声明见 [NOTICE](NOTICE)：

- **SELENE（KAGUYA）LALT 月缘剖面**（`moon/limb/kaguya_lalt.bin`）：ISAS/JAXA，按 ISAS 数据政策分发（兼容 CC BY 4.0）。
- **VSOP2013/TOP2013 与 ELP/MPP02 的再拟合表**（`planet/internal/semianalytic/`）：上游理论属 IMCCE / SYRTE，Apache-2.0 仅适用于本库代码与汇编。
- **JPL DE440/DE441 与 Horizons 校验数据**（`*/testdata/`、`planet/spk/testdata/`）：NASA/JPL。
- **内置星表的天测数据**（`basic/star_catalog.dat`）：源自 Yale Bright Star Catalogue（VizieR [V/50](https://cdsarc.cds.unistra.fr/viz-bin/cat/V/50)）。
- **Natural Earth 1:50m 物理陆地**（`internal/svgmap/land_*.path`）：公有领域。
