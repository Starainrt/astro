package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// FindStarOccultationPaths 搜索单颗点源恒星月掩的全球掩带。
// 调用者显式提供坐标；函数不会加载内嵌星表。查询端点使用核心求解器的 10 ms 数值选择容差。
// FindStarOccultationPaths searches the global lunar-occultation footprint of one point-source star.
// Callers provide the coordinate explicitly; this function does not load the embedded star catalog. Query endpoints use the core solver's 10 ms numerical selection tolerance.
func FindStarOccultationPaths(start, end time.Time, star StarCoordinate, options OccultationPathOptions) ([]StarOccultationPath, error) {
	return basic.FindStarOccultationPaths(start, end, star, options)
}
