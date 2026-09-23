package basic

import (
	"math"
	"sync"
	"sync/atomic"
)

// 固定容量缓存以 ΔT 世代区分模型，RWMutex 保护宿主多线程访问。
const siderealMemoBits = 14

const siderealMemoSize = 1 << siderealMemoBits

type siderealMemoEntry struct {
	key        uint64
	value      float64
	generation uint64
}

var (
	siderealMemoMu     sync.RWMutex
	siderealMemoTable  [siderealMemoSize]siderealMemoEntry
	siderealMemoHits   uint64
	siderealMemoMisses uint64
	siderealMemoStores uint64
)

// siderealMemoIndex 用高低位混合避免相邻儒略日落在相邻槽位而互相驱逐。
func siderealMemoIndex(jd float64) uint64 {
	bits := math.Float64bits(jd)
	return (bits ^ (bits >> 29)) & (siderealMemoSize - 1)
}

// siderealMemoLoad 返回缓存命中值；ΔT 世代不匹配时按未命中处理。
func siderealMemoLoad(jd float64) (float64, bool) {
	generation := deltaTGenerationValue()
	entry := &siderealMemoTable[siderealMemoIndex(jd)]
	siderealMemoMu.RLock()
	key, value, entryGeneration := entry.key, entry.value, entry.generation
	siderealMemoMu.RUnlock()
	if entryGeneration == generation && key == math.Float64bits(jd) {
		atomic.AddUint64(&siderealMemoHits, 1)
		return value, true
	}
	atomic.AddUint64(&siderealMemoMisses, 1)
	return 0, false
}

// 世代须在求值前采样，防止旧模型结果被缓存为新模型结果。
func siderealMemoStore(jd, value float64, generation uint64) {
	if deltaTGenerationValue() != generation {
		return
	}
	index := siderealMemoIndex(jd)
	siderealMemoMu.Lock()
	siderealMemoTable[index] = siderealMemoEntry{
		key:        math.Float64bits(jd),
		value:      value,
		generation: generation,
	}
	siderealMemoMu.Unlock()
	atomic.AddUint64(&siderealMemoStores, 1)
}

// siderealMemoGeneration 在求值前采样 ΔT 世代，供 siderealMemoStore 校验。
func siderealMemoGeneration() uint64 {
	return deltaTGenerationValue()
}

// siderealMemoStats 返回命中/未命中/写入计数，供测试守护记忆表确实生效。
func siderealMemoStats() (hits, misses, stores uint64) {
	return atomic.LoadUint64(&siderealMemoHits), atomic.LoadUint64(&siderealMemoMisses),
		atomic.LoadUint64(&siderealMemoStores)
}

func resetSiderealMemo() {
	siderealMemoMu.Lock()
	for i := range siderealMemoTable {
		siderealMemoTable[i] = siderealMemoEntry{}
	}
	siderealMemoMu.Unlock()
	atomic.StoreUint64(&siderealMemoHits, 0)
	atomic.StoreUint64(&siderealMemoMisses, 0)
	atomic.StoreUint64(&siderealMemoStores, 0)
}
