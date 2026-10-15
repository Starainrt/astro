package moon

import (
	"github.com/starainrt/astro/basic"
)

// StarOccultationDiagramOptions 控制本地恒星月掩轨迹采样。
// StarOccultationDiagramOptions controls local stellar-occultation track sampling.
type StarOccultationDiagramOptions = basic.StarOccultationDiagramOptions

// StarOccultationDiagramFrame 描述一个站心月球与恒星几何采样点。
// StarOccultationDiagramFrame describes one topocentric Moon-star geometry sample.
type StarOccultationDiagramFrame = basic.StarOccultationDiagramFrame

// StarOccultationDiagramResult 包含固定地点的月心恒星轨迹。
// StarOccultationDiagramResult contains a fixed-site Moon-centered stellar track.
type StarOccultationDiagramResult = basic.StarOccultationDiagramResult

// StarOccultationDiagram 为已求解的固定地点恒星月掩计算图形几何数据 / diagram geometry for a solved local stellar occultation.
func (module *Moon) StarOccultationDiagram(
	info StarOccultationInfo,
	star StarCoordinate,
	options StarOccultationDiagramOptions,
) StarOccultationDiagramResult {
	return basic.StarOccultationDiagramWithInstance(info, star, options, module.inst)
}

// LunarLimb 月缘剖面，nil 为圆月面，内置数据表在 moon/limb / lunar limb profile; nil means a circular limb.
type LunarLimb = basic.LunarLimb

// StarOccultationLimbProfileOptions 控制精密恒星月掩的星迹采样 / sampling of a precise stellar occultation.
type StarOccultationLimbProfileOptions = basic.StarOccultationLimbProfileOptions

// StarOccultationLimbSample 星迹上一个时刻的月缘几何 / limb geometry at one instant on the star track.
type StarOccultationLimbSample = basic.StarOccultationLimbSample

// OccultationInterval 一段连续遮挡 / one continuous hidden interval.
type OccultationInterval = basic.OccultationInterval

// StarOccultationLimbProfileResult 一次恒星月掩沿星迹的月缘剖面与全部遮挡区间 / limb profile and every hidden interval of a stellar occultation.
type StarOccultationLimbProfileResult = basic.StarOccultationLimbProfileResult

// StarOccultationLimbProfile 精密恒星月掩；limb 为 nil 时用模块月缘 / precise stellar occultation; nil limb uses the module profile.
func (module *Moon) StarOccultationLimbProfile(info StarOccultationInfo, star StarCoordinate, limb LunarLimb, options StarOccultationLimbProfileOptions) (StarOccultationLimbProfileResult, error) {
	if limb == nil {
		limb = module.lunarLimb
	}
	return basic.StarOccultationLimbProfileWithInstance(info, star, limb, options, module.inst)
}

// StarOccultationLimbProfileInUT1 返回精密恒星月掩的 UT1 时刻副本 / copy in UT1 of a precise stellar occultation.
func (module *Moon) StarOccultationLimbProfileInUT1(profile StarOccultationLimbProfileResult) StarOccultationLimbProfileResult {
	return StarOccultationLimbProfileInUT1(profile)
}
