package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// FindStarOccultations 搜索单颗点源恒星的月掩星 / occultations of a point-source star at a fixed site.
func (module *Moon) FindStarOccultations(start, end time.Time, star StarCoordinate, longitude, latitude, height float64,
	options OccultationSearchOptions) ([]StarOccultationInfo, error) {
	return basic.FindStarOccultationsWithInstance(start, end, star, longitude, latitude, height, module.occultationOptions(options), module.inst)
}

// FindBestStarOccultations 返回窗口内每次恒星月掩的全球海平面几何掩甚点 / global greatest points of stellar occultations in a window.
func (module *Moon) FindBestStarOccultations(start, end time.Time, star StarCoordinate, options OccultationSearchOptions) ([]StarOccultationInfo, error) {
	return basic.FindBestStarOccultationsWithInstance(start, end, star, module.occultationOptions(options), module.inst)
}
