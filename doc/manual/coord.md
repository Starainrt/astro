# 坐标工具

[English](en/coord.md) | [返回 README](../../README.md)

`coord` 提供天球坐标转换与观测辅助计算。没有特殊说明时，角度单位为度；恒星时单位为小时；`time.Time` 按绝对时刻使用，内部转换为 UTC 后计算。

高度角、天顶距与地平可见性的语义沿用[观测角语义](sun-moon.md#观测角语义)；站心几何的完整应用见[月掩·站心事件图](occultation.md#站心事件图)与[日月位置](sun-moon.md#日月位置)；站心日食的初亏、食甚、复圆见[日食](eclipse.md#日食)。时标、观测点高度与整体精度的约定见[时标约定](timescale.md)、[观测点高度约定](coord.md#观测点高度)与[适用范围与精度](accuracy.md)。

## 目录

- [黄道坐标转赤道与地平坐标](#黄道坐标转赤道与地平坐标)
- [API 参考](#api-参考)
  - [黄道 ↔ 赤道](#黄道--赤道)
  - [赤道 ↔ 地平](#赤道--地平)
  - [站心坐标](#站心坐标)
  - [恒星时](#恒星时)
  - [岁差与章动](#岁差与章动)
  - [银道坐标](#银道坐标)
  - [角距离](#角距离)
  - [大气折射](#大气折射)
  - [大气质量](#大气质量)
  - [视差角](#视差角)
- [常用场景](#常用场景)
  - [黄道与地平坐标的往返转换](#黄道与地平坐标的往返转换)
  - [站心改正与恒星时](#站心改正与恒星时)
  - [蒙气差、大气质量与视差角](#蒙气差大气质量与视差角)
  - [岁差章动与交角口径](#岁差章动与交角口径)
- [研究型接口与观测辅助](#研究型接口与观测辅助)
- [参数与返回值约定](#参数与返回值约定)
  - [单位约定](#单位约定)
  - [时标](#时标)
  - [角度象限与归一](#角度象限与归一)
  - [零值与无效输入](#零值与无效输入)
  - [精度与适用范围](#精度与适用范围)
- [观测点高度](#观测点高度)

## 黄道坐标转赤道与地平坐标

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
	fmt.Printf("RA=%.6f Dec=%.6f deg\n", eq.RA, eq.Dec)
	fmt.Printf("azimuth=%.6f altitude=%.6f deg\n", hz.Azimuth, hz.Altitude)
	fmt.Printf("GAST=%.6f h\n", coord.ApparentSiderealTime(date))
}
```

赤经、赤纬和地平角均为度，恒星时为小时。这里的赤道坐标属于观测日期；银道转换要求 ICRS 坐标，不能直接混用。

## API 参考

下面按计算内容列出接口、单位和返回值。

片段沿用首例中的 `date`、`cst` 与 `eq`。恒星时示例还需导入标准库 `math`。

### 黄道 ↔ 赤道

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `EclipticToEquatorial` | 黄道 → 赤道 | 黄经、黄纬、赤经、赤纬均为度；交角取该时刻的瞬时真交角，`RA ∈ [0,360)` |
| `EquatorialToEcliptic` | 赤道 → 黄道 | 同一交角的反变换；`Lon ∈ [0,360)`、`Lat ∈ [−90,90]` |
| `EclipticToEquatorialByObliquity` | 黄道 → 赤道（手工交角） | 三个参数均为度；面向不同自转轴倾角的研究推演，不做日期换算 |
| `EquatorialToEclipticByObliquity` | 赤道 → 黄道（手工交角） | 同上；`Lon ∈ [0,360)`、`Lat ∈ [−90,90]` |
| `Ecliptic` | 黄道坐标结果 | 字段 `Lon`、`Lat`，单位度 |
| `Equatorial` | 赤道坐标结果 | 字段 `RA`、`Dec`，单位度 |

```go
// date 见公共前置；先取该时刻的瞬时黄赤交角，再与按日期的接口对照。
obliquity := coord.EclipticObliquity(date, true)
auto := coord.EclipticToEquatorial(date, 139.686111, 4.875278)
manual := coord.EclipticToEquatorialByObliquity(139.686111, 4.875278, obliquity)
back := coord.EquatorialToEclipticByObliquity(manual.RA, manual.Dec, obliquity)
fmt.Printf("auto=(%.9f, %.9f) manual=(%.9f, %.9f)\n",
	auto.RA, auto.Dec, manual.RA, manual.Dec)
fmt.Printf("back=(%.9f, %.9f)\n", back.Lon, back.Lat)
```

### 赤道 ↔ 地平

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `EquatorialToHorizontal` | 瞬时赤道 → 地平 | 经纬度东正西负、北正南负，单位度；用视恒星时（含章动）；`Azimuth ∈ [0,360)`、`Altitude ∈ [−90,90]`、`Zenith = 90 − Altitude` |
| `EquatorialToHorizontalByLocalSiderealTime` | 赤道 → 地平（手工恒星时） | 恒星时单位为小时，内部 ×15 转度；不查日期，适合手工推演 |
| `HourAngleDeclinationToHorizontal` | 时角 + 赤纬 → 地平 | 时角输入转 `[0,360)` 后归一输出；`Azimuth ∈ [0,360)` |
| `HorizontalToHourAngleDeclination` | 地平 → 时角 + 赤纬 | 返回时角归一到 `[0,360)`（不是 `[−180,180]`），赤纬 `[−90,90]` |
| `HorizontalToEquatorialByLocalSiderealTime` | 地平 → 赤道（手工恒星时） | 恒星时小时；`RA ∈ [0,360)` |
| `Horizontal` | 地平坐标结果 | 字段 `Azimuth`、`Altitude`、`Zenith`、`HourAngle`，均度 |

```go
// 观测点取 115°E, 40°N；eq 见公共前置。
hz := coord.EquatorialToHorizontal(date, eq.RA, eq.Dec, 115, 40)
manual := coord.EquatorialToHorizontalByLocalSiderealTime(10.5, eq.RA, eq.Dec, 40)
hz2 := coord.HourAngleDeclinationToHorizontal(hz.HourAngle, eq.Dec, 40)
ha, dec := coord.HorizontalToHourAngleDeclination(hz2.Azimuth, hz2.Altitude, 40)
eq2 := coord.HorizontalToEquatorialByLocalSiderealTime(10.5, hz2.Azimuth, hz2.Altitude, 40)
fmt.Printf("auto=(%.6f, %.6f, %.6f) manual=(%.6f, %.6f)\n",
	hz.Azimuth, hz.Altitude, hz.Zenith, manual.Azimuth, manual.Altitude)
fmt.Printf("roundtrip ha=%.6f dec=%.6f ra=%.6f\n", ha, dec, eq2.RA)
```

### 站心坐标

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `TopocentricEquatorial` | 地心赤道 → 站心赤道 | `distanceAU` 为地心距、AU；`height` 为椭球高、米；`RA` 只加视差小改正、不做 360° 归一 |
| `TopocentricEcliptic` | 地心黄道 → 站心黄道 | 同上；内部先解站心赤道坐标再转黄道，`Lon` 归一到 `[0,360)`、`Lat` 落在 `[−90,90]` |
| `Ecliptic` / `Equatorial` | 返回值类型 | 字段口径同黄道 ↔ 赤道分组 |

```go
// 目标地心距 0.00257 AU（约月球），观测点 115°E, 40°N，椭球高 53 m。
top := coord.TopocentricEquatorial(date, eq.RA, eq.Dec, 115, 40, 0.00257, 53)
topEcl := coord.TopocentricEcliptic(date, 139.686111, 4.875278, 115, 40, 0.00257, 53)
fmt.Printf("top=(%.9f, %.9f) dRA=%.9f dDec=%.9f\n",
	top.RA, top.Dec, top.RA-eq.RA, top.Dec-eq.Dec)
fmt.Printf("topEcl=(%.9f, %.9f)\n", topEcl.Lon, topEcl.Lat)
```

站心几何在月掩里用于逐地点接触时刻，见[站心事件图](occultation.md#站心事件图)；`sun`、`moon` 的站心物理量见[日月位置](sun-moon.md#日月位置)。

### 恒星时

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `MeanSiderealTime` | 格林尼治平恒星时 | 小时；先把民用时刻换成 UT1，模型为 IAU 2006（ERA 路线） |
| `ApparentSiderealTime` | 格林尼治视恒星时 | 小时；平恒星时叠加 IAU 2000B 黄经章动的投影 |
| `HourAngle` | 由瞬时赤经与站经求时角 | 经度东正西负、度；时角 `[0,360)`；基于视恒星时 |

```go
// 本地恒星时 = 格林尼治恒星时 + 东经/15，再折回 [0,24) 小时；math 仅用于这一步。
gmst := coord.MeanSiderealTime(date)
gast := coord.ApparentSiderealTime(date)
lst := math.Mod(gmst+115.0/15, 24)
fmt.Printf("GMST=%.9f h GAST=%.9f h LST=%.9f h\n", gmst, gast, lst)
fmt.Printf("HA=%.6f deg\n", coord.HourAngle(date, eq.RA, 115))
```

### 岁差与章动

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `Precess` | 赤道坐标从一个日期岁差到另一个日期 | RA/Dec 单位度；两个日期都是民用时刻；只做岁差旋转，不含自行 |
| `EclipticObliquity` | 黄赤交角 | 度；`nutation=false` 给 IAU 1980 平交角，`true` 再叠加 IAU 2000B 交角章动 |
| `Nutation2000B` | IAU 2000B 章动 | 返回 `(黄经章动, 交角章动)`，度 |
| `Nutation1980` | IAU 1980 章动 | 返回 `(黄经章动, 交角章动)`，度 |

```go
// 把 J2000 的赤道坐标岁差到 date，并列出两套章动供对照。
j2000 := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
p := coord.Precess(j2000, date, 83.6331, 22.0145)
dLon2000B, dObl2000B := coord.Nutation2000B(date)
dLon1980, dObl1980 := coord.Nutation1980(date)
fmt.Printf("precessed=(%.9f, %.9f)\n", p.RA, p.Dec)
fmt.Printf("eps mean=%.9f true=%.9f\n",
	coord.EclipticObliquity(date, false), coord.EclipticObliquity(date, true))
fmt.Printf("nut 2000B=(%.9f, %.9f) 1980=(%.9f, %.9f)\n",
	dLon2000B, dObl2000B, dLon1980, dObl1980)
```

### 银道坐标

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `EquatorialToGalactic` | ICRS 赤道 → 银道 | 输入 ICRS 度；`Lon ∈ [0,360)`、`Lat ∈ [−90,90]`；固定旋转矩阵，不做岁差与自行 |
| `GalacticToEquatorial` | 银道 → ICRS 赤道 | 同一矩阵的转置；`RA ∈ [0,360)` |
| `Galactic` | 银道坐标结果 | 字段 `Lon`、`Lat`，单位度 |

```go
// 银心方向 ICRS (266.4051, -28.936175) 应回到银经 0、银纬 0 附近。
gal := coord.EquatorialToGalactic(266.4051, -28.936175)
back := coord.GalacticToEquatorial(gal.Lon, gal.Lat)
fmt.Printf("gal=(%.6f, %.6f) back=(%.9f, %.9f)\n", gal.Lon, gal.Lat, back.RA, back.Dec)
```

### 角距离

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `AngularSeparation` | 两组赤道坐标的角距离 | 输入、输出均为度；大圆角距，与输入顺序无关 |

```go
// 蟹状星云脉冲星方向与银心方向的角距。
sep := coord.AngularSeparation(83.6331, 22.0145, 266.4051, -28.936175)
fmt.Printf("sep=%.9f deg = %.6f arcsec\n", sep, sep*3600)
```

### 大气折射

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `ApparentAltitude` | 真高度 → 视高度 | 高度角与返回值均为度；气压 hPa、气温 ℃ |
| `TrueAltitude` | 视高度 → 真高度 | 对 Saemundsson 模型做数值逆解；无解返回 NaN |
| `AtmosphericRefractionFromTrueAltitude` | 真高度处的折射量 | 度；加到真高度上得视高度 |
| `AtmosphericRefractionFromApparentAltitude` | 视高度处的折射量 | 度；从视高度减去得真高度 |
| `EquatorialToApparentHorizontal` | 赤道 → 视地平 | 在 `EquatorialToHorizontal` 结果上叠加折射，只改 `Altitude`/`Zenith`，不动 `Azimuth`/`HourAngle` |

```go
// 真高度 10°，标准气压 1010 hPa、气温 0 ℃。
apparent := coord.ApparentAltitude(10, 1010, 0)
ref := coord.AtmosphericRefractionFromTrueAltitude(10, 1010, 0)
appHz := coord.EquatorialToApparentHorizontal(date, eq.RA, eq.Dec, 115, 40, 1010, 0)
fmt.Printf("true=10 apparent=%.9f ref=%.9f\n", apparent, ref)
fmt.Printf("back=%.9f appHz alt=%.6f zen=%.6f\n",
	coord.TrueAltitude(apparent, 1010, 0), appHz.Altitude, appHz.Zenith)
```

### 大气质量

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `AirmassPlaneParallelFromTrueAltitude` | 平行平板模型 | 输入真高度、度；等价 `sec(z)`，中高空适用，近地平发散 |
| `AirmassKastenYoungFromApparentAltitude` | Kasten-Young 1989 | 输入视高度、度；不自动做折射修正 |
| `AirmassPickeringFromApparentAltitude` | Pickering 2002 | 输入视高度、度；面向低空观测 |
| `AirmassKastenYoungFromTrueAltitude` | 先折射再 Kasten-Young | 真高度、度；气压 hPa、气温 ℃ |
| `AirmassPickeringFromTrueAltitude` | 先折射再 Pickering | 同上 |

```go
// 真高度 10°，标准气压 1010 hPa、气温 0 ℃；FromTrueAltitude 内部先做折射。
apparent := coord.ApparentAltitude(10, 1010, 0)
fmt.Printf("plane=%.9f\n", coord.AirmassPlaneParallelFromTrueAltitude(10))
fmt.Printf("ky true=%.9f apparent=%.9f\n",
	coord.AirmassKastenYoungFromTrueAltitude(10, 1010, 0),
	coord.AirmassKastenYoungFromApparentAltitude(apparent))
fmt.Printf("pickering true=%.9f apparent=%.9f\n",
	coord.AirmassPickeringFromTrueAltitude(10, 1010, 0),
	coord.AirmassPickeringFromApparentAltitude(apparent))
```

只剩纯公式、不需要坐标层折射时，用 `formula` 的[大气质量模型](formula.md#大气质量模型)三模型。

### 视差角

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `ParallacticAngle` | 由瞬时赤经赤纬求视差角 | 经纬度单位度；内部走时角 + 赤纬，返回 `(−180,180]` |
| `ParallacticAngleByHourAngle` | 由时角求视差角 | 时角、赤纬、纬度单位度；返回 `(−180,180]` |

```go
// 视差角用于相机旋转、光谱缝方向和视场姿态判断。
q := coord.ParallacticAngle(date, eq.RA, eq.Dec, 115, 40)
q2 := coord.ParallacticAngleByHourAngle(coord.HourAngle(date, eq.RA, 115), eq.Dec, 40)
fmt.Printf("q=%.9f qByHA=%.9f\n", q, q2)
```

视差角的方向定义沿用[观测角语义](sun-moon.md#观测角语义)。

## 常用场景

### 黄道与地平坐标的往返转换

```go
eq := coord.EclipticToEquatorial(date, 139.686111, 4.875278)
hz := coord.EquatorialToHorizontal(date, eq.RA, eq.Dec, 115, 40)
back := coord.EquatorialToEcliptic(date, eq.RA, eq.Dec)
fmt.Println(eq.RA, eq.Dec, hz.Altitude, back.Lon)
```

```text
143.72223158223719 19.53512536790277 -17.686511328302952 139.68611100000000
```

只给时刻与坐标就够，库内部按当日视黄赤交角换算；要自己指定交角用 `EclipticToEquatorialByObliquity` / `EquatorialToEclipticByObliquity` 研究入口。往返自洽：`EquatorialToEcliptic(EclipticToEquatorial(...))` 在同口径下回到原值。

### 站心改正与恒星时

```go
top := coord.TopocentricEquatorial(date, eq.RA, eq.Dec, 115, 40, 0.00257, 53)
fmt.Println(top.RA, top.Dec)
fmt.Println(coord.MeanSiderealTime(date), coord.ApparentSiderealTime(date))
// 已有地方恒星时时可直接用它，不必让库重算：
manual := coord.EquatorialToHorizontalByLocalSiderealTime(10.5, 83.6331, 22.0145, 31.2)
fmt.Println(manual.Azimuth, manual.Altitude, manual.HourAngle)
```

```text
144.25512626944376 18.79025523004319
16.852455700085 16.852556781127
281.869347 24.489608 73.8669
```

`distanceAU` 必须给真实地心距（月球约 `0.00257 AU`）；较远天体是否可以忽略视差，取决于所需精度。`height` 是椭球高（米），恒星时单位是小时。

### 蒙气差、大气质量与视差角

```go
fmt.Println(coord.ApparentAltitude(10, 1010, 0))                   // 真高 10° -> 视高
fmt.Println(coord.TrueAltitude(10.093428, 1010, 0))                // 视高 -> 真高（逆解）
fmt.Println(coord.AirmassKastenYoungFromTrueAltitude(10, 1010, 0)) // 真高 10° 的大气质量
fmt.Println(coord.ParallacticAngle(date, eq.RA, eq.Dec, 115, 40))  // 相机/光谱缝旋转角
```

```text
10.093427592862
10.000000410649
5.537933369472
-34.000957202636
```

蒙气差只在真高度角 `(-5°, 90°)` 内生效，区间外不修正；大气质量族另有只按视高度角的简化入口（`...FromApparentAltitude`）。

### 岁差章动与交角口径

```go
fmt.Println(coord.EclipticObliquity(date, true)) // 真交角；false 给平交角
lon, obl := coord.Nutation2000B(date)            // IAU 2000B 黄经/交角章动
pre := coord.Precess(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), date, eq.RA, eq.Dec)
fmt.Println(lon, obl, pre.RA, pre.Dec)
```

```text
23.438261476424
0.001652570533 0.002392788113 144.089973153762 19.416731576422
```

`Nutation1980` 与 `Nutation2000B` 都可显式调用，便于逐项对表；`Precess` 只做岁差，不含自行、章动与光行差。手动恒星时、手动交角这类研究入口集中在[研究型接口与观测辅助](#研究型接口与观测辅助)。

## 研究型接口与观测辅助

`coord` 里的研究型接口不会自动代入当前日期的黄赤交角或恒星时，适合做“不同自转轴倾角”“手工指定时角”这类推演。常规计算可用 `EclipticToEquatorial`、`EquatorialToHorizontal` 等带 `time.Time` 的接口。

观测辅助接口包括：

- `ParallacticAngle` / `ParallacticAngleByHourAngle`：视差角（天顶方向角）
- `Airmass...FromApparentAltitude`：已经有视高度角时，直接套经验式
- `Airmass...FromTrueAltitude`：先按给定气压/气温估算折射，把真高度角换成视高度角后再算

```go
// 目标的视差角，常用于旋转相机、光谱缝方向和视场姿态判断。
q := coord.ParallacticAngle(date, eq.RA, eq.Dec, 115, 40)

// 已知真高度角时，可先估算折射，再按经验模型求大气质量。
x := coord.AirmassKastenYoungFromTrueAltitude(10, 1010, 0)
fmt.Printf("q=%.6f airmass=%.6f\n", q, x)
```

同样的观测辅助接口在 `sun`、`moon`、`star` 以及七大行星包中都有提供。已有视高度角且只需要纯公式时，`formula.Airmass...` 更直接。

## 参数与返回值约定

### 单位约定

- 角度一律为度：黄经/黄纬、赤经/赤纬、方位/高度/天顶距/时角、视差角、角距离。需要角秒时自行 ×3600（如 `AngularSeparation` 的返回值）。
- 恒星时一律为小时：`MeanSiderealTime`、`ApparentSiderealTime` 的返回值和 `*ByLocalSiderealTime` 的 `localSiderealTimeHours` 参数都是小时；内部按 ×15 换成度，不要与赤经的小时表示混用。
- `Equatorial.RA`、`Galactic.Lon`、`Horizontal.Azimuth`、`Horizontal.HourAngle` 都是**度**，不是小时；`Equatorial` 类型不区分 J2000、日期平坐标还是瞬时真坐标，由调用者保证输入口径一致。
- 距离用 AU（`TopocentricEquatorial`、`TopocentricEcliptic` 的 `distanceAU`）；观测点高度用**椭球高（大地高）**，单位米，不建模大地水准面差距，详见[观测点高度约定](coord.md#观测点高度)。
- 气压 hPa、气温 ℃；折射按标准状态 `1010 hPa`、`10 ℃` 定标。
- 本包不产生视直径/视半径；日月与日食手册里的视半径字段以角秒计，不要与本包的“度”互换。

### 时标

- 所有公开接口的 `time.Time` 按绝对时刻（民用时刻）使用，内部先 `date.UTC()` 再转儒略日；数值本身一律按 UTC 标签解释，`1972-01-01` 之前它等于 UT1。完整约定见[时标约定](timescale.md)。
- 恒星时链路显式做 UT1 换算：`MeanSiderealTime` / `ApparentSiderealTime` 走 `UTC2UT1` 后的 IAU 2006 模型；`HourAngle`、`EquatorialToHorizontal`、`TopocentricEquatorial` 内部的视恒星时同样基于 UT1。
- `Precess(from, to)` 的 `from`、`to` 都是民用时刻，按各自的绝对时刻取岁差历元。
- 折射与大气质量不涉及时刻，只依赖高度角与气象参数。

### 角度象限与归一

- 内部 `normalize360` 把所有需要展示的角归一到 `[0,360)`：赤经、黄经、银经、方位角、时角都落在这个区间。
- 赤纬、黄纬、银纬、高度角经 `Asin`（部分路径额外做 `[−1,1]` 夹取）后落 `[−90,90]`。
- `ParallacticAngle` / `ParallacticAngleByHourAngle` 用 `Atan2`，返回 `(−180,180]`，是本包唯一带符号的角。
- 站心族是例外：`TopocentricEquatorial.RA` 是输入赤经加一个小改正，不做 360° 归一，目标靠近 `0°/360°` 边界时结果可能略越界；`TopocentricEcliptic` 的 `Lon` 归一到 `[0,360)`、`Lat` 经 `Asin` 落在 `[−90,90]`（内部先解站心赤道坐标再转黄道）。
- `HourAngleDeclinationToHorizontal` 与 `HorizontalToHourAngleDeclination` 的时角都归一到 `[0,360)`，所以子午线以西的负时角在这里表现为 `360−|HA|`；要带符号时角请自行减 360。
- `EquatorialToHorizontal` 与 `HourAngleDeclinationToHorizontal` 给出的都是**几何**高度角，不含蒙气差；需要视高度用 `EquatorialToApparentHorizontal` 或 `ApparentAltitude`。

### 零值与无效输入

- 折射族要求 `pressureHPa > 0`、`temperatureC > −273.15`，任一参数为 NaN/Inf 或越界即返回 NaN。
- 真高度 `≤ −5°` 或 `≥ 90°` 时折射量按 0 处理：`ApparentAltitude` 原样返回真高度，`TrueAltitude` 与 `AtmosphericRefractionFromApparentAltitude` 原样返回输入视高度；`TrueAltitude` 只在 `(−5°, 90°)` 内做数值逆解，逆解失败返回 NaN。
- 大气质量族把高度角（或天顶距）限制在 `[0,90]`，越界或非有限返回 NaN；平行平板在高度角恰为 `0°`（天顶距恰为 `90°`）时返回 `+Inf`，Kasten-Young 与 Pickering 在 `0°` 处仍是有限值。
- `EquatorialToApparentHorizontal` 把气压/气温直接透传给折射函数：非法气象参数会让 `Altitude` 变 NaN、`Zenith` 随之为 NaN，而 `Azimuth`、`HourAngle` 仍是几何值。
- 站心族对 `distanceAU` 与 `height` 不做校验，`distanceAU` 填 `0` 或负值会得到无意义结果；必须传真实地心距（月球约 `0.00257 AU`）。
- 银道、角距离等纯几何接口不做参数校验，非有限输入会传播成 NaN，黄纬超出 `[−90,90]` 也不会被夹取，而是按球面方向重新解释。
- 四个坐标类型没有 `Valid` 字段：`Ecliptic{}`、`Equatorial{}`、`Horizontal{}`、`Galactic{}` 表示“全 0 角”而不是“未设置”，不能用零值做缺省判断。

### 精度与适用范围

- 黄赤交角：平均项取 IAU 1980 模型，`EclipticObliquity(date, true)` 在其上叠加 IAU 2000B 交角章动；`Nutation1980` 与 `Nutation2000B` 都可用，需要逐项对照时显式调用。
- 恒星时：走 IAU 2006 的 ERA 路线，视恒星时叠加 IAU 2000B 黄经章动；同一儒略日的视恒星时做了有界记忆化，注入 ΔT 覆盖会按世代失效。
- 岁差：`Precess` 在赤道坐标上做长期岁差旋转（赤道极/黄极向量含周期项），不含自行、章动与光行差；它适合把一个历元的平坐标转到另一个历元，不能代替完整的“J2000 → 当日视位置”链路。
- 银道：使用固定的 ICRS ↔ Galactic 旋转矩阵，等价于 SOFA 的 `iauIcrs2g`/`iauG2icrs`，只适用于 ICRS 口径的输入。
- 站心改正：周日视差量级随 `distanceAU` 反比变化，月球最大（约 1°），较远天体是否可忽略应按精度要求判断；`height` 只影响视差项的幅度，量级为米级高程对应的角秒级差异。
- 站心黄道坐标由站心赤道坐标换算（`basic.TopocentricLoBo` 一次求值给出黄经与黄纬）：`TopocentricEcliptic.Lat` 保证落在 `[−90,90]`，并与 `EquatorialToEcliptic(TopocentricEquatorial(...))` 这条独立路径一致。
- 折射：Saemundsson 公式，有效真高度窗口 `(−5°, 90°)`，窗口外不修正；低空（< 5°）折射变化剧烈，气压 ±1 hPa、气温 ±1 ℃ 都会带来可见误差。
- 大气质量：三种经验模型在中高空接近，低空差异最大；平行平板只是几何近似，接近地平线发散，低空精细估算用 Kasten-Young 或 Pickering，纯公式见 `formula` 的[大气质量模型](formula.md#大气质量模型)。
- 视差角：`ParallacticAngle` 用瞬时赤经赤纬与站点经纬度，`ParallacticAngleByHourAngle` 用同一几何的时角形式，两者对同一几何应给出一致结果。

## 观测点高度

站心坐标、升落、地方日月食与月掩接口的高度参数均为椭球高，单位米。正高（海拔）`H` 与椭球高的关系是 `height = H + N`，其中 `N` 为当地大地水准面差距。

库内不提供大地水准面模型，所以 `height=0` 表示参考椭球面，不一定是当地平均海面。需要这一级精度时，应从外部模型取得 `N`；与外部预报对照时也应统一高度定义。

作为量级估算，30 米高差在高度角 30° 的光线方向上对应约 52 米水平位移，即 `|Δh|·cot(高度角)`。接触时刻的差异还取决于当地影子速度和接触几何，不能视为固定秒数的修正。
