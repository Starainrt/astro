# 天象地图、GeoJSON 与 KML

[English](en/map-geojson.md) | [返回 README](../../README.md)

`geojson` 与 `kml` 包用于将日食、月食、水星/金星凌日、月掩及人造卫星的地理几何导出为 RFC 7946 GeoJSON 与 KML 2.2，供 Leaflet、Mapbox、Cesium、QGIS 或 Google Earth 使用。本手册同时汇总全库 SVG 地图通用的投影选项与时标声明约定。

## 目录

- [基本示例](#基本示例)
- [主要功能](#主要功能)
  - [月食、凌日与月掩 GeoJSON](#月食凌日与月掩-geojson)
  - [人造卫星图层](#人造卫星图层)
  - [单时刻日食与月掩阴影足迹](#单时刻日食与月掩阴影足迹)
  - [KML 2.2 转换选项](#kml-22-转换选项)
- [地图投影](#地图投影)
- [时标声明](#时标声明)
- [API 概览](#api-概览)
- [约定与限制](#约定与限制)

## 基本示例

计算 2009-07-22 日全食的偏食足迹与中心路径，导出带 30 分钟时间标记的 GeoJSON，并转换为可在 Google Earth 中按时间轴播放的 KML 2.2：

```go
package main

import (
	"log"
	"os"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/kml"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2009, 7, 22, 0, 0, 0, 0, cst)

	partial, ok := eclipse.SolarEclipsePartialFootprints(date,
		eclipse.SolarEclipsePartialFootprintOptions{Step: 10 * time.Minute, BoundaryPoints: 180})
	if !ok {
		log.Fatal("no solar eclipse on date")
	}

	var centralPtr *eclipse.SolarEclipsePath
	if path, hasCentral := eclipse.SolarEclipseCentralPath(date,
		eclipse.SolarEclipsePathOptions{Step: time.Minute, TargetSpacingKM: 20}); hasCentral {
		centralPtr = &path
	}

	// 导出 GeoJSON（日偏食无中心路径时 centralPtr 允许为 nil）
	gj, err := geojson.MarshalSolarEclipseWithTimeMarkers(partial, centralPtr,
		geojson.TimeMarkerOptions{Step: 30 * time.Minute, Location: cst})
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("eclipse-2009.geojson", gj, 0o644); err != nil {
		log.Fatal(err)
	}

	// 转换为 KML 2.2（仅依赖标准库，按图层 role 分文件夹并应用默认样式）
	kmlData, err := kml.FromGeoJSON(gj, kml.Options{Name: "2009-07-22 日全食", Language: "zh"})
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("eclipse-2009.kml", kmlData, 0o644); err != nil {
		log.Fatal(err)
	}
}
```

如需裁剪输出要素（例如跳过逐时刻瞬时半影轮廓 `partial-footprint` 而保留总偏食区边界 `partial-band` 与中心带），可使用 `geojson.MarshalSolarEclipseWithOptions` 并指定 `SkipRoles`。

## 主要功能

### 月食、凌日与月掩 GeoJSON

全部 `Marshal*` 函数均返回 UTF-8 RFC 7946 `FeatureCollection` JSON，坐标为 WGS84 `[经度, 纬度]`（度）：

- **月食**：`MarshalLunarEclipse`、`MarshalLunarEclipseWithTimeMarkers`、`MarshalLunarEclipseWithOptions`（可跳过时间包络与仅见半影带）。
- **水星 / 金星凌日**：`MarshalMercuryTransit`、`MarshalVenusTransit`，接收 `mercury.NextTransit` / `venus.NextTransit` 结果，支持 `IsochroneStep` 输出 C1–C4 站心等时线：
  ```go
  m := mercury.NextTransit(time.Date(2026, 10, 12, 0, 0, 0, 0, cst))
  data, err := geojson.MarshalMercuryTransit(m, geojson.PlanetTransitOptions{
    BoundaryPoints: 360,
    IsochroneStep:  5 * time.Minute,
    TimeMarkers:    &geojson.TimeMarkerOptions{Step: time.Hour, Location: cst},
  })
  ```
- **月掩**：`MarshalStarOccultation`、`MarshalPlanetOccultation`（及 `WithTimeMarkers` 变体），接收 `moon.Find*OccultationPaths` 返回路径。

### 人造卫星图层

由 [`satellite`](satellite.md) 计算几何后，使用对应函数导出 GeoJSON，均可直接经 `kml.FromGeoJSON` 转换：

| GeoJSON 函数 | 输入来源 | 主要图层 `role` |
| --- | --- | --- |
| `MarshalSatelliteTransitPath` | `satellite.TransitEventPath` | `transit-band`、`center-line`、`left-limit`、`right-limit`、`target-horizon` |
| `MarshalSatelliteTransitShadow` | `satellite.TransitShadowAt` | `current-transit-footprint`、`current-transit-point`、`subsatellite-point` |
| `MarshalSatelliteGroundTrack` | `satellite.GroundTrack` | `ground-track`（按 `sunlit` / `penumbra` / `umbra` 分段） |
| `MarshalSatelliteCoverage` | `satellite.CoverageAt` | `subsatellite-point`、`satellite-footprint`、`night-region`、`satellite-visible` |
| `MarshalSatelliteCoverageSweep` | `satellite.CoverageBetween` | `coverage-sweep`、`visible-sweep`、`ground-track`、`current-satellite-footprint` |

- `SatellitePathOptions` 可附加 `time-marker` 或观测站 `observer`。
- 跨反经线与带内孔区域自动拆分为标准 GeoJSON Polygon（带内环）；转换为 KML 后以 `innerBoundaryIs` 保留孔洞。
- KML 默认配色为几何可见区蓝色、目视可见区青绿色、夜区中性灰、凌带黄色、中心线红色（与 [satellite/svg](satellite-svg.md) 的淡黄系配色区分）。

### 单时刻日食与月掩阴影足迹

交互式地图按时间轴实时查询阴影时，可复用阴影求解器而无需重算整场事件：

```go
solver := eclipse.NewSolarEclipseShadowSolver(eclipse.SolarEclipseShadowSolverOptions{})
instant, ok := solver.ShadowAt(time.Date(2009, 7, 22, 1, 40, 0, 0, time.UTC))
if ok {
	data, err := geojson.MarshalSolarEclipseShadowInstant(instant)
	_ = data
	_ = err
}
```

- 求解器同时提供 `StationStateAt`（站点瞬时食分、遮蔽率与太阳高度）与 `ShadowBetween`（按步长采样）。
- 月掩单时刻足迹对应 `moon.StarOccultationFootprintAt` / `PlanetOccultationFootprintsAt` 与 `geojson.MarshalStarOccultationFootprint` / `MarshalPlanetOccultationFootprints`。
- 相邻帧执行顶点插值前应先比对 `interp_signature` 属性；签名变化（如地平线截断或反经线分段变化）时须改取精确几何。

### KML 2.2 转换选项

`kml.FromGeoJSON(data, options)` 同时支持本库输出与第三方 GeoJSON，按 `role` 分层配置样式并挂载 `<TimeStamp>` / `<TimeSpan>`。`kml.Options` 字段包括：

- `Name` / `Language`（`"zh"` 或 `"en"`）：文档名称与图层文件夹语言。
- `Styles`：按 `role` 或 `"event/role"` 覆盖默认调色板样式。
- `SkipRoles`：整层丢弃指定 `role`（如移动端丢弃大面积 `partial-band` / `partial-footprint` 以减小体积）。
- `FillContext`：是否同时填充中性背景面图层（默认仅填充高亮中心带与掩星带）。
- `NoTimes`：不输出任何时间元素，生成静态叠加图层。
- `NoLookAt`：不输出默认 `Document/LookAt` 初始视角。

## 地图投影

日食、月食、月掩与卫星的 SVG 地图共用四档投影（仅影响 SVG 表达，不改变底层 WGS84 坐标）：

| 投影 | 日月食常量 · 月掩/卫星常量 | 适用场景 | 建议画布 |
| --- | --- | --- | --- |
| 自动（零值） | `EclipseMapProjectionAuto` · `MapProjectionAuto` | 默认；高纬事件自动切换极区投影 | 1200×800 |
| 等经纬 | `...Equirectangular` | 跨反经线的长食带或掩带 | 1200×800 或 1414×1000 |
| 北极 / 南极方位等距 | `...NorthPolar` / `...SouthPolar` | 覆盖极区的高纬事件 | 1200×800 |
| 正射球面 | `...Orthographic` | NASA 风格居中半球视角（以事件中心为视点） | 1000×1414 |

## 时标声明

SVG 与 GeoJSON 默认按 **UTC** 口径输出时刻，也可通过 `TimeScale: astro.TimeScaleUT1` 显式切换为 **UT1（世界时）**：

- **UTC（默认）**：SVG 图注写明 `图中时刻为 UTC`（非 UTC 显示时区同时注明时区）；GeoJSON 省略 `time_scale` 字段，`time` 为 UTC RFC 3339，`label` 按 `Location` 格式化。
- **UT1**：`Location` 须为 `nil` 或 `time.UTC`（否则 SVG 返回 `false` 或错误）；SVG 图注补充当日 `DUT1 = UT1−UTC` 秒数；GeoJSON 写入 `"time_scale": "UT1"`。
- **KML 转换**：KML 规范要求 `<when>` 必须为 UTC，`kml.FromGeoJSON` 会按当前 ΔT 模型将 UT1 标签换回 UTC 并在扩展属性中保留原 `time_scale`。

## API 概览

| 包 | 入口 | 用途 |
| --- | --- | --- |
| `geojson` | `MarshalSolarEclipse` / `...WithTimeMarkers` / `...WithOptions` | 日食全局 GeoJSON |
| `geojson` | `MarshalSolarEclipseShadowInstant` | 单时刻日食阴影 GeoJSON |
| `geojson` | `MarshalLunarEclipse` / `...WithTimeMarkers` / `...WithOptions` | 月食可见区与时间包络 GeoJSON |
| `geojson` | `MarshalMercuryTransit` / `MarshalVenusTransit` | 水星/金星凌日 GeoJSON |
| `geojson` | `MarshalStarOccultation` / `MarshalPlanetOccultation`（及 `WithTimeMarkers`、`Footprint`） | 月掩掩带与瞬时足迹 GeoJSON |
| `geojson` | `MarshalSatelliteTransitPath` / `TransitShadow` / `GroundTrack` / `Coverage` / `CoverageSweep` | 人造卫星五类图层 GeoJSON |
| `kml` | `FromGeoJSON(data, options)` | GeoJSON → KML 2.2 |

## 约定与限制

- **反经线接缝**：跨 ±180° 经线的线与面自动拆分并在接缝处补交点；闭合多边形的 ±180° 接缝段仅用于面填色闭合，绘制物理边界时应使用对应 `*-boundary` 线图层（如 `central-shadow-boundary`）或跳过接缝段。
- **时间标记上限**：`TimeMarkerOptions.Step` 零值默认 30 分钟（正值最小 1 分钟），单次导出上限 1440 个标记。
- **先筛选再算几何**：批量处理多年日食时，宜先用 `eclipse.SolarEclipseCandidates` 筛选事件，再针对目标事件生成路径与 GeoJSON。
- **ΔT 与地面位移**：`SolarEclipseShadowSolverOptions.DeltaTSeconds > 0` 可覆盖单求解器 ΔT；地面经向位移量级约为 `0.4651 × |ΔΔT| × cos(纬度)` km（对应 `basic.DeltaTGroundShiftKM`）。
- **图层来源属性**：SVG 可降级图层带 `data-source` 属性标明实际几何分支（如 `partial-band-union`、`magnitude-contours`、`greatest-time-isochrones`、`topocentric-transit-visibility` 等）。
