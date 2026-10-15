# 显式星历实例

[English](en/ephemeris.md) | [返回 README](../../README.md)

包级函数使用进程当前的星历模型，默认是半解析模型。要在同一个进程里并用多套模型，或接入外部星历（SPK、自研引擎），就用 `astro.Context` 把星历固定下来；需要直接读取几何状态（位置和速度）时，再用底层的 `astro.Instance`。

## 目录

- [换用另一套内置模型](#换用另一套内置模型)
- [用 Context 求坐标、高度与升落](#用-context-求坐标高度与升落)
- [按日期查询事件](#按日期查询事件)
- [接入外部星历](#接入外部星历)
- [多源叠加](#多源叠加)
- [引擎契约](#引擎契约)
- [限制](#限制)

## 换用另一套内置模型

`astro.NewContext` 创建独立上下文，模型在创建时固定：

```go
import (
	"fmt"
	"time"

	"github.com/starainrt/astro"
)

date := time.Date(2026, 2, 17, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))

semi, err := astro.NewContext(astro.ModelSemianalytic) // 默认构建使用的模型
if err != nil {
	panic(err)
}
legacy, err := astro.NewContext(astro.ModelLegacy)
if err != nil {
	panic(err)
}

ra1, dec1 := semi.Sun().ApparentRaDec(date)
ra2, dec2 := legacy.Sun().ApparentRaDec(date)
fmt.Printf("%.8f %.8f\n", ra1, dec1)
fmt.Printf("%.8f %.8f\n", ra2, dec2)
```

```text
330.64554489 -11.99819781
330.64554785 -11.99820252
```

上下文一旦创建就固定模型，之后的 `astro.SetModel` 不会影响它。`SetModel` 改的是包级函数的默认模型，只接受两个内置模型，其他模型返回 `astro.ErrUnknownModel` 且不改动全局状态；自定义引擎必须显式传给 `Context` 或 `Instance`。`NewContext` 对未知模型返回 `ErrUnknownModel`，引擎个数或取值不符（内置模型带引擎、`ModelCustom` 不是恰好一个非 nil 引擎）返回 `ErrInvalidEngine`；`NewModelInstance` 只接受内置模型。

## 用 Context 求坐标、高度与升落

访问器返回的方法与同名包级函数参数、返回值一致，把 `sun.Xxx` 换成 `ctx.Sun().Xxx` 即可：

```go
ctx, err := astro.NewContext(astro.ModelSemianalytic)
if err != nil {
	panic(err)
}
date := time.Date(2026, 2, 17, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
lon, lat := 116.4, 39.9

// 太阳：地心视赤经/赤纬，站心高度角/方位角
ra, dec := ctx.Sun().ApparentRaDec(date)
alt := ctx.Sun().Altitude(date, lon, lat)
az := ctx.Sun().Azimuth(date, lon, lat)
fmt.Printf("太阳 ra=%.6f dec=%.6f alt=%.6f az=%.6f\n", ra, dec, alt, az)

// 月球：地心坐标用 GeocentricApparentRaDec，站心量直接带经纬度
moonRa, moonDec := ctx.Moon().GeocentricApparentRaDec(date)
moonAlt := ctx.Moon().Altitude(date, lon, lat)
moonRise, err := ctx.Moon().RiseTime(date, lon, lat, 0, true)
fmt.Printf("月球 ra=%.6f dec=%.6f alt=%.6f 升起=%s err=%v\n",
	moonRa, moonDec, moonAlt, moonRise.Format("15:04:05"), err)

// 行星按包名取访问器
marsRa, marsDec := ctx.Mars().ApparentRaDec(date)
marsAlt := ctx.Mars().Altitude(date, lon, lat)
fmt.Printf("火星 ra=%.6f dec=%.6f alt=%.6f\n", marsRa, marsDec, marsAlt)
```

```text
太阳 ra=330.645545 dec=-11.998198 alt=37.682918 az=171.221286
月球 ra=327.197499 dec=-14.621205 alt=34.596340 升起=06:59:57 err=<nil>
火星 ra=322.193276 dec=-16.023967 alt=34.060823
```

时刻为 2026-02-17 12:00 CST，观测点东经 116.4°、北纬 39.9°。`RiseTime` 在当天当地民用日内找不到升落时返回错误，不是零值时刻。

| 入口 | 所属包 | 用途 |
| --- | --- | --- |
| `Sun()` | `sun` | 地心坐标、站心高度/方位/天顶距、升落与朦影、视直径、物理星历 |
| `Moon()` | `moon` | 地心与站心坐标、高度/方位、升落、月相、地心距、天平动 |
| `Mercury()` … `Neptune()` | 各行星包 | 视坐标、高度/方位、升落、相位与视星等、冲合和大距等事件 |
| `Eclipse()` | `eclipse` | 日月食查询、地方接触时刻、食带路径 |
| `Phenomena()` | `phenomena` | 黄经穿越、留、最近角距、任意高度升落、偕日可见 |
| `GeoJSON()` | `geojson` | 天象地图数据 |

`...N` 方法走快速表，例如 `ctx.Sun().AltitudeN(date, lon, lat, n)`、`ctx.Moon().TrueLoN(date, n)`；`n` 的含义见[精度与性能](accuracy.md)。

## 按日期查询事件

事件查询也挂在模块上：

```go
date := time.Date(2026, 3, 3, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))

info, ok := ctx.Eclipse().LunarEclipseOnDate(date)
next := ctx.Eclipse().NextLunarEclipse(date)
fmt.Println(ok, info.Maximum.Format("2006-01-02 15:04 MST"), info.Type)
fmt.Println(next.Maximum.Format("2006-01-02 15:04 MST"), next.Type)
```

```text
true 2026-03-03 19:33 CST total
2026-03-03 19:33 CST total
```

日期当天有月食时 `ok` 才为 `true`；示例日期 2026-03-03 当天有月食。筛选与搜索均使用该上下文的实例。内置模型的沙罗表按实例模型选择；自定义引擎可以描述任意轨迹，因此不推断沙罗编号，结果中 `HasSaros` 为 `false`。

## 接入外部星历

`spk.Open` 读取 JPL SPK 核，作为 `astro.Engine` 传给 `NewContext`，之后的调用方式和内置模型完全相同：

```go
import "github.com/starainrt/astro/planet/spk"

kernel, err := spk.Open("de440s.bsp")
if err != nil {
	panic(err)
}
defer kernel.Close()

custom, err := astro.NewContext(astro.ModelCustom, kernel)
if err != nil {
	panic(err)
}

// 沿用上面的 date（2026-02-17 12:00 CST）与经纬度
ra, dec := custom.Sun().ApparentRaDec(date)
alt := custom.Sun().Altitude(date, lon, lat)
marsRa, marsDec := custom.Mars().ApparentRaDec(date)
marsAlt := custom.Mars().Altitude(date, lon, lat)
moonAlt := custom.Moon().Altitude(date, lon, lat)
fmt.Printf("太阳 ra=%.6f dec=%.6f alt=%.6f\n", ra, dec, alt)
fmt.Printf("火星 ra=%.6f dec=%.6f alt=%.6f\n", marsRa, marsDec, marsAlt)
fmt.Printf("月球 alt=%.6f\n", moonAlt)
```

```text
太阳 ra=330.645545 dec=-11.998198 alt=37.682918
火星 ra=322.193276 dec=-16.023967 alt=34.060823
月球 alt=34.596325
```

同一时刻与内置半解析基本一致，这里月球高度差约 `0.05″`。

需要原始几何状态（位置和速度）时改用 `Instance`：

```go
inst := astro.NewInstance(kernel)
jdTDB := calendar.Date2JD(astro.TDBFromTT(date))
pos, vel, err := inst.State(astro.BodyMoon, astro.BodyEarth, jdTDB)
start, end, ok := inst.Coverage(astro.BodyMoon)
fmt.Printf("pos=%.1f,%.1f,%.1f km\nvel=%.1f,%.1f,%.1f km/日\nerr=%v coverage=%.1f..%.1f ok=%v\n",
	pos[0], pos[1], pos[2], vel[0], vel[1], vel[2], err, start, end, ok)
```

```text
pos=312305.1,-204135.4,-98208.0 km
vel=48600.2,64310.3,36543.8 km/日
err=<nil> coverage=2396752.5..2506352.5 ok=true
```

`State` 接收 TDB 历元，返回 ICRF 几何位置（km）与速度（km/日）。`Coverage` 给出该天体的可用历元区间，越界返回 `ErrEphemerisOutOfRange`。

## 多源叠加

`astro.NewStackInstance` 按先后顺序合成多个引擎，先传入的源优先。小天体核常常只给“目标相对太阳”的状态，Stack 会跨源拼出目标相对太阳系质心的状态：

```go
de440, _ := spk.Open("de440s.bsp")
ceres, _ := spk.Open("ceres.bsp")
defer de440.Close()
defer ceres.Close()

stack := astro.NewStackInstance(de440, ceres)

// 小天体核的覆盖往往很窄，查询历元必须落在覆盖内
smallDate := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
jdTT := calendar.Date2JD(astro.TTFromUTC(smallDate))
dir, dist, tau, err := stack.Apparent(astro.Body(20000001), jdTT)
fmt.Printf("%.8f,%.8f,%.8f\ndist=%.0f km tau=%.5f 日 err=%v\n", dir[0], dir[1], dir[2], dist, tau, err)
```

```text
0.98097637,0.17021858,-0.09333272
dist=417020638 km tau=0.01610 日 err=<nil>
```

`astro.Body` 是 `int`，任意 NAIF 编号可以直接写，Ceres 是 `20000001`。示例这份 Ceres 核只覆盖 2026-01-01 至 01-03（`stack.Coverage` 返回 `2461041.5..2461043.5`），超出覆盖返回 `ErrEphemerisOutOfRange`，不会静默改用别的源。

## 引擎契约

- `Engine.State(target, center, jdTDB)`：ICRF 几何位置（km）与速度（km/日），历元为 TDB。自定义引擎必须实现它。
- 可选的 `ApparentEngine.Apparent(target, jdTT)`：TT 历元的地心 ICRF 单位视方向、距离（km）与光行时（日）。只实现 `State` 的引擎（如 SPK）走通用矢量视位置链：迭代光行时，加太阳引力偏折与相对论光行差；这需要引擎能给出地球、太阳和目标相对太阳系质心（`BodySSB`）的状态。
- 引擎会被并发调用，必须只读、可重入；库不会启动 goroutine。
- 自定义引擎的失败不回退到内置表：标量入口返回 `NaN`，月食核心的 `Maximum` 为 `NaN` 且各阶段标志为 `false`，日期查询返回未找到。
- 标量坐标和 `basic` 月食入口按 TT 取历元；几何引擎目前以 TT 近似 TDB，需要严格 TDB 时直接调用 `Instance.State`。

## 限制

- `EphemerisLegacy` 实例只有标量入口，没有几何引擎：`Instance.State` 返回 `astro.ErrEphemerisNoEngine`，`Coverage` 与 `Apparent` 报告不可用。需要位置和速度时用半解析模型或自定义引擎。
- 显式实例只决定用哪份星历，不保证减小二进制体积：两张内置表都会链接进程序。
- 半解析模型的可用区间是 TT JD `-1931455.0 … 5373545.0`，越界仍外推；行星现代段为 JD `2086295.0 … 2816795.0`。误差见[精度与性能](accuracy.md)。
- `Stack.Coverage` 是各源区间的并集，中间可能有断档，也不保证中心链（Stack 为拼接状态在源之间搜索的中心传递路径）所需的每个天体都可用；以 `State` 在具体历元的返回值为准。
- `basic` 中支持实例的入口多数带 `WithInstance` 后缀，带 `n` 参数的入口带 `NWithInstance`，传 `nil` 表示使用进程当前的全局模型。例外是影半径入口 `LunarEclipseWithLunarLimb` 和 18 个大气升落入口，它们把 `*Instance` 放在参数末尾。一般不需要直接调用这些入口。
