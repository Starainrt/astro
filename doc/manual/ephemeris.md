# 显式星历实例

[English](en/ephemeris.md) | [返回 README](../../README.md)

包级函数使用进程当前的默认星历（默认构建为半解析模型）。需要在同一进程内并行使用多套星历模型、加载外部 JPL SPK 文件（`.bsp`）或叠加多个星历源时，使用 `astro.Context` 与 `astro.Instance`。

## 目录

- [基本示例](#基本示例)
- [主要功能](#主要功能)
  - [加载外部 JPL SPK 星历](#加载外部-jpl-spk-星历)
  - [读取原始几何位置与速度](#读取原始几何位置与速度)
  - [多源叠加](#多源叠加)
- [API 概览](#api-概览)
- [约定与限制](#约定与限制)

## 基本示例

通过 `astro.NewContext` 创建独立上下文，模型在创建时固定，互不干扰。上下文暴露的模块访问器（`Sun()`、`Moon()`、`Eclipse()` 等）与同名包级函数签名一致：

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/starainrt/astro"
)

func main() {
	cst := time.FixedZone("CST", 8*3600)
	date := time.Date(2026, 2, 17, 12, 0, 0, 0, cst)

	semi, err := astro.NewContext(astro.ModelSemianalytic) // 默认内置模型
	if err != nil {
		log.Fatal(err)
	}
	legacy, err := astro.NewContext(astro.ModelLegacy) // VSOP87/ELP2000-82
	if err != nil {
		log.Fatal(err)
	}

	ra1, dec1 := semi.Sun().ApparentRaDec(date)
	ra2, dec2 := legacy.Sun().ApparentRaDec(date)
	fmt.Printf("太阳视赤道对比: 半解析 (%.6f°, %.6f°)  Legacy (%.6f°, %.6f°)\n", ra1, dec1, ra2, dec2)

	// 事件查询同样通过上下文调用
	eDate := time.Date(2026, 3, 3, 12, 0, 0, 0, cst)
	le, ok := semi.Eclipse().LunarEclipseOnDate(eDate)
	fmt.Printf("2026-03-03 月食: found=%v  类型=%s  食甚=%s CST\n", ok, le.Type, le.Maximum.Format("15:04:05"))
}
```

输出：

```text
太阳视赤道对比: 半解析 (330.645545°, -11.998198°)  Legacy (330.645548°, -11.998203°)
2026-03-03 月食: found=true  类型=total  食甚=19:33:42 CST
```

`astro.SetModel` 仅修改包级函数的全局默认模型，不影响已创建的 `Context`。

## 主要功能

### 加载外部 JPL SPK 星历

通过 `spk.Open` 打开 `.bsp` 文件并以 `astro.ModelCustom` 传入 `astro.NewContext`，即可在全部高层观测与事件接口中使用该星历核：

```go
kernel, err := spk.Open("de440s.bsp")
if err != nil {
	log.Fatal(err)
}
defer kernel.Close()

ctx, err := astro.NewContext(astro.ModelCustom, kernel)
if err != nil {
	log.Fatal(err)
}
ra, dec := ctx.Sun().ApparentRaDec(date)
alt := ctx.Moon().Altitude(date, 116.4, 39.9)
fmt.Printf("DE440s: 太阳 (%.6f°, %.6f°)  月球高度 %.4f°\n", ra, dec, alt)
```

自定义引擎超出覆盖区间或求值失败时不会静默回退至内置表（标量入口返回 `NaN`，日期查询返回未找到）。

### 读取原始几何位置与速度

需要直接获取 ICRF 坐标系下的几何位置（km）与速度（km/日）时，使用 `astro.NewInstance(engine)` 或 `astro.NewModelInstance(astro.ModelSemianalytic)`：

```go
inst := astro.NewInstance(kernel)
jdTDB := calendar.Date2JD(astro.TDBFromTT(astro.TTFromUTC(date)))

pos, vel, err := inst.State(astro.BodyMoon, astro.BodyEarth, jdTDB)
startJD, endJD, ok := inst.Coverage(astro.BodyMoon)
```

- `State` 历元口径为 **TDB 儒略日**，越界返回 `astro.ErrEphemerisOutOfRange`。
- `ModelLegacy` 仅含标量级数、不提供三维几何引擎，调用 `State` 返回 `astro.ErrEphemerisNoEngine`。

### 多源叠加

小天体 SPK 核通常仅存储“目标相对太阳”的状态。`astro.NewStackInstance(de440, ceres)` 按传入先后顺序叠加多个引擎，并自动跨源拼接目标相对太阳系质心（`BodySSB`）的状态：

```go
de440, _ := spk.Open("de440s.bsp")
ceres, _ := spk.Open("ceres.bsp")
defer de440.Close()
defer ceres.Close()

stack := astro.NewStackInstance(de440, ceres) // 靠前的源优先
jdTT := calendar.Date2JD(astro.TTFromUTC(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)))
dir, distKM, lightTimeDays, err := stack.Apparent(astro.Body(20000001), jdTT) // 20000001 为谷神星 NAIF ID
```

也可通过 `orbit.NewEphemerisEngine` 将轨道根数传播器封装为引擎加入多源栈，见[通用小天体轨道](orbit.md)。

## API 概览

| 入口 | 用途 |
| --- | --- |
| `astro.NewContext(model, engines...)` | 创建隔离的星历上下文；提供 `Sun()`、`Moon()`、`Mercury()`…`Neptune()`、`Eclipse()`、`Phenomena()`、`Satellite()`、`GeoJSON()` 等模块 |
| `ctx.WithLunarLimb(limb)` | 派生启用月缘地形修正的新上下文，不修改原实例 |
| `astro.SetModel(model)` | 切换进程级默认模型（仅接受 `ModelSemianalytic` 或 `ModelLegacy`） |
| `astro.NewModelInstance` / `astro.NewInstance` | 创建底层 `*astro.Instance`，供 `State` / `Apparent` / `Coverage` 使用 |
| `astro.NewStackInstance(engines...)` | 按优先级叠加多个引擎并自动完成中心传递拼接 |
| `spk.Open(path)` | 读取 JPL SPK（`.bsp`）文件并实现 `astro.Engine` |

## 约定与限制

- **引擎契约**：自定义引擎须实现 `Engine.State(target, center, jdTDB)`（返回 ICRF km 与 km/日），且必须只读、可并发重入（库内不启动 goroutine）。未实现可选接口 `ApparentEngine.Apparent` 的引擎自动使用通用视位置链（光行时迭代、太阳引力偏折与相对论光行差），要求能提供地球、太阳及目标相对 `BodySSB` 的状态。
- **隔离范围**：`Context` / `Instance` 仅隔离星历模型；ΔT、TT−UTC 与民用时标政策仍为进程级设置。
- **沙罗编号**：自定义引擎可表达任意轨迹，日月食查询不推断沙罗编号（`HasSaros == false`）。
- **二进制体积**：显式指定单一模型不会减小可执行文件体积，两套内置表随主包链接。
- **时标近似**：高层标量与视位置入口以 TT 近似 TDB；需严格区分 TDB 时直接调用 `Instance.State`。
