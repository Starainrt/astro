package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/ephem"
)

// Moon 指定星历实例的计算模块 / computation module bound to an ephemeris instance.
type Moon struct {
	inst *ephem.Instance
	// lunarLimb 是月掩搜索的默认月缘剖面，nil 为圆月面。
	lunarLimb basic.LunarLimb
}

// NewWithInstance 创建指定星历实例的模块，nil 使用全局模型 / create a module; nil uses the global model.
func NewWithInstance(inst *ephem.Instance) *Moon { return &Moon{inst: inst} }

// WithLunarLimb 返回使用指定月缘剖面的副本，选项中未设置月缘时采用它 / copy with a default lunar limb for occultations.
func (module *Moon) WithLunarLimb(limb basic.LunarLimb) *Moon {
	copied := *module
	copied.lunarLimb = limb
	return &copied
}

// occultationOptions 在选项未设置月缘时补上模块默认值。
func (module *Moon) occultationOptions(options OccultationSearchOptions) OccultationSearchOptions {
	if options.LunarLimb == nil {
		options.LunarLimb = module.lunarLimb
	}
	return options
}

var defaultModule Moon

// EarthDistance 地月距离 / Earth-Moon distance.
func (module *Moon) EarthDistance(date time.Time) float64 {
	jde := basic.UTC2TT(basic.Date2JD(date))
	model := ephem.Active()
	if module.inst != nil {
		model = module.inst.Model()
	}
	if model == ephem.Legacy {
		return basic.MoonAway(jde)
	}
	return basic.HMoonAwayWithInstance(jde, module.inst)
}

// PlanetOccultationInfoInUT1 返回固定地点行星月掩的 UT1 时刻副本 / copy in UT1 of a local planetary result.
func (module *Moon) PlanetOccultationInfoInUT1(info basic.PlanetOccultationInfo) basic.PlanetOccultationInfo {
	return PlanetOccultationInfoInUT1(info)
}

// PlanetOccultationPathInUT1 返回行星月掩全球掩带的 UT1 时刻副本 / copy in UT1 of a planetary path.
func (module *Moon) PlanetOccultationPathInUT1(path basic.PlanetOccultationPath) basic.PlanetOccultationPath {
	return PlanetOccultationPathInUT1(path)
}

// StarCoordinateFromStarData 将星表记录转换为恒星月掩搜索接受的 J2000 输入；不会加载星表 / convert a catalog record to a J2000 occultation input.
func (module *Moon) StarCoordinateFromStarData(star StarData) (StarCoordinate, error) {
	return StarCoordinateFromStarData(star)
}

// StarOccultationInfoInUT1 返回固定地点恒星月掩的 UT1 时刻副本 / copy in UT1 of a local stellar result.
func (module *Moon) StarOccultationInfoInUT1(info basic.StarOccultationInfo) basic.StarOccultationInfo {
	return StarOccultationInfoInUT1(info)
}

// StarOccultationPathInUT1 返回恒星月掩全球路径的 UT1 时刻副本 / copy in UT1 of a stellar path.
func (module *Moon) StarOccultationPathInUT1(path basic.StarOccultationPath) basic.StarOccultationPath {
	return StarOccultationPathInUT1(path)
}
