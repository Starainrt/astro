# 时标

[English](en/timescale.md) | [README](../../README.md)

同一个物理时刻可以用 UTC、UT1 或 TT 表示。UTC（协调世界时）用于民用时间，UT1（世界时）反映地球自转，TT 用于星历计算。它们之间的两个差值是 `ΔT = TT − UT1` 和 `DUT1 = UT1 − UTC`，单位都是秒。

## 目录

- [时间参数](#时间参数)
- [time.Time 换算 API](#timetime-换算-api)
- [儒略日换算 API](#儒略日换算-api)
- [TCG、TCB 与 TDB](#tcgtcb-与-tdb)
- [ΔT 模型](#δt-模型)
  - [比较不同模型](#比较不同模型)
  - [接入外部模型](#接入外部模型)
- [未来的 UTC 对齐口径假设](#未来的-utc-对齐口径假设)
- [SVG、GeoJSON 与 KML 的时间标签](#svggeojson-与-kml-的时间标签)

## 时间参数

一般观测接口接收表示民用时刻的 `time.Time`，先取其 UTC 值，再按计算需要转换到 UT1 或 TT。改变 `Location` 只改变同一时刻的显示方式；日出日落等按日搜索的接口还会用它确定当地日期。

本库约定：1972-01-01 以前的民用时间按 UT1 处理；这是库的历算约定，不表示当时不存在 UTC。1972 年以后，TT−UTC 由内置闰秒表、所选政策或显式覆盖决定，ΔT 模型负责 TT−UT1。

以下输入需要另外区分：

| 输入 | 解释 |
| --- | --- |
| `astro.TTFromUTC` / `UT1FromUTC` 的参数 | 民用时刻，可带任意时区 |
| `astro.UTCFromTT` / `UTCFromUT1` 的参数 | TT / UT1 读数，以 UTC Location 承载 |
| `astro.TCGFromTT` / `TCBFromTT` / `TDBFromTT` 的参数 | TT 读数，不是民用 UTC 时刻 |
| `orbit.Elements.EpochJD` / `TpJD` | TT/TDB 儒略日，见[轨道手册](orbit.md) |
| `basic` 的 JD 数值接口 | 由函数名和参数契约指定时标，数值本身不记录时标 |
| `calendar.Date2JD` / `basic.Date2JD` | 读取年月日和钟面字段，不自动转换时区；需要 UTC JD 时先传入 `date.UTC()` |
| 公农历转换 | 按[历法手册](calendar.md)解释日期，默认使用北京时间 |

`time.Time` 不能表示闰秒的 `23:59:60`。历史日期还须考虑本库的儒略历/格里高利历切换与 Go 前推格里高利历之间的差别。

## time.Time 换算 API

这些接口位于根包 `github.com/starainrt/astro`。

| 函数 | 返回值 |
| --- | --- |
| `TTFromUTC(date)` | 同一时刻的 TT 读数 |
| `UTCFromTT(tt)` | TT 读数对应的UTC时刻 |
| `UT1FromUTC(date)` | 同一时刻的 UT1 读数 |
| `UTCFromUT1(ut1)` | UT1 读数对应的UTC时刻 |
| `DUT1(date)` | UT1−UTC，秒 |
| `LabelIn(scale, date)` | `TimeScaleUTC` 保留输入；`TimeScaleUT1` 返回 UT1 读数 |
| `TCGFromTT(tt)` / `TTFromTCG(tcg)` | TT 与地心坐标时 TCG 互换 |
| `TCBFromTT(tt)` / `TTFromTCB(tcb)` | TT 与太阳系质心坐标时 TCB 互换，采用地心近似 |
| `TDBFromTT(tt)` / `TTFromTDB(tdb)` | TT 与太阳系质心力学时 TDB 互换，采用地心近似 |
| `TCBFromTDB(tdb)` / `TDBFromTCB(tcb)` | TDB 与 TCB 的线性互换 |
| `TCGMinusTT(tt)` / `TCBMinusTT(tt)` / `TDBMinusTT(tt)` | 相应时标与 TT 的秒差，输入为 TT 读数 |

TT、UT1、TCG、TCB、TDB 的换算结果虽然使用 `time.Time` 类型和 UTC Location，但字段表示的是相应时标的读数。不能把这个值直接传回要求民用时刻的太阳、月亮等接口，否则相当于再次移动了计算时刻。需要还原时，调用对应逆变换。

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro"
)

func main() {
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	tt := astro.TTFromUTC(date)
	ut1 := astro.UT1FromUTC(date)
	fmt.Println("UTC:", date.Format(time.RFC3339Nano))
	fmt.Println("TT:", tt.Format("2006-01-02 15:04:05.000000"))
	fmt.Println("UT1:", ut1.Format("2006-01-02 15:04:05.000000"))
	fmt.Printf("DUT1: %.6f s\n", astro.DUT1(date))
	fmt.Println("UTC from TT:", astro.UTCFromTT(tt).Format(time.RFC3339Nano))
}
```

```text
UTC: 2026-04-01T00:00:00Z
TT: 2026-04-01 00:01:09.184014
UT1: 2026-04-01 00:00:00.051015
DUT1: 0.051000 s
UTC from TT: 2026-04-01T00:00:00Z
```

数值换算经过浮点儒略日，往返结果可能有微小舍入差。TT/UT1 的显示格式不加 `Z`，以免被当成 UTC 时间戳。

## 儒略日换算 API

下列函数位于 `basic`，参数与返回值是 `float64`。换算函数返回儒略日，差值函数返回秒。

| 函数 | 输入 → 输出 | 说明 |
| --- | --- | --- |
| `UTC2TT` | UTC JD → TT JD | 1972 年前按 UT1，之后使用 TT−UTC 模型 |
| `TT2UTC` | TT JD → UTC JD | 上述逆变换 |
| `UT12TT` | UT1 JD → TT JD | 使用当前 ΔT 模型 |
| `TT2UT1` | TT JD → UT1 JD | 上述逆变换 |
| `UTC2UT1` | UTC JD → UT1 JD | 等于 `TT2UT1(UTC2TT(jd))` |
| `UT12UTC` | UT1 JD → UTC JD | 上述逆变换 |
| `TTMinusUTCSeconds` | UTC JD → 秒 | 内置闰秒段为 `32.184 + (TAI−UTC)`；也可由注入函数覆盖 |
| `DUT1Seconds` | UTC JD → 秒 | `(TT−UTC) − ΔT` |
| `DeltaT` | JD 或十进制年 → 秒 | 第二参数为 `true` 时输入是 UT 儒略日，为 `false` 时是十进制年 |
| `TT2TCG` / `TCG2TT` | TT JD ↔ TCG JD | 线性换算 |
| `TT2TCB` / `TCB2TT` | TT JD ↔ TCB JD | 地心近似 |
| `TT2TDB` / `TDB2TT` | TT JD ↔ TDB JD | 地心近似 |
| `TCB2TDB` / `TDB2TCB` | TCB JD ↔ TDB JD | 线性换算，含 TDB0 常数 |
| `TCGMinusTTSeconds` / `TCBMinusTTSeconds` / `TDBMinusTTSeconds` | TT JD → 秒 | 相应时标与 TT 的秒差 |

闰秒或模拟闰时带来阶跃时，TT→民用时间的逆变换存在与阶跃宽度相同的不可表示区间；不能要求跨该区间的逐值往返恒等。

`TTMinusUTCSeconds` 与 `DefaultTTMinusUTC()` 读取闰秒表值（前者可被注入函数覆盖），不应用 `UTC2TT` 的未来政策外推。换算未来民用时刻时直接使用 `UTC2TT` 或 `TTFromUTC`。

## TCG、TCB 与 TDB

TCG 是地心坐标时；TCB 与 TDB 分别是太阳系质心坐标时和太阳系质心力学时。TT 与 TCG、TCB 与 TDB 的关系为线性定义；TT 与 TDB 的关系在这里采用地心近似，不含观测者位置引起的日周项。

令 `T0 = 2443144.5003725`、`LG = 6.969290134e-10`、`LB = 1.550519768e-8`、`TDB0 = −65.5e-6` 秒，时标读数用 JD 表示：

```text
TCG − TT = LG / (1 − LG) × (TT − T0)
TDB = TCB − LB × (TCB − T0) + TDB0 / 86400
TCB − TT = [LB × (TT − T0) + (TDB − TT) − TDB0 / 86400] / (1 − LB)
```

参考历元 T0 不表示四个时标的读数在此全部相等。现代日期 TDB−TT 的主要年周期振幅约 1.7 ms；TCG−TT 每年增加约 0.022 秒，TCB−TT 每年增加约 0.489 秒。

TDB−TT 使用 40 项截断级数和质量调整项。在 −3000 至 +6000 年范围，相对完整 787 项地心级数，遗漏项绝对幅度之和给出的保守截断误差上界为 11 µs；每 31 天抽样的最大差约 2.0 µs。抽样最大差不是全时域误差保证，截断误差也不包含完整模型自身的误差；范围外不作精度承诺。

单个 `float64` JD 在现代日期附近的分辨率约为 40 µs，`time.Time` 包装同样经过这一步舍入。比较微秒级时标差值应使用 `*MinusTT` 或 `*MinusTTSeconds`，不要把两个完整 JD 相减后再换算成秒。

## ΔT 模型

默认模型为 SMH2016 与 Morrison 2021 的样条和长期外推，并在实测覆盖期内优先使用逐月 ΔT 表。当前表延伸到 2026 年 10 月 1 日，2026 年 9 月 1 日为最终值，末点为预备扩展值；表外在端点作常值锚定以保持连续。`DeltaT` 对超出 ±40000 年的输入返回 `NaN`。

| 常量 | 模型 |
| --- | --- |
| `DeltaTModelDefault` / `DeltaTModelSMH2016` | 内置默认模型 |
| `DeltaTModelMS2004` | Morrison & Stephenson 2004，`ΔT = −20 + 32u²`，`u = (year−1820)/100` |
| `DeltaTModelEspenakMeeus2006` | Espenak & Meeus 2006 分段多项式 |
| `DeltaTModelNASACanon2006` | 上述多项式加 NASA 目录配对项：1955 年以前加 `−0.000012932(year−1955)²` |
| `DeltaTModelManual` | 查询状态时表示当前使用任意注入函数，不能作为命名模型安装 |

`SetDeltaTModel(model, keepObserved)` 设置进程级模型，返回是否成功；未知模型不会改变当前状态。`keepObserved=true` 保留实测段，模型只用于表外；`false` 则全程使用所选模型。`GetDeltaTModel()` 返回模型标识和该开关。

### 比较不同模型

`DeltaTModelSeconds` 不改变进程状态，适合比较同一时刻的不同 ΔT 取值：

```go
package main

import (
	"fmt"
	"time"

	"github.com/starainrt/astro"
	"github.com/starainrt/astro/basic"
)

func main() {
	date := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	jd := basic.Date2JD(date)
	for _, model := range []astro.DeltaTModel{
		astro.DeltaTModelSMH2016,
		astro.DeltaTModelEspenakMeeus2006,
		astro.DeltaTModelNASACanon2006,
	} {
		fmt.Println(model, astro.DeltaTModelSeconds(model, jd, false))
	}
}
```

```text
smh2016 86.98252070239269
espenak_meeus_2006 202.7448354310582
nasa_canon_2006 202.7448354310582
```

ΔT 的未来值不能精确预知。按现有模型计算，Espenak–Meeus 2006 与默认模型在 2035、2050、2100、2200 年的差异约为 11、21、116、275 秒。远期日月食的时刻和地面路径会受此影响；与日月食目录对照时，应先统一 ΔT 模型。

### 接入外部模型

| 根包接口 | 用途 |
| --- | --- |
| `DeltaT()` / `SetDeltaT(fn)` | 读取、设置 ΔT 函数；函数类型为 `func(float64, bool) float64` |
| `DefaultDeltaT()` | 取得内置默认 ΔT 函数 |
| `TTMinusUTC()` / `SetTTMinusUTC(fn)` | 读取、设置 TT−UTC 覆盖；函数类型为 `func(float64) float64`，参数是民用 JD |
| `DefaultTTMinusUTC()` | 取得内置闰秒表函数，不读取注入覆盖或未来政策 |

两个回调均返回秒。`SetDeltaT(nil)` 恢复默认 ΔT；`SetTTMinusUTC(nil)` 恢复内置闰秒表与未来政策。未设置 TT−UTC 覆盖时，`TTMinusUTC()` 返回 `nil`。

`basic` 中对应的接口是 `GetDeltaTFn` / `SetDeltaTFn`、`GetTTMinusUTCFn` / `SetTTMinusUTCFn`，与根包共享状态。TT−UTC 覆盖作用于 1972 年以后的查询，优先于未来政策；1972 年前仍采用本库的 UT1 约定。

这些设置会影响同一进程内的后续计算。应用应在开始计算前设定；临时对照结束后恢复原来的函数或命名模型，不要把模型切换当成某一次函数调用的局部选项。

## 未来的 UTC 对齐口径假设

> **2027-05-20 起 UTC 不再插入闰秒（闰时取代闰秒）**
>
> 2026 年 10 月 15 日，第 28 届国际计量大会（CGPM）通过[第 5 号决议《关于确保 UTC 连续性所需的技术行动》](https://www.bipm.org/en/cgpm-2026/resolution-5)（`Resolution-CGPM-28-5`，DOI `10.59161/CGPM2026RES5E`），决定**连续 UTC 自 2027 年 5 月 20 日起生效**，并把 `|UT1−UTC|` 的最大值由 0.9 秒放宽到 **3600 秒（1 小时）**。
>
> [IERS Bulletin C 72（2026-07-06）](https://datacenter.iers.org/data/latestVersion/bulletinC.txt) 已确认 **2026 年 12 月底不插入闰秒**；闰秒机制在 2027-05-20 终止，2027 年 6 月底这一可能的插入点同样不会有闰秒。
>
> 本库据此把内置窗口（截至 2026-10-01）之后的默认政策改为 `TimeScaleLeapHour`：1972–2026 仍按确定的闰秒表计算，历史值不变；窗口之后 TT−UTC 将长期保持 69.184 秒，直到数百年后 `|UT1−UTC|` 逼近 1 小时才需要闰时。农历现代快速表的 1900–2026 段也因此保持逐位不变。

`SetTimeScaleFuturePolicy` 与 `GetTimeScaleFuturePolicy` 选择民用时标换算政策，默认是 `TimeScaleLeapHour`。除 `TimeScaleUT1Civil` 替换全时轴外，其余政策只控制实测窗口之后的 TT−UTC；1972 年后显式注入的 TT−UTC 覆盖始终优先。`TimeScaleUTC` / `TimeScaleUT1` 只决定输出标签的口径，本政策决定民用时标与 TT 的换算，两者互不影响。

| 政策 | 计算假设 |
| --- | --- |
| `TimeScaleLeapHour`（零值、默认） | 按 ΔT 外推，以最少整小时校正使 DUT1 回到 ±3600 秒内，对应 2027-05-20 起生效的连续 UTC |
| `TimeScaleLeapSecond` | 从末端 TT−UTC 出发，以最少整数秒校正使外推 DUT1 回到 ±0.9 秒内；对应 1972–2026 的确定闰秒口径，仅供复现与对照 |
| `TimeScaleAssumeUT1Tracking` | 保留末端 DUT1，TT−UTC 随外推 ΔT 平滑变化 |
| `TimeScaleFreezeUTCOffset` | TT−UTC 固定在内置窗口末端值，当前为 69.184 秒 |
| `TimeScaleUT1Civil` | 全时轴将民用时间视为 UT1，含历史闰秒表覆盖期；无显式覆盖时 DUT1 恒为 0 |

默认政策 `TimeScaleLeapHour` 直接对应 2027-05-20 起生效的连续 UTC，不再是情景假设；`TimeScaleLeapSecond` 保留用于复现 1972–2026 的确定口径或作对照。整小时校正由查询时刻的 ΔT 决定，可能不落在公告的日历月边界；步进政策依赖的 ΔT 无效或超出模型范围时返回 `NaN`。

`TimeScaleLeapHour` 下 `|UT1−UTC|` 的容限是 3600 秒，因此窗口之后的 TT−UTC 在很长时间内恒为 69.184 秒；`TimeScaleLeapSecond` 仍按 ±0.9 秒外推，会给出每 1–2 年一次阶跃的旧行为，不代表现实。需要精确民用时刻时，应使用对应日期的已发布时标数据。

## SVG、GeoJSON 与 KML 的时间标签

SVG 和 GeoJSON 默认输出民用时间标签。需要 UT1 时，可使用 `...InUT1` 结果转换函数，或设置导出选项的 `TimeScale: astro.TimeScaleUT1`。UT1 输出的 `Location` 应为 `nil` 或 `time.UTC`。

已有几何换时标时仍对应同一个物理时刻。时间标记若按输出时标的整刻度重新取点，采样时刻会随之改变，标记位置也会移动。GeoJSON 用 `time_scale` 标明标签口径；KML 的 `<when>` 必须是 UTC，转换器会把 UT1 标签按当前模型换回 UTC，并保留原 `time_scale` 属性。生成 GeoJSON 和转成 KML 时应使用一致的时标模型。

具体选项、时间标记步长和 KML 时间轴行为见[地图与数据导出](map-geojson.md)。
