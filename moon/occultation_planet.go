package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// FindPlanetOccultations 搜索固定地点的有限盘面行星月掩。
// 经度东为正、纬度北为正，单位为度；高度为平均海平面以上米数。圆形目标使用行星赤道半径，忽略行星环。
// FindPlanetOccultations searches finite-disk lunar occultations of one planet at a fixed observing site.
// Longitude is east-positive in degrees, latitude is north-positive in degrees, and height is above mean sea level in meters. The circular target uses the planet's equatorial body radius; rings are excluded.
func FindPlanetOccultations(start, end time.Time, planet OccultationPlanet, longitude, latitude, height float64,
	options OccultationSearchOptions) ([]PlanetOccultationInfo, error) {
	return basic.FindPlanetOccultations(start, end, planet, longitude, latitude, height, options)
}

// FindBestPlanetOccultations 返回窗口内每次有限盘面行星月掩的全球海平面几何掩甚点。
// 地平线可见性只报告、不参与点选择；查询端点使用核心求解器的 10 ms 数值选择容差。
// FindBestPlanetOccultations returns the global geometric greatest sea-level point for every finite-disk occultation of one planet in the window.
// Horizon visibility is reported but does not affect point selection. Query endpoints use the core solver's 10 ms numerical selection tolerance.
func FindBestPlanetOccultations(start, end time.Time, planet OccultationPlanet,
	options OccultationSearchOptions) ([]PlanetOccultationInfo, error) {
	return basic.FindBestPlanetOccultations(start, end, planet, options)
}
