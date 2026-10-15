// Package basic 提供底层天文算法、星历实例入口与内置模型口径。
// Package basic provides the low-level astronomical algorithms, ephemeris-instance entry points and built-in model conventions.
// 求值接口的 N 后缀是 n 变体：n < 0 取完整级数；半解析模型下所有 n >= 0 使用同一张快速表；Legacy 按 n 截断。
// The N suffix marks an n-variant: n < 0 uses the full series, every n >= 0 uses one fast table under the semianalytic model, and Legacy truncates by n.
package basic
