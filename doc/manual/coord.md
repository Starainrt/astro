# 坐标工具

[English](en/coord.md) | [返回 README](../../README.md)

`coord` 包提供天球坐标系转换与观测辅助计算，涵盖黄道、赤道、地平、银道及行星赤道坐标互转，以及站心视差、恒星时、岁差章动、角距离、大气折射、大气质量与视差角。

## 目录

- [基本示例](#基本示例)
- [主要功能](#主要功能)
  - [大气折射与大气质量](#大气折射与大气质量)
  - [站心改正、岁差章动与银道坐标](#站心改正岁差章动与银道坐标)
  - [行星赤道坐标](#行星赤道坐标)
  - [观测点高度](#观测点高度)
- [API 概览](#api-概览)
  - [大气质量](#大气质量)
- [约定与限制](#约定与限制)

## 基本示例

将黄道坐标转换为当日赤道坐标与地平坐标，并计算格林尼治视恒星时（GAST）：

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/coord"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2026, 4, 27, 10, 30, 45, 0, cst)

	eq := coord.EclipticToEquatorial(date, 139.686111, 4.875278)
	hz := coord.EquatorialToHorizontal(date, eq.RA, eq.Dec, 115, 40)

	fmt.Printf("赤道: RA=%.4f° Dec=%.4f°  地平: 方位=%.2f° 高度=%.2f°  GAST=%.4f h\n",
		eq.RA, eq.Dec, hz.Azimuth, hz.Altitude, coord.ApparentSiderealTime(date))
}
```

输出：

```text
赤道: RA=143.7222° Dec=19.5351°  地平: 方位=43.47° 高度=-17.69°  GAST=16.8526 h
```

除恒星时（单位**小时**）与高程（单位**米**）外，角度单位均为**度**；经度东正西负，方位角自正北 `0°` 向东增加。

## 主要功能

### 大气折射与大气质量

- 真高度与视高度互换使用 `ApparentAltitude` / `TrueAltitude`（参数含气压 hPa 与气温 ℃）；一步由赤道坐标求含折射地平坐标使用 `EquatorialToApparentHorizontal`。
- 由真高度自动叠加折射并计算大气质量使用 `AirmassKastenYoungFromTrueAltitude` / `AirmassPickeringFromTrueAltitude`。

```go
appAlt := coord.ApparentAltitude(10, 1010, 10)
ky := coord.AirmassKastenYoungFromTrueAltitude(10, 1010, 10)
fmt.Printf("真高度 10° → 视高度 %.4f°（折射 +%.2f'），大气质量 %.3f\n",
	appAlt, (appAlt-10)*60, ky)
```

输出：

```text
真高度 10° → 视高度 10.0901°（折射 +5.41'），大气质量 5.540
```

### 站心改正、岁差章动与银道坐标

- **站心视差**：`TopocentricEquatorial(date, ra, dec, lon, lat, distanceAU, heightM)` 与 `TopocentricEcliptic`（对月球等近地天体尤为重要）。
- **岁差与章动**：`Precess(fromDate, toDate, ra, dec)` 仅执行岁差旋转（不含自行）；`EclipticObliquity(date, withNutation)` 求平/真黄赤交角；章动提供 `Nutation2000B`、`Nutation1980` 及显式高精度入口 `Nutation2000A` / `MeanObliquity2006`。
- **银道坐标**：`EquatorialToGalactic` / `GalacticToEquatorial` 对 ICRS 赤道坐标执行固定矩阵旋转。
- **角距离与视差角**：`AngularSeparation(ra1, dec1, ra2, dec2)` 求大圆角距，`ParallacticAngle` 求天顶方向角。

### 行星赤道坐标

`ToBodyEquatorial(body, date, ra, dec, origin)` 与 `FromBodyEquatorial` 在 ICRF/J2000 赤道方向与行星（水星至海王星）赤道坐标之间执行纯方向旋转：

```go
t := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
marsEq, err := coord.ToBodyEquatorial(basic.SkyMars, t, 101.287155, -16.716116, coord.BodyEquatorOriginEquinox)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("天狼星在火星赤道坐标(春分零点): RA=%.2f°  Dec=%.2f°\n", marsEq.RA, marsEq.Dec)
```

输出：

```text
天狼星在火星赤道坐标(春分零点): RA=174.63°  Dec=-44.02°
```

- `origin` 支持 `BodyEquatorOriginIAUNode`（零值，行星赤道与 ICRF 赤道升交点 Q）与 `BodyEquatorOriginEquinox`（行星春分点；水星春分点退化会返回错误，须使用 Q 零点）。
- 该接口不加入光行时、视差或地平几何；行星表面地平观测请使用[跨行星表面观测](phenomena.md#跨行星表面观测)。

### 观测点高度

本库绝大多数观测接口（`sun`、`moon`、行星、`star`、`eclipse`、`orbit`、`satellite`、`coord`）的 `height` 参数定义为 WGS84 **椭球高（大地高，单位米）**，而非海拔正高 `H`。

若仅知海拔正高 `H`，应结合当地大地水准面差距 `N`（如 EGM2008 / EGM96 模型）换算：

```text
height = H + N
```

在常规日出日落场景中几十米高程差影响极小；但在精密月掩计时、日食带边界与高海拔观测中应使用椭球高。

## API 概览

| 分类 | 入口 | 说明 |
| --- | --- | --- |
| 黄道 ↔ 赤道 | `EclipticToEquatorial` / `EquatorialToEcliptic`（及 `...ByObliquity`） | 按瞬时真交角或指定交角互转 |
| 赤道 ↔ 地平 | `EquatorialToHorizontal` / `EquatorialToApparentHorizontal` | 自动计算恒星时；后者叠加大气折射 |
| | `EquatorialToHorizontalByLocalSiderealTime` / `HorizontalToEquatorialByLocalSiderealTime` | 手工传入地方恒星时（小时）互转 |
| | `HourAngleDeclinationToHorizontal` / `HorizontalToHourAngleDeclination` | 时角与赤纬和地平坐标互转 |
| 站心坐标 | `TopocentricEquatorial` / `TopocentricEcliptic` | 叠加站心视差（`distanceAU` 为 AU，`height` 为椭球高米） |
| 恒星时 | `MeanSiderealTime` / `ApparentSiderealTime` / `HourAngle` | GMST / GAST（小时，IAU 2006），瞬时时角（度） |
| 岁差与章动 | `Precess` / `EclipticObliquity` / `Nutation2000B` / `Nutation1980` / `Nutation2000A` / `MeanObliquity2006` | 岁差旋转、黄赤交角与各代章动模型 |
| 银道与行星赤道 | `EquatorialToGalactic` / `GalacticToEquatorial` / `ToBodyEquatorial` / `FromBodyEquatorial` | ICRS ↔ 银道；ICRF/J2000 ↔ 行星赤道 |
| 角距与视差角 | `AngularSeparation` / `ParallacticAngle` / `ParallacticAngleByHourAngle` | 大圆角距、天顶方向角 `(−180°, 180°]` |
| 大气折射 | `ApparentAltitude` / `TrueAltitude` / `AtmosphericRefractionFromTrueAltitude` / `AtmosphericRefractionFromApparentAltitude` | Saemundsson 折射模型正反解 |

### 大气质量

| 入口 | 输入 | 说明 |
| --- | --- | --- |
| `AirmassPlaneParallelFromTrueAltitude` | 真高度角（度） | 平行平板近似 `sec(z)` |
| `AirmassKastenYoungFromTrueAltitude` / `AirmassPickeringFromTrueAltitude` | 真高度角 + 气压 + 气温 | 先执行折射修正再计算大气质量 |
| `AirmassKastenYoungFromApparentAltitude` / `AirmassPickeringFromApparentAltitude` | 视高度角（度） | 不再叠加折射（亦见 [formula](formula.md#大气质量模型)） |

## 约定与限制

- **角度归一化**：`RA`、`Lon`、`Azimuth`、`HourAngle` 归约至 `[0°, 360°)`，`Dec`、`Lat`、`Altitude` 位于 `[−90°, 90°]`，`ParallacticAngle` 位于 `(−180°, 180°]`。`TopocentricEquatorial` 仅叠加视差改正，不对 `RA` 做 360° 折回。
- **时标口径**：恒星时基于 UT1（IAU 2006 ERA），岁差与章动基于 TT；详见[时标](timescale.md)。
- **大气折射区间**：仅在真高度角 `(-5°, 90°)` 内生效，区间外折射量为 `0`；气压非正、气温低于绝对零度或逆解不存在时返回 `NaN`。
- **行星制图坐标辅助**：`basic.PlanetographicToEastLongitude` / `EastToPlanetographicLongitude` 与 `PlanetocentricToPlanetodeticLatitude` / `PlanetodeticToPlanetocentricLatitude` 提供行星经纬度口径转换。
- **IAU 2006/2000A 矩阵**：完整旋转矩阵可通过 `basic.BiasPrecessionNutation2006` 显式调用，默认高层链路不自动启用。
