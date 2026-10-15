//go:build astro_legacy
// +build astro_legacy

package ephem

// BuildDefault 本次构建的初始模型 / the model active at program start.
//
// 构建标签 astro_legacy 让全库以旧级数模型为默认运行。
const BuildDefault = Legacy
