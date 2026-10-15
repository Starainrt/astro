package planets

import (
	_ "embed"
	"sync"
)

// 现代段为 J2000 ± 1000 儒略年。
const (
	ModernStart = 2086295.0
	ModernEnd   = 2816795.0
)

var (
	//go:embed long.bin
	longTable []byte
	//go:embed modern.bin
	modernTable []byte
)

var (
	defaultOnce  sync.Once
	defaultModel *Model
)

// Default 内置长时段表叠加现代段，首次调用时解析 / built-in long-range plus modern tables, parsed on first use.
func Default() *Model {
	defaultOnce.Do(func() {
		m, err := Parse(longTable)
		if err == nil {
			err = m.WithModernSegment(modernTable, ModernStart, ModernEnd)
		}
		if err != nil {
			panic(err)
		}
		defaultModel = m
	})
	return defaultModel
}

// GM 天体号 1..8 的 GM 与太阳 GM，au³/日²；EMB 为地月合计，越界返回 0 / GM of body id and of the Sun, au³/day².
func GM(id int) (body, sun float64) {
	if id < 1 || id > 8 {
		return 0, gmSol
	}
	return gmp[id-1], gmSol
}

//go:embed fast.bin
var fastTable []byte

var (
	fastOnce  sync.Once
	fastModel *Model
)

// Fast 快速档表（n ≥ 0 用），首次调用时解析 / the fast-tier table used for n ≥ 0, parsed on first use.
func Fast() *Model {
	fastOnce.Do(func() {
		m, err := Parse(fastTable)
		if err != nil {
			panic(err)
		}
		fastModel = m
	})
	return fastModel
}
