# 行星

[English](en/planets.md) | [返回 README](../../README.md)

七大行星各自对应一个同名包：`mercury`、`venus`、`mars`、`jupiter`、`saturn`、`uranus`、`neptune`。各包基础接口命名统一，并按内、外行星及各自特征提供专属能力：
- **内行星（水星、金星）**：上合/下合、东西大距、地心与站心凌日。
- **外行星（火星至海王星）**：冲日、东西方照。
- **专属接口**：木星伽利略卫星位置、现象、接触窗口与互掩互食（`jupiter`），土星环参数（`saturn`），火星太阳日、协调火星时与火星日出日落（`mars`、`sun`）。

木星卫星出图见[木星卫星 SVG](jupiter-svg.md)，凌日出图见[日月食与凌日](eclipse.md#凌日出图)，行星月掩见[月掩](occultation.md)。

## 目录

- [基本示例](#基本示例)
- [主要功能](#主要功能)
  - [冲日、合日、大距与留](#冲日合日大距与留)
  - [水星与金星凌日](#水星与金星凌日)
  - [物理星历与土星环](#物理星历与土星环)
  - [木星伽利略卫星](#木星伽利略卫星)
  - [行星本地时间与木星旋转角](#行星本地时间与木星旋转角)
- [API 概览](#api-概览)
- [约定与限制](#约定与限制)

## 基本示例

计算 2026-01-15 20:00 CST 火星的地心视坐标、视星等、地火距，西安当日火星升落时刻及下一次冲日时刻：

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro/mars"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2026, 1, 15, 20, 0, 0, 0, cst)
	lon, lat := 108.93, 34.27 // 西安（东经/北纬，度）

	ra, dec := mars.ApparentRaDec(date)
	rise, err := mars.RiseTime(date, lon, lat, 0, true)
	if err != nil {
		log.Fatal(err)
	}
	set, err := mars.SetTime(date, lon, lat, 0, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("火星 (2026-01-15 20:00 CST): RA=%.2f° Dec=%.2f° 星等=%.2f 地火距=%.3f AU\n",
		ra, dec, mars.ApparentMagnitude(date), mars.EarthDistance(date))
	fmt.Printf("西安火星升落: %s / %s  下次冲日: %s\n",
		rise.Format("15:04:05"), set.Format("15:04:05"),
		mars.NextOpposition(date).Format("2006-01-02 15:04"))
}
```

输出：

```text
火星 (2026-01-15 20:00 CST): RA=295.93° Dec=-22.29° 星等=1.15 地火距=2.398 AU
西安火星升落: 07:50:30 / 17:46:23  下次冲日: 2027-02-19 23:50
```

其余六颗行星包具有同名基础接口；经 `ctx.Mars()` 等访问器可绑定显式星历实例。

## 主要功能

### 冲日、合日、大距与留

行星天象搜索成对提供 `Last*`（上一次）与 `Next*`（下一次），返回值保留输入 `date` 的时区：

```go
fmt.Printf("金星下次东大距: %s  水星下次下合: %s\n",
	venus.NextGreatestElongationEast(date).Format("2006-01-02 15:04"),
	mercury.NextInferiorConjunction(date).Format("2006-01-02 15:04"))
```

输出：

```text
金星下次东大距: 2026-08-15 14:31  水星下次下合: 2026-03-07 19:01
```

- **留**：七大行星均提供 `Last/NextProgradeToRetrograde`（顺转逆）与 `Last/NextRetrogradeToPrograde`（逆转顺）；水星与金星另提供不区分方向的 `Last/NextRetrograde`。
- **内行星专属**：`SuperiorConjunction`（上合）、`InferiorConjunction`（下合）、`GreatestElongation` / `...East` / `...West`（大距）。
- **外行星专属**：`Opposition`（冲日）、`EasternQuadrature`（东方照）、`WesternQuadrature`（西方照）。

### 水星与金星凌日

通过 `NextTransit` / `LastTransit` / `ClosestTransit` 搜索全球凌日事件（先检查 `tr.Valid`），再调用 `LocalTransit(tr, lon, lat, height)` 计算指定站点的可见性与 C1–C4 接触时刻：

```go
tr := mercury.NextTransit(date)
if tr.Valid {
	ltr := mercury.LocalTransit(tr, 116.4, 39.9, 50) // 北京
	fmt.Printf("下次水星凌日: %s UTC (C1=%s, 凌甚=%s, C4=%s, 北京可见=%v)\n",
		tr.Greatest.UTC().Format("2006-01-02"),
		tr.Start.UTC().Format("15:04:05"),
		tr.Greatest.UTC().Format("15:04:05"),
		tr.End.UTC().Format("15:04:05"),
		ltr.Visible)
}
```

输出：

```text
下次水星凌日: 2032-11-13 UTC (C1=06:41:13, 凌甚=08:54:13, C4=11:07:16, 北京可见=true)
```

掠凌或地心无接触的边缘凌日可能不存在内切（`HasInternal == false`）甚至地心无外切（`HasExternal == false`），站点可见性以 `LocalTransit` 结果为准。SVG 出图见[凌日出图](eclipse.md#凌日出图)，GeoJSON 导出见[天象地图、GeoJSON 与 KML](map-geojson.md)。

### 物理星历与土星环

- 七大行星均提供 `Physical(date)`（盘面子地/子日经纬度、北极位置角）。
- 木星另提供 `CentralMeridians(date)`（System I / II / III 中央经线）；土星与天王星另提供 `PhysicalSystemIII(date)`。
- 土星环参数由 `saturn.Ring(date)` 返回：

```go
ring := saturn.Ring(date)
fmt.Printf("土星环: 张角 B=%.2f°  外缘长轴=%.1f\"  短轴=%.1f\"\n",
	ring.EarthLatitude, ring.MajorAxis, ring.MinorAxis)
```

输出：

```text
土星环: 张角 B=-1.47°  外缘长轴=37.8"  短轴=1.0"
```

### 木星伽利略卫星

`jupiter` 包对四颗伽利略卫星（`GalileanSatelliteIo`、`Europa`、`Ganymede`、`Callisto`）提供四层接口：

1. **瞬时位置与状态**：`Satellites(date)`（相对木星半径的 X/Y 偏移、前后关系、视赤经赤纬）与 `SatellitePhenomena(date)`（凌木、木掩、木影食、影凌标志）。
2. **常规卫星现象**：点中心模型 `NextGalileanPhenomenonEvent`（用于快速区间检索）与有限圆盘 IMCCE 风格接触窗口 `NextGalileanPhenomenonContactEvent`（给出消失 `Disappearance` 与再现 `Reappearance` 窗口，推荐用于观测预报与年表对照）。
3. **木卫互掩与互食**：`NextGalileanMutualEvent(date, satA, satB, kind)` 返回主/被动卫星、起止/食甚时刻、几何遮挡比 `Coverage`（`0..1`）与是否内含/环食 `Annular`。
4. **批量汇总**：`GalileanEventsBetween(start, end, options)` 按时间顺序汇总区间内（上限 366 天）的全部常规现象、接触窗口与互现象。

```go
satName := map[int]string{
	jupiter.GalileanSatelliteIo:       "木卫一(Io)",
	jupiter.GalileanSatelliteEuropa:   "木卫二(Europa)",
	jupiter.GalileanSatelliteGanymede: "木卫三(Ganymede)",
	jupiter.GalileanSatelliteCallisto: "木卫四(Callisto)",
}

t := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

sats := jupiter.Satellites(t)
fmt.Printf("木卫一位置 (2026-06-01 00:00 UTC): X=%+.2f Rj  Y=%+.2f Rj  在木星前=%v\n",
	sats.Io.OffsetXJupiterR, sats.Io.OffsetYJupiterR, sats.Io.InFrontOfJupiter)

occ := jupiter.NextGalileanMutualEvent(t, jupiter.GalileanSatelliteIo, jupiter.GalileanSatelliteEuropa, jupiter.GalileanMutualOccultation)
if occ.Valid {
	fmt.Printf("下次木卫互掩: %s 掩 %s  %s – %s UTC  食甚 %s  遮挡 %.1f%%  持续 %s\n",
		satName[occ.Active], satName[occ.Passive],
		occ.Start.Format("2006-01-02 15:04:05"), occ.End.Format("15:04:05"),
		occ.Greatest.Format("15:04:05"), occ.Coverage*100, occ.Duration.Round(time.Second))
}

ecl := jupiter.NextGalileanMutualEvent(t, jupiter.GalileanSatelliteIo, jupiter.GalileanSatelliteEuropa, jupiter.GalileanMutualEclipse)
if ecl.Valid {
	fmt.Printf("下次木卫互食: %s 食 %s  %s – %s UTC  食甚 %s  本影遮挡 %.1f%%  持续 %s\n",
		satName[ecl.Active], satName[ecl.Passive],
		ecl.Start.Format("2006-01-02 15:04:05"), ecl.End.Format("15:04:05"),
		ecl.Greatest.Format("15:04:05"), ecl.Coverage*100, ecl.Duration.Round(time.Second))
}

all, err := jupiter.GalileanEventsBetween(t, t.Add(7*24*time.Hour), jupiter.GalileanEventsOptions{
	IncludeMutual: true,
})
if err != nil {
	log.Fatal(err)
}
fmt.Printf("2026-06-01 起 7 天内伽利略卫星事件总数: %d（首条: %s %s）\n",
	len(all), all[0].Greatest.Format("01-02 15:04"), all[0].Family)
```

输出：

```text
木卫一位置 (2026-06-01 00:00 UTC): X=-2.21 Rj  Y=+0.56 Rj  在木星前=false
下次木卫互掩: 木卫一(Io) 掩 木卫二(Europa)  2026-06-01 01:56:24 – 02:26:20 UTC  食甚 02:11:26  遮挡 100.0%  持续 29m56s
下次木卫互食: 木卫一(Io) 食 木卫二(Europa)  2026-06-01 13:52:46 – 14:02:19 UTC  食甚 13:57:34  本影遮挡 4.9%  持续 9m33s
2026-06-01 起 7 天内伽利略卫星事件总数: 35（首条: 06-01 02:48 phenomenon）
```

- 互现象返回全部几何事件（含被动卫星处于木星掩蔽或木影中的情况）；可观测性通过 `event.GreatestState[event.Passive-1]` 判断。
- 木影食若无法求得常规 D/F 接触窗口，两段均标记 `PenumbralFallback: true` 表示回退至半影边界口径。出图见[木星卫星 SVG](jupiter-svg.md)。

### 行星本地时间与木星旋转角

接收地球民用时刻 `time.Time`，计算火星太阳日（MSD）、协调火星时（MTC）、太阳黄经（Ls）、地方太阳时、火星日出日落及木星旋转角：

```go
tM := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
mRise, err := sun.NextRiseOnMars(tM, 137.4, -4.6) // 盖尔撞击坑（东经 137.4°，南纬 4.6°）
if err != nil {
	log.Fatal(err)
}

fmt.Printf("火星 (2026-11-14 00:00 UTC): MSD=%.2f  MTC=%.2f h  Ls=%.2f°  盖尔坑LMST=%.2f h  下次日出=%s UTC\n",
	mars.SolDate(tM), mars.CoordinatedTime(tM), mars.SolarLongitude(tM),
	mars.LocalMeanSolarTime(tM, 137.4), mRise.Format("15:04:05"))
fmt.Printf("木星 System III 旋转角: %.2f°\n", jupiter.RotationAngle(tM, jupiter.SystemIII))
```

输出：

```text
火星 (2026-11-14 00:00 UTC): MSD=54342.55  MTC=13.27 h  Ls=21.58°  盖尔坑LMST=22.43 h  下次日出=08:15:55 UTC
木星 System III 旋转角: 130.68°
```

`mars.LocalMeanSolarTime` 遵循 Mars24 线性约定（`MTC + 东经/15`）；`mars.LocalTrueSolarTime` 与[跨行星表面观测](phenomena.md#跨行星表面观测)真太阳时一致。`jupiter.RotationAngle` 支持 `SystemI`、`SystemII`、`SystemIII`，返回请求时刻的 ICRF 本初子午线角 W（不扣除地木光行时）。

## API 概览

| 能力 | 适用包 | 代表入口 | 单位 |
| --- | --- | --- | --- |
| 地心视位置与距离 | 全部 7 包 | `ApparentLo/Bo`、`ApparentRa/Dec/RaDec`、`EarthDistance`、`SunDistance` | 度 / AU |
| 站心地平与升落 | 全部 7 包 | `Altitude`、`Azimuth`、`Zenith`、`RiseTime`、`SetTime`、`CulminationTime` | 度 / `time.Time` |
| 测光与视径 | 全部 7 包 | `ApparentMagnitude`、`Diameter`、`Semidiameter`、`PhaseAngle`、`IlluminatedFraction`（别名 `Phase`）、`BrightLimbPositionAngle` | mag / 角秒 / 度 |
| 轨道节点 | 全部 7 包 | `AscendingNode`、`DescendingNode` | 度 |
| 合日与留 | 全部 7 包 | `Last/NextConjunction`、`Last/NextProgradeToRetrograde`、`Last/NextRetrogradeToPrograde` | `time.Time` |
| 内行星天象 | `mercury`、`venus` | `Last/NextSuperiorConjunction`、`InferiorConjunction`、`GreatestElongation[East/West]`、`Last/Next/ClosestTransit`、`LocalTransit` | `time.Time` / `TransitInfo` |
| 外行星天象 | `mars` … `neptune` | `Last/NextOpposition`、`EasternQuadrature`、`WesternQuadrature` | `time.Time` |
| 物理星历 | 全部 7 包 | `Physical`（木星另含 `CentralMeridians`，土星另含 `Ring`、`PhysicalSystemIII`，天王星另含 `PhysicalSystemIII`） | 度 / 角秒 |
| 伽利略卫星 | `jupiter` | `Satellites`、`SatellitePhenomena`、`*GalileanPhenomenonEvent`、`*GalileanPhenomenonContactEvent`、`*GalileanMutualEvent`、`GalileanEventsBetween` | 木星半径 / 角秒 / `time.Time` |
| 火星钟与木星转角 | `mars`、`sun`、`jupiter` | `mars.SolDate` / `CoordinatedTime` / `LocalMeanSolarTime` / `LocalTrueSolarTime` / `SolarLongitude`、`sun.NextRiseOnMars` / `NextSetOnMars`、`jupiter.RotationAngle` | Sol / 小时 / 度 |

## 约定与限制

- **地心与站心**：`ApparentRaDec`、`HourAngle` 与 `ParallacticAngle` 为地心量；`Altitude`、`Azimuth`、`Zenith` 与 `RiseTime`、`SetTime` 为站心量（`aero=true` 计入标准地平折射，`height` 为椭球高米）。无升落时各包返回对应的 `ERR_<PLANET>_NEVER_RISE` / `ERR_<PLANET>_NEVER_SET`（`DownTime` 为 `SetTime` 旧兼容别名）。
- **`...N` 变体**：瞬时求值函数均提供 `...N(..., n)` 变体；`n < 0` 使用完整级数，半解析模型下 `n >= 0` 使用同一快速表（仅适合粗搜索）。事件搜索函数不提供 `N` 变体。
- **物理经度方向**：`PhysicalInfo` 的 `SubEarthLongitude` / `SubSolarLongitude` 遵循当前 IAU 制图惯例——水星、火星、木星、土星、海王星以**西经为正**，金星与天王星以**东经为正**。
- **伽利略卫星精度**：采用内置木卫理论，与 JPL Horizons 时刻对照一般偏差数秒至数十秒，适用于常规预报与制图；高精度光度曲线计时应配套专用卫星星历。
