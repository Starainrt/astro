package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// FindPlanetOccultations 搜索固定地点的有限盘面行星月掩 / finite-disk planetary occultations at a fixed site.
func (module *Moon) FindPlanetOccultations(start, end time.Time, planet OccultationPlanet, longitude, latitude, height float64,
	options OccultationSearchOptions) ([]PlanetOccultationInfo, error) {
	return basic.FindPlanetOccultationsWithInstance(start, end, planet, longitude, latitude, height, module.occultationOptions(options), module.inst)
}

// FindBestPlanetOccultations 返回窗口内每次有限盘面行星月掩的全球海平面几何掩甚点 / global greatest points of planetary occultations in a window.
func (module *Moon) FindBestPlanetOccultations(start, end time.Time, planet OccultationPlanet,
	options OccultationSearchOptions) ([]PlanetOccultationInfo, error) {
	return basic.FindBestPlanetOccultationsWithInstance(start, end, planet, module.occultationOptions(options), module.inst)
}
