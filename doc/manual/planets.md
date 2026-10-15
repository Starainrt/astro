# 行星

[English](en/planets.md) | [返回 README](../../README.md)

七大行星各对应一个同名包：`mercury`、`venus`、`mars`、`jupiter`、`saturn`、`uranus`、`neptune`。它们的公开接口按同一套形状组织：以 `time.Time` 传入民用时刻，以 `float64` 返回角度或距离，事件搜索返回 `time.Time` 或结构体。内行星（水星、金星）额外提供上合/下合、大距与地心、站心凌日；外行星（火星到海王星）额外提供冲日与方照；木星独有伽利略卫星，土星独有土星环参数。

低层内置半解析模型与 Legacy 的 VSOP87/日月级数放在 `planet` 包，被这七个包与 `sun` / `moon` 共用；默认构建使用半解析模型，可用显式 Context 固定模型。

- 七个包公共能力的函数名一致（例如都提供 `ApparentRa`、`ApparentDec` 与 `ApparentRaDec`），差异只体现在各自额外的那几族接口上；调用时必须带对应包名前缀，不能跨包混用。
- 位置接口给出的是**地心视位置**；地平量单独提供：`Altitude` / `Azimuth` / `Zenith` 是站心量，`HourAngle` / `ParallacticAngle` 是地心量（不含站心视差），公式入口见[坐标工具](coord.md)。
- 事件搜索族都成对出现（`Last...` / `Next...`，部分另有 `Closest...`），一律取"当前或之前/之后最近一次"并包含端点。
- 行星本身没有独立的 SVG 出图入口；月掩相关出图（含土星环月掩）见[月掩手册](occultation.md#月掩出图)。

## 目录

- [火星的位置与升起时刻](#火星的位置与升起时刻)
- [API 参考](#api-参考)
  - [按行星横向对照](#按行星横向对照)
  - [通用能力与单位](#通用能力与单位)
  - [`...N` 族](#n-族)
- [常用场景](#常用场景)
  - [今晚能看到哪颗行星](#今晚能看到哪颗行星)
  - [冲日、合日、大距、留与逆行](#冲日合日大距留与逆行)
  - [水星与金星凌日](#水星与金星凌日)
  - [相位、视直径、视星等与节点](#相位视直径视星等与节点)
  - [物理星历与木星伽利略卫星](#物理星历与木星伽利略卫星)
  - [与外部资料的对照口径](#与外部资料的对照口径)
- [基础示例](#基础示例)
  - [内行星](#内行星)
  - [外行星](#外行星)
- [分主题示例](#分主题示例)
  - [位置与坐标](#位置与坐标)
  - [升落与中天](#升落与中天)
  - [合冲留与方照](#合冲留与方照)
  - [大距与地心凌日](#大距与地心凌日)
  - [节点、相位、视星等、视直径与视差角](#节点相位视星等视直径与视差角)
  - [与其他手册的分工](#与其他手册的分工)
  - [物理星历](#物理星历)
  - [木星伽利略卫星](#木星伽利略卫星)
  - [与 `planet` 包共用的低层入口](#与-planet-包共用的低层入口)
- [参数与返回值约定](#参数与返回值约定)
  - [时标与民用时刻](#时标与民用时刻)
  - [单位与口径](#单位与口径)
  - [零值、越界与哨兵错误](#零值越界与哨兵错误)
  - [站心与地心](#站心与地心)
  - [精度与适用范围](#精度与适用范围)

## 火星的位置与升起时刻

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro/mars"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2020, 1, 1, 8, 8, 8, 0, cst)
	lon, lat, height := 108.93, 34.27, 0.0
	ra, dec := mars.ApparentRaDec(date)
	fmt.Printf("RA=%.6f Dec=%.6f deg\n", ra, dec)
	rise, err := mars.RiseTime(date, lon, lat, height, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(rise.Format(time.RFC3339))
}
```

`ApparentRaDec` 返回地心视赤经、视赤纬，单位度。升落接口还需要观测地和椭球高；没有升起事件时返回错误。

## API 参考

### 按行星横向对照

下表按能力对照七个行星包。调用时加包名前缀，例如 `mars.NextOpposition`；`/` 分隔并列函数，`—` 表示该包没有对应接口。

| 计算项 | `mercury` | `venus` | `mars` | `jupiter` | `saturn` | `uranus` | `neptune` |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 视黄经 / 视黄纬 | `ApparentLo` / `ApparentBo` | `ApparentLo` / `ApparentBo` | `ApparentLo` / `ApparentBo` | `ApparentLo` / `ApparentBo` | `ApparentLo` / `ApparentBo` | `ApparentLo` / `ApparentBo` | `ApparentLo` / `ApparentBo` |
| 视赤经 / 视赤纬 | `ApparentRa` / `ApparentDec` / `ApparentRaDec` | `ApparentRa` / `ApparentDec` / `ApparentRaDec` | `ApparentRa` / `ApparentDec` / `ApparentRaDec` | `ApparentRa` / `ApparentDec` / `ApparentRaDec` | `ApparentRa` / `ApparentDec` / `ApparentRaDec` | `ApparentRa` / `ApparentDec` / `ApparentRaDec` | `ApparentRa` / `ApparentDec` / `ApparentRaDec` |
| 视星等 | `ApparentMagnitude` | `ApparentMagnitude` | `ApparentMagnitude` | `ApparentMagnitude` | `ApparentMagnitude` | `ApparentMagnitude` | `ApparentMagnitude` |
| 地心距 / 日心距 | `EarthDistance` / `SunDistance` | `EarthDistance` / `SunDistance` | `EarthDistance` / `SunDistance` | `EarthDistance` / `SunDistance` | `EarthDistance` / `SunDistance` | `EarthDistance` / `SunDistance` | `EarthDistance` / `SunDistance` |
| 轨道升交点 / 降交点 | `AscendingNode` / `DescendingNode` | `AscendingNode` / `DescendingNode` | `AscendingNode` / `DescendingNode` | `AscendingNode` / `DescendingNode` | `AscendingNode` / `DescendingNode` | `AscendingNode` / `DescendingNode` | `AscendingNode` / `DescendingNode` |
| 视直径 / 视半径 | `Diameter` / `Semidiameter` | `Diameter` / `Semidiameter` | `Diameter` / `Semidiameter` | `Diameter` / `Semidiameter` | `Diameter` / `Semidiameter` | `Diameter` / `Semidiameter` | `Diameter` / `Semidiameter` |
| 相位角 / 照亮比例 / 亮面位置角 | `PhaseAngle` / `Phase` / `IlluminatedFraction` / `BrightLimbPositionAngle` | `PhaseAngle` / `Phase` / `IlluminatedFraction` / `BrightLimbPositionAngle` | `PhaseAngle` / `Phase` / `IlluminatedFraction` / `BrightLimbPositionAngle` | `PhaseAngle` / `Phase` / `IlluminatedFraction` / `BrightLimbPositionAngle` | `PhaseAngle` / `Phase` / `IlluminatedFraction` / `BrightLimbPositionAngle` | `PhaseAngle` / `Phase` / `IlluminatedFraction` / `BrightLimbPositionAngle` | `PhaseAngle` / `Phase` / `IlluminatedFraction` / `BrightLimbPositionAngle` |
| 地平量 | `Altitude` / `Azimuth` / `Zenith` / `HourAngle` / `ParallacticAngle` | `Altitude` / `Azimuth` / `Zenith` / `HourAngle` / `ParallacticAngle` | `Altitude` / `Azimuth` / `Zenith` / `HourAngle` / `ParallacticAngle` | `Altitude` / `Azimuth` / `Zenith` / `HourAngle` / `ParallacticAngle` | `Altitude` / `Azimuth` / `Zenith` / `HourAngle` / `ParallacticAngle` | `Altitude` / `Azimuth` / `Zenith` / `HourAngle` / `ParallacticAngle` | `Altitude` / `Azimuth` / `Zenith` / `HourAngle` / `ParallacticAngle` |
| 升 / 落 / 中天 | `RiseTime` / `SetTime` / `DownTime` / `CulminationTime` | `RiseTime` / `SetTime` / `DownTime` / `CulminationTime` | `RiseTime` / `SetTime` / `DownTime` / `CulminationTime` | `RiseTime` / `SetTime` / `DownTime` / `CulminationTime` | `RiseTime` / `SetTime` / `DownTime` / `CulminationTime` | `RiseTime` / `SetTime` / `DownTime` / `CulminationTime` | `RiseTime` / `SetTime` / `DownTime` / `CulminationTime` |
| 物理星历 | `Physical` / `PhysicalN` | `Physical` / `PhysicalN` | `Physical` / `PhysicalN` | `Physical` / `PhysicalN` / `CentralMeridians` / `CentralMeridiansN` | `Physical` / `PhysicalN` / `PhysicalSystemIII` / `Ring` | `Physical` / `PhysicalN` / `PhysicalSystemIII` | `Physical` / `PhysicalN` |
| 合日 | `LastConjunction` / `NextConjunction` | `LastConjunction` / `NextConjunction` | `LastConjunction` / `NextConjunction` | `LastConjunction` / `NextConjunction` | `LastConjunction` / `NextConjunction` | `LastConjunction` / `NextConjunction` | `LastConjunction` / `NextConjunction` |
| 上合 / 下合 | `LastSuperiorConjunction` / `NextSuperiorConjunction` / `LastInferiorConjunction` / `NextInferiorConjunction` | `LastSuperiorConjunction` / `NextSuperiorConjunction` / `LastInferiorConjunction` / `NextInferiorConjunction` | — | — | — | — | — |
| 留 | `LastProgradeToRetrograde` / `NextProgradeToRetrograde` / `LastRetrogradeToPrograde` / `NextRetrogradeToPrograde`，另有 `LastRetrograde` / `NextRetrograde` | `LastProgradeToRetrograde` / `NextProgradeToRetrograde` / `LastRetrogradeToPrograde` / `NextRetrogradeToPrograde`，另有 `LastRetrograde` / `NextRetrograde` | `LastProgradeToRetrograde` / `NextProgradeToRetrograde` / `LastRetrogradeToPrograde` / `NextRetrogradeToPrograde` | `LastProgradeToRetrograde` / `NextProgradeToRetrograde` / `LastRetrogradeToPrograde` / `NextRetrogradeToPrograde` | `LastProgradeToRetrograde` / `NextProgradeToRetrograde` / `LastRetrogradeToPrograde` / `NextRetrogradeToPrograde` | `LastProgradeToRetrograde` / `NextProgradeToRetrograde` / `LastRetrogradeToPrograde` / `NextRetrogradeToPrograde` | `LastProgradeToRetrograde` / `NextProgradeToRetrograde` / `LastRetrogradeToPrograde` / `NextRetrogradeToPrograde` |
| 冲日 | — | — | `LastOpposition` / `NextOpposition` | `LastOpposition` / `NextOpposition` | `LastOpposition` / `NextOpposition` | `LastOpposition` / `NextOpposition` | `LastOpposition` / `NextOpposition` |
| 方照 | — | — | `LastEasternQuadrature` / `NextEasternQuadrature` / `LastWesternQuadrature` / `NextWesternQuadrature` | `LastEasternQuadrature` / `NextEasternQuadrature` / `LastWesternQuadrature` / `NextWesternQuadrature` | `LastEasternQuadrature` / `NextEasternQuadrature` / `LastWesternQuadrature` / `NextWesternQuadrature` | `LastEasternQuadrature` / `NextEasternQuadrature` / `LastWesternQuadrature` / `NextWesternQuadrature` | `LastEasternQuadrature` / `NextEasternQuadrature` / `LastWesternQuadrature` / `NextWesternQuadrature` |
| 大距 | `LastGreatestElongation` / `NextGreatestElongation` / `LastGreatestElongationEast` / `NextGreatestElongationEast` / `LastGreatestElongationWest` / `NextGreatestElongationWest` | `LastGreatestElongation` / `NextGreatestElongation` / `LastGreatestElongationEast` / `NextGreatestElongationEast` / `LastGreatestElongationWest` / `NextGreatestElongationWest` | — | — | — | — | — |
| 凌日 | `LastTransit` / `NextTransit` / `ClosestTransit` / `LocalTransit` | `LastTransit` / `NextTransit` / `ClosestTransit` / `LocalTransit` | — | — | — | — | — |
| 伽利略卫星 | — | — | — | `Satellites` / `SatellitePhenomena` / `LastGalileanPhenomenonEvent` / `NextGalileanPhenomenonEvent` / `ClosestGalileanPhenomenonEvent` / `LastGalileanPhenomenonContactEvent` / `NextGalileanPhenomenonContactEvent` / `ClosestGalileanPhenomenonContactEvent` | — | — | — |
| 结果类型 | `PhysicalInfo` / `TransitInfo` / `LocalTransitInfo` | `PhysicalInfo` / `TransitInfo` / `LocalTransitInfo` | `PhysicalInfo` | `PhysicalInfo` / `CentralMeridianInfo` / `GalileanSatellitesInfo` / `GalileanPhenomenaInfo` / `GalileanSatellitePosition` / `GalileanSatellitePhenomenon` / `GalileanPhenomenonEvent` / `GalileanPhenomenonContactEvent` | `PhysicalInfo` / `RingInfo` | `PhysicalInfo` | `PhysicalInfo` |

七个包的 **`...N` 族**都叫同样的名字：给任意一个瞬时求值接口加 `N` 后缀即可。`n < 0` 用完整级数，半解析的所有 `n >= 0` 用同一张快速表，Legacy 按 `n` 截断，详见 [`...N` 族](#n-族)。事件搜索族（`Last...` / `Next...` / `Closest...`）没有 `N` 版本。

### 通用能力与单位

| 计算项 | 用途 | 单位与口径 |
| --- | --- | --- |
| `ApparentLo` / `ApparentBo` | 地心视黄经、视黄纬 | 度；当日真春分点，含光行时、光行差与章动 |
| `ApparentRa` / `ApparentDec` / `ApparentRaDec` | 地心视赤经、视赤纬 | 度；当日真赤道；`ApparentRaDec` 一次返回两者 |
| `ApparentMagnitude` | 视星等 | 星等（无量纲） |
| `Altitude` / `Azimuth` / `Zenith` | 站心地平坐标 | 度；方位角自正北向东增加，`Zenith` 等于 `90 - Altitude` |
| `HourAngle` | 地心视时角，不含站心视差 | 度 |
| `RiseTime` / `SetTime` / `DownTime` | 当地民用日内的升起、落下 | `time.Time`，保持输入时区；`DownTime` 是 `SetTime` 的兼容别名 |
| `RiseTimeWithAtmosphere` / `SetTimeWithAtmosphere` | 按实测气压（hPa）、气温（℃）计算升落 | `time.Time`；恒带大气折射 |
| `CulminationTime` | 上中天时刻 | `time.Time`，保持输入时区 |
| `ParallacticAngle` | 地心视差角（天顶方向角），由地心时角与赤纬求得 | 度 |
| `Diameter` / `Semidiameter` | 地心视直径、视半径 | 角秒 |
| `PhaseAngle` | 太阳–行星–地球夹角 | 度 |
| `IlluminatedFraction` / `Phase` | 被照亮比例 | `0–1`；`Phase` 是 `IlluminatedFraction` 的别名 |
| `BrightLimbPositionAngle` | 亮面中心位置角 | 度 |
| `EarthDistance` / `SunDistance` | 地心距、日心距 | AU |
| `AscendingNode` / `DescendingNode` | 轨道面与黄道面交点的黄经 | 度；同一时刻两者相差约 `180°` |
| `Physical` | 盘面朝向、子地/子日经纬度、北极位置角 | 度；经度正方向按各天体 IAU 约定 |
| `planet.WherePlanet` / `planet.WherePlanetN` | 当前内置模型的黄经、黄纬、日心距 | 度 / AU；越界返回 `NaN` 而不 panic |

### `...N` 族

每个瞬时求值接口都有 `...N` 后缀版本，用来在精度与开销之间取舍：`n < 0` 使用完整级数，与非 `N` 接口等价；半解析的所有 `n >= 0` 使用同一张快速表，增大 `n` 不提高精度；Legacy 按 `n` 控制主项并等比缩短高阶项。

带 `N` 的族按用途分组：

- 位置与距离：`ApparentLoN`、`ApparentBoN`、`ApparentRaN`、`ApparentDecN`、`ApparentRaDecN`、`EarthDistanceN`、`SunDistanceN`
- 地平与时刻：`AltitudeN`、`AzimuthN`、`ZenithN`、`HourAngleN`、`CulminationTimeN`、`RiseTimeN`、`SetTimeN`、`DownTimeN`、`ParallacticAngleN`
- 视面与相位：`ApparentMagnitudeN`、`DiameterN`、`SemidiameterN`、`PhaseAngleN`、`PhaseN`、`IlluminatedFractionN`、`BrightLimbPositionAngleN`
- 轨道交点与自转：`AscendingNodeN`、`DescendingNodeN`、`PhysicalN`，以及木星的 `CentralMeridiansN`、土星的 `RingN` 与 `PhysicalSystemIIIN`、天王星的 `PhysicalSystemIIIN`

```go
fmt.Println(mars.ApparentLo(date), mars.ApparentLoN(date, 8))   // 完整级数 / 快速表；Legacy 按 n 截断。
fmt.Println(mars.SunDistance(date), mars.SunDistanceN(date, 8)) // 日心距，AU
```

## 常用场景

### 今晚能看到哪颗行星

```go
fmt.Println(venus.RiseTime(date, lon, lat, height, true))                // 金星当日升起
fmt.Println(jupiter.CulminationTime(date, lon))                          // 木星上中天
fmt.Println(mars.Altitude(date, lon, lat), mars.Azimuth(date, lon, lat)) // 火星此刻高度与方位
```

```text
2020-01-01 10:02:34.350145161 +0800 CST <nil>
2020-01-01 12:32:17.585815787 +0800 CST
31.194578177219057 152.07031660415714
```

方位角自正北向东增加；`aero = true` 按标准大气折射把几何地平线降到约 `-0.5667°`，逐参数口径与极区哨兵错误见[升落与中天](#升落与中天)。

### 冲日、合日、大距、留与逆行

```go
fmt.Println(mars.NextOpposition(date))             // 火星下次冲日
fmt.Println(jupiter.NextConjunction(date))         // 木星下次合日
fmt.Println(saturn.NextProgradeToRetrograde(date)) // 土星下次顺转逆的留
```

```text
2020-10-14 07:25:50.441412627 +0800 CST
2021-01-29 09:39:33.697994649 +0800 CST
2020-05-11 17:26:53.961271941 +0800 CST
```

事件搜索取"当前或之后最近一次"并保持输入时区，成对的 `Last...` 与不区分方向的 `NextRetrograde` 见[合冲留与方照](#合冲留与方照)；水星、金星的 `NextGreatestElongationEast` / `...West` 见[大距与地心凌日](#大距与地心凌日)。

### 水星与金星凌日

```go
transit := mercury.NextTransit(date)                           // 2020 年之后下一场地心水星凌日
fmt.Println(transit.Valid, transit.Start, transit.Greatest)     // 是否有凌日、一触与凌甚
fmt.Println(transit.Duration, transit.MinimumSeparationArcsec)  // 历时与凌甚最小角距
```

```text
true 2032-11-13 14:41:13.161198198 +0800 CST 2032-11-13 16:54:12.821315824 +0800 CST
4h26m2.695272267s 572.0643215495325
```

`TransitInfo.Valid` 为假表示搜索窗口内没有凌日、其余字段是零值；`TransitInfo` 是地心几何结果。某个地点看到的接触时刻用 `LocalTransit` 求站心结果：

```go
transit := mercury.NextTransit(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC))
local := mercury.LocalTransit(transit, -77.0, 38.9, 0)        // 华盛顿，经度、纬度、海拔米
fmt.Println(local.Valid, local.HasInternal, local.Visible)     // 站心有凌日、有内切接触、太阳在视地平上
fmt.Println(local.Start.Time, local.Start.SunAltitude)         // 站心一触与当时太阳高度
```

```text
true true true
2019-11-11 12:36:09.037762284 +0000 UTC 7.933256722170642
```

站心接触对太阳与行星分别做视差修正后重解，所以各地接触时刻与地心值不同：金星可达约 ±7 分钟，水星约 ±2 分钟。

- `SunAltitude` 是日心几何高度。
- `Visible`、`Sunrise`、`Sunset` 按太阳中心是否在视地平上判定，计入标准大气（1010 hPa、10 °C）折射，几何阈值约 −34′。
- 地心有内切接触、站心可能只是掠凌（例如 1999-11-15 水星凌日在南半球高纬），此时 `HasInternal` 为假，`InternalStart` / `InternalEnd` 是零值。
- 全球可见图与站心日面图见[日月食与凌日](eclipse.md#凌日出图)；四触、偏凌与内切口径见[大距与地心凌日](#大距与地心凌日)。

### 相位、视直径、视星等与节点

```go
fmt.Println(venus.PhaseAngle(date), venus.Phase(date))           // 相位角（度）与被照亮比例
fmt.Println(venus.Diameter(date), venus.ApparentMagnitude(date)) // 视直径（角秒）与视星等
fmt.Println(mars.AscendingNode(date), mars.DescendingNode(date)) // 升降交点黄经（度）
```

```text
49.98145049145023 0.8215177914415865
13.059409604614839 -4
49.71479005849112 229.71479005849113
```

`Phase` 是 `IlluminatedFraction` 的别名，取值 `0–1`；升降交点同一时刻相差约 `180°`。各量的定义、别名与 `n` 变体见[节点、相位、视星等、视直径与视差角](#节点相位视星等视直径与视差角)。

### 物理星历与木星伽利略卫星

```go
j := jupiter.Physical(date)                                    // 木星物理星历
fmt.Println(j.DS, j.DE, j.CentralMeridianSystemIII)            // 子日/子地赤纬与 System III 中央经线
sats := jupiter.Satellites(date)                               // 四颗伽利略卫星的瞬时位置
fmt.Println(sats.Io.OffsetXJupiterR, sats.Io.InFrontOfJupiter) // 木卫一偏移与是否在盘面前方
```

```text
-56.55778470155335 -2.039966127259664 311.37430665615585
3.8223302102343975 false
```

土星环另有 `saturn.Ring`（`EarthLatitude`、`MinorAxis` 等，角度单位度、长短轴单位角秒）；盘面朝向与中央经线的完整口径见[物理星历](#物理星历)，伽利略卫星的公开入口在 `jupiter` 包，`basic` 的 `JupiterGalilean*` 是接收儒略日的低层入口，见[木星伽利略卫星](#木星伽利略卫星)。

### 与外部资料的对照口径

```go
fmt.Println(mars.ApparentLo(date), mars.ApparentLoN(date, 8)) // 完整级数 / 快速表；Legacy 按 n 截断。
_, err := mars.RiseTime(date, 0, 89, 0, true)                 // 极区观测点
fmt.Println(errors.Is(err, mars.ERR_MARS_NEVER_RISE), errors.Is(err, mars.ERR_MARS_NEVER_SET))
```

```text
238.38840227925655 238.39637464888327
true false
```

`ApparentLo` 等是**当日真春分点**的地心视位置。与外部 J2000 或平位置对表时，需同时统一坐标框架、章动与视位置修正，不能仅用岁差变换消除全部差异。`n` 的模型相关语义见 [`...N` 族](#n-族)。伽利略卫星接触事件与外部年表的口径差异见[与外部资料对照](#与外部资料对照)，整体边界见[参数与返回值约定](#参数与返回值约定)。

## 基础示例

下面两段是最小可运行示例，沿用 `date = 2020-01-01 08:08:08 CST` 与西安市坐标。

### 内行星

```go
package main

import (
	"github.com/starainrt/astro/mercury"
	"github.com/starainrt/astro/venus"
	"fmt"
	"time"
)

func main() {
	// 以陕西省西安市为例，设置西安市经纬度,设置地平高度为0米
	var lon, lat, height float64 = 108.93, 34.27, 0
	cst := time.FixedZone("CST", 8*3600)
	// 指定观测时刻。
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	//水星上次下合时间
	fmt.Println(mercury.LastInferiorConjunction(date))
	//金星下次上合时间
	fmt.Println(venus.NextSuperiorConjunction(date))
	//水星上次留（顺转逆）时间（水逆）
	fmt.Println(mercury.LastProgradeToRetrograde(date))
	//金星下次留（逆转顺）时间
	fmt.Println(venus.NextRetrogradeToPrograde(date))
	//水星上次东大距时间
	fmt.Println(mercury.LastGreatestElongationEast(date))
	//金星下次西大距时间
	fmt.Println(venus.NextGreatestElongationWest(date))
	//西安市今日金星升起，降落时间
	fmt.Println(venus.RiseTime(date, lon, lat, height, true))
	fmt.Println(venus.SetTime(date, lon, lat, height, true))
	//金星当前视星等
	fmt.Println(venus.ApparentMagnitude(date))
	//金星相位角、被照亮比例、亮面中心位置角
	fmt.Println(venus.PhaseAngle(date))
	fmt.Println(venus.Phase(date))
	fmt.Println(venus.BrightLimbPositionAngle(date))
	//金地距离
	fmt.Println(venus.EarthDistance(date))
	//金日距离
	fmt.Println(venus.SunDistance(date))
}
```

输出结果：

```text
2019-11-11 23:21:41.729008555 +0800 CST // 水星上次下合
2021-03-26 14:57:43.352850079 +0800 CST // 金星下次上合
2019-11-01 04:31:49.712648391 +0800 CST // 水星上次由顺行转逆行的留
2020-06-25 02:07:26.280500292 +0800 CST // 金星下次由逆行转顺行的留
2019-10-20 12:01:37.925546765 +0800 CST // 水星上次东大距
2020-08-13 08:14:45.946914553 +0800 CST // 金星下次西大距
2020-01-01 10:02:34.94517982 +0800 CST <nil> // 西安当天金星升起时刻；无错误
2020-01-01 20:25:36.944952607 +0800 CST <nil> // 西安当天金星落下时刻；无错误
-4 // 金星视星等
49.98142313573699 // 金星相位角，单位度
0.8215179742647736 // 金星被照亮比例
255.63803631896667 // 金星亮面中心位置角，单位度
1.2778823747747166 // 金地距离，单位 AU
0.7262651318993368 // 金日距离，单位 AU
```

内外行星同样提供 `Diameter` / `Semidiameter`（以及 `N` 版），返回地心视直径/视半径，单位为角秒。

行星视直径或轨道节点也可以单独查询：

```go
fmt.Println(mars.Diameter(date), mars.Semidiameter(date))
fmt.Println(venus.AscendingNode(date), venus.DescendingNode(date))
```

这里的“升交点 / 降交点”指天体轨道面与黄道面的两个交点：

- `AscendingNode`：天体从黄道南侧穿到黄道北侧时对应的黄经
- `DescendingNode`：天体从黄道北侧穿到黄道南侧时对应的黄经
- 返回值单位都是度；对同一时刻而言，降交点通常与升交点相差约 `180°`

以上面 `date := 2020-01-01 08:08:08 CST` 的示例来说，输出结果是：

```text
4.287299886569956 2.143649943284978 // 火星视直径、视半径，单位角秒
76.86008484515058 256.8600848451506 // 金星升交点、降交点黄经，单位度
```

水星和金星还提供 `NextTransit` / `LastTransit` / `ClosestTransit` 地心凌日查询。这里的“地心”指从地球中心看到的行星圆面经过太阳圆面，不判断某个地点当时太阳是否在地平线上；观测计划用 `LocalTransit` 求站心接触与当时太阳高度（见[水星与金星凌日](#水星与金星凌日)）。

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/mercury"
	"github.com/starainrt/astro/venus"
)

func main() {
	// 查询 2019 年之后下一次地心水星凌日。
	mercuryTransit := mercury.NextTransit(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC))
	fmt.Println(mercuryTransit.Valid)
	fmt.Println(mercuryTransit.Start)
	fmt.Println(mercuryTransit.InternalStart)
	fmt.Println(mercuryTransit.Greatest)
	fmt.Println(mercuryTransit.InternalEnd)
	fmt.Println(mercuryTransit.End)
	fmt.Println(mercuryTransit.Duration)
	fmt.Println(mercuryTransit.MinimumSeparationArcsec)
	fmt.Println(mercuryTransit.SunSemidiameterArcsec)
	fmt.Println(mercuryTransit.PlanetSemidiameterArcsec)

	// 查询 2012 年之后下一次地心金星凌日。
	venusTransit := venus.NextTransit(time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC))
	fmt.Println(venusTransit.Valid)
	fmt.Println(venusTransit.Start)
	fmt.Println(venusTransit.InternalStart)
	fmt.Println(venusTransit.Greatest)
	fmt.Println(venusTransit.InternalEnd)
	fmt.Println(venusTransit.End)
	fmt.Println(venusTransit.Duration)
}
```

输出结果：

```text
true // 找到一次有效的地心水星凌日
2019-11-11 12:35:31.142292916 +0000 UTC // 一触：水星外切进入太阳圆面
2019-11-11 12:37:12.603219151 +0000 UTC // 二触：水星完全进入太阳圆面
2019-11-11 15:19:48.146160542 +0000 UTC // 凌甚：水星中心最接近太阳中心
2019-11-11 18:02:28.751637339 +0000 UTC // 三触：水星开始离开太阳圆面
2019-11-11 18:04:10.212523341 +0000 UTC // 四触：水星外切离开太阳圆面
5h28m39.070230425s // 一触到四触的地心凌日持续时间
75.93592092696468 // 凌甚时水星中心与太阳中心的最小角距离，单位角秒
968.8881399120452 // 凌甚时太阳视半径，单位角秒
4.978442587642365 // 凌甚时水星视半径，单位角秒
true // 找到一次有效的地心金星凌日
2012-06-05 22:09:48.690456748 +0000 UTC // 一触：金星外切进入太阳圆面
2012-06-05 22:27:37.088886201 +0000 UTC // 二触：金星完全进入太阳圆面
2012-06-06 01:29:36.58495903 +0000 UTC // 凌甚：金星中心最接近太阳中心
2012-06-06 04:31:35.87008953 +0000 UTC // 三触：金星开始离开太阳圆面
2012-06-06 04:49:24.268559217 +0000 UTC // 四触：金星外切离开太阳圆面
6h39m35.578102469s // 一触到四触的地心凌日持续时间
```

### 外行星

```go
package main

import (
	"github.com/starainrt/astro/jupiter"
	"github.com/starainrt/astro/mars"
	"github.com/starainrt/astro/neptune"
	"github.com/starainrt/astro/saturn"
	"github.com/starainrt/astro/uranus"
	"fmt"
	"time"
)

func main() {
	// 以陕西省西安市为例，设置西安市经纬度,设置地平高度为0米
	var lon, lat, height float64 = 108.93, 34.27, 0
	cst := time.FixedZone("CST", 8*3600)
	// 指定观测时刻。
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)
	//火星下次冲日时间
	fmt.Println(mars.NextOpposition(date))
	//木星下次合日时间
	fmt.Println(jupiter.NextConjunction(date))
	//土星上次留（顺转逆）时间（土逆）
	fmt.Println(saturn.LastProgradeToRetrograde(date))
	//土星环观测参数
	ring := saturn.Ring(date)
	fmt.Printf("saturn B=%.6f Bp=%.6f P=%.6f dU=%.6f major=%.6f minor=%.6f\n",
		ring.EarthLatitude,
		ring.SunLatitude,
		ring.PositionAngle,
		ring.DeltaU,
		ring.MajorAxis,
		ring.MinorAxis,
	)
	//天王星下次留（逆转顺）时间
	fmt.Println(uranus.NextRetrogradeToPrograde(date))
	//海王星上次东方照时间
	fmt.Println(neptune.LastEasternQuadrature(date))
	//火星下次西方照时间
	fmt.Println(mars.NextWesternQuadrature(date))
	//西安市今日火星升起，降落时间
	fmt.Println(mars.RiseTime(date, lon, lat, height, true))
	fmt.Println(mars.SetTime(date, lon, lat, height, true))
	//火星当前视星等
	fmt.Println(mars.ApparentMagnitude(date))
	//地火距离
	fmt.Println(mars.EarthDistance(date))
	//日火距离
	fmt.Println(mars.SunDistance(date))
}
```

输出结果：

```text
2020-10-14 07:25:53.541214764 +0800 CST // 火星下次冲日
2021-01-29 09:39:32.556620836 +0800 CST // 木星下次合日
2019-04-30 10:28:17.254818677 +0800 CST // 土星上次由顺行转逆行的留
saturn B=23.577010 Bp=23.266930 P=6.629810 dU=1.171003 major=34.133617 minor=13.652809 // 土星环 B、B'、P、dU、长轴、短轴
2020-01-11 15:22:18.532456755 +0800 CST // 天王星下次由逆行转顺行的留
2019-12-08 17:00:04.494322836 +0800 CST // 海王星上次东方照
2020-06-07 03:10:55.059157311 +0800 CST // 火星下次西方照
2020-01-01 04:41:30.157055854 +0800 CST <nil> // 西安当天火星升起时刻；无错误
2020-01-01 14:55:32.792302966 +0800 CST <nil> // 西安当天火星落下时刻；无错误
1.57 // 火星视星等
2.1844226019239925 // 地火距离，单位 AU
1.5897858533008575 // 日火距离，单位 AU
```

`saturn.Ring` 返回 `RingInfo`：`EarthLatitude` 是土星环张角 B，`SunLatitude` 是 B'，`PositionAngle` 是北半短轴位置角，`DeltaU` 是太阳与地球在环面内的土星心黄经差，`MajorAxis` / `MinorAxis` 是土星环外缘长短轴，单位为角秒。

## 分主题示例

下面按主题给出短片段，每段都接得上本手册的公共变量 `date`、`lon`、`lat`、`height`；完整可运行版本见前面的基础示例。

### 位置与坐标

`ApparentLo` / `ApparentBo` / `ApparentRa` / `ApparentDec` / `ApparentRaDec` 给的是**地心视位置**：几何地心位置经光行时、光行差与章动改正，再换算到当日真赤道与真春分点。这些包不提供 J2000 或平位置输出；要对 J2000 口径时，用 `coord.Precess` 做岁差归算、用 `coord.Nutation2000B` 处理章动，站心量用 `coord.TopocentricEquatorial` 与 `coord.EquatorialToHorizontal`，详见[坐标工具](coord.md)。

```go
// 当日视位置：视黄经/视黄纬与视赤经/视赤纬，单位度。
lo, bo := venus.ApparentLo(date), venus.ApparentBo(date)
ra, dec := venus.ApparentRaDec(date)
fmt.Println(lo, bo, venus.ApparentRa(date), venus.ApparentDec(date), ra, dec)

// 地心距与日心距，单位 AU。
fmt.Println(venus.EarthDistance(date), venus.SunDistance(date))
```

### 升落与中天

`RiseTime` / `SetTime` / `DownTime` 按**当地民用日**搜索：`date` 用于确定当地日期与输出时区，返回的事件属于该当地民用日；`height` 是椭球高（米），`aero` 为 `true` 时加入标准大气折射（几何地平线降到约 `-0.5667°`）。极昼、极夜或当天没有过零时返回[哨兵错误](#参数与返回值约定)而不是时刻。`CulminationTime` 给上中天；`Altitude` / `Azimuth` / `Zenith` 是站心瞬时量，`HourAngle` / `ParallacticAngle` 是地心瞬时量，几何公式见[坐标工具](coord.md)。

```go
// 西安市当地民用日内的升起、落下与中天；aero=true 含标准大气折射。
rise, err := mars.RiseTime(date, lon, lat, height, true)
set, err := mars.SetTime(date, lon, lat, height, true)
fmt.Println(rise, set, err)
fmt.Println(mars.CulminationTime(date, lon))
```

```go
// 地平量：经度东正西负、纬度北正南负，返回值单位度。
fmt.Println(mars.Altitude(date, lon, lat), mars.Azimuth(date, lon, lat))
fmt.Println(mars.Zenith(date, lon, lat), mars.HourAngle(date, lon))
fmt.Println(mars.ParallacticAngle(date, lon, lat))
```

```go
// DownTime 是 SetTime 的兼容别名；N 版按 `n` 变体求值。
set, err := mars.DownTime(date, lon, lat, height, true)
fmt.Println(set, err)
fmt.Println(mars.CulminationTimeN(date, lon, 12))
```

### 合冲留与方照

事件搜索一律取"当前或之前/之后最近一次"（含端点），返回时区与输入一致。内行星有 `LastConjunction` / `NextConjunction` 与上合下合族；外行星有 `LastConjunction` / `NextConjunction`、`LastOpposition` / `NextOpposition` 与东西方照族。

留分两个方向：`LastProgradeToRetrograde` / `NextProgradeToRetrograde`（顺转逆）与 `LastRetrogradeToPrograde` / `NextRetrogradeToPrograde`（逆转顺）；水星、金星另有不区分方向的 `LastRetrograde` / `NextRetrograde`，外行星没有这两个名字。

合日对内行星是太阳同侧，对外行星还额外分冲日与方照：只有地球轨道之外的这五颗行星会出现冲日（太阳–地球–行星成一线）与东西方照（行星与太阳黄经相差约 90°），水星和金星在地球轨道以内，所以这两族接口只出现在外行星包里。

```go
// 内行星：上合/下合，以及不区分方向的留。
fmt.Println(mercury.LastSuperiorConjunction(date), mercury.NextInferiorConjunction(date))
fmt.Println(mercury.LastRetrograde(date), mercury.NextRetrograde(date))
```

```go
// 外行星：合日、冲日与东西方照。
fmt.Println(jupiter.NextConjunction(date), mars.NextOpposition(date))
fmt.Println(mars.NextEasternQuadrature(date), neptune.LastWesternQuadrature(date))
```

```go
// 留的两个方向分别有 Last/Next 两版。
fmt.Println(saturn.LastProgradeToRetrograde(date), saturn.NextProgradeToRetrograde(date))
fmt.Println(saturn.LastRetrogradeToPrograde(date), saturn.NextRetrogradeToPrograde(date))
```

### 大距与地心凌日

大距只对水星、金星有意义：`LastGreatestElongation` / `NextGreatestElongation` 不区分东西，`LastGreatestElongationEast` / `NextGreatestElongationEast` 与 `...West` 区分。

地心凌日的公开入口同样只在 `mercury` / `venus`：`LastTransit` / `NextTransit` / `ClosestTransit` 返回 `TransitInfo`。`Valid` 为假表示搜索窗口内没有凌日，其余字段都是零值；偏凌日没有内切，此时 `HasInternal` 为假、`InternalStart` / `InternalEnd` 是零值。凌日只判断地心几何，不判断某地当时太阳是否在地平线上。

```go
// 大距：区分东西用 ...East / ...West，不区分时用 NextGreatestElongation。
fmt.Println(mercury.NextGreatestElongationEast(date), venus.LastGreatestElongationWest(date))
fmt.Println(venus.NextGreatestElongation(date))
```

```go
// 地心凌日：Valid 为假时其余字段为零值。
transit := mercury.NextTransit(date)
if transit.Valid {
	fmt.Println(transit.Start, transit.InternalStart, transit.Greatest, transit.InternalEnd, transit.End)
	fmt.Println(transit.Duration, transit.MinimumSeparationArcsec, transit.SunSemidiameterArcsec)
}
```

```go
// 凌日结果的其余字段；无内切时 InternalDuration 为 0。
transit := venus.ClosestTransit(date)
fmt.Println(transit.HasInternal, transit.InternalDuration, transit.MinimumSeparationArcsec)
fmt.Println(transit.SunSemidiameterArcsec, transit.PlanetSemidiameterArcsec)
```

### 节点、相位、视星等、视直径与视差角

`AscendingNode` / `DescendingNode` 是行星轨道面与黄道面两个交点的黄经，单位度，同一时刻两者相差约 `180°`。`PhaseAngle` 是太阳–行星–地球夹角（度）；`IlluminatedFraction`（别名 `Phase`）是 `0–1` 的被照亮比例；`BrightLimbPositionAngle` 是亮面中心位置角（度）。`Diameter` / `Semidiameter` 返回地心视直径/视半径，单位角秒。

`ParallacticAngle` 返回地心视差角（天顶方向角，度），由地心时角与赤纬求得，不含站心视差；站心坐标入口见[坐标工具](coord.md)。

```go
fmt.Println(mars.AscendingNode(date), mars.DescendingNode(date))                     // 升降交点黄经，度
fmt.Println(mars.PhaseAngle(date), mars.IlluminatedFraction(date), mars.Phase(date)) // 相位角，度；照亮比例
fmt.Println(mars.ApparentMagnitude(date), mars.BrightLimbPositionAngle(date))        // 视星等、亮面位置角
```

```go
fmt.Println(mars.Diameter(date), mars.Semidiameter(date)) // 视直径、视半径，单位角秒
fmt.Println(mars.ParallacticAngle(date, lon, lat))        // 视差角，度
```

```go
// 交点与视直径同样有 `n` 变体。
fmt.Println(mars.AscendingNodeN(date, 12), mars.DescendingNodeN(date, 12))
fmt.Println(mars.DiameterN(date, 12), mars.SemidiameterN(date, 12))
```

### 与其他手册的分工

- 恒星时、岁差章动、站心与地平坐标换算见[坐标工具](coord.md)。
- 太阳与月亮的位置、月相、月出月落、朔望弦见[日月手册](sun-moon.md)。
- 行星月掩、掩带与月面视圆图见[月掩手册](occultation.md#月掩出图)。
- 小行星、彗星等按轨道根数计算的天体见[通用轨道](orbit.md)。
- 时标声明、UT1 口径与 GeoJSON 输出见[天象地图与 GeoJSON](map-geojson.md)；时标本身的约定见[时标声明](map-geojson.md#时标声明)。
- 权威能力清单、依赖矩阵与精度汇总见根目录 [README](../../README.md)。

### 物理星历

七大行星都提供 `Physical` / `PhysicalN`，用于查看盘面朝向、子地/子日经纬度和北极位置角等物理观测参数。木星额外提供 System I/II/III 中央经线，土星额外提供土星环参数。

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/jupiter"
	"github.com/starainrt/astro/saturn"
)

func main() {
	date := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)

	// 木星：DS/DE 分别是太阳、地球相对木星赤道的行星中心赤纬。
	// CMI/CMII/CMIII 是木星 System I/II/III 中央经线，单位度。
	j := jupiter.Physical(date)
	fmt.Printf("jupiter DS=%.6f DE=%.6f CMI=%.6f CMII=%.6f CMIII=%.6f\n",
		j.DS,
		j.DE,
		j.CentralMeridianSystemI,
		j.CentralMeridianSystemII,
		j.CentralMeridianSystemIII,
	)

	// 土星环：B/B' 是地球、太阳看到的环面纬度，P 是环面短轴位置角。
	ring := saturn.Ring(date)
	fmt.Printf("saturn B=%.6f Bp=%.6f P=%.6f major=%.6f minor=%.6f\n",
		ring.EarthLatitude,
		ring.SunLatitude,
		ring.PositionAngle,
		ring.MajorAxis,
		ring.MinorAxis,
	)
}
```

输出结果：

```text
jupiter DS=54.342201 DE=1.436449 CMI=292.712946 CMII=276.309085 CMIII=147.241848 // 木星子日/子地赤纬，System I/II/III 中央经线，单位度
saturn B=-0.608032 Bp=-2.675667 P=4.480277 major=42.709152 minor=0.453228 // 土星环 B、B'、短轴位置角、外缘长短轴，角度单位度，长短轴单位角秒
```

只需要中央经线时，可以单独调用 `CentralMeridians`：

```go
cm := jupiter.CentralMeridians(date)
fmt.Printf("CMI=%.6f CMII=%.6f CMIII=%.6f\n", cm.SystemI, cm.SystemII, cm.SystemIII) // 木星 System I/II/III 中央经线
```

土星和天王星则额外保留了显式的 `System III` 语义别名，便于按行星自转系统来写调用代码：

```go
sat3 := saturn.PhysicalSystemIII(date)
ura3 := uranus.PhysicalSystemIII(date)
fmt.Printf("saturn systemIII lon=%.6f lat=%.6f P=%.6f\n", sat3.SubEarthLongitude, sat3.SubEarthLatitude, sat3.NorthPolePositionAngle) // 土星子地经纬度与北极位置角
fmt.Printf("uranus systemIII lon=%.6f lat=%.6f P=%.6f\n", ura3.SubEarthLongitude, ura3.SubEarthLatitude, ura3.NorthPolePositionAngle) // 天王星子地经纬度与北极位置角
```

七个包的 `Physical` 都返回各自的 `PhysicalInfo`，字段与 `basic.PlanetPhysicalInfo` 一一对应；`SubEarthLongitude` / `SubSolarLongitude` 的正方向按各天体当前 IAU/Horizons 制图约定，水星、火星、木星、土星、海王星取西经为正，金星、天王星取东经为正。土星环参数只用于盘面与掩带表达，不参与月掩接触计算，出图见[月掩手册](occultation.md#月掩出图)。

```go
p := uranus.Physical(date) // 天王星子地/子日经纬度与北极位置角，单位度
fmt.Println(p.SubEarthLongitude, p.SubEarthLatitude, p.SubSolarLongitude, p.SubSolarLatitude, p.NorthPolePositionAngle)
```

```go
// 土星环与木星中央经线的 `n` 变体。
fmt.Println(saturn.RingN(date, 12).MinorAxis, jupiter.CentralMeridiansN(date, 12).SystemIII)
```

### 木星伽利略卫星

这组接口的公开入口在 **`jupiter` 包**（`jupiter.Satellites`、`jupiter.SatellitePhenomena`、`jupiter.NextGalileanPhenomenonEvent` 等），接收 `time.Time` 并返回 `jupiter` 自己的类型。

`basic` 里的 `JupiterGalilean*`（如 `basic.JupiterGalileanSatelliteObservations`、`basic.NextJupiterGalileanPhenomenonEvent`）是同一实现的低层入口，接收儒略日并返回 `basic` 类型；写应用时用 `jupiter` 这一层即可。

```go
// 入口在 jupiter 包；卫星编号用 jupiter.GalileanSatelliteIo 等常量。
sats := jupiter.Satellites(date)
fmt.Println(sats.Io.OffsetXJupiterR, sats.Io.OffsetYJupiterR, sats.Io.InFrontOfJupiter)
```

卫星编号与现象类型都有具名常量，不必手写数字或字符串：

- 卫星编号：`GalileanSatelliteIo`、`GalileanSatelliteEuropa`、`GalileanSatelliteGanymede`、`GalileanSatelliteCallisto`
- 现象类型（`GalileanPhenomenonType`）：`GalileanPhenomenonTransit`、`GalileanPhenomenonOccultation`、`GalileanPhenomenonEclipse`、`GalileanPhenomenonShadowTransit`
- 接触阶段（`GalileanPhenomenonContactPhase`）：`GalileanPhenomenonContactDisappearance`、`GalileanPhenomenonContactReappearance`

```go
// 卫星编号与现象类型都用常量，避免手写数字与字符串。
event := jupiter.NextGalileanPhenomenonEvent(date, jupiter.GalileanSatelliteCallisto, jupiter.GalileanPhenomenonShadowTransit)
fmt.Println(event.Type == jupiter.GalileanPhenomenonShadowTransit)
```

`jupiter` 包提供四颗伽利略卫星的视位置、瞬时现象和事件搜索。

常用接口：

- `Satellites`：四颗卫星相对木星盘面的瞬时视位置
- `SatellitePhenomena`：瞬时凌日、掩蔽、食、影凌状态
- `LastGalileanPhenomenonEvent` / `NextGalileanPhenomenonEvent` / `ClosestGalileanPhenomenonEvent`：搜索整场现象区间
- `LastGalileanPhenomenonContactEvent` / `NextGalileanPhenomenonContactEvent` / `ClosestGalileanPhenomenonContactEvent`：搜索 IMCCE 风格的 D/F 接触事件

两个口径的区别如下，以木卫一凌日为例：

- `GalileanPhenomenonEvent` 把卫星看作一个点，判断“卫星圆心是否进入/离开木星圆面”。它返回整段凌日的起止区间，适合快速搜索现象和程序内部状态判断。
- `GalileanPhenomenonContactEvent` 把卫星自身的有限圆盘考虑进去，区分初亏到复圆的完整接触过程。它返回消失阶段（D）和再现阶段（R）各自的接触起止与模型中心穿越时刻，适合和 IMCCE 年表中的 `TR.D/TR.F/OC.D/OC.F/EC.D/EC.F/SH.D/SH.F` 逐项对照。

两个口径的差异在持续时间上最多约 7 分钟，差异来自模型定义不同。用于观测预报或与公开年表逐项核对时，取 `GalileanPhenomenonContactEvent`。

#### 代码示例

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/jupiter"
)

func main() {
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	// 四颗卫星相对木星中心的瞬时位置。
	sats := jupiter.Satellites(date)
	fmt.Printf("io x=%.6f y=%.6f front=%v\n", sats.Io.OffsetXJupiterR, sats.Io.OffsetYJupiterR, sats.Io.InFrontOfJupiter)
	fmt.Printf("europa ra=%.6f dec=%.6f\n", sats.Europa.ApparentRA, sats.Europa.ApparentDec)

	// 瞬时现象标志。
	ph := jupiter.SatellitePhenomena(date)
	fmt.Printf("io transit=%v occultation=%v eclipse=%v shadow=%v\n", ph.Io.Transit, ph.Io.Occultation, ph.Io.Eclipse, ph.Io.ShadowTransit)
	fmt.Printf("europa transit=%v occultation=%v eclipse=%v shadow=%v\n", ph.Europa.Transit, ph.Europa.Occultation, ph.Europa.Eclipse, ph.Europa.ShadowTransit)

	// 下一次木卫一凌日整场事件。
	event := jupiter.NextGalileanPhenomenonEvent(date, jupiter.GalileanSatelliteIo, jupiter.GalileanPhenomenonTransit)
	fmt.Printf("event valid=%v sat=%d type=%s\n", event.Valid, event.Satellite, event.Type)
	fmt.Println(event.Start)
	fmt.Println(event.Greatest)
	fmt.Println(event.End)
	fmt.Println(event.Duration)

	// 下一次木卫二掩蔽的 IMCCE 风格接触窗口。
	contact := jupiter.NextGalileanPhenomenonContactEvent(date, jupiter.GalileanSatelliteEuropa, jupiter.GalileanPhenomenonOccultation)
	fmt.Printf("contact valid=%v sat=%d type=%s\n", contact.Valid, contact.Satellite, contact.Type)
	fmt.Println(contact.Disappearance.Start)
	fmt.Println(contact.Disappearance.ModelCrossing)
	fmt.Println(contact.Disappearance.End)
	fmt.Println(contact.Greatest)
	fmt.Println(contact.Reappearance.Start)
	fmt.Println(contact.Reappearance.ModelCrossing)
	fmt.Println(contact.Reappearance.End)
}
```

输出结果：

```text
io x=-0.675023 y=-0.032794 front=true // 木卫一相对木星中心的 X/Y 偏移，单位木星半径；位于木星盘面前方
europa ra=110.769095 dec=22.335877 // 木卫二视赤经、视赤纬，单位度
io transit=true occultation=false eclipse=false shadow=true // 木卫一正在凌日，且影子正在凌日
europa transit=false occultation=false eclipse=false shadow=false // 木卫二此刻无凌日、掩蔽、木星食或影凌
event valid=true sat=1 type=transit // 下一次有效事件为木卫一凌日
2026-01-16 16:32:47.552742362 +0000 UTC // 木卫一凌日开始
2026-01-16 17:40:44.189371168 +0000 UTC // 木卫一凌日中点
2026-01-16 18:48:40.287077128 +0000 UTC // 木卫一凌日结束
2h15m52.734334766s // 木卫一凌日持续时间
contact valid=true sat=2 type=occultation // 下一次有效接触事件为木卫二掩蔽
2026-01-17 01:00:34.99533087 +0000 UTC // 木卫二掩蔽消失阶段开始
2026-01-17 01:02:31.714070141 +0000 UTC // 木卫二掩蔽消失阶段模型中心穿越
2026-01-17 01:04:28.432809412 +0000 UTC // 木卫二掩蔽消失阶段结束
2026-01-17 02:27:37.807798683 +0000 UTC // 木卫二掩蔽最深时刻
2026-01-17 03:50:48.120300471 +0000 UTC // 木卫二掩蔽再现阶段开始
2026-01-17 03:52:43.901527225 +0000 UTC // 木卫二掩蔽再现阶段模型中心穿越
2026-01-17 03:54:39.68275398 +0000 UTC // 木卫二掩蔽再现阶段结束
```

#### 与外部资料对照

木卫能力主要对照了两类外部基线：

- **JPL Horizons**：用于四颗卫星相对木星中心的视位置，以及影凌时影心相对木星盘面的偏移。
- **IMCCE 2026 年表**：用于凌日、掩蔽、木星食、影凌等事件和 D/F 接触窗口。

对照结果：

- `Satellites` 相对木星中心的位置，对 JPL Horizons 的样例最大偏差约为 `X=0.054"`、`Y=0.048"`。
- `SatellitePhenomena` 的影凌影心偏移，对 JPL Horizons 的样例最大偏差约为 `X=0.051"`、`Y=0.016"`，现象布尔标志在样例中一致。
- `GalileanPhenomenonContactEvent` 与 IMCCE 2026 年表的 D/F 接触时刻最大偏差约 `79 s`，接触持续时间最大偏差约 `17 s`。
- `GalileanPhenomenonEvent` 与 IMCCE 的 D/F 接触定义不同，起止时刻的差异可达约 `7` 分钟。

### 与 `planet` 包共用的低层入口

`planet` 包提供内置模型与级数的低层入口，被七个行星包与 `sun` / `moon` 共用。它只导出函数，没有导出类型；行星包共用的是数值口径与模型选择，而不是结果类型。各包的 `PhysicalInfo` 与 `basic.PlanetPhysicalInfo`、`TransitInfo` 与 `basic.PlanetTransitResult` 字段一一对应，类型本身仍属于各自的包。外部核引擎类型位于独立子包 `planet/spk`。

| 入口 | 作用 | 单位 |
| --- | --- | --- |
| `planet.WherePlanet` / `planet.WherePlanetN` | 当前内置模型结果：`xt` 取 `1..7` 依次为水星到海王星、`-1` 或 `0` 为地球，`zn` 取 `0` 黄经、`1` 黄纬、`2` 日心距；`xt` / `zn` 越界返回 `NaN` 而不 panic | 度 / AU |
| `planet.Distance` | 日地距离 | AU |
| `planet.SunLo` / `planet.SunM` / `planet.SunMidFun` / `planet.SunTrueLo` / `planet.SunApparentLo` | 太阳几何黄经、平近点角、中心差、真黄经、视黄经 | 度 |
| `planet.Earthe` / `planet.EarthPI` | 地球轨道偏心率、地球近日点黄经 | 无量纲 / 度 |
| `planet.MoonLo` / `planet.MoonM` / `planet.MoonLonX` / `planet.SunMoonAngle` | 月球平黄经、平近点角、到升交点的平角距、日月距角 | 度 |
| `planet.MoonI` / `planet.MoonB` / `planet.MoonR` | 月球黄经、黄纬、距离周期项（ELP2000/82 风格截断级数） | `10⁻⁶` 度 / `10⁻⁶` 度 / `10⁻³ km` |
| `planet.MoonTrueLo` / `planet.MoonTrueBo` / `planet.MoonAway` | 月球真黄经、真黄纬、地心距 | 度 / 度 / km |

```go
// xt=1..7 对应水星..海王星；zn=0 黄经、1 黄纬、2 日心距（AU）。
fmt.Println(planet.WherePlanet(4, 2, 2460000.5))      // 木星日心距，AU
fmt.Println(planet.WherePlanetN(4, 2, 2460000.5, 12)) // 半解析快速表；Legacy 按 n 截断。
fmt.Println(planet.WherePlanet(8, 0, 2460000.5))      // xt 越界：NaN
```

```go
// 地球日心黄经（xt=-1）与行星日心黄经可放在同一口径下比较。
fmt.Println(planet.WherePlanet(-1, 0, 2460000.5), planet.WherePlanet(4, 0, 2460000.5))
```

## 参数与返回值约定

下面几条是七个包一致的口径；按能力汇总的返回值单位另见[通用能力与单位](#通用能力与单位)。

### 时标与民用时刻

公开 API 一律把 `time.Time` 当**民用时刻**（UTC 标签）使用：位置与物理接口内部先取 `date.UTC()` 再换算到 TT 求星历；升落、中天与站心地平量族额外读取 `date.Zone()` 参与地方时计算。

本库将 1972-01-01 之前的民用时间按 UT1 处理；1972 年以后用内置闰秒表，窗口末端之后按当前的 UTC 跟随政策处理。需要显式换算时用根包的 `astro.UT1FromUTC`、`astro.TTFromUTC` 与 `astro.DUT1`；图内/图注的时标声明与 UT1 口径见[时标声明](map-geojson.md#时标声明)。

### 单位与口径

角度一律为度；视直径与视半径为角秒；距离按函数名区分——`EarthDistance` / `SunDistance` 与 `planet.WherePlanet` 的 `zn=2` 用 AU，`planet.MoonAway` 用 km。

`PhaseAngle` 为度，`IlluminatedFraction` 与别名 `Phase` 为 `0–1`；`ApparentMagnitude` 为星等（无量纲，经典相位多项式）；`ApparentMagnitude2018` 用 Mallama & Hilton 2018 模型，水星、金星大相位角与土星环更准，火星不含自转与季节修正；升落、中天与所有事件搜索返回 `time.Time`，时区与输入一致。

事件结构里的历时字段（`TransitInfo.Duration`、`TransitInfo.InternalDuration`、`GalileanPhenomenonEvent.Duration`、`GalileanPhenomenonContact.Duration`）是 Go 的 `time.Duration`，不是儒略日或天数；`TransitInfo` 的 `Start` / `Greatest` / `End` / `InternalStart` / `InternalEnd` 保持调用者输入的时区。

地心量、站心量与距离是三个不同口径：`ApparentLo` / `ApparentBo` / `ApparentRa` / `ApparentDec` / `ApparentRaDec` 是地心视位置，`Altitude` / `Azimuth` / `Zenith` 是站心量，`HourAngle` / `ParallacticAngle` 是不含站心视差的地心量，`EarthDistance` / `SunDistance` 是地心几何距离。

### 零值、越界与哨兵错误

事件搜索在窗口内没有事件时返回零值或 `Valid=false` 的结构（`TransitInfo.Valid`、`GalileanPhenomenonEvent.Valid`、`GalileanPhenomenonContactEvent.Valid`），不返回错误；`TransitInfo` 在 `HasInternal` 为假时 `InternalStart` / `InternalEnd` 为零值，偏凌日没有内切。

`planet.WherePlanet` / `planet.WherePlanetN` 的 `xt` / `zn` 越界返回 `NaN` 而不 panic；`...N` 的 `n < 0` 用完整级数，半解析的非负 `n` 用同一张快速表，Legacy 按 `n` 截断。`RiseTime` / `SetTime` / `DownTime`（含 `N` 版）当天没有几何升/落时返回下面的哨兵错误，时刻为零值。

每个行星包各有一套极昼/极夜错误，名字里的 `NEVER_RISE` 指"当天永不升起"（极夜），`NEVER_SET` 指"当天永不落下"（极昼），`NEVER_DOWN` 是 `NEVER_SET` 的兼容别名。

| 包 | 永不升起 | 永不落下 | 落下别名 |
| --- | --- | --- | --- |
| `mercury` | `ERR_MERCURY_NEVER_RISE` | `ERR_MERCURY_NEVER_SET` | `ERR_MERCURY_NEVER_DOWN` |
| `venus` | `ERR_VENUS_NEVER_RISE` | `ERR_VENUS_NEVER_SET` | `ERR_VENUS_NEVER_DOWN` |
| `mars` | `ERR_MARS_NEVER_RISE` | `ERR_MARS_NEVER_SET` | `ERR_MARS_NEVER_DOWN` |
| `jupiter` | `ERR_JUPITER_NEVER_RISE` | `ERR_JUPITER_NEVER_SET` | `ERR_JUPITER_NEVER_DOWN` |
| `saturn` | `ERR_SATURN_NEVER_RISE` | `ERR_SATURN_NEVER_SET` | `ERR_SATURN_NEVER_DOWN` |
| `uranus` | `ERR_URANUS_NEVER_RISE` | `ERR_URANUS_NEVER_SET` | `ERR_URANUS_NEVER_DOWN` |
| `neptune` | `ERR_NEPTUNE_NEVER_RISE` | `ERR_NEPTUNE_NEVER_SET` | `ERR_NEPTUNE_NEVER_DOWN` |

```go
// 极昼/极夜：升落接口返回哨兵错误，时刻是 time.Time 零值。
rise, err := mercury.RiseTime(date, lon, lat, height, true)
switch {
case errors.Is(err, mercury.ERR_MERCURY_NEVER_RISE):
	fmt.Println("极夜：当天永不升起", rise.IsZero())
case errors.Is(err, mercury.ERR_MERCURY_NEVER_SET):
	fmt.Println("极昼：当天永不落下", rise.IsZero())
}
```

### 站心与地心

`Altitude` / `Azimuth` / `Zenith` / `HourAngle` / `ParallacticAngle` 使用 `date` 的时区做地方时计算，经度东正西负、纬度北正南负，`height` 是椭球高（米，不是海拔正高）。其中 `Altitude` / `Azimuth` / `Zenith` 计入站心视差，与 `ApparentRa` / `ApparentDec` 的地心视位置不是一个口径；`HourAngle` / `ParallacticAngle` 直接由地心视位置求得，不含站心视差，跨口径使用前先经[坐标工具](coord.md)换算。

### 精度与适用范围

行星包默认使用半解析模型，也可选择 Legacy 或自定义引擎。可用区间不代表全范围精度相同，Legacy 相对完整 VSOP87 的截断误差也不等于对 DE 星历的综合误差，详见[Legacy 截断误差](accuracy.md#legacy-截断误差)。快速表适合粗扫描；需要精修事件或比较外部年表时使用完整级数并统一时标与坐标口径。

`saturn.Ring` 与物理星历只影响盘面与掩带表达：土星环既不参与月掩接触计算、也不作为圆盘边界绘制。行星没有独立出图入口，月掩与土星环的出图见[月掩手册](occultation.md#月掩出图)。
