// Package planet 提供内置星历模型的低层入口：半解析表（默认）与 VSOP87/ELP2000-82 旧级数；jde 为 TT 儒略日，角度单位为度，距离单位为 AU。
// Package planet provides the low-level entry points of the built-in ephemeris models: the semianalytic tables (default) and the VSOP87/ELP2000-82 legacy series; jde is a TT Julian day, angles are degrees and distances are AU.
// 求值接口的 N 后缀是 n 变体：n < 0 取完整级数；半解析模型下所有 n >= 0 使用同一张快速表；Legacy 按 n 截断。
// The N suffix marks an n-variant: n < 0 uses the full series, every n >= 0 uses one fast table under the semianalytic model, and Legacy truncates by n.
package planet
