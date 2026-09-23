// Package astro 提供进程级时标配置与时标换算 / process-wide time-scale configuration and conversions.
package astro

import "github.com/starainrt/astro/basic"

// DeltaT 返回当前的 TT−UT1（ΔT）模型 / the active TT−UT1 (DeltaT) model.
func DeltaT() func(float64, bool) float64 {
	return basic.GetDeltaTFn()
}

// SetDeltaT 替换 ΔT 模型（入参为数值与“该数值是否为儒略日”，为假时是十进制年），传 nil 恢复默认 / replaces the DeltaT model; nil restores the default.
func SetDeltaT(deltaT func(float64, bool) float64) {
	basic.SetDeltaTFn(deltaT)
}

// DefaultDeltaT 返回内置默认 ΔT 模型 / the built-in default DeltaT model.
func DefaultDeltaT() func(float64, bool) float64 {
	return basic.DefaultDeltaTv2
}

// TTMinusUTC 返回当前 TT−UTC 覆盖函数，未设置时为 nil / current TT−UTC override, or nil.
func TTMinusUTC() func(float64) float64 {
	return basic.GetTTMinusUTCFn()
}

// SetTTMinusUTC 替换 TT−UTC 模型，nil 恢复内置表与政策 / overrides TT−UTC; nil restores the built-in table and policy.
func SetTTMinusUTC(ttMinusUTC func(float64) float64) {
	basic.SetTTMinusUTCFn(ttMinusUTC)
}

// DefaultTTMinusUTC 返回内置闰秒表模型，忽略覆盖与未来政策 / built-in leap-table model, ignoring overrides and future policy.
func DefaultTTMinusUTC() func(float64) float64 {
	return basic.TTMinusUTCSecondsDefault
}

// TimeScaleFuturePolicy 民用时标换算政策，与输出用的 TimeScale 无关 / civil-time conversion policy, separate from output TimeScale.
type TimeScaleFuturePolicy = basic.TimeScaleFuturePolicy

const (
	// TimeScaleLeapSecond 默认按 ΔT 越限施加整数秒校正，不代表闰秒公告 / default integer-second corrections driven by extrapolated ΔT, not announcements.
	TimeScaleLeapSecond = basic.TimeScaleLeapSecond
	// TimeScaleAssumeUT1Tracking 窗口外固定末端 DUT1，平滑跟随 UT1 / holds the last observed DUT1 beyond the window.
	TimeScaleAssumeUT1Tracking = basic.TimeScaleAssumeUT1Tracking
	// TimeScaleFreezeUTCOffset 假设 UTC 偏移冻结在精确窗口末端 / freezes the UTC offset at the window end.
	TimeScaleFreezeUTCOffset = basic.TimeScaleFreezeUTCOffset
	// TimeScaleLeapHour 按 ΔT 越限施加整小时校正，仅作情景演算 / applies hour-sized corrections as a scenario assumption.
	TimeScaleLeapHour = basic.TimeScaleLeapHour
	// TimeScaleUT1Civil 全时轴民用时标等同 UT1，TT−UTC 覆盖仍优先 / uses UT1 as civil time everywhere unless TT−UTC is overridden.
	TimeScaleUT1Civil = basic.TimeScaleUT1Civil
)

// SetTimeScaleFuturePolicy 设置与 basic 共享的民用时标换算政策 / sets the civil-time conversion policy shared with basic.
func SetTimeScaleFuturePolicy(policy TimeScaleFuturePolicy) {
	basic.SetTimeScaleFuturePolicy(policy)
}

// GetTimeScaleFuturePolicy 返回当前民用时标换算政策 / current civil-time conversion policy.
func GetTimeScaleFuturePolicy() TimeScaleFuturePolicy {
	return basic.GetTimeScaleFuturePolicy()
}

// DeltaTModel 标识可选的已发布 ΔT 模型 / a selectable published ΔT model.
type DeltaTModel = basic.DeltaTModel

const (
	// DeltaTModelDefault 内置默认：SMH2016 + Morrison 2021 / the built-in SMH2016 + Morrison 2021 model.
	DeltaTModelDefault = basic.DeltaTModelDefault
	// DeltaTModelSMH2016 显式选择内置模型 / selects the built-in model explicitly.
	DeltaTModelSMH2016 = basic.DeltaTModelSMH2016
	// DeltaTModelMS2004 Morrison & Stephenson 2004 长期抛物线 / the M&S2004 long-term parabola.
	DeltaTModelMS2004 = basic.DeltaTModelMS2004
	// DeltaTModelEspenakMeeus2006 Espenak & Meeus 2006 分段多项式 / the Espenak & Meeus 2006 piecewise polynomial.
	DeltaTModelEspenakMeeus2006 = basic.DeltaTModelEspenakMeeus2006
	// DeltaTModelNASACanon2006 上式再加配对修正 c，即 NASA 五千年目录的印刷口径 / plus the pairing term c, the NASA canon convention.
	DeltaTModelNASACanon2006 = basic.DeltaTModelNASACanon2006
	// DeltaTModelManual 表示当前生效的是 SetDeltaT 注入的任意函数 / an arbitrary function injected with SetDeltaT.
	DeltaTModelManual = basic.DeltaTModelManual
)

// SetDeltaTModel 安装命名 ΔT 模型；keepObserved 为真时逐月实测段优先（推荐），未知模型返回 false 且不改动现状。
// SetDeltaTModel installs a named ΔT model; with keepObserved the monthly observed span wins wherever it covers the instant.
func SetDeltaTModel(model DeltaTModel, keepObserved bool) bool {
	return basic.SetDeltaTModel(model, keepObserved)
}

// GetDeltaTModel 返回当前生效的模型与是否保留实测段 / the active model and whether the observed span is kept.
func GetDeltaTModel() (DeltaTModel, bool) {
	return basic.GetDeltaTModel()
}

// DeltaTModelSeconds 按命名模型求 ΔT（秒，入参为 UT 儒略日），不改变进程状态，便于对照 / evaluates a named model without touching process state.
func DeltaTModelSeconds(model DeltaTModel, jd float64, keepObserved bool) float64 {
	return basic.DeltaTModelSeconds(model, jd, keepObserved)
}
