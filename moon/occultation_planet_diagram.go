package moon

import (
	"github.com/starainrt/astro/basic"
)

// PlanetOccultationDiagramOptions 控制本地行星月掩轨迹采样。
// PlanetOccultationDiagramOptions controls local planetary-occultation track sampling.
type PlanetOccultationDiagramOptions = basic.PlanetOccultationDiagramOptions

// PlanetOccultationDiagramFrame 描述一个站心月球与行星几何采样点。
// PlanetOccultationDiagramFrame describes one topocentric Moon-planet geometry sample.
type PlanetOccultationDiagramFrame = basic.PlanetOccultationDiagramFrame

// PlanetOccultationDiagramResult 包含固定地点的月心行星轨迹。
// PlanetOccultationDiagramResult contains a fixed-site Moon-centered planetary track.
type PlanetOccultationDiagramResult = basic.PlanetOccultationDiagramResult

// PlanetOccultationDiagram 为已求解的固定地点行星月掩计算图形几何数据 / diagram geometry for a solved local planetary occultation.
func (module *Moon) PlanetOccultationDiagram(
	info PlanetOccultationInfo,
	options PlanetOccultationDiagramOptions,
) PlanetOccultationDiagramResult {
	return basic.PlanetOccultationDiagramWithInstance(info, options, module.inst)
}
