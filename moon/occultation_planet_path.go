package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// FindPlanetOccultationPaths 搜索窗口内每次有限盘面行星月掩的完整全球外接触掩带。
// 查询端点使用核心求解器的 10 ms 数值选择容差。
// FindPlanetOccultationPaths searches the complete global outer-contact footprint of every finite-disk planetary occultation selected by the window.
// Query endpoints use the core solver's 10 ms numerical selection tolerance.
func FindPlanetOccultationPaths(start, end time.Time, planet OccultationPlanet,
	options OccultationPathOptions) ([]PlanetOccultationPath, error) {
	return basic.FindPlanetOccultationPaths(start, end, planet, options)
}
