# 研究公式

[English](en/formula.md) | [返回 README](../../README.md)

`formula` 包提供不读取星历表的无状态天文公式，涵盖望远镜指标、星等与距离换算、会合周期、黑体辐射、恒星物理参数、大气折射、消光、大气质量以及地球姿态矩阵。

## 目录

- [基本示例](#基本示例)
- [主要功能](#主要功能)
  - [大气折射、消光与大气质量](#大气折射消光与大气质量)
  - [星等、距离换算与恒星物理参数](#星等距离换算与恒星物理参数)
  - [大气质量模型](#大气质量模型)
  - [地球姿态：极移与 CPO 矩阵](#地球姿态极移与-cpo-矩阵)
- [API 概览](#api-概览)
- [约定与限制](#约定与限制)

## 基本示例

```go
package main

import (
	"fmt"

	"github.com/starainrt/astro/formula"
)

func main() {
	// 70 mm 折射镜在裸眼极限 6 等条件下的分辨极限与经验极限星等
	fmt.Printf("70mm 折射镜: Dawes分辨 %.2f\"  Rayleigh分辨 %.2f\"  极限星等 %.2f\n",
		formula.DawesLimitArcsec(70),
		formula.RayleighLimitArcsec(70),
		formula.LimitingMagnitudeEmpirical(70, 6))

	// 地球与金星的会合周期（输入输出同单位，此处为天）
	fmt.Printf("地球-金星会合周期: %.2f 天\n",
		formula.SynodicPeriod(365.25636, 224.70069))

	// 绝对星等 4.83 在 100 pc 处的视星等
	fmt.Printf("太阳放到 100 pc 视星等: %.2f\n",
		formula.ApparentMagnitudeFromAbsolute(4.83, 100))

	// 5772 K 黑体的维恩峰值波长与斯特藩-玻尔兹曼总出射度
	fmt.Printf("5772K 黑体: 峰值波长 %.1f nm  辐射出射度 %.3e W/m²\n",
		formula.WienPeakWavelength(5772)*1e9,
		formula.StefanBoltzmannFlux(5772))
}
```

输出：

```text
70mm 折射镜: Dawes分辨 1.66"  Rayleigh分辨 1.98"  极限星等 11.00
地球-金星会合周期: 583.92 天
太阳放到 100 pc 视星等: 9.83
5772K 黑体: 峰值波长 502.0 nm  辐射出射度 6.294e+07 W/m²
```

## 主要功能

### 大气折射、消光与大气质量

- 真高度与视高度互换：`ApparentAltitude(trueAlt, pressureHPa, tempC)` / `TrueAltitude(appAlt, pressureHPa, tempC)`（标准参考气象为 `1010 hPa / 10 ℃`）。
- Schaefer V 波段天顶消光：`ExtinctionSchaefer(elevM, tempC, humidityPct, latDeg, sunRADeg)`，返回的 `Total`（星等/大气质量）可直接赋值给 `phenomena.HeliacalOptions.ExtinctionCoefficient`。

```go
ext := formula.ExtinctionSchaefer(100, 20, 50, 39.9, 120)
fmt.Printf("V波段天顶消光: %.3f mag/airmass  5°高度大气质量(KY): %.2f\n",
	ext.Total, formula.AirmassKastenYoung(5))
```

输出：

```text
V波段天顶消光: 0.391 mag/airmass  5°高度大气质量(KY): 10.31
```

`ExtinctionSchaefer` 的高程参数为**测站海拔（米）**，湿度范围为 `(0, 100)`。

### 星等、距离换算与恒星物理参数

- 距离单位统一换算为秒差距：`formula.Distance(val, unit)`，支持 `DistanceParsec`、`DistanceLightYear` 与 `DistanceAU`。
- 视星等、绝对星等与距离模数：`DistanceModulus(pc)`、`ApparentMagnitudeFromAbsolute(M, pc)`、`AbsoluteMagnitudeFromApparent(m, pc)`。
- 光度、半径与有效温度互解（`L = 4πR²σT⁴`）：提供 SI 单位（米、瓦特、K）和太阳单位（`...Solar...`，R☉、L☉、K）两套接口。

### 大气质量模型

本包提供三种不自动执行折射修正的纯公式大气质量模型（高度角或天顶距限制在 `[0°, 90°]`）：

| 函数 | 输入 | 说明 |
| --- | --- | --- |
| `AirmassPlaneParallel` / `AirmassPlaneParallelByZenithDistance` | 真高度角 / 天顶距 | 几何 `sec(z)` 近似，`0°` 处返回 `+Inf` |
| `AirmassKastenYoung` | 视高度角 | Kasten-Young (1989)，低空较稳健 |
| `AirmassPickering` | 视高度角 | Pickering (2002)，面向低空观测 |

由真高度角自动叠加大气折射再求大气质量的入口见[坐标工具：大气质量](coord.md#大气质量)。

### 地球姿态：极移与 CPO 矩阵

用于接入外部 IERS EOP 参数进行高精度坐标变换：

- `TIOLocatorArcsec(jdTT)`：计算 TIO 定位角 `s′`（角秒）。
- `PolarMotionMatrix(xpArcsec, ypArcsec, sPrimeArcsec)`：按列向量将 TIRS 转换至 ITRS。
- `CelestialPoleOffsetMatrix(dXMas, dYMas)`：一阶天球中间极偏移修正矩阵（输入单位为毫角秒 `mas`，左乘现有旋转矩阵）。

本包不下载 IERS 数据，亦不将极移或 CPO 自动接入默认升落、日月食或月掩链路。

## API 概览

| 分类 | 入口 | 单位与口径 |
| --- | --- | --- |
| 黑体辐射 | `PlanckRadianceByWavelength`、`WienPeakWavelength`、`StefanBoltzmannFlux`、`SolarEffectiveTemperature` | 波长 m、温度 K → `W·sr⁻¹·m⁻³` / m / `W/m²` / K |
| 会合周期 | `SynodicPeriod(p1, p2)` | 两输入同单位，输出同单位 |
| 星等与距离 | `DistanceModulus`、`ApparentMagnitudeFromAbsolute`、`AbsoluteMagnitudeFromApparent`、`Distance` | 距离 pc、星等 mag |
| 望远镜 | `DawesLimitArcsec`、`RayleighLimitArcsec`、`LightGatheringPowerRatio`、`LimitingMagnitudeEmpirical` | 口径 mm → 分辨角为**角秒**、集光力无量纲、极限星等 mag |
| 恒星物理 | `Luminosity[Solar]FromRadiusTemperature`、`Radius[Solar]FromLuminosityTemperature`、`EffectiveTemperatureFromLuminosity[Solar]Radius` | SI（m / W / K）或太阳单位（R☉ / L☉ / K） |
| 折射与消光 | `ApparentAltitude`、`TrueAltitude`、`RefractionFromTrueAltitude`、`RefractionFromApparentAltitude`、`ExtinctionSchaefer` | 高度角与折射量为度；气压 hPa、气温 ℃ |
| 大气质量 | `AirmassPlaneParallel*`、`AirmassKastenYoung`、`AirmassPickering` | 度；相对天顶无量纲 |
| 地球姿态 | `TIOLocatorArcsec`、`PolarMotionMatrix`、`CelestialPoleOffsetMatrix` | `xp`/`yp`/`s′` 为角秒；`dX`/`dY` 为毫角秒（mas） |

## 约定与限制

- **无效输入**：本包不返回 `error`，非法或越界输入统一返回 `NaN`（使用 `math.IsNaN` 判断）；具明确物理定义的边界除外：`StefanBoltzmannFlux(0) == 0`、`SynodicPeriod(p, p) == +Inf`、`AirmassPlaneParallel(0) == +Inf`。
- **折射有效区间**：Saemundsson 经验公式仅在真高度角 `(-5°, 90°)` 内生效，区间外折射量为 `0`；逆解无解时 `TrueAltitude` 返回 `NaN`。
- **常数口径**：采用 2019 SI 定义精确常数（`h`、`c`、`k`、`σ`），IAU 标称太阳常数 `L☉ = 3.828e26 W`、`R☉ = 6.957e8 m`、`Teff = 5772 K`、`1 AU = 149597870.7 km`、`1 ly = 9460730472580.8 km`。
- **模型局限**：经验极限星等按 `裸眼极限 + 5·log10(D/7)` 计算，不含中心遮挡与光学透过率；星等距离公式不含星际消光与 K 修正。
