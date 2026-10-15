package eclipse

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
)

// Eclipse 指定星历实例的计算模块 / computation module bound to an ephemeris instance.
type Eclipse struct {
	inst *ephem.Instance
	// lunarLimb 是站心日食与月食的默认月缘剖面，nil 为圆月面。
	lunarLimb LunarLimb
}

// NewWithInstance 创建指定星历实例的模块，nil 使用全局模型 / create a module; nil uses the global model.
func NewWithInstance(inst *ephem.Instance) *Eclipse { return &Eclipse{inst: inst} }

// WithLunarLimb 返回使用指定月缘剖面的新模块，原模块不变；nil 恢复圆月面 /
// a copy of the module using the given lunar limb; nil restores the circular limb.
func (module *Eclipse) WithLunarLimb(limb LunarLimb) *Eclipse {
	copied := *module
	copied.lunarLimb = limb
	return &copied
}

// lunarResult 按模块星历与月缘计算一次月食；未设置月缘时与包级计算相同。
func (module *Eclipse) lunarResult(seedJDE float64, chauvenet bool) basic.LunarEclipseResult {
	result := basic.LunarEclipseDanjonWithInstance(seedJDE, module.inst)
	if chauvenet {
		result = basic.LunarEclipseChauvenetWithInstance(seedJDE, module.inst)
	}
	return basic.LunarEclipseWithLunarLimb(result, module.lunarLimb, module.inst)
}

// withLunarLimb 对已求出的全局月食按模块月缘重算接触时刻与接触点。
func (module *Eclipse) withLunarLimb(info LunarEclipseInfo) LunarEclipseInfo {
	if module.lunarLimb == nil || info.Type == LunarEclipseNone || info.Maximum.IsZero() {
		return info
	}
	result := module.lunarResult(timeToTTJDE(info.Maximum), basic.LunarEclipseShadowModel(info.ShadowModel) == basic.LunarEclipseShadowModelChauvenet)
	if result.Type == basic.LunarEclipseNone {
		return info
	}
	return lunarEclipseInfoFromBasicWithInstance(result, info.Maximum.Location(), module.inst)
}

var defaultModule Eclipse

func (module *Eclipse) model() ephem.Model {
	if module.inst == nil {
		return ephem.Active()
	}
	return module.inst.Model()
}

// ClosestLunarEclipse 最近一次月食 / closest lunar eclipse.
func (module *Eclipse) ClosestLunarEclipse(date time.Time) LunarEclipseInfo {
	return module.withLunarLimb(ClosestLunarEclipseWithInstance(date, module.inst))
}

// ClosestLunarEclipseChauvenet 最近一次月食（Chauvenet） / closest lunar eclipse with Chauvenet model.
func (module *Eclipse) ClosestLunarEclipseChauvenet(date time.Time) LunarEclipseInfo {
	return module.withLunarLimb(ClosestLunarEclipseChauvenetWithInstance(date, module.inst))
}

// ClosestLunarEclipseDanjon 最近一次月食（Danjon） / closest lunar eclipse with Danjon model.
func (module *Eclipse) ClosestLunarEclipseDanjon(date time.Time) LunarEclipseInfo {
	return module.withLunarLimb(ClosestLunarEclipseDanjonWithInstance(date, module.inst))
}

// LastLunarEclipse 上次月食 / previous lunar eclipse.
func (module *Eclipse) LastLunarEclipse(date time.Time) LunarEclipseInfo {
	return module.withLunarLimb(LastLunarEclipseWithInstance(date, module.inst))
}

// LastLunarEclipseChauvenet 上次月食（Chauvenet） / previous lunar eclipse with Chauvenet model.
func (module *Eclipse) LastLunarEclipseChauvenet(date time.Time) LunarEclipseInfo {
	return module.withLunarLimb(LastLunarEclipseChauvenetWithInstance(date, module.inst))
}

// LastLunarEclipseDanjon 上次月食（Danjon） / previous lunar eclipse with Danjon model.
func (module *Eclipse) LastLunarEclipseDanjon(date time.Time) LunarEclipseInfo {
	return module.withLunarLimb(LastLunarEclipseDanjonWithInstance(date, module.inst))
}

// LocalSolarEclipseInfoInUT1 返回各阶段与接触点换成 UT1 时刻的副本 / copy in UT1.
func (module *Eclipse) LocalSolarEclipseInfoInUT1(info LocalSolarEclipseInfo) LocalSolarEclipseInfo {
	return LocalSolarEclipseInfoInUT1(info)
}

// LunarEclipseInfoInUT1 返回各阶段与接触点换成 UT1 时刻的副本 / copy in UT1.
func (module *Eclipse) LunarEclipseInfoInUT1(info LunarEclipseInfo) LunarEclipseInfo {
	return LunarEclipseInfoInUT1(info)
}

// LunarEclipseOnDate 当地自然日月食查询 / local-date lunar eclipse query.
func (module *Eclipse) LunarEclipseOnDate(date time.Time) (LunarEclipseInfo, bool) {
	info, ok := LunarEclipseOnDateWithInstance(date, module.inst)
	return module.withLunarLimb(info), ok
}

// LunarEclipseOnDateChauvenet 当地自然日月食查询（Chauvenet） / local-date lunar eclipse query with Chauvenet model.
func (module *Eclipse) LunarEclipseOnDateChauvenet(date time.Time) (LunarEclipseInfo, bool) {
	info, ok := LunarEclipseOnDateChauvenetWithInstance(date, module.inst)
	return module.withLunarLimb(info), ok
}

// LunarEclipseOnDateDanjon 当地自然日月食查询（Danjon） / local-date lunar eclipse query with Danjon model.
func (module *Eclipse) LunarEclipseOnDateDanjon(date time.Time) (LunarEclipseInfo, bool) {
	info, ok := LunarEclipseOnDateDanjonWithInstance(date, module.inst)
	return module.withLunarLimb(info), ok
}

// NextLunarEclipse 下次月食 / next lunar eclipse.
func (module *Eclipse) NextLunarEclipse(date time.Time) LunarEclipseInfo {
	return module.withLunarLimb(NextLunarEclipseWithInstance(date, module.inst))
}

// NextLunarEclipseChauvenet 下次月食（Chauvenet） / next lunar eclipse with Chauvenet model.
func (module *Eclipse) NextLunarEclipseChauvenet(date time.Time) LunarEclipseInfo {
	return module.withLunarLimb(NextLunarEclipseChauvenetWithInstance(date, module.inst))
}

// NextLunarEclipseDanjon 下次月食（Danjon） / next lunar eclipse with Danjon model.
func (module *Eclipse) NextLunarEclipseDanjon(date time.Time) LunarEclipseInfo {
	return module.withLunarLimb(NextLunarEclipseDanjonWithInstance(date, module.inst))
}

// SolarEclipseBesselianMuForPublishedTable 把本库的 Mu 换算成与已发布根数表直接可比的取值 / convert Mu to the convention of published element tables.
func (module *Eclipse) SolarEclipseBesselianMuForPublishedTable(
	mu SolarEclipseBesselianPolynomial, deltaTSeconds float64,
) SolarEclipseBesselianPolynomial {
	return SolarEclipseBesselianMuForPublishedTable(mu, deltaTSeconds)
}

// SolarEclipseGeocentricPanelInUT1 返回两个朔时刻换成 UT1 时刻的副本 / copy in UT1.
func (module *Eclipse) SolarEclipseGeocentricPanelInUT1(panel SolarEclipseGeocentricPanel) SolarEclipseGeocentricPanel {
	return SolarEclipseGeocentricPanelInUT1(panel)
}

// SolarEclipseInfoInUT1 返回各阶段时刻换成 UT1 时刻的副本 / returns a copy in UT1.
func (module *Eclipse) SolarEclipseInfoInUT1(info SolarEclipseInfo) SolarEclipseInfo {
	return SolarEclipseInfoInUT1(info)
}

// SolarEclipsePartialFootprintsInUT1 返回足迹、轮廓与接触点都换成 UT1 时刻的副本 / copy in UT1.
func (module *Eclipse) SolarEclipsePartialFootprintsInUT1(info SolarEclipsePartialFootprintsInfo) SolarEclipsePartialFootprintsInfo {
	return SolarEclipsePartialFootprintsInUT1(info)
}

// SolarEclipsePathInUT1 返回中心线与限界都换成 UT1 时刻的副本 / copy in UT1.
func (module *Eclipse) SolarEclipsePathInUT1(path SolarEclipsePath) SolarEclipsePath {
	return SolarEclipsePathInUT1(path)
}

// TimeLabelsInUT1 返回换成 UT1 时刻的新切片 / returns a copy in UT1.
func (module *Eclipse) TimeLabelsInUT1(times []time.Time) []time.Time { return TimeLabelsInUT1(times) }
