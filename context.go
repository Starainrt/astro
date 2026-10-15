package astro

import (
	"sync"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/jupiter"
	"github.com/starainrt/astro/mars"
	"github.com/starainrt/astro/mercury"
	"github.com/starainrt/astro/moon"
	"github.com/starainrt/astro/neptune"
	"github.com/starainrt/astro/phenomena"
	"github.com/starainrt/astro/saturn"
	"github.com/starainrt/astro/sun"
	"github.com/starainrt/astro/uranus"
	"github.com/starainrt/astro/venus"
)

// Model 星历模型 / ephemeris model.
type Model = EphemerisModel

const (
	// ModelLegacy VSOP87 与 ELP2000 截断级数 / truncated VSOP87 and ELP2000 series.
	ModelLegacy = EphemerisLegacy
	// ModelSemianalytic 半解析表模型 / semi-analytic table model.
	ModelSemianalytic = EphemerisSemianalytic
	// ModelCustom 自定义只读引擎 / custom read-only engine.
	ModelCustom = EphemerisCustom
)

// Context 不可变星历实例与延迟创建的高层模块，须通过 NewContext 创建且不得复制；零值与 nil 没有引擎，
// 标量结果为 NaN，带输入校验的接口（如升落）返回 basic.ErrInvalidObservationInput /
// immutable ephemeris context with lazy modules; construct with NewContext and do not copy. The zero value and nil
// have no engine: scalar results are NaN and validated entry points such as rise/set return basic.ErrInvalidObservationInput.
type Context struct {
	inst *ephem.Instance
	// lunarLimb 日月食与月掩的默认月缘剖面，nil 为圆月面。
	lunarLimb eclipse.LunarLimb

	moonOnce      sync.Once
	moonObj       *moon.Moon
	sunOnce       sync.Once
	sunObj        *sun.Sun
	mercuryOnce   sync.Once
	mercuryObj    *mercury.Mercury
	venusOnce     sync.Once
	venusObj      *venus.Venus
	marsOnce      sync.Once
	marsObj       *mars.Mars
	jupiterOnce   sync.Once
	jupiterObj    *jupiter.Jupiter
	saturnOnce    sync.Once
	saturnObj     *saturn.Saturn
	uranusOnce    sync.Once
	uranusObj     *uranus.Uranus
	neptuneOnce   sync.Once
	neptuneObj    *neptune.Neptune
	eclipseOnce   sync.Once
	eclipseObj    *eclipse.Eclipse
	phenomenaOnce sync.Once
	phenomenaObj  *phenomena.Phenomena
	geojsonOnce   sync.Once
	geojsonObj    *geojson.Generator
}

// NewContext 创建独立模型上下文；内置模型不得带 Engine，ModelCustom 必须且只能带一个非 nil 只读 Engine，
// 违反时返回 ErrInvalidEngine，未知模型返回 ErrUnknownModel /
// create an independent context; built-in models take no Engine and ModelCustom takes exactly one non-nil Engine.
func NewContext(model Model, engines ...Engine) (*Context, error) {
	switch model {
	case ModelLegacy, ModelSemianalytic:
		if len(engines) != 0 {
			return nil, ErrInvalidEngine
		}
		return &Context{inst: ephem.NewModelInstance(model)}, nil
	case ModelCustom:
		if len(engines) != 1 || ephem.IsNilEngine(engines[0]) {
			return nil, ErrInvalidEngine
		}
		return &Context{inst: ephem.NewInstance(engines[0])}, nil
	default:
		return nil, ErrUnknownModel
	}
}

// NewStackContext 创建使用多源星历的独立上下文，先传入的引擎优先；含 nil 引擎时返回 ErrInvalidEngine /
// creates a multi-source context; earlier engines have priority and nil engines return ErrInvalidEngine.
func NewStackContext(engines ...Engine) (*Context, error) {
	for _, engine := range engines {
		if ephem.IsNilEngine(engine) {
			return nil, ErrInvalidEngine
		}
	}
	return &Context{inst: ephem.NewInstance(ephem.NewStack(engines...))}, nil
}

// Model 返回此上下文的固定模型 / the fixed model of this context.
func (context *Context) Model() Model {
	return context.instance().Model()
}

// emptyInstance 零值或 nil 上下文所用的无引擎实例。
var emptyInstance = ephem.NewInstance(nil)

// nilContext nil 接收者的访问器转到这个共享零值上下文。
var nilContext = &Context{}

func (context *Context) self() *Context {
	if context == nil {
		return nilContext
	}
	return context
}

func (context *Context) instance() *ephem.Instance {
	if context == nil || context.inst == nil {
		return emptyInstance
	}
	return context.inst
}

// Moon 返回月球模块 / the lunar module.
func (context *Context) Moon() *moon.Moon {
	context = context.self()
	context.moonOnce.Do(func() { context.moonObj = moon.NewWithInstance(context.instance()).WithLunarLimb(context.lunarLimb) })
	return context.moonObj
}

// Sun 返回太阳模块 / the solar module.
func (context *Context) Sun() *sun.Sun {
	context = context.self()
	context.sunOnce.Do(func() { context.sunObj = sun.NewWithInstance(context.instance()) })
	return context.sunObj
}

// Mercury 返回水星模块 / the Mercury module.
func (context *Context) Mercury() *mercury.Mercury {
	context = context.self()
	context.mercuryOnce.Do(func() { context.mercuryObj = mercury.NewWithInstance(context.instance()) })
	return context.mercuryObj
}

// Venus 返回金星模块 / the Venus module.
func (context *Context) Venus() *venus.Venus {
	context = context.self()
	context.venusOnce.Do(func() { context.venusObj = venus.NewWithInstance(context.instance()) })
	return context.venusObj
}

// Mars 返回火星模块 / the Mars module.
func (context *Context) Mars() *mars.Mars {
	context = context.self()
	context.marsOnce.Do(func() { context.marsObj = mars.NewWithInstance(context.instance()) })
	return context.marsObj
}

// Jupiter 返回木星模块 / the Jupiter module.
func (context *Context) Jupiter() *jupiter.Jupiter {
	context = context.self()
	context.jupiterOnce.Do(func() { context.jupiterObj = jupiter.NewWithInstance(context.instance()) })
	return context.jupiterObj
}

// Saturn 返回土星模块 / the Saturn module.
func (context *Context) Saturn() *saturn.Saturn {
	context = context.self()
	context.saturnOnce.Do(func() { context.saturnObj = saturn.NewWithInstance(context.instance()) })
	return context.saturnObj
}

// Uranus 返回天王星模块 / the Uranus module.
func (context *Context) Uranus() *uranus.Uranus {
	context = context.self()
	context.uranusOnce.Do(func() { context.uranusObj = uranus.NewWithInstance(context.instance()) })
	return context.uranusObj
}

// Neptune 返回海王星模块 / the Neptune module.
func (context *Context) Neptune() *neptune.Neptune {
	context = context.self()
	context.neptuneOnce.Do(func() { context.neptuneObj = neptune.NewWithInstance(context.instance()) })
	return context.neptuneObj
}

// Eclipse 返回日月食模块 / the eclipse module.
func (context *Context) Eclipse() *eclipse.Eclipse {
	context = context.self()
	context.eclipseOnce.Do(func() {
		context.eclipseObj = eclipse.NewWithInstance(context.instance()).WithLunarLimb(context.lunarLimb)
	})
	return context.eclipseObj
}

// WithLunarLimb 共享星历实例、日月食与月掩改用指定月缘剖面的新上下文，nil 为圆月面 / context using the given lunar limb.
func (context *Context) WithLunarLimb(limb eclipse.LunarLimb) *Context {
	return &Context{inst: context.instance(), lunarLimb: limb}
}

// GeoJSON 返回地理数据生成模块 / the geographic data generator.
func (context *Context) GeoJSON() *geojson.Generator {
	context = context.self()
	context.geojsonOnce.Do(func() { context.geojsonObj = geojson.NewWithInstance(context.instance()) })
	return context.geojsonObj
}

var globalMu sync.RWMutex

// SetModel 切换包级函数的全局内置模型，非内置模型返回 ErrUnknownModel 且不改动全局状态；已有显式上下文不受影响 /
// switch the global built-in model; non-built-in models return ErrUnknownModel and leave the global state unchanged.
func SetModel(model Model) error {
	if model != ModelLegacy && model != ModelSemianalytic {
		return ErrUnknownModel
	}
	inst := ephem.NewModelInstance(model)
	globalMu.Lock()
	defer globalMu.Unlock()
	basic.SetEphemerisModel(basic.EphemerisModel(model))
	ephem.SetGlobal(inst)
	return nil
}

// GetModel 返回包级函数当前使用的模型 / the active model used by package-level functions.
func GetModel() Model {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return Model(basic.GetEphemerisModel())
}

// Phenomena 返回通用天象与偕日可见模块 / the generic sky event module.
func (context *Context) Phenomena() *phenomena.Phenomena {
	context = context.self()
	context.phenomenaOnce.Do(func() { context.phenomenaObj = phenomena.NewWithInstance(context.instance()) })
	return context.phenomenaObj
}
