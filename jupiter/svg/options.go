package svg

import "time"

// GalileanSystemSVGOptions 系统图选项，零值为 900×600 中文 UTC+8、主图放大卫星盘 / system chart options, defaulting to 900×600 Chinese UTC+8 with enlarged main-view satellite disks.
type GalileanSystemSVGOptions struct {
	Width, Height int
	Title         string
	Language      string
	Location      *time.Location
	TrueScale     bool
}

// GalileanMutualEventSVGOptions 互现象图选项，TrueScale 不改变输出，放大的影核会标注 / mutual-event chart options; TrueScale has no effect, and enlarged shadow cores are labeled.
type GalileanMutualEventSVGOptions struct {
	Width, Height int
	Title         string
	Language      string
	Location      *time.Location
	TrueScale     bool
}

type chartOptions struct {
	width, height float64
	title         string
	english       bool
	location      *time.Location
	trueScale     bool
}

func normalizeOptions(width, height int, title, language string, location *time.Location, trueScale bool) chartOptions {
	if width <= 0 {
		width = 900
	}
	if height <= 0 {
		height = 600
	}
	if location == nil {
		location = time.FixedZone("UTC+8", 8*3600)
	}
	return chartOptions{float64(width), float64(height), title, language == "en", location, trueScale}
}

// GalileanPhenomenonSVGOptions 单星现象图选项，盘面恒按相对尺寸，TrueScale 不改变输出 / single-satellite phenomenon chart options; disks are always to relative scale.
type GalileanPhenomenonSVGOptions struct {
	Width, Height int
	Title         string
	Language      string
	Location      *time.Location
	TrueScale     bool
}
