# 恒星

[English](en/star.md) | [返回 README](../../README.md)

`star` 包内置 9100 颗恒星数据库（BSC / HR `1–9110`，视星等 `-1.46`～`7.96`，含中文星官名、英文常用名与自行数据），支持任意时刻位置归算、星座判定与升落计算；`star/catalog` 子包支持加载 CSV、Hipparcos、BSC5 与 Stellarium 外挂星表。

## 目录

- [基本示例](#基本示例)
- [主要功能](#主要功能)
  - [星名查询与重名消歧](#星名查询与重名消歧)
  - [亮星样本与批量查询](#亮星样本与批量查询)
  - [外挂星表](#外挂星表)
- [API 概览](#api-概览)
- [约定与限制](#约定与限制)

## 基本示例

星表存储的是 **J2000 历元**坐标，用于具体时刻的升落、高度角或星座判定前，须先调用 `StarData.RaDecByDate(date)` 叠加自行、岁差与章动：

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro/star"
	"github.com/starainrt/astro/tools"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2026, 1, 1, 20, 0, 0, 0, cst)
	lon, lat := 116.4, 39.9 // 北京

	sirius, err := star.StarDataByName("天狼") // 亦支持 "Sirius" 或 StarDataByHR(2491)
	if err != nil {
		log.Fatal(err)
	}
	ra, dec := sirius.RaDecByDate(date)

	rise, err := star.RiseTime(date, ra, dec, lon, lat, 0, true)
	if err != nil {
		log.Fatal(err)
	}
	culm := star.CulminationTime(date, ra, lon)
	set, err := star.SetTime(date, ra, dec, lon, lat, 0, true)
	if err != nil {
		log.Fatal(err)
	}
	alt := star.ApparentAltitude(date, ra, dec, lon, lat, 1010, 10)
	az := star.Azimuth(date, ra, dec, lon, lat)

	fmt.Printf("天狼星 (%s / %s, %.2f等): 当日 RA=%s Dec=%s\n",
		star.Constellation(ra, dec, date), star.ConstellationCode(ra, dec, date),
		sirius.Mag, tools.Format(ra/15, 1), tools.Format(dec, 0))
	fmt.Printf("2026-01-01 20:00 北京: 视高度 %.2f°  方位 %.2f°\n", alt, az)
	fmt.Printf("升起 %s  中天 %s  落下 %s\n",
		rise.Format("01-02 15:04:05"), culm.Format("01-02 15:04:05"), set.Format("01-02 15:04:05"))
}
```

输出：

```text
天狼星 (大犬座 / CMA, -1.46等): 当日 RA=6h46m17.61s Dec=-16°45′12.59″
2026-01-01 20:00 北京: 视高度 7.98°  方位 119.66°
升起 12-31 19:15:16  中天 01-01 00:19:18  落下 01-01 05:23:20
```

## 主要功能

### 星名查询与重名消歧

- `StarDataByName` 同时匹配中文名、中文别名与内置英文常用名（大小写不敏感，如 `"天狼"`、`"织女一"`、`"sirius"`、`"Vega"`）。
- `StarDataByHR` 按亮星编号唯一定位。
- 部分中文星名（如 `"杵增一"`）或英文名（如 `"Gienah"`）对应多颗恒星，此时 `StarDataByName` 返回 `star.ErrStarNameAmbiguous`（未命中返回 `star.ErrStarNotFound`，使用 `errors.Is` 判断）。可通过 `StarDataByNameIn("杵增一", "危宿")` 按宿/垣消歧，或调用 `StarDataByNameAll` 获取全部候选。

### 亮星样本与批量查询

内置星表在首次访问时懒加载并保持只读缓存，可安全跨请求复用；如需启动时显式预加载并提前检查错误，可调用 `star.InitStarDatabase()`。`star.TopBrightStars()` 返回 169 颗较亮恒星样本（视星等约 ≤ 3）：

```go
bright, err := star.TopBrightStars()
if err != nil {
	log.Fatal(err)
}
fmt.Printf("内置最亮星样本数: %d，前三: %s(%.2f), %s(%.2f), %s(%.2f)\n",
	len(bright), bright[0].ChineseName, bright[0].Mag,
	bright[1].ChineseName, bright[1].Mag, bright[2].ChineseName, bright[2].Mag)
```

输出：

```text
内置最亮星样本数: 169，前三: 天狼(-1.46), 老人(-0.72), 大角(-0.04)
```

### 外挂星表

`star/catalog` 独立于内置 9100 星表，不自动下载数据文件。读入记录的 `entry.Coordinate` 为 `basic.StarCoordinate`，可直接传入 `moon.FindStarOccultations` 或 `phenomena.StarApproaches`：

```go
input := strings.NewReader("id,ra,dec,pmra,pmdec,V\nSirius,101.28715533,-16.71611586,-546.01,-1223.07,-1.46\n")
stars, err := catalog.ReadCSV(input, catalog.ColumnMap{
	ID: "id", RA: "ra", Dec: "dec", PMRA: "pmra", PMDec: "pmdec", Magnitude: "V",
	PMRAIncludesCosDec: true, Band: catalog.BandV,
	Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: basic.CoordinateFrameJ2000,
}, catalog.ReadOptions{Strict: true})
if err != nil {
	log.Fatal(err)
}
entry, ok := stars.Lookup("Sirius")
```

| 入口 | 说明 |
| --- | --- |
| `ReadCSV` | 带表头的 CSV / TSV，经 `ColumnMap` 指定列名、历元与参考系 |
| `ReadHipparcos2007` / `ReadBSC5` | 固定宽度文本星表 |
| `ReadStellariumStars` / `ReadStellariumDir` | Stellarium 二进制分片或 `fs.FS` 目录（支持 `MaxMagnitude` 过滤） |
| `Lookup` / `Entries` / `Len` / `Near` / `Band` / `Merge` / `WithNames` | 主 ID/别名查找、锥形与赤纬带检索（按星表自身历元）、合并与名称挂载 |

`ColumnMap` 须显式提供 `Epoch`（或 `EpochColumn`）与 `Frame`；角度默认为度，自行单位 mas/年，视差单位 mas，径向速度单位 km/s。

## API 概览

| 分类 | 入口 | 说明 |
| --- | --- | --- |
| 恒星库 | `InitStarDatabase`、`StarDataByHR`、`StarDataByName`、`StarDataByNameIn`、`StarDataByNameAll`、`TopBrightStars` | 哨兵错误：`ErrStarNotFound`、`ErrStarNameAmbiguous` |
| 坐标归算 | `StarData.RaDecByDate(date)` | J2000 → 指定时刻（自行 + 岁差 + 章动）；`Pc > 0` 时采用三维空间运动 |
| 星座判定 | `Constellation` / `ConstellationEN` / `ConstellationCode` | 中文名、英文名与 IAU 三字母代码（入参为当日视坐标） |
| 升落与中天 | `RiseTime` / `SetTime` / `CulminationTime` / `RiseTimeWithAtmosphere` / `SetTimeWithAtmosphere` | 极区返回 `ERR_STAR_NEVER_RISE` / `ERR_STAR_NEVER_SET` |
| 站心量 | `Altitude` / `ApparentAltitude` / `Azimuth` / `Zenith` / `ApparentZenith` / `HourAngle` / `ParallacticAngle` | 度；`Apparent*` 需气压（hPa）与气温（℃） |
| 恒星时 | `MeanSiderealTime` / `ApparentSiderealTime` | 格林尼治平/真恒星时（小时） |

## 约定与限制

- **拱极与永不升起恒星**：`SetTime` 返回 `star.ERR_STAR_NEVER_SET`，`RiseTime` 返回 `star.ERR_STAR_NEVER_RISE`（`DownTime` 与 `ERR_STAR_NEVER_DOWN` 为废弃兼容别名）。
- **三维自行**：记录包含距离（`Pc > 0`）时，`RaDecByDate` 结合 `PmRA`、`PmDec` 与 `RadVel` 执行三维空间运动推进；无距离记录退回二维角分量推进。
- **大气折射区间**：`ApparentAltitude` 仅在真高度角 `(-5°, 90°)` 内施加 Saemundsson 折射修正；气压非正或气温低于绝对零度返回 `NaN`。
- **高程与时标**：`height` 为 WGS84 椭球高（米），见[观测点高度](coord.md#观测点高度)；时间参数为民用时刻，见[时标](timescale.md)。
