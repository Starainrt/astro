// Package moon 月球视位置、升落、月相、天平动、视直径与物理观测参数；角度单位为度。
// Package moon covers lunar apparent place, rise/set, phase, libration, apparent diameter and physical observing parameters; angles are degrees.
// 求值接口的 N 后缀是 n 变体：n < 0 取完整级数；半解析模型下所有 n >= 0 使用同一张快速表；Legacy 按 n 截断。
// The N suffix marks an n-variant: n < 0 uses the full series, every n >= 0 uses one fast table under the semianalytic model, and Legacy truncates by n.
package moon
