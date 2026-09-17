// Package astro 进程级 ΔT 模型的读取、替换与恢复默认：DeltaT、SetDeltaT、DefaultDeltaT。
// Package astro exposes the process-wide DeltaT model: read the current function, install a custom one, or restore the built-in default.
package astro

import "github.com/starainrt/astro/basic"

func DeltaT() func(float64, bool) float64 {
	return basic.GetDeltaTFn()
}

func SetDeltaT(deltaT func(float64, bool) float64) {
	basic.SetDeltaTFn(deltaT)
}

func DefaultDeltaT() func(float64, bool) float64 {
	return basic.DefaultDeltaTv2
}
