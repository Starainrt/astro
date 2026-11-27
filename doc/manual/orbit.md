# 通用小天体轨道

[English](en/orbit.md) | [返回 README](../../README.md)

`orbit` 包用于按日心轨道根数传播小行星、彗星、矮行星及自定义目标的位置、站心升落与视星等，支持经典椭圆根数、近日点形式、可选的行星数值摄动传播，以及经典视双星轨道求解。

## 目录

- [基本示例](#基本示例)
- [主要功能](#主要功能)
  - [近日点形式（彗星、抛物线与双曲线轨道）](#近日点形式彗星抛物线与双曲线轨道)
  - [行星摄动传播与多源星历接入](#行星摄动传播与多源星历接入)
  - [视双星轨道](#视双星轨道)
- [API 概览](#api-概览)
- [约定与限制](#约定与限制)

## 基本示例

使用 1 号小行星谷神星（1 Ceres）的 J2000 平黄道根数，计算指定时刻的地心视赤道坐标、H-G 视星等以及上海站心的高度角与升起时刻：

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
	lon, lat, height := 121.4737, 31.2304, 20.0 // 上海，椭球高 20 m

	ceres := orbit.Elements{
		EpochJD: 2461000.5, // TT/TDB 儒略日
		A:       2.765615651508659,
		E:       0.07957631994408416,
		I:       10.58788658206854,
		Omega:   80.24963090816965,
		W:       73.29975464616518,
		M0:      231.5397330043706,
	}

	pos := orbit.ApparentGeocentricEquatorial(when, ceres)
	mag := orbit.AsteroidMagnitudeHG(when, ceres, 3.34, 0.12) // H=3.34, G=0.12
	alt := orbit.Altitude(when, ceres, lon, lat, height)
	rise, err := orbit.RiseTime(when, ceres, lon, lat, height, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("谷神星 (2025-11-21 20:00 CST): RA=%.3f° Dec=%.3f° 距地=%.4f AU 视星等=%.2f\n",
		pos.RA, pos.Dec, pos.Distance, mag)
	fmt.Printf("上海高度 %.2f°  当日升起 %s\n", alt, rise.Format("15:04:05"))
}
```

输出：

```text
谷神星 (2025-11-21 20:00 CST): RA=7.128° Dec=-10.028° 距地=2.2625 AU 视星等=8.36
上海高度 48.47°  当日升起 14:41:49
```

根数参考系固定为 **J2000 平黄道 / 平春分点**，`EpochJD` / `TpJD` 为 TT/TDB 儒略日，角度单位为度，距离单位为 AU。与 J2000 星表比对使用 `AstrometricGeocentricEquatorialJ2000`，近地站心位置使用 `ApparentTopocentricEquatorial`。

## 主要功能

### 近日点形式（彗星、抛物线与双曲线轨道）

彗星及 `E >= 1` 的抛物线、双曲线轨道采用近日点距离 `Q`（AU）与过近日点时刻 `TpJD`（TT/TDB 儒略日）：

```go
halley := orbit.Elements{
	Q: 0.5870992, E: 0.9671429, I: 162.26269,
	Omega: 58.42008, W: 111.33249, TpJD: 2446467.395,
}
hp := orbit.ApparentGeocentricEquatorial(time.Date(1986, 2, 9, 0, 0, 0, 0, time.UTC), halley)
fmt.Printf("哈雷彗星 (1986-02-09 UTC): RA=%.2f° Dec=%.2f° 距地=%.4f AU\n", hp.RA, hp.Dec, hp.Distance)
```

输出：

```text
哈雷彗星 (1986-02-09 UTC): RA=312.11° Dec=-11.83° 距地=1.5339 AU
```

当 `Q > 0` 且 `TpJD` 有限时优先按近日点形式求解。抛物线与双曲线无平均角速度与平近点角（`MeanMotion` / `MeanAnomaly` 返回 `NaN`），`TrueAnomaly` 与全部位置接口均正常求解。

### 行星摄动传播与多源星历接入

跨度较长或近距离掠过行星时，可通过 `StateWithOptions` 启用行星引力摄动数值积分，或使用 `NewEphemerisEngine` 将其封装为引擎接入 `astro.NewStackInstance`：

| `PropagationOptions` 字段 | 说明 |
| --- | --- |
| `Perturbers` | 摄动天体列表（支持水星至海王星、地球、月球及地月质心）；非空时须同时提供星历源 |
| `Engine` / `Context` | 摄动天体星历源（如 `*spk.File` 或 `astro.Instance`），`Engine` 优先 |
| `Step` / `Tolerance` | 最大积分步长（日）与步长自适应相对误差阈值 |

`PropagationOptions` 零值严格执行二体解析传播。长期高精度任务建议直接使用 JPL SPK 文件，见[显式星历实例](ephemeris.md)。

### 视双星轨道

`VisualBinary(date, elems)` 与 `VisualBinaryByYear(decimalYear, elems)` 按经典表观轨道公式计算伴星相对主星的位置角（正北 `0°`、向东增加，度）与角距（角秒）：

```go
gammaVir := orbit.VisualBinaryElements{
	PeriodYears: 171.37, PeriastronYear: 1836.433, Eccentricity: 0.8808,
	SemiMajorAxis: 3.746, Inclination: 146.05, AscendingNode: 31.78, PeriastronArgument: 252.88,
}
vb := orbit.VisualBinaryByYear(2026.0, gammaVir)
fmt.Printf("室女座γ视双星 (2026.0): 位置角 %.2f°  角距 %.3f\"\n", vb.PositionAngle, vb.Separation)
```

输出：

```text
室女座γ视双星 (2026.0): 位置角 353.10°  角距 3.383"
```

## API 概览

| 分类 | 入口 | 说明 |
| --- | --- | --- |
| 根数与近点角 | `Elements` / `MeanMotion` / `MeanAnomaly` / `TrueAnomaly` | 经典椭圆（`A/E/.../M0/EpochJD` 及变化率 `ADot…MDot`）与近日点形式（`Q/E/.../TpJD`） |
| 位置层次 | `HeliocentricEcliptic[J2000]` | 日心几何黄道 |
| | `GeocentricEcliptic[J2000]` / `GeocentricEquatorial[J2000]` | 地心几何黄道 / 赤道 |
| | `AstrometricGeocentricEquatorialJ2000` | 地心 J2000 测算坐标（含光行时） |
| | `ApparentGeocentricEcliptic` / `ApparentGeocentricEquatorial` | 地心当日视位置（光行时 + 章动） |
| | `ApparentTopocentricEquatorial` | 站心当日视位置（再叠加站心视差） |
| 摄动传播 | `State` / `StateWithOptions` / `NewEphemerisEngine` | ICRF 几何状态与星历引擎封装 |
| 几何与测光 | `SunDistance` / `EarthDistance` / `Elongation` / `PhaseAngle` / `IlluminatedFraction`（别名 `Phase`） / `AsteroidMagnitudeHG` | 距离（AU）、日距角/相位角（度）、受照比例、H-G 星等 |
| 站心与升落 | `Altitude` / `Azimuth` / `Zenith` / `HourAngle` / `ParallacticAngle` / `CulminationTime` / `RiseTime` / `SetTime` | 极区或当日无升落返回 `ERR_ORBIT_NEVER_RISE` / `ERR_ORBIT_NEVER_SET` |
| 视双星 | `VisualBinary` / `VisualBinaryByYear` | 位置角（度）与角距（角秒） |

## 约定与限制

- **大气折射**：`Apparent*` 与 `Altitude` 均不含大气折射；仅 `RiseTime` / `SetTime` 在 `aero=true` 时计入标准地平折射与椭球高地平俯角。
- **`Phase` 与 `PhaseAngle`**：`Phase` 为 `IlluminatedFraction`（`0–1` 受照比例）的别名，太阳-天体-地球夹角应调用 `PhaseAngle`。
- **非法输入**：`Elements` 零值不构成有效轨道，非法根数在位置与近点角接口中返回 `NaN`；`StateWithOptions` 传入不支持的摄动天体或同时包含地月质心与地球/月球时返回 `ErrUnsupportedPerturber`。
- **线性变化率**：`ADot`…`WDot` 仅对经典椭圆形式生效；近日点形式下仅非零 `MDot` 覆盖平均角速度，摄动积分时若同时给出 `EpochJD` 则以其为积分起点。

相关手册：[恒星](star.md) | [坐标工具](coord.md) | [显式星历实例](ephemeris.md)
