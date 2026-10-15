# 恒星

[English](en/star.md) | [返回 README](../../README.md)

> 本手册的完整示例以仓库根目录为工作目录执行。

本程序自带 9100 颗恒星的数据库（BSC / HR 编号 `1–9110`，视星等 `-1.46`～`7.96`），能够自动计算自行。星表里存的是 **J2000 历元**的赤经赤纬（`InnerStarData.Ra`/`Dec`），要得到某一时刻的位置必须再做自行、岁差与章动修正，入口是 `StarData.RaDecByDate(date)`；升落、站心量与星座判定都应传入修正后的赤经赤纬。

## 目录

- [外挂星表](#外挂星表)
- [查询天狼星的位置与升起时刻](#查询天狼星的位置与升起时刻)
- [API 参考](#api-参考)
  - [星座判定](#星座判定)
  - [恒星库](#恒星库)
  - [升落与中天](#升落与中天)
  - [站心量](#站心量)
  - [恒星时](#恒星时)
  - [遍历亮星并计算观测量](#遍历亮星并计算观测量)
  - [完整示例](#完整示例)
  - [批量查询与缓存](#批量查询与缓存)
- [常用场景](#常用场景)
  - [今晚能不能看到这颗星](#今晚能不能看到这颗星)
  - [把 J2000 位置用到具体时刻](#把-j2000-位置用到具体时刻)
  - [星座判定与亮星表](#星座判定与亮星表)
  - [极区边界与星表口径](#极区边界与星表口径)
- [参数与返回值约定](#参数与返回值约定)
- [相关手册](#相关手册)

## 外挂星表

`star/catalog` 与内置 9100 星数据库独立，不自动下载或加载大型星表。读入的 `Entry.Coordinate` 是带历元、坐标框架和自行的 `basic.StarCoordinate`，可直接用于 `moon.FindStarOccultations` 或 `phenomena.StarApproaches`。

| 入口 | 输入与用途 |
| --- | --- |
| `ReadCSV` | 带表头的 CSV/TSV，使用 `ColumnMap` 指定列名、历元与框架；`Delimiter` 为 0 时自动识别 |
| `ReadHipparcos2007` / `ReadBSC5` | 读取对应的固定宽度文本星表 |
| `ReadStellariumStars` / `ReadStellariumDir` | 单个二进制分片或 `fs.FS` 目录；目录选项可加载名称及交叉 ID，按 `MaxMagnitude` 过滤 |
| `Lookup` / `Entries` / `Len` | 按主 ID 或别名查询，取得记录副本或数量 |
| `Near` / `Band` | 星表自身历元下的锥形搜索/赤纬带搜索，不自动传播到观测日期 |
| `Merge` / `WithNames` | 返回新星表；合并时后者覆盖同主 ID、同分量记录，名称文件使用 `id\|name` 行 |

```go
import (
	"strings"
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/star/catalog"
)

input := strings.NewReader("id,ra,dec,pmra,pmdec,V\nSirius,101.28715533,-16.71611586,-546.01,-1223.07,-1.46\n")
stars, err := catalog.ReadCSV(input, catalog.ColumnMap{
	ID: "id", RA: "ra", Dec: "dec", PMRA: "pmra", PMDec: "pmdec", Magnitude: "V",
	PMRAIncludesCosDec: true, Band: catalog.BandV,
	Epoch: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), Frame: basic.CoordinateFrameJ2000,
}, catalog.ReadOptions{Strict: true})
if err != nil {
	panic(err)
}
entry, found := stars.Lookup("Sirius")
_ = entry.Coordinate
_ = found
```

单位与必填项：

- `RA`、`Dec` 默认是度，`AnglesInRadians` 可改为弧度；自行是 mas/年，视差是 mas，径向速度是 km/s。
- 若 `PMRA` 已经是 `dRA·cos(Dec)`，设置 `PMRAIncludesCosDec: true`，否则读入器补乘余弦。
- 必须给出 `Epoch` 或 `EpochColumn`（儒略年）以及正确的 `Frame`，不得用零历元代替未知历元。

读取行为：

- `ReadOptions.Strict` 为 true 时坏记录返回错误，否则跳过并计入 `Skipped()`；缺少必需表头这类结构错误始终返回错误。
- `Filter` 只控制入选，不计为坏记录。
- 未提供星等时 `Magnitude` 为 `NaN`；`BandV`、`BandG`、`BandHp` 不应不经换算就用于同一 V 波段可见性阈值。
- `Catalog.Near` 在各记录自身历元搜索；混合历元或框架的数据应先统一，观测日期的精确筛选仍需自行传播。

## 查询天狼星的位置与升起时刻

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro/star"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2020, 1, 1, 8, 8, 8, 0, cst)
	lon, lat, height := 115.0, 40.0, 0.0
	sirius, err := star.StarDataByName("天狼")
	if err != nil {
		log.Fatal(err)
	}
	ra, dec := sirius.RaDecByDate(date)
	rise, err := star.RiseTime(date, ra, dec, lon, lat, height, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(star.Constellation(ra, dec, date))
	fmt.Printf("RA=%.6f Dec=%.6f deg\n", ra, dec)
	fmt.Println(rise.Format(time.RFC3339))
}
```

星表坐标的历元是 J2000。`RaDecByDate` 加入自行、岁差与章动后，才用于指定日期的升落、高度角和星座判定。

自行归算按**儒略年**（365.25 日）计，`RaDecByDate` 先把民用时刻换算成 TT 再取历元差。只要记录带距离（`Pc > 0`），`RaDecByJde` 就按**三维空间运动**推进：自行给出切向速度、`RadVel` 给出视向分量，位置矢量线性外推后取方向；没有距离的记录退回二维，即只推进赤经赤纬两个角分量，等价于把恒星当作无穷远。两者的差别是二阶项：三维推进同时改变恒星距离，视位置比二维推进略小，26 年内最亮的高自行星也只有约 0.09 角秒。

## API 参考

| 分组 | 入口 | 用途 | 单位与口径 |
| --- | --- | --- | --- |
| 星座判定 | `Constellation` / `ConstellationEN` / `ConstellationCode` | 星座中文名 / 英文名 / IAU 三字母代码 | 输入当日赤经赤纬（度）与时刻 |
| 恒星库 | `InitStarDatabase` / `StarDataByHR` / `StarDataByName` / `TopBrightStars` | 初始化内置星表、按 HR 编号或中文名取星、取最亮恒星样本 | HR `1–9110`；`Mag` 为视星等 |
| 坐标修正 | `StarData.RaDecByDate`（`basic` 侧对应 `RaDecByJde`） | 把 J2000 位置修正到指定时刻 | 返回赤经赤纬，单位度 |
| 升落与中天 | `RiseTime` / `SetTime` / `CulminationTime`（`DownTime` 是 `SetTime` 的废弃别名） | 升起、落下、中天时刻 | 返回民用时刻；极区返回哨兵错误 |
| 带气象升落 | `RiseTimeWithAtmosphere` / `SetTimeWithAtmosphere` | 按实测气压（hPa）、气温（℃）计算升落 | 恒带大气折射 |
| 站心量 | `Altitude` / `ApparentAltitude` / `Azimuth` / `Zenith` / `ApparentZenith` | （视）高度角、方位角、（视）天顶距 | 度 |
| 时角与视差角 | `HourAngle` / `ParallacticAngle` | 恒星时角、天顶方向视差角 | 度 |
| 恒星时 | `MeanSiderealTime` / `ApparentSiderealTime` | 平恒星时、真恒星时 | 小时 |

`star` 包没有 `...N` 入口；带 `n` 参数的求值接口在 `sun`、`moon`、行星与 `coord` 的对应函数上，语义见各自手册（`n < 0` 用完整级数；半解析模型下 `n >= 0` 用快速表，Legacy 按 `n` 截断）。

以下片段省略公共前置变量：`date`（观测时刻，民用时标）、`lon`/`lat`（观测点经纬度，东经/北纬为正，度）、`height`（观测点高度，**椭球高**，米）、`aero`（是否计入蒙气差与视半径修正）。

### 星座判定

```go
sirius, _ := star.StarDataByName("天狼")
ra, dec := sirius.RaDecByDate(date)
fmt.Println(star.Constellation(ra, dec, date))     // 大犬座
fmt.Println(star.ConstellationEN(ra, dec, date))   // Canis Major
fmt.Println(star.ConstellationCode(ra, dec, date)) // CMA
```

三个入口共用同一份星座边界表，只是输出口径不同；判定用的是**当日视位置**，所以必须传入 `RaDecByDate` 的结果而不是星表里的 J2000 坐标。

### 恒星库

```go
_ = star.InitStarDatabase()
s, _ := star.StarDataByHR(2491)
fmt.Println(s.HR, s.ChineseName, s.CommonName, s.Mag) // 2491 天狼 Sirius -1.46
bright, _ := star.TopBrightStars()
fmt.Println(len(bright), bright[0].HR) // 最亮恒星样本条数与首项 HR
```

星表在首次查询时自动加载，`InitStarDatabase()` 只是提前预热：它幂等，并且能把首次加载的错误显式暴露出来，不调用也会在第一次查询时自动加载。`TopBrightStars()` 返回 169 颗视星等约不高于 3、按亮到暗大致排列的内置样本。

`basic.StarData` 在 `InnerStarData` 之上补了名称字段；字段口径如下：

| 字段 | 含义 |
| --- | --- |
| `HR` / `HD` / `HIP` | 亮星编号（`1–9110`）/ 亨利·德雷伯编号 / 依巴谷编号 |
| `Ra` / `Dec` | **J2000** 赤经、赤纬，单位度（要用当日位置先过 `RaDecByDate`） |
| `Mag` | 视星等 |
| `PmRA` / `PmDec` | 赤经投影年自行 `cos(dec)·dRA/dt` 与赤纬年自行，单位角秒/年 |
| `RadVel` / `RotVel` | 径向速度与自行速度，单位 km/s |
| `Pc` | 距离，单位秒差距；`> 0` 时自行按三维空间运动推进 |
| `ChineseName` / `ChineseAlias` / `ChineseBayerName` | 中文名、别名、中文拜耳名 |
| `CommonName` / `CommonAliasName` | 英文常用名与别名 |
| `Cst` / `CstChinese` | 星座英文名与中文名 |

按中文名查询只匹配库内中文名（如 `天狼`、`织女一`），英文名不参与匹配；需要按编号取星用 `StarDataByHR`。

### 升落与中天

```go
ra, dec := 101.28715533, -16.71611586 // 天狼星 J2000 附近示例坐标
rise, _ := star.RiseTime(date, ra, dec, lon, lat, height, true)
set, _ := star.SetTime(date, ra, dec, lon, lat, height, true)
fmt.Println(rise, set)
fmt.Println(star.CulminationTime(date, ra, lon))
```

`aero` 为真时按标准蒙气差修正后的几何地平求升落，为假时用几何地平；`CulminationTime` 只需要赤经与经度。极夜/极昼下 `RiseTime`/`SetTime` 不返回时刻而是哨兵错误，见下文。

### 站心量

```go
fmt.Println(star.Altitude(date, ra, dec, lon, lat))
fmt.Println(star.ApparentAltitude(date, ra, dec, lon, lat, 1010, 10))
fmt.Println(star.Azimuth(date, ra, dec, lon, lat), star.Zenith(date, ra, dec, lon, lat))
fmt.Println(star.HourAngle(date, ra, lon), star.ParallacticAngle(date, ra, dec, lon, lat))
```

`ApparentAltitude`/`ApparentZenith` 多两个参数：气压（hPa）与气温（°C），用来做蒙气差修正；`ParallacticAngle` 常用于旋转相机与光谱缝方向。

### 恒星时

```go
fmt.Println(star.MeanSiderealTime(date), star.ApparentSiderealTime(date))
```

两者都返回小时；恒星时与恒星时角、地平转换的关系见 `coord` 手册。

### 遍历亮星并计算观测量

```go
bright, _ := star.TopBrightStars()
for _, s := range bright[:3] {
	ra, dec := s.RaDecByDate(date)
	alt := star.ApparentAltitude(date, ra, dec, lon, lat, 1010, 10)
	fmt.Printf("%-6s %-10s mag=%.2f alt=%.3f\n", s.ChineseName, star.ConstellationEN(ra, dec, date), s.Mag, alt)
}
```

上面这段在 `date = 2020-01-01 08:08:08 CST`、观测点 `115°E, 40°N`、气压 `1010 hPa`、气温 `10 °C` 下实际输出：

```text
天狼     Canis Major    mag=-1.46 alt=-30.180
老人     Carina         mag=-0.72 alt=-48.661
大角     Bootes         mag=-0.04 alt=68.926
```

### 完整示例

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/star"
	"github.com/starainrt/astro/tools"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	// 指定观测时刻。
	date := time.Date(2020, 1, 1, 8, 8, 8, 8, cst)

	// 初始化恒星数据库。
	_ = star.InitStarDatabase()
	sirius, _ := star.StarDataByName("天狼")
	ra, dec := sirius.RaDecByDate(date)
	// 天狼星升起时间。
	riseDate, _ := star.RiseTime(date, ra, dec, 115, 40, 0, true)
	fmt.Println(riseDate)
	// 天狼星落下时间。
	setDate, _ := star.SetTime(date, ra, dec, 115, 40, 0, true)
	fmt.Println(setDate)
	fmt.Println(star.Constellation(ra, dec, date))

	// 织女星。
	vega, _ := star.StarDataByName("织女一")
	ra, dec = vega.RaDecByDate(time.Date(13600, 1, 1, 0, 0, 0, 0, time.Local))
	// 织女星在公元 13600 年的赤经。
	fmt.Println(tools.Format(ra/15, 1))
	// 织女星在公元 13600 年的赤纬。
	fmt.Println(tools.Format(dec, 0))

	bright, _ := star.TopBrightStars()
	fmt.Println(bright[0].ChineseName, bright[0].CommonName, bright[0].Mag)
}
```

输出结果：

```text
2019-12-31 19:22:56.32125467 +0800 CST // 天狼星升起时刻
2020-01-01 05:30:39.979518949 +0800 CST // 天狼星落下时刻
大犬座 // 天狼星所在星座
6h4m44.89s // 织女一在公元 13600 年的赤经
84°19′13.71″ // 织女一在公元 13600 年的赤纬
天狼 Sirius -1.46 // 最亮恒星表第一项：中文名、英文常用名、视星等
```

### 批量查询与缓存

```go
for _, hr := range []int{2491, 2326, 5340} {
	s, err := star.StarDataByHR(hr)
	if err != nil {
		continue
	}
	ra, dec := s.RaDecByDate(date)
	fmt.Println(s.ChineseName, star.ConstellationCode(ra, dec, date), s.Mag)
}
```

星表只在第一次访问时加载一次，之后是只读缓存；跨请求、跨线程复用同一份数据即可，不需要自己缓存。查不到的编号或名字会返回错误，批量场景按 `err != nil` 跳过而不必区分错误类型。

## 常用场景

### 今晚能不能看到这颗星

```go
ra, dec := sirius.RaDecByDate(date)
fmt.Println(star.RiseTime(date, ra, dec, lon, lat, height, true)) // 升起
fmt.Println(star.SetTime(date, ra, dec, lon, lat, height, true))  // 落下
fmt.Println(star.CulminationTime(date, ra, lon))                  // 中天
fmt.Println(star.ApparentAltitude(date, ra, dec, lon, lat, 1010, 10), star.Azimuth(date, ra, dec, lon, lat))
```

`aero = true` 用标准蒙气差修正后的地平，更接近"刚露出地平"的目视时刻；`false` 是几何地平，两者差 2–3 分钟量级。判断此刻能否看到用视高度 `ApparentAltitude`，`1010 hPa / 10 °C` 是常用默认气象值。`height` 是**椭球高**（米），只有海拔正高时要先加大地水准面差距，见[观测点高度约定](coord.md#观测点高度)。

### 把 J2000 位置用到具体时刻

```go
ra, dec := sirius.RaDecByDate(date)                        // J2000 -> 当日（自行 + 岁差 + 章动）
fmt.Println(star.ApparentSiderealTime(date))               // 真恒星时，单位小时
fmt.Println(star.HourAngle(date, ra, lon))                 // 时角，单位度
```

星表存的是 J2000，跳过 `RaDecByDate` 会让升落与星座判定出现度级偏差。恒星时、地平转换与岁差章动的细节见[坐标工具](coord.md)；时间参数一律民用时刻（UTC 标签，1972-01-01 前等于 UT1），约定见[时标声明](map-geojson.md#时标声明)。

### 星座判定与亮星表

```go
bright, _ := star.TopBrightStars()
for _, s := range bright[:3] {
	ra, dec := s.RaDecByDate(date)
	fmt.Println(s.ChineseName, star.ConstellationCode(ra, dec, date), s.Mag, s.CommonName)
}
```

```text
天狼 CMA -1.46 Sirius
老人 CAR -0.72 Canopus
大角 BOO -0.04 Arcturus
```

星座三个入口同源、只是输出口径不同：`Constellation` 给中文名、`ConstellationEN` 给英文名、`ConstellationCode` 给 IAU 三字母代码；判定用的是当日视位置。星表字段（`HR`/`HD`/`HIP`、J2000 赤经赤纬、视星等、自行、视差距离与各名称字段）见 [API 参考](#恒星库) 的字段表。

### 极区边界与星表口径

```go
_, err := star.RiseTime(date, ra, dec, 0, 89, 0, true)
if errors.Is(err, star.ERR_STAR_NEVER_RISE) || errors.Is(err, star.ERR_STAR_NEVER_SET) {
	fmt.Println("该日无升落：", err)
}
```

- **极区哨兵错误**：`RiseTime` 在极夜返回 `star.ERR_STAR_NEVER_RISE`（该日永远在地平线下），`SetTime` 在极昼返回 `star.ERR_STAR_NEVER_SET`（该日永远在地平线上），用 `errors.Is` 判定；`ERR_STAR_NEVER_DOWN` 是 `ERR_STAR_NEVER_SET` 的废弃别名，`DownTime` 是 `SetTime` 的废弃别名。
- **蒙气差模型的区间**：Saemundsson 近似只在真高度角 `(-5°, 90°)` 内生效，区间外修正量为 0——深在地平线以下时视高度角与几何高度角完全相等。
- **星表口径**：BSC / HR `1–9110`，视星等 `-1.46`～`7.96`；按名字查询只匹配库内中文名（如 `天狼`、`织女一`），英文名不参与。
- **与外部星表对表**：库内位置是 J2000 平位置叠加自行与岁差章动；与外部星表逐位对表时先用 `RaDecByDate` 统一到同一时刻，再比较。

## 参数与返回值约定

- **单位**：赤经、赤纬、高度角、天顶距、方位角、时角、视差角均为**度**，恒星时为**小时**；`tools.Format` 负责把它们格式化成度分秒或时分秒展示。星表里的 `PmRA`/`PmDec` 是角秒/年，`RadVel`/`RotVel` 是 km/s，`Pc` 是秒差距。
- **时标**：所有公开 API 的时间参数与返回值都是民用时刻（UTC 标签，1972-01-01 之前等于 UT1），约定见[时标声明](map-geojson.md#时标声明)。
- **坐标口径**：星表是 J2000；`RaDecByDate` 叠加自行、岁差与章动。直接用 `InnerStarData.Ra`/`Dec` 会让升落与星座判定出现可观偏差（百年量级的岁差就是度级）。
- **观测点高度**：`height` 是椭球高（大地高），单位米；手上只有海拔正高时要先加大地水准面差距，约定见「观测点高度约定」。
- **蒙气差模型的有效区间**：`ApparentAltitude`/`ApparentZenith` 用的 Saemundsson 近似只在真高度角 `(-5°, 90°)` 内生效，区间外修正量为 0——所以深在地平线以下时视高度角与几何高度角完全相等（示例里天狼星 `alt=-30.180` 与 `Altitude` 同值就是这个原因）。气压必须为正、气温必须高于绝对零度，否则返回 `NaN`。
- **`aero` 语义**：真值按标准蒙气差修正后的地平求升落，假值走几何地平；两者的差值在低纬度约 2–3 分钟，高纬度会明显放大。
- **星表范围**：HR `1–9110`、视星等 `-1.46`～`7.96`，按中文名查询只匹配库内中文名。

## 相关手册

- 恒星时、地平转换、岁差与章动、视差角：[坐标工具](coord.md)
- 升落语义与太阳/月亮对照：[太阳与月亮](sun-moon.md)
- 出图时标：[时标声明](map-geojson.md#时标声明)
