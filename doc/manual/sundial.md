# 日晷与真太阳时

[English](en/sundial.md) | [返回 README](../../README.md)

> 本手册的完整示例以仓库根目录为工作目录执行。

`sundial` 把 `sun` 包的真太阳时、太阳时角与日晷绘制所需的几何量集中在一起，不引入另一套算法。日晷部分按经典平面日晷模型工作：一根指向天极的极轴晷针，其影子投在任意平面上；坐标约定由构造器给定，水平日晷是 **x 轴向东、y 轴向北**。

## 目录

- [计算水平日晷的影子位置](#计算水平日晷的影子位置)
- [API 参考](#api-参考)
  - [真太阳时、平太阳时与均时差](#真太阳时平太阳时与均时差)
  - [时角](#时角)
  - [水平日晷时线角](#水平日晷时线角)
  - [平面日晷核心](#平面日晷核心)
  - [盘面受光区间](#盘面受光区间)
  - [时间线与赤纬曲线](#时间线与赤纬曲线)
  - [赤道、水平与垂直日晷特例](#赤道水平与垂直日晷特例)
  - [返回结构](#返回结构)
  - [完整示例](#完整示例)
  - [综合示例：一天的可用时角与等时线](#综合示例一天的可用时角与等时线)
  - [常见坑](#常见坑)
- [常用场景](#常用场景)
  - [真太阳时与钟表时间的差](#真太阳时与钟表时间的差)
  - [水平日晷的时线角与落影](#水平日晷的时线角与落影)
  - [盘面受光区间与等时线](#盘面受光区间与等时线)
  - [退化情形与各面日晷口径](#退化情形与各面日晷口径)
- [参数与返回值约定](#参数与返回值约定)
- [相关手册](#相关手册)

## 计算水平日晷的影子位置

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
	lon, lat := 121.4737, 31.2304
	fmt.Println(sundial.TrueSolarTime(date, lon))
	fmt.Println(sundial.HourAngle(date, lon))
	dial := sundial.HorizontalDial(lat, 10)
	shadow := dial.ShadowPointAt(date, lon)
	if !shadow.Illuminated {
		fmt.Println("no illuminated shadow")
		return
	}
	fmt.Printf("x=%.6f y=%.6f\n", shadow.X, shadow.Y)
}
```

`HorizontalDial` 的长度参数与返回坐标使用同一单位；例如晷针长度填厘米，影子坐标也是厘米。x 轴向东、y 轴向北，`Illuminated` 为真时才有可用的受光落影。

## API 参考

| 分组 | 入口 | 用途 | 单位与口径 |
| --- | --- | --- | --- |
| 真/平太阳时 | `TrueSolarTime` / `MeanSolarTime` | 该绝对时刻在指定经度上的地方真/平太阳时 | 返回地方太阳时读数；经度东正西负（度） |
| 太阳时角 | `HourAngle` | 真太阳时角 | 度，上午为负、下午为正 |
| 时角换算 | `MeanSolarHourAngle` / `ZoneTimeHourAngle` | 把地方平太阳时或区时钟面读数换成视太阳时角 | 小时数与经度（度） |
| 水平时线角 | `HorizontalHourLineAngle` / `HorizontalHourLineAngleAt` | 水平日晷时线相对午线的角度 | 度 |
| 平面日晷核心 | `PlanarDial`（字段）+ `Geometry` / `ShadowPointByHourAngleDeclination` / `ShadowPointAt` | 任意平面的几何量与落影点 | 坐标与晷针长度为同一长度单位 |
| 盘面受光 | `PlaneIlluminatedHourAngleIntervals` / `IlluminatedHourAngleIntervals` | 盘面受光时角区间与最终可用时角区间 | 度，`[-180, 180]` |
| 时间线 | `MeanSolarTimePoint` / `ZoneTimePoint` / `MeanSolarTimeLine` / `ZoneTimeLine` | 把平太阳时线或区时线直接接到日晷几何 | 返回 `PlanarShadowPoint` / `TimeLineSample` |
| 赤纬曲线 | `DeclinationCurve` / `DeclinationCurveAt` | 按赤纬或日期生成分段采样点列 | 时角步长单位为度 |
| 特例构造 | `EquatorialNorthDial` / `EquatorialSouthDial` / `HorizontalDial` / `VerticalDial` | 赤道（南北面）、水平、垂直日晷 | 纬度与法线方位角单位为度 |

以下片段省略公共前置变量：`date`（时刻，民用时标）、`lon`（经度，东正西负，度）、`lat`（纬度，度）。

### 真太阳时、平太阳时与均时差

```go
fmt.Println(sundial.TrueSolarTime(date, lon))
fmt.Println(sundial.MeanSolarTime(date, lon))
fmt.Println(sundial.TrueSolarTime(date, lon).Sub(sundial.MeanSolarTime(date, lon))) // 均时差
```

三者共用 `sun` 包的口径：`TrueSolarTime` 是视太阳时，`MeanSolarTime` 是地方平太阳时，两者之差即均时差（真 − 平）。

### 时角

```go
fmt.Println(sundial.HourAngle(date, lon))          // 真太阳时角，上午为负
fmt.Println(sundial.MeanSolarHourAngle(date, 9.5)) // 地方平太阳时 9:30 对应的时角
fmt.Println(sundial.ZoneTimeHourAngle(date, lon, 9.5)) // 区时钟面 9:30 对应的时角
```

`HourAngle` 直接用绝对时刻求时角；后两者用于"给定钟面读数求落影方向"，区别在于一个按地方平太阳时、一个按区时。

### 水平日晷时线角

```go
fmt.Println(sundial.HorizontalHourLineAngle(31.2304, -45))
fmt.Println(sundial.HorizontalHourLineAngleAt(date, lon, 31.2304))
```

前者给定纬度与带符号时角，后者直接用时刻与经纬度求当前时线角；返回值是时线相对午线的夹角。

### 平面日晷核心

```go
dial := sundial.PlanarDial{
	Latitude: 31.2304, PlaneNormalAzimuth: 180, PlaneNormalZenithDistance: 90, StylusLength: 10,
}
g := dial.Geometry()
fmt.Println(g.HasFiniteCenter, g.PolarStylusLength, g.PolarStylusPlaneAngle)
p := dial.ShadowPointByHourAngleDeclination(-45, 23.44)
fmt.Println(p.X, p.Y, p.Illuminated)
```

`PlanarDial` 的四个字段分别是纬度、盘面法线方位角、法线天顶距与晷针长度。`Geometry` 给出日晷中心（极轴晷针固定点）、极轴晷针长度以及它与盘面的夹角；`ShadowPointByHourAngleDeclination` 给定带符号时角与太阳赤纬求落影点，`ShadowPointAt` 则直接用时刻与经度。

### 盘面受光区间

```go
for _, iv := range dial.PlaneIlluminatedHourAngleIntervals(23.44) {
	fmt.Println(iv.Start, iv.End)
}
for _, iv := range dial.IlluminatedHourAngleIntervals(23.44) {
	fmt.Println(iv.Start, iv.End)
}
```

前者只回答"盘面是否朝向太阳"（几何受光），后者叠加太阳在地平线以上与盘面朝向两个条件，给出**最终可用**的时角区间；区间约定在 `[-180, 180]` 且 `Start <= End`。

### 时间线与赤纬曲线

```go
dates := []time.Time{date, date.Add(30 * time.Minute), date.Add(time.Hour)}
fmt.Println(len(dial.MeanSolarTimeLine(dates, 9.5)))
segs := dial.DeclinationCurve(23.44, 1.0)
segsAt := dial.DeclinationCurveAt(date, 1.0)
fmt.Println(len(segs), len(segsAt))
```

时间线把"地方平太阳时线上的等时刻点"直接投影成 `TimeLineSample`；赤纬曲线按固定赤纬或当日赤纬分段采样，段内 `Interval` 即该赤纬对应的可用时角区间。

### 赤道、水平与垂直日晷特例

```go
h := sundial.HorizontalDial(31.2304, 10)
n := sundial.EquatorialNorthDial(31.2304, 10)
s := sundial.EquatorialSouthDial(31.2304, 10)
v := sundial.VerticalDial(31.2304, 180, 10)
fmt.Println(h.PlaneNormalZenithDistance, n.PlaneNormalAzimuth, s.PlaneNormalAzimuth)
fmt.Println(v.PlaneNormalAzimuth, v.PlaneNormalZenithDistance)
```

四个构造器给出的盘面法线口径不同，画图时按 `PlanarDial` 字段判断即可：

| 构造器 | `PlaneNormalAzimuth` | `PlaneNormalZenithDistance` |
| --- | --- | --- |
| `HorizontalDial` | `180°` | `0°`（法线指向天顶；x 轴向东、y 轴向北） |
| `EquatorialNorthDial` | `0°` | `90° - 纬度` |
| `EquatorialSouthDial` | `180°` | `90° + 纬度` |
| `VerticalDial` | 入参归一化到 `[0°, 360°)` | `90°` |

北赤道日晷在北半球用于春夏半年（太阳赤纬为正），南赤道日晷用于秋冬半年；垂直日晷的 `planeNormalAzimuth` 按正北 `0°`、向东增加，朝南墙面取 `180°`、朝东墙面取 `90°`。

### 返回结构

| 类型 | 字段 | 含义 |
| --- | --- | --- |
| `PlanarShadowPoint` | `X` / `Y` | 落影点坐标，与 `StylusLength` 同单位 |
| | `DenominatorQ` | 投影分母，趋近 0 表示影子趋于无穷远 |
| | `SunAboveHorizon` / `PlaneIlluminated` / `Illuminated` | 太阳在地平线上、盘面受光、两者同时成立的最终判据 |
| `PlanarGeometry` | `CenterX` / `CenterY` | 日晷中心（极轴晷针固定点）坐标 |
| | `PolarStylusLength` / `PolarStylusPlaneAngle` | 极轴晷针长度、它与盘面的夹角 |
| | `HasFiniteCenter` | 为 false 时中心退化到无穷远，相关量为 `NaN` |
| `HourAngleInterval` | `Start` / `End` | 有符号时角区间（度），保证 `Start <= End` |
| `TimeLineSample` | `Date` / `Declination` / `HourAngle` / `Point` | 时刻、太阳赤纬、真太阳时角与对应落影点 |
| `DeclinationCurveSegment` | `Declination` / `Interval` / `Samples` | 该段的赤纬、可用时角区间与采样点列 |

### 完整示例

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/sundial"
)

func main() {
	date := time.Date(2026, 6, 21, 9, 30, 0, 0, time.FixedZone("CST", 8*3600))
	lon, lat := 121.4737, 31.2304

	trueSolar := sundial.TrueSolarTime(date, lon)
	hourAngle := sundial.HourAngle(date, lon)
	lineAngle := sundial.HorizontalHourLineAngle(lat, -45)
	lineAngleNow := sundial.HorizontalHourLineAngleAt(date, lon, lat)

	fmt.Println(trueSolar)
	fmt.Printf("hour angle=%.6f line@9am=%.6f line@now=%.6f\n", hourAngle, lineAngle, lineAngleNow)
}
```

输出结果：

```text
2026-06-21 09:34:10.404156224 +0805 LTZ
hour angle=-36.456649 line@9am=-27.405871 line@now=-20.959281
```

第一行的时区是合成的当地真太阳时区（经度 `121.4737°` 对应 `+08:05` 的 `LTZ`），所以打印值本身就体现了地方真太阳时与钟表时间的差。

### 综合示例：一天的可用时角与等时线

```go
dial := sundial.HorizontalDial(31.2304, 10)
for _, seg := range dial.DeclinationCurveAt(date, 1.0) {
	fmt.Printf("decl=%.2f usable=%.2f..%.2f samples=%d\n", seg.Declination, seg.Interval.Start, seg.Interval.End, len(seg.Samples))
}
mean := sundial.MeanSolarTime(date, 121.4737)
samples := dial.MeanSolarTimeLine([]time.Time{mean, mean.Add(30 * time.Minute)}, 9.5)
fmt.Println(len(samples), samples[0].HourAngle)
```

实测（`date = 2026-06-21 09:30 CST`、`121.4737°E, 31.2304°N`、晷针长 10）：

```text
decl=23.44 usable=-105.24..105.24 samples=211
2 -37.92998436772365
```

第一行说明当天在纬度 `31.2304°` 的水平日晷上，太阳赤纬 `23.44°` 时盘面能用的时角区间是 `[-105.24°, +105.24°]`（约 14 小时），采样 211 点；第二行是"地方平太阳时 9:30 这条时线"的两个采样点与其时角。可用区间可直接作为时线的绘制裁剪范围。

### 常见坑

- 把 `ZoneTimePoint` 的 `date` 当成目标地点的地方真太阳时 → 它只用年月日与时区，钟面时间来自 `zoneTimeHours`。
- 用 `PlaneIlluminated` 判断"这张日晷一天里能不能用" → 应该看 `Illuminated`，它才叠加了太阳在地平线以上这个条件。
- 垂直日晷传"墙面朝向" → `planeNormalAzimuth` 是**法线**方位角，朝南墙是 `180°`。
- 时角符号写反 → `HourAngle` 上午为负、下午为正，`HorizontalHourLineAngle` 也沿用同一符号。
- 以为时角区间一定是一段 → 跨零点的一天会被拆成多段，必须遍历返回的切片。

## 常用场景

### 真太阳时与钟表时间的差

```go
trueSolar := sundial.TrueSolarTime(date, lon)
mean := sundial.MeanSolarTime(date, lon)
fmt.Println(trueSolar, mean)
fmt.Println(trueSolar.Sub(mean)) // 均时差
```

```text
2026-06-21 09:34:10.438158222 +0805 LTZ 2026-06-21 09:35:53.532790863 +0805 LTZ
-1m43.094632641s
```

- `TrueSolarTime` 返回的是带合成时区的时刻，直接与自己钟表时间相减就得到"真太阳时快/慢多少"。
- 要把**钟面读数**换成时角（例如按区时排影子刻度），用 `MeanSolarHourAngle` / `ZoneTimeHourAngle`，不要手工加减均时差。

### 水平日晷的时线角与落影

```go
dial := sundial.HorizontalDial(lat, 10)
fmt.Println(sundial.HorizontalHourLineAngle(lat, -45))        // 时角 -45° 的时线角
fmt.Println(sundial.HorizontalHourLineAngleAt(date, lon, lat)) // 当前时刻的时线角
p := dial.ShadowPointAt(date, lon)
fmt.Println(p.X, p.Y, p.Illuminated)
```

```text
-27.40587112370779
-20.95918157094186
-6.511723246549 0.507600045956 true
```

- 水平日晷的坐标约定是 `x` 轴向东、`y` 轴向北；`Illuminated` 是"太阳在地平线上且盘面朝向太阳"的最终判据，只画有效影子时看它。
- 只有时角、没有具体时刻时用 `ShadowPointByHourAngleDeclination`，同一几何给 `-6.511402572556 0.506762879485 true`。

### 盘面受光区间与等时线

```go
dial := sundial.HorizontalDial(31.2304, 10)
for _, seg := range dial.DeclinationCurveAt(date, 1.0) {
	fmt.Printf("decl=%.2f usable=%.2f..%.2f samples=%d\n", seg.Declination, seg.Interval.Start, seg.Interval.End, len(seg.Samples))
}
mean := sundial.MeanSolarTime(date, 121.4737)
samples := dial.MeanSolarTimeLine([]time.Time{mean, mean.Add(30 * time.Minute)}, 9.5)
fmt.Println(len(samples), samples[0].HourAngle)
```

```text
decl=23.44 usable=-105.24..105.24 samples=211
2 -37.92998436772365
```

- `PlaneIlluminatedHourAngleIntervals` 只判断"盘面是否朝向太阳"，`IlluminatedHourAngleIntervals` 再叠加太阳在地平线以上，回答"这盘一天能用多久"要用后者。
- 区间约定是 `[-180, 180]` 且 `Start <= End`，跨零点会拆成多段；赤纬曲线与时间线的完整接口见 [API 参考](#时间线与赤纬曲线)。

### 退化情形与各面日晷口径

```go
h := sundial.HorizontalDial(31.2304, 10)
n := sundial.EquatorialNorthDial(31.2304, 10)
s := sundial.EquatorialSouthDial(31.2304, 10)
v := sundial.VerticalDial(31.2304, 180, 10)
fmt.Println(h.PlaneNormalZenithDistance, n.PlaneNormalAzimuth, s.PlaneNormalAzimuth)
fmt.Println(v.PlaneNormalAzimuth, v.PlaneNormalZenithDistance)
```

```text
0 0 180
180 90
```

- **退化条件**：盘面法线与极轴垂直（等价于极轴晷针与盘面平行）时，`Geometry().HasFiniteCenter` 为 `false`，`CenterX`/`CenterY`/`PolarStylusLength` 为 `NaN`、`PolarStylusPlaneAngle` 为 0——纬度 `45°`、法线方位 `180°`、法线天顶距 `45°` 正落在这个点上，此时以中心为基准的绘制不可用。
- **各面口径**：水平盘法线指向天顶（天顶距 `0°`）；垂直盘法线天顶距 `90°`，其 `planeNormalAzimuth` 是**法线**方向（朝南墙 `180°`、朝东墙 `90°`）；赤道南北面分别用 `90-纬度` 与 `90+纬度`。四个构造器的完整表见[赤道、水平与垂直日晷特例](#赤道水平与垂直日晷特例)。

## 参数与返回值约定

- **单位**：时角、时线角、赤纬、纬度、法线方位角与法线天顶距都是**度**；`PlanarDial` 的 `StylusLength` 与返回点 `X`/`Y` 共用同一长度单位，可以是任意自洽单位（毫米、米或画布坐标）。
- **时标**：观测输入为民用时刻；`TrueSolarTime` / `MeanSolarTime` 的返回字段表示地方太阳时读数，不应直接作为新的观测时刻传入星历接口。换算约定见[时标手册](timescale.md)。
- **时角符号**：`HourAngle` 上午为负、下午为正；`HourAngleInterval` 的取值范围是 `[-180, 180]` 且保证 `Start <= End`。跨零点的一天会被拆成多段。
- **`date` 的时区语义（易错）**：`MeanSolarTimePoint` / `MeanSolarTimeLine` 的 `date` 表示**目标地点的地方平太阳时**（通常是 `MeanSolarTime(...)` 的返回值）；`ZoneTimePoint` / `ZoneTimeLine` 会忽略传入 `date` 的时分秒，只用它的年月日与时区，再用参数 `zoneTimeHours` 替换钟面时间。传错会把整条时线平移。
- **落影有效性的三个布尔量**：`SunAboveHorizon` 表示太阳在地平线以上，`PlaneIlluminated` 表示盘面朝向太阳，`Illuminated` 是两者同时成立后的最终判据；只画影子时必须看 `Illuminated`。
- **退化情形**：盘面法线与极轴垂直时 `HasFiniteCenter` 为 `false`、`CenterX`/`CenterY`/`PolarStylusLength` 为 `NaN`，中心在无穷远，以中心为基准的绘制不可用；示例见[退化情形与各面日晷口径](#退化情形与各面日晷口径)。
- **`PlaneIlluminated` 与 `Illuminated` 的区别**：前者只做几何受光判断，后者还要求太阳在地平线以上；做"这张日晷一天里能用多久"的结论时用后者。

## 相关手册

- 真太阳时、均时差与太阳位置：[太阳与月亮](sun-moon.md)
- 时角、恒星时与地平转换：[坐标工具](coord.md)
- 出图时标：[时标声明](map-geojson.md#时标声明)
