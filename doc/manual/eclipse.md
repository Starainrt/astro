# 日月食与凌日

[English](en/eclipse.md) | [返回 README](../../README.md)

`eclipse` 包提供日食与月食的全球搜索、站心接触时刻、中心路径、偏食足迹、贝塞尔根数与沙罗序列计算；`eclipse/svg` 生成站心视圆图、全球见食图、月食穿影图、详细版式以及水星/金星凌日图；`moon/limb` 提供可选的 Kaguya 月缘地形修正。

水星与金星凌日的接触时刻查询见[行星：水星与金星凌日](planets.md#水星与金星凌日)；GeoJSON / KML 导出见[天象地图、GeoJSON 与 KML](map-geojson.md)。

## 目录

- [基本示例](#基本示例)
- [主要功能](#主要功能)
  - [日食查询与中心路径](#日食查询与中心路径)
  - [月食查询与影半径模型](#月食查询与影半径模型)
  - [月缘地形修正](#月缘地形修正)
  - [全球见食图与月食出图](#全球见食图与月食出图)
  - [凌日出图](#凌日出图)
- [API 概览](#api-概览)
- [约定与限制](#约定与限制)

## 基本示例

计算 2009-07-22 上海洋山附近日全食的站心接触时刻，以及 2029-01-01 跨年月全食的各阶段时刻：

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/eclipse"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)

	// 站心日食：上海洋山附近（121.9850°E, 30.6167°N，椭球高 0 m）
	d2009 := time.Date(2009, 7, 22, 12, 0, 0, 0, cst)
	if s, ok := eclipse.LocalSolarEclipseOnDate(d2009, 121.9850, 30.6167, 0); ok {
		fmt.Printf("2009-07-22 洋山日食: 食型=%s  沙罗=%d (%d/%d)\n",
			s.Type, s.Saros.Series, s.Saros.Member, s.Saros.Count)
		fmt.Printf("  初亏 %s  食既 %s  食甚 %s  生光 %s  复圆 %s\n",
			s.PartialStart.Format("15:04:05"), s.CentralStart.Format("15:04:05"),
			s.GreatestEclipse.Format("15:04:05"), s.CentralEnd.Format("15:04:05"),
			s.PartialEnd.Format("15:04:05"))
		fmt.Printf("  全食时长 %s  食分 %.4f  食甚太阳高度 %.1f°\n",
			s.CentralEnd.Sub(s.CentralStart).Round(time.Second), s.Magnitude, s.SunAltitude)
	}

	// 月食：2029-01-01 跨年月全食
	d2029 := time.Date(2029, 1, 1, 12, 0, 0, 0, cst)
	if l, ok := eclipse.LunarEclipseOnDate(d2029); ok {
		fmt.Printf("2029-01-01 月食: 食型=%s  沙罗=%d (%d/%d)\n",
			l.Type, l.Saros.Series, l.Saros.Member, l.Saros.Count)
		fmt.Printf("  半影始 %s  初亏 %s  食既 %s  食甚 %s  生光 %s  复圆 %s  半影终 %s\n",
			l.PenumbralStart.Format("01-02 15:04"), l.PartialStart.Format("15:04"),
			l.TotalStart.Format("15:04"), l.Maximum.Format("01-02 15:04:05"),
			l.TotalEnd.Format("15:04"), l.PartialEnd.Format("15:04"), l.PenumbralEnd.Format("15:04"))
		fmt.Printf("  本影食分 %.3f  半影食分 %.3f\n", l.UmbralMagnitude, l.PenumbralMagnitude)
	}
}
```

输出：

```text
2009-07-22 洋山日食: 食型=total  沙罗=136 (37/71)
  初亏 08:23:55  食既 09:37:23  食甚 09:40:21  生光 09:43:20  复圆 11:03:14
  全食时长 5m57s  食分 1.0770  食甚太阳高度 57.3°
2029-01-01 月食: 食型=total  沙罗=125 (49/72)
  半影始 12-31 22:03  初亏 23:07  食既 00:16  食甚 01-01 00:52:05  生光 01:27  复圆 02:36  半影终 03:40
  本影食分 1.246  半影食分 2.274
```

`ok == false` 表示所查询日期或地点未命中日月食，此时不应读取结果字段。

## 主要功能

### 日食查询与中心路径

- **全球日食**：`SolarEclipseOnDate`、`Last/Next/ClosestSolarEclipse`；跨多年筛选轻量候选列表（不含地面几何）使用 `SolarEclipseCandidates(start, end, opt)`。
- **站心日食**：`LocalSolarEclipseOnDate`、`NextLocalSolarEclipse`、`NextLocalTotalSolarEclipse`、`NextLocalAnnularSolarEclipse`；带搜索预算上限的跨百年中心食检索使用 `SearchLocalCentralSolarEclipse`。
- **路径与足迹**：`SolarEclipseCentralPath` 返回中心线、南北界、食甚点与食带宽度；`SolarEclipsePartialFootprints` 返回偏食可见域、全球 P1–P4 / U1–U4 接触、可选食甚等时线（`GreatestTimeStep`）与瞬时本影轮廓（`CentralShadowStep`）。
- **贝塞尔根数**：`SolarEclipseBesselianElements(jdeTT, opt)` 返回三次多项式 `X`/`Y`/`D`/`L1`/`L2`/`Mu` 与 `TanF1`/`TanF2`；与已发布表核对 `Mu` 时先经 `SolarEclipseBesselianMuForPublishedTable` 换算恒星时基准。

```go
d2035 := time.Date(2035, 9, 2, 12, 0, 0, 0, cst)
if bj, ok := eclipse.LocalSolarEclipseOnDate(d2035, 116.4074, 39.9042, 0); ok {
	fmt.Printf("2035-09-02 北京日食: %s  食甚 %s  全食时长 %s  沙罗 %d\n",
		bj.Type, bj.GreatestEclipse.Format("15:04:05"),
		bj.CentralEnd.Sub(bj.CentralStart).Round(time.Second), bj.Saros.Series)
}
```

输出：

```text
2035-09-02 北京日食: total  食甚 08:33:38  全食时长 1m33s  沙罗 145
```

### 月食查询与影半径模型

- 全球月食：`LunarEclipseOnDate`、`Last/Next/ClosestLunarEclipse`。
- 站心月食：`LocalLunarEclipseOnDate`、`NextLocalLunarEclipse`、`NextLocalTotalLunarEclipse`；`Visibility` 区分 `full`、`moonrise`、`moonset`、`rise-and-set`、`interrupted`、`penumbra-moonrise`、`penumbra-moonset` 与 `invisible`。
- **地影放大模型**：默认采用 **Danjon**（与现代 NASA 月食目录口径一致）；需对照经典旧历表口径时，调用带 `Chauvenet` 后缀的入口（如 `ClosestLunarEclipseChauvenet`），其半影食分通常比 Danjon 大约 `0.025`，本影食分大约 `0.005`。

### 月缘地形修正

默认按光滑圆月面求解。导入独立子包 `github.com/starainrt/astro/moon/limb`（内嵌约 1.5 MB 的 Kaguya LALT 剖面，仅在导入时链接）后，可对接触时刻、中心带南北界与偏食外边界执行月缘地形修正（典型量级为秒级时长与公里级边界偏移）：

```go
import "github.com/starainrt/astro/moon/limb"

// 单次查询启用
info, ok := eclipse.LocalSolarEclipseOnDateWithOptions(date, lon, lat, 0,
	eclipse.SolarEclipseOptions{LunarLimb: limb.Default()})

// 挂载到 Context，统一作用于该上下文的日食、月食与月掩
base, _ := astro.NewContext(astro.ModelSemianalytic)
ctx := base.WithLunarLimb(limb.Default())
info, ok = ctx.Eclipse().LocalSolarEclipseOnDate(date, lon, lat, 0)
```

`eclipse/svg` 五个绘图选项均含 `LunarLimb` 字段，结果上的 `LunarLimbApplied` 记录实际是否生效，图注自动标明“圆月面”或“月缘剖面”。修正覆盖站心/全球日食接触与中心食时长、中心路径南北界与带宽、偏食外边界以及月食六个接触时刻；食甚时刻、食型、中心线、食分线与等时线保持圆月面口径。

### 全球见食图与月食出图

`eclipse/svg` 的五个入口均返回 `(string, bool)`，返回 `false` 表示该日无事件、画布低于下限或参数不合法：

| 入口 | 画面内容 | 建议画布 |
| --- | --- | --- |
| `LocalSolarEclipseSVG` | 站心日面视圆阶段图与接触时刻表 | 920×720 起 |
| `SolarEclipseMapSVG` | 全球见食图（等经纬、正射球面、南北极四种投影） | 1200×800（正射建议 1000×1414） |
| `LunarEclipseSVG` | 月球穿行地影示意图 | 960×620 起 |
| `LunarEclipseMapSVG` | 全球月食可见区图（默认含半影月出/月落带） | 1200×800 |
| `LunarEclipseDetailedSVG` | 穿影示意、坐标摘要与世界可见区组合版式 | 1000×1414 或 1414×1000 |

#### 站心日食视圆图

```go
svg, ok := eclipsesvg.LocalSolarEclipseSVG(
	time.Date(2009, 7, 22, 12, 0, 0, 0, cst),
	121.9850, 30.6167, 0,
	eclipsesvg.LocalSolarEclipseSVGOptions{Width: 920, Height: 720, Step: 5 * time.Minute, Location: cst},
)
```

![2009 长江口洋山日全食](../img/solar-eclipse-yangshan-2009.svg)
![2012 厦门日环食](../img/solar-eclipse-xiamen-2012.svg)
![2035 北京日全食](../img/solar-eclipse-beijing-2035.svg)

#### 全球日食图与投影切换

通过 `Projection` 在等经纬、正射球面（NASA 版式）、北极与南极方位等距投影间切换；`TimeLabelStep` 控制中心线时刻标注，`GreatestTimeStep` 启用食甚等时线，`MagnitudeValues: []float64{}` 关闭食分等值线：

```go
mapSVG, ok := eclipsesvg.SolarEclipseMapSVG(
	time.Date(2009, 7, 22, 12, 0, 0, 0, cst),
	eclipsesvg.SolarEclipseMapSVGOptions{
		Width: 1414, Height: 1000, Location: cst,
		TimeLabelStep: 30 * time.Minute, GreatestTimeStep: 30 * time.Minute,
	},
)

globeSVG, ok := eclipsesvg.SolarEclipseMapSVG(
	time.Date(2009, 7, 22, 12, 0, 0, 0, cst),
	eclipsesvg.SolarEclipseMapSVGOptions{
		Width: 1000, Height: 1414, Location: cst,
		Projection: eclipsesvg.EclipseMapProjectionOrthographic,
		TimeLabelStep: 30 * time.Minute, GreatestTimeStep: 30 * time.Minute,
	},
)

ut1SVG, ok := eclipsesvg.SolarEclipseMapSVG(
	time.Date(2009, 7, 22, 12, 0, 0, 0, cst),
	eclipsesvg.SolarEclipseMapSVGOptions{
		Width: 1200, Height: 800, Location: time.UTC,
		TimeScale: astro.TimeScaleUT1, TimeLabelStep: 30 * time.Minute,
	},
)
```

![2009 长江大日食全球见食图](../img/solar-eclipse-yangshan-2009-global.svg)
![2012 厦门日环食全球见食图](../img/solar-eclipse-xiamen-2012-global.svg)
![2035 北京日全食全球见食图](../img/solar-eclipse-beijing-2035-global.svg)
![2009 长江大日食正射球面图](../img/solar-eclipse-yangshan-2009-globe.svg)
![2012 日环食北极区全球见食图](../img/solar-eclipse-arctic-2012-global.svg)
![2021 南极日全食南极区全球见食图](../img/solar-eclipse-southpolar-2021-12-04.svg)
![2009 长江大日食全球见食图（UT1 口径）](../img/solar-eclipse-yangshan-2009-global-ut1.svg)

#### 月食穿影图、可见区图与详细版式

```go
d2029 := time.Date(2029, 1, 1, 0, 0, 0, 0, cst)
diagram, _ := eclipsesvg.LunarEclipseSVG(d2029, eclipsesvg.LunarEclipseSVGOptions{Width: 960, Height: 620, Step: 10 * time.Minute, Location: cst})
world, _ := eclipsesvg.LunarEclipseMapSVG(d2029, eclipsesvg.LunarEclipseMapSVGOptions{Width: 1200, Height: 800, Location: cst})
detailed, _ := eclipsesvg.LunarEclipseDetailedSVG(d2029, eclipsesvg.LunarEclipseDetailedSVGOptions{Location: cst})
```

世界可见区图默认绘制 `U1–U4` 地平边界与半影月出/月落带；设 `DisablePenumbralPhase: true` 可退回仅按 `P1`、食甚、`P4` 划分的四类可见区。

![2029 跨年月全食穿影图](../img/lunar-eclipse-2029-01-01.svg)
![2029 跨年月全食全球可见图](../img/lunar-eclipse-2029-01-01-global.svg)
![2029 跨年月全食详细版式](../img/lunar-eclipse-2029-01-01-detailed.svg)
![2026 月全食穿影图](../img/lunar-eclipse-2026-03-03.svg)
![2026 月全食详细版式](../img/lunar-eclipse-2026-03-03-detailed.svg)

### 凌日出图

`eclipsesvg.TransitMapSVG` 与 `eclipsesvg.TransitLocalSVG` 渲染水星（`TransitMercury`）和金星（`TransitVenus`）凌日的全球可见图与站心日面图：

```go
venusDate := time.Date(2012, 6, 6, 0, 0, 0, 0, time.UTC)
world, ok := eclipsesvg.TransitMapSVG(eclipsesvg.TransitVenus, venusDate, eclipsesvg.TransitMapSVGOptions{
	Isochrones:    eclipsesvg.TransitIsochroneExternal,
	IsochroneStep: 10 * time.Minute,
})
local, ok := eclipsesvg.TransitLocalSVG(eclipsesvg.TransitVenus, venusDate, 116.4, 39.9, 50,
	eclipsesvg.TransitLocalSVGOptions{Location: cst})
```

站心日面图上北左东，太阳位于视地平线下时轨迹改绘灰色虚线；对于地心无接触的边缘凌日（如 1937-05-11 水星凌日、2854-12-14 金星凌日），`TransitMapSVG` 自动绘制实际站心可见区。

![2032 水星凌日北京站心图](../img/transit-mercury-beijing-2032.svg)
![2032 水星凌日全球可见图](../img/transit-mercury-beijing-2032-global.svg)
![2012 金星凌日北京站心图](../img/transit-venus-beijing-2012.svg)
![2012 金星凌日全球可见图](../img/transit-venus-beijing-2012-global.svg)
![1937 水星凌日（地心无接触）全球可见图](../img/transit-mercury-1937-partial.svg)

## API 概览

| 分类 | 入口 | 说明 |
| --- | --- | --- |
| 日食查询 | `SolarEclipseOnDate`、`LocalSolarEclipseOnDate`、`Last/Next/Closest*SolarEclipse`、`SearchLocalCentralSolarEclipse`、`SolarEclipseCandidates` | 另有 `...WithOptions`（太阳半径档位、月缘剖面）与 `...IAUSingleK` |
| 日食几何 | `SolarEclipseCentralPath`、`SolarEclipsePartialFootprints`、`SolarEclipseBesselianElements`、`NewSolarEclipseShadowSolver` | 中心线、偏食足迹、贝塞尔根数与交互式单时刻阴影求解器 |
| 月食查询 | `LunarEclipseOnDate`、`LocalLunarEclipseOnDate`、`GeometricLocalLunarEclipseOnDate`、`Last/Next/Closest*LunarEclipse` | 默认 Danjon；`Danjon` / `Chauvenet` 后缀显式指定影半径模型 |
| UT1 转换 | `SolarEclipseInfoInUT1`、`LocalSolarEclipseInfoInUT1`、`LunarEclipseInfoInUT1`、`SolarEclipsePathInUT1` 等 | 将结果中的民用时刻换算为同物理瞬间的 UT1 读数 |
| SVG 出图 | `LocalSolarEclipseSVG`、`SolarEclipseMapSVG`、`LunarEclipseSVG`、`LunarEclipseMapSVG`、`LunarEclipseDetailedSVG`、`TransitMapSVG`、`TransitLocalSVG` | 位于 `github.com/starainrt/astro/eclipse/svg` |

## 约定与限制

- **日月半径口径**：日食默认采用 NASA Split-K 月球半径（半影 `k=0.2724880`，本影 `k=0.2722810`）与标准太阳半径（`959.639″` @ 1 AU）；通过 `SolarEclipseOptions.SunRadiusModel` 可切换至边缘档（`959.95″`），通过 `...IAUSingleK` 可切换至 IAU 单一 `k=0.2725076`。
- **与 NASA 目录对照**：在不含 ΔT 差异的 TT/TD 层，现代日食食甚中位差约 `0.4 s`、月食食甚在 `0.1 s` 量级。直接对比 UT 时刻时，NASA 出版页面采用出版时的预测 ΔT（如 2024/2026 年为 `74–75 s`），与实测 ΔT（约 `69.1 s`）存在 `5–6 s` 的时标口径差；可通过 `astro.SetDeltaT` 统一 ΔT 后比对。
- **沙罗编号**：优先匹配内置 NASA 权威锚点（`Saros.Verified == true`），`-3000..+6000` 年其余事件查预计算扩展表，范围外实时推算（`Verified == false`）。
- **画布下限与 UT1**：`SolarEclipseMapSVG` 下限 `800×560`（不足时回落至 `960×640`）；`PartialStep` 下限 2 分钟；`LunarEclipseDetailedSVG` 过小画布返回 `false`。设置 `TimeScale: astro.TimeScaleUT1` 时 `Location` 须为 `nil` 或 `time.UTC`，详见[时标声明](map-geojson.md#时标声明)。
