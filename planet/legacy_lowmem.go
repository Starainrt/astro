//go:build astro_lowmem
// +build astro_lowmem

package planet

import (
	_ "embed"
	"sync"
)

// lowmem 下旧级数系数以只读串常驻，首次用到旧模型时才解码进堆（约 240 KB）。
//
//go:embed legacy.bin
var legacyTableBlob string

var (
	legacyOnce sync.Once
	legacyRaw  [][]float64
	legacyCir  [][][]float64
)

func loadLegacyTables() {
	legacyOnce.Do(func() {
		raw, cir, err := decodeLegacyTables(legacyTableBlob)
		if err != nil {
			panic(err)
		}
		legacyRaw, legacyCir = raw, cir
	})
}

func planetRawTable() [][]float64 {
	loadLegacyTables()
	return legacyRaw
}

func moonCirTable() [][][]float64 {
	loadLegacyTables()
	return legacyCir
}
