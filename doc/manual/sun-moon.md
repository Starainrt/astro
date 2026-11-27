# 太阳与月亮

[English](en/sun-moon.md) | [返回 README](../../README.md)

本手册涵盖太阳与月亮的常用计算：日出日落、晨昏蒙影、月出月落、地平与赤道/黄道位置、月相与朔望弦、近远地点、轨道交点、最大赤纬、天平动、地球近日与远日点，以及面向资源受限环境的 `lite/sun`、`lite/moon` 轻量链路。

日月食见[日月食与凌日](eclipse.md)，月掩见[月掩](occultation.md)，通用坐标系与折射换算见[坐标工具](coord.md)。

## 目录

- [基本示例](#基本示例)
- [主要功能](#主要功能)
  - [日出日落月出月落](#日出日落月出月落)
  - [日月位置](#日月位置)
  - [月相](#月相)
  - [近远地点、轨道交点与地球近日点](#近远地点轨道交点与地球近日点)
  - [天平动与亮边位置角](#天平动与亮边位置角)
  - [轻量链路（lite/sun 与 lite/moon）](#轻量链路litesun-与-litemoon)
- [观测角语义](#观测角语义)
- [API 概览](#api-概览)
- [约定与限制](#约定与限制)

## 基本示例

计算 2026 年 1 月 1 日西安的民用晨昏蒙影、日出日落、上中天以及月出月落时刻：

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
	date := time.Date(2026, 1, 1, 12, 0, 0, 0, cst)
	lon, lat, height := 108.93, 34.27, 0.0 // 西安（东经、北纬、椭球高米）

	twM, err := sun.MorningTwilight(date, lon, lat, -6) // -6° 民用、-12° 航海、-18° 天文
	if err != nil {
		log.Fatal(err)
	}
	sRise, err := sun.RiseTime(date, lon, lat, height, true)
	if err != nil {
		log.Fatal(err)
	}
	sCulm := sun.CulminationTime(date, lon)
	sSet, err := sun.SetTime(date, lon, lat, height, true)
	if err != nil {
		log.Fatal(err)
	}
	twE, err := sun.EveningTwilight(date, lon, lat, -6)
	if err != nil {
		log.Fatal(err)
	}

	mRise, err := moon.RiseTime(date, lon, lat, height, true)
	if err != nil {
		log.Fatal(err)
	}
	mCulm := moon.CulminationTime(date, lon, lat)
	mSet, err := moon.SetTime(date, lon, lat, height, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("西安 2026-01-01")
	fmt.Printf("民用晨光: %s  日出: %s  中天: %s  日落: %s  民用昏影: %s\n",
		twM.Format("15:04:05"), sRise.Format("15:04:05"), sCulm.Format("15:04:05"), sSet.Format("15:04:05"), twE.Format("15:04:05"))
	fmt.Printf("月出: %s  月中天: %s  月落: %s\n",
		mRise.Format("15:04:05"), mCulm.Format("15:04:05"), mSet.Format("15:04:05"))
}
```

输出：

```text
西安 2026-01-01
民用晨光: 07:22:35  日出: 07:49:59  中天: 12:47:50  日落: 17:45:31  民用昏影: 18:12:55
月出: 15:14:55  月中天: 22:56:17  月落: 05:28:16
```

## 主要功能

以下片段沿用西安坐标（`108.93°E, 34.27°N`）与 `date := 2026-01-01 12:00:00 CST`。

### 日出日落月出月落

- 参数 `aero` 为 `true` 时，按标准大气折射（`1010 hPa / 10 ℃`）与实时日月视半径计算**上缘**过地平线时刻；为 `false` 时按几何中心过 `0°` 地平线计算。
- 需要传入实测气压与气温时，使用 `sun.RiseTimeWithAtmosphere` / `SetTimeWithAtmosphere` 及 `moon` 同名函数。
- 月出与月落按指定当地自然日独立计算，升降时刻之间不一定连续（如上例 `05:28` 月落、`15:14` 月出）；当日无事件时返回 `moon.ERR_NOT_TODAY`，跨日查询需将 `date` 顺延一天。
- 极昼或极夜无升落时返回哨兵错误 `ERR_SUN_NEVER_RISE` / `ERR_SUN_NEVER_SET`（月球对应 `ERR_MOON_NEVER_*`），使用 `errors.Is` 判断。

### 日月位置

计算视赤经赤纬、高度方位、视直径、距离与真太阳时：

```go
sRA, sDec := sun.ApparentRaDec(date)
mRA, mDec := moon.ApparentRaDec(date, lon, lat) // 站心视位置
mgRA, mgDec := moon.GeocentricApparentRaDec(date)

fmt.Printf("太阳: 视赤经 %s  视赤纬 %s  高度 %.2f°  方位 %.2f°  日地距 %.6f AU  视直径 %.1f\"\n",
	tools.Format(sRA/15, 1), tools.Format(sDec, 0),
	sun.Altitude(date, lon, lat), sun.Azimuth(date, lon, lat),
	sun.EarthDistance(date), sun.Diameter(date))
fmt.Printf("太阳真太阳时: %s  均时差: %.2f min\n",
	sun.ApparentSolarTime(date, lon).Format("15:04:05"), sun.EquationTime(date)*60)
fmt.Printf("月球(站心): 视赤经 %s  视赤纬 %s  高度 %.2f°  方位 %.2f°  地月距 %.1f km  视直径 %.1f\"\n",
	tools.Format(mRA/15, 1), tools.Format(mDec, 0),
	moon.Altitude(date, lon, lat), moon.Azimuth(date, lon, lat),
	moon.EarthDistance(date), moon.Diameter(date))
fmt.Printf("月球视差(站心-地心): dRA=%+.3f°  dDec=%+.3f°\n", mRA-mgRA, mDec-mgDec)
```

输出：

```text
太阳: 视赤经 18h46m42.87s  视赤纬 -23°0′13.25″  高度 31.61°  方位 167.10°  日地距 0.983324 AU  视直径 1951.0"
太阳真太阳时: 11:12:18  均时差: -3.40 min
月球(站心): 视赤经 4h28m6.57s  视赤纬 25°58′57.50″  高度 -26.01°  方位 22.73°  地月距 360486.1 km  视直径 1986.5"
月球视差(站心-地心): dRA=+0.363°  dDec=-0.848°
```

- 月球站心与地心位置偏差可达约 `1°`，地面观测指向应使用 `moon.ApparentRaDec(date, lon, lat)`；地心年表对照使用 `moon.GeocentricApparentRaDec(date)`。
- 距离单位：`sun.EarthDistance` 为 **AU**，`moon.EarthDistance` 为 **千米（km）**。

### 月相

查询受照比例、中文月相名称与朔望弦时刻（英文别名与拼音名 `NextShuoYue`/`NextWangYue`/`NextShangXianYue`/`NextXiaXianYue` 等价）：

```go
fmt.Printf("月相: %s（照亮 %.1f%%）\n", moon.PhaseDesc(date), moon.Phase(date)*100)
fmt.Printf("下次满月: %s  下弦: %s  朔月: %s  上弦: %s\n",
	moon.NextFullMoon(date).Format("01-02 15:04"),
	moon.NextLastQuarter(date).Format("01-02 15:04"),
	moon.NextNewMoon(date).Format("01-02 15:04"),
	moon.NextFirstQuarter(date).Format("01-02 15:04"))
```

输出：

```text
月相: 盈凸月（照亮 92.5%）
下次满月: 01-03 18:02  下弦: 01-10 23:48  朔月: 01-19 03:51  上弦: 01-26 12:47
```

- 同步提供 `Last*`（上一次）与 `Closest*`（最近一次），返回时刻保持输入时区。
- 行星合月（赤经合）使用 `moon.ClosestConjunctionWithPlanet(date, moon.ConjunctionJupiter)`。

### 近远地点、轨道交点与地球近日点

```go
peris := moon.PerigeesInMonth(2026, time.January)
apos := moon.ApogeesInMonth(2026, time.January)
fmt.Printf("2026-01 近地点: %s UTC (%.0f km)，远地点: %s UTC (%.0f km)\n",
	peris[0].Time.Format("01-02 15:04"), peris[0].Distance,
	apos[0].Time.Format("01-02 15:04"), apos[0].Distance)

ep := earth.Perihelion(2026)
ea := earth.Aphelion(2026)
fmt.Printf("2026 地球近日点: %s UTC (%.6f AU)  远日点: %s UTC (%.6f AU)\n",
	ep.Time.Format("01-02 15:04"), ep.Distance, ea.Time.Format("01-02 15:04"), ea.Distance)
```

输出：

```text
2026-01 近地点: 01-01 21:44 UTC (360348 km)，远地点: 01-13 20:47 UTC (405438 km)
2026 地球近日点: 01-03 17:15 UTC (0.983302 AU)  远日点: 07-06 17:30 UTC (1.016644 AU)
```

月球升降交点黄经使用 `moon.AscendingNode(date)` / `DescendingNode(date)`；月赤纬极值使用 `moon.MaximumNorthDeclinationsInMonth` / `NextMaximumNorthDeclination`（南赤纬对称）。

### 天平动与亮边位置角

`TopocentricPhysical` 返回站心天平动与自转轴位置角，`TopocentricBrightLimbPositionAngle` 返回亮边中心位置角（以月面北点为 `0°`，向东增加）：

```go
p := moon.TopocentricPhysical(date, lon, lat, 0)
fmt.Printf("站心天平动: 经度 %+.2f°  纬度 %+.2f°  自转轴PA %+.2f°  亮边PA %.2f°\n",
	p.LibrationLongitude, p.LibrationLatitude, p.PositionAngle,
	moon.TopocentricBrightLimbPositionAngle(date, lon, lat, 0))
```

输出：

```text
站心天平动: 经度 -0.78°  纬度 -5.66°  自转轴PA -8.88°  亮边PA 267.86°
```

地心版本为 `moon.Physical` 与 `moon.BrightLimbPositionAngle`；太阳日面物理量（P / B0 / L0）为 `sun.Physical`。

### 轻量链路（`lite/sun` 与 `lite/moon`）

`lite/sun` 与 `lite/moon` 为独立简化公式，不加载主链表、不随全局星历模型切换，适用于资源受限环境：

```go
lRise, _ := litesun.RiseTime(date, lon, lat, 0, true)
fmt.Printf("Lite 对比: 日出 %s (主链 %s)  月相 %.2f%%  月龄 %.2f 天\n",
	lRise.Format("15:04:05"), sRise.Format("15:04:05"),
	litemoon.Phase(date)*100, litemoon.PhaseAge(date))
```

输出：

```text
Lite 对比: 日出 07:49:59 (主链 07:49:59)  月相 92.63%  月龄 12.18 天
```

`litemoon.PhaseAge`（月龄，天）仅在轻量链路提供；`litesun.Distance` 对应主链 `sun.EarthDistance`。误差与性能对比见[精度与性能](accuracy.md#lite-轻量链路)。

## 观测角语义

- **`Altitude` / `Zenith`**：几何高度角（地平 `0°`、天顶 `+90°`）与天顶距（满足 `Altitude + Zenith = 90°`），不含大气折射；含折射版本为 `ApparentAltitude` / `ApparentZenith`（传入气压 hPa 与气温 ℃）。
- **`Azimuth`**：方位角，正北 `0°`，向东顺时针增加，取值 `[0°, 360°)`。
- **`HourAngle`**：时角，上中天 `0°`，向西增加，取值 `[0°, 360°)`。
- **`ParallacticAngle`**：天顶方向视差角，带符号（度）。

## API 概览

| 分组 | `sun` 包 | `moon` 包 | 单位与口径 |
| --- | --- | --- | --- |
| 坐标 | `TrueLo`、`ApparentLo`、`TrueBo`、`ApparentRaDec`、`EclipticObliquity` | `TrueLo`、`ApparentLo`、`TrueBo`、`TrueRaDec`、`GeocentricApparentRaDec`、`ApparentRaDec` | 度；`moon.ApparentRaDec` 为站心 |
| 升落与中天 | `RiseTime`、`SetTime`、`CulminationTime`、`MorningTwilight`、`EveningTwilight` | `RiseTime`、`SetTime`、`CulminationTime` | 另提供 `...WithAtmosphere` 实测气象版 |
| 地平量 | `Altitude`、`ApparentAltitude`、`Azimuth`、`Zenith`、`HourAngle`、`ParallacticAngle` | 同左 | 度 |
| 距离与视径 | `EarthDistance`（AU）、`Diameter`、`Semidiameter`（角秒） | `EarthDistance`（km）、`Diameter`、`Semidiameter`（角秒） | 太阳为 AU，月球为 km |
| 太阳时 / 月相 | `ApparentSolarTime`、`EquationTime`（小时） | `Phase`、`PhaseDesc`、`SunMoonLoDiff`、`Next/Last/ClosestNewMoon`（`ShuoYue` 等） | 含拼音与英文双命名 |
| 物理与轨道 | `Physical`（P/B0/L0）；`earth.Perihelion` / `Aphelion` / `EarthEccentricity` | `Physical`、`TopocentricPhysical`、`BrightLimbPositionAngle`、`PerigeesInMonth`、`ApogeesInMonth`、`AscendingNode`、`MaximumNorthDeclinationsInMonth` | 角度为度，月距为 km |

## 约定与限制

- **`...N` 变体**：瞬时求值接口大多提供 `...N(date, ..., n)` 后缀版本。`n < 0` 使用完整级数；默认半解析模型下所有 `n >= 0` 使用同一张快速表（仅适合粗搜索），仅 `ModelLegacy` 按 `n` 截断。
- **时标与高程**：时间参数为民用时刻（1972-01-01 之前按 UT1，见[时标](timescale.md)）；`height` 为 WGS84 椭球高（米），见[观测点高度](coord.md#观测点高度)。`sun.ApparentSolarTime` 返回地方真太阳时读数（合成经度时区），不应直接传回星历接口。
- **哨兵错误与兼容别名**：极区无升落返回 `ERR_SUN_NEVER_RISE` / `ERR_SUN_NEVER_SET`（月球为 `ERR_MOON_NEVER_*`），无朦影返回 `ERR_TWILIGHT_NOT_EXISTS`，当日无月出/月落返回 `ERR_NOT_TODAY`。`DownTime` 与 `ERR_*_NEVER_DOWN` 为废弃兼容别名，新代码应使用 `SetTime` 与 `ERR_*_NEVER_SET`。
