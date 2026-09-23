# 通用小天体轨道

[English](en/orbit.md) | [返回 README](../../README.md)

`orbit` 包用于按日心二体轨道根数传播天体位置，支持小行星、彗星、矮行星和自定义假想轨道。七大行星仍由各行星包使用内置 VSOP87 解析项计算。

`orbit.Elements` 支持两种常见写法：

- 经典椭圆根数：`A/E/I/Omega/W/M0`
- 近日点形式：`Q/E/I/Omega/W/TpJD`，适合彗星和高偏心率轨道

`orbit.Elements` 的参考系固定为 J2000 平黄道/平春分点。

站心与升落接口的经度东正西负、纬度北正南负、椭球高单位米；位置与站心量也能与[恒星](star.md#恒星)、[坐标工具](coord.md#坐标工具)手册里的接口配合使用。

## 目录

- [用轨道根数计算谷神星位置](#用轨道根数计算谷神星位置)
- [API 参考](#api-参考)
  - [轨道根数](#轨道根数)
  - [位置](#位置)
  - [几何量](#几何量)
  - [升落与中天](#升落与中天)
  - [测光](#测光)
  - [视双星](#视双星)
- [常用场景](#常用场景)
  - [按位置层次取小行星坐标](#按位置层次取小行星坐标)
  - [今晚会不会升起、现在多高](#今晚会不会升起现在多高)
  - [距日距地、相位角与 H-G 视星等](#距日距地相位角与-h-g-视星等)
  - [轨道类型与位置层次](#轨道类型与位置层次)
- [参数与返回值约定](#参数与返回值约定)
  - [位置层次](#位置层次)
  - [单位与坐标口径](#单位与坐标口径)
  - [时标与输入格式](#时标与输入格式)
  - [零值与越界](#零值与越界)
  - [精度与适用范围](#精度与适用范围)
  - [常见误用](#常见误用)
- [相关手册](#相关手册)

## 用轨道根数计算谷神星位置

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro/orbit"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	when := time.Date(2025, 11, 21, 20, 0, 0, 0, cst)
	ceres := orbit.Elements{
		EpochJD: 2461000.5, A: 2.765615651508659, E: 0.07957631994408416,
		I: 10.58788658206854, Omega: 80.24963090816965,
		W: 73.29975464616518, M0: 231.5397330043706,
	}
	pos := orbit.ApparentGeocentricEquatorial(when, ceres)
	fmt.Printf("RA=%.6f Dec=%.6f deg distance=%.6f AU\n", pos.RA, pos.Dec, pos.Distance)
	rise, err := orbit.RiseTime(when, ceres, 121.4737, 31.2304, 20, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(rise.Format(time.RFC3339))
}
```

根数使用 J2000 平黄道参考系，历元为 TT/TDB 儒略日。这里是日心二体传播；长时间跨度或近距离掠过行星时，摄动误差需要另外评估。

## API 参考

### 轨道根数

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `Elements` | 日心二体圆锥曲线根数 | `EpochJD`/`TpJD` 为 TT/TDB 儒略日；`A`/`Q` 为 AU，`I`/`Omega`/`W`/`M0` 为度，`E` 无量纲；`ADot`…`MDot` 为每天变化量，只作用于经典椭圆形式 |
| `MeanMotion` | 平均角速度 | 度/日；抛物线与双曲线返回 `NaN`；`MDot` 非零时直接取它 |
| `MeanAnomaly` | 平近点角 | 度（`[0,360)`）；抛物线与双曲线返回 `NaN` |
| `TrueAnomaly` | 真近点角 | 度（`[0,360)`）；椭圆、抛物、双曲线都有解，根数非法时返回 `NaN` |

平近点角与真近点角按同一组根数求解，`MDot` 可以替代默认平均角速度：

```go
fmt.Println(orbit.MeanMotion(ceres), orbit.MeanAnomaly(when, ceres), orbit.TrueAnomaly(when, ceres))

// 抛物线与双曲线只能走近日点形式，平均角速度与平近点角没有定义。
parabolic := orbit.Elements{Q: 0.9, E: 1, I: 30, Omega: 40, W: 50, TpJD: 2461000.5}
hyperbolic := orbit.Elements{Q: 1.2, E: 1.05, I: 30, Omega: 40, W: 50, TpJD: 2461000.5}
fmt.Println(orbit.MeanMotion(parabolic), orbit.MeanAnomaly(when, parabolic))
fmt.Println(orbit.TrueAnomaly(when, parabolic), orbit.TrueAnomaly(when, hyperbolic))
```

完整示例同时给出椭圆根数与近日点根数两条链路：

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro/orbit"
)

func main() {
	// 1 Ceres 的一组经典椭圆根数，参考系为 J2000 平黄道/平春分点。
	ceres := orbit.Elements{
		EpochJD: 2461000.5,
		A:       2.765615651508659,
		E:       0.07957631994408416,
		I:       10.58788658206854,
		Omega:   80.24963090816965,
		W:       73.29975464616518,
		M0:      231.5397330043706,
	}
	ceresPos := orbit.ApparentGeocentricEquatorial(
		time.Date(2025, 11, 12, 0, 0, 0, 0, time.UTC),
		ceres,
	)
	fmt.Printf("ceres ra=%.6f dec=%.6f distance=%.6f\n", ceresPos.RA, ceresPos.Dec, ceresPos.Distance)

	// 哈雷彗星示例：用近日点距离 Q 和近日点通过时刻 TpJD 描述。
	halley := orbit.Elements{
		Q:     0.5870992,
		E:     0.9671429,
		I:     162.26269,
		Omega: 58.42008,
		W:     111.33249,
		TpJD:  2446467.395,
	}
	halleyPos := orbit.ApparentGeocentricEquatorial(
		time.Date(1986, 2, 9, 0, 0, 0, 0, time.UTC),
		halley,
	)
	fmt.Printf("halley ra=%.6f dec=%.6f distance=%.6f\n", halleyPos.RA, halleyPos.Dec, halleyPos.Distance)
}
```

输出结果：

```text
ceres ra=7.739532 dec=-10.625981 distance=2.164391
halley ra=312.112360 dec=-11.826451 distance=1.533936
```

轨道根数本身有历元，离历元越远，静态根数误差越明显。若数据源提供 `ADot/EDot/IDot/OmegaDot/WDot/MDot` 这类长期线性变化率，也可以填入 `Elements`，用于减轻中长期漂移。

### 位置

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `EclipticPosition` | 黄道球坐标返回值 | `Lon`/`Lat` 度，`Distance` AU |
| `EquatorialPosition` | 赤道球坐标返回值 | `RA`/`Dec` 度，`Distance` AU |
| `HeliocentricEclipticJ2000` | 日心 J2000 平黄道 | 几何量，不加光行时 |
| `HeliocentricEcliptic` | 日心历元黄道 | 几何量，参考系为当日平分点 |
| `GeocentricEclipticJ2000` | 地心 J2000 平黄道 | 几何量，地球与目标同取该时刻位置 |
| `GeocentricEcliptic` | 地心历元黄道 | 几何量，参考系为当日平分点 |
| `GeocentricEquatorialJ2000` | 地心 J2000 平赤道 | 几何量，黄赤交角用 J2000 值 |
| `GeocentricEquatorial` | 地心历元平赤道 | 几何量，黄赤交角取当日值 |
| `AstrometricGeocentricEquatorialJ2000` | 地心测算 J2000 赤道 | 在几何量上加光行时，可与 J2000 星表直接比对 |
| `ApparentGeocentricEcliptic` | 地心视黄道 | 光行时 + 章动，不含完整光行差 |
| `ApparentGeocentricEquatorial` | 地心视赤道 | 光行时 + 章动，不含完整光行差 |
| `ApparentTopocentricEquatorial` | 站心视赤道 | 在视赤道上再叠加站心视差修正 |

同一时刻沿"几何 → 光行时 → 章动 → 站心"逐层加码：

```go
h := orbit.HeliocentricEcliptic(when, ceres)                  // 日心历元黄道，几何量
g := orbit.GeocentricEquatorialJ2000(when, ceres)             // 地心 J2000 平赤道
a := orbit.ApparentGeocentricEquatorial(when, ceres)          // 地心视赤道
t := orbit.ApparentTopocentricEquatorial(when, ceres, 121.4737, 31.2304, 20)
e := orbit.ApparentGeocentricEcliptic(when, ceres)
fmt.Printf("h=%.6f %.6f %.6f\n", h.Lon, h.Lat, h.Distance)
fmt.Printf("g=%.6f %.6f %.6f\n", g.RA, g.Dec, g.Distance)
fmt.Printf("a=%.6f %.6f t=%.6f %.6f e=%.6f\n", a.RA, a.Dec, t.RA, t.Dec, e.Lon)
```

`...J2000` 与不带后缀的历元量是两套参考系：前者固定到 J2000 平黄道/平春分点，适合与星表比对和长期存档；后者使用当日平黄道/平春分点，适合表达“当天天空”。地心量里的 `Distance` 在几何接口上是该时刻的瞬时距离，在 `Astrometric...` 上是光行时收敛后的距离，两者相差约光行时对应的位移。

### 几何量

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `SunDistance` | 日心距离 | AU，几何量 |
| `EarthDistance` | 地心距离 | AU，几何量 |
| `Elongation` | 日距角 | 度，地心视方向上的角距 |
| `PhaseAngle` | 相位角 | 度，0° 为全亮面朝向观测者 |
| `IlluminatedFraction` | 被照亮比例 | 无量纲，通常落在 `[0,1]` |
| `Phase` | 被照亮比例别名 | 与 `IlluminatedFraction` 同义 |
| `ParallacticAngle` | 视差角（天顶方向角） | 度，时角与赤纬取自同一次站心求解 |

`orbit` 也提供了常见观测几何量和轻量测光接口：

```go
r := orbit.SunDistance(when, ceres)                         // 日心距离
delta := orbit.EarthDistance(when, ceres)                   // 地心距离
elong := orbit.Elongation(when, ceres)                      // 日距角
phase := orbit.PhaseAngle(when, ceres)                      // 相位角
k := orbit.IlluminatedFraction(when, ceres)                 // 被照亮比例
mag := orbit.AsteroidMagnitudeHG(when, ceres, 3.34, 0.12)   // H-G 视星等
q := orbit.ParallacticAngle(when, ceres, 121.4737, 31.2304, 20) // 站心视差角

fmt.Printf("r=%.6f delta=%.6f elong=%.6f phase=%.6f k=%.6f mag=%.3f q=%.6f\n",
	r, delta, elong, phase, k, mag, q)
```

日距角接近 180° 时相位角接近 0°，`IlluminatedFraction` 接近 1；这三个量都由同一组地心几何推出。

`ParallacticAngle` 不另求赤纬，它和 `HourAngle` 复用同一次站心求解，避免时角与赤纬来自两条相差约 1 ULP 的儒略日路径而引入亚纳度漂移。几何量都只依赖 `date` 的绝对时刻，不含观测者参数；含站心参数的只有 `ParallacticAngle` 一个。

### 升落与中天

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `Altitude` | 视高度角 | 度，站心视位置与观测者当地民用时刻 |
| `Zenith` | 天顶距 | 度，等于 `90 - Altitude` |
| `Azimuth` | 视方位角 | 度，正北 0°、向东增加 |
| `HourAngle` | 站心视时角 | 度 |
| `CulminationTime` | 中天时刻 | `time.Time`，保持输入 `date` 的时区 |
| `RiseTime` | 升起时刻 | `(time.Time, error)`，第二返回值为哨兵错误 |
| `SetTime` | 落下时刻 | `(time.Time, error)`，第二返回值为哨兵错误 |
| `ERR_ORBIT_NEVER_RISE` | 目标当日永不升起的哨兵错误 | 由 `RiseTime` 返回 |
| `ERR_ORBIT_NEVER_SET` | 目标当日永不落下的哨兵错误 | 由 `SetTime` 返回 |

```go
fmt.Println(orbit.Zenith(when, ceres, 121.4737, 31.2304, 20))
fmt.Println(orbit.HourAngle(when, ceres, 121.4737, 31.2304, 20))
fmt.Println(orbit.CulminationTime(when, ceres, 121.4737, 31.2304, 20).Format(time.RFC3339))

day := time.Date(2025, 11, 21, 0, 0, 0, 0, site)
set, err := orbit.SetTime(day, ceres, 121.4737, 31.2304, 20, true)
if errors.Is(err, orbit.ERR_ORBIT_NEVER_SET) {
	fmt.Println("当日不落", set)
}
```

已有轨道根数时，也可以把它当作一个“可观测目标”来求站心观测量：

```go
site := time.FixedZone("CST", 8*3600)
when := time.Date(2025, 11, 21, 20, 0, 0, 0, site)

alt := orbit.Altitude(when, ceres, 121.4737, 31.2304, 20) // 视高度角
az := orbit.Azimuth(when, ceres, 121.4737, 31.2304, 20)   // 视方位角
rise, _ := orbit.RiseTime(time.Date(2025, 11, 21, 0, 0, 0, 0, site), ceres, 121.4737, 31.2304, 20, true) // 升起时刻

fmt.Printf("alt=%.6f az=%.6f rise=%s\n", alt, az, rise.Format(time.RFC3339))
```

这些观测接口基于站心视坐标计算，适合直接拿去做小行星、彗星或自定义二体目标的升落和指向辅助。

`aero` 为假时判据是几何地平线，为真时把目标高度取到 `-0.5667°` 并叠加依椭球高、纬度算出的地平俯角。`RiseTime`/`SetTime` 的第二个返回值是真正的错误：只有当日确实没有升/落才映射成 `ERR_ORBIT_NEVER_RISE` / `ERR_ORBIT_NEVER_SET`，其余失败原样透传。

### 测光

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `AsteroidMagnitudeHG` | 小行星 H-G 模型视星等 | `absoluteMagnitude` 为绝对星等 H，`slopeParameter` 为斜率参数 G；无单位 |

```go
fmt.Printf("H-G magnitude=%.3f\n", orbit.AsteroidMagnitudeHG(when, ceres, 3.34, 0.12))
fmt.Printf("r=%.6f delta=%.6f elong=%.6f\n",
	orbit.SunDistance(when, ceres), orbit.EarthDistance(when, ceres), orbit.Elongation(when, ceres))
fmt.Printf("phase=%.6f k=%.6f k2=%.6f\n",
	orbit.PhaseAngle(when, ceres), orbit.IlluminatedFraction(when, ceres), orbit.Phase(when, ceres))
```

H-G 模型只用日心距、地心距和相位角，不引入目标的半径、反照率或自转；`G` 的取值由外部星表给出，本包不做默认值填充。

### 视双星

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `VisualBinaryElements` | 视双星轨道要素 | `PeriodYears` 平太阳年，`PeriastronYear` 带小数的年，`SemiMajorAxis` 角秒，`Inclination`/`AscendingNode`/`PeriastronArgument` 度，`Eccentricity` 无量纲 |
| `VisualBinaryPosition` | 视双星计算结果 | `MeanAnomaly`/`EccentricAnomaly`/`TrueAnomaly`/`PositionAngle` 度，`Radius`/`Separation` 角秒 |
| `VisualBinary` | 按时刻求视双星位置 | 先把时刻换算为 UTC 小数年，再套经典视轨道公式 |
| `VisualBinaryByYear` | 按小数年求视双星位置 | 直接给小数年，跳过时刻换算 |

```go
gammaVir := orbit.VisualBinaryElements{
	PeriodYears: 171.37, PeriastronYear: 1836.433, Eccentricity: 0.8808,
	SemiMajorAxis: 3.746, Inclination: 146.05, AscendingNode: 31.78, PeriastronArgument: 252.88,
}
vb := orbit.VisualBinaryByYear(2026.0, gammaVir)
fmt.Printf("theta=%.6f rho=%.6f M=%.6f\n", vb.PositionAngle, vb.Separation, vb.MeanAnomaly)
```

`orbit` 里还带了一个视双星求解器，直接按《天文算法》第 55 章的经典表观轨道公式输出位置角和角距：

```go
gammaVir := orbit.VisualBinaryElements{
	PeriodYears:        171.37,
	PeriastronYear:     1836.433,
	Eccentricity:       0.8808,
	SemiMajorAxis:      3.746,
	Inclination:        146.05,
	AscendingNode:      31.78,
	PeriastronArgument: 252.88,
}
vb := orbit.VisualBinary(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), gammaVir)
fmt.Printf("theta=%.6f rho=%.6f\n", vb.PositionAngle, vb.Separation) // 位置角与角距
```

位置角按北为 0°、东为 90° 度量，角距离与径矢 `Radius` 的单位都是角秒。

## 常用场景

### 按位置层次取小行星坐标

```go
helio := orbit.HeliocentricEcliptic(when, ceres)
geo := orbit.GeocentricEclipticJ2000(when, ceres)
ast := orbit.AstrometricGeocentricEquatorialJ2000(when, ceres)
app := orbit.ApparentGeocentricEquatorial(when, ceres)
fmt.Println(helio.Lon, helio.Lat, helio.Distance)
fmt.Println(geo.Lon, geo.Lat, geo.Distance)
fmt.Println(ast.RA, ast.Dec)
fmt.Println(app.RA, app.Dec, app.Distance)
```

```text
19.251489 -9.315340 2.912174
2.147652 -12.026350 2.262445
6.795211 -10.172420
7.125008 -10.028726 2.262489
```

四个层次含义不同：`Heliocentric*` 是相对太阳的位置（第一个数就是日心距），`Geocentric*J2000` 是 J2000 口径的地心位置，`Astrometric*J2000` 去掉光行时、适合与星表对表，`Apparent*` 是当日视位置、观测与出图用它。需要站心视位置用 `ApparentTopocentricEquatorial(when, ceres, lon, lat, height)`。

### 今晚会不会升起、现在多高

```go
fmt.Println(orbit.RiseTime(when, ceres, 121.4737, 31.2304, 20, true))
fmt.Println(orbit.CulminationTime(when, ceres, 121.4737, 31.2304, 20))
fmt.Println(orbit.Altitude(when, ceres, 121.4737, 31.2304, 20),
	orbit.Azimuth(when, ceres, 121.4737, 31.2304, 20))
```

```text
2025-11-21 14:41:48.913 CST <nil>
2025-11-21 20:19:34 CST
48.472628 172.699168
```

- `aero = true` 用蒙气差与视半径修正后的地平；`height` 是椭球高（米）；经度东正西负。
- 环极或极区目标没有升落：`RiseTime`/`SetTime` 返回 `orbit.ERR_ORBIT_NEVER_RISE` / `ERR_ORBIT_NEVER_SET` 哨兵错误，用 `errors.Is` 判定。

### 距日距地、相位角与 H-G 视星等

```go
fmt.Println(orbit.SunDistance(when, ceres), orbit.EarthDistance(when, ceres))
fmt.Println(orbit.Elongation(when, ceres), orbit.PhaseAngle(when, ceres))
fmt.Println(orbit.IlluminatedFraction(when, ceres), orbit.AsteroidMagnitudeHG(when, ceres, 3.34, 0.12))
```

```text
2.912174 2.262445
122.264630 16.670089
0.978986 8.360
```

- `PhaseAngle` 是太阳-天体-地球夹角（度），`IlluminatedFraction` 是照明比例；`Phase` 只是 `IlluminatedFraction` 的别名，别把它当成相位角。
- H-G 视星等要传绝对星等 `H` 与斜率参数 `G`（示例用谷神星的 `3.34` 与 `0.12`）。

### 轨道类型与位置层次

```go
parabolic := orbit.Elements{Q: 0.9, E: 1, I: 30, Omega: 40, W: 50, TpJD: 2461000.5}
hyperbolic := orbit.Elements{Q: 1.2, E: 1.05, I: 30, Omega: 40, W: 50, TpJD: 2461000.5}
fmt.Println(orbit.MeanMotion(parabolic), orbit.MeanAnomaly(when, parabolic))
fmt.Println(orbit.TrueAnomaly(when, parabolic), orbit.TrueAnomaly(when, hyperbolic))
fmt.Println(orbit.MeanMotion(ceres))
```

```text
NaN NaN
0.817533 0.537610
0.21429712142765137
```

- 抛物线与双曲线只能走近日点形式（`Q` + `TpJD`）：平均角速度与平近点角没有定义、返回 `NaN`，真近点角三种轨道都有解。
- 与外部星历对表时先统一位置层次与参考系（`Elements` 固定为 J2000 平黄道/平春分点）；`ADot…WDot` 只作用于经典椭圆形式，`MDot` 非零时可直接替代默认平均角速度。

## 参数与返回值约定

### 位置层次

| 层次 | 代表接口 | 包含的改正 |
| --- | --- | --- |
| 日心几何 | `HeliocentricEcliptic` / `HeliocentricEclipticJ2000` | 无 |
| 地心几何 | `GeocentricEcliptic` / `GeocentricEclipticJ2000` / `GeocentricEquatorial` / `GeocentricEquatorialJ2000` | 减去地球日心位置 |
| 地心测算 | `AstrometricGeocentricEquatorialJ2000` | 光行时 |
| 地心视 | `ApparentGeocentricEcliptic` / `ApparentGeocentricEquatorial` | 光行时 + 章动 |
| 站心视 | `ApparentTopocentricEquatorial` 及全部升落接口 | 光行时 + 章动 + 站心视差 |

### 单位与坐标口径

- 角度一律用度，距离用 AU，站心与升落接口的椭球高用米，时间用 `time.Time`。
- `Elements` 的参考系是 J2000 平黄道/平春分点；`...J2000` 结尾的接口保持该参考系，其余 `HeliocentricEcliptic`/`GeocentricEcliptic`/`GeocentricEquatorial` 是历元（of date）量，换参考系时不要混用。
- 位置分三层：`Heliocentric*`/`Geocentric*` 是几何量，`AstrometricGeocentricEquatorialJ2000` 在几何量上加光行时（按距离迭代求解，上限 8 次、`1e-12` 天收敛），`Apparent*` 在光行时之上再加章动。仓库的行星口径就是“光行时 + 章动、不含完整外部光行差模型”，所以 `Apparent*` 与行星包同级，不应把它当成全项视位置。
- `MeanMotion`/`MeanAnomaly`/`TrueAnomaly` 输出度；`MeanMotion` 与 `MeanAnomaly` 对抛物线和双曲线没有定义。

### 时标与输入格式

- `EpochJD` 与 `TpJD` 都是 TT/TDB 儒略日；坐标、几何量与测光接口把 `date` 当绝对时刻处理（`date.UTC()` 后经 `UTC2TT` 换成 TT/TDB）。`UTC2TT` 在 1972-01-01 之前把民用时刻按 UT1 处理，窗口内使用内置闰秒表，闰秒表可用 `astro.SetTTMinusUTC` 覆盖。
- 瞬时站心量将 `date` 的当地字段与 `date.Zone()` 偏移一起换回绝对时刻。同一时刻用 UTC 或当地时区表示，所得位置相同。升落搜索还按当地日期选事件，因此应选择观测日历所用的时区。
- `CulminationTime`、`RiseTime`、`SetTime` 的结果保持输入 `date` 的时区，是民用时刻；三者内部在 `date.Hour() > 12` 时先回退 12 小时，以保持搜索锚点在所选日期内。
- 公开 API 的含 `time.Time` 输出一律是民用时刻（UTC 标签）；UT1 与 TT 只出现在内部换算里，需要显式换算时用根包的 `astro.UT1FromUTC` / `astro.TTFromUTC`。

### 零值与越界

- `Elements` 零值不是合法轨道。经典椭圆形式要求 `A` 有限且为正、`E` 落在 `[0,1)`、`EpochJD` 与 `M0` 有限；`E >= 1` 的抛物线与双曲线只能走近日点形式，此时要求 `Q` 有限且为正、`TpJD` 有限、`E >= 0`，三个角度 `I`/`Omega`/`W` 都必须有限。
- `Q > 0` 且 `TpJD` 有限时优先按近日点形式解释，`A`/`M0`/`EpochJD` 被忽略；此时 `ADot`/`EDot`/`IDot`/`OmegaDot`/`WDot` 都不生效，只有 `MDot` 非零时会被当作平均角速度使用。
- 根数非法时 `MeanMotion`、`MeanAnomaly` 返回 `NaN`，位置接口返回三个 `NaN`；`AsteroidMagnitudeHG` 在输入非有限或日心距、地心距、相位角非正时返回 `NaN`，H-G 相位混合项为零时返回 `+Inf`。
- `RiseTime` 与 `SetTime` 在给定当地日内找不到升/落时返回哨兵错误 `ERR_ORBIT_NEVER_RISE` / `ERR_ORBIT_NEVER_SET`，其它失败原样透传；成功时第二个返回值为 `nil`。
- `VisualBinary` 与 `VisualBinaryByYear` 在 `PeriodYears <= 0`、`SemiMajorAxis <= 0`、`Eccentricity` 不落在 `[0,1)` 或任一要素非有限时，把 `VisualBinaryPosition` 的全部数值字段填成 `NaN`。

### 精度与适用范围

- `orbit` 是二体圆锥曲线传播：只含所给根数，不含行星摄动、非引力项与相对论改正；离历元越远，静态根数的误差越大，`ADot`…`WDot` 只能缓解线性漂移，不能替代重新拟合或数值积分。
- `Apparent*` 与站心量的差别只在几何与章动，不含大气折射；升落接口的 `aero=true` 才把地平折射算进判据（目标高度取 `-0.5667°`，再叠加依椭球高与纬度算出的地平俯角）。
- `RiseTime`/`SetTime` 内部用 `round(observerLon/15)` 的标称时区迭代升落几何，因此观测点经度最好落在时区中心附近；只调用一次求根，极区或拱极目标的边界情况以哨兵错误为准。
- 视双星求解器用的是《天文算法》第 55 章的经典视轨道公式，`Eccentricity >= 1` 时不适用；它只做几何投影，不含质量、光度或摄动信息。

### 常见误用

- 把 `HeliocentricEcliptic` 的返回值当作地心坐标：日心量的原点是太阳，地心量已经减掉了地球的日心位置。
- 把 `Apparent*` 当作含大气折射的视位置：折射只出现在升落判据与观测类接口里，坐标接口不含折射。
- 用零值 `Elements` 探路：`MeanMotion` 与位置接口会安静地返回 `NaN`，不会报错。
- 在近日点形式里期待 `ADot`…`WDot` 生效：这些变化率只在经典椭圆形式下参与传播。

## 相关手册

- 恒星与星表：[恒星](star.md#恒星)
- 坐标系换算与站心量：[坐标工具](coord.md#坐标工具)
- 日出日落、月出月落的同类接口：[太阳与月亮](sun-moon.md#日出日落月出月落)

> 需要把这里的黄道/赤道坐标接到恒星表或站心量上时，参见[恒星](star.md#恒星)与[坐标工具](coord.md#坐标工具)。
