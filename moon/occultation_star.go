package moon

import (
	"time"

	"github.com/starainrt/astro/basic"
)

// FindStarOccultations 搜索单颗点源恒星的月掩星。
//
// 输入坐标从历元传播并转换到当日视坐标系。此函数不会加载内嵌星表；需要星表搜索时，调用者必须显式加载并选择恒星。经度东为正、纬度北为正，单位为度；高度为平均海平面以上米数。
// FindStarOccultations searches for lunar occultations of one point-source star.
// The input coordinate is propagated from its epoch and converted to the apparent frame of date. This function never loads the embedded star catalog; callers that need catalog searches must load and select stars explicitly. Longitude is east-positive in degrees, latitude is north-positive in degrees, and height is the observer elevation above mean sea level in meters.
func FindStarOccultations(start, end time.Time, star StarCoordinate, longitude, latitude, height float64,
	options OccultationSearchOptions) ([]StarOccultationInfo, error) {
	return basic.FindStarOccultations(start, end, star, longitude, latitude, height, options)
}

// FindBestStarOccultations 返回窗口内每次恒星月掩的全球海平面几何掩甚点。
// 点选择不使用地平线或可见性评分；查询端点使用核心求解器的 10 ms 数值选择容差。
// FindBestStarOccultations returns the global geometric greatest sea-level point for every stellar occultation whose greatest instant is in the window.
// The point is selected without a horizon or visibility score. Query endpoints use the core solver's 10 ms numerical selection tolerance.
func FindBestStarOccultations(start, end time.Time, star StarCoordinate, options OccultationSearchOptions) ([]StarOccultationInfo, error) {
	return basic.FindBestStarOccultations(start, end, star, options)
}
