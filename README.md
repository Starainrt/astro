# Astro

**[English](README.en.md) | 中文**

[![Go Reference](https://pkg.go.dev/badge/github.com/starainrt/astro.svg)](https://pkg.go.dev/github.com/starainrt/astro)

自用多年的天文算法库，用于个人天文历法爱好。

基于 Jean Meeus《天文算法》（Astronomical Algorithms）实现。太阳和行星使用内置 VSOP87 解析项，月球使用 ELP2000/82 风格截断级数，不依赖外部星历文件。适合历法计算、业余观测和天文算法学习。

## 安装

```sh
go get github.com/starainrt/astro
```

## 功能概览

- 公历与农历互转，支持公元前 721 年至公元 3000 年；节气、干支、年号与古历。
- 太阳、月亮、七大行星的位置、升落、中天、距离、视直径与物理星历；另有轻量日月算法。
- 日月食、恒星与行星月掩的事件搜索、地方接触时刻和全球路径。
- SVG 天象图、GeoJSON 地理数据和可供 Google Earth 查看及按时间播放的 KML。
- 内置 9100 颗恒星，支持星座判定、坐标修正与升落计算。
- 天球坐标转换、恒星时、岁差章动、折射、通用小天体轨道、日晷和常用天文公式。

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

## 包概览

| 包 | 用途与文档 |
| --- | --- |
| `calendar` | [历法转换、节气、干支与古历](doc/manual/calendar.md) |
| `sun` / `moon` / `lite/sun` / `lite/moon` | [太阳与月亮的位置、升落、月相和物理量](doc/manual/sun-moon.md) |
| `mercury` / `venus` / `mars` / `jupiter` / `saturn` / `uranus` / `neptune` / `earth` | [行星位置、特殊天象、木星卫星与土星环](doc/manual/planets.md) |
| `eclipse` / `eclipse/svg` | [日月食查询、地方可见性、路径与出图](doc/manual/eclipse.md) |
| `moon` / `moon/svg` | [恒星及行星月掩、接触时刻与掩带](doc/manual/occultation.md) |
| `geojson` / `kml` | [天象地图、GeoJSON 与 KML](doc/manual/map-geojson.md) |
| `star` | [恒星数据库、星座与观测量](doc/manual/star.md) |
| `coord` | [坐标转换、恒星时、岁差、章动与折射](doc/manual/coord.md) |
| `orbit` | [小行星、彗星等二体轨道与视双星](doc/manual/orbit.md) |
| `sundial` | [真太阳时与日晷几何](doc/manual/sundial.md) |
| `formula` | [辐射、星等、会合周期与望远镜公式](doc/manual/formula.md) |
| `astro`（根包） | [UTC、UT1、TT 换算与 ΔT 模型](doc/manual/timescale.md) |

`basic` 是底层算法包，`planet` 保存解析级数，`tools` 提供数值辅助函数。一般使用上表中的包即可；完整函数签名也可查 [Go API 文档](https://pkg.go.dev/github.com/starainrt/astro)。

## 时标约定

一般观测接口接收表示民用时刻的 `time.Time`，内部处理 UTC、UT1 与 TT 换算。1972 年以前，库将民用时间按 UT1 处理，之后按UTC处理；轨道历元和部分底层接口另有时标要求。

角度默认用度，视直径用角秒，恒星时用小时，距离按接口使用 AU 或 km。SVG 和 GeoJSON 默认输出 UTC 时间标签，也可显式选择 UT1。具体约定见[时标手册](doc/manual/timescale.md)。

DUT1 与 TT−UTC 在实测窗口（截至 2026 年 9 月）之后默认沿用现行闰秒规则外推，可用 `SetTimeScaleFuturePolicy` 调整 TT−UTC 口径到"冻结偏移量"、"继续跟随 UT1"、"闰时规则"、"持续等价UT1"四个规则。

## 观测点高度约定

观测点高度为椭球高，单位米。若只有海拔（正高）`H`，应结合当地大地水准面差距 `N` 换算：`height = H + N`。详见[观测点高度](doc/manual/coord.md#观测点高度)。

## 适用范围与精度

本库使用解析模型和截断级数。精度随年代、天体与计算项目变化；专业掩星、航天导航等用途需要更精确的星历和物理模型。

`lite/sun`、`lite/moon` 适合资源受限环境；部分主接口也提供带 `N` 后缀的截断版本。模型范围、误差样本和性能数据见[精度与性能](doc/manual/accuracy.md)。
