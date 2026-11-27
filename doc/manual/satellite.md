# 人造卫星

[English](en/satellite.md) | [返回 README](../../README.md) | [卫星 SVG](satellite-svg.md) | [GeoJSON / KML](map-geojson.md#人造卫星图层)

`satellite` 包提供人造卫星轨道根数（TLE / OMM）解析、SGP4/SDP4 轨道传播，以及站心观测、过境预报、凌日凌月和地面覆盖范围计算。

## 目录

- [基本示例](#基本示例)
- [主要功能](#主要功能)
  - [解析 TLE 与 OMM 根数](#解析-tle-与-omm-根数)
  - [星下点与站心观测](#星下点与站心观测)
  - [凌日与凌月搜索](#凌日与凌月搜索)
  - [可见圈与时段覆盖扫掠](#可见圈与时段覆盖扫掠)
- [API 概览](#api-概览)
- [约定与限制](#约定与限制)

## 基本示例

解析国际空间站（ISS）两行根数，预报北纬 40°、经度 0° 观测者在 2026-11-14 18:00 UTC 之后的下一次仰角 ≥ 10° 过境及可见性：

```go
package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/starainrt/astro/satellite"
)

func main() {
	tle := []byte(`ISS (ZARYA)
1 25544U 98067A   26317.91612726  .00008668  00000+0  16320-3 0  9995
2 25544  51.6324 279.1461 0007381  22.6178 337.5134 15.49359724594966`)

	elements, err := satellite.ParseTLEText(tle)
	if err != nil {
		log.Fatal(err)
	}
	iss := elements[0]

	obs := satellite.Observer{Longitude: 0, Latitude: 40, Height: 0} // 东经/北纬（度），椭球高（米）
	start := time.Date(2026, 11, 14, 18, 0, 0, 0, time.UTC)
	stdMag := 0.0 // 1000 km、50% 相位角标准星等；置 nil 则不估算星等

	pass, err := satellite.NextPass(iss, obs, start, satellite.PassOptions{
		MinElevation:   10,
		ObserveOptions: satellite.ObserveOptions{StandardMagnitude: &stdMag},
	})
	if err != nil {
		log.Fatal(err)
	}

	kindName := map[satellite.PassKind]string{
		satellite.PassVisible:  "目视可见",
		satellite.PassDaylight: "白昼",
		satellite.PassEclipsed: "地影中",
	}

	fmt.Printf("卫星: %s（NORAD %d）\n", iss.Name, iss.CatalogNumber)
	fmt.Printf("升起: %s UTC  方位 %.1f°\n", pass.Rise.Time.Format("2006-01-02 15:04:05"), pass.Rise.Azimuth)
	fmt.Printf("最高: %s UTC  仰角 %.1f°  方位 %.1f°  距离 %.0f km\n",
		pass.Culmination.Time.Format("15:04:05"), pass.Culmination.Elevation, pass.Culmination.Azimuth, pass.Culmination.RangeKM)
	fmt.Printf("落下: %s UTC  方位 %.1f°  持续 %s\n",
		pass.Set.Time.Format("15:04:05"), pass.Set.Azimuth, pass.Duration.Round(time.Second))
	fmt.Printf("可见性: %s", kindName[pass.Kind])
	if pass.Kind == satellite.PassVisible {
		fmt.Printf("（%s – %s UTC", pass.VisibleStart.Format("15:04:05"), pass.VisibleEnd.Format("15:04:05"))
		if !math.IsNaN(pass.BrightestMagnitude) {
			fmt.Printf("，最亮 %.1f 等 @ %s", pass.BrightestMagnitude, pass.BrightestTime.Format("15:04:05"))
		}
		fmt.Print("）")
	}
	fmt.Println()
}
```

输出：

```text
卫星: ISS (ZARYA)（NORAD 25544）
升起: 2026-11-14 18:19:55 UTC  方位 266.1°
最高: 18:22:58 UTC  仰角 31.6°  方位 331.7°  距离 745 km
落下: 18:26:02 UTC  方位 37.2°  持续 6m7s
可见性: 目视可见（18:19:55 – 18:23:47 UTC，最亮 -0.9 等 @ 18:23:43）
```

查询区间内的全部过境使用 `Passes(iss, obs, start, end, opt)`；仅筛选目视可见过境时在 `PassOptions` 中设置 `VisibleOnly: true`。

## 主要功能

以下片段沿用基本示例中的 `iss`、`obs`、`start` 与 `stdMag`。

### 解析 TLE 与 OMM 根数

- 单条两行根数使用 `ParseTLE(line1, line2)`；多条或含名称行的三行文本使用 `ParseTLEText(data)`。
- CCSDS OMM 格式使用 `ParseOMM(data)`，自动识别 JSON、XML 与 KVN；也可显式调用 `ParseOMMJSON`、`ParseOMMXML` 或 `ParseOMMKVN`。

```go
ommJSON := []byte(`[{
	"OBJECT_NAME": "ISS (ZARYA)",
	"OBJECT_ID": "1998-067A",
	"EPOCH": "2026-11-13T21:59:13.395264",
	"MEAN_MOTION": 15.49359724,
	"ECCENTRICITY": 0.0007381,
	"INCLINATION": 51.6324,
	"RA_OF_ASC_NODE": 279.1461,
	"ARG_OF_PERICENTER": 22.6178,
	"MEAN_ANOMALY": 337.5134,
	"EPHEMERIS_TYPE": 0,
	"CLASSIFICATION_TYPE": "U",
	"NORAD_CAT_ID": 25544,
	"ELEMENT_SET_NO": 999,
	"REV_AT_EPOCH": 59496,
	"BSTAR": 0.0001632,
	"MEAN_MOTION_DOT": 0.00008668,
	"MEAN_MOTION_DDOT": 0
}]`)

list, err := satellite.ParseOMM(ommJSON)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("OMM 解析: %s (NORAD %d, 历元 %s UTC)\n",
	list[0].Name, list[0].CatalogNumber, list[0].Epoch.Format("2006-01-02 15:04:05"))
```

输出：

```text
OMM 解析: ISS (ZARYA) (NORAD 25544, 历元 2026-11-13 21:59:13 UTC)
```

低轨卫星受大气阻力影响显著，根数距历元越远预报误差越大，实际应用应定期更新近历元根数。

### 星下点与站心观测

- `GroundPointAt(e, t)` 返回指定时刻的 WGS84 星下点经纬度与轨道高度（km）；`GroundTrack(e, start, end, step)` 按步长采样地面轨迹。
- `LookAt(e, obs, t, opt)` 返回方位角、仰角、斜距、距离变化率、站心真赤道坐标、光照状态与估算星等。

```go
at := start.Add(23 * time.Minute)
gp, err := satellite.GroundPointAt(iss, at)
if err != nil {
	log.Fatal(err)
}
look, err := satellite.LookAt(iss, obs, at, satellite.ObserveOptions{StandardMagnitude: &stdMag})
if err != nil {
	log.Fatal(err)
}
fmt.Printf("星下点: 经度 %.2f°  纬度 %.2f°  高度 %.1f km\n", gp.Longitude, gp.Latitude, gp.HeightKM)
fmt.Printf("站心观测: 方位 %.1f°  仰角 %.1f°  距离 %.0f km  星等 %.1f\n",
	look.Azimuth, look.Elevation, look.RangeKM, look.Magnitude)
```

输出：

```text
星下点: 经度 -3.47°  纬度 44.71°  高度 419.0 km
站心观测: 方位 332.5°  仰角 31.5°  距离 745 km  星等 -0.5
```

`LookAt` 默认计入大气折射；如需与无折射几何覆盖圈保持同一仰角门限，在 `ObserveOptions` 中设置 `NoRefraction: true`。

### 凌日与凌月搜索

- 全球凌带搜索使用 `TransitEvents`（`Targets` 可指定 `TransitSun`、`TransitMoon`，留空则同时搜索两者）。
- 固定测站的接触时刻与持续时长使用 `Transits`；批量筛查多颗卫星距站点一定距离内的凌现象使用 `NearbyTransits`。

```go
tStart := time.Date(2026, 11, 14, 0, 40, 0, 0, time.UTC)
events, err := satellite.TransitEvents(iss, tStart, tStart.Add(40*time.Minute), satellite.TransitSearchOptions{
	Targets: []satellite.TransitTarget{satellite.TransitMoon},
})
if err != nil {
	log.Fatal(err)
}
ev := events[0]
fmt.Printf("凌月带: %s – %s UTC  最大半宽 %.1f km  中心点 (%.2f°E, %.2f°N)\n",
	ev.Start.Format("15:04:05"), ev.End.Format("15:04:05"),
	ev.MaxHalfWidthKM, ev.GreatestPoint.Longitude, ev.GreatestPoint.Latitude)

site := satellite.Observer{Longitude: ev.GreatestPoint.Longitude, Latitude: ev.GreatestPoint.Latitude}
transits, err := satellite.Transits(iss, site, ev.Start.Add(-time.Minute), ev.End.Add(time.Minute), satellite.TransitOptions{
	Targets: []satellite.TransitTarget{satellite.TransitMoon},
})
if err != nil {
	log.Fatal(err)
}
tr := transits[0]
fmt.Printf("本站凌月: 中心时刻 %s UTC  持续 %.2f s  最小角距 %.1f\"  月面视半径 %.1f\"\n",
	tr.Center.Format("15:04:05.00"), tr.Duration.Seconds(), tr.MinSeparationArcsec, tr.TargetRadiusArcsec)
```

输出：

```text
凌月带: 00:43:16 – 01:17:54 UTC  最大半宽 23.5 km  中心点 (-161.77°E, -12.48°N)
本站凌月: 中心时刻 01:00:34.27 UTC  持续 0.53 s  最小角距 0.3"  月面视半径 896.6"
```

- 绘制全球凌带地图时，调用 `TransitEventPath(iss, ev, time.Second)` 生成中心线与左右限，交由 [satellite/svg](satellite-svg.md#1-凌月带地图transitmapsvg) 或 [GeoJSON / KML](map-geojson.md#人造卫星图层) 处理。
- 绘制本站盘面轨迹时，调用 `TransitFrames(iss, site, tr, 50*time.Millisecond)` 交由 [TransitLocalSVG](satellite-svg.md#2-本站凌月盘面图transitlocalsvg)。

### 可见圈与时段覆盖扫掠

- `CoverageAt(e, t, opt)` 计算单时刻的几何可见圈（`Footprint`）、地球夜区（`Night`）及目视可见区（优先使用 `VisibleRegions`，为空时回退至 `Visible`）。
- `CoverageBetween(e, start, end, step, opt)` 计算区间内的地面轨迹、几何扫掠区域（`Swept`）与目视扫掠区域（`VisibleSwept`）。

```go
cov, err := satellite.CoverageAt(iss, start.Add(23*time.Minute), satellite.CoverageOptions{MinElevation: 10})
if err != nil {
	log.Fatal(err)
}
sweep, err := satellite.CoverageBetween(iss, start, start.Add(time.Hour), 30*time.Second, satellite.CoverageOptions{MinElevation: 10})
if err != nil {
	log.Fatal(err)
}
rKM, _ := satellite.SphericalFootprintRadiusKM(cov.SubPoint.HeightKM, 10)
visCount := len(cov.VisibleRegions)
if visCount == 0 {
	visCount = len(cov.Visible)
}
fmt.Printf("可见圈: 高度 %.1f km  10°仰角覆盖半径约 %.0f km  目视区段数=%d\n",
	cov.SubPoint.HeightKM, rKM, visCount)
fmt.Printf("1小时扫掠: 轨迹点=%d  扫掠区域=%d  目视区域=%d\n",
	len(sweep.Track), len(sweep.Swept), len(sweep.VisibleSwept))
```

输出：

```text
可见圈: 高度 419.0 km  10°仰角覆盖半径约 1387 km  目视区段数=1
1小时扫掠: 轨迹点=121  扫掠区域=1  目视区域=2
```

覆盖结果可直接传给 [CoverageSVG](satellite-svg.md#3-地面可见范围图coveragesvg) 渲染，或经 `geojson.MarshalSatelliteCoverage` / `MarshalSatelliteCoverageSweep` 导出。

## API 概览

| 分类 | 入口 | 说明 |
| --- | --- | --- |
| 根数解析 | `ParseTLE` / `ParseTLEText` | 解析单条或多条两行/三行 TLE |
| | `ParseOMM` / `ParseOMMJSON` / `ParseOMMXML` / `ParseOMMKVN` | 解析 CCSDS OMM 根数 |
| 轨道传播 | `Propagate` / `GroundPointAt` / `GroundTrack` | TEME 位置（km）与速度（km/s）、WGS84 星下点 |
| 站心与过境 | `LookAt` / `NextPass` / `Passes` | 方位、仰角、斜距、光照、星等与过境可见性 |
| 凌日与凌月 | `Transits` / `TransitFrames` | 本站接触时刻与盘面轨迹帧；`MaxSeparationFactor > 1` 可检出近掠 |
| | `TransitEvents` / `ClosestTransitEvent` / `TransitEventPath` / `TransitShadowAt` / `NearbyTransits` | 全球凌带、中心线与左右限、瞬时阴影足迹与批量附近筛选 |
| 覆盖范围 | `CoverageAt` / `CoverageBetween` / `SphericalFootprintRadiusKM` | 瞬时可见圈、夜区、目视区、扫掠并集与球面半径估算 |
| 显式星历 | `NewWithInstance` | 绑定显式 `*astro.Instance`（或经 `ctx.Satellite()`） |

哨兵错误（使用 `errors.Is` 判断）：`ErrInvalidElements`、`ErrInvalidInput`、`ErrDecayed`、`ErrMeanMotion`、`ErrMeanEccentricity`、`ErrPerturbedEccentricity`、`ErrSemiLatusRectum`、`ErrPassNotFound`、`ErrTransitNotFound`。

## 约定与限制

- **传播模型与坐标系**：近地轨道采用 SGP4，深空轨道（周期 ≥ 225 分钟）采用 SDP4，在 TEME 坐标系传播后旋转至 WGS84 地固系与站心系。输入与输出均为民用时刻，输出保持查询起点时区。
- **时间跨度与 OMM 校验**：`Propagate` 拒绝超过 292 年时间跨度（`time.Duration` 限制）。OMM 历元无时区后缀时按 UTC 解释，非 SGP4 理论或参数越界时返回错误。
- **凌现象几何口径**：卫星按点、日月按视圆盘在无折射 WGS84 椭球面上求解；`TransitEvent.Complete == false` 表示凌带被查询窗口截断。
- **极区与带孔覆盖**：极地或高轨卫星的可见圈可能包裹极点或包含昼区内孔。绘图与 GIS 导出应优先读取 `Coverage.VisibleRegions`（已做反经线拆分与孔洞表达），为空时再回退至 `Coverage.Visible`。
- **星等估算**：仅当提供 `ObserveOptions.StandardMagnitude` 且卫星不在地球本影中时，`Look.Magnitude` 与 `Pass.BrightestMagnitude` 返回有限值，其余情况为 `NaN`。
