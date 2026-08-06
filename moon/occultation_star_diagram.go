package moon

import "github.com/starainrt/astro/basic"

// StarOccultationDiagramOptions 控制本地恒星月掩轨迹采样。
// StarOccultationDiagramOptions controls local stellar-occultation track sampling.
type StarOccultationDiagramOptions = basic.StarOccultationDiagramOptions

// StarOccultationDiagramFrame 描述一个站心月球与恒星几何采样点。
// StarOccultationDiagramFrame describes one topocentric Moon-star geometry sample.
type StarOccultationDiagramFrame = basic.StarOccultationDiagramFrame

// StarOccultationDiagramResult 包含固定地点的月心恒星轨迹。
// StarOccultationDiagramResult contains a fixed-site Moon-centered stellar track.
type StarOccultationDiagramResult = basic.StarOccultationDiagramResult

// StarOccultationDiagram 为已求解的固定地点恒星月掩计算图形几何数据。
// StarOccultationDiagram computes diagram geometry for an already solved fixed-site stellar occultation.
func StarOccultationDiagram(
	info StarOccultationInfo,
	star StarCoordinate,
	options StarOccultationDiagramOptions,
) StarOccultationDiagramResult {
	return basic.StarOccultationDiagram(info, star, options)
}
