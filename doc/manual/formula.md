# 研究公式

[English](en/formula.md) | [返回 README](../../README.md)

`formula` 包放的是和具体日期、星历表无关的常用公式，适合科普估算、小说设定和教学演示。

它不做任何时标换算，也不读星历表：输入只有温度、波长、距离、口径、高度角这类瞬时或常量参数。高度角与天顶距的口径沿用[观测角语义](sun-moon.md#观测角语义)；需要坐标层的折射修正时改用 `coord` 的[大气质量](coord.md#大气质量)；整体精度与适用范围见[适用范围与精度](accuracy.md)。

## 目录

- [星等、会合周期与黑体辐射](#星等会合周期与黑体辐射)
- [API 参考](#api-参考)
  - [黑体与辐射](#黑体与辐射)
  - [会合周期](#会合周期)
  - [星等与距离](#星等与距离)
  - [距离单位换算](#距离单位换算)
  - [望远镜指标](#望远镜指标)
  - [恒星参数换算](#恒星参数换算)
  - [大气质量模型](#大气质量模型)
- [常用场景](#常用场景)
  - [黑体峰值、总辐射与恒星参数](#黑体峰值总辐射与恒星参数)
  - [会合周期与星等距离](#会合周期与星等距离)
  - [望远镜极限星等与分辨率](#望远镜极限星等与分辨率)
  - [大气质量模型对比](#大气质量模型对比)
- [参数与返回值约定](#参数与返回值约定)
  - [单位约定](#单位约定)
  - [时标](#时标)
  - [角度象限与弧度/度的边界](#角度象限与弧度度的边界)
  - [零值与无效输入](#零值与无效输入)
  - [精度与适用范围](#精度与适用范围)

## 星等、会合周期与黑体辐射

```go
package main

import (
	"fmt"

	"github.com/starainrt/astro/formula"
)

func main() {
	// 70mm 小折射镜，观测地裸眼极限取 6 等。
	fmt.Printf("limiting=%.6f\n", formula.LimitingMagnitudeEmpirical(70, 6))

	// 地球和金星的会合周期，输入周期单位都是天，输出也是天。
	fmt.Printf("synodic=%.6f\n", formula.SynodicPeriod(365.25636, 224.70069))

	// 太阳这样的绝对星等天体放到 100pc 处的视星等。
	fmt.Printf("apparent=%.6f\n", formula.ApparentMagnitudeFromAbsolute(4.83, 100))

	// 把太阳近似为 5772K 黑体，计算峰值波长和单位面积总辐射出射度。
	fmt.Printf("peak=%.9em flux=%.6e\n",
		formula.WienPeakWavelength(5772),
		formula.StefanBoltzmannFlux(5772),
	)
}
```

输出结果：

```text
limiting=11.000000
synodic=583.920635
apparent=9.830000
peak=5.020394932e-07m flux=6.293859e+07
```

## API 参考

下面按计算内容列出接口、单位和返回值。

分组片段省略公共前置：`fmt` 已在文件头导入，`formula` 指 `github.com/starainrt/astro/formula`。

### 黑体与辐射

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `PlanckRadianceByWavelength` | 按波长的普朗克谱辐亮度 | 波长米、温度 K；返回 W·sr⁻¹·m⁻³；非正温度或非正波长返回 NaN |
| `WienPeakWavelength` | 维恩位移峰值波长 | 温度 K；返回米；温度 `≤ 0` 或非有限返回 NaN |
| `StefanBoltzmannFlux` | 单位面积总辐射出射度 | 温度 K；返回 W/m²；温度 `0 K` 合法且返回 `0` |
| `SolarEffectiveTemperature` | 内置太阳有效温度常数 | 无参数；返回 K，当前为 `5772` |

```go
// 把太阳近似为 5772 K 黑体：峰值波长、总出射度，以及 500 nm 处的谱辐亮度。
tSun := formula.SolarEffectiveTemperature()
fmt.Printf("peak=%.9e m flux=%.6e W/m^2\n",
	formula.WienPeakWavelength(tSun), formula.StefanBoltzmannFlux(tSun))
fmt.Printf("radiance@500nm=%.6e W·sr^-1·m^-3\n",
	formula.PlanckRadianceByWavelength(500e-9, tSun))
```

### 会合周期

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `SynodicPeriod` | 两个周期天体的会合周期 | 两个输入单位必须一致，输出同单位；周期 `≤ 0` 或非有限返回 NaN，两周期相等返回 `+Inf` |

```go
// 地球与其他行星的会合周期，输入输出都是天。
earth := 365.25636
fmt.Printf("venus=%.6f mars=%.6f jupiter=%.6f\n",
	formula.SynodicPeriod(earth, 224.70069),
	formula.SynodicPeriod(earth, 686.980),
	formula.SynodicPeriod(earth, 4332.589))
```

### 星等与距离

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `DistanceModulus` | 距离模数 | 距离 pc；返回 `m − M`；`10 pc` 处为 `0`，距离 `≤ 0` 返回 NaN |
| `ApparentMagnitudeFromAbsolute` | 绝对星等 + 距离 → 视星等 | 星等 mag、距离 pc；等于 `M + 距离模数` |
| `AbsoluteMagnitudeFromApparent` | 视星等 + 距离 → 绝对星等 | 星等 mag、距离 pc；等于 `m − 距离模数` |

```go
// 太阳绝对星等 4.83，放到 10 pc / 100 pc / 1 kpc 处的视星等与反解。
for _, d := range []float64{10, 100, 1000} {
	m := formula.ApparentMagnitudeFromAbsolute(4.83, d)
	fmt.Printf("d=%.0f pc m=%.6f M=%.6f mod=%.6f\n",
		d, m, formula.AbsoluteMagnitudeFromApparent(m, d), formula.DistanceModulus(d))
}
```

### 距离单位换算

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `Distance` | 把秒差距、光年或天文单位换算为秒差距 | 输入正数 + `DistanceUnit`；非正数、NaN、未知单位返回 NaN |

| 常量 | 含义 |
| --- | --- |
| `DistanceParsec` | 秒差距 pc，恒等换算 |
| `DistanceLightYear` | 光年 ly |
| `DistanceAU` | 天文单位 AU |

```go
// 天狼星视差 0.375 角秒，折合 2.667 pc，也就是 8.70 光年。
pc := formula.Distance(1/0.375, formula.DistanceParsec)
fmt.Printf("%.3f pc = %.2f ly\n", pc, formula.Distance(pc, formula.DistanceParsec)/formula.Distance(1, formula.DistanceLightYear))
```

口径：天文单位取 `149597870.7` km，光年取 IAU 定义值 `9460730472580.8` km，秒差距由精确关系 `648000/π` 天文单位导出，因此 `1 pc = 3.261563777 ly`。

### 望远镜指标

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `DawesLimitArcsec` | Dawes 极限分辨角 | 口径 mm；返回角秒，经验式 `116 / D` |
| `RayleighLimitArcsec` | Rayleigh 极限分辨角 | 口径 mm；返回角秒，经验式 `138.4 / D` |
| `LightGatheringPowerRatio` | 集光力比值 | 两个口径 mm；返回 `(D1 / D2)²`，无量纲 |
| `LimitingMagnitudeEmpirical` | 经验极限星等 | 口径 mm、裸眼极限 mag；按 `裸眼极限 + 5·log10(D / 7)` 估算，7 mm 为内置暗适应瞳径 |

```go
// 70 mm 小折射镜：分辨极限、相对 7 mm 暗瞳的集光力、裸眼极限 6 等的经验极限星等。
fmt.Printf("dawes=%.6f rayleigh=%.6f\n",
	formula.DawesLimitArcsec(70), formula.RayleighLimitArcsec(70))
fmt.Printf("power=%.6f limiting=%.6f\n",
	formula.LightGatheringPowerRatio(70, 7),
	formula.LimitingMagnitudeEmpirical(70, 6))
```

### 恒星参数换算

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `LuminosityFromRadiusTemperature` | 半径 + 温度 → 光度 | 半径米、温度 K；返回 W；按 `4πR²σT⁴` |
| `LuminositySolarFromRadiusTemperature` | 同上，太阳单位 | 半径 R☉、温度 K；返回 L☉ |
| `RadiusFromLuminosityTemperature` | 光度 + 温度 → 半径 | 光度 W、温度 K；返回米 |
| `RadiusSolarFromLuminosityTemperature` | 同上，太阳单位 | 光度 L☉、温度 K；返回 R☉ |
| `EffectiveTemperatureFromLuminosityRadius` | 光度 + 半径 → 有效温度 | 光度 W、半径米；返回 K |
| `EffectiveTemperatureFromLuminositySolarRadius` | 同上，太阳单位 | 光度 L☉、半径 R☉；返回 K |
| `SolarEffectiveTemperature` | 内置太阳有效温度 | 无参数；返回 K |

```go
// 半径 2.5 R☉、光度 20 L☉ 的主序星：反解温度，再正算回光度与半径。
t := formula.EffectiveTemperatureFromLuminositySolarRadius(20, 2.5)
fmt.Printf("Teff=%.6f K\n", t)
fmt.Printf("L=%.6f Lsun R=%.6f Rsun\n",
	formula.LuminositySolarFromRadiusTemperature(2.5, t),
	formula.RadiusSolarFromLuminosityTemperature(20, t))
// 同一组量的 MKS 版本；太阳半径取内置常数的 6.957e8 m。
rM := 2.5 * 6.957e8
lW := formula.LuminosityFromRadiusTemperature(rM, t)
fmt.Printf("L=%.6e W R=%.6e m Teff=%.6f K\n",
	lW, formula.RadiusFromLuminosityTemperature(lW, t),
	formula.EffectiveTemperatureFromLuminosityRadius(lW, rM))
```

### 大气质量模型

| 名称 | 用途 | 单位与口径 |
| --- | --- | --- |
| `AirmassPlaneParallel` | 平行平板模型 | 输入真高度角、度；等价 `sec(z)`，`0°` 处返回 `+Inf` |
| `AirmassPlaneParallelByZenithDistance` | 平行平板模型（天顶距） | 输入天顶距、度；`90°` 处返回 `+Inf` |
| `AirmassKastenYoung` | Kasten-Young 1989 | 输入视高度角、度；低空比 `sec(z)` 稳健 |
| `AirmassPickering` | Pickering 2002 | 输入视高度角、度；面向低空观测修正 |

四者都把输入限制在 `[0,90]`，越界或非有限返回 NaN；高度角以地平 `0°`、天顶 `+90°` 计，天顶距与高度角互补，口径见[观测角语义](sun-moon.md#观测角语义)。

```go
fmt.Println(formula.AirmassPlaneParallel(30))
fmt.Println(formula.AirmassKastenYoung(5))
fmt.Println(formula.AirmassPickering(5))
fmt.Println(formula.AirmassPlaneParallelByZenithDistance(60))
```

如果不需要坐标层的折射修正，`formula` 也直接提供三种大气质量模型，输入语义更直接：

- `AirmassPlaneParallel`：输入真高度角，等价于 `sec(z)` 几何近似
- `AirmassPlaneParallelByZenithDistance`：直接输入天顶距
- `AirmassKastenYoung` / `AirmassPickering`：输入视高度角，不会自动做折射修正

## 常用场景

### 黑体峰值、总辐射与恒星参数

```go
fmt.Println(formula.WienPeakWavelength(5772))                            // 峰值波长（米）
fmt.Println(formula.StefanBoltzmannFlux(5772))                           // 单位面积总辐射（W/m²）
fmt.Println(formula.RadiusSolarFromLuminosityTemperature(1, 5772))       // 由光度与温度解半径
fmt.Println(formula.EffectiveTemperatureFromLuminositySolarRadius(1, 1)) // 由光度与半径解温度
```

```text
5.020394932432432e-07
6.293859246828887e+07
1.0000011882005775
5772.003429145848
```

三组互算函数互为逆运算，同口径下回到输入值（示例里 `1 → 1.0000012`、`5772 → 5772.0034` 的残差来自双方都取 5772 K 的近似）；带 `...Solar` 的变体用太阳单位，不带的使用 SI。

### 会合周期与星等距离

```go
fmt.Println(formula.SynodicPeriod(365.25636, 224.70069))      // 地球与金星的会合周期（天）
fmt.Println(formula.DistanceModulus(10))                      // 10 pc 处的距离模数
fmt.Println(formula.ApparentMagnitudeFromAbsolute(4.83, 100)) // 绝对星等 4.83 放到 100 pc
fmt.Println(formula.AbsoluteMagnitudeFromApparent(4.83, 100)) // 逆换算
```

```text
583.9206352820089
0
9.83
-0.16999999999999993
```

`DistanceModulus(10)` 为 0，因为 10 pc 就是绝对星等的定义距离；会合周期的输入输出都是天，参数顺序不影响结果。

### 望远镜极限星等与分辨率

```go
fmt.Println(formula.DawesLimitArcsec(70), formula.RayleighLimitArcsec(70)) // 70 mm 口径的两种分辨极限
fmt.Println(formula.LightGatheringPowerRatio(200, 70))                     // 200 mm 相对 70 mm 的集光力
fmt.Println(formula.LimitingMagnitudeEmpirical(70, 6))                     // 裸眼 6 等时 70 mm 的极限星等
```

```text
1.6571428571428573 1.9771428571428573
8.16326530612245
11
```

Dawes 与 Rayleigh 相差一个系数（70 mm 下 1.66″ 与 1.98″），报告里要写明用的是哪一条；`LimitingMagnitudeEmpirical` 的第二个参数是观测地裸眼极限星等，换观测地要一起改。

### 大气质量模型对比

```go
for _, alt := range []float64{5, 30, 60, 90} {
	fmt.Printf("alt=%.0f KY=%.6f Pickering=%.6f plane=%.6f\n", alt,
		formula.AirmassKastenYoung(alt), formula.AirmassPickering(alt), formula.AirmassPlaneParallel(alt))
}
```

```text
alt=5 KY=10.305791 Pickering=10.333706 plane=11.473713
alt=30 KY=1.994293 Pickering=1.993154 plane=2.000000
alt=60 KY=1.153992 Pickering=1.154058 plane=1.154701
alt=90 KY=0.999712 Pickering=1.000000 plane=1.000000
```

三种模型在中高空几乎重合，低空（5°）差异最大（Kasten-Young 与平面平行差约 1.2 个大气质量）；平面平行模型在 0° 发散，低空精细估算用 Kasten-Young 或 Pickering。带气压与气温修正的版本在 [坐标工具](coord.md#大气质量)，本包只给纯公式。

## 参数与返回值约定

### 单位约定

- 黑体族：波长米、温度开尔文；`PlanckRadianceByWavelength` 返回谱辐亮度 `W·sr⁻¹·m⁻³`，`StefanBoltzmannFlux` 返回 `W/m²`，`WienPeakWavelength` 返回米。
- 恒星族：MKS 变体用半径米、光度瓦特、温度开尔文；Solar 变体用太阳半径 R☉、太阳光度 L☉、温度开尔文，输入输出都是无量纲的太阳倍数。
- 星等族：距离秒差距 pc、星等 mag；`DistanceModulus` 返回 `m − M`，单位也是 mag。
- 距离单位：`Distance` 只做单位换算，输入单位由 `DistanceUnit` 指定，返回值一律是秒差距 pc。
- 望远镜族：口径毫米 mm；`DawesLimitArcsec`、`RayleighLimitArcsec` 返回**角秒**，不是度；`LightGatheringPowerRatio` 与 `LimitingMagnitudeEmpirical` 分别为无量纲比值与 mag。
- 大气质量族：高度角（或天顶距）为度；返回值是以天顶为 1 的无量纲相对大气质量。
- 会合周期：单位由调用者自定，两个输入必须一致，输出与之一致；本包不假定“天”。
- 本包不产生视直径/视半径。日食、月掩手册里的视半径字段以角秒计，本包只有 Dawes/Rayleigh 两个极限角用角秒，其余角度一律为度，不要混用。

### 时标

- 全部接口都不接收时刻，也不做任何时标换算：公式只依赖温度、波长、距离、口径、高度角等参数，与 UTC、UT1、TT 无关。
- 会合周期给的是周期长度，不是“下一次会合的时刻”。要落到日期，需要配合行星包的会合接口与民用时刻，时标口径见[时标约定](timescale.md)。
- 本包没有 ΔT、闰秒或 UT1 修正入口；这些只在 `coord`、`eclipse`、`moon` 等依赖时刻的链路里出现。

### 角度象限与弧度/度的边界

- 角度参数一律为度，内部转弧度计算，返回值也回到度或角秒，不会把弧度泄漏给调用者。
- 高度角与天顶距都限制在 `[0,90]`：地平为 `0°`、天顶为 `+90°`，天顶距与高度角互补。本包不做象限折叠，负高度（地平线以下）直接被判为无效，不会自动取绝对值或折到天顶。
- 天顶距 `z` 与高度角 `h` 的关系是 `z = 90° − h`；`AirmassPlaneParallel` 收 `h`，`AirmassPlaneParallelByZenithDistance` 收 `z`，两者对同一几何应给出一致结果。
- Dawes/Rayleigh 的返回值是角秒；若要与其他手册的“度”角度换算，需除以 3600。

### 零值与无效输入

- 黑体族对无效输入的处理刻意不一致：`WienPeakWavelength` 与 `PlanckRadianceByWavelength` 在温度 `≤ 0` 或非有限时返回 NaN；`StefanBoltzmannFlux` 只在温度 `< 0` 或非有限时返回 NaN，`0 K` 是合法输入并返回 `0`（0 K 的通量为 0，而峰值波长无定义）。
- 会合周期：任一周期 `≤ 0` 或非有限返回 NaN；两个周期相等时频率差为 0，返回 `+Inf`。
- 星等族：`distanceParsec ≤ 0` 或非有限时 `DistanceModulus` 返回 NaN，两个换算函数随之返回 NaN；`DistanceModulus(10)` 恒为 `0`。
- 望远镜族：口径（或第二个口径）`≤ 0` 或非有限返回 NaN；`LightGatheringPowerRatio` 的第二个口径为 `0` 会先被判为无效而不是除零；`LimitingMagnitudeEmpirical` 不接收瞳径参数，7 mm 是内置常数。
- 恒星族：所有输入必须 `> 0`，否则返回 NaN；`EffectiveTemperatureFromLuminositySolarRadius`、`LuminositySolarFromRadiusTemperature`、`RadiusSolarFromLuminosityTemperature` 先把太阳单位换成 SI 再计算。
- 大气质量族：高度角或天顶距落在 `[0,90]` 之外（含负值）或非有限都返回 NaN；`AirmassPlaneParallel(0)` 与 `AirmassPlaneParallelByZenithDistance(90)` 返回 `+Inf`；`AirmassKastenYoung(0)`、`AirmassPickering(0)` 仍是有限值。
- 本包没有 `(值, error)` 或 `(值, ok)` 形式的返回：无效输入统一用 NaN 表达，调用者用 `math.IsNaN` 判断。

### 精度与适用范围

- 黑体族是理想黑体模型，不含吸收线、星际消光与大气消光。`WienPeakWavelength` 使用维恩位移常数 `b = 2.897771955e-3 m·K`；按频率与按波长的峰值口径不同，严格意义上 `b/T` 是波长口径的峰值。
- 常数口径：`h = 6.62607015e-34 J·s`、`c = 299792458 m/s`、`k = 1.380649e-23 J/K`、`σ = 5.670374419e-8 W·m⁻²·K⁻⁴`；太阳参数 `L☉ = 3.828e26 W`、`R☉ = 6.957e8 m`、`Teff = 5772 K`，后者的公开入口是 `SolarEffectiveTemperature`。
- 星等族假设无消光、无 K 修正、无宇宙学项；两个换算函数只是加减 `DistanceModulus`，因此精度完全取决于外部给的绝对星等与距离。
- 望远镜族是可见光经验值：Dawes 与 Rayleigh 用固定系数（`116`、`138.4`，口径 mm），不随波长变化；`LightGatheringPowerRatio` 只比较口径平方，不含中心遮挡、透过率与副镜损失；`LimitingMagnitudeEmpirical` 不含天空背景、倍率、透过率与观测经验修正。
- 恒星族由 `L = 4πR²σT⁴` 互相反解，假设球对称、无临边昏暗修正、无自转与磁场效应；Solar 变体与 MKS 变体共用同一组太阳常数，两者之间的差异只来自这一组常数的口径。
- 大气质量族：平行平板是纯几何 `sec(z)`，只在中高空可作近似，接近地平线时发散；Kasten-Young(1989) 与 Pickering(2002) 是经验拟合，中高空彼此接近，低空差异最大。已经有视高度角时直接用 `AirmassKastenYoung` / `AirmassPickering`；只有真高度角而需要折射时用 `coord` 的[大气质量](coord.md#大气质量)。
- 本包全部函数都是无状态纯函数，不做缓存、不读全局时标状态，可以安全地在任意顺序、任意并发（调用方自行同步）下调用。
