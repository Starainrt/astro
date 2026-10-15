package lunar

import (
	_ "embed"
	"sync"
)

//go:embed lunar.bin
var defaultTable []byte

var (
	defaultOnce  sync.Once
	defaultModel *Model
)

// Default 内置表，首次调用时解析 / the built-in table, parsed on first use.
func Default() *Model {
	defaultOnce.Do(func() {
		m, err := Parse(defaultTable)
		if err != nil {
			panic(err)
		}
		defaultModel = m
	})
	return defaultModel
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
