// Package ephem 全局星历模型状态与实例接口 / global ephemeris model state and instance interfaces.
package ephem

import (
	"sync"
	"sync/atomic"
)

// Model 星历模型 / ephemeris model.
type Model int32

const (
	// Legacy VSOP87 + ELP2000 截断级数 / truncated VSOP87 and ELP2000 series.
	Legacy Model = iota
	// Semianalytic 按 DE441 重拟的半解析表 / semi-analytic tables refitted to DE441.
	Semianalytic
	// Custom 外部 Engine 实例，不对应内置模型 / external Engine instance without a built-in model.
	Custom Model = -1
)

var (
	active                = int32(BuildDefault)
	generation     uint64 = 1
	modelEngines          = make(map[Model]func() Engine)
	modelEnginesMu sync.RWMutex
)

// Active 当前模型 / the active model.
func Active() Model { return Model(atomic.LoadInt32(&active)) }

// Generation 模型切换世代，起始为 1 / model switch generation, starting at 1.
func Generation() uint64 { return atomic.LoadUint64(&generation) }

// SetActive 切换模型，返回是否发生变化 / switches the model and reports whether it changed.
func SetActive(m Model) bool {
	if m != Legacy && m != Semianalytic {
		return false
	}
	globalInstanceMu.Lock()
	defer globalInstanceMu.Unlock()
	if atomic.SwapInt32(&active, int32(m)) == int32(m) {
		return false
	}
	atomic.AddUint64(&generation, 1)
	globalInstance.Store(NewModelInstance(m))
	return true
}

// RegisterModelEngine 注册内置模型引擎工厂 / register a built-in model engine factory.
func RegisterModelEngine(m Model, factory func() Engine) {
	if m != Legacy && m != Semianalytic || factory == nil {
		return
	}
	modelEnginesMu.Lock()
	modelEngines[m] = factory
	modelEnginesMu.Unlock()
}

// Instance 星历实例，可包含配置和缓存 / ephemeris instance with configuration and cache.
type Instance struct {
	model  Model
	engine Engine
}

// NewInstance 创建自定义星历实例 / create an instance from a custom engine.
func NewInstance(engine Engine) *Instance {
	if IsNilEngine(engine) {
		engine = nil
	}
	return &Instance{model: Custom, engine: engine}
}

// NewModelInstance 创建内置模型实例 / create an instance for a built-in model.
func NewModelInstance(m Model) *Instance {
	if m != Legacy && m != Semianalytic {
		m = BuildDefault
	}
	modelEnginesMu.RLock()
	factory := modelEngines[m]
	modelEnginesMu.RUnlock()
	var engine Engine
	if factory != nil {
		engine = factory()
	}
	return &Instance{model: m, engine: engine}
}

// Model 获取实例使用的模型 / get model used by this instance.
func (inst *Instance) Model() Model {
	if inst == nil {
		return Custom
	}
	return inst.model
}

// Engine 获取底层星历引擎 / get the underlying ephemeris engine.
func (inst *Instance) Engine() Engine {
	if inst == nil {
		return nil
	}
	return inst.engine
}

// State 计算目标相对中心的状态 / calculate target state relative to center.
func (inst *Instance) State(target, center Body, jdTDB float64) (pos, vel [3]float64, err error) {
	if inst == nil || inst.engine == nil {
		return pos, vel, ErrNoEngine
	}
	return inst.engine.State(target, center, jdTDB)
}

// Coverage 返回目标覆盖范围 / return target coverage.
func (inst *Instance) Coverage(target Body) (start, end float64, ok bool) {
	if inst == nil || inst.engine == nil {
		return 0, 0, false
	}
	return inst.engine.Coverage(target)
}

// Apparent 计算地心视方向 / calculate apparent geocentric direction.
func (inst *Instance) Apparent(target Body, jdTT float64) (dir [3]float64, distKM, tauDays float64, err error) {
	if inst == nil || inst.engine == nil {
		return dir, 0, 0, ErrNoEngine
	}
	if engine, ok := inst.engine.(ApparentEngine); ok {
		return engine.Apparent(target, jdTT)
	}
	// 只给几何状态的引擎（如 SPK）走通用矢量链。
	return genericApparent(inst.engine, target, jdTT)
}

var (
	globalInstance   atomic.Value
	globalInstanceMu sync.Mutex
)

// Global 全局默认星历实例 / global default ephemeris instance.
func Global() *Instance {
	if inst, ok := globalInstance.Load().(*Instance); ok && inst != nil {
		return inst
	}
	globalInstanceMu.Lock()
	defer globalInstanceMu.Unlock()
	if inst, ok := globalInstance.Load().(*Instance); ok && inst != nil {
		return inst
	}
	inst := NewModelInstance(Active())
	globalInstance.Store(inst)
	return inst
}

// SetGlobal 设置全局默认星历实例 / set global default ephemeris instance.
func SetGlobal(inst *Instance) {
	if inst == nil {
		return
	}
	globalInstanceMu.Lock()
	defer globalInstanceMu.Unlock()
	globalInstance.Store(inst)
}
