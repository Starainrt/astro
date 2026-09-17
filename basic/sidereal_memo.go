package basic

import (
	"math"
	"sync"
	"sync/atomic"
)

// 视恒星时是 UT 的纯函数，但月掩全球路径会在同一条计算链里反复向它求值：
// 地球自转、升落上下文、测地投影各自按自己的调用点重算同一个瞬时。实测单场 Saturn
// 2025-01-05 的请求里，191,517 次求值只对应 51,308 个不同的儒略日（重复距离中位数只有
// 2 次调用），而每次求值都要完整算一遍 77 项 IAU2000B 章动。
//
// 这里用一张有界直接映射表把结果记下来：无分配、容量固定（4096 槽 × 24 字节），
// 用 RWMutex 保证 C 共享库被宿主多线程调用时安全。表项记录写入时的 ΔT 世代，
// 因此 astro.SetDeltaT 覆盖之后旧条目自然失效，不会返回陈旧恒星时。
//
// Apparent sidereal time is a pure function of UT, yet one occultation path query evaluates
// it many times for the same instant from independent code paths (Earth rotation, rise/set
// contexts, geodetic projection). A single Saturn 2025-01-05 request performed 191,517
// evaluations for only 51,308 distinct Julian days, and each evaluation ran the full
// 77-term IAU2000B nutation. This bounded direct-mapped memo removes that redundancy without
// allocating: a fixed 4096-slot table guarded by an RWMutex, with the ΔT generation stored in
// each entry so an astro.SetDeltaT override invalidates stale values instead of replaying them.
// 4096 槽对单场月掩的 5 万余个不同儒略日而言明显偏小（重复距离中位数只有 2 次调用），
// 这里扩到 16384 槽（16384×24 B = 384 KB，BSS 静态数组，不参与初始化）。
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
// siderealMemoIndex mixes high and low bits so adjacent Julian days do not evict each other.
func siderealMemoIndex(jd float64) uint64 {
	bits := math.Float64bits(jd)
	return (bits ^ (bits >> 29)) & (siderealMemoSize - 1)
}

// siderealMemoLoad 返回缓存命中值；ΔT 世代不匹配时按未命中处理。
// siderealMemoLoad returns a cached value; a generation mismatch counts as a miss.
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

// siderealMemoStore 只在 ΔT 世代未变时写入：世代必须在**求值前**采样（见 siderealMemoGeneration），
// 否则求值期间发生的 SetDeltaTFn 会把旧 ΔT 的结果打上新世代并长期回放。
// siderealMemoStore writes only while the ΔT generation is unchanged; the generation must be sampled
// before the evaluation, otherwise a SetDeltaTFn during the computation would stamp the old value
// with the new generation and replay it.
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
