package basic

import (
	"math"
	"sync"
	"sync/atomic"
)

// Nutation2000B 是 jd 的纯函数，但升落、星历与掩星路径会对同一批瞬时反复求值：一条计算链里
// 地球自转、状态上下文与几何装配各自求一次同一个瞬时。这里用有界直接映射表记住结果：
// 槽位固定、无分配、RWMutex 保证 c-shared 宿主多线程安全。章动只依赖传入的 jd，与 ΔT 世代无关。
// Nutation2000B is a pure function of jd, yet the rise/set, ephemeris and occultation chains evaluate
// it repeatedly for the same instants from independent code paths. This bounded direct-mapped memo
// removes that redundancy with fixed slots, no allocation and an RWMutex for the c-shared host.
// Nutation depends only on its argument, so no ΔT generation stamp is needed.
const nutationMemoBits = 13

const nutationMemoSize = 1 << nutationMemoBits

type nutationMemoEntry struct {
	// key 是 math.Float64bits(jd)+1，0 表示空槽（避免 jd=0 与空槽同码）。
	key        uint64
	dpsi, deps float64
}

var (
	nutationMemoMu    sync.RWMutex
	nutationMemoTable [nutationMemoSize]nutationMemoEntry
	nutationMemoHits  uint64
	nutationMemoMiss  uint64
)

func nutationMemoIndex(jd float64) uint64 {
	bits := math.Float64bits(jd)
	return (bits ^ (bits >> 31)) & (nutationMemoSize - 1)
}

func nutationMemoLoad(jd float64) (float64, float64, bool) {
	key := math.Float64bits(jd) + 1
	entry := &nutationMemoTable[nutationMemoIndex(jd)]
	nutationMemoMu.RLock()
	entryKey, dpsi, deps := entry.key, entry.dpsi, entry.deps
	nutationMemoMu.RUnlock()
	if key != 0 && entryKey == key {
		atomic.AddUint64(&nutationMemoHits, 1)
		return dpsi, deps, true
	}
	atomic.AddUint64(&nutationMemoMiss, 1)
	return 0, 0, false
}

func nutationMemoStore(jd, dpsi, deps float64) {
	key := math.Float64bits(jd) + 1
	if key == 0 {
		return
	}
	nutationMemoMu.Lock()
	nutationMemoTable[nutationMemoIndex(jd)] = nutationMemoEntry{key: key, dpsi: dpsi, deps: deps}
	nutationMemoMu.Unlock()
}

// nutationMemoStats 返回命中/未命中计数，供测试守护命中率。
func nutationMemoStats() (uint64, uint64) {
	return atomic.LoadUint64(&nutationMemoHits), atomic.LoadUint64(&nutationMemoMiss)
}

func resetNutationMemo() {
	nutationMemoMu.Lock()
	for i := range nutationMemoTable {
		nutationMemoTable[i] = nutationMemoEntry{}
	}
	nutationMemoMu.Unlock()
	atomic.StoreUint64(&nutationMemoHits, 0)
	atomic.StoreUint64(&nutationMemoMiss, 0)
}
