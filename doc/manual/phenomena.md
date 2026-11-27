# 通用天象搜索

[English](en/phenomena.md) | [返回 README](../../README.md)

`phenomena` 包提供不绑定特定天体包的通用天象搜索接口，涵盖黄经穿越、相位（合/冲/方照）、留、最近角距、轨道近远日点与交点、任意高度角穿越、偕日可见以及跨行星表面观测。

## 目录

- [基本示例](#基本示例)
- [主要功能](#主要功能)
  - [黄经穿越、相位与留](#黄经穿越相位与留)
  - [近日点、远日点与轨道交点](#近日点远日点与轨道交点)
  - [任意高度角穿越](#任意高度角穿越)
  - [偕日可见](#偕日可见)
  - [跨行星表面观测](#跨行星表面观测)
- [API 概览](#api-概览)
- [约定与限制](#约定与限制)

## 基本示例

搜索 2026 年春分、木星近月以及 2025–2028 年火星冲日：

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/phenomena"
)

func main() {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)

	// 太阳视黄经穿越 0°（春分）
	vernal := phenomena.LongitudeCrossings(phenomena.Sun, 0, start, end)
	fmt.Printf("2026 春分: %s UTC\n", vernal[0].Time.Format("2006-01-02 15:04:05"))

	// 木星与月球在 5° 以内的最近角距（2026 年 1–2 月）
	approaches := phenomena.ClosestApproaches(phenomena.Jupiter, phenomena.Moon, 5,
		start, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))
	for _, e := range approaches {
		fmt.Printf("木星近月: %s UTC  角距 %.2f°\n", e.Time.Format("2006-01-02 15:04:05"), e.Value)
	}

	// 火星冲日（火星与太阳黄经差 180°）
	opps := phenomena.AspectCrossings(phenomena.Mars, phenomena.Sun, 180,
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2028, 1, 1, 0, 0, 0, 0, time.UTC))
	for _, e := range opps {
		fmt.Printf("火星冲日: %s UTC\n", e.Time.Format("2006-01-02 15:04:05"))
	}
}
```

输出：

```text
2026 春分: 2026-03-20 14:45:57 UTC
木星近月: 2026-01-03 23:14:28 UTC  角距 3.61°
木星近月: 2026-01-31 03:39:08 UTC  角距 3.78°
木星近月: 2026-02-27 07:32:19 UTC  角距 3.91°
火星冲日: 2025-01-16 02:38:35 UTC
火星冲日: 2027-02-19 15:50:55 UTC
```

返回时刻保留 `start` 的时区；查询窗口内无命中事件时返回空切片。通过 `ctx.Phenomena()` 可绑定显式星历实例。

## 主要功能

### 黄经穿越、相位与留

- `LongitudeCrossings(body, deg, start, end)`：天体地心视黄经穿越指定角度（太阳每 15° 对应节气）。
- `AspectCrossings(a, b, angle, start, end)`：两天体视黄经差 `λ(a) - λ(b)` 穿越指定角度（`0°` 为合、`180°` 为冲、`90°`/`270°` 为弦或方照）。
- `Stations(body, start, end)`：行星视黄经速率过零的留，`Event.Value` 为留时视黄经（度）。
- `StarApproaches(body, star, maxSep, start, end)`：天体接近指定恒星；恒星坐标可通过 `StarCoordinateByName` / `StarCoordinateByHR` 从内置星表构造。

```go
start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
end := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

for _, s := range phenomena.Stations(phenomena.Mercury, start, end) {
	fmt.Printf("水星留: %s UTC（视黄经 %.2f°）\n", s.Time.Format("2006-01-02 15:04:05"), s.Value)
}
```

输出：

```text
水星留: 2026-02-26 06:48:10 UTC（视黄经 352.57°）
水星留: 2026-03-20 19:32:51 UTC（视黄经 338.49°）
```

### 近日点、远日点与轨道交点

- `Apsides(body, start, end)`：日心距离极值（`Perihelion` / `Aphelion`），支持 `Earth` 及七大行星。
- `Nodes(body, start, end)`：行星日心黄纬过零的升交点（`AscendingNode`）与降交点（`DescendingNode`）。

```go
y2026 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
y2027 := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)

for _, a := range phenomena.Apsides(phenomena.Earth, y2026, y2027) {
	label := "远日点"
	if a.Kind == phenomena.Perihelion {
		label = "近日点"
	}
	fmt.Printf("地球%s: %s UTC  日心距 %.6f AU\n", label, a.Time.Format("2006-01-02 15:04:05"), a.DistanceAU)
}
```

输出：

```text
地球近日点: 2026-01-03 17:15:43 UTC  日心距 0.983302 AU
地球远日点: 2026-07-06 17:30:19 UTC  日心距 1.016644 AU
```

### 任意高度角穿越

`AltitudeCrossing` 查询太阳、月球或行星在指定当地日内升、落穿越任意目标高度角的时刻（如 `-6°`/`-12°`/`-18°` 晨昏蒙影或自定义遮挡仰角）；恒星对应 `StarAltitudeCrossing`。

```go
cst := time.FixedZone("CST", 8*3600)
day := time.Date(2026, 6, 21, 0, 0, 0, 0, cst)

res, err := phenomena.AltitudeCrossing(phenomena.Sun, day, 116.4, 39.9, 0, -6, phenomena.AltitudeOptions{})
if err != nil {
	log.Fatal(err)
}
fmt.Printf("北京夏至民用晨昏 (-6°): 晨始 %s  昏终 %s\n",
	res.Rise.Format("15:04:05"), res.Set.Format("15:04:05"))
```

输出：

```text
北京夏至民用晨昏 (-6°): 晨始 04:11:26  昏终 20:20:51
```

`res.State` 区分 `AltitudeCrosses`、`AltitudeAlwaysAbove` 与 `AltitudeAlwaysBelow`，`HasRise` / `HasSet` 标识单侧穿越。默认选项按天体中心且不含折射，可通过 `AltitudeOptions` 指定 `Limb`（`LimbCenter`/`LimbUpper`/`LimbLower`）及 `Refraction: true`。

### 偕日可见

`PlanetHeliacalEvents` 与 `StarHeliacalEvents` 估算行星或恒星在晨昏微光中的首次可见与最后可见时刻：

```go
start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
end := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

name := map[phenomena.HeliacalEventKind]string{
	phenomena.HeliacalMorningFirst: "晨始见（偕日升）",
	phenomena.HeliacalMorningLast:  "晨末见",
	phenomena.HeliacalEveningFirst: "昏始见",
	phenomena.HeliacalEveningLast:  "昏末见（偕日落）",
}

for _, e := range phenomena.PlanetHeliacalEvents(phenomena.Mercury, start, end, 116.4, 39.9, phenomena.HeliacalOptions{}) {
	fmt.Printf("水星 %s: %s UTC  目标高度 %.1f°  太阳高度 %.1f°  余量 %.2f 等\n",
		name[e.Kind], e.Time.Format("2006-01-02 15:04"), e.TargetAltitude, e.SunAltitude, e.Margin)
}
```

输出：

```text
水星 昏始见: 2026-02-08 10:20 UTC  目标高度 3.7°  太阳高度 -8.0°  余量 0.10 等
水星 昏末见（偕日落）: 2026-02-25 10:50 UTC  目标高度 5.6°  太阳高度 -10.2°  余量 0.44 等
```

`Margin` 为天空极限星等与目标视星等之差；可通过 `HeliacalOptions.ExtinctionCoefficient` 指定 V 波段消光系数（可配合 `formula.ExtinctionSchaefer` 计算），通过 `VisualThreshold` 调整判据阈值。

### 跨行星表面观测

支持以水星至海王星任意行星表面（气态行星取 1 bar 参考层）为观测点，计算当地恒星时/太阳时、目标站心高度方位以及无折射升落与中天事件。以下以火星盖尔撞击坑（东经 `137.4°`，南纬 `4.6°`）为例：

```go
obs := phenomena.BodyObserver{
	Body:          phenomena.Mars,
	EastLongitude: 137.4, // 东经为正（度）
	Latitude:      -4.6,  // 大地纬度（度）
}
t := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)

lt, err := phenomena.SurfaceLocalTimeAt(obs, t)
if err != nil {
	log.Fatal(err)
}
sunRise, err := phenomena.NextSurfaceRise(phenomena.Sun, obs, t, phenomena.SurfaceOptions{})
if err != nil {
	log.Fatal(err)
}
earthPos, err := phenomena.SurfacePositionAt(phenomena.Earth, obs, sunRise.Time)
if err != nil {
	log.Fatal(err)
}

fmt.Printf("火星盖尔坑真太阳时: %.2f h  平太阳时: %.2f h  太阳日长: %.5f 地球日\n",
	lt.LocalTrueSolarTime, lt.LocalMeanSolarTime, lt.SolarDayLength)
fmt.Printf("下次火星日出: %s UTC（当地真太阳时 %.2f h，方位 %.1f°）\n",
	sunRise.Time.Format("2006-01-02 15:04:05"), sunRise.Local.LocalTrueSolarTime, sunRise.Position.Azimuth)
fmt.Printf("此时从火星看地球: 高度 %.1f°  方位 %.1f°  距离 %.3f AU\n",
	earthPos.Altitude, earthPos.Azimuth, earthPos.DistanceAU)
```

输出：

```text
火星盖尔坑真太阳时: 21.99 h  平太阳时: 22.43 h  太阳日长: 1.02749 地球日
下次火星日出: 2026-11-14 08:15:55 UTC（当地真太阳时 6.04 h，方位 80.9°）
此时从火星看地球: 高度 -35.7°  方位 66.7°  距离 1.322 AU
```

- 区间事件查询使用 `SurfaceEvents`；恒星目标使用 `StarSurfacePositionAt`、`StarSurfaceEvents` 及 `NextStarSurfaceRise/Set/Culmination`。
- 火星 MSD / MTC 与快捷日出日落入口见[行星本地时间与木星旋转角](planets.md#行星本地时间与木星旋转角)。
- 若输入为西经制图坐标或行星中心纬度，先经 `basic.PlanetographicToEastLongitude` / `basic.PlanetocentricToPlanetodeticLatitude` 转换。

## API 概览

| 分类 | 入口 | 说明 |
| --- | --- | --- |
| 天体常量 | `Sun`、`Moon`、`Mercury` … `Neptune`、`Earth` | `Earth` 仅适用于 `Apsides` 与跨行星观测目标 |
| 黄经 / 相位 / 留 | `LongitudeCrossings`、`AspectCrossings`、`Stations` | 返回 `[]Event`（`Time` 与角量 `Value`） |
| 最近角距 | `ClosestApproaches`、`StarApproaches` | 返回角距 `≤ maxSeparation`（度）的局部极小值 |
| 轨道事件 | `Apsides`、`Nodes` | 返回 `[]OrbitalEvent`（含日心距 `DistanceAU`） |
| 任意高度穿越 | `AltitudeCrossing`、`StarAltitudeCrossing` | 单日内穿过指定高度角的升落时刻与状态 |
| 偕日可见 | `PlanetHeliacalEvents`、`StarHeliacalEvents`、`LimitingMagnitude` | 晨始见/晨末见/昏始见/昏末见及极限星等估算 |
| 跨行星表面观测 | `SurfacePositionAt`、`SurfaceLocalTimeAt`、`SurfaceEvents`、`NextSurfaceRise/Set/Culmination` | 恒星版本带 `Star` 前缀；观测行星为水星至海王星 |
| 恒星构造 | `StarCoordinateByName`、`StarCoordinateByHR`、`StarCoordinateFromStarData` | 由内置星表构造 `StarCoordinate` |

## 约定与限制

- **视位置与视差**：黄经穿越、相位、留与最近角距采用地心视位置。`AltitudeCrossing` 仅对月球计入站心视差；需扣除地平俯角时由调用方预先调整目标高度角。
- **偕日可见**：基于经验天空亮度与消光公式（默认消光系数 `0.25 mag/airmass`），实际可见性受局地气象、光污染与观测条件影响，仅作窗口规划参考。
- **跨行星表面观测**：
  - 均为无折射几何模型（不含大气折射、地形遮挡与自转固有时）；`SurfaceOptions.UpperLimb`（默认 `true`）仅对太阳生效，其余天体与恒星按中心判定。
  - `LocalTrueSolarTime` 与 `LocalMeanSolarTime` 的小时为当地太阳日的 1/24，`SolarDayLength` 的单位为地球日；`LocalSiderealTime` 与 `SiderealRotationAngle` 单位为**度**。
  - 行星自转姿态采用 IAU 报告线性模型，近代约 ±500 年内最可靠（`Local.ModelValid` 标识有效年代）。水星轨道倾角接近零导致春分点退化；极昼/极夜超出扫描上限时 `NextSurfaceRise/Set` 返回 `ErrSurfaceEventNotFound`。
  - 若事件 TT 时刻落在历史正闰秒（`23:59:60`）内，`time.Time` 无法精确表示该秒；需要严格 TT 时刻请使用 `basic.SurfaceEvents` / `basic.NextSurfaceEvent`。
