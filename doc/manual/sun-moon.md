# 太阳与月亮

[English](en/sun-moon.md) | [返回 README](../../README.md)

`sun` 与 `moon` 是主链，`lite/sun` 与 `lite/moon` 是独立近似实现。没有特殊说明时，角度单位为度，视直径与视半径为角秒，`sun.EarthDistance` 为 AU、`moon.EarthDistance` 为千米；观测接口的 `time.Time` 通常表示民用时刻，真太阳时的返回值另按地方太阳时解释，见[时标手册](timescale.md)。

本手册覆盖日月本体：位置、升落与中天、站心量、相位与朔望、近远地点与交点、最大赤纬、天平动、视直径与物理星历。日月食几何见[日食与月食手册](eclipse.md#全球见食图与月食出图)，月掩见[月掩手册](occultation.md#月掩出图)，站心几何与折射的通用换算见[坐标工具](coord.md)。

## 目录

- [日出日落与月相](#日出日落与月相)
- [API 参考](#api-参考)
  - [sun](#sun)
  - [moon](#moon)
  - [lite/sun](#litesun)
  - [lite/moon](#litemoon)
  - [截断项族](#截断项族)
- [常用场景](#常用场景)
  - [今天的日出日落与晨昏朦影](#今天的日出日落与晨昏朦影)
  - [月出月落与此刻的月亮高度](#月出月落与此刻的月亮高度)
  - [月相与下次朔望弦](#月相与下次朔望弦)
  - [视直径、地月距离与天平动](#视直径地月距离与天平动)
  - [地心量与站心量的区别](#地心量与站心量的区别)
  - [主链与 lite 的误差对照](#主链与-lite-的误差对照)
- [观测角语义](#观测角语义)
- [综合示例：日月升落与位置](#综合示例日月升落与位置)
  - [日出日落/月出月落](#日出日落月出月落)
  - [日月位置](#日月位置)
- [太阳](#太阳)
  - [位置](#位置)
  - [升落与中天](#升落与中天)
  - [站心量与视差角](#站心量与视差角)
  - [真太阳时与均时差](#真太阳时与均时差)
  - [物理与视直径](#物理与视直径)
  - [地球轨道极值](#地球轨道极值)
- [月亮](#月亮)
  - [位置](#位置-1)
  - [升落与中天](#升落与中天-1)
  - [月相](#月相)
  - [近地点与远地点](#近地点与远地点)
  - [交点](#交点)
  - [最大赤纬](#最大赤纬)
  - [天平动与亮边位置角](#天平动与亮边位置角)
  - [视直径与地月距离](#视直径与地月距离)
- [轻量链路](#轻量链路)
  - [lite/sun](#litesun-1)
  - [lite/moon](#litemoon-1)
  - [与主链的差异与误差量级](#与主链的差异与误差量级)
- [参数与返回值约定](#参数与返回值约定)
  - [单位与角口径](#单位与角口径)
  - [时标](#时标)
  - [高度与 aero](#高度与-aero)
  - [零值与越界](#零值与越界)
  - [精度与适用范围](#精度与适用范围)

## 日出日落与月相

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/sun"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2020, 1, 1, 8, 8, 8, 0, cst)
	lon, lat, height := 108.93, 34.27, 0.0
	rise, err := sun.RiseTime(date, lon, lat, height, true) //日出时间
	if err != nil {
		log.Fatal(err)
	}
	set, err := sun.SetTime(date, lon, lat, height, true) //日落时间
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(rise.Format(time.RFC3339), set.Format(time.RFC3339))
	fmt.Println(moon.Phase(date), moon.PhaseDesc(date)) //月相（被照亮比例）和月相描述
}
```

`aero=true` 采用含折射的升落判据。极昼、极夜或当天没有事件时，升落接口返回错误。`moon.Phase` 是受照比例，不是月龄。

## API 参考

下表按能力分组列出四个包的导出接口。多数求值接口另有 `...N` 截断变体，统一在[截断项族](#截断项族)里说明；表中所有 `time.Time` 参数都是绝对时刻，`lon`/`lat` 均为东正西负、北正南负。

### sun

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `TrueLo` / `ApparentLo` | 太阳真黄经 / 视黄经 | 度，地心 |
| `TrueBo` | 太阳真黄纬 | 度，地心；库内不单独提供视黄纬 |
| `GeometricLo` / `MidFunc` | 太阳几何黄经 / 中心差 | 度 |
| `ApparentRa` / `ApparentDec` / `ApparentRaDec` | 太阳视赤经 / 视赤纬 | 度，地心 |
| `EclipticObliquity` | 黄赤交角 | 度；第二参数为 `true` 时加入交角章动 |
| `EclipticNutation` / `EclipticNutation1980` | 黄经章动 | 度，IAU 2000B / IAU 1980 |
| `AxialtiltNutation` / `AxialtiltNutation1980` | 交角章动 | 度，IAU 2000B / IAU 1980 |
| `RiseTime` / `SetTime` | 日出 / 日落 | `(time.Time, error)`；`aero` 与 `height` 见参数与返回值约定 |
| `DownTime` | 日落别名 | 已废弃，内部转调 `SetTime` |
| `CulminationTime` | 上中天 | `time.Time` |
| `MorningTwilight` / `EveningTwilight` | 晨光始 / 暮光终 | `(time.Time, error)`；角度常用 -6 / -12 / -18 度 |
| `Altitude` / `Zenith` / `Azimuth` / `HourAngle` | 几何高度角 / 天顶距 / 方位角 / 时角 | 度，站心 |
| `ApparentAltitude` / `ApparentZenith` | 视高度角 / 视天顶距 | 度；需要气压 hPa 与气温 ℃ |
| `ParallacticAngle` | 视差角（天顶方向角） | 度，有符号 |
| `ApparentSolarTime` | 真太阳时 | `time.Time`，结果时区按经度换算 |
| `EquationTime` | 均时差 | 小时 |
| `Diameter` / `Semidiameter` | 视直径 / 视半径 | 角秒 |
| `EarthDistance` | 日地距离 | AU |
| `Physical` | 日面物理量 | 返回 `PhysicalInfo`，字段单位为度 |

### moon

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `TrueLo` / `TrueBo` / `ApparentLo` | 地心真黄经 / 真黄纬 / 视黄经 | 度 |
| `TrueRa` / `TrueDec` / `TrueRaDec` | 地心真赤道坐标 | 度 |
| `GeocentricApparentRa` / `GeocentricApparentDec` / `GeocentricApparentRaDec` | 地心视赤道坐标 | 度 |
| `ApparentRa` / `ApparentDec` / `ApparentRaDec` | 站心视赤道坐标 | 度；需要观测者经纬度 |
| `Altitude` / `Zenith` / `Azimuth` / `HourAngle` | 站心高度角 / 天顶距 / 方位角 / 时角 | 度 |
| `ApparentAltitude` / `ApparentZenith` | 视高度角 / 视天顶距 | 度；需要气压 hPa 与气温 ℃ |
| `ParallacticAngle` | 视差角 | 度，有符号，显式依赖观测者经纬度 |
| `RiseTime` / `SetTime` | 月出 / 月落 | `(time.Time, error)` |
| `DownTime` | 月落别名 | 已废弃，内部转调 `SetTime` |
| `CulminationTime` | 上中天 | `time.Time`；需要经度与纬度 |
| `Phase` / `PhaseDesc` | 受照比例 / 中文月相描述 | 比例 `[0,1]` / 字符串 |
| `SunMoonLoDiff` | 日月视黄经差 | 度，`[0,360)` |
| `ShuoYue` / `ShangXianYue` / `WangYue` / `XiaXianYue` | 以小数年为锚点的朔 / 上弦 / 望 / 下弦 | `time.Time`，UTC |
| `NewMoon` / `FullMoon` / `FirstQuarter` / `LastQuarter` | 上四相的英文 alias | `time.Time`，UTC |
| `Next*` / `Last*` / `Closest*` | 下一次 / 上一次 / 最近一次相位与最大赤纬 | `time.Time`；结果保持输入时区 |
| `NextConjunctionWithPlanet` / `LastConjunctionWithPlanet` / `ClosestConjunctionWithPlanet` | 行星合月（赤经合） | `time.Time`；目标用 `ConjunctionPlanet` 常量 |
| `PerigeesInMonth` / `ApogeesInMonth` | 指定年月内的近地点 / 远地点 | `[]ApsisInfo`，距离 km |
| `MaximumNorthDeclinationsInMonth` / `MaximumSouthDeclinationsInMonth` | 指定年月内的最大北 / 南赤纬事件 | `[]MaximumDeclinationInfo`，赤纬为度 |
| `AscendingNode` / `DescendingNode` | 升交点 / 降交点黄经 | 度 |
| `Physical` / `TopocentricPhysical` | 地心 / 站心天平动与自转轴位置角 | 返回 `PhysicalInfo`，字段单位为度 |
| `BrightLimbPositionAngle` / `TopocentricBrightLimbPositionAngle` | 地心 / 站心亮边位置角 | 度 |
| `Diameter` / `Semidiameter` | 视直径 / 视半径 | 角秒 |
| `EarthDistance` | 地月距离 | 千米 |

`moon` 还提供月掩接口，其参数、结果和完整示例见[月掩手册](occultation.md)。

| 分组 | 导出接口与类型 |
| --- | --- |
| 事件与路径 | `FindStarOccultations`, `FindPlanetOccultations`, `FindBestStarOccultations`, `FindBestPlanetOccultations`, `FindStarOccultationPaths`, `FindPlanetOccultationPaths` |
| 瞬时足迹与视圆几何 | `StarOccultationFootprintAt`, `PlanetOccultationFootprintsAt`, `StarOccultationDiagram`, `PlanetOccultationDiagram` |
| UT1 标签换算 | `StarOccultationInfoInUT1`, `StarOccultationPathInUT1`, `PlanetOccultationInfoInUT1`, `PlanetOccultationPathInUT1` |
| 恒星结果 | `StarOccultationInfo`, `StarOccultationPath`, `StarOccultationInstant` |
| 行星结果 | `PlanetOccultationInfo`, `PlanetOccultationPath`, `PlanetOccultationInstant`, `PlanetOccultationFootprint` |
| 路径数据 | `OccultationFootprint`, `OccultationPathPoint`, `OccultationGreatestTimeContour`, `OccultationRiseSetCurve` |
| 搜索与路径选项 | `OccultationSearchOptions`, `OccultationPathOptions`, `OccultationPathAlgorithm` |
| 视圆图数据 | `StarOccultationDiagramFrame`, `StarOccultationDiagramOptions`, `StarOccultationDiagramResult`, `PlanetOccultationDiagramFrame`, `PlanetOccultationDiagramOptions`, `PlanetOccultationDiagramResult` |
| 目标与坐标 | `StarData`, `StarCoordinate`, `StarCoordinateFromStarData`, `Observer`, `CoordinateFrame`, `CoordinateFrameICRS`, `CoordinateFrameJ2000`, `CoordinateFrameApparentOfDate` |
| 事件类型 | `OccultationPlanet`, `OccultationType`, `OccultationTotal`, `OccultationPartial`, `OccultationGrazing` |
| 路径算法 | `OccultationPathAlgorithmOptimized`, `OccultationPathAlgorithmExact` |
| 升落阶段 | `RiseSetPhase`, `RiseSetDirection`, `RiseSetPhaseStart`, `RiseSetPhaseGreatest`, `RiseSetPhaseEnd`, `RiseSetDirectionRise`, `RiseSetDirectionSet` |
| 错误 | `ErrInvalidOccultationInput`, `ErrOccultationPathSamplingLimit` |

行星目标常量为 `OccultationMercury` 至 `OccultationNeptune`。

### lite/sun

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `TrueLo` / `ApparentLo` | 轻量真黄经 / 视黄经 | 度，地心 |
| `TrueRa` / `TrueDec` / `TrueRaDec` | 轻量真赤道坐标 | 度，地心 |
| `ApparentRa` / `ApparentDec` / `ApparentRaDec` | 轻量视赤道坐标 | 度，地心 |
| `Distance` | 轻量日地距离 | AU |
| `HourAngle` / `Azimuth` / `Altitude` / `Zenith` | 轻量时角 / 方位角 / 高度角 / 天顶距 | 度，站心 |
| `RiseTime` / `SetTime` | 轻量日出 / 日落 | `(time.Time, error)` |
| `ERR_SUN_NEVER_RISE` / `ERR_SUN_NEVER_SET` | 极夜 / 极昼 | 错误值 |

### lite/moon

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `TrueLo` / `TrueBo` | 轻量地心真黄经 / 真黄纬 | 度 |
| `TrueRa` / `TrueDec` / `TrueRaDec` | 轻量地心真赤道坐标 | 度 |
| `ApparentRa` / `ApparentDec` / `ApparentRaDec` | 轻量站心视赤道坐标 | 度；需要观测者经纬度 |
| `HourAngle` / `Azimuth` / `Altitude` / `Zenith` | 轻量时角 / 方位角 / 高度角 / 天顶距 | 度，站心 |
| `SunMoonLoDiff` / `Phase` / `PhaseAge` | 轻量日月黄经差 / 受照比例 / 月龄 | 度 / `[0,1]` / 天 |
| `RiseTime` / `SetTime` | 轻量月出 / 月落 | `(time.Time, error)` |
| `ERR_MOON_NEVER_RISE` / `ERR_MOON_NEVER_SET` / `ERR_NOT_TODAY` | 极夜 / 极昼 / 事件不在当天 | 错误值 |

### 截断项族

`sun` 与 `moon` 的多数求值接口提供 `...N` 变体：名字是原接口名加 `N`，参数在末尾多一个 `n int`，返回形状不变。`n < 0` 使用当前仓库内嵌的全部解析项，结果与不带 `N` 的版本一致；`n >= 0` 把级数截断到 `n` 项，适合性能对比、批量粗算或误差敏感性实验。

- sun：`TrueLoN`、`TrueBoN`、`AltitudeN`、`ZenithN`、`AzimuthN`、`HourAngleN`、`ParallacticAngleN`、`ApparentAltitudeN`、`ApparentZenithN`、`DiameterN`、`SemidiameterN`、`PhysicalN`、`RiseTimeN`、`SetTimeN`、`DownTimeN`、`CulminationTimeN`、`MorningTwilightN`、`EveningTwilightN`、`ApparentSolarTimeN`
- moon：`TrueLoN`、`TrueBoN`、`AscendingNodeN`、`DescendingNodeN`、`DiameterN`、`SemidiameterN`、`PhysicalN`、`TopocentricPhysicalN`、`BrightLimbPositionAngleN`、`TopocentricBrightLimbPositionAngleN`

站心赤道坐标（`ApparentRa` / `ApparentDec` / `ApparentRaDec`）、月亮的升落（`RiseTime` / `SetTime`）与相位族没有 `N` 变体，它们不受截断开关控制。

## 常用场景

### 今天的日出日落与晨昏朦影

```go
fmt.Println(sun.MorningTwilight(date, lon, lat, -6)) // 民用晨光始
fmt.Println(sun.RiseTime(date, lon, lat, height, true))
fmt.Println(sun.SetTime(date, lon, lat, height, true))
fmt.Println(sun.EveningTwilight(date, lon, lat, -6)) // 民用暮光终
```

```text
2020-01-01 07:22:28.138198256 +0800 CST <nil>
2020-01-01 07:49:52.591398954 +0800 CST <nil>
2020-01-01 17:45:09.366609156 +0800 CST <nil>
2020-01-01 18:12:33.801986575 +0800 CST <nil>
```

朦影角度换成 `-12` / `-18` 就是航海与天文朦影；`aero = true` 按蒙气差与视半径修正后的地平求升落，与几何地平的差别见[升落与中天](#升落与中天)。

### 月出月落与此刻的月亮高度

```go
rise, _ := moon.RiseTime(date, lon, lat, height, true)
set, _ := moon.SetTime(date, lon, lat, height, true)
fmt.Println(rise)
fmt.Println(set)
fmt.Println(moon.Altitude(date, lon, lat), moon.Azimuth(date, lon, lat))
```

```text
2020-01-01 11:52:50.042243599 +0800 CST
2020-01-01 23:26:49.498263895 +0800 CST
-45.349728852972675 67.63824603392399
```

月球升落按当地自然日计算，升落之间可以没有连续性，`date` 之后的完整周期要看升起与落下时刻的先后关系，细节见[日出日落/月出月落](#日出日落月出月落)；高度角为负说明月亮在地平线下，这里 `-45.35°` 就是此刻看不见。

### 月相与下次朔望弦

```go
fmt.Println(moon.Phase(date), moon.PhaseDesc(date)) // 受照比例与月相名
fmt.Println(moon.NextShuoYue(date))                 // 下次朔
fmt.Println(moon.NextWangYue(date))                 // 下次望
```

```text
0.30004130960877884 上峨眉月
2020-01-25 05:41:58.271192908 +0800 CST
2020-01-11 03:21:17.159625291 +0800 CST
```

`Next*` / `Last*` / `Closest*` 是下一次 / 上一次 / 最近一次三种检索口径，上弦与下弦分别是 `moon.NextShangXianYue` / `moon.NextXiaXianYue`；结果保持输入时区，四种相位与朔望月的完整口径见[月相](#月相)。

### 视直径、地月距离与天平动

```go
fmt.Println(moon.Diameter(date), moon.EarthDistance(date))
p := moon.Physical(date)
fmt.Println(p.LibrationLongitude, p.LibrationLatitude, p.PositionAngle)
```

```text
1774.6658461637385 404238.6096080479
0.7655535663486027 6.382898400777244 -23.672356410246774
```

`Diameter` 单位角秒、`EarthDistance` 单位千米；`1774.67″`（约 29.6′）对应约 40.4 万 km 的远地点附近，比平均视直径小约 5%。`Physical` 给地心天平动，站心版本 `TopocentricPhysical` 与字段口径见[天平动与亮边位置角](#天平动与亮边位置角)。

### 地心量与站心量的区别

```go
geoRa, geoDec := moon.GeocentricApparentRaDec(date) // 地心视位置
topRa, topDec := moon.ApparentRaDec(date, lon, lat) // 站心视位置
fmt.Printf("geocentric %.4f %.4f\n", geoRa, geoDec)
fmt.Printf("topocentric %.4f %.4f\n", topRa, topDec)
fmt.Printf("delta dRA=%.4f dDec=%.4f\n", topRa-geoRa, topDec-geoDec)
fmt.Println(sun.ApparentRaDec(date)) // 太阳只给地心量
```

```text
geocentric 349.2322 -9.9506
topocentric 349.7343 -10.3485
delta dRA=0.5021 dDec=-0.3978
280.8950939694744 -23.05840775453492
```

月亮离得近，站心与地心能差半度左右（这里赤经差 `0.50°`、赤纬差 `0.40°`），合起来比一个月面视直径还大，所以给观测者用的位置必须走 `moon.ApparentRaDec`；太阳在 1 AU 外视差可忽略，只提供地心量（当然也可以调用站心坐标转换接口得到站心位置=-=）。口径定义见[观测角语义](#观测角语义)。

### 主链与 lite 的误差对照

```go
fmt.Println(sun.Altitude(date, lon, lat), litesun.Altitude(date, lon, lat)) // 主链 / 轻量几何高度角
fmt.Println(moon.Phase(date), litemoon.Phase(date))                         // 受照比例
fmt.Println(litemoon.PhaseAge(date))                                        // 轻量月龄（天）
fmt.Println(litesun.RiseTime(date, lon, lat, height, true))                 // 轻量日出
```

```text
2.40091496867759 2.403576774819768
0.30004130960877884 0.2978124633132848
5.42608394367707
2020-01-01 07:49:51.69717729 +0800 CST <nil>
```

轻量链路与主链形状同构、精度不同：同一时刻的高度角只差 `0.0027°` 量级，月相只差 `0.0022`；2026 全年 8 站点统计下 `lite/sun` 日出平均绝对误差 `0.02 min`（P95 `0.04 min`、最大 `0.31 min`），`lite/moon` 月出 `0.28 min`（P95 `0.57 min`、最大 `1.44 min`），`lite/moon` 的 `Phase()` 最大绝对误差 `0.00243`。

完整对照与截断误差见[与主链的差异与误差量级](#与主链的差异与误差量级)与 README 的 [Lite 轻量链路](accuracy.md#lite-轻量链路)。

## 观测角语义

- `Altitude`：高度角，地平线为 `0°`，天顶为 `+90°`
- `Zenith`：天顶距，天顶为 `0°`，地平线为 `90°`
- `Zenith` 与 `Altitude` 互补，两者相加为 `90°`
- `Azimuth`：方位角，正北为 `0°`，向东增加，取值范围 `[0°, 360°)`
- `HourAngle`：时角，上中天为 `0°`，向西（下午）增大，归一化到 `[0°, 360°)`
- `ParallacticAngle`：视差角（天顶方向角），有符号，单位度；方向定义与站心几何的统一口径见[坐标工具](coord.md)
- `Altitude` / `Azimuth` 是几何中心高度与方位，不含大气折射和视半径修正；`ApparentAltitude` / `ApparentZenith` 加入大气折射，需要传入气压（hPa）与气温（℃）
- 太阳的赤道坐标是地心量；月亮的 `TrueRaDec` 与 `GeocentricApparentRaDec` 是地心量，`ApparentRa` / `ApparentDec` / `ApparentRaDec` 是站心量，必须给出观测者经纬度

## 综合示例：日月升落与位置

下面两个完整示例覆盖最常用的入口，公共前置是西安（`108.93°E, 34.27°N`）与 `2020-01-01 08:08:08 CST`。后续各节的片段都省略这类公共变量，只保留与本组能力相关的语句。

### 日出日落/月出月落

> ⚠️ 月球升降时间按当天日期计算，升降时间点之间不一定具有连续性。
>
> 例如月亮可能在凌晨1点落下、中午12点再次升起，此时升起时间会晚于降落时间；这一场景晚上的月落时间对应次日日期。
>
> 完整的升降周期由升起时间与降落时间的先后关系决定：判断升起时间是否在降落时间之后，即可确定后续的正确时间点。

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/sun"
)

func main() {
	// 以陕西省西安市为例，设置西安市经纬度,设置地平高度为0米
	var lon, lat, height float64 = 108.93, 34.27, 0
	cst := time.FixedZone("CST", 8*3600)
	// 指定 2020-01-01 08:08:08 CST，所有"今日"语义都以这个本地自然日为基准。
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	// 西安市2020年1月1日民用晨朦影开始时间
	// 民用朦影，太阳位于地平线下6度，航海朦影=地平线下12度，天文朦影=地平线下18度
	fmt.Println(sun.MorningTwilight(date, lon, lat, -6))
	// 西安市2020年1月1日日出时间，按动态标准折射和实时太阳视半径计算上缘过地平线
	fmt.Println(sun.RiseTime(date, lon, lat, height, true))
	// 西安市2020年1月1日太阳上中天时间
	fmt.Println(sun.CulminationTime(date, lon))
	// 西安市2020年1月1日日落时间，按动态标准折射和实时太阳视半径计算上缘过地平线
	fmt.Println(sun.SetTime(date, lon, lat, height, true))
	// 西安市2020年1月1日民用昏朦影结束时间
	fmt.Println(sun.EveningTwilight(date, lon, lat, -6))

	// 西安市2020年1月1日月出时间，按动态标准折射和实时月球视半径计算上缘过地平线
	fmt.Println(moon.RiseTime(date, lon, lat, height, true))
	// 西安市2020年1月1日月亮上中天时间
	fmt.Println(moon.CulminationTime(date, lon, lat))
	// 西安市2020年1月1日月落时间，按动态标准折射和实时月球视半径计算上缘过地平线
	fmt.Println(moon.SetTime(date, lon, lat, height, true))
}
```

输出结果：

```text
2020-01-01 07:22:27.960488498 +0800 CST <nil>
2020-01-01 07:49:52.413689196 +0800 CST <nil>
2020-01-01 12:47:35.933117866 +0800 CST
2020-01-01 17:45:09.188657999 +0800 CST <nil>
2020-01-01 18:12:33.624035418 +0800 CST <nil>
2020-01-01 11:52:49.860912859 +0800 CST <nil>
2020-01-01 17:36:48.811488747 +0800 CST
2020-01-01 23:26:49.313553571 +0800 CST <nil>
```

### 日月位置

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/star"
	"github.com/starainrt/astro/sun"
	"github.com/starainrt/astro/tools"
)

func main() {
	// 以陕西省西安市为例，设置西安市经纬度,设置地平高度为0米
	var lon, lat float64 = 108.93, 34.27
	cst := time.FixedZone("CST", 8*3600)
	// 指定观测时刻。
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	// 太阳此刻的视黄经，单位度。
	fmt.Println(sun.ApparentLo(date))
	// 此刻黄赤交角，第二个参数 true 表示使用真黄赤交角。
	fmt.Println(sun.EclipticObliquity(date, true))
	//太阳此刻视赤经、视赤纬
	ra, dec := sun.ApparentRaDec(date)
	fmt.Println("赤经：", tools.Format(ra/15, 1), "赤纬：", tools.Format(dec, 0))
	//太阳当前所在星座
	fmt.Println(star.Constellation(ra, dec, date))
	//此刻西安市的太阳方位角、高度角、天顶距
	fmt.Println("方位角：", sun.Azimuth(date, lon, lat), "高度角：", sun.Altitude(date, lon, lat), "天顶距：", sun.Zenith(date, lon, lat))
	//此刻日地距离，单位为天文单位（AU）
	fmt.Println(sun.EarthDistance(date))

	//月亮此刻站心视赤经、视赤纬
	ra, dec = moon.ApparentRaDec(date, lon, lat)
	fmt.Println("赤经：", tools.Format(ra/15, 1), "赤纬：", tools.Format(dec, 0))
	//月亮当前所在星座
	fmt.Println(star.Constellation(ra, dec, date))
	//此刻西安市的月亮方位角、高度角、天顶距
	fmt.Println("方位角：", moon.Azimuth(date, lon, lat), "高度角：", moon.Altitude(date, lon, lat), "天顶距：", moon.Zenith(date, lon, lat))
	//此刻地月距离，单位为千米
	fmt.Println(moon.EarthDistance(date))
}
```

输出结果：

```text
280.01526210031136
23.4362178391013
赤经： 18h43m34.82s 赤纬： -23°3′30.27″
人马座
方位角： 120.19477090015224 高度角： 2.4014437419430097 天顶距： 87.59855625805699
0.983292937163176
赤经： 23h18m56.24s 赤纬： -10°20′54.42″
宝瓶座
方位角： 67.63889332004852 高度角： -45.34916937173283 天顶距： 135.34916937173284
404238.6096080479
```

## 太阳

本节片段省略公共前置：`cst := time.FixedZone("CST", 8*3600)`、`date := time.Date(2026, 1, 1, 12, 0, 0, 0, cst)`、`var lon, lat float64 = 108.93, 34.27`（西安），并假定已导入 `fmt`、`time`、`sun`、`moon`。

### 位置

太阳位置接口只依赖绝对时刻，不依赖观测者。真黄纬由 `TrueBo` 给出，库内不单独提供视黄纬；`EclipticObliquity` 的第二个参数决定是否加入交角章动。

```go
// 太阳真黄经、视黄经与真黄纬，单位度。
fmt.Println(sun.TrueLo(date), sun.ApparentLo(date), sun.TrueBo(date))
// 太阳视赤经与视赤纬。
ra, dec := sun.ApparentRaDec(date)
fmt.Println("视赤经：", ra, "视赤纬：", dec)
fmt.Println(sun.ApparentRa(date), sun.ApparentDec(date))
// 黄赤交角、几何黄经与中心差。
fmt.Println(sun.EclipticObliquity(date, true), sun.GeometricLo(date), sun.MidFunc(date))
```

输出结果：

```text
280.742671383543 280.7383965677222 0.00018280886212040676
视赤经： 281.6786097810291 视赤纬： -23.00369182413533
281.6786097810291 -23.00387403948847
23.438148552330773 280.83169623098 -0.08703499790144194
```

同一族接口都提供 `...N` 截断变体。下面的对照用 `n = 8` 截断，`n < 0` 时结果与不带 `N` 的版本逐位一致：

```go
// n<0 使用全部内嵌 VSOP 项，n>=0 截断解析项。
fmt.Println(sun.TrueLo(date), sun.TrueLoN(date, 8))
fmt.Println(sun.Altitude(date, lon, lat), sun.AltitudeN(date, lon, lat, 8))
fmt.Println(sun.Diameter(date), sun.DiameterN(date, 8))
```

输出结果：

```text
280.742671383543 280.7439900413756
31.61569462953789 31.615524473491835
1950.9979407481142 1950.9994358395434
```

### 升落与中天

`RiseTime` / `SetTime` 以 `date` 所在时区的当地自然日为锚点，返回值保持同一个时区；`height` 是按椭球高（大地高）解读的观测点高程，单位米。`aero` 为 `true` 时按动态标准大气折射与实时视半径计算**上缘**过地平线，为 `false` 时只做几何中心高度过地平线的判定。

朦影接口把目标高度角作为参数：民用朦影 `-6°`、航海朦影 `-12°`、天文朦影 `-18°`，晨昏两侧分别是 `MorningTwilight` 与 `EveningTwilight`。

```go
// 民用、航海、天文晨朦影的目标高度角。
for _, angle := range []float64{-6, -12, -18} {
	t, err := sun.MorningTwilight(date, lon, lat, angle)
	fmt.Println(angle, t.Format("15:04:05"), err)
}
// 上中天与日出；极区无事件时 err 非空。
fmt.Println(sun.CulminationTime(date, lon).Format("15:04:05"))
t, err := sun.RiseTime(date, lon, lat, 0, true)
fmt.Println(t.Format("15:04:05"), err)
```

输出结果：

```text
-6 07:22:47 <nil>
-12 06:51:31 <nil>
-18 06:21:00 <nil>
12:47:50
07:50:10 <nil>
```

### 站心量与视差角

`Altitude` / `Zenith` / `Azimuth` / `HourAngle` 走几何链路，不含折射；`ApparentAltitude` / `ApparentZenith` 加入大气折射，需要给气压与气温；`ParallacticAngle` 是有符号的视差角。站心几何与折射的更一般换算（含大气折射、站心赤道坐标）见[坐标工具](coord.md)。

```go
fmt.Println(sun.Azimuth(date, lon, lat), sun.Altitude(date, lon, lat), sun.Zenith(date, lon, lat))
fmt.Println(sun.ApparentAltitude(date, lon, lat, 1010, 10), sun.ApparentZenith(date, lon, lat, 1010, 10))
// 时角与有符号视差角。
fmt.Println(sun.HourAngle(date, lon, lat), sun.ParallacticAngle(date, lon, lat))
```

输出结果：

```text
167.09774780715728 31.61569462953789 58.38430537046211
31.643010360459822 58.356989639540174
348.07820699607544 -11.564174033740159
```

### 真太阳时与均时差

`ApparentSolarTime` 返回给定经度处的真太阳时，结果时区是按经度换算出来的固定偏移时区，不是调用方传入的时区；`EquationTime` 给出同一时刻的均时差，单位小时。`sundial.TrueSolarTime` 与 `ApparentSolarTime` 口径相同，`sundial` 另提供地方平太阳时与日晷几何，见[日晷与真太阳时](sundial.md)。

```go
// 真太阳时；结果时区按经度换算。
fmt.Println(sun.ApparentSolarTime(date, lon).Format("2006-01-02 15:04:05 -0700"))
// 均时差，单位小时。
fmt.Println(sun.EquationTime(date))
// sundial.TrueSolarTime 与其口径相同。
fmt.Println(sundial.TrueSolarTime(date, lon).Format("2006-01-02 15:04:05 -0700"))
```

输出结果：

```text
2026-01-01 11:12:18 +0715
-0.05674946079069686
2026-01-01 11:12:18 +0715
```

### 物理与视直径

`sun.Physical` 返回 `PhysicalInfo`：`P` 是太阳北极位置角，`B0` 是日面中心的太阳纬度，`L0` 是日面中心的卡林顿经度，单位都是度。`Diameter` / `Semidiameter` 给视直径与视半径，单位角秒；`EarthDistance` 给日地距离，单位 AU。

```go
// 视直径、视半径（角秒）与日地距离（AU）。
fmt.Println(sun.Diameter(date), sun.Semidiameter(date), sun.EarthDistance(date))
// 日面物理量 P/B0/L0，单位度。
p := sun.Physical(date)
fmt.Println(p.P, p.B0, p.L0)
```

输出结果：

```text
1950.9979407481142 975.4989703740571 0.9833237486528845
1.979086377118846 -3.0131029209723916 296.9595333604375
```

日、月与七大行星的同名接口形状一致，可以并排对照：

```go
fmt.Println(sun.Diameter(date), sun.Semidiameter(date))
fmt.Println(sun.Physical(date))
fmt.Println(moon.Diameter(date), moon.Semidiameter(date))
fmt.Println(mars.Diameter(date), mars.Semidiameter(date))
```

### 地球轨道极值

日地距离的极值由 `earth` 包给出，时间按 UTC，距离单位 AU；只关心某一时刻的轨道偏心率时直接调 `EarthEccentricity`：

```go
// 2026 年地球近日点、远日点，时间为 UTC，距离单位 AU。
peri := earth.Perihelion(2026)
aphe := earth.Aphelion(2026)
fmt.Printf("earth perihelion=%s distance=%.9fAU\n", peri.Time.Format(time.RFC3339), peri.Distance)
fmt.Printf("earth aphelion=%s distance=%.9fAU\n", aphe.Time.Format(time.RFC3339), aphe.Distance)
```

输出结果：

```text
earth perihelion=2026-01-03T17:15:35Z distance=0.983302050AU
earth aphelion=2026-07-06T17:31:24Z distance=1.016643936AU
```

```go
fmt.Printf("earth e=%.9f\n", earth.EarthEccentricity(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
```

## 月亮

本节的片段沿用太阳一节的公共前置，`date` 仍是 `2026-01-01 12:00:00 CST`、观测点是西安。

### 位置

月亮的赤道坐标分三层：`TrueRaDec` 是地心真位置，`GeocentricApparentRaDec` 是地心视位置，`ApparentRaDec` 是站心视位置。黄道侧只有地心量：`TrueLo` 真黄经、`TrueBo` 真黄纬、`ApparentLo` 视黄经。

```go
// 地心真赤道坐标。
fmt.Println(moon.TrueRaDec(date))
// 地心视赤道坐标。
fmt.Println(moon.GeocentricApparentRaDec(date))
// 站心视赤道坐标。
fmt.Println(moon.ApparentRaDec(date, lon, lat))
// 真黄经、真黄纬与视黄经。
fmt.Println(moon.TrueLo(date), moon.TrueBo(date), moon.ApparentLo(date))
```

输出结果：

```text
66.66309709020791 26.830370234236764
66.66476688311091 26.830608930005567
67.0275694157259 25.982665403390747
69.21422925147913 5.060516750865828 69.21574418700706
```

### 升落与中天

月球升落与太阳一样以当地自然日为锚点：`RiseTime` / `SetTime` 返回当天的事件时刻，`CulminationTime` 返回当天上中天时刻。月球的升降之间可以没有连续性，`date` 之后的完整周期要靠升起时间与降落时间的先后关系决定；当某个事件落在查询日期之外时，升落接口返回 `ERR_NOT_TODAY`。

```go
// 月出、上中天与月落，按当天日期计算。
rise, err := moon.RiseTime(date, lon, lat, 0, true)
fmt.Println(rise.Format("15:04:05"), err)
fmt.Println(moon.CulminationTime(date, lon, lat).Format("15:04:05"))
set, err := moon.SetTime(date, lon, lat, 0, true)
fmt.Println(set.Format("15:04:05"), err)
```

输出结果：

```text
16:17:11 <nil>
00:03:16
06:41:37 <nil>
```

### 月相

`Phase` 返回 `[0,1]` 的受照比例，`PhaseDesc` 返回中文月相名，`SunMoonLoDiff` 返回归一化到 `[0,360)` 的日月视黄经差（朔附近接近 `0°`、望附近接近 `180°`）。`Next*` / `Last*` / `Closest*` 是三种检索口径，结果保持输入时区；四个相位同时提供拼音名与英文 alias，例如 `ShuoYue` / `NewMoon`、`WangYue` / `FullMoon`、`ShangXianYue` / `FirstQuarter`、`XiaXianYue` / `LastQuarter`。

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/moon"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	// 指定观测时刻。
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	//月亮此刻被照亮的比例（月相）
	fmt.Println(moon.Phase(date))
	//月相具体描述
	fmt.Println(moon.PhaseDesc(date))
	//下次朔月时间；也可用 moon.NextNewMoon(date)
	fmt.Println(moon.NextShuoYue(date))
	//下次上弦月时间；也可用 moon.NextFirstQuarter(date)
	fmt.Println(moon.NextShangXianYue(date))
	//下次望月时间；也可用 moon.NextFullMoon(date)
	fmt.Println(moon.NextWangYue(date))
	//下次下弦月时间；也可用 moon.NextLastQuarter(date)
	fmt.Println(moon.NextXiaXianYue(date))
}
```

输出结果：

```text
0.30004130960877884 // 月面约有 30% 被太阳照亮
上峨眉月 // 当前月相描述
2020-01-25 05:41:58.271192908 +0800 CST // 下一次朔月
2020-01-03 12:45:23.229190707 +0800 CST // 下一次上弦
2020-01-11 03:21:17.159625291 +0800 CST // 下一次望月，也就是满月
2020-01-17 20:58:23.396406769 +0800 CST // 下一次下弦
```

`Last*`、`Closest*` 与按小数年锚点求解的 `ShuoYue` / `FullMoon` 等返回 UTC；`ClosestConjunctionWithPlanet` 求最近一次行星合月，目标用 `ConjunctionPlanet` 常量给出：

```go
// 朔与望的上一次、最近一次。
fmt.Println(moon.LastShuoYue(date), moon.ClosestShuoYue(date))
fmt.Println(moon.LastWangYue(date), moon.ClosestWangYue(date))
// 上弦与下弦。
fmt.Println(moon.LastFirstQuarter(date), moon.ClosestLastQuarter(date))
// 以小数年为锚点的解，结果为 UTC。
fmt.Println(moon.ShuoYue(2025.5).Format(time.RFC3339), moon.FullMoon(2025.5).Format(time.RFC3339))
// 最近一次行星合月（赤经合）。
fmt.Println(moon.ClosestConjunctionWithPlanet(date, moon.ConjunctionJupiter))
```

输出结果：

```text
2025-12-20 09:43:19.074603617 +0800 CST 2025-12-20 09:43:19.074603617 +0800 CST
2025-12-05 07:14:03.670351803 +0800 CST 2026-01-03 18:02:53.55531156 +0800 CST
2025-12-28 03:09:50.141303837 +0800 CST 2026-01-10 23:48:22.03346461 +0800 CST
2025-06-25T10:31:35Z 2025-07-10T20:36:46Z
2026-01-04 05:59:19.9425897 +0800 CST
```

### 近地点与远地点

`PerigeesInMonth` / `ApogeesInMonth` 返回指定公历月内的全部近地点与远地点事件，元素是 `ApsisInfo`，含 `Time`（UTC）与 `Distance`（km），一个月可能有零个、一个或多个事件。

```go
// 2026 年 1 月的月球近地点、远地点，距离单位 km。
perigees := moon.PerigeesInMonth(2026, time.January)
apogees := moon.ApogeesInMonth(2026, time.January)
fmt.Printf("moon perigee=%s distance=%.1fkm count=%d\n", perigees[0].Time.Format(time.RFC3339), perigees[0].Distance, len(perigees))
fmt.Printf("moon apogee=%s distance=%.1fkm count=%d\n", apogees[0].Time.Format(time.RFC3339), apogees[0].Distance, len(apogees))
```

输出结果：

```text
moon perigee=2026-01-01T21:44:24Z distance=360348.1km count=2
moon apogee=2026-01-13T20:47:13Z distance=405437.9km count=1
```

### 交点

月球也提供升交点和降交点黄经，适合做食季、轨道几何和月球轨道研究：

```go
nodeDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
fmt.Println(moon.AscendingNode(nodeDate), moon.DescendingNode(nodeDate))
```

这里的“升交点 / 降交点”与行星章节中的定义相同：

- `AscendingNode`：月球轨道从黄道南侧穿到黄道北侧时的黄经
- `DescendingNode`：月球轨道从黄道北侧穿到黄道南侧时的黄经
- 单位都是度；同一时刻两者通常相差约 `180°`

以上面 `nodeDate := 2026-01-01 00:00:00 UTC` 的示例来说，输出结果是：

```text
340.95708624505863 160.9570862450587
```

### 最大赤纬

月球赤纬在一个交点月内达到南北极值，`MaximumDeclinationInfo` 含 `Time`（事件时刻）与 `Declination`（该时刻的地心赤纬，度）。逐月接口返回当月全部事件，`Next*` / `Last*` / `Closest*` 按时刻检索：

```go
// 最近最大北赤纬与上一次最大南赤纬。
north := moon.ClosestMaximumNorthDeclination(date)
south := moon.LastMaximumSouthDeclination(date)
fmt.Println(north.Time.Format(time.RFC3339), north.Declination)
fmt.Println(south.Time.Format(time.RFC3339), south.Declination)
// 下一次最大北赤纬与当月全部事件。
fmt.Println(moon.NextMaximumNorthDeclination(date).Time.Format(time.RFC3339))
events := moon.MaximumNorthDeclinationsInMonth(2026, time.January)
fmt.Println(len(events))
for _, event := range events {
	fmt.Println(event.Time.Format(time.RFC3339), event.Declination)
}
```

输出结果：

```text
2026-01-02T16:10:49+08:00 28.266373428242343
2025-12-20T07:06:57+08:00 -28.23514705130737
2026-01-02T16:10:49+08:00
2 2026-01-02T08:10:49Z 28.266373428242343
```

同一组事件的逐月列表口径如下（`2026-01-01 00:00:00 UTC` 输入）：

```go
// 2026 年 1 月的月球最大北/南赤纬。
north := moon.MaximumNorthDeclinationsInMonth(2026, time.January)
south := moon.MaximumSouthDeclinationsInMonth(2026, time.January)
fmt.Printf("north=%s dec=%.6f\n", north[0].Time.Format(time.RFC3339), north[0].Declination)
fmt.Printf("south=%s dec=%.6f\n", south[0].Time.Format(time.RFC3339), south[0].Declination)
```

输出结果：

```text
north=2026-01-02T08:10:49Z dec=28.266373
south=2026-01-16T05:15:14Z dec=-28.304184
```

### 天平动与亮边位置角

`Physical` 返回地心天平动，`TopocentricPhysical` 返回站心天平动，二者都是 `PhysicalInfo`，含光学、物理与总天平动的经纬分量以及自转轴位置角 `PositionAngle`，单位度。亮边位置角的 `0°` 从月面北点起、向东增加。

```go
// 地心天平动与自转轴位置角。
p := moon.Physical(date)
fmt.Println(p.LibrationLongitude, p.LibrationLatitude, p.PositionAngle)
// 站心天平动与自转轴位置角。
topo := moon.TopocentricPhysical(date, lon, lat, 0)
fmt.Println(topo.LibrationLongitude, topo.LibrationLatitude, topo.PositionAngle)
// 地心与站心亮边位置角。
fmt.Println(moon.BrightLimbPositionAngle(date), moon.TopocentricBrightLimbPositionAngle(date, lon, lat, 0))
```

输出结果：

```text
-0.9680924808747591 -6.547834757841939 -9.025022841390472
-0.7780085060449551 -5.659649431558411 -8.883877708625544
269.08333384819935 267.8559531949645
```

站心量的完整示例（上海，`height = 4 m`）：

```go
// 月球天平动和自转轴位置角。
physical := moon.Physical(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
fmt.Printf("libration lon=%.6f lat=%.6f pa=%.6f\n", physical.LibrationLongitude, physical.LibrationLatitude, physical.PositionAngle)
// 月亮明亮边缘位置角；0° 从月面北点起，向东增加。
fmt.Printf("bright limb=%.6f\n", moon.BrightLimbPositionAngle(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
// 上海站心看到的月球天平动、自转轴位置角和亮边位置角。
topo := moon.TopocentricPhysical(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 121.4737, 31.2304, 4)
fmt.Printf("topo libration lon=%.6f lat=%.6f pa=%.6f\n", topo.LibrationLongitude, topo.LibrationLatitude, topo.PositionAngle)
fmt.Printf("topo bright limb=%.6f\n", moon.TopocentricBrightLimbPositionAngle(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 121.4737, 31.2304, 4))
```

输出结果：

```text
libration lon=-1.278902 lat=-6.531444 pa=-9.967050
bright limb=267.364849
topo libration lon=-1.736754 lat=-5.780730 pa=-10.072846
topo bright limb=266.038258
```

### 视直径与地月距离

`Diameter` / `Semidiameter` 给视直径与视半径（角秒），`EarthDistance` 给地月距离（千米）。三者都只依赖绝对时刻：

```go
// 视直径、视半径（角秒）与地月距离（千米）。
fmt.Println(moon.Diameter(date), moon.Semidiameter(date), moon.EarthDistance(date))
```

输出结果：

```text
1986.4975069969655 993.2487534984828 360488.4234539985
```

## 轻量链路

`lite/sun` 与 `lite/moon` 是独立于主链的近似实现：不依赖 VSOP87 或主链的 ELP2000/82 级数，面向 CPU / 内存受限环境，调用方式与主链同形。升落搜索用固定步长扫描加二分，不走主链的高精度章动迭代。

```go
package main

import (
	"fmt"
	"time"

	litemoon "github.com/starainrt/astro/lite/moon"
	litesun "github.com/starainrt/astro/lite/sun"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2026, 1, 1, 20, 0, 0, 0, cst)

	fmt.Println(litesun.Altitude(date, 121.4737, 31.2304))
	fmt.Println(litesun.RiseTime(date, 121.4737, 31.2304, 0, true))

	fmt.Println(litemoon.Phase(date))
	fmt.Println(litemoon.PhaseAge(date))
	fmt.Println(litemoon.RiseTime(date, 121.4737, 31.2304, 0, true))
}
```

下面的片段沿用这一节的公共前置：上海（`121.4737°E, 31.2304°N`）与 `2026-01-01 20:00:00 CST`。

### lite/sun

轻量太阳链路提供黄经、赤道坐标、距离、地平坐标与升落，但不提供视直径、日面物理量与朦影接口。`Distance` 是日地距离（AU）：需要注意，主链同名能力叫 `EarthDistance`，轻量链路叫 `Distance`。

```go
// 轻量太阳真黄经、视黄经与日地距离（AU）。
fmt.Println(litesun.TrueLo(date), litesun.ApparentLo(date), litesun.Distance(date))
// 轻量视赤经与视赤纬。
ra, dec := litesun.ApparentRaDec(date)
fmt.Println(ra, dec)
// 轻量时角、方位角、高度角与天顶距。
fmt.Println(litesun.HourAngle(date, 121.4737, 31.2304), litesun.Azimuth(date, 121.4737, 31.2304), litesun.Altitude(date, 121.4737, 31.2304), litesun.Zenith(date, 121.4737, 31.2304))
// 轻量升落。
fmt.Println(litesun.RiseTime(date, 121.4737, 31.2304, 0, true))
fmt.Println(litesun.SetTime(date, 121.4737, 31.2304, 0, true))
```

输出结果：

```text
281.0835889076667 281.0793650422643 0.9833163427233701
282.0475744673639 -22.973803828458102
120.58011725187828 263.4579582386061 -37.07676039520773 127.07676039520773
2026-01-01 06:52:24.6475178 +0800 CST <nil>
2026-01-01 17:02:44.014452695 +0800 CST <nil>
```

### lite/moon

轻量月球链路提供少量摄动项的月球位置、轻量站心修正、月相与月龄、升落，但不提供天平动、视直径、地月距离与交点。`PhaseAge` 返回月龄（天），只在轻量链路提供：

```go
// 轻量真黄经与真黄纬。
fmt.Println(litemoon.TrueLo(date), litemoon.TrueBo(date))
// 轻量地心真赤道坐标。
ra, dec := litemoon.TrueRaDec(date)
fmt.Println(ra, dec)
// 轻量站心视赤道坐标。
fmt.Println(litemoon.ApparentRaDec(date, 121.4737, 31.2304))
// 日月黄经差、受照比例与月龄。
fmt.Println(litemoon.SunMoonLoDiff(date), litemoon.Phase(date), litemoon.PhaseAge(date))
// 轻量地平坐标。
fmt.Println(litemoon.Altitude(date, 121.4737, 31.2304), litemoon.Azimuth(date, 121.4737, 31.2304), litemoon.Zenith(date, 121.4737, 31.2304))
```

输出结果：

```text
74.25630740893 5.078407838127742
72.2518040645064 27.549605139478064
72.74256784654426 27.432483413326057
153.17694236666568 0.9462021494002484 12.565014741082448
63.55513206820331 90.53047230027812 26.444867931796693
```

### 与主链的差异与误差量级

轻量链路与主链的差别集中在实现与误差上，接口形状基本一致。位置与月相等纯求值接口相对主链约快 `8.3–27.3x`，升落接口约 `1.0–3.7x`，计算链路零堆分配；升落搜索步长为 `lite/sun` `30` 分钟、`lite/moon` `15` 分钟。与 `sun` / `moon` 的误差（2026 全年，8 个站点）如下，数据见[适用范围与精度](accuracy.md#lite-轻量链路)：

| 能力 | 平均绝对误差 | P95 | 最大绝对误差 |
| --- | --- | --- | --- |
| `lite/sun` 日出 | `0.02 min` | `0.04 min` | `0.31 min` |
| `lite/sun` 日落 | `0.02 min` | `0.06 min` | `0.35 min` |
| `lite/moon` 月出 | `0.28 min` | `0.57 min` | `1.44 min` |
| `lite/moon` 月落 | `0.36 min` | `0.86 min` | `1.24 min` |
| `lite/moon` `Phase()` | `0.00089` | `0.00185` | `0.00243` |
| `lite/moon` `PhaseAge()` | `0.003 d` | `0.010 d` | `0.014 d` |
| `lite/moon` 地心黄经 | `2.41'` | `6.82'` | `9.91'` |
| `lite/moon` 地心黄纬 | `0.87'` | `1.83'` | `2.92'` |

两个包都不提供 `...N` 截断族。太阳侧的极夜 / 极昼错误是 `ERR_SUN_NEVER_RISE` / `ERR_SUN_NEVER_SET`；月球侧除 `ERR_MOON_NEVER_RISE` / `ERR_MOON_NEVER_SET` 外还有 `ERR_NOT_TODAY`，语义与主链一致。

## 参数与返回值约定

### 单位与角口径

- 角度一律为度，`RA`、`Lon`、`Azimuth` 归一化到 `[0°, 360°)`，赤纬与黄纬取值 `[−90°, 90°]`
- 视直径与视半径为角秒；`sun.EarthDistance` 为 AU，`moon.EarthDistance` 为千米
- `Phase` 是 `[0,1]` 的受照比例，`PhaseAge`（轻量链路）为天，`EquationTime` 为小时
- 距离类极值：`earth.Perihelion` / `earth.Aphelion` 为 AU，`ApsisInfo.Distance` 为千米

### 时标

观测输入与事件时刻采用民用时间约定；本库将 `1972-01-01` 以前的读数按 UT1 处理。`ApparentSolarTime` 另返回地方太阳时读数。`time.Time` 携带的时区只影响“当地自然日”的划分与返回值的时区，不改变绝对时刻。UT1 与图内时标的完整声明见[时标声明](map-geojson.md#时标声明)。

### 高度与 aero

- 升落接口的 `height` 是按椭球高（大地高）解读的观测点高程，单位米，不是正高；换算约定见[观测点高度约定](coord.md#观测点高度)
- `aero = true`：按动态标准大气折射与实时视半径计算上缘过地平线；`aero = false`：只做几何中心高度过地平线的判定
- `ApparentAltitude` / `ApparentZenith` 的 `pressureHPa`、`temperatureC` 是观测时的气压（hPa）与气温（℃）

### 零值与越界

- 极夜时升落接口返回 `ERR_SUN_NEVER_RISE` / `ERR_MOON_NEVER_RISE`；极昼时返回 `ERR_SUN_NEVER_SET` / `ERR_MOON_NEVER_SET`，此时第一个返回值为零值 `time.Time`
- 朦影不存在时返回 `ERR_TWILIGHT_NOT_EXISTS`
- 月亮升降事件落在查询日期之外时返回 `ERR_NOT_TODAY`，对应上面“月出月落按当天日期计算”的语义
- `DownTime` / `DownTimeN` 是 `SetTime` 的废弃别名，`ERR_SUN_NEVER_DOWN` / `ERR_MOON_NEVER_DOWN` 是极昼错误的废弃别名，新代码不应使用
- `...N` 中 `n < 0` 用全部内嵌项、`n >= 0` 截断；`n` 只影响级数长度，不改变返回单位与时区

### 精度与适用范围

太阳与行星用内置 VSOP87 解析项，月球用内置 ELP2000/82 风格截断级数，覆盖 J2000 前后约 4000 年，无需外部星历文件。截断误差、月球链路的能力边界和 lite 链路的量化误差见[适用范围与精度](accuracy.md)。本手册的太阳 / 月球接口适合历法、观测辅助、科普与业余预报；航天导航、精确掩星预报和严格动力学积分需要 JPL DE 等专业星历。
