package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// FindStarOccultationPaths 搜索单颗点源恒星月掩的全球掩带 / global paths of a point-source stellar occultation.
func (module *Moon) FindStarOccultationPaths(start, end time.Time, star StarCoordinate, options OccultationPathOptions) ([]StarOccultationPath, error) {
	if options.LunarLimb == nil {
		options.LunarLimb = module.lunarLimb
	}
	return basic.FindStarOccultationPathsWithInstance(start, end, star, options, module.inst)
}
