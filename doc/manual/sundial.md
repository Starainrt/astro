# 日晷与真太阳时

[English](en/sundial.md) | [返回 README](../../README.md)

`sundial` 包提供地方真太阳时、地方平太阳时、均时差与太阳时角计算，以及平面日晷（水平、垂直、赤道及任意倾角与朝向）的几何参数、落影坐标、受光时角区间、等时线与赤纬曲线求解。

## 目录

- [基本示例](#基本示例)
- [主要功能](#主要功能)
  - [可用受光区间与赤纬曲线](#可用受光区间与赤纬曲线)
  - [平太阳时线与区时线](#平太阳时线与区时线)
  - [赤道、垂直与任意平面日晷](#赤道垂直与任意平面日晷)
- [API 概览](#api-概览)
- [约定与限制](#约定与限制)

## 基本示例

计算 2026 年夏至 09:30（北京时间）上海的真/平太阳时、均时差，以及 10 cm 极轴晷针在水平日晷上的落影坐标：

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/sundial"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2026, 6, 21, 9, 30, 0, 0, cst)
	lon, lat := 121.4737, 31.2304 // 上海（东经/北纬，度）

	ts := sundial.TrueSolarTime(date, lon)
	ms := sundial.MeanSolarTime(date, lon)
	ha := sundial.HourAngle(date, lon)

	dial := sundial.HorizontalDial(lat, 10) // 晷针长度 10 cm，落影坐标单位同为 cm
	sh := dial.ShadowPointAt(date, lon)

	fmt.Printf("上海夏至 09:30 CST: 真太阳时 %s  平太阳时 %s  均时差 %s\n",
		ts.Format("15:04:05"), ms.Format("15:04:05"), ts.Sub(ms).Round(time.Second))
	fmt.Printf("时角 %.2f°  水平日晷落影(晷针10cm): x=%.2f cm (东), y=%.2f cm (北)  受光=%v\n",
		ha, sh.X, sh.Y, sh.Illuminated)
}
```

输出：

```text
上海夏至 09:30 CST: 真太阳时 09:34:10  平太阳时 09:35:53  均时差 -1m43s
时角 -36.46°  水平日晷落影(晷针10cm): x=-6.51 cm (东), y=0.51 cm (北)  受光=true
```

水平日晷坐标约定为 **x 轴向东、y 轴向北**，坐标单位与 `StylusLength` 一致。绘制有效落影时以 `sh.Illuminated` 为最终判据（同时满足太阳在地平线上且盘面朝向太阳）。

## 主要功能

### 可用受光区间与赤纬曲线

`DeclinationCurveAt(date, stepDeg)`（或指定赤纬 `DeclinationCurve(declDeg, stepDeg)`）计算某日或某赤纬下日晷的可用时角区间与影子尖端采样点列：

```go
dial := sundial.HorizontalDial(31.2304, 10)
segs := dial.DeclinationCurveAt(date, 1.0) // 按 1° 时角步长采样
for _, seg := range segs {
	fmt.Printf("夏至可用时角区间: %.1f° .. %.1f°（赤纬 %.2f°，采样 %d 点）\n",
		seg.Interval.Start, seg.Interval.End, seg.Declination, len(seg.Samples))
}
```

输出：

```text
夏至可用时角区间: -105.2° .. 105.2°（赤纬 23.44°，采样 211 点）
```

仅需时角区间时调用 `dial.IlluminatedHourAngleIntervals(declDeg)`（或仅计盘面几何受光的 `PlaneIlluminatedHourAngleIntervals`）。跨零点的区间会拆分为多段，应遍历返回的切片。

### 平太阳时线与区时线

- **区时线**：`dial.ZoneTimePoint(date, lon, 9.5)` / `dial.ZoneTimeLine(dates, lon, 9.5)` 读取 `date` 的年月日与时区，并以 `zoneTimeHours`（如 `9.5` 表示 09:30）作为钟面时刻。
- **地方平太阳时线**：`dial.MeanSolarTimePoint(meanDate, 9.5)` / `dial.MeanSolarTimeLine(meanDates, 9.5)` 接收目标地的地方平太阳时（由 `sundial.MeanSolarTime` 取得）。
- **水平日晷时线角**：`sundial.HorizontalHourLineAngle(lat, hourAngleDeg)` 或 `HorizontalHourLineAngleAt(date, lon, lat)`。

### 赤道、垂直与任意平面日晷

| 构造器 | 日晷类型 | `PlaneNormalAzimuth` | `PlaneNormalZenithDistance` |
| --- | --- | --- | --- |
| `HorizontalDial(lat, len)` | 水平日晷 | `180°` | `0°`（法线指向天顶） |
| `EquatorialNorthDial(lat, len)` | 赤道日晷北面（北半球春夏） | `0°` | `90° − lat` |
| `EquatorialSouthDial(lat, len)` | 赤道日晷南面（北半球秋冬） | `180°` | `90° + lat` |
| `VerticalDial(lat, normalAz, len)` | 垂直日晷 | 盘面**法线**方位角（朝南墙 `180°`，朝东墙 `90°`） | `90°` |

任意倾角与朝向的平面日晷可直接构造 `PlanarDial`，并通过 `dial.Geometry()` 获取日晷中心坐标（`CenterX`、`CenterY`）、极轴晷针长度（`PolarStylusLength`）与极轴倾角（`PolarStylusPlaneAngle`）。

## API 概览

| 分组 | 入口 | 说明 |
| --- | --- | --- |
| 太阳时 | `TrueSolarTime` / `MeanSolarTime` | 地方真/平太阳时读数（合成 `LTZ` 时区） |
| 时角 | `HourAngle` / `MeanSolarHourAngle` / `ZoneTimeHourAngle` | 太阳时角（度，上午负、下午正） |
| 水平时线 | `HorizontalHourLineAngle` / `HorizontalHourLineAngleAt` | 水平日晷时线相对午线夹角（度） |
| 日晷构造 | `HorizontalDial` / `VerticalDial` / `EquatorialNorthDial` / `EquatorialSouthDial` / `PlanarDial` | 平面日晷模型 |
| 几何与落影 | `Geometry` / `ShadowPointAt` / `ShadowPointByHourAngleDeclination` | 日晷中心与落影坐标 |
| 受光与曲线 | `IlluminatedHourAngleIntervals` / `PlaneIlluminatedHourAngleIntervals` / `DeclinationCurve` / `DeclinationCurveAt` | 可用时角区间与赤纬采样点列 |
| 时间线 | `MeanSolarTimePoint` / `MeanSolarTimeLine` / `ZoneTimePoint` / `ZoneTimeLine` | 平太阳时线与区时线采样 |

## 约定与限制

- **时角符号**：本包时角与水平时线角约定为**上午负、下午正**（`[-180°, 180°]`）。
- **太阳时返回值**：`TrueSolarTime` 与 `MeanSolarTime` 返回带有按经度合成偏移时区的太阳时读数，不应直接传回要求民用时刻的星历接口。
- **退化日晷**：当盘面法线与极轴垂直时（如纬度 `45°`、法线方位 `180°`、法线天顶距 `45°`），日晷中心位于无穷远，`Geometry().HasFiniteCenter` 为 `false`，`CenterX` / `CenterY` / `PolarStylusLength` 为 `NaN`；`ShadowPointAt` 仍可正常求落影点。
- **参数语义**：`VerticalDial` 的 `planeNormalAzimuth` 为墙面**法线**方位角（正北 `0°` 顺时针，朝南墙为 `180°`）；`ZoneTimePoint` / `ZoneTimeLine` 忽略输入 `date` 的时分秒；判断有效落影须以 `Illuminated` 而非 `PlaneIlluminated` 为准。

相关手册：[太阳与月亮](sun-moon.md) | [坐标工具](coord.md) | [时标](timescale.md)
