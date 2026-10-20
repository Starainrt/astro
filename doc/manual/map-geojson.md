# 天象地图、GeoJSON 与 KML

[English](en/map-geojson.md) | [返回 README](../../README.md)

> 本手册的完整示例以仓库根目录为工作目录执行，生成的图片写入 `doc/img/`。

## 目录

- [导出日食 GeoJSON](#导出日食-geojson)
- [导出凌日 GeoJSON](#导出凌日-geojson)
- [API 参考](#api-参考)
- [常用场景](#常用场景)
  - [给事件选投影](#给事件选投影)
  - [导出 GeoJSON 与时间标记](#导出-geojson-与时间标记)
  - [时标声明与 UT1](#时标声明与-ut1)
  - [单时刻足迹与图层词表](#单时刻足迹与图层词表)
- [地图投影](#地图投影)
- [时标声明](#时标声明)
- [GeoJSON](#geojson)
  - [图层筛选](#图层筛选)
  - [坐标、时间与反经线](#坐标时间与反经线)
  - [按时刻计算日食阴影](#按时刻计算日食阴影)
  - [地平线闭合与插值](#地平线闭合与插值)
  - [ΔT 与地面位置](#δt-与地面位置)
  - [月掩瞬时足迹](#月掩瞬时足迹)
  - [先查事件再计算几何](#先查事件再计算几何)
- [KML](#kml)
  - [转换文件](#转换文件)
  - [Options](#options)
  - [图层与样式](#图层与样式)
  - [时间轴与静态叠加](#时间轴与静态叠加)
  - [控制文件大小和取景](#控制文件大小和取景)
  - [属性与输入校验](#属性与输入校验)

## 导出日食 GeoJSON

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2009, 7, 22, 0, 0, 0, 0, cst)
	partial, ok := eclipse.SolarEclipsePartialFootprints(date,
		eclipse.SolarEclipsePartialFootprintOptions{Step: 10 * time.Minute, BoundaryPoints: 180})
	if !ok {
		log.Fatal("no solar eclipse")
	}
	path, ok := eclipse.SolarEclipseCentralPath(date,
		eclipse.SolarEclipsePathOptions{Step: time.Minute, TargetSpacingKM: 20})
	var centralPath *eclipse.SolarEclipsePath
	if ok {
		centralPath = &path
	}
	data, err := geojson.MarshalSolarEclipseWithTimeMarkers(partial, centralPath,
		geojson.TimeMarkerOptions{Step: 30 * time.Minute, Location: cst})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))
}
```

偏食可以没有中心路径，因此 `centralPath` 允许为 `nil`。本例把 GeoJSON 写到标准输出，可重定向到文件；SVG 与 KML 用法见后文。

## 导出凌日 GeoJSON

凌日入口接收全球事件的 `TransitInfo`，接触字段仍为地心参考时刻，分区口径与凌日全球图一致：

地心无接触的边缘事件（如 2854-12-14 金星凌日）导出站心实际可见的 `transit-partial` 区域，
不生成不存在的地心 C1/C4 地平线。属性还包含 `has_geocentric_contacts`、`gamma`、`penumbral_radius`、
`series`、`series_period_years`。足迹按 2° 经度、1° 纬度采样并细化边界，这是制图采样而不是接触求解精度，比采样间距更窄的可见区可能未检出；精确判断请用 `LocalTransit`。

```go
m := mercury.NextTransit(time.Date(2026, 10, 12, 0, 0, 0, 0, cst))
data, err := geojson.MarshalMercuryTransit(m, geojson.PlanetTransitOptions{
	BoundaryPoints: 360,
	IsochroneStep:  5 * time.Minute,
	TimeMarkers:    &geojson.TimeMarkerOptions{Step: time.Hour, Location: cst},
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(string(data))
```

金星把 `mercury.NextTransit` 换成 `venus.NextTransit`、`MarshalMercuryTransit` 换成 `MarshalVenusTransit`，选项相同。`date` 早于该次凌日即可：`NextTransit` 返回其后最近的一次，取前后最近一次用 `ClosestTransit`。`TransitInfo.Valid` 为假时导出返回错误。

## API 参考

SVG 片段使用导入别名 `eclipsesvg "github.com/starainrt/astro/eclipse/svg"`；月掩图使用 `moonsvg "github.com/starainrt/astro/moon/svg"`。日期和时区沿用首例。

| 名称 | 用途 | 备注 |
| --- | --- | --- |
| `eclipsesvg.EclipseMapProjectionAuto` / `...Equirectangular` / `...NorthPolar` / `...SouthPolar` / `...Orthographic` | 日月食地图投影 | 零值即自动 |
| `moonsvg.MapProjectionAuto` / `...Equirectangular` / `...NorthPolar` / `...SouthPolar` / `...Orthographic` | 月掩地图投影 | 同上 |
| `SolarEclipseMapSVG` / `LunarEclipseMapSVG` | 日食、月食全球图 | 返回 `(string, bool)` |
| `StarOccultationPathSVG` / `PlanetOccultationPathSVG` | 月掩全球掩带图 | 返回 `(string, error)` |
| `MarshalSolarEclipse` / `MarshalSolarEclipseWithTimeMarkers` | 日食 GeoJSON | 无/带时间标记 |
| `MarshalLunarEclipse` / `MarshalLunarEclipseWithTimeMarkers` / `MarshalLunarEclipseWithOptions` | 月食 GeoJSON | 同上 |
| `MarshalMercuryTransit` / `MarshalVenusTransit` | 凌日 GeoJSON | 选项见 `PlanetTransitOptions` |
| `MarshalStarOccultation` / `MarshalPlanetOccultation`（及 `...WithTimeMarkers`） | 月掩 GeoJSON | 同上 |
| `NewSolarEclipseShadowSolver` / `MarshalSolarEclipseShadowInstant` | 单时刻足迹及其 GeoJSON | 拖动时间轴用 |
| `TimeMarkerOptions` | 时间标记 `Step`、`Location`、`TimeScale` | `Step` 零值 30 分钟，最多 1440 个 |
| `astro.TimeScaleUT1` / `astro.DUT1` | UT1 口径与 DUT1 差值 | UT1 时 `Location` 必须为 UTC |
| `kml.FromGeoJSON` | 把 GeoJSON 转成 KML 2.2 | 只依赖标准库；按 `role` 分层并给默认调色板 |

## 常用场景

### 给事件选投影

```go
for _, spec := range []struct {
	name string
	p    eclipsesvg.EclipseMapProjection
}{
	{"equirectangular", eclipsesvg.EclipseMapProjectionEquirectangular},
	{"north-polar", eclipsesvg.EclipseMapProjectionNorthPolar},
	{"south-polar", eclipsesvg.EclipseMapProjectionSouthPolar},
	{"orthographic", eclipsesvg.EclipseMapProjectionOrthographic},
} {
	svg, ok := eclipsesvg.SolarEclipseMapSVG(date, eclipsesvg.SolarEclipseMapSVGOptions{
		Width: 1200, Height: 800, Location: cst, Projection: spec.p,
	})
	fmt.Println(spec.name, ok, len(svg))
}
```

```text
equirectangular true 265827
north-polar true 181602
south-polar true 91902
orthographic true 537001
```

投影只影响 SVG 表达、不改变底层 WGS84 地理结果；`Auto`（零值）按事件自动选极图，需要固定版式时才显式指定。正射球面视点取食甚点、只画朝向视点的半球，并切换成 NASA 摆法版式——代价与底图反解过程见[地图投影](#地图投影)。

### 导出 GeoJSON 与时间标记

```go
data, err := geojson.MarshalSolarEclipseWithTimeMarkers(partial, centralPath,
	geojson.TimeMarkerOptions{Step: 30 * time.Minute, Location: cst})
fmt.Println(err, json.Valid(data), len(data))
```

```text
<nil> true 426253
```

- `WithTimeMarkers` 会追加 `role=time-marker` 的 Point 要素：`label` 按 `Location` 本地化显示，`time` 始终是 UTC RFC 3339。
- `Step` 零值为 30 分钟，正值至少 1 分钟，单次导出最多 1440 个标记；不需要标记就用不带后缀的 `MarshalSolarEclipse`。
- 月食与月掩有对称入口（`MarshalLunarEclipse*`、`MarshalStarOccultation*`、`MarshalPlanetOccultation*`）。

### 时标声明与 UT1

```go
fmt.Println(astro.DUT1(date)) // UT1−UTC，单位秒
```

```text
0.23
```

出图默认按 UTC 口径并在图上或图注声明；`TimeScale: astro.TimeScaleUT1` 改成 UT1 读数并在图里写出 `DUT1 = UT1−UTC = +0.23 s`，此时 `Location` 必须是 UTC。GeoJSON 的对应成员是 `time_scale`（UTC 口径省略，UT1 口径写 `"UT1"`），完整约定见[时标声明](#时标声明)。

### 单时刻足迹与图层词表

```go
solver := eclipse.NewSolarEclipseShadowSolver(eclipse.SolarEclipseShadowSolverOptions{})
instant, ok := solver.ShadowAt(date)
shadow, err := geojson.MarshalSolarEclipseShadowInstant(instant)
fmt.Println(ok, err, json.Valid(shadow), len(shadow))
```

```text
true <nil> true 4217
```

- 单时刻接口只算"这一瞬间的本影/半影足迹"与站心日月几何，不产生可见带、食分线、升落边界或南北界；本影不在地球上时返回空值而不是错误。
- 插值前先比 `interp_signature`：只有相同的相邻时刻才适合按顶点插值，`closed` 翻转、段数/顶点数变化、空↔非空（U1/U4 附近）都必须改取精确几何。
- 可降级图层用 `data-source` 标注实际几何来源，没有降级分支的图层不带该属性；合法取值见下面的图层词表。

可降级图层的 `data-source` 取值按出图族分组，完整含义见 `eclipse/svg` 包注释：

- 日食边界与可见域：`partial-band-union`、`sampled-footprint-sweep`、`partial-band-contours`、`besselian-critical-envelope`、`paired-limit-chords`、`sampled-open-sweep`、`central-path-limits`
- 日食时刻与食分线：`rise-set-phase-lines`、`magnitude-contours`、`greatest-time-isochrones`
- 瞬时轮廓：`penumbral-outlines`、`central-shadow-outlines`
- 月食：`p1-p4-visibility-regions`、`p1-p4-horizon-boundaries`
- 凌日：`c1-c4-visibility-regions`、`c1-c4-horizon-boundaries`、`transit-contact-isochrones`；地心无接触时可见区填充改用 `topocentric-transit-visibility`

## 地图投影

日食、月食与月掩的 SVG 共用同一套投影：等经纬、北极方位等距、南极方位等距与正射球面。投影只影响 SVG 表达，不改变底层 WGS84 地理结果；日食与月掩的自动投影（零值 `...Auto`）会在适合时选择极图，月食默认使用等经纬投影。

| 投影 | 常量（日食/月食 · 月掩） | 适用场景 | 画布建议 |
| --- | --- | --- | --- |
| 自动 | `EclipseMapProjectionAuto` · `MapProjectionAuto` | 默认；按事件自动选极图 | 1200×800 |
| 等经纬 | `...Equirectangular` | 跨反经线的长食带或掩带 | 1200×800、1414×1000 |
| 北极/南极方位等距 | `...NorthPolar` / `...SouthPolar` | 事件整体落在高纬 | 1200×800、1000×1414 |
| 正射球面 | `...Orthographic` | NASA 版式的半球图，视点取事件中心 | 1000×1414 |

`EclipseMapProjectionOrthographic` / `MapProjectionOrthographic` 以食甚点（月掩取事件中心）为视点，只显示朝向视点的半球。该投影使用 NASA 风格的居中球面版式，建议画布为 `1000×1414`。

同一事件批量出四种投影：

```go
date := time.Date(2009, 7, 22, 12, 0, 0, 0, cst)
for _, spec := range []struct {
	name string
	p    eclipsesvg.EclipseMapProjection
}{
	{"equirectangular", eclipsesvg.EclipseMapProjectionEquirectangular},
	{"north-polar", eclipsesvg.EclipseMapProjectionNorthPolar},
	{"south-polar", eclipsesvg.EclipseMapProjectionSouthPolar},
	{"orthographic", eclipsesvg.EclipseMapProjectionOrthographic},
} {
	options := eclipsesvg.SolarEclipseMapSVGOptions{
		Width: 1200, Height: 800, Location: cst, Projection: spec.p,
	}
	svg, ok := eclipsesvg.SolarEclipseMapSVG(date, options)
	if !ok {
		continue
	}
	_ = os.WriteFile("solar-eclipse-"+spec.name+".svg", []byte(svg), 0o644)
}
```

各族的选项、图层开关与带图示例见[日月食与凌日手册](eclipse.md#全球见食图与月食出图)与[月掩手册](occultation.md#月掩出图)；GeoJSON 侧只携带地理结果，投影由应用自行选择。凌日的可见分区、C1/C4 视地平线与可选站心等时线也可以直接导出为 GeoJSON，见上面的凌日导出示例。

## 时标声明

图内时刻的口径由 `TimeScale` 选项决定，日食全球图、月食全球图、月食详细版式与月掩三族图都会在图上或图注里写明：

- 默认（`astro.TimeScaleUTC`，零值）：写 `图中时刻为 UTC`；展示时区不是 UTC 时写 `图中时刻为 UTC（显示时区 CST，UTC+08:00）`，把时标与展示时区分开声明。
- `astro.TimeScaleUT1`：写 `图中时刻为 UT1（世界时），DUT1 = UT1−UTC = +0.05 s。`，差值随事件日期变化；此时 `Location` 必须是 UTC，否则渲染返回 `false`（月掩侧直接返回错误）。

GeoJSON 侧的对应成员是 `time_scale`：UTC 口径省略该成员，UT1 口径写 `"UT1"`，表示所有 `time` 字符串与 `HH:MM` 标注都是 UT1 读数（RFC 3339 的 `Z` 后缀严格说不等于 UT1）。切时标不改掩带、足迹、地平线与等值线几何，只有写出的时刻文字换成 UT1 读数；时间标记按输出时标的整点取点（整点读数本身就是另一个物理时刻），位置随所标时刻移动。

## GeoJSON

`geojson` 接收已经计算好的日食、月食或月掩结果，返回 `[]byte`。这段字节是完整的 UTF-8 RFC 7946 `FeatureCollection` JSON，可直接写入 `.geojson`，或交给 `encoding/json` 与前端地图组件。

```go
package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

func main() {
	date := time.Date(2024, 4, 8, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipse.SolarEclipsePartialFootprints(
		date,
		eclipse.SolarEclipsePartialFootprintOptions{
			Step: 10 * time.Minute, BoundaryPoints: 180,
		},
	)
	if !ok {
		return
	}
	central, hasCentral := eclipse.SolarEclipseCentralPath(
		date,
		eclipse.SolarEclipsePathOptions{Step: time.Minute, TargetSpacingKM: 20},
	)
	var centralPath *eclipse.SolarEclipsePath
	if hasCentral {
		centralPath = &central
	}

	data, err := geojson.MarshalSolarEclipseWithTimeMarkers(
		partial, centralPath,
		geojson.TimeMarkerOptions{
			Step:     30 * time.Minute,
			Location: time.FixedZone("CST", 8*3600),
		},
	)
	fmt.Println(err, json.Valid(data))
}
```

对应的无时间标记和带时间标记入口包括：

- `MarshalSolarEclipse` / `MarshalSolarEclipseWithTimeMarkers`
- `MarshalLunarEclipse` / `MarshalLunarEclipseWithTimeMarkers`
- `MarshalLunarEclipseWithOptions`：用 `LunarEclipseOptions` 一次给出时间标记、`SkipRoles` 与时间包络的采样参数。`SkipRoles` 可跳过时间包络和仅见半影带；跳过两个时间包络时不会执行包络采样。`EnvelopeSweepSamples` 默认 48，限制在 `[2, 192]`；`EnvelopeLongitudePoints` 默认 `max(360, boundaryPoints)`，限制在 `[12, 720]`。
- `MarshalStarOccultation` / `MarshalStarOccultationWithTimeMarkers`
- `MarshalPlanetOccultation` / `MarshalPlanetOccultationWithTimeMarkers`
- `MarshalMercuryTransit` / `MarshalVenusTransit`：接收 `mercury.TransitInfo` / `venus.TransitInfo`，分区口径同凌日 SVG 全球图（一触、凌甚、四触太阳中心是否在视地平上，计入标准大气折射）。`BoundaryPoints` 是地平线采样点数，同时决定分区的经度列数 `max(360, BoundaryPoints)`（上限 720）。role 有 `transit-entire`、`transit-at-sunrise`、`transit-at-sunset` 三个分区，`visible-at-c1` / `visible-at-c4` 与 `c1-horizon` / `c4-horizon`，`greatest`（太阳直射点）与时间标记 `subsolar-track`；地心无接触事件改用 `transit-partial` 分区，不生成三个可见分区、视地平与可见半球 role；`IsochroneStep` 为正（至少 1 秒）时追加 `c1-isochrone` 到 `c4-isochrone` 站心等时线（地表没有站心内切接触时无二、三触，地心只有外切的掠凌仍可能在部分地区出现；时刻是站心接触时刻，受视差影响可落在地心一触到四触之外），某个接触的等时线不落在可见半球上时该 role 不出现；`SkipRoles` 跳过全部三个分区时不做分区计算。
- `MarshalSolarEclipseWithOptions`：用 `SolarEclipseOptions` 一次给出时间标记与 `SkipRoles`，等价于上面的日食入口再加图层裁剪。`SkipRoles` 里最常用的是 `partial-footprint`（瞬时半影轮廓）；`partial-band` 是日食带边缘（食分 0 界限），通常要留着。

### 图层筛选

`SolarEclipseOptions.SkipRoles` 按 `role` 排除输出要素，零值保留全部图层。可通过 `MarshalSolarEclipseWithOptions` 同时指定时间标记与筛选条件。

筛选发生在编码阶段，几何仍按完整采样计算。因此，跳过 `partial-footprint` 会去掉逐时刻的半影足迹，但不会降低总偏食区边缘 `partial-band` 的精度，也不影响中心带、食分线和升落边界。全部要素被排除时返回错误。

### 坐标、时间与反经线

坐标为 WGS84 `[经度, 纬度]`，单位度。跨反经线的线和面会拆分，并在接缝两侧补出同一个交点。触及极点的环用 `[±180, ±90]` 两个坐标表示同一极点。

闭合环含有用于闭合的 ±180° 接缝段。填充时需要这些边；描绘真实边界时应跳过接缝。`band-outline`、`total-band-outline` 等闭合轮廓线也需按此规则处理。

带时间路径的 `times` 与各段坐标逐点对应。`WithTimeMarkers` 额外加入 `role=time-marker` 的 Point 要素：`label` 用于显示，默认 `time` 为 UTC RFC 3339；显式选择 UT1 时由 `time_scale` 区分。详见[时标](timescale.md)。

| `TimeMarkerOptions` 字段 | 行为 |
| --- | --- |
| `Step` | 零值为 30 分钟，正值至少 1 分钟；一次导出最多 1440 个标记 |
| `Location` | 标签的显示时区，`nil` 使用 UTC |
| `TimeScale` | 默认 UTC；UT1 要求 `Location` 为 `nil` 或 `time.UTC` |

GeoJSON 不携带底图、样式或投影，应用可自行选择瓦片、Web Mercator 或极区投影。

### 按时刻计算日食阴影

交互地图按时间查询阴影时，可复用 `eclipse.NewSolarEclipseShadowSolver` 返回的求解器：

| 方法 | 输入与结果 |
| --- | --- |
| `ShadowAt(date)` | 民用 `time.Time`，返回瞬时全球阴影足迹 |
| `ShadowAtJDE(jdeTT)` | TT 儒略日，返回相同类型的足迹 |
| `StationStateAt` / `StationStateAtJDE` | 指定时刻与站点的食分、遮蔽率、站心角距、视半径、太阳高度/方位与全食/环食状态 |
| `ShadowBetween(start, end, step)` | 按步长返回足迹序列，无阴影的时刻保留空条目 |
| `StationStatesBetween` | 按时间采样站点状态 |

默认计算本影或反本影；`Kind: SolarEclipseShadowPenumbra` 改为半影。单时刻接口不计算整场事件的可见带、食分线、南北限或中心线；阴影不在地球上时返回空结果。

`geojson.MarshalSolarEclipseShadowInstant(instant)` 将结果编码为 GeoJSON。输出属性包括 `time`、`source_boundary_closed`、`geometry_role`、`closure`、`delta_t_seconds`、`model` 与 `interp_signature`。没有阴影时返回空 FeatureCollection。

| 阴影 | 区域 role | 物理边界 role |
| --- | --- | --- |
| 本影/反本影 | `central-shadow-footprint` | `central-shadow-boundary` |
| 半影 | `partial-footprint` | `partial-footprint-boundary` |

单时刻半影与整场偏食采样使用相同的默认边界参数（96 点、200 km 加密），可以对照同一时刻的结果。已有单机测量中，96 点瞬时足迹约 64 µs，站点瞬时状态约 20 µs；这些数据只用于估算成本，首次查询和批量查询会受缓存状态影响。

### 地平线闭合与插值

`central-shadow-footprint` 只输出 Polygon 或 MultiPolygon。被地平线切断时，物理边界延伸到两个地平擦地点，再沿地平弧闭合为覆盖区域；此时 `source_boundary_closed=false`，`closure` 用 `kind`、`time`、`subsolar` 描述闭合弧。

只画真实阴影边缘时使用 `central-shadow-boundary`；需要填色时使用 footprint。阴影在 U1/U4 收缩为空时不输出该要素，不会退化成折线。`source_boundary_closed=true` 表示物理边界本身已经闭合。

采样的半影足迹也带这些属性。足迹只覆盖一个时刻，静态掩带覆盖整场事件，取几何时不要混用。

相邻帧可先比较 `interp_signature`，例如 `umbra-closed-seg1-pt97`。只有签名一致、分段数与顶点数对应时才适合按顶点插值；这不代表插值具有严格误差上界。闭合状态改变、反经线分段改变或空/非空切换时，应查询精确几何。

移动距离也会随事件阶段变化：已有两分钟采样的对照中，足迹质心在中段移动约 78–232 km，接触附近可达约 520 km。因此不能仅凭固定时间步长判断动画误差。

### ΔT 与地面位置

求解器的 `DeltaTSeconds > 0` 可为该句柄指定 ΔT；小于等于 0 时采用进程级模型。结果回传实际使用的值。

同一 TT 下，改变 ΔT 会改变地球自转相位，而不改变日月在空间中的相对几何。地面经向位移的近似量级为 `0.4651 × |ΔΔT| × cos(纬度)` 千米，函数 `basic.DeltaTGroundShiftKM` 可用于换算。库不提供 ΔT 不确定度模型，误差输入需由调用方给出。

### 月掩瞬时足迹

`moon.StarOccultationFootprintAt` / `moon.PlanetOccultationFootprintsAt` 的结果可传给 `geojson.MarshalStarOccultationFootprint` / `MarshalPlanetOccultationFootprints`。

它们同样输出 `delta_t_seconds`、`source_boundary_closed`、`geometry_role` 和 `interp_signature`。月球地平线闭合采用 `closure.kind=target-horizon`、`body=moon` 与 `sublunar` 月下点。月掩计算沿用进程级 ΔT，并在结果中报告实际值。

### 先查事件再计算几何

只需要事件列表时，可先用 `eclipse.SolarEclipseCandidates(start, end, options)` 获取食甚时刻、食型、中心食类型、食分、伽马和可选沙罗信息；这个结果不包含地理几何。

固定地点的中心食搜索使用 `SearchLocalCentralSolarEclipse`，选项包括 `Kind`、`MaxYears`、`Backward`、`Geometric`、`Model`，返回 `(info, SolarEclipseLocalSearchStatus)`，其 `Exhausted` 表示已用尽搜索范围。

`MaxYears<=0` 使用默认搜索预算（6000 次候选步进，约 992 年）。找到事件后再计算路径、足迹或 SVG，可以避免为不需要的事件生成地图数据。

## KML

`kml.FromGeoJSON(data []byte, options kml.Options) ([]byte, error)` 把 GeoJSON FeatureCollection 转成 KML 2.2。输入可来自本库，也可以是第三方文件；转换器读取已有几何和时间属性，不计算新的星历或动画帧。

### 转换文件

下面的程序把 `eclipse.geojson` 转成 `eclipse.kml`，可在 Google Earth 中打开：

```go
package main

import (
	"log"
	"os"

	"github.com/starainrt/astro/kml"
)

func main() {
	data, err := os.ReadFile("eclipse.geojson")
	if err != nil {
		log.Fatal(err)
	}
	result, err := kml.FromGeoJSON(data, kml.Options{
		Name: "2009-07-22 日食",
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("eclipse.kml", result, 0644); err != nil {
		log.Fatal(err)
	}
}
```

### Options

| 字段 | 零值或默认行为 | 设置后的作用 |
| --- | --- | --- |
| `Name` | 按事件类型与最早时刻推导文档名 | 指定 `Document/name` |
| `Language` | `"zh"` | `"en"` 使用英文图层名；未收录的角色保留原始 `role` |
| `Styles` | 内置配色 | 以 `role` 或 `event/role` 为键覆盖样式，后者优先 |
| `NoLookAt` | 自动设置取景 | `true` 不写 `Document/LookAt` |
| `SkipRoles` | 保留所有角色 | 按 `role` 删除整层，包括几何、样式和取景贡献 |
| `FillContext` | 只填充高亮中心带和掩星带 | `true` 为其他面图层添加浅灰半透明填充 |
| `NoTimes` | 写入输入中存在的时间 | `true` 不输出时间元素，生成静态叠加 |

### 图层与样式

要素按 `event` 和 `role` 分组到 Folder。同一集合含多个事件类型时，图层名带事件前缀；第三方 `event` 与 `role` 也可使用相同的样式覆盖规则。

| 图层 | role | 默认样式 |
| --- | --- | --- |
| 日食中心带与本影 | `central-band`、`central-shadow`、`central-shadow-footprint`、`central-shadow-sweep`、`total-footprint` | 红色；面填充约 35% 不透明度 |
| 中心线 | `center-line` | 黑色，3 px |
| 日食食甚等时线 | `greatest-time-line` | 绿色 |
| 食分线 | `magnitude-line`、`magnitude-one-envelope` | 黄色 |
| 掩星全掩与偏掩带 | `total-band`、`partial-band`、`total-band-outline`、`occultation-band` | 黄色，带半透明填充 |
| 升落可见边界、月食时间包络、日食偏食带边缘 | `visibility-boundary`、`p1-horizon`、`p4-horizon`、`visible-at-p1`、`visible-at-p4`、`visible-during-eclipse`、`visible-throughout-eclipse`、日食的 `partial-band` | 橙色 |
| 仅见半影的月出/月落带 | `penumbra-moonrise`、`penumbra-moonset` | 月出蓝紫、月落紫红，带半透明填充 |
| 凌日可见分区与接触线 | `transit-entire`、`transit-at-sunrise`、`transit-at-sunset`、`transit-partial`、`c1-horizon`、`c4-horizon`、`visible-at-c1`、`visible-at-c4`、`c1-isochrone` … `c4-isochrone` | 三个分区与部分凌日可见区各有填充色；C1/C4 视地平与可见半球橙色；四触等时线各一色 |
| 其他限界、轮廓、足迹与时间标记 | 其他 `role` | 灰色，面默认只画轮廓 |

`Style` 的零值字段沿用默认样式：

| 字段 | 含义 |
| --- | --- |
| `LineColor` | 线色，KML `aabbggrr` 十六进制 |
| `FillColor` | 面填充色；点和线忽略此项 |
| `LineWidth` | 像素；小于等于 0 时使用默认线宽 |
| `NoFill` | 强制不填充，优先于 `FillColor` 与默认配色 |

KML 颜色顺序是透明通道、蓝、绿、红。例如 `ff0000ff` 为不透明红，`590000ff` 为约 35% 不透明度的红；它与 CSS `rrggbb` 的顺序不同。

```go
options := kml.Options{
    Language: "zh",
    Styles: map[string]kml.Style{
        "center-line": {LineColor: kml.ColorBlack, LineWidth: 4},
        "solar-eclipse/central-band": {FillColor: "590000ff"},
        "partial-footprint": {NoFill: true},
    },
}
result, err := kml.FromGeoJSON(data, options)
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(result))
```

无填充的面会转成轮廓线。有填充的面与描边分别表示，反经线拆分产生的接缝不作为真实边界绘制。这样既可保持面闭合，也可避免沿 ±180° 经线出现贯穿地图的描边。

### 时间轴与静态叠加

| GeoJSON 时间数据 | KML 输出 |
| --- | --- |
| 单个 `time` 属性 | 包含该要素全部子几何的 Placemark 带一个 `TimeStamp` |
| MultiLineString 的 `times` | 按线段拆成 Placemark，各段取首个时间作为 `TimeStamp` |
| 至少一个有效时间戳 | Document 带覆盖最早至最晚时间戳的 `TimeSpan` |
| `time_scale: "UT1"` | 先按当前时标模型换回 UTC，再写 `<when>` |
| `NoTimes: true` | 不输出 `TimeStamp` 或 `TimeSpan` |

时间戳保留源数据的小数秒。要素含 `label` 时用它作名称；自动生成的名称只显示到秒。UT1 数据的原 `time_scale` 保留在属性中；转换时应与生成 GeoJSON 时使用同一 ΔT 模型。

时间轴中的帧来自输入数据。`times` 不会让折线上的每个顶点自动变成独立动画帧；需要本影或半影逐时刻变化时，应先用 GeoJSON 侧的采样选项生成这些要素。

Google Earth 会按所选时间窗口隐藏带时间戳的要素。查看静态全路径时可设置 `NoTimes: true`；播放时则将时间窗口移到事件日期，并调整可见时间段的宽度。

### 控制文件大小和取景

半影足迹通常覆盖很大区域。密集采样、大量顶点与半透明面重叠都会增加客户端绘制开销；可增大采样步长，或按用途去掉部分图层。

只需要总偏食区边缘和中心食路径时，可跳过 `partial-footprint`，保留 `partial-band`：

```go
result, err := kml.FromGeoJSON(data, kml.Options{
    SkipRoles: []string{"partial-footprint"},
    NoTimes: true,
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(result))
```

若在生成 GeoJSON 时已经确定不需要某层，可通过 `geojson.MarshalSolarEclipseWithOptions` 的 `SkipRoles` 先排除，减少中间数据。KML 的 `SkipRoles` 适用于已有文件。

自动取景优先使用中心线、中心带、限界和本影等路径图层，缺少这些图层时再使用全部要素。包络考虑跨反经线情况，`LookAt/range` 的单位为米，下限 200 km。需要客户端自行定位时设 `NoLookAt: true`。

### 属性与输入校验

通过 `SkipRoles` 筛选后，所有保留要素都具有且值相同的属性才提升到 `Document/ExtendedData`。其余属性留在各自 Placemark，嵌套对象也会保留；`times` 已转换成时间戳，不重复写入属性。

坐标至少含经纬度两项。经度超出 ±180° 时回绕，原本的 +180° 与 −180° 保持不变；纬度必须在 ±90° 内。额外的高度分量被忽略，输出采用 `clampToGround`。

多边形环会补闭合，并调整为外环逆时针、内环顺时针。支持 Point、MultiPoint、LineString、MultiLineString、Polygon、MultiPolygon 和 GeometryCollection。

以下情况返回错误：JSON 或几何类型无效、坐标非有限或纬度越界、时间数组与线段结构不匹配、过滤后没有剩余要素，或所有要素均没有可绘制几何。两个相同坐标组成的单点路径片段会被跳过。输入本来就是空 FeatureCollection 时，返回合法的空 Document。
