package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// FindPlanetOccultationPaths 搜索窗口内每次有限盘面行星月掩的完整全球外接触掩带 / global paths of finite-disk planetary occultations in a window.
func (module *Moon) FindPlanetOccultationPaths(start, end time.Time, planet OccultationPlanet,
	options OccultationPathOptions) ([]PlanetOccultationPath, error) {
	if options.LunarLimb == nil {
		options.LunarLimb = module.lunarLimb
	}
	return basic.FindPlanetOccultationPathsWithInstance(start, end, planet, options, module.inst)
}
