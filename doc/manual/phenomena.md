# 通用天象与偕日可见

[English](en/phenomena.md) | [返回 README](../../README.md)

`phenomena` 包按地心视黄经、视黄纬搜索任意两天体的事件，计算行星与恒星的晨昏始见、末见，另外提供任意高度升落和近日点、远日点、过交点。包级入口使用进程当前的星历模型，也可以用 `ctx.Phenomena()` 固定星历实例；只有恒星的任意高度升落按传入的视坐标计算，不读星历。

## 目录

- [快速示例](#快速示例)
- [天体对事件](#天体对事件)
- [恒星坐标](#恒星坐标)
- [偕日可见](#偕日可见)
- [任意高度升落](#任意高度升落)
- [近日点、远日点与过交点](#近日点远日点与过交点)
- [API 参考](#api-参考)
- [参数与返回值约定](#参数与返回值约定)

## 快速示例

```go
start := time.Date(2020, 10, 1, 0, 0, 0, 0, time.UTC)
end := time.Date(2021, 3, 1, 0, 0, 0, 0, time.UTC)

events := phenomena.ClosestApproaches(phenomena.Jupiter, phenomena.Saturn, 1, start, end)
fmt.Println(events[0].Time.UTC().Format("2006-01-02 15:04"), fmt.Sprintf("%.4f", events[0].Value))
```

```text
2020-12-21 18:21 0.1018
```

`ClosestApproaches` 返回窗口内两天体角距的**局部极小**，`Value` 是当时的角距（度）。窗口内没有小于 `maxSeparation` 的极小值时返回空切片。

固定星历实例时换成模块方法：

```go
ctx, err := astro.NewContext(astro.ModelLegacy)
if err != nil {
	panic(err)
}
events = ctx.Phenomena().ClosestApproaches(phenomena.Jupiter, phenomena.Saturn, 1, start, end)
```

## 天体对事件

| 入口 | 事件 | `Event.Value` |
| --- | --- | --- |
| `LongitudeCrossings(body, lon, start, end)` | 视黄经经过 `lon` | 事件时刻的视黄经（度） |
| `Stations(body, start, end)` | 视黄经的留（行星） | 留时的视黄经（度） |
| `AspectCrossings(a, b, angle, start, end)` | `lon(a) − lon(b) = angle`；`0` 为黄经合，`180` 为冲 | 实际的黄经差（度） |
| `ClosestApproaches(a, b, max, start, end)` | 两天体角距的局部极小 | 角距（度） |
| `StarApproaches(body, star, max, start, end)` | 天体与恒星角距的局部极小 | 角距（度） |

各入口只接受上表列出的参数，结果按时间升序返回。各行星包里按赤经定义的留与这里按黄经定义的留口径不同，时刻可差数小时。

```go
// 2026 年前两次水星留
stations := phenomena.Stations(phenomena.Mercury,
	time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC))
for _, e := range stations {
	fmt.Println(e.Time.UTC().Format("01-02 15:04"), fmt.Sprintf("%.3f", e.Value))
}
```

```text
02-26 06:48 352.565
03-20 19:32 338.491
```

```go
// 太阳视黄经经过 0°（春分）
crossings := phenomena.LongitudeCrossings(phenomena.Sun, 0,
	time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC), time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC))
fmt.Println(crossings[0].Time.UTC().Format("2006-01-02 15:04"))
```

```text
2026-03-20 14:45
```

## 恒星坐标

行星与恒星的角距、恒星偕日可见都要传入 `StarCoordinate`，内嵌 9100 星表可以直接查：

| 入口 | 输入 |
| --- | --- |
| `StarCoordinateByName(name)` | 星表中文名/别名或英文常用名，例如 `"天狼星"`、`"sirius"`；同名多星返回 `ErrStarNameAmbiguous` |
| `StarCoordinateByHR(hr)` | HR 编号 |
| `StarCoordinateFromStarData(star)` | `star.StarDataByHR` 等取到的星表记录 |

```go
star, err := phenomena.StarCoordinateByName("天狼星")
fmt.Println(err, star.ID, fmt.Sprintf("%.6f %.6f %.2f %.2f", star.RA, star.Dec,
	star.ProperMotionRACosDecMasPerYear, star.ProperMotionDecMasPerYear))
```

```text
<nil> 天狼 101.287083 -16.716111 -553.00 -1205.00
```

`StarCoordinateByName` 遇到同名多星会返回 `phenomena.ErrStarNameAmbiguous`，未找到返回 `phenomena.ErrStarNotFound`；两者都可用 `errors.Is` 判断。重名时先用 `star.StarDataByNameIn(name, region)` 取到唯一星表记录，再交给 `StarCoordinateFromStarData`：

```go
data, err := star.StarDataByNameIn("杵增一", "危宿")
if err != nil {
	panic(err)
}
coord, err := phenomena.StarCoordinateFromStarData(data)
if err != nil {
	panic(err)
}
```

`RA`、`Dec` 为度，`Epoch` 与 `Frame` 必填；自行字段（`ProperMotionRACosDecMasPerYear` 等）按星表常见的 `dRA·cos(Dec)` 约定，单位毫角秒/年。`ParallaxMas`（或等价的 `DistanceLightYear`）给出距离，`RadialVelocityKmPerSecond` 给出视向速度；距离缺失时按二维自行传播，有距离时才计入视向速度。

## 偕日可见

`PlanetHeliacalEvents(body, start, end, lon, lat, options)` 与 `StarHeliacalEvents(star, magnitude, start, end, lon, lat, options)` 逐日扫描太阳高度 −18° 到 0° 的晨昏时段，按天光亮度模型求肉眼极限星等，再与目标视星等比较（行星用 Mallama & Hilton 2018 星等），返回四类事件：

| 常量 | 含义 |
| --- | --- |
| `HeliacalMorningFirst` | 晨始见（偕日升） |
| `HeliacalMorningLast` | 晨末见 |
| `HeliacalEveningFirst` | 昏始见 |
| `HeliacalEveningLast` | 昏末见（偕日落） |

`HeliacalEvent` 中 `Time` 是当次晨昏的最佳可见时刻，`TargetAltitude`、`SunAltitude` 是当时目标与太阳的几何高度（度），`Margin` 是极限星等减目标星等。

| `HeliacalOptions` | 默认 | 说明 |
| --- | --- | --- |
| `ExtinctionCoefficient` | `0.25` | 大气消光，星等/大气质量 |
| `MinTargetAltitude` | `0.5` | 目标最低几何高度（度） |
| `Moonlight` | `false` | 计入月光散射（Krisciunas–Schaefer） |

```go
events := phenomena.PlanetHeliacalEvents(phenomena.Venus,
	time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
	31.23, 30.04, phenomena.HeliacalOptions{})
for _, e := range events {
	fmt.Println(e.Kind, e.Time.UTC().Format("2006-01-02"), fmt.Sprintf("margin=%.2f", e.Margin))
}
```

```text
3 2025-03-18 margin=0.10
0 2025-03-25 margin=0.58
1 2025-12-01 margin=0.10
```

`Kind` 打印出来是常量序号，判断时用常量本身比较。`LimitingMagnitude(targetAltitude, sunAltitude, separation, extinction)` 单独给出某一时刻无月夜的极限星等，与事件判定同源：

```go
fmt.Println(fmt.Sprintf("%.3f", phenomena.LimitingMagnitude(20, -12, 60, 0.25)))
```

```text
3.408
```

结果对消光系数敏感，可见日期差一两天属正常。

## 任意高度升落

`AltitudeCrossing(body, date, lon, lat, height, altitude, options)` 求日、月或行星在 `date` 所在**当地民用日**内首次穿过视高度 `altitude` 的升、落时刻。`StarAltitudeCrossing(date, ra, dec, lon, lat, altitude, options)` 对恒星视坐标做同样计算，它不读星历，也没有模块方法。

| `AltitudeOptions` | 默认 | 说明 |
| --- | --- | --- |
| `Limb` | `LimbUpper` | 日月取上缘，也可取 `LimbCenter`、`LimbLower`；行星与恒星忽略 |
| `PressureHPa` / `TemperatureC` | 标准大气（`0` 值） | 指定实测气象 |
| `NoRefraction` | `false` | 为 `true` 时按几何高度比较 |

| `AltitudeCrossingResult` | 说明 |
| --- | --- |
| `Rise`、`Set` | 升、落时刻，没有事件时为零值 |
| `HasRise`、`HasSet` | 当天是否存在升、落 |
| `State` | `AltitudeCrosses`、`AltitudeAlwaysAbove`、`AltitudeAlwaysBelow` 或 `AltitudeTangent`（极值与目标高度相差不足 `1″`） |
| `MinResidual`、`MaxResidual` | 当天高度与目标高度之差的最小、最大值（度） |

```go
cst := time.FixedZone("CST", 8*3600)
r, err := phenomena.AltitudeCrossing(phenomena.Sun, time.Date(2026, 3, 21, 12, 0, 0, 0, cst),
	116.4, 39.9, 0, -18, phenomena.AltitudeOptions{})
fmt.Println(err, r.State, r.Rise.In(cst).Format("15:04:05"), r.Set.In(cst).Format("15:04:05"),
	fmt.Sprintf("%.3f %.3f", r.MinResidual, r.MaxResidual))
```

```text
<nil> 0 04:44:54 19:59:16 -31.806 68.602
```

```go
// 北纬 89° 极昼：太阳整天都在 -18° 之上
polar, err := phenomena.AltitudeCrossing(phenomena.Sun, time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC),
	0, 89, 0, -18, phenomena.AltitudeOptions{})
fmt.Println(err, polar.State, polar.HasRise, polar.HasSet, fmt.Sprintf("%.3f", polar.MinResidual))
```

```text
<nil> 1 false false 40.737
```

```go
// 天狼星穿过 -0.833° 的升落
sirius, err := phenomena.StarAltitudeCrossing(time.Date(2026, 1, 1, 12, 0, 0, 0, cst),
	star.RA, star.Dec, 116.4, 39.9, -0.833, phenomena.AltitudeOptions{Limb: phenomena.LimbCenter})
fmt.Println(err, sirius.State, sirius.Rise.In(cst).Format("15:04:05"), sirius.Set.In(cst).Format("15:04:05"))
```

```text
<nil> 0 19:04:31 05:27:52
```

`height` 只影响月球视差。需要以下沉地平为基准时，由调用方从 `altitude` 中减去地平俯角。天体不是日、月或行星（例如 `phenomena.Earth`），或恒星视坐标不是有限值时，返回 `basic.ErrInvalidObservationInput`。

## 近日点、远日点与过交点

| 常量 | 事件 |
| --- | --- |
| `Perihelion`、`Aphelion` | 近日点、远日点 |
| `AscendingNode`、`DescendingNode` | 穿过历元平黄道的升交点、降交点 |

`Apsides(body, start, end)` 接受行星或地球（`phenomena.Earth`），`Nodes(body, start, end)` 只接受行星，传地球返回空切片。`OrbitalEvent` 的 `Time` 是事件时刻，`Kind` 是上表常量，`DistanceAU` 是该时刻的日心距（AU）。

```go
apsides := phenomena.Apsides(phenomena.Earth,
	time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC))
for _, e := range apsides {
	fmt.Println(e.Kind, e.Time.UTC().Format("2006-01-02 15:04"), fmt.Sprintf("%.9f AU", e.DistanceAU))
}
nodes := phenomena.Nodes(phenomena.Jupiter,
	time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC))
for _, e := range nodes {
	fmt.Println(e.Kind, e.Time.UTC().Format("2006-01-02 15:04"), fmt.Sprintf("%.6f AU", e.DistanceAU))
}
```

```text
0 2026-01-03 09:15 0.983302049 AU
1 2026-07-06 09:30 1.016643962 AU
3 2020-02-25 18:09 5.207056 AU
2 2025-09-19 11:58 5.174312 AU
```

极值取日心距本身：太阳绕质心的反射运动会让天王星、海王星的日心距在近日点附近出现多个小极值，同一段出现多个极小值时只返回前后各 `0.3` 个轨道周期内日心距最小的一个，与吻切轨道近日点可相差数年。

## API 参考

`ctx.Phenomena()` 的方法与同名包级函数签名一致（`StarAltitudeCrossing` 除外）；`NewWithInstance` 用给定星历实例创建模块。各入口的参数与单位见上文各节。

## 参数与返回值约定

- 角度一律为度，`OrbitalEvent.DistanceAU` 为 AU，`StarHeliacalEvents` 的 `magnitude` 为 V 星等。
- 时间参数按绝对时刻解释（读取其 UTC 字段），返回时刻换算到 `start`（升落类为 `date`）所在的时区。
- 事件搜索没有结果时返回空切片，不返回错误；只有任意高度升落在输入非法时返回 `basic.ErrInvalidObservationInput`。
- 事件搜索会用到星历，可以用模块方法固定模型；`StarAltitudeCrossing`、`StarCoordinate*` 与 `LimitingMagnitude` 与星历无关。
- 行星视星等、赤经留等口径差异见[行星手册](planets.md)；日月升落与观测角语义见[太阳与月亮](sun-moon.md)。
